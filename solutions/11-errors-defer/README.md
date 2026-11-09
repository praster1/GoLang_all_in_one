# 11장 해설: 오류 처리와 자원 정리

[연습문제](../../exercises/11-errors-defer/README.md)를 먼저 시도하세요. 모든 명령은 저장소 루트에서 실행합니다. 참고 구현은 오류 문자열이 비슷하게 보이는지뿐 아니라 원인 구조와 자원 수명 계약을 확인합니다.

## 문제 1. 반환식 뒤에도 defer가 실행됩니다

출력 순서는 다음과 같습니다.

```text
정리 중 결과: 12
저장된 결과: 3
최종: 12
```

| 시점 | result | 저장된 인자 |
| --- | ---: | ---: |
| 첫 대입 | 3 | 아직 없음 |
| 첫 defer 등록 | 3 | 3 |
| return 10의 반환 변수 대입 | 10 | 3 |
| 둘째 defer 실행 | 12 | 3 |
| 첫 defer 실행 | 12 | 3 출력 |
| 호출자에게 전달 | 12 | 관련 없음 |

main의 Println은 인자인 calculate 호출이 끝난 다음 실행됩니다. calculate의 defer까지 마쳐야 호출 결과를 받을 수 있습니다.

참고 구현: [predict/main.go](predict/main.go).

```go
package main

import "fmt"

func calculate() (result int) {
	result = 3
	defer fmt.Println("저장된 결과:", result)
	defer func() { result += 2; fmt.Println("정리 중 결과:", result) }()
	return 10
}
func main() { fmt.Println("최종:", calculate()) }
```

```sh
go run ./solutions/11-errors-defer/predict
go test ./solutions/11-errors-defer/predict
```

예상 출력:

```text
정리 중 결과: 12
저장된 결과: 3
최종: 12
```


이름 붙인 반환값을 쓴다는 이유만으로 모든 return을 맨몸 `return`으로 바꿀 필요는 없습니다. 여기서는 `return 10`이 반환 변수를 먼저 설정한다는 규칙을 분명히 보여 줍니다.

## 문제 2. 변환 오류와 사용 범위 오류

문자열을 정수로 만들 수 있는가와 그 정수가 포트로 허용되는가는 별도 질문입니다. 먼저 Atoi 오류를 확인한 뒤 범위 검사를 합니다.

참고 구현: [port/main.go](port/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrPortRange = errors.New("port must be 1..65535")

func parsePort(text string) (int, error) {
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("parse port: %w", err)
	}
	if value < 1 || value > 65535 {
		return 0, fmt.Errorf("port %d: %w", value, ErrPortRange)
	}
	return value, nil
}
func main() {
	port, err := parsePort("8080")
	fmt.Println("포트:", port, "성공:", err == nil)
	_, err = parsePort("0")
	fmt.Println("범위:", errors.Is(err, ErrPortRange))
	_, err = parsePort("abc")
	var detail *strconv.NumError
	fmt.Println("형식:", errors.As(err, &detail))
}
```

```sh
go run ./solutions/11-errors-defer/port
go test ./solutions/11-errors-defer/port
```

예상 출력:

```text
포트: 8080 성공: true
범위: true
형식: true
```


`"65536"`은 int로 변환할 수 있지만 포트 범위를 벗어납니다. 반면 int의 표현 범위를 넘는 아주 긴 숫자는 Atoi 단계에서 실패하므로 NumError를 찾게 됩니다. 이 둘을 모두 ErrPortRange로 뭉뚱그리지 않았습니다.

`errors.As(err, &detail)`의 detail은 `*strconv.NumError` 변수입니다. `&detail`을 전달하는 이유는 함수가 그 변수에 찾은 오류 포인터를 대입해야 하기 때문입니다.

앞뒤 공백 제거는 계약에 없습니다. `" 80"`을 자동으로 허용하도록 바꾼 풀이는 나쁜 기능인 것은 아니지만 이 문제의 계약과 다릅니다. 정책을 바꾸려면 문제·문서·테스트의 기대도 함께 바꿉니다.

## 문제 3. 코드 검토 해설

### A. 같은 글자의 새 오류는 같은 대표 오류가 아닙니다

이미 정의한 `ErrNotFound`를 비교 대상으로 전달합니다.

```go
if errors.Is(err, ErrNotFound) {
    fmt.Println("없음")
}
```

새 errors.New 호출은 같은 메시지의 다른 값을 만듭니다. 오류 문구는 사용자에게 의미를 전달하고, 대표 오류 값은 프로그램의 분류 계약을 전달합니다.

### B. 새 지역 변수가 반환 변수를 가립니다

`err := closeResource()`는 defer 함수 안에 새 err를 만듭니다. 이름 붙인 바깥 반환 err는 nil로 남습니다. 닫기 원인을 다른 이름으로 받고 바깥 err에 대입하거나 Join해야 합니다.

```go
if closeErr := closeResource(); closeErr != nil {
    err = errors.Join(err, closeErr)
}
```

이것은 설명용 발췌입니다. 실제 자원 함수는 아래 문제 5에 있습니다. 단순히 fmt.Println으로 출력했다고 호출자가 실패를 받는 것은 아닙니다.

### C. 성공은 nil 인터페이스

성공 경로는 `return nil`입니다. nil `*InputError`를 반환하면 동적 타입이 있는 error 인터페이스가 만들어집니다. 호출자마다 typed nil을 검사하게 만드는 대신 오류 생산자의 반환을 고칩니다.

### D. 획득 성공 확인 뒤 defer

일반적인 순서는 다음과 같습니다.

```go
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()
```

이 발췌는 등록 순서만 보여 줍니다. 실제 코드에서는 Close 오류를 어떻게 처리할지도 계약에 맞춰 결정해야 합니다. 읽기 자원인지 쓰기 자원인지, 성공 반환 전에 정리 실패를 알려야 하는지에 따라 구체 코드가 달라집니다.

## 문제 4. 오류 타입의 데이터와 원인

맵에서 읽은 값이 0인지가 아니라 존재 여부 ok를 확인합니다. LookupError에는 조회한 키를 보관하고 Unwrap으로 대표 원인을 노출합니다.

참고 구현: [lookup/main.go](lookup/main.go).

```go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type LookupError struct {
	Key   string
	Cause error
}

func (e *LookupError) Error() string { return fmt.Sprintf("lookup %q: %v", e.Key, e.Cause) }
func (e *LookupError) Unwrap() error { return e.Cause }
func lookup(values map[string]int, key string) (int, error) {
	value, ok := values[key]
	if !ok {
		return 0, &LookupError{Key: key, Cause: ErrNotFound}
	}
	return value, nil
}
func main() {
	values := map[string]int{"zero": 0}
	value, err := lookup(values, "zero")
	fmt.Println("존재:", value, err == nil)
	_, err = lookup(values, "missing")
	fmt.Println("없음:", errors.Is(err, ErrNotFound))
	var detail *LookupError
	if errors.As(err, &detail) {
		fmt.Println("키:", detail.Key)
	}
}
```

```sh
go run ./solutions/11-errors-defer/lookup
go test ./solutions/11-errors-defer/lookup
```

예상 출력:

```text
존재: 0 true
없음: true
키: missing
```


`lookup(nil, "missing")`도 정상적으로 같은 조회 실패를 반환합니다. nil 맵 조회는 안전하며 쓰기를 수행하지 않기 때문입니다.

바깥에서 `fmt.Errorf("service: %w", err)`로 감싼 뒤에도 errors.Is와 errors.As는 내부 관계를 탐색합니다. 단순 `err.(*LookupError)`는 바깥 타입만 보므로 그 경우 실패할 수 있습니다.

LookupError의 Key는 표시 문자열을 잘라서 복구한 것이 아닙니다. 메시지 표현을 바꾸어도 이 구조화된 필드의 계약을 유지할 수 있습니다. 반대로 필드 이름과 타입을 바꾸는 것은 공개 오류 API 변경이 될 수 있습니다.

## 문제 5. 소유권과 두 오류 보존

함수는 r을 받은 시점부터 정리 책임을 갖습니다. operation을 호출하기 전에 defer를 등록하므로 정상 반환 또는 오류 반환 모두 Close에 도달합니다.

참고 구현: [cleanup/main.go](cleanup/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"io"
)

// withResource takes ownership of r; operation receives the same resource.
func withResource(r io.Closer, operation func(io.Closer) error) (err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close: %w", closeErr))
		}
	}()
	if callErr := operation(r); callErr != nil {
		return fmt.Errorf("operation: %w", callErr)
	}
	return nil
}

type resource struct {
	closeErr error
	closes   int
}

func (r *resource) Close() error { r.closes++; return r.closeErr }
func main() {
	workErr := errors.New("work failed")
	closeErr := errors.New("close failed")
	r := &resource{closeErr: closeErr}
	err := withResource(r, func(io.Closer) error { return workErr })
	fmt.Println("처리 오류:", errors.Is(err, workErr))
	fmt.Println("정리 오류:", errors.Is(err, closeErr))
	fmt.Println("닫기 횟수:", r.closes)
}
```

```sh
go run ./solutions/11-errors-defer/cleanup
go test ./solutions/11-errors-defer/cleanup
```

예상 출력:

```text
처리 오류: true
정리 오류: true
닫기 횟수: 1
```


닫기 횟수는 1입니다. main은 withResource가 끝난 뒤 다시 Close하지 않습니다. 소유권을 넘긴다는 문장이 실제 호출 구조와 일치해야 합니다.

오류 결과를 네 가지 조합으로 비교합니다.

| operation 오류 | Close 오류 | 최종 err | Is(A) | Is(B) |
| --- | --- | --- | --- | --- |
| nil | nil | nil | false | false |
| A | nil | operation으로 감싼 A | true | false |
| nil | B | close로 감싼 B | false | true |
| A | B | 두 오류 결합 | true | true |

이름 붙인 err에 오류가 담겨 있는 상태로 반환을 시작해도, defer가 끝나기 전에는 호출자에게 확정된 결과를 전달하지 않습니다. Close 오류가 있다면 그때 Join합니다.

### 대안: 첫 오류 우선 정책

프로그램에 따라 첫 처리 오류만 반환하고 정리 오류는 별도 기록하는 정책을 선택할 수 있습니다. 그것이 항상 잘못인 것은 아닙니다. 다만 이번 문제는 두 원인을 모두 보존하라는 계약이 있으므로 첫 오류 하나만 유지하는 풀이는 완료 조건을 만족하지 않습니다.

문자열을 `err.Error()+closeErr.Error()`처럼 이어 붙이면 사람은 읽을 수 있어도 errors.Is로 두 원인을 찾을 수 없습니다. 오류 구조를 유지하는 Join이 요구사항에 맞습니다.

### 적용 범위

r과 operation이 nil이 아니며 operation은 반환한다고 가정했습니다. 모든 typed nil을 추적하거나 임의의 panic을 성공으로 바꾸는 기능은 없습니다. 더 넓은 입력 계약이 필요하면 그 요구에 맞는 검증과 테스트를 추가해야 합니다.

## 문제 6. 독립된 검증 오류 모으기

이름과 나이는 서로 독립적으로 검사할 수 있습니다. 첫 실패에서 반환하지 않고 각 실패를 슬라이스에 추가한 뒤 Join합니다.

참고 구현: [validation/main.go](validation/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrName = errors.New("name required")
var ErrAge = errors.New("age must be 0..150")

func validate(name string, age int) error {
	var all []error
	if strings.TrimSpace(name) == "" {
		all = append(all, ErrName)
	}
	if age < 0 || age > 150 {
		all = append(all, ErrAge)
	}
	return errors.Join(all...)
}
func main() {
	err := validate(" ", 151)
	fmt.Println("이름:", errors.Is(err, ErrName))
	fmt.Println("나이:", errors.Is(err, ErrAge))
	fmt.Println("정상:", validate("Go", 0) == nil)
}
```

```sh
go run ./solutions/11-errors-defer/validation
go test ./solutions/11-errors-defer/validation
```

예상 출력:

```text
이름: true
나이: true
정상: true
```


정상 입력의 오류 슬라이스는 비어 있고 Join 결과는 nil입니다. `errors.New("")`를 반환하면 빈 메시지여도 error 인터페이스는 nil이 아니므로 성공이 아닙니다.

`strings.TrimSpace`는 검증용으로 사용했습니다. name은 문자열 값이며 호출자의 변수를 바꾸지 않습니다. 이 함수는 정규화된 이름을 반환하는 기능이 아니라 유효성만 확인하는 기능입니다.

## 선택 A. 복구와 종료

1. 숫자가 아닌 포트는 예상 가능한 입력 실패입니다. error로 반환하면 호출자가 메시지를 표시하거나 다시 입력받을 수 있습니다.
2. 복구는 실패한 문장 다음부터 다시 시작하지 않습니다. 복구한 함수가 반환하고 호출자가 이어갑니다.
3. recover의 범위는 같은 고루틴의 panic 전파 과정입니다. 다른 고루틴은 별도 실행 흐름입니다.
4. 알 수 없는 panic을 nil로 바꾸면 내부 버그를 성공처럼 숨기고 잘못된 상태를 계속 사용할 수 있습니다.
5. os.Exit는 defer를 실행하지 않고 종료합니다. 자원 처리를 마친 가장 바깥 경계에서 종료 정책을 적용합니다.

## 본문 자기 점검 답

| 질문 | 핵심 답 |
| --- | --- |
| 정상 0과 실패 0 | error의 nil 여부와 함수 계약 |
| 같은 문장의 New 두 번 | 별개의 오류 값 |
| %v와 %w | 원인 래핑 관계 보존 여부 |
| As의 대상 주소 | 찾은 오류 값을 대상 변수에 대입 |
| defer 평가 시점 | 인자는 등록 시점, 본문은 종료 시점 |
| 반환 변수와 defer | 반환식 대입 후 defer가 최종 값을 수정 가능 |
| 처리·정리 동시 실패 | 정한 계약에 따라 두 원인 보존 또는 명시적 우선순위 |
| typed nil | 동적 타입이 있어서 nil 인터페이스 아님 |
| 다른 고루틴 recover | 불가 |
| 중간 os.Exit | defer 정리가 실행되지 않음 |

## 자동 검증

```sh
go test ./examples/11-errors-defer/... ./solutions/11-errors-defer/...
go vet ./examples/11-errors-defer/... ./solutions/11-errors-defer/...
```

테스트는 정상·경계 입력, 원인 유지, 구조화된 세부 정보, 정리 횟수, 처리·정리 오류 조합을 확인합니다. panic 예제의 테스트는 알려진 타입만 변환하고 모르는 panic 값은 원래 값 그대로 전파되는지 검사합니다.

단순한 오류 메시지 문구 변경에만 민감한 테스트와, 약속한 오류 종류·필드·상태를 검사하는 테스트를 구별하세요. 다음 장에서는 이런 계약을 표로 정리하고 자동 테스트로 표현합니다.

[본문](../../book/11-errors-defer/README.md) · [문제로 돌아가기](../../exercises/11-errors-defer/README.md)
