# Golang / Rust API Tutorial

A hands-on tutorial that walks through building CRUD APIs in different systems languages. Each chapter lives in its own sub-folder and is demonstrated through a shared Next.js UI.

Chapter 1 covers **Go**. Future chapters will add **Rust** as a second tab, so you can compare the two languages side by side.

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

The Go chapter builds a simple CRUD API with [chi](https://github.com/go-chi/chi) (a lightweight HTTP router) and [go-playground/validator](https://github.com/go-playground/validator) (struct-tag-based request validation).

See **[docs/go.md](docs/go.md)** for the full guided tour: project setup with Go modules, how the router and handlers work, validation, and how the UI connects to the API.

Quick test with `curl` once the server is running:

```bash
curl http://localhost:3005/api/items

curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Go Item","description":"Created via curl"}'
```

---

## Adding a New Chapter

This project is set up to grow one language/framework at a time. To add the next chapter (e.g. Rust):

1. Create a new API project (e.g. `rust-api/`) on its own port.
2. Add a new CRUD component under `app/components/` (e.g. `RustCrud.tsx`), following the pattern in `app/components/GoCrud.tsx`.
3. Add the new tab to the `MainTab` type and `TABS` array in `app/components/TabNavigation.tsx`.
4. Render the new component conditionally in `app/page.tsx`, following the existing `go-crud` example.
5. Add a new doc file under `docs/` (e.g. `docs/rust.md`) and link it from this README.
6. Add the new server to `scripts/start-servers.sh` so `./scripts/start-servers.sh` starts it alongside the others.
