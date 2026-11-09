# 26. 성능 측정과 진단

프로그램이 올바르게 실행되어도 데이터가 커지면 기다리는 시간이 길어질 수 있습니다.
이때 눈에 복잡해 보이는 줄부터 고치면, 실제 비용이 큰 부분은 남아 있을 수 있습니다.
이 장에서는 **같은 동작을 유지하면서 무엇을 얼마나 측정할지 먼저 정하는 방법**을 배웁니다.

문자열 연결, 슬라이스 생성, 반복 조회처럼 이미 배운 작업을 재료로 사용합니다.
벤치마크(benchmark)는 정해진 작업을 반복해 비용을 측정하는 실험이고,
프로파일(profile)은 실행 중 비용이 어디에 모이는지 살펴보는 자료입니다.
두 도구가 답하는 질문이 다르므로 같은 숫자로 해석하지 않습니다.

## 학습 목표

- `testing.B`로 한 번의 작업을 정의하고 하위 벤치마크를 실행한다.
- `ns/op`, `B/op`, `allocs/op`의 의미와 한계를 설명한다.
- 입력 준비와 측정 대상 작업을 구분하고 크기가 다른 입력을 비교한다.
- CPU·메모리 프로파일을 수집하고 `go tool pprof`의 주요 값을 읽는다.
- 동작 검증 → 기준 측정 → 한 가지 변경 → 재측정의 순서로 개선을 평가한다.

## 사전 지식과 준비

[12장 자동 테스트](../12-testing/README.md), 슬라이스와 맵,
[20장 공유 상태와 동기화](../20-sync/README.md)를 알고 있다고 가정합니다.
새 함수 `strings.Builder`와 프로파일 명령은 여기서 설명합니다.

- 기준 버전: 루트 [`go.mod`](../../go.mod)의 Go 1.27 계열.
- 본문 확인 환경: 2026-11-09, Go 1.27.3, Linux amd64.
- 예제: [`examples/26-performance`](../../examples/26-performance/).
- 모든 명령은 별도 안내가 없으면 **저장소 루트**에서 실행합니다.
- 일반 테스트와 벤치마크에 외부 서비스·계정은 필요하지 않습니다.
- 숫자는 CPU, 실행 중인 다른 프로그램, Go 버전과 입력에 따라 달라집니다.

벤치마크 API와 실행 조건의 기준은 [testing 문서](https://pkg.go.dev/testing#hdr-Benchmarks)입니다.
개발 PC에서 얻은 처리 시간을 운영 서버의 보장값으로 사용하지 않습니다.

## 1. 먼저 같은 일을 하는지 확인하기

두 구현의 출력이 다르면 빠르다는 비교는 의미가 없습니다.
예를 들어 오류 검사를 없앤 숫자 변환과 오류를 반환하는 숫자 변환은 책임이 다릅니다.
같은 정상 입력뿐 아니라 빈 입력, 잘못된 입력, 경계 입력을 먼저 테스트합니다.

성능 문제를 조사할 때 다음 네 항목을 한 문장으로 적어 봅니다.

| 항목 | 이 장의 문자열 예 |
| --- | --- |
| 입력 | 길이 6인 문자열 100개 |
| 한 번의 작업 | 100개를 순서대로 연결하여 문자열 하나 반환 |
| 관찰할 값 | 시간, 할당 바이트, 할당 횟수 |
| 유지할 계약 | 빈 입력은 빈 문자열, 원소의 순서와 내용 보존 |

“문자열 처리를 빠르게 한다”보다 이 문장이 실험을 다시 수행하기 쉽습니다.
입력 개수를 바꾸거나 구분자를 추가했다면 비교 대상도 함께 바꾸어야 합니다.

### 1.1 단위 테스트와 벤치마크의 역할

```sh
go test -count=1 ./examples/26-performance/...
```

위 명령은 결과를 확인합니다.
벤치마크 함수는 보통 `-bench`를 지정해야 측정됩니다.
테스트 통과는 프로그램 전체가 빠르다는 뜻이 아니며,
짧은 실행 시간도 입력 계약이 올바르다는 증거가 되지 않습니다.
[테스트 플래그](https://pkg.go.dev/cmd/go#hdr-Testing_flags)에서 두 실행 경로를 구분할 수 있습니다.

### 1.2 한 번에 바꿀 것 정하기

문자열 연결 방식을 바꾸는 실험에서 입력 크기와 출력 형식까지 함께 바꾸지 않습니다.
먼저 원본 측정값을 남기고, 변경을 적용하고, 동일한 명령으로 다시 측정합니다.
개선 후보가 실패하면 되돌릴 수 있도록 작은 변경 단위를 유지합니다.

측정값이 비슷하다면 더 복잡한 코드를 무조건 채택할 이유가 없습니다.
읽기 쉬움, 메모리 사용량, 실패 처리, 유지 비용도 함께 판단합니다.
이 판단은 언어가 정해 주는 규칙이 아니라 프로젝트의 선택입니다.

## 2. 문자열 연결을 측정하기

소스: [`join/main.go`](../../examples/26-performance/join/main.go)

```go
package main

import (
	"fmt"
	"strings"
)

func joinSlow(parts []string) string {
	result := ""
	for _, part := range parts {
		result += part
	}
	return result
}

func joinBuilder(parts []string) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part)
	}
	return b.String()
}

func main() {
	parts := []string{"Go", "로", " ", "측정"}
	fmt.Println(joinSlow(parts))
	fmt.Println(joinBuilder(parts))
}
```

```sh
go run ./examples/26-performance/join
```

예상 출력:

```text
Go로 측정
Go로 측정
```

`joinSlow`는 원소마다 기존 결과와 새 문자열을 연결합니다.
`joinBuilder`는 `strings.Builder`에 원소를 차례로 쓰고 마지막에 문자열을 얻습니다.
이름의 `Slow`는 비교 실험의 기준 구현을 구분하기 위한 이름입니다.
모든 입력에서 항상 느리다는 보장은 아닙니다.

### 2.1 Builder의 사용 책임

영값인 `strings.Builder`를 바로 사용할 수 있습니다.
`WriteString`은 내용을 이어 붙이고, `String`은 누적된 내용을 문자열로 돌려줍니다.
사용을 시작한 Builder를 값으로 복사하면 안 됩니다.
필요하면 포인터를 전달하고, 동시에 접근하려면 별도 동기화 설계를 해야 합니다.
자세한 계약은 [strings.Builder](https://pkg.go.dev/strings#Builder)에 있습니다.

이번 함수는 Builder를 지역 변수로 만들어 한 호출에서만 사용합니다.
재사용 풀이나 전역 Builder를 넣지 않았으므로 호출 사이의 상태가 섞이지 않습니다.

### 2.2 실제 테스트와 벤치마크

소스: [`join/main_test.go`](../../examples/26-performance/join/main_test.go)

```go
package main

import (
	"strings"
	"testing"
)

func TestJoin(t *testing.T) {
	for _, parts := range [][]string{nil, {}, {""}, {"가", "나"}, {"a", "", "b"}} {
		want := strings.Join(parts, "")
		if got := joinSlow(parts); got != want {
			t.Fatalf("slow=%q want=%q", got, want)
		}
		if got := joinBuilder(parts); got != want {
			t.Fatalf("builder=%q want=%q", got, want)
		}
	}
}

func BenchmarkJoin(b *testing.B) {
	parts := make([]string, 100)
	for i := range parts {
		parts[i] = "golang"
	}
	b.Run("Concat", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			joinSlow(parts)
		}
	})
	b.Run("Builder", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			joinBuilder(parts)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// Go로 측정
	// Go로 측정
}
```

`TestJoin`은 두 구현을 표준 함수 `strings.Join(parts, "")`의 결과와 비교합니다.
빈 입력과 빈 문자열 원소를 포함해 연결 규칙이 보존되는지 확인합니다.
실제 시간 측정은 `BenchmarkJoin`의 두 하위 벤치마크가 담당합니다.

- `parts` 준비는 각 측정 루프 밖에 있습니다.
- 각 `b.Loop()` 본문은 전체 연결 작업을 한 번 실행합니다.
- `b.ReportAllocs()`는 할당 지표도 보고하도록 요청합니다.
- `b.Run` 이름은 결과에서 구현을 구분하는 라벨입니다.

### 2.3 `b.Loop()`를 읽는 방법

처음부터 고정된 반복 횟수를 직접 적지 않습니다.
벤치마크 도구가 충분한 측정을 위해 반복을 조절하고 `b.Loop()`가 계속할지 알려 줍니다.
해당 API는 Go 1.24에 추가되어 이 저장소의 기준 버전에서 사용할 수 있습니다.
[`B.Loop` 문서](https://pkg.go.dev/testing#B.Loop)에 타이머와 최적화 방지 규칙이 설명되어 있습니다.

루프 앞의 준비와 루프 뒤의 정리를 측정에서 제외하는 동작이 있으므로,
측정하려는 작업은 실제로 루프 본문에 있어야 합니다.
항상 같은 값을 돌려주는 빈 함수로 바꾸면 의미 있는 작업을 측정하지 못합니다.
단순 표현식이 최적화로 사라지는지도 주의해야 하므로,
이 장에서는 결과를 계산하는 함수를 호출하여 작업의 경계를 분명히 합니다.

## 3. 측정 명령과 결과 읽기

```sh
go test -run '^$' -bench '^BenchmarkJoin$' -benchmem -benchtime=200ms -count=5 ./examples/26-performance/join
```

이 명령은 PowerShell과 Bash에서 같습니다.
작은따옴표는 정규식의 `$`를 셸이 다른 뜻으로 해석하지 않게 합니다.
`-run '^$'`는 일반 테스트 이름과 일치하지 않도록 지정한 것입니다.
앞 단계에서 일반 테스트를 따로 통과시킨 뒤 측정에 집중합니다.

| 옵션 | 이 명령에서의 의미 |
| --- | --- |
| `-bench` | 실행할 벤치마크 이름을 고르는 정규식 |
| `-benchmem` | 바이트와 횟수 형태의 할당 통계 표시 |
| `-benchtime=200ms` | 도구가 각 측정에 사용할 목표 시간 |
| `-count=5` | 같은 벤치마크를 여러 차례 반복 |
| `-run '^$'` | 일반 테스트 이름을 선택하지 않음 |

정확한 옵션 규칙은 [go test의 플래그](https://pkg.go.dev/cmd/go#hdr-Testing_flags)를 참고합니다.
`-benchtime`은 전체 명령의 최대 실행 시간을 뜻하지 않습니다.
준비, 컴파일, 여러 하위 벤치마크와 반복 횟수에 시간이 추가됩니다.

### 3.1 결과 열의 의미

도구는 실행 환경과 벤치마크별 한 줄을 출력합니다.
아래 표는 출력 형식을 해석하기 위한 설명이며 고정 성능 수치를 제시하지 않습니다.

| 출력 부분 | 해석 |
| --- | --- |
| `goos`, `goarch` | 운영체제와 아키텍처 |
| `cpu` | 도구가 보고한 CPU 정보 |
| `BenchmarkJoin/Builder-9` 같은 이름 | 벤치마크·하위 이름과 사용한 GOMAXPROCS 값 |
| 이름 뒤의 정수 | 측정을 위해 실행한 반복 횟수 |
| `ns/op` | 한 번의 작업에 사용한 평균 시간 |
| `B/op` | 한 번의 작업당 할당한 평균 바이트 |
| `allocs/op` | 한 번의 작업당 평균 할당 횟수 |

마지막 숫자가 고루틴 수나 실제 동시 실행 수를 항상 뜻하는 것은 아닙니다.
일반 순차 벤치마크와 `RunParallel`을 구분해야 합니다.
각 수치의 보고 방식은 [testing.B](https://pkg.go.dev/testing#B)를 기준으로 읽습니다.

### 3.2 단위를 바꾸어 생각하기

`ns/op`가 낮으면 이 실험에서 한 작업의 평균 시간이 짧았다는 뜻입니다.
전체 API 요청의 응답 시간이 같은 비율로 줄었다고 해석하면 안 됩니다.
요청에는 DB 대기, 네트워크, 직렬화와 다른 계산도 포함될 수 있습니다.

`B/op`는 작업 종료 후 살아 남은 메모리 양과 다릅니다.
즉시 사용하고 버린 임시 객체도 할당량에 포함될 수 있습니다.
메모리를 덜 할당하는 구현이 반드시 CPU 시간도 적게 쓰는 것은 아닙니다.

### 3.3 결과를 기록하는 양식

| 기록 항목 | 직접 적을 내용 |
| --- | --- |
| 코드 버전 | 비교한 Git 커밋 또는 변경 내용 |
| 도구 | `go version` 결과 |
| 환경 | OS, 아키텍처, CPU, GOMAXPROCS |
| 입력 | 개수, 길이, 분포, 성공·실패 비율 |
| 명령 | `-bench`, `-count`, `-benchtime`, 기타 옵션 |
| 관찰 | 반복 측정의 범위와 할당 지표 |
| 결정 | 채택 여부, 유지한 계약, 남은 불확실성 |

이 장의 예제와 벤치마크는 Go 1.27.3 Linux amd64에서 실행했습니다.
공유 개발 환경의 수치를 성능 기준선으로 고정하지 않습니다.
독자는 자신의 환경에서 같은 명령을 실행해 결과를 기록합니다.

## 4. 길이를 알고 있을 때 슬라이스 용량 준비하기

소스: [`reserve/main.go`](../../examples/26-performance/reserve/main.go)

```go
package main

import "fmt"

func squaresGrow(n int) []int {
	var result []int
	for i := 0; i < n; i++ {
		result = append(result, i*i)
	}
	return result
}

func squaresReserve(n int) []int {
	if n <= 0 {
		return nil
	}
	result := make([]int, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, i*i)
	}
	return result
}

func main() {
	fmt.Println(squaresGrow(5))
	fmt.Println(squaresReserve(5))
}
```

```sh
go run ./examples/26-performance/reserve
```

```text
[0 1 4 9 16]
[0 1 4 9 16]
```

첫 구현은 빈 슬라이스에서 시작하고, 두 번째는 필요한 용량을 먼저 준비합니다.
`make([]int, 0, n)`의 길이는 0이므로 이어서 `append`해도 앞에 0이 쌓이지 않습니다.
길이를 `n`으로 만든 뒤 같은 횟수로 `append`하면 결과 길이가 `2*n`이 됩니다.
[슬라이스 생성](https://go.dev/ref/spec#Making_slices_maps_and_channels)과
[append](https://go.dev/ref/spec#Appending_and_copying_slices)의 규칙을 다시 확인하세요.

### 4.1 입력의 범위

이 예제는 `n <= 0`이면 빈 결과를 반환합니다.
본문 실험의 양수 범위는 1~1000입니다.
임의의 큰 정수를 허용하는 공개 API라면 메모리 한도와 `i*i` 오버플로도 계약에 포함해야 합니다.
성능 실험을 이유로 입력의 제약을 생략하지 않습니다.

### 4.2 크기별 벤치마크

소스: [`reserve/main_test.go`](../../examples/26-performance/reserve/main_test.go)

```go
package main

import (
	"slices"
	"testing"
)

func TestSquares(t *testing.T) {
	for _, n := range []int{-1, 0, 1, 5, 100} {
		a, b := squaresGrow(n), squaresReserve(n)
		if !slices.Equal(a, b) {
			t.Fatalf("n=%d: %v != %v", n, a, b)
		}
		for i, value := range b {
			if value != i*i {
				t.Fatal("incorrect square")
			}
		}
	}
}

func BenchmarkSquares(b *testing.B) {
	for _, n := range []int{10, 1000} {
		b.Run(fmtSize(n), func(b *testing.B) {
			b.Run("Grow", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					squaresGrow(n)
				}
			})
			b.Run("Reserve", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					squaresReserve(n)
				}
			})
		})
	}
}

func fmtSize(n int) string {
	if n == 10 {
		return "N10"
	}
	return "N1000"
}

func Example_main() {
	main()
	// Output:
	// [0 1 4 9 16]
	// [0 1 4 9 16]
}
```

```sh
go test -run '^$' -bench '^BenchmarkSquares$' -benchmem -count=3 ./examples/26-performance/reserve
```

`N10`과 `N1000`을 비교하면 입력 크기가 달라질 때 비용이 어떻게 변하는지 관찰할 수 있습니다.
작은 입력의 결과만 보고 큰 입력에서도 같은 차이가 난다고 결론 내리지 않습니다.

길이가 알려져 있다는 조건은 이 실험의 중요한 전제입니다.
스트림에서 결과 개수를 알 수 없다면 과도한 용량 예약으로 메모리를 낭비할 수 있습니다.
알고 있는 상한과 실제 데이터 분포를 구분해 판단합니다.

### 4.3 바꾸어 볼 입력

| 변경 | 먼저 예상할 점 | 확인할 값 |
| --- | --- | --- |
| `n=0` | 반환값 길이는 얼마인가? | 두 함수 모두 길이 0 |
| `n=1` | 첫 원소는 무엇인가? | `[0]` |
| `make([]int, n)`으로 변경 | append 앞에 어떤 값이 존재하는가? | 길이와 앞쪽 영값 |
| 예약량을 실제 길이보다 크게 지정 | 할당량과 결과 길이는 같은가? | 용량·B/op·len을 구분 |

## 5. 조회만 잴지, 자료구조 생성까지 잴지 정하기

소스: [`lookup/main.go`](../../examples/26-performance/lookup/main.go)

```go
package main

import "fmt"

func contains(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func makeSet(values []int) map[int]struct{} {
	result := make(map[int]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func containsSet(set map[int]struct{}, target int) bool {
	_, found := set[target]
	return found
}

func main() {
	values := []int{3, 5, 8}
	set := makeSet(values)
	_, found := set[8]
	fmt.Println(contains(values, 8), found)
	_, found = set[7]
	fmt.Println(contains(values, 7), found)
}
```

```sh
go run ./examples/26-performance/lookup
```

```text
true true
false false
```

슬라이스에서 순서대로 찾는 구현과 맵을 만들어 조회하는 구현입니다.
둘 다 같은 존재 여부를 답하지만, 맵을 만드는 데도 시간과 메모리가 듭니다.
조회가 한 번인지 수천 번인지에 따라 준비 비용을 부담하는 이유가 달라집니다.

### 5.1 서로 다른 질문을 가진 세 실험

소스: [`lookup/main_test.go`](../../examples/26-performance/lookup/main_test.go)

```go
package main

import "testing"

func TestLookup(t *testing.T) {
	values := []int{1, 1, 3, 9}
	set := makeSet(values)
	if len(set) != 3 {
		t.Fatal("duplicate key")
	}
	for _, target := range []int{0, 1, 3, 9, 10} {
		_, got := set[target]
		if got != contains(values, target) {
			t.Fatalf("target=%d", target)
		}
	}
	if contains(nil, 0) {
		t.Fatal("empty input")
	}
}

func BenchmarkLookup(b *testing.B) {
	values := make([]int, 1000)
	for i := range values {
		values[i] = i
	}
	set := makeSet(values)
	b.Run("LinearMissing", func(b *testing.B) {
		for b.Loop() {
			contains(values, -1)
		}
	})
	b.Run("MapMissingPrepared", func(b *testing.B) {
		for b.Loop() {
			containsSet(set, -1)
		}
	})
	b.Run("MapBuildAndLookup", func(b *testing.B) {
		for b.Loop() {
			current := makeSet(values)
			containsSet(current, -1)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// true true
	// false false
}
```

```sh
go test -run '^$' -bench '^BenchmarkLookup$' -benchmem -count=3 ./examples/26-performance/lookup
```

| 이름 | 측정에 포함한 일 | 답하려는 질문 |
| --- | --- | --- |
| `LinearMissing` | 슬라이스 전체에서 없는 값 찾기 | 선형 탐색의 실패 조회 비용은? |
| `MapMissingPrepared` | 이미 만든 맵에서 없는 키 찾기 | 준비된 맵의 조회 비용은? |
| `MapBuildAndLookup` | 맵 생성 후 한 번 조회 | 요청마다 새로 만든다면 총 비용은? |

이름에 `Missing`을 넣은 이유는 실패 조회가 끝까지 검사하는 입력이기 때문입니다.
슬라이스 첫 원소를 찾는 경우와 같은 작업으로 취급하지 않습니다.
맵 순회 순서나 내부 버킷 배치는 이 실험의 계약에 넣지 않습니다.

### 5.2 재사용 횟수라는 조건

시작할 때 한 번 만든 맵을 오래 재사용한다면 준비 비용을 여러 조회가 나누어 부담합니다.
요청마다 맵을 다시 만든다면 생성 비용과 할당이 매번 발생합니다.
이 두 상황을 한 줄의 `ns/op`로 섞으면 설계 판단을 잘못할 수 있습니다.

다음 변화는 별도의 하위 벤치마크로 비교할 수 있습니다.

- 성공 조회: 첫 원소, 중간 원소, 마지막 원소.
- 실패 조회: 데이터에 없는 값.
- 데이터 크기: 10개, 1000개.
- 준비 방식: 한 번 준비, 매 작업 준비.

측정 범위가 다른 결과를 단순히 “맵이 몇 배 빠르다”로 요약하지 않습니다.
정확히 어떤 데이터와 재사용 조건에서 관찰한 차이인지 함께 남깁니다.

## 6. 처리 바이트와 오류 경로

소스: [`parse/main.go`](../../examples/26-performance/parse/main.go)

```go
package main

import (
	"fmt"
	"strconv"
)

func sumNumbers(values []string) (int, error) {
	total := 0
	for i, value := range values {
		number, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("item %d: %w", i+1, err)
		}
		total += number
	}
	return total, nil
}

func main() {
	total, err := sumNumbers([]string{"10", "20", "30"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("합계:", total)
}
```

```sh
go run ./examples/26-performance/parse
```

```text
합계: 60
```

`strconv.Atoi`는 문자열 전체를 정수로 해석하며 오류를 반환할 수 있습니다.
숫자를 읽다가 잘못된 원소를 만나면 항목 번호를 붙여 호출자에게 돌려줍니다.
반환된 0은 부분 합계를 의미하지 않습니다.
이 함수의 예제 입력은 합계가 `int` 범위 안에 있도록 작게 제한했습니다.
[정수 변환의 계약](https://pkg.go.dev/strconv#Atoi)을 확인하세요.

### 6.1 `SetBytes`가 표현하는 것

소스: [`parse/main_test.go`](../../examples/26-performance/parse/main_test.go)

```go
package main

import (
	"errors"
	"strconv"
	"testing"
)

func TestSumNumbers(t *testing.T) {
	for _, tc := range []struct {
		values []string
		want   int
		bad    bool
	}{
		{nil, 0, false}, {[]string{"10", "-3"}, 7, false},
		{[]string{"10", "x"}, 0, true}, {[]string{""}, 0, true},
	} {
		got, err := sumNumbers(tc.values)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("got %d, %v", got, err)
		}
		if tc.bad && !errors.Is(err, strconv.ErrSyntax) {
			t.Fatalf("lost cause: %v", err)
		}
	}
}

func BenchmarkSumNumbers(b *testing.B) {
	values := []string{"10", "20", "30", "40"}
	b.SetBytes(int64(len("10203040")))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := sumNumbers(values); err != nil {
			b.Fatal(err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 합계: 60
}
```

`SetBytes`에는 **한 번의 작업**이 처리하는 바이트 수를 전달합니다.
도구는 이를 시간과 조합해 처리량을 표시할 수 있습니다.
여기서는 네 문자열의 숫자 문자 8바이트를 기준으로 합니다.
메모리에 존재하는 슬라이스 헤더의 크기나 HTTP 패킷 크기를 뜻하지 않습니다.
[`B.SetBytes`](https://pkg.go.dev/testing#B.SetBytes)의 정의와 일치하도록 단위를 정합니다.

```sh
go test -run '^$' -bench '^BenchmarkSumNumbers$' -benchmem ./examples/26-performance/parse
```

### 6.2 성공 입력만 측정했다는 한계

이 벤치마크는 정상적인 숫자 네 개만 사용합니다.
오류를 감싸는 `fmt.Errorf` 경로는 결과에 포함되지 않습니다.
실제 입력에 오류가 많다면 오류 위치와 비율을 정해 별도 측정해야 합니다.

첫 원소에서 실패하는 입력과 마지막 원소에서 실패하는 입력도 비용이 다릅니다.
실패를 무시해 루프를 끝까지 돌도록 바꾸면 반환 계약이 바뀌므로 같은 구현 개선이 아닙니다.
성능 수치 옆에 정상·오류 분포를 기록하는 이유입니다.

## 7. CPU 프로파일 만들기

벤치마크는 정해진 작업의 비용을 알려 줍니다.
여러 함수가 호출되는 실제 작업에서 어느 부분이 비용을 사용하는지 찾으려면 프로파일을 봅니다.
CPU 프로파일은 실행 중인 프로그램을 표본으로 관찰합니다.
표본이 적거나 작업이 너무 짧으면 유용한 단서를 얻기 어렵습니다.
[Go 진단 안내](https://go.dev/doc/diagnostics)에서 도구별 목적을 확인할 수 있습니다.

소스: [`workload/main.go`](../../examples/26-performance/workload/main.go)

```go
package main

import "fmt"

func checksum(data []byte, rounds int) uint64 {
	var sum uint64
	for r := 0; r < rounds; r++ {
		for _, value := range data {
			sum = sum*33 + uint64(value)
		}
	}
	return sum
}

func main() {
	fmt.Println("체크섬:", checksum([]byte("Go"), 2))
}
```

```sh
go run ./examples/26-performance/workload
```

```text
체크섬: 2674860
```

이 함수는 바이트를 여러 번 훑으며 `uint64` 값을 누적합니다.
부호 없는 정수의 범위를 넘으면 정해진 모듈러 연산으로 값이 돌아갑니다.
이 함수는 프로파일 연습용이며 암호학적 해시나 무결성 보장 도구가 아닙니다.
[정수 오버플로 규칙](https://go.dev/ref/spec#Integer_overflow)을 적용한 계산입니다.

### 7.1 측정할 실행 경로

소스: [`workload/main_test.go`](../../examples/26-performance/workload/main_test.go)

```go
package main

import "testing"

func TestChecksum(t *testing.T) {
	if got := checksum([]byte("Go"), 2); got != 2674860 {
		t.Fatalf("got %d", got)
	}
	if checksum(nil, 100) != 0 || checksum([]byte("Go"), 0) != 0 {
		t.Fatal("empty work")
	}
}

func BenchmarkChecksum(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data) * 50))
	for b.Loop() {
		checksum(data, 50)
	}
}

func Example_main() {
	main()
	// Output:
	// 체크섬: 2674860
}
```

기본 실행은 짧아서 동작을 확인하기 좋습니다.
벤치마크는 4096바이트를 50회 처리하는 작업을 반복하여 프로파일을 수집합니다.
문자열 연결 예제에서 나온 문제를 조사한다면 그 예제의 벤치마크로 대상을 바꾸어야 합니다.
프로파일 파일 이름만 바꾸어서는 다른 프로그램의 병목을 알 수 없습니다.

### 7.2 파일 저장 위치 준비

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/26-performance/profiles | Out-Null
```

Bash:

```sh
mkdir -p _practice/26-performance/profiles
```

다음 명령은 두 셸에서 같습니다.
같은 이름으로 다시 실행하면 **이 실습의 이전 프로파일 파일을 교체**합니다.
비교할 자료는 이름을 바꾸어 보관하세요.

```sh
go test -run '^$' -bench BenchmarkChecksum -benchtime=1s -cpuprofile ./_practice/26-performance/profiles/cpu.prof -memprofile ./_practice/26-performance/profiles/mem.prof -o ./_practice/26-performance/profiles/workload.test ./examples/26-performance/workload
```

`-o`로 검사에 사용한 테스트 실행 파일도 함께 남깁니다.
실행 파일과 프로파일이 다른 코드 버전이면 소스 위치 해석이 혼동될 수 있습니다.
이번에 만든 두 자료를 한 쌍으로 보관합니다.
수집 방법의 기준은 [runtime/pprof](https://pkg.go.dev/runtime/pprof#hdr-Profiling_a_Go_program)입니다.

### 7.3 터미널에서 읽기

```sh
go tool pprof -top ./_practice/26-performance/profiles/workload.test ./_practice/26-performance/profiles/cpu.prof
```

CPU 결과의 시간과 비율은 실행마다 달라집니다.
이 예제에서는 계산 함수인 `checksum`과 인라이닝된 관련 위치를 중심으로 읽습니다.
함수 이름이 항상 같은 줄에 나타난다고 테스트하지 않습니다.
컴파일러의 인라이닝과 표본 수가 표시에 영향을 줍니다.

| 열 | 살펴볼 의미 |
| --- | --- |
| `flat` | 해당 함수 자체에 귀속된 표본 비용 |
| `cum` | 그 함수와 호출한 함수들을 포함한 비용 |
| `%` | 이 프로파일의 전체 표본에서 차지하는 비율 |

호출만 하는 함수는 `flat`이 작고 `cum`이 클 수 있습니다.
따라서 첫 번째 열만 정렬해 호출 관계를 무시하지 않습니다.
프로파일은 이번 실행에서 관찰한 자료이며 모든 입력의 경로를 보여 주지 않습니다.

### 7.4 필요한 경우 브라우저로 보기

```sh
go tool pprof -http=127.0.0.1:8081 ./_practice/26-performance/profiles/workload.test ./_practice/26-performance/profiles/cpu.prof
```

명령이 실행 중인 동안 로컬 분석 UI가 열립니다.
종료는 해당 터미널에서 `Ctrl+C`로 합니다.
일부 그래프 기능은 Graphviz 같은 별도 도구가 필요할 수 있습니다.
기본 학습은 설치가 추가로 필요 없는 `-top` 결과로 진행할 수 있습니다.
프로파일에는 함수명과 경로 같은 내부 정보가 담길 수 있어 외부 서비스로 자동 업로드하지 않습니다.

## 8. 할당량과 살아 있는 메모리 구분하기

```sh
go tool pprof -top -alloc_space ./_practice/26-performance/profiles/workload.test ./_practice/26-performance/profiles/mem.prof
go tool pprof -top -inuse_space ./_practice/26-performance/profiles/workload.test ./_practice/26-performance/profiles/mem.prof
```

`alloc_space`는 누적 할당 바이트 관점으로,
`inuse_space`는 수집 시점에 사용 중인 바이트 관점으로 자료를 읽습니다.
힙 프로파일도 표본과 수집 시점의 영향을 받습니다.
작은 함수의 결과가 표에 없다는 이유만으로 메모리를 전혀 사용하지 않는다고 단정하지 않습니다.
[힙 프로파일의 설명](https://pkg.go.dev/runtime/pprof#Profile)을 참고하세요.

앞의 체크섬 반복 자체에는 새 컨테이너를 만드는 코드가 없습니다.
그 프로파일에는 테스트 준비나 런타임 비용이 더 눈에 띌 수 있습니다.
문자열 연결의 할당을 조사하려면 `join` 벤치마크로 다시 수집해야 합니다.
관찰 대상과 수집 대상이 일치하는지 항상 먼저 확인합니다.

### 8.1 작은 결과가 큰 저장 공간을 붙잡는 경우

소스: [`retention/main.go`](../../examples/26-performance/retention/main.go)

```go
package main

import (
	"fmt"
	"slices"
)

func firstBytes(data []byte, n int) []byte {
	if n <= 0 || len(data) == 0 {
		return nil
	}
	if n > len(data) {
		n = len(data)
	}
	return slices.Clone(data[:n])
}

func main() {
	source := []byte("Go language")
	prefix := firstBytes(source, 2)
	source[0] = 'N'
	fmt.Println(string(prefix))
	fmt.Println(string(source))
}
```

```sh
go run ./examples/26-performance/retention
```

```text
Go
No language
```

`data[:n]`은 원본 저장 공간을 공유합니다.
작은 부분만 남겨도 큰 원본 배열이 계속 필요해질 수 있습니다.
여기서는 `slices.Clone`으로 필요한 바이트를 독립된 결과에 복사합니다.
요청한 길이가 0 이하이거나 입력이 비어 있으면 `nil`을 반환해 원본을 참조하지 않습니다.
[Clone](https://pkg.go.dev/slices#Clone)은 원소를 얕게 복사하는 함수입니다.
바이트 원소에는 추가 참조가 없지만 포인터를 담은 슬라이스라면 대상 객체까지 복제하지는 않습니다.

### 8.2 복사 비용도 함께 생각하기

독립된 결과를 만들면 복사와 할당 비용이 생깁니다.
원본을 오래 보관할 필요가 없는 경우에는 그 비용이 불필요할 수 있습니다.
`B/op`를 줄이는 것과 장기 보유 메모리를 줄이는 것이 같은 방향이 아닐 수 있는 예입니다.

```sh
go test -run '^$' -bench '^BenchmarkFirstBytes$' -benchmem ./examples/26-performance/retention
```

이 벤치마크는 1MiB 입력 중 앞의 32바이트를 복사하는 비용을 봅니다.
프로그램 전체의 장기 메모리 보유량을 직접 재현하는 실험은 아닙니다.
`retention/main_test.go`는 반환값을 수정해도 원본이 바뀌지 않는 계약을 검증합니다.

## 9. 병렬 벤치마크와 공유 자원

소스: [`parallel/main.go`](../../examples/26-performance/parallel/main.go)

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func count(workers, times int) int64 {
	var total atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < times; j++ {
				total.Add(1)
			}
		}()
	}
	wg.Wait()
	return total.Load()
}

func main() { fmt.Println("처리:", count(4, 100)) }
```

```sh
go run ./examples/26-performance/parallel
```

```text
처리: 400
```

고루틴의 실행 순서는 고정하지 않았지만 합계는 완료 대기 뒤에 확인합니다.
각 증가에는 원자 연산을 사용하므로 공유 정수의 데이터 경쟁을 피합니다.
동기화 정확성은 성능 측정 전에 확인해야 합니다.

### 9.1 `RunParallel` 사용

소스: [`parallel/main_test.go`](../../examples/26-performance/parallel/main_test.go)

```go
package main

import (
	"sync/atomic"
	"testing"
)

func TestCount(t *testing.T) {
	if count(0, 10) != 0 || count(2, 0) != 0 || count(4, 100) != 400 {
		t.Fatal("wrong total")
	}
}

func BenchmarkAtomicParallel(b *testing.B) {
	var total atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			total.Add(1)
		}
	})
	if total.Load() != int64(b.N) {
		b.Fatal("lost update")
	}
}

func Example_main() {
	main()
	// Output:
	// 처리: 400
}
```

```sh
go test -run '^$' -bench '^BenchmarkAtomicParallel$' -benchmem -cpu=1,2,4 ./examples/26-performance/parallel
```

`RunParallel`은 반복 작업을 여러 고루틴에 나누고 `pb.Next()`로 각 작업을 진행합니다.
순차 벤치마크의 `b.Loop()`를 그 안에 중첩하지 않습니다.
한 번의 작업은 공유 카운터를 1 증가시키는 것입니다.
[`B.RunParallel`](https://pkg.go.dev/testing#B.RunParallel)의 동작을 기준으로 읽으세요.

이 수치는 공유 원자 변수 하나에 작업이 몰리는 상황입니다.
실제 서비스의 모든 병렬 처리를 대표하지 않습니다.
고루틴 수를 늘려도 공유 자원 경쟁이나 다른 병목 때문에 처리량이 늘지 않을 수 있습니다.

### 9.2 레이스 검사와 시간 측정을 분리하기

```sh
go test -race -count=1 ./examples/26-performance/...
```

레이스 검사는 별도 검증 실행으로 진행합니다.
검사를 위한 계측이 실행 비용을 바꾸므로 일반 빌드의 시간과 직접 비교하지 않습니다.
검출되지 않았다는 결과도 실행하지 않은 경로까지 안전함을 증명하지 않습니다.
[레이스 검사의 영향](https://go.dev/doc/articles/race_detector)과 검사 범위를 함께 기록하세요.

## 10. 직접 실험하기: 결합 결과에 구분자 추가

개인 실습 경로는 `_practice/26-performance/join/`입니다.
기존 파일을 보존하려면 다음과 같이 준비합니다.

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/26-performance/join | Out-Null
if (-not (Test-Path _practice/26-performance/join/main.go)) {
    Copy-Item examples/26-performance/join/main.go _practice/26-performance/join/main.go
}
if (-not (Test-Path _practice/26-performance/join/main_test.go)) {
    Copy-Item examples/26-performance/join/main_test.go _practice/26-performance/join/main_test.go
}
```

Bash:

```sh
mkdir -p _practice/26-performance/join
if [ ! -e _practice/26-performance/join/main.go ]; then
    cp examples/26-performance/join/main.go _practice/26-performance/join/main.go
fi
if [ ! -e _practice/26-performance/join/main_test.go ]; then
    cp examples/26-performance/join/main_test.go _practice/26-performance/join/main_test.go
fi
```

### 10.1 먼저 출력 계약 바꾸기

두 함수가 원소 사이에만 `,`를 넣도록 변경합니다.
빈 입력에는 빈 문자열을 반환하고, 끝에 구분자를 붙이지 않습니다.
입력 `["Go", "측정"]`의 기대 결과는 `Go,측정`입니다.
입력 `["", ""]`는 빈 원소가 두 개이므로 `,`입니다.

테스트의 기준 함수도 `strings.Join(parts, ",")`로 맞춥니다.
`main`의 출력과 `Example_main`도 새 계약에 맞게 수정해야 합니다.
한 구현만 바꾼 뒤 벤치마크를 비교하지 않습니다.

### 10.2 검증 후 측정

```sh
go test -count=1 ./_practice/26-performance/join
go test -run '^$' -bench '^BenchmarkJoin$' -benchmem -count=5 ./_practice/26-performance/join
```

`_practice`는 `./...`에서 자동으로 제외되므로 정확한 경로로 실행합니다.
이번 실험의 입력 개수, 원소 길이, 구분자 길이를 기록합니다.
수치가 달라진 이유를 연결 횟수만으로 단정하지 말고 할당 지표도 확인합니다.

### 10.3 완료 조건

- 빈 입력, 원소 한 개, 빈 원소 두 개, 한글 입력의 결과가 올바릅니다.
- 두 구현의 테스트와 예제 출력이 같습니다.
- 같은 입력에 대해 5회 측정한 결과를 보관했습니다.
- 시간과 할당량 중 어떤 지표를 개선하려 했는지 설명할 수 있습니다.
- 개선하지 못했더라도 결과와 판단 근거를 남겼습니다.

## 11. 자주 하는 실수

| 증상 | 원인 후보 | 확인과 수정 |
| --- | --- | --- |
| 벤치마크가 출력되지 않음 | `-bench` 누락, 이름 불일치 | `BenchmarkXxx(*testing.B)`와 정규식 확인 |
| 측정이 지나치게 빠름 | 작업 생략·컴파일러 최적화 | 같은 결과를 계산하는 함수가 호출되는지 확인 |
| 구현마다 입력이 다름 | 준비 코드나 성공률 차이 | 입력·계약을 먼저 일치시킴 |
| 생성 비용이 사라짐 | 준비를 루프 밖에 둠 | 측정 질문에 따라 안팎을 결정 |
| 테스트 로그가 시간을 지배 | 반복마다 출력 | 측정 중의 불필요한 로그 제거 |
| 프로파일이 거의 비어 있음 | 작업이 너무 짧음 | 입력과 측정 시간을 늘려 유효한 표본 확보 |
| 메모리 지표가 서로 다름 | 누적 할당과 현재 사용량 혼동 | `alloc_space`와 `inuse_space` 구분 |
| 실행할 때마다 수치가 크게 흔들림 | 부하·전력·입력 조건 변화 | 조건 기록, 반복 측정, 동시 작업 감소 |
| 병렬 결과가 일관되지 않음 | 공유 상태의 경쟁 | 정확성 테스트와 레이스 검사 먼저 수행 |
| 이전 코드의 줄이 표시됨 | 실행 파일과 프로파일 불일치 | 같은 빌드에서 다시 수집 |

### 11.1 평균 한 숫자의 한계

같은 평균을 가진 두 요청 집합도 일부 요청이 매우 느릴 수 있습니다.
벤치마크 한 작업의 평균과 서비스 응답 시간의 분포는 서로 다른 관찰입니다.
이 장에서 측정한 함수의 개선이 실제 지연을 줄였는지는 더 큰 실행 경로에서 확인합니다.
측정 범위를 넓힐 때는 네트워크·DB·동시 부하 조건도 기록해야 합니다.

### 11.2 최적화 때문에 책임을 잃는 경우

오류를 무시하거나 결과 슬라이스를 전역에서 공유하면 숫자가 작아 보일 수 있습니다.
하지만 호출 사이에 값이 섞이거나 오류 계약을 지키지 못한다면 채택할 수 없습니다.
빠르게 보이는 구현을 만들기 전에 테스트가 무엇을 지키는지 다시 읽습니다.

## 12. 연습문제와 해설

[26장 연습문제](../../exercises/26-performance/README.md)는
측정 결과 해석, 입력 크기 비교, 문자열 결합, 조회 준비 비용,
바이트 복사, 보고서 작성을 포함합니다.
[해설과 참고 구현](../../solutions/26-performance/README.md)은
같은 결과를 유지하는 테스트와 실행 가능한 벤치마크를 제공합니다.

```sh
go test -count=1 ./examples/26-performance/... ./solutions/26-performance/...
go test -run '^$' -bench . -benchmem -benchtime=100ms ./solutions/26-performance/...
```

기본 테스트는 성능 임계값을 하드코딩하지 않습니다.
공유 CI 머신의 부하 때문에 의미 없는 실패가 생기는 것을 피하고,
정확성과 측정 가능성을 확인합니다.

## 13. 요약과 자기 점검

측정 전에 입력 계약과 한 번의 작업을 정합니다.
`testing.B`로 반복 측정을 수행하고 시간·할당의 단위를 구분합니다.
프로파일로 비용이 모이는 실행 경로를 살핀 뒤 작은 변경을 적용합니다.
마지막에는 정확성 테스트와 같은 조건의 측정을 다시 실행합니다.

### 확인 질문

1. `B/op`와 현재 살아 있는 힙 메모리 양은 왜 다를 수 있나요?
2. 맵 생성 비용을 루프 밖에 두면 어떤 상황을 측정하나요?
3. `n=10`의 결과를 `n=1000000`에 그대로 적용할 수 있나요?
4. `-run '^$'`와 `-bench`는 각각 무엇을 선택하나요?
5. 결과를 출력하는 코드가 측정 루프 안에 있으면 무엇이 달라지나요?
6. `flat`과 `cum`의 차이를 호출 관계로 설명할 수 있나요?
7. 프로파일 파일과 실행 파일은 왜 함께 보관하나요?
8. 작은 슬라이스를 복사하는 것이 메모리 사용에 도움이 되는 조건은 무엇인가요?
9. `RunParallel` 결과를 전체 HTTP 서버 처리량으로 보아도 되나요?
10. 성능 개선 뒤에도 동일한 테스트를 다시 실행하는 이유는 무엇인가요?

<details>
<summary>답 확인</summary>

1. 누적 할당에는 곧 버린 객체도 포함되지만 현재 사용량은 수집 시점과 수명에 영향을 받습니다.
2. 이미 만든 자료구조를 재사용하며 조회하는 상황을 측정합니다.
3. 데이터 크기가 바뀌면 할당과 탐색 비용도 달라지므로 별도 입력으로 측정해야 합니다.
4. `-run`은 일반 테스트 등을, `-bench`는 벤치마크를 이름으로 선택합니다.
5. 실제 계산 외에 출력·동기화·입출력 비용이 함께 들어갑니다.
6. `flat`은 함수 자체, `cum`은 호출한 함수들을 포함한 표본 비용입니다.
7. 다른 코드의 심벌과 위치를 사용하면 해석을 잘못할 수 있기 때문입니다.
8. 큰 원본을 더 이상 사용할 필요 없이 작은 결과만 오래 유지할 때 검토할 수 있습니다.
9. 원자 변수 하나를 공유하는 제한된 작업이며 전체 서버의 여러 병목을 대표하지 않습니다.
10. 변경이 결과·오류·상태 공유 계약을 깨지 않았는지 확인해야 합니다.

</details>

### 완료 체크리스트

- [ ] 같은 계약을 지키는 두 구현을 테스트했습니다.
- [ ] 준비 비용을 측정에 포함할지 이유를 적었습니다.
- [ ] 입력 크기 두 개 이상으로 벤치마크를 실행했습니다.
- [ ] 시간·할당·처리 바이트의 단위를 설명할 수 있습니다.
- [ ] CPU 프로파일을 생성하고 `-top`으로 읽었습니다.
- [ ] 결과를 일반화할 수 없는 조건을 보고서에 적었습니다.

## 참고 자료

- [testing — 벤치마크](https://pkg.go.dev/testing#hdr-Benchmarks)
- [testing.B.Loop](https://pkg.go.dev/testing#B.Loop)
- [testing.B.RunParallel](https://pkg.go.dev/testing#B.RunParallel)
- [go test 플래그](https://pkg.go.dev/cmd/go#hdr-Testing_flags)
- [Go 진단 도구](https://go.dev/doc/diagnostics)
- [runtime/pprof](https://pkg.go.dev/runtime/pprof)
- [strings.Builder](https://pkg.go.dev/strings#Builder)
- [slices.Clone](https://pkg.go.dev/slices#Clone)
- [레이스 검사](https://go.dev/doc/articles/race_detector)

문서와 실행 확인 기준: 2026-11-09, Go 1.27.3.
OS·하드웨어·런타임 설정을 바꾼 뒤의 성능 수치는 별도로 측정해야 합니다.

## 다음 학습

다음 [27장 퍼징과 견고한 테스트](../27-fuzzing/README.md)에서는
직접 고른 입력 외에 도구가 생성하는 입력으로 불변 조건을 검사합니다.

[이전: 25. 서비스 구성과 운영 기초](../25-service-operations/README.md) · [전체 목차](../README.md) · [다음: 27. 퍼징과 견고한 테스트](../27-fuzzing/README.md)
