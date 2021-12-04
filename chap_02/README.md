2장 gRPC시작
===

2장에서는 다음과 같은 순서로 gRPC의 흐름을 파악한다.

1. 프로토콜 버퍼를 사용해 gRPC서비스 정의 지정
2. 서버 스켈레톤과 클라이언트 스텁생성
3. 서비스 비즈니스 로직 구현
4. 구현된 서비스에 대한 gRPC서버 실행
5. gRPC 클라이언트를 통한 서비스 호출

<br>

## 1. 프로토콜 버퍼를 사용해 gRPC서비스 정의

gRPC를 이용한 어플리케이션을 작성할 때 가장 먼저 해야할 일은 **소비자가 원격으로 호출할 수 있는 메서드와 메서드의 파라미터, 사용할 메세지포멧등을 포함한 서비스 인터페이스을 정의하는 것이다.**

모든 서비스 정의는 gRPC에서 사용되는 IDL인 프로토콜 버퍼 정의로 작성된다.

프로토콜 버퍼 정의를 작성할 때 서비스와 메세지 타입을 정의한다.

서비스는 **메서드로 각 메서드 타입, 입출력 파라미터**로 정의된다.

메세지는 **필드로 구성되며 각 필드는 해당 타입과 고유 인덱스의 값으로 정의된다.**

<Br>

### 2장의 proto 예제

```proto
syntax = "proto3"; // 1

package ecommerec.v1; // 2
option go_package = "grpc_with_go/chap_02/proto/productinfo"; // 3


service ProductInfo { // 4
    rpc addProduct (Product) returns (ProductID);
    rpc getProduct (ProductID) returns (Product);
}

message Product { // 5
    string id = 1; // 6
    string name = 2;
    string description = 3;
    float price = 4;
}

message ProductID{
    string id = 1;
}
```

1: 서비스 정의에 사용되는 **프로토콜 버퍼 버전을 지정**

2: 패키지 이름은 프로토콜 메세지 타입 사이의 이름 충돌을 방지 및 코드 생성에 사용, 버전을 관리할 경우 `. v1`과  같이 패키지 명을 부여할 수 있다.

3: 현 gRPC를 go언어 기준으로 하고 있으며 1장에서 살펴보았듯 go로 code를 생성하려면 **해당 option을 강제하고 있다.**

4: 서비스의 서비스 인터페이스를 정의한다. 추 후 **서버를 생성할 때 해당 인터페이스를 구현한다.** 또한 메서드는 **프로토콜 버퍼 규칙에 따라 하나의 입력과 하나의 출력만 가질 수 있다.**

5. 서비스의 입력이나 출력에 사용되는 메세지를 정의

6. 메세지의 필드는 고유한 식별 번호를 가지고 있다.

<br>

## 2. 서버 스켈레톤과 클라이언트 스텁생성

1장에서 살펴본 것과 같이 build 툴을 사용하지 않았으며 명령어와 프로토콜 버퍼 컴파일러를 통해 직접 생성하였으며 명령어는 다음과 같다.

```
protoc --go_out=./chap_02/proto/productinfo \
--go_opt=paths=source_relative \
--go-grpc_out=./chap_02/proto/productinfo \
--go-grpc_opt=paths=source_relative \
-I ./chap_02/IDL ./chap_02/IDL/product_info.proto
```

자세한 것은 1장 참조 (https://github.com/leewoooo/grpc_with_go/tree/main/chap_01/IDL)

<br>


프로토콜 버퍼 컴파일러를 통해 코드를 생성하면 golang기준으로 2개의 파일이 생성된다.

<img src = https://user-images.githubusercontent.com/74294325/144440330-c3375634-eabd-4151-bb15-b821ce0e6036.png>

<br>

### product_info.pb.go

해당 파일에는 서비스 인터페이스에서 정의한 메세지에 관한 code들이 생성된다. 

```go
type Product struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description string  `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Price       float32 `protobuf:"fixed32,4,opt,name=price,proto3" json:"price,omitempty"`
}
```

<br>

`Product` 메세지를 살펴보자면.. `Product`라는 구조체가 생성되었으며 해당 파일에는 `Product`에 field들에 접근할 수 있는 **Getter** 메소드가 제공되어 있다.

<br>

### product_info_grpc.pb.go

해당 파일에는 gRPC 스켈레톤 서버, 클라이언트 스텁에 관련된 code들이 들어 있다. 

먼저 서버에 관련된 code들을 살펴보자면 서비스 인터페이스는 다음과 같은 code로 생성되었다.

```go
type ProductInfoServer interface {
	AddProduct(context.Context, *Product) (*ProductID, error)
	GetProduct(context.Context, *ProductID) (*Product, error)
	mustEmbedUnimplementedProductInfoServer()
}
```

<Br>

gRPC 서버를 생성할 때 해당 서비스 인터페이스를 구현하여 비지니스 로직을 작성한다. 

`mustEmbedUnimplementedProductInfoServer()` 는 이 서비스에 대한 순방향 호환성을 위해 내장될 수 있으며 이 인터페이스를 사용하지 않는 것을 권장한다.(컴파일 오류) 
>UnsafeProductInfoServer may be embedded to opt out of forward compatibility for this service. Use of this interface is not recommended, as added methods to ProductInfoServer will result in compilation errors.

(https://github.com/grpc/grpc-go/blob/master/cmd/protoc-gen-go-grpc/README.md)

<br>

다음은 클라이언트 스텁에 대한 코드 생성을 살펴보면 다음과 같다.

```go
type ProductInfoClient interface {
	AddProduct(ctx context.Context, in *Product, opts ...grpc.CallOption) (*ProductID, error)
	GetProduct(ctx context.Context, in *ProductID, opts ...grpc.CallOption) (*Product, error)
}

type productInfoClient struct {
	cc grpc.ClientConnInterface
}

func NewProductInfoClient(cc grpc.ClientConnInterface) ProductInfoClient {
	return &productInfoClient{cc}
}
```

`productInfoClient` 구조체는 `ProductInfoClient` 인터페이스를 구현하고 있는 구조체 이며 생성자를 통해 **gRPC Connection**을 주입받아 `productInfoClient`를 return해주게 됩니다. (`ProductInfoClient`를 구현한 메소드는 product_info_grpc.pb.go 파일에 포함되어 있습니다. **직접 구현하지 않아도 된다.**)

<br>

## 3. 서비스 비즈니스 로직 구현

생성된 서비스 인터페이스를 구현하여 서비스 비즈니스 로직을 구현한다. 예제 코드는 다음과 같다.

```go
// Server implement ProductInfoServer
type Server struct { // 1
	database map[string]*pb.Product
	pb.UnimplementedProductInfoServer
}

// NewServer create productinfo grpc Server instance
func NewServer(database map[string]*pb.Product) pb.ProductInfoServer { //2
	return &Server{database: database}
}

// AddProduct add product
func (s *Server) AddProduct(ctx context.Context, product *pb.Product) (*pb.ProductID, error) { // 3
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
func (s *Server) GetProduct(ctx context.Context, productID *pb.ProductID) (*pb.Product, error) { // 4
	logrus.WithContext(ctx).
		Infof("getproduct request with productID: %s", productID.GetId())

	product, ok := s.database[productID.GetId()]

	if !ok {
		logrus.WithContext(ctx).Warnf("GetProduct not exist with ID: %s", productID.GetId())
		return nil, status.Errorf(codes.NotFound, "not exist product")
	}

	return product, nil
}
```

<Br>

1: 서비스 인터페이스를 구현할 구조체를 정의한다.

2: 생성자를 통해 해당 구조체에 필요한 인스턴스들을 DI받아 사용할 수 있다.

3, 4: 서비스 인터페이스에 있는 구조체를 리시버를 이용해 구현하여 서비스 로직을 작성한다.

<br>

## 4. 구현된 서비스에 대한 gRPC서버 실행

서버를 실행하기 위한 code는 다음과 같다.

```go
...
pb "grpc_with_go/chap_02/proto/productinfo" // 1
...

port := fmt.Sprintf(":%s", portNumber)
lis, err := net.Listen("tcp", port) // 2
if err != nil {
	logrus.Fatal(err)
}

srv := grpc.NewServer() // 3

localDatabase := make(map[string]*pb.Product)
productServer := NewServer(localDatabase) // 4

pb.RegisterProductInfoServer(srv, productServer) // 5

logrus.Printf("start gRPC server on %s port", portNumber) //6
if err := srv.Serve(lis); err != nil {
	logrus.Fatal(err)
}
```

<Br>

1: 프로토버프 컴파일러로 생성된 코드가 포함된 패키지를 임포트 받는다.

2: gRPC 서버가 바인딩 하고자 하는 **TCP 리스너를 서버 포트와 함께 생성한다.**

3: RPC Go API를 호출하여 새 **gRPC 인스턴스를 생성한다.**

4: 서비스 인터페이스를 구현한 구조체의 인스턴스를 **생성자 메소드를 통해 생성한다.**

5: 서비스 인터페이스를 구현한 **인스턴스를 새로 생성한 gRPC 서버에 등록한다.**

6: 리스너와 함께 gRPC 서버의 **리스닝을 시작한다.**

<br>

위와 같은 과정을 통해 gRPC 서비스를 할 수 있는 서버는 준비가 되었다.

<br>

## 5. gRPC 클라이언트를 통한 서비스 호출

gRPC 서비스를 호출하는 클라이언트 쪽 코드는 간단하게 작성을 할 수 있다.

```go
...
pb "grpc_with_go/chap_02/proto/productinfo" // 1
...

ctx := context.Background() 
conn, err := grpc.Dial("localhost:9000", grpc. WithInsecure()) // 2
if err != nil {
	logrus.Fatal(err)
}

client := pb.NewProductInfoClient(conn) // 3

givenProduct := &pb.Product{
	Name:        "mac m1",
	Description: "pro",
	Price:       99.99,
}

insertedID, err := client.AddProduct(ctx, givenProduct) // 4
if err != nil {
	logrus.WithContext(ctx).WithError(err).Fatal("failed excute AddProduct")
}

selectedpProduct, err := client.GetProduct(ctx, insertedID) // 5
if err != nil {
	logrus.WithContext(ctx).WithError(err).Fatalf("failed excute GetProduct with insertedID :%s", insertedID.GetId())
}
```

<br>

1: 서버와 같이 프로토버프 컴파일러로 생성된 코드가 포함된 패키지를 임포트 받는다.

2: 서버의 **주소와 포트**를 이용하여 **커넥션 인스턴스를 생성한다.** 위의 예제코드는 서버와 클라이언트 사이에 보안되지 않은 커넥션을 생성한다. (`WithInsecure()`)

3: 생성된 gRPC 커넥션 인스턴스를 이용하여 **클라이언트 스텁을 생성한다.** 생성된 클라이언트 스텁을 통해 gRPC 서버와 통신을 한다.

4, 5: 클라이언트 인스턴스를 통해 **IDL에서 정의한 서비스 메소드**를 **메세지를 인자로 하여 호출한다.** 리턴 값으로는 **IDL**에서 정의한 리턴 타입의 값을 받게 된다.

<br>

## 정리

2장의 내용을 통해 gRPC서비스를 **어떻게 제공해야 하는지에 대한 One cycle을 공부 할 수 있었다.**

전반적인 흐름은 위에서 정의한 것 처럼 아래의 흐름을 따르게 된다.

1. 프로토콜 버퍼를 사용해 gRPC서비스 정의 지정
2. 서버 스켈레톤과 클라이언트 스텁생성
3. 서비스 비즈니스 로직 구현
4. 구현된 서비스에 대한 gRPC서버 실행
5. gRPC 클라이언트를 통한 서비스 호출

rest 방식과는 다르게 client 스텁만 존재하면 **local method를** 호출하는 것과 같은 편안함과 통신의 결과에 대한  **type이 명확하다** 라는 장점을 가지고 있는 것 같다.





