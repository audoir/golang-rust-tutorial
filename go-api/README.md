# Go CRUD API

A simple CRUD API built with [Go](https://go.dev) and [chi](https://github.com/go-chi/chi), with request validation via [go-playground/validator](https://github.com/go-playground/validator). Runs on `http://localhost:3005`.

## Project layout

```
go-api/
├── cmd/
│   ├── go-api/               # the real server binary
│   └── concurrency-demo/     # standalone goroutines/channels demo
└── internal/
    ├── items/                 # Item, Store, validated input structs
    ├── concurrency/           # FanOut — goroutines + WaitGroup + channel
    └── api/                   # chi router + HTTP handlers
```

## Setup

Requires Go 1.21+ (developed against Go 1.27).

```bash
go mod download
```

## Run

```bash
# Quick iteration
go run ./cmd/go-api

# Build then run — the primary workflow (required for graceful shutdown to work)
go build -o bin/go-api ./cmd/go-api
./bin/go-api

# Custom port
./bin/go-api -port 3006
```

Stop the server with Ctrl-C (or `kill`) — it shuts down gracefully, finishing in-flight requests before exiting.

## Try the concurrency demo

A small, standalone program with no HTTP server involved, demonstrating goroutines + `sync.WaitGroup` + channels:

```bash
go run ./cmd/concurrency-demo
```

## Test

```bash
go test ./...
go test -v ./...       # verbose
go test -race ./...    # with the race detector
```

## Test the API

```bash
curl http://localhost:3005/api/items

curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Go Item","description":"Created via curl"}'

# Concurrency showcase endpoint — fans out 3 checks concurrently
curl http://localhost:3005/api/items/1/enrich
```

See [../docs/go/README.md](../docs/go/README.md) for a full guided tour of the code.
