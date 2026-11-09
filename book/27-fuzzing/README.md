# 27. 퍼징과 견고한 테스트

테이블 테스트는 사람이 고른 입력을 확인합니다.
그러나 문자열에는 빈 값, 매우 긴 값, 잘못된 UTF-8, 구분자가 여러 번 나오는 값이 들어올 수 있습니다.
각 경우를 미리 떠올리는 데에는 한계가 있습니다.
퍼징(fuzzing)은 테스트 도구가 입력을 바꾸어 실행하면서 실패하는 사례를 찾는 방법입니다.

도구는 무엇이 올바른 결과인지 스스로 결정하지 않습니다.
개발자가 **유효한 입력의 범위와 지켜야 할 성질**을 작성해야 합니다.
이 장에서는 함수의 결과를 검사하는 규칙을 만들고,
발견한 실패를 재현 가능한 회귀 테스트로 남기는 과정을 배웁니다.

## 학습 목표

- `FuzzXxx(*testing.F)`와 시드 입력으로 퍼즈 테스트를 작성한다.
- 일반 `go test`의 시드 실행과 새로운 입력을 생성하는 퍼징을 구분한다.
- 왕복·멱등성·범위·입력 보존 같은 불변 조건을 설계한다.
- UTF-8, 정수 범위, CSV 형식의 허용·거부 계약을 테스트한다.
- 실패 입력을 보관하고 수정 후 다시 실행해 회귀를 방지한다.

## 사전 지식과 준비

[07장 문자열과 유니코드](../07-strings-unicode/README.md),
[11장 오류 처리](../11-errors-defer/README.md),
[12장 자동 테스트](../12-testing/README.md),
[15장 데이터 형식과 시간](../15-data-time/README.md)을 먼저 확인합니다.

- 기준 환경: 2026-11-09, Go 1.27.3, Linux amd64.
- 예제 위치: [`examples/27-fuzzing`](../../examples/27-fuzzing/).
- 명령은 모두 저장소 루트에서 실행합니다.
- 외부 API나 계정, DB를 호출하지 않습니다.
- Go의 내장 퍼징은 Go 1.18부터 제공되며 현재 기준 버전에서 사용할 수 있습니다.

기본 형식과 지원 조건은 [Go Fuzzing](https://go.dev/doc/security/fuzz/)에서 확인합니다.
퍼징은 모든 입력에 대한 정당성 증명이 아닙니다.
시드와 탐색 시간, 검사한 성질의 범위 안에서 결함을 찾는 도구입니다.

## 1. 테스트의 답을 어떻게 알 것인가

정확한 정답이 알려진 입력은 일반 테스트로 확인합니다.
예를 들어 룬 순서를 뒤집는 함수에 `"Go"`를 넣으면 `"oG"`여야 합니다.
도구가 만든 임의의 문자열에는 예상 결과를 매번 사람이 적기 어렵습니다.
대신 여러 입력에서 항상 성립해야 하는 성질을 검사할 수 있습니다.

| 성질 | 예 | 그 성질만으로 부족한 점 |
| --- | --- | --- |
| 왕복 | 인코딩 뒤 디코딩하면 원래 바이트 | 두 함수가 같은 잘못을 공유할 수 있음 |
| 멱등성 | 공백 정리를 두 번 해도 결과 동일 | 아무 일도 안 하는 함수도 통과할 수 있음 |
| 범위 | 보정 결과는 하한과 상한 사이 | 항상 하한을 반환하는 결함을 놓칠 수 있음 |
| 보존 | 룬 뒤집기 뒤 룬 개수 유지 | 순서를 실제로 뒤집었는지는 따로 확인 필요 |
| 거부 계약 | 잘못된 UTF-8은 오류 | 유효한 모든 입력을 거부하는 결함은 별도 검사 |

**불변 조건(invariant)**은 허용되는 입력에서 계속 유지해야 하는 성질입니다.
조건 하나를 찾았다고 테스트가 끝난 것은 아닙니다.
구체적인 정상·경계 예와 여러 성질을 함께 사용합니다.

### 1.1 “패닉이 없으면 통과”의 범위

호출만 하고 반환값을 보지 않는 퍼즈 테스트는 패닉을 찾는 데 도움이 됩니다.
하지만 잘못된 값을 조용히 반환하는 결함은 놓칩니다.
문자열을 항상 빈 값으로 반환하는 함수도 패닉 없이 실행될 수 있습니다.
결과의 의미를 검증하는 조건을 추가해야 합니다.

### 1.2 오류가 정상 결과인 경우

파서가 잘못된 입력에 오류를 반환하는 것은 기대한 동작입니다.
퍼저가 이상한 문자열을 만들었다는 이유만으로 테스트를 실패시키면
허용하지 않는 입력을 정상적으로 거부한 구현을 결함으로 오해합니다.

입력 계약을 다음과 같이 나누어 적습니다.

1. 허용되는 입력에서 반드시 성공해야 하는 조건.
2. 허용되지 않는 입력에서 반드시 거부해야 하는 조건.
3. 성공했을 때 결과가 만족해야 하는 성질.

## 2. 첫 퍼즈 테스트: 룬 순서 뒤집기

소스: [`reverse/main.go`](../../examples/27-fuzzing/reverse/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

var errUTF8 = errors.New("invalid UTF-8")

func reverseRunes(text string) (string, error) {
	if !utf8.ValidString(text) {
		return "", errUTF8
	}
	values := []rune(text)
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
	return string(values), nil
}

func main() {
	result, err := reverseRunes("Go한글")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
}
```

```sh
go run ./examples/27-fuzzing/reverse
```

예상 출력:

```text
글한oG
```

### 2.1 입력 계약을 먼저 정하기

이 함수는 유효한 UTF-8 문자열을 받아 **룬 순서**를 뒤집습니다.
잘못된 UTF-8은 `errUTF8`로 거부합니다.
`[]rune`으로 바꾸기 전에 유효성을 검사하여 잘못된 바이트가 대체 문자로 바뀌는 것을 피합니다.
[utf8.ValidString](https://pkg.go.dev/unicode/utf8#ValidString)의 계약을 사용합니다.

룬은 화면에 보이는 글자 한 덩어리와 항상 같지 않습니다.
예를 들어 `a`와 결합 악센트로 이루어진 문자열은 두 룬입니다.
이 함수를 이모지나 결합 문자의 시각적 반전 기능이라고 소개하지 않습니다.
문자열의 바이트·룬 구분은 [07장](../07-strings-unicode/README.md)을 참고하세요.

### 2.2 전체 테스트 코드

소스: [`reverse/main_test.go`](../../examples/27-fuzzing/reverse/main_test.go)

```go
package main

import (
	"errors"
	"testing"
	"unicode/utf8"
)

func TestReverse(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"", ""}, {"Go", "oG"}, {"Go한글", "글한oG"}, {"a\u0301", "\u0301a"}} {
		got, err := reverseRunes(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("got=%q err=%v", got, err)
		}
	}
	if _, err := reverseRunes(string([]byte{0xff})); !errors.Is(err, errUTF8) {
		t.Fatal("invalid accepted")
	}
}

func FuzzReverse(f *testing.F) {
	for _, seed := range []string{"", "Go", "한글", "a\u0301", string([]byte{0xff})} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 65536 {
			t.Skip()
		}
		got, err := reverseRunes(input)
		if !utf8.ValidString(input) {
			if !errors.Is(err, errUTF8) {
				t.Fatal("invalid accepted")
			}
			return
		}
		if err != nil || !utf8.ValidString(got) {
			t.Fatalf("invalid result %q: %v", got, err)
		}
		if utf8.RuneCountInString(got) != utf8.RuneCountInString(input) {
			t.Fatal("lost rune")
		}
		again, err := reverseRunes(got)
		if err != nil || again != input {
			t.Fatalf("round trip: %q", again)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 글한oG
}
```

`TestReverse`는 정확히 알려진 결과를 확인합니다.
`FuzzReverse`는 다양한 입력에서 UTF-8 유효성, 룬 개수, 두 번 뒤집기 결과를 확인합니다.
둘은 서로를 보완합니다.

### 2.3 `f.Add`와 `f.Fuzz`

`f.Add`는 시작할 입력인 시드(seed)를 등록합니다.
`f.Fuzz`에 전달한 함수가 도구가 만든 각 입력을 검사하는 퍼즈 타깃입니다.
첫 인자는 `*testing.T`이고, 뒤의 인자 타입은 시드의 타입·순서와 일치해야 합니다.
지원되는 타입은 `string`, `[]byte`, 정수·실수·불리언 등으로 제한됩니다.
[`testing.F`](https://pkg.go.dev/testing#F)에서 정확한 목록을 확인합니다.

`f.Add`는 `f.Fuzz` 호출 전에 합니다.
검사 함수 안에서는 받은 `t`를 사용하여 오류를 보고합니다.
같은 `FuzzXxx`에서 여러 `f.Fuzz`를 호출하지 않습니다.
서로 다른 입력 계약을 검사하려면 별도 `FuzzXxx` 함수로 나눕니다.

## 3. 일반 실행과 퍼징 실행 구분하기

### 3.1 저장된 입력만 검사

```sh
go test -count=1 ./examples/27-fuzzing/...
```

이 명령은 일반 테스트와 예제 테스트뿐 아니라
등록된 시드·저장된 회귀 입력으로 퍼즈 타깃을 실행합니다.
새로운 입력을 계속 생성하는 퍼징 모드는 아닙니다.
기본 CI에서도 이 경로를 실행하므로 저장된 실패 사례가 다시 깨지는지 확인할 수 있습니다.
[testing의 퍼징 설명](https://pkg.go.dev/testing#hdr-Fuzzing)을 참고하세요.

### 3.2 새로운 입력을 생성

```sh
go test -run '^$' -fuzz '^FuzzReverse$' -fuzztime=5s -parallel=2 ./examples/27-fuzzing/reverse
```

PowerShell과 Bash에서 같은 명령을 사용합니다.
이 예제는 단일 패키지와 단일 퍼즈 타깃을 지정합니다.
정규식이 여러 타깃과 일치하면 실행할 대상을 명확하게 좁혀야 합니다.

| 옵션 | 목적 |
| --- | --- |
| `-run '^$'` | 일반 테스트 이름을 선택하지 않음 |
| `-fuzz '^FuzzReverse$'` | 퍼징할 함수 하나 선택 |
| `-fuzztime=5s` | 새로운 입력을 탐색할 시간 제한 |
| `-parallel=2` | 동시에 사용하는 퍼징 작업자 수 제한 |

퍼징은 CPU와 메모리를 사용합니다.
개발 중에는 짧게 실행하고 중요한 경계에서는 더 긴 예산을 정할 수 있습니다.
시간을 늘린다고 모든 결함을 찾았다는 보장은 생기지 않습니다.
[퍼징 플래그](https://pkg.go.dev/cmd/go#hdr-Testing_flags)의 의미를 확인하세요.

### 3.3 실행 결과 읽기

도구는 시드 기준선 수집, 경과 시간, 실행 횟수, 새로 유용해진 입력 수 등을 표시합니다.
정확한 횟수와 속도는 환경마다 달라지므로 고정 출력으로 비교하지 않습니다.
끝에 `PASS`가 나오면 이번 실행에서 검사한 조건을 위반한 입력을 발견하지 못한 것입니다.

새로 유용한 입력은 코드 경로 탐색에 도움이 된 입력이며
모두 결함이라는 뜻은 아닙니다.
실패가 발견되면 원인, 재실행 명령, 저장 위치를 읽습니다.
실패 입력을 지우기 전에 먼저 재현해야 합니다.

## 4. 입력 제한은 계약인가, 실험 예산인가

`FuzzReverse`는 65536바이트보다 큰 입력을 건너뜁니다.
이것은 함수 자체의 길이 제한이 아니라 짧은 실험의 자원 예산입니다.
함수에 길이 제한이 있는 것처럼 문서화하면 안 됩니다.

반면 뒤의 `parseRecord`는 함수 안에서 4096바이트 제한을 검사합니다.
이는 사용자가 의존할 수 있는 입력 계약입니다.
그 경계는 4096과 4097의 테스트로 확인해야 합니다.

| 제한 위치 | 의미 | 추가로 확인할 것 |
| --- | --- | --- |
| 함수 내부 | API의 허용 범위 | 경계의 성공·실패 |
| 퍼즈 타깃 내부 | 이번 탐색의 예산 | 생략한 범위의 별도 테스트 |
| 명령의 `-fuzztime` | 탐색 시간 | 실제 실행 범위와 한계 |

너무 많은 입력을 건너뛰면 퍼저가 중요한 경로에 도달하지 못합니다.
유효한 입력을 만드는 전용 타깃과 임의 입력을 거부하는 타깃을 분리하는 방법도 있습니다.
시드에는 빈 값, 최소·최대 경계, 대표적인 정상 값과 거부 값을 넣습니다.

## 5. 숫자 범위 파서의 성공 계약

소스: [`interval/main.go`](../../examples/27-fuzzing/interval/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errInterval = errors.New("invalid interval")

func parseInterval(text string) (int, int, error) {
	left, right, ok := strings.Cut(text, ":")
	if !ok || left == "" || right == "" {
		return 0, 0, errInterval
	}
	for _, part := range []string{left, right} {
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return 0, 0, errInterval
			}
		}
	}
	low, err := strconv.Atoi(left)
	if err != nil {
		return 0, 0, errInterval
	}
	high, err := strconv.Atoi(right)
	if err != nil || low > high || high > 10000 {
		return 0, 0, errInterval
	}
	return low, high, nil
}

func main() {
	low, high, err := parseInterval("2:10")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("범위: %d~%d\n", low, high)
}
```

```sh
go run ./examples/27-fuzzing/interval
```

```text
범위: 2~10
```

입력은 ASCII 숫자 두 묶음을 콜론 하나로 나눈 형태입니다.
성공 결과는 `0 <= low <= high <= 10000`을 만족해야 합니다.
공백, 부호, 추가 콜론은 허용하지 않고, 선행 0은 허용합니다.
`strconv.Atoi`가 처리할 수 없는 큰 수는 오류로 반환합니다.
[정수 파싱](https://pkg.go.dev/strconv#Atoi)의 범위 오류를 무시하지 않습니다.

### 5.1 전체 테스트

```go
package main

import (
	"fmt"
	"testing"
)

func TestInterval(t *testing.T) {
	for _, input := range []string{"", ":", "-1:2", "3:2", "0:10001", "1:2:3", " 1:2", "+1:2"} {
		if _, _, err := parseInterval(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	for _, tc := range []struct {
		input     string
		low, high int
	}{{"0:0", 0, 0}, {"0002:10", 2, 10}, {"0:10000", 0, 10000}} {
		low, high, err := parseInterval(tc.input)
		if err != nil || low != tc.low || high != tc.high {
			t.Fatal("bad result")
		}
	}
}

func FuzzInterval(f *testing.F) {
	for _, seed := range []string{"", "0:0", "2:10", "0:10000", "1:2:3"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 1024 {
			t.Skip()
		}
		low, high, err := parseInterval(input)
		if err != nil {
			return
		}
		if low < 0 || low > high || high > 10000 {
			t.Fatal("range contract")
		}
		again := fmt.Sprintf("%d:%d", low, high)
		l, h, e := parseInterval(again)
		if e != nil || l != low || h != high {
			t.Fatal("semantic round trip")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 범위: 2~10
}
```

```sh
go test -run '^$' -fuzz '^FuzzInterval$' -fuzztime=5s -parallel=2 ./examples/27-fuzzing/interval
```

성공한 결과를 `%d:%d`로 다시 만들고 파싱하면 같은 숫자를 얻어야 합니다.
문자열 자체는 같지 않을 수 있습니다.
예를 들어 `0002:10`은 정규 형태 `2:10`으로 바뀝니다.
이번 왕복 조건은 **문자열 일치가 아니라 의미 일치**입니다.

### 5.2 성공 결과만 검사하는 타깃의 한계

퍼즈 타깃이 오류이면 곧바로 반환하므로,
모든 입력을 오류로 거부하는 잘못된 구현도 이 타깃만 보면 통과할 수 있습니다.
그래서 `TestInterval`에서 반드시 성공해야 하는 정상·경계 입력을 따로 검사합니다.
좋은 불변 조건도 알려진 정답 테스트를 대체하지 않습니다.

## 6. 바이트 인코딩의 왕복 성질

소스: [`hex/main.go`](../../examples/27-fuzzing/hex/main.go)

```go
package main

import (
	"encoding/hex"
	"fmt"
)

func encode(data []byte) string { return hex.EncodeToString(data) }

func main() {
	data := []byte{0, 15, 255}
	text := encode(data)
	decoded, err := hex.DecodeString(text)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(text)
	fmt.Println(decoded)
}
```

```sh
go run ./examples/27-fuzzing/hex
```

```text
000fff
[0 15 255]
```

16진수 인코딩은 한 바이트를 두 개의 ASCII 문자로 표현합니다.
원본 바이트가 텍스트일 필요는 없습니다.
0이나 255처럼 UTF-8 문자가 아닌 값도 그대로 왕복할 수 있습니다.
[encoding/hex](https://pkg.go.dev/encoding/hex)의 계약을 확인하세요.

### 6.1 전체 테스트

```go
package main

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func FuzzHexRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 15, 255})
	f.Add([]byte("한글"))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 65536 {
			t.Skip()
		}
		encoded := encode(input)
		if len(encoded) != 2*len(input) {
			t.Fatal("wrong encoded length")
		}
		decoded, err := hex.DecodeString(encoded)
		if err != nil || !bytes.Equal(decoded, input) {
			t.Fatalf("round trip: %v", err)
		}
	})
}

func TestHexInvalid(t *testing.T) {
	for _, input := range []string{"0", "zz", "0x12"} {
		if _, err := hex.DecodeString(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 000fff
	// [0 15 255]
}
```

```sh
go test -run '^$' -fuzz '^FuzzHexRoundTrip$' -fuzztime=5s -parallel=2 ./examples/27-fuzzing/hex
```

`bytes.Equal`은 nil과 길이 0의 슬라이스를 같은 바이트열로 비교합니다.
이 문제의 계약은 같은 바이트 내용이며 nil 여부가 아닙니다.
인코딩 길이가 두 배라는 성질과 디코딩 결과가 원본이라는 성질을 함께 확인합니다.

### 6.2 잘못된 인코딩도 별도 검사

왕복 테스트는 올바르게 인코딩한 결과만 디코더에 넣습니다.
따라서 홀수 길이, 잘못된 문자, `0x` 접두어 같은 거부 경로는 별도 테스트가 필요합니다.
테스트가 어떤 입력을 만들 수 없는지 생각하면 빠진 범위를 찾을 수 있습니다.

## 7. 공백 정리와 멱등성

소스: [`normalize/main.go`](../../examples/27-fuzzing/normalize/main.go)

```go
package main

import (
	"fmt"
	"strings"
)

func normalizeSpace(text string) string { return strings.Join(strings.Fields(text), " ") }

func main() { fmt.Println(normalizeSpace("  Go\t 공부\n 시작  ")) }
```

```sh
go run ./examples/27-fuzzing/normalize
```

```text
Go 공부 시작
```

`strings.Fields`는 공백으로 나눈 원소를 반환하고,
`Join`은 그 사이에 ASCII 공백 하나를 넣습니다.
앞뒤 공백과 연속 공백을 정리한 결과를 다시 정리해도 바뀌지 않아야 합니다.
이 성질을 멱등성(idempotence)이라고 부릅니다.
[Fields](https://pkg.go.dev/strings#Fields)의 공백 정의를 사용합니다.

### 7.1 전체 테스트

```go
package main

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"", ""}, {" \t\n", ""}, {"가\u3000나", "가 나"}, {" a  b ", "a b"}} {
		if got := normalizeSpace(tc.in); got != tc.want {
			t.Fatalf("got=%q", got)
		}
	}
}

func FuzzNormalize(f *testing.F) {
	for _, seed := range []string{"", " ", "a\tb\n", "가\u3000나", string([]byte{0xff})} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 65536 {
			t.Skip()
		}
		got := normalizeSpace(input)
		if normalizeSpace(got) != got {
			t.Fatal("not idempotent")
		}
		if strings.HasPrefix(got, " ") || strings.HasSuffix(got, " ") || strings.Contains(got, "  ") {
			t.Fatalf("spacing: %q", got)
		}
		if len(strings.Fields(got)) != len(strings.Fields(input)) {
			t.Fatal("lost field")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// Go 공부 시작
}
```

```sh
go test -run '^$' -fuzz '^FuzzNormalize$' -fuzztime=5s -parallel=2 ./examples/27-fuzzing/normalize
```

멱등성 외에 앞뒤 ASCII 공백, 연속 ASCII 공백, 원소 개수를 검사합니다.
정확한 예제도 있어 항상 빈 문자열을 반환하는 결함을 잡을 수 있습니다.

이 함수는 Unicode 정규화(NFC/NFD)를 수행하지 않습니다.
대소문자나 시각적으로 같은 문자열을 통합하는 기능도 아닙니다.
잘못된 UTF-8을 자동으로 고친다는 계약도 두지 않았습니다.
공백 정리와 텍스트 유효성 검사는 서로 다른 책임입니다.

## 8. CSV 왕복에서 형식의 규칙 확인하기

소스: [`csv-pair/main.go`](../../examples/27-fuzzing/csv-pair/main.go)

```go
package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
)

var errCarriageReturn = errors.New("carriage return not supported")

func encodePair(left, right string) (string, error) {
	if strings.ContainsAny(left+right, "\r") {
		return "", errCarriageReturn
	}
	var result strings.Builder
	w := csv.NewWriter(&result)
	if err := w.Write([]string{left, right}); err != nil {
		return "", err
	}
	w.Flush()
	return result.String(), w.Error()
}

func main() {
	text, err := encodePair("Go,입문", "한글")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(text)
}
```

```sh
go run ./examples/27-fuzzing/csv-pair
```

```text
"Go,입문",한글
```

쉼표가 있는 필드는 따옴표로 감싸야 합니다.
직접 문자열을 쉼표로 이어 붙이면 필드 수가 바뀔 수 있으므로 `csv.Writer`를 사용합니다.
버퍼를 비운 뒤 `Error()`도 확인합니다.
[CSV Writer](https://pkg.go.dev/encoding/csv#Writer)의 오류 확인 방식입니다.

### 8.1 왜 CR을 거부하는가

이 예제는 필드 두 개를 원래 바이트 내용으로 복원하는 계약입니다.
Go CSV Reader는 CRLF를 LF로 정규화하는 동작이 있습니다.
따라서 CR을 포함한 입력까지 원본 바이트와 항상 같다고 주장하면 계약이 맞지 않습니다.
여기서는 CR을 명시적으로 거부하고 LF는 허용합니다.
[CSV Reader의 줄바꿈 규칙](https://pkg.go.dev/encoding/csv#Reader)을 확인합니다.

실제 서비스에서 CR을 허용하려면
정규화된 값의 왕복을 계약으로 정하거나 별도의 표현 방식을 선택할 수 있습니다.
이 장의 거부 정책은 연습용 설계 선택이며 모든 CSV의 필수 규칙이 아닙니다.

### 8.2 전체 테스트

```go
package main

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"testing"
)

func FuzzCSVPair(f *testing.F) {
	f.Add("", "")
	f.Add("Go,입문", "한글")
	f.Add("a\nb", "quote \"here\"")
	f.Add("a\r\nb", "x")
	f.Fuzz(func(t *testing.T, left, right string) {
		if len(left)+len(right) > 65536 {
			t.Skip()
		}
		encoded, err := encodePair(left, right)
		if strings.ContainsAny(left+right, "\r") {
			if !errors.Is(err, errCarriageReturn) {
				t.Fatal("CR accepted")
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		r := csv.NewReader(strings.NewReader(encoded))
		row, err := r.Read()
		if err != nil || len(row) != 2 || row[0] != left || row[1] != right {
			t.Fatalf("round trip: %q %v", row, err)
		}
		if _, err = r.Read(); err != io.EOF {
			t.Fatalf("extra record: %v", err)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// "Go,입문",한글
}
```

```sh
go test -run '^$' -fuzz '^FuzzCSVPair$' -fuzztime=5s -parallel=2 ./examples/27-fuzzing/csv-pair
```

반환된 두 필드가 각각 원본과 같은지 확인한 뒤,
두 번째 `Read()`가 EOF인지 검사합니다.
한 레코드로 약속한 결과에 추가 레코드가 생기면 실패합니다.
필드 안의 LF와 레코드 구분 LF를 구분해서 처리하는지 검증할 수 있습니다.

빈 필드 두 개도 한 레코드로 테스트합니다.
한 개의 빈 필드만 있는 레코드와 빈 줄을 혼동하지 않도록
이 예제는 항상 정확히 두 필드라는 조건을 유지합니다.

## 9. 정수 전체 범위를 다루는 보정 함수

소스: [`clamp/main.go`](../../examples/27-fuzzing/clamp/main.go)

```go
package main

import (
	"errors"
	"fmt"
)

var errBounds = errors.New("lower bound exceeds upper bound")

func clamp(value, low, high int) (int, error) {
	if low > high {
		return 0, errBounds
	}
	if value < low {
		return low, nil
	}
	if value > high {
		return high, nil
	}
	return value, nil
}

func main() {
	value, err := clamp(120, 0, 100)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("보정:", value)
}
```

```sh
go run ./examples/27-fuzzing/clamp
```

```text
보정: 100
```

값이 하한보다 작으면 하한, 상한보다 크면 상한을 돌려줍니다.
이미 범위 안이면 원래 값을 유지합니다.
하한이 상한보다 크면 잘못된 설정이므로 오류입니다.
뺄셈으로 범위를 비교하지 않아 정수 극단값에서의 오버플로를 피합니다.

### 9.1 전체 테스트

```go
package main

import (
	"errors"
	"testing"
)

func FuzzClamp(f *testing.F) {
	f.Add(120, 0, 100)
	f.Add(-1, 0, 100)
	f.Add(5, 5, 5)
	f.Add(1, 10, 0)
	f.Fuzz(func(t *testing.T, value, low, high int) {
		got, err := clamp(value, low, high)
		if low > high {
			if !errors.Is(err, errBounds) {
				t.Fatal("bad bounds accepted")
			}
			return
		}
		if err != nil || got < low || got > high {
			t.Fatal("outside range")
		}
		if value >= low && value <= high && got != value {
			t.Fatal("changed in-range input")
		}
		again, err := clamp(got, low, high)
		if err != nil || again != got {
			t.Fatal("not idempotent")
		}
	})
}

func TestClampEdges(t *testing.T) {
	max := int(^uint(0) >> 1)
	min := -max - 1
	got, err := clamp(max, min, max)
	if err != nil || got != max {
		t.Fatal("integer edge")
	}
}

func Example_main() {
	main()
	// Output:
	// 보정: 100
}
```

```sh
go test -run '^$' -fuzz '^FuzzClamp$' -fuzztime=5s -parallel=2 ./examples/27-fuzzing/clamp
```

퍼저는 값·하한·상한 세 정수를 함께 바꿉니다.
성공한 결과는 범위 안이고, 범위 안 입력은 그대로이며, 다시 보정해도 같아야 합니다.
하한과 상한이 같은 경우에는 성공 결과가 그 한 값이어야 합니다.

정수 최솟값과 최댓값은 일반 테스트에서도 직접 확인합니다.
임의 입력 생성에 맡기기만 하면 특정 경계에 도달하지 않을 수 있기 때문입니다.
[정수 타입과 범위](https://go.dev/ref/spec#Numeric_types)를 기준으로 경계를 만듭니다.

## 10. 구분자는 첫 번째 것만 의미할 수 있다

소스: [`key-value/main.go`](../../examples/27-fuzzing/key-value/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var errRecord = errors.New("invalid record")

func parseRecord(text string) (string, string, error) {
	if len(text) > 4096 || !utf8.ValidString(text) || strings.ContainsAny(text, "\r\n") {
		return "", "", errRecord
	}
	key, value, ok := strings.Cut(text, "=")
	if !ok || key == "" {
		return "", "", errRecord
	}
	for _, ch := range key {
		if (ch < 'a' || ch > 'z') && ch != '_' {
			return "", "", errRecord
		}
	}
	return key, value, nil
}

func main() {
	key, value, err := parseRecord("title=Go=입문")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("키: %s\n값: %s\n", key, value)
}
```

```sh
go run ./examples/27-fuzzing/key-value
```

```text
키: title
값: Go=입문
```

이 형식에서는 첫 `=`만 키와 값을 나눕니다.
값 안의 추가 `=`는 내용으로 보존합니다.
`strings.Cut`은 이 계약을 직접 표현합니다.
[Cut](https://pkg.go.dev/strings#Cut)의 반환값으로 구분자 존재 여부도 확인합니다.

키는 비어 있지 않은 소문자 ASCII와 `_`의 조합입니다.
값은 빈 문자열도 허용하지만 전체 입력은 유효한 UTF-8이며 CR·LF가 없어야 합니다.
전체 바이트 길이는 4096 이하입니다.
앞뒤 공백을 자동으로 제거하지 않으므로 성공 시 원본을 그대로 재구성할 수 있습니다.

### 10.1 전체 테스트

```go
package main

import (
	"strings"
	"testing"
)

func TestRecord(t *testing.T) {
	for _, input := range []string{"", "=x", "A=x", "a-b=x", "a=x\ny"} {
		if _, _, err := parseRecord(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	key, value, err := parseRecord("a_b=x=y")
	if err != nil || key != "a_b" || value != "x=y" {
		t.Fatal("split contract")
	}
}

func FuzzRecord(f *testing.F) {
	for _, seed := range []string{"", "a=", "title=Go=입문", "=x", "a=x\ny"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 8192 {
			t.Skip()
		}
		key, value, err := parseRecord(input)
		if err != nil {
			return
		}
		if key+"="+value != input {
			t.Fatal("lost bytes")
		}
		k, v, e := parseRecord(key + "=" + value)
		if e != nil || k != key || v != value {
			t.Fatal("round trip")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 키: title
	// 값: Go=입문
}

func TestRecordLengthBoundary(t *testing.T) {
	for _, n := range []int{4094, 4095} {
		_, _, err := parseRecord("a=" + strings.Repeat("x", n))
		if (err != nil) != (n == 4095) {
			t.Fatalf("value bytes=%d err=%v", n, err)
		}
	}
}
```

```sh
go test -run '^$' -fuzz '^FuzzRecord$' -fuzztime=5s -parallel=2 ./examples/27-fuzzing/key-value
```

범위 안의 임의 문자열에 대해서 파싱 성공 시 `key + "=" + value`가 원본과 같아야 합니다.
구분자 처리나 잘못된 공백 제거로 바이트가 사라지는 결함을 찾는 성질입니다.
유효한 정상 입력이 항상 성공하는지는 별도의 알려진 정답 테스트가 담당합니다.

### 10.2 길이 경계 추가하기

독자는 값 부분을 `strings.Repeat("x", n)`으로 만들어
전체 길이 4096과 4097의 차이를 확인할 수 있습니다.
`a=` 두 바이트도 전체 길이에 포함됩니다.
글자 수 제한으로 바꾸려면 `utf8.RuneCountInString` 등 별도 기준과 계약이 필요합니다.

## 11. 직접 결함을 만들고 다시 고치기

이번에는 정상 예제를 개인 폴더로 복사한 뒤 **의도적으로 잘못된 구현**을 만듭니다.
실패 코드는 기본 예제 디렉터리에 넣지 않습니다.

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/27-fuzzing/reverse | Out-Null
foreach ($file in @("main.go", "main_test.go")) {
    if (-not (Test-Path "_practice/27-fuzzing/reverse/$file")) {
        Copy-Item "examples/27-fuzzing/reverse/$file" "_practice/27-fuzzing/reverse/$file"
    }
}
```

Bash:

```sh
mkdir -p _practice/27-fuzzing/reverse
for file in main.go main_test.go; do
    if [ ! -e "_practice/27-fuzzing/reverse/$file" ]; then
        cp "examples/27-fuzzing/reverse/$file" "_practice/27-fuzzing/reverse/$file"
    fi
done
```

### 11.1 바이트 반전 결함

개인 `main.go`의 `reverseRunes` 안에서
`values := []rune(text)`를 `values := []byte(text)`로 바꿉니다.
나머지 코드는 유지합니다.
이 변경은 **한글을 구성하는 바이트 순서까지 뒤집는 의도된 결함**입니다.

```sh
go test -run '^$' -fuzz '^FuzzReverse$' -fuzztime=5s -parallel=2 ./_practice/27-fuzzing/reverse
```

시드에 `한글`이 들어 있으므로 새로운 입력을 탐색하기 전 기준선 검사에서 실패할 수 있습니다.
확인한 실행에서는 `invalid result` 메시지와 함께 잘못된 UTF-8 결과가 보고되었습니다.
파일의 줄 번호와 처리한 시드 수는 실행 환경·코드 수정에 따라 달라질 수 있습니다.

### 11.2 두 번 반전만 검사했다면

바이트를 두 번 뒤집으면 원래 바이트 순서로 돌아옵니다.
따라서 두 번 실행한 값만 비교하는 테스트는 이 결함을 놓칠 수 있습니다.
중간 결과도 유효한 UTF-8이어야 한다는 조건이 중요한 이유입니다.

반대로 항상 입력을 그대로 돌려주는 함수는
두 번 실행하기, UTF-8 유효성, 룬 개수 보존을 모두 통과할 수 있습니다.
`Go → oG` 같은 알려진 정답 테스트가 그 결함을 잡습니다.
여러 종류의 검증을 함께 설계해야 합니다.

### 11.3 수정 후 재검증

개인 파일을 다시 `[]rune(text)`로 고치고 다음 두 명령을 실행합니다.

```sh
go test -count=1 ./_practice/27-fuzzing/reverse
go test -run '^$' -fuzz '^FuzzReverse$' -fuzztime=5s -parallel=2 ./_practice/27-fuzzing/reverse
```

첫 명령은 알려진 정답과 시드를 확인합니다.
두 번째는 새 입력을 더 탐색합니다.
실패 입력을 삭제해서 검사만 통과시키지 않고, 입력 계약을 지키도록 구현을 고칩니다.

## 12. 코퍼스와 회귀 테스트 보관

코퍼스(corpus)는 테스트에 사용하는 입력 모음입니다.
`f.Add` 외에도 `testdata/fuzz/FuzzXxx/` 아래의 입력 파일을 읽을 수 있습니다.
퍼징 도구가 실패를 찾으면 재현에 필요한 입력과 명령을 안내할 수 있습니다.
[코퍼스 관리](https://go.dev/doc/security/fuzz/)의 형식을 따릅니다.

### 12.1 저장소에 넣은 입력 파일

파일: [`reverse/testdata/fuzz/FuzzReverse/unicode`](../../examples/27-fuzzing/reverse/testdata/fuzz/FuzzReverse/unicode)

```text
go test fuzz v1
string("한글")
```

이 파일은 실행할 Go 프로그램이 아니라 퍼징 도구가 읽는 입력 형식입니다.
이름 `unicode`는 이 교재에서 의미를 알아보기 쉽게 붙인 시드 이름입니다.
도구가 자동으로 만든 파일은 다른 이름을 가질 수 있습니다.

다음 명령으로 이 저장된 입력만 선택할 수 있습니다.

```sh
go test -run '^FuzzReverse/unicode$' ./examples/27-fuzzing/reverse
```

본문 예제는 이미 수정된 구현이므로 통과합니다.
일반 `go test`도 이 파일을 읽으므로 팀원이 새 입력 탐색을 하지 않아도 같은 사례를 확인합니다.

### 12.2 실패를 찾았을 때의 순서

1. 테스트가 출력한 입력과 재실행 명령을 보관합니다.
2. 같은 코드 버전에서 해당 입력만 실행해 재현합니다.
3. 함수 계약과 테스트 조건 중 어느 쪽이 잘못되었는지 판단합니다.
4. 구현 또는 잘못된 조건을 최소 범위로 수정합니다.
5. 발견한 입력을 회귀 검사에 포함하고 전체 테스트를 실행합니다.
6. 제한된 시간의 퍼징을 다시 실행해 주변 입력도 탐색합니다.

실패 입력은 작게 줄어들 수 있습니다.
작은 입력은 원인을 이해하기 좋지만, 원래 서비스의 모든 조건을 그대로 보존한다는 뜻은 아닙니다.
타깃이 네트워크·전역 상태에 의존하지 않으면 재현이 쉬워집니다.

### 12.3 캐시와 저장된 회귀 입력

탐색에 유용한 입력을 보관하는 도구 캐시와,
저장소의 `testdata/fuzz` 파일은 역할이 다릅니다.
팀에 전달할 회귀 사례를 개인 캐시에만 두면 다른 환경에서 실행되지 않을 수 있습니다.
실패가 된 입력과 수정 이유를 검토 가능한 파일·테스트로 남깁니다.

## 13. 재현 가능한 타깃 만들기

퍼즈 타깃은 같은 입력에 같은 판단을 해야 합니다.
현재 시각, 외부 API 결과, 파일 이름 충돌, 이전 호출에서 남은 전역 상태가 결과를 바꾸면
실패 원인을 입력으로 좁히기 어려워집니다.
[타깃 작성 권고](https://go.dev/doc/security/fuzz/#suggestions)에서 결정성과 상태 격리를 강조합니다.

### 13.1 입력마다 상태를 새로 만들기

맵·버퍼·파서 상태가 필요하면 타깃 호출 안에서 만들거나 명시적으로 초기화합니다.
두 호출이 같은 전역 맵을 수정하면 입력 순서에 따라 결과가 달라질 수 있습니다.
작업자가 병렬로 실행될 수 있다는 점도 고려합니다.

파일이 꼭 필요하면 임시 디렉터리를 사용하고 정리합니다.
하지만 수백만 번 생성할수록 비용이 커질 수 있으므로
가능하면 `io.Reader`·`io.Writer` 또는 순수 계산 부분을 먼저 분리합니다.
이는 파일 시스템의 의미를 검사하지 않아도 되는 경우의 설계 선택입니다.

### 13.2 `recover`로 결함을 숨기지 않기

검사 대상에서 난 모든 패닉을 무조건 복구하고 성공 처리하면
퍼저가 발견한 결함을 정상 결과로 바꾸게 됩니다.
패닉이 입력 계약의 일부인지, 구현 결함인지 먼저 판단합니다.
이 장의 파서들은 잘못된 입력을 오류로 반환하므로 패닉이 나면 테스트 실패로 보아야 합니다.

### 13.3 관계가 있는 여러 입력

하한과 상한처럼 서로 관계가 있는 값은 생성된 조합 대부분이 거부될 수 있습니다.
그렇다고 `low > high`인 값을 모두 교환하여 넣으면
거부 계약을 검사하는 경로가 사라집니다.
거부 조합과 유효 조합을 각각 어떻게 검사할지 정하세요.

구조체 자체를 퍼징 인자로 직접 받을 수 없다면
지원되는 바이트·문자열·정수 인자를 받아 구조체를 구성할 수 있습니다.
구성 단계에서 지나치게 많은 입력을 버리지 않도록 시드와 생성 규칙을 검토합니다.

## 14. 자주 하는 실수

| 증상 | 원인 | 수정 방향 |
| --- | --- | --- |
| 일반 테스트만 끝남 | `-fuzz`를 지정하지 않음 | 타깃 하나와 시간 제한을 지정 |
| 시드 형식 오류 | `f.Add`와 타깃 인자 타입·순서 불일치 | 개수·타입을 맞춤 |
| 타깃을 고를 수 없음 | 정규식이 여러 함수와 일치 | `^이름$`으로 하나 선택 |
| 오류 입력에서 계속 실패 | 정상 거부를 결함으로 취급 | 허용·거부 계약을 분리 |
| 성공 입력을 모두 거부해도 통과 | 오류이면 항상 반환하는 조건만 존재 | 알려진 정상 예 또는 유효 입력 전용 타깃 추가 |
| 같은 입력이 가끔만 실패 | 시각·전역 상태·외부 I/O 의존 | 상태 격리와 제어 가능한 의존성 사용 |
| UTF-8 문자열이 깨짐 | 바이트 순서로 텍스트를 변형 | 문자 단위와 유효성 계약 확인 |
| CSV 왕복이 CRLF에서 다름 | 형식의 줄바꿈 정규화 무시 | 거부 또는 정규화 계약을 명시 |
| 탐색 속도가 너무 느림 | 무제한 크기·파일·로그 비용 | 목표에 맞는 입력·시간·작업자 예산 설정 |
| 패닉이 발견되지 않음 | 모든 패닉을 복구해 성공 처리 | 의도한 복구 경계 외에는 실패가 보이게 함 |

### 14.1 테스트 자체의 오류도 검토하기

“뒤집은 문자열의 바이트 수는 항상 룬 수와 같다”는 조건은 한글에서 잘못되었습니다.
퍼저가 이런 테스트를 실패시켰다면 먼저 계약과 조건을 검토해야 합니다.
테스트가 실패했다고 구현을 무조건 바꾸면 올바른 코드를 잘못 고칠 수 있습니다.

### 14.2 실행 시간과 보장 범위

짧은 퍼징은 빠른 피드백에 좋고 긴 퍼징은 더 넓은 탐색 기회를 줍니다.
어느 쪽도 전체 입력 공간을 다 확인했다는 뜻은 아닙니다.
실제로 실행한 시간과 타깃, 입력 제한을 검증 기록에 남깁니다.

## 15. 연습문제

[연습문제](../../exercises/27-fuzzing/README.md)에는
불변 조건의 빈틈 찾기, 키·값 파서, 정수 보정, 공백 정리, 구간 파서와 회귀 입력 관리가 있습니다.
[해설](../../solutions/27-fuzzing/README.md)은 테스트 조건을 고른 이유와
일반 테스트·퍼즈 테스트를 함께 제공합니다.

```sh
go test -count=1 ./examples/27-fuzzing/... ./solutions/27-fuzzing/...
```

시간 제한을 둔 퍼징은 타깃별로 실행합니다.
여러 패키지 전체에 `-fuzz .`를 붙여 어떤 함수를 실행하는지 불분명하게 만들지 않습니다.

## 16. 요약과 자기 점검

퍼징은 도구가 입력을 바꾸고 개발자가 결과의 계약을 검사하는 협업입니다.
정확한 예, 불변 조건, 거부 입력과 회귀 코퍼스를 함께 사용합니다.
실패를 찾으면 입력을 보존하고 재현한 뒤 고칩니다.

### 확인 질문

1. 일반 `go test`와 `-fuzz` 실행은 무엇이 다른가요?
2. `f.Add`의 타입은 어디와 일치해야 하나요?
3. 두 번 뒤집으면 원본이라는 조건이 바이트 반전 결함을 놓치는 이유는 무엇인가요?
4. 모든 입력을 거부하는 파서가 통과할 수 있는 테스트는 어떤 모습인가요?
5. `0002:10` 파서의 왕복에서 문자열이 달라도 되는 이유는 무엇인가요?
6. 공백 정리의 멱등성만 검사하면 어떤 결함을 놓칠 수 있나요?
7. CSV의 CRLF 규칙은 바이트 왕복과 어떤 관계가 있나요?
8. 타깃의 크기 제한과 함수의 입력 제한을 어떻게 구분하나요?
9. 발견한 입력을 개인 캐시에만 두면 어떤 문제가 있나요?
10. 짧은 퍼징이 통과했다는 결과의 정확한 의미는 무엇인가요?

<details>
<summary>답 확인</summary>

1. 일반 실행은 시드와 저장된 입력을 검사하고, 퍼징 모드는 새 입력을 생성해 탐색합니다.
2. `f.Fuzz` 검사 함수의 `*testing.T` 뒤 인자 타입·순서·개수와 일치해야 합니다.
3. 잘못된 바이트 반전도 두 번 하면 원래 바이트 순서로 돌아올 수 있기 때문입니다.
4. 오류가 나면 무조건 성공 처리하고 반드시 성공해야 하는 입력을 따로 검사하지 않는 테스트입니다.
5. 선행 0을 허용하며 파싱한 두 정수의 의미를 보존하는 계약이기 때문입니다.
6. 항상 빈 값 또는 입력 그대로를 돌려주는 등 멱등성은 지키지만 정리는 하지 않는 결함입니다.
7. Reader가 CRLF를 LF로 바꾸면 원본 바이트와 동일하지 않을 수 있습니다.
8. 함수 내부 제한은 API 계약이고, 타깃 제한은 이번 검사의 자원 예산입니다.
9. 다른 개발자나 CI가 같은 회귀 사례를 실행하지 못할 수 있습니다.
10. 주어진 예산과 조건에서 실행한 입력들에서는 위반을 발견하지 못했다는 뜻입니다.

</details>

### 완료 체크리스트

- [ ] 시드와 타깃의 타입을 맞추었습니다.
- [ ] 정상·거부·경계 입력을 나누었습니다.
- [ ] 알려진 정답과 여러 불변 조건을 함께 검사했습니다.
- [ ] 시간·입력 크기·병렬 작업자 수를 명시했습니다.
- [ ] 의도된 결함이 실제로 실패하는 것을 확인했습니다.
- [ ] 저장된 코퍼스 하나를 이름으로 다시 실행했습니다.
- [ ] 수정 후 일반 테스트와 짧은 퍼징을 다시 실행했습니다.

## 참고 자료

- [Go Fuzzing](https://go.dev/doc/security/fuzz/)
- [공식 퍼징 튜토리얼](https://go.dev/doc/tutorial/fuzz)
- [testing.F](https://pkg.go.dev/testing#F)
- [testing — Fuzzing](https://pkg.go.dev/testing#hdr-Fuzzing)
- [go test 플래그](https://pkg.go.dev/cmd/go#hdr-Testing_flags)
- [unicode/utf8](https://pkg.go.dev/unicode/utf8)
- [encoding/hex](https://pkg.go.dev/encoding/hex)
- [encoding/csv](https://pkg.go.dev/encoding/csv)

공식 문서와 실행 확인: 2026-11-09, Go 1.27.3.
기본 시드 검사와 타깃별 제한 시간 퍼징을 구분하여 검증했습니다.

## 다음 학습

[28장 빌드와 배포 자동화](../28-build-deploy/README.md)에서
테스트·빌드·산출물 확인 과정을 반복 가능한 명령과 CI로 연결합니다.

[이전: 26. 성능 측정과 진단](../26-performance/README.md) · [전체 목차](../README.md) · [다음: 28. 빌드와 배포 자동화](../28-build-deploy/README.md)
