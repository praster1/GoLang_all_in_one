# 15장 연습문제 — JSON·CSV·시간

[본문](../../book/15-data-time/README.md)에서 익힌 API를
입력 계약과 오류 처리에 맞춰 조합합니다.
기본 6문제와 선택 1문제입니다.
정답은 [해설](../../solutions/15-data-time/README.md)에 분리되어 있습니다.

## 시작하기

모든 명령은 저장소 루트에서 실행합니다.
기존 예제·해설을 직접 수정하지 않고 개인 폴더를 사용합니다.
프로그램마다 `main` 함수가 필요하면 서로 다른 디렉터리에 둡니다.

macOS/Linux/WSL Bash:

```sh
mkdir -p _practice/15-data-time/decode-config
if [ ! -e _practice/15-data-time/decode-config/main.go ]; then
  cp examples/15-data-time/json-decode/main.go _practice/15-data-time/decode-config/main.go
fi
```

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/15-data-time/decode-config | Out-Null
if (-not (Test-Path _practice/15-data-time/decode-config/main.go)) {
    Copy-Item examples/15-data-time/json-decode/main.go _practice/15-data-time/decode-config/main.go
}
```

이 준비 명령은 이미 작성한 파일을 보존합니다.
복사한 예제는 시작점이며 문제 2의 요구사항에 맞게 직접 고칩니다.
문제 4·5·6도 별도 개인 폴더에서 작성합니다.

```sh
gofmt -w _practice/15-data-time/decode-config/main.go
go run ./_practice/15-data-time/decode-config
```

테스트를 작성할 때 성공 결과뿐 아니라 오류일 때의 반환값도 확인합니다.
파일이 필요하지 않은 함수는 `strings.NewReader`와 `bytes.Buffer`로 검사합니다.
실제 파일을 만들 경우에는 14장의 임시 디렉터리 방식을 사용합니다.

## 문제 1. 상태를 잃는 JSON 필드

**수준: 확인**

다음 두 구조체는 같은 상태를 표현하지 못합니다.
각각을 새로운 제로 값으로 만들고 같은 JSON을 읽는다고 가정합니다.

```go
type Plain struct {
	Enabled bool `json:"enabled,omitempty"`
}

type Optional struct {
	Enabled *bool `json:"enabled,omitempty"`
}
```

`encoding/json` v1의 동작을 기준으로 답합니다.
태그 조각만으로는 독립 실행할 수 없습니다.
확인할 때는 본문의 `json-presence` 예제를 개인 폴더에 복사해 구조체를 바꿉니다.

| 입력 | Plain의 값 | Optional의 값 | 다시 인코딩할 때 차이 |
| --- | --- | --- | --- |
| `{}` | 직접 작성 | 직접 작성 | 직접 작성 |
| `{"enabled":false}` | 직접 작성 | 직접 작성 | 직접 작성 |
| `{"enabled":true}` | 직접 작성 | 직접 작성 | 직접 작성 |
| `{"enabled":null}` | 직접 작성 | 직접 작성 | 직접 작성 |

### 완료 조건

- false와 nil 포인터를 다른 상태로 설명한다.
- Optional도 누락과 null을 구분하지 못한다는 점을 적는다.
- nil 역참조를 피하는 검사 순서를 적는다.
- `{"enabled":"false"}`가 왜 다른 입력인지 설명한다.
- v2로 import만 바꾸고 같은 생략 규칙을 가정하면 안 되는 이유를 적는다.

### 힌트

먼저 포인터가 nil인지 확인합니다.
nil이 아니면 포인터가 가리키는 bool을 읽을 수 있습니다.
`omitempty`가 검사하는 값이 bool인지 포인터인지 구분합니다.

## 문제 2. 크기가 제한된 설정 JSON

**수준: 변형**

개인 파일: `_practice/15-data-time/decode-config/main.go`

다음 타입과 함수를 구현합니다.
타입과 함수 선언은 요구사항이며 함수 본문은 직접 작성합니다.

```go
type Config struct {
	Course       string `json:"course"`
	DailyMinutes int    `json:"daily_minutes"`
}

func readConfig(source io.Reader) (Config, error)
```

### 계약

| 항목 | 요구사항 |
| --- | --- |
| 최대 크기 | 4096바이트, 공백 포함 |
| 최상위 값 | JSON 하나 |
| 모르는 필드 | 거부 |
| Course | 앞뒤 공백 제거, 빈 값 거부 |
| DailyMinutes | 1~120, 양 끝 포함 |
| 오류 반환 | `Config{}`와 오류 |
| 자원 | 전달받은 Reader를 닫지 않음 |

v1의 대소문자 매칭과 중복 이름을 별도로 제한하는 문제는 아닙니다.
`DisallowUnknownFields`만으로 그런 규칙을 구현했다고 쓰지 않습니다.
외부 Reader 오류는 `%w`로 원인을 보존합니다.

### 실행 예

입력:

```json
{"course":"Go","daily_minutes":25}
```

정상 출력:

```text
Go: 하루 25분
```

### 경계 확인표

| 입력 변경 | 기대 |
| --- | --- |
| Course `" Go "` | `"Go"`로 정리 |
| DailyMinutes 1 | 성공 |
| DailyMinutes 120 | 성공 |
| DailyMinutes 0 또는 121 | 오류 |
| Course 누락 또는 공백 | 오류 |
| `null` | 오류 |
| 잘못된 JSON | 오류 |
| `memo` 필드 추가 | 오류 |
| 객체 뒤에 `{}` 추가 | 오류 |
| 정상 객체 뒤 공백을 채워 총 4096바이트 | 성공 |
| 같은 방식으로 총 4097바이트 | 오류 |
| Reader 자체의 읽기 실패 | 오류 원인 보존 |

총 길이는 UTF-8 바이트 수입니다.
한글 제목의 글자 수를 4096과 비교하지 않습니다.

### 단계별 힌트

1. 먼저 `io.LimitReader`로 최대보다 한 바이트 더 읽습니다.
2. 길이를 확인한 뒤 메모리의 바이트를 디코더에 연결합니다.
3. 첫 `Decode`, 두 번째 `Decode`, 의미 검증을 분리합니다.
4. 오류 경로마다 반쯤 채워진 Config를 성공처럼 반환하지 않습니다.

[참고 구현](../../solutions/15-data-time/decode-config/main.go)과
[테스트](../../solutions/15-data-time/decode-config/main_test.go)는 시도 뒤 확인합니다.

## 문제 3. CSV의 쉼표와 버퍼

**수준: 확인·변형**

다음은 실행되지만 CSV 해석과 출력 확인이 잘못된 전체 프로그램입니다.
개인 폴더에 저장하고 출력부터 예측합니다.
기본 예제에 이 잘못된 구현을 덮어쓰지 않습니다.

```go
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func run() error {
	input := `"Go, basics",25`
	parts := strings.Split(input, ",")
	fmt.Println("직접 나눈 필드:", len(parts))
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write([]string{"Go, basics", "25"}); err != nil {
		return err
	}
	fmt.Println("Flush 전 바이트:", output.Len())
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

### 할 일

- 첫 결과의 필드 개수가 왜 실제 CSV 필드 수와 다른지 설명합니다.
- `csv.NewReader`와 `Read`를 사용해 두 필드를 얻습니다.
- Writer를 Flush한 뒤 Error를 확인합니다.
- 결과 CSV를 Reader로 다시 읽어 원래 두 필드가 보존되는지 검사합니다.
- 실패를 반환하는 Writer에서도 오류가 전달되는지 검사합니다.

### 수정 후 관찰할 값

```text
읽은 필드: 2
첫 필드: Go, basics
"Go, basics",25
```

### 주의

CSV 안의 큰따옴표 규칙과 Go 소스의 문자열 표기 규칙을 구분합니다.
인용이 이미 처리된 문자열 필드에 수동으로 큰따옴표를 더하지 않습니다.
Writer에게는 원래 필드 문자열을 전달합니다.

### 힌트

`csv.Writer.Flush`는 반환값이 없습니다.
`if err := writer.Flush(); ...`는 이 타입에서 성립하지 않습니다.
[Writer.Error](https://pkg.go.dev/encoding/csv#Writer.Error)를 함께 확인합니다.

## 문제 4. CSV 학습 시간 합계

**수준: 변형**

개인 파일: `_practice/15-data-time/csv-total/main.go`

함수 계약:

```go
func sumCSV(source io.Reader) (int, error)
```

### 입력 규칙

- 첫 레코드는 정확히 `name,minutes`입니다.
- 각 레코드의 필드는 정확히 두 개입니다.
- 이름은 앞뒤 공백을 제거한 뒤 비어 있지 않아야 합니다.
- 분은 0~120 정수입니다. 이 문제에서는 0분을 휴식 기록으로 허용합니다.
- 데이터 레코드는 최대 1000개입니다.
- 헤더만 있는 입력은 합계 0으로 성공합니다.
- 입력이 완전히 비어 있으면 필수 헤더 오류입니다.

본문의 `readSessions`는 1~120분입니다.
이 문제는 **0을 허용한다는 차이**를 요구사항에서 확인합니다.
기존 조건문을 복사한 뒤 그대로 두지 않습니다.

### 정상 입력

```csv
name,minutes
"Go, basics",25
테스트,35
휴식,0
```

예상 결과:

```text
합계: 60분
```

### 오류 정책과 경계

| 상황 | 반환값 |
| --- | --- |
| 정상 레코드 1000개, 각각 120분 | 120000, nil |
| 1001번째 데이터 레코드 | 0, 오류 |
| 이름이 공백 | 0, 오류 |
| 분이 -1 또는 121 | 0, 오류 |
| 분이 `x` | 0, 오류 |
| 세 번째 필드 존재 | 0, 오류 |
| 앞 레코드는 정상, 마지막 레코드 잘못됨 | 0, 오류 |
| CRLF 또는 마지막 개행 없음 | CSV가 유효하면 성공 |

최대 합계 120000은 이 문제의 입력 제한으로 정해집니다.
끝에서 오류가 났을 때 앞부분의 합계를 성공처럼 돌려주지 않습니다.
전체 바이트 크기 제한은 별도의 확장 과제입니다.

### 힌트

`FieldsPerRecord`와 헤더 검사는 역할이 다릅니다.
EOF는 데이터 루프에서 정상 종료로 처리합니다.
숫자 변환은 `strconv.Atoi`의 오류와 변환 후 범위를 모두 확인합니다.

[참고 구현](../../solutions/15-data-time/csv-total/main.go) ·
[테스트](../../solutions/15-data-time/csv-total/main_test.go)

## 문제 5. 시작 포함·끝 제외 시간 구간

**수준: 변형**

개인 파일: `_practice/15-data-time/time-window/main.go`

함수 계약:

```go
func withinWindow(value, start, end time.Time) bool
```

구간은 **`start <= value < end`**입니다.
시작을 포함하고 끝은 포함하지 않습니다.
`start >= end`인 비어 있거나 뒤집힌 구간은 모든 값을 거부합니다.

### 기준 시각

- start: 2026-11-09 09:00 UTC
- end: 2026-11-09 10:00 UTC
- 오차 비교 대신 정확한 `time.Time` 비교를 사용합니다.

| 검사 값 | 기대 |
| --- | --- |
| start보다 1나노초 전 | false |
| start | true |
| end보다 1나노초 전 | true |
| end | false |
| end보다 1시간 뒤 | false |
| start와 같은 순간을 UTC+9로 표시 | true |
| start == end인 구간 | false |
| start > end인 구간 | false |

### 완료 조건

- 문자열 비교로 구현하지 않습니다.
- `Time`의 `==`로 순간의 순서를 판단하지 않습니다.
- 시간대 표시가 달라도 같은 순간의 판정이 같습니다.
- 현재 시각을 사용하지 않아 실행마다 결과가 같습니다.

### 힌트

“value가 start보다 앞서지 않다”는 조건을 `!value.Before(start)`로 표현할 수 있습니다.
“end 이전”은 `value.Before(end)`입니다.
구간 자체가 유효한지도 먼저 확인합니다.

[참고 구현](../../solutions/15-data-time/time-window/main.go) ·
[테스트](../../solutions/15-data-time/time-window/main_test.go)

## 문제 6. CSV를 검증한 뒤 JSON 배열로 출력

**수준: 응용**

개인 파일: `_practice/15-data-time/csv-json/main.go`

함수 계약:

```go
func convertCSV(source io.Reader, destination io.Writer) error
```

입력 규칙은 문제 4와 같습니다.
출력 배열의 각 객체는 `name`과 `minutes` 필드를 가집니다.
이름의 앞뒤 공백은 제거합니다.

### 핵심 계약

**전체 CSV 검증이 끝난 뒤에만 결과를 씁니다.**
첫 레코드를 읽자마자 JSON을 쓰기 시작하는 스트리밍 구현과 구분합니다.
이 문제에서는 최대 1000개를 메모리에 모으는 선택을 합니다.

헤더만 있으면 `[]`와 마지막 개행을 출력합니다.
빈 입력은 헤더 오류이며 출력은 없어야 합니다.
nil 슬라이스를 그대로 인코딩해 `null`이 나오지 않도록 합니다.

입력:

```csv
name,minutes
Go,25
테스트,35
```

출력:

```json
[
  {
    "name": "Go",
    "minutes": 25
  },
  {
    "name": "테스트",
    "minutes": 35
  }
]
```

### 검사 순서

1. 정상 입력을 bytes.Buffer에 출력하고 JSON으로 다시 읽어 값이 같은지 확인합니다.
2. 인용된 쉼표가 있는 이름도 확인합니다.
3. 정상 첫 레코드 뒤에 잘못된 숫자 레코드를 넣고 출력 길이가 0인지 확인합니다.
4. 헤더만 있을 때 정확히 `"[]\n"`인지 확인합니다.
5. 실패하는 Writer를 전달하고 원인 오류가 보존되는지 확인합니다.

출력 실패까지 “어떤 바이트도 쓰지 않는다”고 보장하지 않습니다.
Writer는 일부 쓰기 뒤 실패할 수 있습니다.
여기서 보장하는 범위는 **잘못된 입력이 발견된 경우 출력하지 않음**입니다.

### 힌트

빈 슬라이스를 `make([]Session, 0)`로 시작합니다.
`json.NewEncoder`, `SetIndent("", "  ")`, `Encode`를 사용합니다.
Encode는 개행을 추가하므로 별도의 Println을 덧붙이지 않습니다.

[참고 구현](../../solutions/15-data-time/csv-json/main.go) ·
[테스트](../../solutions/15-data-time/csv-json/main_test.go)

## 선택 문제. 숫자 ID의 정밀도

**수준: 응용**

다음 JSON 숫자를 float64를 거치지 않고 int64로 읽습니다.

```text
9007199254740993
```

함수 계약은 `readID(source io.Reader) (int64, error)`입니다.
입력은 양수 정수 하나이고 뒤쪽 공백만 허용합니다.
최대 크기가 작은 입력이라는 전제이며 크기 제한은 별도로 추가할 수 있습니다.

- `Decoder.UseNumber`를 사용합니다.
- 결과가 `json.Number`인지 확인합니다.
- `Int64` 변환 오류를 확인합니다.
- 0과 음수는 거부합니다.
- 문자열 `"123"`, 소수 `1.5`, 두 값 `1 2`, int64 범위 초과를 거부합니다.
- 정상 값의 출력은 처음 입력한 정수와 같아야 합니다.

`UseNumber`만 켜고 형 변환 오류를 무시하면 완료되지 않습니다.
숫자 문법과 프로그램이 원하는 양수 정수 계약을 둘 다 검사합니다.
[공식 UseNumber 문서](https://pkg.go.dev/encoding/json#Decoder.UseNumber)를 참고합니다.

## 해설을 보기 전 완료 점검

- 입력 형식과 의미 검증을 다른 단계로 설명할 수 있다.
- JSON 하나 계약을 두 번째 읽기로 확인했다.
- CSV를 직접 Split하지 않았다.
- CSV Flush 뒤 오류를 확인했다.
- 0개와 최대·최대 초과 입력을 검사했다.
- 시간대가 다른 같은 순간을 비교했다.
- 잘못된 입력과 출력 장치 실패를 구분했다.

[해설로 이동](../../solutions/15-data-time/README.md) ·
[본문으로 돌아가기](../../book/15-data-time/README.md)

확인일: 2026-11-09. 기본 코드는 Go 1.27.3 기준입니다.
