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
