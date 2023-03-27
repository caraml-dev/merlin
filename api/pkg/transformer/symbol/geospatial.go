package symbol

import (
	"strings"

	"github.com/caraml-dev/merlin/pkg/transformer/symbol/function"
)

// Geohash calculates geohash of latitude and longitude with the given character precision
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) Geohash(latitude interface{}, longitude interface{}, precision uint) interface{} {
	lat, err := sr.evalArg(latitude)
	if err != nil {
		panic(err)
	}

	lon, err := sr.evalArg(longitude)
	if err != nil {
		panic(err)
	}

	result, err := function.Geohash(lat, lon, precision)
	if err != nil {
		panic(err)
	}

	return result
}

// S2ID calculates S2 ID of latitude and longitude of the given level
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) S2ID(latitude interface{}, longitude interface{}, level int) interface{} {
	lat, err := sr.evalArg(latitude)
	if err != nil {
		panic(err)
	}

	lon, err := sr.evalArg(longitude)
	if err != nil {
		panic(err)
	}

	result, err := function.S2ID(lat, lon, level)
	if err != nil {
		panic(err)
	}

	return result
}

// HaversineDistance of two points (latitude, longitude) in KM
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) HaversineDistance(latitude1 interface{}, longitude1 interface{}, latitude2 interface{}, longitude2 interface{}) interface{} {
	return sr.HaversineDistanceWithUnit(latitude1, longitude1, latitude2, longitude2, function.KMDistanceUnit)
}

// HaversineDistanceWithUnit of two points (latitude, longitude) and user need to specify distance unit e.g 'km' or 'm'
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) HaversineDistanceWithUnit(latitude1 interface{}, longitude1 interface{}, latitude2 interface{}, longitude2 interface{}, distanceUnit string) interface{} {
	lat1, err := sr.evalArg(latitude1)
	if err != nil {
		panic(err)
	}

	lon1, err := sr.evalArg(longitude1)
	if err != nil {
		panic(err)
	}

	lat2, err := sr.evalArg(latitude2)
	if err != nil {
		panic(err)
	}

	lon2, err := sr.evalArg(longitude2)
	if err != nil {
		panic(err)
	}

	normalizedDistanceUnit := strings.ToLower(distanceUnit)
	result, err := function.HaversineDistance(lat1, lon1, lat2, lon2, normalizedDistanceUnit)
	if err != nil {
		panic(err)
	}

	return result
}

// GeohashDistance will calculate haversine distance of two geohash
// Those two geohashes will be converted back to latitude and longitude that represent center of geohash point
// 'firstGeohash' and 'secondGeohash' can be:
// - Json path string
// - Slice / gota.Series
// - string
func (sr Registry) GeohashDistance(firstGeohash interface{}, secondGeohash interface{}, distanceUnit string) interface{} {
	geohash1, err := sr.evalArg(firstGeohash)
	if err != nil {
		panic(err)
	}
	geohash2, err := sr.evalArg(secondGeohash)
	if err != nil {
		panic(err)
	}
	result, err := function.GeohashDistance(geohash1, geohash2, distanceUnit)
	if err != nil {
		panic(err)
	}
	return result
}

// GeohashAllNeighbors get all neighbors of a geohash from all directions
func (sr Registry) GeohashAllNeighbors(targetGeohash interface{}) interface{} {
	geohashVal, err := sr.evalArg(targetGeohash)
	if err != nil {
		panic(err)
	}
	result, err := function.GeohashAllNeighbors(geohashVal)
	if err != nil {
		panic(err)
	}
	return result
}

// GeohashNeighborForDirection get all neighbor of a geohash from one direction
func (sr Registry) GeohashNeighborForDirection(targetGeohash interface{}, direction string) interface{} {
	geohashVal, err := sr.evalArg(targetGeohash)
	if err != nil {
		panic(err)
	}
	result, err := function.GeohashNeighborForDirection(geohashVal, direction)
	if err != nil {
		panic(err)
	}
	return result
}

// PolarAngle calculate polar angle of two locations given latitude1, longitude1, latitude1, latitude2
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) PolarAngle(latitude1 interface{}, longitude1 interface{}, latitude2 interface{}, longitude2 interface{}) interface{} {
	lat1, err := sr.evalArg(latitude1)
	if err != nil {
		panic(err)
	}

	lon1, err := sr.evalArg(longitude1)
	if err != nil {
		panic(err)
	}

	lat2, err := sr.evalArg(latitude2)
	if err != nil {
		panic(err)
	}

	lon2, err := sr.evalArg(longitude2)
	if err != nil {
		panic(err)
	}
	result, err := function.PolarAngle(lat1, lon1, lat2, lon2)
	if err != nil {
		panic(err)
	}

	return result
}
