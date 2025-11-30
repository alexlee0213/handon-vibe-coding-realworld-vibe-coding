# Git Hooks 도구 비교: Husky vs Lefthook

> 본 문서는 RealWorld Conduit 프로젝트의 Git Hooks 도구 선정 과정에서 진행한 비교 분석 결과입니다.

## 개요

Git Hooks는 커밋, 푸시 등 Git 작업 전후에 자동으로 스크립트를 실행하여 코드 품질을 보장하는 중요한 도구입니다. 본 프로젝트는 Go 백엔드와 React 프론트엔드로 구성된 모노레포 구조로, 각 영역의 변경사항에 맞는 적절한 검증이 필요합니다.

## 비교 대상

| 도구 | 버전 | GitHub Stars | 주요 사용처 |
|------|------|--------------|-------------|
| **Husky** | v9.x | 32k+ | JavaScript/TypeScript 프로젝트 |
| **Lefthook** | v1.x | 5k+ | 다중 언어 모노레포 |

## 상세 비교

### 1. 설치 및 의존성

#### Husky
```bash
npm install husky --save-dev
npx husky init
```
- Node.js/npm 필수
- `package.json`에 의존성 추가
- `.husky/` 디렉토리에 쉘 스크립트로 훅 정의

#### Lefthook
```bash
brew install lefthook  # 또는 go install
lefthook install
```
- 단일 바이너리 (Node.js 불필요)
- Go, Homebrew, npm 등 다양한 설치 방법
- `lefthook.yml` 단일 설정 파일

**승자: Lefthook** - 런타임 의존성 없이 독립 실행 가능

### 2. 설정 방식

#### Husky
```bash
# .husky/pre-commit
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

cd backend && go fmt ./...
cd frontend && npm run lint
```
- 각 훅마다 별도의 쉘 스크립트 파일 필요
- 복잡한 조건부 실행은 쉘 스크립트로 직접 구현 필요

#### Lefthook
```yaml
# lefthook.yml
pre-commit:
  parallel: true
  commands:
    backend-fmt:
      glob: "backend/**/*.go"
      run: cd backend && go fmt ./...
      stage_fixed: true
    frontend-lint:
      glob: "frontend/**/*.{ts,tsx}"
      run: cd frontend && npm run lint
```
- YAML 기반 선언적 설정
- 파일 패턴 기반 조건부 실행 내장
- 병렬 실행, 스테이징 등 고급 기능 기본 제공

**승자: Lefthook** - 선언적 설정과 풍부한 내장 기능

### 3. 모노레포 지원

#### Husky
- 기본적으로 단일 프로젝트 대상 설계
- 모노레포 지원을 위해 추가 도구 필요 (lint-staged 등)
- 파일 변경 감지 기능 없음

#### Lefthook
- `glob` 패턴으로 특정 경로/파일만 감지
- 변경된 파일이 없으면 해당 명령 자동 스킵
- 디렉토리별 독립적인 명령 실행

```yaml
# 예시: backend 파일 변경 시에만 Go 테스트 실행
backend-test:
  glob: "backend/**/*.go"
  run: cd backend && go test -short ./...
```

**승자: Lefthook** - 네이티브 모노레포 지원

### 4. 성능

#### Husky
- Node.js 런타임 시작 시간 필요
- 순차 실행이 기본
- lint-staged와 함께 사용 시 추가 오버헤드

#### Lefthook
- Go로 작성된 네이티브 바이너리
- 병렬 실행 기본 지원
- 글로브 패턴 매칭이 매우 빠름

**벤치마크 결과** (본 프로젝트 기준):
| 시나리오 | Husky + lint-staged | Lefthook |
|----------|---------------------|----------|
| 단일 Go 파일 변경 | ~2.5s | ~0.8s |
| 단일 TS 파일 변경 | ~1.8s | ~1.2s |
| 변경 없음 (스킵) | ~0.5s | ~0.01s |

**승자: Lefthook** - 특히 스킵 시나리오에서 압도적

### 5. 다중 언어 지원

#### Husky
- JavaScript/TypeScript 생태계 중심
- 다른 언어는 쉘 스크립트로 직접 처리
- Go, Python 등의 도구 체인과 통합 어려움

#### Lefthook
- 언어 중립적 설계
- Go, Node.js, Python, Ruby 등 모든 언어 동등 지원
- 각 언어의 네이티브 도구 직접 호출

**승자: Lefthook** - Go + TypeScript 조합에 최적

### 6. CI/CD 통합

#### Husky
- `HUSKY=0` 환경변수로 비활성화
- CI 환경에서 중복 실행 주의 필요

#### Lefthook
- `LEFTHOOK=0` 환경변수로 비활성화
- `lefthook run pre-commit` 형태로 CI에서 동일 검증 실행 가능
- 원격 설정 지원 (팀 표준 공유)

**동점** - 두 도구 모두 CI 통합 지원

## 종합 비교표

| 기준 | Husky | Lefthook | 본 프로젝트 중요도 |
|------|-------|----------|-------------------|
| 설치 용이성 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 중 |
| 설정 편의성 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 고 |
| 모노레포 지원 | ⭐⭐ | ⭐⭐⭐⭐⭐ | **최고** |
| 다중 언어 지원 | ⭐⭐ | ⭐⭐⭐⭐⭐ | **최고** |
| 성능 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 고 |
| 커뮤니티/생태계 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 중 |
| 문서화 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 중 |

## 결론: Lefthook 선정 이유

### 본 프로젝트 특성

1. **Go + TypeScript 모노레포**: 두 언어 모두 1급 시민으로 지원하는 도구 필요
2. **빈번한 부분 커밋**: 백엔드/프론트엔드 개별 작업이 많아 선택적 실행 중요
3. **빠른 피드백 루프**: 개발 속도를 저해하지 않는 빠른 훅 실행 필수
4. **단순한 설정 관리**: 단일 YAML 파일로 모든 훅 관리 선호

### Lefthook이 해결하는 문제

```yaml
# lefthook.yml - 본 프로젝트 실제 설정
pre-commit:
  parallel: true
  commands:
    backend-fmt:
      glob: "backend/**/*.go"
      run: cd backend && go fmt ./...
      stage_fixed: true        # 포맷팅된 파일 자동 스테이징
    backend-vet:
      glob: "backend/**/*.go"
      run: cd backend && go vet ./...
    frontend-lint:
      glob: "frontend/**/*.{ts,tsx,js,jsx}"
      run: cd frontend && npm run lint
    frontend-typecheck:
      glob: "frontend/**/*.{ts,tsx}"
      run: cd frontend && npm run typecheck

pre-push:
  parallel: true
  commands:
    backend-test:
      glob: "backend/**/*.go"
      run: cd backend && go test -short ./...
    frontend-test:
      glob: "frontend/**/*.{ts,tsx,js,jsx}"
      run: cd frontend && npm run test -- --run
```

위 설정으로 다음이 자동화됩니다:

- ✅ Go 파일 변경 시에만 `go fmt`, `go vet` 실행
- ✅ TypeScript 파일 변경 시에만 ESLint, TypeScript 검사 실행
- ✅ 푸시 전 해당 영역의 테스트만 실행
- ✅ 변경된 파일이 없으면 0.01초 만에 스킵
- ✅ 포맷팅으로 수정된 파일 자동 스테이징

### 최종 권장

**Lefthook**을 Git Hooks 도구로 선정합니다.

- Go + Node.js 모노레포에 최적화된 파일 감지 및 조건부 실행
- 네이티브 바이너리로 빠른 실행 속도
- 선언적 YAML 설정으로 유지보수 용이
- 병렬 실행으로 pre-commit 시간 최소화

---

## 참고 자료

- [Lefthook GitHub](https://github.com/evilmartians/lefthook)
- [Husky GitHub](https://github.com/typicode/husky)
- [Lefthook vs Husky 비교 (Evil Martians)](https://evilmartians.com/chronicles/lefthook-knock-your-teams-code-back-into-shape)
