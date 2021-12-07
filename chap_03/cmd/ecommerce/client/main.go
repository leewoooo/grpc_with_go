package main

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "grpc_with_go/chap_03/proto/ecommerce"
)

var (
	serverURL = "localhost:9000"
	once      sync.Once
	client    pb.OrderManagementClient
)

func main() {
	once.Do(func() {
		conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
		if err != nil {
			logrus.WithError(err).Fatalf("failed grpc Dial with serverURL: %s", serverURL)
		}

		client = pb.NewOrderManagementClient(conn)
	})

	ctx := context.Background()

	// unray
	targetID := "102"

	selectedOrder, err := client.GetOrder(ctx, wrapperspb.String(targetID))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed getOrder with targetID: %s", targetID)
		return
	}

	logrus.WithContext(ctx).Infof("GetOrder result: %+v", selectedOrder)

	// server stream
	targetItem := "Amazon Echo"
	serverStream, err := client.SearchOrder(ctx, wrapperspb.String(targetItem))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed get ServerStream with targetItem: %s", targetItem)
		return
	}

	for {
		order, err := serverStream.Recv()
		if err != nil {
			if err == io.EOF {
				logrus.WithContext(serverStream.Context()).Info("server stream END")
				break
			}
			logrus.WithContext(serverStream.Context()).WithError(err).Error("failed Order Recv")
			return
		}

		logrus.WithContext(serverStream.Context()).Infof("selected Order: %+v", order)
	}

	// client stream
	clientStream, err := client.UpdateOrder(ctx)
	if err != nil {
		logrus.WithContext(clientStream.Context()).WithError(err).Error("failed get clientStream")
		return
	}

	err = clientStream.Send(&pb.Order{Id: "102", Items: []string{"Google Pixel 10A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00})
	if err != nil {
		logrus.WithContext(clientStream.Context()).WithError(err).Error("failed Send Order where client stream")
		return
	}
	err = clientStream.Send(&pb.Order{Id: "103", Items: []string{"Apple Watch SE"}, Destination: "San Jose, CA", Price: 400.00})
	if err != nil {
		logrus.WithContext(clientStream.Context()).WithError(err).Error("failed Send Order where client stream")
		return
	}

	// server에서 응답이 올 때 까지 blocking
	response, err := clientStream.CloseAndRecv()
	if err != nil {
		logrus.WithContext(clientStream.Context()).WithError(err).Error("failed clientStream CloseAndRecv")
		return
	}
	logrus.Info(response.GetValue())

	// two way
	twoWayStream, err := client.ProcessOrder(ctx)
	if err != nil {
		logrus.WithContext(twoWayStream.Context()).WithError(err).Error("failed get twoWayStream")
		return
	}

	err = twoWayStream.Send(wrapperspb.String("102"))
	if err != nil {
		logrus.WithContext(twoWayStream.Context()).WithError(err).Error("failed Send where twoWayStream with targetID: %s", "102")
		return
	}
	err = twoWayStream.Send(wrapperspb.String("103"))
	if err != nil {
		logrus.WithContext(twoWayStream.Context()).WithError(err).Error("failed Send where twoWayStream with targetID: %s", "102")
		return
	}
	err = twoWayStream.Send(wrapperspb.String("104"))
	if err != nil {
		logrus.WithContext(twoWayStream.Context()).WithError(err).Error("failed Send where twoWayStream with targetID: %s", "102")
		return
	}
	<-time.After(time.Second * 3)
	channel := make(chan struct{})
	go asyncClientBidrectionalRPC(twoWayStream, channel)

	err = twoWayStream.Send(wrapperspb.String("105"))
	if err != nil {
		logrus.WithContext(twoWayStream.Context()).WithError(err).Error("failed Send where twoWayStream with targetID: %s", "102")
		return
	}

	if err := twoWayStream.CloseSend(); err != nil {
		log.Fatal(err)
	}

	<-channel
}

func asyncClientBidrectionalRPC(twoWayStrem pb.OrderManagement_ProcessOrderClient, channel chan struct{}) {
	defer close(channel)
	for {
		comb, err := twoWayStrem.Recv()
		if err != nil {
			if err == io.EOF {
				logrus.WithContext(twoWayStrem.Context()).Info("Server Stream END")
				break
			}
			logrus.WithContext(twoWayStrem.Context()).WithError(err).Error("failed Recv where stream")
			return
		}

		logrus.WithContext(twoWayStrem.Context()).Infof("Combined shipment: %+v", comb)
	}
}
