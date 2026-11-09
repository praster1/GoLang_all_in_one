# 29장 연습문제 — 종합 실습: CSV 분석 CLI

[본문](../../book/29-csv-cli/README.md)에서 만든 분석기를 확장합니다.
CSV 문법, 업무 규칙, 집계, 출력과 명령행 경계를 각각 확인하세요.
기존 `report` 패키지를 재사용하되 추가 기능의 계약을 먼저 정합니다.

## 문제 구성

| 문제 | 수준 | 결과물 |
| --- | --- | --- |
| 1 | 확인 | 집계 결과와 빈 입력 정책 설명 |
| 2 | 확인·변형 | 오류 위치·입력 한도·출력 정책 검사 |
| 3 | 변형 | 분류 필터와 일관된 요약 |
| 4 | 변형 | 합계순 상위 그룹 선택 |
| 5 | 응용 | 여러 CSV 입력의 집계 병합 |
| 6 | 응용 | 검증 후 완성된 출력 바이트 반환 |
| A | 선택 | 실제 실행 파일의 입출력·종료 코드 기록 |

문제 3~6의 참고 구현은 [해설](../../solutions/29-csv-cli/README.md)에 있습니다.
먼저 직접 작성하고, 관찰 가능한 결과와 경계를 비교하세요.
함수 이름이 같은 것보다 입력 계약과 오류 처리가 일치하는 것이 중요합니다.

## 개인 파일 준비

명령은 저장소 루트에서 실행합니다.
아래 준비 명령은 이미 작성한 `main.go`를 덮어쓰지 않습니다.

PowerShell:

```powershell
$names = @("filter", "ranking", "merge", "render")
foreach ($name in $names) {
    $dir = "_practice/29-csv-cli/$name"
    New-Item -ItemType Directory -Force $dir | Out-Null
    if (-not (Test-Path "$dir/main.go")) {
        Copy-Item examples/01-first-program/hello/main.go "$dir/main.go"
    }
}
```

Bash:

```sh
for name in filter ranking merge render; do
    dir="_practice/29-csv-cli/$name"
    mkdir -p "$dir"
    if [ ! -e "$dir/main.go" ]; then
        cp examples/01-first-program/hello/main.go "$dir/main.go"
    fi
done
```

`main_test.go`는 각 폴더에 직접 작성합니다.
같은 모듈 안이므로 새 `go.mod`를 만들지 않습니다.
공유 패키지는 다음 경로로 가져옵니다.

```go
import "github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
```

본문의 입력 계약은 그대로 유지합니다.
분류는 공백 정리 후 1~40룬, 금액은 0~1000000000의 십진 정수입니다.
분석기 기본 한도는 입력 1 MiB와 데이터 레코드 10000개입니다.
분류 이름은 정규화·대소문자 변환 없이 비교합니다.

## 문제 1. 집계와 빈 입력의 의미

다음 입력의 기본 CSV 출력과 JSON의 Rows·Total 값을 예측하세요.

```csv
category,amount
도서,12000
식품,5000
도서,8000
문구,0
```

### 확인할 내용

1. 데이터 레코드 수와 그룹 수를 각각 적습니다.
2. 전체 금액과 그룹별 금액을 계산합니다.
3. 출력 그룹의 순서를 정합니다.
4. 금액 0인 행이 그룹의 Count에 포함되는지 설명합니다.
5. 같은 분류가 떨어져 있어도 하나로 합쳐지는 이유를 설명합니다.
6. 기본 CSV 출력에 전체 합계 행이 있는지 확인합니다.

### 추가 입력

아래 네 입력의 성공·실패와 결과를 구분하세요.
문자열의 `\n`은 실제 줄바꿈을 뜻합니다.

| 입력 | 생각할 점 |
| --- | --- |
| `category,amount\n` | 헤더만 있을 때 |
| 빈 문자열 | 헤더 자체가 없을 때 |
| `amount,category\n` | 같은 이름이 다른 순서일 때 |
| `category,amount\n"도서,문구",10\n` | 인용 필드에 쉼표가 있을 때 |

### 완료 조건

Rows를 그룹 수와 혼동하지 않아야 합니다.
허용된 0과 금액 변환 실패를 구분해야 합니다.
출력 순서의 기준을 “실행하면 우연히 이렇게 나옴”으로 설명하지 않습니다.
헤더만 있는 성공 결과에서 JSON groups가 `[]`인지 확인합니다.

<details>
<summary>힌트</summary>

먼저 각 데이터 레코드 옆에 분류·금액을 적습니다.
그다음 같은 분류끼리 Count와 Total을 합칩니다.
마지막 단계에서 이름으로 정렬하세요.
CSV 인용 규칙은 `strings.Split`이 처리하지 않습니다.

</details>

## 문제 2. 오류가 발생하는 경계 찾기

아래 사례를 입력 오류, 한도 오류, 명령 사용법 오류로 분류합니다.
CLI에서 표준 출력·표준 오류·종료 코드가 어떻게 달라지는지도 적으세요.

| 사례 | 조건 |
| --- | --- |
| 잘못된 금액 | `category,amount\n도서,bad\n` |
| 음수 금액 | `category,amount\n도서,-1\n` |
| 빈 분류 | `category,amount\n  ,1\n` |
| 열 부족 | `category,amount\n도서\n` |
| 행 수 초과 | 데이터 2개에 `-max-rows 1` |
| 바이트 초과 | 실제 입력 길이보다 1 작은 `-max-bytes` |
| 잘못된 한도 | `-max-rows 0` |
| 지원하지 않는 출력 | `-format xml` |

### 직접 검사할 함수

```go
summary, err := report.Analyze(reader, report.DefaultOptions())
```

- 오류가 있으면 성공한 앞부분의 Summary가 반환되는지 확인합니다.
- `errors.Is(err, report.ErrInput)`과 `errors.Is(err, report.ErrLimit)`을 구분합니다.
- `errors.As`로 `*report.DataError`를 찾습니다.
- DataError의 Record와 Field가 어떤 위치를 뜻하는지 설명합니다.

### 반드시 포함할 경계

1. 입력 길이와 MaxBytes가 정확히 같은 경우.
2. 그 입력에서 MaxBytes를 1 줄인 경우.
3. 데이터 수와 MaxRows가 정확히 같은 경우.
4. 데이터가 한 개 더 있는 경우.
5. 첫 행은 정상이지만 뒤의 행에 오류가 있는 경우.
6. Reader가 입력 도중 읽기 오류를 반환하는 경우.

### 완료 조건

최대 허용값과 초과값을 서로 다른 테스트로 검사합니다.
후반부 오류에서도 부분 집계를 성공 결과로 내보내지 않습니다.
오류 메시지에 든 레코드 번호를 파일의 물리적 줄 번호로 단정하지 않습니다.
한글 입력에서 바이트 수와 룬 수가 다르다는 점을 반영합니다.

<details>
<summary>힌트</summary>

MaxBytes는 `len(input)`으로 계산할 수 있습니다.
이는 문자열의 바이트 수이며 분류 길이 제한의 룬 수와 다릅니다.
`DataError`는 입력 오류의 위치를 설명하고,
기본 오류 분류는 래핑된 sentinel을 통해 확인합니다.

</details>

## 문제 3. 분류 하나를 선택한 보고서

다음 함수를 작성합니다.

```go
func selectCategory(summary report.Summary, category string) report.Summary
```

### 입력 계약

- summary는 `report.Analyze`가 성공한 결과입니다.
- category는 일치 여부를 판단할 이름이며 함수 내부에서 공백을 정리하지 않습니다.
- 비교는 정확한 문자열 일치입니다.
- 원본 summary와 그 Groups를 수정하지 않습니다.

### 출력 계약

- 일치하는 그룹만 새 Groups에 넣습니다.
- Rows와 Total은 선택한 그룹에서 다시 계산합니다.
- 일치하는 그룹이 없으면 Rows와 Total은 0, Groups는 빈 비-nil 슬라이스입니다.
- 결과 그룹의 필드를 수정해도 원본 그룹 값은 유지됩니다.

아래 자료에서 도서만 선택해 CSV로 출력합니다.

```csv
category,amount
도서,12
식품,5
도서,8
```

예상 출력:

```text
category,count,total
도서,2,20
```

### 테스트 기준

| 선택 | Rows | Total | 그룹 수 |
| --- | ---: | ---: | ---: |
| 도서 | 2 | 20 | 1 |
| 식품 | 1 | 5 | 1 |
| 없음 | 0 | 0 | 0 |

결과 Groups의 값을 바꾼 뒤 원본이 그대로인지도 검사합니다.
`Summary`가 값으로 전달된다는 사실만으로 원본 보호가 보장되는 것은 아닙니다.

### 완료 조건

필터 뒤에도 전체 요약과 그룹 요약의 의미가 일치해야 합니다.
선택하지 않을 분류의 잘못된 행을 조용히 무시하는 기능으로 바꾸지 않습니다.
입력 전체를 먼저 분석한 후 결과를 선택하세요.

<details>
<summary>힌트</summary>

새 Summary와 새 Groups를 만들어 시작합니다.
조건을 통과한 Group 값을 추가하면서 Count와 Total을 누적하세요.

</details>

## 문제 4. 합계순 상위 N개 그룹

다음 함수를 작성합니다.

```go
func topGroups(summary report.Summary, n int) ([]report.Group, error)
```

### 정렬·개수 계약

1. Total이 큰 그룹이 먼저 옵니다.
2. Total이 같으면 Category의 문자열 오름차순입니다.
3. n이 0이면 빈 결과입니다.
4. n이 그룹 수보다 크면 모든 그룹을 반환합니다.
5. n이 음수이면 오류입니다.
6. 원본 Groups의 순서와 필드 값은 유지합니다.

같은 합계일 때 입력 순서나 맵 순회 순서에 의존하지 마세요.
비교 함수에서 두 int64를 빼고 int로 변환하는 방식은 피합니다.
대소 비교만으로 필요한 순서를 표현할 수 있습니다.

### 실행 예

아래 입력에서 상위 두 그룹을 선택합니다.

```csv
category,amount
도서,12
식품,20
도서,8
```

예상 출력:

```text
1. 도서: 20
2. 식품: 20
```

### 테스트 기준

- 합계가 서로 다른 세 그룹.
- 합계가 같고 이름 순서가 뒤집힌 두 그룹.
- n=0, n=1, n=그룹 수, n=그룹 수+1, n=-1.
- 그룹이 없는 입력.
- 결과의 Group을 바꿔도 원본 값이 유지되는지.

### 완료 조건

정렬 결과와 원본 보존을 모두 검사합니다.
이 함수가 반환하는 것은 그룹 목록입니다.
원래 Summary의 Rows·Total을 “선택된 상위 그룹만의 합계”라고 표시하지 않습니다.

<details>
<summary>힌트</summary>

`Groups`를 복사한 후 정렬합니다.
합계 비교가 같을 때만 이름을 비교하세요.
슬라이스를 잘라내기 전에 n의 범위를 조정합니다.

</details>

## 문제 5. 여러 CSV 입력 병합

다음 함수를 작성합니다.

```go
func mergeInputs(inputs []io.Reader) (report.Summary, error)
```

### 입력 계약

- 한 번에 최대 10개의 Reader를 받습니다.
- 각 Reader는 서로 독립적인 CSV이며 각자 헤더가 있어야 합니다.
- 각 입력에는 `report.DefaultOptions()`의 한도를 적용합니다.
- Reader는 호출자가 소유합니다. 이 함수는 Close하지 않습니다.
- nil 인터페이스 원소는 오류입니다. 유효한 Reader 구현을 전달해야 합니다.

### 병합 정책

1. 각 입력을 분석하고 같은 분류의 Count·Total을 합칩니다.
2. 마지막에 Category 이름으로 정렬합니다.
3. 모든 입력이 성공했을 때만 Summary를 반환합니다.
4. 어느 입력이 실패했는지 1부터 시작하는 입력 번호를 오류에 덧붙입니다.
5. `%w` 래핑으로 원래 오류의 분류를 보존합니다.
6. 입력 목록 자체가 비어 있으면 빈 성공 결과를 반환합니다.
7. 결과의 Groups는 빈 경우에도 비-nil입니다.

### 실행 예

첫 입력:

```csv
category,amount
도서,12
```

둘째 입력:

```csv
category,amount
식품,7
도서,8
```

예상 출력:

```text
category,count,total
도서,2,20
식품,1,7
```

### 테스트 기준

- 두 입력에 같은 분류가 있는 정상 병합.
- 헤더만 있는 입력과 정상 입력의 조합.
- 빈 Reader 목록.
- 두 번째 입력이 잘못되었을 때 부분 결과를 반환하지 않음.
- 오류를 감싼 뒤에도 `errors.Is(err, report.ErrInput)`이 참임.
- 입력 10개와 11개의 경계.
- nil 인터페이스 원소를 전달한 경우.

### 완료 조건

입력 개수 제한과 개별 입력 한도를 함께 설명합니다.
이 제한에서 총 Rows와 Total이 자료형 범위 안인지 계산합니다.
여러 입력의 바이트를 단순 연결하지 않습니다. 중간 헤더는 데이터가 아닙니다.

<details>
<summary>힌트</summary>

분류 이름을 키로 하는 새 맵에 Group 값을 누적합니다.
분석 도중 실패하면 지금까지 만든 결과도 반환하지 않습니다.
Readers를 열고 닫는 책임은 이 함수 바깥에 두세요.

</details>

## 문제 6. 완성된 출력 바이트 만들기

다음 함수를 작성합니다.

```go
func render(input io.Reader, format string) ([]byte, error)
```

### 처리 계약

- format은 csv 또는 json입니다.
- 지원하지 않는 format은 입력을 읽기 전에 오류로 반환합니다.
- 분석에는 기본 입력 한도를 사용합니다.
- 분석이 실패하면 반환 바이트는 nil입니다.
- 출력은 먼저 `bytes.Buffer`에 완성합니다.
- 전체 직렬화가 성공하면 완성된 바이트를 반환합니다.
- 호출자가 준 Reader는 닫지 않습니다.

### 테스트 기준

CSV에서는 헤더·정렬·개행을 포함한 결과를 비교합니다.
JSON에서는 역직렬화한 Rows·Total·Groups의 의미를 검사합니다.
잘못된 입력에서는 오류와 nil 출력을 함께 확인합니다.
읽기 횟수를 기록하는 Reader로 잘못된 format이 입력을 소비하지 않는지 검사합니다.

### 완료 조건

이 함수는 완성된 바이트를 만드는 경계입니다.
그 바이트를 실제 파일에 쓸 때 발생하는 오류까지 해결한다고 설명하지 않습니다.
메모리에 결과를 모으는 비용과 외부 Writer에 즉시 쓰는 방식의 차이를 설명하세요.

<details>
<summary>힌트</summary>

format 검사 → Analyze → Buffer → WriteCSV 또는 WriteJSON 순서를 사용합니다.
외부 출력에 전달하는 시점은 함수가 성공한 뒤로 미룰 수 있습니다.

</details>

## 선택 A. 실제 바이너리의 결과 기록

본문의 빌드 명령으로 CSV CLI 실행 파일을 만듭니다.
이름이 겹치는 기존 산출물이 있는지 확인하고 전용 개인 경로를 사용하세요.

다음 조건을 실제 실행 파일로 검사합니다.

| 실행 | 확인할 결과 |
| --- | --- |
| 정상 파일을 CSV로 출력 | 데이터 stdout, 종료 0 |
| 정상 파일을 JSON으로 출력 | 해석 가능한 JSON, 종료 0 |
| 표준 입력으로 같은 자료 전달 | 파일 입력과 같은 결과 |
| 잘못된 금액 파일 | stdout 비어 있음, stderr 진단, 종료 1 |
| 없는 파일 | 읽기 오류, 종료 1 |
| `-format xml` | 사용법 오류, 종료 2 |
| `-h` | 도움말, 종료 0 |

명령·환경·실제 종료 코드를 짧게 기록합니다.
`go run`이 중간 실행기라는 점과 빌드한 바이너리의 종료 코드 확인을 구분하세요.
셸 리디렉션은 프로그램 시작 전에 파일을 만들거나 비울 수 있습니다.
“입력 오류 시 stdout이 비어 있다”는 계약만으로 기존 결과 파일이 보존되지는 않습니다.
기존 파일을 보호하려면 임시 파일 작성·성공 확인·교체 과정을 별도로 설계합니다.

## 최종 확인

저장소의 참고 구현은 다음 명령으로 검사할 수 있습니다.

```sh
go test -count=1 ./examples/29-csv-cli/... ./solutions/29-csv-cli/...
```

개인 코드를 검사하려면 명시적인 패키지 경로를 사용합니다.
Go의 `./...` 탐색은 밑줄로 시작하는 `_practice` 폴더를 자동으로 포함하지 않습니다.

```sh
go test ./_practice/29-csv-cli/filter ./_practice/29-csv-cli/ranking ./_practice/29-csv-cli/merge ./_practice/29-csv-cli/render
```

- [ ] 입력 규칙과 출력 규칙을 먼저 적었습니다.
- [ ] 정상·빈 입력·경계·실패 사례를 구분했습니다.
- [ ] 요약 값과 그룹 값이 일치합니다.
- [ ] 정렬과 필터가 원본을 바꾸지 않습니다.
- [ ] 읽기·쓰기·닫기 책임을 설명할 수 있습니다.
- [ ] 실제 실행 파일의 종료 코드와 진단 위치를 확인했습니다.

[본문으로](../../book/29-csv-cli/README.md) · [힌트와 해설](../../solutions/29-csv-cli/README.md) · [연습문제 목록](../README.md)
