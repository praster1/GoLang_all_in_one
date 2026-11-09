# 31장 연습문제 — 동시 배치 처리

[본문](../../book/31-concurrent-batch/README.md)의 실행 엔진을 검토하고 작은 확장을 구현합니다.
기본 5문제와 완료 검증 과제입니다.
[해설](../../solutions/31-concurrent-batch/README.md)은 구현한 뒤 확인합니다.

## 준비

모든 명령은 저장소 루트에서 실행합니다.
공통 엔진은 다음 경로로 import합니다.

```go
import "github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
```

하위 go.mod를 만들지 않습니다.
개인 실습은 `_practice/31-concurrent-batch/` 아래에 문제별 폴더를 만듭니다.
기존 공통 엔진이나 해설을 덮어쓰지 않습니다.

Bash:

```sh
mkdir -p _practice/31-concurrent-batch/csv-report
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/31-concurrent-batch/csv-report | Out-Null
```

현재 완성 프로그램을 먼저 실행합니다.

```sh
go run ./examples/31-concurrent-batch/cmd/batch
go test ./examples/31-concurrent-batch/...
```

## 문제 1. 항목 상태와 전체 오류 추적

**수준: 확인**

작업자 한 명, 최대 시도 세 번으로 다음 입력을 처리한다고 가정합니다.

| 항목 | 콜백의 동작 |
| --- | --- |
| A | 첫 호출 성공 |
| B | 첫 호출 ErrTemporary, 둘째 호출 성공 |
| C | 첫 호출 영구 오류 |

각 항목의 State와 Attempts를 채우고,
전체 오류와 Summary의 Total·Succeeded·Failed·Canceled·Attempts를 적습니다.

### 추가 질문

1. 재시도 세 번은 첫 시도까지 포함한 세 번인가요?
2. B가 두 번 호출되면 Total도 하나 늘어나나요?
3. C가 실패했을 때 이미 성공한 A와 B의 결과는 사라지나요?
4. 같은 입력 문자열이 두 개 있으면 한 Job으로 합쳐지나요?
5. BatchError에서 원래 sentinel 오류를 찾는 방법은 무엇인가요?

### 취소 시나리오

작업자 한 명으로 A, B, C를 처리합니다.
A는 성공하고 B의 콜백에서 부모 취소 함수를 호출한 뒤 ctx.Err를 반환합니다.
C는 아직 실행하지 않았습니다.

| 항목 | State | Attempts | 결과 보존 여부 |
| --- | --- | --- | --- |
| A | 직접 작성 | 직접 작성 | 직접 작성 |
| B | 직접 작성 | 직접 작성 | 직접 작성 |
| C | 직접 작성 | 직접 작성 | 직접 작성 |

작업자가 세 명이면 어떤 부분을 같은 값으로 고정할 수 없는지도 설명합니다.

### 힌트

Total은 입력 수이고 Attempts는 콜백 호출 횟수입니다.
미시작 취소와 실행 중 취소는 시도 횟수로 구분할 수 있습니다.

## 문제 2. 완료 결과를 CSV로 내보내기

**수준: 변형**

개인 폴더: `_practice/31-concurrent-batch/csv-report`

```go
func writeCSV(out io.Writer, results []batch.Result) error
```

### 출력 계약

첫 줄은 다음 헤더입니다.

```text
index,input,state,attempts,output,error
```

그 뒤 입력 순서대로 한 결과씩 기록합니다.
오류가 없으면 error 필드는 빈 문자열입니다.
실패한 결과의 오류 문자열에는 `result.Err.Error()`를 사용합니다.

공통 Run이 반환한 완료 결과를 입력으로 받습니다.
임의로 만든 Result의 상태 전체를 검증하는 파서까지 구현할 필요는 없습니다.

정상 go·한글 입력의 출력:

```text
index,input,state,attempts,output,error
0,go,succeeded,1,GO,
1,한글,succeeded,1,한글,
```

### 필수 조건

- 문자열을 쉼표로 직접 연결하지 않고 encoding/csv를 사용합니다.
- 입력·출력에 쉼표·따옴표·줄바꿈이 있어도 필드가 보존되어야 합니다.
- 빈 결과에서도 헤더는 한 줄 출력합니다.
- Write 오류와 Flush 뒤의 Error를 모두 확인합니다.
- 함수가 Writer를 닫지 않습니다. Writer의 수명은 호출자에게 있습니다.

### 필수 테스트

| 상황 | 확인 |
| --- | --- |
| 정상 두 항목 | 헤더와 필드 순서 |
| 쉼표·따옴표·줄바꿈 포함 | csv.Reader로 읽었을 때 원문 보존 |
| nil 결과 | 헤더 한 줄 |
| 실패하는 Writer | 오류 반환 |
| 실패 항목 | 상태·시도 횟수·오류 설명 보존 |

### 힌트

csv.Writer는 내부 버퍼가 있으므로
각 Write가 성공해도 최종 Flush에서 오류가 나타날 수 있습니다.
[15장](../../book/15-data-time/README.md)의 CSV 쓰기를 다시 확인합니다.

## 문제 3. 항목별 제한 시간 추가

**수준: 응용**

개인 폴더: `_practice/31-concurrent-batch/job-timeout`

```go
func withJobTimeout(
    timeout time.Duration,
    process batch.Processor,
) (batch.Processor, error)
```

실행 엔진을 수정하지 않고 Processor를 감싸는 함수를 만듭니다.
각 호출마다 부모에서 WithTimeout으로 자식 Context를 만들고,
자식 Context를 실제 process에 전달합니다.

### 계약

| 항목 | 규칙 |
| --- | --- |
| timeout | 0~1분 |
| nil process | 생성 오류 |
| 자식 취소 함수 | 각 호출에서 defer로 정리 |
| 이미 마감된 자식 | 실제 process 호출 전 오류 |
| 실제 process 오류 | 원래 오류 보존 |
| process 성공 뒤 마감 | 값 대신 자식 Context 오류 |
| 강제 중단 | 제공하지 않음; process가 협조해야 함 |

테스트에서 timeout=0을 사용하면 실제로 잠들지 않고 이미 마감된 상태를 재현할 수 있습니다.
운영 환경에서 항상 0초를 사용하라는 뜻은 아닙니다.

배치 전체 Context는 Background로 두고,
항목 제한 시간 0, 최대 시도 3으로 한 항목을 실행합니다.

```text
항목 상태: failed
시도 횟수: 1
항목 마감 오류: true
```

### 왜 canceled가 아닌가요?

그 항목에만 만든 자식 Context가 마감되었습니다.
Run에 전달한 전체 Context는 취소되지 않았습니다.
이 차이를 설명하고 실제 결과의 State와 errors.Is 결과를 확인합니다.

### 필수 테스트

- 0초 제한에서는 실제 process를 호출하지 않음.
- 정상 제한 시간에서 성공 결과 보존.
- 부모가 이미 취소된 경우.
- process가 반환한 sentinel 오류 보존.
- 음수 제한 시간·1분 초과·nil process 거부.

### 힌트

Context를 새 Background에서 시작하면 부모 취소와 연결이 끊깁니다.
매번 전달받은 parent에서 자식을 만듭니다.
전체 작업을 고루틴으로 감싸 호출자만 먼저 반환하는 방식은 사용하지 않습니다.

## 문제 4. 재시도 후보 목록 만들기

**수준: 응용**

개인 폴더: `_practice/31-concurrent-batch/retry-plan`

```go
func retryInputs(results []batch.Result) []string
```

다음 조건을 모두 만족하는 결과의 원래 Input만 추출합니다.

1. State가 Failed입니다.
2. Err에서 errors.Is(batch.ErrTemporary)가 true입니다.

### 계약

- 입력 결과의 순서를 유지합니다.
- 중복 입력도 별도 항목으로 보존합니다.
- Succeeded와 Canceled는 자동 후보에 넣지 않습니다.
- 영구 실패는 제외합니다.
- 래핑된 일시 오류도 포함합니다.
- 빈 입력은 nil이 아닌 빈 슬라이스입니다.
- 후보를 만드는 것뿐이며 외부 작업을 실제로 다시 실행하지 않습니다.

예시 자료:

| 입력 | State | 오류 |
| --- | --- | --- |
| ok | Succeeded | 없음 |
| busy | Failed | ErrTemporary를 감싼 오류 |
| invalid | Failed | 일반 입력 오류 |
| canceled | Canceled | 취소 오류 |

```text
검토할 재시도 입력: [busy]
```

### 필수 테스트

래핑되지 않은 ErrTemporary와 래핑된 ErrTemporary를 모두 포함합니다.
입력 순서와 중복을 보존하는지 확인합니다.
Canceled의 시도 횟수가 0 또는 1이어도 이번 자동 후보에서는 제외합니다.

### 설계 질문

취소된 항목을 자동 재실행하기 전에 무엇을 확인해야 하나요?
오류가 반환되었다고 외부 부작용이 전혀 없었다고 단정할 수 없는 이유를 설명합니다.

### 힌트

문자열에 temporary가 포함되었는지 검사하지 않습니다.
상태와 원인 오류를 함께 검사합니다.

## 문제 5. 실행 엔진의 수명·소유권 검토

**수준: 확인·응용**

[batch/batch.go](../../examples/31-concurrent-batch/batch/batch.go)를 읽고 표를 채웁니다.

| 대상 | 누가 쓰거나 닫는가? | 언제 다른 곳에서 읽는가? |
| --- | --- | --- |
| jobs 전송 | 직접 작성 | 직접 작성 |
| jobs 닫기 | 직접 작성 | 직접 작성 |
| results[index] | 직접 작성 | 직접 작성 |
| 미시작 항목 상태 | 직접 작성 | 직접 작성 |
| 재시도 Timer | 직접 작성 | 직접 작성 |
| 호출 Context | 직접 작성 | 직접 작성 |

다음 변경안을 검토합니다.

1. 분배를 마치기 전에 jobs를 작업자 중 한 명이 닫는다.
2. 취소되면 group.Wait를 생략하고 바로 반환한다.
3. 결과 칸 대신 모든 작업자가 공유 슬라이스에 append한다.
4. ErrTemporary 여부를 무시하고 모든 오류를 재시도한다.
5. Timer 대신 time.Sleep을 사용한다.
6. 입력 한 줄마다 고루틴을 새로 만들고 Workers 플래그는 출력만 한다.
7. 실제 콜백이 Context를 무시해도 wrapper가 강제로 종료한다고 설명한다.

각 변경이 어떤 계약을 깨는지 적고,
이미 있는 테스트 중 어떤 검사가 문제를 잡는지 연결합니다.
모든 문제를 race detector 하나로 찾을 수 있다고 답하지 않습니다.

## 완료 검증 과제

**수준: 종합**

다음 명령과 결과를 확인합니다.

```sh
go run ./examples/31-concurrent-batch/cmd/batch
go run ./examples/31-concurrent-batch/cmd/batch -input examples/31-concurrent-batch/testdata/jobs.txt
go run ./examples/31-concurrent-batch/cmd/batch -h
```

실패 예제는 의도적으로 0이 아닌 종료 코드입니다.

```sh
go run ./examples/31-concurrent-batch/cmd/batch -input examples/31-concurrent-batch/testdata/failures.txt
```

go run의 래퍼 종료값과 실제 프로그램 종료 코드를 구분합니다.
정확한 코드 확인은 본문처럼 실행 파일을 빌드해 수행합니다.

### 점검 항목

| 구분 | 확인 |
| --- | --- |
| 기본 실행 | 총 3개, 성공 3개, 시도 합계 4 |
| 파일 입력 | 기본 예제와 같은 결과 |
| 영구 실패 | 성공·실패 결과 함께 출력, 프로그램 코드 1 |
| 잘못된 설정 | 작업 시작 전 코드 2 |
| 이미 취소된 Context 테스트 | 미시작 상태와 시도 0 |
| 출력 실패 테스트 | 성공 코드로 숨기지 않음 |

`-timeout`은 입력을 읽은 뒤의 처리 제한입니다.
임의의 io.Reader의 Read나 Writer의 Write를 강제로 중단하지 않습니다.
표준 입력을 직접 사용하는 경우 EOF와 인터럽트 처리의 경계도 본문에서 확인합니다.

## 제출 전 검증

```sh
go test ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
go vet ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
```

개인 실습은 `_practice` 경로를 명시해 검사합니다.
Linux와 C 컴파일러 환경에서는 `go test -race`도 실행합니다.

- 항목 개수와 시도 횟수를 구분했는가?
- 잘못된 입력에서 콜백을 시작하지 않는가?
- 결과의 입력 순서가 유지되는가?
- 재시도와 취소가 유한하게 끝나는가?
- 함수 반환 전에 내부 작업을 모두 기다리는가?
- 실제로 확인하지 않은 플랫폼 결과를 통과로 표시하지 않았는가?

[본문](../../book/31-concurrent-batch/README.md) ·
[해설](../../solutions/31-concurrent-batch/README.md)
