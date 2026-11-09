# 실행 예제

교재에서 설명하는 코드를 장별로 모았습니다. 모든 명령은 저장소 최상위 폴더에서 실행합니다. 각 예제 폴더는 독립적으로 실행할 수 있는 패키지이며, 루트의 `go.mod`를 함께 사용합니다.

## 01. 첫 Go 프로그램

| 예제 | 실행 명령 | 확인할 내용 |
| --- | --- | --- |
| [hello](01-first-program/hello/main.go) | `go run ./examples/01-first-program/hello` | 인사 한 줄 출력 |
| [output](01-first-program/output/main.go) | `go run ./examples/01-first-program/output` | `Print`·`Println`·`Printf`의 공백과 줄바꿈 |
| [multifile](01-first-program/multifile/main.go) · [message.go](01-first-program/multifile/message.go) | `go run ./examples/01-first-program/multifile` | 같은 패키지의 여러 파일을 함께 실행하는 선택 예제 |

`hello`의 출력 예시는 다음과 같습니다.

```text
안녕하세요, Go!
```

`output`은 문장 사이의 공백과 줄바꿈을 비교합니다. `multifile`은 다른 파일에 정의한 함수를 호출하므로 예제 폴더를 실행 대상으로 지정합니다. 각 예제의 예상 출력과 설명은 [01장](../book/01-first-program/README.md)에 있습니다. 패키지를 지정하는 실행과 파일을 직접 나열하는 실행의 차이는 [go run 공식 문서](https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program)를 참고합니다.

## 02. 변수·상수·자료형·연산자

| 예제 | 실행 명령 | 확인할 내용 |
| --- | --- | --- |
| [variables](02-variables-types/variables/main.go) | `go run ./examples/02-variables-types/variables` | 변수 선언과 값 변경 |
| [zero-values](02-variables-types/zero-values/main.go) | `go run ./examples/02-variables-types/zero-values` | 값을 지정하지 않은 변수의 영값과 자료형 |
| [constants](02-variables-types/constants/main.go) | `go run ./examples/02-variables-types/constants` | 상수를 이용해 분을 시간과 나머지 분으로 계산 |
| [conversions](02-variables-types/conversions/main.go) | `go run ./examples/02-variables-types/conversions` | 형 변환과 나눗셈 순서에 따른 결과 비교 |
| [operators](02-variables-types/operators/main.go) | `go run ./examples/02-variables-types/operators` | 산술·비교·논리 연산과 우선순위 |

각 예제의 코드 해설과 예상 출력은 [02장](../book/02-variables-types/README.md)에서 확인합니다. 다섯 예제만 검사하려면 저장소 루트에서 다음을 실행합니다.

```sh
go test -count=1 ./examples/02-variables-types/...
```

## 03. 조건문과 반복문

| 예제 | 실행 명령 | 확인할 내용 |
| --- | --- | --- |
| [conditions](03-control-flow/conditions/main.go) | `go run ./examples/03-control-flow/conditions` | `if`·`else if`·`else`로 실행할 코드 선택 |
| [switch](03-control-flow/switch/main.go) | `go run ./examples/03-control-flow/switch` | 여러 값을 묶은 `case`와 `default` |
| [for-sum](03-control-flow/for-sum/main.go) | `go run ./examples/03-control-flow/for-sum` | 세 부분으로 구성된 `for`와 누적 합계 |
| [countdown](03-control-flow/countdown/main.go) | `go run ./examples/03-control-flow/countdown` | 조건만 사용하는 `for`와 종료 조건 |
| [loop-control](03-control-flow/loop-control/main.go) | `go run ./examples/03-control-flow/loop-control` | `break`로 반복 종료, `continue`로 현재 회차의 나머지 건너뛰기 |
| [range-count](03-control-flow/range-count/main.go) | `go run ./examples/03-control-flow/range-count` | 정수에 대한 `range`와 0부터 시작하는 반복 값 |
| [nested](03-control-flow/nested/main.go) | `go run ./examples/03-control-flow/nested` | 중첩 반복문의 실행 순서 |

각 예제의 코드 해설과 예상 출력은 [03장](../book/03-control-flow/README.md)에서 확인합니다. 일곱 예제만 검사하려면 저장소 루트에서 다음을 실행합니다.

```sh
go test -count=1 ./examples/03-control-flow/...
```

## 04. 함수와 스코프

| 예제 | 실행 명령 | 확인할 내용 |
| --- | --- | --- |
| [greeting](04-functions-scope/greeting/main.go) | `go run ./examples/04-functions-scope/greeting` | 매개변수로 값을 받는 함수 정의와 호출 |
| [value-return](04-functions-scope/value-return/main.go) | `go run ./examples/04-functions-scope/value-return` | 값 전달, 반환값과 호출한 쪽의 대입 |
| [divide](04-functions-scope/divide/main.go) | `go run ./examples/04-functions-scope/divide` | 여러 값 반환과 입력 확인 결과 처리 |
| [scope](04-functions-scope/scope/main.go) | `go run ./examples/04-functions-scope/scope` | 패키지 상수와 블록 안팎의 이름 사용 범위 |
| [variadic](04-functions-scope/variadic/main.go) | `go run ./examples/04-functions-scope/variadic` | 개수가 다른 인자를 받는 가변 인자 함수 |
| [function-values](04-functions-scope/function-values/main.go) | `go run ./examples/04-functions-scope/function-values` | 함수를 변수에 저장하고 다른 함수에 전달 |
| [closure](04-functions-scope/closure/main.go) | `go run ./examples/04-functions-scope/closure` | 클로저의 상태 유지와 카운터 사이의 독립성 |

각 예제의 코드 해설과 예상 출력은 [04장](../book/04-functions-scope/README.md)에서 확인합니다. 일곱 예제만 검사하려면 저장소 루트에서 다음을 실행합니다.

```sh
go test -count=1 ./examples/04-functions-scope/...
```

## 예제를 읽는 순서

1. 연결된 교재 장에서 학습 목표와 실행 결과를 확인합니다.
2. 소스 파일을 읽고 결과를 예상한 뒤 실행합니다.
3. [연습 문제](../exercises/README.md)를 풀며 한 가지씩 바꿔 봅니다.
4. `go test ./...`로 교재 예제와 풀이의 검증 결과를 확인합니다.

실습을 위한 수정은 `/_practice/` 아래에 복사해서 진행할 수 있습니다. 이 폴더는 Git에 포함되지 않으며, `go test ./...`의 기본 패키지 검색에서도 제외됩니다. 반영할 공식 예제를 작성할 때는 `examples/` 아래에 새 폴더를 추가합니다.

예제를 추가할 때는 [집필 가이드](../docs/AUTHORING.md)를 따르고 이 목록을 갱신합니다. 실행 결과가 의도적으로 달라지는 예제는 그 이유와 확인 방법을 함께 적습니다.

참고: [Go command — package lists and patterns](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns).
