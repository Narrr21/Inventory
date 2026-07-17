# API Contract

Base path: `/api/v1`. All responses use the envelope:

Success: `{ "success": true, "data": ..., "meta"?: ... }`
Success with partial errors: `{ "success": true, "data": ..., "errors": [...] }`
Error: `{ "success": false, "error": { "code": "...", "message": "...", "fields"?: {...} } }`

Every request is bounded by a 10s server-side timeout (`middleware.Timeout`) — a request that
doesn't complete in time gets a `504` from chi, independent of any application logic.

---

## Items — `/api/v1/items`

### Item shape

```json
{
  "_id": "665f1a1a1a1a1a1a1a1a1a1a",
  "jenis": "Laptop",
  "serialNumber": "SN-00123",
  "nama": "RTI-ALPHA-001",
  "idProyek": "665f1a1a1a1a1a1a1a1a1a99",
  "credentials": { "account": "user01", "passwordAccount": "secret01", "passwordPin": "1234" },
  "remoteInfo": { "ipAddress": "10.0.0.12", "anydesk": "123 456 789", "rustdesk": "", "passwordRemote": "rdpass01" },
  "licenseWindows": "Pro",
  "licenseOffice": "365",
  "status": "Healthy",
  "deskripsi": "Contoh data seed",
  "customAttributes": { "warranty": "3 tahun", "garansiTahun": 3 },
  "createdAt": "2026-07-17T12:20:35.684Z",
  "updatedAt": "2026-07-17T12:20:35.684Z"
}
```

`idProyek` is a foreign key into `projects._id` (see below) — `lokasi` is **not** an item field, it
belongs to the referenced project.

`customAttributes` is schemaless (`Map<String, any>`) — any request field that isn't one of the
fixed fields above is automatically folded into `customAttributes` instead of being rejected or
silently dropped. This is the sanctioned mechanism for ad-hoc columns on newly-inventoried item
types; see "Ad-hoc fields" below.

### `POST /items`

Body: any subset of item fields, plus optionally any number of unrecognized fields (see "Ad-hoc
fields" below).

```json
{ "jenis": "Printer", "serialNumber": "SN-00200", "nama": "RTI-DELTA-009", "idProyek": "665f...a99", "status": "Healthy" }
```

Required: `jenis`, `serialNumber`, `nama`, `status`, `idProyek`. `idProyek` must reference an
existing project.

`201`
```json
{ "success": true, "data": { "_id": "665f...a99", "jenis": "Printer", "...": "..." } }
```

`422` — one or more required fields missing/blank, or `idProyek` doesn't reference a real project:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid item fields",
    "fields": { "idProyek": "referenced project does not exist" }
  }
}
```

`500` — write failed:
```json
{ "success": false, "error": { "code": "INTERNAL_ERROR", "message": "Failed to create item" } }
```

### `GET /items`

Query params:

| Param | Behavior |
|---|---|
| `page`, `limit` | pagination (default `page=1`, `limit=20`) |
| `sortBy` | one of `nama, createdAt, updatedAt, status, idProyek, jenis`, or `customAttributes.<key>` |
| `sortOrder` | `asc` or `desc` (default `desc`) |
| `jenis, serialNumber, nama, idProyek, status, licenseWindows, licenseOffice` | exact-match filters |
| `credentials.account, remoteInfo.ipAddress, remoteInfo.anydesk, remoteInfo.rustdesk` | exact-match filters on non-secret nested fields |
| `customAttributes.<key>` | exact-match filter on an ad-hoc attribute, e.g. `?customAttributes.warna=Merah` |
| `q` | case-insensitive substring search over `nama` and `serialNumber` |
| `createdFrom`, `createdTo`, `updatedFrom`, `updatedTo` | ISO-8601 range filter on `createdAt`/`updatedAt` (`$gte`/`$lte`) |

Password-bearing fields (`credentials.passwordAccount`, `credentials.passwordPin`,
`remoteInfo.passwordRemote`) are not filterable, and free-text search doesn't reach into
`customAttributes` (arbitrary keys can't be searched without knowing them in advance).

`200`
```json
{
  "success": true,
  "data": [
    {
      "_id": "665f...a1a", "jenis": "Laptop", "serialNumber": "SN-00123", "nama": "RTI-ALPHA-001",
      "idProyek": "665f...a99",
      "credentials": { "account": "user01" },
      "remoteInfo": { "ipAddress": "10.0.0.12", "anydesk": "123 456 789", "rustdesk": "" },
      "licenseWindows": "Pro", "licenseOffice": "365", "status": "Healthy",
      "deskripsi": "Contoh data dummy", "customAttributes": { "warranty": "3 tahun" },
      "createdAt": "...", "updatedAt": "..."
    }
  ],
  "meta": { "total": 4, "page": 1, "limit": 20, "totalPages": 1 }
}
```
Note: `credentials.passwordAccount`, `credentials.passwordPin`, and `remoteInfo.passwordRemote` are
omitted here — only `GET /items/{id}` returns them. No matches → `"data": []`, `meta.total: 0` (not
an error).

### `GET /items/{id}`

Returns the item with the given `{id}`, including secret fields.

`200`
```json
{ "success": true, "data": { "_id": "665f...a1a", "credentials": { "account": "user01", "passwordAccount": "secret01", "passwordPin": "1234" }, "...": "..." } }
```

Unknown `{id}` → `404`
```json
{ "success": false, "error": { "code": "NOT_FOUND", "message": "Item not found" } }
```

### `PATCH /items/{id}`

Body: any subset of fields to update, plus optionally any number of unrecognized fields (see
"Ad-hoc fields" below).

```json
{ "status": "Broken" }
```

`200` — existing item merged with the body:
```json
{ "success": true, "data": { "_id": "665f...a1a", "...": "...", "status": "Broken" } }
```

Validation runs against the *result* of applying the patch (not just the patch delta) — clearing a
required field to `""` fails validation the same way omitting it on create does. `404` if `{id}`
doesn't exist, `422` on validation failure (same shape as `POST`), `500` on write failure.

### `DELETE /items/{id}`

`200`
```json
{ "success": true, "data": { "_id": "507f191e810c19729de860ea" } }
```
`404` if `{id}` doesn't exist. A successful delete is the client's signal to refresh its item list.

### Ad-hoc fields (dynamic columns)

Any top-level field in a `POST`/`PATCH` body that isn't one of the fixed `Item` fields is **never**
rejected and **never** silently dropped — it's automatically folded into `customAttributes`:

```
POST /items  { "jenis": "Printer", "serialNumber": "SN-1", "nama": "X", "status": "Healthy", "idProyek": "...", "garansiTahun": 3 }
→ stored as customAttributes: { "garansiTahun": 3 }
```

On `PATCH`, ad-hoc fields are written via a dotted `$set` path (`customAttributes.<key>`), not by
replacing the whole `customAttributes` object — so two separate PATCHes adding different ad-hoc
attributes both survive, rather than the second overwriting the first. `credentials` and
`remoteInfo`, by contrast, are still replaced wholesale on `PATCH` (send the full nested object, not
a delta) — they're expected to come from a form that always submits the complete sub-object,
whereas ad-hoc attributes are expected to accumulate one at a time.

There is currently no way to delete a single `customAttributes` key via the API (no `$unset`
support) — only add or overwrite one.

### `GET /items/filter-options`

`200`
```json
{ "success": true, "data": { "jenis": ["Laptop", "Monitor", "PC", "Server"], "status": ["Healthy", "Under Maintenance", "Broken"] } }
```

### `GET /items/filter-options/jenis`

Distinct, non-empty `jenis` values.

`200`
```json
{ "success": true, "data": ["Laptop", "Monitor", "PC", "Server"] }
```

For project options (to populate an `idProyek` dropdown), use `GET /projects` instead — it returns
richer data (`namaProyek`, `lokasi`) than a bare distinct-values list would.

### `GET /items/stats`

Out of scope per the RTI spec (F05) — kept as-is from before the overhaul, field names updated to
match the new schema. Not guaranteed to stay stable; don't build against it.

`200`
```json
{
  "success": true,
  "data": {
    "totalItems": 143,
    "byStatus": [ { "_id": "Healthy", "count": 120 }, { "_id": "Broken", "count": 15 } ],
    "byJenis": [ { "_id": "Laptop", "count": 80 }, { "_id": "PC", "count": 40 } ],
    "byProyek": [ { "_id": "665f...a99", "count": 30 } ],
    "recentlyAdded": [ { "...5 items, secret fields omitted...": "" } ]
  }
}
```

### `POST /items/import`, `GET /items/export`

Out of scope per the RTI spec (F09/F10). Stubs only — `import` accepts anything and reports zero
inserted/updated/failed; `export` always responds `501`.

---

## Projects — `/api/v1/projects`

Minimal reference/lookup entity for `idProyek` — intentionally **no** update or delete endpoint; the
RTI spec defines no use case for editing or removing a project, only for items to reference one.

### Project shape

```json
{ "_id": "665f1a1a1a1a1a1a1a1a1a99", "namaProyek": "ALPHA", "lokasi": "Jakarta HQ" }
```

### `POST /projects`

```json
{ "namaProyek": "DELTA", "lokasi": "Bandung Office" }
```

`201`
```json
{ "success": true, "data": { "_id": "665f...a99", "namaProyek": "DELTA", "lokasi": "Bandung Office" } }
```

`422` if `namaProyek` is blank:
```json
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "Invalid project fields", "fields": { "namaProyek": "namaProyek is required" } } }
```

### `GET /projects`

Full list, sorted by `namaProyek`. Not paginated — it's a lookup source for populating an `idProyek`
dropdown, not a browsable resource.

`200`
```json
{ "success": true, "data": [ { "_id": "665f...a99", "namaProyek": "ALPHA", "lokasi": "Jakarta HQ" } ] }
```

There is currently no way to filter items directly by location — `lokasi` lives only on `Project`.
Filter items by `idProyek` and cross-reference `GET /projects` for location, or ask for a `$lookup`
join to be added if per-item location filtering turns out to be needed.
