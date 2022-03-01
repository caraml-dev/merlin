package main

import (
	"context"
	"fmt"
	"net"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"google.golang.org/grpc"

	"github.com/gojek/merlin/log"
	feastType2 "github.com/gojek/merlin/pkg/transformer/types/feast"
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

type mockFeastOnlineServing struct{}

func (m mockFeastOnlineServing) GetFeastServingInfo(ctx context.Context, req *serving.GetFeastServingInfoRequest) (*serving.GetFeastServingInfoResponse, error) {
	return &serving.GetFeastServingInfoResponse{}, nil
}

// GetOnlineFeaturesV2 return a valid GetOnlineFeaturesResponse but with all features having GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
func (m mockFeastOnlineServing) GetOnlineFeaturesV2(ctx context.Context, req *serving.GetOnlineFeaturesRequestV2) (*serving.GetOnlineFeaturesResponse, error) {
	log.Infof("GetOnlineFeaturesV2.Request: %+v", req)

	var fieldValues []*serving.GetOnlineFeaturesResponse_FieldValues
	for _, entity := range req.EntityRows {
		fields := entity.Fields
		status := make(map[string]serving.GetOnlineFeaturesResponse_FieldStatus)

		for k := range entity.Fields {
			status[k] = serving.GetOnlineFeaturesResponse_PRESENT
		}

		for _, feature := range req.Features {
			featureName := feature.FeatureTable + ":" + feature.Name

			status[featureName] = serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
			fields[featureName] = nil

			if val, ok := mockOutput[feature.Name]; ok {
				status[featureName] = serving.GetOnlineFeaturesResponse_PRESENT
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

func newMockFeastOnlineServing() serving.ServingServiceServer {
	return &mockFeastOnlineServing{}
}
