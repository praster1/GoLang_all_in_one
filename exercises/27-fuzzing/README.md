# 27장 연습문제 — 퍼징과 견고한 테스트

[본문](../../book/27-fuzzing/README.md)의 시드·불변 조건·회귀 입력을 적용합니다.
각 문제는 입력 계약, 알려진 정답, 생성 입력에서 검사할 성질을 먼저 정한 뒤 구현합니다.
무작위 입력을 많이 실행했다는 사실만으로 올바름을 판단하지 않습니다.

## 문제 구성

| 문제 | 수준 | 핵심 목표 |
| --- | --- | --- |
| 1 | 확인 | 불변 조건이 놓치는 결함 찾기 |
| 2 | 확인·변형 | 시드와 퍼징 실행의 차이 확인 |
| 3 | 변형 | 룬 단위 접두어의 계약과 불변 조건 |
| 4 | 변형 | 포트 파서의 허용·거부·유효 입력 생성 |
| 5 | 응용 | 태그 정규화의 정렬·중복 제거·멱등성 |
| 6 | 응용 | 정수 덧셈의 범위를 독립 기준과 비교 |
| A | 선택 | 바이트 길이 경계와 회귀 입력 보관 |

## 준비

모든 명령은 저장소 루트에서 실행합니다.
개인 폴더의 기존 파일을 보존하며 새 파일만 준비하세요.
테스트 파일은 직접 작성합니다.

PowerShell:

```powershell
$names = @("rune-prefix", "port", "normalize-tags", "checked-add")
foreach ($name in $names) {
    $dir = "_practice/27-fuzzing/$name"
    New-Item -ItemType Directory -Force $dir | Out-Null
    if (-not (Test-Path "$dir/main.go")) {
        Copy-Item examples/01-first-program/hello/main.go "$dir/main.go"
    }
}
```

Bash:

```sh
for name in rune-prefix port normalize-tags checked-add; do
    dir="_practice/27-fuzzing/$name"
    mkdir -p "$dir"
    if [ ! -e "$dir/main.go" ]; then
        cp examples/01-first-program/hello/main.go "$dir/main.go"
    fi
done
```

일반 테스트는 다음처럼 명시적인 개인 경로를 지정합니다.

```sh
go test -count=1 ./_practice/27-fuzzing/...
```

퍼징은 패키지 하나와 타깃 하나를 지정합니다.
시간은 개발 환경에 맞게 정하되 실제 실행한 값을 기록하세요.

## 문제 1. 조건 하나로는 부족하다

다음 세 퍼즈 타깃은 각각 하나의 조건만 검사합니다.

1. `reverse(reverse(s)) == s`
2. `normalize(normalize(s)) == normalize(s)`
3. `low <= clamp(value, low, high) && clamp(value, low, high) <= high`

### 요구사항

각 조건을 만족하면서도 의도한 동작을 하지 않는 구현을 한 개씩 설명합니다.
그 결함을 잡기 위한 구체적인 입력과 기대 결과도 적습니다.

추가로, 바이트 단위 반전이 한글에서 왜 문제인지 설명합니다.
두 번 반전 조건만으로 그 문제를 찾을 수 있는지도 판단합니다.

### 완료 조건

- 조건을 만족하는 잘못된 구현 세 가지를 제시했습니다.
- 각 결함에 대해 알려진 정답 테스트를 적었습니다.
- 불변 조건과 정답 예제를 함께 사용해야 하는 이유를 설명했습니다.

<details>
<summary>힌트 1</summary>

아무 처리도 하지 않는 구현, 항상 같은 값을 반환하는 구현을 생각해 보세요.

</details>

<details>
<summary>힌트 2</summary>

범위 안의 값을 보정할 때 원래 값을 유지한다는 조건이 빠져 있습니다.

</details>

## 문제 2. 같은 테스트를 두 방식으로 실행

본문의 `examples/27-fuzzing/reverse`를 대상으로 다음 명령을 실행합니다.

```sh
go test -run '^FuzzReverse$' -count=1 ./examples/27-fuzzing/reverse
go test -run '^$' -fuzz '^FuzzReverse$' -fuzztime=3s -parallel=2 ./examples/27-fuzzing/reverse
```

### 요구사항

1. 첫 명령이 검사하는 입력의 출처를 적습니다.
2. 두 번째 명령의 기준선 수집과 새 입력 탐색을 구분합니다.
3. 실행 횟수가 다른 사람과 달라도 실패가 아닌 이유를 설명합니다.
4. `PASS`를 “모든 문자열에서 안전하다”로 해석하면 안 되는 이유를 적습니다.
5. 저장된 `unicode` 코퍼스 하나만 다시 실행합니다.

### 완료 조건

실제로 실행한 명령, Go 버전, 성공·실패 여부를 남깁니다.
콘솔의 처리 횟수를 예상 출력으로 하드코딩하지 않습니다.

<details>
<summary>힌트</summary>

일반 실행은 등록된 시드와 `testdata/fuzz`의 입력을 사용합니다.
새 입력 생성은 `-fuzz`로 선택합니다.
개별 코퍼스는 `FuzzReverse/unicode`라는 하위 이름으로 선택할 수 있습니다.

</details>

## 문제 3. 룬 단위 접두어

다음 함수를 구현합니다.

```go
func prefixRunes(text string, limit int) (string, error)
```

### 입력·반환 계약

- 유효한 UTF-8이고 `limit >= 0`이면 성공합니다.
- 앞에서 최대 `limit`개 룬을 반환합니다.
- 원본 룬 수보다 제한이 크면 원본을 반환합니다.
- `limit=0`이면 빈 문자열입니다.
- 잘못된 UTF-8 또는 음수 제한은 오류입니다.
- 반환값은 화면 글자 덩어리가 아닌 룬 단위 접두어입니다.

| text | limit | 결과 |
| --- | ---: | --- |
| `""` | 0 | `""` |
| `"Go한글"` | 3 | `"Go한"` |
| `"한글"` | 10 | `"한글"` |
| `"Go"` | -1 | 오류 |
| 바이트 `0xff`로 만든 문자열 | 0 | 오류 |

`main`은 `"Go한글"`, 3을 사용합니다.

```text
Go한
```

### 테스트 요구사항

알려진 정답 테스트에 위 표를 포함합니다.
퍼즈 타깃은 문자열과 정수를 받아 다음을 검사합니다.

1. 유효하지 않은 입력은 오류입니다.
2. 유효한 입력의 결과도 유효한 UTF-8입니다.
3. 결과는 원본의 접두어입니다.
4. 결과 룬 수는 `min(limit, 원본 룬 수)`입니다.
5. 같은 제한으로 다시 자르면 결과가 같습니다.

퍼징의 문자열 예산은 65536바이트로 제한해도 됩니다.
이는 함수의 최대 길이 계약과 다르다고 기록하세요.

```sh
go test -run '^$' -fuzz '^FuzzPrefixRunes$' -fuzztime=3s -parallel=2 ./_practice/27-fuzzing/rune-prefix
```

<details>
<summary>힌트 1</summary>

유효성 검사 후 `[]rune`으로 바꾸고, 슬라이싱 전에 제한을 실제 길이로 조정합니다.

</details>

<details>
<summary>힌트 2</summary>

`utf8.RuneCountInString`, `strings.HasPrefix`를 이용할 수 있습니다.
`min`은 두 정수 중 작은 값을 돌려주는 내장 함수입니다.

</details>

## 문제 4. 포트 번호 파서

다음 함수를 구현합니다.

```go
func parsePort(text string) (uint16, error)
```

이번 설정 형식은 ASCII 숫자로만 된 문자열을 받아 1~65535의 포트 번호를 반환합니다.
선행 0은 허용하고 공백·부호·빈 값·0은 거부합니다.
운영체제에서 특정 포트를 실제로 열 수 있는지는 이 함수의 검증 대상이 아닙니다.

| 입력 | 결과 |
| --- | --- |
| `"1"` | 1 |
| `"65535"` | 65535 |
| `"08080"` | 8080 |
| `"0"`, `"65536"`, `""` | 오류 |
| `"+80"`, `" 80"`, `"８０"` | 오류 |

`main`의 출력:

```text
포트: 8080
```

### 퍼즈 타깃 두 개

- `FuzzPort`: 임의 문자열을 파싱하고, 성공 결과를 정규 십진수 문자열로 바꿔 다시 파싱합니다.
- `FuzzValidPort`: `uint16` 입력을 받아 0 외의 값을 정규 문자열로 만들고 반드시 성공하는지 확인합니다.

두 번째 타깃은 정상 입력을 모두 거부하는 결함을 찾기 위한 검사입니다.
한 명령에서는 타깃 하나만 지정합니다.

```sh
go test -run '^$' -fuzz '^FuzzPort$' -fuzztime=3s -parallel=2 ./_practice/27-fuzzing/port
go test -run '^$' -fuzz '^FuzzValidPort$' -fuzztime=3s -parallel=2 ./_practice/27-fuzzing/port
```

### 완료 조건

반드시 성공하는 입력과 반드시 거부하는 입력의 테스트를 모두 작성합니다.
왕복은 선행 0을 보존하는 문자열 왕복이 아니라 숫자 의미의 왕복입니다.

<details>
<summary>힌트</summary>

`strconv.ParseUint(text, 10, 16)`을 사용하되 숫자 문자의 허용 정책도 검사합니다.
0을 거부하는 정책은 변환 함수만으로 해결되지 않습니다.

</details>

## 문제 5. 태그 목록 정규화

다음 함수를 구현합니다.

```go
func normalizeTags(text string) ([]string, error)
```

입력은 쉼표로 구분한 태그입니다.
각 태그의 앞뒤 공백을 제거하고 빈 태그는 생략합니다.
남은 태그에는 ASCII 영문자·숫자·하이픈만 허용합니다.
영문자는 소문자로 바꾸고 중복을 제거한 뒤 오름차순으로 반환합니다.

| 입력 | 결과 |
| --- | --- |
| `""` | 빈 목록 |
| `" , "` | 빈 목록 |
| `"Go,api,GO"` | `[api go]` |
| `"test-1, GO"` | `[go test-1]` |
| `"go lang"`, `"한글"`, `"İ"` | 오류 |

빈 결과의 nil 여부는 계약에 넣지 않습니다.
임의의 Unicode 문자를 소문자로 바꿔 ASCII가 되었더라도 허용하지 않습니다.
태그 원문이 ASCII 허용 문자로만 이루어졌는지 먼저 검사해야 합니다.

`main`은 `" Go,api, GO, test-1 "`을 사용합니다.

```text
[api go test-1]
```

### 검사할 성질

성공 결과는 정렬되어 있고, 빈 태그·중복·대문자가 없어야 합니다.
반환 태그를 쉼표로 합친 뒤 다시 정규화하면 같은 목록이어야 합니다.
알려진 정답 테스트도 함께 작성합니다.

```sh
go test -run '^$' -fuzz '^FuzzTags$' -fuzztime=3s -parallel=2 ./_practice/27-fuzzing/normalize-tags
```

### 완료 조건

- 입력 순서가 달라도 정렬된 같은 결과를 얻습니다.
- 허용하지 않는 원소가 하나라도 있으면 전체 입력을 오류로 처리합니다.
- 빈 목록과 오류를 구분합니다.
- 정규화의 멱등성을 확인합니다.

<details>
<summary>힌트</summary>

`strings.Split`·`TrimSpace`·`ToLower`, 맵, `slices.Sort`를 조합할 수 있습니다.
정렬된 결과에서 이웃 원소를 비교하면 중복이 남았는지 확인할 수 있습니다.

</details>

## 문제 6. 범위를 넘지 않는 정수 덧셈

다음 함수를 구현합니다.

```go
func addChecked(a, b int) (int, error)
```

수학적 합이 현재 플랫폼의 `int` 범위 안이면 합을 반환합니다.
범위를 벗어나면 0과 오버플로 오류를 반환합니다.
이미 오버플로한 `a+b`의 값을 기준으로 판단하지 말고 덧셈 전에 검사합니다.

`maxInt := int(^uint(0) >> 1)`, `minInt := -maxInt - 1`을 사용할 수 있습니다.
여기서 `^uint(0)`은 해당 부호 없는 정수의 모든 비트를 1로 만들고,
한 비트를 오른쪽으로 이동하면 대응하는 부호 있는 정수의 최댓값을 얻습니다.

| 입력 | 결과 |
| --- | --- |
| 0, 0 | 0 |
| -1, 1 | 0 |
| maxInt, 0 | maxInt |
| maxInt, 1 | 오류 |
| minInt, -1 | 오류 |
| minInt, maxInt | -1 |

`main`은 최대 정수에 1을 더한 오류와 10+20의 결과를 출력합니다.

```text
범위 오류: true
합계: 30
```

### 독립 기준으로 검사하기

퍼즈 타깃에서는 표준 라이브러리 `math/big.Int`로 정확한 합을 계산할 수 있습니다.
`big.NewInt`는 `int64`를 큰 정수로 만들고,
`new(big.Int).Add(x, y)`는 두 값의 합을 새 결과에 저장합니다.
`Cmp`는 작음·같음·큼에 따라 음수·0·양수를 돌려줍니다.

현재 Go의 `int`는 32비트 또는 64비트이므로 각 입력을 `int64`로 옮길 수 있습니다.
정확한 합을 minInt·maxInt와 비교한 뒤 결과가 범위 안일 때만 `Int64`로 꺼냅니다.
[math/big.Int](https://pkg.go.dev/math/big#Int)의 API를 참고합니다.

```sh
go test -run '^$' -fuzz '^FuzzAddChecked$' -fuzztime=3s -parallel=2 ./_practice/27-fuzzing/checked-add
```

### 완료 조건

- 양수 방향과 음수 방향 오버플로를 모두 검사합니다.
- 단위 테스트에 정수 극단값을 직접 넣습니다.
- 검사 함수가 구현과 같은 오버플로 식을 그대로 복사하지 않습니다.
- 오류가 없는 결과를 독립적인 정확한 합과 비교합니다.

<details>
<summary>힌트 1</summary>

b가 양수일 때 `a > maxInt-b`이면 덧셈 결과가 범위를 넘습니다.
음수일 때는 아래쪽 경계를 생각합니다.

</details>

<details>
<summary>힌트 2</summary>

b가 음수일 때 `a < minInt-b`를 검사할 수 있습니다.
해당 분기에서 경계 계산 자체가 범위를 넘지 않는지도 확인하세요.

</details>

## 선택 A. 길이 경계와 회귀 코퍼스

본문의 `parseRecord`에 전체 길이 4096과 4097인 입력을 추가로 검사합니다.
`a=`가 이미 두 바이트라는 점을 반영하세요.

그 뒤 본문의 개인 바이트 반전 결함을 재현하고 다음을 남깁니다.

1. 실패 입력과 위반한 계약.
2. 실패를 다시 실행하는 명령.
3. 수정한 코드와 수정 이유.
4. 회귀 입력의 저장 위치 또는 `f.Add`에 추가한 값.
5. 수정 후 일반 테스트와 제한 시간 퍼징 결과.

자동 생성된 실패 입력의 이름을 임의로 추정하지 말고 도구 출력을 사용합니다.
개인 폴더의 실패 파일을 기본 예제 폴더에 그대로 옮기지 않습니다.

## 완료 체크리스트

- [ ] 알려진 정답과 불변 조건을 함께 검사했습니다.
- [ ] 유효한 입력을 전부 거부하는 결함을 고려했습니다.
- [ ] 입력 단위가 바이트인지 룬인지 명시했습니다.
- [ ] 타깃 하나마다 시간과 작업자 수를 정했습니다.
- [ ] 공유 전역 상태나 외부 API에 의존하지 않습니다.
- [ ] 실패를 숨기기 위한 무조건적인 recover를 넣지 않았습니다.
- [ ] 결과를 “이번 탐색 범위에서 위반 미발견”으로 해석했습니다.

## 참고 자료

[Go Fuzzing](https://go.dev/doc/security/fuzz/) · [testing.F](https://pkg.go.dev/testing#F) · [math/big](https://pkg.go.dev/math/big)

[본문](../../book/27-fuzzing/README.md) · [해설 확인](../../solutions/27-fuzzing/README.md)
