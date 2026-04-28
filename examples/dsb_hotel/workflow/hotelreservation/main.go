package hotelreservation

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

func main() {
	ctx := context.Background()

	var geoDB backend.NoSQLDatabase
	geoService, _ := NewGeoServiceImpl(ctx, geoDB)

	var lat, lon float64
	geoService.Nearby(ctx, lat, lon)
}
