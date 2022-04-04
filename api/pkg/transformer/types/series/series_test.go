package series

import (
	"reflect"
	"testing"

	"github.com/go-gota/gota/series"
	"github.com/stretchr/testify/assert"
)

func TestSeries_New(t *testing.T) {
	s := New([]interface{}{"1111", nil}, String, "string_col")

	assert.Equal(t, String, s.Type())
	assert.Equal(t, *s.Series(), series.New([]interface{}{"1111", nil}, series.String, "string_col"))

	gotaSeries := series.New([]interface{}{"1111", nil}, series.String, "string_col")
	s2 := NewSeries(&gotaSeries)
	assert.Equal(t, String, s2.Type())
	assert.Equal(t, *s2.Series(), gotaSeries)
}

func TestSeries_NewInferType(t *testing.T) {
	type args struct {
		values     interface{}
		seriesName string
	}
	tests := []struct {
		name    string
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name: "single value string",
			args: args{
				values:     "1111",
				seriesName: "string_series",
			},
			want:    New([]interface{}{"1111"}, String, "string_series"),
			wantErr: false,
		},
		{
			name: "string array without nil",
			args: args{
				values:     []string{"1111", "2222", "3333", "4444"},
				seriesName: "string_series",
			},
			want:    New([]string{"1111", "2222", "3333", "4444"}, String, "string_series"),
			wantErr: false,
		},
		{
			name: "string array with nil",
			args: args{
				values:     []interface{}{"1111", "2222", "3333", "4444", nil},
				seriesName: "string_series",
			},
			want:    New([]interface{}{"1111", "2222", "3333", "4444", nil}, String, "string_series"),
			wantErr: false,
		},
		{
			name: "string array with nil and mixed type",
			args: args{
				values:     []interface{}{"1111", false, 33.33, 4444, nil},
				seriesName: "string_series",
			},
			want:    New([]interface{}{"1111", "false", "33.33", "4444", nil}, String, "string_series"),
			wantErr: false,
		},
		{
			name: "string list",
			args: args{
				values:     [][]string{{"1111", "2222"}, {"AAAA", "BBBB"}},
				seriesName: "string_list_series",
			},
			want:    New([]interface{}{[]string{"1111", "2222"}, []string{"AAAA", "BBBB"}}, StringList, "string_list_series"),
			wantErr: false,
		},
		{
			name: "string list via 1-D interface",
			args: args{
				values:     []interface{}{[]string{"1111", "2222"}, []string{"AAAA", "BBBB"}},
				seriesName: "string_list_series",
			},
			want:    New([]interface{}{[]string{"1111", "2222"}, []string{"AAAA", "BBBB"}}, StringList, "string_list_series"),
			wantErr: false,
		},
		{
			name: "single value int",
			args: args{
				values:     int64(1),
				seriesName: "int_series",
			},
			want:    New([]interface{}{1}, Int, "int_series"),
			wantErr: false,
		},
		{
			name: "int array without nil",
			args: args{
				values:     []int64{1111, 2222, 3333, 4444},
				seriesName: "int_series",
			},
			want:    New([]int{1111, 2222, 3333, 4444}, Int, "int_series"),
			wantErr: false,
		},
		{
			name: "int array with nil",
			args: args{
				values:     []interface{}{1111, 2222, 3333, 4444, nil},
				seriesName: "int_series",
			},
			want:    New([]interface{}{1111, 2222, 3333, 4444, nil}, Int, "int_series"),
			wantErr: false,
		},
		{
			name: "int array with nil and mixed type",
			args: args{
				values:     []interface{}{int8(11), int16(2222), int32(3333), int64(4444), nil},
				seriesName: "int_series",
			},
			want:    New([]interface{}{11, 2222, 3333, 4444, nil}, Int, "int_series"),
			wantErr: false,
		},
		{
			name: "int list",
			args: args{
				values:     [][]int{{1111, 2222}, {3333, 4444}},
				seriesName: "int_list_series",
			},
			want:    New([]interface{}{[]int{1111, 2222}, []int{3333, 4444}}, IntList, "int_list_series"),
			wantErr: false,
		},
		{
			name: "single value float",
			args: args{
				values:     float64(1.1),
				seriesName: "float_series",
			},
			want:    New([]interface{}{1.1}, Float, "float_series"),
			wantErr: false,
		},
		{
			name: "float array without nil",
			args: args{
				values:     []float64{1111.11, 2222.22, 3333.33, 4444.44},
				seriesName: "float_series",
			},
			want:    New([]float64{1111.11, 2222.22, 3333.33, 4444.44}, Float, "float_series"),
			wantErr: false,
		},
		{
			name: "float array with nil",
			args: args{
				values:     []interface{}{1111.11, 2222.22, 3333.33, 4444.44, nil},
				seriesName: "float_series",
			},
			want:    New([]interface{}{1111.11, 2222.22, 3333.33, 4444.44, nil}, Float, "float_series"),
			wantErr: false,
		},
		{
			name: "float array with nil and mixed types",
			args: args{
				values:     []interface{}{float32(1111), 2222.22, int8(1), int16(16), int32(1234), int64(555555), nil},
				seriesName: "float_series",
			},
			want:    New([]interface{}{1111, 2222.22, 1, 16, 1234, 555555, nil}, Float, "float_series"),
			wantErr: false,
		},
		{
			name: "double list",
			args: args{
				values:     [][]float64{{0.201, 3.14}, {4.56, 7.89}},
				seriesName: "double_list_series",
			},
			want:    New([]interface{}{[]float64{0.201, 3.14}, []float64{4.56, 7.89}}, FloatList, "double_list_series"),
			wantErr: false,
		},
		{
			name: "single value bool",
			args: args{
				values:     true,
				seriesName: "bool_series",
			},
			want:    New([]bool{true}, Bool, "bool_series"),
			wantErr: false,
		},
		{
			name: "bool array without nil",
			args: args{
				values:     []bool{true, false},
				seriesName: "bool_series",
			},
			want:    New([]bool{true, false}, Bool, "bool_series"),
			wantErr: false,
		},
		{
			name: "bool array with nil",
			args: args{
				values:     []interface{}{true, false, nil},
				seriesName: "bool_series",
			},
			want:    New([]interface{}{true, false, nil}, Bool, "bool_series"),
			wantErr: false,
		},
		{
			name: "bool list",
			args: args{
				values:     [][]bool{{true, true}, {false, false}},
				seriesName: "bool_list_series",
			},
			want:    New([]interface{}{[]bool{true, true}, []bool{false, false}}, BoolList, "bool_list_series"),
			wantErr: false,
		},
		// because the data type is not explicitly provided by the user, we cannot just guess the correct series type
		{
			name: "mixed data type in [][]interface",
			args: args{
				values:     [][]interface{}{{1111, 2.222}, {3333, 4444}, {"A", "B"}},
				seriesName: "string_series",
			},
			want:    New([]string{"[1111 2.222]", "[3333 4444]", "[A B]"}, String, "string_series"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewInferType(tt.args.values, tt.args.seriesName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewInferType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSeries_Append(t *testing.T) {
	tests := []struct {
		name   string
		series *Series
		values interface{}
		want   *Series
	}{
		{
			name:   "strings + strings",
			series: New([]string{"A", "B", "C"}, String, "string"),
			values: []interface{}{"X", "Y", "Z"},
			want:   New([]string{"A", "B", "C", "X", "Y", "Z"}, String, "string"),
		},
		{
			name:   "ints + ints",
			series: New([]int{1, 2, 3}, Int, "int"),
			values: []interface{}{7, 8, 9},
			want:   New([]int{1, 2, 3, 7, 8, 9}, Int, "int"),
		},
		{
			name:   "ints + ints in strings",
			series: New([]int{1, 2, 3}, Int, "int"),
			values: []interface{}{"7", "8", "9"},
			want:   New([]int{1, 2, 3, 7, 8, 9}, Int, "int"),
		},
		{
			name:   "strings list + strings list",
			series: New([][]string{{"A"}, {"B", "C"}}, StringList, "string_list"),
			values: []interface{}{[]string{"X", "Y"}, []string{"Z"}},
			want:   New([][]string{{"A"}, {"B", "C"}, {"X", "Y"}, {"Z"}}, StringList, "string_list"),
		},
		{
			name:   "ints list + ints list",
			series: New([][]int{{1}, {2, 3}}, IntList, "int_list"),
			values: []interface{}{[]int{7, 8}, []int{9}},
			want:   New([][]int{{1}, {2, 3}, {7, 8}, {9}}, IntList, "int_list"),
		},
		{
			name:   "ints + ints in strings",
			series: New([][]int{{1}, {2, 3}}, IntList, "int_list"),
			values: []interface{}{[]string{"7", "8"}, []int{9}},
			want:   New([][]int{{1}, {2, 3}, {7, 8}, {9}}, IntList, "int_list"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.series
			s.Append(tt.values)

			assert.Equal(t, tt.want, s)
		})
	}
}

func TestSeries_Concat(t *testing.T) {
	tests := []struct {
		name    string
		series1 *Series
		series2 *Series
		want    *Series
	}{
		{
			name:    "strings + strings",
			series1: New([]string{"A", "B", "C"}, String, "string"),
			series2: New([]string{"X", "Y", "Z"}, String, "string"),
			want:    New([]string{"A", "B", "C", "X", "Y", "Z"}, String, "string"),
		},
		{
			name:    "ints + ints",
			series1: New([]int{1, 2, 3}, Int, "int"),
			series2: New([]int{7, 8, 9}, Int, "int"),
			want:    New([]int{1, 2, 3, 7, 8, 9}, Int, "int"),
		},
		{
			name:    "ints + ints in strings",
			series1: New([]int{1, 2, 3}, Int, "int"),
			series2: New([]string{"7", "8", "9"}, Int, "int"),
			want:    New([]int{1, 2, 3, 7, 8, 9}, Int, "int"),
		},
		{
			name:    "strings list + strings list",
			series1: New([][]string{{"A"}, {"B", "C"}}, StringList, "string_list"),
			series2: New([][]string{{"X", "Y"}, {"Z"}}, StringList, "string_list"),
			want:    New([][]string{{"A"}, {"B", "C"}, {"X", "Y"}, {"Z"}}, StringList, "string_list"),
		},
		{
			name:    "ints list + ints list",
			series1: New([][]int{{1}, {2, 3}}, IntList, "int_list"),
			series2: New([][]int{{7, 8}, {9}}, IntList, "int_list"),
			want:    New([][]int{{1}, {2, 3}, {7, 8}, {9}}, IntList, "int_list"),
		},
		{
			name:    "ints + ints in strings",
			series1: New([][]int{{1}, {2, 3}}, IntList, "int_list"),
			series2: New([][]string{{"7", "8"}, {"9"}}, IntList, "int_list"),
			want:    New([][]int{{1}, {2, 3}, {7, 8}, {9}}, IntList, "int_list"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.series1.Concat(*tt.series2)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSeries_IsIn(t *testing.T) {
	type args struct {
		comparingValue interface{}
	}
	tests := []struct {
		name        string
		inputSeries *Series
		args        args
		want        *Series
	}{
		{
			name:        "some of record has value in comparing value",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: []int{5, 3, 9},
			},
			want: New([]bool{false, false, true, false, true}, Bool, ""),
		},
		{
			name:        "no records that has value in comparingValue",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: []int{9, 10, 11},
			},
			want: New([]bool{false, false, false, false, false}, Bool, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.inputSeries.IsIn(tt.args.comparingValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.IsIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toPointerInt(val int) *int {
	return &val
}

func TestSeries_Slice(t *testing.T) {
	type args struct {
		start *int
		end   *int
	}
	tests := []struct {
		name        string
		inputSeries *Series
		args        args
		want        *Series
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "start < end",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: toPointerInt(1),
				end:   toPointerInt(5),
			},
			want: New([]int{2, 3, 4, 5}, Int, ""),
		},
		{
			name:        "start exist but end nil",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: toPointerInt(2),
				end:   nil,
			},
			want: New([]int{3, 4, 5}, Int, ""),
		},
		{
			name:        "start nil but end exist",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: nil,
				end:   toPointerInt(3),
			},
			want: New([]int{1, 2, 3}, Int, ""),
		},
		{
			name:        "start is negative number",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: toPointerInt(-1),
				end:   nil,
			},
			want: New([]int{5}, Int, ""),
		},
		{
			name:        "end is negative number",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: nil,
				end:   toPointerInt(-2),
			},
			want: New([]int{1, 2, 3}, Int, ""),
		},
		{
			name:        "start is positive number, end is negative number",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: toPointerInt(1),
				end:   toPointerInt(-2),
			},
			want: New([]int{2, 3}, Int, ""),
		},
		{
			name:        "start > number of row",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: toPointerInt(6),
				end:   nil,
			},
			wantErr: true,
			errMsg:  "slice index out of bounds",
		},
		{
			name:        "start < -1 * number of row",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: toPointerInt(-6),
				end:   nil,
			},
			wantErr: true,
			errMsg:  "slice index out of bounds",
		},
		{
			name:        "end > number of row",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: nil,
				end:   toPointerInt(6),
			},
			wantErr: true,
			errMsg:  "slice index out of bounds",
		},
		{
			name:        "end < -1 * number of row",
			inputSeries: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				start: nil,
				end:   toPointerInt(-6),
			},
			wantErr: true,
			errMsg:  "slice index out of bounds",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.errMsg, func() {
					tt.inputSeries.Slice(tt.args.start, tt.args.end)
				})
				return
			}
			if got := tt.inputSeries.Slice(tt.args.start, tt.args.end); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Slice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Add(t *testing.T) {
	type args struct {
		right   *Series
		indexes *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "adding series with same length to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{1, 3, 5, 7, 9}, Int, ""),
		},
		{
			name:  "adding series broadcasted to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{2, 3, 4, 5, 6}, Int, ""),
		},
		{
			name:  "adding series broadcasted to all rows (broadcast indexes)",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true}, Bool, ""),
			},
			want: New([]int{2, 3, 4, 5, 6}, Int, ""),
		},
		{
			name:  "adding series with same length to certain rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			want: New([]interface{}{1, nil, nil, nil, 9}, Int, ""),
		},
		{
			name:  "adding series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not add boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Add(tt.args.right, tt.args.indexes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Substract(t *testing.T) {
	type args struct {
		right   *Series
		indexes *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "substract series with same length to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{1, 1, 1, 1, 1}, Int, ""),
		},
		{
			name:  "substract series broadcasted to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{0, 1, 2, 3, 4}, Int, ""),
		},
		{
			name:  "substract series broadcasted to all rows (broadcast indexes)",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true}, Bool, ""),
			},
			want: New([]int{0, 1, 2, 3, 4}, Int, ""),
		},
		{
			name:  "substract series with same length to certain rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			want: New([]interface{}{1, nil, nil, nil, 1}, Int, ""),
		},
		{
			name:  "substract series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not substract boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Substract(tt.args.right, tt.args.indexes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Substract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Substract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Multiply(t *testing.T) {
	type args struct {
		right   *Series
		indexes *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "multiply series with same length to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{0, 2, 6, 12, 20}, Int, ""),
		},
		{
			name:  "multiply series broadcasted to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{1, 2, 3, 4, 5}, Int, ""),
		},
		{
			name:  "multiply series broadcasted to all rows (broadcast indexes)",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true}, Bool, ""),
			},
			want: New([]int{1, 2, 3, 4, 5}, Int, ""),
		},
		{
			name:  "multiply series with same length to certain rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			want: New([]interface{}{0, nil, nil, nil, 20}, Int, ""),
		},
		{
			name:  "multiply series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not multiply boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Multiply(tt.args.right, tt.args.indexes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Multiply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Multiply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Divide(t *testing.T) {
	type args struct {
		right   *Series
		indexes *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "divide series with same length to all rows",
			input: New([]int{1, 2, 6, 8, 20}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 4, 5}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]interface{}{nil, 2, 3, 2, 4}, Int, ""),
		},
		{
			name:  "divide series broadcasted to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{1, 2, 3, 4, 5}, Int, ""),
		},
		{
			name:  "divide series broadcasted to all rows (broadcast indexes)",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true}, Bool, ""),
			},
			want: New([]int{1, 2, 3, 4, 5}, Int, ""),
		},
		{
			name:  "divide series with same length to certain rows",
			input: New([]int{1, 2, 6, 8, 20}, Int, ""),
			args: args{
				right:   New([]int{1, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			want: New([]interface{}{1, nil, nil, nil, 5}, Int, ""),
		},
		{
			name:  "divide series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not divide boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Divide(tt.args.right, tt.args.indexes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Divide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Divide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Modulo(t *testing.T) {
	type args struct {
		right   *Series
		indexes *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "modulo series with same length to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]interface{}{nil, 0, 1, 1, 1}, Int, ""),
		},
		{
			name:  "modulo series broadcasted to all rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true, true, true, true, true}, Bool, ""),
			},
			want: New([]int{0, 0, 0, 0, 0}, Int, ""),
		},
		{
			name:  "modulo series broadcasted to all rows (broadcast indexes)",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{1}, Int, ""),
				indexes: New([]bool{true}, Bool, ""),
			},
			want: New([]int{0, 0, 0, 0, 0}, Int, ""),
		},
		{
			name:  "substract series with same length to certain rows",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			want: New([]interface{}{nil, nil, nil, nil, 1}, Int, ""),
		},
		{
			name:  "modulo series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not modulo boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				right:   New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
				indexes: New([]bool{true, false, false, false, true}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Modulo(tt.args.right, tt.args.indexes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Modulo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Modulo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Greater(t *testing.T) {
	type args struct {
		comparingValue interface{}
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 10, 3, 10}, Int, ""),
			},
			want: New([]bool{true, true, false, true, false}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1}, Int, ""),
			},
			want: New([]bool{false, true, true, true, true}, Bool, ""),
		},
		{
			name:  "compare series with scalar",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: 3.5,
			},
			want: New([]interface{}{false, false, false, true, true}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not compare boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Greater(tt.args.comparingValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Greater() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Greater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_GreaterEq(t *testing.T) {
	type args struct {
		comparingValue interface{}
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1, 1, 10, 3, 10}, Int, ""),
			},
			want: New([]bool{true, true, false, true, false}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1}, Int, ""),
			},
			want: New([]bool{true, true, true, true, true}, Bool, ""),
		},
		{
			name:  "compare series with scalar",
			input: New([]float32{1, 2, 3, 4, 5}, Float, ""),
			args: args{
				comparingValue: 3.5,
			},
			want: New([]interface{}{false, false, false, true, true}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not compare boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.GreaterEq(tt.args.comparingValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.GreaterEq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.GreaterEq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Less(t *testing.T) {
	type args struct {
		comparingValue interface{}
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 10, 3, 10}, Int, ""),
			},
			want: New([]bool{false, false, true, false, true}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{3}, Int, ""),
			},
			want: New([]bool{true, true, false, false, false}, Bool, ""),
		},
		{
			name:  "compare series with scalar",
			input: New([]float32{1, 2, 3, 4, 5}, Float, ""),
			args: args{
				comparingValue: 3.5,
			},
			want: New([]interface{}{true, true, true, false, false}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not compare boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Less(tt.args.comparingValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Less() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_LessEq(t *testing.T) {
	type args struct {
		comparingValue interface{}
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1, 2, 10, 3, 10}, Int, ""),
			},
			want: New([]bool{true, true, true, false, true}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1}, Int, ""),
			},
			want: New([]bool{true, false, false, false, false}, Bool, ""),
		},
		{
			name:  "compare series with scalar",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: 3.5,
			},
			want: New([]interface{}{true, true, true, false, false}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not compare boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.LessEq(tt.args.comparingValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.LessEq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.LessEq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Eq(t *testing.T) {
	type args struct {
		comparingValue interface{}
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1, 2, 10, 3, 10}, Int, ""),
			},
			want: New([]bool{true, true, false, false, false}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1}, Int, ""),
			},
			want: New([]bool{true, false, false, false, false}, Bool, ""),
		},
		{
			name:  "compare series with scalar",
			input: New([]float32{1, 2, 3.5, 4, 5}, Float, ""),
			args: args{
				comparingValue: 3.5,
			},
			want: New([]interface{}{false, false, true, false, false}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not compare boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Eq(tt.args.comparingValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Eq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Eq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Neq(t *testing.T) {
	type args struct {
		comparingValue interface{}
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1, 2, 10, 3, 10}, Int, ""),
			},
			want: New([]bool{false, false, true, true, true}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{1}, Int, ""),
			},
			want: New([]bool{false, true, true, true, true}, Bool, ""),
		},
		{
			name:  "compare series with scalar",
			input: New([]float32{1, 2, 3.5, 4, 5}, Float, ""),
			args: args{
				comparingValue: 3.5,
			},
			want: New([]interface{}{true, true, false, true, true}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]int{1, 2, 3, 4, 5}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
		{
			name:  "could not compare boolean series with integer series",
			input: New([]bool{true, true, true, true, true}, Int, ""),
			args: args{
				comparingValue: New([]int{0, 1, 2, 3, 4, 6}, Int, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Neq(tt.args.comparingValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Neq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Neq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_And(t *testing.T) {
	type args struct {
		right *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]bool{true, true, true, false, true}, Bool, ""),
			args: args{
				right: New([]bool{false, true, true, true, false}, Bool, ""),
			},
			want: New([]bool{false, true, true, false, false}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]bool{true, true, true, false, false}, Bool, ""),
			args: args{
				right: New([]bool{true}, Bool, ""),
			},
			want: New([]bool{true, true, true, false, false}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]bool{true, true, true, true, true}, Bool, ""),
			args: args{
				right: New([]bool{true, true, true, true, true, false}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.And(tt.args.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.And() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.And() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_XOr(t *testing.T) {
	type args struct {
		right *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]bool{true, true, true, false, false}, Bool, ""),
			args: args{
				right: New([]bool{false, true, true, true, false}, Bool, ""),
			},
			want: New([]bool{true, false, false, true, false}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]bool{true, true, true, false, false}, Bool, ""),
			args: args{
				right: New([]bool{true}, Bool, ""),
			},
			want: New([]bool{false, false, false, true, true}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]bool{true, true, true, true, true}, Bool, ""),
			args: args{
				right: New([]bool{true, true, true, true, true, false}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.XOr(tt.args.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.XOr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.XOr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_Or(t *testing.T) {
	type args struct {
		right *Series
	}
	tests := []struct {
		name    string
		input   *Series
		args    args
		want    *Series
		wantErr bool
	}{
		{
			name:  "compare series with same length",
			input: New([]bool{true, true, true, false, false}, Bool, ""),
			args: args{
				right: New([]bool{false, true, true, true, false}, Bool, ""),
			},
			want: New([]bool{true, true, true, true, false}, Bool, ""),
		},
		{
			name:  "compare series broadcasted",
			input: New([]bool{true, true, true, false, false}, Bool, ""),
			args: args{
				right: New([]bool{true}, Bool, ""),
			},
			want: New([]bool{true, true, true, true, true}, Bool, ""),
		},
		{
			name:  "compare series with different length",
			input: New([]bool{true, true, true, true, true}, Bool, ""),
			args: args{
				right: New([]bool{true, true, true, true, true, false}, Bool, ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Or(tt.args.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("Series.Or() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Series.Or() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeries_IsBoolean(t *testing.T) {
	tests := []struct {
		name  string
		input *Series
		want  bool
	}{
		{
			name:  "boolean series",
			input: New([]bool{true, true}, Bool, ""),
			want:  true,
		},
		{
			name:  "int series",
			input: New([]int{1, 2}, Int, ""),
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.IsBoolean(); got != tt.want {
				t.Errorf("Series.IsBoolean() = %v, want %v", got, tt.want)
			}
		})
	}
}
