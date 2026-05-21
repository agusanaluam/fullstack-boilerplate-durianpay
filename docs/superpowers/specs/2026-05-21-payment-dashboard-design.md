# Payment Dashboard — Design Spec
_Date: 2026-05-21_

---

## 1. Overview

Internal dashboard for monitoring incoming payments. Go backend (existing boilerplate extended) + Vue 3 frontend. Reviewed on Mac with Go v1.21+, Node v20+, Docker & Docker Compose.

**Deadline:** 72 hours from assignment receipt.

---

## 2. Tech Stack

| Layer | Choice |
|---|---|
| Backend | Go 1.21+, Chi router, oapi-codegen |
| Backend DB | SQLite (existing `database/sql` pattern) |
| Frontend | Vue 3 + TypeScript |
| Frontend UI | shadcn-vue + Tailwind CSS |
| Frontend State | Pinia |
| Frontend API client | @hey-api/openapi-ts (generated from `openapi.yaml`) |
| Containerisation | Docker Compose (2 services) |

---

## 3. System Architecture

```
┌─────────────────┐     HTTP      ┌──────────────────┐
│  Frontend :3000 │ ────────────► │  Backend  :8080  │
│  Vue 3 + TS     │               │  Go + Chi        │
│  Pinia stores   │               │  oapi-codegen    │
│  hey-api client │               │  JWT middleware  │
└─────────────────┘               └────────┬─────────┘
                                           │
                                ┌──────────▼──────────┐
                                │   SQLite (file)     │
                                │   dashboard.db      │
                                │   users table       │
                                │   payments table    │
                                └─────────────────────┘
```

**Docker Compose services:**
- `backend` — Go binary, mounts `dashboard.db` volume
- `frontend` — Vite build served via nginx, depends_on: backend

---

## 4. Backend Design

### 4.1 Directory Changes

```
backend/
├── internal/
│   ├── entity/
│   │   ├── user.go           # existing — unchanged
│   │   └── payment.go        # NEW: Payment struct
│   ├── module/
│   │   ├── auth/             # existing — unchanged
│   │   │   ├── handler/auth.go
│   │   │   ├── usecase/auth.go
│   │   │   └── repository/user.go
│   │   └── payment/          # NEW module
│   │       ├── handler/payment.go
│   │       ├── usecase/payment.go
│   │       └── repository/payment.go
│   ├── config/
│   │   └── env.go            # existing — unchanged (no new env vars needed)
│   └── api/
│       └── api_handler.go    # wire new payment handler
├── script/
│   ├── gen-secret/main.go    # existing
│   └── seed/main.go          # NEW: random payment + user seeder
└── main.go                   # add payments table to initDB(), wire payment module
```

### 4.2 SQLite Schema

Added to `initDB()` in `main.go`:

```sql
CREATE TABLE IF NOT EXISTS payments (
  id         TEXT PRIMARY KEY,
  merchant   TEXT NOT NULL,
  amount     INTEGER NOT NULL,
  status     TEXT NOT NULL CHECK(status IN ('completed','processing','failed')),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Existing `users` table unchanged.

### 4.3 Payment Entity

**`internal/entity/payment.go`:**
```go
package entity

import "time"

type Payment struct {
    ID        string    `json:"id"`
    Merchant  string    `json:"merchant"`
    Amount    int64     `json:"amount"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 4.4 Environment Variables

No new env vars needed. Existing `env.sample`:
```env
HTTP_ADDR=:8080
OPENAPIYAML_LOCATION=../openapi.yaml
JWT_SECRET=your-very-secret
JWT_EXPIRED=24h
```

### 4.5 API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/dashboard/v1/auth/login` | No | Returns JWT + role |
| GET | `/dashboard/v1/payments` | Bearer JWT | List/filter payments |

Query params for GET payments:
- `status` — `completed` / `processing` / `failed`
- `sort` — `-created_at` (desc, default) / `created_at` (asc) / `-amount` / `amount`
- `id` — exact payment ID match

### 4.6 JWT Middleware

Custom Chi middleware that:
1. Reads `Authorization: Bearer <token>` header
2. Validates with `golang-jwt/jwt/v5`
3. Rejects with 401 if missing/expired/invalid
4. Injects claims into `context.Context`

Applied to all `/dashboard/v1/payments*` routes via Chi group in `server.go`.

### 4.7 Payment Repository

**`internal/module/payment/repository/payment.go`** — uses `*sql.DB`, same pattern as existing user repo:

```go
type PaymentRepository interface {
    ListPayments(status, sort, id string) ([]*entity.Payment, error)
}
```

SQL query builds `WHERE` and `ORDER BY` dynamically based on non-empty params.

### 4.8 Backend Unit Tests

- `internal/module/auth/usecase/auth_test.go` (existing pattern, add if missing)
  - Login success → returns token + user
  - Wrong password → `ErrorCodeUnauthorized`
  - User not found → `ErrorCodeNotFound`
- `internal/module/payment/usecase/payment_test.go`
  - List all payments → returns full list
  - Filter by status → passes status to repo
  - Sort param forwarded correctly

Mock interfaces: `UserRepository`, `PaymentRepository`.

---

## 5. Frontend Design

### 5.1 Project Structure

```
frontend/
├── src/
│   ├── api/
│   │   └── generated/        # @hey-api/openapi-ts output — do not edit
│   ├── components/
│   │   ├── ui/               # shadcn-vue primitives (Button, Input, Badge, Select, Table)
│   │   ├── PaymentTable.vue
│   │   ├── SummaryCards.vue
│   │   └── AppSidebar.vue
│   ├── pages/
│   │   ├── LoginPage.vue
│   │   └── DashboardPage.vue
│   ├── stores/
│   │   ├── auth.ts           # Pinia: token, role, login(), logout()
│   │   └── payments.ts       # Pinia: list, filters, summary, fetchPayments()
│   ├── router/
│   │   └── index.ts          # Vue Router: / → login, /dashboard → protected
│   └── main.ts
├── openapi-ts.config.ts      # hey-api config → ../../openapi.yaml
└── vite.config.ts            # proxy /dashboard → http://localhost:8080
```

### 5.2 Pages

**LoginPage (`/`)**
- Split layout: left purple gradient panel (brand + tagline) / right white card form
- Fields: email + password (show/hide toggle)
- On submit: call generated `postDashboardV1AuthLogin()`, store token + role in `authStore`, redirect to `/dashboard`
- On error: inline error banner
- Font: Arial

**DashboardPage (`/dashboard`)**
- Protected: navigation guard checks `authStore.token`, redirects to `/` if null
- Layout: fixed left sidebar + scrollable main content
- Sidebar: logo, "Payments" nav item (active + total badge), user chip (email + role) + logout button
- Main: topbar (title + date) → 4 summary cards → table section
- Summary cards: Total / Completed / Processing / Failed — derived from full unfiltered fetch
- Table columns: Payment ID, Merchant, Date, Amount, Status
- Filters: search input (client-side, by ID or merchant) + status dropdown + sort dropdown
- Status badge colors: green (completed), amber (processing), red (failed)
- Pagination: client-side, 10 rows per page

### 5.3 Pinia Stores

**`authStore`** (`stores/auth.ts`):
```ts
state: { token: string | null, role: string | null }
actions: login(email: string, password: string): Promise<void>
         logout(): void
persist: localStorage via pinia-plugin-persistedstate
```

**`paymentsStore`** (`stores/payments.ts`):
```ts
state: {
  payments: Payment[]
  loading: boolean
  error: string | null
  filterStatus: string   // '' | 'completed' | 'processing' | 'failed'
  filterSort: string     // '-created_at' | 'created_at' | '-amount' | 'amount'
  filterSearch: string   // client-side text filter
}
getters: {
  summary: { total, completed, processing, failed }
  filteredPayments: Payment[]   // search applied client-side
}
actions: fetchPayments(): Promise<void>   // calls API with filterStatus + filterSort
```

### 5.4 OpenAPI Client Generation

Config `openapi-ts.config.ts`:
```ts
import { defineConfig } from '@hey-api/openapi-ts'
export default defineConfig({
  input: '../openapi.yaml',
  output: 'src/api/generated',
  plugins: ['@hey-api/client-fetch'],
})
```

Run: `npm run gen:api` (`npx @hey-api/openapi-ts`). Generated types used directly in stores — no manual API type definitions.

### 5.5 Frontend Tests

- `src/stores/auth.spec.ts` — login stores token/role, logout clears state
- `src/components/SummaryCards.spec.ts` — correct counts from mock payment array
- `src/components/PaymentTable.spec.ts` — renders rows, status badge correct color
- Tool: Vitest + Vue Test Utils

---

## 6. Docker Compose

```yaml
services:
  backend:
    build: ./backend
    ports: ["8080:8080"]
    volumes: ["sqlite_data:/app/data"]
    environment:
      - HTTP_ADDR=:8080
      - JWT_SECRET=dev-secret-replace-me
      - JWT_EXPIRED=24h
      - OPENAPIYAML_LOCATION=../openapi.yaml

  frontend:
    build: ./frontend
    ports: ["3000:80"]
    depends_on: [backend]

volumes:
  sqlite_data:
```

`dashboard.db` stored at `/app/data/dashboard.db` inside container via volume mount.

**One-command startup:**
```bash
docker-compose up --build
```

---

## 7. Root Makefile Targets

```makefile
dev-backend:
    cd backend && make run

dev-frontend:
    cd frontend && npm run dev

gen-api:
    cd frontend && npm run gen:api

seed:
    cd backend && go run ./script/seed/main.go

test-backend:
    cd backend && go test ./...

test-frontend:
    cd frontend && npm run test

up:
    docker-compose up --build

down:
    docker-compose down
```

---

## 8. Seed Data

Dedicated script: `backend/script/seed/main.go`. Run via `make seed`.

**Behavior:**
- Idempotent — skips if `payments` table already has rows
- Opens same SQLite file (`dashboard.db`) as the app
- Seeds 2 users: `cs@test.com` + `operation@test.com`, password `password` (bcrypt hashed) — skips if users already exist
- Seeds 50 payments with `math/rand`:
  - `id`: sequential `PAY-001` … `PAY-050`
  - `merchant`: random pick from pool of 15 merchant names
  - `amount`: random Rp 10,000–500,000 rounded to nearest 1,000
  - `status`: weighted — 60% `completed`, 25% `processing`, 15% `failed`
  - `created_at`: random timestamp within last 90 days

---

## 9. Testing Strategy (for README)

**Backend:** Table-driven unit tests with mock interfaces. Auth usecase tests cover happy path + error branches. Payment usecase tests verify filter/sort params forwarded correctly to repo. Run: `cd backend && go test ./...`

**Frontend:** Vitest + Vue Test Utils. Store tests verify state transitions. Component tests verify rendering with mock data. Run: `cd frontend && npm run test`

---

## 10. Out of Scope

- `PUT /dashboard/v1/payment/{id}/review`
- Dark mode
- Real-time updates (WebSocket)
- Role-based feature differences beyond login
