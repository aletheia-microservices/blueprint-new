package trainticket

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type FoodService interface {
	// Find all food order
	CreateFoodOrder(ctx context.Context, addFoodOrder FoodOrder) (FoodOrder, error)
	CreateFoodBatches(ctx context.Context, foodOrderList []FoodOrder) error
	UpdateFoodOrder(ctx context.Context, updateFoodOrder FoodOrder) error
	DeleteFoodOrder(ctx context.Context, orderID string) error
	FindFoodOrderByOrderId(ctx context.Context, orderID string) (FoodOrder, error)
	GetAllFood(ctx context.Context, date string, startStation string, endStation string, tripId string) ([]Food, map[string]StationFoodStore, error)
}

type FoodServiceImpl struct {
	foodDB             backend.NoSQLDatabase
	deliveryQueue      backend.Queue
	trainFoodService   TrainFoodService
	travelService      TravelService
	stationFoodService StationFoodService
}

func NewFoodServiceImpl(ctx context.Context, foodDB backend.NoSQLDatabase, foodOrderQueue backend.Queue, trainFoodService TrainFoodService, travelService TravelService, stationFoodService StationFoodService) (FoodService, error) {
	return &FoodServiceImpl{foodDB: foodDB, deliveryQueue: foodOrderQueue, trainFoodService: trainFoodService, travelService: travelService, stationFoodService: stationFoodService}, nil
}

func (c *FoodServiceImpl) CreateFoodOrder(ctx context.Context, addFoodOrder FoodOrder) (FoodOrder, error) {
	foodOrder := FoodOrder{
		ID:          uuid.New().String(),
		OrderID:     addFoodOrder.OrderID,
		FoodType:    addFoodOrder.FoodType,
		StationName: addFoodOrder.StationName,
		StoreName:   addFoodOrder.StoreName,
		FoodName:    addFoodOrder.FoodName,
		Price:       addFoodOrder.Price,
	}

	collection, err := c.foodDB.GetCollection(ctx, "food_db", "food_order")
	if err != nil {
		return FoodOrder{}, err
	}
	err = collection.InsertOne(ctx, foodOrder)
	if err != nil {
		return FoodOrder{}, err
	}

	delivery := Delivery{
		OrderID:     addFoodOrder.OrderID,
		FoodName:    addFoodOrder.FoodName,
		StationName: addFoodOrder.StationName,
		StoreName:   addFoodOrder.StoreName,
	}

	_, err = c.deliveryQueue.Push(ctx, delivery)
	if err != nil {
		return FoodOrder{}, err
	}

	return foodOrder, nil
}

func (c *FoodServiceImpl) CreateFoodBatches(ctx context.Context, foodOrderList []FoodOrder) error {
	collection, err := c.foodDB.GetCollection(ctx, "food_db", "food_order")
	if err != nil {
		return err
	}

	for _, order := range foodOrderList {
		filter := bson.D{{Key: "OrderID", Value: order.OrderID}}
		cursor, err := collection.FindOne(ctx, filter)
		if err != nil {
			return err
		}

		var foodOrder FoodOrder
		ok, err := cursor.One(ctx, &foodOrder)
		if err != nil {
			return err
		}
		if ok {
			return fmt.Errorf("food order already exists for order ID (%s)", order.OrderID)
		}
	}

	for _, order := range foodOrderList {
		foodOrder := FoodOrder{
			ID:          uuid.New().String(),
			OrderID:     order.OrderID,
			FoodType:    order.FoodType,
			StationName: order.StationName,
			StoreName:   order.StoreName,
			FoodName:    order.FoodName,
			Price:       order.Price,
		}
		err = collection.InsertOne(ctx, foodOrder)
		if err != nil {
			return err
		}

		delivery := Delivery{
			OrderID:     foodOrder.OrderID,
			FoodName:    foodOrder.FoodName,
			StationName: foodOrder.StationName,
			StoreName:   foodOrder.StoreName,
		}

		_, err = c.deliveryQueue.Push(ctx, delivery)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *FoodServiceImpl) UpdateFoodOrder(ctx context.Context, updateFoodOrder FoodOrder) error {
	collection, err := c.foodDB.GetCollection(ctx, "food_db", "food_order")
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "OrderID", Value: updateFoodOrder.OrderID}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return err
	}

	var foodOrder FoodOrder
	ok, err := cursor.One(ctx, &foodOrder)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("food order not found for orderID (%s)", updateFoodOrder.OrderID)
	}

	foodOrder.FoodName = updateFoodOrder.FoodName
	foodOrder.Price = updateFoodOrder.Price
	err = collection.InsertOne(ctx, foodOrder)
	if err != nil {
		return err
	}
	return nil
}

func (c *FoodServiceImpl) DeleteFoodOrder(ctx context.Context, orderID string) error {
	collection, err := c.foodDB.GetCollection(ctx, "food_db", "food_order")
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "OrderID", Value: orderID}}
	return collection.DeleteOne(ctx, filter)
}

func (c *FoodServiceImpl) FindFoodOrderByOrderId(ctx context.Context, orderID string) (FoodOrder, error) {
	collection, err := c.foodDB.GetCollection(ctx, "food_db", "food_order")
	if err != nil {
		return FoodOrder{}, err
	}

	filter := bson.D{{Key: "OrderID", Value: orderID}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return FoodOrder{}, err
	}

	var foodOrder FoodOrder
	ok, err := cursor.One(ctx, &foodOrder)
	if err != nil {
		return FoodOrder{}, err
	}
	if !ok {
		return FoodOrder{}, fmt.Errorf("food order not found for orderID (%s)", orderID)
	}
	return foodOrder, nil
}

func (c *FoodServiceImpl) GetAllFood(ctx context.Context, date string, startStation string, endStation string, tripId string) ([]Food, map[string]StationFoodStore, error) {
	if len(tripId) <= 2 {
		return nil, nil, fmt.Errorf("trip id (%s) is not suitable", tripId)
	}

	var foodStoreListMap = make(map[string]StationFoodStore)
	trainFoodList, err := c.trainFoodService.ListTrainFoodByTripID(ctx, tripId)
	if err != nil {
		return nil, nil, err
	}

	route, err := c.travelService.GetRouteByTripId(ctx, tripId)
	if err != nil {
		return nil, nil, err
	}

	var stations = route.Stations
	for i, station := range route.Stations {
		if station == startStation {
			break
		} else {
			// remove
			stations = append(stations[:i], stations[i+1:]...)
		}
	}
	for i, station := range route.Stations {
		if station == endStation {
			break
		} else {
			// remove
			stations = append(stations[:i], stations[i+1:]...)
		}
	}

	stationFoodStores, err := c.stationFoodService.ListFoodStores(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, station := range stations {
		for _, stationFoodStore := range stationFoodStores {
			if stationFoodStore.StationName == station {
				foodStoreListMap[station] = stationFoodStore
			}
		}
	}

	return trainFoodList, foodStoreListMap, nil
}
