package hotelreservation

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	geoindex "github.com/hailocab/go-geoindex"
	"go.mongodb.org/mongo-driver/bson"
)

// GeoService implements the GeoService from HotelReservation
type GeoService interface {
	// Returns list of hotel IDs that are near to the provided coordinates (`lat`, `lon`)
	Nearby(ctx context.Context, lat float64, lon float64) ([]string, error)
}

// Implementation of GeoService
type GeoServiceImpl struct {
	geoDB backend.NoSQLDatabase
	index *geoindex.ClusteringIndex
}

// Creates and returns a new GeoService object
func NewGeoServiceImpl(ctx context.Context, geoDB backend.NoSQLDatabase) (GeoService, error) {
	service := &GeoServiceImpl{geoDB: geoDB}
	service.newGeoIndex(ctx)
	return service, nil
}

const (
	MAXSEARCHRESULTS = 5
	MAXSEARCHRADIUS  = 10
)

func (g *GeoServiceImpl) newGeoIndex(ctx context.Context) error {
	collection, err := g.geoDB.GetCollection(ctx, "geo_db", "geo")
	if err != nil {
		return err
	}
	var points []Point
	filter := bson.D{}
	res, err := collection.FindMany(ctx, filter)
	if err != nil {
		return err
	}
	res.All(ctx, &points)
	g.index = geoindex.NewClusteringIndex()
	for _, point := range points {
		g.index.Add(point)
	}
	return nil
}

/* func (g *GeoServiceImpl) getNearbyPoints(lat float64, lon float64) []geoindex.Point {
	center := &Point{Pid: "", Plat: lat, Plon: lon}

	return g.index.KNearest(center, MAXSEARCHRESULTS, geoindex.Km(MAXSEARCHRADIUS), func(p geoindex.Point) bool { return true })
} */

func (g *GeoServiceImpl) Nearby(ctx context.Context, lat float64, lon float64) ([]string, error) {
	//points := g.getNearbyPoints(lat, lon)
	var points []geoindex.Point
	var hotelIds []string
	for _, p := range points {
		hotelIds = append(hotelIds, p.Id())
	}

	return hotelIds, nil
}
