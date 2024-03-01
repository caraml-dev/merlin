package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	hystrixGo "github.com/afex/hystrix-go/hystrix"
	hystrixpkg "github.com/caraml-dev/merlin/pkg/hystrix"
	"github.com/caraml-dev/merlin/pkg/transformer/server/config"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/go-coldbrew/grpcpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type upiPredictorClient interface {
	predict(ctx context.Context, payload *upiv1.PredictValuesRequest) (*upiv1.PredictValuesResponse, error)
	close() error
}

func newUPIPredictorClient(opts *config.Options) (upiPredictorClient, error) {
	if opts.PredictorUPIHTTPEnabled {
		return newHTTPClient(opts)
	}
	return newGRPCClient(opts)
}

type grpcClient struct {
	conn      grpcpool.ConnPool
	upiClient upiv1.UniversalPredictionServiceClient
	opts      *config.Options
}

func newGRPCClient(opts *config.Options) (*grpcClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if opts.ModelGRPCKeepAliveEnabled {
		dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Timeout:             opts.ModelGRPCKeepAliveTime,
			Time:                opts.ModelGRPCKeepAliveTimeout,
			PermitWithoutStream: true,
		}))
	}

	connPool, err := grpcpool.DialContext(context.Background(), opts.ModelPredictURL, uint(opts.ModelServerConnCount), dialOpts...)
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

	modelClient := upiv1.NewUniversalPredictionServiceClient(connPool)
	return &grpcClient{
		conn:      connPool,
		upiClient: modelClient,
		opts:      opts,
	}, nil
}

func (cli *grpcClient) predict(ctx context.Context, payload *upiv1.PredictValuesRequest) (*upiv1.PredictValuesResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	var modelResponse *upiv1.PredictValuesResponse
	err := hystrix.Do(cli.opts.ModelGRPCHystrixCommandName, func() error {
		response, err := cli.upiClient.PredictValues(ctx, (*upiv1.PredictValuesRequest)(payload), grpc.WaitForReady(true))
		if err != nil {
			return err
		}
		modelResponse = response
		return nil
	}, nil)

	if err != nil {
		return nil, err
	}
	return modelResponse, nil
}

func (cli *grpcClient) close() error {
	return cli.conn.Close()
}

func newHTTPClient(opts *config.Options) (*httpClient, error) {
	hystrixConfig := hystrixGo.CommandConfig{
		Timeout:                int(opts.ModelTimeout / time.Millisecond),
		MaxConcurrentRequests:  opts.ModelHystrixMaxConcurrentRequests,
		RequestVolumeThreshold: opts.ModelHystrixRequestVolumeThreshold,
		SleepWindow:            opts.ModelHystrixSleepWindowMs,
		ErrorPercentThreshold:  opts.ModelHystrixErrorPercentageThreshold,
	}
	cl := &http.Client{
		Timeout: opts.ModelTimeout,
	}
	client := hystrixpkg.NewClient(cl, &hystrixConfig, opts.ModelHTTPHystrixCommandName)
	return &httpClient{client: client, opts: opts}, nil
}

type httpClient struct {
	client hystrixHttpClient
	opts   *config.Options
}

type hystrixHttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

func (cli *httpClient) predict(ctx context.Context, payload *upiv1.PredictValuesRequest) (*upiv1.PredictValuesResponse, error) {
	payloadBytes, err := protojson.Marshal(payload)
	if err != nil {
		return nil, err
	}

	predictURL := getUrl(fmt.Sprintf("%s/v1/models/%s:predict", cli.opts.ModelPredictURL, cli.opts.ModelFullName))
	req, err := http.NewRequestWithContext(ctx, "POST", predictURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Length", fmt.Sprint(len(payloadBytes)))

	res, err := cli.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint: errcheck

	modelResponseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var upiResponse upiv1.PredictValuesResponse
	if err := protojson.Unmarshal(modelResponseBody, &upiResponse); err != nil {
		return nil, err
	}
	return &upiResponse, nil
}

func (cli *httpClient) close() error {
	return nil
}
