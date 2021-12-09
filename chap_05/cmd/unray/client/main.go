package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	interceptor "grpc_with_go/chap_05/internal/interceptor/unray"
	pb "grpc_with_go/chap_05/proto/unray"
)

const (
	serverURL = "localhost:9000"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.Dial(
		serverURL,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(interceptor.UnrayOrderManagementClientInterceptor),
	)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
	}

	client := pb.NewOrderManagementClient(conn)

	order := &pb.Order{
		Items:       []string{"mac pro"},
		Description: "apple",
		Price:       199.99,
		Destination: "korea seoul",
	}

	ID, err := client.AddOrder(ctx, order)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed AddOrder: %+v", order)
		return
	}

	selectedOrder, err := client.GetOrder(ctx, ID)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed GetOrder: %s", ID.GetId())
		return
	}

	logrus.WithContext(ctx).Infof("selected Order: %+v", selectedOrder)
}
