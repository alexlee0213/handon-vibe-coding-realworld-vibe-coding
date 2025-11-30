# Building a Full-Stack Application with AI-Assisted "Vibe Coding"

> A practical tutorial on building a RealWorld Conduit application using Claude Code

## Introduction

This tutorial documents the process of building a full-stack social blogging platform (RealWorld Conduit - a Medium.com clone) using AI-assisted development, often called "Vibe Coding" or "Agentic Coding." The project follows the [RealWorld specification](https://realworld-docs.netlify.app/) and applies principles from [Armin Ronacher's Agentic Coding article](https://lucumr.pocoo.org/2025/6/12/agentic-coding/).

### What You'll Learn

- How to effectively communicate with AI coding assistants
- Project planning and documentation strategies for AI collaboration
- Test-Driven Development (TDD) in AI-assisted workflows
- Managing complex full-stack projects with AI tools

### Tech Stack Overview

| Layer | Technologies |
|-------|-------------|
| **Backend** | Go 1.21+, net/http (stdlib), SQLite/PostgreSQL, Pure SQL |
| **Frontend** | React 18+, Vite, TypeScript, TanStack Router/Query, Mantine UI |
| **Infrastructure** | AWS CDK, GitHub Actions, GitHub Pages |

---

## Phase 1: Project Initialization and Planning

### 1.1 Creating the Pre-PRD Document

**Prompt Used:**
```
RealWorld 스펙 (https://realworld-docs.netlify.app/implementation-creation/introduction/) 을
바이브 코딩으로 구현하려고 해. 기술 스택은
https://lucumr.pocoo.org/2025/6/12/agentic-coding/ 을 참고해서 작성해줘.
다만, UI 구현은 Mantine UI를 사용하고 인프라는 AWS CDK,
CI/CD 파이프라인은 깃허브 액션을 사용해줘.

본격적인 요구사항을 작성하기에 앞서 요구사항 작성을 위해서 필요한 내용을
docs/pre-prd.md에 문서화해줘.
```

**Why This Approach:**
- Reference external specifications directly via URLs - AI can fetch and analyze them
- Specify customizations clearly (Mantine UI instead of other UI libraries)
- Request documentation *before* jumping into code - this creates a shared understanding

**Result:**
The AI created `docs/pre-prd.md` containing:
- Project overview and goals
- Tech stack analysis from the Agentic Coding article
- RealWorld API specification summary
- Architecture considerations

### 1.2 Creating the PRD (Product Requirements Document)

**Prompt Used:**
```
@docs/pre-prd.md와 https://lucumr.pocoo.org/2025/6/12/agentic-coding/ 를
기반으로 docs/prd.md 를 작성해줘.
```

**Refinement Prompt:**
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

**Why This Approach:**
- Build on previous documents using `@` file references
- Provide specific technical requirements to override AI defaults
- Iterate on documents rather than rewriting from scratch

**Key Lesson:**
> When the AI's initial output doesn't match your vision, provide concrete examples of what you want. The AI learns from your corrections.

### 1.3 Creating the Implementation Plan

**Prompt Used:**
```
@docs/prd.md를 바탕으로, 프로젝트를 어떻게 구현할지
단계별 작업 계획(docs/plan.md) 을 세워줘.
```

**Result:**
A comprehensive 7-phase plan was created:
- Phase 1: Project Setup (backend, frontend, infra, CI/CD)
- Phase 2: Authentication System
- Phase 3: Article System
- Phase 4: Tag/Comment System
- Phase 5: Profile System
- Phase 6: Feed/Favorites System
- Phase 7: Deployment and Documentation

---

## Phase 2: Automating Project Management

### 2.1 Registering Tasks as GitHub Issues

**Prompt Used:**
```
@docs/plan.md의 태스크들을 깃허브 이슈에 등록해줘.
```

**Why This Works:**
- The AI reads the plan document and understands the task structure
- It creates properly labeled GitHub issues with descriptions
- Maintains traceability between documentation and actual work items

**Result:**
28 GitHub issues were created across 7 phases, with labels like `phase:1-setup`, `backend`, `frontend`, `infra`.

### 2.2 Implementing Issues Sequentially

**Prompt Pattern:**
```
do implement next git issue
```

This simple prompt became the primary workflow driver. The AI:
1. Checks open GitHub issues
2. Identifies the next logical issue to implement
3. Implements the required changes
4. Runs tests
5. Commits and closes the issue

**Key Lesson:**
> Create a consistent prompt pattern for repetitive workflows. The AI learns your project's conventions and applies them automatically.

---

## Phase 3: Setting Up Development Infrastructure

### 3.1 Choosing Git Hooks (Husky vs Lefthook)

**Prompt Used:**
```
git hook에서 프론트와 백엔드 각각의 변경이 있을경우
lint, unit test를 실행하도록 되어 있는지 확인 해줘
```

**Follow-up:**
```
husky와 lefthook 비교하면, 어떤 게 더 지금 프로젝트에 유리한지
의견 제시해줘
```

**Decision:**
```
go with Lefthook
```

**Why This Approach:**
- First, ask the AI to verify current state
- Request a comparison before making decisions
- Make explicit decisions so the AI can proceed

**Result:**
Lefthook was chosen for this Go+Node.js monorepo because:
- Better polyglot support (Go + JavaScript)
- Glob pattern filtering for selective execution
- YAML configuration (no JavaScript runtime for config)

**Lefthook Configuration Created:**
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

### 3.2 Documenting Decisions

**Prompt Used:**
```
위에서 husky와 Lefthook 비교한 적이 있는데, 비교 결과와 왜 Lefthook 이
이 프로젝트에 더 적합한지를 정리해서 docs 폴더 아래 md 파일로 저장해줘
```

**Key Lesson:**
> Ask the AI to document its reasoning. This creates valuable project documentation and helps onboard future contributors.

---

## Phase 4: Backend Implementation with TDD

### 4.1 Establishing TDD Principles

After initial implementation, the user requested stricter TDD adherence:

**Prompt Used:**
```
백엔드 구현은 TDD 적용해서 진행하는 걸 원칙으로 한다는 내용을 CLAUDE.md에 추가해줘.
```

**Result:**
Guidelines were added to CLAUDE.md specifying the TDD cycle:
1. **Red**: Write failing tests first
2. **Green**: Write minimal code to pass tests
3. **Refactor**: Clean up while keeping tests green

### 4.2 TDD in Practice

**Prompt Used:**
```
issue #7에 대해 TDD 적용해서 다시 진행해줘
```

**TDD Workflow Example:**

**Step 1 - Red Phase (Write Failing Tests):**
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

**Step 2 - Green Phase (Minimal Implementation):**
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
- Extract common patterns
- Add proper error handling
- Improve code organization

**Key Lesson:**
> When adopting TDD with AI assistance, explicitly state the expectation. The AI will follow the Red-Green-Refactor cycle systematically.

---

## Phase 5: Frontend Development with Visual Verification

### 5.1 Establishing Playwright MCP Verification

**Prompt Used:**
```
프론트엔드 개발을 할 때 Playwright MCP를 사용해서 구현 상태를 확인해야 한다는
내용도 추가해줘.
```

**Result:**
Frontend development guidelines were added requiring:
- `browser_snapshot` after implementing new components
- `browser_click`/`browser_fill_form` for interaction testing
- Visual verification before closing UI-related issues

### 5.2 Frontend Verification Workflow

**Required Checks Before Closing Frontend Issues:**
```
□ typecheck passes
□ lint passes
□ Playwright MCP browser test performed
  □ Page renders correctly (browser_snapshot)
  □ Key UI elements visible
  □ Interactions work (if applicable)
□ Commit and close issue
```

**Key Lesson:**
> AI can verify visual implementations using browser automation tools. Establish this as a requirement to catch UI issues early.

---

## Phase 6: Handling Context Continuations

### 6.1 Session Continuations

When Claude Code runs out of context, it automatically summarizes the conversation:

**Automatic Summary Structure:**
```
This session is being continued from a previous conversation...

Analysis:
[Chronological breakdown of what was accomplished]

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

**Continuation Prompt:**
```
Please continue the conversation from where we left it off
without asking the user any further questions.
Continue with the last task that you were asked to work on.
```

**Key Lesson:**
> Claude Code maintains context across sessions through summaries. Trust the continuation process - it preserves important details.

---

## Phase 7: Issue Selection and Prioritization

### 7.1 Checking Issue Status

**Prompt Used:**
```
git issues 중에서 아직 open 인 겻 3개만 가져와서 보여줘.
제일 먼저 처리해야 하는 순서로
```

**Result:**
AI lists issues prioritized by:
- Phase number (earlier phases first)
- Dependencies (prerequisites before dependents)
- Logical grouping

### 7.2 Selective Implementation

**Prompt Used:**
```
plz select one issue with high priority from the open issues in git
and implement it.
```

**AI's Decision Process:**
1. Fetches all open issues from GitHub
2. Analyzes phase numbers and dependencies
3. Selects the most impactful issue
4. Explains the selection reasoning
5. Implements the solution

---

## Common Patterns and Best Practices

### Effective Prompt Patterns

| Pattern | Example | Use Case |
|---------|---------|----------|
| **Reference Files** | `@docs/plan.md` | Include file content in context |
| **Sequential Action** | `do implement next git issue` | Automate repetitive workflows |
| **Verify First** | `해결되었는지 검증해줘` | Check state before proceeding |
| **Compare Options** | `husky와 lefthook 비교` | Make informed decisions |
| **Document Decisions** | `정리해서 md 파일로 저장해줘` | Create project documentation |
| **Explicit Commit** | `do git commit` | Control when commits happen |

### Error Handling Patterns

**When Tests Fail:**
```
The AI will automatically:
1. Read the error message
2. Identify the root cause
3. Fix the issue
4. Re-run tests
5. Continue if all pass
```

**When Edits Conflict:**
```
If the AI encounters "Found 2 matches" errors, it will:
1. Provide more unique context around the target
2. Retry the edit with better specificity
```

### Git Workflow Patterns

**After Each Task:**
```bash
# AI automatically:
git add -A
git commit -m "feat(scope): descriptive message"
git push
gh issue close #N --comment "Completed in commit xyz"
```

---

## Key Takeaways

1. **Documentation First**: Create PRD and implementation plans before coding
2. **Small Tasks**: Break work into atomic issues that can be completed in one session
3. **Explicit Guidelines**: Add development principles (TDD, visual verification) to CLAUDE.md
4. **Trust but Verify**: Let AI make decisions, but verify through tests and visual checks
5. **Document Decisions**: Ask AI to document comparisons and architectural choices
6. **Consistent Patterns**: Use repeatable prompts for common workflows
7. **Context Continuity**: Trust session continuations - they preserve essential context

---

## Conclusion

AI-assisted development ("Vibe Coding") is most effective when you:
- Provide clear specifications and constraints
- Establish consistent workflows and patterns
- Document decisions and guidelines in project files
- Verify implementations through automated tests and visual checks
- Trust the AI to handle routine tasks while you focus on decisions

The RealWorld Conduit project demonstrates that complex full-stack applications can be built efficiently with AI assistance, provided you establish clear communication patterns and verification processes.

---

## References

- [RealWorld Specification](https://realworld-docs.netlify.app/)
- [Agentic Coding - Armin Ronacher](https://lucumr.pocoo.org/2025/6/12/agentic-coding/)
- [Claude Code Documentation](https://docs.anthropic.com/claude-code)
- [Mantine UI](https://mantine.dev/)
- [TanStack Router](https://tanstack.com/router)
- [TanStack Query](https://tanstack.com/query)
