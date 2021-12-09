package main

import (
	"context"
	interceptor "grpc_with_go/chap_05/internal/interceptor/clientstream"
	pb "grpc_with_go/chap_05/proto/clientstream"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	serverURL = "localhost:9000"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.Dial(serverURL, grpc.WithInsecure(), grpc.WithStreamInterceptor(interceptor.ClientStreamOrderManagementInterceptor))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
	}
	client := pb.NewOrderManagementClient(conn)

	stream, err := client.UpdateOrders(ctx)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get client stream")
	}

	target1 := &pb.Order{Id: "102", Items: []string{"Google Pixel 10", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	err = stream.Send(target1)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatalf("failed send Order: %+v", target1)
	}

	target2 := &pb.Order{Id: "103", Items: []string{"Apple Watch 7"}, Destination: "San Jose, CA", Price: 400.00}
	err = stream.Send(target2)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatalf("failed send Order: %+v", target2)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatalf("failed recv server resp")
	}

	logrus.WithContext(ctx).Infof("resp: %s", resp.GetValue())
}
