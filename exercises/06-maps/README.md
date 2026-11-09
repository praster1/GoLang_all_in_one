# 06장 연습문제: 맵과 집계

선수 학습: [06. 맵과 집계](../../book/06-maps/README.md).

키의 존재, 원소 개수, 출력 순서, 함수의 입력 변경 여부를 구분합니다. 입력은 코드에 직접 적고 정수와 합계는 int 범위 안의 작은 값으로 제한합니다.

기본 문제는 1~6입니다. 선택 A는 같은 내용의 맵을 비교하는 과제입니다.

| 문제 | 수준 | 개인 폴더 |
| --- | --- | --- |
| 1. 영값과 존재 여부 추적 | 확인 | `predict` |
| 2. 항목 빈도 | 변형 | `tally` |
| 3. nil 맵 쓰기 오류 | 오류 진단 | `repair` |
| 4. 처음 본 순서의 고유 목록 | 변형 | `unique` |
| 5. 재고 출고 | 응용 | `withdraw` |
| 6. 두 집계의 독립 병합 | 응용 | `merge-counts` |
| 선택 A. 맵 내용 비교 | 선택 | `same-counts` |

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("predict", "tally", "repair", "unique", "withdraw", "merge-counts", "same-counts")) {
    $practiceDir = Join-Path "_practice/06-maps" $practiceName
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
for practice_name in predict tally repair unique withdraw merge-counts same-counts; do
    practice_dir="_practice/06-maps/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

한 폴더에는 하나의 `main`만 둡니다. 새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice/`는 공통 `go test ./...`에 포함되지 않으므로 개인 폴더를 직접 지정해서 실행하세요. [공통 연습 안내](../../exercises/README.md)를 참고합니다.


## 문제 1. 0은 어디에서 왔을까? — 확인

작업 파일: `_practice/06-maps/predict/main.go`.

다음 전체 코드의 출력 네 줄을 먼저 예상합니다.

```go
package main

import "fmt"

func main() {
	counts := map[string]int{"zero": 0}
	fmt.Println(counts["missing"], len(counts))
	value, ok := counts["zero"]
	fmt.Println(value, ok)
	value, ok = counts["missing"]
	fmt.Println(value, ok)
	counts["missing"]++
	delete(counts, "zero")
	fmt.Println(counts["missing"], len(counts))
}
```

1. 첫 조회 뒤 키가 늘어나나요?
2. zero와 missing 조회에서 값과 ok가 각각 어떻게 다른가요?
3. missing에 ++하면 어떤 값이 저장되나요?
4. delete 뒤 키는 몇 개인가요?

```sh
go run ./_practice/06-maps/predict
```

<details>
<summary>힌트 1: 조회와 쓰기 구분</summary>

조회만 하면 항목을 등록하지 않습니다. 증가 연산은 조회한 값을 바탕으로 다시 저장하는 문장입니다.

</details>

<details>
<summary>힌트 2: 값과 키 수는 별도 정보입니다</summary>

0인 값도 등록된 항목입니다. 존재 여부와 len을 값의 합계와 혼동하지 마세요.

</details>

완료 조건: 각 줄에서 실제로 등록된 키들을 적고 결과를 설명합니다.

## 문제 2. 항목 빈도 세기 — 변형

작업 파일: `_practice/06-maps/tally/main.go`.

`func tally(items []string) map[string]int`를 작성합니다. 같은 문자열이 몇 번 등장했는지 반환합니다. 빈 입력에도 **nil이 아닌 빈 맵**을 반환하세요.

문자열은 그대로 비교합니다. 대소문자를 자동으로 통일하거나 공백을 지우지 않습니다. 빈 문자열도 유효한 키입니다.

기본 입력은 `[]string{"pen", "book", "pen", "note", "book", "pen"}`입니다. main에서 키를 정렬하여 다음을 출력합니다.

```text
book: 2
note: 1
pen: 3
```

```sh
go run ./_practice/06-maps/tally
```

| 입력 | 반환 항목 |
| --- | --- |
| nil·빈 슬라이스 | 생성된 빈 맵 |
| `[a]` | a:1 |
| `[a a A]` | a:2, A:1 |
| `[]string{"", ""}` | 빈 문자열 키의 값 2 |

<details>
<summary>힌트 1: 처음 본 항목의 출발 값</summary>

없는 키를 읽으면 int의 영값입니다. 항목을 처음 보았는지 따로 분기하지 않고 증가시킬 수 있습니다.

</details>

<details>
<summary>힌트 2: 집계와 표시를 분리하기</summary>

집계 함수는 맵을 반환합니다. main에서 키를 []string에 모으고 sort.Strings로 정렬한 뒤 그 순서로 조회하세요.

</details>

완료 조건: 전체 빈도의 합이 입력 길이와 같고, 키 수는 서로 다른 문자열 개수와 같습니다. 기본 출력 순서가 반복 실행에도 유지되어야 합니다.

## 문제 3. nil 맵 쓰기 오류 고치기 — 오류 진단

작업 파일: `_practice/06-maps/repair/main.go`.

아래는 **의도적으로 실행 중 실패하는 전체 코드**입니다.

```go
package main

import "fmt"

func main() {
	var stock map[string]int
	stock["pen"] = 3
	fmt.Println(stock["pen"])
}
```

실행해 오류를 확인하고 맵을 준비하는 줄만 수정하세요. 항목 대입과 출력은 유지합니다.

```sh
go run ./_practice/06-maps/repair
```

수정 후 출력은 `3` 한 줄입니다. 단순히 쓰기 문장을 지우거나 임의의 복구 코드로 실패를 감추는 방식은 사용하지 않습니다.

<details>
<summary>힌트 1: 선언과 생성</summary>

타입을 가진 맵 변수를 선언하는 것과 쓰기 가능한 맵을 만드는 것은 다릅니다. 선언 직후 값이 무엇인지 생각하세요.

</details>

<details>
<summary>힌트 2: 두 가지 생성 방법</summary>

make 또는 빈 맵 리터럴을 사용할 수 있습니다. 어느 방법이든 쓰기 전에 준비되어야 합니다.

</details>

완료 조건: nil 맵에서 조회·delete는 가능하지만 쓰기는 실패하는 이유를 설명합니다.

## 문제 4. 처음 본 순서의 고유 목록 — 변형

작업 파일: `_practice/06-maps/unique/main.go`.

`func unique(items []string) []string`를 작성합니다. 중복은 한 번만 남기고 첫 등장 순서를 유지합니다. 입력은 바꾸지 않고 결과 원소를 수정해도 입력이 유지되어야 합니다. 빈 입력의 결과는 nil입니다.

기본 입력 `[]string{"go", "map", "go", "slice", "map"}`의 출력:

```text
고유 목록: [go map slice]
원본: [go map go slice map]
```

```sh
go run ./_practice/06-maps/unique
```

| 입력 | 결과 |
| --- | --- |
| nil·빈 슬라이스 | nil |
| `[a a a]` | `[a]` |
| `[b a b A]` | `[b a A]` |
| `[]string{"", "a", ""}` | `[]string{"", "a"}` |

<details>
<summary>힌트 1: 서로 다른 두 책임</summary>

이미 본 항목인지 검사하는 맵과, 처음 본 순서를 보관하는 슬라이스를 함께 사용합니다.

</details>

<details>
<summary>힌트 2: 어느 자료를 순회하나요?</summary>

결과를 맵의 키 순회로 만들면 첫 등장 순서를 잃습니다. 입력을 순회하면서 처음 본 순간에만 결과에 추가하세요.

</details>

완료 조건: 대소문자와 빈 문자열을 구분하고 결과가 정렬 순서가 아니라 첫 등장 순서인지 확인합니다.

## 문제 5. 실패하면 재고를 바꾸지 않는 출고 — 응용

작업 파일: `_practice/06-maps/withdraw/main.go`.

`func withdraw(stock map[string]int, item string, quantity int) bool`을 작성합니다. 이 함수는 성공할 때 **입력 맵의 해당 재고를 직접 줄입니다.**

| 조건 | 처리 |
| --- | --- |
| 상품 없음 또는 nil 맵 | false, 변경 없음 |
| quantity가 0 이하 | false, 변경 없음 |
| quantity가 재고보다 큼 | false, 변경 없음 |
| 나머지 | 수량만큼 차감하고 true |
| 재고가 0이 됨 | 키는 삭제하지 않고 0으로 유지 |

재고 입력은 0 이상의 작은 정수입니다. 기본 맵은 `{"pen":5, "book":0}`입니다. 순서대로 pen 3개, pen 3개, 없는 note 1개를 출고합니다.

```text
첫 출고: true
남은 pen: 2
초과 출고: false
없는 상품: false
최종 pen: 2
```

```sh
go run ./_practice/06-maps/withdraw
```

각 경계 실험은 새 재고 맵으로 시작합니다. 수량 -1·0·6, 재고 전부인 5, 재고 0인 book, 없는 키와 nil을 확인하세요.

<details>
<summary>힌트 1: 없는 상품을 구분하기</summary>

두 결과 조회로 재고와 존재 여부를 함께 읽습니다. 없는 상품에 잘못 0을 등록하지 않아야 합니다.

</details>

<details>
<summary>힌트 2: 변경 전에 검사 끝내기</summary>

유효성 검사가 모두 끝난 뒤 한 번만 재고를 대입합니다. 먼저 차감한 다음 실패를 발견하는 흐름은 입력 복구 문제를 만듭니다.

</details>

완료 조건: 실패한 모든 호출에서 키 개수와 원래 수량이 유지됩니다. 재고 전부를 출고한 뒤 키가 남는지도 확인합니다.

## 문제 6. 두 집계의 독립 병합 — 응용

작업 파일: `_practice/06-maps/merge-counts/main.go`.

`func mergeCounts(a, b map[string]int) map[string]int`를 작성합니다. 양쪽의 키를 모두 포함하고 같은 키는 두 값을 더합니다. 입력은 0 이상의 횟수이고 합은 int 범위 안입니다.

두 입력을 변경하지 않으며 결과를 수정해도 입력은 바뀌지 않습니다. nil 입력은 빈 항목 집합으로 처리하고, 두 입력이 모두 비어도 생성된 빈 맵을 반환합니다. 값이 0인 키도 유지합니다.

기본 입력은 a=`{"go":2,"map":1}`, b=`{"go":3,"slice":2}`입니다. 키를 정렬해 출력합니다.

```text
go: 5
map: 1
slice: 2
원본 go: 2 3
```

```sh
go run ./_practice/06-maps/merge-counts
```

| 입력 | 확인할 결과 |
| --- | --- |
| nil, nil | 생성된 빈 맵 |
| nil, b | b와 같은 항목, 독립 공간 |
| a, nil | a와 같은 항목, 독립 공간 |
| `{"zero":0}`, 빈 맵 | zero 키와 값 0 유지 |
| `{"x":2}`, `{"x":3}` | x:5 |

<details>
<summary>힌트 1: 목적지를 어디에 만들까요?</summary>

`result := a`는 독립 복사가 아닙니다. 새 맵을 만들고 입력 항목을 그곳에 기록합니다.

</details>

<details>
<summary>힌트 2: 두 번째 입력 합산</summary>

첫 맵의 항목을 복사한 뒤 두 번째 맵의 값은 기존 결과에 더합니다. 없는 키의 영값을 이용할 수 있습니다.

</details>

완료 조건: 한쪽 입력이 nil일 때도 결과를 그대로 다른 입력과 공유하지 않습니다. 병합 결과 수정과 삭제가 원본에 영향을 주지 않아야 합니다.

## 선택 A. 영값도 구별하는 내용 비교

작업 파일: `_practice/06-maps/same-counts/main.go`.

`func sameCounts(a, b map[string]int) bool`를 작성합니다. 키와 값이 모두 같을 때만 true입니다. nil과 생성된 빈 맵은 같은 빈 항목 집합으로 취급합니다.

기본 확인 호출의 출력:

```text
빈 비교: true
같은 값: true
다른 키: false
다른 값: false
```

각 호출은 nil과 빈 맵, a:2와 a:2, a:0과 b:0, a:2와 a:3을 비교합니다.

```sh
go run ./_practice/06-maps/same-counts
```

<details>
<summary>힌트 1: 항목 수가 다르면</summary>

길이가 다르면 같은 항목 집합이 아닙니다. 길이가 같을 때도 키와 값을 확인해야 합니다.

</details>

<details>
<summary>힌트 2: 값만 비교하면 놓치는 경우</summary>

없는 키를 읽으면 0입니다. 값 0인 키를 비교할 때 상대 맵에 실제로 존재하는지도 검사하세요.

</details>

완료 조건: 입력을 변경하지 않고, 서로 다른 0 키를 같다고 판단하지 않습니다.

## 점검과 다음 단계

- [ ] nil 맵을 쓰기 전에 생성했다.
- [ ] 저장된 0과 없는 키를 구분했다.
- [ ] 정렬 순서와 첫 등장 순서를 구분했다.
- [ ] 실패한 출고 뒤 재고가 그대로인지 확인했다.
- [ ] 빈 병합 결과와 독립성을 확인했다.
- [ ] 기본 입력으로 복구하고 개인 파일을 포맷했다.

```sh
go fmt ./_practice/06-maps/tally
go run ./_practice/06-maps/tally
```

다른 문제도 폴더 이름을 바꾸어 실행합니다. 참고 테스트는 개인 파일까지 자동으로 검사하지 않습니다.

[해설](../../solutions/06-maps/README.md) · [본문](../../book/06-maps/README.md)
