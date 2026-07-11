# API Endpoints

Base URL (local): `http://localhost:8080`

All JSON responses use `Content-Type: application/json` and this envelope:

```json
{
  "code": 200,
  "status": true,
  "message": "Entries retrieved successfully",
  "data": {}
}
```

- `status`: `true` (success) or `false` (error)
- `data`: payload on success, `null` on error

---

## Overview

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | No | Health check |
| `GET` | `/api/entries` | No | List entries (paginated; optional category / archived) |
| `GET` | `/api/categories` | No | All categories with paginated entries each |
| `GET` | `/api/categories/list` | No | List category definitions (`id`, `name`, `entry_count`) |
| `POST` | `/api/login` | No | Admin login |
| `POST` | `/api/logout` | No | Logout |
| `POST` | `/api/entries` | Bearer | Create entry |
| `PUT` | `/api/entries?id={id}` | Bearer | Update entry |
| `DELETE` | `/api/entries?id={id}` | Bearer | Delete entry permanently |
| `POST` | `/api/entries/archive?id={id}` | Bearer | Archive entry |
| `POST` | `/api/entries/unarchive?id={id}` | Bearer | Unarchive entry |
| `POST` | `/api/categories` | Bearer | Create category |
| `PUT` | `/api/categories?id={id}` | Bearer | Rename category |
| `DELETE` | `/api/categories?id={id}` | Bearer | Delete category |

**Protected routes** require header:

```
Authorization: Bearer <token>
```

See request/response models: **[api-models.md](./api-models.md)**

---

## Authentication

| Access | Endpoints |
|--------|-----------|
| Public | `GET /health`, `GET /api/entries`, `GET /api/categories`, `GET /api/categories/list`, `POST /api/login`, `POST /api/logout` |
| Protected | `POST/PUT/DELETE /api/entries`, `POST /api/entries/archive`, `POST /api/entries/unarchive`, `POST/PUT/DELETE /api/categories` |

Token algorithm:

```
HMAC-SHA256(key=ADMIN_PASSWORD, message="mindex-admin-session").hex()
```

Client stores token in `sessionStorage` key: `mindex_admin_token`.

---

## `GET /health`

Health check.

### Request

No body. No auth.

```bash
curl http://localhost:8080/health
```

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with `data: { "status": "ok" }` |

---

## `GET /api/entries`

List psychology literature entries with pagination. **Public.**

### Query params

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | number | `1` | Page number (>= 1) |
| `limit` | number | `10` | Items per page (max 100) |
| `category` | string | — | Optional category filter |
| `archived` | string | `false` | `false` / `active` = active only; `true` / `archived` = archived only; `all` = both |

### Request

```bash
curl "http://localhost:8080/api/entries?page=1&limit=10"
curl "http://localhost:8080/api/entries?page=1&limit=5&category=Clinical%20Psychology"
curl "http://localhost:8080/api/entries?archived=true"
curl "http://localhost:8080/api/entries?archived=all"
```

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with `data.items` + `data.pagination` |
| `400` | Invalid pagination / category / archived filter |
| `500` | Envelope error |

### Notes

- Ordered by `year DESC, id DESC`.
- Default list excludes archived entries (`is_archived = false`).
- Returns empty `items: []` when no matches.
- Each item includes `is_archived`.

---

## `GET /api/categories`

List **every category**, each with its own paginated entries. **Public.**

### Query params

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | number | `1` | Page number applied **per category** |
| `limit` | number | `10` | Items per category page (max 100) |
| `archived` | string | `false` | Same filter as `GET /api/entries` |

### Request

```bash
curl "http://localhost:8080/api/categories?page=1&limit=5"
curl "http://localhost:8080/api/categories?archived=true"
```

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with `data.categories[]` (each has `category`, `items`, `pagination`) |
| `400` | Invalid pagination / archived filter |
| `500` | Envelope error |

---

## `GET /api/categories/list`

List category definitions (for admin manage UI). **Public.**

### Request

```bash
curl "http://localhost:8080/api/categories/list"
```

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with `data.items[]` (`id`, `name`, `entry_count`) |
| `500` | Envelope error |

Example:

```json
{
  "code": 200,
  "status": true,
  "message": "Category list retrieved successfully",
  "data": {
    "items": [
      { "id": 1, "name": "Clinical Psychology", "entry_count": 3 }
    ]
  }
}
```

---

## `POST /api/categories`

Create a category. **Auth required.**

### Request

```bash
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"Neuropsychology"}'
```

Body: `{ "name": string }` — required, trimmed, max 100 chars.

### Response

| Status | Body |
|--------|------|
| `201` | Envelope with created category in `data` |
| `400` | Invalid category payload |
| `401` | Unauthorized |
| `409` | Category already exists |
| `500` | Envelope error |

---

## `PUT /api/categories?id={id}`

Rename a category. Also updates `entries.category` for matching rows. **Auth required.**

### Request

```bash
curl -X PUT "http://localhost:8080/api/categories?id=1" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"Clinical Psych"}'
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | query | yes | Positive integer category ID |

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with updated category in `data` |
| `400` | Invalid category id / payload |
| `401` | Unauthorized |
| `404` | Category not found |
| `409` | Category name already exists |
| `500` | Envelope error |

---

## `DELETE /api/categories?id={id}`

Delete a category. **Rejected with 409** if any entries still use it. **Auth required.**

### Request

```bash
curl -X DELETE "http://localhost:8080/api/categories?id=8" \
  -H "Authorization: Bearer <token>"
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | query | yes | Positive integer category ID |

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with `data: null` |
| `400` | Invalid category id |
| `401` | Unauthorized |
| `404` | Category not found |
| `409` | Category is still used by entries |
| `500` | Envelope error |

---

## `POST /api/login`

Authenticate admin and receive bearer token.

### Request

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}'
```

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |

Body: `LoginRequest` — see [api-models.md](./api-models.md)

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with `data: { "token": "<hex-string>" }` |
| `400` | Envelope error — Invalid request body |
| `401` | Envelope error — Invalid password |
| `405` | Envelope error — Method not allowed |
| `503` | Envelope error — ADMIN_PASSWORD not configured |
| `500` | Envelope error |

---

## `POST /api/logout`

Logout. **Public** (client clears token locally).

### Request

```bash
curl -X POST http://localhost:8080/api/logout
```

No body required.

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with `data: null` |
| `405` | Envelope error — Method not allowed |
| `500` | Envelope error |

### Client action

After success, remove `mindex_admin_token` from `sessionStorage`.

---

## `POST /api/entries`

Create a new entry. **Auth required.**

### Request

```bash
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
```

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `Authorization` | `Bearer <token>` |

Body: `EntryInput` — see [api-models.md](./api-models.md)

### Response

| Status | Body |
|--------|------|
| `201` | Envelope with created `Entry` in `data` |
| `400` | Envelope error — Invalid entry payload |
| `401` | Envelope error — Unauthorized |
| `500` | Envelope error |

---

## `PUT /api/entries?id={id}`

Update an existing entry. **Auth required.**

### Request

```bash
curl -X PUT "http://localhost:8080/api/entries?id=1" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
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
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | query | yes | Positive integer entry ID |

Body: `EntryInput` — same as create.

> Uses query param `?id=`, **not** path param `/entries/:id`.

### Response

| Status | Body |
|--------|------|
| `200` | Envelope with updated `Entry` in `data` |
| `400` | Envelope error — Invalid entry id / payload |
| `401` | Envelope error — Unauthorized |
| `404` | Envelope error — Entry not found |
| `500` | Envelope error |

---

## `DELETE /api/entries?id={id}`

Permanently delete one or many entries. **Auth required.** Prefer archive when you want to hide without deleting.

### Request — single

```bash
curl -X DELETE "http://localhost:8080/api/entries?id=1" \
  -H "Authorization: Bearer <token>"
```

### Request — bulk (select all / checklist)

```bash
curl -X DELETE "http://localhost:8080/api/entries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"ids":[1,2,3]}'
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | query | one of | Single entry ID |
| `ids` | body | one of | Array of entry IDs |

### Response

| Status | Body |
|--------|------|
| `200` | Single: `data: null` — Bulk: `data: { "affected": 3, "ids": [1,2,3] }` |
| `400` | Envelope error — Invalid entry id |
| `401` | Envelope error — Unauthorized |
| `404` | Envelope error — Entry not found (single only) |
| `500` | Envelope error |

---

## `POST /api/entries/archive?id={id}`

Archive one or many entries (`is_archived = true`). Archived entries are hidden from the default list. **Auth required.**

### Request — single

```bash
curl -X POST "http://localhost:8080/api/entries/archive?id=1" \
  -H "Authorization: Bearer <token>"
```

### Request — bulk (select all / checklist)

```bash
curl -X POST "http://localhost:8080/api/entries/archive" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"ids":[1,2,3]}'
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | query | one of | Single entry ID |
| `ids` | body | one of | Array of entry IDs |

### Response

| Status | Body |
|--------|------|
| `200` | Single: archived `Entry` — Bulk: `{ "affected": 3, "ids": [1,2,3] }` |
| `400` | Envelope error — Invalid entry id |
| `401` | Envelope error — Unauthorized |
| `404` | Envelope error — Entry not found (single only) |
| `500` | Envelope error |

---

## `POST /api/entries/unarchive?id={id}`

Restore one or many archived entries (`is_archived = false`). **Auth required.**

### Request — single

```bash
curl -X POST "http://localhost:8080/api/entries/unarchive?id=1" \
  -H "Authorization: Bearer <token>"
```

### Request — bulk (select all / checklist)

```bash
curl -X POST "http://localhost:8080/api/entries/unarchive" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"ids":[1,2,3]}'
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | query | one of | Single entry ID |
| `ids` | body | one of | Array of entry IDs |

### Response

| Status | Body |
|--------|------|
| `200` | Single: restored `Entry` — Bulk: `{ "affected": 3, "ids": [1,2,3] }` |
| `400` | Envelope error — Invalid entry id |
| `401` | Envelope error — Unauthorized |
| `404` | Envelope error — Entry not found (single only) |
| `500` | Envelope error |

---

## Full workflow (curl)

```bash
# 1. Health
curl http://localhost:8080/health

# 2. List active (public)
curl "http://localhost:8080/api/entries?page=1&limit=10"

# 3. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}' | jq -r '.data.token')

# 4. Create
curl -X POST http://localhost:8080/api/entries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @docs/mocks/entry-create-request.json

# 5. Update
curl -X PUT "http://localhost:8080/api/entries?id=1" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @docs/mocks/entry-update-request.json

# 6. Archive (single)
curl -X POST "http://localhost:8080/api/entries/archive?id=1" \
  -H "Authorization: Bearer $TOKEN"

# 7. Archive (bulk / select all)
curl -X POST "http://localhost:8080/api/entries/archive" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ids":[1,2,3]}'

# 8. List archived
curl "http://localhost:8080/api/entries?archived=true"

# 9. Unarchive (bulk)
curl -X POST "http://localhost:8080/api/entries/unarchive" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ids":[1,2,3]}'

# 10. Delete permanently (bulk)
curl -X DELETE "http://localhost:8080/api/entries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ids":[1,2,3]}'

# 10. Categories CRUD
curl "http://localhost:8080/api/categories/list"
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Neuropsychology"}'

# 11. Logout
curl -X POST http://localhost:8080/api/logout
```

---

## Related docs

- [API Models & Mock Data](./api-models.md)
- [Deployment](./deployment.md)
- [Project README](../README.md)
