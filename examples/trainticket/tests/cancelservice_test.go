package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/trainticket/workflow/trainticket"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/stretchr/testify/assert"
)

var cancelServiceRegistry = registry.NewServiceRegistry[trainticket.CancelService]("cancel_service")
var cancelOrderDB *simplenosqldb.SimpleNoSQLDB
var cancelUserDB *simplenosqldb.SimpleNoSQLDB
var cancelInsidePaymentDB *simplenosqldb.SimpleNoSQLDB

func init() {
	cancelServiceRegistry.Register("local", func(ctx context.Context) (trainticket.CancelService, error) {
		var err error
		cancelOrderDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		orderService, err := trainticket.NewOrderServiceImpl(ctx, cancelOrderDB)
		if err != nil {
			return nil, err
		}

		cancelUserDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		userService, err := trainticket.NewUserServiceImpl(ctx, cancelUserDB)
		if err != nil {
			return nil, err
		}

		cancelInsidePaymentDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		payDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		moneyDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		paymentService, err := trainticket.NewPaymentServiceImpl(ctx, payDB, moneyDB)
		if err != nil {
			return nil, err
		}
		insidePaymentService, err := trainticket.NewInsidePaymentServiceImpl(ctx, cancelInsidePaymentDB, paymentService, orderService)
		if err != nil {
			return nil, err
		}

		queue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}

		return trainticket.NewCancelServiceImpl(ctx, orderService, userService, insidePaymentService, queue)
	})
}

func TestCancelServiceCalculateRefundNotPaid(t *testing.T) {
	ctx := context.Background()
	service, err := cancelServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	orderService, err := trainticket.NewOrderServiceImpl(ctx, cancelOrderDB)
	assert.NoError(t, err)

	order := trainticket.Order{
		ID:     "cancel_ord001",
		Status: trainticket.ORDER_STATUS_NOT_PAID,
		Price:  "100.00",
	}
	_, err = orderService.CreateNewOrder(ctx, order)
	assert.NoError(t, err)

	refund, err := service.CalculateRefund(ctx, "cancel_ord001")
	assert.NoError(t, err)
	assert.Equal(t, "not paid", refund)
}

func TestCancelServiceCalculateRefundInvalidStatus(t *testing.T) {
	ctx := context.Background()
	service, err := cancelServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	orderService, err := trainticket.NewOrderServiceImpl(ctx, cancelOrderDB)
	assert.NoError(t, err)

	order := trainticket.Order{
		ID:     "cancel_ord002",
		Status: trainticket.ORDER_STATUS_USED,
		Price:  "100.00",
	}
	_, err = orderService.CreateNewOrder(ctx, order)
	assert.NoError(t, err)

	_, err = service.CalculateRefund(ctx, "cancel_ord002")
	assert.Error(t, err)
}

func TestCancelServiceCancelOrder(t *testing.T) {
	ctx := context.Background()
	service, err := cancelServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	orderService, err := trainticket.NewOrderServiceImpl(ctx, cancelOrderDB)
	assert.NoError(t, err)
	userService, err := trainticket.NewUserServiceImpl(ctx, cancelUserDB)
	assert.NoError(t, err)

	user := trainticket.User{
		UserID:   "cancel_user001",
		Username: "alice",
		Email:    "alice@example.com",
	}
	err = userService.SaveUser(ctx, user)
	assert.NoError(t, err)

	order := trainticket.Order{
		ID:         "cancel_ord003",
		AccountID:  "cancel_user001",
		Status:     trainticket.ORDER_STATUS_NOT_PAID,
		Price:      "0.00",
		TravelDate: "2020-01-01 00:00:00",
		TravelTime: "2020-01-01 00:00:00",
	}
	_, err = orderService.CreateNewOrder(ctx, order)
	assert.NoError(t, err)

	err = service.CancelOrder(ctx, "cancel_ord003", "cancel_user001")
	assert.NoError(t, err)
}
