# 18장 해설 — 고루틴과 완료 대기

[연습문제](../../exercises/18-goroutines/README.md)의 참고 구현입니다.
모든 명령은 저장소 루트에서 실행합니다.
실행 코드와 테스트는 외부 서비스 없이 종료됩니다.

이 해설의 기준은 함수의 모양보다 **수명·순서·오류의 계약**입니다.
작업을 만들었다면 기다리고, 오류가 있어도 시작한 작업의 수명을 남기지 않습니다.

| 문제 | 소스 | 확인하는 계약 |
| --- | --- | --- |
| 2 | [parallel-labels](parallel-labels/main.go) | 입력 순서와 인덱스 소유권 |
| 4 | [parallel-sum](parallel-sum/main.go) | 고정된 작업자 수와 지역 합계 |
| 5 | [parallel-transform](parallel-transform/main.go) | 전체 합류와 오류 보존 |
| 6 | [joined-task](joined-task/main.go) | 취소 요청 후 실제 종료 |

## 문제 1. 완료를 기다리는 것은 성공을 보장하지 않는다

1. main은 다른 고루틴의 종료를 자동으로 기다리지 않으므로 출력이 보장되지 않습니다.
2. 새 고루틴이 Add를 하기 전에 호출자가 Wait를 만나면 카운터 0을 보고 통과할 수 있습니다.
3. 입력 오류나 건너뛰기로 중간 반환하면 마지막 Done까지 도달하지 않습니다.
4. Wait는 등록된 작업이 끝났음을 의미합니다. 오류 결과는 별도로 확인합니다.
5. cancel은 취소 의사를 전달합니다. 종료는 함수 반환이나 완료 신호로 확인합니다.
6. 다른 원소를 각각 한 작업이 쓰고 호출자가 Wait 뒤에서 읽기 때문입니다.
7. append는 같은 슬라이스 정보와 배열의 일부를 공유해 변경합니다.
8. 고루틴은 런타임의 실행 단위이며 CPU 코어와 일대일로 대응하지 않습니다.

| 사건 | 개념적인 카운터 | 결과 읽기 |
| --- | --- | --- |
| 생성 | 0 | 아직 작업 없음 |
| Add(2) | 2 | 진행 중 작업의 결과를 읽지 않음 |
| 첫 Done | 1 | 남은 작업과의 관계가 필요함 |
| 둘째 Done | 0 | Wait 완료 조건 성립 |
| Wait 반환 | 0 | 완료 전에 쓴 결과를 읽음 |

실제 필드에 접근하거나 카운터 수치를 조회하는 코드는 만들지 않습니다.
공개 메서드로 수명을 표현합니다.

## 문제 2. 이름표를 인덱스에 맞춰 기록하기

입력을 모두 검증한 뒤 결과 슬라이스를 만듭니다.
이 시점에서 실패하면 아직 고루틴을 하나도 시작하지 않았습니다.


각 작업은 자기 결과 칸만 씁니다.
TrimSpace와 ToUpper는 입력 문자열을 반환값으로 변환합니다.
중복 이름은 서로 다른 입력 항목이므로 결과도 두 개 유지합니다.

소스: [parallel-labels/main.go](parallel-labels/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

func labels(names []string) ([]string, error) {
	if len(names) > 16 {
		return nil, errors.New("이름은 16개까지 허용합니다")
	}
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			return nil, errors.New("빈 이름은 허용하지 않습니다")
		}
	}
	results := make([]string, len(names))
	var group sync.WaitGroup
	for index, name := range names {
		group.Go(func() {
			results[index] = strings.ToUpper(strings.TrimSpace(name))
		})
	}
	group.Wait()
	return results, nil
}

func main() {
	results, err := labels([]string{" go ", "test", "한글"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(results)
}
```

```sh
go run ./solutions/18-goroutines/parallel-labels
```

예상 출력:

```text
[GO TEST 한글]
```

### 경계에서 확인할 것

| 입력 | 결과 |
| --- | --- |
| nil | 빈 결과 |
| go, test, 한글 | GO, TEST, 한글 |
| same, same | SAME 두 개 |
| 16개 정상 이름 | 성공 |
| 17개 | 오류와 nil 결과 |
| 공백뿐인 이름 포함 | 오류와 nil 결과 |

결과가 빈 경우에도 성공과 실패를 오류값으로 구분합니다.
허용된 빈 입력을 잘못된 입력으로 보지 않습니다.

### 반복 변수의 의미

현재 교재의 Go 버전에서는 range의 `:=` 선언 변수가 각 반복에서 구분됩니다.
코드의 index와 name은 각 반복의 값입니다.
기존 변수를 `=`로 재사용하거나 외부 변수를 직접 바꾸는 형태는 별도 문제입니다.

입력 슬라이스를 호출 중 다른 곳에서 바꾸지 않는 계약도 유지해야 합니다.
결과를 안전하게 나누었다고 입력의 외부 변경까지 막아 주지는 않습니다.

## 문제 3. 잘못된 구조의 수정

### A. 시작 전에 등록한다

```go
var group sync.WaitGroup
group.Add(1)
go func() {
    defer group.Done()
    // work
}()
group.Wait()
```

또는 현재 버전에서는 다음처럼 등록과 실행을 묶을 수 있습니다.

```go
var group sync.WaitGroup
group.Go(func() {
    // work
})
group.Wait()
```

두 번째 코드를 `go group.Go(...)`로 다시 감싸지 않습니다.
그렇게 하면 등록을 미루는 원래 문제를 다시 만들 수 있습니다.

### B. 쓰는 위치를 나눈다

```go
results := make([]int, len(values))
for index, value := range values {
    group.Go(func() {
        results[index] = value * value
    })
}
group.Wait()
```

values의 범위와 곱셈의 범위는 호출 전에 검증해야 합니다.
이 발췌는 소유권 수정만 보여 주며 모든 입력 검증을 대신하지 않습니다.

### C. 대기 객체를 복사하지 않는다

등록과 대기를 한 함수에서 모두 소유하는 구조가 가장 분명합니다.
외부 helper가 꼭 필요하면 `*sync.WaitGroup`을 전달할 수 있습니다.
그러나 어느 쪽이 언제 등록을 끝내는지 계약이 더 복잡해집니다.

```go
func launch(group *sync.WaitGroup, work func()) {
    group.Add(1)
    go func() {
        defer group.Done()
        work()
    }()
}
```

이 발췌에서도 work의 nil·패닉·공유 상태·취소 책임은 별도입니다.
WaitGroup 포인터만 바꾼다고 작업의 모든 문제가 해결되지는 않습니다.

## 문제 4. 부분 합계를 지역 변수로 계산하기

입력 길이보다 작업자 수가 많으면 작업자를 줄입니다.
각 작업자는 인덱스를 일정 간격으로 건너뛰며 자기 합계를 계산합니다.
반복 중 공유 합계에 접근하지 않고 마지막에 자기 결과 칸 한 번만 씁니다.

소스: [parallel-sum/main.go](parallel-sum/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

func sumParts(values []int, workers int) (int, error) {
	if workers < 1 || workers > 8 || len(values) > 1000 {
		return 0, errors.New("작업자 1~8명과 값 1000개 이하가 필요합니다")
	}
	for _, value := range values {
		if value < 0 || value > 1000 {
			return 0, errors.New("값은 0~1000이어야 합니다")
		}
	}
	if len(values) == 0 {
		return 0, nil
	}
	if workers > len(values) {
		workers = len(values)
	}
	partials := make([]int, workers)
	var group sync.WaitGroup
	for worker := range workers {
		group.Go(func() {
			total := 0
			for index := worker; index < len(values); index += workers {
				total += values[index]
			}
			partials[worker] = total
		})
	}
	group.Wait()
	total := 0
	for _, partial := range partials {
		total += partial
	}
	return total, nil
}

func main() {
	total, err := sumParts([]int{10, 20, 30, 40, 50}, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("합계:", total)
}
```

```sh
go run ./solutions/18-goroutines/parallel-sum
```

예상 출력:

```text
합계: 150
```

### 나눠 맡은 인덱스

값이 다섯 개이고 작업자가 두 명이면 다음과 같습니다.

| 작업자 | 값 | 지역 합계 |
| --- | --- | --- |
| 0 | 10, 30, 50 | 90 |
| 1 | 20, 40 | 60 |
| 호출자 | 90, 60 | 150 |

루프의 total은 각 고루틴이 독립적으로 갖는 지역 변수입니다.
partials의 서로 다른 원소를 쓰므로 잠금 없이 결합할 수 있습니다.
최종 total은 Wait 뒤의 호출자만 수정합니다.

### 범위를 먼저 정한 이유

1000개에 각 1000 이하이므로 합계는 최대 1,000,000입니다.
지원하는 32비트 int 범위에서도 안전하게 계산할 수 있습니다.
실제 업무에 이 범위를 그대로 강요할 이유는 없지만,
범위를 늘리면 오버플로 정책도 함께 정해야 합니다.

작업자 수를 늘린다고 더 빨라진다는 주장은 하지 않습니다.
이 정도 작은 합계는 순차 계산이 더 효율적일 수 있습니다.
학습 목표는 제한된 수의 실행 단위와 결과 결합입니다.

## 문제 5. 성공값과 오류의 소유권을 함께 나누기

작업마다 인덱스별 결과와 오류 칸을 둡니다.
errors.Join은 Wait 뒤에 호출하므로 아직 쓰는 중인 오류 슬라이스를 읽지 않습니다.
콜백 전에 취소되었으면 작업을 시작하지 않고,
콜백이 성공해도 그 사이 취소되었다면 실패 정책을 적용합니다.

소스: [parallel-transform/main.go](parallel-transform/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

func transformAll(ctx context.Context, inputs []string, transform func(context.Context, string) (string, error)) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(inputs) > 16 || transform == nil {
		return nil, errors.New("변환 함수와 입력 16개 이하가 필요합니다")
	}
	results := make([]string, len(inputs))
	failures := make([]error, len(inputs))
	var group sync.WaitGroup
	for index, input := range inputs {
		group.Go(func() {
			if err := ctx.Err(); err != nil {
				failures[index] = err
				return
			}
			value, err := transform(ctx, input)
			if err == nil {
				err = ctx.Err()
			}
			if err != nil {
				failures[index] = fmt.Errorf("항목 %d: %w", index, err)
				return
			}
			results[index] = value
		})
	}
	group.Wait()
	if err := errors.Join(failures...); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func run() error {
	results, err := transformAll(context.Background(), []string{"go", "test"}, func(ctx context.Context, input string) (string, error) {
		return strings.ToUpper(input), ctx.Err()
	})
	if err != nil {
		return err
	}
	fmt.Println("정상 결과:", results)
	results, err = transformAll(context.Background(), []string{"go", " "}, func(ctx context.Context, input string) (string, error) {
		if strings.TrimSpace(input) == "" {
			return "", errors.New("빈 입력")
		}
		return strings.ToUpper(input), ctx.Err()
	})
	fmt.Println("실패 결과 없음:", results == nil)
	fmt.Println("오류 확인:", err != nil)
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
go run ./solutions/18-goroutines/parallel-transform
```

예상 출력:

```text
정상 결과: [GO TEST]
실패 결과 없음: true
오류 확인: true
```

### 실패와 취소가 겹치는 경우

여러 오류를 errors.Join으로 결합하므로
작업 오류와 취소 오류가 함께 들어갈 수 있습니다.
반환 오류를 하나의 문자열로만 분류하지 않고 errors.Is로 원인을 찾습니다.

콜백 자체가 실패하면 원래 오류를 보존합니다.
콜백이 성공했을 때는 바로 다음의 ctx.Err 확인으로 취소를 반영합니다.
모든 고루틴을 기다린 뒤 최종 Context 상태도 다시 확인합니다.

이 정책은 “모든 결과를 성공으로 제공할 조건”을 엄격하게 정한 것입니다.
이미 외부 시스템에 반영된 부작용을 되돌린다는 뜻은 아닙니다.
취소와 트랜잭션 롤백은 서로 다른 문제입니다.

### 오류 보존 테스트의 발췌

다음 테스트는 두 실패를 모두 찾고 모든 항목이 실행되었는지 확인합니다.
테스트용 seen도 다른 인덱스에 쓰며, 함수 반환 뒤 읽습니다.

```go
func TestTransformJoinsAllErrors(t *testing.T) {
	first := errors.New("first")
	second := errors.New("second")
	seen := make([]bool, 3)
	got, err := transformAll(context.Background(), []string{"a", "b", "c"}, func(ctx context.Context, input string) (string, error) {
		index := int(input[0] - 'a')
		seen[index] = true
		switch input {
		case "a":
			return "", first
		case "c":
			return "", second
		}
		return input, nil
	})
	if got != nil || !errors.Is(err, first) || !errors.Is(err, second) || !reflect.DeepEqual(seen, []bool{true, true, true}) {
		t.Fatal(got, err, seen)
	}
}
```

## 문제 6. 함수 반환이 합류를 의미하도록 만들기

이 함수의 약속은 내부 작업이 끝난 뒤 반환하는 것입니다.
고루틴을 시작해 작업을 옮기는 것만으로 취소 기능이 새로 생기지는 않습니다.
task 자체도 Context를 관찰해야 종료가 가능합니다.

소스: [joined-task/main.go](joined-task/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

func runJoined(ctx context.Context, task func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if task == nil {
		return errors.New("작업 함수가 필요합니다")
	}
	var result error
	var group sync.WaitGroup
	group.Go(func() {
		result = task(ctx)
		if result == nil {
			result = ctx.Err()
		}
	})
	group.Wait()
	return result
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runJoined(ctx, func(ctx context.Context) error {
		cancel()
		<-ctx.Done()
		return ctx.Err()
	})
	fmt.Println("종료까지 기다림:", errors.Is(err, context.Canceled))
}
```

```sh
go run ./solutions/18-goroutines/joined-task
```

예상 출력:

```text
종료까지 기다림: true
```

### 수명 테스트 전체

테스트는 시작 여부를 신호로 확인하고 취소합니다.
함수가 반환했을 때 종료 신호가 이미 닫혀 있어야 합니다.
5초 제한은 교착 상태를 실패로 표시하기 위한 장치이며,
5초 기다리면 정상이라는 의미가 아닙니다.

테스트의 발췌:

```go
func TestJoinedCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	finished := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- runJoined(ctx, func(ctx context.Context) error { defer close(finished); close(started); <-ctx.Done(); return ctx.Err() })
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("start timeout")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		select {
		case <-finished:
		default:
			t.Fatal("returned before task finished")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("join timeout")
	}
}
```

실제 `task`가 Context를 무시하고 영원히 반환하지 않으면
이 함수도 기다립니다.
강제로 task를 죽이고 안전하게 자원을 회수하는 기능은 제공하지 않습니다.
따라서 task 계약에 취소 협조와 유한한 종료 조건이 포함되어야 합니다.

## 문제 7. 선택 과제의 설계

입력 위치마다 시작 게이트 채널을 하나 만듭니다.
작업은 자기 게이트가 닫힐 때까지 기다린 뒤 자기 결과 칸에 씁니다.
테스트는 게이트를 2, 0, 1 순서로 닫고 각 단계의 완료 신호를 기다립니다.

게이트를 연 순서만으로 실제 완료 순서까지 단정할 수는 없습니다.
완료 순서 자체를 확인하려면 완료 확인 신호도 함께 받아야 합니다.
모든 작업을 기다린 뒤 결과 슬라이스를 읽으면 입력 순서가 유지됩니다.

이 구조는 다음 장의 채널과 select 연습에서 구현합니다.
Sleep 길이를 10ms, 20ms, 30ms로 다르게 주는 방식은
운영체제 스케줄링에 따라 흔들리므로 순서 증명의 수단으로 사용하지 않습니다.

## 검증

```sh
go fmt ./solutions/18-goroutines/...
go vet ./solutions/18-goroutines/...
go test ./solutions/18-goroutines/...
go build ./solutions/18-goroutines/...
```

Linux와 C 컴파일러가 준비된 환경:

```sh
go test -race ./solutions/18-goroutines/...
```

각 프로그램에는 문서의 출력과 같은 `Example_main`이 있습니다.
함수 테스트는 정상·상한·오류·취소와 완료 관계를 따로 검증합니다.

## 스스로 평가하기

| 항목 | 확인 기준 |
| --- | --- |
| 등록 | 새 고루틴보다 먼저 작업 수를 등록 |
| 종료 | 조기 반환에서도 Done 또는 Go의 완료 처리가 실행 |
| 소유권 | 공유 쓰기 대상과 각 인덱스의 담당자가 명확 |
| 결과 순서 | 입력 위치와 출력 위치가 대응 |
| 오류 | 실패한 결과를 성공으로 반환하지 않음 |
| 취소 | 요청뿐 아니라 실제 종료를 기다림 |
| 상한 | 입력과 작업자 수를 검증 |
| 테스트 | Sleep 대신 신호·반환값으로 확인 |

## 참고 자료

- [WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [WaitGroup.Go](https://pkg.go.dev/sync#WaitGroup.Go)
- [Go 메모리 모델](https://go.dev/ref/mem)
- [errors.Join](https://pkg.go.dev/errors#Join)
- [Context](https://pkg.go.dev/context)

[문제로 돌아가기](../../exercises/18-goroutines/README.md) ·
[본문으로 돌아가기](../../book/18-goroutines/README.md)
