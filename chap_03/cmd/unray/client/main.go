package main

import (
	"context"

	pb "grpc_with_go/chap_03/proto/unray"

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

	client := pb.NewOrderMangementClient(conn)

	order, err := client.GetOrder(ctx, wrapperspb.String("1"))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatalf("failed get order with ID:%s", "1")
	}

	logrus.WithContext(ctx).Infof("%+v", order)
}
