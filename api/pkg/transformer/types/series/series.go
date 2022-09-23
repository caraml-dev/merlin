package series

import (
	"fmt"
	"reflect"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/go-gota/gota/series"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

type (
	Type       string
	Comparator string
)

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

// NormalizeIndex to find correct start and end index of a series
// if index is negative, the index value will be counted backward
func NormalizeIndex(start, end *int, numOfRow int) (startIdx, endIdx int) {
	startIdx = 0
	if start != nil {
		startIdx = *start
	}
	endIdx = numOfRow
	if end != nil {
		endIdx = *end
	}

	if startIdx < 0 {
		startIdx = numOfRow + startIdx
	}
	if endIdx < 0 {
		endIdx = numOfRow + endIdx
	}

	return startIdx, endIdx
}

// ConvertToUPIColumns convert list of columns/series into list of upi column
func ConvertToUPIColumns(cols []*Series) ([]*upiv1.Column, error) {
	upiCols := make([]*upiv1.Column, len(cols))
	for idx, col := range cols {
		upiCol, err := col.toUPIColumn()
		if err != nil {
			return nil, err
		}
		upiCols[idx] = upiCol
	}
	return upiCols, nil
}

func (s *Series) toUPIColumn() (*upiv1.Column, error) {
	var upiColType upiv1.Type
	switch s.Type() {
	case Int:
		upiColType = upiv1.Type_TYPE_INTEGER
	case Float:
		upiColType = upiv1.Type_TYPE_DOUBLE
	case String:
		upiColType = upiv1.Type_TYPE_STRING
	default:
		return nil, fmt.Errorf("type %v is not supported in UPI", s.Type())
	}

	return &upiv1.Column{
		Name: s.series.Name,
		Type: upiColType,
	}, nil
}

func (s *Series) Series() *series.Series {
	return s.series
}

func (s *Series) Type() Type {
	return Type(s.series.Type())
}

func (s *Series) Len() int {
	return s.series.Len()
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

func (s *Series) And(right *Series) (*Series, error) {
	newSeries := s.series.And(right.series)
	if newSeries.Err != nil {
		return nil, newSeries.Err
	}
	return &Series{&newSeries}, nil
}

func (s *Series) XOr(right *Series) (*Series, error) {
	newSeries := s.series.XOr(right.series)
	if newSeries.Err != nil {
		return nil, newSeries.Err
	}
	return &Series{&newSeries}, nil
}

func (s *Series) Or(right *Series) (*Series, error) {
	newSeries := s.series.Or(right.series)
	if newSeries.Err != nil {
		return nil, newSeries.Err
	}
	return &Series{&newSeries}, nil
}

func getIndexes(indexes *Series) interface{} {
	if indexes == nil {
		return nil
	}
	if indexes.series == nil {
		return nil
	}
	return *indexes.series
}

func (s *Series) Add(right *Series, indexes *Series) (*Series, error) {
	res := s.series.Add(*right.series, getIndexes(indexes))
	if res.Err != nil {
		return nil, res.Err
	}
	return &Series{&res}, nil
}

func (s *Series) Substract(right *Series, indexes *Series) (*Series, error) {
	res := s.series.Substract(*right.series, getIndexes(indexes))
	if res.Err != nil {
		return nil, res.Err
	}
	return &Series{&res}, nil
}

func (s *Series) Multiply(right *Series, indexes *Series) (*Series, error) {
	res := s.series.Multiply(*right.series, getIndexes(indexes))
	if res.Err != nil {
		return nil, res.Err
	}
	return &Series{&res}, nil
}

func (s *Series) Divide(right *Series, indexes *Series) (*Series, error) {
	res := s.series.Divide(*right.series, getIndexes(indexes))
	if res.Err != nil {
		return nil, res.Err
	}
	return &Series{&res}, nil
}

func (s *Series) Modulo(right *Series, indexes *Series) (*Series, error) {
	res := s.series.Modulo(*right.series, getIndexes(indexes))
	if res.Err != nil {
		return nil, res.Err
	}
	return &Series{&res}, nil
}

func (s *Series) Greater(comparingValue interface{}) (*Series, error) {
	return s.compare(series.Greater, comparingValue)
}

func (s *Series) GreaterEq(comparingValue interface{}) (*Series, error) {
	return s.compare(series.GreaterEq, comparingValue)
}

func (s *Series) Less(comparingValue interface{}) (*Series, error) {
	return s.compare(series.Less, comparingValue)
}

func (s *Series) LessEq(comparingValue interface{}) (*Series, error) {
	return s.compare(series.LessEq, comparingValue)
}

func (s *Series) Eq(comparingValue interface{}) (*Series, error) {
	return s.compare(series.Eq, comparingValue)
}

func (s *Series) Neq(comparingValue interface{}) (*Series, error) {
	return s.compare(series.Neq, comparingValue)
}

func (s *Series) compare(comparator series.Comparator, comparingValue interface{}) (*Series, error) {
	var result series.Series
	switch cVal := comparingValue.(type) {
	case Series:
		result = s.series.Compare(comparator, cVal.series)
	case *Series:
		if cVal == nil {
			result = s.series.Compare(comparator, nil)
		} else {
			result = s.series.Compare(comparator, cVal.series)
		}
	default:
		result = s.series.Compare(comparator, cVal)
	}
	if result.Err != nil {
		return nil, result.Err
	}
	return NewSeries(&result), nil
}

// Slice series based on start and end index
// the result will include start index but exclude end index
func (s *Series) Slice(start, end *int) *Series {
	startIdx, endIdx := NormalizeIndex(start, end, s.Len())
	sliced := s.series.Slice(startIdx, endIdx)
	if sliced.Error() != nil {
		panic(sliced.Error())
	}
	return NewSeries(&sliced)
}

func (s *Series) IsBoolean() bool {
	return s.Type() == Bool
}

// IsIn check whether whether row has value that part of `comparingValue`
func (s *Series) IsIn(comparingValue interface{}) *Series {
	result, err := s.compare(series.In, comparingValue)
	if err != nil {
		panic(err)
	}
	return result
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
