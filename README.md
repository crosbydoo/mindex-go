# Mindex API

REST API for **Mindex** — a minimalist psychology journal/literature catalog. Drop-in compatible with the existing React + Vite frontend.

**Base URL (local):** `http://localhost:8080`

All API routes use `Content-Type: application/json`. Errors return:

```json
{ "error": "Human readable message" }
```

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
CORS_ORIGIN=http://localhost:3005
PORT=8080
```

`make run` loads `.env` automatically.

### 3. Run

```bash
make run          # load .env + go run
make dev            # hot reload (requires air)
make test           # unit tests
make build          # compile binary to bin/api
```

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
├── data/seed-entries.json    # 18 seed psychology entries
├── migrations/               # SQL migrations
├── pkg/response/             # JSON response helpers
├── .env.example
├── Dockerfile
└── Makefile
```

---

## Data Models

### `Entry` (API response)

Returned by list, create, and update endpoints.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `number` | Auto-generated database ID |
| `title` | `string` | Entry title |
| `abstract` | `string` | Summary / abstract text |
| `category` | `string` | Psychology category (see enums) |
| `year` | `number` | Publication year |
| `author` | `string` | Author name |
| `source` | `string` | Journal, publisher, or source |
| `type` | `string` | Entry type (see enums) |
| `url` | `string` | Link to source (default `#`) |

```json
{
  "id": 1,
  "title": "Cognitive Behavioral Therapy for Depression: A Meta-Analysis",
  "abstract": "This meta-analysis examines the efficacy of cognitive behavioral therapy...",
  "category": "Clinical Psychology",
  "year": 2023,
  "author": "Sarah Mitchell",
  "source": "Journal of Clinical Psychology",
  "type": "Journal",
  "url": "https://example.com/cbt-depression"
}
```

> `created_at` exists in the database but is **not** exposed in API responses.

### `EntryInput` (create / update body)

Same fields as `Entry` **without** `id`.

| Field | Required | Rules |
|-------|----------|-------|
| `title` | yes | Non-empty after trim |
| `abstract` | yes | Non-empty after trim |
| `category` | yes | Must be a valid category enum |
| `year` | yes | Positive integer (`> 0`) |
| `author` | yes | Non-empty after trim |
| `source` | yes | Non-empty after trim |
| `type` | yes | Must be a valid entry type enum |
| `url` | no | Defaults to `"#"` if missing or empty |

### Enums

#### `category` — allowed values

- `Clinical Psychology`
- `Developmental Psychology`
- `Cognitive Psychology`
- `Social Psychology`
- `Educational Psychology`
- `Mental Health`
- `Research Methods`

#### `type` — allowed values

- `Journal`
- `Article`
- `Thesis`
- `Literature Review`

### Auth models

#### Login request

```json
{ "password": "string" }
```

#### Login response

```json
{ "token": "<hex-string>" }
```

#### Logout response

```json
{ "ok": true }
```

#### Health response

```json
{ "status": "ok" }
```

---

## Authentication

This API uses a **static HMAC bearer token**, not JWT.

```
token = HMAC-SHA256(key=ADMIN_PASSWORD, message="mindex-admin-session").hex()
```

### Flow

1. `POST /api/login` with admin password → receive `token`
2. Store token in client (`sessionStorage` key: `mindex_admin_token`)
3. Send token on protected routes:

```
Authorization: Bearer <token>
```

4. `POST /api/logout` → client removes token from storage

### Route access

| Access | Routes |
|--------|--------|
| Public | `GET /health`, `GET /api/entries`, `POST /api/login`, `POST /api/logout` |
| Protected (Bearer) | `POST /api/entries`, `PUT /api/entries`, `DELETE /api/entries` |

> Logout is stateless — the server cannot invalidate the static token. The client must clear `sessionStorage` after logout.

---

## API Endpoints

### `GET /health`

Health check.

**Response `200`**

```json
{ "status": "ok" }
```

---

### `GET /api/entries`

List all psychology literature entries. **No auth required.**

**Response `200`** — array of `Entry`, ordered by `year DESC, id DESC`

```json
[
  {
    "id": 4,
    "title": "Social Identity and Group Behavior in Online Communities",
    "abstract": "...",
    "category": "Social Psychology",
    "year": 2024,
    "author": "Michael Okafor",
    "source": "Social Psychology Quarterly",
    "type": "Article",
    "url": "https://example.com/social-identity"
  }
]
```

---

### `POST /api/login`

Admin login.

**Request body**

```json
{ "password": "admin123" }
```

**Response `200`**

```json
{ "token": "a1b2c3..." }
```

**Errors**

| Status | Message |
|--------|---------|
| `401` | `Invalid password` |
| `503` | `ADMIN_PASSWORD is not configured on the server` |
| `405` | `Method not allowed` |
| `400` | `Invalid request body` |

---

### `POST /api/logout`

Logout. **No auth required.**

**Response `200`**

```json
{ "ok": true }
```

Client should remove `mindex_admin_token` from `sessionStorage` after a successful response.

---

### `POST /api/entries`

Create a new entry. **Auth required.**

**Request body** — `EntryInput`

```json
{
  "title": "Test Entry",
  "abstract": "Abstract text",
  "category": "Clinical Psychology",
  "year": 2024,
  "author": "Jane Doe",
  "source": "Test Journal",
  "type": "Journal",
  "url": "https://example.com"
}
```

**Response `201`** — created `Entry`

**Errors**

| Status | Message |
|--------|---------|
| `400` | `Invalid entry payload` |
| `401` | `Unauthorized` |
| `500` | `Internal server error` |

---

### `PUT /api/entries?id={id}`

Update an existing entry. **Auth required.**

**Query params**

| Param | Type | Description |
|-------|------|-------------|
| `id` | `number` | Entry ID (positive integer) |

**Request body** — `EntryInput` (same as create)

**Response `200`** — updated `Entry`

**Errors**

| Status | Message |
|--------|---------|
| `400` | `Invalid entry id` or `Invalid entry payload` |
| `401` | `Unauthorized` |
| `404` | `Entry not found` |
| `500` | `Internal server error` |

---

### `DELETE /api/entries?id={id}`

Delete an entry. **Auth required.**

**Query params**

| Param | Type | Description |
|-------|------|-------------|
| `id` | `number` | Entry ID (positive integer) |

**Response `204`** — empty body

**Errors**

| Status | Message |
|--------|---------|
| `400` | `Invalid entry id` |
| `401` | `Unauthorized` |
| `404` | `Entry not found` |
| `500` | `Internal server error` |

---

## How to Use

### curl

```bash
# 1. Health check
curl http://localhost:8080/health

# 2. List entries (public)
curl http://localhost:8080/api/entries

# 3. Login and save token
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}' | jq -r '.token')

# 4. Create entry
curl -X POST http://localhost:8080/api/entries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Test Entry",
    "abstract": "Abstract text",
    "category": "Clinical Psychology",
    "year": 2024,
    "author": "Jane Doe",
    "source": "Test Journal",
    "type": "Journal",
    "url": "https://example.com"
  }'

# 5. Update entry
curl -X PUT "http://localhost:8080/api/entries?id=1" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Updated Entry",
    "abstract": "Updated abstract",
    "category": "Clinical Psychology",
    "year": 2024,
    "author": "Jane Doe",
    "source": "Test Journal",
    "type": "Journal",
    "url": "https://example.com"
  }'

# 6. Delete entry
curl -X DELETE "http://localhost:8080/api/entries?id=1" \
  -H "Authorization: Bearer $TOKEN"

# 7. Logout
curl -X POST http://localhost:8080/api/logout
```

### JavaScript / fetch

```ts
const API_BASE = 'http://localhost:8080';

// List entries (public)
const entries = await fetch(`${API_BASE}/api/entries`).then((r) => r.json());

// Login
const { token } = await fetch(`${API_BASE}/api/login`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ password: 'admin123' }),
}).then((r) => r.json());

sessionStorage.setItem('mindex_admin_token', token);

// Create entry (protected)
const created = await fetch(`${API_BASE}/api/entries`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
  },
  body: JSON.stringify({
    title: 'Test Entry',
    abstract: 'Abstract text',
    category: 'Clinical Psychology',
    year: 2024,
    author: 'Jane Doe',
    source: 'Test Journal',
    type: 'Journal',
    url: 'https://example.com',
  }),
}).then((r) => r.json());

// Update entry
const updated = await fetch(`${API_BASE}/api/entries?id=1`, {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
  },
  body: JSON.stringify({ /* EntryInput fields */ }),
}).then((r) => r.json());

// Delete entry (204, no body)
await fetch(`${API_BASE}/api/entries?id=1`, {
  method: 'DELETE',
  headers: { Authorization: `Bearer ${token}` },
});

// Logout
await fetch(`${API_BASE}/api/logout`, { method: 'POST' });
sessionStorage.removeItem('mindex_admin_token');
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
| `CORS_ORIGIN` | no | Allowed frontend origin (e.g. `http://localhost:3005`) |

---

## Docker

Build and run locally:

```bash
docker build -t mindex-api .
docker run -p 8080:8080 \
  -e POSTGRES_URL="postgres://user:password@host:5432/mindex?sslmode=disable" \
  -e ADMIN_PASSWORD="your-admin-password" \
  -e CORS_ORIGIN="http://localhost:3005" \
  mindex-api
```

## Production Deployment (GitHub → Docker Hub → VPS)

VPS **does not build** the app — it only **pulls the image** from Docker Hub.

```
GitHub (push main) → GitHub Actions (build + push) → Docker Hub → VPS (pull + run)
```

See full guide: **[docs/deployment.md](docs/deployment.md)**

Quick VPS setup:

```bash
# On VPS (one-time)
mkdir -p /opt/mindex-api && cd /opt/mindex-api
# copy deploy/* files here
cp .env.example .env && nano .env
chmod +x deploy.sh && ./deploy.sh
```

Required GitHub Secrets: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, `VPS_DEPLOY_PATH`

---

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Load `.env` and run with `go run` |
| `make dev` | Hot reload with Air |
| `make build` | Build binary to `bin/api` |
| `make test` | Run all unit tests |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build Docker image |
