# 연습문제 안내

연습문제는 설명을 읽은 뒤 스스로 코드를 작성하는 공간입니다. 교재의 실행 예제는 `examples/`, 참고 풀이는 `solutions/`에 있습니다.

## 학습 순서

1. 문제의 입력·출력·조건을 읽습니다.
2. 예상 결과를 적고 직접 작성합니다.
3. 실행 결과와 예상 결과를 비교합니다.
4. 막히면 문제의 힌트를 펼칩니다.
5. 참고 풀이를 읽고 자신의 코드와 차이를 설명합니다.

동일한 결과를 만드는 코드는 여러 가지일 수 있습니다. 참고 풀이를 외우기보다 요구 사항이 무엇인지, 자신의 코드가 어떤 입력까지 처리하는지 확인하세요.

## 개인 연습 폴더

개인 코드는 저장소 루트의 `/_practice/` 아래에 둡니다. 이 폴더는 Git에 포함하지 않습니다. 또한 Go의 `./...` 패키지 탐색에서 이름이 `_`로 시작하는 디렉터리는 제외되므로, 실습 도중 아직 완성되지 않은 코드가 교재 전체 검사에 섞이지 않습니다. 개인 코드의 경로를 직접 지정하면 실행하고 검사할 수 있습니다. 패키지 패턴의 규칙은 [Go 명령 문서](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)를 참고합니다.

### Windows PowerShell

```powershell
New-Item -ItemType Directory -Force _practice/01-first-program | Out-Null
Copy-Item examples/01-first-program/hello/main.go _practice/01-first-program/main.go
```

### macOS·Linux·WSL

```bash
mkdir -p _practice/01-first-program
cp examples/01-first-program/hello/main.go _practice/01-first-program/main.go
```

이미 연습한 `main.go`가 있다면 복사 명령을 다시 실행하기 전에 내용을 확인하세요. 새로 복사하면 이전 실습을 덮어쓸 수 있습니다.

복사한 파일을 편집한 뒤 저장소 루트에서 실행합니다.

```text
go run ./_practice/01-first-program
go fmt ./_practice/01-first-program
```

연습 폴더에도 새 `go.mod`는 만들지 않습니다. 루트의 모듈 설정을 사용합니다. 한 폴더에는 실행 프로그램 하나를 두고, 이전 답안을 보관하려면 다른 이름의 폴더를 만드세요. 같은 폴더에 `main` 함수가 있는 파일을 여러 개 쌓아 두면 중복 선언이 발생합니다.

## 현재 제공하는 문제

| 장 | 문제 | 참고 풀이 |
| --- | --- | --- |
| 00. 개발 환경과 학습 준비 | [환경 확인·실행·문제 해결 실습](00-setup/README.md) | [00장 실습 해설](../solutions/00-setup/README.md) |
| 01. 첫 Go 프로그램 | [변형·선택 응용 문제와 확인 질문 안내](01-first-program/README.md) | [풀이 안내](../solutions/README.md) |

새로운 장의 문제는 해당 장이 작성될 때 추가합니다.
