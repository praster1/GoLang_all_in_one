# 22장 해설: HTTP 서버와 라우팅

문제: [22장 연습문제](../../exercises/22-http-server/README.md)

핸들러는 요청마다 상태·헤더·본문을 정하고 반환합니다.
실패를 응답한 경로와 성공 경로를 나눠 읽으면 중복 응답을 찾기 쉽습니다.
모든 실행 명령은 **저장소 루트** 기준입니다.
아래 프로그램은 자신의 로컬 서버를 준비하고 요청을 마친 뒤 자동 종료합니다.

## 문제 1. 응답 상태 예측

실제로 기록되는 최종 상태는 200입니다.
본문을 처음 쓰는 문장이 아직 정해지지 않은 최종 상태를 200으로 정했기 때문입니다.
그 뒤의 `WriteHeader(400)`이 앞선 상태를 되돌리지 않습니다.
[ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)의 계약에 따른 동작입니다.

개인 확인 경로: `_practice/22-http-server/predict/main.go`.
기존 파일이 있다면 내용을 보존한 뒤 사용합니다.
다음은 **잘못된 응답 순서를 관찰하는 전체 프로그램**입니다.
정상 API 구현에 이 순서를 복사하지 않습니다.

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "잘못된 입력"); err != nil {
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println("기록된 상태:", recorder.Code)
	fmt.Println("본문:", strings.TrimSpace(recorder.Body.String()))
}
```

```text
go run ./_practice/22-http-server/predict
```

예상 출력:

```text
기록된 상태: 200
본문: 잘못된 입력
```

수정하려면 본문보다 먼저 `w.WriteHeader(http.StatusBadRequest)`를 실행합니다.
`Content-Type` 등 응답 헤더도 최종 상태와 본문을 보내기 전에 설정합니다.
응답 본문의 문장이 오류를 나타내는 것과 HTTP 상태가 실패를 나타내는 것은 별개입니다.

## 문제 2. 상태 확인 경로

참고 코드: [health/main.go](health/main.go) · [동작 테스트](health/behavior_test.go)

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
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, "ready\n"); err != nil {
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
		{http.MethodGet, "/health", ""},
		{http.MethodPost, "/health", ""},
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

참고 구현:

```text
go run ./solutions/22-http-server/health
go test ./solutions/22-http-server/health
```

개인 풀이:

```text
go run ./_practice/22-http-server/health
```

예상 출력:

```text
GET /health -> 200: ready
POST /health -> 405: Method Not Allowed
```

### 등록한 규칙의 범위

`GET /health`는 GET과 HEAD에 대응합니다.
POST는 같은 경로에서 허용되지 않으므로 라우터가 405와 `Allow: GET, HEAD`를 반환합니다.
알 수 없는 `/missing`은 404입니다.
[ServeMux](https://pkg.go.dev/net/http#ServeMux)의 메서드 규칙을 활용했습니다.

| 요청 | 상태 | 전송 본문 | 비고 |
| --- | --- | --- | --- |
| GET | 200 | `ready\n` | 본문과 Content-Type 검사 |
| HEAD | 200 | 없음 | 실제 임시 서버로 검사 |
| POST | 405 | 기본 오류 텍스트 | Allow 검사 |
| 알 수 없는 경로 | 404 | 기본 오류 텍스트 | 등록 범위 확인 |

핸들러 직접 호출의 recorder와 실제 HTTP 전송을 구분해야 합니다.
이 풀이의 HEAD 검사는 임시 서버를 통과하므로 전송 계층이 본문을 생략한 결과를 확인합니다.
서버가 쓰려 한 바이트를 기록하는 검사만으로 같은 결론을 내리지 않습니다.

## 문제 3. 오류 뒤 실행 고치기

`http.Error`는 오류 응답을 작성하지만 현재 핸들러의 Go 실행을 끝내지는 않습니다.
따라서 오류 다음에 `return`을 넣습니다.
이름이 있는 정상 경로의 성공 응답은 그대로 남깁니다.

개인 파일: `_practice/22-http-server/repair/main.go`

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
		if r.URL.Query().Get("name") == "" {
			http.Error(w, "이름 필요", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, "완료\n"); err != nil {
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
		{http.MethodGet, "/hello?name=Go", ""},
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

```text
go run ./_practice/22-http-server/repair
```

예상 출력:

```text
GET /hello -> 400: 이름 필요
GET /hello?name=Go -> 200: 완료
```

| 조건 | 수정 전 | 수정 후 |
| --- | --- | --- |
| 이름 없음 | 400 본문 뒤 `완료`가 붙음 | 400 본문만 반환 |
| 이름 있음 | 200과 `완료` | 동일하게 유지 |

기본 Go 제어 흐름이 HTTP 처리에서도 그대로 적용됩니다.
에러 helper의 이름만 보고 자동으로 흐름이 끝난다고 추측하지 않습니다.
[http.Error](https://pkg.go.dev/net/http#Error)는 응답을 쓰는 함수이며 호출자의 반환을 대신하지 않습니다.

이 문제는 빈 문자열만 검사합니다.
공백 이름이나 다른 언어의 이름을 정규화하는 요구사항은 추가하지 않았습니다.
새 입력 정책이 필요하면 문제 계약과 테스트를 함께 바꿉니다.

## 문제 4. 제곱 계산 경로

참고 코드: [square/main.go](square/main.go) · [동작 테스트](square/behavior_test.go)

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
	mux.HandleFunc("GET /squares/{n}", func(w http.ResponseWriter, r *http.Request) {
		n, err := strconv.Atoi(r.PathValue("n"))
		if err != nil || n < 0 || n > 1000 {
			writeProblem(w, http.StatusBadRequest, "invalid_number")
			return
		}
		result := struct {
			Input  int `json:"input"`
			Square int `json:"square"`
		}{Input: n, Square: n * n}
		data, err := json.Marshal(result)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "encode_failed")
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
		{http.MethodGet, "/squares/12", ""},
		{http.MethodGet, "/squares/0", ""},
		{http.MethodGet, "/squares/1001", ""},
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

참고 구현:

```text
go run ./solutions/22-http-server/square
go test ./solutions/22-http-server/square
```

개인 풀이:

```text
go run ./_practice/22-http-server/square
```

예상 출력:

```text
GET /squares/12 -> 200: {"input":12,"square":144}
GET /squares/0 -> 200: {"input":0,"square":0}
GET /squares/1001 -> 400: {"error":"invalid_number"}
```

### 유효한 수만 계산하기

경로 패턴이 맞는지, 정수로 변환되는지, 허용 범위인지를 순서대로 확인합니다.
범위는 0~1000이므로 제곱 결과도 0~1000000입니다.
주어진 범위 안에서 현재 Go의 최소 int 크기로도 표현 가능한 계산을 합니다.

| 경로 값 | 결과 |
| --- | --- |
| `0` | 입력 0, 제곱 0 |
| `1` | 입력 1, 제곱 1 |
| `12` | 입력 12, 제곱 144 |
| `1000` | 입력 1000, 제곱 1000000 |
| `-1`, `1001` | 범위 오류 400 |
| `x`, 표현 불가능한 긴 숫자 | 변환 오류 400 |

변환 오류와 범위 오류는 이번 계약에서 같은 `invalid_number` 코드로 표현합니다.
하지만 어떤 경로에서 발견했는지는 구현과 테스트에서 구분할 수 있습니다.
성공과 이 경로의 입력 실패 모두 JSON Content-Type을 보냅니다.

`json.Marshal`로 데이터를 준비한 뒤 응답을 시작합니다.
응답 쓰기 중 실패한 뒤에 다른 오류 JSON을 이어 붙이지 않습니다.
새로운 실패 상태로 되돌릴 수 있는 시점과 이미 전송을 시작한 시점을 나눕니다.

## 문제 5. 학습 계획 미리 보기

참고 코드: [plan-preview/main.go](plan-preview/main.go) · [동작 테스트](plan-preview/behavior_test.go)

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

type PlanInput struct {
	Days    int `json:"days"`
	Minutes int `json:"minutes"`
}

type PlanPreview struct {
	TotalMinutes int `json:"total_minutes"`
}

func readInput(w http.ResponseWriter, r *http.Request) (PlanInput, error) {
	var input PlanInput
	r.Body = http.MaxBytesReader(w, r.Body, 128)
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
	mux.HandleFunc("POST /plans/preview", func(w http.ResponseWriter, r *http.Request) {
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
		if input.Days < 0 || input.Days > 7 || input.Minutes < 1 || input.Minutes > 120 {
			http.Error(w, "계획 값을 확인하세요", http.StatusUnprocessableEntity)
			return
		}
		data, err := json.Marshal(PlanPreview{TotalMinutes: input.Days * input.Minutes})
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
		{http.MethodPost, "/plans/preview", `{"days":3,"minutes":25}`},
		{http.MethodPost, "/plans/preview", `{"days":0,"minutes":25}`},
		{http.MethodPost, "/plans/preview", `{"days":8,"minutes":25}`},
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

참고 구현:

```text
go run ./solutions/22-http-server/plan-preview
go test ./solutions/22-http-server/plan-preview
```

개인 풀이:

```text
go run ./_practice/22-http-server/plan-preview
```

예상 출력:

```text
POST /plans/preview -> 200: {"total_minutes":75}
POST /plans/preview -> 200: {"total_minutes":0}
POST /plans/preview -> 422: 계획 값을 확인하세요
```

### 생성과 미리 보기의 차이

이 API는 데이터를 저장하지 않고 결과를 계산합니다.
따라서 성공 코드를 200으로 정하고 새 리소스 주소를 나타내는 Location을 보내지 않습니다.
본문의 생성 응답 형식을 그대로 복사하는 대신 이 경로의 책임에 맞게 결정했습니다.

`days`가 0이면 0분이 유효한 계산 결과입니다.
그래도 `minutes`는 1~120이어야 합니다.
0일이라는 이유로 나머지 입력의 검증을 건너뛰지 않습니다.

| 날짜·하루 분량 | 상태 | 총 분량 |
| --- | --- | --- |
| 3·25 | 200 | 75 |
| 0·25 | 200 | 0 |
| 1·1 | 200 | 1 |
| 7·120 | 200 | 840 |
| -1·25, 8·25 | 422 | 반환하지 않음 |
| 0·0, 5·121 | 422 | 반환하지 않음 |

이 계약에서는 `days`가 없으면 정수의 영값인 0일로 해석합니다.
`minutes`가 없으면 0으로 남아 범위 검사에서 실패합니다.
필드의 존재 자체를 필수로 구분하려면 포인터 필드 등의 별도 표현과 검증이 필요합니다.
이번 문제는 그 요구사항을 포함하지 않습니다.

### 형식 검증과 값 검증

| 실패 입력 | 상태 | 발견 단계 |
| --- | --- | --- |
| 다른 Content-Type | 415 | 미디어 타입 |
| 129바이트 | 413 | 제한된 본문 읽기 |
| 빈 본문·잘못된 타입·알 수 없는 필드 | 400 | 첫 디코딩 |
| 두 객체·뒤의 쓰레기 문자 | 400 | 두 번째 디코딩 |
| `null`, `{}`, 범위 밖 값 | 422 | 업무 조건 |

128바이트에 맞게 공백을 붙인 정상 객체는 성공합니다.
129바이트에서는 같은 의미의 JSON이어도 크기 계약을 넘으므로 실패합니다.
`MaxBytesReader`의 오류를 `errors.As`로 구분하는 이유입니다.

`DisallowUnknownFields`는 중복 키를 거절하는 별도 기능이 아니며 필드 대소문자 매칭도 기본 `encoding/json` 규칙을 따릅니다.
이 풀이의 “알려진 필드” 계약을 그보다 더 엄격한 JSON 검증이라고 부르지 않습니다.
[encoding/json](https://pkg.go.dev/encoding/json)의 기본 규칙을 참고합니다.

## 문제 6. 정상 종료하는 상태 서버

참고 코드: [shutdown-health/main.go](shutdown-health/main.go) · [종료 테스트](shutdown-health/behavior_test.go)

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
		if _, err := io.WriteString(w, "healthy\n"); err != nil {
			return
		}
	})
	return mux
}

func stopServer(ctx context.Context, server *http.Server, served <-chan error) error {
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("종료 대기: %w", err)
	}
	if err := <-served; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("서버 실행: %w", err)
	}
	return nil
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
	if err := stopServer(ctx, server, serveResult); err != nil {
		return err
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

참고 구현:

```text
go run ./solutions/22-http-server/shutdown-health
go test ./solutions/22-http-server/shutdown-health
```

개인 풀이:

```text
go run ./_practice/22-http-server/shutdown-health
```

예상 출력:

```text
응답: healthy
정상 종료: true
```

### 기다리는 대상의 순서

`stopServer`는 Shutdown의 완료를 기다리고 그 뒤 Serve의 결과를 확인합니다.
Serve의 반환을 먼저 서버 전체 정리 완료로 해석하지 않습니다.
진행 중 요청의 완료를 기다리는 책임은 Shutdown에 있습니다.
[Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)의 문서를 참고합니다.

| 신호 | 뜻 | 이것만으로 부족한 이유 |
| --- | --- | --- |
| 리스너 닫힘 | 새 요청 수신 중지 | 활성 요청이 남을 수 있음 |
| Serve 반환 | 수신 루프 종료 | 요청 정리 완료와 다를 수 있음 |
| Shutdown 반환 | 종료 대기 성공 또는 기한 오류 | 반환 오류를 확인해야 함 |
| 요청 결과 수신 | 특정 요청의 응답 완료 | 다른 요청도 있으면 별도 대기 필요 |

테스트는 실제 요청 하나를 핸들러 안에서 기다리게 합니다.
종료로 리스너가 닫힌 뒤에도 작업을 풀기 전에는 `stopServer`가 완료되지 않아야 합니다.
작업을 풀어 준 뒤 전체 본문과 정상 종료를 확인합니다.
고정된 sleep을 두지 않아 기계 속도에 맞춰 대기 시간을 추측하지 않습니다.

중간 오류 경로에서도 `defer server.Close()`가 남은 연결을 정리합니다.
정상 종료의 기다림과 마지막 강제 정리를 나누어 둔 것입니다.
서비스 전체의 신호 처리·로그·의존성 종료 순서는 운영 기초 장에서 확장합니다.

## 선택 A. HEAD 전용 처리

개인 파일: `_practice/22-http-server/custom-head/main.go`

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

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Source", "get")
		if _, err := io.WriteString(w, "data"); err != nil {
			return
		}
	})
	mux.HandleFunc("HEAD /summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Source", "head")
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		req, err := http.NewRequest(method, server.URL+"/summary", nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 17))
		closeErr := resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		if len(body) > 16 {
			return fmt.Errorf("응답 한도 초과")
		}
		fmt.Printf("%s: %d, 본문=%d, 출처=%s\n", method, resp.StatusCode, len(body), resp.Header.Get("X-Source"))
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

```text
go run ./_practice/22-http-server/custom-head
```

예상 출력:

```text
GET: 200, 본문=4, 출처=get
HEAD: 200, 본문=0, 출처=head
```

GET 패턴도 HEAD에 대응하지만 HEAD 전용 패턴은 그보다 구체적입니다.
따라서 HEAD 요청은 `X-Source: head`를 설정하는 핸들러로 갑니다.
단순히 본문이 비었다는 사실만으로는 어느 핸들러가 실행되었는지 알 수 없어 헤더로 구분했습니다.
[ServeMux](https://pkg.go.dev/net/http#ServeMux)의 메서드·우선순위 규칙을 확인합니다.

클라이언트의 응답 본문은 반복마다 읽고 즉시 닫습니다.
짧은 실험이더라도 반복 전체가 끝날 때까지 여러 본문을 열어 둘 필요는 없습니다.
읽기 오류와 닫기 결과를 확인하고 다음 요청으로 넘어갑니다.

## 검증과 다시 설명할 질문

```text
go test -count=1 ./solutions/22-http-server/...
go vet ./solutions/22-http-server/...
```

참고 구현은 기본 출력뿐 아니라 여러 요청·입력 경계·종료 순서를 확인합니다.
개인 코드는 `_practice` 경로를 직접 지정하여 실행합니다.
참고 구현 테스트의 성공을 자신의 구현 검사로 대신하지 않습니다.

- 응답 상태를 먼저 정하지 않으면 어떤 기본값이 적용되는가?
- 같은 경로에서 404와 405가 나뉘는 기준은 무엇인가?
- 0일 계획을 허용하면서 0분 하루 계획은 거절하는 이유는 무엇인가?
- 본문 바이트 한도와 JSON 의미 검사는 어떻게 다른가?
- Shutdown을 기다리지 않고 Serve 반환만 기다리면 무엇을 놓칠 수 있는가?

[22장 본문](../../book/22-http-server/README.md) · [문제](../../exercises/22-http-server/README.md) · [해설 목록](../README.md)
