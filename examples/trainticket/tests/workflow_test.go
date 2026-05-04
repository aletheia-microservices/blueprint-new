package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/trainticket/workflow/trainticket"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/stretchr/testify/assert"
)

// magicDate is a past date that causes calculateRefund to return "0" (no refund)
// It is required for any test that drives CancelOrder through the calculateRefund path
const magicDate = "2020-01-01 00:00:00"

// ============================================================
// Cancel workflow environment
// ============================================================

type cancelWorkflowEnv struct {
	cancelService        trainticket.CancelService
	orderService         trainticket.OrderService
	userService          trainticket.UserService
	insidePaymentService trainticket.InsidePaymentService
}

func newCancelWorkflowEnv(ctx context.Context) (*cancelWorkflowEnv, error) {
	orderDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	orderService, err := trainticket.NewOrderServiceImpl(ctx, orderDB)
	if err != nil {
		return nil, err
	}

	userDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	userService, err := trainticket.NewUserServiceImpl(ctx, userDB)
	if err != nil {
		return nil, err
	}

	insidePayDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
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
	insidePaymentService, err := trainticket.NewInsidePaymentServiceImpl(ctx, insidePayDB, paymentService, orderService)
	if err != nil {
		return nil, err
	}

	emailQueue, err := simplequeue.NewSimpleQueue(ctx)
	if err != nil {
		return nil, err
	}

	cancelService, err := trainticket.NewCancelServiceImpl(ctx, orderService, userService, insidePaymentService, emailQueue)
	if err != nil {
		return nil, err
	}

	return &cancelWorkflowEnv{
		cancelService:        cancelService,
		orderService:         orderService,
		userService:          userService,
		insidePaymentService: insidePaymentService,
	}, nil
}

// ============================================================
// Rebook workflow environment
// ============================================================

type rebookWorkflowEnv struct {
	rebookService trainticket.RebookService
	orderService  trainticket.OrderService
}

func newRebookWorkflowEnv(ctx context.Context) (*rebookWorkflowEnv, error) {
	orderDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	orderService, err := trainticket.NewOrderServiceImpl(ctx, orderDB)
	if err != nil {
		return nil, err
	}

	configDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	configService, err := trainticket.NewConfigServiceImpl(ctx, configDB)
	if err != nil {
		return nil, err
	}
	seatService, err := trainticket.NewSeatServiceImpl(ctx, orderService, configService)
	if err != nil {
		return nil, err
	}

	stationDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	stationService, err := trainticket.NewStationServiceImpl(ctx, stationDB)
	if err != nil {
		return nil, err
	}
	trainDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	trainService, err := trainticket.NewTrainServiceImpl(ctx, trainDB)
	if err != nil {
		return nil, err
	}
	routeDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	routeService, err := trainticket.NewRouteServiceImpl(ctx, routeDB)
	if err != nil {
		return nil, err
	}
	priceDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	priceService, err := trainticket.NewPriceServiceImpl(ctx, priceDB)
	if err != nil {
		return nil, err
	}
	basicService, err := trainticket.NewBasicServiceImpl(ctx, stationService, trainService, routeService, priceService)
	if err != nil {
		return nil, err
	}
	travelDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	travelService, err := trainticket.NewTravelServiceImpl(ctx, basicService, seatService, routeService, trainService, travelDB)
	if err != nil {
		return nil, err
	}

	insidePayDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
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
	insidePaymentService, err := trainticket.NewInsidePaymentServiceImpl(ctx, insidePayDB, paymentService, orderService)
	if err != nil {
		return nil, err
	}

	rebookService, err := trainticket.NewRebookServiceImpl(ctx, seatService, travelService, orderService, trainService, routeService, insidePaymentService)
	if err != nil {
		return nil, err
	}

	return &rebookWorkflowEnv{
		rebookService: rebookService,
		orderService:  orderService,
	}, nil
}

// ============================================================
// InsidePayment workflow environment
// ============================================================

type insidePaymentWorkflowEnv struct {
	insidePaymentService trainticket.InsidePaymentService
	orderService         trainticket.OrderService
}

func newInsidePaymentWorkflowEnv(ctx context.Context) (*insidePaymentWorkflowEnv, error) {
	orderDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	orderService, err := trainticket.NewOrderServiceImpl(ctx, orderDB)
	if err != nil {
		return nil, err
	}

	insidePayDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
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
	insidePaymentService, err := trainticket.NewInsidePaymentServiceImpl(ctx, insidePayDB, paymentService, orderService)
	if err != nil {
		return nil, err
	}

	return &insidePaymentWorkflowEnv{
		insidePaymentService: insidePaymentService,
		orderService:         orderService,
	}, nil
}

// ============================================================
// Preserve workflow environment
// ============================================================

type preserveWorkflowEnv struct {
	preserveService trainticket.PreserveService
	contactsService trainticket.ContactsService
	userService     trainticket.UserService
	travelService   trainticket.TravelService
	stationService  trainticket.StationService
	trainService    trainticket.TrainService
	routeService    trainticket.RouteService
	priceService    trainticket.PriceService
}

func newPreserveWorkflowEnv(ctx context.Context) (*preserveWorkflowEnv, error) {
	assuranceDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	assuranceService, err := trainticket.NewAssuranceServiceImpl(ctx, assuranceDB)
	if err != nil {
		return nil, err
	}

	stationDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	stationService, err := trainticket.NewStationServiceImpl(ctx, stationDB)
	if err != nil {
		return nil, err
	}
	trainDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	trainService, err := trainticket.NewTrainServiceImpl(ctx, trainDB)
	if err != nil {
		return nil, err
	}
	routeDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	routeService, err := trainticket.NewRouteServiceImpl(ctx, routeDB)
	if err != nil {
		return nil, err
	}
	priceDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	priceService, err := trainticket.NewPriceServiceImpl(ctx, priceDB)
	if err != nil {
		return nil, err
	}
	basicService, err := trainticket.NewBasicServiceImpl(ctx, stationService, trainService, routeService, priceService)
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
	configDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	configService, err := trainticket.NewConfigServiceImpl(ctx, configDB)
	if err != nil {
		return nil, err
	}
	seatService, err := trainticket.NewSeatServiceImpl(ctx, orderService, configService)
	if err != nil {
		return nil, err
	}

	travelDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	travelService, err := trainticket.NewTravelServiceImpl(ctx, basicService, seatService, routeService, trainService, travelDB)
	if err != nil {
		return nil, err
	}

	consignPriceDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	consignPriceService, err := trainticket.NewConsignPriceServiceImpl(ctx, consignPriceDB)
	if err != nil {
		return nil, err
	}
	consignDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	consignService, err := trainticket.NewConsignServiceImpl(ctx, consignPriceService, consignDB)
	if err != nil {
		return nil, err
	}

	contactsDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	contactsService, err := trainticket.NewContactsServiceImpl(ctx, contactsDB)
	if err != nil {
		return nil, err
	}

	trainFoodDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	trainFoodService, err := trainticket.NewTrainFoodServiceImpl(ctx, trainFoodDB)
	if err != nil {
		return nil, err
	}
	stationFoodDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	stationFoodService, err := trainticket.NewStationFoodServiceImpl(ctx, stationFoodDB)
	if err != nil {
		return nil, err
	}
	foodDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	foodQueue, err := simplequeue.NewSimpleQueue(ctx)
	if err != nil {
		return nil, err
	}
	foodService, err := trainticket.NewFoodServiceImpl(ctx, foodDB, foodQueue, trainFoodService, travelService, stationFoodService)
	if err != nil {
		return nil, err
	}

	userDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	if err != nil {
		return nil, err
	}
	userService, err := trainticket.NewUserServiceImpl(ctx, userDB)
	if err != nil {
		return nil, err
	}

	emailQueue, err := simplequeue.NewSimpleQueue(ctx)
	if err != nil {
		return nil, err
	}

	preserveService, err := trainticket.NewPreserveServiceImpl(
		ctx, assuranceService, basicService, consignService, contactsService,
		foodService, orderService, seatService, stationService, travelService, userService, emailQueue,
	)
	if err != nil {
		return nil, err
	}

	return &preserveWorkflowEnv{
		preserveService: preserveService,
		contactsService: contactsService,
		userService:     userService,
		travelService:   travelService,
		stationService:  stationService,
		trainService:    trainService,
		routeService:    routeService,
		priceService:    priceService,
	}, nil
}

// ============================================================
// Cancel workflow tests
// ============================================================

// TestCancelWorkflow_OrderNotFound verifies that cancelling a non-existent order returns an error
func TestCancelWorkflow_OrderNotFound(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.cancelService.CancelOrder(ctx, "nonexistent_order", "any_user")
	assert.Error(t, err)
}

// TestCancelWorkflow_CancelNotPaidOrder verifies the full cancel path for a NOT_PAID order:
// the order status is persisted as CANCELED after a successful call
func TestCancelWorkflow_CancelNotPaidOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.userService.SaveUser(ctx, trainticket.User{
		UserID:   "wf_cancel_user1",
		Username: "alice",
		Email:    "alice@example.com",
	})
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:         "wf_cancel_ord1",
		AccountID:  "wf_cancel_user1",
		Status:     trainticket.ORDER_STATUS_NOT_PAID,
		Price:      "0.00",
		TravelDate: magicDate,
		TravelTime: magicDate,
	})
	assert.NoError(t, err)

	err = env.cancelService.CancelOrder(ctx, "wf_cancel_ord1", "wf_cancel_user1")
	assert.NoError(t, err)

	updated, err := env.orderService.GetOrderById(ctx, "wf_cancel_ord1")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_CANCELED, updated.Status)
}

// TestCancelWorkflow_CancelChangeStatusOrder verifies that a CHANGE-status order can be cancelled
func TestCancelWorkflow_CancelChangeStatusOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.userService.SaveUser(ctx, trainticket.User{
		UserID:   "wf_cancel_user2",
		Username: "bob",
		Email:    "bob@example.com",
	})
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:         "wf_cancel_ord2",
		AccountID:  "wf_cancel_user2",
		Status:     trainticket.ORDER_STATUS_CHANGE,
		Price:      "0.00",
		TravelDate: magicDate,
		TravelTime: magicDate,
	})
	assert.NoError(t, err)

	err = env.cancelService.CancelOrder(ctx, "wf_cancel_ord2", "wf_cancel_user2")
	assert.NoError(t, err)

	updated, err := env.orderService.GetOrderById(ctx, "wf_cancel_ord2")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_CANCELED, updated.Status)
}

// TestCancelWorkflow_CancelUsedOrder verifies that a USED order is not modified by CancelOrder
// (falls outside the eligible status set, so the call is a silent no-op)
func TestCancelWorkflow_CancelUsedOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_cancel_ord3",
		Status: trainticket.ORDER_STATUS_USED,
	})
	assert.NoError(t, err)

	err = env.cancelService.CancelOrder(ctx, "wf_cancel_ord3", "any_user")
	assert.NoError(t, err)

	unchanged, err := env.orderService.GetOrderById(ctx, "wf_cancel_ord3")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_USED, unchanged.Status)
}

// TestCancelWorkflow_CancelAlreadyCanceledOrder verifies that cancelling an already CANCELED
// order is a silent no-op that does not modify the status
func TestCancelWorkflow_CancelAlreadyCanceledOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_cancel_ord4",
		Status: trainticket.ORDER_STATUS_CANCELED,
	})
	assert.NoError(t, err)

	err = env.cancelService.CancelOrder(ctx, "wf_cancel_ord4", "any_user")
	assert.NoError(t, err)

	unchanged, err := env.orderService.GetOrderById(ctx, "wf_cancel_ord4")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_CANCELED, unchanged.Status)
}

// TestCancelWorkflow_CalculateRefund_NotPaid verifies that the refund label for a NOT_PAID
// order is the string "not paid"
func TestCancelWorkflow_CalculateRefund_NotPaid(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_refund_ord1",
		Status: trainticket.ORDER_STATUS_NOT_PAID,
		Price:  "100.00",
	})
	assert.NoError(t, err)

	refund, err := env.cancelService.CalculateRefund(ctx, "wf_refund_ord1")
	assert.NoError(t, err)
	assert.Equal(t, "not paid", refund)
}

// TestCancelWorkflow_CalculateRefund_NonRefundableStatuses checks that COLLECTED, USED, and
// CANCELED orders are rejected with an error from CalculateRefund
func TestCancelWorkflow_CalculateRefund_NonRefundableStatuses(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	for _, status := range []int{
		trainticket.ORDER_STATUS_COLLECTED,
		trainticket.ORDER_STATUS_USED,
		trainticket.ORDER_STATUS_CANCELED,
	} {
		orderId := fmt.Sprintf("wf_refund_status_%d", status)
		_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
			ID:     orderId,
			Status: status,
		})
		assert.NoError(t, err)

		_, err = env.cancelService.CalculateRefund(ctx, orderId)
		assert.Error(t, err, "expected error for order status %d", status)
	}
}

// TestCancelWorkflow_FullLifecycle exercises the full create-then-cancel lifecycle:
// a user and a NOT_PAID order are created, the order is cancelled, and the final
// status is asserted to be CANCELED
func TestCancelWorkflow_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.userService.SaveUser(ctx, trainticket.User{
		UserID:   "wf_full_user1",
		Username: "charlie",
		Email:    "charlie@example.com",
	})
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:         "wf_full_ord1",
		AccountID:  "wf_full_user1",
		Status:     trainticket.ORDER_STATUS_NOT_PAID,
		Price:      "120.00",
		TravelDate: magicDate,
		TravelTime: magicDate,
	})
	assert.NoError(t, err)

	err = env.cancelService.CancelOrder(ctx, "wf_full_ord1", "wf_full_user1")
	assert.NoError(t, err)

	o, err := env.orderService.GetOrderById(ctx, "wf_full_ord1")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_CANCELED, o.Status)
}

// TestCancelWorkflow_MultipleOrdersCancelIndependently verifies that cancelling one order
// does not affect the status of other orders belonging to the same user
func TestCancelWorkflow_MultipleOrdersCancelIndependently(t *testing.T) {
	ctx := context.Background()
	env, err := newCancelWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.userService.SaveUser(ctx, trainticket.User{
		UserID:   "wf_multi_user1",
		Username: "diana",
		Email:    "diana@example.com",
	})
	assert.NoError(t, err)

	for i := 1; i <= 3; i++ {
		_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
			ID:         fmt.Sprintf("wf_multi_ord%d", i),
			AccountID:  "wf_multi_user1",
			Status:     trainticket.ORDER_STATUS_NOT_PAID,
			Price:      "0.00",
			TravelDate: magicDate,
			TravelTime: magicDate,
		})
		assert.NoError(t, err)
	}

	// Cancel only the first order
	err = env.cancelService.CancelOrder(ctx, "wf_multi_ord1", "wf_multi_user1")
	assert.NoError(t, err)

	o1, err := env.orderService.GetOrderById(ctx, "wf_multi_ord1")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_CANCELED, o1.Status)

	o2, err := env.orderService.GetOrderById(ctx, "wf_multi_ord2")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_NOT_PAID, o2.Status)

	o3, err := env.orderService.GetOrderById(ctx, "wf_multi_ord3")
	assert.NoError(t, err)
	assert.Equal(t, trainticket.ORDER_STATUS_NOT_PAID, o3.Status)
}

// ============================================================
// Rebook workflow tests
// ============================================================

// TestRebookWorkflow_RejectNotPaid verifies that rebooking a NOT_PAID order fails
func TestRebookWorkflow_RejectNotPaid(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_rebook_ord1",
		Status: trainticket.ORDER_STATUS_NOT_PAID,
	})
	assert.NoError(t, err)

	err = env.rebookService.Rebook(ctx, trainticket.RebookInfo{
		OrderId:  "wf_rebook_ord1",
		TripId:   "G_WF_001",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "haven't paid")
}

// TestRebookWorkflow_RejectPaidOrder verifies that a PAID order is rejected with
// "not suitable to rebook" (the first status guard in Rebook checks status==1==PAID)
func TestRebookWorkflow_RejectPaidOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_rebook_ord2",
		Status: trainticket.ORDER_STATUS_PAID,
	})
	assert.NoError(t, err)

	err = env.rebookService.Rebook(ctx, trainticket.RebookInfo{
		OrderId:  "wf_rebook_ord2",
		TripId:   "G_WF_002",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not suitable")
}

// TestRebookWorkflow_RejectCollectedOrder verifies that a COLLECTED order cannot be rebooked
func TestRebookWorkflow_RejectCollectedOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_rebook_ord3",
		Status: trainticket.ORDER_STATUS_COLLECTED,
	})
	assert.NoError(t, err)

	err = env.rebookService.Rebook(ctx, trainticket.RebookInfo{
		OrderId:  "wf_rebook_ord3",
		TripId:   "G_WF_003",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already collected")
}

// TestRebookWorkflow_RejectUsedOrder verifies that a USED order cannot be rebooked
func TestRebookWorkflow_RejectUsedOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_rebook_ord4",
		Status: trainticket.ORDER_STATUS_USED,
	})
	assert.NoError(t, err)

	err = env.rebookService.Rebook(ctx, trainticket.RebookInfo{
		OrderId:  "wf_rebook_ord4",
		TripId:   "G_WF_004",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can't change")
}

// TestRebookWorkflow_RejectCanceledOrder verifies that a CANCELED order cannot be rebooked
func TestRebookWorkflow_RejectCanceledOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_rebook_ord5",
		Status: trainticket.ORDER_STATUS_CANCELED,
	})
	assert.NoError(t, err)

	err = env.rebookService.Rebook(ctx, trainticket.RebookInfo{
		OrderId:  "wf_rebook_ord5",
		TripId:   "G_WF_005",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can't change")
}

// TestRebookWorkflow_RejectChangeStatusOrder verifies that a CHANGE-status order cannot be rebooked
func TestRebookWorkflow_RejectChangeStatusOrder(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_rebook_ord6",
		Status: trainticket.ORDER_STATUS_CHANGE,
	})
	assert.NoError(t, err)

	err = env.rebookService.Rebook(ctx, trainticket.RebookInfo{
		OrderId:  "wf_rebook_ord6",
		TripId:   "G_WF_006",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can't change")
}

// TestRebookWorkflow_PayDifference_NotPaid verifies that PayDifference for a NOT_PAID order
// returns immediately with no error (special early-exit path in PayDifference)
func TestRebookWorkflow_PayDifference_NotPaid(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.orderService.CreateNewOrder(ctx, trainticket.Order{
		ID:     "wf_paydiff_ord1",
		Status: trainticket.ORDER_STATUS_NOT_PAID,
	})
	assert.NoError(t, err)

	err = env.rebookService.PayDifference(ctx, trainticket.RebookInfo{
		OrderId:  "wf_paydiff_ord1",
		TripId:   "G_WF_DIFF_001",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.NoError(t, err)
}

// TestRebookWorkflow_PayDifference_OrderNotFound verifies that PayDifference for a
// non-existent order returns an error
func TestRebookWorkflow_PayDifference_OrderNotFound(t *testing.T) {
	ctx := context.Background()
	env, err := newRebookWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.rebookService.PayDifference(ctx, trainticket.RebookInfo{
		OrderId:  "nonexistent_order",
		TripId:   "G_WF_DIFF_002",
		SeatType: 2,
		Date:     "2099-01-01",
	})
	assert.Error(t, err)
}

// ============================================================
// InsidePayment workflow tests
// ============================================================

// TestInsidePaymentWorkflow_CreateAccountAndQuery verifies that CreateAccount populates
// the balance and QueryAccount reflects it
func TestInsidePaymentWorkflow_CreateAccountAndQuery(t *testing.T) {
	ctx := context.Background()
	env, err := newInsidePaymentWorkflowEnv(ctx)
	assert.NoError(t, err)

	ok, err := env.insidePaymentService.CreateAccount(ctx, trainticket.AccountInfo{
		UserId: "wf_pay_user1",
		Money:  "500",
	})
	assert.NoError(t, err)
	assert.True(t, ok)

	balance, err := env.insidePaymentService.QueryAccount(ctx, "wf_pay_user1")
	assert.NoError(t, err)
	assert.Equal(t, "500", balance)
}

// TestInsidePaymentWorkflow_DuplicateAccountRejected verifies that creating an account
// for a user that already has one returns false without error
func TestInsidePaymentWorkflow_DuplicateAccountRejected(t *testing.T) {
	ctx := context.Background()
	env, err := newInsidePaymentWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.insidePaymentService.CreateAccount(ctx, trainticket.AccountInfo{UserId: "wf_pay_user2", Money: "100"})
	assert.NoError(t, err)

	ok, err := env.insidePaymentService.CreateAccount(ctx, trainticket.AccountInfo{UserId: "wf_pay_user2", Money: "200"})
	assert.NoError(t, err)
	assert.False(t, ok)
}

// TestInsidePaymentWorkflow_AddMoneyIncreasesBalance verifies that AddMoney appends a
// new money record, and QueryAccount returns the sum of all records
func TestInsidePaymentWorkflow_AddMoneyIncreasesBalance(t *testing.T) {
	ctx := context.Background()
	env, err := newInsidePaymentWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.insidePaymentService.CreateAccount(ctx, trainticket.AccountInfo{UserId: "wf_pay_user3", Money: "100"})
	assert.NoError(t, err)

	err = env.insidePaymentService.AddMoney(ctx, "wf_pay_user3", "200")
	assert.NoError(t, err)

	balance, err := env.insidePaymentService.QueryAccount(ctx, "wf_pay_user3")
	assert.NoError(t, err)
	assert.Equal(t, "300", balance)
}

// TestInsidePaymentWorkflow_MultipleAddMoneyOperations verifies that multiple AddMoney
// calls accumulate correctly in the balance
func TestInsidePaymentWorkflow_MultipleAddMoneyOperations(t *testing.T) {
	ctx := context.Background()
	env, err := newInsidePaymentWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.insidePaymentService.CreateAccount(ctx, trainticket.AccountInfo{UserId: "wf_pay_user4", Money: "50"})
	assert.NoError(t, err)

	for _, amount := range []string{"100", "150", "200"} {
		err = env.insidePaymentService.AddMoney(ctx, "wf_pay_user4", amount)
		assert.NoError(t, err)
	}

	balance, err := env.insidePaymentService.QueryAccount(ctx, "wf_pay_user4")
	assert.NoError(t, err)
	assert.Equal(t, "500", balance) // 50 + 100 + 150 + 200
}

// TestInsidePaymentWorkflow_DrawbackDoesNotError verifies that Drawback on an existing
// account record succeeds without error
func TestInsidePaymentWorkflow_DrawbackDoesNotError(t *testing.T) {
	ctx := context.Background()
	env, err := newInsidePaymentWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.insidePaymentService.CreateAccount(ctx, trainticket.AccountInfo{UserId: "wf_pay_user5", Money: "1000"})
	assert.NoError(t, err)

	err = env.insidePaymentService.Drawback(ctx, "wf_pay_user5", "200")
	assert.NoError(t, err)
}

// TestInsidePaymentWorkflow_EmptyAccountQuery verifies that querying balance for a user
// with no account records returns "0"
func TestInsidePaymentWorkflow_EmptyAccountQuery(t *testing.T) {
	ctx := context.Background()
	env, err := newInsidePaymentWorkflowEnv(ctx)
	assert.NoError(t, err)

	balance, err := env.insidePaymentService.QueryAccount(ctx, "nonexistent_user")
	assert.NoError(t, err)
	assert.Equal(t, "0", balance)
}

// TestInsidePaymentWorkflow_QueryPaymentsEmpty verifies that querying an empty payment
// store returns a non-nil (possibly empty) slice without error
func TestInsidePaymentWorkflow_QueryPaymentsEmpty(t *testing.T) {
	ctx := context.Background()
	env, err := newInsidePaymentWorkflowEnv(ctx)
	assert.NoError(t, err)

	payments, err := env.insidePaymentService.QueryPayment(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, payments)
}

// ============================================================
// Preserve workflow tests
// ============================================================

// TestPreserveWorkflow_ContactNotFound verifies that Preserve returns an error when
// the referenced contact does not exist in the contacts store
func TestPreserveWorkflow_ContactNotFound(t *testing.T) {
	ctx := context.Background()
	env, err := newPreserveWorkflowEnv(ctx)
	assert.NoError(t, err)

	_, err = env.preserveService.Preserve(ctx, trainticket.OrderTicketsInfo{
		AccountID:  "wf_pres_user1",
		ContactsID: "nonexistent_contact",
		TripID:     "G_WF_PRES_001",
		SeatType:   2,
		Date:       "2099-01-01",
		From:       "shanghai",
		To:         "beijing",
	})
	assert.Error(t, err)
}

// TestPreserveWorkflow_ContactFoundTripMissing verifies that when a contact exists but
// the requested trip is not registered, Preserve silently discards the GetTripAllDetailInfo
// error and returns an empty Order with no error. No actual order is created
func TestPreserveWorkflow_ContactFoundTripMissing(t *testing.T) {
	ctx := context.Background()
	env, err := newPreserveWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.contactsService.CreateContacts(ctx, trainticket.Contact{
		ID:           "wf_pres_contact1",
		AccountID:    "wf_pres_user1",
		Name:         "Alice",
		DocumentType: int(trainticket.ID_CARD),
	})
	assert.NoError(t, err)

	order, err := env.preserveService.Preserve(ctx, trainticket.OrderTicketsInfo{
		AccountID:  "wf_pres_user1",
		ContactsID: "wf_pres_contact1",
		TripID:     "G_WF_PRES_001",
		SeatType:   2,
		Date:       "2099-01-01",
		From:       "shanghai",
		To:         "beijing",
	})
	// Preserve swallows the trip-not-found error and returns an empty Order
	assert.NoError(t, err)
	assert.Empty(t, order.ID)
}

// TestPreserveWorkflow_TripExistsButDateInvalid verifies that when a trip is registered but
// the departure date fails internal validation, Preserve similarly returns an empty Order
// with no error (the internal error from GetTripAllDetailInfo is swallowed)
func TestPreserveWorkflow_TripExistsButDateInvalid(t *testing.T) {
	ctx := context.Background()
	env, err := newPreserveWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.contactsService.CreateContacts(ctx, trainticket.Contact{
		ID:           "wf_pres_contact2",
		AccountID:    "wf_pres_user2",
		Name:         "Bob",
		DocumentType: int(trainticket.ID_CARD),
	})
	assert.NoError(t, err)

	// Register the trip so GetTripAllDetailInfo finds it, but uses an invalid date
	// so getTickets' afterToday check fails internally
	_, err = env.travelService.CreateTrip(ctx, trainticket.TravelInfo{
		TripID:              "G_WF_PRES_002",
		TrainTypeName:       "GaoTie",
		RouteID:             "route_pres_001",
		StartStationName:    "shanghai",
		TerminalStationName: "beijing",
		StartTime:           "08:00:00",
		EndTime:             "13:00:00",
	})
	assert.NoError(t, err)

	order, err := env.preserveService.Preserve(ctx, trainticket.OrderTicketsInfo{
		AccountID:  "wf_pres_user2",
		ContactsID: "wf_pres_contact2",
		TripID:     "G_WF_PRES_002",
		SeatType:   2,
		Date:       "2099-01-01",
		From:       "shanghai",
		To:         "beijing",
	})
	// The departure date validation fails inside GetTripAllDetailInfo, but Preserve
	// swallows that error too, returning an empty Order
	assert.NoError(t, err)
	assert.Empty(t, order.ID)
}

// TestPreserveWorkflow_ContactCreateAndLookup verifies that a contact can be created
// and immediately retrieved by ID with all fields intact
func TestPreserveWorkflow_ContactCreateAndLookup(t *testing.T) {
	ctx := context.Background()
	env, err := newPreserveWorkflowEnv(ctx)
	assert.NoError(t, err)

	contact := trainticket.Contact{
		ID:             "wf_pres_contact3",
		AccountID:      "wf_pres_user3",
		Name:           "Charlie",
		DocumentType:   int(trainticket.ID_CARD),
		DocumentNumber: "CN123456",
		PhoneNumber:    "13800138000",
	}
	err = env.contactsService.CreateContacts(ctx, contact)
	assert.NoError(t, err)

	found, err := env.contactsService.FindContactsById(ctx, "wf_pres_contact3")
	assert.NoError(t, err)
	assert.Equal(t, "Charlie", found.Name)
	assert.Equal(t, "CN123456", found.DocumentNumber)
	assert.Equal(t, "wf_pres_user3", found.AccountID)
}

// TestPreserveWorkflow_MultipleContactsForSameAccount verifies that multiple contacts
// can be associated with the same account and retrieved individually
func TestPreserveWorkflow_MultipleContactsForSameAccount(t *testing.T) {
	ctx := context.Background()
	env, err := newPreserveWorkflowEnv(ctx)
	assert.NoError(t, err)

	contacts := []trainticket.Contact{
		{ID: "wf_mc_contact1", AccountID: "wf_mc_user1", Name: "Passenger A", DocumentNumber: "DOC001"},
		{ID: "wf_mc_contact2", AccountID: "wf_mc_user1", Name: "Passenger B", DocumentNumber: "DOC002"},
	}
	for _, c := range contacts {
		err = env.contactsService.CreateContacts(ctx, c)
		assert.NoError(t, err)
	}

	all, err := env.contactsService.FindContactsByAccountId(ctx, "wf_mc_user1")
	assert.NoError(t, err)
	assert.Len(t, all, 2)

	c1, err := env.contactsService.FindContactsById(ctx, "wf_mc_contact1")
	assert.NoError(t, err)
	assert.Equal(t, "Passenger A", c1.Name)

	c2, err := env.contactsService.FindContactsById(ctx, "wf_mc_contact2")
	assert.NoError(t, err)
	assert.Equal(t, "Passenger B", c2.Name)
}

// TestPreserveWorkflow_UserAndTravelServiceIntegration verifies that the user and travel
// service dependencies are correctly wired: creating a user and trip independently,
// then confirming both can be fetched
func TestPreserveWorkflow_UserAndTravelServiceIntegration(t *testing.T) {
	ctx := context.Background()
	env, err := newPreserveWorkflowEnv(ctx)
	assert.NoError(t, err)

	err = env.userService.SaveUser(ctx, trainticket.User{
		UserID:   "wf_int_user1",
		Username: "eve",
		Email:    "eve@example.com",
	})
	assert.NoError(t, err)

	user, err := env.userService.FindByUserID(ctx, "wf_int_user1")
	assert.NoError(t, err)
	assert.Equal(t, "eve", user.Username)
	assert.Equal(t, "eve@example.com", user.Email)

	_, err = env.travelService.CreateTrip(ctx, trainticket.TravelInfo{
		TripID:              "G_INT_001",
		TrainTypeName:       "GaoTie",
		RouteID:             "route_int_001",
		StartStationName:    "shanghai",
		TerminalStationName: "beijing",
		StartTime:           "08:00:00",
		EndTime:             "13:00:00",
	})
	assert.NoError(t, err)

	trip, err := env.travelService.Retrieve(ctx, "G_INT_001")
	assert.NoError(t, err)
	assert.Equal(t, "G_INT_001", trip.TripID)
	assert.Equal(t, "GaoTie", trip.TrainTypeName)
}
