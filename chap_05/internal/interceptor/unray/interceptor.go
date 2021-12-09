package interceptor

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// UnrayOrderManagementInterceptorfunc unray interceptor
func UnrayOrderManagementInterceptorfunc(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
