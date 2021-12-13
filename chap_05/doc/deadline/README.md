gRPC 고급기능 (DeadLine)
===

데드라인과 타임아웃은 **분산 컴퓨팅에서 일반적으로 사용되는 패턴이다.**

데드라인과 타임아웃은 **클라이언트 애플리케이션이 RPC가 완료 될 때 까지 에러로 종료되기 전 얼마의 시간 동안 기다릴지를 지정한다.**

하나의 요청이 하나 이상의 서비스를 함께 묶는 여러 **다운 스트림 RPC**로 구성되는 될 때, 이 경우 각 서비스 호출마다 **개별 RPC**를 기준으로 **데드라인 혹은 타임아웃을 적용할 수 있지만 요청 전체 수명주기에는 직접 적용할 수 없다.** 

데드라인은 요청 시작 기준으로 특정 시간으로 표현되며 여러 서비스 호출에 걸쳐 적용된다. **요청을 시작하는 애플리케이션이 데드라인을 설정하면 전체 요청 체인은 데드라인까지 응답을 해야 한다.**

gRPC API에서도 RPC 데드라인을 지원하는데 **여러가지 이유로 항상 gRPC 애플리케이션에서 데드라인을 지정하는 것이 바람직하다.** 

데드라인을 사용하지 않고 클라이언트 애플리케이션을 개발하면 **시작된 RPC요청에 대한 응답을 무한정 기다리며, 진행 중인 요청에 대해 리소스가 계속 유지된다.**

이로 인해 **리소스가 부족해 줄 수 있으므로 서비스 대기 시간이 길어지게 되며 결국 전체 gRPC 서비스가 중단될 수도 있다.**

<img src = https://user-images.githubusercontent.com/74294325/145820715-fa6728c2-3f0c-4e12-9f26-d56fd5a38028.png>

<br>

클라이언트 애플리케이션은 **gRPC 서비스를 처음 연결 할 때 데드라인을 설정한다.** 

위의 사진을 보면 클라이언트가 **데드라인을 50ms로 설정 후 요청을 보낸다.** service1이 작업을 하는데 20ms를 소요하고 service2는 30ms를 소요한다. 이렇게 되면 **클라이언트가 설정한 데드라인을 초가하게 되는데 이렇게 데드라인이 넘어가게 되면 클라이언트는 `DEADLINE_EXCEEDED`에러와 함께 해당 RPC 호출이 종료된다.**

<br>

## 데드라인 적용해보기

golang기준으로 `Context` 패키지를 이용하여 DeadLine을 설정 할 수 있다. 이전에 작성해 둔 예제코드에 데드라인을 적용하면 다음과 같다.

```go
conn, err := grpc.Dial(
		serverURL,
		grpc.WithInsecure(),
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

ctx, cancel := context.WithDeadline(context.Bacground(), time.Now().Add(time.Microsecond*30)) // 1
defer cancel()

ID, err := client.AddOrder(ctx, order) // 3
if err != nil {
    logrus.WithContext(ctx).WithError(err).Errorf("failed AddOrder: %+v", order)
    return
}
```

1. context의 **데드라인을 30ms로 설정한다.**

2. 생성된 context를 gRPC 클라이언트 스텁의 method를 호출할 때 **인자로 넣어준다.**

<Br>

만약 context가 **데드라인이 지나게 되면 아래와 같은 error를 받게된다.**
```zsh
failed Reqeust RPC met error="rpc error: code = DeadlineExceeded desc = context deadline exceeded"
```
<br>

context 패키지의 `withTimeOut()` API를 이용하여도 지정이 가능하다. 

`withTimeOut()`와 `WithDeadline()`의 차이점은 두번 째 인자로 `time.Duration`을 받거나 `time.Time을` 받는 것이다. 

`withTimeOut()`를 이용하여 context가 타임아웃 되도 **error의 내용은 동일하다.**
```zsh
failed Reqeust RPC met error="rpc error: code = DeadlineExceeded desc = context deadline exceeded"
```

<Br>

## 데드라인의 이상적인 값은?

데드라인의 이상적인 값을 결정할 때 몇 가지 요소를 고려할 수 있다. 

1. 개별 서비스의 엔트 투 엔드 지연시간

2. RPC가 직렬화 되는지

3. 병렬로 호출 될 수 있는지
   
4. 기본 네트워크의 지연시간과 다운 스트림 서비스의 데드라인 값

<Br>

## 서버 측에서의 데드라인 핸들링

gRPC 데드라인과 관련해 **클라이언트와 서버 모두 RPC의 성공여부에 대해 독립적이고 개별적인 결정을 내릴 수 있다.**

예를 들어 클라이언트가 요청을 보내고 서버에서 작업을 진행하다가 `DEADLINE_EXCEEDE`가 되었고 **ctx의 데드라인을 체크하지 않는 다면 그 시점  이후에 서비스의 로직은 여전이 요청에 대한 응답을 시도할 수도 있다.**

그렇기 때문에 서버측에서도 요청의 **데드라인이 초가되었는지 확인하는 로직이 필요하다.** golang 기준으로 `ctx.Err() == context.DeadlineExceeded`를 확인하여 클라이언트가 이미 **데드라인 초과 상태인지를 확인 후 서버에서 RPC를 더 이상 진행하지 않고 에러를 반환한다.**

대부분 golang에서는 블로킹 되지 않는 `select`구문을 이용하여 처리하게 된다.