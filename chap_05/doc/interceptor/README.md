gRPC 고급기능 (interceptor)
===

## Interceptor

gRPC 애플케이션을 만들 때 클라이언트나 서버에 원격 함수 실행 전후 **몇 가지 공통적인 로직을 실행할 필요가 있다.**

**인터셉터라는 확장 메커니즘을 사용해 로깅, 인증, 메트릭 등과 같은 특정 요구사항 충족을** 위해 RPC 실행을 가로챌 수 있다. (gRPC를 지원하는 모든 언어에서 인터셉터가 지원되는 것은 아니며, 각 언어별로 인터셉터의 구현이 다를 수 있다.)

인터셉터는 가로채는 RPC 호출 타입에 따라 두가지 유형으로 분류되는데 **단순 RPC의 경우 단일 인터셉터를 사용할 수 있지만 스트리밍 RPC의 경우 스트리밍 인터셉터를 사용해야 한다.**

go언어 기준으로 gRPC 인터셉터를 지원하는 github repo도 있다. (https://github.com/grpc-ecosystem/go-grpc-middleware)

<br>

## 서버 측 인터셉터

클라이언트가 gRPC 서비스의 원격 메서드를 호출할 때 **서버에서 인터셉터를 사용해 원격 메서드 실행 전에 공통 로직을 실행할 수 있다.**


gRPC서버에 **하나 이상의 인터셉터를 연결할 수 있다.** 

구조는 다음과 같다.

<img src = https://user-images.githubusercontent.com/74294325/145669681-fb159247-8a6c-473e-9d91-e5c91a09f250.png>

<br>

### 서버 단일 인터셉터

서버에서 gRPC 서비스의 단일 RPC를 가로채려면 **gRPC 서버에서 단일 인터셉터를 구현해야 한다.** 

타입은 `UnaryServerInterceptor`이며 Syntax는 다음과 같다.

```go
//https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md#unary-interceptor-1
func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)
```

위의 Syntax에 따라 **함수를 정의 한 후 gRPC서버를 생성할 때 추가해주면 된다.**

```go
// ex
// UnrayOrderManagementServerInterceptor unray interceptor
func UnrayOrderManagementServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) { // 1
	// 전처리
	logrus.WithContext(ctx).
		Infof("==== [Server Interceptor] method: %s", info.FullMethod)

	// 실행
	resp, err = handler(ctx, req) // 2

    // error handling

	// 후처리
	logrus.WithContext(ctx).
		Infof("Post Proc Message: %+v", resp)

	return resp, err
}

// gRPC 서버
srv := grpc.NewServer(grpc.UnaryInterceptor( // 3
		unray.UnrayOrderManagementServerInterceptor,
	))
```

1.  `UnaryServerInterceptor` 타입에 맞춰서 **단일 RPC Server Interceptor**를 정의한다.

2. `handler()로 gRPC를 호출하며` 이 함수를 기준으로 위는 **전처리 영역**이고 **아래는** 후처리 영역이다.

3. gRPC 인스턴스를 생성할 때 `UnaryInterceptor()`의 인자로 작성한 **인터셉터**를 넣어주게 된다. 만약 gRPC서버에 여러 개의 인터셉터를 작성하려면 `ChainUnaryInterceptor()`를 이용할 수 있다.


<br>

정리하자면 서버측 단일 인터셉터 구현은 **일반적으로 전처리, RPC 호출, 후처리** 세부분으로 나눌 수 있다. 전처리 단계에서는 **개발자는 RPC의 컨텍스트, RPC Request, 서버 정보와 같이 전달된 인자를 검사해 현재 RPC호출에 대한 정보를 얻을 수 있다.**

필요하다면 RPC를 호출한 결과에서 반환되는 `error` 처리 또한 가능하다.

<br>

### 서버 스트리밍 인터셉터

서버 스트리밍 인터셉터는 **gRPC 서버가 처리하는 모든 스트리밍 RPC 호출을 인터셉트 한다.**

스트리밍 인터셉터는 **전처리 단계와 스트림 동작 인터셉트 단계가 있다.**

타입은 `StreamServerInterceptor`이며 Syntax는 다음과 같다.
```go
// https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md#stream-interceptor-1
func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error
```

<br>

단일 인터셉터 처럼 전처리 단계에서 **스트리밍 RPC 호출이 서비스 구현으로 이동하기 전에 인터셉트 할 수 있다.** 

단일 RPC와 다른점은 `grpc.ServerStream` 인터페이스를 구현하는 **래퍼 스트림이라는 인터페이스를 사용해 스트리밍 RPC메세지를 가로챌 수 있다.** 간단하게 이야기하면 기존 `grpc.ServerStream`을 **래퍼 구조체로 한번 감싸는 것이다. (데코레이터 패턴)** 

wrapper 구조체는 다음과 같다.
```go
// WrappedStream wrapping grpc server stream
type wrappedStream struct {
	grpc.ServerStream
}

// RecvMsg inst
func (w *wrappedStream) RecvMsg(m interface{}) error { // 1
	// 받기 전 스트리밍 RPC 메세지에 대한 interceptor
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error { // 2
    // 보내기 전 스트리밍 RPC 메세지에 대한 interceptor
	return w.ServerStream.SendMsg(m)
}
```

1. 스트림 RPC로 수신된 메세지를 처리하기 위한 래퍼의 `RecvMsg`함수를 구현한다.

2. 스트림 RPC로 전송되는 메세지를 처리하기 위한 래퍼의 `SendMsg`함수를 구현한다.

<Br>

이처럼 `wrappedStream`는 내부적으로 `grpc.ServerStream`를 가지고 있다

`grpc.ServerStream` 인터페이스는 `RecvMsg(m interface{}) error`와 `SendMsg(m interface{}) error`를 method로 가지고 있으며(두개의 method 말고도 더 가지고 있다.) **래퍼 구조체가 이를 구현함으로 스트리밍을 이용해 메세지를 보낼 때 기존 method대신 구현된 로직이 호출 된다.**

그 이후 **서버 스트리밍 인터셉터를 Syntax에 맞게 정의한다.**

```go
// ServerStreamOrderManagementInterceptor server stream interceptor
func ServerStreamOrderManagementInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error { // 1
	// 전처리
	logrus.WithContext(ss.Context()).Infof("==== server stream interceptor method: %s ====", info.FullMethod)

	err := handler(srv, NewWrappedStream(ss)) // 2
	// 후처리
	if err != nil {
		logrus.WithContext(ss.Context()).WithError(err).Error("failed RPC")
		return err
	}

	return nil
}

// gRPC 서버
srv := grpc.NewServer(grpc.StreamInterceptor( // 3
		interceptor.ServerStreamOrderManagementInterceptor,
	))
```

1. 서버 스트리밍 인터셉터의 **Syntax**에 맞게 함수를 정의한다.

2. 기존 `grpc.ServerStream`를 위에서 정의한 **래퍼 구조체로 감싸준 후 RPC 호출. (스트리밍 RPC를 인터셉터 할 수 있게 된다.)**

3. gRPC 인스턴스를 생성할 때 `StreamInterceptor()`의 인자로 작성한 **인터셉터**를 넣어주게 된다. 서버 스트리밍 또한 동일하게 여러개의 인터셉터를 등록하려면 `ChainStreamInterceptor()`를 이용할 수 있다.

<br>


### 서버 측 정리

단일 요청에 대한 인터셉터는 서버로 들어오는 **RPC요청을** 인터셉터하여

**전처리 -> RPC 호출 -> 후처리** 순으로 진행이 된다.

스트리밍으로 들어오는 요청에 대한 인터셉터는 

**전처리  -> [ Wrapper 구조체에서 구현한 `Recv(interface{}) error` 처리 -> RPC 호출 -> Wrapper 구조체에서 구현한 `Send(interface{}) error` 처리 ] -> 후처리** 순으로 진행된다.

전처리와 후처리는 1번만 실행 되고 스트림이 종료될 때 까지는 래핑 인터페이스의 method가 동작한다.

<br>

## 클라이언트 측 인터셉터

클라이언트 측 인터셉터는 **서버 측 인터셉터와 대부분이 유사하다.** 다른 점은 **인터셉터의 타입, Syntax, 등록방법이 다르다는 것 이외는 동일하다.**

클라이언트 애필리케이션 코드 외부에서 **gRPC 서비스를 안전하게 호출하는 재사용 가능한 특정 기능을 구현해야 할 때 유용하게 사용된다.**

구조는 다음과 같다.

<img src = https://user-images.githubusercontent.com/74294325/145670930-1df78d8c-5a3d-40fb-b948-e7a80e9b7323.png>

<br>

### 클라이언트 단일 인터셉터

타입은 `UnaryClientInterceptor`이며 Syntax는 다음과 같다.

```go
// https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md#unary-interceptor
func(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, invoker UnaryInvoker, opts ...CallOption) error
```

<br>

서버 단일 인터셉터는 `handler()`를 호출하여 RPC를 처리하지만 클라이언트 단일 인터셉터는 `invoker()`를 호출하여 처리한다.

```go
// ex
// UnrayOrderManagementClientInterceptor unray client interceptor
func UnrayOrderManagementClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error { // 1
	// 전처리
	logrus.WithContext(ctx).Infof("==== client unray interceptor method: %s req: %T ====", method, req)

	// 실행
	err := invoker(ctx, method, req, reply, cc, opts...) // 2

	// 후처리
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed Reqeust RPC met")
		return err
	}

	return nil
}

// gRPC client 
conn, err := grpc.Dial( // 3
    serverURL,
    grpc.WithInsecure(),
    grpc.WithUnaryInterceptor(interceptor.UnrayOrderManagementClientInterceptor),
)
```

1.  `UnaryClientInterceptor` 타입에 맞춰서 **단일 RPC Client Interceptor**를 정의한다.

2. `invoker()로 gRPC를 호출하며` 이 함수를 기준으로 위는 **전처리 영역**이고 **아래는** 후처리 영역이다.

3. gRPC 클라이언트 인스턴스를 생성하기 위해 `Dial()` 을 할 때 `WithUnaryInterceptor()`의 인자로 작성한 **인터셉터**를 넣어주게 된다. 만약 gRPC 클라이언트에게 여러 개의 인터셉터를 작성하려면 `WithChainUnaryInterceptor()`를 이용할 수 있다.

<Br>

정리하자면 클라이언트 단일 인터셉터 또한 **전처리, RPC 호출, 후처리** 세부분으로 나눌 수 있다. 전처리 단계에서는 **RPC 컨텍스트, 메세지 문자열, 전송요청, CallOpton등의 호출에 대한 정보에 접근을 할 수 있다.**

전송되기 전 원래 **RPC호출을 수정하는 것도 가능하다.**

<br>

## 클라이언트 스트리밍 인터셉터

클라이언트 스트리밍 인터셉터는 서버 스트리밍 인터셉터 처럼 **gRPC 클라이언트가 처리하는 모든 스트리밍 RPC호출을 인터셉트 한다.**

타입은 `StreamClientInterceptor`이며 Syntax는 다음과 같다.
```go
func(ctx context.Context, desc *StreamDesc, cc *ClientConn, method string, streamer Streamer, opts ...CallOption) (ClientStream, error)
```

클라이언트 스트리밍 인터셉터를 **래핑할 구조체와 클라이언트 스트림에 대한 인터페이스를 구현한다.**

`grpc.ClientStream`를 감싸고 있는 구조체를 정의 한 후 **해당 인터페이스의 method를 구현한다.**

```go
type wrappedStream struct {
	grpc.ClientStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error { // 1
	logrus.Infof("==== client stream interceptor wrapped Recv Message Type: %T====", m)
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error { // 2
	logrus.Infof("==== client stream interceptor wrapped Send Message Type: %T====", m)
	return w.ClientStream.SendMsg(m)
}
```

1. 스트림 RPC로 수신된 메세지를 처리하기 위한 래퍼의 `RecvMsg`함수를 구현한다.

2. 스트림 RPC로 전송되는 메세지를 처리하기 위한 래퍼의 `SendMsg`함수를 구현한다.

<br>

그 이후 **클라이언트 스트리밍 인터셉터를 Syntax에 맞게 정의한다.**

```go
// ClientStreamOrderManagementInterceptor client stream interceptor
func ClientStreamOrderManagementInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) { // 1
	// 전처리
	logrus.WithContext(ctx).Infof("==== client stream interceptor method: %s ====", method)

	stream, err := streamer(ctx, desc, cc, method, opts...) // 2
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed get client stream")
		return nil, err
	}

	// 후처리
	return NewWrappedStream(stream), nil // 3
}

// gRPC client
conn, err := grpc.Dial( // 4
    serverURL,
    grpc.WithInsecure(),
    grpc.WithStreamInterceptor(interceptor.ClientStreamOrderManagementInterceptor))
if err != nil {
    logrus.WithContext(ctx).WithError(err).Fatal("failed get gRPC connection")
}
```

1. 클라이언트 스트리밍 인터셉터의 **Syntax**에 맞게 함수를 정의한다.

2. `streamer()`를 이용하여 **RPC 요청**

3. **래퍼 구조체로 스트림을 감싸줌으로 클라이언트가 서버로 보내는 메세지와 클라이언트로 들어오는 메세지를 가로챈다.**.

4. gRPC 클라이언트 인스턴스를 생성하기 위해 `Dial()` 을 할 때 `WithStreamInterceptor()`의 인자로 작성한 **인터셉터**를 넣어주게 된다. 만약 gRPC 클라이언트에게 여러 개의 인터셉터를 작성하려면 `WithChainWithStreamInterceptor`를 이용할 수 있다.

<br>

### 클라이언트 측 정리

단일 요청에 대한 인터셉터는 클라이언트가 서버로 RPC 요청을 보낼 때

**전처리 -> RPC 요청 -> 후처리** 순으로 진행이 된다.

스트리밍으로 서버에게 요청을 보낼 때에는

**전처리  -> [ Wrapper 구조체에서 구현한 `Send(interface{}) error` 처리 -> RPC 요청 -> Wrapper 구조체에서 구현한 `Recv(interface{}) error` 처리 ] -> 후처리** 순으로 진행된다.

전처리와 후처리는 1번만 실행 되고 스트림이 종료될 때 까지는 래핑 인터페이스의 method가 동작한다.

<br>

## REFERENCE

https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md

https://yeongcheon.github.io/posts/2020-05-30-grpc-interceptor/