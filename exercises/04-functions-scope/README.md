# 04장 연습문제: 함수와 스코프

선수 학습: [04. 함수와 스코프](../../book/04-functions-scope/README.md)

이번 문제에서는 계산을 함수로 나누고, 인자와 반환값을 따라가며, 같은 이름의 변수가 어느 범위에 속하는지 확인합니다. 마지막에는 호출 사이의 상태를 기억하는 클로저를 만듭니다. 입력은 코드에 직접 적으며, 키보드 입력이나 외부 패키지는 사용하지 않습니다.

**문제 1~6이 기본 과정**입니다. 선택 A에서는 함수를 다른 함수의 인자로 전달합니다. 모든 명령은 **저장소 루트**에서 실행합니다. 새 `go.mod`나 `go.work`는 만들지 않습니다.

## 문제와 작업 폴더

| 문제 | 수준 | 개인 작업 폴더 |
| --- | --- | --- |
| 1. 인자와 같은 이름의 변수 추적 | 확인 | 먼저 학습 노트에 값과 출력 기록 |
| 2. 점수 판정을 함수로 분리 | 변형 | `_practice/04-functions-scope/grade/` |
| 3. 변경되지 않는 반환값 고치기 | 오류 진단 | `_practice/04-functions-scope/repair/` |
| 4. 여러 수에서 짝수만 더하기 | 변형 | `_practice/04-functions-scope/sum-even/` |
| 5. 학습 계획을 두 함수로 계산 | 응용 | `_practice/04-functions-scope/study-plan/` |
| 6. 호출할 때마다 움직이는 값 | 응용 | `_practice/04-functions-scope/stepper/` |
| 선택 A. 같은 연산을 두 번 적용 | 선택 | `_practice/04-functions-scope/apply-twice/` |

## 준비: 기존 개인 파일을 보존하며 복사하기

자신의 터미널에 맞는 블록 하나를 실행합니다. 기본 문제의 다섯 폴더에 01장의 `main.go`만 복사합니다. 기존 `main.go`가 있으면 덮어쓰지 않으므로, 안내가 나오면 그 내용을 확인하고 이어서 작성하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("grade", "repair", "sum-even", "study-plan", "stepper")) {
    $practiceDir = Join-Path "_practice/04-functions-scope" $practiceName
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
for practice_name in grade repair sum-even study-plan stepper; do
    practice_dir="_practice/04-functions-scope/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

위 블록은 셸 명령입니다. Go 코드의 `for` 문과 구분합니다. 한 폴더에 `main` 함수가 여러 개 생기지 않도록, 각 문제의 코드는 해당 폴더의 `main.go`에서 작성합니다. 공통 규칙은 [연습문제 안내](../README.md)에 있습니다.

## 문제 1. 어느 변수가 바뀔까? — 확인

아래 전체 프로그램을 실행하기 전에 출력 두 줄을 예상하세요.

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

다음 내용을 학습 노트에 기록합니다.

1. 호출 직후 `addOne`의 `value`는 얼마이며, `value++` 뒤에는 얼마인가?
2. `next`에 저장되는 값은 무엇인가? `main`의 첫 번째 `number`도 바뀌는가?
3. `if` 안의 `number := 99`는 어느 범위에서 사용할 변수를 만드는가?
4. 마지막 `Printf`의 `number`는 어느 선언을 가리키는가?

실행으로 확인하고 싶다면 개인 내용이 없는 `_practice/04-functions-scope/predict/main.go`를 편집기에서 만들고 위 코드를 저장합니다. 기존 파일이 있으면 먼저 내용을 보존하세요.

```text
go run ./_practice/04-functions-scope/predict
```

<details>
<summary>힌트 1: 함수 호출부터 추적하기</summary>

이 문제의 `int` 인자는 값이 복사되어 매개변수에 전달됩니다. `value`를 바꾸는 일과 호출자의 변수에 반환값을 대입하는 일을 나누어 생각하세요.

</details>

<details>
<summary>힌트 2: 같은 이름이어도 선언 위치가 다릅니다</summary>

`main`의 바깥쪽 `number`와 `if` 블록에서 `:=`로 선언한 `number`를 서로 다른 칸에 적어 보세요. [선언과 스코프](https://go.dev/ref/spec#Declarations_and_scope)를 참고할 수 있습니다.

</details>

완료 조건: 출력만 맞히는 데 그치지 않고 `value`, 바깥 `number`, 안쪽 `number`, `next`를 구분해서 설명합니다.

## 문제 2. 점수 판정을 함수로 분리하기 — 변형

작업 파일: `_practice/04-functions-scope/grade/main.go`

03장의 점수 판정을 `func classifyScore(score int) (string, bool)`로 분리하세요. 첫 반환값은 판정 문구이고, 두 번째 반환값은 입력 점수가 유효한지 나타냅니다.

| 입력 `score` | 첫 반환값 | 두 번째 반환값 |
| --- | --- | --- |
| 0 미만 또는 100 초과 | 빈 문자열 `""` | `false` |
| 90~100 | `"우수"` | `true` |
| 60~89 | `"통과"` | `true` |
| 0~59 | `"다시 연습"` | `true` |

`classifyScore`는 값을 반환하고, 화면 출력은 `main`에서만 합니다. `main`에서는 `score := 72`로 시작하고 두 반환값을 `label`, `ok`에 받으세요. 유효하지 않으면 `점수 확인`을 출력하고 `return`으로 `main`을 끝냅니다. 유효하면 점수와 판정 문구를 출력합니다.

기본 출력:

```text
72점: 통과
```

실행:

```text
go run ./_practice/04-functions-scope/grade
```

`score`를 `-1`, `0`, `59`, `60`, `89`, `90`, `100`, `101`로 하나씩 바꾸어 경계를 확인한 뒤 `72`로 복구합니다. 범위 밖 입력에서는 `점수 확인` 한 줄만 나와야 합니다.

<details>
<summary>힌트 1: 함수의 결과와 출력 문장을 분리하기</summary>

범위 밖 값도 반환값 두 개가 필요합니다. 먼저 그 경우를 반환하면 나머지 경로는 유효한 점수라고 생각할 수 있습니다. 높은 점수 기준부터 비교하세요.

</details>

<details>
<summary>힌트 2: 두 값을 받고 유효성을 확인하기</summary>

호출 결과는 `label, ok := classifyScore(score)`처럼 두 변수에 받습니다. `ok`를 확인한 뒤에만 `label`을 판정 문구로 사용하세요. 실패 시 출력 후 `return`하면 아래의 성공 출력까지 이어지지 않습니다.

</details>

완료 조건: 모든 경계를 확인하고, 함수 내부에 `fmt.Print...` 호출 없이 같은 판정 결과를 얻습니다.

## 문제 3. 블록 안에서는 바뀌는데 반환값은 그대로? — 오류 진단

작업 파일: `_practice/04-functions-scope/repair/main.go`

아래 코드는 **컴파일되지만 의도와 다르게 동작하는 예**입니다. 기존 개인 내용을 확인한 뒤 전체 코드를 옮겨 실행하세요.

```go
package main

import "fmt"

func lessonLabel(score int) string {
	label := "계속 연습"
	if score >= 60 {
		label := "통과"
		fmt.Println("블록 안:", label)
	}
	return label
}

func main() {
	fmt.Println("반환값:", lessonLabel(72))
}
```

요구사항은 60점 이상일 때 함수의 반환값도 `"통과"`가 되는 것입니다. **블록 안의 대입 부분만 고치고**, 흐름을 확인하는 두 출력 문장은 유지하세요. 72점의 목표 출력은 다음과 같습니다.

```text
블록 안: 통과
반환값: 통과
```

실행:

```text
go run ./_practice/04-functions-scope/repair
```

`lessonLabel(59)`에서는 `반환값: 계속 연습` 한 줄만 나와야 합니다. `60`과 `72`에서는 위와 같은 두 줄이 나와야 합니다. 확인 후 호출 인자를 `72`로 복구합니다.

<details>
<summary>힌트 1: 출력에 사용한 이름의 선언 찾기</summary>

`return label`이 실행되는 위치에서 `if` 안의 선언을 사용할 수 있을까요? 두 `label` 선언과 각각의 중괄호를 찾아보세요.

</details>

<details>
<summary>힌트 2: 새 변수와 기존 변수 변경 구분하기</summary>

이 위치에서 필요한 것은 바깥의 `label`에 새 값을 저장하는 일입니다. 새 변수를 만드는 `:=`와 이미 있는 변수에 대입하는 `=` 중 목적에 맞는 것을 고르세요.

</details>

완료 조건: 수정 전후의 출력 차이를 설명하고, 출력 문장을 지우거나 마지막 반환값을 상수로 바꾸지 않고 고칩니다.

## 문제 4. 여러 수에서 짝수만 더하기 — 변형

작업 파일: `_practice/04-functions-scope/sum-even/main.go`

가변 인자 함수 `func sumEven(numbers ...int) int`를 작성하세요. 전달받은 정수 중 짝수만 더한 결과를 반환합니다. 인자가 없거나 짝수가 없으면 `0`을 반환합니다. 함수 내부에서는 출력하지 않습니다.

`main`에서 `sumEven(1, 2, 3, 4, 6)`을 호출해 다음 한 줄을 출력하세요.

```text
짝수 합계: 12
```

가변 인자 `numbers`는 함수 안에서 `[]int` 형태의 값으로 다뤄집니다. 이번 문제에는 본문에서 배운 `for _, number := range numbers`만 있으면 됩니다. `_`는 순서 번호를 쓰지 않겠다는 뜻이고, `number`는 이번 순서의 값입니다. 슬라이스를 직접 만들거나 변경할 필요는 없습니다. [가변 인자 명세](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)를 참고합니다.

실행:

```text
go run ./_practice/04-functions-scope/sum-even
```

호출의 인자만 바꾸어 아래 결과를 확인한 뒤 기본 호출로 복구합니다. 입력과 합계는 `int` 범위를 넘지 않는 작은 정수로 제한합니다.

| 호출 | 반환값 |
| --- | --- |
| `sumEven()` | `0` |
| `sumEven(8)` | `8` |
| `sumEven(1, 3, 5)` | `0` |
| `sumEven(-2, 3, 4)` | `2` |

<details>
<summary>힌트 1: 아무것도 더하지 않는 경우부터 생각하기</summary>

합계를 0으로 시작하면 반복이 한 번도 실행되지 않을 때도 반환할 결과가 있습니다. 반환은 모든 값을 확인한 뒤에 합니다.

</details>

<details>
<summary>힌트 2: 포함할 값을 조건으로 고르기</summary>

짝수는 2로 나눈 나머지가 0인 값입니다. `number%2 == 0`은 음수 짝수에도 성립합니다. `return`을 반복 안에 두면 첫 값 뒤에 함수가 끝나지 않는지 확인하세요.

</details>

완료 조건: 인자 개수가 달라도 같은 함수로 처리하고, 빈 호출·홀수만 있는 호출·음수가 섞인 호출을 확인합니다.

## 문제 5. 학습 계획을 두 함수로 계산하기 — 응용

작업 파일: `_practice/04-functions-scope/study-plan/main.go`

일주일 안의 학습 계획을 다음 두 함수로 계산하세요. 함수는 계산 결과만 반환하며, 출력은 `main`에서 합니다.

| 함수 | 입력 조건과 반환 계약 |
| --- | --- |
| `func totalMinutes(days, minutesPerDay int) (int, bool)` | `days`가 0~7이고 `minutesPerDay`가 1~120이면 총 분량과 `true`; 하나라도 벗어나면 `0, false` |
| `func splitMinutes(total int) (int, int)` | 유효한 총 분량 0~840을 받아 시간과 남은 분을 순서대로 반환 |

`days == 0`은 쉬는 계획으로 유효합니다. 이때도 `minutesPerDay`의 범위 조건은 적용됩니다. 따라서 `total == 0`만으로 실패 여부를 판단할 수 없습니다.

`main`에서는 `days := 5`, `minutesPerDay := 25`로 시작합니다. `totalMinutes`의 유효성 결과를 먼저 확인하고, 유효하지 않으면 `계획 확인`을 출력한 뒤 `return`합니다. 유효할 때만 `splitMinutes`로 환산합니다. `splitMinutes`에는 유효 범위의 값만 넘긴다는 전제이며, 이 함수에 별도 오류 판정 기능을 추가할 필요는 없습니다.

기본 출력:

```text
총 계획: 125분
환산: 2시간 5분
```

실행:

```text
go run ./_practice/04-functions-scope/study-plan
```

### 확인할 입력

`days`와 `minutesPerDay`를 함께 바꾸고 결과를 기록하세요. 확인 후 `5`와 `25`로 복구합니다.

| `days` | `minutesPerDay` | `totalMinutes`의 결과 |
| --- | --- | --- |
| `0` | `25` | `0, true` |
| `1` | `1` | `1, true` |
| `7` | `120` | `840, true` |
| `-1` | `25` | `0, false` |
| `8` | `25` | `0, false` |
| `5` | `0` | `0, false` |
| `5` | `121` | `0, false` |
| `0` | `0` | `0, false` |

`splitMinutes`는 `0`, `59`, `60`, `125`, `840`을 직접 넣어 시간과 분을 확인하세요. 잠시 `main`의 계산 부분을 바꾸어 확인했다면 기본 프로그램으로 되돌립니다. `main`을 한 파일에 두 개 만들지 않습니다.

<details>
<summary>힌트 1: 유효성은 네 경계를 모두 확인합니다</summary>

날짜의 최솟값·최댓값과 하루 분량의 최솟값·최댓값을 나눠 적어 보세요. 유효하지 않은 경우를 먼저 반환하면 곱셈은 유효한 입력에서만 실행됩니다.

</details>

<details>
<summary>힌트 2: 계산 결과를 다음 함수의 입력으로 연결하기</summary>

총 분량과 `ok`를 받은 뒤 `ok`를 확인합니다. 시간은 총 분량을 60으로 나눈 몫이고, 남은 분은 나머지입니다. 첫 함수의 반환 순서와 두 번째 함수의 반환 순서를 혼동하지 않도록 변수 이름을 붙여 받으세요.

</details>

완료 조건: 유효한 0분과 잘못된 입력을 구분하고, 계산 함수 둘에 출력 코드를 넣지 않으며, 환산 결과의 남은 분이 0~59인지 설명합니다.

## 문제 6. 호출할 때마다 움직이는 값 — 응용

작업 파일: `_practice/04-functions-scope/stepper/main.go`

`func newStepper(start, step int) func() int`를 작성하세요. 이 함수는 인자를 받지 않고 `int`를 반환하는 새 함수를 만듭니다. 반환된 함수는 호출할 때마다 자신이 기억하는 현재 값에 `step`을 더하고 그 값을 반환해야 합니다.

다음 조건을 지킵니다.

- 현재 값은 `start`에서 시작합니다. **첫 호출의 반환값은 `start + step`**입니다.
- `newStepper`를 각각 호출해서 만든 함수는 서로 다른 현재 값을 기억합니다.
- `step`이 0이면 같은 값이 유지되고, 음수이면 값이 줄어듭니다.
- 전역 변수를 사용하지 않습니다. 입력과 반복 호출 횟수는 계산 결과가 `int` 범위를 넘지 않도록 작게 정합니다.

`main`에서 `first := newStepper(10, 3)`, `second := newStepper(100, 10)`을 만듭니다. `first`, `first`, `second`, `first` 순서로 호출하면서 다음과 같이 출력하세요.

```text
첫 번째: 13
첫 번째: 16
두 번째: 110
첫 번째: 19
```

실행:

```text
go run ./_practice/04-functions-scope/stepper
```

아래는 각각 새 함수를 만든 뒤 두 번 호출했을 때의 결과입니다. 확인 후 기본 두 함수와 네 번의 출력으로 복구하세요.

| 생성 호출 | 첫 호출 | 두 번째 호출 |
| --- | --- | --- |
| `newStepper(0, 1)` | `1` | `2` |
| `newStepper(10, 0)` | `10` | `10` |
| `newStepper(5, -2)` | `3` | `1` |

<details>
<summary>힌트 1: 어떤 변수가 호출 사이에 살아 있어야 할까요?</summary>

`current := start`를 반환할 함수의 **바깥**, `newStepper`의 **안**에 둡니다. 반환된 함수가 그 변수를 사용하면 클로저가 됩니다. 현재 값을 반환된 함수 안에서 매번 처음 선언하면 이전 호출의 결과가 남지 않습니다.

</details>

<details>
<summary>힌트 2: 함수를 만드는 일과 실행하는 일 구분하기</summary>

`newStepper(10, 3)`의 결과는 숫자가 아니라 나중에 호출할 함수입니다. 이를 `first`에 저장하고 `first()`로 실행합니다. 반환할 익명 함수에서는 현재 값을 바꾼 뒤 그 값을 반환하세요.

</details>

완료 조건: 두 함수의 호출을 섞어도 값이 독립적으로 변하고, 왜 첫 번째 함수의 마지막 출력이 19인지 설명합니다.

## 선택 A. 같은 연산을 두 번 적용하기

작업 파일: `_practice/04-functions-scope/apply-twice/main.go`

함수를 값으로 전달하는 연습입니다. 기본 여섯 문제를 마친 뒤 진행하세요. 먼저 자신의 셸에 맞는 블록으로 선택 과제 폴더를 준비합니다. 기존 파일이 있으면 보존합니다.

Windows PowerShell:

```powershell
$practiceDir = "_practice/04-functions-scope/apply-twice"
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
practice_dir="_practice/04-functions-scope/apply-twice"
mkdir -p "$practice_dir"
if [ -e "$practice_dir/main.go" ]; then
    printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
else
    cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
fi
```

`func applyTwice(value int, operation func(int) int) int`를 작성합니다. `operation`을 `value`에 적용한 뒤, **첫 결과에 같은 연산을 한 번 더 적용**하여 반환하세요. `operation`에는 실제로 호출할 수 있는 함수를 전달한다고 전제합니다. 함수가 없는 값인 `nil`은 이번 문제의 입력에 포함하지 않습니다.

`main`에서 “입력에 3을 더하는 함수”와 “입력을 2배로 만드는 함수”를 각각 함수 값으로 준비합니다. 시작 값 `5`에 대해 다음을 출력하세요.

```text
3 더하기 두 번: 11
2배 두 번: 20
```

실행:

```text
go run ./_practice/04-functions-scope/apply-twice
```

<details>
<summary>힌트 1: 두 번째 연산의 입력 확인하기</summary>

3을 더하는 경우의 흐름은 `5 → 8 → 11`입니다. 원래 5에 같은 연산을 두 번 각각 적용해서 두 결과를 더하는 문제가 아닙니다.

</details>

<details>
<summary>힌트 2: 함수 매개변수도 호출할 수 있습니다</summary>

`operation(value)`의 결과를 지역 변수에 저장한 다음 그 변수를 `operation`에 전달해 보세요. `applyTwice` 내부에 덧셈이나 곱셈 종류를 판정하는 `if`는 필요하지 않습니다.

</details>

완료 조건: `applyTwice`의 본문을 바꾸지 않고 두 연산을 모두 처리하고, 두 번째 호출에 전달되는 값을 설명합니다.

## 스스로 오류를 좁혀 보기

| 증상 | 먼저 확인할 것 |
| --- | --- |
| 반환값 개수와 변수가 맞지 않는 오류 | 함수 선언의 반환 타입 개수와 호출 결과를 받는 변수 개수 |
| `missing return` | 반환값이 있는 함수의 끝까지 내려오는 경로에 반환이 빠졌는지 |
| 블록 밖의 값이 그대로임 | 안쪽에서 `:=`로 같은 이름의 새 변수를 만들었는지 |
| 빈 가변 인자에서 기대와 다른 결과 | 합계의 시작 값과 반복 뒤의 `return` 위치 |
| 클로저의 값이 호출할 때마다 초기화됨 | 현재 값의 선언을 반환된 함수 안에 두었는지 |
| 함수가 만들어졌지만 숫자가 나오지 않음 | 함수 값을 저장한 뒤 `first()`처럼 호출했는지 |

개인 폴더는 이름이 `_`로 시작하므로 일반 `go test ./...`의 검사 대상에서 제외됩니다. 이번 문제는 자신의 코드를 직접 실행하여 확인합니다. [패키지 탐색 규칙](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)을 참고하세요.

포맷도 저장소 루트에서 개인 폴더를 직접 지정합니다.

```text
go fmt ./_practice/04-functions-scope/grade
go fmt ./_practice/04-functions-scope/repair
go fmt ./_practice/04-functions-scope/sum-even
go fmt ./_practice/04-functions-scope/study-plan
go fmt ./_practice/04-functions-scope/stepper
```

선택 과제를 작성했다면 `go fmt ./_practice/04-functions-scope/apply-twice`도 실행합니다. 포맷 후 각 문제의 실행 명령으로 결과를 다시 확인하세요.

## 해설을 열기 전 점검

- [ ] 인자 값, 함수의 지역 변수, 반환값을 서로 구분해 설명할 수 있다.
- [ ] 범위 밖 점수와 유효한 0분 계획을 반환값 두 개로 구분한다.
- [ ] `:=`로 만드는 새 변수와 `=`로 바꾸는 기존 변수를 구분한다.
- [ ] 가변 인자가 없을 때와 음수가 섞였을 때 결과를 확인했다.
- [ ] 계산 함수와 화면 출력을 분리했다.
- [ ] 서로 따로 만든 클로저의 상태가 독립적임을 확인했다.

이제 [04장 해설과 참고 구현](../../solutions/04-functions-scope/README.md)을 읽고 자신의 결과와 비교합니다. 참고 구현의 테스트가 통과해도 개인 풀이가 자동으로 검사되는 것은 아닙니다.

[04장 본문](../../book/04-functions-scope/README.md) · [연습문제 목록](../README.md)
