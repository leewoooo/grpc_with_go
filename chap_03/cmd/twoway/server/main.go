package main

import (
	"fmt"
	pb "grpc_with_go/chap_03/proto/twoway"
	"io"
	"log"
	"net"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	port      = "9000"
	batchSize = 3
)

// OrderManagement implement OrderManagement Service interface
type OrderManagement struct {
	database map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

// NewOrderManagement create OrderManagement instace
func NewOrderManagement(database map[string]*pb.Order) *OrderManagement {
	return &OrderManagement{database: database}
}

// ProcessOrders ...
func (o *OrderManagement) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	orderBatchSize := 1
	var combinedShipmentMap = make(map[string]*pb.CombinedShipment)
	for {
		orderID, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// client에서 stream의 종료 시그널을 받았기 때문에 남아있는 주문건 모두 배송
				for _, comb := range combinedShipmentMap {
					err = stream.Send(comb)
					if err != nil {
						logrus.WithContext(stream.Context()).
							WithError(err).
							Error("failed Send where stream")
						return err
					}
				}

				logrus.WithContext(stream.Context()).Info("Client stream end")
				return nil
			}
			logrus.WithContext(stream.Context()).
				WithError(err).
				Error("failed Recv where stream")
			return err
		}

		logrus.Infof("Request With OrderID: %s", orderID.GetValue())

		// 주문을 묶음 배송하기 위한 로직
		order := o.database[orderID.GetValue()]
		shipment, found := combinedShipmentMap[order.Destination]
		destination := order.GetDestination()
		if found {
			ord := o.database[orderID.GetValue()]
			shipment.OrdersList = append(shipment.OrdersList, ord)
			combinedShipmentMap[destination] = shipment
		} else {
			id, _ := gonanoid.New(0)
			comShip := &pb.CombinedShipment{Id: id, Status: "Processed"}
			ord := o.database[orderID.GetValue()]
			comShip.OrdersList = []*pb.Order{ord}
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		if batchSize == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				err = stream.Send(comb)
				if err != nil {
					logrus.WithContext(stream.Context()).
						WithError(err).
						Error("failed Send where stream")
					return err
				}
			}

			orderBatchSize = 0
			combinedShipmentMap = make(map[string]*pb.CombinedShipment)
		} else {
			orderBatchSize++
		}

		log.Println("orderBatchSize", orderBatchSize)
	}
}

func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).
			Fatal("failed create listener instance")
	}

	srv := grpc.NewServer()

	database := generateMockDatabase
	orderManagement := NewOrderManagement(database())

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
			Items:       []string{"mac"},
			Description: "m1",
			Price:       99.99,
			Destination: "korea",
		},
		{
			Id:          "2",
			Items:       []string{"iphone"},
			Description: "13 pro",
			Price:       129.99,
			Destination: "usa",
		},
		{
			Id:          "3",
			Items:       []string{"mac mini"},
			Description: "m1",
			Price:       79.99,
			Destination: "usa",
		},
		{
			Id:          "4",
			Items:       []string{"airpods"},
			Description: "pro",
			Price:       29.99,
			Destination: "korea",
		},
	} {
		database[v.Id] = v
	}

	return database
}
