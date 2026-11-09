# 02장 연습문제: 변수·상수·자료형·연산자

선수 학습: [02. 변수·상수·자료형·연산자](../../book/02-variables-types/README.md)

이번 문제에서는 값을 변수에 보관하고, 필요한 값을 계산하고, 계산 결과를 알맞은 형식으로 출력합니다. 입력값은 코드에 직접 적습니다. 새 함수, 조건문, 반복문, 키보드 입력이나 외부 패키지는 필요하지 않습니다.

**문제 1~6이 기본 과정**이며, 선택 A는 `iota`를 사용하는 추가 연습입니다. 모든 명령은 **저장소 루트**에서 실행합니다. 개인 파일은 `_practice/02-variables-types/` 아래에 두고, 새 `go.mod`나 `go.work`를 만들지 않습니다.

## 문제와 개인 작업 폴더

| 문제 | 수준 | 작업 폴더 |
| --- | --- | --- |
| 1. 값의 변화 예측 | 확인 | 먼저 학습 노트에 결과 기록 |
| 2. 학습 시간 기록 | 변형 | `_practice/02-variables-types/study-log/` |
| 3. 초를 시·분·초로 나누기 | 변형 | `_practice/02-variables-types/time-parts/` |
| 4. 다른 정수형의 덧셈 고치기 | 오류 진단 | `_practice/02-variables-types/repair/` |
| 5. 소수 부분을 포함한 평균 | 응용 | `_practice/02-variables-types/average/` |
| 6. 구매 금액과 예산 비교 | 응용 | `_practice/02-variables-types/receipt/` |
| 선택 A. 상태 번호 만들기 | 선택 | `_practice/02-variables-types/iota/` |

각 폴더의 파일 이름은 `main.go`입니다. 서로 다른 문제의 `main` 함수를 같은 폴더에 모으지 않습니다.

## 준비: 개인 파일 만들기

자신의 터미널에 맞는 블록 하나를 실행합니다. 기본 문제의 다섯 폴더에 01장의 첫 프로그램을 복사합니다. **기존 `main.go`가 있으면 보존**하며, 테스트 파일은 복사하지 않습니다. 이미 파일이 있다는 안내가 나오면 기존 내용을 확인하고 이어서 작성하세요.

Windows PowerShell:

```powershell
foreach ($practiceName in @("study-log", "time-parts", "repair", "average", "receipt")) {
    $practiceDir = Join-Path "_practice/02-variables-types" $practiceName
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
for practice_name in study-log time-parts repair average receipt; do
    practice_dir="_practice/02-variables-types/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

여기서 사용한 반복문은 **파일을 준비하는 셸 문법**입니다. 이번 장의 Go 풀이에는 반복문을 사용하지 않습니다. 개인 폴더의 공통 규칙은 [연습문제 안내](../README.md)를 참고하세요.

## 문제 1. 값은 언제 계산될까? — 확인

다음은 `main` 함수 안의 **발췌 코드**입니다. 실행하기 전에 출력 한 줄을 예상하고, 그 이유를 적으세요.

```go
var count int
label := "Go"
next := count + 2
count = 5
fmt.Printf("%s: %d, %d\n", label, count, next)
```

### 설명할 내용

1. 처음 선언한 `count`에는 어떤 값이 들어 있는가?
2. `next := count + 2`를 실행할 때 `next`에 어떤 값이 저장되는가?
3. `count = 5`가 실행되면 `next`도 자동으로 바뀌는가?
4. `label`, `count`, `next`의 자료형은 각각 무엇인가?

<details>
<summary>힌트</summary>

한 줄씩 실행했을 때의 값을 표로 적어 보세요. 대입은 그 시점의 계산 결과를 저장합니다. 스프레드시트의 셀 수식처럼 다른 값이 바뀔 때마다 이전 식을 다시 계산한다고 가정하지 마세요.

</details>

완료 조건: 출력뿐 아니라 `count`를 바꾼 뒤 `next`가 어떻게 되는지 설명할 수 있습니다. 정답은 모든 문제 뒤의 해설 링크에서 확인합니다.

## 문제 2. 학습 시간 기록 — 변형

작업 파일: `_practice/02-variables-types/study-log/main.go`

다음 값을 변수로 관리하여 학습 기록을 출력하세요.

| 이름 | 시작 값과 조건 |
| --- | --- |
| `course` | `var`로 선언한 `string`, 값은 `"Go"` |
| `days` | `:=`로 선언한 정수 `2`; 이후 `+=`로 `1` 증가 |
| `minutesPerDay` | `:=`로 선언한 정수 `25` |
| `totalMinutes` | 증가시킨 학습 일수와 하루 학습 시간을 곱한 값 |
| `ready` | `:=`로 선언한 불리언 `true` |

`totalMinutes`는 `days`를 증가시킨 **뒤에** 계산합니다. 다음 네 줄을 정확히 출력하세요. 문자열에는 `%s`, 정수에는 `%d`, 불리언에는 `%t`를 사용합니다.

```text
Go 학습
3일, 하루 25분
총 75분
준비: true
```

실행:

```text
go run ./_practice/02-variables-types/study-log
```

<details>
<summary>힌트</summary>

이미 선언한 `days`를 다시 선언하지 않고 값을 변경합니다. 전체 학습 시간을 먼저 계산한 뒤 일수만 바꾸면 계산해 둔 시간은 자동으로 바뀌지 않습니다. 문제 1에서 확인한 대입의 의미를 적용하세요.

</details>

완료 조건: 예상한 네 줄을 출력하고, 선언과 값 변경을 구분하며, 계산 순서를 설명할 수 있습니다.

## 문제 3. 초를 시·분·초로 나누기 — 변형

작업 파일: `_practice/02-variables-types/time-parts/main.go`

다음 두 상수를 선언하세요.

- `secondsPerMinute`: 1분에 해당하는 초 수 `60`
- `minutesPerHour`: 1시간에 해당하는 분 수 `60`

변수 `totalSeconds`에는 정수 `9865`를 저장합니다. 정수 나눗셈 `/`와 나머지 `%`를 이용해 `hours`, `minutes`, `seconds`를 구하고 다음을 출력하세요.

```text
9865초 = 2시간 44분 25초
```

### 조건

- 초 수는 0 이상의 정수이며, 이 문제에 제시한 작은 값들을 대상으로 합니다.
- `hours`에는 전체 시간 수를 저장합니다.
- `minutes`와 `seconds`는 각각 0~59 사이의 나머지 단위입니다.
- 계산에는 선언한 두 상수를 사용합니다. 문자열로 숫자를 나누거나 부동소수점으로 변환할 필요는 없습니다.

실행:

```text
go run ./_practice/02-variables-types/time-parts
```

### 입력을 바꾸어 확인하기

`totalSeconds`의 숫자만 아래 값으로 바꾸어 한 번씩 실행합니다. 반복문을 작성할 필요는 없습니다.

| `totalSeconds` | 예상 출력 |
| --- | --- |
| `0` | `0초 = 0시간 0분 0초` |
| `59` | `59초 = 0시간 0분 59초` |
| `60` | `60초 = 0시간 1분 0초` |
| `3600` | `3600초 = 1시간 0분 0초` |
| `3661` | `3661초 = 1시간 1분 1초` |

확인한 뒤 기본 값 `9865`로 되돌립니다.

<details>
<summary>힌트 1: 시간과 남은 초</summary>

1시간에 해당하는 초 수는 두 상수를 곱해 구합니다. 전체 초를 그 값으로 나누면 온전한 시간 수가 됩니다. 남은 초는 1분에 해당하는 초 수로 나눈 나머지입니다.

</details>

<details>
<summary>힌트 2: 분</summary>

전체 초를 60으로 나누면 전체 분 수가 됩니다. 그중 시간으로 이미 묶인 부분을 빼고 0~59 범위의 분을 얻으려면 다시 나머지를 사용합니다.

</details>

완료 조건: 기본 입력과 추가 입력에서 시간·분·초가 모두 일치하고, `/`와 `%`가 각각 무엇을 구하는지 설명할 수 있습니다.

## 문제 4. 다른 정수형의 덧셈 고치기 — 오류 진단

작업 파일: `_practice/02-variables-types/repair/main.go`

아래는 **의도적으로 컴파일에 실패하는 전체 코드**입니다. 교재의 정상 예제 파일은 바꾸지 말고 개인 실습 파일에서 확인합니다.

```go
package main

import "fmt"

func main() {
	var apples int = 4
	var oranges int64 = 3
	fmt.Println("합계:", apples+oranges)
}
```

실행:

```text
go run ./_practice/02-variables-types/repair
```

### 수행할 작업

1. 오류 메시지에서 서로 다른 두 자료형을 찾습니다.
2. `apples`와 `oranges`의 선언은 유지합니다.
3. 덧셈하는 위치에서 `apples` 값을 `int64`로 명시적으로 변환합니다.
4. 다시 실행해 다음 출력이 나오는지 확인합니다.

```text
합계: 7
```

<details>
<summary>힌트</summary>

자료형 이름 뒤에 괄호로 값을 감싸면 명시적으로 변환할 수 있습니다. 두 변수가 정수 계열이라는 이유만으로 다른 자료형 사이의 덧셈을 자동으로 맞춰 주지는 않습니다.

</details>

완료 조건: 수정 전 실패와 수정 후 성공을 확인하고, 변환한 것은 값이지 원래 변수의 선언 자료형이 아니라는 점을 설명할 수 있습니다.

## 문제 5. 소수 부분을 포함한 평균 — 응용

작업 파일: `_practice/02-variables-types/average/main.go`

정수 변수 `total`에 `245`, `count`에 `4`를 저장합니다. `count`는 **양수**라는 조건으로 계산하며, 이 문제에서는 0으로 나누는 입력을 다루지 않습니다.

나누는 두 값을 먼저 `float64`로 변환하여 `average`를 계산한 뒤, 소수점 아래 두 자리로 출력하세요.

```text
평균: 61.25
```

실행:

```text
go run ./_practice/02-variables-types/average
```

### 추가 확인

| `total` | `count` | 예상 출력 |
| --- | --- | --- |
| `0` | `4` | `평균: 0.00` |
| `10` | `3` | `평균: 3.33` |
| `5` | `2` | `평균: 2.50` |

두 변수의 숫자를 바꾸어 하나씩 확인한 뒤 기본 값 `245`, `4`로 되돌립니다. 계산 결과와 화면에 표시하는 소수 자릿수가 어떻게 다른지도 설명하세요.

<details>
<summary>힌트 1: 변환 순서</summary>

정수끼리 먼저 나눈 결과를 `float64`로 바꾸면 이미 버린 소수 부분을 되살릴 수 없습니다. 나눗셈이 일어나기 전에 두 피연산자를 변환합니다.

</details>

<details>
<summary>힌트 2: 표시 형식</summary>

`fmt.Printf`의 `%.2f`는 부동소수점 값을 소수점 아래 두 자리로 표시할 때 사용합니다. 표시 형식을 정하는 것과 변수에 저장한 값을 변경하는 것은 다릅니다. [fmt의 출력 서식](https://pkg.go.dev/fmt#hdr-Printing)을 참고하세요.

</details>

완료 조건: 소수 부분이 유지되고, 추가 입력의 표시 결과가 맞으며, `float64(total/count)`와 다른 이유를 설명할 수 있습니다.

## 문제 6. 구매 금액과 예산 비교 — 응용

작업 파일: `_practice/02-variables-types/receipt/main.go`

다음 조건으로 작은 구매 내역을 계산합니다. 모든 금액은 **정수 원 단위**입니다. 할인은 주어진 금액을 그대로 빼고, 배송비는 항상 더합니다. 할인 조건이나 무료 배송 규칙을 추가하지 않습니다.

| 이름 | 선언과 값 |
| --- | --- |
| `unitPrice` | 상수, 상품 하나의 가격 `1200` |
| `shippingFee` | 상수, 배송비 `2500` |
| `quantity` | 변수, 수량 `3` |
| `discount` | 변수, 할인액 `500` |
| `budget` | 변수, 예산 `6000` |

`subtotal`에 상품 금액, `total`에 할인과 배송비를 반영한 합계를 저장합니다. `withinBudget`에는 합계가 예산 이하인지 비교한 불리언 값을 저장합니다.

다음 다섯 줄을 출력하세요.

```text
상품 금액: 3600원
할인: 500원
배송비: 2500원
합계: 5600원
예산 이내: true
```

실행:

```text
go run ./_practice/02-variables-types/receipt
```

수량을 `4`로 바꾸고 다시 실행하세요. 상품 금액은 `4800원`, 합계는 `6800원`, 예산 이내 여부는 `false`가 되어야 합니다. 할인과 배송비는 그대로입니다.

이어서 수량을 `3`, 예산을 `5600`으로 바꾸어 합계와 예산이 정확히 같은 경우도 확인하세요. 합계는 `5600원`, 예산 이내 여부는 `true`가 되어야 합니다. 확인한 뒤 **기본 수량 `3`과 예산 `6000`으로 되돌립니다.**

<details>
<summary>힌트 1: 계산 순서</summary>

단가와 수량을 곱한 뒤 할인액을 빼고 배송비를 더합니다. 각 중간 결과를 이름이 있는 변수로 두면 어떤 값을 출력하는지 확인하기 쉽습니다.

</details>

<details>
<summary>힌트 2: 예산 비교</summary>

“예산 이하”는 같은 금액도 포함합니다. 비교 결과 자체가 불리언이므로 조건문으로 `true`나 `false`를 따로 대입할 필요가 없습니다.

</details>

완료 조건: 기본 입력, 수량 변경, 합계와 예산이 같은 경계에서 결과가 모두 맞고, 비교 연산자의 결과를 불리언 변수에 저장해 출력할 수 있습니다.

## 선택 A. `iota`로 상태 번호 만들기

본문의 선택 설명을 읽었다면 다음 상수를 한 `const` 블록 안에서 `iota`로 선언해 보세요.

- `statusQueued`: 대기
- `statusRunning`: 진행
- `statusDone`: 완료

첫 값은 0이며 순서대로 증가해야 합니다. 예상 출력:

```text
대기=0 진행=1 완료=2
```

개인 파일은 `_practice/02-variables-types/iota/main.go`입니다. 자신의 터미널에 맞게 준비합니다. 이미 파일이 있으면 복사하지 않습니다.

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/02-variables-types/iota | Out-Null
if (-not (Test-Path _practice/02-variables-types/iota/main.go)) {
    Copy-Item examples/01-first-program/hello/main.go _practice/02-variables-types/iota/main.go
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
mkdir -p _practice/02-variables-types/iota
if [ ! -e _practice/02-variables-types/iota/main.go ]; then
    cp examples/01-first-program/hello/main.go _practice/02-variables-types/iota/main.go
fi
```

실행:

```text
go run ./_practice/02-variables-types/iota
```

<details>
<summary>힌트</summary>

`const` 블록의 첫 상수 식을 `iota`로 두고 다음 선언들의 식을 생략할 수 있습니다. `iota`는 프로그램 실행 중 증가시키는 변수가 아니라 상수 선언에서 쓰는 값입니다. [Go 명세의 iota 설명](https://go.dev/ref/spec#Iota)을 참고하세요.

</details>

## 실습을 마치며

입력을 바꿔 확인한 문제는 기본 값으로 돌려놓습니다. 문제 4의 오류를 고친 뒤 저장소 루트에서 개인 코드의 형식을 정리하고 다시 실행합니다.

```text
go fmt ./_practice/02-variables-types/study-log
go fmt ./_practice/02-variables-types/time-parts
go fmt ./_practice/02-variables-types/repair
go fmt ./_practice/02-variables-types/average
go fmt ./_practice/02-variables-types/receipt
go run ./_practice/02-variables-types/study-log
go run ./_practice/02-variables-types/time-parts
go run ./_practice/02-variables-types/repair
go run ./_practice/02-variables-types/average
go run ./_practice/02-variables-types/receipt
```

- [ ] 변수 선언과 이미 있는 변수의 값 변경을 구분했다.
- [ ] 계산 결과가 언제 변수에 저장되는지 설명할 수 있다.
- [ ] 정수 나눗셈의 몫과 나머지로 시간 단위를 나눴다.
- [ ] 자료형이 다른 값을 명시적으로 변환해 계산했다.
- [ ] 부동소수점 계산과 출력 자릿수를 구분했다.
- [ ] 비교 결과를 불리언 값으로 저장하고 출력했다.
- [ ] 제시된 추가 입력도 실행해 결과를 비교했다.

먼저 문제와 힌트를 충분히 확인한 뒤 [02장 해설](../../solutions/02-variables-types/README.md)에서 계산 과정과 선택 이유를 비교하세요. 참고 코드가 실행되는 것과 자신의 코드가 실행되는 것은 구분해서 확인합니다.

| 교재 | 문제 목록 | 해설 |
| --- | --- | --- |
| [02장 본문](../../book/02-variables-types/README.md) | [전체 연습문제](../README.md) | [02장 해설](../../solutions/02-variables-types/README.md) |
