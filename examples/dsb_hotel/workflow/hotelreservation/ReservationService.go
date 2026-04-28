package hotelreservation

import (
	"context"
	//"strconv"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// ReservationService implements ReservationService from hotel reservation application
type ReservationService interface {
	// Makes a reservation at the desired hotel (`hotelIds[0]`, len(hotelIds) == 1). Returns the hotelID if the reservation is successful.
	MakeReservation(ctx context.Context, customerName string, hotelIds []string, inDate string, outDate string, roomNumber int64) ([]string, error)
	// Returns the subset of hotels from desired hotels that are available for reservation
	CheckAvailability(ctx context.Context, customerName string, hotelIDs []string, inDate string, outDate string, roomNumber int64) ([]string, error)
}

type ReservationServiceImpl struct {
	reserveCache backend.Cache
	reserveDB    backend.NoSQLDatabase
	CacheHits    int64
	NumRequests  int64
}

func NewReservationServiceImpl(ctx context.Context, reserveCache backend.Cache, reserveDB backend.NoSQLDatabase) (ReservationService, error) {
	r := &ReservationServiceImpl{reserveCache: reserveCache, reserveDB: reserveDB}
	return r, nil
}

func (r *ReservationServiceImpl) MakeReservation(ctx context.Context, customerName string, hotelIds []string, inDate string, outDate string, roomNumber int64) ([]string, error) {
	reservation_collection, err := r.reserveDB.GetCollection(ctx, "reservation_db", "reservation")
	if err != nil {
		return []string{}, err
	}
	hnumber_collection, err := r.reserveDB.GetCollection(ctx, "reservation_db", "number")
	if err != nil {
		return []string{}, err
	}
	newInDate, _ := time.Parse(time.RFC3339, inDate+"T12:00:00+00:00")
	newOutDate, _ := time.Parse(time.RFC3339, outDate+"T12:00:00+00:00")
	hotelId := hotelIds[0]
	indate := newInDate.String()[0:10]

	reservation_update_map := make(map[string]int64)
	for newInDate.Before(newOutDate) {
		newInDate = newInDate.AddDate(0, 0, 1)
		outdate := newInDate.String()[0:10]

		key := hotelId + "_" + newInDate.String()[0:10] + "_" + outdate
		var room_number int64
		exists, err := r.reserveCache.Get(ctx, key, &room_number)
		if err != nil {
			return []string{}, err
		}
		if !exists {
			var reservations []Reservation

			query := bson.D{{Key: "HotelId", Value: hotelId}, {Key: "InDate", Value: indate}, {Key: "OutDate", Value: outdate}}
			res, err := reservation_collection.FindMany(ctx, query)
			if err != nil {
				return []string{}, err
			}
			res.All(ctx, &reservations)
			for _, reservation := range reservations {
				room_number += reservation.Number
			}
		}
		reservation_update_map[key] = room_number + roomNumber

		// Check capacity
		cap_key := hotelId + "_cap"
		var hotelNumber HotelNumber
		var capacity int64
		exists, err = r.reserveCache.Get(ctx, cap_key, &capacity)
		if err != nil {
			return []string{}, err
		}
		if !exists {
			query := bson.D{{Key: "HotelId", Value: hotelId}}
			res, err := hnumber_collection.FindOne(ctx, query)
			if err != nil {
				return []string{}, err
			}
			res.One(ctx, &hotelNumber)
			capacity = hotelNumber.Number
			err = r.reserveCache.Put(ctx, cap_key, capacity)
			if err != nil {
				return []string{}, err
			}
		}
		if room_number+roomNumber > capacity {
			return []string{}, nil
		}
		indate = outdate
	}

	newInDate, _ = time.Parse(time.RFC3339, inDate+"T12:00:00+00:00")
	indate = newInDate.String()[0:10]

	for newInDate.Before(newOutDate) {
		newInDate = newInDate.AddDate(0, 0, 1)
		outdate := newInDate.String()[0:10]
		reservation := Reservation{HotelId: hotelId, CustomerName: customerName, InDate: indate, OutDate: outdate, Number: roomNumber}
		err := reservation_collection.InsertOne(ctx, reservation)
		if err != nil {
			return []string{}, err
		}
	}

	for key, val := range reservation_update_map {
		err := r.reserveCache.Put(ctx, key, val)
		if err != nil {
			return []string{}, err
		}
	}

	return []string{hotelId}, nil
}

func (r *ReservationServiceImpl) CheckAvailability(ctx context.Context, customerName string, hotelIds []string, inDate string, outDate string, roomNumber int64) ([]string, error) {
	reservation_collection, err := r.reserveDB.GetCollection(ctx, "reservation_db", "reservation")
	if err != nil {
		return []string{}, err
	}
	hnumber_collection, err := r.reserveDB.GetCollection(ctx, "reservation_db", "number")
	if err != nil {
		return []string{}, err
	}

	var available_hotels []string

	for _, hotelId := range hotelIds {
		newInDate, _ := time.Parse(time.RFC3339, inDate+"T12:00:00+00:00")
		newOutDate, _ := time.Parse(time.RFC3339, outDate+"T12:00:00+00:00")
		indate := newInDate.String()[0:10]

		for newInDate.Before(newOutDate) {
			newInDate = newInDate.AddDate(0, 0, 1)
			outdate := newInDate.String()[0:10]
			key := hotelId + "_" + newInDate.String()[0:10] + "_" + outdate
			r.NumRequests += 1
			var count int64
			exists, err := r.reserveCache.Get(ctx, key, &count)
			if err != nil {
				return []string{}, err
			}
			if !exists {
				// Check Database
				var reservations []Reservation
				//query := `{"HotelId":"` + hotelId + `", "InDate":"` + indate + `", "OutDate":"` + outdate + `"}`
				query := bson.D{{Key: "HotelId", Value: hotelId}, {Key: "InDate", Value: indate}, {Key: "OutDate", Value: outdate}} // TODO fix this?
				res, err := reservation_collection.FindMany(ctx, query)
				if err != nil {
					return []string{}, err
				}
				res.All(ctx, &reservations)
				for _, reservation := range reservations {
					count += reservation.Number
				}
				err = r.reserveCache.Put(ctx, key, count)
				if err != nil {
					return []string{}, err
				}
			} else {
				r.CacheHits += 1
			}

			// Check capacity
			cap_key := hotelId + "_cap"
			var capacity int64
			var hotelNumber HotelNumber
			exists, err = r.reserveCache.Get(ctx, cap_key, &capacity)
			r.NumRequests += 1
			if err != nil {
				return []string{}, err
			}
			if !exists {
				query := bson.D{{Key: "HotelId", Value: hotelId}}
				res, err := hnumber_collection.FindOne(ctx, query)
				if err != nil {
					return []string{}, err
				}
				res.One(ctx, &hotelNumber)
				capacity = hotelNumber.Number
				err = r.reserveCache.Put(ctx, cap_key, capacity)
				if err != nil {
					return []string{}, err
				}
			} else {
				r.CacheHits += 1
			}
			if count+roomNumber > capacity {
				break
			}

			if newInDate.Equal(newOutDate) {
				available_hotels = append(available_hotels, hotelId)
			}
		}
	}

	return available_hotels, nil
}
