# 실행 예제

교재에서 설명하는 코드를 장별로 모았습니다. 모든 명령은 저장소 최상위 폴더에서 실행합니다. 각 예제 폴더는 독립적으로 실행할 수 있는 패키지이며, 루트의 `go.mod`를 함께 사용합니다.

| 장 | 예제 | 실행 명령 | 확인할 내용 |
| --- | --- | --- | --- |
| [01. 첫 Go 프로그램](../book/01-first-program/README.md) | [hello](01-first-program/hello/main.go) | `go run ./examples/01-first-program/hello` | 인사 한 줄 출력 |
| [01. 첫 Go 프로그램](../book/01-first-program/README.md) | [output](01-first-program/output/main.go) | `go run ./examples/01-first-program/output` | `Print`·`Println`·`Printf`의 공백과 줄바꿈 |
| [01. 첫 Go 프로그램](../book/01-first-program/README.md) | [multifile](01-first-program/multifile/main.go) · [message.go](01-first-program/multifile/message.go) | `go run ./examples/01-first-program/multifile` | 같은 패키지의 여러 파일을 함께 실행하는 선택 예제 |

`hello`의 출력 예시는 다음과 같습니다.

```text
안녕하세요, Go!
```

`output`은 문장 사이의 공백과 줄바꿈을 비교합니다. `multifile`은 다른 파일에 정의한 함수를 호출하므로 예제 폴더를 실행 대상으로 지정합니다. 각 예제의 예상 출력과 설명은 [01장](../book/01-first-program/README.md)에 있습니다. 패키지를 지정하는 실행과 파일을 직접 나열하는 실행의 차이는 [go run 공식 문서](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)를 참고합니다.

## 예제를 읽는 순서

1. 연결된 교재 장에서 학습 목표와 실행 결과를 확인합니다.
2. 소스 파일을 읽고 결과를 예상한 뒤 실행합니다.
3. [연습 문제](../exercises/README.md)를 풀며 한 가지씩 바꿔 봅니다.
4. `go test ./...`로 교재 예제와 풀이의 검증 결과를 확인합니다.

실습을 위한 수정은 `/_practice/` 아래에 복사해서 진행할 수 있습니다. 이 폴더는 Git에 포함되지 않으며, `go test ./...`의 기본 패키지 검색에서도 제외됩니다. 반영할 공식 예제를 작성할 때는 `examples/` 아래에 새 폴더를 추가합니다.

예제를 추가할 때는 [집필 가이드](../docs/AUTHORING.md)를 따르고 이 목록을 갱신합니다. 실행 결과가 의도적으로 달라지는 예제는 그 이유와 확인 방법을 함께 적습니다.

참고: [Go command — package lists and patterns](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns).
