Proto 작성
===

## 예제 Proto

```proto
syntax = "proto3"; //1

package ecommerce; //2

service ProductionInfo { //3
    rpc addProduct (Product) returns (ProductID); //4
    rpc getProduct (ProductID) returns (Product);
}

message Product{ //5
   string id = 1;
   string name = 2;
   string description = 3;
}

message ProductID{
    string value = 1;
}
```

<br>

1. 서비스 정의는 사용하는 프로토콜 버퍼의 **버전(proto3)**지정으로 시작한다.
2. 패키지 이름은 **프로토콜 메세지 타입 사이의 이름 충동을 방지하고자 사용하며** 코드 생성에도 사용한다.
3. gRPC 서비스의 **서비스 인터페이스를 정의한다.**
4. 원격 메소드로 클라이언트 측과 서버측이 통신할 때 호출 될 메소드 이다.
5. 원격 메소드의 파라미터로 받는 **메세지**를 정의한 것이며 메세지의 필드에는 **필드를 식별하는데 사용되는 고유 필드 번호가 부여된다.**

<br>

## 서비스 정의 코드 생성 (golang 기준)

서비스 정의 코드를 생성하려면 protoc가 필요합니다. mac OS를 기준으로 아래와 같이 brew를 이용해 설치한다.

<br>

### protobuf 설치

```zsh
brew install protobuf

protoc --version
```
<br>

### plugin설치

golang protobuf plugin을 설치한다. 
>설치 관련 google: https://developers.google.com/protocol-buffers/docs/reference/go-generated

```zsh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

<br>

공식 문서에는 아래와 같이 설명되어 있다.
>This will install a protoc-gen-go binary in $GOBIN. Set the $GOBIN environment variable to change the installation location. It must be in your $PATH for the protocol buffer compiler to find it.

protoc-gen-go 바이너리는 $GOBIN에 설치될 것이다. $GOBIN의 환경변수를 설정하여 install 되는 위치를 변경할 수 있다. protocol buffer 컴파일러가 찾기위에해서는 $PATH에 등록되어 있어야 한다.

<br>

### code 생성

명령어를 입력하여 서비스 정의 코드를 생성한다.
```zsh
protoc --go_out=<pb파일이 export될 경로> --go_opt=paths=source_relative \
--go-grpc_out=<pb파일이 export될 경로> --go-grpc_opt=paths=source_relative \
-I <proto file이 있는 dir 경로> \
<proto file 경로>
```

여기서 `-I` 옵션은 proto file이 존재하는 디렉토리 경로를 지정하는 option이다. 그 이후 **import** 할 proto파일을 지정한다.(여러개 가능하다.)

<Br>

`paths=source_relative` 옵션은 다음과 같이 설명하고 있다.
>If the paths=source_relative flag is specified, the output file is placed in the same relative directory as the input file. For example, an input file protos/buzz.proto results in an output file at protos/buzz.pb.go.
요약하면 `paths=source_relative`를 이용하면 출력파일은 입력파일과 동일한 상대 디렉토리에 배치된다는 것이다.

<br>

이 후 go언어로 코드를 생성하려면 **option go_package**를 강제하고 있다.
>In order to generate Go code, the Go package's import path must be provided for every .proto file

`.proto`에 아래와 같은 형식으로 생성될 서비스 정의 코드에 package를 정의할 수 있다. 경로상 **가장 뒤에 있는 것이 서비스 정의 코드의 package명이 된다.**
```proto
option go_package = <모듈명>/<사용자 정의 package경로>

// ex (module명이 grpc_study일 경우)
option go_package = gprc_study/internal/proto
```
혹은 `M${PROTO_FILE}=${GO_IMPORT_PATH}`와 같이 CLI로 명령어를 입력할 때 package를 지정할 수 있다.

<br>

만약 package를 정의하지 않고 서비스 정의 코드를 생성하려 할 때 발생하는 error는 다음과 같다.
```zsh
Please specify either:
    • a "go_package" option in the .proto source file, or
    • a "M" argument on the command line.
```

<br>

## REFERENCE

https://developers.google.com/protocol-buffers/docs/reference/go-generated

https://developers.google.com/protocol-buffers/docs/proto3


