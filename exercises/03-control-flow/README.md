# 03장 연습문제: 조건문과 반복문

선수 학습: [03. 조건문과 반복문](../../book/03-control-flow/README.md)

이번 문제에서는 조건에 맞는 실행 경로를 고르고, 정해진 범위 또는 목표에 도달할 때까지 반복합니다. 입력은 코드의 변수에 직접 적습니다. 배열·슬라이스, 별도 함수, 키보드 입력은 사용하지 않습니다.

**문제 1~6이 기본 과정**이며, 선택 A는 조건과 반복을 함께 사용하는 FizzBuzz입니다. 모든 명령은 **저장소 루트**에서 실행합니다. 개인 코드는 `_practice/03-control-flow/` 아래에 두고, 새 `go.mod`나 `go.work`를 만들지 않습니다.

## 문제와 작업 폴더

| 문제 | 수준 | 개인 작업 폴더 |
| --- | --- | --- |
| 1. `continue`의 흐름 예측 | 확인 | 먼저 학습 노트에 반복별 값 기록 |
| 2. 점수에 따른 안내 | 변형 | `_practice/03-control-flow/grade/` |
| 3. 신호에 따른 안내 | 변형 | `_practice/03-control-flow/signal/` |
| 4. 반복의 마지막 값 고치기 | 오류 진단 | `_practice/03-control-flow/repair/` |
| 5. 3의 배수만 합산 | 응용 | `_practice/03-control-flow/multiples/` |
| 6. 목표 학습 시간 채우기 | 응용 | `_practice/03-control-flow/goal/` |
| 선택 A. FizzBuzz | 선택 | `_practice/03-control-flow/fizzbuzz/` |

각 폴더의 파일은 `main.go` 하나로 시작합니다. 서로 다른 프로그램의 `main` 함수를 같은 폴더에 두지 않습니다.

## 준비: 기존 개인 파일을 보존하며 복사하기

자신의 터미널에 맞는 블록 하나를 실행합니다. 기본 문제의 다섯 폴더에 01장의 첫 프로그램을 복사합니다. `main.go`가 이미 있으면 복사하지 않습니다. 안내가 나오면 기존 내용을 확인하고 이어서 작성하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("grade", "signal", "repair", "multiples", "goal")) {
    $practiceDir = Join-Path "_practice/03-control-flow" $practiceName
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
for practice_name in grade signal repair multiples goal; do
    practice_dir="_practice/03-control-flow/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

위 블록은 파일을 준비하는 셸 명령입니다. Go의 `for`와 문법이 다르므로 구분해서 읽습니다. 공통 실습 규칙은 [연습문제 안내](../README.md)에 있습니다.

## 문제 1. 어떤 값이 더해질까? — 확인

다음은 `main` 함수 안의 **발췌 코드**입니다. 실행하기 전에 마지막 출력과 반복 중 `total`의 변화를 예상하세요.

```go
total := 0
for number := 1; number <= 5; number++ {
	if number%2 == 0 {
		continue
	}
	total += number
}
fmt.Println(total)
```

### 기록할 내용

`number`가 1, 2, 3, 4, 5인 다섯 번을 모두 추적합니다. 각 반복에서 다음을 기록하세요.

- `number%2 == 0`이 참인가?
- `total += number`가 실행되는가?
- 이번 반복이 끝났을 때 `total`은 얼마인가?
- `continue`를 만난 뒤 `number++`는 실행되는가?

<details>
<summary>힌트</summary>

`continue`가 생략하는 범위는 현재 반복의 남은 본문입니다. 이 코드에서 증가식은 중괄호 안이 아니라 `for`의 마지막 부분에 있습니다. [continue 명세](https://go.dev/ref/spec#Continue_statements)와 본문의 반복 순서를 확인하세요.

</details>

완료 조건: 출력 숫자와 다섯 단계의 변화를 설명하고, 짝수를 만났을 때 반복 자체가 끝나는 것은 아니라는 점을 구분할 수 있습니다.

## 문제 2. 점수에 따른 안내 — 변형

작업 파일: `_practice/03-control-flow/grade/main.go`

정수 변수 `score`를 `72`로 시작합니다. 하나의 `if`·`else if`·`else` 연결로 아래 규칙에 맞는 문장 **하나만** 출력하세요.

| 점수 | 출력 |
| --- | --- |
| 0 미만 또는 100 초과 | `점수 확인` |
| 90~100 | `우수` |
| 60~89 | `통과` |
| 0~59 | `다시 연습` |

유효 범위를 벗어난 점수를 먼저 확인합니다. 기본 입력 `72`의 예상 출력은 다음과 같습니다.

```text
통과
```

실행:

```text
go run ./_practice/03-control-flow/grade
```

### 경계 확인

`score`의 숫자만 바꾸어 아래 입력을 하나씩 확인합니다.

| 입력 | 예상 출력 |
| --- | --- |
| `-1` | `점수 확인` |
| `0` | `다시 연습` |
| `59` | `다시 연습` |
| `60` | `통과` |
| `89` | `통과` |
| `90` | `우수` |
| `100` | `우수` |
| `101` | `점수 확인` |

확인 후 기본 점수 `72`로 되돌립니다.

<details>
<summary>힌트 1: 유효 범위 밖</summary>

“0보다 작거나 100보다 크다”는 두 조건 중 하나만 참이어도 됩니다. 논리 연산자 `||`를 사용합니다.

</details>

<details>
<summary>힌트 2: 조건의 순서</summary>

90점은 60점 이상이라는 조건도 만족합니다. 먼저 만족한 분기가 실행된다는 점을 생각하여, 유효한 점수의 범위는 높은 기준부터 검사하세요.

</details>

완료 조건: 기본 입력과 모든 경계 입력에서 한 줄만 출력되고, 조건 순서가 중요한 이유를 설명할 수 있습니다.

## 문제 3. 신호에 따른 안내 — 변형

작업 파일: `_practice/03-control-flow/signal/main.go`

문자열 변수 `signal`을 `"yellow"`로 시작합니다. `switch signal`로 값을 비교하고 아래 문장 하나를 출력하세요. 문자열의 대소문자는 구분하며, `fallthrough`는 사용하지 않습니다.

| `signal` | 출력 |
| --- | --- |
| `"red"` | `정지` |
| `"yellow"` | `주의` |
| `"green"` | `진행` |
| 위의 세 문자열이 아닌 값 | `신호 확인` |

기본 출력:

```text
주의
```

실행:

```text
go run ./_practice/03-control-flow/signal
```

`"red"`, `"yellow"`, `"green"`, `"blue"`를 각각 확인한 뒤 기본 값 `"yellow"`로 되돌립니다.

<details>
<summary>힌트</summary>

각 문자열은 별도의 `case`에, 일치하는 값이 없을 때의 동작은 `default`에 둡니다. 일반적인 Go `switch`는 일치한 `case`의 본문을 마친 뒤 다음 `case`의 본문을 자동으로 실행하지 않습니다.

</details>

완료 조건: 네 경우 모두 기대한 한 줄만 출력되고, `default`의 역할을 설명할 수 있습니다.

## 문제 4. 마지막 값이 빠진 합계 고치기 — 오류 진단

작업 파일: `_practice/03-control-flow/repair/main.go`

아래는 **컴파일과 실행은 성공하지만 요구한 합계와 다른 결과를 만드는 코드**입니다. 목표는 1부터 `limit`까지, 마지막 값도 포함하여 더하는 것입니다.

```go
package main

import "fmt"

func main() {
	limit := 5
	total := 0
	for number := 1; number < limit; number++ {
		total += number
	}
	fmt.Printf("합계: %d\n", total)
}
```

실행:

```text
go run ./_practice/03-control-flow/repair
```

### 1단계: 조건 고치기

1. 수정 전의 실제 출력을 확인합니다.
2. 실행된 `number` 값과 실행되지 않은 마지막 값을 기록합니다.
3. 반복 조건을 고쳐 `limit`도 더하도록 만듭니다.

수정 후 기본 출력:

```text
합계: 15
```

### 2단계: 정수 `range`로 다시 작성하기

같은 개인 파일에서 반복문을 `for index := range limit` 형태로 바꿉니다. `index` 자체를 그대로 더하지 말고, 1부터 `limit`까지의 값으로 바꾸어 더하세요. 이전 버전을 보관한다면 별도 폴더를 사용합니다.

두 방식 모두 아래 결과를 확인하세요. 입력 범위는 이 실습에서 사용하는 0 이상의 작은 정수입니다.

| `limit` | 예상 출력 |
| --- | --- |
| `0` | `합계: 0` |
| `1` | `합계: 1` |
| `5` | `합계: 15` |
| `10` | `합계: 55` |

확인 후 기본 값 `5`로 되돌립니다.

<details>
<summary>힌트 1: 끝점 포함</summary>

`number < limit`는 `number`가 `limit`와 같아지는 순간 거짓이 됩니다. 마지막 값도 포함하는 비교 연산자를 선택하세요.

</details>

<details>
<summary>힌트 2: 정수 range의 값</summary>

양의 정수 `limit`를 순회하면 `index`는 0부터 `limit-1`까지입니다. 이번 합계에 필요한 값은 1부터 `limit`까지입니다. [for 문 명세](https://go.dev/ref/spec#For_statements)를 참고하세요.

</details>

완료 조건: 실행 성공과 정답 여부를 구분하고, 두 반복 방식이 같은 합계를 만드는 이유를 설명할 수 있습니다.

## 문제 5. 3의 배수만 합산하기 — 응용

작업 파일: `_practice/03-control-flow/multiples/main.go`

`limit := 10`, `total := 0`, `count := 0`으로 시작합니다. 1부터 `limit`까지 반복하며 다음을 수행하세요.

1. 3의 배수가 아니면 `continue`로 이번 반복의 나머지를 건너뜁니다.
2. 3의 배수이면 `total`에 더하고 `count`를 1 증가시킵니다.
3. 반복이 끝나면 개수와 합계를 출력합니다.

이 문제의 `limit`는 0~100 사이입니다. 범위는 **1부터** 시작하므로 0을 배수 개수에 포함하지 않습니다.

기본 출력:

```text
3의 배수: 3개
합계: 18
```

실행:

```text
go run ./_practice/03-control-flow/multiples
```

### 추가 확인

| `limit` | 개수 | 합계 |
| --- | --- | --- |
| `0` | `0` | `0` |
| `2` | `0` | `0` |
| `3` | `1` | `3` |
| `9` | `3` | `18` |
| `10` | `3` | `18` |

한 번에 입력 하나씩 바꾸어 확인한 뒤 기본 값 `10`으로 되돌립니다.

<details>
<summary>힌트</summary>

3으로 나눈 나머지가 0이 아닌 값을 먼저 건너뜁니다. 합계와 개수의 변경은 `continue` 다음에 둡니다. 누적용 변수는 반복문 밖에서 한 번만 초기화하세요.

</details>

완료 조건: 배수가 없는 경우에도 0개·합계 0을 출력하고, `continue`를 만나도 다음 숫자로 진행하는 이유를 설명할 수 있습니다.

## 문제 6. 목표 학습 시간 채우기 — 응용

작업 파일: `_practice/03-control-flow/goal/main.go`

`targetMinutes := 100`, `dailyMinutes := 30`으로 시작합니다. 하루에 정해진 시간만큼 학습한다고 할 때, 누적 시간이 목표 **이상**이 되는 데 필요한 일수와 최종 누적 시간을 구하세요.

### 조건

- 목표 시간이 음수이거나 하루 학습 시간이 0 이하이면 `입력 확인` 한 줄을 출력합니다. 이 경우 반복을 실행하지 않습니다.
- 유효한 입력이면 `days := 0`, `total := 0`에서 시작합니다.
- `for`의 조건식으로 `total`이 목표보다 작은 동안 반복합니다.
- 매 반복에서 하루 학습 시간을 더하고 일수를 1 증가시킵니다.
- 목표를 처음 달성하면 반복을 마칩니다. 마지막 날은 하루 학습 시간을 전부 더하므로 누적 시간이 목표를 넘을 수 있습니다.
- 이번 실습은 아래에 제시한 작은 정수 입력을 대상으로 합니다.

기본 출력:

```text
4일, 누적 120분
```

실행:

```text
go run ./_practice/03-control-flow/goal
```

### 경계와 잘못된 입력 확인

| 목표 시간 | 하루 학습 시간 | 예상 출력 |
| --- | --- | --- |
| `100` | `30` | `4일, 누적 120분` |
| `90` | `30` | `3일, 누적 90분` |
| `0` | `30` | `0일, 누적 0분` |
| `-1` | `30` | `입력 확인` |
| `100` | `0` | `입력 확인` |
| `100` | `-1` | `입력 확인` |

확인 후 기본 목표 `100`과 하루 학습 시간 `30`으로 되돌립니다.

<details>
<summary>힌트 1: 반복하기 전에 확인할 것</summary>

하루 학습 시간이 0이면 누적 시간이 증가하지 않습니다. 잘못된 입력을 확인하는 분기와, 누적하는 반복이 들어 있는 분기를 나누세요.

</details>

<details>
<summary>힌트 2: 반복이 끝나는 순간</summary>

목표를 이미 달성했다면 한 번 더 반복하면 안 됩니다. 목표가 0일 때 처음부터 조건이 거짓인지도 확인하세요. 반복문 안에는 조건을 언젠가 거짓으로 만드는 값 변경이 있어야 합니다.

</details>

완료 조건: 목표를 넘는 경우·정확히 달성하는 경우·처음부터 달성한 경우를 구분하며, 잘못된 입력에서 반복을 시작하지 않습니다.

## 선택 A. FizzBuzz

1부터 15까지 순서대로 반복하고, 각 숫자마다 다음 규칙에 따른 결과를 한 줄 출력하세요. **조건식들을 나열하는 `switch`**, 즉 `switch` 뒤에 비교할 값을 쓰지 않는 형태를 사용합니다.

| 조건 | 출력 |
| --- | --- |
| 3과 5의 배수 | `FizzBuzz` |
| 3의 배수 | `Fizz` |
| 5의 배수 | `Buzz` |
| 나머지 | 해당 숫자 |

15에서는 한 줄의 `FizzBuzz`가 나와야 합니다. `Fizz`와 `Buzz`를 두 줄로 출력하지 않습니다.

개인 파일은 `_practice/03-control-flow/fizzbuzz/main.go`입니다. 자신의 터미널에 맞게 준비합니다. 기존 파일이 있다면 복사하지 않습니다.

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/03-control-flow/fizzbuzz | Out-Null
if (-not (Test-Path _practice/03-control-flow/fizzbuzz/main.go)) {
    Copy-Item examples/01-first-program/hello/main.go _practice/03-control-flow/fizzbuzz/main.go
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
mkdir -p _practice/03-control-flow/fizzbuzz
if [ ! -e _practice/03-control-flow/fizzbuzz/main.go ]; then
    cp examples/01-first-program/hello/main.go _practice/03-control-flow/fizzbuzz/main.go
fi
```

실행:

```text
go run ./_practice/03-control-flow/fizzbuzz
```

<details>
<summary>힌트</summary>

15의 배수는 3의 배수이기도 하고 5의 배수이기도 합니다. 먼저 만족하는 `case`가 선택되므로 두 조건을 함께 만족하는 경우를 먼저 검사하세요.

</details>

## 실습을 마치며

변경한 입력을 각 문제의 기본 값으로 되돌린 뒤, 개인 파일을 정리하고 다시 실행합니다.

```text
go fmt ./_practice/03-control-flow/grade
go fmt ./_practice/03-control-flow/signal
go fmt ./_practice/03-control-flow/repair
go fmt ./_practice/03-control-flow/multiples
go fmt ./_practice/03-control-flow/goal
go run ./_practice/03-control-flow/grade
go run ./_practice/03-control-flow/signal
go run ./_practice/03-control-flow/repair
go run ./_practice/03-control-flow/multiples
go run ./_practice/03-control-flow/goal
```

- [ ] 조건에 맞는 분기가 하나만 실행되는 이유를 설명했다.
- [ ] 비교 경계의 바로 전·경계·바로 다음 값을 확인했다.
- [ ] 반복의 시작 값과 마지막 포함 값을 구분했다.
- [ ] 반복을 한 번도 하지 않는 경우를 확인했다.
- [ ] `continue` 이후의 증가식과 다음 조건 검사를 설명했다.
- [ ] 누적 변수의 초기화 위치와 반복 종료 조건을 확인했다.
- [ ] 잘못된 입력에서는 목표 누적 반복을 실행하지 않았다.

먼저 문제와 힌트를 확인한 뒤 [03장 해설](../../solutions/03-control-flow/README.md)에서 풀이와 실행 흐름을 비교하세요. 정답 숫자뿐 아니라 어떤 분기가 실행되었고 몇 번 반복했는지를 자신의 말로 설명해 봅니다.

| 교재 | 문제 목록 | 해설 |
| --- | --- | --- |
| [03장 본문](../../book/03-control-flow/README.md) | [전체 연습문제](../README.md) | [03장 해설](../../solutions/03-control-flow/README.md) |
