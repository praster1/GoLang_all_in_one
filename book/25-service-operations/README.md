# 25. 서비스 구성과 운영 기초

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 중급 | [16장 CLI](../16-cli/README.md), [17장 context](../17-context/README.md), [22장 HTTP 서버](../22-http-server/README.md), [24장 SQL](../24-sql-transactions/README.md) | Go, 이 저장소 |

앞 장에서 요청을 처리하고 데이터를 저장하는 코드를 만들었습니다. 이제 같은 코드를 오래 실행되는 서비스로 묶습니다. 서비스에는 요청을 처리하는 시간뿐 아니라 **시작하는 시간과 멈추는 시간**이 있습니다. 설정이 잘못되었는데 포트를 먼저 열거나, 응답을 쓰는 중에 DB를 닫으면 개별 함수가 올바르더라도 전체 프로그램은 기대와 다르게 동작합니다.

이 장은 설정, 로그, 상태 확인, 인증 경계와 종료를 작은 프로그램으로 나누어 실습합니다. 처음부터 거대한 서버에 모두 넣지 않고 각각의 계약을 관찰한 다음, 하나의 실행 흐름으로 연결합니다. 30장에서는 이 구조 위에 메모리 저장소와 SQL 저장소를 교체하는 할 일 API를 완성합니다.

## 학습 목표

- 기본값과 외부 설정을 합친 뒤 시작 전에 검증한다.
- 구조화 로그에 필요한 정보만 남기고 비밀값의 출력 경로를 제한한다.
- 요청 결과, 프로세스 생존 상태와 요청 수용 가능 상태를 구분한다.
- 토큰 확인의 범위와 실제 사용자 권한 관리의 차이를 설명한다.
- 요청 수신을 멈추고 진행 중인 작업을 기다린 뒤 자원을 정리한다.

## 사전 지식과 실행 방법

모든 명령은 **저장소 루트**에서 실행합니다. 이번 장은 표준 라이브러리만 사용하며 실제 계정과 외부 서비스가 필요하지 않습니다. 예제에 들어 있는 토큰 문자열은 동작을 설명하는 공개된 학습 값입니다. 개인 서비스의 비밀값을 코드에 넣지 않습니다.

| 실행 예제 | 관찰할 내용 |
| --- | --- |
| [configuration](../../examples/25-service-operations/configuration/main.go) | 기본값, 존재 여부, 파싱, 범위 검증 |
| [structured-log](../../examples/25-service-operations/structured-log/main.go) | JSON 로그, 수준, 공통 속성 |
| [secret-value](../../examples/25-service-operations/secret-value/main.go) | `LogValuer`와 허용한 필드만 기록하기 |
| [request-log](../../examples/25-service-operations/request-log/main.go) | 응답 상태와 경로 패턴 기록 |
| [health-readiness](../../examples/25-service-operations/health-readiness/main.go) | 생존 확인과 준비 상태 |
| [bearer-auth](../../examples/25-service-operations/bearer-auth/main.go) | 요청 앞에서 자격 증명 확인 |
| [shutdown](../../examples/25-service-operations/shutdown/main.go) | 실제 소켓을 열고 정상 종료 |
| [service 패키지](../../examples/25-service-operations/service/lifecycle.go) | 취소 신호와 서버 수명 연결 |

`shutdown`만 loopback 임시 포트를 사용합니다. 나머지 HTTP 예제는 `httptest.ResponseRecorder`로 요청을 직접 전달하므로 별도 서버를 실행하지 않습니다. 기본 예제는 모두 자동으로 끝납니다. 직접 요청을 보내는 실행 모드는 연습문제의 `managed-server`에 따로 있습니다.

## 1. 기능과 실행 환경 사이에 경계를 둔다

예를 들어 `GET /lessons/7`이 잘 동작한다고 합시다. 운영 환경에서는 다음 질문도 답할 수 있어야 합니다.

| 질문 | 코드가 결정해야 할 것 |
| --- | --- |
| 어디에서 요청을 받는가? | 주소와 포트, 수신 실패 처리 |
| 얼마나 기다리는가? | 헤더 읽기, 작업, 종료의 서로 다른 제한 시간 |
| 무슨 일이 있었는가? | 요청 ID, 경로 패턴, 결과 상태, 소요 시간 |
| 요청을 받아도 되는가? | 초기화·의존 자원·종료 상태 |
| 누가 기능을 사용할 수 있는가? | 인증과 권한 검사의 위치 |
| 종료 시 진행 중인 작업은 어떻게 되는가? | 신규 수신 중지, 대기, 강제 종료 조건 |

이 결정을 모두 핸들러 내부에 넣으면 요청마다 환경 변수를 다시 읽거나 매번 DB를 새로 열 수 있습니다. 대신 시작 코드가 설정과 의존성을 만들고 핸들러에 전달합니다. 핸들러는 요청의 입력과 결과에 집중하고, 프로그램의 가장 바깥쪽이 자원의 전체 수명을 소유합니다.

### 1.1 권장 시작 순서

1. 플래그와 환경 변수를 읽습니다.
2. 숫자·주소·시간 값을 파싱하고 허용 범위를 검증합니다.
3. 로그 출력과 사용할 자원을 준비합니다.
4. 저장소와 핸들러를 연결합니다.
5. 포트를 열고 요청을 받기 시작합니다.
6. 종료 요청이 오면 준비 상태를 내리고 서버를 종료합니다.
7. 요청을 기다린 뒤 저장소 같은 공유 자원을 닫습니다.

이 순서는 이 교재의 설계 선택입니다. 모든 서비스에 하나의 디렉터리 구조가 강제되는 것은 아닙니다. 다만 설정 오류와 요청 오류, 시작 실패와 종료 실패를 구분하면 각 실패를 재현하기 쉬워집니다.

### 1.2 서로 다른 세 가지 제한 시간

`http.Server.ReadHeaderTimeout`은 요청 헤더를 읽는 시간입니다. `context.WithTimeout`은 전달된 작업이 취소를 관찰할 수 있는 수명입니다. `Shutdown`에 전달하는 context는 종료 대기의 상한입니다. 세 값을 모두 `timeout` 하나로 이름 붙이면 어떤 기다림을 제한하는지 알기 어렵습니다.

`ReadTimeout`, `WriteTimeout`, `IdleTimeout`의 정확한 범위는 [http.Server](https://pkg.go.dev/net/http#Server) 문서를 확인합니다. 특히 쓰기 제한 시간과 핸들러 작업 취소는 같은 기능이 아닙니다. context를 무시하는 작업은 제한 시간이 지났다고 자동으로 강제 중단되지 않습니다.

## 2. 설정을 읽는 함수와 실제 환경을 분리하기

저장 경로: [`examples/25-service-operations/configuration/main.go`](../../examples/25-service-operations/configuration/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address     string
	ReadTimeout time.Duration
	Level       slog.Level
}

func loadConfig(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{Address: "127.0.0.1:8080", ReadTimeout: 5 * time.Second, Level: slog.LevelInfo}
	if value, ok := lookup("APP_ADDR"); ok {
		host, port, err := net.SplitHostPort(value)
		if err != nil || host == "" {
			return Config{}, errors.New("APP_ADDR는 호스트:포트 형식이어야 합니다")
		}
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return Config{}, errors.New("APP_ADDR 포트는 1~65535여야 합니다")
		}
		cfg.Address = value
	}
	if value, ok := lookup("APP_READ_TIMEOUT"); ok {
		d, err := time.ParseDuration(value)
		if err != nil || d <= 0 || d > time.Minute {
			return Config{}, errors.New("APP_READ_TIMEOUT은 0보다 크고 1m 이하여야 합니다")
		}
		cfg.ReadTimeout = d
	}
	if value, ok := lookup("APP_LOG_LEVEL"); ok {
		switch value {
		case "debug":
			cfg.Level = slog.LevelDebug
		case "info":
			cfg.Level = slog.LevelInfo
		case "warn":
			cfg.Level = slog.LevelWarn
		case "error":
			cfg.Level = slog.LevelError
		default:
			return Config{}, errors.New("APP_LOG_LEVEL은 debug/info/warn/error 중 하나여야 합니다")
		}
	}
	return cfg, nil
}

func run() error {
	env := map[string]string{"APP_ADDR": "127.0.0.1:9090", "APP_READ_TIMEOUT": "2s", "APP_LOG_LEVEL": "warn"}
	cfg, err := loadConfig(func(key string) (string, bool) { value, ok := env[key]; return value, ok })
	if err != nil {
		return err
	}
	fmt.Println("주소:", cfg.Address)
	fmt.Println("읽기 제한:", cfg.ReadTimeout)
	fmt.Println("로그 수준:", cfg.Level)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./examples/25-service-operations/configuration
```

실행 결과:

```text
주소: 127.0.0.1:9090
읽기 제한: 2s
로그 수준: WARN
```


### 2.1 외부 값을 곧바로 신뢰하지 않는다

`loadConfig`는 문자열을 읽는 `lookup` 함수를 받습니다. 운영 실행에서는 `os.LookupEnv`를 전달할 수 있고, 테스트에서는 작은 맵을 전달할 수 있습니다. 함수가 실제 프로세스 환경을 바꾸지 않으므로 테스트끼리 환경 변수 값을 두고 충돌하지 않습니다.

예제 `main`은 항상 같은 결과를 볼 수 있도록 맵을 전달합니다. 따라서 셸에서 `APP_ADDR`를 바꿔도 이 예제의 출력은 변하지 않습니다. 실제 환경을 사용하는 한 줄은 다음과 같습니다. **발췌**이며 기존 `run`에서 설정을 만드는 부분을 바꿀 때 사용합니다.

```go
cfg, err := loadConfig(os.LookupEnv)
```

환경 변수가 없는 경우와 빈 문자열로 존재하는 경우를 구분하려면 [os.LookupEnv](https://pkg.go.dev/os#LookupEnv)가 적합합니다. `os.Getenv`만 사용하면 두 경우 모두 빈 문자열로 보입니다.

| 입력 상태 | 이번 예제의 정책 |
| --- | --- |
| 변수가 없음 | 기본값 사용 |
| 변수가 있고 값이 유효 | 해당 필드 덮어쓰기 |
| 빈 값이 명시됨 | 설정 오류 |
| 숫자로 변환되지만 범위를 벗어남 | 설정 오류 |
| 알려지지 않은 로그 수준 | 설정 오류 |

빈 문자열을 기본값으로 해석하는 정책도 만들 수 있습니다. 중요한 것은 정책을 설명하고 테스트하는 것입니다. 여기서는 배포 설정에서 잘못 비워 둔 값을 빨리 발견하려고 오류로 처리합니다.

### 2.2 문법 검증과 의미 검증

`time.ParseDuration("0s")`는 파싱에 성공합니다. 하지만 이번 서비스의 읽기 제한으로는 허용하지 않습니다. `strconv.Atoi("65536")`도 정수를 반환하지만 TCP 포트의 허용 범위가 아닙니다. 파싱 성공은 문자열의 문법이 맞다는 의미이고, 서비스가 사용할 수 있는 값인지는 별도로 확인합니다.

주소는 콜론을 직접 `Split`하지 않고 [net.SplitHostPort](https://pkg.go.dev/net#SplitHostPort)로 나눕니다. IPv6 주소의 콜론과 포트 구분 기호를 혼동하지 않기 위해서입니다. `[::1]:8080`처럼 IPv6 호스트에는 대괄호가 필요합니다.

### 2.3 잘못된 설정을 조용히 고치지 않는다

`APP_READ_TIMEOUT=oops`가 들어왔을 때 기본 5초로 돌아가는 코드를 만들면 프로그램은 실행됩니다. 하지만 사용자는 5초로 실행 중이라는 사실을 모르기 쉽습니다. 이 예제는 오류를 반환하여 시작 단계에서 원인을 찾게 합니다.

오류에는 어떤 설정 항목이 잘못되었는지 적되 원래 입력을 무조건 붙이지 않습니다. 주소와 시간처럼 공개 가능한 항목도 있고, 비밀번호나 API 토큰처럼 기록하면 안 되는 항목도 있기 때문입니다. 일관된 규칙은 “전체 설정 구조체를 한 번에 출력하지 않는다”입니다.

### 2.4 바꾸어 볼 값

| 변경 | 예상 결과 |
| --- | --- |
| `APP_READ_TIMEOUT` 키 삭제 | 5초 기본값 |
| 값을 `500ms`로 변경 | 500밀리초 |
| 값을 `0s`로 변경 | 실행 실패 |
| 주소를 `[::1]:9090`으로 변경 | 주소 파싱 성공 |
| 주소를 `:9090`으로 변경 | 이 예제의 호스트 필수 정책으로 실패 |
| 로그 수준을 `DEBUG`로 변경 | 명시한 소문자 목록에 없으므로 실패 |

대소문자 구분은 예제의 정책입니다. 필요하면 문자열을 소문자로 정규화할 수 있지만, 그때는 문서와 테스트도 함께 바꿉니다.

## 3. 구조화 로그로 사건과 속성을 나누기

저장 경로: [`examples/25-service-operations/structured-log/main.go`](../../examples/25-service-operations/structured-log/main.go).

```go
package main

import (
	"log/slog"
	"os"
)

func withoutTime(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 && attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, ReplaceAttr: withoutTime,
	})).With("service", "study-api")
	logger.Debug("설정 읽기")
	logger.Info("요청 완료", slog.Group("request", "method", "GET", "status", 200))
	logger.Warn("의존 서비스 지연", "dependency", "catalog", "retryable", true)
}
```

```sh
go run ./examples/25-service-operations/structured-log
```

실행 결과:

```text
{"level":"INFO","msg":"요청 완료","service":"study-api","request":{"method":"GET","status":200}}
{"level":"WARN","msg":"의존 서비스 지연","service":"study-api","dependency":"catalog","retryable":true}
```


### 3.1 메시지와 속성

`fmt.Println("사용자 7의 요청이 200으로 끝남")`은 사람이 읽기는 쉽지만 자동 집계하려면 문자열을 다시 해석해야 합니다. 구조화 로그(structured logging)는 메시지와 속성을 나눕니다. 예제에서는 메시지가 `요청 완료`, 속성이 `method=GET`, `status=200`입니다.

[slog](https://pkg.go.dev/log/slog)의 `Logger`는 사건을 전달하고 `Handler`는 수준 필터와 출력 형태를 담당합니다. `JSONHandler`를 사용하면 숫자는 숫자로, 논리값은 논리값으로 기록할 수 있습니다. JSON 한 줄이 사건 하나를 나타내도록 만들면 수집 도구에서 각 속성을 다루기 쉽습니다.

`logger.With("service", "study-api")`는 이후 로그마다 공통 속성을 포함하는 새 로거를 만듭니다. 요청마다 같은 문자열을 직접 붙일 필요가 없습니다. 전역 기본 로거를 바꾸지 않고 필요한 함수에 로거를 전달하면 테스트 출력도 분리하기 쉽습니다.

### 3.2 수준은 장식이 아니라 정책이다

| 수준 | 예제에서 사용할 상황 |
| --- | --- |
| DEBUG | 개발 중 세부 흐름, 기본 설정에서는 제외 |
| INFO | 시작·종료, 정상 처리 결과 |
| WARN | 복구 가능하지만 확인할 필요가 있는 상태 |
| ERROR | 작업 실패 또는 대응이 필요한 문제 |

모든 404를 오류 수준으로 남기거나 매번 같은 시작 설정을 반복 기록하면 중요한 사건을 찾기 어려워집니다. 수준은 함수 이름에 맞춰 기계적으로 고르기보다 사건의 의미와 운영 방식에 맞춰 정합니다.

기본 예제의 `DEBUG` 메시지가 출력되지 않는 이유는 `LevelInfo` 이상만 처리하기 때문입니다. `ReplaceAttr`로 시간을 제외한 이유는 교재 출력과 테스트를 재현하기 위해서입니다. 실제 서비스에서 사건의 시각까지 없애라는 뜻이 아닙니다.

### 3.3 로그 자체의 수명

표준 오류에 로그를 기록하면 서비스 실행 환경이 수집하기 쉬운 경우가 많습니다. 파일에 기록한다면 파일을 누가 열고 닫는지, 용량이 계속 늘 때 어떻게 교체할지까지 결정해야 합니다. `slog`를 사용한다는 사실만으로 파일 회전이나 원격 전송이 자동으로 생기지는 않습니다.

이번 장은 직접 만든 원격 로그 전송기를 추가하지 않습니다. 필요한 정보가 구조화되어 있고 오류 경로에서도 로그가 남는지 먼저 확인합니다. 출력 대상의 장애 대응은 사용 중인 실행 환경과 연결해서 별도로 설계할 내용입니다.

## 4. 비밀값을 로그에서 제외하는 타입

저장 경로: [`examples/25-service-operations/secret-value/main.go`](../../examples/25-service-operations/secret-value/main.go).

```go
package main

import (
	"log/slog"
	"os"
)

type Secret string

func (Secret) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }

type Credentials struct {
	Client string
	Token  Secret
}

func (c Credentials) LogValue() slog.Value {
	return slog.GroupValue(slog.String("client", c.Client), slog.Any("token", c.Token))
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	credentials := Credentials{Client: "lesson-client", Token: Secret("classroom-placeholder")}
	logger.Info("클라이언트 설정", "credentials", credentials)
}
```

```sh
go run ./examples/25-service-operations/secret-value
```

실행 결과:

```text
level=INFO msg="클라이언트 설정" credentials.client=lesson-client credentials.token=[REDACTED]
```


### 4.1 값이 로그로 변환되는 경로를 정의한다

`Secret`은 문자열을 바탕으로 만든 타입이지만 `LogValue`를 구현합니다. `slog`가 값을 해석할 때 원문 대신 `[REDACTED]`를 사용하게 합니다. `Credentials`도 `LogValue`를 구현하여 출력할 필드를 직접 선택합니다. 공식 계약은 [slog.LogValuer](https://pkg.go.dev/log/slog#LogValuer)에 있습니다.

이 설계는 구조체에 비밀 필드가 늘어날 때 효과가 있습니다. `fmt`로 전체 구조체를 무심코 찍는 대신, 로그용 표현에서 허용한 필드만 남깁니다. 이번 예제의 `client`는 공개해도 되는 학습용 식별자입니다. 실제 식별자가 개인정보라면 출력 여부를 별도로 결정합니다.

### 4.2 타입이 모든 출력 경로를 막지는 않는다

다음은 **잘못된 사용을 설명하는 발췌**입니다. 원문으로 변환한 뒤 로거에 전달하면 `Secret.LogValue`를 거치지 않습니다.

```go
logger.Info("잘못된 기록", "token", string(credentials.Token))
```

`fmt.Printf`, 문자열 연결, 요청 URL, 오류 메시지에도 값이 들어갈 수 있습니다. 따라서 `LogValuer`는 올바른 사용을 돕는 경계이며 비밀 관리 전체를 대신하는 기능은 아닙니다. 로그 테스트에서 원문이 없는지 확인하는 이유도 여기에 있습니다.

### 4.3 제외할 값과 남길 값을 함께 정한다

| 기록 후보 | 이번 장의 선택 |
| --- | --- |
| 처리 결과 상태 | 기록 |
| 라우트 패턴 | 기록 |
| 사용자 입력 본문 전체 | 기본적으로 기록하지 않음 |
| Authorization 헤더 | 기록하지 않음 |
| 토큰이 포함될 수 있는 전체 URL | 기록하지 않음 |
| 비밀정보가 없는 내부 오류 분류 | 필요하면 기록 |

민감한 내용을 지우는 방식만 고민하면 새 필드를 빠뜨리기 쉽습니다. 필요한 필드만 선택해서 기록하는 방식을 함께 사용합니다. 다음 예제는 URL 대신 라우트 패턴을 남깁니다.

## 5. 요청의 결과를 기록하는 미들웨어

저장 경로: [`examples/25-service-operations/request-log/main.go`](../../examples/25-service-operations/request-log/main.go).

```go
package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

type responseCapture struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseCapture) WriteHeader(status int) {
	if status >= 100 && status < 200 {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseCapture) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}
func (w *responseCapture) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func requestLog(logger *slog.Logger, now func() time.Time, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := now()
		capture := &responseCapture{ResponseWriter: w}
		next.ServeHTTP(capture, r)
		if capture.status == 0 {
			capture.status = http.StatusOK
		}
		logger.InfoContext(r.Context(), "request",
			"method", r.Method, "route", r.Pattern, "status", capture.status,
			"bytes", capture.bytes, "elapsed_ms", now().Sub(start).Milliseconds())
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "ok\n"); err != nil {
			return
		}
	})
	fixed := time.Unix(0, 0)
	handler := requestLog(logger, func() time.Time { return fixed }, mux)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/lessons/7?token=hidden", nil))
	fmt.Println("HTTP 상태:", recorder.Code)
}
```

```sh
go run ./examples/25-service-operations/request-log
```

실행 결과:

```text
{"level":"INFO","msg":"request","method":"GET","route":"GET /lessons/{id}","status":200,"bytes":3,"elapsed_ms":0}
HTTP 상태: 200
```


### 5.1 실행 전과 실행 후를 잇는다

미들웨어는 시작 시각을 읽고 응답 작성자를 감싼 뒤 실제 핸들러를 호출합니다. 핸들러가 돌아오면 상태 코드, 쓴 바이트 수와 소요 시간을 기록합니다. `next.ServeHTTP`가 가운데에 있는 이유는 결과가 정해진 뒤 로그를 남겨야 하기 때문입니다.

순서를 따라 읽어 봅니다.

1. `now()`로 시작 시각을 얻습니다.
2. `responseCapture`에 원래 `ResponseWriter`를 보관합니다.
3. mux가 경로를 찾고 실제 핸들러가 응답을 씁니다.
4. `Write`가 호출되었으면 상태와 실제 기록된 바이트를 모읍니다.
5. 아무것도 쓰지 않았다면 기본 상태 200으로 기록합니다.
6. 원래 URL 대신 mux가 선택한 `r.Pattern`을 로그에 남깁니다.

`now` 함수도 주입하므로 테스트에서 실제 시계를 기다리지 않습니다. 예제의 시간 함수는 같은 값을 두 번 반환하여 `elapsed_ms`가 0입니다. 실제 실행에서는 `time.Now`를 전달합니다.

### 5.2 상태 코드의 첫 결정

`WriteHeader`를 호출하지 않고 `Write`하면 HTTP 응답은 200으로 시작합니다. 이미 최종 상태를 보낸 뒤 500을 다시 설정해도 앞선 상태를 바꿀 수 없습니다. `responseCapture`는 첫 최종 상태를 보관해서 로그가 실제 응답과 다른 값을 주장하지 않게 합니다. [ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)의 계약을 확인합니다.

100~199 정보 응답은 일반 최종 응답과 구분합니다. 이 래퍼는 일반 요청·응답 학습용이며 WebSocket 업그레이드, 직접 연결 제어와 스트리밍을 완전히 추상화하지 않습니다. 지원해야 할 프로토콜이 늘면 추가 인터페이스와 상태 전이를 검토해야 합니다.

### 5.3 패턴과 실제 경로

`/lessons/7?token=hidden` 대신 `GET /lessons/{id}`가 기록됩니다. 토큰 문자열을 실수로 남길 가능성을 줄이고, 항목 번호가 다른 요청도 같은 경로 종류로 묶을 수 있습니다. `r.Pattern`은 mux가 라우트를 고른 뒤에 의미가 있으므로 핸들러 실행 후 읽습니다.

일치하는 라우트가 없으면 패턴이 비어 있을 수 있습니다. 이 경우 상태와 메서드만으로도 기본 분류는 가능합니다. 실제 경로가 필요한 진단에서는 로그용 정규화와 민감한 경로값의 처리 정책을 먼저 정합니다.

### 5.4 래퍼가 숨길 수 있는 기능

원래 `ResponseWriter`가 `http.Flusher`를 구현해도 단순히 구조체에 포함했다고 래퍼가 그 인터페이스를 자동으로 구현하는 것은 아닙니다. 이번 `Unwrap`은 [ResponseController](https://pkg.go.dev/net/http#NewResponseController)가 내부 작성자를 찾을 수 있게 합니다. 다른 코드가 직접 `w.(http.Flusher)`를 검사하는 상황까지 자동으로 해결하지는 않습니다.

또한 이 미들웨어는 정상적으로 돌아온 요청을 기록합니다. 패닉 복구 미들웨어를 추가한다면 호출 순서와 오류 응답 작성 책임을 함께 정해야 합니다. 패닉을 잡았다는 이유로 이미 일부 전송한 응답을 정상 JSON 오류로 바꿀 수는 없습니다.

## 6. 살아 있는가, 요청을 받을 준비가 되었는가

저장 경로: [`examples/25-service-operations/health-readiness/main.go`](../../examples/25-service-operations/health-readiness/main.go).

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

type Health struct {
	Ready atomic.Bool
	Check func(context.Context) error
}

func (h *Health) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !h.Ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
		defer cancel()
		if err := h.Check(ctx); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return mux
}
func main() {
	health := &Health{Check: func(ctx context.Context) error { return ctx.Err() }}
	handler := health.Handler()
	for _, path := range []string{"/livez", "/readyz"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
		fmt.Printf("%s: %d\n", path, recorder.Code)
	}
	health.Ready.Store(true)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/readyz", nil))
	fmt.Println("준비 완료:", recorder.Code)
}
```

```sh
go run ./examples/25-service-operations/health-readiness
```

실행 결과:

```text
/livez: 200
/readyz: 503
준비 완료: 200
```


### 6.1 두 상태의 의미

생존 확인(liveness)은 프로세스가 응답할 수 있는지 보는 용도입니다. 준비 확인(readiness)은 현재 실제 요청을 받아도 되는지 보는 용도입니다. 둘을 같은 검사로 만들면 일시적인 DB 장애 때문에 살아 있는 프로세스를 불필요하게 반복 재시작할 수 있습니다.

이번 예제에서 `/livez`는 핸들러가 응답할 수 있으면 200을 반환합니다. `/readyz`는 `Ready` 상태가 참이고 의존 확인도 성공해야 200입니다. 이는 이 교재의 상태 확인 계약이며, 특정 배포 플랫폼 설정이 자동으로 생긴다는 뜻은 아닙니다.

| 서비스 단계 | livez | readyz |
| --- | --- | --- |
| 핸들러는 동작하지만 초기화 전 | 200 | 503 |
| 초기화 완료, 의존 확인 성공 | 200 | 200 |
| 프로세스는 동작하지만 의존 확인 실패 | 200 | 503 |
| 요청 수용을 중지하기 시작함 | 200 | 503 |

### 6.2 준비 상태를 공유한다

HTTP 서버는 여러 요청을 동시에 처리할 수 있습니다. 한 고루틴이 상태를 쓰고 다른 고루틴이 읽을 때 일반 `bool`만 사용하면 경쟁이 생길 수 있습니다. `atomic.Bool`의 `Store`와 `Load`로 한 값을 공유합니다. 복잡한 여러 필드의 불변 조건까지 자동으로 보호되는 것은 아닙니다.

준비 상태가 거짓일 때는 의존 확인 함수를 호출하지 않습니다. 아직 초기화되지 않은 자원에 접근하지 않게 하기 위해서입니다. `Check`는 구성 시 항상 유효한 함수를 전달하는 계약이며, 실행 중 다른 고루틴에서 임의로 바꾸지 않습니다.

### 6.3 상태 확인도 작업이다

상태 확인이 DB 연결을 기다리다가 영원히 멈추면 확인 기능 자체가 자원을 차지합니다. 예제는 요청 context에서 100밀리초 제한을 만들어 전달합니다. DB를 확인할 때는 이 context를 `PingContext`에 넘길 수 있습니다.

context는 협조적 취소입니다. `Check` 구현이 전달된 context를 무시하고 계속 대기한다면 `WithTimeout`이 함수 실행을 강제로 멈추지 않습니다. 별도의 무제한 고루틴을 만들고 시간만 기다리는 방식은 멈춘 작업을 숨길 수 있으므로 여기서는 사용하지 않습니다.

실제 서비스에서는 의존 확인 빈도와 비용도 고려합니다. 모든 상태 요청마다 비싼 전체 조회를 실행할 필요는 없습니다. 작은 검사를 선택하거나 최근 상태를 관리하는 방식의 정확성과 지연을 비교합니다.

### 6.4 준비 상태를 내린 즉시 모든 요청이 멈추는가

아닙니다. 외부 라우터가 새 상태를 관찰하는 데 시간이 걸릴 수 있습니다. 이미 연결되어 있거나 라우팅 중인 요청도 있습니다. readiness를 내리는 것, 신규 수신을 멈추는 것, 이미 들어온 작업을 기다리는 것은 서로 다른 단계입니다.

연습문제에서는 종료가 시작된 후 새로운 작업을 503으로 거절하는 게이트도 만듭니다. 이미 통과한 요청까지 되돌리는 장치는 아닙니다. 이 구분을 이해한 뒤 서버 종료와 연결합니다.

## 7. 요청 앞에서 토큰 확인하기

저장 경로: [`examples/25-service-operations/bearer-auth/main.go`](../../examples/25-service-operations/bearer-auth/main.go).

```go
package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func bearerAuth(expected string, next http.Handler) (http.Handler, error) {
	if len(expected) < 16 || strings.ContainsAny(expected, " \t\r\n") {
		return nil, errors.New("설정 토큰은 공백 없는 16바이트 이상이어야 합니다")
	}
	want := sha256.Sum256([]byte(expected))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.Header.Values("Authorization")
		if len(values) == 1 && strings.HasPrefix(values[0], "Bearer ") {
			got := sha256.Sum256([]byte(strings.TrimPrefix(values[0], "Bearer ")))
			if subtle.ConstantTimeCompare(got[:], want[:]) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}), nil
}
func run() error {
	token := "classroom-demo-token"
	handler, err := bearerAuth(token, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	if err != nil {
		return err
	}
	for _, credential := range []string{"", "wrong-token", token} {
		request := httptest.NewRequest("POST", "/private", nil)
		if credential != "" {
			request.Header.Set("Authorization", "Bearer "+credential)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		fmt.Println("상태:", recorder.Code)
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./examples/25-service-operations/bearer-auth
```

실행 결과:

```text
상태: 401
상태: 401
상태: 204
```


### 7.1 인증과 권한을 구분한다

인증(authentication)은 요청이 제시한 자격 증명을 확인하는 과정입니다. 권한 확인(authorization)은 확인된 주체가 특정 행동을 해도 되는지 판단하는 과정입니다. 예제는 하나의 학습 토큰이 일치하는지만 확인합니다. 여러 사용자, 역할, 만료, 회수, 토큰 발급을 구현한 시스템이 아닙니다.

토큰이 맞으면 다음 핸들러를 정확히 한 번 호출하고, 틀리면 호출하지 않습니다. 이 경계가 명확해야 인증 실패 요청이 DB 변경 코드까지 도달하지 않습니다. 성공 시 204, 실패 시 401이 나오는지 테스트에서 상태뿐 아니라 다음 핸들러 호출 횟수도 확인합니다.

### 7.2 고정 길이 값 비교

예제는 토큰을 SHA-256으로 변환한 뒤 [subtle.ConstantTimeCompare](https://pkg.go.dev/crypto/subtle#ConstantTimeCompare)로 같은 길이의 바이트 배열을 비교합니다. 이는 비교 과정의 동작을 명확하게 하는 선택입니다. 전체 요청 처리가 길이나 입력에 전혀 의존하지 않는다고 주장하는 것은 아닙니다.

이 해시는 비밀번호 저장 알고리즘을 가르치는 예제가 아닙니다. 이미 정해진 임의의 토큰을 비교하는 상황입니다. 사용자 비밀번호를 SHA-256 한 번으로 저장하는 설계와 혼동하지 않습니다.

### 7.3 헤더와 설정의 경계

설정 토큰이 비어 있으면 서버 구성을 실패시킵니다. 빈 설정을 “인증을 끈다”는 의미로 해석하지 않습니다. 요청의 Authorization 헤더가 여러 개이면 모호한 해석을 피하려고 실패로 처리합니다. 이 학습 예제의 스킴 문자열은 정확히 `Bearer `입니다.

실패 응답에는 `WWW-Authenticate: Bearer`와 짧은 일반 메시지만 넣습니다. 전달된 토큰을 응답하거나 로그에 남기지 않습니다. 자세한 프로토콜 정의는 [RFC 6750](https://www.rfc-editor.org/rfc/rfc6750)을 참고합니다. 예제는 RFC의 전체 토큰 발급·검증 체계를 구현하지 않습니다.

Bearer 토큰은 가진 사람이 사용할 수 있는 자격 증명입니다. 실제 네트워크에서는 전송 보호와 저장 위치가 중요합니다. 여기서는 요청을 메모리 안에서 만들기 때문에 실제 토큰을 전송하지 않습니다. 학습 문자열을 복사해 외부 서비스의 비밀값으로 사용하지 않습니다.

## 8. 서버 종료를 하나의 책임으로 만들기

### 8.1 종료 함수가 소유하는 것

`service` 패키지는 이미 만들어진 서버, listener와 취소 context를 받습니다. 핸들러나 DB 구조는 알지 못합니다. listener를 넘긴 뒤에는 함수가 닫는 책임을 가집니다. 양수인 종료 제한 시간을 전달해야 합니다.

저장 경로: [`examples/25-service-operations/service/lifecycle.go`](../../examples/25-service-operations/service/lifecycle.go).

```go
// Package service connects a server lifetime to a cancellation signal.
package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// Serve takes ownership of listener. A canceled ctx requests graceful shutdown.
// Handlers must honor request cancellation if a forced close becomes necessary.
func Serve(ctx context.Context, server *http.Server, listener net.Listener, timeout time.Duration) error {
	if timeout <= 0 {
		return errors.Join(errors.New("종료 제한 시간은 양수여야 합니다"), listener.Close())
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(err, listener.Close())
	}
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	select {
	case err := <-result:
		return errors.Join(err, server.Close())
	case <-ctx.Done():
		drain, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		shutdownErr := server.Shutdown(drain)
		var closeErr error
		if shutdownErr != nil {
			closeErr = server.Close()
		}
		serveErr := <-result
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		return errors.Join(shutdownErr, closeErr, serveErr)
	}
}
```

이 파일은 `package service`이므로 `go run` 대상이 아닙니다. 다음 실행 예제가 import합니다. 작은 프로그램마다 같은 종료 코드를 다시 복사하지 않고 패키지 계약으로 재사용하는 구성입니다.

### 8.2 왜 종료 context는 Background에서 시작하는가

종료 요청이 오면 전달받은 `ctx`는 이미 취소되었습니다. 그 context로 `WithTimeout`을 만들어도 즉시 취소된 자식이 생깁니다. 진행 중인 응답을 기다릴 시간이 없습니다. 따라서 종료 대기는 별도의 `context.Background()`에서 제한 시간을 만듭니다.

이는 취소를 무시한다는 뜻이 아닙니다. “새 요청을 받으며 실행하는 수명”과 “이미 시작한 요청을 정리하는 수명”을 구분한 것입니다. 정리 시간도 무한하게 두지 않고 상한을 정합니다.

### 8.3 Serve와 Shutdown을 모두 기다린다

[Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)은 listener를 닫고 유휴 연결을 닫은 다음 활성 연결의 작업이 끝나기를 기다립니다. `Serve`는 종료가 시작되면 `http.ErrServerClosed`로 먼저 돌아올 수 있습니다. 이 값 하나만 보고 프로세스를 끝내면 아직 응답 중인 작업을 기다리지 못합니다.

함수는 `Shutdown`이 끝난 뒤 `Serve` 결과도 받습니다. 정상 종료 경로에서 `ErrServerClosed`는 예상한 결과이므로 nil로 변환하고, 다른 오류는 보존합니다. `errors.Join`은 종료 대기 오류와 강제 닫기 오류를 하나가 다른 하나를 지우지 않게 모읍니다.

### 8.4 제한 시간이 지나면

`Shutdown`이 시간 안에 끝나지 않으면 `Close`로 연결을 닫습니다. 핸들러가 요청 context를 관찰하고 있다면 취소를 보고 작업을 마칠 수 있습니다. 그러나 context를 무시하는 계산이나 외부 작업까지 `Close`가 강제로 중단하는 것은 아닙니다.

따라서 다음 두 결과를 구분합니다.

| 결과 | 호출자가 할 수 있는 해석 |
| --- | --- |
| 반환 오류 없음 | 일반 HTTP 요청의 정상 종료 대기를 마침 |
| `context.DeadlineExceeded` 포함 | 제한 내 종료 실패, 강제 연결 닫기 시도 |
| 서버 실행 오류 | 수신 루프 자체가 실패함 |
| 시작 전 취소 | listener를 닫고 실행을 시작하지 않음 |

정상 종료 대기와 별도로 고루틴을 직접 만들었다면 그 고루틴의 완료도 별도로 기다려야 합니다. hijack한 연결이나 WebSocket도 `Shutdown`이 자동으로 모두 처리하는 대상이 아닙니다. 이 패키지는 일반 `net/http` 요청의 수명에 범위를 둡니다.

## 9. 실제 포트를 열고 종료까지 확인하기

저장 경로: [`examples/25-service-operations/shutdown/main.go`](../../examples/25-service-operations/shutdown/main.go).

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/25-service-operations/service"
)

func run() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	server := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})}
	result := make(chan error, 1)
	go func() { result <- service.Serve(ctx, server, listener, time.Second) }()
	client := &http.Client{Timeout: time.Second}
	defer client.CloseIdleConnections()
	response, requestErr := client.Get("http://" + listener.Addr().String())
	var readErr, closeErr error
	if requestErr == nil {
		_, readErr = io.Copy(io.Discard, response.Body)
		closeErr = response.Body.Close()
		fmt.Println("요청 상태:", response.StatusCode)
	}
	cancel()
	stopErr := <-result
	if err := errors.Join(requestErr, readErr, closeErr, stopErr); err != nil {
		return err
	}
	fmt.Println("종료 완료:", true)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./examples/25-service-operations/shutdown
```

실행 결과:

```text
요청 상태: 204
종료 완료: true
```


listener는 `127.0.0.1:0`으로 엽니다. 포트 0은 사용 가능한 포트를 운영체제가 배정하도록 요청하는 값입니다. 테스트가 고정 포트를 두고 경쟁하지 않으며 외부 인터페이스에 공개하지 않습니다.

프로그램은 다음 순서로 끝납니다.

1. listener와 서버를 준비합니다.
2. 별도 고루틴에서 `service.Serve`를 실행합니다.
3. 같은 프로세스의 HTTP 클라이언트가 요청합니다.
4. 응답 본문을 읽고 닫습니다.
5. `cancel()`로 종료를 요청합니다.
6. `result`에서 종료 결과를 받을 때까지 기다립니다.
7. 모든 오류를 모아 확인하고 완료를 출력합니다.

클라이언트 오류가 나도 종료 요청과 결과 대기를 건너뛰지 않습니다. 이것이 예제에서 요청 처리 직후 곧바로 `return`하지 않는 이유입니다. 이미 시작한 서버를 남겨 둔 채 테스트를 끝내지 않도록 모든 분기에서 수명을 정리합니다.

### 9.1 실제 운영 신호와 연결하기

직접 실행하는 프로그램에서는 [signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)로 Ctrl+C 같은 신호를 context 취소에 연결할 수 있습니다. 연습문제 `managed-server`는 `os.Interrupt`를 사용하며 신호 등록을 해제하는 `stop`도 호출합니다.

컨테이너 환경에서 사용하는 SIGTERM 등은 운영체제와 실행 환경에 맞춰 처리해야 합니다. Ctrl+C만 처리하는 예제를 그대로 어떤 배포 환경의 완성된 종료 설정이라고 생각하지 않습니다. 28장에서 컨테이너 실행과 연결하고, 30장의 실제 서버 코드를 보며 종료 신호와 자원 수명을 다시 확인합니다.

`signal.NotifyContext`를 등록한 동안 해당 신호의 기본 동작이 달라질 수 있습니다. `stop`은 관련 자원을 해제하고 기본 처리 복원을 돕습니다. 종료 도중 두 번째 신호를 어떻게 처리할지도 실제 서비스의 정책에 포함됩니다.

### 9.2 DB는 어디에서 닫는가

DB를 사용한다면 시작 코드에서 연 뒤 서버 종료가 끝난 다음 닫습니다. 각 요청의 끝에서는 Rows와 Tx 등 요청에 속한 자원을 닫고, 공유 `*sql.DB`는 요청마다 닫지 않습니다.

정상 종료를 기다리는 동안 DB를 먼저 닫으면 진행 중인 요청이 남은 조회를 못 할 수 있습니다. 종료 제한을 넘겼다면 이미 정상 종료가 실패한 상황이므로 그 사실을 오류로 남기고 강제 종료 정책을 적용합니다. 실패를 정상 완료 메시지로 바꾸지 않습니다.

## 10. 종료 테스트를 시계 대신 사건으로 동기화하기

핵심 테스트는 “충분히 오래 기다렸으니 요청이 실행됐을 것”이라고 가정하지 않습니다. 테스트 핸들러에 들어갔을 때 채널을 닫고, listener가 닫혔을 때도 신호를 보냅니다. 테스트는 관찰한 사건 순서대로 다음 단계를 진행합니다.

실제 테스트 파일: [`service/behavior_test.go`](../../examples/25-service-operations/service/behavior_test.go). 아래는 **정상 종료 테스트에 사용한 타입과 함수의 발췌**입니다.

```go
type watchedListener struct {
	net.Listener
	closed chan struct{}
	once   sync.Once
}

func (l *watchedListener) Close() error {
	err := l.Listener.Close()
	l.once.Do(func() { close(l.closed) })
	return err
}
func wait(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("operation did not finish")
	}
}
func TestServeDrainsActiveRequest(t *testing.T) {
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener := &watchedListener{Listener: raw, closed: make(chan struct{})}
	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-release; w.WriteHeader(204) })}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- Serve(ctx, server, listener, time.Second) }()
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); _ = server.Close() })
	client := &http.Client{Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	requestDone := make(chan error, 1)
	go func() {
		response, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_, readErr := io.Copy(io.Discard, response.Body)
			err = errors.Join(readErr, response.Body.Close())
			if response.StatusCode != 204 {
				err = errors.New("wrong response")
			}
		}
		requestDone <- err
	}()
	wait(t, entered)
	cancel()
	wait(t, listener.closed)
	select {
	case err := <-stopped:
		t.Fatalf("returned before handler: %v", err)
	default:
	}
	releaseOnce.Do(func() { close(release) })
	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}
```

활성 요청이 대기 중일 때 종료를 요청한 뒤 서버 결과가 아직 오지 않았는지 확인합니다. 그다음 요청의 대기를 풀고 HTTP 응답과 서버 종료를 모두 확인합니다. 이렇게 해야 `Shutdown`을 호출하기만 하고 기다리지 않는 잘못된 구현을 잡을 수 있습니다.

테스트의 몇 초짜리 시간 제한은 시스템이 멈췄을 때 검사 전체가 무한 대기하지 않게 하는 실패 경계입니다. “정확히 10밀리초 뒤에 끝난다” 같은 시간을 정답으로 요구하지 않습니다. 종료 제한 테스트는 결과가 `DeadlineExceeded`인지와 핸들러가 요청 취소를 관찰했는지 확인합니다.

```sh
go test -count=1 ./examples/25-service-operations/...
go test -count=1 ./solutions/25-service-operations/...
go test -race ./examples/25-service-operations/... ./solutions/25-service-operations/...
```

레이스 검사는 지원 운영체제와 C 컴파일러가 준비된 환경에서 실행합니다. 테스트가 통과하더라도 실제 배포 환경의 네트워크, 프록시, 로그 수집기까지 검증했다는 뜻은 아닙니다. 이 장의 테스트 범위는 입력 계약과 로컬 서버 수명입니다.

## 11. 개인 실습 폴더에서 변경하기

원본 예제를 수정하기 전에 `_practice/25-service-operations/configuration`에 복사합니다. 다음 명령은 기존 파일이 있으면 덮어쓰지 않습니다.

Windows PowerShell:

```powershell
$practiceDir = "_practice/25-service-operations/configuration"
New-Item -ItemType Directory -Force $practiceDir | Out-Null
if (-not (Test-Path "$practiceDir/main.go")) {
    Copy-Item "examples/25-service-operations/configuration/main.go" "$practiceDir/main.go"
}
```

macOS·Linux·WSL:

```bash
practice_dir=_practice/25-service-operations/configuration
mkdir -p "$practice_dir"
if [ ! -e "$practice_dir/main.go" ]; then
    cp examples/25-service-operations/configuration/main.go "$practice_dir/main.go"
fi
```

```sh
go run ./_practice/25-service-operations/configuration
```

1. `loadConfig`에 전달하는 함수를 `os.LookupEnv`로 바꿉니다.
2. 설정 변수가 없을 때 기본값을 확인합니다.
3. 현재 터미널에서만 `APP_READ_TIMEOUT=500ms`를 설정하고 실행합니다.
4. 값을 `oops`로 바꿔 설정 오류가 출력되는지 확인합니다.
5. 환경 변수를 제거하고 기본값으로 돌아오는지 확인합니다.

PowerShell에서는 `$env:APP_READ_TIMEOUT = "500ms"`, 제거에는 `Remove-Item Env:APP_READ_TIMEOUT`을 사용합니다. POSIX 셸에서는 한 명령에만 `APP_READ_TIMEOUT=500ms go run ...` 형태로 전달할 수 있습니다. 실제 토큰을 셸 명령 이력에 넣는 연습은 하지 않습니다.

## 12. 자주 하는 실수와 수정 방향

| 증상 | 원인 | 확인과 수정 |
| --- | --- | --- |
| 환경 설정이 틀려도 서버가 뜬다 | 파싱 오류를 기본값으로 덮음 | 시작 전에 오류 반환 |
| 테스트 순서에 따라 설정 결과가 바뀐다 | 전역 환경·플래그를 공유 | lookup, args, 출력자를 주입 |
| 로그에서 비밀값이 보인다 | 구조체·URL·헤더 전체 출력 | 허용할 필드만 기록 |
| 클라이언트는 200인데 로그는 500 | 최종 상태를 여러 번 덮어씀 | 첫 최종 상태 보관 |
| DB 장애 때 프로세스가 계속 재시작된다 | liveness에 모든 의존 검사를 넣음 | 준비 상태와 의미 분리 |
| 종료가 즉시 시간 초과된다 | 이미 취소된 context로 Shutdown | 별도 제한 시간으로 정리 |
| 종료 완료라고 찍고 응답이 끊긴다 | Serve의 반환만 기다림 | Shutdown 완료도 대기 |
| 종료 후 고루틴이 남는다 | 직접 만든 작업의 완료 미수집 | 작업 소유자와 종료 조건 정의 |
| 인증 실패인데 데이터가 바뀐다 | 다음 핸들러 실행 후 토큰 확인 | 인증을 변경 처리 앞에 배치 |
| 한 요청을 여러 번 기록한다 | 같은 미들웨어 중복 적용 | 구성 지점에서 한 번 연결 |

문제의 시작점은 “무슨 일이 일어났는가”와 “누가 그 자원을 소유하는가”입니다. 더 많은 로그를 무조건 남기기보다, 설정 검증·요청 실행·종료 중 어느 단계에서 실패했는지 먼저 분류합니다.

## 연습문제

[25장 연습문제](../../exercises/25-service-operations/README.md)는 설정 우선순위, 비밀정보 제외, 종료 게이트, 실제 서버 구성으로 이어집니다. 문제의 계약을 먼저 읽고 직접 구현한 다음 [해설](../../solutions/25-service-operations/README.md)의 코드와 테스트를 비교합니다.

## 요약과 자기 점검

서비스의 바깥쪽은 설정과 자원의 수명을 소유합니다. 요청의 안쪽은 입력을 검증하고 결과를 만듭니다. 로그는 필요한 사건과 속성을 남기고, readiness는 현재 요청 수용 가능성을 나타냅니다. 종료는 신규 수신 중지와 진행 중인 요청 대기를 구분해야 합니다.

<details>
<summary>1. 환경 변수가 없을 때와 빈 문자열일 때 왜 다르게 처리하나요?</summary>

정해진 기본값을 쓸 상황과 사용자가 잘못 비워 둔 설정을 구분하기 위해서입니다. 이는 이 예제의 정책이며 LookupEnv의 bool 결과로 구현합니다.

</details>

<details>
<summary>2. LogValuer를 구현하면 fmt 출력도 자동으로 가려지나요?</summary>

아닙니다. slog의 값 변환 경로에서 사용하는 계약입니다. 원래 문자열로 변환하거나 다른 출력 함수를 사용하면 해당 보호를 우회할 수 있습니다.

</details>

<details>
<summary>3. 준비 상태를 false로 바꾸면 진행 중인 요청도 즉시 취소되나요?</summary>

아닙니다. 새 요청을 받을 준비가 되지 않았다는 상태입니다. 기존 작업의 수명과 외부 라우터의 반영 시점은 별도로 관리해야 합니다.

</details>

<details>
<summary>4. 정상 종료가 시작되었는데 Serve가 먼저 반환하는 이유는 무엇인가요?</summary>

listener를 닫아 신규 수신을 중지했기 때문입니다. 활성 요청의 완료를 기다리는 Shutdown은 그 뒤에도 진행될 수 있습니다.

</details>

<details>
<summary>5. Close가 성공했으면 모든 핸들러 계산이 끝났다고 말할 수 있나요?</summary>

아닙니다. 연결을 닫는 것과 모든 고루틴의 계산 완료는 다릅니다. 작업은 요청 context를 관찰해야 하며 별도로 시작한 고루틴은 소유자가 기다려야 합니다.

</details>

## 참고 자료

- [log/slog](https://pkg.go.dev/log/slog): 로거, 핸들러, 속성과 LogValuer
- [os.LookupEnv](https://pkg.go.dev/os#LookupEnv): 환경 변수의 존재 여부
- [net.SplitHostPort](https://pkg.go.dev/net#SplitHostPort): 호스트·포트 구분
- [http.Server](https://pkg.go.dev/net/http#Server): 서버 제한 시간과 종료
- [http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown): 활성 요청 대기와 제외되는 연결
- [http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter): 응답 상태와 본문 작성
- [http.ResponseController](https://pkg.go.dev/net/http#ResponseController): 래퍼와 Unwrap
- [os/signal](https://pkg.go.dev/os/signal): 운영체제별 신호 처리
- [crypto/subtle](https://pkg.go.dev/crypto/subtle#ConstantTimeCompare): 고정 길이 비교
- [RFC 6750](https://www.rfc-editor.org/rfc/rfc6750): Bearer 토큰의 HTTP 사용

공식 API 확인일: **2026-11-09**. 실행 환경의 정책은 해당 배포 도구 문서를 추가로 확인합니다. 예제는 외부 인증 제공자나 운영 로그 시스템을 구성하지 않습니다.

[이전: 24장 SQL과 트랜잭션](../24-sql-transactions/README.md) · [전체 목차](../README.md) · [다음: 26장 성능 측정과 진단](../26-performance/README.md)
