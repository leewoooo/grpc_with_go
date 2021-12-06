package main

import (
	"context"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "grpc_with_go/chap_03/proto/twoway"
)

const (
	serverURL = "localhost:9000"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Errorf("failed get gRPC connection with serverURL: %s", serverURL)
		return
	}

	client := pb.NewOrderManagementClient(conn)
	stream, err := client.ProcessOrders(ctx)
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Errorf("failed get gRPC stream instance with value: %s", "mac")
		return
	}

	// server batch size = 3
	if err := stream.Send(wrapperspb.String("1")); err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed Send where stream with value: %s", "1")
		return
	}

	if err := stream.Send(wrapperspb.String("2")); err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed Send where stream with value: %s", "2")
		return
	}

	if err := stream.Send(wrapperspb.String("3")); err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed Send where stream with value: %s", "3")
		return
	}

	// 3개 들어가면 응답이 올 것이다. gorutine을 이용하여 응답과 요청을 동시에 처리
	channel := make(chan struct{})
	go asyncClientBidrectionalRPC(stream, channel)
	time.Sleep(time.Second * 5)

	if err := stream.Send(wrapperspb.String("4")); err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed Send where stream with value: %s", "4")
		return
	}

	if err := stream.CloseSend(); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed CloseSend")
		return
	}

	<-channel
}

func asyncClientBidrectionalRPC(stream pb.OrderManagement_ProcessOrdersClient, channel chan struct{}) {
	for {
		comb, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logrus.WithContext(stream.Context()).
					Info("server stream End")
				break
			}
			logrus.WithContext(stream.Context()).WithError(err).Error("failed Recv where stream")
			return
		}

		logrus.WithContext(stream.Context()).Infof("Combined shipment: %v", comb.OrdersList)
	}

	close(channel)
}
