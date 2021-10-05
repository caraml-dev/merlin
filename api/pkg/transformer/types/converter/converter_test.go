package converter

import (
	"encoding/base64"
	"reflect"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
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
			name: "from float",
			args: args{
				v: 1.1,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToBool(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, got, tt.want)
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
		val         *feastTypes.Value
		want        interface{}
		wantValType feastTypes.ValueType_Enum
		wantErr     bool
	}{
		{
			name:        "string",
			val:         feast.StrVal("hello"),
			want:        "hello",
			wantValType: feastTypes.ValueType_STRING,
			wantErr:     false,
		},
		{
			name:        "double",
			val:         feast.DoubleVal(123456789.123456789),
			want:        123456789.123456789,
			wantValType: feastTypes.ValueType_DOUBLE,
			wantErr:     false,
		},
		{
			name:        "float",
			val:         feast.FloatVal(1.1),
			want:        float32(1.1),
			wantValType: feastTypes.ValueType_FLOAT,
			wantErr:     false,
		},
		{
			name:        "int32",
			val:         feast.Int32Val(1234),
			want:        int32(1234),
			wantValType: feastTypes.ValueType_INT32,
			wantErr:     false,
		},
		{
			name:        "int64",
			val:         feast.Int64Val(12345678),
			want:        int64(12345678),
			wantValType: feastTypes.ValueType_INT64,
			wantErr:     false,
		},
		{
			name:        "bool",
			val:         feast.BoolVal(true),
			want:        true,
			wantValType: feastTypes.ValueType_BOOL,
			wantErr:     false,
		},
		{
			name:        "bytes",
			val:         feast.BytesVal([]byte("hello")),
			want:        base64.StdEncoding.EncodeToString([]byte("hello")),
			wantValType: feastTypes.ValueType_STRING,
			wantErr:     false,
		},
		{
			name: "string list",
			val: &feastTypes.Value{Val: &feastTypes.Value_StringListVal{StringListVal: &feastTypes.StringList{Val: []string{
				"hello",
				"world",
			}}}},
			want: []string{
				"hello",
				"world",
			},
			wantValType: feastTypes.ValueType_STRING_LIST,
			wantErr:     false,
		},
		{
			name: "double list",
			val: &feastTypes.Value{Val: &feastTypes.Value_DoubleListVal{DoubleListVal: &feastTypes.DoubleList{Val: []float64{
				123.45,
				123.45,
			}}}},
			want: []float64{
				123.45,
				123.45,
			},
			wantValType: feastTypes.ValueType_DOUBLE_LIST,
			wantErr:     false,
		},
		{
			name: "float list",
			val: &feastTypes.Value{Val: &feastTypes.Value_FloatListVal{FloatListVal: &feastTypes.FloatList{Val: []float32{
				123.45,
				123.45,
			}}}},
			want: []float32{
				123.45,
				123.45,
			},
			wantValType: feastTypes.ValueType_FLOAT_LIST,
			wantErr:     false,
		},
		{
			name: "int32 list",
			val: &feastTypes.Value{Val: &feastTypes.Value_Int32ListVal{Int32ListVal: &feastTypes.Int32List{Val: []int32{
				int32(1234),
				int32(1234),
			}}}},
			want: []int32{
				1234,
				1234,
			},
			wantValType: feastTypes.ValueType_INT32_LIST,
			wantErr:     false,
		},
		{
			name: "int64 list",
			val: &feastTypes.Value{Val: &feastTypes.Value_Int64ListVal{Int64ListVal: &feastTypes.Int64List{Val: []int64{
				int64(1234),
				int64(1234),
			}}}},
			want: []int64{
				1234,
				1234,
			},
			wantValType: feastTypes.ValueType_INT64_LIST,
			wantErr:     false,
		},
		{
			name: "bool list",
			val: &feastTypes.Value{Val: &feastTypes.Value_BoolListVal{BoolListVal: &feastTypes.BoolList{Val: []bool{
				true,
				false,
			}}}},
			want: []bool{
				true,
				false,
			},
			wantValType: feastTypes.ValueType_BOOL_LIST,
			wantErr:     false,
		},
		{
			name: "bytes list",
			val: &feastTypes.Value{Val: &feastTypes.Value_BytesListVal{BytesListVal: &feastTypes.BytesList{Val: [][]byte{
				[]byte("hello"),
				[]byte("world"),
			}}}},
			want: []string{
				base64.StdEncoding.EncodeToString([]byte("hello")),
				base64.StdEncoding.EncodeToString([]byte("world")),
			},
			wantValType: feastTypes.ValueType_STRING_LIST,
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
