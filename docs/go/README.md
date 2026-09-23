# Chapter 1 — Go

← [Back to top-level README](../../README.md)

This chapter is written for developers who already know **TypeScript** and/or **Python** and are learning **Go** for the first time. It's split into four parts, each building on the last — from basic syntax, to concurrency, to testing, to a real CRUD API that ties everything together.

Read them in order the first time through; each part links back to earlier concepts as they come up again in more advanced code.

## Parts

1. **[Go Language Basics](01-basics.md)** — syntax, types, structs, methods, interfaces, slices/maps, error handling, struct tags. Everything you need before touching concurrency or the API.
2. **[Concurrency](02-concurrency.md)** — goroutines, `sync.WaitGroup`, and channels, demonstrated in a small standalone program with no HTTP server involved. Also introduces the `cmd/` + `internal/` project layout used by the rest of this chapter.
3. **[Testing](03-testing.md)** — the `testing` package, table-driven tests, `testify`'s `assert`/`require`, and `httptest` for testing HTTP handlers.
4. **[The CRUD API](04-api.md)** — a real, runnable REST API built with [chi](https://github.com/go-chi/chi) and [go-playground/validator](https://github.com/go-playground/validator), composing the store, the concurrency helper, and everything from Parts 1–3 into one working server — including a concurrency showcase endpoint, graceful shutdown, and running it as a compiled binary instead of `go run`.

## Where the code lives

```
go-api/
├── cmd/
│   ├── go-api/               # the real server binary (Part 4)
│   └── concurrency-demo/     # standalone concurrency demo (Part 2)
└── internal/
    ├── concurrency/          # FanOut — goroutines + WaitGroup + channel (Part 2)
    ├── items/                # Item, Store, validated input structs (Parts 1, 3, 4)
    └── api/                  # chi router + HTTP handlers (Part 4)
```

Start with [Go Language Basics](01-basics.md).
