# Web Frontend

Server-rendered Go frontend using templ (HTML templates), Tailwind v4, and HTMX.

## Running

```bash
just generate        # Rebuild templ + CSS (run after template changes)
just run-server      # Start on :8080 (needs NATS, id, todo, and log services)
just css-watch       # Auto-rebuild CSS during development
```

Or from the repo root: `just run-all` starts all four services.

## Architecture

The web service is a thin frontend — it owns no database. All data flows through NATS to backend services (id, todo, log).

```
Handler (net/http) → NATSClient → id/todo/log services
```

- **Handlers** (`internal/handlers/`) — HTTP handlers, form parsing, render templates
- **NATSClient** (`internal/natsclient/`) — Typed wrappers around NATS request/reply
- **Middleware** (`internal/middleware/`) — Auth middleware validates tokens via `id.auth.validate`
- **Templates** (`internal/templates/`) — Templ components

## Templates (templ)

Page templates live in `internal/templates/`, reusable UI components in `internal/templates/ui/`.

After editing any `.templ` file, run `just generate` to produce the `*_templ.go` files. Generated files are committed to source.

### UI Components

| Component | File | Usage |
|-----------|------|-------|
| `Layout` | `layout.templ` | HTML shell (head, body, scripts) |
| `PageNarrow` / `PageWide` | `ui/page.templ` | Content width wrappers |
| `Navbar` | `ui/navbar.templ` | Top nav with children slot |
| `Card` | `ui/card.templ` | Surface container |
| `Button` / `ButtonDanger` | `ui/button.templ` | Submit buttons |
| `Input` | `ui/input.templ` | Form input |
| `FormField` | `ui/form_field.templ` | Label + input wrapper |
| `AlertError` | `ui/alert.templ` | Error message banner |

## Styling

Tailwind v4 with a custom dark color scheme. All colors are defined as semantic tokens in `static/css/input.css` using OKLCH.

**Use semantic tokens, not raw Tailwind colors.** Never use `text-gray-*`, `bg-blue-*`, etc.

| Purpose | Token examples |
|---------|---------------|
| Surfaces | `bg-base`, `bg-surface`, `bg-raised` |
| Borders | `border-border`, `border-border-subtle` |
| Text | `text-text-primary`, `text-text-secondary`, `text-text-muted` |
| Brand | `bg-brand`, `text-brand-text`, `hover:bg-brand-hover` |
| Links | `text-link`, `hover:text-link-hover` |
| Inputs | `bg-input-bg`, `border-input-border` |
| Errors | `bg-error-bg`, `text-error-text`, `border-error-border` |
| Danger/Delete | `bg-danger`, `text-delete`, `hover:text-delete-hover` |

## Auth Flow

- Auth pages (login, register) use plain HTML forms with 303 redirects — no HTMX
- Token stored in `auth_token` cookie, treated as opaque by the web service
- Protected routes use `RequireAuth` middleware which validates via NATS
- Validated `user_id` is stored in request context and passed to backend services

## HTMX

Used for in-page interactions on authenticated pages (delete list, toggle item, edit title). Auth pages intentionally avoid HTMX to prevent nested HTML issues with redirects.
