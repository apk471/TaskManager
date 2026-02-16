# TaskManager

A production-ready task management application with a Go backend API, PostgreSQL, Redis, background jobs (Asynq), cron jobs, observability (New Relic), Clerk authentication, and shared TypeScript packages for API contracts and Zod schemas.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Backend](#backend)
  - [Entry Point & Startup](#entry-point--startup)
  - [Configuration](#configuration)
  - [Server](#server)
  - [Database](#database)
  - [Router & Middleware](#router--middleware)
  - [Handlers](#handlers)
  - [Domain Models](#domain-models)
  - [Services & Repositories](#services--repositories)
  - [Errors](#errors)
  - [Logging & Observability](#logging--observability)
  - [Background Jobs](#background-jobs)
  - [Cron Jobs](#cron-jobs)
  - [Email](#email)
  - [Validation](#validation)
- [Packages (TypeScript)](#packages-typescript)
- [API Endpoints](#api-endpoints)
- [Tooling](#tooling)
- [Environment Variables](#environment-variables)
- [Running the Project](#running-the-project)
- [Extending the TaskManager](#extending-the-taskmanager)

---

## Overview

TaskManager is a full-featured task management system designed for production use. It provides a REST API for managing todos, categories, and comments with features like hierarchical todos, due date reminders, overdue notifications, weekly productivity reports, and automatic archiving.

---

## Features

- **Todo Management**: Create, update, delete, and organize todos with priorities, statuses, due dates, and metadata
- **Hierarchical Todos**: Support for parent-child todo relationships (subtasks)
- **Categories**: Organize todos into color-coded categories
- **Comments**: Add comments to todos for collaboration
- **Attachments**: Upload and manage file attachments on todos
- **Pagination & Filtering**: Advanced query support with sorting, filtering, and search
- **Background Jobs**: Async email processing via Asynq (Redis-backed)
- **Cron Jobs**: Scheduled tasks for reminders, notifications, reports, and auto-archiving
- **Authentication**: Clerk-based JWT authentication with user roles and permissions
- **Observability**: Full New Relic APM integration with distributed tracing and log forwarding
- **API Documentation**: OpenAPI 3.0 spec with Scalar UI

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| **API Framework** | [Echo v4](https://echo.labstack.com/) |
| **Database** | PostgreSQL via [pgx v5](https://github.com/jackc/pgx), [tern](https://github.com/jackc/tern) migrations |
| **Cache/Queue** | Redis ([go-redis](https://github.com/redis/go-redis)), [Asynq](https://github.com/hibiken/asynq) for background jobs |
| **Authentication** | [Clerk](https://clerk.com/) via `clerk-sdk-go` |
| **Configuration** | [Koanf](https://github.com/knadh/koanf) with env vars, [go-playground/validator](https://github.com/go-playground/validator) |
| **Logging** | [zerolog](https://github.com/rs/zerolog) with New Relic log forwarding |
| **Observability** | New Relic (APM, distributed tracing, nrpgx5, nrecho, nrredis) |
| **API Docs** | OpenAPI 3 via [ts-rest](https://ts-rest.com/), served with Scalar |
| **CLI** | [Cobra](https://github.com/spf13/cobra) for cron job runner |
| **Monorepo** | [Turborepo](https://turbo.build/), Bun package manager |

---

## Project Structure

```
TaskManagement/
├── backend/                    # Go API server
│   ├── cmd/go-taskmanager/     # main entry
│   ├── internal/
│   │   ├── config/             # config structs, load, observability
│   │   ├── database/           # pgx pool, migrations (embed)
│   │   ├── errs/               # HTTP error types and constructors
│   │   ├── handler/            # health, openapi, base (typed Handle/HandleNoContent/HandleFile)
│   │   ├── lib/
│   │   │   ├── email/          # Resend client, templates, welcome email
│   │   │   ├── jobs/           # Asynq job service, welcome email task
│   │   │   └── utils/          # small helpers (e.g. PrintJSON)
│   │   ├── logger/             # zerolog + New Relic LoggerService, pgx logger
│   │   ├── middleware/         # CORS, secure, request ID, tracing, context, auth, rate limit, recover, global error
│   │   ├── repository/         # repository layer (currently empty struct)
│   │   ├── router/             # Echo router, system routes registration
│   │   ├── server/             # Server struct (config, DB, Redis, Job, HTTP server)
│   │   ├── service/            # Auth (Clerk), Job service ref
│   │   ├── sqlerr/             # PG error → HTTP error mapping
│   │   └── validation/         # BindAndValidate, Validatable, tag→message mapping
│   ├── static/                 # openapi.html, openapi.json (from packages/openapi gen)
│   ├── templates/emails/       # HTML email templates (e.g. welcome.html)
│   ├── Taskfile.yml            # run, migrations:new, migrations:up, tidy
│   ├── .golangci.yml           # linter config
│   ├── go.mod
│   └── go.sum
├── packages/
│   ├── openapi/                # ts-rest contracts, OpenAPI 3 generation, writes openapi.json
│   ├── zod/                    # shared Zod schemas (e.g. health response)
│   └── emails/                 # (optional) React email templates
├── package.json                # workspace root, turbo scripts
├── turbo.json
└── README.md
```

---

## Backend

### Entry Point & Startup

- **`cmd/go-taskmanager/main.go`**
  - Loads config via `config.LoadConfig()` (env-only, `TASKMANAGER_` prefix).
  - Creates `LoggerService` (New Relic optional) and zerolog logger.
  - Runs DB migrations when `env != "local"` via `database.Migrate(...)`.
  - Builds `server.Server` (DB, Redis, Asynq job service), repositories, services, handlers, router.
  - Sets up HTTP server on `server.Port`, starts it and graceful shutdown on interrupt (30s timeout).
  - Shuts down HTTP server, DB pool, and job server.

### Configuration

- **`internal/config/config.go`**

  - **Config** struct: Primary (env), Server (port, timeouts, CORS origins), Database (host, port, user, password, name, ssl_mode, pool settings), Auth (secret for Clerk), Redis (address), Integration (e.g. Resend API key), Observability (optional).
  - Load: Koanf with `env.Provider("TASKMANAGER_", ".", lowerAndTrimPrefix)` so env vars like `TASKMANAGER_SERVER_PORT` map to `server.port`.
  - Validation with `go-playground/validator`; on failure the process exits.
  - Observability defaults: `DefaultObservabilityConfig()` and override with `observability.service_name`, `observability.environment` from primary env.

- **`internal/config/observability.go`**
  - **ObservabilityConfig:** service_name, environment, logging (level, format, slow_query_threshold), new_relic (license_key, app_log_forwarding_enabled, distributed_tracing_enabled, debug_logging), health_checks (enabled, interval, timeout, checks list).
  - `Validate()`: service_name required, log level in [debug, info, warn, error], slow_query_threshold >= 0.
  - `GetLogLevel()`: uses environment default (e.g. debug for development) when level empty.
  - `IsProduction()`: true when environment == "production".

### Server

- **`internal/server/server.go`**
  - **Server** holds: Config, Logger, LoggerService, DB (*database.Database), Redis (go-redis Client), Job (*jobs.JobService), and the HTTP server.
  - **New:** Creates DB (with optional New Relic nrpgx5 tracer, local pgx tracelog in local env), Redis client (with optional nrredis hook), Job service (Asynq client + server), starts the job server (registers task handlers).
  - **SetupHTTPServer(handler):** Sets `http.Server` (Addr from config, read/write/idle timeouts).
  - **Start:** Calls `ListenAndServe()`.
  - **Shutdown:** Shuts down HTTP server, closes DB pool, stops job server.

### Database

- **`internal/database/database.go`**

  - **Database** wraps `*pgxpool.Pool` and a logger.
  - **New:** Builds DSN (password URL-encoded), parses pool config; if LoggerService has New Relic app, sets `nrpgx5.NewTracer()`; in local env adds pgx-zerolog tracelog (or multi-tracer with both). Pool created with `pgxpool.NewWithConfig`, then ping with 10s timeout.
  - **Close:** Logs and closes pool.

- **`internal/database/migrator.go`**

  - Uses embedded `migrations/*.sql` and [tern](https://github.com/jackc/tern) with table `schema_version`.
  - **Migrate:** Connects with same DSN, creates tern migrator, loads migrations from embed, runs migrate, logs version.

- **`internal/database/migrations/001_setup.sql`**
  - Placeholder migration (empty up/down). New migrations: `task migrations:new name=something`.

### Router & Middleware

- **`internal/router/router.go`**

  - **NewRouter:** Creates Echo instance, sets **GlobalErrorHandler** from middlewares, then applies in order:
    - Rate limiter (20 req/s, memory store), DenyHandler returns 429 and records rate limit hit in New Relic.
    - CORS (origins from config), Secure(), RequestID (X-Request-ID, uuid if missing), NewRelic (nrecho), EnhanceTracing (request id, user id, status code, NoticeError), ContextEnhancer (request-scoped logger with request_id, method, path, ip, trace context, user_id, user_role), RequestLogger, Recover.
  - Registers system routes via `registerSystemRoutes`; `/api/v1` group exists for future versioned routes.

- **`internal/router/system.go`**

  - **GET /status** → HealthHandler.CheckHealth
  - **/static** → static files (e.g. openapi.json)
  - **GET /docs** → OpenAPIHandler.ServeOpenAPIUI (serves static/openapi.html, which loads /static/openapi.json and Scalar)

- **Middleware details**
  - **global (internal/middleware/global.go):** CORS, Secure, RequestLogger (status, latency, URI, etc., uses context logger and request_id/user_id), Recover, GlobalErrorHandler (sqlerr handling, then HTTP/echo error → JSON response, logging).
  - **auth (auth.go):** Clerk `WithHeaderAuthorization`; on success sets `user_id`, `user_role`, `permissions` in context; on failure returns 401 JSON.
  - **context (context.go):** Puts request-scoped logger (with request_id, method, path, ip, trace id/span id if New Relic, user_id/user_role) in context; `GetLogger(c)`, `GetUserID(c)`.
  - **request_id (request_id.go):** Reads or generates X-Request-ID, sets in context and response header.
  - **tracing (tracing.go):** Wraps nrecho middleware; EnhanceTracing adds http.real_ip, http.user_agent, request.id, user.id, http.status_code, and NoticeError on handler error.
  - **rate_limit (rate_limit.go):** RecordRateLimitHit(endpoint) for New Relic custom event when rate limit is hit.

### Handlers

- **`internal/handler/handlers.go`**

  - **Handlers** contains Health and OpenAPI. **NewHandlers** builds them from server and services.

- **`internal/handler/base.go`**

  - **Handler** is a base with server reference.
  - **HandlerFunc[Req, Res], HandlerFuncNoContent[Req]** for typed handlers.
  - **ResponseHandler** interface: Handle(c, result), GetOperation(), AddAttributes(txn, result). Implementations: **JSONResponseHandler**, **NoContentResponseHandler**, **FileResponseHandler** (filename, content-type, blob).
  - **handleRequest:** Binds and validates payload with `validation.BindAndValidate`, runs handler, records validation/handler duration and status on New Relic transaction, uses context logger; on error uses `nrpkgerrors.Wrap` and returns err; on success calls responseHandler.Handle(c, result).
  - **Handle**, **HandleNoContent**, **HandleFile** wrap handler funcs with handleRequest and the appropriate response handler.

- **`internal/handler/health.go`**

  - **CheckHealth:** Returns JSON with status (healthy/unhealthy), timestamp, environment, and **checks** (database ping, redis ping when Redis not nil). On DB/Redis failure sets check to unhealthy and records **HealthCheckError** custom event in New Relic. Returns 503 when unhealthy.

- **`internal/handler/openapi.go`**
  - **ServeOpenAPIUI:** Serves `static/openapi.html` as HTML (Cache-Control: no-cache). The HTML page loads Scalar with `/static/openapi.json`.

- **`internal/handler/todo.go`**
  - **TodoHandler:** CRUD operations for todos (CreateTodo, GetTodos, GetTodoByID, UpdateTodo, DeleteTodo).

- **`internal/handler/category.go`**
  - **CategoryHandler:** CRUD operations for categories (CreateCategory, GetCategories, UpdateCategory, DeleteCategory).

- **`internal/handler/comment.go`**
  - **CommentHandler:** CRUD operations for comments (AddComment, GetCommentsByTodoID, UpdateComment, DeleteComment).

### Domain Models

**Todo** (`internal/model/todo/todo.go`)

```go path=null start=null
type Todo struct {
    ID           uuid.UUID
    UserID       string
    Title        string
    Description  *string
    Status       Status    // draft, active, completed, archived
    Priority     Priority  // low, medium, high
    DueDate      *time.Time
    CompletedAt  *time.Time
    ParentTodoID *uuid.UUID  // For subtasks
    CategoryID   *uuid.UUID
    Metadata     *Metadata   // tags, reminder, color, difficulty
    SortOrder    int
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Category** (`internal/model/category/category.go`)

```go path=null start=null
type Category struct {
    ID          uuid.UUID
    UserID      string
    Name        string
    Color       string
    Description *string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Comment** (`internal/model/comment/comment.go`)

```go path=null start=null
type Comment struct {
    ID        uuid.UUID
    TodoID    uuid.UUID
    UserID    string
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Errors

- **`internal/errs/type.go`**

  - **HTTPError:** Code, Message, Status, Override, Errors (field-level), Action (e.g. redirect). Implements `error` and `Is(*HTTPError)`.

- **`internal/errs/http.go`**

  - Constructors: **NewUnauthorizedError**, **NewForbiddenError**, **NewBadRequestError**, **NewNotFoundError**, **NewInternalServerError**, **ValidationError**. **MakeUpperCaseWithUnderscores** for code formatting.

- **`internal/sqlerr/error.go`**

  - **Code** constants: Other, NotNullViolation, ForeignKeyViolation, UniqueViolation, CheckViolation, etc., with **MapCode** from PostgreSQL codes (23502, 23503, 23505, …).
  - **Severity** and **Error** struct (Code, Severity, Message, TableName, ColumnName, ConstraintName, …). **ConvertPgError** from pgconn.PgError.

- **`internal/sqlerr/handler.go`**
  - **HandleError(err):** If already HTTPError, return as-is. If pgconn.PgError, convert and map to user-facing message and **errs** (BadRequest with optional field errors for not_null, NotFound for no rows, InternalServerError for rest). **ErrNoRows** / **sql.ErrNoRows** → NotFound. Otherwise InternalServerError.
  - Global error handler (in global.go) calls **sqlerr.HandleError** for non-HTTP errors before formatting response.

### Logging & Observability

- **`internal/logger/logger.go`**

  - **LoggerService:** Holds optional New Relic Application. **NewLoggerService** from ObservabilityConfig (app name, license, log forwarding, distributed tracing, optional debug logger). **Shutdown** flushes New Relic.
  - **NewLoggerWithService:** Builds zerolog with level from config, time format, pkgerrors stack marshaler; in production with JSON format and NR app, wraps writer with **zerologWriter** for log forwarding; otherwise console writer in dev. Logger has service, environment; in non-production adds Stack().
  - **WithTraceContext:** Adds trace.id and span.id from New Relic transaction to logger.
  - **NewPgxLogger,** **GetPgxTraceLogLevel:** Used for local DB query logging when env is local.

- New Relic integrations used: main agent, nrecho-v4, nrpgx5, nrredis-v9, nrpkgerrors, logcontext-v2/zerologWriter.

### Services & Repositories

**Repositories** (`internal/repository/`):
- **TodoRepository**: Todo database operations (CRUD, filtering, pagination, batch operations)
- **CategoryRepository**: Category database operations
- **CommentRepository**: Comment database operations

**Services** (`internal/service/services.go`):
- **AuthService**: Clerk integration for authentication
- **TodoService**: Todo business logic (create, update, delete, query with filters)
- **CategoryService**: Category business logic
- **CommentService**: Comment business logic
- **JobService**: Asynq job enqueuing (via server reference)

### Background Jobs

**`internal/lib/jobs/`** - Asynq-based background job processing

- **JobService**: Asynq client + server with priority queues:
  - `critical`: weight 6
  - `default`: weight 3  
  - `low`: weight 1

- **Task Types**:
  - `email:welcome`: Welcome email on signup
  - `email:reminder`: Due date reminder emails
  - `email:weekly_report`: Weekly productivity reports

- **Configuration**: MaxRetry(3), Queue("default"), Timeout(30s)

### Cron Jobs

**`cmd/cron/main.go`** - CLI runner using Cobra

```bash path=null start=null
# List available jobs
go run ./cmd/cron list

# Run a specific job
go run ./cmd/cron due-date-reminders
go run ./cmd/cron weekly-reports
```

**`internal/cron/`** - Job definitions:

| Job | Command | Description |
|-----|---------|-------------|
| `DueDateRemindersJob` | `due-date-reminders` | Enqueue reminders for todos due within N hours |
| `OverdueNotificationsJob` | `overdue-notifications` | Enqueue notifications for overdue todos |
| `WeeklyReportsJob` | `weekly-reports` | Generate and enqueue weekly productivity reports |
| `AutoArchiveJob` | `auto-archive` | Archive completed todos older than N days |

**Scheduling** (example crontab):

```bash path=null start=null
# Daily at 8 AM - due date reminders
0 8 * * * cd /path/to/backend && go run ./cmd/cron due-date-reminders

# Every 4 hours - overdue notifications  
0 */4 * * * cd /path/to/backend && go run ./cmd/cron overdue-notifications

# Weekly on Monday at 9 AM - productivity reports
0 9 * * 1 cd /path/to/backend && go run ./cmd/cron weekly-reports

# Daily at 2 AM - auto-archive old todos
0 2 * * * cd /path/to/backend && go run ./cmd/cron auto-archive
```

### Email

- **`internal/lib/email/client.go`**

  - **Client** wraps Resend client. **SendEmail(to, subject, templateName, data):** Loads HTML from `templates/emails/{templateName}.html`, executes with data, sends via Resend (from: TaskManager &lt;onboarding@resend.dev&gt;).

- **`internal/lib/email/emails.go`**

  - **SendWelcomeEmail(to, firstName):** Uses TemplateWelcome and data UserFirstName.

- **`internal/lib/email/template.go`**

  - **Template** type; **TemplateWelcome** = `"welcome"`.

- **`internal/lib/email/preview.go`**

  - **PreviewData** map for template preview (e.g. welcome → UserFirstName: "John").

- **`templates/emails/welcome.html`**
  - Go HTML template with `{{.UserFirstName}}`, “Welcome to TaskManager!”, CTA, support link.

### Validation

- **`internal/validation/utils.go`**
  - **Validatable** interface: `Validate() error`.
  - **BindAndValidate(c, payload):** Binds payload with `c.Bind(payload)`, then validates with `validateStruct(payload)`. On bind error returns BadRequest with message; on validation error returns BadRequest with **extractValidationErrors** (field + message per tag).
  - **extractValidationErrors:** Handles **validator.ValidationErrors** (required, min, max, oneof, email, e164, uuid, uuidList, dive) and custom **CustomValidationErrors**.
  - **IsValidUUID:** regex for UUID string.

---

## Packages (TypeScript)

- **`packages/zod`**

  - Shared Zod schemas; **@anatine/zod-openapi** for OpenAPI metadata. Exports e.g. **ZHealthResponse** (status, timestamp, environment, checks.database, checks.redis).

- **`packages/openapi`**

  - **ts-rest** contract: health contract (GET /status, response ZHealthResponse). **apiContract** aggregates contracts.
  - **generateOpenApi** with security (bearerAuth, x-service-token), operationMapper for security metadata. **gen.ts** string-replaces custom “file” type with OpenAPI binary, then writes **openapi.json** to repo and (in script) to `../../apps/backend/static/openapi.json` For this repo, add or change the output path in `packages/openapi/src/gen.ts` to `../../backend/static/openapi.json` so `/docs` loads the generated spec.
  - Backend serves `/docs` with Scalar and `/static/openapi.json` so docs stay in sync when you run the openapi package gen.

- **`packages/emails`**
  - Optional React-based email templates (e.g. welcome.tsx); can be used to generate or mirror HTML for backend.

---

## API Endpoints

### System Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/status` | Health check (DB, Redis status) |
| GET | `/docs` | OpenAPI documentation (Scalar UI) |
| GET | `/static/*` | Static files (openapi.json) |

### Todo Routes (Protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/todos` | Create a new todo |
| GET | `/api/v1/todos` | List todos (paginated, filterable) |
| GET | `/api/v1/todos/:id` | Get todo by ID (with category, children, comments) |
| PUT | `/api/v1/todos/:id` | Update a todo |
| DELETE | `/api/v1/todos/:id` | Delete a todo |

**Query Parameters for GET /todos**:
- `page`, `limit`: Pagination (default: page=1, limit=20)
- `sort`: `created_at`, `updated_at`, `title`, `priority`, `due_date`, `status`
- `order`: `asc`, `desc`
- `search`: Full-text search
- `status`: `draft`, `active`, `completed`, `archived`
- `priority`: `low`, `medium`, `high`
- `categoryId`, `parentTodoId`: Filter by relations
- `dueFrom`, `dueTo`: Date range
- `overdue`, `completed`: Boolean filters

### Category Routes (Protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/categories` | Create a category |
| GET | `/api/v1/categories` | List categories |
| PUT | `/api/v1/categories/:id` | Update a category |
| DELETE | `/api/v1/categories/:id` | Delete a category |

### Comment Routes (Protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/todos/:todoId/comments` | Add comment to todo |
| GET | `/api/v1/todos/:todoId/comments` | Get comments for todo |
| PUT | `/api/v1/comments/:id` | Update a comment |
| DELETE | `/api/v1/comments/:id` | Delete a comment |

---

## Tooling

- **Taskfile (backend/Taskfile.yml)**

  - **run:** `go run ./cmd/go-taskmanager`
  - **migrations:new:** `tern new -m ./internal/database/migrations {{.NAME}}` (requires `name=...`)
  - **migrations:up:** `tern migrate -m ./internal/database/migrations --conn-string {{.TASKMANAGER_DB_DSN}}` (with confirm)
  - **tidy:** `go fmt ./...`, `go mod tidy`, `go mod verify`

- **Golangci-lint (backend/.golangci.yml)**

  - Large set of linters (errcheck, staticcheck, gosec, revive, gocritic, etc.) with sensible limits (e.g. cyclop, funlen, gocognit). **gomodguard** blocks old uuid/protobuf modules. **exhaustruct** exclusions for std and third-party structs. **govet** with shadow strict.

- **Root**
  - **package.json** + **turbo.json**: Workspaces `apps/*`, `packages/*`; scripts: build, dev, format, lint, typecheck, clean. Turbo runs tasks with dependency order (^build, etc.).

---

## Environment Variables

All backend config is read from environment with prefix **TASKMANAGER\_**. Keys are lowercased and the prefix is stripped (e.g. `TASKMANAGER_SERVER_PORT` → `server.port`). Nested keys use underscore (e.g. `TASKMANAGER_DATABASE_HOST`).

Example (replace values as needed):

```bash
# Primary
TASKMANAGER_PRIMARY_ENV=local

# Server
TASKMANAGER_SERVER_PORT=8080
TASKMANAGER_SERVER_READ_TIMEOUT=30
TASKMANAGER_SERVER_WRITE_TIMEOUT=30
TASKMANAGER_SERVER_IDLE_TIMEOUT=60
TASKMANAGER_SERVER_CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Database
TASKMANAGER_DATABASE_HOST=localhost
TASKMANAGER_DATABASE_PORT=5432
TASKMANAGER_DATABASE_USER=postgres
TASKMANAGER_DATABASE_PASSWORD=secret
TASKMANAGER_DATABASE_NAME=taskmanager
TASKMANAGER_DATABASE_SSL_MODE=disable
TASKMANAGER_DATABASE_MAX_OPEN_CONNS=25
TASKMANAGER_DATABASE_MAX_IDLE_CONNS=5
TASKMANAGER_DATABASE_CONN_MAX_LIFETIME=300
TASKMANAGER_DATABASE_CONN_MAX_IDLE_TIME=60

# Auth (Clerk)
TASKMANAGER_AUTH_SECRET_KEY=sk_test_...

# Redis
TASKMANAGER_REDIS_ADDRESS=localhost:6379

# Integration (Resend)
TASKMANAGER_INTEGRATION_RESEND_API_KEY=re_...

# Observability (optional)
TASKMANAGER_OBSERVABILITY_SERVICE_NAME=taskmanager
TASKMANAGER_OBSERVABILITY_ENVIRONMENT=development
TASKMANAGER_OBSERVABILITY_LOGGING_LEVEL=debug
TASKMANAGER_OBSERVABILITY_LOGGING_FORMAT=json
TASKMANAGER_OBSERVABILITY_NEW_RELIC_LICENSE_KEY=
TASKMANAGER_OBSERVABILITY_NEW_RELIC_APP_LOG_FORWARDING_ENABLED=true
TASKMANAGER_OBSERVABILITY_NEW_RELIC_DISTRIBUTED_TRACING_ENABLED=true
TASKMANAGER_OBSERVABILITY_NEW_RELIC_DEBUG_LOGGING=false
TASKMANAGER_OBSERVABILITY_HEALTH_CHECKS_ENABLED=true
TASKMANAGER_OBSERVABILITY_HEALTH_CHECKS_INTERVAL=30s
TASKMANAGER_OBSERVABILITY_HEALTH_CHECKS_TIMEOUT=5s
TASKMANAGER_OBSERVABILITY_HEALTH_CHECKS_CHECKS=database,redis
```

For **Taskfile** migrations: set **TASKMANAGER_DB_DSN** (e.g. `postgres://user:pass@localhost:5432/taskmanager?sslmode=disable`).

---

## Running the Project

1. **Prerequisites:** Go 1.25+, PostgreSQL, Redis, Node/Bun for packages.
2. **Env:** Copy or set the variables above (e.g. `.env` and use `godotenv/autoload` or export).
3. **Backend:**
   - From repo root: `cd backend && task run` (or `go run ./cmd/go-taskmanager`).
   - Migrations (non-local): run automatically on startup; for manual run: `TASKMANAGER_DB_DSN=... task migrations:up`.
   - New migration: `task migrations:new name=add_users_table`.
4. **OpenAPI:** From repo root, build/openapi gen so `backend/static/openapi.json` exists (e.g. `cd packages/zod && bun run build && cd ../openapi && bun run gen` if gen writes there). Then open `http://localhost:8080/docs`.
5. **Health:** `GET http://localhost:8080/status`.

---

## Extending the TaskManager

- **New route:** Add to `router/system.go` or a versioned group in `router/router.go`; use `middlewares.Auth.RequireAuth(next)` for protected routes.
- **New handler:** Implement handler func with request/response types implementing **Validatable** where needed; register with **Handle**, **HandleNoContent**, or **HandleFile** from `handler/base.go`.
- **New migration:** `task migrations:new name=your_change` in `backend`, then edit the new file under `internal/database/migrations/`.
- **New job:** Define task type and payload in `internal/lib/jobs`, add handler in `job.go` (mux.HandleFunc), enqueue via `Job.Client.Enqueue(...)` from services/handlers.
- **New email template:** Add template name in `internal/lib/email/template.go`, HTML in `templates/emails/`, and send method in `internal/lib/email/`.
- **OpenAPI:** Add contract in `packages/openapi/src/contracts/`, add Zod types in `packages/zod`, run openapi package gen and copy/openapi.json to `backend/static/` if needed.
- **Config:** Add fields to `config.Config` or `ObservabilityConfig` and corresponding env vars with `TASKMANAGER_` prefix.
