# 14장 해설: 파일과 스트림

문제: [14장 연습문제](../../exercises/14-files-io/README.md)

이 장의 정답은 출력 문자열만으로 평가하지 않습니다.
누가 Reader와 Writer를 소유하는지, 실패 때 어떤 결과를 반환하는지,
기존 파일이 그대로 남는지도 계약에 포함됩니다.

모든 명령은 저장소 루트에서 실행합니다.
참고 코드의 테스트는 개인 폴더를 대신 검사하지 않습니다.
개인 풀이를 먼저 확인한 뒤 흐름과 경계 입력을 비교하세요.

## 문제 1. 마지막 데이터보다 먼저 오류를 검사한 경우

원래 프로그램의 출력은 다음과 같습니다.

```text
내용: ""
```

`Read`는 `n=2`, `err=io.EOF`를 함께 반환했습니다.
그러나 `readBad`는 오류를 먼저 보아 빈 문자열을 반환하므로,
이미 받은 `Go`를 사용하지 않습니다.

### 수정한 전체 코드

작업 파일: `_practice/14-files-io/eof/main.go`

```go
package main

import (
	"errors"
	"fmt"
	"io"
)

type finalReader struct{}

func (finalReader) Read(p []byte) (int, error) {
	return copy(p, "Go"), io.EOF
}

func readGood(source io.Reader) (string, error) {
	buffer := make([]byte, 8)
	n, err := source.Read(buffer)
	text := string(buffer[:n])
	if errors.Is(err, io.EOF) {
		return text, nil
	}
	return text, err
}

func main() {
	text, err := readGood(finalReader{})
	fmt.Printf("내용: %q\n", text)
	fmt.Println("오류 없음:", err == nil)
}
```

```sh
go run ./_practice/14-files-io/eof
```

예상 출력:

```text
내용: "Go"
오류 없음: true
```

이 함수는 문제에서 정한 **한 번 읽기**를 다룹니다.
임의 길이 입력을 끝까지 읽으려면 본문의 `readChunks`처럼 반복하거나 `io.ReadAll`을 사용합니다.
작은 Reader의 결과를 일반적인 전체 읽기 함수로 오해하지 않습니다.

| 순서 | 원래 코드 | 수정 코드 |
| --- | --- | --- |
| `Read` 반환 | 2바이트와 EOF | 2바이트와 EOF |
| 데이터 보관 | 하지 않음 | `buffer[:2]` 보관 |
| 종료 판단 | 모든 오류를 빈 결과로 처리 | EOF는 정상 끝으로 처리 |
| 반환 | 빈 문자열 | `Go`, `nil` |

다른 오류가 데이터와 함께 오면 수정 함수는 부분 문자열과 오류를 반환합니다.
호출자는 오류가 있는 결과를 완전한 문서로 취급하면 안 됩니다.
이 순서의 근거는 [Reader 계약](https://pkg.go.dev/io#Reader)입니다.

## 문제 2. 주석과 빈 줄을 제외한 레코드

참고 코드: [count-records/main.go](count-records/main.go) · [테스트](count-records/main_test.go)

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func countRecords(source io.Reader) (int, error) {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("레코드 읽기: %w", err)
	}
	return count, nil
}

func main() {
	count, err := countRecords(strings.NewReader("# 메모\nGo\n\n테스트\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("레코드:", count)
}
```

실행과 검사:

```sh
go run ./solutions/14-files-io/count-records
go test ./solutions/14-files-io/count-records
```

예상 출력:

```text
레코드: 2
```

### 읽기 규칙과 내용 규칙 분리

Scanner는 줄 경계를 찾습니다.
공백 제거와 주석 판단은 이 프로그램의 입력 정책입니다.
같은 Scanner를 사용해도 어떤 줄을 셀지는 프로그램마다 다를 수 있습니다.

| 입력 토큰 | `TrimSpace` 뒤 | 판단 |
| --- | --- | --- |
| `# 메모` | `# 메모` | 주석 |
| `Go` | `Go` | 레코드 |
| 빈 줄 | 빈 문자열 | 제외 |
| `테스트` | `테스트` | 레코드 |
| `값 # 본문` | `값 # 본문` | 시작이 `#`가 아니므로 레코드 |

주석 문자를 단순히 포함하는지 검사하면 마지막 행을 잘못 제외합니다.
여기서는 `HasPrefix`가 계약에 맞습니다.

### 오류가 있으면 개수를 쓰지 않기

정상 끝에서는 `scanner.Err()`가 `nil`입니다.
읽기에 실패하거나 토큰이 너무 길면 오류를 반환합니다.
참고 함수는 오류가 발생한 입력의 부분 개수를 사용하지 않도록 `0, err`를 반환합니다.

따라서 결과가 0이라는 사실만으로 빈 문서라고 판단할 수 없습니다.
호출자는 먼저 오류를 확인해야 합니다.
이는 앞 장의 유효한 0과 실패를 구분하는 문제와 같은 원리입니다.

테스트는 빈 입력, CRLF, 마지막 줄바꿈 없음, 주석 위치, 읽기 실패를 확인합니다.
Scanner의 동작은 [ScanLines](https://pkg.go.dev/bufio#ScanLines)와
[Scanner.Err](https://pkg.go.dev/bufio#Scanner.Err)를 참고합니다.

## 문제 3. 버퍼의 출력이 사라진 것처럼 보이는 이유

Flush를 제거한 코드는 데이터를 `bufio.Writer`의 메모리에 남겨 둡니다.
아래 목적지의 길이를 바로 읽으면 0이 나올 수 있습니다.
이번 작은 입력에서는 버퍼가 가득 차지 않으므로 직접 전달도 발생하지 않습니다.

수정한 전체 코드는 본문 예제와 같습니다.

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

```sh
go run ./_practice/14-files-io/flush
```

예상 출력:

```text
Flush 전: 0
Flush 후: 10
Go
buffer
```

| 책임 | 코드 위치 |
| --- | --- |
| 바이트를 버퍼에 넣음 | `fmt.Fprintln(writer, ...)` |
| 쓰기 호출 자체의 실패 확인 | 각 호출 뒤 오류 검사 |
| 남은 버퍼를 목적지로 전달 | `writer.Flush()` |
| 최종 내용 확인 | `destination.String()` |

버퍼링을 제거해 바로 쓰는 것도 다른 설계로는 가능합니다.
그러나 이 문제의 학습 목표는 버퍼와 목적지 사이의 시점을 이해하는 것입니다.
Flush만 호출하고 오류를 버리면 실패를 성공처럼 처리할 수 있습니다.

파일을 아래 Writer로 사용했다면 Flush 다음에 파일을 닫아야 합니다.
Flush가 파일을 닫아 주지는 않습니다.
[Writer.Flush](https://pkg.go.dev/bufio#Writer.Flush)의 역할을 다시 확인합니다.

## 문제 4. 줄 번호를 출력하는 함수

참고 코드: [number-lines/main.go](number-lines/main.go) · [테스트](number-lines/main_test.go)

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func writeNumbered(destination io.Writer, source io.Reader) error {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	number := 0
	for scanner.Scan() {
		number++
		if _, err := fmt.Fprintf(destination, "%d: %s\n", number, scanner.Text()); err != nil {
			return fmt.Errorf("줄 %d 쓰기: %w", number, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("줄 읽기: %w", err)
	}
	return nil
}

func main() {
	if err := writeNumbered(os.Stdout, strings.NewReader("Go\n테스트\n끝")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./solutions/14-files-io/number-lines
go test ./solutions/14-files-io/number-lines
```

예상 출력:

```text
1: Go
2: 테스트
3: 끝
```

### 빈 줄도 하나의 토큰

번호를 늘리는 조건은 내용의 길이가 아니라 `Scan()`의 성공입니다.
따라서 빈 줄도 번호를 받습니다.
출력 서식에는 콜론 뒤 공백이 있으며 빈 내용일 때도 유지됩니다.

| 입력 | 정확한 출력 문자열 |
| --- | --- |
| `""` | `""` |
| `"Go"` | `"1: Go\n"` |
| `"Go\n"` | `"1: Go\n"` |
| `"Go\n\n끝"` | `"1: Go\n2: \n3: 끝\n"` |
| `"가\r\n나\r\n"` | `"1: 가\n2: 나\n"` |

마지막에 LF가 있는 문서와 없는 문서는 여기서 같은 결과가 될 수 있습니다.
이 함수는 원본의 줄 구분자 바이트를 그대로 복제하는 함수가 아닙니다.
줄을 읽고 정해진 형식으로 다시 출력합니다.

### 읽기와 쓰기 모두 실패할 수 있다

`fmt.Fprintf`가 실패하면 해당 줄의 번호를 오류에 덧붙입니다.
원래 오류는 `%w`로 유지하므로 호출자가 `errors.Is`로 확인할 수 있습니다.

출력 실패는 이미 쓴 줄을 자동으로 지우지 않습니다.
일부 결과를 받을 수 있다는 점을 Writer를 전달하는 호출자가 알아야 합니다.
완전한 결과가 필요하면 임시 저장과 검증을 결합하는 설계를 추가합니다.

이 함수는 Reader와 Writer를 닫지 않습니다.
파일을 연 적이 없고, 호출자가 그 객체를 계속 사용하려 할 수도 있기 때문입니다.
본문의 `writeReport`가 파일 수명을 맡는 구조와 연결해 보세요.

## 문제 5. 기존 파일을 보존하는 복사

참고 코드: [copy-file/main.go](copy-file/main.go) · [테스트](copy-file/main_test.go)

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyFile(sourcePath, destinationPath string) (written int64, err error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return 0, fmt.Errorf("원본 열기: %w", err)
	}
	defer func() {
		err = errors.Join(err, source.Close())
	}()

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return 0, fmt.Errorf("대상 만들기: %w", err)
	}
	defer func() {
		err = errors.Join(err, destination.Close())
	}()
	written, err = io.Copy(destination, source)
	return written, err
}

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-copy-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()
	source := filepath.Join(directory, "source.txt")
	destination := filepath.Join(directory, "copy.txt")
	if err := os.WriteFile(source, []byte("Go\n"), 0o600); err != nil {
		return err
	}
	written, err := copyFile(source, destination)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		return err
	}
	_, err = fmt.Printf("복사한 바이트: %d\n%s", written, data)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./solutions/14-files-io/copy-file
go test ./solutions/14-files-io/copy-file
```

예상 출력:

```text
복사한 바이트: 3
Go
```

### 대상 생성 플래그 읽기

| 플래그 | 이 문제에서의 의미 |
| --- | --- |
| `O_WRONLY` | 쓰기용으로 열기 |
| `O_CREATE` | 대상이 없으면 생성 |
| `O_EXCL` | 이미 있으면 실패 |

이 조합은 기존 내용을 지우고 복사하는 요구사항과 다릅니다.
원본과 대상 경로가 같아도 대상이 이미 존재하므로 실패하고 원본을 유지합니다.
[OpenFile](https://pkg.go.dev/os#OpenFile)과 [파일 열기 플래그](https://pkg.go.dev/os#pkg-constants)를 참고합니다.

### 원본을 먼저 여는 이유

원본을 열 수 없는데 대상을 먼저 만들면 빈 파일이 남을 수 있습니다.
이 구현은 원본 열기를 먼저 시도하여 그 실패에서는 대상을 만들지 않습니다.
실패 순서도 관찰 가능한 계약입니다.

두 파일을 성공적으로 열면 복사 뒤 지연 닫기가 역순으로 실행됩니다.
출력 파일이 먼저 닫히고 입력 파일이 뒤에 닫힙니다.
각 닫기 오류는 최종 반환 오류에 합칩니다.

### 이 함수가 보장하지 않는 것

복사 도중 실패하면 새로 만든 대상에 일부 바이트가 남을 수 있습니다.
복사를 취소했다고 기존 상태로 자동 복원되는 함수는 아닙니다.
문제는 기존 파일 보존과 자원 수명을 확인하는 데 범위를 둡니다.

| 상황 | 파일 상태 |
| --- | --- |
| 원본 없음 | 대상 생성하지 않음 |
| 대상 이미 있음 | 대상 내용 유지 |
| 정상 복사 | 새 대상에 전체 내용 |
| 복사 도중 실패 | 새 대상에 일부 내용이 남을 수 있음 |

테스트는 기존 대상의 바이트를 다시 읽어 보존 여부를 확인합니다.
오류가 났다는 사실만 확인하면 이미 잘못 덮어쓴 구현을 놓칠 수 있습니다.
성공 내용과 실패 뒤 상태를 함께 검사하는 이유입니다.

## 문제 6. 허용 크기보다 한 바이트 더 확인

참고 코드: [read-limit/main.go](read-limit/main.go) · [테스트](read-limit/main_test.go)

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var errTooLarge = errors.New("크기 제한 초과")

func readBounded(source io.Reader, maximum int64) ([]byte, error) {
	if maximum < 0 || maximum > 1024*1024 {
		return nil, fmt.Errorf("제한은 0부터 1048576바이트까지 지정하세요")
	}
	data, err := io.ReadAll(io.LimitReader(source, maximum+1))
	if err != nil {
		return nil, fmt.Errorf("읽기: %w", err)
	}
	if int64(len(data)) > maximum {
		return nil, errTooLarge
	}
	return data, nil
}

func main() {
	data, err := readBounded(strings.NewReader("hello"), 5)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("내용: %s\n", data)
	_, err = readBounded(strings.NewReader("hello!"), 5)
	fmt.Println("초과 확인:", errors.Is(err, errTooLarge))
}
```

```sh
go run ./solutions/14-files-io/read-limit
go test ./solutions/14-files-io/read-limit
```

예상 출력:

```text
내용: hello
초과 확인: true
```

### 상한 검사 순서

먼저 제한값 자체가 0~1MiB인지 확인합니다.
그다음 `maximum+1`바이트까지 읽고 실제 받은 길이를 비교합니다.
이 순서라면 한 바이트를 더하는 계산의 범위도 명확합니다.

`LimitReader`는 경계에서 정상 EOF를 만듭니다.
초과 오류를 만드는 것은 그 위에 작성한 길이 검사입니다.
표준 함수와 프로그램 정책이 각각 맡는 역할을 나누어 읽습니다.

| 입력 | 최대값 | 최대값+1까지 읽은 길이 | 결과 |
| --- | --- | --- | --- |
| 빈 입력 | 0 | 0 | 정상 |
| `x` | 0 | 1 | 초과 |
| `hello` | 5 | 5 | 정상 |
| `hello!` | 5 | 6 | 초과 |
| `한` | 2 | 3 | 초과 |

### 원본 오류와 크기 초과 구분

`errTooLarge`는 이 함수의 정책 오류입니다.
원본 Reader가 반환한 오류는 `읽기: %w`로 문맥을 붙여 전달합니다.
두 실패를 같은 문자열로 뭉개지 않으므로 호출자는 원인을 분류할 수 있습니다.

실패할 때 반환 슬라이스는 `nil`입니다.
부분 데이터를 사용해야 하는 다른 작업이라면 별도의 계약이 필요합니다.
이 구현은 제한을 만족한 완전한 입력만 사용하는 목적에 맞췄습니다.

[LimitReader](https://pkg.go.dev/io#LimitReader)와
[ReadAll](https://pkg.go.dev/io#ReadAll)을 함께 읽어 보세요.
바이트 제한과 시간 제한은 별개이며, 느린 Reader의 취소는 17장에서 다룹니다.

## 선택 A. 텍스트 파일 목록

다음은 전체 참고 코드입니다.
기본 네 패키지와 별도로 개인 `list-text/main.go`에서 실행합니다.

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func listText(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".txt" {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-list-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()
	for _, name := range []string{"z.txt", "a.txt", "memo.md"} {
		if err := os.WriteFile(filepath.Join(directory, name), nil, 0o600); err != nil {
			return err
		}
	}
	if err := os.Mkdir(filepath.Join(directory, "folder.txt"), 0o700); err != nil {
		return err
	}
	names, err := listText(directory)
	if err != nil {
		return err
	}
	_, err = fmt.Println(names)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```sh
go run ./_practice/14-files-io/list-text
```

예상 출력:

```text
[a.txt z.txt]
```

`ReadDir`의 이름순 결과에서 조건에 맞는 항목만 유지하므로 반환 순서도 이름순입니다.
디렉터리 `folder.txt`는 확장자가 같아도 제외합니다.
파일명 `.TXT`는 문자열상 `.txt`와 다르므로 이 계약에서는 포함되지 않습니다.

이 선택 문제의 입력은 일반 파일과 디렉터리만 포함합니다.
심볼릭 링크를 실제 파일처럼 읽을지, 목록에 포함할지는 별도의 정책입니다.
하위 디렉터리를 재귀적으로 순회하는 구현도 이 계약에 포함되지 않습니다.

## 검증 결과를 해석하기

```sh
go test -count=1 ./solutions/14-files-io/...
```

| 패키지 | 핵심 검사 |
| --- | --- |
| `count-records` | 빈 줄·주석·CRLF·읽기 오류 |
| `number-lines` | 빈 입력·빈 줄·마지막 줄바꿈 없음·쓰기 오류 |
| `copy-file` | 복사 내용·기존 대상 보존·없는 원본·같은 경로 |
| `read-limit` | 정확한 제한·초과·한글 바이트·원본 오류 |

실제 파일 테스트는 `t.TempDir`를 사용합니다.
정상 출력 테스트와 함수의 경계 테스트를 함께 두어
예상 출력만 맞춘 구현이 실패 처리를 빠뜨리지 않게 합니다.

개인 풀이에서는 반환 오류를 먼저 확인하고 성공 결과를 비교합니다.
오류가 있을 때 파일이나 출력에 무엇이 남을 수 있는지도 기록하세요.
부분 결과를 허용하는 함수와 버리는 함수를 섞어 읽지 않습니다.

[본문](../../book/14-files-io/README.md) · [문제](../../exercises/14-files-io/README.md) · [해설 목록](../README.md)
