# 31. 종합 실습: 동시 배치 처리

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 종합 응용 | Context·고루틴·채널·동기화·CLI·테스트 | Go, 터미널, 이 저장소 |

입력 목록을 여러 작업으로 처리한다고 해서
모든 항목에 고루틴을 하나씩 붙이는 것만으로 배치 프로그램이 완성되지는 않습니다.
동시에 실행할 수 있는 수, 실패한 항목의 재시도,
취소 시 남길 결과, 모든 작업의 종료를 함께 정해야 합니다.

이번 종합 실습은 문자열 목록을 정규화하는 작은 배치를 만듭니다.
처리 내용은 의도적으로 작게 두고,
그 주변의 **입력 검증 → 제한된 작업자 → 재시도 → 결과 수집 → 보고서**를 완성합니다.
외부 네트워크나 실제 계정 없이 실패와 취소를 재현할 수 있습니다.

앞 장에서 만든 HTTP API와 달리,
이번 프로그램은 주어진 작업 묶음을 처리하고 종료합니다.
서비스의 한 요청이나 정기 배치 안에서도 사용할 수 있도록
실행 엔진과 명령행 입출력을 별도 패키지로 나눕니다.

## 학습 목표

- 입력 개수와 동시 작업자 수를 분리하고 자원 상한을 검증한다.
- 실행 순서와 무관하게 입력 위치에 대응하는 결과를 반환한다.
- 다시 시도할 오류만 구분하고 시도 횟수·대기 시간에 상한을 둔다.
- 취소 시 새 작업을 줄이고 이미 시작한 작업을 모두 기다린다.
- 성공·실패·취소·미시작 항목을 설명할 수 있는 CLI 보고서를 만든다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
기본 실행에는 파일·네트워크·DB가 필요하지 않습니다.
명령행의 기본 입력은 프로그램 안에 준비되어 있습니다.

선수 학습은 [16장 CLI](../16-cli/README.md),
[17장 Context](../17-context/README.md),
[18장 고루틴](../18-goroutines/README.md),
[19장 채널](../19-channels-select/README.md),
[20장 동기화](../20-sync/README.md)입니다.
출력과 입력 오류는 [14장](../14-files-io/README.md),
구조화된 자료는 [15장](../15-data-time/README.md)과 연결합니다.

| 단계 | 소스 | 목적 |
| --- | --- | --- |
| 1 | [sequential](../../examples/31-concurrent-batch/sequential/main.go) | 처리 계약을 순차 함수로 확인 |
| 2 | [bounded](../../examples/31-concurrent-batch/bounded/main.go) | 고정 작업자와 순서 보존 |
| 3 | [limits](../../examples/31-concurrent-batch/limits/main.go) | 실행 전 설정·입력 검증 |
| 4 | [retries](../../examples/31-concurrent-batch/retries/main.go) | 일시 오류와 영구 오류 구분 |
| 5 | [cancellation](../../examples/31-concurrent-batch/cancellation/main.go) | 취소·미시작·부분 성공 |
| 6 | [report](../../examples/31-concurrent-batch/report/main.go) | 완료 결과를 요약 |
| 완성 | [cmd/batch](../../examples/31-concurrent-batch/cmd/batch/main.go) | 파일·표준 입력·JSON·종료 코드 |
| 공통 | [batch](../../examples/31-concurrent-batch/batch/batch.go) | 재사용 가능한 실행 엔진 |

## 1. 먼저 성공과 실패의 계약을 정하기

### 1.1 이번 프로그램의 범위

실행 엔진은 문자열 하나를 받아 문자열 하나 또는 오류를 반환하는 함수를 실행합니다.
엔진은 그 문자열이 URL인지, 파일명인지, 업무 식별자인지 해석하지 않습니다.
처리 내용은 Processor 콜백에 맡깁니다.

최종 CLI는 양끝 공백을 제거하고 영문자를 대문자로 바꿉니다.
실패 경로를 실습하기 위해 두 개의 접두사를 해석합니다.

| CLI 입력 | 동작 |
| --- | --- |
| 일반 문자열 | 공백 제거 후 대문자로 변환 |
| `retry:내용` | 첫 시도에 일시 오류, 다음 시도에 내용 처리 |
| `fail:내용` | 영구 오류로 실패 |
| 공백뿐인 문자열 | 입력 내용 오류 |

이 접두사들은 **CLI의 학습용 시뮬레이션**입니다.
공통 batch 패키지는 접두사를 모르며 ErrTemporary라는 오류 표지만 봅니다.
실제 서비스에서 사용자 입력에 이 표기를 요구하라는 설계가 아닙니다.

### 1.2 자원과 입력의 상한

| 대상 | 허용 범위 | 이유 |
| --- | --- | --- |
| 입력 개수 | 최대 10,000개 | 결과와 입력을 메모리에 유지하는 범위 제한 |
| 입력 문자열 | 유효한 UTF-8, 최대 4,096바이트 | 항목 크기 제한 |
| 성공 출력 | 유효한 UTF-8, 최대 4,096바이트 | 결과 자료 크기 제한 |
| 작업자 수 | 1~32명 | 동시에 실행할 콜백 수 제한 |
| 항목별 최대 시도 | 1~5회 | 무한 재시도 방지 |
| 기본 재시도 대기 | 0~1초 | 학습 실습의 시간 상한 |
| 한 번의 재시도 대기 | 최대 1초 | 지수 증가의 상한 |
| CLI 처리 제한 시간 | 0 초과, 최대 1분 | 호출 전체의 처리 예산 |

이 값들은 Go 자체의 제한이 아니라 교재의 설계 선택입니다.
업무에 맞게 바꾸려면 메모리·처리 시간·외부 서비스 수용량을 함께 검토합니다.

입력 개수가 커도 동시에 실행되는 콜백 수는 작업자 수를 넘지 않습니다.
그러나 입력과 결과 전체를 보관하므로 메모리 사용이 입력 수와 무관한 것은 아닙니다.
진정한 스트림 처리와 디스크 체크포인트는 이 실습의 범위를 넘어섭니다.

### 1.3 반환 정책

| 상황 | 결과 슬라이스 | 반환 오류 |
| --- | --- | --- |
| 잘못된 설정·입력·nil 콜백 | nil | 검증 오류 |
| 정상 빈 입력 | 길이 0인 결과 | nil |
| 모두 성공 | 입력과 같은 길이 | nil |
| 하나 이상 항목 실패 | 입력과 같은 길이 | BatchError |
| 호출 Context 취소·마감 | 입력과 같은 길이 | ctx.Err |

취소되어도 완료된 성공 결과는 보존합니다.
시작하지 못한 항목도 결과에 남기고 canceled, attempts=0으로 표시합니다.
진행 중이던 작업은 반환한 뒤 상태를 기록합니다.

이 정책은 [18장](../18-goroutines/README.md)의 전체 성공 시에만 값 반환 정책과 다릅니다.
배치에서는 어떤 항목이 성공했고 무엇을 다시 검토해야 하는지 기록할 필요가 있기 때문입니다.

### 1.4 검증과 취소의 우선순위

설정과 입력 검증을 먼저 마칩니다.
그 다음 Context 상태를 확인하고 작업을 시작합니다.
따라서 잘못된 설정과 이미 취소된 Context가 함께 오면 설정 오류를 반환합니다.

작업을 실행한 뒤에는 호출 Context의 취소를 전체 반환 오류로 우선합니다.
개별 항목의 기존 실패·성공은 결과 슬라이스에 남아 있습니다.
오류 하나만 읽고 상세 결과를 버리면 배치의 일부 정보를 잃을 수 있습니다.

## 2. 처리 규칙을 순차 실행으로 먼저 확인하기

동시성을 추가하기 전에 문자열 처리와 오류의 의미를 확인합니다.
이 첫 단계는 입력 최대 1000개인 작은 순차 버전입니다.
완성 엔진에서는 앞의 요구사항 표에 맞춰 검증과 결과 계약을 확장합니다.

소스: [sequential/main.go](../../examples/31-concurrent-batch/sequential/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

func sequential(ctx context.Context, inputs []string) ([]string, error) {
	if len(inputs) > 1000 {
		return nil, errors.New("입력은 1000개까지 허용합니다")
	}
	results := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		value := strings.TrimSpace(input)
		if value == "" {
			return nil, errors.New("빈 항목")
		}
		results = append(results, strings.ToUpper(value))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
func main() {
	values, err := sequential(context.Background(), []string{"go", " test ", "한글"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("순차 결과:", values)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/31-concurrent-batch/sequential
```

예상 출력:

```text
순차 결과: [GO TEST 한글]
```

### 2.1 먼저 확인할 것

공백 제거와 대문자 변환이 기대한 결과를 내는지 확인합니다.
입력이 비면 오류인지, 항목 일부가 실패하면 전체 결과를 버릴지 정합니다.
순차 버전은 하나라도 실패하면 nil 결과를 반환합니다.

동시성을 추가한 뒤 처리 규칙까지 함께 바꾸면
문제가 변환 함수인지 작업 분배인지 구별하기 어려워집니다.
작은 처리 규칙을 먼저 확인하는 이유입니다.

### 2.2 Context가 하는 일

각 반복 전에 취소 상태를 확인합니다.
이 문자열 처리는 짧아서 중간 중단 API가 필요하지 않습니다.
실제 HTTP·DB 호출을 넣는다면 같은 Context를 하위 API에도 전달해야 합니다.

Context를 매개변수로 받는 것만으로 모든 작업이 자동 취소되지는 않습니다.
콜백의 종료 협조가 필요하다는 조건은 완성 엔진에서도 유지됩니다.

## 3. 입력 순서를 유지하며 제한된 작업자로 실행하기

공통 batch 패키지를 사용해 두 명의 작업자로 세 항목을 처리합니다.
실제 완료 순서는 정하지 않지만 반환 결과는 원래 입력 순서를 유지합니다.

소스: [bounded/main.go](../../examples/31-concurrent-batch/bounded/main.go)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	results, err := batch.Run(context.Background(), []string{"c", "a", "한글"}, batch.Config{Workers: 2, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) { return strings.ToUpper(job.Input), ctx.Err() })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, result := range results {
		fmt.Printf("%d: %s (%s)\n", result.Job.Index, result.Output, result.State)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/31-concurrent-batch/bounded
```

예상 출력:

```text
0: C (succeeded)
1: A (succeeded)
2: 한글 (succeeded)
```

### 3.1 Job의 인덱스

각 Job에는 Index와 Input을 함께 둡니다.
작업자는 results[job.Index]에 자기 결과만 기록합니다.
동시에 같은 결과 칸을 쓰는 작업자가 없습니다.

고루틴 안에서 공유 결과 슬라이스에 append하지 않습니다.
결과의 길이와 위치는 실행 전에 정하고,
호출자는 모든 작업자를 기다린 뒤 결과를 읽습니다.

### 3.2 작업자 수와 입력 수

실제 작업자 수는 min(Workers, len(inputs))입니다.
세 입력에 Workers=2이면 두 고루틴이 여러 항목을 나누어 처리합니다.
입력이 하나이면 불필요한 작업자들을 더 만들지 않습니다.

작업자 수를 늘리는 것이 항상 성능 개선은 아닙니다.
이번 문자열 변환은 매우 작으므로 순차 처리보다 비용이 늘 수 있습니다.
학습 목표는 자원 상한과 수명의 설계입니다.

## 4. 잘못된 입력에서 작업을 시작하지 않기

작업자 수가 0이거나 항목이 너무 크면 콜백을 호출하기 전에 거부합니다.
처리 도중 설정이 잘못되었다는 사실을 발견해 일부 항목만 실행하지 않습니다.

소스: [limits/main.go](../../examples/31-concurrent-batch/limits/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	calls := 0
	process := func(context.Context, batch.Job) (string, error) { calls++; return "ok", nil }
	results, err := batch.Run(context.Background(), []string{"go"}, batch.Config{Workers: 0, Attempts: 1}, process)
	fmt.Println("설정 거부:", results == nil && errors.Is(err, batch.ErrConfig))
	results, err = batch.Run(context.Background(), []string{strings.Repeat("x", batch.MaxItemBytes+1)}, batch.Config{Workers: 1, Attempts: 1}, process)
	fmt.Println("입력 거부:", results == nil && errors.Is(err, batch.ErrInput))
	fmt.Println("실행된 작업:", calls)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/31-concurrent-batch/limits
```

예상 출력:

```text
설정 거부: true
입력 거부: true
실행된 작업: 0
```

### 4.1 무엇을 먼저 검증하는가?

Context·콜백·설정의 유효성, 입력 개수, 각 입력의 바이트 길이와 UTF-8을 확인합니다.
그 뒤 결과 저장 공간을 만들고 Context의 취소 상태를 확인합니다.

UTF-8 문자열의 바이트 길이와 글자 수는 다릅니다.
4096바이트 제한은 문자 개수 제한이 아닙니다.
한글을 포함한 입력은 [07장](../07-strings-unicode/README.md)의 구분을 적용합니다.

### 4.2 콜백이 만드는 결과도 경계다

외부 입력만 작아도 콜백이 큰 출력이나 잘못된 UTF-8을 만들 수 있습니다.
성공 반환값도 같은 크기·인코딩 계약을 확인합니다.
잘못된 성공 출력은 ErrOutput을 가진 failed 결과가 됩니다.

이 검사는 이미 콜백 내부에서 발생한 할당을 되돌려 제한하는 기능은 아닙니다.
콜백 자체의 메모리·시간 사용은 콜백 구현이 관리해야 합니다.

## 5. 공통 실행 엔진 읽기

소스: [batch/batch.go](../../examples/31-concurrent-batch/batch/batch.go)

다음은 엔진의 전체 코드입니다.
아래 설명에서는 검증, 분배, 실행, 합류의 순서로 나누어 읽습니다.

```go
// Package batch runs a bounded number of cooperative workers and preserves input order.
package batch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	MaxItems      = 10000
	MaxItemBytes  = 4096
	MaxWorkers    = 32
	MaxAttempts   = 5
	MaxRetryDelay = time.Second
)

var (
	ErrConfig    = errors.New("invalid batch configuration")
	ErrInput     = errors.New("invalid batch input")
	ErrTemporary = errors.New("temporary failure")
	ErrOutput    = errors.New("invalid processor output")
)

type State string

const (
	Succeeded State = "succeeded"
	Failed    State = "failed"
	Canceled  State = "canceled"
)

type Config struct {
	Workers    int
	Attempts   int
	RetryDelay time.Duration
}

type Job struct {
	Index int
	Input string
}

type Result struct {
	Job      Job
	Output   string
	State    State
	Attempts int
	Err      error
}

// Processor must be safe for concurrent calls, cooperate with ctx, and not panic.
// A job's retries are sequential, and returning nil means that attempt succeeded.
type Processor func(ctx context.Context, job Job) (string, error)

// BatchError reports item failures after all workers have joined.
type BatchError struct {
	Count  int
	causes []error
}

func (e *BatchError) Error() string   { return fmt.Sprintf("%d batch item(s) failed", e.Count) }
func (e *BatchError) Unwrap() []error { return append([]error(nil), e.causes...) }

func ValidateConfig(cfg Config) error {
	if cfg.Workers < 1 || cfg.Workers > MaxWorkers {
		return fmt.Errorf("%w: workers must be 1..%d", ErrConfig, MaxWorkers)
	}
	if cfg.Attempts < 1 || cfg.Attempts > MaxAttempts {
		return fmt.Errorf("%w: attempts must be 1..%d", ErrConfig, MaxAttempts)
	}
	if cfg.RetryDelay < 0 || cfg.RetryDelay > MaxRetryDelay {
		return fmt.Errorf("%w: retry delay must be 0..%s", ErrConfig, MaxRetryDelay)
	}
	return nil
}

// Run returns one result per input after every worker has stopped.
// Invalid arguments return nil results. Cancellation preserves completed results
// and marks unfinished items canceled. The caller must not mutate inputs during Run.
func Run(ctx context.Context, inputs []string, cfg Config, process Processor) ([]Result, error) {
	return run(ctx, inputs, cfg, process, waitRetry)
}

func run(ctx context.Context, inputs []string, cfg Config, process Processor,
	wait func(context.Context, time.Duration) error) ([]Result, error) {
	if ctx == nil || process == nil || wait == nil {
		return nil, fmt.Errorf("%w: context, processor and wait function are required", ErrConfig)
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	if len(inputs) > MaxItems {
		return nil, fmt.Errorf("%w: at most %d items", ErrInput, MaxItems)
	}
	for i, input := range inputs {
		if len(input) > MaxItemBytes || !utf8.ValidString(input) {
			return nil, fmt.Errorf("%w: item %d must be valid UTF-8 and at most %d bytes", ErrInput, i, MaxItemBytes)
		}
	}
	results := make([]Result, len(inputs))
	for i, input := range inputs {
		results[i].Job = Job{Index: i, Input: input}
	}
	if err := ctx.Err(); err != nil {
		markUnfinished(results, err)
		return results, err
	}
	if len(inputs) == 0 {
		return results, nil
	}
	workers := min(cfg.Workers, len(inputs))
	jobs := make(chan Job)
	var group sync.WaitGroup
	for range workers {
		group.Go(func() {
			for job := range jobs {
				results[job.Index] = execute(ctx, job, cfg, process, wait)
			}
		})
	}
	// This goroutine is the only sender and the only closer of jobs.
dispatch:
	for i, input := range inputs {
		if ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- Job{Index: i, Input: input}:
		}
	}
	close(jobs)
	group.Wait()
	if err := ctx.Err(); err != nil {
		markUnfinished(results, err)
		return results, err
	}
	failures := []error{}
	for _, result := range results {
		if result.State == Failed {
			failures = append(failures, fmt.Errorf("item %d: %w", result.Job.Index, result.Err))
		}
	}
	if len(failures) > 0 {
		return results, &BatchError{Count: len(failures), causes: failures}
	}
	return results, nil
}

func execute(ctx context.Context, job Job, cfg Config, process Processor,
	wait func(context.Context, time.Duration) error) Result {
	result := Result{Job: job}
	for attempt := 1; attempt <= cfg.Attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			result.State, result.Err = Canceled, err
			return result
		}
		result.Attempts = attempt
		value, err := process(ctx, job)
		if err == nil {
			if len(value) > MaxItemBytes || !utf8.ValidString(value) {
				result.State, result.Err = Failed, ErrOutput
				return result
			}
			result.State, result.Output = Succeeded, value
			return result
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			result.State, result.Err = Canceled, ctxErr
			return result
		}
		if !errors.Is(err, ErrTemporary) || attempt == cfg.Attempts {
			result.State, result.Err = Failed, err
			return result
		}
		if err := wait(ctx, retryDelay(cfg.RetryDelay, attempt)); err != nil {
			result.State, result.Err = Canceled, err
			return result
		}
	}
	// ValidateConfig makes the loop execute at least once.
	return result
}

func markUnfinished(results []Result, err error) {
	for i := range results {
		if results[i].State == "" {
			results[i].State, results[i].Err = Canceled, err
		}
	}
}

func retryDelay(base time.Duration, attempt int) time.Duration {
	delay := base
	for i := 1; i < attempt; i++ {
		if delay >= MaxRetryDelay/2 {
			return MaxRetryDelay
		}
		delay *= 2
	}
	return delay
}

func waitRetry(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay == 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}
```

### 5.1 타입이 드러내는 책임

| 타입 | 책임 |
| --- | --- |
| Config | 실행 개수·재시도 횟수·대기 정책 |
| Job | 원래 입력 위치와 값 |
| Processor | 실제 한 항목의 처리 |
| Result | 한 항목의 상태·시도 횟수·값·오류 |
| BatchError | 모든 작업을 마친 뒤 항목 실패들을 결합 |

Config를 함수마다 명시적으로 전달합니다.
전역 작업자 수나 전역 취소 상태를 두지 않습니다.
서로 다른 배치를 다른 설정으로 호출할 수 있습니다.

### 5.2 jobs 채널의 소유권

Run을 수행하는 고루틴이 jobs에 보내는 유일한 주체입니다.
같은 고루틴이 전송을 끝낸 뒤 jobs를 닫습니다.
작업자들은 수신만 하며 jobs를 닫지 않습니다.

| 객체 | 수정·닫기 소유자 |
| --- | --- |
| jobs 전송 | Run의 분배 반복 |
| jobs 닫기 | 같은 분배 반복을 마친 Run |
| results[i] | 해당 Job을 받은 작업자 하나 |
| 최종 상태의 빈 칸 정리 | 모든 작업자 합류 후 Run |
| Context 취소 | 호출자 또는 호출자의 상위 수명 |

결과를 위한 별도 채널은 사용하지 않습니다.
입력 개수가 이미 정해져 있어 결과 슬롯을 나누는 방식이 간결하기 때문입니다.

### 5.3 버퍼 없는 접수와 역압력

jobs는 버퍼 없는 채널입니다.
작업자가 다음 입력을 받을 준비가 되어야 분배가 진행됩니다.
모든 작업자가 바쁘면 분배자도 기다립니다.
이를 **역압력(backpressure)**의 한 형태로 볼 수 있습니다.

분배자는 기다리는 동안에도 Context 취소를 선택할 수 있습니다.
취소되면 분배 반복에서 빠져나와 jobs를 닫고 모든 작업자를 기다립니다.

채널을 크게 버퍼링해 모든 입력을 미리 쌓지 않습니다.
입력과 결과는 이미 메모리에 있지만,
접수된 작업의 진행 관계를 제한된 작업자 수로 유지합니다.

### 5.4 작업자는 계속 같은 고루틴을 사용한다

각 작업자는 jobs를 range로 읽고 처리합니다.
하나가 끝나면 같은 고루틴에서 다음 Job을 받습니다.
재시도도 같은 작업자가 순서대로 수행합니다.

따라서 하나의 Job에 대해 콜백 두 번이 동시에 실행되지는 않습니다.
서로 다른 Job의 콜백들은 동시에 실행될 수 있습니다.
콜백이 공유 맵이나 카운터를 쓰면 [20장](../20-sync/README.md)의 보호 규칙이 필요합니다.

### 5.5 오류가 있어도 합류한다

한 항목의 영구 오류는 그 항목을 failed로 만들고 다음 항목으로 진행합니다.
개별 실패 하나가 다른 작업을 자동으로 취소하지 않습니다.
이것은 전체 배치에서 가능한 결과를 최대한 모으는 정책입니다.

작업자들이 모두 종료한 뒤 실패 원인을 입력 순서로 결합합니다.
BatchError는 Count와 Unwrap을 제공하므로
오류 개수와 원래 원인을 errors.As·errors.Is로 확인할 수 있습니다.

### 5.6 취소 후 빈 슬롯 정리

취소 시 모든 입력이 jobs로 전달되지는 않았을 수 있습니다.
전달되지 않은 항목은 아직 State가 비어 있습니다.
모든 작업자를 기다린 뒤에만 이 칸들을 canceled로 바꿉니다.

작업자와 동시에 결과를 수정하지 않습니다.
이 순서를 지켜야 작업자가 쓴 성공 결과를 취소 정리가 덮어쓰는 일이 없습니다.

### 5.7 호출자와 콜백의 계약

- 호출자는 Run이 반환할 때까지 inputs를 다른 고루틴에서 바꾸지 않습니다.
- Processor는 여러 Job의 동시 호출에 안전해야 합니다.
- Processor는 전달받은 Context를 관찰하고 유한한 시간에 반환해야 합니다.
- 예상 가능한 실패는 오류로 반환하고 패닉하지 않습니다.
- 같은 Job의 재시도에서 외부 효과가 중복되어도 괜찮은지 별도로 판단합니다.

Run은 임의의 함수를 강제로 종료하거나 외부 부작용을 되돌리지 않습니다.
이 조건을 만족하지 않는 콜백을 고루틴으로 감싸서 숨기지 않습니다.

## 6. 재시도 가능한 실패만 다시 실행하기

정상 항목, 첫 번만 일시 실패하는 항목, 계속 영구 실패하는 항목을 함께 실행합니다.
이 예제는 예상한 항목 실패를 검사하는 실습이므로 실패를 출력으로 확인한 뒤 정상 종료합니다.
완성 CLI의 종료 코드 정책과는 구분합니다.

소스: [retries/main.go](../../examples/31-concurrent-batch/retries/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	calls := make([]int, 3)
	results, err := batch.Run(context.Background(), []string{"ok", "temporary", "permanent"}, batch.Config{Workers: 2, Attempts: 3}, func(ctx context.Context, job batch.Job) (string, error) {
		calls[job.Index]++
		if job.Input == "temporary" && calls[job.Index] < 2 {
			return "", fmt.Errorf("busy: %w", batch.ErrTemporary)
		}
		if job.Input == "permanent" {
			return "", errors.New("bad input")
		}
		return strings.ToUpper(job.Input), ctx.Err()
	})
	for _, result := range results {
		fmt.Printf("%d: %s attempts=%d\n", result.Job.Index, result.State, result.Attempts)
	}
	var failure *batch.BatchError
	fmt.Println("항목 실패 보고:", errors.As(err, &failure))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/31-concurrent-batch/retries
```

예상 출력:

```text
0: succeeded attempts=1
1: succeeded attempts=2
2: failed attempts=1
항목 실패 보고: true
```

### 6.1 errors.Is로 일시 오류 표지 찾기

엔진은 `errors.Is(err, ErrTemporary)`가 true일 때만 다시 시도합니다.
오류 문자열에 temporary라는 단어가 들어 있는지 검색하지 않습니다.
래핑된 오류에서도 의미를 보존하기 위해 `%w`와 errors.Is를 사용합니다.

모든 네트워크 오류나 모든 상태 코드가 자동으로 재시도 대상은 아닙니다.
실제 Processor에서 재시도가 가능한 원인을 분류해 ErrTemporary로 감싸야 합니다.
인증 실패·입력 오류·영구적인 규칙 위반은 반복해도 해결되지 않을 수 있습니다.

### 6.2 Attempts는 콜백 호출 횟수다

최대 시도 횟수에는 첫 실행도 포함됩니다.
Attempts=3이면 첫 실행과 최대 두 번의 재시도입니다.
재시도만 세 번 더 한다는 뜻이 아닙니다.

미시작 취소 항목은 Attempts=0입니다.
첫 시도에서 영구 오류가 나면 1입니다.
한 번 실패하고 다음에 성공하면 2입니다.

### 6.3 지수 대기와 상한

기본 대기가 100ms일 때 첫 재시도 전에 100ms,
둘째 재시도 전에 200ms, 셋째 전에 400ms처럼 늘립니다.
각 대기는 최대 1초로 제한합니다.

| 실패한 시도 | 다음 시도 전 대기, 기본 400ms |
| --- | --- |
| 1 | 400ms |
| 2 | 800ms |
| 3 | 1s |
| 4 | 1s |

예제의 기본값은 0이어서 출력 확인에 불필요한 대기가 없습니다.
이 값이 실제 외부 서비스의 장애 대응에 적절하다는 뜻은 아닙니다.
실제 호출에서는 요청 예산·서버의 안내·부하를 고려해야 합니다.

### 6.4 대기도 취소 가능한 작업이다

대기 중에는 Timer와 ctx.Done을 select로 기다립니다.
취소가 오면 다음 시도를 시작하지 않고 반환합니다.
단순 Sleep은 대기 중 취소를 관찰하지 못하므로 사용하지 않습니다.

테스트는 실제로 수백 ms를 기다리지 않고
내부 대기 함수를 주입해 요청한 Duration을 기록합니다.
이 주입 지점은 테스트용 내부 함수이며 공개 Run은 실제 타이머를 사용합니다.

### 6.5 재시도와 중복 효과

외부 요청이 시간 초과되었다고 그 작업이 반드시 반영되지 않았다고 단정할 수 없습니다.
다시 호출하면 같은 효과가 두 번 생길 수 있습니다.

이번 문자열 변환은 같은 입력을 다시 처리해도 외부 부작용이 없습니다.
실제 송금·메일 전송·파일 업로드 같은 작업은
업무 식별자와 중복 처리 정책, 결과 확인 방법을 별도로 설계해야 합니다.
재시도 엔진이 정확히 한 번의 외부 반영을 보장하는 것은 아닙니다.

## 7. 취소된 항목과 미시작 항목 구분하기

설명을 결정적으로 만들기 위해 이 예제만 작업자 한 명을 사용합니다.
첫 항목이 성공하고, 두 번째 항목의 콜백이 취소를 요청합니다.
마지막 항목은 콜백을 실행하지 않습니다.

소스: [cancellation/main.go](../../examples/31-concurrent-batch/cancellation/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results, err := batch.Run(ctx, []string{"first", "cancel", "last"}, batch.Config{Workers: 1, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) {
		if job.Index == 1 {
			cancel()
			return "", ctx.Err()
		}
		return job.Input, nil
	})
	for _, result := range results {
		fmt.Printf("%d: %s attempts=%d\n", result.Job.Index, result.State, result.Attempts)
	}
	fmt.Println("취소 오류:", errors.Is(err, context.Canceled))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/31-concurrent-batch/cancellation
```

예상 출력:

```text
0: succeeded attempts=1
1: canceled attempts=1
2: canceled attempts=0
취소 오류: true
```

### 7.1 부분 성공의 의미

0번 항목의 성공과 출력은 보존합니다.
1번 항목은 콜백을 한 번 호출했으므로 attempts=1인 canceled입니다.
2번 항목은 호출하지 않았으므로 attempts=0인 canceled입니다.

canceled라는 한 상태만 보고 외부 작업이 전혀 시작되지 않았다고 단정하지 않습니다.
시도 횟수와 실제 Processor의 부작용 계약을 함께 확인해야 합니다.

### 7.2 작업자가 여러 명이면 무엇이 달라지는가?

취소 시 이미 여러 항목이 진행 중일 수 있습니다.
따라서 취소 이전에 성공한 항목 수나 시도 횟수는 실행 순서에 따라 달라질 수 있습니다.
데모의 정확한 상태 배열을 다중 작업자의 일반적인 정답으로 사용하지 않습니다.

보장하는 것은 작업자 수의 상한,
입력 위치와 결과 위치의 대응,
반환 전 모든 작업자의 종료입니다.

### 7.3 select는 취소 우선순위가 아니다

분배 전에 ctx.Err를 확인해 이미 취소된 작업을 줄입니다.
그 직후 취소와 jobs 전송이 동시에 준비될 가능성은 있습니다.
case를 위에 쓴다고 취소가 반드시 먼저 선택되는 것은 아닙니다.

작업자의 execute도 콜백 전에 ctx.Err를 확인합니다.
그 확인 이후 콜백 호출 사이의 모든 경주를 없애는 강제 중단 장치는 아닙니다.
협력적으로 취소를 관찰하고 완료까지 기다린다는 계약을 유지합니다.

### 7.4 전체 마감과 항목 마감

Run에 전달된 Context가 끝나면 전체 배치가 중단됩니다.
콜백 내부에서 별도 자식 Context의 제한 시간을 사용하면
그 한 항목만 실패하고 다른 항목은 계속 진행하도록 만들 수 있습니다.

이때 오류 체인에 DeadlineExceeded가 있다고
전체 배치 Context까지 반드시 마감된 것은 아닙니다.
어느 Context의 상태인지와 결과의 State를 함께 봅니다.
연습문제에서 항목별 제한 시간을 구현합니다.

## 8. 완료 결과를 보고서로 바꾸기

결과의 상태와 시도 횟수를 합쳐 요약합니다.
작업이 끝난 뒤 한 곳에서 보고서를 만들므로 공유 출력 경쟁이 없습니다.

소스: [report/main.go](../../examples/31-concurrent-batch/report/main.go)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	results, err := batch.Run(context.Background(), []string{"go", "한글"}, batch.Config{Workers: 2, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) { return strings.ToUpper(job.Input), ctx.Err() })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	report := batch.NewReport(results)
	fmt.Printf("총계=%d 성공=%d 실패=%d 취소=%d 시도=%d\n", report.Summary.Total, report.Summary.Succeeded, report.Summary.Failed, report.Summary.Canceled, report.Summary.Attempts)
	for _, item := range report.Items {
		fmt.Printf("%d %s -> %s\n", item.Index, item.Input, item.Output)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/31-concurrent-batch/report
```

예상 출력:

```text
총계=2 성공=2 실패=0 취소=0 시도=2
0 go -> GO
1 한글 -> 한글
```

### 8.1 합계의 두 기준

Total은 입력 항목 수입니다.
Attempts는 실제 Processor 호출 횟수의 합계입니다.
재시도가 있으면 Attempts가 Total보다 클 수 있고,
미시작 취소가 많으면 Total보다 작을 수도 있습니다.

성공·실패·취소 수의 합은 정상적으로 반환된 결과의 Total과 같습니다.
출력 항목은 원래 입력 순서입니다.
작업자가 먼저 끝낸 순서대로 화면에 섞어 출력하지 않습니다.

### 8.2 error를 그대로 JSON으로 보내지 않는 이유

error 인터페이스를 직접 직렬화하면 원하는 오류 설명이 나오지 않을 수 있습니다.
보고서용 Item에는 Error 문자열 필드를 따로 둡니다.
기계가 읽을 상태는 State로, 사람을 위한 설명은 Error로 구분합니다.

기본 NewReport는 Run이 반환한 완료 결과를 입력으로 받는 함수입니다.
외부에서 임의로 만든 Result를 모두 검증하는 파서 역할은 하지 않습니다.
새 상태를 추가하면 요약과 직렬화도 함께 갱신해야 합니다.

## 9. JSON 보고서 구현

소스: [batch/report.go](../../examples/31-concurrent-batch/batch/report.go)

```go
package batch

import (
	"encoding/json"
	"io"
)

type Summary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Canceled  int `json:"canceled"`
	Attempts  int `json:"attempts"`
}

type Item struct {
	Index    int    `json:"index"`
	Input    string `json:"input"`
	Output   string `json:"output,omitempty"`
	State    State  `json:"state"`
	Attempts int    `json:"attempts"`
	Error    string `json:"error,omitempty"`
}

type Report struct {
	Summary Summary `json:"summary"`
	Items   []Item  `json:"items"`
}

func NewReport(results []Result) Report {
	report := Report{Items: make([]Item, len(results))}
	report.Summary.Total = len(results)
	for i, result := range results {
		item := Item{Index: result.Job.Index, Input: result.Job.Input, Output: result.Output, State: result.State, Attempts: result.Attempts}
		if result.Err != nil {
			item.Error = result.Err.Error()
		}
		report.Items[i] = item
		report.Summary.Attempts += result.Attempts
		switch result.State {
		case Succeeded:
			report.Summary.Succeeded++
		case Failed:
			report.Summary.Failed++
		case Canceled:
			report.Summary.Canceled++
		}
	}
	return report
}

func WriteJSON(w io.Writer, results []Result) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(NewReport(results))
}
```

WriteJSON은 io.Writer를 받습니다.
테스트에서는 메모리 버퍼나 실패하는 Writer를 주입할 수 있습니다.
출력 실패는 오류로 반환하며 성공으로 무시하지 않습니다.

직렬화는 작업자 안에서 하지 않습니다.
한 결과를 여러 고루틴이 동시에 같은 Writer에 쓰는 구조를 피하고,
보고서 전체를 하나의 계약으로 테스트합니다.

## 10. 완성 CLI 실행하기

소스: [cmd/batch/main.go](../../examples/31-concurrent-batch/cmd/batch/main.go)

기본 실행은 내장된 세 항목을 처리합니다.
외부 네트워크에 접속하거나 파일을 만들지 않습니다.

```sh
go run ./examples/31-concurrent-batch/cmd/batch
```

예상 출력:

```json
{
  "summary": {
    "total": 3,
    "succeeded": 3,
    "failed": 0,
    "canceled": 0,
    "attempts": 4
  },
  "items": [
    {
      "index": 0,
      "input": "go",
      "output": "GO",
      "state": "succeeded",
      "attempts": 1
    },
    {
      "index": 1,
      "input": "retry:channels",
      "output": "CHANNELS",
      "state": "succeeded",
      "attempts": 2
    },
    {
      "index": 2,
      "input": "한글",
      "output": "한글",
      "state": "succeeded",
      "attempts": 1
    }
  ]
}
```

### 10.1 파일과 표준 입력

준비한 UTF-8 텍스트 파일은 한 줄에 한 항목입니다.

```sh
go run ./examples/31-concurrent-batch/cmd/batch -input examples/31-concurrent-batch/testdata/jobs.txt
```

Bash에서 표준 입력:

```sh
printf 'go\nretry:channels\n한글\n' | go run ./examples/31-concurrent-batch/cmd/batch -input -
```

Windows PowerShell에서는 UTF-8 파일 입력 명령을 그대로 사용할 수 있습니다.
문자 인코딩 설정에 따른 파이프 차이를 피하려면 위 파일 예제로 먼저 확인합니다.

빈 파일은 0개 항목입니다.
마지막 줄바꿈 하나가 불필요한 추가 항목을 만들지는 않습니다.
중간의 빈 줄은 빈 항목으로 읽혀 CLI 처리 단계에서 실패합니다.
CRLF 줄바꿈은 Scanner의 줄 읽기 규칙으로 처리합니다.

### 10.2 플래그

| 플래그 | 기본값 | 의미 |
| --- | --- | --- |
| `-workers` | 3 | 동시 작업자 수 |
| `-attempts` | 3 | 첫 실행을 포함한 최대 시도 |
| `-retry-delay` | 0s | 지수 증가의 기본 대기 |
| `-timeout` | 3s | 입력을 읽은 뒤 배치 처리 제한 시간 |
| `-input` | 생략 | 파일 경로, `-`는 표준 입력 |
| `-h` | | 도움말 후 정상 종료 |

```sh
go run ./examples/31-concurrent-batch/cmd/batch -workers 2 -attempts 3 -retry-delay 10ms -timeout 2s
```

위치 인자는 받지 않습니다.
잘못된 플래그·설정은 작업 시작 전에 거부합니다.

### 10.3 종료 코드

| 코드 | 의미 |
| --- | --- |
| 0 | 모든 항목 성공 또는 도움말 |
| 1 | 입력 읽기·항목 처리·출력 실패 |
| 2 | 플래그·설정 사용법 오류 |
| 130 | 호출 Context의 취소 또는 배치 전체 마감 |

JSON은 표준 출력, 오류 설명은 표준 오류로 보냅니다.
실패한 항목이 있어도 유효한 결과가 있으면 보고서를 출력하고 1로 종료합니다.

`go run`은 Go 도구가 실행 파일을 감싸므로
실제 프로그램의 종료 코드를 쉘에서 정확히 확인하려면 빌드한 파일을 실행합니다.

Bash:

```sh
mkdir -p bin
go build -o bin/batch ./examples/31-concurrent-batch/cmd/batch
./bin/batch -input examples/31-concurrent-batch/testdata/failures.txt
echo $?
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o bin/batch.exe ./examples/31-concurrent-batch/cmd/batch
./bin/batch.exe -input examples/31-concurrent-batch/testdata/failures.txt
$LASTEXITCODE
```

실패 예제의 프로그램 종료 코드는 1입니다.
출력에 succeeded와 failed가 함께 나타나는 것이 의도된 결과입니다.

### 10.4 입력 읽기와 처리 시간의 경계

`-timeout`은 **입력 전체를 읽은 뒤 시작하는 batch.Run의 시간 예산**입니다.
일반 io.Reader의 Read나 임의의 io.Writer의 Write를 강제로 중단하는 기능은 없습니다.
표준 입력이 끝나지 않으면 처리 단계에 아직 들어가지 않은 상태입니다.

파일 입력을 사용하거나 표준 입력에 EOF를 보내 입력을 끝냅니다.
입력 대기 중 Ctrl+C를 누르면 전체 Context를 취소하고 신호 가로채기를 해제합니다.
읽기가 계속 막혀 있다면 **다음 Ctrl+C가 기본 종료 동작**으로 돌아갈 수 있습니다.
첫 취소가 임의의 Reader를 즉시 깨운다고 설명하지 않습니다.

일반 터미널의 인터럽트 동작을 기준으로 하며,
다른 프로그램이 같은 프로세스에서 별도 신호 처리를 등록한 경우는 이 독립 CLI의 범위가 아닙니다.
[signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)의 stop 규칙을 참고합니다.

### 10.5 완료 전 os.Exit를 호출하지 않기

실제 정리는 realMain과 run 안에서 마치고 정수 코드를 반환합니다.
가장 바깥 main만 os.Exit를 호출합니다.
os.Exit는 defer를 실행하지 않으므로
취소·신호 복원·작업 합류를 그 전에 끝내야 합니다.

## 11. CLI 코드의 경계 읽기

다음은 완성 CLI의 전체 코드입니다.
입력·플래그·출력과 신호 처리가 공통 실행 엔진 밖에 있는 점을 확인합니다.

```go
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func readLines(input io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), batch.MaxItemBytes+2)
	lines := []string{}
	for scanner.Scan() {
		if len(lines) == batch.MaxItems {
			return nil, fmt.Errorf("입력은 %d줄까지 허용합니다", batch.MaxItems)
		}
		if len(scanner.Text()) > batch.MaxItemBytes {
			return nil, fmt.Errorf("한 줄은 %d바이트까지 허용합니다", batch.MaxItemBytes)
		}
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func run(parent context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("batch", flag.ContinueOnError)
	flags.SetOutput(stderr)
	workers := flags.Int("workers", 3, "동시에 처리할 작업자 수 (1~32)")
	attempts := flags.Int("attempts", 3, "항목별 최대 시도 횟수 (1~5)")
	retryDelay := flags.Duration("retry-delay", 0, "재시도 기본 대기 시간 (0~1s)")
	timeout := flags.Duration("timeout", 3*time.Second, "배치 처리 제한 시간 (0 초과, 최대 1m)")
	inputPath := flags.String("input", "", "입력 파일; - 는 표준 입력, 생략하면 내장 예제")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	cfg := batch.Config{Workers: *workers, Attempts: *attempts, RetryDelay: *retryDelay}
	if flags.NArg() != 0 || *timeout <= 0 || *timeout > time.Minute {
		fmt.Fprintln(stderr, "위치 인자는 받지 않으며 timeout은 0 초과 1m 이하여야 합니다")
		return 2
	}
	if err := batch.ValidateConfig(cfg); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	inputs := []string{"go", "retry:channels", "한글"}
	if *inputPath != "" {
		var err error
		if *inputPath == "-" {
			inputs, err = readLines(stdin)
		} else {
			var file *os.File
			file, err = os.Open(*inputPath)
			if err == nil {
				inputs, err = readLines(file)
				err = errors.Join(err, file.Close())
			}
		}
		if err != nil {
			fmt.Fprintln(stderr, "입력 오류:", err)
			return 1
		}
	}
	ctx, cancel := context.WithTimeout(parent, *timeout)
	defer cancel()
	// Each job index belongs to one worker; retries for that index are sequential.
	calls := make([]int, len(inputs))
	process := func(ctx context.Context, job batch.Job) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		calls[job.Index]++
		value := strings.TrimSpace(job.Input)
		if value == "" {
			return "", errors.New("empty item")
		}
		if strings.HasPrefix(value, "fail:") {
			return "", errors.New("permanent example failure")
		}
		if strings.HasPrefix(value, "retry:") {
			if calls[job.Index] == 1 {
				return "", fmt.Errorf("first attempt: %w", batch.ErrTemporary)
			}
			value = strings.TrimSpace(strings.TrimPrefix(value, "retry:"))
			if value == "" {
				return "", errors.New("empty retry item")
			}
		}
		return strings.ToUpper(value), nil
	}
	results, err := batch.Run(ctx, inputs, cfg, process)
	if results == nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if writeErr := batch.WriteJSON(stdout, results); writeErr != nil {
		fmt.Fprintln(stderr, "출력 오류:", writeErr)
		return 1
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return 130
		}
		return 1
	}
	return 0
}

// restoreOnCancel releases signal interception after the first interrupt.
// A second interrupt can then terminate even if a plain Reader is still blocked.
func restoreOnCancel(ctx context.Context, stop func()) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		stop()
		close(done)
	}()
	return done
}

func realMain() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	restored := restoreOnCancel(ctx, stop)
	defer func() {
		stop()
		<-restored
	}()
	return run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}
func main() { os.Exit(realMain()) }
```

### 11.1 Scanner의 크기 제한

Scanner 버퍼의 상한과 읽은 줄의 길이 검사를 함께 사용합니다.
줄 구분자의 여유를 포함하되, 실제 항목은 4096바이트까지입니다.
전체 줄 수도 10000개로 제한합니다.

readLines는 오류가 나면 부분 입력을 성공으로 반환하지 않습니다.
한 줄을 너무 크게 읽거나 입력 장치가 실패하면 콜백을 시작하지 않습니다.
파일을 열었다면 읽기 오류와 닫기 오류를 모두 확인합니다.

### 11.2 재시도 시뮬레이션의 상태

calls 슬라이스는 Job 인덱스별로 나뉩니다.
같은 Job의 재시도는 같은 작업자가 순차로 수행하므로
같은 calls[index]에 동시에 접근하는 다른 작업자가 없습니다.

전체 calls를 콜백 중간에 다른 고루틴이 읽지 않습니다.
서로 다른 Job이 같은 카운터를 공유하는 방식으로 바꾸면 별도 동기화가 필요합니다.

### 11.3 신호 복원 고루틴도 합류한다

restoreOnCancel은 첫 취소를 관찰하면 stop을 호출합니다.
이로써 다음 인터럽트를 계속 가로채고 있는 상태를 해제합니다.
realMain은 정상 반환에서도 stop을 호출하고 이 보조 고루틴의 종료를 기다립니다.

정리를 돕기 위해 만든 고루틴도 수명 소유자가 있어야 한다는 같은 원칙입니다.
보조 고루틴이라고 시작만 하고 잊어버리지 않습니다.

## 12. 무엇을 테스트하는가?

### 12.1 입력·설정·결과 계약

| 테스트 | 기대 |
| --- | --- |
| 작업자 0·33 | 실행 전 ErrConfig |
| 시도 0·6 | 실행 전 ErrConfig |
| 입력 10001개·4097바이트·잘못된 UTF-8 | 실행 전 ErrInput |
| 정상 빈 입력 | nil이 아닌 빈 결과 |
| 입력 순서 c,a,한글,a | 결과 위치도 같은 순서 |
| 잘못된 출력 인코딩·크기 | 해당 항목 Failed, ErrOutput |
| JSON Writer 실패 | 출력 오류 반환 |
| CSV 내 따옴표·쉼표·줄바꿈 | 표준 인코더로 왕복 보존 |

### 12.2 재시도 횟수와 대기

일시 오류 두 번 뒤 성공하는 콜백에 대해 Attempts=3을 확인합니다.
주입한 대기 함수에 10ms와 20ms가 순서대로 전달되는지 확인합니다.
실제로 잠들어 시간이 얼추 흘렀는지 비교하지 않습니다.

영구 오류는 첫 시도에서 멈춥니다.
일시 오류가 계속되어도 설정한 최대 횟수에서 끝납니다.
재시도 대기 중 취소되면 다음 호출이 없음을 검사합니다.

### 12.3 작업자 상한과 합류

테스트 콜백은 시작한 사실을 채널로 알리고 해제 신호를 기다립니다.
세 명이 시작했을 때 현재 실행 수가 3임을 확인합니다.
그 뒤 취소하고 Run이 반환했을 때 실행 중인 수가 0인지 확인합니다.

고루틴 개수를 런타임 전체에서 세는 대신
실제 Processor의 동시 실행 수를 측정합니다.
다른 테스트나 런타임의 고루틴 때문에 흔들리지 않습니다.

### 12.4 제한 시간은 실패 감지용

시작·종료 신호가 오지 않아 테스트가 영원히 멈추는 것을 막기 위해
실패용 제한 시간을 사용합니다.
성공 여부는 신호와 결과로 판단합니다.
Sleep을 길게 늘려 통과시키지 않습니다.

### 12.5 실행 명령

```sh
go test ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
go vet ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
go build ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
```

Linux와 C 컴파일러가 준비된 환경:

```sh
go test -race ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
```

검증한 범위를 넘어 Windows의 실제 인터럽트 동작이나
특정 외부 서비스의 취소 응답까지 확인했다고 말하지 않습니다.
라이브러리 테스트는 네트워크·실제 계정 없이 실행됩니다.

## 13. 흔한 실수

| 증상 | 원인 | 수정 방향 |
| --- | --- | --- |
| 고루틴 수가 입력만큼 증가 | 항목마다 시작 | 고정된 작업자가 jobs를 반복 수신 |
| 결과가 입력과 맞지 않음 | 완료 순서로 append | Job.Index로 결과 칸 지정 |
| 실패가 끝없이 반복 | 모든 오류·무제한 재시도 | 재시도 표지와 횟수 상한 |
| 취소해도 오래 대기 | Sleep 또는 취소 없는 하위 작업 | Timer/Done, Context 지원 호출 |
| 반환 뒤 결과가 바뀜 | 합류 전 반환 | 모든 작업자를 기다린 뒤 반환 |
| 취소하면 성공 결과도 사라짐 | 부분 결과 정책 누락 | 항목별 상태와 시도 횟수 유지 |
| 두 번 실행해 중복 효과 | 재시도의 업무 의미 미검토 | 중복 처리 계약과 확인 절차 |
| stdin에서 계속 기다림 | 입력 EOF가 없고 Reader 취소 미지원 | 입력 종료 또는 인터럽트 처리 확인 |
| 요약의 총계와 시도 횟수 혼동 | 항목과 실행 횟수 구분 없음 | Total과 Attempts 별도 집계 |

## 14. 연습문제와 완료 기준

[연습문제](../../exercises/31-concurrent-batch/README.md)에서
상태 추적, CSV 보고서, 항목별 제한 시간,
재시도 후보 추출, 수명 검토를 수행합니다.
작성한 뒤 [해설](../../solutions/31-concurrent-batch/README.md)과 비교합니다.

완성 기준은 다음과 같습니다.

- 정상·실패·취소에서 입력마다 결과의 위치가 유지됩니다.
- 설정한 작업자 수보다 많은 Processor가 동시에 실행되지 않습니다.
- 재시도 대상과 최대 시도 횟수가 코드와 문서에서 일치합니다.
- 취소 요청만 하고 반환하지 않고 모든 내부 작업을 기다립니다.
- 검증 실패와 일부 항목 실패를 구분합니다.
- 표준 출력의 보고서와 표준 오류의 설명을 구분합니다.
- 실제 실행 결과와 본문의 출력·종료 코드가 맞습니다.

## 15. 이 실습 이후의 설계 경계

이 배치는 한 프로세스의 메모리에서 동작합니다.
프로세스가 종료된 뒤 작업을 이어 가는 내구성 있는 큐는 아닙니다.
외부 시스템에 정확히 한 번만 반영하는 분산 실행 엔진도 아닙니다.

더 큰 시스템에서는 작업 식별자, 저장된 진행 상태,
중복 실행 처리, 외부 속도 제한, 재시도 분산, 관측 자료가 필요할 수 있습니다.
추가 기능을 넣기 전에 현재의 입력·결과·취소 계약이 어떻게 바뀌는지 정합니다.

지금 완성한 엔진의 장점은 그 경계가 명시적이라는 점입니다.
순차 처리 함수를 받아 제한된 실행 흐름으로 조합하고,
모든 성공과 실패를 검사 가능한 결과로 돌려줍니다.

## 요약과 자기 점검

1. 입력 10000개와 Workers=4이면 어느 수가 콜백 동시 실행 수를 결정하나요?
2. 결과 배열에 Mutex가 없어도 안전한 이유는 무엇인가요?
3. jobs를 닫는 주체가 하나여야 하는 이유는 무엇인가요?
4. 일시 오류와 영구 오류의 표지를 누가 결정하나요?
5. 최대 시도 3회는 재시도를 세 번 더 한다는 뜻인가요?
6. canceled 항목에서 Attempts=0과 Attempts=1의 차이는 무엇인가요?
7. Context 취소가 외부 시스템의 부작용을 되돌리나요?
8. -timeout이 표준 입력의 모든 Read를 강제로 중단하나요?
9. race 검사와 작업자 상한 테스트는 각각 무엇을 확인하나요?
10. 프로세스 재시작 후 이어서 처리하려면 어떤 정보가 추가로 필요할까요?

## 공식 참고 자료

공식 API와 취소·동기화 규칙을 2026-11-09에 확인했습니다.

- [Context](https://pkg.go.dev/context)
- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [Go 메모리 모델](https://go.dev/ref/mem)
- [파이프라인과 취소](https://go.dev/blog/pipelines)
- [time.Timer](https://pkg.go.dev/time#Timer)
- [errors.Is와 As](https://pkg.go.dev/errors)
- [bufio.Scanner](https://pkg.go.dev/bufio#Scanner)
- [encoding/json](https://pkg.go.dev/encoding/json)
- [signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)
- [Data Race Detector](https://go.dev/doc/articles/race_detector)

[이전: 30장 할 일 HTTP API](../30-todo-api/README.md) ·
[목차](../README.md) ·
[프로젝트 홈](../../README.md)
