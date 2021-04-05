package main

import (
	"context"
	"fmt"
	"net"

	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"google.golang.org/grpc"

	"github.com/gojek/merlin/log"
)

const (
	grpcPort = 6565
)

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
}

func (m mockFeastOnlineServing) GetFeastServingInfo(ctx context.Context, req *serving.GetFeastServingInfoRequest) (*serving.GetFeastServingInfoResponse, error) {
	return &serving.GetFeastServingInfoResponse{}, nil
}

// GetOnlineFeaturesV2 return a valid GetOnlineFeaturesResponse but with all features having GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
func (m mockFeastOnlineServing) GetOnlineFeaturesV2(ctx context.Context, req *serving.GetOnlineFeaturesRequestV2) (*serving.GetOnlineFeaturesResponse, error) {
	var fieldValues []*serving.GetOnlineFeaturesResponse_FieldValues
	for _, entity := range req.EntityRows {
		fields := entity.Fields
		status := make(map[string]serving.GetOnlineFeaturesResponse_FieldStatus)

		for k := range entity.Fields {
			status[k] = serving.GetOnlineFeaturesResponse_PRESENT
		}

		for _, feature := range req.Features {
			fields[feature.Name] = nil
			status[feature.Name] = serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
		}

		fv := &serving.GetOnlineFeaturesResponse_FieldValues{
			Fields:   fields,
			Statuses: status,
		}
		fieldValues = append(fieldValues, fv)
	}

	return &serving.GetOnlineFeaturesResponse{
		FieldValues: fieldValues,
	}, nil
}

func newMockFeastOnlineServing() serving.ServingServiceServer {
	return &mockFeastOnlineServing{}
}
