package main

import (
	"context"
	"fmt"
	pb "grpc_with_go/chap_02/proto/productinfo"
	"net"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const portNumber = "9000"

// Server implement ProductInfoServer
type Server struct {
	database map[string]*pb.Product
	pb.UnimplementedProductInfoServer
}

// NewServer create productinfo grpc Server instance
func NewServer(database map[string]*pb.Product) pb.ProductInfoServer {
	return &Server{database: database}
}

// AddProduct add product
func (s *Server) AddProduct(ctx context.Context, product *pb.Product) (*pb.ProductID, error) {
	logrus.WithContext(ctx).
		Infof("AddProduct reqeust with name:%s, description: %s, price:%v", product.GetName(), product.GetDescription(), product.GetPrice())

	id, err := gonanoid.New(0)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("AddProduct failed gen gonanoid")
		return nil, status.Errorf(codes.Internal, "failed gen gonanoid")
	}

	s.database[id] = product

	return &pb.ProductID{Id: id}, nil
}

// GetProduct get product with productID
func (s *Server) GetProduct(ctx context.Context, productID *pb.ProductID) (*pb.Product, error) {
	logrus.WithContext(ctx).
		Infof("getproduct request with productID: %s", productID.GetId())

	product, ok := s.database[productID.GetId()]

	if !ok {
		logrus.WithContext(ctx).Warnf("GetProduct not exist with ID: %s", productID.GetId())
		return nil, status.Errorf(codes.NotFound, "not exist product")
	}

	return product, nil
}

func main() {
	port := fmt.Sprintf(":%s", portNumber)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.Fatal(err)
	}

	srv := grpc.NewServer()

	localDatabase := make(map[string]*pb.Product)
	productServer := NewServer(localDatabase)

	pb.RegisterProductInfoServer(srv, productServer)

	logrus.Printf("start gRPC server on %s port", portNumber)
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}
