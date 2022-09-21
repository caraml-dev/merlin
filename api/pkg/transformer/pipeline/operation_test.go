package pipeline

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/stretchr/testify/assert"
)

func TestOperationTracing_AddInputOutput(t *testing.T) {
	type args struct {
		input  map[string]interface{}
		output map[string]interface{}
	}
	tests := []struct {
		name             string
		operationTracing *OperationTracing
		args             args
		expectedInput    []map[string]interface{}
		expectedOutput   []map[string]interface{}
	}{
		{
			name: "add output only",
			operationTracing: &OperationTracing{
				Specs: &spec.Variable{
					Name: "var1",
					Value: &spec.Variable_Literal{
						Literal: &spec.Literal{
							LiteralValue: &spec.Literal_IntValue{
								IntValue: 4,
							},
						},
					},
				},
				OpType: types.VariableOpType,
			},
			args: args{
				input: nil,
				output: map[string]interface{}{
					"var1": 4,
				},
			},
			expectedInput: []map[string]interface{}{
				nil,
			},
			expectedOutput: []map[string]interface{}{
				{
					"var1": 4,
				},
			},
		},
		{
			name: "add input and output only",
			operationTracing: &OperationTracing{
				Specs: &spec.TableJoin{
					LeftTable:   "leftTable",
					RightTable:  "rightTable",
					OutputTable: "outputTable",
					How:         spec.JoinMethod_INNER,
					OnColumn:    "col1",
				},
				OpType: types.VariableOpType,
			},
			args: args{
				input: map[string]interface{}{
					"leftTable": []map[string]interface{}{
						{
							"col1": 1,
							"col2": 2,
							"col3": 3,
						},
						{
							"col1": 2,
							"col2": 4,
							"col3": 6,
						},
					},
					"rightTable": []map[string]interface{}{
						{
							"col1": 1,
							"col4": 2,
							"col5": 3,
						},
						{
							"col1": 2,
							"col4": 4,
							"col5": 6,
						},
					},
				},
				output: map[string]interface{}{
					"outputTable": []map[string]interface{}{
						{
							"col1": 1,
							"col2": 2,
							"col3": 3,
							"col4": 2,
							"col5": 3,
						},
						{
							"col1": 2,
							"col2": 4,
							"col3": 6,
							"col4": 4,
							"col5": 6,
						},
					},
				},
			},
			expectedInput: []map[string]interface{}{
				{
					"leftTable": []map[string]interface{}{
						{
							"col1": 1,
							"col2": 2,
							"col3": 3,
						},
						{
							"col1": 2,
							"col2": 4,
							"col3": 6,
						},
					},
					"rightTable": []map[string]interface{}{
						{
							"col1": 1,
							"col4": 2,
							"col5": 3,
						},
						{
							"col1": 2,
							"col4": 4,
							"col5": 6,
						},
					},
				},
			},
			expectedOutput: []map[string]interface{}{
				{
					"outputTable": []map[string]interface{}{
						{
							"col1": 1,
							"col2": 2,
							"col3": 3,
							"col4": 2,
							"col5": 3,
						},
						{
							"col1": 2,
							"col2": 4,
							"col3": 6,
							"col4": 4,
							"col5": 6,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := tt.operationTracing
			err := ot.AddInputOutput(tt.args.input, tt.args.output)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedInput, ot.Input)
			assert.Equal(t, tt.expectedOutput, ot.Output)
		})
	}
}

func TestOperationTracing_GetOperationTracingDetail(t *testing.T) {
	tests := []struct {
		name             string
		operationTracing *OperationTracing
		want             []types.TracingDetail
		wantErr          error
	}{
		{
			name: "list of variable specs",
			operationTracing: &OperationTracing{
				Specs: []*spec.Variable{
					{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 4,
								},
							},
						},
					},
					{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 5,
								},
							},
						},
					},
				},
				Input: []map[string]interface{}{
					nil, nil,
				},
				Output: []map[string]interface{}{
					{
						"var1": 4,
					},
					{
						"var1": 5,
					},
				},
				OpType: types.VariableOpType,
			},
			want: []types.TracingDetail{
				{
					Input: nil,
					Output: map[string]interface{}{
						"var1": 4,
					},
					Spec: &spec.Variable{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 4,
								},
							},
						},
					},
					OpType: types.VariableOpType,
				},
				{
					Input: nil,
					Output: map[string]interface{}{
						"var1": 5,
					},
					Spec: &spec.Variable{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 5,
								},
							},
						},
					},
					OpType: types.VariableOpType,
				},
			},
		},
		{
			name: "single jsonoutput",
			operationTracing: &OperationTracing{
				Specs: &spec.JsonOutput{
					JsonTemplate: &spec.JsonTemplate{
						Fields: []*spec.Field{
							{
								FieldName: "key1",
								Value: &spec.Field_Expression{
									Expression: "outputTable",
								},
							},
						},
					},
				},
				Input: []map[string]interface{}{nil},
				Output: []map[string]interface{}{
					{
						"key1": []map[string]interface{}{
							{
								"col1": 1,
							},
							{
								"col1": 2,
							},
						},
					},
				},
				OpType: types.JsonOutputOpType,
			},
			want: []types.TracingDetail{
				{
					Spec: &spec.JsonOutput{
						JsonTemplate: &spec.JsonTemplate{
							Fields: []*spec.Field{
								{
									FieldName: "key1",
									Value: &spec.Field_Expression{
										Expression: "outputTable",
									},
								},
							},
						},
					},
					Input: nil,
					Output: map[string]interface{}{
						"key1": []map[string]interface{}{
							{
								"col1": 1,
							},
							{
								"col1": 2,
							},
						},
					},
					OpType: types.JsonOutputOpType,
				},
			},
		},
		{
			name: "list of variable specs - number of inputs does not match with length of specs",
			operationTracing: &OperationTracing{
				Specs: []*spec.Variable{
					{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 4,
								},
							},
						},
					},
					{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 5,
								},
							},
						},
					},
				},
				Input: []map[string]interface{}{
					nil,
				},
				Output: []map[string]interface{}{
					{
						"var1": 4,
					},
					{
						"var1": 5,
					},
				},
				OpType: types.VariableOpType,
			},
			wantErr: fmt.Errorf("variable_op: number of inputs (1) does not match with number of specs (2)"),
		},
		{
			name: "list of variable specs - number of outputs does not match with length of specs",
			operationTracing: &OperationTracing{
				Specs: []*spec.Variable{
					{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 4,
								},
							},
						},
					},
					{
						Name: "var1",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{
									IntValue: 5,
								},
							},
						},
					},
				},
				Input: []map[string]interface{}{
					nil, nil,
				},
				Output: []map[string]interface{}{
					{
						"var1": 4,
					},
					{
						"var1": 5,
					},
					{
						"var1": 6,
					},
				},
				OpType: types.VariableOpType,
			},
			wantErr: fmt.Errorf("variable_op: number of outputs (3) does not match with number of specs (2)"),
		},
		{
			name: "single jsonoutput - no input recorded",
			operationTracing: &OperationTracing{
				Specs: &spec.JsonOutput{
					JsonTemplate: &spec.JsonTemplate{
						Fields: []*spec.Field{
							{
								FieldName: "key1",
								Value: &spec.Field_Expression{
									Expression: "outputTable",
								},
							},
						},
					},
				},
				Input: nil,
				Output: []map[string]interface{}{
					{
						"key1": []map[string]interface{}{
							{
								"col1": 1,
							},
							{
								"col1": 2,
							},
						},
					},
				},
				OpType: types.JsonOutputOpType,
			},
			wantErr: fmt.Errorf("json_output_op: input should has one record"),
		},
		{
			name: "single jsonoutput - no output recorded",
			operationTracing: &OperationTracing{
				Specs: &spec.JsonOutput{
					JsonTemplate: &spec.JsonTemplate{
						Fields: []*spec.Field{
							{
								FieldName: "key1",
								Value: &spec.Field_Expression{
									Expression: "outputTable",
								},
							},
						},
					},
				},
				Input: []map[string]interface{}{nil},
				Output: []map[string]interface{}{
					{
						"key1": []map[string]interface{}{
							{
								"col1": 1,
							},
							{
								"col1": 2,
							},
						},
					},
					{
						"key1": []map[string]interface{}{
							{
								"col1": 1,
							},
							{
								"col1": 2,
							},
						},
					},
				},
				OpType: types.JsonOutputOpType,
			},
			wantErr: fmt.Errorf("json_output_op: output should has one record"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := tt.operationTracing
			got, err := ot.GetOperationTracingDetail()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OperationTracing.GetOperationTracingDetail() = %v, want %v", got, tt.want)
			}
		})
	}
}
