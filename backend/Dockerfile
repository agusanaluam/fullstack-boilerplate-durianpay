FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
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
