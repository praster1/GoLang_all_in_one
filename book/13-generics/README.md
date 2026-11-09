# 13. 제네릭

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초에서 응용 | [12. 자동 테스트](../12-testing/README.md), 슬라이스·맵·인터페이스 | Go, 편집기, 터미널 |

정수 슬라이스에서 첫 값을 얻는 함수와 문자열 슬라이스에서 첫 값을 얻는 함수는 처리 순서가 같습니다. 입력이 비었는지 확인하고 첫 값을 반환합니다. 다른 것은 원소 타입뿐입니다. 이런 반복을 타입 정보를 유지하면서 줄이는 방법이 **제네릭(generics)**입니다.

제네릭은 모든 값을 any로 바꾸는 기술이 아닙니다. 호출할 때 정해지는 타입을 매개변수로 받아, 그 타입에 허용되는 연산만 사용하도록 컴파일 단계에서 검사하는 방식입니다. 이번 장은 “더 추상적인 코드”를 만드는 데보다 **같은 알고리즘과 다른 타입을 구별하고 필요한 범위만 일반화하는 것**에 초점을 둡니다.

## 학습 목표

- 타입 매개변수와 타입 인자를 구별하며 제네릭 함수의 선언·호출을 읽는다.
- any·comparable·타입 집합·`~`가 허용하는 연산과 타입 범위를 설명한다.
- 정의된 슬라이스 타입을 보존하고 입력·결과의 공유 여부를 명시한다.
- 제네릭 타입과 Go 1.27의 제네릭 메서드를 구분하여 구현한다.
- 표준 라이브러리의 제네릭 함수를 활용하고 경계·오류·타입별 동작을 테스트한다.

## 사전 지식과 준비

슬라이스의 복사와 공유, 맵의 키 조건, 함수 값, 구조체와 메서드, 인터페이스, 테스트를 사용합니다. `any`와 동적 타입은 09장, 테스트 비교와 사례 선택은 12장에서 다뤘습니다.

모든 명령은 **저장소 루트**에서 실행합니다. 루트의 go.mod를 사용하며 외부 의존성을 추가하지 않습니다.

| 기능 | 처음 도입된 계열 또는 이 장의 조건 |
| --- | --- |
| 타입 매개변수·제네릭 함수·제네릭 타입 | Go 1.18 이상 |
| 표준 cmp·slices 패키지 | Go 1.21 이상 |
| 메서드 자체의 별도 타입 매개변수 | **Go 1.27 이상** |
| 이 저장소의 학습 환경 | 루트 go.mod 기준 Go 1.27 |

특히 오래된 자료의 “Go 메서드는 자기 타입 매개변수를 가질 수 없다”는 설명을 현재 버전에 그대로 적용하지 않습니다. **Go 1.27은 제네릭 메서드를 지원합니다. 다만 인터페이스의 메서드 선언에는 자체 타입 매개변수를 붙일 수 없습니다.** [Go 1.27 언어 변경](https://go.dev/doc/go1.27#language)을 기준으로 10절에서 실제 컴파일되는 코드를 확인합니다.

| 예제 | 중심 주제 | 실행 명령 |
| --- | --- | --- |
| [first](../../examples/13-generics/first/main.go) | 타입 인자·추론·영값 | `go run ./examples/13-generics/first` |
| [index](../../examples/13-generics/index/main.go) | comparable과 비교 | `go run ./examples/13-generics/index` |
| [sum](../../examples/13-generics/sum/main.go) | 타입 집합과 기반 타입 | `go run ./examples/13-generics/sum` |
| [filter](../../examples/13-generics/filter/main.go) | 정의된 슬라이스 타입 보존 | `go run ./examples/13-generics/filter` |
| [transform](../../examples/13-generics/transform/main.go) | 입력·결과 타입이 다른 변환 | `go run ./examples/13-generics/transform` |
| [stack](../../examples/13-generics/stack/main.go) | 제네릭 구조체와 수신자 | `go run ./examples/13-generics/stack` |
| [generic-methods](../../examples/13-generics/generic-methods/main.go) | Go 1.27 메서드의 타입 매개변수 | `go run ./examples/13-generics/generic-methods` |
| [ordered](../../examples/13-generics/ordered/main.go) | 순서 제약과 NaN 정책 | `go run ./examples/13-generics/ordered` |
| [stdlib](../../examples/13-generics/stdlib/main.go) | slices·cmp 사용 | `go run ./examples/13-generics/stdlib` |

## 1. 같은 알고리즘에서 타입만 달라질 때

### 1.1 두 함수의 공통 부분 찾기

정수 슬라이스와 문자열 슬라이스에서 첫 값을 읽는 함수를 따로 작성하면 다음처럼 될 수 있습니다. 아래는 개념 비교를 위한 함수 선언 발췌입니다.

```go
func firstInt(values []int) (int, bool) {
    if len(values) == 0 {
        return 0, false
    }
    return values[0], true
}

func firstString(values []string) (string, bool) {
    if len(values) == 0 {
        return "", false
    }
    return values[0], true
}
```

두 함수의 분기와 인덱스 접근은 같습니다. 달라지는 것은 `int`와 `string`, 빈 입력에서 반환할 영값입니다. 제네릭 함수에서는 이 타입 자리에 이름을 붙입니다.

모든 중복이 타입 매개변수를 도입할 이유는 아닙니다. 한 함수는 금액을 검증하고 다른 함수는 이름을 정리한다면 겉으로 반복문이 비슷해도 의미가 다릅니다. 공통 알고리즘과 공통 계약이 있는지 먼저 확인합니다.

### 1.2 첫 제네릭 함수

저장 경로: [`examples/13-generics/first/main.go`](../../examples/13-generics/first/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

func first[T any](values []T) (T, bool) {
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[0], true
}
func main() {
	number, ok := first([]int{7, 9})
	fmt.Println("정수:", number, ok)
	word, ok := first([]string{"Go", "학습"})
	fmt.Println("문자열:", word, ok)
	missing, ok := first[int](nil)
	fmt.Println("빈 입력:", missing, ok)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/first
```

예상 출력:

```text
정수: 7 true
문자열: Go true
빈 입력: 0 false
```


`T`는 타입 매개변수(type parameter)입니다. `first[int]`의 `int`는 실제로 선택한 타입 인자(type argument)입니다. 일반 함수의 매개변수와 인자를 구별했던 것처럼, 타입 수준에서도 선언 자리와 사용 값을 구별합니다.

| 선언 부분 | 의미 |
| --- | --- |
| `first` | 함수 이름 |
| `[T any]` | T라는 타입 매개변수, 제약은 any |
| `values []T` | 선택된 T의 슬라이스를 입력으로 받음 |
| `(T, bool)` | 같은 T의 값과 존재 여부를 반환 |

`first([]int{7, 9})`에서는 T가 int입니다. 함수의 결과를 int 변수에 바로 사용할 수 있고, any로 받은 뒤 타입 단언을 할 필요가 없습니다. [타입 매개변수 선언](https://go.dev/ref/spec#Type_parameter_declarations)과 [공식 제네릭 입문](https://go.dev/doc/tutorial/generics)을 참고합니다.

### 1.3 T의 영값 만들기

T는 int일 수도, string일 수도, 구조체·슬라이스일 수도 있습니다. 모든 경우에 `return 0`이나 `return nil`이 맞지는 않습니다.

```go
var zero T
return zero, false
```

변수를 초기값 없이 선언하면 그 타입의 영값을 얻습니다. int라면 0, string이라면 빈 문자열, 포인터라면 nil입니다. 빈 입력인지 여부는 bool이 알려 주므로 정상 첫 값이 영값인 경우와 구별됩니다.

| 입력 | 값 | bool |
| --- | --- | --- |
| []int{0} | 0 | true |
| 빈 []int | 0 | false |
| []string{""} | 빈 문자열 | true |
| 빈 []string | 빈 문자열 | false |

## 2. 타입 추론과 명시적인 타입 인자

### 2.1 인자에서 타입을 알 수 있는 경우

`first([]int{7, 9})`는 슬라이스 원소 타입을 통해 T를 추론합니다. `first[int]([]int{7, 9})`처럼 명시적으로 적어도 같은 의미입니다. 구체 타입을 적용해 사용하는 과정을 인스턴스화(instantiation)라고 부릅니다. [Instantiations](https://go.dev/ref/spec#Instantiations)를 참고합니다.

다음 발췌는 앞 예제의 main에서 사용할 수 있습니다.

```go
number, ok := first[int]([]int{7, 9})
fmt.Println(number, ok)
```

명시적 표기가 항상 더 좋은 것은 아닙니다. 호출 인자가 타입을 충분히 보여 주면 생략하는 편이 읽기 쉽고, 타입을 알 수 없는 위치에서는 명시합니다.

### 2.2 nil만으로는 원소 타입을 알 수 없습니다

`first(nil)`은 어떤 []T를 의도했는지 정보가 부족합니다. `first[int](nil)` 또는 타입이 있는 nil 슬라이스를 전달합니다.

```go
var values []int
number, ok := first(values)
```

nil은 많은 타입의 값으로 사용할 수 있지만 모든 nil이 같은 타입은 아닙니다. 런타임 값의 부재와 컴파일 시 타입 정보의 부재를 나누어 생각하세요.

### 2.3 반환받을 변수만으로 호출 타입이 정해지지는 않습니다

다음은 **의도적으로 컴파일되지 않는 전체 프로그램**입니다.

```go
package main

func zero[T any]() T {
    var value T
    return value
}

func main() {
    var value int = zero()
    _ = value
}
```

호출 인자가 없어서 T를 추론할 정보가 없습니다. Go 1.27에서 실행하면 `cannot infer T`에 해당하는 오류를 확인할 수 있습니다. `zero[int]()`로 명시합니다.

Go 1.27에는 제네릭 **함수 값을 대입**할 때의 추론 확장이 있지만, 이것을 모든 함수 호출의 반환 문맥 추론으로 확대하면 안 됩니다. 규칙은 [Type inference](https://go.dev/ref/spec#Type_inference)와 [Go 1.27 변경](https://go.dev/doc/go1.27#language)을 확인합니다.

### 2.4 정의된 타입도 구체 타입입니다

`type StudentID int`를 정의하고 []StudentID를 전달하면 T는 StudentID입니다. 함수가 자동으로 int로 바꾸는 것이 아닙니다.

제네릭의 장점은 이 타입 관계를 유지할 수 있다는 것입니다. 값의 표현이 int와 같더라도 새로 정의한 타입이라는 정보를 결과에서도 보존할 수 있습니다.

## 3. 제약은 허용되는 연산을 정합니다

### 3.1 any는 무엇이든 할 수 있다는 뜻이 아닙니다

`any`는 모든 타입을 허용하는 제약입니다. 하지만 허용하는 타입이 많을수록 공통으로 보장되는 연산은 적어집니다.

`first`는 원소 값을 더하거나 비교하지 않고 그대로 반환합니다. 어떤 T에서도 가능한 일을 하므로 any로 충분합니다. 반면 `a+b`나 `a==b`를 하려면 그 연산이 가능한 타입만 허용하는 제약이 필요합니다.

다음은 **의도적으로 컴파일되지 않는 함수**입니다.

```go
func same[T any](a, b T) bool {
    return a == b
}
```

T에는 슬라이스나 맵 같은 비교 불가능 타입도 포함될 수 있습니다. 컴파일러는 제약에 속한 타입들에서 필요한 연산이 성립하는지 확인합니다. 현재 실제 호출이 int뿐이라고 이 일반 선언의 제약을 무시하지 않습니다.

### 3.2 ==가 필요하면 comparable

저장 경로: [`examples/13-generics/index/main.go`](../../examples/13-generics/index/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

func index[T comparable](values []T, target T) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

type Code int

func main() {
	fmt.Println("정수:", index([]int{3, 7, 3}, 3))
	fmt.Println("문자열:", index([]string{"Go", "한글"}, "한글"))
	fmt.Println("없음:", index([]Code{1, 2}, Code(9)))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/index
```

예상 출력:

```text
정수: 0
문자열: 1
없음: -1
```


`comparable`은 비교 가능한 타입을 위한 미리 선언된 제약입니다. 정수, 문자열, 포인터, 그리고 필드 조건을 만족하는 배열·구조체 등을 사용할 수 있습니다. 슬라이스나 맵을 이 함수의 원소 타입으로 직접 사용하는 것은 허용되지 않습니다. [Comparison operators](https://go.dev/ref/spec#Comparison_operators), [Satisfying a type constraint](https://go.dev/ref/spec#Satisfying_a_type_constraint)를 참고합니다.

index는 첫 번째 일치 위치를 반환하고 없으면 -1입니다. 빈 입력도 -1입니다. 반환 위치가 0인 경우가 정상 첫 원소와 일치한다는 점을 테스트합니다.

### 3.3 비교 가능과 정렬 가능은 다릅니다

bool과 포인터는 == 비교가 가능하지만 `<`로 크기를 비교할 수는 없습니다. comparable을 썼다고 모든 정렬 연산이 가능해지는 것은 아닙니다.

| 하려는 일 | 필요한 제약의 방향 |
| --- | --- |
| 값을 저장·전달 | any |
| == 비교, 맵 키 | comparable |
| 숫자·문자열의 순서 비교 | cmp.Ordered 또는 목적에 맞는 타입 집합 |
| 특정 행동 호출 | 필요한 메서드가 있는 인터페이스 제약 |

사용하는 연산에 맞춰 가장 작은 요구를 표현합니다. 비교하지 않는 함수에 comparable을 붙이면 슬라이스 등 유효한 입력을 불필요하게 제한합니다.

### 3.4 인터페이스 타입을 넣을 때의 런타임 경계

현재 Go에서는 `any` 같은 인터페이스 타입도 특정 규칙 아래 comparable 제약을 만족하는 타입 인자로 사용할 수 있습니다. 그러나 인터페이스 안에 비교 불가능한 동적 값이 들어 있으면 실제 비교는 패닉을 일으킬 수 있습니다.

예를 들어 `index[any]`의 입력에 `[]int`를 담은 인터페이스가 있고 같은 동적 슬라이스 타입과 비교하게 되면 09장의 인터페이스 비교 문제가 다시 나타납니다. 제네릭이 동적 값의 비교 가능성까지 마법처럼 바꾸는 것은 아닙니다.

기본 예제는 int·string·정의된 정수 타입으로 실행합니다. 인터페이스를 타입 인자로 허용하는 공개 API를 만든다면 동적 값의 계약도 설명해야 합니다. “comparable이면 모든 경우에 패닉이 절대 없다”라고 단정하지 않습니다.

## 4. 타입 집합과 기반 타입

### 4.1 어떤 숫자 타입을 허용할지 적기

서로 다른 정수 타입의 합계를 내는 알고리즘은 + 연산이 필요합니다. 다음 제약은 int 또는 int64를 기반으로 하는 타입을 허용합니다.

저장 경로: [`examples/13-generics/sum/main.go`](../../examples/13-generics/sum/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

type Integer interface{ ~int | ~int64 }

func sum[T Integer](values []T) T {
	var total T
	for _, value := range values {
		total += value
	}
	return total
}

type Cents int64

func main() {
	fmt.Println("int:", sum([]int{1, 2, 3}))
	total := sum([]Cents{100, 250})
	fmt.Printf("정의 타입: %d (%T)\n", total, total)
	fmt.Println("빈 입력:", sum([]int64(nil)))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/sum
```

예상 출력:

```text
int: 6
정의 타입: 350 (main.Cents)
빈 입력: 0
```


`|`는 타입 집합의 합집합을 나타냅니다. `~int`는 기반 타입(underlying type)이 int인 타입들을, `~int64`는 기반 타입이 int64인 타입들을 포함합니다. [General interfaces](https://go.dev/ref/spec#General_interfaces)를 참고합니다.

Cents는 `type Cents int64`로 정의한 새 타입입니다. 정확히 int64라는 타입 자체와 같지는 않지만 기반 타입이 int64이므로 Integer 제약을 만족합니다.

| 제약 | int64 자체 | type Cents int64 |
| --- | --- | --- |
| int64 | 허용 | 허용되지 않음 |
| ~int64 | 허용 | 허용 |
| ~int \| ~int64 | 허용 | 허용 |

### 4.2 ~는 값 변환이 아닙니다

`~int64`라고 썼다고 Cents를 int64로 변환해서 반환하는 것이 아닙니다. 타입 인자 T가 Cents이면 total도 Cents이고 결과도 Cents입니다. 출력의 타입 이름으로 이를 확인합니다.

기반 타입이 같은 다른 정의 타입들도 자동으로 서로 대입되는 것은 아닙니다. 제약의 타입 집합과 일반 대입 규칙은 다릅니다.

### 4.3 허용할 타입을 무한히 넓히지 않기

이 Integer 제약은 학습을 위해 int와 int64 계열만 포함합니다. uint, int32, float64는 이 함수의 입력 계약 밖입니다. 실제 프로젝트에서는 필요한 범위를 정해 확장할 수 있습니다.

정수 덧셈을 허용한다고 오버플로 검사가 자동으로 생기지 않습니다. sum은 선택된 정수 타입의 연산 규칙을 그대로 사용하며, 예제의 작은 입력은 범위를 넘지 않습니다. 금액 합계가 반드시 안전해야 하는 함수라면 범위·오류 반환 정책을 별도로 구현해야 합니다.

부동소수점까지 한꺼번에 포함하면 정밀도와 NaN 같은 문제도 함께 들어옵니다. 타입을 더 많이 허용하는 것과 계약이 더 좋아지는 것은 같은 일이 아닙니다.

### 4.4 제약용 인터페이스는 일반 값 타입과 다를 수 있습니다

Integer처럼 타입 항을 포함하는 비기본 인터페이스는 타입 제약으로 사용합니다. 아래 선언은 **의도적으로 실패하는 발췌**입니다.

```go
var value Integer
```

Integer는 int 또는 int64 값을 담는 런타임 상자 역할을 하지 않습니다. 이 제약으로 받을 타입을 정하는 제네릭 함수의 매개변수를 선언하는 데 사용합니다.

09장에서 배운 메서드 목록만 있는 기본 인터페이스는 일반 변수 타입으로 사용할 수 있습니다. interface라는 키워드가 같아도 값 타입으로 사용할 수 있는지와 제약에서 표현하는 타입 집합은 구별합니다.

## 5. 슬라이스 타입을 보존하는 함수

### 5.1 원소 타입과 슬라이스 자체 타입

`func copyValues[T any](values []T) []T`는 원소 타입 T를 유지하지만 반환 타입은 []T입니다. `type Scores []int`처럼 정의한 슬라이스 타입의 이름과 메서드까지 유지하려면 슬라이스 자체를 타입 매개변수로 표현할 수 있습니다.

저장 경로: [`examples/13-generics/filter/main.go`](../../examples/13-generics/filter/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

// filter preserves the slice type and nilness; keep must be non-nil.
func filter[S ~[]E, E any](values S, keep func(E) bool) S {
	if values == nil {
		return nil
	}
	result := make(S, 0, len(values))
	for _, value := range values {
		if keep(value) {
			result = append(result, value)
		}
	}
	return result
}

type Scores []int

func (scores Scores) Label() string { return fmt.Sprintf("점수 %d개", len(scores)) }
func main() {
	values := Scores{30, 80, 95}
	passed := filter(values, func(score int) bool { return score >= 60 })
	fmt.Println("통과:", passed, passed.Label())
	passed[0] = 100
	fmt.Println("원본:", values)
	var empty Scores
	fmt.Println("nil 유지:", filter(empty, func(int) bool { return true }) == nil)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/filter
```

예상 출력:

```text
통과: [80 95] 점수 2개
원본: [30 80 95]
nil 유지: true
```


시그니처를 두 층으로 읽습니다.

```go
func filter[S ~[]E, E any](values S, keep func(E) bool) S
```

| 타입 매개변수 | 의미 |
| --- | --- |
| S | 기반 타입이 []E인 슬라이스 타입 |
| E | 슬라이스의 원소 타입 |

values가 Scores이면 S는 Scores, E는 int로 추론됩니다. 반환 타입이 S이므로 결과에도 Scores의 Label 메서드가 있습니다. 원소 타입만 보존하는 함수였다면 이 메서드 호출을 그대로 할 수 없을 수 있습니다.

### 5.2 입력 순서와 결과 저장 공간

filter는 원소를 입력 순서대로 검사하고 통과한 값을 결과에 append합니다. 따라서 선택된 원소의 상대 순서는 유지됩니다. 새 슬라이스를 만들므로 결과의 int 원소를 바꿔도 원본 int 원소는 그대로입니다.

이것은 **얕은 복사**입니다. E가 포인터나 슬라이스를 포함한다면 원소 안의 참조 대상까지 재귀적으로 복제하지 않습니다. 05·08장의 공유 규칙은 제네릭 함수 안에서도 그대로 적용됩니다.

| 변화 | 원본 영향 |
| --- | --- |
| 결과 []int의 원소 대입 | 새 바깥 저장 공간이므로 원본 int 유지 |
| 결과 []*Node 원소가 가리키는 Node 필드 변경 | 같은 대상을 가리키면 원본 쪽에서도 보임 |
| 결과 구조체 안의 슬라이스 원소 변경 | 내부 저장 공간을 공유하면 보임 |

### 5.3 nil과 빈 슬라이스 계약

이 함수는 nil 입력이면 nil을 반환하고, nil이 아닌 빈 입력이면 nil이 아닌 빈 결과를 반환합니다. 조건에 맞는 값이 없어도 입력이 non-nil이면 non-nil 빈 결과입니다.

이 구분은 Go가 모든 필터 함수에 강제하는 규칙이 아닙니다. 해당 함수의 API 선택입니다. JSON 직렬화 등에서 nil과 빈 슬라이스가 다르게 나타날 수 있으므로 정책을 문서화하고 테스트합니다.

keep은 nil이 아닌 함수여야 합니다. 입력 원소마다 한 번 호출합니다. 이 함수가 임의의 콜백 패닉을 복구하거나 callback의 부작용까지 없애는 것은 아닙니다.

## 6. 입력과 결과 타입이 다른 변환

### 6.1 타입 매개변수 두 개 사용

저장 경로: [`examples/13-generics/transform/main.go`](../../examples/13-generics/transform/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"strconv"
)

// transform preserves element order; fn must be non-nil.
func transform[T, U any](values []T, fn func(T) U) []U {
	if values == nil {
		return nil
	}
	result := make([]U, len(values))
	for i, value := range values {
		result[i] = fn(value)
	}
	return result
}
func main() {
	labels := transform([]int{2, 4}, func(value int) string { return "#" + strconv.Itoa(value) })
	fmt.Println("변환:", labels)
	lengths := transform(labels, func(value string) int { return len(value) })
	fmt.Println("바이트 길이:", lengths)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/transform
```

예상 출력:

```text
변환: [#2 #4]
바이트 길이: [2 2]
```


T는 입력 원소 타입이고 U는 출력 원소 타입입니다. 첫 호출은 int에서 string으로, 둘째 호출은 string에서 int로 변환합니다. 두 호출은 하나의 transform 구현을 사용합니다.

`fn func(T) U`는 변환 규칙을 호출자가 제공한다는 뜻입니다. 타입 매개변수가 구체 계산 정책을 만들어 주는 것은 아닙니다. 함수 값과 타입 매개변수를 조합하여 알고리즘과 개별 변환을 나눴습니다.

### 6.2 길이와 순서 계약

결과는 입력과 같은 길이를 가지며 i번째 결과는 i번째 입력을 변환한 값입니다. callback은 입력마다 한 번 호출합니다. 결과의 저장 공간은 새로 만듭니다.

첫 호출의 `#2`, `#4`는 ASCII 문자여서 바이트 길이가 각각 2입니다. 한글이나 이모지가 있는 일반 문자열의 화면 글자 수를 len으로 측정할 수 있다는 의미는 아닙니다. 문자열의 단위 구별은 07장과 연결됩니다.

### 6.3 변환이 실패할 수 있다면

현재 callback은 U만 반환하므로 실패 정책이 없습니다. 실제 변환이 오류를 반환한다면 함수 시그니처와 결과 계약을 추가로 정해야 합니다.

예를 들어 다음은 가능한 **설계 형태를 보여 주는 시그니처 발췌**입니다.

```go
func transformChecked[T, U any](
    values []T,
    convert func(T) (U, error),
) ([]U, error)
```

첫 오류에서 중단할지, 부분 결과를 반환할지, 여러 오류를 모을지는 제네릭 문법이 결정하지 않습니다. 11장에서 배운 오류 계약이 그대로 필요합니다. 함수를 만들기 전에 그 정책을 문서로 정합니다.

## 7. 제네릭 구조체

### 7.1 스택의 원소 타입을 매개변수로

스택(stack)은 마지막에 넣은 값을 먼저 꺼내는 자료구조입니다. 구조체 안의 슬라이스 원소 타입을 매개변수로 만들면 같은 자료구조를 여러 값 타입에 사용할 수 있습니다.

저장 경로: [`examples/13-generics/stack/main.go`](../../examples/13-generics/stack/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(value T) { s.items = append(s.items, value) }
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	last := len(s.items) - 1
	value := s.items[last]
	var zero T
	s.items[last] = zero
	s.items = s.items[:last]
	return value, true
}
func (s *Stack[T]) Len() int { return len(s.items) }

type Popper[T any] interface{ Pop() (T, bool) }

var _ Popper[string] = (*Stack[string])(nil)

func main() {
	var stack Stack[string]
	stack.Push("첫째")
	stack.Push("둘째")
	for stack.Len() > 0 {
		value, ok := stack.Pop()
		fmt.Println("꺼내기:", value, ok)
	}
	value, ok := stack.Pop()
	fmt.Printf("빈 스택: %q %t\n", value, ok)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/stack
```

예상 출력:

```text
꺼내기: 둘째 true
꺼내기: 첫째 true
빈 스택: "" false
```


`Stack[int]`, `Stack[string]`은 각각 구체 타입입니다. 서로 다른 원소 타입의 스택은 같은 타입이 아니며, Stack[int]에 문자열을 Push할 수 없습니다.

| 코드 | 의미 |
| --- | --- |
| `type Stack[T any] struct` | T를 받는 제네릭 타입 선언 |
| `items []T` | 선택된 원소 타입의 저장 공간 |
| `var stack Stack[string]` | string 스택의 영값 |
| `func (s *Stack[T]) Push(value T)` | 수신자 타입의 T를 사용 |

타입을 선언할 때 받은 T를 메서드 수신자에서 다시 이름으로 적어 메서드 본문에서 사용할 수 있게 합니다. 여기의 Push는 메서드 자체에 새로운 타입 매개변수를 추가한 것이 아닙니다. 그 차이는 10절에서 봅니다.

### 7.2 영값과 수신자

Stack의 영값에는 nil 슬라이스가 있고 append가 이를 확장할 수 있으므로 별도 생성자 없이 사용합니다. 비어 있을 때 Pop은 T의 영값과 false를 반환합니다.

메서드는 nil이 아닌 *Stack 수신자를 전제로 합니다. `var stack Stack[int]`는 유효한 영값 구조체이지만, `var stack *Stack[int]`는 nil 포인터입니다. “영값 사용 가능”이라는 말이 모든 nil 포인터 메서드 호출까지 안전하다는 뜻은 아닙니다.

### 7.3 꺼낸 슬롯을 영값으로 지우는 이유

길이를 줄이기 전 마지막 슬롯을 T의 영값으로 지웁니다. T가 포인터라면 더 이상 스택이 보관하지 않는 값을 바깥 저장 공간이 불필요하게 계속 참조하지 않도록 하는 목적입니다.

슬라이스의 길이만 줄여도 기존 배열의 뒤쪽 값은 남을 수 있습니다. 단순히 길이 0으로 보인다는 사실과 내부 저장 공간의 참조가 사라졌다는 사실은 다릅니다. 이 동작은 포인터 원소 테스트로 확인합니다.

이 정리가 스택 전체의 메모리를 즉시 운영체제에 돌려준다는 뜻은 아닙니다. 할당과 회수의 구체 시점은 런타임 동작이며 성능 장에서 측정합니다.

### 7.4 상태를 가진 제네릭 값을 복사할 때

Stack을 값으로 복사하면 안의 슬라이스 헤더도 복사됩니다. 기존 배열을 공유할 수 있으므로 두 독립 스택을 얻었다고 가정하면 안 됩니다. 일반 구조체의 얕은 복사와 같은 문제입니다.

참고 예제는 하나의 스택 변수에 포인터 수신자로 접근합니다. 스냅샷이나 복제 기능이 필요하다면 저장 공간을 따로 복사하는 계약을 구현합니다. 타입 매개변수는 소유권이나 동시 접근을 자동으로 해결하지 않습니다.

## 8. 제네릭 인터페이스

### 8.1 인터페이스 선언 자체의 타입 매개변수

스택 예제에는 다음 인터페이스가 있습니다.

```go
type Popper[T any] interface {
    Pop() (T, bool)
}
```

Popper[int]는 Pop() (int, bool)을, Popper[string]은 Pop() (string, bool)을 요구합니다. 인터페이스를 구체 타입으로 사용하기 전에 타입 인자를 지정합니다.

이것은 인터페이스 메서드 Pop이 자기 타입 매개변수를 받는 것이 아닙니다. 인터페이스 선언의 T를 메서드 시그니처에서 사용하는 것입니다. 현재 Go에서도 허용되는 형태입니다.

### 8.2 구현 여부를 컴파일 때 확인

```go
var _ Popper[string] = (*Stack[string])(nil)
```

이 선언은 *Stack[string]이 Popper[string]을 구현하는지 컴파일 시점에 확인합니다. nil 포인터를 실제로 Pop 호출하는 코드가 아닙니다. 이름을 `_`로 두어 런타임에 사용할 변수를 만들 필요가 없음을 표시합니다.

메서드의 인자·반환 타입이 달라지면 이 선언이 컴파일 오류로 알려 줄 수 있습니다. 모든 타입에 이런 선언을 반드시 붙일 필요는 없지만 중요한 계약을 명시하는 데 사용할 수 있습니다.

### 8.3 일반 인터페이스와 제네릭의 선택

서로 다른 구현이 같은 동작을 제공하면 일반 인터페이스로 충분할 때가 많습니다. 같은 알고리즘이 입력 타입과 결과 타입의 관계를 유지해야 하면 타입 매개변수가 도움됩니다.

| 요구 | 가능한 선택 |
| --- | --- |
| 다양한 저장소가 같은 Save 메서드 제공 | 작은 일반 인터페이스 |
| int·string 등 같은 타입의 슬라이스에서 첫 값 | 제네릭 함수 |
| 값의 구체 타입을 유지한 스택 | 제네릭 타입 |
| 이질적인 여러 값을 한 슬라이스에 담기 | any 또는 공통 인터페이스와 명시적 처리 |

어느 쪽이 항상 더 현대적이거나 더 빠른 것은 아닙니다. 필요한 계약을 가장 분명하게 드러내는 형태를 선택합니다.

## 9. 순서 제약과 부동소수점 정책

### 9.1 cmp.Ordered 사용

정수·부동소수점·문자열처럼 순서 비교를 제공하는 타입에는 표준 cmp.Ordered 제약을 사용할 수 있습니다. [cmp.Ordered](https://pkg.go.dev/cmp#Ordered)를 참고합니다.

저장 경로: [`examples/13-generics/ordered/main.go`](../../examples/13-generics/ordered/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"cmp"
	"fmt"
	"math"
)

// smallest uses cmp.Compare's total ordering, including its NaN policy.
func smallest[T cmp.Ordered](values []T) (T, bool) {
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	result := values[0]
	for _, value := range values[1:] {
		if cmp.Compare(value, result) < 0 {
			result = value
		}
	}
	return result, true
}
func main() {
	n, ok := smallest([]int{8, 3, 5})
	fmt.Println("정수:", n, ok)
	word, ok := smallest([]string{"banana", "apple"})
	fmt.Println("문자열:", word, ok)
	number, _ := smallest([]float64{2, math.NaN(), 1})
	fmt.Println("NaN 정책:", math.IsNaN(number))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/ordered
```

예상 출력:

```text
정수: 3 true
문자열: apple true
NaN 정책: true
```


이 함수는 cmp.Compare를 사용합니다. 정수와 문자열에서는 비교 결과가 -1·0·1로 나타납니다. NaN은 일반 `<` 비교와 다른 명시적 정책을 가집니다. cmp.Compare는 NaN을 NaN이 아닌 값보다 앞에 두고 NaN끼리는 같은 순서로 취급합니다. [cmp.Compare](https://pkg.go.dev/cmp#Compare)를 참고합니다.

따라서 이 예제의 smallest는 NaN이 하나라도 있으면 NaN을 결과로 선택합니다. 이것은 해당 비교 규칙을 선택한 함수의 계약입니다.

### 9.2 < 하나로 모든 데이터 의미가 정해지지는 않습니다

float의 NaN은 일반 비교에서 특수한 성질을 갖습니다. 문자열의 기본 순서도 사람이 쓰는 모든 언어의 사전식 정렬이나 대소문자 무시 정렬과 같지는 않습니다.

타입 제약이 연산의 존재를 보장해도 업무상 원하는 순서까지 결정하지 않습니다. 이름 정렬·가격 정렬·결측값 처리에는 별도의 정책이 필요할 수 있습니다.

### 9.3 빈 입력을 어떻게 처리하나요?

smallest는 빈 입력에서 영값과 false를 반환합니다. 최소값이 0일 수 있으므로 숫자 하나로 빈 입력을 구별하지 않습니다.

표준 라이브러리의 비슷한 함수가 다른 빈 입력 계약을 가질 수 있습니다. 함수를 바꿀 때 이름이 비슷하다는 이유로 결과와 실패 정책까지 같다고 가정하지 마세요.

## 10. Go 1.27의 제네릭 메서드

### 10.1 수신자의 T와 메서드의 U

Go 1.27에서는 메서드가 자체 타입 매개변수를 선언할 수 있습니다. 아래 Box는 T 타입의 값을 보관하고, Map 메서드는 호출마다 U 타입의 결과를 만드는 변환을 받습니다.

저장 경로: [`examples/13-generics/generic-methods/main.go`](../../examples/13-generics/generic-methods/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"strconv"
)

type Box[T any] struct{ Value T }

// Map has its own U parameter in addition to the receiver's T (Go 1.27+).
func (box Box[T]) Map[U any](convert func(T) U) Box[U] {
	return Box[U]{Value: convert(box.Value)}
}
func main() {
	number := Box[int]{Value: 7}
	text := number.Map(func(value int) string { return "#" + strconv.Itoa(value) })
	fmt.Printf("원본: %d (%T)\n", number.Value, number)
	fmt.Printf("변환: %s (%T)\n", text.Value, text)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/generic-methods
```

예상 출력:

```text
원본: 7 (main.Box[int])
변환: #7 (main.Box[string])
```


선언을 나누어 읽습니다.

```go
func (box Box[T]) Map[U any](convert func(T) U) Box[U]
```

| 부분 | 언제 정해지는 타입 |
| --- | --- |
| Box[T]의 T | 수신자 Box 값을 만들 때 |
| Map의 U | 해당 메서드를 사용할 때 |
| convert의 입력 | 기존 T |
| convert의 결과 | 새 U |
| 반환 Box[U] | 새 원소 타입을 가진 상자 |

앞의 Stack.Push는 수신자 타입에서 이미 정해진 T를 사용했습니다. Box.Map은 추가로 U를 선언하므로 **메서드 자체가 제네릭**입니다. 메서드가 있는 제네릭 타입과 제네릭 메서드는 같은 말이 아닙니다.

### 10.2 명시적 인스턴스화와 추론

예제의 callback이 int를 받아 string을 반환하므로 U는 string으로 추론됩니다. `number.Map[string](...)`처럼 명시해도 됩니다. 제네릭 함수와 마찬가지로 구체 타입이 적용되어야 호출하거나 함수 값으로 사용할 수 있습니다.

이 프로그램을 Go 1.26 이하의 도구로 컴파일하면 메서드의 자체 타입 매개변수 구문을 지원하지 않을 수 있습니다. 설치 버전만 추정하지 말고 `go version`, `go env GOMOD GOTOOLCHAIN`으로 실제 선택된 도구 체인을 확인합니다.

### 10.3 인터페이스 메서드의 제한은 남아 있습니다

다음은 **의도적으로 컴파일되지 않는 선언**입니다.

```go
type Mapper interface {
    Map[U any](func(int) U) U
}
```

현재 Go 1.27에서도 인터페이스 메서드에는 자체 타입 매개변수를 붙일 수 없습니다. 또한 제네릭 메서드 자체로 인터페이스의 메서드를 구현하는 방식은 지원되지 않습니다. [Method declarations](https://go.dev/ref/spec#Method_declarations), [Basic interfaces](https://go.dev/ref/spec#Basic_interfaces)를 참고합니다.

필요한 관계를 인터페이스 선언의 타입 매개변수로 옮길 수 있는지, 구체 타입을 가진 일반 메서드로 표현할지, 일반 제네릭 함수로 둘지 실제 소비 형태에 맞춰 판단합니다.

### 10.4 기존 버전도 지원해야 한다면

같은 변환을 일반 제네릭 함수로 표현할 수도 있습니다. 다음은 별도의 함수 선언 발췌입니다.

```go
func mapBox[T, U any](box Box[T], convert func(T) U) Box[U] {
    return Box[U]{Value: convert(box.Value)}
}
```

알고리즘이 달라지는 것이 아니라 API를 어디에 두고 어떻게 호출하는지가 달라집니다. 이 저장소는 Go 1.27이 기준이므로 실제 실행 예제에서는 새 메서드 구문을 검증합니다.

신규 문법이 있다고 모든 함수를 메서드로 옮길 필요는 없습니다. 수신자와 동작의 관계가 자연스럽고 호출자가 그 위치에서 기능을 찾기 쉬운지 판단합니다.

## 11. 표준 라이브러리의 제네릭 함수 사용

### 11.1 먼저 있는 함수를 살펴보기

검색·정렬·복제처럼 흔한 작업은 slices·maps·cmp 패키지에 이미 제공될 수 있습니다. 직접 알고리즘을 만드는 학습과 실제 코드에서 검증된 공통 함수를 사용하는 선택을 구별합니다.

저장 경로: [`examples/13-generics/stdlib/main.go`](../../examples/13-generics/stdlib/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Student struct {
	Name  string
	Score int
}

func main() {
	original := []Student{{Name: "민수", Score: 80}, {Name: "영희", Score: 95}, {Name: "재성", Score: 80}}
	ranked := slices.Clone(original)
	slices.SortFunc(ranked, func(a, b Student) int {
		if order := cmp.Compare(b.Score, a.Score); order != 0 {
			return order
		}
		return cmp.Compare(a.Name, b.Name)
	})
	for _, student := range ranked {
		fmt.Println(student.Name, student.Score)
	}
	fmt.Println("원본 첫째:", original[0].Name)
	fmt.Println("포함:", slices.Contains([]string{"Go", "SQL"}, "Go"))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/13-generics/stdlib
```

예상 출력:

```text
영희 95
민수 80
재성 80
원본 첫째: 민수
포함: true
```


이 프로그램은 원본 순서를 유지하려고 먼저 slices.Clone으로 바깥 슬라이스를 복사합니다. 그 뒤 복사본을 점수 내림차순, 같은 점수에서는 이름 오름차순으로 정렬합니다.

| 호출 | 역할 |
| --- | --- |
| slices.Clone | 원소를 복사한 슬라이스 생성 |
| slices.SortFunc | 전달한 비교 규칙으로 제자리 정렬 |
| cmp.Compare | 정수 또는 문자열의 비교 결과 반환 |
| slices.Contains | 비교 가능한 원소의 포함 여부 |

정렬은 ranked를 변경합니다. Clone 없이 original을 그대로 정렬하면 원본 순서도 바뀝니다. 이것은 제네릭 함수라서 달라지는 특별 규칙이 아니라 슬라이스 공유 규칙입니다.

### 11.2 비교 함수의 의미

SortFunc 비교 함수는 a가 앞이면 음수, 같은 순서면 0, 뒤면 양수를 반환합니다. 점수를 내림차순으로 정렬하려고 cmp.Compare의 인자 순서를 뒤집었습니다.

`return b.Score-a.Score`처럼 뺄셈으로 비교를 구현하면 큰 정수에서 오버플로가 생겨 순서가 깨질 수 있습니다. cmp.Compare는 순서 판단 의도를 분명히 나타냅니다.

동점에 대해 이름까지 비교하므로 이 예제의 최종 순서는 결정적입니다. 그래도 모든 필드가 같은 요소들끼리 원래 상대 순서를 보존해야 한다면 안정 정렬의 계약을 가진 함수를 따로 확인합니다. [slices.SortFunc](https://pkg.go.dev/slices#SortFunc), [slices.SortStableFunc](https://pkg.go.dev/slices#SortStableFunc)를 참고합니다.

### 11.3 Clone은 깊은 복제가 아닙니다

Student는 문자열과 int 필드만 있으므로 이 예제에서 원본 필드 값과 정렬 복사본의 관계가 단순합니다. 구조체 안에 슬라이스나 포인터가 있으면 Clone 뒤에도 그 내부 대상을 공유할 수 있습니다.

함수 이름이 Clone이어도 어떤 층까지 복사하는지 문서를 확인합니다. 표준 slices.Clone은 원소 대입을 통한 얕은 복사입니다. [slices.Clone](https://pkg.go.dev/slices#Clone)을 참고합니다.

### 11.4 표준 함수의 시그니처 읽기

표준 문서의 타입 매개변수가 길어 보이면 입력·출력 관계부터 찾습니다. 예를 들어 S ~[]E가 있으면 원소 타입 E와 슬라이스 타입 S를 구분한 것으로 읽을 수 있습니다.

모든 제약을 암기하기보다 “함수가 무슨 연산을 필요로 하는가”, “입력의 이름 붙인 타입을 보존하는가”, “값을 복사하는가 수정하는가”를 물어보세요. 이 질문들이 사용자 코드의 계약과 직접 연결됩니다.

## 12. 테스트로 확인할 계약

### 12.1 타입 수보다 의미 있는 경우

같은 코드를 int·int64·string에 한 번씩 실행했다고 모든 경계를 검증한 것은 아닙니다. 빈 입력, 원소 한 개, 중복, 없는 값, 정의 타입, 공유 상태, 비교 정책 등 알고리즘의 위험을 확인합니다.

| 예제 | 주요 검증 |
| --- | --- |
| first | 빈 입력과 정상 영값 구별 |
| index | 첫 일치 위치와 -1 |
| sum | 빈 입력, 음수, 정의된 정수 타입 유지 |
| filter | 순서, nil 구분, callback 횟수, 바깥 저장 공간 독립 |
| transform | 길이·순서·서로 다른 입력/출력 타입 |
| stack | 마지막 입력 먼저 출력, 빈 상태, 꺼낸 포인터 슬롯 제거 |
| ordered | 빈 입력, NaN 위치에 무관한 명시적 정책 |
| generic-methods | 결과 Box 타입, 변환 횟수, 원본 값 유지 |

### 12.2 컴파일 오류도 학습 자료입니다

any로 비교하려는 함수, 제약 인터페이스를 값 변수로 쓰는 선언, 인터페이스 메서드의 자체 타입 매개변수는 컴파일 단계에서 실패합니다. 실패 이유를 읽는 것이 타입 제약을 이해하는 중요한 연습입니다.

이런 파일은 정상 `examples/` 아래에 두지 않습니다. 개인 실습 파일에서 의도적으로 실행한 뒤 정상 상태로 고칩니다. 교재 전체 검사는 항상 성공 가능한 코드만 포함합니다.

### 12.3 제네릭이라고 성능이 자동으로 좋아지지는 않습니다

제네릭은 타입 안전한 재사용의 도구입니다. 성능은 알고리즘, 할당, 데이터 크기, 컴파일러 구현과 호출 형태 등에 영향을 받습니다. “제네릭은 항상 복사보다 빠르다” 또는 “인터페이스보다 항상 빠르다” 같은 단정을 하지 않습니다.

성능 때문에 구조를 바꾸려면 26장에서 배울 벤치마크와 프로파일링으로 같은 작업을 같은 조건에서 비교합니다. 이 장의 예제에는 측정하지 않은 속도 수치를 붙이지 않습니다.

## 13. 언제 제네릭을 사용하지 않을까요?

### 13.1 업무 규칙이 서로 다른 경우

정수 금액과 문자열 이름을 모두 처리한다는 이유로 하나의 `Process[T any]` 함수에 타입 스위치를 몰아넣으면 타입별 차이를 숨길 수 있습니다. 각 타입에서 전혀 다른 정책이 필요하면 별도의 함수가 더 명확합니다.

함수 이름과 타입 매개변수를 읽고 호출자가 결과를 예상할 수 있어야 합니다. 실제 작업을 설명하기 위해 매번 T의 동적 타입에 따라 긴 예외 목록이 필요하다면 일반화 범위를 다시 검토합니다.

### 13.2 작은 일반 함수로 충분한 경우

한 종류의 값만 다루며 앞으로 다른 타입을 사용할 근거도 없다면 일반 함수로 시작할 수 있습니다. 같은 알고리즘이 실제로 반복될 때 공통 계약을 찾아도 늦지 않습니다.

제네릭을 도입하면 유지 관리해야 할 제약·타입 추론·사용 가능 타입의 범위가 생깁니다. 코드 줄 수만 줄어드는 것이 전체 복잡도가 줄어드는 것과 같지는 않습니다.

### 13.3 인터페이스가 실제 교체 지점인 경우

HTTP 핸들러가 저장소의 조회·저장 동작만 필요하다면 저장되는 데이터 타입은 이미 구체적으로 정해졌을 수 있습니다. 그때는 작은 인터페이스가 구현 교체를 표현하는 데 충분할 수 있습니다.

뒤의 할 일 API에서는 메모리 저장소와 SQL 저장소를 같은 동작으로 연결합니다. 제네릭 컨테이너를 만들었으니 모든 서비스도 제네릭으로 바꾸는 식으로 확장하지 않습니다.

## 14. 자주 하는 실수

| 증상 | 원인 후보 | 수정 방향 |
| --- | --- | --- |
| any 함수에서 + 또는 ==가 실패 | 제약이 필요한 연산을 보장하지 않음 | 사용 연산에 맞는 제약 |
| Cents가 int64 제약에 맞지 않음 | 정확한 타입과 기반 타입 혼동 | 의도에 맞으면 ~int64 |
| first(nil)의 타입 추론 실패 | 원소 타입 정보 없음 | 명시적 타입 인자 또는 typed nil |
| 필터 뒤 Label 메서드를 호출 못함 | 반환 타입을 []E로 고정 | S ~[]E와 반환 S 검토 |
| comparable인데 런타임 비교 panic | 인터페이스 안의 비교 불가능 동적 값 | 입력 계약과 구체 타입 범위 확인 |
| 제네릭 스택 복사 후 서로 영향 | 내부 슬라이스 저장 공간 공유 | 소유권·복사 계약 설계 |
| 꺼낸 포인터가 배열에 남음 | 길이만 줄이고 슬롯을 지우지 않음 | 제거 슬롯을 T의 영값으로 설정 |
| 메서드의 [U]가 컴파일 안 됨 | Go 1.27보다 낮은 실제 도구 체인 | 버전·go.mod·GOTOOLCHAIN 확인 |
| 인터페이스 Map[U]가 컴파일 안 됨 | 인터페이스 메서드 자체 타입 매개변수 제한 | 인터페이스 타입 매개변수 또는 API 구조 변경 |
| 같은 입력인데 정렬 결과가 불안정 | 비교 계약·동점 정책·NaN 정책 미정 | 비교 규칙과 안정성 요구 명시 |

## 15. 연습문제

[연습문제](../../exercises/13-generics/README.md)에서 직접 제약과 반환 계약을 정합니다. [해설](../../solutions/13-generics/README.md)은 구현 뒤 비교하세요.

| 번호 | 수준 | 목표 |
| --- | --- | --- |
| 1 | 확인 | 타입 추론과 제약 충족 판단 |
| 2 | 변형 | 마지막 값을 타입 안전하게 반환 |
| 3 | 변형 | 정의 타입을 유지하며 중복 제거 |
| 4 | 응용 | 제한된 순서 키의 맵 키 정렬 |
| 5 | 응용 | 키 함수로 여러 타입의 값 그룹화 |
| 6 | 응용 | 조건에 따라 두 슬라이스로 분리 |
| 선택 A | 응용 | Box 제네릭 메서드 호출과 함수형 대안 비교 |

```sh
go test ./examples/13-generics/...
go test ./solutions/13-generics/...
```

## 16. 요약과 자기 점검

- 타입 매개변수는 호출마다 정해지는 타입 관계를 이름으로 표현합니다.
- 제약은 허용 타입과 사용할 수 있는 연산을 결정합니다.
- `~`는 기반 타입을 기준으로 집합을 넓히며 값 변환을 수행하지 않습니다.
- 제네릭 함수도 빈 입력, 오류, 소유권, 공유 상태의 계약이 필요합니다.
- 제네릭 타입의 수신자 매개변수와 Go 1.27 메서드 자체의 매개변수는 구별합니다.
- 표준 slices·maps·cmp를 먼저 살펴보고, 일반 함수·인터페이스·제네릭 중 필요한 관계를 가장 잘 표현하는 것을 선택합니다.

책을 보지 않고 답해 보세요.

1. `first([]int{0})`와 `first[int](nil)`의 결과를 어떻게 구별하나요?
2. T any에서 a==b가 일반적으로 허용되지 않는 이유는 무엇인가요?
3. int64와 ~int64 제약이 Cents에 대해 어떻게 다른가요?
4. 반환 타입 []E와 S ~[]E의 반환 S는 어떤 차이가 있나요?
5. 새 슬라이스를 만들면 내부 포인터의 대상도 자동 복제되나요?
6. Stack[T].Push와 Box[T].Map[U]는 타입 매개변수가 어느 위치에서 다른가요?
7. Go 1.27에서도 인터페이스 메서드 Map[U]를 선언할 수 있나요?
8. cmp.Compare의 NaN 정책을 사용한 smallest는 어떤 결과를 선택하나요?
9. generic이라는 이유만으로 성능이 더 좋다고 할 수 있나요?
10. 이미 표준 slices 함수로 계약을 만족할 때 직접 새 함수를 만드는 이유가 있나요?

## 참고 자료

공식 문서 및 Go 1.27.3 컴파일 동작 확인일: **2026-11-09**.

- [Go 제네릭 입문](https://go.dev/doc/tutorial/generics), [When To Use Generics](https://go.dev/blog/when-generics)
- [Go 명세: Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations), [Instantiations](https://go.dev/ref/spec#Instantiations), [Type inference](https://go.dev/ref/spec#Type_inference)
- [Go 명세: General interfaces](https://go.dev/ref/spec#General_interfaces), [Satisfying a type constraint](https://go.dev/ref/spec#Satisfying_a_type_constraint)
- [Go 명세: Method declarations](https://go.dev/ref/spec#Method_declarations), [Basic interfaces](https://go.dev/ref/spec#Basic_interfaces)
- [Go 1.27 언어 변경](https://go.dev/doc/go1.27#language), [Go 1.18](https://go.dev/doc/go1.18), [Go 1.21](https://go.dev/doc/go1.21)
- [cmp](https://pkg.go.dev/cmp), [slices](https://pkg.go.dev/slices), [maps](https://pkg.go.dev/maps)

## 다음 학습

다음 장에서는 파일과 스트림을 읽고 씁니다. 이미 배운 인터페이스·오류·자원 수명을 실제 입출력 경계에 적용합니다.

[이전: 12장 자동 테스트](../12-testing/README.md) · [다음: 14장 파일과 스트림](../14-files-io/README.md) · [전체 목차](../README.md)
