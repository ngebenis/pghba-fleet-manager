# ── Stage 1: Build Frontend ───────────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/index.html ./
COPY frontend/vite.config.ts ./
COPY frontend/tsconfig*.json ./
COPY frontend/src/ ./src/
COPY frontend/public/ ./public/
RUN npm run build

# ── Stage 2: Build Backend ────────────────────────────────────────────────────
FROM golang:1.26-alpine AS backend-builder

ENV GOTOOLCHAIN=local
ENV GONOSUMCHECK=*
ENV GOFLAGS=-mod=mod
WORKDIR /app

COPY backend/go.mod ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o pghba-manager ./cmd/server

# ── Stage 3: Final Image ──────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

COPY --from=backend-builder /app/pghba-manager .
COPY --from=frontend-builder /app/frontend/dist ./static

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget -qO- http://localhost:8080/api/auth/ping || exit 1

CMD ["./pghba-manager"]
