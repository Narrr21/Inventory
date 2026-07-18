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

**Headers:** `Content-Type: application/json` is expected but not actually enforced by the server —
the handler decodes whatever bytes are in the body as JSON regardless of the header. Sending a
non-JSON body or omitting the header entirely doesn't get you a `415`; it gets you a decode failure
that's swallowed (see below), which then surfaces as a `422` for missing required fields.

**Body:** a flat JSON object. Every field is optional at the JSON-decoding level; required-ness is
enforced by validation, not by the shape of the request. Recognized top-level fields:

| Field | Type | Notes |
|---|---|---|
| `jenis` | string | **required**, non-blank after trim |
| `serialNumber` | string | **required**, non-blank after trim |
| `nama` | string | **required**, non-blank after trim |
| `idProyek` | string | **required**, non-blank, and must be the `_id` of an existing project |
| `status` | string | **required**, non-blank after trim |
| `credentials` | object `{ account?, passwordAccount?, passwordPin? }` | optional, all sub-fields strings |
| `remoteInfo` | object `{ ipAddress?, anydesk?, rustdesk?, passwordRemote? }` | optional, all sub-fields strings |
| `licenseWindows` | string | optional |
| `licenseOffice` | string | optional |
| `deskripsi` | string | optional |
| `customAttributes` | object `Map<string, any>` | optional, merged with any ad-hoc top-level fields (see below) |

Any other top-level key you send is **not** one of the above — it's schemaless and gets folded into
`customAttributes` automatically (see "Ad-hoc fields"). `_id`, `createdAt`, `updatedAt` are ignored if
present in the body — they're always server-assigned on create.

A minimal valid request only needs the five required fields:

```json
{ "jenis": "Printer", "serialNumber": "SN-00200", "nama": "RTI-DELTA-009", "idProyek": "665f...a99", "status": "Healthy" }
```

A fuller request exercising every field, including an ad-hoc one:

```json
{
  "jenis": "Printer",
  "serialNumber": "SN-00200",
  "nama": "RTI-DELTA-009",
  "idProyek": "665f...a99",
  "status": "Healthy",
  "credentials": { "account": "user02", "passwordAccount": "secret02", "passwordPin": "5678" },
  "remoteInfo": { "ipAddress": "10.0.0.13", "anydesk": "987 654 321", "rustdesk": "", "passwordRemote": "rdpass02" },
  "licenseWindows": "Pro",
  "licenseOffice": "365",
  "deskripsi": "New printer for Delta floor",
  "garansiTahun": 3
}
```

`garansiTahun` isn't a known field, so it lands in the stored document as `customAttributes: { "garansiTahun": 3 }`.

**Malformed or empty body:** an empty body, invalid JSON, or a JSON value that isn't an object (e.g.
a bare array or string) all fail to decode. The decode error is discarded rather than returned to the
client — the request proceeds as if an empty object `{}` had been sent, which then fails validation
with all five required fields reported missing (`422`), rather than a `400 Bad Request`.

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

No request body. All input is via query string; every param is optional.

| Param | Behavior |
|---|---|
| `page` | 1-indexed page number. Anything that doesn't parse as an integer (missing, empty, non-numeric, negative, `0`) **silently falls back to `1`** — there's no `400` for a bad value. |
| `limit` | page size. Same silent-fallback rule as `page`, defaulting to `20`. There is **no upper bound/clamp** — `?limit=100000` is honored as-is and will hit Mongo with that page size. |
| `sortBy` | one of `nama`, `createdAt`, `updatedAt`, `status`, `idProyek`, `jenis`, or any `customAttributes.<key>`. An unrecognized value (typo, unrelated field name) is **not rejected** — it silently falls back to sorting by `createdAt`. |
| `sortOrder` | `asc` or `desc`. Anything other than exactly `"asc"` (including empty/missing) is treated as `desc` — there's no validation error for a bad value. |
| `jenis`, `serialNumber`, `nama`, `idProyek`, `status`, `licenseWindows`, `licenseOffice` | exact-match (case-sensitive) filters; combined with AND if several are given |
| `credentials.account`, `remoteInfo.ipAddress`, `remoteInfo.anydesk`, `remoteInfo.rustdesk` | exact-match filters on non-secret nested fields, e.g. `?remoteInfo.ipAddress=10.0.0.12` |
| `customAttributes.<key>` | exact-match filter on an ad-hoc attribute, e.g. `?customAttributes.warna=Merah`. The key is passed straight through unvalidated — a typo'd key just matches nothing rather than erroring |
| `q` | case-insensitive **substring** (regex) search, OR'd across `nama` and `serialNumber`. Combined with the exact-match filters above via AND (e.g. `?status=Healthy&q=alpha` = must be Healthy AND match "alpha") |
| `createdFrom`, `createdTo`, `updatedFrom`, `updatedTo` | ISO-8601 datetime strings, compared as strings (`$gte`/`$lte`) against `createdAt`/`updatedAt`. Only meaningful with full ISO-8601 timestamps (`2026-07-01T00:00:00.000Z`) since the comparison is lexicographic, not a real date comparison — a malformed date string is passed to Mongo as-is and will just fail to match anything sensible rather than erroring |

Example combining several params:
```
GET /items?jenis=Laptop&status=Healthy&q=alpha&sortBy=nama&sortOrder=asc&page=2&limit=10
```

Password-bearing fields (`credentials.passwordAccount`, `credentials.passwordPin`,
`remoteInfo.passwordRemote`) are not filterable — they're excluded from the allowed-filter list
regardless of what's in the query string. Free-text search (`q`) doesn't reach into
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

**Headers:** same as `POST /items` — `Content-Type: application/json` is expected but not enforced.

**Body:** a flat JSON object containing only the fields you want to change — this is a true partial
update, unlike `POST` where the required fields must all be present. Same field table as `POST
/items` applies to what's recognized vs. folded into `customAttributes`, with two semantic
differences to be aware of per field:

| Field(s) | PATCH semantics |
|---|---|
| `jenis`, `serialNumber`, `nama`, `idProyek`, `status`, `licenseWindows`, `licenseOffice`, `deskripsi` | scalar overwrite — only the fields present in the body are changed |
| `credentials`, `remoteInfo` | **whole-object replace** — send the complete sub-object, not a delta. `{ "credentials": { "account": "newuser" } }` replaces the *entire* `credentials` object, so `passwordAccount`/`passwordPin` on the existing item are wiped unless you re-send them too |
| unrecognized top-level keys / `customAttributes.<key>` | **additive `$set`** by dotted path — each ad-hoc key is written independently, so this PATCH won't clobber ad-hoc keys set by a previous PATCH |

Simplest case, a single scalar field:
```json
{ "status": "Broken" }
```

Replacing nested credentials wholesale (must include every sub-field you want to keep):
```json
{ "credentials": { "account": "user01", "passwordAccount": "newsecret01", "passwordPin": "1234" } }
```

Adding/overwriting one ad-hoc attribute without touching others already on the item:
```json
{ "garansiTahun": 5 }
```

Sending `_id` in the body has no effect — it's stripped before the update is applied, so you can't
change or spoof an item's ID via `PATCH`.

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

**Headers:** same caveat as items — `Content-Type: application/json` expected, not enforced.

**Body:**

| Field | Type | Notes |
|---|---|---|
| `namaProyek` | string | **required**, non-blank after trim |
| `lokasi` | string | optional — no validation, can be omitted or blank |

Unlike items, there is **no ad-hoc/`customAttributes` mechanism** for projects — `Project` decodes
strictly into its two fields, so any other top-level key in the body is silently ignored (not
stored, not rejected).

```json
{ "namaProyek": "DELTA", "lokasi": "Bandung Office" }
```

An empty or malformed body decodes to a zero-value `Project{}` (both fields `""`), which then fails
validation on `namaProyek` the same way an explicit `{}` would — no `400`, only a `422`.

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
