package symbol

import "github.com/gojek/merlin/pkg/transformer/symbol/function"

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

// HaversineDistance of two points (latitude, longitude)
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) HaversineDistance(latitude1 interface{}, longitude1 interface{}, latitude2 interface{}, longitude2 interface{}) interface{} {
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
	result, err := function.HaversineDistance(lat1, lon1, lat2, lon2)
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
