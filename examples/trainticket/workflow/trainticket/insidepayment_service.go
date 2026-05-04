package trainticket

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type InsidePaymentService interface {
	Pay(ctx context.Context, info PaymentInfo) (bool, error)
	CreateAccount(ctx context.Context, info AccountInfo) (bool, error)
	AddMoney(ctx context.Context, userId string, money string) error
	QueryAccount(ctx context.Context, userId string) (string, error)
	QueryAddMoney(ctx context.Context) ([]InsideMoney, error)
	QueryPayment(ctx context.Context) ([]Payment, error)
	PayDifference(ctx context.Context, info PaymentInfo) error
	Drawback(ctx context.Context, userID string, money string) error
}

type InsidePaymentServiceImpl struct {
	insidePaymentDB backend.NoSQLDatabase
	paymentService  PaymentService
	orderService    OrderService
}

func NewInsidePaymentServiceImpl(ctx context.Context, insidePaymentDB backend.NoSQLDatabase, paymentService PaymentService, orderService OrderService) (InsidePaymentService, error) {
	return &InsidePaymentServiceImpl{insidePaymentDB: insidePaymentDB, paymentService: paymentService, orderService: orderService}, nil
}

func (s *InsidePaymentServiceImpl) PayDifference(ctx context.Context, info PaymentInfo) error {
	userId := info.UserId
	payment := Payment{
		ID:      uuid.New().String(),
		OrderID: info.OrderId,
		Price:   info.Price,
		UserID:  info.UserId,
	}
	payments, err := s.findPaymentByUserId(ctx, userId)
	if err != nil {
		return err
	}
	addMonies, err := s.findMoneyByUserId(ctx, userId)
	if err != nil {
		return err
	}
	totalExpand := 0
	for _, p := range payments {
		v, _ := strconv.Atoi(p.Price)
		totalExpand += v
	}
	v, _ := strconv.Atoi(info.Price)
	totalExpand += v
	money := 0
	for _, m := range addMonies {
		v, _ := strconv.Atoi(m.Money)
		money += v
	}

	if totalExpand > money {
		outsidePaymentInfo := Payment{
			OrderID: info.OrderId,
			UserID:  userId,
			Price:   info.Price,
		}
		err = s.paymentService.Pay(ctx, outsidePaymentInfo)
		if err != nil {
			return fmt.Errorf("pay difference failed: %w", err)
		}
		payment.PaymentType = PaymentType_E
		s.savePayment(ctx, payment)
	} else {
		payment.PaymentType = PaymentType_E
		s.savePayment(ctx, payment)
	}
	return nil
}

func (s *InsidePaymentServiceImpl) Drawback(ctx context.Context, userID string, money string) error {
	collection, err := s.insidePaymentDB.GetCollection(ctx, "inside_payment_db", "money")
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "UserID", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "Money", Value: money}, {Key: "Type", Value: INSIDE_MONEY_TYPE_DRAWBACK}}}}
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *InsidePaymentServiceImpl) Pay(ctx context.Context, info PaymentInfo) (bool, error) {
	userId := info.UserId
	var order Order
	if strings.HasPrefix(info.TripId, "G") || strings.HasPrefix("D", info.TripId) {
		var err error
		order, err = s.orderService.GetOrderById(ctx, info.OrderId)
		if err != nil {
			return false, err
		}
	}
	if order.Status == ORDER_STATUS_NOT_PAID {
		return false, fmt.Errorf("order status not allowed to pay")
	}

	payment := Payment{
		ID:      uuid.New().String(),
		OrderID: info.OrderId,
		Price:   order.Price,
		UserID:  userId,
	}

	payments, err := s.findPaymentByUserId(ctx, userId)
	if err != nil {
		return false, err
	}
	addMonies, err := s.findMoneyByUserId(ctx, userId)
	if err != nil {
		return false, err
	}
	totalExpand := 0
	for _, p := range payments {
		v, _ := strconv.Atoi(p.Price)
		totalExpand += v
	}
	money := 0
	for _, m := range addMonies {
		v, _ := strconv.Atoi(m.Money)
		money += v
	}

	if totalExpand > money {
		outsidePaymentInfo := Payment{
			OrderID: info.OrderId,
			UserID:  userId,
			Price:   info.Price,
		}
		err = s.paymentService.Pay(ctx, outsidePaymentInfo)
		if err != nil {
			return false, fmt.Errorf("pay failed: %w", err)
		}
		payment.PaymentType = PaymentType_O
		s.savePayment(ctx, payment)
		s.orderService.ModifyOrder(ctx, info.OrderId, ORDER_STATUS_PAID)
	} else {
		s.orderService.ModifyOrder(ctx, info.OrderId, ORDER_STATUS_PAID)
		payment.PaymentType = PaymentType_P
		s.savePayment(ctx, payment)
	}
	return true, nil
}

func (s *InsidePaymentServiceImpl) CreateAccount(ctx context.Context, info AccountInfo) (bool, error) {
	moneys, err := s.findMoneyByUserId(ctx, info.UserId)
	if err != nil {
		return false, err
	}
	if len(moneys) == 0 {
		addMoney := InsideMoney{
			UserID: info.UserId,
			Money:  info.Money,
			Type:   MoneyType_A,
		}
		err = s.saveMoney(ctx, addMoney)
		if err != nil {
			return false, fmt.Errorf("add money failed for user id (%s)", info.UserId)
		}
		return true, nil
	}
	return false, nil
}

func (s *InsidePaymentServiceImpl) AddMoney(ctx context.Context, userId string, money string) error {
	_, err := s.findMoneyByUserId(ctx, userId)
	if err != nil {
		return err
	}
	addMoney := InsideMoney{
		UserID: userId,
		Money:  money,
		Type:   MoneyType_A,
	}
	err = s.saveMoney(ctx, addMoney)
	if err != nil {
		return fmt.Errorf("add money failed for user id (%s)", userId)
	}
	return nil
}

func (s *InsidePaymentServiceImpl) QueryAccount(ctx context.Context, userId string) (string, error) {
	payments, err := s.findPaymentByUserId(ctx, userId)
	if err != nil {
		return "", err
	}
	addMonies, err := s.findMoneyByUserId(ctx, userId)
	if err != nil {
		return "", err
	}
	totalExpand := 0
	for _, p := range payments {
		v, _ := strconv.Atoi(p.Price)
		totalExpand += v
	}
	money := 0
	for _, m := range addMonies {
		v, _ := strconv.Atoi(m.Money)
		money += v
	}

	return strconv.Itoa(money - totalExpand), nil
}

func (s *InsidePaymentServiceImpl) QueryPayment(ctx context.Context) ([]Payment, error) {
	return s.findAllPayments(ctx)
}

func (s *InsidePaymentServiceImpl) QueryAddMoney(ctx context.Context) ([]InsideMoney, error) {
	return s.findAllMoneys(ctx)
}

func (s *InsidePaymentServiceImpl) findAllPayments(ctx context.Context) ([]Payment, error) {
	collection, err := s.insidePaymentDB.GetCollection(ctx, "inside_payment_db", "payment")
	if err != nil {
		return nil, err
	}

	cursor, err := collection.FindMany(ctx, nil)
	if err != nil {
		return nil, err
	}
	var payments []Payment
	err = cursor.All(ctx, &payments)
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (s *InsidePaymentServiceImpl) findAllMoneys(ctx context.Context) ([]InsideMoney, error) {
	collection, err := s.insidePaymentDB.GetCollection(ctx, "inside_payment_db", "money")
	if err != nil {
		return nil, err
	}

	cursor, err := collection.FindMany(ctx, nil)
	if err != nil {
		return nil, err
	}
	var money []InsideMoney
	err = cursor.All(ctx, &money)
	if err != nil {
		return nil, err
	}
	return money, nil
}

func (s *InsidePaymentServiceImpl) findPaymentByUserId(ctx context.Context, userID string) ([]Payment, error) {
	collection, err := s.insidePaymentDB.GetCollection(ctx, "inside_payment_db", "payment")
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "UserID", Value: userID}}
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	var payments []Payment
	err = cursor.All(ctx, &payments)
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (s *InsidePaymentServiceImpl) findMoneyByUserId(ctx context.Context, userID string) ([]Money, error) {
	collection, err := s.insidePaymentDB.GetCollection(ctx, "inside_payment_db", "money")
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "UserID", Value: userID}}
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	var moneys []Money
	err = cursor.All(ctx, &moneys)
	if err != nil {
		return nil, err
	}
	return moneys, nil
}

func (s *InsidePaymentServiceImpl) savePayment(ctx context.Context, payment Payment) error {
	collection, err := s.insidePaymentDB.GetCollection(ctx, "inside_payment_db", "payment")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, payment)
}

func (s *InsidePaymentServiceImpl) saveMoney(ctx context.Context, money InsideMoney) error {
	collection, err := s.insidePaymentDB.GetCollection(ctx, "inside_payment_db", "money")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, money)
}
