# 08장 해설: 포인터와 구조체

먼저 [문제 문서](../../exercises/08-pointers-structs/README.md)를 풀어 보세요. 이 파일은 출력, 계약의 해석, 참고 구현과 경계 검사를 함께 설명합니다. 함수 이름이 다른 풀이도 같은 계약을 만족하면 가능합니다.

모든 명령은 저장소 루트에서 실행합니다. 참고 코드는 별도 실행 프로그램이며, 개인 파일에 참고 구현과 `main`을 중복해서 붙이지 않습니다.

## 문제 1. 포인터와 구조체 값의 흐름

전체 프로그램:

```go
package main

import "fmt"

type Counter struct {
	Value int
}

func main() {
	original := Counter{Value: 10}
	copied := original
	pointer := &original
	copied.Value = 20
	pointer.Value += 3
	fmt.Println(original.Value, copied.Value)
	alias := pointer
	alias = &copied
	alias.Value = 30
	fmt.Println(original.Value, copied.Value)
	fmt.Println(pointer == alias)
}
```

예상 출력:

```text
13 20
13 30
false
```

| 시점 | 원본 값 | 복사본 값 | `pointer`의 대상 | `alias`의 대상 |
| --- | ---: | ---: | --- | --- |
| `copied.Value = 20` | 10 | 20 | original | 선언 전 |
| `pointer.Value += 3` | 13 | 20 | original | 선언 전 |
| `alias = &copied` | 13 | 20 | original | copied |
| `alias.Value = 30` | 13 | 30 | original | copied |

`copied`는 `Counter` 값의 복사본입니다. `pointer`와 처음의 `alias`는 같은 `original`을 가리키는 포인터 값입니다. 이후 `alias`에 다른 주소를 대입해도 별도 포인터 변수인 `pointer`는 바뀌지 않습니다.

마지막 비교는 같은 정수를 담았는지 비교하지 않습니다. 두 포인터가 같은 `Counter` 변수를 가리키는지 확인합니다. `Counter`는 int 필드가 있어 크기가 0인 구조체에 관한 특수한 비교 상황도 아닙니다.

## 문제 2. 포인터가 가리키는 값을 교환

두 포인터 중 하나라도 nil이면 어떤 값도 읽거나 쓰지 않습니다. 유효한 두 포인터를 받은 다음 기존 값을 잃지 않도록 교환합니다.

참고 구현: [swap/main.go](swap/main.go) · [테스트](swap/main_test.go).

```go
package main

import "fmt"

func swap(left, right *int) bool {
	if left == nil || right == nil {
		return false
	}
	*left, *right = *right, *left
	return true
}

func main() {
	left, right := 3, 8
	fmt.Println("교환:", swap(&left, &right))
	fmt.Println("값:", left, right)
	fmt.Println("nil:", swap(nil, &right))
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/08-pointers-structs/swap
go test ./solutions/08-pointers-structs/swap
```

예상 출력:

```text
교환: true
값: 8 3
nil: false
```


`*left, *right = *right, *left`는 대입할 오른쪽 값들을 먼저 구한 뒤 왼쪽에 저장합니다. `*left = *right`와 `*right = *left`를 순서대로 쓰면 두 번째 문장에서 이미 바뀐 값을 읽기 때문에 교환에 실패합니다.

임시 변수로 이전 `*left`를 보관하는 풀이도 맞습니다. 덧셈·뺄셈을 이용한 교환은 이 문제에 필요하지 않고 정수 오버플로 위험까지 만듭니다.

`left, right = right, left`라고 쓰면 포인터 매개변수의 지역 값만 바뀝니다. 호출자가 가진 정수 두 개는 그대로입니다. 문제의 계약은 포인터 자체의 교환이 아니라 가리키는 값의 교환입니다.

같은 포인터가 두 번 들어오면 오른쪽 두 값이 같고 같은 변수에 같은 값을 대입하므로 성공한 뒤 값이 유지됩니다. 테스트는 양쪽·한쪽 nil과 이 별칭 상황을 함께 확인합니다.

다중 대입의 평가 순서는 [Assignment statements](https://go.dev/ref/spec#Assignment_statements)를 참고합니다.

## 문제 3. 맵 항목을 다시 저장

수정한 전체 프로그램:

```go
package main

import "fmt"

type Stock struct {
	Quantity int
}

func main() {
	stock := map[string]Stock{"pen": {Quantity: 3}}
	item := stock["pen"]
	item.Quantity += 2
	stock["pen"] = item
	fmt.Println(stock["pen"].Quantity)
}
```

예상 출력:

```text
5
```

맵에서 꺼낸 구조체를 지역 변수 `item`에 저장합니다. 지역 변수의 필드는 수정할 수 있습니다. 마지막 `stock["pen"] = item`이 수정한 구조체 값을 맵에 반영합니다.

이 문제는 항목이 이미 있다고 정했으므로 조회 성공 여부를 생략했습니다. 사용자가 상품 이름을 입력하는 실제 기능이라면 `item, ok := stock[name]`으로 항목 존재 여부를 확인해야 합니다. 없는 항목의 제로 값을 기존 재고로 취급할지, 새로 등록할지는 별도 계약입니다.

`stock["pen"].Quantity`를 읽는 것과 그 필드에 직접 대입하는 것은 구분됩니다. 맵 항목 전체 대입은 언어가 허용하지만, 맵 내부 구조체 필드는 주소를 얻을 수 있는 변수처럼 다룰 수 없습니다.

시작 수량 `0`, 추가량 `1`에서도 같은 경로로 갱신하면 맵에서 `1`이 조회됩니다. 지역 변수만 출력하는 구현은 이 동작을 확인하지 못합니다.

## 문제 4. 학생 생성 함수의 성공·실패 값

UTF-8과 점수 범위를 먼저 검사합니다. 이름의 양끝 공백을 제거한 뒤 남는 내용이 있는지 확인합니다. 모든 검사를 통과한 경우에만 입력을 반영한 구조체를 반환합니다.

참고 구현: [new-student/main.go](new-student/main.go) · [테스트](new-student/main_test.go).

```go
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Student struct {
	Name  string
	Score int
}

func newStudent(name string, score int) (Student, bool) {
	if !utf8.ValidString(name) || score < 0 || score > 100 {
		return Student{}, false
	}
	cleaned := strings.TrimSpace(name)
	if cleaned == "" {
		return Student{}, false
	}
	return Student{Name: cleaned, Score: score}, true
}

func main() {
	student, ok := newStudent(" Mina ", 80)
	fmt.Printf("생성: %t, 이름: %q, 점수: %d\n", ok, student.Name, student.Score)
	invalid, ok := newStudent("Jin", -1)
	fmt.Printf("잘못된 점수: %t, 값: %+v\n", ok, invalid)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/08-pointers-structs/new-student
go test ./solutions/08-pointers-structs/new-student
```

예상 출력:

```text
생성: true, 이름: "Mina", 점수: 80
잘못된 점수: false, 값: {Name: Score:0}
```


`Student{}`의 필드는 `Name: ""`, `Score: 0`입니다. 그러나 점수 `0` 자체는 성공 가능한 값입니다. 성공 여부를 `student.Score == 0` 같은 방법으로 추측하면 안 됩니다. 별도로 반환한 `ok`가 판단 기준입니다.

이름이 `"\u3000한글\t"`이면 유니코드 공백과 탭을 제거하여 `"한글"`을 저장합니다. 이름 내부의 공백은 `TrimSpace`가 제거하지 않으므로 `"Mina Kim"`은 그대로입니다.

유효하지 않은 UTF-8을 문자열 변환으로 먼저 고치지 않습니다. 이 함수는 잘못된 입력을 거절하는 계약이며, 대체 문자로 손상 정보를 덮는 기능이 아닙니다. 정리·검증·정규화는 각각 다른 작업입니다.

이 생성 함수는 일반 함수입니다. `newStudent`라는 이름을 썼다고 Go가 직접 `Student{}`를 만드는 코드를 금지하지는 않습니다. 패키지 경계에서 상태를 보호하는 설계는 뒤에서 확장합니다.

테스트는 0·100의 포함 경계, -1·101의 제외 경계, 공백뿐인 이름, Unicode 공백과 무효 UTF-8을 확인합니다. 사용한 API는 [strings.TrimSpace](https://pkg.go.dev/strings#TrimSpace)와 [utf8.ValidString](https://pkg.go.dev/unicode/utf8#ValidString)입니다.

## 문제 5. 두 좌표의 일괄 갱신

이 함수의 핵심은 조건의 개수보다 대입 시점입니다. 입력 검사를 통과한 다음 새 좌표를 지역 변수에 계산하고, 결과 검사를 통과한 뒤 두 필드에 반영합니다.

참고 구현: [move-point/main.go](move-point/main.go) · [테스트](move-point/main_test.go).

```go
package main

import "fmt"

type Point struct{ X, Y int }

func movePoint(point *Point, dx, dy int) bool {
	if point == nil || point.X < -1000 || point.X > 1000 || point.Y < -1000 || point.Y > 1000 {
		return false
	}
	if dx < -1000 || dx > 1000 || dy < -1000 || dy > 1000 {
		return false
	}
	nextX, nextY := point.X+dx, point.Y+dy
	if nextX < -1000 || nextX > 1000 || nextY < -1000 || nextY > 1000 {
		return false
	}
	point.X, point.Y = nextX, nextY
	return true
}

func main() {
	point := Point{X: 10, Y: 20}
	ok := movePoint(&point, -3, 5)
	fmt.Printf("이동: %t, 좌표: (%d, %d)\n", ok, point.X, point.Y)
	ok = movePoint(&point, 1000, 0)
	fmt.Printf("범위 초과: %t, 좌표: (%d, %d)\n", ok, point.X, point.Y)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/08-pointers-structs/move-point
go test ./solutions/08-pointers-structs/move-point
```

예상 출력:

```text
이동: true, 좌표: (7, 25)
범위 초과: false, 좌표: (7, 25)
```


`point.X += dx`를 먼저 실행하고 `Y`만 나중에 검사하면 `(1, 2)`에 `(2, 1000)`을 요청했을 때 실패하면서 `X`가 `3`이 되는 버그가 생깁니다. 참고 구현은 `nextX`, `nextY`에 계산하므로 그 실패에서 원본은 `(1, 2)`입니다.

현재 값이 `1001`이고 이동량이 `-1`이라 최종 값이 `1000`이 되는 경우도 실패합니다. 문제는 현재 상태 자체가 -1000~1000이어야 한다고 정했습니다. 최종 결과만 검사하면 계약 일부를 놓칩니다.

허용 입력을 먼저 검사한 결과 덧셈의 범위는 -2000~2000입니다. 이 실습에서 검사 전 계산을 피한 이유를 설명할 수 있어야 합니다. 더 큰 숫자를 받는 범용 좌표 계산은 오버플로 정책을 따로 정해야 합니다.

| 실패 위치 | 실제 구조체에 대입했는가? | 원본 상태 |
| --- | --- | --- |
| nil 또는 현재 좌표 검사 | 아니요 | 보존 |
| 이동량 검사 | 아니요 | 보존 |
| 다음 좌표 검사 | 아니요 | 보존 |
| 모든 검사 성공 | 예 | 두 필드 갱신 |

테스트는 어느 한 축만 한도를 넘는 경우에도 두 필드가 모두 보존되는지 확인합니다. 정상 이동 하나만 확인하는 테스트로는 이 오류를 찾기 어렵습니다.

## 문제 6. 주문의 내부 슬라이스 복사

`result := order`로 ID와 슬라이스 값을 먼저 복사합니다. 그 뒤 nil이 아닌 `Items`를 새 슬라이스로 대체하고 원소들을 복사합니다.

참고 구현: [clone-order/main.go](clone-order/main.go) · [테스트](clone-order/main_test.go).

```go
package main

import "fmt"

type Item struct {
	Name     string
	Quantity int
}

type Order struct {
	ID    string
	Items []Item
}

func cloneOrder(order Order) Order {
	result := order
	if order.Items != nil {
		result.Items = make([]Item, len(order.Items))
		copy(result.Items, order.Items)
	}
	return result
}

func main() {
	original := Order{ID: "A-1", Items: []Item{{Name: "book", Quantity: 2}, {Name: "pen", Quantity: 3}}}
	cloned := cloneOrder(original)
	cloned.ID = "A-2"
	cloned.Items[0].Quantity = 9
	fmt.Printf("원본: %s, %s %d개\n", original.ID, original.Items[0].Name, original.Items[0].Quantity)
	fmt.Printf("복사본: %s, %s %d개\n", cloned.ID, cloned.Items[0].Name, cloned.Items[0].Quantity)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/08-pointers-structs/clone-order
go test ./solutions/08-pointers-structs/clone-order
```

예상 출력:

```text
원본: A-1, book 2개
복사본: A-2, book 9개
```


`copy`는 `Item` 값을 복사합니다. 현재 `Item`의 필드는 문자열과 정수이므로 새 원소의 필드에 대입해도 원본 원소는 바뀌지 않습니다. 문자열의 기존 바이트를 수정할 수 없다는 성질도 이 판단에 포함됩니다.

반면 `Item`에 `Options []string`을 추가한다면 `copy`만으로는 옵션의 기반 배열까지 분리되지 않습니다. 새로운 참조 성격의 필드가 추가될 때 복사 계약을 다시 검토해야 합니다. 범용 깊은 복사라고 이름 붙여 실제 보장 범위를 감추지 않습니다.

입력이 nil이면 처음 복사한 `result.Items`도 nil로 남습니다. nil이 아닌 길이 0 슬라이스에는 `make([]Item, 0)`을 사용하므로 결과 역시 nil이 아닌 빈 슬라이스입니다.

테스트는 내용과 순서 보존, 복사본을 바꿔도 원본 유지, 원본을 바꿔도 복사본 유지, nil·빈 슬라이스 구분을 검사합니다. 이 중 내용만 확인하면 기반 배열 공유 버그를 놓칠 수 있습니다.

## 선택 A. 원래 원소의 포인터 반환

이 선택 과제의 전체 코드는 개인 파일에서 실행합니다. 별도 참고 구현 디렉터리는 만들지 않았습니다.

```go
package main

import "fmt"

type Task struct {
	Title string
	Done  bool
}

func findTask(tasks []Task, title string) *Task {
	for i := range tasks {
		if tasks[i].Title == title {
			return &tasks[i]
		}
	}
	return nil
}

func main() {
	tasks := []Task{{Title: "읽기"}, {Title: "실행"}}
	found := findTask(tasks, "실행")
	if found != nil {
		found.Done = true
	}
	for _, task := range tasks {
		fmt.Printf("%s: %t\n", task.Title, task.Done)
	}
	fmt.Println("없는 항목:", findTask(tasks, "배포") == nil)
}
```

예상 출력:

```text
읽기: false
실행: true
없는 항목: true
```

`&tasks[i]`는 슬라이스 기반 배열의 실제 원소 주소입니다. `for _, task := range tasks`에서 `&task`를 반환하면 원소 값을 복사받은 range 변수의 주소이므로 호출자가 바꿔도 원소 자체는 갱신되지 않습니다.

Go 1.22 이후의 반복 변수 규칙은 각 반복 변수의 구분에 관한 것입니다. range 변수와 슬라이스 원소가 다른 변수라는 사실은 그대로입니다. 과거의 “모든 주소가 같은 변수다”라는 설명으로 현재 동작을 설명하지 않습니다.

빈·nil 슬라이스에서는 반복이 0번이라 nil을 반환합니다. 제목이 중복되면 첫 일치에서 반환하므로 첫 원소만 수정됩니다. 이미 `Done`이 true인 원소도 검색은 성공합니다.

이 포인터는 특정 원소를 가리킵니다. 이후 `append`가 새 배열을 만들면 이전 포인터가 새 배열로 자동 연결되지 않습니다. 문제에서 포인터 사용 중 슬라이스를 늘리지 않도록 제한한 이유입니다.

## 본문 자기 점검 해설

### 1. 포인터 재대입과 역참조 대입

`pointer = &other`는 포인터 변수의 대상을 바꿉니다. `*pointer = other`는 지금 가리키는 변수에 `other`의 값을 저장합니다. 왼쪽에 `*`가 있는지 먼저 읽습니다.

### 2. nil 포인터와 제로 값을 가리키는 포인터

`var pointer *int`는 nil이며 역참조할 대상이 없습니다. `pointer := new(int)`는 새 정수 변수의 포인터이고 nil이 아닙니다. 그 정수의 초기 값은 0이므로 읽고 수정할 수 있습니다.

### 3. 함수 안의 포인터 재대입

포인터 값도 매개변수로 복사됩니다. 매개변수에 다른 주소를 대입하는 것은 지역 변수 수정입니다. 호출자의 정수를 바꾸려면 그 정수를 가리키는 포인터를 역참조해야 합니다.

### 4. 생략한 구조체 필드

`Book{Title: "Go"}`에서 나머지 필드는 각 타입의 제로 값입니다. 예제의 `Price`는 0, `Available`은 false입니다. 업무상 기본값을 설정하는 기능이 자동 실행되지는 않습니다.

### 5. range 값 변수

`item`은 원소 값을 받은 변수입니다. 정수 필드 `Count`에 대입하면 그 복사본의 필드가 바뀝니다. 실제 원소를 수정하려면 `items[i].Count` 또는 그 원소 포인터를 사용합니다.

### 6. 맵 항목 재대입

맵에서 꺼낸 구조체를 지역 변수에서 수정한 것만으로 맵 값이 교체되지 않습니다. 수정한 구조체 전체를 `stock[key] = item`으로 다시 넣어야 변경이 반영됩니다.

### 7. 문자열 필드와 슬라이스 원소

`Name = "Jin"`은 복사본의 문자열 필드만 새 값으로 바꿉니다. 반면 얕게 복사된 `Scores[0] = 99`는 공유하는 기반 배열에 쓰므로 원본에서도 99가 보입니다.

### 8. 실패 시 원본 보존

입력 검사 → 다음 값 계산 → 결과 검사 → 실제 대입 순서입니다. 일부 필드를 먼저 바꾸었다가 실패하는 경로를 없앱니다. 이것은 순차 실행의 계약이며 동시 접근을 자동으로 안전하게 만들지는 않습니다.

## 확인 명령과 검증 범위

```sh
go test ./examples/08-pointers-structs/...
go test ./solutions/08-pointers-structs/...
go vet ./examples/08-pointers-structs/... ./solutions/08-pointers-structs/...
```

각 참고 구현의 `Example_main`은 문서의 예시 출력을 확인합니다. 별도 테스트는 nil, 입력 범위, 별칭, 첫 일치, 실패 시 상태 보존, 복사 독립성을 확인합니다. 위 명령은 참고 구현을 검사하며 개인 `_practice/` 풀이를 자동 검사하지 않습니다.

## 완료 체크리스트

- [ ] 포인터가 가리키는 변수와 포인터 변수 자체를 구분했다.
- [ ] nil 검사 전에 필드를 읽거나 역참조하지 않았다.
- [ ] 교환 함수에서 같은 포인터와 nil 입력을 확인했다.
- [ ] 맵 구조체를 수정한 뒤 항목 전체를 다시 대입했다.
- [ ] 생성 함수의 점수 0·100과 무효 이름을 확인했다.
- [ ] 좌표 이동 실패에서 두 필드 모두 보존했다.
- [ ] 주문 복사 후 양방향 변경이 독립인지 확인했다.
- [ ] nil 슬라이스와 nil이 아닌 빈 슬라이스의 계약을 유지했다.

## 참고 자료

공식 문서 확인일: **2026-11-09**.

- [Go 명세: Address operators](https://go.dev/ref/spec#Address_operators), [Assignment statements](https://go.dev/ref/spec#Assignment_statements)
- [Go 명세: Struct types](https://go.dev/ref/spec#Struct_types), [Representation of values](https://go.dev/ref/spec#Representation_of_values), [For range](https://go.dev/ref/spec#For_range)
- [strings.TrimSpace](https://pkg.go.dev/strings#TrimSpace), [utf8.ValidString](https://pkg.go.dev/unicode/utf8#ValidString)

[문제 다시 보기](../../exercises/08-pointers-structs/README.md) · [본문](../../book/08-pointers-structs/README.md) · [전체 목차](../../book/README.md)
