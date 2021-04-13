package series

import (
	"fmt"
	"reflect"

	"github.com/go-gota/gota/series"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

type Type string

const (
	String Type = "string"
	Int    Type = "int"
	Float  Type = "float"
	Bool   Type = "bool"
)

type Series struct {
	series *series.Series
}

func NewSeries(s *series.Series) *Series {
	return &Series{s}
}

func New(values interface{}, t Type, name string) *Series {
	s := series.New(values, series.Type(t), name)
	return &Series{&s}
}

func NewInferType(values interface{}, seriesName string) (*Series, error) {
	seriesType := detectType(values)
	seriesValues, err := castValues(values, seriesType)
	if err != nil {
		return nil, err
	}

	return New(seriesValues, seriesType, seriesName), nil
}

func (s *Series) Series() *series.Series {
	return s.series
}

func (s *Series) Type() Type {
	return Type(s.series.Type())
}

func detectType(values interface{}) Type {
	var hasBool, hasString, hasInt, hasFloat bool
	v := reflect.ValueOf(values)
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Interface() == nil {
				continue
			}

			switch v.Index(i).Interface().(type) {
			case float64, float32:
				hasFloat = true
			case int, int8, int16, int32, int64:
				hasInt = true
			case bool:
				hasBool = true
			default:
				hasString = true
			}
		}
	default:
		switch values.(type) {
		case float64, float32:
			hasFloat = true
		case int, int8, int16, int32, int64:
			hasInt = true
		case bool:
			hasBool = true
		default:
			hasString = true
		}
	}

	switch {
	case hasString:
		return String
	case hasBool:
		return Bool
	case hasFloat:
		return Float
	case hasInt:
		return Int
	default:
		return String
	}
}

func castValues(values interface{}, colType Type) ([]interface{}, error) {
	v := reflect.ValueOf(values)
	var seriesValues []interface{}
	switch v.Kind() {
	case reflect.Slice:
		seriesValues = make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			iVal := v.Index(i).Interface()
			if iVal == nil {
				seriesValues[i] = iVal
				continue
			}
			cVal, err := castValue(iVal, colType)
			if err != nil {
				return nil, err
			}

			seriesValues[i] = cVal
		}
	default:
		seriesValues = make([]interface{}, 1)
		iVal := v.Interface()
		if iVal == nil {
			return seriesValues, nil
		}

		v, err := castValue(iVal, colType)
		if err != nil {
			return nil, err
		}

		seriesValues[0] = v
	}

	return seriesValues, nil
}

func castValue(singleValue interface{}, seriesType Type) (interface{}, error) {
	switch seriesType {
	case Int:
		return converter.ToInt(singleValue)
	case Float:
		return converter.ToFloat64(singleValue)
	case Bool:
		return converter.ToBool(singleValue)
	case String:
		return converter.ToString(singleValue)
	default:
		return nil, fmt.Errorf("unknown series type %s", seriesType)
	}
}
