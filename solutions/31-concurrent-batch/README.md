# 31장 해설 — 동시 배치 처리

[연습문제](../../exercises/31-concurrent-batch/README.md)의 참고 구현입니다.
모든 명령은 저장소 루트에서 실행합니다.
공통 실행 엔진은 examples의 batch 패키지를 그대로 사용합니다.

이번 해설의 확장은 실행 엔진의 계약을 유지합니다.
출력 형식, 콜백의 항목 수명, 재시도 후보 선택을 각각 별도 함수로 만듭니다.

## 문제 1. 상태와 횟수

첫 시나리오의 결과는 다음과 같습니다.

| 항목 | State | Attempts |
| --- | --- | --- |
| A | succeeded | 1 |
| B | succeeded | 2 |
| C | failed | 1 |

| Summary 필드 | 값 |
| --- | --- |
| Total | 3 |
| Succeeded | 2 |
| Failed | 1 |
| Canceled | 0 |
| Attempts | 4 |

반환 오류는 Count=1인 BatchError입니다.
A와 B의 결과는 보존됩니다.
오류 체인에서 errors.Is로 C의 sentinel을 찾을 수 있습니다.

최대 시도 횟수는 첫 실행을 포함합니다.
재시도가 있어도 입력 항목 수 Total은 증가하지 않습니다.
같은 입력 문자열이 두 개 있어도 Job.Index가 다르므로 별도 항목입니다.

### 취소 시나리오

| 항목 | State | Attempts | 결과 |
| --- | --- | --- | --- |
| A | succeeded | 1 | 성공 출력 보존 |
| B | canceled | 1 | 취소 오류 |
| C | canceled | 0 | 미시작 취소 |

전체 반환 오류는 context.Canceled입니다.
이미 성공한 A를 되돌리거나 없애지 않습니다.

작업자가 여러 명이면 B의 취소 전에 C가 시작되거나 끝날 수 있습니다.
따라서 특정 항목의 시도 횟수를 이 표처럼 고정할 수 없습니다.
입력 위치에 맞는 결과, 동시 실행 상한, 반환 전 합류가 유지되는지를 검사합니다.

## 참고 구현

| 문제 | 소스 | 핵심 계약 |
| --- | --- | --- |
| 2 | [csv-report](csv-report/main.go) | 표준 인코더·Flush 오류 |
| 3 | [job-timeout](job-timeout/main.go) | 항목 자식 Context의 정리 |
| 4 | [retry-plan](retry-plan/main.go) | 상태와 일시 오류를 함께 선택 |

## 문제 2. CSV 보고서의 입출력 경계


CSV 출력은 처리 엔진과 분리된 함수입니다.
완료된 결과만 읽으므로 동시 상태를 잠글 필요가 없습니다.
호출자가 소유한 Writer를 닫지 않고, 쓰기 결과를 오류로 돌려줍니다.

소스: [csv-report/main.go](csv-report/main.go)

```go
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func writeCSV(out io.Writer, results []batch.Result) error {
	writer := csv.NewWriter(out)
	if err := writer.Write([]string{"index", "input", "state", "attempts", "output", "error"}); err != nil {
		return err
	}
	for _, result := range results {
		message := ""
		if result.Err != nil {
			message = result.Err.Error()
		}
		record := []string{strconv.Itoa(result.Job.Index), result.Job.Input, string(result.State), strconv.Itoa(result.Attempts), result.Output, message}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
func run() error {
	results, err := batch.Run(context.Background(), []string{"go", "한글"}, batch.Config{Workers: 2, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) { return strings.ToUpper(job.Input), ctx.Err() })
	if err != nil {
		return err
	}
	return writeCSV(os.Stdout, results)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./solutions/31-concurrent-batch/csv-report
```

예상 출력:

```text
index,input,state,attempts,output,error
0,go,succeeded,1,GO,
1,한글,succeeded,1,한글,
```


### 직접 문자열을 이어 붙이지 않는 이유

입력 문자열에 쉼표·따옴표·줄바꿈이 있을 수 있습니다.
단순 Join으로 만들면 필드와 행 경계가 바뀝니다.
csv.Writer는 CSV의 인용 규칙을 적용하고,
테스트는 csv.Reader로 다시 읽어 원래 필드가 보존되는지 확인합니다.

### Flush 뒤의 Error

Writer는 내부 버퍼를 사용할 수 있으므로
각 Write 호출에서 아직 실제 출력 장치 오류를 보지 못할 수 있습니다.
마지막 Flush 후 Error를 확인해야 늦은 실패를 놓치지 않습니다.

이 함수는 out을 닫지 않습니다.
파일을 생성한 호출자가 모든 쓰기와 닫기 오류를 책임집니다.
원자적인 파일 교체가 필요하면 임시 파일·동기화·이름 변경 같은 별도 계약이 필요하지만,
현재 함수는 주어진 Writer에 CSV를 쓰는 범위만 담당합니다.

### 테스트

[CSV 경계 테스트](csv-report/behavior_test.go)는 특수 문자와 Writer 오류를 확인합니다.
빈 결과의 헤더도 유지합니다.
오류 문자열 자체를 새 CSV 규칙으로 손으로 이스케이프하지 않습니다.

## 문제 3. 전체 수명 아래에 항목 수명 만들기


각 호출은 전달받은 parent에서 WithTimeout으로 자식을 만듭니다.
함수가 끝날 때 cancel을 호출해 관련 자원을 정리합니다.
콜백을 새 고루틴에 옮겨 호출자만 먼저 반환시키지 않습니다.

소스: [job-timeout/main.go](job-timeout/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func withJobTimeout(timeout time.Duration, process batch.Processor) (batch.Processor, error) {
	if timeout < 0 || timeout > time.Minute || process == nil {
		return nil, errors.New("제한 시간은 0~1m이며 작업 함수가 필요합니다")
	}
	return func(parent context.Context, job batch.Job) (string, error) {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()
		if err := ctx.Err(); err != nil {
			return "", err
		}
		value, err := process(ctx, job)
		if err != nil {
			return "", err
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return value, nil
	}, nil
}
func main() {
	process, err := withJobTimeout(0, func(ctx context.Context, job batch.Job) (string, error) { return job.Input, ctx.Err() })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	results, err := batch.Run(context.Background(), []string{"go"}, batch.Config{Workers: 1, Attempts: 3}, process)
	fmt.Println("항목 상태:", results[0].State)
	fmt.Println("시도 횟수:", results[0].Attempts)
	fmt.Println("항목 마감 오류:", errors.Is(err, context.DeadlineExceeded))
}
```

```sh
go run ./solutions/31-concurrent-batch/job-timeout
```

예상 출력:

```text
항목 상태: failed
시도 횟수: 1
항목 마감 오류: true
```


### 0초 제한의 의미

0초 제한은 이미 끝난 마감을 만들어 테스트를 결정적으로 수행합니다.
실제 process를 호출하기 전 ctx.Err를 확인하므로 호출 횟수는 0입니다.
배치 관점에서는 wrapper를 한 번 호출했으므로 Result.Attempts는 1입니다.

여기서 **배치가 호출한 Processor 횟수**와
**wrapper 내부의 원래 함수 호출 횟수**를 구분합니다.
두 함수 경계가 다르므로 서로 다른 수치가 나오는 것이 맞습니다.

### 왜 failed인가?

Run의 전체 Context는 Background로 계속 유효합니다.
항목 내부의 자식만 DeadlineExceeded입니다.
엔진은 해당 콜백 오류를 영구 오류로 분류하고 failed로 기록합니다.

ErrTemporary를 감싼 오류가 아니므로 최대 시도가 3이어도 반복하지 않습니다.
전체 BatchError의 원인에서 DeadlineExceeded를 찾을 수 있습니다.
이것이 전체 Context가 취소되었다는 의미는 아닙니다.

### 강제 중단을 제공하지 않는다

실제 process가 Context를 확인하지 않고 멈추면 wrapper도 기다립니다.
마감이 있는 Context를 전달하는 것과 임의의 실행을 강제로 죽이는 것은 다릅니다.
실제 입출력에는 Context·기한을 지원하는 API를 사용해야 합니다.

콜백이 성공해도 반환 직전에 자식 마감이 확인되면 값 대신 오류를 돌려줍니다.
이미 발생한 외부 부작용을 되돌리는 정책은 아닙니다.

## 문제 4. 후보 선택과 실제 재실행 분리


후보 함수는 Failed이면서 일시 오류인 항목만 골라 냅니다.
errors.Is를 사용하므로 래핑된 오류도 같은 의미로 처리합니다.
오류 문자열을 비교하지 않습니다.

소스: [retry-plan/main.go](retry-plan/main.go)

```go
package main

import (
	"errors"
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func retryInputs(results []batch.Result) []string {
	inputs := []string{}
	for _, result := range results {
		if result.State == batch.Failed && errors.Is(result.Err, batch.ErrTemporary) {
			inputs = append(inputs, result.Job.Input)
		}
	}
	return inputs
}
func main() {
	results := []batch.Result{
		{Job: batch.Job{Index: 0, Input: "ok"}, State: batch.Succeeded},
		{Job: batch.Job{Index: 1, Input: "busy"}, State: batch.Failed, Err: fmt.Errorf("server busy: %w", batch.ErrTemporary)},
		{Job: batch.Job{Index: 2, Input: "invalid"}, State: batch.Failed, Err: errors.New("invalid input")},
		{Job: batch.Job{Index: 3, Input: "canceled"}, State: batch.Canceled},
	}
	fmt.Println("검토할 재시도 입력:", retryInputs(results))
}
```

```sh
go run ./solutions/31-concurrent-batch/retry-plan
```

예상 출력:

```text
검토할 재시도 입력: [busy]
```


### 순서와 중복

결과 슬라이스를 앞에서부터 읽고 조건에 맞는 Input을 추가합니다.
따라서 원래 입력 순서가 유지됩니다.
같은 문자열의 실패가 두 개 있으면 두 항목 모두 남습니다.
자동으로 중복 제거하면 별도의 업무 요청을 합쳐 버릴 수 있습니다.

### 취소를 자동 후보에서 제외한 이유

취소된 항목은 호출자가 더 이상 진행을 원하지 않았을 수 있습니다.
시도 횟수가 0인 항목과 이미 외부 효과가 발생했을 가능성이 있는 항목도 섞입니다.
이 문제에서는 사용자의 취소 의도와 외부 효과를 검토한 뒤 별도로 판단하도록 제외합니다.

일시 오류도 무조건 안전한 재실행 보장은 아닙니다.
현재 함수는 후보 목록만 만들며 실제로 다시 실행하지 않습니다.
실제 재실행 전에는 처리의 중복 허용성과 업무 식별자를 확인해야 합니다.

### 영속적인 재시도 목록으로 확장할 때

문자열 Input만으로 원래 Job의 모든 이력을 보존할 수 있는 것은 아닙니다.
프로세스 재시작 뒤 이어서 처리하려면 안정적인 식별자·이전 시도·결과 기록이 필요합니다.
이 실습의 후보 함수는 메모리 결과를 읽는 작은 변환으로 범위를 한정합니다.

## 문제 5. 실행 엔진의 수명 표

| 대상 | 쓰기·닫기 소유자 | 관찰 시점 |
| --- | --- | --- |
| jobs 전송 | Run의 분배 반복 | 작업자가 range로 수신 |
| jobs 닫기 | 분배를 마친 Run | 모든 작업자의 range 종료 |
| results[index] | 그 Job을 받은 작업자 | group.Wait 뒤의 Run과 호출자 |
| 미시작 상태 | group.Wait 뒤의 Run | 반환된 결과를 읽는 호출자 |
| 재시도 Timer | execute가 부른 대기 함수 | Timer 또는 ctx.Done 선택 |
| 호출 Context | 호출자의 수명 소유자 | 분배·실행·대기·최종 확인 |

### 잘못된 변경안

1. 작업자가 jobs를 닫으면 분배자가 아직 보내는 중일 수 있어 패닉이 발생합니다.
2. Wait를 생략하면 반환 뒤에도 결과를 쓰는 작업자가 남습니다.
3. 공유 append는 결과 위치를 잃고 슬라이스 변경에 레이스를 만듭니다.
4. 모든 오류 재시도는 영구 오류를 반복하고 부작용을 늘릴 수 있습니다.
5. Sleep은 대기 중 Context 취소를 관찰하지 않습니다.
6. 항목마다 고루틴을 만들면 설정한 동시 실행 상한이 실제로 적용되지 않습니다.
7. wrapper는 Context를 전달할 뿐 강제 중단을 제공하지 않습니다.

### 어느 테스트가 무엇을 확인하는가?

| 잘못된 변경 | 필요한 검사 |
| --- | --- |
| 합류 생략 | 반환 시 active=0, 종료 신호 |
| 무제한 작업자 | 시작 게이트로 동시에 실행 중인 수 관찰 |
| 공유 결과 변경 | race detector와 입력 순서 결과 비교 |
| 영구 오류 재시도 | 정확한 콜백 호출 횟수 |
| 일시 오류 무한 반복 | 최대 Attempts와 오류 종류 |
| 잘못된 출력 | 크기·UTF-8 경계 |
| 취소를 무시한 대기 | 취소 중 대기 종료와 다음 호출 없음 |

레이스 검사만으로 재시도 횟수나 업무 결과의 오류를 모두 찾을 수는 없습니다.
각 계약에 맞는 관찰값을 함께 확인합니다.

## 공통 엔진의 상한·합류 테스트 읽기

아래는 [batch_test.go](../../examples/31-concurrent-batch/batch/batch_test.go)의 발췌입니다.
시작한 콜백 수를 채널로 확인하므로 Sleep으로 진행을 추측하지 않습니다.

```go
func TestWorkerBoundAndJoin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const workers = 3
	started := make(chan struct{}, workers)
	release := make(chan struct{})
	defer close(release)
	var active, peak atomic.Int32
	finished := make(chan []Result, 1)
	errCh := make(chan error, 1)
	process := func(ctx context.Context, job Job) (string, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for {
			old := peak.Load()
			if old >= n || peak.CompareAndSwap(old, n) {
				break
			}
		}
		if job.Index < workers {
			started <- struct{}{}
		}
		select {
		case <-release:
			return job.Input, nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	go func() {
		results, err := Run(ctx, []string{"a", "b", "c", "d", "e", "f"}, Config{Workers: workers, Attempts: 1}, process)
		finished <- results
		errCh <- err
	}()
	for range workers {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("workers not started")
		}
	}
	if active.Load() != workers {
		t.Fatal(active.Load())
	}
	cancel()
	select {
	case results := <-finished:
		if len(results) != 6 {
			t.Fatal(results)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("workers not joined")
	}
	if !errors.Is(<-errCh, context.Canceled) || active.Load() != 0 || peak.Load() != workers {
		t.Fatal(active.Load(), peak.Load())
	}
}
```

세 작업자가 모두 시작했음을 확인한 뒤 취소합니다.
Run이 반환한 시점에 active가 0이어야 합니다.
peak는 실행 중 관찰한 최대 콜백 수이며 설정한 작업자 수와 같습니다.

이 테스트는 전체 프로세스의 goroutine 수를 세지 않습니다.
다른 테스트와 런타임의 고루틴이 섞여도 관심 있는 처리 함수의 상한을 직접 측정합니다.

5초 타이머는 시작·합류가 되지 않을 때 실패를 보고하기 위한 것입니다.
그 시간이 지나면 성공이라고 가정하지 않습니다.

## CLI 검증 결과의 해석

기본 입력은 go, retry:channels, 한글입니다.
성공 항목은 3개이고 총 시도 횟수는 4입니다.
가운데 항목의 첫 실패는 엔진 안에서 재시도되어 최종 결과는 succeeded입니다.

failures.txt에는 영구 실패 항목이 하나 있습니다.
보고서는 성공과 실패를 함께 담고 실제 프로그램은 코드 1로 종료합니다.
go run의 래퍼 코드와 실제 실행 파일의 코드를 구분해야 합니다.

설정 오류는 작업을 시작하기 전에 코드 2입니다.
-h는 도움말을 출력하고 코드 0입니다.
Writer 실패는 완료된 계산을 stdout으로 전달하지 못한 것이므로 코드 1입니다.

### 입력 대기의 경계

임의의 Reader.Read는 Context만으로 자동 중단되지 않습니다.
CLI의 -timeout은 입력을 읽은 뒤 배치 처리에 적용됩니다.
입력 대기에서는 EOF를 보내거나 인터럽트 정책을 사용합니다.

첫 Ctrl+C가 Context를 취소하면 신호 가로채기를 해제합니다.
읽기가 계속 막힌 경우 이후 Ctrl+C로 기본 종료 동작을 사용할 수 있습니다.
보조 고루틴도 정상 반환 때 합류하므로 정리용 작업이 남지 않습니다.

## 검증 명령

```sh
go fmt ./solutions/31-concurrent-batch/...
go vet ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
go test ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
go build ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
```

Linux와 C 컴파일러가 준비된 환경:

```sh
go test -race ./examples/31-concurrent-batch/... ./solutions/31-concurrent-batch/...
```

## 최종 평가 기준

| 항목 | 통과 기준 |
| --- | --- |
| 제한 | 동시 콜백 수·입력·출력·시도 횟수의 상한 |
| 순서 | 원래 입력 위치에 대응하는 결과 |
| 오류 | 검증·항목 실패·전체 취소를 구분 |
| 재시도 | 일시 오류만 유한하게 반복 |
| 수명 | 취소 뒤에도 모든 작업의 종료를 기다림 |
| 보고 | 성공·실패·취소·시도 횟수 보존 |
| 입출력 | 읽기·쓰기·닫기 오류를 숨기지 않음 |
| 테스트 | 시간 추측 없이 시작·완료·상한을 확인 |

## 참고 자료

- [Context](https://pkg.go.dev/context)
- [WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [time.Timer](https://pkg.go.dev/time#Timer)
- [encoding/csv](https://pkg.go.dev/encoding/csv)
- [errors.Is](https://pkg.go.dev/errors#Is)
- [파이프라인과 취소](https://go.dev/blog/pipelines)
- [signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)

[문제로 돌아가기](../../exercises/31-concurrent-batch/README.md) ·
[본문으로 돌아가기](../../book/31-concurrent-batch/README.md)
