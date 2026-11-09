# 10. 패키지와 모듈

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 | [09. 메서드와 인터페이스](../09-methods-interfaces/README.md) | Go, 편집기, 터미널, 이 저장소 |

지금까지는 한 프로그램을 작은 함수와 타입으로 나눴습니다. 프로그램이 커지면 관련 기능은 여러 파일에 흩어지고, 다른 프로그램에서도 쓰고 싶은 계산이 생깁니다. 모든 코드를 복사하면 수정 내용이 서로 달라질 수 있습니다. 이번 장에서는 **함께 컴파일할 코드, 외부에 공개할 이름, 함께 버전을 관리할 묶음**을 구분합니다.

파일을 나누는 일과 패키지를 나누는 일은 다릅니다. 하나의 패키지를 여러 파일로 나눌 수도 있고, 하나의 모듈에 여러 패키지를 둘 수도 있습니다. 이 세 층을 구별하면 import 오류, 공개 범위, 버전 변경을 훨씬 정확하게 읽을 수 있습니다.

## 학습 목표

- 파일·패키지·모듈의 차이를 실제 디렉터리와 import 경로로 설명한다.
- 공개 이름과 비공개 상태를 구분하고 패키지의 작은 사용 계약을 만든다.
- import 별칭과 internal 경계를 사용하며 순환 의존성을 피한다.
- go.mod와 go.sum의 역할, 버전 선택과 모듈 검사 명령을 구별한다.
- 여러 파일과 패키지로 나눈 프로그램을 실행·검사하고 오류의 원인을 찾는다.

## 사전 지식과 준비

구조체, 메서드, 인터페이스, 슬라이스, 맵과 문자열을 사용합니다. 예제의 성공 여부는 앞 장의 `(값, bool)` 형태를 사용합니다. 실패 원인을 표현하는 `error`는 다음 장에서 본격적으로 다룹니다.

모든 명령은 별도 안내가 없으면 **저장소 루트**에서 실행합니다. Go 명령은 Windows PowerShell, macOS, Linux, WSL에서 동일합니다. 이 장의 기본 예제는 표준 라이브러리와 같은 저장소의 패키지만 사용합니다. 루트의 단일 [`go.mod`](../../go.mod)를 그대로 사용하며 예제 디렉터리 안에 새 모듈을 만들지 않습니다.

| 실행 예제 | 중심 주제 | 실행 명령 |
| --- | --- | --- |
| [multifile](../../examples/10-packages-modules/multifile/main.go) | 같은 패키지의 여러 파일 | `go run ./examples/10-packages-modules/multifile` |
| [greeting](../../examples/10-packages-modules/greeting/main.go) | 다른 패키지의 공개 함수 | `go run ./examples/10-packages-modules/greeting` |
| [private-state](../../examples/10-packages-modules/private-state/main.go) | 필드 숨기기와 변경 메서드 | `go run ./examples/10-packages-modules/private-state` |
| [aliases](../../examples/10-packages-modules/aliases/main.go) | 같은 패키지 이름과 import 별칭 | `go run ./examples/10-packages-modules/aliases` |
| [internal-demo](../../examples/10-packages-modules/internal-demo/main.go) | 내부 패키지의 사용 범위 | `go run ./examples/10-packages-modules/internal-demo` |
| [dependency-direction](../../examples/10-packages-modules/dependency-direction/main.go) | 소비 측 인터페이스와 의존 방향 | `go run ./examples/10-packages-modules/dependency-direction` |
| [initialization](../../examples/10-packages-modules/initialization/main.go) | 패키지 초기화와 main | `go run ./examples/10-packages-modules/initialization` |
| [commands/show](../../examples/10-packages-modules/commands/cmd/show/main.go) | 공유 패키지의 첫 명령 | `go run ./examples/10-packages-modules/commands/cmd/show` |
| [commands/check](../../examples/10-packages-modules/commands/cmd/check/main.go) | 공유 패키지의 둘째 명령 | `go run ./examples/10-packages-modules/commands/cmd/check` |

## 1. 파일·패키지·모듈을 구별하기

### 1.1 세 단위가 담당하는 일

**소스 파일(source file)**은 편집기에서 열어 작성하는 `.go` 파일입니다. **패키지(package)**는 같은 디렉터리에서 함께 컴파일되는 소스의 단위입니다. **모듈(module)**은 함께 버전과 의존성을 관리하는 패키지 모음이며, 루트의 `go.mod`로 식별합니다. [공식 코드 구성 안내](https://go.dev/doc/code)와 [모듈의 정의](https://go.dev/ref/mod#modules-overview)를 참고합니다.

| 단위 | 이 저장소의 예 | 바꾸면 영향을 받는 것 |
| --- | --- | --- |
| 파일 | `greeting/greet/greet.go` | 코드의 배치와 읽는 위치 |
| 패키지 | `greeting/greet/` | import 경로와 공개 API |
| 모듈 | 저장소 루트의 `go.mod` | 공통 import 접두사와 의존 버전 |

파일 이름이 `main.go`라고 해서 모두 같은 프로그램은 아닙니다. 서로 다른 디렉터리에 있는 `package main`들은 별도 패키지입니다. 앞 장의 예제를 동시에 `go test ./...` 할 수 있는 이유입니다.

반대로 같은 디렉터리에 `main.go`, `calculation.go`가 있고 둘 다 `package main`이면 같은 패키지에 속합니다. 함수가 어느 파일에 있는지는 이름의 접근 가능성을 바꾸지 않습니다.

### 1.2 import 경로를 계산하기

이 저장소의 모듈 경로는 다음과 같습니다.

```text
github.com/praster1/GoLang_all_in_one
```

그 아래 `examples/10-packages-modules/greeting/greet/` 패키지의 import 경로는 모듈 경로에 상대 디렉터리를 붙인 것입니다.

```text
github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/greeting/greet
```

| 구성 요소 | 역할 |
| --- | --- |
| `github.com/praster1/GoLang_all_in_one` | 모듈을 식별하는 접두사 |
| `examples/10-packages-modules/greeting/greet` | 모듈 안의 패키지 위치 |
| `package greet` | 소스에서 선언한 패키지 이름 |
| `greet.Hello` | import한 파일에서 접근하는 공개 이름 |

같은 모듈에 있는 패키지를 import할 때도 전체 import 경로를 사용합니다. `import "../greet"`처럼 상대 파일 경로를 쓰는 방식은 이 모듈 기반 학습 과정의 사용법이 아닙니다.

import 경로가 GitHub 주소처럼 보여도 같은 모듈의 현재 작업 파일은 로컬 소스에서 찾습니다. 파일을 수정할 때마다 GitHub에 먼저 업로드해야 실행되는 것이 아닙니다.

### 1.3 패키지 이름과 디렉터리 이름

관례적으로 마지막 디렉터리 이름과 패키지 이름을 맞추지만, 문법적으로 항상 같아야 하는 것은 아닙니다. 뒤의 별칭 예제는 이 차이를 의도적으로 보여 줍니다.

사용하는 코드에서는 import 경로로 패키지를 찾고, 패키지 이름 또는 별칭으로 식별자에 접근합니다. 경로만 보고 사용 이름까지 무조건 추정하지 말고 파일의 `package` 선언을 확인하세요.

## 2. 같은 패키지를 여러 파일로 나누기

### 2.1 main과 계산을 다른 파일에 두기

다음 프로그램은 강의 시간의 합계를 계산합니다. 입력은 이 예제처럼 작은 음이 아닌 정수 목록이라고 가정합니다. 정수 범위를 넘는 임의 크기 집계를 해결하는 예제는 아닙니다.

저장 경로: [`examples/10-packages-modules/multifile/calculation.go`](../../examples/10-packages-modules/multifile/calculation.go).

```go
package main

// totalMinutes adds a small list of nonnegative lesson durations.
func totalMinutes(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}
```

저장 경로: [`examples/10-packages-modules/multifile/main.go`](../../examples/10-packages-modules/multifile/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	fmt.Println("총 시간:", totalMinutes([]int{20, 30, 10}))
	fmt.Println("빈 입력:", totalMinutes(nil))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/multifile
```

예상 출력:

```text
총 시간: 60
빈 입력: 0
```


`main.go`에서 `totalMinutes` 앞에 패키지 이름을 붙이지 않았습니다. 두 파일은 같은 패키지에 속하므로 서로의 패키지 수준 선언을 사용합니다. 이 함수는 이름이 소문자로 시작하지만 같은 패키지 안에서는 사용할 수 있습니다.

### 2.2 파일 하나보다 패키지 디렉터리를 실행하기

이 예제는 다음 명령으로 실행합니다.

```sh
go run ./examples/10-packages-modules/multifile
```

명령의 대상은 패키지 디렉터리입니다. Go 도구가 해당 빌드에 포함되는 소스 파일을 함께 읽습니다.

다음은 **의도적으로 잘못 지정한 실행 명령**입니다.

```sh
go run ./examples/10-packages-modules/multifile/main.go
```

파일 목록을 직접 전달하면 지정한 파일만으로 임시 패키지를 구성하므로 `calculation.go`의 함수가 함께 들어오지 않습니다. 오류에 `undefined: totalMinutes`가 표시됩니다. 함수 이름의 오타만 의심하기 전에 실행 대상을 확인하세요. [go 명령의 패키지 인자 규칙](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)을 참고합니다.

같은 패키지의 모든 파일에서 import가 공유되는 것은 아닙니다. `fmt`를 사용하는 `main.go`는 직접 `import "fmt"`를 적습니다. `calculation.go`가 `fmt`를 사용한다면 그 파일도 자신의 import를 선언해야 합니다. 패키지 수준 선언의 스코프와 import 이름의 파일 스코프는 다릅니다.

### 2.3 두 main 함수를 한 폴더에 넣지 않기

같은 패키지에 같은 이름의 일반 함수 두 개를 선언할 수 없습니다. 서로 다른 예제를 한 폴더에 복사하면 `main redeclared` 오류가 날 수 있습니다.

| 상황 | 결과 |
| --- | --- |
| 같은 폴더의 main.go와 calculation.go에 서로 다른 함수 | 함께 컴파일 |
| 같은 폴더의 두 파일에 각각 func main | 이름 중복 |
| 서로 다른 폴더에 각각 package main과 func main | 별도 프로그램 |
| 같은 일반 소스 폴더에 package main과 package greet 혼합 | 패키지 이름 충돌 |

외부 테스트용 `package name_test`는 뒤의 12장에서 배울 특별한 테스트 구성입니다. 일반 소스에서 패키지 이름을 섞어도 된다는 뜻은 아닙니다.

## 3. 별도 패키지의 공개 함수 사용하기

### 3.1 출력과 인사 문자열 만들기 분리

인사 문자열을 만드는 기능을 `greet` 패키지로 옮깁니다. 이 패키지는 문자열을 반환하고 터미널에 직접 출력하지 않습니다. 같은 결과를 웹 응답이나 파일에 쓸 수 있게 하기 위한 선택입니다.

저장 경로: [`examples/10-packages-modules/greeting/greet/greet.go`](../../examples/10-packages-modules/greeting/greet/greet.go).

```go
// Package greet formats short greetings without producing output.
package greet

import "strings"

// Hello returns a greeting after trimming surrounding spaces.
func Hello(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "방문자"
	}
	return "안녕하세요, " + name + "!"
}
```

저장 경로: [`examples/10-packages-modules/greeting/main.go`](../../examples/10-packages-modules/greeting/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/greeting/greet"
)

func main() {
	fmt.Println(greet.Hello(" Go "))
	fmt.Println(greet.Hello(""))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/greeting
```

예상 출력:

```text
안녕하세요, Go!
안녕하세요, 방문자!
```


`Hello`는 대문자로 시작하는 공개 함수입니다. 다른 패키지가 `greet.Hello`로 호출할 수 있습니다. 함수에 전달한 이름의 앞뒤 공백은 제거하며, 남은 이름이 없으면 `방문자`를 사용합니다.

`main`은 결과를 어떻게 보여 줄지 정합니다. 기능 패키지는 문자열을 만들고 호출자는 출력합니다. 이처럼 패키지의 입력과 결과가 작고 명확하면 사용법과 테스트를 이해하기 쉽습니다.

### 3.2 공개 이름은 사용자에게 하는 약속

Go에서 다른 패키지가 직접 접근할 수 있는 식별자는 공개(exported) 조건을 만족해야 합니다. 보통 첫 글자를 대문자로 쓰며, 정확한 규칙은 유니코드 대문자와 선언 위치를 포함합니다. [공개 식별자 명세](https://go.dev/ref/spec#Exported_identifiers)를 참고합니다.

```go
func Hello(name string) string
```

이 선언을 공개하면 호출자는 이름, 인자 타입, 결과 타입과 문서화된 동작에 의존합니다. 나중에 결과를 `[]byte`로 바꾸면 기존 호출 코드는 영향을 받습니다. 파일 이름을 바꾸는 일보다 공개 함수의 계약을 바꾸는 일이 더 큰 경계 변화인 이유입니다.

공개 함수와 타입에는 이름으로 시작하는 설명 주석을 작성하는 관례가 있습니다. 예제의 `// Hello returns ...`가 그 역할을 합니다. `go doc`은 이런 주석을 사용 안내로 보여 줍니다.

```sh
go doc ./examples/10-packages-modules/greeting/greet
go doc ./examples/10-packages-modules/greeting/greet Hello
```

출력의 경로나 서식은 도구 버전에 따라 달라질 수 있지만 `Hello`의 시그니처와 설명을 확인할 수 있습니다. 로컬 문서 확인에는 외부 사이트에 먼저 게시할 필요가 없습니다.

### 3.3 패키지는 폴더로 나누고 사용 이유로 묶기

함수 하나마다 새 패키지를 만들면 독자는 import를 오가느라 개념의 전체 흐름을 놓치기 쉽습니다. 반대로 `utils`라는 큰 패키지에 관련 없는 것을 모으면 이름만 보고 책임을 알기 어렵습니다.

다음 질문으로 분리 기준을 정합니다.

- 같은 데이터와 규칙을 함께 설명할 수 있나요?
- 다른 패키지에 공개해야 하는 기능은 무엇인가요?
- 구현을 바꾸어도 호출자가 유지해야 하는 약속은 무엇인가요?
- 새 패키지를 만들면 의존 방향이 더 분명해지나요?

이는 언어가 강제하는 규칙이 아니라 이 교재의 설계 판단입니다. 작은 프로그램은 여러 파일을 가진 하나의 패키지로도 충분합니다. [공식 모듈 구성 안내](https://go.dev/doc/modules/layout)는 프로그램 규모에 따른 여러 구성을 보여 줍니다.

## 4. 공개 타입과 비공개 상태

### 4.1 타입을 공개해도 모든 필드를 공개할 필요는 없습니다

다음 `Counter`는 외부에서 만들 수 있지만, 내부 정수 필드는 직접 대입할 수 없습니다. 상태 변경을 `Add` 메서드에 모으기 위한 구성입니다.

저장 경로: [`examples/10-packages-modules/private-state/counter/counter.go`](../../examples/10-packages-modules/private-state/counter/counter.go).

```go
// Package counter keeps a count in the range 0..100.
package counter

// Counter has a useful zero value. The count is changed through Add.
type Counter struct{ value int }

func (c Counter) Value() int { return c.value }
func (c *Counter) Add(amount int) bool {
	if c == nil || amount < 0 || amount > 100-c.value {
		return false
	}
	c.value += amount
	return true
}
```

저장 경로: [`examples/10-packages-modules/private-state/main.go`](../../examples/10-packages-modules/private-state/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/private-state/counter"
)

func main() {
	var c counter.Counter
	fmt.Println("추가:", c.Add(30))
	fmt.Println("초과:", c.Add(80))
	fmt.Println("현재:", c.Value())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/private-state
```

예상 출력:

```text
추가: true
초과: false
현재: 30
```


| 이름 | 외부에서 사용 | 이유 |
| --- | --- | --- |
| `counter.Counter` | 가능 | 공개 타입 |
| `c.Value()` | 가능 | 공개 읽기 메서드 |
| `c.Add(30)` | 가능 | 공개 변경 메서드 |
| `c.value = 200` | 불가능 | 비공개 필드 |

`value`가 소문자로 시작한다는 사실은 같은 패키지의 다른 파일까지 막는 것이 아닙니다. **패키지가 비공개 상태의 경계**입니다. `counter` 내부 파일들은 그 필드에 접근할 수 있습니다.

영값 `counter.Counter{}`는 값 0으로 사용할 수 있습니다. 따라서 이 예제는 생성자 없이 시작해도 안전합니다. 모든 타입이 반드시 `New` 함수로 만들어져야 하는 것은 아닙니다.

### 4.2 내부 상태와 사용 계약

`Add`가 성공하는 조건은 0 이상의 증가량이며 합계가 100 이하인 것입니다. 기존 값은 외부에서 잘못 대입할 수 없고 정상 메서드는 범위를 유지합니다. 실패한 메서드는 필드를 바꾸지 않습니다.

`Value`는 값 수신자 메서드이므로 nil `*Counter`에 호출하는 사용법은 지원하지 않습니다. `Add`는 포인터 수신자 메서드이며 nil을 명시적으로 거부합니다. 공개 메서드라고 해서 nil 처리 방식이 모두 같아야 하는 것은 아니지만 계약은 분명해야 합니다.

비공개 필드는 외부 호출자가 규칙을 우회하기 어렵게 합니다. 이것은 보안 암호화나 악성 코드 격리를 뜻하지 않습니다. 정상 Go 코드 사이의 패키지 접근 규칙입니다.

### 4.3 생성자가 필요한 경우

어떤 타입은 영값으로 사용할 수 없거나, 초기 입력 검증이 필요합니다. 그런 타입에는 관례적으로 `New` 또는 `New이름` 함수를 제공합니다. 다음 의존 방향 예제의 `order.New`가 수량과 가격을 확인합니다.

`New`는 Go 키워드가 아니라 개발자가 정한 함수 이름입니다. 컴파일러가 특별한 생성자 실행 규칙을 부여하지 않습니다. 필드가 비공개라도 `Type{}` 형태로 영값을 만드는 것은 가능하므로, 영값을 어떻게 다룰지 함께 설계하세요.

## 5. import 별칭과 이름 충돌

### 5.1 경로가 다르고 패키지 이름이 같은 경우

아래 두 폴더는 학습을 위해 둘 다 `package format`을 선언합니다.

저장 경로: [`examples/10-packages-modules/aliases/brief/format.go`](../../examples/10-packages-modules/aliases/brief/format.go).

```go
package format

func Title(name string) string { return "[" + name + "]" }
```

저장 경로: [`examples/10-packages-modules/aliases/detail/format.go`](../../examples/10-packages-modules/aliases/detail/format.go).

```go
package format

import "fmt"

func Title(name string, count int) string { return fmt.Sprintf("%s: %d개", name, count) }
```

저장 경로: [`examples/10-packages-modules/aliases/main.go`](../../examples/10-packages-modules/aliases/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	brief "github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/aliases/brief"
	detail "github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/aliases/detail"
)

func main() {
	fmt.Println(brief.Title("메모"))
	fmt.Println(detail.Title("메모", 3))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/aliases
```

예상 출력:

```text
[메모]
메모: 3개
```


`brief`, `detail`은 이 파일에서 정한 import 별칭(alias)입니다. 다른 패키지의 실제 선언 이름이나 공개 API를 수정하지 않습니다. 어떤 구현인지 명확히 하고 이름 충돌을 피합니다.

일반적으로 패키지 이름과 폴더 이름을 맞추면 이런 설명이 덜 필요합니다. 이 예제의 불일치는 두 이름이 별개라는 사실을 관찰하기 위한 선택입니다.

### 5.2 사용하지 않는 import

일반 import는 해당 파일에서 사용해야 합니다. 필요 없는 import를 남기면 컴파일 오류가 발생합니다. 기능을 지운 뒤 `fmt`가 더 이상 필요하지 않다면 import도 지웁니다.

별칭을 `_`로 두는 빈 import는 이름을 사용하지 않고 패키지 초기화 효과만 가져올 때 사용합니다. 뒤의 SQL 장에서 드라이버 등록을 설명할 때 등장합니다. 여기서는 사용하지 않는 import 오류를 없애려고 무조건 `_`를 붙이지 않습니다.

점 import `import . "..."`는 공개 이름을 접두사 없이 사용할 수 있게 하지만 이름의 출처가 흐려지고 충돌하기 쉽습니다. 이 교재의 기본 예제에서는 사용하지 않습니다. 문법의 존재와 권장 사용법은 구분하세요. [Import declarations](https://go.dev/ref/spec#Import_declarations)를 참고합니다.

## 6. internal로 사용 범위 한정하기

### 6.1 공개 이름과 공개 패키지는 별개입니다

`internal`이라는 디렉터리 이름은 Go 도구에서 특별한 접근 규칙을 가집니다. internal 아래 패키지는 그 internal 디렉터리의 부모를 기준으로 허용된 소스 트리 안에서만 import할 수 있습니다. 단순히 이름에 `internal` 글자가 포함된 아무 폴더가 아니라 경로 요소가 정확히 `internal`인 경우입니다. [Internal packages](https://pkg.go.dev/cmd/go#hdr-Internal_packages)를 참고합니다.

저장 경로: [`examples/10-packages-modules/internal-demo/internal/token/token.go`](../../examples/10-packages-modules/internal-demo/internal/token/token.go).

```go
package token

import "strings"

func Normalize(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
```

저장 경로: [`examples/10-packages-modules/internal-demo/main.go`](../../examples/10-packages-modules/internal-demo/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/internal-demo/internal/token"
)

func main() { fmt.Println("키:", token.Normalize(" GO ")) }
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/internal-demo
```

예상 출력:

```text
키: go
```


이 예제에서 `internal`의 부모는 `internal-demo/`입니다. 그 아래의 `main.go`는 `token`을 import할 수 있습니다. 반면 같은 저장소라도 `examples/01-first-program/hello/`처럼 이 부모 트리 밖의 패키지는 import할 수 없습니다.

| import하는 위치 | 이 token 패키지 사용 |
| --- | --- |
| `internal-demo/` | 허용 |
| `internal-demo/cmd/tool/` | 허용 |
| `examples/01-first-program/hello/` | 허용되지 않음 |
| 별도 외부 모듈 | 허용되지 않음 |

`Normalize`가 대문자인 이유는 다른 **허용된 패키지**가 호출해야 하기 때문입니다. internal은 패키지를 import할 수 있는 범위를 제한하고, 소문자 이름은 패키지 안의 선언 접근을 제한합니다. 둘은 다른 층의 규칙입니다.

### 6.2 경계를 실제로 확인하기

잘못된 import를 정상 예제에 추가하지 마세요. 오류를 직접 보려면 개인 실습 폴더에서 위 token 패키지를 import한 코드를 만들어 봅니다. `_practice/10-packages-modules/...`는 `internal-demo`의 부모 트리 밖이므로 `use of internal package ... not allowed` 종류의 오류가 발생합니다.

오류를 해결하려고 internal이라는 이름만 지우기 전에, 정말 외부 API로 공개하려는 기능인지 판단합니다. 구현 세부사항이면 허용된 위치에서 사용하거나 공개 API를 통해 간접 호출하는 구성이 맞을 수 있습니다.

## 7. 의존 방향과 소비 측 인터페이스

### 7.1 양방향 import는 허용되지 않습니다

패키지 A가 B를 import하고 B가 A를 import하면 순환 의존성(import cycle)이 됩니다. 사이에 다른 패키지가 있어도 다시 출발점으로 돌아오면 같은 문제입니다. Go 패키지는 직접·간접적으로 자신을 import할 수 없습니다. [Import declarations](https://go.dev/ref/spec#Import_declarations)를 참고합니다.

순환이 생겼을 때 가장 먼저 할 일은 `init`이나 전역 변수로 연결을 숨기는 것이 아닙니다. 어느 패키지가 어느 기능을 실제로 필요로 하는지 그려 봅니다.

아래 예제는 주문 값을 계산하는 `order`와 한 줄을 만드는 `report`를 `main`이 연결합니다.

저장 경로: [`examples/10-packages-modules/dependency-direction/order/order.go`](../../examples/10-packages-modules/dependency-direction/order/order.go).

```go
package order

// Basket stores a validated quantity and unit price for this exercise.
type Basket struct{ quantity, price int }

func New(quantity, price int) (Basket, bool) {
	if quantity < 0 || quantity > 1000 || price < 0 || price > 1000000 {
		return Basket{}, false
	}
	return Basket{quantity: quantity, price: price}, true
}
func (b Basket) Total() int { return b.quantity * b.price }
```

저장 경로: [`examples/10-packages-modules/dependency-direction/report/report.go`](../../examples/10-packages-modules/dependency-direction/report/report.go).

```go
package report

import "fmt"

// Totaler describes the single behavior the report needs.
type Totaler interface{ Total() int }

func Summary(value Totaler) string { return fmt.Sprintf("합계: %d원", value.Total()) }
```

저장 경로: [`examples/10-packages-modules/dependency-direction/main.go`](../../examples/10-packages-modules/dependency-direction/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/dependency-direction/order"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/dependency-direction/report"
)

func main() {
	basket, ok := order.New(3, 2000)
	if !ok {
		fmt.Println("입력이 잘못되었습니다")
		return
	}
	fmt.Println(report.Summary(basket))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/dependency-direction
```

예상 출력:

```text
합계: 6000원
```


`report`는 주문 타입의 이름과 내부 필드를 모릅니다. 합계를 얻는 `Total() int` 동작만 필요합니다. 그래서 소비하는 쪽인 `report`가 작은 `Totaler` 인터페이스를 선언했습니다.

| 패키지 | 의존하는 것 | 모르는 것 |
| --- | --- | --- |
| `order` | 기본 타입 | report의 표시 방식 |
| `report` | fmt와 Totaler 계약 | order.Basket의 내부 구조 |
| `main` | order, report | 필요 이상으로 내부 상태를 수정하지 않음 |

`order.Basket`은 맞는 메서드를 제공하므로 별도 등록 없이 `report.Totaler`를 구현합니다. 09장에서 배운 암시적 인터페이스 구현이 패키지 경계에서도 작동합니다.

### 7.2 모든 순환을 인터페이스로 풀 필요는 없습니다

두 패키지가 사실상 같은 책임의 절반이라면 합치는 것이 더 단순할 수 있습니다. 공통 데이터 구조가 독립된 의미를 가진다면 더 낮은 공통 패키지로 옮길 수도 있습니다. 아주 작은 프로그램이면 한 패키지 안의 여러 파일로 시작하는 것도 합리적입니다.

선택 기준은 파일 수를 늘리는 것이 아니라 책임과 의존 방향을 설명할 수 있는가입니다. 이 예제는 서로 다른 계산 구현을 같은 보고서에 전달할 수 있으므로 인터페이스가 실제 교체 지점을 나타냅니다.

`report.Summary(nil)`이나 계약을 지키지 않는 구현까지 안전하게 처리한다고 약속하지 않았습니다. 입력 계약이 필요하면 그 경계에서 별도로 검사해야 합니다.

### 7.3 의존 목록 직접 확인하기

```sh
go list -f '{{.ImportPath}}: {{join .Imports ", "}}' ./examples/10-packages-modules/dependency-direction/...
```

출력에서 `report`가 `order`를 import하지 않고, `order`도 `report`를 import하지 않는지 확인합니다. Go 템플릿의 자세한 문법을 지금 외울 필요는 없습니다. `go list`가 실제 도구 관점의 패키지 정보를 보여 준다는 점을 기억하세요.

## 8. 패키지 초기화와 init

### 8.1 main 이전에 준비되는 값

패키지 수준 변수는 프로그램 실행 준비 과정에서 초기화됩니다. 해당 패키지의 변수 초기화 이후 `init` 함수들이 실행되며, import한 패키지가 준비된 뒤 이를 사용하는 패키지가 초기화됩니다. 전체 패키지 초기화가 끝나고 main 패키지의 `main`이 실행됩니다. [Program initialization and execution](https://go.dev/ref/spec#Program_initialization_and_execution)을 참고합니다.

저장 경로: [`examples/10-packages-modules/initialization/settings/settings.go`](../../examples/10-packages-modules/initialization/settings/settings.go).

```go
package settings

var base = 2
var Limit = base * 3

func init() { Limit++ }
```

저장 경로: [`examples/10-packages-modules/initialization/main.go`](../../examples/10-packages-modules/initialization/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/initialization/settings"
)

func main() { fmt.Println("초기화된 한도:", settings.Limit) }
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/initialization
```

예상 출력:

```text
초기화된 한도: 7
```


이 예제에서는 `base`가 2, `Limit`이 6으로 초기화된 뒤 `init`이 1을 더합니다. 따라서 `main`에서 보는 값은 7입니다.

`init`은 직접 호출하는 일반 API가 아닙니다. 인자와 결과를 선언하지 않으며, 컴파일된 패키지의 초기화 과정에서 도구가 실행합니다. 같은 패키지에 여러 `init`을 둘 수 있지만 그 사실을 복잡한 설정 순서의 설계 도구로 삼지 않습니다.

### 8.2 자동 실행과 명시적 준비 함수의 차이

초기화 중 파일을 열거나 네트워크에 접속하면 호출자가 실패와 설정을 제어하기 어려워질 수 있습니다. 실제 설정은 다음 장에서 배울 오류 반환을 가진 준비 함수에서 처리하고, main이 실패 여부를 결정하는 구성이 읽기 쉽습니다.

이 장의 `init` 예제는 규칙을 관찰하기 위한 최소 사례입니다. 모든 프로그램에 전역 변수와 init을 추가하라는 권장사항은 아닙니다. 실제 카운터 예제처럼 인스턴스별 상태가 필요한 경우 지역 변수와 명시적인 생성 흐름을 사용합니다.

여러 파일에 걸친 서로 관련 없는 초기화 순서에 프로그램의 의미를 의존시키지 마세요. 초기화에 순서가 필요하다면 그 관계를 데이터 의존성 또는 명시적 함수 호출로 나타냅니다.


### 8.3 명시적인 조립: 두 명령이 규칙을 공유하기

설정과 객체 생성을 init에 숨기지 않고 main에서 명시적으로 수행하면, 같은 패키지를 여러 실행 프로그램이 사용하기 쉽습니다. 이번에는 학습 시간의 입력 규칙을 하나의 패키지에 두고 표시 명령과 검사 명령에서 함께 사용합니다.

구성은 다음과 같습니다.

| 경로 | 역할 |
| --- | --- |
| `commands/internal/lesson/` | 제목·시간 규칙과 한 줄 요약 |
| `commands/cmd/show/` | 유효한 강의를 만들고 요약 출력 |
| `commands/cmd/check/` | 잘못된 입력의 유효성 확인 |

`cmd`라는 이름은 여러 실행 프로그램을 정리할 때 자주 사용하는 관례입니다. 컴파일러가 이름 자체에 특별한 접근 권한을 부여하는 것은 아닙니다. `internal`은 실제 도구의 접근 규칙이 있는 디렉터리라는 차이를 구별하세요.

공유 기능은 다음 파일에 있습니다.

저장 경로: [`examples/10-packages-modules/commands/internal/lesson/lesson.go`](../../examples/10-packages-modules/commands/internal/lesson/lesson.go).

```go
// Package lesson defines the shared input policy of two small commands.
package lesson

import (
	"fmt"
	"strings"
)

type Lesson struct {
	title   string
	minutes int
}

func New(title string, minutes int) (Lesson, bool) {
	title = strings.TrimSpace(title)
	if title == "" || minutes < 1 || minutes > 240 {
		return Lesson{}, false
	}
	return Lesson{title: title, minutes: minutes}, true
}
func (value Lesson) Summary() string { return fmt.Sprintf("%s: %d분", value.title, value.minutes) }
```

제목의 앞뒤 공백을 제거하고 시간이 1~240분인지 확인합니다. 이 범위는 두 명령이 공유하는 학습용 정책입니다. 구현을 각각 main에 복사하지 않으므로 경계를 바꾸면 같은 규칙을 쓰는 호출자들을 함께 검증할 수 있습니다.

`Lesson`의 필드는 비공개입니다. `New`가 유효한 값인지 확인하고, `Summary`가 표시 문자열을 만듭니다. 여기서 영값 Lesson은 초기화되지 않은 상태로서 유효한 강의라는 계약을 갖지 않습니다. 사용자는 New의 bool을 확인하고 성공한 값을 사용해야 합니다.

이 타입이 학습용으로 Summary만 제공하는 것과, 모든 프로그램에서 생성자가 필요하다는 주장은 다릅니다. 앞의 Counter는 영값이 유효했습니다. 타입별로 계약을 정합니다.

첫 번째 실행 프로그램은 다음과 같습니다.

저장 경로: [`examples/10-packages-modules/commands/cmd/show/main.go`](../../examples/10-packages-modules/commands/cmd/show/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/commands/internal/lesson"
)

func main() {
	value, ok := lesson.New(" Go ", 30)
	if !ok {
		fmt.Println("입력 오류")
		return
	}
	fmt.Println(value.Summary())
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/commands/cmd/show
```

예상 출력:

```text
Go: 30분
```

두 번째 프로그램도 같은 패키지를 import합니다.

저장 경로: [`examples/10-packages-modules/commands/cmd/check/main.go`](../../examples/10-packages-modules/commands/cmd/check/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/commands/internal/lesson"
)

func main() {
	_, ok := lesson.New("Go", 0)
	fmt.Println("유효:", ok)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/10-packages-modules/commands/cmd/check
```

예상 출력:

```text
유효: false
```

`show`와 `check`는 각자 package main과 main 함수를 갖습니다. 디렉터리가 다르므로 이름 충돌이 없습니다. 두 실행 파일이 같은 모듈에 있다는 이유로 하나의 프로세스에서 자동으로 함께 실행되는 것은 아닙니다.

### 8.4 여러 실행 프로그램의 빌드 범위

두 프로그램과 공유 패키지를 한 번에 검사합니다.

```sh
go test ./examples/10-packages-modules/commands/...
go build ./examples/10-packages-modules/commands/...
```

여러 패키지를 대상으로 하는 위 `go build`는 각 패키지가 빌드 가능한지 확인합니다. 두 이름의 실행 파일을 현재 폴더에 자동으로 모두 배포해 주는 명령으로 이해하지 않습니다. 실행 산출물을 만들 때는 하나의 main 패키지를 선택하고 출력 파일명을 정합니다. 운영체제별 산출물 생성은 28장에서 자세히 다룹니다.

단일 main을 빌드할 때도 공유 internal 패키지를 먼저 따로 빌드할 필요는 없습니다. Go 도구가 import 의존 관계를 따라 필요한 패키지를 함께 처리합니다. 사람이 파일 순서를 나열하여 실행 순서를 맞추는 방식이 아닙니다.

### 8.5 같은 모듈이 보장하지 않는 것

| 사실 | 자동으로 따라오지 않는 보장 |
| --- | --- |
| 두 main이 같은 패키지를 import | 실행 중 메모리 상태 공유 |
| 같은 모듈에서 함께 빌드 | 같은 프로세스로 실행 |
| 같은 파일의 입력 검증 사용 | 서로 다른 실행 파일의 배포 버전이 항상 같음 |
| 공유 규칙 테스트 통과 | 모든 운영 환경의 설정까지 검증 |

이 구별은 나중에 CLI와 HTTP 서버를 같은 저장소에서 만들 때 도움이 됩니다. 코드 재사용과 실행 시점의 상태 공유는 다른 문제입니다.

패키지의 내부 전역 변수에 값을 보관해도 서로 다른 실행 프로그램 사이에서 메모리가 공유되지는 않습니다. 공통 상태가 필요하다면 파일·데이터베이스·네트워크 등 실제 공유 매체를 설계해야 합니다. 그 주제는 입출력과 서비스 장에서 이어집니다.

### 8.6 값을 바꾸며 확인하기

| 바꿔 볼 입력 | 실행 전 예상 | 확인할 내용 |
| --- | --- | --- |
| 시간 0 → 1 | 실패에서 성공으로 바뀜 | 하한 포함 |
| 시간 240 → 241 | 성공에서 실패로 바뀜 | 상한 포함 |
| 제목 Go → 공백 | 실패 | 제목 정규화 뒤 검사 |
| show에서 New 호출 제거 후 영값 사용 | 문법상 가능하나 사용 계약 위반 | 생성자의 존재와 언어 강제를 구별 |

기본 예제를 직접 덮어쓰지 말고 개인 폴더의 실습에서 입력을 바꿔 봅니다. 두 명령이 어느 공유 패키지를 import하는지 확인하고, 개인 복사본을 만든 경우 경로도 함께 조정합니다.

## 9. go.mod를 읽는 방법

### 9.1 모듈 이름과 Go 버전

현재 저장소의 루트 `go.mod`를 편집기로 열어 봅니다. 앞부분에는 모듈 경로와 Go 버전 조건이 있습니다.

```go
module github.com/praster1/GoLang_all_in_one

go 1.27.0
```

이 블록은 Go 소스 코드가 아니라 `go.mod`의 형식입니다. `.go` 파일에 붙여 넣지 않습니다. 실제 의존 목록은 뒤의 장에서 사용하는 드라이버 때문에 더 길 수 있으므로 현재 루트 파일을 기준으로 확인합니다.

`go` 지시문은 모듈의 최소 Go 버전 요구와 언어 버전 선택에 영향을 줍니다. 특정 운영체제 이름이나 라이브러리 버전을 의미하지 않습니다. `toolchain` 지시문이 있다면 주 모듈을 작업할 때 사용할 권장 도구 체인 선택에도 영향을 줍니다. [Go Toolchains](https://go.dev/doc/toolchain)를 참고합니다.

```sh
go version
go env GOMOD GOWORK GOTOOLCHAIN
go list -m
```

- `go version`: 실제 선택되어 실행 중인 도구 체인 버전.
- `GOMOD`: 현재 디렉터리에서 적용되는 go.mod 경로.
- `GOWORK`: 적용 중인 작업 공간 파일 위치 또는 비활성 상태.
- `GOTOOLCHAIN`: 도구 체인 선택 설정.
- `go list -m`: 주 모듈의 경로.

경로와 일부 출력은 운영체제·사용자 설정에 따라 달라집니다. 저장소 밖의 엉뚱한 `go.mod`나 `go.work`를 읽고 있지 않은지 확인하는 것이 목적입니다.

### 9.2 require와 간접 의존성

`require`에는 필요한 모듈과 버전이 기록됩니다. `// indirect`는 주 모듈의 패키지에서 직접 import하지 않는 의존성에 붙을 수 있습니다. 정확한 포함 조건은 모듈 그래프와 Go 버전에 따른 규칙을 따릅니다. 주석이 있다고 해서 필요 없는 파일이거나 삭제해도 된다는 뜻은 아닙니다.

```sh
go list -m all
go mod graph
```

첫 명령은 선택된 모듈 목록을, 둘째 명령은 모듈 요구 관계를 보여 줍니다. 의존성 메타데이터가 로컬 캐시에 없으면 네트워크 접근이 필요할 수 있습니다. 이 장의 기본 실행 예제 자체는 외부 Go 패키지를 추가로 요구하지 않습니다.

다른 장이 의존하는 모듈을 지워서 이 장만 깔끔하게 만드는 것은 전체 교재를 깨뜨릴 수 있습니다. 단일 모듈 저장소에서는 전체 사용 범위를 함께 봅니다.

### 9.3 최소 버전 선택의 의미

Go의 기본 모듈 버전 선택은 최소 버전 선택(minimal version selection, MVS) 규칙을 사용합니다. 같은 모듈에 대해 그래프가 요구하는 버전들을 충족하는 선택을 구성하며, 단순히 인터넷의 최신 버전을 매번 고르는 것이 아닙니다. [MVS](https://go.dev/ref/mod#minimal-version-selection)를 참고합니다.

예를 들어 어떤 의존 경로가 가상의 `example.com/math` v1.2.0을, 다른 경로가 v1.4.0을 요구한다면 같은 모듈 경로의 선택은 v1.4.0 쪽으로 올라갑니다. 이것은 설명용 가상 그래프이며 해당 주소에 실제 패키지가 있다는 뜻은 아닙니다.

| 오해 | 실제 확인할 점 |
| --- | --- |
| require의 버전은 언제나 최종 고정 버전 | 다른 요구 관계가 더 높은 버전을 선택할 수 있음 |
| 빌드 때마다 최신 버전으로 자동 갱신 | 기록된 그래프와 도구 규칙에 따라 선택 |
| 최신 패치면 검증이 필요 없음 | 공개 계약 변화와 회귀 가능성은 테스트로 확인 |
| 하위 모듈은 각각 자기 의존 버전을 동시에 사용 | 같은 모듈 경로의 선택은 전체 빌드 그래프에서 조정 |

`replace`, `exclude`, 작업 공간과 주요 버전이 추가되면 더 많은 규칙이 있습니다. 이번 장에서는 기본 그래프를 읽는 것부터 익힙니다.

## 10. go.sum, tidy, verify의 서로 다른 역할

### 10.1 go.sum은 의존성 무결성 자료

`go.sum`에는 모듈 내용과 메타데이터의 체크섬이 기록됩니다. 다운로드한 내용이 기대한 내용과 일치하는지 확인하는 데 사용합니다. 파일에 한 모듈의 여러 버전이 보일 수 있고, 모든 줄이 현재 빌드에서 선택된 모듈이라는 뜻은 아닙니다. [go.sum files](https://go.dev/ref/mod#go-sum-files)를 참고합니다.

`go.sum`을 다른 생태계의 잠금 파일과 완전히 같은 의미라고 설명하면 버전 선택을 오해하기 쉽습니다. 버전 요구와 선택은 `go.mod`와 모듈 그래프를 함께 봅니다.

### 10.2 검사와 변경을 구별하기

| 명령 | 주된 목적 | 주의할 점 |
| --- | --- | --- |
| `go list -m all` | 선택된 의존 모듈 확인 | 필요한 메타데이터를 내려받을 수 있음 |
| `go mod why -m 모듈경로` | 왜 필요한지 확인 | 테스트 의존도 관련될 수 있음 |
| `go mod tidy` | 실제 소스·테스트에 맞게 의존 기록 정리 | go.mod와 go.sum 변경 가능 |
| `go mod verify` | 로컬 모듈 캐시 내용의 변경 검사 | 코드 보안성·테스트 성공을 보장하지 않음 |
| `go test ./...` | 현재 모듈 패키지의 테스트 실행 | 실행한 조건에서의 동작 검증 |

명령은 각각 다른 질문에 답합니다. `verify`가 성공했다고 새 버전의 동작이 호환된다는 뜻은 아닙니다. `tidy`가 성공했다고 기능이 올바르다는 뜻도 아닙니다.

### 10.3 의존성을 바꾼 뒤 검토하기

일반적인 의존 변경에서는 이유와 목표 버전을 정한 뒤 변경하고, diff와 전체 테스트를 함께 봅니다. `go.mod`와 `go.sum`을 모두 버전 관리합니다.

이 장의 예제를 실행하려고 루트에서 무작정 `go mod init`이나 `go get ...@latest`를 실행할 필요는 없습니다. 이미 모듈이 준비되어 있습니다. 새 모듈을 만드는 실습은 아래의 분리된 임시 프로젝트에서 진행합니다.

캐시를 수동으로 지우거나 체크섬을 지워 오류를 숨기기 전에, 버전·네트워크·모듈 경로·실제 오류 메시지를 확인하세요. 재현 정보를 없애면 원인 분석이 더 어려워집니다.

## 11. 별도 실습: 새 모듈을 만들 때

### 11.1 현재 저장소와 분리된 위치 선택

모듈을 직접 만드는 과정을 연습하고 싶다면 **이 저장소 바깥의 새 빈 디렉터리**를 선택합니다. 기존 폴더를 덮어쓰지 않습니다. 여기서는 편집기나 파일 탐색기로 `go-package-practice`라는 빈 폴더를 준비했다고 가정합니다.

그 새 폴더의 터미널에서 실행합니다.

```sh
go mod init example.com/learn/packages
```

`example.com/learn/packages`는 연습용 모듈 이름입니다. 이 명령 자체가 인터넷에 저장소를 만들거나 코드를 공개하지 않습니다.

그 폴더에 다음 전체 `main.go`를 작성합니다.

```go
package main

import "fmt"

func main() {
    fmt.Println("새 모듈에서 실행")
}
```

같은 새 폴더에서 실행합니다.

```sh
go run .
go list -m
```

예상 출력은 각각 `새 모듈에서 실행`, `example.com/learn/packages`입니다. `go.mod`의 Go 버전 줄은 명령을 실행한 도구 체인에 따라 결정되므로 이 교재의 루트 파일과 글자까지 같아야 하는 것은 아닙니다.

### 11.2 외부 모듈과 주요 버전

실제 외부 패키지를 추가할 때는 패키지 경로·릴리스 버전·유지 관리 상태를 확인합니다. v2 이상의 주요 버전은 일반적으로 모듈 경로에 `/v2` 같은 접미사를 포함하는 규칙이 적용됩니다. 예외와 전환 규칙은 [Major version suffixes](https://go.dev/ref/mod#major-version-suffixes)에 있습니다.

학습용 명령에서 임의로 만든 경로를 실제 `go get` 대상으로 실행하지 않습니다. 현재 교재의 SQL 장처럼 외부 패키지가 필요한 곳에서 목적과 실제 버전을 함께 안내합니다.

`go install 패키지경로@버전`은 실행 도구를 설치하는 명령입니다. 라이브러리 의존성을 코드에서 import하는 작업과 구별합니다. 도구 설치와 모듈 의존 변경을 같은 작업이라고 생각하면 go.mod 변화 여부를 오해할 수 있습니다.

### 11.3 작업 공간은 필요할 때만

`go.work`는 여러 모듈을 함께 개발할 때 사용하는 작업 공간 설정입니다. 여러 **패키지**를 만들었다는 이유만으로 필요한 파일이 아닙니다. 이 교재의 기본 예제는 단일 모듈이므로 새 `go.work`가 필요하지 않습니다.

명령이 예상과 다르게 동작하면 `go env GOWORK`로 상위 디렉터리의 작업 공간 영향을 확인합니다. 전역 설정을 무조건 바꾸기보다 현재 명령이 어떤 파일을 읽는지부터 봅니다.

## 12. 자주 하는 실수

| 증상 | 원인 후보 | 확인·수정 |
| --- | --- | --- |
| 같은 파일에서는 되는데 다른 파일에서는 fmt를 모름 | import의 파일 스코프 | 사용하는 파일에 import 추가 |
| 함수가 존재하는데 undefined 오류 | main.go 한 파일만 실행 | 디렉터리 단위 go run |
| found packages 오류 | 같은 일반 소스 폴더에 다른 패키지 선언 | 디렉터리 분리 또는 선언 일치 |
| exported처럼 보여도 필드 접근 불가 | 타입과 필드 공개 여부 혼동 | 필드 이름·메서드 계약 확인 |
| import cycle not allowed | 직접·간접 순환 | 책임과 의존 방향 다시 설계 |
| use of internal package not allowed | internal 부모 트리 밖에서 사용 | 허용 경계 또는 공개 API 확인 |
| 외부 import에서 package 이름 충돌 | 서로 다른 경로의 같은 이름 | 명확한 import 별칭 |
| module을 찾지 못함 | 잘못된 위치·경로·버전·네트워크 | GOMOD, GOWORK, 오류 전문 확인 |
| 새 go.mod 뒤 전체 테스트에서 예제가 사라짐 | 중첩 모듈로 경계 변경 | 기본 과정의 단일 모듈 유지 |

### 12.1 순환 예제를 정상 소스에 남기지 않기

의도적으로 순환 의존성을 만들어 보면 오류 문구를 이해하기 좋습니다. 하지만 항상 실패하는 예제를 `examples/`에 두면 모든 장의 검사까지 실패합니다. 본문의 설명이나 별도 개인 실습에서만 재현하고, 정상 코드 경로는 계속 빌드할 수 있게 유지합니다.

### 12.2 패키지 분리 후 검증할 항목

파일 이동은 내용이 같아도 import 경로와 공개 접근을 바꿀 수 있습니다. 이동 뒤에는 파일 하나의 실행만 확인하지 말고 관련 패키지들과 호출자를 함께 검사합니다.

```sh
go test ./examples/10-packages-modules/...
go vet ./examples/10-packages-modules/...
go build ./examples/10-packages-modules/...
```

본문 예제의 main 출력은 예제 테스트로 고정되어 있습니다. `greet`, `counter`, `order`, `report`의 주요 계약도 테스트합니다. 테스트 문법을 지금 모두 외울 필요는 없으며 다음 12장에서 직접 작성합니다.

## 13. 연습문제

[연습문제](../../exercises/10-packages-modules/README.md)에서 직접 패키지를 나눠 봅니다. [해설](../../solutions/10-packages-modules/README.md)은 시도한 뒤 비교합니다.

| 번호 | 수준 | 할 일 |
| --- | --- | --- |
| 1 | 확인 | 파일·패키지·모듈과 공개 경로 판단 |
| 2 | 변형 | 평균 함수를 같은 패키지의 별도 파일로 분리 |
| 3 | 변형 | import·공개 이름·internal 오류 진단 |
| 4 | 응용 | 라벨 처리 패키지와 출력 프로그램 분리 |
| 5 | 응용 | 비공개 잔액과 소비 메서드를 가진 지갑 패키지 |
| 6 | 응용 | 소비 측 인터페이스로 작업과 표시 연결 |
| 선택 A | 확인 | 가상 의존 그래프에서 버전과 체크섬 역할 설명 |

## 14. 요약과 자기 점검

- 파일은 편집 단위, 패키지는 함께 컴파일하고 이름을 공개하는 단위, 모듈은 함께 버전과 의존성을 관리하는 단위입니다.
- 같은 패키지는 여러 파일로 나눌 수 있고 import는 각 파일에서 선언합니다.
- 공개 타입 안에서도 비공개 필드를 사용해 변경 경로를 모을 수 있습니다.
- internal과 소문자 이름은 각각 다른 층의 접근 범위를 제한합니다.
- 의존성은 한 방향으로 구성하며, 인터페이스는 실제 필요한 동작을 소비하는 위치에서 작게 정의할 수 있습니다.
- go.mod, go.sum, 테스트는 서로 다른 사실을 기록하고 검증합니다.

책을 보지 않고 답해 보세요.

1. 같은 모듈의 패키지를 import할 때 GitHub에 먼저 올릴 필요가 없는 이유는 무엇인가요?
2. 같은 패키지의 함수는 공유되는데 import는 파일마다 필요한 이유는 무엇인가요?
3. 공개 타입의 소문자 필드는 같은 패키지의 다른 파일에서도 숨겨지나요?
4. internal 아래 공개 함수와 일반 패키지의 비공개 함수는 어떤 경계가 다른가요?
5. report가 order를 import하지 않아도 Basket을 사용할 수 있는 이유는 무엇인가요?
6. go.sum이 있다고 해서 모든 의존성의 현재 선택 버전을 그 파일만으로 알 수 있나요?
7. go mod tidy와 go mod verify가 각각 무엇을 바꾸거나 확인하나요?
8. 여러 패키지를 가진 단일 모듈에 go.work가 반드시 필요한가요?

## 참고 자료

공식 문서와 도구 동작 확인일: **2026-11-09**.

- [How to Write Go Code](https://go.dev/doc/code)
- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Go 명세: Packages](https://go.dev/ref/spec#Packages), [Import declarations](https://go.dev/ref/spec#Import_declarations)
- [Go 명세: Exported identifiers](https://go.dev/ref/spec#Exported_identifiers), [Program initialization](https://go.dev/ref/spec#Program_initialization_and_execution)
- [Go Modules Reference](https://go.dev/ref/mod), [MVS](https://go.dev/ref/mod#minimal-version-selection), [go.sum](https://go.dev/ref/mod#go-sum-files)
- [Go Toolchains](https://go.dev/doc/toolchain)
- [go 명령](https://pkg.go.dev/cmd/go), [internal 디렉터리](https://pkg.go.dev/cmd/go#hdr-Internal_packages)

## 다음 학습

다음 장에서는 입력 오류·파일 오류·실패 원인을 `error`로 표현하고, `defer`로 자원을 정리합니다. 패키지의 사용 계약에 실패 이유까지 포함하는 방법을 배웁니다.

[이전: 09장 메서드와 인터페이스](../09-methods-interfaces/README.md) · [다음: 11장 오류 처리와 자원 정리](../11-errors-defer/README.md) · [전체 목차](../README.md)
