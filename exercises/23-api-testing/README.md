# 23장 연습문제: API 설계와 핸들러 테스트

선수 학습: [23장 본문](../../book/23-api-testing/README.md)

이번 문제는 구현과 테스트를 함께 작성합니다.
기본 여섯 문제를 마친 뒤 선택 A에서 경로·메서드 오류까지 JSON으로 통일합니다.
모든 명령은 **저장소 루트**에서 실행하며 외부 네트워크나 데이터베이스는 필요하지 않습니다.

## 작업 폴더와 준비

| 문제 | 수준 | 개인 폴더 |
| --- | --- | --- |
| 1. 테스트가 놓치는 실패 | 확인 | 학습 노트에 기록 |
| 2. 짧은 메시지 API | 변형 | `_practice/23-api-testing/echo-contract/` |
| 3. 기록기 재사용 고치기 | 오류 진단 | `_practice/23-api-testing/repair/` |
| 4. 필수 점수 입력 | 응용 | `_practice/23-api-testing/strict-score/` |
| 5. 저장소 실패 처리 | 응용 | `_practice/23-api-testing/failing-repository/` |
| 6. 요청 ID 미들웨어 | 응용 | `_practice/23-api-testing/request-middleware/` |
| 선택 A. 전체 오류 형식 | 선택 | `_practice/23-api-testing/uniform-errors/` |

테이블 테스트 예제의 `main.go`를 출발점으로 복사합니다.
문제에 맞게 핸들러와 실행 요청을 바꾸고, 자신이 작성할 테스트는 같은 폴더의 `behavior_test.go`에 둡니다.
기존 개인 파일을 먼저 확인하고 보존합니다. 새 모듈은 만들지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("echo-contract", "repair", "strict-score", "failing-repository", "request-middleware", "uniform-errors")) {
    $practiceDir = Join-Path "_practice/23-api-testing" $practiceName
    New-Item -ItemType Directory -Force $practiceDir | Out-Null
    $practiceMain = Join-Path $practiceDir "main.go"
    if (Test-Path -LiteralPath $practiceMain) {
        Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
    } else {
        Copy-Item examples/23-api-testing/handler-table/main.go $practiceMain
    }
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
for practice_name in echo-contract repair strict-score failing-repository request-middleware uniform-errors; do
    practice_dir="_practice/23-api-testing/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/23-api-testing/handler-table/main.go "$practice_dir/main.go"
    fi
done
```

## 문제 1. 상태 200만 확인하면? — 확인

정상 계약은 상태 200, `Content-Type: application/json`, 본문 `{"message":"Go"}`입니다.
아래 네 응답을 상태 코드만 검사하는 테스트에 넣었다고 생각해 보세요.

| 사례 | 상태 | Content-Type | 본문 |
| --- | --- | --- | --- |
| A | 200 | application/json | `{"message":"Go"}` |
| B | 200 | application/json | `{"message":"wrong"}` |
| C | 200 | text/plain | `not json` |
| D | 400 | application/json | `{"code":"invalid_message"}` |

1. 상태만 검사할 때 통과하는 사례를 고릅니다.
2. 전체 계약을 만족하는 사례를 고릅니다.
3. B를 발견할 assertion과 C를 발견할 assertion을 적습니다.
4. JSON 필드 순서가 달라졌을 때도 필드 의미를 비교하려면 어떻게 해야 하나요?

<details>
<summary>힌트 1</summary>

200이라는 사실은 응답의 필드 값과 형식을 보장하지 않습니다.
상태, 헤더와 JSON 해석 결과를 별도의 관찰값으로 생각하세요.

</details>

<details>
<summary>힌트 2</summary>

`json.Unmarshal`로 응답을 구조체에 읽은 뒤 `Message`를 비교할 수 있습니다.
헤더는 `Header().Get("Content-Type")`로 확인합니다.

</details>

완료 조건: 테스트가 통과한 것과 계약이 맞는 것을 구분해서 설명합니다.

## 문제 2. 짧은 메시지 API와 테이블 테스트 — 변형

작업 파일: `_practice/23-api-testing/echo-contract/main.go`
테스트 파일: 같은 폴더의 `behavior_test.go`

`GET /echo?message=...`를 구현하세요.
메시지의 앞뒤 공백을 제거한 뒤 1~8룬을 허용합니다.
성공은 200과 `{"message":"정규화한 값"}`입니다.
실패는 400과 다음 JSON입니다.

```json
{"code":"invalid_message","message":"메시지는 1~8글자입니다"}
```

성공·실패 모두 `Content-Type: application/json`이며 본문 뒤 줄바꿈을 붙입니다.
기본 프로그램은 `Go`와 빈 메시지를 보내 다음을 출력합니다.

```text
200 {"message":"Go"}
400 {"code":"invalid_message","message":"메시지는 1~8글자입니다"}
```

```text
go run ./_practice/23-api-testing/echo-contract
go test -v ./_practice/23-api-testing/echo-contract
```

테이블에는 다음 입력을 포함하세요.

| 입력 | 기대 |
| --- | --- |
| `Go` | 메시지 Go |
| ` Go ` | 메시지 Go |
| 한글 8룬 | 성공 |
| 빈 문자열 | invalid_message |
| 공백만 있는 문자열 | invalid_message |
| 한글 9룬 | invalid_message |

상태만 검사하지 말고 Content-Type, JSON 문법, 메시지와 오류 코드를 확인합니다.
각 사례에서 새로운 요청과 recorder를 만들며, 라우터는 여러 요청에 재사용합니다.

<details>
<summary>힌트 1</summary>

계약의 길이는 UTF-8 바이트 수가 아니라 룬 수입니다.
`strings.TrimSpace` 뒤 `utf8.RuneCountInString`으로 검사합니다.

</details>

<details>
<summary>힌트 2</summary>

테스트의 검색어도 `url.QueryEscape`로 구성하여 URL 문법과 입력 데이터를 분리합니다.
기대 메시지는 계약 표에 직접 적고, 핸들러의 조건문을 다시 복제해 계산하지 않습니다.

</details>

완료 조건: 잘못된 메시지를 200으로 반환하거나 본문만 바꾸면 테스트가 실패합니다.

## 문제 3. 앞선 응답이 다음 테스트에 남는다 — 오류 진단

작업 파일: `_practice/23-api-testing/repair/main.go`
테스트 파일: 같은 폴더의 `behavior_test.go`

준비 단계에서 복사한 `GET /greet` 구현은 그대로 사용합니다.
아래는 **의도적으로 잘못 구성한 전체 테스트 파일**입니다.
개인 테스트 파일이 이미 있다면 먼저 보존한 뒤 별도 파일·폴더에서 비교하세요.

```go
package main

import (
    "net/http/httptest"
    "testing"
)

func TestGreetingResponses(t *testing.T) {
    handler := newHandler()
    recorder := httptest.NewRecorder()
    for _, tc := range []struct {
        path   string
        status int
    }{
        {"/greet?name=Go", 200},
        {"/greet", 400},
    } {
        request := httptest.NewRequest("GET", tc.path, nil)
        handler.ServeHTTP(recorder, request)
        if recorder.Code != tc.status {
            t.Errorf("path=%s status=%d want=%d", tc.path, recorder.Code, tc.status)
        }
    }
}
```

```text
go test -v ./_practice/23-api-testing/repair
```

두 번째 사례에서 이전 응답의 상태·본문이 남는 원인을 찾고 고치세요.
프로덕션 핸들러를 바꾸어 이 잘못된 테스트에 맞추지 않습니다.
수정한 테스트는 각 사례의 상태뿐 아니라 JSON 메시지·오류 코드도 검사해야 합니다.

<details>
<summary>힌트 1</summary>

기록기는 한 번의 응답을 담는 상태를 가집니다.
현재 어디에서 만들고 몇 번 사용하는지 확인하세요.

</details>

<details>
<summary>힌트 2</summary>

recorder를 각 사례 안에서 새로 만듭니다.
요청 본문도 소비되는 스트림이므로 request 역시 각 사례에서 새로 준비합니다.

</details>

완료 조건: 잘못된 버전의 실패를 확인한 뒤 수정 버전이 통과하고, 이전 JSON이 다음 응답에 붙지 않습니다.

## 문제 4. 필수 점수 입력 — 응용

작업 파일: `_practice/23-api-testing/strict-score/main.go`

`POST /score`는 JSON의 필수 `score` 필드를 받아 판정합니다.
`score`는 0~100의 정수입니다. 명시적 0은 유효하지만 필드가 없거나 null이면 실패입니다.
요청 구조체에서 `*int`로 존재 여부를 표현하세요.

| 점수 | 성공 JSON |
| --- | --- |
| 0~59 | `{"label":"다시 연습"}` |
| 60~89 | `{"label":"통과"}` |
| 90~100 | `{"label":"우수"}` |

성공 상태는 200입니다.
실패는 JSON의 `code`, `message` 필드로 표현합니다.

| 실패 | 상태 | code |
| --- | --- | --- |
| JSON이 아닌 미디어 타입 | 415 | unsupported_media_type |
| 전체 본문 64바이트 초과 | 413 | body_too_large |
| 잘못된 JSON·타입·알 수 없는 필드·추가 값 | 400 | invalid_json |
| score 없음·null | 422 | missing_score |
| score가 범위 밖 | 422 | invalid_score |

기본 요청은 `{"score":0}`, `{"score":72}`, `{}`, `{"score":null}`입니다.

```text
200 {"label":"다시 연습"}
200 {"label":"통과"}
422 {"code":"missing_score","message":"score가 필요합니다"}
422 {"code":"missing_score","message":"score가 필요합니다"}
```

```text
go run ./_practice/23-api-testing/strict-score
go test -v ./_practice/23-api-testing/strict-score
```

점수 -1·0·59·60·89·90·100·101과 누락·null·문자열 타입을 테스트합니다.
본문 64바이트와 65바이트, 객체 두 개와 뒤의 쓰레기 문자도 확인합니다.
응답 헤더와 JSON 오류 코드를 함께 검사하세요.

이 문제는 JSON v1의 알 수 없는 필드 제한을 사용합니다.
중복 키 거절이나 대소문자까지 정확히 일치하는 별도 파서를 요구하지 않습니다.

<details>
<summary>힌트 1</summary>

nil 검사를 먼저 한 뒤 점수를 역참조합니다.
0이라는 숫자만으로 누락을 판정하면 유효한 0점이 잘못 거절됩니다.

</details>

<details>
<summary>힌트 2</summary>

HTTP 검증을 통과한 0~100의 정수는 `classifyScore` 같은 순수 함수에서 판정할 수 있습니다.
입력 읽기 오류와 점수 범위를 HTTP 오류로 바꾸는 일은 핸들러에 둡니다.

</details>

완료 조건: 0과 누락을 구분하고 정상·경계·오류 상태를 테스트로 확인합니다.

## 문제 5. 저장소 실패를 HTTP 계약으로 바꾸기 — 응용

작업 파일: `_practice/23-api-testing/failing-repository/main.go`

본문의 `LessonStore` 인터페이스를 받는 조회 핸들러를 작성하세요.
`GET /lessons/{id}`의 계약은 다음과 같습니다.

| 조건 | 저장소 호출 | 상태 | 응답 |
| --- | --- | --- | --- |
| ID 변환 실패 또는 1 미만 | 0회 | 400 | invalid_id |
| 데이터 있음 | 1회 | 200 | id·title |
| ErrNotFound 또는 그것을 감싼 오류 | 1회 | 404 | not_found |
| 다른 저장소 오류 | 1회 | 500 | internal_error |

요청의 컨텍스트를 저장소에 전달합니다.
내부 오류의 세부 문자열을 JSON 메시지로 그대로 내보내지 않습니다.
테스트에서는 `storeFunc` 어댑터로 결과·오류·호출 횟수·전달된 ID와 컨텍스트를 관찰합니다.

기본 메모리 구현은 ID 1의 Go 데이터만 가지며 ID 3에서 학습용 내부 오류를 반환합니다.
순서대로 1, 2, 3, x를 요청합니다.

```text
200 {"id":1,"title":"Go"}
404 {"code":"not_found","message":"학습을 찾을 수 없습니다"}
500 {"code":"internal_error","message":"요청을 처리하지 못했습니다"}
400 {"code":"invalid_id","message":"ID는 양수 정수입니다"}
```

```text
go run ./_practice/23-api-testing/failing-repository
go test -v ./_practice/23-api-testing/failing-repository
```

테스트용 저장소는 HTTP 응답을 직접 쓰면 안 됩니다.
실제 핸들러가 저장소 결과를 상태와 JSON으로 바꾸는 경계를 검사해야 합니다.

<details>
<summary>힌트 1</summary>

인터페이스는 조회 결과와 error만 반환합니다.
존재하지 않는 데이터는 sentinel 오류로 표현하고 `errors.Is`로 분류합니다.

</details>

<details>
<summary>힌트 2</summary>

잘못된 ID에서는 저장소 구현의 호출 횟수가 0인지 확인합니다.
컨텍스트에 테스트용 값을 넣어 전달 여부를 확인하되 개인 정보나 운영 설정은 넣지 않습니다.

</details>

완료 조건: 실제 장애를 만들지 않고 404·500과 부작용 없는 400 경로를 재현합니다.

## 문제 6. 요청 ID 미들웨어 — 응용

작업 파일: `_practice/23-api-testing/request-middleware/main.go`

`withRequestID(next http.Handler) http.Handler`를 작성합니다.
미들웨어를 만들 때 로컬 원자 카운터를 준비하고 요청마다 `r-1`, `r-2`처럼 ID를 생성합니다.
응답의 `X-Request-ID`와 private 타입 키의 컨텍스트 값이 같아야 합니다.
외부에서 보낸 X-Request-ID는 사용하지 않습니다.

`GET /ready`는 컨텍스트의 ID와 `ready: true`를 JSON으로 돌려줍니다.
미들웨어를 라우터 바깥에 두어 알 수 없는 경로의 404에도 ID 헤더가 있어야 합니다.
404 본문 자체는 이 문제에서 라우터의 기본 텍스트를 유지합니다.

기본 프로그램은 `/ready`, `/ready`, `/missing`을 요청합니다.

```text
200 {"request_id":"r-1","ready":true} id=r-1
200 {"request_id":"r-2","ready":true} id=r-2
404 404 page not found id=r-3
```

```text
go run ./_practice/23-api-testing/request-middleware
go test -race ./_practice/23-api-testing/request-middleware
```

다음 계약을 검사하세요.

- 다음 핸들러를 정확히 한 번 호출한다.
- 헤더와 컨텍스트에 같은 ID가 있다.
- 클라이언트 제공 ID를 채택하지 않는다.
- 50개 동시 요청의 ID가 비어 있지 않고 중복되지 않는다.
- 404에도 ID 헤더가 있다.

카운터는 전역 변수로 만들지 않고 미들웨어 인스턴스가 소유합니다.
여러 서버 전체의 분산 ID 생성이나 인증을 구현하는 문제는 아닙니다.

<details>
<summary>힌트 1</summary>

같은 이름의 문자열 컨텍스트 키 대신 자체 타입을 사용합니다.
`r.WithContext`로 얻은 요청을 다음 핸들러에 넘깁니다.

</details>

<details>
<summary>힌트 2</summary>

공유 카운터는 원자적으로 증가시킵니다.
동시 테스트에서는 반환 순서가 아니라 ID 집합의 중복 여부를 검사합니다.
각 고루틴에는 별도의 request와 recorder가 필요합니다.

</details>

완료 조건: 값 연결·호출 횟수·적용 범위와 동시성을 모두 확인합니다.

## 선택 A. 경로·메서드 오류도 JSON으로 통일하기

작업 파일: `_practice/23-api-testing/uniform-errors/main.go`

`/health`만 제공하는 작은 API 전체에서 오류 형식을 통일합니다.
등록 경로 밖이면 404의 `not_found`, GET·HEAD 이외 메서드이면 405의 `method_not_allowed`를 보냅니다.
405에는 `Allow: GET, HEAD`를 설정합니다.
성공은 200과 `{"ready":true}`이며 실제 HEAD 응답 본문은 비어 있어야 합니다.

모든 실패 본문은 JSON `code`, `message` 구조입니다.
라우터 기본 텍스트 오류가 남지 않도록 경로·메서드의 처리 범위를 직접 정합니다.
실제 임시 서버로 네 요청을 확인합니다.

```text
GET /health -> 200 {"ready":true}
POST /health -> 405 {"code":"method_not_allowed","message":"GET 또는 HEAD를 사용하세요"}
GET /missing -> 404 {"code":"not_found","message":"경로를 찾을 수 없습니다"}
HEAD /health -> 200 (빈 본문)
```

```text
go run ./_practice/23-api-testing/uniform-errors
```

<details>
<summary>힌트</summary>

경로가 맞는지 먼저 확인하고, 그 경로의 메서드를 검사합니다.
성공 HEAD는 상태·헤더만 정하고 반환할 수 있습니다.
23장 기본 예제의 JSON 오류가 적용되는 범위와 이번 선택 과제의 전체 범위를 비교하세요.

</details>

## 완료 점검

개인 테스트는 일반 `go test ./...`에 자동 포함되지 않으므로 경로를 직접 지정합니다.
의도적인 실패 실험을 끝내면 올바른 테스트로 복구합니다.

```text
go fmt ./_practice/23-api-testing/echo-contract
go fmt ./_practice/23-api-testing/repair
go fmt ./_practice/23-api-testing/strict-score
go fmt ./_practice/23-api-testing/failing-repository
go fmt ./_practice/23-api-testing/request-middleware
```

- [ ] 상태·헤더·JSON 필드를 따로 검사했다.
- [ ] 각 테스트 사례의 관찰 상태를 분리했다.
- [ ] 누락과 명시적 0을 구분했다.
- [ ] 저장소 오류와 전달된 컨텍스트를 재현했다.
- [ ] 동시 요청에서 순서 대신 불변 조건을 검사했다.

[23장 해설](../../solutions/23-api-testing/README.md) · [본문](../../book/23-api-testing/README.md) · [연습문제 목록](../README.md)
