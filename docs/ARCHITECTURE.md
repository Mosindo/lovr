# Architecture

This document explains how the fullstack boilerplate is organized so new developers and AI agents can quickly navigate the repository, understand responsibilities, and extend the system without breaking its boundaries.

## 1. Monorepo Structure

The repository is organized into a small number of top-level directories:

- `apps/`: frontend applications
- `services/`: backend services
- `infra/`: local infrastructure and container orchestration
- `scripts/`: smoke tests, QA helpers, and local automation
- `docs/`: architecture and contributor-facing documentation

At the moment, the main applications are:
- `apps/mobile`: React Native mobile client
- `services/api`: Go API

This layout keeps frontend, backend, infrastructure, and documentation concerns separated while still living in one repository.

## 2. Backend Architecture

The backend uses a feature-first architecture in Go with Gin.

```text
services/api/
├─ cmd/api/main.go
└─ internal/
   ├─ platform/
   │  ├─ config/
   │  ├─ db/
   │  ├─ middleware/
   │  ├─ logger/
   │  └─ errors/
   └─ features/
      ├─ auth/
      ├─ users/
      ├─ files/
      ├─ chat/
      ├─ posts/
      ├─ comments/
      └─ notifications/
```

`cmd/api/main.go` is the composition root. It wires the application together by loading configuration, opening the database, registering feature routes, and starting the HTTP server.

Each feature folder groups everything needed for that domain. Instead of scattering handlers, services, and repositories across separate global folders, the code stays close to the feature it belongs to.

Each backend feature follows the same structure:
- `handler.go`: receives HTTP requests, validates input, and returns HTTP responses
- `service.go`: contains business rules and orchestration
- `repository.go`: contains explicit PostgreSQL queries and persistence logic
- `model.go`: defines domain, request, and response models
- `routes.go`: registers the feature's endpoints with Gin

This pattern improves maintainability because:
- feature code is easier to find
- boundaries are explicit
- business logic stays out of handlers
- SQL stays out of handlers and services
- new modules can follow a predictable template

## 3. Platform Layer

The `platform` layer contains infrastructure shared by all backend features.

- `config/`: loads runtime configuration from environment variables such as `DATABASE_URL`, `JWT_SECRET`, and `PORT`
- `db/`: initializes the PostgreSQL connection pool and runs embedded migrations
- `middleware/`: contains reusable HTTP middleware such as auth checks, request metadata handling, and logging hooks
- `logger/`: provides shared logging utilities so features do not each invent their own logging style
- `errors/`: provides shared API error types and helpers for consistent error responses

The purpose of this layer is to centralize technical infrastructure so feature modules stay focused on domain behavior.

## 4. Database Design

The API uses PostgreSQL with explicit SQL queries. There is no ORM.

This is an intentional choice:
- SQL stays visible and reviewable
- query behavior is explicit
- performance characteristics are easier to reason about
- schema changes stay close to the database design

The generic modules rely on core tables such as:
- `users`: account records and identity data
- `posts`: generic user-generated content
- `comments`: responses attached to posts
- `conversations`: chat threads
- `messages`: individual chat messages
- `notifications`: system or user-facing notifications
- `files`: uploaded file metadata and ownership

Related support tables such as `sessions`, `conversation_participants`, and `votes` extend these core capabilities.

## 5. Frontend Architecture

The mobile frontend lives under `apps/mobile/src/`:

```text
apps/mobile/src/
├─ api/
├─ components/
├─ screens/
├─ store/
├─ theme/
└─ utils/
```

Responsibilities:
- `api/`: HTTP client functions and endpoint wrappers
- `components/`: reusable UI building blocks
- `screens/`: route-level UI and screen orchestration
- `store/`: token persistence and lightweight client-side state
- `theme/`: design tokens and shared styling primitives
- `utils/`: shared helpers used across the app

Screens should orchestrate the UI, while shared logic belongs in reusable modules such as API clients, helpers, components, and storage utilities. This keeps screens smaller and makes behavior easier to reuse across future clients.

## 6. Adding a New Feature

New backend features should follow the same feature-first pattern:

```text
internal/features/example/
├─ handler.go
├─ service.go
├─ repository.go
├─ model.go
└─ routes.go
```

Recommended process:

1. Add or update the database migration if the feature needs persistence.
2. Create a new folder under `internal/features/<feature>`.
3. Define models in `model.go`.
4. Implement database access in `repository.go`.
5. Implement business logic in `service.go`.
6. Expose HTTP handlers in `handler.go`.
7. Register routes in `routes.go`.
8. Wire the feature from `cmd/api/main.go`.

When adding a feature, preserve the same architectural rules:
- thin handlers
- explicit services
- SQL in repositories
- reusable platform infrastructure
- clear, predictable module boundaries

This consistency is what makes the boilerplate scalable as more features are added.
