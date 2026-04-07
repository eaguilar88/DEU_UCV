# Admin Dashboard API — Frontend Developer Guide

## Context
The admin dashboard allows reviewers to list and act on four types of requests:
- **Course requests** — a professor/group asking to create a new course
- **Group requests** — an extension group asking for authorization
- **Provider requests** — an individual or organization applying to become an approved provider
- **Course cycle close requests** — a request to close an active course cycle

All admin endpoints are prefixed with `/admin` and require a valid JWT Bearer token with admin-level roles.

---

## Authentication

Every request must include the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

Requests without a valid token return `401 Unauthorized`.

---

## Pagination

List endpoints accept these query parameters:

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number (1-based) |
| `per_page` | int | 10 | Results per page |

The response always includes a `pagination` object:
```json
{
  "pagination": {
    "page": 1,
    "per_page": 10,
    "count": 42
  }
}
```

`count` is the total number of matching records. Total pages = `Math.ceil(count / per_page)`.

---

## 1. Course Requests

Course requests are scoped to a faculty. The admin must pass the faculty they are reviewing for.

### List requests
```
GET /admin/course-requests?faculty=<faculty>&page=1&per_page=10
```

**Query params:**
- `faculty` (required) — faculty identifier string

**Response `200`:**
```json
{
  "requests": [
    {
      "id": "12",
      "status": "created",
      "comments": "",
      "reviewed_at": "",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "course": {
        "id": "5",
        "name": "Introducción a Python",
        "description": "...",
        "objectives": "...",
        "duration": "...",
        "content": "...",
        "type": "...",
        "faculty": "FCYT",
        "cost": "...",
        "location": "...",
        "is_active": false,
        "created_at": "2024-01-15T10:00:00Z",
        "updated_at": "2024-01-15T10:00:00Z"
      }
    }
  ],
  "pagination": { "page": 1, "per_page": 10, "count": 3 }
}
```

### Get single request
```
GET /admin/course-requests/:id
```

**Response `200`:** Same shape as one item in the list above.
**Response `404`:** Request not found.

### Approve
```
POST /admin/course-requests/:id/approve
Content-Type: application/json

{
  "tipo_curso": "diplomado",   // optional — course type classification
  "observaciones": "OK"        // optional — reviewer comments
}
```

**Response `200`:** `{}`
**Response `404`:** Request not found.

### Reject
```
POST /admin/course-requests/:id/reject
Content-Type: application/json

{
  "observaciones": "Incomplete documentation"   // optional
}
```

**Response `200`:** `{}`
**Response `404`:** Request not found.

### Redirect (send to another faculty)
```
POST /admin/course-requests/:id/redirect
Content-Type: application/json

{
  "facultad": "FCYT",       // required — target faculty
  "motivo": "Wrong faculty" // required — reason for redirect
}
```

**Response `200`:** `{}`
**Response `400`:** Missing `facultad` or `motivo`.
**Response `404`:** Request not found.

---

## 2. Group Requests

Also scoped to a faculty.

### List requests
```
GET /admin/group-requests?faculty=<faculty>&page=1&per_page=10
```

**Response `200`:**
```json
{
  "requests": [
    {
      "id": "7",
      "group_id": "3",
      "comments": "",
      "status": "created",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 10, "count": 1 }
}
```

### Get single request
```
GET /admin/group-requests/:id
```

**Response `200`:** Same shape as one item above.
**Response `404`:** Request not found.

### Approve
```
POST /admin/group-requests/:id/approve
```
No request body required.

**Response `200`:** `{}`
**Response `404`:** Request not found.

### Reject
```
POST /admin/group-requests/:id/reject
```
No request body required.

**Response `200`:** `{}`
**Response `404`:** Request not found.

---

## 3. Provider Requests

Not faculty-scoped — covers all providers across the system.

### List requests
```
GET /admin/provider-requests?page=1&per_page=10
```

**Response `200`:**
```json
{
  "requests": [
    {
      "id": "9",
      "provider_id": "4",
      "status": "created",
      "comments": "",
      "reviewed_at": "",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 10, "count": 2 }
}
```

### Get single request
```
GET /admin/provider-requests/:id
```

**Response `200`:** Same shape as one item above.
**Response `404`:** Request not found.

### Approve
```
POST /admin/provider-requests/:id/approve
```
No request body required. The backend generates the provider code automatically.

**Response `200`:** `{}`
**Response `404`:** Request not found.

### Reject
```
POST /admin/provider-requests/:id/reject
Content-Type: application/json

{
  "observaciones": "Missing credentials"   // optional
}
```

**Response `200`:** `{}`
**Response `404`:** Request not found.

---

## 4. Course Cycle Close Requests

Requests to close an active course cycle (submitted by instructors or group admins).

### List requests
```
GET /admin/course-cycle-close-requests?page=1&per_page=10
```

**Response `200`:**
```json
{
  "requests": [
    {
      "id": "2",
      "course_cycle_id": "11",
      "submitted_by_id": "5",
      "status": "created",
      "comments": "",
      "reviewed_at": "",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 10, "count": 1 }
}
```

### Get single request
```
GET /admin/course-cycle-close-requests/:id
```

**Response `200`:** Same shape as one item above.
**Response `404`:** Request not found.

### Approve
```
POST /admin/course-cycle-close-requests/:id/approve
```
No request body required.

**Response `200`:** `{}`
**Response `404`:** Request not found.

### Reject
```
POST /admin/course-cycle-close-requests/:id/reject
Content-Type: application/json

{
  "observaciones": "Cycle still has active participants"   // optional
}
```

**Response `200`:** `{}`
**Response `404`:** Request not found.

---

## Request Status Values

All request types share the same status lifecycle:

| Status | Meaning |
|--------|---------|
| `created` | Submitted, awaiting review |
| `under_review` | Being reviewed |
| `approved` | Approved by admin |
| `rejected` | Rejected by admin |
| `redirected` | Redirected to another faculty (course requests only) |

---

## Common HTTP Error Responses

| Code | When |
|------|------|
| `400` | Validation error — missing required field |
| `401` | Missing or invalid JWT token |
| `403` | Valid token but insufficient permissions |
| `404` | Resource not found |
| `409` | Duplicate entry conflict |
| `500` | Unexpected server error |

Error body shape:
```json
{
  "message": "course request not found"
}
```
