# Platform

A personal playground for vibe coding and trying out different things. Only public to use GHCR and GitHub Actions.

## What's here

Microservices platform with a server-rendered web frontend. Services communicate over NATS using protobuf, with Postgres for storage.

**Services:**
- **id** — Authentication (register, login, JWT tokens)
- **todo** — Todo lists with shared lists and templates
- **log** — Centralized logging with correlation ID tracing
- **web** — Server-rendered frontend (templ + Tailwind + HTMX)

**Stack:** Go, NATS, PostgreSQL, protobuf, templ, Tailwind v4, HTMX

## Running

```bash
# Start Postgres + NATS, run migrations
just compose-up && sleep 3 && just migrate-apply

# Run all services (web on :8080)
just run-all
```
