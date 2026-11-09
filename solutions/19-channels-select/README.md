# 19장 해설 — 채널과 select

[연습문제](../../exercises/19-channels-select/README.md)의 참고 풀이입니다.
명령은 모두 저장소 루트에서 실행합니다.

핵심은 채널 개수보다 전달과 종료의 관계입니다.
각 출력의 닫기 소유자를 하나로 정하고,
중간 소비 중단을 예상한 전송에는 취소 경로를 둡니다.

## 문제 1. 채널의 상태

| 상태 | 전송 | 수신 | close |
| --- | --- | --- | --- |
| nil | 계속 기다림 | 계속 기다림 | 패닉 |
| 열린 버퍼 없음, 상대 없음 | 수신자 대기 | 송신자 대기 | 가능 |
| 열린 버퍼 가득 참 | 여유 공간 대기 | 준비된 값 수신 | 가능 |
| 닫힘, 버퍼 값 있음 | 패닉 | 남은 값과 true | 다시 닫으면 패닉 |
| 닫힘, 버퍼 비었음 | 패닉 | 영값과 false 즉시 반환 | 다시 닫으면 패닉 |

정상 데이터 0과 종료는 ok로 구분합니다.
select의 첫 case가 우선되는 것은 아닙니다.
default가 있는 빈 반복은 바쁘게 돌며 CPU를 사용할 수 있습니다.

버퍼가 끝없이 커지는 것이 아니므로 소비 중단에는 취소가 필요합니다.
고정된 한 값을 주고받고 더 사용하지 않는 모든 채널을 닫을 의무는 없습니다.
슬라이스를 전달해도 내부 배열은 공유할 수 있습니다.
송신 전용 타입은 여러 함수가 모두 닫아도 된다는 소유권 보장이 아닙니다.

## 풀이 코드

| 문제 | 소스 | 소유권 |
| --- | --- | --- |
| 2 | [filter-positive](filter-positive/main.go) | 각 단계가 자신의 출력 닫기 |
| 3 | [take-values](take-values/main.go) | 외부 입력을 닫지 않는 소비자 |
| 4 | [merge-results](merge-results/main.go) | 모든 생산자를 기다린 담당자가 닫기 |
| 6 | [ordered-gates](ordered-gates/main.go) | 시작 허가와 완료 확인을 분리 |

## 문제 2. 한 방향으로 흐르는 양수 필터


생산자는 입력을 모두 보내거나 취소되면 in을 닫습니다.
필터는 in의 닫힘 또는 취소를 확인하고 out을 닫습니다.
소비자는 out을 끝까지 읽은 뒤 두 단계를 모두 기다립니다.

필터가 0과 음수를 버리는 것은 요구된 데이터 변환입니다.
채널이 가득 차서 데이터를 우연히 버리는 것과 다릅니다.

소스: [filter-positive/main.go](filter-positive/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
)

func positive(ctx context.Context, values []int) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(values) > 1000 {
		return nil, errors.New("값은 1000개까지 허용합니다")
	}
	in := make(chan int)
	out := make(chan int)
	var group sync.WaitGroup
	group.Go(func() {
		defer close(in)
		for _, value := range values {
			select {
			case <-ctx.Done():
				return
			case in <- value:
			}
		}
	})
	group.Go(func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-in:
				if !ok {
					return
				}
				if value > 0 {
					select {
					case <-ctx.Done():
						return
					case out <- value:
					}
				}
			}
		}
	})
	results := []int{}
	for value := range out {
		results = append(results, value)
	}
	group.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
func main() {
	values, err := positive(context.Background(), []int{-1, 2, 0, 3})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("양수:", values)
}
```

```sh
go run ./solutions/19-channels-select/filter-positive
```

예상 출력:

```text
양수: [2 3]
```


### 순서와 종료

생산자가 입력 순서로 보내고 필터가 한 고루틴에서 처리하므로
양수가 등장한 상대적 순서는 유지됩니다.
취소가 전달되면 각 단계의 대기 지점에 빠져나갈 경로가 있습니다.

함수는 Context 상태를 마지막에도 확인합니다.
이미 모은 양수가 있더라도 취소되었다면 전체 성공 결과로 반환하지 않습니다.
부분 결과를 보존할 것인지는 다른 설계이며 현재 계약에서는 nil입니다.

### 테스트의 관찰 대상

테스트는 입력 순서·중복·빈 입력·상한·취소를 확인합니다.
고루틴이 어떤 CPU에서 실행했는지 확인하지 않습니다.
그 정보는 필터의 외부 계약에 포함되지 않습니다.

## 문제 3. 외부 생산자의 수명을 소유하지 않는 소비자


take는 입력의 일부만 읽는 작은 함수입니다.
자신이 만들지 않은 채널을 닫거나 생산자를 강제로 끝내지 않습니다.
생산자가 있는 호출 구조에서는 호출자가 Context 취소와 합류를 맡아야 합니다.

소스: [take-values/main.go](take-values/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func take(ctx context.Context, in <-chan int, limit int) ([]int, error) {
	if limit < 0 || limit > 1000 {
		return nil, errors.New("개수는 0~1000이어야 합니다")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	results := make([]int, 0, limit)
	for len(results) < limit {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case value, ok := <-in:
			if !ok {
				return results, nil
			}
			results = append(results, value)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
func main() {
	in := make(chan int, 4)
	for value := 1; value <= 4; value++ {
		in <- value
	}
	close(in)
	values, err := take(context.Background(), in, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("받은 값:", values)
	fmt.Println("남은 첫 값:", <-in)
}
```

```sh
go run ./solutions/19-channels-select/take-values
```

예상 출력:

```text
받은 값: [1 2]
남은 첫 값: 3
```


### 닫힘을 정상 종료로 처리하기

입력 채널이 limit 전에 닫히면 모은 값을 정상 반환합니다.
최소 개수를 강제하는 함수가 아니라 최대 개수만 제한하는 함수이기 때문입니다.
`ok`를 확인하지 않고 영값을 채우면 계약이 달라집니다.

limit이 0이면 채널에 손대지 않고 빈 결과를 만들 수 있습니다.
이미 취소되었다면 현재 계약에 따라 취소 오류를 우선 반환합니다.

### 남은 값을 읽을 수 있는 이유

버퍼에 네 값을 넣고 닫았어도 앞의 두 값만 소비했습니다.
나머지 3과 4는 그대로 남아 있습니다.
take의 반환이 입력의 버퍼를 비우거나 닫는 작업은 아닙니다.

이 함수에 nil 채널과 Background를 주고 limit을 양수로 두면 기다립니다.
호출자는 실제 생산자가 있거나 취소·마감이 있는지 보장해야 합니다.

## 문제 4. 닫기 책임을 한 곳으로 모으기


각 입력 묶음마다 생산자를 하나 둡니다.
모든 생산자가 같은 out으로 보내므로 개별 생산자는 out을 닫지 않습니다.
합류 담당자만 모든 생산자를 기다린 뒤 out과 done을 닫습니다.

소스: [merge-results/main.go](merge-results/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
)

func collect(ctx context.Context, parts [][]int) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(parts) > 8 {
		return nil, errors.New("입력 묶음은 8개까지 허용합니다")
	}
	total := 0
	for _, part := range parts {
		if len(part) > 1000-total {
			return nil, errors.New("전체 값은 1000개까지 허용합니다")
		}
		total += len(part)
	}
	out := make(chan int)
	var group sync.WaitGroup
	for _, part := range parts {
		group.Go(func() {
			for _, value := range part {
				select {
				case <-ctx.Done():
					return
				case out <- value:
				}
			}
		})
	}
	done := make(chan struct{})
	go func() { group.Wait(); close(out); close(done) }()
	results := make([]int, 0, total)
	for value := range out {
		results = append(results, value)
	}
	<-done
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sort.Ints(results)
	return results, nil
}
func main() {
	values, err := collect(context.Background(), [][]int{{3, 1}, {2, 3}, nil})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("모은 값:", values)
}
```

```sh
go run ./solutions/19-channels-select/merge-results
```

예상 출력:

```text
모은 값: [1 2 3 3]
```


### 결과를 모으는 사람은 한 명

results에 append하는 곳은 호출자 하나입니다.
생산자들은 채널 전송만 수행하므로 공유 append 경쟁이 없습니다.
완료 순서는 달라도 모든 값을 모아 정렬하므로 반환 순서는 안정적입니다.

정렬은 중복을 없애지 않습니다.
서로 다른 묶음의 3 두 개가 모두 남아야 합니다.

### 입력의 상한

각 묶음의 길이를 더하기 전에 `len(part) > 1000-total`을 검사합니다.
상한을 넘기는 입력을 작업 시작 전에 거부합니다.
단순 합산 후 비교보다 범위를 함께 관리한다는 의도가 드러납니다.

8개 묶음이더라도 빈 묶음은 정상입니다.
입력 묶음이 없어도 합류 담당자가 출력 채널을 닫아 소비자가 종료됩니다.

## 문제 5. 조기 중단에 필요한 두 단계

원래 generate는 네 번째 값을 보내다 기다릴 수 있습니다.
소비자가 이미 반환해 수신할 곳이 없기 때문입니다.
버퍼 3으로 바꾸면 기다리는 시점이 뒤로 밀릴 뿐 생산자 종료를 보장하지 않습니다.

소비자가 out을 닫으면 아직 전송 중인 생산자가 패닉을 일으킬 수 있습니다.
생산자의 채널을 닫는 대신 Context 취소를 요청합니다.

수정된 전송의 발췌:

```go
select {
case <-ctx.Done():
    return
case out <- value:
}
```

호출자의 정리는 다음 순서여야 합니다.

```go
defer func() {
    cancel()
    group.Wait()
}()
```

먼저 Wait하고 나중에 cancel하면 생산자가 수신자를 기다리는 동안
호출자도 생산자를 기다리게 되어 끝나지 않을 수 있습니다.

완성된 두 단계 구현은
[pipeline/main.go](../../examples/19-channels-select/pipeline/main.go)에 있습니다.
검증은 [pipeline의 경계 테스트](../../examples/19-channels-select/pipeline/behavior_test.go)에서 확인합니다.

| 경계 | 확인할 의미 |
| --- | --- |
| count=0, take=0 | 빈 데이터에서도 합류 |
| count=1000, take=0 | 소비하지 않고 취소해도 합류 |
| count=1000, take=3 | 일부 소비 뒤 양쪽 단계 종료 |
| count=3, take=3 | 모든 데이터 정상 소비 |
| take>count | 시작 전에 오류 |

```sh
go run ./examples/19-channels-select/pipeline
go test ./examples/19-channels-select/pipeline
```

## 문제 6. 시작 허가와 완료 확인


전체 순열이 올바른지 먼저 검사합니다.
그 다음 작업마다 독립된 게이트를 만들고 모든 작업자를 시작합니다.
호출자는 하나의 게이트를 연 뒤 해당 완료값을 받고 다음으로 진행합니다.

소스: [ordered-gates/main.go](ordered-gates/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

func ordered(order []int) ([]int, []int, error) {
	if len(order) > 16 {
		return nil, nil, errors.New("작업은 16개까지 허용합니다")
	}
	seen := make([]bool, len(order))
	for _, index := range order {
		if index < 0 || index >= len(order) || seen[index] {
			return nil, nil, errors.New("0부터 작업 수-1까지의 순열이 필요합니다")
		}
		seen[index] = true
	}
	gates := make([]chan struct{}, len(order))
	for index := range gates {
		gates[index] = make(chan struct{})
	}
	results := make([]int, len(order))
	completed := make(chan int)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() { <-gates[index]; results[index] = index * 10; completed <- index })
	}
	actual := make([]int, 0, len(order))
	for _, index := range order {
		close(gates[index])
		actual = append(actual, <-completed)
	}
	group.Wait()
	return results, actual, nil
}
func main() {
	values, completed, err := ordered([]int{2, 0, 1})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("완료 순서:", completed)
	fmt.Println("입력 순서 결과:", values)
}
```

```sh
go run ./solutions/19-channels-select/ordered-gates
```

예상 출력:

```text
완료 순서: [2 0 1]
입력 순서 결과: [0 10 20]
```


### 왜 완료 채널도 필요한가?

게이트를 2, 0, 1 순서로 빠르게 닫는 것만으로는
스케줄러가 작업을 같은 순서로 끝내도록 보장할 수 없습니다.
시작할 수 있다는 허가와 실제 완료는 서로 다른 사건입니다.

코드는 게이트 하나를 열고 완료값 하나를 받은 다음 다음 게이트를 엽니다.
따라서 완료 순서도 입력 order와 같아집니다.
최종 결과는 각 인덱스에 저장하므로 원래 입력 순서를 유지합니다.

### 모든 채널을 닫지 않는 이유

게이트는 한 번 열리는 신호이므로 닫습니다.
completed는 작업 수만큼 정확히 받는 프로토콜이며 range로 끝을 기다리지 않습니다.
모든 작업을 Wait한 뒤 더 사용하지 않으므로 completed를 닫을 필요는 없습니다.

순열 검증을 실행 중에 하면 일부 작업의 게이트가 열리지 않은 채 남을 수 있습니다.
이런 작업은 Wait를 끝내지 못하므로 검증은 시작 전에 끝냅니다.

## 의미 있는 테스트를 작성하기

### 취소의 종료 확인

단순히 취소 함수를 호출했다고 테스트를 끝내지 않습니다.
생산자의 완료 신호를 기다린 뒤 닫힌 출력 상태를 확인합니다.

본문 예제의 테스트는 하나의 실제 값을 받은 뒤 취소합니다.
그 다음 더 이상 출력값을 소비하지 않습니다.
그래도 송신자가 종료해야 하므로 막힌 전송의 취소 경로를 검증할 수 있습니다.

[취소 전송 테스트](../../examples/19-channels-select/cancel-send/behavior_test.go)를 참고합니다.

### nil 입력과 빈 입력의 차이

닫힌 빈 채널은 즉시 종료를 알립니다.
nil 채널은 통신하지 않습니다.
따라서 nil을 “빈 채널과 같은 것”으로 처리하는 테스트는 잘못된 가정을 포함합니다.

병합 예제는 nil 입력을 명시적으로 건너뜁니다.
take는 입력의 존재를 보장하지 않으므로 nil 입력에서 취소를 기다릴 수 있다고 문서화합니다.
같은 언어 동작 위에서도 API의 선택은 다를 수 있습니다.

### 제한 시간을 쓰는 목적

테스트가 무한히 멈추지 않도록 제한 시간을 두는 것은 가능합니다.
그 제한 시간이 경과하면 성공이 아니라 실패입니다.
정상 여부는 완료 신호·결과·오류로 판단합니다.

## 검증 명령

```sh
go fmt ./solutions/19-channels-select/...
go vet ./solutions/19-channels-select/...
go test ./solutions/19-channels-select/...
go build ./solutions/19-channels-select/...
```

Linux와 C 컴파일러가 준비된 환경:

```sh
go test -race ./solutions/19-channels-select/...
```

## 평가 기준

| 항목 | 통과 기준 |
| --- | --- |
| 전달 | 값과 개수를 계약대로 보존 |
| 순서 | 입력·완료·정렬 순서를 구분 |
| 닫기 | 마지막 전송을 아는 하나의 주체만 닫음 |
| 취소 | 전송과 수신의 대기 모두 빠져나올 수 있음 |
| 합류 | 시작한 작업이 끝난 뒤 반환 |
| 경계 | 빈 입력·nil·상한·취소를 의도대로 처리 |

## 참고 자료

- [채널 타입](https://go.dev/ref/spec#Channel_types)
- [close](https://go.dev/ref/spec#Close)
- [select](https://go.dev/ref/spec#Select_statements)
- [파이프라인과 취소](https://go.dev/blog/pipelines)
- [Context](https://pkg.go.dev/context)

[문제로 돌아가기](../../exercises/19-channels-select/README.md) ·
[본문으로 돌아가기](../../book/19-channels-select/README.md)
