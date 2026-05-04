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

var foodServiceRegistry = registry.NewServiceRegistry[trainticket.FoodService]("food_service")

func init() {
	foodServiceRegistry.Register("local", func(ctx context.Context) (trainticket.FoodService, error) {
		trainFoodDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		trainFoodService, err := trainticket.NewTrainFoodServiceImpl(ctx, trainFoodDB)
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
		queue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		return trainticket.NewFoodServiceImpl(ctx, foodDB, queue, trainFoodService, travelService, stationFoodService)
	})
}

func TestFoodServiceCreateAndFind(t *testing.T) {
	ctx := context.Background()
	service, err := foodServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	foodOrder := trainticket.FoodOrder{
		OrderID:     "food_ord001",
		FoodType:    1,
		FoodName:    "Sandwich",
		StationName: "shanghai",
		StoreName:   "ShangHai Store",
		Price:       15.0,
	}
	created, err := service.CreateFoodOrder(ctx, foodOrder)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "food_ord001", created.OrderID)

	found, err := service.FindFoodOrderByOrderId(ctx, "food_ord001")
	assert.NoError(t, err)
	assert.NotEmpty(t, found.ID)
	assert.Equal(t, "Sandwich", found.FoodName)
}

func TestFoodServiceUpdateFoodOrder(t *testing.T) {
	ctx := context.Background()
	service, err := foodServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	foodOrder := trainticket.FoodOrder{
		OrderID:  "food_ord002",
		FoodType: 1,
		FoodName: "Noodles",
		Price:    20.0,
	}
	_, err = service.CreateFoodOrder(ctx, foodOrder)
	assert.NoError(t, err)

	foodOrder.FoodName = "Rice"
	foodOrder.Price = 18.0
	err = service.UpdateFoodOrder(ctx, foodOrder)
	assert.NoError(t, err)
}

func TestFoodServiceDeleteFoodOrder(t *testing.T) {
	ctx := context.Background()
	service, err := foodServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	foodOrder := trainticket.FoodOrder{
		OrderID:  "food_ord003",
		FoodType: 1,
		FoodName: "Dumpling",
		Price:    12.0,
	}
	_, err = service.CreateFoodOrder(ctx, foodOrder)
	assert.NoError(t, err)

	err = service.DeleteFoodOrder(ctx, "food_ord003")
	assert.NoError(t, err)

	_, err = service.FindFoodOrderByOrderId(ctx, "food_ord003")
	assert.Error(t, err)
}

func TestFoodServiceCreateFoodBatches(t *testing.T) {
	ctx := context.Background()
	service, err := foodServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	orders := []trainticket.FoodOrder{
		{OrderID: "food_batch001", FoodType: 1, FoodName: "Tea", Price: 5.0},
		{OrderID: "food_batch002", FoodType: 1, FoodName: "Coffee", Price: 8.0},
	}
	err = service.CreateFoodBatches(ctx, orders)
	assert.NoError(t, err)

	found, err := service.FindFoodOrderByOrderId(ctx, "food_batch001")
	assert.NoError(t, err)
	assert.Equal(t, "Tea", found.FoodName)
}
