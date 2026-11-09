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

개인 코드는 저장소 루트의 `/_practice/` 아래에 둡니다. 각 장의 문제 문서에 **기존 파일을 보존하는 준비 명령**이 있습니다. 자신의 PowerShell 또는 Bash/Zsh에 맞는 블록을 선택하세요.

| 장 | 준비 안내 | 개인 파일의 위치 |
| --- | --- | --- |
| 00. 개발 환경과 학습 준비 | [00장 실습 준비](00-setup/README.md) | `_practice/00-setup/main.go` |
| 01. 첫 Go 프로그램 | [01장 문제별 준비](01-first-program/README.md) | `_practice/01-first-program/` 아래의 `message/`, `report/`, `repair/`; 선택 과제는 `greeting/` |
| 02. 변수·상수·자료형·연산자 | [02장 문제별 준비](02-variables-types/README.md) | `_practice/02-variables-types/` 아래의 `study-log/`, `time-parts/`, `repair/`, `average/`, `receipt/`; 선택 과제는 `iota/` |
| 03. 조건문과 반복문 | [03장 문제별 준비](03-control-flow/README.md) | `_practice/03-control-flow/` 아래의 `grade/`, `signal/`, `repair/`, `multiples/`, `goal/`; 선택 과제는 `fizzbuzz/` |
| 04. 함수와 스코프 | [04장 문제별 준비](04-functions-scope/README.md) | `_practice/04-functions-scope/` 아래의 `grade/`, `repair/`, `sum-even/`, `study-plan/`, `stepper/`; 선택 과제는 `apply-twice/` |

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

새로운 장의 문제는 해당 장이 작성될 때 추가합니다. [전체 교재 목차](../book/README.md)에서 집필 상태를 확인할 수 있습니다.
