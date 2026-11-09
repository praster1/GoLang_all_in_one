# 13장 해설: 제네릭

먼저 [연습문제](../../exercises/13-generics/README.md)를 풀어 보세요. 모든 명령은 저장소 루트에서 실행합니다. 해설은 타입 관계뿐 아니라 순서·빈 입력·저장 공간의 계약을 함께 설명합니다.

## 문제 1. 타입 판단

| 호출·선언 | 결과 | 이유 |
| --- | --- | --- |
| `first([]int{0})` | 가능, T=int | []int에서 추론 |
| `first[string](nil)` | 가능, T=string | 명시적 타입 인자 |
| `first(nil)` | 불가 | T를 알 정보 부족 |
| `index([]string{"a"}, "a")` | 가능, T=string | string은 comparable |
| `index([][]int{{1}}, []int{1})` | 불가 | 원소 타입 []int는 comparable 아님 |
| `sum([]Cents{100,200})` | 가능, T=Cents | 기반 타입 int64를 ~int64가 포함 |
| `sum([]float64{1.5})` | 불가 | Integer 타입 집합에 없음 |
| `var value Integer` | 불가 | 비기본 제약 인터페이스는 일반 값 변수로 사용 불가 |

Exact가 정확히 int64만 포함하면 새 정의 타입 Cents는 포함되지 않습니다. `~int64`는 int64 자체와 그 기반 타입을 가진 정의 타입을 포함합니다.

comparable에 인터페이스 타입을 대입할 수 있는 경우에도 그 인터페이스 안의 동적 값이 비교 불가능하면 런타임 패닉이 가능합니다. 타입 인자 규칙과 동적 값의 비교 계약을 나누어 봅니다.

Labels를 []T 기반 함수에 전달하면 T는 string이지만 반환 타입은 []string입니다. S ~[]E를 사용하고 S를 반환하면 결과 타입 Labels를 유지할 수 있습니다.

## 문제 2. 영값과 존재 여부 분리

빈 입력을 먼저 검사하여 len(values)-1이 음수인 상태로 접근하지 않도록 합니다. 일반적인 원소 값을 복사하여 반환하므로 포인터 원소라면 그 포인터 값이 그대로 반환됩니다.

참고 구현: [last/main.go](last/main.go).

```go
package main

import "fmt"

func last[T any](values []T) (T, bool) {
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[len(values)-1], true
}
func main() {
	value, ok := last([]int{2, 4, 6})
	fmt.Println(value, ok)
	word, ok := last[string](nil)
	fmt.Printf("%q %t\n", word, ok)
}
```

```sh
go run ./solutions/13-generics/last
go test ./solutions/13-generics/last
```

예상 출력:

```text
6 true
"" false
```


빈 []string에는 빈 문자열과 false를, 원소가 빈 문자열인 []string에는 빈 문자열과 true를 반환합니다. result만으로 성공 여부를 판단하지 않습니다.

첫 값 함수와 거의 같은 구조이지만 인덱스 계산은 길이 검사가 끝난 뒤에 합니다. 함수 제약이 any인 이유는 비교나 더하기를 하지 않고 값만 전달하기 때문입니다.

## 문제 3. 순서는 슬라이스, 존재 여부는 맵

map은 이미 나온 값을 기억하는 데만 사용하고, 결과 순서는 입력 순회가 결정합니다. 맵 키를 나중에 순회해서 결과를 만들면 첫 등장 순서가 사라집니다.

참고 구현: [unique/main.go](unique/main.go).

```go
package main

import "fmt"

func unique[S ~[]E, E comparable](values S) S {
	if values == nil {
		return nil
	}
	result := make(S, 0, len(values))
	seen := make(map[E]bool)
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

type Names []string

func main() {
	fmt.Println(unique([]int{2, 1, 2, 3, 1}))
	names := unique(Names{"Go", "SQL", "Go"})
	fmt.Printf("%v (%T)\n", names, names)
}
```

```sh
go run ./solutions/13-generics/unique
go test ./solutions/13-generics/unique
```

예상 출력:

```text
[2 1 3]
[Go SQL] (main.Names)
```


### 단계 추적

입력이 [2,1,2,3,1]일 때:

| 입력 값 | 이미 봄 | 결과 |
| ---: | --- | --- |
| 2 | 아니요 | [2] |
| 1 | 아니요 | [2 1] |
| 2 | 예 | [2 1] |
| 3 | 아니요 | [2 1 3] |
| 1 | 예 | [2 1 3] |

S를 반환하므로 Names 타입이 유지됩니다. 새 바깥 저장 공간을 만들기 때문에 int 또는 string 원소를 대입해 바꿔도 원본 요소는 유지됩니다. 원소 안에 참조가 들어 있으면 일반 얕은 복사 규칙은 그대로 적용됩니다.

### 적용 범위

comparable에는 부동소수점과 인터페이스 관련 경계가 있습니다. NaN은 자기 자신과 ==가 false인 특성이 있어 “같아 보이는 NaN을 한 번만 남긴다”라는 별도 정책을 이 구현이 제공하지 않습니다. 문제의 입력은 그런 정책을 피한 타입과 값으로 한정했습니다.

공개 범용 유틸리티로 확장한다면 같음의 기준을 함수로 받는 방식도 가능합니다. 그 경우 맵 키 기반 존재 확인을 그대로 사용할 수 있는지 별도로 판단해야 합니다.

## 문제 4. 키 타입만 순서가 필요합니다

정렬하는 대상은 키 K입니다. 값 V는 정렬·비교에 사용하지 않으므로 any면 충분합니다. map[string][]int도 함수의 타입 조건을 만족합니다.

참고 구현: [keys/main.go](keys/main.go).

```go
package main

import (
	"fmt"
	"slices"
)

type Key interface{ ~int | ~string }

func sortedKeys[K Key, V any](values map[K]V) []K {
	result := make([]K, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	slices.Sort(result)
	return result
}
func main() {
	fmt.Println(sortedKeys(map[string]int{"b": 2, "a": 1}))
	fmt.Println(sortedKeys(map[int]string{10: "ten", 2: "two"}))
}
```

```sh
go run ./solutions/13-generics/keys
go test ./solutions/13-generics/keys
```

예상 출력:

```text
[a b]
[2 10]
```


문제의 Key는 ~int와 ~string만 허용합니다. 이는 float의 NaN·특수 키 동작까지 이 함수에서 다루지 않기 위한 명시적 범위입니다. cmp.Ordered로 무조건 넓히는 것이 문제의 요구는 아닙니다.

make로 결과를 만들기 때문에 nil 맵에서도 non-nil 빈 슬라이스를 반환합니다. 입력이 nil일 때만 별도 nil 반환하는 계약이었다면 코드와 테스트가 달라졌을 것입니다.

맵 순회가 어떤 순서여도 마지막에 정렬하므로 최종 키 목록은 결정적입니다. 숫자 키는 정수 순서, 문자열 키는 기본 문자열 비교 순서를 따릅니다.

## 문제 5. 그룹 내부 순서를 유지

각 입력에서 키를 한 번 계산하고 해당 슬라이스에 원소를 append합니다. 같은 키의 기존 슬라이스 끝에 추가하므로 그룹 안의 입력 순서가 유지됩니다.

참고 구현: [group/main.go](group/main.go).

```go
package main

import (
	"fmt"
	"slices"
)

func groupBy[T any, K comparable](values []T, key func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, value := range values {
		k := key(value)
		result[k] = append(result[k], value)
	}
	return result
}

type Lesson struct {
	Language string
	Minutes  int
}

func main() {
	groups := groupBy([]Lesson{{"Go", 20}, {"SQL", 30}, {"Go", 40}}, func(v Lesson) string { return v.Language })
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		fmt.Println(key, groups[key])
	}
}
```

```sh
go run ./solutions/13-generics/group
go test ./solutions/13-generics/group
```

예상 출력:

```text
Go [{Go 20} {Go 40}]
SQL [{SQL 30}]
```


### 그룹 내부와 바깥 키의 순서

| 대상 | 순서 계약 |
| --- | --- |
| Go 그룹의 Lesson | 입력의 20분 다음 40분 |
| SQL 그룹의 Lesson | 30분 한 항목 |
| 맵의 Go와 SQL 순회 | 보장하지 않음 |
| main에서 출력하는 키 | 별도 정렬하여 결정 |

결과 맵은 make로 생성합니다. nil 맵에 슬라이스를 대입하는 실수를 피하고, 빈 입력에서도 non-nil 결과라는 계약을 지킵니다.

callback을 지역 변수 k에 한 번 저장합니다. `result[key(value)] = append(result[key(value)], value)`처럼 작성하면 key 함수를 같은 원소에서 두 번 호출할 수 있습니다. 부작용이나 계산 비용을 고려하지 않더라도 명시한 호출 횟수 계약에 어긋납니다.

입력이 구조체 값이면 구조체를 일반 대입 방식으로 복사합니다. 구조체 안의 슬라이스·포인터까지 독립적으로 복제하는 함수는 아닙니다. 그룹 결과를 바꾸는 API를 제공할 때 어느 층을 공유하는지 구분합니다.

## 문제 6. 한 번의 순회에서 정확히 한 분기

입력마다 keep을 한 번 호출한 뒤 두 결과 중 하나에만 추가합니다. 따라서 두 길이의 합은 입력 길이이고, 각 결과 안의 상대 순서는 유지됩니다.

참고 구현: [partition/main.go](partition/main.go).

```go
package main

import "fmt"

func partition[S ~[]E, E any](values S, keep func(E) bool) (S, S) {
	if values == nil {
		return nil, nil
	}
	yes, no := make(S, 0, len(values)), make(S, 0, len(values))
	for _, value := range values {
		if keep(value) {
			yes = append(yes, value)
		} else {
			no = append(no, value)
		}
	}
	return yes, no
}
func main() {
	even, odd := partition([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	fmt.Println("짝수:", even)
	fmt.Println("홀수:", odd)
}
```

```sh
go run ./solutions/13-generics/partition
go test ./solutions/13-generics/partition
```

예상 출력:

```text
짝수: [2 4]
홀수: [1 3]
```


### 왜 filter를 두 번 호출하지 않았나요?

순수한 함수라면 true 조건과 false 조건으로 두 번 필터해 같은 값 목록을 얻을 수 있습니다. 그러나 callback 호출이 입력마다 두 번이 되며, 원소마다 한 번이라는 계약을 어깁니다. callback이 상태를 읽거나 부작용을 가지면 결과 자체가 달라질 수도 있습니다.

### 왜 결과를 둘 다 따로 만드나요?

입력 앞부분을 결과 버퍼로 재사용하면 입력을 덮어쓸 수 있습니다. 같은 배열의 앞뒤를 두 결과로 나누면 append나 원소 대입에서 공유 영향을 고려해야 합니다. 이번 문제는 독립된 바깥 저장 공간을 요구하므로 두 make가 직접적인 구현입니다.

예제는 간단한 구현을 위해 두 결과 모두 입력 길이만큼의 용량을 예약했습니다. 실제 큰 데이터에서는 메모리와 할당 패턴을 측정해 다른 전략을 선택할 수 있습니다. 이 선택의 속도 우위를 측정 없이 단정하지 않습니다.

### 빈 입력 정책

| 입력 | 첫 결과 | 둘째 결과 |
| --- | --- | --- |
| nil | nil | nil |
| non-nil 빈 슬라이스 | non-nil 빈 슬라이스 | non-nil 빈 슬라이스 |
| 모든 값이 true | 전체 값 | non-nil 빈 슬라이스 |
| 모든 값이 false | non-nil 빈 슬라이스 | 전체 값 |

이 구분을 테스트에서 직접 확인합니다. slices.Equal만으로는 nil과 빈 슬라이스가 같은 원소 목록으로 취급될 수 있으므로 필요한 nilness 조건은 따로 확인합니다.

## 선택 A. Pair의 타입 관계

Swap은 기존 수신자 타입 매개변수 A와 B의 위치를 교환합니다. MapLeft는 메서드 자체의 새로운 C를 추가하여 왼쪽만 변환합니다.

참고 구현: [pair/main.go](pair/main.go).

```go
package main

import (
	"fmt"
	"strconv"
)

type Pair[A, B any] struct {
	Left  A
	Right B
}

func (pair Pair[A, B]) Swap() Pair[B, A] { return Pair[B, A]{Left: pair.Right, Right: pair.Left} }
func (pair Pair[A, B]) MapLeft[C any](convert func(A) C) Pair[C, B] {
	return Pair[C, B]{Left: convert(pair.Left), Right: pair.Right}
}
func main() {
	pair := Pair[int, string]{Left: 7, Right: "Go"}
	fmt.Println("교환:", pair.Swap())
	mapped := pair.MapLeft(func(v int) string { return "#" + strconv.Itoa(v) })
	fmt.Printf("왼쪽 변환: %v (%T)\n", mapped, mapped)
	fmt.Println("원본:", pair)
}
```

```sh
go run ./solutions/13-generics/pair
go test ./solutions/13-generics/pair
```

예상 출력:

```text
교환: {Go 7}
왼쪽 변환: {#7 Go} (main.Pair[string,string])
원본: {7 Go}
```


| 메서드 | 수신자 타입 인자 | 메서드 자체 타입 인자 | 결과 |
| --- | --- | --- | --- |
| Swap | A, B | 없음 | Pair[B,A] |
| MapLeft | A, B | C | Pair[C,B] |

MapLeft의 [C any]는 Go 1.27의 제네릭 메서드 구문입니다. 반면 제네릭 타입에 기존 A·B를 사용하는 일반 메서드를 붙이는 것은 이전 Go 제네릭 기능에서도 가능했습니다.

값 수신자에서 새 Pair를 만들기 때문에 원본의 int·string 필드는 그대로입니다. A나 B가 포인터라면 일반적인 포인터 값 복사이며 대상까지 깊게 복제하지 않습니다.

테스트는 결과를 Pair[string,int]와 Pair[string,string] 변수에 명시적으로 대입하여 컴파일 시점의 타입 관계를 확인합니다. 런타임 문자열에 타입 이름이 비슷하게 찍히는지만 검사하는 것보다 실제 사용 계약에 가깝습니다.

## 본문 자기 점검 답

1. 값과 bool을 함께 확인합니다. 정상 0은 true이고 빈 입력의 0은 false입니다.
2. any에는 == 비교를 지원하지 않는 타입도 포함됩니다.
3. int64는 정확한 타입, ~int64는 기반 타입이 int64인 정의 타입까지 포함합니다.
4. []E는 원소 타입을 유지하고, S 반환은 정의된 슬라이스 타입 이름과 메서드도 유지할 수 있습니다.
5. 바깥 슬라이스의 새 저장 공간이 내부 포인터 대상의 깊은 복제를 뜻하지 않습니다.
6. Push는 수신자 타입의 T를 사용하고, Map은 메서드 자체의 U를 추가합니다.
7. 인터페이스 메서드 자체 타입 매개변수는 Go 1.27에서도 허용되지 않습니다.
8. NaN을 일반 숫자보다 앞에 두는 cmp.Compare 정책 때문에 NaN을 선택합니다.
9. 성능은 실제 작업과 구현·할당 조건에 따라 측정해야 합니다.
10. 학습·다른 계약·실제 필요가 없다면 표준 함수의 계약을 확인해 사용하는 편이 유지 관리 부담을 줄일 수 있습니다.

## 검증 명령

```sh
go fmt ./examples/13-generics/... ./solutions/13-generics/...
go vet ./examples/13-generics/... ./solutions/13-generics/...
go test ./examples/13-generics/... ./solutions/13-generics/...
go build ./examples/13-generics/... ./solutions/13-generics/...
```

정상 출력 외에도 nil·빈 입력·경계 위치·정의 타입 보존·callback 횟수·입력 저장 공간 독립을 검사합니다. Go 1.27 메서드 구문은 실제 컴파일과 실행 예제로 검증합니다.

본문의 의도적 컴파일 실패 예제는 정상 테스트 대상에 넣지 않았습니다. 그 실패는 제한을 설명하는 학습 결과이며, 기본 예제가 빌드되지 않는 상태와 구별해야 합니다.

[본문](../../book/13-generics/README.md) · [문제로 돌아가기](../../exercises/13-generics/README.md)
