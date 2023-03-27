package symbol

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/stretchr/testify/assert"
)

func TestSymbolRegistry_JsonExtract(t *testing.T) {
	testJsonString := []byte(`{
		"details": "{\"merchant_id\": 9001}",
		"nested": "{\"child_node\": { \"grandchild_node\": \"gen-z\"}}",
		"not_json": "i_am_not_json_string",
		"not_string": 1024,
		"array": "{\"child_node\": { \"array\": [1, 2]}}"
	}`)

	var jsonObject types.JSONObject
	err := json.Unmarshal(testJsonString, &jsonObject)
	if err != nil {
		panic(err)
	}

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequest(jsonObject)

	type args struct {
		parentJsonPath string
		nestedJsonPath string
	}
	tests := []struct {
		name       string
		args       args
		want       interface{}
		wantErr    bool
		wantErrVal error
	}{
		{
			"should be able to extract value from JSON string",
			args{
				"$.details",
				"$.merchant_id",
			},
			float64(9001),
			false,
			nil,
		},
		{
			"should be able to extract value using nested key from JSON string",
			args{
				"$.nested",
				"$.child_node.grandchild_node",
			},
			"gen-z",
			false,
			nil,
		},
		{
			"should be able to extract array value using nested key from JSON string",
			args{
				"$.array",
				"$.child_node.array[*]",
			},
			[]interface{}{float64(1), float64(2)},
			false,
			nil,
		},
		{
			"should not throw error when value specified by key does not exist in nested JSON",
			args{
				"$.nested",
				"$.child_node.does_not_exist_node",
			},
			nil,
			false,
			nil,
		},
		{
			"should throw error when value obtained by key is not valid json",
			args{
				"$.not_json",
				"$.not_exist",
			},
			nil,
			true,
			fmt.Errorf("the value specified in path `$.not_json` should be a valid JSON"),
		},
		{
			"should throw error when value obtained by key is not string",
			args{
				"$.not_string",
				"$.not_exist",
			},
			nil,
			true,
			errors.New("the value specified in path `$.not_string` should be of string type"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.wantErrVal.Error(), func() {
					sr.JsonExtract(tt.args.parentJsonPath, tt.args.nestedJsonPath)
				})
				return
			}

			if got := sr.JsonExtract(tt.args.parentJsonPath, tt.args.nestedJsonPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonExtract() = %v, want %v", got, tt.want)
			}
		})
	}
}
