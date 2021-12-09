package main

import (
	"context"
	"io"

	pb "grpc_with_go/chap_05/proto/serverstream"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	serverURL = "localhost:9000"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
	}

	client := pb.NewOrderManagementClient(conn)

	stream, err := client.SearchOrder(ctx, wrapperspb.String("Amazon Echo"))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatalf("failed get stream with item:%s", "Amazon Echo")
	}

	for {
		order, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logrus.WithContext(stream.Context()).Info("server stream End")
				break
			}

			logrus.WithContext(stream.Context()).WithError(err).Error("failed get Order Where server stream")
			return
		}

		logrus.WithContext(stream.Context()).Infof("selected Order: %+v", order)
	}
}
