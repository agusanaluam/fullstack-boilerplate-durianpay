# Payment Dashboard

Internal dashboard for monitoring incoming payments. Go backend + Vue 3 frontend.

## Prerequisites

- Docker & Docker Compose (for quick start)
- Go 1.21+ (local dev only)
- Node 22+ (local dev only)
- make (local dev only — or run commands manually on Windows)

## Quick Start (Docker)

```bash
# Start full stack
make up

# In a separate terminal, seed the database
docker-compose exec backend ./seed

# Open in browser
open http://localhost:3000
```

## Local Development

### Backend

```bash
cd backend
cp env.sample .env
make dep
make openapi-gen
make gen-secret
make seed
make run
```

### Frontend

```bash
cd frontend
npm install
npm run gen:api
npm run dev
```

## Test Accounts

| Email | Password | Role |
|---|---|---|
| cs@test.com | password | cs |
| operation@test.com | password | operation |

## API

Spec: `openapi.yaml` — interactive docs at `http://localhost:8080/swagger`

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/dashboard/v1/auth/login` | No | Login, returns JWT + role |
| GET | `/dashboard/v1/payments` | Bearer JWT | List payments (filter: status, sort, id) |
| GET | `/swagger` | No | Swagger UI |

## Tests

```bash
make test           # run all tests
make test-backend   # Go unit tests
make test-frontend  # Vitest component + store tests
```

### Testing Strategy

**Backend:** Table-driven unit tests with mock interfaces. `auth_test.go` covers login success, wrong password, user not found. `payment_test.go` covers list all, filter by status, sort forwarding.

**Frontend:** Vitest + Vue Test Utils. `auth.spec.ts` verifies store state transitions. `SummaryCards.spec.ts` and `PaymentTable.spec.ts` verify rendering with mock data.

## Architecture

```
frontend (Vue 3 + TS)  →  backend (Go + Chi)  →  SQLite (dashboard.db)
     Pinia + hey-api          oapi-codegen
```

See `docs/superpowers/specs/2026-05-21-payment-dashboard-design.md` for full design spec.
