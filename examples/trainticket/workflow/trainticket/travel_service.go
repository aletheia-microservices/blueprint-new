package trainticket

import (
	"context"
	"fmt"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type TravelService interface {
	GetTrainTypeByTripId(ctx context.Context, tripId string) (TrainType, error)
	GetRouteByTripId(ctx context.Context, tripId string) (Route, error)
	GetTripsByRouteId(ctx context.Context, routeIds []string) ([]Trip, error)
	CreateTrip(ctx context.Context, info TravelInfo) (Trip, error)
	Retrieve(ctx context.Context, tripId string) (Trip, error)
	UpdateTrip(ctx context.Context, info TravelInfo) error
	DeleteTrip(ctx context.Context, tripID string) error
	QueryInfo(ctx context.Context, info TripInfo) ([]TripResponse, error)
	GetTripAllDetailInfo(ctx context.Context, gtdi TripAllDetailInfo) (TripAllDetail, error)
	QueryAll(ctx context.Context) ([]Trip, error)
	AdminQueryAll(ctx context.Context) ([]AdminTrip, error)
}

type TravelServiceImpl struct {
	basicService BasicService
	seatService  SeatService
	routeService RouteService
	trainService TrainService
	travelDB     backend.NoSQLDatabase
}

func NewTravelServiceImpl(ctx context.Context, basicService BasicService, seatService SeatService, routeService RouteService, trainService TrainService, travelDB backend.NoSQLDatabase) (TravelService, error) {
	return &TravelServiceImpl{basicService: basicService, seatService: seatService, routeService: routeService, trainService: trainService, travelDB: travelDB}, nil
}

func (t *TravelServiceImpl) GetTrainTypeByTripId(ctx context.Context, tripId string) (TrainType, error) {
	if len(tripId) >= 2 {
		trip, err := t.findByTripId(ctx, tripId)
		if err != nil {
			return TrainType{}, err
		}

		trainType, err := t.trainService.RetrieveByName(ctx, trip.TrainTypeName)
		if err != nil {
			return TrainType{}, err
		}

		return trainType, nil
	}
	return TrainType{}, fmt.Errorf("trip not found for trip id (%s)", tripId)
}

func (t *TravelServiceImpl) GetRouteByTripId(ctx context.Context, tripId string) (Route, error) {
	if len(tripId) >= 2 {
		trip, err := t.findByTripId(ctx, tripId)
		if err != nil {
			return Route{}, err
		}

		route, err := t.routeService.GetRouteById(ctx, trip.RouteID)
		if err != nil {
			return Route{}, err
		}
		return route, nil
	}
	return Route{}, fmt.Errorf("trip not found for trip id (%s)", tripId)
}

func (t *TravelServiceImpl) GetTripsByRouteId(ctx context.Context, routeIds []string) ([]Trip, error) {
	var tripList []Trip
	for _, routeID := range routeIds {
		trip, err := t.findByRouteId(ctx, routeID)
		if err != nil {
			return nil, err
		}
		tripList = append(tripList, trip)
	}
	return tripList, nil
}

func (t *TravelServiceImpl) CreateTrip(ctx context.Context, info TravelInfo) (Trip, error) {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return Trip{}, err
	}

	filter := bson.D{{Key: "TripID", Value: info.TripID}}
	res, err := collection.FindOne(ctx, filter)
	if err != nil {
		return Trip{}, err
	}
	var trip Trip
	ok, err := res.One(ctx, &trip)
	if err != nil {
		return Trip{}, err
	}
	if ok {
		return Trip{}, fmt.Errorf("trip (%s) already exists", info.TripID)
	}

	trip = Trip{
		TripID:              info.TripID,
		TrainTypeName:       info.TrainTypeName,
		RouteID:             info.RouteID,
		StartStationName:    info.StartStationName,
		StationsName:        info.StationsName,
		TerminalStationName: info.TerminalStationName,
		StartTime:           info.StartTime,
		EndTime:             info.EndTime,
	}
	err = collection.InsertOne(ctx, trip)
	if err != nil {
		return Trip{}, err
	}
	return trip, err
}

func (t *TravelServiceImpl) Retrieve(ctx context.Context, tripId string) (Trip, error) {
	return t.findByTripId(ctx, tripId)
}

func (t *TravelServiceImpl) UpdateTrip(ctx context.Context, info TravelInfo) error {
	trip, err := t.findByTripId(ctx, info.TripID)
	if err != nil {
		return err
	}
	trip.TrainTypeName = info.TrainTypeName
	trip.StartStationName = info.StartStationName
	trip.StationsName = info.StationsName
	trip.TerminalStationName = info.TerminalStationName
	trip.StartTime = info.StartTime
	trip.EndTime = info.EndTime
	trip.RouteID = info.RouteID
	return t.save(ctx, trip)
}

func (t *TravelServiceImpl) DeleteTrip(ctx context.Context, tripID string) error {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "TripID", Value: tripID}}
	return collection.DeleteOne(ctx, filter)
}

func (t *TravelServiceImpl) QueryInfo(ctx context.Context, info TripInfo) ([]TripResponse, error) {
	if info.startPlace == "" || info.endPlace == "" || info.departureTime == "" {
		return nil, fmt.Errorf("something null")
	}
	allTripList, err := t.findAll(ctx)
	if err != nil {
		return nil, err
	}
	return t.getTicketsByBatch(ctx, allTripList, info.startPlace, info.endPlace, info.departureTime)

}

func (t *TravelServiceImpl) getTicketsByBatch(ctx context.Context, trips []Trip, startPlaceName string, endPlaceName string, departureTime string) ([]TripResponse, error) {
	if !afterToday(departureTime) {
		return nil, fmt.Errorf("departure time (%s) not valid", departureTime)
	}
	var infos []Travel
	var tripMap = make(map[string]Trip)

	for _, trip := range trips {
		query := Travel{
			Trip:          trip,
			StartPlace:    startPlaceName,
			EndPlace:      endPlaceName,
			DepartureTime: departureTime,
		}
		infos = append(infos, query)
		tripMap[trip.TripID] = trip
	}

	travelResultMap, err := t.basicService.QueryForTravels(ctx, infos)
	if err != nil {
		return nil, err
	}

	var responses []TripResponse
	for trKey, trVal := range travelResultMap {
		tripNumber := trKey
		travelResult := trVal
		trip := tripMap[tripNumber]
		tripResponse, err := t.setResponse(ctx, trip, travelResult, startPlaceName, endPlaceName, departureTime)
		if err != nil {
			return nil, err
		}
		responses = append(responses, tripResponse)
	}
	return responses, nil
}

func (t *TravelServiceImpl) GetTripAllDetailInfo(ctx context.Context, gtdi TripAllDetailInfo) (TripAllDetail, error) {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return TripAllDetail{}, err
	}
	filter := bson.D{{Key: "TripID", Value: gtdi.TripID}}
	res, err := collection.FindOne(ctx, filter)
	if err != nil {
		return TripAllDetail{}, err
	}
	var trip Trip
	ok, err := res.One(ctx, &trip)
	if err != nil {
		return TripAllDetail{}, err
	}
	if !ok {
		return TripAllDetail{}, fmt.Errorf("trip (%s) not found", gtdi.TripID)
	}

	tripResponse, err := t.getTickets(ctx, trip, Route{}, gtdi.From, gtdi.To, gtdi.TravelDate)
	if err != nil {
		return TripAllDetail{}, err
	}

	gtdr := TripAllDetail{
		trip:         trip,
		tripResponse: tripResponse,
	}
	return gtdr, nil
}

func (t *TravelServiceImpl) QueryAll(ctx context.Context) ([]Trip, error) {
	return t.findAll(ctx)
}

func (t *TravelServiceImpl) AdminQueryAll(ctx context.Context) ([]AdminTrip, error) {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return nil, err
	}

	cursor, err := collection.FindMany(ctx, nil)
	if err != nil {
		return nil, err
	}

	var trips []Trip
	err = cursor.All(ctx, &trips)
	if err != nil {
		return nil, err
	}

	var adminTrips []AdminTrip
	for _, trip := range trips {
		route, err := t.routeService.GetRouteById(ctx, trip.RouteID)
		if err != nil {
			return nil, err
		}
		trainType, err := t.trainService.RetrieveByName(ctx, trip.TrainTypeName)
		if err != nil {
			return nil, err
		}
		adminTrip := AdminTrip{
			Trip:      trip,
			Route:     route,
			TrainType: trainType,
		}
		adminTrips = append(adminTrips, adminTrip)
	}
	return adminTrips, nil
}

func (t *TravelServiceImpl) save(ctx context.Context, trip Trip) error {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return err
	}
	_, err = collection.Upsert(ctx, bson.D{{Key: "TripID", Value: trip.TripID}}, trip)
	return err
}

func (t *TravelServiceImpl) findByTripId(ctx context.Context, tripId string) (Trip, error) {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return Trip{}, err
	}

	filter := bson.D{{Key: "TripID", Value: tripId}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return Trip{}, err
	}

	var trip Trip
	ok, err := cursor.One(ctx, &trip)
	if err != nil {
		return Trip{}, err
	}
	if !ok {
		return Trip{}, fmt.Errorf("trip not found for trip id (%s)", tripId)
	}
	return trip, nil
}

func (t *TravelServiceImpl) findByRouteId(ctx context.Context, routeId string) (Trip, error) {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return Trip{}, err
	}

	filter := bson.D{{Key: "RouteID", Value: routeId}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return Trip{}, err
	}

	var trip Trip
	_, err = cursor.One(ctx, &trip)
	if err != nil {
		return Trip{}, err
	}
	return trip, nil
}

func (t *TravelServiceImpl) findAll(ctx context.Context) ([]Trip, error) {
	collection, err := t.travelDB.GetCollection(ctx, "travel_db", "trip")
	if err != nil {
		return nil, err
	}

	cursor, err := collection.FindMany(ctx, nil)
	if err != nil {
		return nil, err
	}

	var trips []Trip
	return trips, cursor.All(ctx, &trips)
}

func (t *TravelServiceImpl) getTickets(ctx context.Context, trip Trip, route1 Route, startPlaceName string, endPlaceName string, departureTime string) (TripResponse, error) {
	if !afterToday(departureTime) {
		return TripResponse{}, fmt.Errorf("departure time (%s) not valid", departureTime)
	}
	travelQuery := Travel{
		Trip:          trip,
		StartPlace:    startPlaceName,
		EndPlace:      endPlaceName,
		DepartureTime: departureTime,
	}
	travelResult, err := t.basicService.QueryForTravel(ctx, travelQuery)
	if err != nil {
		return TripResponse{}, err
	}
	return t.setResponse(ctx, trip, travelResult, startPlaceName, endPlaceName, departureTime)

}

func (t *TravelServiceImpl) setResponse(ctx context.Context, trip Trip, travelResult TravelResult, startPlaceName string, endPlaceName string, departureTime string) (TripResponse, error) {
	var response TripResponse
	response.ConfortClass = 50
	response.EconomyClass = 50

	route := travelResult.Route
	stationList := route.Stations

	firstClassTotalNum := travelResult.TrainType.ComfortClass
	secondClassTotalNum := travelResult.TrainType.EconomyClass

	first, err := t.seatService.GetLeftTicketOfInterval(ctx, SeatRequest{departureTime, trip.TripID, startPlaceName, endPlaceName, 1, firstClassTotalNum, stationList})
	if err != nil {
		return TripResponse{}, err
	}

	second, err := t.seatService.GetLeftTicketOfInterval(ctx, SeatRequest{departureTime, trip.TripID, startPlaceName, endPlaceName, 2, secondClassTotalNum, stationList})
	if err != nil {
		return TripResponse{}, err
	}

	indexStart := indexOf(route.Stations, startPlaceName)
	indexEnd := indexOf(route.Stations, endPlaceName)
	distanceStart := route.Distances[indexStart] - route.Distances[0]
	distanceEnd := route.Distances[indexEnd] - route.Distances[0]
	trainType := travelResult.TrainType
	minutesStart := 60 * distanceStart / trainType.AvgSpeed
	minutesEnd := 60 * distanceEnd / trainType.AvgSpeed

	start, err := time.ParseInLocation(CALENDAR_LAYOUT, trip.StartTime, time.Local)
	if err != nil {
		return TripResponse{}, err
	}
	calendarStart := start.Add(time.Duration(minutesStart) * time.Minute)
	calendarEnd := start.Add(time.Duration(minutesEnd) * time.Minute)

	startTime := calendarStart.Format(CALENDAR_LAYOUT)
	endTime := calendarEnd.Format(CALENDAR_LAYOUT)

	tripResponse := TripResponse{
		TripID:               trip.TripID,
		TrainTypeName:        trip.TrainTypeName,
		ConfortClass:         first,
		EconomyClass:         second,
		PriceForConfortClass: travelResult.Prices["confortClass"],
		PriceForEconomyClass: travelResult.Prices["economyClass"],
		StartStation:         startPlaceName,
		TerminalStation:      endPlaceName,
		StartTime:            startTime,
		EndTime:              endTime,
	}
	return tripResponse, nil
}

func afterToday(dateStr string) bool {
	d, err := time.ParseInLocation(CALENDAR_LAYOUT, dateStr, time.Local)
	if err != nil {
		return false
	}

	now := time.Now()
	yearA, monthA, dayA := now.Date()
	yearB, monthB, datB := d.Date()

	if yearA > yearB {
		return false
	} else if yearA == yearB {
		if int(monthA) > int(monthB) {
			return false
		} else if int(monthA) == int(monthB) {
			return dayA <= datB
		} else {
			return true
		}
	}
	return true
}
