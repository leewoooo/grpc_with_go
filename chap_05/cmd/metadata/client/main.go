package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "grpc_with_go/chap_05/proto/metadata"
)

var (
	serverURL = "localhost:50051"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.Dial(
		serverURL,
		grpc.WithInsecure(),
	)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
	}
	orderClient := pb.NewOrderManagementClient(conn)

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"Authorization", "foobar",
		"client-key", "client-val",
		"client-key", "client-val2",
	))

	ctx = metadata.AppendToOutgoingContext(ctx,
		"client-key", "client-val3",
	)

	var header, trailer metadata.MD
	order, err := orderClient.GetOrder(
		ctx,
		wrapperspb.String("102"),
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	if err != nil {
		errStatus := status.Convert(err)
		logrus.WithContext(ctx).WithError(err).Errorf("failed get Order with status: :%s, ID: %s", errStatus.Code(), "102")
	}

	logrus.WithContext(ctx).Infof("selected Order: %+v", order)
}
