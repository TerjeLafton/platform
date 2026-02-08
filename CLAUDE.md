# Platform Development Guide

## Architecture

**3-Layer Design:**
- **Handler** (NATS) - Request/response, protobuf marshaling
- **Service** - Business logic, validation, orchestration
- **Data** (sqlc) - Database queries

Keep layers thin. Service validates, database enforces.

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
- **INFO** - Normal operations (request received, list created)
- **WARN** - Expected errors (validation failures)
- **ERROR** - Unexpected errors (database failures, bugs)
- **Component context:** `logger.With("component", "service")`

**Why:** ERROR triggers alerts. Validation failures are WARN, not ERROR.

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

# Run services
cd apps/id && just run-server      # Terminal 1 - Auth service
cd apps/todo && just run-server    # Terminal 2 - Todo service

# Test services
cd apps/id && just run-client      # Test auth
cd apps/todo && just run-client    # Test todo

# Clean database (drops and recreates)
just db-clean && sleep 3 && just migrate-apply

# Add new operation to a service
1. Add SQL to db/queries.sql
2. just generate (from root)
3. Add service method
4. Add NATS handler + register to subjects map
5. Test with client
```

## Tools

- **sqlc** - Type-safe SQL queries
- **Atlas** - Schema-based migrations (auto-creates atlas_dev)
- **Just** - Task runner (root + per-service)
- **NATS** - Messaging (request/reply + pub/sub)
- **Protobuf** - Serialization

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

**Validation:**
- App: Fast fail with friendly errors
- DB: CHECK constraints as safety net
- Both layers = defense in depth
