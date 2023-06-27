package interceptors

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TracingPropagatorInterceptor interceptor to use load context that has parent tracing
func TracingPropagatorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md := getMetadata(ctx)
		headerCarier := propagation.MapCarrier(md)
		ctx = otel.GetTextMapPropagator().Extract(ctx, headerCarier)

		resp, err = handler(ctx, req)
		return resp, err
	}
}

func getMetadata(ctx context.Context) map[string]string {
	meta, _ := metadata.FromIncomingContext(ctx)
	resultHeaders := make(map[string]string, len(meta))
	for k, v := range meta {
		resultHeaders[k] = strings.Join(v, ",")
	}
	return resultHeaders
}
