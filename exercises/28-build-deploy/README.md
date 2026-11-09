# 28장 연습문제 — 빌드와 배포 자동화

[본문](../../book/28-build-deploy/README.md)의 검사·빌드·실행·보관 과정을 적용합니다.
빌드할 수 있다는 사실과 실제 대상에서 동작한다는 사실을 구분하여 기록하세요.
실행할 환경이 없는 검사는 미실행으로 남기고 확인한 범위를 설명합니다.

## 문제 구성

| 문제 | 수준 | 결과물 |
| --- | --- | --- |
| 1 | 확인 | 단계별 검증 범위 표 |
| 2 | 변형 | 크로스 빌드와 환경 복원 기록 |
| 3 | 변형 | 산출물 이름 검증 함수 |
| 4 | 응용 | 파일명·크기·SHA-256 목록 CLI |
| 5 | 응용 | 실행 파일에 포함한 기본 설정과 검증 |
| 6 | 응용 | 상태 확인 응답 계약과 테스트 |
| A | 선택 | 자동화 실패 원인 찾기와 개선 제안 |

## 개인 파일 준비

저장소 루트에서 개인 작업 디렉터리를 만듭니다.
기존 코드가 있으면 보존합니다.

PowerShell:

```powershell
$names = @("metadata", "manifest", "embedded-config", "health-contract")
foreach ($name in $names) {
    $dir = "_practice/28-build-deploy/$name"
    New-Item -ItemType Directory -Force $dir | Out-Null
    if (-not (Test-Path "$dir/main.go")) {
        Copy-Item examples/01-first-program/hello/main.go "$dir/main.go"
    }
}
```

Bash:

```sh
for name in metadata manifest embedded-config health-contract; do
    dir="_practice/28-build-deploy/$name"
    mkdir -p "$dir"
    if [ ! -e "$dir/main.go" ]; then
        cp examples/01-first-program/hello/main.go "$dir/main.go"
    fi
done
```

문제 5의 JSON 파일과 각 `main_test.go`는 직접 작성합니다.
실행 파일을 같은 이름으로 다시 빌드하면 기존 산출물을 교체한다는 점을 기억하세요.

## 문제 1. 어디까지 검증했는가

다음 상황마다 확인한 것과 아직 확인하지 못한 것을 적습니다.

1. Linux에서 Windows amd64용 실행 파일을 빌드했습니다.
2. Linux에서 현재 환경용 실행 파일을 만들고 인사 한 번을 출력했습니다.
3. 일반 테스트가 모두 통과했습니다.
4. 컨테이너 이미지 빌드가 성공했지만 실행하지는 않았습니다.
5. 내려받은 파일의 SHA-256이 제공된 목록과 같습니다.
6. `/healthz`가 200을 반환하지만 응답 본문은 읽지 않았습니다.

### 완료 조건

“성공”이라는 표현의 대상을 분명히 해야 합니다.
예를 들어 1번을 Windows의 실제 실행 검증이라고 쓰지 않습니다.
5번에서는 파일 내용 비교와 출처 신뢰를 구분합니다.

<details>
<summary>힌트</summary>

검사한 파일·입력·환경·시간이 무엇인지 각각 적어 보세요.
각 검사에는 그 바깥의 조건이 남아 있습니다.

</details>

## 문제 2. 크로스 빌드 후 환경 복원

본문의 CLI를 자신의 현재 환경과 다른 OS 대상으로 빌드합니다.
대상은 Windows amd64 또는 Linux amd64 중에서 선택하세요.

### 요구사항

1. 실행 전 `go env GOOS GOARCH CGO_ENABLED`를 기록합니다.
2. 명령 한 번 또는 `try/finally` 범위 안에서 대상 환경 변수를 설정합니다.
3. 파일명에 OS·아키텍처를 포함합니다.
4. 명령 뒤 같은 환경 조회를 실행해 원래 설정이 유지되었는지 확인합니다.
5. 현재 OS용 테스트를 실행합니다.
6. 크로스 빌드 파일을 실제 대상에서 실행했는지 별도로 기록합니다.

### 확인 질문

- `go env -w`로 전역 기본값을 바꾸지 않은 이유는 무엇인가요?
- 다른 OS를 지정한 상태에서 `go test`를 실행하면 어떤 문제가 생길 수 있나요?
- 파일명에 `.exe`만 붙이면 Windows 프로그램으로 바뀌나요?
- 모든 의존성이 CGO 없이 빌드된다고 가정해도 되나요?

### 완료 조건

원래 환경 값을 복원하거나 명령 범위로 제한해야 합니다.
빌드 성공과 실제 실행 확인을 각각 보고합니다.
다른 사람의 전역 설정이나 파일을 바꾸지 않습니다.

<details>
<summary>힌트</summary>

Bash는 명령 앞 환경 변수 접두어를,
PowerShell은 기존 값을 저장한 뒤 `finally`에서 복원하는 방식을 사용할 수 있습니다.

</details>

## 문제 3. 일관된 산출물 이름 만들기

다음 자료형과 함수를 작성합니다.

```go
type release struct {
    Name, Version, OS, Arch string
}

func artifactName(r release) (string, error)
```

### 입력 계약

- Name과 Version은 1~64개의 ASCII 영문자·숫자·점·밑줄·하이픈입니다.
- 첫 문자는 영문자 또는 숫자여야 합니다.
- OS는 linux, windows, darwin 중 하나입니다.
- Arch는 amd64, arm64 중 하나입니다.
- 그 외 입력은 오류입니다.

### 이름 규칙

`이름-버전-OS-아키텍처`를 사용합니다.
Windows에는 끝에 `.exe`를 붙입니다.
이 함수는 이름만 만들며 실행 파일을 빌드하거나 경로에 쓰지 않습니다.

| 입력 요약 | 결과 |
| --- | --- |
| hello / 1.0.0 / linux / amd64 | `hello-1.0.0-linux-amd64` |
| hello / dev-1 / windows / arm64 | `hello-dev-1-windows-arm64.exe` |
| 이름 `../bad` | 오류 |
| 버전 `two words` | 오류 |
| OS `other` | 오류 |
| Arch `386` | 오류 |

`main`은 hello·1.0.0·windows·amd64를 사용합니다.

```text
hello-1.0.0-windows-amd64.exe
```

### 완료 조건

허용한 조합을 올바르게 만들고, 경로 구분자나 공백이 들어간 라벨은 거부합니다.
이름 검증을 실제 파일의 내용·출처 검증으로 혼동하지 않습니다.

<details>
<summary>힌트</summary>

정규식 또는 문자 순회로 라벨을 검사할 수 있습니다.
확장자 선택과 대상 목록 검증은 별도로 표현하면 읽기 쉽습니다.

</details>

## 문제 4. 산출물 목록 CLI

`-dir` 플래그로 지정한 폴더의 일반 파일을 읽어 JSON 목록을 출력합니다.
다음 필드를 사용합니다.

```go
type entry struct {
    Name   string `json:"name"`
    Size   int64  `json:"size"`
    SHA256 string `json:"sha256"`
}
```

### 처리 계약

- `-dir`이 없거나 남은 위치 인자가 있으면 종료 코드 2.
- 파일·디렉터리 읽기 또는 출력 실패는 종료 코드 1.
- 정상 출력은 종료 코드 0.
- 파일명 오름차순으로 반환합니다.
- 하위 디렉터리는 재귀 탐색하지 않습니다.
- 기존 `manifest.json`은 결과에서 제외합니다.
- 심볼릭 링크 등 일반 파일이 아닌 항목은 오류로 처리합니다.
- 빈 디렉터리의 JSON 결과는 `[]`입니다.
- 크기는 실제 읽은 바이트 수이고 SHA-256은 소문자 16진수 문자열입니다.

빌드가 끝나 더 이상 수정되지 않는 산출물 폴더를 대상으로 합니다.
파일을 다른 프로세스가 동시에 바꾸는 상황의 스냅샷 보장은 요구하지 않습니다.

### 테스트 자료

테스트의 임시 디렉터리에 다음을 만듭니다.

| 이름 | 내용 |
| --- | --- |
| a.bin | 빈 바이트열 |
| b.bin | 줄바꿈 없는 abc |
| manifest.json | 이전 목록을 나타내는 임의 문자열 |
| nested/ | 빈 하위 디렉터리 |

결과 항목은 a.bin, b.bin 두 개입니다.
b.bin의 크기는 3이고 SHA-256은 다음과 같습니다.

```text
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
```

### 실행

실제 산출물 폴더를 먼저 만든 뒤 실행합니다.
다음 명령은 `dist/`가 존재한다고 가정합니다.

```sh
go run ./_practice/28-build-deploy/manifest -dir dist
```

### 완료 조건

- 파일을 스트림으로 읽어 큰 입력을 전부 메모리에 담지 않습니다.
- 읽은 파일을 작업마다 닫습니다.
- 읽기 오류와 JSON 출력 오류를 호출자에게 전달합니다.
- 테스트는 임시 디렉터리를 사용합니다.
- 자신의 출력 파일을 목록에 다시 넣지 않습니다.

<details>
<summary>힌트 1</summary>

`os.ReadDir`의 정렬된 항목에서 디렉터리와 manifest.json을 건너뜁니다.
파일 하나를 해시하는 함수를 분리하면 Close의 수명을 짧게 유지하기 쉽습니다.

</details>

<details>
<summary>힌트 2</summary>

`sha256.New`, `io.Copy`, `hash.Sum(nil)`을 사용할 수 있습니다.
`json.Encoder.SetIndent`로 읽기 좋은 출력을 만들 수 있습니다.

</details>

## 문제 5. 기본 설정을 실행 파일에 포함

`defaults.json` 파일을 만들고 `go:embed`로 포함합니다.

```json
{"name":"demo","port":8080}
```

다음 자료형과 파서를 작성합니다.

```go
type config struct {
    Name string `json:"name"`
    Port int    `json:"port"`
}

func parseConfig(text string) (config, error)
```

### 계약

- JSON 값은 하나만 허용합니다.
- 알 수 없는 필드는 거부합니다.
- 이름은 공백 제거 후 비어 있으면 안 됩니다.
- 포트는 1~65535입니다.
- 잘못된 JSON, null, 빈 객체, 범위 밖 값은 오류입니다.

이 문제는 `encoding/json` v1을 사용합니다.
`DisallowUnknownFields`가 중복 키 금지나 필드 대소문자 엄격성을 모두 제공한다고 가정하지 않습니다.
그런 계약이 필요하다면 별도 구현과 테스트가 필요합니다.

`main`의 출력:

```text
demo:8080
```

### 테스트 요구사항

임시 작업 디렉터리로 이동한 뒤에도 임베드 기본값을 읽을 수 있어야 합니다.
`{}`, `null`, 포트 0, 알 수 없는 필드, JSON 값 두 개를 거부하는지 검사합니다.

### 완료 조건

- 실행 시 로컬 defaults.json의 존재에 의존하지 않습니다.
- 설정 값에 대한 검증은 임베드 여부와 관계없이 실행합니다.
- 비밀값을 기본 파일에 넣지 않습니다.
- 정적 기본값과 운영 환경별 설정의 역할을 설명합니다.

<details>
<summary>힌트</summary>

두 번째 Decode가 EOF인지 확인하면 추가 JSON 값이나 잘못된 꼬리를 구분할 수 있습니다.
처음 Decode의 성공만으로 입력 전체를 처리했다고 판단하지 않습니다.

</details>

## 문제 6. 상태 확인 응답 계약

다음 함수를 작성합니다.

```go
func newHandler() http.Handler
```

`GET /healthz`가 상태 200과 `Content-Type: application/json`을 반환하도록 합니다.
응답 JSON은 `{"status":"ok"}`입니다.
`POST /healthz`는 405, 존재하지 않는 경로는 404를 반환합니다.

`main`에서는 `httptest.NewRecorder`와 `NewRequest`로 GET 요청을 한 번 처리합니다.
계속 실행되는 서버를 만들 필요는 없습니다.

```text
200 {"status":"ok"}
```

### 테스트 요구사항

- GET의 상태·미디어 타입·JSON 필드.
- POST의 405.
- 없는 경로의 404.
- 여러 요청을 같은 핸들러에 전달해도 결과가 안정적임.

### 완료 조건

200만 확인하는 검사와 약속한 응답을 확인하는 검사의 차이를 설명합니다.
DB가 없는 이 핸들러를 모든 서비스 의존성 준비 검사로 소개하지 않습니다.
GET 패턴이 HEAD에도 대응하는 표준 ServeMux 규칙을 유지합니다.

<details>
<summary>힌트</summary>

`http.NewServeMux`의 `GET /healthz` 패턴을 이용할 수 있습니다.
정상 응답에서 JSON을 디코딩해 status 값까지 확인하세요.

</details>

## 선택 A. 자동화 점검

저장소의 `build-artifacts.yml`을 읽고 다음을 찾아 설명합니다.

1. 어떤 변경에서 자동 실행되는가?
2. 어느 단계가 테스트하고 어느 단계가 실제 실행 파일을 실행하는가?
3. 산출물 이름은 OS를 어떻게 구분하는가?
4. manifest 파일은 언제 계산되고 어디로 옮겨지는가?
5. 보관 기간은 얼마인가?
6. 컨테이너 빌드와 실행은 각각 어느 단계인가?
7. 레지스트리 업로드나 운영 배포를 수행하는가?

자신의 변형 워크플로를 제안한다면
필요한 권한·실패 조건·테스트한 파일과 배포한 파일의 관계를 함께 설명합니다.
실제 계정이나 운영 자격 증명은 문제의 답안에 넣지 않습니다.

## 최종 확인

```sh
go fmt ./_practice/28-build-deploy/...
go vet ./_practice/28-build-deploy/...
go test -count=1 ./_practice/28-build-deploy/...
```

- [ ] 테스트·빌드·실행·배포의 범위를 구분했습니다.
- [ ] 현재 셸의 환경 설정을 보존했습니다.
- [ ] 출력과 종료 코드를 검사했습니다.
- [ ] 파일명과 내용 체크섬을 함께 기록했습니다.
- [ ] 임베드 자원과 실행 설정을 구분했습니다.
- [ ] 현재 실행하지 못한 환경을 정확히 남겼습니다.

## 참고 자료

[go 명령](https://pkg.go.dev/cmd/go) · [embed](https://pkg.go.dev/embed) · [SHA-256](https://pkg.go.dev/crypto/sha256) · [GitHub workflow artifacts](https://docs.github.com/en/actions/concepts/workflows-and-actions/workflow-artifacts)

[본문](../../book/28-build-deploy/README.md) · [해설 확인](../../solutions/28-build-deploy/README.md)
