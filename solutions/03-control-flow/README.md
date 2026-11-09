# 03장 해설: 조건문과 반복문

문제: [03장 연습문제](../../exercises/03-control-flow/README.md)

먼저 문제와 힌트를 확인한 뒤 해설을 읽습니다. 조건문은 **어느 경로를 실행하는지**, 반복문은 **언제 시작하고 무엇을 바꾸며 언제 끝나는지**를 따라 읽습니다.

모든 명령은 **저장소 루트**에서 실행합니다. `solutions/`는 참고 구현이며, `_practice/`는 개인 풀이입니다. 참고 구현의 테스트와 개인 코드의 실행 결과를 구분해서 확인하세요.

## 문제 1. `continue`의 흐름 예측

최종 출력은 다음 한 줄입니다.

```text
9
```

직접 실행해 확인하고 싶다면 편집기에서 `_practice/03-control-flow/predict/main.go`를 준비합니다. 기존 파일이 있으면 개인 내용을 먼저 보존하세요. 전체 코드는 다음과 같습니다.

```go
package main

import "fmt"

func main() {
	total := 0
	for number := 1; number <= 5; number++ {
		if number%2 == 0 {
			continue
		}
		total += number
	}
	fmt.Println(total)
}
```

실행:

```text
go run ./_practice/03-control-flow/predict
```

### 다섯 번의 반복 추적

`total`은 반복 전에 0으로 시작합니다.

| `number` | 짝수 조건 | 이번 본문의 동작 | 본문 처리 후 `total` | 증가식 처리 후 `number` |
| --- | --- | --- | --- | --- |
| `1` | 거짓 | 1을 더함 | `1` | `2` |
| `2` | 참 | `continue`로 덧셈을 건너뜀 | `1` | `3` |
| `3` | 거짓 | 3을 더함 | `4` | `4` |
| `4` | 참 | `continue`로 덧셈을 건너뜀 | `4` | `5` |
| `5` | 거짓 | 5를 더함 | `9` | `6` |

그 다음 조건 `6 <= 5`가 거짓이 되어 반복을 마칩니다. 실제로 더한 값은 1, 3, 5이므로 합계는 9입니다.

이 세 부분 형태의 `for`에서는 `continue` 뒤에도 증가식 `number++`를 실행한 뒤 조건을 다시 검사합니다. `continue`는 현재 본문의 남은 문장을 건너뛰며, 반복 전체를 종료하지 않습니다. 이 동작을 [for 문](https://go.dev/ref/spec#For_statements)과 [continue 문](https://go.dev/ref/spec#Continue_statements)의 정의로 확인할 수 있습니다.

## 문제 2. 점수에 따른 안내

참고 코드: [grade/main.go](grade/main.go) · [출력 테스트](grade/main_test.go)

개인 파일: `_practice/03-control-flow/grade/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	score := 72

	if score < 0 || score > 100 {
		fmt.Println("점수 확인")
	} else if score >= 90 {
		fmt.Println("우수")
	} else if score >= 60 {
		fmt.Println("통과")
	} else {
		fmt.Println("다시 연습")
	}
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/03-control-flow/grade
go test -v ./solutions/03-control-flow/grade
```

출력:

```text
통과
```

먼저 0~100 범위를 벗어난 입력을 제외합니다. 그렇지 않으면 101이 `score >= 90`을 만족해 잘못 분류될 수 있습니다. 유효한 점수에서는 높은 기준을 먼저 검사합니다. 90점은 60점 이상이기도 하므로 60점 기준을 앞에 두면 의도와 다른 분기가 선택됩니다.

`else if`로 연결한 분기는 앞에서 참인 조건을 찾으면 그 본문을 실행하고 연결을 마칩니다. 서로 독립적인 `if`를 여러 개 두는 것과 구분합니다. [if 문 명세](https://go.dev/ref/spec#If_statements)를 참고하세요.

### 경계 입력

개인 파일의 `score`를 하나씩 바꾸고 `go run ./_practice/03-control-flow/grade`로 확인합니다.

| 점수 | 출력 | 확인하는 경계 |
| --- | --- | --- |
| `-1` | `점수 확인` | 유효 범위 아래 |
| `0` | `다시 연습` | 유효 범위 시작 |
| `59` | `다시 연습` | 통과 기준 바로 전 |
| `60` | `통과` | 통과 기준 포함 |
| `89` | `통과` | 우수 기준 바로 전 |
| `90` | `우수` | 우수 기준 포함 |
| `100` | `우수` | 유효 범위 마지막 |
| `101` | `점수 확인` | 유효 범위 위 |

모든 경우 출력은 한 줄입니다. 확인 후 기본 점수 `72`로 되돌립니다.

## 문제 3. 신호에 따른 안내

참고 코드: [signal/main.go](signal/main.go) · [출력 테스트](signal/main_test.go)

개인 파일: `_practice/03-control-flow/signal/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	signal := "yellow"

	switch signal {
	case "red":
		fmt.Println("정지")
	case "yellow":
		fmt.Println("주의")
	case "green":
		fmt.Println("진행")
	default:
		fmt.Println("신호 확인")
	}
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/03-control-flow/signal
go test -v ./solutions/03-control-flow/signal
```

출력:

```text
주의
```

이 `switch`는 `signal`의 값과 일치하는 `case`를 선택합니다. `fallthrough`를 사용하지 않으므로 선택한 본문을 마친 뒤 다음 `case`의 본문이 자동으로 이어지지 않습니다. 각 경우에 `break`를 덧붙이지 않아도 한 문장만 출력됩니다. [switch 문 명세](https://go.dev/ref/spec#Switch_statements)를 참고하세요.

개인 파일에서 다음 네 입력을 확인합니다.

| 입력 | 출력 |
| --- | --- |
| `"red"` | `정지` |
| `"yellow"` | `주의` |
| `"green"` | `진행` |
| `"blue"` | `신호 확인` |

`default`는 위 세 값과 일치하지 않을 때 실행됩니다. 이 문제는 문자열을 정확히 비교하며 대소문자를 자동으로 바꾸지 않습니다. 확인 후 `"yellow"`로 되돌립니다.

## 문제 4. 반복의 마지막 값 고치기

개인 파일: `_practice/03-control-flow/repair/main.go`

수정 전 프로그램의 실제 출력은 `합계: 10`입니다. `number < limit`일 때 5는 포함되지 않으므로 1+2+3+4만 계산됩니다. 컴파일과 실행이 성공해도 문제의 요구 사항을 만족하는지 따로 확인해야 합니다.

### 끝값을 포함하는 반복

반복 조건을 `number <= limit`로 바꿉니다. 수정 후 전체 코드:

```go
package main

import "fmt"

func main() {
	limit := 5
	total := 0
	for number := 1; number <= limit; number++ {
		total += number
	}
	fmt.Printf("합계: %d\n", total)
}
```

실행:

```text
go run ./_practice/03-control-flow/repair
```

출력:

```text
합계: 15
```

### 정수 `range`로 같은 합계 구하기

같은 개인 파일의 전체 코드를 다음 방식으로 바꿀 수도 있습니다. 이전 코드를 보관하려면 다른 폴더를 사용합니다.

```go
package main

import "fmt"

func main() {
	limit := 5
	total := 0
	for index := range limit {
		total += index + 1
	}
	fmt.Printf("합계: %d\n", total)
}
```

실행 명령과 기본 출력은 앞의 방식과 같습니다. `limit`가 5이면 `index`는 0, 1, 2, 3, 4이므로, 1을 더한 값을 누적해야 1부터 5까지 더하게 됩니다. `index` 자체를 더하면 다시 합계 10이 됩니다.

정수 `range`의 범위는 반복 시작 전에 정해지며, 이 문제에서는 양의 `limit`에 대해 0부터 `limit-1`까지 진행합니다. `limit`가 0이면 본문은 한 번도 실행하지 않습니다. [for 문 명세](https://go.dev/ref/spec#For_statements)를 참고하세요.

### 두 방식의 경계 확인

두 코드 각각에서 `limit`만 바꾸어 확인합니다.

| `limit` | 더하는 값 | 출력 |
| --- | --- | --- |
| `0` | 없음 | `합계: 0` |
| `1` | `1` | `합계: 1` |
| `5` | `1+2+3+4+5` | `합계: 15` |
| `10` | `1`부터 `10`까지 | `합계: 55` |

합계 변수는 반복문 밖에서 0으로 초기화하므로, 반복하지 않는 경우에도 0을 출력할 수 있습니다. 확인 후 기본 값 `5`로 되돌립니다.

## 문제 5. 3의 배수만 합산하기

참고 코드: [multiples/main.go](multiples/main.go) · [출력 테스트](multiples/main_test.go)

개인 파일: `_practice/03-control-flow/multiples/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	limit := 10
	total := 0
	count := 0

	for number := 1; number <= limit; number++ {
		if number%3 != 0 {
			continue
		}
		total += number
		count++
	}

	fmt.Printf("3의 배수: %d개\n합계: %d\n", count, total)
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/03-control-flow/multiples
go test -v ./solutions/03-control-flow/multiples
```

출력:

```text
3의 배수: 3개
합계: 18
```

1~10 중 누적하는 값은 3, 6, 9입니다. 다른 값에서는 `continue`를 만나 합계와 개수를 바꾸지 않습니다. 그 후 `number++`와 다음 조건 검사로 진행합니다.

`total`과 `count`는 반복 밖에서 한 번만 초기화합니다. 매 반복의 시작에서 다시 0으로 만들면 지금까지의 결과가 사라집니다. `limit`가 0이면 1부터 시작하는 조건이 처음부터 거짓이므로 결과는 0개·합계 0입니다. 수학적으로 0도 3의 배수지만, 이 문제에서 순회하는 범위에는 0이 포함되지 않습니다.

### 추가 입력

개인 파일의 `limit`를 바꾸어 `go run ./_practice/03-control-flow/multiples`로 확인합니다. 입력은 문제에서 정한 0~100 범위입니다.

| `limit` | 출력 첫 줄 | 출력 둘째 줄 |
| --- | --- | --- |
| `0` | `3의 배수: 0개` | `합계: 0` |
| `2` | `3의 배수: 0개` | `합계: 0` |
| `3` | `3의 배수: 1개` | `합계: 3` |
| `9` | `3의 배수: 3개` | `합계: 18` |
| `10` | `3의 배수: 3개` | `합계: 18` |

확인 후 기본 값 `10`으로 되돌립니다.

## 문제 6. 목표 학습 시간 채우기

참고 코드: [goal/main.go](goal/main.go) · [출력 테스트](goal/main_test.go)

개인 파일: `_practice/03-control-flow/goal/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	targetMinutes := 100
	dailyMinutes := 30

	if targetMinutes < 0 || dailyMinutes <= 0 {
		fmt.Println("입력 확인")
	} else {
		days := 0
		total := 0
		for total < targetMinutes {
			total += dailyMinutes
			days++
		}
		fmt.Printf("%d일, 누적 %d분\n", days, total)
	}
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/03-control-flow/goal
go test -v ./solutions/03-control-flow/goal
```

출력:

```text
4일, 누적 120분
```

### 누적 과정과 종료 조건

| 처리한 일수 | 누적 시간 | 다음 반복의 조건 `total < 100` |
| --- | --- | --- |
| `0` | `0` | 참 |
| `1` | `30` | 참 |
| `2` | `60` | 참 |
| `3` | `90` | 참 |
| `4` | `120` | 거짓 |

하루 단위로 학습 시간을 더하므로 목표 100분을 넘은 120분에서 멈춥니다. 최종 누적 시간을 100으로 강제로 바꾸지 않습니다.

목표가 90이면 3일째의 90분에서 이미 달성했으므로 더 반복하지 않습니다. 조건을 `total <= targetMinutes`로 쓰면 목표와 같은 순간에도 한 번 더 실행하게 됩니다. 목표가 0이면 처음의 `0 < 0`이 거짓이므로 0일·누적 0분입니다. `for`는 본문을 실행하기 전에 조건을 검사합니다. [for 문 명세](https://go.dev/ref/spec#For_statements)를 참고하세요.

### 반복 전에 입력을 확인하는 이유

하루 학습 시간이 0이면 양의 목표를 향해 누적 시간이 증가하지 않습니다. 음수라면 목표와 반대 방향으로 움직입니다. 이 문제는 목표가 음수이거나 하루 학습 시간이 0 이하인 경우 `입력 확인`을 출력하고 반복이 있는 `else`에 들어가지 않습니다.

문제에서 사용하는 작은 유효 입력에서는 하루 학습 시간이 양수이므로 반복할 때마다 누적 시간이 증가하고, 결국 목표 이상이 되어 조건이 거짓이 됩니다. **조건을 확인하는 것과 그 조건을 바꿀 값을 갱신하는 것**을 함께 살펴야 반복의 종료를 설명할 수 있습니다.

개인 파일에서 아래 입력을 확인합니다.

| 목표 시간 | 하루 학습 시간 | 출력 |
| --- | --- | --- |
| `100` | `30` | `4일, 누적 120분` |
| `90` | `30` | `3일, 누적 90분` |
| `0` | `30` | `0일, 누적 0분` |
| `-1` | `30` | `입력 확인` |
| `100` | `0` | `입력 확인` |
| `100` | `-1` | `입력 확인` |

잘못된 입력에서는 안내 한 줄만 출력되고 누적 결과는 출력되지 않습니다. 확인 후 기본 목표 `100`과 하루 학습 시간 `30`으로 되돌립니다.

## 선택 A. FizzBuzz

개인 파일: `_practice/03-control-flow/fizzbuzz/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	for number := 1; number <= 15; number++ {
		switch {
		case number%15 == 0:
			fmt.Println("FizzBuzz")
		case number%3 == 0:
			fmt.Println("Fizz")
		case number%5 == 0:
			fmt.Println("Buzz")
		default:
			fmt.Println(number)
		}
	}
}
```

실행:

```text
go run ./_practice/03-control-flow/fizzbuzz
```

출력:

```text
1
2
Fizz
4
Buzz
Fizz
7
8
Fizz
Buzz
11
Fizz
13
14
FizzBuzz
```

비교할 값을 생략한 `switch`는 참인 조건을 찾는 방식으로 사용할 수 있습니다. 15의 배수는 3과 5의 배수이므로 먼저 검사합니다. 3의 배수를 먼저 검사하면 15에서 `Fizz` 분기가 선택되어 `FizzBuzz`에 도달하지 못합니다.

각 숫자에서 하나의 `case`만 실행합니다. `switch`를 마친 뒤에는 바깥 `for`가 다음 숫자로 진행하므로 출력은 총 15줄입니다. [switch 문 명세](https://go.dev/ref/spec#Switch_statements)를 참고하세요.

## 검증 범위 구분하기

참고 구현 네 개의 기본 출력은 다음으로 확인합니다.

```text
go test -v ./solutions/03-control-flow/...
```

각 `Example_main`은 참고 구현에 적힌 기본 입력을 검증합니다. 개인 코드와 경계 입력은 문제 문서의 실행 명령으로 따로 확인합니다. 기본 출력 하나만 맞는지 확인하는 데서 멈추지 않고, 분기가 바뀌는 값과 반복 횟수가 0이 되는 값도 비교하세요.

| 문제로 돌아가기 | 교재 | 해설 목록 |
| --- | --- | --- |
| [03장 문제](../../exercises/03-control-flow/README.md) | [03장 본문](../../book/03-control-flow/README.md) | [전체 해설](../README.md) |
