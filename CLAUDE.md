# Platform Development Guide

## Architecture

**Backend services (todo, id) — 3-Layer Design:**
- **Handler** (NATS) - Request/response, protobuf marshaling
- **Service** - Business logic, validation, orchestration
- **Data** (sqlc) - Database queries

Keep layers thin. Service validates, database enforces.

**Web service — Thin frontend:**
- **Handler** (HTTP) - Form parsing, render templates
- **NATSClient** - Typed wrappers around NATS request/reply to backend services
- **Middleware** - Auth validation via `id.auth.validate`

The web service owns no database. All data flows through NATS to backend services.

## Key Decisions

### Database
- **PostgreSQL 18 with UUID v7** - Time-ordered, non-enumerable, distributed-friendly
- **Defense in depth** - Validate in app (friendly errors) AND database (data integrity)
- **Specific queries** - `UpdateListTitle` not generic `Update`. Explicit is better.
- **Manual updated_at** - No triggers. Set explicitly: `updated_at = NOW()`
- **Commit generated code** - sqlc and protobuf are part of source

### Communication
- **NATS + Protobuf** - Request/reply for sync, pub/sub for events
- **Subject pattern:** `service.resource.action` (e.g., `todo.list.create`)
- **Direct publish** - Don't wrap in goroutines, `nc.Publish()` is already non-blocking

### Logging (slog)
- **INFO** - Normal operations (list created, user registered)
- **WARN** - Expected errors (validation failures, bad input)
- **ERROR** - Unexpected errors (database failures, NATS errors)
- **Root logger:** `logger.With("service", "todo")` — set in main.go
- **Module context:** `logger.With("module", "service")` / `"handler"` / `"middleware"`
- **Service layer owns success logs** — handlers only log errors/warnings, never duplicate the service's success log
- **Always include context** — `user_id`, `list_id`, `item_id`, `error` as appropriate

**Why:** ERROR triggers alerts. Validation failures are WARN, not ERROR. Single success log per operation avoids noise.

### Error Handling
- **ValidationError** for expected user errors
- **Translate DB errors** to friendly messages
- **ErrorResponse protobuf** with code + message (no separate field)
- **Shared ErrorResponse** - Use `common/v1/ErrorResponse` across all services

### Authentication
- **ID service** - Handles auth (register, login, token validation)
- **JWT tokens** - Stateless, 7-day expiration, signed with JWT_SECRET
- **Frontend validates** - Web service calls `id.auth.validate` via NATS
- **Backend services trust user_id** - Frontend passes validated user_id to services
- **Database enforces** - All queries include `WHERE user_id = $1`

**Why:** Frontend-first architecture. Users access only through web service (no direct NATS/DB access). Single token validation per request. Backend services focus on business logic, not auth.

## Code Style

- **Simple over complex** - No abstractions until needed 3+ times
- **Explicit over implicit** - Manual `updated_at`, specific queries
- **Don't over-engineer** - No feature flags, premature optimization, or unused helpers

## Project Structure

```
apps/
  todo/               # Todo service
    cmd/server/       # Production entrypoint
    cmd/client/       # Testing client
    internal/
      db/             # sqlc generated
      service/        # Business logic
      handlers/nats/  # NATS handlers
    db/
      schema.sql      # Database schema
      queries.sql     # SQL queries
  id/                 # ID/Auth service
    cmd/server/       # Production entrypoint
      config.go       # Environment config
    cmd/client/       # Testing client
    internal/
      db/             # sqlc generated
      service/        # Auth logic (JWT, bcrypt)
      handlers/nats/  # NATS handlers
    db/
      schema.sql      # Users table
      queries.sql     # User queries
    .env              # JWT_SECRET, DB config
  web/                # Web frontend (no database)
    cmd/server/       # Production entrypoint
    internal/
      handlers/       # HTTP handlers (auth, todo)
      middleware/      # Auth middleware
      natsclient/     # NATS request/reply wrappers
      templates/      # Templ pages
      templates/ui/   # Reusable UI components
    static/css/       # Tailwind input + generated CSS

proto/
  common/v1/          # Shared messages (ErrorResponse)
  todo/v1/            # Todo protobuf definitions
  id/v1/              # Auth protobuf definitions
libs/proto-stubs/     # Generated Go code (committed)
migrations/           # Atlas migrations (all services)
```

## Quick Start

```bash
# Start infrastructure and migrations
just compose-up && sleep 3 && just migrate-apply

# Run all services (id + todo + web on :8080)
just run-all

# Or run individually
cd apps/id && just run-server      # Auth service
cd apps/todo && just run-server    # Todo service
cd apps/web && just run-server     # Web frontend (:8080)

# Test backend services
cd apps/id && just run-client      # Test auth
cd apps/todo && just run-client    # Test todo

# Clean database (drops and recreates)
just db-clean && sleep 3 && just migrate-apply

# Regenerate all generated code (proto stubs, sqlc, templ, CSS)
just generate

# Add new operation to a backend service
1. Add SQL to db/queries.sql
2. just generate (from root)
3. Add service method
4. Add NATS handler + register to subjects map
5. Test with client
```

## Tools

- **sqlc** - Type-safe SQL queries
- **Atlas** - Schema-based migrations (auto-creates atlas_dev)
- **Just** - Task runner (root + per-service Justfiles via `mod`)
- **NATS** - Messaging (request/reply + pub/sub)
- **Protobuf** - Serialization
- **templ** - Type-safe HTML templates (run `templ generate`, commit `*_templ.go`)
- **Tailwind v4** - CSS with semantic tokens (see `apps/web/CLAUDE.md` for token list)
- **HTMX** - In-page interactions on authenticated pages (not used for auth forms)

## Conventions

**Service methods:**
1. Validate input (trim, check required, check length)
2. Call database
3. Log success
4. Publish event (optional)
5. Return

**NATS handler registration:**
```go
// Use map for clean registration
subjects := map[string]nats.MsgHandler{
    "service.resource.action": h.HandleAction,
}
for subject, handler := range subjects {
    nc.Subscribe(subject, handler)
    logger.Info("subscribed to subject", "subject", subject)
}
```

**Authorization:**
- Include `user_id` in WHERE: `WHERE id = $1 AND user_id = $2`
- Database enforces ownership

**Web handlers:**
- Only log errors/warnings (success is logged by the backend service)
- Always include `user_id` and entity IDs in error logs
- Auth forms use plain HTML + 303 redirects (no HTMX)
- Todo pages use HTMX for in-page interactions

**Validation:**
- App: Fast fail with friendly errors
- DB: CHECK constraints as safety net
- Both layers = defense in depth
