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
| Backend DB | MongoDB (replaces SQLite) |
| Backend Cache | Redis (payments query cache) |
| Frontend | Vue 3 + TypeScript |
| Frontend UI | shadcn-vue + Tailwind CSS |
| Frontend State | Pinia |
| Frontend API client | @hey-api/openapi-ts (generated from `openapi.yaml`) |
| Containerisation | Docker Compose (4 services) |

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
                               ┌───────────┴───────────┐
                               │                       │
                    ┌──────────▼──────────┐  ┌─────────▼────────┐
                    │   MongoDB  :27017   │  │  Redis   :6379   │
                    │  users collection  │  │  payments cache  │
                    │  payments collection│  │  TTL: 60s        │
                    └─────────────────────┘  └──────────────────┘
```

**Docker Compose services:**
- `mongodb` — official mongo:7 image
- `redis` — official redis:7-alpine image
- `backend` — Go binary, depends_on: mongodb + redis
- `frontend` — Node build served via Vite preview or nginx, depends_on: backend

---

## 4. Backend Design

### 4.1 Directory Changes

```
backend/
├── internal/
│   ├── db/
│   │   ├── mongo.go          # MongoDB client + collection helpers
│   │   └── redis.go          # Redis client setup
│   ├── entity/
│   │   ├── user.go           # existing — unchanged
│   │   └── payment.go        # NEW: Payment struct
│   ├── module/
│   │   ├── auth/             # existing — repo rewritten for MongoDB
│   │   │   ├── handler/auth.go
│   │   │   ├── usecase/auth.go
│   │   │   └── repository/user.go   # rewrite: SQL → MongoDB
│   │   └── payment/          # NEW module
│   │       ├── handler/payment.go
│   │       ├── usecase/payment.go   # Redis cache logic here
│   │       └── repository/payment.go
│   └── api/
│       └── api_handler.go    # wire new payment handler
└── main.go                   # wire MongoDB + Redis clients, seed payments
```

### 4.2 MongoDB Collections

**`users`**
```json
{ "_id": ObjectId, "email": "cs@test.com", "password_hash": "...", "role": "cs" }
```

**`payments`** — seeded with 50 records at startup
```json
{
  "_id": ObjectId,
  "id": "PAY-001",
  "merchant": "Merchant Alpha",
  "amount": 150000,
  "status": "completed",
  "created_at": ISODate
}
```
Status distribution: ~60% completed, ~25% processing, ~15% failed.

### 4.3 Environment Variables

```env
HTTP_ADDR=:8080
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DB=dashboard
REDIS_ADDR=redis:6379
JWT_SECRET=your-secret
JWT_EXPIRED=24h
```

### 4.4 API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/dashboard/v1/auth/login` | No | Returns JWT + role |
| GET | `/dashboard/v1/payments` | Bearer JWT | List/filter payments |

Query params for GET payments:
- `status` — `completed` / `processing` / `failed`
- `sort` — `-created_at` (desc) / `created_at` (asc) / `-amount` / `amount`
- `id` — exact payment ID match

### 4.5 Redis Cache Strategy

- **Key:** `payments:status={s}&sort={s}&id={s}` (empty string for unset params)
- **TTL:** 60 seconds
- **Hit:** return cached JSON bytes, skip MongoDB
- **Miss:** query MongoDB → marshal → SET with TTL → return
- **Invalidation:** TTL-based only (payments are read-only in this scope)

### 4.6 JWT Middleware

Custom Chi middleware that:
1. Reads `Authorization: Bearer <token>` header
2. Validates with `golang-jwt/jwt/v5`
3. Rejects with 401 if missing/expired/invalid
4. Injects claims into `context.Context`

Applied to all `/dashboard/v1/payments*` routes.

### 4.7 Backend Unit Tests

- `internal/module/auth/usecase/auth_test.go`
  - Login success
  - Wrong password → 401
  - User not found → 404
- `internal/module/payment/usecase/payment_test.go`
  - List all payments (cache miss → DB)
  - Filter by status
  - Cache hit skips DB call
  - Sort applied correctly

Mock interfaces: `UserRepository`, `PaymentRepository`, `CacheClient`.

---

## 5. Frontend Design

### 5.1 Project Structure

```
frontend/
├── src/
│   ├── api/              # @hey-api/openapi-ts generated client
│   │   └── generated/    # auto-generated, do not edit
│   ├── components/
│   │   ├── ui/           # shadcn-vue primitives (Button, Input, Badge, Select, Table…)
│   │   ├── PaymentTable.vue
│   │   ├── SummaryCards.vue
│   │   ├── StatusFilter.vue
│   │   └── AppSidebar.vue
│   ├── pages/
│   │   ├── LoginPage.vue
│   │   └── DashboardPage.vue
│   ├── stores/
│   │   ├── auth.ts       # Pinia: token, role, login(), logout()
│   │   └── payments.ts   # Pinia: list, filters, summary counts, fetchPayments()
│   ├── router/
│   │   └── index.ts      # Vue Router: / → login, /dashboard → protected
│   └── main.ts
├── openapi-ts.config.ts  # hey-api config pointing to ../../openapi.yaml
└── vite.config.ts        # proxy /dashboard → http://backend:8080
```

### 5.2 Pages

**LoginPage (`/`)**
- Split layout: left purple gradient panel (brand + tagline) / right white card form
- Fields: email, password (with show/hide toggle)
- On submit: call generated `postDashboardV1AuthLogin()`, store token + role in `authStore`, redirect to `/dashboard`
- On error: show inline error banner
- Font: Arial

**DashboardPage (`/dashboard`)**
- Protected: Vue Router navigation guard checks `authStore.token`, redirects to `/` if missing
- Layout: fixed left sidebar + scrollable main content
- Sidebar: logo, nav item "Payments" (active + badge count), user chip with role + logout button
- Main: topbar (title + date) → 4 summary cards → table section
- Summary cards: Total / Completed / Processing / Failed — counts derived from full unfiltered list
- Table: columns Payment ID, Merchant, Date, Amount, Status
- Filters: search input (by ID or merchant, client-side) + status dropdown + sort dropdown
- Status badge colors: green (completed), amber (processing), red (failed)
- Pagination: client-side, 10 rows per page

### 5.3 Pinia Stores

**`authStore`**
```ts
state: { token: string | null, role: string | null }
actions: login(email, password), logout()
persist: localStorage
```

**`paymentsStore`**
```ts
state: {
  payments: Payment[],
  loading: boolean,
  error: string | null,
  filterStatus: string,
  filterSort: string,
  filterSearch: string
}
getters: { summary, filteredPayments }
actions: fetchPayments(status?, sort?, id?)
```

### 5.4 OpenAPI Client Generation

```bash
cd frontend && npx @hey-api/openapi-ts \
  --input ../openapi.yaml \
  --output src/api/generated \
  --client fetch
```

Run via `npm run gen:api`. Generated types used directly in stores and components — no manual type definitions for API shapes.

### 5.5 Frontend Tests

- `PaymentTable.spec.ts` — renders rows, filter pills update visible rows
- `SummaryCards.spec.ts` — correct counts from mock payment list
- `authStore.spec.ts` — login sets token, logout clears state
- Tool: Vitest + Vue Test Utils

---

## 6. Docker Compose

```yaml
services:
  mongodb:
    image: mongo:7
    ports: ["27017:27017"]
    volumes: [mongo_data:/data/db]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  backend:
    build: ./backend
    ports: ["8080:8080"]
    env_file: ./backend/.env.docker
    depends_on: [mongodb, redis]

  frontend:
    build: ./frontend
    ports: ["3000:3000"]
    depends_on: [backend]

volumes:
  mongo_data:
```

`backend/.env.docker` uses `mongodb://mongodb:27017` and `redis:6379` (Docker network hostnames).

**One-command startup:**
```bash
docker-compose up --build
```

---

## 7. Root Makefile Targets

```makefile
dev-backend   # cd backend && make run
dev-frontend  # cd frontend && npm run dev
gen-api       # cd frontend && npm run gen:api
test-backend  # cd backend && go test ./...
test-frontend # cd frontend && npm run test
up            # docker-compose up --build
down          # docker-compose down
```

---

## 8. Seed Data

Backend `main.go` seeds on startup (idempotent — skips if payments collection non-empty):
- 2 users: `cs@test.com` / `operation@test.com`, password: `password`
- 50 payments: random merchants, amounts (Rp 10,000–Rp 500,000), mixed statuses

---

## 9. Testing Strategy (for README)

**Backend:** Table-driven unit tests with mock interfaces. Auth usecase tests cover happy path + error cases. Payment usecase tests cover cache hit/miss + filter/sort logic. Run: `go test ./...`

**Frontend:** Vitest component tests for critical views (SummaryCards, PaymentTable) and Pinia store tests (authStore). Run: `npm run test`

---

## 10. Out of Scope

- `PUT /dashboard/v1/payment/{id}/review` — not in PDF requirements
- Dark mode
- Real-time payment updates (WebSocket)
- Role-based feature differences beyond login
