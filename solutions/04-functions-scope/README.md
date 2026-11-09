# 04장 해설: 함수와 스코프

문제: [04장 연습문제](../../exercises/04-functions-scope/README.md)

함수를 읽을 때는 **무엇을 받아서 무엇을 반환하는지**, 변수를 읽을 때는 **어디에서 선언되었는지**를 먼저 확인합니다. 이 장에서는 계산 결과를 반환하는 함수와 화면에 출력하는 `main`의 역할도 나눕니다.

모든 명령은 **저장소 루트**에서 실행합니다. `solutions/`는 참고 구현이며 `_practice/`는 개인 풀이입니다. 참고 구현의 테스트가 통과하는 것과 자신의 프로그램이 요구사항을 만족하는 것은 각각 확인해야 합니다.

## 문제 1. 인자와 같은 이름의 변수 추적

출력은 다음 두 줄입니다.

```text
블록 안: 99
호출 뒤: 10, 11
```

전체 코드는 문제와 같습니다. 직접 실행하려면 기존 개인 내용이 없는 `_practice/04-functions-scope/predict/main.go`에 저장합니다. 기존 파일이 있으면 내용을 먼저 보존하세요.

```go
package main

import "fmt"

func addOne(value int) int {
	value++
	return value
}

func main() {
	number := 10
	next := addOne(number)
	if next > 10 {
		number := 99
		fmt.Println("블록 안:", number)
	}
	fmt.Printf("호출 뒤: %d, %d\n", number, next)
}
```

실행:

```text
go run ./_practice/04-functions-scope/predict
```

### 같은 이름보다 선언 위치를 먼저 보기

| 순서 | 실행한 일 | 확인할 값 |
| --- | --- | --- |
| 1 | `main`에서 `number` 선언 | 바깥 `number`는 `10` |
| 2 | `addOne(number)` 호출 | 매개변수 `value`가 `10`을 받음 |
| 3 | `value++` 실행 후 반환 | `value`는 `11`, 반환값도 `11` |
| 4 | 반환값을 `next`에 저장 | `next`는 `11`, 바깥 `number`는 `10` |
| 5 | 조건이 참이므로 `if` 안에 진입 | 안쪽 `number`를 새로 만들고 `99` 저장 |
| 6 | 블록 안에서 출력 | 이 위치의 `number`는 안쪽 변수이므로 `99` |
| 7 | 블록 뒤에서 출력 | 바깥 `number`는 `10`, `next`는 `11` |

이 문제의 `int` 값은 호출 시 매개변수에 복사됩니다. `value++`는 `addOne` 안의 매개변수를 바꿉니다. 호출자가 반환값을 `next`에 저장했으므로 바깥 `number`를 변경하는 대입은 없습니다. [함수 호출 명세](https://go.dev/ref/spec#Calls)를 참고합니다.

`if` 안의 `number := 99`는 새 변수를 선언하여 바깥의 같은 이름을 가립니다. 이를 **이름 가림(shadowing)**이라고 합니다. 블록을 나와서는 바깥 `number`를 다시 사용합니다. 만약 안쪽 문장이 `number = 99`였다면 바깥 변수에 대입하므로 마지막 출력도 달라집니다. [선언과 스코프](https://go.dev/ref/spec#Declarations_and_scope)에서 이 이름의 유효 범위를 확인할 수 있습니다.

## 문제 2. 점수 판정을 함수로 분리

참고 코드: [grade/main.go](grade/main.go) · [함수·출력 테스트](grade/main_test.go)

개인 파일: `_practice/04-functions-scope/grade/main.go`

전체 코드:

```go
package main

import "fmt"

func classifyScore(score int) (string, bool) {
	if score < 0 || score > 100 {
		return "", false
	}
	if score >= 90 {
		return "우수", true
	}
	if score >= 60 {
		return "통과", true
	}
	return "다시 연습", true
}

func main() {
	score := 72
	label, ok := classifyScore(score)
	if !ok {
		fmt.Println("점수 확인")
		return
	}
	fmt.Printf("%d점: %s\n", score, label)
}
```

참고 구현 실행과 검사:

```text
go run ./solutions/04-functions-scope/grade
go test ./solutions/04-functions-scope/grade
```

개인 풀이 실행:

```text
go run ./_practice/04-functions-scope/grade
```

예상 출력:

```text
72점: 통과
```

### 반환값에 의미를 붙이기

`classifyScore`는 입력 점수만 보고 판정 문구와 유효성을 반환합니다. 함수 내부에서 출력하거나 외부 상태를 바꾸지 않으므로 같은 점수에는 같은 결과를 얻습니다. 이런 계산을 **순수 함수(pure function)**로 나누면 화면 출력과 별개로 입력별 결과를 검사하기 좋습니다.

범위 밖 입력은 먼저 `"", false`를 반환합니다. 이후에는 0~100이라는 전제에서 높은 등급부터 확인합니다. `return`을 만난 호출은 그 자리에서 끝나므로, 한 등급을 반환한 뒤 아래 조건을 계속 검사하지 않습니다. [return 문](https://go.dev/ref/spec#Return_statements)을 참고합니다.

`label, ok := ...`는 두 반환값을 순서대로 받습니다. `ok`는 특별한 예약어가 아니라, 여기서는 유효성을 나타내도록 정한 변수 이름입니다. 이 예제의 `bool` 규칙도 우리가 정한 함수의 계약입니다. 유효하지 않을 때 `main`에서 출력 후 반환하므로 성공 문구가 추가로 나오지 않습니다.

| `score` | `classifyScore`의 반환값 | 프로그램 출력 |
| --- | --- | --- |
| `-1` | `"", false` | `점수 확인` |
| `0` | `"다시 연습", true` | `0점: 다시 연습` |
| `59` | `"다시 연습", true` | `59점: 다시 연습` |
| `60` | `"통과", true` | `60점: 통과` |
| `89` | `"통과", true` | `89점: 통과` |
| `90` | `"우수", true` | `90점: 우수` |
| `100` | `"우수", true` | `100점: 우수` |
| `101` | `"", false` | `점수 확인` |

표의 프로그램 출력은 각각 한 줄이며 마지막에 줄바꿈이 있습니다. 개인 코드에서 점수를 바꾸어 확인했다면 `score := 72`로 복구합니다.

## 문제 3. 변경되지 않는 반환값 고치기

문제의 원래 프로그램은 다음을 출력합니다.

```text
블록 안: 통과
반환값: 계속 연습
```

블록 안에서 새로 선언한 `label`은 `"통과"`지만, 함수가 반환하는 바깥 `label`은 여전히 `"계속 연습"`입니다. 기존 바깥 변수에 대입하도록 **안쪽 `:=`를 `=`로 바꾸면** 요구사항을 만족합니다.

개인 파일: `_practice/04-functions-scope/repair/main.go`

수정한 전체 코드:

```go
package main

import "fmt"

func lessonLabel(score int) string {
	label := "계속 연습"
	if score >= 60 {
		label = "통과"
		fmt.Println("블록 안:", label)
	}
	return label
}

func main() {
	fmt.Println("반환값:", lessonLabel(72))
}
```

실행:

```text
go run ./_practice/04-functions-scope/repair
```

수정 후 예상 출력:

```text
블록 안: 통과
반환값: 통과
```

바깥 `label := "계속 연습"`은 처음 변수를 만드는 선언이므로 그대로 둡니다. 안쪽의 `=`는 이미 선언된 변수의 값을 바꿉니다. `lessonLabel`의 결과가 나온 뒤에야 `main`의 `Println`이 실행되므로, 함수 안의 추적 출력이 먼저 나옵니다. 이 문제에서는 흐름을 관찰하기 위해 함수 안의 출력을 유지하지만, 문제 2의 판정 함수는 출력 없이 결과만 반환하도록 요구했습니다.

| 호출 | 수정 후 출력 순서 |
| --- | --- |
| `lessonLabel(59)` | `반환값: 계속 연습` |
| `lessonLabel(60)` | `블록 안: 통과` → `반환값: 통과` |
| `lessonLabel(72)` | `블록 안: 통과` → `반환값: 통과` |

화살표는 서로 다른 출력 줄을 뜻합니다. 59점에서는 `if` 본문이 실행되지 않으므로 블록 안 출력도 없습니다. 경계를 확인한 뒤 인자를 `72`로 복구하세요. 규칙의 근거는 [짧은 변수 선언](https://go.dev/ref/spec#Short_variable_declarations)과 [대입문](https://go.dev/ref/spec#Assignment_statements)에 있습니다.

## 문제 4. 여러 수에서 짝수만 더하기

참고 코드: [sum-even/main.go](sum-even/main.go) · [함수·출력 테스트](sum-even/main_test.go)

개인 파일: `_practice/04-functions-scope/sum-even/main.go`

전체 코드:

```go
package main

import "fmt"

func sumEven(numbers ...int) int {
	total := 0
	for _, number := range numbers {
		if number%2 == 0 {
			total += number
		}
	}
	return total
}

func main() {
	fmt.Println("짝수 합계:", sumEven(1, 2, 3, 4, 6))
}
```

참고 구현 실행과 검사:

```text
go run ./solutions/04-functions-scope/sum-even
go test ./solutions/04-functions-scope/sum-even
```

개인 풀이 실행:

```text
go run ./_practice/04-functions-scope/sum-even
```

예상 출력:

```text
짝수 합계: 12
```

### 가변 인자를 순서대로 확인하기

`numbers ...int`는 정수를 0개 이상 받을 수 있는 마지막 매개변수입니다. 함수 안에서 `numbers`의 타입은 `[]int`이며, 여기서는 받은 값들을 순서대로 꺼내는 데만 사용합니다. `range`가 제공하는 순서 번호는 `_`로 버리고, 값은 `number`로 받습니다. 슬라이스의 생성·변경·공유 방식은 05장에서 학습합니다. [가변 인자](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)와 [range 문](https://go.dev/ref/spec#For_range)의 규칙을 참고합니다.

기본 호출에서 1과 3은 더하지 않고 2, 4, 6을 더하므로 12가 됩니다. `return`을 반복 뒤에 두어 모든 인자를 확인합니다. 빈 호출에서는 반복이 한 번도 실행되지 않으므로 초기 합계인 0을 그대로 반환합니다.

| 호출 | 더해지는 값 | 반환값 |
| --- | --- | --- |
| `sumEven()` | 없음 | `0` |
| `sumEven(8)` | `8` | `8` |
| `sumEven(1, 3, 5)` | 없음 | `0` |
| `sumEven(-2, 3, 4)` | `-2`, `4` | `2` |
| `sumEven(-4, -3, 2, 0)` | `-4`, `2`, `0` | `-2` |

음수에서도 짝수의 나머지는 0입니다. 반면 `-3 % 2`는 `-1`이므로, 홀수를 제외하려고 `%2 == 1`만 검사하면 음수 홀수를 놓칠 수 있습니다. 짝수의 포함 조건인 `%2 == 0`을 쓰면 양수와 음수를 함께 처리합니다. [정수의 나눗셈과 나머지](https://go.dev/ref/spec#Arithmetic_operators)를 참고하세요.

개인 코드의 호출을 바꿔 확인했다면 `sumEven(1, 2, 3, 4, 6)`으로 복구합니다. 이 문제는 입력과 합계가 `int`의 범위 안인 작은 수만 다룹니다.

## 문제 5. 학습 계획을 두 함수로 계산

참고 코드: [study-plan/main.go](study-plan/main.go) · [함수·출력 테스트](study-plan/main_test.go)

개인 파일: `_practice/04-functions-scope/study-plan/main.go`

전체 코드:

```go
package main

import "fmt"

func totalMinutes(days, minutesPerDay int) (int, bool) {
	if days < 0 || days > 7 || minutesPerDay < 1 || minutesPerDay > 120 {
		return 0, false
	}
	return days * minutesPerDay, true
}

func splitMinutes(total int) (int, int) {
	return total / 60, total % 60
}

func main() {
	days := 5
	minutesPerDay := 25
	total, ok := totalMinutes(days, minutesPerDay)
	if !ok {
		fmt.Println("계획 확인")
		return
	}
	hours, minutes := splitMinutes(total)
	fmt.Printf("총 계획: %d분\n환산: %d시간 %d분\n", total, hours, minutes)
}
```

참고 구현 실행과 검사:

```text
go run ./solutions/04-functions-scope/study-plan
go test ./solutions/04-functions-scope/study-plan
```

개인 풀이 실행:

```text
go run ./_practice/04-functions-scope/study-plan
```

예상 출력:

```text
총 계획: 125분
환산: 2시간 5분
```

### 각 함수의 책임을 좁히기

`totalMinutes`는 날짜와 하루 분량이 계약의 범위에 있는지 확인하고 총 분량을 구합니다. `days, minutesPerDay int`는 두 매개변수의 타입이 같아서 묶어 쓴 선언입니다. 범위를 벗어난 경우의 반환값은 `0, false`로 정했습니다.

`splitMinutes`는 이미 유효하다고 확인한 총 분량을 환산합니다. 이 예제에서 `main`은 첫 함수의 `ok`가 참일 때만 두 번째 함수를 호출하므로 입력은 0~840입니다. 다른 곳에서 `splitMinutes`를 직접 호출한다면 그 호출자가 같은 전제를 지켜야 합니다. 입력 검증이 언어에 의해 자동으로 추가되는 것은 아닙니다.

마지막으로 `main`은 두 계산을 순서대로 연결하고 결과를 출력합니다. 이렇게 나누면 총 분량 계산과 시간 환산을 각각 확인할 수 있습니다.

### 유효한 0과 실패를 구분하기

| `days`, `minutesPerDay` | `totalMinutes`의 반환값 | 프로그램 출력 |
| --- | --- | --- |
| `0, 25` | `0, true` | `총 계획: 0분` → `환산: 0시간 0분` |
| `1, 1` | `1, true` | `총 계획: 1분` → `환산: 0시간 1분` |
| `7, 120` | `840, true` | `총 계획: 840분` → `환산: 14시간 0분` |
| `-1, 25` | `0, false` | `계획 확인` |
| `8, 25` | `0, false` | `계획 확인` |
| `5, 0` | `0, false` | `계획 확인` |
| `5, 121` | `0, false` | `계획 확인` |
| `0, 0` | `0, false` | `계획 확인` |

화살표는 줄바꿈을 뜻합니다. 날짜가 0이라도 하루 분량의 조건은 적용됩니다. 예를 들어 `0, 121`도 유효하지 않습니다. 유효한 쉬는 계획과 잘못된 입력이 모두 숫자 0을 반환할 수 있으므로 `total == 0` 대신 `ok`를 확인해야 합니다.

`splitMinutes`의 반환 순서는 시간, 남은 분입니다.

| `total` | 시간 | 남은 분 |
| --- | --- | --- |
| `0` | `0` | `0` |
| `59` | `0` | `59` |
| `60` | `1` | `0` |
| `125` | `2` | `5` |
| `840` | `14` | `0` |

항상 `시간 * 60 + 남은 분 == total`이며 남은 분은 0~59입니다. 02장에서 배운 몫과 나머지 계산을 함수로 옮긴 것입니다. 여러 반환값의 순서와 타입은 [함수 타입](https://go.dev/ref/spec#Function_types)에 선언되어 있으며, 호출할 때도 그 순서대로 받습니다.

개인 코드에서 입력이나 환산 호출을 바꾸었다면 `days := 5`, `minutesPerDay := 25`와 기본 계산 흐름으로 복구합니다.

## 문제 6. 호출할 때마다 움직이는 값

참고 코드: [stepper/main.go](stepper/main.go) · [함수·출력 테스트](stepper/main_test.go)

개인 파일: `_practice/04-functions-scope/stepper/main.go`

전체 코드:

```go
package main

import "fmt"

func newStepper(start, step int) func() int {
	current := start
	return func() int {
		current += step
		return current
	}
}

func main() {
	first := newStepper(10, 3)
	second := newStepper(100, 10)
	fmt.Println("첫 번째:", first())
	fmt.Println("첫 번째:", first())
	fmt.Println("두 번째:", second())
	fmt.Println("첫 번째:", first())
}
```

참고 구현 실행과 검사:

```text
go run ./solutions/04-functions-scope/stepper
go test ./solutions/04-functions-scope/stepper
```

개인 풀이 실행:

```text
go run ./_practice/04-functions-scope/stepper
```

예상 출력:

```text
첫 번째: 13
첫 번째: 16
두 번째: 110
첫 번째: 19
```

### 함수를 반환하고 그 함수가 상태를 기억하기

`newStepper`의 반환 타입 `func() int`는 “인자가 없고 `int`를 반환하는 함수”입니다. `return func() int { ... }`는 이름 없는 함수를 만들어 반환합니다. 반환된 함수는 바깥의 `current`와 `step`을 사용하므로 클로저입니다. 바깥 함수의 호출이 끝나도 그 변수는 이 클로저를 통해 계속 사용됩니다. [함수 리터럴과 클로저](https://go.dev/ref/spec#Function_literals)의 정의를 참고합니다.

`newStepper(10, 3)`을 실행하는 순간에는 `current`를 10으로 준비하고 함수 값을 반환합니다. 첫 `first()`를 실행할 때 비로소 `current += step`이 실행되어 13을 반환합니다. 두 번째 호출은 앞선 결과인 13에서 다시 3을 더합니다.

| 실행 순서 | 첫 번째 함수의 현재 값 | 두 번째 함수의 현재 값 | 출력 값 |
| --- | --- | --- | --- |
| 두 함수 생성 직후 | `10` | `100` | 아직 없음 |
| `first()` | `13` | `100` | `13` |
| `first()` | `16` | `100` | `16` |
| `second()` | `16` | `110` | `110` |
| `first()` | `19` | `110` | `19` |

두 번의 `newStepper` 호출은 각각의 `current`를 만듭니다. 따라서 두 번째 함수를 호출해도 첫 번째 함수의 현재 값은 바뀌지 않습니다. `current := start`를 반환되는 함수 안으로 옮기면 매 호출마다 초기화되어 이 누적 동작을 잃습니다.

| 새로 만든 함수 | 두 번 호출한 결과 | 의미 |
| --- | --- | --- |
| `newStepper(0, 1)` | `1`, `2` | 1씩 증가 |
| `newStepper(10, 0)` | `10`, `10` | 현재 값 유지 |
| `newStepper(5, -2)` | `3`, `1` | 2씩 감소 |

`start`와 `step`은 음수도 받을 수 있습니다. 다만 이 문제에서는 계산 결과가 `int` 범위를 넘지 않는 작은 값과 호출 횟수를 사용합니다. 변형을 확인한 뒤 기본 두 함수와 네 번의 호출로 복구하세요.

## 선택 A. 같은 연산을 두 번 적용

개인 파일: `_practice/04-functions-scope/apply-twice/main.go`

전체 코드:

```go
package main

import "fmt"

func applyTwice(value int, operation func(int) int) int {
	first := operation(value)
	return operation(first)
}

func main() {
	addThree := func(value int) int {
		return value + 3
	}
	double := func(value int) int {
		return value * 2
	}
	fmt.Println("3 더하기 두 번:", applyTwice(5, addThree))
	fmt.Println("2배 두 번:", applyTwice(5, double))
}
```

실행:

```text
go run ./_practice/04-functions-scope/apply-twice
```

예상 출력:

```text
3 더하기 두 번: 11
2배 두 번: 20
```

`operation func(int) int`는 정수 하나를 받아 정수 하나를 반환하는 함수를 인자로 받겠다는 뜻입니다. `addThree`와 `double` 모두 그 함수 타입에 맞으므로 같은 `applyTwice`에 전달할 수 있습니다.

| 연산 | 첫 적용 | 두 번째 적용 | 반환값 |
| --- | --- | --- | --- |
| 3 더하기 | `5 + 3 = 8` | `8 + 3 = 11` | `11` |
| 2배 | `5 * 2 = 10` | `10 * 2 = 20` | `20` |

두 번째 적용에는 첫 결과를 전달합니다. `applyTwice`는 연산의 종류를 직접 판단하지 않고 전달받은 함수를 호출합니다. 여기서는 단계가 보이도록 `first`를 두었으며, 같은 계산을 `return operation(operation(value))`로 표현할 수도 있습니다.

이번 계약에서는 `operation`에 호출할 함수가 반드시 있습니다. 함수 값이 없는 `nil`을 호출하면 실행 중 패닉이 발생하므로, 이 경우까지 다루려면 함수의 입력 계약과 처리 방법을 별도로 정해야 합니다. [함수 호출 명세](https://go.dev/ref/spec#Calls)를 참고합니다.

## 테스트가 확인하는 범위

참고 구현의 네 패키지를 함께 확인할 수 있습니다.

```text
go test ./solutions/04-functions-scope/...
```

| 참고 구현 | 함수 테스트의 핵심 | `Example_main`이 확인하는 것 |
| --- | --- | --- |
| `grade` | 0·60·90·100의 경계, 범위 밖 반환값과 유효성 | 72점의 판정 출력 |
| `sum-even` | 빈 인자, 하나·여러 인자, 홀수만 있는 경우, 음수와 0 | 기본 다섯 수의 합계 출력 |
| `study-plan` | 날짜·하루 분량의 경계, 유효한 0과 실패 구분, 시간 환산 | 5일·하루 25분의 두 줄 출력 |
| `stepper` | 증가·정지·감소, 두 클로저를 섞어 호출했을 때의 독립성 | 기본 네 번의 호출 출력 |

`main_test.go`에는 입력 사례를 모아 검사하기 위한 슬라이스·구조체 문법이 있습니다. 테스트 작성법은 12장에서 배우므로, 지금은 문제를 풀기 위한 선수 지식이 아닙니다. 함수 선언과 문제의 경계 입력을 먼저 이해하고, 테스트 실행으로 참고 구현의 약속을 확인하면 됩니다.

개인 코드에는 이 테스트가 자동으로 적용되지 않습니다. 먼저 각 문제의 `go run ./_practice/...` 명령으로 자신의 출력을 확인하세요. 문제 1·3·선택 A는 이 문서의 전체 코드를 개인 폴더에서 실행하며, 별도의 영구 참고 패키지는 두지 않았습니다.

## 다시 설명해 볼 질문

- 함수의 매개변수를 바꾸는 것과 호출자의 변수에 반환값을 대입하는 것은 어떻게 다른가?
- 반환값이 `0`일 때도 별도의 `bool`이 필요한 이유는 무엇인가?
- 안쪽 블록의 `:=`를 `=`로 바꾸면 어느 변수에 영향을 주는가?
- `sumEven()`이 별도 분기 없이도 0을 반환하는 이유는 무엇인가?
- `newStepper`를 두 번 호출하는 것과 반환된 함수를 두 번 호출하는 것은 어떻게 다른가?
- `applyTwice`는 연산 내용을 모르고도 어떻게 두 종류의 계산을 수행하는가?

[04장 본문](../../book/04-functions-scope/README.md) · [04장 연습문제](../../exercises/04-functions-scope/README.md) · [해설 목록](../README.md)
