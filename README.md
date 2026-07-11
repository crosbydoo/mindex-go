# Mindex API

REST API for **Mindex** — a minimalist psychology journal/literature catalog. Drop-in compatible with the existing React + Vite frontend.

**Base URL (local):** `http://localhost:8080`

All API routes use `Content-Type: application/json` with this response envelope:

```json
{
  "code": 200,
  "status": true,
  "message": "Entries retrieved successfully",
  "data": {}
}
```

- `status`: `true` (success) or `false` (error)

---

## Quick Start

### 1. Prerequisites

- Go 1.22+
- PostgreSQL

### 2. Environment

Copy and edit the env file:

```bash
cp .env.example .env
```

Example `.env`:

```env
POSTGRES_URL=postgres://ristudev@localhost:5432/mindex?sslmode=disable
ADMIN_PASSWORD=admin123
CORS_ORIGIN=http://localhost:3005,http://localhost:5173
PORT=8080
```

`make run` loads `.env` automatically.

### 3. Run (hot reload)

```bash
make run          # hot reload with Air (auto-installs Air if missing)
make run-plain    # plain go run (no hot reload)
make test         # unit tests
make build        # compile binary to bin/api
```

Air watches `.go` / `.env` changes and restarts the server automatically.

Migrations run on startup. If `entries` is empty, 18 seed records are inserted automatically.

---

## Tech Stack

- Go 1.22+
- Gin HTTP router
- PostgreSQL via pgx
- Structured logging with slog
- Clean Architecture (`core/` layers)

## Project Structure

```text
mindex-api/
├── cmd/api/main.go           # Application entry point
├── core/
│   ├── auth/                 # HMAC token generation/verification
│   ├── config/               # Environment configuration
│   ├── database/             # Pool, migrations, seeding
│   ├── domain/               # Models and validation
│   ├── handler/              # HTTP handlers
│   ├── middleware/           # Auth, CORS, logging, recovery
│   ├── repository/           # PostgreSQL data access
│   ├── router/               # Route registration
│   └── service/              # Business logic
├── docs/
│   ├── api-endpoints.md      # Endpoint reference
│   ├── api-models.md         # Request/response models & mocks
│   ├── mocks/                # Ready-to-use JSON mock files
│   ├── vps-setup.md          # VPS setup (pull from Docker Hub)
│   ├── cloudflare-tunnel.md  # Free HTTPS tunnel (no domain)
│   └── deployment.md         # Full CI/CD GitHub → Docker Hub → VPS
├── data/seed-entries.json    # 18 seed psychology entries
├── migrations/               # SQL migrations
├── pkg/response/             # JSON response helpers
├── .env.example
├── Dockerfile
└── Makefile
```

---

## API Documentation

Detailed docs are split into separate files:

| Document | Description |
|----------|-------------|
| **[docs/api-endpoints.md](docs/api-endpoints.md)** | All endpoints, auth, status codes, curl examples |
| **[docs/api-models.md](docs/api-models.md)** | Request/response models, enums, TypeScript types |
| **[docs/mocks/](docs/mocks/)** | Ready-to-use JSON mock files for testing |

### Endpoint summary

| Method | Path | Auth |
|--------|------|------|
| `GET` | `/health` | No |
| `GET` | `/api/entries?page=&limit=&category=&archived=` | No |
| `GET` | `/api/categories?page=&limit=&archived=` | No |
| `POST` | `/api/login` | No |
| `POST` | `/api/logout` | No |
| `POST` | `/api/entries` | Bearer |
| `PUT` | `/api/entries?id={id}` | Bearer |
| `DELETE` | `/api/entries?id={id}` | Bearer |
| `POST` | `/api/entries/archive?id={id}` | Bearer |
| `POST` | `/api/entries/unarchive?id={id}` | Bearer |

### Quick test with mock files

```bash
# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d @docs/mocks/login-request.json

# Create entry (after login, set TOKEN)
curl -X POST http://localhost:8080/api/entries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @docs/mocks/entry-create-request.json
```

### Frontend (Vite) integration

```env
VITE_API_BASE_URL=http://localhost:8080
```

Or proxy in `vite.config.ts`:

```ts
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/health': 'http://localhost:8080',
  },
}
```

---
## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `POSTGRES_URL` | yes | PostgreSQL connection string |
| `ADMIN_PASSWORD` | yes* | Admin password for login (*required at login time) |
| `PORT` | no | Server port (default `8080`) |
| `CORS_ORIGIN` | no | Allowed frontend origin(s), comma-separated (e.g. `http://localhost:3005,https://app.vercel.app`) |

---

## Docker

Build and run locally:

```bash
docker build -t ristudev/mindex-go-server .
docker run -p 8080:8080 \
  -e POSTGRES_URL="postgres://user:password@host:5432/mindex?sslmode=disable" \
  -e ADMIN_PASSWORD="your-admin-password" \
  -e CORS_ORIGIN="http://localhost:3005" \
  ristudev/mindex-go-server
```

## Production Deployment (GitHub → Docker Hub → VPS)

VPS **does not build** the app — it only **pulls the image** from Docker Hub.

```
GitHub (push main) → GitHub Actions (build + push) → Docker Hub → VPS (pull + run)
```

### Setup VPS (image sudah ada di Docker Hub)

Panduan step-by-step: **[docs/vps-setup.md](docs/vps-setup.md)**

### Free HTTPS tanpa domain (Cloudflare Tunnel)

Supaya frontend Vercel bisa panggil API via HTTPS:

**[docs/cloudflare-tunnel.md](docs/cloudflare-tunnel.md)**

```bash
# Di VPS (setelah API sudah up)
cloudflared tunnel --url http://localhost:8080
# → https://xxxx.trycloudflare.com
```

Full CI/CD guide: **[docs/deployment.md](docs/deployment.md)**

Required GitHub Secrets (untuk auto-deploy): `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, `VPS_DEPLOY_PATH`

---

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Hot reload with Air (loads `.env`) |
| `make run-plain` | Run with `go run` (no hot reload) |
| `make install-air` | Install Air manually |
| `make build` | Build binary to `bin/api` |
| `make test` | Run all unit tests |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build Docker image |
