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

메타데이터를 생성하고 RPC 호출 컨텍스트를 지정함으로써 **클라이언트에서 gRPC서버로 메타데이터를 보낼 수 있다.**

메타데이터를 서버로 보내는 방법은 두가지가 있는데 **권장하는 방법은 key,value를 context에 추가를 하는 것이다.** 

// TODO: 병합되는 것 직접 해봐야 함.
첫번째는 `AppendToOutgoingContext`를 이용하여 context에 메타데이터를 추가할 수 있는데 메타데이터를 추가할 때 키가 존재한다면 **병합이 되어 메타데이터에 추가된다.** 

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








## REFERENCE

https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md

https://stackoverflow.com/questions/57060602/what-is-the-difference-between-metadata-fromoutgoingcontext-and-metadata-frominc

