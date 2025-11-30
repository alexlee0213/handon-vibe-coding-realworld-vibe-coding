#!/usr/bin/env bash
#
# PR Auto Review Script
# GitHub PR을 자동으로 리뷰하고 정적 분석 및 테스트를 수행합니다.
#
# Usage: ./scripts/pr-review.sh <PR_NUMBER> [OPTIONS]
#
# Options:
#   --dry-run       실제 승인/거부 없이 검증만 수행
#   --skip-backend  백엔드 검증 건너뛰기
#   --skip-frontend 프론트엔드 검증 건너뛰기
#   --auto-merge    승인 후 자동 머지 (squash)
#   --verbose       상세 로그 출력
#
# Examples:
#   ./scripts/pr-review.sh 42
#   ./scripts/pr-review.sh 42 --dry-run
#   ./scripts/pr-review.sh 42 --auto-merge --verbose
#

set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKTREE_BASE="${PROJECT_ROOT}/.worktrees"
LOG_FILE="${PROJECT_ROOT}/pr-review.log"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Options (defaults)
DRY_RUN=false
SKIP_BACKEND=false
SKIP_FRONTEND=false
AUTO_MERGE=false
VERBOSE=false

# ============================================================================
# Helper Functions
# ============================================================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    case "$level" in
        INFO)  echo -e "${BLUE}[INFO]${NC} $message" ;;
        OK)    echo -e "${GREEN}[OK]${NC} $message" ;;
        WARN)  echo -e "${YELLOW}[WARN]${NC} $message" ;;
        ERROR) echo -e "${RED}[ERROR]${NC} $message" ;;
        STEP)  echo -e "${CYAN}==>${NC} $message" ;;
    esac

    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

verbose() {
    if [[ "$VERBOSE" == true ]]; then
        log INFO "$@"
    fi
}

die() {
    log ERROR "$@"
    exit 1
}

cleanup() {
    local exit_code=$?
    if [[ -n "${WORKTREE_PATH:-}" && -d "$WORKTREE_PATH" ]]; then
        log INFO "Cleaning up worktree..."
        cd "$PROJECT_ROOT"
        git worktree remove --force "$WORKTREE_PATH" 2>/dev/null || true
    fi
    exit $exit_code
}

trap cleanup EXIT

usage() {
    cat << EOF
Usage: $(basename "$0") <PR_NUMBER> [OPTIONS]

GitHub PR을 자동으로 리뷰하고 정적 분석 및 테스트를 수행합니다.

Options:
  --dry-run       실제 승인/거부 없이 검증만 수행
  --skip-backend  백엔드 검증 건너뛰기
  --skip-frontend 프론트엔드 검증 건너뛰기
  --auto-merge    승인 후 자동 머지 (squash)
  --verbose       상세 로그 출력
  -h, --help      이 도움말 표시

Examples:
  $(basename "$0") 42
  $(basename "$0") 42 --dry-run
  $(basename "$0") 42 --auto-merge --verbose
EOF
}

# ============================================================================
# Validation Functions
# ============================================================================

check_dependencies() {
    log STEP "Checking dependencies..."

    local missing=()

    command -v gh >/dev/null 2>&1 || missing+=("gh (GitHub CLI)")
    command -v git >/dev/null 2>&1 || missing+=("git")
    command -v jq >/dev/null 2>&1 || missing+=("jq")
    command -v go >/dev/null 2>&1 || missing+=("go")
    command -v npm >/dev/null 2>&1 || missing+=("npm")

    if [[ ${#missing[@]} -gt 0 ]]; then
        die "Missing dependencies: ${missing[*]}"
    fi

    # Check gh auth status
    if ! gh auth status >/dev/null 2>&1; then
        die "GitHub CLI not authenticated. Run: gh auth login"
    fi

    log OK "All dependencies available"
}

# ============================================================================
# PR Information
# ============================================================================

fetch_pr_info() {
    local pr_number="$1"
    log STEP "Fetching PR #${pr_number} information..."

    PR_JSON=$(gh pr view "$pr_number" --json number,title,state,headRefName,baseRefName,author,mergeable,reviewDecision,additions,deletions,changedFiles,commits 2>/dev/null) || {
        die "Failed to fetch PR #${pr_number}. Check if PR exists."
    }

    PR_TITLE=$(echo "$PR_JSON" | jq -r '.title')
    PR_STATE=$(echo "$PR_JSON" | jq -r '.state')
    PR_HEAD_BRANCH=$(echo "$PR_JSON" | jq -r '.headRefName')
    PR_BASE_BRANCH=$(echo "$PR_JSON" | jq -r '.baseRefName')
    PR_AUTHOR=$(echo "$PR_JSON" | jq -r '.author.login')
    PR_MERGEABLE=$(echo "$PR_JSON" | jq -r '.mergeable')
    PR_ADDITIONS=$(echo "$PR_JSON" | jq -r '.additions')
    PR_DELETIONS=$(echo "$PR_JSON" | jq -r '.deletions')
    PR_CHANGED_FILES=$(echo "$PR_JSON" | jq -r '.changedFiles')
    PR_COMMITS=$(echo "$PR_JSON" | jq -r '.commits | length')

    log OK "PR #${pr_number}: ${PR_TITLE}"
    log INFO "  Author: ${PR_AUTHOR}"
    log INFO "  Branch: ${PR_HEAD_BRANCH} → ${PR_BASE_BRANCH}"
    log INFO "  State: ${PR_STATE}"
    log INFO "  Changes: +${PR_ADDITIONS} -${PR_DELETIONS} (${PR_CHANGED_FILES} files, ${PR_COMMITS} commits)"
    log INFO "  Mergeable: ${PR_MERGEABLE}"

    if [[ "$PR_STATE" != "OPEN" ]]; then
        die "PR is not open (state: ${PR_STATE})"
    fi
}

get_changed_paths() {
    local pr_number="$1"
    log STEP "Analyzing changed files..."

    CHANGED_FILES_LIST=$(gh pr diff "$pr_number" --name-only 2>/dev/null) || {
        die "Failed to get changed files list"
    }

    HAS_BACKEND_CHANGES=false
    HAS_FRONTEND_CHANGES=false
    HAS_INFRA_CHANGES=false

    while IFS= read -r file; do
        case "$file" in
            backend/*) HAS_BACKEND_CHANGES=true ;;
            frontend/*) HAS_FRONTEND_CHANGES=true ;;
            infra/*) HAS_INFRA_CHANGES=true ;;
        esac
    done <<< "$CHANGED_FILES_LIST"

    log INFO "  Backend changes: $HAS_BACKEND_CHANGES"
    log INFO "  Frontend changes: $HAS_FRONTEND_CHANGES"
    log INFO "  Infra changes: $HAS_INFRA_CHANGES"
}

# ============================================================================
# Git Worktree
# ============================================================================

setup_worktree() {
    local pr_number="$1"
    log STEP "Setting up git worktree for PR #${pr_number}..."

    # Ensure worktree base directory exists
    mkdir -p "$WORKTREE_BASE"

    WORKTREE_PATH="${WORKTREE_BASE}/pr-${pr_number}"

    # Clean up existing worktree if present
    if [[ -d "$WORKTREE_PATH" ]]; then
        log WARN "Removing existing worktree at ${WORKTREE_PATH}"
        git worktree remove --force "$WORKTREE_PATH" 2>/dev/null || true
    fi

    # Fetch the PR branch
    log INFO "Fetching PR branch..."
    git fetch origin "pull/${pr_number}/head:pr-${pr_number}" --force 2>/dev/null || {
        die "Failed to fetch PR branch"
    }

    # Create worktree
    git worktree add "$WORKTREE_PATH" "pr-${pr_number}" 2>/dev/null || {
        die "Failed to create worktree"
    }

    log OK "Worktree created at ${WORKTREE_PATH}"
}

# ============================================================================
# Code Review & Analysis
# ============================================================================

REVIEW_ERRORS=()
REVIEW_WARNINGS=()

run_backend_checks() {
    if [[ "$SKIP_BACKEND" == true ]]; then
        log WARN "Skipping backend checks (--skip-backend)"
        return 0
    fi

    if [[ "$HAS_BACKEND_CHANGES" != true ]]; then
        log INFO "No backend changes detected, skipping backend checks"
        return 0
    fi

    log STEP "Running backend checks..."
    cd "${WORKTREE_PATH}/backend"

    # Go mod tidy check
    verbose "Checking go.mod consistency..."
    if ! go mod tidy 2>&1; then
        REVIEW_ERRORS+=("Backend: go mod tidy failed")
    fi

    # Check if go.mod/go.sum changed (means dependencies weren't tidied)
    if git diff --quiet go.mod go.sum 2>/dev/null; then
        verbose "go.mod is consistent"
    else
        REVIEW_WARNINGS+=("Backend: go.mod/go.sum may need 'go mod tidy'")
    fi

    # Go fmt check
    verbose "Checking Go formatting..."
    local fmt_output
    fmt_output=$(gofmt -l . 2>&1) || true
    if [[ -n "$fmt_output" ]]; then
        REVIEW_ERRORS+=("Backend: Files need formatting: $(echo "$fmt_output" | tr '\n' ' ')")
    else
        log OK "Backend: Go formatting OK"
    fi

    # Go vet
    verbose "Running go vet..."
    if ! go vet ./... 2>&1; then
        REVIEW_ERRORS+=("Backend: go vet found issues")
    else
        log OK "Backend: go vet OK"
    fi

    # Go build
    verbose "Building backend..."
    if ! go build -o /dev/null ./cmd/server/main.go 2>&1; then
        REVIEW_ERRORS+=("Backend: Build failed")
    else
        log OK "Backend: Build OK"
    fi

    # Go test
    verbose "Running backend tests..."
    if ! go test -v ./... 2>&1; then
        REVIEW_ERRORS+=("Backend: Tests failed")
    else
        log OK "Backend: Tests passed"
    fi

    cd "$PROJECT_ROOT"
}

run_frontend_checks() {
    if [[ "$SKIP_FRONTEND" == true ]]; then
        log WARN "Skipping frontend checks (--skip-frontend)"
        return 0
    fi

    if [[ "$HAS_FRONTEND_CHANGES" != true ]]; then
        log INFO "No frontend changes detected, skipping frontend checks"
        return 0
    fi

    log STEP "Running frontend checks..."
    cd "${WORKTREE_PATH}/frontend"

    # Install dependencies
    verbose "Installing frontend dependencies..."
    if ! npm ci --silent 2>&1; then
        REVIEW_ERRORS+=("Frontend: npm install failed")
        cd "$PROJECT_ROOT"
        return 1
    fi

    # TypeScript check
    verbose "Running TypeScript check..."
    if ! npm run typecheck 2>&1; then
        REVIEW_ERRORS+=("Frontend: TypeScript errors found")
    else
        log OK "Frontend: TypeScript OK"
    fi

    # ESLint
    verbose "Running ESLint..."
    if ! npm run lint 2>&1; then
        REVIEW_ERRORS+=("Frontend: ESLint errors found")
    else
        log OK "Frontend: ESLint OK"
    fi

    # Build check
    verbose "Building frontend..."
    if ! npm run build 2>&1; then
        REVIEW_ERRORS+=("Frontend: Build failed")
    else
        log OK "Frontend: Build OK"
    fi

    # Tests
    verbose "Running frontend tests..."
    if ! npm run test -- --run 2>&1; then
        REVIEW_ERRORS+=("Frontend: Tests failed")
    else
        log OK "Frontend: Tests passed"
    fi

    cd "$PROJECT_ROOT"
}

# ============================================================================
# PR Review Actions
# ============================================================================

generate_review_body() {
    local status="$1"
    local body=""

    body+="## 🤖 Automated PR Review\n\n"
    body+="**PR:** #${PR_NUMBER} - ${PR_TITLE}\n"
    body+="**Branch:** \`${PR_HEAD_BRANCH}\` → \`${PR_BASE_BRANCH}\`\n"
    body+="**Changes:** +${PR_ADDITIONS} -${PR_DELETIONS} (${PR_CHANGED_FILES} files)\n\n"

    if [[ "$status" == "APPROVE" ]]; then
        body+="### ✅ All checks passed\n\n"
        body+="| Check | Status |\n"
        body+="|-------|--------|\n"

        if [[ "$HAS_BACKEND_CHANGES" == true && "$SKIP_BACKEND" != true ]]; then
            body+="| Backend Format | ✅ Pass |\n"
            body+="| Backend Vet | ✅ Pass |\n"
            body+="| Backend Build | ✅ Pass |\n"
            body+="| Backend Tests | ✅ Pass |\n"
        fi

        if [[ "$HAS_FRONTEND_CHANGES" == true && "$SKIP_FRONTEND" != true ]]; then
            body+="| Frontend TypeScript | ✅ Pass |\n"
            body+="| Frontend ESLint | ✅ Pass |\n"
            body+="| Frontend Build | ✅ Pass |\n"
            body+="| Frontend Tests | ✅ Pass |\n"
        fi

        body+="\n✨ This PR is ready to merge!\n"
    else
        body+="### ❌ Some checks failed\n\n"
        body+="Please fix the following issues:\n\n"

        for error in "${REVIEW_ERRORS[@]}"; do
            body+="- ❌ ${error}\n"
        done

        if [[ ${#REVIEW_WARNINGS[@]} -gt 0 ]]; then
            body+="\n**Warnings:**\n"
            for warning in "${REVIEW_WARNINGS[@]}"; do
                body+="- ⚠️ ${warning}\n"
            done
        fi

        body+="\n---\n"
        body+="Fix the issues and push again to trigger a new review.\n"
    fi

    body+="\n---\n"
    body+="_Automated review by pr-review.sh_"

    echo -e "$body"
}

submit_review() {
    local pr_number="$1"
    local status="$2"  # APPROVE or REQUEST_CHANGES

    if [[ "$DRY_RUN" == true ]]; then
        log WARN "[DRY RUN] Would submit review with status: ${status}"
        echo "--- Review Body ---"
        generate_review_body "$status"
        echo "--- End Review Body ---"
        return 0
    fi

    log STEP "Submitting PR review (${status})..."

    local review_body
    review_body=$(generate_review_body "$status")

    if [[ "$status" == "APPROVE" ]]; then
        echo -e "$review_body" | gh pr review "$pr_number" --approve --body-file - || {
            die "Failed to submit approval review"
        }
        log OK "PR #${pr_number} approved!"

        if [[ "$AUTO_MERGE" == true ]]; then
            log STEP "Auto-merging PR..."
            gh pr merge "$pr_number" --squash --delete-branch || {
                log WARN "Auto-merge failed. PR may require additional approvals or checks."
            }
        fi
    else
        echo -e "$review_body" | gh pr review "$pr_number" --request-changes --body-file - || {
            die "Failed to submit changes-requested review"
        }
        log WARN "PR #${pr_number} requires changes"
    fi
}

# ============================================================================
# Main
# ============================================================================

main() {
    # Parse arguments
    if [[ $# -lt 1 ]]; then
        usage
        exit 1
    fi

    PR_NUMBER=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --skip-backend)
                SKIP_BACKEND=true
                shift
                ;;
            --skip-frontend)
                SKIP_FRONTEND=true
                shift
                ;;
            --auto-merge)
                AUTO_MERGE=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            -*)
                die "Unknown option: $1"
                ;;
            *)
                if [[ -z "$PR_NUMBER" ]]; then
                    PR_NUMBER="$1"
                else
                    die "Unexpected argument: $1"
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$PR_NUMBER" ]]; then
        die "PR number is required"
    fi

    if ! [[ "$PR_NUMBER" =~ ^[0-9]+$ ]]; then
        die "Invalid PR number: ${PR_NUMBER}"
    fi

    # Initialize log
    echo "=== PR Review Started: $(date) ===" >> "$LOG_FILE"
    echo "PR #${PR_NUMBER}" >> "$LOG_FILE"

    log INFO "Starting automated PR review for PR #${PR_NUMBER}"
    [[ "$DRY_RUN" == true ]] && log WARN "Running in DRY RUN mode"

    # Run workflow
    cd "$PROJECT_ROOT"

    check_dependencies
    fetch_pr_info "$PR_NUMBER"
    get_changed_paths "$PR_NUMBER"
    setup_worktree "$PR_NUMBER"

    # Run checks
    run_backend_checks || true
    run_frontend_checks || true

    # Determine result
    log STEP "Review Summary"

    if [[ ${#REVIEW_ERRORS[@]} -eq 0 ]]; then
        log OK "All checks passed! ✨"
        submit_review "$PR_NUMBER" "APPROVE"
    else
        log ERROR "Found ${#REVIEW_ERRORS[@]} error(s):"
        for error in "${REVIEW_ERRORS[@]}"; do
            log ERROR "  - ${error}"
        done
        submit_review "$PR_NUMBER" "REQUEST_CHANGES"
        exit 1
    fi
}

main "$@"
