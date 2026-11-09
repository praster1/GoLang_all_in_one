# 13장 연습문제: 제네릭

[본문](../../book/13-generics/README.md)을 읽고 풉니다. 타입만 일반화한 뒤 빈 입력과 공유 상태를 잊지 않는 것이 이 장의 목표입니다.

## 준비와 검증 방법

모든 명령은 저장소 루트에서 실행합니다. 개인 코드는 `_practice/13-generics/` 아래 `last`, `unique`, `keys`, `group`, `partition`, `pair` 폴더로 나눕니다. 각 폴더에는 main.go와 필요한 `_test.go` 파일을 둡니다. 기존 코드는 덮어쓰지 않습니다.

```sh
go run ./_practice/13-generics/last
go test ./_practice/13-generics/last
```

새 go.mod나 go.work를 만들지 않습니다. 마지막 선택 문제의 메서드 자체 타입 매개변수는 Go 1.27 이상이 필요합니다. `go version`으로 실제 실행 버전을 확인하세요.

풀이 전 다음 계약을 적습니다.

| 항목 | 확인할 질문 |
| --- | --- |
| 타입 | 어떤 연산이 필요하고 어떤 제약이 맞는가? |
| 결과 | 입력 타입 이름을 유지해야 하는가? |
| 빈 입력 | nil과 non-nil 빈 값을 구별하는가? |
| 순서 | 입력 순서나 정렬 순서를 보장하는가? |
| 저장 공간 | 원본을 바꾸거나 공유하는가? |
| callback | 원소마다 몇 번 호출하고 nil을 허용하는가? |

정답 함수와 같은 이름을 썼다는 이유만으로 완료가 되지는 않습니다. 아래 계약을 만족하고 중요한 경계를 테스트하면 됩니다.

## 문제 1. 타입과 제약 판단

**수준: 확인**

다음 선언을 알고 있다고 가정합니다.

```go
func first[T any](values []T) (T, bool)
func index[T comparable](values []T, target T) int
type Integer interface { ~int | ~int64 }
func sum[T Integer](values []T) T

type Cents int64
type Labels []string
```

이 블록은 시그니처만 보여 주는 발췌입니다. 구현 없는 함수를 그대로 실행하지 않습니다.

다음 호출이 가능한지, 가능하면 타입 매개변수가 무엇인지 적으세요.

| 호출·선언 | 가능 여부 | 이유 또는 추론 타입 |
| --- | --- | --- |
| `first([]int{0})` | | |
| `first[string](nil)` | | |
| `first(nil)` | | |
| `index([]string{"a"}, "a")` | | |
| `index([][]int{{1}}, []int{1})` | | |
| `sum([]Cents{100, 200})` | | |
| `sum([]float64{1.5})` | | |
| `var value Integer` | | |

추가 질문:

- `type Exact interface { int64 }`라면 Cents도 포함되나요?
- comparable 제약에 인터페이스 타입을 사용했을 때 모든 동적 비교가 항상 안전한가요?
- `func filter[T any]([]T) []T`와 `func filter[S ~[]E, E any](S) S`가 Labels에 주는 결과 타입은 어떻게 다른가요?

완료 조건: 정적 타입 제약과 런타임 동적 값의 문제를 구별합니다. 실패하는 코드는 정상 예제 트리에 저장하지 않습니다.

<details>
<summary>힌트</summary>

nil만으로는 원소 타입을 추론할 정보가 부족합니다. 슬라이스는 == 비교를 지원하지 않습니다. 기반 타입 조건을 쓰는 `~`와 정확한 타입 조건을 구별합니다.

</details>

## 문제 2. 마지막 값 반환

**수준: 변형**

다음 함수를 작성합니다.

```go
func last[T any](values []T) (T, bool)
```

계약:

- 원소가 있으면 마지막 값과 true입니다.
- nil 또는 빈 입력이면 T의 영값과 false입니다.
- 입력을 수정하지 않습니다.
- 반환값은 마지막 원소 값의 일반적인 복사입니다. 깊은 복제를 수행하지 않습니다.

| 입력 | 결과 |
| --- | --- |
| []int{2,4,6} | 6, true |
| []int{0} | 0, true |
| 빈 []int | 0, false |
| []string{"a","b"} | b, true |
| 빈 []string | 빈 문자열, false |

완료 조건: 정수와 문자열의 영값을 하드코딩하는 타입 스위치 없이 구현합니다. 정상 영값과 비어 있는 입력을 구별하는 테스트를 작성합니다.

<details>
<summary>힌트 1</summary>

인덱스를 계산하기 전에 길이가 0인지 확인합니다. 빈 입력에서 len(values)-1은 유효한 인덱스가 아닙니다.

</details>

<details>
<summary>힌트 2</summary>

`var zero T`로 모든 허용 타입의 영값을 얻습니다. 모든 타입에서 0 또는 nil을 반환할 수 있는 것은 아닙니다.

</details>

## 문제 3. 첫 등장 순서의 중복 제거

**수준: 변형**

다음 함수를 작성합니다.

```go
func unique[S ~[]E, E comparable](values S) S
```

계약:

- 같은 값은 처음 등장한 것만 남깁니다.
- 남은 원소의 입력 순서를 유지합니다.
- 결과는 입력과 같은 정의된 슬라이스 타입 S입니다.
- 입력 슬라이스의 원소를 바꾸지 않습니다.
- 결과의 바깥 저장 공간은 입력과 독립적입니다.
- nil 입력은 nil 결과입니다.
- non-nil 빈 입력은 non-nil 빈 결과입니다.
- 테스트 입력 E는 int·string·정의된 int/string처럼 동적 비교 패닉이나 NaN의 특수성을 피한 값으로 한정합니다.

예상 결과:

```text
[2 1 2 3 1] -> [2 1 3]
[Go SQL Go] -> [Go SQL]
```

`type Names []string`을 만든 뒤 `var result Names = unique(Names{...})`가 컴파일되는지도 확인하세요.

완료 조건: 결과 첫 원소를 바꿔도 원본 첫 원소는 유지됩니다. 정렬해서 중복을 제거하는 방식은 요구된 첫 등장 순서를 바꾸므로 그대로는 맞지 않습니다.

<details>
<summary>힌트 1</summary>

이미 나온 값인지 확인하는 맵과 결과 슬라이스를 분리합니다. 맵은 존재 확인에만 사용하고 출력 순서는 입력 순회로 결정합니다.

</details>

<details>
<summary>힌트 2</summary>

`make(S, 0, len(values))`는 정의된 슬라이스 타입 S의 결과를 만듭니다. nil 입력은 먼저 따로 처리합니다.

</details>

## 문제 4. 맵의 키를 정렬해 반환

**수준: 응용**

다음 제약과 함수를 작성합니다.

```go
type Key interface { ~int | ~string }
func sortedKeys[K Key, V any](values map[K]V) []K
```

계약:

- 키를 오름차순으로 정렬하여 반환합니다.
- 입력 맵은 변경하지 않습니다.
- 값 타입 V에는 불필요한 comparable 제약을 붙이지 않습니다.
- nil/빈 맵은 non-nil 빈 슬라이스를 반환합니다.
- 정의된 정수·문자열 키 타입도 유지합니다.
- float 키를 지원하지 않습니다. 이번 문제의 정책 범위입니다.

| 입력 | 결과 |
| --- | --- |
| map[string]int{"b":2,"a":1} | [a b] |
| map[int]string{10:"ten",2:"two"} | [2 10] |
| nil map[string]int | non-nil, 길이 0 |

완료 조건: 키 수집 순서가 맵마다 달라질 수 있어도 최종 결과는 정렬되어야 합니다. map[string][]int처럼 비교 불가능한 값 타입을 가진 맵도 허용해야 합니다.

<details>
<summary>힌트</summary>

키만 슬라이스에 모으고 slices.Sort를 사용합니다. 맵의 값은 비교하거나 정렬하지 않으므로 V는 any면 충분합니다.

</details>

## 문제 5. 키 함수로 값 그룹화

**수준: 응용**

다음 함수를 작성합니다.

```go
func groupBy[T any, K comparable](
    values []T,
    key func(T) K,
) map[K][]T
```

계약:

- 원소마다 key를 한 번 호출합니다.
- 같은 키의 값은 입력 순서대로 묶습니다.
- nil/빈 입력도 non-nil 빈 맵을 반환합니다.
- 입력의 바깥 슬라이스를 변경하지 않습니다.
- 결과 그룹의 바깥 저장 공간은 입력의 저장 공간과 독립적입니다.
- key는 nil이 아닌 함수입니다.
- key가 반환하는 값은 맵 키로 런타임에 안전하게 사용할 수 있어야 합니다.
- 맵 바깥 키의 순서는 보장하지 않습니다. 표시할 때 따로 키를 정렬합니다.

입력 예:

```go
type Lesson struct {
    Language string
    Minutes int
}

[]Lesson{{"Go", 20}, {"SQL", 30}, {"Go", 40}}
```

Language로 묶으면 Go에는 20분·40분 순서의 두 항목, SQL에는 30분 한 항목이 들어갑니다. 별도 테스트에서는 정수의 짝수 여부를 bool 키로 사용해 보세요.

완료 조건: 그룹 안의 순서가 보존됩니다. 표시 문자열 비교만 하지 말고 각 키의 슬라이스 내용과 callback 호출 횟수도 확인합니다.

<details>
<summary>힌트 1</summary>

`result[k] = append(result[k], value)`는 처음 보는 키의 nil 슬라이스에도 사용할 수 있습니다. 결과 맵 자체는 make로 초기화합니다.

</details>

<details>
<summary>힌트 2</summary>

key(value)를 맵 읽기와 쓰기에서 각각 다시 호출하지 말고 지역 변수에 한 번 저장합니다. callback의 비용과 부작용이 중복되지 않게 합니다.

</details>

## 문제 6. 두 슬라이스로 분리

**수준: 응용**

다음 함수를 작성합니다.

```go
func partition[S ~[]E, E any](
    values S,
    keep func(E) bool,
) (S, S)
```

첫 결과는 keep이 true인 값, 둘째 결과는 false인 값입니다.

계약:

- 각 결과 안의 상대 순서는 입력과 같습니다.
- 모든 입력은 두 결과 중 정확히 한 곳에 들어갑니다.
- callback은 원소마다 정확히 한 번 호출합니다.
- nil 입력은 두 nil 결과입니다.
- non-nil 입력은 두 non-nil 결과이며, 한쪽에 값이 없어도 non-nil입니다.
- 입력과 두 결과의 바깥 저장 공간은 서로 독립적입니다.
- keep은 nil이 아닌 함수입니다.
- 원소 내부의 포인터까지 깊게 복제하지는 않습니다.

예상 결과:

```text
입력: [1 2 3 4]
짝수: [2 4]
홀수: [1 3]
```

완료 조건: 두 결과 중 하나를 수정해도 입력이나 반대 결과가 바뀌지 않습니다. 정의된 슬라이스 타입을 유지하는지도 검사합니다.

<details>
<summary>힌트 1</summary>

두 번 filter를 호출하면 callback이 원소마다 두 번 실행되어 계약에 어긋날 수 있습니다. 한 번의 순회에서 true/false 결과 중 하나에 추가합니다.

</details>

<details>
<summary>힌트 2</summary>

원본 슬라이스의 앞부분을 임시 결과로 재사용하면 입력을 변경하거나 결과들이 저장 공간을 공유할 수 있습니다. 독립된 두 슬라이스로 시작합니다.

</details>

## 선택 A. 수신자 타입과 메서드 타입 구별

**수준: 응용, Go 1.27 이상**

두 값을 보관하는 Pair를 만들고 다음 메서드를 구현합니다.

```go
type Pair[A, B any] struct {
    Left A
    Right B
}

func (pair Pair[A, B]) Swap() Pair[B, A]
func (pair Pair[A, B]) MapLeft[C any](convert func(A) C) Pair[C, B]
```

계약:

- Swap은 좌우 값과 그 타입 순서를 바꾼 새 Pair를 반환합니다.
- MapLeft는 왼쪽만 변환하고 오른쪽 값과 B 타입을 유지합니다.
- convert는 nil이 아닌 함수이며 한 번 호출합니다.
- 원본 Pair의 필드 값을 바꾸지 않습니다.

완료 조건: Swap은 수신자의 A·B만 사용하고, MapLeft는 메서드 자체의 C를 추가로 받는다는 차이를 설명합니다. `Pair[int,string]`에서 `Pair[string,string]`을 얻는 코드를 컴파일합니다.

<details>
<summary>힌트</summary>

반환할 타입의 대괄호 안도 값의 위치와 함께 바뀝니다. 값 수신자의 필드를 읽어 새 구조체 리터럴을 만듭니다. 메서드의 C 제약은 수신자 괄호가 아니라 메서드 이름 뒤에 적습니다.

</details>

## 완료 체크리스트

- [ ] 필요한 연산을 기준으로 제약을 선택했습니다.
- [ ] nil·빈 입력·정상 영값을 구별했습니다.
- [ ] 정의된 타입을 유지해야 하는 결과를 확인했습니다.
- [ ] 정렬과 입력 순서 보존 중 요구된 계약을 지켰습니다.
- [ ] 원본 수정·바깥 저장 공간 공유·내부 포인터 공유를 구별했습니다.
- [ ] Go 1.27 제네릭 메서드와 인터페이스의 제한을 구별했습니다.

[본문](../../book/13-generics/README.md) · [해설](../../solutions/13-generics/README.md)
