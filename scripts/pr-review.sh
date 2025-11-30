#!/bin/bash
set -e

# =============================================================================
# PR Review Automation Script
# =============================================================================
# Usage: ./scripts/pr-review.sh [OPTIONS] <PR_NUMBER>
#
# Options:
#   --dry-run    Run checks but don't submit review to GitHub
#   --help       Show this help message
#
# This script:
# 1. Fetches PR information from GitHub
# 2. Creates a git worktree for the PR branch
# 3. Runs static analysis and tests
# 4. Approves or requests changes based on results
# =============================================================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
WORKTREE_BASE="$(dirname "$PROJECT_DIR")"
REPO_NAME="$(basename "$PROJECT_DIR")"

# Options
DRY_RUN=false

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_dry_run() {
    echo -e "${CYAN}[DRY-RUN]${NC} $1"
}

show_help() {
    cat << EOF
Usage: $0 [OPTIONS] <PR_NUMBER>

Options:
  --dry-run    Run checks but don't submit review to GitHub
  --help       Show this help message

Examples:
  $0 33              # Review PR #33 and submit results
  $0 --dry-run 33    # Review PR #33 without submitting
EOF
    exit 0
}

cleanup() {
    local worktree_path="$1"
    if [ -d "$worktree_path" ]; then
        log_info "Cleaning up worktree: $worktree_path"
        cd "$PROJECT_DIR"
        git worktree remove "$worktree_path" --force 2>/dev/null || true
    fi
}

# Parse arguments
PR_NUMBER=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help|-h)
            show_help
            ;;
        -*)
            log_error "Unknown option: $1"
            show_help
            ;;
        *)
            PR_NUMBER="$1"
            shift
            ;;
    esac
done

# Check PR number
if [ -z "$PR_NUMBER" ]; then
    log_error "PR number is required"
    show_help
fi

if $DRY_RUN; then
    echo ""
    log_dry_run "Running in DRY-RUN mode - no review will be submitted"
    echo ""
fi

# =============================================================================
# Step 1: Fetch PR Information
# =============================================================================
log_info "Fetching PR #${PR_NUMBER} information..."

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    log_error "GitHub CLI (gh) is not installed. Please install it first."
    exit 1
fi

# Check gh auth status
if ! gh auth status &> /dev/null; then
    log_error "Not authenticated with GitHub. Run 'gh auth login' first."
    exit 1
fi

# Get PR details
PR_JSON=$(gh pr view "$PR_NUMBER" --json title,headRefName,baseRefName,state,author,body,files,additions,deletions 2>/dev/null)
if [ $? -ne 0 ]; then
    log_error "Failed to fetch PR #${PR_NUMBER}. Does it exist?"
    exit 1
fi

PR_TITLE=$(echo "$PR_JSON" | jq -r '.title')
PR_BRANCH=$(echo "$PR_JSON" | jq -r '.headRefName')
PR_BASE=$(echo "$PR_JSON" | jq -r '.baseRefName')
PR_STATE=$(echo "$PR_JSON" | jq -r '.state')
PR_AUTHOR=$(echo "$PR_JSON" | jq -r '.author.login')
PR_ADDITIONS=$(echo "$PR_JSON" | jq -r '.additions')
PR_DELETIONS=$(echo "$PR_JSON" | jq -r '.deletions')

echo ""
echo "=============================================="
echo "PR #${PR_NUMBER}: ${PR_TITLE}"
echo "=============================================="
echo "Author:    ${PR_AUTHOR}"
echo "Branch:    ${PR_BRANCH} -> ${PR_BASE}"
echo "State:     ${PR_STATE}"
echo "Changes:   +${PR_ADDITIONS} / -${PR_DELETIONS}"
echo "=============================================="
echo ""

# Check PR state
if [ "$PR_STATE" != "OPEN" ]; then
    log_warning "PR #${PR_NUMBER} is not open (state: ${PR_STATE}). Skipping."
    exit 0
fi

# =============================================================================
# Step 2: Create Git Worktree
# =============================================================================
WORKTREE_PATH="${WORKTREE_BASE}/${REPO_NAME}-pr-${PR_NUMBER}"

log_info "Creating worktree at: ${WORKTREE_PATH}"

# Clean up existing worktree if exists
if [ -d "$WORKTREE_PATH" ]; then
    log_warning "Worktree already exists. Removing..."
    cleanup "$WORKTREE_PATH"
fi

# Fetch the PR branch
cd "$PROJECT_DIR"
git fetch origin "pull/${PR_NUMBER}/head:pr-${PR_NUMBER}" 2>/dev/null || \
    git fetch origin "${PR_BRANCH}:pr-${PR_NUMBER}" 2>/dev/null || \
    { log_error "Failed to fetch PR branch"; exit 1; }

# Create worktree
git worktree add "$WORKTREE_PATH" "pr-${PR_NUMBER}"
log_success "Worktree created successfully"

# =============================================================================
# Step 3: Run Static Analysis and Tests
# =============================================================================
cd "$WORKTREE_PATH"

REVIEW_RESULT="APPROVE"
REVIEW_COMMENTS=""
FAILED_CHECKS=()

# Function to run a check
run_check() {
    local check_name="$1"
    local check_cmd="$2"
    local check_dir="$3"

    log_info "Running: ${check_name}..."

    if [ -n "$check_dir" ] && [ -d "$check_dir" ]; then
        cd "$check_dir"
    fi

    if eval "$check_cmd" > /tmp/check_output_$$.txt 2>&1; then
        log_success "${check_name} passed"
        cd "$WORKTREE_PATH"
        return 0
    else
        log_error "${check_name} failed"
        FAILED_CHECKS+=("$check_name")
        REVIEW_COMMENTS="${REVIEW_COMMENTS}\n### ❌ ${check_name} Failed\n\`\`\`\n$(tail -50 /tmp/check_output_$$.txt)\n\`\`\`\n"
        cd "$WORKTREE_PATH"
        return 1
    fi
}

echo ""
log_info "Starting code review checks..."
echo ""

# Backend checks (if backend directory exists)
if [ -d "backend" ]; then
    log_info "=== Backend Checks ==="

    # Go vet
    run_check "Go Vet" "go vet ./..." "backend" || REVIEW_RESULT="REQUEST_CHANGES"

    # Go fmt check
    run_check "Go Format Check" "test -z \"\$(gofmt -l .)\"" "backend" || REVIEW_RESULT="REQUEST_CHANGES"

    # Go tests
    run_check "Go Tests" "go test ./..." "backend" || REVIEW_RESULT="REQUEST_CHANGES"
fi

# Frontend checks (if frontend directory exists)
if [ -d "frontend" ]; then
    log_info "=== Frontend Checks ==="

    # Install dependencies if needed
    if [ ! -d "frontend/node_modules" ]; then
        log_info "Installing frontend dependencies..."
        cd frontend && npm ci --silent && cd ..
    fi

    # TypeScript check
    run_check "TypeScript Check" "npm run typecheck" "frontend" || REVIEW_RESULT="REQUEST_CHANGES"

    # ESLint
    run_check "ESLint" "npm run lint" "frontend" || REVIEW_RESULT="REQUEST_CHANGES"

    # Vitest tests
    run_check "Frontend Tests" "npm run test -- --run" "frontend" || REVIEW_RESULT="REQUEST_CHANGES"
fi

# Infrastructure checks (if infra directory exists)
if [ -d "infra" ]; then
    log_info "=== Infrastructure Checks ==="

    if [ ! -d "infra/node_modules" ]; then
        log_info "Installing infra dependencies..."
        cd infra && npm ci --silent && cd ..
    fi

    # CDK synth check
    if [ -f "infra/package.json" ]; then
        run_check "CDK Synth" "npx cdk synth --quiet" "infra" || REVIEW_RESULT="REQUEST_CHANGES"
    fi
fi

echo ""

# =============================================================================
# Step 4: Submit Review
# =============================================================================
if [ "$REVIEW_RESULT" == "APPROVE" ]; then
    REVIEW_BODY="## ✅ Automated Review Passed

All checks have passed successfully:

### Checks Performed
- Go Vet
- Go Format
- Go Tests
- TypeScript Check
- ESLint
- Frontend Tests

**Recommendation:** Ready to merge.

---
*This review was generated automatically by pr-review.sh*"

    if $DRY_RUN; then
        log_dry_run "Would APPROVE PR #${PR_NUMBER} with the following review:"
        echo ""
        echo "$REVIEW_BODY"
        echo ""
    else
        log_info "Submitting review..."
        gh pr review "$PR_NUMBER" --approve --body "$REVIEW_BODY"
        log_success "PR #${PR_NUMBER} approved!"
    fi

else
    REVIEW_BODY="## ❌ Automated Review Failed

Some checks have failed. Please address the following issues:

### Failed Checks
$(for check in "${FAILED_CHECKS[@]}"; do echo "- ❌ $check"; done)

### Details
${REVIEW_COMMENTS}

---
*This review was generated automatically by pr-review.sh*"

    if $DRY_RUN; then
        log_dry_run "Would REQUEST CHANGES on PR #${PR_NUMBER} with the following review:"
        echo ""
        echo -e "$REVIEW_BODY"
        echo ""
    else
        log_info "Submitting review..."
        gh pr review "$PR_NUMBER" --request-changes --body "$(echo -e "$REVIEW_BODY")"
        log_warning "PR #${PR_NUMBER} needs changes"
    fi
fi

# =============================================================================
# Step 5: Cleanup
# =============================================================================
log_info "Cleaning up..."
cd "$PROJECT_DIR"
cleanup "$WORKTREE_PATH"

# Delete the temporary branch
git branch -D "pr-${PR_NUMBER}" 2>/dev/null || true

echo ""
if [ "$REVIEW_RESULT" == "APPROVE" ]; then
    log_success "PR #${PR_NUMBER} review complete - APPROVED"
    echo ""
    echo "To merge the PR, run:"
    echo "  gh pr merge ${PR_NUMBER} --squash --delete-branch"
else
    log_warning "PR #${PR_NUMBER} review complete - CHANGES REQUESTED"
    echo ""
    echo "Failed checks: ${FAILED_CHECKS[*]}"
fi

exit 0
