# Payment Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a full-stack payment dashboard — Go backend (SQLite, JWT, Chi) + Vue 3 TypeScript frontend (shadcn-vue, Pinia, @hey-api/openapi-ts) — packaged in Docker Compose.

**Architecture:** Go backend extends the existing oapi-codegen boilerplate with a new `payment` module following the existing handler/usecase/repository pattern. Frontend is a Vue 3 SPA with Pinia state management and an auto-generated OpenAPI client. Two Docker services: backend + frontend.

**Tech Stack:** Go 1.21+, Chi, oapi-codegen, SQLite (`database/sql`), Vue 3, TypeScript, Vite, shadcn-vue, Tailwind CSS, Pinia, @hey-api/openapi-ts, Vitest, Docker Compose

---

## File Map

### Backend (new/modified)
| File | Action | Purpose |
|---|---|---|
| `backend/internal/entity/payment.go` | Create | Payment domain struct |
| `backend/internal/module/payment/repository/payment.go` | Create | SQL queries for payments |
| `backend/internal/module/payment/usecase/payment.go` | Create | Business logic + interface |
| `backend/internal/module/payment/usecase/payment_test.go` | Create | Unit tests (mock repo) |
| `backend/internal/module/auth/usecase/auth_test.go` | Create | Unit tests (mock repo) |
| `backend/internal/module/payment/handler/payment.go` | Create | HTTP handler |
| `backend/internal/middleware/auth.go` | Create | JWT bearer middleware |
| `backend/internal/api/api_handler.go` | Modify | Wire payment handler |
| `backend/internal/service/http/server.go` | Modify | Add JWT middleware to payments route group |
| `backend/main.go` | Modify | Add payments table to initDB(), wire payment module |
| `backend/script/seed/main.go` | Create | Random data seeder |
| `backend/Makefile` | Modify | Add `seed` target |
| `backend/env.sample` | Modify | Ensure all vars documented |

### Frontend (new)
| File | Action | Purpose |
|---|---|---|
| `frontend/` | Create | Scaffolded via `npm create vue@latest` |
| `frontend/openapi-ts.config.ts` | Create | hey-api config |
| `frontend/src/api/generated/` | Generate | Auto-generated OpenAPI client |
| `frontend/src/stores/auth.ts` | Create | Pinia auth store |
| `frontend/src/stores/payments.ts` | Create | Pinia payments store |
| `frontend/src/router/index.ts` | Create | Vue Router with navigation guard |
| `frontend/src/pages/LoginPage.vue` | Create | Login UI |
| `frontend/src/pages/DashboardPage.vue` | Create | Dashboard layout |
| `frontend/src/components/AppSidebar.vue` | Create | Sidebar nav |
| `frontend/src/components/SummaryCards.vue` | Create | 4 metric cards |
| `frontend/src/components/PaymentTable.vue` | Create | Table + filters + pagination |
| `frontend/src/stores/auth.spec.ts` | Create | Auth store tests |
| `frontend/src/components/SummaryCards.spec.ts` | Create | SummaryCards tests |
| `frontend/src/components/PaymentTable.spec.ts` | Create | PaymentTable tests |
| `frontend/vite.config.ts` | Modify | Add proxy `/dashboard` → `http://localhost:8080` |
| `frontend/Dockerfile` | Create | nginx build |

### Root
| File | Action | Purpose |
|---|---|---|
| `docker-compose.yml` | Create | 2-service stack |
| `Makefile` | Create | Root dev + Docker targets |
| `README.md` | Modify | Full setup instructions |

---

## Task 1: Add payments table + Payment entity

**Files:**
- Modify: `backend/main.go`
- Create: `backend/internal/entity/payment.go`

- [ ] **Step 1: Create payment entity**

Create `backend/internal/entity/payment.go`:
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

- [ ] **Step 2: Add payments table to initDB()**

In `backend/main.go`, inside the `stmts` slice in `initDB()`, add after the users table statement:
```go
`CREATE TABLE IF NOT EXISTS payments (
  id         TEXT PRIMARY KEY,
  merchant   TEXT NOT NULL,
  amount     INTEGER NOT NULL,
  status     TEXT NOT NULL CHECK(status IN ('completed','processing','failed')),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,
```

- [ ] **Step 3: Build to verify no compile errors**

```bash
cd backend && CGO_ENABLED=1 go build ./...
```
Expected: no output (success).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/entity/payment.go backend/main.go
git commit -m "feat(backend): add payment entity and payments table schema"
```

---

## Task 2: Payment repository

**Files:**
- Create: `backend/internal/module/payment/repository/payment.go`

- [ ] **Step 1: Create payment repository**

Create `backend/internal/module/payment/repository/payment.go`:
```go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
)

type PaymentRepository interface {
	ListPayments(status, sort, id string) ([]*entity.Payment, error)
}

type Payment struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *Payment {
	return &Payment{db: db}
}

func (r *Payment) ListPayments(status, sort, id string) ([]*entity.Payment, error) {
	query := `SELECT id, merchant, amount, status, created_at FROM payments WHERE 1=1`
	args := []any{}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	if id != "" {
		query += ` AND id = ?`
		args = append(args, id)
	}

	orderCol := "created_at"
	orderDir := "DESC"
	switch sort {
	case "created_at":
		orderDir = "ASC"
	case "-amount":
		orderCol = "amount"
		orderDir = "DESC"
	case "amount":
		orderCol = "amount"
		orderDir = "ASC"
	}
	query += fmt.Sprintf(` ORDER BY %s %s`, orderCol, orderDir)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, entity.WrapError(err, entity.ErrorCodeInternal, "db error")
	}
	defer rows.Close()

	var payments []*entity.Payment
	for rows.Next() {
		var p entity.Payment
		var createdAt string
		if err := rows.Scan(&p.ID, &p.Merchant, &p.Amount, &p.Status, &createdAt); err != nil {
			return nil, entity.WrapError(err, entity.ErrorCodeInternal, "scan error")
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		payments = append(payments, &p)
	}
	return payments, nil
}
```

- [ ] **Step 2: Build to verify**

```bash
cd backend && CGO_ENABLED=1 go build ./...
```
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/module/payment/repository/payment.go
git commit -m "feat(backend): add payment repository with list/filter/sort"
```

---

## Task 3: Payment usecase + unit tests

**Files:**
- Create: `backend/internal/module/payment/usecase/payment.go`
- Create: `backend/internal/module/payment/usecase/payment_test.go`

- [ ] **Step 1: Write failing tests first**

Create `backend/internal/module/payment/usecase/payment_test.go`:
```go
package usecase_test

import (
	"errors"
	"testing"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
)

type mockPaymentRepo struct {
	payments []*entity.Payment
	err      error
	gotStatus, gotSort, gotID string
}

func (m *mockPaymentRepo) ListPayments(status, sort, id string) ([]*entity.Payment, error) {
	m.gotStatus = status
	m.gotSort = sort
	m.gotID = id
	return m.payments, m.err
}

func TestListPayments_All(t *testing.T) {
	repo := &mockPaymentRepo{
		payments: []*entity.Payment{
			{ID: "PAY-001", Merchant: "A", Amount: 100, Status: "completed"},
		},
	}
	uc := usecase.NewPaymentUsecase(repo)
	result, err := uc.ListPayments("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(result))
	}
}

func TestListPayments_FilterStatus(t *testing.T) {
	repo := &mockPaymentRepo{payments: []*entity.Payment{}}
	uc := usecase.NewPaymentUsecase(repo)
	_, _ = uc.ListPayments("completed", "", "")
	if repo.gotStatus != "completed" {
		t.Errorf("expected status=completed, got %q", repo.gotStatus)
	}
}

func TestListPayments_SortForwarded(t *testing.T) {
	repo := &mockPaymentRepo{payments: []*entity.Payment{}}
	uc := usecase.NewPaymentUsecase(repo)
	_, _ = uc.ListPayments("", "-amount", "")
	if repo.gotSort != "-amount" {
		t.Errorf("expected sort=-amount, got %q", repo.gotSort)
	}
}

func TestListPayments_RepoError(t *testing.T) {
	repo := &mockPaymentRepo{err: errors.New("db down")}
	uc := usecase.NewPaymentUsecase(repo)
	_, err := uc.ListPayments("", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
```

- [ ] **Step 2: Run tests — expect compile failure (usecase not yet created)**

```bash
cd backend && go test ./internal/module/payment/usecase/...
```
Expected: compile error "no required module provides package".

- [ ] **Step 3: Create payment usecase**

Create `backend/internal/module/payment/usecase/payment.go`:
```go
package usecase

import (
	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/payment/repository"
)

type PaymentUsecase interface {
	ListPayments(status, sort, id string) ([]*entity.Payment, error)
}

type Payment struct {
	repo repository.PaymentRepository
}

func NewPaymentUsecase(repo repository.PaymentRepository) *Payment {
	return &Payment{repo: repo}
}

func (u *Payment) ListPayments(status, sort, id string) ([]*entity.Payment, error) {
	return u.repo.ListPayments(status, sort, id)
}
```

- [ ] **Step 4: Run tests — expect all pass**

```bash
cd backend && go test ./internal/module/payment/usecase/... -v
```
Expected:
```
--- PASS: TestListPayments_All
--- PASS: TestListPayments_FilterStatus
--- PASS: TestListPayments_SortForwarded
--- PASS: TestListPayments_RepoError
PASS
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/module/payment/usecase/
git commit -m "feat(backend): add payment usecase with unit tests"
```

---

## Task 4: Auth usecase unit tests

**Files:**
- Create: `backend/internal/module/auth/usecase/auth_test.go`

- [ ] **Step 1: Create auth usecase tests**

Create `backend/internal/module/auth/usecase/auth_test.go`:
```go
package usecase_test

import (
	"testing"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	user *entity.User
	err  error
}

func (m *mockUserRepo) GetUserByEmail(email string) (*entity.User, error) {
	return m.user, m.err
}

func hashPassword(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}

func TestLogin_Success(t *testing.T) {
	repo := &mockUserRepo{user: &entity.User{
		ID: "1", Email: "cs@test.com",
		PasswordHash: hashPassword(t, "password"), Role: "cs",
	}}
	uc := usecase.NewAuthUsecase(repo, []byte("secret"), time.Hour)
	token, user, err := uc.Login("cs@test.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if user.Email != "cs@test.com" {
		t.Errorf("expected email cs@test.com, got %s", user.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockUserRepo{user: &entity.User{
		ID: "1", Email: "cs@test.com",
		PasswordHash: hashPassword(t, "password"), Role: "cs",
	}}
	uc := usecase.NewAuthUsecase(repo, []byte("secret"), time.Hour)
	_, _, err := uc.Login("cs@test.com", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{err: entity.ErrorNotFound("user not found")}
	uc := usecase.NewAuthUsecase(repo, []byte("secret"), time.Hour)
	_, _, err := uc.Login("nobody@test.com", "password")
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd backend && go test ./internal/module/auth/usecase/... -v
```
Expected: all 3 tests PASS.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/module/auth/usecase/auth_test.go
git commit -m "test(backend): add auth usecase unit tests"
```

---

## Task 5: JWT middleware

**Files:**
- Create: `backend/internal/middleware/auth.go`

- [ ] **Step 1: Create JWT middleware**

Create `backend/internal/middleware/auth.go`:
```go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func JWTAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				transport.WriteAppError(w, entity.ErrorUnauthorized("missing token"))
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, entity.ErrorUnauthorized("invalid signing method")
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				transport.WriteAppError(w, entity.ErrorUnauthorized("invalid token"))
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, token.Claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

- [ ] **Step 2: Update `transport.CodeToStatus` to handle Unauthorized**

In `backend/internal/transport/jsonerror.go`, update `CodeToStatus`:
```go
func CodeToStatus(code entity.Code) int {
	switch code {
	case entity.ErrorCodeBadRequest:
		return http.StatusBadRequest
	case entity.ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case entity.ErrorCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
```

- [ ] **Step 3: Build to verify**

```bash
cd backend && CGO_ENABLED=1 go build ./...
```
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/middleware/auth.go backend/internal/transport/jsonerror.go
git commit -m "feat(backend): add JWT auth middleware and fix CodeToStatus"
```

---

## Task 6: Payment HTTP handler + wire everything

**Files:**
- Create: `backend/internal/module/payment/handler/payment.go`
- Modify: `backend/internal/api/api_handler.go`
- Modify: `backend/internal/service/http/server.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Create payment handler**

Create `backend/internal/module/payment/handler/payment.go`:
```go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	paymentUsecase "github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

type PaymentHandler struct {
	paymentUC paymentUsecase.PaymentUsecase
}

func NewPaymentHandler(paymentUC paymentUsecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{paymentUC: paymentUC}
}

func (h *PaymentHandler) GetDashboardV1Payments(w http.ResponseWriter, r *http.Request, params openapigen.GetDashboardV1PaymentsParams) {
	var status, sort, id string
	if params.Status != nil {
		status = *params.Status
	}
	if params.Sort != nil {
		sort = *params.Sort
	}
	if params.Id != nil {
		id = *params.Id
	}

	payments, err := h.paymentUC.ListPayments(status, sort, id)
	if err != nil {
		transport.WriteError(w, err)
		return
	}

	type response struct {
		Payments any `json:"payments"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response{Payments: payments}); err != nil {
		transport.WriteAppError(w, entity.ErrorInternal("encode error"))
	}
}
```

- [ ] **Step 2: Wire payment handler into APIHandler**

Replace `backend/internal/api/api_handler.go`:
```go
package api

import (
	"net/http"

	ah "github.com/durianpay/fullstack-boilerplate/internal/module/auth/handler"
	ph "github.com/durianpay/fullstack-boilerplate/internal/module/payment/handler"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
)

type APIHandler struct {
	Auth    *ah.AuthHandler
	Payment *ph.PaymentHandler
}

var _ openapigen.ServerInterface = (*APIHandler)(nil)

func (h *APIHandler) PostDashboardV1AuthLogin(w http.ResponseWriter, r *http.Request) {
	h.Auth.PostDashboardV1AuthLogin(w, r)
}

func (h *APIHandler) GetDashboardV1Payments(w http.ResponseWriter, r *http.Request, params openapigen.GetDashboardV1PaymentsParams) {
	h.Payment.GetDashboardV1Payments(w, r, params)
}
```

- [ ] **Step 3: Add JWT middleware to server payments route group**

In `backend/internal/service/http/server.go`, update `NewServer` to add a protected sub-router:
```go
package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/config"
	mw "github.com/durianpay/fullstack-boilerplate/internal/middleware"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/go-chi/chi/v5"
	oapinethttpmw "github.com/oapi-codegen/nethttp-middleware"
)

type Server struct {
	router http.Handler
}

const (
	readTimeout  = 10
	writeTimeout = 10
	idleTimeout  = 60
)

func NewServer(apiHandler openapigen.ServerInterface, openapiYamlPath string) *Server {
	swagger, err := openapigen.GetSwagger()
	if err != nil {
		log.Fatalf("failed to load swagger: %v", err)
	}

	r := chi.NewRouter()

	r.Post("/dashboard/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		apiHandler.PostDashboardV1AuthLogin(w, r)
	})

	r.Group(func(protected chi.Router) {
		protected.Use(mw.JWTAuth(config.JwtSecret))
		protected.Use(oapinethttpmw.OapiRequestValidatorWithOptions(
			swagger,
			&oapinethttpmw.Options{
				DoNotValidateServers:  true,
				SilenceServersWarning: true,
			},
		))
		openapigen.HandlerFromMux(apiHandler, protected)
	})

	return &Server{router: r}
}

func (s *Server) Start(addr string) {
	service := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  readTimeout * time.Second,
		WriteTimeout: writeTimeout * time.Second,
		IdleTimeout:  idleTimeout * time.Second,
	}
	go func() {
		log.Printf("listening on %s", addr)
		if err := service.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down gracefully...")

	const shutdownTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := service.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("Server stopped cleanly")
}

func (s *Server) Routes() http.Handler {
	return s.router
}
```

- [ ] **Step 4: Wire payment module in main.go**

In `backend/main.go`, add payment imports and wiring after `authH`:
```go
import (
    // existing imports...
    pr "github.com/durianpay/fullstack-boilerplate/internal/module/payment/repository"
    pu "github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
    ph "github.com/durianpay/fullstack-boilerplate/internal/module/payment/handler"
)

// inside main(), after authH:
paymentRepo := pr.NewPaymentRepo(db)
paymentUC   := pu.NewPaymentUsecase(paymentRepo)
paymentH    := ph.NewPaymentHandler(paymentUC)

apiHandler := &api.APIHandler{
    Auth:    authH,
    Payment: paymentH,
}
```

- [ ] **Step 5: Build and run to verify**

```bash
cd backend && cp env.sample .env && CGO_ENABLED=1 go run main.go &
sleep 2
curl -s -X POST http://localhost:8080/dashboard/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"cs@test.com","password":"password"}' | grep token
kill %1
```
Expected: JSON response containing `"token":"..."`.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/module/payment/handler/ \
        backend/internal/api/api_handler.go \
        backend/internal/service/http/server.go \
        backend/main.go
git commit -m "feat(backend): wire payment handler with JWT-protected route"
```

---

## Task 7: Seed script

**Files:**
- Create: `backend/script/seed/main.go`
- Modify: `backend/Makefile`

- [ ] **Step 1: Create seed script**

Create `backend/script/seed/main.go`:
```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var merchants = []string{
	"Tokopedia", "Shopee", "Bukalapak", "Lazada", "Blibli",
	"JD.id", "Zalora", "Sociolla", "Alfamart", "Indomaret",
	"Gojek", "Grab", "OVO Store", "Dana Merchant", "LinkAja Shop",
}

var statuses = []string{
	"completed", "completed", "completed", "completed", "completed", "completed",
	"processing", "processing", "processing",
	"failed", "failed",
}

func main() {
	_ = godotenv.Load()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dashboard.db"
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := seedUsers(db); err != nil {
		log.Fatalf("seed users: %v", err)
	}
	if err := seedPayments(db); err != nil {
		log.Fatalf("seed payments: %v", err)
	}
	log.Println("Seed complete.")
}

func seedUsers(db *sql.DB) error {
	var cnt int
	if err := db.QueryRow("SELECT COUNT(1) FROM users").Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		log.Println("users already seeded, skipping")
		return nil
	}
	users := []struct{ email, password, role string }{
		{"cs@test.com", "password", "cs"},
		{"operation@test.com", "password", "operation"},
	}
	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := db.Exec(
			"INSERT INTO users(email, password_hash, role) VALUES (?, ?, ?)",
			u.email, string(hash), u.role,
		); err != nil {
			return err
		}
	}
	log.Printf("seeded %d users", len(users))
	return nil
}

func seedPayments(db *sql.DB) error {
	var cnt int
	if err := db.QueryRow("SELECT COUNT(1) FROM payments").Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		log.Println("payments already seeded, skipping")
		return nil
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	for i := 1; i <= 50; i++ {
		id := fmt.Sprintf("PAY-%03d", i)
		merchant := merchants[rng.Intn(len(merchants))]
		amount := int64((rng.Intn(490)+10)*1000)
		status := statuses[rng.Intn(len(statuses))]
		createdAt := now.Add(-time.Duration(rng.Intn(90*24)) * time.Hour)

		if _, err := db.Exec(
			"INSERT INTO payments(id, merchant, amount, status, created_at) VALUES (?, ?, ?, ?, ?)",
			id, merchant, amount, status, createdAt.Format("2006-01-02 15:04:05"),
		); err != nil {
			return err
		}
	}
	log.Println("seeded 50 payments")
	return nil
}
```

- [ ] **Step 2: Add seed target to Makefile**

In `backend/Makefile`, add after `gen-secret`:
```makefile
seed:
	@echo "Seeding database..."
	CGO_ENABLED=1 $(go_bin) run ./script/seed/main.go
```

- [ ] **Step 3: Run seed and verify**

```bash
cd backend && make seed
sqlite3 dashboard.db "SELECT COUNT(*) FROM payments; SELECT COUNT(*) FROM users;"
```
Expected:
```
50
2
```

- [ ] **Step 4: Verify idempotent (run again, should skip)**

```bash
cd backend && make seed
```
Expected: `payments already seeded, skipping` and `users already seeded, skipping`.

- [ ] **Step 5: Commit**

```bash
git add backend/script/seed/main.go backend/Makefile
git commit -m "feat(backend): add random payment seed script"
```

---

## Task 8: Verify full backend E2E

- [ ] **Step 1: Run all backend tests**

```bash
cd backend && go test ./... -v
```
Expected: all tests PASS, no failures.

- [ ] **Step 2: Start server and test payments endpoint**

```bash
cd backend && CGO_ENABLED=1 go run main.go &
sleep 2

# Login and capture token
TOKEN=$(curl -s -X POST http://localhost:8080/dashboard/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"cs@test.com","password":"password"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# List all payments
curl -s http://localhost:8080/dashboard/v1/payments \
  -H "Authorization: Bearer $TOKEN" | grep -o '"id":"PAY' | wc -l

# Filter by status
curl -s "http://localhost:8080/dashboard/v1/payments?status=completed" \
  -H "Authorization: Bearer $TOKEN" | grep -o '"status":"completed"' | wc -l

# Test without token (expect 401)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/dashboard/v1/payments

kill %1
```
Expected: 50 payments total, all filtered have status=completed, 401 without token.

- [ ] **Step 3: Commit**

```bash
git commit --allow-empty -m "chore: backend E2E verified"
```

---

## Task 9: Scaffold Vue 3 frontend

**Files:**
- Create: `frontend/` (scaffolded)
- Modify: `frontend/vite.config.ts`
- Create: `frontend/openapi-ts.config.ts`
- Create: `frontend/src/api/generated/`

- [ ] **Step 1: Scaffold Vue 3 + TypeScript project**

```bash
cd "C:/Users/USER/projects/fullstack-boilerplate-durianpay"
npm create vue@latest frontend -- --typescript --router --pinia --vitest --eslint-with-prettier
cd frontend && npm install
```
When prompted, select: TypeScript ✓, Vue Router ✓, Pinia ✓, Vitest ✓, ESLint ✓.

- [ ] **Step 2: Install UI + API deps**

```bash
cd frontend
npm install tailwindcss @tailwindcss/vite
npm install @hey-api/openapi-ts @hey-api/client-fetch
npm install pinia-plugin-persistedstate
npm install -D @vitejs/plugin-vue
npx shadcn-vue@latest init
```
When shadcn-vue asks: style=Default, base color=Slate, CSS variables=Yes.

- [ ] **Step 3: Install shadcn components needed**

```bash
cd frontend
npx shadcn-vue@latest add button input label badge select table card
```

- [ ] **Step 4: Configure Tailwind**

In `frontend/vite.config.ts`, add Tailwind plugin and dev proxy:
```ts
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    proxy: {
      '/dashboard': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 5: Add Tailwind to main CSS**

Replace contents of `frontend/src/assets/main.css`:
```css
@import "tailwindcss";
```

- [ ] **Step 6: Create OpenAPI client config**

Create `frontend/openapi-ts.config.ts`:
```ts
import { defineConfig } from '@hey-api/openapi-ts'

export default defineConfig({
  input: '../openapi.yaml',
  output: {
    path: 'src/api/generated',
    format: 'prettier',
  },
  plugins: ['@hey-api/client-fetch'],
})
```

Add script to `frontend/package.json` scripts:
```json
"gen:api": "openapi-ts"
```

- [ ] **Step 7: Generate API client**

```bash
cd frontend && npm run gen:api
```
Expected: `src/api/generated/` created with `types.gen.ts`, `services.gen.ts`, `client.ts`.

- [ ] **Step 8: Verify dev server starts**

```bash
cd frontend && npm run dev
```
Expected: Vite starts on `http://localhost:5173` without errors. Kill with Ctrl+C.

- [ ] **Step 9: Commit**

```bash
git add frontend/
git commit -m "feat(frontend): scaffold Vue 3 + TS + shadcn-vue + OpenAPI client"
```

---

## Task 10: Auth Pinia store

**Files:**
- Create: `frontend/src/stores/auth.ts`
- Modify: `frontend/src/main.ts`

- [ ] **Step 1: Write failing auth store test**

Create `frontend/src/stores/auth.spec.ts`:
```ts
import { setActivePinia, createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useAuthStore } from './auth'

vi.mock('@/api/generated', () => ({
  postDashboardV1AuthLogin: vi.fn(),
}))

import { postDashboardV1AuthLogin } from '@/api/generated'

describe('authStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('login stores token and role on success', async () => {
    vi.mocked(postDashboardV1AuthLogin).mockResolvedValueOnce({
      data: { token: 'jwt-token-123', role: 'cs', email: 'cs@test.com' },
      error: undefined,
    } as any)

    const store = useAuthStore()
    await store.login('cs@test.com', 'password')

    expect(store.token).toBe('jwt-token-123')
    expect(store.role).toBe('cs')
  })

  it('logout clears token and role', async () => {
    const store = useAuthStore()
    store.token = 'some-token'
    store.role = 'cs'
    store.logout()

    expect(store.token).toBeNull()
    expect(store.role).toBeNull()
  })

  it('login throws on API error', async () => {
    vi.mocked(postDashboardV1AuthLogin).mockResolvedValueOnce({
      data: undefined,
      error: { message: 'invalid credentials' },
    } as any)

    const store = useAuthStore()
    await expect(store.login('bad@test.com', 'wrong')).rejects.toThrow()
  })
})
```

- [ ] **Step 2: Run test — expect failure**

```bash
cd frontend && npm run test -- --run src/stores/auth.spec.ts
```
Expected: FAIL — `useAuthStore` not found.

- [ ] **Step 3: Create auth store**

Create `frontend/src/stores/auth.ts`:
```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { postDashboardV1AuthLogin } from '@/api/generated'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const role = ref<string | null>(null)
  const email = ref<string | null>(null)

  async function login(emailInput: string, password: string) {
    const { data, error } = await postDashboardV1AuthLogin({
      body: { email: emailInput, password },
    })
    if (error || !data?.token) {
      throw new Error((error as any)?.message ?? 'Login failed')
    }
    token.value = data.token
    role.value = data.role ?? null
    email.value = data.email ?? emailInput
  }

  function logout() {
    token.value = null
    role.value = null
    email.value = null
  }

  return { token, role, email, login, logout }
}, {
  persist: true,
})
```

- [ ] **Step 4: Wire persistedstate plugin in main.ts**

In `frontend/src/main.ts`:
```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import App from './App.vue'
import router from './router'
import './assets/main.css'

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

const app = createApp(App)
app.use(pinia)
app.use(router)
app.mount('#app')
```

- [ ] **Step 5: Run tests — expect all pass**

```bash
cd frontend && npm run test -- --run src/stores/auth.spec.ts
```
Expected: 3 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/stores/auth.ts frontend/src/stores/auth.spec.ts frontend/src/main.ts
git commit -m "feat(frontend): add auth Pinia store with persistence and tests"
```

---

## Task 11: Payments Pinia store

**Files:**
- Create: `frontend/src/stores/payments.ts`

- [ ] **Step 1: Create payments store**

Create `frontend/src/stores/payments.ts`:
```ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getDashboardV1Payments } from '@/api/generated'
import { useAuthStore } from './auth'

export interface Payment {
  id: string
  merchant: string
  amount: number
  status: 'completed' | 'processing' | 'failed'
  created_at: string
}

export const usePaymentsStore = defineStore('payments', () => {
  const payments = ref<Payment[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filterStatus = ref('')
  const filterSort = ref('-created_at')
  const filterSearch = ref('')

  const summary = computed(() => ({
    total: payments.value.length,
    completed: payments.value.filter(p => p.status === 'completed').length,
    processing: payments.value.filter(p => p.status === 'processing').length,
    failed: payments.value.filter(p => p.status === 'failed').length,
  }))

  const filteredPayments = computed(() => {
    const search = filterSearch.value.toLowerCase()
    if (!search) return payments.value
    return payments.value.filter(
      p => p.id.toLowerCase().includes(search) || p.merchant.toLowerCase().includes(search)
    )
  })

  async function fetchPayments() {
    const authStore = useAuthStore()
    loading.value = true
    error.value = null
    try {
      const { data, error: apiError } = await getDashboardV1Payments({
        query: {
          status: filterStatus.value || undefined,
          sort: filterSort.value || undefined,
        },
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (apiError) throw new Error('Failed to fetch payments')
      payments.value = ((data as any)?.payments ?? []) as Payment[]
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  return { payments, loading, error, filterStatus, filterSort, filterSearch, summary, filteredPayments, fetchPayments }
})
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/stores/payments.ts
git commit -m "feat(frontend): add payments Pinia store"
```

---

## Task 12: Vue Router with navigation guard

**Files:**
- Modify: `frontend/src/router/index.ts`

- [ ] **Step 1: Replace router with protected routes**

Replace `frontend/src/router/index.ts`:
```ts
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/pages/DashboardPage.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.token) {
    return { name: 'login' }
  }
  if (to.name === 'login' && auth.token) {
    return { name: 'dashboard' }
  }
})

export default router
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/router/index.ts
git commit -m "feat(frontend): add Vue Router with auth navigation guard"
```

---

## Task 13: AppSidebar component

**Files:**
- Create: `frontend/src/components/AppSidebar.vue`

- [ ] **Step 1: Create sidebar component**

Create `frontend/src/components/AppSidebar.vue`:
```vue
<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentsStore } from '@/stores/payments'

const router = useRouter()
const authStore = useAuthStore()
const paymentsStore = usePaymentsStore()

function logout() {
  authStore.logout()
  router.push({ name: 'login' })
}

const initials = authStore.role?.slice(0, 2).toUpperCase() ?? 'U'
</script>

<template>
  <aside class="w-[220px] flex-shrink-0 bg-white border-r border-slate-200 flex flex-col h-screen">
    <!-- Brand -->
    <div class="p-5 border-b border-slate-100 flex items-center gap-3">
      <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white font-bold text-sm">
        P
      </div>
      <div>
        <div class="font-bold text-sm text-slate-900">PayDash</div>
        <div class="text-[11px] text-slate-400">Internal Dashboard</div>
      </div>
    </div>

    <!-- Nav -->
    <nav class="flex-1 p-3 flex flex-col gap-1">
      <div class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider px-3 mb-1 mt-2">Main</div>
      <router-link
        to="/dashboard"
        class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold text-indigo-600 bg-indigo-50"
      >
        <span>📊</span>
        <span>Payments</span>
        <span class="ml-auto bg-indigo-600 text-white text-[10px] font-bold px-2 py-0.5 rounded-full">
          {{ paymentsStore.summary.total }}
        </span>
      </router-link>
    </nav>

    <!-- User -->
    <div class="p-4 border-t border-slate-100">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-cyan-500 flex items-center justify-center text-white text-[11px] font-bold flex-shrink-0">
          {{ initials }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-xs font-bold text-slate-900 truncate">{{ authStore.email }}</div>
          <div class="text-[10px] text-slate-400 capitalize">{{ authStore.role }} role</div>
        </div>
        <button @click="logout" class="text-slate-400 hover:text-slate-600 text-sm">↩</button>
      </div>
    </div>
  </aside>
</template>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/AppSidebar.vue
git commit -m "feat(frontend): add AppSidebar component"
```

---

## Task 14: SummaryCards component + tests

**Files:**
- Create: `frontend/src/components/SummaryCards.vue`
- Create: `frontend/src/components/SummaryCards.spec.ts`

- [ ] **Step 1: Write failing test**

Create `frontend/src/components/SummaryCards.spec.ts`:
```ts
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import SummaryCards from './SummaryCards.vue'

describe('SummaryCards', () => {
  const summary = { total: 50, completed: 30, processing: 12, failed: 8 }

  it('renders all four counts', () => {
    const wrapper = mount(SummaryCards, { props: { summary } })
    expect(wrapper.text()).toContain('50')
    expect(wrapper.text()).toContain('30')
    expect(wrapper.text()).toContain('12')
    expect(wrapper.text()).toContain('8')
  })

  it('renders correct labels', () => {
    const wrapper = mount(SummaryCards, { props: { summary } })
    expect(wrapper.text()).toContain('Total Payments')
    expect(wrapper.text()).toContain('Completed')
    expect(wrapper.text()).toContain('Processing')
    expect(wrapper.text()).toContain('Failed')
  })
})
```

- [ ] **Step 2: Run test — expect failure**

```bash
cd frontend && npm run test -- --run src/components/SummaryCards.spec.ts
```
Expected: FAIL — component not found.

- [ ] **Step 3: Create SummaryCards component**

Create `frontend/src/components/SummaryCards.vue`:
```vue
<script setup lang="ts">
defineProps<{
  summary: {
    total: number
    completed: number
    processing: number
    failed: number
  }
}>()

const cards = [
  { key: 'total',      label: 'Total Payments', icon: '💳', bg: 'bg-violet-50',  color: 'text-slate-900' },
  { key: 'completed',  label: 'Completed',       icon: '✅', bg: 'bg-emerald-50', color: 'text-emerald-600' },
  { key: 'processing', label: 'Processing',      icon: '⏳', bg: 'bg-amber-50',   color: 'text-amber-600' },
  { key: 'failed',     label: 'Failed',          icon: '❌', bg: 'bg-red-50',     color: 'text-red-600' },
] as const
</script>

<template>
  <div class="grid grid-cols-4 gap-4">
    <div
      v-for="card in cards"
      :key="card.key"
      class="bg-white rounded-xl border border-slate-200 p-4 flex items-start gap-3"
    >
      <div :class="[card.bg, 'w-10 h-10 rounded-xl flex items-center justify-center text-lg flex-shrink-0']">
        {{ card.icon }}
      </div>
      <div>
        <div :class="[card.color, 'text-2xl font-bold leading-none']">{{ summary[card.key] }}</div>
        <div class="text-xs text-slate-400 mt-1">{{ card.label }}</div>
      </div>
    </div>
  </div>
</template>
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd frontend && npm run test -- --run src/components/SummaryCards.spec.ts
```
Expected: 2 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/SummaryCards.vue frontend/src/components/SummaryCards.spec.ts
git commit -m "feat(frontend): add SummaryCards component with tests"
```

---

## Task 15: PaymentTable component + tests

**Files:**
- Create: `frontend/src/components/PaymentTable.vue`
- Create: `frontend/src/components/PaymentTable.spec.ts`

- [ ] **Step 1: Write failing test**

Create `frontend/src/components/PaymentTable.spec.ts`:
```ts
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PaymentTable from './PaymentTable.vue'
import type { Payment } from '@/stores/payments'

const mockPayments: Payment[] = [
  { id: 'PAY-001', merchant: 'Tokopedia', amount: 150000, status: 'completed', created_at: '2025-05-01 10:00:00' },
  { id: 'PAY-002', merchant: 'Shopee',    amount: 75000,  status: 'processing', created_at: '2025-05-02 11:00:00' },
  { id: 'PAY-003', merchant: 'Gojek',     amount: 50000,  status: 'failed',    created_at: '2025-05-03 12:00:00' },
]

describe('PaymentTable', () => {
  it('renders all payment rows', () => {
    const wrapper = mount(PaymentTable, {
      props: { payments: mockPayments, loading: false },
    })
    expect(wrapper.text()).toContain('PAY-001')
    expect(wrapper.text()).toContain('PAY-002')
    expect(wrapper.text()).toContain('PAY-003')
  })

  it('renders merchant names', () => {
    const wrapper = mount(PaymentTable, {
      props: { payments: mockPayments, loading: false },
    })
    expect(wrapper.text()).toContain('Tokopedia')
    expect(wrapper.text()).toContain('Shopee')
  })

  it('shows loading state', () => {
    const wrapper = mount(PaymentTable, {
      props: { payments: [], loading: true },
    })
    expect(wrapper.text()).toContain('Loading')
  })
})
```

- [ ] **Step 2: Run test — expect failure**

```bash
cd frontend && npm run test -- --run src/components/PaymentTable.spec.ts
```
Expected: FAIL.

- [ ] **Step 3: Create PaymentTable component**

Create `frontend/src/components/PaymentTable.vue`:
```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Payment } from '@/stores/payments'

const props = defineProps<{
  payments: Payment[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'update:filterStatus', val: string): void
  (e: 'update:filterSort', val: string): void
  (e: 'update:filterSearch', val: string): void
}>()

const search = ref('')
const statusFilter = ref('')
const sortFilter = ref('-created_at')

const PAGE_SIZE = 10
const currentPage = ref(1)

const displayed = computed(() => {
  const s = search.value.toLowerCase()
  let list = props.payments
  if (s) list = list.filter(p => p.id.toLowerCase().includes(s) || p.merchant.toLowerCase().includes(s))
  const start = (currentPage.value - 1) * PAGE_SIZE
  return list.slice(start, start + PAGE_SIZE)
})

const totalPages = computed(() => {
  const s = search.value.toLowerCase()
  let list = props.payments
  if (s) list = list.filter(p => p.id.toLowerCase().includes(s) || p.merchant.toLowerCase().includes(s))
  return Math.max(1, Math.ceil(list.length / PAGE_SIZE))
})

function onStatusChange(val: string) {
  statusFilter.value = val
  currentPage.value = 1
  emit('update:filterStatus', val)
}

function onSortChange(val: string) {
  sortFilter.value = val
  emit('update:filterSort', val)
}

function onSearch(val: string) {
  search.value = val
  currentPage.value = 1
  emit('update:filterSearch', val)
}

const statusBadge: Record<string, string> = {
  completed: 'bg-emerald-100 text-emerald-700',
  processing: 'bg-amber-100 text-amber-700',
  failed: 'bg-red-100 text-red-700',
}

function formatAmount(amount: number) {
  return `Rp ${amount.toLocaleString('id-ID')}`
}

function formatDate(dt: string) {
  return new Date(dt).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<template>
  <div class="bg-white rounded-xl border border-slate-200 overflow-hidden">
    <!-- Header + Filters -->
    <div class="px-5 py-4 border-b border-slate-100 flex items-center justify-between gap-4">
      <h2 class="text-sm font-bold text-slate-900">All Transactions</h2>
      <div class="flex gap-2 items-center">
        <input
          :value="search"
          @input="onSearch(($event.target as HTMLInputElement).value)"
          placeholder="Search by ID or merchant..."
          class="border border-slate-200 rounded-lg px-3 py-1.5 text-xs w-48 outline-none focus:border-indigo-400"
          style="font-family: Arial, sans-serif"
        />
        <select
          :value="statusFilter"
          @change="onStatusChange(($event.target as HTMLSelectElement).value)"
          class="border border-slate-200 rounded-lg px-2 py-1.5 text-xs outline-none appearance-none pr-6"
          style="font-family: Arial, sans-serif"
        >
          <option value="">All Status</option>
          <option value="completed">Completed</option>
          <option value="processing">Processing</option>
          <option value="failed">Failed</option>
        </select>
        <select
          :value="sortFilter"
          @change="onSortChange(($event.target as HTMLSelectElement).value)"
          class="border border-slate-200 rounded-lg px-2 py-1.5 text-xs outline-none appearance-none pr-6"
          style="font-family: Arial, sans-serif"
        >
          <option value="-created_at">Newest First</option>
          <option value="created_at">Oldest First</option>
          <option value="-amount">Amount ↓</option>
          <option value="amount">Amount ↑</option>
        </select>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="py-16 text-center text-sm text-slate-400">Loading...</div>

    <!-- Table -->
    <table v-else class="w-full border-collapse">
      <thead>
        <tr class="bg-slate-50 border-b border-slate-100">
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Payment ID</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Merchant</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Date</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Amount</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in displayed" :key="p.id" class="border-b border-slate-50 hover:bg-slate-50/50">
          <td class="px-5 py-3 text-xs text-indigo-600 font-bold">{{ p.id }}</td>
          <td class="px-5 py-3 text-xs font-semibold text-slate-900">{{ p.merchant }}</td>
          <td class="px-5 py-3 text-xs text-slate-500">{{ formatDate(p.created_at) }}</td>
          <td class="px-5 py-3 text-xs font-bold text-slate-900">{{ formatAmount(p.amount) }}</td>
          <td class="px-5 py-3">
            <span :class="[statusBadge[p.status], 'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-bold']">
              <span class="w-1.5 h-1.5 rounded-full bg-current"></span>
              {{ p.status }}
            </span>
          </td>
        </tr>
        <tr v-if="displayed.length === 0">
          <td colspan="5" class="px-5 py-12 text-center text-sm text-slate-400">No payments found</td>
        </tr>
      </tbody>
    </table>

    <!-- Pagination -->
    <div class="px-5 py-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-400">
      <span>Showing {{ displayed.length }} of {{ payments.length }} payments</span>
      <div class="flex gap-1">
        <button
          v-for="page in totalPages"
          :key="page"
          @click="currentPage = page"
          :class="['w-7 h-7 rounded-lg border text-xs flex items-center justify-center',
            page === currentPage
              ? 'bg-indigo-600 text-white border-indigo-600'
              : 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50']"
        >{{ page }}</button>
      </div>
    </div>
  </div>
</template>
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd frontend && npm run test -- --run src/components/PaymentTable.spec.ts
```
Expected: 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/PaymentTable.vue frontend/src/components/PaymentTable.spec.ts
git commit -m "feat(frontend): add PaymentTable component with filters and tests"
```

---

## Task 16: LoginPage

**Files:**
- Create: `frontend/src/pages/LoginPage.vue`

- [ ] **Step 1: Create login page**

Create `frontend/src/pages/LoginPage.vue`:
```vue
<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(email.value, password.value)
    router.push({ name: 'dashboard' })
  } catch {
    error.value = 'Invalid email or password. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex h-screen" style="font-family: Arial, sans-serif">
    <!-- Left panel -->
    <div class="w-[45%] flex-shrink-0 bg-gradient-to-br from-indigo-500 via-purple-500 to-purple-400 flex flex-col justify-center items-start p-12 relative overflow-hidden">
      <div class="absolute -top-16 -right-16 w-72 h-72 rounded-full bg-white/[0.07]"></div>
      <div class="absolute -bottom-20 -left-10 w-80 h-80 rounded-full bg-white/[0.05]"></div>
      <div class="relative z-10">
        <div class="w-12 h-12 bg-white rounded-xl flex items-center justify-center text-indigo-600 font-black text-xl mb-8">P</div>
        <h1 class="text-3xl font-bold text-white leading-snug mb-4">Payment<br>Monitor Dashboard</h1>
        <p class="text-white/70 text-sm leading-relaxed max-w-xs">
          Internal tool for monitoring and tracking all incoming payment transactions in real time.
        </p>
        <div class="flex gap-8 mt-10">
          <div>
            <div class="text-white text-2xl font-bold">50+</div>
            <div class="text-white/60 text-xs mt-0.5">Transactions</div>
          </div>
          <div>
            <div class="text-white text-2xl font-bold">2</div>
            <div class="text-white/60 text-xs mt-0.5">User Roles</div>
          </div>
          <div>
            <div class="text-white text-2xl font-bold">99%</div>
            <div class="text-white/60 text-xs mt-0.5">Uptime</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Right form -->
    <div class="flex-1 flex items-center justify-center bg-slate-50 p-8">
      <div class="bg-white rounded-2xl border border-slate-200 p-10 w-full max-w-sm">
        <h2 class="text-xl font-bold text-slate-900 mb-1">Welcome back</h2>
        <p class="text-slate-400 text-sm mb-8">Sign in to your account to continue</p>

        <!-- Error -->
        <div v-if="error" class="mb-4 bg-red-50 border border-red-200 rounded-lg px-4 py-3 text-xs text-red-600 flex items-center gap-2">
          ⚠ {{ error }}
        </div>

        <form @submit.prevent="handleLogin" class="flex flex-col gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Email address</label>
            <input
              v-model="email"
              type="email"
              placeholder="cs@test.com"
              required
              class="w-full border border-slate-200 rounded-lg px-3 py-2.5 text-sm outline-none focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              style="font-family: Arial, sans-serif"
            />
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Password</label>
            <div class="relative">
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="••••••••"
                required
                class="w-full border border-slate-200 rounded-lg px-3 py-2.5 text-sm outline-none focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 pr-10"
                style="font-family: Arial, sans-serif"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 text-sm"
              >{{ showPassword ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-bold py-2.5 rounded-lg mt-1 disabled:opacity-50"
            style="font-family: Arial, sans-serif"
          >
            {{ loading ? 'Signing in...' : 'Sign in →' }}
          </button>
        </form>

        <div class="flex items-center gap-3 my-5">
          <hr class="flex-1 border-slate-100" /><span class="text-[11px] text-slate-300">test accounts</span><hr class="flex-1 border-slate-100" />
        </div>
        <div class="bg-slate-50 border border-slate-200 rounded-lg px-4 py-3 text-[11px] text-slate-500 leading-relaxed">
          <strong class="text-slate-700">cs@test.com</strong> / password — CS role<br>
          <strong class="text-slate-700">operation@test.com</strong> / password — Operation role
        </div>

        <div class="mt-6 text-center text-[11px] text-slate-400">DurianPay Internal</div>
      </div>
    </div>
  </div>
</template>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/LoginPage.vue
git commit -m "feat(frontend): add LoginPage with split layout"
```

---

## Task 17: DashboardPage + App.vue cleanup

**Files:**
- Create: `frontend/src/pages/DashboardPage.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Create dashboard page**

Create `frontend/src/pages/DashboardPage.vue`:
```vue
<script setup lang="ts">
import { onMounted, watch } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'
import SummaryCards from '@/components/SummaryCards.vue'
import PaymentTable from '@/components/PaymentTable.vue'
import { usePaymentsStore } from '@/stores/payments'

const paymentsStore = usePaymentsStore()

onMounted(() => paymentsStore.fetchPayments())

watch([() => paymentsStore.filterStatus, () => paymentsStore.filterSort], () => {
  paymentsStore.fetchPayments()
})

const today = new Date().toLocaleDateString('en-US', { weekday: 'short', day: 'numeric', month: 'long', year: 'numeric' })
</script>

<template>
  <div class="flex h-screen bg-slate-50" style="font-family: Arial, sans-serif">
    <AppSidebar />

    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Topbar -->
      <div class="bg-white border-b border-slate-200 px-7 py-4 flex items-center justify-between flex-shrink-0">
        <div>
          <h1 class="text-base font-bold text-slate-900">Payment Monitor</h1>
          <div class="text-xs text-slate-400 mt-0.5">Showing all incoming transactions</div>
        </div>
        <div class="text-xs text-slate-400">{{ today }}</div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-7 flex flex-col gap-5">
        <SummaryCards :summary="paymentsStore.summary" />

        <PaymentTable
          :payments="paymentsStore.filteredPayments"
          :loading="paymentsStore.loading"
          v-model:filterStatus="paymentsStore.filterStatus"
          v-model:filterSort="paymentsStore.filterSort"
          v-model:filterSearch="paymentsStore.filterSearch"
        />
      </div>
    </div>
  </div>
</template>
```

- [ ] **Step 2: Simplify App.vue to just router-view**

Replace `frontend/src/App.vue`:
```vue
<template>
  <router-view />
</template>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/DashboardPage.vue frontend/src/App.vue
git commit -m "feat(frontend): add DashboardPage with sidebar, summary, and table"
```

---

## Task 18: Run all frontend tests

- [ ] **Step 1: Run all frontend tests**

```bash
cd frontend && npm run test -- --run
```
Expected: all tests PASS (auth store: 3, SummaryCards: 2, PaymentTable: 3).

- [ ] **Step 2: Verify dev build works with backend running**

```bash
# Terminal 1:
cd backend && CGO_ENABLED=1 go run main.go

# Terminal 2:
cd frontend && npm run dev
```
Open `http://localhost:5173`, log in with `cs@test.com` / `password`, verify dashboard loads with 50 payments.

- [ ] **Step 3: Commit**

```bash
git commit --allow-empty -m "chore: all frontend tests pass, dev build verified"
```

---

## Task 19: Docker Compose

**Files:**
- Create: `docker-compose.yml`
- Create: `backend/Dockerfile`
- Create: `frontend/Dockerfile`
- Create: `frontend/nginx.conf`

- [ ] **Step 1: Create backend Dockerfile**

The backend build context must be the repo root so `openapi.yaml` is accessible. Dockerfile lives in `backend/` but docker-compose points context to `.`.

Create `backend/Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
# Copy openapi.yaml from repo root (build context = repo root)
COPY openapi.yaml ./openapi.yaml
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux go build -o server main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o seed ./script/seed/main.go

FROM alpine:3.19
RUN apk add --no-cache sqlite-libs
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/seed .
COPY --from=builder /app/openapi.yaml ./openapi.yaml
EXPOSE 8080
CMD ["./server"]
```

- [ ] **Step 2: Create nginx config for frontend**

Create `frontend/nginx.conf`:
```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /dashboard {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

- [ ] **Step 3: Create frontend Dockerfile**

Frontend build context is also repo root so `openapi.yaml` is accessible for client generation.

Create `frontend/Dockerfile`:
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
# Copy openapi.yaml from repo root (build context = repo root)
COPY openapi.yaml ../openapi.yaml
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run gen:api
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

- [ ] **Step 4: Create docker-compose.yml**

Create `docker-compose.yml` at repo root:
```yaml
version: '3.8'

services:
  backend:
    build:
      context: .
      dockerfile: backend/Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - sqlite_data:/app/data
    environment:
      HTTP_ADDR: ":8080"
      JWT_SECRET: "change-me-in-production"
      JWT_EXPIRED: "24h"
      OPENAPIYAML_LOCATION: "./openapi.yaml"
      DB_PATH: "/app/data/dashboard.db"
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: frontend/Dockerfile
    ports:
      - "3000:80"
    depends_on:
      - backend
    restart: unless-stopped

volumes:
  sqlite_data:
```

- [ ] **Step 5: Update seed script to use DB_PATH env var**

In `backend/script/seed/main.go`, the `DB_PATH` env var is already handled (Task 7 sets default to `dashboard.db`). For Docker, `DB_PATH=/app/data/dashboard.db` is set via docker-compose environment. No code change needed.

- [ ] **Step 6: Build and start Docker Compose**

```bash
docker-compose up --build -d
```
Expected: both services start without errors.

```bash
docker-compose logs backend | tail -5
```
Expected: `listening on :8080`.

- [ ] **Step 7: Seed Docker database**

```bash
docker-compose exec backend ./seed
```
Expected:
```
seeded 2 users
seeded 50 payments
Seed complete.
```

- [ ] **Step 8: Smoke test via Docker**

```bash
curl -s http://localhost:3000 | grep -c "PayDash\|paydash\|app" || echo "frontend up"
TOKEN=$(curl -s -X POST http://localhost:8080/dashboard/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"cs@test.com","password":"password"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
curl -s http://localhost:8080/dashboard/v1/payments \
  -H "Authorization: Bearer $TOKEN" | grep -o '"id"' | wc -l
```
Expected: 50 payment IDs returned.

- [ ] **Step 9: Commit**

```bash
git add docker-compose.yml backend/Dockerfile frontend/Dockerfile frontend/nginx.conf
git commit -m "feat: add Docker Compose with backend and frontend services"
```

---

## Task 20: Root Makefile + README

**Files:**
- Create: `Makefile` (root)
- Modify: `README.md`

- [ ] **Step 1: Create root Makefile**

Create `Makefile` at repo root:
```makefile
.DEFAULT_GOAL := help

help:
	@echo "Available targets:"
	@echo "  make up            - Start full stack (Docker Compose)"
	@echo "  make down          - Stop all containers"
	@echo "  make seed          - Seed database with test data"
	@echo "  make dev-backend   - Run backend locally"
	@echo "  make dev-frontend  - Run frontend dev server"
	@echo "  make gen-api       - Regenerate OpenAPI client"
	@echo "  make test-backend  - Run backend tests"
	@echo "  make test-frontend - Run frontend tests"
	@echo "  make test          - Run all tests"

up:
	docker-compose up --build

down:
	docker-compose down

seed:
	cd backend && CGO_ENABLED=1 go run ./script/seed/main.go

dev-backend:
	cd backend && make run

dev-frontend:
	cd frontend && npm run dev

gen-api:
	cd frontend && npm run gen:api

test-backend:
	cd backend && go test ./... -v

test-frontend:
	cd frontend && npm run test -- --run

test: test-backend test-frontend
```

- [ ] **Step 2: Update README.md**

Replace `README.md` at repo root:
```markdown
# Payment Dashboard

Internal dashboard for monitoring incoming payments. Go backend + Vue 3 frontend.

## Prerequisites

- Go 1.21+
- Node 20+
- Docker & Docker Compose
- make

## Quick Start (Docker)

```bash
# Start full stack
make up

# In a separate terminal, seed the database
make seed

# Open in browser
open http://localhost:3000
```

## Local Development

### Backend

```bash
cd backend
cp env.sample .env          # configure environment
make dep                    # install Go dependencies
make openapi-gen            # regenerate OpenAPI types
make gen-secret             # generate JWT secret, paste into .env
make seed                   # seed database with 50 payments + 2 users
make run                    # start server on :8080
```

### Frontend

```bash
cd frontend
npm install                 # install dependencies
npm run gen:api             # generate OpenAPI client from ../openapi.yaml
npm run dev                 # start dev server on :5173 (proxies /dashboard → :8080)
```

## Test Accounts

| Email | Password | Role |
|---|---|---|
| cs@test.com | password | cs |
| operation@test.com | password | operation |

## API

Spec: `openapi.yaml`

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/dashboard/v1/auth/login` | No | Login, returns JWT + role |
| GET | `/dashboard/v1/payments` | Bearer JWT | List payments (filter: status, sort, id) |

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
```

- [ ] **Step 3: Commit**

```bash
git add Makefile README.md
git commit -m "docs: add root Makefile and complete README setup instructions"
```

---

## Task 21: Final verification

- [ ] **Step 1: Run all backend tests**

```bash
cd backend && go test ./... -v
```
Expected: all PASS.

- [ ] **Step 2: Run all frontend tests**

```bash
cd frontend && npm run test -- --run
```
Expected: all PASS.

- [ ] **Step 3: Full Docker Compose smoke test**

```bash
make down
make up &
sleep 10
curl -s http://localhost:3000 | grep -i "paydash\|login\|payment" | head -3
make down
```
Expected: HTML response containing app content.

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "chore: final verification — all tests pass, Docker build confirmed"
```
