# 15장 해설 — JSON·CSV·시간

[연습문제](../../exercises/15-data-time/README.md)의 해설입니다.
구현을 비교하기 전에 입력과 오류 계약이 같은지 확인합니다.
함수 이름을 따라 쓰는 것보다 경계에서 관찰되는 결과가 중요합니다.

모든 명령은 저장소 루트에서 실행합니다.
참고 구현은 독립 패키지이며 서로의 코드를 import하지 않습니다.
실제 프로그램은 외부 파일이나 네트워크 없이 종료됩니다.

| 문제 | 참고 구현 | 중심 검증 |
| --- | --- | --- |
| 2 | [decode-config](decode-config/main.go) | 크기·JSON·의미 검증 |
| 4 | [csv-total](csv-total/main.go) | 헤더·열·숫자·합계 |
| 5 | [time-window](time-window/main.go) | 시작 포함·끝 제외 |
| 6 | [csv-json](csv-json/main.go) | 검증 후 출력, I/O 오류 |

## 문제 1. 상태를 보존하는 타입

새 제로 값에 읽는다는 조건에서 결과는 다음과 같습니다.

| 입력 | Plain | Optional | Plain 재인코딩 | Optional 재인코딩 |
| --- | --- | --- | --- | --- |
| `{}` | false | nil | `{}` | `{}` |
| `{"enabled":false}` | false | false를 가리킴 | `{}` | `{"enabled":false}` |
| `{"enabled":true}` | true | true를 가리킴 | `{"enabled":true}` | `{"enabled":true}` |
| `{"enabled":null}` | false | nil | `{}` | `{}` |

v1에서 `omitempty`는 bool의 false를 생략합니다.
포인터가 nil이 아니면 가리키는 값이 false여도 포인터 자체는 빈 값이 아닙니다.
그래서 Optional은 명시적인 false를 출력에서 보존합니다.

하지만 누락과 null은 둘 다 nil입니다.
“미지정”과 “삭제 요청”을 구분하는 수정 API에는 상태가 더 필요합니다.
이 예제를 모든 부분 수정 형식의 완성된 해결책으로 사용하지 않습니다.

nil 포인터를 역참조하면 패닉이 발생합니다.
본문의 `describe`는 nil 분기에서 먼저 반환하여 그 뒤의 역참조를 안전하게 합니다.
문자열 `"false"`는 JSON 불리언이 아니므로 bool 포인터로 디코딩할 수 없습니다.

v2는 기본 매칭과 생략 규칙이 다릅니다.
[공식 비교](https://pkg.go.dev/encoding/json#hdr-Migrating_to_v2)를 확인한 뒤
기존 테스트를 사용해 바꾸려는 API의 계약을 다시 검증합니다.

## 문제 2. 크기 제한을 먼저 확인하기

소스: [decode-config/main.go](decode-config/main.go)

```go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Config struct {
	Course       string `json:"course"`
	DailyMinutes int    `json:"daily_minutes"`
}

func readConfig(source io.Reader) (Config, error) {
	const maximum = 4096
	data, err := io.ReadAll(io.LimitReader(source, maximum+1))
	if err != nil {
		return Config{}, fmt.Errorf("설정 읽기: %w", err)
	}
	if len(data) > maximum {
		return Config{}, errors.New("설정은 4096바이트까지 허용합니다")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("설정 JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return Config{}, fmt.Errorf("설정 끝 확인: %w", err)
		}
		return Config{}, errors.New("설정 JSON은 하나만 허용합니다")
	}
	config.Course = strings.TrimSpace(config.Course)
	if config.Course == "" || config.DailyMinutes < 1 || config.DailyMinutes > 120 {
		return Config{}, errors.New("과정과 1~120분이 필요합니다")
	}
	return config, nil
}

func run() error {
	config, err := readConfig(strings.NewReader(`{"course":"Go","daily_minutes":25}`))
	if err != nil {
		return err
	}
	fmt.Printf("%s: 하루 %d분\n", config.Course, config.DailyMinutes)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./solutions/15-data-time/decode-config
```

예상 결과:

```text
Go: 하루 25분
```

### 읽기 단계와 디코딩 단계

4097바이트까지 읽어서 초과 여부를 확인합니다.
정상 입력의 최대 4096바이트만 허용한 뒤 메모리 Reader에서 JSON을 읽습니다.
디코더가 첫 값 뒤의 큰 공백을 아직 읽지 않았다는 이유로 제한 검사가 빠지지 않습니다.

이 구현의 제한은 바이트 수입니다.
Reader가 기다리는 시간, 입력의 구조 깊이, 다른 프로그램의 자원 사용량까지 제한하지 않습니다.
이 기능의 경계를 정확히 이름 붙이는 것이 테스트 설계에도 도움이 됩니다.

| 단계 | 확인할 실패 |
| --- | --- |
| 읽기 | Reader 오류 |
| 크기 | 4096 초과 |
| 첫 Decode | 빈 입력, 문법, 타입, 모르는 필드 |
| 두 번째 Decode | 두 값, 뒤쪽 잘못된 문자 |
| 의미 | 빈 과정, 1~120 범위 밖 |
| 반환 | 검증된 Config 또는 제로 Config와 오류 |

### 왜 부분 Config를 반환하지 않는가?

디코딩 오류 전에 일부 필드가 채워졌을 수 있습니다.
호출자가 그 값을 사용하면 잘못된 설정이 실행에 섞입니다.
이 구현은 모든 실패에서 `Config{}`를 반환해 계약을 단순하게 만듭니다.

`Course`의 공백 제거는 의도적인 정규화입니다.
정확한 공백이 중요한 다른 도메인에서는 그대로 보존할 수 있습니다.
분의 범위는 표준 라이브러리 규칙이 아니라 문제에서 정한 정책입니다.

### 테스트의 목적

[main_test.go](decode-config/main_test.go)는
4096/4097바이트를 실제 문자열 길이로 만들고 검사합니다.
모르는 필드와 두 번째 값은 다른 입력으로 나누어 검사합니다.

정상 입력 뒤에 공백을 붙이는 사례는 첫 JSON 값 자체가 같아도
전체 입력 제한이 적용되는지 확인합니다.
실패하는 Reader는 파일 시스템과 무관하게 원인 전달을 재현합니다.

## 문제 3. CSV 문법과 Flush 복구

잘못된 프로그램의 실제 출력은 다음과 같습니다.

```text
직접 나눈 필드: 3
Flush 전 바이트: 0
```

Split은 인용 규칙을 모르므로 이름 안의 쉼표도 경계로 봅니다.
Writer는 작은 레코드를 버퍼에 보관할 수 있으므로
Flush 전에 아래쪽 bytes.Buffer가 비어 있을 수 있습니다.

수정한 독립 전체 프로그램은 다음과 같습니다.

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
	reader := csv.NewReader(strings.NewReader(input))
	reader.FieldsPerRecord = 2
	record, err := reader.Read()
	if err != nil {
		return err
	}
	fmt.Println("읽은 필드:", len(record))
	fmt.Println("첫 필드:", record[0])
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write(record); err != nil {
		return err
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}
	fmt.Print(output.String())
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

예상 결과:

```text
읽은 필드: 2
첫 필드: Go, basics
"Go, basics",25
```

이 코드는 고정된 레코드 하나를 확인하는 예제입니다.
일반 CSV 전체 파일의 검증은 문제 4처럼 EOF까지 반복해야 합니다.
한 번 Read했다는 사실만으로 뒤쪽 레코드를 모두 검사했다고 보지 않습니다.

쓰기 실패 검증은 본문의 [csv-write 테스트](../../examples/15-data-time/csv-write/main_test.go)에 있습니다.
실패하는 Writer를 넣어 작은 출력이 Flush 때 실패하는 경우를 확인합니다.
[Writer 문서](https://pkg.go.dev/encoding/csv#Writer)가 설명하는 Error 확인 순서를 따릅니다.

## 문제 4. 합계를 성공으로 확정하는 시점

소스: [csv-total/main.go](csv-total/main.go)

```go
package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func sumCSV(source io.Reader) (int, error) {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = 2
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("헤더 읽기: %w", err)
	}
	if header[0] != "name" || header[1] != "minutes" {
		return 0, errors.New("헤더는 name,minutes여야 합니다")
	}
	total := 0
	count := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return total, nil
		}
		if err != nil {
			return 0, fmt.Errorf("레코드 읽기: %w", err)
		}
		count++
		if count > 1000 {
			return 0, errors.New("레코드는 1000개까지 허용합니다")
		}
		minutes, err := strconv.Atoi(record[1])
		if err != nil {
			return 0, fmt.Errorf("분 변환: %w", err)
		}
		if strings.TrimSpace(record[0]) == "" || minutes < 0 || minutes > 120 {
			return 0, errors.New("이름과 0~120분이 필요합니다")
		}
		total += minutes
	}
}

func run() error {
	input := "name,minutes\n\"Go, basics\",25\n테스트,35\n휴식,0\n"
	total, err := sumCSV(strings.NewReader(input))
	if err != nil {
		return err
	}
	fmt.Printf("합계: %d분\n", total)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./solutions/15-data-time/csv-total
```

예상 결과:

```text
합계: 60분
```

### 헤더와 0개 레코드

첫 Read는 필수 헤더를 읽습니다.
입력이 비어 있으면 이 단계에서 EOF가 나므로 오류입니다.
헤더 뒤의 데이터 루프에서 EOF를 만나면 정상 종료입니다.

합계의 초기값은 0입니다.
헤더만 있는 입력은 덧셈을 한 번도 수행하지 않고 0을 반환합니다.
이는 0이 오류 코드가 아니라 정상적인 합계이기도 하다는 뜻입니다.

### 0분을 허용하는 이유

본문 예제와 달리 문제 4는 휴식 기록을 포함합니다.
따라서 조건은 `minutes < 0`입니다.
`minutes < 1`을 그대로 복사하면 정상 0분을 거부하는 버그가 됩니다.

문제를 다시 정의하지 않고 코드만 재사용하면 이런 차이가 숨기 쉽습니다.
테스트에 0을 따로 넣어 요구사항의 차이를 보존했습니다.

### 합계의 상한과 실패

| 데이터 | 결과 |
| --- | --- |
| 1000개 × 120분 | 120000 |
| 1001개 × 1분 | 오류 |
| 정상 25분 뒤 `x` | 0과 오류 |
| 정상 25분 뒤 -1분 | 0과 오류 |
| 이름만 있고 분 필드 없음 | CSV 필드 수 오류 |

레코드 수와 각 분의 상한 때문에 정상 합계가 120000을 넘지 않습니다.
이 값은 Go의 32비트 int에서도 표현됩니다.
입력 제한을 없애거나 다른 단위를 사용하면 산술 범위를 다시 검토해야 합니다.

[테스트](csv-total/main_test.go)는 CRLF, 마지막 개행 없음,
인용된 쉼표, Reader 오류도 검사합니다.
표시 줄 수를 세어 레코드 수로 사용하지 않습니다.

## 문제 5. 반열린 시간 구간

소스: [time-window/main.go](time-window/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func withinWindow(value, start, end time.Time) bool {
	return start.Before(end) && !value.Before(start) && value.Before(end)
}

func main() {
	start := time.Date(2026, time.November, 9, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	values := []time.Time{
		start.Add(-time.Nanosecond),
		start,
		end.Add(-time.Nanosecond),
		end,
	}
	for _, value := range values {
		fmt.Println(value.Format(time.RFC3339Nano), withinWindow(value, start, end))
	}
}
```

실행:

```sh
go run ./solutions/15-data-time/time-window
```

예상 결과:

```text
2026-11-09T08:59:59.999999999Z false
2026-11-09T09:00:00Z true
2026-11-09T09:59:59.999999999Z true
2026-11-09T10:00:00Z false
```

### 세 조건 읽기

`start.Before(end)`는 구간이 양의 길이인지 확인합니다.
`!value.Before(start)`는 value가 시작보다 이르지 않음을 확인합니다.
`value.Before(end)`는 끝보다 이르다는 조건으로 끝을 제외합니다.

이 구간 표기는 **반열린 구간**이라고 부릅니다.
서로 이어진 09:00~10:00과 10:00~11:00에 같은 순간이 중복 포함되지 않습니다.
모든 업무에서 반열린 구간을 써야 하는 것은 아니지만 정책을 명확히 표현할 수 있습니다.

### 시간대와 표시

테스트는 start를 고정 UTC+9 Location으로 표시한 값도 전달합니다.
시·분의 표시가 달라도 같은 순간이면 true입니다.
문자열이나 Location 이름을 직접 비교했다면 이 검증을 통과하지 못할 수 있습니다.

0길이와 뒤집힌 구간을 false로 정한 것은 문제의 계약입니다.
오류로 반환하는 API를 설계할 수도 있지만 이 함수는 bool을 반환하도록 요구했습니다.
잘못된 구간을 자동으로 뒤집는 기능은 넣지 않았습니다.

[테스트](time-window/main_test.go) ·
[Time.Before](https://pkg.go.dev/time#Time.Before)

## 문제 6. 입력을 모두 검사한 뒤 출력하기

소스: [csv-json/main.go](csv-json/main.go)

```go
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Session struct {
	Name    string `json:"name"`
	Minutes int    `json:"minutes"`
}

func convertCSV(source io.Reader, destination io.Writer) error {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = 2
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("헤더 읽기: %w", err)
	}
	if header[0] != "name" || header[1] != "minutes" {
		return errors.New("헤더는 name,minutes여야 합니다")
	}
	sessions := make([]Session, 0)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("레코드 읽기: %w", err)
		}
		if len(sessions) == 1000 {
			return errors.New("레코드는 1000개까지 허용합니다")
		}
		minutes, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("분 변환: %w", err)
		}
		name := strings.TrimSpace(record[0])
		if name == "" || minutes < 0 || minutes > 120 {
			return errors.New("이름과 0~120분이 필요합니다")
		}
		sessions = append(sessions, Session{Name: name, Minutes: minutes})
	}
	encoder := json.NewEncoder(destination)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sessions); err != nil {
		return fmt.Errorf("JSON 쓰기: %w", err)
	}
	return nil
}

func run() error {
	var output bytes.Buffer
	input := "name,minutes\nGo,25\n테스트,35\n"
	if err := convertCSV(strings.NewReader(input), &output); err != nil {
		return err
	}
	fmt.Print(output.String())
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

실행:

```sh
go run ./solutions/15-data-time/csv-json
```

예상 결과:

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

### 메모리에 모으는 이유

입력 첫 레코드를 확인한 즉시 출력하면 뒤 레코드에서 오류가 났을 때
이미 쓴 결과를 일반 Writer에서 되돌리기 어렵습니다.
이 문제는 1000개 제한 아래에서 모두 검증한 다음 한 JSON 배열을 씁니다.

입력 크기가 매우 크면 이 선택이 적합하지 않을 수 있습니다.
그때는 스트리밍과 부분 결과 정책, 임시 파일 후 교체 같은 별도 설계가 필요합니다.
여기서는 그런 기능을 구현했다고 가정하지 않습니다.

### 빈 슬라이스의 외부 의미

v1은 nil 슬라이스를 JSON null로 표현합니다.
이 문제는 결과가 “레코드 0개인 배열”이어야 하므로
`make([]Session, 0)`으로 빈 배열을 의도합니다.

| 입력 | Go에서의 의미 | 출력 |
| --- | --- | --- |
| 헤더 없음 | 입력 계약 오류 | 없음 |
| 헤더만 있음 | 데이터 0개 | `[]`와 개행 |
| 정상 두 레코드 | 데이터 2개 | 객체 2개 배열 |
| 정상 뒤 잘못된 레코드 | 전체 입력 오류 | 없음 |

### 출력 오류의 범위

`Encode`가 오류를 반환해도 Writer가 아무 바이트도 받지 않았다고 단정하지 않습니다.
이 구현이 보장하는 것은 **입력 검증 실패 전에 출력하지 않음**입니다.
출력 장치의 부분 쓰기를 되돌리는 트랜잭션은 제공하지 않습니다.

[테스트](csv-json/main_test.go)는 잘못된 마지막 레코드에서
bytes.Buffer 길이가 0인지 확인합니다.
다른 테스트는 실패하는 Writer의 원인 오류가 `errors.Is`로 보이는지 확인합니다.
두 테스트는 서로 다른 실패 구간을 검증합니다.

## 선택 문제. 큰 정수 ID 읽기

`UseNumber`는 인터페이스에 들어갈 숫자를 `json.Number`로 보존합니다.
그 뒤 int64 변환을 시도하므로 float64로 먼저 반올림되는 경로를 피합니다.
모든 크기의 정수에 성공하는 것은 아닙니다.

다음은 별도 개인 폴더에서 실행할 수 있는 전체 코드입니다.

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func readID(source io.Reader) (int64, error) {
	decoder := json.NewDecoder(source)
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return 0, err
	}
	number, ok := value.(json.Number)
	if !ok {
		return 0, errors.New("숫자가 필요합니다")
	}
	id, err := number.Int64()
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("양수 ID가 필요합니다")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return 0, err
		}
		return 0, errors.New("ID는 하나만 허용합니다")
	}
	return id, nil
}

func run() error {
	id, err := readID(strings.NewReader("9007199254740993"))
	if err != nil {
		return err
	}
	fmt.Println(id)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

예상 결과:

```text
9007199254740993
```

### 경계 확인

| 입력 | 결과 |
| --- | --- |
| `1` | 1 |
| `9007199254740993` | 정확한 정수 |
| `0`, `-1` | 양수 조건 오류 |
| `"123"` | 숫자 타입 아님 |
| `1.5` | 정수 변환 오류 |
| `9223372036854775808` | int64 범위 초과 |
| `1 2` | 두 번째 값 오류 |
| `null` | 숫자 타입 아님 |

크기 제한이 필요한 외부 입력이라면 문제 2의 제한 읽기를 앞에 추가합니다.
`UseNumber`는 크기 제한이나 악의적인 입력 전체를 다루는 보안 정책이 아닙니다.
[Number.Int64](https://pkg.go.dev/encoding/json#Number.Int64)의 변환 계약을 확인합니다.

## 개인 구현 검증

함수 이름과 타입이 문제의 계약과 같아진 뒤 테스트를 참고할 수 있습니다.
원래 테스트를 바꾸어 틀린 구현에 맞추지 않습니다.
필요하면 해당 `main_test.go`를 같은 개인 폴더에 복사해 실행합니다.

```sh
go test ./solutions/15-data-time/...
go vet ./solutions/15-data-time/...
go build ./solutions/15-data-time/...
```

예제까지 함께 검사하려면 다음 명령을 사용합니다.

```sh
go test ./examples/15-data-time/... ./solutions/15-data-time/...
```

이 장의 검사는 외부 파일·네트워크·현재 시각에 의존하지 않습니다.
JSON과 CSV는 실제 바이트 출력 및 다시 읽은 값을 확인합니다.
시간은 고정된 순간과 1나노초 경계를 사용합니다.

[본문](../../book/15-data-time/README.md) ·
[연습문제](../../exercises/15-data-time/README.md)

확인일: 2026-11-09. 코드·출력·오류·경계는 Go 1.27.3에서 확인했습니다.
