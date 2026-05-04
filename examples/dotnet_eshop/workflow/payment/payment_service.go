package payment

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"

	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/order"
)

type PaymentService interface {
	Run(ctx context.Context) error
}

type PaymentServiceImpl struct {
	stockConfirmedQueue backend.Queue
	paymentResultQueue  backend.Queue
}

func NewPaymentServiceImpl(ctx context.Context, stockConfirmedQueue backend.Queue, paymentResultQueue backend.Queue) (PaymentService, error) {
	s := &PaymentServiceImpl{
		stockConfirmedQueue: stockConfirmedQueue,
		paymentResultQueue:  paymentResultQueue,
	}
	return s, nil
}

// Run listens for OrderStatusChangedToStockConfirmedEvents, processes payment,
// and publishes an OrderPaymentResultEvent (always succeeds in this implementation).
func (s *PaymentServiceImpl) Run(ctx context.Context) error {
	for {
		var event order.OrderStatusChangedToStockConfirmedEvent
		ok, err := s.stockConfirmedQueue.Pop(ctx, &event)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		result := order.OrderPaymentResultEvent{
			OrderId:   event.OrderId,
			Succeeded: true,
		}
		_, err = s.paymentResultQueue.Push(ctx, result)
		if err != nil {
			return err
		}
	}
}
