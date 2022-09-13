package feast

import (
	"encoding/json"
	"testing"

	"github.com/antonmedv/expr"
	feast "github.com/feast-dev/feast/sdk/go"
	feastType "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/mmcloughlin/geohash"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	transTypes "github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

func TestEntityExtractor_ExtractValuesFromSymbolRegistry(t *testing.T) {
	testData := []byte(`{
		"integer" : 1234,
		"float" : 1234.111,
		"string" : "1234",
		"boolean" : true,
		"booleanString" : "false",
		"latitude": 1.0,
		"longitude": 2.0,
		"longInteger": 12986086,
		"locations": [
		 {
			"latitude": 1.0,
			"longitude": 2.0
		 },
		 {
			"latitude": 1.0,
			"longitude": 2.0
 		 }
        ],
		"details": "{\"merchant_id\": 9001}",
		"struct" : {
                "integer" : 1234,
				"float" : 1234.111,
				"string" : "value",
				"boolean" : true
		},
		"array" : [
		{
                "integer" : 1111,
				"float" : 1111.1111,
				"string" : "value1",
				"boolean" : true
		},
		{
                "integer" : 2222,
				"float" : 2222.2222,
				"string" : "value2",
				"boolean" : false
		}]
	}`)

	tests := []struct {
		name         string
		entityConfig *spec.Entity
		expValues    []*feastType.Value
		expError     error
		variables    map[string]interface{}
	}{
		{
			name: "integer to int64",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT64",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.integer",
				},
			},
			expValues: []*feastType.Value{
				feast.Int64Val(1234),
			},
			expError: nil,
		},
		{
			name: "integer to int64 -- jsonpath not found key",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT64",
				Extractor: &spec.Entity_JsonPathConfig{
					JsonPathConfig: &spec.FromJson{
						JsonPath:     "$.not_found_key",
						DefaultValue: "2",
						ValueType:    spec.ValueType_INT,
					},
				},
			},
			expValues: []*feastType.Value{
				feast.Int64Val(2),
			},
			expError: nil,
		},
		{
			name: "integer to float",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "FLOAT",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.integer",
				},
			},
			expValues: []*feastType.Value{
				feast.FloatVal(1234),
			},
			expError: nil,
		},
		{
			name: "integer to double",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "DOUBLE",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.integer",
				},
			},
			expValues: []*feastType.Value{
				feast.DoubleVal(1234),
			},
			expError: nil,
		},
		{
			name: "integer to string",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "STRING",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.integer",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1234"),
			},
			expError: nil,
		},
		{
			name: "long integer to string",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "STRING",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.longInteger",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("12986086"),
			},
			expError: nil,
		},
		{
			name: "float to int32",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT32",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.float",
				},
			},
			expValues: []*feastType.Value{
				feast.Int32Val(1234),
			},
			expError: nil,
		},
		{
			name: "float to int64",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT64",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.float",
				},
			},
			expValues: []*feastType.Value{
				feast.Int64Val(1234),
			},
			expError: nil,
		},
		{
			name: "float to float",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "FLOAT",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.float",
				},
			},
			expValues: []*feastType.Value{
				feast.FloatVal(1234.111),
			},
			expError: nil,
		},
		{
			name: "float to double",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "DOUBLE",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.float",
				},
			},
			expValues: []*feastType.Value{
				feast.DoubleVal(1234.111),
			},
			expError: nil,
		},
		{
			name: "float to string",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "STRING",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.float",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1234"),
			},
			expError: nil,
		},
		{
			name: "string to int32",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT32",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.string",
				},
			},
			expValues: []*feastType.Value{
				feast.Int32Val(1234),
			},
			expError: nil,
		},
		{
			name: "string to int64",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT64",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.string",
				},
			},
			expValues: []*feastType.Value{
				feast.Int64Val(1234),
			},
			expError: nil,
		},
		{
			name: "string to float",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "FLOAT",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.string",
				},
			},
			expValues: []*feastType.Value{
				feast.FloatVal(1234),
			},
			expError: nil,
		},
		{
			name: "string to double",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "DOUBLE",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.string",
				},
			},
			expValues: []*feastType.Value{
				feast.DoubleVal(1234),
			},
			expError: nil,
		},
		{
			name: "string to string",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "STRING",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.string",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1234"),
			},
			expError: nil,
		},
		{
			name: "boolean to boolean",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "BOOL",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.boolean",
				},
			},
			expValues: []*feastType.Value{
				feast.BoolVal(true),
			},
			expError: nil,
		},
		{
			name: "string to boolean",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "BOOL",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.booleanString",
				},
			},
			expValues: []*feastType.Value{
				feast.BoolVal(false),
			},
			expError: nil,
		},
		{
			name: "array of integer",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT32",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.array[*].integer",
				},
			},
			expValues: []*feastType.Value{
				feast.Int32Val(1111),
				feast.Int32Val(2222),
			},
			expError: nil,
		},
		{
			name: "array of string",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "STRING",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.array[*].string",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("value1"),
				feast.StrVal("value2"),
			},
			expError: nil,
		},
		{
			name: "struct integer",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "INT32",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.struct.integer",
				},
			},
			expValues: []*feastType.Value{
				feast.Int32Val(1234),
			},
			expError: nil,
		},
		{
			name: "Geohash udf",
			entityConfig: &spec.Entity{
				Name:      "my_geohash",
				ValueType: "STRING",
				Extractor: &spec.Entity_Udf{
					Udf: "Geohash(\"$.latitude\", \"$.longitude\", 12)",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal(geohash.Encode(1.0, 2.0)),
			},
			expError: nil,
		},
		{
			name: "Array of Geohash udf",
			entityConfig: &spec.Entity{
				Name:      "my_geohash",
				ValueType: "STRING",
				Extractor: &spec.Entity_Udf{
					Udf: "Geohash(\"$.locations[*].latitude\", \"$.locations[*].longitude\", 12)",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal(geohash.Encode(1.0, 2.0)),
				feast.StrVal(geohash.Encode(1.0, 2.0)),
			},
			expError: nil,
		},
		{
			name: "JsonExtract udf",
			entityConfig: &spec.Entity{
				Name:      "jsonextract",
				ValueType: "STRING",
				Extractor: &spec.Entity_Udf{
					Udf: "JsonExtract(\"$.details\", \"$.merchant_id\")",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("9001"),
			},
			expError: nil,
		},
		{
			name: "S2ID udf",
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "STRING",
				Extractor: &spec.Entity_Udf{
					Udf: "S2ID(\"$.latitude\", \"$.longitude\", 12)",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1154732743855177728"),
			},
			expError: nil,
		},
		{
			name: "Array of S2ID udf",
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "STRING",
				Extractor: &spec.Entity_Udf{
					Udf: "S2ID(\"$.locations[*].latitude\", \"$.locations[*].longitude\", 12)",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1154732743855177728"),
				feast.StrVal("1154732743855177728"),
			},
			expError: nil,
		},
		{
			name: "unsupported feast type",
			entityConfig: &spec.Entity{
				Name:      "my_entity",
				ValueType: "BYTES",
				Extractor: &spec.Entity_JsonPath{
					JsonPath: "$.booleanString",
				},
			},
			expValues: nil,
			expError:  errors.New("unsupported type BYTES"),
		},
		{
			name: "S2ID expression",
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "STRING",
				Extractor: &spec.Entity_Expression{
					Expression: "S2ID(\"$.latitude\", \"$.longitude\", 12)",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1154732743855177728"),
			},
			expError: nil,
		},
		{
			name: "Array of S2ID expression",
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "STRING",
				Extractor: &spec.Entity_Expression{
					Expression: "S2ID(\"$.locations[*].latitude\", \"$.locations[*].longitude\", 12)",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1154732743855177728"),
				feast.StrVal("1154732743855177728"),
			},
			expError: nil,
		},
		{
			name: "Entity from string variable",
			variables: map[string]interface{}{
				"string_var": "my_string",
			},
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "STRING",
				Extractor: &spec.Entity_Expression{
					Expression: "string_var",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("my_string"),
			},
			expError: nil,
		},
		{
			name: "Entity from int64 variable",
			variables: map[string]interface{}{
				"int64_var": int64(1234),
			},
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "INT64",
				Extractor: &spec.Entity_Expression{
					Expression: "int64_var",
				},
			},
			expValues: []*feastType.Value{
				feast.Int64Val(int64(1234)),
			},
			expError: nil,
		},
		{
			name: "Entity from float64 variable",
			variables: map[string]interface{}{
				"float64_var": float64(1234.5678),
			},
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "DOUBLE",
				Extractor: &spec.Entity_Expression{
					Expression: "float64_var",
				},
			},
			expValues: []*feastType.Value{
				feast.DoubleVal(float64(1234.5678)),
			},
			expError: nil,
		},
		{
			name: "Entity from string series",
			variables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
				),
			},
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "STRING",
				Extractor: &spec.Entity_Expression{
					Expression: "my_table.Col('string_col')",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1111"),
				feast.StrVal("2222"),
				feast.StrVal("3333"),
			},
			expError: nil,
		},
		{
			name: "Entity from int series",
			variables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
				),
			},
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "INT64",
				Extractor: &spec.Entity_Expression{
					Expression: "my_table.Col('int_col')",
				},
			},
			expValues: []*feastType.Value{
				feast.Int64Val(1111),
				feast.Int64Val(2222),
				feast.Int64Val(3333),
			},
			expError: nil,
		},
		{
			name: "Entity from float series",
			variables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
				),
			},
			entityConfig: &spec.Entity{
				Name:      "s2id",
				ValueType: "STRING",
				Extractor: &spec.Entity_Expression{
					Expression: "my_table.Col('int_col')",
				},
			},
			expValues: []*feastType.Value{
				feast.StrVal("1111"),
				feast.StrVal("2222"),
				feast.StrVal("3333"),
			},
			expError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compiledJsonPaths := jsonpath.NewStorage()
			compiledExpressions := expression.NewStorage()

			var nodesBody transTypes.JSONObject
			err := json.Unmarshal(testData, &nodesBody)
			if err != nil {
				panic(err)
			}

			sr := symbol.NewRegistryWithCompiledJSONPath(compiledJsonPaths)
			sr.SetRawRequest(nodesBody)
			for name, value := range test.variables {
				sr[name] = value
			}

			switch test.entityConfig.Extractor.(type) {
			case *spec.Entity_JsonPath:
				compiledJsonPath, _ := jsonpath.Compile(test.entityConfig.GetJsonPath())
				compiledJsonPaths.Set(test.entityConfig.GetJsonPath(), compiledJsonPath)
			case *spec.Entity_JsonPathConfig:
				compiledJsonPath, _ := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
					JsonPath:     test.entityConfig.GetJsonPathConfig().JsonPath,
					DefaultValue: test.entityConfig.GetJsonPathConfig().DefaultValue,
					TargetType:   test.entityConfig.GetJsonPathConfig().ValueType,
				})
				compiledJsonPaths.Set(test.entityConfig.GetJsonPathConfig().GetJsonPath(), compiledJsonPath)
			case *spec.Entity_Udf:
				compiledUdf, _ := expr.Compile(test.entityConfig.GetUdf(), expr.Env(sr), expr.AllowUndefinedVariables())
				compiledExpressions.Set(test.entityConfig.GetUdf(), compiledUdf)
			case *spec.Entity_Expression:
				compiledExpression, _ := expr.Compile(test.entityConfig.GetExpression(), expr.Env(sr), expr.AllowUndefinedVariables())
				compiledExpressions.Set(test.entityConfig.GetExpression(), compiledExpression)
			}

			er := NewEntityExtractor(compiledJsonPaths, compiledExpressions)

			actual, err := er.ExtractValuesFromSymbolRegistry(sr, test.entityConfig)
			if err != nil {
				if test.expError != nil {
					assert.EqualError(t, err, test.expError.Error())
					return
				} else {
					assert.Fail(t, err.Error())
				}
			}
			assert.Equal(t, test.expValues, actual)
		})
	}
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_100Entity(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "",
		ValueType: "INT32",
		Extractor: &spec.Entity_JsonPath{
			JsonPath: "$.array[*].id",
		},
	}

	doRunBenchmark(b, entityConfig)
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_1StringEntity(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "",
		ValueType: "STRING",
		Extractor: &spec.Entity_JsonPath{
			JsonPath: "$.string",
		},
	}
	doRunBenchmark(b, entityConfig)
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_1IntegerEntity(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "",
		ValueType: "INT32",
		Extractor: &spec.Entity_JsonPath{
			JsonPath: "$.integer",
		},
	}
	doRunBenchmark(b, entityConfig)
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_1FloatEntity(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "",
		ValueType: "DOUBLE",
		Extractor: &spec.Entity_JsonPath{
			JsonPath: "$.float",
		},
	}

	doRunBenchmark(b, entityConfig)
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_1GeohashUdf(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "my_geohash",
		ValueType: "STRING",
		Extractor: &spec.Entity_Udf{
			Udf: "Geohash(\"$.latitude\", \"$.longitude\", 7)",
		},
	}

	doRunBenchmark(b, entityConfig)
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_100GeohashUdf(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "my_geohash",
		ValueType: "STRING",
		Extractor: &spec.Entity_Udf{
			Udf: "Geohash(\"$.array[*].latitude\", \"$.array[*].longitude\", 7)",
		},
	}

	doRunBenchmark(b, entityConfig)
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_1S2IDUdf(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "my_geohash",
		ValueType: "STRING",
		Extractor: &spec.Entity_Udf{
			Udf: "S2ID(\"$.latitude\", \"$.longitude\", 11)",
		},
	}

	doRunBenchmark(b, entityConfig)
}

func BenchmarkEntityExtractor_ExtractValuesFromSymbolRegistry_100S2IDUdf(b *testing.B) {
	entityConfig := &spec.Entity{
		Name:      "my_geohash",
		ValueType: "STRING",
		Extractor: &spec.Entity_Udf{
			Udf: "S2ID(\"$.array[*].latitude\", \"$.array[*].longitude\", 11)",
		},
	}

	doRunBenchmark(b, entityConfig)
}

func doRunBenchmark(b *testing.B, entityConfig *spec.Entity) {
	b.StopTimer()
	compiledJsonPaths := jsonpath.NewStorage()
	compiledExpressions := expression.NewStorage()

	switch entityConfig.Extractor.(type) {
	case *spec.Entity_JsonPath:
		c, err := jsonpath.Compile(entityConfig.GetJsonPath())
		if err != nil {
			panic(err)
		}
		compiledJsonPaths.Set(entityConfig.GetJsonPath(), c)
	case *spec.Entity_Expression, *spec.Entity_Udf:
		exp := getExpressionExtractor(entityConfig)
		p, err := expr.Compile(exp, expr.Env(symbol.NewRegistry()), expr.AllowUndefinedVariables())
		if err != nil {
			panic(err)
		}
		compiledExpressions.Set(exp, p)
	}

	var nodesBody transTypes.JSONObject
	err := json.Unmarshal(benchData, &nodesBody)
	if err != nil {
		panic(err)
	}

	sr := symbol.NewRegistryWithCompiledJSONPath(compiledJsonPaths)
	sr.SetRawRequest(nodesBody)
	er := NewEntityExtractor(compiledJsonPaths, compiledExpressions)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Result, _ = er.ExtractValuesFromSymbolRegistry(sr, entityConfig)
	}
}

var (
	Result    []*feastType.Value
	benchData = []byte(`{
"string": "string_value",
"integer" : 1234,
"float" : 1234.111,
"latitude": 103.3,
"longitude": 1.0,
"array": [
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 0
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 1
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 2
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 3
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 4
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 5
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 6
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 7
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 8
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 9
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 10
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 11
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 12
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 13
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 14
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 15
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 16
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 17
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 18
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 19
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 20
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 21
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 22
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 23
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 24
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 25
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 26
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 27
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 28
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 29
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 30
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 31
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 32
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 33
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 34
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 35
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 36
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 37
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 38
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 39
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 40
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 41
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 42
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 43
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 44
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 45
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 46
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 47
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 48
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 49
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 50
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 51
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 52
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 53
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 54
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 55
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 56
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 57
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 58
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 59
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 60
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 61
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 62
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 63
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 64
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 65
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 66
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 67
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 68
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 69
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 70
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 71
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 72
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 73
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 74
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 75
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 76
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 77
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 78
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 79
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 80
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 81
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 82
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 83
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 84
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 85
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 86
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 87
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 88
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 89
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 90
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 91
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 92
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 93
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 94
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 95
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 96
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 97
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 98
  },
  {
    "latitude" : 103.2,
    "longitude" : 1.0,
    "id": 99
  }
]
}`)
)
