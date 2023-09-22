package location

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GeoLocation struct {
	Longtitude float64
	Latitude   float64
	Altitude   float64
	Accuracy   float64
}

var ErrNoGeolocation = errors.New("no geolocation data")

// GeoLocation parse string form
// Longtitude 	=> glong
// Latitude 	=> gla
// Altitude 	=> gal
// Accuracy		=> gacc
func GeoLocationFromCTX(ctx *gin.Context) (geoLocation *GeoLocation, err error) {
	if geoLocation, err = getGeoLocationFromHeader(ctx); err == nil {
		return
	}
	if geoLocation, err = getGeoLocationFromQuery(ctx); err == nil {
		return
	}
	return nil, ErrNoGeolocation
}

func getGeoLocationFromQuery(ctx *gin.Context) (*GeoLocation, error) {
	q := ctx.Request.URL.Query()

	longtitude, err := strconv.ParseFloat(q.Get("glong"), 64)
	if err != nil {
		return nil, err
	}
	latitude, err := strconv.ParseFloat(q.Get("gla"), 64)
	if err != nil {
		return nil, err
	}
	altitude, err := strconv.ParseFloat(q.Get("gal"), 64)
	if err != nil {
		return nil, err
	}
	accuracy, err := strconv.ParseFloat(q.Get("gacc"), 64)
	if err != nil {
		return nil, err
	}

	return &GeoLocation{
		Longtitude: longtitude,
		Latitude:   latitude,
		Altitude:   altitude,
		Accuracy:   accuracy,
	}, nil
}

func getGeoLocationFromHeader(ctx *gin.Context) (*GeoLocation, error) {
	longtitude, err := strconv.ParseFloat(ctx.GetHeader("glong"), 64)
	if err != nil {
		return nil, err
	}
	latitude, err := strconv.ParseFloat(ctx.GetHeader("gla"), 64)
	if err != nil {
		return nil, err
	}
	altitude, err := strconv.ParseFloat(ctx.GetHeader("gal"), 64)
	if err != nil {
		return nil, err
	}
	accuracy, err := strconv.ParseFloat(ctx.GetHeader("gacc"), 64)
	if err != nil {
		return nil, err
	}

	return &GeoLocation{
		Longtitude: longtitude,
		Latitude:   latitude,
		Altitude:   altitude,
		Accuracy:   accuracy,
	}, nil
}
