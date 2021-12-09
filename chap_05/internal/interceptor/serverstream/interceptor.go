package interceptor

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// WrappedStream wrapping grpc server stream
type wrappedStream struct {
	grpc.ServerStream
}

// NewWrappedStream create wrappedStream instance
func NewWrappedStream(stream grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{stream}
}

// RecvMsg inst
func (w *wrappedStream) RecvMsg(m interface{}) error {
	logrus.Infof("==== server stream interceptor wrapper Receive Message Type:%T====", m)
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	logrus.Infof("==== server stream interceptor wrapper Send Message Type:%T====", m)
	return w.ServerStream.SendMsg(m)
}

// ServerStreamOrderManagementInterceptor server stream interceptor
// https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md#stream-interceptor-1
func ServerStreamOrderManagementInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 전처리
	logrus.WithContext(ss.Context()).Infof("==== server stream interceptor method: %s ====", info.FullMethod)

	err := handler(srv, NewWrappedStream(ss))
	// 후처리
	if err != nil {
		logrus.WithContext(ss.Context()).WithError(err).Error("failed RPC")
		return err
	}

	return nil
}
