# 15. JSON·CSV·시간 다루기

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 응용 | 구조체·슬라이스·포인터, 오류, Reader와 Writer | Go, 터미널, 이 저장소 |

파일에서 읽은 바이트가 바로 프로그램의 데이터가 되지는 않습니다.
`{"minutes":25}`라는 바이트를 Go의 정수 필드로 옮기고,
`2026-11-09T09:30:00+09:00`을 비교 가능한 시각으로 해석해야 합니다.

이번 장은 **형식 해석과 값 검증을 분리하는 일**을 다룹니다.
JSON으로 읽을 수 있다는 사실은 학습 시간이 허용 범위라는 뜻이 아닙니다.
CSV의 두 필드가 읽혔다는 사실은 두 번째 필드가 정수라는 뜻이 아닙니다.
시간 문자열이 읽혔다는 사실은 그 시각이 예약 가능한 구간에 있다는 뜻이 아닙니다.

표준 라이브러리에 문법 처리를 맡기고,
프로그램이 결정할 규칙은 함수와 테스트로 드러내겠습니다.

## 학습 목표

- 구조체와 JSON 사이를 변환하고 필드 이름·누락·`null`을 구분한다.
- 한 JSON 값만 받는 입력에서 뒤에 남은 값과 알 수 없는 필드를 검사한다.
- CSV의 인용 규칙을 직접 구현하지 않고 읽기·쓰기 오류를 전달한다.
- 시간의 표시 형식, 시간대, 순간, 기간을 구분한다.
- 크기와 값의 경계를 검증한 뒤 JSON 또는 CSV 결과를 출력한다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
이 장의 예제는 메모리 안의 문자열과 버퍼를 사용합니다.
외부 계정, 네트워크, 현재 시각, 운영체제의 지역 설정이 필요하지 않습니다.

14장에서 배운 `io.Reader`, `io.Writer`, `io.EOF`, `bytes.Buffer`를 사용합니다.
구조체 태그는 이번 장에서 의미를 설명합니다.
포인터가 `nil`일 수 있다는 사실은 08장의 내용을 사용합니다.

JSON 예제의 import 경로는 **`encoding/json`**입니다.
이 패키지는 기존 동작을 유지하는 v1 API입니다.
Go의 공식 문서는 새 코드에서 `encoding/json/v2`도 검토하도록 안내합니다.
두 패키지는 이름만 다른 동등한 교체물이 아닙니다.

이 장은 기존 Go 코드에서 자주 만나는 v1의 계약을 먼저 익힙니다.
대소문자 매칭, 중복 이름, `omitempty`의 의미를 v2에 그대로 적용하지 않습니다.
차이는 [공식 이행 안내](https://pkg.go.dev/encoding/json#hdr-Migrating_to_v2)를 참고합니다.

| 예제 | 학습 대상 | 명령 |
| --- | --- | --- |
| [json-encode](../../examples/15-data-time/json-encode/main.go) | 구조체를 JSON으로 | `go run ./examples/15-data-time/json-encode` |
| [json-decode](../../examples/15-data-time/json-decode/main.go) | JSON과 입력 검증 | `go run ./examples/15-data-time/json-decode` |
| [json-presence](../../examples/15-data-time/json-presence/main.go) | 누락과 명시적 false | `go run ./examples/15-data-time/json-presence` |
| [csv-read](../../examples/15-data-time/csv-read/main.go) | CSV 레코드 읽기 | `go run ./examples/15-data-time/csv-read` |
| [csv-write](../../examples/15-data-time/csv-write/main.go) | CSV 인용과 Flush | `go run ./examples/15-data-time/csv-write` |
| [time-parse](../../examples/15-data-time/time-parse/main.go) | 파싱과 시간대 | `go run ./examples/15-data-time/time-parse` |
| [time-arithmetic](../../examples/15-data-time/time-arithmetic/main.go) | 기간과 시각 계산 | `go run ./examples/15-data-time/time-arithmetic` |

서로 다른 전체 프로그램을 같은 폴더에 이어 붙이지 않습니다.
본문의 발췌는 지정된 함수 안에서만 교체합니다.
공식 예제는 수정하지 않고 `_practice/15-data-time`에 개인 파일을 만듭니다.

## 1. 데이터 형식과 프로그램의 규칙

### 1.1 세 단계로 나누기

바이트를 읽는 일, 문법을 해석하는 일, 값을 허용하는 일은 다릅니다.
각 단계가 실패하면 다음 단계로 넘어가지 않습니다.

| 단계 | 질문 | 실패 사례 |
| --- | --- | --- |
| 입력 | 끝까지 읽을 수 있는가? | Reader 오류, 크기 초과 |
| 형식 | 선택한 문법과 Go 타입에 맞는가? | 닫히지 않은 JSON, 잘못된 CSV 인용 |
| 의미 | 프로그램이 받아도 되는 값인가? | 빈 제목, 음수 학습 시간 |
| 출력 | 결과를 대상에 기록했는가? | Writer 오류, Flush 오류 |

JSON 숫자 `-1`은 올바른 숫자 문법입니다.
하지만 학습 시간이 1분 이상이라는 규칙에는 맞지 않습니다.
파서가 이런 프로그램별 규칙을 대신 결정할 수는 없습니다.

검증 실패 뒤에는 제로 값과 오류를 돌려주는 계약을 사용합니다.
호출자는 오류가 `nil`인 경우에만 반환값을 사용합니다.
오류가 생기기 전까지 해석한 일부 데이터는 성공 결과로 공개하지 않습니다.

### 1.2 형식을 선택하는 기준

JSON은 필드 이름과 중첩 구조를 표현하기 좋습니다.
CSV는 같은 열 구조가 반복되는 표 형태에 적합합니다.
CSV에는 정수 타입이라는 선언이 없으므로 각 필드가 문자열로 읽힙니다.

| 요구 | 이 장의 선택 | 별도로 정할 규칙 |
| --- | --- | --- |
| 이름 있는 설정 값 | JSON 객체 | 필수 필드, 값 범위 |
| 같은 열의 학습 기록 | CSV | 헤더, 열 수, 숫자 변환 |
| 다른 시간대의 시각 교환 | RFC3339 문자열 | 허용 구간, 표시 시간대 |
| 작업의 길이 | `time.Duration` | 최대 길이, 음수 허용 여부 |

형식 선택만으로 크기 제한이나 오류 정책이 정해지지는 않습니다.
짧은 예제 입력은 코드 안에 고정하고,
외부 입력을 받는 개인 실습에서는 읽는 바이트 수부터 제한합니다.

## 2. 구조체를 JSON으로 바꾸기

### 2.1 공개 필드와 태그

JSON의 객체는 문자열 이름과 값의 쌍으로 구성됩니다.
Go의 구조체 필드에는 Go 코드에서 쓸 이름과 외부 형식에서 쓸 이름을 따로 둘 수 있습니다.

`Name`은 대문자로 시작하는 공개 필드입니다.
뒤의 `` `json:"name"` ``은 **구조체 태그(struct tag)**입니다.
태그는 런타임 라이브러리가 읽을 수 있는 필드의 부가 정보입니다.

`json.Marshal`은 이 태그를 읽어 JSON 이름을 `name`으로 정합니다.
태그가 필드를 공개 상태로 바꾸지는 않습니다.
소문자로 시작하는 비공개 필드를 JSON에 내보내려는 용도로 쓰지 않습니다.

[Marshal 문서](https://pkg.go.dev/encoding/json#Marshal)는
Go 값이 JSON 값으로 바뀌는 규칙과 지원하지 않는 타입을 설명합니다.

소스: [json-encode/main.go](../../examples/15-data-time/json-encode/main.go)

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Study struct {
	Name    string `json:"name"`
	Minutes int    `json:"minutes"`
	Done    bool   `json:"done"`
}

func run() error {
	study := Study{Name: "Go", Minutes: 25, Done: false}
	data, err := json.Marshal(study)
	if err != nil {
		return fmt.Errorf("JSON 인코딩: %w", err)
	}
	fmt.Println(string(data))
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
go run ./examples/15-data-time/json-encode
```

예상 결과:

```text
{"name":"Go","minutes":25,"done":false}
```

### 2.2 반환값을 따라가기

| 표현 | 타입 | 역할 |
| --- | --- | --- |
| `study` | `Study` | 프로그램 안의 값 |
| `json.Marshal(study)` | `[]byte, error` | JSON 바이트와 변환 오류 |
| `string(data)` | `string` | 이번 예제의 화면 출력용 변환 |
| `fmt.Println(...)` | 출력 동작 | 마지막 개행 추가 |

`Marshal`이 돌려주는 바이트에는 마지막 개행이 붙지 않습니다.
이 예제의 마지막 개행은 `Println`이 추가합니다.
뒤에서 사용할 `Encoder.Encode`는 JSON 값 뒤에 개행을 씁니다.

`false`가 출력되는 이유는 `Done`에 생략 옵션이 없기 때문입니다.
제로 값이라고 해서 모든 필드가 자동으로 빠지지는 않습니다.

지원하지 않는 값이나 사용자 정의 변환 메서드가 있으면 인코딩이 실패할 수 있습니다.
이번 구조체가 작고 고정되어 있어도 오류 반환값을 확인하는 형태를 유지합니다.
이것은 모든 JSON 인코딩이 실패한다는 뜻이 아니라 함수의 계약을 드러내는 방식입니다.

### 2.3 `omitempty`를 붙이기 전에 의미 정하기

v1의 `omitempty`는 `false`, 숫자 0, nil 포인터, 빈 문자열 같은
빈 Go 값을 JSON에서 생략하도록 합니다.
구체적인 규칙은 [Marshal의 필드 옵션](https://pkg.go.dev/encoding/json#Marshal)에 있습니다.

`Done bool`에 `omitempty`를 붙이면 `false`일 때 필드가 빠집니다.
“아직 지정하지 않음”과 “사용자가 false라고 지정함”이 다른 뜻이라면
bool 하나로 두 상태를 표현할 수 없습니다.

생략 옵션은 출력 크기를 줄이는 장식이 아닙니다.
필드가 없을 때 받는 쪽이 어떤 의미로 처리할지 함께 정해야 합니다.
4절에서는 포인터를 사용해 명시적인 false를 보존합니다.

## 3. JSON을 읽고 한 값인지 확인하기

### 3.1 `Decode` 성공 뒤에도 할 일이 남는다

`json.NewDecoder`는 Reader에서 JSON 값을 읽는 디코더를 만듭니다.
`Decode` 한 번은 다음 JSON 값 하나를 읽습니다.
따라서 첫 호출이 성공해도 뒤에 또 다른 JSON 값이 있을 수 있습니다.

설정 파일 하나에 객체 하나를 받는 계약이라면
두 번째 `Decode`가 `io.EOF`를 반환하는지도 확인합니다.
뒤쪽 공백은 허용하되 두 번째 값과 잘못된 문자는 거부합니다.

`DisallowUnknownFields`는 구조체에 없는 필드를 발견했을 때 오류를 냅니다.
이 옵션과 두 번째 읽기는 서로 다른 조건을 확인합니다.
[Decoder](https://pkg.go.dev/encoding/json#Decoder)와
[DisallowUnknownFields](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields)를 참고합니다.

소스: [json-decode/main.go](../../examples/15-data-time/json-decode/main.go)

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

type Task struct {
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

func decodeTask(source io.Reader) (Task, error) {
	decoder := json.NewDecoder(source)
	decoder.DisallowUnknownFields()
	var task Task
	if err := decoder.Decode(&task); err != nil {
		return Task{}, fmt.Errorf("작업 JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return Task{}, fmt.Errorf("JSON 끝 확인: %w", err)
		}
		return Task{}, errors.New("JSON 값은 하나만 허용합니다")
	}
	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" || task.Minutes < 1 || task.Minutes > 120 {
		return Task{}, errors.New("제목과 1~120분이 필요합니다")
	}
	return task, nil
}

func run() error {
	task, err := decodeTask(strings.NewReader(`{"title":"Go","minutes":25}`))
	if err != nil {
		return err
	}
	fmt.Printf("%s: %d분\n", task.Title, task.Minutes)
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
go run ./examples/15-data-time/json-decode
```

예상 결과:

```text
Go: 25분
```

### 3.2 포인터를 전달하는 이유

`Decode(&task)`는 디코더가 `task`의 필드 값을 바꾸도록 주소를 전달합니다.
`Decode(task)`처럼 값 자체를 넘기면 결과를 저장할 대상이 되지 못합니다.
JSON 디코딩의 포인터 요구는 [Unmarshal 문서](https://pkg.go.dev/encoding/json#Unmarshal)에도 설명되어 있습니다.

`var task Task`로 매번 새 제로 값을 만드는 점도 중요합니다.
기존 값을 재사용하면 JSON에 없는 필드가 이전 값으로 남는 동작을 만날 수 있습니다.
입력 한 건을 새 값에 해석하고 검증을 통과한 값만 돌려주면 판단이 명확합니다.

### 3.3 두 번째 읽기의 세 갈래

| 두 번째 읽기 결과 | 의미 | 처리 |
| --- | --- | --- |
| `io.EOF` | 첫 값 뒤에 공백만 있고 입력 종료 | 검증 계속 |
| `nil` | 다음 JSON 값도 존재 | 값 하나 계약 위반 |
| 다른 오류 | 뒤쪽 문법 오류 또는 입력 오류 | 오류 전달 |

두 번째 값을 저장할 `extra`는 `any`입니다.
어떤 JSON 타입이 오든 “값이 하나 더 있었다”는 사실만 확인할 목적입니다.
두 번째 값이 `{}`이든 숫자든 배열이든 거부합니다.

EOF는 이 위치에서 정상 종료입니다.
첫 번째 읽기의 EOF는 JSON 값이 아예 없다는 뜻이므로 오류입니다.
같은 오류 값도 어느 단계에서 만났는지에 따라 프로그램의 의미가 달라집니다.

### 3.4 문법 검사와 의미 검사 표

| 입력 | 형식 해석 | 최종 결과 |
| --- | --- | --- |
| `{"title":"Go","minutes":25}` | 성공 | `Go`, 25 |
| `{"title":" Go ","minutes":1}` | 성공 | 제목을 다듬어 `Go`, 1 |
| `{"title":"Go","minutes":120}` | 성공 | 상한 포함 |
| `{"title":"Go","minutes":0}` | 성공 | 범위 오류 |
| `{"title":" ","minutes":25}` | 성공 | 빈 제목 오류 |
| `{"title":"Go","minutes":1.5}` | 정수 필드에 맞지 않음 | 변환 오류 |
| `{"title":"Go","minutes":"25"}` | 문자열을 int에 저장할 수 없음 | 변환 오류 |
| `{"title":"Go","minutes":25,"memo":"x"}` | 모르는 필드 | 오류 |
| `{"title":"Go","minutes":25} {}` | 첫 값 성공, 두 번째 값 존재 | 오류 |
| `null` | 새 구조체는 제로 값 상태 | 의미 검사에서 오류 |

“필수”라는 규칙을 태그만으로 표현하지 않았습니다.
누락된 `Minutes`가 0이 되는 사실을 범위 검사에 활용합니다.
반대로 0 자체가 허용되는 도메인이라면 이 방법만으로 누락 여부를 알 수 없습니다.

### 3.5 이 검사가 보장하지 않는 것

v1은 구조체 이름을 대소문자 구분 없이 매칭할 수 있습니다.
JSON 객체에 중복된 이름도 허용합니다.
이 동작은 `DisallowUnknownFields`를 켜는 것만으로 바뀌지 않습니다.

예를 들어 이 구조체에서 `"TITLE"`은 `title`에 대응할 수 있고,
동일한 문자열 필드가 다시 나오면 뒤에서 읽은 값의 영향을 받습니다.
이 장의 함수는 중복 이름 거부나 정확한 키 철자 강제를 약속하지 않습니다.
자세한 차이는 [v1과 v2 비교](https://pkg.go.dev/encoding/json#hdr-Migrating_to_v2)를 확인합니다.

여기서는 짧은 고정 문자열을 사용했습니다.
디코더에 Reader를 연결했다는 사실만으로 전체 입력 크기가 제한되지는 않습니다.
개인 실습과 연습문제 2에서는 14장의 `maximum+1` 읽기를 앞에 붙입니다.

## 4. 누락·`null`·false를 구분하기

### 4.1 bool 포인터로 두 종류의 상태 표현하기

`bool`은 true와 false라는 두 값만 가집니다.
`*bool`은 nil이라는 상태를 추가로 표현할 수 있습니다.

JSON에서 false를 읽으면 false를 담은 값의 주소가 저장됩니다.
포인터 자체는 nil이 아니므로 `omitempty`여도 필드가 출력됩니다.
JSON에서 `null`을 읽으면 포인터가 nil이 됩니다.

이 예제에서는 매 입력마다 새 `Settings`를 만듭니다.
그 조건에서 필드 누락과 null은 둘 다 nil입니다.
**포인터 하나가 누락과 null까지 서로 구분해 주지는 않습니다.**

소스: [json-presence/main.go](../../examples/15-data-time/json-presence/main.go)

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Settings struct {
	Enabled *bool `json:"enabled,omitempty"`
}

func describe(settings Settings) string {
	if settings.Enabled == nil {
		return "미지정"
	}
	if *settings.Enabled {
		return "켜짐"
	}
	return "꺼짐"
}

func run() error {
	inputs := []string{
		`{}`,
		`{"enabled":false}`,
		`{"enabled":true}`,
		`{"enabled":null}`,
	}
	for _, input := range inputs {
		var settings Settings
		if err := json.Unmarshal([]byte(input), &settings); err != nil {
			return err
		}
		output, err := json.Marshal(settings)
		if err != nil {
			return err
		}
		fmt.Printf("%s => %s => %s\n", input, describe(settings), output)
	}
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
go run ./examples/15-data-time/json-presence
```

예상 결과:

```text
{} => 미지정 => {}
{"enabled":false} => 꺼짐 => {"enabled":false}
{"enabled":true} => 켜짐 => {"enabled":true}
{"enabled":null} => 미지정 => {}
```

### 4.2 상태표를 먼저 만든다

| 입력 상태 | `Enabled == nil` | 역참조 가능한가? | 다시 인코딩 |
| --- | --- | --- | --- |
| 필드 누락 | true | 아니오 | 필드 생략 |
| null | true | 아니오 | 필드 생략 |
| false | false | 가능, false | false 보존 |
| true | false | 가능, true | true 보존 |

`describe`는 nil을 확인한 다음에만 `*settings.Enabled`를 읽습니다.
nil 확인보다 역참조를 먼저 하면 패닉이 발생합니다.
포인터를 사용하는 이유가 곧 nil 분기의 이유입니다.

누락은 “기존 값 유지”, null은 “기존 값 삭제”인 수정 API도 있습니다.
그 계약에서는 포인터만으로 충분하지 않습니다.
별도의 존재 여부, `json.RawMessage`, 사용자 정의 디코딩 등을 설계해야 합니다.
이번 기본 예제는 그 계약을 구현했다고 주장하지 않습니다.

입력 `"false"`는 문자열이고 `false`는 불리언입니다.
사람에게 비슷하게 보여도 JSON 타입은 다릅니다.
문자열을 임의로 bool로 바꾸는 동작은 추가하지 않았습니다.

## 5. CSV를 레코드로 읽기

### 5.1 쉼표로 나누기만 해서는 안 되는 이유

CSV는 쉼표가 있는 텍스트지만 쉼표가 항상 필드 경계는 아닙니다.
`"Go, basics"` 안의 쉼표는 이름의 일부입니다.
인용된 필드 안에는 줄바꿈과 큰따옴표도 들어갈 수 있습니다.

그래서 `strings.Split(line, ",")`로 일반적인 CSV를 해석하지 않습니다.
`encoding/csv`가 인용과 줄바꿈 규칙을 처리하도록 합니다.
[패키지 개요](https://pkg.go.dev/encoding/csv)는 지원하는 CSV 규칙을 설명합니다.

**레코드(record)**는 한 묶음의 필드입니다.
인용된 필드에 줄바꿈이 있으면 레코드 하나가 물리적인 여러 줄을 차지할 수 있습니다.
오류 메시지에서 레코드 번호와 파일 줄 번호를 혼동하지 않습니다.

### 5.2 헤더와 필드 수를 계약으로 정하기

이 예제는 첫 레코드를 정확히 `name,minutes`로 요구합니다.
데이터 레코드는 두 필드이며 이름은 공백을 다듬은 뒤 비어 있지 않아야 합니다.
분은 1~120 정수이고 최대 1000개 레코드까지 받습니다.

`FieldsPerRecord = 2`는 레코드의 필드 개수를 확인합니다.
이름이 비어 있는지와 분이 정수인지는 직접 검사합니다.
[Reader](https://pkg.go.dev/encoding/csv#Reader)의 설정과 구분합니다.

소스: [csv-read/main.go](../../examples/15-data-time/csv-read/main.go)

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

type Session struct {
	Name    string
	Minutes int
}

func readSessions(source io.Reader) ([]Session, error) {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = 2
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("CSV 헤더: %w", err)
	}
	if header[0] != "name" || header[1] != "minutes" {
		return nil, errors.New("헤더는 name,minutes여야 합니다")
	}
	sessions := make([]Session, 0)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return sessions, nil
		}
		if err != nil {
			return nil, fmt.Errorf("CSV 레코드: %w", err)
		}
		if len(sessions) == 1000 {
			return nil, errors.New("레코드는 1000개까지 허용합니다")
		}
		name := strings.TrimSpace(record[0])
		minutes, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, fmt.Errorf("분을 정수로 변환: %w", err)
		}
		if name == "" || minutes < 1 || minutes > 120 {
			return nil, errors.New("이름과 1~120분이 필요합니다")
		}
		sessions = append(sessions, Session{Name: name, Minutes: minutes})
	}
}

func run() error {
	input := "name,minutes\n\"Go, basics\",25\n테스트,35\n"
	sessions, err := readSessions(strings.NewReader(input))
	if err != nil {
		return err
	}
	total := 0
	for _, session := range sessions {
		fmt.Printf("%s: %d분\n", session.Name, session.Minutes)
		total += session.Minutes
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
go run ./examples/15-data-time/csv-read
```

예상 결과:

```text
Go, basics: 25분
테스트: 35분
합계: 60분
```

### 5.3 한 레코드가 처리되는 순서

| 단계 | `"Go, basics",25`에서 얻는 값 |
| --- | --- |
| CSV 읽기 | `[]string{"Go, basics", "25"}` |
| 이름 다듬기 | `"Go, basics"` |
| `strconv.Atoi` | 정수 25 |
| 범위 검사 | 1~120에 포함 |
| 슬라이스 추가 | `Session{Name: "Go, basics", Minutes: 25}` |

`Atoi`는 숫자 문자열을 int로 변환합니다.
변환 실패나 int 범위 초과도 오류입니다.
[strconv.Atoi](https://pkg.go.dev/strconv#Atoi)를 참고합니다.

이 예제는 숫자 필드에 자동 `TrimSpace`를 적용하지 않습니다.
`Go, 25`의 두 번째 필드는 공백이 있는 `" 25"`이므로 변환 오류입니다.
허용할 공백 규칙은 프로그램이 명시적으로 선택해야 합니다.

### 5.4 EOF와 부분 결과

헤더를 읽기 전에 EOF를 만나면 필수 헤더가 없다는 뜻입니다.
헤더만 있고 그 다음에 EOF를 만나면 데이터가 0개인 정상 파일입니다.
빈 슬라이스를 반환하므로 호출자는 0번 순회하고 합계를 0으로 출력할 수 있습니다.

도중에 잘못된 레코드가 나오면 `nil, error`를 반환합니다.
앞의 정상 레코드를 성공 결과처럼 돌려주지 않습니다.
부분 성공이 필요한 프로그램이라면 별도의 결과 타입과 정책이 필요합니다.

`csv.Reader`는 기본적으로 빈 줄을 무시합니다.
공백이 있는 줄은 반드시 빈 줄로 취급되는 것이 아닙니다.
이 차이는 [Reader.Read](https://pkg.go.dev/encoding/csv#Reader.Read)와 패키지 개요에서 확인합니다.

레코드 1000개 제한은 슬라이스 크기와 합계의 상한을 설명하기 위한 선택입니다.
레코드 하나의 필드가 아주 길 수 있으므로 전체 바이트 제한과 같지는 않습니다.
외부 파일의 메모리 사용까지 제한하려면 입력 바이트 제한도 별도로 설계합니다.

## 6. CSV를 쓰고 오류 확인하기

### 6.1 인용 규칙은 Writer가 맡는다

CSV를 만들 때도 쉼표와 큰따옴표를 직접 붙이는 방식은 피합니다.
필드를 문자열 슬라이스로 전달하면 Writer가 필요한 인용을 적용합니다.

이름에 쉼표가 있으면 필드 전체를 큰따옴표로 감쌉니다.
필드 안의 큰따옴표는 CSV 규칙에 맞춰 이스케이프합니다.
이것은 Go 문자열 안에서 `\"`를 쓰는 규칙과 다른 층의 처리입니다.

소스: [csv-write/main.go](../../examples/15-data-time/csv-write/main.go)

```go
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Session struct {
	Name    string
	Minutes int
}

func writeSessions(destination io.Writer, sessions []Session) error {
	writer := csv.NewWriter(destination)
	if err := writer.Write([]string{"name", "minutes"}); err != nil {
		return err
	}
	for _, session := range sessions {
		record := []string{session.Name, strconv.Itoa(session.Minutes)}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func run() error {
	sessions := []Session{
		{Name: "Go, basics", Minutes: 25},
		{Name: `문자열 "연습"`, Minutes: 35},
	}
	var output bytes.Buffer
	if err := writeSessions(&output, sessions); err != nil {
		return fmt.Errorf("CSV 쓰기: %w", err)
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
go run ./examples/15-data-time/csv-write
```

예상 결과:

```text
name,minutes
"Go, basics",25
"문자열 ""연습""",35
```

### 6.2 `Flush` 다음에 `Error`

`csv.Writer`는 내부에 버퍼를 둡니다.
`Write`의 성공만 확인하고 끝내면 아직 대상 Writer에 전달하지 않은 데이터가 있을 수 있습니다.

`csv.Writer.Flush`는 오류를 반환하지 않습니다.
호출 뒤 `writer.Error()`를 읽어 쓰기·버퍼 비우기 과정의 오류를 확인합니다.
[Writer.Flush](https://pkg.go.dev/encoding/csv#Writer.Flush)와
[Writer.Error](https://pkg.go.dev/encoding/csv#Writer.Error)를 함께 읽습니다.

14장의 `bufio.Writer.Flush() error`와 호출 모양이 다릅니다.
같은 이름이라고 반환값까지 같다고 가정하지 않습니다.
컴파일러와 해당 타입의 문서로 확인합니다.

| API | 끝에서 해야 하는 일 |
| --- | --- |
| `bufio.Writer` | `Flush`의 오류 확인 |
| `csv.Writer` | `Flush` 후 `Error` 확인 |
| `json.Encoder` | 각 `Encode`의 오류 확인 |
| 직접 연 `os.File` | 위 처리와 별개로 `Close` 확인 |

`writeSessions`는 CSV 표현만 책임집니다.
이 함수 자체는 분이 1~120인지 검사하지 않습니다.
이미 검증된 데이터를 출력하는 함수와 외부 데이터를 검증하는 함수를 구분했습니다.

Reader나 Writer를 전달받았다고 소유권도 받은 것은 아닙니다.
이 함수는 호출자가 준 Writer를 닫지 않습니다.
파일을 열었다면 연 쪽이 닫는 책임을 가집니다.

## 7. 시간 문자열을 순간으로 해석하기

### 7.1 문자열, 순간, 표시 시간대

`2026-11-09T09:30:00+09:00`은 UTC보다 9시간 앞선 표시입니다.
같은 순간을 UTC로 표시하면 `2026-11-09T00:30:00Z`입니다.
숫자가 달라졌어도 시간 여행을 한 것이 아니라 표시 기준을 바꾼 것입니다.

`time.Time`은 시각을 다루는 값입니다.
`Format`은 그 값을 문자열로 만들고,
`Parse`는 정해진 형식의 문자열을 `time.Time`으로 해석합니다.

`UTC()` 또는 `In(location)`은 같은 순간의 표시 위치를 바꿉니다.
순간의 동일성을 비교할 때는 `Equal`을 사용합니다.
[Time](https://pkg.go.dev/time#Time)과 [Time.Equal](https://pkg.go.dev/time#Time.Equal)을 참고합니다.

### 7.2 Go의 레이아웃

Go는 `YYYY-MM-DD` 같은 자리 표시자를 기본 레이아웃으로 사용하지 않습니다.
기준 시각의 숫자 배치를 예시로 적습니다.

| 의미 | Go 레이아웃 조각 |
| --- | --- |
| 네 자리 연도 | `2006` |
| 두 자리 월 | `01` |
| 두 자리 일 | `02` |
| 24시간제 시 | `15` |
| 분 | `04` |
| 초 | `05` |
| 숫자 시간대 | `-07:00` |

`"2006-01-02 15:04"`는 2006년만 읽는 형식이 아닙니다.
연도·월·일·시·분이 어느 자리에 오는지를 나타냅니다.
[레이아웃 문서](https://pkg.go.dev/time#pkg-constants)에 기준 시각이 설명되어 있습니다.

통신이나 저장 형식에는 이미 정해진 상수를 우선 검토합니다.
이번 예제는 `time.RFC3339`를 사용해 숫자 오프셋까지 읽습니다.

소스: [time-parse/main.go](../../examples/15-data-time/time-parse/main.go)

```go
package main

import (
	"fmt"
	"os"
	"time"
)

func run() error {
	start, err := time.Parse(time.RFC3339, "2026-11-09T09:30:00+09:00")
	if err != nil {
		return err
	}
	fmt.Println("원래 표시:", start.Format(time.RFC3339))
	fmt.Println("UTC 표시:", start.UTC().Format(time.RFC3339))
	zone := time.FixedZone("UTC+9", 9*60*60)
	local, err := time.ParseInLocation("2006-01-02 15:04", "2026-11-09 09:30", zone)
	if err != nil {
		return err
	}
	fmt.Println("같은 순간:", start.Equal(local))
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
go run ./examples/15-data-time/time-parse
```

예상 결과:

```text
원래 표시: 2026-11-09T09:30:00+09:00
UTC 표시: 2026-11-09T00:30:00Z
같은 순간: true
```

### 7.3 시간대가 없는 문자열

`"2026-11-09 09:30"`에는 시간대가 없습니다.
이 문자열만으로 세계의 어느 순간인지 정할 수는 없습니다.
프로그램이 어떤 지역의 시각으로 해석할지 정해야 합니다.

`time.Parse`는 시간대 정보가 없으면 UTC로 해석합니다.
`time.ParseInLocation`은 전달한 Location에서 해석합니다.
[ParseInLocation](https://pkg.go.dev/time#ParseInLocation)에서 두 함수의 차이를 확인합니다.

예제의 `FixedZone("UTC+9", 9*60*60)`는 항상 +9시간인 고정 오프셋입니다.
이름은 설명용이고 두 번째 인자는 초 단위 오프셋입니다.
특정 국가의 과거·미래 시간대 규칙 전체를 구현한 것이 아닙니다.

실제 지역 규칙이 필요하면 `LoadLocation("Asia/Seoul")` 같은 이름을 사용합니다.
Location을 읽을 수 없는 환경의 오류도 처리해야 합니다.
선택 부분에서는 시간대 데이터가 필요한 예제를 따로 다룹니다.

### 7.4 비교와 검증

문자열의 사전순 비교를 모든 시각 비교에 쓰지 않습니다.
서로 다른 숫자 오프셋, 소수초 길이, 표시 형식 때문에 순서가 달라질 수 있습니다.
먼저 파싱하고 `Before`, `After`, `Equal`로 비교합니다.

`Time`의 `==`는 표시 위치나 단조 시계 정보까지 포함하는 구조적 비교입니다.
같은 순간이라는 의미의 비교에는 `Equal`이 적합합니다.
[Time의 비교 주의사항](https://pkg.go.dev/time#Time)을 참고합니다.

잘못된 날짜와 문자열은 `Parse`의 오류로 확인합니다.
반면 `time.Date`는 범위를 벗어난 월·일 등을 정규화할 수 있습니다.
예를 들어 “2월 30일을 사용자가 잘못 입력했는가”를 검사할 때
그 값을 `Date`에 넣고 오류를 기다리는 방식은 성립하지 않습니다.

## 8. 기간을 더하고 두 시각의 차이 구하기

### 8.1 `Duration`과 `Time`의 역할

`time.Time`은 순간을 나타냅니다.
`time.Duration`은 두 순간 사이의 길이를 나타내는 나노초 단위 정수 값입니다.
“오전 9시”와 “90분”을 같은 변수 의미로 다루지 않습니다.

`ParseDuration("1h30m")`은 단위가 붙은 문자열을 기간으로 읽습니다.
`Add`는 시각에 기간을 더하고 `Sub`는 두 시각의 차이를 구합니다.
[ParseDuration](https://pkg.go.dev/time#ParseDuration)과
[Duration](https://pkg.go.dev/time#Duration)을 참고합니다.

소스: [time-arithmetic/main.go](../../examples/15-data-time/time-arithmetic/main.go)

```go
package main

import (
	"fmt"
	"os"
	"time"
)

func run() error {
	duration, err := time.ParseDuration("1h30m")
	if err != nil {
		return err
	}
	start := time.Date(2026, time.November, 9, 9, 30, 0, 0, time.UTC)
	end := start.Add(duration)
	fmt.Printf("기간: %.0f분\n", duration.Minutes())
	fmt.Println("종료:", end.Format(time.RFC3339))
	fmt.Println("차이:", end.Sub(start))
	fmt.Println("시작이 먼저:", start.Before(end))
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
go run ./examples/15-data-time/time-arithmetic
```

예상 결과:

```text
기간: 90분
종료: 2026-11-09T11:00:00Z
차이: 1h30m0s
시작이 먼저: true
```

### 8.2 계산을 추적하기

| 단계 | 값 |
| --- | --- |
| 시작 | 2026-11-09 09:30 UTC |
| 기간 | 1시간 30분 |
| 종료 | 2026-11-09 11:00 UTC |
| 종료에서 시작 빼기 | 1h30m0s |
| 시작이 종료보다 앞선가? | true |

`duration.Minutes()`는 float64를 반환합니다.
출력의 `%.0f`는 이번 90분을 소수점 없이 보여 주려는 선택입니다.
기간을 내부에서 문자열로 바꾸어 더하는 계산은 하지 않습니다.

정수 분을 Duration으로 만들 때는 `time.Duration(minutes) * time.Minute`처럼 씁니다.
단위 없는 `time.Duration(25)`는 25분이 아니라 25나노초입니다.
또한 아주 큰 정수 곱셈은 범위를 벗어날 수 있으므로 입력 상한을 정합니다.

`ParseDuration`은 음수 기간도 읽을 수 있습니다.
예약 길이에 음수가 허용되는지는 따로 검사해야 합니다.
`"1d"`는 이 API가 지원하는 기간 단위가 아닙니다.
하루라는 달력 개념은 다음 절과 구분합니다.

### 8.3 선택: 달력의 하루와 24시간

어떤 지역은 날짜 경계 사이에서 UTC 오프셋이 바뀝니다.
그래서 다음 날 같은 지역 시각과 24시간 뒤가 서로 다를 수 있습니다.

다음 전체 코드는 별도 개인 폴더에서 실행합니다.
2024년의 고정된 날짜를 사용하고 현재 시각에 의존하지 않습니다.
`_ "time/tzdata"`는 식별자를 사용하지 않고 패키지 초기화를 실행하는 import입니다.
여기서는 시스템에 시간대 파일이 없을 때 사용할 데이터를 포함합니다.

```go
package main

import (
	"fmt"
	"os"
	"time"
	_ "time/tzdata"
)

func run() error {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}
	start := time.Date(2024, time.March, 9, 12, 0, 0, 0, zone)
	nextDate := start.AddDate(0, 0, 1)
	after24Hours := start.Add(24 * time.Hour)
	fmt.Println("다음 날짜:", nextDate.Format(time.RFC3339))
	fmt.Println("날짜 간격:", nextDate.Sub(start))
	fmt.Println("24시간 뒤:", after24Hours.Format(time.RFC3339))
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
다음 날짜: 2024-03-10T12:00:00-04:00
날짜 간격: 23h0m0s
24시간 뒤: 2024-03-10T13:00:00-04:00
```

`AddDate(0, 0, 1)`은 Location의 달력 규칙을 사용합니다.
`Add(24*time.Hour)`는 고정 길이를 더합니다.
[AddDate](https://pkg.go.dev/time#Time.AddDate)와
[time/tzdata](https://pkg.go.dev/time/tzdata)를 참고합니다.

월말에 한 달을 더하는 것도 결과를 직접 정의하고 검토해야 합니다.
`AddDate`는 존재하지 않는 날짜를 정규화합니다.
청구일을 항상 그 달의 마지막 날로 유지하는 업무 규칙까지 대신 구현하지 않습니다.

## 9. 개인 실습: 학습 예약 JSON 만들기

### 9.1 완료 목표와 계약

JSON으로 받은 제목·시작 시각·분을 검증한 뒤
UTC 시작·종료 시각이 든 JSON 보고서를 만듭니다.

| 항목 | 계약 |
| --- | --- |
| 최대 입력 | 4096바이트, 뒤쪽 공백도 포함 |
| JSON | 객체 하나, 알 수 없는 필드 거부 |
| 제목 | 앞뒤 공백 제거 후 비어 있지 않음 |
| 시작 | RFC3339로 파싱 가능한 문자열 |
| 분 | 1~120 정수 |
| 출력 | 제목, UTC 시작, UTC 종료, 마지막 개행 |
| 오류 | 진단은 오류 값으로 전달, 검증 전에 결과 쓰지 않음 |

v1의 중복 이름·대소문자 매칭 제한은 3절과 같습니다.
입력 검증이 끝난 뒤 출력하므로 잘못된 입력으로 결과 일부를 쓰지는 않습니다.
출력 Writer 자체가 실패할 때 일부 바이트가 이미 쓰였을 가능성은 있습니다.

### 9.2 기존 개인 파일을 보존하며 준비하기

macOS/Linux/WSL Bash:

```sh
mkdir -p _practice/15-data-time
if [ ! -e _practice/15-data-time/main.go ]; then
  cp examples/15-data-time/json-decode/main.go _practice/15-data-time/main.go
fi
```

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/15-data-time | Out-Null
if (-not (Test-Path _practice/15-data-time/main.go)) {
    Copy-Item examples/15-data-time/json-decode/main.go _practice/15-data-time/main.go
}
```

두 명령은 이미 작성한 `main.go`를 덮어쓰지 않습니다.
이제 개인 파일 내용을 아래 전체 코드로 바꿉니다.
기존 `Task` 선언이나 `main`을 남겨 중복하지 않습니다.

### 9.3 전체 코드

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Request struct {
	Title   string `json:"title"`
	Start   string `json:"start"`
	Minutes int    `json:"minutes"`
}

type Report struct {
	Title string `json:"title"`
	Start string `json:"start"`
	End   string `json:"end"`
}

func createReport(source io.Reader, destination io.Writer) error {
	data, err := io.ReadAll(io.LimitReader(source, 4097))
	if err != nil {
		return fmt.Errorf("입력 읽기: %w", err)
	}
	if len(data) > 4096 {
		return errors.New("입력은 4096바이트까지 허용합니다")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var request Request
	if err := decoder.Decode(&request); err != nil {
		return fmt.Errorf("예약 JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return fmt.Errorf("예약 끝 확인: %w", err)
		}
		return errors.New("예약은 하나만 허용합니다")
	}
	request.Title = strings.TrimSpace(request.Title)
	if request.Title == "" || request.Minutes < 1 || request.Minutes > 120 {
		return errors.New("제목과 1~120분이 필요합니다")
	}
	start, err := time.Parse(time.RFC3339, request.Start)
	if err != nil {
		return fmt.Errorf("시작 시각: %w", err)
	}
	end := start.Add(time.Duration(request.Minutes) * time.Minute)
	report := Report{
		Title: request.Title,
		Start: start.UTC().Format(time.RFC3339),
		End:   end.UTC().Format(time.RFC3339),
	}
	if err := json.NewEncoder(destination).Encode(report); err != nil {
		return fmt.Errorf("보고서 쓰기: %w", err)
	}
	return nil
}

func run() error {
	input := `{"title":"Go","start":"2026-11-09T09:30:00+09:00","minutes":25}`
	return createReport(strings.NewReader(input), os.Stdout)
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
gofmt -w _practice/15-data-time/main.go
go run ./_practice/15-data-time
```

예상 결과:

```text
{"title":"Go","start":"2026-11-09T00:30:00Z","end":"2026-11-09T00:55:00Z"}
```

### 9.4 왜 4097바이트를 읽는가?

4096바이트까지만 읽으면 실제로 더 긴 입력이 잘렸는지 알 수 없습니다.
최대보다 한 바이트 더 읽고 길이가 4096보다 큰지 확인합니다.
14장의 제한 읽기와 같은 원리입니다.

제한은 입력 전체의 바이트 수에 적용합니다.
첫 JSON 값이 짧아도 뒤의 공백이 많으면 초과할 수 있습니다.
“파싱한 객체의 크기”나 “글자 수” 제한과 구분합니다.

`LimitReader`는 시간 제한이 아닙니다.
Reader가 더 이상 바이트를 주지 않고 멈추면 그 읽기가 기다릴 수 있습니다.
취소 가능한 작업의 협력 방식은 17장에서 배웁니다.

### 9.5 입력을 바꾸며 확인하기

`run`의 입력 문자열만 한 번에 하나씩 바꿉니다.
오류 사례는 정상 출력이 없고 오류 메시지가 표준 오류에 나와야 합니다.

| 변경 | 결과 |
| --- | --- |
| 원래 입력 | 종료 `00:55:00Z` |
| 분 1 | 종료 `00:31:00Z` |
| 분 120 | 종료 `02:30:00Z` |
| 분 0 | 범위 오류 |
| 분 121 | 범위 오류 |
| 제목 `" Go "` | 결과 제목 `"Go"` |
| 제목 공백만 | 제목 오류 |
| 시작 `"2026-11-09 09:30"` | RFC3339 형식 오류 |
| 시작 `"2026-02-30T09:30:00Z"` | 날짜 오류 |
| 시작 `"2026-11-09T00:30:00Z"` | 원래 입력과 같은 UTC 결과 |
| 뒤에 `{}` 추가 | 두 번째 값 오류 |
| `memo` 필드 추가 | 알 수 없는 필드 오류 |

검증 함수에서 `os.Exit`를 호출하지 않았습니다.
오류가 호출자까지 전달되므로 테스트는 성공·실패를 값으로 관찰할 수 있습니다.
종료 코드를 선택하는 일은 16장에서 더 자세히 분리합니다.

## 10. 자주 하는 실수

| 증상 | 확인할 원인 | 수정 방향 |
| --- | --- | --- |
| JSON이 `{}`로 출력됨 | 구조체 필드가 비공개인지 | 공개 필드와 태그 확인 |
| false가 사라짐 | bool에 `omitempty`를 붙였는지 | 생략 의미를 정하고 필요하면 포인터 사용 |
| 누락과 null을 구분하지 못함 | 포인터만으로 상태를 표현했는지 | 별도 존재 상태 설계 |
| JSON 하나 뒤의 쓰레기 문자가 무시됨 | 첫 `Decode`만 호출했는지 | 끝을 확인하는 두 번째 읽기 |
| 모르는 필드가 조용히 넘어감 | `DisallowUnknownFields` 미사용 | 객체 계약에 맞춰 설정 |
| CSV 이름이 두 칸으로 갈라짐 | 쉼표로 직접 Split했는지 | `csv.Reader` 사용 |
| CSV 출력이 비어 있음 | Flush 누락인지 | Flush 후 Error 확인 |
| CSV 마지막 쓰기 실패를 놓침 | Write 오류만 봤는지 | Writer.Error 확인 |
| 25분 대신 아주 짧은 시간이 더해짐 | Duration 단위를 빠뜨렸는지 | `time.Minute` 곱하기 |
| 같은 순간인데 비교가 다름 | 문자열 또는 `==`로 비교했는지 | 파싱 후 Equal |
| 다음 날이 24시간 뒤와 다름 | 지역의 오프셋 변경인지 | 달력과 고정 기간 구분 |

### 10.1 오류를 문자열 전체로만 검사하지 않기

표준 라이브러리 오류에는 입력 위치나 타입 이름이 포함될 수 있습니다.
테스트에서 모든 오류 문장을 고정하면 구현 세부 변화에 민감해집니다.
우리 함수의 계약인 성공 여부, 제로 결과, 래핑된 원인 등을 먼저 확인합니다.

CSV의 `*csv.ParseError`나 JSON의 타입 오류처럼
추가 정보가 필요하면 `errors.As`를 사용할 수 있습니다.
아직 필요하지 않은 프로그램에 타입별 분기를 모두 추가하지는 않습니다.

예제 테스트는 Reader·Writer가 직접 돌려준 오류가
`errors.Is`로 확인되는지도 검사합니다.
입력과 출력 실패가 단순한 “잘못된 데이터” 오류에 덮이지 않도록 하기 위함입니다.

### 10.2 숫자를 `any`로 읽을 때

알려진 구조를 가진 설정은 구체적인 구조체를 우선 사용했습니다.
v1에서 JSON 숫자를 `any`로 읽으면 기본적으로 float64가 됩니다.
아주 큰 정수를 정확히 다뤄야 한다면 이런 기본값을 주의해야 합니다.

`Decoder.UseNumber`는 인터페이스 값의 숫자를 `json.Number`로 받게 합니다.
그 뒤 `Int64` 같은 변환의 오류와 범위를 확인합니다.
[UseNumber](https://pkg.go.dev/encoding/json#Decoder.UseNumber)를 참고합니다.

다음은 독립된 전체 프로그램입니다.
기본 예제의 작업 구조체를 이 코드로 교체할 필요는 없습니다.

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func run() error {
	decoder := json.NewDecoder(strings.NewReader("9007199254740993"))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	number, ok := value.(json.Number)
	if !ok {
		return fmt.Errorf("숫자가 아닙니다")
	}
	integer, err := number.Int64()
	if err != nil {
		return err
	}
	fmt.Println(integer)
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

구체적인 int64 필드가 있는 구조체로 받을 수 있다면 그 방법이 더 직접적입니다.
`UseNumber`가 JSON 숫자를 무제한 크기 정수로 자동 계산해 주는 것은 아닙니다.
이 예제의 `Int64`도 표현 범위를 넘으면 오류입니다.

## 11. 연습문제와 검증

[연습문제](../../exercises/15-data-time/README.md)는
확인 → 변형 → 응용 순서의 기본 6문제와 선택 1문제입니다.
[해설](../../solutions/15-data-time/README.md)은 요구사항을 시도한 뒤 읽습니다.

| 문제 | 학습 결과 |
| --- | --- |
| 1 | JSON 누락·null·false의 차이 설명 |
| 2 | 4096바이트 설정 JSON 읽기 |
| 3 | CSV 인용과 Flush 오류 수정 |
| 4 | CSV 학습 시간 합계와 잘못된 레코드 처리 |
| 5 | 시작 포함·끝 제외 시간 구간 판정 |
| 6 | 검증한 CSV를 JSON 배열로 변환 |
| 선택 | 큰 숫자를 손실 없이 읽고 변환 오류 검사 |

장 범위 검사는 저장소 루트에서 실행합니다.

```sh
go test ./examples/15-data-time/... ./solutions/15-data-time/...
go vet ./examples/15-data-time/... ./solutions/15-data-time/...
go build ./examples/15-data-time/... ./solutions/15-data-time/...
```

테스트에는 정상 출력뿐 아니라 다음 경계가 포함되어 있습니다.

- JSON의 빈 입력, null, 모르는 필드, 두 번째 값, 잘못된 타입.
- 4096바이트와 4097바이트 설정.
- CSV의 인용된 쉼표·줄바꿈·큰따옴표와 헤더만 있는 입력.
- CSV 레코드 1000개와 1001개.
- 읽기·쓰기 실패와 오류 원인의 전달.
- 시간 구간의 시작·끝과 서로 다른 오프셋의 동일 순간.

샘플 출력은 고정된 데이터와 시각을 사용합니다.
실제 서비스의 현재 시각을 테스트할 때는 값을 주입하는 설계가 도움이 됩니다.
현재 시각 자체를 고정된 출력 정답으로 저장하지 않습니다.

## 12. 자기 점검

해설을 접은 상태에서 각 질문에 한 문장으로 답해 봅니다.

1. JSON 문법이 맞아도 `minutes: -1`을 거부할 이유는 무엇인가?
2. `Decode` 한 번 성공한 뒤에 왜 다시 읽는가?
3. `DisallowUnknownFields`가 중복 이름까지 거부하는가?
4. `*bool`에서 nil과 false 포인터는 어떻게 다른가?
5. 이 예제에서 누락과 null은 서로 구분되는가?
6. CSV의 물리적인 한 줄이 항상 레코드 하나인가?
7. `csv.Writer`에서 Flush만 호출하면 오류 확인이 끝나는가?
8. `Parse`와 `ParseInLocation`은 시간대 없는 입력을 어떻게 해석하는가?
9. 같은 순간인지 비교할 때 왜 Equal을 쓰는가?
10. `AddDate`의 하루와 `24*time.Hour`는 항상 같은가?

<details>
<summary>자기 점검 해설</summary>

1. 문법은 숫자 표현을 확인하고, 허용 범위는 프로그램의 의미 규칙이다.
2. 첫 호출은 값 하나만 읽으므로 뒤에 또 다른 값이나 잘못된 문자가 남을 수 있다.
3. v1에서 이 옵션은 구조체에 대응하지 않는 필드를 검사하며 중복 이름 정책은 별개다.
4. nil은 지정된 bool이 없는 상태이고 false 포인터는 지정된 값이 false인 상태다.
5. 새 구조체로 읽는 이 예제에서는 둘 다 nil이므로 구분되지 않는다.
6. 인용된 필드에 줄바꿈이 있으면 레코드 하나가 여러 줄에 걸칠 수 있다.
7. Flush 뒤 Error를 호출해 내부 쓰기 오류를 확인해야 한다.
8. Parse는 UTC, ParseInLocation은 전달한 Location을 사용한다.
9. Equal은 순간의 동일성을 비교하며 Time의 구조적 동일성과 목적이 다르다.
10. 지역의 오프셋 변경이 있으면 다음 날짜 같은 시각까지의 길이가 24시간과 다를 수 있다.

</details>

## 13. 핵심 정리와 다음 장

형식을 읽는 함수와 프로그램의 의미를 검증하는 코드를 나누면
잘못된 입력이 어느 단계에서 거부되는지 설명할 수 있습니다.
JSON의 상태 표현, CSV의 레코드 규칙, 시간의 단위를 그 경계에서 확인합니다.

입력을 받는 일과 결과를 출력하는 일을 Reader·Writer로 표현했기 때문에
파일이나 표준 입출력으로 연결하는 데 처리 로직 전체를 바꿀 필요는 없습니다.
다음 장에서는 이 구조를 명령줄 프로그램의 인자·환경 변수·종료 코드에 연결합니다.

- 이전: [14. 파일과 스트림](../14-files-io/README.md)
- 다음: [16. 명령줄 프로그램](../16-cli/README.md)
- [전체 목차](../README.md)
- [연습문제](../../exercises/15-data-time/README.md)
- [해설](../../solutions/15-data-time/README.md)

## 참고 자료

- [encoding/json](https://pkg.go.dev/encoding/json)
- [JSON v1과 v2의 차이](https://pkg.go.dev/encoding/json#hdr-Migrating_to_v2)
- [encoding/csv](https://pkg.go.dev/encoding/csv)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
- [time](https://pkg.go.dev/time)
- [time.ParseInLocation](https://pkg.go.dev/time#ParseInLocation)
- [time.Time.AddDate](https://pkg.go.dev/time#Time.AddDate)
- [time/tzdata](https://pkg.go.dev/time/tzdata)

확인일: 2026-11-09. 코드와 출력·경계·오류 처리는 Go 1.27.3에서 확인했습니다.
