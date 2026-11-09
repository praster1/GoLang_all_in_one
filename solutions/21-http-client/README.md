# 21장 해설: HTTP 클라이언트

문제: [21장 연습문제](../../exercises/21-http-client/README.md)

요청이 어디까지 진행되었는지와 누가 자원을 닫는지를 함께 확인합니다.
모든 명령은 **저장소 루트**에서 실행합니다.
참고 구현의 테스트는 개인 폴더의 코드를 대신 검사하지 않습니다.
각 프로그램의 임시 서버는 실행 중에만 존재하고 스스로 종료됩니다.

## 문제 1. 응답과 오류 예측

출력:

```text
오류 없음: true
상태: 404
```

`Get`은 HTTP 응답을 받는 데 성공했습니다.
404는 그 응답이 나타내는 결과이며, 전송 오류로 자동 변환되지 않습니다.
이 API가 200만 허용한다면 본문 닫기를 등록한 뒤 상태를 검사하고 업무 오류를 반환합니다.
상태 때문에 반환하는 경우에도 본문을 닫을 책임은 유지됩니다.
[Client.Do](https://pkg.go.dev/net/http#Client.Do)의 오류·본문 계약에 따른 동작입니다.

개인 확인 경로: `_practice/21-http-client/predict/main.go`.
기존 파일이 있으면 먼저 보존하고 다음 전체 프로그램을 저장합니다.

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	resp, err := client.Get(server.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println("오류 없음:", err == nil)
	fmt.Println("상태:", resp.StatusCode)
	_, err = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```text
go run ./_practice/21-http-client/predict
```

`io.Copy`는 이 실험에서 필요 없는 작은 본문을 버립니다.
읽기에도 한도를 두었으며 반환 오류를 확인합니다.
상태 코드만 확인한다고 정리 책임까지 없어지지는 않습니다.

## 문제 2. 이름과 과정 헤더 전달

참고 코드: [header-greeting/main.go](header-greeting/main.go) · [동작 테스트](header-greeting/behavior_test.go)

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

func greet(ctx context.Context, client *http.Client, address, name string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("name", name)
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Course", "Go")
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
		if _, err := fmt.Fprintf(w, "이름: %s\n과정: %s\n", r.URL.Query().Get("name"), r.Header.Get("X-Course")); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	text, err := greet(context.Background(), client, server.URL, "민지 & Go")
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

참고 구현:

```text
go run ./solutions/21-http-client/header-greeting
go test ./solutions/21-http-client/header-greeting
```

개인 풀이:

```text
go run ./_practice/21-http-client/header-greeting
```

예상 출력:

```text
이름: 민지 & Go
과정: Go
```

### 쿼리와 헤더의 역할

`name`은 URL 쿼리에, 과정 이름은 `X-Course` 헤더에 넣었습니다.
사용한 헤더 이름은 이 실습의 약속이며 HTTP 표준이 과정 이름을 자동 해석하는 것은 아닙니다.
서버도 같은 이름으로 읽도록 맞추었기 때문에 왕복 결과를 관찰할 수 있습니다.

| 입력 이름 | 서버가 읽는 이름 | 설명 |
| --- | --- | --- |
| 빈 문자열 | 빈 문자열 | 이 문제는 빈 이름을 허용 |
| `Go & HTTP` | `Go & HTTP` | `&`가 쿼리 구분자로 오해되지 않음 |
| `한글 + ? #` | `한글 + ? #` | URL 기호와 원래 값 구분 |

`u.Query()`로 기존 쿼리를 얻고 `Set`을 호출한 뒤 다시 인코딩했습니다.
주소의 `keep=1`을 보존하는 테스트도 있습니다.
직접 문자열을 합치는 방식은 이름에 포함된 기호에 따라 동작이 달라질 수 있습니다.
[net/url.Values.Encode](https://pkg.go.dev/net/url#Values.Encode)를 사용한 이유입니다.

서버의 응답 쓰기가 실패하면 핸들러는 돌아옵니다.
클라이언트에서는 상태와 읽기 오류를 따로 확인합니다.
정상 경로뿐 아니라 오류 경로에서도 응답 본문이 닫히도록 `defer`를 등록했습니다.

## 문제 3. 닫기 위치 고치기

문제의 원래 코드는 전송 오류를 확인하기 전에 `resp.Body`를 사용합니다.
`defer`의 실제 호출은 늦춰지지만, 어떤 대상의 메서드를 호출할지는 등록할 때 평가됩니다.
따라서 응답이 없는 경로에서 패닉이 발생할 수 있습니다.
[defer 문](https://go.dev/ref/spec#Defer_statements)의 평가 시점을 참고합니다.

수정 순서는 요청 생성 → 전송 오류 확인 → 닫기 등록 → 상태 판정 → 본문 읽기입니다.
다음 전체 코드는 정상 로컬 응답과 잘못된 URL을 한 번씩 확인합니다.
개인 경로는 `_practice/21-http-client/repair/main.go`입니다.

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
		if _, err := io.WriteString(w, "ok"); err != nil {
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
	fmt.Println("정상:", text)
	_, err = fetchText(context.Background(), client, "://bad")
	if err == nil {
		return fmt.Errorf("잘못된 주소가 허용됨")
	}
	fmt.Println("실패 확인:", err != nil)
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
go run ./_practice/21-http-client/repair
```

예상 출력:

```text
정상: ok
실패 확인: true
```

| 경로 | 응답 본문 소유권 | 함수의 결과 |
| --- | --- | --- |
| URL 생성 실패 | 본문을 얻지 못함 | 오류 반환 |
| 전송 실패 | 정상 응답을 얻지 못함 | 오류 반환 |
| 상태 실패 | 본문을 얻음 | 닫고 상태 오류 반환 |
| 읽기 실패 | 본문을 얻음 | 닫고 읽기 오류 반환 |
| 성공 | 본문을 얻음 | 닫고 텍스트 반환 |

실패 주소는 외부 서버를 찾아가도록 만든 값이 아닙니다.
URL 형식이 잘못되어 요청 생성 단계에서 실패합니다.
정확한 오류 메시지를 문자열로 고정하지 않고 오류가 있는지를 확인합니다.

## 문제 4. 한도가 있는 다운로드

참고 코드: [limited-download/main.go](limited-download/main.go) · [동작 테스트](limited-download/behavior_test.go)

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

var ErrTooLarge = errors.New("본문 한도 초과")

func readAtMost(reader io.Reader, limit int64) ([]byte, error) {
	if limit < 0 || limit > 1024 {
		return nil, errors.New("한도는 0~1024바이트")
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

func download(ctx context.Context, client *http.Client, address string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
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
		if _, err := io.WriteString(w, "hello"); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	_, err := download(context.Background(), client, server.URL, 4)
	if !errors.Is(err, ErrTooLarge) {
		return fmt.Errorf("예상과 다른 결과: %v", err)
	}
	fmt.Println("한도 초과: true")
	body, err := download(context.Background(), client, server.URL, 5)
	if err != nil {
		return err
	}
	fmt.Println("본문:", string(body))
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
go run ./solutions/21-http-client/limited-download
go test ./solutions/21-http-client/limited-download
```

개인 풀이:

```text
go run ./_practice/21-http-client/limited-download
```

예상 출력:

```text
한도 초과: true
본문: hello
```

### 한도와 결과를 연결하기

`limit+1`바이트까지 읽은 결과가 한도보다 길면 `ErrTooLarge`를 반환합니다.
`hello`는 5바이트이므로 한도 4에서는 초과하고 한도 5에서는 정확히 성공합니다.
한도가 0일 때 빈 본문은 허용되지만 `x`는 초과합니다.

바이트 한도는 UTF-8 문자 수와 다릅니다.
`가`의 3바이트를 한도 2에서 잘라 성공으로 돌려주지 않습니다.
함수는 완전한 본문을 허용 범위 안에서 얻었을 때만 성공합니다.

| 실패 종류 | 반환 오류 | 호출자가 확인할 방식 |
| --- | --- | --- |
| 한도 범위가 잘못됨 | 입력 오류 | 오류 존재와 입력 값 확인 |
| 본문이 한도를 넘음 | `ErrTooLarge` | `errors.Is` |
| 읽기 도중 실패 | 읽기 원인 | 원인을 보존하여 처리 |
| 상태가 200이 아님 | 상태 오류 | API 계약과 상태 확인 |
| 요청 취소 | 전송 오류 안의 취소 원인 | `errors.Is(err, context.Canceled)` |

이 예제는 크기가 작은 데이터를 메모리에 읽는 용도입니다.
큰 파일을 내려받아야 한다면 스트리밍 복사와 저장 실패 처리까지 설계합니다.
한도를 키우는 것만으로 모든 다운로드 요구를 해결하지 않습니다.

## 문제 5. 학습 JSON 조회

참고 코드: [lesson-json/main.go](lesson-json/main.go) · [동작 테스트](lesson-json/behavior_test.go)

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
		if err := json.NewEncoder(w).Encode(Lesson{Title: "함수", Minutes: 45}); err != nil {
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

참고 구현:

```text
go run ./solutions/21-http-client/lesson-json
go test ./solutions/21-http-client/lesson-json
```

개인 풀이:

```text
go run ./_practice/21-http-client/lesson-json
```

예상 출력:

```text
학습: 함수, 45분
```

### 작은 검증 단계를 쌓기

`fetchLesson`은 HTTP 상태와 응답 형식을 확인하고 본문을 닫습니다.
`decodeLesson`은 네트워크를 몰라도 `io.Reader`에서 데이터를 읽어 검증합니다.
따라서 JSON의 경계 입력은 문자열 reader만으로 빠르게 검사할 수 있습니다.
HTTP 헤더의 계약은 임시 서버로 별도 검사합니다.

| 입력 | 결과 | 이유 |
| --- | --- | --- |
| 제목과 분량 1 | 성공 | 최솟값 포함 |
| 제목과 분량 180 | 성공 | 최댓값 포함 |
| 분량 0 또는 181 | 실패 | 업무 범위 밖 |
| 공백 제목 | 실패 | 의미 있는 제목 없음 |
| 객체 뒤 공백 | 성공 | 다음 값 없이 EOF |
| 객체 두 개 | 실패 | 두 번째 디코딩이 성공함 |
| 객체 뒤 `x` | 실패 | 뒤에 잘못된 문자가 남음 |
| `null` 또는 `{}` | 실패 | 필수 값의 조건을 만족하지 못함 |
| 알 수 없는 `extra` 필드 | 실패 | `DisallowUnknownFields` |
| 257바이트 | 실패 | 전체 크기 초과 |

제목을 `TrimSpace`로 검사하지만 반환한 제목 자체를 자동으로 바꾸지는 않습니다.
공백만인 제목을 금지하는 계약과 모든 제목을 정규화하는 계약은 다릅니다.
이 문제는 앞뒤 공백이 있어도 다른 글자가 있으면 원래 문자열을 유지합니다.

이 검사는 JSON의 중복 키를 금지하는 기능은 아닙니다.
또한 구조체 필드 매칭의 대소문자 규칙은 기본 JSON 패키지의 규칙을 따릅니다.
그 이상의 엄격한 형식이 필요하면 입력 계약을 먼저 확장해야 합니다.
[encoding/json](https://pkg.go.dev/encoding/json)의 호환성·보안 설명을 참고합니다.

## 문제 6. 취소된 상태 조회

참고 코드: [cancel-check/main.go](cancel-check/main.go) · [동작 테스트](cancel-check/behavior_test.go)

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
	"sync/atomic"
	"time"
)

func requestStatus(ctx context.Context, client *http.Client, address string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, address, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("상태 조회: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func run() error {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := requestStatus(ctx, client, server.URL)
	if !errors.Is(err, context.Canceled) {
		return fmt.Errorf("취소 결과: %v", err)
	}
	fmt.Println("취소 확인: true")
	fmt.Println("서버 요청 수:", calls.Load())
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
go run ./solutions/21-http-client/cancel-check
go test ./solutions/21-http-client/cancel-check
```

개인 풀이:

```text
go run ./_practice/21-http-client/cancel-check
```

예상 출력:

```text
취소 확인: true
서버 요청 수: 0
```

### 컨텍스트를 바꾸지 않고 전달하기

`NewRequestWithContext`에 호출자가 준 `ctx`를 사용했습니다.
함수 안에서 새로운 `Background()`로 바꾸면 이미 취소한 사실이 요청에 전달되지 않습니다.
전송 오류를 `%w`로 감싸도 `errors.Is`로 취소 원인을 찾을 수 있습니다.

`requestStatus`는 “상태 조사”라는 책임을 가지므로 HTTP 상태 자체를 오류로 바꾸지 않습니다.
본문에서 작성한 “200의 학습 JSON 조회” 함수와 다른 계약입니다.
함수 이름과 반환값이 무엇을 약속하는지에 따라 상태 처리 방식이 달라질 수 있습니다.

요청 수는 서버 핸들러에서 증가시키고 클라이언트 쪽에서 읽습니다.
`atomic.Int32`를 사용하여 공유 값의 동시 접근을 처리합니다.
본문의 `Example_main`은 시작 전 취소에서 0회를, 함수 테스트는 정상 204와 진행 중 취소도 확인합니다.

진행 중 취소 테스트는 다음 순서를 사용합니다.

1. 서버가 요청을 받으면 시작 채널을 닫습니다.
2. 테스트가 시작 신호를 받은 뒤 취소합니다.
3. 클라이언트 오류가 취소를 나타내는지 검사합니다.
4. 서버 핸들러도 컨텍스트 종료를 보고 끝났는지 확인합니다.

고정된 sleep 없이 요청의 실제 상태를 확인하므로 스케줄링 속도에 덜 의존합니다.
테스트 전체에는 실행 시간 제한을 두어 예상 밖 정지를 발견할 수 있습니다.

## 선택 A. 리다이렉트 한 번

개인 경로: `_practice/21-http-client/redirect-once/main.go`

전체 코드:

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"time"
)

func run() error {
	var finalCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/b", http.StatusFound) })
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/c", http.StatusFound) })
	mux.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) { finalCalls.Add(1) })
	server := httptest.NewServer(mux)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 2 {
			return http.ErrUseLastResponse
		}
		return nil
	}
	resp, err := client.Get(server.URL + "/a")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, 1024)); err != nil {
		return err
	}
	fmt.Println("중간 상태:", resp.StatusCode)
	fmt.Println("남은 위치:", resp.Header.Get("Location"))
	fmt.Println("최종 요청 수:", finalCalls.Load())
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
go run ./_practice/21-http-client/redirect-once
```

예상 출력:

```text
중간 상태: 302
남은 위치: /c
최종 요청 수: 0
```

| 이동 판정 | `via`에 있는 경로 | 길이 | 정책 |
| --- | --- | --- | --- |
| `/a`에서 `/b`로 | `/a` | 1 | 허용 |
| `/b`에서 `/c`로 | `/a`, `/b` | 2 | 마지막 응답 반환 |

최초 요청은 이동 횟수가 아닙니다.
이전 요청의 개수로 두 번째 이동을 구분하여 `ErrUseLastResponse`를 반환했습니다.
따라서 `/b`의 302를 받고 `/c`는 호출하지 않습니다.
받은 중간 응답의 본문 역시 닫습니다.
[Client.CheckRedirect](https://pkg.go.dev/net/http#Client)와 [ErrUseLastResponse](https://pkg.go.dev/net/http#ErrUseLastResponse)의 계약에 근거합니다.

## 검증 범위와 다시 설명할 질문

```text
go test -count=1 ./solutions/21-http-client/...
go vet ./solutions/21-http-client/...
```

각 참고 구현에는 기본 출력을 확인하는 `Example_main`과 경계 입력을 확인하는 동작 테스트가 있습니다.
테스트는 외부 사이트나 실계정에 의존하지 않습니다.
개인 풀이를 확인할 때는 `_practice` 경로를 명시해서 실행합니다.

- 상태가 실패여도 본문을 닫아야 하는 이유는 무엇인가?
- 읽기 한도를 지키면서 “초과”를 알아내는 방법은 무엇인가?
- HTTP 검사와 JSON 검사를 나누면 어떤 입력을 독립적으로 테스트할 수 있는가?
- 취소된 컨텍스트를 새 컨텍스트로 바꾸면 어떤 계약이 깨지는가?
- 리다이렉트를 중지한 응답의 소유자는 누구인가?

[21장 본문](../../book/21-http-client/README.md) · [문제](../../exercises/21-http-client/README.md) · [해설 목록](../README.md)
