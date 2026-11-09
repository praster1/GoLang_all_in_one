# 17장 해설 — Context

[연습문제](../../exercises/17-context/README.md)의 참고 구현입니다.
모든 명령은 저장소 루트에서 실행합니다.
예제와 테스트는 외부 서비스 없이 종료됩니다.

기본 문제는 Context의 상태를 동기적인 호출로 관찰합니다.
고루틴을 먼저 만들고 잠든 뒤 결과를 추측하지 않습니다.
반환값·호출 횟수·자식의 정리 상태를 직접 확인합니다.

| 문제 | 구현 | 중요한 계약 |
| --- | --- | --- |
| 2 | [checked-sum](checked-sum/main.go) | 취소·입력 오류에서 0 |
| 4 | [scoped-task](scoped-task/main.go) | child 수명의 정리 |
| 5 | [context-values](context-values/main.go) | 값과 취소 연결 |
| 6 | [cancel-batch](cancel-batch/main.go) | 전체 성공 시에만 결과 반환 |

## 문제 1. 취소 방향과 상태

| 사건 | parent.Err | child.Err | sibling.Err |
| --- | --- | --- | --- |
| 생성 직후 | nil | nil | nil |
| child만 cancel | nil | Canceled | nil |
| parent cancel | Canceled | Canceled | Canceled |
| child를 다시 cancel | Canceled | Canceled | Canceled |

child의 취소는 parent와 sibling에 전달되지 않습니다.
공통 parent의 취소는 그 아래 자식들에 전달됩니다.
이미 취소된 Context를 다시 취소해도 상태가 초기화되지 않습니다.

### 질문의 답

1. Done은 취소 상태를 알립니다. 실제 계산이 끝났는지는 함수 반환이나 별도 완료 신호로 확인합니다.
2. Background의 Done은 nil 채널입니다. 단독 수신은 계속 기다립니다.
3. 부모의 마감보다 늦어지지 않는 쪽이 적용됩니다.
4. 마감 전에 수동 취소하면 일반 Canceled가 될 수 있습니다.
5. CancelCauseFunc에 nil을 넘겨도 취소하며 원인은 Canceled가 됩니다.

Context에 취소 메서드가 없는 이유도 이 방향과 관련됩니다.
상태를 관찰하는 모든 하위 함수에 부모를 중단할 권한을 주지 않습니다.
취소 함수를 가진 수명 소유자가 요청합니다.

[Context](https://pkg.go.dev/context#Context) ·
[CancelFunc](https://pkg.go.dev/context#CancelFunc)

## 문제 2. 합계의 성공 경계

소스: [checked-sum/main.go](checked-sum/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func sum(ctx context.Context, values []int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if len(values) > 1000 {
		return 0, errors.New("값은 1000개까지 허용합니다")
	}
	total := 0
	for _, value := range values {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if value < 0 || value > 1000 {
			return 0, errors.New("각 값은 0~1000이어야 합니다")
		}
		total += value
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return total, nil
}

func run() error {
	total, err := sum(context.Background(), []int{10, 20, 30})
	if err != nil {
		return err
	}
	fmt.Println("합계:", total)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = sum(ctx, []int{10, 20, 30})
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./solutions/17-context/checked-sum
```

예상 결과:

```text
합계: 60
취소 확인: true
```

### 입구의 검사

빈 슬라이스여도 입구에서 취소를 확인합니다.
이미 필요 없어진 요청을 “할 일이 없으니 성공”으로 바꾸지 않는 정책입니다.
잘못된 값과 취소가 함께 주어진 경우도 취소 오류가 우선합니다.

값 개수와 각 값의 상한 때문에 정상 합계는 최대 1000000입니다.
이 상한은 32비트 int에서도 표현할 수 있습니다.
입력 규칙을 바꾸면 합계의 범위도 다시 검토해야 합니다.

### 루프 뒤의 검사

마지막 덧셈 이후 취소를 관찰하면 0과 오류를 반환합니다.
중간 합계를 돌려주지 않는 계약이므로 완료된 덧셈 개수는 반환하지 않습니다.
본문 process가 진행 개수를 돌려주는 정책과 구분합니다.

검사 직후 다른 곳에서 취소될 가능성을 완전히 없애지는 못합니다.
함수의 성공 반환과 모든 동시 취소를 원자적으로 직렬화하는 API가 아닙니다.
여기서는 정해진 확인 지점에서 관찰한 상태를 기준으로 합니다.

[테스트](checked-sum/main_test.go)는
0개·상한·상한 초과·취소와 잘못된 입력의 우선순위·과거 마감을 검사합니다.
짧은 덧셈 중간에 우연히 취소가 끼어들기를 기다리지 않습니다.

## 문제 3. 받은 Context를 사용하기

잘못된 프로그램의 출력은 다음과 같습니다.

```text
성공으로 처리: true
```

work가 전달받은 ctx를 버리고 Background의 Err를 확인했기 때문입니다.
Background는 취소되지 않으므로 호출자의 취소를 볼 수 없습니다.

수정한 전체 프로그램은 다음과 같습니다.

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

func work(ctx context.Context) error {
	return ctx.Err()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := work(ctx)
	fmt.Println("성공으로 처리:", err == nil)
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
}
```

예상 결과:

```text
성공으로 처리: false
취소 확인: true
```

새로운 제한이나 독립적인 수명이 필요하지 않으므로 WithCancel을 추가하지 않았습니다.
취소를 확인해 원래 오류를 전달하는 것으로 충분합니다.

### 자식만 반환하는 함수의 문제

함수가 자식을 만들고 `defer cancel()`을 등록한 뒤 자식만 반환하면
반환 과정에서 cancel이 실행됩니다.
호출자가 다음 작업에 쓰려 할 때 이미 취소된 Context입니다.

작업을 함수 안에서 끝내면 defer가 알맞습니다.
반환 뒤 수명이 필요하면 호출자가 정리 책임을 받을 수 있도록 API를 바꿔야 합니다.
defer 자체가 잘못된 것이 아니라 수명 경계가 요구사항과 다른 것입니다.

## 문제 4. 시간 예산은 협력 계약이다

소스: [scoped-task/main.go](scoped-task/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

func withBudget(ctx context.Context, budget time.Duration, task func(context.Context) error) error {
	if budget < 0 || task == nil {
		return errors.New("음수 예산과 nil 작업은 허용하지 않습니다")
	}
	child, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	if err := child.Err(); err != nil {
		return err
	}
	if err := task(child); err != nil {
		return fmt.Errorf("작업: %w", err)
	}
	return child.Err()
}

func run() error {
	var observed context.Context
	err := withBudget(context.Background(), time.Hour, func(ctx context.Context) error {
		observed = ctx
		return ctx.Err()
	})
	if err != nil {
		return err
	}
	fmt.Println("작업 성공:", err == nil)
	fmt.Println("반환 후 정리:", errors.Is(observed.Err(), context.Canceled))
	called := false
	err = withBudget(context.Background(), 0, func(context.Context) error {
		called = true
		return nil
	})
	fmt.Println("0초 마감:", errors.Is(err, context.DeadlineExceeded))
	fmt.Println("0초 작업 실행:", called)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./solutions/17-context/scoped-task
```

예상 결과:

```text
작업 성공: true
반환 후 정리: true
0초 마감: true
0초 작업 실행: false
```

### 정상 완료 뒤 취소 상태

task가 성공했다는 반환값과 child가 정리되었다는 관찰은 서로 모순이 아닙니다.
task 실행 중에는 Err가 nil이었고,
withBudget이 반환할 때 defer cancel이 실행되어 수명이 끝납니다.

이 예제는 관찰을 위해 자식 Context를 변수에 보관합니다.
실제 서비스에서 완료한 요청 Context를 계속 전역에 쌓아 두라는 패턴은 아닙니다.
자식의 정리를 테스트하기 위한 제한된 사용입니다.

### 왜 0초에서 task를 호출하지 않는가?

WithTimeout의 0초는 즉시 마감된 자식이 됩니다.
호출 전에 child.Err를 검사하므로 task는 0번 실행됩니다.
음수는 이 함수의 자체 입력 규칙으로 먼저 거부합니다.

| 경로 | task 호출 | 정리 |
| --- | --- | --- |
| 정상 예산·성공 | 한 번 | child cancel |
| 정상 예산·일반 오류 | 한 번 | child cancel |
| 0초 | 없음 | 생성한 child cancel |
| 부모가 취소됨 | 없음 | 생성한 child cancel |
| 음수·nil task | 없음 | child를 만들지 않음 |

### 일반 오류와 취소가 함께 생기면

task가 일반 오류를 반환하면 그 오류를 감싸서 반환합니다.
task가 nil을 반환한 경우에만 child.Err를 마지막으로 확인합니다.
이 함수는 일반 작업 오류를 우선하는 계약입니다.

다른 정책이 필요하다면 오류를 결합하거나 결과 구조를 바꿀 수 있습니다.
그러나 함수 설명·테스트·호출자의 분류 방식도 함께 바뀌어야 합니다.
코드의 우연한 if 순서를 외부 계약으로 숨기지 않습니다.

### 강제 중단 래퍼가 아니다

task는 같은 고루틴에서 직접 호출합니다.
task가 Context를 무시하면서 반환하지 않으면 withBudget도 반환하지 않습니다.
타임아웃 Context를 만들었다는 사실만으로 강제 중단 기능을 얻지 못합니다.

[테스트](scoped-task/main_test.go)는
정상·오류 뒤 child 정리, 부모 유지, 사전 취소, 0초, nil task를 확인합니다.
큰 양수 예산을 써도 task는 즉시 반환하므로 테스트가 그 시간을 기다리지 않습니다.

## 문제 5. 키 타입과 부모 연결

소스: [context-values/main.go](context-values/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type traceIDKey struct{}

func withTraceID(ctx context.Context, id string) (context.Context, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("추적 ID는 비어 있을 수 없습니다")
	}
	return context.WithValue(ctx, traceIDKey{}, id), nil
}

func traceID(ctx context.Context) (string, bool) {
	id, exists := ctx.Value(traceIDKey{}).(string)
	return id, exists
}

func run() error {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	child, err := withTraceID(parent, " trace-001 ")
	if err != nil {
		return err
	}
	id, exists := traceID(child)
	fmt.Println("추적 ID:", id, exists)
	cancel()
	<-child.Done()
	fmt.Println("취소 전달:", errors.Is(child.Err(), context.Canceled))
	id, exists = traceID(child)
	fmt.Println("값 유지:", id, exists)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./solutions/17-context/context-values
```

예상 결과:

```text
추적 ID: trace-001 true
취소 전달: true
값 유지: trace-001 true
```

### 값의 검증과 조회

withTraceID는 먼저 공백을 다듬고 빈 값을 거부합니다.
따라서 성공한 helper를 통해 저장한 ID는 비어 있지 않습니다.
조회 함수는 그래도 존재 여부를 반환하여 ID가 없는 Context를 처리합니다.

오류일 때 반환된 nil Context는 사용하지 않습니다.
호출자는 오류를 먼저 검사한 뒤 성공한 Context만 넘깁니다.
Context 함수에 nil을 전달하는 관습을 만드는 것이 아닙니다.

### 자체 키 타입

`traceIDKey{}`는 비교 가능한 빈 구조체 값입니다.
같은 자체 타입의 같은 값으로 저장하고 조회합니다.
다른 타입의 키를 사용한 값과 충돌하지 않습니다.

키가 비공개이므로 다른 패키지가 이 타입을 직접 만들어
우연히 같은 키를 사용하는 가능성을 줄입니다.
[WithValue](https://pkg.go.dev/context#WithValue)의 키 지침을 따릅니다.

### 취소와 값은 별개다

WithValue는 취소 연결을 끊지 않습니다.
부모를 취소하면 자식 Err에도 Canceled가 나타납니다.
그 뒤에도 저장된 문자열 ID를 조회할 수 있습니다.

취소를 정보 삭제나 비밀 값 제거 기능으로 해석하지 않습니다.
요청 ID처럼 변경하지 않는 작은 메타데이터를 담습니다.
맵·포인터를 넣고 동시 변경을 자동 보호받는다고 생각하지 않습니다.

[테스트](context-values/main_test.go)는
공백·빈 값·부모와 자식의 다른 값·다른 키 타입·마감 전달을 검사합니다.
정상 예제는 취소 뒤 값 유지도 출력으로 확인합니다.

## 문제 6. 결과 공개와 부수 효과

소스: [cancel-batch/main.go](cancel-batch/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func applyAll(ctx context.Context, values []int, transform func(context.Context, int) (int, error)) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if transform == nil || len(values) > 1000 {
		return nil, errors.New("변환 함수와 1000개 이하의 값이 필요합니다")
	}
	results := make([]int, 0, len(values))
	for _, value := range values {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err := transform(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("값 %d 변환: %w", value, err)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func run() error {
	normal, err := applyAll(context.Background(), []int{1, 2, 3}, func(ctx context.Context, value int) (int, error) {
		return value * 2, ctx.Err()
	})
	if err != nil {
		return err
	}
	fmt.Println("정상 결과:", normal)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results, err := applyAll(ctx, []int{1, 2, 3}, func(ctx context.Context, value int) (int, error) {
		if value == 2 {
			cancel()
		}
		return value * 2, nil
	})
	fmt.Println("실패 결과 없음:", results == nil)
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./solutions/17-context/cancel-batch
```

예상 결과:

```text
정상 결과: [2 4 6]
실패 결과 없음: true
취소 확인: true
```

### 호출 뒤에도 확인하는 이유

transform 안에서 취소를 요청하고 nil을 반환할 수 있습니다.
호출 뒤 검사를 생략하면 마지막 항목의 취소를 놓치거나
취소 후 계산 결과를 슬라이스에 넣는 경로가 생길 수 있습니다.

이 구현은 transform의 성공 반환 뒤 ctx.Err를 확인합니다.
취소를 관찰하면 결과를 추가하지 않고 nil 결과를 반환합니다.
앞에서 모은 슬라이스도 공개하지 않습니다.

### 마지막 항목에서의 취소

| 취소 지점 | transform 호출 횟수 | 결과 |
| --- | --- | --- |
| 호출 전에 취소 | 0 | nil, Canceled |
| 첫 호출에서 취소 | 1 | nil, Canceled |
| 마지막 세 번째 호출에서 취소 | 3 | nil, Canceled |
| 취소 없음 | 3 | 전체 변환 슬라이스 |

입력 순서대로 동기 호출하므로 호출 횟수를 확정적으로 검사할 수 있습니다.
고루틴의 실행 순서나 잠드는 시간에 기대지 않습니다.

### nil과 빈 슬라이스

정상 빈 입력은 `make([]int, 0, 0)`으로 만든 빈 슬라이스입니다.
오류의 nil 결과와 코드에서 구분할 수 있습니다.
실제로 사용하는 쪽은 항상 오류를 먼저 확인해야 합니다.

### 이 함수가 롤백하지 않는 것

transform이 파일을 썼다면 nil 슬라이스를 반환해도 파일 쓰기는 남습니다.
transform이 외부 서비스 요청을 보냈다면 그 요청도 자동 취소·복구되지 않습니다.
이 문제는 반환값 공개의 정책을 구현합니다.

외부 변경을 다루는 프로그램은 멱등성·트랜잭션·재시도 정책을 따로 설계해야 합니다.
31장의 종합 동시 처리에서 작업 오류와 전체 취소를 다시 연결합니다.

[테스트](cancel-batch/main_test.go)는
정상 순서·0개·1000/1001개·nil callback,
사전 취소·첫/마지막 호출 취소·중간 일반 오류를 확인합니다.
오류 원인은 errors.Is로 관찰합니다.

## 선택 문제. 취소 가능한 대기의 경계

본문의 wait는 입구에서 취소를 먼저 확인합니다.
따라서 취소된 Context와 음수 duration을 함께 주면 취소 오류가 먼저입니다.
살아 있는 Context의 음수는 일반 입력 오류입니다.

| 조건 | 관찰 |
| --- | --- |
| Background와 0초 | timer 신호 후 nil |
| 사전 취소와 1시간 | 입구에서 Canceled |
| 과거 마감과 1시간 | 입구에서 DeadlineExceeded |
| Background와 음수 | 입력 오류 |
| 사전 취소와 음수 | 취소 오류 우선 |

0초 정상 대기에서 Background.Done은 nil입니다.
select의 해당 분기는 선택되지 않고 타이머 분기로 진행할 수 있습니다.
nil 채널에 단독 수신하는 코드와 같은 동작이라고 생각하지 않습니다.

### 동시에 준비된 분기

select는 소스 위쪽 case를 우선순위로 사용하지 않습니다.
이 함수는 입구와 타이머 분기 뒤에서 Err를 확인하여
관찰 가능한 취소에 대한 정책을 드러냅니다.

그래도 검사 바로 뒤의 취소를 모든 반환과 원자적으로 묶는 기능은 없습니다.
이 한계를 설명한 뒤 필요한 도메인의 결과·취소 우선순위를 정합니다.
[select 명세](https://go.dev/ref/spec#Select_statements)를 참고합니다.

### 테스트의 전체 제한 시간

검사 프로세스의 timeout은 잘못된 구현이 계속 기다리는 것을 막기 위한 장치입니다.
timeout을 만났으면 성공이 아니라 검사 실패입니다.
정상 사례는 채널 신호·반환 오류·호출 횟수로 판정합니다.

## 검증 명령

```sh
go test ./examples/17-context/... ./solutions/17-context/...
go vet ./examples/17-context/... ./solutions/17-context/...
go build ./examples/17-context/... ./solutions/17-context/...
```

이 장의 기본 테스트에는 직접 만든 비동기 작업이 없습니다.
Context 생성·취소·콜백 호출이 같은 실행 흐름에서 일어납니다.
후속 장의 고루틴 예제에서는 작업 완료와 공유 데이터 동기화를 추가로 검사합니다.

### 비교할 때 남길 설명

- 오류가 생기면 어떤 결과 값을 반환하는가?
- 취소를 어디에서 확인하는가?
- 만들어진 자식 Context는 누가 언제 정리하는가?
- 부모에게 남기는 영향은 무엇인가?
- 일반 작업 오류와 취소가 함께 보이면 무엇을 우선하는가?
- 테스트가 실행 시간을 추측하지 않고 상태를 제어하는가?

[본문](../../book/17-context/README.md) ·
[연습문제](../../exercises/17-context/README.md)

확인일: 2026-11-09. 코드·출력·오류·정리는 Go 1.27.3에서 확인했습니다.
