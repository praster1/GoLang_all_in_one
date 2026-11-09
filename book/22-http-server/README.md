# 22. HTTP 서버와 라우팅

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 중급 | [21장 HTTP 클라이언트](../21-http-client/README.md), JSON·오류·컨텍스트·동기화 | Go, 편집기, 터미널 |

21장에서는 요청을 보내고 응답을 검사했습니다. 이번에는 그 요청을 받는 서버를 만듭니다. 서버는 주소가 맞는지 찾고, 메서드와 입력을 검사하고, 상태·헤더·본문으로 결과를 표현합니다. 입력이 잘못된 경우에도 무엇을 거절했는지 일정한 규칙으로 알려 주어야 합니다.

서버 프로그램은 요청 하나를 처리하고 끝나는 일반 함수와 실행 수명이 다릅니다. 여러 요청을 받는 동안 살아 있어야 하며, 종료할 때도 이미 진행 중인 요청을 고려해야 합니다. 이 장의 자동 예제는 실험 요청을 보낸 뒤 스스로 종료하므로 따로 서버를 켜 둘 필요가 없습니다.

## 학습 목표

- `http.Handler`와 `HandlerFunc`의 관계를 설명하고 핸들러를 만든다.
- `ServeMux`로 메서드·경로를 연결하고 경로 변수를 읽는다.
- 요청 본문의 크기·형식·데이터를 검증하여 다른 실패를 구분한다.
- 상태와 헤더를 본문보다 먼저 정하고 JSON 응답을 작성한다.
- 요청 취소를 작업에 전달하고 서버의 정상 종료를 기다린다.

## 준비와 실행 예제

모든 명령은 **저장소 루트**에서 실행합니다. 루트의 [`go.mod`](../../go.mod)를 사용하며 새 모듈을 만들지 않습니다. 기본 예제에는 외부 계정, 외부 서버와 데이터베이스가 필요하지 않습니다.

| 예제 | 주제 | 실행 명령 |
| --- | --- | --- |
| [hello-handler](../../examples/22-http-server/hello-handler/main.go) | 핸들러와 첫 응답 | `go run ./examples/22-http-server/hello-handler` |
| [method-routing](../../examples/22-http-server/method-routing/main.go) | 메서드·경로·405 | `go run ./examples/22-http-server/method-routing` |
| [path-values](../../examples/22-http-server/path-values/main.go) | 경로 변수와 정수 검증 | `go run ./examples/22-http-server/path-values` |
| [json-create](../../examples/22-http-server/json-create/main.go) | 제한된 JSON 요청 | `go run ./examples/22-http-server/json-create` |
| [json-errors](../../examples/22-http-server/json-errors/main.go) | 구조화한 실패 응답 | `go run ./examples/22-http-server/json-errors` |
| [request-context](../../examples/22-http-server/request-context/main.go) | 서버 작업 취소 | `go run ./examples/22-http-server/request-context` |
| [graceful-shutdown](../../examples/22-http-server/graceful-shutdown/main.go) | 수신 시작과 정상 종료 | `go run ./examples/22-http-server/graceful-shutdown` |

`httptest.NewServer` 예제는 요청을 마치면 `defer server.Close()`로 끝납니다. 마지막 예제는 직접 리스너와 `http.Server`를 구성하고 `Shutdown`을 호출합니다. 실행이 예상과 달리 멈추면 터미널에서 Ctrl+C로 중지하고 테스트의 시간 제한과 종료 경로를 확인합니다.

## 1. 핸들러는 요청 하나를 처리하는 함수입니다

`http.Handler`는 다음 메서드를 가진 인터페이스입니다. 아래는 이해를 위한 **타입 선언 발췌**이며 별도 실행 파일에 추가하지 않습니다.

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

실제 타입은 `net/http` 패키지에 이미 정의되어 있습니다. 핸들러는 응답을 쓰는 `ResponseWriter`와 요청을 나타내는 `*Request`를 받습니다. 서버는 요청이 들어오면 해당 핸들러의 `ServeHTTP`를 호출합니다. [Handler](https://pkg.go.dev/net/http#Handler)의 계약을 참고합니다.

일반 함수를 핸들러로 쓰고 싶다면 `http.HandlerFunc`를 사용합니다. 같은 모양의 함수에 `ServeHTTP` 메서드를 제공하는 어댑터 타입입니다. 라우터의 `HandleFunc` 메서드는 이 형태의 함수를 등록하기 편하게 해 줍니다. [HandlerFunc](https://pkg.go.dev/net/http#HandlerFunc)를 참고합니다.

| 값 | 역할 | 누가 주는가? |
| --- | --- | --- |
| `r.Method` | 요청 메서드 | 클라이언트 |
| `r.URL.Path` | 요청 경로 | 클라이언트 |
| `r.Header` | 요청 부가 정보 | 클라이언트 |
| `r.Body` | 요청 본문 스트림 | HTTP 서버 |
| `w.Header()` | 보낼 응답 헤더 | 핸들러가 설정 |
| `w.WriteHeader` | 보낼 응답 상태 | 핸들러가 결정 |
| `w.Write` | 보낼 응답 본문 | 핸들러가 작성 |

클라이언트의 `resp.Body`와 서버의 `r.Body`를 구분합니다. 앞은 서버가 보낸 응답을 읽는 스트림이고, 뒤는 클라이언트가 보낸 요청을 읽는 스트림입니다. 이름이 같아도 어느 방향의 데이터를 소유하고 있는지 다릅니다.

핸들러가 반환한 뒤에는 그 요청의 `ResponseWriter`를 계속 사용하지 않습니다. 별도 고루틴을 시작하고 핸들러는 먼저 반환한 뒤 나중에 응답을 쓰는 방식은 계약에 맞지 않습니다. 필요한 작업을 기다리거나, 응답과 분리된 작업 큐로 책임을 따로 설계해야 합니다. [ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)의 수명 규칙을 확인합니다.

## 2. 첫 응답 만들기

저장 경로: [hello-handler/main.go](../../examples/22-http-server/hello-handler/main.go)

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
	"strings"
	"time"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, "안녕하세요, HTTP!\n"); err != nil {
			return
		}
	})
	return mux
}

func call(client *http.Client, address, method, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil {
		return 0, "", err
	}
	if len(data) > 4096 {
		return 0, "", fmt.Errorf("응답 한도 초과")
	}
	return resp.StatusCode, string(data), nil
}

func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/hello", ""},
	} {
		status, text, err := call(client, server.URL+tc.path, tc.method, tc.body)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s -> %d: %s\n", tc.method, tc.path, status, strings.TrimSpace(text))
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
go run ./examples/22-http-server/hello-handler
```

예상 출력:

```text
GET /hello -> 200: 안녕하세요, HTTP!
```

### 실행을 세 부분으로 나누어 읽기

`newHandler`는 요청 처리 규칙을 만듭니다. `call`은 21장에서 배운 클라이언트 동작으로 응답을 읽습니다. `run`은 임시 서버의 시작·실험·종료를 연결합니다.

| 부분 | 서버의 책임인가? | 설명 |
| --- | --- | --- |
| `newHandler` | 예 | 경로와 응답 구성 |
| `httptest.NewServer` | 실습 장치 | 임시 수신 서버 시작 |
| `call` | 아니요 | 클라이언트 역할의 관찰 도구 |
| `fmt.Printf` | 아니요 | 받은 응답을 터미널에 표시 |

핸들러 안에서 `fmt.Println`을 호출해도 HTTP 응답이 되지 않습니다. 그 호출은 서버 프로세스의 표준 출력에 씁니다. 클라이언트에게 데이터를 보내려면 `w`에 써야 합니다. 예제는 `io.WriteString(w, ...)`를 사용합니다.

### 기본 상태 200

이 핸들러는 `WriteHeader`를 따로 호출하지 않았습니다. 본문을 처음 쓸 때 최종 상태가 아직 정해지지 않았다면 200이 사용됩니다. 그래서 응답은 200과 인사말입니다. [ResponseWriter.Write](https://pkg.go.dev/net/http#ResponseWriter)의 규칙입니다.

이 규칙은 성공 응답을 짧게 쓰기에 편리합니다. 그러나 실패를 표현하려면 본문을 쓰기 전에 실패 상태를 정해야 합니다. 한번 보낸 최종 상태를 나중에 다른 상태로 고치는 방식으로 작성하지 않습니다.

### 공유 전역 라우터를 사용하지 않은 이유

`http.NewServeMux()`로 라우터를 새로 만들고 반환했습니다. 서로 다른 예제와 테스트가 같은 전역 라우터에 등록되어 충돌하지 않도록 한 선택입니다. 테스트마다 필요한 핸들러를 독립적으로 준비할 수도 있습니다. 이 구조는 HTTP 서비스가 커질 때 의존성을 전달하기에도 편합니다.

## 3. 메서드와 경로를 함께 라우팅하기

라우팅은 요청을 어느 처리 코드로 보낼지 정하는 일입니다. 같은 경로라도 GET은 조회, POST는 입력을 받는 작업으로 구분할 수 있습니다. 메서드가 없는 패턴은 모든 메서드에 대응하므로 의도를 명확히 적는 것이 좋습니다.

저장 경로: [method-routing/main.go](../../examples/22-http-server/method-routing/main.go)

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
	"strings"
	"time"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "목록"); err != nil {
			return
		}
	})
	mux.HandleFunc("POST /lessons", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if _, err := io.WriteString(w, "생성됨"); err != nil {
			return
		}
	})
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "학습 %s", r.PathValue("id")); err != nil {
			return
		}
	})
	return mux
}

func call(client *http.Client, address, method, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil {
		return 0, "", err
	}
	if len(data) > 4096 {
		return 0, "", fmt.Errorf("응답 한도 초과")
	}
	return resp.StatusCode, string(data), nil
}

func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/", ""},
		{http.MethodPost, "/lessons", ""},
		{http.MethodGet, "/lessons/7", ""},
		{http.MethodDelete, "/lessons/7", ""},
		{http.MethodGet, "/missing", ""},
	} {
		status, text, err := call(client, server.URL+tc.path, tc.method, tc.body)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s -> %d: %s\n", tc.method, tc.path, status, strings.TrimSpace(text))
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
go run ./examples/22-http-server/method-routing
```

예상 출력:

```text
GET / -> 200: 목록
POST /lessons -> 201: 생성됨
GET /lessons/7 -> 200: 학습 7
DELETE /lessons/7 -> 405: Method Not Allowed
GET /missing -> 404: 404 page not found
```

### 세 패턴을 표로 읽기

| 패턴 | 허용 요청 | 읽는 값 |
| --- | --- | --- |
| `GET /{$}` | 루트의 GET·HEAD | 별도 경로 변수 없음 |
| `POST /lessons` | 해당 경로의 POST | 이 예제에서는 본문을 사용하지 않음 |
| `GET /lessons/{id}` | 한 경로 조각을 가진 GET·HEAD | `r.PathValue("id")` |

`{$}`는 URL 경로의 끝을 뜻하는 특수 와일드카드입니다. 따라서 `/{$}`는 `/`에만 대응합니다. 반면 `/` 패턴은 경로 아래의 요청까지 넓게 대응합니다. 루트 안내 핸들러가 의도하지 않은 모든 경로를 받지 않도록 구분해야 합니다. [ServeMux 패턴](https://pkg.go.dev/net/http#ServeMux)의 규칙입니다.

### 404와 405를 구분하기

| 요청 | 경로가 있는가? | 메서드가 맞는가? | 결과 |
| --- | --- | --- | --- |
| `GET /lessons/7` | 예 | 예 | 200 |
| `DELETE /lessons/7` | 예 | 아니요 | 405 |
| `GET /missing` | 아니요 | 해당 없음 | 404 |

메서드가 맞지 않으면 `ServeMux`는 405와 허용 메서드를 나타내는 `Allow` 헤더를 작성합니다. 같은 처리 내용을 핸들러마다 중복된 메서드 조건으로 적지 않아도 됩니다. 다만 원하는 API의 오류 본문 형식까지 자동으로 통일되는 것은 아닙니다. 기본 404·405 본문은 이 예제의 성공 응답과 다른 텍스트입니다.

### GET 패턴은 HEAD도 받습니다

`GET /hello`는 HEAD 요청에도 대응합니다. HEAD는 응답의 헤더를 조회하되 전송되는 본문을 받지 않는 메서드입니다. 실제 HTTP 서버를 통한 테스트에서는 `HEAD /hello`의 본문이 비어 있음을 확인합니다. [ServeMux](https://pkg.go.dev/net/http#ServeMux)의 GET·HEAD 매칭 규칙을 참고합니다.

핸들러를 `ResponseRecorder`에 직접 호출하는 검사와 실제 HTTP 전송은 구분합니다. 직접 호출에서는 핸들러가 쓴 바이트를 기록하므로, 전송 계층이 처리하는 HEAD 본문 생략을 똑같이 관찰하지 않을 수 있습니다. 이 차이를 23장의 테스트 범위 선택에서 다시 다룹니다.

### 겹치는 패턴은 등록 순서만으로 해결되지 않습니다

`ServeMux`는 더 구체적인 패턴을 선택하는 규칙을 가집니다. 서로 충돌하는 패턴을 등록하면 시작 단계에서 패닉이 발생할 수 있습니다. 기존 패턴에 새 경로를 추가할 때는 실제 요청의 대응 범위를 비교합니다. 등록 순서를 바꾸어 우연히 동작하게 만드는 방식으로 해결하지 않습니다. [ServeMux 우선순위](https://pkg.go.dev/net/http#ServeMux)를 확인합니다.

## 4. 경로 값은 문자열입니다

`/lessons/7`의 `7`을 경로 변수로 읽었다고 자동으로 정수 검증이 되지는 않습니다. `/lessons/x`도 패턴 모양 자체는 같습니다. 경로를 찾는 일과 경로 값의 의미를 검사하는 일을 구분합니다.

저장 경로: [path-values/main.go](../../examples/22-http-server/path-values/main.go)

전체 코드:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"
)

type Lesson struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 {
			http.Error(w, "ID는 양수 정수여야 합니다", http.StatusBadRequest)
			return
		}
		lesson := Lesson{ID: id, Title: "Go"}
		data, err := json.Marshal(lesson)
		if err != nil {
			http.Error(w, "응답 생성 실패", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(append(data, '\n')); err != nil {
			return
		}
	})
	return mux
}

func call(client *http.Client, address, method, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil {
		return 0, "", err
	}
	if len(data) > 4096 {
		return 0, "", fmt.Errorf("응답 한도 초과")
	}
	return resp.StatusCode, string(data), nil
}

func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/lessons/7", ""},
		{http.MethodGet, "/lessons/0", ""},
		{http.MethodGet, "/lessons/x", ""},
		{http.MethodGet, "/lessons/7/extra", ""},
	} {
		status, text, err := call(client, server.URL+tc.path, tc.method, tc.body)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s -> %d: %s\n", tc.method, tc.path, status, strings.TrimSpace(text))
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
go run ./examples/22-http-server/path-values
```

예상 출력:

```text
GET /lessons/7 -> 200: {"id":7,"title":"Go"}
GET /lessons/0 -> 400: ID는 양수 정수여야 합니다
GET /lessons/x -> 400: ID는 양수 정수여야 합니다
GET /lessons/7/extra -> 404: 404 page not found
```

### 라우팅, 변환, 업무 조건

| 단계 | 입력 예 | 검사 |
| --- | --- | --- |
| 라우팅 | `/lessons/7` | `lessons` 뒤 한 경로 조각인가? |
| 문자열 추출 | `"7"` | `PathValue("id")` |
| 정수 변환 | `7` | `strconv.Atoi` 오류가 없는가? |
| 범위 판정 | `7` | 1 이상인가? |
| 응답 구성 | 구조체 | JSON으로 표현할 수 있는가? |

`Atoi`는 현재 플랫폼의 `int`로 표현할 수 없는 큰 값에도 오류를 반환합니다. 오류를 무시하면 잘못된 ID를 기본값처럼 다룰 수 있습니다. [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)의 결과를 확인한 뒤 범위 조건을 적용합니다.

이 예제는 양수 ID의 모양을 검증한 뒤 `Go` 제목을 붙여 응답합니다. 실제 저장소에서 해당 ID가 존재하는지는 조회하지 않습니다. 유효한 ID와 존재하는 데이터는 서로 다른 조건입니다. 다음 오류 응답 예제에서 400과 404의 역할을 분리합니다.

### 경로 조각과 전체 경로

`{id}`는 기본적으로 한 경로 조각에 대응합니다. 따라서 `/lessons/7/extra`는 이 패턴에 맞지 않습니다. 여러 조각을 받을 때 사용하는 `{name...}` 형태도 있지만, 필요가 있을 때 명시적으로 선택합니다. [Request.PathValue](https://pkg.go.dev/net/http#Request.PathValue)는 라우터가 선택한 값을 제공합니다.

테스트에서 핸들러 함수를 라우터 없이 바로 호출한다면 경로 값이 자동으로 채워지지 않을 수 있습니다. 이 장의 테스트는 `newHandler()`가 반환한 라우터를 통해 요청을 보냅니다. 라우팅과 변환 계약을 함께 검사하려는 선택입니다.

## 5. JSON 입력을 받아 새 응답 만들기

요청 본문은 신뢰할 수 있는 Go 구조체가 아닙니다. 클라이언트는 잘못된 형식, 너무 큰 데이터, 정의하지 않은 필드와 범위 밖 값을 보낼 수 있습니다. 서버가 읽기와 처리의 경계를 정해야 합니다.

다음 예제는 요청을 검증하고 생성 응답의 형식을 보여 줍니다. 저장소를 갖춘 CRUD 구현은 아니며, 매 요청에 학습용 ID 1을 붙입니다. 데이터의 저장·중복·조회·삭제는 후속 종합 API 프로젝트에서 연결합니다. 여기서는 입력 처리와 응답 계약 자체를 완성합니다.

저장 경로: [json-create/main.go](../../examples/22-http-server/json-create/main.go)

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

type LessonInput struct {
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

type Lesson struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

func readInput(w http.ResponseWriter, r *http.Request) (LessonInput, error) {
	var input LessonInput
	r.Body = http.MaxBytesReader(w, r.Body, 256)
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return input, err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return input, fmt.Errorf("JSON 값은 하나여야 함")
	}
	return input, nil
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /lessons", func(w http.ResponseWriter, r *http.Request) {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			http.Error(w, "JSON 형식이 필요합니다", http.StatusUnsupportedMediaType)
			return
		}
		input, err := readInput(w, r)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				http.Error(w, "본문이 너무 큽니다", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "잘못된 JSON입니다", http.StatusBadRequest)
			}
			return
		}
		title := strings.TrimSpace(input.Title)
		if title == "" || input.Minutes < 1 || input.Minutes > 180 {
			http.Error(w, "학습 값의 범위를 확인하세요", http.StatusUnprocessableEntity)
			return
		}
		lesson := Lesson{ID: 1, Title: title, Minutes: input.Minutes}
		data, err := json.Marshal(lesson)
		if err != nil {
			http.Error(w, "응답 생성 실패", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/lessons/1")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write(append(data, '\n')); err != nil {
			return
		}
	})
	return mux
}

func call(client *http.Client, address, method, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil {
		return 0, "", err
	}
	if len(data) > 4096 {
		return 0, "", fmt.Errorf("응답 한도 초과")
	}
	return resp.StatusCode, string(data), nil
}

func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/lessons", `{"title":" Go ","minutes":25}`},
		{http.MethodPost, "/lessons", `{"title":"Go","minutes":0}`},
		{http.MethodPost, "/lessons", `{"title":"Go","minutes":25}{}`},
	} {
		status, text, err := call(client, server.URL+tc.path, tc.method, tc.body)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s -> %d: %s\n", tc.method, tc.path, status, strings.TrimSpace(text))
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
go run ./examples/22-http-server/json-create
```

예상 출력:

```text
POST /lessons -> 201: {"id":1,"title":"Go","minutes":25}
POST /lessons -> 422: 학습 값의 범위를 확인하세요
POST /lessons -> 400: 잘못된 JSON입니다
```

### 요청을 읽는 함수와 응답하는 핸들러

`readInput`은 본문을 작은 크기로 제한한 뒤 JSON을 해석합니다. 핸들러는 미디어 타입, 해석 오류, 업무 범위를 HTTP 상태로 바꿉니다. 읽기 함수가 모든 HTTP 상태를 직접 결정하게 만들지 않아 검사 단계를 구분할 수 있습니다.

| 단계 | 조건 | 실패 상태 |
| --- | --- | --- |
| 미디어 타입 | `application/json` | 415 |
| 본문 한도 | 최대 256바이트 | 413 |
| JSON 구조 | 알려진 필드의 값 하나 | 400 |
| 업무 값 | 제목 존재, 1~180분 | 422 |
| 응답 생성 | 유효한 학습 데이터 | 201 |

상태 코드 선택은 이 API의 계약입니다. 입력 값 오류를 모두 400으로 표현하는 설계도 가능하지만, 이 예제는 문법 오류와 값 범위를 구분합니다. 클라이언트와 테스트가 같은 계약을 사용하도록 문서에 적어 둡니다.

### `MaxBytesReader`와 `LimitReader`의 차이

21장에서는 응답 본문을 읽을 때 `LimitReader`와 한 바이트 추가 읽기를 사용했습니다. 서버의 요청 본문에는 `http.MaxBytesReader`가 제공됩니다. 한도를 넘는 읽기에서 `*http.MaxBytesError`를 반환하므로 413 응답을 구분할 수 있습니다. [MaxBytesReader](https://pkg.go.dev/net/http#MaxBytesReader)의 계약을 참고합니다.

읽기 한도는 요청 헤더에 적힌 `Content-Length`만 확인하는 것과 다릅니다. 서버가 실제로 읽는 스트림에 한도를 적용합니다. 입력은 256바이트까지 허용되며, 뒤의 공백도 본문의 바이트 수에 포함됩니다.

### 요청 본문 닫기

서버는 요청의 본문을 처리 수명에 맞춰 닫습니다. 이 예제의 `readInput`도 자신이 만든 `MaxBytesReader`를 `defer`로 닫아 읽기 범위를 분명히 합니다. 그 래퍼의 닫기는 아래의 원래 reader에도 전달됩니다. [Request.Body](https://pkg.go.dev/net/http#Request)와 [MaxBytesReader](https://pkg.go.dev/net/http#MaxBytesReader)를 참고합니다.

미디어 타입이 잘못되어 본문을 읽지 않는 경로는 핸들러가 오류 응답 후 반환합니다. 클라이언트 응답의 `Body.Close` 책임과 서버 요청 본문의 수명 규칙을 혼동하지 않습니다. 읽지 않은 임의 크기의 본문을 끝까지 메모리에 모으는 방식으로 정리하지 않습니다.

### 한 JSON 값과 업무 검증

첫 디코딩은 구조체의 모양을 검사합니다. 두 번째 디코딩은 EOF를 확인하여 추가 값과 뒤의 쓰레기 문자를 거절합니다. `DisallowUnknownFields`는 정의하지 않은 필드를 거절하지만 중복 키를 금지하는 별도 검증은 아닙니다. 이 장에서도 기본 `encoding/json`의 필드 매칭 규칙을 따릅니다. [Decoder](https://pkg.go.dev/encoding/json#Decoder)의 동작을 확인합니다.

제목은 앞뒤 공백을 제거한 값을 응답에 사용합니다. 이는 21장의 “공백만인 제목인지 검사하되 원문 유지”와 다른 정규화 정책입니다. 이번 생성 응답은 정규화한 값을 돌려준다고 계약했습니다.

| 입력 제목 | 입력 분량 | 처리 결과 |
| --- | --- | --- |
| `" Go "` | 25 | 제목 `Go`로 201 |
| `"   "` | 25 | 422 |
| `"Go"` | 0 | 422 |
| `"Go"` | 1 | 201 |
| `"Go"` | 180 | 201 |
| `"Go"` | 181 | 422 |

`null`은 문법상 JSON이지만 이 데이터 모델에서는 유효한 요청이 아닙니다. 영값으로 남은 필드가 업무 조건을 만족하지 않아 422가 됩니다. 빈 본문은 첫 디코딩에서 실패하여 400입니다. 이 차이도 테스트에 포함했습니다.

### 201과 Location

성공 응답은 201과 `Location: /lessons/1`을 보냅니다. `Location`은 이 실험에서 생성 응답 형식을 설명하기 위한 값입니다. 현재 예제는 그 경로를 조회하는 저장소 기능까지 제공하지 않습니다. 클라이언트가 실제로 후속 조회를 할 수 있는 완전한 서비스에서는 저장과 조회 계약을 함께 구현해야 합니다.

## 6. 상태, 헤더, 본문을 쓰는 순서

응답을 쓰기 시작하면 일부 선택은 되돌릴 수 없습니다. 기본 순서는 데이터를 준비하고, 헤더를 정하고, 최종 상태를 정하고, 본문을 쓰는 것입니다. 이미 본문을 쓴 뒤 오류가 났다고 다른 상태의 JSON을 추가로 쓰면 응답이 섞일 수 있습니다.

저장 경로: [json-errors/main.go](../../examples/22-http-server/json-errors/main.go)

전체 코드:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"
)

type Problem struct {
	Error string `json:"error"`
}

func writeProblem(w http.ResponseWriter, status int, message string) {
	data, err := json.Marshal(Problem{Error: message})
	if err != nil {
		http.Error(w, "응답 생성 실패", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 {
			writeProblem(w, http.StatusBadRequest, "invalid_id")
			return
		}
		if id != 1 {
			writeProblem(w, http.StatusNotFound, "lesson_not_found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, "{\"id\":1,\"title\":\"Go\"}\n"); err != nil {
			return
		}
	})
	return mux
}

func call(client *http.Client, address, method, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil {
		return 0, "", err
	}
	if len(data) > 4096 {
		return 0, "", fmt.Errorf("응답 한도 초과")
	}
	return resp.StatusCode, string(data), nil
}

func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/lessons/x", ""},
		{http.MethodGet, "/lessons/2", ""},
		{http.MethodGet, "/lessons/1", ""},
	} {
		status, text, err := call(client, server.URL+tc.path, tc.method, tc.body)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s -> %d: %s\n", tc.method, tc.path, status, strings.TrimSpace(text))
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
go run ./examples/22-http-server/json-errors
```

예상 출력:

```text
GET /lessons/x -> 400: {"error":"invalid_id"}
GET /lessons/2 -> 404: {"error":"lesson_not_found"}
GET /lessons/1 -> 200: {"id":1,"title":"Go"}
```

### 400과 404의 의미

| 요청 | 판정 | 응답 코드 |
| --- | --- | --- |
| `/lessons/x` | ID를 정수로 해석할 수 없음 | 400 |
| `/lessons/0` | ID 범위가 잘못됨 | 400 |
| `/lessons/2` | 모양은 유효하지만 실험 데이터 없음 | 404 |
| `/lessons/1` | 실험 데이터 있음 | 200 |

실패 본문은 `{"error":"..."}`라는 구조를 사용합니다. 내부 오류 문자열을 그대로 노출하는 대신 호출자가 구분할 수 있는 코드를 정했습니다. 사람에게 보이는 설명이나 요청 식별자를 추가하는 설계는 23장에서 이어갑니다.

### 먼저 직렬화하는 이유

`json.Marshal`로 응답 데이터를 준비한 다음 헤더와 상태를 씁니다. 직렬화 자체에 실패했다면 아직 응답을 시작하지 않았으므로 500을 보낼 수 있습니다. 반면 네트워크에 쓰는 도중 오류가 발생하면 이미 보낸 상태를 바꿀 수 없습니다. 이때 같은 writer에 다른 JSON을 덧붙이는 방식으로 복구하지 않습니다.

이 예제의 `Problem`은 문자열 하나라 일반적인 직렬화 오류가 생길 요소는 적습니다. 그래도 API 호출의 오류를 확인하는 흐름을 유지합니다. 더 일반적인 값이나 커스텀 직렬화 메서드를 쓰는 경우에는 그 실패 가능성도 고려해야 합니다.

### 잘못된 순서의 짧은 예

아래는 **의도적으로 잘못된 발췌 코드**이며 정상 핸들러에 추가하지 않습니다.

```go
fmt.Fprintln(w, "잘못된 입력")
w.WriteHeader(http.StatusBadRequest)
```

첫 줄이 본문을 쓰며 200을 선택할 수 있습니다. 두 번째 줄에서 400으로 되돌리는 것이 아닙니다. 실패를 보낼 때는 `WriteHeader`를 먼저 호출하거나 `http.Error` 같은 helper를 사용합니다. [ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)의 최종 상태 쓰기 규칙을 참고합니다.

`http.Error`도 핸들러 실행을 종료해 주는 함수는 아닙니다. 그 뒤의 성공 코드를 실행하지 않도록 호출한 함수에서 `return`해야 합니다. 연습문제에서 이 흐름을 직접 고칩니다.

### 오류 형식의 적용 범위

이 예제의 `writeProblem`은 직접 등록한 학습 조회의 실패를 JSON으로 만듭니다. 라우터가 자동으로 만드는 알 수 없는 경로의 404와 메서드 오류의 405까지 자동 변환하지 않습니다. API 전체의 실패 형식을 통일하려면 그 범위도 별도 처리해야 합니다. 일부 경로만 보고 “서버의 모든 오류가 JSON”이라고 문서화하지 않습니다.

## 7. 요청의 취소를 실제 작업에 전달하기

클라이언트가 더 이상 기다리지 않는데 서버가 비싼 작업을 계속하면 자원이 낭비될 수 있습니다. 핸들러가 호출하는 함수에도 `r.Context()`를 전달합니다. 그 함수 역시 컨텍스트를 관찰하거나 지원하는 API에 전달해야 합니다.

저장 경로: [request-context/main.go](../../examples/22-http-server/request-context/main.go)

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

func waitForWork(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func run() error {
	started := make(chan struct{})
	finished := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		err := waitForWork(r.Context())
		finished <- err
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		return err
	}
	result := make(chan error, 1)
	go func() {
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		result <- err
	}()
	select {
	case <-started:
		cancel()
	case err := <-result:
		return fmt.Errorf("시작 전 종료: %v", err)
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		return fmt.Errorf("클라이언트 결과: %v", err)
	}
	if err := <-finished; !errors.Is(err, context.Canceled) {
		return fmt.Errorf("서버 결과: %v", err)
	}
	fmt.Println("클라이언트 취소: true")
	fmt.Println("서버 작업 취소: true")
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
go run ./examples/22-http-server/request-context
```

예상 출력:

```text
클라이언트 취소: true
서버 작업 취소: true
```

### 요청 하나의 취소 흐름

| 단계 | 상태 | 신호 |
| --- | --- | --- |
| 요청 도착 | 핸들러 실행 시작 | `started` 닫힘 |
| 클라이언트 취소 | 요청 결과를 더 기다리지 않음 | `cancel()` |
| 서버 요청 컨텍스트 종료 | 작업이 취소를 관찰 | `ctx.Done()` |
| 작업 반환 | `context.Canceled` 반환 | `finished`에 전달 |
| 검증 완료 | 양쪽 종료 확인 | 프로그램 종료 |

실험 작업은 데이터베이스나 외부 API 대신 취소를 기다립니다. 실제 작업 함수도 `ctx`를 첫 인자로 받게 하면 취소를 이어 전달하기 편합니다. 함수 내부에서 새로운 `Background()`를 만들어 상위 취소와 단절하지 않습니다. [Request.Context](https://pkg.go.dev/net/http#Request.Context)를 참고합니다.

클라이언트 연결이 끊어진 뒤에 성공 응답을 쓰려고 하지 않습니다. 이 예제는 작업이 취소를 확인하고 반환하는지를 관찰합니다. 사용자에게 반환할 유효한 HTTP 상태를 새로 만드는 실험은 아닙니다.

### 취소와 공유 데이터는 다른 문제

컨텍스트는 작업 수명을 전달하며 공유 변수에 대한 잠금을 대신하지 않습니다. 여러 요청이 같은 맵이나 카운터를 바꾼다면 20장의 동기화 규칙을 적용해야 합니다. 이 예제의 시작 채널은 요청 한 번을 확인하는 장치이므로 여러 요청에 그대로 재사용하지 않습니다. 서버의 일반 핸들러를 설계할 때 요청별 상태와 서비스 공유 상태를 구분합니다.

핸들러는 요청마다 동시에 실행될 수 있습니다. 여러 요청이 공유하는 값은 Mutex·원자 연산·소유권 구조 등으로 보호합니다. 단순히 테스트 요청을 순서대로 보냈다는 이유로 동시 접근이 안전하다고 결론 내리지 않습니다.

## 8. 서버를 시작하고 정상 종료 기다리기

`http.Server`는 핸들러와 연결 수명 설정을 묶습니다. `Serve`는 리스너에서 요청을 받아 처리하며, 정상 종료 전까지 반환하지 않습니다. 따라서 아래 예제에서는 별도 고루틴에서 실행하고 결과를 채널로 받습니다.

저장 경로: [graceful-shutdown/main.go](../../examples/22-http-server/graceful-shutdown/main.go)

전체 코드:

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
	"strings"
	"time"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "ready\n"); err != nil {
			return
		}
	})
	return mux
}

func run() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           newHandler(),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	defer server.Close()
	serveResult := make(chan error, 1)
	go func() { serveResult <- server.Serve(listener) }()
	client := &http.Client{Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	resp, err := client.Get("http://" + listener.Addr().String() + "/health")
	if err != nil {
		return err
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1025))
	closeErr := resp.Body.Close()
	if readErr != nil {
		return readErr
	}
	if closeErr != nil {
		return closeErr
	}
	if len(body) > 1024 {
		return errors.New("응답 한도 초과")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("상태: %d", resp.StatusCode)
	}
	fmt.Println("응답:", strings.TrimSpace(string(body)))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("종료 대기: %w", err)
	}
	if err := <-serveResult; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("서버 실행: %w", err)
	}
	fmt.Println("정상 종료: true")
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
go run ./examples/22-http-server/graceful-shutdown
```

예상 출력:

```text
응답: ready
정상 종료: true
```

### 수신 주소와 준비 시점

`127.0.0.1:0`은 로컬 루프백 주소와 사용 가능한 포트 선택을 뜻합니다. `net.Listen`이 성공하면 리스너가 준비된 상태입니다. 반환된 주소를 클라이언트가 사용하므로 고정 포트 충돌을 피할 수 있습니다. 실제 선택된 포트는 실행마다 달라질 수 있어 문서의 출력에 넣지 않았습니다. [net.Listen](https://pkg.go.dev/net#Listen)을 참고합니다.

### 시간 제한의 범위

| 설정 | 제한하는 범위 | 예제 값 |
| --- | --- | --- |
| `ReadHeaderTimeout` | 요청 헤더 읽기 | 2초 |
| `ReadTimeout` | 요청 읽기, 본문 포함 | 5초 |
| `WriteTimeout` | 응답 쓰기 | 5초 |
| `IdleTimeout` | 다음 요청을 기다리는 유휴 연결 | 30초 |

값을 0으로 두는 경우와 각 제한의 적용 시점은 [Server](https://pkg.go.dev/net/http#Server) 문서를 확인합니다. 위 값은 짧은 로컬 API 실습을 위한 정책입니다. 큰 업로드, 스트리밍, 장시간 작업이 있다면 같은 값으로 복사하지 말고 동작과 기한을 맞춥니다. 시간 제한은 본문 바이트 한도를 대신하지 않습니다.

### Shutdown의 세 단계

`Shutdown`은 새로운 요청을 받는 리스너를 닫고, 유휴 연결을 정리하고, 진행 중 연결이 쉬는 상태가 되기를 기다립니다. 기다릴 수 있는 시간은 전달한 컨텍스트로 제한합니다. [Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)의 계약에 따른 동작입니다.

| 시점 | 새 요청 | 진행 중 요청 | 호출자 책임 |
| --- | --- | --- | --- |
| 정상 실행 | 받음 | 처리함 | 서버 오류 관찰 |
| Shutdown 시작 | 수신 중단 | 완료를 기다림 | 프로세스를 먼저 종료하지 않음 |
| 모든 요청 완료 | 받지 않음 | 없음 | Shutdown 결과 확인 |
| 종료 기한 초과 | 받지 않음 | 남아 있을 수 있음 | 강제 닫기 등 후속 정책 결정 |

`Shutdown`을 호출하면 `Serve`는 `http.ErrServerClosed`를 반환합니다. 이 오류는 종료 요청에 따른 예상 결과이며 일반적인 실행 장애와 구분합니다. 그러나 `Serve`가 반환했다고 모든 진행 중 요청의 정리가 끝났다고 단정하지 않습니다. `Shutdown` 자체가 반환하는 시점까지 기다려야 합니다.

예제는 `defer server.Close()`도 둡니다. 정상 경로에서는 이미 종료된 서버를 정리하고, 중간 오류 경로에서는 남은 연결을 닫는 마지막 정리 역할을 합니다. 정상 종료 대기와 즉시 닫기는 같은 동작이 아닙니다.

### 종료된 서버 객체는 다시 시작하지 않습니다

`Shutdown`을 호출한 서버는 같은 객체로 다시 `Serve`하는 용도로 재사용하지 않습니다. 새 실행이 필요하면 새 서버를 준비합니다. WebSocket처럼 서버 밖으로 넘겨 관리하는 연결의 수명은 일반 Shutdown과 별도라는 점도 구분합니다. 이 장의 기본 예제에는 그런 연결을 사용하지 않습니다.

## 9. 직접 바꿔 보는 실습

JSON 요청 예제를 개인 폴더에 복사합니다. 기존 파일은 보존하고 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
$practiceDir = "_practice/22-http-server/json-lab"
New-Item -ItemType Directory -Force $practiceDir | Out-Null
$practiceMain = Join-Path $practiceDir "main.go"
if (Test-Path -LiteralPath $practiceMain) {
    Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
} else {
    Copy-Item examples/22-http-server/json-create/main.go $practiceMain
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
practice_dir="_practice/22-http-server/json-lab"
mkdir -p "$practice_dir"
if [ -e "$practice_dir/main.go" ]; then
    printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
else
    cp examples/22-http-server/json-create/main.go "$practice_dir/main.go"
fi
```

```text
go run ./_practice/22-http-server/json-lab
go fmt ./_practice/22-http-server/json-lab
```

### 실험 A: 경계 입력 추가하기

`run`의 요청 목록에 분량 1과 180의 요청을 추가합니다. 두 요청 모두 201이어야 합니다. 0과 181은 422인지 비교하세요. 입력 바이트의 형식은 맞지만 업무 범위를 벗어났다는 차이를 설명합니다.

### 실험 B: 모르는 필드와 값 두 개

정상 객체에 `"extra":true`를 넣어 봅니다. 그다음에는 정상 객체 뒤에 `{}`를 이어 붙입니다. 둘 다 400이지만 실패를 발견한 검증 단계는 다릅니다. 디코더의 필드 검사와 두 번째 값 검사를 찾아 보세요.

### 실험 C: 정확한 바이트 한도

정상 JSON 뒤에 공백을 넣어 전체 길이를 256바이트와 257바이트로 만듭니다. 앞은 201, 뒤는 413이어야 합니다. 공백은 JSON의 의미를 바꾸지 않지만 입력 크기에는 포함됩니다. 검사를 마치면 기본 세 요청으로 복구합니다.

### 실험 D: 상태를 늦게 쓰면 무엇이 달라질까?

별도의 개인 실험에서만 `WriteHeader`를 본문 뒤로 옮겨 봅니다. 이것은 의도한 잘못된 구현입니다. 본문이 실패처럼 보여도 상태가 성공으로 남을 수 있음을 확인한 뒤 올바른 순서로 복구합니다. 기본 예제나 참고 구현의 상태 순서를 바꾸지 않습니다.

## 10. 자주 하는 실수

| 증상 | 원인 후보 | 확인 방법 |
| --- | --- | --- |
| 서버 터미널에만 내용이 보임 | `fmt.Println`으로 출력함 | 응답은 `w`에 쓰는지 확인 |
| 알 수 없는 경로가 루트 응답을 받음 | `/` 패턴이 넓게 대응 | `/{$}`의 의도와 비교 |
| POST를 보냈는데 405 | 등록 메서드와 다름 | 요청 메서드와 Allow 헤더 확인 |
| 숫자가 아닌 ID를 0으로 처리 | 변환 오류 무시 | `Atoi`의 error 확인 |
| 오류 본문인데 상태가 200 | 본문을 먼저 씀 | 상태·헤더·본문 순서 확인 |
| 오류 뒤 성공 데이터까지 붙음 | `http.Error` 뒤 return 없음 | 실패 경로의 종료 확인 |
| 큰 JSON이 메모리를 많이 사용 | 제한 없이 ReadAll | 요청 스트림에 MaxBytesReader 적용 |
| 추가 JSON 객체가 허용됨 | 디코딩 한 번으로 끝냄 | 두 번째 Decode의 EOF 확인 |
| 클라이언트가 끊어도 작업이 계속됨 | 요청 컨텍스트를 전달하지 않음 | 작업의 ctx와 취소 관찰 확인 |
| 종료 때 요청이 중간에 끊김 | 바로 Close·프로세스 종료 | Shutdown 완료를 기다리는지 확인 |

## 11. 테스트로 확인하기

```text
go test -count=1 ./examples/22-http-server/...
go vet ./examples/22-http-server/...
```

핸들러 테스트는 메서드와 경로의 여러 조합을 같은 라우터에 보내 봅니다. 단일 성공 요청만 검사하면 404·405·입력 오류가 뒤섞이는 문제를 놓칠 수 있습니다.

| 범위 | 검사 내용 |
| --- | --- |
| 첫 응답 | GET 본문, HEAD의 빈 전송 본문, 404·405 |
| 메서드 라우팅 | 루트·생성·조회와 Allow 헤더 |
| 경로 값 | 0, 음수, 문자, int 범위 초과, 추가 경로 조각 |
| JSON 입력 | 정확한 한도·초과·미디어 타입·추가 값·값 범위 |
| 오류 응답 | 상태, Content-Type, JSON 오류 코드 |
| 취소 | 이미 취소된 컨텍스트와 기한 초과 |
| 정상 종료 | 활성 요청이 완료되기 전에 Shutdown이 끝나지 않음 |

종료 테스트는 리스너의 `Close`를 관찰하는 테스트용 래퍼를 사용합니다. 수신이 닫힌 것을 확인한 뒤에도 활성 요청을 놓아주기 전에는 Shutdown이 끝나지 않는지 검사합니다. 그다음 작업을 진행시켜 완전한 응답과 서버 종료 결과를 확인합니다. 고정된 sleep으로 “아마 종료 중일 것”이라고 가정하지 않습니다.

```text
go test -race ./examples/22-http-server/...
```

race detector의 성공만으로 모든 동시성 계약이 증명되는 것은 아닙니다. 실제로 실행한 요청 경로와 종료 조건을 함께 검토합니다.

## 12. 연습문제와 자기 점검

[22장 연습문제](../../exercises/22-http-server/README.md)는 기본 여섯 문제와 선택 과제로 구성됩니다. 요구사항과 힌트를 먼저 확인하고 [해설](../../solutions/22-http-server/README.md)과 비교합니다.

| 문제 | 주제 |
| --- | --- |
| 1. 응답 작성 순서 | 암묵적 200과 최종 상태 |
| 2. 상태 확인 경로 | 메서드·본문·404·405 |
| 3. 실패 뒤의 실행 | `http.Error` 뒤 return |
| 4. 제곱 계산 경로 | 경로 값과 범위·JSON 응답 |
| 5. 계획 미리 보기 | JSON 요청 계약과 계산 |
| 6. 정상 종료 | 서버 수명과 실행 결과 |
| 선택 A. 사용자 지정 HEAD | 라우팅 우선순위와 본문 |

### 확인 질문

1. `fmt.Println`과 `fmt.Fprintln(w, ...)`의 출력 대상은 어떻게 다른가요?
2. `/`와 `/{$}` 패턴은 어떤 경로에서 달라지나요?
3. `GET` 패턴은 HEAD를 받을 수 있나요?
4. `/lessons/x`가 라우터를 통과했다고 정수 ID가 유효한가요?
5. JSON 직렬화 후 상태와 헤더를 쓰는 이유는 무엇인가요?
6. `http.Error`를 호출하면 현재 핸들러 함수도 끝나나요?
7. JSON 문법 오류와 업무 범위 오류를 구분한 계약은 무엇인가요?
8. 요청의 취소를 작업 함수에 어떻게 전달하나요?
9. `Serve`가 `ErrServerClosed`를 반환하면 무엇을 추가로 기다려야 하나요?
10. 바이트 크기 제한과 시간 제한은 서로 대체할 수 있나요?

<details>
<summary>답과 설명</summary>

1. 앞은 서버의 표준 출력, 뒤는 클라이언트에게 보낼 응답 writer입니다.
2. `/`는 하위 경로에도 대응하며 `/{$}`는 루트 경로 끝에만 대응합니다.
3. 가능합니다. 실제 HTTP 전송에서는 HEAD 응답 본문이 생략됩니다.
4. 아닙니다. 경로 조각은 문자열이며 변환과 범위 검사가 필요합니다.
5. 직렬화 실패를 아직 응답을 시작하기 전에 처리할 수 있습니다.
6. 끝나지 않습니다. 실패 뒤 성공 코드를 실행하지 않으려면 return이 필요합니다.
7. 이 예제는 문법·필드·추가 값 오류를 400, 업무 범위 오류를 422로 정했습니다.
8. `r.Context()`를 매개변수로 넘기고 작업이 그 신호를 관찰하게 합니다.
9. `Shutdown` 자체가 반환하여 진행 중 요청 정리가 끝났는지 확인합니다.
10. 아닙니다. 적은 바이트를 오래 보내거나 큰 본문을 빠르게 보낼 수 있어 둘 다 필요합니다.

</details>

### 완료 체크리스트

- [ ] 일곱 자동 예제의 응답과 종료를 확인했다.
- [ ] 404·405·400·413·415·422의 계약을 구분했다.
- [ ] 오류 뒤 성공 코드를 실행하지 않도록 작성했다.
- [ ] JSON 값 하나와 크기·업무 범위를 각각 검사했다.
- [ ] 요청 취소를 작업에 전달했다.
- [ ] Shutdown과 Close의 역할을 구분했다.
- [ ] 기본 여섯 문제의 경계 입력을 확인했다.

## 공식 참고 자료

- [http.Handler](https://pkg.go.dev/net/http#Handler) — 요청 처리 인터페이스
- [http.ServeMux](https://pkg.go.dev/net/http#ServeMux) — 메서드·경로 패턴과 우선순위
- [http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter) — 응답 작성 수명과 순서
- [http.MaxBytesReader](https://pkg.go.dev/net/http#MaxBytesReader) — 요청 본문 한도
- [http.Request.Context](https://pkg.go.dev/net/http#Request.Context) — 요청 수명
- [http.Server](https://pkg.go.dev/net/http#Server) — 시간 제한과 서버 구성
- [http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown) — 정상 종료
- [httptest](https://pkg.go.dev/net/http/httptest) — 핸들러·통합 검증

확인일: 2026-11-09. Go 1.27.3의 공식 문서와 로컬 실행을 기준으로 작성했습니다.

## 다음 학습

[23장 API 설계와 핸들러 테스트](../23-api-testing/README.md)에서는 입력·오류 계약을 구조화하고 테스트의 책임을 나눕니다. 핸들러 직접 검사와 실제 클라이언트 왕복의 차이를 이용하여 실패 경로까지 확인합니다.

| 이전 | 목차 | 실습·해설 | 다음 |
| --- | --- | --- | --- |
| [21장](../21-http-client/README.md) | [전체 목차](../README.md) | [문제](../../exercises/22-http-server/README.md) · [해설](../../solutions/22-http-server/README.md) | [23장](../23-api-testing/README.md) |
