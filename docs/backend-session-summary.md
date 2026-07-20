# Backend — session summary

Everything below is **pushed** — both commits are on `origin/backend` and
merged into `origin/dev` (`a70f2d3`). This is a record of what was done and
why, not a to-do list.

---

## At a glance

| # | Module | Status |
|---|---|---|
| 1 | Gap analysis: running schema vs. `api-contract.md` v4 | Done — analysis only, no code |
| 2 | Project CRUD completion (Get/Update/Delete) | Done, tested |
| 3 | `namaProyek` uniqueness validation | Done, tested |
| 4 | Fixed `nama` wrongly required on Item | Done, tested |
| 5 | Pagination `limit` clamp (max 100) | Done, tested |
| 6 | `400 MALFORMED_BODY` handling | Done, tested |
| 7 | Test suite: fixed a latent collision bug, expanded coverage | Done |
| 8 | CORS middleware | Done, verified |
| 9 | `/api/inventory` compatibility layer | Done, verified — **temporary, flagged for removal** |

Two commits carry all of this: `4054174 feat: backend revamp` (modules
2–7) and `87d9e9b feat: cors + inventory compat api` (modules 8–9).

---

## 1. Gap analysis — running schema vs. the target contract

Before writing any code, read `backend/`, `CLAUDE.md`, and `docs/design/`
(`api-contract.md` v4, `knowledge-base.md`, the use-case doc), then compared
what was actually running against what the contract specifies.

**Finding**: the backend was already close — nested `credentials`/
`remoteInfo`, `idProyek` as a real FK, `customAttributes` auto-fold were all
in place from before this session (per earlier project history). The gap
wasn't "old flat schema vs. new," it was that the last implementation pass
predated a *newer* revision of the contract (v4) that expanded Project from
"reference only" to full CRUD, plus a handful of smaller compliance bugs.
Specifically found:

- Project had no `Get`/`Update`/`Delete` — only `Create`/`List`.
- `namaProyek` uniqueness was never actually enforced.
- Item's `nama` field was incorrectly required (contract says optional).
- `GET /items` didn't clamp `limit` to the documented max of 100.
- No code path anywhere returned `400 MALFORMED_BODY` — malformed/empty
  request bodies fell through to a `422` instead.
- Housekeeping: `CLAUDE.md` and `docs/design/api-contract.md` existed on
  disk under download-artifact names (`CLAUDE(1).md`,
  `api-contract-rti-items(1).md`) that didn't match the paths `CLAUDE.md`
  itself references — renamed to fix.

This analysis was presented and confirmed before any implementation
started.

---

## 2. Project CRUD completion

**Added**: `GET/PATCH/DELETE /api/v1/projects/{id}`.

- `GetProject` — straightforward lookup, `404` if missing.
- `UpdateProject` — partial update (`namaProyek`, `lokasi`), re-validates
  non-blank + uniqueness (excluding itself) on every update.
- `DeleteProject` — **not blocked** by referencing items, per contract;
  those items become orphaned (`idProyek` pointing at nothing), no cascade,
  no re-validation. This matches the contract's explicit design decision,
  confirmed via `knowledge-base.md` decision #11.

**Files**: `internal/handlers/projects.go`, `internal/repository/projects.go`
(added `Update`, `Delete`, `ExistsByNamaProyek`), `internal/handlers/router.go`.

---

## 3. `namaProyek` uniqueness validation

Case-sensitive, exact-match uniqueness, enforced on both `POST /projects`
(create) and `PATCH /projects/{id}` (update, excluding the project's own
document from the check — so re-submitting a project's own unchanged name
doesn't false-positive as a duplicate).

**Files**: `internal/repository/projects.go` (`ExistsByNamaProyek`),
`internal/handlers/projects.go`.

**Side effect caught during testing**: the test suite's `createProject`
helper hardcoded `namaProyek: "ALPHA"` as its default — harmless before
uniqueness existed, but once enforced, several existing tests that called
it multiple times against the same shared test database started
colliding and failing. Fixed by making the helper auto-generate a unique
name per call (`ALPHA-1`, `ALPHA-2`, ...).

---

## 4. Fixed `nama` wrongly required on Item

`ItemHandler.validate()` had `nama` in its required-fields map, contradicting
both `CLAUDE.md` ("nama OPSIONAL") and the contract's field table. One-line
fix, but the existing test (`TestItemValidation`) was actually *asserting*
the wrong behavior as correct — updated that test, and added new coverage
proving create/update succeed without `nama`.

---

## 5. Pagination `limit` clamp

`repository.ItemRepository.List` clamped the floor (`limit < 1 → 20`) but
never the ceiling. Added the missing `limit > 100 → 100` clamp, matching
`CLAUDE.md`'s explicit rule ("maksimum 100, clamp bukan reject").

---

## 6. `400 MALFORMED_BODY` handling

Added a shared `decodeJSONObject` helper (`internal/handlers/decode.go`)
used by every mutating endpoint (`CreateItem`, `UpdateItem`, `CreateProject`,
`UpdateProject`). Distinguishes three cases uniformly, matching the
contract's error table:
- Empty body → `400`.
- Invalid JSON → `400`.
- Valid JSON that isn't an object (array, bare string/number) → `400`.
- Valid-but-empty JSON object (`{}`) → **not** malformed — falls through to
  normal required-field validation (`422`), which is the correct
  distinction per the contract.

Previously, all three malformed cases silently fell through to `422`
because the decode error was discarded (`_ = json.NewDecoder(...).Decode(...)`).

---

## 7. Test suite

Extended `backend/test/api_test.go` substantially alongside the above —
every new/fixed behavior (Project CRUD, uniqueness, `nama` optionality,
limit clamp, malformed body) has explicit test coverage, run against a real
MongoDB instance (not mocked), following the existing black-box HTTP-level
test pattern already in place. All green (`go build`, `go vet`, `go test`)
at every step.

---

## 8. CORS middleware

The real API router (`internal/handlers/router.go`) had **zero** CORS
handling — the only CORS code in the whole repo was in the unrelated legacy
`api/hello.go` Vercel-stub handler, which isn't part of this API at all.
Added `github.com/go-chi/cors` with a permissive policy (`*` origin, GET/
POST/PATCH/DELETE/OPTIONS). Needed because the frontend runs on a different
origin/port than the backend — without this, no cross-origin request from
a browser could succeed regardless of correctness elsewhere.

**Verified**: curled an `OPTIONS` preflight and confirmed the
`Access-Control-Allow-*` headers come back correctly.

---

## 9. `/api/inventory` compatibility layer

**This is explicitly temporary, not part of `api-contract.md`.** Added
`GET /api/inventory` and `GET /api/inventory/filter-options` — a read-only
view that translates the real Item/Project storage shape into the flat,
snake_case shape the frontend's dashboard was already written against
(before this session's frontend work, it was mocked and never actually
reachable). Built purely so the frontend could show real data without
frontend code changes being in scope at the time.

Specifics:
- Translates `nama→name`, `serialNumber→serial_number`,
  `licenseWindows→license_windows`, etc.; resolves `idProyek` to the
  project's `namaProyek` for display; flattens `credentials`/`remoteInfo`
  to plain string maps; `customAttributes` becomes `other`.
- Supports **multi-select OR** filtering on project name and `jenis`
  (comma-separated) — a different semantic from `/api/v1/items`'s
  single-value equality filters, so it has its own repository method
  (`ItemRepository.ListCompat`) rather than reusing `ListParams`.
- Does **not** touch or replace `/api/v1/items` / `/api/v1/projects` —
  fully additive.

**Files**: `internal/handlers/inventory_compat.go` (new),
`internal/repository/items.go` (`ListCompat`, `CompatListParams`).

**Status**: flagged in `docs/design/frontend-todo.md` (handed to the
frontend side) as a shim to migrate off of. Once the frontend moves its
list/filter reads onto `/api/v1/items` directly, this entire file and its
route should be deleted — there's no reason to maintain two read paths for
the same data long-term.

---

## Full current endpoint list

```
POST/GET     /api/v1/items
GET/PATCH/DELETE  /api/v1/items/{id}
GET          /api/v1/items/filter-options, /filter-options/jenis
GET          /api/v1/items/stats           (kept as a real feature, not a stub — your call earlier)
POST/GET     /api/v1/items/import, /export (stubs, contract says out of scope)
POST/GET     /api/v1/projects
GET/PATCH/DELETE  /api/v1/projects/{id}
GET          /api/inventory, /api/inventory/filter-options   (temporary, see §9)
```

## Not covered here

Frontend integration work (proxy config, wiring the dashboard/item form to
these endpoints) is a separate, frontend-focused document:
`docs/wip-branch-summary.md` (on the local `dev-frontend-wip` branch, not
pushed).
