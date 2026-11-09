# 06장 해설: 맵과 집계

[문제](../../exercises/06-maps/README.md)를 먼저 풀고 비교합니다. 모든 명령은 저장소 루트 기준입니다. 맵 자체를 직접 출력한 모습만으로 순회 순서를 판단하지 말고 키별 계약을 확인합니다.

문제 2·4·5·6은 참고 패키지와 테스트를 제공합니다. 문제 1·3·선택 A는 아래 전체 코드를 해당 개인 폴더에 저장합니다. 기존 개인 코드는 먼저 보존하세요.

## 문제 1. 값과 존재 추적

예상 출력:

```text
0 1
0 true
0 false
1 1
```

```go
package main

import "fmt"

func main() {
	counts := map[string]int{"zero": 0}
	fmt.Println(counts["missing"], len(counts))
	value, ok := counts["zero"]
	fmt.Println(value, ok)
	value, ok = counts["missing"]
	fmt.Println(value, ok)
	counts["missing"]++
	delete(counts, "zero")
	fmt.Println(counts["missing"], len(counts))
}
```

```sh
go run ./_practice/06-maps/predict
```

| 단계 | 존재하는 키 | 중요한 값 |
| --- | --- | --- |
| 선언 | zero | zero의 값 0 |
| missing 조회 | zero | missing 조회값 0 |
| 두 결과 조회 | zero | zero는 true, missing은 false |
| missing 증가 | zero, missing | missing의 값 1 |
| zero 삭제 | missing | len은 1 |

조회는 등록하지 않고 증가는 값을 저장합니다. 0이라는 값만으로는 키 존재를 알 수 없습니다. 이 구분은 빈도, 가격 0, 재고 0 등의 입력에서 반복해서 필요합니다. [맵 조회](https://go.dev/ref/spec#Index_expressions)의 규칙을 참고합니다.

## 문제 2. 항목 빈도

참고 구현: [tally/main.go](tally/main.go) · [테스트](tally/main_test.go).

```go
package main

import (
	"fmt"
	"sort"
)

func tally(items []string) map[string]int {
	counts := make(map[string]int)
	for _, item := range items {
		counts[item]++
	}
	return counts
}
func main() {
	counts := tally([]string{"pen", "book", "pen", "note", "book", "pen"})
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

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/06-maps/tally
go test ./solutions/06-maps/tally
```

예상 출력:

```text
book: 2
note: 1
pen: 3
```


### 값의 영값을 사용하는 누적

새 항목은 조회값이 0이므로 ++가 1을 저장합니다. 기존 항목은 이전 횟수에 1을 더합니다. 키가 처음인지 검사하는 if가 없어도 됩니다.

| 입력 | 갱신 항목 |
| --- | --- |
| pen | pen:1 |
| book | book:1 |
| pen | pen:2 |
| note | note:1 |
| book | book:2 |
| pen | pen:3 |

집계 함수에서는 정렬하지 않습니다. 반환할 자료가 맵이므로 순서 자체가 계약이 아니기 때문입니다. main에서 키를 []string에 수집한 뒤 sort.Strings로 정렬합니다.

기본 결과는 키 3개이고 값의 합계는 6입니다. 키 개수와 입력 개수의 관계를 검사하면 누락이나 중복 집계 오류를 찾는 데 도움이 됩니다.

| 경계 | 결과 |
| --- | --- |
| nil·빈 입력 | nil이 아닌 빈 맵 |
| `[a a A]` | a:2와 A:1 |
| 빈 문자열 두 개 | 빈 문자열 키의 값 2 |

빈 문자열도 허용된 입력입니다. 빈 문자열 키의 출력은 이름 부분이 비어 보일 수 있으므로 실제 키를 조사할 때는 `%q`로 따옴표를 붙여 확인할 수 있습니다. 문제의 기본 출력 형식은 `%s`를 유지합니다.

```sh
go run ./_practice/06-maps/tally
```

## 문제 3. nil 맵 생성

원래 코드는 항목 대입에서 `assignment to entry in nil map`으로 실패합니다. 타입을 선언한 것만으로 저장 가능한 맵을 만든 것이 아니기 때문입니다.

수정한 전체 코드:

```go
package main

import "fmt"

func main() {
	stock := make(map[string]int)
	stock["pen"] = 3
	fmt.Println(stock["pen"])
}
```

```sh
go run ./_practice/06-maps/repair
```

예상 출력:

```text
3
```

`stock := map[string]int{}`도 올바른 준비입니다. 초기 더미 항목을 넣을 필요가 없습니다. 생성된 빈 맵은 이미 쓰기 가능한 상태이며 실제 키는 대입하면서 생깁니다.

| 연산 | 원래 nil 맵 | 수정한 빈 맵 |
| --- | --- | --- |
| 읽기 | 영값 | 영값 |
| len | 0 | 0 |
| delete | 가능 | 가능 |
| 항목 대입 | panic | 가능 |

실패를 복구하는 처리보다 잘못된 초기 상태를 없애는 것이 이 문제의 목표입니다. [맵 타입](https://go.dev/ref/spec#Map_types)을 참고합니다.

## 문제 4. 고유 목록

참고 구현: [unique/main.go](unique/main.go) · [테스트](unique/main_test.go).

```go
package main

import "fmt"

func unique(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	return result
}
func main() {
	items := []string{"go", "map", "go", "slice", "map"}
	fmt.Println("고유 목록:", unique(items))
	fmt.Println("원본:", items)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/06-maps/unique
go test ./solutions/06-maps/unique
```

예상 출력:

```text
고유 목록: [go map slice]
원본: [go map go slice map]
```


### 맵과 슬라이스의 역할

seen은 이미 만난 문자열을 기억합니다. result는 처음 등장한 순서를 보존합니다. 입력을 한 번씩 확인하면서 처음인 순간에만 추가합니다.

| 입력 | 이미 봄 | 결과 |
| --- | --- | --- |
| go | 아니오 | `[go]` |
| map | 아니오 | `[go map]` |
| go | 예 | `[go map]` |
| slice | 아니오 | `[go map slice]` |
| map | 예 | `[go map slice]` |

마지막에 seen의 키를 순회해서 결과를 만들면 첫 등장 순서를 보장할 수 없습니다. 정렬하는 방법도 요구 순서와 다릅니다.

모든 등록 값은 true입니다. 이 약속이 있기 때문에 단일 조회의 false를 아직 보지 못한 항목으로 해석할 수 있습니다. 일반 bool 설정 맵에서는 false 자체가 저장된 값일 수 있으므로 같은 해석을 그대로 적용하면 안 됩니다.

| 입력 | 결과 |
| --- | --- |
| nil·빈 입력 | nil |
| `[a a a]` | `[a]` |
| `[b a b A]` | `[b a A]` |
| 빈 문자열, a, 빈 문자열 | 빈 문자열, a |

결과 슬라이스의 원소 자리도 독립적입니다. 테스트는 결과의 첫 문자열을 교체해도 원본 슬라이스가 유지되는지 확인합니다.

```sh
go run ./_practice/06-maps/unique
```

## 문제 5. 재고 출고

참고 구현: [withdraw/main.go](withdraw/main.go) · [테스트](withdraw/main_test.go).

```go
package main

import "fmt"

func withdraw(stock map[string]int, item string, quantity int) bool {
	available, ok := stock[item]
	if !ok || quantity <= 0 || quantity > available {
		return false
	}
	stock[item] = available - quantity
	return true
}
func main() {
	stock := map[string]int{"pen": 5, "book": 0}
	fmt.Println("첫 출고:", withdraw(stock, "pen", 3))
	fmt.Println("남은 pen:", stock["pen"])
	fmt.Println("초과 출고:", withdraw(stock, "pen", 3))
	fmt.Println("없는 상품:", withdraw(stock, "note", 1))
	fmt.Println("최종 pen:", stock["pen"])
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/06-maps/withdraw
go test ./solutions/06-maps/withdraw
```

예상 출력:

```text
첫 출고: true
남은 pen: 2
초과 출고: false
없는 상품: false
최종 pen: 2
```


### 검사 뒤 한 번만 변경하기

조회와 검사는 맵을 바꾸지 않습니다. 모든 실패 조건을 통과한 뒤에만 차감한 값을 한 번 저장합니다. 실패 경로에는 복구 대입이 필요하지 않습니다.

nil 맵의 조회는 안전하게 0과 false를 반환합니다. 존재 조건에서 실패하므로 쓰기 문장에 도달하지 않습니다. 이 함수는 없는 상품을 자동 등록하지 않습니다.

| 호출 | 실행 전 pen | 결과 | 실행 후 pen |
| --- | --- | --- | --- |
| pen 3개 | 5 | true | 2 |
| pen 3개 | 2 | false | 2 |
| note 1개 | 2 | false | 2 |

각 경계 검사는 새 재고로 시작해야 합니다. 앞 호출에서 이미 재고를 줄였다면 같은 수량 호출의 결과가 달라질 수 있습니다. 이 함수는 입력 상태를 변경하는 계약입니다.

| 새 재고 pen:5에서 수량 | 결과 | 남은 수량 |
| --- | --- | --- |
| -1 | false | 5 |
| 0 | false | 5 |
| 5 | true | 0 |
| 6 | false | 5 |

전부 출고한 경우 pen 키는 0 값으로 남습니다. delete를 호출하면 문제의 계약이 바뀝니다. 등록된 품목인지와 현재 재고가 있는지는 서로 다른 질문입니다.

```sh
go run ./_practice/06-maps/withdraw
```

## 문제 6. 집계 병합

참고 구현: [merge-counts/main.go](merge-counts/main.go) · [테스트](merge-counts/main_test.go).

```go
package main

import (
	"fmt"
	"sort"
)

func mergeCounts(a, b map[string]int) map[string]int {
	result := make(map[string]int)
	for key, value := range a {
		result[key] = value
	}
	for key, value := range b {
		result[key] += value
	}
	return result
}
func main() {
	first := map[string]int{"go": 2, "map": 1}
	second := map[string]int{"go": 3, "slice": 2}
	merged := mergeCounts(first, second)
	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, merged[key])
	}
	fmt.Println("원본 go:", first["go"], second["go"])
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/06-maps/merge-counts
go test ./solutions/06-maps/merge-counts
```

예상 출력:

```text
go: 5
map: 1
slice: 2
원본 go: 2 3
```


### 독립 목적지에 두 입력 반영

먼저 첫 입력의 항목을 새 맵에 복사하고, 두 번째 입력의 값은 더합니다. 없는 키는 0에서 시작하므로 한쪽에만 있는 키도 같은 문장으로 처리합니다.

| 키 | 첫 입력 | 두 번째 입력 | 결과 |
| --- | --- | --- | --- |
| go | 2 | 3 | 5 |
| map | 1 | 없음 | 1 |
| slice | 없음 | 2 | 2 |

nil 맵은 순회가 0회입니다. 첫 입력이 nil이라고 둘째 입력을 그대로 반환하지 않으므로 그 경우에도 독립성이 유지됩니다.

빈 결과는 make한 맵이며 nil이 아닙니다. 값 0인 키도 첫 복사 또는 두 번째 += 대입에서 실제로 등록됩니다.

### 잘못된 대안 구분하기

`result := a` 뒤 합산하면 a가 함께 바뀝니다. 기본 합계 출력만 보면 맞아 보여도 입력 보존 계약을 어긴 것입니다. 결과 수정·삭제 뒤 두 입력이 유지되는지도 테스트합니다.

`a`와 `b`가 같은 맵 값일 때도 새 결과에 따로 기록하므로 원본은 유지되고 각 값은 두 번 합산됩니다. 이 경우도 합계가 int 범위를 넘지 않아야 합니다.

```sh
go run ./_practice/06-maps/merge-counts
```

## 선택 A. 내용 비교

다음 전체 프로그램을 `_practice/06-maps/same-counts/main.go`에 작성합니다.

```go
package main

import "fmt"

func sameCounts(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		other, exists := b[key]
		if !exists || other != value {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println("빈 비교:", sameCounts(nil, map[string]int{}))
	fmt.Println("같은 값:", sameCounts(map[string]int{"a": 2}, map[string]int{"a": 2}))
	fmt.Println("다른 키:", sameCounts(map[string]int{"a": 0}, map[string]int{"b": 0}))
	fmt.Println("다른 값:", sameCounts(map[string]int{"a": 2}, map[string]int{"a": 3}))
}
```

```sh
go run ./_practice/06-maps/same-counts
```

예상 출력:

```text
빈 비교: true
같은 값: true
다른 키: false
다른 값: false
```

길이가 다르면 즉시 false입니다. 같으면 한쪽의 각 키가 다른 쪽에도 존재하고 값이 같은지 확인합니다. 첫 번째 맵의 모든 항목이 둘째에 있고 길이도 같으면 둘째에 추가 항목이 있을 수 없습니다.

| 비교 | 값만 읽으면 | 존재까지 읽으면 |
| --- | --- | --- |
| a:0과 b:0 | 없는 a 조회값도 0 | a 키 부재로 false |
| a:2와 a:2 | 2와 2 | 키도 있고 같으므로 true |
| nil과 빈 맵 | 순회 0회 | 같은 빈 항목 집합으로 true |

nil 자체를 구별하는 비교가 목적이면 맨 앞에 nil 여부 비교를 추가해야 합니다. 이번 문제는 내용 비교이므로 같은 빈 집합으로 정의했습니다.

## 검사 범위

```sh
go test -count=1 ./solutions/06-maps/...
```

| 참고 구현 | 함수 검사 |
| --- | --- |
| tally | 빈 입력의 non-nil, 대소문자, 빈 문자열 |
| unique | 첫 등장 순서, nil·빈 입력, 독립 원소 |
| withdraw | 잘못된 수량, 없는 키, nil, 전부 출고 후 키 유지 |
| merge-counts | 겹치는 키, 0 키, nil 입력, 결과 수정 후 입력 유지 |

각 패키지의 Example_main은 기본 출력도 검사합니다. 테스트에서 쓰는 보조 타입과 testing은 후속 장에서 다룹니다. 개인 코드는 문제의 실행 명령과 경계 표로 별도 확인합니다.

## 다시 설명해 볼 질문

1. 빈도 집계에는 단일 조회가 충분하지만 재고 등록 여부에는 왜 두 결과가 필요할까요?
2. 처음 등장 순서와 정렬 순서는 언제 다른가요?
3. 실패 시 변경하지 않으려면 검사와 대입을 어떤 순서로 두어야 하나요?
4. nil 입력 분기에서 다른 맵을 그대로 반환하면 어떤 계약을 놓치나요?
5. 맵 값이 슬라이스였다면 독립 병합에서 무엇을 더 확인해야 하나요?

[본문](../../book/06-maps/README.md) · [문제](../../exercises/06-maps/README.md)
