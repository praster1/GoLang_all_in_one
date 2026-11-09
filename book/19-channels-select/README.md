# 19. 채널과 select

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 응용 | 고루틴·완료 대기·Context·슬라이스 | Go, 터미널, 이 저장소 |

앞 장에서는 각 작업이 자기 결과 칸에 쓰고,
호출자가 모든 작업을 기다린 뒤 결과를 읽었습니다.
이번에는 한 작업이 만든 값을 다른 작업이 **진행 중에** 받도록 연결합니다.
이때 사용하는 타입이 **채널(channel)**입니다.

채널은 값의 전달과 실행 흐름의 동기화를 함께 표현합니다.
그러나 채널을 하나 만들었다고 종료나 취소가 자동으로 해결되지는 않습니다.
보내는 쪽이 계속 기다리는데 받는 쪽이 먼저 반환하면 작업이 남을 수 있습니다.
누가 보내고, 누가 받으며, 누가 닫고, 누가 취소 후 기다릴지까지 정해야 합니다.

## 학습 목표

- 채널의 생성·전송·수신과 버퍼의 의미를 설명한다.
- 닫힌 채널과 nil 채널의 동작을 구분한다.
- 방향이 있는 채널 타입으로 함수의 역할을 표현한다.
- select로 값 전달·취소·즉시 실패 중 하나를 선택한다.
- 여러 입력을 병합하거나 단계를 연결하고, 조기 중단 시 모두 종료한다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
예제는 외부 네트워크·계정·파일 없이 종료됩니다.
출력 순서가 중요하면 한 곳에서 모으거나 정렬해 표시합니다.

[18장](../18-goroutines/README.md)의 WaitGroup과
[17장](../17-context/README.md)의 Context를 함께 사용합니다.
채널은 완료 대기나 취소의 모든 역할을 대신하는 만능 도구가 아닙니다.
작업별로 어떤 관계를 표현하는지 확인합니다.

| 예제 | 중심 내용 | 실행 |
| --- | --- | --- |
| [rendezvous](../../examples/19-channels-select/rendezvous/main.go) | 버퍼 없는 전송·수신 | `go run ./examples/19-channels-select/rendezvous` |
| [buffered-close](../../examples/19-channels-select/buffered-close/main.go) | 버퍼와 닫힌 뒤 남은 값 | `go run ./examples/19-channels-select/buffered-close` |
| [directions](../../examples/19-channels-select/directions/main.go) | 송신·수신 전용 타입 | `go run ./examples/19-channels-select/directions` |
| [select-default](../../examples/19-channels-select/select-default/main.go) | 준비된 통신과 default | `go run ./examples/19-channels-select/select-default` |
| [closed-select](../../examples/19-channels-select/closed-select/main.go) | 닫힌 입력을 nil로 비활성화 | `go run ./examples/19-channels-select/closed-select` |
| [cancel-send](../../examples/19-channels-select/cancel-send/main.go) | 막힌 전송의 취소 | `go run ./examples/19-channels-select/cancel-send` |
| [merge](../../examples/19-channels-select/merge/main.go) | 여러 입력을 하나로 병합 | `go run ./examples/19-channels-select/merge` |
| [pipeline](../../examples/19-channels-select/pipeline/main.go) | 단계 연결과 조기 종료 | `go run ./examples/19-channels-select/pipeline` |

## 1. 값을 넘기는 통로와 기다림

### 1.1 채널 타입과 기본 문법

| 구문 | 의미 |
| --- | --- |
| `chan int` | int 값을 보내고 받을 수 있는 채널 타입 |
| `make(chan int)` | 버퍼가 없는 채널 생성 |
| `make(chan int, 3)` | 최대 세 값을 보관하는 버퍼가 있는 채널 생성 |
| `ch <- value` | 채널에 값을 보냄 |
| `value := <-ch` | 채널에서 값을 받음 |
| `value, ok := <-ch` | 값과 채널의 남은 전달 상태를 함께 받음 |
| `close(ch)` | 더 이상 값을 보내지 않음을 알림 |

화살표의 위치를 읽습니다.
`ch <- value`는 값이 채널로 들어가는 전송입니다.
`<-ch`는 채널에서 값을 꺼내는 수신입니다.
둘은 문장 모양이 비슷해도 기다리는 상대가 다릅니다.

### 1.2 버퍼 없는 채널의 만남

버퍼 없는 채널에서는 보내는 쪽과 받는 쪽이 만날 때 값 전달이 진행됩니다.
보내는 쪽만 실행되면 수신자가 올 때까지 기다리고,
받는 쪽만 실행되면 송신자가 올 때까지 기다립니다.

같은 고루틴에서 값을 먼저 보내고 나중에 받는 구조는
첫 전송에서 막혀 다음 줄로 갈 수 없습니다.
다음은 **의도적으로 멈추는 발췌**이며 기본 예제에 넣지 않습니다.

```go
ch := make(chan int)
ch <- 1    // 수신자가 없어서 여기에서 기다림
value := <-ch
_ = value
```

채널을 만들었다는 사실만으로 새로운 고루틴이 생기지는 않습니다.
값을 주고받을 서로 다른 진행 경로가 있거나 버퍼 공간이 있어야 합니다.
[채널 타입](https://go.dev/ref/spec#Channel_types)과
[전송문](https://go.dev/ref/spec#Send_statements)을 참고합니다.

## 2. 한 값을 전달하고 합류하기

고루틴 하나가 42를 보내고 main이 받습니다.
전송이 끝났다는 사실과 작업 함수의 전체 종료를 구분하기 위해 Wait도 둡니다.

소스: [rendezvous/main.go](../../examples/19-channels-select/rendezvous/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	values := make(chan int)
	var group sync.WaitGroup
	group.Go(func() { values <- 42 })
	value := <-values
	group.Wait()
	fmt.Println("전달한 값:", value)
	fmt.Println("전송 작업 종료")
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/rendezvous
```

예상 출력:

```text
전달한 값: 42
전송 작업 종료
```

### 2.1 전송과 함수 종료의 차이

main이 값을 받으면 해당 전송은 진행할 수 있습니다.
그러나 송신자의 함수에 전송 뒤 코드가 더 있다면 아직 실행 중일 수 있습니다.
수신 하나만으로 송신자의 모든 정리가 끝났다고 해석하지 않습니다.

이 예제는 전송 뒤 특별한 작업이 없지만,
Wait로 작업 수명의 끝을 명시합니다.
앞 장에서 배운 완료 대기와 채널의 값 전달을 함께 사용하는 형태입니다.

### 2.2 이 채널은 닫지 않아도 될까?

가능합니다.
한 값을 정확히 받기로 약속했고, 이후 반복 수신으로 종료를 기다리지 않습니다.
채널을 닫는 것은 메모리 정리 함수 호출과 같은 일이 아닙니다.

닫기는 수신자에게 “이후 값이 더 없다”는 정보를 주기 위해 사용합니다.
고정된 개수만 주고받고 더 이상 사용하지 않는 모든 채널을
무조건 닫아야 하는 것은 아닙니다.

수신자가 `range`로 끝까지 읽는 구조라면 종료를 알릴 닫기가 필요합니다.
다음 예제로 그 차이를 확인합니다.

### 2.3 값 복사와 공유 데이터

정수 42는 값으로 전달됩니다.
슬라이스나 포인터를 채널로 보내면 그 값이 가리키는 데이터까지 깊게 복사되는 것은 아닙니다.

송신자가 슬라이스를 보낸 뒤 같은 배열을 수정하고,
수신자도 그 배열을 읽는다면 별도 소유권 규칙이 필요합니다.
“채널을 통해 보냈다”만으로 이후의 모든 접근이 안전해지지 않습니다.
[Go 메모리 모델의 채널 통신](https://go.dev/ref/mem#chan)을 함께 읽습니다.

## 3. 버퍼에 남은 값과 닫힌 채널

버퍼가 있으면 여유 공간만큼은 수신자가 즉시 없어도 전송이 진행됩니다.
버퍼가 가득 차면 다음 전송은 다시 기다립니다.
닫을 때 버퍼에 남은 값은 버려지지 않습니다.

소스: [buffered-close/main.go](../../examples/19-channels-select/buffered-close/main.go)

```go
package main

import "fmt"

func main() {
	values := make(chan int, 2)
	values <- 10
	values <- 20
	close(values)
	collected := []int{}
	for value := range values {
		collected = append(collected, value)
	}
	value, ok := <-values
	fmt.Println("남아 있던 값:", collected)
	fmt.Printf("닫힌 뒤 수신: value=%d ok=%t\n", value, ok)
	fmt.Printf("길이=%d 용량=%d\n", len(values), cap(values))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/buffered-close
```

예상 출력:

```text
남아 있던 값: [10 20]
닫힌 뒤 수신: value=0 ok=false
길이=0 용량=2
```

### 3.1 닫는 시점에 일어나는 일

두 값을 버퍼에 넣은 뒤 채널을 닫습니다.
`range`는 남아 있던 10과 20을 받습니다.
버퍼가 비고 더 이상 값이 오지 않는 상태를 확인하면 반복을 끝냅니다.

닫기 자체가 버퍼 안의 데이터를 지우지는 않습니다.
닫힌 뒤에도 이전에 보낸 값은 순서대로 받을 수 있습니다.
그 이후의 수신은 해당 타입의 영값과 false를 즉시 반환합니다.

| 채널 상태 | 수신 결과 | 기다리는가? |
| --- | --- | --- |
| 값이 있음 | 실제 값, true | 준비된 값을 받음 |
| 열려 있고 값 없음 | 새 전송을 기다림 | 예 |
| 닫혔지만 버퍼에 값 있음 | 남은 값, true | 아니오 |
| 닫혔고 버퍼도 비었음 | 영값, false | 아니오 |
| nil | 수신 불가능 | 계속 기다림 |

`0`을 받았다는 이유만으로 닫혔다고 판단하지 않습니다.
0도 정상 데이터일 수 있으므로 `ok`를 함께 봅니다.

### 3.2 len과 cap의 한계

이 예제에서는 모든 수신이 끝난 뒤 길이 0, 용량 2를 출력합니다.
동시에 다른 고루틴이 사용 중이라면 `len(ch)`는 그 순간의 버퍼 길이일 뿐입니다.

`if len(ch) > 0 { <-ch }` 같은 코드는
확인과 수신 사이에 다른 수신자가 값을 가져가면 기다릴 수 있습니다.
즉시 받을 수 있는지 확인하려면 뒤에서 배울 select와 default를 사용합니다.

용량은 전달을 잠시 완충하는 크기입니다.
버퍼를 늘린다고 소비자가 사라져도 무한히 안전하게 보내는 것은 아닙니다.
결국 꽉 차면 송신자가 기다립니다.

### 3.3 닫기 규칙

- 닫힌 채널에 보내면 패닉이 발생합니다.
- 이미 닫힌 채널을 다시 닫아도 패닉이 발생합니다.
- nil 채널을 닫아도 패닉이 발생합니다.
- 수신 전용 채널 타입에는 close를 호출할 수 없습니다.

따라서 닫기는 “혹시 모르니 여기저기서 한 번씩” 실행하는 정리 코드가 아닙니다.
마지막 전송이 끝났음을 아는 주체 하나가 소유해야 합니다.
[close의 규칙](https://go.dev/ref/spec#Close)을 참고합니다.

## 4. 함수 타입에 채널의 방향을 표시하기

송신자와 수신자에게 같은 양방향 타입을 줄 필요는 없습니다.
필요한 작업만 할 수 있도록 타입에 방향을 표시하면 함수 역할이 보입니다.

소스: [directions/main.go](../../examples/19-channels-select/directions/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func send(out chan<- int) {
	defer close(out)
	for _, value := range []int{2, 4, 6} {
		out <- value
	}
}
func sum(in <-chan int) int {
	total := 0
	for value := range in {
		total += value
	}
	return total
}
func main() {
	values := make(chan int)
	var group sync.WaitGroup
	group.Go(func() { send(values) })
	total := sum(values)
	group.Wait()
	fmt.Println("합계:", total)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/directions
```

예상 출력:

```text
합계: 12
```

### 4.1 세 타입의 차이

| 타입 | 가능한 작업 | 예제의 역할 |
| --- | --- | --- |
| `chan int` | 전송·수신·닫기 | 생성과 연결을 담당하는 main |
| `chan<- int` | 전송·닫기 | 값을 만드는 send |
| `<-chan int` | 수신 | 값을 모으는 sum |

양방향 채널을 송신 전용이나 수신 전용 매개변수로 넘길 수 있습니다.
함수 안에서는 지정된 방향으로만 사용할 수 있습니다.

송신 전용이라는 타입만으로 닫기의 소유권이 유일해지는 것은 아닙니다.
같은 채널을 여러 송신자에게 넘기면 모두 타입상 close가 가능할 수 있습니다.
실제 소유권은 설계와 호출 구조에서 한 곳으로 정해야 합니다.

### 4.2 send가 close하는 이유

이 예제는 send 하나만 전송합니다.
자신의 마지막 전송이 끝났음을 알고 있으므로 출력 채널을 닫습니다.
sum은 range를 사용해 닫힘까지 읽습니다.

sum이 먼저 채널을 닫는다면 send가 아직 보내는 중일 수 있습니다.
수신자가 더 이상 필요 없다는 뜻은 보통 Context 취소로 전달합니다.
수신자가 생산자의 채널을 임의로 닫는 방식은 사용하지 않습니다.

### 4.3 모두 소비하는 계약

이 버전의 send는 취소를 받지 않습니다.
sum이 모든 값을 읽는다는 단순한 계약에 기대고 있습니다.

중간에 sum이 return하도록 바꾸면 남은 전송이 막힐 수 있습니다.
따라서 이후 예제에서는 전송도 Context를 선택할 수 있게 만듭니다.
작은 예제를 확장할 때 달라지는 계약을 함께 살펴봅니다.

## 5. select와 default로 준비된 통신 선택하기

`select`는 여러 채널 통신 중 진행할 수 있는 하나를 선택합니다.
지금 즉시 진행할 수 있는 통신이 없을 때는 default가 실행됩니다.
default가 없으면 적어도 하나가 진행할 수 있을 때까지 기다립니다.

소스: [select-default/main.go](../../examples/19-channels-select/select-default/main.go)

```go
package main

import "fmt"

func main() {
	var disabled <-chan int
	ready := make(chan int, 1)
	ready <- 7
	select {
	case value := <-disabled:
		fmt.Println("비활성:", value)
	case value := <-ready:
		fmt.Println("준비된 값:", value)
	default:
		fmt.Println("즉시 처리 없음")
	}
	select {
	case value := <-ready:
		fmt.Println("추가 값:", value)
	default:
		fmt.Println("즉시 처리 없음")
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/select-default
```

예상 출력:

```text
준비된 값: 7
즉시 처리 없음
```

### 5.1 nil 채널은 선택에서 비활성 상태다

`disabled`는 선언만 한 nil 채널입니다.
그 채널의 수신은 진행할 수 없으므로 현재 select에서 선택되지 않습니다.
ready에는 7이 들어 있으므로 그 수신이 선택됩니다.

두 번째 select에서는 ready가 비어 있습니다.
즉시 받을 값이 없고 default가 있으므로 기다리지 않고 default로 진행합니다.

nil 채널을 단독으로 수신하면 영원히 기다립니다.
select 안에서는 해당 case를 잠시 비활성화하는 데 쓸 수 있습니다.
단, 다른 case도 없고 default도 없으면 select 전체가 기다립니다.

### 5.2 위에 적힌 case가 먼저인가?

아닙니다.
둘 이상의 통신이 준비되어 있으면 그중 하나를 의사 난수 방식으로 선택합니다.
소스에 먼저 쓴 case가 우선권을 갖는 것은 아닙니다.
[select 명세](https://go.dev/ref/spec#Select_statements)를 참고합니다.

Context 취소를 첫 case에 써도
동시에 결과 전송이 준비되면 결과 전송이 선택될 수 있습니다.
따라서 “취소 순간 이후 한 번도 처리하지 않는다” 같은 강한 보장을
case 순서만으로 만들 수는 없습니다.

### 5.3 default가 만드는 바쁜 반복

다음은 **의도적으로 잘못된 발췌**입니다.

```go
for {
    select {
    case value := <-ch:
        _ = value
    default:
        // 아무 일 없이 다시 반복
    }
}
```

값이 없으면 기다리지 않고 반복을 계속합니다.
프로그램이 할 일이 없어도 CPU를 사용할 수 있습니다.
기다리는 프로그램이라면 default를 빼거나 실제 다음 작업을 수행해야 합니다.

### 5.4 비차단 전송은 데이터 정책이다

전송 case와 default를 조합하면 가득 찬 채널에서 바로 실패할 수 있습니다.
그때 값을 버릴지, 오류를 반환할지, 나중에 다시 시도할지는 별도의 정책입니다.

로그처럼 일부 유실을 허용하는 자료와 결제 요청처럼 유실할 수 없는 자료는 다릅니다.
교착을 피한다는 이유만으로 default에서 중요한 값을 조용히 버리지 않습니다.

## 6. 닫힌 입력을 select에서 제거하기

닫힌 채널은 영값을 계속 즉시 받을 수 있습니다.
여러 입력을 select로 읽으면서 닫힘을 처리하지 않으면
같은 입력의 영값을 반복해서 처리할 수 있습니다.

소스: [closed-select/main.go](../../examples/19-channels-select/closed-select/main.go)

```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	left := make(chan int, 1)
	right := make(chan int, 1)
	left <- 3
	right <- 1
	close(left)
	close(right)
	values := []int{}
	closed := 0
	for left != nil || right != nil {
		select {
		case value, ok := <-left:
			if !ok {
				left = nil
				closed++
				continue
			}
			values = append(values, value)
		case value, ok := <-right:
			if !ok {
				right = nil
				closed++
				continue
			}
			values = append(values, value)
		}
	}
	sort.Ints(values)
	fmt.Println("모은 값:", values)
	fmt.Println("닫힘 확인:", closed)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/closed-select
```

예상 출력:

```text
모은 값: [1 3]
닫힘 확인: 2
```

### 6.1 false를 확인하면 nil로 바꾸기

`ok`가 false이면 해당 변수를 nil로 바꿉니다.
그 다음 select부터 그 입력은 선택되지 않습니다.
두 입력이 모두 nil이 되면 반복 조건이 거짓이 되어 끝납니다.

채널 자체를 nil로 바꾸는 것이 아니라 이 함수가 가진 채널 **변수**를 바꿉니다.
같은 채널을 다른 변수가 가리키고 있다면 그 변수까지 nil이 되는 것은 아닙니다.

### 6.2 닫힘과 값의 수신은 별개 사건

채널이 이미 닫혀 있어도 버퍼에 값이 남아 있으면 먼저 값과 true를 받습니다.
다음 수신에서 영값과 false를 받았을 때 비활성화합니다.

닫힌 사실을 두 번 세지 않으므로 최종 닫힘 확인 수는 2입니다.
실제 값은 3과 1이며, 출력 직전에 정렬하므로 항상 [1 3]입니다.

여기서 정렬은 화면 출력을 안정시키는 선택입니다.
select가 원래 작은 값부터 읽었다는 뜻이 아닙니다.
완료 순서나 도착 순서가 업무적으로 의미 있다면 임의로 정렬하면 안 됩니다.

### 6.3 누가 입력을 닫는가?

이 예제는 main이 버퍼를 채우고 전송을 끝낸 뒤 직접 닫습니다.
생산자 고루틴이 있다면 생산자 또는 생산자들을 합류하는 담당자가 닫아야 합니다.
소비자가 false를 확인하는 것은 닫기를 반복할 권한이 아닙니다.

## 7. 받는 쪽이 중단해도 송신자가 끝나게 만들기

생산자는 1000개를 보내려고 하지만 소비자는 앞의 3개만 필요합니다.
3개를 받은 뒤 취소하고 생산자의 종료 오류까지 받습니다.

소스: [cancel-send/main.go](../../examples/19-channels-select/cancel-send/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

func sendNumbers(ctx context.Context, count int, out chan<- int) error {
	defer close(out)
	if count < 0 || count > 1000 {
		return errors.New("개수는 0~1000이어야 합니다")
	}
	for value := 1; value <= count; value++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- value:
		}
	}
	return ctx.Err()
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	values := make(chan int)
	done := make(chan error, 1)
	go func() { done <- sendNumbers(ctx, 1000, values) }()
	first := []int{}
	for range 3 {
		first = append(first, <-values)
	}
	cancel()
	err := <-done
	fmt.Println("먼저 받은 값:", first)
	fmt.Println("전송 취소 확인:", errors.Is(err, context.Canceled))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/cancel-send
```

예상 출력:

```text
먼저 받은 값: [1 2 3]
전송 취소 확인: true
```

### 7.1 전송이 막힐 수 있는 곳에 취소 경로를 둔다

단독 `out <- value`는 수신자가 없으면 기다립니다.
이를 select 안에 넣고 Context의 Done도 선택할 수 있게 합니다.
소비자가 더 이상 받지 않아도 취소가 들어오면 전송 대기에서 빠져나옵니다.

반복 입구의 `ctx.Err()`는 이미 취소된 상황에서 다음 작업을 줄입니다.
그 뒤 취소와 전송이 동시에 준비되는 경쟁까지 완전히 없애는 장치는 아닙니다.
필요한 계약은 신호를 관찰하여 작업이 종료하고, 호출자가 그 종료를 확인하는 것입니다.

### 7.2 종료 오류 채널의 용량이 1인 이유

생산자는 종료 오류 하나만 전달합니다.
용량 1이면 호출자가 아직 그 오류를 받기 전이어도 전달을 마칠 수 있습니다.
무한한 결과를 버퍼로 해결한 것이 아니라, 약속된 하나의 완료값을 위한 공간입니다.

완료값이 한 개라는 약속이 있으므로 done을 range로 읽지 않습니다.
한 번 받아 오류를 확인하면 됩니다.
모든 채널을 닫아야 한다는 규칙은 없습니다.

### 7.3 출력 채널의 닫기 책임

`sendNumbers`가 유일한 송신자이며 `defer close(out)`를 소유합니다.
정상 완료, 잘못된 개수, 취소 등 모든 반환 경로에서 출력이 닫힙니다.
호출자는 이 출력 채널을 직접 닫지 않습니다.

Context는 nil이 아니어야 합니다.
개수는 0~1000이라는 실습 상한을 둡니다.
출력 채널은 호출자가 만든 유효한 채널이며 다른 송신자가 공유하지 않는 계약입니다.

### 7.4 취소 후 기다림을 빼면?

`cancel()`만 하고 함수를 끝내면 생산자가 그 신호를 관찰했는지 모릅니다.
프로그램이 오래 실행되는 서비스라면 다음 요청과 자원 사용이 겹칠 수 있습니다.

종료 결과를 받는 행위는 생산자의 본문이 반환했다는 확인입니다.
고루틴에 후속 정리가 더 있다면 그 정리를 끝낸 뒤 완료 신호를 보내도록 설계합니다.
이번 코드는 함수 반환 뒤 오류값을 보내고 추가 작업을 하지 않습니다.

## 8. 여러 입력을 하나로 합치기

여러 입력 채널을 하나로 합치는 패턴을 **팬인(fan-in)**이라고 부릅니다.
입력 하나마다 전달 작업자를 두고 같은 출력으로 보냅니다.
모든 전달 작업자가 끝났을 때 출력 채널을 닫습니다.

소스: [merge/main.go](../../examples/19-channels-select/merge/main.go)

```go
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

func merge(ctx context.Context, inputs ...<-chan int) (<-chan int, <-chan struct{}) {
	out := make(chan int)
	done := make(chan struct{})
	var group sync.WaitGroup
	for _, input := range inputs {
		if input == nil {
			continue
		}
		group.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case value, ok := <-input:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- value:
					}
				}
			}
		})
	}
	go func() { group.Wait(); close(out); close(done) }()
	return out, done
}
func main() {
	left := make(chan int, 2)
	right := make(chan int, 2)
	left <- 1
	left <- 3
	close(left)
	right <- 2
	right <- 4
	close(right)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, done := merge(ctx, left, right)
	values := []int{}
	for value := range out {
		values = append(values, value)
	}
	<-done
	sort.Ints(values)
	fmt.Println("병합 결과:", values)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/merge
```

예상 출력:

```text
병합 결과: [1 2 3 4]
```

### 8.1 닫기의 주체를 분리하기

여러 작업자가 같은 out에 보냅니다.
각 작업자가 defer close(out)를 하면 먼저 끝난 작업자가 채널을 닫고,
다른 작업자가 이후 전송하다 패닉이 날 수 있습니다.

따라서 별도 합류 담당자가 group.Wait를 마친 뒤 out을 한 번만 닫습니다.
입력을 읽는 작업자들은 자기 종료만 알립니다.

| 대상 | 전송자 | 닫는 주체 | 완료 확인 |
| --- | --- | --- | --- |
| 각 input | 외부 생산자 | 해당 생산자 | 입력의 닫힘 |
| out | 전달 작업자 여러 명 | 모든 전달 작업자를 기다린 담당자 | out의 닫힘 |
| done | 값 전송 없음 | 합류 담당자 | 닫힘 수신 |

### 8.2 두 대기 지점 모두 취소를 지원한다

작업자는 입력을 기다릴 수 있고, 받은 값을 출력으로 보내다 기다릴 수도 있습니다.
두 곳 모두 select에 ctx.Done을 넣습니다.
입력 쪽에만 취소를 넣으면 출력 전송에서 계속 막힐 수 있습니다.

nil 입력은 시작할 작업이 없다고 해석해 건너뜁니다.
입력이 하나도 없으면 작업자 수가 0이므로 출력과 done이 곧 닫힙니다.

### 8.3 순서에 대한 약속

각 입력의 값 순서는 유지되지만 서로 다른 입력 사이의 전송 순서는 정해지지 않습니다.
출력 예제는 모든 값을 받은 뒤 정렬해 표시합니다.
중복값은 제거하지 않습니다.

업무에서 입력 전체의 원래 순서를 복원하려면
값과 함께 인덱스를 전달하고 최종 결과 칸에 배치할 수 있습니다.
[31장](../31-concurrent-batch/README.md)의 배치는 이 방법을 사용합니다.

### 8.4 작업자 수에 대한 제한

이 helper는 전달받은 nil이 아닌 채널 수만큼 작업자를 만듭니다.
실습에서는 두 입력만 연결합니다.
호출자가 임의로 수백만 개의 입력을 만들도록 설계한 API는 아닙니다.

공개 라이브러리로 확장할 때는 입력 채널 수의 상한이나
동시에 처리할 작업자 수를 따로 정해야 합니다.
동시성 구조에도 입력 검증과 자원 한도가 필요합니다.

## 9. 파이프라인을 조기에 멈추기

숫자를 만드는 단계와 제곱하는 단계를 채널로 연결합니다.
이처럼 앞 단계의 출력을 다음 단계가 받는 구조를 **파이프라인(pipeline)**이라고 부릅니다.
소비자는 앞의 세 결과만 필요하므로 나머지는 취소합니다.

소스: [pipeline/main.go](../../examples/19-channels-select/pipeline/main.go)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
)

func numbers(ctx context.Context, group *sync.WaitGroup, count int) <-chan int {
	out := make(chan int)
	group.Go(func() {
		defer close(out)
		for value := 1; value <= count; value++ {
			select {
			case <-ctx.Done():
				return
			case out <- value:
			}
		}
	})
	return out
}
func squared(ctx context.Context, group *sync.WaitGroup, in <-chan int) <-chan int {
	out := make(chan int)
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
				select {
				case <-ctx.Done():
					return
				case out <- value * value:
				}
			}
		}
	})
	return out
}
func firstSquares(count, take int) ([]int, error) {
	if count < 0 || count > 1000 || take < 0 || take > count {
		return nil, errors.New("0 <= take <= count <= 1000이 필요합니다")
	}
	ctx, cancel := context.WithCancel(context.Background())
	var group sync.WaitGroup
	defer func() { cancel(); group.Wait() }()
	out := squared(ctx, &group, numbers(ctx, &group, count))
	results := make([]int, 0, take)
	for range take {
		value, ok := <-out
		if !ok {
			return nil, errors.New("예상보다 일찍 종료되었습니다")
		}
		results = append(results, value)
	}
	return results, nil
}
func main() {
	values, err := firstSquares(1000, 3)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("앞의 제곱값:", values)
	fmt.Println("모든 단계 종료 확인")
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/19-channels-select/pipeline
```

예상 출력:

```text
앞의 제곱값: [1 4 9]
모든 단계 종료 확인
```

### 9.1 단계별 책임

| 단계 | 입력 | 출력 | 닫는 대상 |
| --- | --- | --- | --- |
| numbers | 개수 | 1부터 count까지 | 자신의 out |
| squared | 숫자 채널 | 숫자의 제곱 | 자신의 out |
| firstSquares | 제곱값 채널 | 앞의 take개 슬라이스 | Context 취소, Wait 수행 |

각 단계는 자신이 만든 출력 채널만 닫습니다.
입력 채널은 닫지 않습니다.
입력 생산자가 아직 보내는 중일 수 있기 때문입니다.

### 9.2 정상 종료와 조기 종료

끝까지 읽으면 numbers가 닫히고,
squared가 입력 닫힘을 보고 자신의 출력을 닫습니다.
이것이 정상적인 종료 전파입니다.

앞의 세 개만 읽으면 여전히 남은 데이터가 있습니다.
함수가 반환할 때 defer가 cancel을 호출해 두 단계에 중단을 알립니다.
그 다음 group.Wait로 두 단계의 종료를 확인합니다.

중간에서 입력을 받는 select와 출력으로 보내는 select가 모두 취소를 관찰합니다.
소비자가 읽지 않아도 어느 단계가 기다리든 빠져나올 경로가 있습니다.

### 9.3 defer의 순서를 분명히 쓰기

코드는 하나의 defer 함수 안에서 cancel과 Wait를 순서대로 호출합니다.
먼저 기다리고 나중에 취소하면
생산자가 더 보낼 상대를 기다리는 동안 서로 기다릴 수 있습니다.

```go
defer func() {
    cancel()
    group.Wait()
}()
```

같은 함수를 두 개의 defer 문으로 나눌 수도 있지만,
defer의 역순 실행을 놓치지 않도록 이 예제는 한 곳에서 순서를 보입니다.

### 9.4 0개를 받는 경계

`firstSquares(1000, 0)`도 유효합니다.
소비자는 아무 값을 받지 않고 바로 반환하려고 합니다.
defer가 취소 후 합류하므로 시작한 두 단계가 남지 않습니다.

상한 1000은 제곱 계산과 예제 자원 사용을 제한하는 설계입니다.
잘못된 count·take는 고루틴을 만들기 전에 오류로 반환합니다.

[Go 공식 파이프라인 설명](https://go.dev/blog/pipelines)은
조기 소비 중단과 송신자 누수를 같은 문제로 다룹니다.
이 교재는 Context와 WaitGroup으로 수명 관계를 명시합니다.

## 10. 채널을 사용할 때 확인할 여섯 질문

### 10.1 닫기와 취소를 혼동하지 않기

close는 생산자가 더 보낼 데이터가 없다는 약속입니다.
cancel은 호출자가 더 이상 진행을 원하지 않는다는 신호입니다.
두 신호는 방향도 의미도 다릅니다.

수신자가 out을 닫아 생산자를 멈추려고 하지 않습니다.
그 방식은 생산자를 정상 반환시키는 대신 send-on-closed-channel 패닉을 만들 수 있습니다.

### 10.2 작업자와 채널의 수명표

새 파이프라인을 만들 때 다음 표를 채워 봅니다.

| 질문 | 확인할 내용 |
| --- | --- |
| 누가 만든 채널인가? | 반환된 출력인지 외부 입력인지 |
| 누가 전송하는가? | 한 명인지 여러 명인지 |
| 누가 닫는가? | 마지막 전송을 확실히 아는 주체 |
| 어디에서 기다릴 수 있는가? | 수신·전송·하위 함수 각각 |
| 취소하면 어떻게 나오는가? | 각 대기의 Done 분기 |
| 누가 실제 종료를 확인하는가? | Wait·done·함수 반환 |

여섯 질문 중 하나라도 답이 없으면
예외 상황에서 작업이 남거나 채널이 두 번 닫힐 수 있습니다.

### 10.3 버퍼 크기는 오류를 숨기기 위한 값이 아니다

수신자가 조기에 끝나는 문제를 버퍼 1000으로 늘려 우연히 해결할 수 있습니다.
입력이 1001개가 되면 다시 멈출 수 있습니다.

버퍼 크기는 생산·소비 속도 차이를 얼마나 허용할지 정하는 정책입니다.
그 정책과 별개로 취소·종료 경로가 있어야 합니다.

완료 오류 하나를 보관하는 용량 1 채널처럼
프로토콜에서 최대 전송 수가 확실한 경우에는 그 수를 근거로 용량을 정할 수 있습니다.

### 10.4 채널이 모든 공유 데이터를 보호하지는 않는다

채널로 전달한 값이 포인터나 슬라이스이면 내부 데이터를 공유할 수 있습니다.
전송 뒤 송신자가 그 데이터를 더 이상 수정하지 않는 방식으로 소유권을 넘기거나,
독립된 복사본을 보내거나, 공유 접근을 잠그는 방법을 선택합니다.

다음 장에서는 여러 작업이 같은 상태를 직접 사용하는 경우를 다룹니다.
채널과 잠금 중 무엇이 무조건 더 좋은지 고르는 것이 아니라
값의 소유권과 작업 수명을 가장 분명히 표현하는 방법을 찾습니다.

### 10.5 select 표현식도 실행된다

select에 들어갈 때 채널 표현식과 전송할 값의 표현식은 평가됩니다.
선택되지 않은 전송 case라고 해서 전송값 계산이 무조건 생략되는 것은 아닙니다.

값을 만드는 함수가 비싸거나 외부 부작용을 갖는다면
case 표현식에 넣기 전에 평가 시점이 요구사항과 맞는지 확인해야 합니다.
[select의 실행 단계](https://go.dev/ref/spec#Select_statements)에 이 규칙이 있습니다.

## 11. 흔한 실수와 수정

| 증상 | 원인 | 수정 방향 |
| --- | --- | --- |
| 첫 전송에서 멈춤 | 버퍼 없는 채널에 같은 고루틴이 송수신 | 상대 실행 흐름 또는 필요한 버퍼 설계 |
| range가 끝나지 않음 | 생산자의 닫기 누락 | 마지막 전송 후 소유자가 닫기 |
| 영값이 계속 출력됨 | 닫힌 채널의 ok를 무시 | ok 확인 또는 range |
| select가 한 입력만 계속 읽음 | 닫힌 입력을 남겨 둠 | nil로 비활성화 |
| 조기 반환 뒤 고루틴이 남음 | 생산자의 전송에 취소 경로 없음 | 전송 select와 합류 추가 |
| send on closed channel | 소비자가 닫거나 여러 생산자가 각각 닫음 | 닫기 소유자 하나 지정 |
| CPU 사용률이 계속 높음 | default가 있는 빈 반복 | 필요한 대기 또는 실제 작업 |
| 출력 순서가 매번 다름 | 여러 통신이 함께 준비됨 | 결과 순서 요구를 별도로 설계 |
| 버퍼를 늘리면 잠깐만 개선됨 | 종료 프로토콜 누락 | 취소·종료·완료 확인을 수정 |

실패 예제를 일반 테스트에 영원히 멈추도록 넣지 않습니다.
안전한 구현을 테스트하고, 잘못된 구조는 문서의 발췌로 이유를 설명합니다.

## 12. 테스트와 연습문제

```sh
go test ./examples/19-channels-select/... ./solutions/19-channels-select/...
go vet ./examples/19-channels-select/... ./solutions/19-channels-select/...
go build ./examples/19-channels-select/... ./solutions/19-channels-select/...
```

Linux와 C 컴파일러가 준비된 환경:

```sh
go test -race ./examples/19-channels-select/... ./solutions/19-channels-select/...
```

테스트는 다음 계약을 확인합니다.

- 소비자가 더 받지 않아도 생산자가 취소 후 종료합니다.
- 입력이 없는 병합도 출력 채널을 닫습니다.
- idle 입력을 읽는 작업자도 취소로 종료합니다.
- 빈 결과와 중복값을 올바르게 처리합니다.
- 파이프라인이 0개만 소비해도 합류합니다.
- 잘못된 입력은 작업을 만들기 전에 거부합니다.

[연습문제](../../exercises/19-channels-select/README.md)는
양수 필터, 앞의 N개 수신, 여러 결과의 병합,
누수 구조 수정, 명시적인 완료 순서를 다룹니다.
작성 후 [해설](../../solutions/19-channels-select/README.md)과 비교합니다.

## 13. 요약과 자기 점검

- 채널은 값 전달과 기다림을 표현하며, 만들기만 해서는 고루틴이 생기지 않습니다.
- 닫힌 채널의 버퍼 값은 받을 수 있고 이후에는 영값·false가 즉시 반환됩니다.
- nil 채널의 통신은 진행하지 않으며 select에서 입력을 비활성화하는 데 쓸 수 있습니다.
- 닫기는 마지막 전송을 아는 주체 하나가 소유합니다.
- 조기 소비 중단에는 취소 가능한 전송과 실제 완료 대기가 필요합니다.
- select의 case 순서는 우선순위가 아니며 default는 기다림을 없앱니다.

다음을 자기 말로 설명해 보세요.

1. 버퍼 없는 채널에 보내는 코드가 언제 진행할 수 있나요?
2. 닫힌 채널에서 받은 0과 정상 데이터 0을 어떻게 구별하나요?
3. range 수신은 무엇을 보고 끝나나요?
4. 수신자가 채널을 직접 닫으면 어떤 송신과 충돌할 수 있나요?
5. 취소 case를 첫 줄에 썼다고 취소 우선권이 생기나요?
6. 파이프라인에서 전송·수신 중 어느 쪽에 취소를 넣어야 하나요?
7. 여러 생산자가 같은 출력을 보낼 때 누가 닫아야 하나요?
8. 입력값의 순서와 완료 순서를 각각 어떻게 표현할 수 있나요?

## 공식 참고 자료

공식 명세와 API를 2026-11-09에 확인했습니다.

- [언어 명세: 채널 타입](https://go.dev/ref/spec#Channel_types)
- [언어 명세: 전송](https://go.dev/ref/spec#Send_statements)
- [언어 명세: 수신 연산자](https://go.dev/ref/spec#Receive_operator)
- [언어 명세: close](https://go.dev/ref/spec#Close)
- [언어 명세: select](https://go.dev/ref/spec#Select_statements)
- [Go 메모리 모델](https://go.dev/ref/mem)
- [파이프라인과 취소](https://go.dev/blog/pipelines)
- [context](https://pkg.go.dev/context)
- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)

[이전: 18장 고루틴](../18-goroutines/README.md) ·
[목차](../README.md) ·
[다음: 20장 공유 상태와 동기화](../20-sync/README.md)
