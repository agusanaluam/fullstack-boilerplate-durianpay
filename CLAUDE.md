# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Fullstack payment dashboard — Go backend (SQLite, JWT auth) + Nuxt frontend (not yet scaffolded). OpenAPI-first development: `openapi.yaml` at repo root is the single source of truth for all HTTP contracts.

## Backend Commands

All commands run from `backend/`:

```bash
cp env.sample .env          # first-time setup
make dep                    # go mod tidy + vendor
make tool-openapi           # install oapi-codegen globally
make openapi-gen            # regenerate internal/openapigen/openapi.gen.go from ../openapi.yaml
make gen-secret             # generate JWT_SECRET and print it
make run                    # run server (reads .env)
make build                  # build binary to bin/mygolangapp
```

**Workflow when changing API contracts:** edit `openapi.yaml` → `make openapi-gen` → implement new methods on `APIHandler`.

## Environment Variables

| Var | Default | Notes |
|-----|---------|-------|
| `HTTP_ADDR` | `:8080` | listen address |
| `JWT_SECRET` | `dev-secret-replace-me` | use `make gen-secret` for real value |
| `JWT_EXPIRED` | `24h` | Go duration string |
| `OPENAPIYAML_LOCATION` | `../openapi.yaml` | path from `backend/` to spec file |

## Architecture

```
openapi.yaml  →  oapi-codegen  →  internal/openapigen/openapi.gen.go
                                         ↓ ServerInterface
                               internal/api/api_handler.go (APIHandler)
                                         ↓
                    internal/module/<domain>/handler/   (HTTP decode/encode)
                    internal/module/<domain>/usecase/   (business logic, interface)
                    internal/module/<domain>/repository/ (DB queries, interface)
                    internal/entity/                    (shared domain types)
                    internal/transport/                 (error serialization)
```

- **`APIHandler`** (`internal/api/api_handler.go`) implements the generated `openapigen.ServerInterface` and delegates to domain handlers.
- **Usecase and repository layers use interfaces** — the concrete type wires up in `main.go`.
- **`entity.AppError`** (`internal/entity/error.go`) is the only error type that crosses layer boundaries. Use `entity.ErrorNotFound`, `entity.ErrorUnauthorized`, etc. Transport layer (`transport.WriteError`) maps codes to HTTP status.
- **Database:** SQLite (`dashboard.db`) by default. MySQL driver is also imported. `main.go` runs schema migrations and seeds two users on startup.

## Seeded Test Users

| Email | Password | Role |
|-------|----------|------|
| `cs@test.com` | `password` | cs |
| `operation@test.com` | `password` | operation |

## API Endpoints

- `POST /dashboard/v1/auth/login` — returns `{email, role, token}`
- `GET /dashboard/v1/payments?sort=&status=&id=` — JWT-protected (implementation TODO)

## Adding a New Module

1. Define request/response schemas in `openapi.yaml`.
2. Run `make openapi-gen`.
3. Create `internal/module/<name>/{handler,usecase,repository}/`.
4. Add the handler field to `APIHandler` and implement the generated interface method.
5. Wire the new repo/usecase/handler in `main.go`.
