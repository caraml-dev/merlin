package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	hystrixpkg "github.com/gojek/merlin/pkg/hystrix"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server/config"
	"github.com/gojek/merlin/pkg/transformer/server/instrumentation"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/afex/hystrix-go/hystrix"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

// UPIServer serves GRPC request that implement UPI Service
type UPIServer struct {
	upiv1.UnimplementedUniversalPredictionServiceServer

	opts        *config.Options
	modelClient upiv1.UniversalPredictionServiceClient
	conn        *grpc.ClientConn
	logger      *zap.Logger

	ContextModifier    func(ctx context.Context) context.Context
	PreprocessHandler  func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error)
	PostprocessHandler func(ctx context.Context, response types.Payload, responseHeaders map[string]string) (types.Payload, error)
}

// NewUPIServer creates GRPC server that implement UPI Service
func NewUPIServer(opts *config.Options, handler *pipeline.Handler, logger *zap.Logger) (*UPIServer, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(opts.ModelPredictURL, dialOpts...)
	if err != nil {
		return nil, err
	}

	hystrix.ConfigureCommand(opts.ModelGRPCHystrixCommandName, hystrix.CommandConfig{
		Timeout:                hystrixpkg.DurationToInt(opts.ModelTimeout, time.Millisecond),
		MaxConcurrentRequests:  opts.ModelHystrixMaxConcurrentRequests,
		RequestVolumeThreshold: opts.ModelHystrixRequestVolumeThreshold,
		SleepWindow:            opts.ModelHystrixSleepWindowMs,
		ErrorPercentThreshold:  opts.ModelHystrixErrorPercentageThreshold,
	})

	modelClient := upiv1.NewUniversalPredictionServiceClient(conn)
	svr := &UPIServer{
		opts:        opts,
		modelClient: modelClient,
		conn:        conn,
		logger:      logger,
	}

	if handler != nil {
		svr.ContextModifier = handler.EmbedEnvironment
		svr.PreprocessHandler = handler.Preprocess
		svr.PostprocessHandler = handler.Postprocess
	}

	return svr, nil
}

// RunServer running GRPC Server
func (us *UPIServer) RunServer() {
	// defer us.conn.Close()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", us.opts.GRPCPort))
	if err != nil {
		us.logger.Error(fmt.Sprintf("failed to listen the port %s", us.opts.GRPCPort))
		return
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(PanicRecoveryInterceptor()),
	}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	upiv1.RegisterUniversalPredictionServiceServer(grpcServer, us)

	stopCh := setupSignalHandler()
	errCh := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			errCh <- errors.Wrap(err, "server failed")
		}
	}()

	select {
	case <-stopCh:
		us.logger.Info("got signal to stop server")
	case err := <-errCh:
		us.logger.Error(fmt.Sprintf("failed to run GRPC server %v", err))
	}

	if err := us.conn.Close(); err != nil {
		us.logger.Error(fmt.Sprintf("failed to close connection %v", err))
	}
	us.logger.Info("closed connection to model prediction server")
	grpcServer.GracefulStop()
}

func setupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
	}()

	return stop
}

// PredictValues method to performing model prediction
// it is including preprocessing - model infer - postprocessing
func (us *UPIServer) PredictValues(ctx context.Context, request *upiv1.PredictValuesRequest) (*upiv1.PredictValuesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PredictHandler")
	defer span.Finish()

	if us.ContextModifier != nil {
		ctx = us.ContextModifier(ctx)
	}
	meta := getMetadata(ctx)
	preprocessOutput, err := us.preprocess(ctx, request, meta)
	if err != nil {
		us.logger.Error("preprocess error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "preprocess err: %v", err)
	}

	modelResponse, err := us.predict(ctx, preprocessOutput)
	if err != nil {
		us.logger.Error("predict error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "predict err: %v", err)
	}

	postprocessOutput, err := us.postprocess(ctx, modelResponse, meta)
	if err != nil {
		us.logger.Error("postprocess error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "postprocess err: %v", err)
	}

	return postprocessOutput, nil
}

func (us *UPIServer) preprocess(ctx context.Context, request *upiv1.PredictValuesRequest, meta map[string]string) (*upiv1.PredictValuesRequest, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, string(types.Preprocess))
	defer span.Finish()

	if us.PreprocessHandler == nil {
		return request, nil
	}

	startTime := time.Now()
	output, err := us.PreprocessHandler(ctx, (*types.UPIPredictionRequest)(request), meta)
	durationMs := time.Since(startTime).Milliseconds()
	if err != nil {
		instrumentation.RecordPreprocessLatency(false, float64(durationMs))
		return nil, err
	}

	out, validType := output.(*types.UPIPredictionRequest)
	if !validType {
		instrumentation.RecordPreprocessLatency(false, float64(durationMs))
		return nil, fmt.Errorf("unexpected type for preprocess output %T", output)
	}
	instrumentation.RecordPreprocessLatency(true, float64(durationMs))
	return (*upiv1.PredictValuesRequest)(out), nil
}

func (us *UPIServer) postprocess(ctx context.Context, response *upiv1.PredictValuesResponse, meta map[string]string) (*upiv1.PredictValuesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, string(types.Postprocess))
	defer span.Finish()

	if us.PostprocessHandler == nil {
		return response, nil
	}

	startTime := time.Now()
	output, err := us.PostprocessHandler(ctx, (*types.UPIPredictionResponse)(response), meta)
	durationMs := time.Since(startTime).Milliseconds()
	if err != nil {
		instrumentation.RecordPostprocessLatency(false, float64(durationMs))
		return nil, err
	}
	out, validType := output.(*types.UPIPredictionResponse)
	if !validType {
		instrumentation.RecordPostprocessLatency(false, float64(durationMs))
		return nil, fmt.Errorf("unexpected type for postprocess output %T", output)
	}
	instrumentation.RecordPostprocessLatency(true, float64(durationMs))
	return (*upiv1.PredictValuesResponse)(out), nil
}

func (us *UPIServer) predict(ctx context.Context, payload *upiv1.PredictValuesRequest) (*upiv1.PredictValuesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "predict")
	defer span.Finish()

	predictStartTime := time.Now()
	var modelResponse *upiv1.PredictValuesResponse
	err := hystrix.Do(us.opts.ModelGRPCHystrixCommandName, func() error {
		response, err := us.modelClient.PredictValues(ctx, (*upiv1.PredictValuesRequest)(payload))
		if err != nil {
			return err
		}
		modelResponse = response
		return nil
	}, nil)

	predictionDurationMs := time.Since(predictStartTime).Milliseconds()
	if err != nil {
		instrumentation.RecordPredictionLatency(false, float64(predictionDurationMs))
		return nil, err
	}
	instrumentation.RecordPredictionLatency(true, float64(predictionDurationMs))
	return modelResponse, nil
}

func getMetadata(ctx context.Context) map[string]string {
	meta, _ := metadata.FromIncomingContext(ctx)
	resultHeaders := make(map[string]string, len(meta))
	for k, v := range meta {
		resultHeaders[k] = strings.Join(v, ",")
	}
	return resultHeaders
}
