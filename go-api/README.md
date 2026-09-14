# Go CRUD API

A simple CRUD API built with [Go](https://go.dev) and [chi](https://github.com/go-chi/chi), with request validation via [go-playground/validator](https://github.com/go-playground/validator). Runs on `http://localhost:3005`.

## Setup

Requires Go 1.21+ (developed against Go 1.27).

```bash
go mod download
```

## Run

```bash
go run .
```

## Build

```bash
go build -o bin/go-api .
./bin/go-api
```

## Test the API

```bash
curl http://localhost:3005/api/items

curl -X POST http://localhost:3005/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Go Item","description":"Created via curl"}'
```

See [../docs/go.md](../docs/go.md) for a full guided tour of the code.
