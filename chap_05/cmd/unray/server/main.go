package main

import (
	"context"
	"errors"
	"fmt"
	unray "grpc_with_go/chap_05/internal/interceptor/unray"
	pb "grpc_with_go/chap_05/proto/unray"
	"net"

	epb "google.golang.org/genproto/googleapis/rpc/errdetails"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// OrderMangement implement protobuf service interface
type OrderMangement struct {
	db map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

// NewOrderManagement create pb.OrderManagementServer instance
func NewOrderManagement(db map[string]*pb.Order) pb.OrderManagementServer {
	return &OrderMangement{db: db}
}

// AddOrder implement OrderManagementServer AddOrder
func (o *OrderMangement) AddOrder(ctx context.Context, order *pb.Order) (*pb.OrderId, error) {
	// set ID
	ID, _ := gonanoid.New()
	order.Id = ID

	// save
	o.db[ID] = order

	// return
	return &pb.OrderId{Id: ID}, nil
}

// GetOrder implement OrderManagementServer GetOrder
func (o *OrderMangement) GetOrder(ctx context.Context, ID *pb.OrderId) (*pb.Order, error) {
	if ctx.Err() == context.DeadlineExceeded {
		logrus.WithContext(ctx).Info("client Request DEADLINE_EXCEEDE")
		return nil, errors.New("client Request DEADLINE_EXCEEDE")
	}

	// GetOrder id가 1보다 작은 경우
	if ID.GetId() < "1" {
		logrus.WithContext(ctx).Error("Order ID is Invalid Received OrderID: %s", ID.GetId())

		errStatus := status.New(codes.InvalidArgument, "Invalid information received")
		ds, err := errStatus.WithDetails(
			&epb.BadRequest_FieldViolation{
				Field: "ID",
				Description: fmt.Sprintf(
					"Order ID received is not valid: %s",
					ID.GetId(),
				),
			},
		)
		if err != nil {
			return nil, errStatus.Err()
		}

		return nil, ds.Err()
	}

	order, ok := o.db[ID.GetId()]

	if !ok {
		return nil, grpc.Errorf(codes.NotFound, "Not Exist")
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

	srv := grpc.NewServer(grpc.UnaryInterceptor(
		unray.UnrayOrderManagementServerInterceptor,
	))
	orderManagement := NewOrderManagement(database)

	pb.RegisterOrderManagementServer(srv, orderManagement)

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}
