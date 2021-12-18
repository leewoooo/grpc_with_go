package main

import (
	"context"
	"fmt"
	pb "grpc_with_go/chap_05/proto/metadata"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	port     = "50051"
	database = make(map[string]*pb.Order)
)

func init() {
	database["102"] = &pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	database["103"] = &pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	database["104"] = &pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	database["105"] = &pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	database["106"] = &pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}

// OrderMangement implement protobuf service interface
type OrderMangement struct {
	db map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

// NewOrderManagement create pb.OrderManagementServer instance
func NewOrderManagement(db map[string]*pb.Order) pb.OrderManagementServer {
	return &OrderMangement{db: db}
}

// GetOrder implement OrderManagementServer GetOrder
func (o *OrderMangement) GetOrder(ctx context.Context, reqID *wrapperspb.StringValue) (*pb.Order, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.FailedPrecondition, "metadata should be exist")
	}

	// logging metadata
	for k, v := range md {
		logrus.WithContext(ctx).Infof("key: %s, val:%s", k, v)
	}

	// authorization
	datas := md.Get("Authorization")
	if len(datas) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "Authorization metadata should be exist")
	}
	if token := datas[0]; token != "foobar" {
		return nil, status.Error(codes.FailedPrecondition, "token failed validate")
	}

	// setting header, trailer
	createdMd := metadata.Pairs("header-key", "header-val")
	grpc.SendHeader(ctx, createdMd)

	createdTrailer := metadata.Pairs("trailer-key", "trailer-val")
	grpc.SetTrailer(ctx, createdTrailer)
	return o.db[reqID.GetValue()], nil
}

func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).
			Fatal("failed create listener instance")
	}

	srv := grpc.NewServer()
	orderManagement := NewOrderManagement(database)

	pb.RegisterOrderManagementServer(srv, orderManagement)

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}
