# AGENTS.md - Developer Guide for TaskManager

This document provides guidelines for agentic coding agents working on the TaskManager codebase.

---

## Project Overview

TaskManager is a production-ready monorepo with:
- **Backend**: Go (Echo framework) with PostgreSQL and Redis
- **Frontend packages**: TypeScript packages (zod, openapi, emails)
- **Package Manager**: Bun
- **Build System**: Turborepo

---

## Build, Lint, and Test Commands

### Root Commands (Monorepo)

```bash
# Install dependencies
bun install

# Build all packages
bun run build

# Run all packages in dev mode
bun run dev

# Type check all packages
bun run typecheck

# Lint all packages
bun run lint

# Fix lint issues
bun run lint:fix

# Check formatting
bun run format:check

# Fix formatting
bun run format:fix

# Clean all build artifacts
bun run clean
```

### Backend Commands (Go)

```bash
# Run the server
cd backend
go run ./cmd/taskmanager

# Run with Taskfile
task run

# Run linting
golangci-lint run

# Fix lint issues
golangci-lint run --fix

# Run specific linter
golangci-lint run --disable-all -E errcheck

# Format and tidy Go code
go fmt ./...
go mod tidy
go mod verify

# Using Taskfile
task tidy

# Run migrations
task migrations:new name=migration_name
task migrations:up

# Run cron jobs
go run ./cmd/cron list
go run ./cmd/cron due-date-reminders
go run ./cmd/cron weekly-reports

# Run a single Go test
go test -v -run TestFunctionName ./internal/repository/
go test -v -run TestFunctionName ./internal/service/
go test -v ./internal/repository/todo_test.go
```

### TypeScript Packages

```bash
# Build a specific package
cd packages/zod && bun run build
cd packages/openapi && bun run build

# Generate OpenAPI spec
cd packages/openapi && bun run gen

# Dev mode
cd packages/zod && bun run dev

# Run email dev server
cd packages/emails && bun run dev

# Export emails to backend
cd packages/emails && bun run export
```

---

## Code Style Guidelines

### Go (Backend)

#### Imports
- Use standard Go import grouping:
  1. Standard library
  2. Third-party packages
  3. Internal packages
- Use blank line between groups
- Example:
```go
import (
    "context"
    "encoding/json"
    "time"

    "github.com/apk471/go-taskmanager/internal/errs"
    "github.com/apk471/go-taskmanager/internal/model"
    "github.com/google/uuid"
    "github.com/labstack/echo/v4"
)
```

#### Naming Conventions
- **Variables/Functions**: camelCase
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for enums
- **Types/Interfaces**: PascalCase
- **Packages**: lowercase, short, descriptive (e.g., `errs`, `repo`)
- **Files**: lowercase with underscores (e.g., `todo_handler.go`)
- **Tests**: `*_test.go` suffix

#### Error Handling
- Always handle errors explicitly; avoid bare `err` checks
- Use custom error types from `internal/errs/`
- Log errors with appropriate level (Error for failures, Warn for recoverable)
- Return errors to caller; don't wrap unnecessarily
- Example:
```go
if err != nil {
    logger.Error().Err(err).Msg("failed to fetch todo")
    return nil, err
}
```

#### Logging
- Use zerolog for structured logging
- Include context with `.Str()`, `.Int()`, etc.
- Use appropriate log levels: Debug, Info, Warn, Error
- Example:
```go
logger.Info().
    Str("event", "todo_created").
    Str("todo_id", todoItem.ID.String()).
    Msg("Todo created successfully")
```

#### Types
- Use concrete types over interfaces where possible
- Use pointers for nullable fields (`*string`, `*time.Time`)
- Prefer `uuid.UUID` over string for IDs
- Use custom type aliases for enums:
```go
type Status string
type Priority string

const (
    StatusDraft     Status = "draft"
    StatusCompleted Status = "completed"
)
```

#### Context
- Pass context as first parameter: `func DoSomething(ctx context.Context, ...)`
- Use `ctx.Request().Context()` in Echo handlers

#### Middleware Order (in `router.go`)
1. Global Error Handler
2. Rate Limiter
3. CORS
4. Secure Headers
5. Request ID
6. New Relic
7. Tracing
8. Context Enhancer
9. Request Logger
10. Recover

---

### TypeScript (Packages)

#### Module System
- Use ESM with `NodeNext` module resolution
- Include `.js` extension in imports: `import { Foo } from "./foo.js"`
- Use `verbatimModuleSyntax: true`

#### Configuration
```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "module": "NodeNext",
    "target": "es2022"
  }
}
```

#### Naming Conventions
- **Files**: kebab-case or PascalCase for components
- **Types/Interfaces**: PascalCase
- **Functions/Variables**: camelCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE

#### Imports
- Use path aliases: `@/*` maps to `./src/*`
- Group imports: external, then internal
- Example:
```typescript
import z from "zod";
import { ZCategory } from "../category/index.js";
import { ZTodo } from "./todo/index.js";
```

#### Zod Schemas
- Use `@anatine/zod-openapi` for OpenAPI metadata
- Export schema types matching backend models
- Use `.extend()` for derived types
- Example:
```typescript
export const ZTodo = z.object({
  id: z.string().uuid(),
  title: z.string(),
  status: ZTodoStatus,
});

export const ZPopulatedTodo = ZTodo.extend({
  category: ZTodoCategory.nullable(),
  children: z.array(ZTodo),
});
```

---

## Architecture Patterns

### Backend Layer Order
1. **Handler** (`internal/handler/`) - HTTP request handling
2. **Service** (`internal/service/`) - Business logic
3. **Repository** (`internal/repository/`) - Database operations
4. **Model** (`internal/model/`) - Data structures

### Adding a New Feature
1. Create model in `internal/model/`
2. Add Zod schema in `packages/zod/`
3. Add ts-rest contract in `packages/openapi/`
4. Create repository in `internal/repository/`
5. Create service in `internal/service/`
6. Create handler in `internal/handler/`
7. Register route in `internal/router/v1/`
8. Generate OpenAPI spec

### Database Migrations
- Use `tern` for migrations
- Migration files in `internal/database/migrations/`
- Create with: `task migrations:new name=feature_name`

---

## Testing

### Go Tests
- Tests use standard Go testing package
- No test framework configured (no testify)
- Run specific test: `go test -v -run TestName ./path/`
- Run all tests: `go test ./...`

### Test Naming
- Test functions: `func TestName(t *testing.T)`
- Table-driven tests use `t.Run()` for subtests

---

## Configuration

### Environment Variables
- All variables use `TASKMANAGER_` prefix
- See `backend/.env` for all options
- Required: Database, Redis, Clerk auth

### Development Setup
1. Copy `backend/.env.example` to `backend/.env`
2. Set up PostgreSQL and Redis
3. Run `go mod tidy`
4. Run migrations: `task migrations:up`
5. Start server: `task run`

---

## Dependencies

### Backend Key Dependencies
- Echo v4 - HTTP framework
- pgx v5 - PostgreSQL driver
- go-redis v9 - Redis client
- Asynq - Background jobs
- zerolog - Logging
- Clerk SDK - Authentication
- New Relic - Observability

### Frontend Key Dependencies
- zod - Schema validation
- @anatine/zod-openapi - OpenAPI integration
- @ts-rest/core - Type-safe API contracts
- react-email - Email templates

---

## Notes

- No Cursor or Copilot rules exist in this repository
- This is a monorepo; changes to TypeScript packages require rebuilding dependencies
- OpenAPI spec is generated from ts-rest contracts
- All backend code uses Go 1.25+
- All Node code requires Node 22+ and Bun 1.2+
