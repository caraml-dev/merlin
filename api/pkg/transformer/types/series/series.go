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

type contentType struct {
	hasFloat  bool
	hasInt    bool
	hasBool   bool
	hasString bool
}

func NewSeries(s *series.Series) *Series {
	return &Series{s}
}

func New(values interface{}, t Type, name string) *Series {
	s := series.New(values, series.Type(t), name)
	return &Series{&s}
}

func NewInferType(values interface{}, seriesName string) (*Series, error) {
	s, ok := values.(*Series)
	if ok {
		newSeries := s.Series().Copy()
		newSeries.Name = seriesName
		return NewSeries(&newSeries), nil
	}

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

func (s *Series) GetRecords() []interface{} {
	genericArr := make([]interface{}, s.series.Len())
	for i := 0; i < s.series.Len(); i++ {
		genericArr[i] = s.series.Val(i)
	}

	return genericArr
}

func (s *Series) Get(index int) interface{} {
	return s.series.Elem(index).Val()
}

func detectType(values interface{}) Type {
	contentType := &contentType{}
	v := reflect.ValueOf(values)
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Interface() == nil {
				continue
			}

			contentType = hasType(v.Index(i).Interface(), contentType)
		}
	default:
		contentType = hasType(values, contentType)
	}

	switch {
	case contentType.hasString:
		return String
	case contentType.hasBool:
		return Bool
	case contentType.hasFloat:
		return Float
	case contentType.hasInt:
		return Int
	default:
		return String
	}
}

func hasType(value interface{}, contentType *contentType) *contentType {
	switch value.(type) {
	case float64, float32:
		contentType.hasFloat = true
	case int, int8, int16, int32, int64:
		contentType.hasInt = true
	case bool:
		contentType.hasBool = true
	default:
		contentType.hasString = true
	}
	return contentType
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
