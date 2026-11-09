# 14. 파일과 스트림

| 난이도 | 선수 학습 | 준비물 |
| --- | --- | --- |
| 기초 응용 | 슬라이스·문자열, 인터페이스, 오류와 `defer`, 테스트 | Go, 터미널, 이 저장소 |

지금까지의 예제는 코드 안에 입력을 적었습니다.
프로그램을 다시 실행해도 기록을 남기려면 파일에 저장해야 하고,
다른 프로그램의 결과를 받으려면 바이트가 들어오는 경로를 다룰 수 있어야 합니다.

이번 장에서는 파일 이름보다 먼저 **읽는 쪽과 쓰는 쪽의 계약**을 배웁니다.
같은 처리 함수를 문자열, 파일, 메모리 버퍼에 연결하면
디스크를 직접 사용하지 않고도 계산과 오류 처리를 검사할 수 있습니다.

파일을 열었다는 사실만으로 작업이 끝난 것은 아닙니다.
읽기·쓰기 도중의 실패, 버퍼에 남은 데이터, 닫기 실패까지
어느 함수가 책임질지 정해야 합니다.

## 학습 목표

- `io.Reader`와 `io.Writer`로 입력·출력을 분리한다.
- 읽은 바이트 수와 `EOF`를 구분하고 마지막 데이터를 잃지 않는다.
- 작은 파일 전체 읽기와 스트림 처리를 목적에 맞게 선택한다.
- 파일을 연 함수가 닫기와 오류 전달까지 책임지게 작성한다.
- 줄 처리, 버퍼 비우기, 경로 조합과 입력 크기 제한을 테스트한다.

## 사전 지식과 실행 환경

모든 명령은 **저장소 루트**에서 실행합니다.
현재 모듈의 Go 버전은 [`go.mod`](../../go.mod)를 따릅니다.
하위 디렉터리에 새 모듈을 만들지 않습니다.

인터페이스는 09장, `%w`·`errors.Is`·`defer`는 11장,
`testing`과 `t.TempDir`은 12장의 내용을 사용합니다.
이번 장에서 처음 만나는 API는 사용 위치에서 설명합니다.

실행 예제의 실제 파일은 운영체제의 임시 디렉터리에 만듭니다.
프로그램이 만든 디렉터리만 정리하며, 사용자의 기존 파일은 입력으로 사용하지 않습니다.
임시 경로의 무작위 부분은 출력하지 않아 운영체제별 결과를 비교할 수 있습니다.

| 예제 | 중심 내용 | 실행 명령 |
| --- | --- | --- |
| [stream-copy](../../examples/14-files-io/stream-copy/main.go) | Reader에서 Writer로 복사 | `go run ./examples/14-files-io/stream-copy` |
| [read-chunks](../../examples/14-files-io/read-chunks/main.go) | 작은 버퍼와 EOF | `go run ./examples/14-files-io/read-chunks` |
| [file-roundtrip](../../examples/14-files-io/file-roundtrip/main.go) | 임시 파일 전체 읽기·쓰기 | `go run ./examples/14-files-io/file-roundtrip` |
| [open-close](../../examples/14-files-io/open-close/main.go) | 열린 파일의 소유권 | `go run ./examples/14-files-io/open-close` |
| [scan-lines](../../examples/14-files-io/scan-lines/main.go) | 줄 단위 처리와 오류 확인 | `go run ./examples/14-files-io/scan-lines` |
| [buffered-write](../../examples/14-files-io/buffered-write/main.go) | 버퍼와 Flush | `go run ./examples/14-files-io/buffered-write` |
| [paths](../../examples/14-files-io/paths/main.go) | 경로 조합과 디렉터리 목록 | `go run ./examples/14-files-io/paths` |

본문의 전체 코드는 개인 파일에 따로 작성할 수 있습니다.
서로 다른 프로그램의 `main`을 한 폴더에 이어 붙이지 않습니다.
실패를 보여 주는 입력은 기본 예제의 성공 경로와 구분합니다.

## 1. 파일보다 먼저 바이트의 흐름 생각하기

### 1.1 스트림은 순서가 있는 입력과 출력

**스트림(stream)**은 데이터를 순서대로 읽거나 쓰는 방식입니다.
파일이 가장 익숙한 사례지만 문자열, 압축 데이터, 네트워크 연결도
읽고 쓰는 관점에서는 비슷한 역할을 할 수 있습니다.

이번 장은 네트워크를 사용하지 않습니다.
문자열을 읽는 객체와 메모리에 쓰는 객체를 연결해
파일과 무관한 부분부터 확인합니다.

`io.Reader`의 핵심 메서드는 `Read(p []byte) (n int, err error)`입니다.
호출자는 저장 공간 `p`를 제공하고, 구현은 그 공간에 읽은 바이트를 채웁니다.
반환된 `n`만큼이 이번에 유효한 데이터입니다.

`io.Writer`의 핵심 메서드는 `Write(p []byte) (n int, err error)`입니다.
이쪽에서는 전달한 바이트를 대상에 쓰고 실제 쓴 개수를 돌려줍니다.
구체적인 계약은 [Reader](https://pkg.go.dev/io#Reader)와
[Writer](https://pkg.go.dev/io#Writer)에 있습니다.

| 질문 | 읽는 쪽 | 쓰는 쪽 |
| --- | --- | --- |
| 데이터를 누가 제공하나? | Reader 구현 | 호출자 |
| 바이트 슬라이스는 무엇인가? | 채울 저장 공간 | 쓸 내용 |
| `n`은 무엇인가? | 읽은 바이트 수 | 쓴 바이트 수 |
| 파일만 가능한가? | 문자열·버퍼 등도 가능 | 메모리·파일 등도 가능 |

### 1.2 인터페이스를 매개변수로 받는 이유

함수가 `*os.File`만 받으면 테스트도 파일을 준비해야 합니다.
줄을 세는 작업에 파일 이름이나 권한은 필요하지 않습니다.
그럴 때는 필요한 능력인 `io.Reader`를 매개변수로 받습니다.

반대로 파일을 직접 여는 함수는 경로와 자원 수명까지 책임집니다.
이 두 책임을 분리하면 실패 원인을 좁히기 쉬워집니다.

- 읽는 규칙이 잘못되었는지 확인할 때는 문자열 Reader를 사용합니다.
- 파일이 없는 경우를 확인할 때는 임시 경로를 사용합니다.
- 출력 실패를 확인할 때는 일부러 오류를 돌려주는 Writer를 사용합니다.

이것은 모든 함수를 추상화하라는 규칙이 아닙니다.
입력과 출력의 형태가 바뀌어도 유지할 처리 로직에 적용하는 설계 선택입니다.

## 2. `io.Copy`로 연결하기

`io.Copy`는 EOF까지 읽어 대상에 씁니다.
정상적인 입력 끝은 복사 실패가 아니므로 성공 시 오류는 `nil`입니다.
[Copy 문서](https://pkg.go.dev/io#Copy)의 반환 계약입니다.

저장 경로: [`examples/14-files-io/stream-copy/main.go`](../../examples/14-files-io/stream-copy/main.go).

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func run() error {
	source := strings.NewReader("Go\nstream\n")
	var destination bytes.Buffer
	copied, err := io.Copy(&destination, source)
	if err != nil {
		return fmt.Errorf("복사: %w", err)
	}
	_, err = fmt.Printf("복사한 바이트: %d\n%s", copied, destination.String())
	return err
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
go run ./examples/14-files-io/stream-copy
```

예상 출력:

```text
복사한 바이트: 10
Go
stream
```

`strings.NewReader`는 문자열을 차례로 읽는 Reader를 만듭니다.
`bytes.Buffer`는 메모리에 바이트를 모으며 Writer로 사용할 수 있습니다.
`&destination`을 넘기는 것은 버퍼 자체의 상태를 변경해야 하기 때문입니다.

| 순서 | 입력에 남은 내용 | 출력에 쌓인 내용 |
| --- | --- | --- |
| 시작 | `Go\nstream\n` | 빈 버퍼 |
| 복사 중 | 아직 읽지 않은 부분 | 이미 쓴 부분 |
| 완료 | 없음 | `Go\nstream\n` |

`copied`는 줄 수나 글자 수가 아니라 바이트 수입니다.
이 입력은 영문 8바이트와 줄바꿈 2바이트라서 합계 10입니다.
한글을 넣으면 화면에서 보이는 글자 수와 달라질 수 있습니다.

`io.Copy`가 입력 전체를 한 덩어리로 메모리에 보관해야 하는 것은 아닙니다.
다만 이 예제는 목적지가 `bytes.Buffer`이므로 최종 출력 전체가 메모리에 남습니다.
복사 방식과 목적지가 보관하는 양을 따로 생각해야 합니다.

### 2.1 누가 입력과 출력을 닫는가?

`io.Copy`는 전달받은 Reader와 Writer를 닫지 않습니다.
이 예제의 메모리 객체에는 닫을 파일이 없습니다.
열린 파일을 전달했다면 파일을 연 쪽에서 닫기 책임을 정합니다.

`io.Reader`만 받는 함수가 임의로 형 변환하여 파일을 닫으면,
호출자가 같은 입력을 계속 사용하려 할 때 예상하지 못한 실패가 생깁니다.
함수의 타입에 없는 소유권을 숨겨서 추가하지 않습니다.

### 2.2 복사에 실패하면 이미 쓴 내용은?

복사 함수가 오류를 반환했다고 이미 쓴 바이트가 자동 취소되지는 않습니다.
`n > 0`과 오류가 함께 반환될 수 있으며 대상에는 일부 내용이 남을 수 있습니다.
완전한 교체가 필요한 파일 저장은 임시 파일, 검증, 교체 절차를 별도로 설계합니다.

이번 장의 기본 파일 복사는 기존 대상을 덮어쓰지 않도록 만들지만,
실패한 새 대상이 자동 복구되는 거래 기능까지 제공하지는 않습니다.
이 차이는 뒤의 연습문제 5에서 계약으로 명시합니다.

## 3. 직접 `Read`할 때 지켜야 하는 순서

### 3.1 버퍼 크기는 전체 입력 크기가 아니다

4바이트 버퍼를 만든다고 파일 전체가 4바이트여야 하는 것은 아닙니다.
여러 번 읽어 각 결과를 처리하면 더 긴 입력도 다룰 수 있습니다.

아래 함수는 읽은 바이트를 모두 모아 문자열로 반환합니다.
직접 읽는 규칙을 관찰하기 위한 예이며,
전체 내용만 필요한 일반적인 경우에는 `io.ReadAll`이 더 짧습니다.

저장 경로: [`examples/14-files-io/read-chunks/main.go`](../../examples/14-files-io/read-chunks/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func readChunks(source io.Reader) (string, error) {
	buffer := make([]byte, 4)
	var data []byte
	for {
		n, err := source.Read(buffer)
		data = append(data, buffer[:n]...)
		if errors.Is(err, io.EOF) {
			return string(data), nil
		}
		if err != nil {
			return string(data), err
		}
	}
}

func main() {
	text, err := readChunks(strings.NewReader("Go와 파일"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(text)
	fmt.Println("바이트:", len(text))
}
```

실행:

```sh
go run ./examples/14-files-io/read-chunks
```

예상 출력:

```text
Go와 파일
바이트: 12
```

핵심은 `Read` 뒤에서 **오류보다 먼저 이번 데이터를 반영하는 순서**입니다.
Reader는 마지막 바이트와 `io.EOF`를 같은 호출에서 돌려줄 수도 있습니다.
오류를 먼저 보고 반환하면 그 바이트를 잃을 수 있습니다.
[Reader 계약](https://pkg.go.dev/io#Reader)을 참고합니다.

| 가능한 반환 | 먼저 할 일 | 그다음 판단 |
| --- | --- | --- |
| `n > 0`, `err == nil` | `buffer[:n]` 처리 | 다음 읽기 |
| `n > 0`, `err == io.EOF` | 마지막 데이터 처리 | 정상 종료 |
| `n == 0`, `err == io.EOF` | 추가 데이터 없음 | 정상 종료 |
| `n > 0`, 다른 오류 | 이번 데이터 처리 | 부분 결과와 오류 반환 |

`buffer` 전체를 매번 붙이지 않고 `buffer[:n]`만 사용합니다.
마지막 호출은 버퍼를 모두 채우지 못할 수 있습니다.
남은 영역에는 이전 호출의 바이트가 있을 수 있으므로 입력으로 취급하면 안 됩니다.

### 3.2 한글이 버퍼 경계에서 나뉘어도 되는가?

파일 읽기의 기본 단위는 바이트입니다.
4바이트 경계가 한글 문자 경계와 일치할 필요는 없습니다.
이 예제는 바이트를 모두 모은 뒤 문자열을 만들므로 원래 내용을 복원합니다.

각 조각을 독립적인 문자 집합이라고 가정하여 처리하면 문제가 달라집니다.
UTF-8 문자 단위 처리는 07장에서 배운 룬과 인코딩 규칙을 적용해야 합니다.
Reader가 문자 경계까지 맞춰 준다고 가정하지 않습니다.

### 3.3 `0, nil`은 EOF가 아니다

잘 작성된 Reader는 진행 상황이 없는 `0, nil` 반환을 피해야 하지만,
그 값을 EOF라고 해석해서는 안 됩니다.
이 장의 함수는 정상적으로 진행하는 표준 Reader를 입력으로 사용합니다.

어떤 Reader든 무조건 빠르게 끝난다는 보장은 이 함수에 없습니다.
큰 입력의 메모리 사용, 느린 입력과 시간 제한은 서로 다른 문제입니다.
크기 제한은 이 장 후반에서, 취소와 시간 제한은 17장에서 다룹니다.

### 3.4 테스트로 마지막 데이터를 확인하기

예제의 테스트에는 바이트와 EOF를 동시에 반환하는 작은 Reader가 있습니다.
정상 문자열만 읽는 테스트로는 구현이 허용하는 다른 EOF 형태를 놓칠 수 있기 때문입니다.

또한 바이트와 일반 오류를 함께 돌려주는 경우도 확인합니다.
`readChunks`는 그때 부분 문자열과 오류를 모두 반환하는 계약을 택했습니다.
호출자가 오류를 확인하지 않고 부분 결과를 완전한 문서처럼 사용하면 안 됩니다.

## 4. 작은 파일은 전체 읽기·쓰기부터

### 4.1 임시 디렉터리를 사용하는 전체 예제

`os.WriteFile`은 바이트를 파일에 쓰고,
`os.ReadFile`은 파일 전체를 바이트 슬라이스로 읽습니다.
작고 크기가 통제되는 자료에는 이 방식이 이해하기 쉽습니다.

저장 경로: [`examples/14-files-io/file-roundtrip/main.go`](../../examples/14-files-io/file-roundtrip/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-file-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()

	path := filepath.Join(directory, "notes.txt")
	if err := os.WriteFile(path, []byte("Go\n파일\n"), 0o600); err != nil {
		return fmt.Errorf("쓰기: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("읽기: %w", err)
	}
	_, err = fmt.Printf("읽은 바이트: %d\n%s", len(data), data)
	return err
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
go run ./examples/14-files-io/file-roundtrip
```

예상 출력:

```text
읽은 바이트: 10
Go
파일
```

`os.MkdirTemp("", "go-file-")`는 운영체제의 임시 영역에 새 디렉터리를 만듭니다.
빈 첫 인자는 현재 프로젝트 폴더를 뜻하지 않습니다.
반환받은 경로를 사용하고, 성공적으로 만든 뒤 정리 작업을 등록합니다.

`filepath.Join`은 디렉터리와 파일 이름을 운영체제에 맞게 연결합니다.
무작위 디렉터리 이름을 예상하여 직접 조합하지 않습니다.

### 4.2 정리 실패도 오류다

`run`의 반환값 `err`에 이름을 붙인 이유는 정리 결과를 반영하기 위해서입니다.
지연 함수는 작업 오류와 `RemoveAll` 오류를 `errors.Join`으로 합칩니다.
모두 `nil`이면 결과 역시 `nil`입니다.

이 패턴의 순서를 추적해 보세요.

| 시점 | 할 일 | 책임 |
| --- | --- | --- |
| 임시 디렉터리 생성 성공 | 정리 함수 등록 | `run` |
| 파일 쓰기 실패 | 오류를 반환 경로에 저장 | `run` |
| 함수가 돌아오기 직전 | 임시 디렉터리 정리 | 지연 함수 |
| 최종 반환 | 작업·정리 오류 전달 | `run` |
| 오류가 있으면 | 진단 후 실패 종료 | `main` |

`os.Exit`는 지연 함수를 실행하지 않습니다.
따라서 작업 중간에서 직접 호출하지 않고,
정리를 마친 `run`이 반환한 뒤 `main`에서 사용합니다.
[Exit](https://pkg.go.dev/os#Exit)와 [errors.Join](https://pkg.go.dev/errors#Join)을 참고합니다.

진단을 `os.Stderr`에 쓰는 마지막 문장은 오류를 알리기 위한 최선의 시도입니다.
진단 자체를 쓸 수 없는 상황까지 다시 같은 경로로 출력하며 반복하지는 않습니다.
실제 종료 코드와 스트림 계약은 16장에서 자세히 다룹니다.

### 4.3 `WriteFile`의 기존 파일 처리

`WriteFile`은 기존 파일을 유지하는 복사 함수가 아닙니다.
대상이 이미 있다면 기존 내용을 줄이고 새 내용을 쓰는 동작을 할 수 있습니다.
따라서 기존 개인 파일을 시험 대상으로 지정하면 안 됩니다.

권한 인자 `0o600`은 새 파일을 만들 때의 요청 권한입니다.
기존 파일의 권한을 고쳐 주는 설정이 아니며,
운영체제의 권한 모델과 생성 마스크의 영향도 받습니다.
[WriteFile](https://pkg.go.dev/os#WriteFile)의 설명을 확인합니다.

이 장에서 `0o600`은 읽고 쓰는 파일 예제의 보수적인 기본값입니다.
Windows와 Unix의 권한 동작이 완전히 같다고 단정하지 않습니다.
권한 비트의 출력값을 플랫폼 공통 테스트의 고정 정답으로 사용하지 않습니다.

## 5. `Open`한 파일의 수명 관리

### 5.1 열기·읽기·닫기를 한 함수에 모으기

전체 내용을 읽어도 파일을 직접 여는 방법을 알아 둘 필요가 있습니다.
스트림 처리로 바꿀 때 같은 수명 관리 구조를 사용하기 때문입니다.

저장 경로: [`examples/14-files-io/open-close/main.go`](../../examples/14-files-io/open-close/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func readText(path string) (data []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("열기: %w", err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()
	data, err = io.ReadAll(file)
	return data, err
}

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-open-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()

	path := filepath.Join(directory, "lesson.txt")
	if err := os.WriteFile(path, []byte("열고 읽고 닫기\n"), 0o600); err != nil {
		return err
	}
	data, err := readText(path)
	if err != nil {
		return err
	}
	_, err = fmt.Print(string(data))
	return err
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
go run ./examples/14-files-io/open-close
```

예상 출력:

```text
열고 읽고 닫기
```

`readText`가 `os.Open`을 호출했으므로 이 함수가 파일을 닫습니다.
열기에 실패하면 파일을 사용하지 않고 즉시 오류를 반환합니다.
`defer`는 성공적으로 연 뒤에 등록합니다.

### 5.2 `defer file.Close()`만으로 충분할까?

그 표현은 파일을 닫기는 하지만 `Close`의 반환 오류를 사용하지 않습니다.
읽기 전용 파일에서 어떤 오류를 우선할지,
쓰기 파일에서 닫기 실패를 성공으로 볼지 결정해야 합니다.

예제는 정리 실패도 놓치지 않도록 이름 있는 반환값과 `errors.Join`을 사용합니다.
이는 프로젝트의 선택이며, 모든 함수가 반드시 같은 형식이어야 하는 것은 아닙니다.
중요한 것은 닫기 오류를 어떤 정책으로 처리했는지 설명할 수 있는가입니다.

| 상태 | 반환할 정보 |
| --- | --- |
| 열기 실패 | 경로를 여는 데 실패한 오류 |
| 읽기 실패, 닫기 성공 | 읽기 오류 |
| 읽기 성공, 닫기 실패 | 닫기 오류 |
| 읽기 실패, 닫기 실패 | 두 오류 모두 |

`Close`는 파일을 더 이상 입출력에 사용할 수 없게 합니다.
이미 닫힌 파일을 다시 닫는 것도 오류가 될 수 있으므로,
명시적 닫기와 무조건 실행되는 지연 닫기를 섞지 않습니다.
[File.Close](https://pkg.go.dev/os#File.Close)가 기준입니다.

### 5.3 반복 안의 `defer` 주의

`defer`는 반복 회차가 끝날 때 실행되는 문장이 아닙니다.
현재 함수가 돌아갈 때 실행됩니다.
긴 반복에서 파일을 계속 열고 지연 닫기만 등록하면 열린 파일이 누적될 수 있습니다.

파일 하나의 처리를 별도 함수로 나누면 그 함수가 끝날 때 닫힙니다.
이 예제의 `readText`를 여러 번 호출하면 각 호출이 자신의 수명을 마칩니다.
특정 운영체제의 파일 개수 제한에 기대어 정상 동작을 판단하지 않습니다.

### 5.4 파일이 없다는 사실을 문자열로 비교하지 않기

경로가 없을 때 나오는 전체 오류 문장은 운영체제와 경로에 따라 달라집니다.
`strings.Contains(err.Error(), "not found")` 같은 검사는 안정적인 계약이 아닙니다.
이 예제의 테스트는 `errors.Is(err, os.ErrNotExist)`를 사용합니다.

`fmt.Errorf("열기: %w", err)`로 감쌌어도 원인을 확인할 수 있습니다.
오류에 문맥을 추가하는 일과 원인을 분류하는 일은 함께 사용할 수 있습니다.
[os의 오류 값](https://pkg.go.dev/os#pkg-variables)과 11장의 오류 체인을 참고합니다.

## 6. 줄 단위로 읽기

### 6.1 빈 줄을 제외한 개수 세기

파일 전체가 아니라 한 줄씩 처리하고 싶다면 `bufio.Scanner`가 편리합니다.
기본 분할 규칙은 줄 단위이며, 줄 끝 구분자는 토큰에서 제거됩니다.
[Scanner](https://pkg.go.dev/bufio#Scanner)와 [ScanLines](https://pkg.go.dev/bufio#ScanLines)를 참고합니다.

저장 경로: [`examples/14-files-io/scan-lines/main.go`](../../examples/14-files-io/scan-lines/main.go).

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func countNonEmpty(source io.Reader) (int, error) {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("줄 읽기: %w", err)
	}
	return count, nil
}

func main() {
	count, err := countNonEmpty(strings.NewReader("Go\n\n  \n한글"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("비어 있지 않은 줄:", count)
}
```

실행:

```sh
go run ./examples/14-files-io/scan-lines
```

예상 출력:

```text
비어 있지 않은 줄: 2
```

입력에는 네 줄에 해당하는 내용이 있습니다.
두 번째는 빈 줄, 세 번째는 공백뿐인 줄입니다.
`TrimSpace` 후 비어 있지 않은 두 줄만 셉니다.

| 입력 토큰 | 공백 제거 뒤 | 개수 증가 |
| --- | --- | --- |
| `Go` | `Go` | 예 |
| 빈 문자열 | 빈 문자열 | 아니요 |
| 공백 두 개 | 빈 문자열 | 아니요 |
| `한글` | `한글` | 예 |

마지막 `한글` 뒤에 줄바꿈이 없어도 마지막 토큰으로 읽습니다.
반대로 마지막 줄바꿈 뒤에 자동으로 새 빈 토큰 하나가 더 생긴다고 가정하면 안 됩니다.
줄 수의 정의를 먼저 정한 뒤 입력별 결과를 확인해야 합니다.

### 6.2 반복 종료 뒤 `Err`가 필요한 이유

`Scan()`의 `false`는 정상 끝일 수도 있고 읽기 실패일 수도 있습니다.
반복문 밖에서 `scanner.Err()`를 확인해야 둘을 구분할 수 있습니다.

예제는 오류가 발생하면 누적 개수를 0으로 되돌려 오류와 함께 반환합니다.
불완전한 문서의 개수를 정상 결과처럼 전달하지 않기 위한 계약입니다.
앞의 `readChunks`가 부분 데이터를 돌려주는 정책과 비교해 보세요.

### 6.3 긴 줄의 상한

Scanner에는 토큰 크기 상한이 있습니다.
예제에서는 처음에 1KiB 버퍼를 주고 필요하면 128KiB까지 늘릴 수 있게 했습니다.
이 값은 입력 파일 전체의 크기 제한이 아닙니다.

구분자 등을 담을 공간도 필요하므로,
“텍스트가 정확히 128KiB이면 무조건 성공”이라는 계약으로 읽지 않습니다.
업무상 허용할 최대 줄 길이를 정하고 그 경계의 실제 입력을 테스트합니다.
[Scanner.Buffer](https://pkg.go.dev/bufio#Scanner.Buffer)를 참고합니다.

`Buffer`나 `Split` 설정은 스캔을 시작하기 전에 합니다.
이미 읽기 시작한 뒤 바꾸는 것은 허용되지 않습니다.
이번 예제의 긴 줄 테스트는 제한 초과가 오류로 전달되는지 확인합니다.

### 6.4 `Text`와 `Bytes`를 저장할 때

`Text()`는 현재 토큰의 문자열을 제공합니다.
`Bytes()`는 내부 버퍼를 가리킬 수 있으므로 다음 스캔 뒤에도
그 슬라이스가 그대로 유지된다고 가정하면 안 됩니다.

나중에 사용할 바이트를 저장하려면 복사 책임을 따로 생각해야 합니다.
이 장은 각 줄을 바로 읽고 처리하므로 복사된 토큰을 모으지 않습니다.
[Scanner.Bytes](https://pkg.go.dev/bufio#Scanner.Bytes)의 수명 설명을 확인합니다.

## 7. 버퍼에 쓴 것과 대상에 쓴 것

### 7.1 `Flush` 전후 관찰하기

**버퍼링(buffering)**은 작은 쓰기를 메모리에 모았다가 대상에 전달하는 방식입니다.
성능 이점이 있는지는 실제 사용 조건에서 측정해야 하지만,
우선 데이터가 어느 단계에 있는지 이해해야 합니다.

저장 경로: [`examples/14-files-io/buffered-write/main.go`](../../examples/14-files-io/buffered-write/main.go).

```go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

func run() error {
	var destination bytes.Buffer
	writer := bufio.NewWriter(&destination)
	if _, err := fmt.Fprintln(writer, "Go"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "buffer"); err != nil {
		return err
	}
	fmt.Println("Flush 전:", destination.Len())
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("버퍼 비우기: %w", err)
	}
	_, err := fmt.Printf("Flush 후: %d\n%s", destination.Len(), destination.String())
	return err
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
go run ./examples/14-files-io/buffered-write
```

예상 출력:

```text
Flush 전: 0
Flush 후: 10
Go
buffer
```

`fmt.Fprintln(writer, ...)`는 이 예제에서 작은 내용을 버퍼에 넣습니다.
이 시점에는 실제 목적지인 `destination`의 길이가 0입니다.
`Flush`를 호출하면 모인 바이트를 아래 Writer로 보냅니다.

| 관찰 지점 | 버퍼에 모인 내용 | 목적지 길이 |
| --- | --- | --- |
| 두 번 쓴 뒤 | `Go\nbuffer\n` | 0 |
| Flush 성공 뒤 | 전달 완료 | 10 |

이 숫자는 이번 입력과 기본 버퍼 크기에서의 결과입니다.
버퍼를 채우는 큰 쓰기는 Flush 전에 아래 Writer로 전달될 수 있습니다.
“버퍼 Writer는 Flush 전에는 절대 쓰지 않는다”로 일반화하지 않습니다.

### 7.2 `Flush`는 파일의 `Close`나 `Sync`가 아니다

`Flush`는 사용자 공간 버퍼의 데이터를 아래 Writer로 전달합니다.
아래 Writer가 파일이라면 파일은 여전히 열려 있습니다.
`File.Sync`가 다루는 저장 장치 반영과도 다른 단계입니다.

작업 순서는 보통 **쓰기 → Flush 오류 확인 → 파일 Close 오류 확인**입니다.
지연 함수에 `Flush`를 넣고 반환 오류를 버리면 마지막 저장 실패를 놓칠 수 있습니다.
[Writer.Flush](https://pkg.go.dev/bufio#Writer.Flush)와
[File.Sync](https://pkg.go.dev/os#File.Sync)를 구분합니다.

### 7.3 실패가 Flush에서 보일 수 있다

다음은 개인 파일에서 실행하는 전체 코드입니다.
실제 디스크를 고장 내지 않고, 오류를 반환하는 Writer를 만들어 확인합니다.

```go
package main

import (
	"bufio"
	"errors"
	"fmt"
)

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) {
	return 0, errors.New("저장 장치 오류")
}

func main() {
	writer := bufio.NewWriter(failedWriter{})
	_, writeErr := fmt.Fprintln(writer, "Go")
	fmt.Println("Write 성공:", writeErr == nil)
	flushErr := writer.Flush()
	fmt.Println("Flush 오류:", flushErr != nil)
}
```

예상 출력:

```text
Write 성공: true
Flush 오류: true
```

첫 번째 쓰기는 메모리에 모으는 단계에서 성공했습니다.
실제 대상에 쓰는 Flush에서 실패를 발견합니다.
따라서 쓰기 호출의 성공만으로 최종 결과를 성공 처리하면 안 됩니다.

이것은 버퍼의 결함이 아니라 단계가 분리되어 있다는 뜻입니다.
테스트에서도 입력 실패와 출력 실패를 각각 만들어 확인할 수 있습니다.
파일 권한을 임의로 바꾸어 실패를 유도하면 운영체제별 차이 때문에 불안정해질 수 있습니다.

## 8. 경로와 디렉터리 목록

### 8.1 문자열 연결 대신 `filepath.Join`

경로는 일반 문자열이지만 운영체제가 해석하는 규칙을 갖습니다.
파일을 여는 경로에는 `path/filepath`를 사용합니다.
URL의 경로를 다루는 `path` 패키지와 구분합니다.

저장 경로: [`examples/14-files-io/paths/main.go`](../../examples/14-files-io/paths/main.go).

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-paths-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()

	for _, name := range []string{"zeta.txt", "alpha.txt"} {
		if err := os.WriteFile(filepath.Join(directory, name), nil, 0o600); err != nil {
			return err
		}
	}
	if err := os.Mkdir(filepath.Join(directory, "notes"), 0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fmt.Printf("%s 디렉터리=%t\n", entry.Name(), entry.IsDir())
	}
	fmt.Println("표시용 경로:", filepath.ToSlash(filepath.Join("data", "notes.txt")))
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
go run ./examples/14-files-io/paths
```

예상 출력:

```text
alpha.txt 디렉터리=false
notes 디렉터리=true
zeta.txt 디렉터리=false
표시용 경로: data/notes.txt
```

`os.ReadDir`로 받은 항목은 이 함수의 계약에 따라 이름순으로 정렬됩니다.
디렉터리인지 확인할 때는 `IsDir`를 사용합니다.
다른 목록 API까지 모두 같은 순서를 보장한다고 가정하지 않습니다.
[ReadDir](https://pkg.go.dev/os#ReadDir)에 근거한 출력입니다.

예제는 실제 파일 이름과 디렉터리 여부만 출력합니다.
운영체제별 절대 경로, 생성 시각, 권한 표시는 고정 정답에 넣지 않습니다.
안정적으로 비교할 정보와 환경에 따라 달라질 정보를 구분합니다.

### 8.2 상대 경로의 출발점

`os.Open("data/input.txt")`는 소스 파일의 위치가 아니라
프로그램의 현재 작업 디렉터리를 기준으로 상대 경로를 해석합니다.
그래서 이 교재는 실행 위치를 저장소 루트로 통일합니다.

IDE의 실행 설정이 다른 폴더를 기준으로 하면 같은 코드도 파일을 찾지 못할 수 있습니다.
파일을 다시 만들기 전에 현재 작업 디렉터리와 전달한 경로를 확인하세요.
작업 디렉터리 확인에는 `os.Getwd`를 사용할 수 있습니다.

`filepath.ToSlash`는 이 예제에서 표시용 경로를 일정한 형태로 만들기 위해 사용했습니다.
파일을 여는 실제 경로를 무조건 문자열 치환하는 기법으로 쓰지 않습니다.
[Join](https://pkg.go.dev/path/filepath#Join)과
[ToSlash](https://pkg.go.dev/path/filepath#ToSlash)를 참고합니다.

### 8.3 `Clean`은 접근 권한 검사가 아니다

`filepath.Clean`이나 `Join`은 경로 표현을 정리합니다.
사용자가 전달한 경로를 안전한 디렉터리에 가두는 완전한 기능은 아닙니다.

`IsLocal`도 문자열 수준의 판단이며 심볼릭 링크가 실제로 가리키는 곳까지 검사하지 않습니다.
이번 예제는 코드에서 정한 파일 이름만 사용합니다.
임의의 외부 경로를 받는 프로그램은 별도의 접근 정책을 설계해야 합니다.
[IsLocal](https://pkg.go.dev/path/filepath#IsLocal)의 한계를 확인합니다.

경로를 안전하게 제한하는 문제와 파일을 올바르게 닫는 문제는 별개입니다.
지금은 자원 수명과 데이터 흐름을 익히고,
서비스가 외부 입력을 받게 되면 해당 경계를 다시 검토합니다.

## 9. 전체 읽기의 크기 제한

`io.ReadAll`은 입력이 끝날 때까지 바이트를 모읍니다.
전체 크기가 통제되지 않으면 메모리 사용도 커질 수 있습니다.
파일 전체 읽기를 선택할 때 입력의 상한을 함께 생각합니다.

`io.LimitReader(source, maximum)`은 정해진 바이트까지만 읽게 하지만,
잘린 입력을 자동으로 오류로 보고하지는 않습니다.
이는 제한된 Reader가 EOF에 도달한 것이기 때문입니다.
[LimitReader](https://pkg.go.dev/io#LimitReader)를 참고합니다.

### 9.1 정확히 제한만큼인가, 더 큰가?

최대 5바이트를 허용하면서 정확히 5바이트인 입력도 받아야 한다고 합시다.
5바이트만 읽으면 원본이 `hello`인지 `hello!`인지 알 수 없습니다.

| 원본 | 5바이트까지만 읽기 | 초과 여부를 알 수 있나? |
| --- | --- | --- |
| `hell` | `hell` | 끝을 확인했다면 제한 이내 |
| `hello` | `hello` | 이 결과만으로는 모름 |
| `hello!` | `hello` | 이 결과만으로는 모름 |

그래서 제한보다 한 바이트 더 읽어 초과 여부를 확인할 수 있습니다.
6바이트를 받으면 원래 허용량을 넘은 것이 분명합니다.
예제의 제한 값 자체도 작은 범위로 검증하여 `maximum+1`의 범위를 관리합니다.

이 방법은 바이트 크기 제한입니다.
한글 글자 수 제한이나 줄 수 제한과는 같은 문제가 아닙니다.
연습문제 6에서는 `"한"`을 2바이트와 3바이트 제한에 각각 넣어 확인합니다.

### 9.2 제한값이 0일 때

0을 허용하면 빈 입력만 정상입니다.
한 바이트라도 있으면 초과입니다.
“제한값이 0이면 무제한”이라는 다른 규칙을 원한다면 계약을 명시적으로 바꿔야 합니다.

우리 문제의 제한 범위는 0~1MiB입니다.
이 값은 Go가 모든 입력에 정한 한계가 아니라 해당 함수의 학습용 정책입니다.
정책값과 라이브러리의 자료형 범위를 혼동하지 않습니다.

## 10. 직접 작성하기: 파일에 줄 번호 붙이기

### 10.1 개인 파일 준비

아래에서 자신의 셸에 맞는 블록 하나를 실행합니다.
기존 `main.go`가 있으면 내용을 보존하고 계속합니다.

Windows PowerShell:

```powershell
$practiceDir = "_practice/14-files-io/lab"
New-Item -ItemType Directory -Force $practiceDir | Out-Null
$practiceMain = Join-Path $practiceDir "main.go"
if (Test-Path -LiteralPath $practiceMain) {
    Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
} else {
    Copy-Item examples/01-first-program/hello/main.go $practiceMain
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
practice_dir="_practice/14-files-io/lab"
mkdir -p "$practice_dir"
if [ -e "$practice_dir/main.go" ]; then
    printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
else
    cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
fi
```

### 10.2 책임을 세 함수로 분리하기

작업 파일은 `_practice/14-files-io/lab/main.go`입니다.
입력 파일을 만들고 번호를 붙인 결과 파일을 읽는 전체 프로그램을 작성합니다.

- `numberLines`는 Reader를 읽어 Writer에 번호를 붙여 씁니다.
- `writeReport`는 파일을 열고 닫고 버퍼를 비웁니다.
- `run`은 임시 작업 공간을 만들고 전체 과정을 연결합니다.

```go
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func numberLines(destination io.Writer, source io.Reader) error {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	number := 0
	for scanner.Scan() {
		number++
		if _, err := fmt.Fprintf(destination, "%d: %s\n", number, scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func writeReport(sourcePath, destinationPath string) (err error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, source.Close())
	}()
	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, destination.Close())
	}()
	writer := bufio.NewWriter(destination)
	if err := numberLines(writer, source); err != nil {
		return err
	}
	return writer.Flush()
}

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-report-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()
	sourcePath := filepath.Join(directory, "source.txt")
	destinationPath := filepath.Join(directory, "report.txt")
	if err := os.WriteFile(sourcePath, []byte("Go\n파일\n"), 0o600); err != nil {
		return err
	}
	if err := writeReport(sourcePath, destinationPath); err != nil {
		return err
	}
	data, err := os.ReadFile(destinationPath)
	if err != nil {
		return err
	}
	_, err = fmt.Print(string(data))
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

포맷하고 실행합니다.

```sh
gofmt -w _practice/14-files-io/lab/main.go
go run ./_practice/14-files-io/lab
```

예상 출력:

```text
1: Go
2: 파일
```

### 10.3 처리 순서와 자원 수명 추적

| 순서 | 함수 | 열린 자원 또는 남은 데이터 |
| --- | --- | --- |
| 1 | `run` | 임시 디렉터리 생성 |
| 2 | `WriteFile` | 입력 파일 생성 후 자체 닫기 |
| 3 | `writeReport` | 입력 파일, 출력 파일 열기 |
| 4 | `numberLines` | 한 줄씩 읽고 버퍼에 쓰기 |
| 5 | `Flush` | 버퍼의 나머지 데이터를 출력 파일에 전달 |
| 6 | `writeReport` 반환 | 출력·입력 파일 닫기 |
| 7 | `ReadFile` | 완성된 결과 읽고 자체 닫기 |
| 8 | `run` 반환 | 임시 디렉터리 정리 |

`writeReport`가 돌아온 뒤에 결과 파일을 읽습니다.
따라서 쓰기 버퍼와 파일의 수명을 종료한 뒤 내용을 확인하는 흐름입니다.
읽을 수 있었다는 사실만으로 Flush 누락이 없다고 추정하는 대신 순서를 코드에 드러냅니다.

대상은 `O_EXCL`로 생성하므로 이미 존재하면 실패합니다.
기존 파일을 보존하는 규칙이 이번 실습의 요구사항입니다.
같은 실행 안에서 `writeReport`를 두 번 호출하면 두 번째는 대상 존재 오류가 납니다.

### 10.4 입력을 바꾸기 전에 예상하기

`WriteFile`에 전달하는 문자열만 바꾸어 확인합니다.

| 입력 문자열 | 예상 결과 |
| --- | --- |
| 빈 문자열 | 출력 없음 |
| `"Go"` | `1: Go` 한 줄 |
| `"Go\n"` | `1: Go` 한 줄 |
| `"Go\n\n끝"` | 세 줄, 두 번째 내용은 빈 문자열 |
| `"가\r\n나\r\n"` | 번호를 붙인 두 줄, 출력 줄 끝은 `\n` |

빈 줄의 결과에는 번호·콜론·공백·줄바꿈이 있습니다.
화면에서 보이지 않는 공백을 검사할 때는 문자열을 `%q`로 출력하거나 바이트로 비교하세요.
참고 구현 테스트는 정확한 문자열을 비교합니다.

원본 파일은 변경하지 않는다는 조건도 확인합니다.
번호를 붙인 결과와 원본 경로를 같게 전달해 덮어쓰는 실습은 하지 않습니다.
기존 경로 보존은 오류 반환과 함께 검사해야 합니다.

## 11. 오류를 읽고 책임을 좁히기

### 11.1 실패 위치를 구분하는 질문

파일 처리의 오류를 모두 “파일이 안 된다”로 묶으면 확인할 곳이 너무 넓습니다.
먼저 어느 단계에서 실패했는지 찾습니다.

| 단계 | 확인할 내용 | 예시 원인 |
| --- | --- | --- |
| 경로 준비 | 작업 디렉터리·파일 이름 | 상대 경로의 출발점이 다름 |
| 열기 | 원본 존재·권한 | 없는 파일 |
| 대상 만들기 | 존재 여부·부모 디렉터리 | `O_EXCL` 대상이 이미 있음 |
| 읽기 | Reader 오류·토큰 크기 | Scanner의 긴 줄 |
| 쓰기 | 반환 개수와 오류 | 아래 Writer가 실패 |
| Flush | 버퍼 아래의 실제 대상 | 뒤늦은 쓰기 실패 |
| Close | 자원 종료 결과 | 파일 종료 실패 |
| 정리 | 자신이 만든 경로 | 임시 디렉터리 제거 실패 |

한 단계에서 오류를 반환하면 아래 단계의 성공 출력을 실행하지 않습니다.
반면 정리 작업은 성공과 실패 경로 모두에서 필요할 수 있습니다.
`defer`와 반환 흐름을 함께 읽는 이유입니다.

### 11.2 자주 하는 실수

| 증상 | 원인 | 수정 방향 |
| --- | --- | --- |
| 마지막 데이터 일부가 사라짐 | `n`을 처리하기 전에 EOF 반환 | 데이터를 먼저 처리 |
| 읽은 내용 뒤에 이상한 바이트 | `buffer[:n]` 대신 전체 버퍼 사용 | 유효한 길이만 사용 |
| 정상 파일 끝을 오류로 출력 | EOF 의미를 처리 목적과 혼동 | 전체 읽기의 정상 종료와 구분 |
| 파일 수가 계속 늘거나 열기 실패 | 긴 함수 반복에서 defer 누적 | 파일별 처리 함수를 분리 |
| 결과 파일 일부가 비어 있음 | Flush 누락 | Flush 결과 확인 |
| 스캔이 중간에 끝났는데 성공 처리 | `scanner.Err()` 누락 | 반복 뒤 오류 확인 |
| 일부 긴 줄만 실패 | Scanner 토큰 제한 | 정책을 정하고 시작 전에 설정 |
| 원래 파일이 사라짐 | 새 파일 생성과 덮어쓰기 혼동 | `O_EXCL` 등 보존 계약 사용 |
| OS에 따라 테스트가 달라짐 | 절대 경로·권한 문자열 고정 | 의미와 원인을 비교 |
| 크기 초과 입력을 정상 처리 | LimitReader를 오류 감지기로 오해 | 한 바이트 더 읽고 검사 |

### 11.3 읽은 뒤 닫았는지 테스트하기

참고 구현은 `t.TempDir()` 아래에서 파일을 생성합니다.
테스트가 끝나면 테스트 도구가 그 디렉터리를 정리합니다.
별도로 만든 개인 데이터 경로에 테스트를 실행하지 않습니다.

파일을 읽은 뒤 지우는 검사도 있습니다.
이는 특히 열린 핸들의 영향을 받는 환경에서 수명 관리 실수를 드러내는 데 도움이 됩니다.
모든 자원 누수를 그 한 검사만으로 증명하거나 부정할 수는 없습니다.

테스트가 확인하는 것은 입력·결과·오류 계약입니다.
특정 함수 호출 횟수나 내부 버퍼 배치에만 의존하는 검사를 기본으로 삼지 않습니다.
구현을 바꾸어도 계약이 유지되면 테스트는 계속 유효해야 합니다.

## 12. 선택 학습: 정해진 바이트 수 읽기

문자열 끝까지 읽는 작업과 고정 크기 헤더를 읽는 작업은 다릅니다.
5바이트 헤더가 반드시 필요하다면 3바이트만 읽고 성공이라고 할 수 없습니다.
`io.ReadFull`은 필요한 길이를 채우는 읽기를 표현합니다.

다음은 독립된 전체 코드입니다.

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

func main() {
	header := make([]byte, 5)
	n, err := io.ReadFull(strings.NewReader("Go!"), header)
	fmt.Printf("읽은 내용: %q\n", header[:n])
	fmt.Println("부족한 입력:", errors.Is(err, io.ErrUnexpectedEOF))
}
```

예상 출력:

```text
읽은 내용: "Go!"
부족한 입력: true
```

`EOF`와 `ErrUnexpectedEOF`는 같은 의미가 아닙니다.
이 호출은 일부 데이터를 읽었지만 요청한 길이를 채우지 못했습니다.
[ReadFull](https://pkg.go.dev/io#ReadFull)의 입력 길이별 반환 규칙을 확인합니다.

| 실제 입력 길이 | 5바이트 ReadFull 결과 |
| --- | --- |
| 0 | `n=0`, `io.EOF` |
| 3 | `n=3`, `io.ErrUnexpectedEOF` |
| 5 이상 | `n=5`, 오류 없음 |

더 긴 입력의 나머지는 이후 읽기에 남습니다.
`ReadFull`이 Reader 전체를 소비하는 함수라고 가정하면 안 됩니다.
레코드 형식과 스트림의 끝을 별개의 기준으로 다루는 연습입니다.

## 13. 연습문제와 검증

[14장 연습문제](../../exercises/14-files-io/README.md)에서 다음을 풉니다.
기본 문제 1~6은 이 장의 API와 앞 장의 인터페이스·오류 처리로 해결할 수 있습니다.

| 문제 | 확인할 내용 |
| --- | --- |
| 1. 데이터와 EOF의 순서 | 마지막 데이터를 처리하기 전에 종료하는 오류 |
| 2. 주석과 빈 줄을 제외한 개수 | Scanner와 입력 정책 |
| 3. 사라진 출력 찾기 | Flush와 실제 대상의 차이 |
| 4. 줄 번호 붙이기 | Reader·Writer, 빈 줄과 쓰기 오류 |
| 5. 기존 파일을 보존하는 복사 | OpenFile 플래그, 닫기와 실패 |
| 6. 크기 제한 읽기 | 제한 경계와 한 바이트 추가 읽기 |
| 선택 A. 텍스트 파일 목록 | ReadDir, 확장자와 안정적인 순서 |

먼저 개인 파일에서 구현한 뒤
[해설과 참고 구현](../../solutions/14-files-io/README.md)을 확인합니다.
참고 구현의 테스트는 개인 폴더에 자동으로 적용되지 않습니다.

```sh
go test -count=1 ./examples/14-files-io/... ./solutions/14-files-io/...
```

파일명이나 API 모양만 따라 썼다고 완료한 것은 아닙니다.
입력·출력·소유권·실패 뒤 남는 상태를 설명할 수 있어야 합니다.
원본 보존과 오류 전달도 관찰 가능한 결과에 포함됩니다.

## 14. 요약과 자기 점검

- Reader와 Writer는 파일 이름이 아니라 읽기·쓰기 능력을 표현합니다.
- `Read`에서는 이번 바이트를 처리한 뒤 오류를 판단합니다.
- EOF는 처리 목적에 따라 정상 끝이며, 짧은 고정 길이 읽기는 다른 오류가 될 수 있습니다.
- 파일을 연 함수가 닫기 책임을 분명히 맡습니다.
- Scanner 반복 뒤에는 `Err`, 버퍼 쓰기 뒤에는 `Flush`를 확인합니다.
- 경로 정리와 경로 접근 정책은 다른 문제입니다.
- 전체 읽기에는 입력 크기와 실패 뒤 부분 결과의 의미가 필요합니다.

### 확인 질문

1. 문자열 Reader를 쓰면 어떤 파일 준비 작업을 생략할 수 있나요?
2. `io.Copy`의 성공 결과가 `io.EOF`가 아니라 `nil`인 이유는 무엇인가요?
3. `n > 0`과 EOF가 함께 오면 어느 것을 먼저 처리하나요?
4. 4바이트 버퍼로 한글을 읽을 때 문자 경계가 나뉠 수 있나요?
5. `defer`는 반복문 회차가 끝날 때 실행되나요?
6. Scanner의 `false`만 보고 정상 종료라고 할 수 있나요?
7. Flush와 파일 Close, Sync는 각각 어떤 단계인가요?
8. `WriteFile`을 기존 개인 파일에 시험하면 왜 위험한가요?
9. 5바이트만 읽어서 5바이트 초과 여부를 알 수 있나요?
10. 상대 경로의 시작은 소스 파일이 있는 폴더인가요?

<details>
<summary>답과 설명</summary>

1. 파일 생성과 삭제 없이 읽기 로직을 확인합니다. 파일 접근 오류는 별도의 파일 테스트로 확인합니다.
2. 이 함수의 목적이 끝까지 복사하는 것이므로 입력의 끝을 정상 완료로 해석합니다.
3. 읽은 바이트를 먼저 처리한 뒤 EOF로 종료합니다.
4. 가능합니다. 바이트를 모은 뒤 해석하거나 문자 단위 처리 규칙을 적용해야 합니다.
5. 아닙니다. 현재 함수가 돌아갈 때 실행되므로 긴 반복에서는 자원 누적에 주의합니다.
6. 아닙니다. 반복 뒤 `Err()`로 읽기 실패와 토큰 오류를 확인합니다.
7. Flush는 버퍼를 아래 Writer로 전달하고, Close는 파일 사용을 끝내며, Sync는 저장 장치 반영을 요청합니다.
8. 기존 내용을 보존하는 함수가 아니기 때문입니다. 임시 경로 또는 보존 계약을 사용합니다.
9. 알 수 없습니다. 최대값보다 한 바이트 더 읽는 등의 방법이 필요합니다.
10. 프로그램의 현재 작업 디렉터리입니다. 실행 설정과 경로를 함께 확인합니다.

</details>

### 완료 체크리스트

- [ ] Reader·Writer를 입력과 출력의 역할로 설명했다.
- [ ] EOF와 마지막 바이트가 함께 오는 경우를 확인했다.
- [ ] 자신이 연 파일의 닫기 책임과 오류 정책을 정했다.
- [ ] 빈 줄·마지막 줄바꿈 없음·긴 줄을 구분했다.
- [ ] Flush 실패를 실제 Writer 오류로 테스트했다.
- [ ] 기존 대상 파일이 오류 뒤에도 유지되는지 확인했다.
- [ ] 크기 제한이 바이트 기준임을 한글 입력으로 확인했다.
- [ ] 임시 작업 공간을 정리하고 기본 문제를 마쳤다.

## 참고 자료

- [io](https://pkg.go.dev/io) — Reader·Writer·Copy·ReadFull·LimitReader
- [os](https://pkg.go.dev/os) — 파일 생성·읽기·쓰기·닫기·임시 경로
- [bufio](https://pkg.go.dev/bufio) — Scanner와 버퍼 Writer
- [path/filepath](https://pkg.go.dev/path/filepath) — 운영체제별 경로
- [bytes.Buffer](https://pkg.go.dev/bytes#Buffer) — 메모리의 읽기·쓰기 대상
- [strings.Reader](https://pkg.go.dev/strings#Reader) — 문자열 입력
- [testing.T.TempDir](https://pkg.go.dev/testing#T.TempDir) — 테스트용 임시 디렉터리
- [errors.Join](https://pkg.go.dev/errors#Join) — 작업과 정리 오류를 함께 전달

확인일: 2026-11-09. 코드와 출력·오류 경로는 Go 1.27.3에서 확인했습니다.

## 다음 학습

15장에서는 바이트 스트림 안에 담긴 JSON과 CSV를 구조화된 값으로 바꿉니다.
이 장의 Reader·Writer, 크기 제한과 오류 전달은 그대로 활용합니다.
시간을 문자열로 표현하고 다시 해석할 때도 형식의 계약을 먼저 정합니다.

| 이전 | 목차 | 다음 | 실습 |
| --- | --- | --- | --- |
| [13. 제네릭](../13-generics/README.md) | [전체 목차](../README.md) | [15. 데이터 형식과 시간](../15-data-time/README.md) | [문제](../../exercises/14-files-io/README.md) · [해설](../../solutions/14-files-io/README.md) |
