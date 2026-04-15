# Skill: Docker & Docker Compose

## Amaç
Multi-stage Dockerfile + Docker Compose + Makefile target'ları oluştur.

## Girdiler
- `project-spec.yml` → `docker`, `container`, `stack` bölümleri

---

## Go Dockerfile (multi-stage)

```dockerfile
# ── Stage 1: Frontend build (dashboard varsa) ──
FROM node:20-alpine AS frontend
WORKDIR /app
COPY web/frontend/package*.json ./
RUN npm ci --production=false
COPY web/frontend/ ./
RUN npm run build

# ── Stage 2: Go build ──
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Frontend embed (dashboard varsa)
COPY --from=frontend /app/dist web/frontend/dist
# Build
ARG VERSION=dev
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o /app/bin/{binary_name} .

# ── Stage 3: Runtime ──
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/bin/{binary_name} /usr/local/bin/
EXPOSE {port}
ENTRYPOINT ["{binary_name}"]
CMD ["server", "--port", "{port}"]
```

**Frontend yoksa** Stage 1'i kaldır ve Stage 2'den COPY --from=frontend satırını sil.

---

## Python Dockerfile (multi-stage)

```dockerfile
# ── Stage 1: Frontend build (dashboard varsa) ──
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci --production=false
COPY frontend/ ./
RUN npm run build

# ── Stage 2: Runtime ──
FROM python:3.11-slim
WORKDIR /app

# Non-root user
RUN groupadd -r app && useradd -r -g app app

# Dependencies
COPY backend/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Application
COPY backend/ ./

# Frontend (dashboard varsa)
COPY --from=frontend /app/dist ./static/

# Permissions
RUN chown -R app:app /app
USER app

EXPOSE {port}
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:{port}/health || exit 1
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "{port}"]
```

---

## Docker Compose — Go Projeleri (basit)

```yaml
services:
  {name}:
    build:
      context: .
      args:
        VERSION: "${VERSION:-dev}"
    ports:
      - "{port}:{port}"
    volumes:
      - ./data:/data
      - ./.{name}.yml:/config/.{name}.yml:ro
    environment:
      - SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL:-}
      - DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL:-}
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-}
      - TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID:-}
    restart: unless-stopped
```

---

## Docker Compose — Python Projeleri (full stack)

```yaml
services:
  postgres:
    image: postgres:16-alpine
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: ${DB_NAME:-{name}}
      POSTGRES_USER: ${DB_USER:-{name}}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-{name}}"]
      interval: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      retries: 5

  backend:
    build: .
    ports:
      - "{port}:{port}"
    environment:
      - DATABASE_URL=postgresql+asyncpg://${DB_USER:-{name}}:${DB_PASSWORD:-changeme}@postgres:5432/${DB_NAME:-{name}}
      - REDIS_URL=redis://redis:6379/0
      - SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL:-}
      - DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL:-}
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-}
      - TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID:-}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  celery-worker:
    build: .
    command: celery -A app.tasks.celery_app worker -l info --concurrency=2
    environment:
      - DATABASE_URL=postgresql+asyncpg://${DB_USER:-{name}}:${DB_PASSWORD:-changeme}@postgres:5432/${DB_NAME:-{name}}
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - redis
      - postgres
    restart: unless-stopped

  celery-beat:
    build: .
    command: celery -A app.tasks.celery_app beat -l info
    environment:
      - DATABASE_URL=postgresql+asyncpg://${DB_USER:-{name}}:${DB_PASSWORD:-changeme}@postgres:5432/${DB_NAME:-{name}}
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - redis
      - postgres
    restart: unless-stopped

volumes:
  pgdata:
```

---

## Docker Compose — Dev Override

```yaml
# docker-compose.dev.yml
# Kullanım: docker compose -f docker-compose.yml -f docker-compose.dev.yml up
services:
  backend:
    volumes:
      - ./backend:/app
    environment:
      - DEBUG=true
    command: uvicorn app.main:app --reload --host 0.0.0.0 --port {port}
    # Go projeler için:
    # command: air  (hot reload)

  postgres:
    ports:
      - "5432:5432"  # Dev'de dışarıdan erişim

  redis:
    ports:
      - "6379:6379"
```

---

## Docker Compose — Demo Override

```yaml
# docker-compose.demo.yml
# Demo instance: read-only, seed data, Celery kapalı
services:
  backend:
    environment:
      - DEMO_MODE=true
    # Celery gereksiz — seed data yeterli

  celery-worker:
    profiles: ["disabled"]

  celery-beat:
    profiles: ["disabled"]
```

---

## .dockerignore

```
.git
.gitignore
*.md
!README.md
node_modules
__pycache__
*.pyc
.env
.env.*
*.db
*.sqlite
bin/
dist/
.venv/
.DS_Store
```

---

## Makefile Target'ları

```makefile
# ── Docker ──
docker-build:
	docker build -t {name} .

docker-run:
	docker compose up -d

docker-stop:
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	docker compose down -v
	docker rmi {name} 2>/dev/null || true

docker-dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up
```

---

## Doğrulama
- `docker build -t {name} .` başarılı (multi-stage hatasız)
- `docker compose up -d` tüm servisler healthy
- `docker compose ps` — tüm servisler "Up" durumunda
- Port erişimi çalışıyor: `curl http://localhost:{port}/api/health`
- Graceful shutdown: `docker compose down` temiz kapanma
- Image boyutu makul (Go: <50MB, Python: <200MB)
