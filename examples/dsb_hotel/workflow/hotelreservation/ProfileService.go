package hotelreservation

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// ProfileService implements Profile Service from the hotel reservation application
type ProfileService interface {
	// Returns the profiles of hotels based on the `hotelIds` provided
	GetProfiles(ctx context.Context, hotelIds []string, locale string) ([]HotelProfile, error)
}

// Implementation of Profile Service
type ProfileServiceImpl struct {
	profileCache backend.Cache
	profileDB    backend.NoSQLDatabase
}

// Creates and Returns a new Profile Service object
func NewProfileServiceImpl(ctx context.Context, profileCache backend.Cache, profileDB backend.NoSQLDatabase) (ProfileService, error) {
	return &ProfileServiceImpl{profileCache: profileCache, profileDB: profileDB}, nil
}

func (p *ProfileServiceImpl) GetProfiles(ctx context.Context, hotelIds []string, locale string) ([]HotelProfile, error) {
	var profiles []HotelProfile

	for _, hid := range hotelIds {
		var profile HotelProfile
		exists, err := p.profileCache.Get(ctx, hid, &profile)
		if err != nil {
			return profiles, err
		}
		if !exists {
			// Check Database
			collection, err := p.profileDB.GetCollection(ctx, "profile_db", "hotels")
			if err != nil {
				return []HotelProfile{}, err
			}
			query := bson.D{{Key: "ID", Value: hid}}
			res, err := collection.FindOne(ctx, query)
			if err != nil {
				return profiles, err
			}
			res.One(ctx, &profile)
			err = p.profileCache.Put(ctx, hid, profile)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}
