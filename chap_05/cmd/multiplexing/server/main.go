package main

import (
	"context"
	"fmt"
	gretterpb "grpc_with_go/chap_05/proto/gretter"
	orderpb "grpc_with_go/chap_05/proto/unray"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	port     = "50051"
	database = make(map[string]*orderpb.Order)
)

func init() {
	database["102"] = &orderpb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	database["103"] = &orderpb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	database["104"] = &orderpb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	database["105"] = &orderpb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	database["106"] = &orderpb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}

// OrderManagement Service
type OrderManagement struct {
	db map[string]*orderpb.Order
	orderpb.UnimplementedOrderManagementServer
}

// NewOrderManagement create OrderManagementServer instance
func NewOrderManagement(db map[string]*orderpb.Order) orderpb.OrderManagementServer {
	return &OrderManagement{db: db}
}

// AddOrder unimplement
func (o *OrderManagement) AddOrder(ctx context.Context, order *orderpb.Order) (*orderpb.OrderId, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// GetOrder implement GetOrder
func (o *OrderManagement) GetOrder(ctx context.Context, req *orderpb.OrderId) (*orderpb.Order, error) {
	if strings.HasPrefix(req.GetId(), "-") {
		logrus.WithContext(ctx).Error("Order ID is Invalid Received OrderID: %s", req.GetId())

		errStatus := status.New(codes.InvalidArgument, "invalid information received")
		ds, err := errStatus.WithDetails(
			&epb.BadRequest_FieldViolation{
				Field: "ID",
				Description: fmt.Sprintf(
					"Order ID received is not valid: %s",
					req.GetId(),
				),
			},
		)
		if err != nil {
			return nil, errStatus.Err()
		}

		return nil, ds.Err()
	}

	selectedOrder, ok := o.db[req.GetId()]
	if !ok {
		logrus.WithContext(ctx).Error("Order Not Exist with OrderID: %s", req.GetId())

		errStatus := status.New(codes.NotFound, "resources not exist")
		ds, err := errStatus.WithDetails(
			&epb.BadRequest_FieldViolation{
				Field: "ID",
				Description: fmt.Sprintf(
					"Order Not Exist with ID: %s",
					req.GetId(),
				),
			},
		)
		if err != nil {
			return nil, errStatus.Err()
		}
		return nil, ds.Err()
	}

	return selectedOrder, nil
}

// Gretter service
type Gretter struct {
	gretterpb.UnimplementedGreeterServer
}

// NewGretter create Gretter Server instance
func NewGretter() gretterpb.GreeterServer {
	return &Gretter{}
}

// SayHello implement SayHello service
func (g *Gretter) SayHello(ctx context.Context, req *gretterpb.HelloRequest) (*gretterpb.HelloReply, error) {
	return &gretterpb.HelloReply{
		Message: fmt.Sprintf("Hello %s", req.GetName()),
	}, nil
}

func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).
			Fatal("failed create listener instance")
	}

	srv := grpc.NewServer()
	orderpb.RegisterOrderManagementServer(srv, NewOrderManagement(database))
	gretterpb.RegisterGreeterServer(srv, NewGretter())

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}
