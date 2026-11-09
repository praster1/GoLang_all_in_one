# 10장 연습문제: 패키지와 모듈

[본문](../../book/10-packages-modules/README.md)을 읽고 풉니다. 이 장에서는 결과 숫자뿐 아니라 파일의 위치, import 경로와 공개 범위도 요구사항에 포함됩니다.

모든 명령은 저장소 루트에서 실행합니다. 개인 프로그램은 `_practice/10-packages-modules/` 아래 별도 폴더에 작성합니다. 기존 파일은 덮어쓰지 않습니다. 새 go.mod나 go.work는 만들지 않습니다.

## 준비: 개인 파일과 import 경로

다음 폴더를 편집기로 준비하세요.

| 문제 | 개인 폴더 | 포함할 파일 |
| --- | --- | --- |
| 2 | `_practice/10-packages-modules/split/` | main.go, mean.go |
| 4 | `_practice/10-packages-modules/labels/` | main.go, label/label.go |
| 5 | `_practice/10-packages-modules/balance/` | main.go, wallet/wallet.go |
| 6 | `_practice/10-packages-modules/consumer/` | main.go, task/task.go, display/display.go |

자신의 label 패키지를 import할 경로는 다음과 같습니다.

```text
github.com/praster1/GoLang_all_in_one/_practice/10-packages-modules/labels/label
```

본문의 `examples/...` 경로를 그대로 import하면 개인 풀이 대신 이미 완성된 예제를 사용하게 됩니다. 경로 접두사와 자신의 폴더 위치를 함께 확인하세요.

`_practice`로 시작하는 폴더는 공통 `./...` 검사에서 건너뛰므로 직접 경로를 지정합니다.

```sh
go run ./_practice/10-packages-modules/labels
go test ./_practice/10-packages-modules/labels/...
```

테스트를 아직 작성하지 않았다면 첫 명령으로 정상·경계 입력을 실행합니다. 테스트 작성법은 12장에서 배우며, 이 장의 참고 구현 테스트는 실행해 비교할 수 있습니다.

## 문제 1. 이름과 경계 판별

**수준: 확인**

가상의 모듈 이름이 `example.com/learn`이고 아래 두 파일이 있다고 가정합니다. 실제 네트워크 경로가 아니라 구조를 읽는 문제입니다.

`stats/sum.go`:

```go
package stats

func Total(values []int) int {
    total := 0
    for _, value := range values {
        total += value
    }
    return total
}

func label() string { return "합계" }
```

`cmd/report/main.go`:

```go
package main

import (
    "fmt"
    "example.com/learn/stats"
)

func main() {
    fmt.Println(stats.Total([]int{1, 2, 3}))
}
```

다음 표를 채우세요.

| 질문 | 답 |
| --- | --- |
| 모듈은 몇 개인가요? | |
| 일반 소스 패키지는 몇 개인가요? | |
| stats의 import 경로는 무엇인가요? | |
| main이 stats.label을 호출할 수 있나요? | |
| stats의 다른 파일이 label을 호출할 수 있나요? | |
| sum.go의 파일 이름을 total.go로 바꾸면 import도 바뀌나요? | |
| stats 폴더를 measures로 옮기면 무엇을 확인해야 하나요? | |

완료 조건: 파일 이름과 디렉터리 경로의 변경을 구별하고, 모듈을 새로 만들지 않아도 패키지를 추가할 수 있음을 설명합니다.

<details>
<summary>힌트</summary>

import 경로에는 파일 이름이 아니라 패키지 디렉터리의 위치가 들어갑니다. 소문자 함수는 다른 파일이 아니라 다른 패키지에서 감춰집니다.

</details>

## 문제 2. 평균 함수를 별도 파일로

**수준: 변형**

개인 `split/main.go`와 `split/mean.go`를 작성합니다. 둘 다 `package main`입니다. 계산 함수는 mean.go에, 출력은 main.go에 둡니다.

```go
func mean(scores []int) (float64, bool)
```

계약:

- 점수는 각 0 이상 100 이하입니다.
- 점수 개수는 1 이상 100 이하입니다.
- 조건을 만족하지 않으면 `0, false`입니다.
- 성공하면 소수점을 포함한 평균과 true를 반환합니다.
- 입력 슬라이스를 변경하지 않습니다.

| 입력 | 결과 | 성공 |
| --- | ---: | --- |
| `{10, 20, 30}` | 20 | true |
| `{0}` | 0 | true |
| `{100}` | 100 | true |
| `{1, 2}` | 1.5 | true |
| nil | 0 | false |
| `{-1}` | 0 | false |
| `{101}` | 0 | false |
| 101개 점수 | 0 | false |

완료 조건: 폴더를 대상으로 실행하면 성공합니다. main.go만 직접 실행하는 명령은 왜 실패하는지 설명합니다. 평균 계산에서 정수 나눗셈을 먼저 하면 왜 안 되는지도 확인하세요.

<details>
<summary>힌트 1</summary>

두 파일은 같은 패키지이므로 서로 import하지 않습니다. 함수가 다른 파일에 있다는 이유로 공개 대문자 이름으로 바꿀 필요도 없습니다.

</details>

<details>
<summary>힌트 2</summary>

점수와 개수 범위를 먼저 제한하면 합계가 10,000을 넘지 않습니다. 합계를 float64로 바꾼 뒤 나눕니다.

</details>

## 문제 3. 세 가지 import 문제 진단

**수준: 변형**

다음은 서로 독립된 상황입니다. 실제 정상 예제 파일에 오류를 추가하지 말고, 어떤 규칙에 어긋나는지와 수정 방법을 적습니다.

### 상황 A. 다른 파일의 import 기대

같은 패키지의 `main.go`에 `import "fmt"`가 있습니다. 다른 `format.go`에서는 import 없이 `fmt.Sprintf`를 사용했습니다.

- 왜 함수 선언은 서로 보이는데 fmt는 보이지 않나요?
- 필요한 최소 수정은 무엇인가요?

### 상황 B. 비공개 필드 직접 수정

```go
var c counter.Counter
c.value = 200
```

- Counter 타입이 공개되어 있어도 왜 value는 접근할 수 없나요?
- 이 예제의 한도 계약을 유지하면서 값을 늘리는 방법은 무엇인가요?

### 상황 C. internal 경계

`examples/10-packages-modules/internal-demo/internal/token` 패키지를 `_practice/10-packages-modules/labels`에서 import하려 합니다.

- 같은 모듈인데도 왜 허용되지 않나요?
- 접근 경계를 결정하는 internal의 부모는 어느 폴더인가요?

완료 조건: 문제를 모두 “경로 오류”라고 뭉뚱그리지 않고 파일 스코프·공개 이름·internal 규칙으로 나눠 설명합니다.

<details>
<summary>힌트</summary>

import의 스코프와 패키지 선언의 스코프는 다릅니다. internal 규칙은 모듈 전체가 아니라 internal 바로 위 부모의 트리를 기준으로 판단합니다.

</details>

## 문제 4. 라벨 처리 패키지

**수준: 응용**

개인 `labels/label/label.go`에 `package label`을 만들고 다음 함수를 공개합니다.

```go
func Normalize(value string) (string, bool)
```

계약:

1. 앞뒤 공백을 제거합니다.
2. 결과가 유효한 UTF-8이며 비어 있지 않아야 합니다.
3. 결과의 룬 수가 20 이하여야 합니다.
4. 결과 안에 CR 또는 LF가 남으면 거부합니다.
5. 성공한 문자열을 `strings.ToLower`로 소문자화하여 반환합니다.
6. 실패하면 빈 문자열과 false입니다.

이 규칙은 일반적인 모든 라벨 시스템의 표준이 아니라 이 문제의 사용 계약입니다. 사용자 화면의 글자 수와 룬 수가 항상 같지는 않다는 07장의 설명을 기억하세요.

main 패키지는 함수 결과만 받아 출력합니다. label 패키지에서 fmt.Print 계열로 직접 출력하지 않습니다.

| 입력 | 기대 |
| --- | --- |
| `" GO "` | `"go", true` |
| `" 한글 "` | `"한글", true` |
| 공백뿐인 문자열 | 실패 |
| `"a\nb"` | 실패 |
| 잘못된 UTF-8 바이트 | 실패 |
| 가 20개 | 성공 |
| 가 21개 | 실패 |

완료 조건: main에서 올바른 개인 패키지 경로를 import합니다. label 함수의 출력 정책과 검증 규칙을 함수 주석에 적습니다.

<details>
<summary>힌트 1</summary>

문자열 패키지와 unicode/utf8 패키지는 label.go가 직접 import합니다. main의 import를 대신 사용할 수 없습니다.

</details>

<details>
<summary>힌트 2</summary>

줄바꿈은 앞뒤 공백 제거 뒤 검사합니다. 따라서 마지막 줄바꿈 하나를 제거한 `"한글\n"`은 성공하고, 내부에 남은 `"한\n글"`은 실패합니다.

</details>

## 문제 5. 잔액을 보호하는 지갑

**수준: 응용**

`wallet` 패키지에 공개 타입 `Wallet`과 비공개 `balance int`를 만듭니다.

```go
func New(initial int) (Wallet, bool)
func (wallet Wallet) Balance() int
func (wallet *Wallet) Spend(amount int) bool
```

계약:

- 초기 잔액은 0 이상 10,000 이하입니다. 잘못된 입력이면 영값 Wallet과 false입니다.
- 영값 Wallet은 잔액 0으로 읽을 수 있습니다.
- 지출은 0 이상이며 현재 잔액 이하여야 합니다.
- 지출 실패는 잔액을 바꾸지 않습니다.
- nil 수신자의 Spend는 false입니다.
- nil 포인터에서 Balance를 호출하는 것은 지원하지 않습니다.
- 외부 패키지가 필드에 직접 대입할 수 없어야 합니다.

| 초기 잔액 | 지출 | 성공 | 남은 잔액 |
| ---: | ---: | --- | ---: |
| 3,000 | 1,000 | true | 2,000 |
| 2,000 | 3,000 | false | 2,000 |
| 100 | 100 | true | 0 |
| 0 | 0 | true | 0 |
| 100 | -1 | false | 100 |

완료 조건: 상태 변경을 패키지 안의 메서드로 제한합니다. New 함수가 있다는 사실과 영값을 만들 수 있다는 사실을 함께 설명하세요.

<details>
<summary>힌트</summary>

지출은 잔액과 비교한 뒤 빼면 됩니다. 잘못된 결과를 계산한 다음 되돌리는 대신, 입력 검증이 끝나기 전에는 대입하지 않습니다.

</details>

## 문제 6. 소비 측에서 필요한 행동 정의

**수준: 응용**

개인 `consumer` 폴더 아래 `task`, `display` 패키지를 만듭니다.

- task 패키지의 `Task`는 `Title string`, `Done bool`을 가집니다.
- `Task.Description()`은 제목 뒤에 ` (진행)` 또는 ` (완료)`를 붙입니다.
- display 패키지는 `Description() string` 메서드만 가진 인터페이스 `Describer`를 선언합니다.
- `display.Line`은 받은 설명 앞에 `- `를 붙여 반환합니다.
- task는 display를 import하지 않습니다.
- display는 task를 import하지 않습니다.
- main이 두 패키지를 import하여 연결하고 출력합니다.

예상 출력:

```text
- 패키지 복습 (완료)
```

완료 조건: task를 모르는 별도 타입도 Description 메서드를 제공하면 display.Line에 전달할 수 있습니다. 폴더 이름이나 구조체 필드 이름이 일치해서 구현되는 것이 아님을 설명합니다.

<details>
<summary>힌트 1</summary>

인터페이스는 동작을 소비하는 display에 선언합니다. task 타입에 `implements` 같은 표시를 추가하지 않습니다.

</details>

<details>
<summary>힌트 2</summary>

Line의 결과를 만드는 데 필요한 정보는 Description의 반환 문자열뿐입니다. 구체 타입 단언이나 Task의 필드 접근이 필요하지 않습니다.

</details>

## 선택 A. 모듈 그래프 읽기

**수준: 확인**

다음은 가상의 의존 관계입니다.

| 요구하는 모듈 | 필요한 모듈 | 요구 버전 |
| --- | --- | --- |
| app | helper | v1.2.0 |
| app | report | v1.0.0 |
| report v1.0.0 | helper | v1.4.0 |

질문:

1. 이 단순한 그래프에서 선택할 helper 버전은 무엇인가요?
2. 네트워크에 helper v1.9.0이 있다고 빌드마다 자동 선택되나요?
3. go.sum에 v1.2.0과 v1.4.0 체크섬이 함께 있으면 둘을 동시에 사용하는 뜻인가요?
4. 체크섬 검증 성공은 API 호환성 검증과 같은가요?
5. 같은 기능을 v2로 바꾼 모듈이 왜 import 경로를 바꿔야 할 수 있나요?

완료 조건: 버전 요구·최종 선택·내용 검증·동작 테스트를 구별합니다. 이 가상 주소를 실제 go get 명령으로 실행할 필요는 없습니다.

## 완료 체크리스트

- [ ] 여러 파일을 디렉터리 단위로 실행했습니다.
- [ ] 개인 패키지의 import 경로를 직접 계산했습니다.
- [ ] 공개 함수와 비공개 상태를 구별했습니다.
- [ ] internal 경계와 소문자 이름의 경계가 다름을 설명했습니다.
- [ ] 두 패키지의 import를 순환 없이 구성했습니다.
- [ ] go.mod와 go.sum의 역할을 구별했습니다.

[본문](../../book/10-packages-modules/README.md) · [해설](../../solutions/10-packages-modules/README.md)
