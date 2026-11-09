# 17장 연습문제 — Context

[본문](../../book/17-context/README.md)의 취소·마감·값 전달을 연습합니다.
기본 6문제와 선택 1문제입니다.
[해설](../../solutions/17-context/README.md)은 구현한 뒤 확인합니다.

기본 문제는 직접 고루틴을 만들지 않아도 해결할 수 있습니다.
취소된 입력, 지난 마감, 동기적인 콜백을 이용해 상태를 재현합니다.
테스트 성공을 `time.Sleep`에 의존시키지 않습니다.

## 준비

모든 명령은 저장소 루트에서 실행합니다.
기존 예제와 해설 파일은 직접 수정하지 않습니다.

Bash:

```sh
mkdir -p _practice/17-context/checked-sum
if [ ! -e _practice/17-context/checked-sum/main.go ]; then
  cp examples/17-context/checkpoints/main.go _practice/17-context/checked-sum/main.go
fi
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/17-context/checked-sum | Out-Null
if (-not (Test-Path _practice/17-context/checked-sum/main.go)) {
    Copy-Item examples/17-context/checkpoints/main.go _practice/17-context/checked-sum/main.go
}
```

기존 개인 파일은 보존됩니다.
문제 4·5·6은 별도 이름의 개인 폴더에 작성합니다.
모든 Context 매개변수에는 nil이 아닌 값을 전달합니다.

## 문제 1. 상태와 취소 방향 예측

**수준: 확인**

다음 상태표를 채웁니다.
parent는 Background에서 WithCancel로 만들고,
child와 sibling은 parent에서 각각 WithCancel로 만듭니다.

| 사건 | parent.Err | child.Err | sibling.Err |
| --- | --- | --- | --- |
| 생성 직후 | 직접 작성 | 직접 작성 | 직접 작성 |
| child만 cancel | 직접 작성 | 직접 작성 | 직접 작성 |
| parent cancel | 직접 작성 | 직접 작성 | 직접 작성 |
| child를 다시 cancel | 직접 작성 | 직접 작성 | 직접 작성 |

다음 질문에도 답합니다.

1. child의 Done을 관찰했으면 child에서 하던 모든 계산이 끝났다고 말할 수 있는가?
2. Background의 Done에 단독으로 수신하면 왜 진행하지 못하는가?
3. parent의 마감보다 늦은 child 마감을 요청하면 어떤 시간이 적용되는가?
4. WithTimeout으로 만든 Context를 수동 취소하면 항상 DeadlineExceeded인가?
5. cancel(nil)을 호출하는 CancelCauseFunc는 취소하지 않는다는 뜻인가?

### 완료 조건

취소 신호와 작업 완료를 구분합니다.
상태 오류와 일반 작업 오류를 구분합니다.
취소 방향을 부모·형제까지 무조건 전파되는 것으로 설명하지 않습니다.

### 힌트

[Context](https://pkg.go.dev/context#Context)와
[CancelFunc](https://pkg.go.dev/context#CancelFunc)를 읽습니다.
Done은 Context 상태에 관한 채널이며 작업 결과 채널이 아닙니다.

## 문제 2. 취소 가능한 합계

**수준: 변형**

개인 파일: `_practice/17-context/checked-sum/main.go`

```go
func sum(ctx context.Context, values []int) (int, error)
```

### 계약

| 항목 | 규칙 |
| --- | --- |
| 값 개수 | 최대 1000개 |
| 각 값 | 0~1000 |
| 빈 슬라이스 | 0, nil |
| 취소 | 0과 ctx.Err |
| 입력 범위 오류 | 0과 일반 오류 |
| 성공 | 전체 합계와 nil |
| 확인 위치 | 입구, 각 반복 전, 전체 반복 뒤 |

이미 취소된 Context는 입력 검증보다 먼저 확인합니다.
따라서 취소된 Context와 잘못된 값이 함께 오면 취소 오류입니다.
실패한 합계의 중간 값을 성공 결과로 반환하지 않습니다.

### 출력 예

정상 값은 10, 20, 30입니다.
그 뒤 별도로 취소된 Context를 만들어 같은 값을 전달합니다.

```text
합계: 60
취소 확인: true
```

### 경계표

| 입력 | 기대 |
| --- | --- |
| nil 또는 빈 슬라이스 | 0 |
| 0 하나 | 0 |
| 1000 하나 | 1000 |
| 1000개 × 1000 | 1000000 |
| -1 또는 1001 포함 | 오류 |
| 값 1001개 | 오류 |
| 취소 후 호출 | 작업 시작 전 취소 오류 |
| 과거 마감 후 호출 | DeadlineExceeded |
| 취소 상태와 빈 슬라이스 | 취소 오류 |

### 힌트

`ctx.Err()`가 nil이 아닌 경우 바로 반환합니다.
일반 오류는 오류 메시지를 새로 만들 수 있지만
취소 오류는 원인을 유지해야 errors.Is로 확인할 수 있습니다.

작은 int 덧셈이 너무 빨라서 중간 취소를 재현하기 어렵다고
테스트에 Sleep을 넣지 않습니다.
중간 취소의 결정적인 재현은 콜백이 있는 문제 6에서 수행합니다.

[참고 구현](../../solutions/17-context/checked-sum/main.go) ·
[테스트](../../solutions/17-context/checked-sum/main_test.go)

## 문제 3. 끊어진 Context 연결

**수준: 확인·변형**

다음은 의도적으로 잘못된 전체 프로그램입니다.
취소된 Context를 전달하지만 하위 함수가 그 연결을 버립니다.

```go
package main

import (
	"context"
	"fmt"
)

func work(ctx context.Context) error {
	independent := context.Background()
	return independent.Err()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := work(ctx)
	fmt.Println("성공으로 처리:", err == nil)
}
```

### 할 일

- 실제 출력을 예측합니다.
- 받은 Context의 취소 상태를 전달하도록 work를 고칩니다.
- 수정 후 errors.Is로 Canceled가 확인되는지 검사합니다.
- work가 취소 함수를 새로 만들어야 하는지 설명합니다.

### 완료 조건

수정 후 “성공으로 처리”는 false여야 합니다.
새 Background로 바꾸는 것이 취소를 해결하는 방법이라고 설명하지 않습니다.
필요 없는 WithCancel과 defer를 추가하지 않습니다.

### 추가 질문

함수가 WithCancel로 자식을 만든 뒤
`defer cancel()`을 등록하고 자식 Context만 반환하면,
호출자는 어떤 상태의 Context를 받는지 설명합니다.

### 힌트

Context는 이미 함수 인자로 전달되었습니다.
새로운 수명 범위가 필요한 것이 아니라 받은 연결을 사용하는 것이 목적입니다.
자식을 만들어 반환하는 추가 질문은 반환 뒤 수명의 소유권 문제입니다.

## 문제 4. 작업의 시간 예산과 정리

**수준: 응용**

개인 파일: `_practice/17-context/scoped-task/main.go`

```go
func withBudget(
	ctx context.Context,
	budget time.Duration,
	task func(context.Context) error,
) error
```

### 계약

- 음수 budget과 nil task는 일반 오류입니다.
- 0 budget은 허용하며 즉시 마감된 하위 Context가 됩니다.
- 유효한 입력에서는 전달받은 ctx에서 WithTimeout으로 자식을 만듭니다.
- 자식이 이미 취소됐으면 task를 호출하지 않습니다.
- task에는 자식 Context를 전달합니다.
- task가 오류를 반환하면 그 오류를 `%w`로 감싸서 반환합니다.
- task가 nil을 반환하면 자식의 마지막 Err를 확인합니다.
- 모든 반환 경로에서 만든 자식을 cancel합니다.
- 부모를 직접 cancel하지 않습니다.

task가 Context를 무시하며 영원히 실행되면 이 함수도 강제로 중단시키지 못합니다.
이 함수는 고루틴을 따로 만들지 않고 task를 직접 호출합니다.
“시간이 지나면 어떤 함수든 끊어 준다”는 설명을 하지 않습니다.

### 관찰할 결과

정상 task에서 전달받은 자식을 변수에 저장합니다.
withBudget이 돌아온 뒤 그 자식의 Err가 Canceled인지 확인합니다.
이것은 작업이 실패했다는 뜻이 아니라 수명을 정리했다는 관찰입니다.

```text
작업 성공: true
반환 후 정리: true
0초 마감: true
0초 작업 실행: false
```

### 경계 검사

| 조건 | 기대 |
| --- | --- |
| 1시간 예산, 즉시 성공 task | nil, 반환 뒤 자식 취소 |
| task가 일반 오류 반환 | 원인 보존, 자식 정리 |
| 0 예산 | DeadlineExceeded, task 미실행 |
| 음수 예산 | 일반 오류, task 미실행 |
| nil task | 일반 오류 |
| 부모가 미리 취소됨 | Canceled, task 미실행 |
| task 안에서 부모 취소 후 nil 반환 | Canceled |
| 정상 task 뒤 부모 확인 | 부모는 계속 유지 |

1시간은 실제로 기다리는 시간이 아닙니다.
짧게 끝나는 task에 충분히 큰 예산을 주고 정상 정리 경로를 검증합니다.
호출이 반환했는지를 시계로 추측하지 않습니다.

[참고 구현](../../solutions/17-context/scoped-task/main.go) ·
[테스트](../../solutions/17-context/scoped-task/main_test.go)

## 문제 5. 검증된 추적 ID

**수준: 변형**

개인 파일: `_practice/17-context/context-values/main.go`

함수 계약:

```go
func withTraceID(ctx context.Context, id string) (context.Context, error)
func traceID(ctx context.Context) (string, bool)
```

### 요구사항

- id의 앞뒤 공백을 제거합니다.
- 빈 id는 nil Context와 오류를 반환합니다.
- 유효한 id는 자체 비공개 키 타입으로 WithValue에 저장합니다.
- traceID는 타입과 존재 여부를 확인합니다.
- 부모의 취소·마감 연결을 유지합니다.
- 자식에 같은 키로 다른 ID를 넣어도 부모의 ID는 바뀌지 않습니다.

오류로 반환한 nil Context를 다른 Context 함수에 넘기지 않습니다.
일반적인 입력 Context는 nil이 아니라는 전제입니다.
추적 ID는 변경하지 않는 문자열로 사용합니다.

### 출력 예

```text
추적 ID: trace-001 true
취소 전달: true
값 유지: trace-001 true
```

### 검사할 차이

| 상황 | 기대 |
| --- | --- |
| Background에서 조회 | 빈 문자열, false |
| `" trace-001 "` 저장 | `"trace-001"`, true |
| 빈 값 또는 공백 | 오류 |
| 부모 ID parent, 자식 ID child | 각자 값 유지 |
| 다른 키 타입에 값 저장 | 추적 ID와 충돌하지 않음 |
| 부모 취소 | 자식도 취소 |
| 과거 부모 마감 | 자식에서 DeadlineExceeded |
| 취소 뒤 ID 조회 | 저장한 값 유지 |

### 힌트

`type traceIDKey struct{}`처럼 비교 가능한 자체 타입을 사용합니다.
Value가 any를 반환하므로 두 반환값 타입 단언을 사용합니다.
Context에 옵션과 비밀 값 전체를 넣는 연습으로 확장하지 않습니다.

[참고 구현](../../solutions/17-context/context-values/main.go) ·
[테스트](../../solutions/17-context/context-values/main_test.go)

## 문제 6. 전체 성공 결과만 공개하기

**수준: 응용**

개인 파일: `_practice/17-context/cancel-batch/main.go`

```go
func applyAll(
	ctx context.Context,
	values []int,
	transform func(context.Context, int) (int, error),
) ([]int, error)
```

### 계약

| 항목 | 요구사항 |
| --- | --- |
| 값 개수 | 최대 1000개 |
| transform | nil이면 오류 |
| 호출 순서 | 입력 순서 |
| Context | 같은 ctx를 transform에 전달 |
| 확인 지점 | 입구·각 호출 전·각 호출 뒤·반환 직전 |
| 오류 결과 | nil 슬라이스와 오류 |
| 빈 정상 입력 | nil이 아닌 빈 슬라이스 |
| 정상 결과 | 변환한 값들을 입력 순서로 반환 |

취소된 Context는 입구에서 먼저 거부합니다.
transform이 일반 오류를 반환하면 그 원인을 보존합니다.
그 호출이 동시에 취소도 요청했다면 이 구현 계약에서는 task 오류가 먼저입니다.
task가 nil을 반환했을 때 후속 취소 검사를 수행합니다.

### 정상·취소 예시

정상 변환은 1, 2, 3을 두 배로 만듭니다.
취소 예에서는 값 2를 처리하는 transform이 cancel을 호출하고
계산한 값과 nil을 반환합니다.

예상 출력:

```text
정상 결과: [2 4 6]
실패 결과 없음: true
취소 확인: true
```

### 검증

- 호출 전 취소이면 transform이 0번 호출됩니다.
- 첫 호출에서 취소하면 1번만 호출되고 결과는 nil입니다.
- 마지막 호출에서 취소해도 결과는 nil입니다.
- 두 번째 호출이 일반 오류를 반환하면 세 번째는 실행되지 않습니다.
- 1000개는 허용하고 1001개는 거부합니다.
- transform이 받는 Context가 취소·값 정보를 유지하는지 확인합니다.
- 값의 산술 규칙은 transform이 책임집니다. 본 예제는 작은 정수만 두 배로 만듭니다.

### 부수 효과 주의

nil 결과를 반환한다고 transform 안에서 수행한 외부 쓰기가 되돌아가지는 않습니다.
이 문제는 반환할 결과 슬라이스를 공개하지 않는 정책입니다.
파일·서비스 변경의 롤백은 별도 설계입니다.

### 힌트

결과는 로컬 슬라이스에 모읍니다.
각 오류 경로에서 그 슬라이스를 그대로 반환하지 않습니다.
콜백 안에서 정해진 호출 횟수에 cancel을 실행하면 Sleep 없이 재현할 수 있습니다.

[참고 구현](../../solutions/17-context/cancel-batch/main.go) ·
[테스트](../../solutions/17-context/cancel-batch/main_test.go)

## 선택 문제. 취소 가능한 대기 검사

**수준: 선택 심화**

본문 11절의 wait 함수를 개인 폴더에 옮깁니다.
채널과 select는 19장의 선행 미리 보기이므로 기본 문제와 구분합니다.

확인할 계약은 다음과 같습니다.

| 입력 | 기대 |
| --- | --- |
| 살아 있는 Background, 0초 | 정상 반환 |
| 이미 취소된 Context, 1시간 | Canceled, 장시간 대기 없음 |
| 과거 마감 Context, 1시간 | DeadlineExceeded |
| 살아 있는 Context, 음수 | 일반 오류 |
| 이미 취소된 Context, 음수 | 입구의 취소 오류가 먼저 |

정확히 몇 마이크로초에 반환해야 한다는 단언은 하지 않습니다.
검사 프로세스에는 전체 timeout을 두어 잘못된 코드의 무한 대기를 막을 수 있습니다.
timeout에 도달하는 것을 정상 성공 조건으로 사용하지 않습니다.

두 select 분기가 모두 준비되면 소스 위쪽 분기가 항상 이긴다고 설명하지 않습니다.
취소 확인이 어디에서 수행되는지 코드와 함께 설명합니다.

## 완료 점검

- Context를 첫 인자로 전달했다.
- 받은 ctx를 불필요하게 Background로 바꾸지 않았다.
- 새 자식을 만들었으면 그 수명의 정리를 담당했다.
- 이미 취소된 입력에서는 일을 시작하지 않았다.
- callback의 종료와 취소 신호의 관찰을 구분했다.
- 오류 반환 시 결과 정책을 문서화했다.
- Context 값에 자체 키 타입을 사용했다.
- 테스트 성공에 Sleep을 사용하지 않았다.

[해설](../../solutions/17-context/README.md) ·
[본문](../../book/17-context/README.md)

확인일: 2026-11-09. Go 1.27.3 기준입니다.
