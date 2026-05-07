package order

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/blueprint-uservices/blueprint/examples/eshopmicroservices/workflow/basket"
)

type OrderService interface {
	CreateNewOrder(ctx context.Context, command CreateOrderCommand) (CreateOrderResult, error)
	UpdateOrder(ctx context.Context, command UpdateOrderCommand) (UpdateOrderResult, error)
	DeleteOrder(ctx context.Context, command DeleteOrderCommand) (DeleteOrderResult, error)
	GetOrdersByCustomer(ctx context.Context, query GetOrdersByCustomerQuery) (GetOrdersByCustomerResult, error)
	
	Init(ctx context.Context) error  // new RunnableHTTP in Blueprint
}

type OrderServiceImpl struct {
	database backend.NoSQLDatabase
	queue    backend.Queue
}

func NewOrderServiceImpl(ctx context.Context, database backend.NoSQLDatabase, queue backend.Queue) (OrderService, error) {
	s := &OrderServiceImpl{
		database: database,
		queue:    queue,
	}
	return s, nil
}

func (s *OrderServiceImpl) Init(ctx context.Context) error {
	for {
		var message basket.BasketChekoutEvent
		ok, err := s.queue.Pop(ctx, &message)
		if err != nil {
			log.Printf("[ERROR] init: error reading from queue: %s\n", err.Error())
			continue
		}
		if !ok {
			log.Printf("[ERROR] init: no message in queue")
			continue
		}

		addressDto := AddressDto{message.FirstName, message.LastName, message.EmailAddress, message.AddressLine, message.Country, message.State, message.ZipCode}
		paymentDto := PaymentDto{message.CardName, message.CardNumber, message.Expiration, message.CVV, message.PaymentMethod}
		orderId := uuid.NewString()
		orderDto := OrderDto{
			Id:              orderId,
			CustomerId:      message.CustomerId,
			OrderName:       message.UserName,
			ShippingAddress: addressDto,
			BillingAddress:  addressDto,
			Payment:         paymentDto,
			Status:          Pending,
			OrderItems: []OrderItemDto{
				{orderId, "5334c996-8457-4cf0-815c-ed2b77c4ff61", 2, 500},
				{orderId, "c67d6323-e8b1-4bdf-9a75-b0d0d2e7e914", 1, 400},
			},
		}
		_, err = s.CreateNewOrder(ctx, CreateOrderCommand{OrderDto: orderDto})
		if err != nil {
			log.Printf("[ERROR] init: error creating ne order: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func (s *OrderServiceImpl) CreateNewOrder(ctx context.Context, command CreateOrderCommand) (CreateOrderResult, error) {
	command.OrderDto.Id = uuid.NewString()
	err := s.add(ctx, command.OrderDto)
	if err != nil {
		return CreateOrderResult{}, err
	}
	return CreateOrderResult{Id: command.OrderDto.Id}, nil
}

func (s *OrderServiceImpl) UpdateOrder(ctx context.Context, command UpdateOrderCommand) (UpdateOrderResult, error) {
	orderId := command.OrderDto.Id
	_, err := s.find(ctx, orderId)
	if err != nil {
		return UpdateOrderResult{false}, err
	}

	var newOrder OrderDto
	updateOrderWithNewValues(&newOrder, &command.OrderDto)
	s.update(ctx, newOrder)
	return UpdateOrderResult{true}, nil
}

func (s *OrderServiceImpl) DeleteOrder(ctx context.Context, command DeleteOrderCommand) (DeleteOrderResult, error) {
	orderId := command.Id
	_, err := s.find(ctx, orderId)
	if err != nil {
		return DeleteOrderResult{false}, err
	}
	s.remove(ctx, orderId)
	return DeleteOrderResult{true}, nil
}

func (s *OrderServiceImpl) GetOrdersByCustomer(ctx context.Context, query GetOrdersByCustomerQuery) (GetOrdersByCustomerResult, error) {
	customerId := query.CustomerId
	orders, err := s.findByCustomer(ctx, customerId)
	if err != nil {
		return GetOrdersByCustomerResult{nil}, err
	}
	return GetOrdersByCustomerResult{orders}, nil
}

func (s *OrderServiceImpl) add(ctx context.Context, order OrderDto) error {
	collection, err := s.database.GetCollection(ctx, "order_db", "order")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, order)
}

func (s *OrderServiceImpl) remove(ctx context.Context, id string) error {
	collection, err := s.database.GetCollection(ctx, "order_db", "order")
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "Id", Value: id}}
	return collection.DeleteOne(ctx, filter)
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

func (s *OrderServiceImpl) findByCustomer(ctx context.Context, customerId string) ([]OrderDto, error) {
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

	// sort by OrderName
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

func updateOrderWithNewValues(order *OrderDto, orderUpdate *OrderDto) {
	var updatedShippingAddress AddressDto = orderUpdate.ShippingAddress
	var updatedBillingAddress AddressDto = orderUpdate.BillingAddress
	var updatedPayment PaymentDto = orderUpdate.Payment

	order.OrderName = orderUpdate.OrderName
	order.ShippingAddress = updatedShippingAddress
	order.BillingAddress = updatedBillingAddress
	order.Payment = updatedPayment
	order.Status = orderUpdate.Status
}
