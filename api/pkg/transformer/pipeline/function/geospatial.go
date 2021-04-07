package function

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/golang/geo/s2"
	"github.com/mmcloughlin/geohash"
)

type LatLong struct {
	lat  float64
	long float64
}

func (l *LatLong) toGeoHash(precision uint) interface{} {
	return geohash.EncodeWithPrecision(l.lat, l.long, precision)
}

func (l *LatLong) toS2ID(level int) interface{} {
	return fmt.Sprintf("%d", s2.CellIDFromLatLng(s2.LatLngFromDegrees(l.lat, l.long)).Parent(level))
}

func Geohash(lat interface{}, lon interface{}, precision uint) (interface{}, error) {
	latLong, err := extractLatLong(lat, lon)
	if err != nil {
		panic(err)
	}

	switch latLong.(type) {
	case []*LatLong:
		var value []interface{}
		for _, ll := range latLong.([]*LatLong) {
			value = append(value, ll.toGeoHash(precision))
		}
		return value, nil
	case *LatLong:
		value := latLong.(*LatLong).toGeoHash(precision)
		return value, nil
	default:
		return nil, fmt.Errorf("unknown type: %T", latLong)
	}
}

func S2ID(lat interface{}, lon interface{}, level int) (interface{}, error) {
	latLong, err := extractLatLong(lat, lon)
	if err != nil {
		panic(err)
	}

	switch latLong.(type) {
	case []*LatLong:
		var value []interface{}
		for _, ll := range latLong.([]*LatLong) {
			value = append(value, ll.toS2ID(level))
		}
		return value, nil
	case *LatLong:
		value := latLong.(*LatLong).toS2ID(level)
		return value, nil
	default:
		return nil, fmt.Errorf("unknown type: %T", latLong)
	}
}

func toFloat64(o interface{}) (float64, error) {
	var floatValue float64
	switch o.(type) {
	case float64:
		floatValue = o.(float64)
	case string:
		parsed, err := strconv.ParseFloat(o.(string), 64)
		if err != nil {
			return 0, err
		}
		floatValue = parsed
	case int:
		floatValue = float64(o.(int))
	default:
		return 0, errors.New(fmt.Sprintf("%v cannot be parsed into float64", reflect.TypeOf(o)))
	}
	return floatValue, nil
}

func extractLatLong(latitude, longitude interface{}) (interface{}, error) {
	if reflect.TypeOf(latitude) != reflect.TypeOf(longitude) {
		return nil, errors.New("latitude and longitude must have the same types")
	}

	var latLong interface{}

	latVal := reflect.ValueOf(latitude)
	lonVal := reflect.ValueOf(longitude)
	switch latVal.Kind() {
	case reflect.Slice:
		if latVal.Len() != lonVal.Len() {
			return nil, errors.New("both latitude and longitude arrays must have the same length")
		}

		if latVal.Len() == 0 {
			return nil, errors.New("empty arrays of latitudes and longitude provided")
		}

		var latLongArray []*LatLong
		for index := 0; index < latVal.Len(); index++ {
			lat := latVal.Index(index)
			lon := lonVal.Index(index)

			latitudeFloat, err := toFloat64(lat.Interface())
			if err != nil {
				return nil, err
			}

			longitudeFloat, err := toFloat64(lon.Interface())
			if err != nil {
				return nil, err
			}
			latLongArray = append(latLongArray, &LatLong{
				lat:  latitudeFloat,
				long: longitudeFloat,
			})
		}
		latLong = latLongArray

	default:
		latitudeFloat, err := toFloat64(latitude)
		if err != nil {
			return nil, err
		}
		longitudeFloat, err := toFloat64(longitude)
		if err != nil {
			return nil, err
		}
		latLong = &LatLong{
			lat:  latitudeFloat,
			long: longitudeFloat,
		}
	}

	return latLong, nil
}
