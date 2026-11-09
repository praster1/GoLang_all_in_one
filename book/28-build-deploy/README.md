# 28. 빌드와 배포 자동화

프로그램을 다른 컴퓨터에서 실행하려면 소스가 올바른 것만으로 충분하지 않습니다.
어떤 운영체제와 아키텍처에서 실행할 파일인지,
실행에 필요한 데이터가 어디에 있는지,
검사한 코드와 배포할 파일이 같은 것인지 확인해야 합니다.

이 장에서는 작은 CLI를 기준으로 **검사 → 빌드 → 실행 확인 → 산출물 보관**을 연결합니다.
GitHub Actions로 같은 절차를 반복하고,
컨테이너를 이용해 실행 파일을 전달하는 기본 방법까지 실습합니다.
실제 클라우드나 운영 서버에 서비스를 배포하는 단계는 포함하지 않습니다.
배포 대상별 자격 증명·승인·복구 정책은 그 환경의 별도 결정이 필요합니다.

## 학습 목표

- 빌드 대상 OS·아키텍처와 실행 환경을 구분한다.
- `go build`, `-trimpath`, `-ldflags`, 빌드 제약의 목적을 설명한다.
- 정적 자원을 `go:embed`로 포함하고 실행 시 설정과 구분한다.
- GitHub Actions에서 검사·산출물 생성·실행 확인을 수행한다.
- 여러 단계 Docker 빌드의 책임과 이미지의 제한을 설명한다.

## 사전 지식과 준비

[10장 패키지와 모듈](../10-packages-modules/README.md),
[12장 자동 테스트](../12-testing/README.md),
[16장 명령행 프로그램](../16-cli/README.md),
[25장 서비스 구성과 운영 기초](../25-service-operations/README.md)를 먼저 확인합니다.

- 기준: 2026-11-09에 확인한 Go 1.27.3과 저장소의 Go 1.27 계열.
- 실행 예제: [`examples/28-build-deploy`](../../examples/28-build-deploy/).
- 명령은 저장소 루트에서 실행합니다.
- 기본 Go 빌드와 테스트에는 Docker가 필요하지 않습니다.
- 컨테이너 실습에는 Linux 컨테이너를 실행할 수 있는 Docker Engine 또는 Docker Desktop이 필요합니다.
- 개인 산출물은 `_practice/28-build-deploy/` 또는 무시되는 `dist/`에 둡니다.

도구 옵션은 [go 명령 문서](https://pkg.go.dev/cmd/go)를 기준으로 합니다.
이 장에서 “재현 가능”은 코드·의존성·대상·검사 절차를 다시 실행할 수 있다는 뜻입니다.
모든 환경에서 실행 파일의 바이트가 자동으로 동일하다고 보장하는 표현으로 사용하지 않습니다.

## 1. 빌드와 실행과 배포의 차이

| 단계 | 확인하는 것 | 확인하지 못하는 것 |
| --- | --- | --- |
| 테스트 | 실행한 경로의 동작·오류 계약 | 모든 입력·운영 환경 |
| 빌드 | 선택한 대상에 맞는 프로그램 생성 | 그 대상에서 실제 실행 성공 |
| 스모크 검사 | 대표 기능이 최소 조건에서 실행됨 | 전체 기능·부하·장애 복구 |
| 산출물 보관 | 사용한 파일과 메타데이터 추적 | 파일의 출처 신뢰를 자동 보장 |
| 배포 | 대상 환경에 올리고 전환 | 변경의 안전성을 자동 증명 |

크로스 컴파일(cross compilation)은 현재 컴퓨터와 다른 대상을 위한 실행 파일을 만드는 일입니다.
Linux에서 Windows 실행 파일을 만들 수 있어도 Linux에서 그대로 실행할 수 있다는 뜻은 아닙니다.
해당 대상의 실행 확인은 실제 환경이나 적절한 실행 수단이 필요합니다.

### 1.1 작업 디렉터리와 패키지 경로

`go build ./examples/28-build-deploy/cli`는 패키지를 지정합니다.
`go build ./examples/28-build-deploy/cli/main.go`처럼 파일만 지정하면
다른 소스·빌드 제약·추가 파일을 놓치기 쉽습니다.
이 장에서는 패키지 경로를 사용해 해당 프로그램의 파일을 함께 선택합니다.

## 2. 배포할 작은 CLI

소스: [`cli/main.go`](../../examples/28-build-deploy/cli/main.go)

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func run(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("hello", flag.ContinueOnError)
	fs.SetOutput(errOut)
	name := fs.String("name", "Go", "greeting name")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() != 0 || strings.TrimSpace(*name) == "" {
		fmt.Fprintln(errOut, "name must be non-empty; positional arguments are not accepted")
		return 2
	}
	if _, err := fmt.Fprintf(out, "안녕하세요, %s!\n", *name); err != nil {
		fmt.Fprintln(errOut, "write:", err)
		return 1
	}
	return 0
}

func main() {
	if code := run(os.Args[1:], os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}
```

```sh
go run ./examples/28-build-deploy/cli -name Go
```

```text
안녕하세요, Go!
```

계산과 입출력은 `run`에 두고, 실제 프로세스 종료는 `main`이 담당합니다.
입력 문제는 종료 코드 2, 출력 실패는 1, 성공과 도움말은 0입니다.
`go run` 자체의 종료 코드는 빌드 도구를 거치므로
실행 파일의 종료 동작을 검증할 때는 빌드한 파일을 직접 실행합니다.
[os.Exit](https://pkg.go.dev/os#Exit)과 [go run](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)의 역할을 구분하세요.

### 2.1 테스트에서 프로세스를 끝내지 않기

소스: [`cli/main_test.go`](../../examples/28-build-deploy/cli/main_test.go)

```go
package main

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type brokenWriter struct{}

func (brokenWriter) Write(p []byte) (int, error) { return 0, errors.New("write failed") }

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{nil, 0, "안녕하세요, Go!\n"}, {[]string{"-name", "CI"}, 0, "안녕하세요, CI!\n"},
		{[]string{"-name", " "}, 2, ""}, {[]string{"extra"}, 2, ""}, {[]string{"-bad"}, 2, ""}, {[]string{"-h"}, 0, ""},
	} {
		var out, errOut bytes.Buffer
		if code := run(tc.args, &out, &errOut); code != tc.code || out.String() != tc.want {
			t.Fatalf("args=%v code=%d out=%q", tc.args, code, out.String())
		}
	}
	if run(nil, brokenWriter{}, io.Discard) != 1 {
		t.Fatal("writer failure")
	}
}
```

버퍼를 전달해 표준 출력과 오류를 분리합니다.
실제 `os.Exit`를 테스트 안에서 호출하지 않고 반환한 종료 코드를 확인합니다.
출력 대상에서 오류가 나도 성공으로 처리하지 않는지 검사합니다.

```sh
go test -count=1 ./examples/28-build-deploy/cli
```

이 테스트는 실행 파일 생성 전에 호출할 수 있습니다.
그 뒤 빌드한 파일을 실행하는 스모크 검사가 파일 자체의 동작도 확인합니다.
두 검사의 범위는 서로 다릅니다.

## 3. 현재 환경의 실행 파일 만들기

개인 산출물 디렉터리를 먼저 만듭니다.

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/28-build-deploy | Out-Null
go build -trimpath -o ./_practice/28-build-deploy/hello.exe ./examples/28-build-deploy/cli
./_practice/28-build-deploy/hello.exe -name 개발자
```

Bash(macOS/Linux/WSL):

```sh
mkdir -p _practice/28-build-deploy
go build -trimpath -o ./_practice/28-build-deploy/hello ./examples/28-build-deploy/cli
./_practice/28-build-deploy/hello -name 개발자
```

예상 출력:

```text
안녕하세요, 개발자!
```

이 명령을 다시 실행하면 해당 실습의 이전 실행 파일을 교체합니다.
개인 소스 파일은 수정하지 않습니다.
`-o`는 산출물 경로이고, `-trimpath`는 결과에 기록되는 파일 시스템 경로를 정리합니다.
빌드 옵션의 정확한 범위는 [go build](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)에 있습니다.

### 3.1 종료 코드 확인

PowerShell:

```powershell
./_practice/28-build-deploy/hello.exe -name " "
$LASTEXITCODE
```

Bash:

```sh
./_practice/28-build-deploy/hello -name ' '
printf '%s\n' "$?"
```

오류 메시지는 표준 오류로 출력되고 종료 코드는 2입니다.
Bash에서는 다른 명령을 먼저 실행하면 `$?`가 그 명령의 결과로 바뀌므로 바로 확인합니다.
PowerShell에서는 네이티브 실행 파일의 결과를 `$LASTEXITCODE`로 확인합니다.

### 3.2 소스와 산출물을 구분하기

수정한 소스가 저장되었어도 이미 만든 실행 파일은 자동으로 바뀌지 않습니다.
다시 빌드하고 새 파일을 실행해야 합니다.
다른 폴더의 같은 이름 프로그램을 실행하지 않도록 명시적인 상대 경로를 사용합니다.

## 4. OS와 아키텍처 선택하기

```sh
go env GOOS GOARCH CGO_ENABLED
go tool dist list
```

첫 명령은 현재 설정을, 두 번째는 도구가 지원하는 대상 조합을 보여 줍니다.
지원 목록에 있다는 사실과 프로젝트의 모든 의존성이 그 대상에서 동작한다는 사실은 다릅니다.
실제 프로그램을 빌드하고 필요한 테스트를 수행해야 합니다.

### 4.1 Bash에서 Windows amd64 빌드

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o ./_practice/28-build-deploy/hello-windows-amd64.exe ./examples/28-build-deploy/cli
```

앞의 환경 변수는 이 명령에만 적용됩니다.
운영체제 설정이나 `go env -w`로 사용자의 영구 설정을 바꾸지 않습니다.
만든 파일을 Windows amd64 환경으로 옮긴 뒤 실행 확인을 해야 합니다.

### 4.2 PowerShell에서 Linux amd64 빌드

PowerShell은 Bash의 한 줄 환경 변수 접두어 문법을 사용하지 않습니다.
현재 값을 보관하고 `try/finally`로 복원합니다.

```powershell
$savedGoOS = $env:GOOS
$savedGoArch = $env:GOARCH
$savedCgo = $env:CGO_ENABLED
try {
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    go build -trimpath -o ./_practice/28-build-deploy/hello-linux-amd64 ./examples/28-build-deploy/cli
    if ($LASTEXITCODE -ne 0) { throw "cross build failed" }
} finally {
    $env:GOOS = $savedGoOS
    $env:GOARCH = $savedGoArch
    $env:CGO_ENABLED = $savedCgo
}
```

Windows에서 만든 Linux 파일을 Windows 네이티브 프로그램처럼 실행하지 않습니다.
이 장의 자동화는 Linux와 Windows 러너에서 각각 자신의 실행 파일을 실행합니다.

### 4.3 CGO의 의미

`CGO_ENABLED=0`은 C 연동을 사용하지 않는 빌드를 선택합니다.
C 코드나 특정 시스템 라이브러리가 필요한 패키지는 이 설정으로 빌드되지 않을 수 있습니다.
이 장의 CLI는 표준 라이브러리만 사용하며 해당 설정으로 빌드합니다.
이 설정 하나로 모든 Go 프로그램이 모든 OS의 단일 파일이 된다고 일반화하지 않습니다.
[CGO 문서](https://pkg.go.dev/cmd/cgo)를 참고하세요.

## 5. 빌드 대상에 따라 파일 선택하기

소스: [`platform/main.go`](../../examples/28-build-deploy/platform/main.go)

```go
package main

import (
	"fmt"
	"runtime"
)

func main() { fmt.Printf("target=%s/%s branch=%s\n", runtime.GOOS, runtime.GOARCH, platformName()) }
```

Linux용 파일 [`name_linux.go`](../../examples/28-build-deploy/platform/name_linux.go):

```go
package main

func platformName() string { return "linux" }
```

Windows용 파일 [`name_windows.go`](../../examples/28-build-deploy/platform/name_windows.go):

```go
package main

func platformName() string { return "windows" }
```

그 외 대상용 [`name_other.go`](../../examples/28-build-deploy/platform/name_other.go):

```go
//go:build !linux && !windows

package main

func platformName() string { return "other" }
```

```sh
go run ./examples/28-build-deploy/platform
```

Linux amd64에서는 다음과 같습니다.

```text
target=linux/amd64 branch=linux
```

Windows amd64에서는 다음과 같습니다.

```text
target=windows/amd64 branch=windows
```

아키텍처와 OS가 다르면 해당 부분도 바뀝니다.
일부 OS에는 호환 빌드 태그 규칙이 있으므로 지원 대상을 늘릴 때 공식 규칙을 확인합니다.
[`go:build`와 파일명 제약](https://pkg.go.dev/cmd/go#hdr-Build_constraints)은 빌드 시 파일 선택 규칙입니다.
실행 중 `if runtime.GOOS == ...`로 분기하는 것과 구분하세요.

### 5.1 왜 같은 함수가 충돌하지 않는가

선택된 파일에는 `platformName` 구현이 하나만 남습니다.
조건을 잘못 작성해 둘 이상 선택되면 중복 선언 오류가 납니다.
아무 구현도 선택되지 않으면 선언되지 않은 이름 오류가 납니다.
각 대상의 빌드를 실제로 수행해야 이런 문제를 확인할 수 있습니다.

`go build main.go`처럼 파일 하나만 지정하면
대상별 구현 파일을 함께 선택하지 못해 실패할 수 있습니다.
다음과 같이 패키지 경로를 사용합니다.

```sh
go build ./examples/28-build-deploy/platform
```

파일 이름을 생략한 빌드는 현재 폴더에 기본 실행 파일을 만들 수 있습니다.
실습 결과를 한 곳에 모으려면 앞서 배운 `-o`를 함께 지정하세요.

## 6. 버전 정보를 실행 파일에 넣기

소스: [`version/main.go`](../../examples/28-build-deploy/version/main.go)

```go
package main

import "fmt"

var version = "dev"
var revision = "unknown"

func versionLine(v, r string) string { return fmt.Sprintf("version=%s revision=%s", v, r) }

func main() { fmt.Println(versionLine(version, revision)) }
```

```sh
go run ./examples/28-build-deploy/version
```

```text
version=dev revision=unknown
```

패키지의 문자열 변수 값을 링크할 때 넣을 수 있습니다.
이 예제에서는 릴리스 이름과 코드 식별자를 전달합니다.
실제 비밀키나 개인 접속 정보를 넣는 용도로 사용하지 않습니다.
링커 옵션의 정의는 [cmd/link의 `-X`](https://pkg.go.dev/cmd/link)를 참고합니다.

PowerShell:

```powershell
go build -trimpath -ldflags "-X main.version=1.0.0 -X main.revision=abc123" -o ./_practice/28-build-deploy/version.exe ./examples/28-build-deploy/version
./_practice/28-build-deploy/version.exe
```

Bash:

```sh
go build -trimpath -ldflags '-X main.version=1.0.0 -X main.revision=abc123' -o ./_practice/28-build-deploy/version ./examples/28-build-deploy/version
./_practice/28-build-deploy/version
```

```text
version=1.0.0 revision=abc123
```

여기서 `abc123`은 옵션 사용을 위한 연습용 식별자입니다.
실제 릴리스에서는 빌드한 코드의 커밋 값을 기록해야 합니다.
자동화 파일은 GitHub가 제공하는 해당 실행의 SHA를 사용합니다.

### 6.1 버전 문자열이 신뢰를 자동으로 보장하지는 않는다

실행 파일에 어떤 문자열이 들어 있는지는 빌드한 사람이 정할 수 있습니다.
따라서 버전 표시는 추적을 돕는 정보이며 서명이나 신뢰할 수 있는 빌드 출처의 대체물이 아닙니다.
필요한 검증 수준에 따라 서명·증명·배포 기록을 추가하는 설계가 필요합니다.

### 6.2 도구의 빌드 정보 읽기

```sh
go version -m ./_practice/28-build-deploy/version
```

Windows에서는 경로를 `version.exe`로 바꿉니다.
모듈·빌드 설정·버전 관리 정보가 조건에 따라 표시됩니다.
작업 트리가 수정된 상태인지, 어떤 도구로 만들었는지 함께 살펴보세요.
출력의 모든 항목이 모든 빌드 방식에서 항상 있는 것은 아닙니다.

## 7. 실행 파일에 포함할 자원

소스: [`embed/main.go`](../../examples/28-build-deploy/embed/main.go)

```go
package main

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed message.txt
var message string

func greeting() string { return strings.TrimSpace(message) }

func main() { fmt.Println(greeting()) }
```

함께 포함할 [`message.txt`](../../examples/28-build-deploy/embed/message.txt):

```text
프로그램에 포함된 안내입니다.
```

```sh
go run ./examples/28-build-deploy/embed
```

```text
프로그램에 포함된 안내입니다.
```

`go:embed`는 빌드할 때 패턴에 맞는 자원을 실행 파일에 포함합니다.
실행 시 현재 작업 디렉터리에서 파일을 다시 찾는 방식이 아닙니다.
이 예제의 테스트는 임시 작업 디렉터리로 이동해도 내용이 유지되는지 확인합니다.
[embed 문서](https://pkg.go.dev/embed)에 패턴과 포함 가능한 파일의 제약이 있습니다.

### 7.1 적합한 자원과 부적합한 설정

| 자료 | 포함을 고려할 수 있는가? | 이유 |
| --- | --- | --- |
| 공개 도움말 | 가능 | 코드 버전과 함께 바뀌는 정적 자료 |
| 기본 템플릿 | 가능 | 실행에 필요한 기본 자원을 함께 전달 |
| 실행 환경별 포트 | 보통 실행 설정 | 같은 파일을 다른 환경에서 재사용 |
| 비밀키 | 포함하지 않음 | 실행 파일에서 꺼낼 수 있는 정보가 됨 |
| 사용자 업로드 | 포함 대상 아님 | 빌드 이후 생기는 동적 데이터 |

임베드된 자원을 수정하려면 소스를 바꾸고 다시 빌드해야 합니다.
별도 파일을 읽는 프로그램이라면 실행 디렉터리와 파일 경로 계약을 명확히 해야 합니다.
두 방식을 섞어 현재 실행 파일의 동작을 잘못 이해하지 않습니다.

## 8. 빌드 값과 실행 설정 나누기

소스: [`config/main.go`](../../examples/28-build-deploy/config/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

type config struct{ Port int }

func loadConfig(lookup func(string) (string, bool)) (config, error) {
	value, ok := lookup("BOOK_PORT")
	if !ok {
		return config{Port: 8080}, nil
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return config{}, errors.New("BOOK_PORT must be 1..65535")
	}
	return config{Port: port}, nil
}

func main() {
	values := map[string]string{"BOOK_PORT": "9090"}
	cfg, err := loadConfig(func(key string) (string, bool) { value, ok := values[key]; return value, ok })
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("실행 설정:", cfg.Port)
}
```

```sh
go run ./examples/28-build-deploy/config
```

```text
실행 설정: 9090
```

예제는 설정 조회 함수를 매개변수로 받아 가상의 값으로 동작을 보여 줍니다.
실제 프로세스 환경을 사용하려면 `os.LookupEnv`를 전달할 수 있습니다.
이렇게 의존성을 분리하면 테스트가 사용자 컴퓨터의 환경 변수에 영향을 받지 않습니다.
[LookupEnv](https://pkg.go.dev/os#LookupEnv)는 없는 값과 빈 값의 존재를 구분합니다.

### 8.1 기본값과 잘못된 값

`BOOK_PORT`가 없으면 기본값 8080을 사용합니다.
값이 존재하지만 빈 문자열이거나 범위를 벗어나면 오류입니다.
잘못된 값을 조용히 기본값으로 바꾸면 설정 실수를 늦게 발견할 수 있습니다.
이 예제에서는 시작 단계에서 드러나는 오류를 선택했습니다.

| 입력 상태 | 결과 |
| --- | --- |
| 환경 변수 없음 | 8080 |
| `"1"`, `"65535"` | 해당 포트 |
| `""`, `"0"`, `"65536"`, `"oops"` | 오류 |

이 파서는 `strconv.Atoi`의 정수 표현 규칙을 사용합니다.
앞 장의 ASCII 숫자만 받는 포트 파서와 허용 문법이 다를 수 있습니다.
설정 형식에서 부호나 선행 0을 제한하고 싶다면 별도 정책을 추가합니다.
서로 다른 파서를 같은 계약이라고 가정하지 않습니다.

### 8.2 같은 산출물을 여러 환경에서 사용하기

빌드 때 환경별 비밀과 주소를 넣으면 환경마다 다른 실행 파일이 생깁니다.
대신 코드 버전이 같은 산출물에 실행 시 설정을 전달하면
테스트한 파일과 배포할 파일의 관계를 추적하기 쉬워집니다.
정확한 설정 관리와 인증 경계는 [25장](../25-service-operations/README.md)의 내용을 함께 적용합니다.

## 9. 체크섬으로 파일 내용 확인하기

소스: [`checksum/main.go`](../../examples/28-build-deploy/checksum/main.go)

```go
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

func digest(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func main() {
	sum, err := digest(strings.NewReader("abc"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(sum)
}
```

```sh
go run ./examples/28-build-deploy/checksum
```

```text
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
```

위 값은 줄바꿈 없는 `abc` 세 바이트의 SHA-256입니다.
파일에 마지막 줄바꿈을 넣으면 입력 바이트가 달라져 다른 값이 나옵니다.
[crypto/sha256](https://pkg.go.dev/crypto/sha256)의 해시를 이용해 같은 바이트인지 비교할 수 있습니다.

### 9.1 체크섬이 보장하지 않는 것

공격자가 실행 파일과 체크섬을 함께 바꿀 수 있다면 두 값은 여전히 일치할 수 있습니다.
체크섬만으로 누가 파일을 만들었는지 인증하는 것은 아닙니다.
다운로드한 파일의 내용 비교와 출처를 신뢰하는 문제를 구분합니다.

이 장의 산출물 목록은 파일명·바이트 크기·SHA-256을 기록합니다.
[연습문제의 manifest 구현](../../solutions/28-build-deploy/manifest/main.go)을 자동화에서도 사용합니다.
이미 생성된 `manifest.json` 자체는 목록에서 제외하여 자기 자신을 해시하는 문제를 피합니다.

### 9.2 파일이 바뀌는 중에는 무엇을 해시할까

체크섬을 계산하는 동안 다른 프로세스가 파일을 수정하면 의도한 완성 산출물과 다를 수 있습니다.
빌드가 끝난 디렉터리에 대해 계산하고, 그 뒤 파일을 바꾸지 않는 순서를 사용합니다.
이번 구현은 동시에 변경되는 디렉터리의 일관된 스냅샷까지 제공하는 도구가 아닙니다.

## 10. 최소 실행 확인과 서비스 상태

소스: [`smoke/main.go`](../../examples/28-build-deploy/smoke/main.go)

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "ok")
	})
	return mux
}

func probe(client *http.Client, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4))
	if err != nil {
		return err
	}
	if string(body) != "ok\n" {
		return fmt.Errorf("unexpected health body")
	}
	return nil
}

func main() {
	server := httptest.NewServer(handler())
	defer server.Close()
	client := server.Client()
	client.Timeout = time.Second
	if err := probe(client, server.URL+"/healthz"); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("smoke: ok")
}
```

```sh
go run ./examples/28-build-deploy/smoke
```

```text
smoke: ok
```

이 예제는 자체 로컬 테스트 서버를 열고 한 번 요청한 뒤 닫습니다.
외부 서비스에 연결하거나 계속 실행되는 서버를 남기지 않습니다.
상태 코드만 보지 않고 응답 본문도 약속한 값인지 확인합니다.
요청에 제한 시간을 두고 응답 본문을 닫습니다.
[httptest](https://pkg.go.dev/net/http/httptest)와 [Client.Timeout](https://pkg.go.dev/net/http#Client)을 사용합니다.

### 10.1 읽기 한도를 하나 더 두는 이유

정상 본문은 `ok\n`의 3바이트입니다.
4바이트까지 읽으면 정상 접두어 뒤에 추가 데이터가 붙은 경우를 구분할 수 있습니다.
앞의 3바이트만 읽고 같다고 판단하면 `ok\nextra`도 통과시킬 수 있습니다.
약속한 본문 전체가 맞는지 확인하는 경계입니다.

### 10.2 살아 있음과 준비됨

`/healthz` 하나가 모든 운영 상태를 대표하지는 않습니다.
프로세스가 응답 가능한지와 요청을 실제로 처리할 준비가 되었는지는 다를 수 있습니다.
DB 연결이나 시작 작업이 필요한 서비스에서는 확인할 의존성과 실패 정책을 정해야 합니다.

이 예제의 상태 확인은 자기 프로세스의 간단한 응답 계약입니다.
실제 운영 준비 검사에 추가할 항목은 서비스 요구에 따라 선택합니다.
개인 정보나 상세 내부 오류를 상태 응답에 그대로 노출하지 않습니다.

## 11. 기존 검사 과정을 CI로 반복하기

이 저장소의 공통 검사 파일은 [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml)입니다.

```yaml
name: Go and documentation

on:
  push:
    branches: [main]
  pull_request:
  workflow_dispatch:

permissions:
  contents: read

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  verify:
    name: Verify (${{ matrix.os }})
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    timeout-minutes: 20
    env:
      GOTOOLCHAIN: local
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          persist-credentials: false
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          # go.mod sets the language minimum; CI follows security patches in that minor.
          # Keep this selector in sync when changing the go directive in go.mod.
          go-version: '1.27.x'
          check-latest: true
          cache: true
      - name: Show Go version
        run: go version
      - name: Check formatting
        shell: bash
        run: |
          unformatted=$(gofmt -l .)
          if [ -n "$unformatted" ]; then
            printf '%s\n' "$unformatted"
            printf 'Run go fmt ./... and commit the formatted files.\n'
            exit 1
          fi
      - name: Build all packages
        run: go build ./...
      - name: Vet all packages
        run: go vet ./...
      - name: Test all packages
        run: go test -count=1 ./...
      - name: Check documentation links
        run: go run ./tools/checkdocs
      - name: Run race detector
        if: runner.os == 'Linux'
        run: go test -race -count=1 ./...
```

### 11.1 각 검사의 질문

| 검사 | 확인하는 질문 |
| --- | --- |
| 포맷 | 모든 Go 파일이 gofmt 형식인가? |
| 빌드 | 루트 모듈의 패키지를 컴파일할 수 있는가? |
| vet | 대표적인 잘못된 코드 패턴이 있는가? |
| 테스트 | 정상·경계·오류·예제 출력의 계약을 지키는가? |
| 문서 링크 | 연결한 로컬 파일이 실제로 존재하는가? |
| race | 실행한 경로에서 데이터 경쟁이 검출되는가? |

`go test -count=1`은 테스트 결과 캐시를 재사용하지 않고 실행하도록 합니다.
의존성·빌드 캐시와 테스트 결과 캐시는 서로 다른 용도입니다.
캐시가 없어도 필요한 의존성을 받아 다시 검사할 수 있어야 합니다.
[GitHub의 Go CI 안내](https://docs.github.com/en/actions/tutorials/build-and-test-code/go)를 참고하세요.

### 11.2 Linux와 Windows를 함께 검사하는 이유

경로 구분자, 파일 잠금, 실행 파일 이름, 셸 문법에서 차이가 드러날 수 있습니다.
두 러너에서 실제로 테스트를 실행하면 단순히 다른 OS 대상으로 컴파일하는 것보다 넓게 확인합니다.
다만 이것이 macOS나 모든 아키텍처의 실행 검증까지 포함하는 것은 아닙니다.

### 11.3 권한과 실행 범위

검사 워크플로는 저장소 내용을 읽는 권한으로 시작합니다.
체크아웃 단계에서 자격 증명을 계속 남기지 않도록 설정했습니다.
외부 Action은 실제 확인한 커밋 SHA를 지정해 버전 이동을 명시적으로 관리합니다.
향후 업데이트 때 해당 릴리스와 변경 내용을 다시 확인해야 합니다.
[Actions 보안 사용 안내](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions)를 참고합니다.

## 12. 빌드 산출물을 보관하는 워크플로

별도 파일 [build-artifacts.yml](../../.github/workflows/build-artifacts.yml)은
학습용 실행 파일을 만들고 Linux·Windows에서 실행한 뒤 보관합니다.
같은 파일에서 컨테이너 빌드와 실행도 검사합니다.

```yaml
name: Build learning artifacts

on:
  pull_request:
    paths:
      - 'examples/28-build-deploy/**'
      - 'solutions/28-build-deploy/**'
      - 'go.mod'
      - 'go.sum'
      - '.github/workflows/build-artifacts.yml'
  workflow_dispatch:

permissions:
  contents: read

concurrency:
  group: learning-artifacts-${{ github.ref }}
  cancel-in-progress: true

jobs:
  binary:
    name: Binary (${{ matrix.goos }})
    strategy:
      fail-fast: false
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            suffix: ''
          - os: windows-latest
            goos: windows
            suffix: '.exe'
    runs-on: ${{ matrix.os }}
    timeout-minutes: 15
    env:
      GOTOOLCHAIN: local
      CGO_ENABLED: '0'
      GOARCH: amd64
      GOOS: ${{ matrix.goos }}
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          persist-credentials: false
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version: '1.27.x'
          check-latest: true
          cache: true
      - name: Verify artifact programs
        run: go test -count=1 ./examples/28-build-deploy/... ./solutions/28-build-deploy/...
      - name: Build and smoke test native binaries
        shell: bash
        run: |
          mkdir -p dist
          go build -trimpath -o "dist/hello-${{ matrix.goos }}-amd64${{ matrix.suffix }}" ./examples/28-build-deploy/cli
          go build -trimpath -ldflags "-X main.version=chapter28 -X main.revision=${{ github.sha }}" -o "dist/version-${{ matrix.goos }}-amd64${{ matrix.suffix }}" ./examples/28-build-deploy/version
          actual=$("./dist/hello-${{ matrix.goos }}-amd64${{ matrix.suffix }}" -name CI)
          test "$actual" = '안녕하세요, CI!'
          "./dist/version-${{ matrix.goos }}-amd64${{ matrix.suffix }}"
          go run ./solutions/28-build-deploy/manifest -dir dist > manifest.json
          mv manifest.json dist/manifest.json
      - uses: actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2
        with:
          name: learning-${{ matrix.goos }}-amd64
          path: dist/
          if-no-files-found: error
          retention-days: 7

  container:
    name: Container smoke test
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          persist-credentials: false
      - name: Build container
        run: docker build --file examples/28-build-deploy/cli/Dockerfile --tag golang-book-hello:ci .
      - name: Run container
        shell: bash
        run: |
          actual=$(docker run --rm golang-book-hello:ci -name CI)
          test "$actual" = '안녕하세요, CI!'
```

### 12.1 테스트 뒤에 빌드와 실행

`Verify artifact programs`에서 예제·참고 구현의 계약을 검사합니다.
그 뒤 CLI와 버전 프로그램을 빌드하고 실제 파일을 실행합니다.
CLI의 출력이 `안녕하세요, CI!`와 같은지 확인하므로
컴파일만 되고 실행하지 못하는 산출물을 지나치지 않습니다.

버전 프로그램에는 해당 CI 실행의 SHA가 들어갑니다.
PR 이벤트에서는 GitHub가 검사하는 병합 후보의 SHA일 수 있으므로,
사용자가 로컬에서 보고 있는 기능 브랜치 SHA와 같은 것이라고 추정하지 않습니다.
어떤 코드가 빌드되었는지 실행 메타데이터를 기준으로 추적합니다.

### 12.2 목록 파일의 생성 순서

먼저 `dist/`의 바이너리들을 해시하여 목록을 저장합니다.
목록 파일은 디렉터리 밖에 만들고 완성된 뒤 옮깁니다.
이 순서는 읽는 중인 목록 파일을 동시에 쓰는 혼동을 피합니다.
구현 자체도 기존 `manifest.json`을 목록에서 제외합니다.

```sh
go run ./solutions/28-build-deploy/manifest -dir dist
```

이 명령은 기존 `dist/`의 일반 파일들을 이름순으로 읽습니다.
하위 디렉터리는 재귀 탐색하지 않고, 심볼릭 링크 같은 비일반 파일은 거부합니다.
큰 파일을 메모리에 전부 읽지 않고 스트림으로 해시합니다.

### 12.3 Artifact와 캐시

Artifact는 이번 실행에서 만든 파일을 받거나 검토할 때 사용합니다.
캐시는 다음 실행의 의존성·빌드 준비 비용을 줄이는 데 사용합니다.
두 목적을 섞어 캐시를 릴리스 저장소처럼 취급하지 않습니다.
[워크플로 산출물](https://docs.github.com/en/actions/concepts/workflows-and-actions/workflow-artifacts)의 구분을 참고하세요.

이 워크플로의 보관 기간은 7일입니다.
영구 릴리스 배포나 외부 레지스트리 업로드를 수행하지 않습니다.
장기 배포가 필요하면 버전·서명·보관 정책을 정한 별도 릴리스 절차로 확장합니다.

내려받은 파일은 OS·아키텍처가 맞는지 확인합니다.
압축 전달 과정에서 실행 권한이 유지되지 않았다면 Linux/macOS에서
신뢰할 수 있는 해당 파일에 실행 권한을 설정해야 할 수 있습니다.
파일 내용의 체크섬과 실행 권한 메타데이터는 다른 항목입니다.

## 13. 여러 단계 컨테이너 빌드

파일: [`cli/Dockerfile`](../../examples/28-build-deploy/cli/Dockerfile)

```dockerfile
FROM golang:1.27.3-alpine3.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY examples/28-build-deploy/cli ./examples/28-build-deploy/cli
RUN CGO_ENABLED=0 go test ./examples/28-build-deploy/cli
RUN CGO_ENABLED=0 go build -trimpath -o /out/hello ./examples/28-build-deploy/cli

FROM scratch
COPY --from=build /out/hello /hello
USER 65532:65532
ENTRYPOINT ["/hello"]
CMD ["-name", "Go"]
```

첫 단계는 Go 도구를 사용해 테스트하고 실행 파일을 만듭니다.
두 번째 단계는 결과 파일만 가져와 실행합니다.
컴파일러와 소스 전체를 최종 실행 환경에 함께 넣지 않습니다.
[Docker의 여러 단계 빌드](https://docs.docker.com/build/building/multi-stage/)가 이 구성을 설명합니다.

### 13.1 빌드 컨텍스트

다음 명령의 마지막 `.`은 저장소 루트를 빌드 컨텍스트로 지정합니다.
Dockerfile이 있는 폴더가 아니라 루트를 기준으로 `COPY` 경로를 해석합니다.

```sh
docker build --file examples/28-build-deploy/cli/Dockerfile --tag golang-book-hello:local .
docker run --rm golang-book-hello:local -name 개발자
```

예상 출력:

```text
안녕하세요, 개발자!
```

컨테이너는 인사 한 번을 출력하고 종료됩니다.
`--rm`은 종료한 컨테이너를 지우지만 빌드한 이미지를 지우는 옵션은 아닙니다.
사용한 이미지는 필요에 따라 Docker의 이미지 관리 명령으로 관리합니다.

함께 제공하는 [Dockerfile.dockerignore](../../examples/28-build-deploy/cli/Dockerfile.dockerignore)는
개인 실습·빌드 결과·로컬 설정 등이 컨텍스트에 들어가는 것을 줄입니다.
[컨텍스트와 무시 파일](https://docs.docker.com/build/concepts/context/)의 적용 위치를 확인하세요.

### 13.2 `scratch` 이미지의 제한

최종 단계의 `scratch`에는 셸과 일반 명령 도구가 없습니다.
그래서 `ENTRYPOINT`는 실행 파일을 직접 지정한 JSON 배열 형태를 사용합니다.
이 CLI에는 추가 파일이나 네트워크 인증서가 필요하지 않습니다.

HTTPS 요청, 시간대 데이터, CGO 라이브러리, 런타임 설정 파일이 필요한 프로그램이라면
그 자원을 별도로 포함하거나 적절한 실행 이미지를 선택해야 합니다.
이 Dockerfile을 모든 Go 서비스의 완성된 배포 파일로 복사하지 않습니다.
`USER 65532:65532`도 실행 파일에 필요한 쓰기 권한을 자동으로 만들어 주지는 않습니다.

### 13.3 버전 태그와 다이제스트

빌드 단계는 확인한 공식 Go 이미지 `1.27.3-alpine3.24`를 사용합니다.
[공식 이미지의 태그](https://hub.docker.com/_/golang)를 2026-11-09에 확인했습니다.
태그는 사람이 읽기 쉬운 이름이며 영구 불변의 내용 식별자가 아닐 수 있습니다.
더 엄격한 공급망·재현성 요구가 있다면 검토한 이미지 다이제스트와 업데이트 정책을 함께 관리합니다.

## 14. 직접 실습하기: 테스트한 파일을 전달하기

1. 기본 CLI 테스트를 실행합니다.
2. 자신의 OS용 `hello`를 개인 산출물 폴더에 빌드합니다.
3. 정상 인사와 빈 이름의 종료 코드를 확인합니다.
4. 버전 프로그램에 연습 버전·식별자를 넣어 빌드합니다.
5. 두 파일의 목록·크기·SHA-256을 생성합니다.
6. 임베드 예제를 다른 작업 디렉터리에서 실행해 봅니다.
7. Docker가 있다면 컨테이너 인사까지 실행합니다.

### 14.1 실습 기록 표

| 항목 | 기록할 내용 |
| --- | --- |
| 소스 | 사용한 커밋 또는 개인 변경 사항 |
| 도구 | Go 버전, Docker 사용 시 그 버전 |
| 대상 | GOOS/GOARCH/CGO_ENABLED |
| 검사 | 실행한 테스트와 성공·실패 |
| 산출물 | 파일명·크기·SHA-256 |
| 실행 확인 | 실제 실행한 OS와 정상·오류 결과 |
| 미실행 | 크로스 빌드만 한 대상, 사용할 수 없었던 도구 |

빌드만 한 Windows 파일을 Linux에서 실행 검증했다고 기록하지 않습니다.
이 작업 환경에서 실행하지 못한 도구가 있다면 해당 절차는 구문·공식 문서 확인과 실제 실행을 구분합니다.
저장소의 CI 실행 결과는 해당 PR과 Actions 기록에서 확인할 수 있습니다.

## 15. 자주 하는 실수

| 증상 | 원인 후보 | 확인·수정 |
| --- | --- | --- |
| 실행 형식 오류 | OS·아키텍처가 다름 | GOOS/GOARCH와 대상 환경 확인 |
| 명령을 찾지 못함 | 현재 폴더의 파일을 PATH로 찾음 | `./경로`로 명시 |
| 파일이 바뀌지 않은 듯함 | 이전 실행 파일 실행 | 다시 빌드하고 경로·버전 확인 |
| embed 파일 누락 오류 | 패턴·경로·빌드 컨텍스트 문제 | 소스 기준 상대 경로와 COPY 확인 |
| 대상별 함수 중복 | 빌드 조건이 겹침 | 파일명과 go:build 조건 점검 |
| 대상별 함수 없음 | 파일만 빌드하거나 조건 누락 | 패키지 경로로 빌드, 대상별 검사 |
| Windows 빌드 이후 테스트 실행 실패 | GOOS 설정을 복원하지 않음 | 환경 변수 범위를 제한하고 복원 |
| Docker COPY 실패 | 컨텍스트가 다름 | 루트에서 마지막 `.` 확인 |
| 컨테이너에서 셸이 없음 | scratch의 특성 | 실행 파일 직접 지정, 필요한 이미지 선택 |
| 체크섬 불일치 | 내용·줄바꿈·빌드 조건 차이 | 실제 바이트와 같은 산출물 여부 확인 |
| CI만 느리거나 시간 초과 | 초기 의존성 컴파일·검사 증가 | 실패 단계와 시간 기록, 합리적인 예산 설정 |
| 산출물을 찾을 수 없음 | 경로·이름·보관 기간 문제 | 실행의 artifact 목록과 설정 확인 |

## 16. 연습문제와 해설

[연습문제](../../exercises/28-build-deploy/README.md)에서는
OS별 산출물 이름, 체크섬 목록, 임베드 설정, 상태 확인 계약과 자동화 개선을 다룹니다.
[해설](../../solutions/28-build-deploy/README.md)은
실행 가능한 프로그램과 테스트, 각 선택의 적용 범위를 제공합니다.

```sh
go test -count=1 ./examples/28-build-deploy/... ./solutions/28-build-deploy/...
```

## 17. 요약과 자기 점검

검사할 코드와 전달할 파일의 관계를 추적합니다.
대상 환경에 맞게 빌드하고, 실제 파일을 실행하고, 필요한 메타데이터를 남깁니다.
CI와 컨테이너는 이 절차를 반복하는 도구이며 잘못된 계약을 자동으로 고쳐 주지는 않습니다.

### 확인 질문

1. 크로스 컴파일 성공과 해당 OS 실행 성공은 무엇이 다른가요?
2. `GOOS`를 전역에 남겨 두면 이후 `go test`에 어떤 영향을 줄 수 있나요?
3. `-trimpath`가 모든 빌드 결과의 동일성을 보장하나요?
4. `go:embed`와 실행 시 파일 읽기는 언제 내용을 결정하나요?
5. 버전 문자열과 체크섬만으로 출처를 인증할 수 있나요?
6. Artifact와 캐시는 각각 무엇을 위해 사용하나요?
7. 테스트한 후 산출물을 다시 바꾸면 어떤 연결이 끊어지나요?
8. Dockerfile 경로와 빌드 컨텍스트 경로는 같은가요?
9. scratch 이미지에 없는 자원을 세 가지 말할 수 있나요?
10. 상태 코드 200만 확인하는 스모크 검사의 한계는 무엇인가요?

<details>
<summary>답 확인</summary>

1. 빌드는 대상 형식의 파일 생성이며 실행 검증은 해당 환경에서 실제 동작을 확인하는 일입니다.
2. 다른 OS용 테스트 파일을 만들어 현재 컴퓨터에서 실행하지 못할 수 있습니다.
3. 경로 기록을 정리하는 옵션이며 버전·의존성·플래그 등 다른 입력까지 같게 만들지는 않습니다.
4. embed는 빌드 때, 파일 읽기는 실행 때의 내용을 사용합니다.
5. 둘 다 변경할 수 있으므로 신뢰할 수 있는 출처·서명 등과 구분해야 합니다.
6. Artifact는 결과 전달·보관, 캐시는 반복 실행의 준비 비용 감소가 목적입니다.
7. 검증한 파일과 배포한 파일이 같다는 근거가 사라집니다.
8. -f는 Dockerfile을, 마지막 인자는 COPY가 읽을 컨텍스트를 정합니다.
9. 셸, CA 인증서 묶음, 시간대 데이터나 C 공유 라이브러리 등이 기본으로 없습니다.
10. 약속한 본문이나 실제 의존성 준비 상태가 잘못되어도 통과할 수 있습니다.

</details>

### 완료 체크리스트

- [ ] 테스트와 빌드를 구분하여 실행했습니다.
- [ ] 현재 OS용 실행 파일의 성공·오류 종료를 확인했습니다.
- [ ] 다른 대상의 빌드 결과를 구별되는 이름으로 만들었습니다.
- [ ] 버전 정보와 모듈 빌드 정보를 확인했습니다.
- [ ] 임베드 자원과 실행 설정의 역할을 구분했습니다.
- [ ] 산출물 목록과 체크섬을 생성했습니다.
- [ ] CI의 실패 단계와 산출물 위치를 찾을 수 있습니다.
- [ ] 컨테이너 실행 여부와 미검증 환경을 정확하게 기록했습니다.

## 참고 자료

- [go 명령](https://pkg.go.dev/cmd/go)
- [링커 옵션](https://pkg.go.dev/cmd/link)
- [embed](https://pkg.go.dev/embed)
- [CGO](https://pkg.go.dev/cmd/cgo)
- [GitHub Actions의 Go 검사](https://docs.github.com/en/actions/tutorials/build-and-test-code/go)
- [워크플로 산출물](https://docs.github.com/en/actions/concepts/workflows-and-actions/workflow-artifacts)
- [upload-artifact v7.0.2](https://github.com/actions/upload-artifact/tree/v7.0.2)
- [Docker 여러 단계 빌드](https://docs.docker.com/build/building/multi-stage/)
- [Docker 빌드 컨텍스트](https://docs.docker.com/build/concepts/context/)
- [공식 Go 컨테이너 이미지](https://hub.docker.com/_/golang)

확인 기준: 2026-11-09. Action·이미지 버전과 실행 결과는 변경 시 다시 확인합니다.

## 다음 학습

[29장 CSV 분석 CLI](../29-csv-cli/README.md)부터는 지금까지 배운 기능을
작동하는 프로그램 하나로 묶고 입력·오류·테스트·실행 방법을 완성합니다.

[이전: 27. 퍼징과 견고한 테스트](../27-fuzzing/README.md) · [전체 목차](../README.md) · [다음: 29. 종합 실습 CSV 분석 CLI](../29-csv-cli/README.md)
