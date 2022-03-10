package operation

import (
	"testing"

	"github.com/gojek/merlin/pkg/transformer/types/operator"
)

func TestCompareInt64(t *testing.T) {
	type args struct {
		lValue     interface{}
		rValue     interface{}
		comparator operator.Comparator
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "8 > 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: operator.Greater,
			},
			want: true,
		},
		{
			name: "8 > 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: operator.Greater,
			},
			want: false,
		},
		{
			name: "8 > 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: operator.Greater,
			},
			want: false,
		},
		{
			name: "8 >= 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: operator.GreaterEq,
			},
			want: true,
		},
		{
			name: "8 >= 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: operator.GreaterEq,
			},
			want: false,
		},
		{
			name: "8 >= 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: operator.GreaterEq,
			},
			want: true,
		},
		{
			name: "8 < 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: operator.Less,
			},
			want: false,
		},
		{
			name: "8 < 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: operator.Less,
			},
			want: true,
		},
		{
			name: "8 < 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: operator.Less,
			},
			want: false,
		},
		{
			name: "8 <= 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: operator.LessEq,
			},
			want: false,
		},
		{
			name: "8 <= 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: operator.LessEq,
			},
			want: true,
		},
		{
			name: "8 <= 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: operator.LessEq,
			},
			want: true,
		},
		{
			name: "8 == 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: operator.Eq,
			},
			want: false,
		},
		{
			name: "8 == 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: operator.Eq,
			},
			want: true,
		},
		{
			name: "8 != 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: operator.Neq,
			},
			want: true,
		},
		{
			name: "8 != 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: operator.Neq,
			},
			want: false,
		},
		{
			name: "in operation is not supported",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: operator.In,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareInt64(tt.args.lValue, tt.args.rValue, tt.args.comparator)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareFloat64(t *testing.T) {
	type args struct {
		lValue     interface{}
		rValue     interface{}
		comparator operator.Comparator
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "8 > 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: operator.Greater,
			},
			want: true,
		},
		{
			name: "8 > 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: operator.Greater,
			},
			want: false,
		},
		{
			name: "8 > 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: operator.Greater,
			},
			want: false,
		},
		{
			name: "8 >= 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: operator.GreaterEq,
			},
			want: true,
		},
		{
			name: "8 >= 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: operator.GreaterEq,
			},
			want: false,
		},
		{
			name: "8 >= 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: operator.GreaterEq,
			},
			want: true,
		},
		{
			name: "8 < 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: operator.Less,
			},
			want: false,
		},
		{
			name: "8 < 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: operator.Less,
			},
			want: true,
		},
		{
			name: "8 < 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: operator.Less,
			},
			want: false,
		},
		{
			name: "8 <= 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: operator.LessEq,
			},
			want: false,
		},
		{
			name: "8 <= 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: operator.LessEq,
			},
			want: true,
		},
		{
			name: "8 <= 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: operator.LessEq,
			},
			want: true,
		},
		{
			name: "8 == 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: operator.Eq,
			},
			want: false,
		},
		{
			name: "8 == 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: operator.Eq,
			},
			want: true,
		},
		{
			name: "8 != 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: operator.Neq,
			},
			want: true,
		},
		{
			name: "8 != 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: operator.Neq,
			},
			want: false,
		},
		{
			name: "in operation is not supported",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: operator.In,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareFloat64(tt.args.lValue, tt.args.rValue, tt.args.comparator)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareString(t *testing.T) {
	type args struct {
		lValue     interface{}
		rValue     interface{}
		comparator operator.Comparator
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "ab > aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: operator.Greater,
			},
			want: true,
		},
		{
			name: "a > ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: operator.Greater,
			},
			want: false,
		},
		{
			name: "ab > ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: operator.Greater,
			},
			want: false,
		},
		{
			name: "ab >= aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: operator.GreaterEq,
			},
			want: true,
		},
		{
			name: "a >= ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: operator.GreaterEq,
			},
			want: false,
		},
		{
			name: "ab >= ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: operator.GreaterEq,
			},
			want: true,
		},
		{
			name: "ab < aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: operator.Less,
			},
			want: false,
		},
		{
			name: "a < ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: operator.Less,
			},
			want: true,
		},
		{
			name: "ab < ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: operator.Less,
			},
			want: false,
		},
		{
			name: "ab <= aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: operator.LessEq,
			},
			want: false,
		},
		{
			name: "a <= ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: operator.LessEq,
			},
			want: true,
		},
		{
			name: "ab <= ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: operator.LessEq,
			},
			want: true,
		},
		{
			name: "ab == aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: operator.Eq,
			},
			want: false,
		},
		{
			name: "ab == ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: operator.Eq,
			},
			want: true,
		},
		{
			name: "ab != ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: operator.Neq,
			},
			want: false,
		},
		{
			name: "ab != ac",
			args: args{
				lValue:     "ab",
				rValue:     "ac",
				comparator: operator.Neq,
			},
			want: true,
		},
		{
			name: "in operation is not supported",
			args: args{
				lValue:     "ab",
				rValue:     "ac",
				comparator: operator.In,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareString(tt.args.lValue, tt.args.rValue, tt.args.comparator)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareBool(t *testing.T) {
	type args struct {
		lValue     interface{}
		rValue     interface{}
		comparator operator.Comparator
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "true == true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: operator.Eq,
			},
			want: true,
		},
		{
			name: "true == false",
			args: args{
				lValue:     true,
				rValue:     false,
				comparator: operator.Eq,
			},
			want: false,
		},
		{
			name: "true != false",
			args: args{
				lValue:     true,
				rValue:     false,
				comparator: operator.Neq,
			},
			want: true,
		},
		{
			name: "true != true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: operator.Neq,
			},
			want: false,
		},
		{
			name: "true > true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: operator.Greater,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "true >= true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: operator.GreaterEq,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "true < true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: operator.Less,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "true <= true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: operator.LessEq,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareBool(tt.args.lValue, tt.args.rValue, tt.args.comparator)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareBool() = %v, want %v", got, tt.want)
			}
		})
	}
}
