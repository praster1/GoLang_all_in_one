# 20. 공유 상태와 동기화

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 응용 | 구조체·메서드·맵·고루틴·채널·테스트 | Go, 터미널, 이 저장소 |

여러 요청이 같은 재고를 줄이거나 같은 방문 횟수를 늘리는 상황을 생각해 봅시다.
모든 작업이 자기 결과 칸만 쓰도록 나누기 어려운 경우가 있습니다.
여러 작업이 **같은 상태를 함께 읽고 바꾸는 규칙**이 필요합니다.

공유 상태를 안전하게 다루려면 두 가지를 구분해야 합니다.
하나는 동기화 없이 같은 메모리에 접근하는 **데이터 레이스(data race)**이고,
다른 하나는 접근 자체는 안전하지만 업무의 조건이 깨지는 **논리 오류**입니다.
잠금이나 원자 연산을 사용했다는 사실만으로 둘 다 해결되지는 않습니다.

이 장은 작은 카운터에서 시작해 잔액의 합계, 재고, 읽기 스냅샷,
한 번만 초기화하는 설정, 제한된 예약을 구현합니다.
무엇을 함께 보호해야 하는지 먼저 정한 뒤 도구를 선택합니다.

## 학습 목표

- 데이터 레이스와 원자적인 업무 조건의 차이를 설명한다.
- Mutex로 관련 값의 검사와 변경을 하나의 보호 범위에 둔다.
- 읽기 잠금과 독립된 스냅샷의 의미를 구분한다.
- 원자 연산·CAS·Once가 제공하는 보장과 한계를 설명한다.
- 공유 상태를 줄이는 구조를 선택하고 race detector로 실제 경로를 검사한다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
예제는 외부 네트워크·계정·파일 없이 종료됩니다.
출력은 모든 작업을 기다린 뒤 표시합니다.

[18장](../18-goroutines/README.md)의 WaitGroup은 작업 완료를 기다렸습니다.
이번 장의 Mutex는 함께 접근하면 안 되는 상태를 보호합니다.
두 도구의 역할을 혼동하지 않습니다.

| 예제 | 중심 내용 | 실행 |
| --- | --- | --- |
| [mutex-counter](../../examples/20-sync/mutex-counter/main.go) | 같은 잠금으로 읽기·쓰기 보호 | `go run ./examples/20-sync/mutex-counter` |
| [ledger](../../examples/20-sync/ledger/main.go) | 검사와 두 값 변경의 불변 조건 | `go run ./examples/20-sync/ledger` |
| [rw-snapshot](../../examples/20-sync/rw-snapshot/main.go) | 읽기 잠금과 독립된 맵 복사 | `go run ./examples/20-sync/rw-snapshot` |
| [atomic-counter](../../examples/20-sync/atomic-counter/main.go) | 하나의 정수에 대한 원자 덧셈 | `go run ./examples/20-sync/atomic-counter` |
| [atomic-capacity](../../examples/20-sync/atomic-capacity/main.go) | 검사와 갱신을 CAS로 연결 | `go run ./examples/20-sync/atomic-capacity` |
| [once-config](../../examples/20-sync/once-config/main.go) | 설정 초기화의 결과 공유 | `go run ./examples/20-sync/once-config` |
| [channel-owner](../../examples/20-sync/channel-owner/main.go) | 한 소비자가 맵의 소유권 가짐 | `go run ./examples/20-sync/channel-owner` |
| [lock-scope](../../examples/20-sync/lock-scope/main.go) | 잠금 밖에서 콜백 실행 | `go run ./examples/20-sync/lock-scope` |

## 1. 무엇이 경쟁하고 있는가?

### 1.1 count++는 여러 단계를 포함한다

개념적으로 현재 값을 읽고, 1을 더하고, 새 값을 저장합니다.
여러 작업이 같은 변수를 동시에 수정하면 읽기와 쓰기가 겹칠 수 있습니다.

다음은 **의도적으로 잘못된 발췌**입니다.
기본 빌드·테스트 대상에는 넣지 않습니다.

```go
var count int
var group sync.WaitGroup
for range 2 {
    group.Go(func() {
        count++
    })
}
group.Wait()
```

마지막 Wait는 작업이 끝났음을 기다릴 뿐,
두 작업이 count를 수정하는 동안 서로 배타적으로 실행하게 만들지 않습니다.

잘못된 코드가 어떤 실행에서는 2를 출력할 수도 있습니다.
그 한 번의 결과는 접근이 안전하다는 증거가 아닙니다.
레이스는 결과값이 우연히 맞는 실행에서도 존재할 수 있습니다.

### 1.2 데이터 레이스의 범위

서로 다른 고루틴이 같은 메모리에 동기화 없이 접근하고,
그중 하나 이상이 쓰기이면 데이터 레이스가 발생할 수 있습니다.
둘 다 읽기만 하는 상태와는 구분합니다.

| 접근 관계 | 고려할 점 |
| --- | --- |
| 초기화 후 모두 읽기만 함 | 초기화의 공개 시점과 이후 변경 없음 |
| 같은 값을 여러 작업이 씀 | 배타적 보호 또는 소유권 분할 |
| 한 작업이 쓰고 다른 작업이 읽음 | 읽기도 같은 동기화 규칙에 포함 |
| 다른 배열 원소를 각각 전담 | 원소가 겹치지 않고 슬라이스 정보가 바뀌지 않음 |
| 별도 지역 변수 | 일반적으로 다른 작업과 공유하지 않음 |

[Go 메모리 모델](https://go.dev/ref/mem)과
[공식 race detector 안내](https://go.dev/doc/articles/race_detector)를 참고합니다.

### 1.3 불변 조건을 문장으로 적기

불변 조건(invariant)은 어떤 작업을 수행하더라도 유지해야 하는 관계입니다.
예를 들어 재고는 음수가 되면 안 되고,
두 계정 사이에서 옮긴 금액은 전체 합계를 바꾸면 안 됩니다.

먼저 다음 문장을 완성합니다.

> 다른 작업이 중간 상태를 보면 안 되는 범위는 무엇인가?

변수 하나인지, 여러 필드의 관계인지, 맵 전체인지에 따라
잠글 범위와 반환하는 값의 형태가 달라집니다.
도구 이름을 먼저 고르고 조건을 나중에 끼워 맞추지 않습니다.

## 2. Mutex로 카운터의 읽기와 쓰기 보호하기

`sync.Mutex`는 한 시점에 하나의 보호 구간만 진입하도록 합니다.
잠금은 Lock으로 얻고 Unlock으로 반환합니다.
영값은 잠겨 있지 않은 상태이므로 구조체 안에 바로 선언할 수 있습니다.

소스: [mutex-counter/main.go](../../examples/20-sync/mutex-counter/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Add(delta int) { c.mu.Lock(); defer c.mu.Unlock(); c.value += delta }
func (c *Counter) Value() int    { c.mu.Lock(); defer c.mu.Unlock(); return c.value }
func main() {
	var counter Counter
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			for range 100 {
				counter.Add(1)
			}
		})
	}
	group.Wait()
	fmt.Println("합계:", counter.Value())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/mutex-counter
```

예상 출력:

```text
합계: 400
```

### 2.1 읽기도 같은 잠금을 사용한다

Add만 잠그고 Value는 그냥 필드를 읽으면
Add의 쓰기와 Value의 읽기가 겹칠 수 있습니다.
공유 상태에 대한 모든 접근이 같은 보호 규칙에 참여해야 합니다.

현재는 쓰기와 읽기 모두 `c.mu`를 사용합니다.
다른 Mutex를 각각 사용하면 서로의 진입을 막지 못합니다.
잠금 객체를 공유하는 것과 보호하려는 데이터를 공유하는 것이 함께 맞아야 합니다.

### 2.2 보호된 구간과 함수 반환

Lock 다음에 `defer Unlock`을 두면
이후 함수가 어느 경로로 반환하더라도 잠금이 풀립니다.
잠금을 얻기 전에 Unlock을 등록하거나,
잠금을 얻지 못한 경로에서도 Unlock을 호출하지 않습니다.

Unlock은 지금 잠근 작업과 관계없는 임의의 정리 함수가 아닙니다.
잠금 상태가 아닌 Mutex를 해제하면 런타임 오류가 발생합니다.
[Mutex](https://pkg.go.dev/sync#Mutex)의 계약을 참고합니다.

### 2.3 포인터 수신자를 사용하는 이유

Counter를 값으로 복사하면 mu도 복사될 수 있습니다.
같은 데이터를 서로 다른 잠금으로 보호하거나,
이미 사용 중인 잠금 상태를 복제하는 잘못된 구조가 됩니다.

Mutex는 처음 사용한 뒤 복사하지 않습니다.
메서드는 `*Counter`로 받습니다.
다른 함수에 Counter를 값으로 넘기거나 사용 중인 구조체를 통째로 복사하지 않습니다.

`go vet`는 많은 복사 실수를 찾아 주지만 모든 설계 문제를 증명하는 도구는 아닙니다.
필드를 숨기고 메서드로만 접근하게 만드는 것도 도움이 됩니다.

### 2.4 이 Counter의 숫자 범위

예제는 4개 작업이 각각 100번 1을 더하므로 결과가 400입니다.
일반 목적의 무한 범위 카운터를 만든 것은 아닙니다.
실무에서 외부 delta를 받으면 부호·최댓값·오버플로 정책도 정해야 합니다.

## 3. 여러 값의 관계를 한 번에 보호하기

A와 B의 시작 잔액은 각각 100이며 전체 합계는 200입니다.
A에서 B로 옮길 때 출금 가능 여부를 확인하고 두 잔액을 같은 잠금 안에서 바꿉니다.

소스: [ledger/main.go](../../examples/20-sync/ledger/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"sync"
)

type Ledger struct {
	mu       sync.Mutex
	balances map[string]int
}

func newLedger() *Ledger { return &Ledger{balances: map[string]int{"A": 100, "B": 100}} }
func (l *Ledger) Transfer(from, to string, amount int) error {
	if amount <= 0 || amount > 100 {
		return errors.New("금액은 1~100이어야 합니다")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	source, exists := l.balances[from]
	if !exists {
		return errors.New("출금 계정이 없습니다")
	}
	destination, exists := l.balances[to]
	if !exists {
		return errors.New("입금 계정이 없습니다")
	}
	if from == to {
		return errors.New("같은 계정으로 옮길 수 없습니다")
	}
	if source < amount {
		return errors.New("잔액이 부족합니다")
	}
	l.balances[from] = source - amount
	l.balances[to] = destination + amount
	return nil
}
func (l *Ledger) Snapshot() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()
	copy := make(map[string]int, len(l.balances))
	for key, value := range l.balances {
		copy[key] = value
	}
	return copy
}
func main() {
	ledger := newLedger()
	failures := make([]error, 10)
	var group sync.WaitGroup
	for index := range failures {
		group.Go(func() { failures[index] = ledger.Transfer("A", "B", 1) })
	}
	group.Wait()
	snapshot := ledger.Snapshot()
	fmt.Printf("A=%d B=%d 총합=%d\n", snapshot["A"], snapshot["B"], snapshot["A"]+snapshot["B"])
	fmt.Println("모두 성공:", errors.Join(failures...) == nil)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/ledger
```

예상 출력:

```text
A=90 B=110 총합=200
모두 성공: true
```

### 3.1 검사와 변경은 한 묶음이다

먼저 잔액을 읽고 잠금을 풀었다가 나중에 차감하면,
그 사이 다른 요청이 같은 잔액을 사용할 수 있습니다.
잔액 확인과 실제 변경은 같은 보호 구간에 있어야 합니다.

| 단계 | 보호 범위 안인가? | 이유 |
| --- | --- | --- |
| amount 범위 검사 | 밖 | 호출의 독립된 값이며 공유 상태가 아님 |
| 계정 존재 확인 | 안 | 공유 맵의 내용 |
| 출금 잔액 확인 | 안 | 다른 이체가 변경할 수 있음 |
| 출금 계정 차감 | 안 | 입금과 한 작업으로 보이게 함 |
| 입금 계정 증가 | 안 | 전체 합계 유지 |
| Snapshot 복사 | 안 | 중간 상태를 보여 주지 않음 |

금액 검증을 밖에서 했다고 모든 검사를 밖에서 해도 되는 것은 아닙니다.
어떤 정보가 공유 상태에 의존하는지 구분합니다.

### 3.2 두 필드를 각각 원자적으로 바꾸면 충분할까?

각 잔액의 읽기·쓰기만 따로 안전하게 만들면 데이터 레이스는 피할 수 있습니다.
그러나 출금 후 입금 전의 중간 순간을 다른 조회가 볼 수 있습니다.
그 순간 전체 합계는 200이 아닐 수 있습니다.

여러 값의 관계를 함께 지키려면 그 관계를 하나의 보호 범위로 묶습니다.
이 예제는 Ledger 전체에 Mutex 하나를 둡니다.
복잡한 다중 잠금보다 소유권과 불변 조건을 읽기 쉽습니다.

### 3.3 실패는 아무것도 바꾸지 않는다

계정 없음, 같은 계정, 잘못된 금액, 잔액 부족을 모두 변경 전에 확인합니다.
오류 반환 전에 한쪽 잔액만 바꾸지 않습니다.
테스트도 실패 전후 Snapshot이 같은지 검사합니다.

현재 계정은 A와 B 두 개이며 총합 200은 보존됩니다.
일반 금융 시스템의 원장·감사·영속성·분산 트랜잭션을 구현한 예제는 아닙니다.
이 장에서는 메모리 안의 두 값에 대한 원자적인 처리 범위를 배웁니다.
[24장](../24-sql-transactions/README.md)은 DB 트랜잭션으로 이어집니다.

### 3.4 반환하는 맵도 보호해야 한다

Snapshot이 내부 balances를 그대로 반환하면
호출자가 잠금을 거치지 않고 맵을 수정할 수 있습니다.
따라서 잠금 안에서 새 맵으로 값을 복사합니다.

int 값은 복사하면 독립적입니다.
맵의 값이 슬라이스나 포인터라면 바깥 맵만 복사해도 내부 데이터는 공유할 수 있습니다.
필요한 깊이까지 복사하거나 변경 불가 계약을 정해야 합니다.

## 4. RWMutex와 읽기 스냅샷

읽기가 많고 쓰기가 적은 상태에는 `sync.RWMutex`를 고려할 수 있습니다.
여러 읽기 잠금은 함께 유지될 수 있지만,
쓰기 잠금은 읽기·쓰기와 겹치지 않도록 합니다.

소스: [rw-snapshot/main.go](../../examples/20-sync/rw-snapshot/main.go)

```go
package main

import (
	"fmt"
	"sort"
	"sync"
)

type Store struct {
	mu     sync.RWMutex
	values map[string]string
}

func (s *Store) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.values == nil {
		s.values = make(map[string]string)
	}
	s.values[key] = value
}
func (s *Store) Snapshot() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string, len(s.values))
	for key, value := range s.values {
		result[key] = value
	}
	return result
}
func main() {
	var store Store
	var group sync.WaitGroup
	group.Go(func() { store.Put("language", "Go") })
	group.Go(func() { store.Put("topic", "동기화") })
	group.Wait()
	snapshot := store.Snapshot()
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s=%s\n", key, snapshot[key])
	}
	snapshot["language"] = "changed"
	fmt.Println("원본 유지:", store.Snapshot()["language"])
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/rw-snapshot
```

예상 출력:

```text
language=Go
topic=동기화
원본 유지: Go
```

### 4.1 RLock은 읽기 허가다

Snapshot은 원본 맵을 읽기만 하므로 RLock을 사용합니다.
새 result 맵에 쓰는 것은 호출의 지역 자료를 만드는 작업입니다.
공유 원본 values를 변경하지 않습니다.

Put은 공유 맵을 수정하므로 Lock을 사용합니다.
RLock을 잡았다고 그 안에서 원본 맵에 써도 되는 것은 아닙니다.
여러 읽기 작업이 함께 들어올 수 있기 때문입니다.

### 4.2 잠금 밖에서도 안전한 결과

복사된 스냅샷은 원본 맵과 독립적입니다.
출력에서 복사본의 language를 changed로 바꿔도 원본은 Go로 남습니다.

Snapshot 반환 이후 원본에 새로운 Put이 실행될 수는 있습니다.
스냅샷은 특정 조회 시점의 값이며 이후 모든 변경을 자동으로 반영하지 않습니다.
정확히 어느 시점까지의 데이터가 필요한지 요구사항을 정해야 합니다.

### 4.3 읽기 잠금의 제한

RLock을 잡은 채 Lock으로 승격할 수 없습니다.
읽기 잠금을 재귀적으로 얻는 구조도 피해야 합니다.
쓰기 대기자가 생겼을 때 새로운 읽기 획득이 막히는 규칙과 충돌할 수 있습니다.

읽은 뒤 조건에 따라 쓰려면 Lock 하나로 검사와 수정을 묶거나,
잠금을 다시 얻은 뒤 조건을 재검사해야 합니다.
잠금을 풀고 다시 얻는 사이 상태가 바뀔 수 있음을 잊지 않습니다.
[RWMutex](https://pkg.go.dev/sync#RWMutex)의 규칙을 참고합니다.

### 4.4 Mutex보다 반드시 빠르지는 않다

읽기 비율이 높다는 이유만으로 RWMutex가 항상 이득인 것은 아닙니다.
잠금 구간의 길이, 경쟁 정도, 작업 수와 실행 환경에 따라 다릅니다.

먼저 단순한 Mutex로 정확한 계약을 만들고,
실제 측정이 필요한 상황에서 비교하는 방법도 좋습니다.
[26장](../26-performance/README.md)에서 측정 조건을 함께 기록합니다.

## 5. 하나의 정수 상태를 원자적으로 바꾸기

서로 독립적인 완료 개수를 세는 작업에는 원자 덧셈이 잘 맞습니다.
`atomic.Int64`의 영값은 0이며 Add와 Load를 사용합니다.

소스: [atomic-counter/main.go](../../examples/20-sync/atomic-counter/main.go)

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var counter atomic.Int64
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			for range 100 {
				counter.Add(1)
			}
		})
	}
	group.Wait()
	fmt.Println("완료 수:", counter.Load())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/atomic-counter
```

예상 출력:

```text
완료 수: 400
```

### 5.1 Add와 일반 ++의 차이

원자 Add는 하나의 갱신으로 수행됩니다.
서로 다른 작업의 증가가 같은 이전 값을 덮어쓰는 형태로 사라지지 않습니다.
현재 예제는 총 400번 증가하므로 최종 값은 400입니다.

atomic.Int64의 내부 값을 일반 변수처럼 직접 읽거나 쓰지 않습니다.
같은 상태에 대한 접근은 제공된 원자 메서드로 일관되게 수행합니다.
이 타입도 처음 사용한 뒤 복사하지 않습니다.
[sync/atomic](https://pkg.go.dev/sync/atomic)을 참고합니다.

### 5.2 Load와 Store를 따로 호출하면?

다음은 **논리적으로 잘못된 증가의 발췌**입니다.

```go
old := counter.Load()
counter.Store(old + 1)
```

두 호출 각각은 원자적이지만 둘을 합친 전체 연산은 아닙니다.
두 작업이 old=0을 각각 읽고 모두 1을 저장하면 증가 하나가 사라집니다.

| 순서 | 작업 A | 작업 B |
| --- | --- | --- |
| 1 | Load로 0 읽음 | |
| 2 | | Load로 0 읽음 |
| 3 | Store로 1 기록 | |
| 4 | | Store로 1 기록 |

이 경우 원자 API만 사용했으므로 race detector가 데이터 레이스를 보고하지 않을 수 있습니다.
그래도 결과는 기대한 2가 아니라 1입니다.
동기화의 안전성과 업무 연산의 원자성을 구분해야 하는 이유입니다.

### 5.3 여러 필드의 관계에는 별도 설계가 필요하다

완료 개수 하나를 늘리는 것은 원자 Add로 표현하기 좋습니다.
사용자 잔액과 거래 목록을 동시에 맞추는 일은 여러 값의 관계입니다.
각 필드를 따로 원자적으로 바꾸는 것만으로 전체의 일관성이 보장되지 않습니다.

원자 연산을 쓰면 무조건 잠금보다 빠르거나 좋은 구조가 되는 것도 아닙니다.
이해하기 쉬운 보호 범위와 관찰 가능한 계약을 먼저 택합니다.

## 6. CAS로 한도를 넘지 않게 예약하기

사용량을 읽고 한도를 확인한 뒤 증가시켜야 합니다.
중간에 다른 작업이 사용량을 바꿀 수 있으므로 단순 Load 다음 Store는 부족합니다.
CAS(compare-and-swap, 비교 후 교환)를 사용해 읽은 값이 그대로일 때만 변경합니다.

소스: [atomic-capacity/main.go](../../examples/20-sync/atomic-capacity/main.go)

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Quota struct {
	used  atomic.Int64
	limit int64
}

func (q *Quota) Reserve(amount int64) bool {
	if amount <= 0 || amount > q.limit {
		return false
	}
	for {
		old := q.used.Load()
		if old > q.limit-amount {
			return false
		}
		if q.used.CompareAndSwap(old, old+amount) {
			return true
		}
	}
}
func main() {
	quota := Quota{limit: 10}
	successes := make([]bool, 20)
	var group sync.WaitGroup
	for index := range successes {
		group.Go(func() { successes[index] = quota.Reserve(1) })
	}
	group.Wait()
	count := 0
	for _, success := range successes {
		if success {
			count++
		}
	}
	fmt.Println("예약 성공:", count)
	fmt.Println("사용량:", quota.used.Load())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/atomic-capacity
```

예상 출력:

```text
예약 성공: 10
사용량: 10
```

### 6.1 반복의 의미

1. 현재 사용량 old를 읽습니다.
2. 요청량을 더하면 한도를 넘는지 검사합니다.
3. 값이 여전히 old이면 old+amount로 바꿉니다.
4. 다른 작업이 먼저 바꿨다면 새 값을 읽고 다시 판단합니다.

CAS 실패는 프로그램 오류가 아닙니다.
다른 작업과 경쟁해 상태가 바뀌었다는 뜻입니다.
오래된 검사 결과를 그대로 사용하는 대신 다시 계산합니다.

### 6.2 뺄셈으로 검사하는 이유

`old + amount > limit`를 바로 계산하면
범위를 크게 확장했을 때 덧셈이 오버플로할 수 있습니다.
코드는 먼저 amount가 양수이며 limit 이내인지 검사합니다.
그 다음 `old > limit-amount`로 남은 공간을 비교합니다.

현재 예제의 limit은 생성 뒤 바뀌지 않습니다.
used는 0에서 시작하고 Reserve만 바꾸므로 0~limit 범위를 유지합니다.
limit 자체를 다른 고루틴에서 수정하면 별도의 공유 상태 문제가 생깁니다.

### 6.3 성공한 작업의 정체는 정해지지 않는다

20개 요청 중 정확히 10개가 성공하지만 어떤 인덱스가 성공할지는 정해지지 않습니다.
따라서 출력과 테스트는 성공 개수와 최종 사용량을 검사합니다.
처음 시작한 10개가 반드시 성공한다고 가정하지 않습니다.

이는 공정성이나 선착순 예약을 보장하는 큐가 아닙니다.
순서 요구가 있으면 접수 순서와 처리 순서를 따로 설계해야 합니다.

### 6.4 CAS를 어디까지 사용할까?

상태 하나로 간결하게 표현되는 조건에 사용합니다.
여러 필드가 함께 바뀌거나 수정 과정이 복잡해지면
Mutex 안에서 조건과 변경을 읽기 쉽게 표현하는 편이 나을 수 있습니다.

잠금과 원자 연산 중 하나를 모든 문제의 정답으로 고르지 않습니다.
코드를 검토하는 사람이 불변 조건을 설명할 수 있는지가 중요합니다.

## 7. 설정을 한 번만 읽고 결과 공유하기

여러 요청이 동시에 설정을 필요로 할 수 있습니다.
설정 읽기를 정확히 한 번 실행하고 이후 같은 결과를 제공하려면 Once 계열을 사용할 수 있습니다.
`sync.OnceValues`는 Go 1.21부터 제공되며 두 반환값을 가진 초기화를 감쌉니다.

소스: [once-config/main.go](../../examples/20-sync/once-config/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

type Config struct{ Workers int }

func main() {
	calls := 0
	load := sync.OnceValues(func() (Config, error) { calls++; return Config{Workers: 4}, nil })
	results := make([]Config, 3)
	failures := make([]error, 3)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() { results[index], failures[index] = load() })
	}
	group.Wait()
	fmt.Println("설정 읽기 횟수:", calls)
	fmt.Println("작업자 설정:", results[0].Workers, results[1].Workers, results[2].Workers)
	fmt.Println("오류 없음:", failures[0] == nil && failures[1] == nil && failures[2] == nil)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/once-config
```

예상 출력:

```text
설정 읽기 횟수: 1
작업자 설정: 4 4 4
오류 없음: true
```

### 7.1 한 번의 실행과 여러 호출

load는 여러 고루틴에서 동시에 호출될 수 있습니다.
실제 초기화 함수는 한 번 실행되고,
초기화가 끝난 결과를 모든 호출자가 받습니다.

calls를 여러 고루틴이 직접 증가시키는 구조가 아닙니다.
Once가 실행하는 한 초기화 함수만 calls에 쓰고,
main은 모든 호출자를 기다린 뒤 값을 읽습니다.

### 7.2 오류도 결과에 포함된다

OnceValues가 감싼 함수가 오류를 반환하면 그 오류도 이후 호출에 보존됩니다.
다음 호출에서 자동으로 다시 읽는 재시도 도구가 아닙니다.

일시적인 외부 장애 때문에 다시 시도해야 하는 초기화라면
재시도와 상태 전환을 직접 설계해야 합니다.
이번 예제는 프로세스 수명 동안 고정된 설정을 한 번 준비하는 경우에 맞습니다.

[OnceValues](https://pkg.go.dev/sync#OnceValues)와
[Once](https://pkg.go.dev/sync#Once)를 참고합니다.

### 7.3 반환값의 변경 가능성

예제 Config는 int 필드만 가진 값입니다.
호출자에게 구조체값을 반환하므로 독립된 값을 받습니다.

Config에 맵이나 슬라이스가 있다면 구조체만 복사해도 내부 자료를 공유할 수 있습니다.
초기화 이후 변경하지 않는 계약을 지키거나,
호출자에게 독립된 복사본을 제공해야 합니다.

Once는 초기화 시점의 안전성을 제공할 뿐,
반환한 객체의 모든 미래 변경을 보호하지 않습니다.

### 7.4 다시 읽는 설정과의 차이

서비스 실행 중 설정을 바꾸어야 한다면 Once만으로는 충분하지 않습니다.
새 설정을 검증한 뒤 잠금 아래에서 교체하거나,
변경하지 않는 스냅샷을 안전하게 교체하는 다른 구조가 필요합니다.

한 번만 초기화하는 요구인지, 계속 갱신하는 요구인지 먼저 확인합니다.

## 8. 공유 맵의 주인을 한 명으로 만들기

같은 맵에 잠금을 붙이는 대신 맵을 수정하는 주체를 하나로 제한할 수 있습니다.
여러 작업은 단어만 채널로 전달하고 main이 빈도 맵을 소유합니다.

소스: [channel-owner/main.go](../../examples/20-sync/channel-owner/main.go)

```go
package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	inputs := [][]string{{"go", "test"}, {"go", "한글"}, {"test", "go"}}
	words := make(chan string)
	var group sync.WaitGroup
	for _, part := range inputs {
		group.Go(func() {
			for _, word := range part {
				words <- word
			}
		})
	}
	closed := make(chan struct{})
	go func() { group.Wait(); close(words); close(closed) }()
	counts := map[string]int{}
	for word := range words {
		counts[word]++
	}
	<-closed
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, counts[key])
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/channel-owner
```

예상 출력:

```text
go: 3
test: 2
한글: 1
```

### 8.1 공유 쓰기 자체를 줄이기

counts에 접근하는 코드는 main의 수신 반복과 그 뒤 출력뿐입니다.
생산자들은 counts를 보지 않습니다.
따라서 counts에 Mutex를 붙일 필요가 없습니다.

이 방식이 항상 더 빠르다는 뜻은 아닙니다.
단어마다 채널 통신이 생기고 수집자가 처리량의 제한점이 될 수 있습니다.
장점은 맵의 소유권과 변경 위치가 분명하다는 점입니다.

### 8.2 닫기와 합류는 그대로 필요하다

생산자가 여러 명이므로 합류 담당자가 모든 생산자를 기다린 뒤 words를 닫습니다.
main은 range가 끝난 다음 담당자의 완료 신호도 확인합니다.
맵이 안전해도 고루틴 수명을 잊으면 전체 프로그램이 올바르지 않을 수 있습니다.

현재 입력은 작고 유한하며 소비자는 끝까지 읽습니다.
중간 소비 중단을 허용한다면 [19장](../19-channels-select/README.md)처럼
전송에 Context 취소를 추가해야 합니다.

### 8.3 또 다른 방법: 지역 맵을 합치기

작업마다 지역 빈도 맵을 만들고,
모든 작업을 기다린 뒤 호출자가 지역 맵을 합칠 수도 있습니다.
채널 통신 횟수를 줄이고 공유 쓰기도 피하는 구조입니다.

[연습문제](../../exercises/20-sync/README.md)의 지역 빈도 집계에서 구현합니다.
큰 결과를 모두 보관해야 하는 비용과 단어마다 통신하는 비용 중
어떤 선택이 실제 입력에 맞는지 측정할 수 있습니다.

## 9. 잠금 밖에서 외부 작업 실행하기

잠금을 잡은 채 임의의 콜백을 호출하면
콜백이 같은 객체를 다시 사용하거나 오래 기다릴 수 있습니다.
먼저 필요한 값을 복사한 뒤 잠금을 풀고 콜백을 실행합니다.

소스: [lock-scope/main.go](../../examples/20-sync/lock-scope/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Add()       { c.mu.Lock(); defer c.mu.Unlock(); c.value++ }
func (c *Counter) Value() int { c.mu.Lock(); defer c.mu.Unlock(); return c.value }
func (c *Counter) Report(render func(int)) {
	snapshot := c.Value()
	// The callback can be slow or call Counter again, so invoke it after unlocking.
	render(snapshot)
}
func main() {
	var counter Counter
	counter.Add()
	counter.Report(func(value int) { counter.Add(); fmt.Println("보고서 기준 값:", value) })
	fmt.Println("현재 값:", counter.Value())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/20-sync/lock-scope
```

예상 출력:

```text
보고서 기준 값: 1
현재 값: 2
```

### 9.1 재진입을 생각하기

Report의 콜백은 Counter.Add를 다시 호출합니다.
Report가 같은 Mutex를 계속 잡고 있었다면 Add는 잠금을 기다립니다.
Report는 콜백이 끝나기를 기다리므로 서로 진행하지 못할 수 있습니다.

Mutex는 같은 고루틴이 다시 Lock해도 알아서 허용하는 재진입 잠금이 아닙니다.
잠근 상태에서 같은 객체의 다른 잠금 메서드를 호출하는 구조를 주의합니다.

### 9.2 보고서 시점과 현재 시점

보고서에는 복사 시점의 값 1이 전달됩니다.
콜백이 Add한 뒤 현재 값은 2입니다.
둘이 다르다는 사실은 오류가 아니라 Snapshot 계약의 결과입니다.

보고서 전체를 쓰는 동안 값이 변하지 않아야 하는 요구라면
스냅샷을 하나의 보고서 기준으로 삼거나 다른 일관성 정책을 정해야 합니다.
외부 작업이 끝날 때까지 잠금으로 모든 변경을 막는 선택에는 대기 비용이 있습니다.

### 9.3 잠금 안에서 오래 기다리지 않기

파일 쓰기, 네트워크 호출, 사용자 콜백 같은 작업은
실행 시간과 재진입 여부를 잠금 소유자가 통제하기 어렵습니다.
가능하면 필요한 자료를 잠금 안에서 복사하고 느린 작업은 밖으로 옮깁니다.

잠금 해제 후 다시 상태를 바꾸어야 한다면
그 사이 값이 달라졌을 가능성을 검사해야 합니다.
단순히 잠금 범위를 작게 자르는 것만으로 원래 불변 조건이 유지되지는 않습니다.

Report의 render는 nil이 아닌 콜백이라는 계약입니다.
외부 입력으로 콜백이 생기는 공개 API라면 nil 처리 정책도 정해야 합니다.

## 10. 교착 상태와 잠금 범위

### 10.1 두 잠금의 순서

작업 A가 첫 잠금을 잡고 둘째 잠금을 기다리는데,
작업 B가 둘째 잠금을 잡고 첫 잠금을 기다리면 둘 다 진행하지 못합니다.
이런 상황을 **교착 상태(deadlock)**라고 부릅니다.

| 작업 A | 작업 B |
| --- | --- |
| 첫 잠금 획득 | 둘째 잠금 획득 |
| 둘째 잠금을 기다림 | 첫 잠금을 기다림 |

해결책은 필요한 상태를 하나의 잠금으로 보호하거나,
모든 코드에서 잠금 획득 순서를 동일하게 정하는 것입니다.
잠금 수를 늘리기 전에 보호할 상태의 관계부터 확인합니다.

### 10.2 반복 안의 defer

defer는 반복 한 회차의 끝이 아니라 **함수의 끝**에 실행됩니다.
다음은 **의도적으로 잘못된 발췌**입니다.

```go
for _, value := range values {
    mu.Lock()
    defer mu.Unlock()
    _ = value
}
```

첫 회차에서 잠근 뒤 함수가 아직 반환하지 않았으므로 잠금이 유지됩니다.
다음 회차의 Lock이 다시 기다리게 될 수 있습니다.
한 회차를 작은 함수로 분리하거나 필요한 지점에서 명시적으로 Unlock합니다.

정리 코드를 간결하게 만드는 것보다 잠금의 수명을 정확히 표현하는 것이 우선입니다.

### 10.3 Context는 Lock을 강제로 취소하지 않는다

Mutex.Lock은 Context 인자를 받지 않습니다.
호출자가 취소되었다고 이미 Lock에서 기다리는 코드가 자동으로 반환하지 않습니다.

새 고루틴에서 Lock을 호출하고 호출자만 먼저 반환하면
나중에 잠금을 얻는 고루틴이 남을 수 있습니다.
그렇게 취소 기능을 흉내 내지 않습니다.

짧은 임계 구간을 유지하고,
취소 가능한 작업 접수·대기는 채널이나 다른 명시적인 수명 구조로 분리할 수 있습니다.
잠금을 얻은 뒤 실행할 코드의 종료 조건도 분명해야 합니다.

## 11. 도구 선택과 한계

| 요구 | 우선 생각할 구조 | 주의할 점 |
| --- | --- | --- |
| 각 작업의 독립된 결과 | 인덱스 소유권 분할 | 슬라이스 정보 변경 금지 |
| 여러 필드의 관계 | Mutex | 검사와 변경을 같은 구간에 |
| 읽기 스냅샷이 많음 | Mutex 또는 RWMutex | 실제 이득은 측정 |
| 독립된 숫자 증가 | atomic Add | 복합 조건은 별도 |
| 단일 값의 조건부 변경 | CAS | 실패 시 새 값으로 재판단 |
| 한 번 초기화 | Once·OnceValues | 실패도 다시 시도하지 않을 수 있음 |
| 데이터 흐름으로 소유권 전달 | 채널 | 닫기·취소·합류 설계 |
| 작업별 집계 | 지역 상태 뒤 결합 | 중간 결과 메모리 사용 |

### 11.1 sync.Map을 기본 맵의 대체로 사용하지 않기

sync.Map은 특정 접근 패턴을 위한 동시성 자료구조입니다.
일반적인 강한 타입의 map과 Mutex가 더 읽기 쉬운 경우가 많습니다.
여러 키에 걸친 불변 조건을 자동으로 지켜 주는 트랜잭션 맵도 아닙니다.
[sync.Map](https://pkg.go.dev/sync#Map)의 용도를 확인한 뒤 선택합니다.

### 11.2 sync.Pool은 자원 개수 제한 도구가 아니다

Pool은 재사용 가능한 임시 객체의 할당을 줄이는 용도입니다.
넣은 객체가 계속 남는다는 보장이 없으며,
DB 커넥션처럼 반드시 관리해야 하는 자원 수명과 다릅니다.

작업자 수나 동시 요청 수를 제한하려면 그 개수를 명시적으로 관리합니다.
Pool의 이름만 보고 “작업자 풀”과 같은 것으로 사용하지 않습니다.
[sync.Pool](https://pkg.go.dev/sync#Pool)을 참고합니다.

## 12. 레이스 검사를 정확히 해석하기

### 12.1 실행 명령

```sh
go test ./examples/20-sync/... ./solutions/20-sync/...
go vet ./examples/20-sync/... ./solutions/20-sync/...
go build ./examples/20-sync/... ./solutions/20-sync/...
```

Linux와 C 컴파일러가 준비된 환경:

```sh
go test -race ./examples/20-sync/... ./solutions/20-sync/...
```

race detector는 실행 중 관찰한 메모리 접근 관계를 검사합니다.
일반 테스트와 함께 실행하며, 지원되는 OS·아키텍처와 빌드 도구가 필요합니다.
이 교재의 CI는 지원하는 Linux 환경에서 별도로 실행합니다.

### 12.2 성공한 검사의 의미

실행한 경로에서 감지된 데이터 레이스가 없었다는 뜻입니다.
테스트하지 않은 분기까지 모두 안전하다는 증명은 아닙니다.
논리적 한도 초과나 잘못된 순서처럼 데이터 레이스가 아닌 문제도 남을 수 있습니다.

따라서 다음을 함께 확인합니다.

1. 누가 공유 데이터를 읽고 쓰는가?
2. 어떤 잠금이나 원자 연산이 그 접근을 연결하는가?
3. 여러 값 사이에 반드시 유지할 조건은 무엇인가?
4. 경계와 실패에서 상태가 부분적으로 변하지 않는가?
5. 호출자에게 반환한 값으로 내부 상태를 우회 수정할 수 없는가?

### 12.3 예제 테스트가 검증하는 것

| 테스트 | 관찰하는 계약 |
| --- | --- |
| Counter 동시 읽기·쓰기 | 최종 증가 수와 동시 안전성 |
| Ledger 동시 이체 | 전체 합계 200과 음수가 아닌 잔액 |
| Ledger 실패 | 어떤 잔액도 부분 변경되지 않음 |
| Store Snapshot | 반환 맵 변경이 원본에 영향을 주지 않음 |
| Quota 경쟁 | 한도 초과 없음, 성공한 예약 수와 사용량 일치 |
| 지역 빈도 집계 | 입력 정규화와 합계 |
| Loader 경쟁 | 초기화 함수 한 번과 동일한 실패 반환 |
| Slots 반복 반납 | 중복 반납이 사용량을 두 번 줄이지 않음 |

동시성 테스트에서도 오류 출력과 자료 기록이 안전해야 합니다.
테스트만을 위해 만든 공유 append나 증가 변수가 레이스를 만들 수 있습니다.

## 13. 자주 하는 실수

| 실수 | 왜 문제인가? | 수정 방향 |
| --- | --- | --- |
| 쓰기만 잠그고 읽기 생략 | 쓰기·읽기 경쟁 | 모든 접근에 같은 규칙 |
| 서로 다른 Mutex 사용 | 보호 구간이 연결되지 않음 | 상태와 잠금 소유자 함께 묶기 |
| 값 수신자로 Mutex 복사 | 잠금 상태를 복제 | 포인터 수신자·비복사 구조 |
| 잔액 검사 후 잠금 해제 | 조건과 변경 사이 상태 변화 | 같은 보호 구간에서 검사·변경 |
| 잠금 안에서 콜백 호출 | 재진입·장시간 대기 | 스냅샷 뒤 잠금 밖에서 호출 |
| atomic Load 뒤 Store로 증가 | 두 호출의 묶음은 원자적이지 않음 | Add 또는 CAS |
| 내부 맵을 그대로 반환 | 잠금 우회 접근 | 필요한 깊이의 복사 |
| Once로 일시 장애 재시도 기대 | 한 번의 결과를 보존 | 별도 재시도 상태 설계 |
| 레이스 통과만으로 완료 판단 | 논리 조건은 검사하지 않음 | 불변 조건 테스트 |

## 14. 연습문제와 자기 점검

[연습문제](../../exercises/20-sync/README.md)에서
재고 예약, 지역 빈도 집계, 실패를 캐시하는 설정 읽기,
중복 반납을 막는 슬롯 제한기를 구현합니다.
작성 후 [해설](../../solutions/20-sync/README.md)과 비교합니다.

다음 질문에 답해 보세요.

1. WaitGroup과 Mutex는 각각 무엇을 기다리거나 보호하나요?
2. 읽기만 하는 메서드도 쓰기와 같은 잠금이 필요한 이유는 무엇인가요?
3. atomic Load와 Store를 따로 쓴 증가가 왜 틀릴 수 있나요?
4. 두 잔액을 각각 보호하는 것과 전체 합계를 보호하는 것은 어떻게 다른가요?
5. Snapshot의 바깥 맵만 복사해도 항상 독립된 자료가 되나요?
6. OnceValues가 오류를 반환하면 다음 호출에서 자동 재시도하나요?
7. 잠금 밖에서 보고서를 만들 때 어느 시점의 자료를 기준으로 하나요?
8. race detector 통과 뒤에도 어떤 논리 테스트가 필요한가요?

## 요약

공유 상태를 다룰 때는 먼저 유지해야 할 조건을 정합니다.
관련 값의 검사와 변경은 같은 보호 범위에 두고,
가능하면 공유 쓰기 자체를 줄입니다.

Mutex, RWMutex, 원자 연산, Once와 채널은 서로 다른 계약을 표현합니다.
도구가 제공하는 보장과 프로그램이 지켜야 할 조건을 구분하면
테스트해야 할 경계도 더 분명해집니다.

## 공식 참고 자료

공식 API와 버전 표기를 2026-11-09에 확인했습니다.

- [Go 메모리 모델](https://go.dev/ref/mem)
- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [sync.Mutex](https://pkg.go.dev/sync#Mutex)
- [sync.RWMutex](https://pkg.go.dev/sync#RWMutex)
- [sync.Once](https://pkg.go.dev/sync#Once)
- [sync.OnceValues](https://pkg.go.dev/sync#OnceValues)
- [sync/atomic](https://pkg.go.dev/sync/atomic)
- [sync.Map](https://pkg.go.dev/sync#Map)
- [sync.Pool](https://pkg.go.dev/sync#Pool)

[이전: 19장 채널과 select](../19-channels-select/README.md) ·
[목차](../README.md) ·
[다음: 21장 HTTP 클라이언트](../21-http-client/README.md)
