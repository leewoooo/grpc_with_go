package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	gretterpb "grpc_with_go/chap_05/proto/gretter"
	orderpb "grpc_with_go/chap_05/proto/unray"

	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
)

const (
	serverURL = "localhost:50051"
)

func main() {
	ctx := context.Background()
	conn, err := grpc.Dial(
		serverURL,
		grpc.WithInsecure(),
	)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
	}

	orderClient := orderpb.NewOrderManagementClient(conn)
	order, err := orderClient.GetOrder(ctx, &orderpb.OrderId{Id: "102"})
	if err != nil {
		errCode := status.Code(err)

		errStatus := status.Convert(err)
		for _, detail := range errStatus.Details() {
			switch info := detail.(type) {
			case *epb.BadRequest_FieldViolation:
				logrus.WithContext(ctx).Errorf("order client failed getOrder with status: %s, info: %v", errCode.String(), info)
			default:
				logrus.WithContext(ctx).Errorf("Unexpected Error type:%s", info)
			}
		}
	}
	logrus.WithContext(ctx).Infof("selected Order: %+v", order)

	gretterClient := gretterpb.NewGreeterClient(conn)
	resp, _ := gretterClient.SayHello(ctx, &gretterpb.HelloRequest{Name: "leewoooo"})
	logrus.WithContext(ctx).Infof("SayHello resp: %s", resp.GetMessage())
}
