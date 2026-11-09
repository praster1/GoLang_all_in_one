# 06. 맵과 집계

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 입문 | [05. 배열과 슬라이스](../05-arrays-slices/README.md) | Go, 편집기, 터미널, 이 저장소 |

슬라이스는 0, 1, 2 같은 위치로 원소를 찾습니다. 상품의 재고를 찾을 때는 “세 번째 상품”보다 “pen이라는 상품”이 더 자연스러울 수 있습니다. 이름과 값의 연결을 저장하는 자료형이 **맵(map)**입니다.

이번 장은 맵을 만들고 조회하는 문법에서 출발해 빈도 집계, 중복 제거, 그룹 묶기까지 이어집니다. 특히 **없는 키와 저장된 영값의 차이**, **맵의 순서**, **함수에서의 변경과 복사**를 구분합니다.

## 학습 목표

- 키와 값의 타입을 정하고 맵을 생성·조회·갱신·삭제한다.
- 두 결과 조회로 없는 키와 저장된 영값을 구분한다.
- nil 맵에서 가능한 연산과 쓰기 전 준비를 설명한다.
- 순회 결과를 필요에 따라 정렬하고 빈도·집합·그룹을 만든다.
- 함수의 변경 책임과 독립 복사 계약을 검사한다.

## 사전 지식과 준비

문자열·정수, if·range, 함수의 여러 반환값, 슬라이스와 append를 사용합니다. 문자열의 바이트·유니코드 차이는 다음 장에서 다루며, 여기서는 같은 문자열인지 비교하는 데 사용합니다.

모든 명령은 **저장소 루트**에서 실행합니다. 새 모듈을 만들지 않습니다. 전체 코드는 연결된 파일과 같으며, 발췌는 개인 `main` 안 또는 명시한 함수 선언 위치에 넣습니다.

| 예제 | 질문 | 실행 명령 |
| --- | --- | --- |
| basics | 이름으로 값을 저장하고 삭제하려면? | `go run ./examples/06-maps/basics` |
| lookup | 0은 저장된 값일까, 없는 키일까? | `go run ./examples/06-maps/lookup` |
| nil-map | nil 맵에는 무엇을 할 수 있을까? | `go run ./examples/06-maps/nil-map` |
| frequency | 여러 번 등장한 항목을 세려면? | `go run ./examples/06-maps/frequency` |
| sets | 처음 등장한 순서로 중복을 없애려면? | `go run ./examples/06-maps/sets` |
| groups | 키 하나에 여러 값을 담으려면? | `go run ./examples/06-maps/groups` |
| map-functions | 함수에서 원본 변경과 복사를 구분하려면? | `go run ./examples/06-maps/map-functions` |

개인 실습은 11절의 보존형 준비 명령으로 `_practice/06-maps/lab/main.go`를 만든 뒤 실행합니다. 실제 예제 파일은 읽는 기준으로 사용합니다.

## 1. 키와 값의 연결

맵은 **키(key)**를 사용해 **값(value)**을 찾습니다. `map[string]int`는 키가 문자열이고 값이 정수인 맵 타입입니다. 예를 들어 상품 이름을 키로, 재고 수량을 값으로 저장합니다.

```go
stock := map[string]int{
	"pen":  3,
	"book": 2,
}
fmt.Println(stock["pen"])
```

위는 main 안의 발췌이며 출력은 `3`입니다. 중괄호 안의 `키: 값` 쌍은 맵 리터럴입니다. 여러 줄로 적으면 마지막 항목 뒤에도 쉼표를 둡니다.

| 부분 | 뜻 |
| --- | --- |
| `map` | 맵 타입을 시작하는 키워드 |
| `[string]` | 키의 타입 |
| `int` | 값의 타입 |
| `"pen": 3` | 키 pen에 정수 3 저장 |
| `stock["pen"]` | 그 키로 값 조회 |

하나의 맵에서는 키의 타입과 값의 타입이 각각 일관되어야 합니다. 키가 문자열인 맵에 정수 키를 넣거나, 값이 int인 맵에 문자열 값을 넣을 수 없습니다. [맵 타입](https://go.dev/ref/spec#Map_types)을 참고합니다.

### 1.1 등록, 갱신, 삭제

저장 경로: [`examples/06-maps/basics/main.go`](../../examples/06-maps/basics/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	stock := map[string]int{"pen": 3, "book": 2}
	fmt.Println("pen:", stock["pen"])
	stock["pen"] = 5
	stock["note"] = 4
	fmt.Println("수정한 pen:", stock["pen"])
	fmt.Println("등록 수:", len(stock))
	delete(stock, "book")
	fmt.Println("삭제 뒤 수:", len(stock))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/06-maps/basics
```

예상 출력:

```text
pen: 3
수정한 pen: 5
등록 수: 3
삭제 뒤 수: 2
```

`stock["pen"] = 5`는 이미 있던 키의 값을 바꿉니다. 같은 키에 여러 값을 별개 항목으로 쌓는 것이 아닙니다. `stock["note"] = 4`는 새 키이므로 항목 수가 하나 늘어납니다.

| 실행 | pen | book | note | len |
| --- | --- | --- | --- | --- |
| 리터럴 생성 | 3 | 2 | 없음 | 2 |
| pen 갱신 | 5 | 2 | 없음 | 2 |
| note 추가 | 5 | 2 | 4 | 3 |
| book 삭제 | 5 | 없음 | 4 | 2 |

`len(stock)`은 모든 수량의 합이 아니라 등록된 키의 개수입니다. 수량 합계를 구하려면 값을 순회하며 더해야 합니다.

`delete(stock, "book")`은 해당 키를 제거하며 반환값은 없습니다. 없는 키를 삭제해도 오류가 나지 않습니다. 삭제 성공 여부가 필요하면 먼저 존재를 확인하는 계약을 설계합니다. [맵 원소 삭제](https://go.dev/ref/spec#Deletion_of_map_elements)를 참고합니다.


### 1.2 슬라이스와 무엇이 다른가?

| 질문 | 슬라이스 | 맵 |
| --- | --- | --- |
| 값을 찾는 기준 | 0부터 시작하는 위치 | 정해진 타입의 키 |
| 순서 | 위치 순서가 있음 | range 순서가 지정되지 않음 |
| 같은 항목 반복 | 여러 위치에 저장 가능 | 같은 키에 마지막 대입한 값 저장 |
| 없는 위치·키 | 범위 밖 인덱스는 오류 | 값 타입의 영값 반환 |
| 원소 수 | len | len |

맵이 언제나 슬라이스를 대체하는 것은 아닙니다. 입력 순서를 보존할 목록과 이름으로 빠르게 찾을 연결 정보가 함께 필요할 수도 있습니다. 뒤의 중복 제거에서는 맵과 슬라이스를 함께 사용합니다.

이 장에서는 실행 시간의 우열을 수치로 가정하지 않습니다. 먼저 필요한 조회 방식, 순서, 중복 처리 계약으로 자료형을 선택합니다.

## 2. 없는 키와 값 0을 구분하기

맵에서 없는 키를 읽으면 값 타입의 영값을 얻습니다. int 맵은 0, bool 맵은 false, 슬라이스 맵은 nil입니다. **조회만으로 키가 새로 등록되지는 않습니다.** [인덱스 표현식](https://go.dev/ref/spec#Index_expressions)을 참고합니다.

저장된 점수 0과 미등록 점수를 구분해야 할 때는 두 결과 조회를 사용합니다.

### 2.1 두 결과 조회: value, ok

저장 경로: [`examples/06-maps/lookup/main.go`](../../examples/06-maps/lookup/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	scores := map[string]int{"Go": 0}
	value, ok := scores["Go"]
	fmt.Printf("Go: %d, 존재: %t\n", value, ok)
	value, ok = scores["Rust"]
	fmt.Printf("Rust: %d, 존재: %t\n", value, ok)
	fmt.Println("조회 뒤 수:", len(scores))
	scores["Rust"]++
	fmt.Printf("추가한 Rust: %d, 등록 수: %d\n", scores["Rust"], len(scores))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/06-maps/lookup
```

예상 출력:

```text
Go: 0, 존재: true
Rust: 0, 존재: false
조회 뒤 수: 1
추가한 Rust: 1, 등록 수: 2
```

`value, ok := scores["Go"]`에서 value는 저장된 값이고 ok는 키의 존재 여부입니다. Go는 실제로 0이 저장되어 있으므로 true입니다. Rust는 아직 등록되지 않아 0과 false입니다.

| 조회 | 값 | 존재 여부 |
| --- | --- | --- |
| 저장된 Go | 0 | true |
| 없는 Rust | 0 | false |

`ok`는 특별한 예약어가 아니라 이 결과를 담는 변수 이름입니다. 04장의 성공 여부 반환과 비슷하게 두 값을 받지만, 여기서는 맵 조회 문법이 존재 여부를 함께 제공합니다.

`scores["Rust"]++`는 조회만 하는 문장이 아닙니다. 없는 키의 값 0을 바탕으로 1을 저장합니다. 따라서 항목 수도 2가 됩니다. 읽기와 쓰기의 효과를 구분해야 합니다.


### 2.2 조건문에서 존재 여부 사용하기

다음은 main 안의 독립된 발췌입니다.

```go
prices := map[string]int{"sample": 0, "book": 8000}
if price, ok := prices["sample"]; ok {
	fmt.Println("등록 가격:", price)
} else {
	fmt.Println("미등록 상품")
}
```

출력은 `등록 가격: 0`입니다. `price != 0`으로 등록 여부를 판단했다면 무료 상품을 미등록으로 잘못 처리합니다.

`price`와 `ok`는 if 초기화 문장에서 선언했으므로 해당 if·else 범위 안에서만 사용합니다. 조건문 뒤에서도 필요하면 앞에서 따로 조회합니다. 03·04장의 스코프 규칙이 그대로 적용됩니다.

값은 필요 없고 존재만 궁금하면 `_`를 사용합니다.

```go
_, exists := prices["book"]
fmt.Println(exists)
```

위는 앞 prices가 있는 main에 이어 쓰는 발췌이며 true를 출력합니다. 없는 키와 0이 같은 의미인 계약에서만 단일 조회를 사용해도 충분합니다. 의미가 다르면 두 결과로 확인합니다.

## 3. nil 맵과 쓰기 준비

맵의 영값은 nil입니다. `var counts map[string]int`는 타입을 정하지만 쓰기 가능한 빈 맵을 생성하지는 않습니다.

읽기·길이 확인·순회·삭제는 nil 맵에도 적용할 수 있습니다. 쓰기는 준비된 맵이 필요합니다.

### 3.1 읽기와 쓰기의 차이

저장 경로: [`examples/06-maps/nil-map/main.go`](../../examples/06-maps/nil-map/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	var counts map[string]int
	value, ok := counts["Go"]
	fmt.Printf("nil 조회: %d, %t, 길이: %d\n", value, ok, len(counts))
	delete(counts, "Go")
	clear(counts)
	counts = make(map[string]int)
	counts["Go"] = 1
	fmt.Printf("생성 뒤: %d, nil: %t\n", counts["Go"], counts == nil)
	clear(counts)
	fmt.Printf("비운 뒤 길이: %d, nil: %t\n", len(counts), counts == nil)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/06-maps/nil-map
```

예상 출력:

```text
nil 조회: 0, false, 길이: 0
생성 뒤: 1, nil: false
비운 뒤 길이: 0, nil: false
```

첫 조회는 0과 false이며 길이는 0입니다. nil에 delete와 clear를 호출해도 처리할 항목이 없어 끝납니다. 이후 make로 맵을 만든 뒤에 값을 씁니다.

clear는 모든 항목을 삭제하지만 맵 변수를 nil로 바꾸지는 않습니다. 따라서 마지막 줄의 nil 비교는 false이며, 다시 값을 쓸 수 있습니다.

| 연산 | nil 맵 | 생성된 빈 맵 |
| --- | --- | --- |
| `len(m)` | 0 | 0 |
| `m[key]` 읽기 | 영값 | 영값 |
| `value, ok := m[key]` | 영값, false | 영값, false |
| range | 0회 | 0회 |
| delete·clear | 처리할 항목 없음 | 처리할 항목 없음 |
| `m[key] = value` | 실행 중 panic | 등록 가능 |

Go 1.21부터 지원되는 clear는 이 프로젝트의 Go 1.27에서 사용할 수 있습니다. [맵 타입](https://go.dev/ref/spec#Map_types)과 [clear](https://go.dev/ref/spec#Clear)를 참고합니다.


### 3.2 맵을 만드는 두 방법

```go
first := make(map[string]int)
second := map[string]int{}
first["a"] = 1
second["b"] = 2
fmt.Println(len(first), len(second))
```

이 main 안의 발췌는 `1 1`을 출력합니다. 두 표현 모두 쓰기 가능한 빈 맵을 만듭니다. 초기 항목이 있으면 리터럴, 이후 반복문에서 채울 계획이면 make를 쓰는 식으로 읽기 좋게 선택할 수 있습니다.

`make(map[string]int, 100)`의 숫자는 예상 항목 수에 대한 힌트입니다. 키 100개를 만들거나 len을 100으로 정하지 않습니다. 맵에 cap을 사용할 수는 없습니다. [맵 생성](https://go.dev/ref/spec#Making_slices_maps_and_channels)을 참고합니다.

### 3.3 실패를 재현하고 고치기

다음은 **의도적으로 실행 중 실패하는 전체 코드**입니다. 개인 파일에서만 시험합니다.

```go
package main

func main() {
	var counts map[string]int
	counts["go"]++
}
```

```sh
go run ./_practice/06-maps/lab
```

오류에서 `assignment to entry in nil map`을 찾습니다. `counts := make(map[string]int)`로 선언을 바꾸면 실행은 성공합니다. 출력문이 없으므로 성공 후에는 아무 출력도 없습니다.

이 오류를 고치기 위해 임의의 더미 키를 미리 넣을 필요는 없습니다. 맵의 생성과 항목의 존재는 서로 다른 준비입니다.

## 4. 순회와 출력 순서

맵의 range는 키와 값을 제공합니다. 슬라이스의 인덱스·원소와 비슷한 모양이지만 첫 값은 위치 번호가 아니라 키입니다.

```go
counts := map[string]int{"go": 3, "map": 2}
total := 0
for _, count := range counts {
	total += count
}
fmt.Println("전체 횟수:", total)
```

위 main 안의 발췌는 `전체 횟수: 5`를 출력합니다. 작은 정수 합계는 키를 방문한 순서와 관계없이 같은 결과이므로 정렬이 필요하지 않습니다.

| 필요한 정보 | range 형태 |
| --- | --- |
| 키와 값 | `for key, value := range m` |
| 키만 | `for key := range m` |
| 값만 | `for _, value := range m` |

맵의 순회 순서는 지정되지 않으며 같은 맵이라도 매번 같다고 보장되지 않습니다. 따라서 키별 출력 줄을 고정하고 싶으면 순서를 별도로 만듭니다. [range 절](https://go.dev/ref/spec#For_range)을 참고합니다.

### 4.1 키를 슬라이스에 모아 정렬하기

키를 수집한 다음 `sort.Strings`로 정렬하고 그 순서로 조회합니다. `sort`는 표준 라이브러리입니다. `Strings`는 전달받은 문자열 슬라이스의 원소 순서를 직접 바꾸며 결과 슬라이스를 반환하지 않습니다.

```go
keys := make([]string, 0, len(counts))
for key := range counts {
	keys = append(keys, key)
}
sort.Strings(keys)
for _, key := range keys {
	fmt.Printf("%s: %d\n", key, counts[key])
}
```

이 발췌는 counts가 선언된 main 안에 넣고 파일의 import에 `"sort"`를 추가합니다. 다음 전체 예제에서 실제 형태를 확인합니다. [sort.Strings](https://pkg.go.dev/sort#Strings)의 계약을 참고합니다.

맵 자체를 정렬한 것이 아닙니다. 정렬된 키 목록을 출력 순서로 사용하는 것입니다. 나중에 맵에 키를 추가하면 기존 keys에 자동 반영되지 않으므로 필요한 시점에 다시 수집합니다.

## 5. 빈도 집계

여러 항목이 각각 몇 번 등장했는지 세는 작업을 **빈도 집계(frequency counting)**라고 합니다. 처음 본 키의 값이 0이라는 맵 규칙을 이용하면 별도 초기화 분기 없이 증가시킬 수 있습니다.

### 5.1 입력 순회와 정렬된 결과

저장 경로: [`examples/06-maps/frequency/main.go`](../../examples/06-maps/frequency/main.go). 전체 코드입니다.

```go
package main

import (
	"fmt"
	"sort"
)

func frequency(words []string) map[string]int {
	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}
	return counts
}
func sortedKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func main() {
	counts := frequency([]string{"go", "map", "go", "slice", "map", "go"})
	for _, key := range sortedKeys(counts) {
		fmt.Printf("%s: %d\n", key, counts[key])
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/06-maps/frequency
```

예상 출력:

```text
go: 3
map: 2
slice: 1
```

frequency는 값을 세고, sortedKeys는 키 순서를 만들며, main은 출력합니다. 계산과 표시 책임을 나누어 각각 확인할 수 있습니다.

| 입력 순서 | 현재 단어 | 그 단어의 새 횟수 |
| --- | --- | --- |
| 1 | go | 1 |
| 2 | map | 1 |
| 3 | go | 2 |
| 4 | slice | 1 |
| 5 | map | 2 |
| 6 | go | 3 |

집계 함수는 순서 없는 맵을 반환합니다. 출력 함수가 키를 정렬하므로 예상 출력은 매번 같습니다. 테스트에서도 무작위로 보이는 순회 결과를 우연히 정답으로 저장하지 않습니다.

입력이 nil이나 빈 슬라이스여도 make한 빈 맵을 반환합니다. 이 예제의 함수는 빈 결과도 바로 쓸 수 있는 맵으로 제공한다는 계약을 택했습니다.


### 5.2 항목 수와 전체 횟수 구분하기

위 입력의 전체 개수는 6이지만 키의 개수는 3입니다. `len(counts)`는 서로 다른 항목 수입니다.

| 질문 | 계산 |
| --- | --- |
| 총 몇 개를 입력했나? | `len(words)` |
| 서로 다른 항목은 몇 개인가? | `len(counts)` |
| go가 몇 번 나왔나? | `counts["go"]` |
| 실제로 go라는 키가 있는가? | `_, ok := counts["go"]` |

빈 문자열도 문자열 키로 사용할 수 있고 `"Go"`와 `"go"`는 다른 문자열입니다. 공백 제거, 대소문자 통일 같은 정규화 규칙은 맵이 자동으로 적용하지 않습니다. 다음 07장에서 텍스트 처리와 함께 설계합니다.

항목을 읽는 과정에서 키를 바꿔 정규화하려면 그 규칙을 입력 계약으로 명시합니다. 원문을 보존할 필요가 있는지와 집계 기준이 같은지도 분리해서 생각합니다.

## 6. 집합처럼 사용하기

**집합(set)**은 특정 항목이 있는지를 나타내는 모음입니다. Go에 이 장에서 사용할 별도의 내장 set 타입은 없지만, `map[string]bool`을 이용해 방문 여부를 저장할 수 있습니다.

이 설계에서는 등록한 키의 값을 항상 true로 둡니다. 없는 키는 조회 시 false이므로 “처음 보는가?”를 판단하기 좋습니다.

### 6.1 첫 등장 순서로 중복 제거

저장 경로: [`examples/06-maps/sets/main.go`](../../examples/06-maps/sets/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func uniqueNames(names []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}
func main() {
	names := []string{"Mina", "Jin", "Mina", "Sol", "Jin"}
	fmt.Println("처음 등장 순서:", uniqueNames(names))
	fmt.Println("원본:", names)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/06-maps/sets
```

예상 출력:

```text
처음 등장 순서: [Mina Jin Sol]
원본: [Mina Jin Mina Sol Jin]
```

seen은 이미 본 이름을 확인하는 용도이고 result는 출력 순서를 저장하는 용도입니다. 결과를 맵의 키 순회로 만들지 않고 입력을 순회하며 append하므로 첫 등장 순서가 유지됩니다.

| 이름 | 이미 봤나? | result |
| --- | --- | --- |
| Mina | 아니오 | `[Mina]` |
| Jin | 아니오 | `[Mina Jin]` |
| Mina | 예 | `[Mina Jin]` |
| Sol | 아니오 | `[Mina Jin Sol]` |
| Jin | 예 | `[Mina Jin Sol]` |

원본 슬라이스를 덮어쓰지 않으므로 입력 순서와 중복은 원본에 그대로 남습니다. 결과의 원소를 바꾸어도 원본의 문자열 원소 자리는 바뀌지 않습니다.


### 6.2 false를 저장하면 존재와 의미가 갈라진다

다음은 main 안의 독립된 발췌입니다.

```go
enabled := map[string]bool{"chat": false}
value, exists := enabled["chat"]
fmt.Println(value, exists)
delete(enabled, "chat")
value, exists = enabled["chat"]
fmt.Println(value, exists)
```

예상 출력:

```text
false true
false false
```

집합 구현에서 true만 저장한다는 약속이 중요한 이유입니다. 일반적인 설정 맵에서는 false도 유효한 값이므로 존재 여부와 설정값을 구분해야 합니다.

항목을 집합에서 제거하려면 delete를 사용합니다. 단순히 false를 저장하면 키는 남아 len에도 포함됩니다. 이후 코드가 어떤 기준으로 존재를 판단하는지에 따라 결과가 달라집니다.

구조체를 배운 뒤에는 `map[string]struct{}` 같은 집합 표현도 볼 수 있습니다. 이번 장에서는 이미 아는 bool과 존재 계약으로 동작을 이해합니다.

## 7. 키 하나에 여러 값 묶기

맵의 값이 꼭 정수 하나일 필요는 없습니다. `map[string][]int`는 문자열 키마다 정수 슬라이스를 저장합니다. 분류별로 여러 기록을 모을 때 사용할 수 있습니다.

### 7.1 짝수와 홀수 그룹

저장 경로: [`examples/06-maps/groups/main.go`](../../examples/06-maps/groups/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func groupParity(values []int) map[string][]int {
	groups := make(map[string][]int)
	for _, value := range values {
		key := "odd"
		if value%2 == 0 {
			key = "even"
		}
		groups[key] = append(groups[key], value)
	}
	return groups
}
func main() {
	groups := groupParity([]int{1, 2, 4, 5, 0})
	fmt.Println("짝수:", groups["even"])
	fmt.Println("홀수:", groups["odd"])
	fmt.Println("없는 그룹 길이:", len(groups["unknown"]))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/06-maps/groups
```

예상 출력:

```text
짝수: [2 4 0]
홀수: [1 5]
없는 그룹 길이: 0
```

없는 키의 값은 슬라이스 영값인 nil입니다. nil 슬라이스에 append할 수 있으므로 처음 원소를 넣기 전에 빈 슬라이스를 따로 만들 필요가 없습니다. 하지만 바깥 맵 groups 자체는 make로 준비해야 합니다.

`groups[key] = append(groups[key], value)`는 결과 슬라이스를 다시 그 키에 저장합니다. append 결과를 지역 변수에만 두고 맵에 돌려놓지 않으면 맵 항목의 길이는 갱신되지 않습니다.

| 입력 | 선택 키 | 해당 그룹 |
| --- | --- | --- |
| 1 | odd | `[1]` |
| 2 | even | `[2]` |
| 4 | even | `[2 4]` |
| 5 | odd | `[1 5]` |
| 0 | even | `[2 4 0]` |

각 그룹 안의 순서는 입력 순서입니다. 맵 키의 순서와 그룹 내부 슬라이스의 순서는 별개입니다. main은 even, odd라는 정해진 키를 직접 조회하여 출력 순서를 고정합니다.


### 7.2 바깥 준비와 안쪽 영값

| 자료형 | 없는 키를 읽은 값 | 그 값에 가능한 다음 처리 |
| --- | --- | --- |
| `map[string]int` | 0 | 증가 후 다시 저장 |
| `map[string]bool` | false | 방문 여부 판단 |
| `map[string][]int` | nil 슬라이스 | append 후 다시 저장 |
| `map[string]map[string]int` | nil 맵 | 쓰기 전에 안쪽 make 필요 |

슬라이스의 nil과 맵의 nil은 추가 동작이 다릅니다. nil 슬라이스에는 append할 수 있지만 nil 맵에는 대입할 수 없습니다. 마지막 형태는 12절 선택 학습에서 확인합니다.

`groups["unknown"]`을 조회했다고 unknown 키가 생기지는 않습니다. `len(groups["unknown"])`이 0인 것과 그 키가 실제 존재한다는 것은 다른 정보입니다.

빈 그룹을 명시적으로 보존하고 싶으면 그룹 키를 빈 슬라이스와 함께 미리 등록하는 계약을 정할 수 있습니다. 이 예제는 실제 값을 만난 그룹만 등록합니다.

## 8. 맵 대입과 함수 전달

맵도 함수에 값으로 전달합니다. 다만 맵 값을 복사하면 같은 맵 데이터를 사용할 수 있습니다. `other := original`은 항목의 독립 복사가 아닙니다. [값 전달 FAQ](https://go.dev/doc/faq#pass_by_value)를 참고합니다.

다음은 main 안의 독립된 발췌입니다.

```go
original := map[string]int{"go": 2}
other := original
other["go"] = 9
delete(other, "go")
fmt.Println(len(original), len(other))
```

출력은 `0 0`입니다. 대입 뒤 두 변수는 같은 맵 데이터에 접근합니다. 항목 변경과 삭제가 양쪽에서 관찰됩니다.

### 8.1 변경 함수와 복사 함수의 계약

저장 경로: [`examples/06-maps/map-functions/main.go`](../../examples/06-maps/map-functions/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func increment(counts map[string]int, key string) bool {
	if counts == nil {
		return false
	}
	counts[key]++
	return true
}
func cloneCounts(counts map[string]int) map[string]int {
	if counts == nil {
		return nil
	}
	cloned := make(map[string]int, len(counts))
	for key, value := range counts {
		cloned[key] = value
	}
	return cloned
}
func main() {
	counts := map[string]int{"Go": 2}
	ok := increment(counts, "Go")
	fmt.Printf("변경: %d, 성공: %t\n", counts["Go"], ok)
	cloned := cloneCounts(counts)
	cloned["Go"] = 99
	fmt.Println("원본:", counts["Go"])
	fmt.Println("복사:", cloned["Go"])
	fmt.Println("nil 변경:", increment(nil, "Go"))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/06-maps/map-functions
```

예상 출력:

```text
변경: 3, 성공: true
원본: 3
복사: 99
nil 변경: false
```

increment는 전달받은 맵을 변경하는 함수입니다. nil이면 쓰지 않고 false를 반환합니다. 호출한 쪽은 변경 성공 여부를 확인할 수 있습니다.

cloneCounts는 새 맵에 각 키·정수 값을 복사합니다. 결과의 Go 값을 99로 바꾸어도 원본은 3으로 남습니다. nil 입력은 nil로 유지하고, 생성된 빈 입력은 생성된 빈 결과로 반환합니다.

| 함수 | 원본 변경 | 결과 또는 실패 |
| --- | --- | --- |
| increment | 키 값을 증가시킴 | nil이면 false |
| cloneCounts | 변경하지 않음 | 별도 맵, nil은 nil로 유지 |

정수 값만 담았으므로 항목 값까지 독립적입니다. 값이 슬라이스라면 맵 항목을 복사해도 안쪽 원소가 공유될 수 있다는 점은 10절에서 확인합니다.


### 8.2 지역 매개변수에 재대입하는 경우

다음은 개인 실습 파일에서 실행할 수 있는 전체 코드입니다.

```go
package main

import "fmt"

func reset(counts map[string]int) {
	counts = make(map[string]int)
	counts["new"] = 1
}

func main() {
	counts := map[string]int{"go": 2}
	reset(counts)
	fmt.Println("go:", counts["go"])
	fmt.Println("new:", counts["new"])
}
```

```sh
go run ./_practice/06-maps/lab
```

예상 출력:

```text
go: 2
new: 0
```

reset은 자기 매개변수에 새 맵 값을 대입했습니다. 호출자의 counts 변수는 여전히 원래 맵을 사용합니다. “맵이므로 함수 안의 모든 대입이 호출자에 반영된다”는 설명은 틀립니다.

호출자의 맵 전체를 교체하려면 새 맵을 반환하고 호출한 쪽에서 받아 대입할 수 있습니다. 모든 기존 항목을 지우는 것이 목적이라면 clear를 사용하지만, 이때는 그 데이터를 공유하는 다른 변수에서도 삭제가 보입니다.

## 9. 순회하면서 변경할 때

### 9.1 반복 변수는 값의 복사본

아래는 main 안의 독립된 발췌입니다.

```go
counts := map[string]int{"go": 2}
for _, value := range counts {
	value++
}
fmt.Println(counts["go"])
for key := range counts {
	counts[key]++
}
fmt.Println(counts["go"])
```

예상 출력:

```text
2
3
```

첫 반복은 지역 value만 바꿉니다. 두 번째 반복은 키로 맵 항목을 직접 갱신합니다. 슬라이스에서 반복 변수만 증가시켰을 때와 같은 구분입니다.

### 9.2 삭제와 새 항목 추가

순회 중 아직 방문하지 않은 항목을 삭제하면 그 항목은 해당 순회에서 나오지 않습니다. 순회 중 새로 추가한 항목은 나올 수도 있고 나오지 않을 수도 있습니다. [range 명세](https://go.dev/ref/spec#For_range)가 이 동작을 정합니다.

따라서 새 항목까지 정확히 한 번 처리해야 하는 작업을 맵 순회 중 추가에 의존해서 만들지 않습니다. 먼저 처리 대상을 슬라이스로 모은 뒤 별도 단계에서 변경할 수 있습니다.

다음은 0인 항목을 지우는 main 안의 독립된 발췌입니다.

```go
counts := map[string]int{"go": 2, "map": 0, "slice": 0}
for key, count := range counts {
	if count == 0 {
		delete(counts, key)
	}
}
fmt.Println(len(counts), counts["go"])
```

출력은 `1 2`입니다. 현재 방문한 항목의 값으로 삭제 여부를 정하며, 아직 방문하지 않은 다른 항목은 직접 삭제하지 않습니다. 방문 순서와 관계없이 같은 최종 키가 남습니다.

## 10. 비교와 복사의 범위

### 10.1 맵끼리 ==로 비교할 수 없다

맵은 nil과 비교할 수 있지만 다른 맵과 `==`로 내용 비교를 할 수 없습니다. 내용이 같아 보여도 언어가 허용하는 비교 연산인지 구분해야 합니다.

다음은 **함수 선언 발췌**입니다. main 바깥에 둡니다.

```go
func equalCounts(a, b map[string]int) bool {
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
```

두 맵의 항목 수가 같아도 키가 다를 수 있습니다. `a`에 a:0, `b`에 b:0이 있을 때 값만 조회하면 모두 0이므로 잘못 같다고 판단할 수 있습니다. exists 검사도 필요합니다.

이 함수는 nil과 생성된 빈 맵을 같은 항목 집합으로 취급합니다. nil 상태를 구별해야 한다면 별도 조건을 계약에 추가합니다. [비교 연산자](https://go.dev/ref/spec#Comparison_operators)를 참고합니다.

### 10.2 값이 슬라이스라면 한 단계 더 생각하기

다음은 main 안의 독립된 발췌입니다.

```go
original := map[string][]int{"go": {10, 20}}
copied := make(map[string][]int)
for key, values := range original {
	copied[key] = values
}
copied["go"][0] = 99
fmt.Println(original["go"])
```

출력은 `[99 20]`입니다. 맵 항목은 별도지만 값인 슬라이스가 같은 원소를 공유합니다. map 리터럴의 `"go": {10, 20}`은 값 타입이 이미 `[]int`로 알려져 있어 안쪽 타입 표기를 생략한 형태입니다.

안쪽 정수까지 독립적으로 만들려면 반복 안에서 길이를 맞춘 새 슬라이스를 만들고 copy한 뒤 맵에 넣습니다. 05장의 슬라이스 복사 규칙이 그대로 적용됩니다.

```go
for key, values := range original {
	cloned := make([]int, len(values))
	copy(cloned, values)
	copied[key] = cloned
}
```

이 발췌는 앞 프로그램의 첫 복사 반복을 대체합니다. 이제 결과의 첫 원소를 바꾸어도 원본은 `[10 20]`으로 유지됩니다. 이 구현은 nil인 안쪽 슬라이스를 빈 슬라이스로 바꿀 수 있으므로 nil 보존까지 필요하면 추가로 검사합니다.

## 11. 직접 작성하기: 가격표로 주문 계산

상품 가격표와 수량표를 서로 다른 맵으로 받고 합계를 계산합니다. 다음은 학습용 계약입니다.

- 주문 항목 수는 최대 10개입니다.
- 모든 주문 상품은 가격표에 등록되어 있어야 합니다.
- 사용되는 가격은 1~100,000원, 수량은 0~10개입니다.
- 조건을 벗어나면 `0, false`를 반환합니다.
- 빈 주문은 합계 0의 정상 주문입니다.
- 두 입력 맵을 변경하지 않습니다.

항목 수와 가격·수량을 제한하므로 최대 합계는 10,000,000원으로 기본 int의 최소 범위 안입니다. 이 예제의 제한이며 실제 결제 시스템의 금액 설계를 대신하는 규칙은 아닙니다.

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("lab")) {
    $practiceDir = Join-Path "_practice/06-maps" $practiceName
    New-Item -ItemType Directory -Force $practiceDir | Out-Null
    $practiceMain = Join-Path $practiceDir "main.go"
    if (Test-Path -LiteralPath $practiceMain) {
        Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
    } else {
        Copy-Item examples/01-first-program/hello/main.go $practiceMain
    }
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
for practice_name in lab; do
    practice_dir="_practice/06-maps/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

한 폴더에는 하나의 `main`만 둡니다. 새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice/`는 공통 `go test ./...`에 포함되지 않으므로 개인 폴더를 직접 지정해서 실행하세요. [공통 연습 안내](../../exercises/README.md)를 참고합니다.


### 11.1 전체 프로그램

개인 파일 `_practice/06-maps/lab/main.go`에 작성합니다.

```go
package main

import (
	"fmt"
	"sort"
)

func orderTotal(prices, quantities map[string]int) (int, bool) {
	if len(quantities) > 10 {
		return 0, false
	}
	total := 0
	for item, quantity := range quantities {
		price, exists := prices[item]
		if !exists || price < 1 || price > 100000 || quantity < 0 || quantity > 10 {
			return 0, false
		}
		total += price * quantity
	}
	return total, true
}

func main() {
	prices := map[string]int{"pen": 1000, "book": 8000}
	quantities := map[string]int{"pen": 3, "book": 2}
	total, ok := orderTotal(prices, quantities)
	if !ok {
		fmt.Println("주문 확인")
		return
	}
	keys := make([]string, 0, len(quantities))
	for key := range quantities {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d개\n", key, quantities[key])
	}
	fmt.Printf("합계: %d원\n", total)
}
```

```sh
gofmt -w _practice/06-maps/lab/main.go
go run ./_practice/06-maps/lab
```

예상 출력:

```text
book: 2개
pen: 3개
합계: 19000원
```

가격은 이름으로 조회하고, 수량 맵의 항목을 순회하며 계산합니다. 합계를 구하는 순서는 중요하지 않지만 화면에 표시하는 순서는 sort로 정합니다.

### 11.2 경계 입력

| 변경 | 계산 결과 |
| --- | --- |
| 주문을 빈 맵으로 변경 | `0, true` |
| pen 수량을 0, book 수량을 0 | `0, true` |
| pen 수량을 -1 | `0, false` |
| pen 수량을 11 | `0, false` |
| 없는 note를 주문에 추가 | `0, false` |
| pen 가격을 0 | `0, false` |
| pen 가격을 100001 | `0, false` |

실패하면 `주문 확인` 한 줄만 출력합니다. 빈 주문에서는 상품 출력 없이 `합계: 0원` 한 줄을 출력합니다. 잘못된 입력을 중간에 만났어도 부분 합계를 성공 결과로 돌려주지 않습니다.

어떤 키에서 먼저 실패하는지에 대한 순서는 이 함수의 계약이 아닙니다. 나중에 구체적인 오류 메시지를 반환하려면 검사 순서를 고정할지, 오류 목록을 모을지 별도로 설계해야 합니다. 오류 타입은 11장에서 다룹니다.

## 12. 선택 학습: 맵의 값도 맵인 경우

언어별·요일별 방문 수처럼 기준이 두 개인 자료를 표현할 수 있습니다. 다만 바깥 맵을 만든 것으로 안쪽 맵까지 생기는 것은 아닙니다.

다음 전체 프로그램을 개인 파일에서 실행합니다.

```go
package main

import "fmt"

func main() {
	visits := make(map[string]map[string]int)
	language := "Go"
	day := "Mon"
	if visits[language] == nil {
		visits[language] = make(map[string]int)
	}
	visits[language][day]++
	visits[language][day]++
	fmt.Println("Go Mon:", visits["Go"]["Mon"])
	fmt.Println("없는 언어:", visits["Rust"]["Mon"])
}
```

```sh
go run ./_practice/06-maps/lab
```

예상 출력:

```text
Go Mon: 2
없는 언어: 0
```

Rust 조회에서는 바깥 조회가 nil 맵을 반환하고, 그 nil 맵의 Mon 조회가 0을 반환합니다. 읽기는 가능하지만 `visits["Rust"]["Mon"]++`는 안쪽 맵을 만들기 전에는 실패합니다.

| 준비 | 완료한 것 | 아직 필요한 것 |
| --- | --- | --- |
| 바깥 make | 언어별 값을 저장할 맵 | 각 언어의 안쪽 맵 |
| Go의 안쪽 make | Go의 요일별 맵 | 필요한 키·값의 실제 대입 |
| Mon 증가 | 항목 생성·갱신 | 없음 |

두 기준을 구조체 하나로 묶어 키로 쓰는 대안은 08장에서 생각할 수 있습니다. 어떤 구조가 읽기와 변경에 적합한지는 실제 조회 방식으로 선택합니다.

## 13. 키 타입과 흔한 오류

### 13.1 키는 비교할 수 있어야 한다

문자열과 정수는 맵 키로 사용할 수 있습니다. 정수만 든 배열도 비교 가능하므로 사용할 수 있습니다. 슬라이스·맵·함수는 맵 키로 사용할 수 없습니다. [맵 타입](https://go.dev/ref/spec#Map_types)이 키의 비교 가능 조건을 정합니다.

다음은 main 안의 독립된 발췌입니다.

```go
locations := map[[2]int]string{
	[2]int{1, 2}: "입구",
}
fmt.Println(locations[[2]int{1, 2}])
```

출력은 `입구`입니다. 여기서 `[2]int`는 두 정수로 구성된 키 값이며 길이도 타입의 일부입니다.

다음은 **의도적으로 컴파일에 실패하는 전체 코드**입니다.

```go
package main

func main() {
	values := make(map[[]int]string)
	_ = values
}
```

```sh
go run ./_practice/06-maps/lab
```

오류의 `invalid map key type []int`를 확인합니다. 키가 항상 정수 두 개인 계약이라면 배열을 검토할 수 있습니다. 길이가 달라지는 데이터를 억지로 문자열로 합칠 때는 서로 다른 입력이 같은 키가 되지 않는지 별도 설계가 필요합니다.

### 13.2 증상별 확인표

| 증상 | 원인 | 수정 방향 |
| --- | --- | --- |
| nil map 대입 panic | 선언만 하고 생성하지 않음 | 쓰기 전에 make 또는 리터럴 |
| 값 0을 미등록으로 처리 | 영값과 존재 혼동 | 두 결과 조회 |
| 출력 줄 순서가 달라짐 | range 순서에 의존 | 키 수집·정렬 뒤 출력 |
| 맵 복사 뒤 원본도 바뀜 | 맵 값만 대입 | 새 맵에 항목 복사 |
| 안쪽 목록이 함께 바뀜 | 슬라이스 값만 복사 | 필요한 깊이까지 복사 |
| len이 총 횟수와 다름 | 키 수와 값 합계 혼동 | 값을 순회하여 합산 |
| 반복 변수 수정이 반영되지 않음 | 지역 복사본만 변경 | 키로 항목 재대입 |
| 빈 그룹에 append했는데 저장 안 됨 | 반환 슬라이스를 맵에 넣지 않음 | `groups[key] = append(...)` |
| 중복 제거 순서가 달라짐 | 결과를 맵 순회로 만듦 | 입력 순회와 결과 슬라이스 병행 |
| map끼리 == 오류 | 내용 직접 비교 불가 | 길이·키 존재·값 비교 |

이 장의 예제는 순차 실행합니다. 같은 맵에 여러 고루틴이 동시에 접근하여 한쪽이 쓰는 경우의 동기화는 20장에서 다룹니다. 맵 자체가 이런 접근을 자동으로 안전하게 만드는 것으로 가정하지 않습니다.

## 14. 연습문제와 검증

[06장 연습문제](../../exercises/06-maps/README.md)에서 기본 6문제와 선택 A를 풉니다. 입력 계약을 먼저 읽고 순서·영값·변경 책임을 확인합니다.

| 문제 | 중심 계약 |
| --- | --- |
| 1. 존재와 영값 추적 | 조회는 등록하지 않음 |
| 2. 항목 빈도 | 키별 횟수와 정렬 출력 |
| 3. nil 맵 오류 수정 | 생성과 대입 구분 |
| 4. 고유 목록 | 첫 등장 순서와 입력 보존 |
| 5. 재고 출고 | 실패 시 변경하지 않음 |
| 6. 두 집계 합치기 | 같은 키 합산·입력 독립성 |
| 선택 A. 맵 내용 비교 | 0인 값과 없는 키의 차이 |

[해설](../../solutions/06-maps/README.md)과 비교하기 전, 빈 입력과 없는 키를 직접 실행합니다. 참고 구현은 다음 명령으로 확인합니다.

```sh
go test -count=1 ./examples/06-maps/... ./solutions/06-maps/...
```

검사는 기본 출력과 키의 존재, 실패 후 원본 유지, 빈 결과의 nil 계약, 병합 결과의 독립성을 포함합니다. 개인 파일에는 자동 적용되지 않으므로 문제별 실행 명령도 확인해야 합니다.

## 15. 요약과 자기 점검

- 맵은 정해진 타입의 키와 값을 연결하며 같은 키의 대입은 값을 갱신합니다.
- 없는 키의 조회는 영값을 반환하지만 항목을 만들지 않습니다.
- 존재 여부가 필요하면 value와 ok를 함께 받습니다.
- nil 맵은 읽을 수 있지만 쓰기 전에 생성해야 합니다.
- 맵 range의 방문 순서는 보장되지 않습니다.
- 입력 순서가 필요하면 슬라이스를 함께 사용하고, 정렬 순서가 필요하면 키를 정렬합니다.
- 맵 값의 대입은 항목의 독립 복사가 아닙니다.
- 함수의 항목 변경과 지역 맵 변수 재대입은 다릅니다.

### 확인 질문

1. len이 3인 재고 맵은 총 재고가 3개라는 뜻인가요?
2. 값이 0인 키와 없는 키를 어떻게 구분하나요?
3. 없는 키를 조회하면 len이 늘어나나요?
4. nil 맵에서 delete는 가능한데 대입은 왜 준비가 필요한가요?
5. 키를 정렬하면 맵 자체의 순서도 정렬되나요?
6. 처음 본 순서로 중복 제거할 때 맵과 슬라이스가 각각 맡는 일은 무엇인가요?
7. `groups[key] = append(groups[key], value)`에서 재대입을 빼면 무엇이 빠지나요?
8. 맵을 새로 만들고 슬라이스 값을 대입했을 때 안쪽 정수도 독립적인가요?
9. 함수를 호출한 뒤 nil 여부와 내용이 바뀌는 경우를 어떻게 구분하나요?
10. 같은 개수의 0 항목을 가진 두 맵은 반드시 같은가요?

<details>
<summary>답과 설명</summary>

1. 등록한 키가 3개라는 뜻입니다. 재고 합계는 각 값을 더해야 합니다.
2. `value, ok := m[key]`에서 ok를 확인합니다.
3. 아닙니다. 쓰기나 증가·대입을 해야 항목이 등록됩니다.
4. 삭제할 항목이 없는 읽기·삭제 동작은 정의되어 있지만 nil 맵에는 저장할 준비가 없습니다. make 또는 리터럴로 준비합니다.
5. 아닙니다. 별도로 정렬한 키 슬라이스를 출력 순서로 사용할 뿐입니다.
6. 맵은 이미 본 항목인지 확인하고 슬라이스는 첫 등장 순서를 저장합니다.
7. 새 길이를 가진 결과 슬라이스를 맵 항목에 저장하는 단계가 빠집니다.
8. 아닙니다. 슬라이스 값의 복사로 원소 저장 공간은 공유할 수 있습니다.
9. 항목 대입·delete·clear는 공유 데이터에 작용하고, 지역 매개변수에 새 맵을 대입하는 일은 호출자의 변수를 교체하지 않습니다.
10. 아닙니다. 키도 같아야 합니다. 존재 여부 확인 없이 값만 읽으면 없는 키의 0과 혼동합니다.

</details>

### 완료 체크리스트

- [ ] nil과 생성된 빈 맵을 구분했다.
- [ ] 영값과 존재 여부를 함께 확인했다.
- [ ] 순회 순서에 의존하지 않는 출력을 만들었다.
- [ ] 빈도·집합·그룹의 서로 다른 목적을 설명했다.
- [ ] 함수의 입력 변경 여부를 계약에 적었다.
- [ ] 결과 수정과 원본 유지 여부를 검사했다.
- [ ] 기본 문제 1~6과 경계 입력을 확인했다.

## 참고 자료

- [Go 명세: 맵 타입](https://go.dev/ref/spec#Map_types)
- [Go 명세: 인덱스 표현식](https://go.dev/ref/spec#Index_expressions)
- [Go 명세: 맵 원소 삭제](https://go.dev/ref/spec#Deletion_of_map_elements)
- [Go 명세: range](https://go.dev/ref/spec#For_range)
- [Go 명세: clear](https://go.dev/ref/spec#Clear)
- [Go 블로그: Go maps in action](https://go.dev/blog/maps)
- [sort.Strings](https://pkg.go.dev/sort#Strings)

확인일: 2026-11-09. Go 1.27.3에서 예제와 경계 입력을 확인했습니다. 기본 명령은 운영체제 공통이며 개인 폴더 준비는 셸별로 제공합니다.

## 다음 학습

다음 07장에서는 문자열의 바이트와 룬을 구분하고, 한글을 포함한 텍스트를 읽고 처리합니다. 맵 집계에 넣기 전 공백과 대소문자를 어떻게 다룰지도 연결합니다.

[이전: 05. 배열과 슬라이스](../05-arrays-slices/README.md) · [다음: 07. 문자열과 유니코드](../07-strings-unicode/README.md) · [전체 목차](../README.md) · [문제](../../exercises/06-maps/README.md) · [해설](../../solutions/06-maps/README.md)
