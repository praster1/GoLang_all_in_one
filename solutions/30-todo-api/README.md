# 30장 해설 — 계약을 유지하며 기능 확장하기

[연습문제](../../exercises/30-todo-api/README.md)를 먼저 시도한 뒤 읽습니다. 네 구현은 본문의 todo 패키지를 활용하며, 일괄 가져오기는 별도 표를 사용합니다. 따라서 기존 API의 제목 중복 허용 규칙과 새로운 문제의 고유성 규칙을 혼동하지 않습니다.

모든 실행 명령은 **저장소 루트**에서 실행합니다. 자동 데모는 실행 후 끝나고, SQL 테스트는 메모리 DB를 사용합니다. 사용자의 기존 DB 파일을 수정하거나 삭제하지 않습니다.

## 문제 1. 실패한 수정은 기존 상태를 보존한다

| 순서 | 요청 | 상태 | 저장 결과 |
| --- | --- | --- | --- |
| 1 | title=`" Go "` 생성 | 201 | ID 1, title=`"Go"`, done=false |
| 2 | 제목 변경·완료 | 200 | title=`"Go 완료"`, done=true |
| 3 | 공백 제목으로 수정 | 422 | 2번 상태 유지 |
| 4 | 항목 조회 | 200 | `Go 완료`, true |
| 5 | 삭제 | 204 | 항목 없음 |
| 6 | 같은 항목 삭제 | 404 | 항목 없음 유지 |
| 7 | 목록 조회 | 200 | 빈 배열 `[]` |

3번의 상태가 422라고 해서 저장소가 반드시 보존되었다고 단정할 수는 없습니다. 구현에 따라 일부 값을 먼저 변경한 뒤 검증에 실패했을 수 있습니다. 따라서 4번에서 실제 값도 확인합니다. 코드에서는 title을 정규화하고 검증한 뒤 맵이나 DB를 변경합니다.

반복 삭제의 404는 이번 API가 정한 계약입니다. 모든 삭제 요청이 존재 여부와 관계없이 204가 되어야 한다는 규칙으로 바꾸지 않습니다. 어떤 API는 다른 계약을 택할 수 있지만, 그때는 문서와 테스트를 함께 바꾸어야 합니다.

메모리와 SQLite 모두 같은 결과를 내야 합니다. SQLite에서 SQL 실행이 성공했다는 이유만으로 영향받은 행 수 0을 삭제 성공으로 처리하면 이 계약을 위반합니다. `RowsAffected`를 확인하는 이유입니다.

## 문제 2. 필터 결과의 마지막 항목은 커서가 아닐 수 있다

저장 경로: [`filtered-page/main.go`](filtered-page/main.go).

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

type OpenPage struct {
	Tasks     []todo.Task
	NextAfter int64
	HasMore   bool
}

func openPage(ctx context.Context, repo todo.Repository, after int64, size int) (OpenPage, error) {
	if size < 1 || size > 99 {
		return OpenPage{}, errors.New("size는 1~99여야 합니다")
	}
	source, err := repo.List(ctx, after, size+1)
	if err != nil {
		return OpenPage{}, err
	}
	page := OpenPage{Tasks: make([]todo.Task, 0), NextAfter: after, HasMore: len(source) > size}
	if page.HasMore {
		source = source[:size]
	}
	for _, task := range source {
		page.NextAfter = task.ID
		if !task.Done {
			page.Tasks = append(page.Tasks, task)
		}
	}
	return page, nil
}
func run() error {
	ctx := context.Background()
	repo := todo.NewMemory()
	for i := 1; i <= 5; i++ {
		task, err := repo.Create(ctx, fmt.Sprintf("항목 %d", i))
		if err != nil {
			return err
		}
		if i%2 == 1 {
			if _, err := repo.Replace(ctx, task.ID, task.Title, true); err != nil {
				return err
			}
		}
	}
	var after int64
	for number := 1; ; number++ {
		page, err := openPage(ctx, repo, after, 2)
		if err != nil {
			return err
		}
		ids := make([]int64, 0, len(page.Tasks))
		for _, task := range page.Tasks {
			ids = append(ids, task.ID)
		}
		fmt.Printf("페이지 %d: 미완료=%v 다음=%d 더있음=%t\n", number, ids, page.NextAfter, page.HasMore)
		if !page.HasMore {
			break
		}
		after = page.NextAfter
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
go run ./solutions/30-todo-api/filtered-page
```

실행 결과:

```text
페이지 1: 미완료=[2] 다음=2 더있음=true
페이지 2: 미완료=[4] 다음=4 더있음=true
페이지 3: 미완료=[] 다음=5 더있음=false
```

검증 코드: [`filtered-page/behavior_test.go`](filtered-page/behavior_test.go).

```go
package main

import (
	"context"
	"errors"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func TestEmptyFilteredPageStillAdvances(t *testing.T) {
	ctx := context.Background()
	repo := todo.NewMemory()
	for i := 1; i <= 3; i++ {
		task, err := repo.Create(ctx, "task")
		if err != nil {
			t.Fatal(err)
		}
		if i < 3 {
			if _, err := repo.Replace(ctx, task.ID, task.Title, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	first, err := openPage(ctx, repo, 0, 2)
	if err != nil || len(first.Tasks) != 0 || first.NextAfter != 2 || !first.HasMore {
		t.Fatalf("first=%+v %v", first, err)
	}
	next, err := openPage(ctx, repo, first.NextAfter, 2)
	if err != nil || len(next.Tasks) != 1 || next.Tasks[0].ID != 3 || next.HasMore {
		t.Fatalf("next=%+v %v", next, err)
	}
	empty, err := openPage(ctx, repo, 3, 2)
	if err != nil || empty.Tasks == nil || empty.NextAfter != 3 || empty.HasMore {
		t.Fatalf("empty=%+v %v", empty, err)
	}
	for _, size := range []int{0, 100} {
		if _, err := openPage(ctx, repo, 0, size); err == nil {
			t.Fatal("accepted size")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := openPage(canceled, repo, 0, 2); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
```


### 원본 페이지와 표시 페이지

원본 페이지는 저장소에서 읽은 순서 있는 항목들입니다. 표시 페이지는 그중 미완료 조건을 통과한 항목들입니다. 둘의 길이와 마지막 ID가 다를 수 있으므로 커서는 원본을 기준으로 움직입니다.

예를 들어 원본 ID 1·2가 모두 완료이고 ID 3만 미완료라고 합시다. 첫 표시 페이지는 비어 있지만 검사한 범위는 ID 2까지입니다. `Tasks`가 비었다는 이유로 종료하면 ID 3을 보지 못하고, 커서를 0으로 유지하면 같은 두 항목을 계속 읽을 수 있습니다.

이번 함수는 `size+1`개를 읽어 다음 항목이 있다는 사실을 관찰합니다. 관찰용 마지막 한 항목은 이번 페이지에서 처리하지 않습니다. 그래서 NextAfter를 그 항목의 ID로 올리면 다음 페이지에서 빠지는 버그가 생깁니다. `source[:size]`로 자른 후 검사한 원본의 마지막 ID만 사용합니다.

### 크기 계약의 의미

size는 검사할 원본 개수입니다. 미완료 항목을 size개 채워서 반환하는 기능은 아닙니다. 따라서 소량의 조회로 상한을 설명할 수 있고, 완료 항목이 많다고 내부에서 무한히 다음 페이지를 읽지 않습니다.

| 경계 | 결과 |
| --- | --- |
| size=0 또는 100 | 오류 |
| 원본 없음 | 빈 Tasks, 기존 커서, HasMore=false |
| 원본은 있지만 모두 완료 | 빈 Tasks, 검사한 마지막 커서 |
| size+1개 존재 | HasMore=true |
| 이미 취소된 context | 저장소 취소 오류 반환 |

여러 페이지를 읽는 동안 데이터가 바뀔 수 있습니다. HasMore는 해당 조회 시점에 다음 원본이 있었다는 의미입니다. 이후 다른 요청이 그 항목을 삭제하면 다음 페이지가 달라질 수 있습니다. 전체 스냅샷 계약은 따로 필요합니다.

## 문제 3. 검증 옵션의 실제 범위

| 입력 | 결과 | 이유 |
| --- | --- | --- |
| 알 수 없는 admin 필드 | 400 | DisallowUnknownFields |
| JSON 객체 뒤 두 번째 값 | 400 | 두 번째 Decode가 EOF가 아님 |
| TITLE와 title 중복 | 예제에서는 201, 뒤 값 사용 | JSON v1의 필드 매칭·중복 처리 |
| title:null | 422 | 필수 문자열 포인터가 nil |
| PUT의 done:false | 다른 필드가 정상일 때 허용 | false와 누락은 다름 |
| 본문 1,025바이트 | 413 | 전체 입력 크기를 먼저 제한 |
| 잘못된 UTF-8 원문 | 400 | 바이트 검증 실패 |
| 고립 서로게이트 이스케이프 | v1에서 U+FFFD로 치환될 수 있음 | UTF-16 이스케이프의 해석 규칙 |

고립 서로게이트는 바이트 자체로는 ASCII 문자로 표현될 수 있습니다. 따라서 원문 바이트가 UTF-8인지 확인하는 단계만으로 그 의미까지 검사하지는 않습니다. 반대로 정상적인 U+FFFD 문자는 그 자체로 잘못된 UTF-8이 아닙니다.

중복 키를 금지하려면 중복 이름을 검출하는 API·옵션 또는 토큰 검증 같은 추가 방법을 선택해야 합니다. 이 예제에서는 검증 범위를 과장하지 않고 현재 동작을 명시합니다. 새로운 JSON 구현을 도입하면 기존 클라이언트가 보내던 대소문자·중복·숫자 표현의 호환성도 점검합니다.

관련 공식 설명은 [encoding/json](https://pkg.go.dev/encoding/json)에 있습니다. 옵션 이름만 보고 모든 입력 조건을 추측하지 않고 실제 실패 사례를 테스트로 작성합니다.

## 문제 4. 여러 삽입을 같은 트랜잭션에 넣는다

저장 경로: [`atomic-import/main.go`](atomic-import/main.go).

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
	_ "modernc.org/sqlite"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.ExecContext(ctx, "CREATE TABLE imports(id INTEGER PRIMARY KEY,title TEXT NOT NULL UNIQUE)"); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}
func importTasks(ctx context.Context, db *sql.DB, titles []string) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(titles) < 1 || len(titles) > 100 {
		return errors.New("항목은 1~100개여야 합니다")
	}
	normalized := make([]string, len(titles))
	for i, title := range titles {
		value, err := todo.NormalizeTitle(title)
		if err != nil {
			return fmt.Errorf("항목 %d: %w", i+1, err)
		}
		normalized[i] = value
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, rollbackErr)
		}
	}()
	for _, title := range normalized {
		if _, err := tx.ExecContext(ctx, "INSERT INTO imports(title) VALUES (?)", title); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func count(ctx context.Context, db *sql.DB) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, "SELECT count(*) FROM imports").Scan(&n)
	return n, err
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := importTasks(ctx, db, []string{"기존"}); err != nil {
		return err
	}
	duplicateErr := importTasks(ctx, db, []string{"새 항목", "기존"})
	fmt.Println("중복 실패:", duplicateErr != nil)
	n, err := count(ctx, db)
	if err != nil {
		return err
	}
	fmt.Println("실패 뒤 개수:", n)
	if err := importTasks(ctx, db, []string{"Go", "SQL"}); err != nil {
		return err
	}
	n, err = count(ctx, db)
	if err != nil {
		return err
	}
	fmt.Println("성공 뒤 개수:", n)
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
go run ./solutions/30-todo-api/atomic-import
```

실행 결과:

```text
중복 실패: true
실패 뒤 개수: 1
성공 뒤 개수: 3
```

검증 코드: [`atomic-import/behavior_test.go`](atomic-import/behavior_test.go).

```go
package main

import (
	"context"
	"errors"
	"testing"
)

func TestImportAtomicity(t *testing.T) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := importTasks(ctx, db, []string{"existing"}); err != nil {
		t.Fatal(err)
	}
	if err := importTasks(ctx, db, []string{"new", "existing"}); err == nil {
		t.Fatal("duplicate accepted")
	}
	n, err := count(ctx, db)
	if err != nil || n != 1 {
		t.Fatalf("partial commit n=%d err=%v", n, err)
	}
	if err := importTasks(ctx, db, []string{"valid", " "}); err == nil {
		t.Fatal("invalid title accepted")
	}
	n, err = count(ctx, db)
	if err != nil || n != 1 {
		t.Fatal("invalid batch changed state")
	}
	if err := importTasks(ctx, db, []string{" Go ", "SQL"}); err != nil {
		t.Fatal(err)
	}
	n, err = count(ctx, db)
	if err != nil || n != 3 {
		t.Fatal("valid batch not stored")
	}
	for _, titles := range [][]string{nil, make([]string, 101)} {
		if err := importTasks(ctx, db, titles); err == nil {
			t.Fatal("accepted batch size")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := importTasks(canceled, db, []string{"canceled"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
```


### 입력 검증과 DB 실패를 나누기

제목 정규화는 DB에 쓰기 전에 끝낼 수 있습니다. 공백 제목이 포함된 경우에는 트랜잭션을 시작할 필요도 없습니다. 그러나 DB에 이미 같은 제목이 있는지는 SQL의 UNIQUE 제약이 최종적으로 확인합니다. 입력만 미리 검사한다고 모든 SQL 실패가 사라지는 것은 아닙니다.

`imports` 표는 이 문제만 사용하는 별도 표입니다. 기본 API의 todos는 같은 제목을 허용하므로, 문제의 고유성 정책을 기본 저장소의 계약이라고 설명하지 않습니다. 기능 확장 시 같은 이름의 자료라도 어떤 제약을 갖는지 명시해야 합니다.

### 중요한 실행 순서

1. 항목 개수를 확인합니다.
2. 모든 제목을 공통 규칙으로 정규화합니다.
3. BeginTx로 하나의 트랜잭션을 얻습니다.
4. 모든 INSERT를 `tx.ExecContext`로 실행합니다.
5. 중간 실패는 바로 반환하고 defer에서 Rollback합니다.
6. 전부 성공했을 때만 Commit합니다.

실패 사례에서는 첫 새 항목의 INSERT가 실행된 다음 두 번째 중복에서 실패합니다. 단순 반복 삽입이라면 첫 항목이 남을 수 있습니다. 테스트가 개수 1을 확인하는 이유는 실제 rollback을 검증하기 위해서입니다.

### ErrTxDone을 구분한다

Commit이 성공한 뒤 defer의 Rollback은 이미 끝난 트랜잭션이라는 오류를 반환할 수 있습니다. 이 값은 새 실패로 처리하지 않습니다. 반면 다른 Rollback 오류가 있다면 원래 작업 오류와 함께 보존합니다.

모든 INSERT에 `db.ExecContext`를 쓰면 tx를 열어도 그 쓰기들이 tx에 들어가지 않습니다. 특히 연결 풀을 1개로 제한한 상황에서 tx가 연결을 잡은 채 db 작업이 다른 연결을 기다리면 멈출 수 있습니다. 하나의 작업에 어떤 핸들을 사용하고 있는지 확인해야 합니다.

트랜잭션에 대한 공식 근거는 [Go 트랜잭션 안내](https://go.dev/doc/database/execute-transactions)입니다. 컨텍스트의 취소와 데이터베이스 트랜잭션의 확정 경계를 같은 개념으로 취급하지 않습니다.

## 문제 5. 세 종류의 실패를 나누는 클라이언트

저장 경로: [`create-client/main.go`](create-client/main.go).

```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

type APIError struct {
	Status int
	Code   string
}

func (e *APIError) Error() string { return fmt.Sprintf("API 상태 %d: %s", e.Status, e.Code) }
func createTask(ctx context.Context, client *http.Client, base, title string) (todo.Task, error) {
	title, err := todo.NormalizeTitle(title)
	if err != nil {
		return todo.Task{}, err
	}
	target, err := url.Parse(base)
	if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil || target.RawQuery != "" || target.Fragment != "" || (target.Path != "" && target.Path != "/") {
		return todo.Task{}, errors.New("서버 원점 URL이 필요합니다")
	}
	target.Path = "/todos"
	body, err := json.Marshal(struct {
		Title string `json:"title"`
	}{title})
	if err != nil {
		return todo.Task{}, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", target.String(), bytes.NewReader(body))
	if err != nil {
		return todo.Task{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return todo.Task{}, err
	}
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 4097))
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return todo.Task{}, err
	}
	if len(data) > 4096 {
		return todo.Task{}, errors.New("응답 크기 초과")
	}
	if response.StatusCode != 201 {
		var p struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(data, &p); err != nil || p.Error == "" {
			return todo.Task{}, &APIError{Status: response.StatusCode, Code: "invalid_error_response"}
		}
		return todo.Task{}, &APIError{Status: response.StatusCode, Code: p.Error}
	}
	var task todo.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return todo.Task{}, err
	}
	if task.ID <= 0 || task.Title != title || task.Done {
		return todo.Task{}, errors.New("생성 응답 계약 오류")
	}
	return task, nil
}
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
	task, err := createTask(ctx, client, server.URL, " Go 복습 ")
	if err != nil {
		return err
	}
	fmt.Printf("생성: %d %s 완료=%t\n", task.ID, task.Title, task.Done)
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
go run ./solutions/30-todo-api/create-client
```

실행 결과:

```text
생성: 1 Go 복습 완료=false
```

검증 코드: [`create-client/behavior_test.go`](create-client/behavior_test.go).

```go
package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func TestCreateClient(t *testing.T) {
	handler, err := todo.NewHandler(todo.NewMemory(), nil)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	client := server.Client()
	client.Timeout = time.Second
	defer client.CloseIdleConnections()
	task, err := createTask(context.Background(), client, server.URL, " Go ")
	if err != nil || task.Title != "Go" || task.ID != 1 {
		t.Fatalf("task=%+v %v", task, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := createTask(ctx, client, server.URL, "next"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestClientFailureContracts(t *testing.T) {
	cases := []struct {
		status    int
		body      string
		apiStatus int
	}{{500, `{"error":"internal_error"}`, 500}, {400, "broken", 400}, {201, "bad", 0}, {201, `{"id":0,"title":"Go","done":false}`, 0}, {201, strings.Repeat(" ", 4097), 0}}
	for _, tc := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = io.WriteString(w, tc.body)
		}))
		client := server.Client()
		client.Timeout = time.Second
		_, err := createTask(context.Background(), client, server.URL, "Go")
		client.CloseIdleConnections()
		server.Close()
		if err == nil {
			t.Fatal("accepted invalid response")
		}
		if tc.apiStatus != 0 {
			var api *APIError
			if !errors.As(err, &api) || api.Status != tc.apiStatus {
				t.Fatalf("error=%v", err)
			}
		}
	}
	if _, err := createTask(context.Background(), &http.Client{}, "ftp://example.invalid", "Go"); err == nil {
		t.Fatal("accepted URL")
	}
}
```


### 네트워크 실패와 API 실패

`client.Do`가 오류를 반환하면 요청을 수행하거나 응답을 얻는 경계에서 실패한 것입니다. 오류 없이 응답을 얻었더라도 상태가 500이면 API 기능은 실패입니다. 마지막으로 상태가 201이어도 본문이 잘못되거나 ID가 0이면 약속한 성공 응답이 아닙니다.

| 분류 | 예 | 반환 |
| --- | --- | --- |
| 요청 수행 실패 | 취소, 연결 오류 | 원래 오류 |
| API 실패 | 400·500 JSON 응답 | APIError |
| 성공 응답 형식 실패 | 깨진 JSON | 디코딩 오류 |
| 성공 응답 의미 실패 | ID 0, 다른 제목, done=true | 계약 오류 |
| 읽기 경계 실패 | 본문 초과·읽기·닫기 오류 | 해당 오류 |

상위 호출자는 `errors.As(err, &apiError)`로 상태와 코드를 사용할 수 있습니다. 오류 문자열을 잘라 상태 숫자를 추출하지 않습니다. 서버가 JSON 오류 형식을 지키지 않았을 때도 상태는 유지하고 코드에는 `invalid_error_response`를 사용합니다.

### 응답 크기와 정리

최대 4,096바이트를 허용하면서 정확히 그만큼만 읽으면, 뒤에 더 있는지 알 수 없습니다. 한 바이트를 더 읽어 상한 초과를 감지합니다. 응답은 크기를 확인하기 전에 먼저 닫기 결과까지 수집합니다. 어떤 분기로 반환하더라도 열린 Body가 남지 않게 합니다.

큰 오류 본문 전체를 오류 메시지에 붙이지 않습니다. 이 클라이언트는 API가 약속한 짧은 오류 코드와 상태만 전달합니다. 외부 API와 연결한다면 오류 코드 길이·문자 범위·로그 정책도 해당 계약에 맞춰 검토할 수 있습니다.

### 왜 자동으로 다시 보내지 않는가

요청이 서버에 도착해 저장까지 끝났지만 응답 전달만 실패했을 수 있습니다. 같은 POST를 그대로 다시 보내면 두 번째 항목이 생길 수 있습니다. 이번 함수에는 재시도 루프가 없습니다.

멱등성 키를 사용하는 설계를 추가하려면 서버가 키와 생성 결과를 일관되게 보관하고, 클라이언트가 같은 논리 요청에 같은 키를 사용해야 합니다. 클라이언트 쪽에서 기다렸다가 다시 호출하는 코드만 추가해서 해결되는 문제가 아닙니다.

## 문제 6. 전달된 context를 실제로 사용하게 한다

저장 경로: [`request-timeout/main.go`](request-timeout/main.go).

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func withDeadline(timeout time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type waitingRepository struct{ todo.Repository }

func (waitingRepository) Get(ctx context.Context, id int64) (todo.Task, error) {
	<-ctx.Done()
	return todo.Task{}, ctx.Err()
}
func run() error {
	handler, err := todo.NewHandler(waitingRepository{}, nil)
	if err != nil {
		return err
	}
	limited := withDeadline(time.Millisecond, handler)
	response := httptest.NewRecorder()
	limited.ServeHTTP(response, httptest.NewRequest("GET", "/todos/1", nil))
	fmt.Printf("상태: %d\n본문: %s\n", response.Code, strings.TrimSpace(response.Body.String()))
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
go run ./solutions/30-todo-api/request-timeout
```

실행 결과:

```text
상태: 504
본문: {"error":"deadline_exceeded"}
```

검증 코드: [`request-timeout/behavior_test.go`](request-timeout/behavior_test.go).

```go
package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func TestDeadlinePropagation(t *testing.T) {
	called := false
	handler := withDeadline(time.Second, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("missing deadline")
		}
		w.WriteHeader(204)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	if !called || response.Code != 204 {
		t.Fatal("handler not called")
	}
}
func TestCanceledAndExpiredMapping(t *testing.T) {
	handler, err := todo.NewHandler(waitingRepository{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		parentCanceled bool
		want           int
	}{{false, 504}, {true, 503}} {
		ctx, cancel := context.WithCancel(context.Background())
		if tc.parentCanceled {
			cancel()
		}
		response := httptest.NewRecorder()
		withDeadline(time.Millisecond, handler).ServeHTTP(response, httptest.NewRequest("GET", "/todos/1", nil).WithContext(ctx))
		cancel()
		if response.Code != tc.want {
			t.Fatalf("status=%d want=%d", response.Code, tc.want)
		}
	}
}
```


### context를 만들기만 해서는 충분하지 않다

`WithTimeout`으로 새 값을 만든 뒤 `next.ServeHTTP(w, r)`를 호출하면 다음 핸들러는 원래 context만 봅니다. `r.WithContext(ctx)`가 반환한 요청을 전달해야 합니다. 기존 요청의 다른 필드와 부모 취소 관계는 유지합니다.

시험용 저장소는 Get에서 ctx.Done을 기다립니다. 따라서 제한 시간이 지나면 반환할 수 있습니다. 이 대역은 Get 경로만 시험하기 위해 다른 인터페이스 메서드를 포함한 구조를 사용합니다. 실제 모든 경로를 제공하는 저장소로 쓰는 코드는 아닙니다.

### 작업 제한의 범위

이번 래퍼는 고루틴을 새로 만들어 작업을 강제로 버리지 않습니다. 다음 핸들러가 협조적으로 context를 관찰하는 동안 같은 호출 흐름에서 실행합니다. 저장소가 취소를 무시하면 이 래퍼만으로는 그 함수를 종료할 수 없습니다.

원래 요청이 이미 취소된 경우 `context.Canceled`, 자식의 제한 시간 때문에 끝난 경우 `context.DeadlineExceeded`가 전달됩니다. API는 각각 503과 504로 분류합니다. 실제 연결이 이미 끊겼으면 클라이언트가 이 응답을 보지 못할 수 있습니다.

`Body.Read`처럼 소켓 입력을 읽는 작업도 context 생성만으로 언제나 중단된다고 가정하지 않습니다. 실제 서버의 `ReadTimeout`과 요청 처리 수명은 서로 다른 경계입니다. 본문 서버는 둘을 따로 설정합니다.

### 테스트에서 시간을 정답으로 삼지 않는다

테스트는 정확히 몇 밀리초 걸렸는지 확인하지 않습니다. 마감이 존재하는지, 저장소가 반환한 오류가 올바르게 응답으로 분류되는지 확인합니다. 짧은 timeout은 자동으로 취소 사건을 만들기 위한 것이며 성능 측정 수치가 아닙니다.

## 문제 7. 타입 검사 뒤에도 계약 검사가 필요하다

| 질문 | 해설 |
| --- | --- |
| 메모리가 통과하면 SQLite 생략 가능한가? | 구현마다 정렬·NULL·Rows·연결 수명 등의 실패가 다르므로 같은 계약을 다시 확인합니다. |
| Repository 만족이 순서를 보장하는가? | 메서드 타입만 검사합니다. 정렬 의미는 구현과 테스트로 지킵니다. |
| 반환값을 바꾸는 테스트가 필요한가? | 내부 저장소를 외부에서 잠금 없이 바꾸지 못하게 값 소유권을 확인합니다. |
| 같은 핸들 조회로 파일 영속성이 충분한가? | 닫고 새 핸들로 여는 절차가 더 직접적인 증거입니다. |
| race 통과는 편집 충돌 방지인가? | 아닙니다. 개별 메모리 접근이 안전해도 후속 PUT이 이전 변경을 덮을 수 있습니다. |
| 취소는 완료한 저장을 복구하는가? | 아닙니다. 이미 확정한 쓰기와 응답 전달 실패는 구분해야 합니다. |

`var _ Repository = (*Memory)(nil)` 같은 선언은 구현이 인터페이스의 메서드 집합을 만족하는지 컴파일 때 확인합니다. 그것만으로 `List`가 오름차순인지, 없는 항목을 ErrNotFound로 돌려주는지는 검증하지 못합니다.

같은 테스트 함수를 저장소 팩터리에 적용하면 구현 간 차이를 찾기 쉽습니다. 공통 계약 테스트와 특정 구현 테스트를 함께 둡니다. 예를 들어 파일 재열기는 SQLite 전용이고, 최대 ID 발급 상태는 메모리 구현의 내부 경계 검사입니다.

## 선택 문제 해설. 마지막 쓰기와 충돌 감지는 다르다

A가 제목을 바꾼 뒤 B가 옛 제목을 포함한 전체 PUT을 보내면 B의 제목이 남습니다. B는 done만 바꾸려는 의도였지만 API는 전체 교체 요청을 받았습니다. 서버가 B의 의도를 자동으로 추론할 수는 없습니다.

버전 필드를 사용한다면 다음과 같은 새 계약을 정할 수 있습니다.

1. 조회 결과에 현재 버전을 포함합니다.
2. 수정 요청에 자신이 읽은 버전을 함께 보냅니다.
3. 저장소는 현재 버전이 일치할 때만 값과 버전을 한 번에 변경합니다.
4. 일치하지 않으면 충돌을 반환합니다.
5. 클라이언트는 새 상태를 읽어 사용자의 변경을 다시 판단합니다.

SQL에서는 조건부 UPDATE의 WHERE에 버전을 포함하고 영향받은 행 수를 확인할 수 있습니다. 메모리에서는 같은 잠금 안에서 버전 비교와 갱신을 수행해야 합니다. HTTP에서 사용할 409·412 등의 선택은 ETag·조건부 요청을 쓰는지에 따라 정해야 합니다. 이 선택 문제는 계약 설계를 위한 것이며 기본 API가 이미 이 기능을 제공한다는 뜻은 아닙니다.

## 전체 확인

```sh
go test -count=1 ./examples/30-todo-api/... ./solutions/30-todo-api/...
go vet ./examples/30-todo-api/... ./solutions/30-todo-api/...
go test -race ./examples/30-todo-api/... ./solutions/30-todo-api/...
```

구현이 한 줄 짧아졌다는 사실보다 실패 후 상태, 자원 정리, 취소, 순서와 같은 계약이 유지되는지 확인합니다. 실행하지 못한 운영체제나 외부 환경은 검증한 것으로 기록하지 않습니다.

[문제로 돌아가기](../../exercises/30-todo-api/README.md) · [본문으로 돌아가기](../../book/30-todo-api/README.md)
