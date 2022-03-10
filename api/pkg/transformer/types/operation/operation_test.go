package operation

import (
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/types/operator"
	"github.com/gojek/merlin/pkg/transformer/types/series"
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
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				Operation: operator.Add,
			},
			want: series.New([]int{3, 5, 7, 9, 11}, series.Int, ""),
		},
		{
			name: "substract two series",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				Operation: operator.Substract,
			},
			want: series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
		},
		{
			name: "multiply int with float series",
			node: &OperationNode{
				LeftVal:   2,
				RightVal:  series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				Operation: operator.Multiply,
			},
			want: series.New([]float64{4, 6, 8, 10, 12}, series.Float, ""),
		},
		{
			name: "divide float series",
			node: &OperationNode{
				LeftVal:   series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				RightVal:  series.New([]int{2}, series.Int, ""),
				Operation: operator.Divide,
			},
			want: series.New([]float64{1, 1.5, 2, 2.5, 3}, series.Float, ""),
		},
		{
			name: "modulo int series",
			node: &OperationNode{
				LeftVal:   series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				RightVal:  series.New([]int{2}, series.Int, ""),
				Operation: operator.Modulo,
			},
			want: series.New([]int{0, 1, 0, 1, 0}, series.Int, ""),
		},
		{
			name: "add int with float",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  5.2,
				Operation: operator.Add,
			},
			want: 9.2,
		},
		{
			name: "multiply int with bool",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  true,
				Operation: operator.Multiply,
			},
			wantErr: true,
		},
		{
			name: "greater int with float",
			node: &OperationNode{
				LeftVal:   3,
				RightVal:  3.5,
				Operation: operator.Greater,
			},
			want: false,
		},
		{
			name: "greater series with int",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  3,
				Operation: operator.Greater,
			},
			want: series.New([]bool{false, false, false, true, true}, series.Bool, ""),
		},
		{
			name: "equal series with series",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{1, 2, 3, 5, 5}, series.Int, ""),
				Operation: operator.Eq,
			},
			want: series.New([]bool{true, true, true, false, true}, series.Bool, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.node.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("OperationNode.Execute() error = %v, wantErr %v", err, tt.wantErr)
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
				Operation: operator.Add,
			},
			subset: series.New([]bool{true, true, false, false, false}, series.Bool, ""),
			want:   series.New([]interface{}{3, 5, nil, nil, nil}, series.Int, ""),
		},
		{
			name: "substract two series",
			node: &OperationNode{
				LeftVal:   series.New([]int{1, 2, 3, 4, 5}, series.Int, ""),
				RightVal:  series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				Operation: operator.Substract,
			},
			subset: series.New([]bool{false, false, false, false, true}, series.Bool, ""),
			want:   series.New([]interface{}{nil, nil, nil, nil, -1}, series.Int, ""),
		},
		{
			name: "multiply int with float series",
			node: &OperationNode{
				LeftVal:   2,
				RightVal:  series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				Operation: operator.Multiply,
			},
			subset: series.New([]bool{false, false, false, false, true}, series.Bool, ""),
			want:   series.New([]interface{}{nil, nil, nil, nil, 12}, series.Float, ""),
		},
		{
			name: "divide float series",
			node: &OperationNode{
				LeftVal:   series.New([]float64{2, 3, 4, 5, 6}, series.Float, ""),
				RightVal:  series.New([]int{2, 2, 2, 2, 0}, series.Int, ""),
				Operation: operator.Divide,
			},
			subset: series.New([]bool{true, true, false, true, true}, series.Bool, ""),
			want:   series.New([]interface{}{1, 1.5, nil, 2.5, nil}, series.Float, ""),
		},
		{
			name: "modulo int series",
			node: &OperationNode{
				LeftVal:   series.New([]int{2, 3, 4, 5, 6}, series.Int, ""),
				RightVal:  series.New([]int{2}, series.Int, ""),
				Operation: operator.Modulo,
			},
			subset: series.New([]bool{false, false, true, false, false}, series.Bool, ""),
			want:   series.New([]interface{}{nil, nil, 0, nil, nil}, series.Int, ""),
		},
		{
			name: "add int with float",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  5.2,
				Operation: operator.Add,
			},
			want: 9.2,
		},
		{
			name: "multiply int with bool",
			node: &OperationNode{
				LeftVal:   4,
				RightVal:  true,
				Operation: operator.Multiply,
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
