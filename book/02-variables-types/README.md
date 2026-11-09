# 02. 변수·상수·자료형·연산자

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 입문 | [01. 첫 Go 프로그램](../01-first-program/README.md) | Go, 편집기, 터미널, 이 저장소 |

01장에서는 문자열과 숫자를 출력문에 직접 적었습니다. 학습한 장의 수나 공부한 시간이 바뀔 때마다 여러 출력문을 고쳐야 한다면 실수하기 쉽습니다. 이번 장에서는 **값에 이름을 붙이고, 그 값으로 계산하고, 계산 결과의 자료형을 확인하는 방법**을 배웁니다.

## 학습 목표

이 장을 마치면 다음을 할 수 있습니다.

- `var`, `:=`, `=`를 구분하여 변수를 선언하고 값을 변경한다.
- 기본 자료형과 영값을 설명하고 출력에서 값과 타입을 확인한다.
- 변하지 않는 기준값을 상수로 표현하고 변수와의 차이를 설명한다.
- 산술·비교·논리 연산을 조합하고 필요한 위치에서 명시적으로 형 변환한다.
- 정수 나눗셈, 상수 변환, 범위 초과, 부동소수점 계산에서 생기는 실수를 진단한다.

## 사전 지식과 준비

`package main`, `import "fmt"`, `func main()`과 `fmt.Printf`의 기본 사용을 알고 있다고 가정합니다. 실행과 파일 저장이 낯설다면 [01장](../01-first-program/README.md)을 먼저 마칩니다. 조건문·반복문·사용자 정의 함수는 기본 실습에 필요하지 않습니다.

모든 명령은 **저장소 루트 `GoLang_all_in_one/`**에서 실행합니다. Go 명령은 Windows PowerShell과 macOS·Linux·WSL에서 같습니다. 개인 파일 준비 명령은 운영체제별로 구분합니다. 루트의 [`go.mod`](../../go.mod)를 함께 사용하므로 하위 폴더에 새 모듈을 만들지 않습니다.

| 실행 예제 | 학습할 내용 | 실행 명령 |
| --- | --- | --- |
| [variables](../../examples/02-variables-types/variables/main.go) | 선언과 값 변경 | `go run ./examples/02-variables-types/variables` |
| [zero-values](../../examples/02-variables-types/zero-values/main.go) | 영값과 타입 출력 | `go run ./examples/02-variables-types/zero-values` |
| [constants](../../examples/02-variables-types/constants/main.go) | 상수를 사용한 시간 계산 | `go run ./examples/02-variables-types/constants` |
| [conversions](../../examples/02-variables-types/conversions/main.go) | 변환 위치와 소수 부분 | `go run ./examples/02-variables-types/conversions` |
| [operators](../../examples/02-variables-types/operators/main.go) | 산술·비교·논리·우선순위 | `go run ./examples/02-variables-types/operators` |

본문에서 **발췌**라고 표시한 Go 블록은 `main` 함수 안의 일부입니다. 따로 실행하려면 `package main`, 필요한 import와 `func main()`을 포함한 파일 안에 넣습니다. 9절에서는 이런 코드를 시험할 개인 파일을 준비합니다.

## 1. 값에 이름을 붙여 사용하기

**변수(variable)**는 이름을 통해 값을 보관하고 다시 읽거나 바꾸는 공간입니다. **자료형 또는 타입(type)**은 어떤 종류의 값을 담을 수 있고 어떤 연산이 가능한지를 정합니다. 아래에서는 언어 이름을 문자열로, 완료한 장의 수를 정수로 저장합니다.

저장 경로: [`examples/02-variables-types/variables/main.go`](../../examples/02-variables-types/variables/main.go).

```go
package main

import "fmt"

func main() {
	var language string = "Go"
	completed := 1

	fmt.Printf("변경 전: %s %d장\n", language, completed)
	completed = completed + 1
	fmt.Printf("변경 후: %s %d장\n", language, completed)
}
```

실행:

```sh
go run ./examples/02-variables-types/variables
```

예상 출력:

```text
변경 전: Go 1장
변경 후: Go 2장
```

| 코드 | 실행할 때 하는 일 |
| --- | --- |
| `var language string = "Go"` | `string` 타입의 변수 `language`를 만들고 `"Go"`로 초기화 |
| `completed := 1` | 초기값에서 타입을 정해 `int` 변수 `completed` 선언 |
| 첫 번째 `Printf` | 현재 변수의 값인 `"Go"`와 `1` 출력 |
| `completed = completed + 1` | 기존 값 `1`에 `1`을 더한 결과 `2`를 다시 저장 |
| 두 번째 `Printf` | 변경된 값 `2` 출력 |

변수를 처음 만들고 값을 정하는 것을 **선언(declaration)**과 **초기화(initialization)**라고 합니다. 이미 있는 변수에 값을 저장하는 것은 **대입(assignment)**입니다. 변수의 이름은 그대로지만 값은 바뀔 수 있습니다. [변수 선언](https://go.dev/ref/spec#Variable_declarations)과 [대입문](https://go.dev/ref/spec#Assignment_statements)이 이 동작을 정의합니다.

`completed = completed + 1`은 수학의 등식이 아닙니다. 오른쪽에서 현재 값을 읽어 계산한 뒤 왼쪽 변수에 저장하는 명령입니다. 두 값이 같은지를 물을 때는 뒤에서 배울 `==`를 사용합니다.

## 2. 선언, 타입 추론, 대입 구분하기

### 2.1 `var`와 `:=`의 선택

아래 표의 각 행은 **서로 다른 선언 방법**입니다. 같은 블록에 모두 붙여 넣어 실행하는 코드가 아닙니다.

| 선언 | 초기값 | 타입 | 사용할 상황 |
| --- | --- | --- | --- |
| `var count int = 3` | `3` | 명시한 `int` | 타입과 초기값을 모두 드러낼 때 |
| `var count = 3` | `3` | 추론한 `int` | 값에서 타입을 정할 때 |
| `count := 3` | `3` | 추론한 `int` | 함수 안에서 짧게 선언할 때 |
| `var count int` | 영값 `0` | 명시한 `int` | 초기값을 따로 주지 않을 때 |

**타입 추론(type inference)**은 초기값을 보고 변수의 타입을 결정하는 것입니다. 타입을 적지 않았다고 아무 타입이나 나중에 담을 수 있는 것은 아닙니다. Go 변수의 타입은 선언할 때 정해집니다.

다음은 `main` 안의 발췌 코드입니다.

```go
count := 3
count = 5
fmt.Println(count)
```

결과는 `5`입니다. 마지막 대입을 `count = "다섯"`으로 바꾸면 **의도적으로 컴파일에 실패**합니다. `count`는 `int`인데 문자열을 저장하려 했기 때문입니다. 같은 이름에 다른 타입의 값을 번갈아 담는 방식으로 사용하지 않습니다.

`:=`는 **짧은 변수 선언(short variable declaration)**이며 함수 안에서 사용합니다. 함수 바깥의 패키지 수준에서는 `var`나 `const`를 사용합니다. 이번 장의 데이터는 동작을 따라가기 쉽도록 `main` 안에 둡니다. [짧은 변수 선언 규칙](https://go.dev/ref/spec#Short_variable_declarations)을 참고합니다.

### 2.2 `:=`는 새 선언, `=`는 기존 값 변경

다음 발췌는 **의도적으로 실패하는 예**입니다.

```go
count := 1
count := 2
fmt.Println(count)
```

같은 블록에서 새 변수 없이 `:=`를 반복했기 때문입니다. 대표 메시지는 `no new variables on left side of :=`입니다. 두 번째 줄을 `count = 2`로 고치면 됩니다.

여러 변수를 함께 선언할 때는 다른 경우가 있습니다. 다음 발췌는 정상입니다.

```go
count := 1
count, step := 2, 3
fmt.Println(count, step)
```

출력은 `2 3`입니다. **같은 블록**에서 기존 `count`의 타입을 유지하면서 새 이름 `step`을 함께 선언했으므로 허용됩니다. 기존 변수만으로 이루어진 선언과 구분하세요. 바깥 블록에 있는 이름을 안쪽 블록에서 다시 `:=`로 선언하면 별개의 변수가 될 수 있습니다. 이름을 사용할 수 있는 범위인 **스코프(scope)**와 이 가림 현상은 04장에서 자세히 다룹니다.

### 2.3 값을 복사하는 것과 계산식을 연결하는 것은 다르다

다음은 `main` 안의 발췌입니다.

```go
original := 3
copied := original
original = 8
fmt.Println(original, copied)
```

출력은 `8 3`입니다. `copied`에는 대입 당시의 정수 값 `3`이 저장됩니다. 나중에 `original`이 바뀐다고 `copied`가 자동으로 다시 계산되지 않습니다. 여기서는 정수 값을 다루며, 슬라이스처럼 저장 공간을 공유할 수 있는 자료형은 해당 장에서 별도로 배웁니다.

여러 값을 함께 대입할 수도 있습니다. 다음 발췌는 두 값을 교환합니다.

```go
left, right := 10, 20
left, right = right, left
fmt.Println(left, right)
```

출력은 `20 10`입니다. 오른쪽 값들을 구한 뒤 왼쪽에 대입하기 때문에 중간에 한쪽 값이 사라지지 않습니다. [대입의 평가 순서](https://go.dev/ref/spec#Assignment_statements)를 참고합니다.

### 2.4 이름과 사용 여부

이 교재에서는 `completedChapters`, `totalSeconds`처럼 역할과 단위가 드러나는 영어 이름을 사용합니다. `var`, `func` 같은 키워드는 변수 이름으로 사용할 수 없고, 이름의 첫 글자로 숫자를 사용할 수도 없습니다. `count`와 `Count`는 서로 다른 이름입니다. 근거는 [식별자](https://go.dev/ref/spec#Identifiers)와 [키워드](https://go.dev/ref/spec#Keywords) 규칙입니다.

Go 도구로 함수 안에 선언했지만 사용하지 않는 지역 변수를 컴파일하면 `declared and not used` 오류가 납니다. 필요한 값이라면 계산이나 출력에 사용하고, 필요 없다면 선언을 제거합니다. 단순히 오류를 숨기기보다 그 변수가 왜 필요한지 확인합니다.

## 3. 초기값을 생략하면 영값에서 시작한다

`var count int`처럼 초기값을 적지 않아도 값이 미정인 채로 남지 않습니다. 타입마다 정해진 **영값(zero value)**으로 초기화됩니다.

저장 경로: [`examples/02-variables-types/zero-values/main.go`](../../examples/02-variables-types/zero-values/main.go).

```go
package main

import "fmt"

func main() {
	var count int
	var rate float64
	var ready bool
	var title string

	fmt.Printf("count=%d (%T)\n", count, count)
	fmt.Printf("rate=%.1f (%T)\n", rate, rate)
	fmt.Printf("ready=%t (%T)\n", ready, ready)
	fmt.Printf("title=%q (%T)\n", title, title)
}
```

실행:

```sh
go run ./examples/02-variables-types/zero-values
```

예상 출력:

```text
count=0 (int)
rate=0.0 (float64)
ready=false (bool)
title="" (string)
```

| 타입 | 영값 | 의미 |
| --- | --- | --- |
| 정수 타입 | `0` | 정수 영 |
| 부동소수점 타입 | `0` | 실수 영; 예제에서는 `%.1f`로 `0.0` 표시 |
| `bool` | `false` | 거짓 |
| `string` | `""` | 문자가 없는 빈 문자열 |

빈 문자열 `""`과 공백 한 칸이 든 `" "`는 다릅니다. `%q`를 쓰면 따옴표가 함께 보여 차이를 확인하기 쉽습니다. 영값은 [프로그램 초기화 규칙](https://go.dev/ref/spec#The_zero_value)에 정의되어 있습니다.

이번 장에서 추가로 사용하는 출력 서식도 정리해 둡니다.

| 서식 | 용도 | 사용 예 |
| --- | --- | --- |
| `%d` | 정수를 십진수로 표시 | `3` |
| `%s` | 문자열 내용 표시 | `Go` |
| `%t` | 불리언 값 표시 | `true`, `false` |
| `%q` | 문자열을 따옴표와 이스케이프를 포함해 표시 | `""`, `"Go"` |
| `%T` | 전달한 값의 타입 표시 | `int`, `float64` |
| `%.1f`, `%.2f` | 부동소수점 값을 소수점 아래 1자리·2자리로 표시 | `3.5`, `61.25` |

`%.2f`는 **출력할 때의 표시 자릿수**를 정합니다. 변수에 저장된 값을 반올림한 새 값으로 바꾸거나 계산 정밀도를 높이지 않습니다. [fmt의 서식 규칙](https://pkg.go.dev/fmt#hdr-Printing)을 참고합니다.

## 4. 어떤 자료형을 선택할까?

### 4.1 이번 장에서 사용하는 기본 타입

| 종류 | 타입 | 사용하는 값과 주의점 |
| --- | --- | --- |
| 참·거짓 | `bool` | 준비 여부, 비교 결과; 값은 `true` 또는 `false` |
| 문자열 | `string` | 이름, 메시지, 문장; `"25"`도 문자열 |
| 일반 정수 | `int` | 개수, 장 번호처럼 정수로 계산하는 값 |
| 크기가 정해진 정수 | `int8`, `int16`, `int32`, `int64` | 이름의 숫자는 비트 수; 표현 가능한 범위가 다름 |
| 부호 없는 정수 | `uint`, `uint8`, `uint16`, `uint32`, `uint64` | 음수를 표현하지 않음; 범위 초과가 자동으로 안전해지는 것은 아님 |
| 부동소수점 | `float32`, `float64` | 소수 부분이 있는 수; 모든 십진수를 정확히 저장할 수 있지는 않음 |
| 바이트 별칭 | `byte` | `uint8`과 같은 타입 |
| 룬 별칭 | `rune` | `int32`와 같은 타입; 유니코드 코드 포인트를 다룰 때 사용 |

**비트(bit)**는 이진수 한 자리를 뜻합니다. 예를 들어 `uint8`은 `0`부터 `255`, `int8`은 `-128`부터 `127`을 표현합니다. `int`와 `uint`는 구현에 따라 32비트 또는 64비트입니다. 지금 사용하는 컴퓨터에서 64비트라고 해서 모든 환경에서 같다고 가정하지 않습니다. 같은 크기인 환경에서도 `int`와 `int64`는 서로 다른 타입입니다. [숫자 타입 명세](https://go.dev/ref/spec#Numeric_types)를 참고합니다.

일반적인 개수는 우선 `int`, 소수 계산은 우선 `float64`로 시작하면 이 장의 문제를 풀 수 있습니다. 파일 형식이나 API가 특정 크기를 요구할 때는 그 요구에 맞는 타입을 사용합니다. 단순히 작은 양수라는 이유만으로 `uint8`을 선택하지는 않습니다.

`byte`와 `rune`은 새로운 별개 타입이 아닌 **별칭(alias)**입니다. 이 타입의 변수를 `%T`로 출력하면 각각 `uint8`, `int32`가 표시됩니다. 문자열의 바이트와 문자 수가 다른 이유는 07장에서 다룹니다. 복소수 타입 `complex64`·`complex128`, 특수 용도의 `uintptr` 등은 이번 장의 기본 실습에 사용하지 않습니다.

### 4.2 초기값에서 추론하는 타입

다음은 `main` 안의 발췌입니다.

```go
count := 3
ratio := 3.0
name := "Go"
ready := true
letter := 'G'
fmt.Printf("%T %T %T %T %T\n", count, ratio, name, ready, letter)
```

예상 출력:

```text
int float64 string bool int32
```

`3`과 `3.0`은 이 선언에서 서로 다른 기본 타입을 선택하게 합니다. `'G'`는 01장에서 본 룬 리터럴이며, 여기서는 `rune`의 별칭 대상인 `int32`가 표시됩니다. 기본 타입의 규칙은 [상수](https://go.dev/ref/spec#Constants)와 [변수 선언](https://go.dev/ref/spec#Variable_declarations)에 있습니다.

이미 타입이 있는 값에서 추론하면 그 타입을 이어받습니다. 예를 들어 `var small int32 = 3` 다음에 `copyOfSmall := small`을 쓰면 새 변수도 `int32`입니다. `:=`가 언제나 `int`를 만든다는 뜻은 아닙니다.

`"3"`은 숫자를 적은 문자열이므로 `3 + "3"`처럼 정수와 바로 더할 수 없습니다. 텍스트를 숫자로 해석하는 작업은 단순한 숫자 형 변환과 구분하며, 문자열·입력 처리에서 배웁니다.

## 5. 변하지 않는 기준값은 상수로 표현하기

**상수(constant)**는 실행 중 바꾸지 않는 값을 이름으로 표현합니다. `const minutesPerHour = 60`처럼 선언합니다. 시간 단위처럼 계산의 의미를 설명하는 이름을 붙이면 숫자를 여러 곳에 반복하는 것보다 읽기 쉽습니다.

다음 예제에서 `/`는 정수 나눗셈의 몫을, `%`는 나머지를 구합니다. `135 / 60`은 `2`, `135 % 60`은 `15`입니다. 연산자의 자세한 동작은 7절에서 정리합니다.

저장 경로: [`examples/02-variables-types/constants/main.go`](../../examples/02-variables-types/constants/main.go).

```go
package main

import "fmt"

func main() {
	const minutesPerHour = 60
	const course string = "Go"
	studyMinutes := 135
	hours := studyMinutes / minutesPerHour
	minutes := studyMinutes % minutesPerHour

	fmt.Printf("과정: %s\n", course)
	fmt.Printf("학습 시간: %d시간 %d분\n", hours, minutes)
}
```

실행:

```sh
go run ./examples/02-variables-types/constants
```

예상 출력:

```text
과정: Go
학습 시간: 2시간 15분
```

`studyMinutes`는 실습 중 바꿔 볼 입력값입니다. 반면 한 시간에 60분이라는 기준은 이 프로그램에서 변하지 않으므로 상수로 표현했습니다. `minutesPerHour = 90`처럼 다시 대입하면 컴파일 오류가 납니다.

상수의 값은 컴파일할 때 정할 수 있어야 합니다. 불리언·숫자·문자열 상수를 만들 수 있지만, 실행 중 입력받은 값이나 일반 함수 호출의 결과를 그대로 `const`에 넣을 수는 없습니다. `const`에는 `:=`가 아닌 `=`를 사용합니다. [상수 선언](https://go.dev/ref/spec#Constant_declarations)을 참고합니다.

### 타입을 지정한 상수와 지정하지 않은 숫자 상수

`const course string = "Go"`는 타입을 명시한 상수입니다. `const limit = 10`처럼 타입을 쓰지 않고 숫자 리터럴을 그대로 사용하면 **타입이 지정되지 않은 상수(untyped constant)**가 됩니다. 이러한 상수는 사용할 위치의 타입과 표현 범위에 맞춰 사용할 수 있습니다.

다음은 `main` 안의 발췌입니다.

```go
const limit = 10
var count int = limit
var ratio float64 = limit
fmt.Printf("%d %.1f\n", count, ratio)
```

출력은 `10 10.0`입니다. 이 동작은 서로 다른 타입의 **변수**를 자동으로 바꾸어 계산한다는 뜻이 아닙니다. 또 모든 상수를 아무 타입에나 넣을 수 있는 것도 아닙니다. 예를 들어 `const limit = 300`을 `uint8` 변수에 대입하면 그 타입의 범위를 벗어나 컴파일에 실패합니다.

`value := limit`처럼 별도의 타입 없이 변수로 받으면 이 정수 상수의 기본 타입인 `int`가 선택됩니다. 상수 자체의 타입을 확인하려고 `%T`에 전달해도, 호출에 필요한 타입으로 변환된 결과를 보게 됩니다.

타입을 선언에 쓰지 않았더라도 `const limit = int64(10)`처럼 우변에서 타입을 정하면 타입 있는 상수입니다. 지금은 **값이 고정된 상수와 타입이 정해진 변수를 구분한다**는 점에 집중하세요. 더 자세한 설명은 [Go 공식 블로그: Constants](https://go.dev/blog/constants)에 있습니다.

## 6. 다른 숫자 타입은 명시적으로 변환하기

서로 다른 숫자 타입의 변수는 같은 숫자를 담고 있어도 바로 섞어 계산할 수 없습니다. `int`와 `int64`의 합을 구하려면 한쪽을 다른 쪽 타입으로 바꾸어야 합니다.

**형 변환(type conversion)**은 `int64(value)`, `float64(value)`처럼 `타입(값)` 형태로 씁니다. 변환한 새 값을 식에서 사용할 뿐 원래 변수의 타입이 바뀌지는 않습니다. [변환 명세](https://go.dev/ref/spec#Conversions)를 참고합니다.

저장 경로: [`examples/02-variables-types/conversions/main.go`](../../examples/02-variables-types/conversions/main.go).

```go
package main

import "fmt"

func main() {
	total := 7
	count := 2
	var offset int64 = 3

	fmt.Printf("형을 맞춘 합: %d\n", int64(total)+offset)
	fmt.Printf("정수 나눗셈: %d\n", total/count)
	fmt.Printf("나눈 뒤 변환: %.1f\n", float64(total/count))
	fmt.Printf("나누기 전 변환: %.1f\n", float64(total)/float64(count))

	distance := 3.9
	fmt.Printf("소수 부분 제거: %d\n", int(distance))
}
```

실행:

```sh
go run ./examples/02-variables-types/conversions
```

예상 출력:

```text
형을 맞춘 합: 10
정수 나눗셈: 3
나눈 뒤 변환: 3.0
나누기 전 변환: 3.5
소수 부분 제거: 3
```

### 6.1 계산하기 전에 타입을 맞춘다

`total`은 `int`, `offset`은 `int64`입니다. `int64(total)+offset`은 둘 다 `int64`인 값을 더합니다. 변환을 빼고 `total+offset`으로 작성하면 **의도적으로 컴파일에 실패**하며 `mismatched types int and int64`라는 부분을 확인할 수 있습니다.

타입을 맞추기 위해 무조건 큰 값도 작은 타입으로 바꾸는 것은 좋지 않습니다. 변환 대상이 원래 값을 표현할 수 있는지 확인해야 합니다. 범위를 넘어가는 변환은 8절에서 다시 봅니다.

### 6.2 나눈 뒤에 바꾸면 사라진 소수 부분은 돌아오지 않는다

| 식 | 계산 순서 | 결과 |
| --- | --- | --- |
| `total / count` | 정수 `7`을 정수 `2`로 나눔 | 정수 `3` |
| `float64(total / count)` | 정수 나눗셈 후 그 결과 `3`을 변환 | 실수 `3.0` |
| `float64(total) / float64(count)` | 두 값을 먼저 실수로 변환한 뒤 나눔 | 실수 `3.5` |

평균처럼 소수 부분이 필요하다면 **나누기 전에 변환**합니다. 계산이 끝난 다음 `%.2f`로 출력하는 것만으로 사라진 소수 부분을 복구할 수도 없습니다.

### 6.3 실수 변수에서 정수로 바꾸면 소수 부분을 버린다

`distance := 3.9`는 `float64` 변수입니다. `int(distance)`는 `3`을 만들며 반올림하지 않습니다. `distance`의 값과 타입은 그대로입니다. 음수 변수의 값이 `-3.9`라면 변환 결과는 `-3`입니다. 이를 **0 방향으로 절단한다**고 합니다. 여기의 값들은 변환 대상 정수 타입의 범위 안에 있습니다. [숫자 타입 사이의 변환](https://go.dev/ref/spec#Conversions)을 참고합니다.

다음 발췌는 **의도적으로 실패하는 예**입니다.

```go
whole := int(3.9)
fmt.Println(whole)
```

`3.9`는 상수이며 정수 타입으로 그대로 표현할 수 없으므로 컴파일 오류가 납니다. `int(float64(3.9))`도 여전히 상수 변환이므로 해결되지 않습니다. 반면 `int(3.0)`은 값이 정확히 정수 `3`으로 표현되므로 허용됩니다. **상수의 표현 가능성 검사와 실수 변수의 소수 부분 제거는 다른 규칙**입니다. [상수의 표현 가능성](https://go.dev/ref/spec#Representability)을 참고합니다.

## 7. 연산자로 계산하고 조건 만들기

**연산자(operator)**는 더하기, 비교처럼 값에 적용할 동작을 표시합니다. `a + b`에서 `+`가 연산자이고 `a`, `b`가 **피연산자(operand)**입니다. 값을 계산하는 `a + b`, `a > b` 같은 코드를 **식(expression)**이라고 합니다.

### 7.1 산술 연산

| 연산자 | 의미 | 정수 `a = 7`, `b = 2`일 때 |
| --- | --- | --- |
| `+` | 덧셈 | `a + b` → `9` |
| `-` | 뺄셈 | `a - b` → `5` |
| `*` | 곱셈 | `a * b` → `14` |
| `/` | 나눗셈 | `a / b` → `3` |
| `%` | 정수 나눗셈의 나머지 | `a % b` → `1` |

정수 나눗셈은 소수 부분을 0 방향으로 버립니다. 따라서 `-7 / 2`는 `-3`, `-7 % 2`는 `-1`입니다. 나머지가 0이 아니라면 그 부호는 나누어지는 수의 부호와 같습니다. `%`는 부동소수점 값에 사용할 수 없습니다. [정수 연산 규칙](https://go.dev/ref/spec#Integer_operators)을 참고합니다.

정수를 0으로 나누면 안 됩니다. `7 / 0`처럼 상수식이면 컴파일 오류이고, 실행 중 값이 0인 정수 변수로 나누면 프로그램 실행이 실패합니다. 이 장의 평균 문제는 나누는 수가 양수라는 조건을 두며, 조건에 따라 다른 처리를 하는 방법은 03장에서 배웁니다.

문자열에는 `+`로 이어 붙이는 연산을 사용할 수 있습니다. `"Go" + " 학습"`의 결과는 `"Go 학습"`입니다. 두 문자열을 곱하거나 문자열에서 정수를 빼는 식은 사용할 수 없습니다. [문자열 연결 규칙](https://go.dev/ref/spec#String_concatenation)을 참고합니다.

### 7.2 비교 결과는 `bool`

| 연산자 | 질문 |
| --- | --- |
| `==` | 같은가? |
| `!=` | 다른가? |
| `<`, `<=` | 작은가? 작거나 같은가? |
| `>`, `>=` | 큰가? 크거나 같은가? |

비교 결과는 `true` 또는 `false`입니다. `7 > 2`는 `true`, `7 == 2`는 `false`입니다. 변수에 저장할 수도 있습니다. `withinBudget := total <= budget`은 예산 안에 들어오는지를 나타내는 `bool` 변수를 만듭니다.

`=`는 값을 저장하는 대입이고 `==`는 같은지를 묻는 비교입니다. 숫자 `0`이나 빈 문자열을 자동으로 `false`처럼 사용하지 않습니다. 참·거짓이 필요하면 `count > 0`, `title != ""`처럼 비교식을 작성합니다. [비교 연산 규칙](https://go.dev/ref/spec#Comparison_operators)을 참고합니다.

### 7.3 논리 연산과 단락 평가

| 연산자 | 의미 | 예 |
| --- | --- | --- |
| `&&` | 둘 다 참인가? | `true && false` → `false` |
| `\|\|` | 적어도 하나가 참인가? | `true \|\| false` → `true` |
| `!` | 참·거짓을 뒤집기 | `!true` → `false` |

논리 연산은 `bool` 값에 적용합니다. `a > 0 && b > 0`은 두 숫자가 모두 양수인지 확인합니다. 범위를 검사할 때도 `0 <= count && count <= 10`처럼 비교 두 개를 연결합니다. 수학 표기처럼 `0 <= count <= 10`이라고 이어 쓰지는 않습니다.

`&&`는 왼쪽이 거짓이면 오른쪽을 계산하지 않고, `||`는 왼쪽이 참이면 오른쪽을 계산하지 않습니다. 이를 **단락 평가(short-circuit evaluation)**라고 합니다. 다음은 `main` 안의 발췌입니다.

```go
count := 0
canDivide := count != 0 && 10/count > 1
fmt.Println(canDivide)
```

출력은 `false`입니다. `count != 0`이 거짓이어서 오른쪽 정수 나눗셈을 실행하지 않습니다. 다만 **실행을 생략하는 것과 컴파일 검사를 생략하는 것은 다릅니다.** `false && 1/0 > 0`처럼 상수 나눗셈 오류가 있는 식은 컴파일에 실패합니다. [논리 연산 규칙](https://go.dev/ref/spec#Logical_operators)에 이 평가 순서가 정의되어 있습니다.

### 7.4 우선순위와 괄호

이번 절에서 사용하는 연산자의 우선순위는 다음과 같습니다. 위쪽이 먼저 적용됩니다.

| 순서 | 연산자 |
| --- | --- |
| 1 | 단항 `+`, 단항 `-`, `!` |
| 2 | `*`, `/`, `%` |
| 3 | `+`, `-` |
| 4 | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| 5 | `&&` |
| 6 | `\|\|` |

`2 + 3*4`는 `14`, `(2 + 3)*4`는 `20`입니다. **괄호로 계산 의도를 드러내면** 우선순위를 다시 확인하는 부담을 줄일 수 있습니다. 같은 우선순위의 이항 연산은 왼쪽부터 묶입니다. 예를 들어 정수식 `10 / 3 * 3`은 `(10 / 3) * 3`이 되어 `9`입니다. 비트 연산을 포함한 전체 우선순위는 [연산자 명세](https://go.dev/ref/spec#Operator_precedence)에 있습니다.

### 7.5 대입을 줄여 쓰는 방법

다음은 `main` 안의 발췌입니다.

```go
count := 1
count += 2
count *= 3
count++
count--
fmt.Println(count)
```

출력은 `9`입니다. 이 변수는 차례로 `1`, `3`, `9`, `10`, `9`가 됩니다. `count += 2`는 `count = count + 2`, `count *= 3`은 `count = count * 3`과 같은 동작입니다. `-=`, `/=`, `%=`도 사용할 수 있습니다.

`count++`, `count--`는 각각 1을 더하고 빼는 **문장(statement)**입니다. 값을 반환하는 식이 아니므로 `fmt.Println(count++)`나 `next := count++`는 쓸 수 없습니다. `++count` 형태도 지원하지 않습니다. 값 변경과 출력은 별도 문장으로 적습니다. [증가·감소문](https://go.dev/ref/spec#IncDec_statements)을 참고합니다.

### 7.6 연산 결과를 한 번에 확인하기

저장 경로: [`examples/02-variables-types/operators/main.go`](../../examples/02-variables-types/operators/main.go).

```go
package main

import "fmt"

func main() {
	a, b := 7, 2

	fmt.Println("합:", a+b)
	fmt.Println("차:", a-b)
	fmt.Println("곱:", a*b)
	fmt.Println("몫:", a/b)
	fmt.Println("나머지:", a%b)
	fmt.Println("비교:", a > b, a == b)
	fmt.Println("조건:", a > 0 && b > 0, a == 0 || b == 0, !(a == b))
	fmt.Println("우선순위:", 2+3*4, (2+3)*4)
}
```

실행:

```sh
go run ./examples/02-variables-types/operators
```

예상 출력:

```text
합: 9
차: 5
곱: 14
몫: 3
나머지: 1
비교: true false
조건: true false true
우선순위: 14 20
```

`조건:` 줄의 세 결과는 서로 다른 논리식의 결과입니다. 두 수는 모두 양수이므로 첫째는 참, 둘 다 0이 아니므로 둘째는 거짓, 두 수가 같지 않으므로 셋째는 참입니다.

## 8. 값의 범위와 계산 정밀도 확인하기

### 8.1 정수의 범위 초과가 항상 오류가 되는 것은 아니다

다음은 `main` 안의 발췌입니다.

```go
var small uint8 = 255
small++
fmt.Println(small)
```

출력은 `0`입니다. `uint8` 변수의 덧셈이 표현 범위를 넘으면서 값이 순환한 것입니다. 이런 범위 초과를 **오버플로(overflow)**라고 합니다. 부호 없는 정수 연산의 동작은 [정수 오버플로 규칙](https://go.dev/ref/spec#Integer_overflow)에 정의되어 있습니다.

변환에서도 정보가 사라질 수 있습니다. `var large int = 300` 다음에 `small := uint8(large)`를 계산하면 `small`은 `44`가 됩니다. 변수의 정수 형 변환이 범위를 검사해 자동으로 오류를 반환하는 것은 아닙니다.

반면 `var small uint8 = 256` 또는 `uint8(300)`처럼 **범위를 넘는 상수**를 해당 타입에 넣는 코드는 컴파일에 실패합니다. “상수가 검사되니 변수 연산도 모두 안전하다”고 생각하지 않습니다. 값을 담을 타입의 범위를 먼저 확인합니다.

### 8.2 소수는 항상 정확한 십진수로 저장되지 않는다

`float64`는 정해진 비트 수로 이진 부동소수점 값을 표현합니다. `0.1`과 `0.2` 같은 값은 정확한 십진수 그대로 저장되지 않습니다. 다음은 `main` 안의 발췌입니다.

```go
a, b := 0.1, 0.2
fmt.Printf("변수의 합: %.17f\n", a+b)
fmt.Println("0.3과 같음:", a+b == 0.3)
fmt.Println("상수식:", 0.1+0.2)
```

예상 출력:

```text
변수의 합: 0.30000000000000004
0.3과 같음: false
상수식: 0.3
```

첫 계산은 `float64` 변수 두 개를 더합니다. 마지막 계산은 상수 표현식을 계산한 뒤 출력에 필요한 타입으로 바꿉니다. 그래서 **상수식 `0.1+0.2`만으로 변수의 부동소수점 연산을 설명하면 안 됩니다.** 관련 규칙은 [숫자 타입](https://go.dev/ref/spec#Numeric_types)과 [상수 표현식](https://go.dev/ref/spec#Constant_expressions)에 있습니다.

`%.2f`로 `0.30`을 표시할 수 있어도 내부의 근사값이 정확한 십진수로 바뀌는 것은 아닙니다. 어떤 오차를 허용할지는 계산의 목적에 따라 정해야 합니다. 이번 장의 구매 금액 문제는 **작은 정수 원 단위 금액**으로 범위를 정해 정수 연산으로 해결합니다.

## 9. 자기 파일에서 값을 바꾸며 확인하기

개인 실험용 파일은 `_practice/02-variables-types/lab/main.go`에 둡니다. 아래 준비 명령은 해당 파일이 이미 있으면 보존합니다. 01장의 파일과 다른 폴더이므로 학습 기록을 구분할 수 있습니다.

**Windows PowerShell**

```powershell
New-Item -ItemType Directory -Force _practice/02-variables-types/lab | Out-Null
if (-not (Test-Path _practice/02-variables-types/lab/main.go)) {
    Copy-Item examples/02-variables-types/variables/main.go _practice/02-variables-types/lab/main.go
}
```

**macOS·Linux·WSL**

```bash
mkdir -p _practice/02-variables-types/lab
if [ ! -e _practice/02-variables-types/lab/main.go ]; then
    cp examples/02-variables-types/variables/main.go _practice/02-variables-types/lab/main.go
fi
```

본문의 발췌 코드를 시험할 때는 이 파일의 `main` 내부를 해당 발췌로 바꿉니다. 서로 다른 실험의 선언을 계속 덧붙이면 변수 이름이 중복될 수 있으므로 한 번에 한 실험씩 진행합니다.

다음은 공부 횟수를 분과 시간으로 바꾸는 **개인 실습 파일 전체**입니다. `_practice/02-variables-types/lab/main.go`를 아래 내용으로 수정하여 저장합니다.

```go
package main

import "fmt"

func main() {
	const minutesPerSession = 25
	const minutesPerHour = 60
	sessions := 4
	totalMinutes := sessions * minutesPerSession
	hours := totalMinutes / minutesPerHour
	minutes := totalMinutes % minutesPerHour

	fmt.Printf("합계: %d분\n", totalMinutes)
	fmt.Printf("환산: %d시간 %d분\n", hours, minutes)
}
```

저장소 루트에서 포맷하고 실행합니다.

```sh
gofmt -w _practice/02-variables-types/lab/main.go
go run ./_practice/02-variables-types/lab
```

예상 출력:

```text
합계: 100분
환산: 1시간 40분
```

출력문에 `100`, `1`, `40`을 직접 적지 않았는데도 계산된 값이 표시됩니다. `sessions`를 `5`로 바꾸면 총 `125`분, `2`시간 `5`분이 됩니다. `0`으로 바꾸면 모두 `0`이 됩니다. 각 경우를 실행하기 전에 먼저 계산해 보세요.

`_practice/`는 저장소의 공통 검사인 `go test ./...`에 자동으로 포함되지 않습니다. 위처럼 개인 폴더를 직접 지정하여 확인합니다. 이유와 기본 작업 방식은 [연습문제 안내](../../exercises/README.md)에 있습니다.

## 10. 자주 하는 실수와 오류 읽기

### 10.1 오류가 가리키는 두 타입 확인

아래는 **의도적으로 컴파일에 실패하는 전체 코드**입니다. 개인 파일 `_practice/02-variables-types/lab/main.go`에서만 시험합니다.

```go
package main

import "fmt"

func main() {
	var count int = 3
	var extra int64 = 2
	fmt.Println(count + extra)
}
```

실행:

```sh
go run ./_practice/02-variables-types/lab
```

오류 메시지에서 `mismatched types int and int64`라는 부분을 찾습니다. 파일 위치에서 두 변수의 선언을 확인하고, 덧셈을 `int64(count) + extra`로 고칩니다. 다시 실행하면 `5`가 출력됩니다. 형 변환은 어떤 타입으로 계산할지 명시하는 선택이며, 이 두 값은 `int64`로 표현할 수 있습니다.

### 10.2 증상별 확인표

| 증상·코드 | 원인 | 확인과 수정 |
| --- | --- | --- |
| `count := 1` 뒤 같은 블록에서 `count := 2` | 새 이름 없이 짧은 선언 반복 | 값 변경이면 `count = 2` |
| 함수 바깥의 `count := 1` | 짧은 선언을 허용하지 않는 위치 | 패키지 수준에서는 `var`; 실습 값은 `main` 안에 배치 |
| `declared and not used` | 선언한 지역 변수를 사용하지 않음 | 필요한 계산·출력에 사용하거나 불필요한 선언 제거 |
| `int` 변수에 `"3"` 대입 | 숫자를 적은 문자열과 정수 혼동 | 필요한 값의 종류를 확인하고 정수 값은 `3`으로 표현 |
| 상수에 다시 대입 | `const`로 정한 값 변경 시도 | 바뀌는 데이터는 처음부터 `var` 또는 `:=`로 선언 |
| `mismatched types int and int64` | 서로 다른 숫자 타입의 변수를 연산 | 표현 범위를 확인하고 같은 타입으로 명시적 변환 |
| `int(3.9)` | 상수를 정수로 표현할 수 없음 | 목적에 맞는 정수 상수를 사용하거나 실수 변수 변환의 절단 규칙 이해 |
| `uint8(300)` | 변환 대상 타입의 범위를 넘는 상수 | 값을 표현할 수 있는 타입 선택 |
| `fmt.Println(count++)` | 문장을 값처럼 사용 | `count++`와 `fmt.Println(count)`를 별도 문장으로 작성 |
| 소수가 필요한데 정수 몫만 나옴 | 정수 나눗셈 후 변환 | 나눗셈의 피연산자를 먼저 실수로 변환 |
| 정수 나눗셈에서 제수가 0 | 정의되지 않은 정수 나눗셈 | 입력 조건을 확인하고 0일 때 나누지 않도록 설계 |
| `float64` 계산 결과의 직접 비교가 예상과 다름 | 저장·계산 과정의 근사 오차 | 입력 타입과 계산 과정을 확인하고 목적에 맞는 오차 기준 마련 |

오류 메시지의 전체 문구는 버전에 따라 다를 수 있습니다. **파일과 위치 → 변수 선언 → 실제 타입 → 계산 순서** 순으로 확인합니다. 컴파일에 성공해도 소수 부분을 이미 잃었거나 작은 타입에서 값이 순환할 수 있으므로 예상 결과와의 비교도 필요합니다.

## 11. 선택 학습: `iota`와 비트 연산

이 절은 기본 문제 1–6의 필수 내용이 아닙니다. 변수와 정수 계산에 익숙해진 뒤 읽어도 됩니다.

### 11.1 연속된 상수에 번호 붙이기

`iota`는 `const` 선언 묶음 안에서 0부터 시작해 각 상수 선언을 지날 때 증가하는 값을 제공합니다. 다음은 `main` 안의 발췌입니다.

```go
const (
	levelBeginner = iota
	levelIntermediate
	levelAdvanced
)
fmt.Println(levelBeginner, levelIntermediate, levelAdvanced)
```

출력은 `0 1 2`입니다. 뒤의 두 줄은 앞의 상수 식인 `iota`를 반복하는 형태이며, 각 선언에서 다른 값이 됩니다. 새로운 `const (...)` 묶음에서는 다시 0부터 시작합니다. 한 선언에 상수 이름을 여러 개 적으면 그 선언에서 사용하는 `iota` 값은 같습니다. [iota 규칙](https://go.dev/ref/spec#Iota)을 참고합니다.

`iota`는 정수에 이름을 붙이는 데 편리하지만, 상수의 순서를 바꾸면 번호도 바뀔 수 있습니다. 외부 파일이나 API에 고정된 번호로 저장해야 하는 값은 순서 변경의 영향을 확인하고 명시적인 값을 선택할 수 있습니다. 사용자 정의 타입과 결합하는 방법은 이후에 다룹니다.

### 11.2 정수의 각 비트를 계산하기

정수의 비트 단위 계산에는 다음 연산자가 있습니다. `&&`, `||`의 불리언 논리 연산과 구분합니다.

| 연산자 | 정수 비트에 하는 일 |
| --- | --- |
| `&` | 두 비트가 모두 1이면 1 |
| `\|` | 둘 중 하나 이상이 1이면 1 |
| 이항 `^` | 두 비트가 다르면 1 |
| `&^` | 오른쪽에서 1인 비트를 왼쪽에서 제거 |
| 단항 `^` | 비트 반전 |
| `<<`, `>>` | 비트를 왼쪽·오른쪽으로 이동 |

다음은 `main` 안의 발췌입니다.

```go
var a uint8 = 6
var b uint8 = 3
fmt.Printf("AND: %04b\n", a&b)
fmt.Printf("OR: %04b\n", a|b)
fmt.Printf("XOR: %04b\n", a^b)
fmt.Printf("제거: %04b\n", a&^b)
fmt.Printf("왼쪽 이동: %04b\n", a<<1)
fmt.Printf("오른쪽 이동: %04b\n", a>>1)
```

예상 출력:

```text
AND: 0010
OR: 0111
XOR: 0101
제거: 0100
왼쪽 이동: 1100
오른쪽 이동: 0011
```

`%b`는 이진수 출력이며 `%04b`는 최소 네 자리를 사용하고 빈 앞자리를 `0`으로 채웁니다. `6`은 이진수 `0110`, `3`은 `0011`입니다. 예를 들어 AND 결과는 두 값에서 모두 1인 자리만 남겨 `0010`이 됩니다. 숫자를 지정한 폭으로 잘라 출력한다는 뜻은 아닙니다. [정수 연산](https://go.dev/ref/spec#Integer_operators)과 [fmt 서식](https://pkg.go.dev/fmt#hdr-Printing)을 참고합니다.

## 12. 연습문제와 출력 검증

[02장 연습문제](../../exercises/02-variables-types/README.md)는 기본 문제 6개와 선택 문제 1개로 구성됩니다. 개인 파일 준비, 입력 조건, 예상 출력과 힌트가 함께 있습니다.

| 문제 | 수준 | 확인할 내용 |
| --- | --- | --- |
| 1. 값의 변화 예측 | 확인 | 영값, 계산 시점, 값 복사 |
| 2. 학습 기록 | 변형 | 변수 선언, `+=`, 정수 곱셈, 불리언 출력 |
| 3. 시·분·초 환산 | 변형 | 상수, 몫·나머지, 경계 입력 |
| 4. 타입 오류 수정 | 오류 진단 | `int`·`int64`와 명시적 형 변환 |
| 5. 평균 계산 | 응용 | 나누기 전 변환, 출력 자릿수 |
| 6. 구매 내역 계산 | 응용 | 정수 금액, 계산 순서, 예산 비교 |
| 선택 A. 상태 번호 | 선택 | `iota`를 사용한 상수 선언 |

먼저 요구사항을 읽고 직접 풀어 본 뒤 [02장 해설](../../solutions/02-variables-types/README.md)에서 계산 과정과 구현을 비교합니다. 0초, 59초, 60초처럼 결과의 자리 올림이 달라지는 입력을 확인하는 이유도 설명했습니다.

저장소에 제공한 예제와 참고 풀이에는 01장에서 배운 출력 테스트가 있습니다. 다음 명령으로 실행합니다.

```sh
go test -count=1 ./examples/02-variables-types/... ./solutions/02-variables-types/...
```

성공한 패키지마다 `ok`가 표시됩니다. 이것은 **저장소의 예제와 참고 풀이**에 대한 검사입니다. 자신의 `_practice/` 코드가 맞는지는 개인 경로로 실행하고 문제의 결과와 비교해야 합니다. 특히 고정된 예제 출력만 통과했다고 모든 입력에서 계산이 옳다는 뜻은 아닙니다. 문제에 제시한 추가 입력도 바꾸어 확인합니다.

## 13. 요약과 자기 점검

변수에는 타입과 현재 값이 있습니다. `var`나 `:=`로 선언하고, `=`와 복합 대입으로 값을 바꿉니다. 초기값을 생략하면 타입의 영값으로 시작합니다. 상수는 변하지 않는 기준값을 표현하며, 타입이 지정되지 않은 숫자 상수는 사용하는 문맥의 타입과 범위에 맞춰 처리됩니다.

계산에서는 값만큼 **타입과 계산 순서**가 중요합니다. 정수 나눗셈의 소수 부분은 결과를 나중에 실수로 바꾸어도 되살아나지 않습니다. 작은 정수 타입의 범위, 부동소수점의 근사값, 출력 형식이 실제 저장값과 다를 수 있다는 점도 함께 확인합니다.

### 이해 확인

1. `var ready bool`과 `var title string`의 초기값은 무엇인가요?
2. 같은 블록에서 이미 선언한 `count`만 바꾸려면 `:=`와 `=` 중 무엇을 사용하나요?
3. `ratio := 3.0`으로 만든 변수의 타입은 무엇이며, 이후 문자열을 넣을 수 있나요?
4. `a := 3; b := a; a = 8`을 실행한 뒤 `b`는 무엇인가요?
5. 정수 변수 `total = 7`, `count = 2`에서 `float64(total/count)`와 `float64(total)/float64(count)`는 왜 다른가요?
6. `x := 3.9` 다음의 `int(x)`와 상수식 `int(3.9)`는 어떻게 다른가요?
7. `const limit = 300`을 `uint8` 변수에 저장할 수 있나요?
8. `count != 0 && 10/count > 1`에서 `count`가 0이면 왜 나눗셈을 하지 않나요?
9. `%.2f`로 출력하면 원래 변수의 값도 소수점 아래 두 자리로 바뀌나요?

<details>
<summary>답과 설명</summary>

1. 각각 `false`와 빈 문자열 `""`입니다. 공백 한 칸과 빈 문자열은 다릅니다.
2. `=`를 사용합니다. 같은 블록에서 새 이름 없이 `:=`를 반복하면 오류입니다.
3. `float64`입니다. 타입 추론으로 정해졌으며 문자열로 바꿔 담을 수 없습니다.
4. `3`입니다. 대입 당시의 정수 값을 복사했으며 `a`의 변경과 연결되어 있지 않습니다.
5. 앞은 정수 나눗셈 결과 `3`을 변환하여 `3.0`, 뒤는 먼저 실수로 변환하여 나누므로 `3.5`입니다.
6. `int(x)`는 변수의 소수 부분을 버려 `3`을 만듭니다. `int(3.9)`는 그 상수를 정수로 표현할 수 없어 컴파일 오류입니다.
7. 안 됩니다. `uint8` 범위는 `0`부터 `255`이며, 이 상수는 범위를 넘습니다.
8. `&&` 왼쪽이 거짓이면 결과가 이미 거짓이므로 오른쪽을 실행하지 않습니다. 컴파일 검사까지 생략하는 것은 아닙니다.
9. 아닙니다. 출력의 표시 자릿수만 정하며 저장된 값과 계산 정밀도는 바꾸지 않습니다.

</details>

### 완료 체크리스트

- [ ] 변수를 새로 선언하는 코드와 값을 바꾸는 코드를 구분했다.
- [ ] 기본 타입의 영값과 `%T`·`%q` 출력 결과를 설명할 수 있다.
- [ ] 시간 계산에서 상수와 바꿀 입력값을 나누어 사용했다.
- [ ] 정수 나눗셈과 실수 나눗셈의 차이를 직접 확인했다.
- [ ] 타입 오류를 재현하고 변환이 필요한 이유를 설명하며 수정했다.
- [ ] 기본 문제 1–6을 풀고 추가 입력의 계산 결과도 확인했다.

## 참고 자료

- [Go 명세: 변수 선언](https://go.dev/ref/spec#Variable_declarations) — 초기화와 타입 추론
- [Go 명세: 짧은 변수 선언](https://go.dev/ref/spec#Short_variable_declarations) — `:=`와 기존 변수 재사용
- [Go 명세: 영값](https://go.dev/ref/spec#The_zero_value) — 타입별 초기값
- [Go 명세: 상수 선언](https://go.dev/ref/spec#Constant_declarations) — 타입 있는 상수와 상수 식
- [Go 공식 블로그: Constants](https://go.dev/blog/constants) — 상수와 타입을 이해하는 보충 설명
- [Go 명세: 숫자 타입 사이의 변환](https://go.dev/ref/spec#Conversions) — 절단·범위·정밀도
- [Go 명세: 연산자](https://go.dev/ref/spec#Operators) — 산술·비교·논리·비트 연산
- [fmt: Printing](https://pkg.go.dev/fmt#hdr-Printing) — 타입·값·소수 자릿수 출력

확인일: 2026-11-09. 예제와 의도적인 실패 사례는 Go 1.27.3에서 확인했습니다.

## 다음 학습

다음은 **03. 조건문과 반복문**이며 아직 집필 예정입니다. 이번 장에서는 비교 결과를 `bool`로 만들었습니다. 다음 장에서는 그 결과에 따라 실행할 코드를 선택하고, 같은 동작을 여러 값에 반복하는 방법으로 이어집니다. 집필 상태는 [전체 목차](../README.md)와 [로드맵](../../docs/ROADMAP.md)에서 확인합니다.

| 이전 | 목차 | 실습과 해설 |
| --- | --- | --- |
| [01. 첫 Go 프로그램](../01-first-program/README.md) | [전체 목차](../README.md) | [문제](../../exercises/02-variables-types/README.md) · [해설](../../solutions/02-variables-types/README.md) |
