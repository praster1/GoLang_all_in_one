# 16장 연습문제 — 명령줄 프로그램

[본문](../../book/16-cli/README.md)을 읽고 기본 6문제와 선택 1문제를 풉니다.
CLI의 정답은 화면 문장만으로 결정하지 않습니다.
stdout, stderr, 종료 코드를 각각 확인합니다.

[해설](../../solutions/16-cli/README.md)은 요구사항을 구현한 뒤 비교합니다.
모든 명령은 저장소 루트에서 실행합니다.
표준 입력을 받는 명령에는 본문처럼 파이프 입력을 제공합니다.

## 준비

문제 2의 개인 파일을 준비합니다.
다른 문제는 이름이 다른 개인 디렉터리에 작성합니다.

Bash:

```sh
mkdir -p _practice/16-cli/repeat-cli
if [ ! -e _practice/16-cli/repeat-cli/main.go ]; then
  cp examples/16-cli/flags/main.go _practice/16-cli/repeat-cli/main.go
fi
```

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/16-cli/repeat-cli | Out-Null
if (-not (Test-Path _practice/16-cli/repeat-cli/main.go)) {
    Copy-Item examples/16-cli/flags/main.go _practice/16-cli/repeat-cli/main.go
}
```

이미 작성한 파일은 보존됩니다.
각 문제에서 `main`은 실제 os 경계를 처리 함수에 전달하는 역할만 맡깁니다.
단위 테스트에서는 main을 직접 호출하지 않습니다.

### 공통 종료 코드 정책

| 코드 | 의미 |
| --- | --- |
| 0 | 성공 또는 도움말 |
| 1 | I/O 등 실행 중 오류 |
| 2 | 사용법·옵션·입력 계약 오류 |

도움말은 stderr로 출력하고 코드 0으로 처리합니다.
정상 데이터는 stdout, 오류 진단은 stderr에 씁니다.
최종 진단 출력은 최선의 시도이며 그 쓰기까지 실패해도 재귀적으로 진단하지 않습니다.

## 문제 1. 인자와 플래그 예측

**수준: 확인**

본문의 `flags` 예제를 기준으로 판단합니다.
`name`의 기본값은 Go이고 `upper`의 기본값은 false입니다.
남은 위치 인자와 공백 이름은 허용하지 않습니다.

| 인자 | name | upper | 남은 인자 | 코드 |
| --- | --- | --- | --- | --- |
| 없음 | 직접 작성 | 직접 작성 | 직접 작성 | 직접 작성 |
| `-name learner -upper` | 직접 작성 | 직접 작성 | 직접 작성 | 직접 작성 |
| `-upper=false` | 직접 작성 | 직접 작성 | 직접 작성 | 직접 작성 |
| `-upper false` | 직접 작성 | 직접 작성 | 직접 작성 | 직접 작성 |
| `file -upper` | 직접 작성 | 직접 작성 | 직접 작성 | 직접 작성 |
| `-- -upper` | 직접 작성 | 직접 작성 | 직접 작성 | 직접 작성 |
| `-name " "` | 직접 작성 | 직접 작성 | 직접 작성 | 직접 작성 |
| `-h` | 처리 흐름 설명 | 처리 흐름 설명 | 처리 흐름 설명 | 직접 작성 |

### 완료 조건

- bool 플래그 뒤의 별도 `false`가 옵션 값이 아님을 설명합니다.
- 첫 위치 인자에서 flag 파싱이 멈춘다는 규칙을 적용합니다.
- `--`가 프로그램에 파일이나 문자열을 받으라고 강제하는 것은 아님을 설명합니다.
- 도움말을 인사 결과와 같은 stdout에 출력하지 않는 계약을 확인합니다.

### 힌트

파싱 결과와 의미 검증 결과를 나누어 적습니다.
파싱이 성공해도 남은 인자를 거부하는 조건에서 코드 2가 나올 수 있습니다.
[flag 문법](https://pkg.go.dev/flag#hdr-Command_line_flag_syntax)을 참고합니다.

## 문제 2. 반복 횟수가 제한된 인사 CLI

**수준: 변형**

개인 파일: `_practice/16-cli/repeat-cli/main.go`

```go
func run(args []string, stdout, stderr io.Writer) int
```

### 옵션

| 옵션 | 기본값 | 계약 |
| --- | --- | --- |
| `-name` | Go | 공백을 다듬은 뒤 비어 있지 않음 |
| `-times` | 1 | 1~5 정수 |
| `-h`, `-help` | 도움말 | 정상 코드 0 |

위치 인자는 허용하지 않습니다.
이름은 한 번 정리하고 검증한 뒤 사용합니다.
허용 범위 밖이면 인사 일부를 먼저 쓰지 않습니다.

실행 예:

```sh
go run ./_practice/16-cli/repeat-cli -name 민수 -times 2
```

예상 stdout:

```text
안녕하세요, 민수!
안녕하세요, 민수!
```

stderr는 비어 있고 종료 코드는 0입니다.

### 경계 검사

| 입력 | 기대 |
| --- | --- |
| 없음 | Go에게 한 번 |
| `-times 5` | 다섯 번 |
| `-times 0` | 코드 2, stdout 없음 |
| `-times 6` | 코드 2, stdout 없음 |
| `-times x` | 코드 2, 파싱 진단 |
| `-name " "` | 코드 2 |
| `extra` | 코드 2 |
| 모르는 옵션 | 코드 2 |
| 도움말 | 코드 0, stdout 없음 |
| 실패하는 stdout Writer | 코드 1 |

### 단계별 힌트

1. 매번 새로운 FlagSet을 만듭니다.
2. Parse 오류 중 ErrHelp를 별도로 처리합니다.
3. 모든 입력 조건을 검사한 뒤 반복합니다.
4. 쓰기 실패를 발견하면 즉시 코드 1을 반환합니다.

[참고 구현](../../solutions/16-cli/repeat-cli/main.go) ·
[테스트](../../solutions/16-cli/repeat-cli/main_test.go)

## 문제 3. 결과·진단·종료 코드 분리

**수준: 확인·변형**

다음은 의도적으로 잘못된 전체 CLI입니다.
컴파일은 되지만 외부 계약이 틀렸습니다.
개인 폴더에서 별도 실행 파일로 빌드하여 확인합니다.

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stdout, "사용법: echo-one 값")
		return 2
	}
	fmt.Fprintln(stdout, args[0])
	return 0
}

func main() {
	run(os.Args[1:], os.Stdout, os.Stderr)
}
```

### 찾아야 할 문제

- 사용법 오류 진단이 어느 스트림으로 가는가?
- run이 2를 반환해도 실제 프로그램은 어떤 코드로 끝나는가?
- stdout이 쓰기 오류를 반환하면 코드가 바뀌는가?

### 수정 계약

- 인자가 정확히 하나면 그 문자열과 개행을 stdout에 씁니다.
- 인자가 없거나 둘 이상이면 사용법을 stderr에 쓰고 코드 2입니다.
- stdout 쓰기 오류는 stderr에 진단을 쓰고 코드 1입니다.
- main은 run의 반환값을 실제 프로세스 종료 코드로 전달합니다.

### 완료 조건

실행 파일에 인자를 주지 않았을 때 stdout은 비어 있어야 합니다.
stderr에는 사용법이 있어야 하고 종료 코드는 2여야 합니다.
단순히 화면에 사용법이 보인다는 사실로 완료하지 않습니다.

### 힌트

Go 함수의 정수 반환값을 무시하면 프로세스 종료 코드가 자동으로 바뀌지 않습니다.
`os.Exit`는 defer를 실행하지 않으므로 처리 함수가 반환한 뒤 main에서 호출합니다.
[os.Exit](https://pkg.go.dev/os#Exit)를 참고합니다.

## 문제 4. 명시적인 설정 우선순위

**수준: 응용**

개인 파일: `_practice/16-cli/env-priority/main.go`

이름을 다음 순서로 선택합니다.

1. **명시적으로 지정된** `-name` 플래그.
2. `STUDY_NAME` 환경 변수에 값이 존재하면 그 값.
3. 기본값 Go.

선택한 값의 앞뒤 공백을 제거합니다.
선택한 값이 빈 문자열이면 코드 2입니다.
빈 값을 이유로 더 낮은 우선순위로 넘어가지 않습니다.

### 함수 경계

```go
func resolveName(
	args []string,
	lookup func(string) (string, bool),
	diagnostics io.Writer,
) (string, error)
```

실제 CLI용 run은 이 함수를 호출하여 출력·종료 코드를 처리합니다.
lookup은 실제 실행에서 `os.LookupEnv`, 테스트에서는 가짜 함수를 사용합니다.
함수는 지정한 환경 변수 이름 하나만 조회해야 합니다.

### 정상 출력

`-name CLI`를 주었다면 환경 변수와 무관하게 다음과 같습니다.

```text
이름: CLI
```

### 우선순위 표

| 플래그 | 환경 변수 | 선택 |
| --- | --- | --- |
| 없음 | 없음 | Go |
| 없음 | Env | Env |
| CLI | Env | CLI |
| Go를 명시 | Env | Go |
| 빈 값을 명시 | Env | 오류 |
| 없음 | 존재하지만 빈 값 | 오류 |
| CLI | 존재하지만 빈 값 | CLI |
| 공백을 명시 | Env | 오류 |

기본값과 같은 Go를 명시해도 사용자 선택입니다.
`*name != "Go"` 같은 값 비교만으로 플래그 지정 여부를 판단하지 않습니다.

### 힌트

`FlagSet.Visit`은 실제로 지정된 플래그를 순회합니다.
이 콜백 안에서 `name`이 지정됐는지 bool로 기록할 수 있습니다.
[Visit](https://pkg.go.dev/flag#FlagSet.Visit)을 참고합니다.

파싱을 먼저 수행하면 도움말 요청 때 환경 변수의 잘못된 값 때문에
도움말이 실패하는 상황을 피할 수 있습니다.
설정 우선순위를 코드의 우연한 호출 순서에 맡기지 않습니다.

[참고 구현](../../solutions/16-cli/env-priority/main.go) ·
[테스트](../../solutions/16-cli/env-priority/main_test.go)

## 문제 5. 제한이 있는 항목 수 CLI

**수준: 변형**

개인 파일: `_practice/16-cli/line-cli/main.go`

```go
func run(args []string, source io.Reader, stdout, stderr io.Writer) int
```

### 계약

- `-limit` 기본값은 10이고 범위는 0~100입니다.
- 위치 인자는 허용하지 않습니다.
- 표준 입력을 줄로 읽고 공백을 다듬었을 때 빈 줄은 세지 않습니다.
- 비어 있지 않은 항목 수가 limit을 넘으면 코드 2입니다.
- Scanner 오류는 코드 1입니다.
- 정상 출력은 `항목: N`과 개행입니다.
- 모든 입력을 확인하기 전에는 결과를 stdout에 쓰지 않습니다.

본문 app의 limit은 1~100이었습니다.
이 문제는 0을 허용하고 0개 입력에서 성공해야 합니다.
줄을 잘라 버리는 기능이 아니라 제한 위반을 찾는 기능입니다.

### Bash 실행 예

```sh
printf 'Go\n\nTest\n' | go run ./_practice/16-cli/line-cli -limit 2
```

### PowerShell 실행 예

```powershell
"Go", "", "Test" | go run ./_practice/16-cli/line-cli -limit 2
```

예상 결과:

```text
항목: 2
```

### 경계표

| 옵션·입력 | 기대 |
| --- | --- |
| 기본값, 빈 입력 | 0개, 코드 0 |
| limit 0, 공백 줄만 | 0개, 코드 0 |
| limit 0, Go 한 줄 | 코드 2 |
| limit 2, 두 항목 | 2개, 코드 0 |
| limit 1, 두 항목 | 코드 2, stdout 없음 |
| limit 100, 100개 | 성공 |
| limit 100, 101개 | 코드 2 |
| limit -1 또는 101 | 코드 2 |
| Reader 실패 | 코드 1 |
| stdout Writer 실패 | 코드 1 |

Scanner의 최대 토큰 크기는 본문처럼 128KiB로 정합니다.
전체 입력 바이트 제한은 이 문제의 계약에 포함하지 않습니다.
너무 긴 한 줄의 오류도 테스트에서 확인합니다.

[참고 구현](../../solutions/16-cli/line-cli/main.go) ·
[테스트](../../solutions/16-cli/line-cli/main_test.go)

## 문제 6. 텍스트·JSON 학습 요약 CLI

**수준: 응용**

개인 파일: `_practice/16-cli/study-cli/main.go`

문제 5와 같은 run 경계를 사용합니다.
표준 입력의 비어 있지 않은 각 줄을 하나의 학습 항목으로 봅니다.
항목별 시간은 모든 항목에 같은 `-minutes` 값을 적용합니다.

### 옵션과 제한

| 항목 | 계약 |
| --- | --- |
| `-minutes` | 기본 25, 1~120 |
| `-json` | 기본 false |
| 입력 크기 | 전체 4096바이트까지 |
| 항목 수 | 비어 있지 않은 줄 최대 100개 |
| 빈 입력 | 0개·0분으로 정상 |
| 위치 인자 | 거부 |

한글 글자 수 대신 UTF-8 바이트 수로 크기를 확인합니다.
정상 처리 전에 전체 입력 검증을 끝냅니다.
도움말 요청은 입력을 읽기 전에 처리합니다.

### 기본 출력

Bash:

```sh
printf 'Go\nTest\n' | go run ./_practice/16-cli/study-cli
```

PowerShell:

```powershell
"Go", "Test" | go run ./_practice/16-cli/study-cli
```

stdout:

```text
항목: 2
합계: 50분
```

### JSON 출력

Bash:

```sh
printf 'Go\nTest\n' | go run ./_practice/16-cli/study-cli -minutes 30 -json
```

PowerShell:

```powershell
"Go", "Test" | go run ./_practice/16-cli/study-cli -minutes 30 -json
```

stdout:

```json
{"count":2,"total_minutes":60}
```

마지막 개행도 출력 계약에 포함합니다.
진단 문장이나 프롬프트를 JSON 앞뒤에 추가하지 않습니다.

### 오류 정책

| 상황 | 코드 | stdout |
| --- | --- | --- |
| 분량 범위·타입 오류 | 2 | 없음 |
| 4097바이트 입력 | 2 | 없음 |
| 101개 항목 | 2 | 없음 |
| Reader 실패 | 1 | 없음 |
| 출력 실패 | 1 | 일부가 쓰였을 수 있음 |
| 도움말 | 0 | 없음 |

### 힌트

1. 플래그 파싱과 검증부터 합니다.
2. 입력은 `LimitReader(..., 4097)`로 읽고 4096과 비교합니다.
3. 크기가 검증된 바이트를 Scanner에 연결합니다.
4. 합계는 최대 100×120=12000분입니다.
5. JSON은 구조체와 Encoder를 사용합니다.
6. 각 출력 분기의 오류를 같은 곳에서 처리합니다.

[참고 구현](../../solutions/16-cli/study-cli/main.go) ·
[테스트](../../solutions/16-cli/study-cli/main_test.go)

## 선택 문제. 실행 파일 계약 확인

**수준: 응용**

단위 테스트의 반환 코드와 실제 프로세스 종료 코드가 연결되는지 확인합니다.
개인 구현 하나를 `bin` 아래에 빌드합니다.
기존에 보존할 파일이 있으면 다른 출력 이름을 선택합니다.

Bash:

```sh
mkdir -p bin
go build -o bin/study-cli ./solutions/16-cli/study-cli
printf 'Go\nTest\n' | ./bin/study-cli -json
printf '종료 코드: %s\n' "$?"
./bin/study-cli -minutes 0
printf '종료 코드: %s\n' "$?"
```

첫 실행은 JSON과 코드 0입니다.
두 번째는 입력을 기다리지 않고 옵션 오류로 코드 2여야 합니다.

PowerShell:

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o bin/study-cli.exe ./solutions/16-cli/study-cli
"Go", "Test" | .\bin\study-cli.exe -json
"종료 코드: $LASTEXITCODE"
.\bin\study-cli.exe -minutes 0
"종료 코드: $LASTEXITCODE"
```

### 완료 조건

- `go run`의 상태를 프로그램의 정확한 코드로 기록하지 않습니다.
- 성공과 옵션 오류의 stdout·stderr를 구분합니다.
- 잘못된 옵션일 때 stdin을 기다리지 않는지 확인합니다.
- main과 run 중 어디서 실제 프로세스가 종료되는지 설명합니다.

## 완료 점검

- 인자 경계를 다시 공백으로 쪼개지 않았다.
- 로컬 FlagSet을 썼다.
- 도움말과 오류를 구분했다.
- 환경 변수의 미설정·빈 값을 테스트했다.
- stdout에는 결과만 썼다.
- 입력 끝과 Scanner 오류를 확인했다.
- 출력 오류에 성공 코드를 반환하지 않았다.
- 실제 바이너리의 종료 코드를 확인했다.

[해설로 이동](../../solutions/16-cli/README.md) ·
[본문으로 돌아가기](../../book/16-cli/README.md)

확인일: 2026-11-09. Go 1.27.3 기준입니다.
