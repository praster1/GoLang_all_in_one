# 07. 문자열과 유니코드

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 입문 | [06. 맵과 집계](../06-maps/README.md) | Go, 편집기, 터미널, UTF-8 소스 파일 |

지금까지 문자열을 이름과 출력 문구로 사용했습니다. 이제 길이를 제한하거나 일부를 자르고 단어를 집계하려면 “문자 하나”가 무엇인지 정해야 합니다. `Go`와 `한글`은 화면에서 각각 두 글자로 보이지만 Go의 len 결과는 다릅니다.

이번 장은 **바이트(byte)**, **룬(rune)**, 사람이 보는 글자 단위를 구분합니다. 한글을 깨뜨리지 않고 처리하는 방법과, 잘못된 UTF-8 입력을 어떻게 다룰지 함수의 계약으로 정하는 방법을 배웁니다.

## 학습 목표

- string의 길이·인덱스와 range가 다루는 단위를 구분한다.
- []byte·[]rune 변환으로 필요한 단위의 새 값을 만든다.
- 룬 단위 자르기와 UTF-8 유효성 검사를 구현한다.
- strings로 공백·분리·검색·대소문자를 처리한다.
- 텍스트 집계와 이름 검증에서 처리 규칙과 한계를 명시한다.

## 사전 지식과 준비

문자열 리터럴, 정수, 함수, 슬라이스, 맵과 키 정렬을 사용합니다. strings.Builder는 선택 학습이며 구조체·메서드의 전체 설명은 08·09장에서 이어갑니다. 기본 문제에는 Builder를 사용할 필요가 없습니다.

모든 명령은 **저장소 루트**에서 실행합니다. 개인 파일은 11절에서 준비하는 `_practice/07-strings-unicode/lab/main.go`를 사용합니다. 전체 코드는 실제 예제 파일과 같습니다. 발췌를 넣을 때 필요한 import도 함께 확인합니다.

| 예제 | 핵심 | 실행 명령 |
| --- | --- | --- |
| bytes-runes | 길이와 range 위치 | `go run ./examples/07-strings-unicode/bytes-runes` |
| conversion | 바이트·룬 슬라이스 변환 | `go run ./examples/07-strings-unicode/conversion` |
| strings-ops | 공백·분리·검색 | `go run ./examples/07-strings-unicode/strings-ops` |
| rune-prefix | 룬 개수로 앞부분 선택 | `go run ./examples/07-strings-unicode/rune-prefix` |
| utf8-validation | 잘못된 인코딩과 대체 문자 | `go run ./examples/07-strings-unicode/utf8-validation` |
| word-count | 공백 분리와 소문자 집계 | `go run ./examples/07-strings-unicode/word-count` |
| classification | 유니코드 분류 | `go run ./examples/07-strings-unicode/classification` |

## 1. string은 변경할 수 없는 바이트열

Go의 string 값은 바이트들의 연속입니다. 값 자체의 내용을 한 위치씩 바꾸는 것은 허용되지 않습니다. 문자열 변수에 다른 string을 대입하는 것과 문자열 내부 바이트를 수정하는 것은 다릅니다. [문자열 타입](https://go.dev/ref/spec#String_types)을 참고합니다.

다음은 main 안의 독립된 발췌입니다. import는 fmt입니다.

```go
text := "Go"
text = text + " 학습"
fmt.Println(text)
```

출력은 `Go 학습`입니다. 기존 변수에 새로운 문자열 값을 대입했습니다. `text[0] = 'N'`처럼 문자열 내부를 직접 바꾸는 코드는 컴파일되지 않습니다.

### 1.1 소스 문자와 문자열 값

Go 소스는 UTF-8로 작성합니다. 따옴표 안에 직접 쓴 한글은 UTF-8 바이트로 표현됩니다. 그러나 string 값에 들어 있는 모든 바이트열이 반드시 유효한 UTF-8이라는 뜻은 아닙니다.

| 표기 | 뜻 |
| --- | --- |
| `"한글"` | 소스에 직접 적은 텍스트 |
| `"Go\n학습"` | 줄바꿈 이스케이프를 포함 |
| `"\uD55C"` | 코드 포인트 U+D55C의 한글 한 |
| `"\x41"` | 값 0x41인 바이트 |
| `"\xff"` | UTF-8로 유효하지 않은 바이트를 담을 수 있음 |

`\xNN`은 바이트 값, `\uNNNN`은 유니코드 코드 포인트를 나타내는 표기입니다. 역따옴표의 원시 문자열에서는 일반적인 역슬래시 이스케이프를 처리하지 않습니다. [문자열 리터럴](https://go.dev/ref/spec#String_literals)을 참고합니다.

소스 파일의 인코딩과 실행 중 문자열 데이터의 유효성은 별도로 확인해야 합니다. 외부 데이터 읽기는 후속 장에서 배우지만, 이번 장에서는 작은 바이트열로 이 차이를 재현합니다.

## 2. 바이트, 코드 포인트, 룬

**유니코드 코드 포인트(code point)**는 문자 등에 부여된 번호입니다. **UTF-8**은 그 번호를 바이트열로 인코딩하는 방식입니다. Go에서 코드 포인트를 다루는 정수 타입 이름은 rune이며 int32의 별칭입니다. byte는 uint8의 별칭입니다.

이름만 바뀐 정수 타입이므로 rune 변수가 자동으로 항상 유효한 코드 포인트만 저장하도록 제한되지는 않습니다. 다만 문자열 range가 유효한 UTF-8을 읽으면 해당 코드 포인트 값을 rune으로 제공합니다.

| 단위 | Go에서 자주 쓰는 표현 | 질문 |
| --- | --- | --- |
| 바이트 | byte, []byte, len(string) | 저장된 바이트가 몇 개인가? |
| 코드 포인트 | rune, []rune | 어떤 유니코드 값들이 있는가? |
| 사용자가 보는 글자 단위 | 별도 규칙 필요 | 화면에서 한 덩어리로 보는가? |

마지막 단위가 룬 하나와 항상 같지는 않습니다. 결합 문자와 일부 이모지 조합은 여러 코드 포인트로 한 덩어리를 나타냅니다. 먼저 앞의 두 단위를 구분하고 8절에서 그 한계를 확인합니다.

### 2.1 Go한글의 위치를 따라가기

저장 경로: [`examples/07-strings-unicode/bytes-runes/main.go`](../../examples/07-strings-unicode/bytes-runes/main.go). 전체 코드입니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := "Go한글"
	fmt.Printf("바이트: %d, 룬: %d\n", len(text), utf8.RuneCountInString(text))
	fmt.Printf("바이트 값: %x\n", text)
	for index, value := range text {
		fmt.Printf("%d: %c (%U)\n", index, value, value)
	}
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/07-strings-unicode/bytes-runes
```

예상 출력:

```text
바이트: 8, 룬: 4
바이트 값: 476fed959ceab880
0: G (U+0047)
1: o (U+006F)
2: 한 (U+D55C)
5: 글 (U+AE00)
```

len은 8바이트이고 RuneCountInString은 4룬입니다. G와 o는 각각 한 바이트, 이 예제의 한과 글은 각각 세 바이트입니다. 모든 한글 관련 텍스트나 모든 유니코드 값이 무조건 세 바이트라는 일반 규칙은 아닙니다.

| 룬 | 코드 포인트 | 시작 바이트 위치 | 바이트 수 |
| --- | --- | --- | --- |
| G | U+0047 | 0 | 1 |
| o | U+006F | 1 | 1 |
| 한 | U+D55C | 2 | 3 |
| 글 | U+AE00 | 5 | 3 |

range의 첫 변수는 **바이트 위치**입니다. 룬을 차례로 읽더라도 인덱스가 0, 1, 2, 3으로만 증가하지 않는 이유입니다. 두 번째 변수는 디코딩한 rune 값입니다.

근거는 [문자열 range](https://go.dev/ref/spec#For_range), [RuneCountInString](https://pkg.go.dev/unicode/utf8#RuneCountInString), [공식 문자열 설명](https://go.dev/blog/strings)에 있습니다.


### 2.2 출력 서식을 단위에 맞게 고르기

| 서식 | 읽는 목적 | 예 |
| --- | --- | --- |
| `%s` | 문자열을 텍스트로 표시 | Go한글 |
| `%q` | 따옴표와 이스케이프로 경계 표시 | `"Go\n학습"` |
| `%c` | 정수를 코드 포인트 문자로 표시 | 한 |
| `%U` | 코드 포인트 번호 | U+D55C |
| `%x` | string의 바이트를 16진수로 표시 | 476f… |

화면에 공백이나 잘못된 바이트가 섞여 있으면 `%q` 또는 `%x`가 원인을 찾는 데 도움이 됩니다. 출력 방법이 문자열 자체를 정리하거나 고치는 것은 아닙니다. [fmt 출력 서식](https://pkg.go.dev/fmt#hdr-Printing)을 참고합니다.

## 3. 인덱스와 부분 범위는 바이트 기준

`text[i]`는 i번째 **바이트**를 반환합니다. `text[low:high]`도 바이트 범위입니다. UTF-8 코드 포인트의 중간을 경계로 선택할 수 있으므로 자른 결과가 깨진 UTF-8일 수 있습니다.

다음 전체 프로그램을 개인 파일에서 실행합니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := "Go한글"
	fmt.Printf("첫 바이트: %d\n", text[0])
	fmt.Printf("한의 첫 바이트: %x\n", text[2])
	broken := text[:3]
	fmt.Printf("앞 3바이트: %x, 유효: %t\n", broken, utf8.ValidString(broken))
	fmt.Println("앞 5바이트:", text[:5])
}
```

```sh
go run ./_practice/07-strings-unicode/lab
```

예상 출력:

```text
첫 바이트: 71
한의 첫 바이트: ed
앞 3바이트: 476fed, 유효: false
앞 5바이트: Go한
```

인덱스 2의 바이트 0xed 하나는 한이라는 코드 포인트 전체가 아닙니다. 앞 3바이트는 G, o, 한의 첫 바이트만 포함합니다. 앞 5바이트는 한의 세 바이트까지 포함하므로 유효합니다.

문자열 슬라이싱의 끝은 len 범위 안이어야 합니다. 슬라이스의 용량까지 확장하는 규칙을 문자열에 적용하지 않습니다. string에 cap을 사용할 수도 없습니다. [인덱스](https://go.dev/ref/spec#Index_expressions)와 [슬라이스 표현식](https://go.dev/ref/spec#Slice_expressions)을 참고합니다.

### 3.1 바이트 순회가 필요한 경우

```go
text := "한"
for i := 0; i < len(text); i++ {
	fmt.Printf("%d: %02x\n", i, text[i])
}
```

main 안의 위 발췌는 다음을 출력합니다.

```text
0: ed
1: 95
2: 9c
```

파일 형식이나 바이트 프로토콜처럼 바이트 위치가 필요한 작업도 있습니다. 모든 문제를 rune으로 바꿀 필요는 없습니다. 질문이 바이트 단위인지, 코드 포인트 단위인지 먼저 결정합니다.

## 4. []byte와 []rune으로 변환하기

수정 가능한 원소열이 필요하면 문자열을 슬라이스로 변환하고, 처리 후 다시 string으로 만듭니다. []byte는 바이트별 원소, []rune은 디코딩한 코드 포인트별 원소입니다.

### 4.1 원본 문자열과 수정 결과

저장 경로: [`examples/07-strings-unicode/conversion/main.go`](../../examples/07-strings-unicode/conversion/main.go). 전체 코드입니다.

```go
package main

import "fmt"

func main() {
	text := "Go한글"
	bytes := []byte(text)
	bytes[0] = 'N'
	runes := []rune(text)
	runes[2] = '말'
	fmt.Println("원본:", text)
	fmt.Println("바이트 수정:", string(bytes))
	fmt.Println("룬 수정:", string(runes))
	fmt.Printf("두 길이: %d, %d\n", len(bytes), len(runes))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/07-strings-unicode/conversion
```

예상 출력:

```text
원본: Go한글
바이트 수정: No한글
룬 수정: Go말글
두 길이: 8, 4
```

bytes의 첫 바이트를 N으로 바꾸어도 원본 text는 Go한글입니다. runes의 위치 2는 한이라는 룬이므로 말로 교체하면 Go말글이 됩니다.

| 표현 | 원소 수 | 위치 2의 의미 |
| --- | --- | --- |
| `[]byte(text)` | 8 | 한의 첫 바이트 |
| `[]rune(text)` | 4 | 한이라는 룬 |

슬라이스를 수정해도 이미 존재하던 string 값이 바뀌지 않는다는 의미를 확인합니다. 메모리 최적화의 구체적인 복사 횟수까지 가정할 필요는 없습니다. [문자열 변환](https://go.dev/ref/spec#Conversions_to_and_from_a_string_type)을 참고합니다.

변환 후 사용하는 인덱스도 해당 원소열의 단위입니다. []rune으로 바꾼 뒤의 인덱스는 룬 위치이지만, 원래 string의 인덱스가 바뀐 것은 아닙니다.


### 4.2 정수를 문자열로 바꾼다는 말의 두 뜻

정수 65를 코드 포인트로 인코딩하면 A입니다. 숫자의 십진 표기 `"65"`를 만들려면 strconv.Itoa를 사용합니다.

다음 전체 프로그램은 개인 파일에서 실행합니다.

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	code := rune(65)
	fmt.Printf("코드 포인트: %q\n", string(code))
	fmt.Printf("십진 숫자: %q\n", strconv.Itoa(65))
}
```

```sh
go run ./_practice/07-strings-unicode/lab
```

예상 출력:

```text
코드 포인트: "A"
십진 숫자: "65"
```

string 변환은 모든 타입의 일반적인 출력 변환 함수가 아닙니다. 숫자 표시가 목적이면 [strconv.Itoa](https://pkg.go.dev/strconv#Itoa)나 fmt의 서식을 사용합니다. 문자열에서 정수를 읽어 오류를 처리하는 작업은 11장에서 이어갑니다.

## 5. 룬 개수로 앞부분 자르기

한글이 들어 있는 제목에서 앞 몇 코드 포인트를 선택하려면 []rune을 기준으로 범위를 계산할 수 있습니다. 입력이 UTF-8인지 검사하는 함수를 먼저 소개합니다.

`utf8.ValidString(text)`는 유효한 UTF-8이면 true입니다. 이번 함수는 잘못된 입력이나 음수 개수를 false로 반환합니다. 성공한 빈 문자열과 실패를 구분하기 위해 bool을 함께 반환합니다.

### 5.1 유효성 확인 뒤 룬 범위 선택

저장 경로: [`examples/07-strings-unicode/rune-prefix/main.go`](../../examples/07-strings-unicode/rune-prefix/main.go). 전체 코드입니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func prefixRunes(text string, count int) (string, bool) {
	if count < 0 || !utf8.ValidString(text) {
		return "", false
	}
	runes := []rune(text)
	if count >= len(runes) {
		return text, true
	}
	return string(runes[:count]), true
}
func main() {
	prefix, ok := prefixRunes("Go한글", 3)
	fmt.Printf("앞 3룬: %q, 성공: %t\n", prefix, ok)
	empty, ok := prefixRunes("한글", 0)
	fmt.Printf("앞 0룬: %q, 성공: %t\n", empty, ok)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/07-strings-unicode/rune-prefix
```

예상 출력:

```text
앞 3룬: "Go한", 성공: true
앞 0룬: "", 성공: true
```

먼저 count가 음수인지, 입력이 유효한지 확인합니다. 그 다음 []rune으로 변환하므로 이 함수는 잘못된 바이트를 대체 문자로 고쳐서 받아들이는 방식이 아닙니다.

count가 원소 수 이상이면 전체 문자열을 반환합니다. 더 짧게 요청했다면 runes[:count]를 새 문자열로 만듭니다. count가 0이면 성공한 빈 결과입니다.

| 입력 | 개수 | 결과 | 성공 |
| --- | --- | --- | --- |
| Go한글 | 3 | Go한 | true |
| 한글 | 1 | 한 | true |
| 한글 | 0 | 빈 문자열 | true |
| 한글 | 9 | 한글 | true |
| 한글 | -1 | 빈 문자열 | false |
| 잘못된 바이트열 | 0 | 빈 문자열 | false |

0개 요청이라도 입력 유효성 검사는 적용됩니다. 어떤 검사를 생략할지 여부는 함수의 계약이며, 이 예제는 모든 성공 입력이 유효한 UTF-8이도록 정의했습니다.


### 5.2 이 함수가 약속하지 않는 것

이 함수는 코드 포인트 개수를 제한합니다. 사용자가 화면에서 한 글자로 보는 결합 단위까지 보존하는 함수는 아닙니다. 예를 들어 e와 결합 악센트로 이루어진 두 룬을 1개로 자르면 e만 남습니다.

이름이나 제목을 자르는 실제 제품에서는 길이 제한의 단위를 먼저 정해야 합니다. 저장 공간 제한이면 바이트, 코드 포인트 규칙이면 룬, 화면 표시 단위라면 별도의 유니코드 분할 규칙이 필요할 수 있습니다.

UTF-8을 깨뜨리지 않는 것과 모든 사람이 보는 글자 묶음을 보존하는 것은 다른 요구사항입니다. 이번 연습에서는 룬 단위라고 명시하여 범위를 분명히 합니다.

## 6. 잘못된 UTF-8을 구분하기

string은 임의의 바이트열을 담을 수 있습니다. range는 잘못된 UTF-8을 만나면 그 위치에서 대체 룬 U+FFFD를 제공하고 한 바이트씩 진행합니다. 따라서 range를 했다는 사실만으로 원본이 유효했다고 판단할 수 없습니다.

### 6.1 RuneError와 실제 대체 문자

저장 경로: [`examples/07-strings-unicode/utf8-validation/main.go`](../../examples/07-strings-unicode/utf8-validation/main.go). 전체 코드입니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := string([]byte{'A', 0xff, 'B'})
	fmt.Println("유효한 UTF-8:", utf8.ValidString(text))
	for index, value := range text {
		fmt.Printf("%d: %U\n", index, value)
	}
	value, width := utf8.DecodeRuneInString(text[1:])
	fmt.Printf("잘못된 바이트: %U, 크기: %d\n", value, width)
	value, width = utf8.DecodeRuneInString("\uFFFD")
	fmt.Printf("실제 대체 문자: %U, 크기: %d\n", value, width)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/07-strings-unicode/utf8-validation
```

예상 출력:

```text
유효한 UTF-8: false
0: U+0041
1: U+FFFD
2: U+0042
잘못된 바이트: U+FFFD, 크기: 1
실제 대체 문자: U+FFFD, 크기: 3
```

0xff는 이 위치에서 올바른 UTF-8 인코딩이 아닙니다. range는 인덱스 1에서 U+FFFD를 주고 다음 바이트의 B로 진행합니다.

DecodeRuneInString은 첫 룬과 사용한 바이트 수를 돌려줍니다. 잘못된 바이트는 RuneError와 크기 1입니다. 반면 실제로 UTF-8로 저장된 U+FFFD는 같은 룬 값이지만 크기 3입니다.

| 입력 시작 | 반환 룬 | 바이트 수 |
| --- | --- | --- |
| 잘못된 0xff | U+FFFD | 1 |
| 유효한 U+FFFD | U+FFFD | 3 |
| 빈 문자열 | U+FFFD | 0 |

따라서 RuneError라는 값만으로 유효한 대체 문자와 잘못된 인코딩을 구분할 수 없습니다. 전체 유효성은 ValidString으로, 한 위치 디코딩은 반환 크기와 함께 판단합니다. [DecodeRuneInString](https://pkg.go.dev/unicode/utf8#DecodeRuneInString)을 참고합니다.


### 6.2 처리 정책을 먼저 정하기

| 정책 | 언제 선택할 수 있나? | 관찰할 결과 |
| --- | --- | --- |
| 거부 | 유효한 텍스트만 받는 계약 | false 또는 후속 장의 error |
| 대체 | 사람이 읽을 수 있게 표시하는 계약 | 원본과 다른 텍스트가 될 수 있음 |
| 바이트 보존 | 텍스트가 아닌 원시 자료 취급 | []byte 등으로 정확한 바이트 유지 |

[]rune으로 변환하면 잘못된 부분이 대체 룬으로 바뀝니다. 다시 string으로 바꿔도 원래의 잘못된 바이트가 복원되는 것은 아닙니다. 손실 없는 바이트 보존이 목적이라면 무작정 rune 변환을 사용하면 안 됩니다.

다음은 fmt와 unicode/utf8을 import한 main 안의 발췌입니다.

```go
original := "\xff"
converted := string([]rune(original))
fmt.Printf("원래 바이트: %x\n", original)
fmt.Printf("변환 바이트: %x\n", converted)
fmt.Println(utf8.ValidString(converted))
```

예상 출력:

```text
원래 바이트: ff
변환 바이트: efbfbd
true
```

결과가 유효해졌다는 것과 원본을 보존했다는 것은 다릅니다. RuneCountInString도 잘못된 바이트를 한 룬으로 세므로 유효성 검사 대신 사용하지 않습니다.

## 7. strings로 텍스트 처리하기

자주 하는 문자열 작업은 표준 라이브러리 strings에서 찾을 수 있습니다. 함수가 원본 string을 직접 수정하는 것이 아니라 결과를 반환한다는 점을 기억합니다.

### 7.1 공백, 분리, 연결, 검색

저장 경로: [`examples/07-strings-unicode/strings-ops/main.go`](../../examples/07-strings-unicode/strings-ops/main.go). 전체 코드입니다.

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	text := " \tGo  한글\nGo "
	fmt.Printf("양끝 정리: %q\n", strings.TrimSpace(text))
	fields := strings.Fields(text)
	fmt.Printf("단어: %q\n", fields)
	fmt.Println("연결:", strings.Join(fields, " | "))
	fmt.Printf("쉼표 분리: %q\n", strings.Split("go,,map", ","))
	fmt.Println("글의 바이트 위치:", strings.Index("Go한글", "글"))
	fmt.Println("접두사:", strings.HasPrefix("Go한글", "Go"))
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/07-strings-unicode/strings-ops
```

예상 출력:

```text
양끝 정리: "Go  한글\nGo"
단어: ["Go" "한글" "Go"]
연결: Go | 한글 | Go
쉼표 분리: ["go" "" "map"]
글의 바이트 위치: 5
접두사: true
```

TrimSpace는 양끝의 유니코드 공백을 제거합니다. 문자열 내부의 두 칸과 줄바꿈은 그대로 남습니다. Fields는 연속 공백·탭·줄바꿈 등을 경계로 나누고 빈 항목을 만들지 않습니다.

Join은 슬라이스의 원소 사이에 구분자를 넣어 연결합니다. Split은 지정한 구분자로 나누므로 연속 쉼표 사이의 빈 항목이 유지됩니다.

Index가 반환한 글의 위치 5는 바이트 위치입니다. 앞의 Go가 2바이트, 한이 3바이트이므로 글은 위치 5에서 시작합니다.

| 작업 | 함수 | 반환값 해석 |
| --- | --- | --- |
| 양끝 공백 제거 | TrimSpace | 정리한 새 문자열 값 |
| 공백으로 분리 | Fields | 빈 원소 없는 []string |
| 정확한 구분자로 분리 | Split | 빈 원소가 생길 수 있음 |
| 연결 | Join | 구분자를 사이에 넣은 문자열 |
| 첫 위치 검색 | Index | 바이트 위치, 없으면 -1 |
| 시작 부분 검사 | HasPrefix | bool |

근거: [Fields](https://pkg.go.dev/strings#Fields), [Split](https://pkg.go.dev/strings#Split), [Index](https://pkg.go.dev/strings#Index).


### 7.2 Fields와 Split의 빈 값 차이

다음 전체 프로그램을 개인 파일에서 실행합니다.

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Printf("Fields: %q\n", strings.Fields(" a  b "))
	fmt.Printf("Split: %q\n", strings.Split(" a  b ", " "))
	fmt.Println("빈 Fields 길이:", len(strings.Fields("")))
	fmt.Println("빈 Split 길이:", len(strings.Split("", ",")))
}
```

```sh
go run ./_practice/07-strings-unicode/lab
```

예상 출력:

```text
Fields: ["a" "b"]
Split: ["" "a" "" "b" ""]
빈 Fields 길이: 0
빈 Split 길이: 1
```

입력의 빈 항목 자체가 의미가 있으면 Split 결과를 함부로 지우지 않습니다. 반대로 단어 모음을 얻고 싶다면 Fields가 더 직접적일 수 있습니다. 정확한 CSV 형식은 따옴표 등 별도 규칙이 있으므로 단순 Split으로 일반 CSV 처리를 대신하지 않습니다.

### 7.3 찾지 못한 위치를 먼저 확인하기

Index가 -1이면 인덱스로 쓰거나 슬라이스 경계로 넘기기 전에 처리해야 합니다.

```go
text := "name=Go"
index := strings.Index(text, "=")
if index < 0 {
	fmt.Println("구분자 없음")
} else {
	fmt.Println(text[:index], text[index+1:])
}
```

위는 fmt·strings를 import한 main 안의 발췌이며 `name Go`를 출력합니다. 찾은 구분자가 이 예제처럼 한 바이트인 =이므로 +1을 했습니다. 여러 바이트인 구분자라면 그 바이트 길이를 고려해야 합니다.

두 부분으로 나누는 목적이면 `strings.Cut`이 before, after, found를 함께 반환합니다. [Cut](https://pkg.go.dev/strings#Cut)을 사용하면 구분자 길이 계산을 API에 맡길 수 있습니다.

```go
before, after, found := strings.Cut("name=Go", "=")
fmt.Println(before, after, found)
```

이 독립된 main 발췌는 `name Go true`를 출력합니다. 실패한 결과를 어떻게 사용할지 found로 구분합니다.

### 7.4 Trim은 접두사 제거와 다르다

`strings.Trim(s, cutset)`은 양끝에서 cutset에 들어 있는 **각 룬**을 제거합니다. 전체 문자열을 하나의 접두사로 해석하지 않습니다.

```go
fmt.Printf("Trim: %q\n", strings.Trim("go-go", "go-"))
fmt.Printf("Prefix: %q\n", strings.TrimPrefix("go-go", "go-"))
```

main 안의 위 발췌는 `Trim: ""`, `Prefix: "go"`를 각각 출력합니다. 접두사 문자열 하나를 제거하려면 TrimPrefix를 사용합니다. [Trim](https://pkg.go.dev/strings#Trim)과 [TrimPrefix](https://pkg.go.dev/strings#TrimPrefix)를 참고합니다.

## 8. 대소문자와 정규화의 범위

### 8.1 대소문자 처리

strings.ToLower는 유니코드 소문자 매핑을 적용한 문자열을 반환합니다. 원본을 바꾸려면 결과를 변수에 대입해야 합니다.

```go
text := "Go 한글"
lower := strings.ToLower(text)
fmt.Println(text)
fmt.Println(lower)
fmt.Println(strings.EqualFold("Go", "go"))
```

fmt·strings를 import한 main 안의 발췌입니다. 출력은 차례로 Go 한글, go 한글, true입니다.

EqualFold는 단순 유니코드 대소문자 접기 규칙으로 비교합니다. 모든 언어의 검색 규칙이나 다문자 치환을 구현하는 함수는 아닙니다. 예를 들어 ß와 ss의 비교는 false입니다. [ToLower](https://pkg.go.dev/strings#ToLower)와 [EqualFold](https://pkg.go.dev/strings#EqualFold)를 참고합니다.

### 8.2 비슷하게 보여도 다른 코드 포인트열일 수 있다

다음 전체 프로그램을 개인 파일에서 실행합니다.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	composed := "\u00e9"
	decomposed := "e\u0301"
	fmt.Println("바이트:", len(composed), len(decomposed))
	fmt.Println("룬:", utf8.RuneCountInString(composed), utf8.RuneCountInString(decomposed))
	fmt.Println("같은 값:", composed == decomposed)
}
```

```sh
go run ./_practice/07-strings-unicode/lab
```

예상 출력:

```text
바이트: 2 3
룬: 1 2
같은 값: false
```

첫 값은 é 코드 포인트 하나이고, 둘째는 e와 결합 악센트 두 개입니다. 화면에서 비슷하게 보여도 바이트열이 다릅니다. Go의 문자열 비교가 자동으로 유니코드 정규화를 수행하지는 않습니다.

여기서 **정규화(normalization)**는 서로 대응하는 표현을 정해진 형태로 맞추는 별도 처리입니다. 공백 정리나 소문자 변환을 했다고 이 차이까지 해결된 것은 아닙니다. [공식 문자열 설명](https://go.dev/blog/strings)의 코드 포인트 논의를 참고합니다.

이 교재의 기본 집계는 정규화를 추가하지 않습니다. 실제 사용자 이름이나 검색에서 어떤 표현을 같게 볼지는 요구사항으로 정하고 적절한 라이브러리·검증을 선택해야 합니다.

## 9. 공백으로 단어를 나누어 집계하기

06장의 빈도 집계에 문자열 처리를 연결합니다. 이번 예제는 유효한 UTF-8 입력을 받는다고 가정하고, **소문자 변환 후 공백으로 분리**합니다. 문장 부호를 지우는 단계는 없습니다.

### 9.1 텍스트 처리 규칙을 드러내기

저장 경로: [`examples/07-strings-unicode/word-count/main.go`](../../examples/07-strings-unicode/word-count/main.go). 전체 코드입니다.

```go
package main

import (
	"fmt"
	"sort"
	"strings"
)

func countWords(text string) map[string]int {
	counts := make(map[string]int)
	for _, word := range strings.Fields(strings.ToLower(text)) {
		counts[word]++
	}
	return counts
}
func main() {
	counts := countWords(" Go\t한글\nGO 한글 ")
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

저장소 루트에서 실행합니다.

```sh
go run ./examples/07-strings-unicode/word-count
```

예상 출력:

```text
"go": 2
"한글": 2
```

입력의 Go와 GO는 모두 go가 됩니다. Fields는 탭과 줄바꿈도 구분하므로 한글도 두 번 집계됩니다. 키를 정렬하여 출력 순서를 고정합니다.

| 단계 | 결과 |
| --- | --- |
| 소문자 변환 | Go·GO가 go로 바뀜 |
| 공백 분리 | go, 한글, go, 한글 |
| 집계 | go 2회, 한글 2회 |
| 정렬 출력 | go 다음 한글 |

go와 go,는 서로 다른 키입니다. Fields는 언어학적인 단어 분석기가 아니며 문장 부호를 자동 제거하지 않습니다. 빈 문자열이나 공백뿐인 입력은 빈 맵을 반환합니다.

문제에서 “단어 수”를 요구하면 어떤 경계를 단어로 정의하는지 적어야 합니다. 이 예제의 계약은 공백으로 나눈 항목입니다.


### 9.2 처리 순서와 원본 보존

원문을 그대로 표시하면서 정리된 키로 집계할 수 있습니다. 원문 변수에 처리 결과를 덮어쓸지, 별도 변수에 둘지는 용도에 맞게 결정합니다.

| 필요한 값 | 보관 방식 |
| --- | --- |
| 사용자가 입력한 원문 | 원래 string 유지 |
| 비교용 소문자 키 | ToLower 결과 |
| 공백을 정돈한 표시 | Fields 후 Join |
| 정확한 바이트 | 원래 string 또는 []byte |

문자열 자체가 불변이더라도 처리한 결과를 원문 변수에 대입하면 프로그램에서 원문을 다시 찾기 어려워질 수 있습니다. 데이터가 언제 어떤 규칙으로 바뀌었는지 이름으로 드러냅니다.

## 10. unicode로 룬 분류하기

ASCII 범위만 검사하면 한글이나 다른 문자 체계의 글자를 놓칠 수 있습니다. unicode 패키지는 유니코드 분류를 기준으로 룬을 검사하는 함수를 제공합니다.

### 10.1 글자, 십진 숫자, 공백, 기타

저장 경로: [`examples/07-strings-unicode/classification/main.go`](../../examples/07-strings-unicode/classification/main.go). 전체 코드입니다.

```go
package main

import (
	"fmt"
	"unicode"
)

func classify(text string) (int, int, int, int) {
	letters, digits, spaces, others := 0, 0, 0, 0
	for _, value := range text {
		switch {
		case unicode.IsLetter(value):
			letters++
		case unicode.IsDigit(value):
			digits++
		case unicode.IsSpace(value):
			spaces++
		default:
			others++
		}
	}
	return letters, digits, spaces, others
}
func main() {
	letters, digits, spaces, others := classify("Go 한글 42!")
	fmt.Printf("글자: %d, 숫자: %d, 공백: %d, 기타: %d\n", letters, digits, spaces, others)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./examples/07-strings-unicode/classification
```

예상 출력:

```text
글자: 4, 숫자: 2, 공백: 2, 기타: 1
```

IsLetter는 글자 분류, IsDigit은 십진 숫자, IsSpace는 유니코드 공백 여부를 검사합니다. 먼저 일치한 분류에서만 증가하므로 각 룬은 이 예제의 네 그룹 중 하나에 들어갑니다.

| 예제 원소 | 분류 |
| --- | --- |
| G, o, 한, 글 | 글자 |
| 4, 2 | 십진 숫자 |
| 두 공백 | 공백 |
| ! | 기타 |

IsDigit은 ASCII의 0~9만을 뜻하지 않습니다. 반면 모든 수 관련 기호를 다 포함하는 IsNumber와도 다릅니다. 필요한 계약에 맞는 함수를 선택합니다. [IsLetter](https://pkg.go.dev/unicode#IsLetter), [IsDigit](https://pkg.go.dev/unicode#IsDigit), [IsSpace](https://pkg.go.dev/unicode#IsSpace)를 참고합니다.


### 10.2 코드 포인트 규칙을 입력 검증으로 바꾸기

이름을 검증한다고 할 때 다음을 따로 정합니다.

1. 입력이 유효한 UTF-8이어야 하는가?
2. 길이는 바이트·룬·다른 표시 단위 중 무엇인가?
3. 첫 원소와 나머지 원소의 조건이 같은가?
4. 공백, 숫자, 밑줄, 결합 부호를 허용하는가?
5. 대소문자나 유니코드 표현을 통일하는가?

예를 들어 “2~8룬, 첫 룬은 글자, 나머지는 글자·십진 숫자·밑줄”은 실행 가능한 검사 계약입니다. 이것이 모든 서비스의 올바른 사용자 이름 정책이라는 뜻은 아닙니다. 문제 6에서 이 학습용 계약을 구현합니다.

유효하지 않은 UTF-8을 먼저 거부하면 분류 단계가 대체 룬을 원문으로 오해하는 일을 피할 수 있습니다. 길이를 확인한 뒤 첫 룬을 읽어 빈 입력의 인덱스 오류도 막습니다.

## 11. 직접 작성하기: 공백을 정돈한 문장

이번 실습의 입력은 텍스트 한 줄의 논리적 내용입니다. 탭·줄바꿈·연속 공백을 단일 공백으로 합치고 양끝 공백을 없앱니다. 대소문자와 문장 부호는 유지합니다.

잘못된 UTF-8은 false로 반환합니다. 빈 문자열이나 공백뿐인 문자열은 성공한 빈 결과입니다. 원래의 줄바꿈 배치를 보존하는 함수는 아닙니다.

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다. 자신이 쓰는 셸의 블록 하나만 선택합니다. 기존 `main.go`가 있으면 덮어쓰지 않습니다. 그 내용을 확인하고, 다른 실습은 별도 폴더에 보관하세요. 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("lab")) {
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
for practice_name in lab; do
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


### 11.1 전체 프로그램

개인 파일 `_practice/07-strings-unicode/lab/main.go`에 작성합니다.

```go
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func compactSpaces(text string) (string, bool) {
	if !utf8.ValidString(text) {
		return "", false
	}
	words := strings.Fields(text)
	return strings.Join(words, " "), true
}

func main() {
	text := " Go\t한글\nGO "
	cleaned, ok := compactSpaces(text)
	if !ok {
		fmt.Println("텍스트 확인")
		return
	}
	fmt.Printf("정리: %q\n", cleaned)
	fmt.Printf("바이트: %d, 룬: %d\n", len(cleaned), utf8.RuneCountInString(cleaned))
	fmt.Println("단어:", len(strings.Fields(cleaned)))
}
```

```sh
gofmt -w _practice/07-strings-unicode/lab/main.go
go run ./_practice/07-strings-unicode/lab
```

예상 출력:

```text
정리: "Go 한글 GO"
바이트: 12, 룬: 8
단어: 3
```

정리된 문자열에는 단어 사이의 공백 두 개도 포함되므로 룬 수는 8입니다. 단어 안의 룬만 합산한 수 6과 혼동하지 않습니다.

### 11.2 경계 입력

| 입력 표기 | 정리 결과 | 성공 |
| --- | --- | --- |
| `""` | 빈 문자열 | true |
| `" \t\n"` | 빈 문자열 | true |
| `"  한글  "` | 한글 | true |
| `"Go\u3000한글"` | Go 한글 | true |
| `"Go,  한글!"` | Go, 한글! | true |
| `"\xff"` | 빈 문자열 | false |

U+3000은 유니코드 공백이므로 Fields가 경계로 처리합니다. 쉼표와 느낌표는 공백이 아니므로 유지됩니다. 실험 뒤에는 기본 입력으로 복구합니다.

공백을 모두 지우는 `ReplaceAll(text, " ", "")`는 다른 요구사항입니다. 단어 사이 경계를 없애며 탭과 줄바꿈 처리도 다릅니다. 원하는 결과의 예를 먼저 적으면 API를 선택하기 쉬워집니다.

## 12. 선택 학습: 문자열 조립하기

몇 개의 값만 붙일 때는 +나 fmt.Sprintf, 이미 모아 둔 문자열 슬라이스에는 strings.Join을 사용할 수 있습니다. 많은 조각을 단계적으로 붙이는 경우에는 strings.Builder라는 표준 라이브러리 타입도 있습니다.

여기서 `builder.WriteString`은 특정 값에 연결된 동작을 호출하는 **메서드(method)**입니다. 구조체와 메서드의 선언 방법은 08·09장에서 배우며, 아래는 라이브러리 사용을 미리 보는 선택 예제입니다.

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	parts := []string{"Go", "한글", "학습"}
	var builder strings.Builder
	for i, part := range parts {
		if i > 0 {
			builder.WriteString(" | ")
		}
		builder.WriteString(part)
	}
	fmt.Println(builder.String())
}
```

```sh
go run ./_practice/07-strings-unicode/lab
```

예상 출력:

```text
Go | 한글 | 학습
```

Builder의 영값은 바로 사용할 수 있습니다. 내용을 쓴 뒤의 Builder 값을 대입이나 값 전달로 복사하지 않아야 합니다. 이 예제는 하나의 지역 Builder만 사용합니다. [Builder](https://pkg.go.dev/strings#Builder)의 계약입니다.

WriteString은 쓴 바이트 수와 항상 nil인 오류를 반환하는 API입니다. 이 예제에서는 반환값을 사용하지 않습니다. 실패할 수 있는 파일·네트워크 쓰기에도 오류를 무시해도 된다는 일반 규칙은 아닙니다. [Builder.WriteString](https://pkg.go.dev/strings#Builder.WriteString)을 참고합니다.

이 입력은 이미 []string이므로 `strings.Join(parts, " | ")`도 같은 결과를 만듭니다. 어느 쪽이 항상 빠르다고 가정하지 않고, 데이터가 모여 있는지 단계적으로 생성되는지와 측정 결과를 기준으로 선택합니다.

## 13. 자주 하는 실수와 오류 읽기

| 증상 | 원인 | 수정 방향 |
| --- | --- | --- |
| 한글 길이가 예상보다 큼 | len을 표시 글자 수로 해석 | 필요한 단위 확인 |
| 앞부분을 잘랐더니 깨짐 | UTF-8 중간 바이트에서 자름 | 룬 또는 올바른 바이트 경계 사용 |
| range 인덱스가 건너뜀 | 인덱스를 룬 번호로 해석 | 바이트 시작 위치로 읽기 |
| ToLower 뒤 원문이 그대로 | 반환값을 사용하지 않음 | 결과를 필요한 변수에 대입 |
| 쉼표가 단어에 남음 | Fields가 문장 부호도 지운다고 가정 | 분리 계약 명시 |
| 빈 Split 결과가 한 개 | 빈 항목을 포함하는 규칙 | Fields와 Split 차이 확인 |
| 검색 실패 뒤 범위 오류 | Index의 -1을 경계로 사용 | 실패 여부 먼저 검사 |
| 대체 문자만 보고 잘못된 입력 판정 | 실제 U+FFFD와 혼동 | ValidString 또는 디코딩 크기 확인 |
| 비슷한 글자가 다른 맵 키 | 서로 다른 코드 포인트열 | 정규화 요구사항 검토 |
| RuneCount로 유효성 검사 | 잘못된 바이트도 세는 API | ValidString 사용 |

### 13.1 문자열 원소 대입 오류

다음은 **의도적으로 컴파일에 실패하는 전체 코드**입니다.

```go
package main

func main() {
	text := "Go"
	text[0] = 'N'
}
```

```sh
go run ./_practice/07-strings-unicode/lab
```

오류에서 `cannot assign to text[0]`를 찾습니다. 바이트를 바꾸는 목적이라면 []byte로 변환하고 수정한 뒤 string으로 만듭니다. 코드 포인트를 바꾸는 목적이라면 []rune이 더 알맞을 수 있습니다.

### 13.2 검증 순서

1. 함수 계약이 바이트·룬 중 무엇을 요구하는지 적습니다.
2. 빈 입력과 잘못된 UTF-8의 처리 방식을 정합니다.
3. 길이와 검색 결과를 검사한 뒤 인덱스를 사용합니다.
4. 처리 후 어떤 정보가 사라지는지 확인합니다.
5. ASCII, 한글, 결합 문자, 빈 입력을 비교합니다.

단순히 ASCII 입력 하나가 성공했다고 한글과 모든 유니코드 조합까지 올바르게 처리했다고 결론 내리지 않습니다. 시험한 규칙과 아직 처리하지 않는 단위를 구분합니다.

## 14. 연습문제와 검증

[07장 연습문제](../../exercises/07-strings-unicode/README.md)는 기본 6문제와 선택 A를 제공합니다.

| 문제 | 확인할 것 |
| --- | --- |
| 1. 바이트·룬 출력 예측 | 길이와 range 위치 |
| 2. 룬 제한 요약 | 생략 기호·0개·잘못된 UTF-8 |
| 3. 바이트 자르기 수정 | 문자열 단위와 부분 범위 |
| 4. 태그 정리 | 양끝 공백·소문자·중복·순서 |
| 5. 단어 집계 | 유효성·분리 규칙·정렬 |
| 6. 이름 검증 | 길이·첫 룬·허용 분류 |
| 선택 A. 룬 역순 | 코드 포인트 순서와 표시 단위의 한계 |

먼저 스스로 구현하고 [해설](../../solutions/07-strings-unicode/README.md)과 비교합니다. 참고 구현 검사는 저장소 루트에서 실행합니다.

```sh
go test -count=1 ./examples/07-strings-unicode/... ./solutions/07-strings-unicode/...
```

검사는 한글, 빈 입력, 음수 제한, 잘못된 바이트, 대소문자, 유니코드 공백과 결합 부호를 포함합니다. 테스트가 확인하는 것은 명시한 계약이며 모든 언어의 표시·검색 요구를 대신하지 않습니다.

## 15. 요약과 자기 점검

- string의 len과 인덱스는 바이트 기준입니다.
- 문자열 range는 UTF-8을 디코딩하며 바이트 위치와 rune을 줍니다.
- []byte는 바이트 원소, []rune은 코드 포인트 원소를 다룹니다.
- string 값은 내부 원소를 직접 바꿀 수 없습니다.
- RuneCount와 UTF-8 유효성 검사는 다른 작업입니다.
- 공백 정리·대소문자 처리·유니코드 정규화는 별도 규칙입니다.
- Fields, Split, Trim, TrimPrefix의 의미를 구분합니다.
- 룬 하나가 사용자가 보는 한 글자와 항상 같은 것은 아닙니다.

### 확인 질문

1. Go한글의 len이 8인데 range가 네 번 실행되는 이유는 무엇인가요?
2. 문자열 range의 인덱스 5는 다섯 번째 룬인가요?
3. `text[:3]`이 UTF-8 중간을 자를 수 있는 이유는 무엇인가요?
4. []rune을 바꾸면 원래 string도 바뀌나요?
5. 실제 U+FFFD와 잘못된 바이트를 디코딩한 결과를 어떻게 구분하나요?
6. RuneCountInString만으로 유효한 UTF-8인지 검사할 수 있나요?
7. Fields와 Split은 연속 공백과 빈 입력을 어떻게 다르게 처리하나요?
8. Index가 -1이면 무엇을 먼저 해야 하나요?
9. 소문자 변환이 유니코드 정규화까지 수행하나요?
10. 코드 포인트 하나만 남기는 자르기가 결합 문자 단위를 보존하나요?
11. `string(rune(65))`와 strconv.Itoa(65)의 결과는 어떻게 다른가요?
12. 원문이 필요한 프로그램에서 정리한 값을 어떻게 보관할까요?

<details>
<summary>답과 설명</summary>

1. 바이트 수와 코드 포인트 수가 다르기 때문입니다. G·o는 한 바이트씩, 예제의 한·글은 세 바이트씩입니다.
2. 아닙니다. 현재 룬이 시작하는 바이트 위치입니다.
3. 부분 범위가 바이트 기준이기 때문입니다. UTF-8 코드 포인트는 여러 바이트일 수 있습니다.
4. 바뀌지 않습니다. 수정한 슬라이스로 새 string을 만들어 사용합니다.
5. DecodeRuneInString은 잘못된 바이트에서 RuneError와 크기 1, 유효한 U+FFFD에서 크기 3을 줍니다. 전체 검사는 ValidString으로 할 수 있습니다.
6. 없습니다. 잘못된 바이트도 개수에 포함합니다.
7. Fields는 공백 묶음을 건너뛰고 빈 항목을 만들지 않습니다. 지정 구분자의 Split은 빈 항목을 유지할 수 있습니다.
8. 찾지 못한 경우를 처리한 뒤에만 인덱스나 부분 범위를 사용합니다.
9. 아닙니다. 서로 다른 코드 포인트 표현의 통일은 별도 처리입니다.
10. 보장하지 않습니다. 여러 룬으로 이루어진 표시 단위는 잘릴 수 있습니다.
11. 앞은 코드 포인트 A, 뒤는 숫자 표기 65입니다.
12. 원문과 비교·표시용 결과를 별도 변수로 두고 적용한 규칙을 드러낼 수 있습니다.

</details>

### 완료 체크리스트

- [ ] 길이와 위치의 단위를 명시했다.
- [ ] 한글을 바이트 중간에서 자르는 사례를 설명했다.
- [ ] 빈 결과와 잘못된 입력을 구분했다.
- [ ] 분리·공백·대소문자 처리의 범위를 기록했다.
- [ ] 결합 문자와 잘못된 바이트를 시험했다.
- [ ] 기본 문제 1~6과 지정된 경계 입력을 확인했다.

## 참고 자료

- [Go 명세: 문자열 타입](https://go.dev/ref/spec#String_types)
- [Go 명세: 문자열 변환](https://go.dev/ref/spec#Conversions_to_and_from_a_string_type)
- [Go 명세: range](https://go.dev/ref/spec#For_range)
- [Go 블로그: Strings, bytes, runes and characters](https://go.dev/blog/strings)
- [unicode/utf8](https://pkg.go.dev/unicode/utf8)
- [strings](https://pkg.go.dev/strings)
- [unicode](https://pkg.go.dev/unicode)
- [strconv.Itoa](https://pkg.go.dev/strconv#Itoa)

확인일: 2026-11-09. 예제와 경계 입력은 Go 1.27.3에서 확인했습니다. 기본 명령은 운영체제 공통이며 셸별 준비 명령을 구분했습니다.

## 다음 학습

다음 08장에서는 관련 값을 구조체로 묶고 포인터로 변경 대상을 표현합니다. 슬라이스·맵에서 배운 공유와 값 복사의 차이를 구조체 필드에 적용합니다.

[이전: 06. 맵과 집계](../06-maps/README.md) · [다음: 08. 포인터와 구조체](../08-pointers-structs/README.md) · [전체 목차](../README.md) · [문제](../../exercises/07-strings-unicode/README.md) · [해설](../../solutions/07-strings-unicode/README.md)
