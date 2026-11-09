# 23장 해설: API 설계와 핸들러 테스트

문제: [23장 연습문제](../../exercises/23-api-testing/README.md)

구현이 어떤 결과를 만드는지와 테스트가 어떤 약속을 확인하는지를 함께 읽습니다.
HTTP 요청은 모두 메모리의 recorder 또는 로컬 임시 서버에서 처리합니다.
모든 명령은 **저장소 루트** 기준입니다.
개인 코드와 참고 구현의 테스트 대상은 별개입니다.

## 문제 1. 테스트가 놓치는 실패

상태만 검사하면 A·B·C가 통과하지만 전체 계약을 만족하는 것은 A뿐입니다.
B는 JSON 해석 뒤 `message` 필드 비교가 필요합니다.
C는 Content-Type과 JSON 문법 검사에서 발견할 수 있습니다.
D는 상태 검사부터 실패합니다.

| 사례 | 상태만 검사 | 전체 계약 |
| --- | --- | --- |
| A | 통과 | 만족 |
| B | 통과 | 메시지 값이 다름 |
| C | 통과 | 형식과 본문이 다름 |
| D | 실패 | 상태가 다름 |

다음은 **잘못된 응답을 일부러 만들어 상태만 검사하는 한계를 관찰하는 전체 프로그램**입니다.
개인 경로 `_practice/23-api-testing/predict/main.go`에서 실행할 수 있습니다.
기존 파일이 있다면 먼저 내용을 보존하세요.

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, `{"message":"wrong"}`); err != nil {
			return
		}
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &body)
	fmt.Println("상태만 검사:", recorder.Code == 200)
	fmt.Println("본문 계약:", err == nil && body.Message == "Go")
	fmt.Println("형식 계약:", recorder.Header().Get("Content-Type") == "application/json")
}
```

```text
go run ./_practice/23-api-testing/predict
```

예상 출력:

```text
상태만 검사: true
본문 계약: false
형식 계약: false
```

API의 JSON 필드 의미를 비교할 때는 구조체로 해석하여 필요한 필드를 검사합니다.
필드 순서나 공백이 계약이 아니라면 원본 JSON 문자열의 완전 일치에 과도하게 의존하지 않습니다.
반면 위 stdout은 교재가 보여 주는 결과 자체이므로 출력 테스트에서 고정합니다.

## 문제 2. 짧은 메시지 API

참고 코드: [echo-contract/main.go](echo-contract/main.go)

전체 구현:

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
	mux.HandleFunc("GET /echo", func(w http.ResponseWriter, r *http.Request) {
		message := strings.TrimSpace(r.URL.Query().Get("message"))
		length := utf8.RuneCountInString(message)
		if length < 1 || length > 8 {
			writeProblem(w, http.StatusBadRequest, "invalid_message", "메시지는 1~8글자입니다")
			return
		}
		result := struct {
			Message string `json:"message"`
		}{Message: message}
		writeJSON(w, http.StatusOK, result)
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/echo?message=Go", ""},
		{http.MethodGet, "/echo", ""},
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

참고 구현 실행:

```text
go run ./solutions/23-api-testing/echo-contract
```

예상 출력:

```text
200 {"message":"Go"}
400 {"code":"invalid_message","message":"메시지는 1~8글자입니다"}
```

### 구현을 읽기

쿼리 문자열을 해석한 뒤 앞뒤 공백을 제거합니다.
룬 수가 1~8이면 정규화한 값을 반환하고, 아니면 400과 `invalid_message`를 반환합니다.
실패 응답 뒤에는 return이 있어 성공 응답을 이어 쓰지 않습니다.

[동작 테스트](echo-contract/behavior_test.go) 전체:

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

func TestEchoContract(t *testing.T) {
	tests := []struct {
		name, input   string
		status        int
		message, code string
	}{
		{"normal", "Go", 200, "Go", ""},
		{"trimmed", " Go ", 200, "Go", ""},
		{"maximum", strings.Repeat("가", 8), 200, strings.Repeat("가", 8), ""},
		{"missing", "", 400, "메시지는 1~8글자입니다", "invalid_message"},
		{"blank", "  ", 400, "메시지는 1~8글자입니다", "invalid_message"},
		{"too long", strings.Repeat("가", 9), 400, "메시지는 1~8글자입니다", "invalid_message"},
	}
	handler := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/echo?message="+url.QueryEscape(tc.input), nil)
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

```text
go test -v ./solutions/23-api-testing/echo-contract
```

개인 코드 검사:

```text
go run ./_practice/23-api-testing/echo-contract
go test -v ./_practice/23-api-testing/echo-contract
```

### 테스트의 관찰값

| 관찰값 | 계약상 기대 |
| --- | --- |
| 상태 | 정상 200, 범위 실패 400 |
| Content-Type | application/json |
| JSON 해석 | 오류 없이 값 하나를 해석 |
| 정상 메시지 | 정규화한 입력 |
| 오류 코드 | invalid_message |

테이블의 기대 메시지는 API가 약속한 값입니다.
테스트에서 실제 핸들러의 결과를 다시 기대값으로 사용하지 않습니다.
한글 8룬과 9룬을 나누어 바이트 수로 잘못 검사하는 구현도 발견합니다.

## 문제 3. recorder를 각 사례 안에서 만들기

잘못된 테스트는 recorder 하나에 두 응답을 기록합니다.
첫 번째 200 상태가 남고 본문도 이어지므로 다음 400 요청의 독립된 결과가 아닙니다.
수정은 각 사례에서 새 요청과 새 recorder를 만드는 것입니다.

개인 `main.go`는 준비 단계에서 복사한 [handler-table/main.go](../../examples/23-api-testing/handler-table/main.go)를 그대로 사용합니다.
테스트를 맞추려고 핸들러를 변경하지 않습니다.
개인 `_practice/23-api-testing/repair/behavior_test.go`의 전체 수정 코드는 다음과 같습니다.

```go
package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestGreetingResponses(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		path          string
		status        int
		message, code string
	}{
		{"/greet?name=Go", 200, "안녕하세요, Go!", ""},
		{"/greet", 400, "이름은 1~20글자입니다", "invalid_name"},
	} {
		request := httptest.NewRequest("GET", tc.path, nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != tc.status {
			t.Fatalf("path=%s status=%d want=%d", tc.path, recorder.Code, tc.status)
		}
		if recorder.Header().Get("Content-Type") != "application/json" {
			t.Fatal("wrong response type")
		}
		var body struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Message != tc.message || body.Code != tc.code {
			t.Errorf("path=%s body=%+v", tc.path, body)
		}
	}
}
```

```text
go test -v ./_practice/23-api-testing/repair
```

수정된 테스트는 두 요청을 각각 검사하고 통과해야 합니다.
첫 요청의 JSON과 다음 요청의 JSON을 같은 버퍼에서 해석하지 않습니다.

| 객체 | 재사용 여부 | 이유 |
| --- | --- | --- |
| 핸들러·라우터 | 재사용 | 여러 요청을 처리하는 서비스 규칙 |
| 요청 객체 | 사례마다 새로 생성 | 본문과 컨텍스트는 요청별 수명 |
| recorder | 사례마다 새로 생성 | 상태·헤더·본문은 응답별 기록 |
| 기대값 표 | 읽기 전용 사용 | 계약을 정의하는 데이터 |

이 수정은 테스트 격리를 위한 것입니다.
서비스가 가져야 하는 공유 상태까지 매번 새로 만들어 동시성·누적 결함을 숨기는 것과 구분합니다.
[ResponseRecorder](https://pkg.go.dev/net/http/httptest#ResponseRecorder)의 역할을 참고합니다.

## 문제 4. 필수 점수 입력

참고 코드: [strict-score/main.go](strict-score/main.go)

전체 구현:

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

type ScoreRequest struct {
	Score *int `json:"score"`
}

func decodeRequest(w http.ResponseWriter, r *http.Request) (ScoreRequest, error) {
	var input ScoreRequest
	r.Body = http.MaxBytesReader(w, r.Body, 64)
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

func classifyScore(score int) string {
	if score >= 90 {
		return "우수"
	}
	if score >= 60 {
		return "통과"
	}
	return "다시 연습"
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /score", func(w http.ResponseWriter, r *http.Request) {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			writeProblem(w, 415, "unsupported_media_type", "JSON 형식이 필요합니다")
			return
		}
		input, err := decodeRequest(w, r)
		if err != nil {
			var large *http.MaxBytesError
			if errors.As(err, &large) {
				writeProblem(w, 413, "body_too_large", "본문은 64바이트까지입니다")
			} else {
				writeProblem(w, 400, "invalid_json", "JSON 값을 확인하세요")
			}
			return
		}
		if input.Score == nil {
			writeProblem(w, 422, "missing_score", "score가 필요합니다")
			return
		}
		if *input.Score < 0 || *input.Score > 100 {
			writeProblem(w, 422, "invalid_score", "점수는 0~100입니다")
			return
		}
		result := struct {
			Label string `json:"label"`
		}{Label: classifyScore(*input.Score)}
		writeJSON(w, 200, result)
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/score", `{"score":0}`},
		{http.MethodPost, "/score", `{"score":72}`},
		{http.MethodPost, "/score", `{}`},
		{http.MethodPost, "/score", `{"score":null}`},
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

```text
go run ./solutions/23-api-testing/strict-score
```

예상 출력:

```text
200 {"label":"다시 연습"}
200 {"label":"통과"}
422 {"code":"missing_score","message":"score가 필요합니다"}
422 {"code":"missing_score","message":"score가 필요합니다"}
```

### 값 존재와 범위를 다른 단계로 검사하기

`ScoreRequest.Score`는 `*int`입니다.
없음과 null은 nil이며, 명시적 0은 실제 정수 0을 가리킵니다.
nil을 확인한 뒤 역참조하여 0~100인지 검사합니다.

| 입력 | 상태 | 결과 |
| --- | --- | --- |
| `{"score":0}` | 200 | 다시 연습 |
| `{"score":59}` | 200 | 다시 연습 |
| `{"score":60}` | 200 | 통과 |
| `{"score":89}` | 200 | 통과 |
| `{"score":90}` | 200 | 우수 |
| `{"score":100}` | 200 | 우수 |
| -1 또는 101 | 422 | invalid_score |
| 필드 없음 또는 null | 422 | missing_score |

`classifyScore`는 이미 유효하다고 확인한 점수를 판정하는 순수 함수입니다.
JSON을 읽고 HTTP 상태를 정하는 기능과 분리했습니다.
이 함수에 범위 밖 값을 넘기지 않는 책임은 호출자인 핸들러에 있습니다.

[동작 테스트](strict-score/behavior_test.go) 전체:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScoreBoundaries(t *testing.T) {
	for _, tc := range []struct {
		score, status int
		label, code   string
	}{{-1, 422, "", "invalid_score"}, {0, 200, "다시 연습", ""}, {59, 200, "다시 연습", ""}, {60, 200, "통과", ""}, {89, 200, "통과", ""}, {90, 200, "우수", ""}, {100, 200, "우수", ""}, {101, 422, "", "invalid_score"}} {
		req := httptest.NewRequest("POST", "/score", strings.NewReader(fmt.Sprintf(`{"score":%d}`, tc.score)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newHandler().ServeHTTP(w, req)
		var result struct {
			Label string `json:"label"`
			Code  string `json:"code"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		if w.Code != tc.status || err != nil || result.Label != tc.label || result.Code != tc.code {
			t.Errorf("score=%d status=%d body=%+v err=%v", tc.score, w.Code, result, err)
		}
	}
}
func TestScoreJSONContract(t *testing.T) {
	valid := `{"score":72}`
	for _, tc := range []struct {
		name, body, media string
		status            int
		code              string
	}{
		{"missing", "{}", "application/json", 422, "missing_score"}, {"null field", `{"score":null}`, "application/json", 422, "missing_score"}, {"null body", "null", "application/json", 422, "missing_score"},
		{"wrong type", `{"score":"72"}`, "application/json", 400, "invalid_json"}, {"unknown", `{"score":72,"extra":1}`, "application/json", 400, "invalid_json"},
		{"two values", valid + "{}", "application/json", 400, "invalid_json"}, {"trailing junk", valid + "x", "application/json", 400, "invalid_json"},
		{"empty", "", "application/json", 400, "invalid_json"}, {"media", valid, "text/plain", 415, "unsupported_media_type"},
		{"exact limit", valid + strings.Repeat(" ", 64-len(valid)), "application/json", 200, ""},
		{"too large", valid + strings.Repeat(" ", 65-len(valid)), "application/json", 413, "body_too_large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/score", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.media)
			w := httptest.NewRecorder()
			newHandler().ServeHTTP(w, req)
			if w.Code != tc.status || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("status=%d header=%s body=%s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
			}
			if tc.code != "" {
				var p Problem
				if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil || p.Code != tc.code {
					t.Fatalf("problem=%+v err=%v", p, err)
				}
			}
		})
	}
}
```

```text
go test -v ./solutions/23-api-testing/strict-score
```

개인 풀이:

```text
go run ./_practice/23-api-testing/strict-score
go test -v ./_practice/23-api-testing/strict-score
```

64바이트의 정상 객체와 65바이트 입력을 구분합니다.
공백을 덧붙인 JSON도 전체 바이트 한도에는 포함됩니다.
알 수 없는 필드, 두 번째 값과 뒤의 쓰레기 문자는 400으로 분류합니다.

`DisallowUnknownFields`는 기본 JSON v1의 중복 키·대소문자 규칙을 바꾸지 않습니다.
이 풀이의 입력 계약은 그 별도 제약까지 보장하지 않습니다.
[encoding/json](https://pkg.go.dev/encoding/json)의 규칙을 참고합니다.

## 문제 5. 저장소 실패 처리

참고 코드: [failing-repository/main.go](failing-repository/main.go)

전체 구현:

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
	if id == 3 {
		return Lesson{}, errors.New("private storage detail")
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
		{http.MethodGet, "/lessons/3", ""},
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

```text
go run ./solutions/23-api-testing/failing-repository
```

예상 출력:

```text
200 {"id":1,"title":"Go"}
404 {"code":"not_found","message":"학습을 찾을 수 없습니다"}
500 {"code":"internal_error","message":"요청을 처리하지 못했습니다"}
400 {"code":"invalid_id","message":"ID는 양수 정수입니다"}
```

### 실패를 주입하는 경계

실행용 메모리 저장소는 ID 1을 찾고, 2는 없으며, 3에서는 학습용 내부 오류를 반환합니다.
핸들러는 저장소가 구체적으로 어떤 구현인지 몰라도 인터페이스의 결과를 처리합니다.
400은 저장소를 부르기 전에 결정하고, 나머지 결과는 조회 이후 분류합니다.

`errors.Is`를 사용하므로 `%w`로 감싼 not found도 404가 됩니다.
다른 오류는 500과 안정적인 메시지로 바꾸어 내부 문자열을 그대로 응답하지 않습니다.
저장소가 HTTP writer를 직접 다루지 않기 때문에 그 변환을 실제 핸들러에서 검사할 수 있습니다.

[동작 테스트](failing-repository/behavior_test.go) 전체:

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

```text
go test -v ./solutions/23-api-testing/failing-repository
```

개인 풀이:

```text
go run ./_practice/23-api-testing/failing-repository
go test -v ./_practice/23-api-testing/failing-repository
```

### 결과만 검사하지 않은 이유

| 추가 검사 | 발견하는 문제 |
| --- | --- |
| invalid ID의 호출 횟수 0 | 검증 전 저장소 호출 |
| 유효한 ID의 호출 횟수 1 | 중복 조회 |
| 전달된 ID | 다른 데이터를 조회 |
| 컨텍스트 값 | 요청의 수명을 새 컨텍스트로 단절 |
| 내부 문자열이 응답에 없음 | 내부 오류를 그대로 공개 |

테스트용 컨텍스트 값은 전달 여부를 관찰하는 표식입니다.
실제 업무 인자를 숨겨서 전달하는 범용 저장 공간으로 컨텍스트를 사용하는 예는 아닙니다.
[context.WithValue](https://pkg.go.dev/context#WithValue)의 요청 범위 값 지침을 참고합니다.

메모리 저장소는 생성 뒤 읽기만 합니다.
변경 작업을 추가할 때는 별도의 동기화와 저장 계약이 필요합니다.
이 조회 실습의 인터페이스와 테스트 방식은 실제 SQL 저장소에도 적용할 수 있습니다.

## 문제 6. 요청 ID 미들웨어

참고 코드: [request-middleware/main.go](request-middleware/main.go)

전체 구현:

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
		{http.MethodGet, "/missing", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		fmt.Printf("%d %s id=%s\n", recorder.Code, strings.TrimSpace(recorder.Body.String()), recorder.Header().Get("X-Request-ID"))
	}
}
```

```text
go run ./solutions/23-api-testing/request-middleware
```

예상 출력:

```text
200 {"request_id":"r-1","ready":true} id=r-1
200 {"request_id":"r-2","ready":true} id=r-2
404 404 page not found id=r-3
```

### 미들웨어가 소유하는 상태

카운터는 `withRequestID`가 만들어질 때 하나 생깁니다.
각 요청에서는 그 카운터를 원자적으로 증가시키므로 공유 상태를 안전하게 다룹니다.
새 미들웨어를 만들면 별도 카운터가 생기며 전역 번호를 공유하지 않습니다.

응답 헤더를 먼저 정하고 같은 ID를 컨텍스트에 넣어 다음 핸들러를 호출합니다.
클라이언트에서 받은 ID를 그대로 믿는 정책은 사용하지 않았습니다.
요청 ID는 이 인스턴스의 추적용 값이며 인증이나 분산 시스템 전체의 유일성을 보장하지 않습니다.

[동작 테스트](request-middleware/behavior_test.go) 전체:

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRequestIDReplacesUntrustedHeader(t *testing.T) {
	var calls atomic.Int32
	handler := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		id := requestID(r.Context())
		if id == "" || id != w.Header().Get("X-Request-ID") {
			t.Errorf("context id=%q header=%q", id, w.Header().Get("X-Request-ID"))
		}
		w.WriteHeader(204)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "client-supplied")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if calls.Load() != 1 || w.Code != 204 || w.Header().Get("X-Request-ID") != "r-1" {
		t.Fatalf("calls=%d code=%d id=%q", calls.Load(), w.Code, w.Header().Get("X-Request-ID"))
	}
}
func TestRequestIDConcurrent(t *testing.T) {
	const count = 50
	handler := newHandler()
	ids := make(chan string, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", "/ready", nil))
			if w.Code != 200 {
				t.Errorf("status=%d", w.Code)
			}
			ids <- w.Header().Get("X-Request-ID")
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[string]bool)
	for id := range ids {
		if !strings.HasPrefix(id, "r-") || seen[id] {
			t.Errorf("invalid or duplicate id=%q", id)
		}
		seen[id] = true
	}
	if len(seen) != count {
		t.Errorf("unique=%d", len(seen))
	}
}

func TestRequestIDOnNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	newHandler().ServeHTTP(recorder, httptest.NewRequest("GET", "/missing", nil))
	if recorder.Code != 404 || recorder.Header().Get("X-Request-ID") != "r-1" {
		t.Fatalf("status=%d id=%q", recorder.Code, recorder.Header().Get("X-Request-ID"))
	}
}
```

```text
go test -race -v ./solutions/23-api-testing/request-middleware
```

개인 풀이:

```text
go run ./_practice/23-api-testing/request-middleware
go test -race ./_practice/23-api-testing/request-middleware
```

50개 요청의 완료 순서는 고정하지 않습니다.
ID가 비어 있지 않고 중복되지 않는 집합이라는 사실을 검사합니다.
각 요청의 recorder는 따로 만들고, 결과를 모은 뒤 하나의 고루틴에서 중복을 판단합니다.

404에도 헤더가 있는 이유는 미들웨어를 라우터 바깥에 두었기 때문입니다.
이 문제는 라우터 기본 404 본문을 그대로 사용합니다.
본문 형식까지 통일하는 요구는 다음 선택 과제의 범위입니다.

## 선택 A. API 전체의 오류 형식

개인 경로: `_practice/23-api-testing/uniform-errors/main.go`

전체 프로그램:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

func newHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			writeProblem(w, 404, "not_found", "경로를 찾을 수 없습니다")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			writeProblem(w, 405, "method_not_allowed", "GET 또는 HEAD를 사용하세요")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodHead {
			w.WriteHeader(200)
			return
		}
		writeJSON(w, 200, struct {
			Ready bool `json:"ready"`
		}{Ready: true})
	})
}
func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path string }{{"GET", "/health"}, {"POST", "/health"}, {"GET", "/missing"}, {"HEAD", "/health"}} {
		req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
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
			return fmt.Errorf("응답 한도 초과")
		}
		display := strings.TrimSpace(string(body))
		if display == "" {
			display = "(빈 본문)"
		}
		fmt.Printf("%s %s -> %d %s\n", tc.method, tc.path, resp.StatusCode, display)
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
go run ./_practice/23-api-testing/uniform-errors
```

예상 출력:

```text
GET /health -> 200 {"ready":true}
POST /health -> 405 {"code":"method_not_allowed","message":"GET 또는 HEAD를 사용하세요"}
GET /missing -> 404 {"code":"not_found","message":"경로를 찾을 수 없습니다"}
HEAD /health -> 200 (빈 본문)
```

경로를 먼저 확인하고 해당 경로의 허용 메서드를 판정합니다.
경로가 없는 POST를 무조건 405로 처리하지 않도록 순서를 정했습니다.
405에서는 Allow를 보낸 뒤 JSON 실패 응답을 씁니다.

성공 HEAD는 JSON Content-Type과 200만 보내고 반환합니다.
본문을 쓰지 않으며 실제 클라이언트로 빈 전송 본문을 확인합니다.
실제 API가 더 많은 경로를 가지면 같은 원칙을 라우팅·오류 변환 구조로 정리할 수 있습니다.

### 전체 범위가 달라진 점

| 경로 | 기본 문제의 처리 범위 | 선택 과제의 처리 범위 |
| --- | --- | --- |
| 등록한 정상 경로 | JSON 성공·실패 | JSON 성공·실패 |
| 알 수 없는 경로 | 라우터 기본 텍스트 가능 | JSON not_found |
| 허용하지 않는 메서드 | 라우터 기본 텍스트 가능 | JSON method_not_allowed와 Allow |

모든 경로의 실패가 JSON이라는 계약을 선언하려면 실제 기본 오류 경로까지 검사해야 합니다.
성공 경로와 특정 입력 실패만 보고 적용 범위를 넓혀 말하지 않습니다.

## 실행·검증 정리

```text
go test -count=1 ./solutions/23-api-testing/...
go vet ./solutions/23-api-testing/...
go test -race ./solutions/23-api-testing/...
```

기본 출력은 각 패키지의 `Example_main`으로 확인합니다.
동작 테스트는 상태·헤더·본문, 경계 입력과 의존성 호출을 확인합니다.
개인 풀이의 테스트도 개인 경로를 명시하여 따로 실행하세요.

다시 설명할 질문:

- 같은 200 응답이라도 왜 여러 assertion이 필요한가?
- 요청별로 새로 만들 상태와 서비스가 공유할 상태는 무엇인가?
- nil 포인터 필드와 명시적인 0은 어떤 계약을 표현하는가?
- 저장소 오류를 HTTP로 바꾸는 책임은 어디에 있는가?
- 동시 요청의 순서를 고정하지 않고도 무엇을 검증할 수 있는가?

[23장 본문](../../book/23-api-testing/README.md) · [문제](../../exercises/23-api-testing/README.md) · [해설 목록](../README.md)
