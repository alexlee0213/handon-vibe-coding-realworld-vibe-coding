# AI 기반 "바이브 코딩"으로 풀스택 애플리케이션 구축하기

> Claude Code를 활용한 RealWorld Conduit 애플리케이션 개발 실습 튜토리얼

## 소개

이 튜토리얼은 AI 기반 개발("바이브 코딩" 또는 "에이전틱 코딩")을 활용하여 풀스택 소셜 블로깅 플랫폼(RealWorld Conduit - Medium.com 클론)을 구축하는 과정을 문서화합니다. 이 프로젝트는 [RealWorld 명세](https://realworld-docs.netlify.app/)를 따르며, [Armin Ronacher의 Agentic Coding 아티클](https://lucumr.pocoo.org/2025/6/12/agentic-coding/)의 원칙을 적용합니다.

### 학습 내용

- AI 코딩 어시스턴트와 효과적으로 소통하는 방법
- AI 협업을 위한 프로젝트 계획 및 문서화 전략
- AI 기반 워크플로우에서의 테스트 주도 개발(TDD)
- AI 도구를 활용한 복잡한 풀스택 프로젝트 관리

### 기술 스택 개요

| 레이어 | 기술 |
|-------|-------------|
| **백엔드** | Go 1.21+, net/http (stdlib), SQLite/PostgreSQL, Pure SQL |
| **프론트엔드** | React 18+, Vite, TypeScript, TanStack Router/Query, Mantine UI |
| **인프라** | AWS CDK, GitHub Actions, GitHub Pages |

---

## Phase 1: 프로젝트 초기화 및 계획

### 1.1 Pre-PRD 문서 작성

**사용한 프롬프트:**
```
RealWorld 스펙 (https://realworld-docs.netlify.app/implementation-creation/introduction/) 을
바이브 코딩으로 구현하려고 해. 기술 스택은
https://lucumr.pocoo.org/2025/6/12/agentic-coding/ 을 참고해서 작성해줘.
다만, UI 구현은 Mantine UI를 사용하고 인프라는 AWS CDK,
CI/CD 파이프라인은 깃허브 액션을 사용해줘.

본격적인 요구사항을 작성하기에 앞서 요구사항 작성을 위해서 필요한 내용을
docs/pre-prd.md에 문서화해줘.
```

**이 접근법의 이유:**
- 외부 명세를 URL로 직접 참조 - AI가 이를 가져와서 분석할 수 있음
- 커스터마이징 사항을 명확하게 지정 (다른 UI 라이브러리 대신 Mantine UI)
- 코드 작성 *전에* 문서화 요청 - 공유된 이해 기반 형성

**결과:**
AI가 `docs/pre-prd.md`를 생성하여 다음 내용 포함:
- 프로젝트 개요 및 목표
- Agentic Coding 아티클의 기술 스택 분석
- RealWorld API 명세 요약
- 아키텍처 고려사항

### 1.2 PRD(제품 요구사항 문서) 작성

**사용한 프롬프트:**
```
@docs/pre-prd.md와 https://lucumr.pocoo.org/2025/6/12/agentic-coding/ 를
기반으로 docs/prd.md 를 작성해줘.
```

**개선 프롬프트:**
```
아래 내용 참조해서, prd 내용 보강해줘.
-----
3.1 프론트엔드 기술 스택
- 프레임워크: React with Vite
- 언어: TypeScript
- 라우터: Tanstack Router
- 상태 관리: Tanstack Query (서버 상태), Zustand (클라이언트 상태)
- UI 라이브러리: Mantine UI
...
```

**이 접근법의 이유:**
- `@` 파일 참조를 사용하여 이전 문서를 기반으로 구축
- AI 기본값을 재정의하기 위해 구체적인 기술 요구사항 제공
- 처음부터 다시 작성하는 대신 문서를 반복적으로 개선

**핵심 교훈:**
> AI의 초기 출력이 원하는 비전과 맞지 않을 때, 원하는 것의 구체적인 예시를 제공하세요. AI는 당신의 수정사항에서 학습합니다.

### 1.3 구현 계획 작성

**사용한 프롬프트:**
```
@docs/prd.md를 바탕으로, 프로젝트를 어떻게 구현할지
단계별 작업 계획(docs/plan.md) 을 세워줘.
```

**결과:**
포괄적인 7단계 계획이 작성됨:
- Phase 1: 프로젝트 셋업 (백엔드, 프론트엔드, 인프라, CI/CD)
- Phase 2: 인증 시스템
- Phase 3: 아티클 시스템
- Phase 4: 태그/댓글 시스템
- Phase 5: 프로필 시스템
- Phase 6: 피드/좋아요 시스템
- Phase 7: 배포 및 문서화

---

## Phase 2: 프로젝트 관리 자동화

### 2.1 태스크를 GitHub 이슈로 등록

**사용한 프롬프트:**
```
@docs/plan.md의 태스크들을 깃허브 이슈에 등록해줘.
```

**작동 원리:**
- AI가 계획 문서를 읽고 태스크 구조를 이해
- 설명이 포함된 레이블이 지정된 GitHub 이슈를 생성
- 문서와 실제 작업 항목 간의 추적성 유지

**결과:**
7개 Phase에 걸쳐 28개의 GitHub 이슈가 생성되었으며, `phase:1-setup`, `backend`, `frontend`, `infra` 등의 레이블이 적용됨.

### 2.2 이슈 순차적 구현

**프롬프트 패턴:**
```
do implement next git issue
```

이 간단한 프롬프트가 주요 워크플로우 드라이버가 됨. AI는:
1. 열린 GitHub 이슈 확인
2. 구현할 다음 논리적 이슈 식별
3. 필요한 변경사항 구현
4. 테스트 실행
5. 커밋하고 이슈 닫기

**핵심 교훈:**
> 반복적인 워크플로우를 위한 일관된 프롬프트 패턴을 만드세요. AI는 프로젝트의 관례를 학습하고 자동으로 적용합니다.

---

## Phase 3: 개발 인프라 설정

### 3.1 Git Hooks 선택 (Husky vs Lefthook)

**사용한 프롬프트:**
```
git hook에서 프론트와 백엔드 각각의 변경이 있을경우
lint, unit test를 실행하도록 되어 있는지 확인 해줘
```

**후속 프롬프트:**
```
husky와 lefthook 비교하면, 어떤 게 더 지금 프로젝트에 유리한지
의견 제시해줘
```

**결정:**
```
go with Lefthook
```

**이 접근법의 이유:**
- 먼저 AI에게 현재 상태 확인 요청
- 결정하기 전에 비교 요청
- AI가 진행할 수 있도록 명시적인 결정

**결과:**
이 Go+Node.js 모노레포에 Lefthook이 선택된 이유:
- 더 나은 폴리글랏 지원 (Go + JavaScript)
- 선택적 실행을 위한 Glob 패턴 필터링
- YAML 설정 (설정을 위한 JavaScript 런타임 불필요)

**생성된 Lefthook 설정:**
```yaml
# lefthook.yml
pre-commit:
  parallel: true
  commands:
    backend-fmt:
      glob: "backend/**/*.go"
      run: cd backend && go fmt ./...
      stage_fixed: true
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

### 3.2 결정사항 문서화

**사용한 프롬프트:**
```
위에서 husky와 Lefthook 비교한 적이 있는데, 비교 결과와 왜 Lefthook 이
이 프로젝트에 더 적합한지를 정리해서 docs 폴더 아래 md 파일로 저장해줘
```

**핵심 교훈:**
> AI에게 추론 과정을 문서화하도록 요청하세요. 이는 귀중한 프로젝트 문서를 생성하고 향후 기여자의 온보딩에 도움이 됩니다.

---

## Phase 4: TDD를 적용한 백엔드 구현

### 4.1 TDD 원칙 수립

초기 구현 후, 사용자가 더 엄격한 TDD 준수를 요청:

**사용한 프롬프트:**
```
백엔드 구현은 TDD 적용해서 진행하는 걸 원칙으로 한다는 내용을 CLAUDE.md에 추가해줘.
```

**결과:**
CLAUDE.md에 TDD 사이클을 명시하는 가이드라인 추가:
1. **Red**: 실패하는 테스트 먼저 작성
2. **Green**: 테스트를 통과하는 최소한의 코드 작성
3. **Refactor**: 테스트가 계속 통과하도록 유지하면서 정리

### 4.2 실제 TDD 적용

**사용한 프롬프트:**
```
issue #7에 대해 TDD 적용해서 다시 진행해줘
```

**TDD 워크플로우 예시:**

**Step 1 - Red 단계 (실패하는 테스트 작성):**
```go
// backend/internal/repository/user_test.go
func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)

    user := &domain.User{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "hashedpassword",
    }

    err := repo.Create(context.Background(), user)
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
}
```

**Step 2 - Green 단계 (최소한의 구현):**
```go
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
    result, err := r.db.ExecContext(ctx,
        `INSERT INTO users (username, email, password, created_at, updated_at)
         VALUES (?, ?, ?, ?, ?)`,
        user.Username, user.Email, user.Password, time.Now(), time.Now())
    if err != nil {
        return err
    }
    id, _ := result.LastInsertId()
    user.ID = id
    return nil
}
```

**Step 3 - Refactor:**
- 공통 패턴 추출
- 적절한 에러 처리 추가
- 코드 구조 개선

**핵심 교훈:**
> AI 지원과 함께 TDD를 채택할 때, 기대사항을 명시적으로 언급하세요. AI는 Red-Green-Refactor 사이클을 체계적으로 따릅니다.

---

## Phase 5: 시각적 검증을 통한 프론트엔드 개발

### 5.1 Playwright MCP 검증 수립

**사용한 프롬프트:**
```
프론트엔드 개발을 할 때 Playwright MCP를 사용해서 구현 상태를 확인해야 한다는
내용도 추가해줘.
```

**결과:**
프론트엔드 개발 가이드라인에 다음 요구사항 추가:
- 새 컴포넌트 구현 후 `browser_snapshot`
- 인터랙션 테스트를 위한 `browser_click`/`browser_fill_form`
- UI 관련 이슈 종료 전 시각적 검증

### 5.2 프론트엔드 검증 워크플로우

**프론트엔드 이슈 종료 전 필수 체크사항:**
```
□ typecheck 통과
□ lint 통과
□ Playwright MCP 브라우저 테스트 수행
  □ 페이지 렌더링 정상 확인 (browser_snapshot)
  □ 주요 UI 요소 표시 확인
  □ 인터랙션 동작 확인 (해당 시)
□ 커밋 및 이슈 닫기
```

**핵심 교훈:**
> AI는 브라우저 자동화 도구를 사용하여 시각적 구현을 검증할 수 있습니다. 이를 요구사항으로 설정하여 UI 문제를 조기에 발견하세요.

---

## Phase 6: 컨텍스트 연속성 처리

### 6.1 세션 연속

Claude Code가 컨텍스트가 부족해지면 자동으로 대화를 요약합니다:

**자동 요약 구조:**
```
This session is being continued from a previous conversation...

Analysis:
[완료된 작업의 시간순 분석]

Summary:
1. Primary Request and Intent
2. Key Technical Concepts
3. Files and Code Sections
4. Errors and fixes
5. Problem Solving
6. All user messages
7. Pending Tasks
8. Current Work
9. Optional Next Step
```

**연속 프롬프트:**
```
Please continue the conversation from where we left it off
without asking the user any further questions.
Continue with the last task that you were asked to work on.
```

**핵심 교훈:**
> Claude Code는 요약을 통해 세션 간 컨텍스트를 유지합니다. 연속 프로세스를 신뢰하세요 - 중요한 세부사항을 보존합니다.

---

## Phase 7: 이슈 선택 및 우선순위 지정

### 7.1 이슈 상태 확인

**사용한 프롬프트:**
```
git issues 중에서 아직 open 인 겻 3개만 가져와서 보여줘.
제일 먼저 처리해야 하는 순서로
```

**결과:**
AI가 다음 기준으로 우선순위를 정해 이슈 나열:
- Phase 번호 (이전 Phase 우선)
- 의존성 (선행 조건이 있는 것 먼저)
- 논리적 그룹화

### 7.2 선택적 구현

**사용한 프롬프트:**
```
plz select one issue with high priority from the open issues in git
and implement it.
```

**AI의 결정 프로세스:**
1. GitHub에서 모든 열린 이슈 가져오기
2. Phase 번호와 의존성 분석
3. 가장 영향력 있는 이슈 선택
4. 선택 이유 설명
5. 솔루션 구현

---

## 공통 패턴 및 모범 사례

### 효과적인 프롬프트 패턴

| 패턴 | 예시 | 사용 사례 |
|---------|---------|----------|
| **파일 참조** | `@docs/plan.md` | 컨텍스트에 파일 내용 포함 |
| **순차적 작업** | `do implement next git issue` | 반복 워크플로우 자동화 |
| **먼저 확인** | `해결되었는지 검증해줘` | 진행 전 상태 확인 |
| **옵션 비교** | `husky와 lefthook 비교` | 정보에 입각한 결정 |
| **결정 문서화** | `정리해서 md 파일로 저장해줘` | 프로젝트 문서 생성 |
| **명시적 커밋** | `do git commit` | 커밋 시점 제어 |

### 에러 처리 패턴

**테스트 실패 시:**
```
AI가 자동으로:
1. 에러 메시지 읽기
2. 근본 원인 식별
3. 문제 수정
4. 테스트 다시 실행
5. 모두 통과하면 계속 진행
```

**편집 충돌 시:**
```
AI가 "Found 2 matches" 에러를 만나면:
1. 대상 주변에 더 고유한 컨텍스트 제공
2. 더 나은 특정성으로 편집 재시도
```

### Git 워크플로우 패턴

**각 태스크 완료 후:**
```bash
# AI가 자동으로:
git add -A
git commit -m "feat(scope): descriptive message"
git push
gh issue close #N --comment "Completed in commit xyz"
```

---

## 핵심 요약

1. **문서화 우선**: 코딩 전에 PRD와 구현 계획 작성
2. **작은 태스크**: 작업을 한 세션에 완료할 수 있는 원자적 이슈로 분할
3. **명시적 가이드라인**: 개발 원칙(TDD, 시각적 검증)을 CLAUDE.md에 추가
4. **신뢰하되 검증**: AI가 결정하도록 하되, 테스트와 시각적 확인으로 검증
5. **결정 문서화**: AI에게 비교 및 아키텍처 선택 문서화 요청
6. **일관된 패턴**: 공통 워크플로우에 반복 가능한 프롬프트 사용
7. **컨텍스트 연속성**: 세션 연속을 신뢰 - 필수 컨텍스트 보존

---

## 결론

AI 기반 개발("바이브 코딩")은 다음을 수행할 때 가장 효과적입니다:
- 명확한 명세와 제약조건 제공
- 일관된 워크플로우와 패턴 수립
- 결정사항과 가이드라인을 프로젝트 파일에 문서화
- 자동화된 테스트와 시각적 확인으로 구현 검증
- 루틴 태스크는 AI에게 맡기고 당신은 결정에 집중

RealWorld Conduit 프로젝트는 명확한 커뮤니케이션 패턴과 검증 프로세스를 수립하면 AI 지원으로 복잡한 풀스택 애플리케이션을 효율적으로 구축할 수 있음을 보여줍니다.

---

## 참고 자료

- [RealWorld 명세](https://realworld-docs.netlify.app/)
- [Agentic Coding - Armin Ronacher](https://lucumr.pocoo.org/2025/6/12/agentic-coding/)
- [Claude Code 문서](https://docs.anthropic.com/claude-code)
- [Mantine UI](https://mantine.dev/)
- [TanStack Router](https://tanstack.com/router)
- [TanStack Query](https://tanstack.com/query)
