# 18. 고루틴과 완료 대기

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 응용 | 함수·클로저·슬라이스·오류·Context | Go, 터미널, 이 저장소 |

여러 자료를 계산하거나 여러 요청을 처리할 때,
하나가 끝날 때까지 다른 모든 작업을 멈추어 두어야 할까요?
Go에서는 함수를 별도의 실행 흐름인 **고루틴(goroutine)**으로 시작할 수 있습니다.
시작한 함수는 호출한 함수와 독립적으로 진행합니다.

여기서 새로 생기는 문제는 속도보다 **수명과 결과의 소유권**입니다.
호출한 함수가 먼저 반환하면 누가 작업을 기다릴까요?
여러 작업이 같은 결과 칸에 쓰면 무엇이 일어날까요?
취소했다는 사실은 모든 고루틴이 끝났다는 뜻일까요?
이 장은 이 질문에 답하는 작은 프로그램을 만듭니다.

## 학습 목표

- 일반 함수 호출과 `go` 문이 만드는 실행 흐름을 구분한다.
- 시작한 작업의 수를 등록하고 모든 작업의 완료를 기다린다.
- 인자 평가, 클로저, 공유 저장 공간의 차이를 설명한다.
- 입력 순서대로 결과를 모으고 여러 오류를 전달한다.
- 취소 요청 후 실제 종료를 확인하고 고루틴 수에 상한을 둔다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
기본 예제에는 외부 네트워크·계정·파일이 필요하지 않습니다.
일정 시간 동안 잠들어 우연히 끝나기를 기대하는 코드도 사용하지 않습니다.

[17장](../17-context/README.md)의 Context는 취소 의사를 전달했습니다.
이번 장의 `sync.WaitGroup`은 작업이 끝났음을 확인합니다.
둘은 서로 다른 역할이므로 함께 사용할 수 있습니다.

채널은 다음 장의 중심 주제입니다.
이번 장의 취소 실습에서만 **닫힌 채널을 신호로 받는 최소 구문**을 먼저 사용합니다.
신호가 필요한 이유와 해당 두 줄의 의미를 그 자리에서 설명합니다.

| 예제 | 학습 내용 | 실행 |
| --- | --- | --- |
| [launch-wait](../../examples/18-goroutines/launch-wait/main.go) | 시작·등록·완료·대기 | `go run ./examples/18-goroutines/launch-wait` |
| [arguments](../../examples/18-goroutines/arguments/main.go) | 호출 인자의 평가 시점 | `go run ./examples/18-goroutines/arguments` |
| [indexed-results](../../examples/18-goroutines/indexed-results/main.go) | 결과 칸의 독립된 소유권 | `go run ./examples/18-goroutines/indexed-results` |
| [early-return](../../examples/18-goroutines/early-return/main.go) | 조기 반환과 defer | `go run ./examples/18-goroutines/early-return` |
| [error-results](../../examples/18-goroutines/error-results/main.go) | 오류를 결과로 모으기 | `go run ./examples/18-goroutines/error-results` |
| [cancel-join](../../examples/18-goroutines/cancel-join/main.go) | 취소와 완료 확인 | `go run ./examples/18-goroutines/cancel-join` |
| [waitgroup-go](../../examples/18-goroutines/waitgroup-go/main.go) | Go 1.25부터의 Go 메서드 | `go run ./examples/18-goroutines/waitgroup-go` |

## 1. 동시성은 실행 순서를 설계하는 문제다

### 1.1 순차 실행과 동시 실행

일반 함수 호출은 호출된 함수가 반환할 때까지 기다립니다.
`go work()`는 새 고루틴에서 함수를 실행하도록 시작한 뒤 호출 지점의 다음 문장으로 진행합니다.
새 작업이 먼저 실행될지 다음 문장이 먼저 실행될지는 보장되지 않습니다.

다음은 차이를 설명하기 위한 발췌입니다.
`work`의 구현이 생략되었으므로 그대로 실행하는 프로그램은 아닙니다.

```go
work()    // work가 반환한 뒤 다음 문장으로 진행
// 다음 문장
```

```go
go work() // work의 반환을 기다리지 않고 다음 문장으로 진행
// 다음 문장
```

**동시성(concurrency)**은 여러 작업의 진행을 구조화하는 방법입니다.
**병렬성(parallelism)**은 여러 작업이 실제로 같은 시각에 계산되는 성질입니다.
고루틴이 둘이라고 반드시 CPU 코어 두 개를 동시에 쓰는 것은 아닙니다.
하나의 실행 자원에서도 여러 작업을 번갈아 진행할 수 있습니다.

어떤 작은 계산은 고루틴을 시작하고 합류하는 비용 때문에 더 느릴 수 있습니다.
이 장의 예제는 실행 구조를 가르치기 위한 것이며 속도 개선을 주장하지 않습니다.
실제로 비교하는 방법은 [26장](../26-performance/README.md)에서 다룹니다.

### 1.2 고루틴과 운영체제 스레드

고루틴은 Go 런타임이 관리하는 실행 단위입니다.
운영체제 스레드 하나와 고루틴 하나가 항상 일대일로 대응한다고 생각하지 않습니다.
고루틴 수를 늘리는 것과 CPU 자원을 늘리는 것도 같은 일이 아닙니다.

이 구분이 필요한 이유는 자원이 무한하지 않기 때문입니다.
작업마다 큰 데이터를 붙여 수십만 개의 고루틴을 만들면
대기 중인 고루틴도 메모리와 관리 비용을 사용합니다.

지금은 입력 개수를 16개 이하로 제한하는 작은 실습으로 시작합니다.
많은 입력은 고정된 수의 작업자가 여러 입력을 나누어 맡는 구조로 바꿉니다.
이 장의 합계 연습과 [31장](../31-concurrent-batch/README.md)이 그 연결점입니다.

### 1.3 누가 작업의 수명을 소유하는가?

함수가 새 고루틴을 만들면 다음 중 하나를 분명히 해야 합니다.

| 설계 | 호출자가 받는 약속 |
| --- | --- |
| 함수가 모든 작업을 기다린 뒤 반환 | 반환 후 내부 고루틴이 남지 않음 |
| 함수가 작업 핸들을 반환 | 호출자가 정해진 메서드로 취소하고 기다림 |
| 서비스의 장기 작업으로 등록 | 서비스의 종료 절차가 모두 기다림 |

이 장의 함수들은 첫 번째 방식을 사용합니다.
호출자가 완료된 결과를 받으면 내부 작업도 끝나 있습니다.
함수 안에서 시작만 하고 잊어버리는 방식은 사용하지 않습니다.

`main` 함수가 끝나면 프로그램은 다른 고루틴이 끝나기를 자동으로 기다리지 않습니다.
따라서 “작업을 시작했다”와 “프로그램을 종료해도 된다” 사이에 완료 확인이 필요합니다.
[프로그램 실행](https://go.dev/ref/spec#Program_execution)과
[go 문](https://go.dev/ref/spec#Go_statements)의 규칙을 참고합니다.

## 2. 한 작업을 시작하고 기다리기

`WaitGroup`은 끝나지 않은 작업의 수를 추적합니다.
영값을 바로 사용할 수 있으므로 `var group sync.WaitGroup`으로 선언합니다.
먼저 전통적인 `Add`·`Done` 방식을 배운 뒤 마지막에 `Go` 메서드를 비교합니다.

소스: [launch-wait/main.go](../../examples/18-goroutines/launch-wait/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var result string
	var group sync.WaitGroup
	group.Add(1)
	go func() {
		defer group.Done()
		result = "고루틴 작업 완료"
	}()
	group.Wait()
	fmt.Println(result)
	fmt.Println("main 종료")
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/18-goroutines/launch-wait
```

예상 출력:

```text
고루틴 작업 완료
main 종료
```

### 2.1 네 동작을 순서대로 읽기

| 순서 | 코드 | 역할 |
| --- | --- | --- |
| 1 | `group.Add(1)` | 앞으로 완료될 작업 하나 등록 |
| 2 | `go func() { ... }()` | 별도 실행 흐름 시작 |
| 3 | `defer group.Done()` | 작업 함수 반환 시 등록 하나 해제 |
| 4 | `group.Wait()` | 등록 수가 0이 될 때까지 대기 |

`Add`는 새 고루틴을 시작하기 **전에** 실행합니다.
등록을 작업 함수 안에 넣으면 호출자가 먼저 `Wait`를 만나 0을 볼 수 있습니다.
그 경우 아직 시작되지 않은 작업을 기다리지 않고 통과할 수 있습니다.

`Done()`은 관례적으로 함수 입구에서 `defer`로 등록합니다.
정상 경로 끝에 한 번만 직접 쓰면 중간 `return`에서 빠뜨리기 쉽습니다.
`defer`는 해당 함수가 반환할 때 실행되는 약속이므로 분기를 함께 처리합니다.

### 2.2 결과를 읽는 시점

작업은 `result`에 쓰고, 호출자는 `Wait` 뒤에서 읽습니다.
서로 다른 고루틴이 같은 변수에 접근하지만 두 접근의 순서를 동기화했습니다.
쓰기와 읽기가 정해진 관계 없이 겹치는 것이 아닙니다.

`Done`으로 풀린 `Wait` 이후에는 해당 작업이 완료 전에 기록한 결과를 읽을 수 있습니다.
[WaitGroup](https://pkg.go.dev/sync#WaitGroup)은 이 동기화 관계를 문서화합니다.
단순히 “시간이 충분히 흘렀다”는 추측과 다른 보장입니다.

### 2.3 WaitGroup이 하지 않는 일

- 동시에 같은 값을 쓰는 작업들을 서로 배타적으로 실행하지 않습니다.
- 작업의 오류를 자동으로 저장하거나 반환하지 않습니다.
- 실행 중인 함수를 강제로 중단하지 않습니다.
- 작업을 시작한 순서대로 실행하거나 출력하게 만들지 않습니다.

이 예제는 쓰는 작업자가 한 명이라 결과 변수 하나를 사용했습니다.
작업자가 여러 명이면 결과의 저장 위치도 따로 설계해야 합니다.

## 3. 고루틴에 전달하는 값은 언제 정해질까?

`go` 뒤의 호출에도 함수값과 인자를 평가하는 단계가 있습니다.
평가는 호출한 고루틴에서 먼저 일어납니다.
그 다음 별도 고루틴이 평가된 인자를 받아 함수 본문을 실행합니다.

소스: [arguments/main.go](../../examples/18-goroutines/arguments/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	number := 7
	result := 0
	var group sync.WaitGroup
	group.Add(1)
	go func(value int) {
		defer group.Done()
		result = value * 2
	}(number)
	number = 100
	group.Wait()
	fmt.Println("원래 변수:", number)
	fmt.Println("전달된 값의 결과:", result)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/18-goroutines/arguments
```

예상 출력:

```text
원래 변수: 100
전달된 값의 결과: 14
```

### 3.1 전달된 정수와 원래 변수

`number`가 7일 때 호출 인자로 평가됩니다.
새 함수의 `value`는 그 정수값을 받습니다.
호출한 쪽의 `number`를 100으로 바꿔도 이미 전달된 정수는 바뀌지 않습니다.

작업이 `number`를 직접 읽는 클로저로 바뀌면 이야기가 달라집니다.
다음은 **의도적으로 잘못된 발췌**입니다.
일반 실행 파일에는 넣지 않습니다.

```go
go func() {
    result = number * 2 // number를 다른 고루틴에서 직접 읽음
}()
number = 100           // 읽기와 쓰기를 정렬한 장치가 없음
```

`Wait`가 뒤에 있어도 이 두 접근 사이의 경쟁을 없애 주지는 않습니다.
`Wait`는 완료 뒤의 읽기를 정렬하며, 작업 중인 서로 다른 접근을 자동으로 잠그지 않습니다.
공유 상태의 보호는 [20장](../20-sync/README.md)에서 다룹니다.

### 3.2 값 전달이 깊은 복사를 뜻하지는 않는다

정수는 독립된 값이지만 슬라이스의 복사는 같은 배열을 가리킬 수 있습니다.
맵·포인터도 원래 데이터에 접근하는 경로를 함께 전달할 수 있습니다.

| 인자 | 복사되어 전달되는 것 | 별도 주의점 |
| --- | --- | --- |
| `int`, `bool` | 값 | 원래 변수의 이후 대입과 독립 |
| 구조체 | 필드 값 | 포인터·슬라이스 필드는 내부 저장 공간 공유 가능 |
| 슬라이스 | 배열을 가리키는 정보·길이·용량 | 원소에 대한 동시 변경 주의 |
| 맵 | 같은 맵을 가리키는 값 | 읽기·쓰기 정책 필요 |
| 포인터 | 주소값 | 가리킨 대상의 수명과 동기화 필요 |

따라서 “고루틴 인자로 넘겼으니 공유하지 않는다”는 결론은 틀립니다.
[05장](../05-arrays-slices/README.md)과
[08장](../08-pointers-structs/README.md)의 값 복사를 함께 떠올립니다.

### 3.3 반복 변수와 Go 버전

Go 1.22 이후, `for`나 `range`에서 `:=`로 선언한 반복 변수는
각 반복의 값으로 구분되어 클로저에 잡힙니다.
기존 변수를 `=`로 재사용하는 코드나 별도로 선언된 공유 변수까지 바뀐다는 뜻은 아닙니다.

이 장은 중요한 예제에서 인자를 명시해 데이터 전달을 눈에 보이게 합니다.
마지막 `WaitGroup.Go` 예제에서는 각 반복의 `index`를 바로 사용해 두 형태를 비교합니다.
[for 문](https://go.dev/ref/spec#For_statements)과
[go 문](https://go.dev/ref/spec#Go_statements)을 확인합니다.

## 4. 실행 순서와 결과 순서를 분리하기

작업들이 어떤 순서로 끝날지 모르더라도 결과를 입력 순서로 반환할 수 있습니다.
완료 순서를 강제로 고정하는 대신, 입력마다 쓸 결과 칸을 미리 정합니다.

소스: [indexed-results/main.go](../../examples/18-goroutines/indexed-results/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

func squares(values []int) ([]int, error) {
	if len(values) > 16 {
		return nil, errors.New("값은 16개까지 허용합니다")
	}
	for _, value := range values {
		if value < -1000 || value > 1000 {
			return nil, errors.New("각 값은 -1000~1000이어야 합니다")
		}
	}
	results := make([]int, len(values))
	var group sync.WaitGroup
	for index, value := range values {
		group.Add(1)
		go func(index, value int) {
			defer group.Done()
			results[index] = value * value
		}(index, value)
	}
	group.Wait()
	return results, nil
}

func main() {
	results, err := squares([]int{3, 1, 2})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("입력 순서의 결과:", results)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/18-goroutines/indexed-results
```

예상 출력:

```text
입력 순서의 결과: [9 1 4]
```

### 4.1 결과 배열의 소유권

`results`의 길이는 작업을 시작하기 전에 확정합니다.
작업 0은 `results[0]`, 작업 1은 `results[1]`만 씁니다.
동일한 원소를 두 고루틴이 쓰지 않습니다.
호출자는 모든 작업이 끝난 다음 결과를 읽습니다.

| 객체 | 실행 중 읽기 | 실행 중 쓰기 |
| --- | --- | --- |
| `values` | 입력 검증과 인자 평가 | 없음 |
| `results` 슬라이스 정보 | 각 작업이 인덱싱 | 없음 |
| `results[index]` | 완료 후 호출자 | 그 인덱스의 작업 하나 |
| `group` | sync의 메서드로만 사용 | sync의 메서드로만 사용 |

이 표가 안전성의 설명입니다.
단순히 테스트가 한 번 통과했다는 사실만으로 안전하다고 판단하지 않습니다.

### 4.2 append는 왜 같은 방식이 아닐까?

다음은 **의도적으로 잘못된 발췌**입니다.

```go
go func() {
    results = append(results, value*value)
}()
```

여러 작업이 같은 슬라이스의 길이·용량·배열 연결을 함께 바꿉니다.
서로 다른 값이라는 이유로 안전해지지 않습니다.
결과의 순서도 완료 순서에 따라 달라질 수 있습니다.

현재 예제에서는 길이를 미리 알고 있으므로 인덱스 분할이 잘 맞습니다.
결과 수를 모른다면 결과 채널을 한 소비자가 읽어 `append`하거나,
공유 컬렉션을 잠그는 방법을 선택할 수 있습니다.
다음 두 장에서 각각 다룹니다.

### 4.3 입력 상한은 언어의 제한이 아니다

입력 최대 16개와 값 범위 -1000~1000은 **교재의 설계 선택**입니다.
Go가 그 범위만 지원한다는 뜻은 아닙니다.
이 범위는 한 입력당 한 고루틴을 시작하는 실습을 작게 유지하고,
32비트 환경에서도 제곱값이 int 범위 안에 있도록 합니다.

검증을 모두 마친 뒤 고루틴을 시작하므로 잘못된 입력으로 일부 작업만 시작되지 않습니다.
빈 입력에는 길이 0인 결과와 nil 오류를 반환합니다.
입력이 잘못되면 nil 결과와 오류를 반환합니다.

### 4.4 관찰할 결과

입력은 3, 1, 2이고 결과는 9, 1, 4입니다.
출력은 정렬한 값이 아니라 **원래 입력 위치에 대응하는 값**입니다.

작업이 1→2→0 순서로 끝나더라도 같은 결과가 됩니다.
직접 고루틴에서 `fmt.Println`을 호출하면 줄 순서가 달라질 수 있지만,
지금은 합류한 뒤 한 곳에서 출력하므로 문서와 테스트가 같은 순서를 확인합니다.

## 5. 조기 반환에서도 완료를 알리기

작업은 항상 함수의 마지막 줄까지 도달하지 않습니다.
검증 실패, 건너뛰기, 취소 같은 분기가 추가되어도 완료 등록을 빠뜨리면 안 됩니다.

소스: [early-return/main.go](../../examples/18-goroutines/early-return/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	values := []int{-1, 0, 2}
	results := make([]string, len(values))
	var group sync.WaitGroup
	for index, value := range values {
		group.Add(1)
		go func(index, value int) {
			defer group.Done()
			if value <= 0 {
				results[index] = "건너뜀"
				return
			}
			results[index] = "처리"
		}(index, value)
	}
	group.Wait()
	for index, result := range results {
		fmt.Printf("%d: %s\n", values[index], result)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/18-goroutines/early-return
```

예상 출력:

```text
-1: 건너뜀
0: 건너뜀
2: 처리
```

### 5.1 작업마다 한 번 등록하고 한 번 끝내기

입력 세 개에 대해 `Add(1)`도 세 번 실행됩니다.
각 작업은 시작하자마자 `defer group.Done()`을 등록합니다.
양수가 아니어서 중간에 반환해도 해당 `Done`이 실행됩니다.

등록과 완료의 수가 맞지 않으면 다음 문제가 생깁니다.

| 문제 | 결과 |
| --- | --- |
| `Done`을 빠뜨림 | 대기하는 함수가 계속 기다릴 수 있음 |
| 같은 작업에서 `Done`을 두 번 호출 | 카운터가 음수가 되어 패닉 가능 |
| 등록 전 `Wait`를 시작 | 기다려야 할 작업을 놓칠 수 있음 |
| `WaitGroup`을 값으로 복사 | 서로 다른 상태를 기다리는 잘못된 구조 |

WaitGroup은 처음 사용한 뒤 복사하지 않습니다.
함수에 전달해야 한다면 `*sync.WaitGroup`을 사용하되,
가능하면 작업을 만든 범위에서 기다림까지 소유합니다.

### 5.2 defer와 프로세스 종료의 경계

`defer`가 처리하는 것은 함수의 반환 경로입니다.
`os.Exit`는 프로세스를 즉시 종료하므로 대기 중인 defer를 실행하지 않습니다.
작업 고루틴에서 오류를 만났다고 `os.Exit`를 호출하면
다른 작업의 완료와 자원 정리도 함께 건너뜁니다.

작업은 오류를 값으로 기록하고 반환해야 합니다.
최상위 `main`이 모든 결과를 확인한 뒤 종료 코드를 결정하는 편이 책임을 분명히 합니다.
[os.Exit](https://pkg.go.dev/os#Exit)와
[defer 문](https://go.dev/ref/spec#Defer_statements)을 함께 확인합니다.

### 5.3 패닉을 일상적인 오류 전달로 사용하지 않기

파일을 못 찾거나 입력을 해석하지 못하는 상황은 보통 `error`로 표현합니다.
호출자 고루틴에 둔 `recover`가 다른 고루틴의 패닉을 잡아 주지는 않습니다.
패닉은 발생한 고루틴의 함수 반환 경로를 따라 처리됩니다.

뒤의 `WaitGroup.Go`는 전달한 함수가 패닉하지 않아야 한다는 계약도 갖습니다.
“완료 등록을 했으니 모든 실패를 복구한다”고 해석하지 않습니다.
이 장은 예상 가능한 실패를 명시적인 오류값으로 반환합니다.

## 6. 오류도 결과처럼 소유권을 나누기

고루틴으로 실행한 함수의 반환값은 호출자의 변수에 자동으로 전달되지 않습니다.
성공 결과와 오류를 담을 위치를 미리 만들고, 기다린 뒤 읽습니다.

소스: [error-results/main.go](../../examples/18-goroutines/error-results/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

func parseAll(inputs []string) ([]int, error) {
	if len(inputs) > 16 {
		return nil, errors.New("입력은 16개까지 허용합니다")
	}
	results := make([]int, len(inputs))
	failures := make([]error, len(inputs))
	var group sync.WaitGroup
	for index, input := range inputs {
		group.Add(1)
		go func(index int, input string) {
			defer group.Done()
			value, err := strconv.Atoi(input)
			if err != nil {
				failures[index] = fmt.Errorf("항목 %d: %w", index, err)
				return
			}
			results[index] = value
		}(index, input)
	}
	group.Wait()
	if err := errors.Join(failures...); err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	results, err := parseAll([]string{"10", "x", "30"})
	var numberError *strconv.NumError
	fmt.Println("변환 실패:", err != nil)
	fmt.Println("실패 결과 없음:", results == nil)
	fmt.Println("숫자 오류 확인:", errors.As(err, &numberError))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/18-goroutines/error-results
```

예상 출력:

```text
변환 실패: true
실패 결과 없음: true
숫자 오류 확인: true
```

### 6.1 성공값과 오류값을 분리한 이유

숫자 변환은 성공하면 정수를, 실패하면 오류를 만듭니다.
각 고루틴은 자기 인덱스의 `results`와 `failures`만 접근합니다.
일부 작업이 실패해도 나머지 작업은 완료까지 진행합니다.

모든 작업을 기다린 뒤 `errors.Join(failures...)`를 호출합니다.
nil 오류는 제외하고 실제 오류들을 합칩니다.
실패가 하나라도 있으면 결과 슬라이스는 반환하지 않습니다.

이것은 **전체 성공 시에만 결과를 제공하는 정책**입니다.
부분 성공 결과가 필요한 프로그램이라면
각 항목에 값·오류를 함께 둔 구조체를 반환하는 편이 낫습니다.
[31장](../31-concurrent-batch/README.md)에서 그 정책을 구현합니다.

### 6.2 먼저 실패하면 즉시 반환해도 될까?

바로 반환하면 이미 시작한 다른 작업이 결과 슬라이스에 쓰는 중일 수 있습니다.
호출자는 끝난 것으로 생각하지만 내부 작업의 수명이 남는 셈입니다.

빠른 중단이 필요하다면 다음 두 단계가 필요합니다.

1. 남은 작업에 Context 취소를 요청합니다.
2. 실제로 모든 작업이 돌아왔는지 기다립니다.

이 예제는 입력이 작고 처리도 짧아 취소 없이 모두 완료하는 정책을 택했습니다.
첫 실패만 보여 주는 대신 입력 위치 순서대로 오류를 확인할 수 있습니다.

### 6.3 오류의 종류를 보존하기

`fmt.Errorf("항목 %d: %w", index, err)`는 입력 위치를 추가하면서
원래의 숫자 변환 오류를 감쌉니다.
`errors.As`로 `*strconv.NumError`를 찾을 수 있습니다.

에러 문자열이 정확히 어떤 문장인지에 테스트를 강하게 묶지 않고,
변환 실패 여부·결과 반환 정책·오류 타입을 확인합니다.
[errors.Join](https://pkg.go.dev/errors#Join)과
[errors.As](https://pkg.go.dev/errors#As)의 계약을 참고합니다.

### 6.4 입력 순서와 실패 순서

실제로 가장 먼저 실패한 작업은 매번 다를 수 있습니다.
현재 구현은 실패한 **시각**을 기록하지 않습니다.
입력 인덱스 순서로 모은 오류 목록만 제공합니다.

시간순 오류 기록이 필요한 시스템은 별도의 이벤트 기록 구조가 필요합니다.
단순히 반환 오류의 첫 줄을 “가장 먼저 발생한 오류”라고 부르지 않습니다.

## 7. 취소를 요청하고 종료까지 기다리기

작업이 시작되었다는 신호를 확인한 다음 취소하고,
작업의 종료를 기다리는 예제입니다.
`time.Sleep`으로 시작 시점을 추측하지 않습니다.

소스: [cancel-join/main.go](../../examples/18-goroutines/cancel-join/main.go)

```go
package main

import (
	"context"
	"fmt"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	stopped := false
	var group sync.WaitGroup
	group.Add(1)
	go func() {
		defer group.Done()
		close(started)
		<-ctx.Done()
		stopped = true
	}()
	<-started
	cancel()
	group.Wait()
	fmt.Println("시작 확인:", true)
	fmt.Println("종료 확인:", stopped)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/18-goroutines/cancel-join
```

예상 출력:

```text
시작 확인: true
종료 확인: true
```

### 7.1 여기서 필요한 채널 구문

`make(chan struct{})`는 값 자체보다 신호를 전달할 때 쓸 빈 구조체 채널입니다.
`close(started)`는 시작 신호를 한 번 알립니다.
`<-started`는 그 채널이 닫힐 때까지 기다립니다.

작업자는 시작 신호를 보낸 다음 `<-ctx.Done()`에서 취소를 기다립니다.
호출자는 시작 신호를 받은 뒤 `cancel()`을 호출합니다.
이렇게 두 흐름의 관계를 명시적으로 연결합니다.

채널에 값을 보내고 받는 일반 문법과 닫기 규칙은
[19장](../19-channels-select/README.md)에서 자세히 배웁니다.
지금은 **시작 신호**와 **취소 신호**라는 역할에 집중합니다.

### 7.2 사건의 순서

| 사건 | 수행하는 쪽 | 이후 가능한 일 |
| --- | --- | --- |
| `Add(1)` | 호출자 | 작업을 시작해도 기다릴 수 있음 |
| `close(started)` | 작업자 | 호출자가 시작 확인을 마침 |
| `cancel()` | 호출자 | Done이 닫혀 작업자가 취소를 관찰 |
| `stopped = true` | 작업자 | 종료 상태 기록 |
| `Done()` | 작업자 | Wait의 완료 조건에 도달 |
| `Wait` 반환 | 호출자 | stopped를 안전하게 읽음 |

`cancel()` 다음 줄에서 바로 `stopped`를 읽으면 이 관계를 건너뜁니다.
취소 함수는 종료 요청이며 종료를 기다리는 함수가 아닙니다.
[CancelFunc](https://pkg.go.dev/context#CancelFunc)가 이 차이를 명시합니다.

### 7.3 협력하지 않는 작업

작업 함수가 무한 계산을 하면서 Context를 확인하지 않으면
호출자가 취소해도 종료되지 않을 수 있습니다.
`WaitGroup`이 강제로 그 함수를 끊을 수도 없습니다.

반복 처리에는 적절한 확인 지점을 두고,
입출력 API는 Context나 기한을 지원하는 형태를 선택해야 합니다.
아무런 취소 기능이 없는 외부 호출을 새 고루틴으로 감싼다고 해결되지는 않습니다.
호출자를 먼저 반환시키면서 고루틴만 남길 수 있기 때문입니다.

### 7.4 종료를 확인해야 자원을 정리할 수 있다

고루틴이 쓰는 파일·버퍼·DB 연결을 호출자가 먼저 닫거나 재사용하면 안 됩니다.
자원을 누가 소유하고 어느 시점에 사용이 끝나는지 정해야 합니다.

현재 예제에서는 작업이 기록하는 bool만 있지만,
같은 순서가 파일 쓰기나 HTTP 작업의 종료에도 적용됩니다.
취소 요청 → 작업 반환 → 자원 재사용의 순서를 생각합니다.

## 8. WaitGroup.Go로 등록과 실행 묶기

Go 1.25에 추가된 `WaitGroup.Go`는 작업 등록·고루틴 시작·완료를 묶습니다.
현재 교재의 Go 1.27 기준에서는 사용할 수 있습니다.
앞서 배운 `Add`·`Done`의 책임이 사라지는 것이 아니라 메서드가 함께 수행합니다.

소스: [waitgroup-go/main.go](../../examples/18-goroutines/waitgroup-go/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	results := make([]string, 3)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() {
			results[index] = fmt.Sprintf("작업 %d", index+1)
		})
	}
	group.Wait()
	for _, result := range results {
		fmt.Println(result)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/18-goroutines/waitgroup-go
```

예상 출력:

```text
작업 1
작업 2
작업 3
```

### 8.1 두 형태의 관계

전통적인 형태는 다음과 같습니다.

```go
var group sync.WaitGroup
group.Add(1)
go func() {
    defer group.Done()
    // 작업
}()
group.Wait()
```

새 형태는 다음과 같습니다.

```go
var group sync.WaitGroup
group.Go(func() {
    // 작업
})
group.Wait()
```

`group.Go`를 다시 `go group.Go(...)`처럼 감싸지 않습니다.
그렇게 하면 등록 자체를 비동기로 미루어 기다림과 경쟁하게 만들 수 있습니다.
`Go` 안에서 `Done`을 직접 더 호출하지도 않습니다.

### 8.2 그래도 남는 계약

전달하는 함수는 패닉하지 않아야 합니다.
예상 가능한 실패는 반환값을 기록하는 방식으로 표현합니다.
`Go`는 오류 결과를 반환하는 API가 아니므로 오류 저장 정책도 따로 필요합니다.

WaitGroup이 비어 있을 때 첫 `Go` 호출은 `Wait`보다 먼저 일어나야 합니다.
이미 등록된 작업이 새 하위 작업을 등록하는 형태는 가능하지만,
입문 실습에서는 모든 등록을 한 범위에서 끝내고 기다리는 구조를 유지합니다.

이전 작업 묶음을 기다린 다음 새 작업 묶음에 재사용할 수 있습니다.
그때는 이전 `Wait` 호출들이 모두 끝난 뒤 새 작업을 등록합니다.
[WaitGroup.Go](https://pkg.go.dev/sync#WaitGroup.Go)의 세부 규칙을 참고합니다.

### 8.3 출력이 안정적인 이유를 다시 확인하기

각 반복의 `index`가 구분되고,
각 작업은 자기 `results[index]`만 씁니다.
작업의 출력은 기다린 뒤 호출자가 차례대로 수행합니다.

안전하고 결정적인 결과를 만드는 것은 `Go`라는 메서드 이름 하나가 아닙니다.
등록 시점, 데이터 소유권, 결과를 읽는 시점, 출력 위치가 함께 맞아야 합니다.

## 9. 작업 수를 제한하며 확장하기

### 9.1 한 항목에 한 고루틴의 적용 범위

작은 입력에는 직관적인 구조지만 큰 입력에는 그대로 확장하지 않습니다.
입력 100만 개에 고루틴 100만 개를 만들면 모든 작업의 상태를 함께 유지해야 합니다.
실제 외부 서비스가 그만큼의 요청을 받아들일 수 있다는 보장도 없습니다.

입력 상한을 두는 것은 실습 범위를 정하는 첫 단계입니다.
그 다음은 입력 개수와 동시에 일하는 작업자 수를 분리하는 것입니다.

| 구분 | 한 항목당 고루틴 | 고정된 작업자 |
| --- | --- | --- |
| 고루틴 수 | 입력 수에 비례 | 지정한 작업자 수 이하 |
| 작업 분배 | 인자 하나를 바로 전달 | 인덱스 분할 또는 채널 |
| 어울리는 상황 | 작고 상한이 명확한 작업 묶음 | 입력이 많거나 외부 자원이 제한된 작업 |
| 주의점 | 입력 증가에 따른 자원 사용 | 분배·취소·종료 정책 |

### 9.2 채널 없이 나누는 방법

고정된 배열 계산은 인덱스를 나누어 맡길 수 있습니다.
작업자 두 명이 다섯 값을 처리한다면 다음처럼 나눕니다.

| 작업자 | 맡는 인덱스 |
| --- | --- |
| 0 | 0, 2, 4 |
| 1 | 1, 3 |

각 작업자는 지역 변수에 부분 합계를 만듭니다.
마지막에 자기 결과 칸 하나에 기록합니다.
호출자는 모든 작업자를 기다린 뒤 부분 합계를 더합니다.

이 방식은 작업마다 계산 비용이 크게 다르면 부하가 고르지 않을 수 있습니다.
채널을 사용하면 먼저 끝난 작업자가 다음 항목을 가져가게 만들 수 있습니다.
다음 장에서 데이터 전달 방식을 배우고 31장에서 배치로 연결합니다.

### 9.3 입력과 콜백의 동시 사용 계약

슬라이스를 읽기만 하는 함수라도 호출자가 동시에 원소를 바꾸면 안전하지 않습니다.
이 장의 병렬 함수는 **호출이 끝날 때까지 입력을 다른 곳에서 바꾸지 않는 것**을 전제로 합니다.

콜백 함수도 여러 작업에서 동시에 호출될 수 있습니다.
콜백이 공유 맵이나 카운터를 수정한다면 그 부분은 별도의 동기화가 필요합니다.
입력 인덱스의 결과 저장을 안전하게 나누었다고 콜백의 내부까지 안전해지지는 않습니다.

## 10. 테스트에서 시간 대신 관계를 확인하기

### 10.1 잠들면 왜 확인이 약해질까?

다음은 **의도적으로 잘못된 테스트 전략**입니다.

```go
// 작업 시작
// time.Sleep(100 * time.Millisecond)
// 아마 끝났을 것이라고 가정하고 결과 읽기
```

실행 환경이 느려지면 충분히 기다리지 못할 수 있습니다.
반대로 늘 넉넉히 기다리면 테스트가 필요 없이 느려집니다.
가장 큰 문제는 시간이 지나도 쓰기와 읽기의 동기화 관계가 생기지 않는다는 점입니다.

완료를 알아야 하면 `Wait`나 완료 신호를 사용합니다.
시작한 뒤 취소해야 하면 시작 신호를 먼저 기다립니다.
중간 상태를 재현해야 하면 작업이 해당 지점에 도착했다는 신호를 보냅니다.

### 10.2 테스트의 제한 시간은 실패 감지용이다

완료 신호와 함께 5초 타이머를 두는 테스트가 있습니다.
5초가 지난 뒤 성공으로 간주하는 것이 아닙니다.
5초 안에 필요한 신호가 오지 않으면 테스트를 실패시키는 장치입니다.

| 사용 | 의미 |
| --- | --- |
| 완료 채널 수신 | 올바른 동작을 확인하는 조건 |
| 제한 시간 도달 | 종료 계약을 지키지 못했다는 실패 조건 |
| 잠든 후 무조건 성공 | 동작의 증거가 되지 않음 |

[완료 대기 해설의 테스트](../../solutions/18-goroutines/joined-task/main_test.go)는
시작 신호 → 취소 → 함수 반환 → 종료 신호 확인을 검사합니다.
반환 뒤에도 작업이 남는 잘못된 구현을 잡을 수 있습니다.

### 10.3 경계값과 순서

결과 함수는 다음을 확인합니다.

| 입력 | 확인할 계약 |
| --- | --- |
| 빈 입력 | 빈 결과와 nil 오류 |
| 결과 순서가 눈에 띄는 입력 | 입력 위치와 출력 위치의 대응 |
| 허용되는 최대 개수 | 상한에서 성공 |
| 상한보다 하나 많은 개수 | 실행 전 오류 |
| 잘못된 값 여러 개 | 실패 결과 정책과 오류 결합 |
| 이미 취소된 Context | 작업을 시작하지 않음 |
| 실행 중 취소 | 작업 종료 뒤 반환 |

구현한 고루틴의 식별자나 런타임 내부 스케줄링 순서를 검사하지 않습니다.
호출자가 관찰할 결과와 수명을 검사합니다.

### 10.4 실행 명령

```sh
go test ./examples/18-goroutines/... ./solutions/18-goroutines/...
go vet ./examples/18-goroutines/... ./solutions/18-goroutines/...
go build ./examples/18-goroutines/... ./solutions/18-goroutines/...
```

Linux와 C 컴파일러가 준비된 환경에서는 레이스 검사도 실행합니다.

```sh
go test -race ./examples/18-goroutines/... ./solutions/18-goroutines/...
```

레이스 검사는 실제 실행한 경로의 충돌을 찾습니다.
한 번 통과했다고 모든 가능한 경로가 증명된 것은 아닙니다.
데이터 소유권 설명과 중요한 경로의 테스트가 함께 필요합니다.
[공식 race detector 안내](https://go.dev/doc/articles/race_detector)를 참고합니다.

## 11. 자주 하는 실수

| 증상 | 원인 | 확인과 수정 |
| --- | --- | --- |
| 작업 출력이 가끔 사라짐 | main이 먼저 반환 | 작업을 등록하고 기다림 |
| 프로그램이 끝나지 않음 | Done 누락 또는 작업 자체가 미종료 | 모든 반환 경로와 취소 확인점 검사 |
| Wait가 너무 일찍 반환 | 고루틴 안에서 Add | 시작 전에 등록 |
| 결과 수나 순서가 달라짐 | 여러 작업의 공유 append | 인덱스 분할 또는 단일 수집자 |
| 취소 후에도 자원이 사용됨 | cancel만 호출하고 반환 | 실제 종료를 기다림 |
| 오류가 출력만 되고 호출자는 성공 | 오류 전달 통로 없음 | 결과와 함께 오류 수집 |
| 레이스가 발생 | Wait 전에 결과를 읽음 또는 공유 값 변경 | 접근 관계를 표로 그려 확인 |
| WaitGroup 복사 경고 | 값 수신자·값 인자 | 복사하지 않는 소유 구조 사용 |
| 고루틴 수가 입력만큼 폭증 | 무제한 시작 | 입력 상한·고정 작업자 도입 |

오류를 수정할 때 Sleep을 더 길게 넣어 가리는 방법은 사용하지 않습니다.
누가 무엇을 쓰고, 누가 언제 읽으며, 누가 완료를 알리는지부터 확인합니다.

## 12. 연습문제

[별도 연습문제](../../exercises/18-goroutines/README.md)에서 다음 과제를 수행합니다.

| 문제 | 수준 | 목표 |
| --- | --- | --- |
| 1 | 확인 | 등록·완료·취소의 역할 설명 |
| 2 | 변형 | 입력 순서를 보존하는 이름표 생성 |
| 3 | 확인·수정 | Add 위치·공유 append·복사 문제 찾기 |
| 4 | 응용 | 고정 작업자로 부분 합계 계산 |
| 5 | 응용 | 모든 작업을 합류한 뒤 변환 결과 반환 |
| 6 | 응용 | 취소된 작업의 실제 종료까지 기다리기 |
| 7 | 선택 | 완료 순서와 출력 순서를 의도적으로 분리 |

문제 파일에서 입력·상한·결과 정책·테스트 조건을 확인합니다.
코드를 작성한 뒤 [해설](../../solutions/18-goroutines/README.md)과 비교합니다.
해설과 함수 본문의 모양이 같아야 하는 것은 아닙니다.
수명·오류·순서의 약속이 같아야 합니다.

## 13. 요약과 자기 점검

- `go` 문은 독립된 실행 흐름을 시작하고 완료를 기다리지 않습니다.
- 등록은 시작 전에, 완료 알림은 모든 반환 경로에, 읽기는 합류 뒤에 둡니다.
- WaitGroup은 완료 대기를 제공하며 공유 데이터를 자동으로 보호하지 않습니다.
- 입력마다 결과 칸을 미리 나누면 완료 순서와 반환 순서를 분리할 수 있습니다.
- Context 취소와 고루틴 종료 확인은 각각 필요합니다.
- 큰 입력은 제한된 작업자 수로 나누어 처리합니다.

답을 보지 않고 다음을 설명해 보세요.

1. `go f(x)`와 `f(x)`는 반환 시점에서 어떻게 다른가요?
2. Add를 고루틴 내부에 넣으면 어떤 순서에서 문제가 생기나요?
3. Wait 뒤에서 읽는 결과와 작업 중 함께 수정하는 결과는 어떻게 다른가요?
4. 인자로 전달한 슬라이스는 원래 배열과 반드시 독립적인가요?
5. 결과 배열의 서로 다른 원소를 쓰는 것과 공유 append는 왜 다른가요?
6. 취소 요청 후 Wait 없이 반환하면 어떤 책임이 남나요?
7. WaitGroup.Go 안에서 예상 가능한 오류를 어떻게 전달해야 하나요?
8. 입력이 100만 개가 되면 현재 구조를 어디부터 바꿔야 하나요?

## 공식 참고 자료

공식 문서의 API와 버전 표기를 2026-11-09에 확인했습니다.

- [언어 명세: go 문](https://go.dev/ref/spec#Go_statements)
- [언어 명세: 프로그램 실행](https://go.dev/ref/spec#Program_execution)
- [언어 명세: for 문](https://go.dev/ref/spec#For_statements)
- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [sync.WaitGroup.Go](https://pkg.go.dev/sync#WaitGroup.Go)
- [context.CancelFunc](https://pkg.go.dev/context#CancelFunc)
- [Go 메모리 모델](https://go.dev/ref/mem)
- [Data Race Detector](https://go.dev/doc/articles/race_detector)

[이전: 17장 Context](../17-context/README.md) ·
[목차](../README.md) ·
[다음: 19장 채널과 select](../19-channels-select/README.md)
