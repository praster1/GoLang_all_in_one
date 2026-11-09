# 29장 해설 — 종합 실습: CSV 분석 CLI

[연습문제](../../exercises/29-csv-cli/README.md)의 계약을 구현하고 검사합니다.
핵심은 새 기능을 추가해도 입력 검증, 결과의 의미, 원본 보존과 오류 전달을 유지하는 것입니다.
완성 코드를 읽기 전에 각 함수가 무엇을 받고 무엇을 보장하는지 먼저 확인하세요.

## 해설 구성

| 문제 | 참고 구현 | 중심 판단 |
| --- | --- | --- |
| 1 | 결과 분석 | 레코드·그룹·합계의 단위 |
| 2 | 오류 분석 | 문법·값·한도·명령 사용법의 경계 |
| 3 | [filter](filter/main.go) | 선택한 그룹에서 요약 재계산 |
| 4 | [ranking](ranking/main.go) | 동률 기준과 원본 보존 |
| 5 | [merge](merge/main.go) | 입력별 검증과 전체 실패 정책 |
| 6 | [render](render/main.go) | 분석·직렬화 성공 후 바이트 반환 |

명령은 모두 저장소 루트에서 실행합니다.
네 실행 예는 가상의 고정 입력으로 기능을 보여 줍니다.
실제 파일·표준 입력·종료 코드는 본문의 `cmd/csvstats`가 담당합니다.

## 1. 집계 결과 읽기

문제 1의 데이터 레코드는 네 개, 그룹은 세 개입니다.
금액 0인 문구 행도 유효한 거래 한 건이므로 Count에 포함됩니다.
이름순 출력은 다음과 같습니다.

```text
category,count,total
도서,2,20000
문구,1,0
식품,1,5000
```

JSON의 Rows는 4, Total은 25000입니다.
기본 CSV는 그룹별 행만 출력하므로 별도의 전체 합계 행은 없습니다.
같은 분류는 입력에서 인접하지 않아도 맵의 같은 키에 누적됩니다.

### 빈 입력을 구분하는 이유

| 입력 | 결과 | 이유 |
| --- | --- | --- |
| 헤더만 있음 | 성공, Rows=0·Total=0·Groups=[] | 데이터 0개를 허용한 계약 |
| 완전히 비어 있음 | 오류 | 필수 헤더 없음 |
| 헤더 순서가 바뀜 | 오류 | 열 이름과 순서를 정확히 약속함 |
| 인용된 쉼표 포함 분류 | 성공 | CSV Reader가 인용 규칙을 처리함 |

필수 형식이 없는 입력과 형식은 갖췄지만 자료가 없는 입력을 구분했습니다.
이는 모든 CSV 도구의 공통 규칙이 아니라 이 프로젝트가 정한 입력 계약입니다.
다른 정책을 선택할 수 있지만 문서·코드·테스트를 함께 바꿔야 합니다.

## 2. 오류와 한도 분석

필드 개수, 분류, 금액 검사는 DataError로 위치를 전달합니다.
그 안의 원인을 `%w`로 감싸므로 `errors.Is`로 ErrInput을 찾을 수 있습니다.
입력 크기·레코드 수 한도 초과는 ErrLimit으로 구분합니다.

| 사례 | CLI 종료 코드 | stdout | 주된 진단 |
| --- | ---: | --- | --- |
| 잘못된·음수 금액, 빈 분류, 열 부족 | 1 | 비어 있음 | 레코드·필드 입력 오류 |
| 유효한 한도보다 큰 입력 | 1 | 비어 있음 | 입력 한도 초과 |
| `-max-rows 0` | 2 | 비어 있음 | 잘못된 옵션 값 |
| `-format xml` | 2 | 비어 있음 | 지원하지 않는 형식 |

잘못된 한도를 지정하는 것과 유효한 한도로 실제 입력을 처리하다 초과하는 것은 다릅니다.
전자는 실행 설정의 오류이고 후자는 입력 처리 실패입니다.
지원하지 않는 옵션을 파일을 연 뒤에야 검사할 이유도 없습니다.

### 레코드 위치와 오류의 원인

DataError의 Record는 헤더를 1로 보는 논리 CSV 레코드 번호입니다.
인용 필드 안의 줄바꿈과 빈 줄 처리가 있으므로 물리적 줄 번호와 같다고 단정하지 않습니다.
CSV 따옴표 자체가 잘못되면 표준 `csv.ParseError`를 감싼 오류일 수 있습니다.
파일 장치의 읽기 오류 역시 DataError로 바뀌지 않습니다.
따라서 모든 실패에 대해 같은 구체 오류 자료형을 강제로 가정하지 않습니다.

`errors.As`는 실제로 위치 정보를 제공하는 오류인지 확인하는 수단입니다.
`errors.Is`는 래핑된 원인이 무엇인지 확인하는 수단입니다.
이 두 질문은 함께 쓰일 수 있지만 서로를 대체하지 않습니다.
[오류 래핑 문서](https://pkg.go.dev/errors)를 참고하세요.

### 경계값 확인

입력 바이트가 정확히 한도와 같으면 허용합니다.
추가 바이트를 발견했을 때에만 초과로 판정하도록 한도보다 한 바이트 더 읽습니다.
CSV Reader가 미리 읽을 수 있으므로 제한은 레코드 문자열 길이의 합이 아니라 원본 입력 스트림에 적용합니다.

분석 도중 오류가 나면 함수는 빈 Summary와 오류를 반환합니다.
정상 행이 먼저 여러 개 나왔더라도 부분 합계를 전체 보고서로 사용하지 않습니다.
읽기 오류를 나중에 발생시키는 Reader 테스트가 이 정책을 확인합니다.
쓰기를 시작한 뒤의 오류는 이미 외부에 전달한 바이트를 취소하지 못한다는 점도 구분해야 합니다.

## 3. 분류 필터와 일관된 요약

원본 Summary의 Groups만 줄이고 Rows·Total을 그대로 두면 보고서가 모순됩니다.
따라서 새 결과를 만들면서 선택한 그룹의 Count·Total을 누적합니다.
이 함수의 입력은 분석기가 성공한 Summary이므로 그룹 이름이 유일하다는 전제를 사용합니다.

### 전체 코드

파일: [filter/main.go](filter/main.go).

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func selectCategory(summary report.Summary, category string) report.Summary {
	result := report.Summary{Groups: make([]report.Group, 0)}
	for _, group := range summary.Groups {
		if group.Category == category {
			result.Groups = append(result.Groups, group)
			result.Rows += group.Count
			result.Total += group.Total
		}
	}
	return result
}

func main() {
	summary, err := report.Analyze(strings.NewReader("category,amount\n도서,12\n식품,5\n도서,8\n"), report.DefaultOptions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := report.WriteCSV(os.Stdout, selectCategory(summary, "도서")); err != nil {
		fmt.Println("error:", err)
	}
}
```

### 실행

```sh
go run ./solutions/29-csv-cli/filter
```

예상 출력:

```text
category,count,total
도서,2,20
```

### 왜 새 Groups를 만드는가

Summary를 값으로 받아도 내부 슬라이스는 같은 저장 공간을 가리킬 수 있습니다.
새 슬라이스에 Group 값을 추가하면 결과 그룹의 필드 변경이 원본에 전달되지 않습니다.
현재 Group 필드는 string·int·int64이므로 별도의 중첩 가변 자료는 없습니다.
나중에 포인터·슬라이스·맵 필드를 추가한다면 복사 정책을 다시 검토해야 합니다.

정확한 이름 일치를 사용하며 필터 함수가 공백을 정리하지 않습니다.
입력 자료의 공백 정리는 ParseRow의 책임입니다.
사용자가 입력하는 필터 옵션에 공백을 허용하고 싶다면 그 옵션 경계에서 정책을 정합니다.

### 테스트

파일: [filter/main_test.go](filter/main_test.go).

```go
package main

import (
	"slices"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func TestSelectCategory(t *testing.T) {
	source := report.Summary{Rows: 3, Total: 25, Groups: []report.Group{{Category: "도서", Count: 2, Total: 20}, {Category: "식품", Count: 1, Total: 5}}}
	before := slices.Clone(source.Groups)
	for _, tc := range []struct {
		name  string
		rows  int
		total int64
	}{{"도서", 2, 20}, {"식품", 1, 5}, {"없음", 0, 0}} {
		got := selectCategory(source, tc.name)
		if got.Rows != tc.rows || got.Total != tc.total || got.Groups == nil {
			t.Fatalf("got=%+v", got)
		}
		if len(got.Groups) > 0 {
			got.Groups[0].Total = 999
		}
		if !slices.Equal(source.Groups, before) {
			t.Fatal("mutated input")
		}
	}
}

func Example_main() {
	main()
	// Output:
	// category,count,total
	// 도서,2,20
}
```

결과값만 맞는지 확인하면 원본 훼손을 놓칠 수 있습니다.
테스트는 결과의 Total을 바꾼 뒤 원본이 그대로인지 확인합니다.
일치하는 그룹이 없는 경우에도 비-nil 빈 Groups를 유지하여 JSON 배열 계약을 지킵니다.

### 다른 설계를 선택한다면

입력 레코드를 읽는 즉시 다른 분류를 버리면 저장하는 그룹 수는 줄일 수 있습니다.
하지만 버린 레코드의 오류를 검사할지, 전체 입력 한도에 포함할지부터 결정해야 합니다.
이 문제는 전체 입력 검증을 유지하고 분석이 끝난 결과를 선택하는 방법을 사용했습니다.

## 4. 합계순 상위 그룹

먼저 Groups를 복사하고, Total 내림차순 뒤 Category 오름차순을 적용합니다.
정렬 기준에 두 번째 키가 있어야 같은 합계에서도 결과가 일정합니다.
맵 순회 순서나 입력의 우연한 순서를 순위 결정 기준으로 사용하지 않습니다.

### 전체 코드

파일: [ranking/main.go](ranking/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func topGroups(summary report.Summary, n int) ([]report.Group, error) {
	if n < 0 {
		return nil, errors.New("n must be non-negative")
	}
	groups := slices.Clone(summary.Groups)
	slices.SortFunc(groups, func(a, b report.Group) int {
		if a.Total > b.Total {
			return -1
		}
		if a.Total < b.Total {
			return 1
		}
		return strings.Compare(a.Category, b.Category)
	})
	if n > len(groups) {
		n = len(groups)
	}
	return groups[:n], nil
}

func main() {
	summary, err := report.Analyze(strings.NewReader("category,amount\n도서,12\n식품,20\n도서,8\n"), report.DefaultOptions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	groups, err := topGroups(summary, 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for i, group := range groups {
		fmt.Printf("%d. %s: %d\n", i+1, group.Category, group.Total)
	}
}
```

### 실행

```sh
go run ./solutions/29-csv-cli/ranking
```

예상 출력:

```text
1. 도서: 20
2. 식품: 20
```

합계가 같은 두 분류는 문자열 이름순으로 정렬됩니다.
이는 로캘별 사전 정렬이나 사용자가 기대하는 모든 언어의 자연 정렬을 구현한 것은 아닙니다.
문자열 비교 기준은 [strings.Compare](https://pkg.go.dev/strings#Compare)에 따릅니다.

### 비교 함수에서 뺄셈을 피한 이유

정렬 비교 함수에는 음수·0·양수만 필요합니다.
두 큰 정수의 차를 계산하거나 차이를 int로 변환할 필요가 없습니다.
명시적인 대소 비교는 오버플로와 변환 범위를 고려해야 하는 부분을 줄입니다.

n의 음수 여부를 먼저 검사하고, 너무 큰 n은 실제 길이로 조정합니다.
n=0이나 빈 입력은 길이 0의 정상 결과입니다.
이 함수는 빈 슬라이스가 nil인지까지 약속하지 않습니다.
이 결과를 JSON으로 바로 내보낼 때 배열 표기가 필요하다면 출력 경계에서 정책을 정합니다.

### 테스트

파일: [ranking/main_test.go](ranking/main_test.go).

```go
package main

import (
	"slices"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func TestTopGroups(t *testing.T) {
	source := report.Summary{Groups: []report.Group{{Category: "z", Total: 10}, {Category: "b", Total: 20}, {Category: "a", Total: 20}}}
	before := slices.Clone(source.Groups)
	got, err := topGroups(source, 2)
	if err != nil || len(got) != 2 || got[0].Category != "a" || got[1].Category != "b" {
		t.Fatalf("got=%v err=%v", got, err)
	}
	got[0].Total = 999
	if !slices.Equal(source.Groups, before) {
		t.Fatal("mutated source")
	}
	if got, err := topGroups(source, 0); err != nil || len(got) != 0 {
		t.Fatal("zero")
	}
	if got, err := topGroups(source, 9); err != nil || len(got) != 3 {
		t.Fatal("clamp")
	}
	if _, err := topGroups(source, -1); err == nil {
		t.Fatal("negative")
	}
	if got, err := topGroups(report.Summary{}, 2); err != nil || len(got) != 0 {
		t.Fatal("empty")
	}
}

func Example_main() {
	main()
	// Output:
	// 1. 도서: 20
	// 2. 식품: 20
}
```

테스트는 합계 순서, 동률 이름순, 개수의 경계, 원본 보존을 각각 확인합니다.
정렬 알고리즘의 내부 동작을 따라 쓰지 않고 기대한 순서의 이름을 비교합니다.

### 기본 CSV Writer와 순위 출력의 차이

`report.WriteCSV`는 교재의 기본 계약인 이름순 정렬을 적용합니다.
상위 N개를 선택한 뒤 이 Writer를 호출하면 선택된 집합 안의 출력 순서는 다시 이름순이 됩니다.
순위를 보여 주려면 이 예제처럼 순위 순서를 사용하는 별도 출력 함수를 만들거나,
Writer가 정렬을 맡는 계약 자체를 명시적으로 바꿔야 합니다.
기존 Writer가 어떤 책임을 가지는지 확인한 뒤 기능을 조합하세요.

## 5. 여러 입력 병합

각 입력은 헤더를 포함한 독립적인 CSV입니다.
각각 Analyze를 호출하고 성공한 Summary를 새 맵에 누적합니다.
파일 바이트를 연결하면 중간 헤더를 일반 데이터로 읽게 되므로 다른 처리가 필요합니다.

### 전체 코드

파일: [merge/main.go](merge/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func mergeInputs(inputs []io.Reader) (report.Summary, error) {
	if len(inputs) > 10 {
		return report.Summary{}, errors.New("at most 10 inputs")
	}
	result := report.Summary{Groups: make([]report.Group, 0)}
	groups := make(map[string]report.Group)
	for i, input := range inputs {
		if input == nil {
			return report.Summary{}, fmt.Errorf("input %d: nil reader", i+1)
		}
		part, err := report.Analyze(input, report.DefaultOptions())
		if err != nil {
			return report.Summary{}, fmt.Errorf("input %d: %w", i+1, err)
		}
		result.Rows += part.Rows
		result.Total += part.Total
		for _, group := range part.Groups {
			current := groups[group.Category]
			current.Category = group.Category
			current.Count += group.Count
			current.Total += group.Total
			groups[group.Category] = current
		}
	}
	for _, group := range groups {
		result.Groups = append(result.Groups, group)
	}
	slices.SortFunc(result.Groups, func(a, b report.Group) int { return strings.Compare(a.Category, b.Category) })
	return result, nil
}

func main() {
	inputs := []io.Reader{strings.NewReader("category,amount\n도서,12\n"), strings.NewReader("category,amount\n식품,7\n도서,8\n")}
	summary, err := mergeInputs(inputs)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := report.WriteCSV(os.Stdout, summary); err != nil {
		fmt.Println("error:", err)
	}
}
```

### 실행

```sh
go run ./solutions/29-csv-cli/merge
```

예상 출력:

```text
category,count,total
도서,2,20
식품,1,7
```

### 합계와 자원 한도

입력은 최대 10개이고, 각 입력은 기본 한도인 10000개 데이터 레코드를 허용합니다.
따라서 전체 Rows는 최대 100000입니다.
금액 하나의 최대 1000000000을 곱한 전체 Total 상한은 100000000000000입니다.
이 범위는 int64로 표현할 수 있으며 Rows도 교재에서 지원하는 플랫폼의 int 범위 안입니다.
`TestMergeMaximumRowsAndAmounts`는 실제로 10개 입력에 각각 10000개의 최대 금액을 넣어 이 상한의 결과를 확인합니다.

입력 바이트 한도도 각 입력별 1 MiB입니다.
이 함수는 입력 전체를 동시에 읽어 메모리에 보관하지 않지만 누적 그룹 맵은 유지합니다.
따라서 공간 사용량을 “항상 상수”라고 설명할 수는 없습니다.
입력 수와 각 입력 크기·행 수 제한이 처리량의 상한을 정합니다.

### 전체 실패와 Reader 소유권

두 번째 입력에서 실패했다면 첫 입력의 집계도 성공 결과로 내보내지 않습니다.
오류 메시지는 입력 번호를 추가하고 `%w`로 기존 원인을 보존합니다.
호출자는 메시지 문자열을 비교하지 않고 ErrInput 등 원래 원인을 검사할 수 있습니다.

이 함수는 io.Reader만 빌려 사용하며 닫지 않습니다.
파일을 연 호출자가 파일을 닫는 책임을 가집니다.
인터페이스가 nil인지 검사하는 것은 nil 포인터를 담은 모든 Reader 구현의 안전을 보장하지 않습니다.
호출자는 Read 계약을 지키는 유효한 구현을 전달해야 합니다.
인터페이스와 nil의 차이는 [09장](../../book/09-methods-interfaces/README.md)을 참고하세요.

### 테스트

파일: [merge/main_test.go](merge/main_test.go).

```go
package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func TestMergeInputs(t *testing.T) {
	got, err := mergeInputs([]io.Reader{strings.NewReader("category,amount\nx,1\n"), strings.NewReader("category,amount\nx,2\ny,3\n")})
	if err != nil || got.Rows != 3 || got.Total != 6 || len(got.Groups) != 2 || got.Groups[0].Count != 2 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	got, err = mergeInputs(nil)
	if err != nil || got.Rows != 0 || got.Groups == nil {
		t.Fatal("empty inputs")
	}
	got, err = mergeInputs([]io.Reader{strings.NewReader("category,amount\nx,1\n"), strings.NewReader("category,amount\ny,bad\n")})
	if !errors.Is(err, report.ErrInput) || got.Rows != 0 || got.Total != 0 {
		t.Fatal("partial result")
	}
	inputs := make([]io.Reader, 10)
	for i := range inputs {
		inputs[i] = strings.NewReader("category,amount\n")
	}
	inputs[9] = strings.NewReader("category,amount\nx,7\n")
	got, err = mergeInputs(inputs)
	if err != nil || got.Rows != 1 || got.Total != 7 || len(got.Groups) != 1 {
		t.Fatalf("ten inputs: got=%+v err=%v", got, err)
	}
	if _, err := mergeInputs(make([]io.Reader, 11)); err == nil {
		t.Fatal("too many inputs")
	}
	if _, err := mergeInputs([]io.Reader{nil}); err == nil {
		t.Fatal("nil input")
	}
}

func TestMergeMaximumRowsAndAmounts(t *testing.T) {
	text := "category,amount\n" + strings.Repeat("x,1000000000\n", report.DefaultOptions().MaxRows)
	inputs := make([]io.Reader, 10)
	for i := range inputs {
		inputs[i] = strings.NewReader(text)
	}
	got, err := mergeInputs(inputs)
	if err != nil {
		t.Fatal(err)
	}
	const wantRows = 100000
	const wantTotal int64 = 100000000000000
	if got.Rows != wantRows || got.Total != wantTotal || len(got.Groups) != 1 ||
		got.Groups[0] != (report.Group{Category: "x", Count: wantRows, Total: wantTotal}) {
		t.Fatalf("maximum supported input: %+v", got)
	}
}

func Example_main() {
	main()
	// Output:
	// category,count,total
	// 도서,2,20
	// 식품,1,7
}
```

정상 병합뿐 아니라 뒤쪽 입력 실패, 빈 입력 목록, 정확히 10개와 초과를 확인합니다.
헤더만 있는 입력은 성공한 빈 집계이므로 다른 입력과 함께 병합할 수 있습니다.
에러를 감싸도 `errors.Is`가 원인을 찾는지 검사합니다.

## 6. 완성된 출력 바이트 반환

명령행 도구의 핵심 연산을 함수로 제공하면 다른 프로그램에서도 사용할 수 있습니다.
이 함수는 형식 검증, 분석, 직렬화를 마친 후 바이트를 돌려줍니다.
성공 전에 외부 Writer에 데이터를 내보내지 않는 경계를 선택했습니다.

### 전체 코드

파일: [render/main.go](render/main.go).

```go
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

var errFormat = errors.New("format must be csv or json")

func render(input io.Reader, format string) ([]byte, error) {
	if format != "csv" && format != "json" {
		return nil, errFormat
	}
	summary, err := report.Analyze(input, report.DefaultOptions())
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if format == "csv" {
		err = report.WriteCSV(&output, summary)
	} else {
		err = report.WriteJSON(&output, summary)
	}
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func main() {
	result, err := render(strings.NewReader("category,amount\n도서,12\n도서,8\n"), "csv")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(string(result))
}
```

### 실행

```sh
go run ./solutions/29-csv-cli/render
```

예상 출력:

```text
category,count,total
도서,2,20
```

format을 먼저 검사하면 지원하지 않는 형식 때문에 입력 스트림이 소비되는 일을 피할 수 있습니다.
입력은 한 번 읽으면 되돌릴 수 없는 네트워크 스트림일 수도 있습니다.
인자만으로 판단할 수 있는 오류를 먼저 검사하는 이유입니다.

### Buffer의 범위

`bytes.Buffer`는 Writer 인터페이스를 구현하므로 기존 출력 함수를 그대로 사용할 수 있습니다.
반환한 Bytes는 그 버퍼의 저장 공간을 가리킵니다.
함수가 반환되어도 참조되는 바이트가 유효하게 남으며,
이 함수는 반환 뒤 버퍼를 다시 수정하지 않습니다.
[bytes.Buffer.Bytes](https://pkg.go.dev/bytes#Buffer.Bytes)의 별칭 조건을 확인하세요.

대신 출력 전체를 메모리에 보관해야 합니다.
매우 큰 보고서에는 직접 스트리밍하는 방식과 비교하여 필요한 계약을 선택해야 합니다.
현재 입력 한도는 이 실습이 처리하는 범위를 제한합니다.

완성된 바이트를 돌려주는 것과 파일 교체의 원자성은 별개입니다.
호출자가 `os.WriteFile`로 쓰다 실패할 수도 있고,
셸이 프로그램 시작 전에 결과 파일을 비울 수도 있습니다.
기존 파일 보존이 요구된다면 임시 파일 작성·닫기·교체 과정을 추가로 설계해야 합니다.

### 테스트

파일: [render/main_test.go](render/main_test.go).

```go
package main

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

type countedReader struct{ calls int }

func (r *countedReader) Read(p []byte) (int, error) { r.calls++; return 0, io.EOF }

func TestRender(t *testing.T) {
	input := "category,amount\nx,2\nx,3\n"
	csv, err := render(strings.NewReader(input), "csv")
	if err != nil || string(csv) != "category,count,total\nx,2,5\n" {
		t.Fatal("CSV result")
	}
	data, err := render(strings.NewReader(input), "json")
	if err != nil {
		t.Fatal(err)
	}
	var summary report.Summary
	if err := json.Unmarshal(data, &summary); err != nil || summary.Rows != 2 || summary.Total != 5 || len(summary.Groups) != 1 || summary.Groups[0] != (report.Group{Category: "x", Count: 2, Total: 5}) {
		t.Fatal("JSON result")
	}
	if data, err := render(strings.NewReader("category,amount\nx,bad\n"), "csv"); err == nil || data != nil {
		t.Fatal("partial output")
	}
	counted := &countedReader{}
	if _, err := render(counted, "xml"); !errors.Is(err, errFormat) || counted.calls != 0 {
		t.Fatal("format validation order")
	}
}

func Example_main() {
	main()
	// Output:
	// category,count,total
	// 도서,2,20
}
```

CSV는 헤더·정렬·개행이 계약이므로 바이트를 비교합니다.
JSON은 필드 순서보다 내용이 중요하므로 역직렬화한 값을 비교합니다.
읽기 횟수를 기록하는 작은 대역은 형식 검사 순서를 관찰합니다.
이 테스트는 내부 함수가 어떤 순서로 호출되는지보다 외부 입력이 소비되는지를 확인합니다.

## 7. 실행·검사 명령

네 구현을 순서대로 실행합니다.

```sh
go run ./solutions/29-csv-cli/filter
go run ./solutions/29-csv-cli/ranking
go run ./solutions/29-csv-cli/merge
go run ./solutions/29-csv-cli/render
```

단위 테스트와 실제 예제 출력을 함께 확인합니다.

```sh
go test -count=1 ./examples/29-csv-cli/... ./solutions/29-csv-cli/...
go vet ./examples/29-csv-cli/... ./solutions/29-csv-cli/...
go build ./examples/29-csv-cli/... ./solutions/29-csv-cli/...
```

지원하는 환경에서는 데이터 경쟁 검사도 실행합니다.
이 참고 구현은 순차 실행이지만 공유 패키지를 사용하는 전체 검사 기준을 유지합니다.

```sh
go test -race -count=1 ./examples/29-csv-cli/... ./solutions/29-csv-cli/...
```

분석기의 제한 시간 퍼징은 별도 명령입니다.
일반 `go test`는 저장된 seed를 실행하고, 아래 명령은 새 입력을 탐색합니다.

```sh
go test ./examples/29-csv-cli/report -run '^$' -fuzz '^FuzzAnalyze$' -fuzztime=3s -parallel=2
```

성공한 탐색 횟수와 실행 시간은 환경에 따라 다릅니다.
제한 시간 안에 실패를 찾지 않았다는 결과를 모든 입력의 정확성 증명으로 해석하지 않습니다.
반드시 성공해야 하는 정상·경계 입력을 일반 테스트로도 유지합니다.

## 8. CLI에 확장 기능을 통합한다면

이 참고 구현은 네 개의 독립 프로그램입니다.
기본 csvstats 명령에 새 옵션이 자동으로 추가되지는 않습니다.
예를 들어 `-category` 옵션을 연결한다면 다음 흐름을 설계할 수 있습니다.

1. 플래그를 읽고 옵션 자체의 계약을 검사합니다.
2. 기존 파일·표준 입력 경계를 통해 전체 입력을 분석합니다.
3. 분석이 성공하면 선택한 분류로 새 Summary를 만듭니다.
4. 원래 약속한 출력 형식으로 직렬화합니다.
5. 정상·없는 분류·잘못된 입력의 종료 코드와 stdout을 검사합니다.

`-top`을 추가하면 이름순과 합계순 출력 중 어느 계약을 사용할지 정해야 합니다.
여러 파일을 받으면 각 파일의 수명, 최대 개수, 한 입력의 실패 정책을 정해야 합니다.
옵션 이름을 추가하는 작업과 사용자에게 약속하는 동작을 함께 검토하세요.

## 9. 검토 질문

- 선택된 그룹의 요약과 전체 입력 요약을 구분했나요?
- 입력이 공유하는 슬라이스를 정렬하거나 수정하지 않았나요?
- 동률일 때도 결과가 일정한가요?
- 여러 입력의 실패에서 부분 결과가 새어 나오지 않나요?
- 오류를 감싸면서 원인을 잃지 않았나요?
- 입력 소비와 파일 닫기의 주체를 설명할 수 있나요?
- 메모리에 출력이 완성되는 것과 파일이 안전하게 교체되는 것을 구분했나요?
- 기본 CLI의 종료 코드 검사를 실제 바이너리로 실행했나요?

기준: 2026-11-09, Go 1.27.3.

[본문으로](../../book/29-csv-cli/README.md) · [연습문제로](../../exercises/29-csv-cli/README.md) · [해설 목록](../README.md)
