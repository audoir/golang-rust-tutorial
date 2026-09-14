# Chapter 1 — Go CRUD

← [Back to README](../README.md)

---

## Table of Contents

- [What is chi?](#what-is-chi)
- [Core Go HTTP Concepts](#core-go-http-concepts)
- [Project Setup with Go Modules](#project-setup-with-go-modules)
- [How the API Works](#how-the-api-works)
- [Validation with go-playground/validator](#validation-with-go-playgroundvalidator)
- [How the UI Connects](#how-the-ui-connects)
- [Try It Yourself](#try-it-yourself)

---

## What is chi?

[chi](https://github.com/go-chi/chi) is a **lightweight, idiomatic HTTP router** for Go. Go's standard library (`net/http`) already gives you a working HTTP server, but its built-in `ServeMux` router has no support for path parameters (`/api/items/{id}`) or a composable middleware chain — chi adds exactly those two things and nothing else.

chi is **unopinionated routing plus a middleware pipeline**, built directly on top of `net/http` with zero framework "magic" — if you've used a minimal web framework in another language (e.g. Express in Node.js, Flask in Python), the mental model will feel familiar.

Key selling points:

| Feature | Description |
|---|---|
| **100% compatible with `net/http`** | A chi router *is* an `http.Handler`. Any standard library or third-party `net/http` middleware works unmodified. |
| **Lightweight router** | chi adds URL parameters (`chi.URLParam`) and route grouping (`r.Route(...)`) — nothing more. |
| **Middleware pipeline** | Every request flows through a stack of `func(http.Handler) http.Handler` functions — each middleware wraps the next handler in the chain. |
| **No reflection, no code generation** | chi routes are plain Go function calls — what you see is what runs. |
| **Explicit routing** | Routes are registered imperatively with `r.Get(...)`, `r.Post(...)`, etc. |

---

## Core Go HTTP Concepts

| Concept | Description |
|---|---|
| **Language** | Go (compiles to a single static binary) |
| **Package manager** | Go modules (`go.mod` / `go.sum`) |
| **Routing style** | Imperative (`r.Get(...)`, `r.Post(...)`) |
| **Validation library** | `go-playground/validator` (struct tags) |
| **Body parsing** | Manual `json.NewDecoder(r.Body).Decode(&body)` per handler — there is no automatic body-parsing middleware |
| **Error handling** | Each handler writes its own error response; `middleware.Recoverer` catches panics and converts them into a 500 instead of crashing the process |
| **Concurrency model** | OS threads + goroutines — every incoming request is handled on its own lightweight goroutine automatically |
| **Runtime** | Compiles to a single static native binary — no runtime needed to deploy |
| **Port** | 3005 |

> **Go's concurrency model:** Go's HTTP server spins up a lightweight **goroutine** per incoming request automatically; the Go scheduler multiplexes goroutines across OS threads, so CPU-bound and I/O-bound work can both run truly in parallel across multiple cores without you writing any special "async" code.

---

## Project Setup with Go Modules

The Go project lives in **`go-api/`** and uses Go's built-in dependency manager, **Go modules**.

### What Go modules give you

| Go modules concept | Equivalent in Node.js |
|---|---|
| `go.mod` | `package.json` |
| `go.sum` | `package-lock.json` |
| Global module cache (`$GOPATH/pkg/mod`) | `node_modules/` (but shared across all projects, not per-project) |
| `go get <package>` | `npm install <package>` |
| `go run .` | `node index.js` / `ts-node index.ts` |
| `go build` | `tsc` (but produces a single native binary, not JS) |

### Initialising the project

```bash
# Create a new Go module
cd go-api
go mod init go-api

# Add dependencies — this updates go.mod and go.sum automatically
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get github.com/go-playground/validator/v10
```

### Running the server

```bash
# From the go-api/ directory — compiles and runs in one step
go run .

# Or build a native binary once and run it directly
go build -o bin/go-api .
./bin/go-api

# Or from the repo root via the start script
./scripts/start-servers.sh
```

There is no `--reload` / `--respawn` equivalent built in — Go compiles very fast, so most Go developers just re-run `go run .` after a save, or use a third-party watcher like [air](https://github.com/air-verse/air) for auto-reload during development.

---

## How the API Works

The Go server is split across three files in **`go-api/`**:

```
go-api/
├── main.go       # entry point: router, middleware, server startup
├── items.go      # in-memory store + HTTP handlers (the "controller")
├── schemas.go    # validated input structs
└── go.mod        # module definition + dependencies
```

### 1. Entry point — `go-api/main.go`

```go
func main() {
	r := chi.NewRouter()

	// ── Middleware ──────────────────────────────────────────────────────────
	r.Use(middleware.Logger)    // logs method, path, status, and latency
	r.Use(middleware.Recoverer) // recovers from panics with a 500 instead of crashing

	// Allow requests from the Next.js dev server (port 3000)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	// ── Routes ──────────────────────────────────────────────────────────────
	api := newItemsAPI()
	r.Route("/api/items", api.routes)

	http.ListenAndServe(":3005", r)
}
```

- `chi.NewRouter()` creates the router — it satisfies `http.Handler`, so it can be passed straight to `http.ListenAndServe`.
- `r.Use(...)` registers middleware, run in order for every request.
- `middleware.Logger` and `middleware.Recoverer` ship with chi itself; `cors.Handler(...)` comes from the separate `go-chi/cors` package.
- `r.Route("/api/items", api.routes)` mounts a sub-router at a base path — every route registered inside `api.routes` is automatically prefixed with `/api/items`.

### 2. In-memory store — `go-api/items.go`

```go
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type itemStore struct {
	mu     sync.Mutex
	items  map[int]*Item
	nextID int
}
```

A plain in-memory collection, reset on restart — good enough for a tutorial, but not for production. Go has no built-in single-threaded guarantee, so a `sync.Mutex` protects the store from concurrent access whenever multiple goroutines (i.e. multiple in-flight requests) read or write it at the same time.

### 3. Routes — `go-api/items.go`

```go
func (a *itemsAPI) routes(r chi.Router) {
	r.Get("/", a.list)
	r.Post("/", a.create)
	r.Get("/{id}", a.getOne)
	r.Put("/{id}", a.update)
	r.Patch("/{id}", a.partialUpdate)
	r.Delete("/{id}", a.remove)
}

// GET /api/items — list all items
func (a *itemsAPI) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.list())
}

// POST /api/items — create a new item
func (a *itemsAPI) create(w http.ResponseWriter, r *http.Request) {
	var body CreateItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, firstValidationError(err))
		return
	}
	item := a.store.create(body.Name, body.Description)
	writeJSON(w, http.StatusCreated, item)
}
```

Key things to notice:

- **`{id}`** is chi's path-parameter syntax for URL segments. Read it back with `chi.URLParam(r, "id")`.
- **No automatic body parsing.** Go's standard library requires you to decode the JSON body yourself with `json.NewDecoder(r.Body).Decode(&body)` in every handler that needs one.
- **No automatic error responses.** Every handler writes its own status code and JSON error body via small `writeJSON` / `writeError` helpers — there is no built-in global exception filter, so this tutorial defines that convention itself.
- **`http.ResponseWriter` / `*http.Request`** are the two parameters every handler receives — everything a handler needs to read the request and write the response.

### 4. Entry point details

There is no separate "dev" vs "production" command — `go run .` compiles and runs the program in one step, and `go build` produces a single self-contained native binary with no external runtime dependency (no interpreter, no virtual environment, nothing to install on the target machine besides the binary itself).

---

## Validation with go-playground/validator

Go has no built-in schema validation and no runtime type-introspection library baked into the language. The community standard is [`go-playground/validator`](https://github.com/go-playground/validator), which reads **struct tags** to describe validation rules declaratively — you annotate each struct field with a `validate:"..."` tag, then call `validate.Struct(...)` to check an instance against those rules.

### Schemas — `go-api/schemas.go`

```go
type CreateItemInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

type UpdateItemInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// PatchItemInput uses pointers so "field omitted" can be told apart from
// "field explicitly set to empty string" — a nil pointer means the field
// was not present in the request body at all.
type PatchItemInput struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}
```

### On the server

```go
var body CreateItemInput
json.NewDecoder(r.Body).Decode(&body)
body.Sanitize() // trims whitespace before validating

if err := validate.Struct(body); err != nil {
	writeError(w, http.StatusBadRequest, firstValidationError(err))
	return
}
```

`validate.Struct(body)` walks every field's `validate:"..."` tag and returns a `validator.ValidationErrors` if any rule fails. `firstValidationError` extracts a single human-readable message from the first failing field so the client gets a clean, readable error instead of a raw Go error value.

### How struct tags express rules

| Tag | Meaning |
|---|---|
| `required` | Field must be present and non-zero |
| `min=2` | Minimum length (strings) or value (numbers) |
| `max=100` | Maximum length (strings) or value (numbers) |
| `omitempty` | Skip this rule if the field is the zero value / nil |

Describe the shape once with struct tags, validate every request against it with `validate.Struct(...)`, and read a structured error back out if it fails — the same three-step pattern used by schema-validation libraries in most languages.

---

## How the UI Connects

The `GoCrud` component (`app/components/GoCrud.tsx`) is a plain client component that talks directly to the Go server on port 3005:

```ts
const GO_API = "http://localhost:3005";
const res = await fetch(`${GO_API}/api/items`);
```

The Go API returns `{"error": "..."}` on failure, so `GoCrud` reads `err.error` to display a readable message in the UI. Because the Next.js app (port 3000) and the Go API (port 3005) run on different origins, the Go server enables CORS for `http://localhost:3000` via the `cors.Handler(...)` middleware registered in `main.go` (see [How the API Works](#how-the-api-works) above).

---

## Try It Yourself

1. Start all servers: `./scripts/start-servers.sh`
2. Open **http://localhost:3000** — the **🐹 Go** tab is selected by default.
3. Create, edit, and delete items using the form and item list.
4. Open DevTools → Network and notice the requests go to `http://localhost:3005`.
5. Try the Go API directly with `curl`:

```bash
# List all items
curl http://localhost:3005/api/items

# Create an item
curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Go Item","description":"Created via curl"}'

# Try an invalid body — validator will reject it
curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"x"}'

# Try a missing name — validator will reject it
curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"description":"No name provided"}'

# Update an item (replace 1 with the actual id)
curl -X PUT http://localhost:3005/api/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Go Item","description":"New description"}'

# Partially update an item (PATCH)
curl -X PATCH http://localhost:3005/api/items/1 \
  -H "Content-Type: application/json" \
  -d '{"description":"Only the description changed"}'

# Delete an item
curl -X DELETE http://localhost:3005/api/items/1
```

6. **Read the source side-by-side:** Open `go-api/main.go`, `go-api/items.go`, and `go-api/schemas.go` together. The routing, in-memory store, and struct-tag-based validation all live in a small number of files, which makes it easy to trace a request from `main.go`'s router all the way through to the JSON response.

7. **Compare startup models:** Run `go build -o bin/go-api . && ./bin/go-api` and notice there is no `node_modules`, no interpreter, and no separate install step at runtime — the compiled binary is completely self-contained.

---
