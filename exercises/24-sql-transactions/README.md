# 24장 연습문제: SQL과 트랜잭션

[본문](../../book/24-sql-transactions/README.md)에서 `*sql.DB`, `Rows`, `Tx`가 각각 무엇을 소유하는지 먼저 확인합니다. 이 문제는 실제 SQLite를 사용합니다. 외부 DB 서버나 계정이 필요하지 않으며, 운영 데이터와 기존 파일을 연결하지 않습니다.

모든 명령은 저장소 루트에서 실행합니다. 루트의 `go.mod`에 등록된 드라이버를 사용하고 하위 모듈을 만들지 않습니다. 기본 문제는 메모리 DB이며, `SetMaxOpenConns(1)`과 `SetMaxIdleConns(1)`을 유지합니다. `context`, 오류 확인과 자원 정리는 구현의 일부입니다.

## 개인 파일 준비

문제 2~6과 선택 A의 폴더에 연결 확인 예제만 복사합니다. 각 문제의 요구사항에 맞게 SQL과 함수를 직접 추가합니다. 기존 개인 파일은 그대로 보존하며, 예제 테스트를 복사하지 않습니다.

### Windows PowerShell

```powershell
$names = @("note-search", "repair", "nullable-note", "move-points", "pool-deadline", "prepared-batch")
foreach ($name in $names) {
    $practiceDir = Join-Path "_practice/24-sql-transactions" $name
    New-Item -ItemType Directory -Path $practiceDir -Force | Out-Null
    $targetFile = Join-Path $practiceDir "main.go"
    if (Test-Path $targetFile) {
        Write-Host "기존 main.go를 보존합니다: $targetFile"
    } else {
        Copy-Item "examples/24-sql-transactions/open-ping/main.go" $targetFile
    }
}
```

### macOS·Linux·WSL의 Bash

```bash
for name in note-search repair nullable-note move-points pool-deadline prepared-batch; do
    practice_dir="_practice/24-sql-transactions/$name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '기존 main.go를 보존합니다: %s\n' "$practice_dir/main.go"
    else
        cp examples/24-sql-transactions/open-ping/main.go "$practice_dir/main.go"
    fi
done
```

전체 코드는 [해설](../../solutions/24-sql-transactions/README.md)에 분리되어 있습니다. 먼저 예상 상태와 테스트 입력을 적고, 단계별 힌트 뒤에도 막히는 부분만 비교합니다.

## 문제 1. 확정하지 않은 값 — 확인

새 메모리 DB에 `counter(value INTEGER)`를 만들고 5를 한 행 넣습니다. 다음 순서로 모두 오류 없이 실행한다고 가정합니다.

1. `BeginTx`로 트랜잭션을 시작합니다.
2. `tx.ExecContext`로 `UPDATE counter SET value=value+2`를 실행합니다.
3. `tx.QueryRowContext`로 값을 읽어 `inside`에 저장합니다.
4. `tx.Rollback()`을 실행합니다.
5. `db.QueryRowContext`로 값을 읽어 `outside`에 저장합니다.
6. `트랜잭션 안:`, `복구 뒤:`라는 라벨로 각각 출력합니다.

두 출력 값을 예측하고, 3단계와 5단계의 메서드 수신자가 다른 이유를 설명하세요. Rollback을 Commit으로 바꾸면 두 번째 값은 어떻게 달라질까요? 트랜잭션이 연결을 점유한 동안 같은 단일 연결 `db`에 별도 조회를 요청하면 어떤 문제가 가능한지도 설명합니다.

<details>
<summary>힌트 1</summary>

트랜잭션 안에서 읽은 값과 확정되어 남은 값은 같은 질문이 아닙니다.

</details>

<details>
<summary>힌트 2</summary>

중간 계산이 성공했다는 사실과 Commit이 성공했다는 사실을 분리하세요. Rollback은 아직 확정하지 않은 변경을 취소합니다.

</details>

완료 조건: 출력뿐 아니라 값이 남거나 사라지는 확정 경계를 설명합니다.

## 문제 2. 제목을 정확히 검색하기 — 변형

개인 경로: `_practice/24-sql-transactions/note-search/main.go`.

표는 `notes(id INTEGER PRIMARY KEY, title TEXT NOT NULL UNIQUE)`입니다. ID 1의 제목은 `Go's SQL`, ID 2는 `SQL`입니다. 제목으로 정확히 한 행을 찾는 `findByTitle(ctx, db, title)` 함수를 구현하세요. 반환값은 ID와 제목을 담은 구조체, 그리고 error입니다.

요구사항:

- 값은 `?` 매개변수로 전달합니다.
- 찾지 못하면 `errors.Is(err, sql.ErrNoRows)`로 구별 가능한 오류를 반환합니다.
- 작은따옴표가 포함된 제목도 원래 그대로 검색합니다.
- `"' OR 1=1 --"`라는 입력은 그 문자열과 같은 제목을 찾으며 임의의 행을 선택하면 안 됩니다.
- 실패한 결과 구조체를 정상 데이터처럼 출력하지 않습니다.

기본 실행 결과:

```text
검색: 1, Go's SQL
문자 그대로 검색: true
```

```sh
go run ./_practice/24-sql-transactions/note-search
go test ./_practice/24-sql-transactions/note-search
```

테스트에는 정상 제목, 없는 제목, 빈 문자열, 작은따옴표와 SQL 구문처럼 보이는 입력, 이미 취소한 context를 포함합니다. 없는 행과 취소 오류를 같은 오류로 바꾸지 않습니다.

<details>
<summary>힌트 1</summary>

한 행을 기대하는 조회의 오류를 어느 메서드가 반환하는지 본문 5절에서 확인하세요.

</details>

<details>
<summary>힌트 2</summary>

SQL 문자열을 `fmt.Sprintf`로 조합하지 않습니다. 오류에 설명을 붙일 때는 원인을 검사할 수 있도록 `%w`를 사용합니다.

</details>

완료 조건: SQL처럼 생긴 검색어도 데이터로 처리하며, 실제 DB 조회 결과와 오류 종류를 검사합니다.

## 문제 3. 첫 행만 읽고 연결을 놓치다 — 오류 수정

개인 경로: `_practice/24-sql-transactions/repair/main.go`.

다음은 **의도적으로 자원을 정리하지 않는 함수 선언 발췌**입니다. 단일 연결 DB에서 실행합니다.

```go
func firstValue(ctx context.Context, db *sql.DB) (int, error) {
    rows, err := db.QueryContext(ctx, "SELECT 1 UNION ALL SELECT 2")
    if err != nil {
        return 0, err
    }
    if !rows.Next() {
        return 0, rows.Err()
    }
    var value int
    if err := rows.Scan(&value); err != nil {
        return 0, err
    }
    return value, nil
}
```

이 함수를 호출한 뒤 같은 `db`로 `SELECT 9`를 실행하면 어떤 자원을 기다릴 수 있나요? 다음 조회에는 반드시 짧은 context 마감을 두어 영원히 대기하지 않게 관찰합니다. 잘못된 함수를 일반 예제나 공통 검사 대상에 추가하지 않습니다.

수정 요구사항:

- 첫 행을 읽고 조기 반환하더라도 Rows를 닫습니다.
- Query, Next 뒤 Err, Scan, Close 오류를 잃지 않습니다.
- 결과가 없다면 `sql.ErrNoRows`로 구분합니다.
- 함수 반환 뒤의 `SELECT 9`가 성공해야 합니다.

수정 뒤 출력:

```text
첫 행: 1, 다음 조회: 9
```

```sh
go run ./_practice/24-sql-transactions/repair
```

<details>
<summary>힌트 1</summary>

`rows`는 결과를 담은 일반 슬라이스가 아닙니다. 모두 읽기 전에 함수를 떠나면 누가 그 결과의 수명을 끝내나요?

</details>

<details>
<summary>힌트 2</summary>

Query가 성공한 직후 정리를 등록합니다. 함수가 `rows.Next()`에서 반환하는 경우와 `Scan` 오류로 반환하는 경우에도 같은 정리가 실행되어야 합니다.

</details>

완료 조건: 단순히 최대 연결 수를 늘려 증상을 감추지 않고, 다음 쿼리 전에 연결을 돌려줍니다.

## 문제 4. 제목의 세 상태 — 변형

개인 경로: `_practice/24-sql-transactions/nullable-note/main.go`.

`notes(id INTEGER PRIMARY KEY, title TEXT)`에 ID 1은 `Go`, ID 2는 NULL, ID 3은 빈 문자열을 저장합니다. `noteLabel(ctx, db, id) (string, error)`를 구현하세요.

| 조회 상태 | 반환 문자열 | 오류 |
| --- | --- | --- |
| 제목 Go | `Go` | nil |
| 제목 NULL | `(미정)` | nil |
| 제목 빈 문자열 | `(빈 문자열)` | nil |
| ID 99 없음 | 빈 문자열 | ErrNoRows를 검사할 수 있는 오류 |
| DB 닫힘·쿼리 실패 | 빈 문자열 | 원래 실패를 보존한 오류 |

기본 결과:

```text
1: Go
2: (미정)
3: (빈 문자열)
행 없음: true
```

```sh
go run ./_practice/24-sql-transactions/nullable-note
go test ./_practice/24-sql-transactions/nullable-note
```

닫힌 DB의 오류도 “행 없음”으로 처리하는 구현은 계약을 만족하지 않습니다. 오류 종류를 실제 테스트로 구별하세요.

<details>
<summary>힌트 1</summary>

빈 문자열만으로 NULL을 나타낼 수 없습니다. 본문에서 값과 유효성을 함께 보관한 타입을 찾아보세요.

</details>

<details>
<summary>힌트 2</summary>

Scan의 오류를 먼저 확인하고, 그 다음 Valid, 그 다음 String을 판단하면 세 경우가 겹치지 않습니다.

</details>

완료 조건: NULL·빈 문자열·행 부재·저장소 고장을 서로 구별합니다.

## 문제 5. 포인트 이동은 전부 또는 취소 — 응용

개인 경로: `_practice/24-sql-transactions/move-points/main.go`.

표 `wallets(id INTEGER PRIMARY KEY, points INTEGER NOT NULL CHECK(points >= 0))`의 초기 값은 1번 50, 2번 10입니다. `transfer(ctx, db, from, to, amount) error`를 구현하세요.

입력 계약은 양수 ID 두 개, 서로 다른 ID, 이동량 1~1000입니다. 작은 정수 학습 포인트만 다룹니다. 오류 분류는 다음과 같이 고정합니다.

| 상황 | 오류 |
| --- | --- |
| 입력 계약 위반 | ErrInput |
| 출발 지갑 없음 또는 포인트 부족 | ErrSource |
| 목적지 지갑 없음 | ErrDestination |
| SQL·context·Commit 실패 | 해당 원인을 보존한 오류 |

성공 시 출발에서 차감하고 목적지에 더합니다. 어느 단계든 실패하면 두 지갑 모두 호출 전 값이 유지되어야 합니다. 관련 SQL은 같은 Tx에서 실행하며, 두 UPDATE의 RowsAffected를 검사합니다.

기본 실행은 1번에서 2번으로 15 이동 후, 1번에서 없는 9번으로 5 이동을 시도합니다.

```text
성공: 35, 25
복구 필요: true
유지: 35, 25
```

```sh
go run ./_practice/24-sql-transactions/move-points
go test ./_practice/24-sql-transactions/move-points
```

독립된 새 DB에서 검사할 입력:

| 입력 | 기대 상태 |
| --- | --- |
| 1→2, 15 | 35, 25 |
| 1→2, 50 | 0, 60 |
| 1→2, 51 | ErrSource, 50, 10 |
| 1→9, 5 | ErrDestination, 50, 10 |
| 9→2, 5 | ErrSource, 50, 10 |
| 같은 ID, 0·음수·1001 이동량 | ErrInput, 원래 값 |
| 이미 취소한 context | 취소 오류, 원래 값 |

<details>
<summary>힌트 1</summary>

두 번째 UPDATE가 0행을 변경했을 때 첫 번째 차감이 이미 실행되었음을 생각하세요. 오류 반환만으로는 자동 취소되지 않습니다.

</details>

<details>
<summary>힌트 2</summary>

BeginTx 성공 뒤 취소 경로를 등록하고, 모든 업무 조건이 성공했을 때만 Commit합니다. Commit 뒤 defer Rollback이 돌려주는 예상 종료 오류는 ErrTxDone입니다.

</details>

완료 조건: 모든 사례에서 결과 오류뿐 아니라 두 값과 총합 60을 실제로 다시 조회해 확인합니다.

## 문제 6. 연결을 기다리는 조회 취소 — 응용

개인 경로: `_practice/24-sql-transactions/pool-deadline/main.go`.

`notes(id INTEGER PRIMARY KEY)`에 ID 1과 2를 넣습니다. `countNotes(ctx, db) (int, error)`는 `SELECT COUNT(*) FROM notes`를 실행합니다. 전달받은 context를 그대로 사용합니다.

1. 최대 연결을 1로 유지합니다.
2. `db.Conn`으로 유일한 연결을 빌려 보관합니다.
3. 10ms 마감을 가진 새 context로 `countNotes`를 호출합니다.
4. DeadlineExceeded인지 확인합니다.
5. 전용 연결을 반환한 뒤 만료되지 않은 context로 다시 조회합니다.

```text
대기 취소: true
조회: 2
```

```sh
go run ./_practice/24-sql-transactions/pool-deadline
go test ./_practice/24-sql-transactions/pool-deadline
```

테스트는 풀에서 실제로 대기했는지 `db.Stats().WaitCount`를 확인하고, 다음 조회가 성공하는지도 확인합니다. 정확히 10ms가 걸린다는 단언은 하지 않습니다. 빈 표는 오류 없이 0이며, 이미 취소한 context는 취소 오류입니다.

<details>
<summary>힌트 1</summary>

현재 대기하는 이유는 COUNT의 계산량이 아닙니다. 사용할 연결을 누가 점유하고 있는지 확인합니다.

</details>

<details>
<summary>힌트 2</summary>

전용 연결의 Close와 DB 전체의 Close를 구분하세요. 만료된 context는 연결을 돌려준 후에도 만료 상태입니다.

</details>

완료 조건: 고루틴을 추가해 실패를 숨기지 않고, 마감·정리·새 조회의 순서를 설명합니다.

## 선택 A. 준비된 문장으로 묶음 삽입

개인 경로: `_practice/24-sql-transactions/prepared-batch/main.go`.

`notes(id INTEGER PRIMARY KEY, title TEXT NOT NULL UNIQUE)`에 여러 제목을 넣는 `addTitles(ctx, db, titles []string) error`를 구현합니다. 문장은 Tx에서 한 번 Prepare하고 반복해서 값을 전달합니다. 전부 성공해야 Commit합니다.

첫 묶음 `Go`, `SQL`은 성공합니다. 두 번째 묶음 `HTTP`, `Go`는 중복 때문에 실패합니다. 두 번째 묶음의 HTTP도 남으면 안 됩니다.

```text
중복 실패: true
유지된 행: 2
```

```sh
go run ./_practice/24-sql-transactions/prepared-batch
```

빈 묶음도 정상적으로 종료되어야 하며, Statement와 Tx를 모두 정리합니다. 중복 오류의 문자열 전체를 버전 고정 정답처럼 비교하지 않습니다. 원자성과 남은 행을 검사합니다. [PrepareContext](https://pkg.go.dev/database/sql#Tx.PrepareContext)와 [공식 준비된 문장 안내](https://go.dev/doc/database/prepared-statements)를 참고합니다.

<details>
<summary>힌트</summary>

Statement의 소유자가 DB인지 Tx인지 확인합니다. 반복 중 하나가 실패하면 반복 밖의 Commit에 도달하지 않아야 합니다.

</details>

## 자기 주도 진단과 완료 조건

- [ ] SQL 값이 문장에 직접 이어 붙지 않는다.
- [ ] QueryRow의 Scan 오류와 Rows의 반복 오류를 확인한다.
- [ ] Rows·Conn·Tx·DB의 정리 책임을 설명할 수 있다.
- [ ] NULL·빈 문자열·없는 행·내부 실패를 구별한다.
- [ ] 트랜잭션 실패 뒤 실제 저장 상태가 보존된다.
- [ ] 연결 대기를 취소하고 이후 DB를 다시 사용할 수 있다.
- [ ] 의도적 실패 코드는 개인 실습 범위에만 남겼다.

오류가 났다면 해당 호출, context 상태, 점유 중인 자원, 스키마·Scan 타입 순으로 확인합니다. 무조건 풀 크기를 늘리거나 모든 오류를 “없음”으로 바꾸지 않습니다.

[해설과 실행 가능한 참고 구현](../../solutions/24-sql-transactions/README.md) · [본문으로 돌아가기](../../book/24-sql-transactions/README.md)
