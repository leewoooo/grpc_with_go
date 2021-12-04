package main

import (
	"context"
	"fmt"
	pb "grpc_with_go/chap_03/proto/unray"
	"net"

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
	pb.UnimplementedOrderMangementServer
}

// NewOrderMangementClient create new OrderMangement instance
func NewOrderMangementClient(database map[string]*pb.Order) *OrderManagement {
	return &OrderManagement{database: database}
}

// GetOrder get Order
func (o *OrderManagement) GetOrder(ctx context.Context, ID *wrapperspb.StringValue) (*pb.Order, error) {
	order, ok := o.database[ID.GetValue()]

	if !ok {
		logrus.WithContext(ctx).Warnf("order not exist with id:%s", ID.GetValue())
		return nil, grpc.Errorf(codes.NotFound, "order not exist")
	}

	return order, nil
}

func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).
			Fatal("failed create listener instance")
	}

	database := map[string]*pb.Order{
		"1": {
			Id:          "1",
			Items:       []string{"foo", "bar"},
			Description: "mock Data",
			Price:       99.99,
			Destination: "unray grpc mock data",
		},
	}

	srv := grpc.NewServer()
	orderMangement := NewOrderMangementClient(database)

	pb.RegisterOrderMangementServer(srv, orderMangement)

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}
