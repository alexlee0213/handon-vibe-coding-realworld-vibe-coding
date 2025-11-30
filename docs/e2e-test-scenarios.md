# E2E Test Scenarios

RealWorld Conduit 애플리케이션의 E2E 테스트 시나리오 문서입니다.
Playwright MCP를 사용하여 실제 브라우저에서 테스트를 수행합니다.

## 테스트 환경

- **Backend**: Go net/http server (port 8080)
- **Frontend**: React + Vite (port 5173)
- **Database**: SQLite (development)
- **Browser**: Chromium (via Playwright MCP)

## 테스트 시나리오

### Scenario 1: 회원가입 → 로그인 → 글 작성 → 댓글 → 로그아웃

#### 1.1 회원가입 (Registration)

**Steps:**
1. Navigate to `http://localhost:5173` → redirects to `/login`
2. Click "Sign up" link → navigates to `/register`
3. Fill registration form:
   - Username: `e2etest`
   - Email: `e2etest@example.com`
   - Password: `password123`
4. Click "Sign up" button
5. Verify redirect to home page with user logged in

**Expected Results:**
- [x] User is created successfully
- [x] User is automatically logged in after registration
- [x] Navigation shows "New Article" and user menu
- [x] Home page displays with "Your Feed" and "Global Feed" tabs

**Status:** ✅ PASSED

---

#### 1.2 로그인 (Login)

**Steps:**
1. User logs out first (tested in 1.5)
2. Navigate to `/login`
3. Fill login form:
   - Email: `e2etest@example.com`
   - Password: `password123`
4. Click "Sign in" button
5. Verify redirect to home page

**Expected Results:**
- [x] Login form accepts valid credentials
- [x] User is redirected to home page after login
- [x] Navigation shows authenticated user menu

**Status:** ✅ PASSED

---

#### 1.3 글 작성 (Create Article)

**Steps:**
1. Click "New Article" in navigation
2. Fill article form:
   - Title: `E2E Test Article`
   - Description: `Testing article creation flow`
   - Body: `This is a test article created by E2E testing with Playwright MCP.`
   - Tags: `e2e, test, playwright`
3. Click "Publish Article" button
4. Verify redirect to article detail page

**Expected Results:**
- [x] Article editor page loads correctly
- [x] All form fields accept input
- [x] Article is created with correct slug (`e2e-test-article`)
- [x] Redirect to `/article/e2e-test-article`
- [x] Article displays with title, body, tags, and author info

**Status:** ✅ PASSED

---

#### 1.4 댓글 작성 (Create Comment)

**Steps:**
1. On article detail page, find comment section
2. Type comment: `This is a test comment from E2E testing!`
3. Click "Post Comment" button
4. Verify comment appears in comments list

**Expected Results:**
- [x] Comment textbox is visible
- [x] Comment is submitted successfully
- [x] Comment appears with author info and delete button
- [x] Comment text matches input

**Status:** ✅ PASSED

---

#### 1.5 로그아웃 (Logout)

**Steps:**
1. Click user menu button in navigation
2. Click "Log out" menu item
3. Verify redirect to login page

**Expected Results:**
- [x] User menu opens with Profile, Settings, Log out options
- [x] Clicking "Log out" clears session
- [x] User is redirected to `/login`
- [x] Navigation shows "Sign in" and "Sign up" links

**Status:** ✅ PASSED

---

### Scenario 2: 좋아요 → 프로필에서 확인

#### 2.1 좋아요 (Favorite Article)

**Steps:**
1. Navigate to home page
2. Click "Global Feed" tab
3. Find article in feed
4. Click "Favorite" button (heart icon)
5. Verify favorite count increases

**Expected Results:**
- [x] Global Feed shows test article
- [x] Favorite button shows count "0" initially
- [x] After clicking, button changes to "Unfavorite"
- [x] Count changes to "1"
- [x] Tooltip shows "Unfavorite article"

**Status:** ✅ PASSED

---

#### 2.2 프로필에서 확인 (View Favorited in Profile)

**Steps:**
1. Click user menu → Profile
2. Navigate to profile page
3. Click "Favorited Articles" tab
4. Verify favorited article appears

**Expected Results:**
- [x] Profile page shows user info (avatar, username)
- [x] "My Articles" tab shows user's articles
- [x] "Favorited Articles" tab is accessible
- [x] Favorited article appears in list

**Status:** ✅ PASSED

> **Bug Fixed (2025-11-30):** Initial testing showed content not rendering due to TanStack Router
> nested route issue. Fixed by restructuring routes from flat file naming (`$username.favorites.tsx`)
> to folder-based structure (`$username/favorites.tsx`). See Scenario 2.3 for detailed investigation.

---

#### 2.3 심화 테스트: Favorited Articles 라우팅 검증 (Deep Investigation)

> **Purpose:** Detailed investigation of the Favorited Articles tab routing issue.
> This scenario documents the debugging process and root cause analysis.

**Investigation Steps:**

1. **API Verification**
   ```bash
   # Verify backend API returns favorited articles correctly
   curl -s "http://localhost:8080/api/articles?favorited=e2etest" | jq .
   ```
   - [x] API returns `articlesCount: 1` with correct article data
   - [x] Response includes title, description, tags, author info

2. **Route Configuration Analysis**
   - [x] Check `routeTree.gen.ts` for route hierarchy
   - [x] Verify route path definitions

3. **Browser State Verification**
   - [x] Navigate directly to `/profile/e2etest/favorites`
   - [x] Check which tab is `[selected]` in accessibility snapshot
   - [x] Verify which `tabpanel` is rendered

**Root Cause Identified:**

The issue was a **TanStack Router nested route problem**:

| Aspect | Before (Bug) | After (Fixed) |
|--------|--------------|---------------|
| File Structure | `$username.favorites.tsx` (dot naming) | `$username/favorites.tsx` (folder) |
| Route Relationship | Child of `$username` route | Sibling route at root level |
| Parent Route | `getParentRoute: () => ProfileUsernameRoute` | `getParentRoute: () => rootRouteImport` |
| Required for Child | `<Outlet />` in parent (missing) | Not required |

**Fix Applied:**

1. Created folder structure: `routes/profile/$username/`
2. Moved route files:
   - `$username.tsx` → `$username/index.tsx`
   - `$username.favorites.tsx` → `$username/favorites.tsx`
3. Updated import paths to use `../../../` instead of `../../`

**Verification After Fix:**

- [x] "Favorited Articles" tab shows as `[selected]`
- [x] Article "E2E Test Article" appears in the list
- [x] All article metadata (author, date, tags) displayed correctly
- [x] Unfavorite button with count "1" visible
- [x] Navigation between tabs works correctly

**Status:** ✅ PASSED (Bug Fixed)

---

### Scenario 3: 팔로우 → 피드 확인

> **Note:** This scenario requires multiple users to test properly.
> In a single-user test environment, the follow functionality can be verified
> through API testing or by creating additional test users.

#### 3.1 팔로우 (Follow User)

**Prerequisites:**
- Create second user or use existing user
- Second user must have published articles

**Steps:**
1. Navigate to another user's profile
2. Click "Follow" button
3. Verify button changes to "Unfollow"

**Expected Results:**
- [ ] Follow button appears on other users' profiles
- [ ] Following status updates after click
- [ ] Feed includes followed user's articles

**Status:** ⏭️ SKIPPED (Requires multiple users)

---

#### 3.2 피드 확인 (View Feed)

**Steps:**
1. Navigate to home page
2. Click "Your Feed" tab
3. Verify followed users' articles appear

**Expected Results:**
- [ ] Your Feed shows articles from followed users
- [ ] Empty state shows "Follow some users..." message when no followers

**Status:** ⏭️ SKIPPED (Requires multiple users with articles)

---

## 테스트 결과 요약

| Scenario | Status | Notes |
|----------|--------|-------|
| 1.1 회원가입 | ✅ PASSED | - |
| 1.2 로그인 | ✅ PASSED | - |
| 1.3 글 작성 | ✅ PASSED | - |
| 1.4 댓글 작성 | ✅ PASSED | - |
| 1.5 로그아웃 | ✅ PASSED | - |
| 2.1 좋아요 | ✅ PASSED | - |
| 2.2 프로필 Favorites | ✅ PASSED | Bug fixed (routing restructure) |
| 2.3 심화 라우팅 검증 | ✅ PASSED | Deep investigation documented |
| 3.1 팔로우 | ⏭️ SKIPPED | Requires multi-user setup |
| 3.2 피드 확인 | ⏭️ SKIPPED | Requires multi-user setup |

**Overall: 8/10 scenarios passed, 0 partial, 2 skipped**

---

## Known Issues

### ~~1. Favorited Articles Tab Content~~ (RESOLVED)
- **Issue:** The "Favorited Articles" tab on the profile page navigates correctly to `/profile/{username}/favorites` but doesn't display the favorited articles.
- **Root Cause:** TanStack Router nested route configuration - dot naming (`$username.favorites.tsx`) creates child routes that require `<Outlet />` in parent.
- **Resolution:** Restructured to folder-based routing (`$username/favorites.tsx`) making routes siblings at root level.
- **Fixed:** 2025-11-30

---

## Playwright MCP Commands Used

```typescript
// Navigation
browser_navigate({ url: "http://localhost:5173" })

// Form interactions
browser_fill_form({ fields: [...] })
browser_type({ element: "...", ref: "...", text: "..." })

// Click actions
browser_click({ element: "...", ref: "..." })

// State verification
browser_snapshot()
browser_take_screenshot({ filename: "..." })

// Wait for content
browser_wait_for({ time: 1 })
```

---

## 테스트 실행 방법

### Prerequisites
1. Backend server running on port 8080
2. Frontend dev server running on port 5173
3. Playwright MCP server configured

### Manual Testing Steps
```bash
# Terminal 1: Start backend
cd backend && go run ./cmd/server/main.go

# Terminal 2: Start frontend
cd frontend && npm run dev

# Use Playwright MCP in Claude Code to execute test scenarios
```

### Automated Testing (Future)
Consider implementing Playwright test scripts in `frontend/e2e/` directory for CI/CD integration.

---

## 향후 개선 사항

1. **Multi-user Test Setup**: Create test fixtures for follow/feed scenarios
2. **Automated Test Scripts**: Convert manual scenarios to Playwright test files
3. **CI/CD Integration**: Run E2E tests in GitHub Actions
4. **Visual Regression**: Add screenshot comparison tests
5. **Performance Metrics**: Measure page load times and Core Web Vitals
