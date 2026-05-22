FROM node:22-alpine AS builder
WORKDIR /app
# Copy openapi.yaml from repo root (build context = repo root)
# openapi-ts.config.ts uses '../openapi.yaml' so place it one level up from /app
COPY openapi.yaml /openapi.yaml
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
