package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "grpc_with_go/chap_03/proto/clientstream"
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

	stream, err := client.UpdateOrders(ctx)
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Error("failed get gRPC client stream")
		return
	}

	for _, v := range generateMockData() {
		err := stream.Send(v)
		if err != nil {
			logrus.WithContext(ctx).
				WithError(err).
				Errorf("failed sned Order with: %+v", v)
			return
		}

		logrus.WithContext(ctx).Infof("success send order with ID :%s", v.GetId())
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed get server response")
		return
	}

	logrus.WithContext(ctx).Infof("response: %s", response.GetValue())
}

func generateMockData() []*pb.Order {
	return []*pb.Order{
		{
			Id:          "1",
			Items:       []string{"iphone", "mac"},
			Description: "mock Data 1",
			Price:       99.99,
			Destination: "korea",
		},
		{
			Id:          "2",
			Items:       []string{"mac", "airpod"},
			Description: "mock Data 2",
			Price:       129.99,
			Destination: "korea",
		},
		{
			Id:          "3",
			Items:       []string{"iphone"},
			Description: "mock Data 3",
			Price:       79.99,
			Destination: "korea",
		},
	}
}
