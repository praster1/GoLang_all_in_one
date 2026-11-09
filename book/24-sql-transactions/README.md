# 24. SQL과 트랜잭션

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 중급 | 오류와 defer, context, 인터페이스, [23장](../23-api-testing/README.md) | Go, 이 저장소, 처음 한 번 내려받는 SQLite 드라이버 |

23장에서는 저장소 인터페이스 뒤에 메모리 구현과 실패를 돌려주는 구현을 연결했습니다. 이번 장에서는 실제 데이터베이스에 명령을 보내고 결과를 Go 값으로 읽습니다. 같은 함수가 성공하더라도, 여러 변경 중 일부만 남으면 기능 전체는 실패할 수 있습니다. **연결의 수명, 결과를 읽는 수명, 변경을 확정하는 수명**을 구분하는 것이 중심 목표입니다.

별도 데이터베이스 서버와 계정은 필요하지 않습니다. 기본 예제는 프로세스 안에서 동작하는 SQLite를 사용하며, 마지막 예제만 자신이 만든 임시 폴더에 파일을 저장한 뒤 지웁니다. SQL 문장을 화면에 출력하는 모의 예제가 아니라 실제 SQLite에 실행하고 다시 조회합니다.

## 학습 목표

- `database/sql`과 드라이버의 역할을 구분하고 연결을 확인한다.
- 사용자 값을 SQL 문장과 분리하여 전달하고 결과 행을 Go 값으로 읽는다.
- 결과 없음, NULL, 빈 문자열과 실행 오류를 구분한다.
- 여러 변경을 트랜잭션으로 묶고 실패 뒤 원래 값이 유지되는지 확인한다.
- 연결 풀 대기에도 context가 필요한 이유와 파일 DB의 수명을 설명한다.

## 사전 지식과 준비

모든 명령은 **저장소 루트 `GoLang_all_in_one/`**에서 실행합니다. 하위 폴더에 `go.mod`를 만들지 않습니다. 이 장은 루트에 등록된 **`modernc.org/sqlite v1.62.0`**을 사용합니다. 이 드라이버는 CGo 없는 SQLite 구현을 `database/sql`에 연결합니다. 버전과 설치된 의존성은 [`go.mod`](../../go.mod)·[`go.sum`](../../go.sum), 지원 환경은 [드라이버 공식 문서](https://pkg.go.dev/modernc.org/sqlite)에서 확인합니다. 확인일: **2026-11-09**.

```sh
go version
go mod download
go run ./examples/24-sql-transactions/open-ping
```

`go mod download`는 의존성을 로컬 캐시에 준비합니다. 처음에는 모듈 다운로드를 위한 네트워크가 필요할 수 있습니다. 캐시가 준비된 뒤의 예제·테스트는 외부 서버나 계정에 접속하지 않습니다. 새 프로젝트에 같은 드라이버를 도입한다면 그 프로젝트에서 버전을 고정해 추가해야 하지만, 이 저장소에서는 이미 등록된 의존성을 그대로 사용합니다.

| 실행 예제 | 중심 내용 | 명령 |
| --- | --- | --- |
| [open-ping](../../examples/24-sql-transactions/open-ping/main.go) | DB 핸들과 연결 확인 | `go run ./examples/24-sql-transactions/open-ping` |
| [parameterized](../../examples/24-sql-transactions/parameterized/main.go) | 값의 매개변수 전달 | `go run ./examples/24-sql-transactions/parameterized` |
| [query-rows](../../examples/24-sql-transactions/query-rows/main.go) | 목록, Scan, Close, Err | `go run ./examples/24-sql-transactions/query-rows` |
| [null-not-found](../../examples/24-sql-transactions/null-not-found/main.go) | NULL과 없는 행 | `go run ./examples/24-sql-transactions/null-not-found` |
| [transactions](../../examples/24-sql-transactions/transactions/main.go) | commit과 rollback | `go run ./examples/24-sql-transactions/transactions` |
| [pool-cancellation](../../examples/24-sql-transactions/pool-cancellation/main.go) | 연결 대기와 취소 | `go run ./examples/24-sql-transactions/pool-cancellation` |
| [file-reopen](../../examples/24-sql-transactions/file-reopen/main.go) | 파일 DB를 닫고 다시 열기 | `go run ./examples/24-sql-transactions/file-reopen` |

각 예제는 독립 프로그램이며 자동으로 종료됩니다. 실제 운영 데이터와 기존 DB 파일을 사용하지 않습니다. 여러 예제의 `main`을 같은 폴더에 합치지 않습니다.

## 1. 데이터베이스, 드라이버, 연결 풀

### 1.1 데이터를 표로 저장한다

관계형 데이터베이스는 데이터를 테이블(table)의 행(row)과 열(column)로 다룹니다. 아래 학습 표에서 한 행은 하나의 학습 항목이고, `id`, `title`, `minutes`는 열 이름입니다. SQL은 이런 표의 구조를 만들고 데이터를 추가·검색·변경하는 언어입니다.

| id | title | minutes |
| --- | --- | --- |
| 1 | Go | 25 |
| 2 | 복습 | 15 |
| 3 | SQL | 40 |

이 장에서 사용하는 SQL의 기본 역할은 다음과 같습니다. SQL 전체 문법을 한 번에 익히기보다, 각 명령의 **입력과 결과**를 Go 함수에 연결해 읽습니다.

| 문장 | 역할 | Go에서 주로 사용하는 메서드 |
| --- | --- | --- |
| `CREATE TABLE` | 열과 제약을 선언 | `ExecContext` |
| `INSERT INTO` | 새 행 추가 | `ExecContext` |
| `SELECT` | 조건에 맞는 열 조회 | `QueryContext`, `QueryRowContext` |
| `UPDATE` | 기존 행 변경 | `ExecContext` |
| `DELETE FROM` | 행 삭제 | `ExecContext` |

`PRIMARY KEY`는 행을 식별하는 키입니다. `NOT NULL`은 NULL 저장을 막고, `UNIQUE`는 중복을 제한하며, `CHECK`는 조건을 검사합니다. 이 장에서는 SQLite의 `INTEGER PRIMARY KEY`를 식별자로 사용합니다. 다른 DB의 자동 번호 문법까지 같다고 가정하지 않습니다. [SQLite CREATE TABLE](https://sqlite.org/lang_createtable.html)을 참고합니다.

### 1.2 database/sql은 공통 API다

`database/sql`은 SQL 문장을 실행하고 결과를 읽는 공통 인터페이스를 제공합니다. 실제 DB에 맞는 연결과 실행은 **드라이버(driver)**가 담당합니다. 표준 라이브러리만 import한다고 특정 데이터베이스 드라이버가 자동으로 포함되지는 않습니다. [database/sql 개요](https://pkg.go.dev/database/sql#pkg-overview)를 참고합니다.

이번 예제의 다음 import는 패키지에서 공개한 이름을 직접 사용하지 않지만 드라이버 등록을 실행하게 합니다. 이런 import를 **부수 효과를 위한 import**라고 합니다.

```go
// import 목록의 발췌
_ "modernc.org/sqlite"
```

`sql.Open("sqlite", ":memory:")`의 첫 문자열은 등록된 드라이버 이름입니다. 두 번째 문자열은 연결 대상과 옵션을 담는 **데이터 소스 이름(DSN)**입니다. 여기서는 SQLite가 특별하게 해석하는 메모리 DB 이름입니다. 비밀번호가 들어 있는 DSN을 그대로 로그에 쓰는 습관을 만들지 않습니다.

### 1.3 *sql.DB는 하나의 연결이 아니다

`*sql.DB`는 연결을 얻고 돌려받는 **연결 풀(connection pool)**을 관리합니다. 고루틴 여러 개가 같은 `*sql.DB`를 사용할 수 있으며, 요청마다 `sql.Open`과 `Close`를 반복하는 방식으로 설계하지 않습니다. 애플리케이션 시작에서 열고 종료에서 닫는 수명이 기본입니다. [연결 관리](https://go.dev/doc/database/manage-connections)를 참고합니다.

다만 이 장의 메모리 SQLite는 연결 수를 1로 제한합니다. SQLite의 `:memory:`는 **물리 연결마다 별개의 DB**이기 때문입니다. 연결 A에서 만든 표를 연결 B에서 찾으면 같은 표를 볼 수 없습니다. 최대 연결 1, 유휴 연결 1을 유지하여 한 예제 안의 명령들이 같은 DB를 사용하게 합니다. [SQLite 메모리 DB](https://sqlite.org/inmemorydb.html).

이 설정을 모든 DB의 최적값으로 외우지 않습니다. 풀 크기는 DB 종류, 작업과 측정 결과에 맞춰 결정합니다. 메모리 연결을 닫거나 수명이 만료되어 새 연결로 바뀌면 데이터도 사라집니다. 따라서 이 예제에는 연결 수명 만료 설정을 추가하지 않습니다.

## 2. 열기, 확인하기, 닫기

저장 경로: [`examples/24-sql-transactions/open-ping/main.go`](../../examples/24-sql-transactions/open-ping/main.go).

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
	var value int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&value); err != nil {
		return err
	}
	fmt.Println("연결:", true)
	fmt.Println("확인 값:", value)
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
go run ./examples/24-sql-transactions/open-ping
```

```text
연결: true
확인 값: 1
```

### 2.1 Open과 Ping을 구분한다

`sql.Open`이 성공했다는 사실만으로 실제 연결을 사용할 수 있다고 단정하지 않습니다. 드라이버에 따라 Open은 DSN을 확인하거나 핸들만 준비하고, 실제 연결은 다음 작업까지 미룰 수 있습니다. `PingContext`는 연결을 얻어 사용할 수 있는지 확인합니다. [Open](https://pkg.go.dev/database/sql#Open)과 [PingContext](https://pkg.go.dev/database/sql#DB.PingContext)의 계약을 구분합니다.

| 단계 | 소유하는 자원 | 실패하면 할 일 |
| --- | --- | --- |
| `sql.Open` | 성공하면 DB 핸들 | 오류를 반환 |
| 풀 설정 | 같은 DB 핸들 | 이번 예제의 메모리 연결 정책 적용 |
| `PingContext` | 확인 중 실제 연결 | DB도 닫고 오류 반환 |
| `QueryRowContext().Scan` | 한 행 결과 | 실행·변환 오류 반환 |
| `run` 종료 | DB 전체 | `Close` 결과도 확인 |

`SELECT 1`은 표 없이 정수 한 값을 조회합니다. `Scan(&value)`는 해당 열의 값을 `value`가 있는 곳에 기록합니다. `&`가 필요한 이유는 08장에서 배운 포인터와 같습니다. 이 예제는 첫 SQL을 실행하는 흐름에 집중하므로 데이터 표를 아직 만들지 않습니다.

### 2.2 종료 오류를 잃지 않는다

`run() (err error)`는 이름 있는 오류 반환값을 사용합니다. `defer` 안에서 `db.Close()`의 오류와 원래 오류를 `errors.Join`으로 묶어, 앞선 작업의 실패를 종료 오류로 덮지 않습니다. 모든 오류가 nil이면 결과도 nil입니다. 이름 있는 반환값을 택한 이유가 바로 이 정리 책임입니다.

`main`은 실패 메시지를 표준 오류에 기록하고 비정상 종료합니다. 종료 처리는 `run`의 defer가 끝난 뒤 실행되므로 자원 정리가 먼저 이뤄집니다. `run` 안에서 곧바로 `os.Exit`하면 defer를 건너뛰므로 사용하지 않습니다.

작은 예제라도 “연결 실패”, “쿼리 실패”, “정리 실패”가 모두 같은 nil로 사라지지 않아야 합니다. HTTP API에서는 이 오류를 그대로 고객에게 노출하지 않고, 23장에서 배운 오류 분류와 기록 정책을 적용합니다.

## 3. SQL 문장과 값을 분리하기

저장 경로: [`examples/24-sql-transactions/parameterized/main.go`](../../examples/24-sql-transactions/parameterized/main.go).

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

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT NOT NULL)")
	return err
}

func addNote(ctx context.Context, db *sql.DB, title string) (int64, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO notes(title) VALUES (?)", title)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
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
	for _, title := range []string{"Go's SQL", "x'); DROP TABLE notes; --"} {
		id, err := addNote(ctx, db, title)
		if err != nil {
			return err
		}
		fmt.Printf("저장: %d, %s\n", id, title)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count); err != nil {
		return err
	}
	fmt.Println("행 수:", count)
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
go run ./examples/24-sql-transactions/parameterized
```

```text
저장: 1, Go's SQL
저장: 2, x'); DROP TABLE notes; --
행 수: 2
```

### 3.1 물음표는 값이 들어갈 자리다

`INSERT INTO notes(title) VALUES (?)`는 실행할 SQL 문장입니다. `title`은 그 다음 인자로 따로 전달합니다. 드라이버는 값을 SQL 문법으로 다시 이어 붙이지 않고 매개변수(parameter) 값으로 처리합니다. 그래서 작은따옴표가 포함된 제목도 그대로 저장됩니다. [SQL 주입 피하기](https://go.dev/doc/database/sql-injection)를 참고합니다.

두 번째 제목은 SQL처럼 보이지만, 이 프로그램에서는 한 개의 문자열 데이터입니다. 표를 삭제하지 않습니다. 검증 테스트는 단순히 프로그램이 오류 없이 끝나는지만 보지 않고, 각 제목을 다시 읽어 입력과 같으며 표에 네 행이 남는지 확인합니다.

**잘못된 방식의 발췌**는 다음과 같습니다. 입력을 직접 이어 붙이면 문자열의 일부가 SQL 문법으로 해석될 수 있습니다. 개인 실습에서도 이 방식을 정답으로 사용하지 않습니다.

```go
// 의도적으로 잘못된 SQL 구성: 실행 예제에 포함하지 않는다.
query := "INSERT INTO notes(title) VALUES ('" + title + "')"
```

문자열에서 `'`를 임의로 지우는 것도 해결책이 아닙니다. 정상적인 `Go's SQL`까지 바뀌고, DB별 문법을 애플리케이션이 잘못 흉내 내게 됩니다. 값을 전달하는 API를 사용하면 데이터의 원래 내용과 명령의 구조를 모두 보존할 수 있습니다.

### 3.2 매개변수가 대신하지 않는 것

물음표는 **값**에 사용합니다. 테이블명, 열 이름, `ASC`·`DESC` 같은 SQL 구조를 임의 입력으로 받아 같은 방식으로 바꿀 수는 없습니다. 정렬 선택을 받는다면 허용한 선택을 코드 안의 고정 SQL로 대응시킵니다. 매개변수 사용이 업무 규칙 검증까지 대신하지도 않습니다.

이번 `addNote`는 빈 제목도 저장합니다. 이것은 SQLite와 매개변수 전달을 확인하는 좁은 예제의 계약입니다. 실제 학습 API에서는 빈 제목·길이·정규화 규칙을 먼저 검사한 뒤 저장하도록 설계합니다. 입력 검증과 SQL 매개변수는 서로 다른 책임입니다.

### 3.3 Exec 결과도 확인한다

`ExecContext`는 결과 행을 반환하지 않는 명령에 사용합니다. 반환된 `sql.Result`에서 `LastInsertId`나 `RowsAffected`를 조회할 수 있습니다. 지원 여부와 구체적인 의미는 드라이버에 따라 다를 수 있으므로 이 메서드의 오류도 확인합니다. [Result](https://pkg.go.dev/database/sql#Result)를 참고합니다.

예제의 ID 1과 2는 **새 메모리 DB**에서 SQLite가 생성한 결과입니다. 이미 데이터가 있는 DB에서 항상 1부터 다시 시작한다고 가정하지 않습니다. 삭제된 번호의 재사용 정책과 충돌 처리도 별도 계약입니다. 30장에서는 API의 식별자 정책을 저장소 두 구현에 똑같이 적용합니다.

## 4. 여러 행을 안전하게 읽기

저장 경로: [`examples/24-sql-transactions/query-rows/main.go`](../../examples/24-sql-transactions/query-rows/main.go).

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

type Lesson struct {
	ID      int64
	Title   string
	Minutes int
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE lessons (id INTEGER PRIMARY KEY, title TEXT NOT NULL, minutes INTEGER NOT NULL);
        INSERT INTO lessons VALUES (2, '복습', 15), (1, 'Go', 25), (3, 'SQL', 40)`)
	return err
}

func listLessons(ctx context.Context, db *sql.DB, minimum int) (lessons []Lesson, err error) {
	rows, err := db.QueryContext(ctx, "SELECT id, title, minutes FROM lessons WHERE minutes >= ? ORDER BY id", minimum)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	lessons = make([]Lesson, 0)
	for rows.Next() {
		var lesson Lesson
		if err := rows.Scan(&lesson.ID, &lesson.Title, &lesson.Minutes); err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lessons, nil
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
	lessons, err := listLessons(ctx, db, 20)
	if err != nil {
		return err
	}
	for _, lesson := range lessons {
		fmt.Printf("%d: %s, %d분\n", lesson.ID, lesson.Title, lesson.Minutes)
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
go run ./examples/24-sql-transactions/query-rows
```

```text
1: Go, 25분
3: SQL, 40분
```

### 4.1 조회와 반복의 네 단계

`QueryContext`가 반환한 `*sql.Rows`는 결과를 차례로 읽는 객체입니다. 결과 전체가 이미 완성된 슬라이스라고 생각하면 정리 시점을 놓치기 쉽습니다. [여러 행 조회](https://go.dev/doc/database/querying)와 [Rows](https://pkg.go.dev/database/sql#Rows)의 흐름을 따릅니다.

| 단계 | 질문 | 코드 |
| --- | --- | --- |
| 시작 | 쿼리를 시작할 수 있는가? | `rows, err := db.QueryContext(...)` |
| 이동 | 읽을 다음 행이 있는가? | `rows.Next()` |
| 변환 | 열들을 Go 값에 담을 수 있는가? | `rows.Scan(...)` |
| 완료 | 정상적으로 끝났는가? | `rows.Err()` |

`Next`가 false라는 사실만으로 정상 종료라고 판단하지 않습니다. 결과를 다 읽었을 수도 있고 읽는 중 오류가 생겼을 수도 있습니다. 반복 뒤 `rows.Err()`를 확인해 두 경우를 구분합니다. `Scan` 자체의 타입 변환 오류도 그 자리에서 확인합니다.

SQL에서 선택한 열은 `id, title, minutes`입니다. `Scan` 인자는 같은 순서로 `&lesson.ID, &lesson.Title, &lesson.Minutes`를 줍니다. 구조체 필드 이름을 보고 드라이버가 자동으로 찾아주는 방식이 아닙니다. 열을 추가하거나 순서를 바꿨다면 Scan도 함께 검토합니다.

### 4.2 중간에 나가도 Close한다

`defer rows.Close()`에 해당하는 정리 코드를 쿼리 성공 직후 등록합니다. 끝까지 읽은 정상 경로뿐 아니라 Scan 실패, 첫 행만 읽고 반환하는 경로에서도 자원을 반납하기 위해서입니다. 이 예제는 Close 오류도 반환 오류와 합칩니다.

연결이 하나뿐일 때 Rows를 열어 둔 채 같은 `db`로 추가 작업을 요청하면, 추가 작업은 이미 사용 중인 그 연결이 반환되기를 기다릴 수 있습니다. 그 작업이 끝나야 Close하는 구조라면 서로 기다리는 문제가 생깁니다. **필요한 값을 읽고 Rows를 닫은 뒤 다음 작업을 시작하는 함수 경계**가 이 문제를 줄입니다.

### 4.3 순서는 SQL에 적는다

삽입은 ID 2, 1, 3 순서로 했지만 조회에는 `ORDER BY id`를 적었습니다. 출력 순서는 이 계약으로 정해집니다. ORDER BY 없는 결과가 우연히 원하는 순서로 보였더라도 그것을 목록 API의 약속으로 사용하지 않습니다. [SQLite SELECT와 ORDER BY](https://sqlite.org/lang_select.html)를 참고합니다.

| 최소 학습 분 | 선택되는 ID | 결과 개수 |
| --- | --- | --- |
| 0 | 1, 2, 3 | 3 |
| 20 | 1, 3 | 2 |
| 40 | 3 | 1 |
| 41 | 없음 | 0 |

빈 목록은 이 함수에서 성공한 빈 슬라이스입니다. 여러 행 검색의 “일치한 행 0개”를 곧바로 `sql.ErrNoRows`로 바꾸지 않습니다. 한 행을 기대하는 QueryRow의 계약과 구별합니다.

## 5. NULL, 빈 문자열, 없는 행

저장 경로: [`examples/24-sql-transactions/null-not-found/main.go`](../../examples/24-sql-transactions/null-not-found/main.go).

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
	_, err := db.ExecContext(ctx, `CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT);
        INSERT INTO notes VALUES (1, 'Go'), (2, NULL), (3, '')`)
	return err
}

func findNote(ctx context.Context, db *sql.DB, id int64) (Note, error) {
	var note Note
	err := db.QueryRowContext(ctx, "SELECT id, title FROM notes WHERE id = ?", id).Scan(&note.ID, &note.Title)
	if err != nil {
		return Note{}, fmt.Errorf("find note: %w", err)
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
	for _, id := range []int64{1, 2, 3} {
		note, err := findNote(ctx, db, id)
		if err != nil {
			return err
		}
		if note.Title.Valid {
			fmt.Printf("%d: %q\n", note.ID, note.Title.String)
		} else {
			fmt.Printf("%d: NULL\n", note.ID)
		}
	}
	_, lookupErr := findNote(ctx, db, 99)
	fmt.Println("없음:", errors.Is(lookupErr, sql.ErrNoRows))
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
go run ./examples/24-sql-transactions/null-not-found
```

```text
1: "Go"
2: NULL
3: ""
없음: true
```

### 5.1 행이 있는데 값이 없을 수 있다

`sql.NullString`에는 `String`과 `Valid`가 있습니다. `Valid`가 false이면 SQL NULL입니다. `String == ""`만 검사하면 빈 문자열과 NULL을 구분하지 못합니다. [NullString](https://pkg.go.dev/database/sql#NullString)의 두 필드를 함께 읽습니다.

| 저장된 상태 | 행 존재 | Valid | String | 이번 출력 |
| --- | --- | --- | --- | --- |
| `'Go'` | 있음 | true | `Go` | `"Go"` |
| `NULL` | 있음 | false | 의미 있는 값으로 사용하지 않음 | `NULL` |
| `''` | 있음 | true | 빈 문자열 | `""` |
| ID 99 | 없음 | 결과를 사용하지 않음 | 결과를 사용하지 않음 | `sql.ErrNoRows` |

NULL을 “빈 문자열의 다른 표기”로 설명하면 실제 데이터의 의미가 사라집니다. 미정, 아직 입력하지 않음, 적용되지 않음 중 무엇인지는 서비스 계약에서 정해야 합니다. 이 예제는 상태를 읽고 구분하는 데까지만 책임집니다.

### 5.2 QueryRow의 오류는 Scan에서 받는다

`QueryRowContext`는 `(*Row, error)` 쌍을 반환하지 않습니다. 실행 오류와 결과 없음은 `Scan` 호출에서 전달됩니다. 그래서 `db.QueryRowContext(...).Scan(...)`의 반환 오류를 확인합니다. [QueryRowContext](https://pkg.go.dev/database/sql#DB.QueryRowContext).

`findNote`는 `%w`로 조회 맥락을 더합니다. 호출자는 `errors.Is(err, sql.ErrNoRows)`로 감싼 오류 안의 원인을 구별할 수 있습니다. 문자열에 `no rows`가 포함되어 있는지 검색하거나, 모든 오류를 “없음”으로 바꾸지 않습니다.

| 오류 | 가능한 처리 방향 |
| --- | --- |
| `sql.ErrNoRows` | 찾는 대상이 없다는 업무 결과로 변환 |
| context 취소·마감 | 요청 중단이나 시간 초과 처리 |
| Scan 변환 오류 | 스키마와 Go 타입 불일치 확인 |
| 닫힌 DB·잘못된 SQL | 내부 실패로 기록하고 외부에는 일반 오류 |

결과가 여러 행일 수 있는데 QueryRow를 사용하면 전체 목록을 얻을 수 없습니다. 단일 결과가 필요한 쿼리는 기본키나 UNIQUE 조건 등으로 그 의도를 표현하고, 목록이면 QueryContext를 사용합니다.

## 6. 두 변경을 하나의 결과로 만들기

### 6.1 실패 시점이 두 문장 사이에 있다

학습 포인트를 옮기는 예를 생각합니다. 1번 지갑에서 30을 빼고 2번에 30을 더해야 합니다. 차감만 성공하고 증가가 실패하면 총합이 줄어듭니다. 두 명령이 개별적으로 올바른 SQL이라는 사실만으로 업무 전체의 성공이 보장되지 않습니다.

**트랜잭션(transaction)**은 관련 변경을 묶어 확정하거나 취소하는 단위입니다. `Commit` 성공은 변경 확정, `Rollback`은 아직 확정하지 않은 변경 취소입니다. Go에서는 `BeginTx`로 얻은 `*sql.Tx`의 메서드로 모든 관련 명령을 실행합니다. [트랜잭션 실행](https://go.dev/doc/database/execute-transactions)을 참고합니다.

이 예제는 작은 정수 학습 포인트의 이동입니다. 금액 계산, 결제 중복 방지, 장애 복구 정책을 갖춘 결제 시스템을 구현한 것은 아닙니다. 입력은 양수 ID, 서로 다른 지갑, 1~1000 포인트로 제한합니다.

저장 경로: [`examples/24-sql-transactions/transactions/main.go`](../../examples/24-sql-transactions/transactions/main.go).

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
        INSERT INTO wallets VALUES (1, 100), (2, 20)`)
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
	if err := transfer(ctx, db, 1, 2, 30); err != nil {
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
	fmt.Printf("확정 뒤: %d, %d\n", first, second)
	moveErr := transfer(ctx, db, 1, 99, 10)
	fmt.Println("목적지 없음:", errors.Is(moveErr, ErrDestination))
	first, err = points(ctx, db, 1)
	if err != nil {
		return err
	}
	second, err = points(ctx, db, 2)
	if err != nil {
		return err
	}
	fmt.Printf("복구 뒤: %d, %d\n", first, second)
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
go run ./examples/24-sql-transactions/transactions
```

```text
확정 뒤: 70, 50
목적지 없음: true
복구 뒤: 70, 50
```

### 6.2 실제 상태를 추적한다

첫 이동은 1번에서 2번으로 30을 옮깁니다. 두 번째는 없는 99번으로 10을 보내려 하므로, 첫 UPDATE가 성공했더라도 확정하면 안 됩니다.

| 순서 | 진행 중인 변경 | 함수 결과 뒤 DB에 남은 값 |
| --- | --- | --- |
| 초기 상태 | 없음 | 100, 20 |
| 1번에서 30 차감 | 트랜잭션 안: 70, 20 | 아직 확정 전 |
| 2번에 30 증가 | 트랜잭션 안: 70, 50 | 아직 확정 전 |
| Commit 성공 | 확정 | 70, 50 |
| 다음 이동에서 10 차감 | 트랜잭션 안: 60, 50 | 아직 확정 전 |
| 99번 증가 결과 0행 | 목적지 없음 | rollback 뒤 70, 50 |

테스트는 함수가 오류를 반환하는지만 확인하지 않습니다. 실패한 이동 뒤 두 값을 다시 조회하고 합계가 유지되었는지 확인합니다. 반환 오류와 저장 상태가 함께 계약을 만족해야 합니다.

### 6.3 조건을 UPDATE에 포함한다

차감 SQL의 `WHERE id = ? AND points >= ?`는 대상과 충분한 잔여 포인트를 함께 확인합니다. 적용된 행이 1개가 아니면 `ErrSource`입니다. 이 오류는 “출발 지갑 없음”과 “포인트 부족”을 함께 나타내도록 정한 예제의 계약입니다. 둘을 구분해야 하는 서비스라면 별도의 업무 계약을 설계합니다.

입력 검사와 DB의 `CHECK(points >= 0)`는 서로 보완합니다. 함수 검사는 요청 의미를 검사하고, DB 제약은 저장 상태의 규칙을 지킵니다. 한쪽이 있다는 이유로 다른 쪽이 자동으로 생기는 것은 아닙니다.

### 6.4 defer Rollback과 Commit

BeginTx 직후 rollback을 defer에 등록하면 어느 중간 단계에서 반환해도 취소 경로가 남습니다. Commit 뒤 다시 Rollback하면 `sql.ErrTxDone`이므로 이 예상된 종료 상태만 제외합니다. 다른 정리 오류는 `errors.Join`으로 보존합니다.

Commit의 오류도 반드시 반환합니다. UPDATE 두 개가 성공했다는 사실만 보고 성공 응답을 만들지 않습니다. 확정이 실패했다면 해당 트랜잭션의 결과를 성공한 값처럼 사용하지 않아야 합니다. [Tx.Commit](https://pkg.go.dev/database/sql#Tx.Commit)과 [Tx.Rollback](https://pkg.go.dev/database/sql#Tx.Rollback)을 참고합니다.

트랜잭션 안에서 `tx.ExecContext` 대신 `db.ExecContext`를 사용하면 같은 트랜잭션에 참여하지 않습니다. 특히 최대 연결 1에서 tx가 연결을 점유한 상태로 db에 새 연결을 요청하면 대기 상태가 될 수 있습니다. 같은 변수 이름이 아니라 **메서드를 호출하는 객체**를 확인합니다.

격리 수준(isolation level)은 동시에 실행되는 작업이 무엇을 볼 수 있는지와 관련됩니다. 모든 DB가 모든 수준을 지원하지 않으며, SQLite의 동시 쓰기 동작을 다른 DB와 같다고 가정하지 않습니다. 이 장은 기본 옵션의 로컬 SQLite에서 성공·실패의 원자성을 검증합니다. 격리 수준을 바꿔 성능이 좋아진다는 주장까지 하지 않습니다.

## 7. 쿼리를 보내기 전에도 기다릴 수 있다

저장 경로: [`examples/24-sql-transactions/pool-cancellation/main.go`](../../examples/24-sql-transactions/pool-cancellation/main.go).

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

func readValue(ctx context.Context, db *sql.DB) (int, error) {
	var value int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&value)
	return value, err
}

func run() (err error) {
	db, err := openDB(context.Background())
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	held, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	released := false
	defer func() {
		if !released {
			err = errors.Join(err, held.Close())
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, waitErr := readValue(ctx, db)
	fmt.Println("대기 시간 초과:", errors.Is(waitErr, context.DeadlineExceeded))
	if err := held.Close(); err != nil {
		return err
	}
	released = true
	value, err := readValue(context.Background(), db)
	if err != nil {
		return err
	}
	fmt.Println("반환 뒤 조회:", value)
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
go run ./examples/24-sql-transactions/pool-cancellation
```

```text
대기 시간 초과: true
반환 뒤 조회: 1
```

### 7.1 대기 원인이 SQL 실행 속도만은 아니다

`db.Conn`은 풀에서 하나의 전용 연결을 얻습니다. 이번 예제에서는 허용된 연결 1개를 `held`가 점유합니다. 이 상태의 `readValue`는 `SELECT 1`이 복잡해서 느린 것이 아니라 사용할 연결이 없어 기다립니다. [DB.Conn](https://pkg.go.dev/database/sql#DB.Conn)과 [연결 풀 제한](https://go.dev/doc/database/manage-connections)을 참고합니다.

| 시점 | 연결 소유자 | 새 조회 |
| --- | --- | --- |
| Open 뒤 | 풀의 유휴 연결 | 사용 가능 |
| `db.Conn` 뒤 | `held` | 연결 반환을 기다림 |
| 10ms 마감 뒤 | 여전히 `held` | DeadlineExceeded로 대기 종료 |
| `held.Close` 뒤 | 풀 | 새 context로 조회 가능 |

10ms는 예제를 끝내기 위한 짧은 마감입니다. 실제 경과 시간이 정확히 10ms라는 주장을 하지 않습니다. OS 일정에 따라 늦게 관찰될 수 있으므로 테스트도 정밀한 실행 시간 대신 `errors.Is`와 연결 풀 대기 발생, 다음 조회 성공을 검사합니다.

### 7.2 Close는 전용 연결을 풀로 돌려준다

`sql.Conn.Close`는 전용 연결 사용을 끝내고 풀에 반환합니다. `sql.DB.Close`는 DB 핸들 전체의 종료입니다. 둘을 같은 수명으로 다루지 않습니다. 예제의 `released`는 명시적 반환 뒤 defer에서 같은 정리를 불필요하게 반복하지 않기 위한 표시입니다.

만료된 context는 다시 살아나지 않습니다. 연결을 돌려준 뒤에는 새 작업용 context를 사용합니다. 취소 오류를 받은 뒤 같은 만료 context로 재시도하면 연결이 비어 있어도 성공을 기대할 수 없습니다.

### 7.3 HTTP 요청에서 DB까지 context를 전달한다

핸들러에서는 `r.Context()`에서 시작한 작업용 마감을 쿼리까지 전달합니다. DB 함수 안에서 매번 `context.Background()`로 바꾸면 요청 취소의 연결이 끊어집니다. 호출자가 정한 수명과 DB 대기가 이어져야 합니다. [진행 중 DB 작업 취소](https://go.dev/doc/database/cancel-operations)를 참고합니다.

드라이버와 DB에 따라 실제 실행 중 취소의 세부 동작은 다를 수 있습니다. 이 예제는 로컬 SQLite의 연결 대기 취소를 재현합니다. 타임아웃을 설정했다는 이유만으로 원격 시스템에서 변경이 전혀 일어나지 않았다고 일반화하지 않습니다. 쓰기 재시도에는 확정 결과와 중복 처리 계약이 함께 필요합니다.

## 8. 파일에 저장한 뒤 다시 열기

저장 경로: [`examples/24-sql-transactions/file-reopen/main.go`](../../examples/24-sql-transactions/file-reopen/main.go).

```go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

func openFile(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func save(ctx context.Context, path string) (err error) {
	db, err := openFile(ctx, path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if _, err := db.ExecContext(ctx, "CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT NOT NULL)"); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO notes(title) VALUES (?)", "저장된 Go")
	return err
}

func load(ctx context.Context, path string) (title string, err error) {
	db, err := openFile(ctx, path)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	err = db.QueryRowContext(ctx, "SELECT title FROM notes WHERE id = ?", 1).Scan(&title)
	return title, err
}

func run() (err error) {
	dir, err := os.MkdirTemp("", "go-sql-lesson-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	path := filepath.Join(dir, "lesson.db")
	ctx := context.Background()
	if err := save(ctx, path); err != nil {
		return err
	}
	title, err := load(ctx, path)
	if err != nil {
		return err
	}
	fmt.Println("재열기:", title)
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
go run ./examples/24-sql-transactions/file-reopen
```

```text
재열기: 저장된 Go
```

`save`는 DB를 열고 표와 한 행을 만든 뒤 닫습니다. `load`는 같은 파일을 새 DB 핸들로 열고 값을 읽습니다. 같은 메모리 변수에서 문자열을 꺼내 출력한 것이 아니라, 연결을 닫은 뒤 저장 결과를 다시 확인합니다.

파일 경로는 `os.MkdirTemp`가 새로 만든 폴더 아래 `lesson.db`입니다. 기존 사용자 파일과 충돌하지 않으며 전체 실행이 끝나면 자신이 만든 폴더만 정리합니다. 예제는 재열기 동작을 보여 주기 위한 것이므로 실행 뒤 영구 DB를 남기지 않습니다. 실제 프로젝트에서는 DB 파일의 경로·권한·백업·마이그레이션 정책을 정해야 합니다.

`save`는 새 DB 파일을 전제로 `CREATE TABLE`을 사용합니다. 이미 같은 표가 있는 파일에 반복 호출하는 함수라는 계약은 아닙니다. `CREATE TABLE IF NOT EXISTS`를 붙이는 것만으로 스키마 변경과 버전 이전을 다 해결했다고 설명하지 않습니다.

파일이 다시 열리는 것은 이번 테스트가 확인하는 영속성 범위입니다. 전원 장애 시점별 복구, 파일 시스템 손상, 백업 복구까지 검증한 것은 아닙니다. 30장에서는 이 파일 저장을 API 저장소와 연결하고 재시작 뒤 데이터와 ID 정책이 이어지는지 테스트합니다.

## 9. 준비된 문장과 쿼리 선택

같은 SQL을 여러 값으로 반복 실행할 때는 **준비된 문장(prepared statement)**인 `*sql.Stmt`를 사용할 수 있습니다. `PrepareContext`로 만들고, `ExecContext`·`QueryRowContext`에 값만 전달하며, 사용이 끝나면 Close합니다. [준비된 문장](https://go.dev/doc/database/prepared-statements)을 참고합니다.

이 기능을 사용하기 전에도 앞의 `ExecContext(ctx, sql, args...)` 방식은 올바른 매개변수 전달입니다. 매번 Prepare를 명시해야 SQL 주입을 막을 수 있는 것이 아닙니다. 같은 문장을 재사용할 책임이 명확할 때 선택하며, 성능 효과는 측정 없이 단정하지 않습니다.

트랜잭션 내부에서 재사용할 문장은 `tx.PrepareContext`로 준비합니다. DB에서 만든 문장을 그대로 실행하면서 같은 트랜잭션일 것이라고 가정하지 않습니다. 준비된 문장의 소유 범위와 종료 시점을 코드에 드러냅니다. 선택 문제 A에서 중복 제목 때문에 두 번째 삽입이 실패하면 첫 번째 삽입도 남지 않도록 구현합니다.

| 필요한 결과 | 선택 |
| --- | --- |
| 표 만들기, 행 추가·수정·삭제 | ExecContext와 Result 검사 |
| 존재해야 하는 한 행 | QueryRowContext와 Scan |
| 0개 이상 목록 | QueryContext, Next, Scan, Err, Close |
| 관련 변경의 한 번 확정 | BeginTx, Tx 메서드, Commit/Rollback |
| 같은 SQL의 명시적 재사용 | PrepareContext, Stmt 메서드, Close |

## 10. 직접 바꾸는 실습

본문 실습에서는 `query-rows`의 최소 학습 분을 바꾸고 정렬·빈 목록·Scan 실패를 확인합니다. 개인 복사본만 수정합니다. 다음 명령은 기존 파일이 있으면 보존합니다.

### Windows PowerShell

```powershell
$practiceDir = "_practice/24-sql-transactions/query-lab"
New-Item -ItemType Directory -Path $practiceDir -Force | Out-Null
$targetFile = Join-Path $practiceDir "main.go"
if (Test-Path $targetFile) {
    Write-Host "기존 main.go를 보존합니다: $targetFile"
} else {
    Copy-Item "examples/24-sql-transactions/query-rows/main.go" $targetFile
}
```

### macOS·Linux·WSL의 Bash

```bash
practice_dir="_practice/24-sql-transactions/query-lab"
mkdir -p "$practice_dir"
if [ -e "$practice_dir/main.go" ]; then
    printf '기존 main.go를 보존합니다: %s\n' "$practice_dir/main.go"
else
    cp examples/24-sql-transactions/query-rows/main.go "$practice_dir/main.go"
fi
```

테스트 파일은 복사하지 않습니다. 스스로 예상 결과를 작성한 뒤 개인 폴더에 테스트를 추가할 수 있습니다.

```sh
go run ./_practice/24-sql-transactions/query-lab
go fmt ./_practice/24-sql-transactions/query-lab
go vet ./_practice/24-sql-transactions/query-lab
```

1. 최소 분 20을 40으로 바꿉니다. ID 3 한 행이 남아야 합니다.
2. 41로 바꿉니다. 오류 없이 아무 항목도 출력하지 않아야 합니다.
3. 0으로 바꾸고 결과가 1, 2, 3 순서인지 확인합니다.
4. `ORDER BY id DESC`로 바꾸면 3, 2, 1 순서인지 확인합니다.
5. 개인 복사본에서만 `minutes INTEGER`를 `minutes TEXT`로 바꾸고 한 값에 숫자가 아닌 문자열을 저장해 Scan 오류를 확인합니다.
6. 마지막에 원래 스키마와 최소 분 20으로 돌려놓고 다시 실행합니다.

5번은 **의도적으로 실패하는 실행**입니다. 일반 예제와 공통 테스트에는 실패 코드를 넣지 않습니다. 실제 오류 문장의 세부 표현은 드라이버 버전에 따라 달라질 수 있지만, 숫자가 아닌 값을 int로 Scan할 수 없다는 원인을 설명할 수 있어야 합니다.

## 11. 테스트가 확인하는 것

```sh
go test -count=1 ./examples/24-sql-transactions/...
go test -count=1 ./solutions/24-sql-transactions/...
```

| 검사 | 성공 조건 | 놓치지 말아야 할 실패 |
| --- | --- | --- |
| 매개변수 저장 | 따옴표와 SQL 모양 문자열이 그대로 복원 | 문장에 직접 연결 |
| 목록 조회 | 필터 경계, 정렬, 빈 목록 | Scan 오류를 무시 |
| NULL 조회 | NULL·빈 값·행 없음 구별 | 모두 빈 문자열로 처리 |
| 트랜잭션 | 정상 이동과 실패 후 두 값 보존 | 오류만 확인하고 DB 상태 미확인 |
| 풀 대기 | 마감으로 대기 종료, 연결 반환 후 정상 조회 | 취소를 무시하고 무한 대기 |
| 파일 저장 | 닫은 DB를 새 핸들로 재열어 같은 데이터 확인 | 같은 변수만 다시 출력 |

이 테스트는 SQL 문장을 흉내 내는 가짜 드라이버를 사용하지 않습니다. `t.TempDir` 또는 메모리 SQLite로 각 테스트를 분리합니다. 메모리 DB도 실제 SQL 엔진이므로 제약·조회·rollback을 확인할 수 있습니다. 외부 운영 DB와 같은 성능이나 장애 모델을 검증하는 테스트라는 뜻은 아닙니다.

중간 Scan 오류 뒤에는 다시 Ping하여 연결이 풀로 반환되었는지 확인합니다. 연결 대기 테스트에는 전체 검사가 끝날 수 있도록 context 마감을 둡니다. 예상 시간보다 빠르거나 늦게 실행됐다는 조건을 정답으로 삼지 않습니다.

## 12. 흔한 실수와 진단 순서

| 증상 | 확인할 원인 | 수정 방향 |
| --- | --- | --- |
| unknown driver | 드라이버 import가 없음 | blank import와 모듈 의존성 확인 |
| no such table이 간헐적으로 발생 | :memory:에 여러 연결 사용 | 이번 예제의 단일 연결 정책 확인 |
| 요청이 계속 대기 | Rows·Conn·Tx가 연결을 보유 | 정리 순서와 context 확인 |
| 결과 없음인데 500으로만 처리 | ErrNoRows를 분류하지 않음 | errors.Is로 업무 의미 변환 |
| NULL에서 Scan 실패 | 일반 string만 사용 | NullString 등으로 NULL 표현 |
| 실패했는데 값 일부가 변경됨 | db/tx 혼용, Commit 위치 오류 | 모든 변경을 같은 Tx에 넣기 |
| 목록 순서가 바뀜 | ORDER BY가 없음 | 계약에 맞는 정렬 명시 |
| 취소 뒤 재시도도 바로 실패 | 같은 만료 context 재사용 | 새 작업 수명으로 시작 |
| 파일 재열기에서 표가 중복됨 | 새 DB 전제의 초기화를 반복 | 초기화와 스키마 이전 정책 구분 |

문제를 진단할 때 먼저 어떤 호출에서 오류가 났는지 기록합니다. 그다음 연결을 누가 소유하는지, Rows나 Tx가 아직 열려 있는지, context가 만료되었는지를 확인합니다. DB 비밀번호나 전체 사용자 입력을 로그에 덤프하는 방법으로 해결하지 않습니다.

## 연습문제

[24장 연습문제](../../exercises/24-sql-transactions/README.md)에는 확정 전후 상태 예측, 제목 검색, Rows 수명 오류 수정, NULL 표시, 원자적 포인트 이동, 연결 대기 취소와 준비된 문장 선택 문제가 있습니다. 힌트를 먼저 읽고 구현한 뒤 [해설](../../solutions/24-sql-transactions/README.md)과 비교합니다.

## 요약과 자기 점검

DB 핸들은 연결을 관리하고, Rows는 조회 결과를 읽으며, Tx는 관련 변경의 확정 경계를 관리합니다. SQL 값은 매개변수로 전달하고, 결과의 부재와 실행 실패를 구분합니다. 자원을 닫는 순서와 context 수명을 함께 설계해야 작은 예제도 멈추지 않고 종료됩니다.

<details>
<summary>1. Open이 성공하면 Ping을 생략해도 연결 확인이 끝난 것인가요?</summary>

아닙니다. 실제 연결 시점이 지연될 수 있습니다. PingContext나 실제 작업에서 연결을 사용할 수 있는지 확인합니다.

</details>

<details>
<summary>2. Next가 false이면 왜 Err를 추가로 읽나요?</summary>

결과를 모두 읽었을 때도, 읽는 도중 실패했을 때도 false가 될 수 있기 때문입니다. Err가 정상 종료와 반복 오류를 구분합니다.

</details>

<details>
<summary>3. NULL 제목과 없는 행은 어떻게 다르나요?</summary>

NULL 제목은 행이 존재하고 해당 열이 미정인 상태입니다. NullString.Valid가 false입니다. 없는 행은 QueryRow의 Scan에서 ErrNoRows로 전달됩니다.

</details>

<details>
<summary>4. 트랜잭션에서 db.ExecContext를 써도 같은 DB이므로 괜찮나요?</summary>

같은 트랜잭션에 자동으로 포함되지 않습니다. tx 메서드를 사용해야 합니다. 단일 연결 풀에서는 새 연결을 기다리는 문제도 생길 수 있습니다.

</details>

<details>
<summary>5. 아주 짧은 SELECT가 왜 마감될 수 있나요?</summary>

SQL 실행 전에 연결을 기다릴 수 있습니다. 예제에서는 전용 연결이 풀의 유일한 연결을 점유하므로 그 대기 중 마감됩니다.

</details>

## 참고 자료

공식 문서 확인일: **2026-11-09**.

- [database/sql API](https://pkg.go.dev/database/sql)
- [Go: 결과 조회](https://go.dev/doc/database/querying)
- [Go: SQL 주입 방지](https://go.dev/doc/database/sql-injection)
- [Go: 트랜잭션](https://go.dev/doc/database/execute-transactions)
- [Go: 연결 관리](https://go.dev/doc/database/manage-connections)
- [Go: 작업 취소](https://go.dev/doc/database/cancel-operations)
- [Go: 준비된 문장](https://go.dev/doc/database/prepared-statements)
- [modernc.org/sqlite 드라이버](https://pkg.go.dev/modernc.org/sqlite)
- [SQLite 메모리 DB](https://sqlite.org/inmemorydb.html)

| 이전 | 목차 | 다음 |
| --- | --- | --- |
| [23. API 설계와 테스트](../23-api-testing/README.md) | [교재 목차](../README.md) | [25. 서비스 운영 기초](../25-service-operations/README.md) |
