package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/order"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var orderServiceRegistry = registry.NewServiceRegistry[order.OrderService]("order_service")

func init() {
	orderServiceRegistry.Register("local", func(ctx context.Context) (order.OrderService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		awaitingValidationQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		stockValidationResultQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		stockConfirmedQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		paymentResultQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		catalogPaidQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		return order.NewOrderServiceImpl(ctx, db, awaitingValidationQueue, stockValidationResultQueue, stockConfirmedQueue, paymentResultQueue, catalogPaidQueue)
	})
}

func makeOrderAddress(firstName, lastName string) order.AddressDto {
	return order.AddressDto{
		FirstName:    firstName,
		LastName:     lastName,
		EmailAddress: firstName + "@example.com",
		AddressLine:  "123 Test St",
		Country:      "US",
		State:        "NY",
		ZipCode:      "10001",
	}
}

func makeOrderPayment() order.PaymentDto {
	return order.PaymentDto{
		CardName:      "Test User",
		CardNumber:    "4242424242424242",
		Expiration:    "12/25",
		CCV:           "123",
		PaymentMethod: 1,
	}
}

func TestOrderServiceCreateOrder(t *testing.T) {
	ctx := context.Background()
	orderService, err := orderServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	customerId := uuid.NewString()
	orderDto := order.OrderDto{
		CustomerId:      customerId,
		OrderName:       "order@example.com",
		ShippingAddress: makeOrderAddress("John", "Doe"),
		BillingAddress:  makeOrderAddress("John", "Doe"),
		Payment:         makeOrderPayment(),
		Status:          order.Pending,
	}

	result, err := orderService.CreateOrder(ctx, order.CreateOrderCommand{OrderDto: orderDto})
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Id)
}

func TestOrderServiceGetOrder(t *testing.T) {
	ctx := context.Background()
	orderService, err := orderServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	customerId := uuid.NewString()
	created, err := orderService.CreateOrder(ctx, order.CreateOrderCommand{OrderDto: order.OrderDto{
		CustomerId:      customerId,
		OrderName:       "get@example.com",
		ShippingAddress: makeOrderAddress("Jane", "Smith"),
		BillingAddress:  makeOrderAddress("Jane", "Smith"),
		Payment:         makeOrderPayment(),
		Status:          order.Pending,
	}})
	assert.NoError(t, err)

	result, err := orderService.GetOrder(ctx, order.GetOrderRequest{OrderId: created.Id})
	assert.NoError(t, err)
	assert.Equal(t, created.Id, result.Order.Id)
}

func TestOrderServiceGetOrdersByUser(t *testing.T) {
	ctx := context.Background()
	orderService, err := orderServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	customerId := uuid.NewString()

	_, err = orderService.CreateOrder(ctx, order.CreateOrderCommand{OrderDto: order.OrderDto{
		CustomerId:      customerId,
		OrderName:       "cust1@example.com",
		ShippingAddress: makeOrderAddress("Alice", "Smith"),
		BillingAddress:  makeOrderAddress("Alice", "Smith"),
		Payment:         makeOrderPayment(),
		Status:          order.Pending,
	}})
	assert.NoError(t, err)

	_, err = orderService.CreateOrder(ctx, order.CreateOrderCommand{OrderDto: order.OrderDto{
		CustomerId:      customerId,
		OrderName:       "cust2@example.com",
		ShippingAddress: makeOrderAddress("Alice", "Smith"),
		BillingAddress:  makeOrderAddress("Alice", "Smith"),
		Payment:         makeOrderPayment(),
		Status:          order.Pending,
	}})
	assert.NoError(t, err)

	result, err := orderService.GetOrdersByUser(ctx, order.GetOrdersByUserRequest{CustomerId: customerId})
	assert.NoError(t, err)
	assert.True(t, len(result.Orders) >= 2)
}

func TestOrderServiceCancelOrder(t *testing.T) {
	ctx := context.Background()
	orderService, err := orderServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	customerId := uuid.NewString()
	created, err := orderService.CreateOrder(ctx, order.CreateOrderCommand{OrderDto: order.OrderDto{
		CustomerId:      customerId,
		OrderName:       "cancel@example.com",
		ShippingAddress: makeOrderAddress("Carol", "White"),
		BillingAddress:  makeOrderAddress("Carol", "White"),
		Payment:         makeOrderPayment(),
		Status:          order.Pending,
	}})
	assert.NoError(t, err)

	result, err := orderService.CancelOrder(ctx, order.CancelOrderRequest{OrderId: created.Id})
	assert.NoError(t, err)
	assert.True(t, result.IsSuccess)

	getResult, err := orderService.GetOrder(ctx, order.GetOrderRequest{OrderId: created.Id})
	assert.NoError(t, err)
	assert.Equal(t, order.Cancelled, getResult.Order.Status)
}

func TestOrderServiceShipOrder(t *testing.T) {
	ctx := context.Background()
	orderService, err := orderServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	created, err := orderService.CreateOrder(ctx, order.CreateOrderCommand{OrderDto: order.OrderDto{
		CustomerId:      uuid.NewString(),
		OrderName:       "ship@example.com",
		ShippingAddress: makeOrderAddress("Eve", "Green"),
		BillingAddress:  makeOrderAddress("Eve", "Green"),
		Payment:         makeOrderPayment(),
		Status:          order.StockConfirmed,
	}})
	assert.NoError(t, err)

	result, err := orderService.ShipOrder(ctx, order.ShipOrderRequest{OrderId: created.Id})
	assert.NoError(t, err)
	assert.True(t, result.IsSuccess)

	getResult, err := orderService.GetOrder(ctx, order.GetOrderRequest{OrderId: created.Id})
	assert.NoError(t, err)
	assert.Equal(t, order.Shipped, getResult.Order.Status)
}

func TestOrderServiceCreateOrderDraft(t *testing.T) {
	ctx := context.Background()
	orderService, err := orderServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	result, err := orderService.CreateOrderDraft(ctx, order.CreateOrderDraftCommand{
		BuyerId: uuid.NewString(),
		Items: []order.DraftOrderItem{
			{ProductId: 1, ProductName: "Laptop", UnitPrice: 999.99, Discount: 0, Units: 1},
			{ProductId: 2, ProductName: "Mouse", UnitPrice: 29.99, Discount: 5.0, Units: 2},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, result.OrderItems, 2)
	assert.True(t, result.Total > 0)
}
