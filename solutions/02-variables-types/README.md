# 02장 해설: 변수·상수·자료형·연산자

문제: [02장 연습문제](../../exercises/02-variables-types/README.md)

문제를 직접 풀고 힌트를 확인한 뒤 해설을 읽습니다. 이 장의 기본 풀이는 값을 선언하고, 대입하고, 계산하고, 출력하는 순서로 구성합니다. `main` 외의 함수나 조건문·반복문은 필요하지 않습니다.

모든 명령은 **저장소 루트**에서 실행합니다. `solutions/`의 참고 구현을 실행하는 것과 `_practice/`의 개인 풀이를 실행하는 것을 구분하세요. 참고 구현의 테스트가 통과했다고 개인 코드까지 검사된 것은 아닙니다.

## 문제 1. 값의 변화 예측

예상 출력은 다음과 같습니다.

```text
Go: 5, 2
```

실행해서 확인하고 싶다면 편집기에서 `_practice/02-variables-types/predict/main.go`를 준비하고 아래 전체 코드를 작성합니다. 기존 개인 파일이 있다면 내용을 보관한 뒤 진행하세요.

```go
package main

import "fmt"

func main() {
	var count int
	label := "Go"
	next := count + 2
	count = 5
	fmt.Printf("%s: %d, %d\n", label, count, next)
}
```

실행 명령:

```text
go run ./_practice/02-variables-types/predict
```

| 실행한 문장 | `count` | `next` | 이유 |
| --- | --- | --- | --- |
| `var count int` | `0` | 아직 선언되지 않음 | 초기값을 쓰지 않은 `int`의 영값은 0 |
| `next := count + 2` | `0` | `2` | 현재 값 0에 2를 더한 결과를 저장 |
| `count = 5` | `5` | `2` | `count`에만 새 값을 대입 |

`next`에는 계산식 자체가 아니라 계산 결과인 정수 값이 저장됩니다. 따라서 `count`가 나중에 바뀌어도 `next`의 값은 자동으로 다시 계산되지 않습니다. 두 값의 관계를 다시 반영하려면 그 시점에 계산과 대입을 다시 해야 합니다. 이 설명은 여기서 다루는 정수 값의 대입에 관한 것입니다. [변수의 초기값](https://go.dev/ref/spec#The_zero_value)과 [대입](https://go.dev/ref/spec#Assignment_statements)의 명세를 참고하세요.

자료형은 `label`이 `string`, `count`와 `next`가 `int`입니다. `:=`를 사용했다고 자료형이 없는 변수가 되는 것은 아닙니다.

## 문제 2. 학습 시간 기록

참고 코드: [study-log/main.go](study-log/main.go) · [출력 테스트](study-log/main_test.go)

개인 풀이 파일: `_practice/02-variables-types/study-log/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	var course string = "Go"
	days := 2
	days += 1
	minutesPerDay := 25
	totalMinutes := days * minutesPerDay
	ready := true

	fmt.Printf("%s 학습\n", course)
	fmt.Printf("%d일, 하루 %d분\n", days, minutesPerDay)
	fmt.Printf("총 %d분\n", totalMinutes)
	fmt.Printf("준비: %t\n", ready)
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/02-variables-types/study-log
go test -v ./solutions/02-variables-types/study-log
```

프로그램 출력:

```text
Go 학습
3일, 하루 25분
총 75분
준비: true
```

`course`는 자료형을 명시했고, 나머지 변수는 초기값으로 자료형을 정했습니다. `days += 1`은 이미 선언한 변수의 값에 1을 더해 다시 저장합니다. 그 결과 `days`가 3이 된 뒤 전체 시간을 계산하므로 `totalMinutes`는 75입니다.

만약 전체 시간을 `days` 변경 전에 계산했다면 50이 저장됩니다. 일수만 바꾼다고 미리 계산한 값도 바뀌지는 않습니다. 문제 1의 대입 동작이 이 계산 순서에도 적용됩니다.

`ready`는 문자열 `"true"`가 아니라 불리언 값 `true`입니다. 출력 서식 `%t`는 불리언 값을 표시합니다. [fmt의 출력 서식](https://pkg.go.dev/fmt#hdr-Printing)을 참고하세요.

## 문제 3. 초를 시·분·초로 나누기

참고 코드: [time-parts/main.go](time-parts/main.go) · [출력 테스트](time-parts/main_test.go)

개인 풀이 파일: `_practice/02-variables-types/time-parts/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	const secondsPerMinute = 60
	const minutesPerHour = 60
	totalSeconds := 9865

	hours := totalSeconds / (secondsPerMinute * minutesPerHour)
	minutes := totalSeconds / secondsPerMinute % minutesPerHour
	seconds := totalSeconds % secondsPerMinute

	fmt.Printf("%d초 = %d시간 %d분 %d초\n", totalSeconds, hours, minutes, seconds)
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/02-variables-types/time-parts
go test -v ./solutions/02-variables-types/time-parts
```

프로그램 출력:

```text
9865초 = 2시간 44분 25초
```

### 계산을 나누어 읽기

| 값 | 계산 | 결과 |
| --- | --- | --- |
| 1시간의 초 수 | `60 * 60` | `3600` |
| 온전한 시간 수 | `9865 / 3600` | `2` |
| 전체 분 수 | `9865 / 60` | `164` |
| 시간으로 묶고 남은 분 | `164 % 60` | `44` |
| 분으로 묶고 남은 초 | `9865 % 60` | `25` |

`minutes`의 식은 왼쪽부터 나눗셈을 한 뒤 나머지를 구하므로 `(totalSeconds / secondsPerMinute) % minutesPerHour`로 읽을 수 있습니다. `164`를 그대로 분으로 출력하면 시간과 중복되는 부분이 포함되므로 나머지를 구합니다.

정수 나눗셈은 정수 몫을 구하며 소수 부분을 버립니다. 이 문제의 0 이상 입력에서는 온전한 단위의 개수가 됩니다. `%`는 나누고 남는 값을 구합니다. 정확한 규칙은 [산술 연산자 명세](https://go.dev/ref/spec#Arithmetic_operators)에 있습니다.

결과를 반대 방향으로 계산해 `2 * 3600 + 44 * 60 + 25 = 9865`가 되는지도 확인할 수 있습니다. 단위가 바뀌어도 전체 시간의 양이 유지되는지 보는 방법입니다.

### 추가 입력의 결과

개인 파일의 `totalSeconds`만 바꾸고 `go run ./_practice/02-variables-types/time-parts`로 실행합니다.

| 입력 | 출력 |
| --- | --- |
| `0` | `0초 = 0시간 0분 0초` |
| `59` | `59초 = 0시간 0분 59초` |
| `60` | `60초 = 0시간 1분 0초` |
| `3600` | `3600초 = 1시간 0분 0초` |
| `3661` | `3661초 = 1시간 1분 1초` |

59→60, 3600→3661처럼 단위가 바뀌는 부근을 확인하면 몫과 나머지를 올바르게 조합했는지 점검하기 좋습니다. 기본 출력 테스트는 `9865`에 대한 결과를 확인하며, 위의 변형 입력은 개인 파일에서 따로 실행합니다.

## 문제 4. 다른 정수형의 덧셈 고치기

개인 파일: `_practice/02-variables-types/repair/main.go`

`apples`는 `int`, `oranges`는 `int64`입니다. 같은 정수 계열이어도 서로 다른 자료형의 변수이므로 그대로 더할 수 없습니다. 선언은 유지하고 `apples`의 값을 `int64`로 바꾸어 더합니다.

수정 후 전체 코드:

```go
package main

import "fmt"

func main() {
	var apples int = 4
	var oranges int64 = 3
	fmt.Println("합계:", int64(apples)+oranges)
}
```

실행:

```text
go run ./_practice/02-variables-types/repair
```

출력:

```text
합계: 7
```

`int64(apples)`라는 표현식의 값이 `int64`일 뿐, 원래 `apples` 변수의 자료형이 바뀌는 것은 아닙니다. `apples`는 계속 `int`입니다. 이 문제의 값 4는 `int64`로 변환해 그대로 표현할 수 있습니다. 명시적 변환 규칙은 [Go의 변환 명세](https://go.dev/ref/spec#Conversions)를 참고하세요.

오류를 읽을 때는 파일과 위치 다음에 표시된 자료형 이름을 확인합니다. 수정 전에는 계산을 출력하기 전에 컴파일이 실패하며, 수정 후에는 같은 실행 명령으로 합계가 출력됩니다. 숫자를 문자열로 바꾸어 연결하는 것은 덧셈 요구 사항을 해결하지 못합니다.

## 문제 5. 소수 부분을 포함한 평균

참고 코드: [average/main.go](average/main.go) · [출력 테스트](average/main_test.go)

개인 풀이 파일: `_practice/02-variables-types/average/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	total := 245
	count := 4
	average := float64(total) / float64(count)

	fmt.Printf("평균: %.2f\n", average)
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/02-variables-types/average
go test -v ./solutions/02-variables-types/average
```

프로그램 출력:

```text
평균: 61.25
```

### 변환은 나눗셈 전에

| 표현식 | 계산 순서 | 결과 |
| --- | --- | --- |
| `total / count` | `int` 값끼리 나눔 | 정수 `61` |
| `float64(total / count)` | 정수 몫 61을 구한 뒤 변환 | `float64` 값 61, 두 자리로 표시하면 `61.00` |
| `float64(total) / float64(count)` | 두 값을 변환한 뒤 나눔 | `61.25` |

이미 버린 소수 부분은 나중에 자료형을 바꾸어도 복구되지 않습니다. 두 값 중 하나만 변환하면 다른 쪽은 여전히 `int`이므로, 이 식에서는 두 피연산자를 모두 `float64`로 맞춥니다. [정수 나눗셈](https://go.dev/ref/spec#Arithmetic_operators)과 [숫자 변환](https://go.dev/ref/spec#Conversions)의 순서를 함께 생각하세요.

### 계산한 값과 표시한 값

`%.2f`는 출력할 때 소수점 아래 두 자리로 표시하는 형식입니다. 이 명령이 `average` 변수 자체를 소수 두 자리의 값으로 다시 대입하는 것은 아닙니다. `total = 10`, `count = 3`에서 `float64(total) / float64(count)`로 계산한 값은 근삿값이며, `%.2f`로 `3.33`을 표시합니다. `total = 5`, `count = 2`에서는 같은 변환식으로 계산한 `2.5`를 `2.50`으로 표시합니다. [fmt의 정밀도 설명](https://pkg.go.dev/fmt#hdr-Printing)을 참고하세요.

개인 파일의 숫자를 바꾸어 `go run ./_practice/02-variables-types/average`로 확인합니다.

| `total` | `count` | 출력 |
| --- | --- | --- |
| `0` | `4` | `평균: 0.00` |
| `10` | `3` | `평균: 3.33` |
| `5` | `2` | `평균: 2.50` |

이 문제는 `count`가 양수라는 조건입니다. 입력 검증이나 0으로 나누는 상황은 이번 풀이에 포함하지 않습니다. 확인 후 기본 값 `245`, `4`로 되돌립니다.

## 문제 6. 구매 금액과 예산 비교

참고 코드: [receipt/main.go](receipt/main.go) · [출력 테스트](receipt/main_test.go)

개인 풀이 파일: `_practice/02-variables-types/receipt/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	const unitPrice = 1200
	const shippingFee = 2500
	quantity := 3
	discount := 500
	budget := 6000

	subtotal := unitPrice * quantity
	total := subtotal - discount + shippingFee
	withinBudget := total <= budget

	fmt.Printf("상품 금액: %d원\n", subtotal)
	fmt.Printf("할인: %d원\n", discount)
	fmt.Printf("배송비: %d원\n", shippingFee)
	fmt.Printf("합계: %d원\n", total)
	fmt.Printf("예산 이내: %t\n", withinBudget)
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/02-variables-types/receipt
go test -v ./solutions/02-variables-types/receipt
```

프로그램 출력:

```text
상품 금액: 3600원
할인: 500원
배송비: 2500원
합계: 5600원
예산 이내: true
```

계산은 `1200 * 3 = 3600`, `3600 - 500 + 2500 = 5600` 순서입니다. 비교 `5600 <= 6000`의 결과는 `true`이며, 이 불리언 값을 `withinBudget`에 저장합니다. 비교 결과를 저장하기 위해 조건문을 만들 필요는 없습니다. [비교 연산자의 정의](https://go.dev/ref/spec#Comparison_operators)를 참고하세요.

이 실습은 금액을 정수 원 단위로 표현하므로 부동소수점 변환이 필요하지 않습니다. 할인액은 주어진 500원, 배송비는 항상 2500원이라는 문제의 조건을 그대로 사용했습니다. 실제 상점의 할인·배송 정책을 구현한 예제가 아닙니다.

### 수량을 바꾸면

개인 파일의 `quantity`를 `4`로 바꾸고 `go run ./_practice/02-variables-types/receipt`를 실행하면 다음과 같습니다.

```text
상품 금액: 4800원
할인: 500원
배송비: 2500원
합계: 6800원
예산 이내: false
```

같은 계산식에 다른 수량을 사용했기 때문에 상품 금액과 합계가 바뀌고, 그 합계를 사용한 비교 결과도 바뀝니다. 이는 **프로그램을 다시 실행해 계산했기 때문**입니다. 이미 계산한 변수들이 다른 변수와 자동으로 연결되어 있다는 뜻은 아닙니다.

### 합계와 예산이 같을 때

개인 파일에서 수량을 `3`, 예산을 `5600`으로 바꾸어 같은 명령으로 실행합니다. 합계는 `5600원`이므로 `합계: 5600원`, `예산 이내: true`가 출력되어야 합니다. “이하”에는 같은 값이 포함되므로 `5600 <= 5600`은 `true`입니다. `<`를 사용했다면 이 경계에서 `false`가 나옵니다.

확인한 뒤 **기본 수량 `3`과 예산 `6000`으로 되돌립니다.**

## 선택 A. `iota`로 상태 번호 만들기

개인 파일: `_practice/02-variables-types/iota/main.go`

전체 코드:

```go
package main

import "fmt"

func main() {
	const (
		statusQueued = iota
		statusRunning
		statusDone
	)

	fmt.Printf("대기=%d 진행=%d 완료=%d\n", statusQueued, statusRunning, statusDone)
}
```

실행:

```text
go run ./_practice/02-variables-types/iota
```

출력:

```text
대기=0 진행=1 완료=2
```

`iota`는 이 상수 선언 블록에서 0부터 시작하고, 다음 상수 선언으로 진행하면서 증가합니다. 뒤의 두 선언에서 식을 생략하면 앞선 선언의 식을 반복하며 각 선언에 해당하는 `iota` 값을 사용합니다. `statusQueued`, `statusRunning`, `statusDone`은 실행 중 바꾸는 변수가 아니라 상수입니다. 정확한 규칙은 [iota 명세](https://go.dev/ref/spec#Iota)를 참고합니다.

여기서는 이름이 있는 연속 번호를 선언하는 방법만 확인합니다. 실제 상태에 따라 다른 동작을 수행하는 코드는 이후 조건문에서 다룹니다.

## 검증 결과를 읽는 방법

참고 구현 네 개의 기본 출력을 한 번에 확인하려면 다음을 실행합니다.

```text
go test -v ./solutions/02-variables-types/...
```

각 `Example_main`은 해당 참고 구현에 적힌 **기본 입력의 출력**을 확인합니다. 개인 파일과 문제에 제시한 변형 입력은 문제 문서의 명령으로 따로 실행하고 비교합니다. 테스트를 통과한 범위와 자신이 확인한 입력을 구분해서 기록하세요.

| 문제로 돌아가기 | 교재 | 해설 목록 |
| --- | --- | --- |
| [02장 문제](../../exercises/02-variables-types/README.md) | [02장 본문](../../book/02-variables-types/README.md) | [전체 해설](../README.md) |
