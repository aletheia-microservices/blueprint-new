package hotelreservation

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// RateService implements Rate Service from the hotel reservation application
type RateService interface {
	// GetRates return the rates for the desired hotels (`hotelIDs`) for the provided dates (`inDate`, `outDate`)
	GetRates(ctx context.Context, hotelIDs []string, inDate string, outDate string) ([]RatePlan, error)
}

// Implementation of RateService
type RateServiceImpl struct {
	rateCache backend.Cache
	rateDB    backend.NoSQLDatabase
}

// Creates and Returns a new RateService object
func NewRateServiceImpl(ctx context.Context, rateCache backend.Cache, rateDB backend.NoSQLDatabase) (RateService, error) {
	return &RateServiceImpl{rateCache: rateCache, rateDB: rateDB}, nil
}

func (r *RateServiceImpl) GetRates(ctx context.Context, hotelIDs []string, inDate string, outDate string) ([]RatePlan, error) {
	var rate_plans []RatePlan

	for _, hotel_id := range hotelIDs {
		var hotel_rate_plans []RatePlan
		exists, err := r.rateCache.Get(ctx, hotel_id, &hotel_rate_plans)
		if err != nil {
			return rate_plans, err
		}
		if !exists {
			collection, err2 := r.rateDB.GetCollection(ctx, "rate_db", "inventory")
			if err2 != nil {
				return []RatePlan{}, err2
			}
			query := bson.D{{Key: "HotelID", Value: hotel_id}}
			rs, err := collection.FindMany(ctx, query)
			if err != nil {
				return rate_plans, err
			}
			rs.All(ctx, &hotel_rate_plans)
			err = r.rateCache.Put(ctx, hotel_id, hotel_rate_plans)
		}
		rate_plans = append(rate_plans, hotel_rate_plans...)
	}
	// TODO: Sort rate_plans
	return rate_plans, nil
}
