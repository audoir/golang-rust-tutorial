# Golang / Rust API Tutorial

A hands-on tutorial for developers who already know **TypeScript** and/or **Python** and want to learn systems languages by building the same simple CRUD API in each one. Each chapter lives in its own sub-folder and is demonstrated through a shared Next.js UI.

**Chapter 1 covers Go.** It starts with Go language fundamentals — syntax, error handling, structs, goroutines — explained by comparison to TypeScript/Python, then walks through a real CRUD API built with those concepts. Future chapters will add **Rust** as a second tab, so you can compare all three languages side by side.

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Quick Start](#quick-start)
3. [Chapter 1 — Go](#chapter-1--go)
4. [Adding a New Chapter](#adding-a-new-chapter)

---

## Project Structure

```
golang-rust-tutorial/
├── app/             # Next.js app (shared UI shell + tab navigation)
│   ├── components/
│   │   ├── GoCrud.tsx        # Chapter 1 — Go CRUD demo
│   │   ├── PageHeader.tsx
│   │   └── TabNavigation.tsx
│   ├── layout.tsx
│   └── page.tsx
├── go-api/          # Go REST API                                  → port 3005
│   ├── cmd/
│   │   ├── go-api/            # the real server binary
│   │   └── concurrency-demo/  # standalone goroutines/channels demo
│   └── internal/
│       ├── items/              # Item, Store, validated input structs
│       ├── concurrency/         # FanOut — goroutines + WaitGroup + channel
│       └── api/                 # chi router + HTTP handlers
├── docs/            # Per-chapter documentation
│   └── go/
│       ├── README.md          # Chapter 1 index
│       ├── 01-basics.md       # Part 1 — Go language basics
│       ├── 02-concurrency.md  # Part 2 — Goroutines, WaitGroup, channels
│       ├── 03-testing.md      # Part 3 — testing, testify, httptest
│       └── 04-api.md          # Part 4 — the CRUD API
└── scripts/
    └── start-servers.sh  # Starts all tutorial servers (Ctrl-C to stop all)
```

---

## Quick Start

### Prerequisites

- **Node.js ≥ 20.9** (required by Next.js 16)
- **npm** (comes with Node.js)
- **Go ≥ 1.21** (for the Go chapter)

Install Go if you don't have it (macOS via Homebrew):

```bash
brew install go
```

### Install dependencies

```bash
# Next.js (repo root)
npm install

# Go — download dependencies once (go.sum is committed)
cd go-api && go mod download
```

### Run all servers

From the **repo root**:

```bash
./scripts/start-servers.sh
```

Then open **http://localhost:3000** in your browser.

> The script starts every tutorial server in the background and kills them all when you press Ctrl-C.
>
> You can also start each server manually:
> - Next.js: `npm run dev` → http://localhost:3000
> - Go: `cd go-api && go build -o bin/go-api ./cmd/go-api && ./bin/go-api` → http://localhost:3005 (or `go run ./cmd/go-api` for quick iteration)

---

## Chapter 1 — Go

If you're new to Go, **start with [docs/go/README.md](docs/go/README.md)** — it's written specifically for TypeScript/Python developers, split into four progressively more advanced parts:

1. **[Go Language Basics](docs/go/01-basics.md)** — why Go looks and behaves differently from TS/Python, packages/variables/zero values, functions with multiple return values, error handling without exceptions, `defer`, structs and methods, interfaces, slices/maps, and struct tags.
2. **[Concurrency](docs/go/02-concurrency.md)** — goroutines, `sync.WaitGroup`, and channels, demonstrated in a small standalone program (`go run ./cmd/concurrency-demo` in `go-api/`) before they're used for real in the API. Also introduces the `cmd/`/`internal/` project layout.
3. **[Testing](docs/go/03-testing.md)** — the `testing` package, table-driven tests, `testify`'s `assert`/`require`, and `httptest` for testing HTTP handlers.
4. **[The CRUD API](docs/go/04-api.md)** — built with [chi](https://github.com/go-chi/chi) (a lightweight HTTP router) and [go-playground/validator](https://github.com/go-playground/validator) (struct-tag-based request validation), composing everything from the earlier parts — including a concurrency showcase endpoint (`GET /api/items/{id}/enrich`) and running the server as a compiled binary with graceful shutdown, instead of `go run`.

Quick test with `curl` once the server is running:

```bash
curl http://localhost:3005/api/items

curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Go Item","description":"Created via curl"}'

# Concurrency showcase — fans out 3 checks concurrently, see docs/go/02-concurrency.md
curl http://localhost:3005/api/items/1/enrich
```