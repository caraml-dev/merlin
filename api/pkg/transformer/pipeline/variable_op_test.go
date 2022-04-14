package pipeline

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

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

func mustCompileExpressionWithEnv(expression string, env *Environment) *vm.Program {
	cpl, err := expr.Compile(expression, expr.Env(env.symbolRegistry), expr.Operator("&&", "AndOp"),
		expr.Operator("||", "OrOp"),
		expr.Operator(">", "GreaterOp"),
		expr.Operator(">=", "GreaterEqOp"),
		expr.Operator("<", "LessOp"),
		expr.Operator("<=", "LessEqOp"),
		expr.Operator("==", "EqualOp"),
		expr.Operator("!=", "NeqOp"),
		expr.Operator("+", "AddOp"),
		expr.Operator("-", "SubstractOp"),
		expr.Operator("*", "MultiplyOp"),
		expr.Operator("/", "DivideOp"),
		expr.Operator("%", "ModuloOp"),
	)
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
		  "object_with_null": [
			{"sepal_length":2.8, "petal_length":6.8, "petal_width":0.4},
			{"sepal_length":0.1, "sepal_width":0.5, "petal_width":2.4}
		  ],
          "one_row_array" : [
           {"sepal_length":0.1, "sepal_width":0.5, "petal_length":1.8, "petal_width":2.4}
          ],
          "array_int" : [1,2,3,4],
          "array_float": [1.1, 2.2, 3.3, 4.4],
          "array_float_2": [1.1, 2.2, 3.3, 4.4, 5.5],
          "int": 1234,
		  "long_int": 12345678910,
		  "float": 12345.567,
		  "string" : "hello",
		  "details":"{\"points\": [{\"distanceInMeter\": 0.0}, {\"distanceInMeter\": 8976.0}, {\"distanceInMeter\": 729.0}, {\"distanceInMeter\": 8573.0}]}"
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

	logger, _ := zap.NewDevelopment()
	compiledExpression := expression.NewStorage()
	compiledExpression.AddAll(map[string]*vm.Program{
		"Now()": mustCompileExpression("Now()"),
	})

	compiledJsonPath := jsonpath.NewStorage()
	compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
		"$.signature_name": jsonpath.MustCompileJsonPath("$.signature_name"),
		"$.not_exist_key": jsonpath.MustCompileJsonPathWithOption(jsonpath.JsonPathOption{
			JsonPath:     "$.not_exist_key",
			DefaultValue: "2.2",
			TargetType:   spec.ValueType_FLOAT,
		}),
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
					{
						Name: "myBoolLiteral",
						Value: &spec.Variable_Literal{
							Literal: &spec.Literal{
								LiteralValue: &spec.Literal_BoolValue{true},
							},
						},
					},
				},
			},

			&Environment{
				symbolRegistry: symbol.NewRegistryWithCompiledJSONPath(compiledJsonPath),
				logger:         logger,
			},
			map[string]interface{}{
				"myIntegerLiteral": int64(1),
				"myFloatLiteral":   float64(1.2345),
				"myStringLiteral":  "hello world",
				"myBoolLiteral":    true,
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
				logger: logger,
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
				logger: logger,
			},
			map[string]interface{}{
				"signature_name": "predict",
			},
			false,
		},
		{
			"variables declaration using json path config",
			fields{
				[]*spec.Variable{
					{
						Name: "signature_name",
						Value: &spec.Variable_JsonPathConfig{
							JsonPathConfig: &spec.FromJson{
								JsonPath: "$.signature_name",
							},
						},
					},
					{
						Name: "value",
						Value: &spec.Variable_JsonPathConfig{
							JsonPathConfig: &spec.FromJson{
								JsonPath:     "$.not_exist_key",
								DefaultValue: "2.2",
								ValueType:    spec.ValueType_FLOAT,
							},
						},
					},
				},
			},

			&Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledJsonpath: compiledJsonPath,
				},
				logger: logger,
			},
			map[string]interface{}{
				"signature_name": "predict",
				"value":          2.2,
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
