# 18장 연습문제 — 고루틴과 완료 대기

[본문](../../book/18-goroutines/README.md)의 작업 등록·소유권·종료를 연습합니다.
기본 6문제와 선택 1문제입니다.
[해설](../../solutions/18-goroutines/README.md)은 먼저 구현한 뒤 확인합니다.

## 준비

모든 명령은 저장소 루트에서 실행합니다.
개인 실습은 `_practice/18-goroutines/` 아래에 별도 폴더로 작성합니다.
기존 예제와 해설 파일을 덮어쓰지 않습니다.

Bash:

```sh
mkdir -p _practice/18-goroutines/parallel-labels
if [ ! -e _practice/18-goroutines/parallel-labels/main.go ]; then
  cp examples/18-goroutines/indexed-results/main.go _practice/18-goroutines/parallel-labels/main.go
fi
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/18-goroutines/parallel-labels | Out-Null
if (-not (Test-Path _practice/18-goroutines/parallel-labels/main.go)) {
    Copy-Item examples/18-goroutines/indexed-results/main.go _practice/18-goroutines/parallel-labels/main.go
}
```

입력 슬라이스는 함수 호출이 끝날 때까지 다른 곳에서 변경하지 않습니다.
콜백을 받는 문제에서는 콜백이 여러 고루틴에서 동시에 호출될 수 있습니다.
콜백이 쓰는 공유 상태도 안전해야 합니다.

## 문제 1. 기다림의 의미 예측

**수준: 확인**

다음 질문에 각각 이유와 함께 답합니다.

1. `go printName()`을 호출한 다음 main이 반환하면 이름이 반드시 출력되는가?
2. `Add(1)`을 작업 함수 안에 두면 어떤 실행 순서에서 기다림을 놓치는가?
3. `Done()`을 정상 경로의 마지막 줄에만 두면 어떤 분기가 문제가 되는가?
4. `Wait()`가 반환했다는 사실은 작업이 모두 성공했다는 뜻인가?
5. `cancel()`이 반환했다는 사실은 내부 작업이 모두 끝났다는 뜻인가?
6. 작업마다 서로 다른 `results[index]`에 쓰는 예제는 왜 안전한가?
7. 여러 작업이 같은 `results`에 append하면 왜 상황이 달라지는가?
8. 고루틴 100개는 CPU 코어 100개를 사용한다는 뜻인가?

### 상태표

| 사건 | 등록된 작업 수 | 호출자의 결과 읽기 가능 여부 |
| --- | --- | --- |
| 영값 WaitGroup 생성 | 직접 작성 | 작업 시작 전 상태 |
| Add(2) | 직접 작성 | 직접 작성 |
| 작업 하나의 Done | 직접 작성 | 직접 작성 |
| 나머지 작업의 Done | 직접 작성 | 직접 작성 |
| Wait 반환 | 직접 작성 | 직접 작성 |

카운터는 개념적으로만 추적합니다.
WaitGroup 내부 구현 필드를 읽어 확인하지 않습니다.

### 완료 조건

완료와 성공, 취소와 종료, 순서와 속도를 구분합니다.
“충분히 기다렸으므로 안전하다”를 근거로 사용하지 않습니다.

### 힌트

누가 작업을 등록하고 누가 결과를 쓰는지 나눠 적습니다.
작업 결과와 대기용 카운터는 별개의 정보입니다.

## 문제 2. 순서가 유지되는 이름표

**수준: 변형**

개인 폴더: `_practice/18-goroutines/parallel-labels`

```go
func labels(names []string) ([]string, error)
```

각 이름의 양쪽 공백을 제거하고 영문자를 대문자로 바꿉니다.
한 이름당 고루틴 하나를 시작하되 다음 계약을 지킵니다.

| 항목 | 규칙 |
| --- | --- |
| 입력 개수 | 0~16개 |
| 이름 | TrimSpace 결과가 비어 있으면 오류 |
| 빈 입력 | 길이 0인 결과와 nil 오류 |
| 결과 순서 | 입력 순서 유지 |
| 오류 | nil 결과와 오류 |
| 작업 시작 시점 | 전체 입력 검증 뒤 |
| 반환 시점 | 시작한 모든 작업의 종료 뒤 |

공백 제거와 대문자 변환에는 `strings.TrimSpace`, `strings.ToUpper`를 사용합니다.
한글은 영문 대소문자 변환으로 사라지거나 손상되어서는 안 됩니다.

입력:

```text
[" go ", "test", "한글"]
```

예상 출력:

```text
[GO TEST 한글]
```

### 필수 테스트

- nil 입력과 빈 슬라이스.
- 중복 이름 두 개: 결과가 두 개 유지되어야 합니다.
- 앞뒤 공백과 한글.
- 허용되는 16개와 초과하는 17개.
- 빈 문자열과 공백뿐인 이름.

### 금지 사항

공유 슬라이스에 여러 고루틴이 append하지 않습니다.
테스트 통과를 위해 Sleep을 넣지 않습니다.
출력 순서를 맞추려고 모든 작업을 하나씩 시작하고 바로 기다리지 않습니다.

### 힌트

입력과 같은 길이의 결과 슬라이스를 먼저 만듭니다.
각 입력의 인덱스가 결과를 쓸 권한의 범위가 됩니다.

실행:

```sh
go run ./_practice/18-goroutines/parallel-labels
go test ./_practice/18-goroutines/parallel-labels
```

## 문제 3. 세 가지 잘못된 구조 고치기

**수준: 확인·수정**

다음은 **의도적으로 잘못된 발췌**입니다.
컴파일 오류 또는 레이스를 일반 예제에 추가하지 말고,
무슨 순서에서 문제가 생기는지 설명한 뒤 올바른 형태를 작성합니다.

### A. 늦은 등록

```go
var group sync.WaitGroup
go func() {
    group.Add(1)
    defer group.Done()
    work()
}()
group.Wait()
```

### B. 공유 append

```go
results := []int{}
for _, value := range values {
    group.Go(func() {
        results = append(results, value*value)
    })
}
group.Wait()
```

### C. 대기 객체의 복사

```go
func launch(group sync.WaitGroup) {
    group.Add(1)
    go func() {
        defer group.Done()
        work()
    }()
}
```

`work`와 `values`는 설명용 이름입니다.
각 발췌에 대해 다음을 제출합니다.

1. 문제가 되는 객체와 두 사건.
2. 원래 코드가 성공처럼 보일 수 있는 이유.
3. 수정한 코드 또는 구조 설명.
4. 수정 뒤에도 별도로 확인해야 할 오류·취소 책임.

### 힌트

A는 시작 전에 등록합니다.
B는 결과 칸을 분할하거나 수집자를 하나 둡니다.
C는 대기 객체의 소유자를 정하고 복사하지 않습니다.

## 문제 4. 고정 작업자로 합계 구하기

**수준: 응용**

개인 폴더: `_practice/18-goroutines/parallel-sum`

```go
func sumParts(values []int, workers int) (int, error)
```

| 항목 | 규칙 |
| --- | --- |
| 값 개수 | 0~1000개 |
| 값의 범위 | 0~1000 |
| workers | 1~8 |
| 실제 작업자 수 | min(workers, len(values)) |
| 빈 입력 | 0, nil; 고루틴을 만들 필요 없음 |
| 실패 | 0과 오류 |
| 합계 범위 | 최대 1,000,000 |

작업자 번호가 `worker`이면
`worker, worker+workers, worker+2*workers, ...` 인덱스를 맡깁니다.
각 작업자는 지역 합계를 계산해 자기 부분 결과 칸에 기록합니다.
호출자는 모든 작업이 끝난 뒤 부분 결과를 더합니다.

입력 `10, 20, 30, 40, 50`, 작업자 수 2의 예상 출력:

```text
합계: 150
```

### 필수 테스트

| 입력 | workers | 결과 |
| --- | --- | --- |
| nil | 1 | 0 |
| [10] | 8 | 10 |
| [10,20,30,40,50] | 1~8 각각 | 150 |
| 1000을 1000개 | 8 | 1,000,000 |
| [-1] | 2 | 오류 |
| [1001] | 2 | 오류 |
| 값 1001개 | 2 | 오류 |
| 빈 입력 | 0 또는 9 | 오류 |

빈 입력이어도 잘못된 설정을 성공으로 숨기지 않습니다.
입력 검증 후 빈 입력을 처리합니다.

### 힌트

작업자 수보다 입력이 적으면 작업자를 줄입니다.
부분 합계 슬라이스의 길이는 입력 수가 아니라 실제 작업자 수입니다.
공유 total에 동시에 `+=`하지 않습니다.

## 문제 5. 전체 성공 시에만 변환 결과 반환

**수준: 응용**

개인 폴더: `_practice/18-goroutines/parallel-transform`

```go
func transformAll(
    ctx context.Context,
    inputs []string,
    transform func(context.Context, string) (string, error),
) ([]string, error)
```

### 계약

- Context는 nil이 아니어야 합니다.
- 입구에서 취소되었으면 nil 결과와 ctx.Err를 반환하며 콜백을 호출하지 않습니다.
- 입력은 최대 16개이며 nil 콜백은 오류입니다.
- 각 작업은 콜백 전후에 취소를 확인합니다.
- 모든 작업이 종료된 후 반환합니다.
- 하나라도 실패하면 결과는 nil이며 원래 오류를 `errors.Is`로 찾을 수 있어야 합니다.
- 성공 결과는 입력 순서로 반환합니다.
- 서로 다른 항목의 실패가 함께 있으면 둘 다 보존합니다.
- 개별 실패로 다른 작업을 즉시 취소하지는 않습니다.

정상 입력 `go`, `test`를 대문자로 바꾼 뒤,
공백 입력을 실패시키는 호출도 추가합니다.

예상 출력:

```text
정상 결과: [GO TEST]
실패 결과 없음: true
오류 확인: true
```

### 필수 테스트

두 개의 서로 다른 sentinel 오류를 반환하는 콜백을 만듭니다.
최종 오류에서 두 sentinel 모두 `errors.Is`가 true인지 확인합니다.
별도로 이미 취소된 Context와 콜백 안에서 취소되는 상황을 검사합니다.

콜백이 기록하는 테스트 자료도 인덱스별로 나누거나 동기화합니다.
테스트를 위해 공유 append를 만들면 제품 코드가 안전해도 테스트에 레이스가 생깁니다.

### 힌트

성공값과 오류를 각각 인덱스별로 저장합니다.
Wait 뒤에서 errors.Join을 사용합니다.
작업이 끝났는지 확인하기 전에 오류를 보고 반환하지 않습니다.

## 문제 6. 취소해도 끝까지 합류하는 단일 작업

**수준: 응용**

개인 폴더: `_practice/18-goroutines/joined-task`

```go
func runJoined(ctx context.Context, task func(context.Context) error) error
```

task 하나를 고루틴에서 실행하고 반드시 종료까지 기다립니다.
이미 취소되었다면 task를 호출하지 않습니다.
nil task는 오류입니다.
task 오류를 그대로 보존하고,
task가 nil을 반환했지만 Context가 취소되었다면 취소 오류를 반환합니다.

### 수명 테스트

1. task가 시작 신호를 알립니다.
2. 테스트가 시작 신호를 기다린 뒤 cancel합니다.
3. task가 취소를 보고 정리한 뒤 종료 신호를 알립니다.
4. runJoined가 반환합니다.
5. 테스트가 이미 종료 신호가 왔는지 확인합니다.

시작·종료 신호는 `chan struct{}`를 닫아 전달할 수 있습니다.
자세한 채널 문법은 다음 장에서 읽어도 됩니다.
각 대기에는 테스트가 무한히 멈추지 않도록 실패용 제한 시간을 둡니다.
성공 판단을 Sleep에 의존시키지 않습니다.

### 완료 조건

취소를 요청한 즉시 호출자만 반환하는 구현은 실패입니다.
task가 Context를 무시하면 강제로 중단할 수 없다는 한계도 설명합니다.

## 문제 7. 선택 — 완료 순서와 표시 순서

세 작업이 각각 다른 시작 대기 신호를 받도록 만듭니다.
테스트에서 2→0→1 순서로 작업을 진행시키더라도
최종 결과는 입력 순서 0→1→2로 나와야 합니다.

완료 순서대로 로그를 출력하는 경우와 입력 순서대로 결과를 반환하는 경우를 비교합니다.
두 요구는 다르며 하나의 순서로 모두 표현할 필요가 없습니다.

[19장](../../book/19-channels-select/README.md)의 채널을 배운 뒤 풀어도 됩니다.
고루틴 ID, 실행 시간 차이, Sleep 길이로 순서를 만들지 않습니다.

## 제출 전 확인

```sh
go fmt ./_practice/18-goroutines/parallel-labels
go vet ./_practice/18-goroutines/parallel-labels
go test ./_practice/18-goroutines/parallel-labels
```

다른 문제도 해당 경로로 실행합니다.
`_practice`처럼 밑줄로 시작하는 폴더는 일반 `./...` 탐색에서 제외되므로
개인 실습 경로를 명시합니다.
Linux와 C 컴파일러 환경에서는 각 문제에 `go test -race`도 실행합니다.

- 모든 작업을 등록한 뒤 완료를 기다리는가?
- 어떤 입력도 무한한 수의 고루틴을 만들지 않는가?
- 결과 저장 칸의 소유자가 명확한가?
- 실패 결과와 빈 성공 결과를 구분하는가?
- 취소와 실제 종료를 각각 확인하는가?

[본문으로 돌아가기](../../book/18-goroutines/README.md) ·
[해설 보기](../../solutions/18-goroutines/README.md)
