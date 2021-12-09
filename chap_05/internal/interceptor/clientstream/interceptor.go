package interceptor

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type wrappedStream struct {
	grpc.ClientStream
}

// NewWrappedStream create wrapped Client Stream
func NewWrappedStream(stream grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{ClientStream: stream}
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	logrus.Infof("==== client stream interceptor wrapped Recv Message Type: %T====", m)
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	logrus.Infof("==== client stream interceptor wrapped Send Message Type: %T====", m)
	return w.ClientStream.SendMsg(m)
}

// ClientStreamOrderManagementInterceptor client stream interceptor
// https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md#stream-interceptor
func ClientStreamOrderManagementInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// 전처리
	logrus.WithContext(ctx).Infof("==== client stream interceptor method: %s ====", method)
	stream, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed get client stream")
		return nil, err
	}

	// 후처리
	return NewWrappedStream(stream), nil
}
