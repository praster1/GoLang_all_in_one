# 30장 연습문제 — 교체 가능한 할 일 API

[본문](../../book/30-todo-api/README.md)의 API를 실행하고 계약을 확인한 뒤 시작합니다. 이번 문제는 같은 CRUD 코드를 다시 쓰는 것보다, 기존 경계를 확장할 때 생기는 오류를 찾는 데 중점을 둡니다.

모든 명령은 **저장소 루트**에서 실행합니다. 개인 구현은 `_practice/30-todo-api/문제이름/`에 둡니다. 다른 문제의 `main`을 같은 폴더에 합치지 않습니다. 기존 `todo` 패키지를 import할 때는 루트 모듈 경로를 사용합니다.

```go
import "github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
```

기본 연습에 외부 계정이나 운영 DB는 필요하지 않습니다. SQL 문제는 `:memory:` 또는 자신이 만든 임시 DB를 사용합니다. 실제 서버를 실행하는 별도 확인에서는 loopback 주소를 사용하고 종료 방법을 먼저 확인합니다.

## 준비할 프로그램

| 문제 | 개인 폴더 | 중심 계약 |
| --- | --- | --- |
| 2 | `filtered-page` | 필터 결과와 커서의 차이 |
| 4 | `atomic-import` | 여러 삽입의 원자성 |
| 5 | `create-client` | 네트워크 오류와 API 오류 |
| 6 | `request-timeout` | 작업 제한과 취소 전파 |

Windows PowerShell:

```powershell
$practiceRoot = "_practice/30-todo-api"
foreach ($name in @("filtered-page", "atomic-import", "create-client", "request-timeout")) {
    New-Item -ItemType Directory -Force "$practiceRoot/$name" | Out-Null
}
```

macOS·Linux·WSL:

```bash
practice_root=_practice/30-todo-api
for name in filtered-page atomic-import create-client request-timeout; do
    mkdir -p "$practice_root/$name"
done
```

위 명령은 폴더만 생성하며 기존 파일을 덮어쓰지 않습니다. 문제의 타입과 함수부터 직접 작성한 다음 예제 실행 코드를 붙입니다.

## 문제 1. 확인 — 상태·오류·저장 결과 추적

새 메모리 저장소에 다음 요청을 순서대로 보냅니다. 생성과 수정 요청의 Content-Type은 application/json입니다.

1. `POST /todos`, 본문 `{"title":" Go "}`
2. `PUT /todos/1`, 본문 `{"title":"Go 완료","done":true}`
3. `PUT /todos/1`, 본문 `{"title":"   ","done":false}`
4. `GET /todos/1`
5. `DELETE /todos/1`
6. `DELETE /todos/1`
7. `GET /todos`

### 답할 내용

- 각 응답의 상태 코드를 적습니다.
- 4번의 제목과 done 값을 적습니다.
- 3번이 실패한 뒤 왜 저장 상태도 확인해야 하는지 설명합니다.
- 5번과 6번의 결과가 같은지 구분합니다.
- 7번의 본문이 null인지 빈 배열인지 설명합니다.
- 같은 요청을 새 SQLite 저장소에 보내면 어떤 결과가 같아야 하는지 적습니다.

### 완료 조건

- 정상 응답만 추적하지 않습니다.
- 실패한 수정이 기존 값을 바꾸지 않았다는 조건을 포함합니다.
- API의 삭제 계약을 다른 API의 관행으로 바꾸어 답하지 않습니다.
- 저장소 구현과 HTTP 계약을 구분해서 설명합니다.

<details>
<summary>힌트</summary>

1. 제목의 업무 규칙 오류는 422입니다.
2. Replace는 검증에 성공한 뒤에 저장합니다.
3. 이번 API는 없는 항목의 반복 삭제를 404로 반환합니다.

</details>

## 문제 2. 변형 — 미완료 항목만 보여 주는 페이지

저장소 목록의 한 페이지를 읽고, 그 안에서 미완료 항목만 골라 반환합니다. **page size는 검사할 원본 항목 수**입니다. 미완료 항목을 반드시 size개 채워 달라는 요구가 아닙니다.

```go
type OpenPage struct {
    Tasks     []todo.Task
    NextAfter int64
    HasMore   bool
}

func openPage(ctx context.Context, repo todo.Repository, after int64, size int) (OpenPage, error)
```

### 요구사항

- size는 1~99입니다.
- 다음 페이지가 있는지 보려고 size+1개를 조회합니다.
- 실제로 처리할 원본 항목은 앞에서 size개입니다.
- 그 안에서 done=false인 항목만 Tasks에 담습니다.
- NextAfter는 **마지막으로 검사한 원본 ID**입니다.
- 원본 조회 결과가 없으면 NextAfter는 입력 after를 유지합니다.
- Tasks가 비어 있어도 HasMore가 참일 수 있습니다.
- 빈 Tasks는 nil이 아닌 빈 슬라이스입니다.
- 저장소 오류와 취소 오류를 그대로 반환합니다.

### 입력 데이터

| ID | 완료 여부 |
| --- | --- |
| 1 | true |
| 2 | false |
| 3 | true |
| 4 | false |
| 5 | true |

size=2로 순회할 때 예상 출력:

```text
페이지 1: 미완료=[2] 다음=2 더있음=true
페이지 2: 미완료=[4] 다음=4 더있음=true
페이지 3: 미완료=[] 다음=5 더있음=false
```

### 추가 경계

첫 두 항목이 모두 완료이고 세 번째만 미완료라면 첫 페이지의 Tasks는 비어 있습니다. 그래도 NextAfter=2, HasMore=true여야 합니다. 빈 필터 결과만 보고 전체 순회를 끝내면 세 번째 항목을 놓칩니다.

### 완료 조건

- 마지막 미완료 ID를 커서로 사용하지 않습니다.
- 미완료 항목을 찾으려고 무제한 내부 반복을 만들지 않습니다.
- size+1이 저장소 상한 100을 넘지 않습니다.
- 원본이 없는 경우와 필터 결과만 없는 경우를 구분합니다.

<details>
<summary>힌트</summary>

1. 필터 적용 전에 조회한 원본 목록 길이로 HasMore를 계산합니다.
2. HasMore가 참이면 관찰용 마지막 항목 하나를 제외합니다.
3. 순회한 각 원본의 ID로 NextAfter를 갱신한 뒤 필터를 적용합니다.

</details>

## 문제 3. 확인 — JSON 검증은 무엇을 거절하는가

본문 API에 다음 입력을 전달한다고 가정합니다. 각 상태와 이유를 적습니다.

| 입력 | 질문 |
| --- | --- |
| `{"title":"x","admin":true}` | 알려지지 않은 필드는? |
| `{"title":"x"} {}` | JSON 값이 둘이면? |
| `{"TITLE":"first","title":"second"}` | 대소문자와 중복 키는? |
| `{"title":null}` | 필드가 있지만 null이면? |
| PUT에서 `done:false` | false 자체가 누락인가? |
| 정상 JSON 뒤에 공백을 붙여 1,025바이트 | 첫 JSON이 유효하면 통과하는가? |
| 잘못된 UTF-8 원문 바이트 | 디코더 전에 검사하는가? |
| `"\ud800"` 같은 고립 서로게이트 이스케이프 | JSON v1에서 자동으로 거절되는가? |

### 완료 조건

- DisallowUnknownFields 하나가 모든 엄격성을 제공한다고 답하지 않습니다.
- 바이트 유효성과 JSON 문자열 이스케이프의 해석을 구분합니다.
- 필요한 계약을 바꾸려면 어떤 테스트를 추가해야 할지 적습니다.
- 새로운 JSON API로 바꾸면 호환성도 검토해야 함을 설명합니다.

<details>
<summary>힌트</summary>

1. 본문 `decode`의 검사 순서를 읽습니다.
2. `encoding/json` 문서에서 중복 이름과 Unicode 치환 규칙을 확인합니다.
3. 이 프로젝트의 테스트 중 JSON v1 동작을 명시한 사례를 찾습니다.

</details>

## 문제 4. 응용 — 일괄 가져오기의 전체 성공·전체 실패

다음 함수로 1~100개의 제목을 한 번에 저장합니다.

```go
func importTasks(ctx context.Context, db *sql.DB, titles []string) error
```

이 문제는 본문 todos 표와 **별도의 imports 표**를 사용합니다. imports의 title에는 UNIQUE 제약을 둡니다. 같은 제목을 여러 번 허용하는 본문 API의 규칙을 바꾸지 않습니다.

### 요구사항

- 모든 제목에 `todo.NormalizeTitle`을 적용합니다.
- 제목 검증이 실패하면 어느 항목인지 알 수 있는 오류를 반환합니다.
- 모든 삽입은 같은 `sql.Tx`에서 실행합니다.
- 중간 SQL 오류가 나면 앞서 삽입한 새 항목도 남기지 않습니다.
- 성공한 경우에만 Commit합니다.
- 모든 경로에서 트랜잭션을 정리합니다.
- SQL 문장과 값을 분리합니다.
- 입력이 비거나 100개를 넘으면 오류입니다.
- 이미 취소된 context를 반환하면서 변경하지 않습니다.

### 실행 시나리오

1. 기존 항목 `기존`을 저장합니다.
2. `새 항목`, `기존` 두 제목을 한 번에 가져옵니다.
3. 중복 제약으로 실패했는지 확인합니다.
4. 전체 개수가 여전히 1인지 확인합니다.
5. `Go`, `SQL`을 가져옵니다.
6. 전체 개수가 3인지 확인합니다.

예상 출력:

```text
중복 실패: true
실패 뒤 개수: 1
성공 뒤 개수: 3
```

### 완료 조건

- 오류 반환뿐 아니라 실패 후 DB 상태를 검사합니다.
- tx가 있는 구간에서 db.ExecContext를 섞지 않습니다.
- 원래 오류를 Rollback의 ErrTxDone으로 덮어쓰지 않습니다.
- 오류 문구가 특정 SQLite 패치와 정확히 같아야 한다고 테스트하지 않습니다.

<details>
<summary>힌트</summary>

1. 입력 정규화는 DB 작업을 시작하기 전에 할 수 있습니다.
2. BeginTx 이후에는 모든 쓰기에 같은 tx를 사용합니다.
3. defer Rollback을 두고 Commit 성공 뒤의 ErrTxDone만 정상적으로 구분합니다.

</details>

## 문제 5. 응용 — 생성 결과를 검증하는 클라이언트

다음 함수를 만듭니다.

```go
func createTask(ctx context.Context, client *http.Client, base, title string) (todo.Task, error)
```

### 요구사항

- base는 HTTP 또는 HTTPS 원점 URL입니다. 사용자 정보·쿼리·조각·추가 경로를 받지 않습니다.
- title을 도메인 규칙으로 정규화합니다.
- POST /todos에 JSON을 보냅니다.
- 요청에 받은 context를 연결합니다.
- 응답 본문을 4,096바이트까지 허용하고 초과를 감지합니다.
- 본문을 닫고 읽기 오류와 닫기 오류를 확인합니다.
- 성공 상태는 정확히 201입니다.
- 201의 항목 ID가 양수이고 제목이 정규화한 입력과 같으며 done=false인지 확인합니다.
- 다른 상태는 상태 코드와 오류 코드를 담는 `APIError`로 반환합니다.
- HTTP 요청을 반복 재시도하는 루프를 넣지 않습니다.

`APIError`는 다음 정보를 제공합니다.

```go
type APIError struct {
    Status int
    Code   string
}
```

### 확인할 실패

- 서버가 500 JSON 오류를 반환함
- 400인데 오류 본문이 JSON이 아님
- 201인데 본문이 잘못된 JSON임
- 201인데 ID가 0임
- 응답 본문이 4,097바이트임
- 요청 시작 전에 context가 취소됨

### 완료 조건

- `client.Do`가 nil 오류를 반환했다고 201로 간주하지 않습니다.
- 실제 외부 서버 대신 httptest 서버로 검증합니다.
- 상위 호출자는 errors.As로 APIError를 분류할 수 있습니다.
- 재시도하면 중복 생성될 수 있는 이유를 설명합니다.

<details>
<summary>힌트</summary>

1. JSON 직렬화에는 문자열 연결 대신 Marshal을 사용합니다.
2. 제한+1 바이트를 읽으면 정확한 상한 초과를 구분할 수 있습니다.
3. 네트워크 실패, HTTP 실패, 성공 응답의 계약 실패를 나눕니다.

</details>

## 문제 6. 응용 — 요청 context에 제한 추가하기

핸들러를 감싸 요청 context에서 자식 제한 시간을 만듭니다.

```go
func withDeadline(timeout time.Duration, next http.Handler) http.Handler
```

실습에서는 양수 timeout을 사용합니다. 다음 핸들러에 새 context를 포함한 요청을 전달하고, 돌아온 뒤 취소 함수를 호출합니다.

### 시험용 저장소

Get만 구현한 저장소 대역은 `<-ctx.Done()`에서 기다렸다가 ctx.Err를 반환합니다. 기본 API의 다른 메서드는 이 실습에서 호출하지 않습니다. 이를 `todo.NewHandler`에 연결하고 `GET /todos/1`을 요청합니다.

예상 결과:

```text
상태: 504
본문: {"error":"deadline_exceeded"}
```

### 추가 요구사항

- 이미 취소된 부모 context에서는 503으로 매핑되는지 확인합니다.
- 오류가 없고 바로 끝나는 핸들러도 정상 호출되는지 확인합니다.
- 실제 시간을 얼마나 기다렸는지 숫자로 테스트하지 않습니다.
- 별도 고루틴을 무제한으로 만들지 않습니다.
- timeout이 context를 무시하는 함수 실행을 강제 중단한다고 설명하지 않습니다.

### 완료 조건

- 부모 context를 버리고 Background에서 요청 수명을 새로 만들지 않습니다.
- 요청의 취소와 제한 시간 오류를 구분합니다.
- 테스트의 대기 작업은 반드시 ctx.Done을 관찰합니다.
- 실제 소켓의 Body.Read 제한은 서버의 ReadTimeout과 별도로 이해합니다.

<details>
<summary>힌트</summary>

1. r.WithContext가 반환하는 요청을 next에 전달해야 합니다.
2. context.WithTimeout만 호출하고 원래 r을 넘기면 전달되지 않습니다.
3. 오류 응답은 새로 만들지 말고 기존 API의 분류를 재사용합니다.

</details>

## 문제 7. 확인·검증 — 같은 계약을 다른 구현에 적용

본문 저장소 테스트를 읽고 다음 질문에 답합니다.

1. 메모리 테스트가 통과했다고 SQLite 테스트를 생략해도 되나요?
2. 두 구현이 모두 Repository를 만족하면 List 정렬 순서도 같은가요?
3. 반환된 Task를 수정했을 때 원본이 바뀌지 않는지 왜 검사하나요?
4. SQLite를 같은 핸들로 다시 조회하는 것만으로 파일 영속성을 증명할 수 있나요?
5. race detector가 통과하면 사용자의 동시 편집도 보존되나요?
6. context가 취소되면 이미 성공한 저장도 자동으로 복구되나요?

이어서 아래 명령으로 본문과 해설의 검사를 실행합니다.

```sh
go test -count=1 ./examples/30-todo-api/...
go test -count=1 ./solutions/30-todo-api/...
```

지원 환경에서는 다음도 실행합니다.

```sh
go test -race ./examples/30-todo-api/... ./solutions/30-todo-api/...
```

### 완료 조건

- 언어 수준 인터페이스 만족과 관찰 가능한 계약을 구분합니다.
- 데이터 레이스와 논리적 갱신 충돌을 구분합니다.
- 실행하지 않은 환경의 결과를 통과로 적지 않습니다.

## 선택 문제. 충돌하는 두 편집

두 사용자가 같은 항목을 읽었다고 가정합니다. A는 제목을 바꾸고, B는 예전 제목을 가진 채 done만 바꾼 전체 PUT 요청을 보냅니다. B가 나중에 성공하면 어떤 제목이 남는지 설명합니다.

버전 필드나 ETag를 추가한다면 어떤 조건에서 요청을 거절해야 할지 설계해 봅니다. 모든 저장소를 실제로 변경하기 전에 새 계약, HTTP 상태, 원자적 비교·갱신 방법과 테스트 사례를 먼저 적습니다.

[해설 보기](../../solutions/30-todo-api/README.md) · [본문으로 돌아가기](../../book/30-todo-api/README.md)
