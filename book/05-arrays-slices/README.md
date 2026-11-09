# 05. 배열과 슬라이스

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 입문 | [04. 함수와 스코프](../04-functions-scope/README.md) | Go, 편집기, 터미널, 이 저장소 |

하루 공부 시간을 `monday`, `tuesday`처럼 각각의 변수에 저장하면 날짜가 늘 때마다 선언과 계산을 추가해야 합니다. 여러 값을 한 묶음으로 다루면 반복문 하나로 합계를 구하거나 조건에 맞는 값만 고를 수 있습니다.

이번 장에서는 **배열(array)**과 **슬라이스(slice)**를 사용합니다. 특히 두 변수가 같은 원소 저장 공간을 사용할 때 무엇이 함께 바뀌는지 실행해 확인합니다. 04장의 “인자는 값으로 전달된다”는 규칙을 더 구체적으로 이해하는 장이기도 합니다.

## 학습 목표

- 고정 길이 배열과 길이를 바꿔 다루는 슬라이스를 구분한다.
- 위치, 길이, 용량을 읽고 유효한 인덱스와 부분 범위를 계산한다.
- `append`의 반환값과 저장 공간 재사용을 설명한다.
- `copy`로 원소를 독립적으로 보관하고 공유 변경을 검사한다.
- 함수로 필터·역순 복사·삭제·누적 계산을 구현한다.

## 사전 지식과 준비

정수, 변수 대입, `if`, `for`, 함수의 매개변수와 반환값을 사용합니다. 가변 인자에서 보았던 `[]int`를 이번 장에서 자세히 다룹니다. 포인터 문법이나 구조체를 미리 알 필요는 없습니다.

모든 명령은 **저장소 루트 `GoLang_all_in_one/`**에서 실행합니다. 하위 폴더에 모듈을 새로 만들지 않습니다. 전체 코드는 각 링크의 실제 파일과 같습니다. **main 안의 발췌**는 `package main`, `import "fmt"`, `func main()`이 있는 개인 파일의 본문에 넣습니다.

| 예제 | 중심 질문 | 실행 명령 |
| --- | --- | --- |
| arrays | 배열을 함수에 전달하면 원본도 바뀔까? | `go run ./examples/05-arrays-slices/arrays` |
| slice-basics | nil, 빈 값, 길이와 용량은 무엇일까? | `go run ./examples/05-arrays-slices/slice-basics` |
| slice-view | 부분 범위는 복사본일까? | `go run ./examples/05-arrays-slices/slice-view` |
| append-growth | 추가할 때 저장 공간은 언제 달라질까? | `go run ./examples/05-arrays-slices/append-growth` |
| copy | 원소를 독립적으로 보관하려면? | `go run ./examples/05-arrays-slices/copy` |
| slice-functions | 원소 변경과 길이 변경은 어떻게 다를까? | `go run ./examples/05-arrays-slices/slice-functions` |
| filter | 입력을 보존하며 결과를 고르려면? | `go run ./examples/05-arrays-slices/filter` |

개인 파일은 11절의 보존형 준비 명령으로 만듭니다. 기존 예제 파일을 직접 고치는 대신 `_practice/05-arrays-slices/lab/main.go`에서 변형을 실행합니다.

## 1. 배열: 개수가 정해진 값 묶음

배열은 같은 타입의 원소를 정해진 개수만큼 가집니다. `[3]int`는 정수 세 개를 가진 타입입니다. 대괄호 안의 `3`도 타입의 일부이므로 `[3]int`와 `[4]int`는 서로 다른 타입입니다. [배열 타입](https://go.dev/ref/spec#Array_types)을 참고합니다.

배열의 원소는 **인덱스(index)**라는 정수 위치로 읽습니다. 첫 위치는 0입니다. 길이 3인 배열의 유효한 위치는 0, 1, 2입니다.

| 표현 | 의미 |
| --- | --- |
| `var values [3]int` | 정수 세 개를 모두 0으로 초기화 |
| `[3]int{20, 30}` | 앞의 두 값을 지정하고 나머지는 0 |
| `[3]int{20, 30, 40}` | 세 원소를 모두 지정 |
| `values[0]` | 첫 번째 원소 |
| `len(values)` | 원소 개수인 3 |

중괄호로 실제 값을 만드는 표현은 **복합 리터럴(composite literal)**입니다. 선언의 `[3]int`는 타입이고, 뒤의 `{20, 30}`은 초기 원소를 적는 부분입니다.

### 1.1 배열을 복사하여 변경하기

저장 경로: [`examples/05-arrays-slices/arrays/main.go`](../../examples/05-arrays-slices/arrays/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func replaceFirst(values [3]int) [3]int {
	values[0] = 99
	return values
}
func main() {
	values := [3]int{20, 30}
	fmt.Println("초기:", values)
	changed := replaceFirst(values)
	fmt.Println("원본:", values)
	fmt.Println("복사본:", changed)
	fmt.Println("길이:", len(values))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/05-arrays-slices/arrays
```

예상 출력:

```text
초기: [20 30 0]
원본: [20 30 0]
복사본: [99 30 0]
길이: 3
```

`replaceFirst`의 매개변수는 `[3]int`입니다. 호출할 때 배열 값이 복사되므로 함수 안에서 첫 원소를 바꾸어도 호출자의 배열은 유지됩니다. 함수가 바꾼 배열을 반환했기 때문에 `changed`에는 별도 결과가 저장됩니다.

| 단계 | `main`의 `values` | 함수의 `values` 또는 결과 |
| --- | --- | --- |
| 선언 | `[20 30 0]` | 아직 없음 |
| 함수 진입 | `[20 30 0]` | `[20 30 0]` |
| 첫 원소 변경 | `[20 30 0]` | `[99 30 0]` |
| 반환 후 | `[20 30 0]` | `changed`는 `[99 30 0]` |

배열에 정수만 들어 있으므로 이 예에서는 원소 데이터도 독립적입니다. “모든 타입의 복사는 모든 내부 데이터까지 복제한다”는 일반 규칙으로 넓히지는 않습니다. 슬라이스와 구조체의 내부 공유는 뒤에서 구분합니다.

배열은 크기 자체가 계약일 때 유용합니다. 이 장에서는 요일 수 같은 고정 개수와, 실행 중 원소 개수가 달라지는 목록을 구분하는 데 사용합니다.


### 1.2 길이를 컴파일러가 세도록 하기

다음은 `main` 안의 독립된 발췌입니다.

```go
values := [...]int{10, 20, 30}
fmt.Printf("타입: %T, 값: %v\n", values, values)
```

출력은 `타입: [3]int, 값: [10 20 30]`입니다. 배열 리터럴의 `...`는 적힌 원소로 길이를 계산한다는 뜻입니다. 04장의 가변 인자 `...int`와 위치와 역할이 다릅니다.

`[3]int`를 `[4]int` 매개변수에 직접 전달할 수는 없습니다. 개수에 상관없이 처리하는 함수에는 다음 절의 `[]int`를 사용합니다. [복합 리터럴](https://go.dev/ref/spec#Composite_literals)에 배열 리터럴 규칙이 있습니다.

## 2. 인덱스와 범위 오류를 먼저 이해하기

길이가 `n`이면 읽거나 대입할 수 있는 인덱스는 `0 <= i < n`입니다. 마지막 인덱스는 `n`이 아니라 `n-1`입니다. 길이가 0이면 읽을 수 있는 원소가 하나도 없습니다.

| 길이 | 가능한 위치 | 마지막 위치 |
| --- | --- | --- |
| 0 | 없음 | 없음 |
| 1 | 0 | 0 |
| 3 | 0, 1, 2 | 2 |

아래는 개인 파일에서 실행할 수 있는 전체 프로그램입니다.

```go
package main

import "fmt"

func main() {
	values := [3]int{10, 20, 30}
	for i := 0; i < len(values); i++ {
		fmt.Printf("위치 %d: %d\n", i, values[i])
	}
}
```

저장 위치는 `_practice/05-arrays-slices/lab/main.go`이며 준비 명령은 11절에 있습니다.

```sh
go run ./_practice/05-arrays-slices/lab
```

예상 출력:

```text
위치 0: 10
위치 1: 20
위치 2: 30
```

조건을 `i <= len(values)`로 바꾸면 마지막 반복에서 인덱스 3을 읽으려 합니다. 이 변형은 **의도적인 실행 실패 사례**입니다. 오류에서 `index out of range`를 찾고 `<`로 되돌립니다.

상수 인덱스가 배열 범위를 벗어난 것이 컴파일 때 분명하면 컴파일 오류가 날 수 있습니다. 변수 인덱스의 실제 값이 범위를 벗어나면 실행 중 panic이 발생합니다. 둘 다 같은 범위 계약을 어겼지만 발견 시점이 다릅니다. [인덱스 표현식](https://go.dev/ref/spec#Index_expressions)을 참고합니다.

임의 위치를 받는 함수는 읽기 전에 검사합니다. 다음은 **함수 선언 발췌**이며 `main` 바깥에 둡니다.

```go
func valueAt(values []int, index int) (int, bool) {
	if index < 0 || index >= len(values) {
		return 0, false
	}
	return values[index], true
}
```

이 함수에서 0은 유효한 원소일 수도 있습니다. 04장처럼 두 번째 반환값으로 성공 여부를 구분합니다. 인덱스 검사는 `len`을 기준으로 하며, `cap`으로 바꾸면 안 됩니다.

## 3. 슬라이스: 원소 범위를 다루는 값

`[]int`에는 길이 숫자가 없습니다. 슬라이스 값은 원소가 있는 저장 공간과 그 안에서 다루는 범위를 나타냅니다. 슬라이스의 값에는 길이와 용량 정보도 포함됩니다.

“가변 길이 배열”이라는 표현만 기억하면 대입과 부분 선택에서 복사를 기대하기 쉽습니다. 이 장에서는 **원소 저장 공간**과 **그 공간을 보는 슬라이스 값**을 나눠 생각합니다. [슬라이스 타입](https://go.dev/ref/spec#Slice_types)의 정의에 근거한 구분입니다.

### 3.1 nil, 빈 슬라이스, make

저장 경로: [`examples/05-arrays-slices/slice-basics/main.go`](../../examples/05-arrays-slices/slice-basics/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	var pending []int
	empty := []int{}
	values := make([]int, 3, 5)
	values[1] = 7
	fmt.Printf("nil: %t, 길이: %d\n", pending == nil, len(pending))
	fmt.Printf("empty: %t, 길이: %d\n", empty == nil, len(empty))
	fmt.Printf("원소: %v, 길이: %d, 용량: %d\n", values, len(values), cap(values))
	values[2] = 9
	fmt.Println("변경:", values)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/05-arrays-slices/slice-basics
```

예상 출력:

```text
nil: true, 길이: 0
empty: false, 길이: 0
원소: [0 7 0], 길이: 3, 용량: 5
변경: [0 7 9]
```

`var pending []int`는 슬라이스의 영값인 `nil`입니다. `[]int{}`는 원소가 없는 슬라이스 리터럴이며 nil은 아닙니다. 둘 다 길이가 0이므로 원소 순회는 0회 실행합니다.

`make([]int, 3, 5)`는 길이 3, 용량 5인 슬라이스를 만듭니다. 현재 원소 세 개는 `int`의 영값인 0으로 시작합니다. 용량이 5여도 `values[3]`을 곧바로 읽거나 쓸 수는 없습니다.

| 값 | nil 여부 | 길이 | 현재 원소 접근 |
| --- | --- | --- | --- |
| `var s []int` | 참 | 0 | 불가능 |
| `[]int{}` | 거짓 | 0 | 불가능 |
| `make([]int, 3)` | 거짓 | 3 | 0~2 |
| `make([]int, 0, 5)` | 거짓 | 0 | 불가능 |

`fmt.Println`으로 nil과 빈 슬라이스를 출력하면 둘 다 `[]`처럼 보입니다. 구별이 필요하면 `s == nil`을 확인합니다. 원소가 없는지만 확인하는 계약에는 `len(s) == 0`이 더 알맞습니다.


### 3.2 길이와 용량은 다른 질문이다

- **길이(length)**: 지금 슬라이스에 포함된 원소 수입니다. `len(s)`로 읽습니다.
- **용량(capacity)**: 현재 시작 위치에서 사용할 수 있는 내부 배열 범위의 크기입니다. `cap(s)`로 읽습니다.

항상 `0 <= len(s) <= cap(s)`입니다. 다만 용량 안에 공간이 남았다는 것과 지금 그 위치를 인덱싱할 수 있다는 것은 다릅니다.

슬라이스에서 `make`의 세 번째 인자를 생략하면 용량은 길이와 같습니다. `make([]int, 4)`는 길이 4, 용량 4입니다. 길이 0부터 `append`할 계획이면 `make([]int, 0, 4)`처럼 적습니다. [슬라이스 생성](https://go.dev/ref/spec#Making_slices_maps_and_channels)과 [길이·용량](https://go.dev/ref/spec#Length_and_capacity)을 참고합니다.

```go
values := make([]int, 3)
values = append(values, 10, 20, 30)
fmt.Println(values)
```

위 **main 안의 발췌**는 `[0 0 0 10 20 30]`을 출력합니다. 이미 존재하던 원소 3개 뒤에 세 개를 추가했기 때문입니다. 처음부터 세 값을 채우려면 인덱스로 대입하거나, 길이 0에서 추가해야 합니다.

성능을 이유로 큰 용량부터 지정할 필요는 없습니다. 예상 개수가 분명할 때 예약할 수 있지만, 원소 개수와 메모리 사용량은 실제 사용 맥락에서 판단합니다. 용량을 지정하는 코드는 데이터가 몇 개인지 나타내는 코드와 구분합니다.

## 4. 부분 슬라이스와 공유 저장 공간

`s[low:high]`는 `low`부터 `high` 바로 전까지를 선택합니다. 끝 위치는 포함하지 않습니다. 따라서 결과 길이는 `high-low`입니다.

| 표현 | 길이 5인 `s`에서 선택하는 위치 |
| --- | --- |
| `s[1:4]` | 1, 2, 3 |
| `s[:2]` | 0, 1 |
| `s[2:]` | 2, 3, 4 |
| `s[:]` | 전체 범위 |
| `s[2:2]` | 빈 범위 |

아래 예제에서 부분 슬라이스를 바꾸기 전에 원본이 어떻게 될지 예상해 보세요.

### 4.1 부분 선택은 원소 복사가 아니다

저장 경로: [`examples/05-arrays-slices/slice-view/main.go`](../../examples/05-arrays-slices/slice-view/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	base := []int{10, 20, 30, 40, 50}
	middle := base[1:4]
	middle[0] = 99
	fmt.Println("원본:", base)
	fmt.Printf("부분: %v, 길이: %d, 용량: %d\n", middle, len(middle), cap(middle))
	expanded := middle[:4]
	fmt.Println("범위 확장:", expanded)
	limited := base[1:3:3]
	fmt.Printf("용량 제한: %v, 길이: %d, 용량: %d\n", limited, len(limited), cap(limited))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/05-arrays-slices/slice-view
```

예상 출력:

```text
원본: [10 99 30 40 50]
부분: [99 30 40], 길이: 3, 용량: 4
범위 확장: [99 30 40 50]
용량 제한: [99 30], 길이: 2, 용량: 2
```

`middle[0]`과 `base[1]`은 같은 원소 저장 위치를 가리킵니다. 부분 슬라이스의 인덱스는 새 시작점부터 다시 0으로 셉니다.

| `middle`의 위치 | `base`의 위치 | 변경 후 값 |
| --- | --- | --- |
| 0 | 1 | 99 |
| 1 | 2 | 30 |
| 2 | 3 | 40 |

`middle`의 용량은 4입니다. `base`의 위치 1부터 마지막 위치 4까지 사용할 수 있기 때문입니다. `middle[:4]`는 그 남은 공간까지 포함하는 새 슬라이스 값을 만듭니다.

`base[1:3:3]`은 **전체 슬라이스 표현식(full slice expression)**입니다. 마지막 숫자로 결과 용량을 제한합니다. 길이는 `3-1=2`, 용량도 `3-1=2`입니다. 현재 원소는 여전히 공유하며, 이 표기만으로 복사하지는 않습니다. [슬라이스 표현식](https://go.dev/ref/spec#Slice_expressions)을 참고합니다.


### 4.2 범위를 계산할 때의 기준

슬라이스를 대상으로 한 보통의 `s[low:high]`는 `high`를 용량까지 지정할 수 있습니다. 따라서 길이 3, 용량 4인 `middle`에서 `middle[:4]`는 가능하지만 `middle[3]`은 현재 길이 밖이라 불가능합니다.

| 표현 | 검사 기준 | 예제에서 결과 |
| --- | --- | --- |
| `middle[2]` | `2 < len(middle)` | 가능 |
| `middle[3]` | `3 < len(middle)` | 불가능 |
| `middle[:4]` | `4 <= cap(middle)` | 가능 |
| `middle[:5]` | `5 <= cap(middle)` | 불가능 |

경계를 생략할 때 기본 끝은 길이입니다. `s[:]`는 현재 전체 범위이며, 남은 용량까지 자동으로 늘리는 표현이 아닙니다.

전체 슬라이스 표현식 `s[low:high:max]`에서는 `0 <= low <= high <= max <= cap(s)`를 만족해야 합니다. 결과의 길이는 `high-low`, 용량은 `max-low`입니다. 이 문법은 이후의 추가가 기존 뒷부분을 덮어쓸 수 있는 범위를 줄이는 데 사용할 수 있습니다.

## 5. append: 결과 슬라이스를 다시 받기

`append(s, value)`는 원소를 추가한 **슬라이스 값**을 반환합니다. 보통 `s = append(s, value)`처럼 결과를 대입합니다. `append`가 원래 변수의 길이를 직접 바꿔 주는 명령이라고 생각하면 함수 전달에서 혼동하기 쉽습니다.

| 호출 | 추가하는 내용 |
| --- | --- |
| `append(s, 10)` | 값 하나 |
| `append(s, 10, 20)` | 값 두 개 |
| `append(s, other...)` | 다른 슬라이스의 원소들 |

`other...`는 04장에서 만난 가변 인자 전달 문법입니다. `[]int` 슬라이스를 `[]int`에 이어 붙일 때 원소 타입을 맞추어 사용합니다.

### 5.1 공간 재사용과 새 공간 할당

저장 경로: [`examples/05-arrays-slices/append-growth/main.go`](../../examples/05-arrays-slices/append-growth/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	values := make([]int, 2, 4)
	values[0], values[1] = 10, 20
	before := values
	values = append(values, 30, 40)
	fmt.Println("추가 후:", values)
	fmt.Println("기존 뷰 확장:", before[:4])
	values = append(values, 50)
	values[0] = 99
	fmt.Println("새 저장 공간:", values)
	fmt.Println("기존 값:", before)
	fmt.Println("필요한 용량 확보:", cap(values) >= len(values))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/05-arrays-slices/append-growth
```

예상 출력:

```text
추가 후: [10 20 30 40]
기존 뷰 확장: [10 20 30 40]
새 저장 공간: [99 20 30 40 50]
기존 값: [10 20]
필요한 용량 확보: true
```

처음 `values`는 길이 2, 용량 4입니다. 30과 40을 추가한 뒤 길이는 4가 되며, 기존 저장 공간이 충분하므로 그 공간을 사용합니다. `before`의 길이는 2로 남지만 `before[:4]`로 확장하면 추가된 원소를 볼 수 있습니다.

다음에 50을 더하면 필요한 길이가 5이므로 기존 용량 4를 넘습니다. 새 저장 공간에 기존 원소와 새 원소를 담습니다. 이후 `values[0] = 99`는 이전 공간을 사용하는 `before[0]`을 바꾸지 않습니다.

| 단계 | `len(values)` | 원소 저장 공간 | `len(before)` |
| --- | --- | --- | --- |
| 생성 | 2 | 처음 만든 공간 | 아직 없음 |
| `before := values` | 2 | 공유 | 2 |
| 두 원소 추가 | 4 | 기존 공간 재사용 | 2 |
| 50 추가 | 5 | 새 공간 필요 | 2 |

새 용량이 정확히 얼마로 증가하는지는 이 예제의 계약이 아닙니다. “항상 두 배”처럼 외우지 않습니다. 명세가 보장하는 조건은 필요한 원소를 담을 충분한 용량과 기존 값의 보존입니다. [append와 copy](https://go.dev/ref/spec#Appending_and_copying_slices)를 참고합니다.


### 5.2 용량 때문에 원본의 뒷부분이 바뀌는 경우

다음은 **main 안의 독립된 발췌**입니다.

```go
base := []int{10, 20, 30, 40}
first := base[:2]
first = append(first, 99)
fmt.Println(first)
fmt.Println(base)
```

예상 출력:

```text
[10 20 99]
[10 20 99 40]
```

`first`의 길이는 2였지만 용량은 4였습니다. 추가된 99가 공유 배열의 세 번째 위치에 기록됩니다. `append`는 다른 슬라이스가 그 위치를 사용 중인지 판단해서 복사해 주지 않습니다.

`first := base[:2:2]`로 바꾸어 같은 실험을 하면 `base`는 `[10 20 30 40]`을 유지합니다. 추가할 공간이 제한되어 새 공간을 사용하기 때문입니다. 그러나 **추가하기 전** `first[0] = 7` 같은 변경은 여전히 원본에 보입니다.

따라서 용량 제한은 모든 변경을 차단하는 독립 복사와 다릅니다. 원본을 보존해야 하는 함수라면 새 공간에 원소를 복사하는 설계를 검토합니다.

## 6. copy: 독립된 원소 저장 공간 만들기

`other := values`는 슬라이스 값만 복사합니다. 두 변수의 길이 정보는 별도이지만 원소는 공유할 수 있습니다. 정수 원소까지 독립적으로 보관하려면 목적지 공간을 만들고 `copy`로 값을 옮깁니다.

`copy(dst, src)`의 반환값은 실제 복사한 원소 수입니다. 두 길이 중 작은 값만큼 복사하며 목적지의 길이를 늘리지 않습니다.

### 6.1 복사 수, 독립성, 겹치는 범위

저장 경로: [`examples/05-arrays-slices/copy/main.go`](../../examples/05-arrays-slices/copy/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func cloneInts(values []int) []int {
	if values == nil {
		return nil
	}
	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}
func main() {
	source := []int{1, 2, 3}
	cloned := cloneInts(source)
	cloned[0] = 99
	fmt.Println("원본:", source)
	fmt.Println("독립 복사:", cloned)
	short := make([]int, 2)
	count := copy(short, source)
	fmt.Printf("짧은 목적지: %v, 복사 수: %d\n", short, count)
	shifted := []int{1, 2, 3, 4}
	copy(shifted[1:], shifted[:3])
	fmt.Println("겹치는 복사:", shifted)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/05-arrays-slices/copy
```

예상 출력:

```text
원본: [1 2 3]
독립 복사: [99 2 3]
짧은 목적지: [1 2], 복사 수: 2
겹치는 복사: [1 1 2 3]
```

`cloneInts`는 원본과 같은 길이의 목적지를 만든 뒤 원소를 복사합니다. 결과의 첫 원소를 99로 바꾸어도 원본은 유지됩니다. 이 함수는 nil 입력은 nil로, nil이 아닌 빈 입력은 빈 슬라이스로 반환하도록 계약을 정했습니다.

목적지 `short`의 길이는 2이므로 원본이 세 개여도 두 개만 복사합니다. 반환된 `count`가 2인 이유입니다. 남은 원소를 자동으로 추가하는 동작은 없습니다.

마지막 `copy`에서는 출발 범위와 목적 범위가 겹칩니다. Go의 `copy`는 겹치는 원소 범위도 올바르게 처리하므로 `[1 1 2 3]`이 됩니다. 직접 앞에서부터 대입하는 반복문과는 결과가 달라질 수 있습니다.

| 출발 범위 | 목적 범위 | 복사 후 |
| --- | --- | --- |
| `[1 2 3]` | `short`의 두 자리 | `[1 2]` |
| 원래 위치 0~2 | 위치 1~3 | `[1 1 2 3]` |

명세의 [슬라이스 추가·복사](https://go.dev/ref/spec#Appending_and_copying_slices)는 겹치는 경우에도 동작이 적용된다고 설명합니다.


### 6.2 용량만 있는 목적지는 복사할 길이가 없다

아래는 컴파일되지만 의도와 다른 **main 안의 발췌**입니다.

```go
source := []int{10, 20, 30}
target := make([]int, 0, len(source))
count := copy(target, source)
fmt.Println(target, count)
```

출력은 `[] 0`입니다. `cap(target)`은 3이지만 `len(target)`은 0입니다. `copy`는 현재 목적지 원소 범위에만 씁니다.

다음 중 목적에 맞게 고칩니다.

| 방법 | 수정 |
| --- | --- |
| 원본 전체를 복사할 길이로 생성 | `target := make([]int, len(source))` |
| 준비한 공간을 현재 범위로 확장 | `target = target[:len(source)]` 뒤 `copy` |
| 길이 0에서 원소 추가 | `target = append(target, source...)` |

세 방법 모두 입력 조건을 맞추어야 합니다. 두 번째 방법은 목적지 용량이 충분하다는 전제가 필요합니다. 이번 예제에서는 처음부터 길이를 맞추는 첫 방법이 가장 직접적입니다.

`copy`는 원소 값을 복사합니다. 원소가 또 다른 슬라이스인 `[][]int`라면 바깥 원소의 슬라이스 값만 복사되어 안쪽 데이터는 공유할 수 있습니다. 12절에서 이 차이를 확인합니다.

## 7. 슬라이스를 함수에 전달하기

04장에서 인자 값은 매개변수로 복사된다고 배웠습니다. 슬라이스에서도 규칙은 같습니다. 다만 복사된 슬라이스 값이 원소 저장 공간을 공유합니다.

따라서 원소를 바꾸는 일과 슬라이스의 길이를 바꾸어 돌려주는 일을 구분해야 합니다.

### 7.1 원소 변경과 길이 변경을 추적하기

저장 경로: [`examples/05-arrays-slices/slice-functions/main.go`](../../examples/05-arrays-slices/slice-functions/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func increment(values []int) {
	for i := range values {
		values[i]++
	}
}
func addValue(values []int, value int) []int {
	return append(values, value)
}
func main() {
	values := make([]int, 2, 4)
	values[0], values[1] = 10, 20
	increment(values)
	fmt.Println("원소 변경:", values)
	addValue(values, 30)
	fmt.Println("결과를 버린 길이:", len(values))
	values = addValue(values, 40)
	fmt.Println("반환값 대입:", values)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/05-arrays-slices/slice-functions
```

예상 출력:

```text
원소 변경: [11 21]
결과를 버린 길이: 2
반환값 대입: [11 21 40]
```

`increment`는 공유하는 원소에 접근해서 값을 바꿉니다. `main`에서도 11과 21이 보입니다. 이 함수는 원소 변경 자체가 책임이며 반환값은 없습니다.

`addValue`는 `append` 결과를 반환합니다. 첫 호출은 결과를 버렸으므로 `main`의 슬라이스 길이는 여전히 2입니다. 뒤에 실제로 기록된 30은 현재 길이 밖의 공간에 있습니다.

두 번째 호출의 결과를 `values`에 대입하면 길이가 3인 슬라이스를 갖게 됩니다. 같은 위치에 새 값 40을 추가했으므로 결과는 `[11 21 40]`입니다.

| 실행 | 호출자의 길이 | 현재 보이는 원소 |
| --- | --- | --- |
| 처음 | 2 | `[10 20]` |
| `increment(values)` | 2 | `[11 21]` |
| `addValue(values, 30)` 결과 버림 | 2 | `[11 21]` |
| `values = addValue(values, 40)` | 3 | `[11 21 40]` |

함수가 원소를 바꾸는지, 새 결과를 반환하는지, 두 일을 모두 하는지 계약에 적으면 호출하는 쪽이 올바르게 사용할 수 있습니다. 값 전달의 원칙은 [Go FAQ](https://go.dev/doc/faq#pass_by_value)를 참고합니다.


### 7.2 전달받은 이름에 재대입해도 호출자의 이름은 바뀌지 않는다

다음은 **함수 선언 발췌**입니다.

```go
func reset(values []int) {
	values = nil
}
```

이 함수의 지역 매개변수에 nil을 대입해도 호출자의 슬라이스 변수는 nil이 되지 않습니다. 원소 접근 `values[i] = ...`와 슬라이스 변수 대입 `values = ...`는 다른 위치를 변경합니다.

호출자의 슬라이스를 비우는 것이 목적이라면 호출한 쪽에서 `values = nil` 또는 `values = values[:0]`처럼 명시적으로 처리할 수 있습니다. 두 방식의 저장 공간 유지 여부는 다릅니다.

| 코드 | 길이 | 원래 저장 공간과의 관계 |
| --- | --- | --- |
| `values = nil` | 0 | 그 변수의 연결을 끊음 |
| `values = values[:0]` | 0 | 용량과 시작 위치를 유지 |

다른 슬라이스가 같은 공간을 계속 사용하는지는 별도로 확인해야 합니다. 한 변수에 nil을 대입했다고 다른 별칭까지 사라지거나 원소가 0으로 지워지는 것은 아닙니다.

## 8. range의 원소는 값으로 받는다

슬라이스의 `range`는 인덱스와 해당 원소 값을 제공합니다. 두 번째 변수는 그 반복에서 읽은 값의 복사본입니다.

다음 전체 코드를 개인 실습 파일에서 실행합니다.

```go
package main

import "fmt"

func main() {
	values := []int{10, 20, 30}
	for _, value := range values {
		value++
	}
	fmt.Println("값 변수만 변경:", values)
	for i := range values {
		values[i]++
	}
	fmt.Println("원소에 대입:", values)
}
```

```sh
go run ./_practice/05-arrays-slices/lab
```

예상 출력:

```text
값 변수만 변경: [10 20 30]
원소에 대입: [11 21 31]
```

첫 반복의 `value++`는 지역 반복 변수를 바꿉니다. 두 번째 반복은 인덱스로 원래 원소에 접근합니다. [range 절](https://go.dev/ref/spec#For_range)에 근거한 차이입니다.

읽기만 할 때는 `for _, value := range values`, 원소를 수정할 때는 `for i := range values`가 의도를 드러냅니다. 인덱스와 원소가 모두 필요하면 두 변수에 이름을 붙입니다.

### 8.1 순회 중 append의 범위

다음은 **main 안의 독립된 발췌**입니다.

```go
values := []int{1, 2}
for _, value := range values {
	values = append(values, value+10)
}
fmt.Println(values)
```

출력은 `[1 2 11 12]`입니다. 반복을 시작할 때 평가한 슬라이스의 길이를 따라 순회하므로 새로 추가한 11과 12까지 계속 방문하지는 않습니다.

이것이 원소 전체의 별도 복사본을 만든다는 뜻은 아닙니다. 반복 중 공유 원소를 변경하면 뒤에서 읽는 값에 영향을 줄 수 있습니다. 길이·대입·추가를 동시에 바꾸는 코드는 처음에는 피하고, 입력을 순회하며 별도 결과에 추가하는 방식부터 사용합니다.

## 9. 입력을 보존하며 조건에 맞는 값 고르기

필터(filter)는 여러 원소 중 조건에 맞는 값만 선택하는 처리입니다. 원래 순서를 유지하고 원본을 보존할지 여부를 함수의 계약으로 정합니다.

### 9.1 새 결과에만 append하기

저장 경로: [`examples/05-arrays-slices/filter/main.go`](../../examples/05-arrays-slices/filter/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func atLeast(values []int, minimum int) []int {
	var selected []int
	for _, value := range values {
		if value >= minimum {
			selected = append(selected, value)
		}
	}
	return selected
}
func main() {
	minutes := []int{15, 40, 0, 25, 60}
	selected := atLeast(minutes, 30)
	fmt.Println("전체:", minutes)
	fmt.Println("30분 이상:", selected)
	selected[0] = 99
	fmt.Println("결과 수정 뒤 원본:", minutes)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/05-arrays-slices/filter
```

예상 출력:

```text
전체: [15 40 0 25 60]
30분 이상: [40 60]
결과 수정 뒤 원본: [15 40 0 25 60]
```

`selected`는 nil 슬라이스에서 시작합니다. 조건을 통과하는 정수를 새 결과에 추가하므로 원본 원소 저장 공간을 사용하지 않습니다. 같은 값을 여러 번 만나면 각각 추가하고, 입력 순서를 유지합니다.

| 현재 값 | 30 이상인가? | 누적 결과 |
| --- | --- | --- |
| 15 | 아니오 | `[]` |
| 40 | 예 | `[40]` |
| 0 | 아니오 | `[40]` |
| 25 | 아니오 | `[40]` |
| 60 | 예 | `[40 60]` |

결과의 첫 원소를 99로 바꾸어도 원본은 유지됩니다. 이 확인이 단순히 `[40 60]`이라는 출력만 비교하는 것보다 더 많은 계약을 검사합니다.

이 함수는 조건을 통과한 원소가 없으면 nil을 반환합니다. 호출한 쪽이 원소 유무만 필요하다면 `len(result) == 0`으로 확인합니다. 빈 결과에서 `result[0]`을 읽으면 범위 오류가 나므로 먼저 길이를 검사합니다.


### 9.2 함수 계약을 표로 적기

| 질문 | `atLeast`의 약속 |
| --- | --- |
| 입력 순서를 유지하는가? | 유지 |
| 같은 값이 여러 번 나오면? | 모두 유지 |
| 원본 원소를 변경하는가? | 변경하지 않음 |
| 결과 원소를 바꾸면 원본도 바뀌는가? | 이 정수 결과는 독립적 |
| 통과한 값이 없으면? | 길이 0인 nil 슬라이스 |
| 음수 기준을 받을 수 있는가? | 받을 수 있음 |

`atLeast([]int{-2, 0, 3}, 0)`은 `[0 3]`, 기준 4는 빈 결과입니다. 기준과 같은 값도 `>=` 때문에 포함됩니다.

입력의 공간을 재사용하는 필터도 만들 수 있지만 원본 변경 계약이 달라집니다. 이번 장의 기본 구현은 결과 독립성을 먼저 택합니다. 별도 저장 공간이 불필요한지 여부는 사용처와 측정 결과로 판단하며, 작은 예제만 보고 성능 우열을 단정하지 않습니다.

## 10. 비우기, 비교하기, 유지해야 할 데이터

### 10.1 같은 슬라이스인지 어떻게 확인할까?

슬라이스끼리는 `==`로 원소를 비교할 수 없습니다. `nil`과의 비교는 가능합니다. 정수 목록의 값이 같은지 직접 확인하려면 길이를 비교한 뒤 각 원소를 검사합니다.

다음은 **함수 선언 발췌**입니다.

```go
func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, value := range a {
		if value != b[i] {
			return false
		}
	}
	return true
}
```

이 함수는 nil과 빈 슬라이스를 같은 원소열로 취급합니다. 반면 저장 공간의 동일성이나 nil 여부까지 검사하는 함수는 아닙니다. 비교의 목적을 먼저 정해야 합니다. [비교 연산자](https://go.dev/ref/spec#Comparison_operators)를 참고합니다.

표준 라이브러리 `slices.Equal`도 원소열 비교에 사용할 수 있습니다. 이 장의 테스트는 이 함수를 보조 도구로 사용합니다. 타입 매개변수 문법은 13장에서 배우며, 기본 풀이에서는 직접 반복하는 방법만으로 충분합니다. [slices.Equal](https://pkg.go.dev/slices#Equal)의 계약을 참고합니다.

### 10.2 clear는 길이를 줄이지 않는다

Go 1.21부터 사용할 수 있는 내장 함수 `clear`는 슬라이스의 현재 원소를 영값으로 바꿉니다. 프로젝트의 Go 1.27에서 사용할 수 있습니다.

```go
values := []int{10, 20, 30}
clear(values)
fmt.Println(values, len(values))
```

위 **main 안의 발췌**는 `[0 0 0] 3`을 출력합니다. `clear(values)`와 `values = values[:0]`는 다릅니다. 앞은 원소를 지우고 길이를 유지하며, 뒤는 길이를 줄이지만 공유 저장 공간의 값을 그대로 둘 수 있습니다. [clear](https://go.dev/ref/spec#Clear)를 참고합니다.

원본이 큰 슬라이스의 일부만 오래 보관하면 그 부분이 연결된 저장 공간도 계속 필요할 수 있습니다. 작은 일부를 독립적으로 남겨야 한다면 필요한 원소만 새 슬라이스에 복사합니다. 정확한 메모리 비용은 값 타입과 실행 상황에 따라 다르며, 이 장에서는 데이터 공유 관계를 먼저 확인합니다.

## 11. 직접 작성하기: 학습 기록 요약

다음 실습에서는 일별 학습 시간 중 30분 이상인 날을 고르고 합계를 구합니다. 입력은 작은 0 이상의 정수이며 원본 기록을 보존합니다.

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("lab")) {
    $practiceDir = Join-Path "_practice/05-arrays-slices" $practiceName
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
for practice_name in lab; do
    practice_dir="_practice/05-arrays-slices/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

한 폴더에는 하나의 `main`만 둡니다. 새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice/`는 공통 `go test ./...`에 포함되지 않으므로 개인 폴더를 직접 지정해서 실행하세요. [공통 연습 안내](../../exercises/README.md)를 참고합니다.


### 11.1 책임이 드러나는 함수 만들기

다음 전체 프로그램을 `_practice/05-arrays-slices/lab/main.go`에 작성합니다. 이전 실습은 별도 폴더에 보관하세요.

```go
package main

import "fmt"

func selectMinutes(values []int, minimum int) []int {
	var result []int
	for _, value := range values {
		if value >= minimum {
			result = append(result, value)
		}
	}
	return result
}

func totalMinutes(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func main() {
	records := []int{20, 45, 0, 30, 60}
	selected := selectMinutes(records, 30)
	fmt.Println("기록:", records)
	fmt.Println("선택:", selected)
	fmt.Printf("선택한 날: %d일, 합계: %d분\n", len(selected), totalMinutes(selected))
}
```

포맷과 실행:

```sh
gofmt -w _practice/05-arrays-slices/lab/main.go
go run ./_practice/05-arrays-slices/lab
```

예상 출력:

```text
기록: [20 45 0 30 60]
선택: [45 30 60]
선택한 날: 3일, 합계: 135분
```

선택과 합계를 서로 다른 함수로 나누었으므로 원본 전체의 합계가 필요하면 `totalMinutes(records)`를 별도로 호출할 수 있습니다. 출력 문구는 `main`에서 바꾸면 됩니다.

### 11.2 경계 입력 확인

입력과 기준을 하나씩 바꾼 뒤 마지막에는 기본값으로 복구합니다. 합계가 `int` 범위를 넘지 않는 작은 기록만 사용합니다.

| 기록 | 기준 | 선택 결과 | 합계 |
| --- | --- | --- | --- |
| `[]int{}` | 30 | `[]` | 0 |
| `[]int{29}` | 30 | `[]` | 0 |
| `[]int{30}` | 30 | `[30]` | 30 |
| `[]int{30, 30}` | 30 | `[30 30]` | 60 |
| `[]int{0, 10}` | 0 | `[0 10]` | 10 |

`selected`가 비었을 때 첫 원소를 읽지 않으므로 정상적으로 합계 0을 출력합니다. 합계 함수는 입력 원소 수와 상관없이 같은 반복문을 사용합니다.

한 번 더 확인하려면 `selected`에 원소가 있을 때 첫 값을 바꾸고 `records`를 출력합니다. 원본이 유지되어야 합니다. 이 변경은 독립성 확인을 위한 실험이므로 끝나면 제거합니다.

## 12. 선택 학습: 슬라이스 안의 슬라이스

`[][]int`는 원소 타입이 `[]int`인 슬라이스입니다. 바깥 슬라이스의 각 원소가 또 다른 슬라이스이므로 행마다 길이가 달라도 됩니다. 고정 직사각형 모양이라고 가정하지 않습니다.

다음 전체 프로그램은 개인 파일에서 따로 실행합니다.

```go
package main

import "fmt"

func main() {
	rows := [][]int{{10, 20}, {30}, {}}
	outer := make([][]int, len(rows))
	copy(outer, rows)
	outer[0][0] = 99
	fmt.Println("원본 행:", rows)

	independent := make([][]int, len(rows))
	for i, row := range rows {
		independent[i] = make([]int, len(row))
		copy(independent[i], row)
	}
	independent[0][0] = 7
	fmt.Println("독립 행:", independent)
	fmt.Println("원본 유지:", rows)
}
```

```sh
go run ./_practice/05-arrays-slices/lab
```

예상 출력:

```text
원본 행: [[99 20] [30] []]
독립 행: [[7 20] [30] []]
원본 유지: [[99 20] [30] []]
```

첫 `copy`는 바깥 원소인 슬라이스 값만 복사했습니다. 안쪽 정수 저장 공간은 여전히 공유합니다. 두 번째 방식은 각 행의 정수 저장 공간까지 따로 만들고 복사합니다.

| 복사 단계 | 독립적인 부분 | 공유할 수 있는 부분 |
| --- | --- | --- |
| `other := rows` | 변수 이름 | 바깥 원소와 안쪽 원소 |
| 바깥 `copy` | 바깥 슬라이스 원소 자리 | 각 행의 정수 원소 |
| 각 행까지 `copy` | 바깥과 각 행의 정수 원소 | 이 예제에서는 없음 |

이 실습의 두 번째 방식은 빈 행을 nil이 아닌 빈 슬라이스로 만듭니다. nil 상태까지 보존하는 복사 계약이 필요하다면 행마다 nil 여부도 검사합니다. 독립성, 길이, nil 여부는 각각 확인해야 할 조건입니다.

## 13. 자주 하는 실수와 오류 읽기

| 증상 | 원인 | 확인할 점 |
| --- | --- | --- |
| 마지막 원소에서 `index out of range` | `i <= len(s)` 사용 | 인덱스는 길이 미만 |
| 빈 입력에서만 실패 | `s[0]`을 바로 읽음 | 길이 또는 성공 여부 먼저 확인 |
| 용량은 충분한데 인덱스 오류 | 현재 길이와 용량 혼동 | 인덱싱은 `len`, 재슬라이싱은 허용 범위 확인 |
| 앞에 0들이 추가됨 | 길이 있는 `make` 뒤에 append | 미리 존재하는 원소 수 확인 |
| 함수에서 추가했는데 호출자의 길이는 그대로 | 반환 슬라이스를 버림 | 결과를 받아 대입 |
| 부분 값을 바꿨는데 원본도 바뀜 | 부분 슬라이스가 저장 공간 공유 | 독립 복사가 필요한 계약인지 확인 |
| copy가 0개만 복사 | 목적지 길이가 0 | 목적지 길이를 확보 |
| range 변수 변경이 원본에 반영되지 않음 | 원소 값의 복사본을 수정 | 인덱스로 원소 대입 |
| `a == b` 컴파일 오류 | 슬라이스를 직접 비교 | 원소별 비교 또는 `slices.Equal` |
| copy 후 안쪽 목록이 함께 바뀜 | 바깥 원소만 복사 | 필요한 깊이까지 독립 공간 확보 |

### 13.1 실패 코드를 안전하게 시험하기

아래는 **의도적으로 컴파일에 실패하는 전체 코드**입니다. 개인 실습 파일에서만 실행하고 끝나면 복구합니다.

```go
package main

import "fmt"

func main() {
	a := []int{1, 2}
	b := []int{1, 2}
	fmt.Println(a == b)
}
```

```sh
go run ./_practice/05-arrays-slices/lab
```

오류에서 `slice can only be compared to nil`이라는 설명을 찾습니다. 두 목록의 내용이 같다는 사실은 `==`를 사용할 수 있다는 뜻이 아닙니다. 10절의 `equalInts` 함수를 `main` 바깥에 추가하고 `equalInts(a, b)`로 바꾸면 `true`가 출력됩니다.

`append`와 `copy`의 목적도 구분합니다. `append`는 결과의 길이를 늘릴 수 있지만, `copy`는 목적지의 현재 길이 안에서 원소만 바꿉니다. 오류를 고칠 때는 문법만 바꾸지 말고 원하는 결과의 길이와 공유 조건을 먼저 적습니다.

## 14. 연습문제와 검증

[05장 연습문제](../../exercises/05-arrays-slices/README.md)는 기본 6문제와 선택 A로 구성됩니다. 각 문제에는 입력 조건, 예상 출력, 경계 입력과 단계별 힌트가 있습니다.

| 문제 | 확인하는 계약 |
| --- | --- |
| 1. 부분 슬라이스의 출력 예측 | 인덱스 이동과 원소 공유 |
| 2. 통과한 점수만 복사 | 순서·중복 유지와 원본 보존 |
| 3. 0개 복사되는 코드 고치기 | 목적지 길이와 용량 |
| 4. 역순의 독립 복사 | 양 끝 위치와 입력 보존 |
| 5. 지정 위치 삭제 | 인덱스 검증·성공 여부·새 결과 |
| 6. 누적 합계 | 누적 상태와 같은 길이의 결과 |
| 선택 A. 묶음으로 나누기 | 이중 슬라이스와 독립 복사 |

스스로 작성한 뒤 [해설](../../solutions/05-arrays-slices/README.md)을 읽습니다. 참고 구현은 아래 명령으로 검사할 수 있습니다.

```sh
go test -count=1 ./examples/05-arrays-slices/... ./solutions/05-arrays-slices/...
```

테스트는 기본 출력과 빈 입력, 경계 위치, 원본 보존 및 결과 변경 후의 독립성을 확인합니다. 테스트에서 쓰는 구조체와 `testing`은 08·12장에서 배우며, 지금은 문제의 함수 계약을 먼저 이해하면 됩니다.

## 15. 요약과 자기 점검

- 배열의 길이는 타입의 일부이고, 배열 값의 대입은 배열 원소 값을 복사합니다.
- 슬라이스의 대입과 부분 선택은 원소 저장 공간을 공유할 수 있습니다.
- 인덱스는 길이 미만이며, 길이가 0이면 읽을 원소가 없습니다.
- `len`은 현재 범위, `cap`은 시작 위치에서 사용할 수 있는 내부 범위를 나타냅니다.
- `append`의 결과를 받아야 호출자의 슬라이스 길이도 갱신됩니다.
- `copy`는 두 길이 중 작은 수만큼 복사하고 목적지 길이를 늘리지 않습니다.
- 함수의 원소 변경과 슬라이스 변수 재대입은 서로 다른 변경입니다.
- 입력 보존이 계약이면 원본을 바꾸지 않는지만 아니라 결과를 바꾼 뒤도 확인합니다.

### 확인 질문

1. `[3]int`와 `[]int`는 무엇이 다른가요?
2. 길이 2, 용량 5인 슬라이스에서 `s[2]`를 읽을 수 있나요?
3. `s[1:4]`의 길이는 얼마이며 끝 위치 4도 포함하나요?
4. `other := s` 뒤 `other[0]`을 바꾸면 왜 `s`에서도 보일 수 있나요?
5. append 뒤의 새 용량이 반드시 두 배라고 할 수 있나요?
6. 길이 0, 용량 3인 목적지에 copy하면 몇 개가 복사되나요?
7. `for _, value := range s`의 `value++`는 원래 정수를 바꾸나요?
8. 함수 안의 `s = nil`은 호출자의 슬라이스 변수를 바꾸나요?
9. `s[:2:2]`는 원소 독립 복사와 같은가요?
10. `[][]int`의 바깥 슬라이스만 copy하면 각 행의 정수도 독립적인가요?

<details>
<summary>답과 설명</summary>

1. 배열은 길이 3이 타입에 포함됩니다. 슬라이스는 길이를 값으로 가지며 원소 저장 공간의 범위를 나타냅니다.
2. 읽을 수 없습니다. 현재 인덱스 범위는 0과 1입니다. 용량이 충분하면 먼저 길이를 늘린 새 범위를 만들 수 있습니다.
3. 3이며 위치 1, 2, 3을 포함합니다. 끝 경계는 포함하지 않습니다.
4. 슬라이스 값만 복사되어 원소 저장 공간은 공유하기 때문입니다.
5. 아닙니다. 필요한 원소를 담을 용량은 확보되지만 증가 배율을 프로그램 계약으로 가정하지 않습니다.
6. 0개입니다. copy는 용량이 아니라 두 슬라이스의 길이를 사용합니다.
7. 바꾸지 않습니다. 값을 받은 반복 변수만 바꿉니다. 원소 변경은 `s[i]`에 대입합니다.
8. 바꾸지 않습니다. 함수의 매개변수에 재대입한 것입니다.
9. 아닙니다. 용량만 제한하며 현재 원소 변경은 여전히 공유될 수 있습니다.
10. 아닙니다. 안쪽 슬라이스 값이 복사될 뿐, 각 행의 원소 저장 공간은 공유할 수 있습니다.

</details>

### 완료 체크리스트

- [ ] 배열과 슬라이스의 대입 결과를 실행했다.
- [ ] 길이·용량·부분 범위의 숫자를 계산했다.
- [ ] append의 공간 재사용과 새 공간 사용을 구분했다.
- [ ] copy의 목적지 길이와 반환된 복사 수를 확인했다.
- [ ] 함수 호출 뒤 원소·길이·원본 중 무엇이 바뀌는지 설명했다.
- [ ] 빈 입력, 첫·마지막 위치와 잘못된 인덱스를 검사했다.
- [ ] 기본 문제 1~6의 결과와 원본 보존을 확인했다.

## 참고 자료

- [Go 명세: 배열](https://go.dev/ref/spec#Array_types)
- [Go 명세: 슬라이스](https://go.dev/ref/spec#Slice_types)
- [Go 명세: 인덱스](https://go.dev/ref/spec#Index_expressions)
- [Go 명세: 슬라이스 표현식](https://go.dev/ref/spec#Slice_expressions)
- [Go 명세: append와 copy](https://go.dev/ref/spec#Appending_and_copying_slices)
- [Go 명세: range](https://go.dev/ref/spec#For_range)
- [Go 블로그: Go Slices](https://go.dev/blog/slices-intro)
- [표준 라이브러리 slices](https://pkg.go.dev/slices)

확인일: 2026-11-09. 예제·경계 입력과 의도적인 실패 사례는 Go 1.27.3으로 확인했습니다. 기본 명령은 운영체제 공통이며 PowerShell 준비 명령은 별도로 제공합니다.

## 다음 학습

다음 06장에서는 정수 위치 대신 이름 같은 키로 값을 찾는 맵을 배웁니다. 슬라이스의 순서와 맵의 키를 함께 사용해 빈도와 그룹별 값을 집계합니다.

[이전: 04. 함수와 스코프](../04-functions-scope/README.md) · [전체 목차](../README.md) · [문제](../../exercises/05-arrays-slices/README.md) · [해설](../../solutions/05-arrays-slices/README.md)
