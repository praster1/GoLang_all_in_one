# 16장 해설 — 명령줄 프로그램

[연습문제](../../exercises/16-cli/README.md)의 참고 구현입니다.
외부 계약을 stdout·stderr·종료 코드로 나누어 검토합니다.
실제 명령을 실행할 때는 저장소 루트에서 시작합니다.

| 문제 | 구현 | 핵심 |
| --- | --- | --- |
| 2 | [repeat-cli](repeat-cli/main.go) | 반복 전 입력 검증 |
| 4 | [env-priority](env-priority/main.go) | 명시적 설정의 우선순위 |
| 5 | [line-cli](line-cli/main.go) | 입력 제한과 EOF |
| 6 | [study-cli](study-cli/main.go) | 크기 제한·JSON·종료 코드 |

이 장의 테스트는 main을 직접 호출하지 않습니다.
실제 `os.Exit`로 테스트 프로세스를 끝내지 않고 run의 반환값을 검사합니다.
main의 연결은 빌드한 실행 파일을 실제로 실행해 별도로 확인합니다.

## 문제 1. 파싱과 의미 검증

| 인자 | name | upper | 남은 인자 | 코드 |
| --- | --- | --- | --- | --- |
| 없음 | Go | false | 없음 | 0 |
| `-name learner -upper` | learner | true | 없음 | 0 |
| `-upper=false` | Go | false | 없음 | 0 |
| `-upper false` | Go | true | false | 2 |
| `file -upper` | Go | false | file, -upper | 2 |
| `-- -upper` | Go | false | -upper | 2 |
| `-name " "` | 공백 | false | 없음 | 2 |
| `-h` | 파싱이 도움말 요청 반환 | 본 처리 없음 | 본 처리 없음 | 0 |

불리언 플래그는 이름만으로 true가 됩니다.
바로 다음의 별도 문자열을 bool 값으로 소비하지 않습니다.
`false`를 값으로 정하려면 `-upper=false`를 사용합니다.

첫 위치 인자 뒤의 `-upper`는 표준 flag의 파싱 대상이 아닙니다.
`--`도 마찬가지로 파싱을 끝냅니다.
남은 인자를 허용하는지는 CLI 자체가 결정합니다.

`-name " "`는 문법적으로 정상 문자열입니다.
공백을 다듬은 뒤 비었는지를 확인하는 의미 검증에서 거부됩니다.
[공식 flag 문법](https://pkg.go.dev/flag#hdr-Command_line_flag_syntax)을 확인합니다.

## 문제 2. 인사하기 전에 조건을 확정하기

소스: [repeat-cli/main.go](repeat-cli/main.go)

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
	flags := flag.NewFlagSet("repeat", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "Go", "인사할 이름")
	times := flags.Int("times", 1, "반복 횟수: 1~5")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	value := strings.TrimSpace(*name)
	if flags.NArg() != 0 || value == "" || *times < 1 || *times > 5 {
		fmt.Fprintln(stderr, "오류: 이름과 횟수(1~5)를 확인하세요")
		return 2
	}
	for range *times {
		if _, err := fmt.Fprintf(stdout, "안녕하세요, %s!\n", value); err != nil {
			fmt.Fprintln(stderr, "출력 오류:", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```

실행:

```sh
go run ./solutions/16-cli/repeat-cli -name 민수 -times 2
```

예상 stdout:

```text
안녕하세요, 민수!
안녕하세요, 민수!
```

stderr는 비어 있습니다.
0회는 아무것도 출력하는 정상 동작이 아니라 이 문제의 범위 오류입니다.
5회는 상한을 포함하므로 정상입니다.

### 같은 run을 여러 번 호출하기

플래그 값을 함수 안에 두고 FlagSet도 새로 만듭니다.
테스트가 먼저 `-times 5`를 실행하고 다음에 기본값을 실행해도
기본값은 다시 1입니다.

전역 플래그를 초기화하거나 os.Args를 덮어쓸 필요가 없습니다.
입력 문자열 목록을 전달하는 것으로 충분합니다.

### 출력 실패의 처리

반복 첫 출력은 성공하고 다음 출력이 실패할 수도 있습니다.
이때 이미 쓴 인사를 되돌리는 기능은 없습니다.
함수는 다음 반복을 중단하고 코드 1을 반환합니다.

[테스트](repeat-cli/main_test.go)는 기본값·공백 정리·상한과
0/6/x·모르는 옵션·위치 인자·도움말·Writer 실패를 검사합니다.
반복 횟수만 세고 stderr와 코드를 무시하지 않습니다.

## 문제 3. 정수 반환과 프로세스 상태

잘못된 프로그램은 인자 없이 실행하면
사용법을 stdout에 쓰고 프로세스는 코드 0으로 끝납니다.
run의 정수 2를 main이 사용하지 않았기 때문입니다.

수정한 독립 전체 프로그램은 다음과 같습니다.

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "사용법: echo-one 값")
		return 2
	}
	if _, err := fmt.Fprintln(stdout, args[0]); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```

인자가 `Go`이면 stdout은 `Go`와 개행이고 코드 0입니다.
인자가 없으면 stdout은 비어 있고 stderr에 사용법이 나오며 코드 2입니다.
인자가 두 개여도 같은 사용법 오류입니다.

### 왜 run 안에서 Exit하지 않는가?

run의 반환을 통해 그 함수의 defer가 끝날 기회를 얻습니다.
main은 run이 마친 뒤에 상태를 운영체제에 전달합니다.
main에 따로 등록한 defer는 Exit로 실행되지 않는 점을 구분합니다.

`fmt.Fprintln(stderr, ...)`의 마지막 진단 실패를 무한히 다시 보고하지 않습니다.
핵심 실패는 이미 코드 1 또는 2로 전달합니다.
stdout 결과 쓰기 오류는 성공으로 처리하지 않습니다.

[os.Exit](https://pkg.go.dev/os#Exit)

## 문제 4. 값과 지정 여부를 분리하기

소스: [env-priority/main.go](env-priority/main.go)

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

func resolveName(args []string, lookup func(string) (string, bool), diagnostics io.Writer) (string, error) {
	flags := flag.NewFlagSet("name", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	name := flags.String("name", "", "환경 변수보다 우선하는 이름")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	if flags.NArg() != 0 {
		return "", errors.New("위치 인자는 허용하지 않습니다")
	}
	specified := false
	flags.Visit(func(item *flag.Flag) {
		if item.Name == "name" {
			specified = true
		}
	})
	value := "Go"
	if specified {
		value = *name
	} else if environment, exists := lookup("STUDY_NAME"); exists {
		value = environment
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("이름은 비어 있을 수 없습니다")
	}
	return value, nil
}

func run(args []string, lookup func(string) (string, bool), stdout, stderr io.Writer) int {
	name, err := resolveName(args, lookup, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(stderr, "설정 오류:", err)
		return 2
	}
	if _, err := fmt.Fprintln(stdout, "이름:", name); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}
```

환경 변수를 건드리지 않고 CLI 우선순위를 확인할 수 있습니다.

```sh
go run ./solutions/16-cli/env-priority -name CLI
```

예상 stdout:

```text
이름: CLI
```

Bash에서 환경 변수 우선순위만 확인하려면 다음과 같이 실행합니다.

```sh
STUDY_NAME=Env go run ./solutions/16-cli/env-priority
```

예상 stdout:

```text
이름: Env
```

PowerShell 환경 변수 설정·복원은 [본문 4절](../../book/16-cli/README.md)을 참고합니다.
단위 테스트는 환경을 실제로 변경하지 않습니다.

### `Visit`이 필요한 이유

기본값을 Go로 두고 결과가 Go인지 비교하면
사용자가 `-name Go`를 명시했는지 알 수 없습니다.
명시적 Go도 환경 변수 Env보다 우선해야 합니다.

`Visit`은 실제로 설정된 플래그를 알려 줍니다.
값 자체와 지정 여부를 별도로 보관하여 이 요구사항을 충족합니다.
[FlagSet.Visit](https://pkg.go.dev/flag#FlagSet.Visit)을 참고합니다.

| 후보 | 존재 판단 | 값 검증 시점 |
| --- | --- | --- |
| 플래그 | Visit에서 name 발견 | 가장 우선으로 선택한 뒤 |
| 환경 변수 | LookupEnv의 bool | 플래그가 없을 때 선택한 뒤 |
| 기본값 | 앞의 두 후보 없음 | 이미 정해진 Go |

### 빈 플래그는 왜 오류인가?

사용자가 `-name=`을 명시한 것은 플래그가 없는 상태와 다릅니다.
선택한 값이 빈 문자열이면 오류이며 Env로 돌아가지 않습니다.
이 정책은 잘못된 사용자 선택을 조용히 덮지 않도록 합니다.

반대로 환경 변수만 비어 있어도 명시적 CLI가 유효하면 성공입니다.
선택하지 않은 낮은 우선순위의 값을 먼저 검증하지 않았기 때문입니다.

### 진단 소유권

FlagSet은 파싱 오류를 diagnostics에 출력합니다.
run은 resolveName의 의미 오류도 `설정 오류`로 표시합니다.
이 작은 구현은 파싱 실패 때 요약 문장이 추가될 수 있습니다.

큰 CLI에서 중복 진단을 피하려면 오류 종류와 출력 책임을 더 명확히 나눌 수 있습니다.
여기서 테스트하는 계약은 정확한 도움말 공백이 아니라
결과 위치·코드·선택된 값·오류 여부입니다.

[테스트](env-priority/main_test.go)는
미설정·환경 설정·명시적 기본값·빈 값·공백·도움말·출력 실패를 확인합니다.

## 문제 5. 제한 초과를 정상 결과로 숨기지 않기

소스: [line-cli/main.go](line-cli/main.go)

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
	flags := flag.NewFlagSet("count-lines", flag.ContinueOnError)
	flags.SetOutput(stderr)
	limit := flags.Int("limit", 10, "허용할 최대 항목 수: 0~100")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *limit < 0 || *limit > 100 {
		fmt.Fprintln(stderr, "오류: limit은 0~100이어야 합니다")
		return 2
	}
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		count++
		if count > *limit {
			fmt.Fprintln(stderr, "오류: 항목 수 제한을 넘었습니다")
			return 2
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	if _, err := fmt.Fprintf(stdout, "항목: %d\n", count); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

Bash:

```sh
printf 'Go\n\nTest\n' | go run ./solutions/16-cli/line-cli -limit 2
```

PowerShell:

```powershell
"Go", "", "Test" | go run ./solutions/16-cli/line-cli -limit 2
```

예상 stdout:

```text
항목: 2
```

### limit 0의 의미

limit 0은 입력 항목을 하나도 허용하지 않는 정상 설정입니다.
빈 입력이나 공백 줄만 있으면 0개로 성공합니다.
비어 있지 않은 줄을 처음 읽으면 제한 초과입니다.

입력 검증 단계의 실패이므로 코드 2입니다.
Scanner가 긴 토큰이나 Reader 오류로 실패하면 실행 오류 코드 1입니다.
정상 EOF는 오류가 아니므로 카운트를 출력합니다.

### 출력 시점

합계 문자열은 Scanner가 정상 종료한 뒤 한 번 출력합니다.
따라서 두 번째 항목이 제한을 넘었을 때 첫 항목의 결과가 남지 않습니다.
결과가 하나의 카운트인 프로그램에 맞는 출력 시점입니다.

[테스트](line-cli/main_test.go)는
0·2·100·101 항목, CRLF·빈 줄, 숫자 파싱 실패,
Reader·Writer 오류와 긴 줄을 검사합니다.
전체 입력 바이트 제한을 제공한다고 주장하지 않습니다.

## 문제 6. 크기 제한과 출력 형식 선택

소스: [study-cli/main.go](study-cli/main.go)

```go
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Summary struct {
	Count        int `json:"count"`
	TotalMinutes int `json:"total_minutes"`
}

func run(args []string, source io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("study", flag.ContinueOnError)
	flags.SetOutput(stderr)
	minutes := flags.Int("minutes", 25, "항목별 분: 1~120")
	jsonOutput := flags.Bool("json", false, "JSON으로 출력")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *minutes < 1 || *minutes > 120 {
		fmt.Fprintln(stderr, "오류: minutes는 1~120이어야 합니다")
		return 2
	}
	data, err := io.ReadAll(io.LimitReader(source, 4097))
	if err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	if len(data) > 4096 {
		fmt.Fprintln(stderr, "오류: 입력은 4096바이트까지 허용합니다")
		return 2
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
		if count > 100 {
			fmt.Fprintln(stderr, "오류: 항목은 100개까지 허용합니다")
			return 2
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	summary := Summary{Count: count, TotalMinutes: count * *minutes}
	if *jsonOutput {
		err = json.NewEncoder(stdout).Encode(summary)
	} else {
		_, err = fmt.Fprintf(stdout, "항목: %d\n합계: %d분\n", summary.Count, summary.TotalMinutes)
	}
	if err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

Bash:

```sh
printf 'Go\nTest\n' | go run ./solutions/16-cli/study-cli
```

PowerShell:

```powershell
"Go", "Test" | go run ./solutions/16-cli/study-cli
```

예상 stdout:

```text
항목: 2
합계: 50분
```

JSON 출력 명령:

```sh
printf 'Go\nTest\n' | go run ./solutions/16-cli/study-cli -minutes 30 -json
```

위 파이프 명령은 Bash용입니다.
PowerShell에서는 `"Go", "Test"`를 파이프 왼쪽에 사용합니다.

예상 stdout:

```json
{"count":2,"total_minutes":60}
```

### 검증 순서

| 순서 | 검증 | 실패 코드 |
| --- | --- | --- |
| 1 | flag 문법과 위치 인자 | 2 |
| 2 | 분량 1~120 | 2 |
| 3 | Reader 읽기 | 1 |
| 4 | 4096바이트 제한 | 2 |
| 5 | 비어 있지 않은 항목 100개 | 2 |
| 6 | 결과 쓰기 | 1 |

도움말은 첫 단계에서 0으로 끝나므로 stdin을 읽지 않습니다.
잘못된 분량도 입력을 기다리지 않고 거부됩니다.
사용법이 틀린 명령에 파일 전체를 먼저 읽을 필요가 없습니다.

### Scanner 버퍼를 따로 키우지 않은 이유

앞 단계에서 전체 입력이 최대 4096바이트임을 확인했습니다.
이 조건 아래에서는 기본 Scanner의 토큰 제한보다 작습니다.
이 가정을 바꾸어 더 큰 입력을 받게 되면 Scanner 제한도 다시 검토합니다.

“항상 기본 버퍼로 충분하다”는 일반 규칙으로 확대하지 않습니다.
설계의 앞 단계에서 확보한 상한이 뒤 단계의 선택을 정당화합니다.

### 수치와 출력 타입

최대 합계는 100×120=12000분입니다.
JSON 구조체의 필드는 int이므로 숫자 값으로 출력됩니다.
문자열을 손으로 이어 JSON을 만들지 않아 인용·이스케이프 오류를 피합니다.

`Encoder.Encode`가 마지막 개행을 붙입니다.
추가 Println을 호출하지 않습니다.
텍스트 출력도 정해진 두 줄 뒤에 개행을 포함합니다.

### 실패하면 어느 출력이 남는가?

잘못된 입력으로 거부할 때는 stdout이 비어 있습니다.
Writer 실패는 일부 결과를 쓴 뒤 발생할 수도 있습니다.
이를 입력 검증 실패와 같은 원자적 보장으로 설명하지 않습니다.

[테스트](study-cli/main_test.go)는 4096/4097바이트, 100/101항목,
1/120분, 빈 입력, bool 문법, 도움말, 두 출력 형식의 실패를 검사합니다.
외부 파일이나 현재 환경을 준비하지 않아도 경계를 재현할 수 있습니다.

## 선택 문제. 단위 함수와 실행 파일 연결 확인

Bash 예시의 첫 실행은 다음 stdout을 냅니다.

```json
{"count":2,"total_minutes":50}
```

바로 확인한 코드는 0입니다.
`-minutes 0`은 stdin을 읽기 전에 코드 2로 끝납니다.
결과 stdout은 비어 있고 범위 진단은 stderr에 있습니다.

실행 파일을 빌드하는 이유는 `go run`의 상태가
내부 프로그램의 종료 코드와 동일한 계약이 아니기 때문입니다.
[go run 문서](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)를 참고합니다.

### 검증을 두 층으로 나누기

| 검사 | 대상 | 목적 |
| --- | --- | --- |
| 직접 run 호출 | 인자·Reader·Writer와 int | 많은 경계 입력을 빠르게 확인 |
| 실행 파일 실행 | 프로세스 인자·스트림·종료 | main의 실제 연결 확인 |

단위 테스트에서 매번 Go를 빌드할 필요는 없습니다.
반대로 함수 반환값만 검사하면 main이 그 값을 무시하는 버그를 놓칠 수 있습니다.
이 장에서는 두 방식이 확인하는 책임을 나누었습니다.

## 장 범위 검사

```sh
go test ./examples/16-cli/... ./solutions/16-cli/...
go vet ./examples/16-cli/... ./solutions/16-cli/...
go build ./examples/16-cli/... ./solutions/16-cli/...
```

테스트는 실행 환경의 STUDY_NAME이나 실제 os.Args를 변경하지 않습니다.
입력을 기다리는 테스트도 없습니다.
각 run에 이미 끝이 있는 Reader를 전달합니다.

PowerShell 실행 문법은 본문과 공식 문서를 검토했지만
이 집필 환경에는 PowerShell이 설치되어 있지 않습니다.
실제 검증한 프로세스 실행은 Go 1.27.3·Linux Bash 기준입니다.

[본문](../../book/16-cli/README.md) ·
[연습문제](../../exercises/16-cli/README.md)

확인일: 2026-11-09.
