# 00. 개발 환경 준비

| 난이도 | 선수 지식 | 이 장의 결과물 |
| --- | --- | --- |
| 입문 | 폴더를 열고 터미널에 명령을 입력할 수 있으면 됩니다. | Go 예제를 실행하고 저장소 검사를 수행하는 개발 환경 |

## 학습 목표

- 내 컴퓨터 또는 개발 서버에서 Go 명령을 실행한다.
- 저장소 루트, 모듈, 예제 폴더의 관계를 설명한다.
- 예제 실행, 테스트, 코드 검사 명령을 구분한다.
- 오류가 발생했을 때 버전과 현재 작업 폴더부터 확인한다.

## 1. 사용할 환경 정하기

이 교재의 예제는 터미널에서 실행합니다. Windows에서는 PowerShell, macOS·Linux에서는 기본 터미널을 사용할 수 있습니다. VS Code는 권장 편집기이며, 다른 편집기를 사용해도 됩니다.

| 작업 환경 | Go를 설치할 위치 | 명령을 실행할 터미널 |
| --- | --- | --- |
| Windows에서 직접 개발 | Windows | PowerShell |
| Windows + WSL | WSL의 Linux 환경 | WSL 터미널 |
| macOS 또는 Linux에서 개발 | 해당 운영체제 | 터미널 |
| VS Code Remote SSH로 서버에서 개발 | 접속한 개발 서버 | VS Code의 원격 터미널 |

로컬 Windows에 설치한 Go와 원격 Linux 서버에 설치한 Go는 별개입니다. 이 장의 확인 명령은 **실제로 코드를 실행할 환경**에서 입력합니다. 처음에는 환경 하나를 정해 일관되게 사용하세요.

## 2. 준비물과 버전

준비물은 Git, Go, 편집기입니다. Git이 없다면 [공식 Git 설치 안내](https://git-scm.com/install/)에서 운영체제에 맞는 설치 방법을 확인합니다.

이 저장소의 최소 Go 버전은 루트 [go.mod](../../go.mod)에 기록한 **Go 1.27.0**입니다. 2026-11-09에 확인한 공식 안정 버전은 **Go 1.27.3**입니다. 처음 설치할 때는 [공식 다운로드 페이지](https://go.dev/dl/)에서 1.27 계열의 최신 패치를 선택합니다. 버전 확인 날짜와 근거는 [Go 릴리스 기록](https://go.dev/doc/devel/release)에서 다시 확인할 수 있습니다.

기존 개발 환경에 Go가 있다면 먼저 다음 명령을 실행하세요.

```text
go version
git --version
```

예를 들어 `go version go1.27.3 windows/amd64`처럼 버전·운영체제·CPU 종류가 표시됩니다. 출력의 운영체제와 CPU는 환경마다 다르므로 동일한 문자열일 필요는 없습니다.

### Windows

1. [Go 다운로드 페이지](https://go.dev/dl/)에서 자신의 Windows와 CPU에 맞는 설치 파일을 받습니다.
2. 설치 프로그램을 실행하고 안내에 따릅니다.
3. 열려 있던 PowerShell과 VS Code를 닫았다가 다시 엽니다.
4. 새 PowerShell에서 `go version`을 확인합니다.

명령을 찾을 수 없다면 다음으로 실행 파일 위치를 확인합니다.

```powershell
Get-Command go
```

설치 경로를 임의로 추측해 환경 변수를 덮어쓰지 마세요. 설치 프로그램의 경로와 PATH 설정을 확인한 뒤 새 터미널에서 다시 시도합니다. 설치 후 터미널을 다시 여는 이유는 설치 프로그램이 변경한 환경 변수를 새 프로세스에 적용하기 위해서입니다. [공식 설치 안내](https://go.dev/doc/install)를 참고하세요.

### macOS

공식 다운로드 페이지에서 CPU에 맞는 `.pkg` 설치 파일을 받아 실행합니다. 설치가 끝나면 터미널을 다시 열고 `go version`을 확인합니다. 실행 파일 위치는 다음으로 확인할 수 있습니다.

```bash
command -v go
```

### Linux와 WSL

[공식 설치 안내](https://go.dev/doc/install)의 Linux 절차를 따릅니다. 자신의 CPU에 맞는 압축 파일을 선택하고, 설치한 Go의 `bin` 디렉터리를 PATH에 등록합니다. 이미 Go가 설치되어 있다면 설치 방식과 경로부터 확인합니다. 기존 설치 디렉터리 위에 새 압축 파일을 덮어 풀지 않습니다.

```bash
uname -m
command -v go
go version
```

회사에서 관리하는 개발 서버라면 승인된 설치 위치와 설치 절차를 사용합니다. 이 교재를 위해 기존 프로젝트의 Go 설치를 삭제할 필요는 없습니다.

## 3. 저장소 내려받기

작업할 상위 폴더에서 다음 명령을 실행합니다. 아래 명령은 PowerShell과 Bash에서 동일합니다.

```text
git clone https://github.com/praster1/GoLang_all_in_one.git
cd GoLang_all_in_one
go env GOMOD
go env GOVERSION
```

`go env GOMOD`가 방금 내려받은 저장소의 `go.mod` 경로를 가리키는지 확인합니다. 이것이 앞으로 명령을 실행할 **저장소 루트**입니다. 루트에는 `README.md`, `AGENTS.md`, `go.mod`가 있고, 그 아래에 `book/`, `examples/` 등이 있습니다.

이 저장소에는 이미 모듈 설정이 있습니다. 학습을 시작하면서 `go mod init`을 다시 실행하거나 예제마다 `go.mod`를 추가하지 않습니다. 하나의 모듈 안에 여러 예제 패키지를 두므로 루트에서 전체 코드를 검사할 수 있습니다. 모듈이 외부 의존성과 Go 버전을 기록하는 방법은 [Go 모듈 문서](https://go.dev/ref/mod)에서 설명합니다.

`go version`과 `go env GOVERSION`으로 현재 선택된 Go 버전을 확인할 수 있습니다. Go는 설정과 모듈의 버전 요구 사항에 따라 다른 도구 체인을 선택하거나 내려받을 수 있으므로, 명령의 위치에 따라 사용하는 버전이 달라질 수 있습니다. 다운로드가 제한된 환경에서는 요구 버전을 사전에 설치하세요. 자세한 동작은 [Go Toolchains](https://go.dev/doc/toolchain)를 참고합니다.

## 4. VS Code 설정

1. VS Code에서 저장소 **전체 폴더**를 엽니다.
2. 확장 기능에서 **Go**, 식별자 `golang.Go`를 설치합니다.
3. `examples/01-first-program/hello/main.go`를 엽니다.
4. 필요한 Go 도구 설치 안내가 나타나면 개발 환경의 정책에 맞게 설치합니다.
5. VS Code 터미널에서 `go version`과 `go env GOMOD`를 다시 확인합니다.

공식 Go 확장은 `gopls`라는 언어 서버를 이용해 코드 탐색과 편집을 지원합니다. 도구를 관리할 때는 명령 팔레트의 `Go: Install/Update Tools`를 사용할 수 있습니다. 원격 개발 중이라면 원격 환경에 확장이 활성화되었는지도 확인합니다. 확장 기능의 세부 동작은 [공식 VS Code Go 문서](https://github.com/golang/vscode-go)를 기준으로 합니다.

## 5. 첫 실행

저장소 루트에서 실행합니다.

```text
go run ./examples/01-first-program/hello
```

예상 출력:

```text
안녕하세요, Go!
```

이 명령은 해당 폴더의 Go 프로그램을 빌드한 뒤 실행합니다. 현재 예제는 Go에 포함된 표준 라이브러리만 사용하므로 별도의 외부 패키지를 추가할 필요가 없습니다.

출력이 확인되었다면 다음 명령을 실행합니다.

```text
go test ./...
go vet ./...
go run ./tools/checkdocs
```

| 명령 | 확인하는 것 | 성공했을 때 읽는 방법 |
| --- | --- | --- |
| `go test ./...` | 저장소의 Go 패키지와 테스트 | `ok`, 테스트 없는 패키지의 안내 등이 출력됩니다. |
| `go vet ./...` | 의심스러운 Go 코드 패턴 | 일반적으로 문제가 없으면 별도 출력이 없습니다. |
| `go run ./tools/checkdocs` | 이 저장소가 정한 문서 검사 | 출력과 종료 상태를 확인합니다. |

`./...`는 현재 폴더 아래의 패키지를 대상으로 하는 Go 명령의 패턴입니다. 모든 검사가 성공해도 프로그램의 모든 동작이 자동으로 증명되는 것은 아닙니다. 학습자가 예상 출력과 설명을 함께 확인하는 과정도 필요합니다. Go 명령의 의미는 [공식 명령 문서](https://pkg.go.dev/cmd/go)를 참고하세요.

## 6. 코드 형식 정리

Go 코드를 수정한 후에는 루트에서 다음 명령을 실행합니다.

```text
go fmt ./...
```

이 명령은 **파일을 실제로 수정**합니다. 수정한 내용을 `git diff`로 확인하세요. `go fmt`를 쓰면 PowerShell과 Bash에서 같은 명령으로 Go 코드 형식을 정리할 수 있습니다. 개인 연습 폴더에 대해서는 [연습문제 안내](../../exercises/README.md)의 명령을 사용합니다.

## 7. 자주 만나는 문제

| 증상 | 먼저 확인할 것 | 해결 방향 |
| --- | --- | --- |
| `go`를 찾을 수 없음 | 설치 여부, `Get-Command go` 또는 `command -v go` | PATH와 설치 경로를 확인하고 터미널을 다시 엽니다. |
| Go 버전이 너무 낮다는 안내 | `go version`, `go env GOVERSION`, `go.mod` | 프로젝트가 요구하는 Go 버전을 사용합니다. |
| `go.mod`를 찾을 수 없음 | 현재 폴더, `go env GOMOD` | 저장소 루트로 이동합니다. |
| `no Go files` | 실행한 경로 | 루트가 아니라 `./examples/01-first-program/hello`를 실행 대상으로 지정합니다. |
| VS Code 오류와 터미널 결과가 다름 | 로컬·WSL·원격 환경, Go 실행 경로 | 편집기와 터미널이 같은 환경을 사용하는지 확인합니다. |
| 한글이 깨져 보임 | 파일 인코딩과 터미널 표시 | 파일을 UTF-8로 저장하고 UTF-8을 지원하는 터미널에서 확인합니다. |

오류를 질문할 때는 실행한 명령, 오류 전문, `go version`, 운영체제, 현재 폴더를 함께 적으면 원인을 좁히기 쉽습니다. 계정 비밀번호나 토큰은 포함하지 않습니다.

## 8. 직접 확인하기

1. 터미널을 새로 열고 저장소 루트로 이동해 예제를 다시 실행하세요.
2. `go env GOMOD`의 출력이 무엇을 의미하는지 한 문장으로 설명하세요.
3. Go를 Windows와 원격 Linux 양쪽에 설치했다면 지금 실행 중인 환경이 어느 쪽인지 확인하세요.

### 완료 기준

- [ ] Go 버전과 실행 경로를 확인했다.
- [ ] 저장소 루트의 `go.mod`를 확인했다.
- [ ] `안녕하세요, Go!`를 출력했다.
- [ ] 테스트, 코드 검사, 문서 검사를 실행하고 결과를 확인했다.

## 정리

이제 예제를 읽고 수정하고 실행할 준비가 되었습니다. 다음 장에서는 방금 실행한 몇 줄의 코드가 어떤 역할을 하는지 살펴봅니다.

## 참고 자료

- [Download and install](https://go.dev/doc/install) — 운영체제별 공식 설치 절차
- [All releases](https://go.dev/dl/) / [Release history](https://go.dev/doc/devel/release) — 설치 파일과 릴리스 기록
- [Go modules reference](https://go.dev/ref/mod) — 모듈과 `go.mod`
- [Go Toolchains](https://go.dev/doc/toolchain) — 사용하는 Go 버전의 선택
- [Go for Visual Studio Code](https://github.com/golang/vscode-go) — 공식 편집기 확장
- [Command go](https://pkg.go.dev/cmd/go) — 실행·빌드·검사 명령

확인일: 2026-11-09. 버전과 설치 화면은 바뀔 수 있으므로 해당 공식 문서를 함께 확인합니다.

| 이전 | 목차 | 다음 |
| --- | --- | --- |
| 시작 장 | [전체 목차](../README.md) | [01. 첫 Go 프로그램](../01-first-program/README.md) |
