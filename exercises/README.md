# 연습문제 안내

연습문제는 설명을 읽은 뒤 출력을 예측하고, 직접 코드를 수정하고, 오류 원인을 확인하는 공간입니다. 교재의 실행 예제는 `examples/`, 문제를 푼 뒤 비교할 해설은 `solutions/`에 있습니다.

## 학습 순서

1. 문제의 입력·출력·조건을 읽습니다.
2. 예상 결과를 적고 직접 작성합니다.
3. 실행 결과와 예상 결과를 비교합니다.
4. 막히면 문제의 힌트를 펼칩니다.
5. 마지막에 해설을 읽고 자신의 풀이와 차이를 설명합니다.

동일한 결과를 만드는 코드는 여러 가지일 수 있습니다. 코드 모양을 외우기보다 요구 사항을 만족하는지, 어떤 동작을 확인했는지 설명해 보세요.

## 개인 연습 폴더와 준비

개인 코드는 저장소 루트의 `/_practice/` 아래에 둡니다. 각 장의 문제 문서에 개인 파일을 준비하는 방법이 있습니다. 명령을 제공하는 장에서는 자신의 PowerShell 또는 Bash/Zsh에 맞는 블록을 선택하고, 편집기로 파일을 만드는 장에서는 안내된 디렉터리를 사용합니다. 기존 풀이 파일을 덮어쓰지 않습니다.

| 장 | 준비 안내 | 개인 파일의 위치 |
| --- | --- | --- |
| 00. 개발 환경과 학습 준비 | [00장 실습 준비](00-setup/README.md) | `_practice/00-setup/main.go` |
| 01. 첫 Go 프로그램 | [01장 문제별 준비](01-first-program/README.md) | `_practice/01-first-program/` 아래의 `message/`, `report/`, `repair/`; 선택 과제는 `greeting/` |
| 02. 변수·상수·자료형·연산자 | [02장 문제별 준비](02-variables-types/README.md) | `_practice/02-variables-types/` 아래의 `study-log/`, `time-parts/`, `repair/`, `average/`, `receipt/`; 선택 과제는 `iota/` |
| 03. 조건문과 반복문 | [03장 문제별 준비](03-control-flow/README.md) | `_practice/03-control-flow/` 아래의 `grade/`, `signal/`, `repair/`, `multiples/`, `goal/`; 선택 과제는 `fizzbuzz/` |
| 04. 함수와 스코프 | [04장 문제별 준비](04-functions-scope/README.md) | `_practice/04-functions-scope/` 아래의 `grade/`, `repair/`, `sum-even/`, `study-plan/`, `stepper/`; 선택 과제는 `apply-twice/` |
| 05. 배열과 슬라이스 | [05장 실습 준비](05-arrays-slices/README.md) | `_practice/05-arrays-slices/` 아래 문제별 폴더 |
| 06. 맵과 집계 | [06장 실습 준비](06-maps/README.md) | `_practice/06-maps/` 아래 문제별 폴더 |
| 07. 문자열과 유니코드 | [07장 실습 준비](07-strings-unicode/README.md) | `_practice/07-strings-unicode/` 아래 문제별 폴더 |
| 08. 포인터와 구조체 | [08장 실습 준비](08-pointers-structs/README.md) | `_practice/08-pointers-structs/` 아래 문제별 폴더 |
| 09. 메서드와 인터페이스 | [09장 실습 준비](09-methods-interfaces/README.md) | `_practice/09-methods-interfaces/` 아래 문제별 폴더 |
| 10. 패키지와 모듈 | [10장 실습 준비](10-packages-modules/README.md) | `_practice/10-packages-modules/` 아래 문제별 폴더 |
| 11. 오류 처리와 자원 정리 | [11장 실습 준비](11-errors-defer/README.md) | `_practice/11-errors-defer/` 아래 문제별 폴더 |
| 12. 자동 테스트 | [12장 실습 준비](12-testing/README.md) | `_practice/12-testing/` 아래 문제별 폴더 |
| 13. 제네릭 | [13장 실습 준비](13-generics/README.md) | `_practice/13-generics/` 아래 문제별 폴더 |
| 14. 파일과 스트림 | [14장 실습 준비](14-files-io/README.md) | `_practice/14-files-io/` 아래 문제별 폴더 |
| 15. JSON·CSV·시간 다루기 | [15장 실습 준비](15-data-time/README.md) | `_practice/15-data-time/` 아래 문제별 폴더 |
| 16. 명령줄 프로그램 만들기 | [16장 실습 준비](16-cli/README.md) | `_practice/16-cli/` 아래 문제별 폴더 |
| 17. Context로 취소와 마감 전달하기 | [17장 실습 준비](17-context/README.md) | `_practice/17-context/` 아래 문제별 폴더 |
| 18. 고루틴과 완료 대기 | [18장 실습 준비](18-goroutines/README.md) | `_practice/18-goroutines/` 아래 문제별 폴더 |
| 19. 채널과 select | [19장 실습 준비](19-channels-select/README.md) | `_practice/19-channels-select/` 아래 문제별 폴더 |
| 20. 공유 상태와 동기화 | [20장 실습 준비](20-sync/README.md) | `_practice/20-sync/` 아래 문제별 폴더 |
| 21. HTTP 클라이언트 | [21장 실습 준비](21-http-client/README.md) | `_practice/21-http-client/` 아래 문제별 폴더 |
| 22. HTTP 서버와 라우팅 | [22장 실습 준비](22-http-server/README.md) | `_practice/22-http-server/` 아래 문제별 폴더 |
| 23. API 설계와 핸들러 테스트 | [23장 실습 준비](23-api-testing/README.md) | `_practice/23-api-testing/` 아래 문제별 폴더 |
| 24. SQL과 트랜잭션 | [24장 실습 준비](24-sql-transactions/README.md) | `_practice/24-sql-transactions/` 아래 문제별 폴더 |
| 25. 서비스 구성과 운영 기초 | [25장 실습 준비](25-service-operations/README.md) | `_practice/25-service-operations/` 아래 문제별 폴더 |
| 26. 성능 측정과 진단 | [26장 실습 준비](26-performance/README.md) | `_practice/26-performance/` 아래 문제별 폴더 |
| 27. 퍼징과 견고한 테스트 | [27장 실습 준비](27-fuzzing/README.md) | `_practice/27-fuzzing/` 아래 문제별 폴더 |
| 28. 빌드와 배포 자동화 | [28장 실습 준비](28-build-deploy/README.md) | `_practice/28-build-deploy/` 아래 문제별 폴더 |
| 29. 종합 실습: CSV 분석 CLI | [29장 실습 준비](29-csv-cli/README.md) | `_practice/29-csv-cli/` 아래 문제별 폴더 |
| 30. 종합 실습: 할 일 HTTP API | [30장 실습 준비](30-todo-api/README.md) | `_practice/30-todo-api/` 아래 문제별 폴더 |
| 31. 종합 실습: 동시 배치 처리 | [31장 실습 준비](31-concurrent-batch/README.md) | `_practice/31-concurrent-batch/` 아래 문제별 폴더 |

연습 폴더에도 새 `go.mod`나 `go.work`를 만들지 않습니다. 루트의 모듈 설정을 사용합니다. 한 폴더에는 실행 프로그램 하나를 두고, 답안을 보관하려면 다른 폴더를 사용하세요. 같은 폴더에 `main` 함수가 있는 파일을 여러 개 두면 중복 선언이 발생합니다.

`/_practice/`는 Git에 포함하지 않습니다. 또한 Go의 `./...` 패키지 탐색에서 이름이 `_`로 시작하는 디렉터리는 제외됩니다. 이는 Git의 무시 설정과 별개인 Go 도구의 동작입니다. 실습 중인 코드를 실행하거나 포맷할 때는 **해당 폴더를 직접 지정**합니다. [Go의 패키지 탐색 규칙](https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns)을 참고합니다.

예를 들어 01장의 준비 단계를 마친 뒤에는 저장소 루트에서 다음 명령을 사용합니다.

```text
go run ./_practice/01-first-program/message
go fmt ./_practice/01-first-program/message
```

`go fmt`는 파일 형식을 실제로 수정합니다. 테스트가 제공되는 과제는 그 과제의 안내에 따라 적용하세요. 다른 문제의 테스트를 한 폴더에 섞지 않습니다.

## 현재 제공하는 문제

| 장 | 문제 | 해설 |
| --- | --- | --- |
| 00. 개발 환경과 학습 준비 | [환경 확인·실행·문제 해결 실습](00-setup/README.md) | [00장 실습 해설](../solutions/00-setup/README.md) |
| 01. 첫 Go 프로그램 | [출력 예측·변형·보고서·오류 진단과 선택 과제](01-first-program/README.md) | [01장 해설](../solutions/01-first-program/README.md) |
| 02. 변수·상수·자료형·연산자 | [값 예측·시간 계산·형 변환·평균·예산 비교](02-variables-types/README.md) | [02장 해설](../solutions/02-variables-types/README.md) |
| 03. 조건문과 반복문 | [분기·반복 추적·끝점 수정·배수 합산·목표 누적](03-control-flow/README.md) | [03장 해설](../solutions/03-control-flow/README.md) |
| 04. 함수와 스코프 | [값 전달·다중 반환·스코프 수정·가변 인자·클로저](04-functions-scope/README.md) | [04장 해설](../solutions/04-functions-scope/README.md) |
| 05. 배열과 슬라이스 | [문제·요구사항·힌트](05-arrays-slices/README.md) | [05장 해설](../solutions/05-arrays-slices/README.md) |
| 06. 맵과 집계 | [문제·요구사항·힌트](06-maps/README.md) | [06장 해설](../solutions/06-maps/README.md) |
| 07. 문자열과 유니코드 | [문제·요구사항·힌트](07-strings-unicode/README.md) | [07장 해설](../solutions/07-strings-unicode/README.md) |
| 08. 포인터와 구조체 | [문제·요구사항·힌트](08-pointers-structs/README.md) | [08장 해설](../solutions/08-pointers-structs/README.md) |
| 09. 메서드와 인터페이스 | [문제·요구사항·힌트](09-methods-interfaces/README.md) | [09장 해설](../solutions/09-methods-interfaces/README.md) |
| 10. 패키지와 모듈 | [문제·요구사항·힌트](10-packages-modules/README.md) | [10장 해설](../solutions/10-packages-modules/README.md) |
| 11. 오류 처리와 자원 정리 | [문제·요구사항·힌트](11-errors-defer/README.md) | [11장 해설](../solutions/11-errors-defer/README.md) |
| 12. 자동 테스트 | [문제·요구사항·힌트](12-testing/README.md) | [12장 해설](../solutions/12-testing/README.md) |
| 13. 제네릭 | [문제·요구사항·힌트](13-generics/README.md) | [13장 해설](../solutions/13-generics/README.md) |
| 14. 파일과 스트림 | [문제·요구사항·힌트](14-files-io/README.md) | [14장 해설](../solutions/14-files-io/README.md) |
| 15. JSON·CSV·시간 다루기 | [문제·요구사항·힌트](15-data-time/README.md) | [15장 해설](../solutions/15-data-time/README.md) |
| 16. 명령줄 프로그램 만들기 | [문제·요구사항·힌트](16-cli/README.md) | [16장 해설](../solutions/16-cli/README.md) |
| 17. Context로 취소와 마감 전달하기 | [문제·요구사항·힌트](17-context/README.md) | [17장 해설](../solutions/17-context/README.md) |
| 18. 고루틴과 완료 대기 | [문제·요구사항·힌트](18-goroutines/README.md) | [18장 해설](../solutions/18-goroutines/README.md) |
| 19. 채널과 select | [문제·요구사항·힌트](19-channels-select/README.md) | [19장 해설](../solutions/19-channels-select/README.md) |
| 20. 공유 상태와 동기화 | [문제·요구사항·힌트](20-sync/README.md) | [20장 해설](../solutions/20-sync/README.md) |
| 21. HTTP 클라이언트 | [문제·요구사항·힌트](21-http-client/README.md) | [21장 해설](../solutions/21-http-client/README.md) |
| 22. HTTP 서버와 라우팅 | [문제·요구사항·힌트](22-http-server/README.md) | [22장 해설](../solutions/22-http-server/README.md) |
| 23. API 설계와 핸들러 테스트 | [문제·요구사항·힌트](23-api-testing/README.md) | [23장 해설](../solutions/23-api-testing/README.md) |
| 24. SQL과 트랜잭션 | [문제·요구사항·힌트](24-sql-transactions/README.md) | [24장 해설](../solutions/24-sql-transactions/README.md) |
| 25. 서비스 구성과 운영 기초 | [문제·요구사항·힌트](25-service-operations/README.md) | [25장 해설](../solutions/25-service-operations/README.md) |
| 26. 성능 측정과 진단 | [문제·요구사항·힌트](26-performance/README.md) | [26장 해설](../solutions/26-performance/README.md) |
| 27. 퍼징과 견고한 테스트 | [문제·요구사항·힌트](27-fuzzing/README.md) | [27장 해설](../solutions/27-fuzzing/README.md) |
| 28. 빌드와 배포 자동화 | [문제·요구사항·힌트](28-build-deploy/README.md) | [28장 해설](../solutions/28-build-deploy/README.md) |
| 29. 종합 실습: CSV 분석 CLI | [문제·요구사항·힌트](29-csv-cli/README.md) | [29장 해설](../solutions/29-csv-cli/README.md) |
| 30. 종합 실습: 할 일 HTTP API | [문제·요구사항·힌트](30-todo-api/README.md) | [30장 해설](../solutions/30-todo-api/README.md) |
| 31. 종합 실습: 동시 배치 처리 | [문제·요구사항·힌트](31-concurrent-batch/README.md) | [31장 해설](../solutions/31-concurrent-batch/README.md) |

전체 32장의 문제와 해설을 제공합니다. [교재 목차](../book/README.md)와 [검증 기록](../docs/VALIDATION.md)을 함께 확인하세요.
