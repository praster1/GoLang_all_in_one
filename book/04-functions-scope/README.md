# 04. 함수와 스코프

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 입문 | [03. 조건문과 반복문](../03-control-flow/README.md) | Go, 편집기, 터미널, 이 저장소 |

03장에서는 `main` 안에서 조건을 판단하고 반복했습니다. 처리할 일이 늘어나면 같은 계산을 여러 곳에 적거나, 긴 코드를 읽으며 어떤 값이 어디에서 바뀌는지 찾아야 합니다. 이번 장에서는 **처리에 이름을 붙이고, 필요한 값을 전달하고, 결과를 받아 조합하는 방법**을 배웁니다.

이미 호출해 본 `fmt.Println`도 함수입니다. 이제 직접 함수를 만들어 호출합니다. 함수로 나누는 기준은 줄 수가 아니라 **입력·결과·책임을 설명할 수 있는가**입니다. 값을 계산하는 함수와 화면에 보여 주는 부분을 구분하고, 이름의 사용 범위인 스코프까지 함께 살펴봅니다.

## 학습 목표

이 장을 마치면 다음을 할 수 있습니다.

- 매개변수와 인자를 구분하여 함수를 선언하고 호출한다.
- 반환값과 화면 출력을 구분하고, 여러 결과와 성공 여부를 받아 처리한다.
- 값 전달과 변수의 스코프를 설명하고 의도하지 않은 이름 가리기를 고친다.
- 가변 인자, 함수값과 익명 함수, 클로저의 기본 동작을 실행해 확인한다.
- 계산과 판단을 작은 함수로 나누고, 경계 입력에 대한 결과를 확인한다.

## 사전 지식과 준비

변수 선언과 대입, 기본 자료형, 비교·논리·산술 연산, `if`·`for`·`range`, 출력문을 사용합니다. 앞 장의 반복 흐름이 낯설다면 [03장](../03-control-flow/README.md)을 먼저 마칩니다.

모든 명령은 **저장소 루트 `GoLang_all_in_one/`**에서 실행합니다. Go 명령은 Windows PowerShell과 macOS·Linux·WSL에서 같습니다. 개인 파일을 준비하는 셸 명령만 구분합니다. 루트의 [`go.mod`](../../go.mod)를 사용하므로 하위 폴더에 새 모듈을 만들지 않습니다.

가변 인자에서는 여러 정수를 다루는 `[]int`와 그 값의 순회를 짧게 소개합니다. 슬라이스의 전체 사용법은 다음 05장에서 이어갑니다. `error`·`defer`·포인터·인터페이스는 각 후속 장에서 다루며, 여기서는 기본 자료형으로 입력과 결과를 표현합니다.

| 실행 예제 | 확인할 내용 | 실행 명령 |
| --- | --- | --- |
| [greeting](../../examples/04-functions-scope/greeting/main.go) | 함수를 선언하고 다른 인자로 호출 | `go run ./examples/04-functions-scope/greeting` |
| [value-return](../../examples/04-functions-scope/value-return/main.go) | 인자 값 복사, 반환과 대입 | `go run ./examples/04-functions-scope/value-return` |
| [divide](../../examples/04-functions-scope/divide/main.go) | 다중 반환, 입력 확인과 이른 반환 | `go run ./examples/04-functions-scope/divide` |
| [scope](../../examples/04-functions-scope/scope/main.go) | 패키지·함수·블록의 이름과 가리기 | `go run ./examples/04-functions-scope/scope` |
| [variadic](../../examples/04-functions-scope/variadic/main.go) | 개수가 달라지는 인자의 합계 | `go run ./examples/04-functions-scope/variadic` |
| [function-values](../../examples/04-functions-scope/function-values/main.go) | 함수를 변수에 보관하고 인자로 전달 | `go run ./examples/04-functions-scope/function-values` |
| [closure](../../examples/04-functions-scope/closure/main.go) | 호출 사이에 상태를 유지하는 함수 | `go run ./examples/04-functions-scope/closure` |

**전체 코드**는 개인 파일에 그대로 작성할 수 있습니다. **함수 선언 발췌**는 `main` 바깥에, **main 안의 발췌**는 `main`의 중괄호 안에 넣습니다. 앞의 예제에 있는 함수가 필요한 발췌는 어느 예제를 사용하는지 명시합니다. 여러 프로그램을 한 파일에 이어 붙이지 않습니다. 10절에서 개인 실습 파일을 준비합니다.

## 1. 함수에 이름을 붙이고 호출하기

### 1.1 같은 처리를 다른 값으로 실행하기

**함수(function)**는 호출할 수 있는 처리 단위입니다. 아래의 `printGreeting`은 전달받은 이름으로 인사 한 줄을 출력합니다.

저장 경로: [`examples/04-functions-scope/greeting/main.go`](../../examples/04-functions-scope/greeting/main.go).

```go
package main

import "fmt"

func printGreeting(name string) {
	fmt.Printf("안녕하세요, %s!\n", name)
}

func main() {
	printGreeting("Go")
	printGreeting("개발자")
}
```

실행:

```sh
go run ./examples/04-functions-scope/greeting
```

예상 출력:

```text
안녕하세요, Go!
안녕하세요, 개발자!
```

| 코드 | 의미 |
| --- | --- |
| `func` | 함수를 선언하는 키워드 |
| `printGreeting` | 이 함수의 이름 |
| `(name string)` | 호출할 때 받을 값의 이름과 타입 |
| `{ ... }` | 호출했을 때 실행할 본문 |
| `printGreeting("Go")` | 문자열 `"Go"`를 전달하여 함수 호출 |

함수를 정의해 두는 것만으로 본문이 실행되지는 않습니다. `main`에서 `printGreeting("Go")`를 호출하면 함수 본문으로 들어가 출력하고, 호출을 마친 뒤 다음 문장으로 돌아옵니다. 두 번째 호출에서는 같은 본문을 다른 이름으로 실행합니다. [함수 선언](https://go.dev/ref/spec#Function_declarations)과 [호출](https://go.dev/ref/spec#Calls)을 참고합니다.

이 예제의 `printGreeting`은 **반환값이 없는 함수**입니다. 본문의 끝에 도달하면 호출한 곳으로 돌아갑니다. Go에서는 반환값이 없다는 뜻으로 `void` 같은 타입을 적지 않고, 매개변수 목록 뒤에 결과 타입을 생략합니다.

### 1.2 매개변수와 인자

- **매개변수(parameter)**: 함수를 선언할 때 받을 값에 붙이는 이름입니다. 위에서는 `name`입니다.
- **인자(argument)**: 함수를 호출할 때 실제로 전달하는 값이나 식입니다. 위에서는 `"Go"`, `"개발자"`입니다.

호출하는 곳의 변수 이름과 매개변수 이름은 같을 필요가 없습니다. 다음은 `greeting` 예제의 `main`을 바꾸는 발췌입니다. `printGreeting` 선언은 그대로 둡니다.

```go
learner := "초보자"
printGreeting(learner)
```

출력은 `안녕하세요, 초보자!`입니다. 호출할 때 `learner`의 값을 읽고, 함수 안에서는 그 값을 매개변수 `name`으로 사용합니다.

일반적인 고정 매개변수 함수는 정해진 개수와 타입에 맞게 호출합니다. `printGreeting()`은 필요한 인자가 없어서, `printGreeting(3)`은 문자열 자리에 정수를 전달해서 컴파일 오류가 발생합니다. 인자 개수를 유연하게 받는 문법은 7절에서 배웁니다.

## 2. 반환값을 받아 사용하기

### 2.1 출력하는 것과 값을 돌려주는 것

**반환값(return value)**은 함수가 호출한 쪽으로 돌려주는 값입니다. 계산 결과를 화면에 보여 주는 것은 호출한 쪽이 할 수도 있습니다.

저장 경로: [`examples/04-functions-scope/value-return/main.go`](../../examples/04-functions-scope/value-return/main.go).

```go
package main

import "fmt"

func addMinutes(current, extra int) int {
	current += extra
	return current
}

func main() {
	minutes := 30
	updated := addMinutes(minutes, 15)

	fmt.Printf("반환값: %d\n", updated)
	fmt.Printf("원래 값: %d\n", minutes)
	minutes = updated
	fmt.Printf("대입 후: %d\n", minutes)
}
```

실행:

```sh
go run ./examples/04-functions-scope/value-return
```

예상 출력:

```text
반환값: 45
원래 값: 30
대입 후: 45
```

`func addMinutes(current, extra int) int`에서 앞의 `int`는 두 매개변수의 타입이고, 괄호 뒤의 `int`는 반환할 값의 타입입니다. `current, extra int`는 `current int, extra int`를 줄여 쓴 형태입니다. 같은 타입의 매개변수 이름을 묶을 수 있습니다. 매개변수 타입과 결과 타입으로 구성되는 함수의 형태를 **시그니처(signature)**라고 부릅니다. [함수 타입](https://go.dev/ref/spec#Function_types)을 참고합니다.

`return current`는 현재 값을 반환하고 이 함수의 실행을 마칩니다. `updated := addMinutes(minutes, 15)`는 반환된 정수를 `updated`에 저장합니다. `addMinutes` 안에는 출력문이 없으므로 이 함수 자체는 화면에 아무것도 표시하지 않습니다.

| 하고 싶은 일 | 사용할 코드의 예 |
| --- | --- |
| 값을 계산해 호출한 쪽에 전달 | `return current` |
| 반환값을 새 변수에 저장 | `updated := addMinutes(minutes, 15)` |
| 반환값으로 기존 변수를 변경 | `minutes = addMinutes(minutes, 15)` |
| 반환값을 바로 출력 | `fmt.Println(addMinutes(minutes, 15))` |

마지막 행은 `addMinutes`의 실행 결과를 `Println`의 인자로 전달한 것입니다. 함수가 값을 반환한다는 사실만으로 자동 출력되지는 않습니다. 반환값을 사용하지 않고 `addMinutes(30, 15)`만 호출하는 문장도 허용되지만, 이 함수에서는 결과를 버려 관찰할 출력이나 값 변경이 없습니다. 필요한 결과를 어디에 사용할지 정합니다. [표현식 문장](https://go.dev/ref/spec#Expression_statements)과 [return 문](https://go.dev/ref/spec#Return_statements)을 참고합니다.

### 2.2 인자와 결과의 타입 확인하기

다음은 `addMinutes`를 호출하는 서로 다른 방법입니다. 표의 각 행은 독립적으로 읽습니다.

| 호출 | 결과 |
| --- | --- |
| `addMinutes(30, 15)` | 정수 `45` 반환 |
| `addMinutes(0, 0)` | 정수 `0` 반환 |
| `addMinutes(30)` | 매개변수에 비해 인자가 부족해 컴파일 오류 |
| `addMinutes(30, 15, 5)` | 인자가 많아 컴파일 오류 |
| `addMinutes("30", 15)` | 문자열을 `int` 매개변수에 전달하여 컴파일 오류 |

타입이 다른 **변수**를 전달할 때는 필요한 변환을 명시합니다. 예를 들어 `var extra int64 = 15`라면 `int` 매개변수에 그대로 넣을 수 없습니다. 02장에서 배운 범위와 변환 규칙을 확인한 뒤 `int(extra)`처럼 변환할지 결정합니다. 상수의 대입 가능성은 변수와 다를 수 있으므로 “숫자 표기가 조금만 달라도 무조건 오류”로 외우지 않습니다.

기본 매개변수의 값을 선언부에 지정하거나, 같은 패키지에서 같은 이름의 일반 함수를 인자 타입만 바꾸어 여러 개 선언하는 방식은 사용하지 않습니다. 입력이나 책임이 다르면 의미가 드러나는 이름을 고르고, 인자 개수가 달라지는 경우에는 가변 인자를 검토합니다. [함수 타입](https://go.dev/ref/spec#Function_types)과 [선언 규칙](https://go.dev/ref/spec#Declarations_and_scope)에 근거한 설명입니다.

## 3. 값 전달: 원래 변수는 왜 그대로일까?

앞의 실행에서 `addMinutes`는 매개변수 `current`를 45로 바꾸었지만 `main`의 `minutes`는 대입하기 전까지 30이었습니다. **인자 값이 매개변수에 복사되어 전달되기 때문**입니다. 이를 **값 전달(pass by value)**이라고 합니다.

| 시점 | `main`의 `minutes` | 함수 안의 `current` | 설명 |
| --- | --- | --- | --- |
| 호출 전 | `30` | 아직 없음 | 호출할 쪽의 값 |
| `addMinutes(minutes, 15)`에 진입 | `30` | `30` | 인자 값을 매개변수에 복사 |
| `current += extra` 실행 | `30` | `45` | 함수의 매개변수만 변경 |
| 반환 후 | `30` | 호출이 끝남 | 반환값 `45`를 `updated`에 저장 |
| `minutes = updated` 실행 | `45` | 호출이 끝남 | 호출한 쪽이 자신의 변수에 대입 |

`minutes`와 `current`는 서로 다른 변수입니다. 같은 이름으로 작성하더라도 다른 함수의 매개변수와 지역 변수는 하나의 변수가 되지 않습니다. 이 예제에서는 반환값을 받아 대입하는 위치가 원래 값을 변경하는 시점입니다.

Go의 함수 인자는 모두 값으로 전달됩니다. 다만 나중에 배울 슬라이스·맵·포인터 같은 값은 **복사된 값이 같은 내부 데이터를 가리킬 수 있습니다.** 그래서 “함수 안에서는 원래 데이터에 어떤 변경도 할 수 없다”로 일반화하면 안 됩니다. 지금은 `int` 값의 복사를 확인하고, 공유 데이터와 변경은 05·06·08장에서 구분합니다. [Go FAQ: 값 전달](https://go.dev/doc/faq#pass_by_value)을 참고합니다.

입력만 받아 계산 결과를 반환하고 출력이나 외부 상태 변경을 하지 않는 함수는 여러 곳에서 쓰고 결과를 검사하기에 편리합니다. 모든 함수가 이런 형태여야 하는 것은 아닙니다. `printGreeting`처럼 출력 자체가 책임인 함수도 있습니다. **이 함수가 무엇을 돌려주고 무엇을 바꾸는지**를 이름과 설명에서 알 수 있게 만듭니다.

## 4. 여러 결과와 성공 여부 반환하기

### 4.1 몫과 나머지를 함께 받기

Go 함수는 여러 값을 반환할 수 있습니다. 다음 함수는 몫·나머지·성공 여부를 함께 돌려줍니다. 이 예제의 입력 규칙은 **피제수는 0 이상, 제수는 양수**입니다. Go 자체가 음수 나눗셈을 금지한다는 뜻은 아닙니다. 이 함수가 처리할 범위를 정한 것입니다.

저장 경로: [`examples/04-functions-scope/divide/main.go`](../../examples/04-functions-scope/divide/main.go).

```go
package main

import "fmt"

func divide(dividend, divisor int) (int, int, bool) {
	if dividend < 0 || divisor <= 0 {
		return 0, 0, false
	}
	return dividend / divisor, dividend % divisor, true
}

func main() {
	quotient, remainder, ok := divide(17, 5)
	if !ok {
		fmt.Println("입력 확인")
		return
	}
	fmt.Printf("몫: %d, 나머지: %d\n", quotient, remainder)

	_, _, ok = divide(17, 0)
	fmt.Printf("0으로 나누기 가능: %t\n", ok)
}
```

실행:

```sh
go run ./examples/04-functions-scope/divide
```

예상 출력:

```text
몫: 3, 나머지: 2
0으로 나누기 가능: false
```

반환 타입 `(int, int, bool)`은 세 결과의 순서와 타입을 나타냅니다. 성공하면 `return dividend / divisor, dividend % divisor, true`로 세 값을 돌려줍니다. 호출한 쪽은 `quotient, remainder, ok := ...`로 같은 순서에 맞추어 받습니다. 여러 값을 하나의 변수에 자동으로 묶어 주는 문법은 아닙니다. [반환](https://go.dev/ref/spec#Return_statements)과 [대입](https://go.dev/ref/spec#Assignment_statements)을 참고합니다.

| 입력 | 몫 | 나머지 | `ok` | 의미 |
| --- | --- | --- | --- | --- |
| `17, 5` | `3` | `2` | `true` | 정상 계산 |
| `0, 5` | `0` | `0` | `true` | 0을 나눈 정상 결과 |
| `3, 5` | `0` | `3` | `true` | 제수가 더 큼 |
| `17, 0` | `0` | `0` | `false` | 제수가 0이므로 계산하지 않음 |
| `-1, 5` | `0` | `0` | `false` | 이 함수가 허용하지 않은 입력 |

숫자 두 개가 0이라는 사실만으로 실패를 판단할 수 없습니다. `0, 5`도 정상 결과가 0·0이기 때문입니다. 이 함수의 약속은 **`ok`가 참일 때만 앞의 두 결과를 계산 결과로 사용하는 것**입니다.

`ok`는 특별한 키워드가 아니라 성공 여부를 표현하려고 붙인 변수 이름입니다. 반환 타입이 `bool`이라고 해서 실패가 자동으로 처리되지는 않습니다. 호출한 쪽에서 값을 확인해야 합니다. 실패 이유까지 전달하는 `error`는 11장에서 자세히 다룹니다.

### 4.2 필요한 결과만 선택하기

`_, _, ok = divide(17, 0)`에서는 몫과 나머지를 사용하지 않고 성공 여부만 받았습니다. `_`는 **빈 식별자(blank identifier)**이며 해당 값을 버린다는 뜻입니다. 기존에 선언한 `ok`를 갱신하므로 이 줄은 `:=`가 아니라 `=`를 사용합니다. [빈 식별자](https://go.dev/ref/spec#Blank_identifier)를 참고합니다.

같은 블록에서 `_`를 새 이름으로 세어 `_, _, ok := ...`를 쓰면 새 변수가 없다는 오류가 발생합니다. 반대로 다른 블록에서 새 `ok`를 선언하는 경우는 별개의 변수일 수 있습니다. 이름의 범위를 6절에서 다시 확인합니다.

계산 성공 여부를 무시한 채 숫자 결과만 쓰면 잘못된 입력이 정상 결과처럼 보일 수 있습니다. 여기서는 두 번째 호출에서 **유효한 입력인지 자체를 출력하는 목적**이므로 숫자 두 개를 버렸습니다. 필요한 판단에 쓰이는 결과까지 습관적으로 버리지 않습니다.

## 5. `return`과 함수의 종료 경로

### 5.1 잘못된 입력에서 일찍 돌아오기

`divide`의 처음에는 다음 조건이 있습니다. **함수 선언의 본문에서 발췌**한 코드입니다.

```go
if dividend < 0 || divisor <= 0 {
	return 0, 0, false
}
return dividend / divisor, dividend % divisor, true
```

이 코드는 독립 프로그램이 아니라 앞의 `divide` 함수 안에서 읽습니다. 입력이 규칙을 벗어나면 첫 번째 `return`이 함수를 끝내므로 뒤의 나눗셈은 실행하지 않습니다. 이렇게 조건을 앞에서 확인하고 돌아오는 방식을 **이른 반환(early return)**이라고 합니다.

이 예제에서는 중첩된 `else` 없이도 정상 계산을 이어 쓸 수 있습니다. 첫 경로가 이미 함수를 끝냈기 때문입니다. 함수의 입력 규칙과 각 반환값의 의미를 함께 읽으면 계산을 실행해도 되는 경로가 분명해집니다.

### 5.2 현재 함수만 끝낸다

`return`은 가장 안쪽 반복문만 끝내는 명령이 아닙니다. 실행 중인 **현재 함수**를 끝내고 호출한 쪽으로 돌아갑니다. 반환값이 없는 함수에서는 `return`만 적을 수 있습니다. 다음은 개인 파일에서 실행할 수 있는 전체 코드입니다.

```go
package main

import "fmt"

func announce(minutes int) {
	if minutes <= 0 {
		fmt.Println("학습 시간 확인")
		return
	}
	fmt.Printf("학습 시작: %d분\n", minutes)
}

func main() {
	announce(0)
	announce(25)
	fmt.Println("안내 종료")
}
```

저장 위치: `_practice/04-functions-scope/lab/main.go`. 10절의 준비 명령으로 폴더를 만든 뒤 다른 개인 코드가 있다면 보관하고 작성합니다.

```sh
go run ./_practice/04-functions-scope/lab
```

예상 출력:

```text
학습 시간 확인
학습 시작: 25분
안내 종료
```

첫 호출의 `return`은 `announce(0)`만 끝냈습니다. 그 뒤 `main`은 다음 호출과 마지막 출력까지 진행합니다. 반면 `main`에서 실행한 `return`은 `main`을 끝내며, 이번 장의 순차 프로그램에서는 실행이 종료됩니다. 프로그램 시작·종료 규칙은 [Program execution](https://go.dev/ref/spec#Program_execution)에 있습니다.

| 문장 | 여기서 끝내거나 건너뛰는 범위 |
| --- | --- |
| `continue` | 현재 반복 회차의 남은 본문 |
| `break` | 가장 안쪽의 해당 반복·분기 문장 |
| `return` | 현재 함수의 실행 |

반복문 안에서 `return`을 만나도 그 함수 전체에서 돌아옵니다. 호출한 함수가 반환했다고 호출한 쪽의 반복문까지 자동으로 중단되는 것은 아닙니다.

### 5.3 결과가 있는 함수의 모든 종료 경로 확인

결과 타입을 선언한 함수는 정상적으로 돌아오는 각 경로에서 그 타입과 개수에 맞는 결과를 반환해야 합니다. 예를 들어 점수에 따라 문자열을 돌려준다면 조건을 만족하지 않는 경우의 반환도 필요합니다.

이 장의 함수에서는 **끝까지 도달할 수 있는 경로마다 반환을 둔다**고 확인하면 됩니다. 명세의 정확한 규칙은 함수 본문의 끝이 종료 문장이어야 한다는 것이며, `return` 외에도 종료 문장으로 인정되는 형태가 있습니다. [함수 선언](https://go.dev/ref/spec#Function_declarations)과 [종료 문장](https://go.dev/ref/spec#Terminating_statements)을 참고합니다. 일부러 끝나지 않는 반복으로 반환 누락을 감추는 방식으로 풀이하지 않습니다.

## 6. 스코프와 이름 가리기

### 6.1 이름이 보이는 범위

**스코프(scope)**는 소스 코드에서 어떤 이름을 사용할 수 있는 범위입니다. Go는 선언이 코드의 어느 블록에 있는지로 범위가 정해지는 **렉시컬 스코프(lexical scope)**를 사용합니다. 누가 함수를 호출했는지에 따라 호출자의 지역 변수가 자동으로 보이지는 않습니다. [선언과 스코프](https://go.dev/ref/spec#Declarations_and_scope)를 참고합니다.

| 선언 위치 | 이름을 사용할 수 있는 범위 |
| --- | --- |
| 함수 바깥의 상수·변수·일반 함수 | 같은 패키지의 코드 |
| 함수의 매개변수·이름 있는 결과 변수 | 해당 함수의 본문 |
| 함수 안의 지역 변수·상수 | 선언이 끝난 지점부터 가장 안쪽의 해당 블록 끝까지 |
| `if` 초기화 문장의 변수 | 그 조건문과 이어진 `else` 등의 범위 |
| `for`에서 선언한 반복 변수 | 해당 반복문 안 |
| `import "fmt"`로 들여온 패키지 이름 | 그 import가 적힌 파일 |

안쪽 블록에서는 바깥의 이름을 사용할 수 있지만, 바깥에서 안쪽의 지역 이름을 사용할 수는 없습니다. 같은 패키지에 있는 함수여도 다른 함수의 지역 변수까지 공유하지는 않습니다. 필요한 값은 매개변수로 전달하고 결과는 반환값으로 받습니다.

이 장의 예제에서 `printGreeting`을 `main`보다 아래에 선언해도 호출할 수 있습니다. 일반 함수 이름은 패키지 범위에 있기 때문입니다. 함수들을 파일에 적은 순서가 실행 순서를 결정하지는 않습니다.

함수 안에서 `func helper() { ... }`처럼 **이름 있는 함수 선언**을 중첩하는 형태는 허용하지 않습니다. 함수 안에서 사용할 작은 함수값이 필요하면 8절의 익명 함수를 변수에 담을 수 있습니다.

### 6.2 같은 이름이 다른 변수를 뜻할 때

저장 경로: [`examples/04-functions-scope/scope/main.go`](../../examples/04-functions-scope/scope/main.go).

```go
package main

import "fmt"

const courseName = "Go"

func printCourse() {
	fmt.Println("과정:", courseName)
}

func main() {
	score := 72
	message := "계속 연습"
	if score >= 60 {
		message := "통과"
		fmt.Println("블록 안:", message)
	}
	fmt.Println("블록 밖:", message)
	printCourse()
}
```

실행:

```sh
go run ./examples/04-functions-scope/scope
```

예상 출력:

```text
블록 안: 통과
블록 밖: 계속 연습
과정: Go
```

`message := "통과"`는 안쪽 `if` 블록에서 **새 변수**를 선언합니다. 바깥의 `message`와 이름은 같지만 다른 변수입니다. 안쪽 이름이 바깥 이름을 **가리는 현상(shadowing)**입니다.

| 위치 | 그 자리의 `message`가 가리키는 변수 | 값 |
| --- | --- | --- |
| `if` 앞 | `main`에서 선언한 변수 | `계속 연습` |
| 안쪽 선언 뒤 | `if` 블록의 새 변수 | `통과` |
| `if`가 끝난 뒤 | `main`에서 선언한 변수 | `계속 연습` |

원래 목적이 바깥의 값을 바꾸는 것이라면 안쪽 줄을 `message = "통과"`로 수정합니다. 그러면 두 출력 모두 `통과`가 됩니다. 한 블록 안에서 기존 변수의 값을 바꾸는 것인지, 다른 블록에 별도의 이름을 만드는 것인지 먼저 구분합니다.

같은 블록에서 `:=`로 기존 변수를 다시 사용할 수 있는 경우는 **같은 타입을 유지하고, 빈 식별자가 아닌 새 이름이 적어도 하나 함께 있을 때**입니다. 함수 본문에서는 매개변수도 이 재사용 규칙에 포함됩니다. `:=`가 언제나 새 변수만 만들거나 언제나 기존 변수를 바꾼다고 단정하면 안 됩니다. [짧은 변수 선언](https://go.dev/ref/spec#Short_variable_declarations)을 참고합니다.

### 6.3 패키지 범위와 함수의 입력을 구분하기

예제의 `courseName`은 함수 바깥의 상수이므로 `printCourse`에서도 읽을 수 있습니다. 다음은 `scope` 예제의 `main`을 교체하는 발췌입니다. 패키지 상수와 `printCourse` 선언은 유지합니다.

```go
courseName := "다른 과정"
fmt.Println("main:", courseName)
printCourse()
```

예상 출력:

```text
main: 다른 과정
과정: Go
```

`printCourse`는 자신이 선언된 코드에서 보이는 패키지 상수 `courseName`을 사용합니다. 호출자인 `main`의 지역 변수로 바뀌지 않습니다. 호출에 따라 다른 과정명을 쓰고 싶다면 함수에 매개변수를 추가해서 전달해야 합니다.

여러 함수가 읽는 고정 기준은 패키지 상수로 둘 수 있습니다. 매 호출마다 달라지는 입력이나 결과는 매개변수·반환값으로 드러내면 흐름을 따라가기 쉽습니다. 숨은 패키지 변수에 결과를 저장하는 설계는 이 장의 기본 예제로 사용하지 않습니다.

## 7. 가변 인자로 여러 값 받기

### 7.1 `...int`의 의미

합계를 구할 때 어떤 호출에서는 숫자 하나, 다른 호출에서는 세 개가 필요할 수 있습니다. **가변 인자(variadic parameter)**는 같은 타입의 인자를 0개 이상 받을 수 있게 합니다.

저장 경로: [`examples/04-functions-scope/variadic/main.go`](../../examples/04-functions-scope/variadic/main.go).

```go
package main

import "fmt"

func sum(numbers ...int) int {
	total := 0
	for _, number := range numbers {
		total += number
	}
	return total
}

func main() {
	fmt.Printf("인자 없음: %d\n", sum())
	fmt.Printf("세 값의 합: %d\n", sum(10, 20, 30))
}
```

실행:

```sh
go run ./examples/04-functions-scope/variadic
```

예상 출력:

```text
인자 없음: 0
세 값의 합: 60
```

`numbers ...int`는 호출한 쪽에서 `int` 인자를 여러 개 전달할 수 있다는 뜻입니다. 함수 안에서 `numbers`의 타입은 **`[]int`**, 즉 정수들을 순서대로 다루는 **슬라이스(slice)**입니다. 길이와 저장 공간, 원소 변경 등의 전체 규칙은 05장에서 배우고, 여기서는 전달된 정수를 차례로 읽는 데 사용합니다. [가변 인자 전달 규칙](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)을 참고합니다.

### 7.2 슬라이스의 `range`를 읽기 위한 최소 문법

03장에서는 정수에 대한 `range`가 반복 번호 하나를 제공했습니다. 여기서는 슬라이스를 순회하므로 두 자리에 **위치와 그 위치의 값**을 받을 수 있습니다.

| 코드 | 뜻 |
| --- | --- |
| `numbers ...int` | 호출할 때 받을 개수가 달라질 수 있는 `int` 인자 |
| 함수 내부의 `numbers` | `[]int` 타입으로 전달된 값들을 다룸 |
| `for _, number := range numbers` | 각 위치의 값은 `number`로 받고, 위치는 `_`로 버림 |
| `total += number` | 이번 원소를 합계에 누적 |
| `len(numbers)` | 전달받은 슬라이스의 원소 개수 |

`len`은 길이를 구하는 내장 함수입니다. 슬라이스에서 반복 변수 하나만 적은 `for number := range numbers`는 **값이 아니라 위치**를 받습니다. 합계를 계산하려면 위 예제처럼 두 번째 자리를 사용해야 합니다. `sum(10, 20, 30)`에서 값은 10·20·30이고 위치는 0·1·2이므로, 둘을 혼동하면 전혀 다른 합계가 나옵니다. [range 절](https://go.dev/ref/spec#For_range)과 [길이](https://go.dev/ref/spec#Length_and_capacity)를 참고합니다.

### 7.3 인자가 없을 때와 선언 위치

`sum()`도 유효한 호출입니다. 이때 `numbers`는 `nil` 슬라이스이고 길이는 0입니다. `nil`은 슬라이스의 영값이며, 이 경우 원소가 없어 `range` 본문을 실행하지 않습니다. 처음의 `total`이 0이므로 0을 반환합니다. `nil`에 관한 타입별 세부 동작은 이후 해당 자료형에서 다시 다룹니다.

| 호출 | 순회하는 값 | 반환값 |
| --- | --- | --- |
| `sum()` | 없음 | `0` |
| `sum(7)` | `7` | `7` |
| `sum(10, 20, 30)` | `10, 20, 30` | `60` |

가변 인자는 매개변수 목록의 **마지막 자리**에 하나만 둘 수 있습니다. 예를 들어 `func printLabels(prefix string, labels ...string)`은 고정 인자 다음에 가변 인자를 두는 형태입니다. 반대로 `func bad(values ...int, label string)`은 선언할 수 없습니다. [함수 타입](https://go.dev/ref/spec#Function_types)이 이 제약을 정의합니다.

가변 인자라고 아무 타입이나 섞을 수 있는 것은 아닙니다. 이 함수의 `...int`에는 `int`에 대입할 수 있는 값들을 전달합니다. `sum(10, "20")`은 컴파일 오류입니다. 이미 만들어 둔 슬라이스를 전달하는 `values...` 표기는 12절의 선택 학습에서 확인합니다.

## 8. 함수도 값으로 다루기

### 8.1 함수값을 변수에 담고 전달하기

함수 이름을 호출 괄호 없이 사용하면 함수값을 변수에 담거나 다른 함수의 인자로 전달할 수 있습니다. 다음은 어떤 계산을 할지 함수값으로 선택하는 예제입니다.

저장 경로: [`examples/04-functions-scope/function-values/main.go`](../../examples/04-functions-scope/function-values/main.go).

```go
package main

import "fmt"

func add(a, b int) int {
	return a + b
}

func multiply(a, b int) int {
	return a * b
}

func apply(a, b int, operation func(int, int) int) int {
	return operation(a, b)
}

func main() {
	operation := add
	fmt.Println("덧셈:", operation(3, 4))
	operation = multiply
	fmt.Println("곱셈:", apply(3, 4, operation))

	double := func(value int) int {
		return value * 2
	}
	fmt.Println("두 배:", double(5))
}
```

실행:

```sh
go run ./examples/04-functions-scope/function-values
```

예상 출력:

```text
덧셈: 7
곱셈: 12
두 배: 10
```

`operation := add`는 덧셈을 **지금 실행하는 코드가 아닙니다.** `add`라는 함수값을 변수에 담습니다. 이후 `operation(3, 4)`를 호출해야 덧셈을 수행합니다. `operation = multiply`로 같은 타입의 다른 함수값을 담으면 다음 호출의 동작이 바뀝니다.

| 표현 | 의미 | 여기서의 타입 |
| --- | --- | --- |
| `add` | 함수값 | `func(int, int) int` |
| `add(3, 4)` | 함수를 실행한 결과 | `int` |
| `operation := add` | 함수값을 보관 | `operation`은 함수 타입 |
| `result := add(3, 4)` | 계산 결과를 보관 | `result`는 `int` |

`add`와 `multiply`는 `int` 두 개를 받고 `int` 하나를 반환하므로 같은 함수 타입을 사용합니다. 매개변수의 **이름**은 함수 타입의 일치 여부를 결정하지 않지만, 입력·결과의 타입과 개수 및 가변 인자 여부는 중요합니다. [함수 타입](https://go.dev/ref/spec#Function_types)과 [타입의 동일성](https://go.dev/ref/spec#Type_identity)을 참고합니다.

### 8.2 함수를 매개변수로 받기

`apply`의 매개변수는 다음처럼 읽습니다.

| 매개변수 | 타입 | 역할 |
| --- | --- | --- |
| `a` | `int` | 계산할 첫 값 |
| `b` | `int` | 계산할 두 번째 값 |
| `operation` | `func(int, int) int` | 두 정수를 받아 정수를 반환하는 함수 |

`apply(3, 4, operation)`은 값 두 개와 **수행할 계산**을 전달합니다. `apply` 안에서 `operation(a, b)`를 호출하고 그 결과를 반환합니다. 이 예제에서 `apply`에 전달할 `operation`은 호출 가능한 함수값이라는 전제를 둡니다.

이렇게 함수를 인자로 받거나 결과로 반환하는 함수를 **고차 함수(higher-order function)**라고 부릅니다. 이름보다 동작이 중요합니다. 값을 처리하는 흐름은 같고 수행할 계산만 달라질 때 이런 분리가 유용할 수 있습니다. 단순한 덧셈 하나를 작성할 때마다 반드시 `apply` 같은 함수를 추가하라는 규칙은 아닙니다.

### 8.3 익명 함수

`double := func(value int) int { ... }`는 **익명 함수(anonymous function)**를 만들고 그 함수값을 변수 `double`에 담습니다. Go 명세에서는 이 표기를 **함수 리터럴(function literal)**이라고 합니다. 함수 이름이 `func` 뒤에 없다는 점이 이름 있는 선언과 다릅니다.

`double`은 함수 선언에 붙인 이름이 아니라 **함수값을 보관하는 변수 이름**입니다. 함수 본문 안에서도 이렇게 작은 함수값을 만들 수 있습니다. 앞 예제의 `double(5)`는 10을 반환합니다. [함수 리터럴](https://go.dev/ref/spec#Function_literals)을 참고합니다.

### 8.4 함수 변수의 영값

함수 타입의 변수를 선언하면서 함수값을 넣지 않으면 영값은 `nil`입니다. **아직 호출할 함수가 없다**고 읽으면 됩니다. 다음은 `main` 안의 독립된 발췌입니다.

```go
var operation func(int, int) int
fmt.Println(operation == nil)

operation = func(a, b int) int {
	return a + b
}
fmt.Println(operation(2, 3))
```

예상 출력:

```text
true
5
```

`nil` 상태에서 호출하면 컴파일을 통과하더라도 실행 중 `panic`이 발생합니다. 함수값을 전달받는 설계에서는 호출 가능한 값이 들어오는지 계약을 정하거나 필요하면 `nil`을 확인합니다. 위 코드에서는 호출 전에 실제 함수를 대입했습니다. 이 장의 `apply`도 실제 함수를 전달하는 예제입니다.

함수값은 서로 `==`로 비교할 수 없고 `nil`과의 비교만 허용됩니다. “같은 동작인지”를 함수값 비교로 판단하지 않습니다. [함수 타입](https://go.dev/ref/spec#Function_types), [호출](https://go.dev/ref/spec#Calls), [비교 연산자](https://go.dev/ref/spec#Comparison_operators)를 참고합니다.

## 9. 클로저: 바깥 변수와 함께 동작하는 함수

### 9.1 생성 당시의 값으로 고정되지 않는다

익명 함수는 자신을 둘러싼 함수에서 선언한 변수를 사용할 수 있습니다. 이렇게 바깥 변수에 접근하는 함수값을 **클로저(closure)**라고 부릅니다. 다음은 `main` 안의 독립된 발췌입니다.

```go
bonus := 5
addBonus := func(score int) int {
	return score + bonus
}
bonus = 10
fmt.Println(addBonus(70))
```

출력은 `80`입니다. 익명 함수를 만들 때의 5가 고정되어 75가 나오는 것이 아닙니다. 함수 본문에서 사용하는 `bonus`는 바깥의 같은 변수이므로 호출 시점의 값 10을 읽습니다. 바깥 변수를 이렇게 사용하는 것을 **캡처(capture)**라고도 부릅니다. [함수 리터럴과 클로저](https://go.dev/ref/spec#Function_literals)의 규칙입니다.

매개변수로 정수를 전달할 때는 그 호출의 값을 복사하지만, 이 예제의 `bonus`는 매개변수가 아닙니다. 함수가 둘러싼 범위의 변수를 사용한다는 점을 구분합니다.

### 9.2 함수를 반환하여 상태 이어가기

저장 경로: [`examples/04-functions-scope/closure/main.go`](../../examples/04-functions-scope/closure/main.go).

```go
package main

import "fmt"

func newCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	first := newCounter()
	second := newCounter()

	fmt.Println("첫 번째:", first())
	fmt.Println("첫 번째:", first())
	fmt.Println("두 번째:", second())
	fmt.Println("첫 번째:", first())
}
```

실행:

```sh
go run ./examples/04-functions-scope/closure
```

예상 출력:

```text
첫 번째: 1
첫 번째: 2
두 번째: 1
첫 번째: 3
```

`func newCounter() func() int`는 처음에는 괄호가 많아 보일 수 있습니다. 앞에서부터 다음처럼 나누어 읽습니다.

| 부분 | 의미 |
| --- | --- |
| `func newCounter()` | 인자를 받지 않는 `newCounter` 함수 선언 |
| 뒤의 `func() int` | 반환값은 “인자를 받지 않고 정수를 반환하는 함수” |
| `count := 0` | 이번 `newCounter` 호출에서 사용할 지역 변수 |
| `return func() int { ... }` | `count`를 사용하는 익명 함수값 반환 |

`first := newCounter()`는 새 `count`를 만들고 함수값을 받습니다. 이때 안쪽의 `count++`가 실행되는 것은 아닙니다. 그 뒤 `first()`를 호출할 때마다 안쪽 본문을 실행하고, 같은 `count`를 증가시킵니다.

`second := newCounter()`는 `newCounter`를 **다시 호출**하므로 별도의 `count`를 갖습니다. 따라서 첫 번째 카운터가 이미 2여도 두 번째 카운터의 첫 결과는 1입니다.

| 실행 | 첫 번째 상태 | 두 번째 상태 | 반환값 |
| --- | --- | --- | --- |
| 두 카운터 생성 직후 | `0` | `0` | 각각 함수값 |
| 첫 `first()` | `1` | `0` | `1` |
| 두 번째 `first()` | `2` | `0` | `2` |
| 첫 `second()` | `2` | `1` | `1` |
| 세 번째 `first()` | `3` | `1` | `3` |

### 9.3 스코프와 변수의 수명은 다르다

`count`라는 이름을 `main`에서 직접 사용할 수는 없습니다. `count`는 `newCounter`의 지역 이름이기 때문입니다. 하지만 반환된 함수는 그 변수를 계속 사용할 수 있습니다. 클로저가 접근할 수 있는 동안 변수는 유지됩니다.

즉 **이름이 어디에서 보이는가**라는 스코프와 **변수가 언제까지 사용 가능한가**라는 수명은 다른 질문입니다. “함수 호출이 끝나면 지역 변수와 관련된 상태가 무조건 사라진다”고 외우면 클로저를 설명할 수 없습니다. 메모리가 구체적으로 어디에 배치되는지는 여기서 가정할 필요가 없습니다.

### 9.4 함수값을 다른 변수에 담으면 상태도 새로 생길까?

다음은 `closure` 예제의 `main`을 교체하는 발췌입니다. `newCounter` 선언은 유지합니다.

```go
first := newCounter()
same := first
fmt.Println(first())
fmt.Println(same())
fmt.Println(first())
```

예상 출력:

```text
1
2
3
```

`same := first`는 `newCounter()`를 다시 호출하지 않았습니다. 두 변수에 담긴 함수값은 같은 `count`를 사용하므로 호출 순서에 따라 1·2·3으로 이어집니다. 독립된 상태가 필요하면 `newCounter()`를 새로 호출해야 합니다.

클로저는 상태와 그 상태를 사용하는 동작을 함께 제공할 때 쓸 수 있습니다. 한편 호출할 때마다 결과가 달라질 수 있으므로 테스트에서는 **한 번의 반환값뿐 아니라 연속 호출과 독립된 생성 결과**를 확인합니다. 이 장의 카운터는 순서대로 호출하는 예제이며, 여러 고루틴의 공유 상태 처리는 동시성 장에서 다룹니다.

## 10. 직접 작성하기: 계산을 함수로 나누기

### 10.1 개인 파일 준비

자신의 터미널에 맞는 블록 하나를 실행합니다. `_practice/04-functions-scope/lab/main.go`가 이미 있으면 보존합니다. 앞 절의 발췌나 전체 코드를 시험할 때도 이 파일을 사용할 수 있습니다.

Windows PowerShell:

```powershell
$practiceDir = "_practice/04-functions-scope/lab"
New-Item -ItemType Directory -Force $practiceDir | Out-Null
$practiceMain = Join-Path $practiceDir "main.go"
if (Test-Path -LiteralPath $practiceMain) {
    Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
} else {
    Copy-Item examples/01-first-program/hello/main.go $practiceMain
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
practice_dir="_practice/04-functions-scope/lab"
mkdir -p "$practice_dir"
if [ -e "$practice_dir/main.go" ]; then
    printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
else
    cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
fi
```

기존 개인 코드를 남기려면 별도 폴더에 보관한 뒤 진행합니다.

### 10.2 입력과 처리의 책임 정하기

이번 실습은 가상의 주문 금액을 계산합니다. 규칙은 다음과 같습니다.

- 상품 단가는 1~100,000원, 수량은 1~10개입니다. 이 범위 밖이면 안내 한 줄을 출력하고 끝냅니다.
- 상품 금액은 단가 × 수량입니다.
- 상품 금액이 30,000원 이상이면 배송비는 0원, 그보다 작으면 3,000원입니다.
- 합계는 상품 금액 + 배송비입니다.

이는 학습용으로 정한 규칙입니다. `main`은 입력을 확인하고 출력하며, `itemTotal`과 `shippingFee`는 계산 결과만 반환하도록 나눕니다.

작업 파일: `_practice/04-functions-scope/lab/main.go`. 다음은 전체 코드입니다.

```go
package main

import "fmt"

func itemTotal(unitPrice, quantity int) int {
	return unitPrice * quantity
}

func shippingFee(subtotal int) int {
	if subtotal >= 30000 {
		return 0
	}
	return 3000
}

func main() {
	unitPrice := 12000
	quantity := 2

	if unitPrice < 1 || unitPrice > 100000 || quantity < 1 || quantity > 10 {
		fmt.Println("가격은 1~100000원, 수량은 1~10개로 입력하세요.")
		return
	}

	subtotal := itemTotal(unitPrice, quantity)
	fee := shippingFee(subtotal)
	fmt.Printf("상품 금액: %d원\n", subtotal)
	fmt.Printf("배송비: %d원\n", fee)
	fmt.Printf("합계: %d원\n", subtotal+fee)
}
```

포맷하고 실행합니다.

```sh
gofmt -w _practice/04-functions-scope/lab/main.go
go run ./_practice/04-functions-scope/lab
```

예상 출력:

```text
상품 금액: 24000원
배송비: 3000원
합계: 27000원
```

`itemTotal`이 반환한 값을 `subtotal`에 저장하고, 그 값을 다시 `shippingFee`의 인자로 전달합니다. 함수 사이에 숨은 공통 변수를 만들지 않아도 결과를 조합할 수 있습니다.

이 실습에서는 `main`이 허용 범위를 확인한 뒤 두 함수를 호출합니다. 따라서 계산 함수는 그 입력 규칙을 전제로 동작합니다. 이 함수들을 다른 프로그램에서 사용한다면 호출하는 쪽도 같은 전제조건을 지키거나 함수의 입력 검사 책임을 다시 설계해야 합니다.

### 10.3 경계 입력으로 책임 확인하기

단가와 수량만 바꾸며 실행합니다. 각 경우에 합계를 예상한 뒤 실제 출력과 비교하고 마지막에는 12,000원·2개로 되돌립니다.

| 단가 | 수량 | 상품 금액 | 배송비 | 합계 또는 결과 |
| --- | --- | --- | --- | --- |
| `12000` | `2` | `24000` | `3000` | `27000` |
| `12000` | `3` | `36000` | `0` | `36000` |
| `29999` | `1` | `29999` | `3000` | `32999` |
| `30000` | `1` | `30000` | `0` | `30000` |
| `30001` | `1` | `30001` | `0` | `30001` |
| `1` | `1` | `1` | `3000` | `3001` |
| `100000` | `10` | `1000000` | `0` | `1000000` |
| `0` | `2` | 계산하지 않음 | 계산하지 않음 | 범위 안내 한 줄 |
| `100001` | `1` | 계산하지 않음 | 계산하지 않음 | 범위 안내 한 줄 |
| `12000` | `0` | 계산하지 않음 | 계산하지 않음 | 범위 안내 한 줄 |
| `12000` | `11` | 계산하지 않음 | 계산하지 않음 | 범위 안내 한 줄 |

배송비 기준을 바꾸는 일은 `shippingFee`, 상품 금액 공식을 바꾸는 일은 `itemTotal`, 출력 문구를 바꾸는 일은 `main`에서 할 수 있습니다. **함수를 나누었을 때 어느 변경이 어디에 속하는지** 설명해 보세요.

`_practice/`는 저장소 공통의 `go test ./...`에서 자동으로 찾는 대상에 포함되지 않습니다. 개인 폴더를 직접 지정하여 실행합니다. 공통 규칙은 [연습문제 안내](../../exercises/README.md)에 있습니다.

## 11. 자주 하는 실수와 오류 읽기

### 11.1 반환하지 않는 경로 찾기

아래는 **의도적으로 컴파일에 실패하는 전체 코드**입니다. 개인 파일 `_practice/04-functions-scope/lab/main.go`에서만 시험하며, 앞서 작성한 코드는 보관한 뒤 바꿉니다.

```go
package main

import "fmt"

func resultLabel(score int) string {
	if score >= 60 {
		return "통과"
	}
}

func main() {
	fmt.Println(resultLabel(50))
}
```

실행:

```sh
go run ./_practice/04-functions-scope/lab
```

오류에서 `missing return`을 찾습니다. 함수는 `string`을 반환한다고 선언했지만 60보다 작은 점수의 경로에는 반환값이 없습니다. `if` 블록이 끝난 뒤, 함수가 끝나기 전에 다음 문장을 추가합니다.

```go
return "다시 연습"
```

위 한 줄은 **`resultLabel` 함수 안에 추가하는 발췌**입니다. 수정한 프로그램의 출력은 `다시 연습`이며, 점수를 60으로 바꾸면 `통과`입니다. 호출한 값이 우연히 60 이상인 경우만 사용하더라도 컴파일러는 함수 정의의 반환 경로를 검사합니다.

### 11.2 증상별 확인표

| 증상·코드 | 원인 | 확인과 수정 |
| --- | --- | --- |
| `not enough arguments` | 필요한 인자보다 적게 전달 | 함수 선언의 매개변수 개수 확인 |
| `too many arguments` | 고정 매개변수보다 많이 전달 | 호출할 함수와 인자 개수 확인 |
| 문자열을 `int` 인자로 전달 | 매개변수 타입과 불일치 | 필요한 값의 타입과 변환 목적 확인 |
| `assignment mismatch` | 받는 변수 개수와 반환값 개수 불일치 | 시그니처를 읽고 순서·개수 맞추기 |
| `missing return` | 반환값 있는 함수에서 끝까지 도달하는 경로에 반환 없음 | 모든 종료 경로의 결과를 확인 |
| 반환값 없는 함수의 결과를 변수에 대입 | 출력과 반환을 혼동 | 필요한 계산 결과를 반환하도록 함수 설계 |
| 함수에서 매개변수를 바꾸어도 호출자의 정수가 그대로임 | 별도 변수에 값이 복사됨 | 반환값을 받아 필요한 변수에 대입 |
| 바깥 변수의 값이 바뀌지 않음 | 안쪽에서 `:=`로 새 이름을 선언 | 갱신하려는 변수라면 `=` 사용 |
| `no new variables on left side of :=` | 같은 블록에서 새 이름 없이 짧은 선언 | 기존 변수 갱신에는 `=`; `_`는 새 이름이 아님 |
| 다른 함수의 지역 이름이 `undefined` | 그 이름의 스코프 밖에서 사용 | 매개변수로 전달하거나 반환값으로 받기 |
| 함수 안의 `func helper()`에서 문법 오류 | 이름 있는 함수 선언을 중첩 | 패키지 수준에 선언하거나 익명 함수값 사용 |
| `for number := range numbers`로 합이 잘못 나옴 | 슬라이스의 값 대신 위치를 합산 | `for _, number := range numbers` |
| `operation := add(3, 4)` 뒤 `operation(...)`에서 오류 | 함수값 대신 정수 결과를 저장 | 함수값을 보관하려면 `operation := add` |
| `nil`인 함수값을 호출해 실행 중 실패 | 호출할 함수가 설정되지 않음 | 실제 함수를 대입하거나 호출 전 계약·nil 확인 |
| 카운터를 다른 변수에 담았는데 숫자가 이어짐 | 같은 클로저 상태를 공유 | 독립 상태가 필요하면 생성 함수를 다시 호출 |

오류 문구 전체와 줄 번호는 Go 버전에 따라 달라질 수 있습니다. **호출 위치 → 함수 시그니처 → 반환 경로 → 이름의 선언 위치** 순으로 확인합니다. 컴파일 오류를 수정한 뒤에는 출력과 경계 입력도 확인하세요.

## 12. 선택 학습: 이름 있는 반환값과 슬라이스 전달

기본 문제를 풀 때 이 절의 문법을 반드시 사용할 필요는 없습니다. 함수의 입력·결과와 가변 인자를 이해한 뒤 읽어도 됩니다.

### 12.1 반환할 값에도 이름 붙이기

반환 타입과 함께 이름을 선언할 수 있습니다. 다음 전체 코드는 0 이상의 작은 분 단위를 시간과 나머지 분으로 나누며, 개인 실습 파일에서 따로 실행합니다.

```go
package main

import "fmt"

func splitMinutes(total int) (hours, minutes int) {
	hours = total / 60
	minutes = total % 60
	return hours, minutes
}

func main() {
	hours, minutes := splitMinutes(135)
	fmt.Printf("%d시간 %d분\n", hours, minutes)
}
```

저장·실행 위치는 `_practice/04-functions-scope/lab/main.go`와 저장소 루트입니다.

```sh
go run ./_practice/04-functions-scope/lab
```

예상 출력:

```text
2시간 15분
```

함수 선언의 `hours`, `minutes`는 **이름 있는 반환값(named results)**이며 함수에 진입할 때 각 타입의 영값으로 초기화되는 결과 변수입니다. 여기서는 둘 다 0에서 시작합니다. 함수 본문에서는 이미 선언된 이름이므로 `=`로 값을 바꿉니다.

이 함수에서는 `return hours, minutes` 대신 `return`만 적어 현재 결과 변수들을 반환할 수도 있습니다. 이를 **맨몸 반환(bare return, naked return)**이라고 부릅니다. 다만 코드가 길거나 여러 경로에서 값을 바꾸면 무엇을 반환하는지 추적하기 어려울 수 있어, 이 장의 기본 풀이에서는 반환할 값을 명시합니다.

이름을 붙였어도 함수 끝에서 자동 반환되는 것은 아닙니다. 끝까지 진행하는 경로에는 여전히 반환이 필요합니다. 안쪽 블록에서 결과 변수와 같은 이름을 선언해 가린 위치에서는 맨몸 반환이 컴파일 오류가 될 수도 있습니다. [return 문](https://go.dev/ref/spec#Return_statements)을 참고합니다.

### 12.2 이미 만든 슬라이스를 가변 인자로 전달하기

7절의 `variadic` 예제에서 `sum` 선언을 유지하고, `main`을 다음 발췌로 바꿉니다.

```go
values := []int{10, 20, 30}
fmt.Println(sum(values...))
fmt.Println(len(values))
```

예상 출력:

```text
60
3
```

`[]int{10, 20, 30}`은 정수 세 개를 담은 슬라이스를 만드는 표기입니다. 함수 선언에서는 `numbers ...int`, 호출에서는 `values...`처럼 점 세 개의 위치가 다릅니다.

| 위치 | 표기 | 뜻 |
| --- | --- | --- |
| 매개변수 선언 | `numbers ...int` | 여러 `int` 인자를 받음 |
| 함수 호출 | `sum(values...)` | `[]int`인 `values`를 가변 인자 부분으로 전달 |

이 호출에서 원소들을 별도의 저장 공간에 새로 복사하는 것은 아닙니다. 전달받은 슬라이스는 `values`와 같은 내부 데이터를 사용합니다. 위의 `sum`은 원소를 읽기만 하므로 값을 바꾸지 않습니다. 원소 변경이 어떻게 공유되는지와 복사가 필요한 경우는 05장에서 이어갑니다. [가변 인자 전달](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)과 [슬라이스 타입](https://go.dev/ref/spec#Slice_types)을 참고합니다.

`sum(values)`처럼 `...` 없이 쓰면 이 함수가 기대하는 `int` 인자 자리에 `[]int` 값을 전달하려 하므로 컴파일 오류입니다.

## 13. 연습문제와 검증

[04장 연습문제](../../exercises/04-functions-scope/README.md)에서 다음을 풉니다. 기본 문제 1~6은 이 장까지의 문법으로 해결할 수 있습니다. 함수의 이름·매개변수·반환 타입은 문제에서 지정한 계약에 맞춥니다.

| 문제 | 확인할 내용 |
| --- | --- |
| 1. 함수 호출과 이름 가리기의 출력 예측 | 값 복사, 반환값, 안쪽 블록의 새 변수 |
| 2. 점수 판정 함수 | 분기를 함수로 옮기고 결과·유효성 반환 |
| 3. 바뀌지 않는 반환값 고치기 | `:=`와 `=`의 차이, 실행 추적 |
| 4. 짝수의 합계 함수 | 가변 인자와 조건·반복 조합 |
| 5. 학습 계획을 계산하는 함수들 | 입력 검증과 결과 변환의 책임 분리 |
| 6. 증가량을 기억하는 함수 | 클로저의 상태, 독립된 생성과 연속 호출 |
| 선택 A. 같은 계산 두 번 적용 | 함수를 인자로 전달하고 결과를 다시 입력으로 사용 |

먼저 스스로 작성하고 힌트를 확인한 뒤 [04장 해설](../../solutions/04-functions-scope/README.md)과 비교합니다. 기본값만 맞추는 데서 끝내지 않고, 범위의 양 끝·잘못된 입력·인자 없음·연속 호출을 확인합니다.

공식 예제와 참고 구현의 검사는 저장소 루트에서 실행합니다.

```sh
go test -count=1 ./examples/04-functions-scope/... ./solutions/04-functions-scope/...
```

이번 장에서는 함수를 직접 호출할 수 있으므로 기본 출력뿐 아니라 **함수의 반환값과 경계 입력**도 테스트합니다. 예를 들어 나눗셈의 성공 여부, 유효한 0분 계획, 인자가 없는 합계, 별도로 만든 카운터의 상태를 확인합니다.

테스트 파일의 표·슬라이스·구조체와 `testing` 사용법은 이후 05·08·12장에서 배웁니다. 지금은 검사 명령을 실행하고 어떤 입력과 결과를 확인하는지 읽으면 됩니다. `_practice/`의 개인 코드를 대신 검사하는 것은 아니므로, 개인 답안은 문제의 명령과 입력 표로 따로 실행합니다.

## 14. 요약과 자기 점검

- 매개변수는 함수를 선언할 때 받을 값의 이름이고, 인자는 호출할 때 전달하는 실제 값이나 식입니다.
- 함수 선언은 처리의 정의이며, 함수 호출이 본문을 실행합니다.
- `return`은 결과를 돌려주고 현재 함수를 끝냅니다. 화면 출력은 별개의 동작입니다.
- Go는 인자 값을 복사해 전달합니다. 복사된 값이 다른 데이터와 무엇을 공유하는지는 자료형별로 구분합니다.
- 여러 반환값은 순서·개수·타입에 맞게 받습니다. 성공 여부는 호출한 쪽에서 확인합니다.
- 이름의 스코프는 코드의 선언 위치로 정해집니다. 안쪽의 `:=`는 바깥 이름을 가릴 수 있습니다.
- 가변 인자 `...int`는 함수 안에서 `[]int`이며, 슬라이스의 `range`에서 두 번째 값이 원소입니다.
- 함수값과 함수의 실행 결과는 다릅니다. 함수값은 변수에 담거나 인자·반환값으로 사용할 수 있습니다.
- 클로저는 바깥 변수를 사용할 수 있고 호출 사이에 상태를 유지할 수 있습니다. 함수값의 대입과 생성 함수의 새 호출을 구분합니다.

### 확인 질문

답을 보지 않고 설명해 보세요.

1. `func printGreeting(name string)`과 `printGreeting("Go")`에서 매개변수와 인자는 각각 무엇인가요?
2. `addMinutes(30, 15)`가 45를 반환해도 화면에 45가 보이지 않는 이유는 무엇인가요?
3. 함수 안에서 `current += extra`를 실행했는데 호출자의 `minutes`는 왜 대입 전까지 30인가요?
4. `divide(0, 5)`와 `divide(17, 0)`은 숫자 두 결과가 모두 0인데 무엇으로 구분하나요?
5. `announce(0)` 안에서 `return`을 실행하면 `main`의 다음 문장도 실행되지 않을까요?
6. 안쪽 `if`에서 `message := "통과"`를 쓸 때와 `message = "통과"`를 쓸 때 무엇이 다른가요?
7. `sum(10, 20, 30)`에서 `for number := range numbers`만 쓰면 왜 합계가 잘못될까요?
8. `operation := add`와 `result := add(3, 4)`에서 각 변수의 타입과 값은 어떻게 다른가요?
9. 클로저를 만든 뒤 바깥 변수 `bonus`를 바꾸면 함수에서 읽는 값도 달라질까요?
10. `second := first`와 `second := newCounter()`는 카운터 상태에 어떤 차이를 만들까요?
11. `newCounter`가 반환된 뒤에도 카운터를 계속 호출할 수 있는 이유는 무엇인가요?

<details>
<summary>답과 설명</summary>

1. `name`이 매개변수이고 `"Go"`가 인자입니다. 인자의 값을 함수 안에서 매개변수로 사용합니다.
2. 값을 반환하는 것과 출력하는 것은 다르기 때문입니다. 이 함수에는 출력문이 없으며, 반환값을 `fmt.Println` 등에 전달하면 표시할 수 있습니다.
3. 인자 값이 별도의 매개변수에 복사되었기 때문입니다. `minutes = updated`처럼 호출한 쪽에서 대입해야 그 정수 변수의 값이 바뀝니다.
4. `ok`입니다. 앞은 `true`인 정상 결과, 뒤는 `false`인 잘못된 입력입니다.
5. 실행됩니다. 그 `return`은 현재의 `announce` 호출만 끝내고 `main`으로 돌아옵니다.
6. `:=`는 안쪽 블록에 새 변수를 만들고 바깥 이름을 가립니다. `=`는 그 위치에서 보이는 기존 변수의 값을 바꿉니다.
7. 슬라이스 `range`의 첫 번째 값은 위치이므로 10·20·30 대신 0·1·2를 받습니다. 원소를 쓰려면 `_, number`로 두 번째 값을 받습니다.
8. `operation`은 `func(int, int) int` 타입의 함수값이며 아직 계산을 호출하지 않았습니다. `result`는 호출 결과 7을 담은 `int`입니다.
9. 달라집니다. 이 예제의 클로저는 생성 당시 값의 고정 복사본이 아니라 바깥의 같은 변수를 사용합니다.
10. 앞은 같은 상태를 사용하는 함수값을 대입하고, 뒤는 생성 함수를 새로 호출해 독립된 상태를 만듭니다.
11. 반환된 클로저가 `count`에 접근할 수 있는 동안 그 변수가 유지되기 때문입니다. `main`에서 이름을 직접 사용할 수 있는지는 별개의 스코프 문제입니다.

</details>

### 완료 체크리스트

- [ ] 함수 선언과 호출, 매개변수와 인자를 구분했다.
- [ ] 출력과 반환을 나누고 반환값을 필요한 변수에 대입했다.
- [ ] 모든 종료 경로의 결과와 다중 반환의 순서를 확인했다.
- [ ] 성공한 0과 실패한 결과를 구분했다.
- [ ] 블록의 범위를 따라 이름 가리기를 찾아 수정했다.
- [ ] 가변 인자의 값과 슬라이스의 위치를 구분했다.
- [ ] 함수값의 대입과 함수 호출을 구분했다.
- [ ] 클로저의 연속 호출·독립 생성·같은 함수값 대입을 확인했다.
- [ ] 기본 문제 1~6과 지정된 경계 입력을 확인했다.

## 참고 자료

- [Go 명세: 함수 선언](https://go.dev/ref/spec#Function_declarations) — 이름, 본문과 종료 경로
- [Go 명세: 함수 타입](https://go.dev/ref/spec#Function_types) — 매개변수, 결과, 가변 인자
- [Go 명세: 함수 호출](https://go.dev/ref/spec#Calls) — 인자 전달과 호출 규칙
- [Go FAQ: 값 전달](https://go.dev/doc/faq#pass_by_value) — 복사되는 값과 공유되는 데이터
- [Go 명세: return 문](https://go.dev/ref/spec#Return_statements) — 다중 반환과 이름 있는 결과
- [Go 명세: 선언과 스코프](https://go.dev/ref/spec#Declarations_and_scope) — 이름이 보이는 범위
- [Go 명세: 짧은 변수 선언](https://go.dev/ref/spec#Short_variable_declarations) — 재사용과 새 변수
- [Go 명세: 가변 인자 전달](https://go.dev/ref/spec#Passing_arguments_to_..._parameters) — 함수 안의 슬라이스와 펼쳐 전달하기
- [Go 명세: 함수 리터럴](https://go.dev/ref/spec#Function_literals) — 익명 함수와 클로저
- [Go 명세: range 절](https://go.dev/ref/spec#For_range) — 슬라이스의 위치와 값

확인일: 2026-11-09. 예제와 의도적인 실패 사례는 Go 1.27.3에서 확인했습니다.

## 다음 학습

다음은 **05. 배열과 슬라이스**이며 아직 집필 예정입니다. 가변 인자에서 만난 `[]int`를 출발점으로 여러 값을 저장하고, 길이와 용량을 확인하고, 원소를 추가하고 복사하는 방법을 배웁니다. 특히 함수에 전달한 슬라이스와 원래 값이 저장 공간을 공유하는 경우를 직접 확인합니다. 집필 상태는 [전체 목차](../README.md)와 [로드맵](../../docs/ROADMAP.md)에서 확인합니다.

| 이전 | 목차 | 실습과 해설 |
| --- | --- | --- |
| [03. 조건문과 반복문](../03-control-flow/README.md) | [전체 목차](../README.md) | [문제](../../exercises/04-functions-scope/README.md) · [해설](../../solutions/04-functions-scope/README.md) |
