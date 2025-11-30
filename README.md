# RealWorld Conduit

[![Backend CI](https://github.com/alexlee-dev/realworld-vibe-coding/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/alexlee-dev/realworld-vibe-coding/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/alexlee-dev/realworld-vibe-coding/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/alexlee-dev/realworld-vibe-coding/actions/workflows/frontend-ci.yml)

> A full-stack social blogging platform built with Go and React, following the [RealWorld](https://realworld-docs.netlify.app/) specification and [Agentic Coding](https://lucumr.pocoo.org/2025/6/12/agentic-coding/) principles.

![RealWorld](https://img.shields.io/badge/RealWorld-Fullstack-blueviolet)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18.3+-61DAFB?logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5.3+-3178C6?logo=typescript)

## Overview

This project is a "Medium.com clone" implementing the RealWorld spec. It demonstrates a modern full-stack architecture optimized for AI-assisted development (Agentic Coding).

### Key Features

- User authentication with JWT
- Article CRUD with Markdown support
- Comments system
- User profiles with follow functionality
- Article favorites and tags
- Global and personal feed

## Tech Stack

### Backend
- **Go 1.21+** - Standard library `net/http` router
- **SQLite** (development) / **PostgreSQL 16** (production)
- **Pure SQL** - No ORM, explicit queries
- **golang-jwt v5** - JWT authentication
- **golang-migrate v4** - Database migrations

### Frontend
- **React 18.3+** with **Vite 5+**
- **TypeScript 5.3+** - Type safety
- **TanStack Router** - File-based routing
- **TanStack Query v5** - Server state management
- **Zustand** - Client state (auth)
- **Mantine v7** - UI component library
- **Vitest** - Unit testing

### Infrastructure
- **AWS CDK** - Infrastructure as Code
- **ECS Fargate** - Container orchestration
- **RDS PostgreSQL** - Managed database
- **GitHub Pages** - Frontend hosting
- **GitHub Actions** - CI/CD pipelines

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 20+
- Make

### Installation

```bash
# Clone the repository
git clone https://github.com/alexlee-dev/realworld-vibe-coding.git
cd realworld-vibe-coding

# Install all dependencies
make install

# Initialize the database
make db-init
```

### Running Locally

```bash
# Start both backend and frontend
make dev

# Or run separately
make dev-backend    # Backend on http://localhost:8080
make dev-frontend   # Frontend on http://localhost:5173
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Frontend watch mode
make test-watch
```

### Code Quality

```bash
# Run linters
make lint

# Type checking
make typecheck

# All checks
make lint && make typecheck && make test
```

## Project Structure

```
realworld-vibe-coding/
├── backend/
│   ├── cmd/server/           # Application entry point
│   ├── internal/
│   │   ├── api/              # HTTP handlers, middleware
│   │   ├── config/           # Configuration management
│   │   ├── domain/           # Domain models, errors
│   │   ├── repository/       # Data access layer
│   │   └── service/          # Business logic
│   ├── db/
│   │   ├── migrations/       # SQL migrations
│   │   └── queries/          # SQL query files
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/       # Reusable UI components
│   │   ├── features/         # Feature modules (api, hooks, types)
│   │   ├── routes/           # TanStack Router pages
│   │   ├── stores/           # Zustand stores
│   │   └── lib/              # Utilities
│   └── vite.config.ts
├── infra/                    # AWS CDK stacks
├── docs/                     # Documentation
└── docker-compose.yml
```

## API Documentation

See [docs/api.md](docs/api.md) for complete API reference.

### Quick API Overview

| Endpoint | Method | Description | Auth |
|----------|--------|-------------|------|
| `/api/users` | POST | Register | - |
| `/api/users/login` | POST | Login | - |
| `/api/user` | GET/PUT | Current user | Required |
| `/api/profiles/:username` | GET | Get profile | Optional |
| `/api/profiles/:username/follow` | POST/DELETE | Follow/Unfollow | Required |
| `/api/articles` | GET/POST | List/Create articles | Optional/Required |
| `/api/articles/:slug` | GET/PUT/DELETE | Article CRUD | Optional/Required |
| `/api/articles/:slug/favorite` | POST/DELETE | Favorite | Required |
| `/api/articles/:slug/comments` | GET/POST | Comments | Optional/Required |
| `/api/tags` | GET | List tags | - |

## Environment Variables

```bash
# Backend
DATABASE_URL=sqlite://./data/conduit.db
JWT_SECRET=your-secret-key
JWT_EXPIRY=72h
SERVER_PORT=8080
SERVER_ENV=development

# Frontend
VITE_API_URL=http://localhost:8080/api
```

## Deployment

See [docs/deployment.md](docs/deployment.md) for detailed deployment instructions.

### Quick Deploy (AWS)

```bash
# Prerequisites: AWS CLI configured, Docker running

# Bootstrap CDK (first time only)
cd infra && npx cdk bootstrap

# Deploy infrastructure
make deploy

# Deploy frontend to GitHub Pages
make deploy-frontend
```

## Development Principles

This project follows [Agentic Coding](https://lucumr.pocoo.org/2025/6/12/agentic-coding/) principles:

1. **Descriptive function names** over short/ambiguous ones
2. **Composition** over inheritance
3. **Plain SQL** over ORM magic
4. **Explicit permission checks** in code
5. **Structured logging** with `slog`
6. **Test-Driven Development** for backend

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Security

See [docs/SECURITY.md](docs/SECURITY.md) for security policies.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [RealWorld](https://github.com/gothinkster/realworld) - The specification
- [Armin Ronacher](https://lucumr.pocoo.org/) - Agentic Coding principles
- [Mantine](https://mantine.dev/) - UI components
- [TanStack](https://tanstack.com/) - Router & Query
