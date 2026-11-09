# 26장 해설 — 성능 측정과 진단

[문제](../../exercises/26-performance/README.md)를 먼저 풀고 확인하세요.
이 해설은 입력 계약을 만족하는 한 가지 구현과 측정 구성을 제공합니다.
측정값은 각자의 환경에서 기록하므로 특정 속도 순서를 정답으로 고정하지 않습니다.

## 실행 안내

모든 명령은 저장소 루트에서 실행합니다.

```sh
go test -count=1 ./solutions/26-performance/...
go test -run '^$' -bench . -benchmem -benchtime=100ms ./solutions/26-performance/...
```

| 문제 | 구현 위치 | 측정 대상 |
| --- | --- | --- |
| 3 | [join-separated](join-separated/main.go) | 구분자가 있는 문자열 결합 |
| 4 | [positive](positive/main.go) | 분포에 따른 양수 필터 |
| 5 | [query-batch](query-batch/main.go) | 준비 비용을 구분한 일괄 조회 |
| 6 | [prefix](prefix/main.go) | 독립 바이트 복사 |

각 디렉터리의 테스트는 결과·경계·입력 보존을 확인합니다.
벤치마크는 별도로 실행하며 일반 테스트에 시간 임계값을 넣지 않았습니다.

## 문제 1. 지표를 구분한 판단

가상 수치에서 B의 시간 감소율은 `(1000-800)/1000 = 0.2`, 즉 20%입니다.
B는 할당 횟수가 5회에서 2회로 줄지만 할당 바이트는 800에서 1200으로 늘었습니다.
할당 횟수와 할당량은 서로 다른 지표이므로 “메모리를 적게 쓴다”는 결론은 부족합니다.

현재 살아 있는 메모리는 또 다른 질문입니다.
임시 객체를 많이 만들고 금방 버리는 경우와,
큰 객체 한 개를 오래 보유하는 경우가 다를 수 있습니다.
힙 프로파일의 관점과 객체의 수명을 함께 살펴야 합니다.

함수의 평균 시간이 줄어도 전체 HTTP 요청에는 다른 비용이 남아 있습니다.
이 함수가 전체 시간의 일부라면 전체 지연의 감소율은 달라집니다.
응답 시간의 분포, DB·네트워크 대기까지 포함한 측정이 필요할 수 있습니다.

추가 확인의 예:

- 같은 버전·입력·환경에서 여러 번 측정했는가?
- 정상·오류·크기 분포가 실제 사용과 비슷한가?
- 출력·오류·동시 사용 계약을 계속 지키는가?
- 더 큰 할당량을 운영 환경이 감당할 수 있는가?

이 문제는 가상의 입력에서 비율과 판단 범위를 연습하는 문제입니다.
실제 측정 없이 이 비율을 특정 구현의 성능 설명에 인용하지 않습니다.

## 문제 2. 준비 비용은 없어지는 것이 아니다

서비스 P에서는 준비된 맵 조회가 요청 경로와 가깝습니다.
하지만 시작 시간과 맵을 유지하는 메모리는 별도로 고려해야 합니다.
서비스 Q에서는 매번 생성하므로 `MapBuildAndLookup`에 가까운 범위를 측정합니다.

서비스 R은 “목록 하나를 준비해 1000번 조회한다”를 한 번의 작업으로 정의할 수 있습니다.
이 경우 루프 안에 맵 생성과 조회 1000개를 모두 넣습니다.
혹은 준비와 개별 조회를 별도로 측정한 뒤 실제 반복 횟수와 함께 해석할 수 있습니다.
어느 방식을 썼는지 결과 이름과 보고서에 명시해야 합니다.

첫 원소를 조회한다면 이름을 `LinearFirst`처럼 바꿉니다.
실패 조회와 성공 조회를 섞어 같은 비용으로 간주하지 않습니다.

부족한 문장:

> 맵이 슬라이스보다 항상 빠르다.

수정한 문장:

> 원소 1000개로 준비한 자료구조에서 존재하지 않는 값 하나를 조회하는 조건을 비교했다. 생성 비용은 별도 실험으로 확인했다.

두 번째 문장은 결과의 적용 범위를 명확하게 합니다.
정확한 속도 차이는 실제 측정 후에 기록해야 합니다.

## 문제 3. 구분자 있는 결합

### 전체 구현

```go
package main

import (
	"fmt"
	"strings"
)

func joinSeparated(parts []string, sep string) string {
	var result strings.Builder
	for i, part := range parts {
		if i > 0 {
			result.WriteString(sep)
		}
		result.WriteString(part)
	}
	return result.String()
}

func main() { fmt.Println(joinSeparated([]string{"한글", "Go", "측정"}, ",")) }
```

```sh
go run ./solutions/26-performance/join-separated
```

```text
한글,Go,측정
```

구분자를 넣는 조건은 원소 내용이 아니라 인덱스입니다.
첫 원소 앞에는 구분자가 없고, 두 번째부터 구분자를 먼저 씁니다.
따라서 빈 원소 두 개도 하나의 구분자로 구분됩니다.

| i | part | 먼저 쓰는 값 | 누적 결과 |
| ---: | --- | --- | --- |
| 0 | `한글` | 없음 | `한글` |
| 1 | `Go` | `,` | `한글,Go` |
| 2 | `측정` | `,` | `한글,Go,측정` |

Builder는 호출 안에서만 사용합니다.
매번 새 지역 변수를 만들기 때문에 이전 호출 결과가 섞이지 않습니다.
표준 `strings.Join`을 직접 사용하는 대안도 있지만,
이번 문제에서는 누적 순서와 측정 경계를 연습하기 위해 Builder를 작성했습니다.

### 테스트와 벤치마크 전체

```go
package main

import (
	"strings"
	"testing"
)

func TestJoinSeparated(t *testing.T) {
	for _, tc := range []struct {
		parts []string
		sep   string
	}{
		{nil, ","}, {[]string{""}, ","}, {[]string{"", ""}, ","},
		{[]string{"a", "b"}, ""}, {[]string{"가", "나"}, " / "},
	} {
		if got, want := joinSeparated(tc.parts, tc.sep), strings.Join(tc.parts, tc.sep); got != want {
			t.Fatalf("got=%q want=%q", got, want)
		}
	}
}

func BenchmarkJoinSeparated(b *testing.B) {
	for _, size := range []struct {
		name string
		n    int
	}{{"N10", 10}, {"N100", 100}} {
		b.Run(size.name, func(b *testing.B) {
			parts := make([]string, size.n)
			for i := range parts {
				parts[i] = "Go"
			}
			b.ReportAllocs()
			for b.Loop() {
				joinSeparated(parts, ",")
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 한글,Go,측정
}
```

기준 함수와 비교하면 단순한 정상 출력뿐 아니라 빈 원소 처리도 확인할 수 있습니다.
10개와 100개 입력은 서로 다른 하위 이름을 가지며 준비는 루프 밖에서 수행합니다.
`ns/op`의 한 작업은 전체 원소를 문자열 하나로 결합하는 일입니다.

## 문제 4. 결과 개수와 예약량

### 전체 구현

```go
package main

import "fmt"

func positivesGrow(values []int) []int {
	var result []int
	for _, value := range values {
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}

func positivesReserve(values []int) []int {
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}

func main() {
	values := []int{-2, 3, 0, 5, -1}
	fmt.Println(positivesGrow(values))
	fmt.Println(positivesReserve(values))
}
```

```sh
go run ./solutions/26-performance/positive
```

```text
[3 5]
[3 5]
```

두 구현은 양수만 입력 순서대로 반환합니다.
입력 원소를 덮어쓰지 않고 별도의 결과에 append하므로 반환값 변경이 원본에 전파되지 않습니다.

`positivesReserve`는 최대 가능한 결과 개수인 입력 길이를 예약합니다.
모든 원소가 양수라면 이 예약을 채울 수 있지만,
양수가 드물면 사용하지 않는 용량이 많이 남을 수 있습니다.
할당 횟수와 바이트를 같이 살펴보아야 하는 이유입니다.

### 테스트와 벤치마크 전체

```go
package main

import (
	"slices"
	"testing"
)

func TestPositive(t *testing.T) {
	for _, tc := range []struct{ input, want []int }{
		{nil, nil}, {[]int{-1, 0}, nil}, {[]int{2, 2}, []int{2, 2}},
		{[]int{-2, 3, 0, 5, -1}, []int{3, 5}},
	} {
		before := slices.Clone(tc.input)
		for _, fn := range []func([]int) []int{positivesGrow, positivesReserve} {
			got := fn(tc.input)
			if !slices.Equal(got, tc.want) || !slices.Equal(tc.input, before) {
				t.Fatal("contract broken")
			}
			if len(got) > 0 {
				got[0] = 99
				if !slices.Equal(tc.input, before) {
					t.Fatal("shared result")
				}
			}
		}
	}
}

func BenchmarkPositive(b *testing.B) {
	for _, tc := range []struct {
		name  string
		every int
	}{{"All", 1}, {"Half", 2}, {"Sparse", 100}} {
		values := make([]int, 1000)
		for i := range values {
			if i%tc.every == 0 {
				values[i] = i + 1
			}
		}
		b.Run(tc.name, func(b *testing.B) {
			b.Run("Grow", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					positivesGrow(values)
				}
			})
			b.Run("Reserve", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					positivesReserve(values)
				}
			})
		})
	}
}

func Example_main() {
	main()
	// Output:
	// [3 5]
	// [3 5]
}
```

테스트는 입력의 복사본을 보관합니다.
호출 직후에도 입력이 같아야 하고, 결과의 첫 원소를 바꾼 뒤에도 입력이 같아야 합니다.
이 두 검사는 입력 수정과 결과 공유라는 서로 다른 결함을 찾습니다.

| 분포 | 1000개 중 양수 원소 수 | 준비 방식 |
| --- | ---: | --- |
| All | 1000 | 모든 인덱스에 i+1 |
| Half | 500 | 2의 배수 인덱스에 i+1 |
| Sparse | 10 | 100의 배수 인덱스에 i+1 |

원소 수가 같아도 결과 개수가 달라지면 할당 양상이 달라집니다.
정확성 계약이 허용하는 한, 입력 규모나 예상 비율에 맞춘 작은 초기 용량도 비교 후보가 됩니다.
그 대안을 채택할 때도 같은 분포와 입력으로 다시 측정해야 합니다.

## 문제 5. 한 번의 조회 묶음

### 전체 구현

```go
package main

import "fmt"

func hitsLinear(values, queries []int) int {
	hits := 0
	for _, query := range queries {
		for _, value := range values {
			if value == query {
				hits++
				break
			}
		}
	}
	return hits
}

func index(values []int) map[int]struct{} {
	result := make(map[int]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func hitsIndexed(set map[int]struct{}, queries []int) int {
	hits := 0
	for _, query := range queries {
		if _, ok := set[query]; ok {
			hits++
		}
	}
	return hits
}

func main() {
	values := []int{2, 4, 6}
	queries := []int{4, 5, 4}
	fmt.Println("선형:", hitsLinear(values, queries))
	fmt.Println("인덱스:", hitsIndexed(index(values), queries))
}
```

```sh
go run ./solutions/26-performance/query-batch
```

```text
선형: 2
인덱스: 2
```

선형 구현은 조회 하나가 일치하면 안쪽 반복을 종료합니다.
자료에 1이 두 번 있더라도 조회 1은 한 번만 성공합니다.
반면 조회 목록의 같은 값은 독립된 요청이므로 반복해서 셉니다.

`index`는 자료 중복을 키 하나로 합칩니다.
`hitsIndexed`가 조회를 차례로 순회하므로 조회의 중복까지 사라지지는 않습니다.

### 테스트와 벤치마크 전체

```go
package main

import "testing"

func TestHits(t *testing.T) {
	for _, tc := range []struct {
		values, queries []int
		want            int
	}{
		{nil, []int{0}, 0}, {[]int{1}, nil, 0},
		{[]int{1, 1, 2}, []int{1, 1, 3}, 2}, {[]int{2, 4, 6}, []int{4, 5, 4}, 2},
	} {
		if hitsLinear(tc.values, tc.queries) != tc.want || hitsIndexed(index(tc.values), tc.queries) != tc.want {
			t.Fatal("incorrect hits")
		}
	}
}

func BenchmarkQueryBatch(b *testing.B) {
	values := make([]int, 1000)
	for i := range values {
		values[i] = i
	}
	queries := []int{0, 500, 999, 1001}
	set := index(values)
	b.Run("Linear", func(b *testing.B) {
		for b.Loop() {
			hitsLinear(values, queries)
		}
	})
	b.Run("Prepared", func(b *testing.B) {
		for b.Loop() {
			hitsIndexed(set, queries)
		}
	})
	b.Run("BuildIncluded", func(b *testing.B) {
		for b.Loop() {
			hitsIndexed(index(values), queries)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 선형: 2
	// 인덱스: 2
}
```

`Prepared`는 준비된 맵을 재사용하고,
`BuildIncluded`는 측정 루프마다 새 맵을 만듭니다.
두 실험은 각각 다른 사용 방식을 나타내므로 제목을 유지해야 합니다.

한 번의 작업은 네 번의 조회를 모두 처리하는 것입니다.
출력의 `ns/op`를 개별 조회 한 번의 값이라고 보고하지 않습니다.
개별 평균으로 환산한다면 조회 네 개의 분포가 고정되어 있다는 설명도 남깁니다.

## 문제 6. 바이트 복사의 책임

### 전체 구현

```go
package main

import (
	"bytes"
	"errors"
	"fmt"
)

var errLength = errors.New("invalid prefix length")

func clonePrefix(data []byte, n int) ([]byte, error) {
	if n < 0 || n > len(data) {
		return nil, errLength
	}
	if n == 0 {
		return nil, nil
	}
	return bytes.Clone(data[:n]), nil
}

func main() {
	source := []byte("Go!")
	result, err := clonePrefix(source, 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	source[0] = 'N'
	fmt.Printf("prefix=%s\nsource=%s\n", result, source)
}
```

```sh
go run ./solutions/26-performance/prefix
```

```text
prefix=Go
source=No!
```

길이 범위를 검사한 뒤 슬라이스를 자르므로 범위 밖 슬라이싱을 피합니다.
`n=0`은 유효한 빈 요청이며 오류가 아닙니다.
그 경우 원본 저장 공간을 참조하지 않는 nil 결과를 반환합니다.

양수 길이에서는 `bytes.Clone`이 내용을 복사합니다.
결과가 원본과 분리되므로 이후 원본을 바꿔도 `prefix`는 유지됩니다.
[bytes.Clone](https://pkg.go.dev/bytes#Clone)의 계약을 활용한 구현입니다.

### 테스트와 벤치마크 전체

```go
package main

import (
	"errors"
	"testing"
)

func TestClonePrefix(t *testing.T) {
	for _, tc := range []struct {
		n    int
		want string
		bad  bool
	}{{-1, "", true}, {0, "", false}, {2, "Go", false}, {3, "Go!", false}, {4, "", true}} {
		input := []byte("Go!")
		got, err := clonePrefix(input, tc.n)
		if errors.Is(err, errLength) != tc.bad || string(got) != tc.want {
			t.Fatalf("n=%d got=%q err=%v", tc.n, got, err)
		}
		if len(got) > 0 {
			got[0] = 'X'
			if input[0] != 'G' {
				t.Fatal("alias")
			}
		}
	}
}

func BenchmarkClonePrefix(b *testing.B) {
	data := make([]byte, 4096)
	for _, tc := range []struct {
		name string
		n    int
	}{{"N32", 32}, {"N256", 256}, {"N4096", 4096}} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tc.n))
			for b.Loop() {
				if _, err := clonePrefix(data, tc.n); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// prefix=Go
	// source=No!
}
```

`SetBytes`에는 전체 원본 4096이 아니라 복사하는 길이를 전달합니다.
32바이트를 복사한 작업에서 4096을 적으면 처리량을 과장하게 됩니다.
오류가 발생하면 `b.Fatal`로 측정을 실패시키므로 실패 결과를 정상 비용으로 섞지 않습니다.

이 벤치마크는 복사 작업의 비용을 측정합니다.
반환된 작은 슬라이스를 수백만 개 오래 보관하는 프로그램의 힙 상태를 직접 측정하지 않습니다.
장기 보유 메모리는 수명과 보관 개수까지 정한 별도의 실험이 필요합니다.

## 선택 A. 기록의 예시 구성

아래는 수치를 채우기 전 보고서의 구조입니다.
없는 성능 결과를 만들어 넣지 말고 실제 실행 자료를 붙입니다.

### 질문

입력 원소의 대부분이 양수가 아닐 때,
최대 용량을 예약하는 필터와 필요에 따라 증가하는 필터 중 무엇을 사용할 것인가?

### 계약

양수만 순서대로 반환하고, 입력과 독립된 결과를 만든다.
nil 여부는 계약에 포함하지 않는다.

### 실험

Go 버전·OS·CPU·코드 버전·입력 분포를 적습니다.
`All`, `Half`, `Sparse`를 각각 여러 번 실행합니다.
준비는 측정 밖에 두고 필터 한 번을 작업 단위로 정의합니다.

### 관찰

시간의 반복 범위와 B/op·allocs/op를 함께 기록합니다.
프로파일을 수집했다면 수집 대상과 기간, flat·cum 관찰을 적습니다.
표본에 없는 함수를 비용 0이라고 단정하지 않습니다.

### 결정

목표가 평균 지연인지, 메모리 한도인지, 단순한 구현 유지인지 밝힙니다.
개선이 일관되지 않으면 기존 구현을 유지할 수도 있습니다.
결정에는 미검증 조건과 다음에 측정할 입력을 함께 남깁니다.

## 검증과 출처

```sh
go fmt ./solutions/26-performance/...
go vet ./solutions/26-performance/...
go test -count=1 ./solutions/26-performance/...
go test -race -count=1 ./solutions/26-performance/...
```

시간 임계값을 통과 조건으로 두지 않았습니다.
Go 1.27.3에서 일반 테스트·벤치마크를 실행하여 동작과 측정 경로를 확인했습니다.

[testing.B](https://pkg.go.dev/testing#B) · [strings.Builder](https://pkg.go.dev/strings#Builder) · [bytes.Clone](https://pkg.go.dev/bytes#Clone)

[본문](../../book/26-performance/README.md) · [문제](../../exercises/26-performance/README.md) · [전체 해설 목록](../README.md)
