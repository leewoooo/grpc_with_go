package main

import (
	"fmt"
	interceptor "grpc_with_go/chap_05/internal/interceptor/serverstream"
	pb "grpc_with_go/chap_05/proto/serverstream"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	port     = "9000"
	database = make(map[string]*pb.Order)
)

func init() {
	database["102"] = &pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	database["103"] = &pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	database["104"] = &pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	database["105"] = &pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	database["106"] = &pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}

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

func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).
			Fatal("failed create listener instance")
	}

	srv := grpc.NewServer(grpc.StreamInterceptor(
		interceptor.ServerStreamOrderManagementStream,
	))
	orderManagement := NewOrderManagement(database)

	pb.RegisterOrderManagementServer(srv, orderManagement)

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}
