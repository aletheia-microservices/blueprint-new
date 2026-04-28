package hotelreservation

import (
	"context"
	"math"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/hailocab/go-geoindex"
	"go.mongodb.org/mongo-driver/bson"
)

// RecommendationService implements Recommendation Service from the hotel reservation application
type RecommendationService interface {
	// Returns the recommended hotels based on the desired location (`lat`, `lon`) and the metric (`require`) for ranking recommendations
	GetRecommendations(ctx context.Context, require string, lat float64, lon float64) ([]string, error)
}

// Implements RecommendationService
type RecommendationServiceImpl struct {
	recommendDB backend.NoSQLDatabase
	hotels      map[string]Hotel
}

// Creates and Returns a new RecommendationService object
func NewRecommendationServiceImpl(ctx context.Context, recommendDB backend.NoSQLDatabase) (RecommendationService, error) {
	service := &RecommendationServiceImpl{recommendDB: recommendDB, hotels: make(map[string]Hotel)}
	err := service.LoadRecommendations(context.Background())
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (r *RecommendationServiceImpl) LoadRecommendations(ctx context.Context) error {
	collection, err := r.recommendDB.GetCollection(ctx, "recommendation_db", "recommendation")
	if err != nil {
		return err
	}

	filter := bson.D{}
	res, err := collection.FindMany(ctx, filter)
	if err != nil {
		return err
	}
	var hotels []Hotel
	res.All(ctx, &hotels)
	for _, hotel := range hotels {
		r.hotels[hotel.HId] = hotel
	}

	return nil
}

func (r *RecommendationServiceImpl) GetRecommendations(ctx context.Context, require string, lat float64, lon float64) ([]string, error) {

	var hotelIds []string
	if require == "dis" {
		p1 := &geoindex.GeoPoint{Pid: "", Plat: lat, Plon: lon}
		min := math.MaxFloat64
		dist := make(map[string]float64)
		for _, hotel := range r.hotels {
			hotel_pt := &geoindex.GeoPoint{Pid: "", Plat: hotel.HLat, Plon: hotel.HLon}
			tmp := float64(geoindex.Distance(p1, hotel_pt)) / 1000
			if tmp < min {
				min = tmp
			}
			dist[hotel.HId] = tmp
		}
		for _, hotel := range r.hotels {
			distance := dist[hotel.HId]
			if distance == min {
				hotelIds = append(hotelIds, hotel.HId)
			}
		}
	} else if require == "rate" {
		max := 0.0
		rates := make(map[string]float64)
		for _, hotel := range r.hotels {
			if hotel.HRate > max {
				max = hotel.HRate
			}
			rates[hotel.HId] = hotel.HRate
		}
		for _, hotel := range r.hotels {
			rate := rates[hotel.HId]
			if rate == max {
				hotelIds = append(hotelIds, hotel.HId)
			}
		}
	} else if require == "price" {
		min := math.MaxFloat64
		prices := make(map[string]float64)
		for _, hotel := range r.hotels {
			if hotel.HPrice < min {
				min = hotel.HPrice
			}
			prices[hotel.HId] = hotel.HPrice
		}
		for hid, price := range prices {
			if min == price {
				hotelIds = append(hotelIds, hid)
			}
		}
	}

	return hotelIds, nil
}
