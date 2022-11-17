package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mErrors "github.com/gojek/merlin/pkg/errors"
	hystrixpkg "github.com/gojek/merlin/pkg/hystrix"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server/config"
	"github.com/gojek/merlin/pkg/transformer/server/grpc/interceptors"
	"github.com/gojek/merlin/pkg/transformer/server/instrumentation"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"

	"github.com/afex/hystrix-go/hystrix"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
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

	opts                  *config.Options
	modelClient           upiv1.UniversalPredictionServiceClient
	conn                  *grpc.ClientConn
	instrumentationRouter *mux.Router
	logger                *zap.Logger

	// ContextModifier function to modify or store value in a context
	ContextModifier func(ctx context.Context) context.Context
	// PreprocessHandler function to run all preprocess operation
	// request parameter for this function must be in types.UPIPredictionRequest type
	// output payload  of this function must be in types.UPIPredictionRequest type
	PreprocessHandler func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error)
	// PostprocessHandler function to run all postprocess operation
	// response parameter for this function must be in types.UPIPredictionResponse type
	// output payload of this function must be in types.UPIPredictionResponse type
	PostprocessHandler func(ctx context.Context, response types.Payload, responseHeaders map[string]string) (types.Payload, error)
}

// NewUPIServer creates GRPC server that implement UPI Service
func NewUPIServer(opts *config.Options, handler *pipeline.Handler, instrumentationRouter *mux.Router, logger *zap.Logger) (*UPIServer, error) {
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
		opts:                  opts,
		modelClient:           modelClient,
		conn:                  conn,
		instrumentationRouter: instrumentationRouter,
		logger:                logger,
	}

	if handler != nil {
		svr.ContextModifier = handler.EmbedEnvironment
		svr.PreprocessHandler = handler.Preprocess
		svr.PostprocessHandler = handler.Postprocess
	}

	return svr, nil
}

// Run running GRPC Server
func (us *UPIServer) Run() {
	// bind to all interfaces at port us.opts.GRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", us.opts.GRPCPort))
	if err != nil {
		us.logger.Error(fmt.Sprintf("failed to listen the port %s", us.opts.GRPCPort))
		return
	}

	m := cmux.New(lis)
	// cmux.HTTP2MatchHeaderFieldSendSettings ensures we can handle any gRPC client.
	grpcLis := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpLis := m.Match(cmux.HTTP1Fast())

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptors.PanicRecoveryInterceptor()),
	}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	upiv1.RegisterUniversalPredictionServiceServer(grpcServer, us)

	// add health check service
	healthChecker := newHealthChecker()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthChecker)

	stopCh := setupSignalHandler()
	errCh := make(chan error, 1)
	go func() {
		us.logger.Info("starting grpc server")
		if err := grpcServer.Serve(grpcLis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- errors.Wrap(err, "GRPC server failed")
		}
	}()

	httpServer := &http.Server{Handler: us.instrumentationRouter}
	go func() {
		us.logger.Info("starting http server")
		if err := httpServer.Serve(httpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- errors.Wrapf(err, "instrumentation server failed")
		}
	}()

	go func() {
		us.logger.Info(fmt.Sprintf("serving at port: %s", us.opts.GRPCPort))
		if err := m.Serve(); err != nil && !errors.Is(err, cmux.ErrListenerClosed) {
			errCh <- errors.Wrapf(err, "cmux server failed")
		}
	}()

	select {
	case <-stopCh:
		us.logger.Info("got signal to stop server")
	case err := <-errCh:
		us.logger.Error(fmt.Sprintf("failed to run server %v", err))
	}

	us.logger.Info("shutting down standard transformer")
	if err := us.conn.Close(); err != nil {
		us.logger.Error(fmt.Sprintf("failed to close connection %v", err))
	}
	us.logger.Info("closed connection to model prediction server")

	grpcServer.GracefulStop()
	us.logger.Info("stopped grpc server")
	if err = httpServer.Shutdown(context.Background()); err != nil {
		us.logger.Warn("failed shutting down http server")
		return
	}

	us.logger.Info("stopped http server")
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
		return nil, status.Errorf(getGRPCCode(err), "preprocess err: %v", err)
	}

	modelResponse, err := us.predict(ctx, preprocessOutput)
	if err != nil {
		us.logger.Error("predict error", zap.Error(err))
		return nil, status.Errorf(getGRPCCode(err), "predict err: %v", err)
	}

	postprocessOutput, err := us.postprocess(ctx, modelResponse, meta)
	if err != nil {
		us.logger.Error("postprocess error", zap.Error(err))
		return nil, status.Errorf(getGRPCCode(err), "postprocess err: %v", err)
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
		if errors.Is(err, hystrix.ErrTimeout) {
			return nil, mErrors.NewDeadlineExceededError(err.Error())
		}
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

func getGRPCCode(err error) codes.Code {
	if statusErr, valid := status.FromError(err); valid {
		return statusErr.Code()
	}

	if errors.Is(err, mErrors.InvalidInputError) {
		return codes.InvalidArgument
	} else if errors.Is(err, mErrors.DeadlineExceededError) {
		return codes.DeadlineExceeded
	}
	return codes.Internal
}
