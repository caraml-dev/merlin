package pipeline

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
)

func mustCompileExpression(expression string) *vm.Program {
	cpl, err := expr.Compile(expression)
	if err != nil {
		panic(err)
	}
	return cpl
}

const (
	rawRequestJson = `
		{
		  "signature_name" : "predict",
		  "instances": [
			{"sepal_length":2.8, "sepal_width":1.0, "petal_length":6.8, "petal_width":0.4},
			{"sepal_length":0.1, "sepal_width":0.5, "petal_length":1.8, "petal_width":2.4}
		  ],
          "array_int" : [1,2,3,4],
          "array_float": [1.1, 2.2, 3.3, 4.4],
          "array_float_2": [1.1, 2.2, 3.3, 4.4, 5.5],
          "int": 1234,
		  "long_int": 12345678910,
		  "float": 12345.567,
		  "string" : "hello"
		}
		`

	modelResponseJson = `
		{
		  "predictions": [
			1, 2
		  ],
          "model_name" : "iris-classifier"
		}
    `
)

func TestVariableDeclarationOp_Execute(t *testing.T) {
	type fields struct {
		variableSpec []*spec.Variable
	}

	compiledExpression := expression.NewStorage()
	compiledExpression.AddAll(map[string]*vm.Program{
		"Now()": mustCompileExpression("Now()"),
	})

	compiledJsonPath := jsonpath.NewStorage()
	compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
		"$.signature_name": jsonpath.MustCompileJsonPath("$.signature_name"),
	})

	var rawRequestJSON types.JSONObject
	err := json.Unmarshal([]byte(rawRequestJson), &rawRequestJSON)
	if err != nil {
		panic(err)
	}

	symbolRegistry := symbol.NewRegistryWithCompiledJSONPath(compiledJsonPath)
	symbolRegistry.SetRawRequestJSON(rawRequestJSON)
	tests := []struct {
		name         string
		fields       fields
		env          *Environment
		expVariables map[string]interface{}
		wantErr      bool
	}{
		{
			"literal variables declaration",
			fields{
				[]*spec.Variable{
					{
						Name: "myIntegerLiteral",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_IntValue{1},
							},
						},
					},
					{
						Name: "myFloatLiteral",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_FloatValue{1.2345},
							},
						},
					},
					{
						Name: "myStringLiteral",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_StringValue{"hello world"},
							},
						},
					},
				},
			},

			&Environment{
				symbolRegistry: symbol.NewRegistryWithCompiledJSONPath(compiledJsonPath),
			},
			map[string]interface{}{
				"myIntegerLiteral": int64(1),
				"myFloatLiteral":   float64(1.2345),
				"myStringLiteral":  "hello world",
			},
			false,
		},
		{
			"expression variables declaration",
			fields{
				[]*spec.Variable{
					{
						Name: "currentTime",
						Value: &spec.Variable_Expression{
							Expression: "Now()",
						},
					},
				},
			},

			&Environment{
				symbolRegistry: symbol.NewRegistryWithCompiledJSONPath(compiledJsonPath),
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
				},
			},
			map[string]interface{}{
				"currentTime": time.Now(),
			},
			false,
		},
		{
			"variables declaration using json path",
			fields{
				[]*spec.Variable{
					{
						Name: "signature_name",
						Value: &spec.Variable_JsonPath{
							JsonPath: "$.signature_name",
						},
					},
				},
			},

			&Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledJsonpath: compiledJsonPath,
				},
			},
			map[string]interface{}{
				"signature_name": "predict",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VariableDeclarationOp{
				variableSpec: tt.fields.variableSpec,
			}
			if err := v.Execute(context.Background(), tt.env); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			for varName, varValue := range tt.expVariables {
				switch v := varValue.(type) {
				case time.Time:
					assert.True(t, v.Sub(tt.env.symbolRegistry[varName].(time.Time)) < time.Second)
				default:
					assert.Equal(t, v, tt.env.symbolRegistry[varName])
				}
			}
		})
	}
}
