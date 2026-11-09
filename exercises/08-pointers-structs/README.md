# 08장 연습문제: 포인터와 구조체

[본문](../../book/08-pointers-structs/README.md)에서 포인터의 대상, 구조체 값 복사, 컬렉션 안의 구조체 갱신을 읽은 뒤 풉니다. 문제는 확인 → 변형 → 응용 순서입니다. 정답 코드는 이 파일에 두지 않습니다.

## 풀이 순서

1. 문제의 입력 범위와 실패 시 동작을 먼저 적습니다.
2. 원본을 바꿔야 하는지, 새 값을 반환해야 하는지 구분합니다.
3. 출력 예측 문제는 실행 전에 표로 추적합니다.
4. 개인 파일을 실행하고 정상 입력과 경계 입력을 비교합니다.
5. 막히면 아래 힌트를 하나씩 확인하고, 풀이 뒤 [해설](../../solutions/08-pointers-structs/README.md)을 읽습니다.

모든 명령은 저장소 루트에서 실행합니다. 기존 예제·참고 구현 파일은 수정하지 않습니다. 기본 문제는 앞 장에서 배운 문법과 표준 라이브러리로 해결할 수 있습니다. 테스트 파일 작성은 필수가 아닙니다.

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("predict", "swap", "repair", "new-student", "move-point", "clone-order", "find-task")) {
    $practiceDir = Join-Path "_practice/08-pointers-structs" $practiceName
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
for practice_name in predict swap repair new-student move-point clone-order find-task; do
    practice_dir="_practice/08-pointers-structs/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

한 폴더에는 하나의 `main`만 둡니다. 새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice/`는 공통 `go test ./...`에 포함되지 않으므로 개인 폴더를 직접 지정해서 실행하세요. [공통 연습 안내](../../exercises/README.md)를 참고합니다.


## 문제 1. 복사한 값과 같은 대상을 가리키는 포인터

**수준: 확인**

`_practice/08-pointers-structs/predict/main.go`의 전체 내용을 다음 프로그램으로 바꿉니다. 먼저 실행하지 않고 세 줄 출력을 적으세요.

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

다음 표를 직접 채웁니다.

| 시점 | `original.Value` | `copied.Value` | `pointer`의 대상 | `alias`의 대상 |
| --- | --- | --- | --- | --- |
| `copied.Value = 20` 직후 |  |  |  | 선언 전 |
| `pointer.Value += 3` 직후 |  |  |  | 선언 전 |
| `alias = &copied` 직후 |  |  |  |  |
| `alias.Value = 30` 직후 |  |  |  |  |

질문:

- `copied := original`과 `alias := pointer`는 각각 어떤 타입의 값을 복사하나요?
- `alias = &copied`가 `pointer`까지 바꾸지 않는 이유는 무엇인가요?
- 마지막 비교는 두 `Counter`의 내용과 대상 중 무엇을 확인하나요?

예측 뒤 실행합니다.

```sh
go run ./_practice/08-pointers-structs/predict
```

완료 조건: 세 줄 출력을 맞히고, 포인터 변수의 값과 그 대상의 값을 분리해 설명합니다.

## 문제 2. 두 정수 교환하기

**수준: 변형**

개인 파일: `_practice/08-pointers-structs/swap/main.go`.

`func swap(left, right *int) bool`을 구현합니다.

### 계약

- 두 포인터가 모두 있으면 가리키는 정수 값을 교환하고 `true`를 반환합니다.
- 하나라도 nil이면 `false`를 반환하고 다른 값도 바꾸지 않습니다.
- 같은 포인터를 두 번 전달하는 것은 성공입니다. 값은 그대로입니다.
- 음수·0·양수 모두 사용할 수 있습니다. 덧셈이나 뺄셈으로 교환하지 않습니다.

`main`에서 `left, right := 3, 8`로 시작합니다. 교환 성공 여부, 두 값, `swap(nil, &right)`의 결과를 차례로 출력합니다.

예상 출력:

```text
교환: true
값: 8 3
nil: false
```

경계 확인:

| 시작 값 | 호출 조건 | 성공 여부 | 끝 값 |
| --- | --- | --- | --- |
| -2, 7 | 서로 다른 두 포인터 | true | 7, -2 |
| 0, 0 | 서로 다른 두 포인터 | true | 0, 0 |
| 7 | 같은 포인터 두 번 | true | 7 |
| 7 | 왼쪽 nil | false | 7 보존 |
| 7 | 오른쪽 nil | false | 7 보존 |
| 없음 | 양쪽 nil | false | 접근하지 않음 |

```sh
gofmt -w _practice/08-pointers-structs/swap/main.go
go run ./_practice/08-pointers-structs/swap
```

완료 조건: 값을 덮어써 잃어버리지 않고, nil 검사 전에 역참조하지 않습니다. 매개변수 포인터 자체만 교환해 성공한 것처럼 반환하지 않습니다.

## 문제 3. 맵의 구조체 항목 갱신 오류 고치기

**수준: 변형**

`_practice/08-pointers-structs/repair/main.go`에 다음 전체 프로그램을 넣습니다. **이 코드는 의도적으로 컴파일되지 않습니다.** 일반 예제 경로에는 저장하지 않습니다.

```go
package main

import "fmt"

type Stock struct {
	Quantity int
}

func main() {
	stock := map[string]Stock{"pen": {Quantity: 3}}
	stock["pen"].Quantity += 2
	fmt.Println(stock["pen"].Quantity)
}
```

`Stock` 타입과 맵 값 타입을 유지하면서 수정합니다. 맵을 `map[string]*Stock`으로 바꾸는 방식은 이번 문제의 범위에 포함하지 않습니다.

조건:

- 이 프로그램에서는 `"pen"` 항목이 이미 있습니다.
- 지역 변수에서 구조체 값을 수정한 뒤 맵 항목 전체를 갱신합니다.
- 정답 출력을 상수로 적지 않습니다.

수정 후 예상 출력:

```text
5
```

```sh
go run ./_practice/08-pointers-structs/repair
```

추가 확인: 시작 수량을 `0`, 추가량을 `1`로 바꾸면 `1`이 출력되어야 합니다. 수정한 지역 변수만 출력하지 말고, 맵에서 다시 조회한 값을 출력하세요.

완료 조건: 컴파일 오류의 이유를 설명하고, 재대입을 빠뜨렸을 때 결과가 어떻게 달라지는지 설명합니다.

## 문제 4. 검사한 학생 값 생성하기

**수준: 응용**

개인 파일: `_practice/08-pointers-structs/new-student/main.go`.

`Student` 구조체는 `Name string`, `Score int` 두 필드를 갖습니다. `func newStudent(name string, score int) (Student, bool)`을 구현합니다.

### 계약

- 입력 이름은 유효한 UTF-8이어야 합니다.
- 이름의 양끝 유니코드 공백을 `strings.TrimSpace`로 제거합니다.
- 정리한 이름이 비어 있으면 실패합니다. 이름 내부의 공백은 그대로 둡니다.
- 점수는 양 끝을 포함하여 0~100입니다.
- 성공하면 정리된 이름과 점수를 가진 구조체, `true`를 반환합니다.
- 실패하면 `Student{}`, `false`를 반환합니다.

이 문제는 이름 길이나 특정 문자 종류를 제한하지 않습니다. 07장의 사용자 이름 문제와 다른 계약입니다. 대소문자 변환이나 유니코드 정규화도 하지 않습니다.

`main`에서는 `newStudent(" Mina ", 80)`과 `newStudent("Jin", -1)`을 호출합니다. `%q`와 `%+v`를 사용한 출력은 다음과 같습니다.

```text
생성: true, 이름: "Mina", 점수: 80
잘못된 점수: false, 값: {Name: Score:0}
```

필수 경계:

| 이름 | 점수 | 결과 |
| --- | ---: | --- |
| `"Go"` | 0 | 성공 |
| `"Go"` | 100 | 성공 |
| `"Go"` | -1 또는 101 | 제로 구조체와 false |
| `""` 또는 `" \t\n"` | 50 | 제로 구조체와 false |
| `"\u3000한글\t"` | 0 | 이름 `"한글"`, 성공 |
| `"Mina Kim"` | 80 | 내부 공백 보존, 성공 |
| `"\xff"` | 50 | 제로 구조체와 false |

```sh
gofmt -w _practice/08-pointers-structs/new-student/main.go
go run ./_practice/08-pointers-structs/new-student
```

완료 조건: 무효 입력이 부분적으로 채워진 구조체를 반환하지 않으며, 점수 0을 실패 신호와 혼동하지 않습니다. 반환한 `bool`로 성공 여부를 판단합니다.

## 문제 5. 실패하면 좌표를 그대로 두기

**수준: 응용**

개인 파일: `_practice/08-pointers-structs/move-point/main.go`.

`Point` 구조체는 `X int`, `Y int` 필드를 갖습니다. `func movePoint(point *Point, dx, dy int) bool`을 구현합니다.

### 계약

- nil 포인터는 실패합니다.
- 현재 `X`, `Y`와 이동량 `dx`, `dy`는 모두 -1000~1000이어야 합니다.
- 이동 결과의 `X`, `Y`도 모두 -1000~1000이어야 합니다.
- 성공하면 두 필드를 이동한 값으로 바꾸고 `true`를 반환합니다.
- 실패하면 두 필드를 모두 보존하고 `false`를 반환합니다.
- 이동량 0도 유효합니다.

입력 범위를 먼저 확인하면 덧셈 결과는 -2000~2000 안에 있으므로 이 문제에서는 플랫폼의 `int` 크기 차이에 따른 오버플로가 없습니다.

`main`에서 `(10, 20)`을 `(-3, 5)`만큼 이동한 뒤, 이어서 `(1000, 0)`만큼 이동을 요청합니다.

예상 출력:

```text
이동: true, 좌표: (7, 25)
범위 초과: false, 좌표: (7, 25)
```

필수 경계:

| 현재 좌표 | 이동량 | 성공 | 결과 |
| --- | --- | --- | --- |
| `(0, 0)` | `(-1000, 1000)` | true | `(-1000, 1000)` |
| `(1000, 0)` | `(0, 0)` | true | 그대로 |
| `(1000, 0)` | `(1, 0)` | false | 그대로 |
| `(0, -1000)` | `(0, -1)` | false | 그대로 |
| `(1, 2)` | `(2, 1000)` | false | **X도 1 보존** |
| `(1001, 0)` | `(-1, 0)` | false | 현재 상태가 무효이므로 보존 |
| `(0, 0)` | `(1001, 0)` | false | 이동량이 무효이므로 보존 |

```sh
gofmt -w _practice/08-pointers-structs/move-point/main.go
go run ./_practice/08-pointers-structs/move-point
```

완료 조건: 한 필드를 먼저 바꾼 뒤 다른 필드 검증에서 실패하는 부분 갱신이 없습니다. 모든 검사와 다음 값 계산이 끝난 후 두 필드에 대입합니다.

## 문제 6. 주문의 상품 목록까지 독립 복사하기

**수준: 응용**

개인 파일: `_practice/08-pointers-structs/clone-order/main.go`.

다음 타입을 선언합니다.

- `Item`: `Name string`, `Quantity int`.
- `Order`: `ID string`, `Items []Item`.

`func cloneOrder(order Order) Order`를 구현합니다.

### 계약

- `ID`, 원소 수, 원소 순서와 모든 필드 값을 보존합니다.
- 결과의 `Items` 원소를 바꿔도 원본 원소가 바뀌지 않아야 합니다.
- 원본 원소를 바꿔도 결과 원소가 바뀌지 않아야 합니다.
- 입력 `Items`가 nil이면 결과도 nil입니다.
- 입력 `Items`가 nil이 아닌 빈 슬라이스면 결과도 nil이 아닌 빈 슬라이스입니다.
- 값을 검증·정규화·합산하지 않습니다. 복사만 담당합니다.

`main`의 원본은 ID `"A-1"`, 상품 `book` 2개와 `pen` 3개입니다. 복사본의 ID를 `"A-2"`로, 첫 상품 수량을 `9`로 바꾼 뒤 출력합니다.

```text
원본: A-1, book 2개
복사본: A-2, book 9개
```

추가로 복사본 두 번째 상품 이름을 바꾼 뒤 원본 이름도 확인합니다. 마지막으로 원본 두 번째 상품 수량을 바꾸고 복사본이 보존되는지 확인합니다.

```sh
gofmt -w _practice/08-pointers-structs/clone-order/main.go
go run ./_practice/08-pointers-structs/clone-order
```

완료 조건: 구조체 대입만으로 끝내지 않고, `Items`의 원소 저장 공간을 분리합니다. `Item`에 나중에 슬라이스나 포인터 필드가 추가되면 어떤 부분을 재검토해야 하는지도 설명합니다.

## 선택 A. 검색한 원소의 포인터 반환하기

**수준: 선택 응용**

개인 파일: `_practice/08-pointers-structs/find-task/main.go`.

`Task` 구조체는 `Title string`, `Done bool`을 갖습니다. `func findTask(tasks []Task, title string) *Task`를 구현합니다.

- 제목이 같은 첫 번째 원소의 포인터를 반환합니다.
- 찾지 못하면 nil을 반환합니다. 빈 슬라이스와 nil 슬라이스도 해당합니다.
- 함수 자체는 `Done`을 바꾸지 않습니다.
- 호출자는 결과가 nil이 아닌지 검사한 뒤 `Done`을 바꿉니다.
- 포인터를 사용하는 동안 슬라이스에 `append`하거나 기반 배열을 바꾸지 않습니다.

`"읽기"`, `"실행"` 두 항목에서 `"실행"`을 찾고, 반환한 포인터의 `Done`을 `true`로 바꿉니다. 없는 `"배포"`를 검색한 결과도 확인합니다.

예상 출력:

```text
읽기: false
실행: true
없는 항목: true
```

중복 제목 두 개를 넣었다면 첫 원소만 변경되어야 합니다. range 값 변수의 주소를 반환해서 복사본만 바뀌는 구현은 계약을 만족하지 않습니다.

```sh
go run ./_practice/08-pointers-structs/find-task
```

## 단계별 힌트

힌트가 필요할 때만 해당 항목을 펼칩니다. 전체 구현은 해설에 있습니다.

<details>
<summary>문제 1: 포인터의 대상 표시하기</summary>

대상 변수 이름을 표에 적습니다. 포인터 변수에 새 주소를 대입하는 문장과 그 대상의 필드에 대입하는 문장을 따로 표시하세요.

</details>

<details>
<summary>문제 2: 기존 값을 잃지 않는 교환</summary>

한쪽을 먼저 덮어쓰면 이전 값이 사라집니다. 임시 변수 또는 다중 대입을 사용할 수 있습니다. nil 검사는 어느 역참조보다 앞에 둡니다.

</details>

<details>
<summary>문제 3: 변경한 구조체를 어디에 저장할까?</summary>

맵에서 꺼낸 값은 지역 변수에 놓습니다. 그 변수를 수정한 것만으로는 원래 맵 항목이 바뀌지 않습니다. 수정된 값 전체가 맵으로 돌아가야 합니다.

</details>

<details>
<summary>문제 4: 제로 값과 실패 여부</summary>

반환할 실패 값을 한 가지로 통일합니다. UTF-8 검사는 문자열 정리보다 먼저, 빈 문자열 검사는 양끝 공백 정리 뒤에 진행하면 계약과 순서가 맞습니다.

</details>

<details>
<summary>문제 5: 계산과 대입을 분리하기</summary>

새 좌표를 지역 변수 두 개에 계산합니다. 두 값이 모두 범위 안임을 확인한 뒤에만 원본 필드를 바꿉니다.

</details>

<details>
<summary>문제 6: 복사할 저장 공간의 범위</summary>

구조체 값의 복사는 먼저 해도 됩니다. 그다음 슬라이스 필드를 새 기반 배열을 가진 값으로 바꿉니다. nil과 길이 0은 별도로 판단합니다.

</details>

<details>
<summary>선택 A: 어느 변수의 주소인가?</summary>

슬라이스 원소에 접근하는 인덱스를 유지하세요. range 값 변수와 원래 원소가 같은 변수가 아니라는 점을 떠올립니다.

</details>

## 완료 후 확인

문제별 경계 입력을 직접 바꾸어 실행한 뒤 [해설과 참고 구현](../../solutions/08-pointers-structs/README.md)을 확인합니다. 참고 구현의 테스트를 실행해도 개인 파일이 자동으로 검사되지는 않습니다.

[본문으로](../../book/08-pointers-structs/README.md) · [전체 목차](../../book/README.md) · [공통 연습 안내](../README.md)
