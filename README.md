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
├── go-api/          # Go REST API (chi router, Go modules)          → port 3005
├── docs/            # Per-chapter documentation
│   └── go.md
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
> - Go: `cd go-api && go run .` → http://localhost:3005

---

## Chapter 1 — Go

If you're new to Go, **start with [docs/go.md](docs/go.md)** — it's written specifically for TypeScript/Python developers and covers, in order:

**Part 1 — Go language basics**, explained by comparison to what you already know:
- Why Go looks and behaves differently from TS/Python (compiled vs. interpreted, static typing, etc.)
- Packages, variables, zero values, and `:=`
- Functions with multiple return values
- Error handling without exceptions (`if err != nil`, no `try`/`catch`)
- Structs and methods (Go's alternative to classes)
- Interfaces (structural typing, satisfied implicitly)
- Slices and maps
- **Goroutines** — Go's lightweight concurrency primitive, and how it differs from JS's event loop and Python's `asyncio`/GIL
- Struct tags — how Go expresses metadata like JSON field names and validation rules

**Part 2 — The CRUD API**, built with [chi](https://github.com/go-chi/chi) (a lightweight HTTP router) and [go-playground/validator](https://github.com/go-playground/validator) (struct-tag-based request validation):
- Project setup with Go modules
- How the router, handlers, and in-memory store work
- Request validation
- How the Next.js UI talks to the Go server

Quick test with `curl` once the server is running:

```bash
curl http://localhost:3005/api/items

curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Go Item","description":"Created via curl"}'
```

---

## Adding a New Chapter

This project is set up to grow one language at a time. To add the next chapter (e.g. Rust):

1. Create a new API project (e.g. `rust-api/`) on its own port.
2. Add a new CRUD component under `app/components/` (e.g. `RustCrud.tsx`), following the pattern in `app/components/GoCrud.tsx`.
3. Add the new tab to the `MainTab` type and `TABS` array in `app/components/TabNavigation.tsx`.
4. Render the new component conditionally in `app/page.tsx`, following the existing `go-crud` example.
5. Add a new doc file under `docs/` (e.g. `docs/rust.md`), written the same way as `docs/go.md`: language fundamentals first (compared to whatever the reader already knows), then the API walkthrough. Link it from this README.
6. Add the new server to `scripts/start-servers.sh` so `./scripts/start-servers.sh` starts it alongside the others.
