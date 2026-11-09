# Golang All In One

**Go를 처음 실행하는 순간부터, 테스트 가능한 CLI와 HTTP 서비스를 만드는 과정까지 배우는 한국어 교재입니다.**

설명을 읽고, 작은 예제를 실행하고, 결과를 바꿔 보고, 연습문제로 확인합니다. 문법의 사용법과 함께 왜 그렇게 작성하는지 설명합니다. 기본 과정은 Go 표준 라이브러리를 중심으로 구성하며, 데이터 처리와 백엔드 개발로 확장합니다.

> **00–31장 전체**의 본문·실습·해설과 실행 코드를 제공합니다. 기초 문법, 패키지·테스트, 입출력·동시성, HTTP·SQL, 품질·배포를 거쳐 세 가지 종합 프로그램을 완성합니다. [전체 목차](book/README.md)에서 순서대로 읽거나, [검증 기록](docs/VALIDATION.md)에서 실행 환경과 확인 범위를 볼 수 있습니다.

## 어디에서 시작할까요?

| 목적 | 읽을 문서 |
| --- | --- |
| Go를 설치하고 첫 예제를 실행하고 싶다 | [00장 · 개발 환경 설정](book/00-setup/README.md) |
| 첫 프로그램의 구조를 이해하고 싶다 | [01장 · 첫 Go 프로그램](book/01-first-program/README.md) |
| 값을 저장하고 계산하는 방법을 배우고 싶다 | [02장 · 변수·상수·자료형·연산자](book/02-variables-types/README.md) |
| 조건에 따라 선택하고 같은 작업을 반복하고 싶다 | [03장 · 조건문과 반복문](book/03-control-flow/README.md) |
| 함수를 만들고 값과 이름의 사용 범위를 이해하고 싶다 | [04장 · 함수와 스코프](book/04-functions-scope/README.md) |
| CSV 데이터를 읽고 집계하는 완성 도구를 보고 싶다 | [29장 · CSV 분석 CLI](book/29-csv-cli/README.md) |
| 메모리·SQL 저장소를 갖춘 API를 만들고 싶다 | [30장 · 할 일 HTTP API](book/30-todo-api/README.md) |
| 작업 수를 제한하고 취소하는 배치를 만들고 싶다 | [31장 · 동시 배치 처리](book/31-concurrent-batch/README.md) |
| 전체 학습 순서와 작성 상태를 보고 싶다 | [32개 장의 목차](book/README.md) |
| 예제를 찾아 실행하고 싶다 | [예제 안내](examples/README.md) |
| 직접 문제를 풀고 싶다 | [연습문제 안내](exercises/README.md) |
| 교재를 함께 작성하거나 AI에게 작업을 맡긴다 | [프로젝트 방향](docs/PROJECT_GUIDE.md) · [집필 기준](docs/AUTHORING.md) · [AGENTS.md](AGENTS.md) |

## 빠른 시작

Git과 Go를 설치한 뒤 다음 명령을 실행합니다. 아래 명령은 Windows PowerShell, macOS 터미널, Linux/WSL에서 같습니다.

```sh
git clone https://github.com/praster1/GoLang_all_in_one.git
cd GoLang_all_in_one
go version
go run ./examples/01-first-program/hello
```

예상 출력:

```text
안녕하세요, Go!
```

설치 방법과 PATH 문제 해결은 [00장](book/00-setup/README.md)에 있습니다. 모든 명령은 별도 안내가 없다면 **저장소 최상위 디렉터리**에서 실행합니다.

### Go 버전

- 모듈 기준: **Go 1.27.0 이상**, Go 1.27 계열을 사용합니다.
- 시작 시 권장 설치 버전: **Go 1.27.3**. 2026-11-09에 [공식 다운로드](https://go.dev/dl/)와 [릴리스 이력](https://go.dev/doc/devel/release)을 확인했습니다.
- CI는 `1.27.x`의 최신 패치를 사용합니다. 버전을 올릴 때는 `go.mod`, CI와 설치 안내를 함께 점검합니다.
- 기본 문법·입출력·HTTP·동시성은 표준 라이브러리로 학습합니다. 24장과 30장의 실제 SQL 실습에는 `modernc.org/sqlite v1.62.0`을 사용하며 버전은 `go.mod`·`go.sum`에 고정합니다. 별도 DB 서버나 계정은 필요하지 않습니다. 처음 의존성을 내려받을 때는 네트워크가 필요합니다.

## 학습 방식

1. 각 장의 학습 목표와 사전 지식을 확인합니다.
2. 예제를 실행하기 전에 결과를 예상합니다.
3. 실제 출력과 설명을 비교하고 코드를 조금 바꿔 봅니다.
4. [연습문제](exercises/README.md)를 먼저 풀고, 막히면 힌트를 봅니다.
5. 마지막에 [해설](solutions/README.md)과 자신의 풀이를 비교합니다.

이미 다른 언어를 사용한다면 익숙한 문법 설명은 빠르게 읽어도 됩니다. 값 복사, 슬라이스, 인터페이스, 오류 처리, 고루틴과 취소는 Go의 동작을 직접 확인하며 학습합니다.

## 저장소 구성

| 경로 | 역할 |
| --- | --- |
| [`book/`](book/README.md) | 교재 본문과 전체 목차 |
| [`examples/`](examples/README.md) | 본문에서 설명하는 실행 가능한 예제 |
| [`exercises/`](exercises/README.md) | 먼저 풀어 볼 문제와 힌트 |
| [`solutions/`](solutions/README.md) | 문제 해설과 검증 가능한 모범 코드 |
| [`docs/`](docs/PROJECT_GUIDE.md) | 프로젝트 방향, 집필 기준, 집필 순서 |
| [`templates/chapter.md`](templates/chapter.md) | 새 장을 작성할 때 사용할 양식 |
| [`tools/checkdocs/`](tools/checkdocs/README.md) | 문서의 로컬 파일 링크 검사 도구 |
| [`.github/`](.github/workflows/ci.yml) | 자동 검사, 오류 제보 양식, PR 양식 |

하나의 `go.mod`로 예제를 관리합니다. 독립 실행 프로그램은 각각의 디렉터리에 두므로, 다른 장의 `main` 함수와 충돌하지 않습니다.

## 코드와 문서 확인

```sh
go fmt ./...
go vet ./...
go test ./...
go build ./...
go run ./tools/checkdocs
```

Linux에서 C 컴파일러가 준비되어 있다면 데이터 레이스 검사도 실행합니다.

```sh
go test -race ./...
```

GitHub Actions는 Windows·Linux에서 포맷, 정적 검사, 테스트, 빌드, 로컬 문서 링크를 확인하고 Linux에서 레이스 검사를 실행합니다. 28장의 [산출물 워크플로](.github/workflows/build-artifacts.yml)는 Linux·Windows 실행 파일과 컨테이너의 대표 동작을 검사합니다. 검사 범위와 실제 결과는 [CI 설정](.github/workflows/ci.yml), [기여 안내](CONTRIBUTING.md), [검증 기록](docs/VALIDATION.md)을 참고하세요.

## 집필과 프로젝트 지침

- [AGENTS.md](AGENTS.md): 이 저장소에서 작업하는 AI 에이전트의 공통 지침
- [프로젝트 방향과 운영 기준](docs/PROJECT_GUIDE.md): 대상 독자, 범위, 구조와 완료 기준
- [교재 집필 기준](docs/AUTHORING.md): 설명, 예제, 연습문제와 출처 작성법
- [집필 로드맵](docs/ROADMAP.md): 완료한 집필 범위와 이후 개정 방향
- [ChatGPT 프로젝트용 지침](docs/CHATGPT_PROJECT_INSTRUCTIONS.md): 프로젝트 설정의 지침란에 사용할 문구
- [기여 안내](CONTRIBUTING.md): 변경 작성과 검증 절차

## 공식 참고 자료

[Go 문서](https://go.dev/doc/) · [A Tour of Go](https://go.dev/tour/) · [언어 명세](https://go.dev/ref/spec) · [표준 라이브러리](https://pkg.go.dev/std) · [Go 모듈](https://go.dev/ref/mod)

공식 자료와 이 교재의 설명이 다르다면, 해당 Go 버전의 명세와 문서를 확인하고 교재를 수정합니다.
