# 07장 연습문제: 문자열과 유니코드

선수 학습: [07. 문자열과 유니코드](../../book/07-strings-unicode/README.md).

길이 제한의 단위는 문제에 적힌 **룬(코드 포인트)**입니다. 사용자가 보는 결합 글자 단위와 같다고 가정하지 않습니다. 문자열을 바이트로 자르는 경우와 룬으로 다루는 경우를 비교합니다.

기본 문제는 1~6이며 선택 A에서는 코드 포인트 순서를 뒤집습니다. 외부 모듈이나 파일 입력은 필요하지 않습니다.

| 문제 | 수준 | 개인 폴더 |
| --- | --- | --- |
| 1. 길이·위치 추적 | 확인 | `predict` |
| 2. 룬 제한 요약 | 변형 | `shorten` |
| 3. 바이트 자르기 수정 | 오류 진단 | `repair` |
| 4. 태그 정리 | 변형 | `clean-tags` |
| 5. 단어 집계 | 응용 | `word-counts` |
| 6. 학습용 이름 검증 | 응용 | `valid-name` |
| 선택 A. 룬 역순 | 선택 | `reverse-runes` |

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("predict", "shorten", "repair", "clean-tags", "word-counts", "valid-name", "reverse-runes")) {
    $practiceDir = Join-Path "_practice/07-strings-unicode" $practiceName
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
for practice_name in predict shorten repair clean-tags word-counts valid-name reverse-runes; do
    practice_dir="_practice/07-strings-unicode/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

한 폴더에는 하나의 `main`만 둡니다. 새 `go.mod`나 `go.work`는 만들지 않습니다. `_practice/`는 공통 `go test ./...`에 포함되지 않으므로 개인 폴더를 직접 지정해서 실행하세요. [공통 연습 안내](../../exercises/README.md)를 참고합니다.


## 문제 1. 바이트 길이와 룬 위치 — 확인

작업 파일: `_practice/07-strings-unicode/predict/main.go`.

아래 전체 코드의 출력 네 줄을 예상합니다. 이 예제에서 A는 1바이트, 한은 3바이트, 🙂는 4바이트입니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := "A한🙂"
	fmt.Println("길이:", len(text), utf8.RuneCountInString(text))
	for index, value := range text {
		fmt.Printf("%d: %U\n", index, value)
	}
}
```

1. len과 RuneCountInString은 각각 무엇을 세나요?
2. range의 인덱스가 0, 1 다음에 어떤 수가 되나요?
3. 두 번째 반복 변수의 타입은 무엇인가요?
4. `text[:2]`는 유효한 UTF-8일까요?

```sh
go run ./_practice/07-strings-unicode/predict
```

<details>
<summary>힌트 1: 시작 바이트 위치</summary>

앞선 룬들의 바이트 수를 더하면 다음 룬의 시작 위치입니다. 첫 변수는 순서 번호가 아닙니다.

</details>

<details>
<summary>힌트 2: 부분 범위의 단위</summary>

문자열 슬라이싱도 바이트 기준입니다. 끝 경계가 한의 세 바이트 중간에 있는지 확인하세요.

</details>

완료 조건: 길이와 위치를 표로 설명하고 range와 바이트 반복의 차이를 말합니다.

## 문제 2. 룬 개수로 요약하기 — 변형

작업 파일: `_practice/07-strings-unicode/shorten/main.go`.

`func shorten(text string, limit int) (string, bool)`를 작성합니다.

| 조건 | 결과 |
| --- | --- |
| 음수 limit 또는 잘못된 UTF-8 | 빈 문자열, false |
| 룬 수가 limit 이하 | 원문, true |
| 룬 수가 limit보다 많음 | 앞 limit룬 뒤에 `…` 추가, true |

limit은 **본문으로 남길 룬 수**입니다. 생략 기호 `…` 자체는 이 제한에 포함하지 않습니다. 비어 있지 않은 입력에 limit 0이면 `…`이며, 빈 입력에 limit 0이면 빈 문자열입니다.

기본 입력은 `"안녕하세요 Go"`, limit은 5입니다.

```text
요약: "안녕하세요…", 성공: true
```

```sh
go run ./_practice/07-strings-unicode/shorten
```

| 입력 | limit | 결과·성공 |
| --- | --- | --- |
| 빈 문자열 | 0 | 빈 문자열, true |
| 한글 | 1 | 한…, true |
| 한글 | 2 또는 8 | 한글, true |
| 한글 | 0 | …, true |
| Go | -1 | 빈 문자열, false |
| `"\xff"` | 0 | 빈 문자열, false |

<details>
<summary>힌트 1: 유효성 검사부터</summary>

잘못된 바이트를 rune 변환으로 대체한 뒤 성공 처리하지 않도록 ValidString을 먼저 확인합니다.

</details>

<details>
<summary>힌트 2: 비교하는 길이의 단위</summary>

[]rune의 길이와 limit을 비교합니다. 바이트 수가 아닌 룬 수를 기준으로 부분 범위를 만든 뒤 string으로 바꿉니다.

</details>

완료 조건: 잘못된 입력과 성공한 빈 문자열을 구분하고, 생략 기호가 붙는 조건을 정확히 설명합니다.

## 문제 3. 한글의 첫 룬을 남기기 — 오류 진단

작업 파일: `_practice/07-strings-unicode/repair/main.go`.

아래는 **컴파일되지만 잘못된 UTF-8 결과를 만드는 전체 코드**입니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := "한글"
	prefix := text[:1]
	fmt.Printf("앞 1룬: %q\n", prefix)
	fmt.Println("유효:", utf8.ValidString(prefix))
}
```

prefix를 만드는 줄만 수정합니다. 입력은 비어 있지 않은 유효한 UTF-8이라고 가정하며, 첫 코드 포인트 하나를 남겨야 합니다.

```text
앞 1룬: "한"
유효: true
```

```sh
go run ./_practice/07-strings-unicode/repair
```

입력을 Go, 🙂Go로 바꾸어 각각 G와 🙂가 나오는지 확인한 뒤 한글로 복구합니다. 빈 입력까지 다루는 일반 함수는 본문의 prefixRunes 계약을 참고합니다.

<details>
<summary>힌트 1: 현재 코드의 단위</summary>

text의 [:1]은 첫 룬이 아니라 첫 바이트입니다. 필요한 원소 단위의 슬라이스로 바꾸어 생각하세요.

</details>

<details>
<summary>힌트 2: 결과를 다시 문자열로</summary>

[]rune에서 범위를 고른 뒤 string으로 변환합니다. 원래 문자열 자체를 수정할 필요는 없습니다.

</details>

완료 조건: 기본 입력의 원래 첫 바이트와 수정한 첫 룬을 구분해 설명합니다.

## 문제 4. 태그 정리 — 변형

작업 파일: `_practice/07-strings-unicode/clean-tags/main.go`.

`func cleanTags(items []string) ([]string, bool)`를 작성합니다.

1. 각 입력이 유효한 UTF-8인지 검사합니다. 하나라도 잘못되면 `nil, false`입니다.
2. 양끝 유니코드 공백을 제거하고 소문자로 바꿉니다.
3. 빈 결과는 제외합니다.
4. 정리된 문자열을 기준으로 중복을 제거하고 첫 등장 순서를 유지합니다.
5. 입력을 바꾸지 않습니다. 결과 원소 수정도 입력에 영향을 주지 않아야 합니다.

유효한 빈 입력 또는 전부 빈 태그인 입력은 `nil, true`입니다. 태그 내부 공백은 이번 계약에서 유지합니다.

기본 입력 `[]string{" Go ", "한글", "GO", "\t", " Map "}`:

```text
태그: ["go" "한글" "map"], 성공: true
원본: [" Go " "한글" "GO" "\t" " Map "]
```

```sh
go run ./_practice/07-strings-unicode/clean-tags
```

| 입력 | 결과·성공 |
| --- | --- |
| nil | nil, true |
| 공백·탭만 있는 두 항목 | nil, true |
| GO, go, 양끝 공백을 가진 한글 | go, 한글; true |
| 정상 항목 뒤 잘못된 바이트 항목 | nil, false |

<details>
<summary>힌트 1: 중복 기준은 정리한 값</summary>

원문을 seen에 넣으면 Go와 GO가 다른 키가 됩니다. 비교용 태그를 만든 뒤 이미 보았는지 확인하세요.

</details>

<details>
<summary>힌트 2: 실패 시 부분 결과를 돌려주지 않기</summary>

이 계약은 전체 성공 또는 실패입니다. 앞의 일부 태그가 정상이었어도 잘못된 입력을 만나면 nil과 false를 반환합니다.

</details>

완료 조건: 정리 순서, 중복 기준, 빈 결과와 실패의 차이를 설명합니다.

## 문제 5. 유효한 텍스트의 단어 집계 — 응용

작업 파일: `_practice/07-strings-unicode/word-counts/main.go`.

`func wordCounts(text string) (map[string]int, bool)`를 작성합니다. 잘못된 UTF-8은 `nil, false`로 거부합니다. 유효하면 소문자로 바꾸고 Fields로 분리한 항목별 횟수를 반환합니다.

문장 부호는 제거하지 않습니다. 빈 문자열·공백뿐인 입력에는 생성된 빈 맵과 true를 반환합니다. main은 실패 시 `텍스트 확인` 한 줄만 출력하고 성공하면 키를 정렬합니다.

기본 입력 `" Go\t한글 GO\n한글 "`의 출력:

```text
"go": 2
"한글": 2
```

```sh
go run ./_practice/07-strings-unicode/word-counts
```

| 입력 | 결과 |
| --- | --- |
| 공백·탭·줄바꿈 | 생성된 빈 맵, true |
| `GO go,` | go:1, go,:1 |
| `한글\u3000한글` | 한글:2 |
| 정상 문자와 `\xff` | nil, false |

<details>
<summary>힌트 1: 입력을 처리하기 전에</summary>

원문 UTF-8을 검증한 뒤 소문자 변환과 분리를 합니다. 변환된 결과가 유효해 보이는 것만 확인하면 원본의 잘못된 바이트를 놓칠 수 있습니다.

</details>

<details>
<summary>힌트 2: 06장의 집계와 정렬</summary>

분리한 []string을 순회하며 맵 값을 증가시킵니다. 계산 함수와 정렬·출력을 분리하면 실패 처리도 명확해집니다.

</details>

완료 조건: 쉼표를 포함한 항목을 별도 키로 유지하고 유니코드 공백을 처리합니다.

## 문제 6. 학습용 이름 규칙 검사 — 응용

작업 파일: `_practice/07-strings-unicode/valid-name/main.go`.

`func validName(text string) bool`을 작성합니다. 다음은 이 연습에서 정한 규칙이며 모든 서비스의 이름 정책을 뜻하지 않습니다.

- 유효한 UTF-8이어야 합니다.
- 길이는 2~8룬입니다.
- 첫 룬은 unicode.IsLetter가 참이어야 합니다.
- 나머지는 IsLetter 또는 IsDigit이 참이거나 밑줄 `_`이어야 합니다.
- 공백 제거, 소문자 변환, 유니코드 정규화를 자동으로 하지 않습니다.

기본 출력:

```text
한글_2: true
2Go: false
Go 학습: false
```

```sh
go run ./_practice/07-strings-unicode/valid-name
```

| 입력 | 결과 |
| --- | --- |
| 빈 문자열·a | false |
| a1·abcdefgh | true |
| abcdefghi | false |
| a-b·Go 학습 | false |
| a１ | true |
| `e\u0301` | false |
| é1 | true |
| 잘못된 UTF-8 | false |

전각 숫자 １도 유니코드 십진 숫자입니다. 결합 악센트는 IsLetter·IsDigit이 아니므로 이 계약에서 허용되지 않습니다.

<details>
<summary>힌트 1: 첫 원소를 읽기 전 검사</summary>

먼저 UTF-8과 길이를 검사합니다. 빈 문자열을 []rune으로 바꿔도 첫 원소는 없으므로 무조건 runes[0]을 읽으면 안 됩니다.

</details>

<details>
<summary>힌트 2: 첫 위치와 나머지 위치</summary>

첫 룬의 조건을 검사한 뒤 나머지를 순회합니다. 허용 조건 셋이 모두 거짓인 경우 false입니다.

</details>

완료 조건: 최소·최대 길이와 첫 위치 규칙을 각각 확인하고, 코드 포인트 기준임을 설명합니다.

## 선택 A. 코드 포인트 순서 뒤집기

작업 파일: `_practice/07-strings-unicode/reverse-runes/main.go`.

`func reverseRunes(text string) (string, bool)`을 작성합니다. 유효한 UTF-8은 룬의 순서를 뒤집고 true, 잘못된 UTF-8은 빈 문자열과 false입니다. 빈 문자열은 정상 입력입니다.

기본 입력 Go한글:

```text
원본: Go한글
룬 역순: "글한oG", 성공: true
```

```sh
go run ./_practice/07-strings-unicode/reverse-runes
```

빈 문자열, 한 룬, 🙂Go를 확인합니다. `e\u0301`을 뒤집으면 결합 부호의 위치도 뒤집힙니다. 이 함수는 사용자가 보는 글자 묶음의 역순을 보장하지 않습니다.

<details>
<summary>힌트 1: 05장의 역순 처리</summary>

문자열을 []rune으로 바꾼 뒤 양 끝을 교환하거나 별도 결과에 역순으로 복사합니다.

</details>

<details>
<summary>힌트 2: 원문 유지</summary>

수정한 룬 슬라이스로 새 string을 만들면 됩니다. 원본 string 내부에 대입하는 문법은 사용할 수 없습니다.

</details>

완료 조건: 유효한 입력을 두 번 뒤집으면 원래 코드 포인트열이 되며, 원문이 유지됩니다.

## 해설 전 점검

- [ ] 바이트·룬의 단위를 요구사항에 맞췄다.
- [ ] UTF-8 검사와 룬 개수 검사를 구분했다.
- [ ] 빈 입력과 실패 결과를 구분했다.
- [ ] 공백·대소문자·문장 부호 정책을 지켰다.
- [ ] 결과 원소 수정이 원본에 미치는 영향을 확인했다.
- [ ] 결합 부호와 전각 숫자 사례를 실행했다.

```sh
go fmt ./_practice/07-strings-unicode/shorten
go run ./_practice/07-strings-unicode/shorten
```

다른 문제도 폴더를 바꾸어 실행합니다. 참고 구현 테스트는 개인 코드를 자동 검사하지 않습니다.

[해설](../../solutions/07-strings-unicode/README.md) · [본문](../../book/07-strings-unicode/README.md)
