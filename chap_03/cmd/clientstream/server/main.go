package main

import (
	"fmt"
	pb "grpc_with_go/chap_03/proto/clientstream"
	"io"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	port = "9000"
)

// OrderManagement implement OrderManagement service interface
type OrderManagement struct {
	database map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

// NewOrderManagement create OrderManagement instance
func NewOrderManagement(database map[string]*pb.Order) pb.OrderManagementServer {
	return &OrderManagement{database: database}
}

// UpdateOrders update Orders
func (o *OrderManagement) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	responseValue := "Updated Order Ids:"

	for {
		order, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logrus.WithError(err).Info("Client stream End")

				err := stream.SendAndClose(wrapperspb.String(responseValue))
				if err != nil {
					logrus.WithError(err).Errorf("failed client stream response with value: %s", responseValue)
					return err
				}

				break
			}

			logrus.WithError(err).Error("failed recv client stream")
			return err
		}

		logrus.Infof("Try Update Order with %+v", order)

		o.database[order.GetId()] = order
		responseValue += fmt.Sprintf("%s ", order.GetId())
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

	srv := grpc.NewServer()

	database := generateMockDatabase()
	orderManagement := NewOrderManagement(database)

	pb.RegisterOrderManagementServer(srv, orderManagement)

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}

func generateMockDatabase() map[string]*pb.Order {
	database := make(map[string]*pb.Order)

	for _, v := range []*pb.Order{
		{
			Id:          "1",
			Items:       []string{"foo"},
			Description: "will change",
			Price:       99.99,
			Destination: "korea",
		},
		{
			Id:          "2",
			Items:       []string{"bar"},
			Description: "will change",
			Price:       129.99,
			Destination: "korea",
		},
		{
			Id:          "3",
			Items:       []string{"foobar"},
			Description: "will change",
			Price:       79.99,
			Destination: "korea",
		},
	} {
		database[v.Id] = v
	}

	return database
}
