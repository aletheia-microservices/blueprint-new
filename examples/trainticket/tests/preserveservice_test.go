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

var preserveServiceRegistry = registry.NewServiceRegistry[trainticket.PreserveService]("preserve_service")
var preserveContactsDB *simplenosqldb.SimpleNoSQLDB
var preserveUserDB *simplenosqldb.SimpleNoSQLDB
var preserveTravelDB *simplenosqldb.SimpleNoSQLDB
var preserveStationDB *simplenosqldb.SimpleNoSQLDB
var preserveTrainDB *simplenosqldb.SimpleNoSQLDB
var preserveRouteDB *simplenosqldb.SimpleNoSQLDB
var preservePriceDB *simplenosqldb.SimpleNoSQLDB
var preserveConfigDB *simplenosqldb.SimpleNoSQLDB

func init() {
	preserveServiceRegistry.Register("local", func(ctx context.Context) (trainticket.PreserveService, error) {
		assuranceDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		assuranceService, err := trainticket.NewAssuranceServiceImpl(ctx, assuranceDB)
		if err != nil {
			return nil, err
		}

		preserveStationDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		stationService, err := trainticket.NewStationServiceImpl(ctx, preserveStationDB)
		if err != nil {
			return nil, err
		}
		preserveTrainDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		trainService, err := trainticket.NewTrainServiceImpl(ctx, preserveTrainDB)
		if err != nil {
			return nil, err
		}
		preserveRouteDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		routeService, err := trainticket.NewRouteServiceImpl(ctx, preserveRouteDB)
		if err != nil {
			return nil, err
		}
		preservePriceDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		priceService, err := trainticket.NewPriceServiceImpl(ctx, preservePriceDB)
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

		preserveConfigDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		configService, err := trainticket.NewConfigServiceImpl(ctx, preserveConfigDB)
		if err != nil {
			return nil, err
		}
		seatService, err := trainticket.NewSeatServiceImpl(ctx, orderService, configService)
		if err != nil {
			return nil, err
		}

		preserveTravelDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		travelService, err := trainticket.NewTravelServiceImpl(ctx, basicService, seatService, routeService, trainService, preserveTravelDB)
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

		preserveContactsDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		contactsService, err := trainticket.NewContactsServiceImpl(ctx, preserveContactsDB)
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

		preserveUserDB, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		userService, err := trainticket.NewUserServiceImpl(ctx, preserveUserDB)
		if err != nil {
			return nil, err
		}

		emailQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}

		return trainticket.NewPreserveServiceImpl(
			ctx,
			assuranceService,
			basicService,
			consignService,
			contactsService,
			foodService,
			orderService,
			seatService,
			stationService,
			travelService,
			userService,
			emailQueue,
		)
	})
}

func TestPreserveServiceContactNotFound(t *testing.T) {
	ctx := context.Background()
	service, err := preserveServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	oti := trainticket.OrderTicketsInfo{
		AccountID:  "preserve_acc001",
		ContactsID: "nonexistent_contact",
		TripID:     "G_PRESERVE_001",
		SeatType:   2,
		Date:       "2026-05-01",
		From:       "shanghai",
		To:         "beijing",
	}
	_, err = service.Preserve(ctx, oti)
	assert.Error(t, err)
}

func TestPreserveServiceTripNotFound(t *testing.T) {
	ctx := context.Background()
	service, err := preserveServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// seed a contact so the contacts lookup step succeeds
	contactsService, err := trainticket.NewContactsServiceImpl(ctx, preserveContactsDB)
	assert.NoError(t, err)
	err = contactsService.CreateContacts(ctx, trainticket.Contact{
		ID:             "preserve_trip_contact001",
		AccountID:      "preserve_trip_acc001",
		Name:           "Charlie",
		DocumentType:   int(trainticket.ID_CARD),
		DocumentNumber: "TRIP001",
	})
	assert.NoError(t, err)

	oti := trainticket.OrderTicketsInfo{
		AccountID:  "preserve_trip_acc001",
		ContactsID: "preserve_trip_contact001",
		TripID:     "G_NOTFOUND_001",
		SeatType:   2,
		Date:       "2026-06-01 00:00:00",
		From:       "nanjing",
		To:         "wuhan",
	}
	// trip "G_NOTFOUND_001" does not exist; GetTripAllDetailInfo errors but Preserve swallows it
	order, err := service.Preserve(ctx, oti)
	assert.NoError(t, err)
	assert.Empty(t, order.ID)
}

func TestPreserveServiceSuccess(t *testing.T) {
	const (
		accountID    = "preserve_success_acc001"
		contactID    = "preserve_success_contact001"
		tripID       = "G_SUCCESS_001"
		trainType    = "GS"
		routeID      = "00000000000000000000000000000001"
		travelDate   = "2026-06-01 00:00:00"
		startStation = "nanjing"
		endStation   = "wuhan"
	)

	ctx := context.Background()
	service, err := preserveServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// seed contact
	contactsService, err := trainticket.NewContactsServiceImpl(ctx, preserveContactsDB)
	assert.NoError(t, err)
	err = contactsService.CreateContacts(ctx, trainticket.Contact{
		ID:             contactID,
		AccountID:      accountID,
		Name:           "Alice",
		DocumentType:   int(trainticket.ID_CARD),
		DocumentNumber: "SUCCESS001",
	})
	assert.NoError(t, err)

	// seed user
	userService, err := trainticket.NewUserServiceImpl(ctx, preserveUserDB)
	assert.NoError(t, err)
	err = userService.SaveUser(ctx, trainticket.User{
		UserID:   accountID,
		Username: "alice",
		Email:    "alice@example.com",
	})
	assert.NoError(t, err)

	// seed stations
	stationService, err := trainticket.NewStationServiceImpl(ctx, preserveStationDB)
	assert.NoError(t, err)
	err = stationService.CreateStation(ctx, trainticket.Station{ID: "st_nanjing", Name: startStation})
	assert.NoError(t, err)
	err = stationService.CreateStation(ctx, trainticket.Station{ID: "st_wuhan", Name: endStation})
	assert.NoError(t, err)

	// seed train type (G-prefix so GetLeftTicketOfInterval uses the ordering DB path)
	trainServiceSeed, err := trainticket.NewTrainServiceImpl(ctx, preserveTrainDB)
	assert.NoError(t, err)
	err = trainServiceSeed.Create(ctx, trainticket.TrainType{
		ID:           "train_GS",
		Name:         trainType,
		EconomyClass: 100,
		ComfortClass: 50,
		AvgSpeed:     200,
	})
	assert.NoError(t, err)

	// seed route; ID must be >= 32 chars so CreateAndModify uses it directly
	routeServiceSeed, err := trainticket.NewRouteServiceImpl(ctx, preserveRouteDB)
	assert.NoError(t, err)
	_, err = routeServiceSeed.CreateAndModify(ctx, trainticket.RouteInfo{
		ID:           routeID,
		StartStation: startStation,
		EndStation:   endStation,
		StationList:  startStation + "," + endStation,
		DistanceList: "0,300",
	})
	assert.NoError(t, err)

	// seed price config for this route + train type
	priceServiceSeed, err := trainticket.NewPriceServiceImpl(ctx, preservePriceDB)
	assert.NoError(t, err)
	err = priceServiceSeed.CreateNewPriceConfig(ctx, trainticket.PriceConfig{
		ID:                  "price_GS_001",
		TrainType:           trainType,
		RouteID:             routeID,
		BasicPriceRate:      1.0,
		FirstClassPriceRate: 1.5,
	})
	assert.NoError(t, err)

	// seed DirectTicketAllocationProportion so GetLeftTicketOfInterval returns seatTotalNum seats
	configServiceSeed, err := trainticket.NewConfigServiceImpl(ctx, preserveConfigDB)
	assert.NoError(t, err)
	err = configServiceSeed.Create(ctx, trainticket.Config{
		Name:        "DirectTicketAllocationProportion",
		Value:       "1",
		Description: "allocation proportion for direct tickets",
	})
	assert.NoError(t, err)

	// seed trip using a dedicated TravelService that shares preserveTravelDB
	orderDB4seed, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	assert.NoError(t, err)
	orderService4seed, err := trainticket.NewOrderServiceImpl(ctx, orderDB4seed)
	assert.NoError(t, err)
	configDB4seed, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	assert.NoError(t, err)
	configService4seed, err := trainticket.NewConfigServiceImpl(ctx, configDB4seed)
	assert.NoError(t, err)
	seatService4seed, err := trainticket.NewSeatServiceImpl(ctx, orderService4seed, configService4seed)
	assert.NoError(t, err)
	basicService4seed, err := trainticket.NewBasicServiceImpl(ctx, stationService, trainServiceSeed, routeServiceSeed, priceServiceSeed)
	assert.NoError(t, err)
	travelService4seed, err := trainticket.NewTravelServiceImpl(ctx, basicService4seed, seatService4seed, routeServiceSeed, trainServiceSeed, preserveTravelDB)
	assert.NoError(t, err)
	_, err = travelService4seed.CreateTrip(ctx, trainticket.TravelInfo{
		TripID:              tripID,
		TrainTypeName:       trainType,
		RouteID:             routeID,
		StartStationName:    startStation,
		TerminalStationName: endStation,
		StartTime:           "2026-06-01 08:00:00",
		EndTime:             "2026-06-01 12:00:00",
	})
	assert.NoError(t, err)

	oti := trainticket.OrderTicketsInfo{
		AccountID:  accountID,
		ContactsID: contactID,
		TripID:     tripID,
		SeatType:   2,
		Date:       travelDate,
		From:       startStation,
		To:         endStation,
	}
	order, err := service.Preserve(ctx, oti)
	assert.NoError(t, err)
	assert.NotEmpty(t, order.ID)
	assert.Equal(t, tripID, order.TrainNumber)
	assert.Equal(t, accountID, order.AccountID)
	assert.Equal(t, startStation, order.FromStation)
	assert.Equal(t, endStation, order.ToStation)
}
