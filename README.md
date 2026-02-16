# TaskManager

A production-ready, full-stack task management application built with Go, PostgreSQL, Redis, and TypeScript. This monorepo implements a REST API for managing todos, categories, and comments with advanced features like hierarchical tasks, background job processing, cron jobs, observability, and Clerk-based authentication.

---

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Backend (Go)](#backend-go)
  - [Entry Point](#entry-point)
  - [Configuration Management](#configuration-management)
  - [Database](#database)
  - [Server Architecture](#server-architecture)
  - [Router and Middleware](#router-and-middleware)
  - [Handlers](#handlers)
  - [Domain Models](#domain-models)
  - [Services and Repositories](#services-and-repositories)
  - [Error Handling](#error-handling)
  - [Logging and Observability](#logging-and-observability)
  - [Background Jobs (Asynq)](#background-jobs-asynq)
  - [Cron Jobs](#cron-jobs)
  - [Email System](#email-system)
  - [Validation](#validation)
- [TypeScript Packages](#typescript-packages)
  - [Zod Schemas](#zod-schemas)
  - [OpenAPI Contracts](#openapi-contracts)
  - [Email Templates](#email-templates)
- [API Endpoints](#api-endpoints)
  - [System Routes](#system-routes)
  - [Todo API](#todo-api)
  - [Category API](#category-api)
  - [Comment API](#comment-api)
- [Environment Variables](#environment-variables)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running the Application](#running-the-application)
- [Development Tools](#development-tools)
- [Extending the Application](#extending-the-application)
- [License](#license)

---

## Project Overview

TaskManager is a comprehensive task management system designed for production environments. It provides a robust REST API for creating, organizing, and managing tasks with support for:

- **Hierarchical Todos**: Parent-child task relationships for subtasks
- **Categories**: Color-coded task organization
- **Comments**: Collaboration through task comments
- **Attachments**: File attachments on tasks (model support)
- **Advanced Filtering**: Complex query support with sorting, filtering, and search
- **Background Processing**: Async email processing via Asynq
- **Scheduled Jobs**: Cron-based reminders, notifications, and reports
- **Authentication**: Clerk-based JWT authentication
- **Observability**: Full New Relic APM integration

---

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Layer                                │
│  (Echo Router → Middleware → Handlers → Response)              │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                               │
│  (Business Logic: Auth, Todo, Category, Comment, Job Services)  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Repository Layer                              │
│  (Database Operations: CRUD, Queries, Pagination)               │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Data Layer                                  │
│  PostgreSQL (pgx v5) + Redis (go-redis/Asynq)                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Features

### Core Features
- **Todo Management**: Full CRUD operations for tasks
- **Hierarchical Tasks**: Support for parent-child relationships (subtasks)
- **Task Prioritization**: Low, Medium, High priority levels
- **Status Management**: Draft, Active, Completed, Archived states
- **Due Dates**: Task due dates with overdue tracking
- **Categories**: Organize tasks into color-coded categories
- **Comments**: Add comments to tasks for collaboration
- **Attachments**: File attachment support (model defined)

### Advanced Features
- **Pagination & Filtering**: Advanced query support with sorting, filtering, and full-text search
- **Background Jobs**: Async email processing via Asynq (Redis-backed)
- **Cron Jobs**: Scheduled tasks for reminders, notifications, reports, and auto-archiving
- **Authentication**: Clerk-based JWT authentication with user roles and permissions
- **Observability**: Full New Relic APM integration with distributed tracing and log forwarding
- **API Documentation**: OpenAPI 3.0 spec with Scalar UI

### Task Statistics
- Total, Draft, Active, Completed, Archived counts
- Overdue task tracking
- Weekly productivity reports

---

## Tech Stack

| Component | Technology | Version |
|-----------|------------|---------|
| **API Framework** | Echo | v4 |
| **Database** | PostgreSQL | Latest |
| **Database Driver** | pgx | v5 |
| **Migrations** | tern | v2 |
| **Cache/Queue** | Redis | Latest |
| **Background Jobs** | Asynq | v0.25+ |
| **Redis Client** | go-redis | v9 |
| **Authentication** | Clerk SDK | v2 |
| **Configuration** | Koanf | v2 |
| **Validation** | go-playground/validator | v10 |
| **Logging** | zerolog | v1.34+ |
| **Observability** | New Relic Go Agent | v3.40+ |
| **API Docs** | ts-rest + Scalar | Latest |
| **CLI** | Cobra | Latest |
| **Monorepo** | Turborepo | v2.5+ |
| **Package Manager** | Bun | 1.2+ |
| **Language (Backend)** | Go | 1.25+ |
| **Language (Frontend)** | TypeScript | 5.8+ |

---

## Project Structure

```
TaskManagement/
├── backend/                          # Go API server
│   ├── cmd/
│   │   ├── taskmanager/              # Main application entry
│   │   └── cron/                     # Cron job runner CLI
│   ├── internal/
│   │   ├── config/                   # Configuration structs and loading
│   │   ├── database/                 # Database connection, migrations
│   │   ├── errs/                     # HTTP error types
│   │   ├── handler/                  # HTTP request handlers
│   │   │   ├── base.go               # Base handler with typed responses
│   │   │   ├── health.go             # Health check endpoint
│   │   │   ├── openapi.go            # OpenAPI UI handler
│   │   │   ├── todo.go               # Todo CRUD handlers
│   │   │   ├── category.go           # Category CRUD handlers
│   │   │   └── comment.go            # Comment CRUD handlers
│   │   ├── lib/
│   │   │   ├── email/                # Resend email client and templates
│   │   │   ├── jobs/                # Asynq job service
│   │   │   └── utils/                # Utility helpers
│   │   ├── logger/                   # zerolog + New Relic integration
│   │   ├── middleware/               # HTTP middleware stack
│   │   │   ├── auth.go               # Clerk authentication
│   │   │   ├── context.go            # Context helpers
│   │   │   ├── global.go             # Global error handler
│   │   │   ├── rate_limit.go         # Rate limiting
│   │   │   ├── request_id.go         # Request ID generation
│   │   │   ├── secure.go             # Security headers
│   │   │   └── tracing.go            # Distributed tracing
│   │   ├── model/                    # Domain models
│   │   │   ├── base.go               # Base model with ID, timestamps
│   │   │   ├── todo/                 # Todo model and related types
│   │   │   ├── category/             # Category model
│   │   │   └── comment/              # Comment model
│   │   ├── repository/               # Database repositories
│   │   │   ├── todo.go
│   │   │   ├── category.go
│   │   │   └── comment.go
│   │   ├── router/                   # HTTP router configuration
│   │   │   ├── router.go             # Main router setup
│   │   │   ├── system.go             # System routes (/status, /docs)
│   │   │   └── v1/                   # API v1 routes
│   │   ├── server/                   # Server initialization
│   │   ├── service/                  # Business logic services
│   │   │   ├── auth.go               # Clerk authentication service
│   │   │   ├── todo.go               # Todo business logic
│   │   │   ├── category.go           # Category business logic
│   │   │   ├── comment.go            # Comment business logic
│   │   │   └── services.go           # Service container
│   │   ├── sqlerr/                   # PostgreSQL error handling
│   │   ├── validation/               # Request validation
│   │   └── cron/                     # Cron job definitions
│   ├── static/                       # Static files (OpenAPI)
│   ├── templates/
│   │   └── emails/                   # HTML email templates
│   ├── Taskfile.yml                  # Task automation
│   ├── .golangci.yml                # Linter configuration
│   ├── go.mod
│   └── go.sum
├── packages/                         # TypeScript packages
│   ├── zod/                         # Shared Zod schemas
│   ├── openapi/                     # ts-rest contracts, OpenAPI generation
│   └── emails/                      # React email templates
├── package.json                      # Workspace root
├── turbo.json                       # Turborepo configuration
└── README.md                        # This file
```

---

## Backend (Go)

### Entry Point

The main application entry point is located at `backend/cmd/taskmanager/main.go`. It handles:

1. **Configuration Loading**: Loads configuration from environment variables with `TASKMANAGER_` prefix
2. **Logger Initialization**: Creates logger with optional New Relic integration
3. **Database Migration**: Runs migrations in non-local environments
4. **Server Initialization**: Creates and configures the HTTP server
5. **Dependency Injection**: Builds repositories, services, and handlers
6. **Router Setup**: Configures routes and middleware
7. **Graceful Shutdown**: Handles interrupt signals for clean shutdown

```go
// Key startup sequence
cfg := config.LoadConfig()
loggerService := logger.NewLoggerService(cfg.Observability)
srv, err := server.New(cfg, &log, loggerService)
repos := repository.NewRepositories(srv)
services, _ := service.NewServices(srv, repos)
handlers := handler.NewHandlers(srv, services)
r := router.NewRouter(srv, handlers, services)
srv.SetupHTTPServer(r)
srv.Start()
```

### Configuration Management

Configuration is managed through environment variables using the Koanf library. All configuration keys use the `TASKMANAGER_` prefix.

**Configuration Structure** (`internal/config/config.go`):

```go
type Config struct {
    Primary       PrimaryConfig
    Server        ServerConfig
    Database      DatabaseConfig
    Auth          AuthConfig
    Redis         RedisConfig
    Integration   IntegrationConfig
    Observability ObservabilityConfig
}

type ServerConfig struct {
    Port              int
    ReadTimeout       int
    WriteTimeout      int
    IdleTimeout       int
    CORSAllowedOrigins []string
}

type DatabaseConfig struct {
    Host            string
    Port            int
    User            string
    Password        string
    Name            string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime int
    ConnMaxIdleTime int
}
```

**Observability Configuration** (`internal/config/observability.go`):

The application supports comprehensive observability with New Relic:
- APM (Application Performance Monitoring)
- Distributed tracing
- Custom events (health checks, rate limit hits)
- Log forwarding
- Database query tracing

### Database

**Database Connection** (`internal/database/database.go`):

- Uses pgx v5 for PostgreSQL connection pooling
- Supports New Relic database tracing (nrpgx5)
- Local development: pgx-zerolog for query logging
- Connection pool configuration with health checks

**Migrations** (`internal/database/migrator.go`):

- Uses tern for migrations
- Embedded migration files in `internal/database/migrations/`
- Auto-runs migrations on startup (non-local environments)
- Manual migration commands via Taskfile

```bash
# Create new migration
task migrations:new name=add_users_table

# Run migrations
task migrations:up
```

### Server Architecture

**Server Struct** (`internal/server/server.go`):

The Server encapsulates all application dependencies:

```go
type Server struct {
    Config        *config.Config
    Logger        *zerolog.Logger
    LoggerService *logger.LoggerService
    DB            *database.Database
    Redis         *redis.Client
    Job           *jobs.JobService
    HTTPServer    *http.Server
}
```

**Server Initialization**:
- Creates PostgreSQL connection pool
- Initializes Redis client with optional New Relic hooks
- Sets up Asynq job server with priority queues
- Configures HTTP server with timeouts

### Router and Middleware

**Router** (`internal/router/router.go`):

The router sets up middleware in the following order:

1. **Global Error Handler**: Catches and formats all errors
2. **Rate Limiter**: 20 req/s with memory store
3. **CORS**: Configured from environment
4. **Secure Headers**: Security headers (HSTS, X-Frame-Options, etc.)
5. **Request ID**: Generates/reads X-Request-ID
6. **New Relic**: APM middleware
7. **Tracing**: Adds trace context to requests
8. **Context Enhancer**: Adds request-scoped logger
9. **Request Logger**: Logs all requests
10. **Recover**: Panics recovery

**System Routes** (`internal/router/system.go`):
- `GET /status` - Health check
- `GET /docs` - OpenAPI documentation (Scalar UI)
- `GET /static/*` - Static files

**API Routes** (`internal/router/v1/`):
- `/api/v1/todos` - Todo CRUD
- `/api/v1/categories` - Category CRUD
- `/api/v1/todos/:todoId/comments` - Comment endpoints

### Handlers

**Base Handler** (`internal/handler/base.go`):

Provides typed handler helpers:

```go
type Handler struct {
    Server   *server.Server
    Services *service.Services
}

// HandlerFunc - Generic handler with request/response types
func (h *Handler) Handle[Req any, Res any](
    handler func(ctx echo.Context, req Req) (Res, error)
) echo.HandlerFunc

// HandleNoContent - For responses without body
func (h *Handler) HandleNoContent[Req any](
    handler func(ctx echo.Context, req Req) error
) echo.HandlerFunc

// HandleFile - For file downloads
func (h *Handler) HandleFile(
    handler func(ctx echo.Context) (File, error)
) echo.HandlerFunc
```

**Response Handlers**:
- `JSONResponseHandler` - JSON responses
- `NoContentResponseHandler` - 204 No Content
- `FileResponseHandler` - File downloads

**Handler Implementations**:
- `health.go` - Health check with DB/Redis pings
- `openapi.go` - Serves OpenAPI UI
- `todo.go` - Todo CRUD operations
- `category.go` - Category CRUD operations
- `comment.go` - Comment CRUD operations

### Domain Models

**Todo** (`internal/model/todo/todo.go`):

```go
type Todo struct {
    model.Base
    UserID       string     `json:"userId" db:"user_id"`
    Title        string     `json:"title" db:"title"`
    Description  *string    `json:"description" db:"description"`
    Status       Status     `json:"status" db:"status"`         // draft, active, completed, archived
    Priority     Priority   `json:"priority" db:"priority"`     // low, medium, high
    DueDate      *time.Time `json:"dueDate" db:"due_date"`
    CompletedAt  *time.Time `json:"completedAt" db:"completed_at"`
    ParentTodoID *uuid.UUID `json:"parentTodoId" db:"parent_todo_id"`
    CategoryID   *uuid.UUID `json:"categoryId" db:"category_id"`
    Metadata     *Metadata  `json:"metadata" db:"metadata"`
    SortOrder    int        `json:"sortOrder" db:"sort_order"`
}

type Metadata struct {
    Tags       []string `json:"tags"`
    Reminder   *string  `json:"reminder"`
    Color      *string  `json:"color"`
    Difficulty *int     `json:"difficulty"`
}

// Status values
const (
    StatusDraft     Status = "draft"
    StatusActive    Status = "active"
    StatusCompleted Status = "completed"
    StatusArchived  Status = "archived"
)

// Priority values
const (
    PriorityLow    Priority = "low"
    PriorityMedium Priority = "medium"
    PriorityHigh   Priority = "high"
)

// PopulatedTodo includes related data
type PopulatedTodo struct {
    Todo
    Category    *category.Category  `json:"category"`
    Children    []Todo              `json:"children"`
    Comments    []comment.Comment   `json:"comments"`
    Attachments []TodoAttachment    `json:"attachments"`
}
```

**Category** (`internal/model/category/category.go`):

```go
type Category struct {
    model.Base
    UserID      string  `json:"userId" db:"user_id"`
    Name        string  `json:"name" db:"name"`
    Color       string  `json:"color" db:"color"`
    Description *string `json:"description" db:"description"`
}
```

**Comment** (`internal/model/comment/comment.go`):

```go
type Comment struct {
    model.Base
    TodoID  uuid.UUID `json:"todoId" db:"todo_id"`
    UserID  string    `json:"userId" db:"user_id"`
    Content string    `json:"content" db:"content"`
}
```

**Base Model** (`internal/model/base.go`):

```go
type Base struct {
    ID        uuid.UUID `json:"id" db:"id"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type PaginatedResponse[T any] struct {
    Data       []T `json:"data"`
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"totalPages"`
}
```

### Services and Repositories

**Services** (`internal/service/`):

- `auth.go` - Clerk authentication service
- `todo.go` - Todo business logic (create, update, delete, query, filtering)
- `category.go` - Category business logic
- `comment.go` - Comment business logic
- `services.go` - Service container

**Repositories** (`internal/repository/`):

- `todo.go` - Todo database operations (CRUD, filtering, pagination, batch)
- `category.go` - Category database operations
- `comment.go` - Comment database operations

### Error Handling

**HTTP Errors** (`internal/errs/`):

```go
type HTTPError struct {
    Code     int               `json:"code"`
    Message  string            `json:"message"`
    Status   int               `json:"status"`
    Override bool              `json:"override"`
    Errors   map[string]string `json:"errors,omitempty"`
    Action   string            `json:"action,omitempty"`
}
```

**Error Constructors** (`internal/errs/http.go`):
- `NewUnauthorizedError`
- `NewForbiddenError`
- `NewBadRequestError`
- `NewNotFoundError`
- `NewInternalServerError`
- `ValidationError`

**PostgreSQL Error Handling** (`internal/sqlerr/`):

Maps PostgreSQL error codes to HTTP errors:
- `UniqueViolation` → 409 Conflict
- `NotNullViolation` → 400 Bad Request
- `ForeignKeyViolation` → 400 Bad Request
- `CheckViolation` → 400 Bad Request

### Logging and Observability

**Logger Service** (`internal/logger/logger.go`):

- Uses zerolog for structured logging
- Optional New Relic integration for log forwarding
- Different log formats for development (console) vs production (JSON)
- Request-scoped logging with trace context

**New Relic Integration**:
- APM (Application Performance Monitoring)
- Distributed tracing
- Custom events:
  - `HealthCheckError` - Health check failures
  - `RateLimitHit` - Rate limit exceeded
- Database query tracing (nrpgx5)
- Redis command tracing (nrredis-v9)
- HTTP handler tracing (nrecho-v4)

### Background Jobs (Asynq)

**Job Service** (`internal/lib/jobs/`):

The application uses Asynq for background job processing with priority queues:

```go
// Queue configuration
const (
    CriticalQueue = "critical"  // Weight: 6
    DefaultQueue = "default"    // Weight: 3
    LowQueue     = "low"        // Weight: 1
)
```

**Task Types**:
- `email:welcome` - Welcome email on user signup
- `email:reminder` - Due date reminder emails
- `email:weekly_report` - Weekly productivity reports

**Job Configuration**:
- MaxRetry: 3
- Queue: default
- Timeout: 30s

### Cron Jobs

**Cron Job Runner** (`cmd/cron/main.go`):

Cobra-based CLI for running scheduled jobs:

```bash
# List available jobs
go run ./cmd/cron list

# Run specific jobs
go run ./cmd/cron due-date-reminders
go run ./cmd/cron weekly-reports
go run ./cmd/cron overdue-notifications
go run ./cmd/cron auto-archive
```

**Cron Job Definitions** (`internal/cron/`):

| Job | Command | Description |
|-----|---------|-------------|
| DueDateRemindersJob | due-date-reminders | Enqueue reminders for todos due within N hours |
| OverdueNotificationsJob | overdue-notifications | Enqueue notifications for overdue todos |
| WeeklyReportsJob | weekly-reports | Generate and enqueue weekly productivity reports |
| AutoArchiveJob | auto-archive | Archive completed todos older than N days |

**Example Crontab**:

```bash
# Daily at 8 AM - due date reminders
0 8 * * * cd /path/to/backend && go run ./cmd/cron due-date-reminders

# Every 4 hours - overdue notifications
0 */4 * * * cd /path/to/backend && go run ./cmd/cron overdue-notifications

# Weekly on Monday at 9 AM - productivity reports
0 9 * * 1 cd /path/to/backend && go run ./cmd/cron weekly-reports

# Daily at 2 AM - auto-archive old todos
0 2 * * * cd /path/to/backend && go run ./cmd/cron auto-archive
```

### Email System

**Email Client** (`internal/lib/email/`):

- Uses Resend for email delivery
- HTML template support with Go templates

**Email Templates** (`templates/emails/`):
- `welcome.html` - Welcome email template

**Email Functions**:
- `SendWelcomeEmail(to, firstName)` - Send welcome email

### Validation

**Validation** (`internal/validation/utils.go`):

Uses go-playground/validator for request validation:

```go
// BindAndValidate - Bind request and validate
func BindAndValidate(c echo.Context, payload any) error

// Validatable interface for custom validation
type Validatable interface {
    Validate() error
}
```

**Supported Validations**:
- Required fields
- Min/Max length
- OneOf enums
- Email format
- UUID format
- Custom validation errors

---

## TypeScript Packages

### Zod Schemas

**Location**: `packages/zod`

Shared Zod schemas for request/response validation:

- `ZHealthResponse` - Health check response schema
- TypeScript types matching Go models

Uses `@anatine/zod-openapi` for OpenAPI metadata.

### OpenAPI Contracts

**Location**: `packages/openapi`

- **ts-rest** contracts for type-safe API definitions
- **Contracts**:
  - Health contract (`health.ts`)
  - Todo contract (`todo.ts`)
  - Category contract (`category.ts`)
  - Comment contract (`comment.ts`)
- **OpenAPI Generation**: Generates `openapi.json` for API documentation
- **Scalar UI**: Interactive API documentation at `/docs`

### Email Templates

**Location**: `packages/emails`

React-based email templates (optional):
- Can be used to generate HTML for backend templates
- Type-safe email component library

---

## API Endpoints

### System Routes

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/status` | Health check | No |
| GET | `/docs` | OpenAPI documentation | No |
| GET | `/static/*` | Static files | No |

### Todo API

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/api/v1/todos` | Create a new todo | Required |
| GET | `/api/v1/todos` | List todos (paginated, filterable) | Required |
| GET | `/api/v1/todos/:id` | Get todo by ID | Required |
| PUT | `/api/v1/todos/:id` | Update a todo | Required |
| DELETE | `/api/v1/todos/:id` | Delete a todo | Required |

**Query Parameters for GET /todos**:
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20)
- `sort` - Sort field: `created_at`, `updated_at`, `title`, `priority`, `due_date`, `status`
- `order` - Sort order: `asc`, `desc`
- `search` - Full-text search
- `status` - Filter by status: `draft`, `active`, `completed`, `archived`
- `priority` - Filter by priority: `low`, `medium`, `high`
- `categoryId` - Filter by category UUID
- `parentTodoId` - Filter by parent todo UUID
- `dueFrom` - Due date range start
- `dueTo` - Due date range end
- `overdue` - Boolean filter for overdue tasks
- `completed` - Boolean filter for completed tasks

### Category API

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/api/v1/categories` | Create a category | Required |
| GET | `/api/v1/categories` | List categories | Required |
| GET | `/api/v1/categories/:id` | Get category by ID | Required |
| PUT | `/api/v1/categories/:id` | Update a category | Required |
| DELETE | `/api/v1/categories/:id` | Delete a category | Required |

### Comment API

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/api/v1/todos/:todoId/comments` | Add comment to todo | Required |
| GET | `/api/v1/todos/:todoId/comments` | Get comments for todo | Required |
| PUT | `/api/v1/comments/:id` | Update a comment | Required |
| DELETE | `/api/v1/comments/:id` | Delete a comment | Required |

---

## Environment Variables

All configuration is managed through environment variables with the `TASKMANAGER_` prefix.

### Required Variables

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
```

### Optional Variables

```bash
# Resend Email Integration
TASKMANAGER_INTEGRATION_RESEND_API_KEY=re_...

# Observability (New Relic)
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

---

## Getting Started

### Prerequisites

- **Go**: 1.25 or higher
- **PostgreSQL**: Latest stable version
- **Redis**: Latest stable version
- **Node.js**: 22 or higher (for TypeScript packages)
- **Bun**: 1.2 or higher
- **Task** (optional): For Taskfile automation

### Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd TaskManagement
   ```

2. **Install dependencies**:
   ```bash
   # Install Go dependencies
   cd backend
   go mod download

   # Install Node dependencies (from root)
   bun install
   ```

3. **Configure environment**:
   ```bash
   # Copy the example environment file
   cp backend/.env.example backend/.env

   # Edit backend/.env with your configuration
   ```

4. **Set up the database**:
   ```bash
   # Create PostgreSQL database
   createdb taskmanager

   # Run migrations (non-local environments)
   cd backend
   task migrations:up
   ```

### Running the Application

**Development Mode**:

```bash
# Run the backend server
cd backend
task run

# Or directly
go run ./cmd/taskmanager
```

**Run Cron Jobs**:

```bash
# List available cron jobs
go run ./cmd/cron list

# Run a specific job
go run ./cmd/cron due-date-reminders
go run ./cmd/cron weekly-reports
```

**Generate OpenAPI Documentation**:

```bash
# From packages/openapi
cd packages/openapi
bun run gen
```

**Access the API**:
- API: http://localhost:8080
- Health: http://localhost:8080/status
- Docs: http://localhost:8080/docs

---

## Development Tools

### Taskfile (backend/Taskfile.yml)

```bash
# Run the server
task run

# Create new migration
task migrations:new name=add_users_table

# Run migrations
task migrations:up

# Format and tidy Go code
task tidy
```

### Golangci-lint

The project uses golangci-lint with extensive linters:
- errcheck, staticcheck, gosec, revive, gocritic
- Cyclomatic complexity limits
- Function length limits
- Custom module restrictions

Run linting:
```bash
golangci-lint run
```

### Turborepo

From the root directory:

```bash
# Build all packages
bun run build

# Run dev mode
bun run dev

# Type check all packages
bun run typecheck

# Lint all packages
bun run lint

# Format code
bun run format:fix
```

---

## Extending the Application

### Adding a New Route

1. Create a new handler in `internal/handler/`
2. Register the route in `internal/router/v1/`
3. Add service methods in `internal/service/`
4. Add repository methods in `internal/repository/`

### Adding a New Model

1. Create model in `internal/model/`
2. Create repository in `internal/repository/`
3. Create service in `internal/service/`
4. Create handler in `internal/handler/`
5. Add Zod schema in `packages/zod/`
6. Add ts-rest contract in `packages/openapi/`
7. Generate OpenAPI spec

### Adding a Background Job

1. Define task type in `internal/lib/jobs/`
2. Add handler in job server (`internal/lib/jobs/job.go`)
3. Enqueue from service using `Job.Client.Enqueue()`

### Adding an Email Template

1. Add template name in `internal/lib/email/template.go`
2. Create HTML template in `templates/emails/`
3. Add send function in `internal/lib/email/`

---

## License

MIT License - See LICENSE file for details
