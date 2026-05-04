package basket

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type BasketService interface {
	GetBasket(ctx context.Context, query GetBasketRequest) (CustomerBasketResponse, error)
	UpdateBasket(ctx context.Context, request UpdateBasketRequest) (CustomerBasketResponse, error)
	DeleteBasket(ctx context.Context, query DeleteBasketRequest) (DeleteBasketResponse, error)
}

type BasketServiceImpl struct {
	database backend.NoSQLDatabase
}

func NewBasketServiceImpl(ctx context.Context, database backend.NoSQLDatabase) (BasketService, error) {
	s := &BasketServiceImpl{
		database: database,
	}
	return s, nil
}

func (s *BasketServiceImpl) UpdateBasket(ctx context.Context, command UpdateBasketRequest) (CustomerBasketResponse, error) {
	err := s.storeBasket(ctx, command.Cart)
	if err != nil {
		return CustomerBasketResponse{}, err
	}
	return CustomerBasketResponse{Cart: command.Cart}, nil
}

func (s *BasketServiceImpl) GetBasket(ctx context.Context, query GetBasketRequest) (CustomerBasketResponse, error) {
	basket, err := s.getBasket(ctx, query.UserName)
	if err != nil {
		return CustomerBasketResponse{}, err
	}
	return CustomerBasketResponse{Cart: basket}, nil
}

func (s *BasketServiceImpl) DeleteBasket(ctx context.Context, query DeleteBasketRequest) (DeleteBasketResponse, error) {
	err := s.deleteBasket(ctx, query.UserName)
	if err != nil {
		return DeleteBasketResponse{IsSuccess: false}, err
	}
	return DeleteBasketResponse{IsSuccess: true}, nil
}

func (s *BasketServiceImpl) getBasket(ctx context.Context, username string) (CustomerBasket, error) {
	collection, err := s.database.GetCollection(ctx, "basket_db", "basket")
	if err != nil {
		return CustomerBasket{}, err
	}
	filter := bson.D{{Key: "UserName", Value: username}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return CustomerBasket{}, err
	}
	var cart CustomerBasket
	ok, err := cursor.One(ctx, &cart)
	if err != nil {
		return CustomerBasket{}, err
	}
	if !ok {
		return CustomerBasket{}, fmt.Errorf("basket not found for username (%s)", username)
	}
	return cart, nil
}

func (s *BasketServiceImpl) storeBasket(ctx context.Context, basket CustomerBasket) error {
	collection, err := s.database.GetCollection(ctx, "basket_db", "basket")
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "UserName", Value: basket.UserName}}
	updated, err := collection.ReplaceOne(ctx, filter, basket)
	if err != nil {
		return err
	}
	if updated == 0 {
		return collection.InsertOne(ctx, basket)
	}
	return nil
}

func (s *BasketServiceImpl) deleteBasket(ctx context.Context, username string) error {
	collection, err := s.database.GetCollection(ctx, "basket_db", "basket")
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "UserName", Value: username}}
	return collection.DeleteOne(ctx, filter)
}
