# 11장 연습문제: 오류 처리와 자원 정리

[본문](../../book/11-errors-defer/README.md)을 읽고 풉니다. 이 장의 완료 조건에는 정상 출력뿐 아니라 오류 종류, 실패 시 결과와 정리 횟수도 포함됩니다.

## 풀이 준비

개인 코드는 `_practice/11-errors-defer/` 아래 `predict`, `port`, `lookup`, `cleanup`, `validation` 폴더에 나누어 작성합니다. 각 폴더의 main.go는 하나의 독립 프로그램입니다. 새 go.mod나 go.work는 만들지 않습니다. 기존 파일이 있으면 덮어쓰지 않습니다.

모든 명령은 저장소 루트에서 실행합니다.

```sh
go run ./_practice/11-errors-defer/port
```

12장 이전이므로 테스트 작성은 필수가 아닙니다. main에서 정상·경계·오류 입력을 넣고 오류를 errors.Is/As로 검사해도 됩니다. 참고 해설에는 실행 가능한 코드와 자동 테스트가 있습니다.

계약을 먼저 적어 보세요.

| 항목 | 직접 적을 내용 |
| --- | --- |
| 성공 | 반환값과 상태 |
| 실패 | 오류 종류와 무효 결과 |
| 원인 | 래핑 안에 보존할 정보 |
| 자원 | 누가 언제 닫는지 |
| 중복 처리 | 실패 시 다시 실행해도 되는지 |

## 문제 1. 반환값과 defer의 순서

**수준: 확인**

다음 전체 프로그램의 세 줄 출력을 실행 전에 적습니다.

```go
package main

import "fmt"

func calculate() (result int) {
    result = 3
    defer fmt.Println("저장된 결과:", result)
    defer func() {
        result += 2
        fmt.Println("정리 중 결과:", result)
    }()
    return 10
}

func main() {
    fmt.Println("최종:", calculate())
}
```

다음 표를 채웁니다.

| 시점 | result | 이미 저장된 defer 인자 |
| --- | ---: | --- |
| result = 3 직후 | | |
| 첫 defer 등록 직후 | | |
| return 10의 반환 변수 대입 뒤 | | |
| 둘째 defer 실행 뒤 | | |
| 호출자에게 전달될 때 | | |

질문:

- 첫 Println이 3과 10과 12 중 무엇을 출력하나요?
- 두 번째 등록된 defer는 어느 result를 수정하나요?
- main의 Println은 calculate의 defer보다 먼저 실행되나요?

완료 조건: 마지막 숫자만 맞히지 말고, 반환식 평가·반환 변수 대입·defer 실행·호출자 전달 순서를 설명합니다.

<details>
<summary>힌트 1</summary>

첫 defer의 인자는 등록할 때 평가됩니다. 둘째 defer의 클로저 본문은 실제 정리 시점에 실행됩니다.

</details>

<details>
<summary>힌트 2</summary>

이름 붙인 반환 변수는 return 식의 값으로 먼저 설정됩니다. 그 뒤 defer가 값을 바꿀 수 있습니다.

</details>

## 문제 2. 포트의 형식과 범위 구별

**수준: 변형**

문자열 포트를 정수로 바꾸는 함수를 만듭니다.

```go
var ErrPortRange = errors.New("port must be 1..65535")
func parsePort(text string) (int, error)
```

계약:

- 숫자 변환은 strconv.Atoi를 사용합니다.
- 변환 실패는 원래 오류를 `%w`로 감싸 반환합니다.
- 숫자 변환 뒤 1~65,535 범위를 검사합니다.
- 범위 실패는 errors.Is로 ErrPortRange를 찾을 수 있어야 합니다.
- 실패한 결과 정수는 0입니다.
- 성공하면 포트와 nil을 반환합니다.
- 앞뒤 공백 제거는 하지 않습니다. Atoi의 문자열 처리 계약을 그대로 사용합니다.

| 입력 | 기대 |
| --- | --- |
| `"1"` | 1, nil |
| `"65535"` | 65535, nil |
| `"8080"` | 8080, nil |
| `"0"`, `"65536"`, `"-1"` | 범위 오류 |
| `"abc"`, 빈 문자열, `" 80"` | 숫자 변환 오류 |
| int에 담을 수 없을 정도로 큰 숫자 | 숫자 변환 오류 |

완료 조건: 범위 오류를 문자열 비교로 판정하지 않습니다. 형식 오류는 `var detail *strconv.NumError`에 errors.As로 찾아봅니다.

<details>
<summary>힌트 1</summary>

변환 실패부터 반환하면 이후 코드는 숫자가 있다는 전제에서 범위를 검사할 수 있습니다.

</details>

<details>
<summary>힌트 2</summary>

`fmt.Errorf("parse port: %v", err)`는 원인을 문자열로 표시할 뿐 래핑 관계를 보존하지 않습니다. 포맷 동사를 확인하세요.

</details>

## 문제 3. 오류 처리 코드 검토

**수준: 변형**

다음은 서로 독립된 잘못된 발췌입니다. 정상 예제에 붙이지 말고, 문제와 수정 방향을 적습니다.

### A. 같은 메시지로 비교

```go
if errors.Is(err, errors.New("not found")) {
    fmt.Println("없음")
}
```

ErrNotFound라는 대표 오류가 이미 정의되어 있다고 가정합니다. 왜 새 오류를 만들어서는 분류가 안 되나요?

### B. 반환 오류 가리기

```go
func work() (err error) {
    defer func() {
        err := closeResource()
        if err != nil {
            fmt.Println(err)
        }
    }()
    return nil
}
```

closeResource는 설명용 함수이며 실제 저장소 함수가 아닙니다. 이 코드는 왜 닫기 오류를 호출자에게 반환하지 않나요? 단순 출력과 오류 전달은 어떻게 다른가요?

### C. typed nil 반환

```go
func validate() error {
    var detail *InputError
    return detail
}
```

InputError가 error를 구현한다고 가정합니다. 성공 시 어떤 값을 반환해야 하나요?

### D. defer 등록 전에 자원 획득 확인 누락

```go
file, err := os.Open(path)
defer file.Close()
if err != nil {
    return err
}
```

일반적인 권장 순서를 설명하세요. 이 문제에서는 실제 파일을 열 필요가 없습니다.

완료 조건: 오류 값의 정체성·스코프·인터페이스·자원 수명을 각각 구별하여 설명합니다.

<details>
<summary>힌트</summary>

대표 오류는 이미 존재하는 같은 값을 사용합니다. 이름 붙인 반환 변수를 수정하려면 새 지역 변수를 만들지 않습니다. 자원은 획득 성공을 확인한 뒤 정리 책임을 등록합니다.

</details>

## 문제 4. 어떤 키의 조회가 실패했는지 보존

**수준: 응용**

문자열 키에서 정수를 읽는 함수를 작성합니다.

```go
var ErrNotFound = errors.New("not found")
type LookupError struct {
    Key string
    Cause error
}
func lookup(values map[string]int, key string) (int, error)
```

LookupError의 Error와 Unwrap 메서드를 구현하세요.

계약:

- 키가 있으면 저장된 값과 nil입니다. 저장된 값 0도 정상입니다.
- 키가 없으면 0과 `*LookupError`를 반환합니다.
- LookupError는 찾던 Key와 원인 ErrNotFound를 저장합니다.
- errors.Is로 ErrNotFound를 확인할 수 있어야 합니다.
- errors.As로 LookupError를 찾아 Key를 읽을 수 있어야 합니다.
- nil 맵에서도 없는 키 오류를 정상 반환하며 패닉이 나지 않습니다.
- 입력 맵은 변경하지 않습니다.

완료 조건: `{ "zero": 0 }`의 zero 조회와 missing 조회를 구별합니다. 바깥에서 추가로 `%w`로 감싼 뒤에도 같은 원인·타입 검사가 가능한지 확인합니다.

<details>
<summary>힌트 1</summary>

맵의 두 결과 조회를 사용합니다. 값이 0인지 비교하면 존재하는 0을 잘못 실패로 분류합니다.

</details>

<details>
<summary>힌트 2</summary>

Unwrap은 원인 오류를 반환합니다. Error는 문장을 만들며, 그 문장을 다시 잘라 Key를 추출하지 않습니다.

</details>

## 문제 5. 작업과 자원 정리의 두 오류

**수준: 응용**

다음 함수를 구현합니다.

```go
func withResource(
    resource io.Closer,
    operation func(io.Closer) error,
) (err error)
```

입력 resource는 nil이 아닌 유효한 객체이고 operation은 nil이 아니며 정상적으로 반환한다고 가정합니다. 이 문제에서 모든 panic을 복구하지 않습니다.

계약:

- resource의 소유권을 함수가 넘겨받습니다.
- operation을 한 번 호출하며 같은 resource를 전달합니다.
- operation의 성공·실패와 관계없이 Close를 정확히 한 번 호출합니다.
- operation 오류에는 `operation` 문맥을 더합니다.
- Close 오류에는 `close` 문맥을 더합니다.
- 둘 다 실패하면 errors.Is로 두 원인 모두를 찾을 수 있어야 합니다.
- 둘 다 성공하면 nil입니다.

검사할 네 조합:

| operation | Close | 기대 |
| --- | --- | --- |
| 성공 | 성공 | nil |
| 실패 A | 성공 | A 보존 |
| 성공 | 실패 B | B 보존 |
| 실패 A | 실패 B | A와 B 모두 보존 |

테스트 입력을 위해 `Close`가 호출될 때 카운터를 증가시키고 지정한 오류를 반환하는 작은 구조체를 만들어도 됩니다. 실제 디스크나 네트워크 연결이 필요하지 않습니다.

완료 조건: 출력 문자열뿐 아니라 오류 원인과 호출 횟수를 확인합니다. Close 오류를 처리 오류에 덮어쓰지 않습니다.

<details>
<summary>힌트 1</summary>

이름 붙인 반환값과 defer를 함께 사용하면 operation의 return 뒤에도 Close 결과를 반영할 수 있습니다.

</details>

<details>
<summary>힌트 2</summary>

Close가 실패했을 때 errors.Join으로 기존 err와 새 오류를 합칩니다. 두 오류의 메시지를 문자열 연결만 하는 방식은 구조를 잃습니다.

</details>

## 문제 6. 여러 입력 오류 보고

**수준: 응용**

이름과 나이를 검사하는 함수를 작성합니다.

```go
var ErrName = errors.New("name required")
var ErrAge = errors.New("age must be 0..150")
func validate(name string, age int) error
```

계약:

- strings.TrimSpace(name)이 빈 문자열이면 ErrName을 포함합니다.
- age가 0 미만 또는 150 초과이면 ErrAge를 포함합니다.
- 두 조건이 모두 실패하면 두 오류를 포함합니다.
- 정상 입력은 nil입니다.
- 검증은 이름 문자열을 수정해 반환하는 기능이 아닙니다.

| 이름 | 나이 | 결과 |
| --- | ---: | --- |
| Go | 0 | 성공 |
| Go | 150 | 성공 |
| 공백 | 0 | 이름 오류 |
| Go | -1 | 나이 오류 |
| 빈 문자열 | 151 | 두 오류 |

완료 조건: 두 오류를 errors.Is로 각각 검사합니다. 둘 다 성공했을 때 빈 메시지를 가진 오류 객체를 반환하지 않습니다.

<details>
<summary>힌트</summary>

오류 슬라이스에 실패한 검사만 추가한 뒤 errors.Join을 사용합니다. 빈 슬라이스를 결합한 결과는 nil입니다.

</details>

## 선택 A. 복구·종료 정책 읽기

**수준: 확인**

다음 질문에 설명으로 답합니다.

1. 사용자 포트 입력이 숫자가 아닌 상황에서 panic보다 error가 적합한 이유는 무엇인가요?
2. recover 뒤에 panic을 일으킨 함수의 다음 줄부터 다시 실행되나요?
3. 한 고루틴의 defer가 다른 고루틴의 panic을 잡을 수 있나요?
4. 알 수 없는 panic까지 모두 잡고 nil을 반환하면 어떤 정보가 사라지나요?
5. 자원을 닫는 defer가 있는데 os.Exit를 호출하면 왜 정리가 보장되지 않나요?

고루틴 실습은 18장 이후에 합니다. 여기서는 같은 실행 흐름이라는 범위만 이해하면 됩니다.

## 완료 체크리스트

- [ ] 실패 시 결과를 어떻게 사용할지 계약을 적었습니다.
- [ ] 문맥을 붙인 뒤에도 Is/As로 원인을 찾았습니다.
- [ ] defer 등록 시점과 실행 시점을 따로 추적했습니다.
- [ ] 처리 오류와 정리 오류의 네 조합을 확인했습니다.
- [ ] 자원은 정확히 한 번 닫힙니다.
- [ ] 일반 입력 오류를 panic으로 숨기지 않았습니다.

[본문](../../book/11-errors-defer/README.md) · [해설](../../solutions/11-errors-defer/README.md)
