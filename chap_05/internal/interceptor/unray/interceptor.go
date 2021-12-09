package interceptor

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// UnrayOrderManagementServerInterceptor unray interceptor
// https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md#unary-interceptor-1
func UnrayOrderManagementServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// 전처리
	logrus.WithContext(ctx).
		Infof("==== [Server Interceptor] method: %s", info.FullMethod)

	// 실행
	resp, err = handler(ctx, req)

	// 후처리
	logrus.WithContext(ctx).
		Infof("Post Proc Message: %+v", resp)

	return resp, err
}

// UnrayOrderManagementClientInterceptor unray client interceptor
// https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md#unary-interceptor
func UnrayOrderManagementClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	logrus.WithContext(ctx).Infof("==== client unray interceptor method: %s req: %T ====", method, req)

	err := invoker(ctx, method, req, reply, cc, opts...)

	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed Reqeust RPC met")
		return err
	}

	return nil
}
