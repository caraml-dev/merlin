package symbol

import (
	"encoding/json"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/types"
)

// JsonExtract extract json field pointed by nestedJsonPath within a json string pointed by nestedJsonPath
func (sr Registry) JsonExtract(parentJsonPath, nestedJsonPath string) interface{} {
	jsonBody, err := sr.evalArg(parentJsonPath)
	if err != nil {
		panic(err)
	}

	cplJsonPath := sr.getCompiledJSONPath(nestedJsonPath)
	if cplJsonPath == nil {
		c, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
			JsonPath: nestedJsonPath,
		})
		if err != nil {
			panic(err)
		}
		sr.addCompiledJsonPath(nestedJsonPath, c)
		cplJsonPath = c
	}

	var js types.JSONObject
	jsonBodyStr, ok := jsonBody.(string)
	if !ok {
		panic(fmt.Errorf("the value specified in path `%s` should be of string type", parentJsonPath))
	}

	if err := json.Unmarshal([]byte(jsonBodyStr), &js); err != nil {
		panic(fmt.Errorf("the value specified in path `%s` should be a valid JSON", parentJsonPath))
	}

	value, err := cplJsonPath.Lookup(js)
	if err != nil {
		panic(err)
	}

	return value
}
