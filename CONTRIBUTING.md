# Contributing to RealWorld Conduit

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)

## Code of Conduct

This project follows a simple code of conduct:

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 20+
- Make
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/alexlee-dev/realworld-vibe-coding.git
cd realworld-vibe-coding

# Install dependencies
make install

# Initialize the database
make db-init

# Start development servers
make dev
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run frontend tests in watch mode
make test-watch
```

### Code Quality

Before submitting a PR, ensure all checks pass:

```bash
# Linting
make lint

# Type checking
make typecheck

# All tests
make test
```

## Coding Standards

This project follows [Agentic Coding](https://lucumr.pocoo.org/2025/6/12/agentic-coding/) principles:

### General Principles

1. **Descriptive Names**: Use clear, descriptive function and variable names
2. **Explicit over Implicit**: Make behavior explicit in code
3. **Composition over Inheritance**: Prefer composing smaller functions
4. **No Magic**: Avoid hidden behavior or clever tricks

### Backend (Go)

#### Code Style

```go
// ✅ Good: Descriptive function name
func GetUserByEmail(ctx context.Context, email string) (*User, error)

// ❌ Bad: Unclear name
func GetUser(ctx context.Context, v string) (*User, error)
```

#### Error Handling

```go
// ✅ Good: Explicit error handling
user, err := repo.GetUserByID(ctx, id)
if err != nil {
    if errors.Is(err, domain.ErrNotFound) {
        return nil, domain.ErrUserNotFound
    }
    return nil, fmt.Errorf("get user: %w", err)
}

// ❌ Bad: Silent error
user, _ := repo.GetUserByID(ctx, id)
```

#### Logging

```go
// ✅ Good: Structured logging with context
slog.Info("user created",
    "user_id", user.ID,
    "username", user.Username,
)

// ❌ Bad: Unstructured logging
log.Printf("user created: %v", user)
```

#### Testing (TDD)

Write tests first:

```go
func TestCreateUser(t *testing.T) {
    // Arrange
    repo := setupTestRepo(t)
    user := &domain.User{
        Username: "testuser",
        Email:    "test@example.com",
    }

    // Act
    err := repo.CreateUser(context.Background(), user)

    // Assert
    require.NoError(t, err)
    assert.NotZero(t, user.ID)
}
```

### Frontend (TypeScript/React)

#### Component Structure

```tsx
// ✅ Good: Clear component with typed props
interface ArticleCardProps {
  article: Article;
  onFavorite: (slug: string) => void;
}

export function ArticleCard({ article, onFavorite }: ArticleCardProps) {
  return (
    <Card>
      <Title>{article.title}</Title>
      <Text>{article.description}</Text>
    </Card>
  );
}
```

#### State Management

- Use **TanStack Query** for server state
- Use **Zustand** for client state (auth, UI)
- Use **Mantine Form** for form state

```tsx
// Server state with TanStack Query
const { data: articles } = useArticles({ tag, author });

// Client state with Zustand
const { token, setToken } = useAuthStore();

// Form state with Mantine Form
const form = useForm({
  initialValues: { email: '', password: '' },
  validate: zodResolver(loginSchema),
});
```

#### Testing

```tsx
import { render, screen } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';

const renderWithProviders = (component: React.ReactNode) => {
  return render(
    <MantineProvider>
      {component}
    </MantineProvider>
  );
};

test('renders article title', () => {
  renderWithProviders(<ArticleCard article={mockArticle} />);
  expect(screen.getByText('Test Article')).toBeInTheDocument();
});
```

## Pull Request Process

### Branch Naming

Use descriptive branch names:

```
feature/add-user-avatar
fix/login-validation-error
docs/update-api-documentation
refactor/simplify-auth-middleware
```

### Commit Messages

Follow conventional commits:

```
feat: add user avatar upload
fix: resolve login validation error
docs: update API documentation
refactor: simplify auth middleware
test: add user service tests
chore: update dependencies
```

### PR Checklist

Before submitting a PR, ensure:

- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Type checking passes (`make typecheck`)
- [ ] New features have tests
- [ ] Documentation is updated if needed
- [ ] PR description explains the changes

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How were these changes tested?

## Checklist
- [ ] Tests pass
- [ ] Lint passes
- [ ] Documentation updated
```

### Review Process

1. Create a PR against `main` branch
2. Ensure CI checks pass
3. Request review from maintainers
4. Address feedback
5. Squash and merge after approval

## Issue Guidelines

### Bug Reports

Include:
- Clear title describing the bug
- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment details (OS, browser, versions)

### Feature Requests

Include:
- Clear description of the feature
- Use case / motivation
- Proposed implementation (optional)
- Alternatives considered

### Issue Labels

- `bug`: Something isn't working
- `enhancement`: New feature or improvement
- `documentation`: Documentation updates
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention needed

## Questions?

If you have questions:

1. Check existing issues and documentation
2. Open a new issue with the `question` label
3. Be specific about what you're trying to accomplish

Thank you for contributing!
