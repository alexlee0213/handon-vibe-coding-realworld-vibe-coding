# CLAUDE.md

RealWorld Conduit 애플리케이션 - Agentic Coding 기반 풀스택 소셜 블로깅 플랫폼

## Project Overview

RealWorld 스펙을 기반으로 한 "Medium.com 클론" 프로젝트. Armin Ronacher의 Agentic Coding 원칙을 적용하여 AI 에이전트와의 협업에 최적화된 코드베이스 구축.

### Tech Stack

**Backend**:
- Go 1.21+ (stdlib net/http, slog)
- SQLite (개발) / PostgreSQL 16 (운영)
- Pure SQL (ORM 미사용)
- golang-jwt v5, golang-migrate v4

**Frontend**:
- React 18.3+, Vite 5+, TypeScript 5.3+
- TanStack Router + TanStack Query v5
- Zustand (클라이언트 상태)
- Mantine v7 + Mantine Form + Zod
- Tabler Icons, Vitest + React Testing Library

**Infrastructure**:
- AWS CDK (TypeScript)
- ECS Fargate, GitHub Pages
- GitHub Actions CI/CD

## Development Commands

```bash
# 의존성 설치
make install

# 개발 서버 실행 (backend + frontend)
make dev
make dev-backend    # 백엔드만
make dev-frontend   # 프론트엔드만

# 테스트
make test           # 전체 테스트
make test-watch     # 프론트엔드 watch 모드
make test-coverage  # 커버리지 포함

# 린트 & 타입체크
make lint
make typecheck

# 빌드
make build          # 프로덕션 빌드
make docker-build   # Docker 이미지

# 데이터베이스
make db-init        # SQLite 초기화
make db-up          # PostgreSQL (Docker)
make migrate        # 마이그레이션 실행
make migrate-down   # 롤백
make migrate-status # 상태 확인

# 배포
make deploy         # AWS CDK 배포
make deploy-frontend # GitHub Pages 배포

# 정리
make clean
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
│  React + Vite + TanStack Router + Mantine UI                │
└────────────────────────────┬────────────────────────────────┘
                             │ HTTP/REST
┌────────────────────────────▼────────────────────────────────┐
│                         Backend                              │
│  Go net/http + JWT Auth + slog logging                      │
└────────────────────────────┬────────────────────────────────┘
                             │ SQL
┌────────────────────────────▼────────────────────────────────┐
│              SQLite (Dev) / PostgreSQL (Prod)               │
└─────────────────────────────────────────────────────────────┘
```

### Backend Layer Structure

```
HTTP Layer (net/http)
    ↓
Service Layer (비즈니스 로직, 권한 체크)
    ↓
Repository Layer (Pure SQL)
    ↓
Database
```

### Frontend State Management

- **TanStack Query**: 서버 상태 (API 캐싱, 백그라운드 동기화)
- **Zustand**: 클라이언트 상태 (인증 토큰, UI 상태)
- **Mantine Form + Zod**: 폼 상태 및 검증

## Project Structure

```
realworld-vibe-coding/
├── backend/
│   ├── cmd/server/           # 진입점
│   ├── internal/
│   │   ├── api/              # HTTP 핸들러, 미들웨어
│   │   ├── config/           # 환경 설정
│   │   ├── domain/           # 도메인 모델, 에러
│   │   ├── repository/       # 데이터 접근
│   │   └── service/          # 비즈니스 로직
│   ├── db/
│   │   ├── migrations/       # SQL 마이그레이션
│   │   └── queries/          # SQL 쿼리
│   ├── Dockerfile
│   └── Makefile
├── frontend/
│   ├── src/
│   │   ├── components/       # UI 컴포넌트
│   │   ├── features/         # 기능별 모듈 (api, hooks, store, types)
│   │   ├── routes/           # TanStack Router 페이지
│   │   ├── stores/           # Zustand 스토어
│   │   ├── lib/              # 유틸리티
│   │   └── test/             # 테스트 설정
│   ├── Dockerfile
│   └── vite.config.ts
├── infra/
│   ├── bin/                  # CDK 앱 진입점
│   └── lib/                  # CDK 스택 정의
├── docs/
│   ├── pre-prd.md
│   └── prd.md
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Agentic Coding Principles

### Code Patterns

```go
// ✅ 명확하고 설명적인 함수명
func GetUserByID(ctx context.Context, userID int64) (*User, error)
func ListArticlesByAuthorUsername(ctx context.Context, username string, limit, offset int) ([]Article, error)

// ✅ 코드에서 명시적 권한 체크
if article.AuthorID != currentUserID {
    return ErrForbidden
}

// ✅ 구조화된 로깅
slog.Info("article created",
    "article_id", article.ID,
    "slug", article.Slug,
    "author_id", article.AuthorID,
)
```

### Key Principles

1. **명확하고 긴 함수명** > 짧고 모호한 함수명
2. **합성(Composition)** > 상속(Inheritance)
3. **Plain SQL** > ORM
4. **명시적 권한 체크** > 설정 파일 기반 권한
5. **코드 생성** > 외부 의존성 추가
6. **구조화된 로깅** > printf 디버깅

## Development Guidelines

### Backend: TDD (Test-Driven Development)

백엔드 구현은 TDD 원칙을 따른다:

1. **Red**: 실패하는 테스트 먼저 작성
2. **Green**: 테스트를 통과하는 최소한의 코드 작성
3. **Refactor**: 코드 개선 (테스트는 계속 통과해야 함)

```go
// 1. 테스트 먼저 작성
func TestCreateUser(t *testing.T) {
    // 테스트 케이스 정의
}

// 2. 구현 코드 작성
func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
    // 최소한의 구현
}

// 3. 리팩토링
```

**TDD 적용 범위**:
- Repository Layer: 모든 데이터 접근 메서드
- Service Layer: 비즈니스 로직
- Handler Layer: HTTP 엔드포인트 (통합 테스트)

### Frontend: Playwright MCP 활용 (⚠️ 필수)

> **🚨 중요: Frontend UI 이슈 구현 시 Playwright MCP 테스트는 필수입니다!**
>
> Frontend UI 관련 이슈를 완료하기 전에 반드시 Playwright MCP를 사용하여
> 실제 브라우저에서 UI가 올바르게 렌더링되고 동작하는지 검증해야 합니다.
> **Playwright 테스트 없이 Frontend UI 이슈를 닫지 마세요.**

프론트엔드 개발 시 Playwright MCP를 사용하여 구현 상태를 시각적으로 확인한다:

1. **컴포넌트 개발 후**: 브라우저에서 렌더링 상태 확인
2. **페이지 구현 후**: 네비게이션 및 라우팅 동작 검증
3. **폼 구현 후**: 입력/제출 플로우 테스트
4. **API 연동 후**: 실제 데이터 흐름 확인

```bash
# Playwright MCP 활용 예시
- browser_navigate: 페이지 이동
- browser_snapshot: 현재 상태 캡처
- browser_click: 버튼/링크 클릭 테스트
- browser_fill_form: 폼 입력 테스트
- browser_take_screenshot: 스크린샷 저장
```

**⚠️ 필수 확인 시점** (이슈 완료 전 반드시 수행):
- 새로운 페이지/컴포넌트 구현 완료 시 → **browser_snapshot 필수**
- 스타일링 변경 시 → **시각적 확인 필수**
- 사용자 인터랙션 로직 구현 시 → **browser_click/fill_form 테스트 필수**
- 버그 수정 후 검증 시 → **수정 확인 필수**

**Frontend UI 이슈 완료 체크리스트**:
```
□ typecheck 통과
□ lint 통과
□ Playwright MCP로 브라우저 테스트 수행 ← 필수!
  □ 페이지 렌더링 확인 (browser_snapshot)
  □ 주요 UI 요소 표시 확인
  □ 인터랙션 동작 확인 (해당 시)
□ 커밋 및 이슈 닫기
```

## API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | /api/users | 회원가입 | - |
| POST | /api/users/login | 로그인 | - |
| GET | /api/user | 현재 사용자 조회 | Required |
| PUT | /api/user | 사용자 정보 수정 | Required |
| GET | /api/profiles/:username | 프로필 조회 | Optional |
| POST | /api/profiles/:username/follow | 팔로우 | Required |
| DELETE | /api/profiles/:username/follow | 언팔로우 | Required |
| GET | /api/articles | 아티클 목록 | Optional |
| GET | /api/articles/feed | 피드 | Required |
| POST | /api/articles | 아티클 생성 | Required |
| GET | /api/articles/:slug | 아티클 조회 | Optional |
| PUT | /api/articles/:slug | 아티클 수정 | Required |
| DELETE | /api/articles/:slug | 아티클 삭제 | Required |
| POST | /api/articles/:slug/favorite | 좋아요 | Required |
| DELETE | /api/articles/:slug/favorite | 좋아요 취소 | Required |
| GET | /api/articles/:slug/comments | 댓글 목록 | Optional |
| POST | /api/articles/:slug/comments | 댓글 작성 | Required |
| DELETE | /api/articles/:slug/comments/:id | 댓글 삭제 | Required |
| GET | /api/tags | 태그 목록 | - |

**Authentication**: `Authorization: Token jwt.token.here`

## Frontend Routes

| Route | Page | Auth |
|-------|------|------|
| `/` | 홈 (피드 탭, 태그 필터) | Optional |
| `/login` | 로그인 | - |
| `/register` | 회원가입 | - |
| `/settings` | 설정 | Required |
| `/editor` | 새 글 작성 | Required |
| `/editor/:slug` | 글 수정 | Required |
| `/article/:slug` | 글 상세 | Optional |
| `/profile/:username` | 프로필 | Optional |
| `/profile/:username/favorites` | 좋아요한 글 | Optional |

## Environment Variables

```bash
# Database
DATABASE_URL=sqlite://./data/conduit.db  # 개발
# DATABASE_URL=postgres://user:pass@host:5432/db  # 운영

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=72h

# Server
SERVER_PORT=8080
SERVER_ENV=development

# Frontend
VITE_API_URL=http://localhost:8080/api

# AWS
AWS_REGION=ap-northeast-2
```

## Testing

```bash
# Backend (Go)
cd backend && go test -v ./...
cd backend && go test -v -coverprofile=coverage.out ./...

# Frontend (Vitest)
cd frontend && npm run test
cd frontend && npm run test:watch
cd frontend && npm run test:coverage
```

### Test Patterns

```typescript
// Frontend - MantineProvider 래퍼 사용
import { MantineProvider } from '@mantine/core';

const renderWithMantine = (component: React.ReactNode) => {
  return render(<MantineProvider>{component}</MantineProvider>);
};
```

## References

- [RealWorld Spec](https://realworld-docs.netlify.app/)
- [Agentic Coding - Armin Ronacher](https://lucumr.pocoo.org/2025/6/12/agentic-coding/)
- [Mantine UI](https://mantine.dev/)
- [TanStack Query](https://tanstack.com/query)
- [TanStack Router](https://tanstack.com/router)
- [AWS CDK](https://docs.aws.amazon.com/cdk/)
