# 10장 해설: 패키지와 모듈

먼저 [문제](../../exercises/10-packages-modules/README.md)를 풀어 보세요. 코드가 같은 결과를 내더라도 패키지 경계 조건을 어겼다면 그 부분까지 수정해야 합니다. 모든 명령은 저장소 루트에서 실행합니다.

## 문제 1. 경계 판별

| 질문 | 답 |
| --- | --- |
| 모듈 개수 | 하나. example.com/learn |
| 일반 소스 패키지 개수 | 두 개. stats, cmd/report |
| stats의 import 경로 | example.com/learn/stats |
| main에서 stats.label 호출 | 불가. 소문자 비공개 이름 |
| stats의 다른 파일에서 label 호출 | 가능. 같은 패키지 |
| sum.go를 total.go로 이름 변경 | import 경로는 유지 |
| stats 폴더를 measures로 이동 | import 경로 변경, 선언 이름·사용 이름·테스트 확인 |

폴더를 바꿔도 package 선언이 자동으로 바뀌지는 않습니다. 관례에 맞춰 함께 변경할지 결정해야 합니다. 파일 이동 도구가 일부 import를 고쳐 주더라도 실제 빌드와 테스트로 결과를 확인합니다.

## 문제 2. 계산과 출력 파일 분리

두 파일 모두 같은 main 패키지입니다. main은 mean을 직접 호출하고, mean.go는 fmt를 사용하지 않으므로 fmt import도 없습니다.

파일: [`split/main.go`](split/main.go).

```go
package main

import "fmt"

func main() {
	average, ok := mean([]int{10, 20, 30})
	fmt.Println("평균:", average, ok)
	average, ok = mean(nil)
	fmt.Println("빈 입력:", average, ok)
}
```

파일: [`split/mean.go`](split/mean.go).

```go
package main

// mean accepts at most 100 scores, each in 0..100.
func mean(scores []int) (float64, bool) {
	if len(scores) == 0 || len(scores) > 100 {
		return 0, false
	}
	total := 0
	for _, score := range scores {
		if score < 0 || score > 100 {
			return 0, false
		}
		total += score
	}
	return float64(total) / float64(len(scores)), true
}
```

저장소 루트에서 실행합니다.

```sh
go run ./solutions/10-packages-modules/split
go test ./solutions/10-packages-modules/split/...
```

예상 출력:

```text
평균: 20 true
빈 입력: 0 false
```


점수 개수와 각 점수의 범위를 제한하면 합계는 최대 10,000입니다. 평균 계산 전에 두 정수 중 적어도 나누는 표현식의 피연산자를 float64로 바꿉니다. 참고 구현은 양쪽을 명시해 읽기 쉽게 작성했습니다.

다음은 개인 파일을 실행할 때 흔한 실수입니다.

```sh
go run ./_practice/10-packages-modules/split/main.go
```

이 명령은 mean.go를 포함하지 않습니다. 폴더 경로를 대상으로 실행하세요. 파일을 분리한 뒤 main.go만 본 편집기 실행 버튼도 같은 문제를 만들 수 있으므로 실행 구성을 확인합니다.

## 문제 3. 서로 다른 세 가지 접근 규칙

### 상황 A. import는 파일 스코프

format.go에 `import "fmt"`를 추가합니다. 같은 패키지의 함수와 변수는 다른 파일에서도 이름을 찾지만, import에 붙은 패키지 식별자는 해당 파일에서만 사용합니다.

### 상황 B. 공개 타입과 필드는 별개

counter.Counter가 공개라고 해서 value 필드도 공개가 되는 것은 아닙니다. c.Add를 사용해 허용된 변경을 요청하고 bool 결과를 확인합니다. 이 예제에서 200이라는 상태는 계약을 벗어나므로 외부에서 억지로 설정할 경로가 없어야 합니다.

### 상황 C. internal 부모 트리

허용 범위는 `examples/10-packages-modules/internal-demo/` 아래입니다. `_practice/10-packages-modules/labels/`는 같은 저장소라도 그 트리 밖에 있으므로 접근할 수 없습니다.

해결은 필요한 기능을 적절한 공개 API로 제공하거나, 실제로 내부 프로그램인 경우 허용된 위치에 배치하는 것입니다. 오류를 없애려고 접근 경계를 무조건 삭제하는 것은 설계 의도를 바꿉니다.

## 문제 4. 출력하지 않는 라벨 패키지

label 패키지는 입력을 정리하고 유효성을 확인해 반환합니다. main이 출력하므로 같은 함수는 CLI, HTTP, 테스트 등에서 재사용할 수 있습니다.

파일: [`labels/label/label.go`](labels/label/label.go).

```go
// Package label creates small canonical labels.
package label

import (
	"strings"
	"unicode/utf8"
)

func Normalize(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || value == "" || utf8.RuneCountInString(value) > 20 || strings.ContainsAny(value, "\r\n") {
		return "", false
	}
	return strings.ToLower(value), true
}
```

파일: [`labels/main.go`](labels/main.go).

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/labels/label"
)

func main() {
	value, ok := label.Normalize(" GO ")
	fmt.Println("라벨:", value, ok)
	_, ok = label.Normalize("  ")
	fmt.Println("빈 값:", ok)
}
```

저장소 루트에서 실행합니다.

```sh
go run ./solutions/10-packages-modules/labels
go test ./solutions/10-packages-modules/labels/...
```

예상 출력:

```text
라벨: go true
빈 값: false
```


`TrimSpace`를 먼저 호출하므로 맨 앞뒤의 줄바꿈은 공백처럼 제거됩니다. 안에 남은 CR/LF는 한 줄 규칙에 어긋납니다. 이 순서를 바꾸면 계약도 바뀝니다.

룬 수를 검사하기 전에 UTF-8 유효성을 검사합니다. 잘못된 바이트를 대체 문자 하나처럼 세서 정상 라벨로 받아들이는 것을 막습니다. `strings.ToLower`는 유니코드 소문자 매핑을 수행하며 모든 언어의 정교한 검색 정규화 정책을 대신하지는 않습니다.

참고 테스트는 정상 문자열, 공백, 내부 줄바꿈, 잘못된 UTF-8, 20·21 룬의 경계를 확인합니다. 패키지 안에서 출력하지 않기 때문에 반환값을 바로 비교할 수 있습니다.

## 문제 5. 공개 API로 잔액 변경 제한

Wallet의 필드는 비공개입니다. New가 초기 입력을 검사하고 Spend가 변경을 검사합니다. 영값은 잔액 0으로 유효하므로 `var w wallet.Wallet`도 사용할 수 있습니다.

파일: [`balance/main.go`](balance/main.go).

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/balance/wallet"
)

func main() {
	w, ok := wallet.New(3000)
	if !ok {
		fmt.Println("입력 오류")
		return
	}
	fmt.Println("사용:", w.Spend(1000), w.Balance())
	fmt.Println("초과:", w.Spend(3000), w.Balance())
}
```

파일: [`balance/wallet/wallet.go`](balance/wallet/wallet.go).

```go
// Package wallet provides a small integer balance with explicit limits.
package wallet

type Wallet struct{ balance int }

func New(initial int) (Wallet, bool) {
	if initial < 0 || initial > 10000 {
		return Wallet{}, false
	}
	return Wallet{balance: initial}, true
}
func (w Wallet) Balance() int { return w.balance }
func (w *Wallet) Spend(amount int) bool {
	if w == nil || amount < 0 || amount > w.balance {
		return false
	}
	w.balance -= amount
	return true
}
```

저장소 루트에서 실행합니다.

```sh
go run ./solutions/10-packages-modules/balance
go test ./solutions/10-packages-modules/balance/...
```

예상 출력:

```text
사용: true 2000
초과: false 2000
```


Spend는 amount가 잔액 이하인지 먼저 확인하므로 감소 후 음수가 되는 상황을 허용하지 않습니다. amount가 음수이면 빼기 연산이 사실상 잔액 증가가 될 수 있으므로 음수 검사도 필요합니다.

New는 예약어가 아닙니다. 이름을 잘 지은 일반 함수입니다. 생성자를 제공하면서 영값의 동작을 설명하지 않으면 사용자는 `Wallet{}`이 가능한지, 허용되는지 혼동하기 쉽습니다.

복사한 Wallet은 별도 int 상태를 갖습니다. 공유 지갑이나 동시 접근은 이 장에서 다루지 않습니다. 실제 금융 시스템의 원장·정확한 금액 단위·동시성·감사 기록을 이 작은 학습 타입으로 대체할 수 있다는 의미도 아닙니다.

## 문제 6. 인터페이스를 소비하는 위치에 선언

Task는 Description 메서드를 제공합니다. display는 Describer의 계약만 요구합니다. main이 구체 값과 표시 함수를 연결합니다.

파일: [`consumer/display/display.go`](consumer/display/display.go).

```go
package display

type Describer interface{ Description() string }

func Line(value Describer) string { return "- " + value.Description() }
```

파일: [`consumer/main.go`](consumer/main.go).

```go
package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/consumer/display"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/consumer/task"
)

func main() { fmt.Println(display.Line(task.Task{Title: "패키지 복습", Done: true})) }
```

파일: [`consumer/task/task.go`](consumer/task/task.go).

```go
package task

type Task struct {
	Title string
	Done  bool
}

func (t Task) Description() string {
	if t.Done {
		return t.Title + " (완료)"
	}
	return t.Title + " (진행)"
}
```

저장소 루트에서 실행합니다.

```sh
go run ./solutions/10-packages-modules/consumer
go test ./solutions/10-packages-modules/consumer/...
```

예상 출력:

```text
- 패키지 복습 (완료)
```


Task와 display가 서로를 import하지 않으므로 순환이 없습니다. display 테스트의 별도 문자열 타입도 Description을 구현하여 같은 함수에 전달됩니다. 구체 타입이 무엇인지 검사하지 않고 약속한 행동만 사용합니다.

이 분리가 필요한지는 실제 프로그램의 규모에 달려 있습니다. 한 번만 쓰는 한 줄 함수를 위해 무조건 여러 패키지를 만드는 것이 목표는 아닙니다. 이번 문제는 인터페이스의 구현 관계가 import 관계와 다를 수 있음을 연습하기 위해 구조를 나눴습니다.

## 선택 A. 요구 버전과 선택 버전

1. 주어진 단순 그래프에서는 helper v1.4.0이 선택됩니다. v1.2.0과 v1.4.0 요구를 함께 만족해야 합니다.
2. 외부에 v1.9.0이 있다는 이유만으로 매번 자동 선택하지 않습니다.
3. go.sum의 여러 버전 체크섬은 현재 동시에 사용한다는 뜻이 아닙니다.
4. 내용 무결성 검사는 공개 API 호환성과 프로그램 동작 테스트를 대신하지 않습니다.
5. v2 이상은 보통 모듈 경로에 주요 버전 접미사를 넣습니다. 새로운 계약을 기존 v1과 구분하기 위한 경로 규칙입니다.

실제 그래프에 replace·exclude·작업 공간 등이 있다면 추가 규칙을 확인해야 합니다. 이 문제는 그 예외 없이 기본 원리를 연습했습니다.

## 본문 자기 점검 답

1. 현재 모듈 안의 import는 로컬 소스 경로로 해결할 수 있습니다.
2. 패키지 수준 선언과 import 식별자의 스코프가 다르기 때문입니다.
3. 같은 패키지의 다른 일반 소스 파일에서는 비공개 필드도 접근할 수 있습니다.
4. internal은 패키지 import 허용 트리를, 소문자 이름은 패키지 안의 선언 접근을 제한합니다.
5. report는 필요한 메서드의 인터페이스만 사용하며 Basket이 이를 암시적으로 구현합니다.
6. 최종 선택은 go.mod와 모듈 그래프를 함께 봐야 합니다.
7. tidy는 의존 기록을 소스·테스트에 맞게 정리할 수 있고, verify는 로컬 모듈 캐시 내용의 변경을 검사합니다.
8. 단일 모듈은 여러 패키지를 가질 수 있으므로 go.work가 필수는 아닙니다.

## 검증 명령

```sh
go fmt ./examples/10-packages-modules/... ./solutions/10-packages-modules/...
go vet ./examples/10-packages-modules/... ./solutions/10-packages-modules/...
go test ./examples/10-packages-modules/... ./solutions/10-packages-modules/...
go build ./examples/10-packages-modules/... ./solutions/10-packages-modules/...
```

각 main의 예상 출력은 예제 테스트로 확인합니다. 패키지별 검증은 평균의 입력 경계, 라벨의 문자 경계, 잔액의 실패 시 상태 보존, 소비 측 인터페이스에 다른 구현을 전달하는 동작을 포함합니다.

코드를 개인 폴더로 옮긴 경우 import 경로를 자신의 폴더로 바꿔야 합니다. 해설 코드를 그대로 복사하고 solutions 경로를 import하면 개인 구현을 우회할 수 있습니다.

[본문](../../book/10-packages-modules/README.md) · [문제로 돌아가기](../../exercises/10-packages-modules/README.md)
