# Part 3 — Testing

← [Back to Chapter 1 index](README.md) · ← [Part 2 — Concurrency](02-concurrency.md)

Go treats testing as a first-class part of the toolchain — no separate test runner to install, no config file to write. This part covers the standard-library `testing` package, table-driven tests, the third-party `testify` assertion library, and `httptest` for testing HTTP handlers without a real network connection.

## Table of Contents

- [The testing package](#the-testing-package)
- [Table-driven tests and subtests](#table-driven-tests-and-subtests)
- [testify: assert and require](#testify-assert-and-require)
- [Testing HTTP handlers with httptest](#testing-http-handlers-with-httptest)
- [Testing concurrent code](#testing-concurrent-code)
- [Running the tests](#running-the-tests)
- [Try It Yourself](#try-it-yourself)

---

## The testing package

A Go test is just a function named `TestXxx` taking a `*testing.T`, living in a file named `xxx_test.go`:

```go
// store_test.go
func TestStore_Get_NotFound(t *testing.T) {
	store := items.NewStore()

	_, ok := store.Get(999)

	if ok {
		t.Fatalf("expected ok=false for a missing item, got true")
	}
}
```

- `t.Fatalf(...)` — reports the failure and stops the current test function immediately (like `assert` + early return combined).
- `t.Errorf(...)` — reports the failure but lets the rest of the test function keep running (useful when checking several independent things and you want to see every failure, not just the first).
- Test files live alongside the code they test (`store.go` and `store_test.go` in the same package directory) — there's no separate `tests/` folder the way you might see in a TS or Python project.

---

## Table-driven tests and subtests

Instead of writing one `TestXxx` function per scenario, idiomatic Go collects scenarios into a slice of structs and loops over them — a **table-driven test**:

```go
func TestCreateItemInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   items.CreateItemInput
		wantErr bool
	}{
		{name: "valid input", input: items.CreateItemInput{Name: "Valid Name"}, wantErr: false},
		{name: "name too short", input: items.CreateItemInput{Name: "x"}, wantErr: true},
		{name: "name missing", input: items.CreateItemInput{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := items.Validate.Struct(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
```

- `t.Run(tt.name, func(t *testing.T) { ... })` registers each table entry as its own named **subtest** — `go test -v` reports each one individually (`TestCreateItemInput_Validation/name_too_short`), and a single failing case doesn't stop the others from running.
- This pattern is used throughout `go-api/internal/items/store_test.go` and `schemas_test.go` — see those files for more examples.

---

## testify: assert and require

Go's standard library deliberately ships no assertion library — comparisons are just `if` statements plus `t.Errorf`/`t.Fatalf`. In practice, most real-world Go projects add [`testify`](https://github.com/stretchr/testify) on top, because it removes a lot of that boilerplate. This tutorial uses it too — it's a normal dependency in `go-api/go.mod`, added the same way as `chi` or `validator`.

```go
import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	err := doSomething()

	require.NoError(t, err)              // stops the test immediately on failure
	assert.Equal(t, "expected", actual)   // reports failure but keeps going
	assert.InDelta(t, 3.14, pi, 0.01)     // approximate float comparison
}
```

- **`require`** stops the test function on failure (like `t.Fatalf`) — use it for setup steps where continuing wouldn't make sense (e.g. `require.NoError(t, err)` right after something that must succeed for the rest of the test to be meaningful).
- **`assert`** reports the failure but lets the test keep running (like `t.Errorf`) — use it for the actual assertions you're checking, especially when you want to see every mismatch in one run instead of stopping at the first.
- Compare to Python: `assert.Equal(t, want, got)` plays a similar role to `assertEqual` in `unittest`, or a bare `assert want == got` in `pytest`.
- Compare to TypeScript: similar role to `expect(got).toBe(want)` in Jest/Vitest — Go just doesn't build this in, so `testify` is the community-standard addition.

---

## Testing HTTP handlers with httptest

The standard library's `net/http/httptest` package lets you test real HTTP handlers without opening an actual network port to the outside world:

```go
func TestList_ReturnsSeededItem(t *testing.T) {
	srv := httptest.NewServer(api.NewRouter())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/items")
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
```

- `httptest.NewServer(handler)` starts a real (but local, ephemeral-port) HTTP server backed by any `http.Handler` — here, the exact same `api.NewRouter()` that `cmd/go-api/main.go` uses in production, so the test exercises the real middleware + routing + handler stack, not a mock.
- `srv.Close()` shuts it down — always deferred immediately after creation, mirroring the `defer` pattern from [Part 1](01-basics.md#defer).
- This is different from testing a `Store` directly (no HTTP involved at all) — `httptest` is specifically for verifying the HTTP layer itself: status codes, JSON bodies, routing.

See `go-api/internal/api/items_handlers_test.go` for the full set of handler tests (create/list/get/update/patch/delete), and `go-api/internal/api/enrich_handler_test.go` for the concurrency endpoint.

---

## Testing concurrent code

Concurrency bugs are notoriously hard to catch by reading code, so it helps to write tests that assert on the actual behavior rather than trusting the implementation:

```go
func TestFanOut_RunsChecksConcurrently(t *testing.T) {
	checks := []concurrency.Check{
		sleepyCheck("inventory", 30*time.Millisecond, "in-stock"),
		sleepyCheck("pricing", 30*time.Millisecond, "$19.99"),
		sleepyCheck("reviews", 30*time.Millisecond, "4.5 stars"),
	}

	start := time.Now()
	results := concurrency.FanOut(checks)
	elapsed := time.Since(start)

	require.Len(t, results, 3)
	assert.Less(t, elapsed, 70*time.Millisecond, "expected checks to run concurrently, not sequentially")
}
```

Three checks that each sleep 30ms would take ~90ms if run sequentially; asserting the real elapsed time stays well under that (with headroom for slow CI machines) is a concrete, automated way to prove `FanOut` genuinely runs its checks concurrently — see `go-api/internal/concurrency/fanout_test.go` and [Part 2](02-concurrency.md#putting-it-together-fanout).

Go's test runner also has a built-in **race detector** for catching exactly the kind of shared-memory bugs concurrent code is prone to:

```bash
go test -race ./...
```

`-race` instruments the binary to flag any place two goroutines access the same memory without proper synchronization (e.g. a missing `sync.Mutex` lock) — it's slower than a normal test run, but extremely good at catching bugs that only show up intermittently otherwise.

---

## Running the tests

```bash
# From go-api/
go test ./...            # run every test in every package
go test -v ./...          # verbose — show each test/subtest name and result
go test -race ./...       # also run the race detector
go test ./internal/items  # only this package
```

`./...` is Go's "this package and every package below it" wildcard — the same role `**/*.test.ts` glob patterns play for Jest, or `pytest` discovering every `test_*.py` file recursively.

---

## Try It Yourself

1. From `go-api/`, run the whole suite: `go test -v ./...`
2. Run just the concurrency package: `go test -v ./internal/concurrency`
3. Deliberately break something — e.g. change `Store.Patch` in `go-api/internal/items/store.go` to always overwrite `Description` even when the pointer is `nil` — and rerun `go test ./internal/items`. Confirm the table-driven test in `store_test.go` catches it.
4. Run `go test -race ./...` and confirm it passes.

---

Next: [Part 4 — The CRUD API](04-api.md)
