# 25장 연습문제 — 서비스의 시작과 종료

[본문](../../book/25-service-operations/README.md)에서 설정, 로그와 요청 수명을 확인한 뒤 풀어 봅니다. 이번 문제의 목표는 출력 한 줄을 맞히는 것뿐 아니라, 실패한 설정으로 서비스를 시작하지 않고 종료할 때 남은 작업을 설명하는 것입니다.

모든 명령은 **저장소 루트**에서 실행합니다. 문제용 파일은 `_practice/25-service-operations/` 아래에 만듭니다. 같은 폴더에 여러 `main` 함수를 두지 않습니다. 모범 코드는 마지막의 해설 링크에 있으며, 문제를 읽을 때 먼저 열지 않아도 됩니다.

## 준비

다음 네 폴더에 독립 프로그램을 작성합니다.

| 문제 | 개인 작업 경로 | 시작할 예제 |
| --- | --- | --- |
| 2 | `_practice/25-service-operations/config-precedence` | 설정 예제와 16장 flag |
| 4 | `_practice/25-service-operations/redacted-account` | secret-value |
| 5 | `_practice/25-service-operations/draining-gate` | health-readiness |
| 6 | `_practice/25-service-operations/managed-server` | service 패키지와 22장 서버 |

Windows PowerShell:

```powershell
$practiceRoot = "_practice/25-service-operations"
foreach ($name in @("config-precedence", "redacted-account", "draining-gate", "managed-server")) {
    New-Item -ItemType Directory -Force "$practiceRoot/$name" | Out-Null
}
```

macOS·Linux·WSL:

```bash
practice_root=_practice/25-service-operations
for name in config-precedence redacted-account draining-gate managed-server; do
    mkdir -p "$practice_root/$name"
done
```

위 명령은 폴더만 만듭니다. 이미 작성한 `main.go`를 덮어쓰지 않습니다. 첫 줄에 `package main`을 작성하고 필요한 import와 함수를 직접 추가합니다.

## 문제 1. 확인 — 시작 실패와 요청 실패

다음 상황을 순서대로 읽고 각 질문에 답합니다.

1. 기본 읽기 제한은 5초입니다.
2. 사용자가 `APP_READ_TIMEOUT=oops`를 설정했습니다.
3. 설정을 읽는 함수가 오류를 반환했습니다.
4. 호출자는 오류를 기록하고 실행을 끝냅니다.
5. HTTP listener는 아직 생성하지 않았습니다.

### 질문

1. 서버가 5초 제한으로 실행 중이라고 말해도 될까요?
2. 이 실패를 HTTP 500 응답으로 확인할 수 있을까요?
3. 포트를 먼저 열고 설정을 나중에 검사하면 어떤 상태가 생길 수 있나요?
4. 오류 메시지에 어떤 설정 이름을 넣으면 도움이 되나요?
5. 모든 환경 변수의 원문을 한꺼번에 출력하는 방법에는 어떤 문제가 있나요?

추가로 다음 두 로그를 비교합니다.

```text
요청 GET /items/7?token=abc가 실패했다
method=GET route="GET /items/{id}" status=500
```

어떤 정보를 유지하고 어떤 정보를 제외했는지 적습니다. 비밀값뿐 아니라 항목 번호가 요청마다 달라질 때 집계하기 쉬운지도 생각합니다.

### 완료 조건

- 프로세스 시작 실패와 개별 요청 실패를 구분했습니다.
- 오류가 발생했을 때 실제로 소유한 자원이 무엇인지 설명했습니다.
- 로그의 정보량과 필요한 정보의 차이를 설명했습니다.
- 실패했는데 실행 중이라고 보고하지 않습니다.

<details>
<summary>힌트</summary>

1. 아직 listener가 없으면 요청을 받을 준비도 되지 않았습니다.
2. 기본값은 변수가 없을 때 쓰는 정책이지, 모든 잘못된 값을 숨기는 장치가 아닙니다.
3. 로그에 필요한 것은 단계·설정 항목·오류 분류이며, 모든 원문이 아닙니다.

</details>

## 문제 2. 변형 — 기본값, 환경 변수, 플래그

페이지 크기 `limit`을 결정하는 함수를 만듭니다.

```go
func limitConfig(args []string, lookup func(string) (string, bool)) (int, error)
```

### 요구사항

- 기본값은 20입니다.
- 환경 변수 `APP_LIMIT`이 있으면 먼저 읽습니다.
- 명령행 `-limit`이 있으면 유효한 환경 설정 위에 덮어씁니다.
- 최종 값의 범위는 1~100입니다.
- 존재하는 환경 변수가 빈 값, 정수가 아닌 값, 범위 밖 값이면 오류입니다.
- 잘못된 환경 변수를 정상 플래그로 숨기지 않습니다. 환경 검증이 먼저입니다.
- 알 수 없는 플래그와 불필요한 위치 인자는 오류입니다.
- 함수는 환경을 변경하거나 `os.Exit`을 호출하지 않습니다.

### 확인 표

| 환경 | 인자 | 결과 |
| --- | --- | --- |
| 없음 | 없음 | 20 |
| 30 | 없음 | 30 |
| 30 | `-limit=40` | 40 |
| 없음 | `-limit=1` | 1 |
| 없음 | `-limit=100` | 100 |
| 30 | `-limit=0` | 오류 |
| bad | `-limit=20` | 오류 |
| 30 | `extra` | 오류 |

기본 프로그램은 환경 맵과 인자 슬라이스를 직접 만들어 실행 결과를 비교합니다.

```text
크기: 20
크기: 30
크기: 40
범위 오류: true
```

### 완료 조건

- 파싱 성공과 값의 허용 범위를 따로 검사합니다.
- `flag.CommandLine` 전역 상태에 의존하지 않습니다.
- 테스트에서 실제 환경 변수를 변경하지 않습니다.
- 경계 1·100, 잘못된 0·101을 포함합니다.

<details>
<summary>힌트</summary>

1. 환경 검증을 마친 값을 플래그의 기본값으로 등록합니다.
2. `flag.NewFlagSet`과 `flag.ContinueOnError`를 사용합니다.
3. `Parse` 뒤에는 `NArg`와 최종 범위를 확인합니다.

</details>

## 문제 3. 확인·수정 — 종료 context의 부모

다음은 **의도적으로 잘못된 종료 코드의 발췌**입니다. 정상 예제 파일에 그대로 합치지 않습니다.

```go
<-runContext.Done()
shutdownContext, cancel := context.WithTimeout(runContext, 3*time.Second)
defer cancel()
err := server.Shutdown(shutdownContext)
```

### 질문과 수정

1. `runContext.Done()`에서 돌아왔다는 것은 어떤 상태인가요?
2. 자식 context에 3초를 지정했으므로 3초를 기다릴 수 있나요?
3. 정리를 위한 context는 어디에서 시작해야 할까요?
4. `Shutdown`이 시간 초과되면 어떤 추가 동작을 하고 어떤 오류를 남길까요?
5. `Serve`가 `ErrServerClosed`를 반환했으므로 종료 대기가 끝났다고 판단해도 될까요?

### 완료 조건

- 자식 context가 취소된 부모를 되살리지 못한다는 것을 설명합니다.
- 새 정리 context의 취소 함수도 호출합니다.
- 정상 종료와 강제 연결 닫기를 구분합니다.
- 본문의 `service.Serve`를 읽고 두 결과를 모두 수집하는 위치를 찾습니다.

<details>
<summary>힌트</summary>

1. 실행 수명과 정리 수명을 나누어 봅니다.
2. `context.Background()`에서 제한 시간을 만드는 코드를 찾습니다.
3. `Shutdown`의 반환과 `Serve`의 반환은 같은 사건이 아닙니다.

</details>

## 문제 4. 변형 — 기록할 필드를 허용 목록으로 정하기

다음 타입을 사용합니다.

```go
type Account struct {
    ID       string
    Email    string
    Password string
}
```

`Account`가 `slog.LogValuer`를 구현하게 만듭니다. 로그에는 `id`만 포함하고 이메일과 비밀번호는 키도 값도 넣지 않습니다.

### 요구사항

- `logger.Info("account", "account", value)`로 구조체를 그대로 전달합니다.
- `LogValue`는 그룹 값을 반환합니다.
- 출력 검증에서는 시각 속성을 제거하여 재현 가능하게 만듭니다.
- 필드가 추가되더라도 자동으로 모든 필드를 노출하는 코드를 사용하지 않습니다.

### 입력과 출력

입력 ID는 `student-7`입니다. 나머지 값은 공개용 가짜 문자열로 작성합니다.

```text
{"level":"INFO","msg":"account","account":{"id":"student-7"}}
```

### 완료 조건

- JSON을 파싱하여 `account` 그룹에 `id`만 있는지 확인합니다.
- 이메일과 비밀번호의 원문이 로그 바이트에 없는지도 확인합니다.
- 실제 계정 자료를 실습에 사용하지 않습니다.
- `fmt.Println(account)`에도 같은 효과가 적용된다고 설명하지 않습니다.

<details>
<summary>힌트</summary>

1. `slog.GroupValue`에 허용한 `slog.String`만 넣습니다.
2. 구조체 전체를 문자열로 바꾼 뒤 정규식으로 지우는 방식은 필요하지 않습니다.
3. 테스트에서 전체 JSON 문자열의 시각까지 고정할 필요는 없습니다.

</details>

## 문제 5. 응용 — 종료 중 새 작업 거절하기

준비 상태를 내린 뒤에도 이미 도착한 요청이 있을 수 있습니다. 다음 형태의 미들웨어를 구현합니다.

```go
func gate(accepting *atomic.Bool, next http.Handler) http.Handler
```

### 요구사항

- accepting이 참이면 `next`를 한 번 호출합니다.
- 거짓이면 `next`를 호출하지 않고 503을 반환합니다.
- 거절 응답에는 `Retry-After: 1`을 넣습니다.
- 이미 게이트를 통과한 요청은 상태가 바뀌어도 정상 완료할 수 있습니다.
- 이 게이트가 요청 완료를 기다리는 역할까지 한다고 가정하지 않습니다.

기본 실행에서는 참·거짓 두 상태의 요청 결과를 출력합니다.

```text
상태: 204
상태: 503
```

### 동시성 테스트

1. 첫 요청을 시작하고 핸들러 진입 신호를 받습니다.
2. 핸들러는 채널에서 기다리게 합니다.
3. accepting을 거짓으로 바꿉니다.
4. 두 번째 요청이 503인지 확인합니다.
5. 첫 요청의 대기를 풀고 204인지 확인합니다.

`time.Sleep`으로 핸들러 진입을 추측하지 않습니다. 테스트가 소유한 채널로 사건을 전달합니다.

### 완료 조건

- 원자적 읽기·쓰기를 사용합니다.
- 거절된 요청은 실제 작업에 들어가지 않습니다.
- 이미 시작한 작업을 새 상태 값 하나로 자동 취소한다고 생각하지 않습니다.
- race detector를 실행할 수 있는 환경에서는 해당 검사를 수행합니다.

<details>
<summary>힌트</summary>

1. 상태 읽기와 핸들러 실행은 전체가 하나의 원자적 작업은 아닙니다.
2. 게이트를 통과한 순간부터 기존 요청으로 취급하는 계약을 정합니다.
3. 테스트는 첫 요청이 실제로 들어갔다는 신호를 확인한 후 상태를 바꿉니다.

</details>

## 문제 6. 응용 — 자동 데모와 직접 실행 모드

`GET /hello`에 한국어 인사를 반환하는 프로그램을 만듭니다. 시작 코드가 플래그, listener와 서버 수명을 관리하게 합니다.

### 요구사항

- 기본 실행은 `httptest`로 요청 한 번을 처리하고 자동 종료합니다.
- `-serve`를 지정하면 실제 HTTP 서버를 실행합니다.
- `-addr`의 기본값은 `127.0.0.1:8080`입니다.
- 직접 입력한 IP가 loopback이 아니면 시작을 거절합니다.
- 실제 실행은 `service.Serve`를 재사용합니다.
- `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`을 명시합니다.
- JSON 형식의 시작·종료 로그는 표준 오류에 씁니다.
- Ctrl+C를 종료 요청으로 연결하고 신호 등록도 해제합니다.
- `main` 외 함수에서 `os.Exit`을 호출하지 않습니다.

기본 출력:

```text
상태: 200
안녕하세요
```

직접 실행:

```sh
go run ./_practice/25-service-operations/managed-server -serve -addr=127.0.0.1:8080
```

다른 터미널에서 `http://127.0.0.1:8080/hello`를 요청하고 첫 터미널에서 Ctrl+C로 종료합니다. 포트가 사용 중이면 다른 loopback 포트를 선택합니다.

### 완료 조건

- 기본 데모는 사용자의 입력을 기다리지 않습니다.
- 직접 실행 모드는 종료 방법이 설명되어 있습니다.
- 외부 인터페이스 주소와 잘못된 인자를 거절합니다.
- 시작 전 취소된 context에서는 서버를 열지 않습니다.
- 신호 등록을 해제한 뒤 오류 종료합니다.

<details>
<summary>힌트</summary>

1. `run(ctx, args, out, diagnostics)`를 만들면 테스트가 프로세스 인자와 출력에 의존하지 않습니다.
2. `net.ParseIP(host).IsLoopback()`을 확인합니다. nil 여부부터 검사합니다.
3. 운영체제 신호는 `main`, 서버 실행은 `run`, 종료 대기는 `service`가 맡게 합니다.

</details>

## 선택 문제. 요청 로그의 범위 설명하기

본문의 `responseCapture`가 어떤 응답에 적합한지 적습니다.

- `Write`만 호출한 일반 텍스트 응답
- 최종 상태를 두 번 설정하려는 핸들러
- 아무것도 쓰지 않는 핸들러
- `http.Flusher`를 직접 타입 단언하는 스트리밍 코드
- `http.Hijacker`로 연결을 가져가는 코드

각각 “그대로 지원”, “추가 설계 필요”로 분류하고 이유를 한 문장씩 적습니다. 실제로 사용하지 않는 인터페이스를 추측으로 흉내 내어 구현할 필요는 없습니다.

## 결과 확인

```sh
go fmt ./_practice/25-service-operations/config-precedence
go vet ./_practice/25-service-operations/config-precedence
go run ./_practice/25-service-operations/config-precedence
```

나머지 문제도 해당 폴더를 지정해 실행합니다. `_practice`는 일반 `./...` 탐색에서 제외될 수 있으므로 실습 폴더를 명시합니다. 검사는 자신의 프로그램 경로를 대상으로 합니다.

[해설과 검증 코드 보기](../../solutions/25-service-operations/README.md) · [본문으로 돌아가기](../../book/25-service-operations/README.md)
