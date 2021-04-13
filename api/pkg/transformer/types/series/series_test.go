package series

import (
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
