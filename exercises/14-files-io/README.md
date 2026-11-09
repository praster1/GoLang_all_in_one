# 14장 연습문제: 파일과 스트림

선수 학습: [14. 파일과 스트림](../../book/14-files-io/README.md)

입출력 함수는 정상 문자열만 처리한다고 완성되지 않습니다.
입력의 끝, 빈 입력, 부분 데이터, 읽기·쓰기 실패와 파일 보존을 각각 확인합니다.
기본 문제 1~6을 먼저 풀고 선택 A에서 파일 목록을 다룹니다.

모든 명령은 **저장소 루트**에서 실행합니다.
기본 과정에는 외부 계정이나 실제 개인 파일이 필요하지 않습니다.
디스크를 쓰는 실습은 `os.MkdirTemp` 또는 테스트의 `t.TempDir` 아래에서 진행합니다.

## 개인 폴더 준비

| 문제 | 수준 | 개인 폴더 |
| --- | --- | --- |
| 1. 마지막 데이터 추적 | 확인 | `_practice/14-files-io/eof/` |
| 2. 내용 있는 레코드 세기 | 변형 | `_practice/14-files-io/count-records/` |
| 3. 버퍼에 남은 출력 | 오류 진단 | `_practice/14-files-io/flush/` |
| 4. 줄 번호 출력 | 변형 | `_practice/14-files-io/number-lines/` |
| 5. 기존 파일 보존 복사 | 응용 | `_practice/14-files-io/copy-file/` |
| 6. 제한 있는 읽기 | 응용 | `_practice/14-files-io/read-limit/` |
| 선택 A. 텍스트 파일 목록 | 선택 | `_practice/14-files-io/list-text/` |

자신의 터미널에 맞는 준비 블록 하나를 실행합니다.
기존 파일이 있으면 덮어쓰지 않으며 테스트 파일은 복사하지 않습니다.

Windows PowerShell:

```powershell
foreach ($practiceName in @("eof", "count-records", "flush", "number-lines", "copy-file", "read-limit", "list-text")) {
    $practiceDir = Join-Path "_practice/14-files-io" $practiceName
    New-Item -ItemType Directory -Force $practiceDir | Out-Null
    $practiceMain = Join-Path $practiceDir "main.go"
    if (Test-Path -LiteralPath $practiceMain) {
        Write-Host "$practiceMain 파일이 이미 있습니다. 기존 내용을 확인하세요."
    } else {
        Copy-Item examples/01-first-program/hello/main.go $practiceMain
    }
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
for practice_name in eof count-records flush number-lines copy-file read-limit list-text; do
    practice_dir="_practice/14-files-io/$practice_name"
    mkdir -p "$practice_dir"
    if [ -e "$practice_dir/main.go" ]; then
        printf '%s\n' "$practice_dir/main.go 파일이 이미 있습니다. 기존 내용을 확인하세요."
    else
        cp examples/01-first-program/hello/main.go "$practice_dir/main.go"
    fi
done
```

## 문제 1. 데이터와 EOF가 함께 오면? — 확인

다음 전체 코드는 컴파일되지만 마지막 데이터를 잃습니다.
실행하기 전에 출력과 그 이유를 예상하세요.

```go
package main

import (
	"fmt"
	"io"
)

type finalReader struct{}

func (finalReader) Read(p []byte) (int, error) {
	return copy(p, "Go"), io.EOF
}

func readBad(source io.Reader) string {
	buffer := make([]byte, 8)
	n, err := source.Read(buffer)
	if err != nil {
		return ""
	}
	return string(buffer[:n])
}

func main() {
	fmt.Printf("내용: %q\n", readBad(finalReader{}))
}
```

작업 파일: `_practice/14-files-io/eof/main.go`

```sh
go run ./_practice/14-files-io/eof
```

다음 세 질문에 답합니다.

1. `Read`가 반환한 `n`과 `err`는 각각 무엇인가요?
2. 오류가 있다는 이유만으로 바이트가 없다고 판단할 수 있나요?
3. 반환된 데이터와 종료 판단 중 어떤 것을 먼저 해야 하나요?

수정할 때는 이름을 `readGood`으로 바꾸고 `(string, error)`를 반환하세요.
읽은 데이터를 보관한 뒤 EOF이면 정상 결과로 돌려주고,
다른 오류라면 부분 데이터와 오류를 함께 반환하도록 작성합니다.
이번 문제의 Reader는 한 번의 읽기로 끝나므로 반복 일반화는 선택입니다.

<details>
<summary>힌트 1: 유효한 바이트의 범위를 먼저 적기</summary>

`buffer[:n]`에는 `Go`가 들어 있습니다.
EOF는 앞으로 더 읽을 내용이 없다는 신호이지 이번 바이트를 버리라는 신호가 아닙니다.

</details>

<details>
<summary>힌트 2: 오류의 종류를 구분하기</summary>

`errors.Is(err, io.EOF)`인 경우와 다른 오류를 구분합니다.
예제의 `readChunks`에서 데이터 처리와 오류 판단의 순서를 비교하세요.

</details>

완료 조건: 수정 전 데이터 손실을 설명하고, 수정 후 `"Go"`와 `nil`을 반환합니다.

## 문제 2. 주석과 빈 줄을 제외한 레코드 세기 — 변형

`func countRecords(source io.Reader) (int, error)`를 작성합니다.
입력은 UTF-8 텍스트이며 기본 줄 분할 규칙을 사용합니다.

| 줄의 상태 | 처리 |
| --- | --- |
| 앞뒤 공백을 지우면 비어 있음 | 제외 |
| 공백 제거 뒤 `#`로 시작 | 주석으로 제외 |
| 그 밖의 줄 | 하나 증가 |
| 읽기 또는 Scanner 오류 | `0`과 오류 반환 |

줄 중간의 `#`는 주석 시작으로 처리하지 않습니다.
Scanner는 시작 전에 1KiB 초기 버퍼와 128KiB 상한을 설정합니다.
128KiB의 긴 토큰은 오류로 검사하며, 파일 전체 크기와 토큰 크기를 구분합니다.

`main`에서는 `"# 메모\nGo\n\n테스트\n"`를 문자열 Reader로 전달합니다.

```text
레코드: 2
```

실행:

```sh
go run ./_practice/14-files-io/count-records
```

| 추가 입력 | 기대 개수 |
| --- | --- |
| `""` | 0 |
| `"\n # 주석\r\n\t\n"` | 0 |
| `"Go\r\n테스트"` | 2 |
| `"값 # 본문\n"` | 1 |

<details>
<summary>힌트 1: 읽기와 정책을 나누기</summary>

Scanner는 줄을 읽습니다.
어떤 줄을 레코드로 볼지는 `TrimSpace`와 `HasPrefix`를 사용해 결정합니다.

</details>

<details>
<summary>힌트 2: 반복이 끝난 이유 확인하기</summary>

반복 밖에서 `scanner.Err()`를 확인합니다.
오류가 있으면 앞서 센 값을 정상 결과로 사용하지 않는 계약입니다.

</details>

완료 조건: 마지막 줄바꿈이 없어도 처리하고, 읽기 실패를 성공한 0개와 구분합니다.

## 문제 3. 출력이 목적지에 없는 이유 — 오류 진단

`buffered-write` 예제를 개인 `flush/main.go`에 옮겨 실행합니다.
그다음 `writer.Flush()`를 호출하는 블록만 잠시 제거하세요.
소스는 [예제 파일](../../examples/14-files-io/buffered-write/main.go)에서 확인합니다.

프로그램이 성공적으로 실행되어도 `destination`에 데이터가 없을 수 있습니다.
어느 객체가 내용을 보관하고 있는지 그림 또는 표로 기록하세요.

수정 요구사항:

- 두 번의 버퍼 쓰기 결과를 확인합니다.
- 결과를 읽기 전에 Flush를 실행하고 오류를 반환합니다.
- `"Flush 전: 0"`과 `"Flush 후: 10"`을 설명합니다.
- Flush 실패가 `main`의 실패 처리까지 전달되게 합니다.

실행:

```sh
go run ./_practice/14-files-io/flush
```

<details>
<summary>힌트 1: 출력 대상을 두 개로 나누기</summary>

`writer`와 `destination`은 같은 객체가 아닙니다.
앞은 버퍼를 보관하고 뒤는 최종 데이터를 모으는 목적지입니다.

</details>

<details>
<summary>힌트 2: 성공 시점</summary>

버퍼에 들어간 시점의 성공과 아래 Writer에 전달된 시점의 성공은 다릅니다.
본문의 `failedWriter` 전체 코드로 Flush 오류도 확인할 수 있습니다.

</details>

완료 조건: 단순히 Flush 한 줄을 추가하는 것뿐 아니라 반환 오류의 처리 위치를 설명합니다.

## 문제 4. Reader를 읽어 줄 번호 붙이기 — 변형

`func writeNumbered(destination io.Writer, source io.Reader) error`를 작성합니다.

- 첫 줄의 번호는 1입니다.
- 각 줄을 `번호: 내용\n` 형태로 씁니다.
- 빈 줄도 번호를 받습니다.
- 마지막 줄에 줄바꿈이 없어도 처리합니다.
- CRLF 입력의 줄 구분자는 출력에 남기지 않으며 출력 줄 끝은 LF입니다.
- 읽기·쓰기 오류를 호출자에게 전달합니다.
- 함수가 전달받은 Reader나 Writer를 닫지 않습니다.

`main`에서 `"Go\n테스트\n끝"`을 입력으로 사용하고 `os.Stdout`에 씁니다.

```text
1: Go
2: 테스트
3: 끝
```

```sh
go run ./_practice/14-files-io/number-lines
```

테스트에서는 `bytes.Buffer`를 출력 대상으로 사용합니다.
공백이 보이지 않을 수 있으므로 다음은 `%q` 형식의 기대 문자열입니다.

| 입력 | 기대 출력 문자열 |
| --- | --- |
| `""` | `""` |
| `"Go"` | `"1: Go\n"` |
| `"Go\n"` | `"1: Go\n"` |
| `"Go\n\n끝"` | `"1: Go\n2: \n3: 끝\n"` |

<details>
<summary>힌트 1: 줄 번호를 늘리는 위치</summary>

`Scan()`이 참인 회차마다 번호를 증가시킵니다.
내용이 빈 문자열인지 여부로 번호 증가를 건너뛰지 않습니다.

</details>

<details>
<summary>힌트 2: 출력 함수도 오류를 반환합니다</summary>

`fmt.Fprintf(destination, ...)`의 오류를 확인합니다.
반복이 끝나면 별도로 `scanner.Err()`를 확인합니다.

</details>

완료 조건: 메모리 Writer로 정상 출력을 확인하고, 오류를 반환하는 Writer로 실패 전달을 확인합니다.

## 문제 5. 기존 파일을 보존하는 복사 — 응용

`func copyFile(sourcePath, destinationPath string) (int64, error)`를 작성합니다.

| 조건 | 계약 |
| --- | --- |
| 원본을 열 수 없음 | 대상 생성 전에 오류 반환 |
| 대상이 이미 있음 | 기존 내용 유지, 오류 반환 |
| 대상이 없음 | 새 파일에 복사 |
| 성공 | 복사한 바이트 수와 `nil` |
| 읽기·쓰기·닫기 실패 | 실패를 반환; 이미 생성된 부분 대상은 남을 수 있음 |

파일을 연 함수가 닫기 책임을 맡습니다.
이 문제는 임시 파일 교체나 실패 시 원상 복구까지 제공하지 않습니다.
호출자가 부분 대상을 성공 결과로 사용하지 않도록 실패 계약을 설명하세요.

`main`은 임시 디렉터리에 `source.txt`를 만들고 `"Go\n"`를 저장합니다.
새 `copy.txt`로 복사한 뒤 읽어 다음을 출력하고 임시 디렉터리를 정리합니다.

```text
복사한 바이트: 3
Go
```

```sh
go run ./_practice/14-files-io/copy-file
```

추가 확인:

1. 대상에 미리 `"keep"`을 저장하면 오류 뒤에도 그 내용이 그대로인가?
2. 원본 경로와 대상 경로가 같으면 원본이 유지되는가?
3. 없는 원본을 지정하면 대상이 만들어지지 않는가?
4. 빈 파일을 복사하면 바이트 수가 0이며 정상인가?

<details>
<summary>힌트 1: 덮어쓰지 않는 생성</summary>

`os.OpenFile`의 `O_CREATE`와 `O_EXCL`을 함께 사용합니다.
`os.Create`는 기존 내용을 보존하는 함수가 아닙니다.

</details>

<details>
<summary>힌트 2: 닫기 오류를 잃지 않기</summary>

11장의 이름 있는 오류 반환과 `errors.Join` 패턴을 사용할 수 있습니다.
두 파일의 닫기는 지연 등록의 역순으로 실행됩니다.

</details>

완료 조건: 성공 내용뿐 아니라 실패 뒤 기존 파일의 내용을 검사합니다.

## 문제 6. 허용 크기를 넘는 입력 거절하기 — 응용

`func readBounded(source io.Reader, maximum int64) ([]byte, error)`를 작성합니다.

- `maximum`은 0~1,048,576바이트 범위입니다.
- 범위 밖 제한값은 읽기 전에 오류로 처리합니다.
- 입력이 제한 이하라면 전체 바이트와 `nil`을 반환합니다.
- 입력이 제한을 넘으면 `nil`과 초과 오류를 반환합니다.
- 원본 Reader가 실패하면 `nil`과 원인을 감싼 오류를 반환합니다.
- Reader를 닫지 않습니다.

초과를 분류할 수 있도록 `errors.New("크기 제한 초과")`로 만든
패키지 오류 값 `errTooLarge`를 사용합니다.
호출자는 `errors.Is`로 이 오류를 확인합니다.
이 오류 변수는 실행 중 값을 바꾸지 않습니다.

기본 실행은 `"hello"`를 제한 5로 읽고,
`"hello!"`의 실패를 확인합니다.

```text
내용: hello
초과 확인: true
```

```sh
go run ./_practice/14-files-io/read-limit
```

| 입력 | 제한 | 결과 |
| --- | --- | --- |
| `""` | 0 | 정상 빈 입력 |
| `"x"` | 0 | 초과 |
| `"abcd"` | 5 | 정상 |
| `"hello"` | 5 | 정상 |
| `"hello!"` | 5 | 초과 |
| `"한"` | 3 | 정상 |
| `"한"` | 2 | 초과 |

<details>
<summary>힌트 1: 제한만큼만 읽으면 부족한 정보</summary>

5바이트만 받으면 원본이 정확히 5바이트인지 더 긴지 구분할 수 없습니다.
허용량보다 한 바이트 더 관찰하면 초과 여부를 판단할 수 있습니다.

</details>

<details>
<summary>힌트 2: 제한값 검증의 순서</summary>

`maximum+1`을 계산하기 전에 범위를 검사합니다.
`io.LimitReader`와 `io.ReadAll`을 연결하고 결과 길이를 비교하세요.

</details>

완료 조건: 정확히 제한과 같은 입력을 허용하고, 읽기 오류와 크기 초과를 분류합니다.

## 선택 A. 텍스트 파일 이름 목록

`func listText(directory string) ([]string, error)`를 작성하세요.
바로 아래 항목 중 디렉터리를 제외하고 확장자가 정확히 `.txt`인 이름을 반환합니다.
하위 디렉터리는 탐색하지 않습니다.
실습 입력은 일반 파일과 디렉터리만 포함하며 심볼릭 링크는 만들지 않습니다.

임시 디렉터리에 `z.txt`, `a.txt`, `memo.md`와 디렉터리 `folder.txt`를 준비합니다.
정상 반환값은 `[]string{"a.txt", "z.txt"}`입니다.

```sh
go run ./_practice/14-files-io/list-text
```

어느 함수가 순서를 보장하는지 공식 문서에서 확인합니다.
없는 디렉터리의 오류를 빈 목록과 구분하세요.

<details>
<summary>힌트</summary>

`os.ReadDir`의 정렬된 결과에서 `IsDir`와 `filepath.Ext`를 확인합니다.
전체 경로가 아니라 항목의 `Name()`을 반환합니다.

</details>

## 확인과 정리

```sh
go test -count=1 ./examples/14-files-io/... ./solutions/14-files-io/...
```

공식 테스트는 참고 코드의 계약을 확인합니다.
개인 코드는 자신의 폴더를 지정해 실행하고 추가 입력을 직접 비교하세요.
개인 코드의 포맷도 `go fmt ./_practice/14-files-io/number-lines`처럼 폴더를 명시합니다.

- [ ] EOF와 이번 바이트를 분리해 설명했다.
- [ ] 읽기 실패와 정상 빈 입력을 구분했다.
- [ ] Flush와 닫기의 오류를 전달했다.
- [ ] 기존 파일을 보존하는 실패 경로를 확인했다.
- [ ] 제한 크기와 같은 입력·초과 입력·한글을 확인했다.
- [ ] 작업 후 자신이 만든 임시 디렉터리만 정리했다.

[해설](../../solutions/14-files-io/README.md) · [본문](../../book/14-files-io/README.md) · [연습문제 목록](../README.md)
