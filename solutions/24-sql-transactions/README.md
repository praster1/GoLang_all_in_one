# 24장 해설: SQL과 트랜잭션

[문제](../../exercises/24-sql-transactions/README.md)를 먼저 풀고 예상한 저장 상태와 비교합니다. 모든 실행 명령은 저장소 루트 기준입니다. 전체 코드는 루트의 SQLite 드라이버 의존성을 사용하며, 기본 검증에 DB 서버·계정이 필요하지 않습니다.

이 장의 정답은 SQL 문장이 오류 없이 실행되는 데서 끝나지 않습니다. 값이 매개변수로 전달되는지, 오류의 원인이 남는지, 연결을 반환하는지, 실패한 변경이 남지 않는지를 함께 확인합니다.

## 문제 1. 트랜잭션 안과 복구 뒤

정답은 트랜잭션 안에서 7, Rollback 뒤 5입니다.

```text
트랜잭션 안: 7
복구 뒤: 5
```

전체 확인 프로그램입니다. 실행하려면 별도로 만든 개인 폴더 `_practice/24-sql-transactions/predict/main.go`에 저장합니다. 같은 이름의 기존 파일이 있다면 보존하고 다른 개인 폴더를 사용합니다.

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if _, err := db.ExecContext(ctx, "CREATE TABLE counter(value INTEGER);INSERT INTO counter VALUES(5)"); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE counter SET value=value+2"); err != nil {
		return err
	}
	var inside int
	if err := tx.QueryRowContext(ctx, "SELECT value FROM counter").Scan(&inside); err != nil {
		return err
	}
	if err := tx.Rollback(); err != nil {
		return err
	}
	var outside int
	if err := db.QueryRowContext(ctx, "SELECT value FROM counter").Scan(&outside); err != nil {
		return err
	}
	fmt.Println("트랜잭션 안:", inside)
	fmt.Println("복구 뒤:", outside)
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
go run ./_practice/24-sql-transactions/predict
```

Tx 안의 SELECT는 자신의 UPDATE 결과를 읽습니다. Rollback 뒤의 DB 조회는 취소된 변경을 보지 않습니다. Rollback을 Commit으로 바꾸고 Commit이 성공하면 두 번째 값도 7입니다. 이때 Commit 오류를 무시하지 않는 것까지가 변경 조건입니다.

| 관찰 | 의미 |
| --- | --- |
| UPDATE 성공 | 명령이 트랜잭션 안에서 수행됨 |
| Tx 조회 7 | 아직 확정하지 않은 변경을 포함한 조회 |
| Rollback 성공 | 이 트랜잭션의 변경 취소 |
| DB 조회 5 | 취소 뒤 남은 상태 |

`db`와 `tx`를 바꿔 써도 같은 결과라고 생각하면 안 됩니다. Tx가 단일 연결을 점유한 상태에서 db 조회를 요청하면 그 연결을 기다릴 수 있습니다. 같은 트랜잭션의 작업은 tx에 보내고, 트랜잭션을 끝낸 뒤 바깥 DB로 상태를 검사합니다. [공식 트랜잭션 안내](https://go.dev/doc/database/execute-transactions).

확인 프로그램의 `defer tx.Rollback()`은 예기치 않은 중간 반환을 위한 보조 정리입니다. 명시적 Rollback 뒤에는 이미 끝난 트랜잭션이므로 추가 변경을 하지 않습니다. 일반 구현은 아래 문제 5처럼 ErrTxDone을 예상된 종료 상태로 다루고 다른 정리 오류를 보존합니다.

## 문제 2. 제목 검색

참고 구현: [`note-search/main.go`](note-search/main.go).

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

type Note struct {
	ID    int64
	Title string
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE notes(id INTEGER PRIMARY KEY, title TEXT NOT NULL UNIQUE);
 INSERT INTO notes VALUES(1, 'Go''s SQL'),(2,'SQL')`)
	return err
}
func findByTitle(ctx context.Context, db *sql.DB, title string) (Note, error) {
	var note Note
	err := db.QueryRowContext(ctx, "SELECT id,title FROM notes WHERE title = ?", title).Scan(&note.ID, &note.Title)
	if err != nil {
		return Note{}, fmt.Errorf("find title: %w", err)
	}
	return note, nil
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := createSchema(ctx, db); err != nil {
		return err
	}
	note, err := findByTitle(ctx, db, "Go's SQL")
	if err != nil {
		return err
	}
	fmt.Printf("검색: %d, %s\n", note.ID, note.Title)
	_, missing := findByTitle(ctx, db, "' OR 1=1 --")
	fmt.Println("문자 그대로 검색:", errors.Is(missing, sql.ErrNoRows))
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
go run ./solutions/24-sql-transactions/note-search
```

```text
검색: 1, Go's SQL
문자 그대로 검색: true
```

### 값을 바꾸지 않고 정확히 전달하기

테이블 초기화 문자열 안의 `Go''s SQL`은 **고정 SQL 리터럴**에서 작은따옴표를 표현하는 방식입니다. 사용자 입력을 이런 문법으로 직접 바꿔 붙이는 함수는 만들지 않습니다. 검색할 실제 Go 문자열은 `Go's SQL`이고, QueryRowContext의 인자로 그대로 전달합니다.

`WHERE title = ?`는 제목이 정확히 같은 행을 선택합니다. 입력 `' OR 1=1 --`가 전체 행을 선택하는 조건으로 변하지 않습니다. 해당 제목을 저장하지 않았으므로 ErrNoRows입니다. UNIQUE 제약을 둔 이유도 제목 검색을 최대 한 행의 계약으로 만들기 위해서입니다.

`fmt.Errorf("find title: %w", err)`는 실패 위치를 설명하면서 원인을 보존합니다. 호출자는 `errors.Is`로 ErrNoRows와 context 취소를 구분할 수 있습니다. 모든 오류를 `false` 하나로 줄이면 저장소 고장도 없는 제목처럼 보이므로 여기서는 error를 유지합니다.

### 실제 DB로 검사하는 테스트

[`note-search/behavior_test.go`](note-search/behavior_test.go)의 전체 코드입니다.

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := createSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestSearchExactTitle(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	note, err := findByTitle(ctx, db, "Go's SQL")
	if err != nil || note.ID != 1 || note.Title != "Go's SQL" {
		t.Fatalf("note=%+v err=%v", note, err)
	}
	for _, title := range []string{"", "missing", "' OR 1=1 --", "SQL' --"} {
		if _, err := findByTitle(ctx, db, title); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("title=%q err=%v", title, err)
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := findByTitle(ctx, db, "SQL"); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
```

```sh
go test -v ./solutions/24-sql-transactions/note-search
```

SQL 모양 입력의 테스트는 공격용 외부 서버를 호출하지 않습니다. 로컬 DB에서 문자열이 데이터로 처리되는 계약을 확인합니다. 취소 테스트는 이미 취소한 context를 사용하여 대기 시간에 의존하지 않고 오류 전달을 확인합니다.

## 문제 3. Rows 수명 수정

문제의 함수는 첫 행 뒤에 읽을 두 번째 행이 있는데도 Rows를 닫지 않고 반환합니다. 결과 객체가 연결을 보유하므로 단일 연결 풀의 다음 조회가 기다립니다. 최대 연결 수를 늘리면 증상을 늦출 수 있어도 누락된 정리 책임은 그대로입니다.

수정된 전체 프로그램입니다. 개인 파일 `_practice/24-sql-transactions/repair/main.go`에 둡니다.

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"time"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func firstValue(ctx context.Context, db *sql.DB) (value int, err error) {
	rows, err := db.QueryContext(ctx, "SELECT 1 UNION ALL SELECT 2")
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, sql.ErrNoRows
	}
	if err := rows.Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	first, err := firstValue(ctx, db)
	if err != nil {
		return err
	}
	deadline, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var next int
	if err := db.QueryRowContext(deadline, "SELECT 9").Scan(&next); err != nil {
		return err
	}
	fmt.Printf("첫 행: %d, 다음 조회: %d\n", first, next)
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
go run ./_practice/24-sql-transactions/repair
```

```text
첫 행: 1, 다음 조회: 9
```

### 어느 반환 경로에서도 정리한다

| 경로 | 반환 직전 확인 | 정리 |
| --- | --- | --- |
| Query 실패 | 생성 실패 오류 | Rows가 없으므로 정리할 것 없음 |
| 첫 Next가 false, Err 있음 | 반복 오류 | defer Close |
| 첫 Next가 false, Err 없음 | ErrNoRows | defer Close |
| Scan 실패 | 변환 오류 | defer Close |
| 첫 행 성공 | 읽은 값 | defer Close |

함수는 첫 행만 필요하므로 나머지 행을 슬라이스에 모두 담을 필요가 없습니다. 대신 반환 전에 결과 자원을 끝내야 합니다. 이름 있는 `err`와 `errors.Join`은 정리 오류가 앞선 실패를 지우지 않도록 합니다. [Rows.Close](https://pkg.go.dev/database/sql#Rows.Close)와 [Rows.Err](https://pkg.go.dev/database/sql#Rows.Err)를 참고합니다.

다음 SELECT의 1초 마감은 실패한 실습이 영원히 멈추지 않도록 둔 보호 경계입니다. 정상 프로그램이 1초를 기다린다는 뜻은 아닙니다. 수정 뒤에는 앞선 함수가 연결을 이미 반납하므로 바로 사용할 수 있습니다.

단일 행만 필요하다면 실제 업무 코드에서는 QueryRowContext로 요구를 더 직접 표현할 수 있습니다. 이번 문제는 여러 행 결과의 조기 반환 경로를 학습하기 위해 QueryContext를 유지합니다.

## 문제 4. NULL과 빈 제목

참고 구현: [`nullable-note/main.go`](nullable-note/main.go).

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

type Note struct {
	ID    int64
	Title sql.NullString
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE notes(id INTEGER PRIMARY KEY,title TEXT);INSERT INTO notes VALUES(1,'Go'),(2,NULL),(3,'')`)
	return err
}
func noteLabel(ctx context.Context, db *sql.DB, id int64) (string, error) {
	var title sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT title FROM notes WHERE id = ?", id).Scan(&title); err != nil {
		return "", fmt.Errorf("note label: %w", err)
	}
	if !title.Valid {
		return "(미정)", nil
	}
	if title.String == "" {
		return "(빈 문자열)", nil
	}
	return title.String, nil
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := createSchema(ctx, db); err != nil {
		return err
	}
	for _, id := range []int64{1, 2, 3} {
		label, err := noteLabel(ctx, db, id)
		if err != nil {
			return err
		}
		fmt.Printf("%d: %s\n", id, label)
	}
	_, missing := noteLabel(ctx, db, 99)
	fmt.Println("행 없음:", errors.Is(missing, sql.ErrNoRows))
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
go run ./solutions/24-sql-transactions/nullable-note
```

```text
1: Go
2: (미정)
3: (빈 문자열)
행 없음: true
```

### 상태를 확인하는 순서

1. Scan 오류가 있으면 결과 문자열을 만들지 않습니다.
2. Valid가 false이면 SQL NULL이므로 `(미정)`입니다.
3. Valid가 true이고 String이 비었으면 `(빈 문자열)`입니다.
4. 나머지는 저장된 문자열 그대로입니다.

이 순서는 빈 문자열에 NULL의 의미를 억지로 싣지 않습니다. 표시 문구는 이 문제의 설계 선택입니다. 실제 API에서는 JSON null 또는 필드 생략 등 별도 표현을 정할 수 있지만, DB에서 읽는 단계부터 정보가 사라지지 않게 해야 합니다.

[`nullable-note/behavior_test.go`](nullable-note/behavior_test.go)의 전체 코드입니다.

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := createSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestThreeTitleStates(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	for _, tc := range []struct {
		id   int64
		want string
	}{{1, "Go"}, {2, "(미정)"}, {3, "(빈 문자열)"}} {
		got, err := noteLabel(ctx, db, tc.id)
		if err != nil || got != tc.want {
			t.Fatalf("id=%d got=%q err=%v", tc.id, got, err)
		}
	}
	if _, err := noteLabel(ctx, db, 99); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err=%v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := noteLabel(ctx, db, 1); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("closed database error=%v", err)
	}
}
```

```sh
go test -v ./solutions/24-sql-transactions/nullable-note
```

테스트의 마지막은 DB를 닫은 뒤 조회합니다. 이것은 없는 행이 아니라 사용할 수 없는 저장소입니다. 오류가 nil이 아니며 ErrNoRows와도 다르다는 조건을 확인합니다. 실제 오류 문자열 전체를 고정하지 않아 드라이버의 세부 문구에 기대지 않습니다.

## 문제 5. 원자적 포인트 이동

참고 구현: [`move-points/main.go`](move-points/main.go).

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

var ErrSource = errors.New("source missing or points insufficient")
var ErrDestination = errors.New("destination missing")
var ErrInput = errors.New("invalid transfer input")

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE wallets (id INTEGER PRIMARY KEY, points INTEGER NOT NULL CHECK(points >= 0));
        INSERT INTO wallets VALUES (1, 50), (2, 10)`)
	return err
}

func transfer(ctx context.Context, db *sql.DB, from, to int64, amount int) (err error) {
	if from <= 0 || to <= 0 || from == to || amount <= 0 || amount > 1000 {
		return ErrInput
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
	result, err := tx.ExecContext(ctx, "UPDATE wallets SET points = points - ? WHERE id = ? AND points >= ?", amount, from, amount)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrSource
	}
	result, err = tx.ExecContext(ctx, "UPDATE wallets SET points = points + ? WHERE id = ?", amount, to)
	if err != nil {
		return err
	}
	changed, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrDestination
	}
	return tx.Commit()
}

func points(ctx context.Context, db *sql.DB, id int64) (int, error) {
	var value int
	err := db.QueryRowContext(ctx, "SELECT points FROM wallets WHERE id = ?", id).Scan(&value)
	return value, err
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := createSchema(ctx, db); err != nil {
		return err
	}
	if err := transfer(ctx, db, 1, 2, 15); err != nil {
		return err
	}
	first, err := points(ctx, db, 1)
	if err != nil {
		return err
	}
	second, err := points(ctx, db, 2)
	if err != nil {
		return err
	}
	fmt.Printf("성공: %d, %d\n", first, second)
	moveErr := transfer(ctx, db, 1, 9, 5)
	fmt.Println("복구 필요:", errors.Is(moveErr, ErrDestination))
	first, err = points(ctx, db, 1)
	if err != nil {
		return err
	}
	second, err = points(ctx, db, 2)
	if err != nil {
		return err
	}
	fmt.Printf("유지: %d, %d\n", first, second)
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
go run ./solutions/24-sql-transactions/move-points
```

```text
성공: 35, 25
복구 필요: true
유지: 35, 25
```

### 검증, 변경, 확정의 순서

입력 계약 위반은 DB 작업 전에 ErrInput으로 반환합니다. 그 다음 BeginTx가 연결과 트랜잭션을 시작합니다. 차감 조건에는 잔여 포인트를 포함하고, UPDATE 결과가 한 행인지 확인합니다. 성공한 차감 뒤 목적지가 없으면 defer Rollback이 앞선 차감도 취소합니다.

| 시나리오 | 차감 | 증가 | 확정 | 최종 값 |
| --- | --- | --- | --- | --- |
| 15 이동 | 성공 | 성공 | Commit | 35, 25 |
| 50 이동 | 성공 | 성공 | Commit | 0, 60 |
| 51 이동 | 0행 | 실행 안 함 | Rollback | 50, 10 |
| 목적지 9 | 성공 | 0행 | Rollback | 50, 10 |
| 같은 ID | 실행 안 함 | 실행 안 함 | 시작 안 함 | 50, 10 |

성공 경로도 Commit에서 실패할 수 있으므로 그 오류를 반환합니다. `defer` 정리는 Commit 뒤 예상되는 ErrTxDone만 제외합니다. 한 번 끝난 Tx를 다시 업무 작업에 쓰지 않습니다.

`ErrSource`가 두 가지 상황을 함께 나타내는 것은 이 문제에서 명시한 계약입니다. 없는 출발 지갑과 포인트 부족을 반드시 분리해야 하는 제품이라면 오류 계약과 SQL 흐름을 함께 바꿔야 합니다. 테스트와 본문에 적지 않은 분류를 임의로 추가하지 않습니다.

### 저장된 상태까지 확인하는 테스트

[`move-points/behavior_test.go`](move-points/behavior_test.go)의 전체 코드입니다.

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := createSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestAtomicTransfer(t *testing.T) {
	for _, tc := range []struct {
		name          string
		from, to      int64
		amount        int
		wantErr       error
		first, second int
	}{
		{"normal", 1, 2, 15, nil, 35, 25}, {"empty source after move", 1, 2, 50, nil, 0, 60}, {"insufficient", 1, 2, 51, ErrSource, 50, 10},
		{"missing destination", 1, 9, 5, ErrDestination, 50, 10}, {"missing source", 9, 2, 5, ErrSource, 50, 10},
		{"same", 1, 1, 5, ErrInput, 50, 10}, {"zero", 1, 2, 0, ErrInput, 50, 10}, {"negative", 1, 2, -1, ErrInput, 50, 10},
		{"limit", 1, 2, 1001, ErrInput, 50, 10}, {"invalid id", 0, 2, 1, ErrInput, 50, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := testDB(t)
			ctx := context.Background()
			err := transfer(ctx, db, tc.from, tc.to, tc.amount)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err=%v", err)
			}
			first, err := points(ctx, db, 1)
			if err != nil {
				t.Fatal(err)
			}
			second, err := points(ctx, db, 2)
			if err != nil {
				t.Fatal(err)
			}
			if first != tc.first || second != tc.second || first+second != 60 {
				t.Fatalf("points=%d,%d", first, second)
			}
		})
	}
}
func TestCanceledTransfer(t *testing.T) {
	db := testDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := transfer(ctx, db, 1, 2, 15); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	got, err := points(context.Background(), db, 1)
	if err != nil || got != 50 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}
```

```sh
go test -v ./solutions/24-sql-transactions/move-points
```

각 사례는 새로운 메모리 DB를 받습니다. 앞 사례의 성공이 다음 사례의 초기 포인트를 바꾸지 않습니다. 모든 사례에서 두 지갑을 다시 조회하며, 합계 60도 확인합니다. 오류 이름만 비교하는 테스트보다 부분 반영을 직접 잡아낼 수 있습니다.

이 테스트는 학습용 정수 범위에서 원자성을 확인합니다. 실제 결제, 분산 작업, 요청 재시도에서 필요한 중복 방지 키와 외부 시스템까지의 원자성은 이 함수의 범위가 아닙니다. 로컬 트랜잭션이 외부 HTTP 호출까지 한 번에 되돌려준다고 생각하지 않습니다.

## 문제 6. 연결 대기의 마감

참고 구현: [`pool-deadline/main.go`](pool-deadline/main.go).

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"time"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE notes(id INTEGER PRIMARY KEY);INSERT INTO notes VALUES(1),(2)`)
	return err
}
func countNotes(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count)
	return count, err
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := createSchema(ctx, db); err != nil {
		return err
	}
	held, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	released := false
	defer func() {
		if !released {
			err = errors.Join(err, held.Close())
		}
	}()
	deadline, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, waitErr := countNotes(deadline, db)
	fmt.Println("대기 취소:", errors.Is(waitErr, context.DeadlineExceeded))
	if err := held.Close(); err != nil {
		return err
	}
	released = true
	count, err := countNotes(ctx, db)
	if err != nil {
		return err
	}
	fmt.Println("조회:", count)
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
go run ./solutions/24-sql-transactions/pool-deadline
```

```text
대기 취소: true
조회: 2
```

### context를 전달하는 함수가 작아도 충분하다

`countNotes`는 짧지만 중요한 계약을 갖습니다. 호출자가 전달한 context로 QueryRowContext를 실행하고 Scan 오류를 그대로 돌려줍니다. 내부에서 Background를 새로 만들거나 context 오류를 무시하는 재시도를 추가하지 않습니다.

유일한 연결을 `held`가 갖고 있으므로 조회는 연결 풀에서 기다립니다. 마감으로 기다리는 호출만 종료되어도 `held`의 소유권은 자동으로 바뀌지 않습니다. `held.Close()`로 전용 사용을 끝내고, 새 작업용 context로 정상 조회합니다.

[`pool-deadline/behavior_test.go`](pool-deadline/behavior_test.go)의 전체 코드입니다.

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := createSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestPoolDeadlineReleasesWaiter(t *testing.T) {
	db := testDB(t)
	held, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := countNotes(ctx, db); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	if db.Stats().WaitCount < 1 {
		t.Fatal("query did not wait for pool")
	}
	if err := held.Close(); err != nil {
		t.Fatal(err)
	}
	live, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if got, err := countNotes(live, db); err != nil || got != 2 {
		t.Fatalf("count=%d err=%v", got, err)
	}
}
func TestEmptyTableAndCanceledCall(t *testing.T) {
	db := testDB(t)
	if _, err := db.ExecContext(context.Background(), "DELETE FROM notes"); err != nil {
		t.Fatal(err)
	}
	if got, err := countNotes(context.Background(), db); err != nil || got != 0 {
		t.Fatalf("count=%d err=%v", got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := countNotes(ctx, db); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
```

```sh
go test -v ./solutions/24-sql-transactions/pool-deadline
```

WaitCount는 실제 연결 대기를 관찰하기 위한 풀 통계입니다. 테스트가 단지 이미 만료된 context를 즉시 거절한 사례로 바뀌지 않았는지 확인합니다. 다만 10ms라는 짧은 시간이 정확히 지켜졌는지 성능 수치로 검사하지 않습니다.

빈 표의 `COUNT(*)`는 결과 행이 없다는 뜻이 아닙니다. 집계 결과를 담은 한 행의 값이 0입니다. 이 차이를 ErrNoRows와 구분하는 것도 마지막 테스트의 목적입니다.

## 선택 A. Tx에 준비된 문장 두기

전체 프로그램을 `_practice/24-sql-transactions/prepared-batch/main.go`에 작성합니다.

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func addTitles(ctx context.Context, db *sql.DB, titles []string) (err error) {
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
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO notes(title) VALUES (?)")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, stmt.Close()) }()
	for _, title := range titles {
		if _, err := stmt.ExecContext(ctx, title); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if _, err := db.ExecContext(ctx, "CREATE TABLE notes(id INTEGER PRIMARY KEY,title TEXT NOT NULL UNIQUE)"); err != nil {
		return err
	}
	if err := addTitles(ctx, db, []string{"Go", "SQL"}); err != nil {
		return err
	}
	batchErr := addTitles(ctx, db, []string{"HTTP", "Go"})
	fmt.Println("중복 실패:", batchErr != nil)
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count); err != nil {
		return err
	}
	fmt.Println("유지된 행:", count)
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
go run ./_practice/24-sql-transactions/prepared-batch
```

```text
중복 실패: true
유지된 행: 2
```

### 준비된 문장과 트랜잭션은 다른 역할이다

Prepare는 같은 SQL 구조를 재사용하게 합니다. 여러 실행을 하나의 원자적 변경으로 묶는 책임은 Tx에 있습니다. DB에서 Prepare한 문장을 여러 번 실행하는 것만으로 전체 묶음이 자동 rollback되지는 않습니다.

이 구현은 `tx.PrepareContext`로 문장을 만들고 반복합니다. 두 번째 묶음에서 HTTP 삽입 뒤 Go가 UNIQUE 제약을 위반하면 함수가 Commit 전에 반환합니다. defer 순서로 Statement를 정리하고 Tx를 rollback하므로 HTTP도 남지 않습니다.

| 묶음 | 첫 삽입 | 다음 삽입 | 최종 상태 |
| --- | --- | --- | --- |
| Go, SQL | Go 성공 | SQL 성공 | 두 행 확정 |
| HTTP, Go | HTTP 임시 성공 | Go 중복 실패 | 기존 두 행 유지 |
| 빈 슬라이스 | 실행 없음 | 실행 없음 | 정상 종료, 변화 없음 |

Statement를 닫는 defer가 Rollback defer보다 나중에 등록되었으므로 먼저 실행됩니다. 이미 Commit이 트랜잭션 문장을 정리한 정상 경로에서도 Close는 안전하게 사용됩니다. [Stmt.Close](https://pkg.go.dev/database/sql#Stmt.Close)와 [Tx.PrepareContext](https://pkg.go.dev/database/sql#Tx.PrepareContext)를 참고합니다.

추가 테스트에서는 빈 묶음, 첫 원소 중복, 두 번째 원소 중복, 따옴표 제목을 검사할 수 있습니다. 행 수만 같고 내용이 바뀐 버그를 피하려면 제목을 정렬해 다시 읽어 Go와 SQL이 남았는지도 확인합니다.

## 검증과 완료 기준

```sh
go fmt ./examples/24-sql-transactions/... ./solutions/24-sql-transactions/...
go vet ./examples/24-sql-transactions/... ./solutions/24-sql-transactions/...
go test -count=1 ./examples/24-sql-transactions/... ./solutions/24-sql-transactions/...
go build ./examples/24-sql-transactions/... ./solutions/24-sql-transactions/...
```

지원되는 환경에서는 `go test -race`도 같은 패키지 범위에 실행합니다. race detector가 논리적 SQL 원자성이나 모든 DB 장애를 대신 확인하지는 않습니다. 저장 상태 검사는 별도 테스트로 계속 필요합니다.

이 장을 마쳤다면 연결을 여는 위치, 각 결과를 닫는 위치, Tx의 확정 경계, context의 시작과 끝을 코드에서 표시할 수 있어야 합니다. 반환값과 저장된 상태가 모두 계약과 맞는지 확인한 뒤 해설과의 비교를 마칩니다.

[본문](../../book/24-sql-transactions/README.md) · [문제](../../exercises/24-sql-transactions/README.md) · [공식 database/sql 문서](https://pkg.go.dev/database/sql)
