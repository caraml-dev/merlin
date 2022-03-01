package series

import (
	"fmt"
	"reflect"

	"github.com/go-gota/gota/series"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

type Type string

const (
	String     Type = "string"
	Int        Type = "int"
	Float      Type = "float"
	Bool       Type = "bool"
	StringList Type = "string_list"
	IntList    Type = "int_list"
	FloatList  Type = "float_list"
	BoolList   Type = "bool_list"
)

var numericTypes = []Type{Int, Float}

type Series struct {
	series *series.Series
}

type contentType struct {
	hasFloat  bool
	hasInt    bool
	hasBool   bool
	hasString bool

	hasFloatList  bool
	hasIntList    bool
	hasBoolList   bool
	hasStringList bool
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

func (s *Series) IsNumeric() error {
	seriesType := s.Type()
	for _, sType := range numericTypes {
		if seriesType == sType {
			return nil
		}
	}
	return fmt.Errorf("this series type is not numeric but %s", seriesType)
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

// Append adds new elements to the end of the Series. When using Append, the
// Series is modified in place.
func (s *Series) Append(values interface{}) {
	s.series.Append(values)
}

// Concat concatenates two series together. It will return a new Series with the
// combined elements of both Series.
func (s *Series) Concat(x Series) *Series {
	concat := s.series.Concat(*x.series)
	return &Series{&concat}
}

// Order returns an sorted Series. NaN or nil elements are pushed to the
// end by order of appearance. Empty elements are pushed to the beginning by order of
// appearance. If reverse==true, the order is descending.
func (s *Series) Order(reverse bool) *Series {
	orderedIndex := s.series.Order(reverse)
	newOrder := make([]interface{}, s.series.Len())
	for i := 0; i < s.series.Len(); i++ {
		newOrder[i] = s.series.Elem(orderedIndex[i])
	}
	return New(newOrder, s.Type(), s.series.Name)
}

// StdDev calculates the standard deviation of a series.
// If a series is a list element type, flatten the series first.
func (s *Series) StdDev() float64 {
	return s.series.StdDev()
}

// Mean calculates the average value of a series.
// If a series is a list element type, flatten the series first.
func (s *Series) Mean() float64 {
	return s.series.Mean()
}

// Median calculates the middle or median value, as opposed to
// mean, and there is less susceptible to being affected by outliers.
// If a series is a list element type, flatten the series first.
func (s *Series) Median() float64 {
	return s.series.Median()
}

// Max return the biggest element in the series.
// If a series is a list element type, flatten the series first.
func (s *Series) Max() float64 {
	return s.series.Max()
}

// MaxStr return the biggest element in a series of type String.
// If a series is a list element type, flatten the series first.
func (s *Series) MaxStr() string {
	return s.series.MaxStr()
}

// Min return the lowest element in the series.
// If a series is a list element type, flatten the series first.
func (s *Series) Min() float64 {
	return s.series.Min()
}

// MinStr return the lowest element in a series of type String.
// If a series is a list element type, flatten the series first.
func (s *Series) MinStr() string {
	return s.series.MinStr()
}

// Quantile returns the sample of x such that x is greater than or
// equal to the fraction p of samples.
// Note: gonum/stat panics when called with strings.
// If a series is a list element type, flatten the series first.
func (s *Series) Quantile(p float64) float64 {
	return s.series.Quantile(p)
}

// Sum calculates the sum value of a series.
// If a series is a list element type, flatten the series first.
func (s *Series) Sum() float64 {
	return s.series.Sum()
}

// Flatten returns the flattened elements of series. If the series is list type (2D), it returns the standard type (1D).
// Examples:
// - Strings([]string{"A", "B", "C"}) -> Strings([]string{"A", "B", "C"})
// - IntsList([][]int{{1, 11}, {3, 33}}) -> Ints([]int{1, 11, 3, 33})
func (s *Series) Flatten() *Series {
	flatten := s.series.Flatten()
	return &Series{&flatten}
}

// Unique returns unique values based on a hash table.
// Examples:
// - Strings([]string{"A", "B", "C", "A", "B"}) -> Strings([]string{"A", "B", "C"})
// - IntsList([][]int{{1, 11}, {3, 33}, {3, 33}}) -> IntsList([][]int{{1, 11}, {3, 33}})
func (s *Series) Unique() *Series {
	unique := s.series.Unique()
	return &Series{&unique}
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
	case contentType.hasStringList:
		return StringList
	case contentType.hasBoolList:
		return BoolList
	case contentType.hasFloatList:
		return FloatList
	case contentType.hasIntList:
		return IntList
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
	case string:
		contentType.hasString = true
	case []float64, []float32:
		contentType.hasFloatList = true
	case []int, []int8, []int16, []int32, []int64:
		contentType.hasIntList = true
	case []bool:
		contentType.hasBoolList = true
	case []string:
		contentType.hasStringList = true
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
	case IntList:
		return converter.ToIntList(singleValue)
	case FloatList:
		return converter.ToFloat64List(singleValue)
	case BoolList:
		return converter.ToBoolList(singleValue)
	case StringList:
		return converter.ToStringList(singleValue)
	default:
		return nil, fmt.Errorf("unknown series type %s", seriesType)
	}
}
