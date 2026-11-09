# 20장 해설 — 공유 상태와 동기화

[연습문제](../../exercises/20-sync/README.md)의 참고 구현입니다.
모든 명령은 저장소 루트에서 실행합니다.

잠금이나 원자 타입의 이름만으로 정답을 판단하지 않습니다.
공유 상태, 보호 범위, 성공·실패의 변화, 반환 자료의 소유권을 확인합니다.

## 문제 1. 각각의 주장 검토

질문의 여덟 주장은 모두 그대로는 맞지 않습니다.

1. Wait는 작업 종료를 기다리며 작업 중의 count++를 보호하지 않습니다.
2. 읽기도 쓰기와 경쟁할 수 있어 같은 보호 규칙이 필요합니다.
3. 원자 메서드 여러 개를 조합한 업무 연산은 자동으로 원자적이지 않습니다.
4. 사용한 Mutex를 복사하면 안 됩니다.
5. RLock은 여러 읽기를 허용하므로 그 안에서 원본을 수정하면 안 됩니다.
6. 맵 값이 포인터이면 바깥 맵만 복사해도 대상 자료를 공유할 수 있습니다.
7. OnceValues는 한 번의 결과를 보존하며 오류도 결과에 포함됩니다.
8. 데이터 레이스가 없어도 한도 검사·이체 순서 같은 논리 오류가 남을 수 있습니다.

Load→Store 증가에서 두 작업이 모두 0을 읽고 각각 1을 쓰면 최종 값은 1입니다.
각 메서드는 안전하지만 전체 증가 두 번이 보존되지 않았습니다.
Add를 사용하거나 검사와 변경을 CAS 또는 잠금으로 묶어야 합니다.

## 참고 구현

| 문제 | 소스 | 보호하려는 조건 |
| --- | --- | --- |
| 2 | [inventory](inventory/main.go) | 재고가 음수가 되지 않음 |
| 3 | [local-frequency](local-frequency/main.go) | 지역 맵만 쓰고 합류 후 결합 |
| 4 | [cached-loader](cached-loader/main.go) | 초기화 한 번, 결과와 오류의 공개 |
| 5 | [atomic-slots](atomic-slots/main.go) | 사용량 한도와 정확히 한 번의 반납 |

## 문제 2. 검사와 차감을 하나로 보호하기


Reserve의 수량 범위는 호출의 독립된 입력이므로 잠금 전에 확인합니다.
실제 stock의 확인과 차감은 같은 Mutex 안에서 수행합니다.
Remaining도 같은 Mutex를 사용해 실행 중인 변경과 읽기를 정렬합니다.

소스: [inventory/main.go](inventory/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type Inventory struct {
	mu    sync.Mutex
	stock int
}

func newInventory(stock int) (*Inventory, error) {
	if stock < 0 || stock > 10000 {
		return nil, errors.New("재고는 0~10000이어야 합니다")
	}
	return &Inventory{stock: stock}, nil
}
func (i *Inventory) Reserve(quantity int) bool {
	if quantity < 1 || quantity > 100 {
		return false
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stock < quantity {
		return false
	}
	i.stock -= quantity
	return true
}
func (i *Inventory) Remaining() int { i.mu.Lock(); defer i.mu.Unlock(); return i.stock }
func main() {
	inventory, err := newInventory(10)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	results := make([]bool, 20)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() { results[index] = inventory.Reserve(1) })
	}
	group.Wait()
	successes := 0
	for _, ok := range results {
		if ok {
			successes++
		}
	}
	fmt.Println("예약 성공:", successes)
	fmt.Println("남은 재고:", inventory.Remaining())
}
```

```sh
go run ./solutions/20-sync/inventory
```

예상 출력:

```text
예약 성공: 10
남은 재고: 0
```


### 실패 경로의 상태

잘못된 수량이나 재고 부족에서는 stock을 바꾸지 않습니다.
검증 실패 뒤 원래 값을 복구하는 방식보다
모든 조건을 확인한 뒤 한 번 바꾸는 방식이 읽기 쉽습니다.

재고 1000에서 7개씩 예약하면 최대 142번 성공합니다.
142*7=994이므로 남은 재고는 6입니다.
충분히 많은 예약 시도를 실행하면 어느 작업이 성공하든 이 수치는 같습니다.

실패한 작업의 인덱스나 스케줄링 순서는 정답으로 고정하지 않습니다.
계약은 전체 성공 수와 최종 재고에 있습니다.

### 값 복사 방지

Inventory의 mu와 stock은 한 객체에 묶입니다.
메서드 수신자를 포인터로 두고 사용 후 구조체를 복사하지 않습니다.
stock을 공개 필드로 바꾸어 호출자가 직접 수정하게 만들면 보호 규칙을 우회할 수 있습니다.

## 문제 3. 공유 쓰기 없이 합계 결합하기


각 작업은 local이라는 자기 맵에만 씁니다.
완성한 맵은 자기 partials[index]에 넣습니다.
호출자는 Wait 뒤에서 모든 지역 맵을 읽고 최종 결과 맵에 합칩니다.

소스: [local-frequency/main.go](local-frequency/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

func countWords(parts [][]string) (map[string]int, error) {
	if len(parts) > 8 {
		return nil, errors.New("묶음은 8개까지 허용합니다")
	}
	total := 0
	for _, part := range parts {
		if len(part) > 1000-total {
			return nil, errors.New("전체 단어는 1000개까지 허용합니다")
		}
		total += len(part)
		for _, word := range part {
			if len(word) > 64 || !utf8.ValidString(word) || strings.TrimSpace(word) == "" {
				return nil, errors.New("단어는 공백이 아닌 유효한 UTF-8, 64바이트 이하여야 합니다")
			}
		}
	}
	partials := make([]map[string]int, len(parts))
	var group sync.WaitGroup
	for index, part := range parts {
		group.Go(func() {
			local := map[string]int{}
			for _, word := range part {
				local[strings.ToLower(strings.TrimSpace(word))]++
			}
			partials[index] = local
		})
	}
	group.Wait()
	result := map[string]int{}
	for _, local := range partials {
		for word, count := range local {
			result[word] += count
		}
	}
	return result, nil
}
func main() {
	counts, err := countWords([][]string{{"Go", "test"}, {"go", "한글"}, {" test ", "GO"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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

```sh
go run ./solutions/20-sync/local-frequency
```

예상 출력:

```text
go: 3
test: 2
한글: 1
```


### 잠금이 없는 이유

부분 맵은 하나의 고루틴만 변경합니다.
partials 슬라이스의 각 원소도 하나의 고루틴이 담당합니다.
최종 result는 모든 작업이 끝난 뒤 호출자만 변경합니다.

같은 종류의 map을 사용하더라도 소유권을 나누면
공유 쓰기 경쟁 자체가 생기지 않을 수 있습니다.
이것이 map을 특별한 동시성 컨테이너로 바꾼 것과 다른 설계입니다.

### 입력 검증과 정규화

원문 바이트 길이 64, 유효한 UTF-8, 공백 제거 뒤 비어 있지 않음을 먼저 확인합니다.
작업 중에는 TrimSpace와 ToLower로 정규화합니다.
Go·go·GO는 같은 키가 됩니다.

이 소문자 변환이 모든 언어의 검색·동등성 규칙을 완성하는 것은 아닙니다.
여기서는 지정된 정규화 함수의 결과를 같은 키로 집계합니다.
Unicode 정규화가 별도로 필요한 상황은 [07장](../../book/07-strings-unicode/README.md)과 연결합니다.

### 메모리와 통신의 선택

이 방식은 작업별 맵을 모두 보관합니다.
단어마다 채널로 보내는 방식보다 통신은 줄지만 중간 자료의 메모리가 필요합니다.
실제 성능 판단은 입력 규모와 측정 조건을 확인한 뒤 합니다.

## 문제 4. 한 번의 실패도 보존하기


Loader는 실제 읽기 함수, Once, 결과, 오류를 한 객체에 묶습니다.
Get은 Once.Do가 끝난 뒤 필드를 읽습니다.
초기화 중 다른 호출자도 Get으로 들어올 수 있지만 초기화 결과가 준비될 때까지 기다립니다.

소스: [cached-loader/main.go](cached-loader/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type Config struct{ Workers int }
type Loader struct {
	once  sync.Once
	load  func() (Config, error)
	value Config
	err   error
}

func newLoader(load func() (Config, error)) (*Loader, error) {
	if load == nil {
		return nil, errors.New("읽기 함수가 필요합니다")
	}
	return &Loader{load: load}, nil
}
func (l *Loader) Get() (Config, error) {
	l.once.Do(func() {
		l.value, l.err = l.load()
		if l.err == nil && (l.value.Workers < 1 || l.value.Workers > 32) {
			l.err = errors.New("작업자 수는 1~32여야 합니다")
		}
		if l.err != nil {
			l.value = Config{}
		}
	})
	return l.value, l.err
}
func main() {
	calls := 0
	sourceError := errors.New("설정을 읽을 수 없습니다")
	loader, err := newLoader(func() (Config, error) { calls++; return Config{}, sourceError })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	failures := make([]error, 3)
	var group sync.WaitGroup
	for index := range failures {
		group.Go(func() { _, failures[index] = loader.Get() })
	}
	group.Wait()
	fmt.Println("읽기 시도:", calls)
	fmt.Println("같은 오류 보존:", errors.Is(failures[0], sourceError) && errors.Is(failures[1], sourceError) && errors.Is(failures[2], sourceError))
}
```

```sh
go run ./solutions/20-sync/cached-loader
```

예상 출력:

```text
읽기 시도: 1
같은 오류 보존: true
```


### 오류의 보존

load가 원래 오류를 반환하면 그대로 저장합니다.
호출자는 errors.Is로 같은 원인을 확인할 수 있습니다.
읽기는 성공했지만 Workers 범위가 잘못되면 검증 오류를 저장합니다.
어떤 실패든 Config 영값을 반환해 잘못된 설정을 실수로 사용하지 않게 합니다.

같은 Loader의 다음 Get은 실제 load를 다시 호출하지 않습니다.
시작 시 설정 오류로 서비스를 중단하는 정책에는 이 동작이 어울릴 수 있습니다.
실행 중 회복을 기대하는 초기화에는 별도 재시도 구조가 필요합니다.

### 반환값의 소유권

현재 Config에는 int 필드 하나만 있으므로 값 복사로 독립적인 결과를 받습니다.
앞으로 맵·슬라이스 필드가 추가되면 공유 변경 가능성을 다시 검토해야 합니다.
Once가 그 자료의 미래 변경까지 잠그는 것은 아닙니다.

### 사용 전제

생성자로 만든 Loader를 사용하고 사용 후 복사하지 않습니다.
실제 load는 예상 가능한 실패를 error로 반환해야 합니다.
패닉을 일반적인 입력 오류 처리로 사용하지 않습니다.

## 문제 5. 획득마다 독립된 반납 책임 만들기


TryAcquire는 현재 사용량을 읽고 한도보다 작을 때만 CAS로 증가시킵니다.
성공한 획득마다 새로운 Once를 가진 반납 함수를 만듭니다.
같은 반납 함수를 여러 번 불러도 그 획득의 사용량은 한 번만 줄어듭니다.

소스: [atomic-slots/main.go](atomic-slots/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type Slots struct {
	limit  int64
	active atomic.Int64
}

func newSlots(limit int64) (*Slots, error) {
	if limit < 1 || limit > 32 {
		return nil, errors.New("슬롯 수는 1~32여야 합니다")
	}
	return &Slots{limit: limit}, nil
}
func (s *Slots) TryAcquire() (func(), bool) {
	for {
		old := s.active.Load()
		if old >= s.limit {
			return nil, false
		}
		if s.active.CompareAndSwap(old, old+1) {
			var once sync.Once
			release := func() { once.Do(func() { s.active.Add(-1) }) }
			return release, true
		}
	}
}
func main() {
	slots, err := newSlots(2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	first, ok1 := slots.TryAcquire()
	second, ok2 := slots.TryAcquire()
	_, ok3 := slots.TryAcquire()
	fmt.Println("처음 세 요청:", ok1, ok2, ok3)
	first()
	first()
	third, ok4 := slots.TryAcquire()
	fmt.Println("반납 후 획득:", ok4)
	second()
	third()
	fmt.Println("남은 사용량:", slots.active.Load())
}
```

```sh
go run ./solutions/20-sync/atomic-slots
```

예상 출력:

```text
처음 세 요청: true true false
반납 후 획득: true
남은 사용량: 0
```


### Once의 위치가 중요하다

Slots 전체에 Once 하나를 두면 전체 프로그램에서 첫 반납 하나만 실행됩니다.
그러면 나머지 획득은 반납되지 않아 슬롯이 소진됩니다.

참고 구현은 **각 성공한 획득의 클로저**에 Once를 둡니다.
반납 책임의 단위와 중복 방지 상태의 단위가 같아야 합니다.

### 한도와 음수 방지

획득은 old가 limit보다 작을 때만 1을 더합니다.
반납은 실제 성공한 획득에서 만들어진 함수만 실행합니다.
그리고 그 함수는 정확히 한 번만 감소합니다.
따라서 정상 사용에서는 active가 음수가 되거나 limit을 넘지 않습니다.

반납 함수를 외부에 전달할 때 누가 호출할지 정해야 합니다.
중복 호출은 막지만 아예 호출하지 않는 누락은 자동으로 해결하지 않습니다.
성공 직후 `defer release()`를 등록하면 많은 반환 경로를 함께 처리할 수 있습니다.

### 기다림과 공정성

TryAcquire는 공간이 없으면 즉시 false를 반환합니다.
입장 순서나 먼저 기다린 요청의 우선권을 제공하지 않습니다.
바쁜 무한 재시도 루프를 호출자에 넣으면 CPU를 낭비할 수 있습니다.

대기 가능한 제한기가 필요하다면 Context와 채널을 사용한 접수 구조를 별도로 설계합니다.
현재 문제의 목표는 비차단 획득과 안전한 반납입니다.

## 문제 6. 보호 범위 검토 답안

### 6.1 잔액 확인과 차감의 분리

잔액을 확인한 뒤 잠금을 풀면 다른 작업이 그 잔액을 사용할 수 있습니다.
다시 잠갔을 때 이전 검사 결과는 더 이상 유효하지 않을 수 있습니다.
검사와 차감을 같은 보호 구간에 두거나 재획득 뒤 반드시 재검사합니다.

원자적인 이체라면 출금과 입금도 함께 보호합니다.
실패 테스트에서 어느 한쪽만 바뀌지 않는지 확인합니다.

### 6.2 내부 맵 반환

Snapshot이 원본 맵을 반환하면 호출자가 잠금 없이 읽고 쓸 수 있습니다.
잠금 안에서 새 맵으로 복사하고 독립된 결과를 반환합니다.
값에 포인터·슬라이스가 들어 있다면 필요한 복사 깊이를 다시 정합니다.

테스트에서는 반환 맵을 바꾼 뒤 새 Snapshot을 받아 원본이 유지되는지 확인합니다.

### 6.3 콜백의 재진입

Report가 잠금을 잡고 render를 호출하고,
render가 같은 객체의 Add를 호출하면 Add가 같은 잠금을 기다릴 수 있습니다.
Report는 render가 끝나야 잠금을 풀 수 있어 교착이 생깁니다.

잠금 안에서 값만 복사하고 render는 밖에서 실행합니다.
그 보고서의 기준은 복사 시점이라는 계약을 문서화합니다.

### 6.4 반복 안의 defer

defer는 반복 회차가 아니라 함수 반환에 연결됩니다.
첫 회차의 잠금을 유지한 채 다음 Lock을 시도할 수 있습니다.
한 회차를 작은 함수로 나누거나 모든 경로에서 명시적으로 Unlock합니다.

### 6.5 잠금 순서의 역전

A를 잡고 B를 기다리는 작업과 B를 잡고 A를 기다리는 작업은 서로 막을 수 있습니다.
한 잠금으로 묶거나 모든 코드에서 동일한 획득 순서를 유지합니다.
두 자원의 식별자를 기준으로 항상 작은 것부터 잠그는 정책도 한 방법입니다.

### 6.6 Context로 Lock 감싸기

Lock 자체는 Context를 받지 않습니다.
별도 고루틴을 만들고 호출자만 먼저 반환하면
나중에 잠금을 얻는 고루틴의 책임이 남습니다.

임계 구간을 짧고 유한하게 만들고,
취소 가능한 작업 대기는 잠금 획득과 별도의 계층에서 설계합니다.
호출자 반환 뒤 자원을 소유한 고루틴을 방치하지 않습니다.

## 문제 7. 선택 과제의 평가

같은 데이터와 같은 처리량에서 비교했는지 확인합니다.
읽기 비율만 보고 RWMutex가 반드시 더 빠르다고 결론 내리지 않습니다.

| 조건 | 기록할 내용 |
| --- | --- |
| 실행 환경 | OS, CPU, Go 버전 |
| 입력 | 키 수, 값 크기, 변경 비율 |
| 동시성 | 작업자 수와 GOMAXPROCS |
| 방법 | 벤치마크 명령과 반복 수 |
| 관찰 | 실행 시간, 할당, 변동 |

race detector를 켠 검사는 안전성 검사용입니다.
추가 계측 비용이 있으므로 일반 실행의 성능 숫자와 섞지 않습니다.
구체적인 측정 코드는 26장 과정을 따른 뒤 작성합니다.

## 테스트의 의미

### 메모리 접근 검사와 불변 조건 검사

레이스 검사는 같은 메모리의 충돌 접근을 찾습니다.
재고가 음수가 되지 않는지, 슬롯을 두 번 반납하지 않는지 같은 조건은
반환값과 최종 상태로 따로 검사합니다.

원자 API로만 잘못 구현한 Load→Store 증가는
레이스 검사만으로 잡히지 않을 수 있습니다.
따라서 성공 횟수·잔여량·초기화 횟수 같은 외부 계약을 확인합니다.

### 동시 테스트의 자료 기록

각 호출의 결과는 인덱스별 칸에 저장하거나 안전한 원자 카운터로 기록합니다.
테스트에서 일반 count++를 공유하면 테스트 자체가 레이스를 만들 수 있습니다.

호출한 작업은 모두 Wait한 뒤 결과를 검사합니다.
테스트 성공을 Sleep의 길이에 기대지 않습니다.

## 검증 명령

```sh
go fmt ./solutions/20-sync/...
go vet ./solutions/20-sync/...
go test ./solutions/20-sync/...
go build ./solutions/20-sync/...
```

Linux와 C 컴파일러가 준비된 환경:

```sh
go test -race ./solutions/20-sync/...
```

## 평가 기준

| 항목 | 통과 기준 |
| --- | --- |
| 소유권 | 누가 어떤 상태를 수정하는지 분명 |
| 보호 범위 | 검사와 상태 변경이 같은 불변 조건을 유지 |
| 실패 | 실패 시 원치 않는 부분 변경 없음 |
| 복사 | 잠금 객체와 공유 자료의 복사 조건을 준수 |
| 반환 | 외부에서 내부 상태를 우회 수정할 수 없음 |
| 종료 | 작업을 기다린 뒤 최종 상태 검사 |
| 검증 | race와 논리적 조건 모두 확인 |

## 참고 자료

- [sync.Mutex](https://pkg.go.dev/sync#Mutex)
- [sync.Once](https://pkg.go.dev/sync#Once)
- [sync.OnceValues](https://pkg.go.dev/sync#OnceValues)
- [sync/atomic](https://pkg.go.dev/sync/atomic)
- [Go 메모리 모델](https://go.dev/ref/mem)
- [Data Race Detector](https://go.dev/doc/articles/race_detector)

[문제로 돌아가기](../../exercises/20-sync/README.md) ·
[본문으로 돌아가기](../../book/20-sync/README.md)
