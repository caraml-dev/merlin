package operation

import (
	"math"
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/stretchr/testify/assert"
)

func TestOperationNode_Execute(t *testing.T) {
	tests := []struct {
		name    string
		node    *OperationNode
		want    interface{}
		wantErr bool
	}{
		{
			name: "add two series",
			node: &OperationNode{
				LeftVal:   series.New([]interface{}{1, 2, 3, 4, 5, nil}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6, 7}, series.Int, ""),
				Operation: Add,
			},
			want: series.New([]interface{}{3, 5, 7, 9, 11, nil}, series.Int, ""),
		},
		{
			name: "add int32 with int64",
			node: &OperationNode{
				LeftVal:   int32(5),
				RightVal:  int64(6),
				Operation: Add,
			},
			want: int64(11),
		},
		{
			name: "divide int",
			node: &OperationNode{
				LeftVal:   10,
				RightVal:  2,
				Operation: Divide,
			},
			want: int64(5),
		},
		{
			name: "multiply int",
			node: &OperationNode{
				LeftVal:   10,
				RightVal:  2,
				Operation: Multiply,
			},
			want: int64(20),
		},
		{
			name: "modulo int",
			node: &OperationNode{
				LeftVal:   10,
				RightVal:  3,
				Operation: Modulo,
			},
			want: int64(1),
		},
		{
			name: "add int with string",
			node: &OperationNode{
				LeftVal:   10,
				RightVal:  "abc",
				Operation: Add,
			},
			wantErr: true,
		},
		{
			name: "substract int64",
			node: &OperationNode{
				LeftVal:   int64(4),
				RightVal:  int64(2),
				Operation: Substract,
			},
			want: int64(2),
		},
		{
			name: "add float32 with int64",
			node: &OperationNode{
				LeftVal:   float32(5.0),
				RightVal:  int64(6),
				Operation: Add,
			},
			want: float64(11.0),
		},
		{
			name: "divide float64",
			node: &OperationNode{
				LeftVal:   float64(6),
				RightVal:  float64(1.5),
				Operation: Divide,
			},
			want: float64(4.0),
		},
		{
			name: "divide 0 float64",
			node: &OperationNode{
				LeftVal:   float64(6),
				RightVal:  0,
				Operation: Divide,
			},
			want: math.NaN(),
		},
		{
			name: "multiply float64",
			node: &OperationNode{
				LeftVal:   float64(2.0),
				RightVal:  float64(3.0),
				Operation: Multiply,
			},
			want: float64(6.0),
		},
		{
			name: "modulo float",
			node: &OperationNode{
				LeftVal:   10.1,
				RightVal:  3.2,
				Operation: Modulo,
			},
			wantErr: true,
		},
		{
			name: "substract float",
			node: &OperationNode{
				LeftVal:   float64(10.25),
				RightVal:  float64(10.0),
				Operation: Substract,
			},
			want: float64(0.25),
		},
		{
			name: "add two string",
			node: &OperationNode{
				LeftVal:   "abcd",
				RightVal:  "efgh",
				Operation: Add,
			},
			want: "abcdefgh",
		},
		{
			name: "add string with int - error",
			node: &OperationNode{
				LeftVal:   "abcd",
				RightVal:  5,
				Operation: Add,
			},
			wantErr: true,
		},
		{
			name: "add bool with int - error",
			node: &OperationNode{
				LeftVal:   true,
				RightVal:  5,
				Operation: Add,
			},
			wantErr: true,
		},
		{
			name: "add string with series of string",
			node: &OperationNode{
				LeftVal:   "prefix_",
				RightVal:  series.New([]string{"a", "b", "c", "d"}, series.String, ""),
				Operation: Add,
			},
			want: series.New([]string{"prefix_a", "prefix_b", "prefix_c", "prefix_d"}, series.String, ""),
		},
		{
			name: "add multiple series",
			node: &OperationNode{
				LeftVal:   series.New([]interface{}{1, 2, 3, 4, 5, nil}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6, 7}, series.Int, ""),
				Operation: Add,
				Next: &OperationNode{
					RightVal: &OperationNode{
						LeftVal:   series.New([]interface{}{1, 1, 1, 1, 1, nil}, series.Int, ""),
						RightVal:  series.New([]int{1, 1, 1, 1, 1, 1}, series.Int, ""),
						Operation: Add,
					},
					Operation: Add,
				},
			},
			want: series.New([]interface{}{5, 7, 9, 11, 13, nil}, series.Int, ""),
		},
		{
			name: "substract two series",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				Operation: Substract,
			},
			want: series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
		},
		{
			name: "multiply int with float series",
			node: &OperationNode{
				LeftVal:   2,
				RightVal:  series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				Operation: Multiply,
			},
			want: series.New([]float64{4, 6, 8, 10, 12}, series.Float, ""),
		},
		{
			name: "add float64 with float series",
			node: &OperationNode{
				LeftVal:   float64(2.0),
				RightVal:  series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				Operation: Add,
			},
			want: series.New([]float64{4, 5, 6, 7, 8}, series.Float, ""),
		},
		{
			name: "add float64 with bool",
			node: &OperationNode{
				LeftVal:   float64(2.0),
				RightVal:  true,
				Operation: Add,
			},
			wantErr: true,
		},
		{
			name: "divide float series",
			node: &OperationNode{
				LeftVal:   series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				RightVal:  series.New([]int{2}, series.Int, ""),
				Operation: Divide,
			},
			want: series.New([]float64{1, 1.5, 2, 2.5, 3}, series.Float, ""),
		},
		{
			name: "modulo int series",
			node: &OperationNode{
				LeftVal:   series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				RightVal:  series.New([]int{2}, series.Int, ""),
				Operation: Modulo,
			},
			want: series.New([]int{0, 1, 0, 1, 0}, series.Int, ""),
		},
		{
			name: "add int with float",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  5.2,
				Operation: Add,
			},
			want: 9.2,
		},
		{
			name: "multiply int with bool",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  true,
				Operation: Multiply,
			},
			wantErr: true,
		},
		{
			name: "greater int with float",
			node: &OperationNode{
				LeftVal:   3,
				RightVal:  3.5,
				Operation: Greater,
			},
			want: false,
		},
		{
			name: "and operator for bool",
			node: &OperationNode{
				LeftVal:   true,
				RightVal:  false,
				Operation: And,
			},
			want: false,
		},
		{
			name: "and operator for bool and int",
			node: &OperationNode{
				LeftVal:   true,
				RightVal:  3,
				Operation: And,
			},
			wantErr: true,
		},
		{
			name: "or operator for string and bool",
			node: &OperationNode{
				LeftVal:   3,
				RightVal:  true,
				Operation: Or,
			},
			wantErr: true,
		},
		{
			name: "or operator for bool",
			node: &OperationNode{
				LeftVal:   true,
				RightVal:  false,
				Operation: Or,
			},
			want: true,
		},
		{
			name: "and operation between bool and series of bool",
			node: &OperationNode{
				LeftVal:   true,
				RightVal:  series.New([]bool{false, true, false, true}, series.Bool, ""),
				Operation: And,
			},
			want: series.New([]bool{false, true, false, true}, series.Bool, ""),
		},
		{
			name: "or operation between bool and series of bool",
			node: &OperationNode{
				LeftVal:   true,
				RightVal:  series.New([]bool{false, true, false, true}, series.Bool, ""),
				Operation: Or,
			},
			want: series.New([]bool{true, true, true, true}, series.Bool, ""),
		},
		{
			name: "and operation between series of bool",
			node: &OperationNode{
				LeftVal:   series.New([]bool{false, true, false, true}, series.Bool, ""),
				RightVal:  series.New([]bool{false, true, false, true}, series.Bool, ""),
				Operation: And,
			},
			want: series.New([]bool{false, true, false, true}, series.Bool, ""),
		},
		{
			name: "or operation between series of bool",
			node: &OperationNode{
				LeftVal:   series.New([]bool{false, true, false, true}, series.Bool, ""),
				RightVal:  series.New([]bool{false, true, false, true}, series.Bool, ""),
				Operation: Or,
			},
			want: series.New([]bool{false, true, false, true}, series.Bool, ""),
		},
		{
			name: "or operation between series of bool different dimension",
			node: &OperationNode{
				LeftVal:   series.New([]bool{false, true, false, true, true}, series.Bool, ""),
				RightVal:  series.New([]bool{false, true, false, true}, series.Bool, ""),
				Operation: Or,
			},
			wantErr: true,
		},
		{
			name: "greater: int with float64",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  float64(3.8),
				Operation: Greater,
			},
			want: true,
		},
		{
			name: "greater equal: int64",
			node: &OperationNode{
				LeftVal:   int64(2),
				RightVal:  int64(2),
				Operation: GreaterEq,
			},
			want: true,
		},
		{
			name: "greater equal: int64 with series",
			node: &OperationNode{
				LeftVal:   int64(2),
				RightVal:  series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				Operation: GreaterEq,
			},
			want: series.New([]bool{true, true, false, false, false}, series.Bool, ""),
		},
		{
			name: "greater: series with int",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  3,
				Operation: Greater,
			},
			want: series.New([]bool{false, false, false, true, true}, series.Bool, ""),
		},

		{
			name: "less operation: int with float64",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  float64(3.8),
				Operation: Less,
			},
			want: false,
		},
		{
			name: "less operation: float64",
			node: &OperationNode{
				LeftVal:   float64(2),
				RightVal:  float64(2),
				Operation: Less,
			},
			want: false,
		},
		{
			name: "less operation: int64 with series",
			node: &OperationNode{
				LeftVal:   int64(2),
				RightVal:  series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				Operation: Less,
			},
			want: series.New([]bool{false, false, true, true, true}, series.Bool, ""),
		},
		{
			name: "less operation: series with int",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  3,
				Operation: Less,
			},
			want: series.New([]bool{true, true, false, false, false}, series.Bool, ""),
		},
		{
			name: "lesseq operation: int with float64",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  float64(3.8),
				Operation: LessEq,
			},
			want: false,
		},
		{
			name: "lesseq operation: int64",
			node: &OperationNode{
				LeftVal:   int64(2),
				RightVal:  int64(2),
				Operation: LessEq,
			},
			want: true,
		},
		{
			name: "lesseq operation: int64 with series",
			node: &OperationNode{
				LeftVal:   int64(2),
				RightVal:  series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				Operation: LessEq,
			},
			want: series.New([]bool{false, true, true, true, true}, series.Bool, ""),
		},
		{
			name: "lesseq operation: series with int",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  3,
				Operation: LessEq,
			},
			want: series.New([]bool{true, true, true, false, false}, series.Bool, ""),
		},
		{
			name: "lesseq operation: float with series",
			node: &OperationNode{
				LeftVal:   float64(3.2),
				RightVal:  series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				Operation: LessEq,
			},
			want: series.New([]bool{false, false, false, true, true}, series.Bool, ""),
		},
		{
			name: "equal series with series",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{1, 2, 3, 5, 5}, series.Int, ""),
				Operation: Eq,
			},
			want: series.New([]bool{true, true, true, false, true}, series.Bool, ""),
		},
		{
			name: "equal list series with list series",
			node: &OperationNode{
				LeftVal:   series.New([][]interface{}{{1, 2, 4}, nil}, series.IntList, ""),
				RightVal:  series.New([][]interface{}{{1, 2, 4}, nil}, series.IntList, ""),
				Operation: Eq,
			},
			want: series.New([]bool{true, true}, series.Bool, ""),
		},
		{
			name: "not equal series with series",
			node: &OperationNode{
				LeftVal:   series.New([]interface{}{1, 2, 3, 4, nil}, series.Int, ""),
				RightVal:  nil,
				Operation: Neq,
			},
			want: series.New([]bool{true, true, true, true, false}, series.Bool, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.node.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("OperationNode.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if want, validType := tt.want.(float64); validType && math.IsNaN(want) {
				gotFloat64, ok := got.(float64)
				assert.True(t, ok)
				if !math.IsNaN(gotFloat64) {
					t.Errorf("OperationNode.Execute() = %v, want %v", got, tt.want)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OperationNode.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperationNode_ExecuteSubset(t *testing.T) {
	tests := []struct {
		name    string
		node    *OperationNode
		subset  *series.Series
		want    interface{}
		wantErr bool
	}{
		{
			name: "add two series",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				Operation: Add,
			},
			subset: series.New([]bool{true, true, false, false, false}, series.Bool, ""),
			want:   series.New([]interface{}{3, 5, nil, nil, nil}, series.Int, ""),
		},
		{
			name: "substract two series",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				Operation: Substract,
			},
			subset: series.New([]bool{false, false, false, false, true}, series.Bool, ""),
			want:   series.New([]interface{}{nil, nil, nil, nil, -1}, series.Int, ""),
		},
		{
			name: "multiply int with float series",
			node: &OperationNode{
				LeftVal:   2,
				RightVal:  series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				Operation: Multiply,
			},
			subset: series.New([]bool{false, false, false, false, true}, series.Bool, ""),
			want:   series.New([]interface{}{nil, nil, nil, nil, 12}, series.Float, ""),
		},
		{
			name: "divide float series",
			node: &OperationNode{
				LeftVal:   series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				RightVal:  series.New([]int{2, 2, 2, 2, 0}, series.Int, ""),
				Operation: Divide,
			},
			subset: series.New([]bool{true, true, false, true, true}, series.Bool, ""),
			want:   series.New([]interface{}{1, 1.5, nil, 2.5, nil}, series.Float, ""),
		},
		{
			name: "modulo int series",
			node: &OperationNode{
				LeftVal:   series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				RightVal:  series.New([]int{2}, series.Int, ""),
				Operation: Modulo,
			},
			subset: series.New([]bool{false, false, true, false, false}, series.Bool, ""),
			want:   series.New([]interface{}{nil, nil, 0, nil, nil}, series.Int, ""),
		},
		{
			name: "add int with float",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  5.2,
				Operation: Add,
			},
			want: 9.2,
		},
		{
			name: "multiply int with bool",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  true,
				Operation: Multiply,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.node.ExecuteSubset(tt.subset)
			if (err != nil) != tt.wantErr {
				t.Errorf("OperationNode.ExecuteSubset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OperationNode.ExecuteSubset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterOperation(t *testing.T) {
	type args struct {
		leftVal  interface{}
		rightVal interface{}
		operator Operator
	}
	tests := []struct {
		name string
		args args
		want *OperationNode
	}{
		{
			name: "simple left and right val",
			args: args{
				leftVal:  2,
				rightVal: 3,
				operator: Add,
			},
			want: &OperationNode{
				LeftVal:   2,
				RightVal:  3,
				Operation: Add,
			},
		},
		{
			name: "left value is operation node",
			args: args{
				leftVal: OperationNode{
					LeftVal:   1,
					RightVal:  3,
					Operation: Multiply,
				},
				rightVal: &OperationNode{
					LeftVal:   2,
					RightVal:  4,
					Operation: Add,
				},
				operator: Substract,
			},
			want: &OperationNode{
				LeftVal:   1,
				RightVal:  3,
				Operation: Multiply,
				Next: &OperationNode{
					LeftVal: nil,
					RightVal: &OperationNode{
						LeftVal:   2,
						RightVal:  4,
						Operation: Add,
					},
					Operation: Substract,
				},
			},
		},
		{
			name: "left value is pointer operation node",
			args: args{
				leftVal: &OperationNode{
					LeftVal:   1,
					RightVal:  3,
					Operation: Multiply,
				},
				rightVal: &OperationNode{
					LeftVal:   2,
					RightVal:  4,
					Operation: Add,
				},
				operator: Substract,
			},
			want: &OperationNode{
				LeftVal:   1,
				RightVal:  3,
				Operation: Multiply,
				Next: &OperationNode{
					LeftVal: nil,
					RightVal: &OperationNode{
						LeftVal:   2,
						RightVal:  4,
						Operation: Add,
					},
					Operation: Substract,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RegisterOperation(tt.args.leftVal, tt.args.rightVal, tt.args.operator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegisterOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}
