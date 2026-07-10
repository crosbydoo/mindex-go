# Go API Migration Prompt — Mindex Psychology Journal

## Context

Build a REST API in **Go** to replace the existing Vercel serverless API for **Mindex** — a minimalist psychology journal/literature catalog UI.

The frontend is a **React + Vite** SPA that already calls these endpoints. The Go API must be **drop-in compatible** with the current client contract so the frontend only needs a base URL change.

**Current stack being replaced:**

- Vercel serverless functions (`api/login.ts`, `api/entries.ts`)
- Neon/Postgres via `@neondatabase/serverless`
- Single admin password auth with HMAC-derived bearer token

**App name:** `mindex`  
**Domain:** Psychology literature entries (journals, articles, theses, literature reviews)

---

## Requirements

### 1. Tech stack (Go)

Use:

- **Go 1.22+**
- HTTP router: `chi` or `gin` or `echo` (your choice, but keep routes clean)
- Database: **PostgreSQL** (same as current Neon setup)
- DB access: `pgx` or `sqlc` + `migrate`/`goose` for migrations
- Config: environment variables only (no hardcoded secrets)
- JSON encoding/decoding with standard `encoding/json`
- Structured logging (`slog` or `zerolog`)

Project layout (suggested):

```
cmd/server/main.go
internal/
  config/
  domain/          # models, enums
  handler/         # HTTP handlers
  middleware/      # auth, cors, logging
  repository/      # DB queries
  service/         # business logic (optional)
  auth/            # token + password verification
migrations/
data/seed-entries.json
```

---

### 2. Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | no | Server port (default `8080`) |
| `POSTGRES_URL` | yes | PostgreSQL connection string |
| `ADMIN_PASSWORD` | yes | Single admin password for login |
| `CORS_ORIGIN` | no | Allowed frontend origin (e.g. `http://localhost:3005`) |

If `ADMIN_PASSWORD` is missing at login time → return **503** with `{ "error": "ADMIN_PASSWORD is not configured on the server" }`.

---

### 3. Domain models

#### Enums (validate on write)

**`EntryType`** — allowed values:

- `Journal`
- `Article`
- `Thesis`
- `Literature Review`

**`Category`** — allowed values:

- `Clinical Psychology`
- `Developmental Psychology`
- `Cognitive Psychology`
- `Social Psychology`
- `Educational Psychology`
- `Mental Health`
- `Research Methods`

#### `Entry` (API response)

```go
type Entry struct {
    ID       int64  `json:"id"`
    Title    string `json:"title"`
    Abstract string `json:"abstract"`
    Category string `json:"category"`
    Year     int    `json:"year"`
    Author   string `json:"author"`
    Source   string `json:"source"`
    Type     string `json:"type"`
    URL      string `json:"url"`
}
```

#### `EntryInput` (create/update body — no `id`)

```go
type EntryInput struct {
    Title    string `json:"title"`
    Abstract string `json:"abstract"`
    Category string `json:"category"`
    Year     int    `json:"year"`
    Author   string `json:"author"`
    Source   string `json:"source"`
    Type     string `json:"type"`
    URL      string `json:"url"`
}
```

**Note:** `created_at` exists in DB but is **not** returned to the client (match current API).

---

### 4. Database schema

```sql
CREATE TABLE IF NOT EXISTS entries (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    abstract TEXT NOT NULL,
    category TEXT NOT NULL,
    year INTEGER NOT NULL,
    author TEXT NOT NULL,
    source TEXT NOT NULL,
    type TEXT NOT NULL,
    url TEXT NOT NULL DEFAULT '#',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Indexes (recommended):**

- `CREATE INDEX idx_entries_year_id ON entries (year DESC, id DESC);`

**Seed behavior:** On first startup (or via migration/seed command), if `entries` table is empty, insert **18 records** from `data/seed-entries.json` (same file as the frontend repo). Default `url` to `#` when empty.

**List ordering:** `ORDER BY year DESC, id DESC`

---

### 5. Authentication (must match current behavior)

This is **not JWT**. It is a **static HMAC token** derived from the admin password.

#### Login flow

**`POST /api/login`**

Request:

```json
{ "password": "string" }
```

Success **200**:

```json
{ "token": "<hex-string>" }
```

Errors:

- **401** `{ "error": "Invalid password" }` — wrong password
- **503** — `ADMIN_PASSWORD` not configured
- **405** — non-POST
- **500** — internal error

#### Token generation (must be identical for frontend compatibility)

```
token = HMAC-SHA256(key=ADMIN_PASSWORD, message="mindex-admin-session").hex()
```

#### Password verification

Use **constant-time comparison** (`crypto/subtle.ConstantTimeCompare`) when comparing password to `ADMIN_PASSWORD`.

#### Protected routes

Mutating entry endpoints require header:

```
Authorization: Bearer <token>
```

Verify token by recomputing expected HMAC and comparing with constant-time equality.

If invalid/missing → **401** `{ "error": "Unauthorized" }`

**Public route:** `GET /api/entries` — no auth required.

---

### 6. API endpoints

All paths are under `/api`. All JSON responses use `Content-Type: application/json`.

Error shape (consistent):

```json
{ "error": "Human readable message" }
```

---

#### `GET /api/entries`

- **Auth:** none
- **Response 200:** `Entry[]`
- **Ordering:** year DESC, id DESC

---

#### `POST /api/entries`

- **Auth:** required (admin)
- **Body:** `EntryInput`
- **Validation:**
  - Required non-empty (after trim): `title`, `abstract`, `category`, `author`, `source`, `type`
  - `year` must be a valid integer
  - `category` must be one of the 7 allowed categories
  - `type` must be one of the 4 allowed types
  - `url` optional; default `"#"` if missing/empty
- **Response 201:** created `Entry`
- **Response 400:** `{ "error": "Invalid entry payload" }`

---

#### `PUT /api/entries?id={id}`

- **Auth:** required (admin)
- **Query:** `id` — positive integer
- **Body:** `EntryInput` (same validation as POST)
- **Response 200:** updated `Entry`
- **Response 400:** invalid id or payload
- **Response 404:** `{ "error": "Entry not found" }`

---

#### `DELETE /api/entries?id={id}`

- **Auth:** required (admin)
- **Query:** `id` — positive integer
- **Response 204:** no body
- **Response 400:** `{ "error": "Invalid entry id" }`
- **Response 404:** `{ "error": "Entry not found" }`

---

#### `POST /api/login`

(See auth section above.)

---

### 7. Validation rules (detail)

| Field | Rules |
|-------|-------|
| `title` | required, trim whitespace |
| `abstract` | required, trim whitespace |
| `category` | required, must match enum |
| `year` | required, integer |
| `author` | required, trim whitespace |
| `source` | required, trim whitespace |
| `type` | required, must match enum |
| `url` | optional, default `#` |

Trim all string fields before save.

---

### 8. Frontend contract (do not break)

The React client (`src/lib/api.ts`) expects:

| Client function | HTTP call |
|-----------------|-----------|
| `fetchEntries()` | `GET /api/entries` |
| `loginAdmin(password)` | `POST /api/login` → `{ token }` |
| `createEntry(input)` | `POST /api/entries` + Bearer |
| `updateEntry(id, input)` | `PUT /api/entries?id={id}` + Bearer |
| `deleteEntry(id)` | `DELETE /api/entries?id={id}` + Bearer |

Client stores token in `sessionStorage` key `mindex_admin_token`.

Client treats **204** on delete as success with no body.

Client reads error from `response.json().error` string.

---

### 9. Middleware

Implement:

1. **Request logging** — method, path, status, duration
2. **Recovery** — panic → 500 `{ "error": "Internal server error" }`
3. **CORS** — allow frontend origin, methods `GET, POST, PUT, DELETE, OPTIONS`, headers `Content-Type, Authorization`
4. **Auth middleware** — reusable for protected routes

---

### 10. Health & ops (nice to have)

- `GET /health` → `{ "status": "ok" }`
- Graceful shutdown on SIGINT/SIGTERM
- DB connection pool with sensible limits
- Run migrations on startup OR provide `make migrate` command

---

### 11. Deliverables

1. Full Go source code
2. SQL migration files
3. `README.md` with:
   - env setup
   - how to run locally
   - how to connect frontend (`VITE_API_URL` or Vite proxy)
4. `Dockerfile` (optional but preferred)
5. Example `curl` commands for all endpoints
6. Unit tests for:
   - auth token generation/verification
   - entry validation
   - CRUD repository (use testcontainers or sqlite if needed)

---

### 12. Example curl tests

```bash
# List (public)
curl http://localhost:8080/api/entries

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"password":"your-admin-password"}'

# Create
curl -X POST http://localhost:8080/api/entries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
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

# Update
curl -X PUT "http://localhost:8080/api/entries?id=1" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{ ... }'

# Delete
curl -X DELETE "http://localhost:8080/api/entries?id=1" \
  -H "Authorization: Bearer <token>"
```

---

### 13. Future improvements (out of scope for v1, but design for extension)

- Pagination/filtering on `GET /api/entries` (`?category=`, `?type=`, `?year=`, `?q=`) — frontend currently filters client-side
- Proper JWT with expiry instead of static HMAC token
- Multi-user admin accounts table
- `created_at` / `updated_at` in API response
- OpenAPI/Swagger spec at `/swagger/index.html`

---

### 14. Reference: current TypeScript implementation

Use these as the source of truth for behavior parity:

- Auth: `api/lib/auth.ts` — HMAC-SHA256, session label `mindex-admin-session`
- CRUD: `api/lib/db.ts` — schema, seed, queries
- Handlers: `api/entries.ts`, `api/login.ts`
- Types: `src/lib/types.ts`
- Seed data: `data/seed-entries.json` (18 entries)

---

## Acceptance criteria

- [ ] All 5 client API calls work without frontend code changes (only base URL)
- [ ] `GET /api/entries` is public; create/update/delete require Bearer token
- [ ] Token algorithm matches HMAC-SHA256 with message `mindex-admin-session`
- [ ] Postgres schema + seed matches existing project
- [ ] Same HTTP status codes and error JSON shape
- [ ] `PUT`/`DELETE` use query param `?id=`, not path param `/entries/:id`
- [ ] Delete returns **204** with empty body

---

## Optional add-on prompt (frontend wiring)

After the Go API is ready, update the Vite app to point to it:

```
Add a Vite dev proxy or VITE_API_BASE_URL env var so fetch('/api/...') 
targets the Go server at http://localhost:8080 during development.
```

---

## Summary of current API

| Area | Current implementation |
|------|------------------------|
| **Auth** | Single password → static HMAC bearer token |
| **Public** | List all entries |
| **Protected** | Create, update, delete entries |
| **DB** | PostgreSQL `entries` table |
| **Seed** | 18 psychology literature records |
| **Enums** | 7 categories, 4 entry types |
