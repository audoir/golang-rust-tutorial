# Part 4 — The CRUD API

← [Back to Chapter 1 index](README.md) · ← [Part 3 — Testing](03-testing.md)

With the language basics, concurrency, and testing in place, let's look at how they all come together in a real, running CRUD API — composed from the same `internal/items` and `internal/concurrency` packages introduced earlier, wired up with [chi](https://github.com/go-chi/chi) and [go-playground/validator](https://github.com/go-playground/validator).

## Table of Contents

- [What is chi?](#what-is-chi)
- [Project Setup with Go Modules](#project-setup-with-go-modules)
- [How the API Is Organized](#how-the-api-is-organized)
- [Validation with go-playground/validator](#validation-with-go-playgroundvalidator)
- [The Concurrency Showcase Endpoint](#the-concurrency-showcase-endpoint)
- [Running as a Real Binary: Graceful Shutdown](#running-as-a-real-binary-graceful-shutdown)
- [How the UI Connects](#how-the-ui-connects)
- [Try It Yourself](#try-it-yourself)

---

## What is chi?

[chi](https://github.com/go-chi/chi) is a **lightweight, idiomatic HTTP router** for Go. Go's standard library (`net/http`) already gives you a working HTTP server, but its built-in `ServeMux` router has no support for path parameters (`/api/items/{id}`) or a composable middleware chain — chi adds exactly those two things and nothing else.

chi is unopinionated routing plus a middleware pipeline, built directly on top of `net/http` with zero framework "magic" — if you've used a minimal web framework in another language (e.g. Express in Node.js, Flask in Python), the mental model will feel familiar: register handlers on paths, stack middleware in front of them.

Key selling points:

| Feature | Description |
|---|---|
| **100% compatible with `net/http`** | A chi router *is* an `http.Handler` (see [Interfaces](01-basics.md#interfaces--structural-typing)) — any standard library or third-party `net/http` middleware works unmodified. |
| **Lightweight router** | chi adds URL parameters (`chi.URLParam`) and route grouping (`r.Route(...)`) — nothing more. |
| **Middleware pipeline** | Every request flows through a stack of `func(http.Handler) http.Handler` functions — each middleware wraps the next handler in the chain. |
| **No reflection, no code generation for routing** | chi routes are plain Go function calls — what you see is what runs. |
| **Explicit routing** | Routes are registered imperatively with `r.Get(...)`, `r.Post(...)`, etc. |

### Quick facts about this API

| Concept | Description |
|---|---|
| **Routing style** | Imperative (`r.Get(...)`, `r.Post(...)`) |
| **Validation library** | `go-playground/validator` (struct tags) |
| **Body parsing** | Manual `json.NewDecoder(r.Body).Decode(&body)` per handler — there is no automatic body-parsing middleware |
| **Error handling** | Each handler writes its own error response; `middleware.Recoverer` catches panics and converts them into a 500 instead of crashing the process |
| **Concurrency model** | Goroutine per request, automatically, via `net/http` — plus a deliberate goroutine fan-out inside `/enrich` (see below) |
| **Port** | 3005 (configurable via `-port`) |

---

## Project Setup with Go Modules

The Go project lives in **`go-api/`** and uses Go's built-in dependency manager, **Go modules**.

### What Go modules give you

| Go modules concept | Equivalent in Node.js | Equivalent in Python |
|---|---|---|
| `go.mod` | `package.json` | `pyproject.toml` |
| `go.sum` | `package-lock.json` | `uv.lock` |
| Global module cache (`$GOPATH/pkg/mod`) | `node_modules/` (but shared across all projects, not per-project) | `~/.cache/uv` (shared cache) |
| `go get <package>` | `npm install <package>` | `uv add <package>` |
| `go run ./cmd/go-api` | `node index.js` / `ts-node index.ts` | `python main.py` |
| `go build ./cmd/go-api` | `tsc` (but produces a single native binary, not JS) | (no direct equivalent — Python isn't compiled) |

### Initialising the project

```bash
# Create a new Go module
cd go-api
go mod init go-api

# Add dependencies — this updates go.mod and go.sum automatically
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get github.com/go-playground/validator/v10
go get github.com/stretchr/testify
```

### Running the server

This project has **two** runnable programs (see [`cmd/` and `internal/`](02-concurrency.md#cmd-and-internal-organizing-a-real-go-project)), so commands need to point at a specific `cmd/` folder rather than the module root:

```bash
# Quick iteration — compiles and runs in one step, from go-api/
go run ./cmd/go-api

# Build a native binary once, then run it directly — the primary workflow
# for this tutorial from here on (see "Running as a Real Binary" below)
go build -o bin/go-api ./cmd/go-api
./bin/go-api

# Or from the repo root via the start script
./scripts/start-servers.sh
```

There is no `--reload`/`--watch` flag built in — Go compiles very fast, so most Go developers just re-run after a save, or use a third-party watcher like [air](https://github.com/air-verse/air) for auto-reload during development.

---

## How the API Is Organized

```
go-api/
├── cmd/
│   ├── go-api/main.go              # entry point: flags, router, graceful shutdown
│   └── concurrency-demo/main.go    # standalone demo from Part 2 (not part of the server)
└── internal/
    ├── items/
    │   ├── item.go                  # Item struct
    │   ├── store.go                 # Store — mutex-guarded map + CRUD methods
    │   └── schemas.go                # CreateItemInput/UpdateItemInput/PatchItemInput + validation
    ├── concurrency/
    │   └── fanout.go                 # FanOut — from Part 2, reused for real below
    └── api/
        ├── router.go                 # NewRouter(): chi setup, middleware, route mounting
        ├── items_handlers.go         # list/create/getOne/update/partialUpdate/remove
        ├── enrich_handler.go         # GET /api/items/{id}/enrich
        └── json.go                   # writeJSON/writeError/parseID helpers
```

### `internal/items` — the data layer

```go
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Store struct {
	mu     sync.Mutex
	items  map[int]*Item
	nextID int
}
```

`Item` is a plain struct (see [Structs](01-basics.md#structs--gos-classes)) with `json:"..."` struct tags (see [Struct Tags](01-basics.md#struct-tags)) controlling how it's serialized. `Store` wraps a `map[int]*Item` (see [Slices and Maps](01-basics.md#slices-and-maps)) — a plain in-memory collection, reset on restart, good enough for a tutorial but not for production.

Because every incoming request runs on its own goroutine (see [Part 2](02-concurrency.md#goroutines)), multiple requests could read and write the underlying map at the same time. The `sync.Mutex` field (`mu`) prevents that: every store method calls `s.mu.Lock()` before touching `items` and `defer s.mu.Unlock()` (see [Defer](01-basics.md#defer)) to release it when the method returns. This package has **no HTTP dependency at all** — see `internal/items/store_test.go` for how that makes it trivial to unit test in isolation.

### `internal/api` — the HTTP layer

```go
func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{ /* ... */ }))

	api := &itemsAPI{store: items.NewStore()}
	r.Route("/api/items", api.routes)
	return r
}

func (a *itemsAPI) routes(r chi.Router) {
	r.Get("/", a.list)
	r.Post("/", a.create)
	r.Get("/{id}", a.getOne)
	r.Put("/{id}", a.update)
	r.Patch("/{id}", a.partialUpdate)
	r.Delete("/{id}", a.remove)
	r.Get("/{id}/enrich", a.enrich)
}
```

- `chi.NewRouter()` satisfies the `http.Handler` interface (see [Interfaces](01-basics.md#interfaces--structural-typing)), so it can be passed straight to an `http.Server`.
- `r.Use(...)` registers middleware, run in order for every request. `middleware.Recoverer` is what turns a `panic` (see [Error Handling](01-basics.md#error-handling--no-exceptions)) into a clean 500 response instead of crashing the whole process.
- `r.Route("/api/items", api.routes)` mounts a sub-router at a base path — every route registered inside `api.routes` is automatically prefixed with `/api/items`.
- `itemsAPI` groups the `Store` together with its handler methods (a pointer receiver — see [Methods and Pointer Receivers](01-basics.md#methods-and-pointer-receivers) — since handlers need to call mutating store methods).

```go
func (a *itemsAPI) create(w http.ResponseWriter, r *http.Request) {
	var body items.CreateItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := items.Validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, items.FirstValidationError(err))
		return
	}
	item := a.store.Create(body.Name, body.Description)
	writeJSON(w, http.StatusCreated, item)
}
```

- **`{id}`** is chi's path-parameter syntax for URL segments. Read it back with `chi.URLParam(r, "id")` (wrapped here as `parseID`).
- **No automatic body parsing.** Go's standard library requires you to decode the JSON body yourself with `json.NewDecoder(r.Body).Decode(&body)` in every handler that needs one.
- **`if err := ...; err != nil { ... }`** is the error-checking pattern from [Error Handling](01-basics.md#error-handling--no-exceptions), used at every step that can fail: decoding JSON, validating the struct, looking up the item.
- **No automatic error responses.** Every handler writes its own status code and JSON error body via small `writeJSON`/`writeError` helpers in `internal/api/json.go` — there is no built-in global exception filter, so this tutorial defines that convention itself.

---

## Validation with go-playground/validator

Go has no built-in schema validation and no runtime type-introspection library baked into the language. The community standard is [`go-playground/validator`](https://github.com/go-playground/validator), which reads the `validate:"..."` struct tags introduced in [Struct Tags](01-basics.md#struct-tags) to describe validation rules declaratively — you annotate each struct field, then call `Validate.Struct(...)` to check an instance against those rules.

### Schemas — `go-api/internal/items/schemas.go`

```go
type CreateItemInput struct {
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

`PatchItemInput` using `*string` instead of `string` is a direct application of [zero values](01-basics.md#variables-types-and-): an omitted `string` field would silently decode to `""` (its zero value), which is indistinguishable from the client explicitly sending an empty string. A `*string` defaults to `nil` when omitted, which the handler code can check for explicitly — see `internal/items/schemas_test.go`'s `TestPatchItemInput_OmittedFieldsAreNil`.

### On the server

```go
var body items.CreateItemInput
json.NewDecoder(r.Body).Decode(&body)
body.Sanitize() // trims whitespace before validating

if err := items.Validate.Struct(body); err != nil {
	writeError(w, http.StatusBadRequest, items.FirstValidationError(err))
	return
}
```

`items.Validate.Struct(body)` walks every field's `validate:"..."` tag and returns a `validator.ValidationErrors` if any rule fails. `items.FirstValidationError` extracts a single human-readable message from the first failing field so the client gets a clean, readable error instead of a raw Go error value.

### How struct tags express rules

| Tag | Meaning |
|---|---|
| `required` | Field must be present and non-zero |
| `min=2` | Minimum length (strings) or value (numbers) |
| `max=100` | Maximum length (strings) or value (numbers) |
| `omitempty` | Skip this rule if the field is the zero value / nil |

Describe the shape once with struct tags, validate every request against it with `Validate.Struct(...)`, and read a structured error back out if it fails — the same three-step pattern used by schema-validation libraries in most languages (Zod in TypeScript, Pydantic in Python), just expressed through struct tags instead of a fluent builder API or a base class.

---

## The Concurrency Showcase Endpoint

`GET /api/items/{id}/enrich` is where the `FanOut` function from [Part 2](02-concurrency.md#putting-it-together-fanout) stops being a standalone toy and becomes part of a real API — the exact same building block, now composed into an HTTP handler:

```go
// go-api/internal/api/enrich_handler.go
func (a *itemsAPI) enrich(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	// ...
	item, ok := a.store.Get(id)
	// ...

	checks := []concurrency.Check{
		{Name: "inventory", Run: func() (string, error) { /* simulated slow call */ }},
		{Name: "pricing", Run: func() (string, error) { /* simulated slow call */ }},
		{Name: "reviews", Run: func() (string, error) { /* simulated slow call */ }},
	}

	start := time.Now()
	results := concurrency.FanOut(checks)
	elapsed := time.Since(start)

	// ... build enrichResponse from results, including elapsed.Milliseconds()
	writeJSON(w, http.StatusOK, resp)
}
```

Each simulated check sleeps 150ms, standing in for a real downstream call (a database query, another microservice, a third-party API). Because `FanOut` runs all three concurrently rather than one after another, the whole request completes in roughly 150ms, not 450ms — the `elapsedMs` field in the JSON response makes this directly observable without needing to time it yourself:

```bash
curl http://localhost:3005/api/items/1/enrich
# {"itemId":1,"checks":{"inventory":"...","pricing":"$19.99","reviews":"..."},"elapsedMs":150}
```

`internal/api/enrich_handler_test.go` asserts on this timing directly with `httptest` (see [Part 3](03-testing.md#testing-http-handlers-with-httptest)), the same way `internal/concurrency/fanout_test.go` does for the standalone version.

---

## Running as a Real Binary: Graceful Shutdown

Everything up to this point would run fine with `go run ./cmd/go-api`. But `go-api/cmd/go-api/main.go` does two things that only really make sense — and are only really testable — against a **compiled, standalone process**, not a `go run` child process:

```go
func main() {
	port := flag.String("port", "3005", "port to listen on")
	flag.Parse()

	srv := &http.Server{Addr: ":" + *port, Handler: api.NewRouter()}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop // block here until Ctrl-C or `kill`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
```

- **A `-port` flag**, via the standard `flag` package — `./bin/go-api -port 3006` runs on a different port with no code changes, the same way a real service would take its port from a CLI flag or environment variable rather than a hardcoded constant.
- **Starting the server on its own goroutine** (see [Part 2](02-concurrency.md#goroutines)) so `main()` is free to block on something else — here, waiting for a shutdown signal — instead of blocking forever inside `ListenAndServe` itself.
- **`signal.Notify(stop, os.Interrupt, syscall.SIGTERM)`** registers `stop` (a channel — see [Channels](02-concurrency.md#channels)) to receive OS signals: `os.Interrupt` is Ctrl-C, `syscall.SIGTERM` is what `kill` sends by default, and what container orchestrators (Docker, Kubernetes) send to ask a process to shut down cleanly before forcibly killing it.
- **`<-stop`** blocks `main()` until one of those signals arrives.
- **`context.WithTimeout` + `srv.Shutdown(ctx)`** is Go's standard "wait, but not forever" pattern: `Shutdown` stops accepting new connections and waits for in-flight requests to finish, but only up to the 5-second deadline carried by `ctx` — after that, it gives up and returns, rather than hanging indefinitely on a stuck request. This is the same `context.WithTimeout` pattern used for timeouts and cancellation throughout the Go ecosystem (HTTP clients, database queries, gRPC calls, etc.), introduced here for the first time in this tutorial.

**Why this needs the compiled binary, not `go run`:** `go run` builds your program into a temporary binary and runs *that* as a child process of the `go` command itself. Signals sent to the parent (`go run`) aren't guaranteed to be forwarded to the child the way they are when you run a compiled binary directly — so testing graceful shutdown properly means building first:

```bash
go build -o bin/go-api ./cmd/go-api
./bin/go-api -port 3005
# in another terminal: kill -TERM <pid>, or just Ctrl-C in the first terminal
```

You should see `"Shutting down gracefully..."` followed by `"Server stopped."` in the logs, rather than the process just disappearing — proof the shutdown path actually ran. This is also why `scripts/start-servers.sh` and the primary workflow in this tutorial build the binary first rather than using `go run .` for the real server (the standalone `concurrency-demo` from Part 2 has no server to shut down, so `go run` is fine for that one).

---

## How the UI Connects

The `GoCrud` component (`app/components/GoCrud.tsx`) is a plain React client component that talks directly to the Go server on port 3005:

```ts
const GO_API = "http://localhost:3005";
const res = await fetch(`${GO_API}/api/items`);
```

The Go API returns `{"error": "..."}` on failure, so `GoCrud` reads `err.error` to display a readable message in the UI. `GoCrud` also duplicates the same validation rules in plain TypeScript (`validateName`, `validateDescription`) so the form can show instant feedback before ever making a network request — the source of truth still lives on the server in `go-api/internal/items/schemas.go`.

Because the Next.js app (port 3000) and the Go API (port 3005) run on different origins, the Go server enables CORS for `http://localhost:3000` via the `cors.Handler(...)` middleware registered in `internal/api/router.go` (see [How the API Is Organized](#how-the-api-is-organized) above).

---

## Try It Yourself

1. Start all servers: `./scripts/start-servers.sh` (this now builds the Go binary first, then runs it — see [Running as a Real Binary](#running-as-a-real-binary-graceful-shutdown))
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

# The concurrency showcase endpoint — watch elapsedMs
curl http://localhost:3005/api/items/1/enrich
```

6. **Read the source alongside all four parts:** Open `go-api/cmd/go-api/main.go`, `go-api/internal/api/*.go`, `go-api/internal/items/*.go`, and `go-api/internal/concurrency/fanout.go` and trace how each concept — structs, pointer receivers, error checking, goroutines, channels, `WaitGroup`, struct tags, graceful shutdown — shows up in real, working code.
7. **Compare startup models:** Run `go build -o bin/go-api ./cmd/go-api && ./bin/go-api` and notice there is no `node_modules`, no interpreter, and no separate install step at runtime — the compiled binary is completely self-contained. Stop it with Ctrl-C and watch the graceful shutdown log messages.
8. **Run the test suite:** `cd go-api && go test -v ./...` — see [Part 3](03-testing.md) for a full walkthrough of what's being tested and why.

---

← [Part 3 — Testing](03-testing.md) · [Back to Chapter 1 index](README.md)
