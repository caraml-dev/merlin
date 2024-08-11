package converter

import (
	"encoding/base64"
	"math"
	"reflect"
	"strings"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	types "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
)

func TestToBool(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "from bool",
			args: args{
				v: true,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				v: "true",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: 0,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: 1,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: 3,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "from float",
			args: args{
				v: 1.1,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "from float",
			args: args{
				v: float64(1.0),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "from float",
			args: args{
				v: float64(0.0),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToBool(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToFloat32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    float32
		wantErr bool
	}{
		{
			name: "from float64",
			args: args{
				v: float64(11.111),
			},
			want:    float32(11.111),
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				v: float32(11.111),
			},
			want:    float32(11.111),
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				v: float32(11.111),
			},
			want:    float32(11.111),
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: int(1111),
			},
			want:    float32(1111),
			wantErr: false,
		},
		{
			name: "from int8",
			args: args{
				v: int8(111),
			},
			want:    float32(111),
			wantErr: false,
		},
		{
			name: "from int16",
			args: args{
				v: int16(11111),
			},
			want:    float32(11111),
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				v: int32(11111),
			},
			want:    float32(11111),
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				v: int64(11111),
			},
			want:    float32(11111),
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				v: "1111.1111",
			},
			want:    float32(1111.1111),
			wantErr: false,
		},
		{
			name: "from string - invalid",
			args: args{
				v: "asd",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "from bool",
			args: args{
				v: false,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat32(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.InEpsilon(t, tt.want, got, 0.0001)
		})
	}
}

func TestToFloat64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "from float64",
			args: args{
				v: float64(11.111),
			},
			want:    float64(11.111),
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				v: float32(11.111),
			},
			want:    float64(11.111),
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: int(1111),
			},
			want:    float64(1111),
			wantErr: false,
		},
		{
			name: "from int8",
			args: args{
				v: int8(111),
			},
			want:    float64(111),
			wantErr: false,
		},
		{
			name: "from int16",
			args: args{
				v: int16(11111),
			},
			want:    float64(11111),
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				v: int32(11111),
			},
			want:    float64(11111),
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				v: int64(11111),
			},
			want:    float64(11111),
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				v: "1111.1111",
			},
			want:    float64(1111.1111),
			wantErr: false,
		},
		{
			name: "from bool",
			args: args{
				v: false,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat64(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.InEpsilon(t, tt.want, got, 0.0001)
		})
	}
}

func TestToInt(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "from float64",
			args: args{
				v: float64(1111.111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				v: float32(1111.111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: int(1111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int8",
			args: args{
				v: int8(11),
			},
			want:    11,
			wantErr: false,
		},
		{
			name: "from int16",
			args: args{
				v: int16(1111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				v: int32(111111),
			},
			want:    111111,
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				v: int64(111111),
			},
			want:    111111,
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				v: "11111",
			},
			want:    11111,
			wantErr: false,
		},
		{
			name: "from invalid string",
			args: args{
				v: "hello",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "from bool",
			args: args{
				v: true,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestToInt64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "from float64",
			args: args{
				v: float64(1111.111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				v: float32(1111.111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: int(1111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int8",
			args: args{
				v: int8(11),
			},
			want:    11,
			wantErr: false,
		},
		{
			name: "from int16",
			args: args{
				v: int16(1111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				v: int32(111111),
			},
			want:    111111,
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				v: int64(111111),
			},
			want:    111111,
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				v: "9223372036854775807",
			},
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			name: "from invalid string",
			args: args{
				v: "hello",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "from bool",
			args: args{
				v: true,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt64(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestToInt32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{
			name: "from float64",
			args: args{
				v: float64(1111.111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				v: float32(1111.111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: int(1111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int8",
			args: args{
				v: int8(11),
			},
			want:    11,
			wantErr: false,
		},
		{
			name: "from int16",
			args: args{
				v: int16(1111),
			},
			want:    1111,
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				v: int32(111111),
			},
			want:    111111,
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				v: int64(111111),
			},
			want:    111111,
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				v: "1234",
			},
			want:    1234,
			wantErr: false,
		},
		{
			name: "from invalid string",
			args: args{
				v: "hello",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "from bool",
			args: args{
				v: true,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt32(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestToString(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "from float64",
			args: args{
				v: float64(11111111.111),
			},
			want:    "1.1111111111e+07",
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				v: float32(1111.111),
			},
			want:    "1111.111",
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				v: int(1111),
			},
			want:    "1111",
			wantErr: false,
		},
		{
			name: "from int8",
			args: args{
				v: int8(11),
			},
			want:    "11",
			wantErr: false,
		},
		{
			name: "from int16",
			args: args{
				v: int16(1111),
			},
			want:    "1111",
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				v: int32(111111),
			},
			want:    "111111",
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				v: int64(1111111111),
			},
			want:    "1111111111",
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				v: "11111",
			},
			want:    "11111",
			wantErr: false,
		},
		{
			name: "from bool",
			args: args{
				v: true,
			},
			want:    "true",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToString(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestExtractFeastValue(t *testing.T) {
	tests := []struct {
		name        string
		val         *types.Value
		want        interface{}
		wantValType types.ValueType_Enum
		wantErr     bool
	}{
		{
			name:        "string",
			val:         feast.StrVal("hello"),
			want:        "hello",
			wantValType: types.ValueType_STRING,
			wantErr:     false,
		},
		{
			name:        "double",
			val:         feast.DoubleVal(123456789.123456789),
			want:        123456789.123456789,
			wantValType: types.ValueType_DOUBLE,
			wantErr:     false,
		},
		{
			name:        "float",
			val:         feast.FloatVal(1.1),
			want:        float32(1.1),
			wantValType: types.ValueType_FLOAT,
			wantErr:     false,
		},
		{
			name:        "int32",
			val:         feast.Int32Val(1234),
			want:        int32(1234),
			wantValType: types.ValueType_INT32,
			wantErr:     false,
		},
		{
			name:        "int64",
			val:         feast.Int64Val(12345678),
			want:        int64(12345678),
			wantValType: types.ValueType_INT64,
			wantErr:     false,
		},
		{
			name:        "bool",
			val:         feast.BoolVal(true),
			want:        true,
			wantValType: types.ValueType_BOOL,
			wantErr:     false,
		},
		{
			name:        "bytes",
			val:         feast.BytesVal([]byte("hello")),
			want:        base64.StdEncoding.EncodeToString([]byte("hello")),
			wantValType: types.ValueType_STRING,
			wantErr:     false,
		},
		{
			name: "string list",
			val: &types.Value{Val: &types.Value_StringListVal{StringListVal: &types.StringList{Val: []string{
				"hello",
				"world",
			}}}},
			want: []string{
				"hello",
				"world",
			},
			wantValType: types.ValueType_STRING_LIST,
			wantErr:     false,
		},
		{
			name: "double list",
			val: &types.Value{Val: &types.Value_DoubleListVal{DoubleListVal: &types.DoubleList{Val: []float64{
				123.45,
				123.45,
			}}}},
			want: []float64{
				123.45,
				123.45,
			},
			wantValType: types.ValueType_DOUBLE_LIST,
			wantErr:     false,
		},
		{
			name: "float list",
			val: &types.Value{Val: &types.Value_FloatListVal{FloatListVal: &types.FloatList{Val: []float32{
				123.45,
				123.45,
			}}}},
			want: []float32{
				123.45,
				123.45,
			},
			wantValType: types.ValueType_FLOAT_LIST,
			wantErr:     false,
		},
		{
			name: "int32 list",
			val: &types.Value{Val: &types.Value_Int32ListVal{Int32ListVal: &types.Int32List{Val: []int32{
				int32(1234),
				int32(1234),
			}}}},
			want: []int32{
				1234,
				1234,
			},
			wantValType: types.ValueType_INT32_LIST,
			wantErr:     false,
		},
		{
			name: "int64 list",
			val: &types.Value{Val: &types.Value_Int64ListVal{Int64ListVal: &types.Int64List{Val: []int64{
				int64(1234),
				int64(1234),
			}}}},
			want: []int64{
				1234,
				1234,
			},
			wantValType: types.ValueType_INT64_LIST,
			wantErr:     false,
		},
		{
			name: "bool list",
			val: &types.Value{Val: &types.Value_BoolListVal{BoolListVal: &types.BoolList{Val: []bool{
				true,
				false,
			}}}},
			want: []bool{
				true,
				false,
			},
			wantValType: types.ValueType_BOOL_LIST,
			wantErr:     false,
		},
		{
			name: "bytes list",
			val: &types.Value{Val: &types.Value_BytesListVal{BytesListVal: &types.BytesList{Val: [][]byte{
				[]byte("hello"),
				[]byte("world"),
			}}}},
			want: []string{
				base64.StdEncoding.EncodeToString([]byte("hello")),
				base64.StdEncoding.EncodeToString([]byte("world")),
			},
			wantValType: types.ValueType_STRING_LIST,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotValType, err := ExtractFeastValue(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFeastValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractFeastValue() got = %v, want %v", got, tt.want)
			}
			assert.Equal(t, tt.wantValType, gotValType)
		})
	}
}

func TestToStringList(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "from string",
			args: args{
				val: "hello",
			},
			want:    []string{"hello"},
			wantErr: false,
		},
		{
			name: "from strings - 1",
			args: args{
				val: []string{"hello"},
			},
			want:    []string{"hello"},
			wantErr: false,
		},
		{
			name: "from strings - 2",
			args: args{
				val: []string{"hello", "world"},
			},
			want:    []string{"hello", "world"},
			wantErr: false,
		},
		{
			name: "from float64",
			args: args{
				val: float64(11111111.111),
			},
			want:    []string{"1.1111111111e+07"},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(1111.111),
			},
			want:    []string{"1111.111"},
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				val: int(1111),
			},
			want:    []string{"1111"},
			wantErr: false,
		},
		{
			name: "from int8",
			args: args{
				val: int8(11),
			},
			want:    []string{"11"},
			wantErr: false,
		},
		{
			name: "from int16",
			args: args{
				val: int16(1111),
			},
			want:    []string{"1111"},
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				val: int32(111111),
			},
			want:    []string{"111111"},
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				val: int64(1111111111),
			},
			want:    []string{"1111111111"},
			wantErr: false,
		},
		{
			name: "from bool",
			args: args{
				val: true,
			},
			want:    []string{"true"},
			wantErr: false,
		},
		{
			name: "from []float64",
			args: args{
				val: []float64{float64(11111111.111), float64(21111111.111)},
			},
			want:    []string{"1.1111111111e+07", "2.1111111111e+07"},
			wantErr: false,
		},
		{
			name: "from []float32",
			args: args{
				val: []float32{float32(1111.111), float32(2111.111)},
			},
			want:    []string{"1111.111", "2111.111"},
			wantErr: false,
		},
		{
			name: "from []int",
			args: args{
				val: []int{int(1111), int(2111)},
			},
			want:    []string{"1111", "2111"},
			wantErr: false,
		},
		{
			name: "from []int8",
			args: args{
				val: []int8{int8(11), int8(21)},
			},
			want:    []string{"11", "21"},
			wantErr: false,
		},
		{
			name: "from []int16",
			args: args{
				val: []int16{int16(1111), int16(2111)},
			},
			want:    []string{"1111", "2111"},
			wantErr: false,
		},
		{
			name: "from []int32",
			args: args{
				val: []int32{int32(111111), int32(211111)},
			},
			want:    []string{"111111", "211111"},
			wantErr: false,
		},
		{
			name: "from []int64",
			args: args{
				val: []int64{int64(1111111111), int64(2111111111)},
			},
			want:    []string{"1111111111", "2111111111"},
			wantErr: false,
		},
		{
			name: "from []bool",
			args: args{
				val: []bool{true, false},
			},
			want:    []string{"true", "false"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringList(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStringList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToStringList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToIntList(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{
			name: "from int",
			args: args{
				val: int(1),
			},
			want:    []int{1},
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				val: int32(1),
			},
			want:    []int{1},
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				val: int64(1),
			},
			want:    []int{1},
			wantErr: false,
		},
		{
			name: "from []int",
			args: args{
				val: []int{int(1), int(2)},
			},
			want:    []int{1, 2},
			wantErr: false,
		},
		{
			name: "from []int32",
			args: args{
				val: []int32{int32(1), int32(2)},
			},
			want:    []int{1, 2},
			wantErr: false,
		},
		{
			name: "from []int64",
			args: args{
				val: []int64{int64(1), int64(2)},
			},
			want:    []int{1, 2},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(3.14),
			},
			want:    []int{3},
			wantErr: false,
		},
		{
			name: "from float64",
			args: args{
				val: float64(3.14),
			},
			want:    []int{3},
			wantErr: false,
		},
		{
			name: "from []float32",
			args: args{
				val: []float32{float32(3.14), float32(4.56)},
			},
			want:    []int{3, 4},
			wantErr: false,
		},
		{
			name: "from []float64",
			args: args{
				val: []float64{float64(3.14), float64(4.56)},
			},
			want:    []int{3, 4},
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				val: "23",
			},
			want:    []int{23},
			wantErr: false,
		},
		{
			name: "from string - invalid",
			args: args{
				val: "asd",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []string",
			args: args{
				val: []string{"23", "30"},
			},
			want:    []int{23, 30},
			wantErr: false,
		},
		{
			name: "from []string - invalid",
			args: args{
				val: []string{"asd", "qwe"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from bool: true",
			args: args{
				val: true,
			},
			want:    []int{1},
			wantErr: false,
		},
		{
			name: "from bool: false",
			args: args{
				val: false,
			},
			want:    []int{0},
			wantErr: false,
		},
		{
			name: "from []bool",
			args: args{
				val: []bool{true, false},
			},
			want:    []int{1, 0},
			wantErr: false,
		},
		{
			name: "from []interface",
			args: args{
				val: []interface{}{0, 1, 10},
			},
			want:    []int{0, 1, 10},
			wantErr: false,
		},
		{
			name: "from []interface - invalid",
			args: args{
				val: []interface{}{0, 1, 10, "asd'"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToIntList(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToIntList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToIntList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToInt32List(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []int32
		wantErr bool
	}{
		{
			name: "from int",
			args: args{
				val: int(1),
			},
			want:    []int32{1},
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				val: int32(1),
			},
			want:    []int32{1},
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				val: int64(1),
			},
			want:    []int32{1},
			wantErr: false,
		},
		{
			name: "from []int",
			args: args{
				val: []int{int(1), int(2)},
			},
			want:    []int32{1, 2},
			wantErr: false,
		},
		{
			name: "from []int32",
			args: args{
				val: []int32{int32(1), int32(2)},
			},
			want:    []int32{1, 2},
			wantErr: false,
		},
		{
			name: "from []int64",
			args: args{
				val: []int64{int64(1), int64(2)},
			},
			want:    []int32{1, 2},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(3.14),
			},
			want:    []int32{3},
			wantErr: false,
		},
		{
			name: "from float64",
			args: args{
				val: float64(3.14),
			},
			want:    []int32{3},
			wantErr: false,
		},
		{
			name: "from []float32",
			args: args{
				val: []float32{float32(3.14), float32(4.56)},
			},
			want:    []int32{3, 4},
			wantErr: false,
		},
		{
			name: "from []float64",
			args: args{
				val: []float64{float64(3.14), float64(4.56)},
			},
			want:    []int32{3, 4},
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				val: "23",
			},
			want:    []int32{23},
			wantErr: false,
		},
		{
			name: "from string - invalid",
			args: args{
				val: "asd",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []string",
			args: args{
				val: []string{"23", "30"},
			},
			want:    []int32{23, 30},
			wantErr: false,
		},
		{
			name: "from []string - invalid",
			args: args{
				val: []string{"23", "30", "asd"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from bool: true",
			args: args{
				val: true,
			},
			want:    []int32{1},
			wantErr: false,
		},
		{
			name: "from bool: false",
			args: args{
				val: false,
			},
			want:    []int32{0},
			wantErr: false,
		},
		{
			name: "from []bool",
			args: args{
				val: []bool{true, false},
			},
			want:    []int32{1, 0},
			wantErr: false,
		},
		{
			name: "from []interface",
			args: args{
				val: []interface{}{0, 1, 10},
			},
			want:    []int32{0, 1, 10},
			wantErr: false,
		},
		{
			name: "from []interface - invalid",
			args: args{
				val: []interface{}{0, 1, 10, "asd'"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt32List(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInt32List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToInt32List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToInt64List(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name: "from int",
			args: args{
				val: int(1),
			},
			want:    []int64{1},
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				val: int32(1),
			},
			want:    []int64{1},
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				val: int64(1),
			},
			want:    []int64{1},
			wantErr: false,
		},
		{
			name: "from []int",
			args: args{
				val: []int{int(1), int(2)},
			},
			want:    []int64{1, 2},
			wantErr: false,
		},
		{
			name: "from []int32",
			args: args{
				val: []int32{int32(1), int32(2)},
			},
			want:    []int64{1, 2},
			wantErr: false,
		},
		{
			name: "from []int64",
			args: args{
				val: []int64{int64(1), int64(2)},
			},
			want:    []int64{1, 2},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(3.14),
			},
			want:    []int64{3},
			wantErr: false,
		},
		{
			name: "from float64",
			args: args{
				val: float64(3.14),
			},
			want:    []int64{3},
			wantErr: false,
		},
		{
			name: "from []float32",
			args: args{
				val: []float32{float32(3.14), float32(4.56)},
			},
			want:    []int64{3, 4},
			wantErr: false,
		},
		{
			name: "from []float64",
			args: args{
				val: []float64{float64(3.14), float64(4.56)},
			},
			want:    []int64{3, 4},
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				val: "23",
			},
			want:    []int64{23},
			wantErr: false,
		},
		{
			name: "from string - invalid",
			args: args{
				val: "asd",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []string",
			args: args{
				val: []string{"23", "30"},
			},
			want:    []int64{23, 30},
			wantErr: false,
		},
		{
			name: "from []string - invalid",
			args: args{
				val: []string{"23", "30", "asd"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from bool: true",
			args: args{
				val: true,
			},
			want:    []int64{1},
			wantErr: false,
		},
		{
			name: "from bool: false",
			args: args{
				val: false,
			},
			want:    []int64{0},
			wantErr: false,
		},
		{
			name: "from []bool",
			args: args{
				val: []bool{true, false},
			},
			want:    []int64{1, 0},
			wantErr: false,
		},
		{
			name: "from []interface",
			args: args{
				val: []interface{}{0, 1, 10},
			},
			want:    []int64{0, 1, 10},
			wantErr: false,
		},
		{
			name: "from []interface - invalid",
			args: args{
				val: []interface{}{0, 1, 10, "asd'"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt64List(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInt64List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToInt64List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToFloat32List(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []float32
		wantErr bool
	}{
		{
			name: "from int",
			args: args{
				val: int(1),
			},
			want:    []float32{1},
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				val: int32(1),
			},
			want:    []float32{1},
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				val: int64(1),
			},
			want:    []float32{1},
			wantErr: false,
		},
		{
			name: "from []int",
			args: args{
				val: []int{int(1), int(2)},
			},
			want:    []float32{1, 2},
			wantErr: false,
		},
		{
			name: "from []int32",
			args: args{
				val: []int32{int32(1), int32(2)},
			},
			want:    []float32{1, 2},
			wantErr: false,
		},
		{
			name: "from []int64",
			args: args{
				val: []int64{int64(1), int64(2)},
			},
			want:    []float32{1, 2},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(3.14),
			},
			want:    []float32{3.14},
			wantErr: false,
		},
		{
			name: "from NaN float32",
			args: args{
				val: float32(math.NaN()),
			},
			want:    []float32{},
			wantErr: false,
		},
		{
			name: "from float64",
			args: args{
				val: float64(3.14),
			},
			want:    []float32{3.14},
			wantErr: false,
		},
		{
			name: "from NaN float64",
			args: args{
				val: math.NaN(),
			},
			want:    []float32{},
			wantErr: false,
		},
		{
			name: "from []float32",
			args: args{
				val: []float32{float32(3.14), float32(4.56), float32(math.NaN())},
			},
			want:    []float32{3.14, 4.56},
			wantErr: false,
		},
		{
			name: "from []float64",
			args: args{
				val: []float64{float64(3.14), math.NaN(), float64(4.56)},
			},
			want:    []float32{3.14, 4.56},
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				val: "23.3",
			},
			want:    []float32{23.3},
			wantErr: false,
		},
		{
			name: "from string - invalid",
			args: args{
				val: "asd",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []string",
			args: args{
				val: []string{"23.3", "30.09", "NaN"},
			},
			want:    []float32{23.3, 30.09},
			wantErr: false,
		},
		{
			name: "from []string - invalid",
			args: args{
				val: []string{"23.3", "30.09", "asd"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from bool: true",
			args: args{
				val: true,
			},
			want:    []float32{1},
			wantErr: false,
		},
		{
			name: "from bool: false",
			args: args{
				val: false,
			},
			want:    []float32{0},
			wantErr: false,
		},
		{
			name: "from []bool",
			args: args{
				val: []bool{true, false},
			},
			want:    []float32{1, 0},
			wantErr: false,
		},
		{
			name: "from []interface",
			args: args{
				val: []interface{}{0, 1, 10, math.NaN(), "NaN"},
			},
			want:    []float32{0, 1, 10},
			wantErr: false,
		},
		{
			name: "from []interface - invalid",
			args: args{
				val: []interface{}{0, 1, 10, "asd'"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat32List(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToFloat32List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToFloat32List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToFloat64List(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []float64
		wantErr bool
	}{
		{
			name: "from int",
			args: args{
				val: int(1),
			},
			want:    []float64{1},
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				val: int32(1),
			},
			want:    []float64{1},
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				val: int64(1),
			},
			want:    []float64{1},
			wantErr: false,
		},
		{
			name: "from []int",
			args: args{
				val: []int{int(1), int(2)},
			},
			want:    []float64{1, 2},
			wantErr: false,
		},
		{
			name: "from []int32",
			args: args{
				val: []int32{int32(1), int32(2)},
			},
			want:    []float64{1, 2},
			wantErr: false,
		},
		{
			name: "from []int64",
			args: args{
				val: []int64{int64(1), int64(2)},
			},
			want:    []float64{1, 2},
			wantErr: false,
		},
		{
			name: "from NaN float32",
			args: args{
				val: float32(math.NaN()),
			},
			want:    []float64{},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(3.14),
			},
			want:    []float64{3.14},
			wantErr: false,
		},
		{
			name: "from NaN float64",
			args: args{
				val: math.NaN(),
			},
			want:    []float64{},
			wantErr: false,
		},
		{
			name: "from float64",
			args: args{
				val: float64(3.14),
			},
			want:    []float64{3.14},
			wantErr: false,
		},
		{
			name: "from []float64",
			args: args{
				val: []float64{float64(3.14), float64(4.56), math.NaN()},
			},
			want:    []float64{3.14, 4.56},
			wantErr: false,
		},
		{
			name: "from []float32",
			args: args{
				val: []float32{float32(3.14), float32(4.56), float32(math.NaN())},
			},
			want:    []float64{3.14, 4.56},
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				val: "23.3",
			},
			want:    []float64{23.3},
			wantErr: false,
		},
		{
			name: "from NaN string",
			args: args{
				val: "NaN",
			},
			want:    []float64{},
			wantErr: false,
		},
		{
			name: "from string - invalid",
			args: args{
				val: "asd",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []string",
			args: args{
				val: []string{"23.3", "30.09", "NaN"},
			},
			want:    []float64{23.3, 30.09},
			wantErr: false,
		},
		{
			name: "from []string - invalid",
			args: args{
				val: []string{"23.3", "30.09", "asd"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from bool: true",
			args: args{
				val: true,
			},
			want:    []float64{1},
			wantErr: false,
		},
		{
			name: "from bool: false",
			args: args{
				val: false,
			},
			want:    []float64{0},
			wantErr: false,
		},
		{
			name: "from []bool",
			args: args{
				val: []bool{true, false},
			},
			want:    []float64{1, 0},
			wantErr: false,
		},
		{
			name: "from []interface",
			args: args{
				val: []interface{}{3.14, 1, 10, "NaN", math.NaN()},
			},
			want:    []float64{3.14, 1, 10},
			wantErr: false,
		},
		{
			name: "from []interface - invalid",
			args: args{
				val: []interface{}{3.14, 1, 10, "asd'"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat64List(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToFloat64List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// for boolean case we use assert.Equal because the array contains 0 value
			// assert.InEpsilonSlice expects all element in array to be greater than 0
			if strings.Contains(tt.name, "bool") {
				assert.Equal(t, tt.want, got)
				return
			}

			assert.InEpsilonSlice(t, tt.want, got, 0.0001)
		})
	}
}

func TestToBoolList(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []bool
		wantErr bool
	}{
		{
			name: "from int",
			args: args{
				val: int(1),
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from int",
			args: args{
				val: int(0),
			},
			want:    []bool{false},
			wantErr: false,
		},
		{
			name: "from int32",
			args: args{
				val: int32(1),
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from int64",
			args: args{
				val: int64(1),
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from int - invalid",
			args: args{
				val: int(100),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []int",
			args: args{
				val: []int{int(1), int(0)},
			},
			want:    []bool{true, false},
			wantErr: false,
		},
		{
			name: "from []int, but neither 0 or 1",
			args: args{
				val: []int{int(2), int(3)},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []int32",
			args: args{
				val: []int32{int32(1), int32(0)},
			},
			want:    []bool{true, false},
			wantErr: false,
		},
		{
			name: "from []int64",
			args: args{
				val: []int64{int64(1), int64(0)},
			},
			want:    []bool{true, false},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(1),
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from float32",
			args: args{
				val: float32(0),
			},
			want:    []bool{false},
			wantErr: false,
		},
		{
			name: "from float32 - invalid",
			args: args{
				val: float32(100),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from float64",
			args: args{
				val: float64(1),
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from []float32",
			args: args{
				val: []float32{float32(0), float32(1)},
			},
			want:    []bool{false, true},
			wantErr: false,
		},
		{
			name: "from []float64",
			args: args{
				val: []float64{float64(0), float64(1)},
			},
			want:    []bool{false, true},
			wantErr: false,
		},
		{
			name: "from []float64, neither 0 or 1",
			args: args{
				val: []float64{float64(2), float64(3)},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from string",
			args: args{
				val: "t",
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from string",
			args: args{
				val: "benar",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from []string",
			args: args{
				val: []string{"1", "t", "T", "true", "TRUE", "True"},
			},
			want:    []bool{true, true, true, true, true, true},
			wantErr: false,
		},
		{
			name: "from []string - invalid",
			args: args{
				val: []string{"1", "t", "T", "true", "TRUE", "True", "ASD"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "from bool: true",
			args: args{
				val: true,
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from bool: false",
			args: args{
				val: false,
			},
			want:    []bool{false},
			wantErr: false,
		},
		{
			name: "from []bool",
			args: args{
				val: []bool{true, false},
			},
			want:    []bool{true, false},
			wantErr: false,
		},
		{
			name: "from []interface - valid bool",
			args: args{
				val: []interface{}{true},
			},
			want:    []bool{true},
			wantErr: false,
		},
		{
			name: "from []interface - valid bool",
			args: args{
				val: []interface{}{true, "true", 1, 0, "f", "FALSE"},
			},
			want:    []bool{true, true, true, false, false, false},
			wantErr: false,
		},
		{
			name: "from []interface - invalid",
			args: args{
				val: []interface{}{"QWE", "ASD"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToBoolList(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBoolList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToBoolList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToFeastValue(t *testing.T) {
	type args struct {
		v         interface{}
		valueType types.ValueType_Enum
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Value
		wantErr bool
	}{
		{
			name: "int32",
			args: args{
				v:         int32(1),
				valueType: types.ValueType_INT32,
			},
			want:    &types.Value{Val: &types.Value_Int32Val{Int32Val: 1}},
			wantErr: false,
		},
		{
			name: "int64",
			args: args{
				v:         int64(2),
				valueType: types.ValueType_INT64,
			},
			want:    &types.Value{Val: &types.Value_Int64Val{Int64Val: 2}},
			wantErr: false,
		},
		{
			name: "float32",
			args: args{
				v:         float32(0.201),
				valueType: types.ValueType_FLOAT,
			},
			want:    &types.Value{Val: &types.Value_FloatVal{FloatVal: 0.201}},
			wantErr: false,
		},
		{
			name: "float32 from NaN string",
			args: args{
				v:         "NaN",
				valueType: types.ValueType_FLOAT,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "float32 from NaN ",
			args: args{
				v:         float32(math.NaN()),
				valueType: types.ValueType_FLOAT,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "double",
			args: args{
				v:         float64(0.201),
				valueType: types.ValueType_DOUBLE,
			},
			want:    &types.Value{Val: &types.Value_DoubleVal{DoubleVal: 0.201}},
			wantErr: false,
		},
		{
			name: "double from NaN string",
			args: args{
				v:         "NaN",
				valueType: types.ValueType_DOUBLE,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "double from NaN ",
			args: args{
				v:         math.NaN(),
				valueType: types.ValueType_DOUBLE,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "string",
			args: args{
				v:         "hello",
				valueType: types.ValueType_STRING,
			},
			want:    &types.Value{Val: &types.Value_StringVal{StringVal: "hello"}},
			wantErr: false,
		},
		{
			name: "bool",
			args: args{
				v:         true,
				valueType: types.ValueType_BOOL,
			},
			want:    &types.Value{Val: &types.Value_BoolVal{BoolVal: true}},
			wantErr: false,
		},
		{
			name: "[]int32",
			args: args{
				v:         []int32{1, 11, 111},
				valueType: types.ValueType_INT32_LIST,
			},
			want:    &types.Value{Val: &types.Value_Int32ListVal{Int32ListVal: &types.Int32List{Val: []int32{1, 11, 111}}}},
			wantErr: false,
		},
		{
			name: "[]int64",
			args: args{
				v:         []int64{2, 22, 222},
				valueType: types.ValueType_INT64_LIST,
			},
			want:    &types.Value{Val: &types.Value_Int64ListVal{Int64ListVal: &types.Int64List{Val: []int64{2, 22, 222}}}},
			wantErr: false,
		},
		{
			name: "[]float32",
			args: args{
				v:         []float32{0.201, 2.01},
				valueType: types.ValueType_FLOAT_LIST,
			},
			want:    &types.Value{Val: &types.Value_FloatListVal{FloatListVal: &types.FloatList{Val: []float32{0.201, 2.01}}}},
			wantErr: false,
		},
		{
			name: "[]double",
			args: args{
				v:         []float64{3.14, 6.28},
				valueType: types.ValueType_DOUBLE_LIST,
			},
			want:    &types.Value{Val: &types.Value_DoubleListVal{DoubleListVal: &types.DoubleList{Val: []float64{3.14, 6.28}}}},
			wantErr: false,
		},
		{
			name: "[]string",
			args: args{
				v:         []string{"hello", "world"},
				valueType: types.ValueType_STRING_LIST,
			},
			want:    &types.Value{Val: &types.Value_StringListVal{StringListVal: &types.StringList{Val: []string{"hello", "world"}}}},
			wantErr: false,
		},
		{
			name: "[]bool",
			args: args{
				v:         []bool{true, false},
				valueType: types.ValueType_BOOL_LIST,
			},
			want:    &types.Value{Val: &types.Value_BoolListVal{BoolListVal: &types.BoolList{Val: []bool{true, false}}}},
			wantErr: false,
		},
		{
			name: "[]int32 from string",
			args: args{
				v:         "[1, 11, 111]",
				valueType: types.ValueType_INT32_LIST,
			},
			want:    &types.Value{Val: &types.Value_Int32ListVal{Int32ListVal: &types.Int32List{Val: []int32{1, 11, 111}}}},
			wantErr: false,
		},
		{
			name: "[]int64 from string",
			args: args{
				v:         "[2, 22, 222]",
				valueType: types.ValueType_INT64_LIST,
			},
			want:    &types.Value{Val: &types.Value_Int64ListVal{Int64ListVal: &types.Int64List{Val: []int64{2, 22, 222}}}},
			wantErr: false,
		},
		{
			name: "[]float32 from string",
			args: args{
				v:         "[0.201, 2.01]",
				valueType: types.ValueType_FLOAT_LIST,
			},
			want:    &types.Value{Val: &types.Value_FloatListVal{FloatListVal: &types.FloatList{Val: []float32{0.201, 2.01}}}},
			wantErr: false,
		},
		{
			name: "[]double from string",
			args: args{
				v:         "[3.14, 6.28]",
				valueType: types.ValueType_DOUBLE_LIST,
			},
			want:    &types.Value{Val: &types.Value_DoubleListVal{DoubleListVal: &types.DoubleList{Val: []float64{3.14, 6.28}}}},
			wantErr: false,
		},
		{
			name: "[]string from string",
			args: args{
				v:         "[hello, world]",
				valueType: types.ValueType_STRING_LIST,
			},
			want:    &types.Value{Val: &types.Value_StringListVal{StringListVal: &types.StringList{Val: []string{"hello", "world"}}}},
			wantErr: false,
		},
		{
			name: "[]bool from string",
			args: args{
				v:         "[true, false]",
				valueType: types.ValueType_BOOL_LIST,
			},
			want:    &types.Value{Val: &types.Value_BoolListVal{BoolListVal: &types.BoolList{Val: []bool{true, false}}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFeastValue(tt.args.v, tt.args.valueType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToFeastValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToFeastValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
