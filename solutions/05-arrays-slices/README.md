# 05장 해설: 배열과 슬라이스

[문제](../../exercises/05-arrays-slices/README.md)를 먼저 풀고 비교합니다. 모든 명령은 저장소 루트 기준입니다. 출력이 같아도 원본 변경이나 결과의 공유 관계가 다르면 계약을 만족하지 못할 수 있습니다.

문제 2·4·5·6은 참고 패키지와 함수 테스트가 있습니다. 문제 1·3·선택 A의 전체 코드는 개인 폴더에서 실행합니다. 기존 파일은 별도 폴더에 보관한 뒤 작성하세요.

## 문제 1. 공유 범위 추적

예상 출력:

```text
부분: [99 30]
원본: [10 99 30 40]
추가 뒤 원본: [10 99 30 77]
```

전체 코드:

```go
package main

import "fmt"

func main() {
	base := []int{10, 20, 30, 40}
	part := base[1:3]
	part[0] = 99
	fmt.Println("부분:", part)
	fmt.Println("원본:", base)
	part = append(part, 77)
	fmt.Println("추가 뒤 원본:", base)
}
```

```sh
go run ./_practice/05-arrays-slices/predict
```

| 단계 | part 길이·용량 | base |
| --- | --- | --- |
| 부분 선택 | 2·3 | `[10 20 30 40]` |
| 원소 변경 | 2·3 | `[10 99 30 40]` |
| 추가 | 3·3 | `[10 99 30 77]` |

`part[0]`은 `base[1]`이고 append할 다음 위치는 `base[3]`입니다. 용량이 남아 있으므로 같은 저장 공간을 사용합니다.

`base[1:3:3]`으로 바꾸면 길이와 용량이 모두 2입니다. 첫 원소 변경은 공유하지만 추가에는 새 공간이 필요합니다. 마지막 출력은 `추가 뒤 원본: [10 99 30 40]`입니다.

부분 선택과 append가 항상 복사본을 만든다는 가정은 모두 틀립니다. [슬라이스 표현식](https://go.dev/ref/spec#Slice_expressions)과 [append](https://go.dev/ref/spec#Appending_and_copying_slices)의 조건을 구분합니다.

## 문제 2. 통과 점수 선택

참고 구현: [passing/main.go](passing/main.go) · [테스트](passing/main_test.go).

```go
package main

import "fmt"

func passing(scores []int) []int {
	var result []int
	for _, score := range scores {
		if score >= 60 {
			result = append(result, score)
		}
	}
	return result
}
func main() {
	scores := []int{59, 60, 90, 30, 100}
	fmt.Println("통과 점수:", passing(scores))
	fmt.Println("원본:", scores)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/05-arrays-slices/passing
go test ./solutions/05-arrays-slices/passing
```

예상 출력:

```text
통과 점수: [60 90 100]
원본: [59 60 90 30 100]
```


### 선택한 이유

별도 result에 정수 값을 append하므로 원본 저장 공간을 사용하지 않습니다. 순회 순서대로 추가하므로 순서와 중복도 보존됩니다. 점수 0~100의 전제는 호출자가 지킵니다.

| 현재 값 | 조건 | 누적 결과 |
| --- | --- | --- |
| 59 | 거짓 | `[]` |
| 60 | 참 | `[60]` |
| 90 | 참 | `[60 90]` |
| 30 | 거짓 | `[60 90]` |
| 100 | 참 | `[60 90 100]` |

nil·빈 입력에서는 순회하지 않습니다. 아무 점수도 통과하지 않으면 처음의 nil 결과를 반환합니다.

| 입력 | 결과 |
| --- | --- |
| nil·빈 슬라이스 | nil |
| `[0 59]` | nil |
| `[60]` | `[60]` |
| `[100 60 100]` | `[100 60 100]` |

테스트는 결과를 수정해도 원본이 유지되는지 확인합니다. 함수가 실행 중 입력을 바꾸지 않았다는 확인만으로는 이후 결과 변경의 영향을 알아낼 수 없습니다.

```sh
go run ./_practice/05-arrays-slices/passing
```

## 문제 3. 목적지 길이 수정

원래 목적지는 길이 0, 용량 3입니다. copy가 쓸 현재 자리가 없어 0개를 복사했습니다. 생성 시 길이를 원본 길이로 맞춥니다.

```go
package main

import "fmt"

func main() {
	source := []int{10, 20, 30}
	target := make([]int, len(source))
	count := copy(target, source)
	fmt.Println("복사 수:", count)
	fmt.Println("복사 값:", target)
}
```

```sh
go run ./_practice/05-arrays-slices/repair
```

예상 출력:

```text
복사 수: 3
복사 값: [10 20 30]
```

| 상태 | 길이 | 용량 | 복사 수 |
| --- | --- | --- | --- |
| 수정 전 | 0 | 3 | 0 |
| 수정 후 | 3 | 3 | 3 |

빈 입력에서는 길이 0인 목적지를 만들고 0개를 복사합니다. nil 여부를 보존하는 요구는 이 문제에 없습니다. 필요한 계약이라면 입력 nil을 따로 검사해야 합니다.

길이 0에서 append하는 대안도 있지만 이 문제는 목적지 생성 줄만 고치도록 요구했습니다. 길이를 준비하는 작업과 원소 값을 복사하는 작업을 분리해 이해하기 위한 제한입니다.

## 문제 4. 역순의 독립 복사

참고 구현: [reverse-copy/main.go](reverse-copy/main.go) · [테스트](reverse-copy/main_test.go).

```go
package main

import "fmt"

func reversed(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	result := make([]int, len(values))
	for i, value := range values {
		result[len(values)-1-i] = value
	}
	return result
}
func main() {
	values := []int{10, 20, 30, 40}
	fmt.Println("뒤집은 복사:", reversed(values))
	fmt.Println("원본:", values)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/05-arrays-slices/reverse-copy
go test ./solutions/05-arrays-slices/reverse-copy
```

예상 출력:

```text
뒤집은 복사: [40 30 20 10]
원본: [10 20 30 40]
```


### 인덱스 대응

길이 n에서 입력 위치 i는 결과 위치 `n-1-i`로 이동합니다. `n-i`로 쓰면 첫 대입이 범위를 벗어납니다.

| 입력 위치 | 값 | 결과 위치 |
| --- | --- | --- |
| 0 | 10 | 3 |
| 1 | 20 | 2 |
| 2 | 30 | 1 |
| 3 | 40 | 0 |

한 원소는 위치 0에서 0으로 옮겨집니다. 홀수 길이의 가운데 값도 별도 분기 없이 가운데로 이동합니다.

새 결과에 끝에서부터 append하는 풀이도 가능합니다. 원본의 양 끝을 교환하는 방식은 입력 보존 계약에 맞지 않습니다.

| 입력 | 결과 |
| --- | --- |
| 빈 입력 | nil |
| `[7]` | `[7]` |
| `[1 2 3]` | `[3 2 1]` |
| `[1 2 2 1]` | `[1 2 2 1]` |

빈 입력을 nil로 통일한 것은 문제의 설계 선택입니다. 본문의 cloneInts처럼 nil과 빈 입력을 구별하여 보존하는 다른 계약도 가능합니다.

```sh
go run ./_practice/05-arrays-slices/reverse-copy
```

## 문제 5. 지정 위치 삭제

참고 구현: [remove-at/main.go](remove-at/main.go) · [테스트](remove-at/main_test.go).

```go
package main

import "fmt"

func removeAt(values []int, index int) ([]int, bool) {
	if index < 0 || index >= len(values) {
		return nil, false
	}
	result := make([]int, 0, len(values)-1)
	result = append(result, values[:index]...)
	result = append(result, values[index+1:]...)
	return result, true
}
func main() {
	values := []int{10, 20, 30, 40}
	result, ok := removeAt(values, 1)
	fmt.Println("삭제 성공:", ok)
	fmt.Println("결과:", result)
	fmt.Println("원본:", values)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/05-arrays-slices/remove-at
go test ./solutions/05-arrays-slices/remove-at
```

예상 출력:

```text
삭제 성공: true
결과: [10 30 40]
원본: [10 20 30 40]
```


### 검사와 복사 순서

부분 범위를 만들기 전에 인덱스를 검사합니다. 검사를 뒤로 미루면 범위 표현식 자체에서 먼저 실패할 수 있습니다.

유효한 경우 새 공간에 삭제 위치 앞과 뒤의 원소를 각각 추가합니다. 결과의 순서와 중복은 그대로 유지되고 길이는 하나 줄어듭니다.

| 기본 위치 | 값 | 처리 |
| --- | --- | --- |
| 0 | 10 | 앞부분 복사 |
| 1 | 20 | 제외 |
| 2 | 30 | 뒷부분 복사 |
| 3 | 40 | 뒷부분 복사 |

`append(values[:index], values[index+1:]...)`만 사용하면 원본 공간을 재사용할 수 있습니다. 새 목적지를 선택한 이유는 원본과 결과의 독립성이 계약이기 때문입니다.

| 입력·위치 | 반환 |
| --- | --- |
| 빈 입력, 0 | nil, 거짓 |
| `[1]`, -1 또는 1 | nil, 거짓 |
| `[1]`, 0 | 빈 결과, 참 |
| `[1 2 3]`, 0 | `[2 3]`, 참 |
| `[1 2 3]`, 2 | `[1 2]`, 참 |

성공한 빈 삭제와 실패는 길이만으로 구분할 수 없습니다. 호출자는 ok를 확인한 뒤 결과를 삭제 결과로 사용합니다.

```sh
go run ./_practice/05-arrays-slices/remove-at
```

## 문제 6. 누적 합계

참고 구현: [running-totals/main.go](running-totals/main.go) · [테스트](running-totals/main_test.go).

```go
package main

import "fmt"

func runningTotals(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	result := make([]int, len(values))
	total := 0
	for i, value := range values {
		total += value
		result[i] = total
	}
	return result
}
func main() {
	minutes := []int{20, 0, 30, 15}
	fmt.Println("일별:", minutes)
	fmt.Println("누적:", runningTotals(minutes))
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/05-arrays-slices/running-totals
go test ./solutions/05-arrays-slices/running-totals
```

예상 출력:

```text
일별: [20 0 30 15]
누적: [20 20 50 65]
```


### 반복 사이의 상태

합계를 0에서 시작하고 현재 원소를 더한 다음 결과에 기록합니다. 0인 기록도 한 자리를 차지하므로 결과 길이는 줄지 않습니다.

| 위치 | 현재 값 | 이전 합계 | 기록할 합계 |
| --- | --- | --- | --- |
| 0 | 20 | 0 | 20 |
| 1 | 0 | 20 | 20 |
| 2 | 30 | 20 | 50 |
| 3 | 15 | 50 | 65 |

원본을 차례로 덮어쓰는 구현도 합계를 만들지만 이 문제의 입력 보존 계약에는 맞지 않습니다. 결과를 별도로 만들어 대입합니다.

검사할 관계는 다음과 같습니다.

- 결과 길이는 입력 길이와 같습니다.
- 첫 결과는 첫 입력과 같습니다.
- 마지막 결과는 전체 입력의 합계입니다.
- 두 번째 이후 결과에서 이전 결과를 빼면 현재 입력입니다.
- 결과 수정이 입력에 전파되지 않습니다.

빈 입력에는 첫·마지막 원소가 없으므로 그 관계를 검사하기 전에 길이를 확인합니다. `[5 -5 2]`는 `[5 0 2]`이므로 중간 합계가 0인 경우도 정상입니다.

```sh
go run ./_practice/05-arrays-slices/running-totals
```

## 선택 A. 독립 묶음

개인 파일 `_practice/05-arrays-slices/chunks/main.go`에서 실행할 전체 코드입니다.

```go
package main

import "fmt"

func chunks(values []int, size int) ([][]int, bool) {
	if size < 1 || size > 100 {
		return nil, false
	}
	var result [][]int
	for start := 0; start < len(values); {
		end := len(values)
		if size < len(values)-start {
			end = start + size
		}
		part := make([]int, end-start)
		copy(part, values[start:end])
		result = append(result, part)
		start = end
	}
	return result, true
}

func main() {
	values := []int{10, 20, 30, 40, 50}
	result, ok := chunks(values, 2)
	fmt.Println("성공:", ok)
	fmt.Println("묶음:", result)
	fmt.Println("원본:", values)
}
```

```sh
go run ./_practice/05-arrays-slices/chunks
```

예상 출력:

```text
성공: true
묶음: [[10 20] [30 40] [50]]
원본: [10 20 30 40 50]
```

| 시작 | 남은 수 | 끝 | 복사한 원소 |
| --- | --- | --- | --- |
| 0 | 5 | 2 | 10, 20 |
| 2 | 3 | 4 | 30, 40 |
| 4 | 1 | 5 | 50 |

끝을 먼저 전체 길이로 두고, 묶음 크기가 남은 수보다 작을 때만 줄입니다. 처리 후 `start = end`로 옮겨 실제 처리한 범위만큼 진행합니다.

각 묶음을 새 공간에 copy하므로 안쪽 결과를 바꾸어도 원본이 유지됩니다. 부분 슬라이스를 그대로 바깥 결과에 append하면 이 독립성은 없어집니다.

| 입력·크기 | 결과 |
| --- | --- |
| 빈 입력, 2 | nil, 참 |
| `[1 2]`, 1 | `[[1] [2]]`, 참 |
| `[1 2]`, 10 | `[[1 2]]`, 참 |
| `[1 2]`, 0 또는 101 | nil, 거짓 |

빈 입력에서도 크기는 검사합니다. 따라서 빈 입력에 잘못된 크기를 전달하면 false입니다. 입력 두 조건의 우선순위도 함수 계약에 포함됩니다.

## 테스트가 확인하는 범위

```sh
go test -count=1 ./solutions/05-arrays-slices/...
```

| 패키지 | 중요 검사 |
| --- | --- |
| passing | 59·60, 빈 결과, 순서·중복, 결과 변경 뒤 입력 유지 |
| reverse-copy | 빈·단일·홀수 길이, 입력 보존, 원소 독립성 |
| remove-at | 음수·길이와 같은 인덱스, 첫·중간·끝, 빈 성공 |
| running-totals | 빈·단일, 0·음수, 입력 보존과 독립성 |

Example_main은 기본 출력도 확인합니다. 테스트의 구조체·보조 함수 사용법은 후속 장에서 배우며, 지금은 함수 계약과 입력 사례를 먼저 읽으면 됩니다.

참고 구현 검사는 개인 파일까지 검사하지 않습니다. 문제의 명령과 경계 표로 자신의 결과를 따로 확인합니다.

## 다시 설명해 볼 질문

1. 실행 중 입력을 변경하지 않아도 결과의 나중 수정이 입력을 바꿀 수 있나요?
2. 결과 길이가 미리 정해지는 문제와 조건에 따라 달라지는 문제의 생성 방식은 어떻게 다른가요?
3. 부분 범위보다 인덱스 검사를 먼저 해야 하는 이유는 무엇인가요?
4. 바깥 복사와 안쪽 원소 복사는 어떻게 다른가요?
5. nil을 유지해야 하는지는 어디에서 결정하나요?

[본문](../../book/05-arrays-slices/README.md) · [문제](../../exercises/05-arrays-slices/README.md)
