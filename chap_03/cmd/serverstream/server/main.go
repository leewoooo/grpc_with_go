package main

import (
	"fmt"
	pb "grpc_with_go/chap_03/proto/serverstream"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	port = "9000"
)

// OrderManagement implement unray proto service interface
type OrderManagement struct {
	database map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

// NewOrderManagement create OrderManagement instance
func NewOrderManagement(database map[string]*pb.Order) pb.OrderManagementServer {
	return &OrderManagement{database: database}
}

// SearchOrder search order with item and response with stream
func (o *OrderManagement) SearchOrder(searchQuery *wrapperspb.StringValue, stream pb.OrderManagement_SearchOrderServer) error {
	for _, order := range o.database {
		for _, v := range order.Items {
			if strings.Contains(v, searchQuery.GetValue()) {
				logrus.Infof("Try gRPC server send order contain :%v", v)
				err := stream.Send(order)
				if err != nil {
					logrus.WithError(err).Errorf("failed send order: %+v", order)
					return grpc.Errorf(codes.Internal, "failed send order")
				}
			}
		}
	}
	return nil
}

func createDatabasseWithMockData() map[string]*pb.Order {
	database := make(map[string]*pb.Order)

	for _, v := range []*pb.Order{
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
	} {
		database[v.Id] = v
	}

	return database
}

func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).
			Fatal("failed create listener instance")
	}

	srv := grpc.NewServer()

	database := createDatabasseWithMockData()
	orderManagement := NewOrderManagement(database)

	pb.RegisterOrderManagementServer(srv, orderManagement)

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}
