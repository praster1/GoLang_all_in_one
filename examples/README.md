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

## 05. 배열과 슬라이스

[본문과 실행 절차](../book/05-arrays-slices/README.md) · [전체 소스](05-arrays-slices/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [append-growth](05-arrays-slices/append-growth/main.go) | `go run ./examples/05-arrays-slices/append-growth` | 추가할 때 저장 공간은 언제 달라질까? |
| [arrays](05-arrays-slices/arrays/main.go) | `go run ./examples/05-arrays-slices/arrays` | 배열을 함수에 전달하면 원본도 바뀔까? |
| [copy](05-arrays-slices/copy/main.go) | `go run ./examples/05-arrays-slices/copy` | 원소를 독립적으로 보관하려면? |
| [filter](05-arrays-slices/filter/main.go) | `go run ./examples/05-arrays-slices/filter` | 입력을 보존하며 결과를 고르려면? |
| [slice-basics](05-arrays-slices/slice-basics/main.go) | `go run ./examples/05-arrays-slices/slice-basics` | nil, 빈 값, 길이와 용량은 무엇일까? |
| [slice-functions](05-arrays-slices/slice-functions/main.go) | `go run ./examples/05-arrays-slices/slice-functions` | 원소 변경과 길이 변경은 어떻게 다를까? |
| [slice-view](05-arrays-slices/slice-view/main.go) | `go run ./examples/05-arrays-slices/slice-view` | 부분 범위는 복사본일까? |

이 장의 예제 검사: `go test ./examples/05-arrays-slices/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 06. 맵과 집계

[본문과 실행 절차](../book/06-maps/README.md) · [전체 소스](06-maps/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [basics](06-maps/basics/main.go) | `go run ./examples/06-maps/basics` | 이름으로 값을 저장하고 삭제하려면? |
| [frequency](06-maps/frequency/main.go) | `go run ./examples/06-maps/frequency` | 여러 번 등장한 항목을 세려면? |
| [groups](06-maps/groups/main.go) | `go run ./examples/06-maps/groups` | 키 하나에 여러 값을 담으려면? |
| [lookup](06-maps/lookup/main.go) | `go run ./examples/06-maps/lookup` | 0은 저장된 값일까, 없는 키일까? |
| [map-functions](06-maps/map-functions/main.go) | `go run ./examples/06-maps/map-functions` | 함수에서 원본 변경과 복사를 구분하려면? |
| [nil-map](06-maps/nil-map/main.go) | `go run ./examples/06-maps/nil-map` | nil 맵에는 무엇을 할 수 있을까? |
| [sets](06-maps/sets/main.go) | `go run ./examples/06-maps/sets` | 처음 등장한 순서로 중복을 없애려면? |

이 장의 예제 검사: `go test ./examples/06-maps/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 07. 문자열과 유니코드

[본문과 실행 절차](../book/07-strings-unicode/README.md) · [전체 소스](07-strings-unicode/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [bytes-runes](07-strings-unicode/bytes-runes/main.go) | `go run ./examples/07-strings-unicode/bytes-runes` | 길이와 range 위치 |
| [classification](07-strings-unicode/classification/main.go) | `go run ./examples/07-strings-unicode/classification` | 유니코드 분류 |
| [conversion](07-strings-unicode/conversion/main.go) | `go run ./examples/07-strings-unicode/conversion` | 바이트·룬 슬라이스 변환 |
| [rune-prefix](07-strings-unicode/rune-prefix/main.go) | `go run ./examples/07-strings-unicode/rune-prefix` | 룬 개수로 앞부분 선택 |
| [strings-ops](07-strings-unicode/strings-ops/main.go) | `go run ./examples/07-strings-unicode/strings-ops` | 공백·분리·검색 |
| [utf8-validation](07-strings-unicode/utf8-validation/main.go) | `go run ./examples/07-strings-unicode/utf8-validation` | 잘못된 인코딩과 대체 문자 |
| [word-count](07-strings-unicode/word-count/main.go) | `go run ./examples/07-strings-unicode/word-count` | 공백 분리와 소문자 집계 |

이 장의 예제 검사: `go test ./examples/07-strings-unicode/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 08. 포인터와 구조체

[본문과 실행 절차](../book/08-pointers-structs/README.md) · [전체 소스](08-pointers-structs/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [clone-struct](08-pointers-structs/clone-struct/main.go) | `go run ./examples/08-pointers-structs/clone-struct` | 실행 예제: 수정 가능한 필드 분리하기 |
| [pointer-functions](08-pointers-structs/pointer-functions/main.go) | `go run ./examples/08-pointers-structs/pointer-functions` | 실행 예제: 복사·수정·재대입 비교 |
| [pointers](08-pointers-structs/pointers/main.go) | `go run ./examples/08-pointers-structs/pointers` | 실행 예제: 두 이름으로 같은 변수에 접근하기 |
| [struct-basics](08-pointers-structs/struct-basics/main.go) | `go run ./examples/08-pointers-structs/struct-basics` | 실행 예제: 제로 값·필드·구조체 복사 |
| [struct-functions](08-pointers-structs/struct-functions/main.go) | `go run ./examples/08-pointers-structs/struct-functions` | 실행 예제: 검사한 뒤 원본 갱신하기 |
| [struct-maps](08-pointers-structs/struct-maps/main.go) | `go run ./examples/08-pointers-structs/struct-maps` | 실행 예제: 재고 항목 갱신하기 |
| [struct-slices](08-pointers-structs/struct-slices/main.go) | `go run ./examples/08-pointers-structs/struct-slices` | 실행 예제: range 값과 인덱스 구분하기 |

이 장의 예제 검사: `go test ./examples/08-pointers-structs/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 09. 메서드와 인터페이스

[본문과 실행 절차](../book/09-methods-interfaces/README.md) · [전체 소스](09-methods-interfaces/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [embedding](09-methods-interfaces/embedding/main.go) | `go run ./examples/09-methods-interfaces/embedding` | 실행 예제: 내부 메서드의 승격 |
| [interfaces](09-methods-interfaces/interfaces/main.go) | `go run ./examples/09-methods-interfaces/interfaces` | 실행 예제: 서로 다른 타입을 같은 함수에 전달하기 |
| [method-sets](09-methods-interfaces/method-sets/main.go) | `go run ./examples/09-methods-interfaces/method-sets` | 실행 예제: 값 복사와 포인터 공유를 인터페이스에서도 확인하기 |
| [pointer-methods](09-methods-interfaces/pointer-methods/main.go) | `go run ./examples/09-methods-interfaces/pointer-methods` | 실행 예제: 포인터 수신자와 nil 계약 |
| [type-assertions](09-methods-interfaces/type-assertions/main.go) | `go run ./examples/09-methods-interfaces/type-assertions` | 실행 예제: 안전한 단언과 분기 |
| [typed-nil](09-methods-interfaces/typed-nil/main.go) | `go run ./examples/09-methods-interfaces/typed-nil` | 실행 예제: typed nil과 메서드 호출 |
| [value-methods](09-methods-interfaces/value-methods/main.go) | `go run ./examples/09-methods-interfaces/value-methods` | 실행 예제: 값을 받아 새 값 반환하기 |

이 장의 예제 검사: `go test ./examples/09-methods-interfaces/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 10. 패키지와 모듈

[본문과 실행 절차](../book/10-packages-modules/README.md) · [전체 소스](10-packages-modules/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [aliases](10-packages-modules/aliases/main.go) | `go run ./examples/10-packages-modules/aliases` | 같은 패키지 이름과 import 별칭 |
| [commands/cmd/check](10-packages-modules/commands/cmd/check/main.go) | `go run ./examples/10-packages-modules/commands/cmd/check` | 공유 패키지의 둘째 명령 |
| [commands/cmd/show](10-packages-modules/commands/cmd/show/main.go) | `go run ./examples/10-packages-modules/commands/cmd/show` | 공유 패키지의 첫 명령 |
| [dependency-direction](10-packages-modules/dependency-direction/main.go) | `go run ./examples/10-packages-modules/dependency-direction` | 소비 측 인터페이스와 의존 방향 |
| [greeting](10-packages-modules/greeting/main.go) | `go run ./examples/10-packages-modules/greeting` | 다른 패키지의 공개 함수 |
| [initialization](10-packages-modules/initialization/main.go) | `go run ./examples/10-packages-modules/initialization` | 패키지 초기화와 main |
| [internal-demo](10-packages-modules/internal-demo/main.go) | `go run ./examples/10-packages-modules/internal-demo` | 내부 패키지의 사용 범위 |
| [multifile](10-packages-modules/multifile/main.go) | `go run ./examples/10-packages-modules/multifile` | 같은 패키지의 여러 파일 |
| [private-state](10-packages-modules/private-state/main.go) | `go run ./examples/10-packages-modules/private-state` | 필드 숨기기와 변경 메서드 |

이 장의 예제 검사: `go test ./examples/10-packages-modules/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 11. 오류 처리와 자원 정리

[본문과 실행 절차](../book/11-errors-defer/README.md) · [전체 소스](11-errors-defer/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [defer-order](11-errors-defer/defer-order/main.go) | `go run ./examples/11-errors-defer/defer-order` | 등록 시점과 역순 실행 |
| [defer-scope](11-errors-defer/defer-scope/main.go) | `go run ./examples/11-errors-defer/defer-scope` | 항목별 함수 수명과 정리 |
| [join](11-errors-defer/join/main.go) | `go run ./examples/11-errors-defer/join` | 여러 오류의 집합 |
| [panic-boundary](11-errors-defer/panic-boundary/main.go) | `go run ./examples/11-errors-defer/panic-boundary` | 제한된 복구와 전파 |
| [resource](11-errors-defer/resource/main.go) | `go run ./examples/11-errors-defer/resource` | 읽기와 닫기의 오류 보존 |
| [return-error](11-errors-defer/return-error/main.go) | `go run ./examples/11-errors-defer/return-error` | 오류를 반환하고 확인 |
| [typed-error](11-errors-defer/typed-error/main.go) | `go run ./examples/11-errors-defer/typed-error` | 오류 타입과 errors.As |
| [typed-nil](11-errors-defer/typed-nil/main.go) | `go run ./examples/11-errors-defer/typed-nil` | error 안의 nil 포인터 |
| [wrapping](11-errors-defer/wrapping/main.go) | `go run ./examples/11-errors-defer/wrapping` | 문맥, %w, errors.Is |

이 장의 예제 검사: `go test ./examples/11-errors-defer/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 12. 자동 테스트

[본문과 실행 절차](../book/12-testing/README.md) · [전체 소스](12-testing/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [errors](12-testing/errors/main.go) | `go run ./examples/12-testing/errors` | 오류를 문자열만으로 검사하지 않기 |
| [example](12-testing/example/main.go) | `go run ./examples/12-testing/example` | 실제 프로그램에서도 같은 패키지 사용 |
| [fixtures](12-testing/fixtures/main.go) | `go run ./examples/12-testing/fixtures` | 임시 자료와 자원 정리 |
| [grade](12-testing/grade/main.go) | `go run ./examples/12-testing/grade` | 먼저 계약을 적기 |
| [settings](12-testing/settings/main.go) | `go run ./examples/12-testing/settings` | 테스트 사이에 상태를 남기지 않기 |
| [subtests](12-testing/subtests/main.go) | `go run ./examples/12-testing/subtests` | 하위 테스트와 공통 검사 함수 |
| [table](12-testing/table/main.go) | `go run ./examples/12-testing/table` | 테이블 테스트로 경계 정리하기 |

이 장의 예제 검사: `go test ./examples/12-testing/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 13. 제네릭

[본문과 실행 절차](../book/13-generics/README.md) · [전체 소스](13-generics/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [filter](13-generics/filter/main.go) | `go run ./examples/13-generics/filter` | 정의된 슬라이스 타입 보존 |
| [first](13-generics/first/main.go) | `go run ./examples/13-generics/first` | 타입 인자·추론·영값 |
| [generic-methods](13-generics/generic-methods/main.go) | `go run ./examples/13-generics/generic-methods` | Go 1.27 메서드의 타입 매개변수 |
| [index](13-generics/index/main.go) | `go run ./examples/13-generics/index` | comparable과 비교 |
| [ordered](13-generics/ordered/main.go) | `go run ./examples/13-generics/ordered` | 순서 제약과 NaN 정책 |
| [stack](13-generics/stack/main.go) | `go run ./examples/13-generics/stack` | 제네릭 구조체와 수신자 |
| [stdlib](13-generics/stdlib/main.go) | `go run ./examples/13-generics/stdlib` | slices·cmp 사용 |
| [sum](13-generics/sum/main.go) | `go run ./examples/13-generics/sum` | 타입 집합과 기반 타입 |
| [transform](13-generics/transform/main.go) | `go run ./examples/13-generics/transform` | 입력·결과 타입이 다른 변환 |

이 장의 예제 검사: `go test ./examples/13-generics/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 14. 파일과 스트림

[본문과 실행 절차](../book/14-files-io/README.md) · [전체 소스](14-files-io/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [buffered-write](14-files-io/buffered-write/main.go) | `go run ./examples/14-files-io/buffered-write` | 버퍼와 Flush |
| [file-roundtrip](14-files-io/file-roundtrip/main.go) | `go run ./examples/14-files-io/file-roundtrip` | 임시 파일 전체 읽기·쓰기 |
| [open-close](14-files-io/open-close/main.go) | `go run ./examples/14-files-io/open-close` | 열린 파일의 소유권 |
| [paths](14-files-io/paths/main.go) | `go run ./examples/14-files-io/paths` | 경로 조합과 디렉터리 목록 |
| [read-chunks](14-files-io/read-chunks/main.go) | `go run ./examples/14-files-io/read-chunks` | 작은 버퍼와 EOF |
| [scan-lines](14-files-io/scan-lines/main.go) | `go run ./examples/14-files-io/scan-lines` | 줄 단위 처리와 오류 확인 |
| [stream-copy](14-files-io/stream-copy/main.go) | `go run ./examples/14-files-io/stream-copy` | Reader에서 Writer로 복사 |

이 장의 예제 검사: `go test ./examples/14-files-io/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 15. JSON·CSV·시간 다루기

[본문과 실행 절차](../book/15-data-time/README.md) · [전체 소스](15-data-time/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [csv-read](15-data-time/csv-read/main.go) | `go run ./examples/15-data-time/csv-read` | CSV 레코드 읽기 |
| [csv-write](15-data-time/csv-write/main.go) | `go run ./examples/15-data-time/csv-write` | CSV 인용과 Flush |
| [json-decode](15-data-time/json-decode/main.go) | `go run ./examples/15-data-time/json-decode` | JSON과 입력 검증 |
| [json-encode](15-data-time/json-encode/main.go) | `go run ./examples/15-data-time/json-encode` | 구조체를 JSON으로 |
| [json-presence](15-data-time/json-presence/main.go) | `go run ./examples/15-data-time/json-presence` | 누락과 명시적 false |
| [time-arithmetic](15-data-time/time-arithmetic/main.go) | `go run ./examples/15-data-time/time-arithmetic` | 기간과 시각 계산 |
| [time-parse](15-data-time/time-parse/main.go) | `go run ./examples/15-data-time/time-parse` | 파싱과 시간대 |

이 장의 예제 검사: `go test ./examples/15-data-time/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 16. 명령줄 프로그램 만들기

[본문과 실행 절차](../book/16-cli/README.md) · [전체 소스](16-cli/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [app](16-cli/app/main.go) | [표준 입력 파이프 명령](../book/16-cli/README.md) | 인자·입력·출력·종료 코드 연결 |
| [arguments](16-cli/arguments/main.go) | `go run ./examples/16-cli/arguments Go "file input"` | 인자 배열 |
| [environment](16-cli/environment/main.go) | `go run ./examples/16-cli/environment` | LookupEnv와 기본값 |
| [exit-code](16-cli/exit-code/main.go) | `go run ./examples/16-cli/exit-code 5` | 0·1·2 종료 코드 |
| [flags](16-cli/flags/main.go) | `go run ./examples/16-cli/flags -name learner -upper` | 이름·불리언 플래그 |
| [stdin-lines](16-cli/stdin-lines/main.go) | [표준 입력 파이프 명령](../book/16-cli/README.md) | 표준 입력의 EOF까지 읽기 |
| [streams](16-cli/streams/main.go) | `go run ./examples/16-cli/streams` | 결과와 진단 분리 |

이 장의 예제 검사: `go test ./examples/16-cli/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 17. Context로 취소와 마감 전달하기

[본문과 실행 절차](../book/17-context/README.md) · [전체 소스](17-context/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [background](17-context/background/main.go) | `go run ./examples/17-context/background` | 취소 없는 시작 Context |
| [cancel](17-context/cancel/main.go) | `go run ./examples/17-context/cancel` | 명시적 취소 |
| [cancel-cause](17-context/cancel-cause/main.go) | `go run ./examples/17-context/cancel-cause` | 취소의 구체적인 원인 |
| [checkpoints](17-context/checkpoints/main.go) | `go run ./examples/17-context/checkpoints` | 작업 사이의 취소 확인 |
| [deadlines](17-context/deadlines/main.go) | `go run ./examples/17-context/deadlines` | 마감과 제한 시간 |
| [hierarchy](17-context/hierarchy/main.go) | `go run ./examples/17-context/hierarchy` | 부모·자식·형제 |
| [values](17-context/values/main.go) | `go run ./examples/17-context/values` | 요청 범위 메타데이터 |

이 장의 예제 검사: `go test ./examples/17-context/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 18. 고루틴과 완료 대기

[본문과 실행 절차](../book/18-goroutines/README.md) · [전체 소스](18-goroutines/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [arguments](18-goroutines/arguments/main.go) | `go run ./examples/18-goroutines/arguments` | 호출 인자의 평가 시점 |
| [cancel-join](18-goroutines/cancel-join/main.go) | `go run ./examples/18-goroutines/cancel-join` | 취소와 완료 확인 |
| [early-return](18-goroutines/early-return/main.go) | `go run ./examples/18-goroutines/early-return` | 조기 반환과 defer |
| [error-results](18-goroutines/error-results/main.go) | `go run ./examples/18-goroutines/error-results` | 오류를 결과로 모으기 |
| [indexed-results](18-goroutines/indexed-results/main.go) | `go run ./examples/18-goroutines/indexed-results` | 결과 칸의 독립된 소유권 |
| [launch-wait](18-goroutines/launch-wait/main.go) | `go run ./examples/18-goroutines/launch-wait` | 시작·등록·완료·대기 |
| [waitgroup-go](18-goroutines/waitgroup-go/main.go) | `go run ./examples/18-goroutines/waitgroup-go` | Go 1.25부터의 Go 메서드 |

이 장의 예제 검사: `go test ./examples/18-goroutines/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 19. 채널과 select

[본문과 실행 절차](../book/19-channels-select/README.md) · [전체 소스](19-channels-select/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [buffered-close](19-channels-select/buffered-close/main.go) | `go run ./examples/19-channels-select/buffered-close` | 버퍼와 닫힌 뒤 남은 값 |
| [cancel-send](19-channels-select/cancel-send/main.go) | `go run ./examples/19-channels-select/cancel-send` | 막힌 전송의 취소 |
| [closed-select](19-channels-select/closed-select/main.go) | `go run ./examples/19-channels-select/closed-select` | 닫힌 입력을 nil로 비활성화 |
| [directions](19-channels-select/directions/main.go) | `go run ./examples/19-channels-select/directions` | 송신·수신 전용 타입 |
| [merge](19-channels-select/merge/main.go) | `go run ./examples/19-channels-select/merge` | 여러 입력을 하나로 병합 |
| [pipeline](19-channels-select/pipeline/main.go) | `go run ./examples/19-channels-select/pipeline` | 단계 연결과 조기 종료 |
| [rendezvous](19-channels-select/rendezvous/main.go) | `go run ./examples/19-channels-select/rendezvous` | 버퍼 없는 전송·수신 |
| [select-default](19-channels-select/select-default/main.go) | `go run ./examples/19-channels-select/select-default` | 준비된 통신과 default |

이 장의 예제 검사: `go test ./examples/19-channels-select/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 20. 공유 상태와 동기화

[본문과 실행 절차](../book/20-sync/README.md) · [전체 소스](20-sync/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [atomic-capacity](20-sync/atomic-capacity/main.go) | `go run ./examples/20-sync/atomic-capacity` | 검사와 갱신을 CAS로 연결 |
| [atomic-counter](20-sync/atomic-counter/main.go) | `go run ./examples/20-sync/atomic-counter` | 하나의 정수에 대한 원자 덧셈 |
| [channel-owner](20-sync/channel-owner/main.go) | `go run ./examples/20-sync/channel-owner` | 한 소비자가 맵의 소유권 가짐 |
| [ledger](20-sync/ledger/main.go) | `go run ./examples/20-sync/ledger` | 검사와 두 값 변경의 불변 조건 |
| [lock-scope](20-sync/lock-scope/main.go) | `go run ./examples/20-sync/lock-scope` | 잠금 밖에서 콜백 실행 |
| [mutex-counter](20-sync/mutex-counter/main.go) | `go run ./examples/20-sync/mutex-counter` | 같은 잠금으로 읽기·쓰기 보호 |
| [once-config](20-sync/once-config/main.go) | `go run ./examples/20-sync/once-config` | 설정 초기화의 결과 공유 |
| [rw-snapshot](20-sync/rw-snapshot/main.go) | `go run ./examples/20-sync/rw-snapshot` | 읽기 잠금과 독립된 맵 복사 |

이 장의 예제 검사: `go test ./examples/20-sync/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 21. HTTP 클라이언트

[본문과 실행 절차](../book/21-http-client/README.md) · [전체 소스](21-http-client/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [basic-get](21-http-client/basic-get/main.go) | `go run ./examples/21-http-client/basic-get` | 요청·응답의 기본 수명 |
| [bounded-body](21-http-client/bounded-body/main.go) | `go run ./examples/21-http-client/bounded-body` | 본문 크기 한도와 경계 |
| [cancellation](21-http-client/cancellation/main.go) | `go run ./examples/21-http-client/cancellation` | 진행 중인 요청 취소 |
| [json-response](21-http-client/json-response/main.go) | `go run ./examples/21-http-client/json-response` | 제한된 JSON 해석과 검증 |
| [query-headers](21-http-client/query-headers/main.go) | `go run ./examples/21-http-client/query-headers` | 쿼리 인코딩과 요청 헤더 |
| [redirect-policy](21-http-client/redirect-policy/main.go) | `go run ./examples/21-http-client/redirect-policy` | TLS 테스트와 이동 정책 |
| [status-error](21-http-client/status-error/main.go) | `go run ./examples/21-http-client/status-error` | 전송 오류와 상태 오류 |

이 장의 예제 검사: `go test ./examples/21-http-client/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 22. HTTP 서버와 라우팅

[본문과 실행 절차](../book/22-http-server/README.md) · [전체 소스](22-http-server/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [graceful-shutdown](22-http-server/graceful-shutdown/main.go) | `go run ./examples/22-http-server/graceful-shutdown` | 수신 시작과 정상 종료 |
| [hello-handler](22-http-server/hello-handler/main.go) | `go run ./examples/22-http-server/hello-handler` | 핸들러와 첫 응답 |
| [json-create](22-http-server/json-create/main.go) | `go run ./examples/22-http-server/json-create` | 제한된 JSON 요청 |
| [json-errors](22-http-server/json-errors/main.go) | `go run ./examples/22-http-server/json-errors` | 구조화한 실패 응답 |
| [method-routing](22-http-server/method-routing/main.go) | `go run ./examples/22-http-server/method-routing` | 메서드·경로·405 |
| [path-values](22-http-server/path-values/main.go) | `go run ./examples/22-http-server/path-values` | 경로 변수와 정수 검증 |
| [request-context](22-http-server/request-context/main.go) | `go run ./examples/22-http-server/request-context` | 서버 작업 취소 |

이 장의 예제 검사: `go test ./examples/22-http-server/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 23. API 설계와 핸들러 테스트

[본문과 실행 절차](../book/23-api-testing/README.md) · [전체 소스](23-api-testing/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [concurrent-counter](23-api-testing/concurrent-counter/main.go) | `go run ./examples/23-api-testing/concurrent-counter` | 동시 변경과 읽기 |
| [handler-table](23-api-testing/handler-table/main.go) | `go run ./examples/23-api-testing/handler-table` | 상태·헤더·JSON 계약 |
| [injected-store](23-api-testing/injected-store/main.go) | `go run ./examples/23-api-testing/injected-store` | 저장소 결과와 오류 분류 |
| [integration](23-api-testing/integration/main.go) | `go run ./examples/23-api-testing/integration` | TLS·전송·HEAD |
| [json-contract](23-api-testing/json-contract/main.go) | `go run ./examples/23-api-testing/json-contract` | 필수 필드·크기·값 하나 |
| [request-id](23-api-testing/request-id/main.go) | `go run ./examples/23-api-testing/request-id` | 미들웨어와 컨텍스트 |
| [validation](23-api-testing/validation/main.go) | `go run ./examples/23-api-testing/validation` | 순수 검증과 정규화 |

이 장의 예제 검사: `go test ./examples/23-api-testing/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 24. SQL과 트랜잭션

[본문과 실행 절차](../book/24-sql-transactions/README.md) · [전체 소스](24-sql-transactions/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [file-reopen](24-sql-transactions/file-reopen/main.go) | `go run ./examples/24-sql-transactions/file-reopen` | 파일 DB를 닫고 다시 열기 |
| [null-not-found](24-sql-transactions/null-not-found/main.go) | `go run ./examples/24-sql-transactions/null-not-found` | NULL과 없는 행 |
| [open-ping](24-sql-transactions/open-ping/main.go) | `go run ./examples/24-sql-transactions/open-ping` | DB 핸들과 연결 확인 |
| [parameterized](24-sql-transactions/parameterized/main.go) | `go run ./examples/24-sql-transactions/parameterized` | 값의 매개변수 전달 |
| [pool-cancellation](24-sql-transactions/pool-cancellation/main.go) | `go run ./examples/24-sql-transactions/pool-cancellation` | 연결 대기와 취소 |
| [query-rows](24-sql-transactions/query-rows/main.go) | `go run ./examples/24-sql-transactions/query-rows` | 목록, Scan, Close, Err |
| [transactions](24-sql-transactions/transactions/main.go) | `go run ./examples/24-sql-transactions/transactions` | commit과 rollback |

이 장의 예제 검사: `go test ./examples/24-sql-transactions/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 25. 서비스 구성과 운영 기초

[본문과 실행 절차](../book/25-service-operations/README.md) · [전체 소스](25-service-operations/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [bearer-auth](25-service-operations/bearer-auth/main.go) | `go run ./examples/25-service-operations/bearer-auth` | 요청 앞에서 자격 증명 확인 |
| [configuration](25-service-operations/configuration/main.go) | `go run ./examples/25-service-operations/configuration` | 기본값, 존재 여부, 파싱, 범위 검증 |
| [health-readiness](25-service-operations/health-readiness/main.go) | `go run ./examples/25-service-operations/health-readiness` | 생존 확인과 준비 상태 |
| [request-log](25-service-operations/request-log/main.go) | `go run ./examples/25-service-operations/request-log` | 응답 상태와 경로 패턴 기록 |
| [secret-value](25-service-operations/secret-value/main.go) | `go run ./examples/25-service-operations/secret-value` | `LogValuer`와 허용한 필드만 기록하기 |
| [shutdown](25-service-operations/shutdown/main.go) | `go run ./examples/25-service-operations/shutdown` | 실제 소켓을 열고 정상 종료 |
| [structured-log](25-service-operations/structured-log/main.go) | `go run ./examples/25-service-operations/structured-log` | JSON 로그, 수준, 공통 속성 |

이 장의 예제 검사: `go test ./examples/25-service-operations/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 26. 성능 측정과 진단

[본문과 실행 절차](../book/26-performance/README.md) · [전체 소스](26-performance/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [join](26-performance/join/main.go) | `go run ./examples/26-performance/join` | 문자열 연결을 측정하기 |
| [lookup](26-performance/lookup/main.go) | `go run ./examples/26-performance/lookup` | 조회만 잴지, 자료구조 생성까지 잴지 정하기 |
| [parallel](26-performance/parallel/main.go) | `go run ./examples/26-performance/parallel` | 병렬 벤치마크와 공유 자원 |
| [parse](26-performance/parse/main.go) | `go run ./examples/26-performance/parse` | 처리 바이트와 오류 경로 |
| [reserve](26-performance/reserve/main.go) | `go run ./examples/26-performance/reserve` | 길이를 알고 있을 때 슬라이스 용량 준비하기 |
| [retention](26-performance/retention/main.go) | `go run ./examples/26-performance/retention` | 작은 결과가 큰 저장 공간을 붙잡는 경우 |
| [workload](26-performance/workload/main.go) | `go run ./examples/26-performance/workload` | CPU 프로파일 만들기 |

이 장의 예제 검사: `go test ./examples/26-performance/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 27. 퍼징과 견고한 테스트

[본문과 실행 절차](../book/27-fuzzing/README.md) · [전체 소스](27-fuzzing/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [clamp](27-fuzzing/clamp/main.go) | `go run ./examples/27-fuzzing/clamp` | 정수 전체 범위를 다루는 보정 함수 |
| [csv-pair](27-fuzzing/csv-pair/main.go) | `go run ./examples/27-fuzzing/csv-pair` | CSV 왕복에서 형식의 규칙 확인하기 |
| [hex](27-fuzzing/hex/main.go) | `go run ./examples/27-fuzzing/hex` | 바이트 인코딩의 왕복 성질 |
| [interval](27-fuzzing/interval/main.go) | `go run ./examples/27-fuzzing/interval` | 숫자 범위 파서의 성공 계약 |
| [key-value](27-fuzzing/key-value/main.go) | `go run ./examples/27-fuzzing/key-value` | 구분자는 첫 번째 것만 의미할 수 있다 |
| [normalize](27-fuzzing/normalize/main.go) | `go run ./examples/27-fuzzing/normalize` | 공백 정리와 멱등성 |
| [reverse](27-fuzzing/reverse/main.go) | `go run ./examples/27-fuzzing/reverse` | 첫 퍼즈 테스트: 룬 순서 뒤집기 |

이 장의 예제 검사: `go test ./examples/27-fuzzing/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 28. 빌드와 배포 자동화

[본문과 실행 절차](../book/28-build-deploy/README.md) · [전체 소스](28-build-deploy/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [checksum](28-build-deploy/checksum/main.go) | `go run ./examples/28-build-deploy/checksum` | 체크섬으로 파일 내용 확인하기 |
| [cli](28-build-deploy/cli/main.go) | `go run ./examples/28-build-deploy/cli` | 배포할 작은 CLI |
| [config](28-build-deploy/config/main.go) | `go run ./examples/28-build-deploy/config` | 빌드 값과 실행 설정 나누기 |
| [embed](28-build-deploy/embed/main.go) | `go run ./examples/28-build-deploy/embed` | 실행 파일에 포함할 자원 |
| [platform](28-build-deploy/platform/main.go) | `go run ./examples/28-build-deploy/platform` | 빌드 대상에 따라 파일 선택하기 |
| [smoke](28-build-deploy/smoke/main.go) | `go run ./examples/28-build-deploy/smoke` | 최소 실행 확인과 서비스 상태 |
| [version](28-build-deploy/version/main.go) | `go run ./examples/28-build-deploy/version` | 버전 정보를 실행 파일에 넣기 |

이 장의 예제 검사: `go test ./examples/28-build-deploy/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 29. 종합 실습: CSV 분석 CLI

[본문과 실행 절차](../book/29-csv-cli/README.md) · [전체 소스](29-csv-cli/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [aggregate](29-csv-cli/aggregate/main.go) | `go run ./examples/29-csv-cli/aggregate` | 최소 집계 실행 |
| [cmd/csvstats](29-csv-cli/cmd/csvstats/main.go) | `go run ./examples/29-csv-cli/cmd/csvstats -input examples/29-csv-cli/testdata/sales.csv` | 파일·표준 입력 검증과 CSV·JSON 집계 CLI |
| [diagnostics](29-csv-cli/diagnostics/main.go) | `go run ./examples/29-csv-cli/diagnostics` | 위치 있는 오류 읽기 |
| [json-output](29-csv-cli/json-output/main.go) | `go run ./examples/29-csv-cli/json-output` | 빈 목록의 JSON 표현 |
| [limits](29-csv-cli/limits/main.go) | `go run ./examples/29-csv-cli/limits` | 경계에서 전체 실패하기 |
| [parse-row](29-csv-cli/parse-row/main.go) | `go run ./examples/29-csv-cli/parse-row` | 최소 실행 예 |
| [sort-results](29-csv-cli/sort-results/main.go) | `go run ./examples/29-csv-cli/sort-results` | 출력 함수가 원본을 정렬하지 않게 하기 |

이 장의 예제 검사: `go test ./examples/29-csv-cli/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 30. 종합 실습: 할 일 HTTP API

[본문과 실행 절차](../book/30-todo-api/README.md) · [전체 소스](30-todo-api/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [client-demo](30-todo-api/client-demo/main.go) | `go run ./examples/30-todo-api/client-demo` | 실제 로컬 HTTP 왕복 |
| [domain-demo](30-todo-api/domain-demo/main.go) | `go run ./examples/30-todo-api/domain-demo` | 제목 정규화와 입력 오류 |
| [http-demo](30-todo-api/http-demo/main.go) | `go run ./examples/30-todo-api/http-demo` | 핸들러에 CRUD 요청 전달 |
| [memory-demo](30-todo-api/memory-demo/main.go) | `go run ./examples/30-todo-api/memory-demo` | 저장소 생성·수정·삭제 |
| [pagination-demo](30-todo-api/pagination-demo/main.go) | `go run ./examples/30-todo-api/pagination-demo` | ID 기준 페이지 순회 |
| [server](30-todo-api/server/main.go) | `go run ./examples/30-todo-api/server -store=memory -addr=127.0.0.1:8080` (Ctrl+C로 종료) | 메모리·SQLite 저장소 선택과 실제 HTTP 서버 |
| [sqlite-demo](30-todo-api/sqlite-demo/main.go) | `go run ./examples/30-todo-api/sqlite-demo` | 파일 DB를 닫고 다시 열기 |

이 장의 예제 검사: `go test ./examples/30-todo-api/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 31. 종합 실습: 동시 배치 처리

[본문과 실행 절차](../book/31-concurrent-batch/README.md) · [전체 소스](31-concurrent-batch/)

| 프로그램 | 기본 실행 | 학습 내용 |
| --- | --- | --- |
| [bounded](31-concurrent-batch/bounded/main.go) | `go run ./examples/31-concurrent-batch/bounded` | 고정 작업자와 순서 보존 |
| [cancellation](31-concurrent-batch/cancellation/main.go) | `go run ./examples/31-concurrent-batch/cancellation` | 취소·미시작·부분 성공 |
| [cmd/batch](31-concurrent-batch/cmd/batch/main.go) | `go run ./examples/31-concurrent-batch/cmd/batch` | 파일·표준 입력·JSON·종료 코드 |
| [limits](31-concurrent-batch/limits/main.go) | `go run ./examples/31-concurrent-batch/limits` | 실행 전 설정·입력 검증 |
| [report](31-concurrent-batch/report/main.go) | `go run ./examples/31-concurrent-batch/report` | 완료 결과를 요약 |
| [retries](31-concurrent-batch/retries/main.go) | `go run ./examples/31-concurrent-batch/retries` | 일시 오류와 영구 오류 구분 |
| [sequential](31-concurrent-batch/sequential/main.go) | `go run ./examples/31-concurrent-batch/sequential` | 처리 계약을 순차 함수로 확인 |

이 장의 예제 검사: `go test ./examples/31-concurrent-batch/...`. 서버·입력 파일·설정 인자가 필요한 프로그램은 본문의 실행·종료 절차를 함께 확인합니다.

## 예제를 읽는 순서

1. 연결된 교재 장에서 학습 목표와 실행 결과를 확인합니다.
2. 소스 파일을 읽고 결과를 예상한 뒤 실행합니다.
3. [연습 문제](../exercises/README.md)를 풀며 한 가지씩 바꿔 봅니다.
4. `go test ./...`로 교재 예제와 풀이의 검증 결과를 확인합니다.

실습을 위한 수정은 `/_practice/` 아래에 복사해서 진행할 수 있습니다. 이 폴더는 Git에 포함되지 않으며, `go test ./...`의 기본 패키지 검색에서도 제외됩니다. 반영할 공식 예제를 작성할 때는 `examples/` 아래에 새 폴더를 추가합니다.

예제를 추가할 때는 [집필 가이드](../docs/AUTHORING.md)를 따르고 이 목록을 갱신합니다. 실행 결과가 의도적으로 달라지는 예제는 그 이유와 확인 방법을 함께 적습니다.

참고: [Go command — package lists and patterns](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns).
