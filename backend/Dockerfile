FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
# Copy openapi.yaml from repo root (build context = repo root)
COPY openapi.yaml ./openapi.yaml
COPY backend/go.mod backend/go.sum ./
# Patch go version directive to be compatible with Go 1.24
RUN sed -i 's/^go 1\.25\.5/go 1.24/' go.mod && go mod download
COPY backend/ .
# Re-patch after copying backend sources (go.mod was overwritten by COPY)
RUN sed -i 's/^go 1\.25\.5/go 1.24/' go.mod && go mod tidy
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
