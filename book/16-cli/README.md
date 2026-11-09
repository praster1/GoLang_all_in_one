# 16. 명령줄 프로그램 만들기

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 응용 | 함수·오류·테스트, Reader와 Writer, JSON | Go, 터미널, 이 저장소 |

이제 코드를 수정하지 않고 프로그램의 행동을 바꿔 보겠습니다.
명령줄에서 이름을 고르고, 파일이나 다른 명령의 결과를 입력으로 받고,
성공 여부를 다음 프로그램이 판단할 수 있게 돌려줍니다.

**명령줄 인터페이스(command-line interface, CLI)**는
프로그램과 사용자가 문자열 인자·입출력 스트림·종료 코드로 약속하는 방식입니다.
화면에 보기 좋은 문장을 출력하는 일과 자동화에 사용할 결과를 내보내는 일을 구분해야 합니다.

이번 장은 작은 CLI를 끝까지 구성합니다.
외부 CLI 프레임워크 없이 표준 라이브러리 `flag`, `os`, `io`를 사용합니다.
사용법 오류, 입력 오류, 출력 오류를 테스트할 수 있는 구조로 만듭니다.

## 학습 목표

- 프로그램 이름과 사용자가 넘긴 인자를 구분한다.
- 로컬 `FlagSet`으로 옵션·기본값·도움말·남은 인자를 처리한다.
- 환경 변수의 미설정과 빈 값을 구분하고 설정 우선순위를 정한다.
- 표준 입력·표준 출력·표준 오류의 역할을 분리한다.
- 처리 함수는 값을 반환하고 `main`이 종료 코드를 전달하게 만든다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
Bash는 macOS/Linux/WSL을, PowerShell은 Windows를 기준으로 표기합니다.
Go 실행 명령이 같아도 환경 변수 설정·파이프·종료 코드 확인 문법은 다릅니다.

입력을 읽는 예제는 본문처럼 파이프로 데이터를 전달합니다.
입력 없이 실행하면 표준 입력이 끝나기를 기다릴 수 있습니다.
이 기다림을 무한 계산이나 컴파일 실패로 해석하지 않습니다.

기본 파이프 예제의 입력은 `Go`, `Test` 같은 ASCII 문자열입니다.
Windows PowerShell과 최신 PowerShell의 문자 인코딩 기본값이 다르므로
한글 외부 입력은 사용하는 셸의 UTF-8 설정도 확인해야 합니다.
[PowerShell 인코딩 문서](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_character_encoding)를 참고합니다.

| 예제 | 중심 내용 | 기본 실행 |
| --- | --- | --- |
| [arguments](../../examples/16-cli/arguments/main.go) | 인자 배열 | `go run ./examples/16-cli/arguments Go "file input"` |
| [flags](../../examples/16-cli/flags/main.go) | 이름·불리언 플래그 | `go run ./examples/16-cli/flags -name learner -upper` |
| [environment](../../examples/16-cli/environment/main.go) | LookupEnv와 기본값 | `go run ./examples/16-cli/environment` |
| [stdin-lines](../../examples/16-cli/stdin-lines/main.go) | 표준 입력의 끝까지 읽기 | 5절의 파이프 명령 |
| [streams](../../examples/16-cli/streams/main.go) | 결과와 진단 분리 | `go run ./examples/16-cli/streams` |
| [exit-code](../../examples/16-cli/exit-code/main.go) | 0·1·2 종료 코드 | `go run ./examples/16-cli/exit-code 5` |
| [app](../../examples/16-cli/app/main.go) | 인자·입력·출력 통합 | 8절의 파이프 명령 |

## 1. 프로그램과 셸의 경계

### 1.1 입력을 받는 세 통로

인자, 환경 변수, 표준 입력은 모두 문자열과 관련 있지만 역할이 다릅니다.

| 통로 | 주로 담는 값 | Go의 진입점 |
| --- | --- | --- |
| 인자 | 이번 실행의 옵션과 대상 | `os.Args` |
| 환경 변수 | 실행 환경에 설정된 기본값 | `os.LookupEnv` |
| 표준 입력 | 처리할 연속 데이터 | `os.Stdin` |

이 구분은 절대적인 규칙이 아닙니다.
설정 파일이나 환경 변수 이름도 인자로 받을 수 있습니다.
핵심은 어떤 값이 어디에서 왔고 무엇이 우선하는지 문서화하는 것입니다.

예를 들어 `-name learner`는 이번 실행에서 선택한 이름입니다.
`STUDY_COURSE=Go`는 프로그램이 참조하도록 정한 환경 설정입니다.
파이프로 들어오는 여러 학습 항목은 처리할 데이터입니다.

### 1.2 셸이 먼저 해석한다

터미널에 적은 줄 전체가 하나의 문자열로 Go에 전달되는 것은 아닙니다.
셸이 인용과 변수 확장 등을 처리한 뒤 프로그램 인자를 구성합니다.
리디렉션 기호도 보통 셸이 해석하여 파일과 스트림을 연결합니다.

Bash에서 `"file input"`은 공백을 포함한 인자 하나입니다.
따옴표를 제거하고 `file input`이라고 쓰면 두 인자가 됩니다.
[GNU Bash 명령 확장](https://www.gnu.org/s/bash/manual/html_node/Simple-Command-Expansion.html)을 참고합니다.

입력 경로에 공백이 있을 때도 같은 이유로 인용이 필요합니다.
프로그램 안에서 모든 인자를 다시 공백으로 이어 붙였다가 나누면
원래 인자 경계를 잃어버릴 수 있습니다.

## 2. `os.Args`와 위치 인자

`os.Args[0]`은 프로그램 이름에 해당하는 문자열입니다.
사용자가 넘긴 인자는 보통 `os.Args[1:]`에 있습니다.
[os.Args](https://pkg.go.dev/os#pkg-variables)의 계약을 확인합니다.

실행 파일 이름이나 임시 빌드 경로는 환경에 따라 달라질 수 있습니다.
이번 출력은 사용자의 인자만 보여 주어 결과를 안정적으로 비교합니다.

소스: [arguments/main.go](../../examples/16-cli/arguments/main.go)

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func showArgs(args []string, output io.Writer) error {
	if _, err := fmt.Fprintf(output, "인자 수: %d\n", len(args)); err != nil {
		return err
	}
	for index, argument := range args {
		if _, err := fmt.Fprintf(output, "%d: %q\n", index, argument); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := showArgs(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

Bash와 PowerShell 공통:

```sh
go run ./examples/16-cli/arguments Go "file input"
```

예상 결과:

```text
인자 수: 2
0: "Go"
1: "file input"
```

### 2.1 출력에서 무엇을 확인하는가?

| 값 | 의미 |
| --- | --- |
| `len(args)` | 사용자가 전달한 인자 개수 |
| `args[0]` | 첫 사용자 인자 |
| `%q` | 따옴표와 이스케이프가 보이는 문자열 출력 |
| `os.Args[1:]` | 프로그램 이름을 제외한 슬라이스 |

여기서 출력 인덱스 0은 `os.Args[0]`을 뜻하지 않습니다.
`showArgs`에 전달한 사용자 인자 슬라이스의 첫 요소입니다.
슬라이스를 잘라 넘긴 뒤에도 원래 인덱스가 유지된다고 생각하지 않습니다.

인자를 주지 않으면 정상적으로 `인자 수: 0`만 출력합니다.
인자를 반드시 하나 받는 프로그램은 별도로 길이를 검사해야 합니다.
길이 확인 없이 `args[0]`을 읽으면 빈 입력에서 패닉이 날 수 있습니다.

### 2.2 `go run`의 옵션과 프로그램 옵션

다음 명령에서 패키지 경로 뒤의 옵션은 실행할 프로그램에 전달됩니다.

```sh
go run ./examples/16-cli/flags -name Go
```

`go run -name Go ...`처럼 쓰면 Go 도구의 옵션으로 해석될 수 있습니다.
프로그램의 옵션은 프로그램 패키지 뒤에 둡니다.
[go run 문서](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)를 참고합니다.

빌드한 실행 파일은 Go 도구를 거치지 않고 인자를 직접 받습니다.
나중에 종료 코드를 확인할 때는 이 차이가 중요해집니다.

## 3. `flag.FlagSet`으로 옵션 정의하기

### 3.1 전역 옵션 대신 로컬 집합

표준 `flag` 패키지는 플래그를 이름과 타입으로 등록합니다.
`FlagSet`은 독립적인 플래그 집합입니다.
이번 장은 처리 함수가 호출될 때 새 FlagSet을 만듭니다.

이렇게 하면 테스트마다 다른 인자를 넣어도 이전 실행의 플래그 값이 남지 않습니다.
전역 `flag.CommandLine`이나 실제 `os.Args`를 테스트에서 바꿀 필요도 없습니다.
[FlagSet](https://pkg.go.dev/flag#FlagSet)을 참고합니다.

`ContinueOnError`는 파싱 실패를 오류 값으로 돌려줍니다.
프로그램 전체를 즉시 끝내는 정책을 처리 함수 안에서 사용하지 않습니다.
호출자는 오류와 상황에 맞는 종료 코드를 선택합니다.

소스: [flags/main.go](../../examples/16-cli/flags/main.go)

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

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("greet", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "Go", "인사할 이름")
	upper := flags.Bool("upper", false, "이름의 영문자를 대문자로")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || strings.TrimSpace(*name) == "" {
		fmt.Fprintln(stderr, "오류: 빈 이름과 위치 인자는 허용하지 않습니다")
		return 2
	}
	value := strings.TrimSpace(*name)
	if *upper {
		value = strings.ToUpper(value)
	}
	if _, err := fmt.Fprintf(stdout, "안녕하세요, %s!\n", value); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```

실행:

```sh
go run ./examples/16-cli/flags -name learner -upper
```

예상 결과:

```text
안녕하세요, LEARNER!
```

### 3.2 등록 → 파싱 → 검증 → 출력

| 단계 | 수행 내용 |
| --- | --- |
| 등록 | `name` 문자열, `upper` bool, 기본값·설명 |
| 파싱 | 문자열 인자를 각 타입의 값으로 변환 |
| 검증 | 빈 이름과 남은 위치 인자 거부 |
| 처리 | 공백 정리, 필요하면 대문자 변환 |
| 출력 | 결과 한 줄, 쓰기 오류 확인 |

`flags.String`과 `flags.Bool`은 값을 가리키는 포인터를 반환합니다.
파싱 후 `*name`, `*upper`로 실제 값을 읽습니다.
파싱 전에 읽으면 기본값만 확인하게 됩니다.

이름을 빈 문자열로 만드는 것도 문자열 플래그의 문법에는 맞습니다.
그래서 `TrimSpace` 후 비었는지는 프로그램이 검사합니다.
타입 파싱과 의미 검증을 구분하는 원칙은 15장과 같습니다.

### 3.3 불리언 플래그의 형태

다음 형태는 서로 다릅니다.

| 인자 | 의미 |
| --- | --- |
| `-upper` | upper를 true로 |
| `-upper=true` | upper를 true로 |
| `-upper=false` | upper를 false로 |
| `-upper false` | upper는 true, `false`는 남은 인자 |

불리언 false를 지정할 때는 `=` 형식을 사용합니다.
[flag 문법](https://pkg.go.dev/flag#hdr-Command_line_flag_syntax)에 이 구분이 설명되어 있습니다.

이 예제는 남은 위치 인자를 받지 않으므로 `-upper false`를 사용법 오류로 거부합니다.
파싱만 성공했다고 해서 사용자가 의도한 옵션 조합이라고 판단하지 않습니다.

일반 문자열 플래그는 `-name Go`와 `-name=Go` 모두 가능합니다.
`--name=Go`처럼 두 하이픈도 허용됩니다.
표준 flag가 모든 다른 CLI 도구의 단축 옵션 규칙을 지원한다고 가정하지 않습니다.

### 3.4 첫 위치 인자와 `--`

표준 flag 파싱은 첫 번째 플래그가 아닌 인자에서 멈춥니다.
`--`는 플래그 파싱 종료를 명시하는 구분자입니다.
뒤쪽 값은 플래그처럼 생겼어도 남은 인자가 됩니다.

| 인자 배열 | 플래그 파싱 뒤 남는 인자 |
| --- | --- |
| `-name Go` | 없음 |
| `file -upper` | `file`, `-upper` |
| `-- -upper` | `-upper` |
| `-upper=false` | 없음 |

현재 예제는 남은 인자를 모두 거부합니다.
파일 경로를 받는 CLI라면 `flags.Args()`를 경로 목록으로 사용할 수 있습니다.
이때 옵션을 파일 뒤에도 자유롭게 둘 수 있다고 설명하면 안 됩니다.

### 3.5 도움말도 계약이다

`-h` 또는 `-help`를 별도로 등록하지 않았을 때
FlagSet은 도움말을 출력하고 `flag.ErrHelp`를 반환합니다.
이번 프로그램은 이를 사용자의 정상 요청으로 보고 0을 반환합니다.
[ErrHelp](https://pkg.go.dev/flag#ErrHelp)를 참고합니다.

도움말과 파싱 진단은 `SetOutput(stderr)`로 표준 오류에 보냅니다.
도움말을 표준 출력에 보내는 CLI도 있을 수 있습니다.
여기서는 출력 위치를 테스트에서 확인할 수 있도록 명시적으로 정했습니다.

```sh
go run ./examples/16-cli/flags -h
```

이 명령은 인사 결과를 표준 출력에 쓰지 않습니다.
도움말에는 프로그램 이름과 등록한 옵션·기본값 설명이 포함됩니다.
표준 라이브러리가 만드는 공백 배치 전체를 교재의 별도 프로토콜로 고정하지 않습니다.

## 4. 환경 변수: 미설정과 빈 문자열

### 4.1 `LookupEnv`의 두 반환값

`os.Getenv`는 값을 문자열로 돌려줍니다.
`os.LookupEnv`는 값과 존재 여부를 함께 돌려주므로
설정되지 않음과 빈 문자열로 설정됨을 구분할 수 있습니다.
[LookupEnv](https://pkg.go.dev/os#LookupEnv)를 참고합니다.

이번 정책은 다음과 같습니다.

| 상태 | 처리 |
| --- | --- |
| `STUDY_COURSE` 없음 | 기본값 Go |
| 값이 `" Rust "` | 공백을 다듬어 Rust |
| 값이 빈 문자열 또는 공백만 | 오류 |

미설정일 때만 기본값을 사용합니다.
잘못 지정된 값을 조용히 기본값으로 바꾸지 않는 선택입니다.

소스: [environment/main.go](../../examples/16-cli/environment/main.go)

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func courseName(lookup func(string) (string, bool)) (string, error) {
	value, exists := lookup("STUDY_COURSE")
	if !exists {
		return "Go", nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("STUDY_COURSE는 비어 있을 수 없습니다")
	}
	return value, nil
}

func run() error {
	course, err := courseName(os.LookupEnv)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, "과정:", course)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

환경 변수를 지정하지 않았다면 예상 결과는 다음과 같습니다.

```text
과정: Go
```

기존 셸 환경에 `STUDY_COURSE`가 있다면 그 값이 적용됩니다.
정해진 출력을 확인하려면 값을 명시적으로 전달합니다.

### 4.2 Bash에서 이번 명령에만 지정

```sh
STUDY_COURSE=Rust go run ./examples/16-cli/environment
```

예상 결과:

```text
과정: Rust
```

이 형태는 실행할 명령의 환경에 값을 추가합니다.
현재 셸의 같은 변수 값을 영구적으로 바꾸기 위한 명령이 아닙니다.
[GNU Bash 명령 확장](https://www.gnu.org/s/bash/manual/html_node/Simple-Command-Expansion.html)을 참고합니다.

### 4.3 PowerShell 7.5 이상에서 설정 후 복원

PowerShell에서는 `$env:`를 사용합니다.
이 예제는 원래 환경 변수 값을 저장했다가 `finally`에서 복원합니다.
빈 문자열을 환경 값으로 보존하는 동작은 PowerShell 7.5 이상 기준입니다.
이전 버전은 빈 값과 제거의 동작이 다르므로 미설정·빈 값 검증은 Go 단위 테스트로 확인합니다.
[PowerShell 환경 변수 문서](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_environment_variables)를 참고합니다.

```powershell
$studyCourseExisted = Test-Path Env:STUDY_COURSE
$studyCoursePrevious = $env:STUDY_COURSE
try {
    $env:STUDY_COURSE = "Rust"
    go run ./examples/16-cli/environment
}
finally {
    if ($studyCourseExisted) {
        $env:STUDY_COURSE = $studyCoursePrevious
    }
    else {
        Remove-Item Env:STUDY_COURSE -ErrorAction SilentlyContinue
    }
}
```

다른 프로그램의 환경 설정을 지우는 `Clearenv` 같은 동작은 필요하지 않습니다.
전체 환경을 출력해 확인하는 대신 필요한 이름 하나만 읽습니다.
예제에는 비밀 값이나 토큰을 사용하지 않습니다.

### 4.4 환경 의존성을 함수로 전달하기

`courseName`은 `os.LookupEnv`를 직접 호출하지 않습니다.
같은 시그니처의 함수를 매개변수로 받습니다.
04장에서 배운 함수 값이 외부 환경을 대체하는 자리에 쓰인 것입니다.

실행 시에는 실제 `os.LookupEnv`를 전달합니다.
테스트에서는 원하는 `(value, exists)`를 돌려주는 함수를 전달합니다.
이렇게 하면 테스트끼리 환경 변수를 변경하며 간섭할 필요가 없습니다.

플래그와 환경 변수를 함께 쓰는 경우에는 우선순위를 정해야 합니다.
연습문제 4는 **명시적 플래그 > 존재하는 환경 변수 > 기본값**으로 정합니다.
빈 플래그를 지정한 경우도 “미지정”과 구분합니다.

## 5. 표준 입력을 끝까지 읽기

### 5.1 `os.Stdin`은 Reader

14장에서 문자열 Reader를 넘겼던 자리에 `os.Stdin`을 전달할 수 있습니다.
처리 함수는 입력이 터미널인지 파일인지 파이프인지 알 필요가 없습니다.
읽는 규칙만 같으면 같은 코드를 사용합니다.

소스: [stdin-lines/main.go](../../examples/16-cli/stdin-lines/main.go)

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func summarize(source io.Reader, destination io.Writer) error {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("입력 읽기: %w", err)
	}
	_, err := fmt.Fprintf(destination, "학습 항목: %d\n", count)
	return err
}

func main() {
	if err := summarize(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

Bash:

```sh
printf 'Go\n\nTest\n' | go run ./examples/16-cli/stdin-lines
```

PowerShell:

```powershell
"Go", "", "Test" | go run ./examples/16-cli/stdin-lines
```

예상 결과:

```text
학습 항목: 2
```

### 5.2 입력이 끝난다는 뜻

파이프의 앞 프로그램이 데이터를 쓰고 종료하면 뒤쪽은 입력 끝을 만납니다.
Scanner는 정상 EOF에서 반복을 마칩니다.
정상 EOF는 Scanner.Err에서 오류로 보고하지 않습니다.

터미널에서 직접 실행하면 사용자가 입력을 끝내야 합니다.
Unix 계열 터미널은 보통 빈 입력 위치에서 Ctrl+D로 EOF를 전달하고,
Windows 콘솔에서는 Ctrl+Z 다음 Enter를 사용하는 경우가 있습니다.
터미널 종류가 다를 수 있으므로 기본 실습은 파이프 명령을 사용합니다.

출력을 확인하겠다고 무작정 기다릴 필요가 없습니다.
먼저 실행한 명령에 입력 공급과 종료가 포함되어 있는지 확인합니다.
중단이 필요하면 터미널의 Ctrl+C로 프로그램을 종료할 수 있습니다.

### 5.3 프롬프트를 출력하지 않은 이유

자동화용 필터가 표준 출력에 `입력하세요:`를 쓰면
다음 프로그램이 그 문장도 데이터로 받게 됩니다.
이 예제는 대화형 질문 없이 표준 입력을 처리합니다.

사용자에게 진행 상황을 알려야 한다면 표준 오류를 별도로 사용할 수 있습니다.
표준 오류는 실패 문장만 쓰는 유일한 통로라는 뜻이 아닙니다.
결과 데이터와 구분할 진단·진행 메시지의 통로로 설계할 수 있습니다.

### 5.4 줄 크기와 전체 입력

Scanner의 버퍼 상한은 한 토큰을 읽을 때 필요한 저장 공간과 관련됩니다.
128KiB 설정을 전체 입력이 128KiB 이하라는 뜻으로 설명하지 않습니다.
짧은 줄이 아주 많으면 전체 입력은 훨씬 클 수 있습니다.

이 예제는 기본 스트림 처리 구조를 보여 줍니다.
고정된 테스트 입력과 일반적인 크기의 줄을 전제로 합니다.
개인 응용에서는 항목 수·전체 바이트·시간 제한을 요구사항에 맞춰 추가합니다.

## 6. 결과와 진단을 분리하기

### 6.1 stdout과 stderr

| 스트림 | 이 장에서의 역할 | 연결 예 |
| --- | --- | --- |
| 표준 입력 stdin | 처리할 데이터 | 파이프 또는 입력 파일 |
| 표준 출력 stdout | 다음 프로그램이 읽을 결과 | 결과 파일 또는 다음 파이프 |
| 표준 오류 stderr | 진단과 진행 메시지 | 터미널 또는 진단 파일 |

두 출력이 같은 터미널에 보이면 분리가 느껴지지 않을 수 있습니다.
하지만 서로 다른 파일에 보내 보면 계약을 확인할 수 있습니다.

소스: [streams/main.go](../../examples/16-cli/streams/main.go)

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func writeResult(stdout, stderr io.Writer) error {
	if _, err := fmt.Fprintln(stderr, "처리: 2개"); err != nil {
		return err
	}
	_, err := fmt.Fprint(stdout, "Go\nTest\n")
	return err
}

func main() {
	if err := writeResult(os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "출력 오류:", err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./examples/16-cli/streams
```

표준 출력:

```text
Go
Test
```

표준 오류:

```text
처리: 2개
```

터미널에서 두 스트림이 합쳐져 보이는 순서를 프로토콜로 의존하지 않습니다.
테스트에서는 각각 다른 bytes.Buffer를 전달해 내용과 위치를 따로 확인합니다.

### 6.2 Bash 리디렉션으로 확인

다음 명령은 개인 실습의 결과 파일을 새 내용으로 씁니다.
같은 이름으로 다시 실행하면 이전 실습 결과를 덮어씁니다.
보존할 결과가 있으면 다른 이름을 선택합니다.

```sh
mkdir -p _practice/16-cli/output
go run ./examples/16-cli/streams > _practice/16-cli/output/result.txt 2> _practice/16-cli/output/diagnostic.txt
cat _practice/16-cli/output/result.txt
cat _practice/16-cli/output/diagnostic.txt
```

`>`는 표준 출력을, `2>`는 표준 오류를 파일에 연결합니다.
이 기호는 Go 소스의 연산자가 아니라 셸 문법입니다.
[GNU Bash 리디렉션](https://www.gnu.org/s/bash/manual/html_node/Redirections.html)을 참고합니다.

PowerShell에도 리디렉션이 있지만 버전에 따른 파일 인코딩 차이가 있습니다.
바이트를 그대로 비교하는 검증과 화면에 내용을 확인하는 검증을 구분합니다.
이 장의 Go 단위 테스트는 셸 인코딩 변환을 거치지 않습니다.

### 6.3 출력 실패도 실패다

프로그램이 계산을 끝냈어도 결과를 쓸 수 없으면 사용자는 결과를 받지 못합니다.
따라서 `Fprintf`, `Encode`, Writer의 오류를 확인합니다.
표준 출력이 항상 성공한다고 가정하지 않습니다.

오류 진단을 표준 오류에 쓰는 마지막 동작은 최선의 시도입니다.
그 진단 출력까지 실패하면 다시 같은 곳에 무한히 진단을 쓰지 않습니다.
프로그램은 이미 정한 실패 종료 코드로 실패를 전달합니다.

## 7. 종료 코드를 처리 함수에서 반환하기

### 7.1 이 장의 코드 정책

종료 코드는 프로그램을 호출한 셸이나 다른 프로그램이 읽는 정수입니다.
일반적으로 0은 성공이고 0이 아닌 값은 실패입니다.
1과 2의 세부 분류는 이 장의 프로그램에서 정한 정책입니다.

| 코드 | 이 장의 의미 |
| --- | --- |
| 0 | 정상 처리 또는 도움말 |
| 1 | 읽기·쓰기 등 실행 중 오류 |
| 2 | 사용법·옵션·입력 값 검증 오류 |

모든 Go 프로그램에서 2가 같은 뜻이라는 언어 규칙은 아닙니다.
호출자와 테스트가 사용할 수 있도록 프로그램 문서에서 약속합니다.

소스: [exit-code/main.go](../../examples/16-cli/exit-code/main.go)

```go
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "사용법: days 0~7")
		return 2
	}
	days, err := strconv.Atoi(args[0])
	if err != nil || days < 0 || days > 7 {
		fmt.Fprintln(stderr, "오류: 일수는 0~7 정수여야 합니다")
		return 2
	}
	if _, err := fmt.Fprintf(stdout, "학습일: %d\n", days); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```

실행:

```sh
go run ./examples/16-cli/exit-code 5
```

예상 표준 출력:

```text
학습일: 5
```

`0`, `7`은 정상 입력입니다.
인자 없음, `-1`, `8`, `x`, 두 개의 인자는 코드 2입니다.
오류일 때 정상 결과를 표준 출력에 쓰지 않습니다.

### 7.2 `os.Exit`와 defer

`os.Exit`는 프로세스를 즉시 종료하고 defer 함수를 실행하지 않습니다.
[os.Exit](https://pkg.go.dev/os#Exit)의 명시적인 계약입니다.
그래서 처리 함수 안에서 파일을 열고 곧바로 Exit하는 형태를 피합니다.

`os.Exit(run(...))`에서는 먼저 `run`이 호출되어 반환해야 합니다.
그 과정에서 run과 그 하위 함수의 defer가 실행됩니다.
반환값이 계산된 뒤 main이 Exit를 호출합니다.

반대로 **main 자체에 등록된 defer는 이 Exit로 실행되지 않습니다.**
정리가 필요한 자원은 run 또는 그 아래의 함수가 소유하도록 합니다.
11장의 자원 수명과 이 호출 순서를 연결해 이해합니다.

### 7.3 종료 코드는 실행 파일로 검사하기

`go run`은 Go 도구가 빌드한 프로그램을 실행하는 명령입니다.
Go 도구의 종료 상태가 내부 실행 파일의 종료 코드와 같다는 보장은 없습니다.
[공식 go run 문서](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)도 이 차이를 명시합니다.

정확히 2가 나오는지 확인하려면 먼저 빌드합니다.

Bash:

```sh
mkdir -p bin
go build -o bin/days ./examples/16-cli/exit-code
./bin/days 8
printf '종료 코드: %s\n' "$?"
```

오류 메시지는 표준 오류에, 다음 결과는 표준 출력에 보입니다.

```text
종료 코드: 2
```

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o bin/days.exe ./examples/16-cli/exit-code
.\bin\days.exe 8
"종료 코드: $LASTEXITCODE"
```

Bash의 `$?`는 직전에 실행한 명령의 상태입니다.
중간에 다른 명령을 실행하면 확인하려던 값이 바뀔 수 있습니다.
[GNU Bash 특수 매개변수](https://www.gnu.org/s/bash/manual/html_node/Special-Parameters.html)를 참고합니다.

PowerShell의 `$LASTEXITCODE`는 네이티브 프로그램의 종료 코드를 담습니다.
[자동 변수 문서](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_automatic_variables#lastexitcode)를 참고합니다.
셸 스크립트의 오류 중단 설정에 따라 비정상 명령 뒤 실행이 멈출 수 있으므로
기본 확인은 대화형 터미널에서 명령을 순서대로 실행합니다.

## 8. 인자·입력·출력을 한 처리 함수에 모으기

### 8.1 경계를 주입하는 구조

이제 최대 항목 수와 접두사를 플래그로 받고,
표준 입력의 비어 있지 않은 줄을 번호와 함께 출력합니다.

처리 함수는 실제 os 전역 값 대신 필요한 경계를 매개변수로 받습니다.
`main`은 그 자리에 실제 인자와 표준 스트림을 연결하는 얇은 어댑터입니다.

소스: [app/main.go](../../examples/16-cli/app/main.go)

```go
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func run(args []string, source io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("study-lines", flag.ContinueOnError)
	flags.SetOutput(stderr)
	prefix := flags.String("prefix", "학습", "각 항목 앞에 붙일 이름")
	limit := flags.Int("limit", 3, "허용할 최대 항목 수: 1~100")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *limit < 1 || *limit > 100 || strings.TrimSpace(*prefix) == "" {
		fmt.Fprintln(stderr, "오류: prefix와 limit(1~100)을 확인하세요")
		return 2
	}
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	lines := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if len(lines) == *limit {
			fmt.Fprintln(stderr, "오류: 항목 수 제한을 넘었습니다")
			return 2
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	for index, line := range lines {
		if _, err := fmt.Fprintf(stdout, "%s %d: %s\n", strings.TrimSpace(*prefix), index+1, line); err != nil {
			fmt.Fprintln(stderr, "출력 오류:", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

Bash:

```sh
printf 'Go\n\nTest\n' | go run ./examples/16-cli/app
```

PowerShell:

```powershell
"Go", "", "Test" | go run ./examples/16-cli/app
```

예상 결과:

```text
학습 1: Go
학습 2: Test
```

### 8.2 처리 순서와 출력 시점

| 단계 | 실패하면 | stdout |
| --- | --- | --- |
| 옵션 파싱·검증 | 코드 2 | 비어 있음 |
| 비어 있지 않은 줄 수 제한 | 코드 2 | 비어 있음 |
| Scanner 읽기 | 코드 1 | 비어 있음 |
| 모든 항목 출력 | 쓰기 실패면 코드 1 | 일부가 이미 쓰였을 수 있음 |
| 완료 | 코드 0 | 번호가 붙은 항목 |

이 구현은 최대 100개 항목을 모은 뒤 출력합니다.
입력 제한 위반이 뒤에서 발견돼도 앞 항목을 결과로 공개하지 않습니다.
출력 Writer 자체의 부분 쓰기를 되돌리는 보장은 없습니다.

빈 입력은 항목 0개로 정상 종료하며 출력도 없습니다.
“출력이 없음”만으로 실패 여부를 판단할 수 없다는 예입니다.
종료 코드와 표준 오류를 함께 확인합니다.

### 8.3 옵션을 바꾸기

Bash:

```sh
printf ' Go \nTest\n' | go run ./examples/16-cli/app -prefix 항목 -limit 2
```

PowerShell:

```powershell
" Go ", "Test" | go run ./examples/16-cli/app -prefix 항목 -limit 2
```

예상 결과:

```text
항목 1: Go
항목 2: Test
```

같은 입력에서 `-limit 1`이면 항목 수 오류입니다.
첫 줄 하나만 출력하고 조용히 나머지를 버리는 옵션이 아닙니다.
이름이 limit이라고 해서 동작을 추측하지 않고 설명과 테스트로 확인합니다.

### 8.4 단위 테스트가 main을 호출하지 않는 이유

`main`은 실제 프로세스 인자를 읽고 `os.Exit`를 호출합니다.
테스트 프로세스 안에서 그대로 호출하면 테스트 전체가 끝날 수 있습니다.
따라서 이 장은 `run`을 직접 호출하는 표 기반 테스트를 사용합니다.

`bytes.Buffer`로 stdout과 stderr를 각각 받고
반환 코드·정확한 결과 문자열·진단 존재 여부를 함께 확인합니다.
외부 프로세스 검증은 실제 빌드한 CLI의 연결이 맞는지 확인하는 데 사용합니다.

테스트가 통과하도록 전역 `os.Args`를 매번 덮어쓰지 않습니다.
입출력 함수의 경계를 설계하면 운영체제 상태를 바꾸지 않고도 실패를 재현할 수 있습니다.

## 9. 개인 실습: 주간 학습 계획 CLI

### 9.1 만들 기능

`-days`, `-minutes`, `-json`을 받아 주간 학습 계획을 계산합니다.
입력 값이 잘못되면 결과를 stdout에 쓰지 않고 코드 2를 반환합니다.
출력 실패는 코드 1, 도움말은 코드 0입니다.

| 옵션 | 기본값 | 허용 범위 |
| --- | --- | --- |
| `-days` | 5 | 0~7 정수 |
| `-minutes` | 25 | 1~120 정수 |
| `-json` | false | 불리언 플래그 문법 |

0일은 이번 주 학습이 없는 정상 계획입니다.
하루 분량은 1 이상이어야 합니다.
합계의 최대는 7×120=840분입니다.

### 9.2 개인 파일 준비

Bash:

```sh
mkdir -p _practice/16-cli/plan
if [ ! -e _practice/16-cli/plan/main.go ]; then
  cp examples/16-cli/flags/main.go _practice/16-cli/plan/main.go
fi
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/16-cli/plan | Out-Null
if (-not (Test-Path _practice/16-cli/plan/main.go)) {
    Copy-Item examples/16-cli/flags/main.go _practice/16-cli/plan/main.go
}
```

기존 개인 파일은 보존됩니다.
이제 개인 `main.go` 전체를 다음 코드로 바꿉니다.

```go
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Plan struct {
	Days         int `json:"days"`
	DailyMinutes int `json:"daily_minutes"`
	TotalMinutes int `json:"total_minutes"`
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	days := flags.Int("days", 5, "학습일: 0~7")
	minutes := flags.Int("minutes", 25, "하루 분량: 1~120")
	jsonOutput := flags.Bool("json", false, "JSON 출력")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *days < 0 || *days > 7 || *minutes < 1 || *minutes > 120 {
		fmt.Fprintln(stderr, "오류: days(0~7)와 minutes(1~120)를 확인하세요")
		return 2
	}
	plan := Plan{
		Days:         *days,
		DailyMinutes: *minutes,
		TotalMinutes: *days * *minutes,
	}
	var err error
	if *jsonOutput {
		err = json.NewEncoder(stdout).Encode(plan)
	} else {
		_, err = fmt.Fprintf(stdout, "학습일: %d\n하루: %d분\n총: %d분\n",
			plan.Days, plan.DailyMinutes, plan.TotalMinutes)
	}
	if err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```

실행:

```sh
gofmt -w _practice/16-cli/plan/main.go
go run ./_practice/16-cli/plan
```

예상 결과:

```text
학습일: 5
하루: 25분
총: 125분
```

JSON 출력:

```sh
go run ./_practice/16-cli/plan -days 7 -minutes 30 -json
```

예상 결과:

```json
{"days":7,"daily_minutes":30,"total_minutes":210}
```

### 9.3 변경할 값과 관찰할 결과

| 인자 | 정상 출력 또는 결과 |
| --- | --- |
| 없음 | 총 125분, 코드 0 |
| `-days 0` | 총 0분, 코드 0 |
| `-days 7 -minutes 120` | 총 840분, 코드 0 |
| `-minutes 1` | 총 5분, 코드 0 |
| `-days -1` | stdout 없음, 코드 2 |
| `-days 8` | stdout 없음, 코드 2 |
| `-minutes 0` | stdout 없음, 코드 2 |
| `-minutes 121` | stdout 없음, 코드 2 |
| `-days x` | 파싱 오류, 코드 2 |
| `-json false` | 남은 인자 오류, 코드 2 |
| `-json=false` | 텍스트 출력, 코드 0 |
| `-h` | 도움말, 코드 0 |
| `extra -days 1` | 남은 인자 오류, 코드 2 |

검사하려는 코드가 정확히 2인지 보려면 실행 파일을 빌드합니다.
함수를 직접 호출하는 테스트에서도 반환값 2를 확인할 수 있습니다.
단순히 오류 메시지에 `2`가 포함되는지를 검사하지 않습니다.

### 9.4 테스트를 설계하는 질문

- 기본값만 사용하는 실행이 정상인가?
- 양 끝 값인 0일·7일·1분·120분이 계약에 맞는가?
- 잘못된 옵션이면 입력 처리나 결과 출력 전에 종료하는가?
- 도움말이 데이터 결과와 섞이지 않는가?
- JSON의 숫자가 문자열이 아닌 숫자로 나오는가?
- Writer가 실패하면 성공 코드 0으로 끝나지 않는가?
- 같은 run을 여러 번 호출해도 이전 플래그 값이 남지 않는가?

이 질문들을 표 기반 테스트 행으로 옮길 수 있습니다.
테스트는 시간을 기다리거나 실제 셸 환경을 바꿀 필요가 없습니다.
실제 바이너리 확인은 main의 연결과 종료 코드 전달을 검증합니다.

## 10. 흔한 실수와 해결

| 증상 | 원인 후보 | 확인·수정 |
| --- | --- | --- |
| 인자 없는 실행에서 패닉 | 길이 검사 없이 args[0] | len 확인 후 접근 |
| 옵션이 적용되지 않음 | 파싱 전에 값 사용 | 등록 뒤 Parse, 그 뒤 사용 |
| 두 번째 테스트가 이전 값을 사용 | 전역 FlagSet 재사용 | run마다 로컬 FlagSet |
| `-flag false`가 남은 인자가 됨 | bool 문법 혼동 | `-flag=false` |
| 파일 뒤 옵션이 적용되지 않음 | 첫 위치 인자에서 파싱 종료 | 옵션을 앞에 놓거나 계약 변경 |
| 미설정과 빈 설정을 구분하지 못함 | Getenv만 사용 | LookupEnv 존재 여부 확인 |
| 입력 프로그램이 멈춘 듯 보임 | stdin EOF를 기다림 | 파이프 입력과 종료 확인 |
| 결과 파일에 진행 문장이 들어감 | stdout에 진단 출력 | stderr로 분리 |
| 오류 메시지만 있고 성공으로 끝남 | 반환 코드 누락 | 모든 실패 경로 코드 확인 |
| defer가 실행되지 않음 | 처리 중 os.Exit | run 반환 뒤 main에서 Exit |
| 코드 2가 그대로 관찰되지 않음 | go run 상태를 확인함 | 빌드한 실행 파일 검사 |

### 10.1 모든 오류를 같은 상태로 뭉개지 않기

옵션을 잘못 준 사용자는 사용법을 확인해야 합니다.
출력 장치가 실패한 사용자는 대상이나 파이프를 확인해야 합니다.
같은 “실패”여도 원인을 구분하면 다음 행동이 달라집니다.

다만 많은 종료 코드를 무작정 만드는 것이 목적은 아닙니다.
프로그램을 호출하는 쪽에 의미 있는 작은 분류를 문서화합니다.
이 장은 0·1·2만 사용합니다.

### 10.2 도움말 출력의 책임

FlagSet은 파싱 오류 때 진단과 사용법을 출력할 수 있습니다.
그 뒤 호출자가 같은 오류를 다시 출력하면 문장이 중복될 수 있습니다.
예제 flags는 파싱 오류의 진단을 FlagSet에 맡깁니다.

설정 우선순위 같은 함수를 별도로 뽑으면
파싱 오류와 의미 오류의 진단 소유자를 정해야 합니다.
연습문제 해설에서는 학습을 위해 오류를 한 단계에서 표시하며,
더 큰 CLI에서는 사용자 오류 타입 등을 통해 출력 책임을 정리할 수 있습니다.

### 10.3 파이프의 성공 상태

Bash의 기본 파이프 상태는 마지막 명령의 상태를 따릅니다.
앞 명령 실패까지 자동으로 반영하는지 확인하려면 셸의 정책을 알아야 합니다.
[GNU Bash 파이프 문서](https://www.gnu.org/software/bash/manual/html_node/Pipelines.html)를 참고합니다.

이번 기본 실습의 앞 명령은 고정 문자열을 내보내는 printf입니다.
서비스 자동화 스크립트의 오류 정책 전체를 이 예제로 대신하지 않습니다.
Go 프로그램의 코드와 셸 전체 파이프의 상태를 구분해 검사합니다.

## 11. 연습문제와 검증

[연습문제](../../exercises/16-cli/README.md)는
기본 6문제와 선택 1문제입니다.
[해설](../../solutions/16-cli/README.md)에는 네 개의 실행 가능한 참고 구현이 있습니다.

| 문제 | 다룰 내용 |
| --- | --- |
| 1 | 인자와 bool 플래그 경계 예측 |
| 2 | 횟수가 제한된 인사 CLI |
| 3 | 결과와 진단을 분리하고 종료 코드 수정 |
| 4 | 플래그·환경 변수·기본값 우선순위 |
| 5 | 항목 수 제한이 있는 표준 입력 CLI |
| 6 | 입력 요약을 텍스트 또는 JSON으로 출력 |
| 선택 | main의 실행 파일 종료 코드를 직접 검사 |

저장소 루트에서 장 범위를 검사합니다.

```sh
go test ./examples/16-cli/... ./solutions/16-cli/...
go vet ./examples/16-cli/... ./solutions/16-cli/...
go build ./examples/16-cli/... ./solutions/16-cli/...
```

단위 테스트는 실제 표준 입력을 기다리지 않습니다.
테스트마다 strings.Reader와 bytes.Buffer를 전달합니다.
Writer·Reader 실패는 오류를 돌려주는 작은 구현으로 재현합니다.

## 12. 자기 점검

1. `os.Args[0]`과 처리 함수의 `args[0]`이 같은 값을 뜻하는가?
2. 로컬 FlagSet을 사용하는 이유는 무엇인가?
3. `-upper false`와 `-upper=false`는 어떻게 다른가?
4. 표준 flag는 첫 위치 인자 뒤의 플래그를 계속 읽는가?
5. LookupEnv의 bool은 무엇을 의미하는가?
6. stdin을 읽는 프로그램이 결과를 내기 전에 기다리는 이유는 무엇인가?
7. 결과와 진단을 서로 다른 스트림에 보내면 무엇이 가능한가?
8. os.Exit(run(...))에서 run의 defer와 main의 defer는 각각 어떻게 되는가?
9. go run으로 정확한 프로그램 종료 코드를 검사해도 되는가?
10. Writer 실패를 테스트하려면 실제 디스크 장애가 필요한가?

<details>
<summary>자기 점검 해설</summary>

1. main이 os.Args[1:]를 전달했다면 처리 함수의 0은 첫 사용자 인자다.
2. 실행과 테스트 사이의 플래그 상태를 독립적으로 만들기 위해서다.
3. 앞 표현의 false는 남은 위치 인자이고, 뒤 표현은 bool 값을 false로 지정한다.
4. 첫 위치 인자 또는 `--`에서 멈춘다.
5. 해당 이름이 환경에 존재하는지이며 값이 비어 있어도 true일 수 있다.
6. 입력 스트림이 끝나야 전체 처리를 확정할 수 있기 때문이다.
7. 결과만 파일·파이프로 전달하면서 진단을 따로 확인할 수 있다.
8. run이 반환할 때 그 defer는 실행되지만 Exit 뒤 main의 defer는 실행되지 않는다.
9. Go 도구의 상태와 내부 프로그램 상태가 다르므로 바이너리를 빌드해 확인한다.
10. 오류를 반환하는 io.Writer를 전달하면 재현할 수 있다.

</details>

## 13. 핵심 정리와 다음 장

CLI의 외부 계약은 옵션 이름만으로 완성되지 않습니다.
기본값과 범위, 환경 설정의 우선순위, 입력의 종료,
결과와 진단의 위치, 성공·실패 코드를 함께 정해야 합니다.

처리 함수를 인자·Reader·Writer에 의존하게 만들면
실제 터미널을 사용하지 않고도 중요한 경계를 테스트할 수 있습니다.
다음 장에서는 작업을 언제 중단해야 하는지 Context로 전달하는 방법을 배웁니다.

- 이전: [15. JSON·CSV·시간](../15-data-time/README.md)
- 다음: [17. Context](../17-context/README.md)
- [전체 목차](../README.md)
- [연습문제](../../exercises/16-cli/README.md)
- [해설](../../solutions/16-cli/README.md)

## 참고 자료

- [flag](https://pkg.go.dev/flag)
- [os.Args](https://pkg.go.dev/os#pkg-variables)
- [os.LookupEnv](https://pkg.go.dev/os#LookupEnv)
- [os.Exit](https://pkg.go.dev/os#Exit)
- [go run](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)
- [GNU Bash 리디렉션](https://www.gnu.org/s/bash/manual/html_node/Redirections.html)
- [GNU Bash 특수 매개변수](https://www.gnu.org/s/bash/manual/html_node/Special-Parameters.html)
- [PowerShell 자동 변수](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_automatic_variables)
- [PowerShell 문자 인코딩](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_character_encoding)

확인일: 2026-11-09. Go 1.27.3과 Linux Bash에서 코드·출력·종료 코드를 확인했습니다.
PowerShell 명령은 문서와 문법을 검토했으며 이 집필 환경에서는 실행하지 않았습니다.
