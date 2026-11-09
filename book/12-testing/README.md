# 12. 자동 테스트

프로그램을 고친 뒤에도 이전에 약속한 동작이 유지되는지 어떻게 확인할까요?
매번 화면을 보고 판단하면 경계값과 오류 경로를 빠뜨리기 쉽습니다.
자동 테스트는 입력과 기대하는 결과를 코드로 남겨 반복해서 확인하는 방법입니다.

이번 장에서는 Go 표준 `testing` 패키지로 정상·경계·오류 입력을 검사합니다.
테이블 테스트와 하위 테스트로 사례를 정리하고,
임시 파일·환경 변수·예제 출력까지 테스트 가능한 경계로 다룹니다.

[이전: 11. 오류 처리와 자원 정리](../11-errors-defer/README.md) · [전체 목차](../README.md) · [연습문제](../../exercises/12-testing/README.md)

## 학습 목표

이 장을 마치면 다음을 할 수 있습니다.

1. `_test.go`와 `TestXxx`를 작성하고 실패 메시지를 읽습니다.
2. 테이블·하위 테스트로 정상값과 경계값, 오류 입력을 구분합니다.
3. 오류 원인·결과값·원본 보존처럼 관찰 가능한 계약을 검사합니다.
4. 임시 디렉터리·정리 함수·환경 변수 복원으로 테스트를 독립시킵니다.
5. 예제 테스트, 선택 실행, 캐시와 커버리지의 범위를 설명합니다.

## 사전 지식과 준비

[함수](../04-functions-scope/README.md), [슬라이스](../05-arrays-slices/README.md),
[맵](../06-maps/README.md), [구조체](../08-pointers-structs/README.md),
[인터페이스](../09-methods-interfaces/README.md), [패키지](../10-packages-modules/README.md),
[오류 처리](../11-errors-defer/README.md)를 사용합니다.

파일 예제의 목적은 테스트 자료와 정리 시점입니다.
io.Reader·임시 파일에 필요한 API를 짧게 소개하며,
일반적인 파일 입출력과 스트림 한도는 [14장](../14-files-io/README.md)에서 자세히 다룹니다.
이 장의 파일 자료는 테스트가 만든 작은 가상 데이터입니다.

명령은 별도 표시가 없다면 **저장소 루트**에서 실행합니다.
PowerShell과 Bash에서 공통으로 쓸 수 있으며 준비 명령만 구분합니다.
Go 버전은 루트의 `go.mod`를 따릅니다. 예제는 Go 1.27.3에서 확인합니다.
외부 서비스·계정·네트워크 연결은 테스트에 필요하지 않습니다.

| 예제 | 중심 주제 | 실행 경로 |
| --- | --- | --- |
| grade | 첫 테스트와 오류 중단 | `./examples/12-testing/grade` |
| table | 입력 경계와 테이블 테스트 | `./examples/12-testing/table` |
| subtests | 중첩 하위 테스트와 Helper | `./examples/12-testing/subtests` |
| errors | 오류 원인과 구체 정보 검사 | `./examples/12-testing/errors` |
| fixtures | 임시 파일과 Cleanup | `./examples/12-testing/fixtures` |
| settings | 의존 함수 전달과 Setenv | `./examples/12-testing/settings` |
| example | 공개 API와 예제 테스트 | `./examples/12-testing/example` |

마지막 프로그램은 같은 장의 `publicapi` 패키지를 사용합니다.
프로그램과 라이브러리가 각각 별도 패키지라는 10장의 구성을 유지합니다.

## 1. 첫 자동 테스트

### 1.1 먼저 계약을 적기

점수는 0~100의 정수입니다.
80 이상은 우수, 60 이상은 통과, 나머지는 보충으로 판정합니다.
범위를 벗어나면 빈 문자열과 오류를 반환합니다.

| 입력 | 기대 결과 | 이유 |
| --- | --- | --- |
| 80 | 우수 | 우수의 시작 경계 |
| 60 | 통과 | 통과의 시작 경계 |
| 0 | 보충 | 허용 범위의 최솟값 |
| -1 | 오류 | 허용 범위 바깥 |
| 101 | 오류 | 허용 범위 바깥 |

이 표는 코드의 조건식을 복사한 정답이 아니라 요구사항에서 정한 결과입니다.
구현과 기대값을 같은 잘못된 식으로 계산하면 테스트가 그 실수를 그대로 반복할 수 있습니다.

파일: [grade/main.go](../../examples/12-testing/grade/main.go).

```go
package main

import (
	"errors"
	"fmt"
)

func grade(score int) (string, error) {
	if score < 0 || score > 100 {
		return "", errors.New("score must be 0..100")
	}
	switch {
	case score >= 80:
		return "우수", nil
	case score >= 60:
		return "통과", nil
	default:
		return "보충", nil
	}
}

func main() {
	result, err := grade(80)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("판정:", result)
}
```

실행:

```sh
go run ./examples/12-testing/grade
```

예상 출력:

```text
판정: 우수
```

### 1.2 테스트 파일과 함수

같은 폴더에 `main_test.go`를 둡니다.
파일 이름의 `_test.go` 접미사는 Go 도구가 테스트용 파일을 구분하는 규칙입니다.
일반 테스트 함수는 `func TestXxx(t *testing.T)` 형태를 사용합니다.
Test 뒤 이름은 소문자로 시작하지 않게 정합니다.
[testing 개요](https://pkg.go.dev/testing#pkg-overview)를 참고하세요.

파일: [grade/main_test.go](../../examples/12-testing/grade/main_test.go).

```go
package main

import "testing"

func TestGrade(t *testing.T) {
	got, err := grade(80)
	if err != nil {
		t.Fatalf("grade(80) error: %v", err)
	}
	if got != "우수" {
		t.Errorf("grade(80) = %q; want %q", got, "우수")
	}
}

func TestGradeInvalid(t *testing.T) {
	got, err := grade(-1)
	if err == nil {
		t.Fatal("grade(-1) accepted an invalid score")
	}
	if got != "" {
		t.Errorf("result on error = %q; want empty", got)
	}
}

func Example_main() {
	main()
	// Output:
	// 판정: 우수
}
```

테스트는 보통 다음 순서로 읽습니다.

1. 입력과 기대값을 준비합니다.
2. 검사할 함수를 호출합니다.
3. 오류나 결과값이 계약과 맞는지 확인합니다.
4. 다르면 어떤 입력에서 무엇이 달랐는지 실패 메시지로 남깁니다.

`got`은 실제로 받은 값, `want`는 기대한 값을 나타내는 흔한 이름입니다.
이 이름 자체에 특별한 문법 기능은 없습니다.
`%q`는 문자열의 따옴표·공백·개행을 읽기 쉽게 보여 줍니다.

### 1.3 테스트 실행

```sh
go test -count=1 ./examples/12-testing/grade
```

성공 출력은 패키지 경로와 실행 시간에 따라 달라지지만 `ok`로 시작합니다.
실패하면 `FAIL`과 실패한 테스트 이름, 파일 위치와 메시지를 확인합니다.
`go run`은 프로그램을 실행하고, `go test`는 테스트용 실행 파일을 만들어 검사합니다.
패키지의 `main`이 테스트마다 자동으로 실행되는 것은 아닙니다.
위의 Example_main은 자신이 명시적으로 main을 호출한 경우입니다.

### 1.4 Errorf와 Fatalf

| 호출 | 테스트 표시 | 현재 테스트의 뒤 코드 |
| --- | --- | --- |
| `t.Errorf(...)` | 실패 | 계속 실행 |
| `t.Fatalf(...)` | 실패 | 중단 |
| `t.Logf(...)` | 기록 | 계속 실행 |

필수 준비나 반환 오류 때문에 다음 비교가 의미 없거나 위험하다면 Fatalf를 사용합니다.
독립된 여러 결과를 비교할 수 있다면 Errorf로 더 많은 차이를 모을 수 있습니다.
Fatalf는 전체 저장소의 모든 테스트 프로세스를 종료하는 명령이 아닙니다.
해당 테스트 실행을 중단하며 다른 테스트는 검사할 수 있습니다.

Fatal·FailNow 계열은 현재 테스트 함수를 실행하는 고루틴에서 호출해야 합니다.
나중에 고루틴 테스트를 만들 때 작업 고루틴에서 직접 Fatal을 호출하지 마세요.
결과를 테스트 쪽으로 전달한 뒤 판정하는 방법은 동시성 장에서 다룹니다.
[testing.T.FailNow](https://pkg.go.dev/testing#T.FailNow)의 조건을 확인합니다.

## 2. 테이블 테스트로 경계 정리하기

하나의 정상값만 확인하면 조건문의 경계가 틀려도 놓칠 수 있습니다.
배송비 예제는 무게 1000g·5000g의 양쪽을 검사합니다.

| 무게 | 배송비 |
| --- | ---: |
| 1~1000g | 3000 |
| 1001~5000g | 5000 |
| 5001~10000g | 8000 |
| 그 외 | 오류 |

파일: [table/main.go](../../examples/12-testing/table/main.go).

```go
package main

import (
	"errors"
	"fmt"
)

func shippingFee(grams int) (int, error) {
	if grams < 1 || grams > 10000 {
		return 0, errors.New("grams must be 1..10000")
	}
	switch {
	case grams <= 1000:
		return 3000, nil
	case grams <= 5000:
		return 5000, nil
	default:
		return 8000, nil
	}
}

func main() {
	fee, err := shippingFee(1200)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("배송비:", fee)
}
```

실행:

```sh
go run ./examples/12-testing/table
```

예상 출력:

```text
배송비: 5000
```

### 2.1 입력·기대값·오류 여부를 한 줄에

테이블 테스트(table-driven test)는 사례를 구조체 슬라이스로 정리하고 같은 검사 코드를 적용합니다.
새 사례를 넣을 때 함수 호출·실패 메시지 전체를 복사할 필요가 없습니다.
테이블의 각 행은 특정한 계약을 확인해야 합니다.
숫자를 많이 늘리는 것보다 경계와 서로 다른 오류 원인을 고르는 것이 중요합니다.

파일: [table/main_test.go](../../examples/12-testing/table/main_test.go).

```go
package main

import "testing"

func TestShippingFee(t *testing.T) {
	tests := []struct {
		name        string
		grams, want int
		wantErr     bool
	}{
		{"minimum", 1, 3000, false}, {"first_edge", 1000, 3000, false},
		{"after_first", 1001, 5000, false}, {"second_edge", 5000, 5000, false},
		{"after_second", 5001, 8000, false}, {"maximum", 10000, 8000, false},
		{"zero", 0, 0, true}, {"negative", -1, 0, true}, {"too_heavy", 10001, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := shippingFee(tc.grams)
			if (err != nil) != tc.wantErr {
				t.Fatalf("shippingFee(%d) error = %v; wantErr %v", tc.grams, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("shippingFee(%d) = %d; want %d", tc.grams, got, tc.want)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 배송비: 5000
}
```

`wantErr`는 오류가 있어야 하는지를 나타냅니다.
`(err != nil) != tc.wantErr`는 실제 오류 여부와 기대가 다른지 비교합니다.
이 검사가 통과한 뒤 결과값을 검사하여 실패 원인을 분명히 합니다.

### 2.2 왜 1000과 1001을 모두 넣는가

1000만 검사하면 그 값을 잘 처리하는지 알 수 있습니다.
1001도 검사해야 그다음 구간으로 이동하는지 알 수 있습니다.
같은 원리로 5000·5001, 최소 1·불허 0, 최대 10000·불허 10001을 선택합니다.

- 정상 대표값: 구간 안에서 일반적인 입력.
- 경계값: 허용 구간이나 분기 조건의 시작·끝.
- 오류값: 계약 바깥의 입력.

정상 입력이 모두 성공하더라도 잘못된 입력을 받아들일 수 있습니다.
오류 입력이 모두 거부되더라도 정상 입력까지 거부하는 구현일 수 있습니다.
성공과 실패의 계약을 함께 검사하세요.

### 2.3 하위 테스트 이름

`t.Run(name, func(t *testing.T) { ... })`은 이름 있는 하위 테스트를 만듭니다.
실패 결과에서 `TestShippingFee/first_edge`처럼 문제를 일으킨 사례를 찾을 수 있습니다.
이름은 순번만 적기보다 확인하려는 조건을 담는 편이 유용합니다.

```sh
go test -v -count=1 ./examples/12-testing/table
```

`-v`를 붙이면 통과한 사례의 이름도 표시됩니다.
소요 시간과 들여쓰기 일부는 환경에 따라 달라질 수 있습니다.
테이블 테스트의 구성은 [Go 공식 하위 테스트 소개](https://go.dev/blog/subtests)에서도 확인할 수 있습니다.

## 3. 하위 테스트와 공통 검사 함수

시각 문자열은 정확히 HH:MM 형식의 ASCII 숫자로 받습니다.
시각 범위는 00:00~23:59이고 결과는 자정부터 지난 분입니다.
이는 날짜·시간대 계산이 아니라 정해진 짧은 입력 형식을 검증하는 예입니다.

파일: [subtests/main.go](../../examples/12-testing/subtests/main.go).

```go
package main

import (
	"errors"
	"fmt"
)

func parseClock(text string) (int, error) {
	if len(text) != 5 || text[2] != ':' {
		return 0, errors.New("expected HH:MM")
	}
	for _, i := range []int{0, 1, 3, 4} {
		if text[i] < '0' || text[i] > '9' {
			return 0, errors.New("expected ASCII digits")
		}
	}
	hours := int(text[0]-'0')*10 + int(text[1]-'0')
	minutes := int(text[3]-'0')*10 + int(text[4]-'0')
	if hours > 23 || minutes > 59 {
		return 0, errors.New("time out of range")
	}
	return hours*60 + minutes, nil
}

func main() {
	minutes, err := parseClock("09:30")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("경과 분:", minutes)
}
```

실행:

```sh
go run ./examples/12-testing/subtests
```

예상 출력:

```text
경과 분: 570
```

### 3.1 검사 전에 길이를 확인하기

길이가 5인지 확인하기 전에는 text[2]에 접근하지 않아야 합니다.
`||`의 왼쪽 조건이 참이면 오른쪽을 평가하지 않는 단락 평가를 사용했습니다.
숫자 위치를 검사한 뒤에만 '0'을 빼서 숫자 값을 계산합니다.

시각을 `int`로 바꾸는 계산의 최대값은 23×60+59=1439입니다.
이 범위는 앞에서 배운 int의 범위 안입니다.
큰 수나 임의 문자열의 일반 숫자 파서를 직접 구현한 것은 아닙니다.

### 3.2 valid와 invalid로 사례 묶기

파일: [subtests/main_test.go](../../examples/12-testing/subtests/main_test.go).

```go
package main

import "testing"

func requireMinutes(t *testing.T, text string, want int) {
	t.Helper()
	got, err := parseClock(text)
	if err != nil {
		t.Fatalf("parseClock(%q): %v", text, err)
	}
	if got != want {
		t.Errorf("parseClock(%q) = %d; want %d", text, got, want)
	}
}

func TestParseClock(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, tc := range []struct {
			name, text string
			want       int
		}{
			{"midnight", "00:00", 0}, {"morning", "09:30", 570}, {"last_minute", "23:59", 1439},
		} {
			t.Run(tc.name, func(t *testing.T) { requireMinutes(t, tc.text, tc.want) })
		}
	})
	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct{ name, text string }{
			{"short", "9:30"}, {"separator", "09-30"}, {"hour", "24:00"},
			{"minute", "12:60"}, {"sign", "+1:00"}, {"space", " 9:30"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got, err := parseClock(tc.text)
				if err == nil || got != 0 {
					t.Fatalf("parseClock(%q) = %d, %v; want rejection", tc.text, got, err)
				}
			})
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 경과 분: 570
}
```

정상 사례의 이름은 `TestParseClock/valid/morning`처럼 계층을 이룹니다.
하위 테스트를 지나치게 깊게 만들 필요는 없습니다.
분류가 실패 원인을 찾거나 일부 실행에 도움이 될 때 사용합니다.

### 3.3 Helper의 역할

`requireMinutes`는 반복되는 호출과 비교를 묶은 테스트 보조 함수입니다.
첫 줄의 `t.Helper()`는 실패 위치를 보고할 때 보조 함수의 프레임을 건너뛰도록 표시합니다.
독자가 고쳐야 할 사례의 호출 위치를 더 쉽게 찾게 합니다.
검사 자체의 성공·실패 규칙을 바꾸는 기능은 아닙니다.
[testing.T.Helper](https://pkg.go.dev/testing#T.Helper)를 참고하세요.

보조 함수에 t를 넘겼다면 그 테스트와 수명을 공유한다는 점을 기억합니다.
테스트가 끝난 뒤에 이 t를 잡고 새 로그나 실패를 남기는 식으로 사용하지 않습니다.

### 3.4 한 사례만 실행하기

PowerShell과 Bash 모두 정규식을 작은따옴표로 감쌉니다.

```sh
go test -v -count=1 -run '^TestParseClock$/^valid$/^morning$' ./examples/12-testing/subtests
```

출력에는 부모 이름과 valid/morning 경로가 나타납니다.
부모 테스트는 선택한 자식 테스트를 찾기 위해 실행됩니다.
`-run`의 슬래시 단위 패턴을 단순 문자열 접두어와 혼동하지 마세요.
테스트 이름에 슬래시가 있다면 별도의 계층처럼 선택될 수 있으므로 이름 설계에도 주의합니다.
[go test 플래그](https://pkg.go.dev/cmd/go#hdr-Testing_flags)에 선택 규칙이 있습니다.

## 4. 오류를 문자열만으로 검사하지 않기

11장에서 배운 오류 래핑을 테스트에 적용합니다.
수량 입력은 ASCII 십진 숫자로만 구성된 1~100이며 앞의 0은 허용합니다.
잘못된 입력에는 공통 원인과 원래 입력값을 담아 반환합니다.

파일: [errors/main.go](../../examples/12-testing/errors/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

var errCount = errors.New("count must be 1..100 decimal digits")

type countError struct{ Input string }

func (e *countError) Error() string { return fmt.Sprintf("count %q: %v", e.Input, errCount) }
func (e *countError) Unwrap() error { return errCount }

func parseCount(text string) (int, error) {
	fail := func() (int, error) { return 0, &countError{Input: text} }
	if text == "" {
		return fail()
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fail()
		}
	}
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 || n > 100 {
		return fail()
	}
	return n, nil
}

func main() {
	n, err := parseCount("12")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("수량:", n)
}
```

실행:

```sh
go run ./examples/12-testing/errors
```

예상 출력:

```text
수량: 12
```

### 4.1 원인과 상세 정보 검사

파일: [errors/main_test.go](../../examples/12-testing/errors/main_test.go).

```go
package main

import (
	"errors"
	"testing"
)

func TestParseCount(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int
	}{{"1", 1}, {"0012", 12}, {"100", 100}} {
		got, err := parseCount(tc.text)
		if err != nil || got != tc.want {
			t.Errorf("parseCount(%q) = %d, %v; want %d, nil", tc.text, got, err, tc.want)
		}
	}
}

func TestParseCountError(t *testing.T) {
	for _, text := range []string{"", "0", "101", "-1", "+1", " 1", "１", "999999999999999999999"} {
		t.Run(text, func(t *testing.T) {
			got, err := parseCount(text)
			if !errors.Is(err, errCount) {
				t.Fatalf("error = %v; want errCount", err)
			}
			var detail *countError
			if !errors.As(err, &detail) {
				t.Fatalf("error type = %T; want *countError", err)
			}
			if got != 0 || detail.Input != text {
				t.Errorf("got=%d input=%q; want 0, %q", got, detail.Input, text)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 수량: 12
}
```

첫 비교는 `errors.Is`로 약속한 원인을 검사합니다.
그다음 `errors.As`로 `*countError`가 담겼는지 확인합니다.
As가 성공한 뒤에만 detail.Input을 읽습니다.
이 순서가 없으면 다른 오류를 반환한 경우 nil 포인터 접근으로 실패 원인을 흐릴 수 있습니다.

에러 메시지에 어떤 문장을 덧붙이더라도 원인과 Input 계약을 유지한다면 테스트가 허용합니다.
반대로 사용자에게 약속한 명령행 문구 자체가 계약이라면 출력 문자열을 검사할 수도 있습니다.
무엇을 공개 계약으로 삼는지에 따라 검사 대상을 고르세요.
[errors.Is와 As](https://pkg.go.dev/errors)의 의미를 기준으로 합니다.

### 4.2 실패 결과도 검사하기

오류가 있는지만 확인하면 잘못된 값이 함께 반환되는 것을 놓칠 수 있습니다.
이 함수는 실패 때 결과 0을 약속했으므로 함께 검사합니다.
부분 결과를 허용하는 함수라면 그 부분 결과의 의미를 별도로 테스트해야 합니다.
오류가 있으면 모든 Go 함수의 반환값이 자동으로 0이 되는 규칙은 없습니다.

## 5. 임시 자료와 자원 정리

테스트가 특정 컴퓨터에 있는 파일을 전제로 하면 다른 환경에서 실행하기 어렵습니다.
작은 가상 자료를 테스트 안에서 만들고, 끝나면 정리하도록 구성합니다.

이번 함수는 작은 Reader를 끝까지 읽어 라벨을 만듭니다.
앞뒤 공백을 정리한 뒤 1~20룬의 유효한 UTF-8 한 줄이어야 합니다.
중간 CR·LF는 허용하지 않습니다. 빈 줄만 있는 자료는 빈 라벨이므로 오류입니다.
이 예제는 신뢰하는 작은 테스트 입력을 위한 것이며 일반 외부 스트림의 크기 제한은 14장에서 배웁니다.

파일: [fixtures/main.go](../../examples/12-testing/fixtures/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

func readLabel(input io.Reader) (string, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return "", fmt.Errorf("read label: %w", err)
	}
	label := strings.TrimSpace(string(data))
	if !utf8.ValidString(label) || label == "" || utf8.RuneCountInString(label) > 20 || strings.ContainsAny(label, "\r\n") {
		return "", errors.New("label must be one line of 1..20 UTF-8 runes")
	}
	return label, nil
}

func main() {
	label, err := readLabel(strings.NewReader("  한글\n"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("라벨:", label)
}
```

실행:

```sh
go run ./examples/12-testing/fixtures
```

예상 출력:

```text
라벨: 한글
```

### 5.1 이번 예제에서 쓰는 입출력 API

| API | 이번 장에서의 역할 |
| --- | --- |
| `io.Reader` | 바이트를 읽을 수 있는 행동을 가진 인터페이스 |
| `strings.NewReader` | 문자열을 Reader로 제공 |
| `io.ReadAll` | Reader에서 EOF까지 읽고 바이트와 오류 반환 |
| `os.WriteFile` | 테스트용 작은 파일 작성 |
| `os.Open` | 파일을 열어 읽을 수 있는 값 반환 |
| `file.Close` | 열린 파일 자원 정리 |
| `filepath.Join` | 운영체제에 맞는 파일 경로 조합 |

readLabel은 Reader를 빌려 읽으며 닫지 않습니다.
파일을 연 테스트 보조 함수가 닫기 책임을 가집니다.
이렇게 역할을 나누면 문자열·파일·오류를 내는 대역을 같은 함수에 전달할 수 있습니다.
[io.Reader](https://pkg.go.dev/io#Reader)와 [os](https://pkg.go.dev/os)의 계약을 참고하세요.

### 5.2 TempDir와 Cleanup

파일: [fixtures/main_test.go](../../examples/12-testing/fixtures/main_test.go).

```go
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func openFixture(t *testing.T, text string) *os.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), "label.txt")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close fixture: %v", err)
		}
	})
	return file
}

func TestReadLabel(t *testing.T) {
	file := openFixture(t, "  한글\n")
	got, err := readLabel(file)
	if err != nil || got != "한글" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func TestReadLabelBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		bad        bool
	}{
		{"twenty", strings.Repeat("가", 20), false}, {"twenty_one", strings.Repeat("가", 21), true},
		{"empty", " \n", true}, {"multiple_lines", "a\nb", true}, {"invalid_utf8", string([]byte{0xff}), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readLabel(openFixture(t, tc.text))
			if (err != nil) != tc.bad {
				t.Fatalf("got=%q err=%v", got, err)
			}
			if tc.bad && got != "" {
				t.Error("result on failure must be empty")
			}
		})
	}
}

type failedReader struct{ err error }

func (r failedReader) Read(p []byte) (int, error) { return 0, r.err }

func TestReadLabelReadError(t *testing.T) {
	cause := errors.New("disk read failed")
	got, err := readLabel(failedReader{cause})
	if got != "" || !errors.Is(err, cause) {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 라벨: 한글
}
```

`t.TempDir()`는 해당 테스트가 사용할 고유한 임시 디렉터리를 만듭니다.
테스트와 그 하위 테스트가 끝나면 자동으로 제거됩니다.
소스 폴더나 사용자의 홈 디렉터리에 임시 파일을 남기지 않아도 됩니다.

`t.Cleanup`은 테스트와 하위 테스트의 종료 뒤 실행할 정리 함수를 등록합니다.
여러 개를 등록하면 나중에 등록한 함수부터 실행합니다.
여기서는 TempDir 생성 뒤 file.Close를 등록하므로 파일을 닫고 디렉터리를 정리합니다.
[TempDir](https://pkg.go.dev/testing#T.TempDir)와 [Cleanup](https://pkg.go.dev/testing#T.Cleanup)의 수명 규칙을 확인합니다.

보조 함수 안에서 `defer file.Close()`를 사용하면 보조 함수가 반환할 때 이미 파일이 닫힙니다.
호출한 테스트가 파일을 읽어야 하므로 테스트 수명에 맞춘 Cleanup이 필요합니다.
일반 함수의 지역 정리에는 defer를, 테스트에 넘겨 사용할 자원의 정리에는 Cleanup을 고려하세요.

### 5.3 실제 오류 장치를 만들 필요는 없다

failedReader는 Read 호출에 정해진 오류를 반환하는 작은 테스트 대역입니다.
실제 디스크를 고장 내지 않고 함수가 읽기 오류를 보존하는지 확인합니다.
하나의 작은 인터페이스를 구현하여 필요한 실패를 재현했습니다.
모든 의존성을 큰 가짜 프레임워크로 바꿀 필요는 없습니다.

## 6. 테스트 사이에 상태를 남기지 않기

같은 함수가 환경 변수에 따라 달라진다면 테스트 입력을 어떻게 고정할까요?
이 예제는 “이름을 읽는 동작”을 함수값으로 받습니다.
실제 실행에서는 os.Getenv를 넘길 수 있고 테스트에서는 고정 값을 주는 함수를 넘길 수 있습니다.

파일: [settings/main.go](../../examples/12-testing/settings/main.go).

```go
package main

import (
	"fmt"
	"strings"
)

const nameKey = "GOLANG_BOOK12_NAME"

func readName(getenv func(string) string) string {
	name := strings.TrimSpace(getenv(nameKey))
	if name == "" {
		return "방문자"
	}
	return name
}

func main() {
	name := readName(func(key string) string { return " 학습자 " })
	fmt.Println("이름:", name)
}
```

실행:

```sh
go run ./examples/12-testing/settings
```

예상 출력:

```text
이름: 학습자
```

이 프로그램의 main은 고정 자료를 주는 함수를 사용하므로 개인 환경에 따라 출력이 변하지 않습니다.
실제 환경 변수 어댑터는 별도 테스트에서 확인합니다.

파일: [settings/main_test.go](../../examples/12-testing/settings/main_test.go).

```go
package main

import (
	"os"
	"testing"
)

func TestReadName(t *testing.T) {
	for _, tc := range []struct{ value, want string }{{"지수", "지수"}, {" 민수 ", "민수"}, {"", "방문자"}, {"  ", "방문자"}} {
		t.Run(tc.want+tc.value, func(t *testing.T) {
			got := readName(func(key string) string {
				if key != nameKey {
					t.Fatalf("key=%q; want %q", key, nameKey)
				}
				return tc.value
			})
			if got != tc.want {
				t.Errorf("got=%q; want %q", got, tc.want)
			}
		})
	}
}

func TestReadNameFromEnvironment(t *testing.T) {
	t.Setenv(nameKey, " 지수 ")
	if got := readName(os.Getenv); got != "지수" {
		t.Errorf("got=%q; want 지수", got)
	}
}

func Example_main() {
	main()
	// Output:
	// 이름: 학습자
}
```

### 6.1 Setenv의 복원

`t.Setenv`는 환경 변수를 설정하고 테스트 뒤 원래 상태로 복원합니다.
기존에 없던 값인지, 어떤 값이 있었는지를 테스트 코드가 직접 보관할 필요가 없습니다.
환경 변수는 프로세스 전체에 영향을 주므로 병렬 테스트나 병렬 부모를 가진 테스트에서 사용할 수 없습니다.
[testing.T.Setenv](https://pkg.go.dev/testing#T.Setenv)의 제약입니다.

`t.Parallel()`을 모든 테스트에 기계적으로 붙이지 마세요.
입력 슬라이스, 전역 상태, 환경 변수, 작업 디렉터리와 외부 자원을 공유하는지 먼저 검토합니다.
이 장의 테스트는 병렬 실행을 직접 요청하지 않습니다.

### 6.2 의존 함수 전달의 범위

getenv 함수는 설정을 읽는 경계 하나를 바꿔 끼웁니다.
전체 운영체제나 환경 변수 체계를 흉내 내지 않습니다.
이 정도의 작은 경계로 대부분의 값 처리 사례를 외부 상태 없이 검사할 수 있습니다.
실제 어댑터와의 연결도 한 번 확인하여 함수 전달과 환경 변수 접근을 함께 검증합니다.

## 7. 공개 API와 예제 테스트

### 7.1 패키지 밖의 사용자처럼 테스트하기

앞의 테스트는 제품 코드와 같은 `package main`에 있었습니다.
같은 패키지의 테스트는 공개하지 않은 이름도 사용할 수 있습니다.
라이브러리의 공개 API를 확인할 때는 `package publicapi_test`처럼 별도 테스트 패키지를 사용할 수 있습니다.

파일: [publicapi/greeting.go](../../examples/12-testing/publicapi/greeting.go).

```go
package publicapi

import "strings"

// Greeting returns a greeting, using 방문자 for an empty or whitespace-only name.
func Greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "방문자"
	}
	return name + "님, 안녕하세요."
}
```

이 패키지는 공개된 Greeting만 제공합니다.
외부 테스트는 모듈의 import 경로로 패키지를 가져와 사용합니다.
10장에서 배운 대문자 공개 규칙이 그대로 적용됩니다.

파일: [publicapi/greeting_test.go](../../examples/12-testing/publicapi/greeting_test.go).

```go
package publicapi_test

import (
	"fmt"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/12-testing/publicapi"
)

func TestGreeting(t *testing.T) {
	for _, tc := range []struct{ name, want string }{{" 지수 ", "지수님, 안녕하세요."}, {"", "방문자님, 안녕하세요."}, {" \n", "방문자님, 안녕하세요."}} {
		if got := publicapi.Greeting(tc.name); got != tc.want {
			t.Errorf("Greeting(%q)=%q; want %q", tc.name, got, tc.want)
		}
	}
}

func ExampleGreeting() {
	fmt.Println(publicapi.Greeting("지수"))
	// Output:
	// 지수님, 안녕하세요.
}

func ExampleGreeting_empty() {
	fmt.Println(publicapi.Greeting(" "))
	// Output:
	// 방문자님, 안녕하세요.
}
```

### 7.2 실행 예제와 출력 비교

`ExampleGreeting`은 Greeting을 사용하는 문서 예제이면서 출력 검사입니다.
`// Output:` 뒤의 주석이 기대 출력입니다.
동작이 바뀌면 예제와 설명도 오래된 상태인지 확인할 수 있습니다.

예제 출력 비교는 앞뒤 공백을 무시합니다.
내부 공백·줄 구성까지 포함한 바이트 수준의 출력 계약이 필요하다면 bytes.Buffer를 받아 직접 비교하는 테스트를 작성합니다.
출력 주석이 없는 Example은 컴파일되지만 실행되지 않습니다.
[예제 함수 규칙](https://pkg.go.dev/testing#hdr-Examples)에 자세한 조건이 있습니다.

| 예제 이름 | 연결 대상 |
| --- | --- |
| `Example` | 패키지 전체 |
| `ExampleGreeting` | Greeting 함수 |
| `ExampleGreeting_empty` | Greeting의 별도 사례 |
| `ExampleType_Method` | Type의 Method |

추가 사례의 접미사는 소문자로 시작합니다.
`Unordered output:`은 행 순서를 무시하는 별도 옵션이지만,
순서가 프로그램의 계약이라면 이 옵션으로 잘못된 순서를 숨기지 않습니다.

### 7.3 실제 프로그램에서도 같은 패키지 사용

파일: [example/main.go](../../examples/12-testing/example/main.go).

```go
package main

import (
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/12-testing/publicapi"
)

func main() { fmt.Println(publicapi.Greeting("지수")) }
```

실행:

```sh
go run ./examples/12-testing/example
```

예상 출력:

```text
지수님, 안녕하세요.
```

공개 API를 호출하는 프로그램, 같은 API를 호출하는 외부 테스트와 예제 테스트가 같은 결과를 공유합니다.
테스트를 위해 제품 코드를 불필요하게 공개하지는 않습니다.
패키지 사용자에게 필요한 기능인지 먼저 판단하세요.

## 8. 테스트 명령을 목적에 맞게 고르기

### 8.1 전체와 일부 실행

| 명령 | 목적 |
| --- | --- |
| `go test ./...` | 모듈의 선택된 모든 패키지 검사 |
| `go test ./examples/12-testing/table` | 한 패키지 검사 |
| `go test -v ./examples/12-testing/table` | 사례별 자세한 출력 |
| `go test -count=1 ./examples/12-testing/...` | 성공 결과 캐시를 사용하지 않고 재실행 |
| `go test -run '^TestShippingFee$' ./examples/12-testing/table` | 이름에 맞는 테스트 실행 |
| `go test -shuffle=on -count=1 ./examples/12-testing/...` | 실행 순서 의존성 점검 |

`./...`는 현재 위치 아래의 패키지를 선택하는 패턴입니다.
한 패키지의 `.go` 파일을 임의로 나열해 실행하면 필요한 다른 파일이나 테스트를 빠뜨릴 수 있습니다.
기본적으로 패키지 경로를 지정하세요.

### 8.2 캐시 표시 읽기

패키지 목록 방식의 go test는 조건이 맞으면 성공 결과를 재사용할 수 있습니다.
출력의 `(cached)`는 이번에 테스트 함수를 다시 실행했다는 뜻이 아닙니다.
실제 재실행이 필요한 검증 기록에는 `-count=1`을 사용합니다.
이 옵션이 컴파일 캐시까지 모두 지우는 것은 아닙니다.
[go test 실행과 캐시](https://pkg.go.dev/cmd/go#hdr-Test_packages)를 참고하세요.

`-run`이 아무 테스트에도 맞지 않아도 명령이 성공할 수 있습니다.
테스트 이름이 맞는지 `-v` 출력과 `go test -list`를 확인합니다.

```sh
go test -list . ./examples/12-testing/subtests
```

이 명령은 최상위 테스트·예제 이름을 보여 주며 모든 하위 사례 이름을 나열하는 명령은 아닙니다.
원하는 하위 사례는 앞에서 본 `-v -run`으로 확인합니다.

### 8.3 순서 의존성 찾기

`-shuffle=on`은 테스트 순서를 섞고 사용한 seed를 출력합니다.
실패를 재현할 때는 표시된 숫자를 `-shuffle=숫자`로 다시 사용할 수 있습니다.
테스트가 앞의 테스트가 남긴 전역 값이나 파일을 기대한다면 순서에 따라 실패할 수 있습니다.

이 옵션이 입력값 자체를 새로 만들어 주는 것은 아닙니다.
또한 모든 하위 테스트 테이블의 순서를 자동으로 무작위화한다고 가정하지 않습니다.
순서 점검은 테스트가 스스로 준비·정리할 수 있는지 확인하는 보조 수단입니다.
새 입력 탐색은 27장의 퍼징과 구분합니다.

## 9. 커버리지를 읽는 방법

커버리지(coverage)는 테스트 실행 중 어떤 코드 구문이 실행되었는지 보는 자료입니다.
높은 숫자만으로 기대값이 맞는지, 중요한 경계를 검사했는지까지 증명되지는 않습니다.

```sh
go test -cover ./examples/12-testing/...
```

함수별 보고서를 만들려면 다음 명령을 사용합니다.
기존 같은 이름의 보고서가 있으면 새 검사 결과로 교체합니다.

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/12-testing | Out-Null
go test -count=1 -coverprofile=./_practice/12-testing/coverage.out ./examples/12-testing/...
go tool cover -func=./_practice/12-testing/coverage.out
```

Bash:

```sh
mkdir -p _practice/12-testing
go test -count=1 -coverprofile=./_practice/12-testing/coverage.out ./examples/12-testing/...
go tool cover -func=./_practice/12-testing/coverage.out
```

필요하면 `go tool cover -html=./_practice/12-testing/coverage.out`으로 보고서를 엽니다.
GUI·브라우저가 없는 환경에서는 함수별 텍스트 보고서로 확인할 수 있습니다.
[Go 커버리지 소개](https://go.dev/blog/cover)를 참고하세요.

### 9.1 숫자보다 질문을 먼저

- 오류 분기가 전혀 실행되지 않았다면 어떤 입력이 빠졌나요?
- 같은 줄을 실행하더라도 <=와 <의 경계를 구분했나요?
- 값이 틀려도 실패하지 않는 테스트는 없나요?
- 부분 결과·원본 훼손·자원 누수처럼 반환값 밖의 계약도 필요한가요?

단순 getter나 출력문만 추가로 호출해 숫자를 높이는 것보다,
실제 요구사항을 깨뜨리는 결함을 찾을 수 있는 검사가 유용합니다.
커버리지를 달성하려고 모든 코드 구조를 테스트에 그대로 복제하지 않습니다.

## 10. 직접 실패시키고 수정하기

테스트는 성공 결과를 보는 것뿐 아니라 잘못된 구현에서 실패하는지도 확인할 수 있습니다.
공식 예제는 유지하고 개인 폴더에 배송비 예제와 테스트를 복사합니다.

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/12-testing/table | Out-Null
foreach ($file in @("main.go", "main_test.go")) {
    if (-not (Test-Path "_practice/12-testing/table/$file")) {
        Copy-Item "examples/12-testing/table/$file" "_practice/12-testing/table/$file"
    }
}
```

Bash:

```sh
mkdir -p _practice/12-testing/table
for file in main.go main_test.go; do
    if [ ! -e "_practice/12-testing/table/$file" ]; then
        cp "examples/12-testing/table/$file" "_practice/12-testing/table/$file"
    fi
done
```

### 10.1 정상 상태 확인

```sh
go test -count=1 ./_practice/12-testing/table
```

그다음 개인 main.go의 첫 배송비 조건을 의도적으로 바꿉니다.
아래는 **의도적으로 잘못된 발췌**입니다.

```go
case grams < 1000:
    return 3000, nil
```

요구사항에서 1000g은 첫 구간에 포함되므로 이 수정은 결함입니다.

### 10.2 경계 테스트만 실행

```sh
go test -v -count=1 -run '^TestShippingFee$/^first_edge$' ./_practice/12-testing/table
```

실패 메시지에서 다음 내용을 확인합니다.
파일의 줄 번호·패키지 경로·시간은 개인 환경에 따라 달라집니다.

```text
shippingFee(1000) = 5000; want 3000
```

main 프로그램은 1200g을 사용하므로 여전히 배송비 5000을 출력할 수 있습니다.
대표 실행 한 번으로 보이지 않던 경계 결함을 테스트가 드러낸 것입니다.

### 10.3 수정 후 다시 확인

조건을 `grams <= 1000`으로 되돌리고 같은 테스트를 실행합니다.
그다음 개인 패키지의 전체 테스트도 실행합니다.
한 경계를 고치면서 다른 경계가 바뀌지 않았는지 확인해야 합니다.

```sh
go test -count=1 ./_practice/12-testing/table
```

`_practice`는 Go의 `./...` 탐색에 자동으로 포함되지 않습니다.
개인 테스트를 검사할 때는 위처럼 경로를 명시합니다.

## 11. 흔한 실수와 확인 방법

| 증상 | 원인 후보 | 확인·수정 |
| --- | --- | --- |
| 테스트가 하나도 실행되지 않음 | 파일·함수 이름 또는 -run 불일치 | `_test.go`, TestXxx, -v 출력 확인 |
| 준비 실패 뒤 nil 접근 panic | 오류를 기록하고 계속 진행 | 필수 전제에는 Fatal 사용 |
| 실패 메시지에 입력이 없음 | 결과만 출력 | 입력·got·want를 함께 기록 |
| 경계 하나가 빠짐 | 정상 대표값만 검사 | 경계와 양옆 값 추가 |
| 래핑 뒤 오류 비교 실패 | err == 원인 또는 문구 전체 비교 | 약속한 원인을 errors.Is로 검사 |
| 보조 함수의 위치만 보임 | Helper 누락 | t.Helper 등록 |
| Windows 임시 폴더 정리 실패 | 파일이 열린 채 남음 | Cleanup에서 먼저 Close |
| 환경에 따라 결과가 달라짐 | 실제 환경·시간·외부 서비스 의존 | 경계를 주입하고 고정 자료 사용 |
| 개별 실행은 통과하지만 전체 실패 | 테스트 간 공유 상태 | 준비·정리·순서 의존성 점검 |
| 테스트가 수정 후에도 cached | 성공 결과 재사용 | -count=1로 재실행 |
| Example이 실제 실행되지 않음 | Output 주석 없음 | 예제 규칙 확인 |
| 커버리지는 높지만 버그를 놓침 | 실행만 하고 의미 있는 비교 없음 | 요구사항과 기대값 재검토 |

### 11.1 비교 방법도 계약에 맞추기

단순 int·string·구조체는 비교 가능 여부를 확인한 뒤 ==로 비교할 수 있습니다.
슬라이스는 ==로 원소 전체를 비교할 수 없으므로 slices.Equal 등 적절한 도구를 사용합니다.
맵의 출력 문자열을 순서가 있는 정답처럼 비교하지 않습니다.
부동소수점 계산은 문제에서 허용하는 오차와 범위를 정한 뒤 검사합니다.

오차 허용값이나 시간을 임의로 크게 잡아 테스트를 통과시키지 않습니다.
실패가 무엇을 뜻하는지부터 분석하고 요구사항에 맞게 고쳐야 합니다.

### 11.2 테스트를 과하게 늘리지 않기

함수 내부의 모든 줄을 그대로 따라 만드는 테스트는 구현 변경에 민감합니다.
반환값, 오류, 출력, 원본 보존처럼 호출자가 관찰하는 약속을 중심으로 검사합니다.
회귀 결함이 발견되면 그 결함을 재현하는 가장 작은 사례를 추가하세요.

성능은 [26장](../26-performance/README.md), 새로운 입력 탐색은 [27장](../27-fuzzing/README.md),
HTTP 요청·응답 계약은 [23장](../23-api-testing/README.md)에서 이어집니다.
이번 장에서 모든 테스트 기법을 한 번에 도입할 필요는 없습니다.

## 12. 연습문제와 참고 구현

[연습문제](../../exercises/12-testing/README.md)에는 기본 문제 6개와 선택 과제 1개가 있습니다.
실패 메시지 분석, 배송비 경계 수정, 범위 보정, 중복 제거,
설정 검증과 출력 오류 테스트를 차례로 다룹니다.
[해설](../../solutions/12-testing/README.md)은 네 가지 함수 구현과 테스트의 선택 이유를 제공합니다.

```sh
go test -count=1 ./examples/12-testing/... ./solutions/12-testing/...
```

문제에서 요구한 정상·경계·오류 사례를 먼저 작성합니다.
그다음 참고 구현과 비교하여 어떤 결함을 잡을 수 있는지 설명해 보세요.
테스트가 많다는 사실보다 계약이 분명하고 실패 원인을 읽을 수 있다는 점이 중요합니다.

## 13. 요약과 자기 점검

자동 테스트는 입력과 기대 결과를 코드로 남깁니다.
테이블과 하위 테스트는 사례를 정리하고 실패한 경계를 찾는 데 도움을 줍니다.
오류의 원인·상세 정보, 원본 보존과 자원 수명도 테스트할 수 있습니다.
각 테스트가 자료를 준비하고 정리하면 실행 순서와 개인 환경의 영향을 줄일 수 있습니다.

### 확인 질문

1. 테스트 파일이 일반 빌드 파일과 구분되는 이름 규칙은 무엇인가요?
2. Errorf와 Fatalf는 뒤의 검사 실행에 어떤 차이를 만드나요?
3. 1000g과 1001g을 모두 검사해야 하는 이유는 무엇인가요?
4. errors.As가 성공하기 전에 detail의 필드를 읽으면 왜 위험한가요?
5. 보조 함수에서 defer Close 대신 t.Cleanup을 사용한 이유는 무엇인가요?
6. 환경 변수를 바꾸는 테스트에 무조건 t.Parallel을 붙일 수 있나요?
7. 외부 테스트 패키지는 어떤 이름에 접근할 수 있나요?
8. Output 주석이 없는 Example은 어떻게 처리되나요?
9. `(cached)`와 `-count=1`은 각각 무엇을 뜻하나요?
10. 커버리지 100%만으로 모든 경계값이 정확하다고 말할 수 있나요?

<details>
<summary>답 확인</summary>

1. 파일 이름은 `_test.go`로 끝나고 일반 테스트 함수는 TestXxx 규칙을 따릅니다.
2. Errorf는 실패를 기록하고 계속하며 Fatalf는 해당 테스트 실행을 중단합니다.
3. 첫 구간의 끝과 다음 구간의 시작을 각각 검증하기 위해서입니다.
4. 다른 오류가 반환되면 원하는 구체 오류를 얻지 못했을 수 있기 때문입니다.
5. 보조 함수가 반환한 뒤에도 테스트가 파일을 사용해야 하기 때문입니다.
6. 환경 변수는 프로세스 전체 상태이며 Setenv는 병렬 테스트에서 허용되지 않습니다.
7. import한 패키지의 공개된 이름을 사용합니다.
8. 컴파일되지만 실행되지 않습니다.
9. 성공 결과의 재사용과 그 테스트 결과 캐시를 사용하지 않는 재실행을 뜻합니다.
10. 실행된 구문과 입력·기대값의 정확성은 다른 문제이므로 말할 수 없습니다.

</details>

### 완료 체크리스트

- [ ] 테스트 파일과 함수를 작성하고 실제로 실행했습니다.
- [ ] 정상·경계·오류 입력의 기대값을 먼저 정했습니다.
- [ ] 실패 메시지에 입력·실제값·기대값을 담았습니다.
- [ ] 필요한 전제가 실패하면 안전하게 검사를 멈췄습니다.
- [ ] 오류 원인과 구체 정보를 구분했습니다.
- [ ] 임시 파일·환경 변수를 테스트 뒤 정리했습니다.
- [ ] 공개 API의 예제 테스트를 확인했습니다.
- [ ] 의도적으로 넣은 경계 결함을 테스트가 찾는지 확인했습니다.

## 참고 자료

- [testing 패키지](https://pkg.go.dev/testing)
- [testing.T.Run](https://pkg.go.dev/testing#T.Run)
- [testing.T.Helper](https://pkg.go.dev/testing#T.Helper)
- [testing.T.Cleanup](https://pkg.go.dev/testing#T.Cleanup)
- [testing.T.TempDir](https://pkg.go.dev/testing#T.TempDir)
- [testing.T.Setenv](https://pkg.go.dev/testing#T.Setenv)
- [go test 명령](https://pkg.go.dev/cmd/go#hdr-Test_packages)
- [go test 플래그](https://pkg.go.dev/cmd/go#hdr-Testing_flags)
- [하위 테스트 소개](https://go.dev/blog/subtests)
- [커버리지 소개](https://go.dev/blog/cover)
- [errors](https://pkg.go.dev/errors)

확인 기준: 2026-11-09, Go 1.27.3.

## 다음 학습

[13장 제네릭](../13-generics/README.md)에서는 타입이 다른 값에 공통 알고리즘을 적용합니다.
여러 타입을 지원하더라도 입력·결과의 계약을 테스트하는 원칙은 유지합니다.
공통화할 필요가 있는 곳을 고르고 제약과 타입 매개변수를 사용합니다.

[이전: 11. 오류 처리와 자원 정리](../11-errors-defer/README.md) · [전체 목차](../README.md) · [다음: 13. 제네릭](../13-generics/README.md)
