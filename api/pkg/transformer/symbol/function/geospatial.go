package function

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/golang/geo/s2"
	"github.com/mmcloughlin/geohash"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

const (
	earthRadiusKm       = 6371 // radius of the earth in kilometers.
	pointFive           = 0.5
	zero                = 0
	minimumDistanceInKM = 0.001 // 1 meter

	KMDistanceUnit    = "km"
	MeterDistanceUnit = "m"
)

var (
	earthRadiusPerUnit = map[string]int{
		KMDistanceUnit:    earthRadiusKm,
		MeterDistanceUnit: earthRadiusKm * 1000,
	}
	geohashDirectionMapping = map[string]geohash.Direction{
		"north":     geohash.North,
		"northeast": geohash.NorthEast,
		"northwest": geohash.NorthWest,
		"south":     geohash.South,
		"southeast": geohash.SouthEast,
		"southwest": geohash.SouthWest,
		"west":      geohash.West,
		"east":      geohash.East,
	}
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

// Geohash convert latitude longitude to geohash value
func Geohash(lat interface{}, lon interface{}, precision uint) (interface{}, error) {
	rawLatLong, err := extractLatLong(lat, lon)
	if err != nil {
		panic(err)
	}

	switch latLong := rawLatLong.(type) {
	case []*LatLong:
		var value []interface{}
		for _, ll := range latLong {
			value = append(value, ll.toGeoHash(precision))
		}
		return value, nil
	case *LatLong:
		value := latLong.toGeoHash(precision)
		return value, nil
	default:
		return nil, fmt.Errorf("unknown type: %T", rawLatLong)
	}
}

// S2ID convert latitude and longitude to S2ID in certain level
func S2ID(lat interface{}, lon interface{}, level int) (interface{}, error) {
	rawLatLong, err := extractLatLong(lat, lon)
	if err != nil {
		panic(err)
	}

	switch latLong := rawLatLong.(type) {
	case []*LatLong:
		var value []interface{}
		for _, ll := range latLong {
			value = append(value, ll.toS2ID(level))
		}
		return value, nil
	case *LatLong:
		value := latLong.toS2ID(level)
		return value, nil
	default:
		return nil, fmt.Errorf("unknown type: %T", rawLatLong)
	}
}

// PolarAngle calculate polar angle between two points
func PolarAngle(lat1 interface{}, lon1 interface{}, lat2 interface{}, lon2 interface{}) (interface{}, error) {
	firstPoint, err := extractLatLong(lat1, lon1)
	if err != nil {
		return nil, err
	}

	secondPoint, err := extractLatLong(lat2, lon2)
	if err != nil {
		return nil, err
	}

	switch fp := firstPoint.(type) {
	case []*LatLong:
		secondPoints, ok := secondPoint.([]*LatLong)
		if !ok {
			return nil, errors.New("first point and second point has different format")
		}
		if len(secondPoints) != len(fp) {
			return nil, errors.New("both first point and second point arrays must have the same length")
		}
		var values []interface{}
		for idx, point1 := range fp {
			point2 := secondPoints[idx]
			values = append(values, calculatePolarAngle(point1, point2))
		}
		return values, nil

	case *LatLong:
		sp, ok := secondPoint.(*LatLong)
		if !ok {
			return nil, errors.New("first point and second point has different format")
		}
		value := calculatePolarAngle(fp, sp)
		return value, nil

	default:
		return nil, fmt.Errorf("unknown type: %T", firstPoint)
	}
}

// HaversineDistance calculate distance between to points (lat, lon) in km
// lat1, lon1, lat2, lon2 should be float64 or []float64 or any value that can be converted to those types (float64 or []float64)
func HaversineDistance(lat1 interface{}, lon1 interface{}, lat2 interface{}, lon2 interface{}, distanceUnit string) (interface{}, error) {
	firstPoint, err := extractLatLong(lat1, lon1)
	if err != nil {
		return nil, err
	}

	secondPoint, err := extractLatLong(lat2, lon2)
	if err != nil {
		return nil, err
	}

	return distanceBetweenPoints(firstPoint, secondPoint, distanceUnit)
}

func distanceBetweenPoints(firstPoint, secondPoint interface{}, distanceUnit string) (interface{}, error) {
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
			values = append(values, calculateDistance(point1, point2, distanceUnit))
		}
		return values, nil

	case *LatLong:
		fp := firstPoint.(*LatLong)
		sp, ok := secondPoint.(*LatLong)
		if !ok {
			return nil, errors.New("first point and second point has different format")
		}
		value := calculateDistance(fp, sp, distanceUnit)
		return value, nil

	default:
		return nil, fmt.Errorf("unknown type: %T", firstPoint)
	}
}

// GeohashDistance calculate distance from center of two geohashes
func GeohashDistance(firstGeohash interface{}, secondGeohash interface{}, distanceUnit string) (interface{}, error) {
	firstPoint, err := extractLatLongFromGeohash(firstGeohash)
	if err != nil {
		return nil, err
	}

	secondPoint, err := extractLatLongFromGeohash(secondGeohash)
	if err != nil {
		return nil, err
	}

	return distanceBetweenPoints(firstPoint, secondPoint, distanceUnit)
}

// GeohashNeighborForDirection finding neighbor given geohash and `direction`
func GeohashNeighborForDirection(targetGeohash interface{}, direction string) (interface{}, error) {
	geohashVal := reflect.ValueOf(targetGeohash)
	getNeighborFn := func(val interface{}, direction string) (string, error) {
		geohashStr, err := converter.ToString(val)
		if err != nil {
			return "", err
		}

		lowercaseDirection := strings.ToLower(direction)
		geohashDirection, ok := geohashDirectionMapping[lowercaseDirection]
		if !ok {
			return "", fmt.Errorf("direction '%s' is not valid", direction)
		}
		return geohash.Neighbor(geohashStr, geohashDirection), nil
	}

	var neighbor interface{}
	switch geohashVal.Kind() {
	case reflect.Slice:
		var neighborArr []string
		for index := 0; index < geohashVal.Len(); index++ {
			val := geohashVal.Index(index)
			neighborVal, err := getNeighborFn(val.Interface(), direction)
			if err != nil {
				return nil, err
			}
			neighborArr = append(neighborArr, neighborVal)
		}
		neighbor = neighborArr
	default:
		neighborVal, err := getNeighborFn(targetGeohash, direction)
		if err != nil {
			return nil, err
		}
		neighbor = neighborVal
	}
	return neighbor, nil
}

// GeohashAllNeighbors is finding all the neighbors from a geohash
func GeohashAllNeighbors(targetGeohash interface{}) (interface{}, error) {
	geohashVal := reflect.ValueOf(targetGeohash)
	getAllNeighborsFn := func(val interface{}) ([]string, error) {
		geohashStr, err := converter.ToString(val)
		if err != nil {
			return nil, err
		}
		return geohash.Neighbors(geohashStr), nil
	}

	var neighbors interface{}
	switch geohashVal.Kind() {
	case reflect.Slice:
		var neighborsArr [][]string
		for index := 0; index < geohashVal.Len(); index++ {
			val := geohashVal.Index(index)
			neighborsVal, err := getAllNeighborsFn(val.Interface())
			if err != nil {
				return nil, err
			}
			neighborsArr = append(neighborsArr, neighborsVal)
		}
		neighbors = neighborsArr
	default:
		neighborsVal, err := getAllNeighborsFn(targetGeohash)
		if err != nil {
			return nil, err
		}
		neighbors = neighborsVal
	}
	return neighbors, nil
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

func calculateDistance(firstPoint *LatLong, secondPoint *LatLong, distanceUnit string) float64 {
	lat1 := degreesToRadians(firstPoint.lat)
	lon1 := degreesToRadians(firstPoint.long)
	lat2 := degreesToRadians(secondPoint.lat)
	lon2 := degreesToRadians(secondPoint.long)

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	earthRadius, ok := earthRadiusPerUnit[distanceUnit]
	if !ok {
		earthRadius = earthRadiusKm
	}
	return c * float64(earthRadius)
}

func calculatePolarAngle(firstPoint *LatLong, secondPoint *LatLong) float64 {
	distance := calculateDistance(firstPoint, secondPoint, KMDistanceUnit)

	minimumDistance := minimumDistanceInKM
	if distance < minimumDistance {
		return zero
	}
	lat1 := degreesToRadians(firstPoint.lat)
	lon1 := degreesToRadians(firstPoint.long)
	lat2 := degreesToRadians(secondPoint.lat)
	lon2 := degreesToRadians(secondPoint.long)

	y := lat2 - lat1
	x := (lon2 - lon1) * math.Cos(pointFive*(lat2+lat1))
	return math.Atan2(y, x)
}

func extractLatLongFromGeohash(geohashValue interface{}) (interface{}, error) {
	var latLong interface{}

	pointConversionFn := func(val interface{}) (*LatLong, error) {
		geohashStr, err := converter.ToString(val)
		if err != nil {
			return nil, err
		}
		lat, lon := geohash.DecodeCenter(geohashStr)
		return &LatLong{
			lat:  lat,
			long: lon,
		}, nil
	}
	geohashVal := reflect.ValueOf(geohashValue)
	switch geohashVal.Kind() {
	case reflect.Slice:
		var latLongArray []*LatLong
		for index := 0; index < geohashVal.Len(); index++ {
			val := geohashVal.Index(index)
			point, err := pointConversionFn(val.Interface())
			if err != nil {
				return nil, err
			}
			latLongArray = append(latLongArray, point)
		}
		latLong = latLongArray
	default:
		point, err := pointConversionFn(geohashValue)
		if err != nil {
			return nil, err
		}
		latLong = point
	}
	return latLong, nil
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
