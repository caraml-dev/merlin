package main

import (
	"context"
	"fmt"
	"net"
	"sort"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"google.golang.org/grpc"

	"github.com/caraml-dev/merlin/log"
	feastType2 "github.com/caraml-dev/merlin/pkg/transformer/types/feast"
)

const (
	grpcPort = 6565
)

var mockOutput = map[string]*types.Value{
	"test_double":      feast.DoubleVal(0.201),
	"test_double_list": feastType2.DoubleListVal([]float64{3.1415, 4.5678}),
	"test_int64":       feast.Int64Val(64),
	"test_int64_list":  feastType2.Int64ListVal([]int64{128, 256}),
	"test_string":      feast.StrVal("hello"),
	"test_string_list": feastType2.StrListVal([]string{"hello", "world"}),
}

func main() {
	log.Infof("running mock feast serving at port: %d", grpcPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := newMockFeastOnlineServing()
	grpcServer := grpc.NewServer()
	serving.RegisterServingServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

type mockFeastOnlineServing struct {
	serving.UnimplementedServingServiceServer
}

func (m *mockFeastOnlineServing) GetFeastServingInfo(ctx context.Context, req *serving.GetFeastServingInfoRequest) (*serving.GetFeastServingInfoResponse, error) {
	return &serving.GetFeastServingInfoResponse{}, nil
}

// GetOnlineFeaturesV2 return a valid GetOnlineFeaturesResponse but with all features having GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
func (m *mockFeastOnlineServing) GetOnlineFeaturesV2(ctx context.Context, req *serving.GetOnlineFeaturesRequestV2) (*serving.GetOnlineFeaturesResponse, error) {
	log.Infof("GetOnlineFeaturesV2.Request: %+v", req)

	var fieldValues []*serving.GetOnlineFeaturesResponse_FieldValues
	for _, entity := range req.EntityRows {
		fields := entity.Fields
		status := make(map[string]serving.FieldStatus)

		for k := range entity.Fields {
			status[k] = serving.FieldStatus_PRESENT
		}

		for _, feature := range req.Features {
			featureName := feature.FeatureTable + ":" + feature.Name

			status[featureName] = serving.FieldStatus_OUTSIDE_MAX_AGE
			fields[featureName] = nil

			if val, ok := mockOutput[feature.Name]; ok {
				status[featureName] = serving.FieldStatus_PRESENT
				fields[featureName] = val
			}
		}

		fv := &serving.GetOnlineFeaturesResponse_FieldValues{
			Fields:   fields,
			Statuses: status,
		}
		fieldValues = append(fieldValues, fv)
	}

	resp := &serving.GetOnlineFeaturesResponse{
		FieldValues: fieldValues,
	}
	log.Infof("GetOnlineFeaturesV2.Response: %+v", resp)

	return resp, nil
}

func (m *mockFeastOnlineServing) GetOnlineFeatures(ctx context.Context, req *serving.GetOnlineFeaturesRequestV2) (*serving.GetOnlineFeaturesResponseV2, error) {
	log.Infof("GetOnlineFeatures.Request: %+v", req)
	if len(req.EntityRows) == 0 {
		log.Panicf("There should be at least one row in the entity rows")
	}
	sortedEntityFieldNames := make([]string, len(req.EntityRows[0].Fields))
	cnt := 0
	for fieldName := range req.EntityRows[0].Fields {
		sortedEntityFieldNames[cnt] = fieldName
		cnt++
	}
	sort.Strings(sortedEntityFieldNames)
	featureReference := make([]string, len(req.Features))
	for index, feature := range req.Features {
		featureReference[index] = feature.Name
	}

	var fieldVectors []*serving.GetOnlineFeaturesResponseV2_FieldVector
	for entityIndex, entity := range req.EntityRows {
		status := make([]serving.FieldStatus, len(sortedEntityFieldNames)+len(req.Features))
		entityFeatureValues := make([]*types.Value, len(status))

		for index, field := range sortedEntityFieldNames {
			entityFeatureValues[index] = entity.Fields[field]
			status[index] = serving.FieldStatus_PRESENT
		}

		for index, feature := range req.Features {
			status[index+len(sortedEntityFieldNames)] = serving.FieldStatus_OUTSIDE_MAX_AGE
			entityFeatureValues[index+len(sortedEntityFieldNames)] = nil

			if val, ok := mockOutput[feature.Name]; ok {
				status[index+len(sortedEntityFieldNames)] = serving.FieldStatus_PRESENT
				entityFeatureValues[index+len(sortedEntityFieldNames)] = val
			}
		}
		fieldVector := &serving.GetOnlineFeaturesResponseV2_FieldVector{
			Values:   entityFeatureValues,
			Statuses: status,
		}
		fieldVectors[entityIndex] = fieldVector
	}

	resp := &serving.GetOnlineFeaturesResponseV2{
		Metadata: &serving.GetOnlineFeaturesResponseMetadata{
			FieldNames: &serving.FieldList{
				Val: append(sortedEntityFieldNames, featureReference...),
			},
		},
		Results: fieldVectors,
	}
	log.Infof("GetOnlineFeaturesV2.Response: %+v", resp)

	return resp, nil
}

func newMockFeastOnlineServing() serving.ServingServiceServer {
	return &mockFeastOnlineServing{}
}
