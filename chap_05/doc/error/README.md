gRPC 고급기능 (error 처리)
===

gRPC를 호출하면 **클라이언트는 성공 상태의 응답을 받거나** 에러 상태를 갖는 에러를 받는다.

클라이언트 애플리케이션은 **발생 가능한 모든 에러와 에러 상태를 처리하는 방식으로 작성해야 한다.** 또한 서버 애플리케이션도 **에러를 처리하고 해당 상태 코드로 적절한 에러를 생성해야 한다.**

에러가 발생하면 gRPC는 에러 상태의 자세한 정보를 제공하는 **선택적 에러 메세지와 함께 에러 상태 코드를 반환한다.** 상태 객체는 **다른 언어에 대한 모든 gRPC 구현에 공통적인 정수 코드와 문자열 메세지로 구성되어 있다.**

<br>

## 상태 표

코드 | 숫자 | 설명 
:-- | :-- | :--
OK | 0 | 성공 상태
CANCELLED | 1 | 처리가 취소됨 (호출자에 의해)
UNKNOWN | 2 | 알 수 없는 에러
INVALID_ARGUMENT | 3 | 클라이언트에 의해 유효하지 않은 인수 지정
DEADLINE_EXCEEDED | 4 | 처리 완료 전 데드라인 만료
NOT_FOUND | 5 | 일부 요청 엔티티를 찾을 수 없음
PERMISSION_DENIED | 7 | 호출자가 지정한 처리의 실행 권한 없음
UNAUTHENTICATED | 16 | 요청 처리에 대한 유효한 인증 자격증명 없음
INTERNAL | 13 | 내부 에러

<br>

위의 표에는 자주 사용되는 상태 코드와 설명을 적어놓았다. 다양한 상태 코드를 확인하려면 [googleapis grpc상태](https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto)를 참조하면 좋을 것 같다.

<br>

## 본격적인 error 처리

gRPC오 함께 제공되는 에러모델은 **기본적으로 gRPC 데이터 형식과 무관하며 매우 제한적이다.**

프로토콜 버퍼를 데이터 형식으로 사용하년 경우 **google.rpc** 패키지의 **Google API**가 제공하는 더 풍부한 에러모델을 활용할 수 있다. (하지만 이 에러모델은 **C++, GO, Java, 파이썬, 루비**라이브러리에서만 지원되므로 다른 언어를 사용할 때는 유의해야한다.)

예제코드를 보면서 살펴보자면 클라이언트에게 OrderID를 받아 해당하는 Order를 return해주는 서비스가 있다고 할 때 ID값은 **음수가 될 수 없기 때문에 이 경우에 대한 error처리이다.**

```go
//server
...
// GetOrder id가 1보다 작은 경우
if ID.GetId() < "1" {
    logrus.WithContext(ctx).Error("Order ID is Invalid Received OrderID: %s", ID.GetId())

    errStatus := status.New(codes.InvalidArgument, "Invalid information received") // 1
    ds, err := errStatus.WithDetails( // 2
        &epb.BadRequest_FieldViolation{
            Field: "ID",
            Description: fmt.Sprintf(
                "Order ID received is not valid: %s",
                ID.GetId(),
            ),
        },
    )
    if err != nil {
        return nil, errStatus.Err() // 3
    }

    return nil, ds.Err() // 4
}
```

1. `google.golang.org/grpc` 패키지를 이용하여 에러코드 `InvalidArgument`를 사용하여 새로운 에러 상태를 만든다.

2. `google.golang.org/genproto/googleapis/rpc/errdetails`를 import받아 에러 타입인 `BadRequest_FieldViolation`와 함께 에러의 세부사항을 포함한다.

3. 만약 디테일을 포함한 에러를 생성함에 있어 에러가 발생한다면 **에러 상태가 가지고 있는 에러를 반환한다.**

4. 디테일을 포함한 **에러를 반환한다.**

<br>

클라이언트에서는 RPC호출로 인해 **반환된 에러를 처리하면 된다.**

```go
// client
selectedOrder, err := client.GetOrder(ctx, ID)
if err != nil { // 1
    errCode := status.Code(err) // 2
    if errCode == codes.InvalidArgument { // 3
        logrus.WithContext(ctx).WithError(err).Error("failed GetOrder")

        errorStatus := status.Convert(err) // 4
        for _, detail := range errorStatus.Details() {
            switch info := detail.(type) {
            case *epb.BadRequest_FieldViolation: // 5
                logrus.WithContext(ctx).Errorf("Reqeust Field Invalid: %s", info)
            default:
                logrus.WithContext(ctx).Errorf("Unexpected Error type:%s", info)
            }
        }
    } else {
        logrus.WithContext(ctx).WithError(err).Error("unhandled error with errorCode: %s", errCode)
    }
}
```

1. RPC의 요청으로 온 응답중에 error가 **nil인지 check한다.**

2. `google.golang.org/grpc/status` 패키지를 통해 error에서 **에러 코드를 얻는다.**

3. 서버에서 보낸 **에러 코드가 맞는지 확인한다.**

4. `google.golang.org/grpc/status` 패키지를 통해 error에서 **에러의 상태를 얻어낸다.**

5. 에러 상태에서 Detail을 얻어오면 **자료형은 배열이 된다.** 얻어온 배열을 순회하며 **서버 측에서 지정한 에러타입인지 확인하여 에러를 처리하게 된다.**

<br>

이 처럼 gRPC 애플리케이션에 가능하면 적절한 **gRPC 에러코드와 더 풍부한 error 모델을 사용하는 것이 좋다.**


