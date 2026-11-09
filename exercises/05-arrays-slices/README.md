# 05장 연습문제: 배열과 슬라이스

선수 학습: [05. 배열과 슬라이스](../../book/05-arrays-slices/README.md).

출력뿐 아니라 **입력 보존과 결과의 독립성**을 확인합니다. 모든 정수와 누적값은 `int` 범위를 넘지 않는 작은 값으로 제한합니다. 입력은 코드에 직접 적습니다.

문제 1~6이 기본 과정입니다. 선택 A는 슬라이스 안에 슬라이스를 담는 본문 12절을 읽은 뒤 풉니다.

| 문제 | 수준 | 개인 폴더 |
| --- | --- | --- |
| 1. 공유 범위 추적 | 확인 | `predict` |
| 2. 통과 점수 선택 | 변형 | `passing` |
| 3. 0개 복사 오류 | 오류 진단 | `repair` |
| 4. 역순 복사 | 변형 | `reverse-copy` |
| 5. 지정 위치 삭제 | 응용 | `remove-at` |
| 6. 누적 합계 | 응용 | `running-totals` |
| 선택 A. 묶음으로 나누기 | 선택 | `chunks` |

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("predict", "passing", "repair", "reverse-copy", "remove-at", "running-totals", "chunks")) {
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
for practice_name in predict passing repair reverse-copy remove-at running-totals chunks; do
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


## 문제 1. 부분 슬라이스는 무엇을 공유할까? — 확인

작업 파일: `_practice/05-arrays-slices/predict/main.go`.

전체 프로그램의 출력 세 줄을 실행 전에 학습 노트에 적으세요.

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

1. `part`의 길이와 용량은 각각 얼마인가요?
2. `part[0]`은 `base`의 어느 원소인가요?
3. append할 때 새 저장 공간이 필요한가요?
4. `part := base[1:3:3]`으로 바꾸면 마지막 출력은 어떻게 달라지나요?

```sh
go run ./_practice/05-arrays-slices/predict
```

<details>
<summary>힌트 1: 두 슬라이스의 위치 대응</summary>

부분 슬라이스의 첫 위치는 다시 0입니다. 원본에서 선택을 시작한 위치와 연결해 보세요.

</details>

<details>
<summary>힌트 2: 길이와 용량을 따로 적기</summary>

길이는 끝에서 시작을 뺀 값입니다. 일반적인 부분 슬라이스 용량은 시작부터 원래 저장 공간의 끝까지 셉니다. 세 번째 경계는 용량을 제한합니다.

</details>

완료 조건: 공유 위치와 추가 위치를 설명합니다. 변형 후 원래 코드로 복구합니다.

## 문제 2. 통과 점수만 독립적으로 모으기 — 변형

작업 파일: `_practice/05-arrays-slices/passing/main.go`.

`func passing(scores []int) []int`를 작성합니다. 입력은 0~100 범위이며 60 이상인 점수만 반환합니다.

| 계약 | 요구사항 |
| --- | --- |
| 순서·중복 | 입력 순서대로 모두 유지 |
| 입력 변경 | 금지 |
| 결과 변경 | 원본에 영향을 주지 않음 |
| 통과 점수가 없음 | nil 반환 |

함수는 출력하지 않습니다. `main`의 기본 입력은 `[]int{59, 60, 90, 30, 100}`이며 다음 두 줄을 출력합니다.

```text
통과 점수: [60 90 100]
원본: [59 60 90 30 100]
```

```sh
go run ./_practice/05-arrays-slices/passing
```

| 경계 입력 | 결과 |
| --- | --- |
| nil·빈 슬라이스 | nil |
| `[0 59]` | nil |
| `[60]` | `[60]` |
| `[100 60 100]` | `[100 60 100]` |

결과가 있을 때 첫 원소를 0으로 바꾸고 원본을 확인합니다. 원본은 유지되어야 합니다. 실험 뒤 결과 변경 문장을 제거하고 기본 입력으로 복구합니다.

<details>
<summary>힌트 1: 결과를 어디에서 시작할까요?</summary>

입력의 부분 범위를 결과로 쓰면 공유할 수 있습니다. 별도의 결과를 빈 상태에서 시작하고 통과한 값만 추가해 보세요.

</details>

<details>
<summary>힌트 2: 아무 값도 통과하지 않을 때</summary>

`var result []int`는 nil입니다. 추가한 값이 없다면 그대로 반환할 수 있습니다. 반환은 모든 원소를 확인한 뒤 합니다.

</details>

완료 조건: 경계 60, 중복, 빈 입력과 원본 독립성을 설명합니다.

## 문제 3. 용량이 있는데 왜 0개 복사될까? — 오류 진단

작업 파일: `_practice/05-arrays-slices/repair/main.go`.

아래는 **컴파일되지만 의도와 다르게 동작하는 전체 코드**입니다.

```go
package main

import "fmt"

func main() {
	source := []int{10, 20, 30}
	target := make([]int, 0, len(source))
	count := copy(target, source)
	fmt.Println("복사 수:", count)
	fmt.Println("복사 값:", target)
}
```

목적지 생성 줄만 고쳐 세 원소를 모두 복사하세요. copy 호출과 출력은 유지합니다.

```text
복사 수: 3
복사 값: [10 20 30]
```

```sh
go run ./_practice/05-arrays-slices/repair
```

빈 `source`에서도 오류 없이 복사 수 0과 빈 목록을 출력해야 합니다. 확인 후 원래 입력으로 복구합니다.

<details>
<summary>힌트 1: copy가 사용하는 길이</summary>

실행 직전 `len(target)`과 `cap(target)`을 따로 적어 보세요. copy는 현재 원소 범위에만 씁니다.

</details>

<details>
<summary>힌트 2: 쓸 원소 자리가 먼저 필요합니다</summary>

`make`의 길이 인자를 원본 원소 수와 연결해 보세요. copy 자체가 목적지 길이를 늘려 주지는 않습니다.

</details>

완료 조건: 수정 전 복사 수와 수정 후 결과를 설명합니다.

## 문제 4. 원본을 유지하는 역순 복사 — 변형

작업 파일: `_practice/05-arrays-slices/reverse-copy/main.go`.

`func reversed(values []int) []int`를 작성합니다. 순서를 뒤집은 새 슬라이스를 반환하며 입력과 결과는 독립적입니다. 빈 입력은 nil을 반환합니다.

기본 입력은 `[]int{10, 20, 30, 40}`입니다.

```text
뒤집은 복사: [40 30 20 10]
원본: [10 20 30 40]
```

```sh
go run ./_practice/05-arrays-slices/reverse-copy
```

| 입력 | 결과 |
| --- | --- |
| nil·빈 슬라이스 | nil |
| `[7]` | `[7]` |
| `[1 2 3]` | `[3 2 1]` |
| `[1 2 2 1]` | `[1 2 2 1]` |

<details>
<summary>힌트 1: 결과 길이는 미리 알 수 있습니다</summary>

원소 수는 바뀌지 않습니다. 결과 길이를 정해서 만들면 각 위치에 직접 대입할 수 있습니다.

</details>

<details>
<summary>힌트 2: 양 끝의 대응</summary>

입력 첫 원소는 결과 마지막 위치로 갑니다. 길이 `n`, 입력 위치 `i`를 사용해 결과 위치를 계산해 보세요. 마지막 위치는 `n-1`입니다.

</details>

완료 조건: 홀수·짝수·단일 길이를 확인합니다. 결과를 두 번 뒤집으면 원래 순서가 되고, 결과 수정은 입력에 전파되지 않아야 합니다.

## 문제 5. 지정 위치를 삭제한 새 목록 — 응용

작업 파일: `_practice/05-arrays-slices/remove-at/main.go`.

`func removeAt(values []int, index int) ([]int, bool)`를 작성합니다.

| 상황 | 계약 |
| --- | --- |
| `index < 0` 또는 `index >= len(values)` | `nil, false` |
| 유효한 위치 | 그 위치 하나를 제외한 새 목록과 `true` |
| 하나뿐인 원소를 삭제 | 길이 0의 결과와 `true` |
| 원본·결과 변경 | 서로 독립적 |

나머지 원소의 순서와 중복은 유지합니다. 같은 값을 모두 지우는 것이 아니라 위치 하나를 지우는 문제입니다.

기본 입력 `[10 20 30 40]`, 위치 1의 출력:

```text
삭제 성공: true
결과: [10 30 40]
원본: [10 20 30 40]
```

```sh
go run ./_practice/05-arrays-slices/remove-at
```

| 입력 | 위치 | 결과 | 성공 |
| --- | --- | --- | --- |
| 빈 슬라이스 | 0 | 빈 결과 | 거짓 |
| `[1]` | -1 또는 1 | 빈 결과 | 거짓 |
| `[1]` | 0 | 빈 결과 | 참 |
| `[1 2 3]` | 0 | `[2 3]` | 참 |
| `[1 2 3]` | 2 | `[1 2]` | 참 |

<details>
<summary>힌트 1: 검사 순서</summary>

부분 슬라이스를 만들기 전에 인덱스를 검사합니다. 유효한 경로에서만 앞부분과 뒷부분을 계산하세요.

</details>

<details>
<summary>힌트 2: 추가할 목적지는 어디인가요?</summary>

`append(values[:index], ...)`는 원본을 바꿀 수 있습니다. 새 공간에 앞과 뒤의 원소를 추가하거나 복사합니다.

</details>

완료 조건: 첫·중간·끝과 잘못된 위치를 검사합니다. 성공한 빈 결과와 실패를 bool로 구분합니다.

## 문제 6. 하루별 누적 합계 — 응용

작업 파일: `_practice/05-arrays-slices/running-totals/main.go`.

`func runningTotals(values []int) []int`를 작성합니다. 결과의 각 위치에는 처음부터 해당 위치까지의 합계를 저장합니다. 입력과 같은 길이이고 원본과 독립적이어야 합니다. 빈 입력은 nil을 반환합니다.

음수는 차감 기록으로 허용하며 모든 누적값은 `int` 범위 안이어야 합니다.

```text
일별: [20 0 30 15]
누적: [20 20 50 65]
```

```sh
go run ./_practice/05-arrays-slices/running-totals
```

| 입력 | 결과 |
| --- | --- |
| 빈 슬라이스 | nil |
| `[5]` | `[5]` |
| `[5 -5 2]` | `[5 0 2]` |
| `[0 0]` | `[0 0]` |

<details>
<summary>힌트 1: 반복 사이에 유지할 값</summary>

합계를 반복 전에 0으로 둡니다. 현재 원소를 더한 뒤 그 시점의 합계를 결과에 씁니다.

</details>

<details>
<summary>힌트 2: 결과 위치</summary>

range에서 인덱스와 원소를 함께 받습니다. 원본 원소에 덮어쓰지 말고 별도 결과 위치에 씁니다.

</details>

완료 조건: 빈 입력, 0, 음수를 확인합니다. 마지막 결과가 전체 합계와 같음을 설명합니다.

## 선택 A. 독립적인 묶음으로 나누기

작업 파일: `_practice/05-arrays-slices/chunks/main.go`.

`func chunks(values []int, size int) ([][]int, bool)`를 작성합니다. size는 1~100만 유효하며 그 밖에는 `nil, false`를 반환합니다.

앞에서부터 최대 size개씩 묶고 마지막에는 남은 값을 담습니다. 각 묶음은 원본과 독립적입니다. 빈 입력과 유효한 크기에서는 `nil, true`입니다.

기본 입력 `[10 20 30 40 50]`, 크기 2:

```text
성공: true
묶음: [[10 20] [30 40] [50]]
원본: [10 20 30 40 50]
```

```sh
go run ./_practice/05-arrays-slices/chunks
```

크기 1, 입력보다 큰 10, 잘못된 0·101과 빈 입력을 확인합니다. `result[0][0]`을 변경하는 실험은 결과가 있을 때만 실행하세요.

<details>
<summary>힌트 1: 마지막 경계</summary>

남은 값이 size보다 적으면 전체 길이가 끝입니다. 다음 시작 위치로 실제 끝 경계를 사용하면 반복이 진행됩니다.

</details>

<details>
<summary>힌트 2: 안쪽 원소의 독립성</summary>

각 묶음 길이만큼 새 슬라이스를 만들고 copy한 뒤 바깥 결과에 append합니다. 부분 범위를 그대로 추가하면 원소를 공유합니다.

</details>

완료 조건: 묶음을 이어 붙이면 원래 원소열이 되며 각 길이는 size 이하이고 원본과 공유하지 않습니다.

## 해설을 열기 전

- [ ] 길이와 용량을 따로 계산했다.
- [ ] 빈 입력을 안전하게 처리했다.
- [ ] 순서와 중복을 유지했다.
- [ ] 결과 수정 후 원본 독립성을 확인했다.
- [ ] 성공한 빈 결과와 실패를 구분했다.
- [ ] 개인 코드를 포맷하고 기본 입력으로 복구했다.

```sh
go fmt ./_practice/05-arrays-slices/passing
go run ./_practice/05-arrays-slices/passing
```

다른 문제도 폴더 이름을 바꾸어 실행합니다. 참고 구현의 테스트가 개인 파일까지 검사하지는 않습니다.

[해설](../../solutions/05-arrays-slices/README.md) · [본문](../../book/05-arrays-slices/README.md)
