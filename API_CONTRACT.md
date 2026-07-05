# API Contract — Phase 1 (Stub)

Base path: `/api/v1`. All responses use the envelope:

Success: `{ "success": true, "data": ..., "meta"?: ... }`
Error: `{ "success": false, "error": { "code": "...", "message": "...", "fields"?: {...} } }`

---

### `POST /items`
Body: any subset of item fields.
```json
{ "jenisProduct": "Printer", "name": "RTI-DELTA-009", "status": "Active", "lokasi": "Bandung Office" }
```
`201`
```json
{ "success": true, "data": { "_id": "665f1a1a1a1a1a1a1a1a1a99", "jenisProduct": "Printer", "serialNumber": "SN-00123", "name": "RTI-DELTA-009", "proyek": "ALPHA", "passwordPin": "1234", "account": "user01", "passwordAccount": "secret01", "ipAddress": "10.0.0.12", "anydesk": "123 456 789", "rustdesk": "", "passwordAnydeskRustdesk": "rdpass01", "licenseWindows": "Pro", "licenseOffice": "365", "status": "Active", "lokasi": "Bandung Office", "deskripsi": "Contoh data dummy", "createdAt": "2026-07-05T05:32:51.470Z", "updatedAt": "2026-07-05T05:32:51.470Z" } }
```

### `GET /items`
Query: `page, limit, sortBy, sortOrder, q, jenisProduct, proyek, status, lokasi, serialNumber, name, account, ipAddress, anydesk, rustdesk, licenseWindows, licenseOffice, createdAtFrom, createdAtTo, updatedAtFrom, updatedAtTo` — accepted but ignored in this phase.
`200`
```json
{
  "success": true,
  "data": [ { "_id": "...", "jenisProduct": "Laptop", "serialNumber": "SN-00123", "name": "RTI-ALPHA-001", "proyek": "ALPHA", "account": "user01", "ipAddress": "10.0.0.12", "anydesk": "123 456 789", "rustdesk": "", "licenseWindows": "Pro", "licenseOffice": "365", "status": "Active", "lokasi": "Jakarta HQ", "deskripsi": "Contoh data dummy", "createdAt": "...", "updatedAt": "..." } ],
  "meta": { "total": 4, "page": 1, "limit": 20, "totalPages": 1 }
}
```
Note: sensitive fields (`passwordPin`, `passwordAccount`, `passwordAnydeskRustdesk`) are omitted here.

### `GET /items/{id}`
Returns the mock item regardless of `{id}`, including sensitive fields.
`200`
```json
{ "success": true, "data": { "_id": "665f1a1a1a1a1a1a1a1a1a1a", "jenisProduct": "Laptop", "serialNumber": "SN-00123", "name": "RTI-ALPHA-001", "proyek": "ALPHA", "passwordPin": "1234", "account": "user01", "passwordAccount": "secret01", "ipAddress": "10.0.0.12", "anydesk": "123 456 789", "rustdesk": "", "passwordAnydeskRustdesk": "rdpass01", "licenseWindows": "Pro", "licenseOffice": "365", "status": "Active", "lokasi": "Jakarta HQ", "deskripsi": "Contoh data dummy", "createdAt": "...", "updatedAt": "..." } }
```
`{id} = "notfound"` → `404`
```json
{ "success": false, "error": { "code": "NOT_FOUND", "message": "Item not found" } }
```

### `PATCH /items/{id}`
Body: any subset of fields to update.
```json
{ "status": "Rusak", "lokasi": "Gudang" }
```
`200` — mock item merged with the body:
```json
{ "success": true, "data": { "_id": "665f1a1a1a1a1a1a1a1a1a1a", "...": "...", "status": "Rusak", "lokasi": "Gudang" } }
```

### `DELETE /items/{id}`
`200`
```json
{ "success": true, "data": { "_id": "507f191e810c19729de860ea" } }
```

### `GET /items/filter-options`
`200`
```json
{ "success": true, "data": { "proyek": ["ALPHA", "BETA", "GAMMA"], "jenisProduct": ["Laptop", "PC", "Monitor"], "lokasi": ["Jakarta HQ", "Surabaya Branch"], "status": ["Active", "Idle", "Maintenance"] } }
```

### `GET /items/stats`
`200`
```json
{
  "success": true,
  "data": {
    "totalItems": 143,
    "byStatus": [ { "_id": "Active", "count": 120 }, { "_id": "Maintenance", "count": 15 } ],
    "byJenisProduct": [ { "_id": "Laptop", "count": 80 }, { "_id": "PC", "count": 40 } ],
    "byProyek": [ { "_id": "ALPHA", "count": 30 } ],
    "recentlyAdded": [ { "...4 mock items, sensitive fields omitted..." } ]
  }
}
```

### `POST /items/import` (multipart, field `file` — not parsed in this phase)
`200`
```json
{ "success": true, "data": { "inserted": 0, "updated": 0, "failed": 0 }, "errors": [] }
```

### `GET /items/export`
`501`
```json
{ "success": false, "error": { "code": "NOT_IMPLEMENTED", "message": "Export is not implemented in this phase" } }
```
