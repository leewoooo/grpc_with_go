package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "grpc_with_go/chap_02/proto/productinfo"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		logrus.Fatal(err)
	}

	client := pb.NewProductInfoClient(conn)

	givenProduct := &pb.Product{
		Name:        "mac m1",
		Description: "pro",
		Price:       99.99,
	}

	insertedID, err := client.AddProduct(ctx, givenProduct)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed excute AddProduct")
	}

	selectedpProduct, err := client.GetProduct(ctx, insertedID)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatalf("failed excute GetProduct with insertedID :%s", insertedID.GetId())
	}

	logrus.Infof("product: %v", selectedpProduct)
}
