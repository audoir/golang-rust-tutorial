# Part 2 — Concurrency

← [Back to Chapter 1 index](README.md) · ← [Part 1 — Go Language Basics](01-basics.md)

This is the concept most different from anything in single-threaded JavaScript or GIL-bound Python (outside of `asyncio`/`threading`). We'll build a small, standalone program in this part — no HTTP server, no web framework, nothing else going on — so the concurrency behavior itself is easy to see before [Part 4](04-api.md) puts it to work inside a real API endpoint.

## Table of Contents

- [Goroutines](#goroutines)
- [The sync.WaitGroup](#the-syncwaitgroup)
- [Channels](#channels)
- [Putting it together: FanOut](#putting-it-together-fanout)
- [`cmd/` and `internal/`: organizing a real Go project](#cmd-and-internal-organizing-a-real-go-project)
- [Try It Yourself](#try-it-yourself)

---

## Goroutines

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

**In this tutorial's API (Part 4), you don't have to write `go` yourself for basic request handling.** Go's standard `net/http` server automatically spins up a new goroutine for every incoming HTTP request. This is why the in-memory item store needs a `sync.Mutex` (see [Structs — Go's "Classes"](01-basics.md#structs--gos-classes) and `go-api/internal/items/store.go`) — without that lock, concurrent requests (i.e. concurrent goroutines) could corrupt the shared map.

But `net/http`'s automatic per-request goroutine is a different thing from **starting your own goroutines deliberately** to run independent pieces of work in parallel *within* a single request — that's what the rest of this part covers.

---

## The sync.WaitGroup

If you start several goroutines, how does the calling code know when they're all done? A `sync.WaitGroup` is a simple counter built for exactly this:

```go
var wg sync.WaitGroup

for _, task := range tasks {
	wg.Add(1)          // increment the counter before starting each goroutine
	go func(t Task) {
		defer wg.Done() // decrement the counter when this goroutine finishes
		t.Run()
	}(task)
}

wg.Wait() // blocks until the counter reaches zero
```

- `wg.Add(1)` — increment the counter. Always call this *before* starting the goroutine, not inside it (otherwise `Wait()` might return before the goroutine even starts).
- `wg.Done()` — decrement the counter. Almost always called via `defer` at the top of the goroutine's function, so it runs even if the function returns early or panics (see [Defer](01-basics.md#defer)).
- `wg.Wait()` — blocks the calling goroutine until the counter is back to zero, i.e. until every `Done()` has been called.

Note the loop variable capture: `go func(t Task) { ... }(task)` passes `task` in as a parameter rather than closing over the loop variable directly — this avoids a classic bug where every goroutine ends up seeing the same (final) value of a loop variable.

---

## Channels

Channels (`chan T`) are Go's built-in tool for goroutines to communicate safely — a typed, one-directional-in-practice pipe for sending values between goroutines.

```go
ch := make(chan int)     // unbuffered channel
go func() {
	ch <- 353         // send (must happen in a goroutine, or this blocks forever)
}()
val := <-ch                // receive
close(ch)                  // signal "no more values will be sent"

ch2 := make(chan int, 1)   // buffered channel — send doesn't block until the buffer is full
```

- An **unbuffered** channel (`make(chan T)`) blocks the sender until a receiver is ready, and blocks the receiver until a sender is ready — it's a synchronization point as much as a data pipe.
- A **buffered** channel (`make(chan T, n)`) lets up to `n` sends happen without blocking, even if nothing has received yet — useful when you know exactly how many results are coming, which is exactly the case for the fan-out pattern below.
- `for v := range ch` reads values from a channel until it's `close`d — this is how a receiver knows to stop waiting for "just one more" value.
- Reading from or writing to a `nil` channel, or writing to a closed channel, blocks or panics respectively — channels are a sharp tool, used carefully.

---

## Putting it together: FanOut

`go-api/internal/concurrency/fanout.go` combines goroutines, `sync.WaitGroup`, and a buffered channel into one reusable function — the **fan-out, fan-in** pattern: start several independent units of work concurrently ("fan out"), then collect all their results back into one place ("fan in").

```go
type Check struct {
	Name string
	Run  func() (string, error)
}

type Result struct {
	Name  string
	Value string
	Err   error
}

func FanOut(checks []Check) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, len(checks)) // buffered so no goroutine blocks on send

	for _, check := range checks {
		wg.Add(1)
		go func(c Check) {
			defer wg.Done()
			value, err := c.Run()
			results <- Result{Name: c.Name, Value: value, Err: err}
		}(check)
	}

	go func() {
		wg.Wait()      // wait for every goroutine above to call Done()
		close(results) // then close, so the range loop below can terminate
	}()

	out := make([]Result, 0, len(checks))
	for result := range results {
		out = append(out, result)
	}
	return out
}
```

Walking through why each piece is there:

- The channel is **buffered with exactly `len(checks)` slots** — every worker goroutine can send its `Result` and return immediately, even before anything has started receiving.
- The **`wg.Wait()` + `close(results)` happens in its own goroutine**, not in the main flow — if we called `wg.Wait()` directly before the `for range results` loop, we'd deadlock: `Wait()` would block forever waiting for goroutines that are themselves blocked trying to send on a full buffer that nothing is draining yet. Running the "wait, then close" step concurrently with the "receive results" loop avoids that.
- `close(results)` is what lets `for result := range results` **terminate** instead of blocking forever waiting for "just one more" value.

### Try it standalone

```bash
cd go-api
go run ./cmd/concurrency-demo
```

`go-api/cmd/concurrency-demo/main.go` calls `FanOut` with three simulated slow checks (each just `time.Sleep`s for a different duration, standing in for a call to a separate downstream service — an inventory service, a pricing service, a reviews service). Run it and you'll see all three complete in roughly the time of the *slowest* single check, not the sum of all three — direct, visible proof that they ran concurrently rather than one after another.

---

## `cmd/` and `internal/`: organizing a real Go project

`FanOut` needs to be usable from two different places in this tutorial:

1. **`go-api/cmd/concurrency-demo/main.go`** — a tiny standalone program, so the concurrency behavior is easy to see in isolation.
2. **`go-api/internal/api/enrich_handler.go`** (see [Part 4](04-api.md)) — a real HTTP handler that fans out the exact same kind of checks for real.

A Go package can only have **one** `func main()`, so two runnable programs in the same module need two separate `package main` files — and the code they share (`FanOut`) needs to live somewhere both of them can import it from. This is exactly what Go's `cmd/` + `internal/` convention is for:

- **`cmd/<binary-name>/main.go`** — one folder per runnable program, each with a thin `main.go` that mostly just wires other packages together. This project has two: `cmd/go-api` (the real server) and `cmd/concurrency-demo` (this demo).
- **`internal/<package-name>/`** — shared library code, importable by anything else inside this module (`go-api/...`), but the Go compiler itself **refuses to let anything outside the module import it** — this is a real, enforced privacy boundary, not just a naming convention. There's no clean TypeScript or Python equivalent to this — the closest analogues (an ESLint import-boundaries rule, or a Python `_private` naming convention) are just linting conventions, not something the compiler/interpreter enforces.

```
go-api/
├── cmd/
│   ├── go-api/               # func main() — the real server
│   └── concurrency-demo/     # func main() — the standalone demo
└── internal/
    ├── concurrency/          # FanOut — imported by both binaries above
    ├── items/                # Item, Store, validated input structs
    └── api/                  # chi router + HTTP handlers (imports items + concurrency)
```

You'll see this same shape again in [Part 4](04-api.md), where `internal/api` composes `internal/items` and `internal/concurrency` together into the real CRUD server.

---

## Try It Yourself

1. From `go-api/`, run the standalone demo and note the total elapsed time printed at the end:
   ```bash
   go run ./cmd/concurrency-demo
   ```
2. Open `go-api/internal/concurrency/fanout.go` and try changing one check's `Run` function to `time.Sleep` for much longer (e.g. 2 seconds) — rerun the demo and confirm the total time tracks the *slowest* check, not the sum.
3. Open `go-api/internal/concurrency/fanout_test.go` and read `TestFanOut_RunsChecksConcurrently` — it asserts on wall-clock time directly, which is how [Part 3 — Testing](03-testing.md) proves this concurrency behavior automatically rather than just eyeballing printed output.

---

Next: [Part 3 — Testing](03-testing.md)
