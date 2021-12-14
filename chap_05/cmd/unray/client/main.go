package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	interceptor "grpc_with_go/chap_05/internal/interceptor/unray"
	pb "grpc_with_go/chap_05/proto/unray"

	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
)

const (
	serverURL = "localhost:9000"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.Dial(
		serverURL,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(interceptor.UnrayOrderManagementClientInterceptor),
	)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
	}

	client := pb.NewOrderManagementClient(conn)

	order := &pb.Order{
		Items:       []string{"mac pro"},
		Description: "apple",
		Price:       199.99,
		Destination: "korea seoul",
	}

	// ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Microsecond*30))
	// defer cancel()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	ID, err := client.AddOrder(ctx, order)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed AddOrder: %+v", order)
		return
	}

	selectedOrder, err := client.GetOrder(ctx, ID)
	if err != nil {
		errCode := status.Code(err)
		if errCode == codes.InvalidArgument {
			logrus.WithContext(ctx).WithError(err).Error("failed GetOrder")

			errorStatus := status.Convert(err)
			for _, detail := range errorStatus.Details() {
				switch info := detail.(type) {
				case *epb.BadRequest_FieldViolation:
					logrus.WithContext(ctx).Errorf("Reqeust Field Invalid: %s", info)
				default:
					logrus.WithContext(ctx).Errorf("Unexpected Error type:%s", info)
				}
			}
		} else {
			logrus.WithContext(ctx).WithError(err).Error("unhandled error with errorCode: %s", errCode)
		}
	}

	logrus.WithContext(ctx).Infof("selected Order: %+v", selectedOrder)
}
