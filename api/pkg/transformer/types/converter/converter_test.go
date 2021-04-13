package converter

import (
	"testing"

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
