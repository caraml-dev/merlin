package function

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/golang/geo/s2"
	"github.com/mmcloughlin/geohash"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

const (
	earthRaidusKm = 6371 // radius of the earth in kilometers.
	pointFive     = 0.5
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

func PolarAngle(lat1 interface{}, lon1 interface{}, lat2 interface{}, lon2 interface{}) (interface{}, error) {
	firstPoint, err := extractLatLong(lat1, lon1)
	if err != nil {
		return nil, err
	}

	secondPoint, err := extractLatLong(lat2, lon2)
	if err != nil {
		return nil, err
	}

	switch firstPoint.(type) {
	case []*LatLong:
		firstPoints := firstPoint.([]*LatLong)
		secondPoints, ok := secondPoint.([]*LatLong)
		if !ok {
			return nil, errors.New("first point and second point has different format")
		}
		if len(secondPoints) != len(firstPoints) {
			return nil, errors.New("both first point and second point arrays must have the same length")
		}
		var values []interface{}
		for idx, point1 := range firstPoints {
			point2 := secondPoints[idx]
			values = append(values, calculatePolarAngle(point1, point2))
		}
		return values, nil

	case *LatLong:
		firstPoint := firstPoint.(*LatLong)
		secondPoint, ok := secondPoint.(*LatLong)
		if !ok {
			return nil, errors.New("first point and second point has different format")
		}
		value := calculatePolarAngle(firstPoint, secondPoint)
		return value, nil

	default:
		return nil, fmt.Errorf("unknown type: %T", firstPoint)
	}
}

func HaversineDistance(lat1 interface{}, lon1 interface{}, lat2 interface{}, lon2 interface{}) (interface{}, error) {
	firstPoint, err := extractLatLong(lat1, lon1)
	if err != nil {
		return nil, err
	}

	secondPoint, err := extractLatLong(lat2, lon2)
	if err != nil {
		return nil, err
	}

	switch firstPoint.(type) {
	case []*LatLong:
		firstPoints := firstPoint.([]*LatLong)
		secondPoints, ok := secondPoint.([]*LatLong)
		if !ok {
			return nil, errors.New("first point and second point has different format")
		}
		if len(secondPoints) != len(firstPoints) {
			return nil, errors.New("both first point and second point arrays must have the same length")
		}
		var values []interface{}
		for idx, point1 := range firstPoints {
			point2 := secondPoints[idx]
			values = append(values, calculateDistance(point1, point2))
		}
		return values, nil

	case *LatLong:
		firstPoint := firstPoint.(*LatLong)
		secondPoint, ok := secondPoint.(*LatLong)
		if !ok {
			return nil, errors.New("first point and second point has different format")
		}
		value := calculateDistance(firstPoint, secondPoint)
		return value, nil

	default:
		return nil, fmt.Errorf("unknown type: %T", firstPoint)
	}
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

func calculateDistance(firstPoint *LatLong, secondPoint *LatLong) float64 {
	lat1 := degreesToRadians(firstPoint.lat)
	lon1 := degreesToRadians(firstPoint.long)
	lat2 := degreesToRadians(secondPoint.lat)
	lon2 := degreesToRadians(secondPoint.long)

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return c * earthRaidusKm
}

func calculatePolarAngle(firstPoint *LatLong, secondPoint *LatLong) float64 {
	lat1 := degreesToRadians(firstPoint.lat)
	lon1 := degreesToRadians(firstPoint.long)
	lat2 := degreesToRadians(secondPoint.lat)
	lon2 := degreesToRadians(secondPoint.long)

	y := lat2 - lat1
	x := (lon2 - lon1) * math.Cos(pointFive*(lat2+lat1))
	return math.Atan2(y, x)
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

			latitudeFloat, err := converter.ToFloat64(lat.Interface())
			if err != nil {
				return nil, err
			}

			longitudeFloat, err := converter.ToFloat64(lon.Interface())
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
		latitudeFloat, err := converter.ToFloat64(latitude)
		if err != nil {
			return nil, err
		}
		longitudeFloat, err := converter.ToFloat64(longitude)
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
