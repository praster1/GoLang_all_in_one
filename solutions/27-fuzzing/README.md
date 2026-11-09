# 27장 해설 — 퍼징과 견고한 테스트

[연습문제](../../exercises/27-fuzzing/README.md)를 먼저 시도한 뒤 확인합니다.
테스트가 무엇을 검사하고 무엇을 검사하지 않는지 함께 설명합니다.
정답 코드와 같은 구현을 작성했는지보다 입력·출력·오류 계약을 지켰는지 비교하세요.

## 실행 안내

모든 명령은 저장소 루트에서 실행합니다.

```sh
go test -count=1 ./solutions/27-fuzzing/...
```

| 문제 | 실행 구현 | 핵심 계약 |
| --- | --- | --- |
| 3 | [rune-prefix](rune-prefix/main.go) | UTF-8 유효성과 룬 접두어 |
| 4 | [port](port/main.go) | ASCII 숫자와 1~65535 |
| 5 | [normalize-tags](normalize-tags/main.go) | ASCII 태그, 소문자·정렬·중복 제거 |
| 6 | [checked-add](checked-add/main.go) | int 범위의 정확한 합 또는 오류 |

각 `main_test.go`에는 알려진 정답 검사와 퍼즈 타깃이 함께 있습니다.
기본 `go test`는 시드를 실행하고, 아래 개별 명령은 새 입력을 탐색합니다.
실행 횟수나 발견한 유용한 입력 수는 고정하지 않습니다.

## 문제 1. 조건의 빈틈 찾기

### 두 번 뒤집기

입력을 그대로 반환하는 함수는 두 번 실행해도 원본입니다.
하지만 `Go → oG`라는 구체적인 정답 테스트를 통과하지 못합니다.
따라서 왕복 성질과 알려진 출력 검사가 함께 필요합니다.

바이트 반전도 두 번 수행하면 원래 바이트열로 돌아올 수 있습니다.
이 경우 중간 결과가 유효한 UTF-8인지 검사해야 한글 손상을 발견할 수 있습니다.
룬 반전의 계약을 시각적 글자 덩어리 반전과 구분하는 것도 필요합니다.

### 정규화의 멱등성

항상 빈 문자열을 반환하는 구현은 두 번 실행해도 같은 빈 문자열입니다.
하지만 `" a  b " → "a b"`라는 정상 예를 통과하지 못합니다.
원소 수 보존 같은 추가 조건도 도움이 됩니다.

### 범위 보정

항상 하한을 반환하는 함수는 결과 범위 조건을 만족할 수 있습니다.
하지만 `clamp(5, 0, 10)`은 5를 유지해야 합니다.
“입력이 이미 범위 안이면 바꾸지 않는다”는 조건을 추가합니다.

| 결함 | 기존 조건을 통과하는 이유 | 추가 검사 |
| --- | --- | --- |
| 입력 그대로 반환 | 두 번 실행해도 원본 | 알려진 반전 결과 |
| 항상 빈 정규화 결과 | 두 번 실행해도 빈 값 | 정상 예와 원소 수 |
| 항상 하한 반환 | 결과는 범위 안 | 범위 안 입력 보존 |
| 바이트 순서 반전 | 두 번이면 바이트 복구 | 중간 UTF-8 유효성 |

## 문제 2. 시드 실행과 탐색

일반 실행은 `f.Add`와 `testdata/fuzz`의 입력을 검사합니다.
이름에 `Fuzz`가 있다는 사실만으로 계속 새로운 입력을 생성하는 것은 아닙니다.
`-fuzz`를 지정해야 탐색 모드로 진입합니다.

퍼징 모드도 먼저 등록된 입력을 검사합니다.
시드부터 실패하면 새 입력 생성 전에 멈출 수 있습니다.
입력의 개수·처리 속도는 CPU, 작업자 수, 실행 환경과 탐색 경로에 따라 달라집니다.

저장된 입력 하나를 실행하는 명령은 다음과 같습니다.

```sh
go test -run '^FuzzReverse/unicode$' ./examples/27-fuzzing/reverse
```

이 명령이 통과하면 해당 입력이 검사한 조건을 만족했다는 뜻입니다.
모든 가능한 입력의 동작을 증명한 결과는 아닙니다.
[기본 시드 실행과 퍼징](https://pkg.go.dev/testing#hdr-Fuzzing)의 차이를 참고하세요.

## 문제 3. 룬 접두어

### 전체 구현

```go
package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

var errInput = errors.New("invalid text or limit")

func prefixRunes(text string, limit int) (string, error) {
	if limit < 0 || !utf8.ValidString(text) {
		return "", errInput
	}
	runes := []rune(text)
	if limit > len(runes) {
		limit = len(runes)
	}
	return string(runes[:limit]), nil
}

func main() {
	result, err := prefixRunes("Go한글", 3)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
}
```

```sh
go run ./solutions/27-fuzzing/rune-prefix
```

```text
Go한
```

검사 순서가 중요합니다.
제한이 0이어도 잘못된 UTF-8은 거부한다는 계약이므로,
빈 결과를 먼저 반환하지 않고 유효성을 먼저 확인합니다.
그 뒤 제한을 실제 룬 개수 이하로 조정하여 범위 밖 슬라이싱을 피합니다.

| 단계 | 입력 `Go한글`, 제한 3 |
| --- | --- |
| UTF-8 검사 | 유효 |
| 룬으로 변환 | G, o, 한, 글 |
| 제한 조정 | min(3, 4) = 3 |
| 접두어 선택 | G, o, 한 |
| 문자열 반환 | Go한 |

### 테스트 전체

```go
package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPrefixRunes(t *testing.T) {
	for _, tc := range []struct {
		text string
		n    int
		want string
		bad  bool
	}{
		{"", 0, "", false}, {"Go한글", 3, "Go한", false}, {"한글", 10, "한글", false},
		{"Go", -1, "", true}, {string([]byte{0xff}), 0, "", true},
	} {
		got, err := prefixRunes(tc.text, tc.n)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("got=%q err=%v", got, err)
		}
	}
}

func FuzzPrefixRunes(f *testing.F) {
	f.Add("Go한글", 3)
	f.Add("", 0)
	f.Add("한", -1)
	f.Add(string([]byte{0xff}), 1)
	f.Fuzz(func(t *testing.T, text string, n int) {
		if len(text) > 65536 {
			t.Skip()
		}
		got, err := prefixRunes(text, n)
		if n < 0 || !utf8.ValidString(text) {
			if err == nil {
				t.Fatal("invalid accepted")
			}
			return
		}
		if err != nil || !utf8.ValidString(got) || !strings.HasPrefix(text, got) {
			t.Fatal("prefix contract")
		}
		wantCount := min(n, utf8.RuneCountInString(text))
		if utf8.RuneCountInString(got) != wantCount {
			t.Fatal("wrong length")
		}
		again, err := prefixRunes(got, n)
		if err != nil || again != got {
			t.Fatal("not idempotent")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// Go한
}
```

```sh
go test -run '^$' -fuzz '^FuzzPrefixRunes$' -fuzztime=3s -parallel=2 ./solutions/27-fuzzing/rune-prefix
```

`strings.HasPrefix`는 결과가 원본의 앞부분인지 확인합니다.
룬 개수 검사는 너무 짧게 자르는 결함을 찾습니다.
멱등성은 같은 제한으로 다시 잘랐을 때 안정적인지를 확인합니다.

이 구현은 전체 문자열을 룬 슬라이스로 만듭니다.
작은 접두어만 필요한 큰 입력에서는 순회하며 바이트 경계를 찾는 대안도 있습니다.
그 대안을 구현한다면 같은 정답·성질을 유지하는지 먼저 검사하고,
성능 차이는 26장의 방법으로 별도 측정합니다.

## 문제 4. 숫자 문자열과 포트 값

### 전체 구현

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

var errPort = errors.New("invalid port")

func parsePort(text string) (uint16, error) {
	if text == "" {
		return 0, errPort
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return 0, errPort
		}
	}
	value, err := strconv.ParseUint(text, 10, 16)
	if err != nil || value == 0 {
		return 0, errPort
	}
	return uint16(value), nil
}

func main() {
	port, err := parsePort("08080")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("포트:", port)
}
```

```sh
go run ./solutions/27-fuzzing/port
```

```text
포트: 8080
```

ASCII 숫자 검사와 숫자 범위 검사는 서로 다른 조건입니다.
숫자처럼 보이는 전각 문자는 이번 형식에서 허용하지 않습니다.
`ParseUint`는 16비트 상한을 검사하지만 값 0의 거부 정책은 별도로 확인합니다.
[ParseUint](https://pkg.go.dev/strconv#ParseUint)의 비트 크기 인자가 이 역할을 합니다.

### 테스트 전체

```go
package main

import (
	"strconv"
	"testing"
)

func TestPort(t *testing.T) {
	for _, tc := range []struct {
		text string
		want uint16
		bad  bool
	}{
		{"1", 1, false}, {"65535", 65535, false}, {"08080", 8080, false},
		{"", 0, true}, {"0", 0, true}, {"65536", 0, true}, {"+80", 0, true}, {" 80", 0, true}, {"８０", 0, true},
	} {
		got, err := parsePort(tc.text)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("%q: %d %v", tc.text, got, err)
		}
	}
}

func FuzzPort(f *testing.F) {
	for _, seed := range []string{"", "0", "1", "65535", "65536", "08080"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 1024 {
			t.Skip()
		}
		got, err := parsePort(text)
		if err != nil {
			return
		}
		if got == 0 {
			t.Fatal("zero accepted")
		}
		canonical := strconv.FormatUint(uint64(got), 10)
		again, err := parsePort(canonical)
		if err != nil || again != got {
			t.Fatal("round trip")
		}
	})
}

func FuzzValidPort(f *testing.F) {
	f.Add(uint16(1))
	f.Add(uint16(65535))
	f.Fuzz(func(t *testing.T, value uint16) {
		if value == 0 {
			return
		}
		got, err := parsePort(strconv.FormatUint(uint64(value), 10))
		if err != nil || got != value {
			t.Fatal("valid value rejected")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 포트: 8080
}
```

첫 타깃은 임의 문자열에서 성공한 경우의 의미 왕복을 검사합니다.
두 번째 타깃은 유효한 숫자를 먼저 만들어 반드시 성공해야 할 입력을 제공합니다.

```sh
go test -run '^$' -fuzz '^FuzzPort$' -fuzztime=3s -parallel=2 ./solutions/27-fuzzing/port
go test -run '^$' -fuzz '^FuzzValidPort$' -fuzztime=3s -parallel=2 ./solutions/27-fuzzing/port
```

0은 포트 설정의 계약에서 거부하므로 유효 입력 전용 타깃에서는 제외합니다.
그러나 거부 경로 자체는 `TestPort`와 임의 문자열 타깃의 시드에 남겨 둡니다.
어떤 입력을 생략했는지와 어디에서 검사하는지를 함께 관리합니다.

`08080`의 정규 문자열은 `8080`입니다.
선행 0이 사라져도 숫자 값이 같으면 이 문제의 왕복 계약을 만족합니다.
원본 표기 보존이 필요하다면 반환 데이터에 원문을 따로 보관하는 다른 설계가 필요합니다.

## 문제 5. 태그 정규화

### 전체 구현

```go
package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var errTag = errors.New("invalid tag")

func normalizeTags(text string) ([]string, error) {
	seen := make(map[string]struct{})
	for _, part := range strings.Split(text, ",") {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		for _, ch := range tag {
			valid := ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-'
			if !valid {
				return nil, errTag
			}
		}
		seen[strings.ToLower(tag)] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for tag := range seen {
		result = append(result, tag)
	}
	slices.Sort(result)
	return result, nil
}

func main() {
	tags, err := normalizeTags(" Go,api, GO, test-1 ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(tags)
}
```

```sh
go run ./solutions/27-fuzzing/normalize-tags
```

```text
[api go test-1]
```

처리 순서는 다음과 같습니다.

1. 쉼표로 입력을 나눕니다.
2. 각 원소의 앞뒤 공백을 제거합니다.
3. 빈 원소는 생략합니다.
4. 원문 태그가 허용된 ASCII 문자로만 이루어졌는지 확인합니다.
5. 소문자로 바꾸어 맵에 담아 중복을 없앱니다.
6. 키를 슬라이스로 옮기고 정렬합니다.

소문자 변환 전에 ASCII 검사를 합니다.
Unicode 대문자가 소문자 변환으로 ASCII 문자처럼 되는 경우까지 허용하지 않기 위해서입니다.
이 정책은 문제의 설정 형식에 맞춘 선택입니다.
다국어 태그를 지원하려면 허용 문자와 정규화 정책을 다시 정해야 합니다.

### 테스트 전체

```go
package main

import (
	"slices"
	"strings"
	"testing"
)

func TestTags(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []string
		bad  bool
	}{
		{"", nil, false}, {" , ", nil, false}, {"Go,api,GO", []string{"api", "go"}, false},
		{"go lang", nil, true}, {"한글", nil, true}, {"İ", nil, true},
	} {
		got, err := normalizeTags(tc.in)
		if (err != nil) != tc.bad || !slices.Equal(got, tc.want) {
			t.Fatalf("got=%v err=%v", got, err)
		}
	}
}

func FuzzTags(f *testing.F) {
	for _, seed := range []string{"", "Go,api,GO", ",,", "한글", "test-1"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 65536 {
			t.Skip()
		}
		got, err := normalizeTags(text)
		if err != nil {
			return
		}
		if !slices.IsSorted(got) {
			t.Fatal("unsorted")
		}
		for i, tag := range got {
			if tag == "" || tag != strings.ToLower(tag) {
				t.Fatal("unnormalized")
			}
			if i > 0 && got[i-1] == tag {
				t.Fatal("duplicate")
			}
		}
		again, err := normalizeTags(strings.Join(got, ","))
		if err != nil || !slices.Equal(got, again) {
			t.Fatal("not idempotent")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// [api go test-1]
}
```

```sh
go test -run '^$' -fuzz '^FuzzTags$' -fuzztime=3s -parallel=2 ./solutions/27-fuzzing/normalize-tags
```

정렬 검사는 맵 순회 순서에 의존하는 결함을 드러냅니다.
인접 원소의 같음을 확인하면 중복이 남은 것도 찾을 수 있습니다.
마지막에는 쉼표로 합쳐 다시 정규화하여 결과가 안정적인지 검사합니다.

nil과 빈 슬라이스는 같은 원소 목록으로 비교합니다.
어떤 JSON 표현을 사용할지까지 요구하는 API라면 nil 여부를 별도의 계약으로 추가해야 합니다.
이번 함수의 테스트는 그 범위를 넘어 단정하지 않습니다.

## 문제 6. 정확한 합과 int 범위

### 전체 구현

```go
package main

import (
	"errors"
	"fmt"
)

var errOverflow = errors.New("integer overflow")

func addChecked(a, b int) (int, error) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if b > 0 && a > maxInt-b || b < 0 && a < minInt-b {
		return 0, errOverflow
	}
	return a + b, nil
}

func main() {
	maxInt := int(^uint(0) >> 1)
	_, err := addChecked(maxInt, 1)
	fmt.Println("범위 오류:", errors.Is(err, errOverflow))
	total, err := addChecked(10, 20)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("합계:", total)
}
```

```sh
go run ./solutions/27-fuzzing/checked-add
```

```text
범위 오류: true
합계: 30
```

b가 양수일 때는 `maxInt-b`보다 a가 큰지 검사합니다.
이 분기에서 `maxInt-b`는 정수 범위 안에 있습니다.
b가 음수일 때는 `minInt-b`보다 a가 작은지 검사합니다.
이때도 경계 계산이 유효한 범위에 있어 먼저 오버플로한 결과를 이용하지 않습니다.

| b의 부호 | 넘을 수 있는 방향 | 검사 |
| --- | --- | --- |
| 양수 | 위쪽 | `a > maxInt-b` |
| 음수 | 아래쪽 | `a < minInt-b` |
| 0 | 없음 | a 그대로 |

### 독립 기준을 사용하는 테스트 전체

```go
package main

import (
	"errors"
	"math/big"
	"testing"
)

func TestAddChecked(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	for _, tc := range []struct {
		a, b, want int
		bad        bool
	}{
		{0, 0, 0, false}, {-1, 1, 0, false}, {maxInt, 0, maxInt, false},
		{maxInt, 1, 0, true}, {minInt, -1, 0, true}, {minInt, maxInt, -1, false},
	} {
		got, err := addChecked(tc.a, tc.b)
		if errors.Is(err, errOverflow) != tc.bad || got != tc.want {
			t.Fatalf("got=%d err=%v", got, err)
		}
	}
}

func FuzzAddChecked(f *testing.F) {
	f.Add(10, 20)
	f.Add(-1, 1)
	f.Add(int(^uint(0)>>1), 1)
	f.Fuzz(func(t *testing.T, a, b int) {
		maxInt := int(^uint(0) >> 1)
		minInt := -maxInt - 1
		exact := new(big.Int).Add(big.NewInt(int64(a)), big.NewInt(int64(b)))
		overflow := exact.Cmp(big.NewInt(int64(minInt))) < 0 || exact.Cmp(big.NewInt(int64(maxInt))) > 0
		got, err := addChecked(a, b)
		if errors.Is(err, errOverflow) != overflow {
			t.Fatal("range classification")
		}
		if !overflow && int64(got) != exact.Int64() {
			t.Fatal("wrong sum")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 범위 오류: true
	// 합계: 30
}
```

```sh
go test -run '^$' -fuzz '^FuzzAddChecked$' -fuzztime=3s -parallel=2 ./solutions/27-fuzzing/checked-add
```

검사 기준은 같은 조건문을 복사한 것이 아니라 `math/big.Int`의 정확한 합입니다.
범위를 넘는지를 큰 정수 비교로 판단하고,
범위 안일 때만 `Int64`로 값을 꺼내 구현 결과와 비교합니다.
검사 기준을 구현과 다른 방식으로 만들면 같은 실수를 양쪽에 복제할 위험을 줄일 수 있습니다.

큰 정수 연산도 무제한 입력을 무조건 허용하는 이유는 아닙니다.
여기서는 원래 입력이 `int` 두 개라 필요한 크기가 작고 정해져 있습니다.
외부 문자열로 거대한 정수를 받는 프로그램이라면 입력 길이와 자원 예산을 별도로 제한해야 합니다.

## 선택 A. 길이 경계와 회귀 관리

`parseRecord`의 전체 바이트 길이는 키·구분자·값을 모두 포함합니다.
따라서 다음 입력은 각각 경계 안과 밖입니다.

| 구성 | 전체 길이 | 결과 |
| --- | ---: | --- |
| `"a=" + strings.Repeat("x", 4094)` | 4096 | 성공 |
| `"a=" + strings.Repeat("x", 4095)` | 4097 | 오류 |

한글 글자 수로 계산하면 바이트 제한과 다른 입력을 만들 수 있습니다.
이번 계약의 단위는 바이트입니다.

바이트 반전 실습의 검증 기록은 다음 내용을 포함하면 됩니다.

- 변경: `[]rune`을 `[]byte`로 바꾸었습니다.
- 실패: 한글 시드에서 중간 결과가 유효하지 않은 UTF-8입니다.
- 수정: 룬 단위 순서 반전으로 되돌렸습니다.
- 회귀: 한글 시드와 알려진 출력 테스트를 유지했습니다.
- 재검증: 일반 테스트와 타깃 하나의 제한 시간 퍼징을 실행했습니다.

실패 메시지의 줄 번호나 자동 생성 파일 이름은 자신의 실행 결과에서 옮깁니다.
검증하지 않은 경로나 실행 횟수를 예상으로 채우지 않습니다.

## 검증 명령

```sh
go fmt ./solutions/27-fuzzing/...
go vet ./solutions/27-fuzzing/...
go test -count=1 ./solutions/27-fuzzing/...
go test -race -count=1 ./solutions/27-fuzzing/...
```

작성 시 Go 1.27.3에서 시드 검사와 다섯 퍼즈 타깃의 제한 시간 탐색을 각각 실행했습니다.
통과 결과는 검사한 조건과 예산 안에서의 관찰입니다.
외부 I/O·운영 트래픽·모든 Unicode 표시 규칙까지 검증했다는 뜻은 아닙니다.

## 참고 자료

[Go Fuzzing](https://go.dev/doc/security/fuzz/) · [testing.F](https://pkg.go.dev/testing#F) · [utf8](https://pkg.go.dev/unicode/utf8) · [strconv.ParseUint](https://pkg.go.dev/strconv#ParseUint) · [math/big.Int](https://pkg.go.dev/math/big#Int)

[본문](../../book/27-fuzzing/README.md) · [연습문제](../../exercises/27-fuzzing/README.md) · [전체 해설](../README.md)
