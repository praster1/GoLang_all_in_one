# 28장 해설 — 빌드와 배포 자동화

[문제](../../exercises/28-build-deploy/README.md)를 먼저 풀고 확인합니다.
여기서는 파일 이름·내용·설정·실행 계약을 각각 검증합니다.
실제 클라우드 계정이나 운영 자격 증명은 필요하지 않습니다.

## 실행 안내

저장소 루트에서 실행합니다.

```sh
go test -count=1 ./solutions/28-build-deploy/...
```

| 문제 | 구현 | 핵심 확인 |
| --- | --- | --- |
| 3 | [metadata](metadata/main.go) | 허용 라벨·대상·확장자 |
| 4 | [manifest](manifest/main.go) | 파일명 순서·크기·내용 해시·오류 |
| 5 | [embedded-config](embedded-config/main.go) | 포함된 자원과 JSON 계약 |
| 6 | [health-contract](health-contract/main.go) | 상태·미디어 타입·본문·메서드 |

## 문제 1. 검증 범위 분리

| 상황 | 확인한 것 | 추가 확인이 필요한 것 |
| --- | --- | --- |
| Windows 대상 빌드 | 해당 대상 형식으로 컴파일 가능 | 실제 Windows 실행과 기능 |
| Linux 인사 실행 | 현재 환경에서 대표 정상 경로 동작 | 오류 경로·다른 OS·복잡한 입력 |
| 일반 테스트 통과 | 실행한 입력에서 약속한 동작 | 검사하지 않은 입력·운영 조건 |
| 이미지 빌드 | 레이어와 명령으로 이미지를 구성 가능 | 최종 ENTRYPOINT 실행·권한·필요 자원 |
| 체크섬 일치 | 비교한 파일 내용이 같은지에 대한 근거 | 기준 체크섬의 신뢰할 수 있는 출처 |
| 상태 코드 200 | 응답 상태가 약속한 번호 | 본문 계약·의존성 준비·실제 처리 능력 |

기록에서 “Windows 지원 확인”처럼 범위를 넓히기보다
“Linux에서 Windows amd64 대상 빌드 성공, 실제 실행 미확인”처럼 적으면
다음 검증에서 무엇이 필요한지 알 수 있습니다.

## 문제 2. 환경 변수 수명

Bash의 `GOOS=... GOARCH=... go build ...`는 해당 명령의 환경으로 사용됩니다.
PowerShell에서는 환경 변수를 바꾸기 전에 값을 보관하고 `finally`에서 복원합니다.
중간에 빌드가 실패해도 복원 단계를 거치도록 만드는 것이 핵심입니다.

`go env -w`는 이후 Go 명령의 기본 설정에 영향을 줄 수 있으므로
이 일회성 실습에서는 사용하지 않습니다.
이름 끝에 `.exe`를 붙이는 것은 파일 이름 선택일 뿐 대상 형식을 바꾸지 않습니다.
대상은 GOOS·GOARCH와 관련 빌드 설정으로 결정됩니다.

CGO가 필요한 의존성은 C 도구와 대상 라이브러리를 요구할 수 있습니다.
이번 인사 CLI는 그 의존성이 없으므로 `CGO_ENABLED=0`으로 검증합니다.
다른 프로그램에 같은 조건을 적용하려면 실제 빌드와 실행을 확인해야 합니다.

## 문제 3. 산출물 이름

### 전체 구현

```go
package main

import (
	"errors"
	"fmt"
	"regexp"
)

type release struct{ Name, Version, OS, Arch string }

var validLabel = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

func artifactName(r release) (string, error) {
	if !validLabel.MatchString(r.Name) || !validLabel.MatchString(r.Version) {
		return "", errors.New("invalid release label")
	}
	if r.OS != "linux" && r.OS != "windows" && r.OS != "darwin" {
		return "", errors.New("unsupported OS")
	}
	if r.Arch != "amd64" && r.Arch != "arm64" {
		return "", errors.New("unsupported architecture")
	}
	suffix := ""
	if r.OS == "windows" {
		suffix = ".exe"
	}
	return fmt.Sprintf("%s-%s-%s-%s%s", r.Name, r.Version, r.OS, r.Arch, suffix), nil
}

func main() {
	name, err := artifactName(release{"hello", "1.0.0", "windows", "amd64"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(name)
}
```

```sh
go run ./solutions/28-build-deploy/metadata
```

```text
hello-1.0.0-windows-amd64.exe
```

정규식은 첫 문자 하나와 뒤의 최대 63자를 나누어 전체 1~64자를 제한합니다.
허용 문자에 경로 구분자와 공백이 없으므로 이름에 경로 구성 요소를 끼워 넣지 않습니다.
OS·아키텍처는 이 교재에서 지원하기로 정한 목록을 사용합니다.

실제 Go가 지원하는 모든 대상 목록은 더 넓을 수 있습니다.
따라서 이 함수의 오류는 언어가 그 OS를 지원하지 않는다는 뜻이 아니라,
이번 릴리스 규칙에서 허용하지 않는다는 뜻입니다.

### 테스트 전체

```go
package main

import (
	"strings"
	"testing"
)

func TestArtifactName(t *testing.T) {
	for _, tc := range []struct {
		r    release
		want string
		bad  bool
	}{
		{release{"hello", "1.0.0", "linux", "amd64"}, "hello-1.0.0-linux-amd64", false},
		{release{"hello", "dev-1", "windows", "arm64"}, "hello-dev-1-windows-arm64.exe", false},
		{release{"../bad", "1", "linux", "amd64"}, "", true},
		{release{"hello", "two words", "linux", "amd64"}, "", true},
		{release{"hello", "1", "other", "amd64"}, "", true},
		{release{"hello", "1", "linux", "386"}, "", true},
		{release{strings.Repeat("a", 64), "1", "linux", "amd64"}, strings.Repeat("a", 64) + "-1-linux-amd64", false},
		{release{strings.Repeat("a", 65), "1", "linux", "amd64"}, "", true},
	} {
		got, err := artifactName(tc.r)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("got=%q err=%v", got, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// hello-1.0.0-windows-amd64.exe
}
```

테스트는 정상 파일명과 허용하지 않는 라벨·대상을 비교합니다.
확장자가 붙었다는 이유만으로 파일의 실행 형식이 맞다고 판단하지 않습니다.
이 함수는 파일 내용까지 읽는 도구가 아닙니다.

## 문제 4. 체크섬 목록

### 전체 구현

```go
package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type entry struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

func fileDigest(path string) (sum string, size int64, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		if e := file.Close(); err == nil {
			err = e
		}
	}()
	hash := sha256.New()
	size, err = io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), size, nil
}

func manifest(dir string) ([]entry, error) {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := make([]entry, 0)
	for _, item := range items {
		if item.IsDir() || item.Name() == "manifest.json" {
			continue
		}
		info, err := item.Info()
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("not a regular file: %s", item.Name())
		}
		sum, size, err := fileDigest(filepath.Join(dir, item.Name()))
		if err != nil {
			return nil, err
		}
		result = append(result, entry{item.Name(), size, sum})
	}
	return result, nil
}

func run(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("manifest", flag.ContinueOnError)
	fs.SetOutput(errOut)
	dir := fs.String("dir", "", "artifact directory")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *dir == "" || fs.NArg() != 0 {
		fmt.Fprintln(errOut, "provide -dir and no positional arguments")
		return 2
	}
	entries, err := manifest(*dir)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		fmt.Fprintln(errOut, err)
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

### 처리 순서

1. 플래그를 파싱해 디렉터리 인자가 있는지 확인합니다.
2. `os.ReadDir`로 이름순 항목을 얻습니다.
3. 하위 디렉터리와 기존 manifest.json을 제외합니다.
4. 일반 파일인지 확인합니다.
5. 파일을 열어 스트림으로 해시와 실제 읽은 크기를 계산합니다.
6. 파일을 닫고 다음 항목으로 이동합니다.
7. 완성된 목록을 JSON으로 한 번 출력합니다.

`fileDigest`에 파일 열기·읽기·닫기를 묶었습니다.
반복문 바깥 함수가 끝날 때까지 열린 파일을 쌓지 않도록
파일 하나의 작업이 끝나면 해당 함수의 defer가 실행됩니다.
읽기 오류가 이미 있으면 그 오류를 유지하고,
읽기가 성공했는데 닫기가 실패하면 닫기 오류를 반환합니다.

### 테스트 전체

```go
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestManifest(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{"b.bin": "abc", "a.bin": "", "manifest.json": "old"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"-dir", dir}, &out, &errOut); code != 0 {
		t.Fatalf("code=%d: %s", code, errOut.String())
	}
	var got []entry
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "a.bin" || got[1].Name != "b.bin" || got[1].Size != 3 {
		t.Fatalf("got=%v", got)
	}
	if got[1].SHA256 != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("hash mismatch")
	}
}

func TestManifestErrors(t *testing.T) {
	var out, errOut bytes.Buffer
	if run(nil, &out, &errOut) != 2 {
		t.Fatal("missing argument")
	}
	if run([]string{"-dir", filepath.Join(t.TempDir(), "missing")}, &out, &errOut) != 1 {
		t.Fatal("missing directory")
	}
	empty, err := manifest(t.TempDir())
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatal("empty directory contract")
	}
}
```

테스트는 `t.TempDir()`에 자료를 만들고 외부 파일을 읽지 않습니다.
파일명 순서, 제외 규칙, 크기와 알려진 SHA-256을 확인합니다.
빈 디렉터리는 nil이 아닌 빈 목록으로 만들어 JSON `[]` 계약을 유지합니다.

### 직접 실행할 작은 자료 준비

저장소에는 줄바꿈 없는 abc 파일을 제공합니다.
기존 개인 파일이 있으면 보존하는 다음 명령으로 복사합니다.

PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/28-build-deploy/manifest-demo | Out-Null
if (-not (Test-Path _practice/28-build-deploy/manifest-demo/abc.txt)) {
    Copy-Item examples/28-build-deploy/checksum/testdata/abc.txt _practice/28-build-deploy/manifest-demo/abc.txt
}
```

Bash:

```sh
mkdir -p _practice/28-build-deploy/manifest-demo
if [ ! -e _practice/28-build-deploy/manifest-demo/abc.txt ]; then
    cp examples/28-build-deploy/checksum/testdata/abc.txt _practice/28-build-deploy/manifest-demo/abc.txt
fi
```

```sh
go run ./solutions/28-build-deploy/manifest -dir ./_practice/28-build-deploy/manifest-demo
```

새로 만든 폴더에 제공된 파일 하나만 있다면 다음과 같습니다.
다른 개인 파일이 있으면 목록 항목도 추가됩니다.

```json
[
  {
    "name": "abc.txt",
    "size": 3,
    "sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
  }
]
```

다른 내용이나 줄바꿈·BOM이 들어간 파일의 해시는 다릅니다.
텍스트 에디터에서 다시 저장한 파일과 처음 제공된 바이트가 같은지도 확인해야 합니다.

### 이 구현의 경계

이 도구는 빌드가 끝나 변경되지 않는 폴더를 대상으로 합니다.
파일 검사 후 다른 프로세스가 파일을 바꾸는 경쟁까지 막지 않습니다.
신뢰할 수 없는 사용자가 동시에 조작하는 파일 시스템의 보안 경계를 제공한다고 설명하지 않습니다.

또한 체크섬 목록은 파일 내용 확인 자료입니다.
공급자를 인증하려면 목록을 어디서 받았는지, 서명·빌드 출처를 어떻게 확인할지 별도 설계가 필요합니다.

## 문제 5. 임베드한 기본 설정

함께 빌드할 [defaults.json](embedded-config/defaults.json):

```json
{"name":"demo","port":8080}
```

### 전체 구현

```go
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

//go:embed defaults.json
var defaults string

type config struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func parseConfig(text string) (config, error) {
	var cfg config
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return config{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return config{}, errors.New("expected one JSON value")
	}
	if strings.TrimSpace(cfg.Name) == "" || cfg.Port < 1 || cfg.Port > 65535 {
		return config{}, errors.New("invalid config values")
	}
	return cfg, nil
}

func main() {
	cfg, err := parseConfig(defaults)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s:%d\n", cfg.Name, cfg.Port)
}
```

```sh
go run ./solutions/28-build-deploy/embedded-config
```

```text
demo:8080
```

`go:embed`가 기본 파일을 문자열 변수에 포함합니다.
파서에는 문자열을 전달하므로 테스트에서 다른 JSON도 쉽게 검사할 수 있습니다.
임베드된 기본값도 검증을 거칩니다.
빌드에 포함되어 있다는 이유만으로 내용이 올바르다고 가정하지 않습니다.

### 테스트 전체

```go
package main

import "testing"

func TestConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg, err := parseConfig(defaults)
	if err != nil || cfg.Name != "demo" || cfg.Port != 8080 {
		t.Fatal("embedded defaults")
	}
	for _, text := range []string{`{}`, `null`, `{"name":"x","port":0}`, `{"name":"x","port":80,"extra":1}`, `{"name":"x","port":80} {}`} {
		if _, err := parseConfig(text); err == nil {
			t.Fatalf("accepted %s", text)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// demo:8080
}
```

임시 작업 디렉터리로 이동해도 기본값을 읽습니다.
두 번째 Decode의 결과가 EOF인지 확인하여 추가 값·잘못된 꼬리를 거부합니다.
`null`은 디코딩 자체가 성공할 수 있지만 비어 있는 설정 값의 검증에서 거부됩니다.

이 코드는 `encoding/json` v1을 사용합니다.
알 수 없는 필드를 거부하는 옵션만으로 중복 키와 대소문자 매칭 정책이 달라지지는 않습니다.
그 차이는 [15장](../../book/15-data-time/README.md)과 공식 JSON 문서를 참고합니다.

### 실행 설정으로 확장하기

기본값 위에 실행 환경 값을 적용하는 설계도 가능합니다.
그 경우 적용 순서를 문서화하고 마지막 설정 전체를 검증합니다.
기본값·환경 변수·명령행 인자가 서로 충돌할 때의 우선순위를
코드 작성자의 기억에만 의존하지 않도록 테스트에 남깁니다.

## 문제 6. 상태 확인 계약

### 전체 구현

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
		}{"ok"}); err != nil {
			// 응답을 이미 쓰기 시작했으므로 두 번째 오류 응답을 덧붙이지 않습니다.
			return
		}
	})
	return mux
}

func main() {
	recorder := httptest.NewRecorder()
	newHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	fmt.Println(recorder.Code, strings.TrimSpace(recorder.Body.String()))
}
```

```sh
go run ./solutions/28-build-deploy/health-contract
```

```text
200 {"status":"ok"}
```

메서드를 포함한 ServeMux 패턴이 GET 경로를 등록합니다.
정상 응답에서는 헤더를 먼저 설정한 뒤 JSON을 씁니다.
기본 라우터가 다른 메서드와 없는 경로의 응답을 처리합니다.

### 테스트 전체

```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthContract(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		method, path string
		want         int
	}{{"GET", "/healthz", 200}, {"POST", "/healthz", 405}, {"GET", "/missing", 404}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.want {
			t.Fatalf("%s %s: %d", tc.method, tc.path, w.Code)
		}
		if tc.want == http.StatusOK {
			var body struct {
				Status string `json:"status"`
			}
			if w.Header().Get("Content-Type") != "application/json" {
				t.Fatal("media type")
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Status != "ok" {
				t.Fatal("body contract")
			}
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 200 {"status":"ok"}
}
```

200의 본문을 JSON으로 디코딩해 status 필드까지 확인합니다.
단순히 응답이 왔다는 사실을 넘어 약속한 형식이 맞는지 검사합니다.
POST와 없는 경로도 같은 핸들러로 실행합니다.

GET 패턴은 표준 ServeMux 규칙에 따라 HEAD 요청에도 대응합니다.
실제 HTTP 서버는 HEAD 본문 전송을 처리하며,
직접 핸들러를 부른 ResponseRecorder 관찰과 네트워크 응답은 구분해야 합니다.
자세한 차이는 [22장 HTTP 서버](../../book/22-http-server/README.md)에서 다룹니다.

이 상태 응답에는 DB 연결 같은 의존성이 없습니다.
실제 서비스의 준비 상태를 검사하려면 필요한 의존성·시간 제한·오류 정책을 더 정해야 합니다.
그 내용을 숨긴 채 모든 환경에 완성된 readiness 검사라고 소개하지 않습니다.

## 선택 A. 워크플로 읽기

현재 [build-artifacts.yml](../../.github/workflows/build-artifacts.yml)은
관련 예제·해설·모듈·워크플로 파일을 바꾼 PR과 수동 실행을 대상으로 합니다.
주요 흐름은 다음과 같습니다.

| 순서 | 역할 |
| --- | --- |
| 체크아웃·Go 준비 | 검사할 코드와 도구 선택 |
| 예제·해설 테스트 | 입력·오류 계약 확인 |
| 실행 파일 빌드 | Linux·Windows amd64 산출물 생성 |
| 실제 파일 실행 | 인사 출력과 버전 정보 확인 |
| manifest 생성 | 완성 파일의 이름·크기·해시 기록 |
| Artifact 업로드 | 7일 동안 실행 결과 보관 |
| 별도 Docker 작업 | 컨테이너를 빌드하고 인사 결과 확인 |

레지스트리 푸시나 운영 배포는 수행하지 않습니다.
추가 배포를 설계한다면 어느 검사가 끝나야 진행하는지,
어떤 동일 산출물을 전환할지, 실패 시 무엇을 복구할지 정해야 합니다.

### 개선 제안의 평가 기준

- 실제 문제를 해결하는 단계인가?
- 필요한 권한이 명확한가?
- 잘못된 산출물이 업로드되기 전에 실패할 수 있는가?
- 검사한 코드·파일·메타데이터를 추적할 수 있는가?
- 보관 기간과 다운로드 대상이 명확한가?
- 사용하지 않은 환경을 검증했다고 주장하지 않는가?

워크플로 줄 수를 늘리는 것이 목적은 아닙니다.
기존 단계로 해결되지 않는 위험과 검증 범위를 설명할 수 있어야 합니다.

## 검증 명령

```sh
go fmt ./solutions/28-build-deploy/...
go vet ./solutions/28-build-deploy/...
go test -count=1 ./solutions/28-build-deploy/...
go test -race -count=1 ./solutions/28-build-deploy/...
```

본문의 실제 빌드 명령과 각 환경의 실행 확인은 별도로 기록합니다.
이 장의 Go 구현은 임시 디렉터리와 로컬 요청으로 기본 검사가 가능합니다.

## 참고 자료

[embed](https://pkg.go.dev/embed) · [os.ReadDir](https://pkg.go.dev/os#ReadDir) · [crypto/sha256](https://pkg.go.dev/crypto/sha256) · [encoding/json](https://pkg.go.dev/encoding/json) · [http.ServeMux](https://pkg.go.dev/net/http#ServeMux)

[본문](../../book/28-build-deploy/README.md) · [연습문제](../../exercises/28-build-deploy/README.md) · [전체 해설](../README.md)
