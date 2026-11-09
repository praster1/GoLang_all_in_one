# 01장 연습문제: 첫 Go 프로그램

선수 학습: [01. 첫 Go 프로그램](../../book/01-first-program/README.md)

먼저 출력을 예측하고, 문자열을 바꾸고, 여러 출력 함수를 조합해 봅니다. 마지막에는 의도적으로 넣은 오류를 읽고 고칩니다. **문제 1~5가 기본 과정**이며, 선택 A는 이후 장의 함수·조건문·문자열 처리를 미리 경험하는 과제입니다.

모든 명령은 **저장소 루트**에서 실행합니다. 교재 원본은 그대로 두고 `_practice/01-first-program/` 아래에서 연습합니다. 개인 폴더에는 새 `go.mod`나 `go.work`를 만들지 않습니다.

## 문제와 작업 파일

| 문제 | 수준 | 작업 파일 또는 방법 |
| --- | --- | --- |
| 1. 출력 예측 | 확인 | 먼저 종이나 학습 노트에 결과 기록 |
| 2. 한 줄 출력 | 변형 | `_practice/01-first-program/message/main.go` |
| 3. 두 줄 출력 | 변형 | 문제 2의 `message/main.go`를 이어서 수정 |
| 4. 학습 기록 보고서 | 응용 | `_practice/01-first-program/report/main.go` |
| 5. 오류 고치기 | 진단 | `_practice/01-first-program/repair/main.go` |
| 선택 A. 이름에 맞는 인사말 | 후속 내용 미리보기 | `_practice/01-first-program/greeting/main.go` |

문제 2의 답을 보관하고 싶다면 문제 3을 시작하기 전에 다른 폴더에 복사하세요. 같은 폴더에 `main` 함수를 포함한 답안 파일을 여러 개 두면 중복 선언 오류가 발생합니다.

## 개인 실습 파일 준비

아래에서 자신의 터미널에 맞는 블록 하나만 실행합니다. `message`, `report`, `repair` 폴더에 첫 프로그램을 복사합니다. **기존 `main.go`가 있으면 복사하지 않으므로** 본문 실습이나 이전 학습 내용이 유지됩니다. 이미 파일이 있다는 안내가 나오면 내용을 확인한 뒤 해당 문제에 맞게 수정하세요.

Windows PowerShell:

```powershell
foreach ($practiceName in @("message", "report", "repair")) {
    $practiceDir = Join-Path "_practice/01-first-program" $practiceName
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
for practice_name in message report repair; do
    practice_dir="_practice/01-first-program/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

위의 반복문은 폴더와 파일을 준비하는 **셸 명령**입니다. 아직 Go의 반복문을 배울 필요는 없습니다. 기본 문제에서는 Go 소스 파일 하나만 복사하며, 선택 과제의 `greeting` 테스트를 복사하지 않습니다.

## 문제 1. 줄과 공백을 예측하기 — 확인

다음은 `main` 함수 안에 들어갈 **발췌 코드**입니다. 실행하기 전에 전체 출력과 줄 수를 학습 노트에 적으세요.

```go
fmt.Print("A")
fmt.Println("B", "C")
fmt.Println("D\nE")
```

### 확인할 내용

1. `A`와 `B` 사이에 공백이나 줄바꿈이 생기는가?
2. `B`와 `C` 사이에는 어떤 문자가 들어가는가?
3. 문자열 안의 `\n`과 `Println`이 붙이는 줄바꿈은 각각 어디에 생기는가?
4. 문자 내용이 있는 줄은 모두 몇 줄인가?

<details>
<summary>힌트</summary>

각 호출을 마쳤을 때 커서가 같은 줄에 남아 있는지, 다음 줄로 이동했는지 차례대로 추적하세요. `Println`에 문자열 두 개를 전달한 경우와, 문자열 하나 안에 `\n`이 있는 경우를 구분합니다.

</details>

완료 조건: 출력의 글자뿐 아니라 **공백과 줄바꿈의 위치**를 설명할 수 있습니다. 정답은 아래 해설 링크에서 확인합니다.

## 문제 2. 인사말 바꾸기 — 변형

작업 파일: `_practice/01-first-program/message/main.go`

첫 예제의 출력 문자열을 바꾸어 다음 한 줄을 출력하세요. 문장 끝의 마침표와 마지막 줄바꿈까지 포함합니다. 입력을 받거나 변수를 선언할 필요는 없습니다.

```text
Go를 공부합니다.
```

저장소 루트에서 실행합니다.

```text
go run ./_practice/01-first-program/message
```

<details>
<summary>힌트</summary>

`package main`, `import "fmt"`, `func main()`의 구조를 유지하고 출력 함수에 전달하는 문자열을 바꿉니다. 문자열을 감싸는 큰따옴표는 출력 내용에 포함되지 않습니다.

</details>

완료 조건: 오류 없이 실행되며, 지정한 문장이 한 줄로 출력됩니다.

## 문제 3. 순서대로 두 줄 출력하기 — 변형

작업 파일: 문제 2의 `_practice/01-first-program/message/main.go`

다음 두 문장을 순서대로 출력하세요. 두 문장 사이에 빈 줄을 넣지 않습니다. 이번 문제는 **`fmt.Println` 호출 두 개**를 사용하고 문자열 안에는 `\n`을 넣지 않습니다.

```text
Go를 공부합니다.
작은 프로그램부터 직접 실행합니다.
```

실행 명령:

```text
go run ./_practice/01-first-program/message
```

<details>
<summary>힌트</summary>

`main` 함수의 중괄호 안에 출력문을 두 개 둡니다. 각 `Println` 호출은 전달한 값을 출력한 뒤 줄바꿈을 붙입니다. 별도로 빈 문자열을 출력할 필요는 없습니다.

</details>

완료 조건: 정확히 두 줄이며, 두 문장의 순서·공백·마침표가 예상 결과와 같습니다.

## 문제 4. 학습 기록 보고서 — 응용

작업 파일: `_practice/01-first-program/report/main.go`

배운 출력 함수 세 가지를 사용해 다음 네 줄을 출력하세요.

```text
학습 기록
언어: Go
완료한 장: 1
다음 목표: 변수와 자료형
```

### 요구 사항

- 첫 줄과 마지막 줄은 각각 `fmt.Println`으로 출력합니다.
- 둘째 줄은 `fmt.Print`로 `언어: `를 출력한 뒤, `fmt.Println`으로 `Go`를 이어서 출력합니다. 콜론 뒤의 공백을 확인하세요.
- 셋째 줄은 `fmt.Printf`와 `%d`를 사용합니다. 정수 `1`을 별도 인자로 전달하고, 서식 문자열에 줄바꿈을 넣습니다.
- 변수, 키보드 입력, 반복문은 사용하지 않습니다. 이 장의 출력 함수 호출만 조합하면 됩니다.

실행 명령:

```text
go run ./_practice/01-first-program/report
```

<details>
<summary>힌트 1: 둘째 줄</summary>

첫 호출이 줄바꿈을 붙이지 않아야 다음 호출의 내용이 같은 줄에 이어집니다. 두 호출 사이의 공백은 출력할 문자열에 직접 넣습니다.

</details>

<details>
<summary>힌트 2: 셋째 줄</summary>

`%d`는 정수를 넣을 자리입니다. 서식 문자열 뒤에 쉼표와 정수 값을 전달합니다. `Printf`는 자동으로 마지막 줄바꿈을 붙이지 않으므로 문자열 안의 `\n`이 필요합니다. [fmt 공식 문서](https://pkg.go.dev/fmt#Printf)를 참고할 수 있습니다.

</details>

완료 조건: 정확히 네 줄이 출력되고, `Print`·`Println`·`Printf`가 각각 어떤 역할을 맡았는지 설명할 수 있습니다.

## 문제 5. 출력 함수 이름 고치기 — 오류 진단

작업 파일: `_practice/01-first-program/repair/main.go`

아래는 **의도적으로 컴파일에 실패하는 코드**입니다. 교재 원본을 변경하지 말고 개인 실습 파일에서만 확인합니다.

```go
package main

import "fmt"

func main() {
	fmt.println("Go를 공부합니다.")
}
```

실행 명령:

```text
go run ./_practice/01-first-program/repair
```

### 수행할 작업

1. 오류 메시지에서 파일 경로, 위치, 문제가 된 이름을 찾습니다.
2. 출력 함수 이름에서 대소문자가 잘못된 부분만 고칩니다.
3. 저장한 뒤 같은 명령을 다시 실행합니다.
4. `Go를 공부합니다.`가 한 줄로 출력되는지 확인합니다.
5. “Go가 설치되지 않았다”가 아닌, 이 실패의 실제 원인을 한 문장으로 적습니다.

<details>
<summary>힌트</summary>

Go는 식별자의 대소문자를 구분합니다. 첫 예제에서 사용한 함수 이름과 비교해 보세요. `fmt` 패키지의 공식 문서에서도 동일한 이름을 확인할 수 있습니다.

</details>

완료 조건: 수정 전에는 실패하고 수정 후에는 성공하며, 바꾼 부분과 실패 원인을 설명할 수 있습니다. 오류 메시지의 정확한 문구는 Go 버전에 따라 달라질 수 있습니다.

## 선택 A. 이름에 맞춰 인사말 만들기 — 후속 내용 미리보기

이 과제는 이후 03장의 조건문, 04장의 함수, 07장의 문자열 처리를 먼저 사용합니다. **기본 문제 1~5를 마치면 이 장의 학습을 완료할 수 있으며**, 선택 과제는 관련 장을 배운 뒤 돌아와도 됩니다. 예정된 장의 위치는 [전체 목차](../../book/README.md)에서 확인합니다.

작업 파일은 `_practice/01-first-program/greeting/main.go`입니다. 앞에서 만든 기본 문제의 폴더와 분리합니다. 자신의 터미널에 맞는 준비 명령을 실행하세요. 기존 파일이 있으면 그대로 두고 내용을 확인합니다.

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/01-first-program/greeting | Out-Null
if (-not (Test-Path _practice/01-first-program/greeting/main.go)) {
    Copy-Item examples/01-first-program/hello/main.go _practice/01-first-program/greeting/main.go
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
mkdir -p _practice/01-first-program/greeting
if [ ! -e _practice/01-first-program/greeting/main.go ]; then
    cp examples/01-first-program/hello/main.go _practice/01-first-program/greeting/main.go
fi
```

### 요구 사항

`greeting(name string) string` 함수를 작성합니다.

1. 이름의 앞뒤 공백을 제거합니다.
2. 결과가 빈 문자열이면 이름을 `Go`로 사용합니다.
3. `안녕하세요, 이름!` 형태의 문자열을 반환합니다.
4. `greeting` 함수 안에서는 출력하지 않습니다.
5. `main` 함수는 `greeting("Go")`의 반환값을 출력합니다.

키보드 입력을 읽을 필요는 없습니다. 함수에 코드로 값을 전달합니다.

| 입력 이름 | 함수가 반환할 문자열 |
| --- | --- |
| `"Go"` | `"안녕하세요, Go!"` |
| `"민수"` | `"안녕하세요, 민수!"` |
| `"  민수  "` | `"안녕하세요, 민수!"` |
| `""` | `"안녕하세요, Go!"` |
| `" \t\n"` | `"안녕하세요, Go!"` |
| `"Go learner"` | `"안녕하세요, Go learner!"` |

실행 명령:

```text
go run ./_practice/01-first-program/greeting
```

`main`이 `greeting("Go")`를 출력할 때의 예상 결과:

```text
안녕하세요, Go!
```

<details>
<summary>힌트 1: 입력과 반환값</summary>

`greeting(name string) string`에서 괄호 안은 문자열 입력을 `name`으로 받는다는 뜻이고, 괄호 뒤의 `string`은 문자열을 반환한다는 뜻입니다. `main`과 별도의 함수를 선언하고 `return`으로 결과를 돌려줍니다.

</details>

<details>
<summary>힌트 2: 공백과 기본 이름</summary>

앞뒤 공백은 [strings.TrimSpace](https://pkg.go.dev/strings#TrimSpace)로 제거합니다. `strings` 패키지를 import에 추가해야 합니다. 공백 제거 후 `if name == ""` 조건으로 빈 이름을 확인합니다. 이름 가운데의 공백은 그대로 둡니다.

</details>

<details>
<summary>힌트 3: 문자열 만들기</summary>

[fmt.Sprintf](https://pkg.go.dev/fmt#Sprintf)는 서식에 맞는 문자열을 반환합니다. `Printf`처럼 바로 출력하는 함수와 구분하세요. 만든 문자열을 `greeting`이 반환하고, `main`에서 `Println`으로 출력합니다.

</details>

선택 과제의 구현을 마쳤다면 [해설의 선택 A 검증 안내](../../solutions/01-first-program/README.md)에서 전용 테스트를 사용하는 방법을 확인합니다. 이 테스트를 기본 문제의 `message`, `report`, `repair` 폴더로 복사하지 않습니다.

## 실습을 마치며

문제 1~5를 마친 뒤 다음을 확인합니다.

```text
go fmt ./_practice/01-first-program/message
go fmt ./_practice/01-first-program/report
go fmt ./_practice/01-first-program/repair
go run ./_practice/01-first-program/message
go run ./_practice/01-first-program/report
go run ./_practice/01-first-program/repair
```

- [ ] 실행 전에 공백과 줄바꿈을 포함한 출력을 예상했다.
- [ ] 문제 3의 두 줄과 문제 4의 네 줄이 정확히 출력된다.
- [ ] 문제 5의 오류를 고쳤고 원인을 설명할 수 있다.
- [ ] 서로 다른 프로그램을 서로 다른 폴더에 두었다.
- [ ] 실행한 파일 경로와 편집한 파일 경로가 일치한다.

막혔다면 명령·현재 폴더·저장 여부·오류 위치 순서로 확인합니다. 이후 [01장 해설](../../solutions/01-first-program/README.md)과 자신의 풀이를 비교하세요. 해설과 코드 모양이 같은지보다, 요구한 결과를 만들고 그 이유를 설명할 수 있는지가 중요합니다.

| 교재 | 문제 목록 | 해설 |
| --- | --- | --- |
| [01장 본문](../../book/01-first-program/README.md) | [전체 연습문제](../README.md) | [01장 해설](../../solutions/01-first-program/README.md) |
