# 00. 개발 환경과 학습 준비

Go 공부를 시작할 때 가장 먼저 익힐 것은 **코드를 어디에 저장하고, 어떤 도구로 실행하며, 결과를 어떻게 확인하는가**입니다. 이 장에서는 개발 환경을 준비한 뒤 저장소의 예제를 실행하고, 자신의 연습 파일을 수정하고, 실행 파일까지 만들어 봅니다.

설정 항목을 모두 외울 필요는 없습니다. 명령이 실패했을 때 **현재 작업 위치, 사용하는 Go, 실행 대상**을 차례로 확인할 수 있으면 다음 장의 학습을 시작할 수 있습니다.

| 항목 | 내용 |
| --- | --- |
| 난이도 | 입문 |
| 선수 지식 | 파일·폴더를 열고 프로그램을 설치할 수 있는 정도. Go 문법은 몰라도 됩니다. |
| 준비물 | Go, Git, 편집기, 최초 설치와 저장소 다운로드를 위한 인터넷 연결 |
| 기준 환경 | Go 1.27 계열. 설치 예시는 2026-11-09에 확인한 Go 1.27.3 |
| 이 장의 결과물 | 첫 예제의 실행 결과, 환경 점검 기록, 직접 빌드한 프로그램 |
| 이어지는 학습 | [01. 첫 Go 프로그램](../01-first-program/README.md) |

## 학습 목표

이 장을 마치면 다음을 할 수 있습니다.

1. 편집기, 터미널, Go 도구 체인, Git이 맡은 역할을 구분한다.
2. Windows·macOS·Linux·WSL·원격 서버 중 자신이 코드를 실행할 환경을 정하고 Go 설치를 확인한다.
3. 저장소 루트와 `go.mod`를 확인하고 올바른 패키지 경로로 예제를 실행한다.
4. 실행·빌드·포맷·테스트·정적 검사 명령의 목적과 결과를 구분한다.
5. 설치 경로, 버전, 작업 폴더가 달라 발생한 오류를 진단한다.

이미 Go를 사용한다면 5절의 환경 확인부터 진행해도 됩니다. 처음 설치한다면 1절부터 읽되, 운영체제별 설치와 편집기 설정은 **자신에게 해당하는 항목 하나**를 선택하세요.

## 1. 개발 도구가 맡은 역할

예를 들어 화면에 인사말을 출력하려면 먼저 코드를 파일에 적고, 그 파일로 프로그램을 만들고, 실행 결과를 확인합니다. 이 과정에서 여러 도구가 등장합니다.

| 용어 | 맡은 역할 | 이 교재에서 보는 예 |
| --- | --- | --- |
| 소스 코드(source code) | 수행할 동작을 표현한 글 | `main.go` 파일 |
| 편집기(editor)·통합 개발 환경(IDE) | 코드를 작성하고 탐색·수정하도록 도움 | VS Code, IntelliJ IDEA, GoLand |
| 터미널(terminal)과 셸(shell) | 명령을 입력하고 결과를 확인하는 환경 | PowerShell, Bash, Zsh |
| Go 도구 체인(toolchain) | Go 코드를 빌드하고 테스트하는 도구 묶음 | `go run`, `go build`, `go test` |
| 실행 파일(executable) | 운영체제에서 실행할 수 있도록 만든 프로그램 | Windows의 `hello.exe`, Linux의 `hello` |
| Git | 파일 변경 이력을 관리하고 저장소를 내려받는 도구 | `git clone`, `git status` |
| GitHub | Git 저장소와 문서를 공유하는 서비스 | 이 교재의 온라인 저장소 |

편집기의 실행 버튼도 내부적으로 특정 Go와 실행 설정을 사용합니다. 따라서 버튼을 누르는 방법과 함께 터미널에서 같은 프로그램을 실행하는 방법을 알아 두면, 문제가 코드에 있는지 편집기 설정에 있는지 구분하기 쉽습니다.

Go 도구의 세부 동작은 [공식 Go 명령 문서](https://pkg.go.dev/cmd/go), 프로그램과 패키지의 관계는 [How to Write Go Code](https://go.dev/doc/code)를 참고합니다.

### 명령 블록을 읽는 방법

이 장의 명령은 한 줄씩 입력하고 Enter를 누르면 됩니다. 앞 명령에서 오류가 발생했다면 메시지를 확인한 뒤 다음 단계로 넘어갑니다. `$`, `PS>` 같은 프롬프트는 명령에 포함하지 않습니다.

명령 위에 **PowerShell**이라고 쓰여 있으면 Windows PowerShell에서, **Bash**라고 쓰여 있으면 Linux·WSL의 Bash 또는 호환되는 셸에서 실행합니다. 별도 구분이 없는 `go`·`git` 명령은 두 환경에서 공통입니다. **예상 출력** 블록은 직접 입력하는 내용이 아닙니다.

## 2. 실제로 코드를 실행할 환경 정하기

같은 노트북에서도 Windows, WSL, SSH로 접속한 서버는 서로 다른 실행 환경입니다. Go와 Git을 설치할 위치도 달라집니다.

| 선택한 환경 | Go·Git을 설치할 곳 | 저장소를 둘 곳 | 명령을 입력할 곳 |
| --- | --- | --- | --- |
| Windows에서 직접 실행 | Windows | Windows의 작업 폴더 | PowerShell |
| macOS에서 실행 | macOS | macOS의 작업 폴더 | 터미널 |
| Linux에서 실행 | Linux | Linux의 작업 폴더 | 터미널 |
| Windows에서 WSL 사용 | WSL의 Linux 환경 | WSL의 작업 폴더 | WSL 터미널 |
| SSH로 Linux 서버에 접속 | 접속한 개발 서버 | 서버의 작업 폴더 | 원격 터미널 |

예를 들어 Windows에서 VS Code를 띄웠더라도 Remote SSH로 서버 폴더를 열었다면 코드는 서버에서 빌드하고 실행합니다. 이때 로컬 Windows에 설치한 Go만으로 서버의 Go 설치가 준비되는 것은 아닙니다. [VS Code Remote SSH 공식 문서](https://code.visualstudio.com/docs/remote/ssh)에서도 원격 환경의 도구와 확장을 구분합니다.

처음에는 환경 하나를 정하고 설치, 저장소 다운로드, 실행을 그 환경에서 일관되게 진행하세요. WSL이나 원격 서버를 새로 구성하는 일 자체는 이 장의 필수 조건이 아닙니다. Windows에서 바로 시작해도 됩니다.

회사에서 관리하는 개발 서버를 사용한다면 지정된 설치 경로와 절차를 따릅니다. 특정 서버 주소나 계정은 교재의 예제에 필요하지 않습니다.

## 3. 설치 전에 운영체제와 CPU 확인하기

Go 설치 파일은 **운영체제와 CPU 아키텍처**에 맞게 골라야 합니다. 파일 이름의 `amd64`는 일반적인 Intel·AMD의 64비트 x86 계열을 뜻하며, AMD 제조사의 CPU에만 쓰는 파일이라는 의미는 아닙니다.

| 실행 환경 | 다운로드 이름에서 확인할 부분 | 일반적인 설치 형식 |
| --- | --- | --- |
| Windows, Intel·AMD 64비트 | `windows-amd64` | `.msi` |
| Windows on ARM | `windows-arm64` | 해당 플랫폼 설치 파일 |
| Apple Silicon Mac | `darwin-arm64` | `.pkg` |
| Intel Mac | `darwin-amd64` | `.pkg` |
| Linux·WSL, x86-64 | `linux-amd64` | `.tar.gz` |
| Linux·WSL, ARM64 | `linux-arm64` | `.tar.gz` |

Windows는 **설정 → 시스템 → 정보 → 시스템 종류**, macOS는 **이 Mac에 관하여**에서 칩 정보를 확인할 수 있습니다. Linux·WSL에서는 다음을 확인합니다.

```bash
uname -m
```

보통 `x86_64`라면 `amd64`, `aarch64`라면 `arm64` 파일을 선택합니다. 에뮬레이션이나 가상 환경을 사용한다면 물리 컴퓨터의 모델명만으로 결정하지 말고, **Go를 실행할 운영체제와 환경**을 기준으로 확인하세요.

지원 플랫폼, 설치 파일과 체크섬은 [Go 공식 다운로드 페이지](https://go.dev/dl/)에서 확인합니다. 이 장의 버전 예시는 **Go 1.27.3**이며, 프로젝트의 최소 버전은 [go.mod](../../go.mod)의 **1.27.0**입니다. 이후 패치를 사용할 때는 파일 이름도 해당 버전으로 맞춥니다.

## 4. Go와 Git 설치하기

### 4.1. 이미 설치되어 있는지 먼저 확인하기

현재 사용할 터미널에서 다음 명령을 실행합니다.

```text
go version
git --version
```

정상 설치되어 있다면 버전이 출력됩니다. 예를 들어 Windows에서는 다음과 비슷한 결과를 볼 수 있습니다.

```text
go version go1.27.3 windows/amd64
```

이는 형식을 보여 주는 예입니다. 운영체제·CPU와 실제 패치 버전에 따라 결과가 달라집니다. 이미 필요한 Go가 있다면 같은 버전을 중복 설치할 필요는 없습니다.

### 4.2. Windows에서 Go 설치

1. [공식 다운로드 페이지](https://go.dev/dl/)에서 자신의 환경에 맞는 Windows 설치 파일을 받습니다.
2. `.msi`를 실행하고 설치 안내에 따릅니다.
3. 설치가 끝나면 PowerShell과 사용 중인 편집기를 닫았다가 다시 엽니다.
4. 새 PowerShell에서 아래 명령을 실행합니다.

```powershell
go version
Get-Command go
```

`Get-Command go`의 `Source` 항목으로 실제 실행 파일 위치를 확인할 수 있습니다. 일반적인 설치에서는 `C:\Program Files\Go\bin\go.exe`와 같은 경로지만, 자신이 선택한 경로가 기준입니다.

설치 프로그램이 환경 변수를 변경해도 이미 열려 있던 터미널에는 이전 값이 남을 수 있습니다. 새 창을 여는 이유가 여기에 있습니다. 절차의 기준은 [Go 공식 설치 안내](https://go.dev/doc/install)입니다.

### 4.3. macOS에서 Go 설치

1. [공식 다운로드 페이지](https://go.dev/dl/)에서 Apple Silicon 또는 Intel 환경에 맞는 `.pkg`를 받습니다.
2. 설치 프로그램을 실행합니다.
3. 터미널과 편집기를 다시 열고 다음을 확인합니다.

```bash
go version
command -v go
```

설치 위치를 확인할 때는 기존에 패키지 관리자로 설치한 Go가 있는지도 살펴봅니다. `command -v go`가 가리키는 경로와 설치한 버전이 예상과 다르다면, 여러 Go가 설치되어 PATH에서 다른 것이 먼저 선택되는 상황일 수 있습니다.

### 4.4. Linux·WSL에서 Go 설치

시스템 전체 설치는 [공식 Linux 설치 절차](https://go.dev/doc/install)를 따릅니다. 기존 Go 설치 디렉터리에 새 배포 파일을 덮어 풀지 않습니다.

다음은 **Linux x86-64에서 관리자 권한 없이 사용자 폴더에 새 버전을 설치하는 예시**입니다. 다른 설치를 삭제하지 않고 별도 경로에 설치하며, 시스템 설치를 이미 사용 중이면 이 절차를 추가로 수행하지 않아도 됩니다. 아래 블록은 Bash에서 순서대로 실행합니다.

#### 1단계: 공식 배포 파일 받기

먼저 `uname -m`이 `x86_64`인지 확인합니다. 다음 명령은 `curl`, `tar`, `sha256sum`이 준비된 환경을 기준으로 합니다. ARM64 환경에서는 아래 파일 이름을 그대로 사용하지 말고 공식 페이지의 해당 파일을 선택합니다.

```bash
mkdir -p "$HOME/Downloads"
curl --fail --location \
  https://go.dev/dl/go1.27.3.linux-amd64.tar.gz \
  --output "$HOME/Downloads/go1.27.3.linux-amd64.tar.gz"
sha256sum "$HOME/Downloads/go1.27.3.linux-amd64.tar.gz"
```

Go 1.27.3의 이 파일에 대해 공식 페이지에서 확인한 SHA256은 다음과 같습니다.

```text
65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40
```

`sha256sum` 출력의 첫 번째 값을 비교합니다. 값이 다르면 설치를 진행하지 말고 파일과 선택한 버전을 다시 확인합니다. 다른 CPU용 파일이나 다른 패치 버전은 체크섬도 다릅니다.

#### 2단계: 새 디렉터리에 풀기

```bash
go_archive="$HOME/Downloads/go1.27.3.linux-amd64.tar.gz"
go_install_dir="$HOME/.local/lib/go1.27.3"

mkdir -p "$HOME/.local/lib"
if [ ! -f "$go_archive" ]; then
  printf '다운로드 파일의 경로를 확인하세요.\n'
elif mkdir "$go_install_dir"; then
  tar -xzf "$go_archive" -C "$go_install_dir" --strip-components=1 &&
    GOTOOLCHAIN=local "$go_install_dir/bin/go" version
else
  printf '설치 경로가 이미 있거나 생성할 수 없습니다. 기존 경로를 확인하세요.\n'
fi
```

`go_archive`와 `go_install_dir`는 이 작업에서 사용할 파일·폴더 경로에 붙인 이름입니다. `mkdir`에 `-p`를 붙이지 않은 설치 디렉터리 생성 단계는 같은 경로가 이미 있으면 실패하므로 기존 설치 위에 자동으로 덮어 풀지 않습니다.

`--strip-components=1`은 압축 파일 맨 위의 `go/` 한 단계를 제거합니다. 따라서 결과는 `$HOME/.local/lib/go1.27.3/bin/go`가 됩니다. `go version go1.27.3 linux/amd64`가 출력되었는지 확인한 뒤 진행합니다. 실패했다면 아래 PATH 설정부터 바꾸지 말고 위 명령의 오류를 먼저 확인합니다.

앞의 `GOTOOLCHAIN=local`은 그 명령 한 번에만 적용하여 방금 설치한 Go 자체의 버전을 확인합니다. 모듈에 따라 다른 도구 체인으로 전환하는 동작은 5절에서 설명합니다.

#### 3단계: 현재 셸의 PATH에 추가하기

```bash
export PATH="$HOME/.local/lib/go1.27.3/bin:$PATH"
command -v go
go version
```

PATH는 셸이 실행 파일을 찾는 디렉터리 목록입니다. 앞에 추가한 Go가 우선 검색됩니다. 이 `export`는 현재 셸과 여기서 시작한 프로그램에 적용됩니다.

새 터미널에서도 유지하려면 자신이 사용하는 셸의 시작 설정 파일에 **위 `export` 한 줄**을 추가합니다. 예를 들어 대화형 비로그인 Bash는 `~/.bashrc`, 대화형 Zsh는 기본적으로 `~/.zshrc`를 사용합니다. 로그인 셸과 원격 접속은 읽는 파일이 다를 수 있으므로, 파일을 덮어쓰지 말고 기존 설정을 확인하여 편집한 뒤 새 터미널에서 다시 검사합니다. 시작 파일의 규칙은 [Bash 문서](https://www.gnu.org/software/bash/manual/html_node/Bash-Startup-Files)와 [Zsh 문서](https://zsh.sourceforge.io/Doc/Release/Files.html)를 참고하세요.

Go 도구는 일반적인 설치에서 실행 파일 위치를 바탕으로 설치 루트를 찾으므로, 이 예시를 위해 `GOROOT` 환경 변수를 별도로 설정하지 않습니다. 과거에 GOROOT를 직접 설정했다면 그 값이 선택에 영향을 줄 수 있으므로 `go env GOROOT`로 실제 설치 루트를 확인합니다. 관련 동작은 [Go의 GOROOT 자동 인식 설명](https://go.dev/doc/go1.10#goroot)을 참고합니다.

### 4.5. Git 설치

[Git 공식 설치 안내](https://git-scm.com/install/)에서 운영체제를 선택합니다. Windows는 Git for Windows 설치 안내를, macOS와 Linux는 해당 환경에서 제공하는 설치 방법을 따릅니다. 이미 `git --version`이 동작한다면 그대로 사용할 수 있습니다.

Git 설치 후 새 터미널에서 확인합니다.

```text
git --version
```

공개 교재를 내려받아 실행하는 데는 GitHub에 업로드할 권한이나 Git 커밋 작성자 설정이 필요하지 않습니다. 코드 변경을 커밋하고 기여하는 절차는 [기여 안내](../../CONTRIBUTING.md)에서 다룹니다.

## 5. 설치 확인과 환경 변수 이해하기

Go가 실행되면 다음 정보를 확인합니다. 이 시점에는 아직 저장소를 내려받지 않았으므로 `GOMOD` 확인은 다음 절에서 진행합니다.

```text
go version
go env GOVERSION
go env GOHOSTOS GOHOSTARCH
go env GOOS GOARCH
```

| 명령·항목 | 읽는 방법 |
| --- | --- |
| `go version` | 현재 명령에 선택된 Go 도구 체인의 버전과 환경 |
| `GOVERSION` | 선택된 도구 체인의 버전 문자열 |
| `GOHOSTOS`, `GOHOSTARCH` | Go 도구 바이너리가 동작하는 운영체제·아키텍처 |
| `GOOS`, `GOARCH` | 빌드할 프로그램의 대상 운영체제·아키텍처 |

보통 처음 설치한 상태에서는 호스트와 빌드 대상이 같습니다. 예를 들어 Linux용 Go를 사용하면서 Windows용 프로그램을 만들도록 설정하면 두 값이 달라질 수 있습니다. 따라서 `GOOS`만 보고 지금 Windows에 접속해 있다고 단정하지 않습니다. 변수의 정의는 [공식 환경 변수 문서](https://pkg.go.dev/cmd/go#hdr-Environment_variables)를 참고합니다.

### PATH·GOROOT·GOPATH·GOMOD의 차이

| 이름 | 뜻 | 처음 학습할 때 할 일 |
| --- | --- | --- |
| `PATH` | 셸이 실행 파일을 찾는 경로 목록 | `go` 명령을 찾을 수 있는지 확인 |
| `GOROOT` | 선택된 Go 도구 체인의 설치 루트 | 보통 자동 인식. 필요할 때 `go env GOROOT`로 확인 |
| `GOPATH` | 설치한 Go 도구·모듈 캐시 등에 사용하는 기본 위치 | 기본값을 사용해도 됨 |
| `GOMOD` | 현재 모듈의 `go.mod` 경로 | 저장소에서 실행한 뒤 올바른 파일을 가리키는지 확인 |

이 프로젝트는 모듈을 사용하므로 저장소를 반드시 `GOPATH/src` 아래에 둘 필요가 없습니다. `Documents`, `src` 등 자신의 작업 폴더에 둘 수 있습니다. 기본 `GOPATH`의 역할은 [Go Wiki: GOPATH](https://go.dev/wiki/GOPATH)에 설명되어 있습니다.

### 실행 위치에 따라 Go 버전이 달라질 수 있는 이유

Go는 `go.mod`의 요구 버전과 `GOTOOLCHAIN` 설정에 따라 사용할 도구 체인을 선택합니다. 따라서 셸에서 처음 찾은 Go 실행 파일과 실제 명령을 처리하는 도구 체인의 버전이 항상 같다고 생각하면 안 됩니다. 필요하면 다음을 확인합니다.

```text
go env GOVERSION GOTOOLCHAIN
```

이 교재에서는 먼저 Go 1.27 계열을 설치하여 예제를 실행합니다. 다른 버전이 선택되거나 다운로드가 시도되면 [Go Toolchains](https://go.dev/doc/toolchain)를 참고해 이유를 확인합니다. `go version`을 “PATH에 있는 원본 실행 파일의 버전을 무조건 출력하는 명령”으로 설명하지 않는 이유입니다.

네트워크가 제한된 환경에서는 요구하는 Go 버전과 필요한 편집기 도구를 미리 준비합니다. 첫 예제 자체는 외부 Go 모듈 없이 실행할 수 있습니다.

## 6. 교재 저장소 내려받기와 폴더 확인

### 6.1. 처음 내려받는 경우

작업할 상위 폴더를 연 뒤 다음을 실행합니다. 저장소를 둘 부모 폴더에서 시작하며, 복제가 끝난 뒤 `cd`로 저장소 안에 들어갑니다.

```text
git clone https://github.com/praster1/GoLang_all_in_one.git
cd GoLang_all_in_one
git status --short --branch
go env GOMOD
```

`git clone`은 저장소의 파일과 Git 이력을 내려받습니다. 다음 `cd`는 현재 작업 디렉터리를 바꿉니다. `go env GOMOD`의 결과가 내려받은 폴더의 `go.mod`를 가리키면 모듈이 인식된 것입니다. [Git clone 문서](https://git-scm.com/docs/git-clone)와 [Go 모듈 문서](https://go.dev/ref/mod)를 참고합니다.

예를 들어 다음처럼 출력될 수 있습니다. 실제 사용자 이름과 상위 폴더는 달라집니다.

```text
Windows 예: C:\work\GoLang_all_in_one\go.mod
Linux 예:   /home/learner/work/GoLang_all_in_one/go.mod
```

### 6.2. 현재 작업 폴더 확인

**저장소 루트**란 이 프로젝트의 최상위 폴더입니다. 이 교재에서는 별도 표시가 없다면 여기에서 명령을 실행합니다.

PowerShell:

```powershell
Get-Location
Get-ChildItem
Get-Content go.mod
```

macOS·Linux·WSL:

```bash
pwd
ls
cat go.mod
```

`go.mod`의 내용은 다음과 같습니다.

```go
module github.com/praster1/GoLang_all_in_one

go 1.27.0
```

첫 줄은 모듈의 식별 경로이고, `go` 줄은 최소 Go 버전입니다. 이 파일은 이미 저장소에 있으므로 학습을 시작하면서 `go mod init`을 다시 실행하지 않습니다.

### 6.3. 저장소 안에서 길 찾기

| 경로 | 들어 있는 것 | 사용 시점 |
| --- | --- | --- |
| `book/` | 교재 본문 | 개념과 설명 읽기 |
| `examples/` | 본문의 실행 예제 | 약속된 동작을 직접 확인 |
| `exercises/` | 연습문제와 힌트 | 자신의 코드로 실습 |
| `solutions/` | 해설과 참고 코드 | 직접 시도한 뒤 비교 |
| `tools/checkdocs/` | 문서 링크 검사 도구 | 저장소 문서를 수정한 뒤 확인 |
| `_practice/` | 자신이 만드는 개인 실습 폴더 | 교재 예제를 복사해 자유롭게 변경 |
| `bin/` | 자신이 빌드하는 실행 파일 | 실행 파일 생성·실행 |

처음 내려받았을 때 `_practice/`와 `bin/`이 없어도 정상입니다. 실습할 때 만듭니다. 두 폴더는 [.gitignore](../../.gitignore)에 따라 Git 추적에서 제외됩니다.

**모듈(module)**은 여러 패키지를 함께 관리하는 단위이고, **패키지(package)**는 관련 Go 코드를 묶는 단위입니다. 지금은 “루트 모듈 하나 안에 예제별 패키지 폴더가 있다”는 정도로 이해하면 됩니다. 폴더마다 `go.mod`를 새로 만들지 않습니다.

`go env GOMOD`는 하위 예제 폴더에서도 같은 루트 모듈을 찾을 수 있습니다. 따라서 이것만으로 현재 위치가 정확히 루트인지 판정하지는 않습니다. 현재 위치는 `Get-Location` 또는 `pwd`로 함께 확인합니다.

### 6.4. 이미 내려받은 경우

같은 저장소 폴더 안에서 다시 `git clone`하지 않습니다. 기존 폴더로 이동해 먼저 변경과 브랜치를 확인합니다.

```text
git status --short
git branch --show-current
```

첫 명령에 변경 내용이 없고 두 번째 결과가 `main`이며 교재를 갱신하려는 경우 다음을 실행합니다.

```text
git pull --ff-only origin main
```

`--ff-only`는 이력이 단순히 앞으로 이동할 수 있을 때 갱신하며, 로컬 커밋과 원격 이력이 갈라졌으면 실패합니다. 변경이 있거나 다른 브랜치에서 작업 중이면 바로 실행하지 말고 작업 내용을 먼저 확인합니다. 실패했다고 파일을 지우거나 강제 초기화할 필요는 없습니다. 자세한 동작은 [Git pull 문서](https://git-scm.com/docs/git-pull)를 참고합니다.

## 7. 편집기에서 프로젝트 열기

편집기는 익숙한 것을 하나 선택하면 됩니다. 이 교재의 실행 명령은 편집기와 무관하게 사용할 수 있습니다. 아래 메뉴 이름은 공식 문서를 기준으로 했으며 제품 버전과 표시 언어에 따라 조금 다를 수 있습니다.

### 7.1. VS Code

1. VS Code에서 **File → Open Folder**로 `go.mod`가 있는 `GoLang_all_in_one` 전체 폴더를 엽니다.
2. 확장 화면에서 공식 **Go** 확장, 식별자 `golang.go`를 설치합니다.
3. `examples/01-first-program/hello/main.go`를 엽니다.
4. 필요한 도구 설치 안내를 확인합니다. Go 언어 서버 `gopls`는 코드 탐색·진단을 제공합니다.
5. **Terminal → New Terminal**을 열고 다음을 실행합니다.

```text
go version
go env GOMOD
```

터미널의 현재 위치와 모듈이 저장소를 가리키는지 확인합니다. 이 저장소에는 Go 파일을 저장할 때 포맷을 정리하는 [.vscode/settings.json](../../.vscode/settings.json)이 포함되어 있습니다.

추가 도구는 명령 팔레트의 **Go: Install/Update Tools**에서 관리할 수 있습니다. `gopls` 같은 편집기 도구는 첫 프로그램에서 사용하는 `fmt` 패키지와 별개입니다. 도구 설치에 네트워크가 필요한 상황과, 프로그램에 외부 모듈을 추가하는 상황을 구분하세요. [VS Code Go 문서](https://code.visualstudio.com/docs/languages/go)와 [확장 명령 목록](https://github.com/golang/vscode-go/blob/master/docs/commands.md)을 참고합니다.

### 7.2. IntelliJ IDEA와 GoLand

**IntelliJ IDEA에서도 Go를 개발할 수 있습니다.** JetBrains의 Go 플러그인이 Go 코드 분석, 모듈 관리, 실행·디버깅 기능을 제공합니다. 공식 문서는 이 구성을 **IntelliJ IDEA Ultimate용 Go 플러그인**으로 안내합니다. 현재 IDE의 플러그인 지원과 활성화된 이용 권한을 확인하세요. IntelliJ IDEA가 하나의 제품으로 통합된 뒤에도 무료 기본 기능과 Ultimate 기능의 구분은 유지됩니다. [Go 플러그인 소개](https://www.jetbrains.com/help/idea/go-plugin.html), [GoLand FAQ](https://www.jetbrains.com/help/go/faq-about-goland.html), [IntelliJ IDEA 통합 제품 안내](https://www.jetbrains.com/help/idea/intellij-idea-single-distribution.html)를 기준으로 확인했습니다.

GoLand는 Go 개발용 별도 IDE입니다. 이미 사용 가능한 IntelliJ IDEA 또는 GoLand가 있다면 해당 도구를 활용하고, VS Code를 선택해도 같은 예제를 학습할 수 있습니다.

#### 저장소와 Go SDK 연결

1. IntelliJ IDEA는 **Settings → Plugins**에서 JetBrains의 **Go** 플러그인을 설치·활성화합니다. GoLand는 Go 기능을 기본으로 제공합니다.
2. **File → Open**으로 `GoLang_all_in_one` 폴더를 엽니다. 새 모듈을 생성하는 화면에서 시작하지 않습니다.
3. Go SDK 설정을 확인합니다. IntelliJ IDEA는 **Settings → Languages & Frameworks → Go → GOROOT**, GoLand는 **Settings → Go → GOROOT**에 있습니다.
4. 이미 설치한 Go를 쓰려면 **Add SDK → Local**에서 설치 최상위 폴더를 선택합니다. 이 폴더 안에 `bin`과 `src`가 있어야 합니다. `go.exe` 파일 자체나 `bin` 폴더만 고르지 않습니다.
5. `go.mod`가 인식되지 않으면 Go Modules 설정에서 모듈 통합 여부와 열어 둔 폴더를 확인합니다.

IDE에서 SDK의 GOROOT를 선택하는 것은 터미널 전체의 `GOROOT` 환경 변수를 임의로 만드는 것과 다른 작업입니다. 메뉴 경로와 SDK 조건은 [IntelliJ GOROOT 설정](https://www.jetbrains.com/help/idea/configuring-goroot-and-gopath.html), [GoLand GOROOT 설정](https://www.jetbrains.com/help/go/configuring-goroot-and-gopath.html), [Go 모듈 통합](https://www.jetbrains.com/help/idea/create-a-project-with-go-modules-integration.html)을 참고하세요.

#### 예제 실행 구성

`examples/01-first-program/hello/main.go`를 열고 `main` 함수 옆 실행 아이콘을 누릅니다. 구성을 직접 확인하려면 **Run → Edit Configurations**에서 다음을 확인합니다.

| 설정 항목 | 값 |
| --- | --- |
| 구성 종류 | Go Build |
| Run kind | Package |
| Package path | `github.com/praster1/GoLang_all_in_one/examples/01-first-program/hello` |
| Working directory | 자신의 `GoLang_all_in_one` 저장소 루트 경로 |
| Run after build | 활성화 |

이 IDE의 `Package path`는 전체 import 경로를 받습니다. 아래 절의 터미널 명령은 루트 기준 상대 경로를 사용하므로 입력 형식이 다릅니다. `Working directory`는 프로그램이 상대 파일 경로를 해석하는 기준입니다. [Go Build 공식 설명](https://www.jetbrains.com/help/go/go-build.html)을 참고합니다.

### 7.3. 편집기와 터미널의 결과 맞추기

일반 터미널과 IDE에 내장된 터미널에서 다음을 각각 실행해 비교합니다. IDE의 실행 버튼이 사용하는 SDK는 별도로 확인합니다.

```text
go version
go env GOROOT
go env GOMOD
```

IntelliJ IDEA·GoLand의 SDK 설정과 터미널의 PATH가 다른 Go를 가리킬 수 있습니다. 내장 터미널의 명령이 성공했다는 사실만으로 Run 구성의 SDK까지 같다고 판단하지 않습니다. VS Code는 **Go: Locate Configured Go Tools**에서 사용하는 도구 위치를 확인할 수 있습니다. 문제를 해결할 때는 설정값을 여러 개 바꾸기보다 먼저 **Go 버전, 설치 경로, 모듈 경로**를 비교하세요.

### 7.4. WSL·Remote SSH에서 편집할 때

VS Code의 원격 연결을 사용한다면 창이 로컬인지 WSL인지 SSH인지 확인합니다. Go 확장과 도구도 실제 개발 환경에서 사용할 수 있어야 합니다. 로컬 폴더를 열어 놓고 서버의 Go 설정을 바꾸어도 로컬 편집기의 문제가 해결되지 않을 수 있습니다.

원격 터미널에서 `pwd`, `command -v go`, `go version`, `go env GOMOD`를 확인한 뒤 이후 실습을 진행합니다. 연결 자체의 설정은 [VS Code Remote SSH](https://code.visualstudio.com/docs/remote/ssh)와 [WSL 개발 안내](https://code.visualstudio.com/docs/remote/wsl)를 참고합니다.

## 8. 첫 프로그램 실행하기

편집기 터미널 또는 일반 터미널에서 **저장소 루트**로 이동합니다. 다음 실제 파일을 열어 봅니다.

[examples/01-first-program/hello/main.go](../../examples/01-first-program/hello/main.go)

```go
package main

import "fmt"

func main() {
	fmt.Println("안녕하세요, Go!")
}
```

지금은 각 키워드를 모두 이해하지 않아도 됩니다. `fmt.Println` 안에 적힌 문장이 터미널에 표시될 것이라고 예상하고 다음을 실행하세요.

```text
go run ./examples/01-first-program/hello
```

예상 출력:

```text
안녕하세요, Go!
```

명령을 나누어 읽으면 다음과 같습니다.

| 부분 | 의미 |
| --- | --- |
| `go` | Go 명령 실행 |
| `run` | 지정한 프로그램을 빌드한 뒤 실행 |
| `./` | 현재 작업 폴더 기준 경로 |
| `examples/01-first-program/hello` | 실행할 예제 패키지가 있는 폴더 |

루트에는 실행할 Go 파일이 없으므로 이 위치에서 `go run .`을 실행하면 원하는 예제가 실행되지 않습니다. 예제의 폴더를 명시합니다. 파일 하나가 아니라 패키지 경로를 사용하면 나중에 같은 폴더에 코드가 여러 파일로 나뉘어도 함께 다룰 수 있습니다. [go run 문서](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)를 참고합니다.

출력이 나오지 않았다면 다음 순서로 확인합니다.

1. `go version`이 동작하는가?
2. 현재 위치가 저장소 루트인가?
3. 지정한 예제 디렉터리와 파일이 있는가?
4. 복사한 명령에 프롬프트나 다른 문자가 섞이지 않았는가?

## 9. 실행과 빌드를 구분하기

`go run`은 코드를 빌드한 뒤 바로 실행합니다. 그 결과를 빠르게 확인하기에 편리합니다. 사용자가 계속 보관해서 실행할 파일을 만들고 싶다면 `go build -o`로 출력 위치를 정합니다.

아래 명령도 **저장소 루트**에서 실행합니다.

### Windows PowerShell

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o bin/hello.exe ./examples/01-first-program/hello
.\bin\hello.exe
```

### macOS·Linux·WSL

```bash
mkdir -p bin
go build -o bin/hello ./examples/01-first-program/hello
./bin/hello
```

`bin` 폴더를 만들고, 예제를 빌드한 뒤, 만들어진 파일을 실행한 것입니다. 두 경우 모두 마지막 실행에서 다음이 출력됩니다.

```text
안녕하세요, Go!
```

빌드에 성공했다고 인사말이 바로 나오는 것은 아닙니다. `go build`와 생성한 파일의 실행은 별도 단계입니다. 코드의 문자열을 수정한 뒤 이전 실행 파일을 그대로 실행하면 이전 내용이 나오므로 다시 빌드해야 합니다.

`go build ./...`는 여러 패키지의 빌드 가능 여부를 확인할 때 사용합니다. 위처럼 보관할 실행 파일 하나를 지정하여 만드는 명령과 목적이 다릅니다. 옵션과 출력 동작은 [go build 문서](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)에 설명되어 있습니다.

## 10. 코드와 문서 검사하기

저장소 루트에서 다음 명령을 하나씩 실행합니다.

```text
go fmt ./...
go test ./...
go vet ./...
go build ./...
go run ./tools/checkdocs
```

| 명령 | 무엇을 하는가 | 결과를 읽는 방법 |
| --- | --- | --- |
| `go fmt ./...` | 대상 Go 파일을 표준 형식으로 수정 | 수정한 파일 경로가 나올 수 있음. 내용 변경을 확인 |
| `go test ./...` | 패키지의 테스트를 실행 | 성공한 패키지에 `ok`. 실패하면 실패 메시지 확인 |
| `go vet ./...` | 의심스러운 코드 패턴을 검사 | 문제가 없으면 보통 별도 출력이 없음 |
| `go build ./...` | 대상 패키지를 빌드할 수 있는지 확인 | 문제가 없으면 보통 별도 출력이 없음 |
| `go run ./tools/checkdocs` | 이 저장소의 로컬 문서 링크 검사 | 검사한 문서·링크 수 또는 오류 위치 표시 |

형식을 정리하는 것은 코드를 실제로 바꾸는 작업입니다. `go fmt`를 실행한 뒤 변경이 있다면 `git diff`로 확인합니다. 테스트에 `(cached)`가 표시되면 Go가 재사용 가능한 이전 성공 결과를 사용한 것입니다. 꼭 새로 실행해야 할 때는 `go test -count=1 ./...`을 쓸 수 있습니다. 관련 동작은 [go fmt](https://pkg.go.dev/cmd/go#hdr-Gofmt__reformat__package_sources), [go test](https://pkg.go.dev/cmd/go#hdr-Test_packages), [go vet](https://pkg.go.dev/cmd/go#hdr-Report_likely_mistakes_in_packages)를 참고하세요.

`./...`는 현재 위치 아래의 패키지를 찾는 패턴입니다. Go는 이름이 `.` 또는 `_`로 시작하는 디렉터리 등을 이 탐색에서 제외합니다. 따라서 개인 실습 폴더는 아래처럼 **직접 경로를 지정해서** 실행·검사합니다. 이 규칙은 [Go의 패키지 목록·패턴 설명](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)에 있습니다.

```text
go fmt ./_practice/00-setup
go run ./_practice/00-setup
```

이 두 명령은 다음 절의 실습 폴더를 만든 뒤 사용합니다. 개인 실습 파일이 저장소 전체 테스트에 포함된다고 생각하지 않도록 주의합니다.

문서 검사 도구는 로컬 파일 링크를 확인하며, 인터넷 주소의 접속 가능 여부나 제목 앵커까지 검사하지는 않습니다. 자세한 범위는 [도구 설명](../../tools/checkdocs/README.md)에 있습니다. 검사 성공은 확인한 조건을 만족했다는 뜻이며, 프로그램과 설명의 모든 오류가 사라졌다는 뜻은 아닙니다.

데이터 레이스 검사는 동시성 학습에서 다룹니다. 현재 장의 필수 완료 조건은 위 기본 검사이며, 지원 환경의 추가 검증 명령은 [기여 안내](../../CONTRIBUTING.md)에서 확인할 수 있습니다.

## 11. 직접 해 보기

[00장 환경 준비 실습](../../exercises/00-setup/README.md)을 순서대로 진행합니다. Go 문법을 새로 외우기보다 환경을 확인하고 도구를 사용하는 실습입니다.

| 실습 | 할 일 | 완료 조건 |
| --- | --- | --- |
| 1. 내 실행 환경 기록 | Go·Git 버전, 작업 폴더, 모듈과 대상 환경 확인 | 각 값이 무엇을 뜻하는지 설명 |
| 2. 개인 프로그램 실행·빌드 | 예제를 `_practice/00-setup/`에 복사해 출력 문구 변경 | `개발 환경 준비 완료!`가 실행과 빌드한 파일 양쪽에서 출력 |
| 3. 잘못된 실행 위치 진단 | 루트에서의 `go run .` 실패와 올바른 경로 비교 | 실패 이유와 수정 명령 설명 |

실습 문서에는 PowerShell과 Bash 명령, 단계별 힌트가 있습니다. 직접 시도한 뒤 [00장 해설](../../solutions/00-setup/README.md)에서 결과를 비교하세요. 기존 교재 예제의 출력은 유지하고 자신의 폴더에서 실험합니다.

## 12. 자주 만나는 문제와 해결 순서

### 12.1. `go` 명령을 찾지 못함

설치 여부와 PATH를 확인합니다.

PowerShell:

```powershell
Get-Command go -All
```

macOS·Linux·WSL:

```bash
command -v go
type -a go
```

설치 직후라면 터미널과 IDE를 다시 엽니다. Linux 사용자별 설치라면 해당 설치의 `bin/go`를 절대 경로로 실행해 봅니다. 절대 경로로는 동작하지만 `go`만 입력하면 실패한다면 PATH 확인부터 시작합니다.

명령은 존재하는데 예상과 전혀 다른 사용법이 출력되는 경우에도 실행 경로를 확인합니다. 이름이 같은 다른 프로그램이 선택되었을 수 있습니다.

### 12.2. Go 버전이 낮거나 갑자기 다운로드를 시도함

```text
go version
go env GOVERSION GOTOOLCHAIN
go env GOMOD
```

프로젝트는 `go.mod`에 최소 버전을 적습니다. 도구 체인 자동 선택 설정에서는 필요한 Go 다운로드가 시도될 수 있습니다. 네트워크가 제한되었다면 사용할 버전을 설치하고 경로를 맞춥니다. 오류를 없애기 위해 교재의 `go.mod` 버전부터 낮추지 않습니다.

### 12.3. `go.mod`를 찾지 못하거나 다른 모듈이 표시됨

현재 폴더를 확인하고 저장소 루트로 이동합니다. 모듈 모드에서 해당 폴더가 모듈에 속하지 않으면 `go env GOMOD`가 Unix에서는 `/dev/null`, Windows에서는 `NUL`을 표시할 수 있습니다. 모듈 모드가 비활성화된 경우에는 빈 값일 수 있습니다. [공식 GOMOD 설명](https://pkg.go.dev/cmd/go#hdr-Environment_variables)을 참고하세요.

다른 프로젝트의 `go.mod`가 나오면 터미널이 그 프로젝트를 보고 있는 것입니다. 해당 위치에서 새 `go.mod`를 만들기보다 먼저 작업 경로를 수정합니다.

### 12.4. `no Go files`가 표시됨

이 저장소의 최상위에는 실행할 `.go` 파일이 없습니다. 루트에서 다음을 실행합니다.

```text
go run ./examples/01-first-program/hello
```

이미 예제 폴더 안에 들어가 있다면 `go run .`도 가능합니다. 같은 명령이라도 **현재 위치가 달라지면 `.`의 대상이 달라진다**는 점이 핵심입니다.

### 12.5. 편집기에서만 빨간 밑줄이 보이거나 실행이 실패함

먼저 일반 터미널에서 예제를 실행합니다. 성공한다면 IDE의 Go SDK, 열린 프로젝트 폴더, 플러그인·확장 상태를 확인합니다. 설치 경로가 같은지도 `go env GOROOT`와 비교합니다. JetBrains IDE는 저장소 루트가 아니라 예제 파일 하나만 열었는지도 확인합니다.

### 12.6. 한글이 깨져 보임

소스 파일을 UTF-8로 저장했는지와 터미널의 문자 표시 환경을 확인합니다. 편집기에서부터 글자가 깨졌다면 파일 인코딩을, 파일은 정상인데 실행 결과만 깨졌다면 터미널 설정을 살펴봅니다. 이미 손상된 텍스트를 저장하지 않도록 원본 예제와 비교합니다.

### 12.7. `exec format error` 또는 실행할 수 없는 파일이라는 오류

현재 실행 환경과 다른 운영체제·CPU용 파일을 받았거나 빌드했는지 확인합니다. Linux에서는 `uname -m`, Go가 동작한다면 `go env GOHOSTOS GOHOSTARCH GOOS GOARCH`를 비교합니다. 실행 권한을 바꾸는 것만으로 운영체제·아키텍처 차이가 해결되지는 않습니다.

### 12.8. 편집한 문장이 실행 파일에 반영되지 않음

편집한 파일을 저장한 뒤 같은 패키지로 다시 빌드합니다. 이전에 만든 `bin/hello`나 `bin/hello.exe`는 소스를 편집하는 것만으로 자동 갱신되지 않습니다. 실행 파일 경로가 다른 복사본을 가리키는지도 확인합니다.

### 질문할 때 남길 기록

오류를 질문할 때는 다음을 함께 적습니다.

- 실행 환경: Windows 직접 실행, WSL, macOS, Linux, Remote SSH 중 무엇인지
- 실행한 명령과 오류 메시지 전체
- 현재 작업 폴더와 `go env GOMOD` 결과
- `go version`, `go env GOROOT` 결과
- 편집기에서만 실패하는지, 터미널에서도 실패하는지

공유할 때 사용자 이름, 실제 서버 주소와 업무 경로는 필요한 부분만 남기고 가릴 수 있습니다. 비밀번호나 토큰은 오류 분석에 필요하지 않습니다.

## 13. 학습을 이어 가는 방법과 완료 기준

앞으로는 다음 순서로 공부합니다.

1. 학습 목표를 읽고 이번 장에서 설명할 수 있어야 할 내용을 확인합니다.
2. 예제를 실행하기 전에 결과를 예상합니다.
3. 실제 결과를 비교하고 입력·출력·코드를 조금 바꿔 봅니다.
4. 막힌 부분은 오류 메시지와 실행 환경을 기록합니다.
5. 연습문제를 시도하고 힌트·해설과 비교합니다.

명령을 여러 번 따라 입력하는 것만으로 이해를 판단하기보다, **무엇을 실행했고 왜 그 결과가 나왔는지** 자신의 말로 설명합니다.

### 완료 체크리스트

- [ ] 코드를 실제로 실행하는 환경이 로컬·WSL·원격 중 어디인지 안다.
- [ ] Go 버전과 실행 파일 경로를 확인했다.
- [ ] Git으로 교재를 내려받고 저장소 루트를 찾았다.
- [ ] `go env GOMOD`가 이 저장소의 `go.mod`를 가리킨다.
- [ ] 선택한 편집기에서 저장소 전체 폴더를 열었다.
- [ ] 첫 예제로 `안녕하세요, Go!`를 출력했다.
- [ ] `go run`과 실행 파일을 만드는 `go build -o`의 차이를 설명할 수 있다.
- [ ] 개인 실습에서 `개발 환경 준비 완료!`를 출력하고 실행 파일도 만들었다.
- [ ] 기본 검사 명령의 결과를 읽었고, 개인 실습 경로는 따로 지정한다는 점을 안다.
- [ ] 오류가 나면 현재 위치·Go 버전·실행 대상부터 확인할 수 있다.

### 자기 점검

1. Windows에 Go를 설치했는데 원격 Linux 터미널에서 `go`를 찾지 못하는 이유는 무엇일까요?
2. `go env GOMOD`가 루트의 파일을 가리키면 현재 폴더도 반드시 저장소 루트일까요?
3. `go run .`과 `go run ./examples/01-first-program/hello`는 무엇이 다를까요?
4. 코드의 문자열을 바꾸었는데 빌드한 실행 파일이 이전 문장을 출력한다면 무엇을 확인할까요?
5. 전체 테스트가 성공했는데 `_practice/`의 오타를 찾지 못할 수 있는 이유는 무엇일까요?

설명은 [00장 해설](../../solutions/00-setup/README.md)에서 확인합니다.

## 정리

이 장에서는 Go와 Git을 준비하고, 저장소와 모듈을 확인하고, 같은 프로그램을 터미널과 편집기에서 실행하는 흐름을 익혔습니다. 실행할 패키지를 지정하고, 필요할 때 파일로 빌드하고, 검사 결과를 읽을 수 있으면 학습을 시작할 준비가 된 것입니다.

다음 [01장](../01-first-program/README.md)에서는 방금 실행한 소스의 `package`, `import`, `func main()`과 `fmt.Println`이 각각 어떤 역할을 하는지 살펴봅니다.

## 참고 자료와 확인 범위

공식 자료 확인일: **2026-11-09**. 설치 파일과 제품 메뉴·이용 조건은 변경될 수 있으므로 해당 문서를 기준으로 확인합니다.

| 자료 | 이 장에서 참고한 내용 |
| --- | --- |
| [Go 다운로드](https://go.dev/dl/) · [설치](https://go.dev/doc/install) · [릴리스 이력](https://go.dev/doc/devel/release) | 배포 버전, 운영체제별 설치와 체크섬 |
| [Go 코드 작성](https://go.dev/doc/code) · [Go 모듈](https://go.dev/ref/mod) | 모듈, 패키지와 `go.mod` |
| [Go 명령](https://pkg.go.dev/cmd/go) · [Go Toolchains](https://go.dev/doc/toolchain) | 실행·빌드·검사, 환경 변수와 도구 체인 선택 |
| [Go Wiki: GOPATH](https://go.dev/wiki/GOPATH) | 현재 모듈 환경에서 GOPATH가 맡는 역할 |
| [Git 설치](https://git-scm.com/install/) · [clone](https://git-scm.com/docs/git-clone) · [pull](https://git-scm.com/docs/git-pull) | 저장소 내려받기와 기존 작업 갱신 |
| [VS Code Go](https://code.visualstudio.com/docs/languages/go) · [공식 Go 확장](https://github.com/golang/vscode-go) | 확장, 언어 도구와 편집기 설정 |
| [Remote SSH](https://code.visualstudio.com/docs/remote/ssh) · [WSL](https://code.visualstudio.com/docs/remote/wsl) | 로컬과 원격 개발 환경의 구분 |
| [IntelliJ Go 플러그인](https://www.jetbrains.com/help/idea/go-plugin.html) · [GoLand FAQ](https://www.jetbrains.com/help/go/faq-about-goland.html) | Go 지원과 제품별 기능 범위 |
| [IntelliJ GOROOT](https://www.jetbrains.com/help/idea/configuring-goroot-and-gopath.html) · [Go Build](https://www.jetbrains.com/help/go/go-build.html) | SDK와 실행 구성 |

명령 검증 범위는 이 장을 반영한 변경 기록에 남깁니다. 설치 프로그램과 IDE의 GUI 절차는 공식 문서로 검토하며, 실제 실행한 플랫폼의 명령 검사와 구분합니다. GitHub Actions는 저장소의 공통 코드·문서 검사를 수행하며 모든 설치 화면이나 편집기 조작을 자동으로 검증하지는 않습니다.

| 이전 | 목차 | 다음 |
| --- | --- | --- |
| 시작 장 | [전체 목차](../README.md) | [01. 첫 Go 프로그램](../01-first-program/README.md) |
