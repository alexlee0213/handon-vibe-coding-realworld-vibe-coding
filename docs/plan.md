# Implementation Plan: RealWorld Conduit

> 본 문서는 PRD 기반 단계별 구현 계획서입니다.
> Agentic Coding 방식으로 AI 에이전트와 협업하여 구현합니다.

## 구현 원칙

### Agentic Coding 워크플로우

```
1. 작업 단위를 작게 유지 (한 번에 하나의 기능)
2. 각 작업마다 테스트 작성 → 구현 → 검증
3. Makefile 명령어로 빠른 피드백 루프
4. 구조화된 로깅으로 디버깅 용이성 확보
```

### 작업 순서 원칙

```
Backend 우선 → Frontend 연동
각 Phase 내: API 구현 → 테스트 → UI 구현 → 통합 테스트
```

---

## Phase 1: 프로젝트 셋업

### 1.1 프로젝트 구조 생성

- [x] 루트 디렉토리 구조 생성
- [x] docs/ 폴더 및 PRD 문서
- [x] .gitignore 설정
- [x] CLAUDE.md 작성

### 1.2 Backend 초기 설정

```
backend/
├── cmd/server/main.go
├── internal/
│   ├── api/
│   │   ├── router.go
│   │   ├── middleware/
│   │   └── handler/
│   ├── config/config.go
│   ├── domain/
│   └── repository/
├── db/
│   └── migrations/
├── go.mod
├── Dockerfile
└── Makefile
```

**작업 목록:**

- [x] `go mod init` 실행
- [x] 기본 디렉토리 구조 생성
- [x] `cmd/server/main.go` - HTTP 서버 기본 구조
- [x] `internal/config/config.go` - 환경변수 로드 (envconfig)
- [x] `internal/api/router.go` - 기본 라우터 설정
- [x] `internal/api/middleware/logging.go` - 요청 로깅
- [x] `internal/api/middleware/cors.go` - CORS 설정
- [x] Health check 엔드포인트 (`GET /health`)
- [x] Backend Makefile 작성
- [x] Backend Dockerfile 작성

**검증:**
```bash
make dev-backend
curl http://localhost:8080/health
# {"status": "ok"}
```

### 1.3 Frontend 초기 설정

```
frontend/
├── src/
│   ├── components/
│   ├── features/
│   ├── routes/
│   ├── stores/
│   ├── lib/
│   ├── test/
│   ├── App.tsx
│   └── main.tsx
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── Dockerfile
```

**작업 목록:**

- [x] `npm create vite@latest` 실행 (React + TypeScript)
- [x] 의존성 설치:
  ```bash
  npm install @mantine/core @mantine/hooks @mantine/form @mantine/notifications
  npm install @tanstack/react-router @tanstack/react-query
  npm install zustand zod ky
  npm install @tabler/icons-react
  npm install -D vitest @testing-library/react @testing-library/jest-dom jsdom
  ```
- [x] Mantine Provider 설정 (`App.tsx`)
- [x] 기본 테마 설정 (RealWorld 브랜드 색상)
- [x] TanStack Router 기본 설정
- [x] TanStack Query Provider 설정
- [x] Vitest 설정 (`vitest.config.ts`)
- [x] 테스트 셋업 파일 (`src/test/setup.ts`)
- [x] API 클라이언트 설정 (`src/lib/api.ts`)
- [x] Frontend Dockerfile 작성

**검증:**
```bash
make dev-frontend
# http://localhost:5173 접속 확인
npm run test
```

### 1.4 Infrastructure 초기 설정

```
infra/
├── bin/infra.ts
├── lib/
│   └── ...stacks
├── cdk.json
├── package.json
└── tsconfig.json
```

**작업 목록:**

- [x] `npx cdk init app --language typescript`
- [x] VPC 스택 기본 구조
- [x] CDK 환경 변수 설정

### 1.5 개발 환경 통합

**작업 목록:**

- [x] 루트 Makefile 작성 (PRD 9.4 참조)
- [x] `docker-compose.yml` 작성 (PostgreSQL)
- [x] `.env.example` 파일 생성
- [x] `scripts/dev.sh` 스크립트 작성

**검증:**
```bash
make install
make dev
# Backend: http://localhost:8080
# Frontend: http://localhost:5173
```

### 1.6 데이터베이스 설정

**작업 목록:**

- [x] golang-migrate 설치
- [x] 마이그레이션 파일 생성:
  - `000001_create_users_table.up.sql`
  - `000001_create_users_table.down.sql`
  - `000002_create_articles_table.up.sql`
  - `000002_create_articles_table.down.sql`
  - `000003_create_tags_table.up.sql`
  - `000003_create_tags_table.down.sql`
  - `000004_create_comments_table.up.sql`
  - `000004_create_comments_table.down.sql`
  - `000005_create_follows_table.up.sql`
  - `000005_create_follows_table.down.sql`
  - `000006_create_favorites_table.up.sql`
  - `000006_create_favorites_table.down.sql`
- [x] SQLite 초기화 스크립트

**검증:**
```bash
make db-init
make migrate
make migrate-status
```

### 1.7 CI/CD 파이프라인

**작업 목록:**

- [x] `.github/workflows/ci.yml` 작성
  - Backend 테스트
  - Frontend 테스트 (lint, typecheck, vitest)
  - 빌드 검증
- [x] `.github/workflows/deploy-frontend.yml` 작성 (GitHub Pages)

**검증:**
- GitHub에 push 후 Actions 실행 확인

---

## Phase 2: 인증 시스템

### 2.1 Backend - 사용자 도메인

**작업 목록:**

- [x] `internal/domain/user.go` - User 구조체
- [x] `internal/domain/errors.go` - 도메인 에러 정의
- [x] `internal/repository/user.go` - 사용자 저장소 인터페이스 및 구현
  - `CreateUser`
  - `GetUserByID`
  - `GetUserByEmail`
  - `GetUserByUsername`
  - `UpdateUser`
- [x] 단위 테스트 작성

### 2.2 Backend - 인증 서비스

**작업 목록:**

- [x] `internal/service/auth.go` - 인증 서비스
  - `Register` - 회원가입
  - `Login` - 로그인
  - `GenerateToken` - JWT 생성
  - `ValidateToken` - JWT 검증
- [x] 비밀번호 해싱 (bcrypt)
- [x] 단위 테스트 작성

### 2.3 Backend - 인증 API

**작업 목록:**

- [x] `internal/api/handler/user.go` - 사용자 핸들러
  - `POST /api/users` - 회원가입
  - `POST /api/users/login` - 로그인
  - `GET /api/user` - 현재 사용자 조회
  - `PUT /api/user` - 사용자 정보 수정
- [x] `internal/api/middleware/auth.go` - JWT 인증 미들웨어
- [x] `internal/api/request/user.go` - 요청 DTO
- [x] `internal/api/response/user.go` - 응답 DTO
- [x] 통합 테스트 작성

**검증:**
```bash
make test
# API 테스트
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"user":{"username":"test","email":"test@test.com","password":"password"}}'
```

### 2.4 Frontend - 인증 상태 관리

**작업 목록:**

- [x] `src/features/auth/types.ts` - 타입 정의
- [x] `src/features/auth/api.ts` - API 호출 함수
- [x] `src/features/auth/store.ts` - Zustand 스토어 (토큰, 사용자 정보)
- [x] `src/features/auth/hooks.ts` - TanStack Query 훅
- [x] `src/features/auth/schemas.ts` - Zod 검증 스키마
- [x] `src/lib/api.ts` - 인증 헤더 인터셉터 추가

### 2.5 Frontend - 인증 UI

**작업 목록:**

- [x] `src/components/layout/Header.tsx` - 네비게이션 (로그인 상태 반영)
- [x] `src/components/layout/Layout.tsx` - 기본 레이아웃
- [x] `src/routes/login.tsx` - 로그인 페이지
- [x] `src/routes/register.tsx` - 회원가입 페이지
- [x] 인증 필요 라우트 가드 구현
- [x] 컴포넌트 테스트 작성

**검증:**
```bash
npm run test
# 브라우저에서 회원가입 → 로그인 플로우 테스트
```

---

## Phase 3: 아티클 CRUD

### 3.1 Backend - 아티클 도메인

**작업 목록:**

- [x] `internal/domain/article.go` - Article 구조체
- [x] `internal/repository/article.go` - 아티클 저장소
  - `CreateArticle`
  - `GetArticleBySlug`
  - `ListArticles` (필터링, 페이지네이션)
  - `UpdateArticle`
  - `DeleteArticle`
- [x] Slug 생성 유틸리티
- [x] 단위 테스트 작성

### 3.2 Backend - 아티클 API

**작업 목록:**

- [x] `internal/api/handler/article.go` - 아티클 핸들러
  - `POST /api/articles` - 아티클 생성
  - `GET /api/articles` - 아티클 목록
  - `GET /api/articles/:slug` - 아티클 상세
  - `PUT /api/articles/:slug` - 아티클 수정
  - `DELETE /api/articles/:slug` - 아티클 삭제
- [x] 권한 체크 (작성자만 수정/삭제)
- [x] 통합 테스트 작성

### 3.3 Frontend - 아티클 상태 관리

**작업 목록:**

- [x] `src/features/article/types.ts` - 타입 정의
- [x] `src/features/article/api.ts` - API 호출 함수
- [x] `src/features/article/hooks.ts` - TanStack Query 훅
- [x] `src/features/article/schemas.ts` - Zod 검증 스키마

### 3.4 Frontend - 아티클 UI

**작업 목록:**

- [x] `src/components/article/ArticleCard.tsx` - 아티클 카드
- [x] `src/components/article/ArticleList.tsx` - 아티클 목록
- [x] `src/components/article/ArticleMeta.tsx` - 아티클 메타 정보
- [x] `src/routes/index.tsx` - 홈 페이지 (아티클 목록)
- [x] `src/routes/article/$slug.tsx` - 아티클 상세 페이지
- [x] `src/routes/editor/index.tsx` - 새 글 작성 페이지
- [x] `src/routes/editor/$slug.tsx` - 글 수정 페이지
- [x] 컴포넌트 테스트 작성

---

## Phase 4: 댓글 및 태그

### 4.1 Backend - 댓글/태그 도메인

**작업 목록:**

- [x] `internal/domain/comment.go` - Comment 구조체
- [x] `internal/domain/tag.go` - Tag 구조체
- [x] `internal/repository/comment.go` - 댓글 저장소
- [x] `internal/repository/tag.go` - 태그 저장소
- [x] 단위 테스트 작성

### 4.2 Backend - 댓글/태그 API

**작업 목록:**

- [x] `internal/api/handler/comment.go` - 댓글 핸들러
  - `GET /api/articles/:slug/comments`
  - `POST /api/articles/:slug/comments`
  - `DELETE /api/articles/:slug/comments/:id`
- [x] `internal/api/handler/tag.go` - 태그 핸들러 (ArticleHandler.GetTags에 이미 구현됨)
  - `GET /api/tags`
- [x] 통합 테스트 작성

### 4.3 Frontend - 댓글/태그 UI

**작업 목록:**

- [ ] `src/features/comment/` - 댓글 기능 모듈
- [ ] `src/components/comment/CommentCard.tsx`
- [ ] `src/components/comment/CommentForm.tsx`
- [ ] `src/components/comment/CommentList.tsx`
- [ ] `src/components/common/TagList.tsx`
- [ ] 홈 페이지에 태그 사이드바 추가
- [ ] 아티클 상세 페이지에 댓글 섹션 추가
- [ ] 컴포넌트 테스트 작성

---

## Phase 5: 프로필 및 팔로우

### 5.1 Backend - 프로필/팔로우 도메인

**작업 목록:**

- [ ] `internal/domain/profile.go` - Profile 구조체
- [ ] `internal/repository/follow.go` - 팔로우 저장소
  - `Follow`
  - `Unfollow`
  - `IsFollowing`
- [ ] 단위 테스트 작성

### 5.2 Backend - 프로필/팔로우 API

**작업 목록:**

- [ ] `internal/api/handler/profile.go` - 프로필 핸들러
  - `GET /api/profiles/:username`
  - `POST /api/profiles/:username/follow`
  - `DELETE /api/profiles/:username/follow`
- [ ] 통합 테스트 작성

### 5.3 Frontend - 프로필 UI

**작업 목록:**

- [ ] `src/features/profile/` - 프로필 기능 모듈
- [ ] `src/components/profile/ProfileCard.tsx`
- [ ] `src/routes/profile/$username.tsx` - 프로필 페이지
- [ ] `src/routes/profile/$username.favorites.tsx` - 좋아요한 글
- [ ] `src/routes/settings.tsx` - 설정 페이지
- [ ] 팔로우/언팔로우 버튼 구현
- [ ] 컴포넌트 테스트 작성

---

## Phase 6: 피드 및 좋아요

### 6.1 Backend - 피드/좋아요 기능

**작업 목록:**

- [ ] `internal/repository/favorite.go` - 좋아요 저장소
  - `Favorite`
  - `Unfavorite`
  - `IsFavorited`
  - `GetFavoritesCount`
- [ ] `internal/repository/article.go` 확장
  - `GetFeed` - 팔로우한 사용자의 아티클 목록
- [ ] 단위 테스트 작성

### 6.2 Backend - 피드/좋아요 API

**작업 목록:**

- [ ] `internal/api/handler/article.go` 확장
  - `GET /api/articles/feed`
  - `POST /api/articles/:slug/favorite`
  - `DELETE /api/articles/:slug/favorite`
- [ ] 통합 테스트 작성

### 6.3 Frontend - 피드/좋아요 UI

**작업 목록:**

- [ ] 홈 페이지 피드 탭 구현 (Your Feed / Global Feed / Tag)
- [ ] 좋아요 버튼 컴포넌트
- [ ] 낙관적 업데이트 구현 (TanStack Query)
- [ ] 컴포넌트 테스트 작성

---

## Phase 7: 최종 테스트 및 배포

### 7.1 테스트 보강

**작업 목록:**

- [ ] E2E 테스트 시나리오 작성 (Playwright 또는 수동)
  - 회원가입 → 로그인 → 글 작성 → 댓글 → 로그아웃
  - 팔로우 → 피드 확인
  - 좋아요 → 프로필에서 확인
- [ ] 테스트 커버리지 확인 (목표: 80%+)
- [ ] 성능 테스트 (기본적인 부하 테스트)

### 7.2 보안 검토

**작업 목록:**

- [ ] SQL Injection 방지 확인
- [ ] XSS 방지 확인
- [ ] CSRF 보호 확인
- [ ] JWT 보안 설정 검토
- [ ] 환경변수 노출 확인
- [ ] CORS 설정 검토

### 7.3 프로덕션 배포

**작업 목록:**

- [ ] AWS CDK 스택 완성
  - VPC 스택
  - RDS PostgreSQL 스택
  - ECS Fargate 스택 (Backend)
  - S3 + CloudFront 스택 (대안: GitHub Pages)
- [ ] 환경별 설정 분리 (staging / production)
- [ ] CloudWatch 로그 설정
- [ ] 배포 테스트

### 7.4 문서화 완료

**작업 목록:**

- [ ] README.md 작성 (프로젝트 소개, 실행 방법)
- [ ] API 문서 정리
- [ ] 배포 가이드 작성
- [ ] 기여 가이드 작성

---

## 체크리스트 요약

| Phase | 주요 산출물 | 검증 방법 |
|-------|------------|-----------|
| 1 | 프로젝트 구조, DB 스키마, CI/CD | `make dev`, `make test` |
| 2 | 인증 API, 로그인/회원가입 UI | API 테스트, UI 플로우 |
| 3 | 아티클 CRUD API, 에디터 UI | CRUD 전체 플로우 |
| 4 | 댓글/태그 API, UI | 댓글 작성/삭제, 태그 필터 |
| 5 | 프로필/팔로우 API, UI | 팔로우/언팔로우 플로우 |
| 6 | 피드/좋아요 API, UI | 피드 탭, 좋아요 반영 |
| 7 | E2E 테스트, 배포 | 프로덕션 환경 동작 |

---

## 작업 시 유의사항

### 커밋 전 체크리스트

```bash
# 항상 실행
make lint
make test

# 변경사항 확인
git diff
git status
```

### 디버깅 시

```bash
# Backend 로그 확인
make dev-backend 2>&1 | grep -E "(ERROR|WARN|INFO)"

# Frontend 콘솔 확인
# 브라우저 개발자 도구 → Console 탭
```

### 문제 발생 시

1. 에러 메시지 전체 확인
2. 관련 로그 확인
3. 최근 변경사항 검토
4. 테스트 실행하여 범위 좁히기
