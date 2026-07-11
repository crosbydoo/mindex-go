# API Models & Mock Data

Request and response schemas with mock JSON examples for frontend integration and testing.

Base URL: `http://localhost:8080`

---

## Common response envelope

All endpoints return this shape:

| Field | Type | Description |
|-------|------|-------------|
| `code` | `number` | HTTP status code |
| `status` | `boolean` | `true` on success, `false` on error |
| `message` | `string` | Human readable message |
| `data` | `any \| null` | Payload on success, `null` on error |

### Success example

```json
{
  "code": 200,
  "status": true,
  "message": "Entries retrieved successfully",
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 0,
      "total_pages": 0,
      "has_next": false,
      "has_prev": false
    }
  }
}
```

### Error example

```json
{
  "code": 400,
  "status": false,
  "message": "Invalid entry payload",
  "data": null
}
```

### Pagination object

| Field | Type | Description |
|-------|------|-------------|
| `page` | `number` | Current page |
| `limit` | `number` | Items per page |
| `total` | `number` | Total matching records |
| `total_pages` | `number` | Total pages |
| `has_next` | `boolean` | `true` if another page exists after current |
| `has_prev` | `boolean` | `true` if a previous page exists |

---

## Enums

### `Category`

| Value |
|-------|
| `Clinical Psychology` |
| `Developmental Psychology` |
| `Cognitive Psychology` |
| `Social Psychology` |
| `Educational Psychology` |
| `Mental Health` |
| `Research Methods` |

### `EntryType`

| Value |
|-------|
| `Journal` |
| `Article` |
| `Thesis` |
| `Literature Review` |

---

## Health

### `HealthResponse`

| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | Always `"ok"` when server is healthy |

### Mock: success `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Server is healthy",
  "data": {
    "status": "ok"
  }
}
```

---

## Auth

### `LoginRequest`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `password` | `string` | yes | Admin password |

#### Mock: request

```json
{
  "password": "admin123"
}
```

### `LoginResponse`

| Field | Type | Description |
|-------|------|-------------|
| `token` | `string` | HMAC-SHA256 hex bearer token |

#### Mock: success `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Login successful",
  "data": {
    "token": "7f3a9c2e1b8d4f6a0e5c3b9d2a1f8e4c6b0d7a3f9e2c1b5d8a4f7e0c3b6a9d2f"
  }
}
```

#### Mock: error `401`

```json
{
  "code": 401,
  "status": false,
  "message": "Invalid password",
  "data": null
}
```

#### Mock: error `503`

```json
{
  "code": 503,
  "status": false,
  "message": "ADMIN_PASSWORD is not configured on the server",
  "data": null
}
```

### Logout

#### Mock: success `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Logout successful",
  "data": null
}
```

---

## Entry

### `Entry` (response model)

Returned by list/create/update/archive/unarchive endpoints.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `number` | Auto-generated ID |
| `title` | `string` | Entry title |
| `abstract` | `string` | Summary text |
| `category` | `Category` | Psychology category |
| `year` | `number` | Publication year |
| `author` | `string` | Author name |
| `source` | `string` | Journal or publisher |
| `type` | `EntryType` | Literature type |
| `url` | `string` | Source link (default `#`) |
| `is_archived` | `boolean` | `true` when archived |

> `created_at` / `archived_at` exist in DB but are **not** returned to clients.

#### Mock: single entry

```json
{
  "id": 1,
  "title": "Cognitive Behavioral Therapy for Depression: A Meta-Analysis",
  "abstract": "This meta-analysis examines the efficacy of cognitive behavioral therapy across 87 randomized controlled trials, finding moderate to large effect sizes for depression treatment.",
  "category": "Clinical Psychology",
  "year": 2023,
  "author": "Sarah Mitchell",
  "source": "Journal of Clinical Psychology",
  "type": "Journal",
  "url": "https://example.com/cbt-depression",
  "is_archived": false
}
```

#### Mock: list response `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Entries retrieved successfully",
  "data": {
    "items": [
      {
        "id": 4,
        "title": "Social Identity and Group Behavior in Online Communities",
        "abstract": "Explores how social identity theory explains collective behavior patterns in digital communities and social media platforms.",
        "category": "Social Psychology",
        "year": 2024,
        "author": "Michael Okafor",
        "source": "Social Psychology Quarterly",
        "type": "Article",
        "url": "https://example.com/social-identity",
        "is_archived": false
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 18,
      "total_pages": 2,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

#### Mock: empty list `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Entries retrieved successfully",
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 0,
      "total_pages": 0,
      "has_next": false,
      "has_prev": false
    }
  }
}
```

#### Mock: archive response `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Entry archived successfully",
  "data": {
    "id": 1,
    "title": "Cognitive Behavioral Therapy for Depression: A Meta-Analysis",
    "abstract": "This meta-analysis examines the efficacy of cognitive behavioral therapy across 87 randomized controlled trials, finding moderate to large effect sizes for depression treatment.",
    "category": "Clinical Psychology",
    "year": 2023,
    "author": "Sarah Mitchell",
    "source": "Journal of Clinical Psychology",
    "type": "Journal",
    "url": "https://example.com/cbt-depression",
    "is_archived": true
  }
}
```

---

### `EntryInput` (request model)

Used by `POST /api/entries` and `PUT /api/entries`. Same as `Entry` **without** `id`.

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `title` | `string` | yes | Non-empty after trim |
| `abstract` | `string` | yes | Non-empty after trim |
| `category` | `Category` | yes | Must match enum |
| `year` | `number` | yes | Positive integer (`> 0`) |
| `author` | `string` | yes | Non-empty after trim |
| `source` | `string` | yes | Non-empty after trim |
| `type` | `EntryType` | yes | Must match enum |
| `url` | `string` | no | Defaults to `"#"` if empty |

#### Mock: create request

```json
{
  "title": "Test Entry",
  "abstract": "Abstract text for a new psychology literature entry.",
  "category": "Clinical Psychology",
  "year": 2024,
  "author": "Jane Doe",
  "source": "Test Journal",
  "type": "Journal",
  "url": "https://example.com/article"
}
```

#### Mock: create request (minimal — url omitted)

```json
{
  "title": "Mindfulness in Schools",
  "abstract": "A review of mindfulness-based programs in K-12 education.",
  "category": "Educational Psychology",
  "year": 2022,
  "author": "Lisa Park",
  "source": "Educational Psychology Review",
  "type": "Literature Review"
}
```

Server normalizes missing `url` to `"#"`.

#### Mock: update request

```json
{
  "title": "Updated Entry Title",
  "abstract": "Updated abstract with revised findings.",
  "category": "Mental Health",
  "year": 2025,
  "author": "Jane Doe",
  "source": "Updated Journal",
  "type": "Article",
  "url": "https://example.com/updated"
}
```

#### Mock: create response `201`

```json
{
  "code": 201,
  "status": true,
  "message": "Entry created successfully",
  "data": {
    "id": 19,
    "title": "Test Entry",
    "abstract": "Abstract text for a new psychology literature entry.",
    "category": "Clinical Psychology",
    "year": 2024,
    "author": "Jane Doe",
    "source": "Test Journal",
    "type": "Journal",
    "url": "https://example.com/article",
    "is_archived": false
  }
}
```

#### Mock: update response `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Entry updated successfully",
  "data": {
    "id": 1,
    "title": "Updated Entry Title",
    "abstract": "Updated abstract with revised findings.",
    "category": "Mental Health",
    "year": 2025,
    "author": "Jane Doe",
    "source": "Updated Journal",
    "type": "Article",
    "url": "https://example.com/updated",
    "is_archived": false
  }
}
```

#### Mock: validation error `400`

```json
{
  "code": 400,
  "status": false,
  "message": "Invalid entry payload",
  "data": null
}
```

#### Mock: not found `404`

```json
{
  "code": 404,
  "status": false,
  "message": "Entry not found",
  "data": null
}
```

#### Mock: unauthorized `401`

```json
{
  "code": 401,
  "status": false,
  "message": "Unauthorized",
  "data": null
}
```

#### Mock: delete success `200`

```json
{
  "code": 200,
  "status": true,
  "message": "Entry deleted successfully",
  "data": null
}
```

---

## TypeScript interfaces (frontend reference)

```ts
export interface ApiResponse<T = unknown> {
  code: number;
  status: boolean;
  message: string;
  data: T | null;
}

export type Category =
  | 'Clinical Psychology'
  | 'Developmental Psychology'
  | 'Cognitive Psychology'
  | 'Social Psychology'
  | 'Educational Psychology'
  | 'Mental Health'
  | 'Research Methods';

export type EntryType =
  | 'Journal'
  | 'Article'
  | 'Thesis'
  | 'Literature Review';

export interface Entry {
  id: number;
  title: string;
  abstract: string;
  category: Category;
  year: number;
  author: string;
  source: string;
  type: EntryType;
  url: string;
  is_archived: boolean;
}

export type EntryInput = Omit<Entry, 'id' | 'is_archived'>;

export interface LoginRequest {
  password: string;
}

export interface LoginData {
  token: string;
}

export interface HealthData {
  status: 'ok';
}
```

---

## Mock JSON files

Ready-to-use files in `docs/mocks/`:

| File | Used for |
|------|----------|
| `login-request.json` | `POST /api/login` |
| `entry-create-request.json` | `POST /api/entries` |
| `entry-update-request.json` | `PUT /api/entries?id=1` |
| `entry-response.json` | Example single entry response |
| `entries-list-response.json` | Example list response |
| `entry-archive-response.json` | `POST /api/entries/archive` |
| `entry-unarchive-response.json` | `POST /api/entries/unarchive` |
| `entry-ids-request.json` | Bulk body `{ "ids": [1,2,3] }` |
| `error-unauthorized.json` | `401` example |
| `error-not-found.json` | `404` example |
| `error-invalid-payload.json` | `400` example |

---

## Related docs

- [API Endpoints](./api-endpoints.md)
- [Deployment](./deployment.md)
