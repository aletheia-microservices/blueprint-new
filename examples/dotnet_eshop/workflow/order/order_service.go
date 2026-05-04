package order

import (
	"context"
	"fmt"
	"sort"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type OrderService interface {
	// Routes for modifying orders
	CancelOrder(ctx context.Context, command CancelOrderRequest) (CancelOrderResponse, error)
	ShipOrder(ctx context.Context, command ShipOrderRequest) (ShipOrderResponse, error)

	// Routes for querying orders
	GetOrder(ctx context.Context, query GetOrderRequest) (GetOrderResponse, error)
	GetOrdersByUser(ctx context.Context, query GetOrdersByUserRequest) (GetOrdersByUserResponse, error)

	// Routes for creating orders
	CreateOrderDraft(ctx context.Context, command CreateOrderDraftCommand) (CreateOrderDraftResponse, error)
	CreateOrder(ctx context.Context, command CreateOrderCommand) (CreateOrderResult, error)

	// Integration event handler
	Init(ctx context.Context) error
}

type OrderServiceImpl struct {
	database                   backend.NoSQLDatabase
	awaitingValidationQueue    backend.Queue // order -> catalog: stock validation requests
	stockValidationResultQueue backend.Queue // catalog -> order: stock validation results
	stockConfirmedQueue        backend.Queue // order -> payment: after stock confirmed
	paymentResultQueue         backend.Queue // payment -> order: payment result
	catalogPaidQueue           backend.Queue // order -> catalog: order paid events
}

func NewOrderServiceImpl(ctx context.Context, database backend.NoSQLDatabase, awaitingValidationQueue backend.Queue, stockValidationResultQueue backend.Queue, stockConfirmedQueue backend.Queue, paymentResultQueue backend.Queue, catalogPaidQueue backend.Queue) (OrderService, error) {
	s := &OrderServiceImpl{
		database:                   database,
		awaitingValidationQueue:    awaitingValidationQueue,
		stockValidationResultQueue: stockValidationResultQueue,
		stockConfirmedQueue:        stockConfirmedQueue,
		paymentResultQueue:         paymentResultQueue,
		catalogPaidQueue:           catalogPaidQueue,
	}
	return s, nil
}

// Routes for modifying orders

func (s *OrderServiceImpl) CancelOrder(ctx context.Context, command CancelOrderRequest) (CancelOrderResponse, error) {
	order, err := s.find(ctx, command.OrderId)
	if err != nil {
		return CancelOrderResponse{IsSuccess: false}, err
	}
	order.Status = Cancelled
	_, err = s.update(ctx, order)
	if err != nil {
		return CancelOrderResponse{IsSuccess: false}, err
	}
	return CancelOrderResponse{IsSuccess: true}, nil
}

func (s *OrderServiceImpl) ShipOrder(ctx context.Context, command ShipOrderRequest) (ShipOrderResponse, error) {
	order, err := s.find(ctx, command.OrderId)
	if err != nil {
		return ShipOrderResponse{IsSuccess: false}, err
	}
	order.Status = Shipped
	_, err = s.update(ctx, order)
	if err != nil {
		return ShipOrderResponse{IsSuccess: false}, err
	}
	return ShipOrderResponse{IsSuccess: true}, nil
}

// Routes for querying orders

func (s *OrderServiceImpl) GetOrder(ctx context.Context, query GetOrderRequest) (GetOrderResponse, error) {
	order, err := s.find(ctx, query.OrderId)
	if err != nil {
		return GetOrderResponse{}, err
	}
	return GetOrderResponse{Order: order}, nil
}

func (s *OrderServiceImpl) GetOrdersByUser(ctx context.Context, query GetOrdersByUserRequest) (GetOrdersByUserResponse, error) {
	orders, err := s.findByUser(ctx, query.CustomerId)
	if err != nil {
		return GetOrdersByUserResponse{}, err
	}
	return GetOrdersByUserResponse{Orders: orders}, nil
}

// Routes for creating orders

func (s *OrderServiceImpl) CreateOrderDraft(ctx context.Context, command CreateOrderDraftCommand) (CreateOrderDraftResponse, error) {
	var orderItems []OrderItemDto
	var total float64
	for _, item := range command.Items {
		unitPrice := item.UnitPrice - item.Discount
		orderItem := OrderItemDto{
			ProductID: item.ProductId,
			Quantity:  item.Units,
			Price:     unitPrice,
		}
		orderItems = append(orderItems, orderItem)
		total += unitPrice * float64(item.Units)
	}
	return CreateOrderDraftResponse{
		OrderItems: orderItems,
		Total:      total,
	}, nil
}

// CreateOrder persists the order and publishes OrderStatusChangedToAwaitingValidationEvent
// so catalog can validate stock before payment is triggered.
func (s *OrderServiceImpl) CreateOrder(ctx context.Context, command CreateOrderCommand) (CreateOrderResult, error) {
	command.OrderDto.Id = uuid.NewString()
	command.OrderDto.Status = AwaitingValidation
	err := s.add(ctx, command.OrderDto)
	if err != nil {
		return CreateOrderResult{}, err
	}
	var stockItems []OrderStockItem
	for _, item := range command.OrderDto.OrderItems {
		stockItems = append(stockItems, OrderStockItem{
			ProductId: item.ProductID,
			Units:     item.Quantity,
		})
	}
	s.awaitingValidationQueue.Push(ctx, OrderStatusChangedToAwaitingValidationEvent{
		OrderId:         command.OrderDto.Id,
		OrderStockItems: stockItems,
	})
	return CreateOrderResult{Id: command.OrderDto.Id}, nil
}

// Integration event handler

// Init runs two concurrent background loops:
// - processStockValidationResultEvents: handles catalog's stock check response
// - processPaymentResultEvents: handles payment outcome
func (s *OrderServiceImpl) Init(ctx context.Context) error {
	go s.processStockValidationResultEvents(ctx)
	return s.processPaymentResultEvents(ctx)
}

// processStockValidationResultEvents handles OrderStockValidationResultEvent from catalog
// Matches OrderStockConfirmedIntegrationEventHandler / OrderStockRejectedIntegrationEventHandler
func (s *OrderServiceImpl) processStockValidationResultEvents(ctx context.Context) error {
	for {
		var event OrderStockValidationResultEvent
		ok, err := s.stockValidationResultQueue.Pop(ctx, &event)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		order, err := s.find(ctx, event.OrderId)
		if err != nil {
			continue
		}
		if event.Confirmed {
			order.Status = StockConfirmed
			if _, err := s.update(ctx, order); err == nil {
				s.stockConfirmedQueue.Push(ctx, OrderStatusChangedToStockConfirmedEvent{OrderId: order.Id})
			}
		} else {
			order.Status = Cancelled
			_, _ = s.update(ctx, order)
		}
	}
}

func (s *OrderServiceImpl) processPaymentResultEvents(ctx context.Context) error {
	for {
		var event OrderPaymentResultEvent
		ok, err := s.paymentResultQueue.Pop(ctx, &event)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		order, err := s.find(ctx, event.OrderId)
		if err != nil {
			continue
		}
		if event.Succeeded {
			order.Status = Paid
			if updated, err := s.update(ctx, order); err == nil {
				s.catalogPaidQueue.Push(ctx, OrderPaidEvent{OrderId: updated.Id, OrderItems: updated.OrderItems})
			}
		} else {
			order.Status = Cancelled
			_, _ = s.update(ctx, order)
		}
	}
}

func (s *OrderServiceImpl) add(ctx context.Context, order OrderDto) error {
	collection, err := s.database.GetCollection(ctx, "order_db", "order")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, order)
}

func (s *OrderServiceImpl) find(ctx context.Context, id string) (OrderDto, error) {
	collection, err := s.database.GetCollection(ctx, "order_db", "order")
	if err != nil {
		return OrderDto{}, err
	}
	filter := bson.D{{Key: "Id", Value: id}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return OrderDto{}, err
	}
	var order OrderDto
	ok, err := cursor.One(ctx, &order)
	if err != nil {
		return OrderDto{}, err
	}
	if !ok {
		return OrderDto{}, fmt.Errorf("order not found for id (%s)", id)
	}
	return order, nil
}

func (s *OrderServiceImpl) findByUser(ctx context.Context, customerId string) ([]OrderDto, error) {
	collection, err := s.database.GetCollection(ctx, "order_db", "order")
	if err != nil {
		return nil, err
	}
	filter := bson.D{{Key: "CustomerId", Value: customerId}}
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	var orders []OrderDto
	err = cursor.All(ctx, &orders)
	if err != nil {
		return nil, err
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].OrderName <= orders[j].OrderName
	})
	return orders, nil
}

func (s *OrderServiceImpl) update(ctx context.Context, order OrderDto) (OrderDto, error) {
	collection, err := s.database.GetCollection(ctx, "order_db", "order")
	if err != nil {
		return OrderDto{}, err
	}
	filter := bson.D{{Key: "Id", Value: order.Id}}
	updated, err := collection.ReplaceOne(ctx, filter, order)
	if err != nil {
		return OrderDto{}, err
	}
	if updated == 0 {
		return OrderDto{}, fmt.Errorf("order not found for id (%s)", order.Id)
	}
	return order, nil
}
