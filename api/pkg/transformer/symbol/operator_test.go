package symbol

import (
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Greater(t *testing.T) {
	type args struct {
		left  interface{}
		right interface{}
	}
	tests := []struct {
		name       string
		sr         Registry
		args       args
		want       interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name: "compare int8 with int32",
			args: args{
				left:  int8(32),
				right: int32(20),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int8 with bool",
			args: args{
				left:  int8(32),
				right: true,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int8 and bool",
		},
		{
			name: "compare int64 with string",
			args: args{
				left:  int64(64),
				right: "abcde",
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int64 and string",
		},
		{
			name: "compare float64 with bool",
			args: args{
				left:  float64(64),
				right: false,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between float64 and bool",
		},
		{
			name: "compare float32 with float64",
			args: args{
				left:  float32(32),
				right: float64(64),
			},
			sr:   NewRegistry(),
			want: false,
		},
		{
			name: "compare int with *series",
			args: args{
				left:  3,
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, false, false}, series.Bool, ""),
		},
		{
			name: "compare float with bool series",
			args: args{
				left:  3.5,
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, true, false}, series.Bool, ""),
		},
		{
			name: "compare bool with bool",
			args: args{
				left:  true,
				right: false,
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "> operator is not supported",
		},
		{
			name: "compare float series with int series",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, false, true, true}, series.Bool, ""),
		},
		{
			name: "compare series with string",
			args: args{
				left:  series.New([]string{"a", "b", "c", "d"}, series.String, "string_series"),
				right: "a",
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, true, true, true}, series.Bool, ""),
		},
		{
			name: "compare series with series different dimension",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2, 5}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "can't compare: length mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.errMessage, func() {
					tt.sr.Greater(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.Greater(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.Greater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_GreaterEq(t *testing.T) {
	type args struct {
		left  interface{}
		right interface{}
	}
	tests := []struct {
		name       string
		sr         Registry
		args       args
		want       interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name: "compare int8 with int32",
			args: args{
				left:  int8(32),
				right: int32(20),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int8 with bool",
			args: args{
				left:  int8(32),
				right: true,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int8 and bool",
		},
		{
			name: "compare int64 with string",
			args: args{
				left:  int64(64),
				right: "abcde",
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int64 and string",
		},
		{
			name: "compare float64 with bool",
			args: args{
				left:  float64(64),
				right: false,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between float64 and bool",
		},
		{
			name: "compare float32 with float64",
			args: args{
				left:  float32(32),
				right: float64(64),
			},
			sr:   NewRegistry(),
			want: false,
		},
		{
			name: "compare int with *series",
			args: args{
				left:  3,
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, true, false}, series.Bool, ""),
		},
		{
			name: "compare bool with bool",
			args: args{
				left:  true,
				right: false,
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: ">= operator is not supported",
		},
		{
			name: "compare float series with int series",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, true, true}, series.Bool, ""),
		},
		{
			name: "compare series with string",
			args: args{
				left:  series.New([]string{"a", "b", "c", "d"}, series.String, "string_series"),
				right: "a",
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, true, true}, series.Bool, ""),
		},
		{
			name: "compare series with series different dimension",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2, 5}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "can't compare: length mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.errMessage, func() {
					tt.sr.GreaterEq(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.GreaterEq(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GreaterEq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_Less(t *testing.T) {
	type args struct {
		left  interface{}
		right interface{}
	}
	tests := []struct {
		name       string
		sr         Registry
		args       args
		want       interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name: "compare int8 with int32",
			args: args{
				left:  int8(10),
				right: int32(20),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int8 with bool",
			args: args{
				left:  int8(32),
				right: true,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int8 and bool",
		},
		{
			name: "compare int64 with string",
			args: args{
				left:  int64(64),
				right: "abcde",
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int64 and string",
		},
		{
			name: "compare float64 with bool",
			args: args{
				left:  float64(64),
				right: false,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between float64 and bool",
		},
		{
			name: "compare float32 with float64",
			args: args{
				left:  float32(64),
				right: float64(32),
			},
			sr:   NewRegistry(),
			want: false,
		},
		{
			name: "compare int with *series",
			args: args{
				left:  3,
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, false, false, true}, series.Bool, ""),
		},
		{
			name: "compare bool with bool",
			args: args{
				left:  true,
				right: false,
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "< operator is not supported",
		},
		{
			name: "compare float series with int series",
			args: args{
				left:  series.New([]float64{1, 2, 3, 4.2}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 4, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, false, true, false}, series.Bool, ""),
		},
		{
			name: "compare series with string",
			args: args{
				left:  series.New([]string{"a", "b", "c", "d"}, series.String, "string_series"),
				right: "c",
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, false, false}, series.Bool, ""),
		},
		{
			name: "compare series with series different dimension",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2, 5}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "can't compare: length mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.errMessage, func() {
					tt.sr.Less(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.Less(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_LessThan(t *testing.T) {
	type args struct {
		left  interface{}
		right interface{}
	}
	tests := []struct {
		name       string
		sr         Registry
		args       args
		want       interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name: "compare int8 with int32",
			args: args{
				left:  int8(10),
				right: int32(20),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int8 with bool",
			args: args{
				left:  int8(32),
				right: true,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int8 and bool",
		},
		{
			name: "compare int64 with string",
			args: args{
				left:  int64(64),
				right: "abcde",
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int64 and string",
		},
		{
			name: "compare float64 with bool",
			args: args{
				left:  float64(64),
				right: false,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between float64 and bool",
		},
		{
			name: "compare float32 with float64",
			args: args{
				left:  float32(64),
				right: float64(32),
			},
			sr:   NewRegistry(),
			want: false,
		},
		{
			name: "compare int with *series",
			args: args{
				left:  3,
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, false, true, true}, series.Bool, ""),
		},
		{
			name: "compare bool with bool",
			args: args{
				left:  true,
				right: false,
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "<= operator is not supported",
		},
		{
			name: "compare float series with int series",
			args: args{
				left:  series.New([]float64{1, 2, 3, 4.2}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 4, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, true, false}, series.Bool, ""),
		},
		{
			name: "compare series with string",
			args: args{
				left:  series.New([]string{"a", "b", "c", "d"}, series.String, "string_series"),
				right: "c",
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, true, false}, series.Bool, ""),
		},
		{
			name: "compare series with series different dimension",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2, 5}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "can't compare: length mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.errMessage, func() {
					tt.sr.LessEq(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.LessEq(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_Equal(t *testing.T) {
	type args struct {
		left  interface{}
		right interface{}
	}
	tests := []struct {
		name       string
		sr         Registry
		args       args
		want       interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name: "compare int8 with int32",
			args: args{
				left:  int8(10),
				right: int32(10),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int8 with bool",
			args: args{
				left:  int8(32),
				right: true,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int8 and bool",
		},
		{
			name: "compare int64 with string",
			args: args{
				left:  int64(64),
				right: "abcde",
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int64 and string",
		},
		{
			name: "compare float64 with bool",
			args: args{
				left:  float64(64),
				right: false,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between float64 and bool",
		},
		{
			name: "compare float32 with float64",
			args: args{
				left:  float32(64),
				right: float64(64),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int with *series",
			args: args{
				left:  3,
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, false, true, false}, series.Bool, ""),
		},
		{
			name: "compare bool with bool",
			args: args{
				left:  true,
				right: true,
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare bool with bool, false",
			args: args{
				left:  true,
				right: false,
			},
			sr:   NewRegistry(),
			want: false,
		},
		{
			name: "compare float series with int series",
			args: args{
				left:  series.New([]float64{1, 2, 3, 4.2}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 4, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, false, false}, series.Bool, ""),
		},
		{
			name: "compare series with string",
			args: args{
				left:  series.New([]string{"a", "b", "c", "d"}, series.String, "string_series"),
				right: "c",
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, false, true, false}, series.Bool, ""),
		},
		{
			name: "compare series with series different dimension",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2, 5}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "can't compare: length mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.errMessage, func() {
					tt.sr.Equal(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.Equal(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_Neq(t *testing.T) {
	type args struct {
		left  interface{}
		right interface{}
	}
	tests := []struct {
		name       string
		sr         Registry
		args       args
		want       interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name: "compare int8 with int32",
			args: args{
				left:  int8(10),
				right: int32(9),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int8 with bool",
			args: args{
				left:  int8(32),
				right: true,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int8 and bool",
		},
		{
			name: "compare int64 with string",
			args: args{
				left:  int64(64),
				right: "abcde",
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between int64 and string",
		},
		{
			name: "compare float64 with bool",
			args: args{
				left:  float64(64),
				right: false,
			},
			sr:         NewRegistry(),
			want:       false,
			wantErr:    true,
			errMessage: "comparison not supported between float64 and bool",
		},
		{
			name: "compare float64 with float64",
			args: args{
				left:  float32(32),
				right: float64(64),
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare int with *series",
			args: args{
				left:  3,
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, false, true}, series.Bool, ""),
		},
		{
			name: "compare bool with bool",
			args: args{
				left:  true,
				right: true,
			},
			sr:   NewRegistry(),
			want: false,
		},
		{
			name: "compare bool with bool, false",
			args: args{
				left:  true,
				right: false,
			},
			sr:   NewRegistry(),
			want: true,
		},
		{
			name: "compare float series with int series",
			args: args{
				left:  series.New([]float64{1, 2, 3, 4.2}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 4, 4}, series.Int, "int_series"),
			},
			sr:   NewRegistry(),
			want: series.New([]bool{false, false, true, true}, series.Bool, ""),
		},
		{
			name: "compare series with string",
			args: args{
				left:  series.New([]string{"a", "b", "c", "d"}, series.String, "string_series"),
				right: "c",
			},
			sr:   NewRegistry(),
			want: series.New([]bool{true, true, false, true}, series.Bool, ""),
		},
		{
			name: "compare series with series different dimension",
			args: args{
				left:  series.New([]float64{1, 2, 3.2, 4.2, 5}, series.Float, "float_series"),
				right: series.New([]int{1, 2, 3, 4}, series.Int, "int_series"),
			},
			sr:         NewRegistry(),
			wantErr:    true,
			errMessage: "can't compare: length mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.errMessage, func() {
					tt.sr.Neq(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.Neq(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.Neq() = %v, want %v", got, tt.want)
			}
		})
	}
}
