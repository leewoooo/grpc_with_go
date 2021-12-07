package main

import (
	"context"
	"fmt"
	pb "grpc_with_go/chap_03/proto/ecommerce"
	"io"
	"net"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	orderBatchSize = 3
	port           = "9000"
	database       = make(map[string]*pb.Order)
)

func init() {
	database["102"] = &pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	database["103"] = &pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	database["104"] = &pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	database["105"] = &pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	database["106"] = &pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}

// OrderManagementServerImpl implement OrderManagement service interface
type OrderManagementServerImpl struct {
	db map[string]*pb.Order
	pb.UnimplementedOrderManagementServer
}

// NewOrderManagementServer create OrderManagementServer
func NewOrderManagementServer(db map[string]*pb.Order) pb.OrderManagementServer {
	return &OrderManagementServerImpl{db: database}
}

// AddOrder add order
func (o *OrderManagementServerImpl) AddOrder(ctx context.Context, order *pb.Order) (*wrapperspb.StringValue, error) {
	logrus.WithContext(ctx).Infof("Try Add Order with order: %+v", order)

	id, _ := gonanoid.New()
	order.Id = id
	o.db[id] = order
	logrus.WithContext(ctx).Infof("Success Add Order with id: %s", id)

	return &wrapperspb.StringValue{Value: id}, nil
}

// GetOrder get Order by ID
func (o *OrderManagementServerImpl) GetOrder(ctx context.Context, orderID *wrapperspb.StringValue) (*pb.Order, error) {
	logrus.WithContext(ctx).Infof("Try Get Order with id :%s", orderID.GetValue())

	order, found := o.db[orderID.GetValue()]
	if !found {
		logrus.WithContext(ctx).Warnf("order not exist with id: %s", orderID.GetValue())
		return nil, grpc.Errorf(codes.NotFound, "order not exist  with id: %s", orderID.GetValue())
	}

	return order, nil
}

// SearchOrder search Order by item Name (with server stream)
func (o *OrderManagementServerImpl) SearchOrder(key *wrapperspb.StringValue, stream pb.OrderManagement_SearchOrderServer) error {
	logrus.WithContext(stream.Context()).Infof("Try Search Order with itemName: %s", key.GetValue())
	// database에 order들에서
	for _, order := range o.db {
		// key에 해당하는 order를 찾아 send
		for _, item := range order.Items {
			if strings.Contains(item, key.GetValue()) {
				err := stream.Send(order)
				if err != nil {
					logrus.WithContext(stream.Context()).WithError(err).Error("failed Send Order")
					return err
				}

				logrus.WithContext(stream.Context()).Info("Matching Order key: %s", key.GetValue())
				break
			}
		}
	}

	// nil이 return 되면 client 쪽에서는 EOF를 받게 된다.
	return nil
}

// UpdateOrder ...
func (o *OrderManagementServerImpl) UpdateOrder(stream pb.OrderManagement_UpdateOrderServer) error {
	responseValue := "Updated Order Ids:"

	for {
		order, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err = stream.SendAndClose(wrapperspb.String(responseValue))
				if err != nil {
					logrus.WithError(err).Errorf("failed client stream response with value: %s", responseValue)
					return err
				}
				break
			}
			logrus.WithContext(stream.Context()).WithError(err).Error("failed recv client stream")
			return err
		}

		logrus.Infof("Try Update Order with %+v", order)

		o.db[order.GetId()] = order
		responseValue += fmt.Sprintf("%s ", order.GetId())
	}

	return nil
}

// ProcessOrder ...
func (o *OrderManagementServerImpl) ProcessOrder(stream pb.OrderManagement_ProcessOrderServer) error {
	batchSize := 1
	combinedShipmentMap := make(map[string]*pb.CombinedShipment)

	for {
		orderID, err := stream.Recv()
		if err != nil {
			// eof check
			if err == io.EOF {
				// client의 stream이 종료되었기 때문에 map에 있는 모든 것을 배송
				for _, comb := range combinedShipmentMap {
					err = stream.Send(comb)
					if err != nil {
						logrus.WithContext(stream.Context()).WithError(err).Error("failed send stream")
						return err
					}
				}
				break
			}

			// error handling
			logrus.WithContext(stream.Context()).WithError(err).Error("failed recv stream")
			return err
		}

		// db에서 검색하여 묶음 배송처리
		destination := o.db[orderID.GetValue()].Destination
		comb, ok := combinedShipmentMap[destination]
		if !ok {
			combinedShipmentMap[destination] = &pb.CombinedShipment{
				Id:        fmt.Sprintf("comb-%s", destination),
				Status:    "PROCESSED",
				OrderList: []*pb.Order{o.db[orderID.GetValue()]},
			}
		} else {
			comb.OrderList = append(comb.OrderList, o.db[orderID.GetValue()])
		}

		// batch 크기가 가득 찼는지?
		if orderBatchSize == batchSize {
			for _, comb := range combinedShipmentMap {
				if err = stream.Send(comb); err != nil {
					logrus.WithContext(stream.Context()).WithError(err).Error("failed send stream")
					return err
				}

				// reset batch size && map
				batchSize = 0
				combinedShipmentMap = make(map[string]*pb.CombinedShipment)
			}
		} else {
			batchSize++
		}
	}
	return nil
}

func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).Fatalf("failed get Listener with port: %s", port)
	}

	srv := grpc.NewServer()
	orderManagement := NewOrderManagementServer(database)

	//register
	pb.RegisterOrderManagementServer(srv, orderManagement)

	logrus.Infof("grpc Server starting withg port: %s...", port)
	if err = srv.Serve(lis); err != nil {
		logrus.WithError(err).Fatalf("failed starting Server with port: %s", port)
	}
}
