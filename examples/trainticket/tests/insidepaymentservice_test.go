package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/trainticket/workflow/trainticket"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

var insidePaymentServiceRegistry = registry.NewServiceRegistry[trainticket.InsidePaymentService]("inside_payment_service")

func init() {
	insidePaymentServiceRegistry.Register("local", func(ctx context.Context) (trainticket.InsidePaymentService, error) {
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
		orderDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		orderService, err := trainticket.NewOrderServiceImpl(ctx, orderDB)
		if err != nil {
			return nil, err
		}
		insidePaymentDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return trainticket.NewInsidePaymentServiceImpl(ctx, insidePaymentDB, paymentService, orderService)
	})
}

func TestInsidePaymentServiceCreateAccount(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	info := trainticket.AccountInfo{
		UserId: "ip_user001",
		Money:  "1000",
	}
	created, err := service.CreateAccount(ctx, info)
	assert.NoError(t, err)
	assert.True(t, created)
}

func TestInsidePaymentServiceCreateAccountDuplicate(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	info := trainticket.AccountInfo{
		UserId: "ip_user002",
		Money:  "500",
	}
	_, err = service.CreateAccount(ctx, info)
	assert.NoError(t, err)

	// second create should return false (already exists)
	created, err := service.CreateAccount(ctx, info)
	assert.NoError(t, err)
	assert.False(t, created)
}

func TestInsidePaymentServiceAddMoney(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	err = service.AddMoney(ctx, "ip_user003", "200")
	assert.NoError(t, err)
}

func TestInsidePaymentServiceQueryAccount(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	err = service.AddMoney(ctx, "ip_user004", "300")
	assert.NoError(t, err)

	balance, err := service.QueryAccount(ctx, "ip_user004")
	assert.NoError(t, err)
	assert.NotEmpty(t, balance)
}

func TestInsidePaymentServiceQueryAddMoney(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	err = service.AddMoney(ctx, "ip_user005", "100")
	assert.NoError(t, err)

	moneys, err := service.QueryAddMoney(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, moneys)
}

func TestInsidePaymentServiceQueryPayment(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	payments, err := service.QueryPayment(ctx)
	assert.NoError(t, err)
	// may be empty initially; just verify no error
	_ = payments
}

func TestInsidePaymentServicePayDifferenceSavesWithID(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// fund the account so PayDifference takes the direct-save path (no outside payment needed)
	_, err = service.CreateAccount(ctx, trainticket.AccountInfo{UserId: "ip_user_paydiff001", Money: "1000"})
	assert.NoError(t, err)

	err = service.PayDifference(ctx, trainticket.PaymentInfo{
		UserId:  "ip_user_paydiff001",
		OrderId: "ip_order_paydiff001",
		Price:   "50",
	})
	assert.NoError(t, err)

	payments, err := service.QueryPayment(ctx)
	assert.NoError(t, err)
	for _, p := range payments {
		if p.OrderID == "ip_order_paydiff001" {
			assert.NotEmpty(t, p.ID)
			return
		}
	}
	t.Fatal("payment for ip_order_paydiff001 not found")
}

func TestInsidePaymentServiceDrawback(t *testing.T) {
	ctx := context.Background()
	service, err := insidePaymentServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	err = service.AddMoney(ctx, "ip_user006", "500")
	assert.NoError(t, err)

	err = service.Drawback(ctx, "ip_user006", "100")
	assert.NoError(t, err)
}
