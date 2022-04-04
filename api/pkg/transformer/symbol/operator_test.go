package symbol

import (
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/types/operation"
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
}

func createTestCases(operator operation.Operator) []testCases {
	sr := NewRegistry()
	return []testCases{
		{
			name: "left and right is int",
			args: args{
				left:  3,
				right: 4,
			},
			sr: sr,
			want: &operation.OperationNode{
				LeftVal:   3,
				RightVal:  4,
				Operation: operator,
				Next:      nil,
			},
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
}

func TestRegistry_GreaterOp(t *testing.T) {
	tests := createTestCases(operation.Greater)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.GreaterOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GreaterOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_GreaterEqOp(t *testing.T) {
	tests := createTestCases(operation.GreaterEq)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.GreaterEqOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GreaterEqOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_LessOp(t *testing.T) {
	tests := createTestCases(operation.Less)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.LessOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.LessOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_LessEqOp(t *testing.T) {
	tests := createTestCases(operation.LessEq)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.LessEqOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.LessEqOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_EqualOp(t *testing.T) {
	tests := createTestCases(operation.Eq)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.EqualOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.EqualOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_NeqOp(t *testing.T) {
	tests := createTestCases(operation.Neq)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.NeqOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.NeqOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_AndOp(t *testing.T) {
	tests := createTestCases(operation.And)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.AndOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.AndOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_OrOp(t *testing.T) {
	tests := createTestCases(operation.Or)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.OrOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.OrOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_AddOp(t *testing.T) {
	tests := createTestCases(operation.Add)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.AddOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.AddOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_SubstractOp(t *testing.T) {
	tests := createTestCases(operation.Substract)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.SubstractOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.SubstractOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_MultiplyOp(t *testing.T) {
	tests := createTestCases(operation.Multiply)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.MultiplyOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.MultiplyOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_DivideOp(t *testing.T) {
	tests := createTestCases(operation.Divide)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.DivideOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.DivideOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_ModuloOp(t *testing.T) {
	tests := createTestCases(operation.Modulo)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.ModuloOp(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.ModuloOp() = %v, want %v", got, tt.want)
			}
		})
	}
}
