package pipeline

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
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
		  ]
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

	compiledExpression := map[string]*vm.Program{
		"Now()": mustCompileExpression("Now()"),
	}

	compiledJsonPath := map[string]*jsonpath.CompiledJSONPath{
		"$.signature_name": jsonpath.MustCompileJsonPath("$.signature_name"),
	}

	var rawRequestData types.UnmarshalledJSON
	json.Unmarshal([]byte(rawRequestJson), &rawRequestData)

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
				symbolRegistry: NewRegistry(),
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
				symbolRegistry: NewRegistry(),
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
				symbolRegistry: NewRegistry(),
				compiledPipeline: &CompiledPipeline{
					compiledJsonpath: compiledJsonPath,
				},
				sourceJSONs: map[spec.FromJson_SourceEnum]types.UnmarshalledJSON{
					spec.FromJson_RAW_REQUEST: rawRequestData,
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
			if err := v.Execute(tt.env); (err != nil) != tt.wantErr {
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
