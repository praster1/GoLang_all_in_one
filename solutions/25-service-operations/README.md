# 25장 해설 — 서비스 경계를 코드로 확인하기

[문제](../../exercises/25-service-operations/README.md)를 먼저 풀고 비교합니다. 아래 코드는 유일한 구조가 아닙니다. 설정 우선순위와 오류, 기록할 필드, 요청 수용 상태와 종료 수명을 같은 계약으로 지키는지가 중요합니다.

모든 명령은 **저장소 루트**에서 실행합니다. 실제 사용자 자료나 외부 인증 계정을 사용하지 않습니다. 기본 실행은 모두 자동으로 끝나며, 마지막 프로그램은 `-serve`를 명시했을 때만 기다립니다.

## 문제 1. 시작 단계가 끝나지 않았다

서버가 실행 중이라고 말할 수 없습니다. `APP_READ_TIMEOUT`의 파싱이 실패했고 listener를 아직 만들지 않았습니다. 5초 기본값을 적용하는 상황은 해당 변수가 없는 경우입니다. 명시한 잘못된 값까지 기본값으로 숨기는 정책은 이번 문제의 요구사항과 다릅니다.

HTTP 500은 HTTP 요청에 대한 응답입니다. 아직 서버가 시작되지 않았으므로 이 실패는 프로세스의 표준 오류와 종료 코드에서 확인해야 합니다. 같은 오류라도 발생한 단계가 다르면 관찰하는 방법이 달라집니다.

포트를 먼저 열면 설정이 유효하지 않은 상태에서도 외부에서 연결할 수 있는 순간이 생깁니다. 초기화가 덜 된 핸들러가 요청을 받거나 수신은 되는데 응답할 준비가 없는 상태가 될 수 있습니다. 필요한 설정을 먼저 검증하는 편이 상태를 단순하게 만듭니다.

| 기록하면 좋은 정보 | 그대로 기록하지 않을 정보 |
| --- | --- |
| 설정 항목 이름 `APP_READ_TIMEOUT` | 전체 환경 변수 목록 |
| 실패 단계 `configuration` | 비밀번호·토큰 원문 |
| 요구되는 시간 형식과 범위 | 인증 헤더 전체 |
| 성공적으로 검증한 공개 주소 | 비밀값이 섞일 수 있는 DSN 전체 |

두 요청 로그 중 패턴 기반 로그는 항목 번호와 쿼리를 제외하고 메서드·라우트·결과를 유지합니다. 여러 항목에 대한 요청을 같은 종류로 집계하기 쉽습니다. 다만 실제 문제를 진단하려면 요청 ID 같은 다른 안전한 상관관계 정보가 필요할 수 있습니다. “모든 로그에서 값을 전부 삭제”하는 규칙이 아니라 어떤 질문에 답할 필드가 필요한지 정하는 작업입니다.

## 문제 2. 검증한 환경 위에 플래그를 적용한다

저장 경로: [`config-precedence/main.go`](config-precedence/main.go).

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
)

func limitConfig(args []string, lookup func(string) (string, bool)) (int, error) {
	limit := 20
	if value, ok := lookup("APP_LIMIT"); ok {
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 100 {
			return 0, errors.New("APP_LIMIT은 1~100 정수여야 합니다")
		}
		limit = n
	}
	flags := flag.NewFlagSet("limit", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.IntVar(&limit, "limit", limit, "페이지 크기")
	if err := flags.Parse(args); err != nil {
		return 0, err
	}
	if flags.NArg() != 0 {
		return 0, errors.New("위치 인자는 받지 않습니다")
	}
	if limit < 1 || limit > 100 {
		return 0, errors.New("limit는 1~100이어야 합니다")
	}
	return limit, nil
}
func main() {
	absent := func(string) (string, bool) { return "", false }
	env := func(string) (string, bool) { return "30", true }
	for _, tc := range []struct {
		args   []string
		lookup func(string) (string, bool)
	}{{nil, absent}, {nil, env}, {[]string{"-limit", "40"}, env}} {
		n, err := limitConfig(tc.args, tc.lookup)
		if err != nil {
			fmt.Println("오류:", err)
			return
		}
		fmt.Println("크기:", n)
	}
	_, err := limitConfig([]string{"-limit", "0"}, env)
	fmt.Println("범위 오류:", err != nil)
}
```

```sh
go run ./solutions/25-service-operations/config-precedence
```

```text
크기: 20
크기: 30
크기: 40
범위 오류: true
```

검증 코드: [`config-precedence/behavior_test.go`](config-precedence/behavior_test.go).

```go
package main

import "testing"

func TestPrecedence(t *testing.T) {
	env := func(string) (string, bool) { return "30", true }
	for _, tc := range []struct {
		args []string
		want int
	}{{nil, 30}, {[]string{"-limit=1"}, 1}, {[]string{"-limit=100"}, 100}, {[]string{"-limit=40"}, 40}} {
		got, err := limitConfig(tc.args, env)
		if err != nil || got != tc.want {
			t.Fatalf("got=%d error=%v", got, err)
		}
	}
	for _, args := range [][]string{{"-limit=0"}, {"-limit=-1"}, {"-limit=101"}, {"-limit=x"}, {"-unknown"}, {"extra"}} {
		if _, err := limitConfig(args, env); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if _, err := limitConfig([]string{"-limit=20"}, func(string) (string, bool) { return "bad", true }); err == nil {
		t.Fatal("invalid environment silently replaced")
	}
}
```


### 읽는 순서

기본 20을 만든 뒤 `lookup`으로 환경 존재 여부를 확인합니다. 환경이 있으면 바로 정수 파싱과 범위 검증을 마칩니다. 이 단계에서 실패하면 뒤에 정상 플래그가 있더라도 오류입니다. 설정 오류를 다른 경로로 가리지 않겠다는 문제의 계약을 따릅니다.

플래그에는 현재 `limit`을 기본값으로 등록합니다. 그래서 환경 30이 있고 플래그가 없으면 30, 플래그 40이 있으면 40이 됩니다. 값을 각각 따로 저장했다가 마지막에 복잡한 존재 여부 조건으로 합칠 필요가 없습니다.

`flag.NewFlagSet`은 이 호출의 파싱 상태만 보관합니다. 테스트의 `-test.*` 인자나 다른 예제의 플래그를 읽지 않습니다. `ContinueOnError`는 오류를 반환하게 하므로 함수가 프로세스를 직접 끝내지 않습니다.

### 검사할 경계

| 검사 | 잡으려는 오류 |
| --- | --- |
| 환경 없음 | 기본값을 초기화하지 않음 |
| 환경만 있음 | 환경을 읽었지만 사용하지 않음 |
| 플래그 덮어쓰기 | 우선순위 반대로 적용 |
| 1·100 | 올바른 양끝을 거절 |
| 0·101 | 파싱 성공만 확인 |
| bad 환경 + 정상 플래그 | 잘못된 환경을 조용히 숨김 |
| 위치 인자 | 오타를 무시하고 실행 |

대안으로 설정 출처까지 구조체에 저장할 수 있습니다. 진단이 필요한 서비스라면 `limit`의 값과 출처가 기본값·환경·플래그 중 무엇인지 기록할 수 있습니다. 그 경우에도 비밀값의 원문을 함께 기록하지 않습니다.

## 문제 3. 취소된 부모에서는 정리 시간을 만들 수 없다

`runContext.Done()`에서 돌아왔다면 해당 context는 취소되었거나 제한 시간이 지났습니다. 그 자식에 3초를 추가해도 부모 취소가 전파되므로 이미 취소된 상태입니다. “현재부터 3초”라는 새 기회가 생기는 것이 아닙니다.

수정할 핵심 부분은 다음과 같습니다. **발췌**이며 전체 동작은 [service.Serve](../../examples/25-service-operations/service/lifecycle.go)에 있습니다.

```go
drain, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
shutdownErr := server.Shutdown(drain)
```

별도의 부모를 쓰는 것은 실행 취소를 무시하고 계속 서비스하겠다는 의미가 아닙니다. 새 요청 수신은 중지하되, 이미 받은 요청을 끝내기 위한 별도 수명을 부여하는 것입니다. 종료 대기의 상한이 지나면 `Close`를 시도하고 시간 초과 오류를 보존합니다.

`Serve`가 `ErrServerClosed`로 끝난 것은 수신 루프가 닫혔다는 의미입니다. 활성 요청이 모두 끝났다는 증거는 아닐 수 있습니다. 따라서 `Shutdown`의 결과를 먼저 기다리고 `Serve`의 결과도 수집합니다.

### 실패를 유지하는 이유

종료 제한을 넘겼는데 “정상 종료 완료”만 기록하면 배포 도중 응답이 끊긴 원인을 찾기 어렵습니다. `errors.Join`으로 종료 제한 오류와 강제 닫기 오류를 함께 보존하면 `errors.Is(err, context.DeadlineExceeded)`로 원인을 확인할 수 있습니다.

호출자가 더 이상 복구할 수 없더라도 오류의 의미를 보존하는 일은 필요합니다. 정상 종료와 제한을 넘긴 종료를 구분하는 것은 다음 실행의 기능과 별개로 운영 결과를 설명하기 위한 계약입니다.

## 문제 4. 새 필드가 자동으로 로그에 나오지 않게 한다

저장 경로: [`redacted-account/main.go`](redacted-account/main.go).

```go
package main

import (
	"log/slog"
	"os"
)

type Account struct{ ID, Email, Password string }

func (a Account) LogValue() slog.Value { return slog.GroupValue(slog.String("id", a.ID)) }
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if len(groups) == 0 && a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))
	logger.Info("account", "account", Account{ID: "student-7", Email: "demo@example.invalid", Password: "placeholder"})
}
```

```sh
go run ./solutions/25-service-operations/redacted-account
```

```text
{"level":"INFO","msg":"account","account":{"id":"student-7"}}
```

검증 코드: [`redacted-account/behavior_test.go`](redacted-account/behavior_test.go).

```go
package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestAllowList(t *testing.T) {
	var out bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&out, nil))
	l.Info("account", "account", Account{"id-1", "private@example.invalid", "private-pass"})
	if strings.Contains(out.String(), "private") {
		t.Fatal("private value leaked")
	}
	var got struct {
		Account map[string]string `json:"account"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Account) != 1 || got.Account["id"] != "id-1" {
		t.Fatalf("fields=%v", got.Account)
	}
}
```


### 허용 목록의 장점

`LogValue`는 `ID`만 고릅니다. 구조체에 새 비밀 필드가 추가되어도 이 메서드를 수정하지 않는 한 로그에 자동으로 들어오지 않습니다. 구조체 전체를 출력한 뒤 특정 단어를 지우는 방식보다 출력의 범위를 읽기 쉽습니다.

테스트는 두 방향으로 확인합니다. JSON을 파싱해서 허용한 필드가 있는지 확인하고, 출력 바이트에서 비밀값 원문이 없는지 확인합니다. “값이 가려졌다”만 확인하면 필요한 ID까지 사라지는 구현을 놓칠 수 있습니다. 반대로 ID만 확인하면 비밀 필드가 추가로 남는 구현을 놓칠 수 있습니다.

이메일 키를 남기되 별표로 치환하는 설계도 가능합니다. 문제의 계약은 아예 키도 넣지 않는 것이므로 이번 해설은 한 필드만 포함합니다. 요구되는 관찰 정보가 달라졌을 때 바꾸어야 할 결정입니다.

### 주의할 출력 경로

`LogValuer`가 있어도 `fmt.Printf("%+v", account)`나 문자열로 만든 오류에는 원래 필드가 들어갈 수 있습니다. 로거 사용법과 타입을 함께 검토해야 합니다. 비밀값을 노출하는 실패 예제를 자동 테스트 출력에 실제로 남길 필요는 없습니다.

## 문제 5. 새 요청과 이미 받은 요청을 구분한다

저장 경로: [`draining-gate/main.go`](draining-gate/main.go).

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
)

func gate(accepting *atomic.Bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !accepting.Load() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "draining", 503)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	var accepting atomic.Bool
	accepting.Store(true)
	handler := gate(&accepting, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	for _, state := range []bool{true, false} {
		accepting.Store(state)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("POST", "/work", nil))
		fmt.Println("상태:", w.Code)
	}
}
```

```sh
go run ./solutions/25-service-operations/draining-gate
```

```text
상태: 204
상태: 503
```

검증 코드: [`draining-gate/behavior_test.go`](draining-gate/behavior_test.go).

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAlreadyAcceptedWorkFinishes(t *testing.T) {
	var accepting atomic.Bool
	accepting.Store(true)
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	handler := gate(&accepting, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-release; w.WriteHeader(204) }))
	first := httptest.NewRecorder()
	go func() { defer close(done); handler.ServeHTTP(first, httptest.NewRequest("POST", "/work", nil)) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("not entered")
	}
	accepting.Store(false)
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest("POST", "/work", nil))
	close(release)
	<-done
	if first.Code != 204 || second.Code != 503 || second.Header().Get("Retry-After") != "1" {
		t.Fatalf("first=%d second=%d", first.Code, second.Code)
	}
}
```


### 게이트의 계약

`Load`에서 참을 본 요청은 다음 핸들러로 들어갑니다. 그 직후 상태가 거짓으로 바뀌더라도 이미 통과한 요청입니다. 이 계약은 “상태를 false로 바꾼 시각 이후에는 어떤 실행도 시작되지 않는다”는 엄격한 동기화 장치와 다릅니다.

더 강한 경계가 필요하다면 상태 확인과 작업 등록을 같은 잠금 아래 묶고, 등록된 작업 수의 감소를 기다리는 구조를 설계할 수 있습니다. 이 문제는 HTTP 서버의 종료 대기와 결합할 가벼운 수용 게이트를 만드는 범위입니다.

### 테스트의 사건 순서

첫 핸들러가 `entered`를 닫기 전에는 accepting을 바꾸지 않습니다. 테스트가 첫 요청이 실제로 통과했음을 관찰한 뒤 상태를 바꾸어야 “기존 요청”과 “새 요청”을 확실히 나눌 수 있기 때문입니다.

두 번째 요청은 기다리는 작업을 추가하지 않고 바로 503으로 끝납니다. 그다음 `release`를 닫아 첫 요청을 끝내고 응답을 확인합니다. 상태 변경이 기존 요청을 자동으로 취소하지 않았는지 검증합니다.

`Retry-After: 1`은 이 예제의 안내값입니다. 실제 호출자는 모든 POST를 자동 재시도하면 안 됩니다. 요청이 실제 처리됐는지, 중복 실행이 안전한지, 재시도 정책이 있는지는 별도 계약입니다. 게이트는 자체적으로 재시도를 실행하지 않습니다.

## 문제 6. 입력과 수명을 주입한 실행 함수

저장 경로: [`managed-server/main.go`](managed-server/main.go).

```go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/25-service-operations/service"
)

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, "안녕하세요\n"); err != nil {
			return
		}
	})
	return mux
}
func run(ctx context.Context, args []string, out, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("managed-server", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	serve := flags.Bool("serve", false, "직접 HTTP 요청을 받습니다")
	address := flags.String("addr", "127.0.0.1:8080", "loopback 주소")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("위치 인자는 받지 않습니다")
	}
	host, _, err := net.SplitHostPort(*address)
	if err != nil {
		return errors.New("주소는 호스트:포트 형식이어야 합니다")
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("학습 서버는 loopback 주소만 사용합니다")
	}
	if !*serve {
		w := httptest.NewRecorder()
		handler().ServeHTTP(w, httptest.NewRequest("GET", "/hello", nil))
		_, err := fmt.Fprintf(out, "상태: %d\n%s", w.Code, w.Body.String())
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(diagnostics, nil))
	server := &http.Server{Handler: handler(), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	logger.Info("listening", "address", listener.Addr().String())
	err = service.Serve(ctx, server, listener, 3*time.Second)
	if err != nil {
		logger.Error("shutdown failed", "error", err)
		return err
	}
	logger.Info("stopped")
	return nil
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./solutions/25-service-operations/managed-server
```

```text
상태: 200
안녕하세요
```

검증 코드: [`managed-server/behavior_test.go`](managed-server/behavior_test.go).

```go
package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

func TestDemo(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), nil, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
	if out.String() != "상태: 200\n안녕하세요\n" {
		t.Fatalf("output=%q", out.String())
	}
}
func TestArguments(t *testing.T) {
	for _, args := range [][]string{{"-addr=0.0.0.0:8080"}, {"-addr=bad"}, {"extra"}, {"-unknown"}} {
		if err := run(context.Background(), args, io.Discard, io.Discard); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
func TestCanceledBeforeListen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := run(ctx, []string{"-serve", "-addr=127.0.0.1:0"}, io.Discard, io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
```


### 기본 모드와 직접 실행 모드

기본 모드는 HTTP 핸들러만 실행합니다. 운영체제 신호나 실제 포트에 의존하지 않으므로 결과를 빠르게 비교할 수 있습니다. `-serve` 모드는 listener를 만들고 `service.Serve`가 끝날 때까지 대기합니다.

함수 인자는 프로세스 인자·표준 출력·표준 오류를 직접 참조하지 않습니다. `main`이 `os.Args[1:]`, `os.Stdout`, `os.Stderr`를 전달합니다. 테스트는 빈 인자와 버퍼를 전달하므로 실제 시험 실행기의 플래그를 잘못 파싱하지 않습니다.

| 책임 | 위치 |
| --- | --- |
| 운영체제 종료 신호 등록 | `main` |
| 플래그와 주소 검사 | `run` |
| 라우트와 응답 내용 | `handler` |
| 서버 대기·정상 종료·강제 닫기 | `service.Serve` |
| 최종 프로세스 종료 코드 | `main` |

`signal.NotifyContext`의 `stop`은 `run`에서 돌아온 뒤 호출합니다. 오류가 났을 때도 신호 등록을 정리한 다음 `os.Exit(1)`을 실행합니다. `main`에서 `defer stop()`만 하고 곧바로 `os.Exit`하는 구조라면 defer가 실행되지 않는 점을 놓치기 쉽습니다.

### 직접 실행하기

```sh
go run ./solutions/25-service-operations/managed-server -serve -addr=127.0.0.1:8080
```

다른 터미널에서 요청합니다. Windows PowerShell에서는 다음 명령을 사용할 수 있습니다.

```powershell
Invoke-WebRequest -Uri "http://127.0.0.1:8080/hello"
```

macOS·Linux·WSL에서 curl을 사용할 수 있다면 다음과 같습니다.

```sh
curl -i http://127.0.0.1:8080/hello
```

예상 상태는 200이고 본문은 `안녕하세요`와 줄바꿈입니다. 응답 날짜와 로그 시각은 실제 실행마다 달라집니다. 첫 터미널에서 Ctrl+C를 눌러 종료합니다.

이 실습의 서버는 기본적으로 loopback에서만 접근하도록 주소를 검증합니다. 공개 서비스의 인증·전송 암호화·사용자별 권한을 완성한 프로그램은 아닙니다. 마지막 종합 실습에서도 API 기능과 공개 운영의 요건을 구분합니다.

## 선택 문제 해설. 래퍼가 보장하는 범위

| 경우 | 판단 | 이유 |
| --- | --- | --- |
| Write만 사용하는 일반 텍스트 | 지원 | 첫 Write에서 200을 기록 |
| 최종 상태 두 번 설정 | 첫 상태 유지 | 이미 보낸 상태를 뒤에서 바꿀 수 없음 |
| 아무것도 쓰지 않음 | 지원 | 정상 반환 시 기본 200으로 기록 |
| Flusher 직접 타입 단언 | 추가 설계 필요 | 래퍼가 해당 인터페이스를 자동 구현하지 않음 |
| Hijacker로 연결 가져오기 | 추가 설계 필요 | 일반 요청·응답과 다른 자원 수명 |

`Unwrap`은 ResponseController가 내부 작성자를 찾을 수 있게 합니다. 모든 기존 인터페이스 검사 코드가 자동으로 래퍼를 통과한다는 뜻은 아닙니다. 실제 사용하는 기능을 확인하고 지원 범위를 문서화합니다.

## 검증 결과를 읽는 방법

```sh
go test -count=1 ./examples/25-service-operations/... ./solutions/25-service-operations/...
go vet ./examples/25-service-operations/... ./solutions/25-service-operations/...
go test -race ./examples/25-service-operations/... ./solutions/25-service-operations/...
```

검사는 외부 서비스 없이 설정 경계, 로그 필드, 인증 호출 경계와 실제 로컬 HTTP 종료를 확인합니다. 레이스 검사는 지원 환경에서 실행합니다. 직접 운영체제 신호를 보내는 과정은 모든 단위 테스트에 섞지 않고 수명 관리 함수에 취소 context를 전달해 종료 동작을 검사합니다.

[문제로 돌아가기](../../exercises/25-service-operations/README.md) · [본문으로 돌아가기](../../book/25-service-operations/README.md)
