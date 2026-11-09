# 21. HTTP 클라이언트

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 중급 | 함수·오류·인터페이스·JSON·테스트·컨텍스트 | Go, 편집기, 터미널, 이 저장소 |

프로그램이 HTTP 서비스를 호출할 때는 “주소에서 문자열을 가져온다”보다 많은 일이 일어납니다.
주소를 만들고, 요청을 보내고, 응답의 상태와 형식을 확인하고, 본문을 읽고, 사용한 자원을 정리해야 합니다.
응답이 늦거나 중간에 끊어질 때 어디까지 기다릴지도 정해야 합니다.

이번 장에서는 그 경계를 작은 함수로 나눕니다.
각 예제는 자신의 컴퓨터 안에 임시 HTTP 서버를 만들고 실제 요청을 보냅니다.
외부 사이트의 가용성이나 계정에 의존하지 않으며, 실행이 끝나면 서버도 종료됩니다.
화면에 나타나는 포트나 시간에 기대지 않도록 출력은 고정했습니다.

## 학습 목표

- 컨텍스트·메서드·URL·헤더를 포함한 요청을 만들고 `http.Client`로 전송한다.
- 전송 오류, HTTP 상태 오류, 본문 해석 오류를 구분한다.
- 응답 본문을 닫고 바이트 한도를 적용하여 읽는다.
- JSON 형식과 데이터 조건을 각각 검사한다.
- 요청 취소, 클라이언트 재사용, 리다이렉트와 TLS 신뢰 범위를 설명한다.

## 사전 지식과 준비

이 장의 코드는 구조체, 메서드, `error`, `%w`, `errors.Is/As`, `defer`, `io.Reader`를 사용합니다.
JSON과 컨텍스트가 낯설다면 [전체 목차](../README.md)에서 관련 장을 먼저 확인합니다.
취소 예제는 채널로 시작과 종료를 알리며 `time.Sleep`으로 순서를 맞추지 않습니다.

모든 명령은 **저장소 루트**에서 실행합니다.
Go 명령은 PowerShell과 Bash/Zsh에서 같습니다.
개인 작업 파일의 복사 명령만 셸별로 나눕니다.
루트의 [`go.mod`](../../go.mod)를 사용하며 새 모듈을 만들지 않습니다.

| 예제 | 중심 주제 | 실행 명령 |
| --- | --- | --- |
| [basic-get](../../examples/21-http-client/basic-get/main.go) | 요청·응답의 기본 수명 | `go run ./examples/21-http-client/basic-get` |
| [query-headers](../../examples/21-http-client/query-headers/main.go) | 쿼리 인코딩과 요청 헤더 | `go run ./examples/21-http-client/query-headers` |
| [status-error](../../examples/21-http-client/status-error/main.go) | 전송 오류와 상태 오류 | `go run ./examples/21-http-client/status-error` |
| [bounded-body](../../examples/21-http-client/bounded-body/main.go) | 본문 크기 한도와 경계 | `go run ./examples/21-http-client/bounded-body` |
| [json-response](../../examples/21-http-client/json-response/main.go) | 제한된 JSON 해석과 검증 | `go run ./examples/21-http-client/json-response` |
| [cancellation](../../examples/21-http-client/cancellation/main.go) | 진행 중인 요청 취소 | `go run ./examples/21-http-client/cancellation` |
| [redirect-policy](../../examples/21-http-client/redirect-policy/main.go) | TLS 테스트와 이동 정책 | `go run ./examples/21-http-client/redirect-policy` |

먼저 한 예제를 실행하고 출력을 확인하세요.
실습 서버를 따로 켜거나 브라우저를 열 필요가 없습니다.
각 프로그램은 요청과 확인을 마친 뒤 `server.Close()`를 실행하고 스스로 끝납니다.

## 1. 응답을 받았다는 말의 세 가지 의미

HTTP는 요청과 응답을 주고받는 프로토콜입니다.
클라이언트는 일을 요청하는 쪽이며, 서버는 요청을 받아 응답하는 쪽입니다.
이번 장의 관심은 클라이언트가 받은 결과를 자신의 프로그램에서 어떻게 해석할지에 있습니다.

응답에는 상태 코드, 헤더, 본문이 있습니다.
상태 코드는 결과의 종류를 나타내고, 헤더는 본문 형식 등의 부가 정보를 담습니다.
본문은 실제 데이터입니다. 본문이 없는 응답도 있습니다.

| 단계 | 확인할 질문 | 대표 실패 |
| --- | --- | --- |
| 요청 생성 | URL과 메서드를 요청 객체로 만들 수 있는가? | 잘못된 URL |
| 전송 | 서버와 HTTP 대화를 마칠 수 있는가? | 연결 실패, 취소, 시간 초과 |
| 상태 판정 | 이 API에서 성공으로 약속한 코드인가? | 404, 503 |
| 본문 읽기 | 한도 안에서 끝까지 읽을 수 있는가? | 읽기 오류, 한도 초과 |
| 데이터 해석 | 형식과 업무 조건이 맞는가? | 잘못된 JSON, 빈 제목 |

`client.Do`가 오류 없이 반환했다는 사실은 두 번째 단계의 성공을 뜻합니다.
404나 503도 정상적으로 전달된 HTTP 응답이므로 별도로 판정해야 합니다.
이는 [Client.Do의 계약](https://pkg.go.dev/net/http#Client.Do)에 명시되어 있습니다.

이 교재의 작은 조회 API는 정확히 `200 OK`만 성공으로 받습니다.
모든 HTTP API가 200만 성공인 것은 아닙니다.
생성 API는 201, 본문 없는 작업은 204 등을 성공으로 정할 수 있습니다.
호출 대상의 계약에 맞게 허용 코드를 정해야 합니다.

### 상태를 읽기 전에 본문부터 해석하면 생기는 일

서버가 503과 함께 HTML 오류 페이지를 보냈다고 생각해 보세요.
JSON만 기대하고 바로 디코딩하면 오류가 “JSON 문법 문제”처럼 보입니다.
상태를 먼저 확인하면 서버가 요청을 처리하지 못했다는 원인을 드러낼 수 있습니다.

반대로 200이더라도 본문이 깨진 JSON이면 성공 데이터로 사용할 수 없습니다.
전송, 상태, 형식, 데이터 조건을 한 번의 `err == nil`로 대신하지 않는 이유입니다.

## 2. 외부 서비스 없이 HTTP를 연습하는 방법

`httptest.NewServer`는 HTTP 핸들러를 받아 로컬 테스트 서버를 시작합니다.
`server.URL`은 실행마다 달라질 수 있는 주소이며, 코드에서 직접 사용합니다.
정해진 포트 번호를 외워 다른 프로그램과 충돌시킬 필요가 없습니다.
[httptest.Server](https://pkg.go.dev/net/http/httptest#Server)의 사용 계약을 참고합니다.

이 장에서는 서버 코드가 요청을 받은 뒤 무엇을 돌려줄지 정하는 실험 장치 역할을 합니다.
핸들러의 인자 `w`는 응답을 쓰는 대상이고 `r`은 받은 요청입니다.
서버 설계와 라우팅은 다음 장에서 자세히 다룹니다.

| 코드 | 역할 |
| --- | --- |
| `httptest.NewServer(handler)` | 실습용 서버 시작 |
| `defer server.Close()` | 함수가 끝날 때 서버 종료 |
| `server.Client()` | 그 서버에 요청하도록 준비된 클라이언트 |
| `client.Timeout = 2 * time.Second` | 예제의 최대 대기 시간 설정 |
| `client.CloseIdleConnections()` | 마지막에 사용하지 않는 연결 정리 |

임시 서버도 실제 요청을 처리합니다.
서버를 닫는 시점에 아직 끝나지 않은 핸들러가 있으면 종료가 기다려질 수 있습니다.
따라서 취소 예제는 클라이언트 취소와 서버 작업 종료를 함께 확인합니다.
[Server.Close](https://pkg.go.dev/net/http/httptest#Server.Close)는 진행 중 요청의 종료를 기다립니다.

## 3. 첫 요청: 생성, 전송, 검사, 읽기, 닫기

저장 경로: [basic-get/main.go](../../examples/21-http-client/basic-get/main.go)

전체 코드:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func fetchText(ctx context.Context, client *http.Client, address string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return "", fmt.Errorf("요청 생성: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("요청 전송: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("응답 상태: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1025))
	if err != nil {
		return "", fmt.Errorf("본문 읽기: %w", err)
	}
	if len(body) > 1024 {
		return "", fmt.Errorf("본문이 1024바이트를 초과함")
	}
	return string(body), nil
}

func run() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, "Go client"); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	text, err := fetchText(context.Background(), client, server.URL)
	if err != nil {
		return err
	}
	fmt.Println("본문:", text)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```text
go run ./examples/21-http-client/basic-get
```

예상 출력:

```text
본문: Go client
```

### `fetchText`의 수명 따라가기

`http.NewRequestWithContext`는 요청 객체를 만듭니다.
이 호출 자체가 서버에 접속하는 것은 아닙니다.
실제 전송은 `client.Do(req)`에서 시작됩니다.
[NewRequestWithContext](https://pkg.go.dev/net/http#NewRequestWithContext)는 요청의 컨텍스트를 연결합니다.

| 순서 | 작업 | 실패하면 어떻게 되는가? |
| --- | --- | --- |
| 1 | 요청 객체 생성 | 생성 오류를 감싸 반환 |
| 2 | `client.Do`로 전송 | 전송 오류를 감싸 반환 |
| 3 | 본문 닫기를 `defer`로 등록 | 이후 성공·실패 경로에 공통 적용 |
| 4 | `StatusCode` 확인 | 상태 오류 반환 |
| 5 | 최대 1025바이트 읽기 | 읽기 오류 반환 |
| 6 | 실제 길이가 1024를 넘는지 확인 | 한도 초과 반환 |
| 7 | 문자열 반환 | 호출자가 화면에 출력 |

`defer resp.Body.Close()`는 `err`가 없는 것을 확인한 다음에 둡니다.
응답을 얻지 못한 경로에서 `resp.Body`를 먼저 사용하면 nil 포인터 문제가 생길 수 있습니다.
정상 응답을 얻은 뒤에는 상태가 기대와 달라도 본문을 닫습니다.

### 함수가 반환하는 것과 본문 자원의 소유자

`fetchText`는 응답 본문을 모두 다룬 뒤 문자열을 반환합니다.
따라서 호출자는 `resp.Body`를 닫을 필요가 없습니다.
이 함수가 응답 객체나 `io.ReadCloser` 자체를 반환하는 설계라면 닫을 책임도 호출자에게 전달해야 합니다.
반환 타입을 고를 때 값의 형태와 자원의 소유권을 함께 정합니다.

HTTP 본문은 네트워크에서 읽는 스트림입니다.
본문을 닫는 일은 단순히 문자열 변수를 버리는 일과 다릅니다.
[Client.Do](https://pkg.go.dev/net/http#Client.Do)는 성공 시 반환된 본문을 호출자가 닫도록 요구합니다.

예제에서는 읽기 오류를 반드시 확인하고, 닫기는 `defer`로 보장합니다.
이미 읽은 응답의 `Close` 오류를 새 업무 결과로 사용하지 않는 정리 정책입니다.
다른 파일·압축·쓰기 자원의 닫기 오류 처리와 무조건 같은 규칙이라고 일반화하지 않습니다.

### `main`보다 `run`에 실습 자원을 둔 이유

`os.Exit`는 등록된 `defer`를 실행하지 않고 프로세스를 종료합니다.
서버와 클라이언트 정리를 `run` 안에 두면 `run`이 반환할 때 정리가 끝납니다.
그 뒤 `main`이 오류를 보고하고 종료 코드를 정합니다.
[os.Exit](https://pkg.go.dev/os#Exit)의 동작을 참고합니다.

## 4. URL과 헤더를 문자열 이어 붙이기로 만들지 않기

검색어에 공백, `&`, `+`, 한글이 들어가면 URL에서의 의미를 구분해야 합니다.
`?q=` 뒤에 입력을 그대로 붙이면 `&`가 검색어의 일부가 아니라 다른 쿼리의 시작처럼 해석될 수 있습니다.
`url.Values.Encode`로 쿼리 값을 인코딩합니다.
[Values.Encode](https://pkg.go.dev/net/url#Values.Encode)는 URL 인코딩한 쿼리 문자열을 만듭니다.

저장 경로: [query-headers/main.go](../../examples/21-http-client/query-headers/main.go)

전체 코드:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"
)

func search(ctx context.Context, client *http.Client, address, term string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("q", term)
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept-Language", "ko-KR")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("상태: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1025))
	if err != nil {
		return "", err
	}
	if len(body) > 1024 {
		return "", fmt.Errorf("본문 한도 초과")
	}
	return string(body), nil
}

func run() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "검색어: %s\n언어: %s\n", r.URL.Query().Get("q"), r.Header.Get("Accept-Language")); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	text, err := search(context.Background(), client, server.URL, "Go & HTTP")
	if err != nil {
		return err
	}
	fmt.Print(text)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```text
go run ./examples/21-http-client/query-headers
```

예상 출력:

```text
검색어: Go & HTTP
언어: ko-KR
```

### 쿼리를 수정하는 네 단계

| 표현 | 값의 의미 |
| --- | --- |
| `url.Parse(address)` | 문자열 주소를 구성 요소로 해석 |
| `u.Query()` | 기존 쿼리를 값 모음으로 읽음 |
| `query.Set("q", term)` | `q`의 값을 하나로 설정 |
| `u.RawQuery = query.Encode()` | 수정한 값을 URL에 다시 반영 |

`u.Query()`로 얻은 값을 바꿨다고 URL이 저절로 바뀌지는 않습니다.
인코딩한 결과를 `RawQuery`에 다시 넣어야 합니다.
기존 주소에 `keep=1`이 있었다면 `u.Query()`에서 시작하기 때문에 그 값도 남습니다.

서버에서는 `r.URL.Query().Get("q")`로 해석된 검색어를 읽습니다.
전송 URL에는 이스케이프된 표현이 들어가지만, 서버에서 읽는 값은 `Go & HTTP`입니다.
본문의 출력은 인코딩된 URL 문자열이 아니라 서버가 해석한 값입니다.

### 헤더는 요청을 만든 뒤 설정하기

`req.Header.Set("Accept-Language", "ko-KR")`는 언어 선호를 전달합니다.
이는 서버가 반드시 그 언어를 반환한다는 보장은 아닙니다.
이번 실험 서버는 받은 헤더를 응답에 써서 무엇이 전송되었는지 보여 줍니다.

| 헤더 | 일반적인 용도 | 주의점 |
| --- | --- | --- |
| `Accept` | 받을 수 있는 응답 형식 표시 | 실제 응답 형식을 별도로 확인 |
| `Content-Type` | 현재 보내는 본문의 형식 표시 | GET에 본문이 없다면 자동으로 필요한 값은 아님 |
| `Accept-Language` | 선호 언어 표시 | 서버 정책에 따라 다르게 응답 가능 |
| `Authorization` | 인증 정보 전달 | 출력·오류 로그에 값 자체를 남기지 않음 |

이번 실습에는 인증 정보가 필요하지 않습니다.
헤더 값을 비교하는 테스트에는 고정된 학습용 값만 사용합니다.
같은 요청 객체를 여러 고루틴에서 수정하며 재사용하지 않습니다.
클라이언트는 재사용하되 요청과 헤더는 요청마다 구성합니다.

## 5. 상태 코드를 오류 타입으로 표현하기

전송 오류와 HTTP 상태를 모두 문자열로만 처리하면 이후 분기가 어려워집니다.
상태 코드를 필드로 가진 오류 타입을 만들면 호출자가 의미 있는 조건으로 분기할 수 있습니다.
문자열에 `404`가 포함되어 있는지 검색할 필요가 없습니다.

저장 경로: [status-error/main.go](../../examples/21-http-client/status-error/main.go)

전체 코드:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("HTTP 상태 %d", e.Code)
}

func fetch(ctx context.Context, client *http.Client, address string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("전송: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", &StatusError{Code: resp.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 129))
	if err != nil {
		return "", err
	}
	if len(body) > 128 {
		return "", errors.New("본문 한도 초과")
	}
	return string(body), nil
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "준비 완료"); err != nil {
			return
		}
	})
	mux.HandleFunc("/busy", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "busy", http.StatusServiceUnavailable)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, path := range []string{"/missing", "/busy", "/ok"} {
		text, err := fetch(context.Background(), client, server.URL+path)
		var status *StatusError
		if errors.As(err, &status) {
			fmt.Println("실패 상태:", status.Code)
			continue
		}
		if err != nil {
			return err
		}
		fmt.Println("본문:", text)
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

실행:

```text
go run ./examples/21-http-client/status-error
```

예상 출력:

```text
실패 상태: 404
실패 상태: 503
본문: 준비 완료
```

### 세 번의 요청을 분리해서 읽기

| 주소 경로 | 서버 응답 | `fetch`의 결과 | 호출자의 동작 |
| --- | --- | --- | --- |
| `/missing` | 404 | `*StatusError{Code: 404}` | 상태를 출력하고 다음 요청 |
| `/busy` | 503 | `*StatusError{Code: 503}` | 상태를 출력하고 다음 요청 |
| `/ok` | 200과 본문 | 문자열, nil | 본문 출력 |

`errors.As`는 오류가 특정 타입인지 확인하고 그 값을 받을 수 있게 합니다.
오류를 `%w`로 감싼 경우에도 연결된 오류 안에서 찾을 수 있습니다.
[errors.As](https://pkg.go.dev/errors#As)의 타입 기반 검사를 사용합니다.

이 예제에서는 서버가 준 실패 본문을 오류 문자열에 복사하지 않습니다.
업무 코드가 판정에 필요한 값은 상태 코드뿐이기 때문입니다.
실제 서비스에서 실패 본문을 진단용으로 읽는다면 길이 제한과 민감 정보 처리 정책을 별도로 정합니다.

### 왜 204도 이 함수에서는 오류일까?

204는 일반적으로 성공 계열 상태지만, 이 함수의 계약은 “200의 텍스트 본문”입니다.
따라서 204를 `*StatusError`로 돌려주는 것은 함수 계약에 따른 선택입니다.
허용 상태를 넓히려면 본문 없는 경우 어떤 값을 반환할지도 정해야 합니다.

### 재시도는 상태 출력 다음에 자동으로 따라오는 일이 아닙니다

503을 만났다고 같은 요청을 무한히 보내면 부하가 더 커질 수 있습니다.
재시도를 설계할 때는 최대 횟수, 전체 기한, 대기 방식, 취소, 작업의 중복 영향을 함께 정합니다.
조회와 결제·생성 요청을 같은 방식으로 재시도하면 안 되는 이유도 여기에 있습니다.

이번 장의 함수는 응용 수준의 재시도를 구현하지 않습니다.
한 번의 호출에서 결과와 실패 원인을 명확하게 전달하는 데 집중합니다.
`Transport`가 내부적으로 수행하는 제한된 재전송 동작과 업무에서 설계하는 재시도 정책도 구분합니다.
[Transport](https://pkg.go.dev/net/http#Transport)의 문서를 참고합니다.

## 6. 본문을 제한해서 읽기

`io.ReadAll`은 끝까지 읽은 데이터를 메모리에 모읍니다.
작은 실험 서버에서는 괜찮아 보여도, 실제 응답의 크기가 항상 작다고 가정할 수는 없습니다.
읽을 바이트 수를 먼저 제한하고, 한도를 넘는지 판정합니다.

저장 경로: [bounded-body/main.go](../../examples/21-http-client/bounded-body/main.go)

전체 코드:

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

var ErrTooLarge = errors.New("본문 한도 초과")

func readAtMost(reader io.Reader, limit int64) ([]byte, error) {
	if limit < 0 || limit > 1<<20 {
		return nil, errors.New("한도는 0~1048576바이트")
	}
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, ErrTooLarge
	}
	return body, nil
}

func download(client *http.Client, address string, limit int64) ([]byte, error) {
	resp, err := client.Get(address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("상태: %d", resp.StatusCode)
	}
	return readAtMost(resp.Body, limit)
}

func run() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "ABCDE"); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	_, err := download(client, server.URL, 4)
	if !errors.Is(err, ErrTooLarge) {
		return fmt.Errorf("예상과 다른 결과: %v", err)
	}
	fmt.Println("4바이트 제한: 초과")
	body, err := download(client, server.URL, 5)
	if err != nil {
		return err
	}
	fmt.Println("5바이트 제한:", string(body))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```text
go run ./examples/21-http-client/bounded-body
```

예상 출력:

```text
4바이트 제한: 초과
5바이트 제한: ABCDE
```

### 한도보다 한 바이트 더 읽는 이유

4바이트까지만 읽으면 원본이 `ABCD`인지 `ABCDE`인지 구분할 수 없습니다.
둘 다 읽은 결과가 `ABCD`일 수 있기 때문입니다.
한도가 4라면 최대 5바이트를 읽고 실제 길이를 검사합니다.
[io.LimitReader](https://pkg.go.dev/io#LimitReader)는 한도 뒤에서 EOF처럼 동작합니다.

| 실제 본문 | 허용 한도 | 최대 읽기 | 관찰한 길이 | 판정 |
| --- | --- | --- | --- | --- |
| 빈 본문 | 0 | 1 | 0 | 허용 |
| `A` | 0 | 1 | 1 | 초과 |
| `ABCD` | 4 | 5 | 4 | 허용 |
| `ABCDE` | 4 | 5 | 5 | 초과 |
| 훨씬 긴 본문 | 4 | 5 | 최대 5 | 초과 |

`limit` 자체도 0~1 MiB로 제한합니다.
따라서 `limit+1` 계산이 정수 범위를 넘거나, 실수로 지나치게 큰 한도를 허용하지 않습니다.
이 상한은 교재의 정책이며 모든 API에 적합한 공통 값은 아닙니다.
실제 제한은 데이터 크기와 자원 예산을 기준으로 정합니다.

### 바이트와 글자 수는 다릅니다

UTF-8에서 `가`는 3바이트입니다.
이 함수는 글자 수가 아니라 메모리로 읽을 바이트 수를 제한합니다.
한도 3에서는 `가`를 허용하고 한도 2에서는 초과 오류가 납니다.
잘린 일부 문자열을 성공 데이터로 돌려주지 않는 것이 중요합니다.

### `Content-Length`만 믿으면 충분할까요?

헤더에 본문 길이가 없을 수 있고, 전달 방식에 따라 미리 알 수 없을 수도 있습니다.
길이 정보가 있더라도 실제 읽기를 제한하는 기능은 따로 필요합니다.
이번 함수는 본문에서 실제로 읽은 길이를 기준으로 판정합니다.
[Response.ContentLength](https://pkg.go.dev/net/http#Response)는 알 수 없는 길이를 -1로 표현할 수 있습니다.

### 한도를 넘으면 계속 읽어서 연결을 살려야 할까요?

정상적인 작은 본문을 EOF까지 읽고 닫으면 연결 재사용에 도움이 됩니다.
하지만 큰 본문을 끝없이 읽어 버리는 일은 우리가 정한 자원 한도와 충돌합니다.
이 예제는 초과를 확인하면 닫고 실패를 반환합니다.
연결이 재사용될 가능성보다 요청의 자원 경계를 우선한 선택입니다.

응답을 끝까지 읽지 않고 닫았을 때 재사용 여부를 단정하지 않습니다.
프로토콜, Transport와 읽은 상태에 따라 달라질 수 있습니다.
[Client.Do](https://pkg.go.dev/net/http#Client.Do)의 EOF와 닫기 설명을 확인하세요.

## 7. JSON: 문법이 맞는 것과 데이터가 유효한 것은 다릅니다

HTTP 응답을 구조체로 디코딩하면 필드를 이름으로 다룰 수 있습니다.
그러나 JSON 디코딩 한 번만으로 업무 조건까지 검사되는 것은 아닙니다.
이 예제는 본문 크기, 미디어 타입, 알 수 없는 필드, 추가 JSON 값, 필드 범위를 구분합니다.

저장 경로: [json-response/main.go](../../examples/21-http-client/json-response/main.go)

전체 코드:

```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

type Lesson struct {
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

func decodeLesson(reader io.Reader) (Lesson, error) {
	var lesson Lesson
	body, err := io.ReadAll(io.LimitReader(reader, 257))
	if err != nil {
		return lesson, err
	}
	if len(body) > 256 {
		return lesson, errors.New("본문 한도 초과")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&lesson); err != nil {
		return Lesson{}, fmt.Errorf("JSON 해석: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Lesson{}, errors.New("JSON 값은 하나여야 함")
	}
	if strings.TrimSpace(lesson.Title) == "" || lesson.Minutes < 1 || lesson.Minutes > 180 {
		return Lesson{}, errors.New("학습 데이터 범위 오류")
	}
	return lesson, nil
}

func fetchLesson(ctx context.Context, client *http.Client, address string) (Lesson, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return Lesson{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Lesson{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Lesson{}, fmt.Errorf("상태: %d", resp.StatusCode)
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return Lesson{}, errors.New("JSON Content-Type 필요")
	}
	return decodeLesson(resp.Body)
}

func run() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(Lesson{Title: "Go", Minutes: 25}); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	lesson, err := fetchLesson(context.Background(), client, server.URL)
	if err != nil {
		return err
	}
	fmt.Printf("학습: %s, %d분\n", lesson.Title, lesson.Minutes)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```text
go run ./examples/21-http-client/json-response
```

예상 출력:

```text
학습: Go, 25분
```

### 검증 순서를 고정하기

| 순서 | 검사 | 거절하는 예 |
| --- | --- | --- |
| 1 | 상태가 200인가? | 503과 JSON 본문 |
| 2 | 미디어 타입이 JSON인가? | `text/html` |
| 3 | 본문이 256바이트 이내인가? | 공백 257바이트 |
| 4 | 구조체로 해석할 수 있는가? | minutes가 문자열 |
| 5 | 정의하지 않은 필드가 없는가? | `extra` 필드 |
| 6 | JSON 값이 정확히 하나인가? | `{...}{}` 또는 뒤의 쓰레기 문자 |
| 7 | 제목·시간이 업무 범위에 있는가? | 빈 제목, 0분, 181분 |

미디어 타입은 `mime.ParseMediaType`으로 읽습니다.
`application/json; charset=utf-8`은 JSON 미디어 타입과 매개변수를 함께 가진 표현입니다.
전체 문자열이 정확히 `application/json`인지 비교하면 이 정상 표현을 불필요하게 거절할 수 있습니다.
[ParseMediaType](https://pkg.go.dev/mime#ParseMediaType)의 반환값을 사용합니다.

### 두 번째 `Decode`의 역할

첫 `Decode`가 첫 번째 JSON 값을 읽고 성공해도 뒤에 다른 값이 남아 있을 수 있습니다.
두 번째 디코딩 결과가 `io.EOF`인지 확인하면 공백 뒤에 더 읽을 값이 없는지 확인할 수 있습니다.

| 입력 | 첫 디코딩 | 두 번째 디코딩 | 결과 |
| --- | --- | --- | --- |
| 정상 객체 하나 | 성공 | EOF | 다음 데이터 검사 |
| 정상 객체와 공백 | 성공 | EOF | 다음 데이터 검사 |
| 객체 두 개 | 성공 | 성공 | 값이 둘이므로 거절 |
| 객체 뒤에 `x` | 성공 | 문법 오류 | 남은 문자가 있으므로 거절 |
| 빈 본문 | EOF | 실행하지 않음 | 첫 단계에서 거절 |

`DisallowUnknownFields`는 구조체에 대응하지 않는 필드를 거절합니다.
제목의 공백 여부나 분량의 범위까지 검사하는 기능은 아닙니다.
그 조건은 디코딩 뒤의 `if`에서 확인합니다.
[Decoder.DisallowUnknownFields](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields)를 참고합니다.

### 이 검사가 보장하지 않는 것

기본 `encoding/json`의 객체 키 처리에는 대소문자 매칭과 중복 키 처리 규칙이 있습니다.
알 수 없는 필드를 금지했다고 해서 중복 키까지 자동으로 거절하는 것은 아닙니다.
이 장의 계약은 알려진 필드와 값 하나, 필드 범위를 검사하는 수준입니다.
중복 키까지 금지해야 한다면 토큰 단계의 검증 등 별도 정책을 설계해야 합니다.
[encoding/json의 보안 고려 사항](https://pkg.go.dev/encoding/json#hdr-Security_Considerations)을 확인합니다.

`null`은 구문상 JSON이지만 이 예제에서는 유효한 학습 데이터가 아닙니다.
구조체의 필드가 영값으로 남아 이후 범위 검사에서 거절됩니다.
필수 필드의 “없음”과 명시적 영값을 구분해야 하는 모델은 포인터 필드 등 별도 표현을 검토합니다.
이번에는 두 경우 모두 유효하지 않다는 계약으로 정했습니다.

## 8. 기다림의 경계: 타임아웃과 취소

서버가 응답하지 않는 요청도 언젠가는 끝나야 합니다.
`Client.Timeout`은 연결, 리다이렉트, 응답 본문 읽기를 포함한 시간 제한입니다.
`Do`가 헤더를 받고 반환한 뒤에도 본문 읽기가 제한 대상입니다.
0이면 이 필드에 의한 제한이 없습니다.
[Client.Timeout](https://pkg.go.dev/net/http#Client)의 문서를 참고합니다.

컨텍스트는 개별 요청의 수명을 전달합니다.
상위 작업이 취소되었거나 사용자가 더 이상 결과를 기다리지 않으면 요청도 취소할 수 있습니다.
클라이언트 전체의 기본 시간 제한과 요청별 컨텍스트를 함께 사용할 수 있습니다.

저장 경로: [cancellation/main.go](../../examples/21-http-client/cancellation/main.go)

전체 코드:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func request(ctx context.Context, client *http.Client, address string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 호출: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func run() error {
	started := make(chan struct{})
	stopped := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(stopped)
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		result <- request(ctx, client, server.URL)
	}()
	select {
	case <-started:
		cancel()
	case err := <-result:
		return fmt.Errorf("취소 전 요청 종료: %v", err)
	}
	err := <-result
	<-stopped
	if !errors.Is(err, context.Canceled) {
		return fmt.Errorf("취소 결과: %v", err)
	}
	fmt.Println("요청 취소 확인: true")
	fmt.Println("서버 작업 종료: true")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```text
go run ./examples/21-http-client/cancellation
```

예상 출력:

```text
요청 취소 확인: true
서버 작업 종료: true
```

### 잠깐 자고 취소하는 대신 시작 신호를 기다리기

프로그램이 1밀리초를 기다렸다고 서버가 요청을 받았다는 보장은 없습니다.
이 예제는 서버 핸들러가 `started`를 닫은 후에 취소합니다.
따라서 실제로 진행 중인 요청을 취소한다는 조건이 분명합니다.

| 단계 | 클라이언트 | 서버 |
| --- | --- | --- |
| 1 | 요청 고루틴 시작 | 요청 대기 |
| 2 | 시작 신호 대기 | `started` 닫기 |
| 3 | `cancel()` 호출 | 요청 컨텍스트의 종료 대기 |
| 4 | 전송 오류 받기 | 컨텍스트 종료 후 반환 |
| 5 | `stopped` 확인 | 종료 신호 닫기 |

`result`는 오류 하나를 전달할 크기 1의 채널입니다.
시작 전에 전송이 실패한 경우에는 `select`의 오류 경로로 돌아옵니다.
요청이 시작되기를 무조건 기다려 초기 실패를 놓치지 않도록 했습니다.

`errors.Is(err, context.Canceled)`는 감싼 오류 안의 취소 원인을 확인합니다.
오류 문자열은 주소나 내부 세부 사항에 따라 달라질 수 있으므로 그대로 고정된 정답으로 사용하지 않습니다.
[context.Canceled](https://pkg.go.dev/context#Canceled)와 [errors.Is](https://pkg.go.dev/errors#Is)를 참고합니다.

### 취소가 임의의 코드를 강제로 멈추지는 않습니다

컨텍스트 취소는 신호입니다.
HTTP 클라이언트는 요청의 컨텍스트를 관찰하고, 이 실습 서버도 `r.Context().Done()`을 기다립니다.
컨텍스트를 보지 않는 무한 계산이 자동으로 중단되는 것은 아닙니다.
서버 쪽에서 다른 작업을 시작한다면 그 작업에도 취소를 전달해야 합니다.

`defer cancel()`은 정상 종료 경로에서도 컨텍스트 자원을 정리합니다.
이미 `cancel()`을 호출했어도 다시 호출할 수 있습니다.
[CancelFunc](https://pkg.go.dev/context#CancelFunc)의 계약을 확인하세요.

## 9. 클라이언트 재사용, TLS와 리다이렉트

`http.Client`와 그 전송 계층은 요청 사이에서 연결을 재사용할 수 있습니다.
서비스를 호출할 때마다 전역 설정을 바꾸기보다, 필요한 정책을 가진 클라이언트를 준비해 전달합니다.
클라이언트는 여러 고루틴에서 요청하는 데 사용할 수 있지만, 사용하는 중에 설정 필드를 동시에 바꾸면 안 됩니다.
[Client](https://pkg.go.dev/net/http#Client)와 [Transport](https://pkg.go.dev/net/http#Transport)의 재사용 설명을 참고합니다.

다음 예제는 TLS 실습 서버에서 302 응답을 만납니다.
처음에는 기본 정책으로 최종 주소까지 따라갑니다.
그 다음에는 이동을 멈추고 302와 `Location` 자체를 확인합니다.

저장 경로: [redirect-policy/main.go](../../examples/21-http-client/redirect-policy/main.go)

전체 코드:

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func inspect(client *http.Client, address string) (int, string, error) {
	resp, err := client.Get(address)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, 1024)); err != nil {
		return 0, "", err
	}
	return resp.StatusCode, resp.Header.Get("Location"), nil
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	})
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "완료"); err != nil {
			return
		}
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	code, _, err := inspect(client, server.URL+"/start")
	if err != nil {
		return err
	}
	fmt.Println("기본 정책:", code)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	code, location, err := inspect(client, server.URL+"/start")
	if err != nil {
		return err
	}
	fmt.Println("중지 정책:", code)
	fmt.Println("위치:", location)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```text
go run ./examples/21-http-client/redirect-policy
```

예상 출력:

```text
기본 정책: 200
중지 정책: 302
위치: /final
```

### 같은 최초 응답에서 서로 다른 관찰 결과

| 정책 | 최초 서버 응답 | 추가 요청 | 호출자가 받는 상태 |
| --- | --- | --- | --- |
| 기본 리다이렉트 | 302와 `/final` | `/final` 요청 | 200 |
| `ErrUseLastResponse` | 302와 `/final` | 없음 | 302 |

`http.ErrUseLastResponse`를 반환하면 마지막 리다이렉트 응답을 오류 없이 받을 수 있습니다.
그 본문도 호출자가 닫아야 합니다.
다른 오류를 반환해 리다이렉트를 실패시키는 경우와 구분합니다.
[ErrUseLastResponse](https://pkg.go.dev/net/http#ErrUseLastResponse)를 참고합니다.

예제는 앞 요청을 모두 마친 뒤에만 `CheckRedirect`를 바꿉니다.
동시에 요청을 보내는 프로그램에서는 동작 중인 클라이언트 필드를 변경하지 말고, 시작 전에 정책을 정합니다.
고정된 정책을 생성자나 구성 단계에서 주입하는 형태가 이해하기 쉽습니다.

### 학습용 TLS 인증서를 신뢰하는 방법

`httptest.NewTLSServer`는 학습용 TLS 인증서로 HTTPS 서버를 시작합니다.
`server.Client()`는 그 서버의 인증서를 신뢰하도록 준비되어 있습니다.
인증서 검증을 전부 끄는 설정이 필요하지 않습니다.
[Server.Client](https://pkg.go.dev/net/http/httptest#Server.Client)의 신뢰 설정을 사용합니다.

이 신뢰는 실습 서버를 위한 것입니다.
실제 서비스의 인증서 오류를 없애려고 검증을 끄는 방법으로 옮기지 않습니다.
서버 이름, 인증서 체인, 운영 환경의 신뢰 저장소를 올바르게 구성해야 합니다.

### 본문을 버리는 것과 한도 검증의 차이

`inspect`는 상태와 헤더만 필요하므로 작은 범위의 본문을 `io.Discard`로 보냅니다.
여기의 1024바이트 제한은 버릴 데이터의 양을 제한하는 정책입니다.
본문 전체가 1024바이트 이내라는 검증 결과를 반환하지는 않습니다.
전체 크기를 판정해야 한다면 앞의 `readAtMost`처럼 한 바이트를 더 읽어 구분합니다.

## 10. 직접 바꿔 보는 실습

본문의 JSON 예제를 개인 폴더에 복사합니다.
기존 파일이 있다면 덮어쓰지 않습니다.
테스트 파일은 복사하지 않으므로 자신의 변경은 직접 실행하고 경계 입력으로 확인합니다.

Windows PowerShell:

```powershell
$practiceDir = "_practice/21-http-client/json-lab"
New-Item -ItemType Directory -Force $practiceDir | Out-Null
$practiceMain = Join-Path $practiceDir "main.go"
if (Test-Path -LiteralPath $practiceMain) {
    Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
} else {
    Copy-Item examples/21-http-client/json-response/main.go $practiceMain
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
practice_dir="_practice/21-http-client/json-lab"
mkdir -p "$practice_dir"
if [ -e "$practice_dir/main.go" ]; then
    printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
else
    cp examples/21-http-client/json-response/main.go "$practice_dir/main.go"
fi
```

저장소 루트에서 실행합니다.

```text
go run ./_practice/21-http-client/json-lab
go fmt ./_practice/21-http-client/json-lab
```

### 실험 A: 유효한 최댓값

실험 서버가 보내는 `Minutes: 25`를 `Minutes: 180`으로 바꿉니다.
예상 출력은 `학습: Go, 180분`입니다.
조회 함수의 코드를 바꾸지 않고 서버 데이터만 바꿔 확인합니다.

### 실험 B: HTTP 성공과 데이터 실패

같은 값을 181로 바꿉니다.
서버의 HTTP 상태는 계속 200이지만 프로그램은 학습 데이터 범위 오류를 보고합니다.
이것은 의도한 실패 실행이며 종료 코드도 성공이 아닙니다.
상태 검사와 데이터 검사가 서로 다른 역할이라는 근거를 적어 보세요.

### 실험 C: 형식 헤더

분량을 25로 복구하고 `Content-Type`을 `text/plain`으로 바꿉니다.
JSON 모양의 본문이어도 이 함수의 미디어 타입 계약에 맞지 않으므로 거절됩니다.
확인 후 `application/json; charset=utf-8`로 되돌립니다.

### 실험 D: 제목에 공백만 넣기

`Title: "Go"`를 `Title: "   "`로 바꾸어 봅니다.
문법상 정상 JSON이어도 업무 조건 때문에 실패해야 합니다.
확인 후 제목과 분량을 `Go`, 25로 모두 복구하고 다시 실행합니다.

## 11. 자주 생기는 오류와 확인 순서

| 증상 | 가능한 원인 | 확인·수정 |
| --- | --- | --- |
| 404인데 `Do` 오류가 nil | HTTP 응답은 정상 수신됨 | `StatusCode`를 별도로 검사 |
| 오류 뒤 nil 포인터 문제가 남 | 응답을 얻기 전에 본문 접근 | `err` 검사 후 `defer Close` |
| 연결이나 파일 디스크립터 사용이 늘어남 | 응답 본문을 닫지 않음 | 성공·실패 모든 응답 경로의 닫기 확인 |
| 한도가 있는데 잘린 본문을 성공으로 사용 | `limit`만 읽고 초과 여부 미검사 | `limit+1` 읽기 후 길이 판정 |
| 두 JSON 객체가 허용됨 | 디코딩 한 번으로 끝냄 | 두 번째 디코딩의 EOF 확인 |
| `&`가 든 검색어가 달라짐 | URL 문자열을 직접 결합 | `url.Values.Encode` 사용 |
| 취소 테스트가 가끔 다른 결과 | 고정 sleep에 시작 순서를 의존 | 채널로 실제 시작 알림 |
| 실습 TLS에서 인증서 오류 | 일반 클라이언트로 학습 인증서 사용 | `server.Client()` 사용 |
| 본문 읽기 중 타임아웃 | 전체 제한 시간이 아직 적용 중 | 시간 제한 범위와 본문 크기 확인 |

오류를 좁힐 때 URL 전체와 헤더 전체를 무조건 출력하지 않습니다.
필요한 작업 이름, 상태 코드, 분류 가능한 오류를 먼저 기록합니다.
주소에 민감한 쿼리가 포함될 수 있다면 로그 표현도 따로 정합니다.
이 장의 모든 실습 URL과 데이터는 로컬의 고정 학습용 값입니다.

## 12. 테스트로 확인하는 경계

각 예제 폴더의 `main_test.go`는 문서의 기본 출력을 확인합니다.
`behavior_test.go`는 입력별 동작을 검사합니다.
성공 출력만 맞는 구현이 한도 초과나 취소에서도 안전하다고 단정할 수는 없습니다.

```text
go test -count=1 ./examples/21-http-client/...
go vet ./examples/21-http-client/...
```

| 예제 | 추가로 확인하는 내용 |
| --- | --- |
| `basic-get` | 빈 본문, 정확히 1024, 1025, 상태 실패, 본문 닫기 |
| `query-headers` | 빈 검색어, 한글·기호, 기존 쿼리와 헤더 유지 |
| `status-error` | 200·204·404·503의 계약 |
| `bounded-body` | 0·정확한 한도·초과, UTF-8 바이트, 읽기 오류 |
| `json-response` | 단일 값, 알 수 없는 필드, null, 범위·형식·크기 |
| `cancellation` | 시작 전 취소, 이미 지난 기한, 진행 중 취소 |
| `redirect-policy` | 추가 요청 여부, Location, TLS 테스트 신뢰 |

본문 닫기 검사는 테스트용 `RoundTripper`로 응답을 돌려주고 `Close` 호출 여부를 기록합니다.
외부 네트워크 장애를 만들어 자원 정리를 추정하는 방식보다 조건이 분명합니다.
이 장은 그 인터페이스를 테스트 안에서만 사용하며 실행 예제는 실제 임시 서버를 호출합니다.

Linux에서 race detector를 사용할 수 있다면 다음도 실행합니다.

```text
go test -race ./examples/21-http-client/...
```

## 13. 연습문제

[21장 연습문제](../../exercises/21-http-client/README.md)에서 다음을 구현합니다.
완성 코드는 [해설](../../solutions/21-http-client/README.md)에 분리되어 있습니다.

| 문제 | 확인할 능력 |
| --- | --- |
| 1. 응답과 오류 예측 | 상태와 전송 오류 구분 |
| 2. 이름과 과정 헤더 전달 | 쿼리·헤더 구성 |
| 3. 본문 닫기의 위치 고치기 | 오류 경로와 수명 |
| 4. 작은 다운로드 함수 | 바이트 한도·경계 입력 |
| 5. 학습 JSON 조회 | 상태·미디어 타입·형식·데이터 검사 |
| 6. 취소된 요청 | 컨텍스트 전달과 요청 횟수 확인 |
| 선택 A. 이동 멈추기 | 리다이렉트 정책과 응답 소유권 |

## 핵심 정리와 자기 점검

- 요청 생성은 전송과 다르며, `Do`의 성공도 업무 성공과 다릅니다.
- 응답을 얻은 뒤에는 본문을 닫을 책임이 생깁니다.
- 본문 크기 제한은 헤더 검사와 별개로 실제 읽기에 적용합니다.
- JSON 문법, 값의 개수, 필드 범위는 각각 확인합니다.
- 시간 제한과 취소를 요청에 연결하고 정리 완료를 관찰합니다.
- 클라이언트 설정은 사용 전에 정하고 요청마다 재사용합니다.

### 확인 질문

1. 404 응답에서 `err == nil`이 가능한 이유는 무엇인가요?
2. `defer resp.Body.Close()`를 `Do` 오류 검사보다 먼저 쓰면 어떤 문제가 있나요?
3. 한도 10바이트인 본문을 왜 11바이트까지 읽나요?
4. `Content-Length`가 10이라는 정보만으로 실제 읽기 제한을 생략해도 될까요?
5. JSON을 한 번 디코딩하면 객체 두 개를 거절할 수 있나요?
6. `DisallowUnknownFields`가 제목의 공백도 검사하나요?
7. `Client.Timeout`은 헤더를 받은 순간 끝나나요?
8. 컨텍스트 취소는 임의의 무한 반복을 강제로 끝내나요?
9. TLS 실습에서 `server.Client()`를 쓰는 이유는 무엇인가요?
10. 리다이렉트를 따르지 않을 때 받은 응답 본문은 누가 닫나요?

<details>
<summary>답과 설명</summary>

1. 서버와의 HTTP 대화는 성공했으며 상태가 요청의 실패를 나타내기 때문입니다.
2. 응답이 없을 수 있어 nil 접근이 발생합니다. 오류가 없는 뒤에 본문을 사용합니다.
3. 정확히 10인 경우와 더 긴 본문을 구분하기 위해서입니다.
4. 생략하지 않습니다. 실제 본문 스트림에도 제한을 적용합니다.
5. 첫 값 뒤에 다른 값이 남을 수 있습니다. 두 번째 디코딩의 EOF를 확인합니다.
6. 알 수 없는 필드만 검사합니다. 값의 의미는 별도 검증이 필요합니다.
7. 아닙니다. 응답 본문 읽기에도 전체 제한이 적용됩니다.
8. 아닙니다. 작업이 취소 신호를 관찰하거나 취소를 지원하는 API에 전달해야 합니다.
9. 해당 테스트 서버의 TLS 인증서를 신뢰하도록 구성되어 있기 때문입니다.
10. 오류 없이 응답을 받은 호출자가 닫습니다. 이동 중지 정책이 그 책임을 없애지 않습니다.

</details>

### 완료 체크리스트

- [ ] 일곱 예제를 실행하고 종료되는 것을 확인했다.
- [ ] 상태 오류와 전송 오류를 다른 경로로 설명했다.
- [ ] 정확한 한도와 한 바이트 초과를 구분했다.
- [ ] 잘못된 JSON·추가 값·범위 오류를 각각 확인했다.
- [ ] 요청 취소와 서버 작업 종료를 연결해서 추적했다.
- [ ] 개인 실습의 기본 값을 복구했다.
- [ ] 기본 문제 여섯 개와 경계 입력을 확인했다.

## 공식 참고 자료

- [net/http](https://pkg.go.dev/net/http) — 클라이언트·요청·응답
- [Client.Do](https://pkg.go.dev/net/http#Client.Do) — 오류와 본문 소유권
- [Client](https://pkg.go.dev/net/http#Client) — 타임아웃·리다이렉트·재사용
- [httptest](https://pkg.go.dev/net/http/httptest) — 임시 HTTP·TLS 서버
- [io.LimitReader](https://pkg.go.dev/io#LimitReader) — 읽기 제한
- [net/url](https://pkg.go.dev/net/url) — URL 구성과 쿼리 인코딩
- [encoding/json](https://pkg.go.dev/encoding/json) — 디코딩과 호환 규칙
- [context](https://pkg.go.dev/context) — 요청 기한과 취소

확인일: 2026-11-09. 예제는 Go 1.27.3 기준이며 자동 검증은 외부 HTTP 서비스 없이 실행합니다.

## 다음 학습

다음 [22장 HTTP 서버와 라우팅](../22-http-server/README.md)에서는 요청을 받는 쪽을 구현합니다.
클라이언트가 기대하는 상태·헤더·본문을 핸들러가 어떻게 구성하고, 여러 요청을 어떻게 처리하는지 배웁니다.

| 이전 | 목차 | 실습 | 해설 | 다음 |
| --- | --- | --- | --- | --- |
| [20장 공유 상태와 동기화](../20-sync/README.md) | [전체 목차](../README.md) | [문제](../../exercises/21-http-client/README.md) | [해설](../../solutions/21-http-client/README.md) | [22장](../22-http-server/README.md) |
