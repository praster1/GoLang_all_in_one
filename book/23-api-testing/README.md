# 23. API 설계와 핸들러 테스트

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 중급 | [22장 HTTP 서버](../22-http-server/README.md), 테스트·인터페이스·동기화 | Go, 편집기, 터미널 |

서버가 요청 하나에 200을 보냈다고 API가 올바르게 동작한다고 결론 내릴 수는 없습니다. 본문이 다른 데이터이거나, 필수 헤더가 없거나, 잘못된 입력도 성공으로 받아들이고 있을 수 있습니다. 이번 장에서는 먼저 관찰 가능한 계약을 적고 그 계약을 여러 층의 테스트로 확인합니다.

테스트는 코드와 똑같은 조건문을 다시 쓰는 일이 아닙니다. 호출자가 기대하는 동작을 독립된 입력과 결과로 표현하는 일입니다. HTTP 처리, 업무 검증, 저장소 실패, 미들웨어와 동시 요청을 각각 확인할 수 있게 책임을 나눕니다.

## 학습 목표

- 입력·상태·헤더·본문·부작용을 포함한 API 계약을 작성한다.
- 순수 함수, 핸들러 직접 호출과 실제 HTTP 왕복의 검사 범위를 구분한다.
- 필수 필드 누락, 형식 오류와 업무 범위 오류를 다르게 검사한다.
- 저장소 인터페이스를 주입하여 실패와 컨텍스트 전달을 재현한다.
- 미들웨어와 동시 요청의 공통 규칙을 테스트한다.

## 준비와 예제

모든 명령은 **저장소 루트**에서 실행합니다. 새 모듈을 만들지 않으며 외부 HTTP 서비스나 데이터베이스는 필요하지 않습니다. 핸들러 직접 실행 예제는 메모리의 recorder를 사용하고, 통합 예제는 로컬 TLS 서버를 시작한 뒤 자동 종료합니다.

| 예제 | 확인할 경계 | 실행 명령 |
| --- | --- | --- |
| [validation](../../examples/23-api-testing/validation/main.go) | 순수 검증과 정규화 | `go run ./examples/23-api-testing/validation` |
| [handler-table](../../examples/23-api-testing/handler-table/main.go) | 상태·헤더·JSON 계약 | `go run ./examples/23-api-testing/handler-table` |
| [json-contract](../../examples/23-api-testing/json-contract/main.go) | 필수 필드·크기·값 하나 | `go run ./examples/23-api-testing/json-contract` |
| [request-id](../../examples/23-api-testing/request-id/main.go) | 미들웨어와 컨텍스트 | `go run ./examples/23-api-testing/request-id` |
| [injected-store](../../examples/23-api-testing/injected-store/main.go) | 저장소 결과와 오류 분류 | `go run ./examples/23-api-testing/injected-store` |
| [integration](../../examples/23-api-testing/integration/main.go) | TLS·전송·HEAD | `go run ./examples/23-api-testing/integration` |
| [concurrent-counter](../../examples/23-api-testing/concurrent-counter/main.go) | 동시 변경과 읽기 | `go run ./examples/23-api-testing/concurrent-counter` |

각 폴더의 `main_test.go`는 기본 출력, `behavior_test.go`는 동작의 경계를 확인합니다. 이번 장에서는 테스트 파일도 학습할 코드입니다. 표현식의 문법보다 “이 검사가 어떤 결함을 발견하는가”를 먼저 설명해 보세요.

## 1. API 계약에는 상태보다 많은 것이 있습니다

API 계약은 호출자와 구현이 함께 지켜야 하는 약속입니다. 같은 URL이라도 메서드, 입력 형식, 성공 결과와 실패 결과가 달라질 수 있습니다. 작은 API부터 표로 적으면 구현과 테스트가 같은 기준을 사용할 수 있습니다.

| 항목 | 예시 질문 |
| --- | --- |
| 경로·메서드 | GET `/greet`인가? POST도 받는가? |
| 입력 | 이름을 어디에서 읽고 어떤 길이를 허용하는가? |
| 정규화 | 앞뒤 공백을 제거하는가? 원문을 유지하는가? |
| 성공 상태 | 200인가, 201인가, 204인가? |
| 성공 본문 | 어떤 필드와 타입을 반환하는가? |
| 오류 | 어떤 상태와 안정적인 오류 코드를 반환하는가? |
| 부작용 | 입력이 잘못되어도 저장소를 호출하는가? |
| 취소 | 상위 요청의 컨텍스트를 전달하는가? |

예를 들어 상태만 200인 테스트는 `{"message":"wrong"}`도 통과할 수 있습니다. 헤더를 검사하지 않으면 JSON이지만 `text/plain`으로 선언한 응답을 놓칠 수 있습니다. 입력 실패에서 저장소 호출 횟수를 확인하지 않으면 불필요한 부작용을 놓칠 수 있습니다.

### 어떤 내용을 고정된 문자열로 비교할까요?

상태 코드, 오류 코드와 숫자 결과는 계약으로 고정하기 좋습니다. JSON 필드 순서나 보기용 공백은 계약에 꼭 필요한 경우가 아니면 구조체로 해석하여 비교합니다. 반대로 최종 사용자 출력 자체를 가르치는 `Example_main`은 문서의 출력 형식을 확인합니다. 두 검사는 목적이 다릅니다.

| 검사 방식 | 적합한 목적 |
| --- | --- |
| JSON 디코딩 후 필드 비교 | API의 데이터 의미 |
| 헤더별 값 비교 | Content-Type, 요청 ID, Allow 등 |
| 정확한 stdout 비교 | 교재 실행 예제의 관찰 결과 |
| 오류 타입·sentinel 비교 | 원인 분류와 오류 감싸기 |
| 호출 횟수 비교 | 검증 실패의 부작용 방지 |

### 읽기와 변경의 의미도 계약입니다

조회 요청이 업무 데이터를 바꾸면 재조회나 자동 재요청의 영향을 이해하기 어려워집니다. 이번 장의 카운터는 GET으로 현재 값을 읽고 POST로 증가시킵니다. HTTP 메서드의 의미는 [RFC 9110의 안전한 메서드](https://www.rfc-editor.org/rfc/rfc9110.html#section-9.2.1)를 참고합니다. 로그 같은 내부 관찰과 API의 요청된 업무 상태 변경은 구분합니다.

## 2. HTTP 없이 검증 규칙부터 확인하기

입력 문자열을 정리하고 범위를 판정하는 데 실제 HTTP 서버는 필요하지 않습니다. 순수 함수로 나누면 작은 입력 표를 빠르게 검사할 수 있습니다. HTTP 상태로 변환하는 일은 이후 핸들러의 책임으로 둡니다.

저장 경로: [validation/main.go](../../examples/23-api-testing/validation/main.go)

전체 코드:

```go
package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type CreateInput struct {
	Title   string
	Minutes int
}

type FieldError struct{ Field string }

func (e *FieldError) Error() string { return "invalid field: " + e.Field }

func validate(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	length := utf8.RuneCountInString(input.Title)
	if length < 1 || length > 40 {
		return CreateInput{}, &FieldError{Field: "title"}
	}
	if input.Minutes < 1 || input.Minutes > 180 {
		return CreateInput{}, &FieldError{Field: "minutes"}
	}
	return input, nil
}

func main() {
	input, err := validate(CreateInput{Title: " Go ", Minutes: 25})
	if err != nil {
		fmt.Println("예상하지 못한 오류:", err)
		return
	}
	fmt.Printf("정규화: %s, %d분\n", input.Title, input.Minutes)
	_, err = validate(CreateInput{Title: "Go", Minutes: 0})
	var field *FieldError
	if errors.As(err, &field) {
		fmt.Println("필드 오류:", field.Field)
	}
}
```

실행:

```text
go run ./examples/23-api-testing/validation
```

예상 출력:

```text
정규화: Go, 25분
필드 오류: minutes
```

### 검증과 정규화의 순서

제목은 앞뒤 공백을 제거한 뒤 길이를 검사합니다. 따라서 공백 세 개는 “길이 3인 제목”으로 허용되지 않습니다. 유효한 경우 정규화한 새 입력 값을 반환합니다. 원래 구조체는 값으로 전달되므로 호출자의 문자열 필드를 직접 바꾸지 않습니다.

| 입력 | 결과 |
| --- | --- |
| 제목 ` Go `, 25분 | 제목 `Go`, 25분 |
| 공백 제목 | `FieldError{Field: "title"}` |
| 제목 40룬 | 허용 |
| 제목 41룬 | 제목 오류 |
| 0분·181분 | 분량 오류 |
| 1분·180분 | 허용 |

길이는 `utf8.RuneCountInString`으로 셉니다. 룬 수와 UTF-8 바이트 수, 화면에서 보이는 글자 묶음 수는 같지 않을 수 있습니다. 오류 메시지의 “글자”는 이 교재의 계약에서 유니코드 룬 수를 뜻합니다. [utf8.RuneCountInString](https://pkg.go.dev/unicode/utf8#RuneCountInString)의 의미를 확인합니다.

### 오류 타입을 만드는 목적

오류 문자열의 일부를 검색하는 대신 `errors.As`로 필드를 받습니다. 호출자는 제목 오류와 분량 오류를 분류할 수 있습니다. 실제 API의 사용자 메시지는 이 오류 정보를 보고 별도로 정할 수 있습니다. [errors.As](https://pkg.go.dev/errors#As)를 참고합니다.

이 예제는 여러 오류를 한 번에 모두 모으지 않고 첫 오류를 반환합니다. 제목과 분량이 모두 틀렸다면 제목 오류를 먼저 받습니다. 여러 필드 오류를 함께 반환하는 API가 필요하면 결과 모델과 순서를 추가로 정해야 합니다.

## 3. 상태·헤더·JSON을 함께 검사하기

`httptest.NewRequest`는 서버 핸들러 검사에 사용할 요청을 만듭니다. `httptest.NewRecorder`는 핸들러가 쓴 응답을 메모리에 기록합니다. 둘을 사용하면 리스너를 열지 않고 핸들러를 직접 호출할 수 있습니다. [httptest](https://pkg.go.dev/net/http/httptest)의 도구를 참고합니다.

저장 경로: [handler-table/main.go](../../examples/23-api-testing/handler-table/main.go)

전체 코드:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"unicode/utf8"
)

type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /greet", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		length := utf8.RuneCountInString(name)
		if length < 1 || length > 20 {
			writeProblem(w, http.StatusBadRequest, "invalid_name", "이름은 1~20글자입니다")
			return
		}
		result := struct {
			Message string `json:"message"`
		}{Message: "안녕하세요, " + name + "!"}
		writeJSON(w, http.StatusOK, result)
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/greet?name=Go", ""},
		{http.MethodGet, "/greet", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		fmt.Printf("%d %s\n", recorder.Code, strings.TrimSpace(recorder.Body.String()))
	}
}
```

실행:

```text
go run ./examples/23-api-testing/handler-table
```

예상 출력:

```text
200 {"message":"안녕하세요, Go!"}
400 {"code":"invalid_name","message":"이름은 1~20글자입니다"}
```

### 실제 테이블 테스트 읽기

다음은 같은 폴더의 [behavior_test.go](../../examples/23-api-testing/handler-table/behavior_test.go) 전체입니다. `go run` 대상 파일에 붙이지 말고 테스트 파일로 읽습니다.

```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGreetingContract(t *testing.T) {
	tests := []struct {
		name, input   string
		status        int
		message, code string
	}{
		{"normal", "Go", 200, "안녕하세요, Go!", ""},
		{"trimmed", " Go ", 200, "안녕하세요, Go!", ""},
		{"maximum", strings.Repeat("가", 20), 200, "안녕하세요, " + strings.Repeat("가", 20) + "!", ""},
		{"missing", "", 400, "이름은 1~20글자입니다", "invalid_name"},
		{"blank", "  ", 400, "이름은 1~20글자입니다", "invalid_name"},
		{"too long", strings.Repeat("가", 21), 400, "이름은 1~20글자입니다", "invalid_name"},
	}
	handler := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/greet?name="+url.QueryEscape(tc.input), nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status=%d, want %d", recorder.Code, tc.status)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type=%q", got)
			}
			var body struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid response JSON: %v", err)
			}
			if body.Message != tc.message || body.Code != tc.code {
				t.Errorf("body=%+v, want message=%q code=%q", body, tc.message, tc.code)
			}
		})
	}
}
```

실행:

```text
go test -v ./examples/23-api-testing/handler-table
```

성공하면 각 이름의 하위 테스트가 PASS로 표시됩니다. 테스트 실행 시간과 패키지 경로 등 부가 표시는 환경에 따라 달라질 수 있습니다. 프로그램의 API 출력과 테스트 도구의 보고 형식을 구분합니다.

### 각 assertion의 목적

| 검사 | 발견할 수 있는 잘못된 구현 |
| --- | --- |
| `recorder.Code` | 실패 입력을 200으로 반환 |
| Content-Type | JSON 본문을 다른 형식으로 선언 |
| JSON 해석 | 깨진 JSON 또는 두 응답을 이어 씀 |
| `Message` | 이름을 버리거나 잘못 정규화 |
| `Code` | 실패 종류를 잘못 분류 |

`t.Fatalf`는 현재 하위 테스트를 끝냅니다. 뒤의 검사가 의미 없을 때 먼저 멈추게 했습니다. 최종 필드 값 비교는 `t.Errorf`로 차이를 보고합니다. [testing.T.Run](https://pkg.go.dev/testing#T.Run)과 테스트 보고 API를 참고합니다.

### recorder는 요청마다 새로 만듭니다

기록기는 이전 요청의 상태와 본문을 기억합니다. 같은 recorder에 다음 요청을 보내면 앞선 상태나 본문이 섞일 수 있습니다. 요청 본문도 한 번 읽은 스트림이므로 같은 요청 객체를 무심코 재사용하지 않습니다. 각 테스트 사례 안에서 요청과 recorder를 새로 만들면 경계가 분명합니다.

라우터는 여기서 한 번 만들고 재사용합니다. 서비스의 처리 규칙은 여러 요청에 재사용되는 것이 정상입니다. 요청별 관찰 장치와 서비스 객체의 재사용 여부를 다르게 결정한 것입니다.

## 4. 필수 필드와 값의 유효성을 구분하기

일반 `int` 필드는 JSON에서 없을 때와 명시적인 0일 때 모두 영값으로 남을 수 있습니다. 그 차이가 계약에 중요하면 표현을 바꿔야 합니다. 요청용 구조체에 포인터 필드를 사용하면 nil과 실제 값을 구분할 수 있습니다.

저장 경로: [json-contract/main.go](../../examples/23-api-testing/json-contract/main.go)

전체 코드:

```go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"unicode/utf8"
)

type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

type CreateRequest struct {
	Title   *string `json:"title"`
	Minutes *int    `json:"minutes"`
}

type Lesson struct {
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

func decodeRequest(w http.ResponseWriter, r *http.Request) (CreateRequest, error) {
	var input CreateRequest
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
		return input, errors.New("expected one JSON value")
	}
	return input, nil
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /lessons/preview", func(w http.ResponseWriter, r *http.Request) {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			writeProblem(w, 415, "unsupported_media_type", "JSON 형식이 필요합니다")
			return
		}
		input, err := decodeRequest(w, r)
		if err != nil {
			var large *http.MaxBytesError
			if errors.As(err, &large) {
				writeProblem(w, 413, "body_too_large", "본문은 256바이트까지입니다")
			} else {
				writeProblem(w, 400, "invalid_json", "JSON 값을 확인하세요")
			}
			return
		}
		if input.Title == nil || input.Minutes == nil {
			writeProblem(w, 422, "missing_field", "title과 minutes가 필요합니다")
			return
		}
		title := strings.TrimSpace(*input.Title)
		if n := utf8.RuneCountInString(title); n < 1 || n > 40 {
			writeProblem(w, 422, "invalid_title", "제목은 1~40글자입니다")
			return
		}
		if *input.Minutes < 1 || *input.Minutes > 180 {
			writeProblem(w, 422, "invalid_minutes", "분량은 1~180입니다")
			return
		}
		writeJSON(w, 200, Lesson{Title: title, Minutes: *input.Minutes})
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/lessons/preview", `{"title":" Go ","minutes":25}`},
		{http.MethodPost, "/lessons/preview", `{"title":"Go"}`},
		{http.MethodPost, "/lessons/preview", `{"title":"Go","minutes":0}`},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		fmt.Printf("%d %s\n", recorder.Code, strings.TrimSpace(recorder.Body.String()))
	}
}
```

실행:

```text
go run ./examples/23-api-testing/json-contract
```

예상 출력:

```text
200 {"title":"Go","minutes":25}
422 {"code":"missing_field","message":"title과 minutes가 필요합니다"}
422 {"code":"invalid_minutes","message":"분량은 1~180입니다"}
```

### 누락과 null, 0

| `minutes` 입력 | 디코딩한 포인터 | 결과 |
| --- | --- | --- |
| 필드 없음 | nil | `missing_field` |
| `null` | nil | `missing_field` |
| `0` | 0을 가리키는 포인터 | `invalid_minutes` |
| `1` | 1을 가리키는 포인터 | 성공 |

포인터 필드는 “없음 또는 null”과 “실제 값”을 구분합니다. 없음과 null 자체를 다시 구분하는 모델은 아닙니다. 이번 API는 둘 다 필수 값이 없다는 같은 오류로 처리합니다. 포인터를 역참조하기 전에 nil을 확인하는 순서를 지킵니다.

### 네 층의 실패 계약

| 실패 | 상태 | 코드 |
| --- | --- | --- |
| JSON이 아닌 미디어 타입 | 415 | `unsupported_media_type` |
| 256바이트 초과 | 413 | `body_too_large` |
| 문법·알 수 없는 필드·추가 값 | 400 | `invalid_json` |
| 필수 값 없음 | 422 | `missing_field` |
| 제목 범위 오류 | 422 | `invalid_title` |
| 분량 범위 오류 | 422 | `invalid_minutes` |

상태는 큰 분류이고 `code`는 이 API 안의 안정적인 세부 분류입니다. 사람에게 보이는 한국어 메시지를 바꾸어도 호출자가 오류 코드로 분기할 수 있게 합니다. 메시지 문장을 프로그램의 유일한 분기 기준으로 사용하지 않습니다.

### 엄격함의 범위를 정확히 말하기

이 장은 `encoding/json`의 기존 v1 API를 사용합니다. `DisallowUnknownFields`가 중복 키를 거절하거나 필드 이름의 대소문자까지 정확히 일치하도록 만드는 것은 아닙니다. 기본 v1의 필드 매칭과 중복 키 처리 규칙은 그대로 적용됩니다. 테스트에는 그 호환 동작을 확인하는 입력도 있습니다. [encoding/json의 보안 고려 사항](https://pkg.go.dev/encoding/json#hdr-Security_Considerations)을 확인합니다.

따라서 문서에는 “알려진 필드와 단일 JSON 값, 크기와 업무 범위를 검사한다”라고 적습니다. “모든 모호한 JSON을 거절한다”라고 확대해서 말하지 않습니다. 필드 추가에 대한 호환 정책도 요청과 응답에서 다르게 정할 수 있습니다. 알 수 없는 필드를 거절하면 오타를 찾기 좋지만 미래의 확장 입력도 거절할 수 있습니다.

## 5. 미들웨어를 독립된 경계로 만들기

미들웨어는 다른 핸들러의 앞뒤에 공통 처리를 붙이는 함수입니다. 여기서는 요청마다 ID를 만들고 응답 헤더와 요청 컨텍스트에 넣습니다. 개별 핸들러는 ID 생성 방법을 몰라도 컨텍스트에서 값을 읽을 수 있습니다.

저장 경로: [request-id/main.go](../../examples/23-api-testing/request-id/main.go)

전체 코드:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
)

type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

type requestIDKey struct{}

func requestID(ctx context.Context) string {
	id, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return id
}

func withRequestID(next http.Handler) http.Handler {
	var sequence atomic.Uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("r-%d", sequence.Add(1))
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			RequestID string `json:"request_id"`
			Ready     bool   `json:"ready"`
		}{RequestID: requestID(r.Context()), Ready: true}
		writeJSON(w, 200, result)
	})
	return withRequestID(mux)
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/ready", ""},
		{http.MethodGet, "/ready", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		fmt.Printf("%d %s\n", recorder.Code, strings.TrimSpace(recorder.Body.String()))
	}
}
```

실행:

```text
go run ./examples/23-api-testing/request-id
```

예상 출력:

```text
200 {"request_id":"r-1","ready":true}
200 {"request_id":"r-2","ready":true}
```

### 미들웨어의 입구와 출구

| 순서 | 동작 | 목적 |
| --- | --- | --- |
| 1 | 원자적 카운터 증가 | 동시에 와도 다른 번호 선택 |
| 2 | X-Request-ID 설정 | 호출자가 결과를 식별 |
| 3 | 컨텍스트에 ID 저장 | 내부 핸들러에 같은 값 전달 |
| 4 | `next.ServeHTTP` 호출 | 실제 요청 처리 |
| 5 | 반환 | 같은 요청의 처리를 마침 |

컨텍스트 키는 자체 타입으로 정의했습니다. 다른 코드가 같은 문자열 키를 사용하여 우연히 충돌하는 일을 줄이기 위한 방법입니다. [context.WithValue](https://pkg.go.dev/context#WithValue)의 키 사용 지침을 참고합니다.

생성된 ID는 이 프로세스의 이 미들웨어 인스턴스 안에서 증가합니다. 서비스 여러 대 전체에서 유일한 분산 ID를 구현한 것은 아닙니다. 또한 요청 추적용 값이지 인증 증거가 아닙니다. 외부에서 보낸 `X-Request-ID`는 이번 계약에서는 채택하지 않고 새 ID로 대체합니다.

### 무엇을 테스트해야 할까요?

헤더에 값이 있다는 것만으로는 충분하지 않습니다. 핸들러가 받은 컨텍스트의 ID와 응답 헤더가 같아야 합니다. 다음 핸들러를 정확히 한 번 호출해야 하며, 동시에 들어온 요청의 ID도 중복되지 않아야 합니다.

래퍼를 라우터 바깥에 두었으므로 라우터가 반환하는 404에도 요청 ID 헤더가 적용됩니다. 반면 특정 핸들러만 감싸면 그 범위에만 적용됩니다. 미들웨어의 위치도 테스트할 계약입니다.

## 6. 저장소 실패를 재현 가능하게 만들기

실제 저장소가 가끔 실패할 때까지 기다리는 방식으로 500 응답을 검사할 필요는 없습니다. 핸들러가 필요한 작은 인터페이스를 정의하고 결과를 돌려줄 구현을 주입합니다. 실행 예제는 읽기 전용 메모리 저장소를 사용하며 테스트는 원하는 오류를 반환하는 구현을 사용합니다.

저장 경로: [injected-store/main.go](../../examples/23-api-testing/injected-store/main.go)

전체 코드:

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

var ErrNotFound = errors.New("lesson not found")

type Lesson struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type LessonStore interface {
	Lookup(context.Context, int) (Lesson, error)
}

type memoryStore struct{ lessons map[int]Lesson }

func (s memoryStore) Lookup(ctx context.Context, id int) (Lesson, error) {
	if err := ctx.Err(); err != nil {
		return Lesson{}, err
	}
	lesson, ok := s.lessons[id]
	if !ok {
		return Lesson{}, ErrNotFound
	}
	return lesson, nil
}

func newHandler(store LessonStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 {
			writeProblem(w, 400, "invalid_id", "ID는 양수 정수입니다")
			return
		}
		lesson, err := store.Lookup(r.Context(), id)
		if err != nil {
			if r.Context().Err() != nil {
				return
			}
			if errors.Is(err, ErrNotFound) {
				writeProblem(w, 404, "not_found", "학습을 찾을 수 없습니다")
			} else {
				writeProblem(w, 500, "internal_error", "요청을 처리하지 못했습니다")
			}
			return
		}
		writeJSON(w, 200, lesson)
	})
	return mux
}

func main() {
	handler := newHandler(memoryStore{lessons: map[int]Lesson{1: {ID: 1, Title: "Go"}}})
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/lessons/1", ""},
		{http.MethodGet, "/lessons/2", ""},
		{http.MethodGet, "/lessons/x", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		fmt.Printf("%d %s\n", recorder.Code, strings.TrimSpace(recorder.Body.String()))
	}
}
```

실행:

```text
go run ./examples/23-api-testing/injected-store
```

예상 출력:

```text
200 {"id":1,"title":"Go"}
404 {"code":"not_found","message":"학습을 찾을 수 없습니다"}
400 {"code":"invalid_id","message":"ID는 양수 정수입니다"}
```

### 요청 처리 순서와 호출 횟수

| 상황 | 저장소 호출 | 상태 | 코드 |
| --- | --- | --- | --- |
| 잘못된 ID | 하지 않음 | 400 | `invalid_id` |
| 데이터 발견 | 한 번 | 200 | 성공 데이터 |
| `ErrNotFound` | 한 번 | 404 | `not_found` |
| 다른 저장소 오류 | 한 번 | 500 | `internal_error` |
| 요청 컨텍스트 종료 | 작업에 전달 | 전송 종료 상황 | 새 성공 응답을 쓰지 않음 |

핸들러는 입력을 검증한 뒤 저장소를 호출합니다. 따라서 잘못된 ID가 저장소에 전달되지 않는 것도 계약으로 검사할 수 있습니다. 오류를 `%w`로 감쌌어도 `errors.Is(err, ErrNotFound)`로 분류합니다.

내부 저장소 오류의 문자열은 사용자 응답에 넣지 않습니다. 외부에는 안정적인 일반 오류를, 내부 운영에는 별도의 로그·관측 정보를 남기는 구조로 확장합니다. 로그와 요청 식별자는 운영 기초 장에서 연결합니다. 여기서는 오류를 무시하는 대신 상태와 코드로 실패를 명확히 반환합니다.

### 실제 실패 주입 테스트

다음은 [injected-store/behavior_test.go](../../examples/23-api-testing/injected-store/behavior_test.go) 전체입니다. 테스트의 `storeFunc`는 함수 값을 저장소 인터페이스로 바꾸는 작은 어댑터입니다.

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

type storeFunc func(context.Context, int) (Lesson, error)

func (f storeFunc) Lookup(ctx context.Context, id int) (Lesson, error) { return f(ctx, id) }

type markerKey struct{}

func TestLookupOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"found", nil, 200, ""},
		{"missing", ErrNotFound, 404, "not_found"},
		{"wrapped missing", fmt.Errorf("lookup: %w", ErrNotFound), 404, "not_found"},
		{"internal", errors.New("private database detail"), 500, "internal_error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			store := storeFunc(func(ctx context.Context, id int) (Lesson, error) {
				calls++
				if id != 7 || ctx.Value(markerKey{}) != "passed" {
					t.Errorf("id=%d context=%v", id, ctx.Value(markerKey{}))
				}
				return Lesson{ID: 7, Title: "Go"}, tc.err
			})
			request := httptest.NewRequest("GET", "/lessons/7", nil)
			request = request.WithContext(context.WithValue(request.Context(), markerKey{}, "passed"))
			recorder := httptest.NewRecorder()
			newHandler(store).ServeHTTP(recorder, request)
			if recorder.Code != tc.status || calls != 1 {
				t.Fatalf("status=%d calls=%d", recorder.Code, calls)
			}
			if strings.Contains(recorder.Body.String(), "private database detail") {
				t.Fatal("internal error leaked")
			}
			if tc.status == 200 {
				var lesson Lesson
				if err := json.Unmarshal(recorder.Body.Bytes(), &lesson); err != nil || lesson.ID != 7 {
					t.Fatalf("lesson=%+v err=%v", lesson, err)
				}
			} else {
				var problem Problem
				if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil || problem.Code != tc.code {
					t.Fatalf("problem=%+v err=%v", problem, err)
				}
			}
		})
	}
}

func TestInvalidIDDoesNotCallStore(t *testing.T) {
	calls := 0
	store := storeFunc(func(context.Context, int) (Lesson, error) { calls++; return Lesson{}, nil })
	handler := newHandler(store)
	for _, path := range []string{"/lessons/x", "/lessons/0", "/lessons/-1"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
		if recorder.Code != 400 {
			t.Errorf("path=%s status=%d", path, recorder.Code)
		}
	}
	if calls != 0 {
		t.Errorf("invalid requests called store %d times", calls)
	}
}
```

실행:

```text
go test -v ./examples/23-api-testing/injected-store
```

### 테스트가 구현을 대신하지 않게 하기

테스트용 저장소는 응답 상태를 결정하지 않습니다. 오직 조회 결과와 오류를 반환합니다. 그 결과를 404나 500으로 바꾸는 일은 실제 핸들러가 하므로, 그 변환이 잘못되면 테스트가 실패합니다.

테스트 안에서 핸들러와 똑같은 조건문으로 기대 상태를 다시 계산하지 않았습니다. 입력 표에 계약상 기대 결과를 직접 적었습니다. 그래야 구현의 조건문을 잘못 바꾸었을 때 기대값도 함께 같은 방식으로 틀리는 일을 줄일 수 있습니다.

메모리 저장소의 맵은 생성한 뒤 읽기만 합니다. 동시에 쓰는 기능을 추가하려면 그 순간 동기화 책임도 추가해야 합니다. 읽기 전용 예제의 통과를 임의의 맵 변경까지 안전하다는 근거로 사용하지 않습니다.

## 7. 실제 HTTP 왕복으로 다른 경계를 확인하기

recorded 응답만으로 TLS 연결, 클라이언트 읽기, HEAD 전송 본문을 모두 검증할 수는 없습니다. 이 예제는 같은 조회 계약을 실제 로컬 TLS 서버와 클라이언트 사이에 둡니다. 앞선 단위 검사를 없애는 것이 아니라 다른 경계를 추가합니다.

저장 경로: [integration/main.go](../../examples/23-api-testing/integration/main.go)

전체 코드:

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
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
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

var ErrNotFound = errors.New("lesson not found")

type Lesson struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type LessonStore interface {
	Lookup(context.Context, int) (Lesson, error)
}

type memoryStore struct{ lessons map[int]Lesson }

func (s memoryStore) Lookup(ctx context.Context, id int) (Lesson, error) {
	if err := ctx.Err(); err != nil {
		return Lesson{}, err
	}
	lesson, ok := s.lessons[id]
	if !ok {
		return Lesson{}, ErrNotFound
	}
	return lesson, nil
}

func newHandler(store LessonStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 {
			writeProblem(w, 400, "invalid_id", "ID는 양수 정수입니다")
			return
		}
		lesson, err := store.Lookup(r.Context(), id)
		if err != nil {
			if r.Context().Err() != nil {
				return
			}
			if errors.Is(err, ErrNotFound) {
				writeProblem(w, 404, "not_found", "학습을 찾을 수 없습니다")
			} else {
				writeProblem(w, 500, "internal_error", "요청을 처리하지 못했습니다")
			}
			return
		}
		writeJSON(w, 200, lesson)
	})
	return mux
}

func inspect(client *http.Client, address, method string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1025))
	if err != nil {
		return 0, "", err
	}
	if len(body) > 1024 {
		return 0, "", errors.New("response too large")
	}
	return resp.StatusCode, string(body), nil
}

func run() error {
	handler := newHandler(memoryStore{lessons: map[int]Lesson{1: {ID: 1, Title: "Go"}}})
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path string }{{"GET", "/lessons/1"}, {"HEAD", "/lessons/1"}, {"GET", "/lessons/2"}} {
		code, body, err := inspect(client, server.URL+tc.path, tc.method)
		if err != nil {
			return err
		}
		if body == "" {
			fmt.Printf("%s %d (빈 본문)\n", tc.method, code)
		} else {
			fmt.Printf("%s %d %s\n", tc.method, code, strings.TrimSpace(body))
		}
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
go run ./examples/23-api-testing/integration
```

예상 출력:

```text
GET 200 {"id":1,"title":"Go"}
HEAD 200 (빈 본문)
GET 404 {"code":"not_found","message":"학습을 찾을 수 없습니다"}
```

### 직접 호출과 전송 검사의 차이

| 관찰 대상 | recorder 직접 호출 | 실제 임시 서버 |
| --- | --- | --- |
| 핸들러의 상태·본문 작성 | 확인 가능 | 확인 가능 |
| 라우터의 경로 선택 | 라우터를 호출하면 가능 | 확인 가능 |
| TLS 신뢰와 연결 | 포함하지 않음 | 포함 |
| HEAD의 전송 본문 생략 | 전송 계층 없음 | 확인 가능 |
| 응답 본문의 읽기·닫기 | 클라이언트 없음 | 확인 가능 |
| 외부 서비스 장애 | 포함하지 않음 | 이 로컬 실험에도 포함하지 않음 |

`server.Client()`는 테스트 TLS 인증서를 신뢰하도록 구성되어 있습니다. 응답 본문은 크기 제한을 두어 읽고 닫습니다. 임시 서버는 `defer`로 종료되므로 테스트마다 별도 프로세스나 포트를 수동 관리하지 않습니다. [httptest.Server.Client](https://pkg.go.dev/net/http/httptest#Server.Client)를 참고합니다.

### 통합 테스트의 기대값도 구체적으로

“요청이 오류 없이 끝났다”만 확인하면 404를 성공처럼 넘길 수 있습니다. 이 테스트는 각 메서드·경로의 상태를 확인하고 성공 데이터의 ID도 검사합니다. HEAD에서는 본문이 실제로 빈 문자열인지 확인합니다. 21장의 클라이언트 계약과 22장의 서버 계약을 연결하는 검사입니다.

통합 테스트를 무조건 많이 늘리는 것이 목적은 아닙니다. 순수 값 검증은 함수 테스트로, HTTP 변환은 핸들러 테스트로, 전송의 특성은 왕복 테스트로 나누면 실패 원인을 찾기 쉽습니다. 외부 서비스가 없어도 기본 오류 경로와 계약을 충분히 재현할 수 있습니다.

## 8. 동시 요청의 결과를 검사하기

요청 하나가 정확하더라도 동시에 들어오면 공유 상태가 틀릴 수 있습니다. 카운터의 `읽기 → 1 더하기 → 쓰기`를 보호하지 않으면 일부 증가가 사라질 수 있습니다. 이 예제는 원자적 증가와 읽기를 사용합니다.

저장 경로: [concurrent-counter/main.go](../../examples/23-api-testing/concurrent-counter/main.go)

전체 코드:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
)

type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

func newHandler() http.Handler {
	var visits atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("POST /visits", func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			Count int64 `json:"count"`
		}{Count: visits.Add(1)}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("GET /visits", func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			Count int64 `json:"count"`
		}{Count: visits.Load()}
		writeJSON(w, 200, result)
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/visits", ""},
		{http.MethodPost, "/visits", ""},
		{http.MethodPost, "/visits", ""},
		{http.MethodGet, "/visits", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		fmt.Printf("%d %s\n", recorder.Code, strings.TrimSpace(recorder.Body.String()))
	}
}
```

실행:

```text
go run ./examples/23-api-testing/concurrent-counter
```

예상 출력:

```text
200 {"count":1}
200 {"count":2}
200 {"count":3}
200 {"count":3}
```

### 순서가 아니라 불변 조건을 검사하기

기본 프로그램은 출력을 이해하기 위해 요청을 순서대로 보냅니다. 동작 테스트는 50개의 POST를 동시에 보냅니다. 어느 고루틴이 먼저 끝날지는 정답으로 고정하지 않습니다.

| 검사 | 이유 |
| --- | --- |
| 응답 50개가 모두 성공 | 일부 요청의 누락 확인 |
| 각 값이 1~50 | 범위 밖 결과 확인 |
| 중복 값 없음 | 증가 결과의 중복 확인 |
| 최종 GET이 50 | 사라진 증가 확인 |
| GET은 값을 바꾸지 않음 | 조회의 업무 계약 확인 |

`atomic.Int64.Add(1)`은 원자적으로 증가한 새 값을 반환합니다. [atomic.Int64.Add](https://pkg.go.dev/sync/atomic#Int64.Add)의 계약을 사용합니다. 여러 필드를 함께 바꾸어야 하는 거래 전체를 원자 변수 하나로 자동 해결하는 것은 아닙니다. 그 경우 상태 구조와 잠금·트랜잭션 경계를 따로 정합니다.

POST의 응답 전송이 끊겼다고 카운터 증가가 없었다고 단정할 수는 없습니다. 서버에서 이미 변경한 뒤 응답만 전달되지 않았을 수도 있습니다. 중복 영향을 피하는 재시도나 멱등성 키는 추가 API 계약이며, 이 예제는 단일 요청의 증가와 동시성에 집중합니다.

### race detector와 의미 검사는 함께 사용합니다

```text
go test -race -count=1 ./examples/23-api-testing/...
```

race detector는 실행 중 충돌하는 메모리 접근을 찾는 도구입니다. 데이터 경쟁이 없더라도 잘못된 경로가 두 번 증가시키는 논리 오류는 있을 수 있습니다. 그래서 값의 범위, 중복, 최종 합계를 별도로 검사합니다. [Go race detector](https://go.dev/doc/articles/race_detector)를 참고합니다.

## 9. 직접 바꿔 보는 테스트 실습

테이블 테스트 예제의 실행 코드와 동작 테스트를 개인 폴더에 복사합니다. 기존 두 파일은 각각 보존합니다. `Example_main` 파일은 복사하지 않아 API 변경 실험과 문서 출력의 검사를 섞지 않습니다.

Windows PowerShell:

```powershell
$practiceDir = "_practice/23-api-testing/contract-lab"
New-Item -ItemType Directory -Force $practiceDir | Out-Null
foreach ($practiceFile in @("main.go", "behavior_test.go")) {
    $targetFile = Join-Path $practiceDir $practiceFile
    if (Test-Path -LiteralPath $targetFile) {
        Write-Host "$targetFile 파일이 이미 있습니다. 기존 내용을 확인하세요."
    } else {
        Copy-Item (Join-Path "examples/23-api-testing/handler-table" $practiceFile) $targetFile
    }
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
practice_dir="_practice/23-api-testing/contract-lab"
mkdir -p "$practice_dir"
for practice_file in main.go behavior_test.go; do
    if [ -e "$practice_dir/$practice_file" ]; then
        printf '%s\n' "$practice_dir/$practice_file 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp "examples/23-api-testing/handler-table/$practice_file" "$practice_dir/$practice_file"
    fi
done
```

```text
go test -v ./_practice/23-api-testing/contract-lab
go fmt ./_practice/23-api-testing/contract-lab
```

### 실험 A: 정상 상태지만 잘못된 데이터

개인 핸들러에서 인사말의 `name` 대신 고정된 다른 이름을 사용해 봅니다. 이는 의도한 회귀 실험입니다. 상태 코드는 여전히 200이지만 본문 필드 비교가 실패해야 합니다. 실패 메시지를 읽고 원래 동작으로 복구합니다.

### 실험 B: 오류 뒤에 성공 응답까지 작성

개인 코드에서 `invalid_name`을 응답한 뒤의 `return`을 잠시 제거합니다. 같은 요청에서 두 JSON 응답이 이어질 수 있습니다. 상태 검사만으로는 놓칠 수 있지만 JSON 해석이나 필드 검사가 실패해야 합니다. 복구한 뒤 전체 테이블이 다시 통과하는지 확인합니다.

### 실험 C: 경계 사례 하나 추가

이름 1룬인 `가`의 기대 메시지를 추가합니다. 기존 20·21룬 경계와 함께 검사를 실행합니다. 구현을 그대로 베껴 기대값을 계산하기보다 API 계약의 결과를 직접 적습니다.

### 실험 D: recorder를 루프 밖으로 옮기기

개인 테스트에서 recorder를 모든 사례가 공유하게 옮겨 봅니다. 이것 역시 의도한 잘못된 테스트 구성입니다. 이전 응답과 다음 응답이 섞이는 증상을 확인한 뒤 각 하위 테스트 안으로 복구합니다. 기본 예제의 테스트 파일은 변경하지 않습니다.

## 10. 자주 하는 실수와 검사 선택

| 증상 | 원인 후보 | 개선 |
| --- | --- | --- |
| 본문이 틀려도 테스트 통과 | 상태만 검사 | JSON 필드와 헤더 확인 |
| 두 번째 사례부터 이상한 본문 | recorder·request 재사용 | 사례마다 새로 구성 |
| 실패 테스트가 오류를 출력하지만 PASS | Log만 사용 | Error/Fatal로 실패 보고 |
| 래핑된 not found가 500 | 문자열 또는 직접 동일성만 검사 | `errors.Is` 사용 |
| 저장소 장애를 재현하기 어려움 | 핸들러가 저장소를 직접 생성 | 작은 인터페이스 주입 |
| 이름 길이가 한글에서 달라짐 | 바이트와 룬 혼동 | 계약의 길이 단위 고정 |
| 누락과 0이 같은 결과 | 일반 정수 필드만 사용 | 존재 구분이 필요하면 포인터 모델 |
| 동시 테스트 출력 순서가 바뀜 | 실행 순서를 정답으로 가정 | 결과 집합·최종 상태 검사 |
| 요청 ID는 있는데 내부 값이 다름 | 헤더·컨텍스트를 따로 생성 | 동일한 값 전달 검사 |
| HEAD 검사 결과가 예상과 다름 | 직접 호출과 전송 검사 혼동 | 실제 HTTP 왕복 추가 |

테스트 실패를 없애기 위해 기대값을 실제 출력으로 무조건 바꾸지 않습니다. 계약이 바뀐 것인지 구현이 틀린 것인지 먼저 판단합니다. 계약 변경이라면 본문·문제·클라이언트 기대와 테스트를 함께 개정합니다.

## 11. 연습문제

[23장 연습문제](../../exercises/23-api-testing/README.md)를 먼저 풀고 [해설](../../solutions/23-api-testing/README.md)과 비교합니다. 이번 문제에는 구현뿐 아니라 계약을 확인하는 테스트도 포함됩니다.

| 문제 | 중심 목표 |
| --- | --- |
| 1. 상태만 검사하면? | 테스트가 놓치는 실패 설명 |
| 2. 짧은 메시지 API | 입력 표와 헤더·본문 검사 |
| 3. 기록기 재사용 고치기 | 테스트 사례의 상태 분리 |
| 4. 필수 점수 입력 | 누락·null·0·경계 분류 |
| 5. 저장소 실패 처리 | 의존성 주입과 안정적인 오류 |
| 6. 요청 ID 미들웨어 | 호출 횟수·컨텍스트·동시성 |
| 선택 A. API 전체의 JSON 오류 | 경로·메서드 실패 형식 통일 |

## 핵심 정리와 자기 점검

- 계약은 입력과 상태뿐 아니라 헤더·본문·부작용·취소까지 포함합니다.
- 순수 검증, 핸들러 직접 검사, 실제 왕복은 다른 경계를 확인합니다.
- 요청별 recorder와 request를 새로 만들고 서비스의 공유 상태는 의도적으로 재사용합니다.
- 포인터 필드는 nil과 실제 영값을 구분할 수 있습니다.
- 의존성은 실제 핸들러가 결과를 처리하도록 두고 테스트에서 입력 결과를 주입합니다.
- 동시 테스트는 순서보다 결과의 불변 조건을 검사합니다.

### 확인 질문

1. 상태 200만 검사하는 테스트는 어떤 오류를 놓치나요?
2. JSON 필드 비교와 정확한 stdout 비교의 목적은 어떻게 다른가요?
3. 요청을 여러 번 보낼 때 recorder를 새로 만드는 이유는 무엇인가요?
4. `*int` 필드에서 없음·null·0은 어떻게 구분되나요?
5. `DisallowUnknownFields`가 중복 키도 금지하나요?
6. 잘못된 ID에서 저장소 호출 횟수를 확인하는 이유는 무엇인가요?
7. 테스트 저장소가 404를 직접 쓰게 만들면 어떤 경계를 놓치나요?
8. 같은 요청 ID를 헤더와 컨텍스트에 연결해야 하는 이유는 무엇인가요?
9. HEAD 본문을 검사할 때 실제 전송 검사가 필요한 이유는 무엇인가요?
10. race detector와 최종 카운터 값 검사를 둘 다 하는 이유는 무엇인가요?

<details>
<summary>답과 설명</summary>

1. 잘못된 데이터, 없는 헤더, 중복 본문, 불필요한 저장소 호출 등을 놓칠 수 있습니다.
2. 앞은 API 데이터 의미, 뒤는 교재 프로그램이 보여 주는 출력 형식을 확인합니다.
3. 앞선 상태·헤더·본문이 다음 사례에 남지 않게 하기 위해서입니다.
4. 없음과 null은 nil, 명시적 0은 0을 가리키는 포인터입니다.
5. 아닙니다. 기본 JSON v1의 중복 키·대소문자 규칙은 별개입니다.
6. 입력 검증 전에 의존성을 호출하는 불필요한 부작용을 찾을 수 있습니다.
7. 저장소 오류를 HTTP 상태로 바꾸는 실제 핸들러의 로직을 검사하지 못합니다.
8. 호출자가 받은 응답과 내부 처리 기록을 같은 요청으로 연결하기 위해서입니다.
9. 직접 recorder 호출에는 HEAD 본문을 생략하는 HTTP 전송 계층이 없기 때문입니다.
10. 메모리 충돌과 의미상의 누락·중복은 서로 다른 종류의 결함이기 때문입니다.

</details>

### 완료 체크리스트

- [ ] 일곱 예제의 출력과 동작 테스트를 확인했다.
- [ ] 상태·헤더·본문·의존성 호출을 나눠 검사했다.
- [ ] 누락과 영값을 구분하는 요청 모델을 이해했다.
- [ ] 내부 오류를 주입하여 404·500 계약을 확인했다.
- [ ] 미들웨어 값의 연결과 동시 요청의 불변 조건을 검사했다.
- [ ] 개인 회귀 실험을 복구하고 테스트를 다시 통과시켰다.
- [ ] 기본 여섯 문제와 경계 입력을 확인했다.

## 공식 참고 자료

- [httptest](https://pkg.go.dev/net/http/httptest) — 요청·기록기·임시 서버
- [testing](https://pkg.go.dev/testing) — 하위 테스트와 실패 보고
- [net/http](https://pkg.go.dev/net/http) — 핸들러·응답·라우팅 계약
- [encoding/json](https://pkg.go.dev/encoding/json) — 필드 매칭과 JSON 해석
- [context.WithValue](https://pkg.go.dev/context#WithValue) — 요청 범위 값과 키
- [sync/atomic](https://pkg.go.dev/sync/atomic) — 원자적 카운터
- [Go race detector](https://go.dev/doc/articles/race_detector) — 동시 접근 검사
- [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html) — HTTP 메서드와 상태의 의미

확인일: 2026-11-09. Go 1.27.3과 `encoding/json` v1 API를 기준으로 확인했습니다.

## 다음 학습

[24장 SQL과 트랜잭션](../24-sql-transactions/README.md)에서 저장소 인터페이스 뒤에 실제 SQL 동작을 연결합니다. 쿼리 결과·자원 수명·실패 시 롤백을 테스트하며 HTTP 처리와 저장소 책임을 이어 갑니다.

| 이전 | 목차 | 실습·해설 | 다음 |
| --- | --- | --- | --- |
| [22장](../22-http-server/README.md) | [전체 목차](../README.md) | [문제](../../exercises/23-api-testing/README.md) · [해설](../../solutions/23-api-testing/README.md) | [24장](../24-sql-transactions/README.md) |
