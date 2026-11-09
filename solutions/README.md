# 해설 안내

문제를 먼저 풀고 힌트를 확인한 뒤 해설을 읽습니다. 해설은 가능한 풀이와 그 이유를 설명합니다. 요구 사항을 만족하는 다른 구현도 정답이 될 수 있습니다.

## 장별 해설

| 장 | 문제 | 해설의 내용 |
| --- | --- | --- |
| 00. 개발 환경과 학습 준비 | [00장 실습](../exercises/00-setup/README.md) | [환경 기록, 실행·빌드, 경로 오류 해설](00-setup/README.md) |
| 01. 첫 Go 프로그램 | [01장 연습문제](../exercises/01-first-program/README.md) | [문제 1~5와 선택 A 해설](01-first-program/README.md) |
| 02. 변수·상수·자료형·연산자 | [02장 연습문제](../exercises/02-variables-types/README.md) | [문제 1~6과 선택 A 해설](02-variables-types/README.md) |
| 03. 조건문과 반복문 | [03장 연습문제](../exercises/03-control-flow/README.md) | [문제 1~6과 선택 A 해설](03-control-flow/README.md) |
| 04. 함수와 스코프 | [04장 연습문제](../exercises/04-functions-scope/README.md) | [문제 1~6과 선택 A 해설](04-functions-scope/README.md) |

## 실행 가능한 참고 구현

| 과제 | 코드 | 검증 |
| --- | --- | --- |
| 01장 문제 4: 학습 기록 보고서 | [report/main.go](01-first-program/report/main.go) | [출력 테스트](01-first-program/report/main_test.go) |
| 01장 선택 A: 이름에 맞는 인사말 | [greeting/main.go](01-first-program/greeting/main.go) | [입력별 결과와 출력 테스트](01-first-program/greeting/main_test.go) |
| 02장 문제 2: 학습 시간 기록 | [study-log/main.go](02-variables-types/study-log/main.go) | [출력 테스트](02-variables-types/study-log/main_test.go) |
| 02장 문제 3: 시·분·초 계산 | [time-parts/main.go](02-variables-types/time-parts/main.go) | [출력 테스트](02-variables-types/time-parts/main_test.go) |
| 02장 문제 5: 평균 | [average/main.go](02-variables-types/average/main.go) | [출력 테스트](02-variables-types/average/main_test.go) |
| 02장 문제 6: 구매 내역 | [receipt/main.go](02-variables-types/receipt/main.go) | [출력 테스트](02-variables-types/receipt/main_test.go) |
| 03장 문제 2: 점수 안내 | [grade/main.go](03-control-flow/grade/main.go) | [출력 테스트](03-control-flow/grade/main_test.go) |
| 03장 문제 3: 신호 안내 | [signal/main.go](03-control-flow/signal/main.go) | [출력 테스트](03-control-flow/signal/main_test.go) |
| 03장 문제 5: 배수 합산 | [multiples/main.go](03-control-flow/multiples/main.go) | [출력 테스트](03-control-flow/multiples/main_test.go) |
| 03장 문제 6: 목표 학습 시간 | [goal/main.go](03-control-flow/goal/main.go) | [출력 테스트](03-control-flow/goal/main_test.go) |
| 04장 문제 2: 점수 판정 함수 | [grade/main.go](04-functions-scope/grade/main.go) | [함수·출력 테스트](04-functions-scope/grade/main_test.go) |
| 04장 문제 4: 짝수 합계 | [sum-even/main.go](04-functions-scope/sum-even/main.go) | [함수·출력 테스트](04-functions-scope/sum-even/main_test.go) |
| 04장 문제 5: 학습 계획 | [study-plan/main.go](04-functions-scope/study-plan/main.go) | [함수·출력 테스트](04-functions-scope/study-plan/main_test.go) |
| 04장 문제 6: 상태를 기억하는 함수 | [stepper/main.go](04-functions-scope/stepper/main.go) | [함수·출력 테스트](04-functions-scope/stepper/main_test.go) |

실행 위치와 명령, 예상 결과는 각 장의 해설에 있습니다. 참고 구현을 실행한 결과와 개인 풀이를 실행한 결과를 구분해서 확인하세요.

## 해설을 추가할 때

해당 문제 링크, 실행 명령, 기대하는 동작, 핵심 선택 이유를 함께 기록합니다. 테스트는 요구 사항과 중요한 경계 조건을 확인합니다. 자세한 기준은 [집필 가이드](../docs/AUTHORING.md)를 따릅니다.
