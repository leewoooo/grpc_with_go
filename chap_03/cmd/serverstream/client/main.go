package main

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "grpc_with_go/chap_03/proto/serverstream"
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

	stream, err := client.SearchOrder(ctx, wrapperspb.String("mac"))
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Errorf("failed get gRPC stream instance with value: %s", "mac")
		return
	}

	for {
		order, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logrus.WithContext(ctx).
					WithError(err).
					Info("server stream data end")
				break
			}
			logrus.WithContext(ctx).
				WithError(err).
				Error("failed get order where stream instance")
			return
		}

		logrus.WithContext(ctx).Infof("get order: %+v", order)
	}

}
