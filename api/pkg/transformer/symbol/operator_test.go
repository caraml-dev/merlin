package symbol

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/caraml-dev/merlin/pkg/transformer/types/operation"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	"github.com/stretchr/testify/assert"
)

type args struct {
	left  interface{}
	right interface{}
}

type testCases struct {
	name string
	sr   Registry
	args args
	want interface{}
	err  error
}

func TestRegistry_GreaterOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Greater
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 4,
			},
			sr:   sr,
			want: false,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "abcd",
				right: "aaaa",
			},
			sr:   sr,
			want: true,
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.GreaterOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GreaterOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_GreaterEqOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.GreaterEq
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 3,
			},
			sr:   sr,
			want: true,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "aaaa",
				right: "abcd",
			},
			sr:   sr,
			want: false,
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.GreaterEqOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GreaterEqOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_LessOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Less
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 4,
			},
			sr:   sr,
			want: true,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "abcd",
				right: "aaaa",
			},
			sr:   sr,
			want: false,
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.LessOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.LessOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_LessEqOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.LessEq
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 3,
			},
			sr:   sr,
			want: true,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "abcd",
				right: "aaaa",
			},
			sr:   sr,
			want: false,
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.LessEqOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.LessEqOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_EqualOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Eq
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 3,
			},
			sr:   sr,
			want: true,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "abcd",
				right: "aaaa",
			},
			sr:   sr,
			want: false,
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.EqualOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.EqualOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_NeqOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Neq
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 4,
			},
			sr:   sr,
			want: true,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "abcd",
				right: "abcd",
			},
			sr:   sr,
			want: false,
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.NeqOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.NeqOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_AndOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.And
	tests := []testCases{
		{
			name: "left and right is boolean",
			args: args{
				left:  true,
				right: false,
			},
			sr:   sr,
			want: false,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "abcd",
				right: "aaaa",
			},
			sr:  sr,
			err: fmt.Errorf("logical operation && is not supported for type string and string"),
		},

		{
			name: "left is pointer series and right is boolean",
			args: args{
				left:  series.New([]bool{true, true, true}, series.Bool, ""),
				right: true,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]bool{true, true, true}, series.Bool, ""),
				RightVal:  true,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   true,
					RightVal:  false,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   true,
					RightVal:  false,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   true,
				RightVal:  false,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   true,
						RightVal:  false,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithError(t, tt.err.Error(), func() {
					tt.sr.AndOp(tt.args.left, tt.args.right)
				})
				return
			}

			if got := tt.sr.AndOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.AndOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_OrOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Or
	tests := []testCases{
		{
			name: "left and right is boolean",
			args: args{
				left:  true,
				right: false,
			},
			sr:   sr,
			want: true,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "abcd",
				right: "aaaa",
			},
			sr:  sr,
			err: fmt.Errorf("logical operation || is not supported for type string and string"),
		},

		{
			name: "left is pointer series and right is boolean",
			args: args{
				left:  series.New([]bool{true, true, true}, series.Bool, ""),
				right: true,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]bool{true, true, true}, series.Bool, ""),
				RightVal:  true,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   true,
					RightVal:  false,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   true,
					RightVal:  false,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   true,
				RightVal:  false,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   true,
						RightVal:  false,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithError(t, tt.err.Error(), func() {
					tt.sr.OrOp(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.OrOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.OrOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_AddOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Add
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 3,
			},
			sr:   sr,
			want: int64(6),
		},
		{
			name: "left and right is string",
			args: args{
				left:  "aaaa",
				right: "abcd",
			},
			sr:   sr,
			want: "aaaaabcd",
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.AddOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.AddOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_SubstractOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Substract
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 3,
			},
			sr:   sr,
			want: int64(0),
		},
		{
			name: "left is float and right is int",
			args: args{
				left:  3.5,
				right: 3,
			},
			sr:   sr,
			want: 0.5,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "aaaa",
				right: "abcd",
			},
			sr:  sr,
			err: fmt.Errorf("- operation is not supported"),
		},
		{
			name: "left is int and right is bool",
			args: args{
				left:  2,
				right: true,
			},
			sr:  sr,
			err: fmt.Errorf("comparison not supported between int and bool"),
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithError(t, tt.err.Error(), func() {
					tt.sr.SubstractOp(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.SubstractOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.SubstractOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_MultiplyOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Multiply
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 3,
			},
			sr:   sr,
			want: int64(9),
		},
		{
			name: "left is float and right is int",
			args: args{
				left:  3.5,
				right: 3,
			},
			sr:   sr,
			want: 10.5,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "aaaa",
				right: "abcd",
			},
			sr:  sr,
			err: fmt.Errorf("* operation is not supported"),
		},
		{
			name: "left is int and right is bool",
			args: args{
				left:  2,
				right: true,
			},
			sr:  sr,
			err: fmt.Errorf("comparison not supported between int and bool"),
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithError(t, tt.err.Error(), func() {
					tt.sr.MultiplyOp(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.MultiplyOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.MultiplyOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_DivideOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Divide
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 3,
			},
			sr:   sr,
			want: int64(1),
		},
		{
			name: "left is float and right is int",
			args: args{
				left:  3.0,
				right: 2,
			},
			sr:   sr,
			want: 1.5,
		},
		{
			name: "left and right is string",
			args: args{
				left:  "aaaa",
				right: "abcd",
			},
			sr:  sr,
			err: fmt.Errorf("/ operation is not supported"),
		},
		{
			name: "left is int and right is bool",
			args: args{
				left:  2,
				right: true,
			},
			sr:  sr,
			err: fmt.Errorf("comparison not supported between int and bool"),
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithError(t, tt.err.Error(), func() {
					tt.sr.DivideOp(tt.args.left, tt.args.right)
				})
				return
			}
			if got := tt.sr.DivideOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.DivideOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_ModuloOp(t *testing.T) {
	sr := NewRegistry()
	operator := operation.Modulo
	tests := []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  14,
				right: 3,
			},
			sr:   sr,
			want: int64(2),
		},
		{
			name: "left is float and right is int",
			args: args{
				left:  3.0,
				right: 2,
			},
			sr:  sr,
			err: fmt.Errorf(`%% operation is not supported`),
		},
		{
			name: "left and right is string",
			args: args{
				left:  "aaaa",
				right: "abcd",
			},
			sr:  sr,
			err: fmt.Errorf("%% operation is not supported"),
		},
		{
			name: "left is int and right is bool",
			args: args{
				left:  2,
				right: true,
			},
			sr:  sr,
			err: fmt.Errorf("comparison not supported between int and bool"),
		},
		{
			name: "left is pointer operationNode and right is float",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
		{
			name: "left is pointer series and right is float",
			args: args{
				left:  series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   series.New([]float64{1.1, 2.2, 3.3}, series.Float, ""),
				RightVal:  4.2,
				Operation: operator,
			},
		},
		{
			name: "left is pointer operationNode and right is operationNode",
			args: args{
				left: &operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: &operation.OperationNode{
					LeftVal:   2,
					RightVal:  5,
					Operation: operator,
				},
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					LeftVal: nil,
					RightVal: &operation.OperationNode{
						LeftVal:   2,
						RightVal:  5,
						Operation: operator,
					},
					Operation: operator,
				},
			},
		},
		{
			name: "left is  operationNode and right is float",
			args: args{
				left: operation.OperationNode{
					LeftVal:   3,
					RightVal:  4,
					Operation: operator,
				},
				right: 4.2,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next: &operation.OperationNode{
					RightVal:  4.2,
					Operation: operator,
				},
			},
		},
	}
	for _, tt := range tests {
		if tt.err != nil {
			assert.PanicsWithError(t, tt.err.Error(), func() {
				tt.sr.ModuloOp(tt.args.left, tt.args.right)
			})
			return
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.ModuloOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.ModuloOp() = %v, want %v", got, tt.want)
			}
		})
	}
}
