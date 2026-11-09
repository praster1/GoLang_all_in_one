# 30. 종합 실습: 할 일 HTTP API

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 종합 | [09장 인터페이스](../09-methods-interfaces/README.md), [20장 동기화](../20-sync/README.md), [22장 서버](../22-http-server/README.md), [23장 API 테스트](../23-api-testing/README.md), [24장 SQL](../24-sql-transactions/README.md), [25장 운영 기초](../25-service-operations/README.md) | Go, 이 저장소, 캐시에 준비한 SQLite 드라이버 |

할 일을 추가하고, 목록을 읽고, 완료 상태를 바꾸고, 삭제하는 API를 만듭니다. 이번 장의 중심은 화면에 JSON을 출력하는 데 있지 않습니다. **HTTP 처리와 저장 방식을 분리하고, 저장소를 바꾸어도 같은 동작을 유지하는 것**이 목표입니다.

메모리 구현에서는 맵과 잠금을 사용하고, SQLite 구현에서는 실제 SQL을 실행합니다. 두 구현에 같은 테스트를 적용합니다. 마지막에는 플래그로 저장소를 선택하는 서버를 실행하며, 자동 데모는 사용자 입력 없이 결과를 출력하고 끝납니다.

## 학습 목표

- 사용자 입력 규칙과 저장소 계약을 Go 타입으로 표현한다.
- 메모리와 SQLite 구현을 같은 인터페이스로 교체한다.
- 정상 응답뿐 아니라 잘못된 입력·없는 항목·내부 실패를 HTTP 계약으로 만든다.
- 여러 페이지의 순서, 취소, 동시 요청과 영속 저장을 검증한다.
- 실제 서버의 설정·요청 수명·정상 종료와 저장소 정리를 연결한다.

## 준비와 빠른 실행

모든 명령은 **저장소 루트**에서 실행합니다. 하위 `go.mod`를 만들지 않습니다. SQLite 드라이버는 [24장](../24-sql-transactions/README.md)에서 도입한 `modernc.org/sqlite v1.62.0`을 그대로 사용합니다. 실제 은행 시스템이나 외부 DB 계정에 연결하지 않습니다.

```sh
go mod download
go run ./examples/30-todo-api/http-demo
go test ./examples/30-todo-api/todo
```

첫 명령은 의존성이 캐시에 없을 때 네트워크가 필요할 수 있습니다. 예제와 테스트는 준비된 드라이버를 사용해 로컬 메모리나 임시 DB에서 실행합니다.

| 경로 | 역할 | 자동 종료 |
| --- | --- | --- |
| [todo/model.go](../../examples/30-todo-api/todo/model.go) | 입력 규칙·오류·저장소 계약 | 라이브러리 |
| [todo/memory.go](../../examples/30-todo-api/todo/memory.go) | 메모리 저장소 | 라이브러리 |
| [todo/sqlite.go](../../examples/30-todo-api/todo/sqlite.go) | 실제 SQLite 저장소 | 라이브러리 |
| [todo/http.go](../../examples/30-todo-api/todo/http.go) | JSON API와 오류 응답 | 라이브러리 |
| [domain-demo](../../examples/30-todo-api/domain-demo/main.go) | 제목 정규화와 입력 오류 | 예 |
| [memory-demo](../../examples/30-todo-api/memory-demo/main.go) | 저장소 생성·수정·삭제 | 예 |
| [sqlite-demo](../../examples/30-todo-api/sqlite-demo/main.go) | 파일 DB를 닫고 다시 열기 | 예 |
| [pagination-demo](../../examples/30-todo-api/pagination-demo/main.go) | ID 기준 페이지 순회 | 예 |
| [http-demo](../../examples/30-todo-api/http-demo/main.go) | 핸들러에 CRUD 요청 전달 | 예 |
| [client-demo](../../examples/30-todo-api/client-demo/main.go) | 실제 로컬 HTTP 왕복 | 예 |
| [server](../../examples/30-todo-api/server/main.go) | 플래그로 선택하는 서버 | Ctrl+C까지 실행 |

`go run ./examples/30-todo-api/todo`는 실행 명령이 아닙니다. `todo`는 다른 프로그램이 import하는 패키지입니다. 위 표의 데모 중 하나를 실행합니다.

## 1. 먼저 API 계약을 정한다

할 일 하나는 양수 ID, 제목, 완료 여부를 가집니다. 같은 제목의 할 일을 여러 번 등록할 수 있습니다. 제목의 고유성이 업무 규칙이 아니기 때문에 기본 스키마에 `UNIQUE`를 넣지 않습니다. 연습문제의 일괄 가져오기는 별도 규칙을 사용하며 그 차이를 명시합니다.

| 항목 | 계약 |
| --- | --- |
| ID | 양수 `int64`, 저장소가 발급 |
| 제목 | 올바른 UTF-8, 앞뒤 공백 제거 후 1~80룬 |
| 제목 제어 문자 | 정규화 후 내부에 남은 제어 문자는 거절 |
| 완료 여부 | 생성 시 false, 전체 수정 시 true·false 모두 허용 |
| 목록 순서 | ID 오름차순 |
| 목록 범위 | 기본 20개, 한 번에 1~100개 |
| 삭제한 ID | 같은 저장소 수명에서 다시 발급하지 않음 |
| 메모리 수명 | 프로세스가 끝나면 데이터 소멸 |
| 파일 DB 수명 | 같은 파일을 다시 열면 데이터 유지 |

룬 수는 화면에 보이는 글자 묶음 수와 항상 같지 않습니다. 예를 들어 결합 문자나 이모지 조합은 여러 룬일 수 있습니다. 07장에서 배운 단위 차이를 그대로 적용합니다. 이번 규칙은 룬을 기준으로 한 명시적인 API 선택입니다.

### 1.1 HTTP 경로

| 메서드와 경로 | 입력 | 성공 응답 |
| --- | --- | --- |
| `POST /todos` | `{"title":"Go 복습"}` | 201, 생성한 항목, Location |
| `GET /todos` | 선택 쿼리 `after`, `limit` | 200, 항목 배열 |
| `GET /todos/{id}` | 양수 ID | 200, 항목 |
| `HEAD /todos` | GET과 같은 목록 쿼리 | GET과 같은 상태, 본문 없음 |
| `HEAD /todos/{id}` | 양수 ID | GET과 같은 상태, 본문 없음 |
| `PUT /todos/{id}` | `title`, `done` 둘 다 필수 | 200, 수정한 항목 |
| `DELETE /todos/{id}` | 양수 ID | 204, 본문 없음 |

PUT은 이번 API에서 전체 필드를 교체하는 요청입니다. `done`만 보낸 부분 수정을 지원한다고 가정하지 않습니다. PATCH, 부분 수정, 사용자별 권한을 추가하려면 별도의 계약과 테스트가 필요합니다.

### 1.2 실패 응답

실패 본문은 `{"error":"코드"}` 형식입니다. 내부 SQL 문장과 파일 경로를 그대로 보내지 않습니다.

| 상태 | 상황 | 대표 오류 코드 |
| --- | --- | --- |
| 400 | JSON 문법·필드·쿼리·ID 형식 오류 | `invalid_json`, `invalid_parameter` |
| 404 | 항목 또는 경로 없음 | `not_found` |
| 405 | 경로에서 허용하지 않는 메서드 | `method_not_allowed` |
| 413 | 본문 1,024바이트 초과 | `body_too_large` |
| 415 | JSON 이외의 미디어 타입 | `json_required` |
| 422 | 필수 필드 없음·제목 범위 오류 | `invalid_title`, `missing_field` |
| 500 | 분류하지 않은 저장소 실패 | `internal_error` |
| 503 | 작업 context가 취소됨 | `request_canceled` |
| 504 | 작업 context의 제한 시간 만료 | `deadline_exceeded` |

취소된 실제 클라이언트는 연결을 이미 닫았을 수 있으므로 503 본문을 반드시 받는다는 뜻은 아닙니다. 이는 핸들러가 응답을 작성할 수 있는 상황에서의 오류 매핑입니다. 모든 서비스가 취소를 같은 상태로 매핑해야 한다는 일반 규칙은 아닙니다.

## 2. HTTP가 모르는 저장소 계약

저장 경로: [`todo/model.go`](../../examples/30-todo-api/todo/model.go).

```go
// Package todo defines the task domain and interchangeable repositories.
package todo

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrNotFound = errors.New("할 일을 찾을 수 없습니다")
var ErrInvalidID = errors.New("ID는 양수여야 합니다")
var ErrInvalidPage = errors.New("after는 0 이상, limit는 1~100이어야 합니다")
var ErrIDExhausted = errors.New("새 ID를 발급할 수 없습니다")

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

func NormalizeTitle(title string) (string, error) {
	if !utf8.ValidString(title) {
		return "", &ValidationError{"title", "올바른 UTF-8이어야 합니다"}
	}
	title = strings.TrimSpace(title)
	if n := utf8.RuneCountInString(title); n < 1 || n > 80 {
		return "", &ValidationError{"title", "공백 제거 후 1~80룬이어야 합니다"}
	}
	for _, r := range title {
		if unicode.IsControl(r) {
			return "", &ValidationError{"title", "제어 문자를 포함할 수 없습니다"}
		}
	}
	return title, nil
}
func validatePage(after int64, limit int) error {
	if after < 0 || limit < 1 || limit > 100 {
		return ErrInvalidPage
	}
	return nil
}

// Repository returns tasks by value. List is sorted by ID, ascending.
// Create never reuses a previously issued ID during this repository's lifetime.
// Canceled contexts must be observed, but cancellation does not undo a completed write.
type Repository interface {
	Create(context.Context, string) (Task, error)
	Get(context.Context, int64) (Task, error)
	List(context.Context, int64, int) ([]Task, error)
	Replace(context.Context, int64, string, bool) (Task, error)
	Delete(context.Context, int64) error
}
```

### 2.1 실패의 종류를 먼저 정한다

`ErrNotFound`는 데이터가 없다는 업무 결과입니다. `ErrInvalidID`와 `ErrInvalidPage`는 호출 인자의 계약 위반입니다. `ValidationError`는 필드 이름을 가진 입력 오류입니다. 저장소 구현이 달라도 호출자가 같은 방식으로 분류할 수 있도록 정의했습니다.

오류 문자열이 같다는 것만으로 같은 계약이 되지는 않습니다. `errors.Is`와 `errors.As`로 분류할 수 있도록 래핑 관계와 타입을 유지합니다. SQLite의 `sql.ErrNoRows`는 저장소 안에서 `ErrNotFound`로 변환합니다. HTTP 코드가 특정 DB의 오류 타입을 직접 알아야 할 이유를 줄입니다.

### 2.2 반환값의 소유권

`Task`에는 문자열과 값 필드만 있습니다. 반환된 구조체를 호출자가 수정해도 저장소 안의 값은 바뀌지 않습니다. 목록도 새 슬라이스를 만들어 반환합니다. 저장소의 맵이나 내부 슬라이스를 그대로 노출하면 잠금을 사용하더라도 외부 수정으로 계약이 깨질 수 있습니다.

이 설계가 나중에 슬라이스·맵·포인터 필드를 추가해도 자동으로 유지되는 것은 아닙니다. 참조하는 필드가 생기면 05·08장에서 배운 깊이와 공유 범위를 다시 검토해야 합니다.

### 2.3 Close를 인터페이스에 넣지 않은 이유

핸들러는 생성·조회·수정·삭제를 사용하며 저장소 전체를 닫지 않습니다. 메모리 저장소는 외부 자원이 없고 SQLite는 DB 핸들을 닫아야 합니다. 따라서 시작 코드가 구체적인 SQLite를 만들고 종료 책임을 보관합니다.

인터페이스에는 사용자가 필요한 행동을 넣습니다. 두 구현을 똑같아 보이게 하려고 메모리에 의미 없는 `Close`를 추가할 필요는 없습니다. 30장의 서버 구성에서 이 책임을 연결합니다.

## 3. 제목 정규화와 입력 경계

저장 경로: [`domain-demo/main.go`](../../examples/30-todo-api/domain-demo/main.go).

```go
package main

import (
	"errors"
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func main() {
	title, err := todo.NormalizeTitle("  Go 복습  ")
	fmt.Println("제목:", title, "성공:", err == nil)
	_, err = todo.NormalizeTitle(" \t ")
	var validation *todo.ValidationError
	fmt.Println("입력 오류:", errors.As(err, &validation))
	if validation != nil {
		fmt.Println("필드:", validation.Field)
	}
}
```

```sh
go run ./examples/30-todo-api/domain-demo
```

실행 결과:

```text
제목: Go 복습 성공: true
입력 오류: true
필드: title
```


정규화 함수는 먼저 바이트가 올바른 UTF-8인지 검사합니다. 그 뒤 앞뒤 공백을 제거하고 룬 수와 제어 문자를 검사합니다. 순서를 바꾸면 공백 제거 후 빈 제목이 되었는데 허용하거나, 잘못된 UTF-8이 다른 문자로 바뀐 뒤 통과할 수 있습니다.

다음 예를 직접 분류해 봅니다.

| 입력 | 결과 |
| --- | --- |
| `"  Go 복습  "` | `"Go 복습"` |
| 빈 문자열 또는 공백만 | 오류 |
| 한글 80룬 | 허용 |
| 한글 81룬 | 오류 |
| `"줄\n바꿈"` | 내부 제어 문자로 오류 |
| 정상적인 `�` 문자를 포함 | 허용 |
| UTF-8로 해석할 수 없는 바이트 | 오류 |

UTF-8 오류와 정상적인 대체 문자 U+FFFD는 다릅니다. 07장에서 확인한 것처럼 문자 모양만 보고 원래 바이트가 깨졌다고 판단하지 않습니다. 정규화 결과가 오류이면 저장 전에 반환하므로 ID를 새로 발급하거나 기존 항목을 변경하지 않습니다.

## 4. 메모리 저장소를 만든다

저장 경로: [`todo/memory.go`](../../examples/30-todo-api/todo/memory.go).

```go
package todo

import (
	"context"
	"math"
	"slices"
	"sync"
)

type Memory struct {
	mu     sync.Mutex
	nextID int64
	tasks  map[int64]Task
}

func NewMemory() *Memory { return &Memory{nextID: 1, tasks: make(map[int64]Task)} }

var _ Repository = (*Memory)(nil)

func (m *Memory) Create(ctx context.Context, title string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	title, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if m.nextID <= 0 {
		return Task{}, ErrIDExhausted
	}
	task := Task{ID: m.nextID, Title: title}
	if m.nextID == math.MaxInt64 {
		m.nextID = 0
	} else {
		m.nextID++
	}
	m.tasks[task.ID] = task
	return task, nil
}
func (m *Memory) Get(ctx context.Context, id int64) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if id <= 0 {
		return Task{}, ErrInvalidID
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	task, ok := m.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return task, nil
}
func (m *Memory) List(ctx context.Context, after int64, limit int) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validatePage(after, limit); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(m.tasks))
	for id := range m.tasks {
		if id > after {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	if len(ids) > limit {
		ids = ids[:limit]
	}
	result := make([]Task, 0, len(ids))
	for _, id := range ids {
		result = append(result, m.tasks[id])
	}
	return result, nil
}
func (m *Memory) Replace(ctx context.Context, id int64, title string, done bool) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if id <= 0 {
		return Task{}, ErrInvalidID
	}
	title, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if _, ok := m.tasks[id]; !ok {
		return Task{}, ErrNotFound
	}
	task := Task{ID: id, Title: title, Done: done}
	m.tasks[id] = task
	return task, nil
}
func (m *Memory) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id <= 0 {
		return ErrInvalidID
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := m.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}
```

### 4.1 ID 발급과 저장은 같은 잠금 안에 둔다

고루틴 두 개가 같은 nextID를 읽고 각각 저장하면 항목이 덮어써질 수 있습니다. `Create`는 ID 확인, 증가, 맵 저장을 하나의 잠금 범위로 묶습니다. 원자적 숫자 하나만 사용한다고 맵 변경과의 관계까지 자동으로 안전해지는 것은 아닙니다.

이 예제는 읽기·쓰기 모두 하나의 `Mutex`를 사용합니다. 코드가 짧고 상태 관계를 설명하기 쉽기 때문입니다. 읽기 전용 경로에 `RWMutex`를 적용할 수 있지만 언제나 더 빠르다고 가정하지 않습니다. 26장에서 배운 측정 결과와 실제 접근 패턴이 필요합니다.

### 4.2 context 확인은 잠금 전후에 있다

이미 취소된 호출은 잠금을 얻기 전에 반환합니다. 잠금을 기다리는 동안 취소되었을 수도 있으므로 얻은 뒤 다시 확인합니다. `Mutex.Lock` 자체가 context를 받아 취소 가능한 대기를 제공하는 것은 아닙니다.

잠금 안에서 느린 네트워크 호출을 하지 않습니다. 이 저장소는 작은 학습 데이터의 복사·정렬에 범위를 둡니다. 데이터가 많아지면 목록을 만들기 위해 전체 맵을 훑고 정렬하는 비용도 커집니다. 반환 개수 100개 제한이 모든 내부 작업을 100개로 제한한다는 뜻은 아닙니다.

### 4.3 목록은 정렬한 뒤 자른다

맵 순회 순서는 보장되지 않습니다. 페이지 결과가 매번 달라지면 다음 페이지를 찾는 규칙도 불안정해집니다. 먼저 after보다 큰 ID를 모으고 오름차순으로 정렬한 다음 limit만큼 자릅니다.

결과가 없을 때도 빈 슬라이스를 반환합니다. JSON으로 변환할 때 `null`이 아닌 `[]`가 되어 목록이라는 계약을 유지합니다. nil 슬라이스와 빈 슬라이스가 Go 안에서 비슷하게 동작하는 지점이 있어도 외부 직렬화 결과까지 같지는 않습니다.

### 4.4 ID의 한계도 정의한다

최대 int64를 넘기면 음수로 되돌아가는 방식으로 두지 않습니다. 마지막 ID를 발급한 뒤 nextID를 종료 표시로 바꾸고 이후 생성은 `ErrIDExhausted`를 반환합니다. 정상적으로 쓰는 작은 교재 데이터에서 만나기 어렵지만 ID가 양수라는 불변 조건을 보존합니다.

메모리 저장소의 수명이 끝나면 ID 발급 이력도 끝납니다. 프로그램을 새로 시작하면 다시 1부터 시작합니다. 여러 프로세스 전체에서 고유한 ID라는 뜻이 아닙니다.

## 5. 저장소를 직접 실행한다

저장 경로: [`memory-demo/main.go`](../../examples/30-todo-api/memory-demo/main.go).

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	ctx := context.Background()
	repository := todo.NewMemory()
	first, err := repository.Create(ctx, "Go 복습")
	if err != nil {
		return err
	}
	second, err := repository.Create(ctx, "SQL 복습")
	if err != nil {
		return err
	}
	first, err = repository.Replace(ctx, first.ID, first.Title, true)
	if err != nil {
		return err
	}
	tasks, err := repository.List(ctx, 0, 20)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		fmt.Printf("%d %s 완료=%t\n", task.ID, task.Title, task.Done)
	}
	if err := repository.Delete(ctx, second.ID); err != nil {
		return err
	}
	_, err = repository.Get(ctx, second.ID)
	fmt.Println("삭제 확인:", errors.Is(err, todo.ErrNotFound))
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
go run ./examples/30-todo-api/memory-demo
```

실행 결과:

```text
1 Go 복습 완료=true
2 SQL 복습 완료=false
삭제 확인: true
```


HTTP 없이 저장소 기능부터 실행합니다. 이 단계에서 제목 정규화·완료 상태·삭제를 검증하면 나중에 HTTP 테스트가 실패했을 때 어느 층의 문제인지 나누기 쉽습니다.

`first.Title = "다른 값"`처럼 반환한 구조체만 바꾸어 보고 다시 Get했을 때 저장 내용이 유지되는지도 확인해 봅니다. 실제 수정은 `Replace`를 호출해야 합니다. 이것은 호출자에게 내부 저장소 변경 권한을 직접 넘기지 않는 계약입니다.

## 6. SQLite 구현으로 바꾸기

저장 경로: [`todo/sqlite.go`](../../examples/30-todo-api/todo/sqlite.go).

```go
package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type SQLite struct{ db *sql.DB }

var _ Repository = (*SQLite)(nil)

func OpenSQLite(ctx context.Context, dsn string) (*SQLite, error) {
	if dsn == "" {
		return nil, errors.New("SQLite 경로가 필요합니다")
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// :memory: is per physical connection. The tutorial keeps a single connection.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL CHECK(length(title) BETWEEN 1 AND 80),
        done INTEGER NOT NULL DEFAULT 0 CHECK(done IN (0,1))
    )`)
	if err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return &SQLite{db: db}, nil
}
func (s *SQLite) Close() error                   { return s.db.Close() }
func (s *SQLite) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var task Task
	if err := row.Scan(&task.ID, &task.Title, &task.Done); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, err
	}
	return task, nil
}
func (s *SQLite) Create(ctx context.Context, title string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	title, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}
	row := s.db.QueryRowContext(ctx, "INSERT INTO todos(title) VALUES (?) RETURNING id,title,done", title)
	task, err := scanTask(row)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}
func (s *SQLite) Get(ctx context.Context, id int64) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if id <= 0 {
		return Task{}, ErrInvalidID
	}
	return scanTask(s.db.QueryRowContext(ctx, "SELECT id,title,done FROM todos WHERE id=?", id))
}
func (s *SQLite) List(ctx context.Context, after int64, limit int) (result []Task, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validatePage(after, limit); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, "SELECT id,title,done FROM todos WHERE id>? ORDER BY id LIMIT ?", after, limit)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	result = make([]Task, 0)
	for rows.Next() {
		task, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
func (s *SQLite) Replace(ctx context.Context, id int64, title string, done bool) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if id <= 0 {
		return Task{}, ErrInvalidID
	}
	title, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}
	return scanTask(s.db.QueryRowContext(ctx, "UPDATE todos SET title=?,done=? WHERE id=? RETURNING id,title,done", title, done, id))
}
func (s *SQLite) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id <= 0 {
		return ErrInvalidID
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM todos WHERE id=?", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
```

### 6.1 DB를 여는 책임

`OpenSQLite`는 DB 핸들을 열고 연결을 확인하며 표를 준비합니다. 중간에 실패하면 이미 연 DB도 닫습니다. 성공한 호출자가 나중에 `Close`해야 합니다.

연결 수를 1로 제한한 이유는 `:memory:` 테스트에서도 같은 물리 연결의 DB를 사용하기 위해서입니다. SQLite의 메모리 DB는 기본적으로 연결마다 별개입니다. [SQLite 메모리 DB](https://sqlite.org/inmemorydb.html)의 계약을 참고합니다. 파일 DB에 더 큰 풀을 설정하려면 잠금 경합과 쓰기 패턴을 따로 측정해야 합니다.

### 6.2 같은 업무 규칙을 유지한다

메모리와 SQLite 모두 `NormalizeTitle`을 사용합니다. HTTP에서만 검증하면 다른 CLI나 배치가 저장소를 직접 사용할 때 잘못된 제목을 저장할 수 있습니다. 공통 입력 규칙을 저장소 경계에서도 지킵니다.

DB의 CHECK는 NOT NULL·길이·완료 값 같은 일부 조건을 추가로 제한합니다. 모든 Unicode 정규화 규칙을 SQL로 복제한 것은 아닙니다. 이 교재의 API를 통해 접근할 때 같은 도메인 함수를 사용하며, 외부에서 DB 파일을 직접 수정하는 상황은 별도의 접근 통제와 운영 범위입니다.

### 6.3 값을 SQL 문장에 붙이지 않는다

제목, ID와 limit는 `?` 자리에 별도 인자로 전달합니다. 따옴표나 SQL처럼 보이는 문자열도 값으로 저장됩니다. 문자열 연결로 SQL을 만들지 않는 이유는 [Go의 SQL 매개변수 안내](https://go.dev/doc/database/sql-injection)에서 확인합니다.

`ORDER BY id`와 표 이름은 프로그램에 고정되어 있습니다. 값 매개변수로 열 이름·정렬 방향까지 임의로 바꿀 수 있다고 생각하지 않습니다. 정렬 선택 기능을 추가한다면 허용 목록과 SQL 구성 경계를 별도로 정해야 합니다.

### 6.4 RETURNING으로 변경 결과를 받는다

생성과 수정은 SQL의 `RETURNING id,title,done`으로 변경된 행을 바로 받습니다. UPDATE 뒤에 별도 SELECT를 하면 그 사이 다른 요청이 값을 바꿀 수 있습니다. 한 문장의 결과를 받으면 이번 변경과 반환값을 연결하기 쉽습니다.

[SQLite RETURNING](https://sqlite.org/lang_returning.html)은 SQLite가 제공하는 확장 문법입니다. 모든 데이터베이스에서 같은 문장을 사용할 수 있다는 뜻은 아닙니다. 다른 DB로 옮길 때 공통 `database/sql` API와 SQL 방언의 차이를 구분합니다.

이번 각 CRUD 변경은 한 SQL 문장입니다. 여러 문장의 변경을 하나의 업무 단위로 묶는 요구사항은 연습문제에서 `sql.Tx`로 구현합니다. 인터페이스 뒤에 넣었다고 여러 메서드 호출이 자동으로 하나의 트랜잭션이 되지는 않습니다.

### 6.5 삭제와 ID 재사용

`Delete`는 영향을 받은 행 수를 확인합니다. SQL 실행 자체가 성공해도 일치한 행이 0개이면 이 API에서는 `ErrNotFound`입니다. 이 오류가 HTTP에서 404로 매핑됩니다.

SQLite 스키마에 `AUTOINCREMENT`를 넣은 이유는 삭제한 가장 큰 ID를 나중에 재사용하지 않는 계약을 맞추기 위해서입니다. 이 키워드에는 비용이 있으며 일반적으로 모든 테이블에 붙여야 하는 것은 아닙니다. [SQLite AUTOINCREMENT](https://sqlite.org/autoinc.html)에 나온 의미를 현재 계약과 연결합니다.

### 6.6 IF NOT EXISTS의 한계

기존 표가 있으면 그대로 사용하도록 한 문법은 스키마 이전(migration) 도구가 아닙니다. 기존 열 구조가 다른 경우 이를 자동으로 변환하거나 검증하지 않습니다. 이 예제는 새 학습 DB를 시작점으로 하고 같은 버전의 스키마를 다시 여는 흐름입니다.

표 구조를 변경하는 서비스에서는 스키마 버전과 순서 있는 이전 절차가 필요합니다. 여기서 해당 기능까지 숨겨서 구현하지 않습니다. 기존 운영 DB 파일을 연결하지 않는 이유도 학습 코드의 범위를 명확히 하기 위해서입니다.

## 7. 파일을 닫고 다시 열어 영속성을 확인한다

저장 경로: [`sqlite-demo/main.go`](../../examples/30-todo-api/sqlite-demo/main.go).

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() (err error) {
	dir, err := os.MkdirTemp("", "todo-book-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	ctx := context.Background()
	path := filepath.Join(dir, "tasks.db")
	first, err := todo.OpenSQLite(ctx, path)
	if err != nil {
		return err
	}
	task, createErr := first.Create(ctx, "저장한 할 일")
	if err := errors.Join(createErr, first.Close()); err != nil {
		return err
	}
	second, err := todo.OpenSQLite(ctx, path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, second.Close()) }()
	found, err := second.Get(ctx, task.ID)
	if err != nil {
		return err
	}
	fmt.Printf("다시 열기: %d %s 완료=%t\n", found.ID, found.Title, found.Done)
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
go run ./examples/30-todo-api/sqlite-demo
```

실행 결과:

```text
다시 열기: 1 저장한 할 일 완료=false
```


동일한 변수에서 Get을 한 번 더 호출하는 것으로는 파일 영속성을 검증하기 어렵습니다. 이 예제는 첫 DB 핸들을 닫고 같은 경로로 두 번째 핸들을 열어 조회합니다. 메모리 변수에 남아 있는 값을 출력하는 방식과 구분합니다.

임시 폴더는 이 실행이 직접 만들었으므로 실행이 끝날 때 정리합니다. 사용자에게서 받은 기존 폴더를 지우지 않습니다. 실제 서버 모드에서 `-db`로 지정한 파일은 종료할 때 삭제하지 않으며, 다시 실행하면 기존 데이터가 남습니다.

## 8. ID를 기준으로 다음 페이지 찾기

저장 경로: [`pagination-demo/main.go`](../../examples/30-todo-api/pagination-demo/main.go).

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	ctx := context.Background()
	repository := todo.NewMemory()
	for _, title := range []string{"첫째", "둘째", "셋째", "넷째", "다섯째"} {
		if _, err := repository.Create(ctx, title); err != nil {
			return err
		}
	}
	var after int64
	for number := 1; ; number++ {
		tasks, err := repository.List(ctx, after, 2)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			break
		}
		ids := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		fmt.Printf("페이지 %d: %v\n", number, ids)
		after = tasks[len(tasks)-1].ID
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
go run ./examples/30-todo-api/pagination-demo
```

실행 결과:

```text
페이지 1: [1 2]
페이지 2: [3 4]
페이지 3: [5]
```


`after=2`는 두 행을 건너뛰라는 의미가 아닙니다. **ID가 2보다 큰 항목**을 달라는 뜻입니다. ID 2가 삭제되었더라도 같은 기준으로 다음 항목을 찾을 수 있습니다. offset으로 위치를 세는 방식과 기준이 다릅니다.

| 요청 | 후보 조건 | 반환 상한 |
| --- | --- | --- |
| `after=0, limit=2` | ID > 0 | 2개 |
| `after=2, limit=2` | ID > 2 | 2개 |
| `after=4, limit=2` | ID > 4 | 2개 |

다음 기준은 이번에 읽은 마지막 ID입니다. 빈 결과가 나올 때 끝냅니다. 페이지마다 정렬 순서가 같아야 이 규칙이 동작합니다. 데이터가 동시에 추가·삭제된다면 전체 여러 페이지가 한 시점의 스냅샷이라는 보장은 없습니다.

마지막으로 본 ID보다 큰 새 항목은 다음 조회에 보일 수 있습니다. 고정된 시점의 전체 목록을 요구한다면 별도의 스냅샷·버전·트랜잭션 설계가 필요합니다. 페이지 조회의 일관성 요구를 먼저 정의합니다.

## 9. HTTP 입력과 저장소 사이를 연결한다

저장 경로: [`todo/http.go`](../../examples/30-todo-api/todo/http.go).

```go
package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

const MaxBodyBytes int64 = 1024

type API struct {
	repository Repository
	logger     *slog.Logger
}
type problem struct {
	Error string `json:"error"`
}
type createInput struct {
	Title *string `json:"title"`
}
type replaceInput struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func NewHandler(repository Repository, logger *slog.Logger) (http.Handler, error) {
	if repository == nil {
		return nil, errors.New("repository가 필요합니다")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	api := &API{repository: repository, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("/todos", api.collection)
	mux.HandleFunc("/todos/{id}", api.item)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { api.reply(w, r, 404, problem{"not_found"}) })
	var sequence atomic.Uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A local correlation ID, not a security token or cross-process global ID.
		w.Header().Set("X-Request-ID", strconv.FormatUint(sequence.Add(1), 10))
		mux.ServeHTTP(w, r)
	}), nil
}
func (a *API) reply(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		a.logger.ErrorContext(r.Context(), "encode response", "error", err)
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		a.logger.WarnContext(r.Context(), "write response", "request_id", w.Header().Get("X-Request-ID"), "error", err)
	}
}
func (a *API) failure(w http.ResponseWriter, r *http.Request, err error) {
	var validation *ValidationError
	switch {
	case errors.As(err, &validation):
		a.reply(w, r, 422, problem{"invalid_" + validation.Field})
	case errors.Is(err, ErrInvalidID), errors.Is(err, ErrInvalidPage):
		a.reply(w, r, 400, problem{"invalid_parameter"})
	case errors.Is(err, ErrNotFound):
		a.reply(w, r, 404, problem{"not_found"})
	case errors.Is(err, context.DeadlineExceeded):
		a.reply(w, r, 504, problem{"deadline_exceeded"})
	case errors.Is(err, context.Canceled):
		a.reply(w, r, 503, problem{"request_canceled"})
	default:
		a.logger.ErrorContext(r.Context(), "repository failure", "request_id", w.Header().Get("X-Request-ID"), "error", err)
		a.reply(w, r, 500, problem{"internal_error"})
	}
}
func (a *API) decode(w http.ResponseWriter, r *http.Request, target any) bool {
	media, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" || (params["charset"] != "" && !strings.EqualFold(params["charset"], "utf-8")) {
		a.reply(w, r, 415, problem{"json_required"})
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var max *http.MaxBytesError
		if errors.As(err, &max) {
			a.reply(w, r, 413, problem{"body_too_large"})
		} else {
			a.reply(w, r, 400, problem{"invalid_body"})
		}
		return false
	}
	if !utf8.Valid(body) {
		a.reply(w, r, 400, problem{"invalid_json"})
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		a.reply(w, r, 400, problem{"invalid_json"})
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		a.reply(w, r, 400, problem{"invalid_json"})
		return false
	}
	return true
}
func singleQuery(values url.Values, key string) (string, bool) {
	all, exists := values[key]
	if !exists {
		return "", true
	}
	if len(all) != 1 || all[0] == "" {
		return "", false
	}
	return all[0], true
}
func page(r *http.Request) (int64, int, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return 0, 0, ErrInvalidPage
	}
	for key := range values {
		if key != "after" && key != "limit" {
			return 0, 0, ErrInvalidPage
		}
	}
	afterText, ok := singleQuery(values, "after")
	if !ok {
		return 0, 0, ErrInvalidPage
	}
	limitText, ok := singleQuery(values, "limit")
	if !ok {
		return 0, 0, ErrInvalidPage
	}
	var after int64
	limit := 20
	if afterText != "" {
		after, err = strconv.ParseInt(afterText, 10, 64)
		if err != nil {
			return 0, 0, ErrInvalidPage
		}
	}
	if limitText != "" {
		limit, err = strconv.Atoi(limitText)
		if err != nil {
			return 0, 0, ErrInvalidPage
		}
	}
	if err := validatePage(after, limit); err != nil {
		return 0, 0, err
	}
	return after, limit, nil
}
func (a *API) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		after, limit, err := page(r)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		tasks, err := a.repository.List(r.Context(), after, limit)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		a.reply(w, r, 200, tasks)
	case http.MethodPost:
		if r.URL.RawQuery != "" {
			a.reply(w, r, 400, problem{"invalid_parameter"})
			return
		}
		var input *createInput
		if !a.decode(w, r, &input) {
			return
		}
		if input == nil || input.Title == nil {
			a.reply(w, r, 422, problem{"invalid_title"})
			return
		}
		task, err := a.repository.Create(r.Context(), *input.Title)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		w.Header().Set("Location", "/todos/"+strconv.FormatInt(task.ID, 10))
		a.reply(w, r, 201, task)
	default:
		w.Header().Set("Allow", "GET, HEAD, POST")
		a.reply(w, r, 405, problem{"method_not_allowed"})
	}
}
func (a *API) item(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "GET, HEAD, PUT, DELETE")
		a.reply(w, r, 405, problem{"method_not_allowed"})
		return
	}
	if r.URL.RawQuery != "" {
		a.reply(w, r, 400, problem{"invalid_parameter"})
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		a.failure(w, r, ErrInvalidID)
		return
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		task, err := a.repository.Get(r.Context(), id)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		a.reply(w, r, 200, task)
	case http.MethodPut:
		var input *replaceInput
		if !a.decode(w, r, &input) {
			return
		}
		if input == nil || input.Title == nil || input.Done == nil {
			a.reply(w, r, 422, problem{"missing_field"})
			return
		}
		task, err := a.repository.Replace(r.Context(), id, *input.Title, *input.Done)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		a.reply(w, r, 200, task)
	case http.MethodDelete:
		if err := a.repository.Delete(r.Context(), id); err != nil {
			a.failure(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
```

### 9.1 경로를 찾고 메서드를 분기한다

`ServeMux`는 `/todos`와 `/todos/{id}`를 찾습니다. 각 핸들러는 허용한 메서드를 명시적으로 분기합니다. 이 구성에서는 일반적인 404·405도 JSON으로 맞추기 위해 메서드별 오류 응답을 직접 만듭니다.

허용하지 않는 메서드에는 405와 `Allow` 헤더를 반환합니다. 존재하지 않는 경로와 메서드가 틀린 기존 경로를 구분합니다. mux 자체가 수행하는 경로 정리·리다이렉트 같은 표준 동작까지 모두 JSON으로 바꿨다는 뜻은 아닙니다. 자세한 라우팅 계약은 [http.ServeMux](https://pkg.go.dev/net/http#ServeMux)를 참고합니다.

### 9.2 입력 검증 순서

POST의 처리 순서를 따라 봅니다.

1. 이 경로가 받지 않는 쿼리 문자열을 거절합니다.
2. Content-Type이 application/json인지 확인합니다.
3. 본문을 1,024바이트까지 읽고 초과하면 413으로 끝냅니다.
4. 원래 본문 바이트가 UTF-8인지 확인합니다.
5. 모르는 필드가 없는 JSON 하나인지 검사합니다.
6. 필수 title 필드가 존재하고 null이 아닌지 확인합니다.
7. 저장소가 제목의 업무 규칙을 검증하고 저장합니다.
8. Location과 생성한 항목을 201로 반환합니다.

오류가 난 단계 뒤의 처리를 계속하지 않습니다. 잘못된 JSON이 들어왔는데 빈 제목을 저장하거나, 실패 응답을 쓴 뒤 성공 응답까지 추가하는 일을 피합니다.

### 9.3 전체 크기를 먼저 확인한다

JSON 디코더는 값 하나를 읽고 성공할 수 있습니다. 정상 JSON 뒤에 아주 긴 공백이 붙어 있다면 첫 Decode만으로는 입력 전체의 크기를 확인하기 어렵습니다. 이 예제는 `MaxBytesReader`를 거쳐 `ReadAll`로 전체 입력을 제한한 뒤 파싱합니다.

최대 1,024바이트라는 작은 계약이 있으므로 읽는 메모리 상한도 설명할 수 있습니다. 대형 파일을 같은 방식으로 무제한 읽으라는 뜻이 아닙니다. 스트리밍 입력에서는 제한과 처리 단위를 다르게 설계합니다. [MaxBytesReader](https://pkg.go.dev/net/http#MaxBytesReader)와 [MaxBytesError](https://pkg.go.dev/net/http#MaxBytesError)를 확인합니다.

### 9.4 누락과 false를 구분한다

PUT의 done을 일반 bool로 받으면 누락과 명시적인 false가 모두 false로 보입니다. `*bool`을 사용하면 nil로 누락·null을 감지할 수 있습니다. 같은 이유로 title은 `*string`입니다.

| PUT 입력 | 해석 |
| --- | --- |
| `{"title":"Go","done":true}` | 완료로 교체 |
| `{"title":"Go","done":false}` | 미완료로 교체 |
| `{"title":"Go"}` | 필수 done 없음 |
| `{"title":"Go","done":null}` | done이 유효한 논리값이 아님 |
| `null` | 수정할 객체 자체가 없음 |

구조체 포인터 자체도 nil일 수 있어 `input == nil`을 먼저 확인합니다. nil을 역참조하는 패닉으로 입력 오류를 처리하지 않습니다.

### 9.5 JSON v1에서 더 엄격한 것과 그렇지 않은 것

`DisallowUnknownFields`는 모르는 필드를 거절합니다. 하지만 `encoding/json`의 기존 동작에서 중복 키를 자동 거절하거나 구조체 필드 이름을 대소문자까지 엄격하게 비교한다는 뜻은 아닙니다. 이 API는 `{"TITLE":"first","title":"second"}`에서 뒤의 값을 사용할 수 있으며 그 동작을 테스트로 명시합니다.

원래 바이트의 잘못된 UTF-8은 디코더 앞에서 거절합니다. JSON 안의 잘못된 UTF-16 서로게이트 이스케이프를 기존 디코더가 U+FFFD로 바꾸는 동작까지 별도로 금지하지는 않습니다. 더 엄격한 상호 운용 계약이 필요하면 적절한 JSON API와 옵션을 선택하고 호환성 테스트를 바꿔야 합니다. [encoding/json](https://pkg.go.dev/encoding/json)의 동작 차이를 확인합니다.

이처럼 “엄격한 JSON”이라는 한 문장보다 실제로 거절하는 사례를 표와 테스트로 적는 편이 정확합니다. 이 장의 범위는 unknown 필드·추가 JSON 값·잘못된 바이트·크기·도메인 필드 검증입니다.

### 9.6 쿼리도 입력이다

`url.ParseQuery`의 오류를 확인합니다. `limit=1&limit=2`처럼 같은 키가 여러 번 나오면 하나를 임의로 고르지 않습니다. 빈 값과 모르는 키도 거절합니다. `limit`의 범위는 1~100이며 `after`는 0 이상입니다.

URL에 있는 문자열을 단순히 `r.URL.Query().Get`으로 읽으면 일부 파싱 오류나 중복 정책을 놓치기 쉽습니다. 이 예제는 그 결정을 `page`와 `singleQuery` 함수로 모읍니다. 숫자의 오버플로도 파싱 오류로 반환됩니다.

### 9.7 요청 ID와 로그

각 핸들러 인스턴스는 원자적 카운터로 `X-Request-ID`를 발급합니다. 사용자가 보낸 같은 이름의 값을 그대로 신뢰하지 않고 새 값을 만듭니다. 이 ID는 로컬 실행의 상관관계 식별자이며 여러 프로세스에 걸친 전역 고유 ID나 인증 수단이 아닙니다.

내부 저장소 오류는 로그에 남기고 응답에는 `internal_error`만 보냅니다. 실제 저장소 오류에 비밀값이 섞이는 환경이라면 로깅 정책도 추가로 제한해야 합니다. 학습용 SQLite에는 실제 계정 비밀값이 없으며 토큰이나 요청 본문을 통째로 기록하지 않습니다.

## 10. HTTP 계약을 눈으로 확인한다

저장 경로: [`http-demo/main.go`](../../examples/30-todo-api/http-demo/main.go).

```go
package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	handler, err := todo.NewHandler(todo.NewMemory(), nil)
	if err != nil {
		return err
	}
	cases := []struct{ method, path, body string }{
		{"POST", "/todos", `{"title":" Go 복습 "}`},
		{"GET", "/todos/1", ""},
		{"PUT", "/todos/1", `{"title":"Go 복습","done":true}`},
		{"GET", "/todos?limit=2", ""},
		{"DELETE", "/todos/1", ""},
		{"GET", "/todos/1", ""},
	}
	for _, tc := range cases {
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		line := fmt.Sprintf("%s %s -> %d", tc.method, tc.path, response.Code)
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			line += " " + body
		}
		fmt.Println(line)
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
go run ./examples/30-todo-api/http-demo
```

실행 결과:

```text
POST /todos -> 201 {"id":1,"title":"Go 복습","done":false}
GET /todos/1 -> 200 {"id":1,"title":"Go 복습","done":false}
PUT /todos/1 -> 200 {"id":1,"title":"Go 복습","done":true}
GET /todos?limit=2 -> 200 [{"id":1,"title":"Go 복습","done":true}]
DELETE /todos/1 -> 204
GET /todos/1 -> 404 {"error":"not_found"}
```


각 요청마다 새 Recorder를 만듭니다. 이전 요청의 상태와 본문이 다음 요청에 남으면 API 자체가 잘못된 것처럼 보일 수 있습니다. 메모리 저장소와 핸들러는 같은 객체를 재사용하므로 요청 사이 데이터는 유지됩니다.

이 데모는 POST에서 만든 항목을 GET, PUT, 목록 조회, DELETE로 이어서 확인합니다. 마지막 GET의 404는 앞선 삭제가 실제 저장소 상태에 반영되었다는 증거입니다. `DELETE` 성공 뒤 204에는 JSON 본문을 붙이지 않습니다.

테스트에서는 출력 문자열만 비교하지 않고 Location, Content-Type, Allow, 요청 ID, 필수 필드와 본문 없음도 확인합니다. 상태 코드 200 하나만 맞는다는 이유로 계약 전체가 맞다고 판단하지 않습니다.

## 11. 실제 HTTP 왕복을 한 번 더 검사한다

저장 경로: [`client-demo/main.go`](../../examples/30-todo-api/client-demo/main.go).

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
	"strings"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	handler, err := todo.NewHandler(todo.NewMemory(), nil)
	if err != nil {
		return err
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	client := server.Client()
	client.Timeout = time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/todos", strings.NewReader(`{"title":"HTTP 실습"}`))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 4097))
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return err
	}
	if len(body) > 4096 {
		return errors.New("응답 크기 초과")
	}
	if response.StatusCode != 201 {
		return fmt.Errorf("예상과 다른 상태: %d", response.StatusCode)
	}
	var task todo.Task
	if err := json.Unmarshal(body, &task); err != nil {
		return err
	}
	fmt.Printf("생성: %d %s 완료=%t\n", task.ID, task.Title, task.Done)
	fmt.Println("위치:", response.Header.Get("Location"))
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
go run ./examples/30-todo-api/client-demo
```

실행 결과:

```text
생성: 1 HTTP 실습 완료=false
위치: /todos/1
```


이 데모는 `httptest.NewServer`가 연 로컬 listener로 요청을 보냅니다. `ResponseRecorder`로 핸들러를 직접 호출할 때와 달리 요청·응답 직렬화와 실제 HTTP 클라이언트 경계를 통과합니다. 외부 서버는 필요하지 않습니다.

클라이언트는 제한 시간을 설정하고 요청에도 context를 연결합니다. `Do`의 오류와 HTTP 상태는 다릅니다. 500 응답을 정상적으로 수신하면 네트워크 호출 자체는 오류 없이 돌아올 수 있으므로 201인지 별도로 확인합니다.

본문은 4,097바이트까지 읽어 4,096바이트 상한 초과를 구분합니다. 읽기 오류와 Close 오류를 모두 확인합니다. HTTP 테스트에는 `httptest.NewTLSServer`도 사용하여 테스트 클라이언트가 신뢰하는 TLS 연결에서 HEAD의 본문 없음까지 확인합니다.

## 12. 같은 테스트를 두 저장소에 적용한다

검증 파일은 [`repository_test.go`](../../examples/30-todo-api/todo/repository_test.go)와 [`http_test.go`](../../examples/30-todo-api/todo/http_test.go)입니다. 각 테스트는 새 저장소를 만들어 다른 사례와 데이터를 공유하지 않습니다.

### 12.1 저장소 계약 테스트

| 검사 | 중요한 관찰 |
| --- | --- |
| 빈 목록 | nil 대신 빈 슬라이스 |
| 생성 | 제목 정규화, 새 양수 ID, 초기 done=false |
| 값 복사 | 호출자가 수정해도 내부 값 유지 |
| 교체 | false도 명시적으로 저장, 잘못된 제목은 기존 값 보존 |
| 삭제 | 실제 사라짐, 반복 삭제는 ErrNotFound |
| ID 재발급 | 삭제한 ID를 그대로 재사용하지 않음 |
| 페이지 | ID 순서, after 경계, limit 범위 |
| 취소 | 취소된 호출이 반환되고 기존 상태 유지 |
| 동시 생성 | 중복 ID 없음, 최종 개수 일치 |
| 영속성 | 닫고 새 핸들로 열어 같은 값 확인 |

메모리와 SQLite가 같은 테스트를 통과해야 “교체 가능한 저장소”라는 설명을 뒷받침할 수 있습니다. 메서드 이름이 같다는 사실만으로 오류·정렬·취소의 의미까지 같아지는 것은 아닙니다.

### 12.2 HTTP 실패 테스트

입력 오류 테스트는 실패 응답만 보고 끝나지 않습니다. 모든 실패 요청 뒤 저장소 목록이 여전히 비어 있는지도 확인합니다. 이는 잘못된 요청이 내부 변경까지 진행한 버그를 잡기 위한 조건입니다.

특히 본문은 1,024바이트일 때 성공하고 1,025바이트일 때 실패하는 경계를 검사합니다. UTF-8 한글은 여러 바이트이므로 제목의 80룬 제한과 HTTP의 바이트 제한이 서로 다른 경계임을 확인합니다.

알려진 오류·감싼 ErrNotFound·취소·마감·일반 내부 오류를 반환하는 저장소 대역도 연결합니다. 해당 테스트에서 원래 요청 context가 저장소에 그대로 전달되는지, 잘못된 ID가 저장소 호출 전에 거절되는지 확인합니다.

```sh
go test -count=1 ./examples/30-todo-api/...
go test -count=1 ./solutions/30-todo-api/...
go test -race ./examples/30-todo-api/... ./solutions/30-todo-api/...
```

데이터 레이스 검사는 지원 환경과 C 컴파일러가 필요합니다. 동시성 테스트는 goroutine의 진입 시각을 Sleep으로 추정하지 않습니다. 모든 생성의 결과를 채널과 WaitGroup으로 수집한 뒤 최종 상태를 확인합니다.

## 13. 저장소를 선택하는 실제 서버

저장 경로: [`server/main.go`](../../examples/30-todo-api/server/main.go).

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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/25-service-operations/service"
	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func requestDeadline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func run(ctx context.Context, args []string, diagnostics io.Writer) (err error) {
	flags := flag.NewFlagSet("todo-api", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	address := flags.String("addr", "127.0.0.1:8080", "loopback 호스트:포트")
	store := flags.String("store", "memory", "memory 또는 sqlite")
	path := flags.String("db", "", "sqlite를 선택했을 때 사용할 파일 경로")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("위치 인자는 받지 않습니다")
	}
	host, _, err := net.SplitHostPort(*address)
	if err != nil {
		return errors.New("올바른 호스트:포트가 필요합니다")
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("학습 서버는 loopback IP만 사용합니다")
	}
	if *store != "memory" && *store != "sqlite" {
		return errors.New("store는 memory 또는 sqlite여야 합니다")
	}
	if *store == "sqlite" && *path == "" {
		return errors.New("sqlite에는 -db 경로가 필요합니다")
	}
	if *store == "memory" && *path != "" {
		return errors.New("memory 모드에는 -db를 지정하지 않습니다")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var repository todo.Repository
	if *store == "sqlite" {
		startup, cancel := context.WithTimeout(ctx, 3*time.Second)
		db, openErr := todo.OpenSQLite(startup, *path)
		cancel()
		if openErr != nil {
			return openErr
		}
		defer func() { err = errors.Join(err, db.Close()) }()
		repository = db
	} else {
		repository = todo.NewMemory()
	}
	logger := slog.New(slog.NewJSONHandler(diagnostics, nil)).With("service", "todo-api")
	handler, err := todo.NewHandler(repository, logger)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		return err
	}
	server := &http.Server{Handler: requestDeadline(handler), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	logger.Info("listening", "address", listener.Addr().String(), "store", *store)
	if err := service.Serve(ctx, server, listener, 3*time.Second); err != nil {
		return err
	}
	logger.Info("stopped")
	return nil
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.Stderr)
	stop()
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

### 13.1 메모리 서버 실행

```sh
go run ./examples/30-todo-api/server -store=memory -addr=127.0.0.1:8080
```

이 명령은 자동으로 끝나지 않습니다. 로그에 listening이 나타나면 다른 터미널에서 요청합니다. 첫 터미널에서 **Ctrl+C**를 눌러 종료합니다. Linux 등에서 사용하는 SIGTERM도 등록합니다. 실제 신호 전달 방식은 운영체제와 실행 환경의 영향을 받습니다.

메모리 모드에서는 종료 후 데이터가 사라집니다. 같은 명령으로 다시 실행하면 새 저장소가 생깁니다. 포트가 이미 사용 중이면 다른 loopback 포트를 선택합니다.

### 13.2 PowerShell에서 요청하기

```powershell
$base = "http://127.0.0.1:8080"
$createBody = @{ title = "Go 복습" } | ConvertTo-Json -Compress
$created = Invoke-RestMethod -Method Post -Uri "$base/todos" -ContentType "application/json; charset=utf-8" -Body ([System.Text.Encoding]::UTF8.GetBytes($createBody))
$created
Invoke-RestMethod -Method Get -Uri "$base/todos/$($created.id)"
$updateBody = @{ title = "Go 복습"; done = $true } | ConvertTo-Json -Compress
Invoke-RestMethod -Method Put -Uri "$base/todos/$($created.id)" -ContentType "application/json; charset=utf-8" -Body ([System.Text.Encoding]::UTF8.GetBytes($updateBody))
Invoke-RestMethod -Method Get -Uri "$base/todos?limit=2"
Invoke-RestMethod -Method Delete -Uri "$base/todos/$($created.id)"
```

생성 결과의 ID를 사용하므로 이미 데이터가 있는 DB에서도 동작합니다. PowerShell 버전에 따른 요청 문자열 인코딩 차이를 줄이려고 UTF-8 바이트로 전달합니다. 표에 나타나는 속성 순서는 PowerShell의 표시 방식에 따라 다를 수 있습니다.

### 13.3 macOS·Linux·WSL에서 요청하기

curl이 설치되어 있다면 먼저 생성 요청을 보냅니다.

```sh
curl -i -X POST http://127.0.0.1:8080/todos -H 'Content-Type: application/json' -d '{"title":"Go 복습"}'
```

응답의 `Location`과 `id`를 확인합니다. 아래는 **새 메모리 서버에서 ID 1을 받은 경우**의 명령입니다. 다른 ID가 나왔다면 해당 값으로 바꿉니다.

```sh
curl -i http://127.0.0.1:8080/todos/1
curl -i -X PUT http://127.0.0.1:8080/todos/1 -H 'Content-Type: application/json' -d '{"title":"Go 복습","done":true}'
curl -i 'http://127.0.0.1:8080/todos?limit=2'
curl -i -X DELETE http://127.0.0.1:8080/todos/1
```

Date 같은 응답 헤더와 요청 ID는 실행 시점에 따라 달라질 수 있습니다. 생성·조회·수정·삭제의 상태와 JSON 필드를 확인합니다.

### 13.4 파일 DB 서버 실행

먼저 학습 폴더를 만듭니다. Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force "_practice" | Out-Null
```

macOS·Linux·WSL:

```bash
mkdir -p _practice
```

공통 실행 명령:

```sh
go run ./examples/30-todo-api/server -store=sqlite -db=./_practice/todo.db -addr=127.0.0.1:8080
```

새 항목을 만든 뒤 Ctrl+C로 종료하고 같은 명령으로 다시 실행합니다. GET /todos에서 데이터가 유지되는지 확인합니다. 이 프로그램은 지정한 DB 파일을 종료 시 삭제하지 않습니다.

`-store=sqlite`에서 `-db`를 생략하면 시작 오류입니다. 메모리 모드에 `-db`를 함께 주면 무시하지 않고 오류를 반환합니다. 사용자가 파일에 저장된다고 오해한 채 메모리 모드로 실행하지 않게 하는 선택입니다.

### 13.5 실행·요청·종료의 세 수명

DB 초기화에는 최대 3초를 줍니다. HTTP 요청은 요청 context에서 최대 2초를 만들고 저장소에 전달합니다. 종료 신호가 오면 25장의 `service.Serve`가 별도의 3초 정리 수명을 사용합니다.

요청의 2초 제한은 context를 관찰하는 저장소 작업에 전달하는 협조적 취소입니다. context를 만들었다고 모든 `Body.Read`나 임의의 계산을 강제로 중단하지는 않습니다. 이 실제 서버에서 소켓의 본문 읽기는 `ReadTimeout` 5초로 제한합니다. 전체 요청이 어떤 입력에서도 정확히 2초 안에 종료된다는 보장은 아닙니다.

이 세 context는 용도가 다릅니다. 서버의 BaseContext를 종료 신호 context로 곧바로 연결하지 않았으므로 정상 종료 시작만으로 모든 활성 요청을 즉시 취소하지 않습니다. 요청을 마칠 시간을 주고, 끝나지 않으면 강제 연결 닫기 오류를 반환합니다.

SQLite Close는 `run`의 defer에 있습니다. 정상 종료를 기다린 뒤 DB를 닫습니다. 종료 대기가 실패해도 오류와 Close 결과를 함께 보존합니다. 단순히 “stopped”를 항상 출력하지 않고 실제 정상 경로에서만 기록합니다.

## 14. 구현 범위를 정확하게 읽기

이번 프로그램은 학습 가능한 작은 CRUD API입니다. 다음 경계도 분명히 알고 사용합니다.

| 현재 선택 | 확장할 때 고려할 점 |
| --- | --- |
| loopback 주소만 수신 | 외부 노출, TLS 종료 지점, 접근 제어 |
| 사용자 구분 없음 | 인증된 사용자와 항목 소유권 |
| PUT은 마지막 성공 값으로 교체 | 버전 필드, ETag, 조건부 갱신 |
| ID는 JSON 숫자 | JavaScript의 큰 정수 정밀도가 필요한 경우 문자열 ID |
| 목록은 페이지마다 조회 | 동일 시점 스냅샷이 필요한지 |
| 한 DB 연결의 SQLite | 여러 프로세스·큰 부하의 잠금과 풀 정책 |
| CREATE TABLE IF NOT EXISTS | 버전이 있는 스키마 이전 |
| POST 자동 재시도 없음 | 멱등성 키와 중복 생성 방지 |

특히 테스트의 레이스 통과와 업무 수준의 갱신 충돌은 다릅니다. 두 요청이 같은 항목을 읽은 뒤 서로 다른 수정본을 PUT하면, 각 Replace가 안전하게 실행되어도 나중 값이 앞선 변경을 덮을 수 있습니다. 메모리 경쟁이 없다는 사실만으로 사용자의 동시 편집 의도까지 보존되는 것은 아닙니다.

context 취소도 완료된 변경을 되돌리는 기능이 아닙니다. 서버에서 이미 저장했지만 응답을 받기 전에 연결이 끊겼을 수 있습니다. 네트워크 오류를 보고 POST를 무조건 반복하면 같은 제목의 항목이 두 개 생길 수 있습니다. 이 API는 그러한 재시도를 자동으로 수행하지 않습니다.

## 15. 자주 하는 실수

| 증상 | 원인 | 확인·수정 |
| --- | --- | --- |
| 메모리와 DB의 목록 순서가 다름 | 맵 순회·ORDER BY 생략 | ID 오름차순 계약 검사 |
| done=false가 빠진 값으로 취급됨 | bool만으로 필수 여부 검사 | *bool로 존재와 값을 구분 |
| 빈 목록이 null로 나옴 | nil 슬라이스 반환 | 빈 슬라이스 초기화 |
| 수정 뒤 이전 값 일부가 사라짐 | 읽기·쓰기 사이 동시 갱신 | 필요한 충돌 정책을 별도 설계 |
| 404 대신 내부 오류 문자열 노출 | 저장소 오류를 응답에 그대로 사용 | 오류 분류와 일반 코드 매핑 |
| 1,024바이트 제한을 우회함 | 첫 JSON 값만 읽고 끝냄 | 전체 제한 입력을 먼저 읽기 |
| SQLite 테스트가 가끔 표를 못 찾음 | :memory: 여러 연결 | 연결 수·DB 수명 확인 |
| 같은 본문 요청이 두 번 저장됨 | POST를 무조건 재시도 | 멱등성 요구 정의 |
| 서버 종료 직전 DB 오류 | 요청 대기 전 DB를 닫음 | 서버 대기 뒤 공유 DB 정리 |
| 새 버전에서 기존 DB가 깨짐 | IF NOT EXISTS를 migration으로 오해 | 스키마 버전 절차 설계 |

실패를 찾을 때 HTTP 응답, 저장소 상태, 서버 로그를 함께 봅니다. 오류 응답만 맞고 데이터가 잘못 바뀌었을 수 있으며, 데이터는 맞지만 Content-Type이나 종료 동작이 틀릴 수도 있습니다.

## 연습문제

[30장 연습문제](../../exercises/30-todo-api/README.md)는 상태 추적, 필터링한 페이지의 커서, 원자적 일괄 저장, 오류를 분류하는 클라이언트와 요청 제한 시간으로 이어집니다. 먼저 요구사항과 실패 상태를 적은 뒤 [해설](../../solutions/30-todo-api/README.md)과 비교합니다.

## 요약과 자기 점검

도메인 타입이 입력과 오류를 정의하고, 저장소가 같은 CRUD 계약을 지키며, HTTP 층이 상태·헤더·JSON으로 변환합니다. 실제 실행은 저장소를 선택하고 요청 수명과 종료 순서를 관리합니다. 각 경계를 별도로 시험하고 같은 계약 테스트를 여러 구현에 적용하면 교체의 의미를 확인할 수 있습니다.

<details>
<summary>1. 인터페이스가 같으면 두 저장소가 자동으로 교체 가능하나요?</summary>

아닙니다. 입력 검증, 정렬, 없는 항목, ID와 취소 의미도 같아야 합니다. 같은 동작 테스트를 적용해 확인합니다.

</details>

<details>
<summary>2. title을 HTTP에서 검사했는데 저장소에서도 검사하는 이유는 무엇인가요?</summary>

HTTP 이외의 CLI나 배치도 저장소를 호출할 수 있기 때문입니다. 공통 도메인 규칙을 저장소 경계에서 지켜야 구현을 직접 사용해도 계약이 유지됩니다.

</details>

<details>
<summary>3. 빈 목록을 nil로 두지 않은 이유는 무엇인가요?</summary>

JSON에서 배열을 일관되게 반환하기 위해서입니다. 이 API의 빈 목록 계약은 null이 아니라 []입니다.

</details>

<details>
<summary>4. after=2인데 ID 2가 삭제되면 어떻게 되나요?</summary>

ID가 2보다 큰 항목을 찾으므로 기준 행 자체가 없어도 조회할 수 있습니다. 다만 여러 페이지가 한 시점의 스냅샷인 것은 아닙니다.

</details>

<details>
<summary>5. ReadTimeout이나 WriteTimeout이 있으면 저장소 context 제한은 필요 없나요?</summary>

아닙니다. 네트워크 읽기·쓰기 제한과 내부 작업의 협조적 취소 수명은 다릅니다. 예제는 별도의 요청 context 제한을 전달합니다.

</details>

<details>
<summary>6. 클라이언트가 시간 초과를 받으면 저장은 반드시 실패했나요?</summary>

아닙니다. 저장은 끝났지만 응답 전달이 실패했을 수 있습니다. 특히 생성 요청 재시도에는 중복 처리 계약이 필요합니다.

</details>

## 참고 자료

- [net/http](https://pkg.go.dev/net/http): 서버·라우팅·본문 제한·응답 계약
- [net/http/httptest](https://pkg.go.dev/net/http/httptest): 핸들러와 실제 HTTP 테스트
- [encoding/json](https://pkg.go.dev/encoding/json): 누락·null·중복 키·Unicode 처리
- [database/sql](https://pkg.go.dev/database/sql): context, Scan, Rows와 연결 수명
- [Go 트랜잭션 안내](https://go.dev/doc/database/execute-transactions): 여러 변경의 확정 경계
- [Go SQL 매개변수 안내](https://go.dev/doc/database/sql-injection): 값과 문장 분리
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite): 사용한 드라이버
- [SQLite RETURNING](https://sqlite.org/lang_returning.html): 변경한 행 반환
- [SQLite AUTOINCREMENT](https://sqlite.org/autoinc.html): ID 재사용 정책
- [SQLite 메모리 DB](https://sqlite.org/inmemorydb.html): 연결별 수명

공식 API와 드라이버 확인일: **2026-11-09**. 기본 의존성 버전은 루트 go.mod와 go.sum을 기준으로 합니다.

[이전: 29장 CSV 분석 CLI](../29-csv-cli/README.md) · [전체 목차](../README.md) · [다음: 31장 동시 배치 처리](../31-concurrent-batch/README.md)
