# Frontend Deployment Troubleshooting Guide

이 문서는 RealWorld Conduit 프론트엔드 개발 및 배포 과정에서 발생했던 오류와 해결책을 정리한 것입니다.

## 목차

1. [TypeScript verbatimModuleSyntax 오류](#1-typescript-verbatimmodulesyntax-오류)
2. [Vitest matchMedia 오류](#2-vitest-matchmedia-오류)
3. [Vitest 설정 위치 오류](#3-vitest-설정-위치-오류)
4. [ESLint react-refresh 오류](#4-eslint-react-refresh-오류)
5. [GitHub Actions 빌드 오류 (global 미정의)](#5-github-actions-빌드-오류-global-미정의)
6. [GitHub Pages SPA 라우팅](#6-github-pages-spa-라우팅)
7. [CORS 설정 문제](#7-cors-설정-문제)

---

## 1. TypeScript verbatimModuleSyntax 오류

### 오류 메시지

```
src/lib/theme.ts(1,23): error TS1484: 'MantineColorsTuple' is a type and must be imported using a type-only import when 'verbatimModuleSyntax' is enabled.
src/test/utils.tsx(1,10): error TS1484: 'ReactNode' is a type and must be imported using a type-only import when 'verbatimModuleSyntax' is enabled.
src/test/utils.tsx(2,18): error TS1484: 'RenderOptions' is a type and must be imported using a type-only import when 'verbatimModuleSyntax' is enabled.
```

### 원인

TypeScript 5.0+에서 `verbatimModuleSyntax`가 활성화되어 있을 때, 타입 전용 import는 `import type` 구문을 사용해야 합니다.

### 해결 방법

```typescript
// ❌ Before
import { MantineColorsTuple } from '@mantine/core';

// ✅ After
import { createTheme, type MantineColorsTuple } from '@mantine/core';
// 또는
import type { MantineColorsTuple } from '@mantine/core';
```

```typescript
// ❌ Before
import { ReactNode } from 'react';
import { RenderOptions } from '@testing-library/react';

// ✅ After
import type { ReactNode } from 'react';
import type { RenderOptions } from '@testing-library/react';
```

### 수정된 파일

- `frontend/src/lib/theme.ts`: `type MantineColorsTuple` 사용
- `frontend/src/test/utils.tsx`: `type ReactNode`, `type RenderOptions` 사용

---

## 2. Vitest matchMedia 오류

### 오류 메시지

```
TypeError: window.matchMedia is not a function
```

### 원인

Mantine UI 컴포넌트가 반응형 디자인을 위해 `window.matchMedia`를 사용하는데, jsdom 환경에서는 이 API가 구현되어 있지 않습니다.

### 해결 방법

`frontend/src/test/setup.ts`에 mock 추가:

```typescript
import { vi } from 'vitest';

// Mock matchMedia for Mantine
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock ResizeObserver (Mantine에서도 사용)
window.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));
```

### 관련 설정

`vitest.config.ts`에서 setup 파일 지정:

```typescript
export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,
  },
});
```

---

## 3. Vitest 설정 위치 오류

### 오류 메시지

```
vite.config.ts(8,3): error TS2769: No overload matches this call.
  The last overload gave the following error.
    Object literal may only specify known properties, and 'test' does not exist in type 'UserConfigExport'.
```

### 원인

Vitest 설정을 `vite.config.ts`에 직접 넣으면 타입 오류가 발생합니다. Vite의 `defineConfig`는 `test` 속성을 인식하지 못합니다.

### 해결 방법

**Option 1**: 별도의 `vitest.config.ts` 파일 생성 (권장)

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,
  },
});
```

**Option 2**: Vitest의 `defineConfig` 사용

```typescript
// vite.config.ts
import { defineConfig } from 'vitest/config'; // 'vite' 대신 'vitest/config' 사용
// ...
```

### 프로젝트 구조

```
frontend/
├── vite.config.ts      # Vite 빌드 설정
├── vitest.config.ts    # Vitest 테스트 설정 (별도 파일)
└── src/test/setup.ts   # 테스트 환경 setup
```

---

## 4. ESLint react-refresh 오류

### 오류 메시지

```
error  Fast refresh only works when a file only exports components. Use a named export for both components and helpers  react-refresh/only-export-components
```

### 원인

`react-refresh` ESLint 플러그인이 테스트 파일에서도 적용되어, 테스트 유틸리티 함수가 컴포넌트가 아니라는 경고를 표시합니다.

### 해결 방법

**Option 1**: 테스트 파일에서 ESLint 규칙 비활성화

```typescript
// src/test/utils.tsx
/* eslint-disable react-refresh/only-export-components */
```

**Option 2**: ESLint 설정에서 테스트 파일 제외

```javascript
// eslint.config.js
export default [
  // ... 기존 설정
  {
    ignores: ['**/test/**', '**/*.test.*', '**/*.spec.*'],
  },
];
```

**Option 3**: 특정 파일 패턴에 규칙 비활성화

```javascript
// eslint.config.js
{
  files: ['**/test/**/*.{ts,tsx}', '**/*.test.{ts,tsx}'],
  rules: {
    'react-refresh/only-export-components': 'off',
  },
}
```

---

## 5. GitHub Actions 빌드 오류 (global 미정의)

### 오류 메시지

```
src/test/setup.ts(21,1): error TS2304: Cannot find name 'global'.
```

### 원인

Node.js의 `global` 객체를 브라우저 환경(jsdom)의 테스트 설정에서 사용하려고 할 때 발생합니다.

### 해결 방법

`global` 대신 `window` 또는 `globalThis` 사용:

```typescript
// ❌ Before
global.ResizeObserver = vi.fn();

// ✅ After
window.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));
```

또는 `@types/node`가 설치되어 있는지 확인:

```bash
npm install -D @types/node
```

그리고 `tsconfig.json`에서 node 타입 포함:

```json
{
  "compilerOptions": {
    "types": ["node", "vitest/globals"]
  }
}
```

---

## 6. GitHub Pages SPA 라우팅

### 문제

GitHub Pages에서 SPA를 호스팅할 때, 직접 URL 접근 시 404 오류 발생.
예: `https://username.github.io/repo/article/my-article` → 404

### 원인

GitHub Pages는 정적 파일 서버이므로 `/article/my-article` 경로에 해당하는 파일이 없으면 404를 반환합니다.

### 해결 방법

**Option 1**: 404.html로 리다이렉트 (권장)

```html
<!-- public/404.html -->
<!DOCTYPE html>
<html>
<head>
  <script>
    // URL을 쿼리 파라미터로 변환하여 index.html로 리다이렉트
    const path = window.location.pathname + window.location.search + window.location.hash;
    window.location.replace('/' + window.location.pathname.split('/')[1] + '/?p=' + encodeURIComponent(path));
  </script>
</head>
</html>
```

```typescript
// main.tsx에서 복원
const params = new URLSearchParams(window.location.search);
const redirectPath = params.get('p');
if (redirectPath) {
  window.history.replaceState(null, '', decodeURIComponent(redirectPath));
}
```

**Option 2**: HashRouter 사용

```typescript
// TanStack Router는 기본적으로 history 라우터 사용
// hash 라우터가 필요하면 createHashHistory 사용
import { createHashHistory } from '@tanstack/react-router';

const hashHistory = createHashHistory();
```

### Vite 설정

GitHub Pages 배포 시 base path 설정 필요:

```typescript
// vite.config.ts
export default defineConfig({
  base: '/repository-name/',  // GitHub Pages용
  // ...
});
```

---

## 7. CORS 설정 문제

### 오류 메시지 (브라우저 콘솔)

```
Access to fetch at 'https://api.example.com/api/articles' from origin 'https://username.github.io' has been blocked by CORS policy
```

### 원인

프론트엔드(GitHub Pages)와 백엔드(AWS ECS)가 다른 도메인에 있을 때 CORS 설정이 필요합니다.

### 해결 방법

**백엔드에서 CORS 헤더 추가** (Go):

```go
// backend/internal/api/middleware/cors.go
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
    origins := strings.Split(allowedOrigins, ",")

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            for _, allowed := range origins {
                if strings.TrimSpace(allowed) == origin {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    w.Header().Set("Access-Control-Allow-Credentials", "true")
                    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                    w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
                    break
                }
            }

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

**환경 변수 설정**:

```bash
# ECS Task Definition 또는 환경 설정
CORS_ALLOWED_ORIGINS=https://alexlee0213.github.io,http://localhost:5173
```

### 프론트엔드 API 클라이언트 설정

```typescript
// frontend/src/lib/api.ts
import ky from 'ky';

export const api = ky.create({
  prefixUrl: import.meta.env.VITE_API_URL || '/api',
  credentials: 'include',  // 쿠키 포함 시 필요
  headers: {
    'Content-Type': 'application/json',
  },
});
```

---

## 빌드 및 배포 체크리스트

### 로컬 개발

```bash
# 의존성 설치
cd frontend && npm install

# 타입 체크
npm run typecheck

# 린트
npm run lint

# 테스트
npm run test:run

# 빌드
npm run build
```

### GitHub Actions CI

CI 파이프라인에서 확인하는 항목:
1. `npm run lint` - ESLint 검사
2. `npm run typecheck` - TypeScript 타입 검사
3. `npm run test -- --run` - Vitest 테스트
4. `npm run build` - 프로덕션 빌드

### GitHub Pages 배포

```bash
# 수동 배포 (gh-pages 패키지 사용)
npm run predeploy && npm run deploy

# 또는 GitHub Actions 자동 배포
# .github/workflows/deploy-frontend.yml 참조
```

---

## 관련 파일 구조

```
frontend/
├── src/
│   ├── test/
│   │   ├── setup.ts        # Vitest 환경 설정 (matchMedia mock 등)
│   │   └── utils.tsx       # 테스트 유틸리티
│   └── lib/
│       ├── api.ts          # API 클라이언트
│       └── theme.ts        # Mantine 테마
├── vite.config.ts          # Vite 빌드 설정
├── vitest.config.ts        # Vitest 테스트 설정
├── tsconfig.json           # TypeScript 설정
├── eslint.config.js        # ESLint 설정
└── package.json
```

---

## 참고 자료

- [Mantine Testing Guide](https://mantine.dev/guides/jest/)
- [Vitest Configuration](https://vitest.dev/config/)
- [GitHub Pages SPA Support](https://github.com/rafgraph/spa-github-pages)
- [TypeScript verbatimModuleSyntax](https://www.typescriptlang.org/tsconfig#verbatimModuleSyntax)
- [TanStack Router Documentation](https://tanstack.com/router/latest)
