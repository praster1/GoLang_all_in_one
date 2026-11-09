# 09장 해설: 메서드와 인터페이스

먼저 [연습문제](../../exercises/09-methods-interfaces/README.md)를 풀어 보세요. 각 참고 구현은 독립된 프로그램입니다. 모든 명령은 저장소 루트에서 실행합니다.

## 문제 1. 값의 저장 시점과 대상

첫 줄은 `5 2 5`, 둘째 줄은 `2`, 마지막 줄은 `true false`입니다.

| 이름 | 보관하는 것 | 변경 뒤 관찰 |
| --- | --- | --- |
| `c` | 실제 Counter 변수 | 5 |
| `copied` | 대입 당시 Counter 값 | 2 |
| `shared` | c를 가리키는 포인터 값 | 5 |
| `saved` | 저장 당시 값 수신자가 결합된 함수 | 2 |
| `value` | 동적 타입 *Counter, 동적 값 nil | 인터페이스 자체는 nil 아님 |

`value.Read()`는 안전하지 않습니다. `Read`는 값 수신자 메서드이므로 nil 포인터에서 필요한 구조체 값을 만들 수 없습니다. nil 여부 출력만 하는 현재 프로그램은 정상입니다.

참고 구현: [predict/main.go](predict/main.go).

```go
package main

import "fmt"

type Counter struct{ N int }

func (c Counter) Read() int  { return c.N }
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

```sh
go run ./solutions/09-methods-interfaces/predict
go test ./solutions/09-methods-interfaces/predict
```

예상 출력:

```text
5 2 5
2
true false
```


## 문제 2. 읽기 메서드는 계산 결과를 반환

이 입력 범위의 곱과 합은 지원하는 int 범위 안에 들어갑니다. 이 문제는 임의 크기 정수의 곱셈을 안전하게 하는 문제가 아닙니다. 값 수신자는 원본을 바꾸지 않고 두 정수 필드를 읽는 목적에 맞습니다.

참고 구현: [rectangle/main.go](rectangle/main.go).

```go
package main

import "fmt"

type Rectangle struct{ Width, Height int }

func (rectangle Rectangle) Area() int      { return rectangle.Width * rectangle.Height }
func (rectangle Rectangle) Perimeter() int { return 2 * (rectangle.Width + rectangle.Height) }

func main() {
	rectangle := Rectangle{Width: 3, Height: 4}
	fmt.Println("넓이:", rectangle.Area())
	fmt.Println("둘레:", rectangle.Perimeter())
	fmt.Println("원본:", rectangle.Width, rectangle.Height)
}
```

```sh
go run ./solutions/09-methods-interfaces/rectangle
go test ./solutions/09-methods-interfaces/rectangle
```

예상 출력:

```text
넓이: 12
둘레: 14
원본: 3 4
```


`rectangle`의 `Width`나 `Height`를 대입하는 코드가 없으므로 호출 뒤에도 3과 4입니다. 포인터 수신자를 사용해도 읽기만 한다면 계산은 가능하지만, 원본 변경이 필요하다는 인상을 주지 않도록 여기서는 값 수신자를 사용했습니다.

## 문제 3. 인터페이스에는 *Counter를 전달

`Counter`에는 `Add`가 포함되지 않고, `*Counter`에는 포함됩니다. 지역 변수 `c`에서 `c.Add(2)`를 쓰는 자동 주소 취득은 직접 메서드 호출에 적용되는 규칙입니다. 인터페이스 대입이 임의로 `&c`를 만드는 것은 아닙니다.

참고 구현: [repair/main.go](repair/main.go).

```go
package main

import "fmt"

type Counter struct{ N int }

func (c *Counter) Add(n int) { c.N += n }

type Adder interface{ Add(int) }

func addTwo(target Adder) { target.Add(2) }
func main() {
	c := Counter{N: 3}
	var target Adder = &c
	addTwo(target)
	fmt.Println(c.N)
}
```

```sh
go run ./solutions/09-methods-interfaces/repair
go test ./solutions/09-methods-interfaces/repair
```

예상 출력:

```text
5
```


값 수신자로 바꾸면 복사본의 `N`을 수정하므로 원본이 5가 되지 않습니다. 요구사항은 대입을 컴파일시키는 것과 원본을 바꾸는 것을 모두 포함합니다.

## 문제 4. 검증을 끝낸 뒤 한 번만 상태 변경

nil 수신자는 필드 접근 전에 거부합니다. 이어서 기존 상태와 증가량을 확인합니다. `amount > 100-counter.value`는 정상 상태에서 0~100 범위의 여유와 비교하므로 `counter.value+amount`를 미리 계산하지 않습니다.

참고 구현: [limited-counter/main.go](limited-counter/main.go).

```go
package main

import "fmt"

type Counter struct{ value int }

func (counter Counter) Value() int { return counter.value }
func (counter *Counter) Add(amount int) bool {
	if counter == nil || counter.value < 0 || counter.value > 100 || amount < 0 || amount > 100-counter.value {
		return false
	}
	counter.value += amount
	return true
}
func (counter *Counter) Reset() bool {
	if counter == nil {
		return false
	}
	counter.value = 0
	return true
}

func main() {
	counter := Counter{}
	ok := counter.Add(60)
	fmt.Println("추가:", ok, counter.Value())
	ok = counter.Add(41)
	fmt.Println("한도 초과:", ok, counter.Value())
	ok = counter.Reset()
	fmt.Println("초기화:", ok, counter.Value())
}
```

```sh
go run ./solutions/09-methods-interfaces/limited-counter
go test ./solutions/09-methods-interfaces/limited-counter
```

예상 출력:

```text
추가: true 60
한도 초과: false 60
초기화: true 0
```


실패 경로는 `counter.value += amount`에 도달하지 않습니다. 따라서 실패 시 상태 보존은 각 분기에서 되돌리는 코드를 써서 얻는 결과가 아니라, 변경 시점을 뒤로 미뤄 얻는 결과입니다.

`Reset`은 비정상 상태도 회복합니다. 반면 `Value`는 값 수신자이므로 nil 포인터에 호출하지 않습니다. 다른 정책도 만들 수 있지만 메서드별 계약을 함께 바꿔야 합니다.

## 문제 5. 검증과 호출의 두 단계

nil 인터페이스를 하나라도 발견하면 아무 구현도 호출하지 않아야 합니다. 한 번의 순회에서 즉시 호출하면 뒤늦게 nil을 발견하기 전에 앞선 메서드가 이미 실행됩니다. 두 번 순회하는 이유가 여기에 있습니다.

참고 구현: [render/main.go](render/main.go).

```go
package main

import "fmt"

type Describer interface{ Describe() string }
type Note struct{ Text string }

func (note *Note) Describe() string {
	if note == nil {
		return "메모 없음"
	}
	return "메모: " + note.Text
}

type Number int

func (number Number) Describe() string { return fmt.Sprintf("숫자: %d", number) }

func render(values []Describer) ([]string, bool) {
	for _, value := range values {
		if value == nil {
			return nil, false
		}
	}
	var result []string
	for _, value := range values {
		result = append(result, value.Describe())
	}
	return result, true
}

func main() {
	var missing *Note
	lines, ok := render([]Describer{&Note{Text: "Go"}, Number(7), missing})
	fmt.Println("성공:", ok)
	for _, line := range lines {
		fmt.Println(line)
	}
	_, ok = render([]Describer{nil})
	fmt.Println("nil 인터페이스:", ok)
}
```

```sh
go run ./solutions/09-methods-interfaces/render
go test ./solutions/09-methods-interfaces/render
```

예상 출력:

```text
성공: true
메모: Go
숫자: 7
메모 없음
nil 인터페이스: false
```


`missing`은 `*Note`라는 동적 타입을 가지므로 `value == nil` 검사를 통과합니다. `Note.Describe`가 nil 수신자를 처리하여 `메모 없음`을 반환합니다. 다른 구현이 nil 수신자를 처리하지 않는다면 `render`는 그 구현의 패닉까지 자동으로 막아 주지 않습니다.

`render`의 입력 타입은 `[]Describer`입니다. `[]*Note`를 통째로 같은 타입으로 대입할 수는 없으며 필요한 경우 원소별로 인터페이스 슬라이스를 구성합니다.

## 문제 6. 같은 기반 타입과 같은 동적 타입은 다릅니다

`int64(2)`는 수치가 작아도 `int`가 아닙니다. `type Count int`로 정의한 새 타입도 동적 타입이 `Count`입니다. `type Count = int`처럼 별칭을 만들면 새 타입을 만들지 않으므로 int 분기에 들어갑니다.

참고 구현: [count-types/main.go](count-types/main.go).

```go
package main

import "fmt"

type Counts struct{ Ints, Strings, Bools, Others int }

func countTypes(values []any) Counts {
	var result Counts
	for _, value := range values {
		switch value.(type) {
		case int:
			result.Ints++
		case string:
			result.Strings++
		case bool:
			result.Bools++
		default:
			result.Others++
		}
	}
	return result
}

func main() {
	values := []any{1, "Go", true, int64(2), nil, "한글", false}
	result := countTypes(values)
	fmt.Printf("int: %d, string: %d, bool: %d, 기타: %d\n", result.Ints, result.Strings, result.Bools, result.Others)
}
```

```sh
go run ./solutions/09-methods-interfaces/count-types
go test ./solutions/09-methods-interfaces/count-types
```

예상 출력:

```text
int: 1, string: 2, bool: 2, 기타: 2
```


## 선택 A. 내부 메서드를 명시하여 조합

임베딩은 모든 메서드를 바깥 타입 대상으로 다시 실행하는 상속 기능이 아닙니다. `task.Label.Describe()`는 내부 `Label`을 수신자로 사용합니다. 바깥 메서드는 이 결과에 상태 문자열을 덧붙입니다.

참고 구현: [embedding/main.go](embedding/main.go).

```go
package main

import "fmt"

type Label struct{ Text string }

func (label Label) Describe() string { return label.Text }

type Task struct {
	Label
	Done bool
}

func (task Task) Describe() string {
	state := "진행"
	if task.Done {
		state = "완료"
	}
	return task.Label.Describe() + " / " + state
}
func main() {
	task := Task{Label: Label{Text: "Go 읽기"}, Done: true}
	fmt.Println(task.Describe())
	fmt.Println(task.Label.Describe())
}
```

```sh
go run ./solutions/09-methods-interfaces/embedding
go test ./solutions/09-methods-interfaces/embedding
```

예상 출력:

```text
Go 읽기 / 완료
Go 읽기
```


## 본문 자기 점검 답

1. 값 수신자는 구조체 값의 복사본을 받습니다. int 필드의 변경은 원본과 별개입니다.
2. 주소를 얻을 수 있는 변수의 직접 호출은 주소 취득 편의가 적용될 수 있습니다. 인터페이스 구현은 메서드 집합으로 판단합니다.
3. 같은 이름이어도 반환 타입이 다르면 필요한 메서드 시그니처와 다릅니다.
4. 동적 타입이 있으므로 인터페이스 전체 비교는 false입니다.
5. 단언은 변환을 수행하지 않습니다. int64를 확인한 뒤 명시적으로 int로 변환해야 합니다.
6. 동적 타입이 비교 불가능하면 인터페이스 비교에서 패닉이 발생할 수 있습니다.
7. 승격된 메서드의 실제 수신자는 임베딩된 내부 타입입니다.
8. 함수는 인터페이스의 약속만 사용하며, 새 구현도 그 약속을 만족하기 때문입니다.

## 검증과 적용 범위

참고 구현의 출력 예제 테스트와 경계 테스트는 다음 명령으로 실행합니다.

```sh
go test ./examples/09-methods-interfaces/... ./solutions/09-methods-interfaces/...
```

카운터 테스트는 0·100 경계, 음수, 잘못된 기존 상태와 nil을 확인합니다. `render` 테스트는 잘못된 입력에서 앞선 구현의 메서드조차 호출하지 않는다는 계약을 호출 횟수로 확인합니다. 타입 집계 테스트는 별도 정의 정수 타입, nil 포인터와 슬라이스를 기타로 분류하는지 검사합니다.

메서드가 있는 모든 타입을 인터페이스로 만들 필요는 없습니다. 서로 다른 구현을 같은 방식으로 쓰거나 소비 측의 필요한 행동을 분리할 때 작은 인터페이스부터 도입합니다.

[문제로 돌아가기](../../exercises/09-methods-interfaces/README.md) · [본문](../../book/09-methods-interfaces/README.md)
