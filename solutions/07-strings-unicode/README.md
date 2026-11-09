# 07장 해설: 문자열과 유니코드

[문제](../../exercises/07-strings-unicode/README.md)를 먼저 풀고 비교합니다. 모든 명령은 저장소 루트 기준입니다. “한 글자”라는 표현 대신 실제 코드가 바이트·룬 중 무엇을 다루는지 확인합니다.

문제 2·4·5·6은 참고 패키지와 테스트가 있습니다. 문제 1·3·선택 A는 아래 전체 코드를 개인 폴더에서 실행합니다. 기존 개인 파일은 보존하세요.

## 문제 1. 길이와 시작 위치

예상 출력:

```text
길이: 8 3
0: U+0041
1: U+D55C
4: U+1F642
```

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

```sh
go run ./_practice/07-strings-unicode/predict
```

| 룬 | 바이트 수 | 시작 위치 |
| --- | --- | --- |
| A | 1 | 0 |
| 한 | 3 | 1 |
| 🙂 | 4 | 4 |

바이트 합계는 8이고 코드 포인트는 3개입니다. range의 두 번째 값은 rune입니다. `text[:2]`는 A와 한의 첫 바이트만 포함하므로 유효한 UTF-8이 아닙니다.

첫 반복 변수는 디코딩한 룬의 번호가 아니라 바이트 오프셋입니다. [range](https://go.dev/ref/spec#For_range)와 [utf8](https://pkg.go.dev/unicode/utf8)을 참고합니다.

## 문제 2. 룬 제한 요약

참고 구현: [shorten/main.go](shorten/main.go) · [테스트](shorten/main_test.go).

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func shorten(text string, limit int) (string, bool) {
	if limit < 0 || !utf8.ValidString(text) {
		return "", false
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text, true
	}
	return string(runes[:limit]) + "…", true
}
func main() {
	result, ok := shorten("안녕하세요 Go", 5)
	fmt.Printf("요약: %q, 성공: %t\n", result, ok)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/07-strings-unicode/shorten
go test ./solutions/07-strings-unicode/shorten
```

예상 출력:

```text
요약: "안녕하세요…", 성공: true
```


### 검사와 자르기의 순서

음수 제한과 잘못된 UTF-8을 먼저 거부합니다. 그다음 []rune의 길이를 비교하고, 잘라야 할 때만 생략 기호를 추가합니다.

| 입력·제한 | 선택한 경로 |
| --- | --- |
| 빈 문자열, 0 | 길이 이하이므로 원문 반환 |
| 한글, 2 | 정확히 같으므로 생략 기호 없음 |
| 한글, 1 | 첫 룬 뒤에 생략 기호 |
| 한글, 0 | 본문은 빈 범위, 기호만 추가 |
| 잘못된 바이트, 0 | 유효성 검사에서 실패 |

limit에 생략 기호를 포함하지 않는 것이 문제의 계약입니다. 따라서 줄인 결과의 룬 수는 보통 limit+1입니다. 전체 표시 길이까지 제한하는 함수라면 다른 규칙을 정해야 합니다.

`e\u0301`을 1룬으로 줄이면 e…가 됩니다. 결합 악센트까지 유지하는 함수는 아닙니다. 룬 단위라는 계약의 한계를 테스트에서도 드러냅니다.

```sh
go run ./_practice/07-strings-unicode/shorten
```

## 문제 3. 바이트 범위에서 룬 범위로

원래 출력은 다음과 같습니다.

```text
앞 1룬: "\xed"
유효: false
```

0xed는 한을 인코딩한 세 바이트 중 첫 바이트입니다. 문자열 부분 범위를 룬 부분 범위로 바꿉니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := "한글"
	prefix := string([]rune(text)[:1])
	fmt.Printf("앞 1룬: %q\n", prefix)
	fmt.Println("유효:", utf8.ValidString(prefix))
}
```

```sh
go run ./_practice/07-strings-unicode/repair
```

수정 후 출력:

```text
앞 1룬: "한"
유효: true
```

| 입력 | 첫 룬 |
| --- | --- |
| 한글 | 한 |
| Go | G |
| 🙂Go | 🙂 |

이번 문제는 비어 있지 않은 유효한 입력만 다룹니다. 일반 함수로 만들 때는 유효성·길이를 검사해야 합니다. 단순한 한 줄 수정이 모든 입력에 대한 안전한 함수가 된다고 일반화하지 않습니다.

## 문제 4. 태그 정리

참고 구현: [clean-tags/main.go](clean-tags/main.go) · [테스트](clean-tags/main_test.go).

```go
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func cleanTags(items []string) ([]string, bool) {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !utf8.ValidString(item) {
			return nil, false
		}
		tag := strings.ToLower(strings.TrimSpace(item))
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		result = append(result, tag)
	}
	return result, true
}
func main() {
	tags := []string{" Go ", "한글", "GO", "\t", " Map "}
	result, ok := cleanTags(tags)
	fmt.Printf("태그: %q, 성공: %t\n", result, ok)
	fmt.Printf("원본: %q\n", tags)
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/07-strings-unicode/clean-tags
go test ./solutions/07-strings-unicode/clean-tags
```

예상 출력:

```text
태그: ["go" "한글" "map"], 성공: true
원본: [" Go " "한글" "GO" "\t" " Map "]
```


### 처리 순서

UTF-8을 검사한 원문에서 양끝 공백을 제거하고 소문자 키를 만듭니다. 그 키가 비었거나 이미 있으면 건너뜁니다. 처음 보았다면 seen과 결과에 기록합니다.

| 원문 | 정리한 키 | 처리 |
| --- | --- | --- |
| 앞뒤 공백이 있는 Go | go | 추가 |
| 한글 | 한글 | 추가 |
| GO | go | 중복 제외 |
| 탭 | 빈 문자열 | 제외 |
| 앞뒤 공백이 있는 Map | map | 추가 |

중복 검사를 원문으로 하면 Go와 GO를 다르게 취급합니다. 공백 정리와 대소문자 처리 뒤의 키를 기준으로 하는 것이 요구사항입니다.

### 빈 결과와 실패

nil 입력이나 전부 공백인 태그는 `nil, true`입니다. 중간에 잘못된 UTF-8이 있으면 앞에서 일부 결과를 만들었어도 `nil, false`입니다. 성공 여부를 보지 않고 결과 길이만 확인하면 둘을 구분할 수 없습니다.

원본 items에 대입하지 않고 결과를 append하므로 원본 원소 자리가 유지됩니다. 결과의 첫 문자열을 바꾸어도 입력은 바뀌지 않습니다.

내부 공백은 제거하지 않습니다. 태그가 `Go 학습`이면 결과는 `go 학습`입니다. 여러 단어를 한 태그로 허용할지는 별도 계약이며 이 문제에서는 그대로 유지합니다.

```sh
go run ./_practice/07-strings-unicode/clean-tags
```

## 문제 5. 단어 집계

참고 구현: [word-counts/main.go](word-counts/main.go) · [테스트](word-counts/main_test.go).

```go
package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

func wordCounts(text string) (map[string]int, bool) {
	if !utf8.ValidString(text) {
		return nil, false
	}
	counts := make(map[string]int)
	for _, word := range strings.Fields(strings.ToLower(text)) {
		counts[word]++
	}
	return counts, true
}
func main() {
	counts, ok := wordCounts(" Go\t한글 GO\n한글 ")
	if !ok {
		fmt.Println("텍스트 확인")
		return
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%q: %d\n", key, counts[key])
	}
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/07-strings-unicode/word-counts
go test ./solutions/07-strings-unicode/word-counts
```

예상 출력:

```text
"go": 2
"한글": 2
```


### 실패를 먼저 처리하기

원문이 유효한지 확인한 뒤 변환과 집계를 수행합니다. 잘못된 바이트를 ToLower 등의 처리에서 대체한 후 검사하면 원문의 잘못된 인코딩을 놓칠 수 있으므로 순서를 지킵니다.

성공한 빈 입력은 생성된 빈 맵입니다. 실패한 입력은 nil입니다. 그래도 호출자는 결과의 nil 여부에 의존하기보다 반환된 ok를 검사합니다.

| 입력 | 키별 결과 |
| --- | --- |
| 공백뿐 | 항목 없음 |
| GO go, | go:1, go,:1 |
| 한글과 U+3000, 한글 | 한글:2 |
| 잘못된 바이트 포함 | 실패 |

Fields는 유니코드 공백으로 분리합니다. 쉼표는 남습니다. 이 함수를 문장 부호를 제거하는 단어 분석기처럼 설명하면 실제 동작과 달라집니다. [Fields](https://pkg.go.dev/strings#Fields)의 계약을 참고합니다.

sort.Strings는 출력 순서를 고정합니다. map을 직접 range하며 줄을 출력하면 같은 입력에서도 줄 순서가 고정되지 않습니다.

```sh
go run ./_practice/07-strings-unicode/word-counts
```

## 문제 6. 이름 검증

참고 구현: [valid-name/main.go](valid-name/main.go) · [테스트](valid-name/main_test.go).

```go
package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

func validName(text string) bool {
	if !utf8.ValidString(text) {
		return false
	}
	runes := []rune(text)
	if len(runes) < 2 || len(runes) > 8 {
		return false
	}
	if !unicode.IsLetter(runes[0]) {
		return false
	}
	for _, value := range runes[1:] {
		if !unicode.IsLetter(value) && !unicode.IsDigit(value) && value != '_' {
			return false
		}
	}
	return true
}
func main() {
	fmt.Println("한글_2:", validName("한글_2"))
	fmt.Println("2Go:", validName("2Go"))
	fmt.Println("Go 학습:", validName("Go 학습"))
}
```

저장소 루트에서 참고 구현을 실행합니다.

```sh
go run ./solutions/07-strings-unicode/valid-name
go test ./solutions/07-strings-unicode/valid-name
```

예상 출력:

```text
한글_2: true
2Go: false
Go 학습: false
```


### 첫 룬을 읽기 전에

먼저 UTF-8을 검사하고 룬 슬라이스의 길이를 확인합니다. 최소 길이가 2라는 조건을 통과했으므로 그 뒤의 runes[0]과 runes[1:] 접근이 가능합니다.

첫 룬에는 글자만 허용합니다. 나머지는 글자·십진 숫자·밑줄 중 하나이면 됩니다. 세 조건이 모두 거짓일 때 실패합니다.

| 입력 | 검사 결과의 이유 |
| --- | --- |
| a | 최소 길이 미달 |
| a1 | 글자로 시작, 숫자 허용 |
| abcdefgh | 최대 8룬 |
| abcdefghi | 길이 초과 |
| 2Go | 첫 룬이 글자가 아님 |
| Go 학습 | 공백이 허용되지 않음 |
| a１ | 전각 숫자도 십진 숫자 |
| e와 결합 악센트 | 결합 부호가 허용 분류 아님 |
| é1 | 첫 룬 글자, 다음 숫자 |

IsDigit은 ASCII 범위만 검사하지 않습니다. 이 조건을 원하지 않는 계약이라면 `'0' <= value && value <= '9'`처럼 ASCII 숫자 범위를 명시할 수 있습니다. 어떤 규칙이 맞는지는 요구사항에 달려 있습니다.

이 함수는 등록 가능성이나 중복 이름을 확인하지 않습니다. 입력 형식만 판단하고, 대소문자·공백을 자동 수정하지도 않습니다.

```sh
go run ./_practice/07-strings-unicode/valid-name
```

## 선택 A. 룬 순서 뒤집기

개인 파일 `_practice/07-strings-unicode/reverse-runes/main.go`에서 실행합니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func reverseRunes(text string) (string, bool) {
	if !utf8.ValidString(text) {
		return "", false
	}
	runes := []rune(text)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		runes[left], runes[right] = runes[right], runes[left]
	}
	return string(runes), true
}

func main() {
	text := "Go한글"
	result, ok := reverseRunes(text)
	fmt.Println("원본:", text)
	fmt.Printf("룬 역순: %q, 성공: %t\n", result, ok)
}
```

```sh
go run ./_practice/07-strings-unicode/reverse-runes
```

예상 출력:

```text
원본: Go한글
룬 역순: "글한oG", 성공: true
```

왼쪽·오른쪽 위치를 동시에 이동하며 양 끝을 교환합니다. 다중 대입의 오른쪽 값들을 먼저 읽으므로 임시 변수를 따로 만들지 않아도 두 값을 교환할 수 있습니다.

| 입력 룬 | 교환 뒤 |
| --- | --- |
| G, o, 한, 글 | 글, 한, o, G |
| 한 | 한 |
| 빈 입력 | 빈 입력 |

빈 입력에서는 right가 -1이고 left<right가 거짓이므로 원소에 접근하지 않습니다. 한 룬일 때도 교환할 필요가 없습니다.

결합 부호를 포함한 입력은 부호의 위치까지 뒤집습니다. 이 함수는 **코드 포인트열의 역순**이며 표시 글자 묶음의 역순을 보장하지 않습니다. 이 한계를 함수 이름과 설명에 드러내는 것이 중요합니다.

## 검사 범위

```sh
go test -count=1 ./solutions/07-strings-unicode/...
```

| 참고 구현 | 경계 검사 |
| --- | --- |
| shorten | 빈·0·음수·정확한 길이·초과, 한글·이모지·결합 부호, 잘못된 UTF-8 |
| clean-tags | 빈·공백·대소문자·유니코드 공백, 실패 시 부분 결과 제거, 입력 보존 |
| word-counts | 빈 결과, 공백 정의, 문장 부호 유지, 잘못된 원문 거부 |
| valid-name | 1·2·8·9룬, 첫 문자, 숫자 분류, 공백·부호·잘못된 인코딩 |

각 Example_main은 기본 출력도 검사합니다. 개인 파일은 문제별 명령과 입력 표로 별도 실행해야 합니다.

## 다시 설명해 볼 질문

1. 입력 검증을 문자열 변환 전에 하는 이유는 무엇인가요?
2. 결과 길이 제한에 생략 기호를 포함할지 어떻게 계약으로 드러내나요?
3. 빈 문자열, 공백뿐인 문자열, 잘못된 바이트열은 같은 입력인가요?
4. 대소문자 통일과 유니코드 정규화는 어떤 차이가 있나요?
5. 룬 기반 구현이 사용자가 보는 모든 글자 단위를 보존하지 못하는 이유는 무엇인가요?

[본문](../../book/07-strings-unicode/README.md) · [문제](../../exercises/07-strings-unicode/README.md)
