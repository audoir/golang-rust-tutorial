# Project Context — Golang / Rust API Tutorial

_Last updated: 2026-09-22 (post-concurrency/testing/packaging pass). This file is a working-context dump for AI coding agents / contributors picking up this repo. It is not linked from the public README and can be deleted or updated freely as the project evolves._

---

## What this project is

A hands-on tutorial repo for developers who already know **TypeScript** and/or **Python** and want to learn systems languages (starting with **Go**, with **Rust** planned as a future chapter) by building the *same simple CRUD API* in each language.

- Each "chapter" (one language) lives in its own sub-folder with its own backend server on its own port.
- All chapters are demonstrated through a **single shared Next.js frontend** with a tab-navigation UI — one tab per chapter/language.
- Documentation for each chapter lives in **`docs/<language>/`** (a folder, not a single file) — split into multiple progressively-more-advanced **parts** (basics → concurrency → testing → API), rather than one giant file. See "docs/go/ structure" below.

**Origin note:** This repo started as a fork/copy of a larger multi-framework tutorial project (which had tabs for Next.js, Express, NestJS, Flask, FastAPI, and Go). Everything except Go was stripped out in a cleanup pass so this repo could stand alone as "Chapter 1: Go" of a new Golang/Rust-focused tutorial. If you see any stray references to Express/NestJS/Flask/FastAPI/Zod/Pydantic anywhere, they are leftover cruft that should be removed — the intent is for this repo to only ever talk about Go and (eventually) Rust.

**Restructure note (this pass):** Chapter 1 was substantially expanded at the user's request to be more code-comprehensive: real goroutines/WaitGroup/channels (not just described in prose), a compiled-binary + graceful-shutdown run model instead of `go run`, and a full test suite (testify + httptest). This required splitting the Go backend from one flat `package main` into a `cmd/` + `internal/` layout (see "Go backend architecture notes" below) and splitting `docs/go.md` into four files under `docs/go/`. This same shape (`docs/<lang>/{README,01-basics,02-concurrency,03-testing,04-api}.md` + `<lang>-api/{cmd,internal}`) is the intended template for the future Rust chapter too.

---

## Current state (as of last commit)

- Git: `main` branch. Chapter 1 (Go) was substantially reworked in this pass (see "Restructure note" above) — check `git status`/`git log` for the latest commit state before assuming anything below is uncommitted.
- Only **Chapter 1 — Go** exists right now. No Rust chapter yet.
- Full build/lint/runtime validation has been done and passes:
  - `npm run build` — succeeds (Next.js/Turbopack)
  - `npm run lint` — clean, no errors
  - `go build ./...` + `go vet ./...` + `go test ./...` + `go test -race ./...` in `go-api/` — all succeed
  - The compiled binary (`go build -o bin/go-api ./cmd/go-api && ./bin/go-api`) starts, serves requests, and shuts down gracefully on SIGTERM/Ctrl-C
  - The standalone concurrency demo (`go run ./cmd/concurrency-demo`) runs and prints timing proving concurrent execution
  - Both dev servers (`npm run dev` on :3000, the built Go binary on :3005 via `scripts/start-servers.sh`) start and respond correctly together.

---

## Repo file structure

```
golang-rust-tutorial/
├── app/                          # Next.js App Router frontend (shared UI shell)
│   ├── components/
│   │   ├── GoCrud.tsx            # Chapter 1 — Go CRUD demo (client component, hits GO_API on :3005)
│   │   ├── PageHeader.tsx        # Site header: "🐹 Golang / Rust API Tutorial"
│   │   └── TabNavigation.tsx     # Tab bar; MainTab type + TABS array (currently only "go-crud")
│   ├── favicon.ico
│   ├── globals.css               # Tailwind v4 import + CSS variables
│   ├── layout.tsx                # Root layout; title "Golang / Rust API Tutorial"
│   └── page.tsx                  # Home page: renders PageHeader + TabNavigation + GoCrud
├── go-api/                       # Chapter 1 backend — Go REST API, port 3005
│   ├── cmd/
│   │   ├── go-api/main.go        # Real server entrypoint: -port flag, router, graceful shutdown (SIGINT/SIGTERM)
│   │   └── concurrency-demo/main.go  # Standalone goroutines/WaitGroup/channel demo, no HTTP involved
│   ├── internal/
│   │   ├── items/                # Item struct, Store (mutex-guarded map), CreateItemInput/UpdateItemInput/PatchItemInput + validation
│   │   │   ├── item.go, store.go, store_test.go, schemas.go, schemas_test.go
│   │   ├── concurrency/          # FanOut(checks) []Result — goroutines + sync.WaitGroup + buffered channel
│   │   │   ├── fanout.go, fanout_test.go
│   │   └── api/                  # chi router + HTTP handlers, composes items + concurrency
│   │       ├── router.go, items_handlers.go, enrich_handler.go, json.go
│   │       ├── items_handlers_test.go, enrich_handler_test.go
│   ├── go.mod / go.sum           # Go module: chi/v5, go-chi/cors, go-playground/validator/v10, stretchr/testify
│   ├── .gitignore                # ignores /bin (compiled binary output)
│   └── README.md                 # Self-contained Go API README (setup/run/build/test with curl)
├── docs/
│   └── go/                       # THE teaching docs for Chapter 1, split into 4 progressive parts (see below)
│       ├── README.md             # Chapter 1 index — links to all 4 parts
│       ├── 01-basics.md          # Part 1 — Go language basics
│       ├── 02-concurrency.md     # Part 2 — goroutines, WaitGroup, channels (new)
│       ├── 03-testing.md         # Part 3 — testing package, testify, httptest (new)
│       └── 04-api.md             # Part 4 — the CRUD API (chi, validator, enrich endpoint, graceful shutdown)
├── scripts/
│   └── start-servers.sh          # Starts npm run dev (3000) + builds & runs go-api binary (3005), Ctrl-C kills both
├── public/                       # Default Next.js SVG assets (unused/untouched)
├── README.md                     # Top-level project README (see structure below)
├── package.json / package-lock.json   # Next.js app deps (next, react, react-dom, tailwind, eslint, typescript)
├── next.config.ts, tsconfig.json, eslint.config.mjs, postcss.config.mjs  # Standard Next.js config, untouched defaults
└── CONTEXT.md                    # This file
```

**Note:** There is no `app/lib/` anymore (Zod schemas were removed along with the other framework tabs). `GoCrud.tsx` does its own lightweight client-side validation in plain TypeScript functions (`validateName`, `validateDescription`) that mirror the Go server's `go-playground/validator` rules — no schema library is used on the frontend.

---

## Ports

| Service | Port | Start command |
|---|---|---|
| Next.js frontend | 3000 | `npm run dev` (repo root) |
| Go API (Chapter 1) | 3005 (configurable via `-port`) | `cd go-api && go build -o bin/go-api ./cmd/go-api && ./bin/go-api` (or `go run ./cmd/go-api` for quick iteration) |

(Port 3005 was kept even though the old multi-framework project used 3001–3004 for other frameworks — no need to renumber since those chapters don't exist here.)

---

## README.md structure (top-level)

1. **Title + intro** — framed for TS/Python devs learning systems languages; Chapter 1 = Go, Rust planned next.
2. **Table of Contents**
3. **Project Structure** — ASCII tree (kept in sync with the actual repo — update both if structure changes)
4. **Quick Start** — prerequisites (Node ≥20.9, Go ≥1.21), install deps, run all servers via `./scripts/start-servers.sh`
5. **Chapter 1 — Go** — summarizes `docs/go/README.md`'s four parts as a preview/TOC, points reader there, includes a quick `curl` smoke test (including the `/enrich` concurrency endpoint)
6. **Adding a New Chapter** — referenced from the TOC but the actual section content doesn't currently exist in the file (pre-existing gap, not introduced by this pass) — if/when the Rust chapter is added, this section needs to be written for the first time, following the `docs/<lang>/` + `<lang>-api/{cmd,internal}` template established by Go.

---

## docs/go/ structure (the teaching documents — one folder, four files)

Written specifically for developers coming from TypeScript/Python, learning Go for the first time. Split into a `README.md` index plus four progressively-more-advanced parts (previously a single `docs/go.md`; split apart in this pass so each part stays focused and the concurrency/testing content — which didn't exist before — has room to breathe):

### `README.md` — Chapter 1 index
Short landing page: links to all four parts in order, plus an ASCII tree of where the corresponding code lives in `go-api/`.

### `01-basics.md` — Part 1: Go Language Basics (concepts only, no API/concurrency/testing code yet)
Each section introduces a Go concept by direct comparison to TS/Python equivalents:
- Why Go? (comparison table: execution, typing, errors, concurrency, object model, package manager, null handling, deploy deps)
- Setting Up (`brew install go`)
- Packages and `main` (incl. a note that most `.go` files aren't `package main` — they're shared library packages, foreshadowing `cmd/`/`internal/`)
- Variables, Types, and `:=` (incl. **zero values** as Go's alternative to `null`/`undefined`/`None`)
- Functions and Multiple Return Values
- **Error Handling — No Exceptions** (the `if err != nil` pattern; no `try`/`catch`/`throw`; brief note on `panic`/`recover`)
- **Defer** (new section — LIFO order, argument-eval-immediately semantics, ties to `defer s.mu.Unlock()` in the store)
- Structs — Go's "Classes" (vs. TS `interface`, Python `@dataclass`)
- Methods and Pointer Receivers (`func (x *T) Method()` vs. value receivers)
- Interfaces — Structural Typing (implicitly satisfied, unlike TS `implements`)
- Slices and Maps (`[]T`, `map[K]V`, the `value, ok := m[key]` idiom)
- Struct Tags (the backtick metadata syntax that `encoding/json` and `go-playground/validator` read)
- **Goroutines/Channels were moved OUT of this file** — they're now the entirety of Part 2.

### `02-concurrency.md` — Part 2: Concurrency (NEW — standalone, no HTTP)
- Goroutines (comparison table vs. JS event loop/Promises and Python asyncio/GIL; true multi-core parallelism)
- The `sync.WaitGroup` (`Add`/`Done`/`Wait`, loop-variable-capture gotcha)
- Channels (unbuffered vs. buffered, `close`, `range`)
- **Putting it together: FanOut** — walks through `go-api/internal/concurrency/fanout.go` line by line, including *why* the `wg.Wait()`+`close()` step must run in its own goroutine (deadlock otherwise)
- **`cmd/` and `internal/`: organizing a real Go project** — explains why `FanOut` had to move into its own importable package (two `func main()`s can't coexist), and that `internal/` is a compiler-enforced privacy boundary with no clean TS/Python equivalent
- Try It Yourself: `go run ./cmd/concurrency-demo`, tweak sleep durations, read the test file

### `03-testing.md` — Part 3: Testing (NEW)
- The `testing` package basics (`TestXxx(t *testing.T)`, `t.Fatalf` vs `t.Errorf`)
- Table-driven tests + `t.Run` subtests
- **testify** (`assert`/`require`) — explicitly justified as "most real-world Go projects add this" rather than stdlib-only
- `httptest.NewServer(api.NewRouter())` for full-stack handler tests (real router, real middleware, no mocking)
- Testing concurrent code — asserting on wall-clock elapsed time to prove `FanOut` runs concurrently; `go test -race`
- Try It Yourself: run the suite, deliberately break `Store.Patch`, confirm tests catch it, run with `-race`

### `04-api.md` — Part 4: The CRUD API (ties back to Parts 1–3 throughout)
- What is chi? (lightweight router, `http.Handler` interface, middleware pipeline)
- Project Setup with Go Modules (`go mod init`, `go get`, comparison table to npm/pip/uv) — commands now point at `./cmd/go-api` specifically, not the module root
- How the API Is Organized — full `cmd/`+`internal/` tree, then walks through `internal/items` (data layer, no HTTP) and `internal/api` (HTTP layer), cross-referencing Part 1 sections throughout
- Validation with go-playground/validator — struct tags, `Validate.Struct(...)`, the `*string` pointer trick for PATCH semantics
- **The Concurrency Showcase Endpoint** (NEW) — `GET /api/items/{id}/enrich`, walks through how it reuses `internal/concurrency.FanOut` from Part 2 inside a real handler, `elapsedMs` in the JSON response as observable proof of concurrent execution
- **Running as a Real Binary: Graceful Shutdown** (NEW) — the full `cmd/go-api/main.go`: `-port` flag, server started on its own goroutine, `signal.Notify`+`context.WithTimeout`+`srv.Shutdown`, and *why* this specifically requires the compiled binary rather than `go run .` (signal forwarding through the `go run` parent process is not reliable)
- How the UI Connects — `GoCrud.tsx`, CORS setup
- Try It Yourself — numbered hands-on steps: curl every CRUD op + the `/enrich` endpoint, read all four packages side by side, build+run+Ctrl-C the real binary, run the test suite

**Convention to preserve for future chapters:** Any new `docs/<language>/` should follow this same shape — a `README.md` index plus ordered parts (fundamentals → concurrency/async model → testing → API), each part cross-referencing earlier ones. The corresponding `<language>-api/` should use the same `cmd/` (runnable entrypoints) + `internal/` (or that language's equivalent private-package convention) split demonstrated by `go-api/`.

---

## go-api/README.md structure

Short, self-contained, no cross-references to other chapters:
- One-line description + link to chi and go-playground/validator
- Project layout (small ASCII tree of `cmd/`/`internal/`)
- Setup (`go mod download`)
- Run — `go run ./cmd/go-api` for quick iteration, `go build -o bin/go-api ./cmd/go-api && ./bin/go-api` as the primary workflow, `-port` flag example
- Try the concurrency demo (`go run ./cmd/concurrency-demo`)
- Test (`go test ./...`, `-v`, `-race`)
- Test the API (curl examples, incl. `/enrich`)
- Link to `../docs/go/README.md` for the full guided tour

---

## Frontend architecture notes

- **`app/page.tsx`**: `"use client"` component. `MainTab` type is currently just `"go-crud"`. Renders `<PageHeader />`, `<TabNavigation />`, and conditionally shows `<GoCrud />` (currently always shown since it's the only tab, toggled via CSS `hidden` class rather than unmounting).
- **`app/components/TabNavigation.tsx`**: `TABS` array currently has one entry (`{ id: "go-crud", label: "🐹 Go" }`) with an inline comment showing exactly how to add a new tab (e.g. `{ id: "rust-crud", label: "🦀 Rust" }`) — **when adding Rust, update the `MainTab` type union in both `TabNavigation.tsx` and `page.tsx`.**
- **`app/components/GoCrud.tsx`**: Full CRUD UI (create/edit/delete/list) talking to `http://localhost:3005`. Has its own client-side validation functions duplicating the Go validator rules. Has one eslint-disable comment (`react-hooks/set-state-in-effect`) on the initial `fetchItems()` call inside `useEffect` — this is intentional/expected, not a bug to "fix" later.
- **`app/components/PageHeader.tsx`**: Static header, title "🐹 Golang / Rust API Tutorial", subtitle "Chapter 1 — Go: a hands-on CRUD API walkthrough".
- No Zod, no shared schema file — validation logic lives independently in Go (`go-api/internal/items/schemas.go`) and TS (`GoCrud.tsx`), kept manually in sync.

---

## Go backend architecture notes (`go-api/`)

Restructured in this pass from one flat `package main` (3 files) into `cmd/` + `internal/` (see [Repo file structure](#repo-file-structure) tree above and `docs/go/02-concurrency.md#cmd-and-internal-organizing-a-real-go-project` for the full rationale).

- **`cmd/go-api/main.go`**: the real server entrypoint. Parses a `-port` flag (default `3005`), builds `&http.Server{Addr: ..., Handler: api.NewRouter()}`, starts it on its own goroutine, blocks on `signal.Notify(stop, os.Interrupt, syscall.SIGTERM)`, then does `context.WithTimeout(..., 5*time.Second)` + `srv.Shutdown(ctx)` for graceful shutdown. **This is why `go run .` was replaced with build-then-run everywhere** — signal delivery to a `go run` child process isn't reliable, so graceful shutdown needs the compiled binary.
- **`cmd/concurrency-demo/main.go`**: standalone, no HTTP at all. Calls `internal/concurrency.FanOut` with 3 simulated slow checks and prints elapsed time, proving concurrent (not sequential) execution.
- **`internal/items/`**: the data layer, zero HTTP dependency.
  - `item.go`: `Item` struct (`ID`, `Name`, `Description` w/ `json` tags).
  - `store.go`: `Store` (mutex-guarded `map[int]*Item` + `nextID` counter, seeded with one sample item), exported methods `List`/`Get`/`Create`/`Replace`/`Patch`/`Delete`.
  - `schemas.go`: `CreateItemInput`, `UpdateItemInput` (both require `Name`, `min=2,max=100`; optional `Description` `max=500`), `PatchItemInput` (both fields are `*string`/pointers so "omitted" vs "empty string" can be distinguished — ties to the zero-value concept taught in `docs/go/01-basics.md`). `Sanitize()` methods trim whitespace. `FirstValidationError()` extracts one human-readable message from `validator.ValidationErrors`. Exported `Validate` is the shared validator instance.
  - `store_test.go` / `schemas_test.go`: testify-based table-driven tests (new in this pass).
- **`internal/concurrency/fanout.go`**: `Check`/`Result` types + `FanOut(checks []Check) []Result` — goroutines + `sync.WaitGroup` + buffered channel, the fan-out/fan-in pattern (new in this pass). `fanout_test.go` asserts on real wall-clock timing to prove concurrency.
- **`internal/api/`**: the HTTP layer, imports `items` and `concurrency`.
  - `router.go`: `NewRouter() *chi.Mux` — `chi.NewRouter()`, middleware stack (`middleware.Logger`, `middleware.Recoverer`, `cors.Handler` allowing `http://localhost:3000`), mounts `itemsAPI.routes` at `/api/items`. (No more package-level `http.ListenAndServe` call here — that moved to `cmd/go-api/main.go` so the router itself stays a pure, testable `http.Handler`.)
  - `items_handlers.go`: list/create/getOne/update/partialUpdate/remove — same logic as before, just calling exported `Store`/`Validate` methods from `internal/items` instead of unexported package-`main` versions.
  - `enrich_handler.go` (new): `GET /api/items/{id}/enrich` — fans out 3 simulated checks (inventory/pricing/reviews, 150ms sleep each) via `concurrency.FanOut`, returns `{"itemId", "checks", "errors", "elapsedMs"}`.
  - `json.go`: `writeJSON`/`writeError`/`parseID` helpers (unchanged logic, just relocated).
  - `items_handlers_test.go` / `enrich_handler_test.go` (new): `httptest.NewServer(api.NewRouter())`-based full-stack handler tests.
- All source comments were cleaned of cross-references to Express/NestJS/Zod/Pydantic (leftovers from the original multi-framework project) — comments now describe the Go code on its own terms.
- Routes: `GET/POST /api/items`, `GET/PUT/PATCH/DELETE /api/items/{id}`, `GET /api/items/{id}/enrich` (new). Errors return `{"error": "..."}` JSON with appropriate HTTP status codes (400/404/201/200).
- `go.mod` gained `github.com/stretchr/testify` as a new dependency in this pass (used only in `_test.go` files).

---

## Conventions to follow for future work

1. **No cross-framework comparisons other than TS/Python vs. the new language.** This is a TS/Python-dev-learning-a-new-language tutorial, not a framework bake-off. Don't reintroduce Express/NestJS/Flask/FastAPI/Zod/Pydantic references.
2. **Keep `docs/<language>/` multi-part**: language fundamentals → concurrency/async model → testing → API walkthrough, each as its own file, with explicit cross-references back to earlier parts (see `docs/go/` for the reference shape).
3. **Keep the tab-navigation architecture extensible.** `TabNavigation.tsx` is intentionally structured (via the `MainTab` type + `TABS` array + inline comment) to make adding a new language tab a small, well-defined diff.
4. **Every chapter's backend is self-contained** with its own `README.md`, own port, own dependency manifest, own `cmd/`+`internal/`-style layout separating runnable entrypoints from shared/private packages — no shared code between chapters' backends.
5. **Validate before finishing:** `npm run build`, `npm run lint`, `go build ./...`/`go vet ./...`/`go test ./...`/`go test -race ./...` in `go-api/`, and a live smoke test (start both servers, curl the API incl. any concurrency endpoints, curl the frontend, confirm graceful shutdown on Ctrl-C) should all be run and pass before considering a task done.
6. **Keep README.md's "Project Structure" ASCII tree in sync** with the actual repo whenever files are added/removed/renamed.

---

## Likely next steps (not yet started)

- **Chapter 2 — Rust.** Following the "Adding a New Chapter" checklist in `README.md`:
  1. New `rust-api/` folder (likely using a lightweight Rust web framework — e.g. Axum or Actix-web — TBD, needs a decision).
  2. New `app/components/RustCrud.tsx`, modeled on `GoCrud.tsx`.
  3. Update `TabNavigation.tsx`'s `MainTab` type and `TABS` array to add `"rust-crud"` / `{ id: "rust-crud", label: "🦀 Rust" }`.
  4. Update `app/page.tsx` to import and conditionally render `RustCrud`.
  5. Write `docs/rust/{README,01-basics,02-concurrency,03-testing,04-api}.md` following the same multi-part structure as `docs/go/` (Rust fundamentals compared to TS/Python and possibly Go, then async/concurrency, then testing, then the API walkthrough) — ownership/borrowing, `Result<T, E>` vs Go's `(T, error)`, async runtimes (tokio) vs goroutines, traits vs interfaces, etc. would be natural Part 1 topics; tokio tasks/channels vs goroutines/channels would be the natural Part 2.
  6. Update `scripts/start-servers.sh` to also launch the Rust server.
  7. Update top-level `README.md`'s Project Structure tree, Table of Contents, and add a "Chapter 2 — Rust" section.
- No specific port has been reserved for Rust yet — pick one that doesn't conflict with 3000/3005 (e.g. 3006 would be a natural next choice given the old project's numbering scheme).
