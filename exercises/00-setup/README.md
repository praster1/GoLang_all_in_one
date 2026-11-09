# 00장 실습: 개발 환경 확인과 문제 해결

선수 학습: [00. 개발 환경과 학습 준비](../../book/00-setup/README.md)

이 실습에서는 Go 문법을 새로 배우기보다 **어느 환경의 어느 폴더에서 어떤 명령을 실행하는지** 확인합니다. 실행 예제를 복사해 출력 문장 하나만 바꾸고, 잘못된 위치에서 실행했을 때 원인을 찾습니다.

준비물은 Go, Git, 편집기, 내려받은 이 저장소입니다. 설치가 끝나지 않았다면 먼저 00장 본문을 따라 준비하세요. 문제마다 별도로 안내한 경우를 제외하면 저장소 루트에서 실행합니다. 이 실습은 시스템 설정, 전역 Git 설정, 다른 프로젝트의 설정을 변경하지 않습니다.

## 실습 1. 지금 사용하는 환경 기록하기 — 확인

### 목표

Go와 Git을 실행할 수 있는지 확인하고, 현재 터미널의 환경과 저장소 위치를 설명합니다.

### 수행할 작업

먼저 저장소 루트로 이동합니다. 현재 폴더를 확인하는 명령은 운영체제의 터미널에 맞게 고릅니다.

Windows PowerShell:

```powershell
Get-Location
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
pwd
```

다음 명령은 PowerShell과 Bash/Zsh에서 같습니다.

```text
go version
git --version
go env GOVERSION GOOS GOARCH GOMOD
```

출력 내용을 자신의 학습 노트에 기록하고 다음 표를 채우세요. 경로는 자신의 환경에 맞는 값으로 기록하면 됩니다.

| 확인 항목 | 기록할 내용 |
| --- | --- |
| 터미널을 실행한 환경 | Windows 직접 실행 / WSL / macOS / Linux / 원격 서버 중 해당하는 환경 |
| 현재 작업 폴더 | `Get-Location` 또는 `pwd`로 확인한 위치 |
| Go 버전 | `go version`의 결과와 `GOVERSION` |
| Git 버전 | `git --version`의 결과 |
| 빌드 대상 운영체제 | `GOOS` |
| 빌드 대상 CPU 구조 | `GOARCH` |
| 현재 모듈 파일 | `GOMOD`가 가리키는 파일 |

`go env GOVERSION GOOS GOARCH GOMOD`는 요청한 항목을 순서대로 한 줄씩 출력합니다. 출력된 네 줄을 표의 항목과 대응해 보세요. `GOOS`와 `GOARCH`는 **빌드 대상**이며, 별도 설정이 없다면 보통 실행 환경을 따릅니다. 이 값만으로 로컬 Linux, WSL, 원격 Linux를 구별할 수는 없습니다. 터미널을 어떻게 열었는지와 편집기의 원격 접속 상태도 함께 확인합니다. [공식 Go 환경 변수 설명](https://pkg.go.dev/cmd/go#hdr-Environment_variables)을 참고하세요.

### 완료 조건

- [ ] Go와 Git의 버전 명령이 오류 없이 실행된다.
- [ ] 현재 저장소의 [go.mod](../../go.mod)가 요구하는 Go 버전을 충족한다.
- [ ] `GOMOD`가 지금 학습 중인 저장소의 `go.mod`를 가리킨다.
- [ ] 로컬·WSL·원격 중 실제 코드를 실행하는 환경을 설명할 수 있다.
- [ ] 운영체제나 사용자 폴더가 다르면 출력 경로도 달라진다는 점을 설명할 수 있다.

<details>
<summary>힌트 1: 실행 환경 확인</summary>

Windows 컴퓨터에서 VS Code를 사용하더라도 원격 터미널은 Linux 서버에서 실행될 수 있습니다. WSL 터미널을 열었는지, SSH로 다른 컴퓨터에 접속했는지, Windows PowerShell을 직접 열었는지부터 확인하세요.

</details>

<details>
<summary>힌트 2: 경로와 버전 확인</summary>

`GOMOD`의 마지막 파일 이름만 보지 말고, 그 앞의 폴더가 내려받은 저장소인지 확인하세요. Go 버전은 현재 저장소의 `go.mod`와 비교합니다. 설치한 도구와 모듈의 요구 버전에 따라 Go가 선택하는 도구 체인이 달라질 수 있습니다. [Go Toolchains](https://go.dev/doc/toolchain)에 선택 규칙이 설명되어 있습니다.

</details>

## 실습 2. 내 파일을 수정하고 실행 파일 만들기 — 변형

### 목표

교재 예제를 개인 실습 폴더로 복사하고, 출력 문장을 바꾸어 실행·포맷·빌드를 수행합니다.

### 1단계: 개인 실습 파일 준비

저장소 루트에서 자신의 터미널에 맞는 명령만 실행합니다. 아래 명령은 `main.go`가 이미 있으면 복사하지 않습니다. 기존 파일이 있다는 안내가 나오면 파일을 열어 이전 작업을 확인하고 이어서 진행하세요. 이전 풀이를 보관하려면 편집기에서 다른 폴더로 복사해 두세요.

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force _practice/00-setup | Out-Null
if (Test-Path _practice/00-setup/main.go) {
    Write-Host "main.go가 이미 있습니다. 기존 내용을 확인하세요."
} else {
    Copy-Item examples/01-first-program/hello/main.go _practice/00-setup/main.go
}
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
mkdir -p _practice/00-setup
if [ -e _practice/00-setup/main.go ]; then
    printf '%s\n' 'main.go가 이미 있습니다. 기존 내용을 확인하세요.'
else
    cp examples/01-first-program/hello/main.go _practice/00-setup/main.go
fi
```

이 명령은 폴더를 만들고 파일을 복사하기 위한 셸 명령입니다. 여기서 Go의 조건문 문법을 배우는 것은 아닙니다. 개인 실습 폴더의 기본 규칙은 [연습문제 안내](../README.md)에 있습니다.

### 2단계: 출력 문자열만 변경

편집기에서 `_practice/00-setup/main.go`를 엽니다. 원본의 `안녕하세요, Go!` 문자열만 `개발 환경 준비 완료!`로 바꾸고 저장하세요. `examples/`의 교재 원본은 수정하지 않습니다. 아직 `package`, `import`, `func`의 의미를 모두 이해하지 않아도 됩니다. 다음 장에서 설명합니다.

저장소 루트에서 실행합니다.

```text
go run ./_practice/00-setup
```

정확한 예상 출력:

```text
개발 환경 준비 완료!
```

글자와 공백, 느낌표가 일치하는지 확인합니다. 출력은 한 줄이며 끝에는 줄바꿈이 있습니다.

### 3단계: 형식 정리 후 다시 실행

```text
go fmt ./_practice/00-setup
go run ./_practice/00-setup
```

`go fmt`는 코드 형식을 정리하며 파일을 수정할 수 있습니다. 이미 형식이 맞으면 별도 출력이 없을 수 있습니다. 형식을 정리한 뒤에도 위의 예상 출력은 같아야 합니다.

### 4단계: 실행 파일 만들기

다음의 `bin/setup-check` 또는 `bin/setup-check.exe`는 이 실습의 빌드 결과물입니다. 다시 빌드하면 이 실습에서 만든 이전 실행 파일을 갱신합니다. 같은 경로에 다른 용도의 파일을 두었다면 먼저 별도 위치에 보관하거나 이 실습의 출력 이름을 바꾸세요.

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o bin/setup-check.exe ./_practice/00-setup
.\bin\setup-check.exe
```

macOS·Linux·WSL의 Bash/Zsh:

```bash
mkdir -p bin
go build -o bin/setup-check ./_practice/00-setup
./bin/setup-check
```

빌드가 성공했는지 먼저 확인한 뒤 실행 파일을 실행하세요. 빌드가 실패했는데도 이전에 만든 파일을 실행하면 수정 전 결과가 나올 수 있습니다. 빌드한 실행 파일에서도 `개발 환경 준비 완료!`가 출력되어야 합니다.

### 완료 조건

- [ ] `_practice/00-setup/main.go`에서만 출력 문자열을 바꿨다.
- [ ] `go run`으로 지정한 한 줄을 출력했다.
- [ ] 해당 개인 패키지에 `go fmt`를 적용했다.
- [ ] 실행 파일을 빌드하고 직접 실행해 같은 결과를 확인했다.
- [ ] 실행 과정에서 새로운 `go.mod`나 `go.work`를 만들지 않았다.

<details>
<summary>힌트 1: 수정이 반영되지 않을 때</summary>

열어 둔 파일이 교재 원본인지 개인 복사본인지 확인하세요. 파일을 저장한 뒤 `go run`에 전달한 경로가 그 파일의 폴더를 가리키는지 확인합니다.

</details>

<details>
<summary>힌트 2: 전체 포맷 명령과 개인 폴더</summary>

이 저장소의 개인 연습 폴더는 이름이 `_`로 시작합니다. Go의 `./...` 패턴에서 제외되므로 `go fmt ./...`만 실행하면 개인 코드의 형식을 정리하지 않습니다. 문제에서 안내한 것처럼 개인 패키지의 경로를 명시하세요. [패키지 탐색 규칙](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)을 참고합니다.

</details>

## 실습 3. 실행 경로 오류 설명하기 — 응용

### 목표

실행 명령의 기준 폴더와 대상 패키지를 구분하고, 오류를 설치 문제로 단정하기 전에 경로를 확인합니다.

### 수행할 작업

1. 저장소 루트에서 실행 중인지 현재 폴더를 확인합니다.
2. 아래의 **의도적으로 실패하는 명령**을 한 번 실행하고 오류 메시지를 읽습니다.
3. 루트에 있는 `go.mod`와 예제에 있는 `main.go`의 역할을 구분해 원인을 한 문장으로 설명합니다.
4. 현재 폴더를 유지한 채 실행 대상 경로를 고쳐 교재 원본에서 `안녕하세요, Go!`를 출력합니다.
5. 개인 실습 파일을 실행해 `개발 환경 준비 완료!`가 나오는지도 확인합니다.

의도적으로 실패하는 예 — 현재 저장소 루트에는 실행할 Go 소스 파일이 없습니다.

```text
go run .
```

오류 메시지의 정확한 문구와 표시 경로는 Go 버전과 환경에 따라 달라질 수 있습니다. 오류가 발생했다는 이유만으로 Go를 다시 설치하거나 전역 설정을 바꾸지 않습니다. `.`이 가리키는 위치와 그 안의 Go 파일부터 확인하세요.

### 완료 조건

- [ ] 이 실패가 Go 설치 실패를 뜻하지 않는 이유를 설명했다.
- [ ] 저장소 루트에서 교재 원본을 정확한 패키지 경로로 실행했다.
- [ ] 저장소 루트에서 개인 실습 파일도 정확한 패키지 경로로 실행했다.
- [ ] 모듈 루트와 실행할 패키지의 디렉터리를 구분했다.

<details>
<summary>힌트 1: 점 하나의 의미</summary>

명령의 `.`은 현재 폴더를 뜻합니다. 이 저장소 루트에는 모듈 설정이 있지만, 프로그램 소스는 예제의 하위 폴더에 있습니다.

</details>

<details>
<summary>힌트 2: 실행 대상 찾기</summary>

실행 대상은 `main.go`를 둔 폴더입니다. 개인 실습 파일의 위치와 교재 원본의 위치를 각각 확인하고, 저장소 루트에서 시작하는 상대 경로를 `go run`에 전달해 보세요.

</details>

## 실습 후 자기 점검

아래 질문에 자신의 말로 답할 수 있으면 00장의 환경 실습을 마친 것입니다.

1. 지금의 터미널은 어느 운영체제와 컴퓨터에서 실행 중인가?
2. 내 명령의 현재 작업 폴더는 어디인가?
3. `go.mod`가 있는 폴더와 `main.go`가 있는 폴더가 달라도 되는가?
4. `go run`으로 실행하는 것과 빌드한 실행 파일을 직접 실행하는 것은 어떻게 다른가?
5. 개인 연습 폴더에는 왜 검사 경로를 명시해야 하는가?

## 해설과 다음 학습

먼저 각 실습을 수행하고 힌트를 확인한 뒤 [00장 실습 해설](../../solutions/00-setup/README.md)에서 결과와 설명을 비교합니다. 이후 [01. 첫 Go 프로그램](../../book/01-first-program/README.md)에서 방금 실행한 코드의 구조를 배웁니다.

| 본문 | 문제 목록 | 해설 |
| --- | --- | --- |
| [00장](../../book/00-setup/README.md) | [전체 연습문제](../README.md) | [00장 해설](../../solutions/00-setup/README.md) |
