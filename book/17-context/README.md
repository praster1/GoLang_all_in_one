# 17. Context로 취소와 마감 전달하기

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 응용 | 함수 값·오류·defer·테스트, 시간과 Duration | Go, 터미널, 이 저장소 |

사용자가 작업을 취소했는데 내부 처리가 계속되는 상황을 생각해 봅시다.
명령을 받은 함수만 중단하고 그 아래 함수가 계속 읽거나 계산하면
사용자가 원하지 않는 일을 오래 수행하게 됩니다.

`context.Context`는 호출의 경계를 넘어
**언제 작업을 중단해야 하는지**와 관련 정보를 전달하는 표준 방법입니다.
작업을 강제로 끊는 스위치나 자동으로 결과를 기다려 주는 객체는 아닙니다.

이번 장은 순차 실행부터 시작합니다.
직접 고루틴을 만들지 않고도 취소된 Context의 상태,
부모·자식 관계, 반복문의 확인 지점을 배울 수 있습니다.
고루틴과 채널 전체 문법은 18·19장에서 이어서 배웁니다.

## 학습 목표

- Context의 마감·취소 상태·Done 신호·값의 역할을 구분한다.
- 부모에서 자식으로 전달되는 취소와 더 이른 마감의 의미를 설명한다.
- 취소 함수를 소유한 범위에서 정리하고 호출 아래로 Context를 전달한다.
- 반복과 하위 작업에 취소 확인 지점을 두고 오류 정책을 정한다.
- 취소 상태와 구체적인 원인을 구분하고 잠들지 않는 테스트를 작성한다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
예제는 외부 네트워크·계정·파일을 사용하지 않습니다.
현재 시각을 고정된 출력으로 요구하지 않으며,
시간 초과를 보여 주기 위해 몇 초 동안 잠들지도 않습니다.

기본 시간 예제는 이미 지난 마감 또는 0초 제한을 사용합니다.
이런 입력은 작업을 시작하기 전에 마감된 상황을 재현합니다.
실제 서비스의 시간 예산을 0초로 정하라는 권장 사항은 아닙니다.

| 예제 | 중심 내용 | 실행 |
| --- | --- | --- |
| [background](../../examples/17-context/background/main.go) | 취소 없는 시작 Context | `go run ./examples/17-context/background` |
| [cancel](../../examples/17-context/cancel/main.go) | 명시적 취소 | `go run ./examples/17-context/cancel` |
| [deadlines](../../examples/17-context/deadlines/main.go) | 마감과 제한 시간 | `go run ./examples/17-context/deadlines` |
| [hierarchy](../../examples/17-context/hierarchy/main.go) | 부모·자식·형제 | `go run ./examples/17-context/hierarchy` |
| [checkpoints](../../examples/17-context/checkpoints/main.go) | 작업 사이의 취소 확인 | `go run ./examples/17-context/checkpoints` |
| [values](../../examples/17-context/values/main.go) | 요청 범위 메타데이터 | `go run ./examples/17-context/values` |
| [cancel-cause](../../examples/17-context/cancel-cause/main.go) | 취소의 구체적인 원인 | `go run ./examples/17-context/cancel-cause` |

`Context`를 받는 함수에는 nil을 전달하지 않습니다.
시작점에서는 `context.Background()`를,
아직 연결할 호출 관계를 정하지 못한 코드에는 임시로 `context.TODO()`를 사용할 수 있습니다.
[context 패키지의 사용 규칙](https://pkg.go.dev/context#pkg-overview)을 참고합니다.

## 1. 취소 요청과 작업 종료는 다르다

### 1.1 무엇을 전달하는가?

Context는 다음 네 메서드로 정보를 제공합니다.

| 메서드 | 의미 | 반환값을 읽는 방식 |
| --- | --- | --- |
| `Deadline()` | 더 이상 계속하면 안 되는 마감 | 시간과 존재 여부 |
| `Done()` | 취소 시 닫히는 채널 | 취소 신호를 기다리는 데 사용 |
| `Err()` | 취소 상태의 이유 | nil, Canceled, DeadlineExceeded |
| `Value(key)` | 요청에 연결된 값 | 존재와 타입을 확인해 사용 |

[Context 인터페이스](https://pkg.go.dev/context#Context)가 이 계약을 정의합니다.
작업의 결과나 완료 개수는 이 인터페이스에 없습니다.
결과는 함수의 반환값이나 별도의 결과 통로로 전달해야 합니다.

### 1.2 Context가 대신하지 않는 일

취소를 요청해도 이미 실행 중인 모든 코드가 자동으로 반환하지는 않습니다.
작업이 Context를 확인하거나 취소를 지원하는 API에 전달해야 합니다.
이를 **협력적 취소(cooperative cancellation)**라고 생각할 수 있습니다.

| 요구 | Context만으로 자동 보장되는가? |
| --- | --- |
| 취소 상태를 전달 | 가능 |
| 자식에게 취소 전달 | 파생 관계가 유지되면 가능 |
| 임의의 함수 실행을 강제로 중단 | 아니오 |
| 파일 쓰기를 되돌림 | 아니오 |
| 모든 작업이 끝날 때까지 기다림 | 아니오 |
| 결과 배열의 정합성 정책 결정 | 아니오 |

취소 함수의 반환은 “모든 일을 정리했다”는 확인서가 아닙니다.
[CancelFunc](https://pkg.go.dev/context#CancelFunc)는 작업 종료를 기다리지 않는다고 명시합니다.
18장 이후에는 작업 완료 신호를 따로 기다리는 구조를 함께 만듭니다.

### 1.3 함수 시그니처에 드러내기

Context를 사용하는 함수는 보통 첫 인자를 `ctx context.Context`로 둡니다.
필요한 호출마다 명시적으로 전달하면 호출의 수명을 따라가기 쉽습니다.

다음은 함수 형태를 설명하는 발췌입니다.
구현 없이 그대로 실행하는 코드는 아닙니다.

```go
func loadReport(ctx context.Context, name string) (Report, error)
```

Context를 오래 살아 있는 전역 변수나 서비스 구조체에 묶어 두면
어느 요청의 취소·마감인지 혼동하기 쉽습니다.
[공식 Context와 구조체 설명](https://go.dev/blog/context-and-structs)은
호출마다 인자로 전달하는 이유를 자세히 다룹니다.

이 규칙은 구조체가 Context를 담으면 언어 차원에서 컴파일되지 않는다는 뜻이 아닙니다.
수명과 소유권을 명확히 하기 위한 API 설계 지침입니다.

## 2. `Background`에서 시작하기

`context.Background()`는 비어 있는 시작 Context를 반환합니다.
취소되지 않고 마감도 없으며 값도 없습니다.
`Done()`은 nil입니다.

소스: [background/main.go](../../examples/17-context/background/main.go)

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	_, hasDeadline := ctx.Deadline()
	fmt.Println("마감 있음:", hasDeadline)
	fmt.Println("취소됨:", ctx.Err() != nil)
	fmt.Println("Done 없음:", ctx.Done() == nil)
}
```

실행:

```sh
go run ./examples/17-context/background
```

예상 결과:

```text
마감 있음: false
취소됨: false
Done 없음: true
```

### 2.1 `Deadline`의 bool

`Deadline()`은 시간과 bool을 함께 반환합니다.
bool이 false이면 반환된 시간을 실제 마감으로 사용하지 않습니다.
제로 시각을 보고 “이미 마감됐다”고 판단하면 안 됩니다.

`Err() == nil`은 현재 취소 상태가 아니라는 뜻입니다.
작업이 성공적으로 끝났다는 뜻은 아닙니다.
아직 시작하지 않은 Background도 Err가 nil입니다.

### 2.2 nil Done을 기다리지 않기

`Done()`의 반환 타입은 `<-chan struct{}`입니다.
이번에는 “취소 신호를 받기만 할 수 있는 채널”로 이해하면 됩니다.
19장에서 채널의 송신·수신·닫기를 자세히 다룹니다.

nil 채널에서 단독 수신을 시도하면 계속 기다립니다.
그래서 Background의 Done에 무조건 `<-ctx.Done()`을 실행하지 않습니다.
이 장의 단독 수신 예제는 취소를 먼저 요청했거나 이미 지난 마감을 사용합니다.

`context.TODO()`도 빈 Context이지만 코드의 의도가 다릅니다.
Background는 의도적인 시작점,
TODO는 아직 적절한 Context 연결을 정하지 못했다는 표시로 사용합니다.
[Background](https://pkg.go.dev/context#Background)와
[TODO](https://pkg.go.dev/context#TODO)를 참고합니다.

## 3. 명시적으로 취소하기

### 3.1 Context와 취소 함수를 함께 받는다

`WithCancel(parent)`는 자식 Context와 취소 함수를 반환합니다.
자식은 작업에 전달하고, 취소 함수는 그 작업의 수명을 소유한 쪽이 보관합니다.

소스: [cancel/main.go](../../examples/17-context/cancel/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Println("취소 전:", ctx.Err() == nil)
	cancel()
	cancel()
	<-ctx.Done()
	fmt.Println("취소 확인:", errors.Is(ctx.Err(), context.Canceled))
	fmt.Println("부모 취소:", context.Background().Err() != nil)
}
```

실행:

```sh
go run ./examples/17-context/cancel
```

예상 결과:

```text
취소 전: true
취소 확인: true
부모 취소: false
```

### 3.2 코드의 순서

| 순서 | 동작 | 상태 |
| --- | --- | --- |
| 1 | WithCancel | 자식 Err는 nil |
| 2 | defer cancel | 함수 종료 때 정리 예약 |
| 3 | cancel 호출 | 자식 취소 요청 |
| 4 | cancel 다시 호출 | 이미 취소한 상태 유지 |
| 5 | Done 수신 | 취소 신호 관찰 |
| 6 | Err 검사 | Canceled 확인 |

`<-ctx.Done()`은 채널에서 받는 표현입니다.
취소로 닫힌 채널에서는 수신이 진행할 수 있습니다.
우리는 값을 가져오려는 것이 아니라 취소 신호를 확인하려는 것입니다.

Context의 Done을 사용자가 직접 닫지 않습니다.
이 채널은 수신 전용이고, 닫는 책임은 Context 구현에 있습니다.
작업을 멈추라는 의도는 반환받은 cancel 함수로 전달합니다.

### 3.3 cancel을 여러 번 호출해도 되는 이유

CancelFunc는 여러 번 호출할 수 있으며 첫 호출 이후의 호출은 상태를 되돌리지 않습니다.
[CancelFunc 문서](https://pkg.go.dev/context#CancelFunc)의 계약입니다.
그래서 중간에 취소한 경로와 함수 끝의 defer가 함께 있어도 사용할 수 있습니다.

이 성질이 cancel을 생략해도 된다는 뜻은 아닙니다.
정상 종료 경로에서도 cancel을 호출해 연결된 자원을 정리해야 합니다.
WithTimeout을 쓰면 관련 타이머의 수명도 포함됩니다.

### 3.4 취소된 Context는 다시 켜지지 않는다

한 번 취소된 Context의 Err는 다시 nil로 돌아가지 않습니다.
새 작업을 시작할 때는 적절한 부모에서 새로운 자식을 만듭니다.
이전 작업의 취소된 Context를 재사용하고 “재시도니까 다시 살아나겠지”라고 기대하지 않습니다.

반대로 부모가 이미 취소되었다면 그 아래에서 새 자식을 만들어도 취소를 피할 수 없습니다.
취소를 숨기기 위해 무조건 Background로 바꾸면 호출자의 계약을 끊습니다.
부모·자식 관계는 5절에서 확인합니다.

## 4. 마감 시각과 제한 시간

### 4.1 절대 마감과 상대 기간

| 함수 | 입력 | 의미 |
| --- | --- | --- |
| `WithDeadline` | `time.Time` | 그 시각까지 |
| `WithTimeout` | `time.Duration` | 생성한 때부터 그 기간까지 |

WithTimeout은 현재 시각에 기간을 더한 마감을 만드는 형태입니다.
[WithTimeout](https://pkg.go.dev/context#WithTimeout)의 계약을 참고합니다.
Duration의 단위는 15장의 내용을 사용합니다.

소스: [deadlines/main.go](../../examples/17-context/deadlines/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	<-ctx.Done()
	deadline, exists := ctx.Deadline()
	fmt.Println("마감 있음:", exists)
	fmt.Println("마감:", deadline.UTC().Format(time.RFC3339))
	fmt.Println("마감 초과:", errors.Is(ctx.Err(), context.DeadlineExceeded))

	immediate, stop := context.WithTimeout(context.Background(), 0)
	defer stop()
	<-immediate.Done()
	fmt.Println("0초 제한:", errors.Is(immediate.Err(), context.DeadlineExceeded))
}
```

실행:

```sh
go run ./examples/17-context/deadlines
```

예상 결과:

```text
마감 있음: true
마감: 1970-01-01T00:00:00Z
마감 초과: true
0초 제한: true
```

### 4.2 왜 지난 시각을 쓰는가?

특정 밀리초만큼 잠든 뒤 “이제 분명히 마감됐을 것”이라고 검사하면
실행 환경의 스케줄링과 부하에 따라 테스트가 느리거나 불안정해질 수 있습니다.
이미 지난 마감은 시작 전 마감 상태를 직접 만들 수 있습니다.

0초 제한도 같은 목적입니다.
예제는 Context 상태를 확인하고 끝나며 오래 기다리지 않습니다.
마감이 지나도 cancel을 호출할 책임은 그대로 유지합니다.

### 4.3 취소와 마감 초과의 구분

| 상황 | `Err()` |
| --- | --- |
| 아직 취소되지 않음 | nil |
| 명시적 취소 또는 부모의 일반 취소 | context.Canceled |
| 마감이 지나 취소됨 | context.DeadlineExceeded |

실패 원인을 문자열 전체로 비교하기보다 `errors.Is`로 분류합니다.
하위 함수가 `%w`로 감싸도 원인을 확인할 수 있습니다.
[Canceled와 DeadlineExceeded](https://pkg.go.dev/context#pkg-variables)를 참고합니다.

시간 제한을 가진 Context를 수동으로 먼저 cancel하면
그 Context의 상태는 일반 취소로 끝날 수 있습니다.
“타임아웃 Context에서 나온 모든 오류는 DeadlineExceeded”라고 가정하지 않습니다.

### 4.4 부모의 마감을 늦출 수 없다

부모에게 1시간 마감이 있는데 자식에게 24시간을 요청해도
자식의 작업 수명이 부모보다 길어지지는 않습니다.
두 조건 중 더 이른 마감이 적용됩니다.

다음 전체 프로그램은 실제 시각을 출력하지 않고
자식의 마감이 부모의 마감과 같은지만 확인합니다.

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	parentDeadline := time.Now().Add(time.Hour)
	parent, cancelParent := context.WithDeadline(context.Background(), parentDeadline)
	defer cancelParent()
	child, cancelChild := context.WithTimeout(parent, 24*time.Hour)
	defer cancelChild()
	childDeadline, exists := child.Deadline()
	fmt.Println("자식 마감 있음:", exists)
	fmt.Println("부모 마감 유지:", childDeadline.Equal(parentDeadline))
}
```

예상 결과:

```text
자식 마감 있음: true
부모 마감 유지: true
```

[WithDeadline](https://pkg.go.dev/context#WithDeadline)은
부모의 기존 마감보다 늦어지지 않는다는 조건을 설명합니다.
요청 전체의 시간 예산을 하위 호출마다 새 Background로 초기화하지 않습니다.

## 5. 부모·자식·형제의 취소 방향

### 5.1 취소는 아래로 전달된다

부모 작업이 더 이상 필요 없으면 그 작업을 위해 만든 자식들도 멈춰야 합니다.
반대로 자식 하나가 끝났다는 이유로 부모의 다른 일을 모두 취소할 필요는 없습니다.

소스: [hierarchy/main.go](../../examples/17-context/hierarchy/main.go)

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	child, cancelChild := context.WithCancel(parent)
	defer cancelChild()
	sibling, cancelSibling := context.WithCancel(parent)
	defer cancelSibling()

	cancelChild()
	<-child.Done()
	fmt.Println("자식 취소:", child.Err() != nil)
	fmt.Println("부모 유지:", parent.Err() == nil)
	fmt.Println("형제 유지:", sibling.Err() == nil)

	cancelParent()
	<-sibling.Done()
	fmt.Println("부모 취소 후 형제:", sibling.Err() != nil)
}
```

실행:

```sh
go run ./examples/17-context/hierarchy
```

예상 결과:

```text
자식 취소: true
부모 유지: true
형제 유지: true
부모 취소 후 형제: true
```

### 5.2 상태표

| 시점 | parent | child | sibling |
| --- | --- | --- | --- |
| 생성 직후 | 유지 | 유지 | 유지 |
| cancelChild 뒤 | 유지 | 취소 | 유지 |
| cancelParent 뒤 | 취소 | 취소 유지 | 취소 |

자식 취소는 형제에게 전달되지 않습니다.
공통 부모를 취소하면 두 자식에 영향을 줍니다.
어느 cancel 함수를 호출하는지가 수명 경계를 결정합니다.

이미 취소된 자식을 다시 취소해도 상태는 변하지 않습니다.
취소 함수마다 defer를 둔 이유는 어떤 중간 경로로 끝나도
그 함수가 만든 자식의 정리를 빠뜨리지 않기 위해서입니다.

### 5.3 cancel 함수의 소유자

일반적인 처리 함수는 Context를 받아도 부모의 cancel 함수는 받지 않습니다.
자기 작업을 중단할 필요가 있으면 오류를 반환해 호출자에게 결정하게 할 수 있습니다.
자식 범위를 만들었다면 그 자식의 취소 함수는 직접 정리합니다.

취소 권한과 상태 관찰을 나누면
한 내부 함수가 관련 없는 형제 작업까지 임의로 끊는 일을 줄일 수 있습니다.
[공식 Context 소개](https://go.dev/blog/context)는 이 방향을 설명합니다.

## 6. 취소 확인 지점을 두기

### 6.1 반복문의 각 작업 전에 확인하기

Context를 매개변수에 추가하기만 해서는 동작이 바뀌지 않습니다.
작업이 중단할 수 있는 지점에서 Err를 확인하고 반환해야 합니다.

이번 함수는 각 항목 처리 전에 취소를 확인합니다.
성공한 항목 개수는 오류와 함께 돌려줄 수 있습니다.
이는 이 함수가 정한 **부분 진행 정보**의 계약입니다.

소스: [checkpoints/main.go](../../examples/17-context/checkpoints/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

func process(ctx context.Context, items []string, step func(string) error) (int, error) {
	completed := 0
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return completed, err
		}
		if err := step(item); err != nil {
			return completed, fmt.Errorf("항목 %q: %w", item, err)
		}
		completed++
	}
	return completed, ctx.Err()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	completed, err := process(ctx, []string{"Go", "Test", "Code"}, func(item string) error {
		if item == "Go" {
			cancel()
		}
		return nil
	})
	fmt.Println("완료 항목:", completed)
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
}
```

실행:

```sh
go run ./examples/17-context/checkpoints
```

예상 결과:

```text
완료 항목: 1
취소 확인: true
```

### 6.2 한 항목을 처리한 뒤 무슨 일이 일어나는가?

| 단계 | item | completed | 상태 |
| --- | --- | --- | --- |
| 첫 반복 전 확인 | Go | 0 | 아직 취소 아님 |
| step 안에서 cancel | Go | 0 | 취소 요청 |
| step은 nil 반환 | Go | 1 | 이 항목은 성공으로 셈 |
| 다음 반복 전 확인 | Test | 1 | 취소 오류 반환 |
| Code | 실행하지 않음 | 1 | 호출 없음 |

함수의 step 인자는 nil이 아닌 함수라는 전제입니다.
step이 오류를 반환하면 그 항목은 completed에 포함하지 않습니다.
예제의 step은 동기적으로 호출되며 별도 고루틴을 만들지 않습니다.

마지막 항목이 끝난 직후의 취소도 확인하도록
루프 끝에서 `ctx.Err()`를 다시 반환합니다.
그래서 세 번째 항목 안에서 취소하면 completed는 3이어도 오류는 Canceled입니다.
“전체 항목 수만 채웠으니 반드시 성공”이라는 정책을 사용하지 않았습니다.

### 6.3 확인 사이의 작업은 자동으로 중단되지 않는다

step이 내부에서 오래 기다리고 Context를 받지 않는다면
다음 반복의 확인 지점까지 돌아오지 못할 수 있습니다.
바깥의 Err 검사만으로 그 기다림을 강제로 중단할 수는 없습니다.

실제 하위 API가 Context를 받는다면 같은 ctx를 전달합니다.
함수가 긴 계산을 수행한다면 내부 반복에도 적절한 확인 지점을 둡니다.
취소를 얼마나 자주 확인할지는 작업 단위와 비용에 맞춰 정합니다.

한 번의 검사와 함수 반환 사이에도 다른 곳에서 취소가 발생할 수 있습니다.
Context는 모든 부수 효과와 취소를 하나의 원자적 연산으로 묶어 주지 않습니다.
실패 시 결과를 반환할지, 이미 쓴 데이터를 어떻게 취급할지는 별도 계약입니다.

### 6.4 부분 결과와 전체 성공 결과

| 함수 설계 | 실패 시 결과 | 적합한 설명 |
| --- | --- | --- |
| 진행 집계 | 완료 개수와 오류 | 어디까지 진행했는지 관찰 |
| 전체 검증 변환 | nil 결과와 오류 | 전체 성공일 때만 결과 사용 |
| 외부 파일·서비스 변경 | 별도 부수 효과 존재 | 롤백·재시도 정책 필요 |

본문의 process는 진행 집계입니다.
연습문제 6의 applyAll은 실패하면 결과 슬라이스를 공개하지 않습니다.
둘 다 이미 수행한 외부 부수 효과를 자동으로 되돌리는 것은 아닙니다.

## 7. 요청 범위의 값 전달

### 7.1 Context 값은 일반 옵션 목록이 아니다

요청 ID처럼 여러 API를 통과하며 따라가야 하는 작은 메타데이터가 있습니다.
이런 값은 Context의 Value로 전달할 수 있습니다.
함수의 필수 입력이나 선택 옵션 전체를 Context에 숨기지는 않습니다.

`limit`, `filename`, `minutes`처럼 함수의 동작을 결정하는 명확한 인자는
보통 함수 매개변수나 설정 구조체로 드러내는 편이 낫습니다.
[WithValue 문서](https://pkg.go.dev/context#WithValue)의 권고를 참고합니다.

### 7.2 키는 자체 타입으로 정의하기

여러 패키지가 `"id"`라는 문자열 키를 사용하면 서로 충돌할 수 있습니다.
이번에는 비공개 이름의 자체 타입 `requestIDKey`를 정의합니다.
비교 가능한 키여야 하므로 슬라이스나 맵은 키로 쓰지 않습니다.

소스: [values/main.go](../../examples/17-context/values/main.go)

```go
package main

import (
	"context"
	"fmt"
)

type requestIDKey struct{}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func requestID(ctx context.Context) (string, bool) {
	id, exists := ctx.Value(requestIDKey{}).(string)
	return id, exists
}

func main() {
	parent := withRequestID(context.Background(), "req-001")
	child, cancel := context.WithCancel(parent)
	defer cancel()
	id, exists := requestID(child)
	fmt.Println("요청 ID:", id)
	fmt.Println("값 있음:", exists)
	cancel()
	<-child.Done()
	id, exists = requestID(child)
	fmt.Println("취소 후 값:", id, exists)
}
```

실행:

```sh
go run ./examples/17-context/values
```

예상 결과:

```text
요청 ID: req-001
값 있음: true
취소 후 값: req-001 true
```

### 7.3 타입 단언과 존재 여부

`Value`의 반환 타입은 any입니다.
`.(string)`의 두 반환값 형태를 사용해 기대한 문자열이 있는지 확인합니다.
다른 타입이나 값이 없을 때 패닉을 일으키는 단일 반환값 단언을 사용하지 않습니다.

이 기본 helper는 빈 문자열 저장도 허용합니다.
빈 문자열이 있어도 타입이 문자열이면 exists는 true입니다.
연습문제 5에서는 ID를 검증해 빈 문자열을 거부하는 정책을 추가합니다.

### 7.4 파생 값과 원래 Context

WithValue는 원래 Context를 직접 바꾸지 않고 파생 Context를 만듭니다.
자식에 같은 키를 새 값으로 저장하면 그 자식에서 새 값이 보입니다.
부모를 읽으면 원래 값이 유지됩니다.

취소와 마감은 부모로부터 이어집니다.
값을 붙였다고 부모 취소가 끊기지는 않습니다.
취소되었다고 Value의 내용이 자동으로 지워지는 것도 아닙니다.

### 7.5 Context가 안전해도 내부 값까지 보호되는 것은 아니다

Context의 메서드는 여러 고루틴에서 함께 사용할 수 있습니다.
하지만 Context에 넣은 맵이나 포인터의 내부 데이터를
여러 곳에서 변경해도 자동으로 보호되는 것은 아닙니다.

이번 예제는 변경하지 않는 문자열 ID를 사용합니다.
공유 데이터의 동기화는 20장에서 다룹니다.
Context를 잠금이나 데이터 복사 기능으로 설명하지 않습니다.

## 8. 취소 상태와 구체적인 원인

### 8.1 `Err`와 `Cause`

일반 취소의 상태는 `context.Canceled`입니다.
하지만 사용자가 중단했는지, 상위 단계의 오류로 중단했는지 등의
더 구체적인 이유를 남기고 싶을 수 있습니다.

`WithCancelCause`는 오류를 받는 취소 함수를 돌려줍니다.
상태 분류는 Err에서, 기록한 이유는 `context.Cause`에서 확인합니다.
이 API는 Go 1.20부터 제공되며 저장소의 Go 1.27 기준에서 사용할 수 있습니다.
[WithCancelCause](https://pkg.go.dev/context#WithCancelCause)를 참고합니다.

소스: [cancel-cause/main.go](../../examples/17-context/cancel-cause/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	reason := errors.New("사용자가 작업 중단")
	cancel(reason)
	<-ctx.Done()
	fmt.Println("상태:", ctx.Err())
	fmt.Println("원인:", context.Cause(ctx))
	cancel(errors.New("나중의 원인"))
	fmt.Println("첫 원인 유지:", errors.Is(context.Cause(ctx), reason))
}
```

실행:

```sh
go run ./examples/17-context/cancel-cause
```

예상 결과:

```text
상태: context canceled
원인: 사용자가 작업 중단
첫 원인 유지: true
```

### 8.2 첫 취소의 원인이 유지된다

이미 취소된 Context에 다른 이유로 cancel을 다시 호출해도
처음 정해진 원인을 덮어쓰지 않습니다.
취소 함수의 반복 호출이 안전하다는 성질과 연결됩니다.

`cancel(nil)`은 “취소를 하지 않음”이 아닙니다.
일반 취소 원인인 Canceled로 취소합니다.
그래서 정상 종료 경로의 정리용 defer에도 사용할 수 있습니다.

| 상태 | Err | Cause |
| --- | --- | --- |
| 아직 취소 아님 | nil | nil |
| cancel(사용자 오류) | Canceled | 사용자 오류 |
| cancel(nil) | Canceled | Canceled |
| 일반 WithCancel 취소 | Canceled | Canceled |

부모가 먼저 취소되면 자식도 그 원인을 물려받을 수 있습니다.
자식이 별도의 원인으로 먼저 취소된 경우에는 그 자식의 원인이 유지됩니다.
[CancelCauseFunc](https://pkg.go.dev/context#CancelCauseFunc)의 순서 규칙을 참고합니다.

### 8.3 오류 정책에 두 정보를 사용하기

호출자는 `errors.Is(err, context.Canceled)`로 작업 취소를 분류할 수 있습니다.
구체적인 진단이 필요하면 해당 Context의 Cause를 별도로 확인합니다.
함수의 반환 오류를 임의로 Cause로 바꾸면 기존 분류 계약이 바뀔 수 있습니다.

마감에 자체 원인을 붙이는 `WithTimeoutCause`와 `WithDeadlineCause`도 있습니다.
이 함수들의 일반 CancelFunc를 직접 호출하는 경우와
마감이 실제로 지나서 취소되는 경우의 원인 정책은 다릅니다.
사용할 때는 [WithTimeoutCause](https://pkg.go.dev/context#WithTimeoutCause)를 확인합니다.

## 9. 호출 수명과 정리 설계

### 9.1 만든 범위에서 정리하기

다음은 호출 범위를 보여 주는 함수 발췌입니다.
work 함수가 정의된 프로그램 안에서 사용하는 형태입니다.

```go
func limited(ctx context.Context, work func(context.Context) error) error {
	child, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return work(child)
}
```

defer는 성공·실패 모두에서 자식의 cancel을 호출합니다.
work가 1초보다 빨리 끝나면 남은 타이머 수명을 정리할 수 있습니다.
부모를 취소하지 않고 이번 하위 작업의 수명만 끝냅니다.

이 발췌는 work 자체가 Context를 존중한다는 전제입니다.
1초가 지나면 work의 실행을 강제로 중단하는 래퍼가 아닙니다.
연습문제 4에서 이런 전제를 함수 계약으로 적고 테스트합니다.

### 9.2 루프에서 defer를 계속 쌓지 않기

루프 안에서 만든 cancel을 바깥 함수의 defer로만 모으면
각 반복 작업이 끝나도 정리가 함수 전체의 끝까지 늦어집니다.
오래 실행되는 함수에서는 자식과 타이머가 불필요하게 유지될 수 있습니다.

각 반복을 작은 함수로 나누어 그 안에서 defer하거나,
반복 작업이 끝난 직후 cancel을 호출하는 방식으로 수명을 맞춥니다.
오류 분기에서도 cancel을 빠뜨리지 않도록 주의합니다.

### 9.3 취소를 숨기지 않기

하위 함수가 받은 ctx 대신 Background를 사용하면
부모의 마감·취소·값 연결이 끊깁니다.
호출자가 작업을 중단해도 하위 작업은 그 사실을 모르고 계속할 수 있습니다.

`WithoutCancel`은 의도적으로 취소 연결을 끊는 별도 API입니다.
이것은 일반적인 오류 수정 방법이 아닙니다.
사용한다면 새 수명, 종료, 자원 정리 책임을 다시 설계해야 합니다.
[WithoutCancel](https://pkg.go.dev/context#WithoutCancel)을 참고합니다.

### 9.4 반환 뒤에 사용할 자식은 누가 소유하는가?

함수가 자식 Context를 만들고 defer cancel한 뒤 그 자식만 반환하면
호출자는 이미 취소된 자식을 받게 됩니다.
반환 뒤까지 수명이 이어져야 하는 API는 소유권을 별도로 설계해야 합니다.

작업을 함수 안에서 끝낸다면 defer cancel이 잘 맞습니다.
작업의 수명을 호출자에게 넘기려면 정리 함수도 함께 전달하는 등의
명확한 계약이 필요합니다.

예제는 수명이 함수 안에서 끝나도록 작성합니다.
추가 고루틴의 수명은 다음 장에서 완료를 기다리는 방법과 함께 다룹니다.

## 10. 개인 실습: 단계별 취소를 관찰하기

### 10.1 실습의 목적

세 단계를 순서대로 실행하고 지정한 단계 뒤에 취소합니다.
실제 파일 읽기·저장을 수행하는 프로그램이 아니라
취소 지점과 진행 정보 반환을 관찰하는 모의 작업입니다.

단계 함수가 nil을 반환하면 해당 단계는 완료 개수에 포함합니다.
단계 함수가 오류를 반환하면 그 단계는 포함하지 않습니다.
모든 단계 뒤에도 Context 상태를 마지막으로 확인합니다.

### 10.2 개인 파일 준비

Bash:

```sh
mkdir -p _practice/17-context
if [ ! -e _practice/17-context/main.go ]; then
  cp examples/17-context/checkpoints/main.go _practice/17-context/main.go
fi
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/17-context | Out-Null
if (-not (Test-Path _practice/17-context/main.go)) {
    Copy-Item examples/17-context/checkpoints/main.go _practice/17-context/main.go
}
```

기존 개인 파일은 덮어쓰지 않습니다.
개인 `main.go` 전체를 다음 코드로 바꿉니다.

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

func performSteps(ctx context.Context, names []string, step func(context.Context, string) error) (int, error) {
	if step == nil {
		return 0, errors.New("단계 함수가 필요합니다")
	}
	completed := 0
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return completed, err
		}
		if err := step(ctx, name); err != nil {
			return completed, fmt.Errorf("단계 %q: %w", name, err)
		}
		completed++
	}
	return completed, ctx.Err()
}

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	cancelAfter := 2
	calls := 0
	reason := errors.New("실습 중단")
	completed, err := performSteps(ctx, []string{"읽기", "변환", "저장"}, func(ctx context.Context, name string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		calls++
		if calls == cancelAfter {
			cancel(reason)
		}
		return nil
	})
	fmt.Println("완료 단계:", completed)
	fmt.Println("호출 횟수:", calls)
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
	fmt.Println("원인:", context.Cause(ctx))
}
```

실행:

```sh
gofmt -w _practice/17-context/main.go
go run ./_practice/17-context
```

예상 결과:

```text
완료 단계: 2
호출 횟수: 2
취소 확인: true
원인: 실습 중단
```

### 10.3 취소 위치 바꾸기

`cancelAfter`만 바꾸어 순서대로 확인합니다.

| cancelAfter | 호출 횟수 | 완료 단계 | 취소 확인 |
| --- | --- | --- | --- |
| 0 | 3 | 3 | false |
| 1 | 1 | 1 | true |
| 2 | 2 | 2 | true |
| 3 | 3 | 3 | true |
| 4 | 3 | 3 | false |

0과 4에서는 조건에 맞는 호출이 없어 실행 중 취소하지 않습니다.
이때 출력 시점의 Cause는 nil이며 화면에는 `<nil>`로 보입니다.
main이 끝날 때 defer cancel(nil)이 나중에 실행됩니다.

3에서는 모든 단계가 nil을 반환했지만 마지막 Context 검사에서 취소를 발견합니다.
완료 개수와 전체 성공 상태가 별도 정보임을 보여 줍니다.
이 정책을 원하지 않는 도메인에서는 완료·취소가 겹칠 때의 우선순위를 다르게 정할 수 있습니다.

### 10.4 실행 전 취소와 작업 자체의 오류

다음 변형은 개인 파일에서 한 번에 하나씩 적용합니다.

- `performSteps` 호출 전에 cancel(reason)을 호출하면 단계 함수가 0번 실행됩니다.
- 빈 names를 사용하면 완료 개수는 0이며 살아 있는 Context에서는 오류가 없습니다.
- 두 번째 단계가 일반 오류를 반환하면 첫 단계만 완료 개수에 포함됩니다.
- 일반 작업 오류를 반환했다고 Context 자체가 자동으로 취소되지는 않습니다.

일반 오류를 시험할 때는 `cancelAfter := 0`으로 두고,
콜백의 calls 증가 뒤 두 번째 호출에서 오류를 반환합니다.
완료 1, 호출 2가 되어야 합니다.
이때 `errors.Is(err, context.Canceled)`는 false입니다.

### 10.5 검증 방법의 선택

`time.Sleep`으로 취소가 일어날 때를 추측하지 않습니다.
콜백이 정확히 몇 번째 호출에서 cancel하는지 직접 제어합니다.
따라서 테스트는 작업 횟수와 오류를 확정적으로 검사할 수 있습니다.

실제 동시성 작업에서는 호출 순서를 이렇게 단순하게 제어할 수 없을 수 있습니다.
그때는 “시작됨”, “진행 허용”, “종료됨” 같은 명시적인 신호를 사용합니다.
18~20장에서 그런 동기화 구조를 배웁니다.

## 11. 선택: 취소 가능한 대기

이 절은 19장의 채널·select를 미리 사용하는 선택 내용입니다.
기본 7개 예제를 마친 뒤 건너뛰고 다음 장으로 가도 됩니다.
새 구문을 읽는 데 필요한 최소 의미를 먼저 정리합니다.

| 표현 | 이 절에서의 의미 |
| --- | --- |
| `timer.C` | 타이머가 준비되었음을 알리는 채널 |
| `<-ctx.Done()` | 취소 신호를 받는 분기 |
| `select` | 준비된 채널 동작 중 하나를 선택 |
| nil 채널 분기 | select에서 선택되지 않음 |

default가 없고 모든 채널이 준비되지 않았으면 select는 기다립니다.
여러 분기가 동시에 준비되었다고 위에 적은 분기가 우선하는 것은 아닙니다.
[select 명세](https://go.dev/ref/spec#Select_statements)를 참고합니다.

다음은 별도 개인 폴더에서 실행하는 전체 프로그램입니다.
이미 취소된 Context와 0초 정상 대기를 사용하므로 오래 기다리지 않습니다.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func wait(ctx context.Context, duration time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if duration < 0 {
		return errors.New("대기 시간은 음수일 수 없습니다")
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fmt.Println("취소 대기:", errors.Is(wait(ctx, time.Hour), context.Canceled))
	fmt.Println("0초 대기 성공:", wait(context.Background(), 0) == nil)
}
```

예상 결과:

```text
취소 대기: true
0초 대기 성공: true
```

처음에 Err를 검사하여 이미 취소된 요청이 새 대기를 시작하지 않도록 합니다.
타이머 분기 뒤에도 Err를 확인하는 정책입니다.
그래도 검사와 반환 사이에 발생하는 모든 동시 사건을 원자적으로 묶는 것은 아닙니다.

`NewTimer`는 타이머를 만들고 `Stop`은 더 이상 필요 없는 타이머를 멈추려는 호출입니다.
[time.NewTimer](https://pkg.go.dev/time#NewTimer)와
[Timer.Stop](https://pkg.go.dev/time#Timer.Stop)을 참고합니다.
이 프로그램은 타이머를 재설정하거나 같은 채널을 여러 곳에서 받지 않습니다.

대기를 취소 가능하게 만들었다고 일반 Reader의 Read까지 취소되는 것은 아닙니다.
실제 I/O API가 Context를 받는지, 자체 마감을 지원하는지 확인해야 합니다.
새 고루틴으로 읽기를 감싸고 결과를 무시하는 것만으로는 읽던 작업이 종료되지 않을 수 있습니다.

## 12. 흔한 실수

| 증상 | 확인할 원인 | 수정 방향 |
| --- | --- | --- |
| 취소해도 계산이 계속됨 | ctx를 전달만 하고 확인하지 않음 | 작업 경계에서 Err 검사 |
| 자식의 제한 시간이 너무 길어짐 | 매번 Background에서 시작 | 받은 부모에서 파생 |
| 함수 반환 뒤 자식이 바로 취소됨 | defer cancel 후 자식을 반환 | 소유권·수명 재설계 |
| 타이머·자식 수명이 길어짐 | cancel 누락 또는 루프 밖까지 defer 누적 | 작업 범위에서 정리 |
| Background의 Done에서 멈춤 | nil 채널 단독 수신 | 취소 가능한 Context만 기다리기 |
| 작업 완료로 착각 | Done을 완료 신호로 해석 | 별도 완료 신호·반환 대기 |
| 값 조회에서 패닉 | 단일 타입 단언 또는 nil Context | 존재·타입 검사 |
| Context 키 충돌 | 기본 문자열 키 사용 | 비공개 자체 키 타입 |
| 취소 원인을 덮어쓰려 해도 안 됨 | 첫 취소 원인 유지 | 첫 사건과 후속 오류 구분 |
| 값이 공유되어 race 발생 | Value가 내부 데이터도 보호한다고 생각 | 불변 값 또는 별도 동기화 |

### 12.1 작업 오류와 취소 오류

step이 오류를 반환하면 취소 없이도 작업이 끝날 수 있습니다.
Context의 Err가 nil이어도 작업의 반환 오류는 nil이 아닐 수 있습니다.
두 오류 채널을 같은 것으로 취급하지 않습니다.

취소 시에는 errors.Is로 Canceled나 DeadlineExceeded를 분류합니다.
단순 문자열 비교는 `%w`로 감싼 오류에서 의도를 놓치기 쉽습니다.
일반 작업 오류는 작업 자체의 의미를 보존해 전달합니다.

### 12.2 취소 확인 빈도

매 연산마다 확인하면 비용이 있고,
너무 드물게 확인하면 중단 반응이 늦을 수 있습니다.
이 장은 항목 단위라는 자연스러운 경계를 사용합니다.

대량 계산에서는 적절한 묶음마다 확인하는 선택도 가능합니다.
그때의 반응 시간과 처리량은 측정할 문제입니다.
이 장은 특정 빈도가 모든 프로그램에서 가장 빠르다고 주장하지 않습니다.

### 12.3 검사 도구의 범위

`go vet`는 취소 함수가 사용되지 않는 일부 경로를 찾아낼 수 있습니다.
하지만 모든 작업이 종료하는지, callback이 ctx를 존중하는지,
입력 중단이 안전한지는 자동으로 모두 증명하지 않습니다.

테스트는 취소된 입력·부모 취소·마감·정상 정리 경로를 직접 확인합니다.
후속 동시성 장에서는 race detector와 완료 신호 검사도 함께 사용합니다.

## 13. 연습문제와 검사

[연습문제](../../exercises/17-context/README.md)는 기본 6문제와 선택 1문제입니다.
[해설](../../solutions/17-context/README.md)에는 네 개의 실행 가능한 참고 구현이 있습니다.

| 문제 | 목표 |
| --- | --- |
| 1 | Context 상태와 취소 방향 예측 |
| 2 | 취소를 확인하는 제한된 합계 |
| 3 | Background로 끊긴 호출 연결 복구 |
| 4 | 시간 예산의 소유권과 정리 |
| 5 | 검증된 추적 ID와 파생 Context |
| 6 | 전체 성공일 때만 결과를 공개하는 변환 |
| 선택 | 취소 가능한 대기의 계약 검사 |

장 범위 검사:

```sh
go test ./examples/17-context/... ./solutions/17-context/...
go vet ./examples/17-context/... ./solutions/17-context/...
go build ./examples/17-context/... ./solutions/17-context/...
```

테스트에는 임의의 Sleep이 없습니다.
과거 마감, 0초 예산, 콜백에서 정해진 횟수의 취소로 상태를 재현합니다.
검증을 위한 timeout은 실패한 테스트가 영원히 매달리지 않도록 하는 상한입니다.
정상 동작의 성공 조건으로 그 시간의 경과를 사용하지 않습니다.

## 14. 자기 점검

1. cancel 호출은 모든 작업이 끝날 때까지 기다리는가?
2. Err가 nil이면 작업이 성공했다는 뜻인가?
3. Background의 Done은 어떤 값인가?
4. 자식 취소는 부모와 형제도 취소하는가?
5. 자식의 마감을 부모보다 늦출 수 있는가?
6. 함수가 받은 ctx를 Background로 바꾸면 무엇을 잃는가?
7. WithTimeout을 만들고 작업이 빨리 끝나도 cancel해야 하는가?
8. 취소된 Context를 재시도용으로 초기화할 수 있는가?
9. Value에 저장한 맵이 자동으로 동시성 안전해지는가?
10. Err와 Cause는 각각 어떤 정보를 주는가?

<details>
<summary>자기 점검 해설</summary>

1. 기다리지 않는다. 작업의 반환이나 별도 완료 신호를 관찰해야 한다.
2. 현재 취소되지 않았다는 뜻이며 작업의 반환 오류·완료 여부는 별개다.
3. nil 채널이다. 단독 수신으로 기다리면 진행할 수 없다.
4. 자식과 그 아래만 취소되며 부모·형제는 그대로다.
5. 더 이른 부모 마감을 넘길 수 없다.
6. 부모의 취소·마감·값 연결을 잃는다.
7. 정상 경로에서도 관련 자원을 정리하도록 cancel한다.
8. 이미 취소된 상태를 되돌리지 못하며 적절한 부모에서 새 작업 범위를 만든다.
9. Context 메서드의 안전성과 저장한 데이터의 동기화는 별개다.
10. Err는 취소 상태 분류, Cause는 첫 취소에 기록한 구체적인 이유를 제공한다.

</details>

## 15. 핵심 정리와 다음 장

Context를 사용한다는 것은 시그니처에 타입 하나를 추가하는 데서 끝나지 않습니다.
호출 아래로 연결을 유지하고, 중단할 지점에서 확인하며,
만든 취소 범위의 정리 책임을 지키는 일까지 포함합니다.

취소는 요청이고 완료는 관찰할 결과입니다.
다음 장에서는 고루틴을 만들고 그 결과와 종료를 기다리면서
두 개념을 실제 동시 실행에 연결하겠습니다.

- 이전: [16. 명령줄 프로그램](../16-cli/README.md)
- 다음: [18. 고루틴](../18-goroutines/README.md)
- [전체 목차](../README.md)
- [연습문제](../../exercises/17-context/README.md)
- [해설](../../solutions/17-context/README.md)

## 참고 자료

- [context 패키지](https://pkg.go.dev/context)
- [Context 인터페이스](https://pkg.go.dev/context#Context)
- [CancelFunc](https://pkg.go.dev/context#CancelFunc)
- [WithDeadline](https://pkg.go.dev/context#WithDeadline)
- [WithValue](https://pkg.go.dev/context#WithValue)
- [WithCancelCause](https://pkg.go.dev/context#WithCancelCause)
- [공식 Context 소개](https://go.dev/blog/context)
- [Context와 구조체](https://go.dev/blog/context-and-structs)
- [select 명세](https://go.dev/ref/spec#Select_statements)
- [Timer](https://pkg.go.dev/time#Timer)

확인일: 2026-11-09. 코드·출력·취소·마감·정리는 Go 1.27.3에서 확인했습니다.
