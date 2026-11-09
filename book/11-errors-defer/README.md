# 11. 오류 처리와 자원 정리

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 | [10. 패키지와 모듈](../10-packages-modules/README.md) | Go, 편집기, 터미널 |

계산 결과가 없다는 사실만 알면 되는 함수에는 bool이 충분할 수 있습니다. 하지만 파일을 읽지 못했을 때는 파일이 없는지, 접근할 수 없는지, 읽는 중 문제가 생겼는지에 따라 다음 행동이 달라집니다. 이번 장에서는 **실패를 값으로 반환하고, 필요한 문맥을 더하며, 성공과 실패 어느 쪽에서도 자원을 정리하는 방법**을 배웁니다.

오류를 처리한다는 것은 화면에 오류 문자열을 출력하는 것만을 뜻하지 않습니다. 호출자가 실패 원인을 판단할 수 있게 전달하는 것, 실패했을 때 결과와 상태가 어떤지 약속하는 것, 이미 얻은 자원을 닫는 것까지 포함합니다.

## 학습 목표

- `(값, error)` 반환을 읽고 성공·실패 시 결과의 계약을 설명한다.
- 오류 문맥을 래핑하고 `errors.Is`와 `errors.As`로 원인과 세부 정보를 확인한다.
- `defer`의 평가 시점·실행 순서·함수 수명을 코드로 추적한다.
- 본 처리 오류와 정리 오류를 함께 보존하며 자원을 정확히 한 번 닫는다.
- 일반 오류 반환과 panic/recover의 용도를 구별하고 복구 경계를 좁게 설계한다.

## 사전 지식과 준비

다중 반환, 이른 반환, 함수 값·클로저, 포인터, 구조체와 인터페이스를 사용합니다. `error`도 인터페이스라는 점은 09장과 연결됩니다. 패키지 사용 경로는 10장에서 배웠습니다.

모든 명령은 별도 표시가 없으면 **저장소 루트**에서 실행합니다. 기본 예제에는 외부 파일이나 서버가 필요하지 않습니다. 자원 정리 예제는 메모리 안의 읽기 객체로 성공·실패를 결정적으로 재현합니다. 실제 파일과 스트림은 14장에서 확장합니다.

| 예제 | 중심 주제 | 실행 명령 |
| --- | --- | --- |
| [return-error](../../examples/11-errors-defer/return-error/main.go) | 오류를 반환하고 확인 | `go run ./examples/11-errors-defer/return-error` |
| [wrapping](../../examples/11-errors-defer/wrapping/main.go) | 문맥, %w, errors.Is | `go run ./examples/11-errors-defer/wrapping` |
| [typed-error](../../examples/11-errors-defer/typed-error/main.go) | 오류 타입과 errors.As | `go run ./examples/11-errors-defer/typed-error` |
| [defer-order](../../examples/11-errors-defer/defer-order/main.go) | 등록 시점과 역순 실행 | `go run ./examples/11-errors-defer/defer-order` |
| [resource](../../examples/11-errors-defer/resource/main.go) | 읽기와 닫기의 오류 보존 | `go run ./examples/11-errors-defer/resource` |
| [defer-scope](../../examples/11-errors-defer/defer-scope/main.go) | 항목별 함수 수명과 정리 | `go run ./examples/11-errors-defer/defer-scope` |
| [join](../../examples/11-errors-defer/join/main.go) | 여러 오류의 집합 | `go run ./examples/11-errors-defer/join` |
| [typed-nil](../../examples/11-errors-defer/typed-nil/main.go) | error 안의 nil 포인터 | `go run ./examples/11-errors-defer/typed-nil` |
| [panic-boundary](../../examples/11-errors-defer/panic-boundary/main.go) | 제한된 복구와 전파 | `go run ./examples/11-errors-defer/panic-boundary` |

## 1. 실패를 값으로 반환하기

### 1.1 error는 메시지를 제공하는 인터페이스

Go의 내장 `error` 인터페이스는 다음 형태의 메서드를 요구합니다.

```go
type error interface {
    Error() string
}
```

이 코드는 이미 정의된 인터페이스의 모양을 설명한 것입니다. 자신의 프로그램에서 다시 `type error`를 선언하지 않습니다. 필요한 메서드를 가진 타입은 error를 암시적으로 구현합니다. [Errors 명세](https://go.dev/ref/spec#Errors)와 [errors 패키지](https://pkg.go.dev/errors)를 참고합니다.

함수가 `error`를 반환할 때 성공은 보통 nil로 표현합니다. 실패하면 원인이나 상황을 표현하는 값을 반환합니다. error라는 변수 이름이나 err라는 관례 자체가 예외 처리 기능을 실행시키지는 않습니다. 반환된 값을 호출자가 확인해야 합니다.

### 1.2 결과와 오류를 함께 받기

저장 경로: [`examples/11-errors-defer/return-error/main.go`](../../examples/11-errors-defer/return-error/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"errors"
	"fmt"
)

func divide(total, count int) (int, error) {
	if count <= 0 {
		return 0, errors.New("count must be positive")
	}
	return total / count, nil
}
func main() {
	value, err := divide(12, 3)
	if err != nil {
		fmt.Println("실패:", err)
		return
	}
	fmt.Println("몫:", value)
	_, err = divide(12, 0)
	if err != nil {
		fmt.Println("실패:", err)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/return-error
```

예상 출력:

```text
몫: 4
실패: count must be positive
```


성공한 호출은 몫 4와 nil을 반환합니다. count가 0 이하이면 나눗셈을 하기 전에 실패합니다. 따라서 이 함수의 계약에서는 오류가 있을 때 첫 번째 결과는 사용하지 않습니다.

| 호출 | 첫 번째 결과 | error | 호출자의 다음 행동 |
| --- | ---: | --- | --- |
| divide(12, 3) | 4 | nil | 결과 사용 |
| divide(12, 0) | 0 | 실패 값 | 실패 처리 |
| divide(0, 1) | 0 | nil | 정상 0 사용 |

결과가 0이라고 실패인 것은 아닙니다. `err != nil`이 실패를 판단하는 기준입니다. 이는 맵 조회의 값과 존재 여부를 분리했던 원리와 비슷하지만, error는 실패의 세부 정보도 전달할 수 있습니다.

`divide`는 정수 나눗셈입니다. 분수가 필요하면 다른 결과 타입과 계약을 정해야 합니다. 이 장의 중심은 숫자 계산의 일반화가 아니라 실패 전달입니다.

### 1.3 오류 검사 뒤 정상 경로 진행

실패할 수 있는 호출 바로 뒤에서 오류를 확인하면 나머지 코드를 정상 흐름으로 읽기 쉽습니다. 오류를 무시하고 결과를 계속 사용하면 정상 0과 실패 0을 구별하지 못합니다.

다음은 앞 예제의 호출부를 발췌한 것입니다.

```go
value, err := divide(12, 3)
if err != nil {
    fmt.Println("실패:", err)
    return
}
fmt.Println("몫:", value)
```

main에서 return하면 main의 실행을 마칩니다. 라이브러리 함수라면 보통 오류를 다시 반환합니다. 누가 출력할지, 누가 재시도할지, 누가 종료할지는 프로그램의 경계에서 결정합니다.

### 1.4 부분 결과가 있는 API도 있습니다

모든 Go 함수가 오류 시 첫 결과를 버리라는 규칙을 갖는 것은 아닙니다. 예를 들어 스트림 읽기는 일부 바이트와 오류를 함께 반환할 수 있습니다. 반환값의 의미는 해당 함수의 문서가 결정합니다.

이 장의 직접 작성한 함수들은 별도 안내가 없으면 실패 시 유효한 결과를 반환하지 않는 계약을 사용합니다. “Go에서는 error가 있으면 모든 다른 반환값은 무효”라는 언어 규칙으로 확대하지 마세요. [io.Reader](https://pkg.go.dev/io#Reader)는 이 차이를 잘 보여 주며 14장에서 직접 다룹니다.

## 2. 오류 메시지와 오류의 식별

### 2.1 같은 글자라도 같은 오류 값은 아닙니다

`errors.New("name is empty")`는 메시지를 가진 새 오류 값을 만듭니다. 동일한 메시지로 두 번 호출해도 별개 오류 값입니다. 문자열의 모양과 프로그램이 비교하는 원인은 다릅니다.

오류 종류를 공유해야 한다면 패키지 수준에 대표 오류(sentinel error)를 정의할 수 있습니다. 아래 `ErrEmpty`가 그 예입니다. 호출자에게 이 값을 공개하는 것은 그 오류 종류를 API의 일부로 약속하는 선택입니다.

### 2.2 문맥을 더하면서 원인 보존

저장 경로: [`examples/11-errors-defer/wrapping/main.go`](../../examples/11-errors-defer/wrapping/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrEmpty = errors.New("name is empty")

func normalizeName(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ErrEmpty
	}
	return text, nil
}
func loadName(text string) (string, error) {
	value, err := normalizeName(text)
	if err != nil {
		return "", fmt.Errorf("load name: %w", err)
	}
	return value, nil
}
func main() {
	_, err := loadName(" ")
	fmt.Println("오류:", err)
	fmt.Println("직접 비교:", err == ErrEmpty)
	fmt.Println("원인 비교:", errors.Is(err, ErrEmpty))
	fmt.Println("새 오류 비교:", errors.Is(err, errors.New("name is empty")))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/wrapping
```

예상 출력:

```text
오류: load name: name is empty
직접 비교: false
원인 비교: true
새 오류 비교: false
```


`normalizeName`은 이름 자체의 유효성을 확인합니다. `loadName`은 그 실패가 이름을 불러오는 과정에서 발생했다고 문맥을 더합니다. `%w`가 원래 오류를 감싼 관계를 보존합니다.

**래핑(wrapping)**은 오류에 설명을 붙이면서 내부 원인을 추적할 수 있게 하는 방식입니다. `fmt.Errorf("load name: %w", err)`를 사용하면 사람이 읽는 문장과 프로그램이 확인하는 원인 관계를 함께 유지합니다. [fmt.Errorf](https://pkg.go.dev/fmt#Errorf), [Go 오류 래핑 안내](https://go.dev/blog/go1.13-errors)를 참고합니다.

| 방식 | 메시지에 원인 포함 | errors.Is로 원인 확인 |
| --- | --- | --- |
| 원래 err 그대로 반환 | 원래 메시지 | 가능 |
| fmt.Errorf("문맥: %w", err) | 가능 | 래핑 관계로 가능 |
| fmt.Errorf("문맥: %v", err) | 가능 | 그 포맷만으로는 보존되지 않음 |
| errors.New(err.Error()) | 글자 복사 | 기존 정체성·래핑 관계 소실 |

### 2.3 직접 비교와 errors.Is

`err == ErrEmpty`는 지금 들고 있는 오류 값이 대표 오류 그 자체인지 확인합니다. 래핑된 오류는 바깥의 다른 값이므로 false입니다.

`errors.Is(err, ErrEmpty)`는 오류 자신과 감싼 원인을 따라가며 원하는 오류와 일치하는지 확인합니다. 사용자 정의 Is 메서드가 있다면 그 규칙도 적용될 수 있습니다. 이 장에서는 단순 대표 오류 비교를 사용합니다. [errors.Is](https://pkg.go.dev/errors#Is)를 참고합니다.

`errors.Is`는 오류 문장을 검색하는 함수가 아닙니다. 같은 메시지로 새 오류를 만들어 비교하면 기존 ErrEmpty와 일치하지 않습니다.

### 2.4 어떤 문맥을 붙일까요?

유용한 문맥은 실패한 작업이나 안전하게 표시할 수 있는 식별 정보입니다. 예를 들어 `load profile`, `decode row 3`처럼 어느 단계인지 알려 주면 호출 경로를 이해하기 쉽습니다.

| 문맥 | 독자가 얻는 정보 |
| --- | --- |
| `failed` | 구체적으로 무엇이 실패했는지 부족 |
| `read config` | 실패 단계 |
| `parse age` | 실패한 필드 또는 작업 |
| 이미 같은 문구를 여러 번 반복 | 원인 파악보다 길이만 늘 수 있음 |

비밀번호나 접근 토큰 같은 실제 비밀값을 오류에 넣지 않습니다. 이 교재의 예제는 가상의 입력만 사용합니다. 실제 서비스에서 외부 응답으로 그대로 보여 줄 메시지는 내부 오류와 별도 정책을 둘 수 있으며 23장에서 다룹니다.

### 2.5 래핑은 공개 계약의 선택

내부 구현 오류를 `%w`로 반환하면 호출자가 그 원인을 `errors.Is/As`로 검사할 수 있습니다. 나중에 구현을 바꾸면서 같은 원인이 사라지면 호출자의 분기가 달라질 수 있습니다.

따라서 무조건 모든 내부 오류를 노출하는 것이 정답은 아닙니다. 호출자에게 필요한 원인인가, 패키지의 추상화 경계 안에 둘 세부사항인가를 판단합니다. 이 장의 예제는 원인 추적을 학습하려고 의도적으로 보존합니다.

## 3. 구조를 가진 오류와 errors.As

### 3.1 오류 메시지에서 값을 잘라내지 않기

필드 이름과 잘못된 값이 필요하다고 오류 문자열을 콜론 기준으로 분리하면 메시지가 바뀔 때 프로그램이 깨집니다. 데이터가 필요하면 구조체 필드를 가진 오류 타입으로 표현할 수 있습니다.

저장 경로: [`examples/11-errors-defer/typed-error/main.go`](../../examples/11-errors-defer/typed-error/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrOutOfRange = errors.New("value out of range")

type FieldError struct {
	Field string
	Value string
	Cause error
}

func (e *FieldError) Error() string { return fmt.Sprintf("%s=%q: %v", e.Field, e.Value, e.Cause) }
func (e *FieldError) Unwrap() error { return e.Cause }
func parseAge(text string) (int, error) {
	age, err := strconv.Atoi(text)
	if err != nil {
		return 0, &FieldError{Field: "age", Value: text, Cause: err}
	}
	if age < 0 || age > 150 {
		return 0, &FieldError{Field: "age", Value: text, Cause: ErrOutOfRange}
	}
	return age, nil
}
func main() {
	_, err := parseAge("151")
	err = fmt.Errorf("profile: %w", err)
	var fieldErr *FieldError
	if errors.As(err, &fieldErr) {
		fmt.Println("필드:", fieldErr.Field, fieldErr.Value)
	}
	fmt.Println("범위 오류:", errors.Is(err, ErrOutOfRange))
	_, err = parseAge("many")
	var numberErr *strconv.NumError
	fmt.Println("숫자 오류:", errors.As(err, &numberErr))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/typed-error
```

예상 출력:

```text
필드: age 151
범위 오류: true
숫자 오류: true
```


`FieldError`는 필드 이름, 원본 문자열, 내부 원인을 보관합니다. Error 메서드는 사람이 읽을 설명을 만들고 Unwrap 메서드는 내부 원인을 돌려줍니다. Error 메서드가 포인터 수신자이므로 `*FieldError`가 error를 구현합니다.

`parseAge`는 0부터 150까지를 허용합니다. 숫자로 바꿀 수 없으면 `strconv.Atoi`의 오류를 원인으로 보관하고, 숫자이지만 범위를 벗어나면 `ErrOutOfRange`를 보관합니다.

이 범위는 학습용 프로필 검증 정책입니다. Go의 int 타입 범위와 다릅니다. 타입 변환에 성공해도 도메인 규칙 검사에 실패할 수 있습니다.

### 3.2 Is는 종류, As는 타입별 정보

| 질문 | 도구 |
| --- | --- |
| 범위를 벗어난 오류인가요? | errors.Is(err, ErrOutOfRange) |
| FieldError를 찾아 필드명을 읽을 수 있나요? | errors.As(err, &fieldErr) |
| 숫자 변환 오류의 NumError가 있나요? | errors.As(err, &numberErr) |

두 함수는 같은 오류에 함께 사용할 수 있습니다. `profile`이라는 바깥 문맥이 있어도 내부의 FieldError와 ErrOutOfRange를 확인할 수 있습니다.

### 3.3 왜 포인터의 주소를 넘기나요?

다음 발췌를 한 줄씩 읽습니다.

```go
var fieldErr *FieldError
if errors.As(err, &fieldErr) {
    fmt.Println(fieldErr.Field, fieldErr.Value)
}
```

`fieldErr`는 찾은 `*FieldError` 값을 저장할 변수입니다. errors.As가 그 변수에 값을 대입해야 하므로 **변수의 주소인 `&fieldErr`**를 전달합니다. 변수 자체가 포인터 타입이라는 사실과, 그 변수에 결과를 써야 한다는 사실은 별개입니다.

처음부터 `fieldErr := &FieldError{}`를 만들고 그 필드를 억지로 채우는 함수가 아닙니다. 오류 트리에서 타입이 맞는 실제 오류 값을 찾아 대상 변수에 저장합니다. [errors.As](https://pkg.go.dev/errors#As)를 참고합니다.

As의 대상 인자는 필요한 타입의 값을 받을 수 있는 nil이 아닌 포인터여야 합니다. 잘못된 대상 형태는 패닉을 만들 수 있습니다. 성공 여부를 확인하기 전에 찾은 오류 필드를 읽지 않습니다.

### 3.4 타입 단언과의 차이

```go
fieldErr, ok := err.(*FieldError)
```

이 단언은 현재 바깥 오류 값의 동적 타입만 확인합니다. `fmt.Errorf("profile: %w", err)`로 감싸면 바깥 타입이 달라져 실패할 수 있습니다. errors.As는 래핑 내부까지 탐색합니다.

09장에서 배운 단언이 잘못된 문법은 아닙니다. 지금 필요한 질문이 바깥 타입 하나인지, 전체 원인 관계 안의 타입인지에 따라 도구를 선택합니다.

## 4. defer의 실행 시점과 순서

### 4.1 함수가 끝날 때 할 일 등록

**defer**는 현재 함수가 반환할 때 실행할 함수 호출을 등록합니다. 정상 경로·이른 반환·panic으로 빠져나가는 경로에서 정리 동작을 모으는 데 사용할 수 있습니다. [Defer statements](https://go.dev/ref/spec#Defer_statements)를 참고합니다.

저장 경로: [`examples/11-errors-defer/defer-order/main.go`](../../examples/11-errors-defer/defer-order/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

func show() {
	value := 1
	defer fmt.Println("인자 저장:", value)
	defer func() { fmt.Println("나중에 읽기:", value) }()
	defer fmt.Println("마지막 등록")
	value = 2
	fmt.Println("본문:", value)
}
func main() { show(); fmt.Println("호출자") }
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/defer-order
```

예상 출력:

```text
본문: 2
마지막 등록
나중에 읽기: 2
인자 저장: 1
호출자
```


실행 흐름을 표로 추적해 봅니다.

| 시점 | value | 등록하거나 실행하는 일 |
| --- | ---: | --- |
| 변수 선언 | 1 | 아직 defer 없음 |
| 첫 defer | 1 | Println의 인자 값 1을 저장 |
| 둘째 defer | 1 | 나중에 value를 읽을 클로저 등록 |
| 셋째 defer | 1 | 마지막 등록 문자열 출력 호출 등록 |
| 대입과 본문 출력 | 2 | 본문: 2 |
| show 반환 시작 | 2 | 셋째, 둘째, 첫째 순으로 실행 |
| main 복귀 | 해당 없음 | 호출자 출력 |

### 4.2 인자는 등록할 때 평가

`defer fmt.Println("인자 저장:", value)`는 Println을 나중에 호출하지만, 인자 value는 defer 문장을 실행할 때 평가합니다. 당시 값 1이 저장됩니다.

반면 클로저 안의 `value` 읽기는 그 함수 본문이 실제 실행될 때 일어납니다. 같은 지역 변수의 변경된 값 2를 읽습니다. 함수 값과 클로저의 차이를 함께 적용한 사례입니다.

아래 두 발췌는 같지 않습니다.

```go
defer fmt.Println(value)
```

```go
defer func() {
    fmt.Println(value)
}()
```

클로저라는 이유만으로 무조건 옛 값이나 새 값을 읽는 것은 아닙니다. 값을 인자로 넘기는지, 본문에서 바깥 변수를 읽는지 살펴보세요.

### 4.3 마지막에 등록한 것부터 실행

여러 defer는 마지막에 등록한 호출부터 역순으로 실행합니다. 자원 A를 얻은 다음 B를 얻었다면 B를 먼저 닫고 A를 닫는 순서를 만들기 좋습니다.

하지만 등록 순서가 자원 의존성과 맞아야 합니다. 역순이라는 사실만으로 모든 정리가 자동으로 올바른 것은 아닙니다. 어떤 자원이 다른 자원을 필요로 하는지 이해해야 합니다.

### 4.4 defer는 블록이 아닌 함수 수명

for의 한 반복이 끝났다고 그 안에서 등록한 defer가 바로 실행되지는 않습니다. 주변 함수가 끝날 때까지 남습니다.

다음은 설명용 발췌입니다.

```go
for _, name := range names {
    resource := openForExample(name)
    defer resource.Close()
}
```

`openForExample`은 실제 저장소의 함수가 아니라 수명 문제를 설명하는 가상 이름입니다. 이 형태는 반복마다 자원을 연 뒤 함수 끝까지 정리를 미룹니다. 많은 자원을 처리한다면 한 항목을 처리하고 닫는 별도 함수를 만들거나 명시적으로 Close하는 구조를 고려합니다.

실제 파일 예제에서 open 실패 확인 없이 defer를 등록해서는 안 됩니다. 자원을 얻는 데 성공한 뒤 정리 책임을 등록합니다.

## 5. 읽기와 닫기 오류를 함께 처리하기

### 5.1 자원을 누가 닫나요?

함수가 받은 객체를 닫는지, 호출자가 계속 쓰는지 계약이 필요합니다. 두 곳이 모두 닫거나 어느 곳도 닫지 않으면 수명 관리가 모호해집니다.

아래 `readAndClose`는 전달받은 입력의 소유권을 넘겨받고 정확히 한 번 닫습니다. 입력은 nil이 아닌 유효한 io.ReadCloser이며 이 최소 예제의 작은 메모리 입력을 전제로 합니다. 일반 외부 데이터의 크기 제한은 14장에서 다룹니다.

`io.ReadCloser`는 Read와 Close 동작을 가진 인터페이스입니다. 실제 파일도 이런 행동을 제공할 수 있고, 이 예제의 가짜 자원도 제공합니다. [io.ReadCloser](https://pkg.go.dev/io#ReadCloser)를 참고합니다.

저장 경로: [`examples/11-errors-defer/resource/main.go`](../../examples/11-errors-defer/resource/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type trackedReader struct {
	io.Reader
	closeErr error
	closed   bool
}

func (r *trackedReader) Close() error { r.closed = true; return r.closeErr }

// readAndClose owns input and closes it once. Any error discards the result.
func readAndClose(input io.ReadCloser) (text string, err error) {
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close input: %w", closeErr))
			text = ""
		}
	}()
	data, readErr := io.ReadAll(input)
	if readErr != nil {
		return "", fmt.Errorf("read input: %w", readErr)
	}
	return string(data), nil
}
func main() {
	input := &trackedReader{Reader: strings.NewReader("Go")}
	value, err := readAndClose(input)
	fmt.Println("결과:", value, "성공:", err == nil, "닫힘:", input.closed)
	closeFailure := errors.New("close failed")
	input = &trackedReader{Reader: strings.NewReader("Go"), closeErr: closeFailure}
	value, err = readAndClose(input)
	fmt.Println("실패 결과:", value == "", "닫기 오류:", errors.Is(err, closeFailure))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/resource
```

예상 출력:

```text
결과: Go 성공: true 닫힘: true
실패 결과: true 닫기 오류: true
```


### 5.2 성공·읽기 실패·닫기 실패

`readAndClose`의 계약은 **어떤 오류가 있어도 빈 문자열을 반환**하는 것입니다. 읽기를 성공했어도 닫기에서 오류가 발생하면 실패로 취급합니다. 다른 API가 부분 결과를 허용할 수도 있지만 여기서는 결과 사용 여부를 단순하게 유지합니다.

| 읽기 | 닫기 | 반환 문자열 | 반환 오류 |
| --- | --- | --- | --- |
| 성공 | 성공 | 읽은 내용 | nil |
| 실패 | 성공 | 빈 문자열 | read 문맥 + 읽기 원인 |
| 성공 | 실패 | 빈 문자열 | close 문맥 + 닫기 원인 |
| 실패 | 실패 | 빈 문자열 | 두 원인 모두 보존 |

정리 오류를 처리 오류로 덮어쓰면 처음 실패 원인이 사라질 수 있습니다. 반대로 닫기 오류를 항상 무시하면 완료되지 않은 처리를 성공으로 알릴 수 있습니다. 어떤 자원에서 닫기 오류가 의미 있는지 API의 계약에 맞춰 결정합니다.

### 5.3 이름 붙인 반환값과 defer

함수 선언을 다시 봅니다.

```go
func readAndClose(input io.ReadCloser) (text string, err error)
```

text와 err는 이름 붙인 반환 변수입니다. return 문장이 반환값을 이 변수들에 설정한 뒤 defer가 실행됩니다. 따라서 defer에서 err와 text를 수정하면 호출자가 받는 최종 결과도 바뀝니다.

읽기를 성공한 `return string(data), nil`도 곧바로 호출자에게 결과가 확정되어 전달되는 것은 아닙니다. defer가 닫기를 수행하고, 실패하면 err를 추가하고 text를 비웁니다.

| 순서 | 동작 |
| --- | --- |
| 1 | return 식을 계산 |
| 2 | 이름 붙인 반환 변수에 값 대입 |
| 3 | 등록한 defer들을 역순 실행 |
| 4 | 최종 반환값을 호출자에게 전달 |

이 기능을 사용하면 정리 오류를 반영할 수 있지만, 긴 함수에서 반환값이 숨은 위치에서 바뀌면 읽기 어렵습니다. 여기서는 정리 책임을 가진 작은 함수 안에서만 사용합니다.

### 5.4 짧은 변수 선언으로 오류 가리기

defer 안에서 최종 반환 오류를 바꾸려는데 `err := ...`로 새 지역 변수를 만들면 바깥 반환 변수는 그대로일 수 있습니다. 04장에서 배운 이름 가리기가 오류 처리에서도 중요합니다.

참고 구현은 `closeErr`라는 별도 이름을 사용하고, 결과를 이름 붙인 `err`에 `=`로 대입합니다. 어떤 변수를 바꾸는지 눈으로 확인할 수 있게 하기 위한 선택입니다.


### 5.5 반복마다 닫고 싶다면 함수 수명을 나누기

여러 자원을 차례로 처리하는 상황에서 모든 정리를 바깥 main의 끝까지 미룰 필요는 없습니다. 한 자원의 처리·정리를 작은 함수에 묶으면 그 함수가 반환할 때마다 defer가 실행됩니다.

다음 프로그램은 실제 파일을 열지 않습니다. 이름을 가진 작은 객체가 열기·처리·닫기 시점을 출력하도록 만들어 수명을 관찰합니다. 프로그램의 중심은 객체 종류가 아니라 **어느 함수가 정리 책임을 갖는가**입니다.

저장 경로: [`examples/11-errors-defer/defer-scope/main.go`](../../examples/11-errors-defer/defer-scope/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"errors"
	"fmt"
)

var ErrName = errors.New("resource name required")

type resource struct{ name string }

func openResource(name string) (*resource, error) {
	if name == "" {
		return nil, ErrName
	}
	fmt.Println("열기:", name)
	return &resource{name: name}, nil
}
func (r *resource) Close() error { fmt.Println("닫기:", r.name); return nil }
func processOne(name string) (err error) {
	input, err := openResource(name)
	if err != nil {
		return fmt.Errorf("open resource: %w", err)
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close resource: %w", closeErr))
		}
	}()
	fmt.Println("처리:", name)
	return nil
}
func main() {
	for _, name := range []string{"A", "B"} {
		if err := processOne(name); err != nil {
			fmt.Println("실패:", err)
			return
		}
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/defer-scope
```

예상 출력:

```text
열기: A
처리: A
닫기: A
열기: B
처리: B
닫기: B
```

A의 닫기가 B의 열기보다 먼저 나옵니다. processOne("A")의 반환과 defer가 끝난 뒤 main의 다음 반복이 시작되기 때문입니다. defer가 반복 블록을 감지해서 실행되는 것이 아니라, 반복마다 별도로 호출한 함수가 끝나는 것입니다.

### 5.6 획득·사용·정리를 한 함수에서 추적

| 단계 | processOne의 동작 | 실패할 때 |
| --- | --- | --- |
| 획득 | openResource(name) | 원인을 감싸 즉시 반환 |
| 정리 등록 | defer에 Close 등록 | 획득에 성공한 경우만 등록 |
| 사용 | 처리 동작 | 실제 예제에서는 오류 반환 가능 |
| 반환 준비 | 최종 작업 결과 설정 | 아직 defer는 남아 있음 |
| 정리 | Close 호출 | 이름 붙인 err에 정리 오류 추가 |

openResource가 빈 이름을 받으면 자원을 만들지 않고 ErrName을 반환합니다. 이 경우에는 닫을 자원이 없으므로 Close도 호출하지 않습니다. 테스트는 이 실패 원인이 보존되는지 확인합니다.

이 예제의 Close는 항상 성공합니다. 본문 앞의 readAndClose 예제는 Close 실패까지 재현합니다. 두 예제는 서로 다른 질문을 분리합니다. 하나는 함수 수명과 등록 위치, 다른 하나는 두 오류의 보존입니다.

### 5.7 긴 함수의 여러 return과 정리 중복

정리를 각 return 앞에 직접 적으면 초기에는 이해하기 쉬울 수 있습니다. 하지만 새로운 실패 분기를 추가할 때 한 경로에서 Close를 빠뜨리거나, 앞에서 닫은 자원을 다시 닫을 위험이 있습니다.

정리 책임을 defer 한 곳에 모으면 사용 경로가 여러 개여도 등록한 정리가 함수 종료 때 실행됩니다. 다만 획득하기도 전에 등록하거나, 함수 수명이 지나치게 길다면 별도의 문제가 생깁니다. “defer를 사용했다”보다 “정리 책임과 수명이 맞는다”가 중요한 확인 기준입니다.

### 5.8 이른 반환 뒤에 실행되는 것

개인 실습에서 processOne의 처리 부분을 오류 반환으로 바꿔 볼 수 있습니다. 반환을 시작해도 이미 등록된 Close는 실행됩니다. main이 다음 자원을 처리할지는 main의 오류 정책에 달려 있습니다.

현재 main은 실패하면 그 시점에서 반환하므로 이후 자원은 열지 않습니다. 모든 항목을 검사하면서 실패를 모으고 싶다면 main의 반복 정책과 결과 형식을 바꿔야 합니다. defer는 그 결정을 대신하지 않습니다.

| 정책 | 항목 하나의 오류 뒤 | 항목 하나의 자원 정리 |
| --- | --- | --- |
| 즉시 중단 | 다음 항목을 시작하지 않음 | processOne이 담당 |
| 전체 처리 후 오류 모음 | 다음 항목도 진행 | processOne이 담당 |
| 재시도 | 정한 조건에서 다시 호출 | 각 호출이 자기 자원 담당 |

항목별 함수가 자원을 정확하게 정리하면 바깥에서 정책을 바꿀 때 수명 코드를 덜 흔들 수 있습니다. 실제 재시도에는 중복 처리 가능성과 취소 조건도 필요하며 31장에서 다룹니다.

### 5.9 메모리 회수와 Close는 별개입니다

Go에는 사용하지 않는 메모리를 회수하는 가비지 컬렉터가 있습니다. 하지만 그 사실이 파일·소켓·트랜잭션의 정상 종료 정책을 자동으로 대신하지는 않습니다. 자원의 외부 상태와 오류 처리 시점은 코드의 계약으로 정합니다.

이 예제의 resource는 메모리 객체이지만 Close를 명시적인 수명 이벤트로 사용합니다. 실제 파일에서 Close를 언제 하는지, 데이터베이스에서 Commit 또는 Rollback을 누가 하는지 역시 메모리 주소를 잃는 시점과 구분해야 합니다.

### 5.10 직접 확인할 질문

1. A를 열고 사용한 뒤 A를 닫기 전에 B가 열리나요?
2. 빈 이름을 전달해 획득이 실패하면 어떤 defer가 등록되나요?
3. openResource가 돌려준 오류를 `%v`로만 포맷하면 errors.Is 결과가 달라지나요?
4. processOne 안에서 os.Exit를 호출하면 Close가 실행되나요?
5. main이 모든 오류를 모으는 정책으로 바뀌어도 누가 자원을 닫아야 하나요?

답은 각각 “아니요”, “정리 defer는 등록되지 않음”, “원인 관계가 보존되지 않을 수 있음”, “아니요”, “자원을 소유한 processOne”입니다. 결과 문구만 외우지 말고 출력 순서와 함수 반환을 연결해 설명해 봅니다.

## 6. 여러 오류를 모으기

### 6.1 첫 오류만으로 충분하지 않은 입력 검증

필수 제목과 양수 개수를 동시에 검사할 때 하나씩만 알려 주면 사용자가 여러 번 수정해야 할 수 있습니다. 서로 독립된 검증은 오류들을 모아서 반환할 수 있습니다.

저장 경로: [`examples/11-errors-defer/join/main.go`](../../examples/11-errors-defer/join/main.go). 다음은 전체 코드입니다.

```go
package main

import (
	"errors"
	"fmt"
)

var ErrTitle = errors.New("title required")
var ErrCount = errors.New("count must be positive")

func validate(title string, count int) error {
	var problems []error
	if title == "" {
		problems = append(problems, ErrTitle)
	}
	if count <= 0 {
		problems = append(problems, ErrCount)
	}
	return errors.Join(problems...)
}
func main() {
	err := validate("", 0)
	fmt.Println("제목 오류:", errors.Is(err, ErrTitle))
	fmt.Println("개수 오류:", errors.Is(err, ErrCount))
	fmt.Println("정상:", validate("Go", 2) == nil)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/join
```

예상 출력:

```text
제목 오류: true
개수 오류: true
정상: true
```


`errors.Join`은 nil이 아닌 오류들을 하나의 오류로 결합합니다. 모든 입력이 nil이거나 오류 목록이 비어 있으면 nil을 반환합니다. 따라서 정상 입력에서는 반환 오류가 없습니다. [errors.Join](https://pkg.go.dev/errors#Join)을 참고합니다.

Join으로 만든 오류는 내부에 여러 자식을 가질 수 있습니다. errors.Is와 errors.As는 이 관계도 탐색합니다. 오류 구조가 항상 한 줄짜리 체인이라고만 생각하면 Join을 이해하기 어렵습니다.

### 6.2 결합된 오류의 문장과 구조

Join의 문자열 표현은 여러 오류 메시지를 줄바꿈으로 이어 붙일 수 있습니다. 출력 모양이 필요하면 API 문서를 확인하지만, 프로그램의 판단은 문자열 줄 수나 부분 문자열에 의존하지 않습니다.

`errors.Unwrap`은 `Unwrap() error` 형태를 다루는 함수이며 `Unwrap() []error`의 모든 자식을 차례로 꺼내 주는 탐색 함수가 아닙니다. 여러 오류를 포함할 수 있는 상황에서 직접 Unwrap만 반복하기보다 Is와 As를 사용합니다.

### 6.3 함께 모아도 되는 검사와 중단해야 하는 실패

독립적인 입력 필드 오류를 모으는 것과, 반드시 성공해야 다음 작업이 가능한 처리 오류는 다릅니다.

| 상황 | 보통 적절한 흐름 |
| --- | --- |
| 이름·나이 두 입력이 각각 잘못됨 | 둘 다 검사해 모을 수 있음 |
| 파일 열기 실패 | 읽기 진행 불가, 즉시 반환 |
| 읽기 실패 후 열린 파일 닫기 | 정리는 계속 수행하고 오류를 함께 보존 |
| 검증 실패 후 데이터베이스 쓰기 | 정책상 보통 실행하지 않음 |

“오류가 있어도 끝까지 진행한다”는 하나의 일반 규칙은 없습니다. 각 단계의 의존성과 상태 변경 계약을 기준으로 판단합니다.

## 7. error 안의 typed nil

### 7.1 nil 포인터를 error로 반환하는 함정

저장 경로: [`examples/11-errors-defer/typed-nil/main.go`](../../examples/11-errors-defer/typed-nil/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

type InputError struct{ Message string }

func (e *InputError) Error() string { return e.Message }
func wrong() error                  { var err *InputError; return err }
func correct() error                { return nil }
func main() {
	fmt.Println("typed nil:", wrong() == nil)
	fmt.Println("진짜 nil:", correct() == nil)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/typed-nil
```

예상 출력:

```text
typed nil: false
진짜 nil: true
```


`wrong` 안의 지역 변수 err는 nil `*InputError`입니다. 그 값을 error 인터페이스로 반환하면 동적 타입이 `*InputError`, 동적 값이 nil인 인터페이스가 됩니다. 인터페이스 전체는 nil이 아닙니다.

이 때문에 호출자가 `if err != nil`을 검사하면 실패로 들어갑니다. 성공을 표현하려면 `return nil`을 직접 반환합니다. [Go FAQ의 nil 오류 설명](https://go.dev/doc/faq#nil_error)을 참고합니다.

### 7.2 메시지를 출력하기 전에 원인을 이해하기

오류가 이상한 nil처럼 출력되거나 Error 메서드를 호출할 때 문제가 생기면, 인터페이스의 동적 타입을 확인하세요. 모든 Error 메서드가 nil 수신자를 지원한다고 가정할 수 없습니다.

이 예제의 wrong은 **잘못된 반환 패턴을 안전하게 관찰하기 위한 함수**입니다. nil 여부만 출력하며 그 Error 메서드를 직접 호출하지 않습니다. 실제 정상 경로에서는 correct처럼 nil을 반환합니다.

typed nil을 해결하려고 호출자마다 reflect를 사용해 내부 포인터를 검사하는 것보다, 오류를 만드는 함수의 성공 반환을 고치는 것이 이 문제의 직접적인 수정입니다.

## 8. panic과 recover의 범위

### 8.1 일반적인 실패와 panic

`panic`은 현재 함수의 정상 실행을 중단하고 호출 스택을 거슬러 올라가며 defer를 실행합니다. 복구되지 않은 panic은 프로그램 종료로 이어질 수 있습니다. 배열 인덱스 범위 위반 같은 런타임 문제도 panic을 발생시킬 수 있습니다. [Handling panics](https://go.dev/ref/spec#Handling_panics)를 참고합니다.

잘못된 사용자 입력, 없는 파일, 예상 가능한 조회 실패를 무조건 panic으로 표현하지 않습니다. 그런 실패는 호출자가 정상적으로 선택해야 하는 경우가 많으므로 error 반환이 명확합니다.

### 8.2 복구는 실행하던 문장으로 돌아가는 기능이 아닙니다

아래 예제는 복구 경계를 설명하기 위해 일부러 특정 타입의 panic을 발생시킵니다. 실제 업무 입력 검증을 이렇게 작성하라는 권장 패턴은 아닙니다.

저장 경로: [`examples/11-errors-defer/panic-boundary/main.go`](../../examples/11-errors-defer/panic-boundary/main.go). 다음은 전체 코드입니다.

```go
package main

import "fmt"

type stopped struct{ Reason string }

// invoke contains only the known stopped panic used by this demonstration.
func invoke(step func()) (err error) {
	defer func() {
		if value := recover(); value != nil {
			if known, ok := value.(stopped); ok {
				err = fmt.Errorf("step stopped: %s", known.Reason)
				return
			}
			panic(value)
		}
	}()
	step()
	return nil
}
func main() {
	err := invoke(func() { defer fmt.Println("정리 실행"); panic(stopped{Reason: "demonstration"}) })
	fmt.Println("경계 결과:", err)
	fmt.Println("호출자 계속")
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/11-errors-defer/panic-boundary
```

예상 출력:

```text
정리 실행
경계 결과: step stopped: demonstration
호출자 계속
```


`step`의 defer가 먼저 실행됩니다. 그다음 invoke의 defer에서 recover가 panic 값을 받습니다. 알려진 stopped 타입이면 error로 바꾸고, 모르는 타입이면 다시 panic하여 전파합니다.

`panic` 다음 문장부터 step이 다시 계속되는 것은 아닙니다. 복구한 invoke가 반환하고, invoke를 호출한 main이 다음 문장을 실행합니다.

| 위치 | 역할 |
| --- | --- |
| step 안의 defer | step이 소유한 정리 |
| invoke의 defer | 좁은 경계에서 알려진 panic 판별 |
| main | 반환된 오류를 처리하고 다음 흐름 결정 |

### 8.3 recover가 동작하는 조건

recover는 같은 고루틴에서 panic을 풀어 가는 과정 중 **직접 호출되는 deferred 함수 안**에서 사용해야 의미가 있습니다. 평범한 본문에서 recover를 호출한다고 나중에 발생할 panic을 예약해서 잡는 것은 아닙니다.

다른 고루틴의 panic은 여기의 defer로 잡지 못합니다. 고루틴과 오류 전달은 18~20장에서 다룹니다. 지금은 복구 경계가 현재 실행 흐름에 한정된다는 점을 기억하세요.

알 수 없는 panic을 모두 잡고 nil을 반환하면 프로그램 버그를 성공처럼 숨길 수 있습니다. 참고 구현은 자신이 처리하기로 정한 stopped만 변환하고 나머지는 그대로 전파합니다.

### 8.4 defer가 항상 실행되는 것은 아닙니다

일반적인 함수 반환과 panic 전파에서는 defer가 실행됩니다. 하지만 `os.Exit`는 프로그램을 즉시 종료하므로 defer를 실행하지 않습니다. `log.Fatal` 계열도 종료 동작을 포함합니다. [os.Exit](https://pkg.go.dev/os#Exit), [log.Fatal](https://pkg.go.dev/log#Fatal)을 참고합니다.

자원을 가진 함수 안에서 오류를 만나 곧바로 os.Exit를 호출하면 정리 등록의 이점이 사라집니다. 보통 작업 함수가 오류나 종료 코드를 반환하고, 가장 바깥 main에서 필요한 정리가 끝난 뒤 종료 정책을 적용합니다. 16장에서 CLI 종료 코드를 구체적으로 구현합니다.

## 9. 오류 처리 흐름을 설계하기

### 9.1 실패 시 무엇이 남는지 먼저 정하기

함수 시그니처가 `(값, error)`라고 해서 상태 변화 정책까지 자동으로 정해지는 것은 아닙니다. 다음 질문을 명시합니다.

1. 오류가 있으면 반환된 값은 사용할 수 있나요?
2. 입력 데이터나 공유 상태를 일부 바꿀 수 있나요?
3. 호출자는 같은 입력을 다시 시도해도 되나요?
4. 이 함수가 얻거나 받은 자원을 누가 닫나요?
5. 원인 오류를 호출자가 분류할 수 있나요?

예를 들어 카운터 변경은 검증을 끝내기 전 상태를 바꾸지 않도록 만들 수 있습니다. 파일 쓰기는 중간에 일부 바이트가 이미 저장되었을 수 있으므로 전혀 다른 계약이 필요합니다. 같은 error 문법이라도 실제 작업의 의미는 다릅니다.

### 9.2 반환·기록·재시도의 책임

낮은 수준 함수가 오류를 출력하고 다시 반환하면 여러 상위 계층이 같은 오류를 반복 기록할 수 있습니다. 함수는 필요한 문맥을 추가해 반환하고, 실행 경계를 가진 계층이 한 번 기록하도록 구성할 수 있습니다.

| 계층 | 학습 예제의 기본 책임 |
| --- | --- |
| 작은 계산·검증 함수 | 값과 오류 반환 |
| 작업을 조합하는 함수 | 실패 단계 문맥 추가, 소유 자원 정리 |
| CLI의 실행 경계 | 사용자 메시지와 종료 코드 선택 |
| HTTP의 실행 경계 | 응답 상태·공개 메시지·내부 로그 분리 |

이는 모든 프로그램에 강제되는 언어 규칙이 아니라 중복과 숨은 부작용을 줄이기 위한 구성 방식입니다. 로그와 서비스 운영은 25장에서 이어집니다.

### 9.3 재시도는 오류가 있다는 이유만으로 하지 않기

같은 요청을 반복하면 중복 처리될 수 있습니다. 입력이 잘못되었거나 권한이 없으면 기다려도 결과가 달라지지 않을 수 있습니다. 재시도가 필요한 작업은 실패 종류, 부작용, 횟수와 취소 조건을 함께 정해야 합니다.

이 장의 예제는 오류를 올바르게 전달하고 구별하는 데 집중합니다. 최종 동시 배치 실습에서 재시도 가능한 실패를 따로 분류합니다.

## 10. 자주 하는 실수

| 증상 | 원인 후보 | 수정 방향 |
| --- | --- | --- |
| 실패했는데 0을 정상 결과로 사용 | err 검사 누락 | 결과 사용 전 오류 확인 |
| 같은 메시지인데 Is가 false | 새 errors.New로 다른 값 생성 | 공유된 대표 오류 또는 타입 계약 사용 |
| 바깥 문맥을 붙인 뒤 원인을 못 찾음 | %v로 문자열만 포맷 | 공개할 원인이면 %w 사용 |
| As에서 panic | 대상 인자 형태가 잘못됨 | 받을 변수의 주소 전달 |
| Error 메서드에서 panic | error 안의 nil 포인터 | 성공 경로에서 진짜 nil 반환 |
| 파일을 닫지 못함 | 성공 반환 경로에만 Close | 획득 성공 뒤 defer 또는 명시적 수명 분리 |
| 처리 오류가 사라짐 | 닫기 오류로 덮어씀 | 오류 우선순위 계약 또는 Join |
| 반복 처리 중 자원이 계속 쌓임 | 반복 안의 defer를 즉시 정리로 오해 | 항목별 함수 수명 또는 명시적 Close |
| defer에서 err를 바꿨는데 결과가 nil | 다른 지역 err를 새로 선언 | 이름 붙인 반환 변수와 대입 확인 |
| panic 뒤 문장이 실행되지 않음 | recover를 재개 기능으로 오해 | 복구한 함수의 반환과 호출자 흐름 확인 |
| 오류 처리 후 정리가 안 됨 | 중간에서 os.Exit/log.Fatal | 바깥 종료 경계로 이동 |

### 10.1 실패를 재현하는 테스트 입력

오류 처리는 성공 예제 한 번으로 검증하기 어렵습니다. 읽기와 닫기를 따로 실패하게 만들어 보면 한 오류가 다른 오류를 덮는 문제를 찾을 수 있습니다.

이 장의 resource 테스트는 읽기 성공/실패와 닫기 성공/실패의 네 조합을 모두 실행합니다. 닫기 횟수가 정확히 한 번인지도 확인합니다. 실제 디스크 고장을 기다리지 않고 인터페이스의 작은 구현으로 실패를 재현하는 방법입니다.

### 10.2 예측을 먼저 적기

defer 코드를 읽을 때는 등록과 실행을 같은 줄에서 한 번에 이해하려 하지 않습니다. 먼저 등록 순서와 저장되는 인자를 표에 적습니다. 반환 시점에는 표를 역순으로 읽으며 최종 변수와 출력이 어떻게 되는지 확인합니다.

## 11. 연습문제

[문제 문서](../../exercises/11-errors-defer/README.md)에서 오류와 자원 수명 계약을 직접 구현합니다. [해설](../../solutions/11-errors-defer/README.md)은 시도한 뒤 확인하세요.

| 번호 | 수준 | 목표 |
| --- | --- | --- |
| 1 | 확인 | 이름 붙인 반환값과 defer 출력 추적 |
| 2 | 변형 | 포트 숫자 형식·범위 오류 분리 |
| 3 | 변형 | 래핑과 오류 검사 오류 진단 |
| 4 | 응용 | 조회 오류 타입으로 키와 원인 보존 |
| 5 | 응용 | 처리·정리 오류와 정확히 한 번 닫기 |
| 6 | 응용 | 여러 필드 오류를 Join으로 반환 |
| 선택 A | 확인 | panic 복구 범위와 종료 정책 설명 |

```sh
go test ./examples/11-errors-defer/...
go test ./solutions/11-errors-defer/...
```

## 12. 요약과 자기 점검

- error는 실패를 전달하는 값이며, 성공은 보통 nil입니다.
- 오류의 문장과 오류 종류·타입은 서로 다른 정보입니다.
- `%w`는 문맥과 원인 관계를 보존하며, Is와 As는 그 관계 안에서 비교·탐색합니다.
- defer의 인자는 등록할 때 평가되고, 실행은 함수 종료 시 역순입니다.
- 자원 소유권과 실패 시 반환값·상태 계약을 함께 정해야 합니다.
- panic/recover는 일반 입력 오류의 기본 흐름을 대신하지 않으며, 좁은 경계에서 이해한 실패만 처리합니다.

스스로 답해 보세요.

1. 정상 결과 0과 실패 결과 0을 무엇으로 구별하나요?
2. errors.New를 같은 문장으로 두 번 호출하면 같은 오류인가요?
3. `%v`와 `%w`의 차이가 호출자의 어떤 동작에 영향을 주나요?
4. errors.As에서 포인터 변수의 주소를 전달하는 이유는 무엇인가요?
5. defer 함수의 인자를 평가하는 시점과 함수 본문 실행 시점은 언제인가요?
6. return 식의 값은 defer에서 이름 붙인 반환값을 바꾸기 전에 설정되나요?
7. 처리와 Close가 모두 실패하면 어떤 원인을 보존해야 하나요?
8. nil `*InputError`를 error로 반환하면 왜 성공으로 보이지 않나요?
9. recover가 다른 고루틴의 panic을 잡을 수 있나요?
10. os.Exit를 자원 처리 함수 안에서 호출하면 어떤 문제가 생기나요?

## 참고 자료

공식 문서 확인일: **2026-11-09**.

- [Go 명세: Errors](https://go.dev/ref/spec#Errors), [Defer statements](https://go.dev/ref/spec#Defer_statements), [Handling panics](https://go.dev/ref/spec#Handling_panics)
- [errors 패키지](https://pkg.go.dev/errors), [Is](https://pkg.go.dev/errors#Is), [As](https://pkg.go.dev/errors#As), [Join](https://pkg.go.dev/errors#Join)
- [fmt.Errorf](https://pkg.go.dev/fmt#Errorf), [strconv.NumError](https://pkg.go.dev/strconv#NumError)
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Go FAQ: nil error](https://go.dev/doc/faq#nil_error)
- [io.Reader](https://pkg.go.dev/io#Reader), [io.Closer](https://pkg.go.dev/io#Closer), [os.Exit](https://pkg.go.dev/os#Exit)

## 다음 학습

다음 장에서는 정상·경계·오류 동작을 `testing`으로 자동 확인합니다. 이번 장에서 명시한 계약이 테스트의 기대 결과가 됩니다.

[이전: 10장 패키지와 모듈](../10-packages-modules/README.md) · [다음: 12장 자동 테스트](../12-testing/README.md) · [전체 목차](../README.md)
