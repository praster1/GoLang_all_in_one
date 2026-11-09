# 21장 연습문제: HTTP 클라이언트

선수 학습: [21장 본문](../../book/21-http-client/README.md)

기본 여섯 문제에서 요청 구성, 응답의 소유권, 데이터 경계와 취소를 구현합니다.
선택 A는 리다이렉트를 한 번만 따라가는 정책입니다.
모든 명령은 **저장소 루트**에서 실행하며 외부 사이트를 호출하지 않습니다.
각 프로그램이 `httptest` 서버를 시작하고 요청을 마친 뒤 스스로 종료하도록 작성합니다.

## 문제와 개인 작업 폴더

| 문제 | 수준 | 개인 폴더 |
| --- | --- | --- |
| 1. 응답과 오류 예측 | 확인 | 먼저 학습 노트에 기록 |
| 2. 이름과 과정 전달 | 변형 | `_practice/21-http-client/header-greeting/` |
| 3. 닫기 위치 고치기 | 오류 진단 | `_practice/21-http-client/repair/` |
| 4. 한도가 있는 다운로드 | 변형 | `_practice/21-http-client/limited-download/` |
| 5. 학습 JSON 조회 | 응용 | `_practice/21-http-client/lesson-json/` |
| 6. 취소된 상태 조회 | 응용 | `_practice/21-http-client/cancel-check/` |
| 선택 A. 리다이렉트 한 번 | 선택 | `_practice/21-http-client/redirect-once/` |

## 준비

본문의 첫 클라이언트 예제를 출발점으로 사용합니다.
문제마다 함수와 실험 서버의 동작을 바꾸되, 요청의 오류 검사와 서버 종료는 유지합니다.
기존 파일은 덮어쓰지 않고 테스트 파일도 복사하지 않습니다.
새 `go.mod`나 `go.work`는 만들지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("header-greeting", "repair", "limited-download", "lesson-json", "cancel-check", "redirect-once")) {
    $practiceDir = Join-Path "_practice/21-http-client" $practiceName
    New-Item -ItemType Directory -Force $practiceDir | Out-Null
    $practiceMain = Join-Path $practiceDir "main.go"
    if (Test-Path -LiteralPath $practiceMain) {
        Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
    } else {
        Copy-Item examples/21-http-client/basic-get/main.go $practiceMain
    }
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
for practice_name in header-greeting repair limited-download lesson-json cancel-check redirect-once; do
    practice_dir="_practice/21-http-client/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/21-http-client/basic-get/main.go "$practice_dir/main.go"
    fi
done
```

## 문제 1. 404이면 `Do`도 오류일까? — 확인

실험 서버는 요청을 받으면 `http.StatusNotFound`와 `missing` 본문을 보냅니다.
클라이언트는 정상적으로 서버에 연결하고 응답 헤더를 받습니다.
다음은 `run` 안에서 응답을 처리하는 **발췌 코드**입니다.

```go
resp, err := client.Get(server.URL)
if err != nil {
    return err
}
defer resp.Body.Close()
fmt.Println("오류 없음:", err == nil)
fmt.Println("상태:", resp.StatusCode)
```

출력 두 줄을 예측하고 다음을 설명하세요.

1. 연결에 실패한 경우와 404 응답을 받은 경우는 어떻게 다른가?
2. 이 API가 200만 허용한다면 어느 시점에 실패로 바꾸어 반환해야 하는가?
3. 상태가 404여도 본문을 닫아야 하는가?
4. 응답 본문을 읽지 않을 때와 읽다가 실패할 때도 닫을 책임이 같은가?

<details>
<summary>힌트 1</summary>

`Client.Do`의 `error`는 HTTP 상태 코드 그 자체를 나타내지 않습니다.
본문에서 전송, 상태, 데이터의 세 단계를 다시 구분하세요.

</details>

<details>
<summary>힌트 2</summary>

응답을 얻었다면 상태를 검사하는 경로도 `defer Close`의 적용 대상이어야 합니다.
[Client.Do](https://pkg.go.dev/net/http#Client.Do)의 반환값 계약을 읽습니다.

</details>

완료 조건: 출력과 자원 정리의 책임을 둘 다 설명합니다.

## 문제 2. 이름과 과정 헤더 전달 — 변형

작업 파일: `_practice/21-http-client/header-greeting/main.go`

`greet(ctx context.Context, client *http.Client, address, name string) (string, error)`를 작성합니다.
GET 요청의 쿼리 `name`에 이름을 넣고 요청 헤더 `X-Course`를 `Go`로 설정합니다.
쿼리는 `net/url`로 인코딩하며 주소에 이미 있는 다른 쿼리는 보존합니다.

실험 서버는 받은 요청에서 이름과 과정 헤더를 읽어 다음 형식의 텍스트를 돌려줍니다.
`greet`는 상태 200을 확인하고 본문을 최대 1024바이트까지 허용하며, 항상 본문을 닫습니다.
클라이언트는 한 번 준비하고 2초의 시간 제한을 둡니다.

기본 이름 `민지 & Go`의 출력:

```text
이름: 민지 & Go
과정: Go
```

실행:

```text
go run ./_practice/21-http-client/header-greeting
```

이름을 빈 문자열, `Go & HTTP`, `한글 + ? #`로 바꾸어도 서버가 같은 값을 읽어야 합니다.
주소에 `?keep=1`을 붙였을 때 기존 쿼리도 남는지 확인합니다.
빈 이름은 이 문제에서 허용하며 서버의 `이름:` 뒤가 비어 있게 됩니다.
검사 후 기본 이름으로 복구합니다.

<details>
<summary>힌트 1</summary>

`u.Query()`에서 시작하면 기존 쿼리를 보존할 수 있습니다.
새 값을 넣은 뒤에는 인코딩 결과를 `u.RawQuery`에 다시 저장해야 합니다.

</details>

<details>
<summary>힌트 2</summary>

헤더를 지정하려면 요청 객체를 만든 뒤 `req.Header.Set`을 사용합니다.
문자열 안에 `&`나 `?`가 있다는 이유로 직접 치환하기보다 값 전체를 인코더에 맡기세요.

</details>

완료 조건: 특별한 문자가 있는 이름과 기존 쿼리가 유지되고, 함수가 본문을 닫습니다.

## 문제 3. 본문 닫기의 위치 고치기 — 오류 진단

작업 파일: `_practice/21-http-client/repair/main.go`

다음은 **잘못된 순서의 발췌 코드**입니다. 정상 응답에서는 문제가 보이지 않을 수 있습니다.

```go
resp, err := client.Get(address)
defer resp.Body.Close()
if err != nil {
    return "", err
}
```

이 순서를 고치고, 상태 200일 때만 최대 1024바이트의 텍스트를 반환하는 함수를 완성하세요.
로컬 정상 서버에는 `ok`를 보내게 하고, 잘못된 주소 `://bad`도 함수에 전달합니다.
오류 문자열 전체는 구현에 따라 달라질 수 있으므로 고정 문자열로 비교하지 않습니다.

목표 출력:

```text
정상: ok
실패 확인: true
```

실행:

```text
go run ./_practice/21-http-client/repair
```

<details>
<summary>힌트 1</summary>

`resp`가 실제 응답을 가리킨다는 것을 어느 줄에서 확인할 수 있나요?
`defer`는 나중에 호출되지만 그 호출의 수신자·인자는 등록 시점에 평가됩니다.

</details>

<details>
<summary>힌트 2</summary>

전송 오류는 먼저 반환합니다. 이후 본문 닫기를 등록한 뒤 상태와 읽기 결과를 확인합니다.
상태 오류 경로에서도 정리되도록 닫기 등록을 너무 뒤로 미루지 마세요.

</details>

완료 조건: 정상 응답과 생성·전송 실패에서 패닉 없이 약속한 결과를 얻습니다.

## 문제 4. 한도가 있는 다운로드 — 변형

작업 파일: `_practice/21-http-client/limited-download/main.go`

다음 두 기능을 구현합니다.

| 함수 | 계약 |
| --- | --- |
| `readAtMost(reader io.Reader, limit int64) ([]byte, error)` | 한도 0~1024 허용, 정확한 한도까지 성공, 초과 시 `ErrTooLarge` |
| `download(ctx context.Context, client *http.Client, address string, limit int64) ([]byte, error)` | 컨텍스트를 붙인 GET, 200만 허용, 위 함수로 읽고 본문 닫기 |

한도가 음수이거나 1024를 넘으면 오류입니다.
초과를 판정할 때 결과를 잘라 성공으로 반환하지 않습니다.
`ErrTooLarge`는 `errors.New`로 선언한 값이며 호출자는 `errors.Is`로 구분합니다.
읽는 도중의 오류는 원인을 잃지 않고 전달합니다.

실험 서버는 `hello`를 반환합니다.
먼저 한도 4의 실패를 확인하고, 같은 클라이언트로 한도 5의 성공을 확인합니다.

```text
한도 초과: true
본문: hello
```

실행:

```text
go run ./_practice/21-http-client/limited-download
```

| 입력 본문 | 한도 | 결과 |
| --- | --- | --- |
| 빈 본문 | 0 | 빈 바이트열, 성공 |
| `x` | 0 | `ErrTooLarge` |
| `abcd` | 4 | 성공 |
| `abcde` | 4 | `ErrTooLarge` |
| `가` | 3 | 성공 |
| `가` | 2 | `ErrTooLarge` |

<details>
<summary>힌트 1</summary>

“최대 N바이트 읽기”와 “N바이트를 초과했는지 확인하기”는 다른 일입니다.
N+1번째 바이트를 관찰해야 두 경우를 구분할 수 있습니다.

</details>

<details>
<summary>힌트 2</summary>

`io.LimitReader(reader, limit+1)`로 읽고 길이가 한도보다 큰지 확인합니다.
한도의 범위를 먼저 검사하면 `limit+1` 계산도 정한 작은 범위 안에 있게 됩니다.

</details>

완료 조건: 정확한 한도와 초과를 구분하고, 한글도 바이트 기준으로 판정합니다.

## 문제 5. 학습 JSON 조회 — 응용

작업 파일: `_practice/21-http-client/lesson-json/main.go`

`Lesson`은 `title` 문자열과 `minutes` 정수 필드를 가집니다.
`decodeLesson(io.Reader) (Lesson, error)`와 `fetchLesson(ctx, client, address)`를 분리합니다.
조회 함수의 반환 타입은 `(Lesson, error)`입니다.

다음 계약을 모두 만족하세요.

- HTTP 상태는 200이며 미디어 타입은 `application/json`입니다. 매개변수는 허용합니다.
- 본문 전체가 256바이트 이내여야 합니다.
- 알 수 없는 필드와 JSON 값이 둘 이상인 입력, 뒤의 쓰레기 문자를 거절합니다.
- 제목은 공백만으로 구성될 수 없고 분량은 1~180입니다.
- `null`, 빈 본문, 필수 값이 없는 객체는 오류입니다.
- 상태·형식·해석 오류를 포함한 모든 응답 경로에서 본문을 닫습니다.

실험 서버는 `{"title":"함수","minutes":45}`를 보냅니다.

```text
학습: 함수, 45분
```

실행:

```text
go run ./_practice/21-http-client/lesson-json
```

확인할 입력은 `minutes`의 0·1·180·181, 빈 제목·공백 제목, 알 수 없는 필드, 객체 두 개, 객체 뒤의 공백, 객체 뒤의 `x`입니다.
256바이트는 허용하고 257바이트는 거절해야 합니다.
필드에 허용된 값이라도 본문 전체 크기의 계약을 함께 지켜야 합니다.

<details>
<summary>힌트 1</summary>

작은 크기로 제한한 본문을 얻은 뒤 디코더를 만듭니다.
헤더는 `mime.ParseMediaType`으로 본문 형식과 매개변수를 나누어 확인하세요.

</details>

<details>
<summary>힌트 2</summary>

`DisallowUnknownFields` 뒤 첫 값을 읽고, 두 번째 `Decode` 결과가 `io.EOF`인지 확인합니다.
이후 제목과 분량의 업무 조건을 검사합니다. JSON 문법 검사가 그 조건을 대신하지 않습니다.

</details>

완료 조건: 크기·상태·미디어 타입·문법·필드·값 개수·업무 조건을 따로 설명하고 검사합니다.

## 문제 6. 취소된 상태 조회 — 응용

작업 파일: `_practice/21-http-client/cancel-check/main.go`

`requestStatus(ctx context.Context, client *http.Client, address string) (int, error)`를 작성하세요.
본문이 필요 없는 상태 조회이므로 HEAD 요청을 보냅니다.
응답을 받으면 상태 코드를 반환하고 본문을 닫습니다.
이 함수는 상태 자체를 조사하므로 200만 허용하는 함수와 달리 모든 HTTP 상태를 반환합니다.

호출 전에 취소한 컨텍스트를 전달하면 오류 연결 안에서 `context.Canceled`를 확인할 수 있어야 합니다.
`%w`로 오류를 감싸 원인을 보존합니다.
실험 서버는 요청 수를 `atomic.Int32`로 세고, 요청이 들어오면 204를 반환합니다.

기본 프로그램은 컨텍스트를 취소한 뒤 한 번 호출합니다.

```text
취소 확인: true
서버 요청 수: 0
```

실행:

```text
go run ./_practice/21-http-client/cancel-check
```

취소하지 않은 새 컨텍스트로 호출하면 204와 nil 오류를 받아야 합니다.
추가로 본문의 시작·종료 채널 방식을 적용해 진행 중 취소도 테스트하세요.
고정된 `time.Sleep`으로 요청이 시작되었을 것이라고 추정하지 않습니다.

<details>
<summary>힌트 1</summary>

전달받은 컨텍스트를 요청 생성에 사용해야 합니다.
함수 내부에서 새 `context.Background()`로 바꾸면 호출자가 전달한 취소가 사라집니다.

</details>

<details>
<summary>힌트 2</summary>

`errors.Is(err, context.Canceled)`로 확인합니다.
요청 횟수를 세는 서버와 클라이언트가 서로 다른 고루틴에서 동작하므로 일반 정수의 동시 읽기·쓰기를 피합니다.

</details>

완료 조건: 시작 전 취소에서는 서버에 요청이 도착하지 않고, 진행 중 취소에서는 요청과 핸들러가 모두 끝납니다.

## 선택 A. 리다이렉트를 한 번만 따라가기

작업 파일: `_practice/21-http-client/redirect-once/main.go`

로컬 서버에 `/a → /b → /c` 이동 경로를 만드세요.
`/a`와 `/b`는 각각 302를 반환하고 `/c`는 최종 요청 수를 셉니다.
클라이언트는 첫 이동만 따라가고 두 번째 이동은 `http.ErrUseLastResponse`로 멈춥니다.

목표 출력:

```text
중간 상태: 302
남은 위치: /c
최종 요청 수: 0
```

`CheckRedirect`의 `via`에는 앞서 보낸 요청들이 들어 있습니다.
허용할 이동 횟수와 `len(via)`의 관계를 확인하세요.
받은 중간 응답의 본문도 닫습니다.

```text
go run ./_practice/21-http-client/redirect-once
```

<details>
<summary>힌트</summary>

첫 이동을 판정할 때 이전 요청은 `/a` 하나입니다.
두 번째 이동을 판정할 때는 `/a`와 `/b`가 있습니다.
`ErrUseLastResponse`와 일반 오류를 반환했을 때의 호출자 계약을 구분하세요.

</details>

## 완료 점검과 해설

개인 폴더는 `go test ./...`의 자동 탐색에서 제외됩니다.
각 문제의 개인 경로를 직접 실행하고, 테스트를 작성했다면 해당 폴더를 직접 지정합니다.

```text
go fmt ./_practice/21-http-client/header-greeting
go fmt ./_practice/21-http-client/repair
go fmt ./_practice/21-http-client/limited-download
go fmt ./_practice/21-http-client/lesson-json
go fmt ./_practice/21-http-client/cancel-check
```

- [ ] 서버·클라이언트·응답 본문의 정리 시점을 설명했다.
- [ ] 한도 경계와 JSON 실패 입력을 직접 확인했다.
- [ ] 상태 오류를 전송 오류와 혼동하지 않는다.
- [ ] 취소된 오류의 원인을 보존한다.
- [ ] 외부 서비스 없이 모든 기본 문제를 실행했다.

[21장 해설](../../solutions/21-http-client/README.md) · [본문](../../book/21-http-client/README.md) · [연습문제 목록](../README.md)
