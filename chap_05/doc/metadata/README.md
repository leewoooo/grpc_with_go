gRPC 고급기능 (metaData)
===

비지니스 로직 및 소비자와 직접 관련된 정보는 원격 메서드의 호출 인자의 일부이다. **그러나 특정 조건에서 RPC 인자의 일부가 돼서는 안 된다.**

이런 경우 gRPC 서비스나 gRPC 클라이언트에서 보내거나 받을 수 있는 **gRPC 메타데이터를 사용할 수 있다.** 메타데이터는 **키(문자열)/ 값 에 대한 목록 형식으로 구성된다.**

메타데이터의 가장 일반적인 사용은 gRPC 애플리케이션 간에 **보안 헤더를 교환하는 것이다.** 마찬가지로 임의의 정보를 교환하는데도 사용할 수 있고, gRPC 메타데이터 API는 개발되는 **인터셉터 내부에서 많이 사용된다.**

<img src =https://user-images.githubusercontent.com/74294325/146209037-111fe248-eb8d-4441-82b4-c840fb426f32.png>

<br>

## 메타데이터 생성과 조회

gRPC 애플리케이션에서 메타데이터를 생성하는 일은 매우 간단하다. go언어 기준으로 2가지 방법이 있다. 메타데이터는 **go에서 일반 맵으로 표현된다.** 아울러 `metadata.Pairs()`를 이용하여 메타데이터를 쌍으로 만들 수 있다.(**같은 key가 존재할 경우 배열로 만들 수 있다.**)

```go
// 방법 1
md := metadata.New(map[string]string{ "key1" : "val1", "key2" : "val2" })

// 방법 2
md := metadata.Pairs(
    "key1", "val1",
    "key1", "val2", // key1는 []string{"val1", "val2"} 를 갖게 된다.
    "key2", "val3",
)
```

<br>

클라이언트나 서버에서 메타데이터를 읽으려면 `metadata.FromIncomingContext()`를 호출하여 인자로 넘어온 **context에서 메타데이터를 얻어올 수 있다.** go언어 기준으로는 **맵을 반환한다.**
```go
func (s *server) SomeRPC(ctx context.Context, in *pb.SomeRequest) (*pb.SomeResponse, err) {
    md, ok := metadata.FromIncomingContext(ctx)
    // do something with metadata
}
```

<br>

## 메타데이터 전송과 수신 : 클라이언트 side

### 전송

메타데이터를 생성하고 RPC 호출 컨텍스트를 지정함으로써 **클라이언트에서 gRPC서버로 메타데이터를 보낼 수 있다.**

메타데이터를 서버로 보내는 방법은 두가지가 있는데 **권장하는 방법은 key,value를 context에 추가를 하는 것이다.** 

첫번째는 `AppendToOutgoingContext`를 이용하여 context에 메타데이터를 추가할 수 있는데 메타데이터를 추가할 때 키가 존재한다면 **병합이 되어 메타데이터에 추가된다. (해당 key는 []stirng을 값으로 가짐)** 

코드는 아래와 같다.

```go
// create a new context with some metadata
ctx := metadata.AppendToOutgoingContext(ctx, "k1", "v1", "k1", "v2", "k2", "v3")

// later, add some more metadata to the context (e.g. in an interceptor)
ctx := metadata.AppendToOutgoingContext(ctx, "k3", "v4")

// make unary RPC
response, err := client.SomeRPC(ctx, someRequest)

// or make streaming RPC
stream, err := client.SomeStreamingRPC(ctx)
```

<br>

두번째는 `NewOutgoingContext`를 이용하는 것이다. 하지만 이 방법을 사용하는 시점 이전에 **context에 메타데이터가 포함되어 있는 경우 새로 추가하는 메타데이터로 대체된다. (기존 것은 지워지고 새로운 것으로 덮어 쓰여 진다.)** 공식 문서에 보면 `AppendToOutgoingContext`를 이용할 때 보다 느리다고 적혀 있다.
>This is slower than using `AppendToOutgoingContext`

코드는 아래와 같다.

```go
// create a new context with some metadata
md := metadata.Pairs("k1", "v1", "k1", "v2", "k2", "v3")
ctx := metadata.NewOutgoingContext(context.Background(), md) // 1

// later, add some more metadata to the context (e.g. in an interceptor)
send, _ := metadata.FromOutgoingContext(ctx) // 2
newMD := metadata.Pairs("k3", "v3") // 3
ctx = metadata.NewOutgoingContext(ctx, metadata.Join(send, newMD)) // 4

// make unary RPC
response, err := client.SomeRPC(ctx, someRequest)

// or make streaming RPC
stream, err := client.SomeStreamingRPC(ctx)
```

1. `context.Background()`에 새로 생성한 메타데이터를 `NewOutgoingContext()`를 이용하여 추가한다.

2. context로 부터 **현재 context가 가지고 있는 메타데이터를 추출한다.**

3. 새로운 메타데이터 생성

4. `Join()`을 이용하여 **추출한 메타데이터와 새로운 메타데이터를 합친 후** `NewOutgoingContext()`를 이용하여 기존의 메타데이터와 **대체한다.**
> Joint()은 두 개의 메타데이터를 받아 반복문을 돌며 새로운 메타데이터에 추가 한 후 메타데이터를 return한다. <br> (https://github.com/grpc/grpc-go/blob/79e9c9571a1949d3abae203a127fa5d4f02fb071/metadata/metadata.go#L137)

<br>

### 수신

컨텍스트에서 설정한 메타데이터는 **gRPC 헤더나 트레일러 레벨로 변환되기 때문에 클라이언트가 해당 헤더를 보내면 수신자가 헤더로 수신을 해야한다.**

따라서 클라이언트로에서 메타데이터를 수신할 때는 **헤더나 트레일러로 취급해야 한다.** 

```go
// unray
var header, trailer metadata.MD // variable to store header and trailer // 1
r, err := client.SomeRPC( // 2
    ctx,
    someRequest,
    grpc.Header(&header),    // will retrieve header
    grpc.Trailer(&trailer),  // will retrieve trailer
)

// stream
header, err := stream.Header() // 3

trailer := stream.Trailer() // 4

// do something with header and trailer
```

1. RPC 호출에서 반환될 헤더(메타데이터)와 트레일러를 저장할 변수이다.

2. 단일 RPC에 대해 반환되는 값을 저장하려면 gRPC 서비스 method를 호출할 때 `grpc.CallOption`으로 **헤더와 트레일러의 참조(주소 값)을 전달해야 한다.** 

3. 스트리밍 RPC의 경우 스트림에서 **헤더를 가져온다.**

4. 스트림에서 **트레일러**를 가져오는데 트레일러는 상태코드와 상태 메세지를 보내는데 사용된다.

<br>

## 메타데이터 전송과 수신 : 서버 side

### 수신

서버에서 메타데이터를 수신하는 것도 매우 간단하다. `metadata.FromIncomingContext(ctx)`를 이용하여 컨텍스트로 부터 간단하게 메타데이터를 얻을 수 있다.

```go
// unray
func (s *server) SomeRPC(ctx context.Context, in *pb.someRequest) (*pb.someResponse, error) {
    md, ok := metadata.FromIncomingContext(ctx) // 1
    // do something with metadata
}

// stream
func (s *server) SomeStreamingRPC(stream pb.Service_SomeStreamingRPCServer) error {
    md, ok := metadata.FromIncomingContext(stream.Context()) // get context from stream // 2
    // do something with metadata
}
```

1. 단일 RPC의 경우 인자로 들어온 **컨텍스트로 부터 메타데이터를 얻는다.**

2. 스트리밍 RPC의 경우 **스트림에서 얻어온 컨텍스트로 부터 메타데이터를 추출한다.**

<br>

### 전송

서버에서 메타데이터를 보내려면 메타데이터가 있는 **헤더를 보내거나 메타데이터가 있는 트레일러를 지정한다.**

메타데이터를 생성하는 것은 **클라이언트와 동일하다** 예제 코드를 살펴보자면 다음과 같다.

```go
// unray
func (s *server) SomeRPC(ctx context.Context, in *pb.someRequest) (*pb.someResponse, error) {
    // create and send header
    header := metadata.Pairs("header-key", "val") // 1
    grpc.SendHeader(ctx, header) // 2
    // create and set trailer
    trailer := metadata.Pairs("trailer-key", "val") // 1
    grpc.SetTrailer(ctx, trailer) // 3
}

// stream
func (s *server) SomeStreamingRPC(stream pb.Service_SomeStreamingRPCServer) error {
    // create and send header
    header := metadata.Pairs("header-key", "val") // 1
    stream.SendHeader(header) // 4
    // create and set trailer
    trailer := metadata.Pairs("trailer-key", "val") // 1
    stream.SetTrailer(trailer) // 5
}
```

1. `Pairs()`를 이용하여 메타데이터를 생성한다.

2. grpc 패키지의 `SendHeader()`를 이용하여 생성한 메타데이터를 컨텍스트에 포함시킨다.

3. `SetTrailer()`를 이용하여 생성한 메타데이터를 트레일러로 지정한다.

4. **stream 인스턴스**에 생성한 메타데이터를 포함시킨다.

5. **stream 인스턴스**에 생성한 메타데이터를 트레일러로 지정한다.







## REFERENCE

https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md

https://stackoverflow.com/questions/57060602/what-is-the-difference-between-metadata-fromoutgoingcontext-and-metadata-frominc

