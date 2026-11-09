# 09. 메서드와 인터페이스

08장에서는 구조체로 데이터를 묶고 함수에 값이나 포인터를 전달했습니다. 이번 장에서는 그 타입에 동작을 연결하고, 서로 다른 타입을 공통 동작으로 사용하는 방법을 배웁니다.

책과 사용자는 서로 다른 데이터를 갖습니다. 그래도 둘 다 “설명 문자열을 반환한다”는 동작을 제공하면 같은 출력 함수에 전달할 수 있습니다. 인터페이스는 이런 필요한 동작을 타입으로 표현합니다.

## 학습 목표

- 메서드를 선언하고 수신자(receiver)의 타입과 값을 읽습니다.
- 값 수신자와 포인터 수신자가 원본에 주는 영향을 구분합니다.
- 메서드 집합을 기준으로 `T`와 `*T`가 구현하는 인터페이스를 판단합니다.
- nil 인터페이스와 nil 포인터를 담은 인터페이스를 구분합니다.
- 타입 단언·타입 스위치를 필요한 곳에 사용하고, 인터페이스 비교의 제한을 설명합니다.

## 사전 지식과 준비

[04장 함수](../04-functions-scope/README.md)의 매개변수·반환값·함수 값, [05장 슬라이스](../05-arrays-slices/README.md)의 원소 순회, [08장 포인터와 구조체](../08-pointers-structs/README.md)의 값 복사를 사용합니다. 문자열을 조합하는 개인 실습에서는 07장의 `strings` 패키지도 사용합니다.

명령은 모두 저장소 루트에서 실행합니다. 별도 모듈이나 외부 서비스를 준비하지 않습니다. 예제는 Go 1.27.3에서 확인했으며 공식 자료 확인일은 **2026-11-09**입니다.

이 장에서 말하는 인터페이스는 주로 메서드 목록으로 정의하는 **기본 인터페이스**입니다. 타입 집합을 제한하는 제네릭 제약과 Go 1.27의 제네릭 메서드는 13장에서 다룹니다. 오류를 반환하는 표준 인터페이스는 11·14장에서 연결합니다.

| 읽기 단계 | 확인할 질문 |
| --- | --- |
| 메서드 | 어떤 값을 수신자로 받는가? |
| 포인터 수신자 | 원본을 바꾸는가, 포인터만 재대입하는가? |
| 인터페이스 | 호출자가 요구하는 메서드는 무엇인가? |
| 메서드 집합 | 값 타입과 포인터 타입 중 어느 쪽이 조건을 만족하는가? |
| 동적 타입 | 인터페이스 안에 실제로 어떤 타입의 값이 들어 있는가? |

## 1. 타입에 연결한 함수: 메서드

**메서드(method)**는 수신자가 있는 함수입니다. 수신자는 함수 이름 앞 괄호에 적는 특별한 매개변수이며, 어떤 타입의 값에 대한 동작인지를 나타냅니다.

다음은 선언의 모양을 읽기 위한 조각입니다.

```go
func (score Score) Label() string {
	return fmt.Sprintf("%d점", score.Points)
}
```

`Score` 타입과 `fmt` import가 있는 패키지 수준에 두는 메서드입니다. `main` 함수 안에 선언하지 않습니다. 수신자 `score`는 이 메서드 본문에서 사용하는 이름이고, `Score`는 수신자 타입입니다.

| 부분 | 역할 |
| --- | --- |
| `func` | 함수 또는 메서드 선언 시작 |
| `(score Score)` | 수신자 이름과 타입 |
| `Label` | 메서드 이름 |
| `()` | 일반 인자 목록, 여기서는 없음 |
| `string` | 반환 타입 |
| `score.Points` | 수신자의 필드 읽기 |

`value.Label()`로 호출하면 `value`가 수신자 값으로 전달됩니다. 메서드 안에서도 04장과 같은 값 전달 규칙이 적용됩니다. 메서드의 공식 문법은 [Method declarations](https://go.dev/ref/spec#Method_declarations)를 참고합니다.

### 1.1 실행 예제: 값을 받아 새 값 반환하기

저장 경로: [`examples/09-methods-interfaces/value-methods/main.go`](../../examples/09-methods-interfaces/value-methods/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Score struct{ Points int }

func (score Score) Added(amount int) Score {
	score.Points += amount
	return score
}

func (score Score) Label() string {
	return fmt.Sprintf("%d점", score.Points)
}

func main() {
	original := Score{Points: 70}
	changed := original.Added(5)
	fmt.Println("원본:", original.Label())
	fmt.Println("새 값:", changed.Label())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/09-methods-interfaces/value-methods
```

예상 출력:

```text
원본: 70점
새 값: 75점
```

`original.Added(5)`는 `original`의 값을 수신자 `score`에 복사합니다. 메서드가 `score.Points`를 바꾸어도 원본의 정수 필드는 바뀌지 않습니다. 변경한 수신자 값을 반환하고 호출자는 그것을 `changed`에 저장합니다.

`Label`은 문자열을 만들어 반환합니다. `fmt.Sprintf`는 `fmt.Printf`와 같은 서식 문법을 사용하지만 화면에 출력하는 대신 문자열을 반환합니다. 실제 출력은 `main`의 `fmt.Println`이 합니다.

| 동작 | 원본 점수 | 지역 수신자 또는 결과 |
| --- | ---: | ---: |
| `original` 생성 | 70 | 없음 |
| `Added(5)` 수신자 복사 | 70 | 70 |
| 메서드 안의 증가 | 70 | 75 |
| 반환값 저장 | 70 | `changed.Points == 75` |

이 메서드는 수신자 문법을 배우는 작은 계산 예제입니다. 모든 정수 입력에 대한 오버플로 검사를 약속하지 않습니다. 원본을 변경하는 예제에서는 유효 범위를 명시하여 조건 검사도 함께 연습합니다.

메서드를 사용하지 않고 `func added(score Score, amount int) Score`로 작성해도 같은 값 전달을 표현할 수 있습니다. 메서드 문법은 동작이 어느 타입과 연결되는지 호출 지점에서 드러내 줍니다.


### 1.2 메서드를 선언할 수 있는 타입

구조체에만 메서드를 붙일 수 있는 것은 아닙니다. 같은 패키지에서 정의한 정수·문자열 기반 타입에도 메서드를 선언할 수 있습니다.

다음은 개인 파일에 넣을 수 있는 전체 프로그램입니다.

```go
package main

import "fmt"

type Points int

func (points Points) Label() string {
	return fmt.Sprintf("%d점", points)
}

func main() {
	points := Points(80)
	fmt.Println(points.Label())
}
```

출력은 `80점`입니다. `type Points int`는 `int`와 구분되는 새 타입을 정의합니다. `Points(80)`은 값의 타입을 지정한 표현입니다.

수신자의 기반 타입은 메서드와 같은 패키지에서 선언한 정의된 타입이어야 하며, 그 기반 타입 자체가 포인터나 인터페이스일 수는 없습니다. `*Score` 수신자는 `Score`를 기반으로 하므로 허용됩니다. 다른 패키지의 `time.Time`에 현재 패키지에서 직접 새 메서드를 추가할 수는 없습니다.

이 설명은 일반적인 정의된 타입을 기준으로 합니다. 타입 별칭과 제네릭 수신자의 세부 제약은 [현재 메서드 선언 규칙](https://go.dev/ref/spec#Method_declarations)을 참고합니다. 같은 타입의 필드와 메서드에 같은 이름을 사용할 수도 없습니다.

## 2. 원본을 바꾸는 포인터 수신자

수신자 타입을 `*Counter`로 선언하면 포인터 값을 받습니다. 그 포인터를 통해 원래 구조체에 접근할 수 있습니다. 일반 함수에 `*Counter`를 전달한 08장의 규칙과 같습니다.

다음 예제의 카운터는 0~100 범위를 사용합니다. 추가량은 0 이상이고, 추가 후 값이 100 이하여야 합니다. 실패하면 원본을 그대로 두고 `false`를 반환합니다.

### 2.1 실행 예제: 포인터 수신자와 nil 계약

저장 경로: [`examples/09-methods-interfaces/pointer-methods/main.go`](../../examples/09-methods-interfaces/pointer-methods/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Counter struct{ Value int }

func (counter *Counter) Add(amount int) bool {
	if counter == nil || counter.Value < 0 || counter.Value > 100 || amount < 0 || amount > 100-counter.Value {
		return false
	}
	counter.Value += amount
	return true
}

func main() {
	counter := Counter{}
	ok := counter.Add(3)
	fmt.Println("값 변수 호출:", ok, counter.Value)
	pointer := &counter
	ok = pointer.Add(2)
	fmt.Println("포인터 호출:", ok, counter.Value)
	var missing *Counter
	fmt.Println("nil 수신자:", missing.Add(1))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/09-methods-interfaces/pointer-methods
```

예상 출력:

```text
값 변수 호출: true 3
포인터 호출: true 5
nil 수신자: false
```

`counter`는 `Counter` 값 변수인데 `counter.Add(3)`를 호출할 수 있습니다. 이 변수는 주소를 얻을 수 있고, `*Counter`에 해당 메서드가 있으므로 호출이 `(&counter).Add(3)`로 해석됩니다.

`pointer.Add(2)`는 이미 포인터를 가지고 있으므로 그 포인터를 수신자로 전달합니다. 두 호출 모두 같은 구조체에 접근하여 최종 값이 `5`가 됩니다.

`missing`은 nil 포인터입니다. 포인터 수신자 메서드 호출 자체가 항상 즉시 패닉인 것은 아닙니다. 이 메서드는 처음에 `counter == nil`을 검사하므로 `false`를 반환합니다. nil 상태에서 필드에 접근하는 메서드였다면 패닉이 발생할 수 있습니다.

모든 포인터 수신자 메서드가 nil을 허용해야 한다는 뜻은 아닙니다. nil 허용 여부는 메서드의 계약입니다. 이 예제에서는 없음의 처리 방식을 명시적으로 보여 주기 위해 허용했습니다.

자동 주소 취득은 [Calls](https://go.dev/ref/spec#Calls), 포인터 필드 접근은 [Selectors](https://go.dev/ref/spec#Selectors)에 근거합니다.


### 2.2 자동 주소 취득에는 조건이 있습니다

`counter.Add(1)`이 된다고 모든 `Counter` 표현식에 호출할 수 있는 것은 아닙니다. 구조체 리터럴이나 맵에서 꺼낸 구조체 값처럼 주소를 얻을 수 없는 경우에는 같은 축약이 적용되지 않습니다.

| 표현식 | `(*Counter).Add` 호출 | 이유 |
| --- | --- | --- |
| 지역 변수 `counter.Add(1)` | 가능 | 주소를 얻을 수 있는 변수 |
| 포인터 변수 `pointer.Add(1)` | 가능 | 이미 `*Counter` |
| `(&Counter{}).Add(1)` | 가능 | 리터럴 주소가 명시됨 |
| `Counter{}.Add(1)` | 불가능 | 리터럴 값에서 자동 주소 취득 불가 |
| `items[0].Add(1)` (`[]Counter`) | 가능 | 슬라이스 원소의 주소 취득 가능 |
| `items["a"].Add(1)` (`map[string]Counter`) | 불가능 | 맵 원소 주소 취득 불가 |

맵의 값 타입이 `*Counter`라면 조회 결과가 포인터이므로 호출할 수 있습니다. 다만 없는 키에서 나온 nil을 메서드가 어떻게 처리할지는 다시 계약에 따라 결정됩니다.

### 2.3 값 수신자라고 내부 데이터까지 보호되는 것은 아닙니다

값 수신자에서 구조체 전체가 복사되어도 내부 슬라이스·맵·포인터가 공유될 수 있습니다. `func (group Group) RenameFirst()`가 `group.Names[0]`에 쓰면 `Names`의 기반 배열이 공유되는 경우 원본에도 영향이 있습니다.

따라서 “값 수신자니까 읽기 전용이다”라고 판단하지 않습니다. 필드 타입과 실제로 어떤 저장 공간에 쓰는지 함께 읽어야 합니다. 08장의 얕은 복사 규칙이 메서드에도 적용됩니다.

수신자를 고를 때는 변경 의미와 사용 일관성을 먼저 확인합니다. 구조체가 크다는 이유만으로 측정 없이 속도 차이를 단정하지 않습니다. 성능 검토는 26장에서 다룹니다.

## 3. 호출자가 필요한 동작을 인터페이스로 표현하기

**인터페이스(interface)**는 이 장에서는 필요한 메서드의 목록을 나타내는 타입입니다. 다음 선언은 `Describe() string`을 호출할 수 있는 값을 요구합니다.

```go
type Describer interface {
	Describe() string
}
```

인터페이스의 메서드 선언에는 구현 본문이 없습니다. 구현할 타입의 메서드 이름, 매개변수와 반환 타입이 맞아야 합니다. 매개변수 이름은 같을 필요가 없지만 매개변수 타입과 반환 타입은 시그니처 규칙에 맞아야 합니다.

메서드가 맞으면 타입은 인터페이스를 **암묵적으로 구현**합니다. 별도의 `implements` 선언을 쓰지 않습니다. 구체 타입을 정의한 사람이 해당 인터페이스의 이름을 미리 알고 있어야 하는 것도 아닙니다.

인터페이스 타입의 정의와 구현 조건은 [Interface types](https://go.dev/ref/spec#Interface_types), [Implementing an interface](https://go.dev/ref/spec#Implementing_an_interface)에 있습니다.

### 3.1 실행 예제: 서로 다른 타입을 같은 함수에 전달하기

저장 경로: [`examples/09-methods-interfaces/interfaces/main.go`](../../examples/09-methods-interfaces/interfaces/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Describer interface{ Describe() string }

type Book struct{ Title string }

func (book Book) Describe() string { return "책: " + book.Title }

type User struct{ Name string }

func (user User) Describe() string { return "사용자: " + user.Name }

func printDescription(value Describer) {
	fmt.Println(value.Describe())
}

func main() {
	values := []Describer{Book{Title: "Go"}, User{Name: "Mina"}}
	for _, value := range values {
		printDescription(value)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/09-methods-interfaces/interfaces
```

예상 출력:

```text
책: Go
사용자: Mina
```

`Book`에는 `Title`, `User`에는 `Name`이 있습니다. `printDescription`은 이 필드들을 알지 않습니다. 매개변수 `Describer`가 약속한 `Describe`만 호출합니다.

`[]Describer`는 인터페이스 값을 원소로 갖는 슬라이스입니다. 각 원소는 조건을 만족하는 서로 다른 구체 타입의 값을 담을 수 있습니다. 순회할 때도 `value`의 정적 타입은 `Describer`입니다.

실제로 어떤 메서드 구현이 호출되는지는 인터페이스 안에 담긴 구체 값의 타입에 따라 결정됩니다. `Book`이면 책 설명, `User`이면 사용자 설명을 반환합니다. 이 예제에서는 타입 이름을 비교하는 분기문이 필요하지 않습니다.

| 표현 | 정적 타입 | 담긴 구체 타입 | 호출 결과 |
| --- | --- | --- | --- |
| `values[0]` | Describer | Book | 책: Go |
| `values[1]` | Describer | User | 사용자: Mina |

`printDescription`을 새로 작성하지 않아도, `Describe() string`을 가진 새로운 타입을 전달할 수 있습니다. 이 점은 12장에서 작은 대역을 넣어 테스트하는 방법으로 연결됩니다.


### 3.2 인터페이스를 작게 정하는 이유

출력 함수가 설명 문자열만 필요하다면 `Describe` 하나면 충분합니다. 저장·삭제·네트워크 전송까지 포함한 큰 인터페이스를 요구하면 설명만 제공하는 타입도 불필요한 메서드를 구현해야 합니다.

인터페이스를 사용하는 쪽의 필요에서 출발하면 각 함수의 의존 관계를 읽기 쉽습니다. 모든 구조체마다 미리 인터페이스를 하나씩 만들 필요는 없습니다. 서로 다른 구현을 같은 동작으로 다루거나 경계를 분리할 이유가 있을 때 도입합니다.

[Effective Go의 인터페이스 설명](https://go.dev/doc/effective_go#interfaces)에서도 적은 수의 메서드로 행동을 표현하는 방식을 확인할 수 있습니다. 구체 타입을 그대로 받는 함수도 유효한 설계입니다.

### 3.3 인터페이스에 담았다고 모든 필드를 사용할 수 있는 것은 아닙니다

`var value Describer = Book{Title: "Go"}` 뒤에 `value.Title`을 읽을 수는 없습니다. 컴파일러가 이 변수에 대해 아는 약속은 `Describe` 메서드입니다. 어떤 시점에는 `User`가 들어 있을 수도 있습니다.

구체 필드가 꼭 필요하다면 인터페이스 설계가 적절한지 먼저 검토합니다. 특정 구체 타입인지 확인하는 문법은 뒤의 타입 단언에서 배웁니다. 모든 호출마다 구체 타입을 확인해야 한다면 공통 동작의 설계가 맞는지도 점검하세요.

## 4. 메서드 집합: `T`와 `*T`를 구분하기

**메서드 집합(method set)**은 어떤 타입에 속하는 메서드들의 집합입니다. 인터페이스 구현 여부를 판단할 때 이 집합을 사용합니다.

임베딩이 없는 일반적인 정의된 타입 `T`를 기준으로 읽습니다.

| 선언한 수신자 | `T`의 메서드 집합 | `*T`의 메서드 집합 |
| --- | --- | --- |
| `func (value T) Read() int` | 포함 | 포함 |
| `func (value *T) Add(int)` | 포함하지 않음 | 포함 |

지역 변수에서 포인터 수신자 메서드를 편하게 호출할 수 있는 규칙과, 타입이 인터페이스를 구현하는 규칙은 구분됩니다. `counter.Add(1)`이 가능해도 `Counter` 값이 `Add(int)` 인터페이스를 구현하는 것은 아닙니다.

공식 기준은 [Method sets](https://go.dev/ref/spec#Method_sets)입니다. 뒤에서 임베딩에 따른 승격 규칙을 추가합니다.

### 4.1 실행 예제: 값 복사와 포인터 공유를 인터페이스에서도 확인하기

저장 경로: [`examples/09-methods-interfaces/method-sets/main.go`](../../examples/09-methods-interfaces/method-sets/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Reader interface{ Read() int }
type Adder interface{ Add(int) }

type Counter struct{ N int }

func (counter Counter) Read() int       { return counter.N }
func (counter *Counter) Add(amount int) { counter.N += amount }

var _ Reader = Counter{}
var _ Reader = (*Counter)(nil)
var _ Adder = (*Counter)(nil)

func main() {
	counter := Counter{N: 2}
	var snapshot Reader = counter
	var live Reader = &counter
	var adder Adder = &counter
	adder.Add(3)
	fmt.Println("값을 담은 인터페이스:", snapshot.Read())
	fmt.Println("포인터를 담은 인터페이스:", live.Read())
	counter.Add(1)
	fmt.Println("직접 호출 뒤:", counter.Read())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/09-methods-interfaces/method-sets
```

예상 출력:

```text
값을 담은 인터페이스: 2
포인터를 담은 인터페이스: 5
직접 호출 뒤: 6
```

`Counter.Read`는 값 수신자이므로 `Counter`와 `*Counter` 모두 `Reader`를 구현합니다. `Counter.Add`는 포인터 수신자이므로 이 예제의 `Adder`는 `*Counter`가 구현합니다.

`snapshot`에 넣은 것은 `counter`의 구조체 값입니다. 정수 필드가 복사되어 처음 값 `2`를 유지합니다. `live`에는 `&counter`를 넣었으므로 원본을 가리킵니다. `adder.Add(3)` 뒤에 원본 값 `5`를 읽습니다.

인터페이스도 복사 규칙을 없애지 않습니다. 어떤 구체 타입의 값을 담았는지와 그 타입의 공유 성격을 함께 봅니다. 슬라이스·맵 필드를 가진 구조체를 인터페이스에 넣으면 그 내부 데이터까지 항상 분리되는 것은 아닙니다.

`var _ Reader = Counter{}`는 대입 가능한지 컴파일 단계에서 확인하는 선언입니다. `_`는 값을 사용할 변수 이름이 아니라 버리는 식별자입니다. `(*Counter)(nil)`은 타입이 `*Counter`인 nil 포인터를 만드는 표현이며 역참조하지 않습니다.

이런 선언은 의도한 구현 관계를 보여 주는 데 사용할 수 있습니다. 모든 타입에 기계적으로 붙일 필요는 없습니다. 이미 실제 인터페이스 대입이 있다면 그 대입도 같은 조건을 검사합니다.

여기의 `Reader`는 학습용 `Read() int`입니다. 표준 라이브러리 `io.Reader`와 이름이 같아도 메서드 시그니처가 다르므로 다른 인터페이스입니다. `io.Reader`는 14장에서 다룹니다.


### 4.2 의도적으로 실패하는 인터페이스 대입

다음 전체 프로그램은 **의도적으로 컴파일에 실패합니다**. 개인 파일에서 원인을 확인한 뒤 정상 코드로 바꿉니다.

```go
package main

type Adder interface {
	Add(int)
}

type Counter struct {
	N int
}

func (counter *Counter) Add(amount int) {
	counter.N += amount
}

func main() {
	var value Adder = Counter{}
	value.Add(1)
}
```

오류에는 `Counter does not implement Adder (method Add has pointer receiver)`가 포함됩니다. `var value Adder = &Counter{}`로 바꾸면 필요한 메서드 집합을 가진 포인터 값을 담을 수 있습니다.

컴파일러가 값으로부터 포인터를 알아서 만들어 인터페이스에 넣어 주지는 않습니다. 이 규칙 덕분에 인터페이스에 값을 넣을 때 복사할지, 원본을 공유할지가 코드에 드러납니다.

### 4.3 시그니처는 이름만 맞추는 규칙이 아닙니다

`Read() int`가 필요한 인터페이스에 `Read() int64`를 가진 타입을 넣을 수 없습니다. `Add(int)`와 `Add(int) bool`도 다릅니다. 메서드 이름이 같다는 이유만으로 구현한 것으로 보지 않습니다.

메서드가 포인터를 반환하는 경우에도 반환 타입을 임의로 더 구체적인 타입으로 바꿔 맞출 수는 없습니다. 인터페이스가 약속한 함수 타입과 맞는지 확인해야 합니다. 타입 동일성의 기준은 [Type identity](https://go.dev/ref/spec#Type_identity)를 참고합니다.

## 5. 인터페이스의 nil을 정확하게 읽기

인터페이스 값은 개념적으로 **동적 타입**과 그 타입의 **동적 값**을 함께 담습니다. 이는 동작을 이해하기 위한 모형이며 실제 메모리 배치를 직접 조작하라는 의미가 아닙니다.

`var value Describer`는 아무 구체 타입도 담지 않은 nil 인터페이스입니다. 여기에 타입이 있는 nil 포인터를 담으면 구체 타입 정보가 생깁니다. 내부 포인터가 nil이어도 인터페이스 전체는 nil이 아닙니다.

### 5.1 실행 예제: typed nil과 메서드 호출

저장 경로: [`examples/09-methods-interfaces/typed-nil/main.go`](../../examples/09-methods-interfaces/typed-nil/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Labeler interface{ Label() string }
type User struct{ Name string }

func (user *User) Label() string {
	if user == nil {
		return "사용자 없음"
	}
	return "사용자: " + user.Name
}

func main() {
	var empty Labeler
	var pointer *User
	var filled Labeler = pointer
	fmt.Println("빈 인터페이스:", empty == nil)
	fmt.Println("포인터:", pointer == nil)
	fmt.Println("포인터를 담은 인터페이스:", filled == nil)
	fmt.Printf("동적 타입: %T\n", filled)
	fmt.Println(filled.Label())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/09-methods-interfaces/typed-nil
```

예상 출력:

```text
빈 인터페이스: true
포인터: true
포인터를 담은 인터페이스: false
동적 타입: *main.User
사용자 없음
```

`empty`에는 동적 타입이 없습니다. `pointer`는 `*User` 타입의 nil 포인터입니다. `filled`에는 그 포인터 값과 타입 `*User`가 들어 있으므로 `filled == nil`은 false입니다.

| 변수 | 정적 타입 | 동적 타입 | 저장한 값의 의미 | nil 비교 |
| --- | --- | --- | --- | --- |
| `empty` | Labeler | 없음 | 아무 구체 값도 없음 | true |
| `pointer` | *User | 인터페이스가 아님 | 가리키는 사용자 없음 | true |
| `filled` | Labeler | *User | nil인 *User 포인터 | false |

`filled.Label()`은 `(*User).Label`을 호출합니다. 이 구현이 nil을 검사하므로 `사용자 없음`을 반환합니다. `user.Name`부터 접근했다면 같은 호출이 패닉을 일으켰을 것입니다.

`filled != nil`은 “메서드를 실행하면 절대 실패하지 않는다”는 보장이 아닙니다. 인터페이스에 값이 들어 있다는 검사일 뿐입니다. nil 수신자의 처리 방식은 구현과 계약을 읽어야 합니다.

이 규칙은 [Go FAQ의 nil 인터페이스 설명](https://go.dev/doc/faq#nil_error)에 있습니다. 다음 11장에서 `error`도 인터페이스임을 배울 때 같은 함정이 다시 등장합니다.


### 5.2 빈 인터페이스 값을 그대로 반환하기

어떤 함수가 `Labeler`를 반환하며 “없음”을 nil 인터페이스로 약속했다면 없는 경우에는 `return nil`을 사용합니다. 지역 `var pointer *User`를 만들어 `return pointer`로 돌려주면 타입이 담긴 인터페이스가 됩니다.

이번 예제는 그 차이를 보여 주기 위해 typed nil을 의도적으로 담았습니다. 일반적인 생성 함수에서 없음의 표현이 필요하다면 계약을 한 가지로 정하고 성공·실패 경로에서 지키세요.

### 5.3 nil 인터페이스의 메서드는 호출할 수 없습니다

nil 포인터 수신자가 검사 가능한 것과, nil 인터페이스를 호출하는 것은 다릅니다. nil 인터페이스에는 어떤 구현을 실행해야 하는지 정할 동적 타입이 없습니다.

다음 전체 프로그램은 **의도적으로 실행 중 실패합니다**.

```go
package main

type Labeler interface {
	Label() string
}

func main() {
	var value Labeler
	value.Label()
}
```

nil 포인터 역참조 관련 패닉이 발생합니다. 정상 코드에서는 먼저 `value == nil`을 검사하거나 함수의 입력 계약에서 nil을 거절합니다. `recover`를 일반 입력 검사의 대체물로 쓰지 않습니다.

## 6. `any`, 타입 단언과 타입 스위치

### 6.1 `any`는 메서드가 없는 인터페이스의 별칭

`any`는 `interface{}`의 별칭입니다. 모든 타입의 값을 담을 수 있지만, 변수에 아무 연산이나 허용한다는 뜻은 아닙니다. 컴파일러는 `any` 변수에 특정 숫자 연산이나 필드 접근을 바로 허용하지 않습니다.

`var value any = 42`의 정적 타입은 `any`, 동적 타입은 `int`입니다. 구체 타입이 무엇인지 확인해야 하는 경계에서 **타입 단언(type assertion)**을 사용할 수 있습니다.

`number, ok := value.(int)`는 동적 값이 `int`인지 확인합니다. 맞으면 해당 값을 받고 `ok == true`, 아니면 `int`의 제로 값과 `false`를 받습니다. 숫자를 다른 타입으로 변환하는 `int(x)`와는 다른 문법입니다.

### 6.2 실행 예제: 안전한 단언과 분기

저장 경로: [`examples/09-methods-interfaces/type-assertions/main.go`](../../examples/09-methods-interfaces/type-assertions/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func describeValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "값 없음"
	case string:
		return fmt.Sprintf("문자열 %q", typed)
	case int:
		return fmt.Sprintf("정수 %d", typed)
	default:
		return fmt.Sprintf("다른 타입 %T", typed)
	}
}

func main() {
	var value any = 42
	number, ok := value.(int)
	fmt.Println("int 확인:", number, ok)
	text, ok := value.(string)
	fmt.Printf("string 확인: %q %t\n", text, ok)
	for _, item := range []any{"Go", 7, true, nil} {
		fmt.Println(describeValue(item))
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/09-methods-interfaces/type-assertions
```

예상 출력:

```text
int 확인: 42 true
string 확인: "" false
문자열 "Go"
정수 7
다른 타입 bool
값 없음
```

첫 단언은 동적 타입이 `int`라 성공합니다. 둘째는 `string`이 아니어서 빈 문자열과 false를 받습니다. 실패한 경우에도 두 결과 형태를 사용했으므로 패닉이 발생하지 않습니다.

`switch typed := value.(type)`은 **타입 스위치(type switch)**입니다. 일반 값 대신 동적 타입에 따라 분기합니다. 단일 타입 case 안에서 `typed`는 그 구체 타입으로 사용할 수 있습니다.

`case nil`은 nil 인터페이스를 처리합니다. `default`는 나머지 타입을 처리합니다. `*int` 타입의 nil 포인터가 인터페이스에 들어 있으면 동적 타입이 존재하므로 `case nil`이 아니라 다른 타입 쪽으로 갑니다.

`value.(type)`은 타입 스위치의 지정된 위치에서 사용하는 문법입니다. 일반 대입문에서 타입을 얻는 표현식으로 쓸 수 없습니다. 타입 이름을 출력할 때는 `%T`를 사용할 수 있습니다.

명세의 [Type assertions](https://go.dev/ref/spec#Type_assertions)와 [Type switches](https://go.dev/ref/spec#Type_switches)를 참고합니다.


### 6.3 타입 단언은 변환이 아닙니다

다음은 `main` 안에 넣는 조각입니다. `fmt`를 import합니다.

```go
var value any = int64(42)
number, ok := value.(int)
fmt.Println(number, ok)
wide, ok := value.(int64)
fmt.Println(wide, ok)
```

예상 출력:

```text
0 false
42 true
```

`int64`와 `int`는 별도 타입입니다. 표현할 수 있는 숫자가 같아도 단언은 숫자 변환을 하지 않습니다. 필요하다면 성공한 `int64` 값에 대해 범위를 확인하고 명시적 변환을 해야 합니다.

`number := value.(int)`처럼 결과 하나만 받는 형태는 실패 시 패닉을 일으킵니다. 타입을 확신할 수 없는 입력에는 `number, ok := ...` 형태를 사용하세요.

### 6.4 구체 타입 대신 인터페이스로 단언하기

`value.(Describer)`는 동적 타입이 `Describe() string`을 제공하는지 확인합니다. 그 구체 타입이 정확히 `Book`인지 묻는 것과는 다릅니다. 필요한 추가 동작을 선택적으로 사용하는 코드에 응용할 수 있습니다.

인터페이스 안에 또 다른 인터페이스를 겹겹이 쌓는 것으로 이해하지 않습니다. 다른 인터페이스에 대입하거나 단언해도 그 안에는 실제 구체 타입과 값이 들어 있습니다.

### 6.5 `[]T`가 자동으로 `[]any`가 되지는 않습니다

각 `int`는 `any`에 담을 수 있지만 `[]int` 자체를 `[]any`에 대입할 수는 없습니다. 서로 다른 슬라이스 타입이기 때문입니다. 필요하면 원소를 순회하면서 새 `[]any`에 넣습니다.

아래는 `main` 안의 코드 조각입니다.

```go
numbers := []int{1, 2}
values := make([]any, len(numbers))
for i, number := range numbers {
	values[i] = number
}
fmt.Println(values)
```

출력은 `[1 2]`입니다. 변환을 위해 새 슬라이스가 필요하다는 점을 고려하면, 숫자 목록만 처리하는 함수는 `[]int`를 그대로 받는 것이 더 명확할 수 있습니다. [Go FAQ의 인터페이스 슬라이스 설명](https://go.dev/doc/faq#convert_slice_of_interface)을 참고합니다.

## 7. 인터페이스 비교와 동적 타입

인터페이스 두 개의 `==` 비교는 코드가 컴파일되더라도 동적 타입에 따라 실행 중 실패할 수 있습니다. 인터페이스 자체의 비교 가능성과 그 안에 담긴 값의 비교 가능성을 나누어 봅니다.

다음은 `main` 안의 조각입니다.

```go
var left any = int(3)
var same any = int(3)
var different any = int64(3)
fmt.Println(left == same)
fmt.Println(left == different)
```

출력은 `true`, `false`입니다. 값뿐 아니라 동적 타입도 같아야 두 인터페이스가 같습니다.

### 7.1 의도적으로 실패하는 비교

다음 전체 프로그램은 **컴파일은 되지만 실행 중 패닉이 발생합니다**.

```go
package main

import "fmt"

func main() {
	var left any = []int{1, 2}
	var right any = []int{1, 2}
	fmt.Println(left == right)
}
```

오류에는 `comparing uncomparable type []int`가 포함됩니다. 같은 동적 타입인 `[]int`는 전체 값을 비교할 수 없으므로 인터페이스 비교도 수행할 수 없습니다.

이 성질은 인터페이스 필드를 가진 구조체나 인터페이스 원소 배열의 비교에도 영향을 줍니다. `map[any]...`에 비교 불가능한 동적 값을 키로 넣는 것도 실행 중 실패할 수 있습니다. 키 종류를 알고 있다면 `map[string]...`처럼 구체 타입을 사용하는 편이 계약을 명확히 합니다.

`value == nil`로 인터페이스가 비었는지 확인하는 것은 위와 다릅니다. 그 비교는 안에 담긴 슬라이스 원소끼리 비교하는 연산이 아닙니다. 예를 들어 `var value any = []int{1}`이면 `value == nil`은 false입니다.

정확한 규칙은 [Comparison operators](https://go.dev/ref/spec#Comparison_operators)를 참고합니다. “인터페이스는 비교 가능하다”라는 한 문장만으로 입력 전체의 안전성을 판단하지 않습니다.

## 8. 임베딩으로 구성하기

08장에서는 `Address Address`처럼 이름 있는 필드를 사용했습니다. **임베딩(embedding)**은 구조체 필드 자리에 타입을 적고 별도 필드 이름을 생략하는 방식입니다. 내부 필드와 메서드에 대한 선택자를 짧게 사용할 수 있게 합니다.

다음 예제는 계기 값 `Meter`를 이름 있는 계기 `NamedMeter` 안에 포함합니다. 내부 데이터가 없어지는 것이 아니라 `Meter` 필드로 여전히 존재합니다.

### 8.1 실행 예제: 내부 메서드의 승격

저장 경로: [`examples/09-methods-interfaces/embedding/main.go`](../../examples/09-methods-interfaces/embedding/main.go). 전체 코드입니다.

```go
package main

import "fmt"

type Reader interface{ Read() int }
type Meter struct{ Value int }

func (meter Meter) Read() int       { return meter.Value }
func (meter *Meter) Add(amount int) { meter.Value += amount }

type NamedMeter struct {
	Meter
	Name string
}

func main() {
	meter := NamedMeter{Meter: Meter{Value: 2}, Name: "학습"}
	meter.Add(3)
	var reader Reader = meter
	fmt.Printf("%s: %d\n", meter.Name, reader.Read())
	fmt.Println("명시적 내부 접근:", meter.Meter.Read())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/09-methods-interfaces/embedding
```

예상 출력:

```text
학습: 5
명시적 내부 접근: 5
```

`meter.Read()`는 `meter.Meter.Read()`로 접근할 수 있는 메서드를 짧게 선택한 것입니다. 이를 **메서드 승격(promotion)**이라고 합니다. `Read`의 수신자는 실제 내부 `Meter`이며 바깥 `NamedMeter` 전체가 아닙니다.

`meter.Add(3)`는 주소를 얻을 수 있는 바깥 변수의 내부 `Meter`에 접근하여 값을 바꿉니다. `Read`는 값 수신자이므로 `NamedMeter` 값의 메서드 집합에 승격되어 `Reader` 대입이 가능합니다.

이 구성은 다른 언어의 하위 클래스 상속과 동일한 개념이 아닙니다. `NamedMeter` 값을 `Meter` 매개변수에 자동으로 대입할 수는 없습니다. 필요한 내부 값은 `meter.Meter`로 선택합니다.

초기화에는 `NamedMeter{Meter: Meter{Value: 2}, Name: "학습"}`처럼 포함 관계를 직접 적었습니다. Go 1.27은 구조체 리터럴의 키에 사용할 수 있는 필드 선택자를 확장했지만, 이 장은 구조가 드러나는 초기화를 사용합니다. [Go 1.27 언어 변경](https://go.dev/doc/go1.27#language)을 참고합니다.


### 8.2 임베딩의 메서드 집합

일반적인 `T`가 포인터나 인터페이스가 아닌 정의된 타입일 때, 구조체 `S`가 필드로 무엇을 임베딩하는지에 따라 승격되는 메서드가 달라집니다.

| 구조체 필드 | `S`에 승격 | `*S`에 승격 |
| --- | --- | --- |
| `T` | 수신자 `T` 메서드 | 수신자 `T`, `*T` 메서드 |
| `*T` | 수신자 `T`, `*T` 메서드 | 수신자 `T`, `*T` 메서드 |

`*T`를 임베딩하면 구조체의 제로 값에서 내부 포인터가 nil입니다. 메서드가 보인다는 것과 안전하게 실행할 수 있다는 것은 다시 구분해야 합니다. nil 내부 포인터의 필드 접근이 필요한 경로는 패닉을 일으킬 수 있습니다.

동일한 깊이의 임베딩 경로에 같은 이름이 둘 있으면 짧은 선택자가 모호해질 수 있습니다. 이때는 `value.Left.Read()`처럼 경로를 명시하거나 바깥 타입에 의도를 가진 메서드를 작성합니다. 깊이와 승격의 정확한 기준은 [Struct types](https://go.dev/ref/spec#Struct_types)와 [Selectors](https://go.dev/ref/spec#Selectors)에 있습니다.

### 8.3 인터페이스를 조합하기

기본 인터페이스끼리 필요한 메서드를 조합할 수도 있습니다. 다음은 타입 선언 조각입니다.

```go
type Reader interface {
	Read() int
}

type Resetter interface {
	Reset()
}

type ResettableReader interface {
	Reader
	Resetter
}
```

`ResettableReader`를 구현하려면 `Read() int`와 `Reset()`을 모두 제공해야 합니다. “둘 중 하나”라는 뜻이 아닙니다. 여러 경로에서 같은 이름의 메서드가 있으면 시그니처가 일치해야 합니다.

선택적으로 읽어도 되는 구성 규칙입니다. 핵심 연습은 한두 메서드의 작은 인터페이스부터 작성합니다. [Embedded interfaces](https://go.dev/ref/spec#Embedded_interfaces)를 참고하세요.

## 9. 직접 해 보기: 교체 가능한 문자열 형식

문자열을 그대로 꾸미는 형식과 대문자로 바꾸는 형식을 같은 함수에 전달합니다. 출력 목록을 만드는 함수는 형식의 종류를 알 필요가 없고 `Format`만 호출합니다.

### 9.1 개인 파일 준비

아래 명령으로 `_practice/09-methods-interfaces/format/main.go`를 준비합니다. 이미 파일이 있으면 보존합니다. 준비 뒤 그 개인 파일의 전체 내용을 아래 프로그램으로 바꿉니다.

### 준비 명령

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("format")) {
    $practiceDir = Join-Path "_practice/09-methods-interfaces" $practiceName
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
for practice_name in format; do
    practice_dir="_practice/09-methods-interfaces/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

한 폴더에는 하나의 `main`만 둡니다. 새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice/`는 공통 `go test ./...`에 포함되지 않으므로 개인 폴더를 직접 지정해서 실행하세요. [공통 연습 안내](../../exercises/README.md)를 참고합니다.


### 9.2 전체 프로그램

```go
package main

import (
	"fmt"
	"strings"
)

type Formatter interface {
	Format(string) string
}

type PrefixFormatter struct {
	Prefix string
}

func (formatter PrefixFormatter) Format(text string) string {
	return formatter.Prefix + text
}

type UpperFormatter struct{}

func (formatter UpperFormatter) Format(text string) string {
	return strings.ToUpper(text)
}

func formatAll(words []string, formatter Formatter) ([]string, bool) {
	if formatter == nil {
		return nil, false
	}
	var result []string
	for _, word := range words {
		result = append(result, formatter.Format(word))
	}
	return result, true
}

func main() {
	words := []string{"Go", "한글"}
	prefixed, ok := formatAll(words, PrefixFormatter{Prefix: "[학습] "})
	fmt.Printf("접두사: %q, 성공: %t\n", prefixed, ok)
	upper, ok := formatAll(words, UpperFormatter{})
	fmt.Printf("대문자: %q, 성공: %t\n", upper, ok)
	fmt.Printf("원본: %q\n", words)
}
```

저장소 루트에서 실행합니다.

```sh
gofmt -w _practice/09-methods-interfaces/format/main.go
go run ./_practice/09-methods-interfaces/format
```

예상 출력:

```text
접두사: ["[학습] Go" "[학습] 한글"], 성공: true
대문자: ["GO" "한글"], 성공: true
원본: ["Go" "한글"]
```

`UpperFormatter struct{}`는 필드가 없는 구조체입니다. 개별 설정 없이 동작만 제공하기 때문에 빈 구조체를 사용했습니다. `UpperFormatter{}`로 그 값을 만듭니다.

`Format(string) string`에서 인터페이스 쪽 매개변수 이름은 생략했습니다. 구현의 `Format(text string) string`과 시그니처가 맞습니다. 반환 타입도 string으로 같아야 합니다.

`formatAll`은 새 결과 슬라이스에 문자열 값을 추가합니다. 원본 `words`의 원소를 대입으로 바꾸지 않으므로 입력 목록이 유지됩니다. 여기에 사용한 두 구현은 모두 값 수신자이며, 일반 구조체 값을 전달하므로 nil 포인터 수신자 문제를 만들지 않습니다.

### 9.3 경계와 새 구현 확인

아래 표는 각 호출을 별도로 시도하는 기준입니다. 반환 목록과 `ok`를 함께 확인하세요.

| 입력 목록 | 형식 | 기대 결과 |
| --- | --- | --- |
| nil | `PrefixFormatter{Prefix: "!"}` | nil 목록, true |
| 빈 목록 | `UpperFormatter{}` | nil 목록, true |
| `[]string{""}` | `PrefixFormatter{Prefix: "!"}` | `[]string{"!"}`, true |
| `[]string{"Go"}` | 빈 접두사 | `[]string{"Go"}`, true |
| `[]string{"Go"}` | nil 인터페이스 | nil 목록, false |

이 프로그램은 유효한 UTF-8 문자열을 입력으로 가정합니다. 외부 바이트를 읽는 프로그램이라면 07장에서 배운 검증 정책을 별도로 적용합니다.

추가 실습으로 `SuffixFormatter`를 정의하고 `Format`이 `text + Suffix`를 반환하게 해 보세요. `formatAll` 안에는 새로운 타입을 알아보는 조건문을 추가하지 않아야 합니다. 동일한 계약을 만족하는 새 구현을 끼워 넣는 연습입니다.

### 9.4 추적 질문

1. `formatAll`이 `Formatter` 대신 `PrefixFormatter`를 받으면 어떤 사용이 어려워지나요?
2. `UpperFormatter`에 필드가 없어도 메서드를 선언할 수 있는 이유는 무엇인가요?
3. 입력에 nil 인터페이스가 들어오는 경우와 `(*PrefixFormatter)(nil)`이 들어오는 경우는 같나요?
4. `Format`을 포인터 수신자로 바꾸면 `PrefixFormatter{...}`를 인터페이스에 넣는 코드도 그대로 가능한가요?

3번의 typed nil 포인터는 인터페이스의 nil 검사를 통과하지만, 값 수신자 메서드를 호출하기 위해 값이 필요하므로 안전하지 않습니다. 이번 계약에서는 그런 입력을 전달하지 않습니다. 일반 함수가 임의의 구현 내부 상태를 모두 검사할 수 있다고 가정하지 마세요.

## 10. 선택 읽기: 표준 인터페이스와 메서드 값

### 10.1 `fmt.Stringer`로 출력 표현 제공하기

표준 라이브러리의 `fmt.Stringer`는 `String() string` 메서드를 요구합니다. 해당 메서드가 있으면 `fmt.Println` 등에서 사용자 정의 문자열 표현을 사용할 수 있습니다.

다음은 개인 파일에서 실행하는 전체 프로그램입니다.

```go
package main

import "fmt"

type Ticket struct {
	Title string
	Done  bool
}

func (ticket Ticket) String() string {
	return fmt.Sprintf("%s (완료=%t)", ticket.Title, ticket.Done)
}

func main() {
	ticket := Ticket{Title: "Go 읽기", Done: true}
	var value fmt.Stringer = ticket
	fmt.Println(value)
}
```

출력은 `Go 읽기 (완료=true)`입니다. `fmt` 패키지 코드를 수정하거나 타입을 등록하지 않아도 메서드로 약속을 만족합니다.

`String` 메서드 안에서 `fmt.Sprintf("%v", ticket)`처럼 자기 자신을 다시 포맷하면 같은 메서드를 반복 호출할 수 있습니다. 예제처럼 각 필드를 포맷하여 반환합니다. [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)의 계약과 [출력 규칙](https://pkg.go.dev/fmt#hdr-Printing)을 참고합니다.

### 10.2 메서드 값을 저장할 때 수신자도 정해집니다

`saved := value.Read`처럼 메서드를 함수 값으로 저장할 수 있습니다. 이때 수신자 표현식이 평가되어 보관됩니다. 값 수신자는 당시의 값, 포인터 수신자는 당시의 포인터 값이 저장됩니다.

다음 전체 프로그램으로 비교합니다.

```go
package main

import "fmt"

type Counter struct {
	N int
}

func (counter Counter) Read() int { return counter.N }
func (counter *Counter) Add(amount int) { counter.N += amount }

func main() {
	counter := Counter{N: 2}
	read := counter.Read
	add := counter.Add
	counter.N = 10
	add(3)
	fmt.Println(read(), counter.Read())
}
```

예상 출력은 `2 13`입니다. `read`는 당시의 구조체 값을 보관합니다. `add`는 원본 변수의 포인터를 통해 현재 값을 바꿉니다. 04장의 함수 값 개념에 수신자의 저장 규칙을 더한 것입니다. [Method values](https://go.dev/ref/spec#Method_values)를 참고합니다.

메서드 표현식 `Counter.Read`는 수신자를 첫 인자로 받는 함수로 사용할 수 있습니다. 이 형태는 지금 필수는 아니며 [Method expressions](https://go.dev/ref/spec#Method_expressions)에 자세한 규칙이 있습니다.

## 11. 자주 하는 실수

| 증상 | 원인 후보 | 확인·수정 |
| --- | --- | --- |
| 메서드가 바꾼 정수 필드가 원본에 없음 | 값 수신자 복사본을 수정 | 반환값을 받거나 포인터 수신자 계약 사용 |
| `T`가 인터페이스를 구현하지 못함 | 필요한 메서드가 `*T`에만 있음 | 메서드 집합을 적어 보고 포인터 전달 여부 결정 |
| 지역 변수에서는 호출되는데 대입은 실패 | 자동 주소 취득과 구현 조건 혼동 | 두 규칙을 따로 적용 |
| nil처럼 출력되는데 `== nil`은 false | typed nil이 인터페이스에 들어감 | 동적 타입과 동적 값을 구분 |
| nil 체크 뒤에도 메서드가 패닉 | 내부 포인터 또는 구현의 입력 조건 위반 | 인터페이스 nil과 구현 상태를 분리 |
| `int64` 값의 `.(int)`가 실패 | 타입 단언을 숫자 변환으로 오해 | 먼저 실제 타입 확인 후 필요한 변환 |
| 인터페이스 비교 중 패닉 | 비교 불가능한 동적 타입 | 비교 정책을 정하거나 구체 타입으로 제한 |
| 임베딩한 이름을 선택할 수 없음 | 동일 깊이의 경로가 모호 | 내부 경로를 명시하거나 바깥 메서드 구현 |
| String 메서드가 반복 호출됨 | 자신의 값을 다시 같은 형식으로 포맷 | 원시 필드들을 조합하여 반환 |

모든 문제를 `any`나 타입 단언으로 우회하지 않습니다. 컴파일 오류가 필요한 타입 관계를 알려 주는 경우가 많습니다. 메서드 이름만 확인하지 말고 수신자·인자·반환 타입까지 읽으세요.

### 11.1 의도적으로 실패하는 단일 결과 단언

다음 전체 프로그램은 실행 중 실패합니다.

```go
package main

func main() {
	var value any = "Go"
	_ = value.(int)
}
```

오류에는 `interface conversion`과 `string`, `int`가 포함됩니다. `number, ok := value.(int)`로 바꾸고 `ok`를 검사하면 실패를 정상 분기로 처리할 수 있습니다.

이 실패 예제를 테스트 대상 디렉터리에 저장하지 않습니다. 개인 파일에서 확인한 뒤 되돌립니다. 예상된 무효 입력에는 보통 두 결과 단언과 명시적인 분기가 적합합니다.

## 12. 연습문제

[문제 문서](../../exercises/09-methods-interfaces/README.md)에서 개인 파일을 준비합니다. [해설](../../solutions/09-methods-interfaces/README.md)은 문제를 푼 뒤 확인하세요.

| 번호 | 수준 | 연습 목표 |
| --- | --- | --- |
| 1 | 확인 | 값·포인터 수신자와 인터페이스 복사 출력 예측 |
| 2 | 변형 | 직사각형에 읽기 메서드 두 개 구현 |
| 3 | 변형 | 인터페이스 대입의 포인터 수신자 오류 수정 |
| 4 | 응용 | 한도 있는 카운터의 변경·조회 계약 구현 |
| 5 | 응용 | 여러 구현의 설명 문자열 수집과 nil 계약 |
| 6 | 응용 | 타입 스위치로 동적 타입별 개수 집계 |
| 선택 A | 응용 | 임베딩한 동작을 바깥 메서드에서 조합 |

참고 코드를 검사하려면 저장소 루트에서 실행합니다.

```sh
go test ./examples/09-methods-interfaces/...
go test ./solutions/09-methods-interfaces/...
```

## 13. 요약과 자기 점검

- 메서드는 수신자가 있는 함수이며, 수신자도 값으로 전달됩니다.
- `T`와 `*T`의 메서드 집합은 다릅니다. 직접 호출의 자동 주소 취득과 인터페이스 구현 조건을 구분합니다.
- 인터페이스는 필요한 동작을 표현합니다. 구체 타입은 맞는 메서드들을 제공하여 암묵적으로 구현합니다.
- nil 인터페이스와 typed nil은 다릅니다. 내부 값의 nil 계약은 구현에 달려 있습니다.
- 타입 단언은 변환이 아니며, 인터페이스 비교는 동적 타입에 따라 패닉이 날 수 있습니다.

스스로 답해 보세요.

1. 값 수신자 메서드가 정수 필드를 바꿨는데 원본이 유지되는 이유는 무엇인가요?
2. `counter.Add(1)`이 가능해도 `Counter`가 `Adder`를 구현하지 않을 수 있는 이유는 무엇인가요?
3. `Read() int`가 필요한 인터페이스에 `Read() int64`는 왜 맞지 않나요?
4. 인터페이스에 `(*User)(nil)`을 넣으면 전체 nil 비교는 어떻게 되나요?
5. `any` 변수의 동적 타입이 `int64`일 때 `.(int)`는 변환을 수행하나요?
6. 슬라이스를 담은 두 인터페이스 비교는 어떤 위험이 있나요?
7. 임베딩된 메서드를 호출할 때 그 메서드의 수신자는 내부 타입과 바깥 타입 중 무엇인가요?
8. `formatAll`에 새 구현을 추가할 때 함수 본문을 수정하지 않아도 되는 이유는 무엇인가요?

자기 점검의 답과 완료 체크리스트는 [해설 문서](../../solutions/09-methods-interfaces/README.md)에 있습니다.

## 참고 자료

공식 문서 확인일: **2026-11-09**.

- [Go 명세: Method declarations](https://go.dev/ref/spec#Method_declarations), [Method sets](https://go.dev/ref/spec#Method_sets), [Calls](https://go.dev/ref/spec#Calls)
- [Go 명세: Interface types](https://go.dev/ref/spec#Interface_types), [Implementing an interface](https://go.dev/ref/spec#Implementing_an_interface)
- [Go 명세: Type assertions](https://go.dev/ref/spec#Type_assertions), [Type switches](https://go.dev/ref/spec#Type_switches), [Comparison operators](https://go.dev/ref/spec#Comparison_operators)
- [Go 명세: Struct types](https://go.dev/ref/spec#Struct_types), [Method values](https://go.dev/ref/spec#Method_values)
- [Go FAQ: nil 인터페이스](https://go.dev/doc/faq#nil_error), [인터페이스 슬라이스](https://go.dev/doc/faq#convert_slice_of_interface)
- [Effective Go: 인터페이스](https://go.dev/doc/effective_go#interfaces), [임베딩](https://go.dev/doc/effective_go#embedding)
- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer), [Go 1.27 언어 변경](https://go.dev/doc/go1.27#language)

## 다음 장

다음 주제는 **10장 패키지와 모듈**입니다. 여러 파일과 디렉터리로 코드를 나누고, 공개할 이름과 의존성 경계를 관리합니다.

[이전: 08장 포인터와 구조체](../08-pointers-structs/README.md) · [다음: 10장 패키지와 모듈](../10-packages-modules/README.md) · [전체 목차](../README.md) · [연습문제](../../exercises/09-methods-interfaces/README.md)
