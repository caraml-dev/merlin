package feast

import (
	"context"
	"fmt"
	"strings"

	feastsdk "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/go-coldbrew/grpcpool"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcClient as wrapper of feastsdk
// TO-DO: Remove this after official feastsdk support connection pool
type GrpcClient struct {
	cli          serving.ServingServiceClient
	conn         grpcpool.ConnPool
	waitForReady bool
}

func newInsecureGRPCClientWithDialOptions(host string, port int, numConn int, waitForReady bool, opts ...grpc.DialOption) (*GrpcClient, error) {
	feastCli := &GrpcClient{}
	adr := fmt.Sprintf("%s:%d", host, port)

	// Compile grpc dial options from security config.
	options := append(opts, []grpc.DialOption{grpc.WithStatsHandler(&ocgrpc.ClientHandler{}), grpc.WithTransportCredentials(insecure.NewCredentials())}...)
	// Configure client TLS.

	// Enable tracing if a global tracer is registered
	tracingInterceptor := grpc.WithUnaryInterceptor(
		otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer()))
	options = append(options, tracingInterceptor)

	conn, err := grpcpool.DialContext(context.Background(), adr, uint(numConn), options...)
	if err != nil {
		return nil, err
	}
	feastCli.cli = serving.NewServingServiceClient(conn)
	feastCli.conn = conn
	feastCli.waitForReady = waitForReady
	return feastCli, nil
}

func (fc *GrpcClient) GetOnlineFeatures(ctx context.Context, req *feastsdk.OnlineFeaturesRequest) (*feastsdk.OnlineFeaturesResponse, error) {
	featuresRequest, err := buildRequest(*req)
	if err != nil {
		return nil, err
	}

	resp, err := fc.cli.GetOnlineFeatures(ctx, featuresRequest, grpc.WaitForReady(fc.waitForReady))

	// collect unique entity refs from entity rows
	entityRefs := make(map[string]struct{})
	for _, entityRows := range req.Entities {
		for ref := range entityRows {
			entityRefs[ref] = struct{}{}
		}
	}
	return &feastsdk.OnlineFeaturesResponse{RawResponse: resp}, err
}

// GetFeastServingInfo gets information about the feast serving instance this client is connected to.
func (fc *GrpcClient) GetFeastServingInfo(ctx context.Context, in *serving.GetFeastServingInfoRequest) (
	*serving.GetFeastServingInfoResponse, error) {
	return fc.cli.GetFeastServingInfo(ctx, in, grpc.WaitForReady(fc.waitForReady))
}

// Close the grpc connection.
func (fc *GrpcClient) Close() error {
	return fc.conn.Close()
}

// Builds the feast-specified request payload from the wrapper.
func buildRequest(r feastsdk.OnlineFeaturesRequest) (*serving.GetOnlineFeaturesRequestV2, error) {
	featureRefs, err := buildFeatureRefs(r.Features)
	if err != nil {
		return nil, err
	}

	// build request entity rows from native entities
	entityRows := make([]*serving.GetOnlineFeaturesRequestV2_EntityRow, len(r.Entities))
	for i, entity := range r.Entities {
		entityRows[i] = &serving.GetOnlineFeaturesRequestV2_EntityRow{
			Fields: entity,
		}
	}

	return &serving.GetOnlineFeaturesRequestV2{
		Features:   featureRefs,
		EntityRows: entityRows,
		Project:    r.Project,
	}, nil
}

// Creates a slice of FeatureReferences from string representation in
// the format featuretable:feature.
// featureRefStrs - string feature references to parse.
// Returns parsed FeatureReferences.
// Returns an error when the format of the string feature reference is invalid
func buildFeatureRefs(featureRefStrs []string) ([]*serving.FeatureReferenceV2, error) {
	var featureRefs []*serving.FeatureReferenceV2

	for _, featureRefStr := range featureRefStrs {
		featureRef, err := parseFeatureRef(featureRefStr)
		if err != nil {
			return nil, err
		}
		featureRefs = append(featureRefs, featureRef)
	}
	return featureRefs, nil
}

// Parses a string FeatureReference into FeatureReference proto
// featureRefStr - the string feature reference to parse.
// Returns parsed FeatureReference.
// Returns an error when the format of the string feature reference is invalid
func parseFeatureRef(featureRefStr string) (*serving.FeatureReferenceV2, error) {
	if len(featureRefStr) == 0 {
		return nil, fmt.Errorf(feastsdk.ErrInvalidFeatureRef, featureRefStr)
	}

	var featureRef serving.FeatureReferenceV2
	if strings.Contains(featureRefStr, "/") || !strings.Contains(featureRefStr, ":") {
		return nil, fmt.Errorf(feastsdk.ErrInvalidFeatureRef, featureRefStr)
	}
	// parse featuretable if specified
	if strings.Contains(featureRefStr, ":") {
		refSplit := strings.Split(featureRefStr, ":")
		featureRef.FeatureTable, featureRefStr = refSplit[0], refSplit[1]
	}
	featureRef.Name = featureRefStr

	return &featureRef, nil
}
