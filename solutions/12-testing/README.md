# 12장 해설 — 자동 테스트

[연습문제](../../exercises/12-testing/README.md)의 구현과 검사 기준을 설명합니다.
테스트가 통과했다는 결과뿐 아니라, 어떤 입력에서 어떤 약속을 확인했는지 살펴보세요.
요구사항을 만족하는 다른 구현도 가능합니다.

## 해설 구성

| 문제 | 구현 | 주된 검증 |
| --- | --- | --- |
| 1 | 검사 순서 해설 | 실패 기록과 현재 테스트 중단 |
| 2 | 경계 수정 해설 | <=와 <의 차이, 회귀 사례 |
| 3 | [clamp](clamp/main.go) | 범위·끝점·한 점·오류 원인 |
| 4 | [dedupe](dedupe/main.go) | 첫 등장 순서·빈 값·원본 보존 |
| 5 | [config](config/main.go) | 없음과 빈 값·범위·환경 어댑터 |
| 6 | [report](report/main.go) | 출력 바이트·검증 실패·쓰기 실패 |

모든 명령은 저장소 루트에서 실행합니다.
개인 풀이를 검사하는 명령과 저장소의 참고 구현을 검사하는 명령은 경로가 다릅니다.

## 1. 실패 뒤에 계속해도 되는가

`t.Errorf`는 실패를 기록한 뒤 계속 실행합니다.
`t.Fatalf`는 실패를 기록하고 현재 테스트 실행을 중단합니다.
뒤 검사가 앞 결과의 유효성을 전제로 한다면 먼저 그 전제를 확인해야 합니다.

문제의 grade 결과는 문자열이어서 오류 후 비교 자체가 panic을 만들지는 않습니다.
하지만 오류가 없는 정상 결과의 계약을 비교하려는 목적이라면 Fatalf로 전제 실패를 분리하는 편이 명확합니다.
서로 독립된 두 필드의 차이는 Errorf로 각각 기록해도 됩니다.

### 오류 상세 정보 검사 수정

아래는 `errors.As` 결과를 확인한 뒤에만 상세 값을 읽는 발췌입니다.

```go
var detail *countError
if !errors.As(err, &detail) {
    t.Fatalf("unexpected error type: %T (%v)", err, err)
}
if detail.Input != "bad" {
    t.Errorf("input = %q; want %q", detail.Input, "bad")
}
```

As가 실패했는데 계속 진행하면 detail이 nil인 상태에서 Input에 접근할 수 있습니다.
원래 발견한 “다른 오류 종류”가 nil 포인터 panic으로 가려집니다.
검사 순서는 제품 코드의 오류 처리와 마찬가지로 안전한 다음 단계의 전제를 세워야 합니다.

### 메시지의 정보

실패 메시지에는 가능한 한 입력, 실제값, 기대값을 남깁니다.
문자열은 `%q`로 표시하면 빈 값·공백·개행을 구분하기 쉽습니다.
`fmt.Println`으로 기록만 하면 testing에 실패가 전달되지 않습니다.
의미 있는 차이를 발견했다면 Error·Fatal 등 테스트 실패 API를 사용해야 합니다.

## 2. 배송비 경계 수정

요구사항은 1~1000g을 첫 구간으로 정했습니다.
따라서 `grams < 1000`은 1000g만 다음 구간으로 밀어내는 결함입니다.
main은 1200g을 사용하므로 이 결함이 있어도 5000을 출력할 수 있습니다.

| 입력 | 잘못된 조건 결과 | 계약 |
| ---: | ---: | ---: |
| 999 | 3000 | 3000 |
| 1000 | 5000 | 3000 |
| 1001 | 5000 | 5000 |

first_edge 사례의 기대값을 5000으로 바꾸면 테스트는 통과할 수 있지만 요구사항을 깨뜨립니다.
기대값은 구현의 현재 결과가 아니라 약속한 요금 구간에서 가져와야 합니다.
조건을 `<=`로 수정한 뒤 같은 사례와 전체 패키지를 검사합니다.

```sh
go test -v -count=1 -run '^TestShippingFee$/^first_edge$' ./examples/12-testing/table
go test -count=1 ./examples/12-testing/table
```

위 명령은 이미 수정된 공식 예제를 검사합니다.
의도적인 실패 실습은 본문에서 복사한 개인 경로에 적용합니다.
실제 실패 기록에는 `shippingFee(1000) = 5000; want 3000`이라는 핵심 비교가 나타납니다.
파일 위치와 실행 시간은 환경마다 다릅니다.

## 3. 범위 보정 함수

잘못된 범위는 값의 보정에 앞서 검사합니다.
minimum==maximum이면 한 값만 허용하는 유효한 범위입니다.
대소 비교만 사용하므로 범위 차이나 value-minimum을 계산할 필요가 없습니다.

파일: [clamp/main.go](clamp/main.go).

```go
package main

import (
	"errors"
	"fmt"
)

var errBounds = errors.New("minimum must not exceed maximum")

func clamp(value, minimum, maximum int) (int, error) {
	if minimum > maximum {
		return 0, errBounds
	}
	if value < minimum {
		return minimum, nil
	}
	if value > maximum {
		return maximum, nil
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

실행:

```sh
go run ./solutions/12-testing/clamp
```

예상 출력:

```text
보정: 100
```

### 테스트

파일: [clamp/main_test.go](clamp/main_test.go).

```go
package main

import (
	"errors"
	"testing"
)

func TestClamp(t *testing.T) {
	tests := []struct {
		name                  string
		value, min, max, want int
	}{
		{"below", -1, 0, 100, 0}, {"minimum", 0, 0, 100, 0}, {"inside", 50, 0, 100, 50},
		{"maximum", 100, 0, 100, 100}, {"above", 101, 0, 100, 100}, {"one_point", 20, 7, 7, 7},
		{"negative_range", -5, -10, -2, -5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := clamp(tc.value, tc.min, tc.max)
			if err != nil || got != tc.want {
				t.Fatalf("got=%d err=%v; want %d", got, err, tc.want)
			}
		})
	}
}

func TestClampInvalid(t *testing.T) {
	if got, err := clamp(5, 10, 0); got != 0 || !errors.Is(err, errBounds) {
		t.Fatalf("got=%d err=%v", got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 보정: 100
}
```

입력 표는 아래·최솟값·내부·최댓값·위를 구분합니다.
한 점과 음수 범위를 별도 사례로 두어 “0~100만 처리하는 함수”로 우연히 구현하지 않았는지 확인합니다.
잘못된 범위에서는 실패 결과와 errBounds 원인을 함께 검사합니다.

### 이 테스트가 지키는 것

테스트는 함수가 어떤 switch나 if 순서를 쓰는지 확인하지 않습니다.
호출한 쪽이 받는 값과 오류라는 외부 계약을 검사합니다.
구현을 더 읽기 쉽게 바꾸더라도 계약이 유지되면 같은 테스트를 사용할 수 있습니다.

## 4. 첫 등장 순서로 중복 제거

맵은 이미 본 값인지 판정하는 데 사용합니다.
출력 순서는 입력을 읽으며 새 슬라이스에 추가한 순서로 결정합니다.
맵을 마지막에 순회하여 결과를 만들면 이 문제의 순서 계약을 지킬 수 없습니다.

파일: [dedupe/main.go](dedupe/main.go).

```go
package main

import "fmt"

func unique(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func main() { fmt.Println(unique([]string{"Go", "한글", "Go", "go"})) }
```

실행:

```sh
go run ./solutions/12-testing/dedupe
```

예상 출력:

```text
[Go 한글 go]
```

### 테스트

파일: [dedupe/main_test.go](dedupe/main_test.go).

```go
package main

import (
	"slices"
	"testing"
)

func TestUnique(t *testing.T) {
	for _, tc := range []struct {
		name        string
		input, want []string
	}{
		{"nil", nil, []string{}}, {"empty", []string{}, []string{}},
		{"order", []string{"b", "a", "b", "c"}, []string{"b", "a", "c"}},
		{"case", []string{"Go", "go", "Go"}, []string{"Go", "go"}},
		{"empty_string", []string{"", "x", ""}, []string{"", "x"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := slices.Clone(tc.input)
			got := unique(tc.input)
			if got == nil || !slices.Equal(got, tc.want) {
				t.Fatalf("got=%v; want %v", got, tc.want)
			}
			if len(got) > 0 {
				got[0] = "changed"
			}
			if !slices.Equal(tc.input, before) {
				t.Error("input changed")
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// [Go 한글 go]
}
```

slices.Equal은 원소와 순서를 비교합니다.
nil 슬라이스와 길이 0인 슬라이스는 원소 비교에서 같을 수 있으므로,
이 문제의 비-nil 계약은 `got == nil`로 따로 확인합니다.
[slices.Equal](https://pkg.go.dev/slices#Equal)의 비교 의미를 참고하세요.

### 원본 보존을 두 단계로 확인

함수를 호출한 직후 입력이 바뀌었는지 확인할 수 있습니다.
여기서는 한 단계 더 나아가 결과의 첫 원소를 바꾼 뒤 입력이 유지되는지 확인합니다.
이를 통해 반환 슬라이스가 원본 저장 공간을 공유하지 않는다는 계약도 검사합니다.

값이 string인 이번 문제에서는 원소 값 복사로 필요한 독립성이 생깁니다.
중첩 슬라이스나 포인터가 들어간 자료형이라면 복사 범위를 다시 정의해야 합니다.
05장의 얕은 복사·공유 저장 공간 개념을 테스트로 확인한 예입니다.

### 빈 문자열을 버리지 않은 이유

문제에서 빈 문자열을 유효한 값으로 정했습니다.
실무의 특정 목록에서는 빈 값을 거부하거나 제거할 수도 있지만,
그 정책은 중복 제거와 다른 요구사항입니다.
이 구현은 대소문자 변환이나 공백 정리도 하지 않습니다.

## 5. 설정의 존재 여부와 값 검증

lookup의 두 번째 반환값은 키의 존재 여부입니다.
없을 때 기본값을 쓰는 정책과, 존재하지만 빈 값이면 오류라는 정책을 분리합니다.
`os.Getenv`처럼 문자열만 받으면 두 상태를 구분하기 어려우므로 lookup의 시그니처에 bool을 포함했습니다.

파일: [config/main.go](config/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

const limitKey = "GOLANG_BOOK12_LIMIT"

var errLimit = errors.New("limit must be 1..100 decimal digits")

func loadLimit(lookup func(string) (string, bool)) (int, error) {
	text, present := lookup(limitKey)
	if !present {
		return 20, nil
	}
	if text == "" {
		return 0, errLimit
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return 0, errLimit
		}
	}
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 || n > 100 {
		return 0, errLimit
	}
	return n, nil
}

func main() {
	n, err := loadLimit(func(key string) (string, bool) { return "30", true })
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("한도:", n)
}
```

실행:

```sh
go run ./solutions/12-testing/config
```

예상 출력:

```text
한도: 30
```

### 문자열 형식과 수치 범위

Atoi에 전달하기 전에 ASCII 숫자로만 이루어졌는지 검사합니다.
따라서 `+1`, 공백, 소수점과 전각 숫자를 허용하는 정책이 섞이지 않습니다.
숫자로 변환한 뒤에는 1~100을 검사합니다.
너무 긴 숫자에서 Atoi가 오류를 반환해도 errLimit으로 정한 입력 오류 계약을 유지합니다.

### 테스트

파일: [config/main_test.go](config/main_test.go).

```go
package main

import (
	"errors"
	"os"
	"testing"
)

func TestLoadLimit(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		present    bool
		want       int
		bad        bool
	}{
		{"unset", "", false, 20, false}, {"empty", "", true, 0, true}, {"minimum", "1", true, 1, false},
		{"leading_zero", "0030", true, 30, false}, {"maximum", "100", true, 100, false},
		{"zero", "0", true, 0, true}, {"over", "101", true, 0, true}, {"sign", "+1", true, 0, true},
		{"space", " 1", true, 0, true}, {"overflow", "999999999999999999999", true, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loadLimit(func(key string) (string, bool) {
				if key != limitKey {
					t.Fatalf("key=%q", key)
				}
				return tc.text, tc.present
			})
			if (err != nil) != tc.bad || (tc.bad && !errors.Is(err, errLimit)) || got != tc.want {
				t.Fatalf("got=%d err=%v; want %d bad=%v", got, err, tc.want, tc.bad)
			}
		})
	}
}

func TestLoadLimitEnvironment(t *testing.T) {
	t.Setenv(limitKey, "42")
	got, err := loadLimit(os.LookupEnv)
	if err != nil || got != 42 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 한도: 30
}
```

대부분의 사례는 환경 변수를 바꾸지 않고 값과 존재 여부를 주는 작은 함수를 사용합니다.
이 방식은 운영체제 설정·실행 순서와 관계없이 같은 입력을 재현합니다.
별도 테스트 하나에서 실제 os.LookupEnv와의 연결을 확인합니다.

`t.Setenv`는 테스트 뒤 기존 환경을 복원합니다.
프로세스 전역 상태를 바꾸므로 이 테스트나 부모를 병렬로 만들지 않습니다.
일반 사용자의 설정을 바꾼 상태로 끝내는 os.Setenv 호출을 테스트 곳곳에 흩어 놓지 마세요.

### 오류 조건을 빠뜨리지 않는 비교

정상 사례에서 errors.Is(err, errLimit)이 거짓이라는 사실만 확인하면,
다른 종류의 예상하지 못한 오류를 놓칠 수 있습니다.
따라서 먼저 실제 오류 유무가 기대와 같은지 검사하고,
오류 사례라면 약속한 errLimit인지 추가로 검사합니다.
정상값과 실패값을 한 테이블에 넣을 때 비교 조건도 신중히 설계해야 합니다.

## 6. 출력 계약과 쓰기 실패

이름을 정리하고 검사하는 단계와 출력하는 단계를 분리했습니다.
잘못된 이름이 뒤쪽에 있더라도 첫 출력 전에 발견합니다.
정리 결과를 새 슬라이스에 담아 원본의 공백과 문자열 값도 보존합니다.

파일: [report/main.go](report/main.go).

```go
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

var errName = errors.New("name must be 1..20 UTF-8 runes without control characters")

func writeRoster(output io.Writer, names []string) error {
	cleaned := make([]string, len(names))
	for i, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 20 {
			return fmt.Errorf("name %d: %w", i+1, errName)
		}
		for _, ch := range name {
			if unicode.IsControl(ch) {
				return fmt.Errorf("name %d: %w", i+1, errName)
			}
		}
		cleaned[i] = name
	}
	if _, err := fmt.Fprintf(output, "학습자: %d\n", len(cleaned)); err != nil {
		return fmt.Errorf("write heading: %w", err)
	}
	for _, name := range cleaned {
		if _, err := fmt.Fprintf(output, "- %s\n", name); err != nil {
			return fmt.Errorf("write name: %w", err)
		}
	}
	return nil
}

func main() {
	var output bytes.Buffer
	if err := writeRoster(&output, []string{" 민수 ", "지수"}); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(output.String())
}
```

실행:

```sh
go run ./solutions/12-testing/report
```

예상 출력:

```text
학습자: 2
- 민수
- 지수
```

이 예제는 main에서 bytes.Buffer에 출력한 뒤 화면에 보여 줍니다.
writeRoster는 Writer를 통해 출력하므로 테스트에서 stdout을 전역으로 바꿀 필요가 없습니다.
같은 함수에 파일이나 다른 Writer를 전달할 수 있지만 그 수명은 호출자가 관리합니다.

### 테스트

파일: [report/main_test.go](report/main_test.go).

```go
package main

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestWriteRoster(t *testing.T) {
	names := []string{" 민수 ", "지수"}
	before := slices.Clone(names)
	var output bytes.Buffer
	if err := writeRoster(&output, names); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "학습자: 2\n- 민수\n- 지수\n"; got != want {
		t.Errorf("got=%q; want %q", got, want)
	}
	if !slices.Equal(names, before) {
		t.Error("input changed")
	}
}

func TestWriteRosterEmpty(t *testing.T) {
	var output bytes.Buffer
	if err := writeRoster(&output, nil); err != nil || output.String() != "학습자: 0\n" {
		t.Fatalf("output=%q err=%v", output.String(), err)
	}
}

func TestWriteRosterInvalid(t *testing.T) {
	for _, name := range []string{" ", "a\nb", "a\u0085b", strings.Repeat("가", 21), string([]byte{0xff})} {
		var output bytes.Buffer
		err := writeRoster(&output, []string{"정상", name})
		if !errors.Is(err, errName) || output.Len() != 0 {
			t.Fatalf("output=%q err=%v", output.String(), err)
		}
	}
}

func TestWriteRosterMaximumName(t *testing.T) {
	var output bytes.Buffer
	name := strings.Repeat("가", 20)
	if err := writeRoster(&output, []string{name}); err != nil || !strings.Contains(output.String(), name) {
		t.Fatalf("output=%q err=%v", output.String(), err)
	}
}

type failedWriter struct {
	err       error
	remaining int
}

func (w *failedWriter) Write(p []byte) (int, error) {
	if w.remaining == 0 {
		return 0, w.err
	}
	w.remaining--
	return len(p), nil
}

func TestWriteRosterWriteError(t *testing.T) {
	cause := errors.New("output closed")
	for _, remaining := range []int{0, 1} {
		if err := writeRoster(&failedWriter{err: cause, remaining: remaining}, []string{"지수"}); !errors.Is(err, cause) {
			t.Fatalf("remaining=%d error=%v", remaining, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 학습자: 2
	// - 민수
	// - 지수
}
```

### 바이트를 정확히 비교할 때

출력의 헤더·줄바꿈·공백을 약속했으므로 Buffer.String과 정확한 기대 문자열을 비교합니다.
사용 예를 보여 주는 Example도 있지만, 앞뒤 공백을 무시하는 예제 비교만으로 모든 출력 계약을 대신하지 않았습니다.

이름을 한 줄씩 출력하므로 중간 제어 문자나 줄바꿈을 허용하지 않습니다.
앞뒤 공백은 먼저 정리하므로 검사 대상은 정리 후 남은 문자입니다.
한글을 포함한 길이 제한은 바이트 수가 아니라 룬 수입니다.

### 검증 실패와 쓰기 실패

| 실패 시점 | 출력 상태 | 원인 |
| --- | --- | --- |
| 이름 검증 | 비어 있음 | 출력을 시작하지 않음 |
| 헤더 쓰기 | Writer 동작에 따라 일부 가능 | 첫 외부 쓰기 실패 |
| 이름 쓰기 | 헤더 등이 이미 나갔을 수 있음 | 다음 외부 쓰기 실패 |

검증 단계에서 출력이 없다는 계약이 쓰기 단계의 원자성까지 보장하지는 않습니다.
일반 Writer에 이미 전달된 바이트는 이 함수가 되돌릴 수 없습니다.
출력 오류를 반환하면 호출자가 실패를 인식하고 별도 저장 정책을 적용할 수 있습니다.

### 대역이 재현하는 실패

failedWriter의 remaining은 성공적으로 허용할 Write 호출 수입니다.
0이면 헤더 쓰기, 1이면 다음 이름 쓰기에서 오류를 냅니다.
두 경로 모두 원인이 보존되는지 errors.Is로 확인합니다.
정상 Writer의 내부 구현을 복제하는 것이 아니라 필요한 실패 경계만 제공합니다.

출력 함수에 호출이 하나 더 추가되면 대역의 의미도 검토해야 합니다.
이 작은 대역은 현재의 “헤더 다음 이름” 계약을 설명하는 학습용입니다.
실무에서는 관찰해야 할 출력·오류의 성격에 맞추어 대역을 설계하세요.

## 7. 공개 API와 예제 테스트 검토

외부 테스트 패키지는 import한 패키지의 공개 이름을 사용합니다.
공개된 Greeting의 일반 이름·빈 이름 정책을 문서 예제로 보여 줍니다.

```sh
go test -v -count=1 -run '^ExampleGreeting' ./examples/12-testing/publicapi
```

ExampleGreeting과 ExampleGreeting_empty가 실행되는지 확인합니다.
예제 이름 뒤의 소문자 접미사는 같은 API의 다른 사례를 구분합니다.
Output 주석을 제거하면 컴파일만 확인하고 예제 본문은 실행하지 않습니다.
[예제 함수 규칙](https://pkg.go.dev/testing#hdr-Examples)을 참고하세요.

내부 공백·개행까지 엄격한 출력 계약에는 Buffer를 이용한 일반 테스트를 함께 사용합니다.
행 순서가 계약이라면 Unordered output으로 순서를 무시하지 않습니다.
테스트가 편하다는 이유로 패키지의 내부 구현을 모두 공개하지 않습니다.

## 8. 전체 실행과 검사

네 참고 구현을 실행합니다.

```sh
go run ./solutions/12-testing/clamp
go run ./solutions/12-testing/dedupe
go run ./solutions/12-testing/config
go run ./solutions/12-testing/report
```

일반 검사:

```sh
go test -count=1 ./examples/12-testing/... ./solutions/12-testing/...
go vet ./examples/12-testing/... ./solutions/12-testing/...
go build ./examples/12-testing/... ./solutions/12-testing/...
```

실행 순서 의존성을 확인할 수 있습니다.

```sh
go test -shuffle=on -count=1 ./examples/12-testing/... ./solutions/12-testing/...
```

지원하는 환경에서는 race detector도 적용합니다.
순서 섞기, race detector, 일반 테스트는 확인하는 범위가 서로 다릅니다.
어느 하나의 성공을 모든 결함이 없다는 증명으로 해석하지 않습니다.

```sh
go test -race -count=1 ./examples/12-testing/... ./solutions/12-testing/...
```

## 9. 풀이를 비교하는 기준

- 입력이 계약의 대표값과 경계, 실패를 구분하는가?
- 결과가 틀리면 테스트가 실제로 실패하는가?
- 기대값을 구현과 같은 방식으로 잘못 계산하지 않았는가?
- 오류 유무와 원인, 실패 반환값을 필요한 범위에서 검사하는가?
- 결과 슬라이스를 바꿨을 때 원본이 유지되는가?
- 임시 파일과 환경 변수가 다음 테스트에 남지 않는가?
- 출력 오류가 발생하는 앞·뒤 경로를 놓치지 않았는가?
- 실패 메시지를 읽고 어떤 계약이 깨졌는지 알 수 있는가?

모든 함수에 같은 수의 테스트를 맞출 필요는 없습니다.
함수의 계약과 실패 위험에 맞는 사례를 선택하고,
새 결함이 발견되면 재현 가능한 회귀 사례로 남깁니다.

확인 기준: 2026-11-09, Go 1.27.3.

[본문으로](../../book/12-testing/README.md) · [연습문제로](../../exercises/12-testing/README.md) · [해설 목록](../README.md)
