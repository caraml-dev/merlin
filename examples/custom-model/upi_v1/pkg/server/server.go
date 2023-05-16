package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"custom-model-upi/pkg/interceptors"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

type UPIServer struct {
	upiv1.UnimplementedUniversalPredictionServiceServer
	logger        *zap.Logger
	modelFullName string
}

func RunUPIServer(cfg *Config, instrumentationRouter *mux.Router, logger *zap.Logger) {
	srv := &UPIServer{
		logger:        logger,
		modelFullName: cfg.CaraMLConfig.ModelFullName,
	}
	// bind to all interfaces at port cfg.CaraMLConfig.GRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.CaraMLConfig.GRPCPort))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen the port %d", cfg.CaraMLConfig.GRPCPort))
	}

	m := cmux.New(lis)
	// cmux.HTTP2MatchHeaderFieldSendSettings ensures we can handle any gRPC client.
	grpcLis := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpLis := m.Match(cmux.HTTP1Fast())
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptors.PanicRecoveryInterceptor()),
	}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	upiv1.RegisterUniversalPredictionServiceServer(grpcServer, srv)

	stopCh := setupSignalHandler()
	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting grpc server")
		if err := grpcServer.Serve(grpcLis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- errors.Wrap(err, "GRPC server failed")
		}
	}()

	httpServer := &http.Server{Handler: instrumentationRouter}
	go func() {
		logger.Info("starting http server")
		if err := httpServer.Serve(httpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- errors.Wrapf(err, "instrumentation server failed")
		}
	}()

	go func() {
		logger.Info(fmt.Sprintf("serving at port: %d", cfg.CaraMLConfig.GRPCPort))
		if err := m.Serve(); err != nil && !errors.Is(err, cmux.ErrListenerClosed) {
			errCh <- errors.Wrapf(err, "cmux server failed")
		}
	}()

	select {
	case <-stopCh:
		logger.Info("got signal to stop server")
	case err := <-errCh:
		logger.Error(fmt.Sprintf("failed to run server %v", err))
	}

	grpcServer.GracefulStop()
	logger.Info("stopped grpc server")
	if err = httpServer.Shutdown(context.Background()); err != nil {
		logger.Warn("failed shutting down http server")
		return
	}

	logger.Info("stopped http server")
}

var (
	shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

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
// for the sake of simplicity this method only pass value from request payload
func (srv *UPIServer) PredictValues(ctx context.Context, request *upiv1.PredictValuesRequest) (response *upiv1.PredictValuesResponse, grpcErr error) {
	startTime := time.Now()
	defer func() {
		statusCode := getGRPCCode(grpcErr)
		requestCounter.WithLabelValues(srv.modelFullName, statusCode.String())
		duration := time.Since(startTime)
		requestLatencyBucket.WithLabelValues(srv.modelFullName, statusCode.String()).Observe(duration.Seconds())

	}()
	return &upiv1.PredictValuesResponse{
		PredictionResultTable: request.PredictionTable,
		TargetName:            request.TargetName,
		PredictionContext:     request.PredictionContext,
	}, nil
}

func getGRPCCode(err error) codes.Code {
	if err != nil {
		return codes.OK
	}
	if statusErr, valid := status.FromError(err); valid {
		return statusErr.Code()
	}
	return codes.Internal
}
