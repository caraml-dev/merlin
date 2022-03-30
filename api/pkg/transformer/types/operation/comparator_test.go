package operation

import (
	"testing"
)

func TestCompareInt64(t *testing.T) {
	type args struct {
		lValue     interface{}
		rValue     interface{}
		comparator Comparator
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
				comparator: Greater,
			},
			want: true,
		},
		{
			name: "8 > 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: Greater,
			},
			want: false,
		},
		{
			name: "8 > 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: Greater,
			},
			want: false,
		},
		{
			name: "8 >= 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: GreaterEq,
			},
			want: true,
		},
		{
			name: "8 >= 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: GreaterEq,
			},
			want: false,
		},
		{
			name: "8 >= 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: GreaterEq,
			},
			want: true,
		},
		{
			name: "8 < 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: Less,
			},
			want: false,
		},
		{
			name: "8 < 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: Less,
			},
			want: true,
		},
		{
			name: "8 < 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: Less,
			},
			want: false,
		},
		{
			name: "8 <= 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: LessEq,
			},
			want: false,
		},
		{
			name: "8 <= 9",
			args: args{
				lValue:     8,
				rValue:     9,
				comparator: LessEq,
			},
			want: true,
		},
		{
			name: "8 <= 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: LessEq,
			},
			want: true,
		},
		{
			name: "8 == 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: Eq,
			},
			want: false,
		},
		{
			name: "8 == 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: Eq,
			},
			want: true,
		},
		{
			name: "8 != 7",
			args: args{
				lValue:     8,
				rValue:     7,
				comparator: Neq,
			},
			want: true,
		},
		{
			name: "8 != 8",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: Neq,
			},
			want: false,
		},
		{
			name: "in operation is not supported",
			args: args{
				lValue:     8,
				rValue:     8,
				comparator: In,
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
		comparator Comparator
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
				comparator: Greater,
			},
			want: true,
		},
		{
			name: "8 > 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: Greater,
			},
			want: false,
		},
		{
			name: "8 > 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: Greater,
			},
			want: false,
		},
		{
			name: "8 >= 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: GreaterEq,
			},
			want: true,
		},
		{
			name: "8 >= 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: GreaterEq,
			},
			want: false,
		},
		{
			name: "8 >= 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: GreaterEq,
			},
			want: true,
		},
		{
			name: "8 < 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: Less,
			},
			want: false,
		},
		{
			name: "8 < 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: Less,
			},
			want: true,
		},
		{
			name: "8 < 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: Less,
			},
			want: false,
		},
		{
			name: "8 <= 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: LessEq,
			},
			want: false,
		},
		{
			name: "8 <= 9",
			args: args{
				lValue:     8.0,
				rValue:     9.0,
				comparator: LessEq,
			},
			want: true,
		},
		{
			name: "8 <= 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: LessEq,
			},
			want: true,
		},
		{
			name: "8 == 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: Eq,
			},
			want: false,
		},
		{
			name: "8 == 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: Eq,
			},
			want: true,
		},
		{
			name: "8 != 7",
			args: args{
				lValue:     8.0,
				rValue:     7.0,
				comparator: Neq,
			},
			want: true,
		},
		{
			name: "8 != 8",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: Neq,
			},
			want: false,
		},
		{
			name: "in operation is not supported",
			args: args{
				lValue:     8.0,
				rValue:     8.0,
				comparator: In,
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
		comparator Comparator
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
				comparator: Greater,
			},
			want: true,
		},
		{
			name: "a > ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: Greater,
			},
			want: false,
		},
		{
			name: "ab > ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: Greater,
			},
			want: false,
		},
		{
			name: "ab >= aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: GreaterEq,
			},
			want: true,
		},
		{
			name: "a >= ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: GreaterEq,
			},
			want: false,
		},
		{
			name: "ab >= ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: GreaterEq,
			},
			want: true,
		},
		{
			name: "ab < aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: Less,
			},
			want: false,
		},
		{
			name: "a < ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: Less,
			},
			want: true,
		},
		{
			name: "ab < ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: Less,
			},
			want: false,
		},
		{
			name: "ab <= aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: LessEq,
			},
			want: false,
		},
		{
			name: "a <= ab",
			args: args{
				lValue:     "a",
				rValue:     "ab",
				comparator: LessEq,
			},
			want: true,
		},
		{
			name: "ab <= ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: LessEq,
			},
			want: true,
		},
		{
			name: "ab == aa",
			args: args{
				lValue:     "ab",
				rValue:     "aa",
				comparator: Eq,
			},
			want: false,
		},
		{
			name: "ab == ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: Eq,
			},
			want: true,
		},
		{
			name: "ab != ab",
			args: args{
				lValue:     "ab",
				rValue:     "ab",
				comparator: Neq,
			},
			want: false,
		},
		{
			name: "ab != ac",
			args: args{
				lValue:     "ab",
				rValue:     "ac",
				comparator: Neq,
			},
			want: true,
		},
		{
			name: "in operation is not supported",
			args: args{
				lValue:     "ab",
				rValue:     "ac",
				comparator: In,
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
		comparator Comparator
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
				comparator: Eq,
			},
			want: true,
		},
		{
			name: "true == false",
			args: args{
				lValue:     true,
				rValue:     false,
				comparator: Eq,
			},
			want: false,
		},
		{
			name: "true != false",
			args: args{
				lValue:     true,
				rValue:     false,
				comparator: Neq,
			},
			want: true,
		},
		{
			name: "true != true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: Neq,
			},
			want: false,
		},
		{
			name: "true > true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: Greater,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "true >= true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: GreaterEq,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "true < true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: Less,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "true <= true",
			args: args{
				lValue:     true,
				rValue:     true,
				comparator: LessEq,
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
