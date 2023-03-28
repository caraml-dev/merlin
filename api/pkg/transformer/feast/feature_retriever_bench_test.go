package feast

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer/feast/mocks"
	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	transTypes "github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/caraml-dev/merlin/pkg/transformer/types/expression"
)

// goos: darwin
// goarch: amd64
// pkg: github.com/caraml-dev/merlin/pkg/transformer/feast
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_buildEntitiesRequest_geohashArrays-8   	   94250	     12922 ns/op	    4818 B/op	     196 allocs/op
// PASS
func Benchmark_buildEntitiesRequest_geohashArrays(b *testing.B) {
	b.StopTimer()
	mockFeast := &mocks.Client{}
	feastClients := Clients{}
	feastClients[spec.ServingSource_BIGTABLE] = mockFeast

	logger, _ := zap.NewDevelopment()

	request := []byte(`{"merchants":[{"id": "M111", "latitude": 1.0, "longitude": 1.0}, {"id": "M222", "latitude": 2.0, "longitude": 2.0}]}`)

	sr := symbol.NewRegistry()
	featureTableSpecs := []*spec.FeatureTable{
		{
			Entities: []*spec.Entity{
				{
					Name:      "merchant_id",
					ValueType: "STRING",
					Extractor: &spec.Entity_JsonPath{
						JsonPath: "$.merchants[*].id",
					},
				},
				{
					Name:      "geohash",
					ValueType: "STRING",
					Extractor: &spec.Entity_Udf{
						Udf: "Geohash(\"$.merchants[*].latitude\", \"$.merchants[*].longitude\", 12)",
					},
				},
			},
		},
	}
	compiledJSONPaths, err := CompileJSONPaths(featureTableSpecs, jsonpath.Map)
	if err != nil {
		panic(err)
	}

	compiledExpressions, err := CompileExpressions(featureTableSpecs, symbol.NewRegistry())
	if err != nil {
		panic(err)
	}

	jsonPathStorage := jsonpath.NewStorage()
	jsonPathStorage.AddAll(compiledJSONPaths)
	expressionStorage := expression.NewStorage()
	expressionStorage.AddAll(compiledExpressions)
	entityExtractor := NewEntityExtractor(jsonPathStorage, expressionStorage)
	fr := NewFeastRetriever(feastClients,
		entityExtractor,
		featureTableSpecs,
		&Options{
			StatusMonitoringEnabled:       true,
			ValueMonitoringEnabled:        true,
			FeastClientHystrixCommandName: "Benchmark_buildEntitiesRequest_geohashArrays",
		},
		logger,
	)

	var requestJson transTypes.JSONObject
	err = json.Unmarshal(request, &requestJson)
	if err != nil {
		panic(err)
	}
	sr.SetRawRequest(requestJson)

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		request, _ := fr.buildEntityRows(sr, featureTableSpecs[0].Entities)
		_ = request
	}
}
