# Chapter 1 — Go

← [Back to README](../README.md)

This chapter is written for developers who already know **TypeScript** and/or **Python** and are learning **Go** for the first time. Every concept is introduced by comparing it to something you already know, before we look at the actual CRUD API implemented in `go-api/`.

---

## Table of Contents

**Part 1 — Go Language Basics**
- [Why Go?](#why-go)
- [Setting Up](#setting-up)
- [Packages and `main`](#packages-and-main)
- [Variables, Types, and `:=`](#variables-types-and-)
- [Functions and Multiple Return Values](#functions-and-multiple-return-values)
- [Error Handling — No Exceptions](#error-handling--no-exceptions)
- [Structs — Go's "Classes"](#structs--gos-classes)
- [Methods and Pointer Receivers](#methods-and-pointer-receivers)
- [Interfaces — Structural Typing](#interfaces--structural-typing)
- [Slices and Maps](#slices-and-maps)
- [Goroutines and Channels](#goroutines-and-channels)
- [Struct Tags](#struct-tags)

**Part 2 — The CRUD API**
- [What is chi?](#what-is-chi)
- [Project Setup with Go Modules](#project-setup-with-go-modules)
- [How the API Works](#how-the-api-works)
- [Validation with go-playground/validator](#validation-with-go-playgroundvalidator)
- [How the UI Connects](#how-the-ui-connects)
- [Try It Yourself](#try-it-yourself)

---

# Part 1 — Go Language Basics

## Why Go?

Go (also called **Golang**) is a compiled, statically-typed language created at Google. Compared to TypeScript and Python, the biggest mental shifts are:

| | TypeScript / Python | Go |
|---|---|---|
| **Execution** | Interpreted / JIT-compiled at runtime (Node.js, CPython) | Compiled ahead of time to a single native binary |
| **Typing** | TypeScript: static but erased at runtime. Python: dynamic. | Static and enforced at runtime — no `any`, no duck typing |
| **Errors** | `throw` / `try` / `catch` exceptions | Errors are ordinary return values you check explicitly |
| **Concurrency** | Single-threaded event loop (`async`/`await`, Promises) | Goroutines + channels, scheduled across real OS threads |
| **Object model** | Classes with inheritance | Structs + interfaces, no inheritance |
| **Package manager** | npm / pip / uv | Go modules (built into the toolchain) |
| **Null handling** | `null` / `undefined` / `None` | Every type has a "zero value"; pointers can be `nil` |
| **Dependencies to deploy** | `node_modules/`, a Node runtime, or a `.venv` + Python interpreter | None — `go build` produces one self-contained binary |

None of this makes Go "better" or "worse" — it's simply a different set of trade-offs, optimized for building simple, fast, and highly concurrent network services. Keep this table in mind as a reference as you read the rest of this chapter.

---

## Setting Up

Install Go (macOS via Homebrew):

```bash
brew install go
go version   # should print go1.21 or later
```

There is no separate package-manager binary to install (unlike `npm` for Node or `pip`/`uv` for Python) — dependency management is built into the `go` command itself, as you'll see in [Project Setup with Go Modules](#project-setup-with-go-modules).

---

## Packages and `main`

Every Go file starts with a `package` declaration. A runnable program needs exactly one package called `main`, and that package needs a `func main()` — the entry point, conceptually the same role as a script's top-level code in Python or the file you point `node` at in Node.js.

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

- `package main` — this file belongs to the `main` package (every `.go` file in a directory must declare the same package name).
- `import "fmt"` — `fmt` is a standard-library package for formatted I/O, roughly Go's equivalent of `console.log` (`fmt.Println`) plus Python's `str.format` (`fmt.Sprintf`).
- `func main()` — the entry point. Run it with `go run .`, or compile it to a binary with `go build`.

Unlike Node.js or Python, **unused imports and unused local variables are compile errors** in Go, not warnings — the compiler is intentionally strict about keeping code clean.

---

## Variables, Types, and `:=`

Go is statically typed, but it has type inference, so you rarely have to write types out by hand — similar to TypeScript's `let x = 5` inferring `number`.

```go
var name string = "Go"   // explicit type
var age = 25             // inferred type (int)
count := 0                // shorthand declare + infer — only inside functions
count = count + 1
```

- `var` declares a variable, optionally with an explicit type. It works both inside and outside functions.
- `:=` is shorthand for "declare and initialize with an inferred type" — it can only be used inside a function body, not at package level. This is the form you'll see most often.
- There is no `let`/`const` split by mutability the way JS has `let` vs `const` — Go has a separate `const` keyword for compile-time constants, and everything else declared with `var`/`:=` is mutable.

### Zero values — Go's alternative to `null`/`undefined`

Every declared-but-unassigned variable in Go gets a **zero value** rather than being `null`, `undefined`, or `None`:

| Type | Zero value |
|---|---|
| `int`, `float64` | `0` |
| `string` | `""` (empty string) |
| `bool` | `false` |
| pointers, slices, maps, functions, interfaces | `nil` |

This is why you'll see `nil` used as Go's rough equivalent of `null`/`None`, but only for types that can actually be "nothing" (pointers, slices, maps, etc.) — a plain `int` or `string` can never be `nil`, it just defaults to `0` or `""`.

---

## Functions and Multiple Return Values

```go
func add(a int, b int) int {
	return a + b
}

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil
}

result, err := divide(10, 2)
```

- Parameter types come **after** the name (`a int`, not `int a`) — the reverse of TypeScript's `a: number`.
- Go functions can return **multiple values** — there's no need to bundle them into an object/tuple the way you might in TS (`{ result, error }`) or Python (`return result, error`). This is used everywhere in Go, most importantly for error handling (next section).

---

## Error Handling — No Exceptions

This is the single biggest adjustment coming from TypeScript or Python. **Go has no `try`/`catch`/`throw`.** Instead, any function that can fail returns an `error` as its last return value, and the caller is expected to check it immediately:

```go
data, err := someFunction()
if err != nil {
	// handle the error — log it, return it, wrap it, etc.
	return err
}
// use `data` here — it's safe because err was nil
```

Compare this to what you're used to:

```typescript
// TypeScript
try {
  const data = someFunction();
  // use data
} catch (err) {
  // handle err
}
```

```python
# Python
try:
    data = some_function()
    # use data
except Exception as err:
    # handle err
```

The Go version has no hidden control flow — nothing "throws" and unwinds the stack automatically. Every error must be explicitly checked at the call site, which means Go code tends to have a lot of `if err != nil { ... }` blocks. It's more verbose, but it makes every possible failure point visible in the code, rather than hidden behind an implicit exception path.

> Go does have `panic`/`recover`, which behaves a bit like throwing/catching an exception, but it's reserved for truly unexpected, unrecoverable situations (e.g. a bug, an out-of-bounds array access) — not for ordinary error handling like "the request body was invalid JSON". You'll see `recover` used once in this tutorial's API, in the form of chi's `middleware.Recoverer` (see [How the API Works](#how-the-api-works)), which exists purely as a safety net so a single bad request can't crash the whole server.

---

## Structs — Go's "Classes"

Go has no classes and no inheritance. Instead, it has **structs**: plain data containers, similar to a TypeScript `interface`/`type` or a Python `dataclass`.

```go
type Item struct {
	ID          int
	Name        string
	Description string
}

item := Item{ID: 1, Name: "Sample", Description: "A sample item"}
fmt.Println(item.Name) // "Sample"
```

```typescript
// TypeScript equivalent
interface Item {
  id: number;
  name: string;
  description: string;
}
const item: Item = { id: 1, name: "Sample", description: "A sample item" };
```

```python
# Python equivalent (dataclass)
@dataclass
class Item:
    id: int
    name: str
    description: str
```

A key difference: struct fields starting with an **uppercase letter** are exported (public, visible outside the package), while lowercase fields are unexported (private to the package). There's no `public`/`private` keyword — visibility is determined entirely by the capitalization of the name. This is why you'll see `Item.Name` (uppercase, exported) but `itemStore.mu` (lowercase, package-private) in the API code.

---

## Methods and Pointer Receivers

Go doesn't have classes with methods baked in, but you can attach a method to any struct type using a **receiver**:

```go
type Counter struct {
	count int
}

// Pointer receiver — can modify the struct
func (c *Counter) Increment() {
	c.count++
}

// Value receiver — gets a copy, cannot modify the original
func (c Counter) Value() int {
	return c.count
}
```

- `(c *Counter)` is a **pointer receiver** — `c` is a pointer to the actual struct, so mutations inside the method affect the original. This is the closest equivalent to a regular mutating method (`this.count++`) in a TS/Python class.
- `(c Counter)` is a **value receiver** — `c` is a full copy of the struct, so changes inside the method are local and don't affect the caller's original value.
- The convention: if a method needs to modify the struct (or the struct is large and copying it would be wasteful), use a pointer receiver. Read-only methods on small structs can use value receivers.

You'll see this pattern throughout `go-api/items.go` — e.g. `func (s *itemStore) create(...)` uses a pointer receiver because it mutates the store.

---

## Interfaces — Structural Typing

Go interfaces describe a set of methods a type must implement — but unlike TypeScript's `implements` keyword, **Go interfaces are satisfied implicitly**. There's no explicit "this struct implements this interface" declaration; if a type has all the right methods, it automatically satisfies the interface.

```go
type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

// Circle automatically satisfies Shape — no "implements Shape" needed
var s Shape = Circle{Radius: 2}
```

This is structurally similar to how TypeScript's structural typing works for object shapes (`interface Shape { area(): number }` — any object with an `area()` method matches), but Go extends that same idea to concrete types with methods, not just object literals.

The most important interface in this tutorial is `http.Handler` (from the standard library), which just requires a `ServeHTTP(w http.ResponseWriter, r *http.Request)` method — see [What is chi?](#what-is-chi) for why that matters.

---

## Slices and Maps

- A **slice** (`[]T`) is Go's dynamic array — the rough equivalent of a TypeScript `Array<T>` / `T[]` or a Python `list`.
- A **map** (`map[K]V`) is Go's hash map — the equivalent of a TypeScript `Map<K, V>` / plain object, or a Python `dict`.

```go
names := []string{"alice", "bob"}       // slice literal
names = append(names, "carol")          // append (returns a new slice header)

ages := map[string]int{"alice": 30}     // map literal
ages["bob"] = 25                        // set
value, ok := ages["carol"]              // read + "does it exist?" check
```

The `value, ok := ages["carol"]` pattern is idiomatic Go: reading a missing map key doesn't throw or return `undefined`/`None` — it silently returns the value type's zero value, so `ok` is how you distinguish "key present with a zero value" from "key absent". You'll see this exact pattern in `go-api/items.go`'s `itemStore.get`.

---

## Goroutines and Channels

This is the concept most different from anything in single-threaded JavaScript or GIL-bound Python (outside of `asyncio`/`threading`).

A **goroutine** is a lightweight, independently-scheduled function execution — not an OS thread itself, but multiplexed onto a small pool of OS threads by the Go runtime. You start one with the `go` keyword:

```go
go doSomething() // runs concurrently, doesn't block the caller
```

Compare this to what you already know:

| | JavaScript/TypeScript | Python | Go |
|---|---|---|---|
| **Concurrency unit** | Promise / `async function` | `async def` coroutine, or OS thread (`threading`) | Goroutine |
| **Scheduler** | Single-threaded event loop | Single-threaded event loop (asyncio) or GIL-limited threads | Go runtime scheduler across multiple OS threads |
| **True parallelism (multi-core)** | No (needs Worker Threads / cluster) | No for `asyncio`/`threading` (GIL); yes for `multiprocessing` | **Yes, by default** |
| **How you start one** | `async function f() {}`, awaited or not | `async def f()`, scheduled with `asyncio` | `go f()` |
| **Cost per unit** | Cheap (callback-based) | Cheap (coroutines) / expensive (OS threads) | Very cheap (~2KB starting stack, grows as needed) |

The key takeaway: because Go goroutines are scheduled across real OS threads, CPU-bound work can run truly in parallel across multiple cores — something Node.js's single-threaded event loop and Python's GIL cannot do without extra tooling (worker threads, multiprocessing, etc.).

**In this tutorial's API, you don't have to write `go` yourself.** Go's standard `net/http` server automatically spins up a new goroutine for every incoming HTTP request — see [How the API Works](#how-the-api-works). This is why the in-memory item store needs a `sync.Mutex` to guard against multiple requests (i.e. multiple goroutines) reading and writing it at the same time — without that lock, concurrent requests could corrupt the shared map.

Channels (`chan T`) are Go's built-in tool for goroutines to communicate safely, but this tutorial's API is simple enough that it doesn't need one directly — it's worth knowing they exist as the next concept to learn once goroutines feel comfortable.

---

## Struct Tags

You'll see snippets like this throughout the API code:

```go
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name" validate:"required,min=2,max=100"`
}
```

The backtick-delimited text after each field is a **struct tag** — a plain string attached to the field that libraries can read via reflection at runtime. Go itself does nothing with these strings; they're metadata that specific packages opt into reading:

- `encoding/json` (standard library) reads `json:"..."` tags to decide how to name each field when marshalling to/from JSON — the equivalent of how Pydantic maps Python field names to JSON keys, or how a TypeScript type just naturally matches the JSON shape.
- `go-playground/validator` (third-party) reads `validate:"..."` tags to know what rules to check — see [Validation with go-playground/validator](#validation-with-go-playgroundvalidator).

This tag-based approach is how Go achieves declarative, schema-like behavior without a fluent builder API like Zod's `z.object({...})` or a base class like Pydantic's `BaseModel` — the "schema" is just annotations on a plain struct.

---

# Part 2 — The CRUD API

With the language basics in place, let's look at how they come together in a real, working CRUD API.

## What is chi?

[chi](https://github.com/go-chi/chi) is a **lightweight, idiomatic HTTP router** for Go. Go's standard library (`net/http`) already gives you a working HTTP server, but its built-in `ServeMux` router has no support for path parameters (`/api/items/{id}`) or a composable middleware chain — chi adds exactly those two things and nothing else.

chi is unopinionated routing plus a middleware pipeline, built directly on top of `net/http` with zero framework "magic" — if you've used a minimal web framework in another language (e.g. Express in Node.js, Flask in Python), the mental model will feel familiar: register handlers on paths, stack middleware in front of them.

Key selling points:

| Feature | Description |
|---|---|
| **100% compatible with `net/http`** | A chi router *is* an `http.Handler` (see [Interfaces](#interfaces--structural-typing) above) — any standard library or third-party `net/http` middleware works unmodified. |
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
| **Concurrency model** | Goroutine per request, automatically, via `net/http` |
| **Port** | 3005 |

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
| `go run .` | `node index.js` / `ts-node index.ts` | `python main.py` |
| `go build` | `tsc` (but produces a single native binary, not JS) | (no direct equivalent — Python isn't compiled) |

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

There is no `--reload`/`--watch` flag built in — Go compiles very fast, so most Go developers just re-run `go run .` after a save, or use a third-party watcher like [air](https://github.com/air-verse/air) for auto-reload during development.

---

## How the API Works

The Go server is split across three files in **`go-api/`**:

```
go-api/
├── main.go       # entry point: router, middleware, server startup
├── items.go      # in-memory store + HTTP handlers
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

- `chi.NewRouter()` creates the router — it satisfies the `http.Handler` interface (see [Interfaces](#interfaces--structural-typing)), so it can be passed straight to `http.ListenAndServe`.
- `r.Use(...)` registers middleware, run in order for every request — each one is a function that wraps the next handler in the chain.
- `middleware.Logger` and `middleware.Recoverer` ship with chi itself; `cors.Handler(...)` comes from the separate `go-chi/cors` package. `middleware.Recoverer` is what turns a `panic` (see [Error Handling](#error-handling--no-exceptions)) into a clean 500 response instead of crashing the whole process.
- `r.Route("/api/items", api.routes)` mounts a sub-router at a base path — every route registered inside `api.routes` is automatically prefixed with `/api/items`.
- `http.ListenAndServe` starts the HTTP server and blocks forever, spawning a new goroutine for every incoming request automatically (see [Goroutines and Channels](#goroutines-and-channels)) — you never write `go` yourself for this.

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

`Item` is a plain struct (see [Structs](#structs--gos-classes)) with `json:"..."` struct tags (see [Struct Tags](#struct-tags)) controlling how it's serialized. `itemStore` wraps a `map[int]*Item` (see [Slices and Maps](#slices-and-maps)) — a plain in-memory collection, reset on restart, good enough for a tutorial but not for production.

Because every incoming request runs on its own goroutine, multiple requests could read and write `items` at the same time. The `sync.Mutex` field (`mu`) prevents that: every store method calls `s.mu.Lock()` before touching `items` and `defer s.mu.Unlock()` to release it when the method returns.

```go
func (s *itemStore) create(name, description string) *Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := &Item{ID: s.nextID, Name: name, Description: description}
	s.items[item.ID] = item
	s.nextID++
	return item
}
```

Notice the pointer receiver `(s *itemStore)` (see [Methods and Pointer Receivers](#methods-and-pointer-receivers)) — every store method needs to mutate `s`, so it must take a pointer.

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

Key things to notice, tying back to Part 1:

- **`{id}`** is chi's path-parameter syntax for URL segments. Read it back with `chi.URLParam(r, "id")`.
- **No automatic body parsing.** Go's standard library requires you to decode the JSON body yourself with `json.NewDecoder(r.Body).Decode(&body)` in every handler that needs one.
- **`if err := ...; err != nil { ... }`** is the error-checking pattern from [Error Handling](#error-handling--no-exceptions), used at every step that can fail: decoding JSON, validating the struct, looking up the item.
- **No automatic error responses.** Every handler writes its own status code and JSON error body via small `writeJSON` / `writeError` helpers — there is no built-in global exception filter, so this tutorial defines that convention itself.
- **`http.ResponseWriter` / `*http.Request`** are the two parameters every handler receives — everything a handler needs to read the request and write the response.

### 4. Entry point details

There is no separate "dev" vs "production" command — `go run .` compiles and runs the program in one step, and `go build` produces a single self-contained native binary with no external runtime dependency (no interpreter, no virtual environment, nothing to install on the target machine besides the binary itself).

---

## Validation with go-playground/validator

Go has no built-in schema validation and no runtime type-introspection library baked into the language. The community standard is [`go-playground/validator`](https://github.com/go-playground/validator), which reads the `validate:"..."` struct tags introduced in [Struct Tags](#struct-tags) to describe validation rules declaratively — you annotate each struct field, then call `validate.Struct(...)` to check an instance against those rules.

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

`PatchItemInput` using `*string` instead of `string` is a direct application of [zero values](#variables-types-and-): an omitted `string` field would silently decode to `""` (its zero value), which is indistinguishable from the client explicitly sending an empty string. A `*string` defaults to `nil` when omitted, which the handler code below can check for explicitly.

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

Describe the shape once with struct tags, validate every request against it with `validate.Struct(...)`, and read a structured error back out if it fails — the same three-step pattern used by schema-validation libraries in most languages (Zod in TypeScript, Pydantic in Python), just expressed through struct tags instead of a fluent builder API or a base class.

---

## How the UI Connects

The `GoCrud` component (`app/components/GoCrud.tsx`) is a plain React client component that talks directly to the Go server on port 3005:

```ts
const GO_API = "http://localhost:3005";
const res = await fetch(`${GO_API}/api/items`);
```

The Go API returns `{"error": "..."}` on failure, so `GoCrud` reads `err.error` to display a readable message in the UI. `GoCrud` also duplicates the same validation rules in plain TypeScript (`validateName`, `validateDescription`) so the form can show instant feedback before ever making a network request — the source of truth still lives on the server in `go-api/schemas.go`.

Because the Next.js app (port 3000) and the Go API (port 3005) run on different origins, the Go server enables CORS for `http://localhost:3000` via the `cors.Handler(...)` middleware registered in `main.go` (see [How the API Works](#how-the-api-works) above).

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

6. **Read the source alongside Part 1:** Open `go-api/main.go`, `go-api/items.go`, and `go-api/schemas.go` and trace how each language concept from Part 1 — structs, pointer receivers, error checking, goroutines-per-request, struct tags — shows up in real, working code.

7. **Compare startup models:** Run `go build -o bin/go-api . && ./bin/go-api` and notice there is no `node_modules`, no interpreter, and no separate install step at runtime — the compiled binary is completely self-contained.

8. **Experiment with concurrency:** Add a `time.Sleep(2 * time.Second)` at the top of the `list` handler in `go-api/items.go`, then open two browser tabs and click refresh on both at nearly the same time. Notice both requests complete in roughly 2 seconds total, not 4 — because each is handled on its own goroutine, running concurrently rather than queued one after another.

---
