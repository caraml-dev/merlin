package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	upiV1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

type UPIServer struct {
	upiV1.UnimplementedUniversalPredictionServiceServer

	opts        *Options
	modelClient upiV1.UniversalPredictionServiceClient
	conn        *grpc.ClientConn

	ContextModifier    func(ctx context.Context) context.Context
	PreprocessHandler  func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error)
	PostprocessHandler func(ctx context.Context, response types.Payload, responseHeaders map[string]string) (types.Payload, error)
}

func NewUPIServer(opts *Options, handler *pipeline.Handler, logger *zap.Logger) (*UPIServer, error) {
	conn, err := grpc.Dial(opts.ModelPredictURL)
	if err != nil {
		return nil, err
	}

	hystrix.ConfigureCommand(opts.ModelGRPCHystrixCommandName, hystrix.CommandConfig{
		Timeout:                durationToInt(opts.ModelTimeout, time.Millisecond),
		MaxConcurrentRequests:  opts.ModelHystrixMaxConcurrentRequests,
		RequestVolumeThreshold: opts.ModelHystrixRequestVolumeThreshold,
		SleepWindow:            opts.ModelHystrixSleepWindowMs,
		ErrorPercentThreshold:  opts.ModelHystrixErrorPercentageThreshold,
	})

	modelClient := upiV1.NewUniversalPredictionServiceClient(conn)
	svr := &UPIServer{
		opts:        opts,
		modelClient: modelClient,
		conn:        conn,
	}

	if handler != nil {
		svr.ContextModifier = handler.EmbedEnvironment
		svr.PreprocessHandler = handler.Preprocess
		svr.PostprocessHandler = handler.Postprocess
	}

	return svr, nil
}

func durationToInt(duration, unit time.Duration) int {
	durationAsNumber := duration / unit

	if int64(durationAsNumber) > int64(maxInt) {
		// Returning max possible value seems like best possible solution here
		// the alternative is to panic as there is no way of returning an error
		// without changing the NewClient API
		return maxInt
	}
	return int(durationAsNumber)
}

func (us *UPIServer) RunServer() error {
	defer us.conn.Close()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", us.opts.GRPCPort))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	upiV1.RegisterUniversalPredictionServiceServer(grpcServer, us)
	grpcServer.Serve(lis)
	return nil
}

func (us *UPIServer) PredictValues(ctx context.Context, request *upiV1.PredictValuesRequest) (*upiV1.PredictValuesResponse, error) {
	if us.ContextModifier != nil {
		ctx = us.ContextModifier(ctx)
	}
	var preprocessRequestBody types.Payload
	preprocessRequestBody = (*types.UPIPredictionRequest)(request)
	if us.PreprocessHandler != nil {
		preprocessOutput, err := us.PreprocessHandler(ctx, preprocessRequestBody, nil)
		if err != nil {
			return nil, err
		}
		preprocessRequestBody = preprocessOutput
	}

	modelRequestPayload, validType := preprocessRequestBody.(*types.UPIPredictionRequest)
	if !validType {
		return nil, fmt.Errorf("unknown type for preprocess output %T", modelRequestPayload)
	}

	var modelResponse *upiV1.PredictValuesResponse
	err := hystrix.Do(us.opts.ModelGRPCHystrixCommandName, func() error {
		response, err := us.modelClient.PredictValues(ctx, (*upiV1.PredictValuesRequest)(modelRequestPayload))
		if err != nil {
			return err
		}
		modelResponse = response
		return nil
	}, func(err error) error {
		return err
	})
	if err != nil {
		return nil, err
	}

	if us.PostprocessHandler == nil {
		return modelResponse, nil
	}

	var postprocessRequestBody types.Payload
	postprocessRequestBody = (*types.UPIPredictionResponse)(modelResponse)
	result, err := us.PostprocessHandler(ctx, postprocessRequestBody, nil)
	if err != nil {
		return nil, err
	}
	if result.IsNil() {
		return nil, nil
	}

	transformerOutput, validType := result.(*types.UPIPredictionResponse)
	if !validType {
		return nil, fmt.Errorf("unknown type for postprocess output %T", result)
	}
	return (*upiV1.PredictValuesResponse)(transformerOutput), nil
}
