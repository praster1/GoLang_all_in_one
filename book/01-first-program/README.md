# 01. 첫 Go 프로그램

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 입문 | [00. 개발 환경과 학습 준비](../00-setup/README.md) | Go, 편집기, 터미널, 이 저장소 |

00장에서 프로그램을 실행할 환경을 마련했습니다. 이번 장에서는 그 프로그램이 **왜 실행되는지, 어떤 순서로 출력하는지, 잘못 작성하면 무엇을 확인해야 하는지** 알아봅니다. 마지막에는 배운 출력 함수를 조합하여 짧은 학습 기록을 만듭니다.

## 학습 목표

이 장을 마치면 다음을 할 수 있습니다.

- `package main`, `import`, `func main()`의 역할을 구분한다.
- 문자열을 출력하고, 공백과 줄바꿈을 포함한 결과를 예측한다.
- `Print`, `Println`, `Printf`를 목적에 맞게 선택한다.
- 자신의 프로그램을 수정·저장·실행·빌드하고 결과를 비교한다.
- 컴파일 오류에서 파일·위치·원인을 찾아 고친다.

## 사전 지식과 준비

00장의 `go version`과 첫 예제 실행이 성공하면 시작할 수 있습니다. 설치나 PATH 문제가 남았다면 [00장](../00-setup/README.md)의 문제 해결 안내를 먼저 확인합니다. 변수·조건문·반복문은 아직 몰라도 됩니다.

**별도 안내가 없는 모든 터미널 명령은 저장소 루트 `GoLang_all_in_one/`에서 실행합니다.** `go.mod`, `book/`, `examples/`가 보이는 위치입니다. `go` 명령은 Windows PowerShell과 macOS·Linux·WSL에서 같고, 폴더 생성과 실행 파일 실행처럼 다른 명령은 나누어 안내합니다.

| 자료 | 위치 | 역할 |
| --- | --- | --- |
| 첫 인사 | [hello/main.go](../../examples/01-first-program/hello/main.go) | 프로그램의 기본 구조 |
| 출력 비교 | [output/main.go](../../examples/01-first-program/output/main.go) | 공백·줄바꿈·서식 확인 |
| 여러 파일, 선택 학습 | [multifile/](../../examples/01-first-program/multifile/) | 같은 패키지의 함수 호출 |
| 연습문제 | [01장 문제](../../exercises/01-first-program/README.md) | 확인·변형·응용 실습 |
| 해설 | [01장 해설](../../solutions/01-first-program/README.md) | 직접 푼 뒤 결과와 이유 비교 |

버전 기준은 루트의 [`go.mod`](../../go.mod)를 따릅니다. 이 장의 예제에는 외부 라이브러리나 계정이 필요하지 않습니다.

## 1. 프로그램을 읽고 실행하기

프로그램은 컴퓨터가 수행할 동작을 코드로 적은 것입니다. Go 소스 파일은 `.go`라는 확장자를 사용합니다. 아래 파일은 터미널에 인사말을 출력하는 완전한 프로그램입니다.

저장 경로: [`examples/01-first-program/hello/main.go`](../../examples/01-first-program/hello/main.go).

```go
package main

import "fmt"

func main() {
	fmt.Println("안녕하세요, Go!")
}
```

코드에서 한글 문장을 찾아보고, 실행하면 무엇이 보일지 예상합니다. 그런 다음 터미널에서 실행합니다.

```sh
go run ./examples/01-first-program/hello
```

예상 출력:

```text
안녕하세요, Go!
```

`go run`은 코드를 컴파일하여 실행합니다. **컴파일(compile)**은 소스 코드를 실행할 수 있는 형태로 바꾸는 과정입니다. 이 프로그램은 인사말을 출력한 뒤 종료하므로, 터미널에서 다시 명령을 입력할 수 있게 됩니다. 키보드 입력을 기다리는 코드는 없습니다. [공식 첫 프로그램 안내](https://go.dev/doc/tutorial/getting-started)와 [go run 설명](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)에서도 이 흐름을 확인할 수 있습니다.

터미널의 `$`, `PS>` 같은 프롬프트와 예상 출력은 입력하는 내용이 아닙니다. 이 교재의 명령 블록에는 직접 입력할 명령만 적습니다.

## 2. 일곱 줄의 역할 이해하기

### 2.1 `package main`: 이 파일이 속한 묶음

**패키지(package)**는 관련된 Go 코드를 묶는 단위입니다. `package main`은 이 파일이 `main`이라는 패키지에 속한다는 선언입니다. 실행 프로그램에는 이름이 `main`인 패키지와 그 안의 `main` 함수가 필요합니다. [프로그램 실행 규칙](https://go.dev/ref/spec#Program_execution)이 정한 이름입니다.

이 교재에서는 독립적으로 실행할 프로그램 하나를 폴더 하나에 둡니다. `hello/`와 `output/`은 각각 다른 프로그램입니다. 같은 폴더에 여러 프로그램의 `main` 함수를 모으면 이름이 중복되므로 폴더를 구분합니다.

`main.go`는 진입점을 찾기 쉽게 붙이는 파일명입니다. **파일명이 실행의 시작점을 결정하지는 않습니다.** 이 예제의 파일 이름을 `hello.go`로 바꾸어도 패키지를 대상으로 실행하면 동작합니다. 파일 이름에 관한 도구의 특별한 규칙은 뒤에서 보는 `_test.go`처럼 별도로 존재합니다.

00장에서 본 **모듈(module)**은 패키지들을 관리하는 더 큰 단위이며, 이 저장소는 루트의 `go.mod` 하나를 사용합니다. 이번 장의 예제나 개인 실습 폴더에 `go.mod`를 추가할 필요가 없습니다. 패키지와 모듈의 자세한 구성은 10장에서 다룰 예정입니다.

### 2.2 `import "fmt"`: 필요한 기능 가져오기

`import`는 이 파일에서 다른 패키지의 기능을 사용하겠다는 선언입니다. `fmt`는 값을 문자로 표현하고 출력하는 기능 등을 제공하는 **표준 라이브러리(standard library)** 패키지입니다. Go에 함께 제공되므로 이 예제를 위해 별도로 설치하지 않습니다.

`"fmt"`는 가져올 패키지의 경로이고, 이 예제에서는 코드 안에서도 `fmt`라는 이름으로 사용합니다. 아래처럼 패키지 이름 뒤에 점을 붙여 필요한 기능을 선택합니다.

```go
fmt.Println("안녕하세요, Go!")
```

위 코드는 `main` 함수 안에서 발췌한 문장입니다. 이 한 줄만 파일에 저장하면 완전한 프로그램이 되지 않습니다.

Go는 이름의 대소문자를 구분합니다. `Println`과 `println`은 다른 이름이며, 다른 패키지에 공개하는 이름은 대문자로 시작합니다. 여기서는 **`fmt.Println`의 `P`를 대문자로 쓴다**는 점부터 기억합니다. 공개 이름의 조건은 [언어 명세](https://go.dev/ref/spec#Exported_identifiers)에 있습니다.

일반적인 import는 가져온 파일에서 실제로 사용해야 합니다. `fmt`를 가져온 뒤 출력문을 모두 지우면 사용하지 않는 import 때문에 컴파일 오류가 납니다. 또한 한 파일의 `import`가 같은 폴더의 다른 파일에 자동으로 적용되지는 않습니다. [import 선언 규칙](https://go.dev/ref/spec#Import_declarations)을 참고합니다.

### 2.3 `func main() { ... }`: 시작할 동작 정의하기

**함수(function)**는 이름을 붙여 둔 동작의 묶음입니다. `func`는 함수를 선언하는 키워드이며, 다음 코드는 `main`이라는 함수를 정의합니다.

```go
func main() {
	// 여기에 실행할 문장을 적습니다.
}
```

이 블록은 함수 모양만 보여 주는 발췌입니다. `//` 뒤의 글은 설명을 위한 주석이며 실행할 동작이 아닙니다.

| 부분 | 의미 |
| --- | --- |
| `func` | 함수 선언을 시작하는 키워드 |
| `main` | 이 함수의 이름 |
| `()` | 이 함수가 입력으로 받는 매개변수가 없음 |
| `{`와 `}` | 함수에서 실행할 문장을 묶는 범위 |

실행 프로그램의 `main` 함수는 입력 매개변수와 반환값이 없습니다. 프로그램은 필요한 초기화가 끝난 뒤 `main`을 호출하고, 이 함수가 끝나면 종료합니다. 첫 예제에서는 출력문 하나를 실행한 뒤 끝납니다. 초기화와 일반 함수의 입력·반환값은 이후 장에서 자세히 배웁니다. [프로그램 실행 명세](https://go.dev/ref/spec#Program_execution)가 이 시작과 종료를 정의합니다.

여는 중괄호 `{`는 `func main()`과 **같은 줄**에 둡니다. Go는 줄 끝의 특정 토큰 뒤에 세미콜론을 자동으로 넣기 때문에, 이 선언의 `{`를 다음 줄에 쓰면 문법 오류가 납니다. 지금은 문장 끝마다 `;`를 붙이지 않고 위 형태를 사용하는 것으로 충분합니다. 근거는 [세미콜론 삽입 규칙](https://go.dev/ref/spec#Semicolons)입니다.

### 2.4 `fmt.Println(...)`: 동작 실행하기

함수를 정의한 것과 사용하는 것은 다릅니다. 함수 이름 뒤에 괄호를 붙여 실행하는 것을 **함수 호출(function call)**이라고 합니다. 괄호 안에 넣어 전달하는 값은 **인자(argument)**입니다.

`fmt.Println("안녕하세요, Go!")`에서는 `Println`을 호출하며 문자열 하나를 인자로 전달합니다. **문자열(string)**은 글자들을 담는 값입니다. 이 예제의 큰따옴표는 문자열의 시작과 끝을 표시하고, 화면에는 따옴표 안의 내용만 나옵니다.

`Println`은 전달받은 값을 **표준 출력(standard output)**에 쓰고 끝에 줄바꿈을 붙입니다. 지금처럼 터미널에서 실행하면 그곳에 결과가 보입니다. 표준 출력을 파일로 보내는 방법은 명령행 프로그램을 다룰 때 배웁니다. [fmt.Println 공식 문서](https://pkg.go.dev/fmt#Println)를 참고합니다.

이 프로그램의 실행을 말로 읽으면 다음과 같습니다.

1. `main` 함수에 들어갑니다.
2. `Println`을 호출하여 인사말과 줄바꿈을 출력합니다.
3. `main` 함수가 끝나고 프로그램이 종료됩니다.

## 3. 공백과 줄바꿈을 직접 제어하기

인사말을 두 줄로 나누거나, 한 줄을 여러 번에 걸쳐 출력하려면 함수마다 어떤 내용을 덧붙이는지 알아야 합니다. 다음 프로그램의 출력이 몇 줄일지 먼저 예상합니다.

저장 경로: [`examples/01-first-program/output/main.go`](../../examples/01-first-program/output/main.go).

```go
package main

import "fmt"

func main() {
	fmt.Print("첫째")
	fmt.Print(" 둘째")
	fmt.Println(" 셋째")
	fmt.Println("Go", "시작")
	fmt.Printf("%s %d단계\n", "Go", 1)
	fmt.Println("한 줄\n다음 줄")
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/01-first-program/output
```

예상 출력:

```text
첫째 둘째 셋째
Go 시작
Go 1단계
한 줄
다음 줄
```

출력문은 여섯 개지만 화면의 내용은 다섯 줄입니다. 첫 세 문장은 한 줄을 이어 쓰고, 마지막 문장은 문자열 안의 줄바꿈 때문에 두 줄을 만듭니다.

### 3.1 `Print`와 `Println`

| 함수 | 값 사이의 공백 | 끝의 줄바꿈 | 이 장에서의 용도 |
| --- | --- | --- | --- |
| `fmt.Print` | 인접한 두 값이 모두 문자열이 아닐 때만 자동 삽입 | 자동으로 붙이지 않음 | 한 줄을 나누어 출력 |
| `fmt.Println` | 인자 사이에 항상 삽입 | 자동으로 붙임 | 값들을 출력하고 줄 마무리 |
| `fmt.Printf` | 서식 문자열에 적은 대로 출력 | 자동으로 붙이지 않음 | 문장 안에 값을 정해진 형태로 넣기 |

첫 번째 `Print`는 `첫째`를 출력하고 줄을 끝내지 않습니다. 다음 문자열 `" 둘째"`와 `" 셋째"`의 맨 앞에는 **직접 쓴 공백**이 있습니다. 세 번째 문장의 `Println`이 마지막에 줄바꿈을 붙여 첫 줄을 마무리합니다.

`fmt.Println("Go", "시작")`에서는 쉼표가 두 인자를 구분합니다. 쉼표 자체가 출력되지는 않으며, `Println`이 두 값 사이에 공백을 하나 넣습니다. `Print`와 `Println`은 이름이 비슷해도 공백 규칙이 다릅니다. 정확한 정의는 [Print](https://pkg.go.dev/fmt#Print), [Println](https://pkg.go.dev/fmt#Println)에 있습니다.

### 3.2 `Printf`: 문장 안에 값 넣기

다음은 출력 예제에서 발췌한 문장입니다.

```go
fmt.Printf("%s %d단계\n", "Go", 1)
```

첫 번째 인자는 출력의 틀인 **서식 문자열(format string)**입니다. 뒤의 값들을 순서대로 틀에 넣습니다.

| 표기 | 이 예제에서 하는 일 |
| --- | --- |
| `%s` | 문자열 `"Go"`를 출력 |
| 두 서식 사이의 공백 | 공백 한 칸을 그대로 출력 |
| `%d` | 정수 `1`을 십진수로 출력 |
| `단계` | 글자를 그대로 출력 |
| `\n` | 줄바꿈 문자 하나를 출력 |

`"1"`은 문자열이고 `1`은 숫자 값입니다. 여기서는 숫자 `1`에 `%d`를 사용한다는 것만 기억하고, 자료형은 02장에서 배웁니다. `%s`, `%d`와 실제 값의 종류·개수가 맞아야 의도한 결과를 얻을 수 있습니다. [fmt의 출력 형식](https://pkg.go.dev/fmt#hdr-Printing)을 참고합니다.

`Printf`는 끝에 줄바꿈을 자동으로 붙이지 않습니다. 위의 `\n`을 지우면 뒤에서 출력하는 `한 줄`이 `Go 1단계` 바로 뒤에 붙습니다. 편집기의 소스 코드에서 문장을 다음 줄에 적는 것과, 실행 결과에 줄바꿈을 넣는 것은 별개입니다.

## 4. 문자열, 주석, 코드 모양

### 4.1 문자열 안의 특별한 문자

큰따옴표 문자열에서 `\`로 시작하는 일부 표기는 특별한 문자를 나타냅니다. 이를 **이스케이프 시퀀스(escape sequence)**라고 합니다.

| 코드 안의 표기 | 뜻 | `main` 안에 넣을 수 있는 코드 조각 | 출력되는 내용 |
| --- | --- | --- | --- |
| `\n` | 줄바꿈 | `fmt.Print("위\n아래")` | `위` 다음 줄에 `아래` |
| `\"` | 큰따옴표 문자 | `fmt.Println("\"Go\"를 배웁니다.")` | `"Go"를 배웁니다.` |
| `\\` | 역슬래시 문자 | `fmt.Println("C:\\work")` | `C:\work` |

`fmt.Println("한 줄\n다음 줄")`에는 문자열 안의 줄바꿈 하나와 `Println`이 마지막에 덧붙이는 줄바꿈 하나가 있습니다. 따라서 “`Println` 한 번은 언제나 화면의 한 줄”이라고 외우면 안 됩니다.

큰따옴표 `"Go"`와 작은따옴표는 서로 바꾸어 쓸 수 없습니다. Go에서 `'G'`처럼 작은따옴표로 감싼 표기는 문자 하나의 값을 나타내는 **룬 리터럴(rune literal)**이며, `'Go'`는 올바른 룬 리터럴이 아닙니다. 이번 장의 문장은 큰따옴표로 쓰고, 문자·룬·유니코드의 관계는 07장에서 배웁니다. [문자열 리터럴](https://go.dev/ref/spec#String_literals)과 [룬 리터럴](https://go.dev/ref/spec#Rune_literals)이 정확한 규칙을 설명합니다.

### 4.2 주석은 실행할 동작이 아닌 설명

다음은 `main` 안에 넣을 수 있는 코드 조각입니다.

```go
// 두 번의 출력을 같은 줄에 이어 씁니다.
fmt.Print("오늘의 언어: ")
fmt.Println("Go")
```

`//`부터 그 줄의 끝까지는 주석입니다. Go는 `/* ... */` 형태의 여러 줄 주석도 지원합니다. 주석은 코드의 목적이나 선택 이유를 설명할 때 사용합니다. 문자열 안의 `"//"`는 문자열의 일부이므로 주석이 아닙니다. [주석 문법](https://go.dev/ref/spec#Comments)을 참고합니다.

### 4.3 들여쓰기는 도구에 맡기기

들여쓰기는 함수 안의 문장들을 읽기 쉽게 합니다. Go에는 표준 포맷 도구 `gofmt`가 있으므로 탭과 공백을 수동으로 맞추는 데 시간을 쓸 필요가 없습니다. 다음 절의 개인 파일을 만든 뒤 `gofmt -w 파일경로`로 정리해 봅니다.

`-w`는 정리한 내용을 해당 파일에 저장한다는 뜻입니다. **포맷은 코드 모양을 정리하는 작업**이므로 `fmt.println`을 `fmt.Println`으로 바꾸거나 잘못된 출력 내용을 고쳐 주지는 않습니다. 문법이 깨진 파일은 포맷 도중 오류가 날 수도 있습니다. [gofmt 문서](https://pkg.go.dev/cmd/gofmt)에서 옵션을 확인할 수 있습니다.

## 5. 자기 파일에서 수정하고 실행하기

이제 예제를 자기 파일로 복사하여 바꿉니다. 개인 실습 경로는 `_practice/01-first-program/message/main.go`입니다. 편집기에서 저장소 루트를 연 상태로 작업하면 파일과 터미널의 위치를 맞추기 쉽습니다.

### 5.1 개인 실습 폴더 준비

다음 중 사용하는 운영체제의 명령만 실행합니다. 파일이 이미 있으면 복사하지 않으므로 이어서 사용하면 됩니다. `main_test.go`는 복사하지 않습니다. 교재의 인사말을 확인하는 테스트를 그대로 복사하면, 자기 인사말로 바꾸었을 때 그 테스트가 실패하기 때문입니다.

**Windows PowerShell**

```powershell
New-Item -ItemType Directory -Force _practice/01-first-program/message | Out-Null
if (-not (Test-Path _practice/01-first-program/message/main.go)) {
    Copy-Item examples/01-first-program/hello/main.go _practice/01-first-program/message/main.go
}
```

**macOS·Linux·WSL**

```bash
mkdir -p _practice/01-first-program/message
if [ ! -f _practice/01-first-program/message/main.go ]; then
    cp examples/01-first-program/hello/main.go _practice/01-first-program/message/main.go
fi
```

이 저장소의 `_practice/`는 개인 학습용이며 Git 추적에서 제외됩니다. 또한 `go test ./...` 같은 패키지 패턴은 `_`로 시작하는 디렉터리를 건너뜁니다. 따라서 **개인 코드는 경로를 명시하여 직접 실행하고 확인**합니다. 공식 [패키지 패턴 규칙](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)을 참고합니다.

### 5.2 문장을 바꾸고 결과 확인

편집기에서 `_practice/01-first-program/message/main.go`를 열고 아래 내용으로 수정하여 저장합니다. 파일은 UTF-8로 저장합니다.

```go
package main

import "fmt"

func main() {
	fmt.Println("첫 프로그램을 수정했습니다.")
	fmt.Println("다음에는 출력을 바꿔 봅니다.")
}
```

저장소 루트에서 포맷하고 실행합니다. 두 운영체제에서 같습니다.

```sh
gofmt -w _practice/01-first-program/message/main.go
go run ./_practice/01-first-program/message
```

예상 출력:

```text
첫 프로그램을 수정했습니다.
다음에는 출력을 바꿔 봅니다.
```

소스에 적은 두 출력문이 순서대로 실행됩니다. 다음 변경을 **한 번에 하나씩** 시도하고, 예상과 실제 결과를 비교합니다.

| 바꾸어 볼 내용 | 실행 전에 생각할 점 |
| --- | --- |
| 두 출력문의 순서 교환 | 어느 문장이 먼저 나올까? |
| 첫 `Println`만 `Print`로 변경 | 두 문장 사이에 공백이나 줄바꿈이 생길까? |
| 첫 문자열 끝에 `\n` 추가 | 원래의 `Println`과 함께 쓰면 빈 줄이 생길까? |

실행할 때마다 **파일 저장 → 실행 명령의 경로 확인 → 출력 비교**를 반복합니다. 결과가 예전과 같으면 코드를 저장했는지, 수정한 파일이 있는 폴더를 실행했는지 먼저 확인합니다.

## 6. 바로 실행하기와 실행 파일 만들기

`go run`은 결과를 바로 확인할 때, `go build`는 실행 파일을 따로 만들어 둘 때 사용합니다. 둘 다 컴파일 과정을 거칩니다. `go run`이 소스를 한 줄씩 읽어 곧바로 실행하는 인터프리터라는 뜻은 아닙니다.

| 명령 | 이 장에서의 동작 | 인사말이 출력되는 시점 |
| --- | --- | --- |
| `go run ./_practice/01-first-program/message` | 빌드한 프로그램을 바로 실행 | 명령 실행 중 |
| `go build -o ... ./_practice/01-first-program/message` | 지정한 경로에 실행 파일 생성 | 생성한 파일을 별도로 실행할 때 |

다음은 저장소 루트에서 개인 프로그램을 빌드하는 명령입니다. `-o`는 결과 파일의 위치와 이름을 정합니다.

**Windows PowerShell**

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o bin/first-program.exe ./_practice/01-first-program/message
.\bin\first-program.exe
```

**macOS·Linux·WSL**

```bash
mkdir -p bin
go build -o bin/first-program ./_practice/01-first-program/message
./bin/first-program
```

5.2절의 코드를 그대로 사용했다면 같은 두 문장이 나옵니다. 변경 실습을 했다면 그 변경 결과가 출력됩니다. 빌드에 성공했을 때 `go build` 자체는 보통 아무것도 출력하지 않습니다.

**소스를 수정해도 이미 만든 실행 파일은 자동으로 바뀌지 않습니다.** 새 내용으로 실행 파일을 갱신하려면 다시 빌드해야 합니다. 프로그램을 계속 수정하는 동안에는 `go run`으로 확인하고, 파일로 실행하는 흐름이 필요할 때 `go build`를 사용합니다. [go build 공식 설명](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)을 참고합니다.

## 7. 오류를 읽고 수정하기

오류는 코드를 고칠 위치를 찾는 단서입니다. 처음에는 여러 부분을 한꺼번에 바꾸기보다 오류 하나를 재현하고, 메시지를 읽고, 수정한 뒤 다시 실행합니다.

### 7.1 파일·줄·열 확인

다음은 **의도적으로 실패시키는 실습**입니다. 5.2절의 개인 파일에서 첫 번째 `fmt.Println`만 `fmt.println`으로 바꾸어 저장합니다. 원본 예제 폴더 대신 `_practice/` 안에서 진행합니다.

```sh
go run ./_practice/01-first-program/message
```

오류 메시지에서 다음과 같은 부분을 찾습니다. 아래는 Go 1.27.3에서 확인한 형태이며, 편집한 위치에 따라 숫자가 달라집니다.

```text
_practice/01-first-program/message/main.go:6:6: undefined: fmt.println (but have Println)
```

| 메시지 부분 | 읽는 방법 |
| --- | --- |
| `_practice/01-first-program/message/main.go` | 고칠 파일 |
| `6:6` | 6번째 줄, 6번째 열 부근 |
| `undefined: fmt.println` | 그 이름을 찾을 수 없음 |

해당 줄의 이름을 `fmt.Println`으로 고치고 저장한 뒤 다시 실행합니다. 컴파일에 실패한 동안에는 이 프로그램의 출력문이 실행되지 않습니다.

### 7.2 자주 만나는 증상

| 증상·실수 | 원인 | 확인과 수정 |
| --- | --- | --- |
| `undefined: fmt.println` | 함수 이름의 대소문자가 다름 | `Println`의 `P`를 대문자로 수정 |
| `undefined: fmt` | 그 파일에 `fmt` import가 없음 | `import "fmt"` 추가 |
| `"fmt" imported and not used` | import를 남기고 사용하는 코드를 지움 | 출력문이 필요하면 복구하고, 아니면 import 삭제 |
| 문자열 관련 문법 오류 | 닫는 따옴표 누락, 큰따옴표 안에 실제 줄바꿈 입력 등 | `"`의 짝을 맞추고 줄바꿈은 `\n` 사용 |
| `'Go'`에서 룬 관련 오류 | 작은따옴표 안에 문자 여러 개를 넣음 | 문자열 `"Go"`로 수정 |
| `{` 근처의 문법 오류 | `func main()` 다음 줄에 `{`를 씀 | 여는 중괄호를 선언과 같은 줄로 이동 |
| `main redeclared` | 한 패키지에 `main`을 두 번 선언 | 독립 프로그램은 별도 폴더로 분리 |
| 실행할 Go 파일이 없다는 오류 | 저장소 루트에서 `go run .` 실행 | `go run ./examples/01-first-program/hello`처럼 대상 지정 |
| 에러 없이 예전 문장이 출력됨 | 저장하지 않은 소스, 다른 경로 실행, 이전 실행 파일 사용 | 저장·실행 경로·재빌드 여부 확인 |
| `gofmt` 이후에도 이름 오류 발생 | 포맷과 이름 확인은 다른 작업 | 이름을 고친 뒤 다시 실행 |

오류 문구 전체를 외우지는 않아도 됩니다. 도구 버전과 코드 상태에 따라 표현이 달라질 수 있습니다. 먼저 파일과 위치를 찾고, 그 문장과 바로 앞 문장을 함께 확인합니다. 따옴표나 중괄호가 앞에서 빠졌다면 뒤쪽 줄에서 오류가 보고될 수도 있습니다.

**컴파일 성공과 의도한 결과는 서로 다른 확인 사항**입니다. 두 문장의 순서를 바꾸어도 컴파일은 성공합니다. 출력 내용·공백·순서까지 맞는지는 실행 결과를 비교해야 합니다.

## 8. 선택 학습: 한 프로그램을 두 파일에 나누기

기본 연습문제를 먼저 풀고 싶은 독자는 9절로 넘어가도 됩니다. 이 절에서는 입력과 반환값이 없는 함수 하나만 더 사용합니다. 함수 설계와 스코프는 04장에서, 패키지 구성은 10장에서 자세히 배웁니다.

저장 경로: [`examples/01-first-program/multifile/main.go`](../../examples/01-first-program/multifile/main.go).

```go
package main

func main() {
	printMessage()
}
```

저장 경로: [`examples/01-first-program/multifile/message.go`](../../examples/01-first-program/multifile/message.go).

```go
package main

import "fmt"

func printMessage() {
	fmt.Println("같은 패키지의 함수를 호출했습니다.")
}
```

두 파일의 `package main`이 같고 같은 폴더에 있습니다. `main`은 같은 패키지에 선언한 `printMessage`를 호출합니다. 그 함수의 출력문이 끝나면 `main`으로 돌아오고 프로그램이 종료됩니다. `printMessage`가 소문자로 시작해도 같은 패키지 안에서 사용할 수 있습니다.

저장소 루트에서 **폴더를 지정**하여 실행합니다.

```sh
go run ./examples/01-first-program/multifile
```

예상 출력:

```text
같은 패키지의 함수를 호출했습니다.
```

반면 아래 명령은 **의도적으로 실패하는 예**입니다.

```sh
go run ./examples/01-first-program/multifile/main.go
```

대표 오류:

```text
examples/01-first-program/multifile/main.go:4:2: undefined: printMessage
```

파일을 나열하는 방식은 지정한 파일만 묶습니다. 이 명령에는 `message.go`가 없으므로 그 파일에 정의한 함수를 찾지 못합니다. 폴더를 지정하면 해당 패키지의 빌드 대상 Go 파일을 함께 사용합니다. `_test.go` 파일은 일반 실행 대상에 포함되지 않습니다. 그래서 이 교재는 기본적으로 **예제 폴더를 지정하는 실행 방식**을 사용합니다. [패키지와 파일 목록 규칙](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)에 근거한 차이입니다.

또 하나 확인할 점은 `fmt`를 사용하는 파일입니다. `message.go`가 `fmt.Println`을 호출하므로 그 파일에 `import "fmt"`를 둡니다. `main.go`에만 import를 적어도 `message.go`에서 쓸 수 있게 되지는 않습니다. **패키지의 함수는 함께 사용하지만, import한 패키지 이름의 범위는 파일 단위**입니다. [선언과 스코프 규칙](https://go.dev/ref/spec#Declarations_and_scope)을 참고합니다.

## 9. 예상 출력이 유지되는지 테스트하기

눈으로 비교한 결과를 반복해서 확인하려면 테스트를 사용할 수 있습니다. 저장소의 `hello` 예제에는 다음 파일이 함께 있습니다.

저장 경로: [`examples/01-first-program/hello/main_test.go`](../../examples/01-first-program/hello/main_test.go).

```go
package main

func Example_main() {
	main()
	// Output:
	// 안녕하세요, Go!
}
```

`_test.go`는 테스트용 파일입니다. 이 함수는 `main()`을 호출하고, 실제 표준 출력을 `Output:` 주석에 적은 기대 결과와 비교합니다. `Example_main`이라는 이름만으로 프로그램의 `main`이 자동 실행되는 것은 아닙니다. 함수 안의 `main()` 호출이 그 일을 합니다.

저장소 루트에서 실행합니다.

```sh
go test -count=1 -v ./examples/01-first-program/hello
```

`-v`는 테스트 이름과 진행 상황을 보여 주고, `-count=1`은 이전 성공 결과를 재사용하지 않고 이번에 테스트를 실행하게 합니다. 결과에 `Example_main`, `PASS`, 마지막의 `ok`가 있는지 확인합니다. 패키지 경로와 소요 시간도 함께 나옵니다.

이 예제 테스트는 출력 비교 시 앞뒤 공백을 무시합니다. 마지막 줄바꿈까지 바이트 단위로 동일하다는 뜻은 아닙니다. 테스트는 **작성한 기대 결과에 맞는지** 확인하며, 모든 가능한 오류가 없다는 보장도 아닙니다. 자세한 규칙은 [실행 가능한 예제 테스트](https://pkg.go.dev/testing#hdr-Examples)와 [go test의 캐시 설명](https://pkg.go.dev/cmd/go#hdr-Test_packages)에 있습니다.

이번 장에서 테스트 문법 전체를 외울 필요는 없습니다. “실제 동작을 기대 결과와 비교할 수 있다”는 목적을 이해하면 됩니다. 자동 테스트의 작성법은 12장에서 다룰 예정입니다.

## 10. 연습문제

[01장 연습문제](../../exercises/01-first-program/README.md)에서 준비 방법, 요구사항, 실행 명령과 힌트를 제공합니다. 문제 1–5는 이번 장에서 배운 내용으로 해결할 수 있습니다.

| 문제 | 수준 | 할 일 |
| --- | --- | --- |
| 1 | 확인 | `Print`·`Println`·`\n`이 만드는 출력 예측 |
| 2 | 변형 | 정해진 한 문장 출력 |
| 3 | 변형 | 두 문장을 정해진 순서로 출력 |
| 4 | 응용 | 세 출력 함수를 조합하여 학습 기록 작성 |
| 5 | 오류 수정 | 잘못된 함수 이름을 메시지에서 찾아 수정 |
| 선택 A | 이후 내용 미리보기 | 이름의 공백을 정리하고 인사말을 만드는 함수 작성 |

선택 A는 함수의 입력·반환값, 조건문과 문자열 처리를 미리 사용합니다. 기본 완료 조건에는 포함하지 않으며 03·04·07장 학습 후 다시 풀어도 됩니다. 스스로 시도한 다음 [해설과 실행 코드](../../solutions/01-first-program/README.md)를 확인합니다.

## 11. 요약과 자기 점검

이번 장의 핵심은 **프로그램 구조를 읽고, 출력을 예측하고, 작은 변경을 실행으로 확인하는 것**입니다.

- 실행 프로그램의 `main` 패키지와 `main` 함수가 시작 구조를 이룹니다.
- `import`로 가져온 기능을 `패키지이름.함수이름(...)` 형태로 호출합니다.
- `Print`, `Println`, `Printf`는 공백·줄바꿈·서식을 다르게 처리합니다.
- 수정한 소스를 저장하고 `go run`으로 확인하거나, 다시 빌드한 실행 파일을 실행합니다.
- 오류 위치와 이름을 읽고 하나씩 고칩니다. 성공한 실행의 출력도 요구사항과 비교합니다.

### 이해 확인

코드를 보지 않고 답한 뒤 펼쳐서 확인합니다.

1. `main.go`를 `hello.go`로 바꾸고 폴더를 지정하여 실행하면 시작점이 사라질까요?
2. `Println`을 두 번 호출하면 언제나 정확히 두 줄만 출력될까요?
3. `fmt.Printf("%s %d단계\n", "Go", 1)`의 `%s`, `%d`, `\n`은 각각 무엇을 할까요?
4. 소스를 고쳤는데 이전 실행 파일에서 예전 문장이 나온다면 무엇을 해야 할까요?
5. 사용하지 않는 `import "fmt"`가 있어도 실행할 수 있을까요?
6. `gofmt`를 실행하면 `fmt.println`의 대소문자 오류도 고쳐질까요?

<details>
<summary>답과 설명</summary>

1. 사라지지 않습니다. 이 예제의 진입점은 파일명이 아닌 `main` 패키지의 `main` 함수입니다.
2. 아닙니다. 문자열 자체에 `\n`이 있으면 더 많은 줄이 생길 수 있습니다. 각 호출은 끝에 줄바꿈 하나를 추가합니다.
3. `%s`에는 `Go`, `%d`에는 십진수 `1`이 들어가고, `\n`은 줄바꿈을 만듭니다. 결과는 `Go 1단계`와 끝의 줄바꿈입니다.
4. 소스를 저장하고 다시 빌드한 파일을 실행합니다. 즉시 확인하려면 수정한 폴더를 대상으로 `go run`을 사용할 수도 있습니다.
5. 이 예제에서는 컴파일 오류입니다. 사용하는 코드를 복구하거나 필요 없는 import를 지웁니다.
6. 아닙니다. `gofmt`는 모양을 정리합니다. 올바른 이름인 `fmt.Println`으로 직접 수정해야 합니다.

</details>

### 완료 체크리스트

- [ ] 첫 프로그램에서 패키지 선언·import·함수 선언·호출을 찾아 설명할 수 있다.
- [ ] 출력 예제의 다섯 줄이 만들어지는 이유를 설명할 수 있다.
- [ ] 개인 폴더의 프로그램을 수정·저장하고 올바른 경로로 실행했다.
- [ ] 실행 파일을 만들고, 변경 사항을 반영하려면 다시 빌드해야 함을 확인했다.
- [ ] 함수 이름 오류를 재현하고 메시지의 위치를 따라 수정했다.
- [ ] 문제 1–5를 풀고 실제 출력 또는 해설과 비교했다.

## 참고 자료

- [Tutorial: Get started with Go](https://go.dev/doc/tutorial/getting-started) — 첫 프로그램의 작성과 실행
- [Go 명세: 프로그램 실행](https://go.dev/ref/spec#Program_execution) — `main` 패키지와 함수
- [Go 명세: 소스 파일 구성](https://go.dev/ref/spec#Source_file_organization) — 패키지·import·선언
- [Go 명세: 문자열 리터럴](https://go.dev/ref/spec#String_literals) — 큰따옴표와 이스케이프
- [fmt](https://pkg.go.dev/fmt) — 출력 함수와 서식
- [go 명령](https://pkg.go.dev/cmd/go) — 실행·빌드·테스트·패키지 지정
- [gofmt](https://pkg.go.dev/cmd/gofmt) — 코드 포맷
- [testing: Examples](https://pkg.go.dev/testing#hdr-Examples) — 실행 결과를 비교하는 예제 테스트

확인일: 2026-11-09. 실행 예제와 실패 사례는 Go 1.27.3에서 확인했습니다.

## 다음 학습

다음은 [**02. 변수·상수·자료형·연산자**](../02-variables-types/README.md)입니다. 이번 장에서 코드에 직접 적었던 문자열과 숫자에 이름을 붙이고 계산하는 방법으로 이어집니다. 현재 제공하는 장과 집필 순서는 [전체 목차](../README.md)와 [로드맵](../../docs/ROADMAP.md)에서 확인합니다.

| 이전 | 목차 | 다음 | 실습과 해설 |
| --- | --- | --- | --- |
| [00. 개발 환경과 학습 준비](../00-setup/README.md) | [전체 목차](../README.md) | [02. 변수·상수·자료형·연산자](../02-variables-types/README.md) | [문제](../../exercises/01-first-program/README.md) · [해설](../../solutions/01-first-program/README.md) |
