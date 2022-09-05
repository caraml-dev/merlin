package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

func PanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func(ctx context.Context) {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("panic: %s", r)
				}
			}
		}(ctx)

		resp, err = handler(ctx, req)
		return resp, err
	}
}
