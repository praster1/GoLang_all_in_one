# 09장 연습문제: 메서드와 인터페이스

[본문](../../book/09-methods-interfaces/README.md)을 읽고 풉니다. 인터페이스는 같은 함수 이름을 가진 값을 아무렇게나 묶는 장치가 아니라, 인자·결과까지 일치하는 메서드의 약속입니다. 아래 문제에서는 그 약속과 값의 변화를 함께 확인합니다.

## 풀이 준비

모든 명령은 저장소 루트에서 실행합니다. 개인 실습은 `_practice/09-methods-interfaces/` 아래 문제별 디렉터리에 저장합니다. 편집기로 `predict`, `rectangle`, `repair`, `limited-counter`, `render`, `count-types`, `embedding` 폴더를 각각 만들고 각 폴더에 `main.go` 하나를 작성하세요. 기존 파일이 있으면 덮어쓰지 않습니다.

```sh
go run ./_practice/09-methods-interfaces/predict
```

새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice` 폴더는 전체 `./...` 검사에 자동으로 포함되지 않으므로 직접 경로를 지정합니다. 테스트 작성법은 12장에서 배우며, 지금은 `main`에서 경계 입력을 넣어 확인해도 됩니다.

풀이 순서:

1. 읽기 동작인지 변경 동작인지 나눕니다.
2. 값 수신자와 포인터 수신자 중 필요한 것을 정합니다.
3. nil 인터페이스와 nil 포인터에 대한 계약을 별도로 적습니다.
4. 직접 실행한 결과와 예상 결과를 비교합니다.
5. 막히면 각 문제의 힌트를 보고, 마지막에 [해설](../../solutions/09-methods-interfaces/README.md)을 확인합니다.

## 문제 1. 복사·공유·저장된 메서드

**수준: 확인**

다음 전체 코드를 개인 `predict/main.go`에 작성합니다. 실행 전에 세 줄 출력을 적으세요.

```go
package main

import "fmt"

type Counter struct{ N int }
func (c Counter) Read() int { return c.N }
func (c *Counter) Add(n int) { c.N += n }
type Reader interface{ Read() int }

func main() {
    c := Counter{N: 2}
    var copied Reader = c
    var shared Reader = &c
    saved := c.Read
    c.Add(3)
    fmt.Println(c.Read(), copied.Read(), shared.Read())
    fmt.Println(saved())
    var missing *Counter
    var value Reader = missing
    fmt.Println(missing == nil, value == nil)
}
```

질문:

- `copied`에 보관되는 값과 `shared`에 보관되는 값은 각각 무엇인가요?
- `saved := c.Read`는 호출 시점과 저장 시점 중 언제의 수신자를 사용하나요?
- 마지막 `value`의 동적 타입과 동적 값은 무엇인가요?
- 마지막 줄 다음에 `value.Read()`를 추가하면 안전한가요? 값 수신자 메서드와 nil 포인터의 관계를 설명하세요.

완료 조건: 출력뿐 아니라 값 복사, 포인터 공유, 메서드 값 저장 규칙을 분리하여 설명합니다.

<details>
<summary>힌트 1</summary>

인터페이스에 구체 값을 넣는 것도 대입입니다. 구조체 값을 넣으면 구조체 값이 복사되고, 포인터를 넣으면 포인터 값이 복사됩니다.

</details>

<details>
<summary>힌트 2</summary>

`missing == nil`은 포인터 값만 확인합니다. `value == nil`은 동적 타입도 없는 인터페이스인지를 확인합니다.

</details>

## 문제 2. 직사각형의 읽기 메서드

**수준: 변형**

`Rectangle`은 정수 `Width`, `Height` 필드를 가집니다. 두 값은 0 이상 1,000 이하라고 가정합니다. 입력 검증은 이 문제의 범위가 아닙니다.

다음 메서드를 구현하세요.

```go
func (rectangle Rectangle) Area() int
func (rectangle Rectangle) Perimeter() int
```

두 메서드는 원본 필드를 바꾸지 않습니다. 넓이는 가로×세로, 둘레는 2×(가로+세로)입니다. 한 변이 0이어도 정의한 수식 그대로 계산합니다.

| 입력 | 넓이 | 둘레 |
| --- | ---: | ---: |
| 3, 4 | 12 | 14 |
| 0, 0 | 0 | 0 |
| 0, 5 | 0 | 10 |
| 1,000, 1,000 | 1,000,000 | 4,000 |

완료 조건: 위 입력들을 실행하고, 호출 전후 구조체 값이 같은지 확인합니다. 왜 포인터 수신자가 필요하지 않은지 설명하세요.

<details>
<summary>힌트</summary>

메서드는 `main` 밖에 선언합니다. 수신자 값의 필드를 읽어 계산한 결과를 반환하세요. 화면 출력은 `main`이 담당합니다.

</details>

## 문제 3. 호출은 되는데 대입은 안 됩니다

**수준: 변형**

아래 프로그램은 의도적으로 컴파일되지 않습니다. 기존 정상 예제에 넣지 말고 개인 `repair/main.go`에서만 확인합니다.

```go
package main

import "fmt"

type Counter struct{ N int }
func (c *Counter) Add(n int) { c.N += n }
type Adder interface{ Add(int) }
func addTwo(target Adder) { target.Add(2) }

func main() {
    c := Counter{N: 3}
    var target Adder = c
    addTwo(target)
    fmt.Println(c.N)
}
```

`Add`의 수신자와 함수 시그니처를 바꾸지 않고 대입 부분만 고치세요. 출력은 `5`여야 합니다.

질문:

- `c.Add(2)`가 컴파일되어도 `var target Adder = c`가 실패하는 이유는 무엇인가요?
- `Add`를 값 수신자로 바꾸면 대입 오류는 없어질 수 있습니다. 그때 원본 변경이라는 요구도 만족하나요?

완료 조건: `Counter`와 `*Counter`의 메서드 집합을 적고, 자동 주소 취득과 인터페이스 구현 여부를 구별합니다.

<details>
<summary>힌트</summary>

필요한 메서드는 `*Counter`의 메서드 집합에 있습니다. 그 타입의 값을 전달하세요.

</details>

## 문제 4. 한도 있는 카운터

**수준: 응용**

정수 비공개 필드 `value`를 가진 `Counter`를 만듭니다. 정상 상태는 0 이상 100 이하입니다. 영값은 정상 상태입니다.

```go
func (counter Counter) Value() int
func (counter *Counter) Add(amount int) bool
func (counter *Counter) Reset() bool
```

계약:

- `Value`는 현재 값을 읽습니다. nil 포인터에 호출하는 동작은 지원하지 않습니다.
- `Add`는 수신자가 nil이거나 현재 상태가 범위를 벗어나거나, 음수 증가량이거나, 결과가 100을 넘으면 false입니다.
- 실패한 `Add`는 기존 상태를 바꾸지 않습니다.
- 0을 더하거나 결과가 정확히 100이 되는 것은 성공입니다.
- `Reset`은 nil이면 false, 그 외에는 현재 상태가 무엇이든 0으로 만들고 true입니다.

| 시작 값 | 증가량 | 성공 | 최종 값 |
| ---: | ---: | --- | ---: |
| 0 | 0 | true | 0 |
| 0 | 100 | true | 100 |
| 60 | 41 | false | 60 |
| 3 | -1 | false | 3 |
| 101 | 0 | false | 101 |

완료 조건: 경계값과 nil을 검사하며, 실패 시 상태가 유지되는지 확인합니다. 정수 최댓값을 증가량으로 넣어도 범위 검사가 덧셈 오버플로 때문에 잘못 통과하지 않아야 합니다.

<details>
<summary>힌트 1</summary>

기존 상태가 정상임을 확인한 뒤, 증가량을 남은 여유 `100-counter.value`와 비교하면 큰 양수를 먼저 더할 필요가 없습니다.

</details>

<details>
<summary>힌트 2</summary>

`counter == nil` 검사를 필드 접근보다 먼저 합니다. 값 수신자 `Value`까지 nil을 지원한다고 가정하지 마세요.

</details>

## 문제 5. 여러 구현의 설명을 모으기

**수준: 응용**

다음 계약을 구현합니다.

```go
type Describer interface{ Describe() string }
func render(values []Describer) ([]string, bool)
```

`*Note`는 메모를, 정의된 정수 타입 `Number`는 숫자를 설명합니다.

- `&Note{Text: "Go"}`의 설명: `메모: Go`.
- `Number(7)`의 설명: `숫자: 7`.
- nil `*Note`의 설명: `메모 없음`. 이 타입의 메서드가 명시적으로 nil을 처리합니다.
- `render`는 입력 순서대로 결과를 반환합니다.
- 입력에 nil 인터페이스가 하나라도 있으면 `nil, false`를 반환하며, 어떤 `Describe` 메서드도 호출하지 않습니다.
- nil/빈 슬라이스 입력은 성공입니다. 참고 구현은 결과로 nil 슬라이스를 반환합니다.
- 이 함수가 모든 typed nil의 안전성을 보장하는 것은 아닙니다. 각 구현이 자신의 메서드 계약을 지켜야 합니다.

완료 조건: 정상 메모·숫자·nil `*Note`를 섞은 입력은 성공하고, 마지막 요소가 nil 인터페이스인 입력에서는 앞선 요소의 메서드도 호출하지 않게 설계합니다.

<details>
<summary>힌트 1</summary>

입력 검증과 메서드 호출을 두 번의 순회로 나눕니다. 인터페이스 자체가 nil인지 먼저 전부 검사합니다.

</details>

<details>
<summary>힌트 2</summary>

모든 구현에 대해 `reflect`로 내부 포인터를 검사하려 하지 마세요. 이 문제는 인터페이스 nil과 구현별 nil 계약을 분리하는 연습입니다.

</details>

## 문제 6. 동적 타입별 개수 집계

**수준: 응용**

다음 타입과 함수를 구현하세요.

```go
type Counts struct{ Ints, Strings, Bools, Others int }
func countTypes(values []any) Counts
```

정확히 `int`, `string`, `bool`인 값의 개수를 해당 필드에 더하고 나머지는 `Others`에 더합니다. `nil`, `int64`, 별도로 정의한 `type Count int`, 슬라이스도 기타입니다.

입력:

```go
[]any{1, "Go", true, int64(2), nil, "한글", false}
```

출력 계약:

```text
int: 1, string: 2, bool: 2, 기타: 2
```

완료 조건: 빈 입력은 모든 필드가 0입니다. `true`와 `false`는 모두 bool 개수에 포함하며, 문자열 내용이나 숫자 크기로 분류하지 않습니다.

<details>
<summary>힌트</summary>

타입 스위치는 값이 아니라 동적 타입으로 분기합니다. `case int:`는 `int`에 기반을 둔 모든 정의 타입을 포함하지 않습니다.

</details>

## 선택 A. 임베딩한 동작 조합

**수준: 응용**

`Label`의 `Describe`는 `Text`를 반환합니다. `Task`는 `Label`을 임베딩하고 `Done bool` 필드를 가집니다. 바깥 타입의 `Describe`는 내부 `Label.Describe()` 결과 뒤에 ` / 진행` 또는 ` / 완료`를 붙입니다.

완료 조건: `task.Describe()`는 상태까지 반환하고, `task.Label.Describe()`는 원래 레이블만 반환합니다. 바깥 메서드가 자기 자신을 재귀 호출하지 않아야 합니다.

<details>
<summary>힌트</summary>

바깥 메서드 안에서 `task.Describe()`를 다시 호출하지 않습니다. 임베딩된 필드를 명시해 호출할 메서드를 선택합니다.

</details>

## 완료 체크리스트

- [ ] 읽기와 변경 메서드의 수신자를 설명했습니다.
- [ ] 값 복사와 포인터 공유를 실제 출력으로 확인했습니다.
- [ ] nil 인터페이스와 typed nil을 구별했습니다.
- [ ] 카운터의 실패 시 상태 보존을 확인했습니다.
- [ ] 구현을 추가해도 `render` 본문을 바꿀 필요가 없는 이유를 설명했습니다.
- [ ] 타입 스위치를 값 변환과 혼동하지 않습니다.

[본문으로](../../book/09-methods-interfaces/README.md) · [해설 확인](../../solutions/09-methods-interfaces/README.md)
