# 01장 해설: 첫 Go 프로그램

문제: [01장 연습문제](../../exercises/01-first-program/README.md)

문제와 힌트를 먼저 읽고 직접 시도한 뒤 해설을 확인합니다. 다른 코드를 작성했더라도 요구한 출력과 조건을 만족하면 됩니다. 각 호출이 공백과 줄바꿈에 어떤 영향을 주는지 설명해 보세요.

모든 명령은 **저장소 루트**에서 실행합니다. 아래 개인 실습 경로는 문제 문서에서 준비한 파일을 뜻합니다. 완성 코드를 확인한다고 교재 원본이나 다른 문제의 파일에 덮어쓸 필요는 없습니다.

## 문제 1. 줄과 공백 예측

문제의 세 출력문이 만드는 결과는 다음과 같습니다.

```text
AB C
D
E
```

문자 내용이 있는 줄은 **세 줄**이고, 마지막 `E` 뒤에도 줄바꿈이 있습니다. 줄바꿈을 보이게 표기하면 `AB C\nD\nE\n`입니다. 이 표기의 `\n`은 실제 출력에서 보이는 두 문자가 아니라 줄바꿈을 나타냅니다.

| 호출 | 출력에 더해지는 내용 | 호출 후 위치 |
| --- | --- | --- |
| `fmt.Print("A")` | `A` | 첫째 줄의 `A` 바로 뒤 |
| `fmt.Println("B", "C")` | `B C`와 줄바꿈 | 둘째 줄의 시작 |
| `fmt.Println("D\nE")` | `D`, 줄바꿈, `E`, 마지막 줄바꿈 | 다음 줄의 시작 |

둘째 호출은 앞선 출력과 이어지므로 `A`와 `B` 사이에 자동 공백이 생기지 않습니다. `Println`이 공백을 넣는 곳은 **한 호출에 전달된 인자 사이**입니다. 마지막 호출은 문자열 안에 줄바꿈이 하나 있고, 함수가 끝에 줄바꿈을 하나 더 붙이므로 두 줄을 출력합니다. 마지막 줄을 끝내는 줄바꿈을 별도의 빈 출력 줄로 세지 않습니다. [Print](https://pkg.go.dev/fmt#Print)와 [Println](https://pkg.go.dev/fmt#Println)의 공식 동작을 비교해 보세요.

## 문제 2. 한 줄 출력

개인 파일 `_practice/01-first-program/message/main.go`의 전체 코드:

```go
package main

import "fmt"

func main() {
	fmt.Println("Go를 공부합니다.")
}
```

실행:

```text
go run ./_practice/01-first-program/message
```

출력:

```text
Go를 공부합니다.
```

첫 예제에서 바뀐 것은 출력 함수에 전달하는 문자열입니다. 함수 호출과 패키지 선언의 구조를 함께 바꿀 필요는 없습니다. `Println`이 마지막 줄바꿈을 붙이므로 문자열 끝에 `\n`을 따로 넣지 않았습니다.

## 문제 3. 두 줄 출력

같은 개인 파일 `_practice/01-first-program/message/main.go`를 다음처럼 수정합니다. 문제 2와 3의 파일을 같은 폴더에 각각 두는 것이 아니라, 문제 2를 확장하는 단계입니다.

```go
package main

import "fmt"

func main() {
	fmt.Println("Go를 공부합니다.")
	fmt.Println("작은 프로그램부터 직접 실행합니다.")
}
```

실행:

```text
go run ./_practice/01-first-program/message
```

출력:

```text
Go를 공부합니다.
작은 프로그램부터 직접 실행합니다.
```

호출 순서대로 첫 문장과 둘째 문장이 출력됩니다. `fmt.Println("")`을 가운데에 추가하면 불필요한 빈 줄이 생깁니다. 두 문장 사이에 빈 줄이 생겼다면 추가 호출이나 문자열 안의 `\n`을 확인하세요.

## 문제 4. 학습 기록 보고서

참고 구현은 [report/main.go](report/main.go), 출력 테스트는 [report/main_test.go](report/main_test.go)에 있습니다. 개인 풀이 파일은 `_practice/01-first-program/report/main.go`입니다.

참고 구현의 전체 코드:

```go
package main

import "fmt"

func main() {
	fmt.Println("학습 기록")
	fmt.Print("언어: ")
	fmt.Println("Go")
	fmt.Printf("완료한 장: %d\n", 1)
	fmt.Println("다음 목표: 변수와 자료형")
}
```

참고 구현을 실행하고 테스트하는 명령:

```text
go run ./solutions/01-first-program/report
go test -v ./solutions/01-first-program/report
```

프로그램 출력:

```text
학습 기록
언어: Go
완료한 장: 1
다음 목표: 변수와 자료형
```

| 부분 | 선택한 출력 방식 | 이유 |
| --- | --- | --- |
| 제목 | `Println` | 제목을 출력한 뒤 다음 줄로 이동 |
| `언어: ` | `Print` | 같은 줄에 뒤의 내용을 이어서 출력 |
| `Go` | `Println` | 앞의 내용과 합쳐 둘째 줄을 완성 |
| 완료한 장 번호 | `Printf` | `%d` 자리에 정수 `1`을 넣고 `\n`으로 줄을 끝냄 |
| 다음 목표 | `Println` | 마지막 문장을 한 줄로 출력 |

`Printf`는 서식 문자열에 따라 값을 배치합니다. 여기서 `%d`는 정수의 십진수 표시이며, 줄바꿈은 서식 문자열에 직접 넣었습니다. [fmt의 서식 설명](https://pkg.go.dev/fmt#hdr-Printing)을 참고하세요.

자신의 풀이를 확인할 때는 아래처럼 **개인 경로**를 실행해야 합니다. 참고 구현의 테스트를 통과했다고 개인 코드까지 검증된 것은 아닙니다.

```text
go run ./_practice/01-first-program/report
```

출력이 다르면 콜론 뒤 공백, `Printf`의 `\n`, 함수 호출 순서를 확인합니다. 이 문제에서 별도 변수를 만들지 않은 이유는 값을 보관하는 문법을 배우기 전에 출력 함수의 차이를 먼저 확인하기 위해서입니다.

## 문제 5. 함수 이름 오류 수정

잘못된 이름은 `fmt.println`이며, 고친 이름은 `fmt.Println`입니다. `Println`의 첫 글자 `P`를 대문자로 바꿉니다.

개인 파일 `_practice/01-first-program/repair/main.go`의 수정 후 전체 코드:

```go
package main

import "fmt"

func main() {
	fmt.Println("Go를 공부합니다.")
}
```

다시 실행합니다.

```text
go run ./_practice/01-first-program/repair
```

출력:

```text
Go를 공부합니다.
```

Go는 이름의 대소문자를 구분합니다. `fmt` 패키지에서 사용하는 함수 이름은 `Println`이며, 소문자로 바꾸면 다른 이름이 됩니다. 다른 패키지의 공개 이름에 접근하는 규칙은 [Go 명세의 공개 식별자 설명](https://go.dev/ref/spec#Exported_identifiers)에 있습니다.

오류를 읽을 때는 다음 순서로 확인합니다.

1. 메시지에 표시된 파일이 내가 편집한 `repair/main.go`인지 확인합니다.
2. 표시된 위치에서 문제가 된 이름을 찾습니다.
3. 본문 예제나 공식 API 문서와 이름을 비교합니다.
4. 필요한 부분을 고쳐 저장하고 같은 명령을 다시 실행합니다.

예상한 인사말이 나오지 않은 이유는 출력에 도달하기 전에 컴파일이 실패했기 때문입니다. 실패한 호출을 지워서 아무것도 출력하지 않는 프로그램으로 만드는 것은 이 문제의 해결이 아닙니다. 수정 후 약속한 한 줄이 출력되는지까지 확인합니다.

## 선택 A. 이름에 맞는 인사말

이 과제는 조건문·함수·문자열을 더 배운 뒤 이해해도 되는 미리보기입니다. 기본 문제 1~5의 완료 기준에는 포함되지 않습니다.

- [실행 가능한 참고 구현](greeting/main.go)
- [입력별 결과와 기본 출력을 확인하는 테스트](greeting/main_test.go)

참고 구현의 전체 코드:

```go
package main

import (
	"fmt"
	"strings"
)

func greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Go"
	}
	return fmt.Sprintf("안녕하세요, %s!", name)
}

func main() {
	fmt.Println(greeting("Go"))
}
```

참고 구현의 실행과 검증:

```text
go run ./solutions/01-first-program/greeting
go test -v ./solutions/01-first-program/greeting
```

프로그램 출력:

```text
안녕하세요, Go!
```

### 처리 순서

먼저 `strings.TrimSpace`로 이름 앞뒤의 공백을 정리합니다. 다음으로 결과가 빈 문자열이면 기본 이름 `Go`를 선택합니다. 마지막으로 `fmt.Sprintf`가 만든 인사말을 반환합니다. 화면에 출력하는 일은 `main`이 담당합니다.

`"   "`는 처음에는 빈 문자열이 아니므로, 공백을 정리한 **뒤에** 빈 이름인지 확인해야 합니다. 반면 `"Go learner"`의 가운데 공백은 유지해야 합니다. [strings.TrimSpace](https://pkg.go.dev/strings#TrimSpace)와 [fmt.Sprintf](https://pkg.go.dev/fmt#Sprintf)의 문서를 참고하세요.

함수가 출력 대신 문자열을 반환하므로 테스트는 이름을 전달하고 반환된 문자열을 바로 비교할 수 있습니다. 기존 테스트는 일반 이름, 한글 이름, 앞뒤 공백, 빈 이름, 공백만 있는 이름, 가운데 공백을 포함한 이름을 확인합니다. `main`의 기본 출력도 예제 테스트로 확인합니다.

### 개인 풀이에 테스트 적용하기

이 절은 선택 과제의 `_practice/01-first-program/greeting/main.go`에 `greeting` 함수를 작성한 **후에만** 실행합니다. `main`의 호출도 `greeting("Go")`로 맞춥니다. `message`, `report`, `repair`에는 이 테스트를 복사하지 않습니다.

다음 복사 명령은 개인 `main_test.go`가 이미 있으면 보존합니다. 이전 테스트를 직접 수정했다면 내용을 먼저 확인하세요.

Windows PowerShell:

```powershell
if (-not (Test-Path _practice/01-first-program/greeting/main_test.go)) {
    Copy-Item solutions/01-first-program/greeting/main_test.go _practice/01-first-program/greeting/main_test.go
}
go test -v ./_practice/01-first-program/greeting
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
if [ ! -e _practice/01-first-program/greeting/main_test.go ]; then
    cp solutions/01-first-program/greeting/main_test.go _practice/01-first-program/greeting/main_test.go
fi
go test -v ./_practice/01-first-program/greeting
```

함수를 아직 작성하지 않았다면 해당 이름을 찾지 못해 테스트가 컴파일되지 않습니다. 반환 문자열은 맞는데 `Example_main`만 실패한다면 `main`이 어떤 이름으로 호출하고 무엇을 출력하는지 확인합니다. 테스트 문법 자체는 이후 자동 테스트 장에서 다룹니다.

## 다음 학습으로 연결하기

기본 문제에서는 고정된 문자열과 정수를 직접 출력했습니다. 다음에는 값을 변수에 보관하고 자료형에 따라 어떤 연산을 할 수 있는지 배웁니다. 이후 집필 상태는 [전체 목차](../../book/README.md)에서 확인할 수 있습니다.

| 문제로 돌아가기 | 교재 | 해설 목록 |
| --- | --- | --- |
| [01장 문제](../../exercises/01-first-program/README.md) | [01장 본문](../../book/01-first-program/README.md) | [전체 해설](../README.md) |
