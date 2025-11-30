# Pre-PRD: RealWorld Conduit 프로젝트

> 본 문서는 RealWorld 스펙 기반 Conduit 애플리케이션 구현을 위한 사전 요구사항 분석 문서입니다.

## 1. 프로젝트 개요

### 1.1 RealWorld란?

RealWorld는 "Medium.com 클론"으로, 실제 운영 환경 수준의 풀스택 애플리케이션을 다양한 기술 스택으로 구현하는 오픈소스 프로젝트입니다. 모든 구현체는 동일한 API 스펙을 따르므로, 프론트엔드와 백엔드를 자유롭게 조합할 수 있습니다.

### 1.2 프로젝트 목표

- RealWorld 스펙을 완벽히 준수하는 풀스택 애플리케이션 구현
- Agentic Coding(바이브 코딩) 방식으로 개발 진행
- 현대적인 기술 스택과 모범 사례 적용
- 프로덕션 수준의 품질과 배포 자동화 구현

---

## 2. RealWorld 기능 명세

### 2.1 사용자 인증 (Authentication)

| 기능 | 설명 | API 엔드포인트 |
|------|------|----------------|
| 회원가입 | 사용자명, 이메일, 비밀번호로 계정 생성 | `POST /api/users` |
| 로그인 | 이메일, 비밀번호로 인증 | `POST /api/users/login` |
| 현재 사용자 조회 | JWT 토큰으로 현재 사용자 정보 조회 | `GET /api/user` |
| 사용자 정보 수정 | 프로필 정보(이메일, 사용자명, 비밀번호, 이미지, 소개) 업데이트 | `PUT /api/user` |

**인증 방식**: JWT 토큰 (헤더: `Authorization: Token jwt.token.here`)

### 2.2 프로필 (Profiles)

| 기능 | 설명 | API 엔드포인트 |
|------|------|----------------|
| 프로필 조회 | 특정 사용자의 프로필 정보 조회 | `GET /api/profiles/:username` |
| 팔로우 | 사용자 팔로우 | `POST /api/profiles/:username/follow` |
| 언팔로우 | 사용자 언팔로우 | `DELETE /api/profiles/:username/follow` |

### 2.3 아티클 (Articles)

| 기능 | 설명 | API 엔드포인트 |
|------|------|----------------|
| 아티클 목록 조회 | 전체 아티클 목록 (필터링, 페이지네이션 지원) | `GET /api/articles` |
| 피드 조회 | 팔로우한 사용자의 아티클 목록 | `GET /api/articles/feed` |
| 아티클 상세 조회 | slug로 특정 아티클 조회 | `GET /api/articles/:slug` |
| 아티클 생성 | 새 아티클 작성 | `POST /api/articles` |
| 아티클 수정 | 기존 아티클 수정 | `PUT /api/articles/:slug` |
| 아티클 삭제 | 아티클 삭제 | `DELETE /api/articles/:slug` |
| 좋아요 | 아티클 좋아요 | `POST /api/articles/:slug/favorite` |
| 좋아요 취소 | 아티클 좋아요 취소 | `DELETE /api/articles/:slug/favorite` |

**쿼리 파라미터**:
- `tag`: 태그로 필터링
- `author`: 작성자로 필터링
- `favorited`: 좋아요한 사용자로 필터링
- `limit`: 결과 수 제한 (기본값: 20)
- `offset`: 오프셋 (기본값: 0)

### 2.4 댓글 (Comments)

| 기능 | 설명 | API 엔드포인트 |
|------|------|----------------|
| 댓글 목록 조회 | 아티클의 댓글 목록 | `GET /api/articles/:slug/comments` |
| 댓글 작성 | 아티클에 댓글 추가 | `POST /api/articles/:slug/comments` |
| 댓글 삭제 | 댓글 삭제 | `DELETE /api/articles/:slug/comments/:id` |

### 2.5 태그 (Tags)

| 기능 | 설명 | API 엔드포인트 |
|------|------|----------------|
| 태그 목록 조회 | 사용된 모든 태그 목록 | `GET /api/tags` |

---

## 3. 프론트엔드 라우팅 명세

| 라우트 | 페이지 | 인증 필요 |
|--------|--------|-----------|
| `/` | 홈 (글로벌/피드/태그별 아티클 목록) | 선택적 |
| `/login` | 로그인 | 불필요 |
| `/register` | 회원가입 | 불필요 |
| `/settings` | 설정 | 필수 |
| `/editor` | 새 아티클 작성 | 필수 |
| `/editor/:slug` | 아티클 수정 | 필수 |
| `/article/:slug` | 아티클 상세 | 선택적 |
| `/profile/:username` | 사용자 프로필 (작성 글) | 선택적 |
| `/profile/:username/favorites` | 사용자 프로필 (좋아요 글) | 선택적 |

---

## 4. 기술 스택 결정

### 4.1 기술 스택 선정 원칙 (Armin Ronacher's Agentic Coding 참고)

Armin Ronacher의 "Agentic Coding" 글에서 제안하는 핵심 원칙:

1. **명시적이고 단순한 코드**: LLM이 이해하기 쉬운 명확한 코드 구조
2. **빠른 도구**: 빠른 피드백 루프를 위한 빠른 개발 도구
3. **Plain SQL 선호**: ORM보다 직접 SQL 사용
4. **로깅 기반 관찰성**: 에이전트가 로그를 읽어 문제 진단
5. **Makefile 활용**: 중요한 개발 명령어를 Makefile로 관리

### 4.2 최종 기술 스택

#### Backend

| 영역 | 기술 | 선정 이유 |
|------|------|-----------|
| **언어** | Go | Armin 권장: 명시적 컨텍스트 시스템, 간단한 테스트 실행, LLM이 이해하기 쉬운 구조적 인터페이스 |
| **웹 프레임워크** | Chi 또는 Echo | 경량, 표준 라이브러리 기반, 명시적 라우팅 |
| **데이터베이스** | PostgreSQL | 신뢰성, 성능, 풍부한 기능 |
| **데이터 접근** | sqlc | Plain SQL로 타입 안전한 Go 코드 생성 |
| **마이그레이션** | golang-migrate | 버전 관리 가능한 DB 마이그레이션 |
| **인증** | JWT (golang-jwt) | RealWorld 스펙 준수 |

#### Frontend

| 영역 | 기술 | 선정 이유 |
|------|------|-----------|
| **프레임워크** | React 18+ | Armin 권장, 풍부한 생태계 |
| **빌드 도구** | Vite | Armin 권장, 빠른 개발 서버 |
| **상태/데이터** | TanStack Query | Armin 권장, 서버 상태 관리 |
| **라우팅** | TanStack Router | Armin 권장, 타입 안전 라우팅 |
| **UI 라이브러리** | Mantine UI | 사용자 요청, 120+ 컴포넌트, 접근성 지원 |
| **폼 관리** | React Hook Form | Mantine과 잘 통합됨 |
| **HTTP 클라이언트** | ky 또는 fetch | 경량, 간단한 API |

#### Infrastructure

| 영역 | 기술 | 선정 이유 |
|------|------|-----------|
| **IaC** | AWS CDK (TypeScript) | 사용자 요청, 타입 안전한 인프라 정의 |
| **컨테이너** | Docker | 일관된 개발/배포 환경 |
| **오케스트레이션** | AWS ECS Fargate | 서버리스 컨테이너 관리 |
| **데이터베이스** | AWS RDS PostgreSQL | 관리형 PostgreSQL |
| **CI/CD** | GitHub Actions | 사용자 요청, GitHub 통합 |
| **CDN/호스팅** | AWS CloudFront + S3 | 정적 자산 배포 |

---

## 5. 프로젝트 구조 (제안)

```
realworld-vibe-coding/
├── .github/
│   └── workflows/           # GitHub Actions CI/CD
├── backend/
│   ├── cmd/
│   │   └── server/          # 메인 진입점
│   ├── internal/
│   │   ├── api/             # HTTP 핸들러
│   │   ├── domain/          # 도메인 모델
│   │   ├── repository/      # 데이터 접근 계층
│   │   └── service/         # 비즈니스 로직
│   ├── db/
│   │   ├── migrations/      # SQL 마이그레이션
│   │   └── queries/         # sqlc 쿼리 파일
│   ├── Dockerfile
│   └── Makefile
├── frontend/
│   ├── src/
│   │   ├── components/      # 공통 컴포넌트
│   │   ├── features/        # 기능별 모듈
│   │   ├── hooks/           # 커스텀 훅
│   │   ├── routes/          # 페이지 라우트
│   │   └── lib/             # 유틸리티
│   ├── Dockerfile
│   └── package.json
├── infra/
│   ├── lib/                 # CDK 스택 정의
│   ├── bin/                 # CDK 앱 진입점
│   └── cdk.json
├── docs/
│   ├── pre-prd.md           # 본 문서
│   └── prd.md               # 상세 요구사항 (작성 예정)
├── Makefile                 # 프로젝트 루트 명령어
└── README.md
```

---

## 6. 개발 원칙 (Agentic Coding)

### 6.1 코드 작성 원칙

1. **단순하고 설명적인 코드**: 긴 함수명도 괜찮음, 명확성 우선
2. **상속 지양**: 상속보다 합성(composition) 선호
3. **명시적 권한 체크**: 설정 파일이 아닌 코드 내 명시적 권한 확인
4. **코드 생성 활용**: 의존성보다 코드 생성 선호 (sqlc 등)

### 6.2 도구 설정 원칙

1. **Makefile 활용**: 모든 중요 명령어를 Makefile로 관리
2. **빠른 피드백**: 빠른 테스트 실행, 빠른 빌드
3. **로깅 중심 관찰성**: 에이전트가 로그로 문제 진단 가능하도록

### 6.3 개발 워크플로우

```makefile
# 예상 Makefile 타겟
make dev          # 개발 서버 실행 (backend + frontend)
make test         # 전체 테스트 실행
make lint         # 린트 검사
make build        # 프로덕션 빌드
make migrate      # DB 마이그레이션
make generate     # sqlc 코드 생성
make deploy       # CDK 배포
```

---

## 7. Mantine UI 활용 계획

### 7.1 핵심 컴포넌트 매핑

| RealWorld 기능 | Mantine 컴포넌트 |
|----------------|------------------|
| 네비게이션 | AppShell, NavLink, Burger |
| 아티클 카드 | Card, Badge, Avatar, Group |
| 아티클 에디터 | Textarea, TextInput, MultiSelect |
| 댓글 | Paper, Stack, ActionIcon |
| 폼 (로그인/가입) | TextInput, PasswordInput, Button |
| 태그 | Badge, Chip, Group |
| 페이지네이션 | Pagination |
| 피드 탭 | Tabs |
| 알림 | Notifications |
| 로딩 상태 | Skeleton, Loader |

### 7.2 테마 설정

```typescript
// 예상 테마 구성
const theme = createTheme({
  primaryColor: 'green',  // Medium 스타일 녹색
  fontFamily: 'Inter, sans-serif',
  headings: {
    fontFamily: 'Inter, sans-serif',
  },
});
```

---

## 8. 다음 단계

### 8.1 PRD 작성 시 포함할 내용

1. **상세 사용자 스토리**: 각 기능별 상세 시나리오
2. **API 스키마 정의**: 요청/응답 JSON 스키마
3. **에러 처리 명세**: 에러 코드 및 메시지 정의
4. **UI/UX 와이어프레임**: 주요 페이지 레이아웃
5. **테스트 전략**: 단위/통합/E2E 테스트 계획
6. **배포 전략**: 환경별 배포 파이프라인

### 8.2 구현 우선순위 (제안)

1. **Phase 1**: 프로젝트 셋업 및 인프라 구성
2. **Phase 2**: 인증 시스템 (회원가입/로그인)
3. **Phase 3**: 아티클 CRUD
4. **Phase 4**: 댓글 및 태그
5. **Phase 5**: 프로필 및 팔로우
6. **Phase 6**: 피드 및 좋아요
7. **Phase 7**: 최종 테스트 및 배포

---

## 9. 참고 자료

- [RealWorld Spec](https://realworld-docs.netlify.app/)
- [Agentic Coding - Armin Ronacher](https://lucumr.pocoo.org/2025/6/12/agentic-coding/)
- [Mantine UI Documentation](https://mantine.dev/)
- [AWS CDK Documentation](https://docs.aws.amazon.com/cdk/)
- [TanStack Query](https://tanstack.com/query)
- [TanStack Router](https://tanstack.com/router)
- [sqlc Documentation](https://sqlc.dev/)
