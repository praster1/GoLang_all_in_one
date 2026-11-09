# 22장 연습문제: HTTP 서버와 라우팅

선수 학습: [22장 본문](../../book/22-http-server/README.md)

기본 여섯 문제에서 응답 작성 순서, 라우팅, 요청 검증과 종료를 구현합니다.
모든 명령은 **저장소 루트**에서 실행합니다.
각 프로그램은 로컬 서버를 만들고 정해진 요청을 보낸 뒤 스스로 종료합니다.
외부 사이트나 별도 서버 프로세스는 필요하지 않습니다.

## 문제와 개인 폴더

| 문제 | 수준 | 개인 폴더 |
| --- | --- | --- |
| 1. 응답 상태 예측 | 확인 | 학습 노트에 기록 |
| 2. 상태 확인 경로 | 변형 | `_practice/22-http-server/health/` |
| 3. 오류 뒤 실행 고치기 | 오류 진단 | `_practice/22-http-server/repair/` |
| 4. 제곱 계산 경로 | 변형 | `_practice/22-http-server/square/` |
| 5. 계획 미리 보기 | 응용 | `_practice/22-http-server/plan-preview/` |
| 6. 정상 종료하는 서버 | 응용 | `_practice/22-http-server/shutdown-health/` |
| 선택 A. HEAD 전용 처리 | 선택 | `_practice/22-http-server/custom-head/` |

## 준비

첫 핸들러 예제를 출발점으로 복사합니다.
기존 파일이 있다면 내용을 먼저 확인하고 보존합니다.
서버·클라이언트 정리와 오류 처리를 유지하면서 문제에 맞게 경로와 응답을 변경하세요.
테스트 파일은 복사하지 않으며 하위 모듈은 만들지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("health", "repair", "square", "plan-preview", "shutdown-health", "custom-head")) {
    $practiceDir = Join-Path "_practice/22-http-server" $practiceName
    New-Item -ItemType Directory -Force $practiceDir | Out-Null
    $practiceMain = Join-Path $practiceDir "main.go"
    if (Test-Path -LiteralPath $practiceMain) {
        Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
    } else {
        Copy-Item examples/22-http-server/hello-handler/main.go $practiceMain
    }
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
for practice_name in health repair square plan-preview shutdown-health custom-head; do
    practice_dir="_practice/22-http-server/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/22-http-server/hello-handler/main.go "$practice_dir/main.go"
    fi
done
```

## 문제 1. 본문 뒤의 상태는 바뀔까? — 확인

다음은 **의도적으로 잘못된 핸들러의 발췌 코드**입니다.

```go
fmt.Fprintln(w, "잘못된 입력")
w.WriteHeader(http.StatusBadRequest)
```

다음 질문에 먼저 답하세요.

1. 클라이언트가 받는 최종 상태는 200인가, 400인가?
2. 첫 줄은 어떤 상태를 암묵적으로 선택하는가?
3. 상태를 400으로 보내려면 어느 순서를 바꿔야 하는가?
4. `Content-Type`도 본문을 쓴 뒤에 정하면 되는가?

직접 확인하려면 `httptest.NewRecorder()`에 위 핸들러를 호출하여 기록된 상태를 봅니다.
전송 계층이 필요한 검사가 아니므로 이 확인에는 실제 서버를 띄우지 않아도 됩니다.

<details>
<summary>힌트 1</summary>

`ResponseWriter.Write`는 아직 최종 상태가 정해지지 않았다면 200을 사용합니다.
본문을 전송하기 시작한 뒤 최종 상태를 되돌리는 방식으로 생각하지 않습니다.

</details>

<details>
<summary>힌트 2</summary>

데이터 준비 → 헤더 → 최종 상태 → 본문 순서로 적어 보세요.
[ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)의 계약을 확인합니다.

</details>

완료 조건: 실제 상태와 수정할 순서를 설명하고, 본문에 오류 문장이 있다는 사실만으로 실패 상태가 되지 않음을 구분합니다.

## 문제 2. 상태 확인 경로 — 변형

작업 파일: `_practice/22-http-server/health/main.go`

`newHandler() http.Handler`가 새 라우터를 반환하도록 작성하세요.
`GET /health`에서 `text/plain; charset=utf-8` 형식의 `ready\n`을 보냅니다.
명시적인 HEAD 전용 경로는 등록하지 않습니다.

| 요청 | 기대 상태 | 기대 본문 |
| --- | --- | --- |
| `GET /health` | 200 | `ready\n` |
| `HEAD /health` | 200 | 실제 HTTP 전송 본문 없음 |
| `POST /health` | 405 | 라우터 기본 메서드 오류 |
| `GET /missing` | 404 | 라우터 기본 경로 오류 |

405에는 `Allow: GET, HEAD`가 있어야 합니다.
기본 프로그램은 GET과 POST를 차례대로 보내 다음을 출력합니다.

```text
GET /health -> 200: ready
POST /health -> 405: Method Not Allowed
```

```text
go run ./_practice/22-http-server/health
```

<details>
<summary>힌트 1</summary>

메서드가 없는 `/health` 대신 메서드가 포함된 패턴을 등록합니다.
GET 패턴과 HEAD의 관계를 본문에서 다시 확인하세요.

</details>

<details>
<summary>힌트 2</summary>

HEAD의 전송 본문은 실제 임시 서버와 클라이언트로 확인합니다.
핸들러를 recorder에 직접 호출하는 검사와 전송 계층을 통과한 결과를 구분합니다.

</details>

완료 조건: 성공 응답뿐 아니라 HEAD·405·404·Allow를 함께 확인합니다.

## 문제 3. 오류 뒤에도 성공 문장이 붙는다 — 오류 진단

작업 파일: `_practice/22-http-server/repair/main.go`

다음은 `GET /hello`의 **잘못된 핸들러 발췌 코드**입니다.

```go
if r.URL.Query().Get("name") == "" {
    http.Error(w, "이름 필요", http.StatusBadRequest)
}
fmt.Fprintln(w, "완료")
```

이름이 없으면 오류 본문 뒤에 성공 문장도 붙습니다.
실패 응답을 보낸 뒤 현재 핸들러를 끝내도록 고치세요.
이름이 있으면 200과 `완료\n`을 반환합니다.
공백 정규화는 이 문제의 범위에 넣지 않습니다.

기본 요청 두 개의 목표 출력:

```text
GET /hello -> 400: 이름 필요
GET /hello?name=Go -> 200: 완료
```

```text
go run ./_practice/22-http-server/repair
```

<details>
<summary>힌트 1</summary>

`http.Error`는 응답을 쓰는 함수입니다.
호출자의 나머지 Go 문장을 자동으로 건너뛰게 만들지는 않습니다.

</details>

<details>
<summary>힌트 2</summary>

실패 응답 다음에 `return`을 넣어 실행 경로를 분리합니다.
출력 문구를 바꾸거나 성공 문장을 항상 지우는 것으로 문제를 피하지 않습니다.

</details>

완료 조건: 실패 본문에 성공 문장이 없고, 유효한 요청의 성공 동작도 유지됩니다.

## 문제 4. 제곱 계산 경로 — 변형

작업 파일: `_practice/22-http-server/square/main.go`

`GET /squares/{n}`을 구현합니다.
`n`을 `int`로 변환한 뒤 0~1000 범위인지 확인합니다.
범위 안이면 JSON으로 입력과 제곱을 반환합니다.

| 조건 | 상태 | JSON 본문 |
| --- | --- | --- |
| `n=12` | 200 | `{"input":12,"square":144}` |
| `n=0` | 200 | `{"input":0,"square":0}` |
| 변환 실패·범위 밖 | 400 | `{"error":"invalid_number"}` |

본문 마지막에 줄바꿈을 붙이고 `Content-Type`은 `application/json`으로 정합니다.
계산과 JSON 준비를 끝낸 뒤 응답을 쓰세요.
기본 프로그램은 12, 0, 1001을 요청합니다.

```text
GET /squares/12 -> 200: {"input":12,"square":144}
GET /squares/0 -> 200: {"input":0,"square":0}
GET /squares/1001 -> 400: {"error":"invalid_number"}
```

```text
go run ./_practice/22-http-server/square
```

추가로 -1, 1, 1000, `x`, `int`로 표현할 수 없는 긴 숫자를 검사합니다.
1000의 제곱은 1000000입니다.
`/squares/1/extra`는 등록한 경로와 다르므로 404가 됩니다.

<details>
<summary>힌트 1</summary>

라우터가 채운 값은 `r.PathValue("n")`으로 읽습니다.
정수 변환 성공과 0~1000 범위 검사를 모두 통과해야 계산합니다.

</details>

<details>
<summary>힌트 2</summary>

오류 응답 helper는 상태와 오류 코드를 받아 JSON으로 씁니다.
성공과 실패의 `Content-Type`도 검사하세요.

</details>

완료 조건: 0을 유효하게 처리하고, 과도한 정수·음수·문자 입력은 계산 전에 거절합니다.

## 문제 5. 학습 계획 미리 보기 — 응용

작업 파일: `_practice/22-http-server/plan-preview/main.go`

`POST /plans/preview`는 JSON `days`, `minutes`를 받아 총 분량을 계산합니다.
새 데이터를 저장하는 기능이 아니므로 성공은 200이며 `Location`을 보내지 않습니다.

| 검증 | 계약 |
| --- | --- |
| 형식 | `application/json`, 미디어 타입 매개변수 허용 |
| 크기 | 전체 본문 최대 128바이트 |
| JSON | 객체 하나, 정의하지 않은 필드·추가 값·잘못된 타입 거절 |
| 날짜 | 0~7 |
| 하루 분량 | 1~120 |
| 성공 | `{"total_minutes": days * minutes}` |

실패 코드는 미디어 타입 415, 크기 초과 413, JSON 해석 400, 업무 범위 422로 구분합니다.
`null`과 `{}`는 유효한 계획 값이 아니므로 422입니다.
0일이어도 하루 분량의 범위 검사는 생략하지 않습니다.
이 계약에서 `days`를 생략하면 정수의 영값인 0일로 해석합니다. `minutes`를 생략한 경우의 0은 범위 밖이므로 거절합니다. 두 필드의 존재 자체를 필수로 구분하는 별도 검증은 요구하지 않습니다.

기본 요청과 출력:

```text
POST /plans/preview -> 200: {"total_minutes":75}
POST /plans/preview -> 200: {"total_minutes":0}
POST /plans/preview -> 422: 계획 값을 확인하세요
```

차례로 `{"days":3,"minutes":25}`, `{"days":0,"minutes":25}`, `{"days":8,"minutes":25}`를 보냅니다.

```text
go run ./_practice/22-http-server/plan-preview
```

확인할 경계는 `(1,1) → 1`, `(7,120) → 840`, `(-1,25)`, `(8,25)`, `(0,0)`, `(5,121)`입니다.
정상 객체 뒤 공백으로 전체 128바이트와 129바이트를 만들어 한도도 확인하세요.
정상 객체 뒤 `{}`와 `x`를 붙인 경우는 400이어야 합니다.

<details>
<summary>힌트 1</summary>

입력 읽기와 HTTP 응답 결정을 나눕니다.
`MaxBytesReader` 오류는 `*http.MaxBytesError`인지 검사하여 다른 JSON 오류와 구분할 수 있습니다.

</details>

<details>
<summary>힌트 2</summary>

날짜와 분량을 검사한 뒤 곱합니다.
본문 한도·단일 JSON·알 수 없는 필드 검사는 업무 조건과 별개의 단계입니다.

</details>

완료 조건: 유효한 0분과 잘못된 입력을 구분하며 128/129바이트와 실패 종류별 상태를 확인합니다.

## 문제 6. 정상 종료하는 상태 서버 — 응용

작업 파일: `_practice/22-http-server/shutdown-health/main.go`

직접 `net.Listener`와 `http.Server`를 준비하여 `GET /health`에서 `healthy\n`을 보냅니다.
리스너는 `127.0.0.1:0`에 바인딩하여 사용 가능한 포트를 선택합니다.
`Serve`는 별도 고루틴에서 실행하고 결과 채널은 크기 1로 만듭니다.

정상 요청을 한 번 보낸 뒤 제한 시간이 있는 컨텍스트로 종료합니다.
`stopServer(ctx, server, served) error`에 다음 책임을 모으세요.

1. `Shutdown`을 호출하고 완료를 기다립니다.
2. 종료 오류를 원인을 보존하여 반환합니다.
3. `Serve`의 결과가 `http.ErrServerClosed`인지 확인합니다.

예상 출력:

```text
응답: healthy
정상 종료: true
```

```text
go run ./_practice/22-http-server/shutdown-health
```

서버에는 헤더 읽기 2초, 요청 읽기·응답 쓰기 5초, 유휴 연결 30초의 제한을 둡니다.
실험 요청의 클라이언트 제한은 2초, 종료 대기는 1초입니다.
정상 종료와 중간 실패 경로 모두 남은 서버 자원을 정리해야 합니다.

추가 테스트에서는 활성 요청 하나를 채널로 대기시키세요.
수신 리스너가 닫혀도 그 요청이 끝나기 전에 `stopServer`가 성공 반환하면 안 됩니다.
요청을 풀어 준 뒤 응답과 서버 종료가 모두 완료되어야 합니다.

<details>
<summary>힌트 1</summary>

`Shutdown`이 시작되면 `Serve`는 빨리 반환할 수 있습니다.
`Serve` 결과만 기다리고 프로세스를 끝내면 진행 중 요청을 기다리는 일을 놓칠 수 있습니다.

</details>

<details>
<summary>힌트 2</summary>

`Shutdown`을 먼저 기다린 다음 실행 결과를 확인합니다.
중간 오류에서 마지막 정리가 가능하도록 `defer server.Close()`도 둡니다.
고정 sleep 대신 시작·완료·수신 종료 신호를 사용하세요.

</details>

완료 조건: 자동 예제가 스스로 끝나고 활성 요청의 완료를 기다리는 종료를 검증합니다.

## 선택 A. HEAD를 직접 등록하기

작업 파일: `_practice/22-http-server/custom-head/main.go`

`GET /summary`와 `HEAD /summary`를 함께 등록합니다.
GET은 헤더 `X-Source: get`과 본문 `data`를 보냅니다.
HEAD는 헤더 `X-Source: head`를 보내고 본문을 쓰지 않습니다.
둘 다 상태는 200입니다.

실제 임시 서버에 요청해 상태, 받은 본문 바이트 수와 헤더를 출력합니다.

```text
GET: 200, 본문=4, 출처=get
HEAD: 200, 본문=0, 출처=head
```

```text
go run ./_practice/22-http-server/custom-head
```

<details>
<summary>힌트</summary>

GET 패턴도 HEAD를 받지만, 명시적인 HEAD 패턴은 더 구체적입니다.
본문 길이만 보면 어느 핸들러가 실행됐는지 모를 수 있어 `X-Source`로 구분합니다.

</details>

## 완료 점검과 해설

개인 폴더는 일반 `./...` 검사에서 제외되므로 실행·포맷·테스트에 경로를 직접 지정합니다.

```text
go fmt ./_practice/22-http-server/health
go fmt ./_practice/22-http-server/repair
go fmt ./_practice/22-http-server/square
go fmt ./_practice/22-http-server/plan-preview
go fmt ./_practice/22-http-server/shutdown-health
```

- [ ] 응답의 상태·헤더·본문을 서로 구분했다.
- [ ] 오류 응답 뒤 실행을 끝내는 위치를 확인했다.
- [ ] 정상 요청 외에 메서드·경로·입력 실패를 검사했다.
- [ ] JSON 크기와 값 범위를 별도로 확인했다.
- [ ] 종료 요청 후 서버 작업이 끝나는 것을 기다렸다.

[22장 해설](../../solutions/22-http-server/README.md) · [본문](../../book/22-http-server/README.md) · [연습문제 목록](../README.md)
