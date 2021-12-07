gRPC 통신 패턴
===

gRPC 통신패턴에는 4가지 패턴이 존재한다. 패턴은 다음과 같다.

1. 단순 RPC
2. 서버 스트리밍
3. 클라이언트 스트리밍
4. 양방향 스트리밍

이번 chap_03에서는 gRPC IDL을 사용해 서비스를 정의하고 구현하여 각 통신 패턴을 살펴본다.
(공부를 위해 각 패턴마다 IDL을 정의하고 서비스 인터페이스를 구현 후 server와 client를 작성하였다.)

통신 패턴을 적용할 때에는 **비지니스 적인 사례를 분석한 후 적합한 통신 패턴을 선택하는 것이 바람직 하다.**

<br>

## 단순 RPC

단순 RPC는 1장, 2장의 내용과 동일하다. **클라이언트는 단일 요청을 서버에 보내고 서버는 상태에 대한 세부 정보 및 후행 메타데이터와 함께 단일 응답을 한다.**

흐름은 다음과 같다. (code는 1장 2장과 크게 다른 것이 없다.)

<img src =https://user-images.githubusercontent.com/74294325/145038124-700f35f8-5fc5-4a63-8f37-bfb8b1e88b08.png>

<Br>

## 서버 스트리밍

단순 RPC에서는 **항상 단일 요청과 단일 응답을 갖는다.**

서버 스트리밍 RPC에서는 **서버가 클라이언트의 요청 메세지를 받은 후 일련의 응답을 다시 보낸다.**

흐름은 다음과 같다.

<img src = https://user-images.githubusercontent.com/74294325/145038762-10915d5c-81dc-454d-8b54-f053a3365fc1.png>

<Br>

서비스 인터페이스를 정의할 때 서버 스트리밍을 이용하려면 **return 되는 message 앞에 `stream`을 명시해주면 된다.**

```proto
rpc method명 (request) returns (stream response)

// ex
rpc searchOrder (google.protobuf.StringValue) returns (stream Order);
```

<br>

먼저 서비스 인터페이스를 구현하는 서버쪽 코드는 다음과 같다.
```go
// SearchOrder search Order by item Name (with server stream)
func (o *OrderManagementServerImpl) SearchOrder(key *wrapperspb.StringValue, stream pb.OrderManagement_SearchOrderServer) error { // 1
	// database에 order들에서
	for _, order := range o.db {
		// key에 해당하는 order를 찾아 send
		for _, item := range order.Items {
			if strings.Contains(item, key.GetValue()) {
				err := stream.Send(order) // 2
				if err != nil {
					logrus.WithContext(stream.Context()).WithError(err).Error("failed Send Order")
					return err
				}
                // error Handling
				break
			}
		}
	}

	// nil이 return 되면 client 쪽에서는 EOF를 받게 된다.
	return nil // 3
}
```

<Br>

1. 해당 method는 파라미터로 `stream pb.OrderManagement_SearchOrderServer`를 받게 된다. 이 stream을 통해 client에게 **단일 요청이 아닌 여러개의 응답을 보낼 수 있다.**

2. stream 인스턴스의 `Send` API를 이용하여 **Client에게 응답하게 된다.** 

3. return nil을 하게 되면 Client에서는 **EOF**를 받게 된다. 즉 **스트림의 끝을 의미한다.**

<br>

이제 여러 응답을 받는 client의 code는 다음과 같다.

```go
serverStream, err := client.SearchOrder(ctx, wrapperspb.String(targetItem)) // 1
	if err != nil {
		// error Handling
		return
	}

	for {
		order, err := serverStream.Recv() // 2
		if err != nil {
			if err == io.EOF { // 3
				logrus.WithContext(serverStream.Context()).Info("server stream END")
				break
			}
			// error Handling
			return
		}

		logrus.WithContext(serverStream.Context()).Infof("selected Order: %+v", order)
	}
```

<Br>

1. client 스텁에서 method를 request Message와 함께 호출하여 **server stream 인스턴스를 받는다.**

2. 반복문을 돌면서 `Recv()` method를 사용하여 **서버 측 응답을 받는다.** 

3. 서버 스트림의 끝(EOF)을 **check** 하여 반복문을 탈출한다.

<Br>

## 클라이언트 스트리밍

클라이언트 스트림은 서버 스트리밍과 반대로 **클라이언트가 하나의 요청이 아닌 여러 메세지를 보내고 서버는 클라이언트에게 단일 응답을 한다.**

**그러나** 서버는 클라이언트에서 모든 메세지를 수신해 응답을 보낼 필요는 없다. 

**필요의 로직에 따라 하나 또는 여러개의 메세지를 읽은 후 또는 모든 메세지를 읽은 후 응답을 보낼 수 있다.**

흐름은 다음과 같다.

<img src =https://user-images.githubusercontent.com/74294325/145041272-50ab96ec-d254-4ba4-98ec-88694f6d73b5.png>


<Br>

서비스 인터페이스를 정의할 때 클라이언트 스트리밍을 이용하려면 **요청 되는 message 앞에 `stream`을 명시해주면 된다.**

```proto
rpc method명 (stream request) returns (response)

// ex
rpc updateOrder (stream Order) returns (google.protobuf.StringValue);
```

<br>

서비스 인터페이스를 구현하는 서버쪽 코드는 다음과 같다. (현재는 클라이언트의 모든 요청을 받은 후 응답한다.)

```go
func (o *OrderManagementServerImpl) UpdateOrder(stream pb.OrderManagement_UpdateOrderServer) error { // 1
	responseValue := "Updated Order Ids:"

	for {
		order, err := stream.Recv() // 2
		if err != nil {
			if err == io.EOF { // 3
				err = stream.SendAndClose(wrapperspb.String(responseValue)) // 4
				if err != nil {
					// error Handling
					return err
				}
				break
			}

			// error Handling
			return err
		}

		logrus.Infof("Try Update Order with %+v", order)

		o.db[order.GetId()] = order
		responseValue += fmt.Sprintf("%s ", order.GetId())
	}

	return nil
}
```

1. method의 파라미터로 client stream 인스턴스를 받게 된다. `pb.OrderManagement_UpdateOrderServer` 인스턴스를 통해 서버에서 **클라이언트가 보낸 여러개의 요청을 받을 수 있게 된다.**

2. 반복문을 돌며 `Recv()`를 호출하여 Stream에서 **메세지를 얻어온다.**

3. 클라이언트 스트림의 끝(EOF)을 **check** 한다.

4. `SendAndClose()` 를 호출하여 클라이언트 스트림에 대한 **요청 읽기를 종료하고 요청에 대한 응답을 클라이언트에게 보낸다.**

<br>


여러 요청을 받는 client의 code는 다음과 같다. (2개의 요청을 보내고 서버에 대한 응답을 받는다.)

```go
clientStream, err := client.UpdateOrder(ctx) // 1
if err != nil {
    // error Handling
    return
}

err = clientStream.Send(//order message) // 2
if err != nil {
    // error Handling
    return
}
err = clientStream.Send(// order message) // 2
if err != nil {
    // error Handling
    return
}

// server에서 응답이 올 때 까지 blocking
response, err := clientStream.CloseAndRecv() // 3
if err != nil {
    // error Handling
    return
}
```

<br>

1. client 스텁에서 method를 `context`와 호출하여 **클라이언트 스트림 인스턴스를 얻는다.**

2. `Send() API를 이용하여 서버에게 요청을 보낸다.` 단일 요청이 아닌 여러개의 요청을 보낸다.

3. **클라이언트 스트림에 대한 요청 보내기를 종료하고 서버의 응답을 기다린다.** 서버에서 응답이 올 때 까지 blocking 된다.

<br>

## 양방향 스트리밍

양방향 스트리밍 RPC에서는 **클라이언트는 메세지 스트림을 서버에 요청을 보내고 서버는 메세지 스트림으로 응답한다.**

**호출은 클라이언트에서 시작하지만** 그 후 통신은 gRPC 클라이언트와 서버의 애플리케이션 로직에 따라 완전히 다르다.

흐름은 다음과 같다.

<img src =https://user-images.githubusercontent.com/74294325/145046810-a071010f-ea1f-4a5b-aab6-e074b836562d.png>

<br>

서비스 인터페이스를 정의할 때 양방향 스트리밍을 이용하려면 **요청 되는 message 앞에 `stream`을 명시하고 응답 하는 message 앞에 `stream`을 명시하면 된다.**

```proto
rpc method명 (stream request) returns (stream response)

// ex
rpc processOrder (stream google.protobuf.StringValue) returns (stream CombinedShipment);
```

<Br>

서비스 인터페이스를 구현하는 서버쪽 코드는 다음과 같다.

```go
func (o *OrderManagementServerImpl) ProcessOrder(stream pb.OrderManagement_ProcessOrderServer) error { // 1
    // ...
    for {
            orderID, err := stream.Recv() // 2
            if err != nil {
                // eof check
                if err == io.EOF { // 3
                    // client의 stream이 종료되었기 때문에 map에 있는 모든 것을 배송
                    for _, comb := range combinedShipmentMap {
                        err = stream.Send(comb) // 4
                        if err != nil {
                            // error Handling
                            return err
                        }
                    }
                    break
                }

                // error handling
                return err
            }

            // 목적지를 구분하는 로직

            // batch 크기가 가득 찼는지?
            if orderBatchSize == batchSize {
                for _, comb := range combinedShipmentMap {
                    if err = stream.Send(comb); err != nil { // 4
                        // error handling
                        return err
                    }

                    // reset batch size && map
                    batchSize = 0
                    combinedShipmentMap = make(map[string]*pb.CombinedShipment)
                }
            } else {
                batchSize++
            }
        }
    return nil
}
```

<br>

1. method의 파라미터로 양방향 스트리밍 인스턴스를 받게 된다. `pb.OrderManagement_ProcessOrderServer` 인스턴스를 통해 **클라이언트 메세지 스트림에서 요청을 읽어드리고 서버 메세지 스트림을 통해 클라이언트에게 응답을 보낼 수 있다.**

2. 클라이언트 스트리밍과 동일하게 반복문 안에서 `Recv()`를 호출하여 **클라이언트 메세지 스트림에서 요청을 읽어드린다.**

3. 클라이언트 스트림의 끝(EOF)를 **check**한다.

4. 클라이언트에게 응답을 할 때는 서버 스트리밍과 동일하게 `Send()`를 호출하여 **클라이언트에게 응답을 보낼 수 있다.**

<br>

양방향 통신을 하는 client의 code는 다음과 같다.

```go
// two way
twoWayStream, err := client.ProcessOrder(ctx) // 1
if err != nil {
    // error Handling
}

err = twoWayStream.Send(wrapperspb.String("102")) // 2
if err != nil {
    // error Handling
}
err = twoWayStream.Send(wrapperspb.String("103")) // 2
if err != nil {
    // error Handling
}
err = twoWayStream.Send(wrapperspb.String("104")) // 2
if err != nil {
    // error Handling
}
<-time.After(time.Second * 3)
channel := make(chan struct{})
go asyncClientBidrectionalRPC(twoWayStream, channel) // 3

err = twoWayStream.Send(wrapperspb.String("105"))
if err != nil {
    // error Handling
}

if err := twoWayStream.CloseSend(); err != nil { // 4
    log.Fatal(err)
}

<-channel

func asyncClientBidrectionalRPC(twoWayStrem pb.OrderManagement_ProcessOrderClient, channel chan struct{}) { // 5
	defer close(channel)
	for {
		comb, err := twoWayStrem.Recv() // 6
		if err != nil {
			if err == io.EOF { // 7
				logrus.WithContext(twoWayStrem.Context()).Info("Server Stream END")
				break
			}
			logrus.WithContext(twoWayStrem.Context()).WithError(err).Error("failed Recv where stream")
			return
		}

		logrus.WithContext(twoWayStrem.Context()).Infof("Combined shipment: %+v", comb)
	}
}
```

<Br>

1. 1. client 스텁에서 method를 `context`와 호출하여 **양방향 스트림 인스턴스를 얻는다.**

2. 클라이언스 스트리밍과 동일하게 `Send()` API를 호출하여 **클라이언스 메세지 스트림을 통해 서버로 요청을 보낸다.**

3. 양방향 스트리밍을 진행할 때 **고루틴을 이용하여 서버로 부터 오는 응답을 다른 병렬로 읽어 낸다.**

4. **클라이언트 메세지 스트림의 종료를 처리한다.**

5. 서버의 응답을 병렬로 처리 하기 위한 function

6. **양방향 스트림을 통해** 서버 메세지 스트림에서 응답을 읽어낸다.

7. 서버 메세지 스트림의 끝(EOF)를 `check`한다.

<br>

위의 코드 처럼 **양방향 스트리밍** 통신 패턴을 이용할 때 **고루틴을 이용하여 동시에 읽고 쓸 수 있으며 수신 스트림과 발신 스트림 모두 독립적으로 동작한다.**
