# 01장 연습문제: 첫 Go 프로그램

선수 학습: [01. 첫 Go 프로그램](../../book/01-first-program/README.md)

확인 단계는 01장 본문의 **이해 확인** 질문으로 시작합니다. 아래 문제 1·2는 예제 변형, 문제 3은 선택 응용 과제입니다.

먼저 [개인 연습 폴더 안내](../README.md)를 따라 `/_practice/01-first-program/main.go`를 준비합니다. 아래 문제는 순서대로 풀면서 이 파일을 수정합니다. 명령은 저장소 루트에서 실행합니다.

```text
go run ./_practice/01-first-program
```

## 문제 1. 인사말 바꾸기 — 변형 기초

### 요구 사항

프로그램을 실행하면 다음 한 줄이 출력되도록 작성하세요. 문장 끝의 마침표까지 같아야 합니다.

```text
Go를 공부합니다.
```

입력은 없으며, 터미널에 이 문장과 줄바꿈을 출력하면 됩니다.

### 스스로 확인할 것

- `package main`, `import "fmt"`, `func main()`이 있는가?
- 출력하는 문장이 정확히 일치하는가?
- 실행할 때 컴파일 오류가 없는가?

<details>
<summary>힌트</summary>

첫 예제에서 `fmt.Println`의 괄호 안 문자열만 바꾸면 됩니다. 문자열을 감싸는 큰따옴표는 출력 내용에 포함되지 않습니다.

</details>

## 문제 2. 순서대로 두 줄 출력하기 — 변형 확장

### 요구 사항

다음 두 문장을 순서대로 출력하세요. 두 문장 사이에는 빈 줄이 없어야 합니다.

```text
Go를 공부합니다.
작은 프로그램부터 직접 실행합니다.
```

두 개의 `fmt.Println` 호출을 사용하세요. 이번 문제에서는 문자열 내부의 `\n`을 사용하지 않습니다.

### 스스로 확인할 것

- 첫째 문장과 둘째 문장의 순서가 맞는가?
- 출력 줄 수가 두 줄인가?
- 별도의 빈 문자열을 출력해 불필요한 빈 줄을 만들지 않았는가?

<details>
<summary>힌트</summary>

`main` 함수의 중괄호 안에 출력문 두 개를 작성합니다. `Println`은 각각의 호출 끝에 줄바꿈을 붙이므로 따로 줄바꿈 문자를 넣을 필요가 없습니다.

</details>

## 문제 3. 이름에 맞춰 인사말 만들기 — 응용 선택

이 문제는 함수의 입력·반환값과 문자열 처리를 먼저 경험하는 선택 문제입니다. 처음 배우는 내용이므로 힌트를 읽으며 진행해도 됩니다. 변형 문제 1·2만 먼저 마치고 이후 함수 장을 배운 뒤 돌아와도 좋습니다.

### 요구 사항

`greeting(name string) string` 함수를 만들어 이름에 맞는 인사말을 반환하세요.

1. 입력 이름의 앞뒤 공백을 제거합니다.
2. 공백을 제거한 결과가 빈 문자열이면 이름을 `Go`로 사용합니다.
3. `안녕하세요, 이름!` 형태의 문자열을 반환합니다.
4. `greeting` 함수 안에서는 출력하지 않습니다.
5. `main` 함수는 `greeting("Go")`의 반환값을 출력합니다.

이 문제는 키보드 입력을 읽는 문제가 아닙니다. 먼저 코드에서 함수에 값을 전달하는 방법을 익힙니다.

### 확인할 입력과 결과

| 함수에 전달하는 이름 | 함수가 반환할 문자열 |
| --- | --- |
| `"Go"` | `"안녕하세요, Go!"` |
| `"민수"` | `"안녕하세요, 민수!"` |
| `"  민수  "` | `"안녕하세요, 민수!"` |
| `""` | `"안녕하세요, Go!"` |
| `"   "` | `"안녕하세요, Go!"` |

`main`을 실행했을 때의 예상 출력:

```text
안녕하세요, Go!
```

<details>
<summary>힌트 1: 함수 선언 읽기</summary>

다음은 함수 선언의 형태를 설명하는 **미완성 코드 조각**입니다. 그대로 실행하는 전체 프로그램이 아닙니다.

```go
func greeting(name string) string {
	// name을 사용해 문자열을 만들고 return으로 반환합니다.
}
```

괄호 안의 `name string`은 문자열 입력을 `name`이라는 이름으로 받는다는 뜻입니다. 괄호 뒤의 `string`은 문자열을 반환한다는 뜻입니다. `main`과 별도의 함수로 작성합니다.

</details>

<details>
<summary>힌트 2: 문자열의 앞뒤 공백 제거하기</summary>

표준 라이브러리의 `strings.TrimSpace`를 사용하면 앞뒤 공백이 제거된 문자열을 얻을 수 있습니다. `strings` 패키지를 import에 추가해야 합니다.

```go
name = strings.TrimSpace(name)
```

이 문장은 함수 내부에서 사용할 코드 조각입니다. 공백이 전부인 이름은 정리한 뒤 `""`가 됩니다. [strings.TrimSpace 공식 문서](https://pkg.go.dev/strings#TrimSpace)를 참고하세요.

</details>

<details>
<summary>힌트 3: 빈 이름과 반환값 처리하기</summary>

`if name == ""`로 이름이 비어 있는지 확인하고, 해당할 때 `name = "Go"`로 바꿉니다. 문자열을 만들 때는 `fmt.Sprintf("안녕하세요, %s!", name)`을 사용할 수 있습니다. `%s` 자리에 문자열이 들어갑니다.

`fmt.Sprintf`는 화면에 출력하지 않고 문자열을 돌려줍니다. `greeting`에서는 그 값을 `return`하고, `main`에서 `fmt.Println`으로 출력하세요. [fmt.Sprintf 공식 문서](https://pkg.go.dev/fmt#Sprintf)를 참고합니다.

</details>

### 검증 방법

먼저 `main`에서 `greeting`에 전달하는 이름을 바꾸면서 표의 결과를 확인합니다. 확인한 뒤 기본 호출은 `greeting("Go")`로 돌려놓습니다. 참고 풀이의 테스트 파일을 이용하면 같은 요구 사항을 자동으로 확인할 수도 있습니다.

Windows PowerShell:

```powershell
Copy-Item solutions/01-first-program/greeting/main_test.go _practice/01-first-program/main_test.go
go test -v ./_practice/01-first-program
```

macOS·Linux·WSL:

```bash
cp solutions/01-first-program/greeting/main_test.go _practice/01-first-program/main_test.go
go test -v ./_practice/01-first-program
```

이 테스트 복사는 문제 3의 함수를 작성한 뒤 수행합니다. 문제 1·2만 작성한 상태에는 `greeting` 함수가 없으므로 테스트가 컴파일되지 않습니다. 테스트 파일을 복사할 때 기존의 같은 이름 파일이 있다면 먼저 내용을 확인하세요.

## 풀이 확인

직접 시도했다면 [풀이 안내](../../solutions/README.md)를 읽습니다. 문제 3의 실행 가능한 풀이와 테스트는 [greeting 폴더](../../solutions/01-first-program/greeting)에 있습니다.

풀이를 읽은 뒤에는 다음을 자신의 말로 설명해 보세요.

- 왜 공백 제거를 빈 문자열 검사보다 먼저 했을까?
- 문자열을 만드는 함수와 출력하는 함수를 나누면 무엇을 확인하기 쉬울까?
- 테스트에서 정상적인 이름 외에 빈 이름도 확인하는 이유는 무엇일까?

마지막으로 개인 코드를 정리합니다.

```text
go fmt ./_practice/01-first-program
```

| 교재로 돌아가기 | 문제 목록 |
| --- | --- |
| [01. 첫 Go 프로그램](../../book/01-first-program/README.md) | [전체 연습문제](../README.md) |
