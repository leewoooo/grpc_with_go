gRPC 고급기능 (multiplexing)
===

지금까지는 하나의 gRPC서비스가 등록되고 gRPC 클라이언트 연결이 하나의 클라이언트 스텁에서만 사용되는 gRPC를 살펴보았지만 gRPC를 사용하면 **동일한 gRPC 서버에서 여러 gRPC 서비스를 실행할 수 있고, 여려 gRPC 클라이언트 스텁에 동일한 gRPC 클라이언트 연결을 사용할 수 있다.**

이 기능을 멀티플렉싱(multiplexing)이라고 한다.

<img src = https://user-images.githubusercontent.com/74294325/145997941-5cbfaf90-ac59-4294-b9f4-aaeaa6eec405.png>

<Br>

간단하게 이야기 하자면 서버측은 하나의 gRPC 서버 인스턴스에 **여러개의 gRPC 서비스를 등록시키는 것을 말한다.** 

클라이언트측은 한번의 `Dial()`을 통해 얻어온 커넥션을 가지고 여러개의 **Client 인스턴스를 생성하는 것을 말한다.**

여러 서비스를 실행하거나 여러 스텁간 동일한 연결을 사용하는 것은 **gRPC개념과는 상관이 없는 설계 선택의 문제이다.** MSA와 같은 대부분의 일반적인 사례에서는 **두 서비스 간에 동일한 gRPC 서버 인스턴스를 공유하지 않는다.**

<br>

## 예제 

서버측 예제 코드는 다음과 같다. 

```go
func main() {
	port := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithError(err).
			Fatal("failed create listener instance")
	}

	srv := grpc.NewServer() // 1

	orderpb.RegisterOrderManagementServer(srv, NewOrderManagement(database)) // 2

	gretterpb.RegisterGreeterServer(srv, NewGretter()) // 3

	logrus.Infof("gRPC server starting with port:%s...", port)
	if err := srv.Serve(lis); err != nil {
		logrus.WithError(err).
			Fatal("failed serve gRPC server with port: %s", port)
	}
}
```

1. gRPC 서버 인스턴스를 생성한다.

2. `RegisterOrderManagementServer`를 이용하여 OrderManagement 서비스를 등록한다.

3. 동일한 gRPC서버에 `RegisterGreeterServer`를 이용하여 Gretter 서비스를 등록한다.

<br>

클라이언트측 예제코드는 다음과 같다.

```go
ctx := context.Background()
	conn, err := grpc.Dial( // 1
		serverURL,
		grpc.WithInsecure(),
	)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
	}

	orderClient := orderpb.NewOrderManagementClient(conn) // 2
    // ...
    gretterClient := gretterpb.NewGreeterClient(conn) // 3
    // ...
```

1. gRPC 클라이언트 커넥션을 얻어온다.

2. `NewOrderManagementClient`를 호출하여 OrderMenagement **클라이언트 인스턴스를 생성한다.**

3. 동일한 gRPC 커넥션으로 `NewGreeterClient`를 호출하여 Gretter **클라이언트 인스턴스를 생성한다.**

<br>

