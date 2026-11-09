# 08. 포인터와 구조체

04장에서는 함수에 인자를 전달하면 매개변수가 그 값을 받는다고 배웠습니다. 05장과 06장에서는 슬라이스와 맵의 값을 복사해도 일부 데이터가 공유될 수 있음을 확인했습니다. 이제 어떤 변수를 바꾸고 있는지 직접 추적하고, 관련 데이터를 하나의 타입으로 묶습니다.

이 장을 마치면 학생의 이름과 점수를 하나의 값으로 표현하고, 원본을 바꾸는 함수와 새 값을 반환하는 함수를 구분할 수 있습니다. 포인터를 많이 쓰는 것이 목표가 아닙니다. **변경할 대상과 공유 범위를 설명할 수 있는 코드**를 만드는 것이 목표입니다.

## 학습 목표

- `T`, `*T`, `&value`, `*pointer`를 구분하고 포인터를 통해 원래 변수를 읽고 수정합니다.
- `nil` 포인터의 의미를 설명하고, 사용 전에 필요한 검사를 합니다.
- 구조체 타입과 필드를 선언하고 리터럴·제로 값·복사 동작을 사용합니다.
- 슬라이스와 맵에 들어 있는 구조체를 올바른 방법으로 갱신합니다.
- 구조체에 슬라이스나 맵이 들어 있을 때 독립 복사가 필요한 부분을 찾아 구현합니다.

## 사전 지식과 준비

[04장 함수와 스코프](../04-functions-scope/README.md)의 인자·반환값, [05장 배열과 슬라이스](../05-arrays-slices/README.md)의 값 복사와 기반 배열, [06장 맵](../06-maps/README.md)의 조회·갱신을 사용합니다. 이름을 검사하는 연습문제에서는 [07장 문자열과 유니코드](../07-strings-unicode/README.md)의 `strings.TrimSpace`와 `utf8.ValidString`도 사용합니다.

모든 명령은 저장소 루트에서 실행합니다. 새 `go.mod`를 만들지 않습니다. 기본 예제는 표준 라이브러리만 사용하며 외부 파일이나 서비스에 접근하지 않습니다.

본문은 Go 1.27 계열을 기준으로 설명합니다. 예제와 경계 입력은 Go 1.27.3에서 확인했으며, 공식 문서 확인일은 **2026-11-09**입니다. 설치는 00장, 테스트 문법 자체는 12장에서 다룹니다.

| 학습 경로 | 권장 순서 |
| --- | --- |
| 처음 읽는 독자 | 포인터 → 값 전달 → 구조체 → 컬렉션 안의 구조체 → 개인 실습 |
| 다른 언어 경험이 있는 독자 | 2·3절의 Go 포인터 규칙 → 7·8·9절의 복사·갱신 → 연습문제 |
| 지금 이해하지 않아도 되는 부분 | 선택 읽기의 메모리 배치와 `new(표현식)`, 다음 장의 메서드 |

## 1. 값 자체와 그 값이 저장된 변수

`number := 10`은 `number`라는 변수에 정수 값 `10`을 저장합니다. `copied := number`는 그 값을 다른 변수에 복사합니다. 이후 `copied`에 대입해도 `number`의 값은 바뀌지 않습니다.

**포인터(pointer)**는 다른 변수를 가리키는 값입니다. `&number`는 `number`를 가리키는 포인터를 만들고, 그 타입은 `*int`입니다. 포인터를 통해 접근할 원래 값의 타입이 `int`라는 뜻입니다.

포인터를 숫자 주소로 외울 필요는 없습니다. 우선 “이 값으로 어느 변수에 접근할 수 있는가?”를 물어보세요. 주소의 실제 숫자는 실행 환경에 따라 달라질 수 있어 이 장의 예상 출력에는 넣지 않습니다.

### 1.1 `&`와 `*`를 읽는 순서

| 코드 | 문맥 | 의미 |
| --- | --- | --- |
| `var pointer *int` | 타입 | `int` 변수를 가리키는 포인터 타입 |
| `pointer := &number` | 표현식 | `number`의 주소를 얻어 포인터 값으로 저장 |
| `fmt.Println(*pointer)` | 표현식 | 가리키는 변수의 값을 읽음 |
| `*pointer = 20` | 대입 왼쪽 | 가리키는 변수에 `20`을 저장 |
| `pointer = &other` | 대입 왼쪽 | 포인터 변수 자체가 다른 변수를 가리키게 바뀜 |

`*pointer`를 통해 대상 변수에 접근하는 동작을 **역참조(indirection 또는 dereference)**라고 부릅니다. 타입의 `*int`와 표현식의 `*pointer`는 위치에 따라 역할이 다릅니다. 곱셈의 `left * right`와도 구분합니다.

포인터 타입과 주소 연산의 정확한 규칙은 [언어 명세: Pointer types](https://go.dev/ref/spec#Pointer_types), [Address operators](https://go.dev/ref/spec#Address_operators)를 참고합니다.

### 1.2 실행 예제: 두 이름으로 같은 변수에 접근하기

저장 경로: [`examples/08-pointers-structs/pointers/main.go`](../../examples/08-pointers-structs/pointers/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	value := 10
	pointer := &value
	fmt.Printf("타입: %T, 읽은 값: %d\n", pointer, *pointer)
	*pointer += 5
	fmt.Println("원래 변수:", value)
	alias := pointer
	*alias = 20
	other := 20
	fmt.Println("같은 변수:", pointer == &value)
	fmt.Println("같은 숫자, 다른 변수:", pointer == &other)
	fmt.Println("두 경로:", value, *alias)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/08-pointers-structs/pointers
```

예상 출력:

```text
타입: *int, 읽은 값: 10
원래 변수: 15
같은 변수: true
같은 숫자, 다른 변수: false
두 경로: 20 20
```

`pointer`와 `alias`는 서로 다른 포인터 변수입니다. `alias := pointer`가 포인터 **값을 복사**했기 때문에 두 포인터가 같은 `value`를 가리킵니다. `*alias = 20`은 `alias`에 숫자 `20`을 넣는 문장이 아닙니다. 그 포인터가 가리키는 `value`에 넣는 문장입니다.

| 실행한 문장 | `value` | `*pointer` | `*alias` |
| --- | ---: | ---: | ---: |
| `pointer := &value` | 10 | 10 | 아직 선언 전 |
| `*pointer += 5` | 15 | 15 | 아직 선언 전 |
| `alias := pointer` | 15 | 15 | 15 |
| `*alias = 20` | 20 | 20 | 20 |

`other`에도 `20`이 들어 있지만 `pointer == &other`는 `false`입니다. 이 비교는 두 정수 값이 같은지 확인하지 않습니다. 포인터가 같은 변수를 가리키는지 확인합니다.

이 예제의 `int` 변수처럼 크기가 0이 아닌 별도 변수는 구분됩니다. 크기가 0인 타입에 대한 특수한 주소 동일성까지 이 결과로 일반화하지 않습니다. [포인터 비교 규칙](https://go.dev/ref/spec#Comparison_operators)을 참고하세요.

코드의 `%T`는 값의 타입을 출력합니다. `%p`로 실제 주소를 출력해 볼 수는 있지만, 주소 숫자를 정답으로 비교하지 않습니다. 출력 형식은 [fmt 문서](https://pkg.go.dev/fmt#hdr-Printing)에서 확인할 수 있습니다.


### 1.3 포인터는 대상 변수와 별도로 존재합니다

다음은 `main` 안에 넣는 독립적인 코드 조각입니다. `fmt`를 import한 상태에서 실행합니다.

```go
left, right := 10, 30
pointer := &left
pointer = &right
*pointer = 40
fmt.Println(left, right)
```

예상 출력은 `10 40`입니다. `pointer = &right`가 대상을 바꾼 뒤에는 `*pointer = 40`이 `right`를 수정합니다. `left`의 값은 그대로입니다.

한 문장을 볼 때 다음 두 질문을 나누면 흐름을 추적하기 쉽습니다.

1. 왼쪽이 `pointer`인가, `*pointer`인가?
2. 지금 `pointer`가 어느 변수를 가리키는가?

포인터의 포인터인 `**int`도 만들 수 있지만, 이 장의 문제에는 필요하지 않습니다. 호출자가 들고 있는 포인터를 교체해야 한다면 새 포인터를 반환하는 설계를 먼저 고려할 수 있습니다.

## 2. 포인터를 전달해도 인자 전달 규칙은 같습니다

함수에 `number`를 전달하면 정수 값이 복사됩니다. `&number`를 전달하면 포인터 값이 복사됩니다. Go가 갑자기 다른 인자 전달 방식을 사용하는 것은 아닙니다.

복사된 포인터를 역참조하면 호출자의 변수에 접근할 수 있다는 점이 다릅니다. 따라서 **포인터 매개변수 자체의 재대입**과 **포인터가 가리키는 변수의 수정**을 구분해야 합니다.

이 규칙은 [함수 호출 시 매개변수 초기화](https://go.dev/ref/spec#Calls)와 [값의 표현](https://go.dev/ref/spec#Representation_of_values)에 설명되어 있습니다.

### 2.1 실행 예제: 복사·수정·재대입 비교

저장 경로: [`examples/08-pointers-structs/pointer-functions/main.go`](../../examples/08-pointers-structs/pointer-functions/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func added(value, amount int) int {
	value += amount
	return value
}

func addInPlace(value *int, amount int) bool {
	if value == nil {
		return false
	}
	*value += amount
	return true
}

func redirect(value *int) {
	replacement := 99
	value = &replacement
	*value = 100
}

func main() {
	number := 10
	result := added(number, 3)
	fmt.Println("값으로 전달:", number, result)
	ok := addInPlace(&number, 5)
	fmt.Println("주소로 접근:", number, ok)
	redirect(&number)
	fmt.Println("포인터 재대입 뒤:", number)
	fmt.Println("nil 전달:", addInPlace(nil, 5))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/08-pointers-structs/pointer-functions
```

예상 출력:

```text
값으로 전달: 10 13
주소로 접근: 15 true
포인터 재대입 뒤: 15
nil 전달: false
```

`added`는 입력 정수를 복사한 매개변수 `value`를 바꾼 뒤 결과를 반환합니다. 호출자는 반환값을 `result`에 받았습니다. `number`에는 대입하지 않았으므로 여전히 `10`입니다.

`addInPlace`의 매개변수 타입은 `*int`입니다. `&number`를 받으면 `*value`가 `number`에 접근합니다. `*value += amount`가 호출자의 정수를 바꿉니다.

`redirect`는 매개변수 `value`에 새 포인터를 대입합니다. 이 매개변수는 호출자가 전달한 포인터 값의 **지역 복사본**이므로, 호출자의 `number`와 그 주소는 바뀌지 않습니다. 이후 `*value = 100`은 지역 변수 `replacement`를 수정합니다.

| 함수 | 받은 값 | 실제 변경 대상 | 호출 뒤 `number` |
| --- | --- | --- | ---: |
| `added(number, 3)` | 정수 `10` | 지역 매개변수 | 10 |
| `addInPlace(&number, 5)` | `number`를 가리키는 포인터 | `number` | 15 |
| `redirect(&number)` | 같은 포인터의 복사본 | 지역 포인터와 `replacement` | 15 |

여기서 `addInPlace`는 포인터 전달을 보여 주는 작은 예제입니다. 정수 오버플로를 막는 범용 계산 함수라는 계약은 없습니다. 본문의 입력은 작은 정수로 제한하며, 범위 검사가 필요한 프로그램은 뒤의 좌표·이체 실습처럼 입력 한도를 명시합니다.


### 2.2 반환값과 포인터 매개변수 중 선택하기

원본을 보존하면서 새 결과만 필요한 함수라면 `func changed(value T) T` 형태가 읽기 쉽습니다. 원래 객체의 상태를 변경하는 것이 함수의 목적이면 `func change(value *T) bool`처럼 변경 대상을 드러낼 수 있습니다.

| 요구사항 | 출발점으로 삼을 형태 | 호출자가 확인할 것 |
| --- | --- | --- |
| 새 계산 결과를 얻기 | 값 매개변수와 반환값 | 반환값을 어디에 저장하는가 |
| 기존 변수의 상태 바꾸기 | 포인터 매개변수 | 실패 시에도 일부 값이 바뀌는가 |
| 값이 없을 수 있음 표현 | 포인터와 `nil` 계약, 또는 값과 `bool` | 없음과 제로 값을 어떻게 구분하는가 |
| 슬라이스 원소 변경 | 보통 슬라이스 매개변수 | 기반 배열이 공유되는가 |
| 맵 항목 변경 | 보통 맵 매개변수 | 맵이 초기화되어 있는가 |

슬라이스 원소나 맵 항목을 바꾸기 위해 무조건 `*[]T`나 `*map[K]V`를 쓸 필요는 없습니다. 05·06장에서 본 것처럼 복사한 슬라이스·맵 값도 원소 저장 공간에 접근합니다. 슬라이스를 `append`해 새 길이를 호출자에게 전달하려면 반환값으로 받는 방식부터 떠올리세요.

## 3. `nil`, `new`, 변수의 수명

### 3.1 초기화하지 않은 포인터

`var pointer *int`의 제로 값은 `nil`입니다. `nil`은 여기서 가리키는 변수가 없다는 뜻입니다. 정수 `0`이 저장된 변수를 가리키는 포인터와는 다릅니다.

다음 전체 프로그램은 그 차이를 보여 줍니다. 개인 파일에서 실행할 수 있습니다.

```go
package main

import "fmt"

func main() {
	var missing *int
	value := 0
	present := &value
	fmt.Println("없음:", missing == nil)
	fmt.Println("0을 가리킴:", present == nil, *present)
}
```

예상 출력:

```text
없음: true
0을 가리킴: false 0
```

`missing == nil` 비교는 가능합니다. 그러나 `*missing`을 읽거나 쓰면 실행 중 패닉이 발생합니다. 패닉을 정상 입력 검사의 대체물로 사용하지 않습니다. 함수가 `nil`을 받을 수 있다면 역참조하기 전에 검사합니다. [명세의 nil 역참조 규칙](https://go.dev/ref/spec#Address_operators)을 참고합니다.

`if pointer == nil || *pointer < 0`에서 왼쪽이 참이면 오른쪽은 평가되지 않습니다. 03장의 단락 평가 때문에 안전하게 검사할 수 있습니다. 순서를 바꾸어 `*pointer < 0 || pointer == nil`로 쓰면 보호되지 않습니다.

### 3.2 `new(T)`는 제로 값 변수의 포인터를 만듭니다

다음은 `main` 안의 코드 조각입니다. `fmt`를 import합니다.

```go
pointer := new(int)
fmt.Println(pointer == nil, *pointer)
*pointer = 7
fmt.Println(*pointer)
```

예상 출력:

```text
false 0
7
```

`new(int)`는 `int` 변수를 제로 값 `0`으로 초기화하고 `*int` 값을 반환합니다. 이름이 붙은 지역 변수의 `&value`를 사용하는 방법과 함께 익혀 두면 됩니다.

`new`와 `make`는 같은 기능이 아닙니다.

| 표현식 | 결과 타입 | 초기 상태 |
| --- | --- | --- |
| `new(int)` | `*int` | 가리키는 정수는 0 |
| `new(Book)` | `*Book` | 구조체의 각 필드는 제로 값 |
| `make([]int, 2)` | `[]int` | 길이 2인 사용 가능한 슬라이스 |
| `make(map[string]int)` | `map[string]int` | 항목을 쓸 수 있는 빈 맵 |
| `new(map[string]int)` | `*map[string]int` | 가리키는 맵 값은 `nil` |

마지막 행의 포인터가 nil이 아니라고 해서 맵까지 초기화된 것은 아닙니다. 맵에 항목을 쓰려면 그 맵 값이 준비되어 있어야 합니다. 포인터 유무와 내부 값의 상태를 따로 확인하세요.

`new(T)`의 규칙은 [Allocation](https://go.dev/ref/spec#Allocation), `make`는 [Making slices, maps and channels](https://go.dev/ref/spec#Making_slices_maps_and_channels)에 있습니다. Go 1.26부터는 `new(표현식)`도 지원하며, 이 장 끝에서 선택 사항으로 확인합니다.

### 3.3 주소가 필요한 이유부터 정합니다

`new`를 사용해야만 포인터를 얻는 것은 아닙니다. 이미 있는 변수에는 `&value`, 구조체 리터럴에는 `&Book{...}`를 사용할 수 있습니다. 주소를 만들었다는 이유만으로 어떤 방식이 항상 더 빠르다고 판단하지 않습니다.

변수의 저장 위치와 수명에 관한 선택 읽기는 뒤에서 다룹니다. 지금은 **포인터를 통해 접근하는 동안 그 변수의 값이 유지된다**는 언어 수준의 동작에 집중합니다.

## 4. 관련 데이터를 구조체로 묶기

책 제목·가격·판매 가능 여부를 서로 다른 변수로 관리하면 어떤 값들이 한 책을 나타내는지 이름만으로 추적해야 합니다. **구조체(struct)**는 서로 관련된 값을 이름 있는 필드로 묶는 타입입니다.

```go
type Book struct {
	Title     string
	Price     int
	Available bool
}
```

이 블록은 타입 선언 조각입니다. 아래 전체 예제에서는 `main` 밖, `package`와 `import` 뒤에 둡니다. 타입 선언 자체가 책 한 권을 생성하지는 않습니다. `Book`이라는 이름으로 만들 값의 모양을 정합니다.

| 부분 | 의미 |
| --- | --- |
| `type Book` | `Book`이라는 타입 이름을 선언 |
| `struct` | 필드들을 묶는 구조체 타입 |
| `Title string` | 이름이 `Title`이고 타입이 `string`인 필드 |
| `Price int` | 이름이 `Price`이고 타입이 `int`인 필드 |
| `Available bool` | 이름이 `Available`이고 타입이 `bool`인 필드 |

타입 이름과 변수 이름은 다릅니다. `book := Book{...}`에서 왼쪽 `book`은 변수, 오른쪽 `Book`은 타입입니다. 같은 타입으로 서로 다른 책 값을 여러 개 만들 수 있습니다.

필드는 구조체 안의 각 값에 붙인 이름입니다. `book.Price`처럼 점을 통해 선택합니다. 다른 패키지에 공개할 이름은 대문자로 시작한다는 규칙은 10장에서 자세히 다룹니다. 지금 예제는 같은 `main` 패키지 안에서 사용합니다.

구조체 정의는 [언어 명세: Struct types](https://go.dev/ref/spec#Struct_types), 이름 공개는 [Exported identifiers](https://go.dev/ref/spec#Exported_identifiers)를 참고합니다.

### 4.1 실행 예제: 제로 값·필드·구조체 복사

저장 경로: [`examples/08-pointers-structs/struct-basics/main.go`](../../examples/08-pointers-structs/struct-basics/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Book struct {
	Title     string
	Price     int
	Available bool
}

func main() {
	var empty Book
	fmt.Printf("제로 값: %q, %d, %t\n", empty.Title, empty.Price, empty.Available)
	book := Book{Title: "Go 첫걸음", Price: 12000, Available: true}
	copied := book
	copied.Price = 8000
	fmt.Printf("원본: %s, %d원\n", book.Title, book.Price)
	fmt.Printf("복사본: %+v\n", copied)
	fmt.Println("같은 구조체 값:", book == copied)
	pointer := &book
	pointer.Available = false
	fmt.Println("판매 가능:", book.Available)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/08-pointers-structs/struct-basics
```

예상 출력:

```text
제로 값: "", 0, false
원본: Go 첫걸음, 12000원
복사본: {Title:Go 첫걸음 Price:8000 Available:true}
같은 구조체 값: false
판매 가능: false
```

`var empty Book`에는 문자열의 `""`, 정수의 `0`, 불리언의 `false`가 들어갑니다. 구조체의 제로 값은 각 필드의 제로 값으로 구성됩니다. 필드마다 별도 초기화 코드를 넣어야 사용할 수 있는 것은 아닙니다.

`Book{Title: ..., Price: ...}`처럼 필드 이름과 값을 적는 표현식을 구조체 리터럴이라고 합니다. 값을 생략한 필드는 제로 값이 됩니다. `Book{Title: "메모"}`라면 가격은 `0`, 판매 가능 여부는 `false`입니다.

`copied := book`은 구조체 값을 복사합니다. 이 구조체의 필드는 문자열·정수·불리언이므로 `copied.Price = 8000`은 원본 가격에 영향을 주지 않습니다.

`pointer := &book` 뒤의 `pointer.Available = false`는 `(*pointer).Available = false`와 같은 필드에 접근합니다. 포인터를 통한 구조체 필드 선택에서는 역참조 표기를 생략할 수 있습니다. nil 포인터에서는 이 접근도 안전하지 않습니다.

`%+v`는 이 구조체의 필드 이름을 함께 출력합니다. 디버깅에는 편하지만 사용자에게 보여 줄 형식을 세밀하게 정할 때는 각 필드를 지정하여 출력하세요.

필드 선택은 [Selectors](https://go.dev/ref/spec#Selectors), 리터럴 초기화는 [Composite literals](https://go.dev/ref/spec#Composite_literals), 제로 값은 [The zero value](https://go.dev/ref/spec#The_zero_value)에 근거합니다.


### 4.2 필드 이름을 적는 이유

`Book{"Go 첫걸음", 12000, true}`처럼 필드 순서대로 값을 나열할 수도 있습니다. 이 장의 실행 코드는 `Book{Title: ..., Price: ..., Available: ...}`처럼 이름을 적습니다. 독자가 각 값의 뜻을 바로 읽고, 같은 타입의 필드를 혼동하지 않게 하기 위한 선택입니다.

필드 이름을 쓰는 리터럴에서는 필요한 필드만 지정할 수 있습니다. 이름을 생략하는 형식은 선언 순서와 필드 수를 정확하게 맞춰야 합니다. 두 형식을 같은 리터럴 안에서 섞지 않습니다.

아래는 `Book` 선언과 `fmt` import가 있는 파일의 `main` 안에 넣는 조각입니다.

```go
book := Book{Title: "메모"}
fmt.Println(book.Price, book.Available)
book.Price = 3000
fmt.Println(book.Price)
```

예상 출력은 첫 줄 `0 false`, 둘째 줄 `3000`입니다. `Title`을 지정했다고 해서 언급하지 않은 필드가 자동으로 특별한 기본값을 갖지는 않습니다.

### 4.3 제로 값이 업무상 유효하다는 뜻은 아닙니다

언어는 `Book{}`를 만들 수 있게 합니다. 이 값이 “판매할 수 있는 책”인지는 프로그램의 규칙입니다. 제목이 비어 있으면 안 되고 가격이 양수여야 한다면, 생성 또는 사용 시점에 검사해야 합니다.

초기화와 검증을 묶은 함수를 흔히 생성 함수라고 부릅니다. `NewBook` 같은 이름은 관례이며 예약어나 특별한 생성자 구문이 아닙니다. 함수는 구조체 값이나 포인터를 반환할 수 있습니다. 실패 여부가 필요하면 이전 장처럼 `bool`을 함께 반환할 수 있습니다.

구조체를 생성할 때 어떤 필드를 검사하는지, 생성 뒤 누구나 그 필드를 바꿀 수 있는지는 서로 다른 문제입니다. 공개 범위를 통한 상태 보호는 09·10장의 메서드·패키지와 연결됩니다. [Effective Go의 생성 함수 설명](https://go.dev/doc/effective_go#composite_literals)을 참고합니다.

### 4.4 구조체를 비교할 수 있는 경우

`Book`의 필드는 모두 비교 가능하므로 두 `Book` 값을 `==`로 비교할 수 있습니다. 필드별 값이 같으면 두 구조체 값도 같습니다. 가격이 다르면 `false`입니다.

그러나 **구조체라면 항상 비교할 수 있는 것은 아닙니다.** `[]int`나 `map[string]int` 필드가 들어 있으면 구조체 전체의 `==` 비교는 허용되지 않습니다. 어떤 필드를 어떤 규칙으로 비교할지 직접 정해야 합니다.

| 필드 예 | 구조체 전체 `==` | 이유 |
| --- | --- | --- |
| `Name string`, `Score int` | 가능 | 모든 필드가 비교 가능 |
| `Position [2]int` | 가능 | 배열 원소인 `int`가 비교 가능 |
| `Owner *Book` | 가능 | 포인터는 비교 가능 |
| `Scores []int` | 불가능 | 슬라이스 전체 비교 불가 |
| `Labels map[string]string` | 불가능 | 맵 전체 비교 불가 |

포인터 필드를 비교할 때는 가리키는 구조체의 모든 내용이 아니라 포인터 값이 비교됩니다. 내용의 동일성과 대상의 동일성을 나누어 판단하세요. [Comparison operators](https://go.dev/ref/spec#Comparison_operators)를 참고합니다.

## 5. 구조체를 변경하는 함수의 계약

학생의 점수나 계좌 잔액처럼 값에 조건이 붙으면, 함수를 만들기 전에 유효 범위와 실패 시 동작을 정하는 것이 도움이 됩니다.

다음 예제의 규칙은 다음과 같습니다.

- 계좌 포인터는 nil이 아니어야 합니다.
- 기존 잔액은 0 이상이어야 합니다.
- 출금액은 양수이며 기존 잔액 이하여야 합니다.
- 성공하면 잔액만 줄이고 `true`를 반환합니다.
- 실패하면 모든 필드를 그대로 두고 `false`를 반환합니다.

이 예제는 정수 단위 금액으로 포인터 변경을 배우기 위한 프로그램입니다. 실제 금융 시스템의 거래 처리나 동시 접근을 구현하지는 않습니다.

### 5.1 실행 예제: 검사한 뒤 원본 갱신하기

저장 경로: [`examples/08-pointers-structs/struct-functions/main.go`](../../examples/08-pointers-structs/struct-functions/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Account struct {
	Owner   string
	Balance int
}

func withdraw(account *Account, amount int) bool {
	if account == nil || account.Balance < 0 || amount <= 0 || amount > account.Balance {
		return false
	}
	account.Balance -= amount
	return true
}

func main() {
	account := Account{Owner: "Mina", Balance: 1000}
	ok := withdraw(&account, 300)
	fmt.Printf("출금: %t, 잔액: %d\n", ok, account.Balance)
	ok = withdraw(&account, 800)
	fmt.Printf("초과 출금: %t, 잔액: %d\n", ok, account.Balance)
	fmt.Println("계좌 없음:", withdraw(nil, 100))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/08-pointers-structs/struct-functions
```

예상 출력:

```text
출금: true, 잔액: 700
초과 출금: false, 잔액: 700
계좌 없음: false
```

조건문이 실패 사례를 먼저 제거합니다. 특히 `account == nil`이 맨 앞에 있어 nil 포인터의 `Balance` 필드를 읽지 않습니다. 뒤의 잔액 검사는 포인터가 있는 경우에만 진행합니다.

모든 검사를 통과한 다음 한 문장으로 잔액을 변경합니다. `amount`가 잔액 이하인 양수이므로 이 뺄셈 결과는 음수가 되지 않습니다. 정수 최솟값으로 내려가는 상황도 입력 계약에서 배제됩니다.

`withdraw`는 성공 여부만 반환합니다. 호출자는 이미 `account`를 가지고 있으므로 변경된 잔액은 `account.Balance`에서 읽습니다. 상태 변경과 결과 반환을 구분해 보세요.

| 호출 전 잔액 | 출금액 | 성공 | 호출 뒤 잔액 |
| ---: | ---: | --- | ---: |
| 1000 | 300 | true | 700 |
| 700 | 800 | false | 700 |
| 10 | 10 | true | 0 |
| 10 | 0 | false | 10 |
| 10 | -1 | false | 10 |

정확히 잔액만큼 출금하는 경계가 성공해야 합니다. `amount >= account.Balance`로 실패 조건을 쓰면 이 규칙과 어긋납니다. 비교 연산자의 등호 하나가 계약을 바꿉니다.


### 5.2 여러 필드를 바꿀 때도 검사가 먼저입니다

한 필드를 바꾼 뒤 다른 조건을 검사하다가 실패하면 일부 상태만 바뀔 수 있습니다. 두 좌표를 이동하거나 두 계좌 사이에서 금액을 옮길 때는 필요한 검사를 먼저 끝내고 수정해야 합니다.

다음 설계 순서를 사용해 보세요.

1. 입력 포인터가 존재하는지 확인합니다.
2. 현재 값과 요청 값이 허용 범위인지 확인합니다.
3. 다음 값을 지역 변수로 계산합니다.
4. 계산한 다음 값이 유효한지 확인합니다.
5. 실제 대상에 대입합니다.

여기서 “검사를 먼저 한다”는 규칙은 이 장의 순차 실행 프로그램에서 실패 시 원본을 보존하기 위한 설계입니다. 여러 고루틴이 동시에 같은 값을 바꿔도 안전하다는 의미는 아닙니다. 공유 데이터의 동시 접근은 20장에서 따로 배웁니다.

## 6. 구조체 안에 구조체를 넣기

필드의 타입은 기본 타입만 가능한 것이 아닙니다. 주소를 하나의 값으로 묶고, 사람의 필드로 넣을 수 있습니다.

다음은 개인 파일에 넣을 수 있는 전체 프로그램입니다.

```go
package main

import "fmt"

type Address struct {
	City   string
	Street string
}

type Person struct {
	Name    string
	Address Address
}

func main() {
	person := Person{
		Name: "Mina",
		Address: Address{
			City:   "Seoul",
			Street: "Sample-ro",
		},
	}
	copied := person
	copied.Address.City = "Busan"
	fmt.Println(person.Name, person.Address.City)
	fmt.Println(copied.Name, copied.Address.City)
}
```

예상 출력:

```text
Mina Seoul
Mina Busan
```

`person.Address.City`는 `Person`의 `Address` 필드 안에 있는 `City`를 선택합니다. 타입 이름과 필드 이름이 모두 `Address`이지만 등장하는 위치가 달라 구분됩니다.

이 예제의 주소에는 문자열 필드만 있으므로 구조체 값을 복사한 후 `copied.Address.City`에 대입해도 원본은 바뀌지 않습니다. 필드를 `*Address`로 선언하면 다른 결과가 생길 수 있습니다. 이때 복사되는 것은 주소 구조체 자체가 아니라 그 구조체를 가리키는 포인터 값이기 때문입니다.

이름을 붙인 중첩 필드와 필드 이름을 생략하는 임베딩은 다른 표기입니다. 임베딩과 메서드 승격은 다음 장에서 다룹니다. 이 장의 핵심 예제는 `person.Address.City`처럼 접근 경로를 직접 드러냅니다.

## 7. 슬라이스 안의 구조체 수정하기

`[]Task`는 여러 `Task` 값을 순서대로 담는 슬라이스입니다. 각 원소는 구조체 값이고, `tasks[i]`를 통해 원래 원소에 접근할 수 있습니다.

반면 `for _, task := range tasks`의 `task`는 각 반복에서 원소 값을 받은 변수입니다. `task.Done = true`는 그 변수의 필드만 수정합니다. `Task`에 `string`과 `bool`만 있는 다음 예제에서는 슬라이스 원소의 값이 바뀌지 않습니다.

### 7.1 실행 예제: range 값과 인덱스 구분하기

저장 경로: [`examples/08-pointers-structs/struct-slices/main.go`](../../examples/08-pointers-structs/struct-slices/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Task struct {
	Title string
	Done  bool
}

func completeByTitle(tasks []Task, title string) bool {
	for i := range tasks {
		if tasks[i].Title == title {
			tasks[i].Done = true
			return true
		}
	}
	return false
}

func main() {
	tasks := []Task{{Title: "읽기"}, {Title: "실행"}}
	for _, task := range tasks {
		task.Done = true
	}
	fmt.Println("복사본 수정 뒤:", tasks[0].Done, tasks[1].Done)
	fmt.Println("검색 결과:", completeByTitle(tasks, "실행"))
	for _, task := range tasks {
		fmt.Printf("%s: %t\n", task.Title, task.Done)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/08-pointers-structs/struct-slices
```

예상 출력:

```text
복사본 수정 뒤: false false
검색 결과: true
읽기: false
실행: true
```

첫 번째 반복문이 끝나도 두 원소의 `Done`은 `false`입니다. 그 반복문은 필드를 쓰기는 했지만 실제 슬라이스 원소에 대입하지 않았습니다.

`completeByTitle`은 인덱스만 순회합니다. `tasks[i].Title`로 원소를 읽고, 이름이 같으면 `tasks[i].Done`에 직접 대입합니다. 찾자마자 반환하므로 같은 제목이 여러 번 있으면 첫 번째 원소만 완료 처리합니다. 이미 완료된 항목을 다시 찾아도 `true`를 반환합니다.

| 표현식 | 가리키는 대상 | `Done`에 대입한 영향 |
| --- | --- | --- |
| `task.Done` | range 변수의 구조체 복사본 | 이 복사본만 변경 |
| `tasks[i].Done` | 슬라이스 기반 배열의 원소 | 호출자가 보는 원소 변경 |
| `pointer := &tasks[i]` 뒤 `pointer.Done` | 같은 원소의 포인터 | 호출자가 보는 원소 변경 |

`[]Task{{Title: "읽기"}, {Title: "실행"}}`는 슬라이스 원소 타입이 `Task`임을 이미 적었기 때문에 내부 리터럴의 타입 이름을 생략한 표기입니다. `[]Task{Task{Title: "읽기"}, Task{Title: "실행"}}`로 읽어도 됩니다.

범위 반복은 [For statements with range clause](https://go.dev/ref/spec#For_range), 중첩 리터럴의 타입 생략은 [Composite literals](https://go.dev/ref/spec#Composite_literals)를 참고합니다.


### 7.2 `&task`와 `&tasks[i]`는 다릅니다

`for _, task := range tasks`에서 `&task`를 얻으면 range 변수의 주소입니다. 원래 슬라이스 원소의 주소가 아닙니다. Go 1.22 이후 `:=`로 선언한 반복 변수는 각 반복에 별도로 생기지만, 이 사실이 그 변수를 원소 자체로 바꾸지는 않습니다.

원래 원소를 가리키려는 목적이라면 인덱스 기반의 `&tasks[i]`를 사용합니다. 다음은 `Task` 타입과 `fmt` import가 있는 `main` 안의 조각입니다.

```go
tasks := []Task{{Title: "읽기"}}
pointer := &tasks[0]
pointer.Done = true
fmt.Println(tasks[0].Done)
```

출력은 `true`입니다. 포인터가 가리키는 원소와 슬라이스의 0번 원소가 같은 변수이기 때문입니다.

슬라이스에 `append`해서 새 기반 배열이 생기면 이전 원소 포인터가 새 배열의 대응 원소로 자동 이동하지는 않습니다. 기존 포인터는 기존 원소를 계속 가리킵니다. 이 장의 검색 포인터 연습은 원소 포인터를 사용하는 동안 슬라이스를 늘리지 않는 조건으로 진행합니다.

## 8. 맵 안의 구조체는 꺼내서 수정한 뒤 넣기

맵에서 얻은 구조체 값도 값입니다. `copied := stock["pen"]`은 구조체 값을 변수에 복사합니다. 이 변수의 정수 필드를 바꾸는 것만으로 맵이 갱신되지는 않습니다.

또한 맵 원소는 일반 변수처럼 주소를 얻을 수 있는 대상이 아닙니다. `&stock["pen"]`이나 `stock["pen"].Quantity = 5`처럼 맵 안의 구조체 필드를 직접 갱신하는 형태는 허용되지 않습니다. 맵 항목 전체 대입인 `stock["pen"] = item`은 가능합니다.

따라서 다음 세 단계를 사용합니다.

1. `item, ok := stock[name]`으로 구조체 값을 꺼냅니다.
2. 지역 변수 `item`의 필드를 수정합니다.
3. `stock[name] = item`으로 수정한 값을 맵에 다시 저장합니다.

맵 인덱스와 대입 가능 대상에 대한 규칙은 [Index expressions](https://go.dev/ref/spec#Index_expressions), [Assignment statements](https://go.dev/ref/spec#Assignment_statements), [Address operators](https://go.dev/ref/spec#Address_operators)에 있습니다.

### 8.1 실행 예제: 재고 항목 갱신하기

저장 경로: [`examples/08-pointers-structs/struct-maps/main.go`](../../examples/08-pointers-structs/struct-maps/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Stock struct {
	Quantity int
}

func addStock(stock map[string]Stock, name string, amount int) bool {
	if stock == nil || amount < 1 || amount > 100 {
		return false
	}
	item, ok := stock[name]
	if !ok || item.Quantity < 0 || item.Quantity > 100-amount {
		return false
	}
	item.Quantity += amount
	stock[name] = item
	return true
}

func main() {
	stock := map[string]Stock{"pen": {Quantity: 3}}
	fmt.Println("입고:", addStock(stock, "pen", 2))
	fmt.Println("맵 수량:", stock["pen"].Quantity)
	copied := stock["pen"]
	copied.Quantity++
	fmt.Println("복사본과 맵:", copied.Quantity, stock["pen"].Quantity)
	fmt.Println("없는 상품:", addStock(stock, "book", 1))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/08-pointers-structs/struct-maps
```

예상 출력:

```text
입고: true
맵 수량: 5
복사본과 맵: 6 5
없는 상품: false
```

이 함수는 등록된 상품만 갱신합니다. 현재 수량은 0 이상, 추가량은 1~100, 최종 수량은 100 이하로 제한합니다. `item.Quantity > 100-amount`를 검사한 뒤 더하므로 허용 범위 밖 수량이나 정수 오버플로를 만들지 않습니다.

`item.Quantity += amount`만 남기고 `stock[name] = item`을 지우면 함수는 `true`를 반환하면서 맵은 바뀌지 않습니다. 코드 한 줄을 지웠을 때 예상 결과가 어떻게 어긋나는지 말해 보세요.

마지막 `copied.Quantity++`도 지역 복사본만 바꿉니다. 출력의 `6 5`가 그 차이를 보여 줍니다. 조회에 성공했는지 확인하지 않고 제로 값을 갱신하면 없는 상품을 새 상품처럼 만들 수 있으므로, 이 계약에서는 반드시 `ok`를 검사합니다.

맵 값 타입을 `*Stock`으로 바꾸는 설계도 가능합니다. 그러면 항목의 포인터를 통해 원본 구조체를 바꿀 수 있지만, nil 값과 공유 대상의 수명까지 함께 정해야 합니다. 이 예제에서는 값 타입과 재대입을 사용해 변경 지점을 드러냅니다.


## 9. 구조체 복사와 내부 데이터의 공유

“구조체는 값 복사된다”와 “구조체 안의 모든 데이터가 완전히 독립된다”는 같은 말이 아닙니다. 구조체 필드에 슬라이스·맵·포인터가 있으면 그 필드의 값도 해당 타입의 규칙대로 복사됩니다.

| 필드 타입 | 구조체 대입 시 복사하는 것 | 복사 뒤 확인할 점 |
| --- | --- | --- |
| `int`, `bool` | 해당 값 | 한쪽 대입은 다른 쪽에 영향 없음 |
| `string` | 문자열 값 | 문자열은 바이트를 직접 수정할 수 없음 |
| `[3]int` | 배열 값과 각 정수 원소 | 배열 원소가 독립됨 |
| `[]int` | 길이·용량·기반 배열을 참조하는 값 | 원소 저장 공간이 공유될 수 있음 |
| `map[string]string` | 맵 데이터에 접근하는 값 | 같은 항목 집합을 공유함 |
| `*Book` | 포인터 값 | 같은 책 변수를 가리킴 |

복사된 상위 값 내부에서 이런 공유가 남는 경우를 흔히 **얕은 복사(shallow copy)**라고 부릅니다. 필요한 내부 데이터까지 새 저장 공간으로 복사하는 작업은 **깊은 복사(deep copy)**라고 부르지만, 어떤 깊이까지 필요한지는 실제 타입 구조와 계약으로 정해야 합니다.

이 장의 `Profile`은 `[]int`와 `map[string]string`을 가지고 있습니다. 슬라이스의 정수 원소와 맵의 문자열 값까지 새 저장 공간에 복사하면 이 타입에서 수정 가능한 내부 데이터를 분리할 수 있습니다. 더 안쪽에 슬라이스나 포인터가 있다면 추가 설계가 필요합니다.

### 9.1 실행 예제: 수정 가능한 필드 분리하기

저장 경로: [`examples/08-pointers-structs/clone-struct/main.go`](../../examples/08-pointers-structs/clone-struct/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Profile struct {
	Name   string
	Scores []int
	Labels map[string]string
}

func cloneProfile(profile Profile) Profile {
	result := profile
	if profile.Scores != nil {
		result.Scores = make([]int, len(profile.Scores))
		copy(result.Scores, profile.Scores)
	}
	if profile.Labels != nil {
		result.Labels = make(map[string]string, len(profile.Labels))
		for key, value := range profile.Labels {
			result.Labels[key] = value
		}
	}
	return result
}

func main() {
	original := Profile{Name: "Mina", Scores: []int{70, 80}, Labels: map[string]string{"team": "go"}}
	shallow := original
	shallow.Name = "Jin"
	shallow.Scores[0] = 99
	shallow.Labels["team"] = "web"
	fmt.Printf("얕은 복사 뒤 원본: %s, %v, %s\n", original.Name, original.Scores, original.Labels["team"])
	cloned := cloneProfile(original)
	cloned.Name = "Sol"
	cloned.Scores[0] = 7
	cloned.Labels["team"] = "cli"
	fmt.Printf("독립 복사 뒤 원본: %s, %v, %s\n", original.Name, original.Scores, original.Labels["team"])
	fmt.Printf("독립 복사본: %s, %v, %s\n", cloned.Name, cloned.Scores, cloned.Labels["team"])
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/08-pointers-structs/clone-struct
```

예상 출력:

```text
얕은 복사 뒤 원본: Mina, [99 80], web
독립 복사 뒤 원본: Mina, [99 80], web
독립 복사본: Sol, [7 80], cli
```

먼저 `shallow := original` 뒤에 세 필드를 바꿉니다. `Name`은 새 문자열 값을 복사본 필드에 대입하므로 원본 이름은 `Mina`입니다. `Scores[0]`과 `Labels["team"]`은 공유한 데이터에 쓰므로 원본에서도 `99`, `web`으로 보입니다.

그다음 `cloneProfile`은 구조체 값을 먼저 복사하고, 수정 가능한 슬라이스와 맵 필드를 각각 새 저장 공간으로 바꿉니다. 결과의 `Scores[0]`과 `Labels["team"]`을 바꿔도 원본은 더 이상 영향을 받지 않습니다.

| 단계 | 원본 이름 | 원본 점수 | 원본 팀 |
| --- | --- | --- | --- |
| 최초 생성 | Mina | `[70 80]` | go |
| 얕은 복사본 수정 | Mina | `[99 80]` | web |
| 독립 복사본 수정 | Mina | `[99 80]` | web |

nil인 필드는 nil로 남기는 계약을 택했습니다. nil이 아닌 빈 슬라이스·맵은 새 빈 값으로 만듭니다. 내용이 비어 있다는 점만 같아도 nil 여부가 호출자에게 의미 있을 수 있기 때문입니다.

이 함수는 모든 타입을 자동 복제하는 범용 함수가 아닙니다. `Profile`에 새로운 참조 성격의 필드를 추가하면 복사 계약과 테스트도 함께 검토해야 합니다. 값 표현과 공유 관계의 언어 규칙은 [Representation of values](https://go.dev/ref/spec#Representation_of_values)를 참고합니다.


### 9.2 원본 보존을 어떻게 검사할까?

값이 같다는 검사만으로 독립 복사가 끝났다고 판단하기 어렵습니다. 원본과 복사본이 같은 배열을 공유해도 최초 출력은 같기 때문입니다.

다음 순서로 확인합니다.

1. 복사 직후 내용이 같은지 확인합니다.
2. 복사본의 수정 가능한 필드를 바꿉니다.
3. 원본의 해당 필드가 그대로인지 확인합니다.
4. 원본도 바꾸어 복사본에 영향이 없는지 확인합니다.
5. nil과 빈 입력의 약속이 지켜지는지 확인합니다.

제공한 테스트는 주요 경계와 관찰 가능한 원본 보존을 검사합니다. 테스트 파일에 보이는 `testing.T`, 테스트 표 구조체, `slices.Equal`의 사용법은 12장에서 익힙니다. 지금은 `go test`로 계약이 유지되는지 확인하면 됩니다.

## 10. 직접 해 보기: 두 구조체를 함께 변경하기

두 계좌 사이에서 금액을 옮기는 함수를 만들어 포인터·구조체·검증 순서를 함께 사용합니다. 이 실습은 순차 실행에서 성공과 실패의 상태 변화를 비교합니다.

### 10.1 개인 파일 준비

아래 명령은 `_practice/08-pointers-structs/transfer/main.go`만 준비합니다. 기존 파일이 있으면 보존합니다. 편집기에서 **그 개인 파일 전체**를 아래 프로그램으로 바꾸세요. 실행 예제 파일을 덮어쓰지 않습니다.

### 준비 명령

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("transfer")) {
    $practiceDir = Join-Path "_practice/08-pointers-structs" $practiceName
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
for practice_name in transfer; do
    practice_dir="_practice/08-pointers-structs/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

한 폴더에는 하나의 `main`만 둡니다. 새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice/`는 공통 `go test ./...`에 포함되지 않으므로 개인 폴더를 직접 지정해서 실행하세요. [공통 연습 안내](../../exercises/README.md)를 참고합니다.


### 10.2 전체 프로그램

`Balance`의 범위를 0~100000으로 정하고, 요청 금액은 1~100000으로 제한합니다. 같은 계좌를 두 번 전달하면 실패하도록 정합니다. 이 규칙들은 언어 규칙이 아니라 이 연습의 계약입니다.

```go
package main

import "fmt"

type Account struct {
	Owner   string
	Balance int
}

func transfer(from, to *Account, amount int) bool {
	if from == nil || to == nil || from == to {
		return false
	}
	if from.Balance < 0 || from.Balance > 100000 || to.Balance < 0 || to.Balance > 100000 {
		return false
	}
	if amount < 1 || amount > 100000 || amount > from.Balance || to.Balance > 100000-amount {
		return false
	}
	from.Balance -= amount
	to.Balance += amount
	return true
}

func main() {
	from := Account{Owner: "Mina", Balance: 1000}
	to := Account{Owner: "Jin", Balance: 200}
	ok := transfer(&from, &to, 300)
	fmt.Printf("이체: %t, 보내는 잔액: %d, 받는 잔액: %d\n", ok, from.Balance, to.Balance)
	ok = transfer(&from, &to, 800)
	fmt.Printf("초과 요청: %t, 보내는 잔액: %d, 받는 잔액: %d\n", ok, from.Balance, to.Balance)
}
```

저장소 루트에서 실행합니다.

```sh
gofmt -w _practice/08-pointers-structs/transfer/main.go
go run ./_practice/08-pointers-structs/transfer
```

예상 출력:

```text
이체: true, 보내는 잔액: 700, 받는 잔액: 500
초과 요청: false, 보내는 잔액: 700, 받는 잔액: 500
```

마지막 조건을 통과하기 전까지 두 계좌의 필드에 쓰지 않습니다. 두 번째 호출에서는 잔액 부족을 발견하고 반환하므로 첫 번째 호출 이후 상태가 유지됩니다.

`from == to`는 계좌 내용이 같은지 비교하지 않습니다. 두 포인터가 같은 계좌 변수를 가리키는지 확인합니다. 서로 다른 계좌의 잔액이 우연히 같아도 이 조건에는 걸리지 않습니다.

`to.Balance > 100000-amount`는 받는 쪽의 한도를 넘는지 덧셈 전에 검사합니다. 이미 `amount`와 잔액의 범위를 검사했기 때문에 이 뺄셈도 범위 안에서 이뤄집니다.

### 10.3 경계 입력을 바꾸어 확인하기

매번 프로그램의 `main`을 처음부터 실행하는 기준입니다. 표의 호출을 단독으로 시험할 때는 두 번째 호출을 제거하거나 주석으로 바꾸고 첫 호출 뒤 상태를 확인합니다.

| 보내는 잔액 | 받는 잔액 | 금액 | 기대 성공 | 기대 잔액 쌍 |
| ---: | ---: | ---: | --- | --- |
| 1000 | 200 | 1000 | true | 0, 1200 |
| 1000 | 99000 | 1000 | true | 0, 100000 |
| 1000 | 99001 | 1000 | false | 1000, 99001 |
| 1000 | 200 | 0 | false | 1000, 200 |
| 1000 | 200 | -1 | false | 1000, 200 |
| -1 | 200 | 1 | false | -1, 200 |
| 100001 | 200 | 1 | false | 100001, 200 |

`transfer(nil, &to, 1)`과 `transfer(&from, nil, 1)`은 실패해야 합니다. `transfer(&from, &from, 1)`도 실패하고 원래 잔액이 유지되어야 합니다. 실패 전후 `Owner` 필드도 바뀌지 않는지 확인하세요.

### 10.4 설계를 말로 설명하기

다음 세 문장을 스스로 완성해 보세요.

- 이 함수는 `Account` 대신 `*Account`를 받으므로 호출자의 ______에 접근한다.
- 포인터 값은 복사되지만, 두 포인터가 가리키는 ______는 같을 수 있다.
- 실패 시 원본을 보존하기 위해 대입보다 ______를 먼저 수행한다.

구현을 외우는 것보다 입력과 상태 변화의 계약을 설명하는 것이 중요합니다. 실패 여부와 실패 이유를 함께 전달하는 방식은 11장의 오류 처리에서 확장합니다.

## 11. 선택 읽기

### 11.1 지역 변수의 주소를 반환해도 되는가?

다음은 개인 파일에서 실행할 수 있는 전체 프로그램입니다.

```go
package main

import "fmt"

func newScore(value int) *int {
	score := value
	return &score
}

func main() {
	first := newScore(70)
	second := newScore(70)
	*first = 90
	fmt.Println(*first, *second)
}
```

예상 출력은 `90 70`입니다. 함수 호출마다 별도의 정수 변수가 생깁니다. 함수가 반환되어도 결과 포인터로 그 변수를 계속 사용할 수 있습니다.

Go 구현은 필요한 수명을 유지합니다. 지역 변수의 이름을 함수 밖에서 사용할 수 없다는 **스코프**와, 그 변수의 값을 포인터로 계속 사용할 수 있는 **수명**은 별개입니다.

메모리가 스택에 있는지 힙에 있는지는 컴파일러의 판단과 프로그램에 따라 달라질 수 있습니다. `&`나 `new`가 등장했다는 이유만으로 항상 같은 저장 위치라고 외우지 않습니다. 실제 성능과 할당 분석은 26장에서 다룹니다. [Go FAQ: stack or heap](https://go.dev/doc/faq#stack_or_heap)을 참고하세요.

### 11.2 `new(표현식)`과 `&변수`는 다른 대상을 만듭니다

Go 1.26부터 `new`는 타입뿐 아니라 표현식도 받을 수 있습니다. 저장소의 Go 1.27 계열에서는 다음 `main` 조각을 실행할 수 있습니다. `fmt`를 import합니다.

```go
number := 10
pointer := new(number)
*pointer = 20
fmt.Println(number, *pointer)
```

예상 출력은 `10 20`입니다. `new(number)`는 `number`의 값을 초기값으로 삼아 **새 변수**를 만들었습니다. `&number`처럼 원래 변수를 가리키지 않습니다.

`new(int)`는 타입을 받아 제로 값, `new(10)`은 표현식 값을 받아 `10`으로 시작하는 새 `int` 변수를 만듭니다. 이는 이전 버전 코드에서 보던 `new(T)` 형식을 확장한 기능입니다. [현재 명세의 Allocation](https://go.dev/ref/spec#Allocation)과 [Go 1.26 언어 변경](https://go.dev/doc/go1.26#language)을 참고합니다.

### 11.3 주소를 얻을 수 있는 표현식

`&`를 모든 표현식 앞에 붙일 수 있는 것은 아닙니다. 일반 변수, 접근 가능한 구조체 필드, 슬라이스 원소 등은 주소를 얻을 수 있는 대상입니다. 구조체 리터럴처럼 별도로 허용한 경우도 있습니다.

| 표현식 | 주소 취득 | 설명 |
| --- | --- | --- |
| `&number` | 가능 | 변수 |
| `&book.Price` | 가능 | 주소를 얻을 수 있는 구조체의 필드 |
| `&scores[0]` | 가능 | 유효한 인덱스의 슬라이스 원소 |
| `&Book{Title: "Go"}` | 가능 | 복합 리터럴에 허용하는 예외 |
| `&stock["pen"]` | 불가능 | 맵 원소의 주소는 얻을 수 없음 |
| `&text[0]` | 불가능 | 문자열 바이트의 주소는 얻을 수 없음 |
| `&10` | 불가능 | 정수 상수 자체는 주소 대상이 아님 |

`new(10)`은 가능하지만 `&10`은 불가능하다는 차이도 여기서 설명됩니다. 전자는 새 변수를 만드는 연산이고, 후자는 주어진 대상의 주소를 요구합니다.

이 장은 일반 Go 포인터만 사용합니다. 포인터를 정수처럼 더해서 다음 위치로 이동하는 연산은 제공하지 않습니다. 배열·슬라이스의 인덱스로 이동하세요. 저수준 `unsafe` 패키지는 이 입문 과정의 필수 내용이 아닙니다.

## 12. 자주 하는 실수와 진단

| 증상 | 먼저 확인할 것 | 수정 방향 |
| --- | --- | --- |
| nil pointer 패닉 | 역참조 또는 필드 접근 전에 nil 검사했는가 | 포인터를 초기화하거나 없음의 계약 처리 |
| 함수 안에서 바꿨는데 호출자 값이 같음 | 정수·구조체 복사본을 바꾼 것은 아닌가 | 반환값을 받거나 원본 포인터를 전달 |
| 포인터를 재대입했는데 원본이 같음 | 포인터 변수와 대상 변수를 혼동했는가 | `pointer`와 `*pointer`의 왼쪽을 구분 |
| range에서 구조체를 바꿨는데 원소가 같음 | range 값 변수만 수정했는가 | `items[i].Field`로 원소 변경 |
| 맵 구조체 필드를 직접 수정할 수 없음 | 맵 원소를 일반 변수처럼 사용했는가 | 꺼내기 → 수정 → 전체 항목 재대입 |
| 구조체 복사 뒤 원본 슬라이스도 바뀜 | 내부 필드가 공유하는 저장 공간이 있는가 | 필요한 내부 데이터까지 복사 |
| 구조체 `==`가 컴파일되지 않음 | 슬라이스·맵 필드가 있는가 | 비교할 필드와 동등성 규칙 설계 |

### 12.1 의도적으로 실패하는 nil 역참조

다음 전체 프로그램은 **의도적으로 실행 중 실패합니다**. 개인 파일에서만 시험하고, 확인 뒤 정상 코드로 되돌립니다.

```go
package main

import "fmt"

func main() {
	var pointer *int
	fmt.Println(*pointer)
}
```

`invalid memory address or nil pointer dereference`가 포함된 패닉이 발생합니다. 경로·줄 번호·스택 정보는 환경에 따라 다릅니다. `pointer := new(int)`로 바꾸면 `0`을 읽을 수 있지만, 실제 프로그램에서는 “값 없음”을 무조건 0으로 바꾸어도 되는지 먼저 정해야 합니다.

### 12.2 의도적으로 실패하는 구조체 비교

다음 전체 프로그램은 **의도적으로 컴파일에 실패합니다**.

```go
package main

func main() {
	type Record struct {
		Scores []int
	}
	left := Record{Scores: []int{1, 2}}
	right := Record{Scores: []int{1, 2}}
	_ = left == right
}
```

`struct containing []int cannot be compared`가 포함된 오류가 납니다. 정수 슬라이스 내용을 비교하려면 길이와 각 원소를 확인하거나 적절한 라이브러리 함수를 사용합니다. 포인터로 바꾸어 비교하는 것은 내용 비교의 해결책이 아닙니다. 그것은 서로 다른 질문을 하는 코드가 됩니다.

## 13. 연습문제

[연습문제 문서](../../exercises/08-pointers-structs/README.md)에서 개인 파일을 준비하고 문제를 풉니다. 정답 코드는 별도 [해설 문서](../../solutions/08-pointers-structs/README.md)에 있습니다.

| 번호 | 수준 | 목표 |
| --- | --- | --- |
| 1 | 확인 | 구조체 복사와 포인터 공유의 출력을 예측 |
| 2 | 변형 | 두 정수 값을 포인터로 교환 |
| 3 | 변형 | 맵에 저장한 구조체 갱신 오류 수정 |
| 4 | 응용 | 이름·점수를 검사해 학생 값 생성 |
| 5 | 응용 | 좌표 이동 실패 시 두 필드 모두 보존 |
| 6 | 응용 | 주문의 원소 슬라이스까지 독립 복사 |
| 선택 A | 응용 | 슬라이스에서 원소 포인터를 찾아 상태 변경 |

참고 구현과 테스트를 확인할 때는 저장소 루트에서 실행합니다.

```sh
go test ./examples/08-pointers-structs/...
go test ./solutions/08-pointers-structs/...
```

테스트가 통과했다고 해서 개인 풀이까지 검사된 것은 아닙니다. 개인 파일은 문제의 명령으로 직접 실행하고, 경계 입력을 바꾸어 보세요.

## 14. 요약과 자기 점검

- 포인터는 다른 변수를 가리키는 값입니다. `&`는 주소를 얻고, `*`는 대상 변수에 접근합니다.
- 포인터 매개변수도 값으로 전달됩니다. 포인터 재대입과 역참조를 통한 수정은 다릅니다.
- 구조체는 관련된 필드를 묶는 타입입니다. 값을 복사하면 필드도 각 타입의 규칙대로 복사됩니다.
- 슬라이스의 구조체 원소를 바꾸려면 실제 원소에 접근합니다. 맵의 구조체는 값을 꺼내 수정한 뒤 다시 저장합니다.
- 내부 슬라이스·맵·포인터가 있으면 공유가 남을 수 있습니다. 복사의 범위를 계약과 테스트로 확인합니다.

해설을 보지 않고 다음 질문에 답해 보세요.

1. `pointer = &other`와 `*pointer = other`는 각각 무엇을 바꾸나요?
2. `var pointer *int`와 `pointer := new(int)`는 nil 여부와 역참조 가능 여부가 어떻게 다른가요?
3. 함수 안에서 포인터 매개변수를 다른 주소로 바꾸어도 호출자의 정수 값이 그대로인 이유는 무엇인가요?
4. `book := Book{Title: "Go"}`에서 생략한 필드는 무엇으로 초기화되나요?
5. `for _, item := range items`에서 `item.Count++`가 원래 구조체 원소를 바꾸지 않는 이유는 무엇인가요?
6. `map[string]Stock`의 수량을 갱신하려면 왜 마지막 항목 대입이 필요한가요?
7. `Profile`을 대입으로 복사한 뒤 `Name` 변경과 `Scores[0]` 변경이 원본에 주는 영향은 어떻게 다른가요?
8. 실패 시 두 필드를 모두 보존하려면 검증·계산·대입을 어떤 순서로 해야 하나요?

자기 점검 해설과 완료 체크리스트는 [해설 문서](../../solutions/08-pointers-structs/README.md)에 있습니다.

## 참고 자료

공식 문서 확인일: **2026-11-09**.

- [Go 명세: Pointer types](https://go.dev/ref/spec#Pointer_types), [Address operators](https://go.dev/ref/spec#Address_operators)
- [Go 명세: Struct types](https://go.dev/ref/spec#Struct_types), [Composite literals](https://go.dev/ref/spec#Composite_literals), [Selectors](https://go.dev/ref/spec#Selectors)
- [Go 명세: Representation of values](https://go.dev/ref/spec#Representation_of_values), [Calls](https://go.dev/ref/spec#Calls), [Comparison operators](https://go.dev/ref/spec#Comparison_operators)
- [Go 명세: Allocation](https://go.dev/ref/spec#Allocation), [Go 1.26 언어 변경](https://go.dev/doc/go1.26#language)
- [Go FAQ: 스택과 힙](https://go.dev/doc/faq#stack_or_heap)
- [Effective Go: 생성 함수와 복합 리터럴](https://go.dev/doc/effective_go#composite_literals)
- [fmt의 출력 형식](https://pkg.go.dev/fmt#hdr-Printing)

## 다음 장

다음 주제는 **09장 메서드와 인터페이스**입니다. 이번 장에서 만든 타입에 동작을 연결하고, 서로 다른 타입이 공통 동작을 제공하는 방식을 배웁니다.

[이전: 07장 문자열과 유니코드](../07-strings-unicode/README.md) · [다음: 09. 메서드와 인터페이스](../09-methods-interfaces/README.md) · [전체 목차](../README.md) · [연습문제](../../exercises/08-pointers-structs/README.md)
