## Languages | 언어

- [English](README.md)
- [한국어](README_ko.md) (현재)

---

# be-schema

이 라이브러리는 `Google Batchexecute` 를 사용하는 API Request, Response Payload 를 직렬화 혹은 역직렬화 하기 위한 라이브러리 입니다.

`Google` 에서 실제로 어떤 필드를 사용하는지는 알 수 없으나 임의의 명시적 구조체 필드를 사용하여 라이브러리 사용자가 필드를 유추하여 파싱할 수 있도록 만들었습니다.

## 주요 기능

- `beschema` 태그를 기반으로 Go 구조체를 순서가 있는 배열로 변환
- 배열을 적절한 필드 매핑으로 구조체로 다시 변환
- 중첩된 구조체 지원
- 명시적 필드 순서 제어
- 스키마 기반 순서를 사용한 JSON 마샬링/언마샬링

## 설치

```bash
go get github.com/starpia-forge/be-schema
```

## 사용법

### 예제

이 라이브러리는 스키마 마샬링/언마샬링을 처리하는 세 가지 방법을 제공합니다:

#### 1. 명시적 스키마 (Explicit Schema)
미리 정의된 구조체 타입과 `beschema` 태그를 사용하는 구조화된 데이터용:
- 참조: [`cmd/example/explicit/explicit.go`](cmd/example/explicit/explicit.go)

#### 2. 암시적 스키마 (Implicit Schema)  
미리 정의된 구조체 타입 없이 동적 데이터용:
- 참조: [`cmd/example/implicit/implicit.go`](cmd/example/implicit/implicit.go)

#### 3. 암시적 스트림 (Implicit Stream)
암시적 스키마를 사용하는 스트리밍 데이터용:
- 참조: [`cmd/example/implicit_stream/implicit_stream.go`](cmd/example/implicit_stream/implicit_stream.go)

### 예제 실행

다음 명령어로 예제를 실행할 수 있습니다:

```bash
# 명시적 스키마 예제
go run cmd/example/explicit/explicit.go

# 암시적 스키마 예제  
go run cmd/example/implicit/implicit.go

# 암시적 스트림 예제
go run cmd/example/implicit_stream/implicit_stream.go
```

## API 참조

### 함수

#### `MarshalExplicitSchema[T any](v T) ([]byte, error)`

`beschema` 태그 순서를 사용하여 구조체를 JSON 배열 형식으로 마샬링합니다.

**매개변수:**
- `v`: 마샬링할 구조체

**반환값:**
- `[]byte`: JSON으로 인코딩된 배열
- `error`: 마샬링 실패 시 오류

#### `UnmarshalExplicitSchema[T any](data []byte) (T, error)`

`beschema` 태그 순서를 사용하여 JSON 배열 데이터를 구조체로 다시 언마샬링합니다.

**매개변수:**
- `data`: JSON으로 인코딩된 배열 데이터

**반환값:**
- `T`: 언마샬링된 구조체
- `error`: 언마샬링 실패 시 오류

## 스키마 태그

결과 배열에서 필드의 순서를 지정하려면 `beschema` 태그를 사용하세요:

```go
type Example struct {
    Third  string `beschema:"3"`
    First  string `beschema:"1"`
    Second string `beschema:"2"`
}
```

구조체에서 선언된 순서와 관계없이 필드는 배열에서 `[First, Second, Third]` 순서로 정렬됩니다.

## 요구사항

- Go 1.24 이상

## 라이선스

이 프로젝트는 [MIT 라이선스](LICENSE) 하에 배포됩니다.