// Package concurrency demonstrates Go's core concurrency building blocks —
// goroutines, sync.WaitGroup, and channels — in one small, dependency-free
// function: FanOut.
//
// It's deliberately kept separate from any HTTP or CLI concerns so it can be
// used two different ways in this tutorial:
//   - cmd/concurrency-demo runs it standalone.
//   - internal/api composes it into a real HTTP handler
//     (GET /api/items/{id}/enrich), showing the exact same building block
//     doing real work inside a bigger program.
//
// See docs/go/02-concurrency.md for the full walkthrough.
package concurrency

import "sync"

// Check is a single unit of (simulated) work to run concurrently — e.g.
// "call the inventory service" or "call the pricing service". In a real
// system, Run would make a network call; here it's a plain function so the
// tutorial has no external dependencies.
type Check struct {
	Name string
	Run  func() (string, error)
}

// Result is what came back from running one Check.
type Result struct {
	Name  string
	Value string
	Err   error
}

// FanOut runs every Check concurrently on its own goroutine and waits for
// all of them to finish before returning — the "fan-out, fan-in" pattern.
//
// How it works, tying back to docs/go/02-concurrency.md:
//   - results is a buffered channel with exactly len(checks) slots, so every
//     goroutine can send its Result and return immediately without blocking,
//     even if nothing has read from the channel yet.
//   - wg (a sync.WaitGroup) is how the main goroutine knows when every
//     worker goroutine has finished: wg.Add(1) before starting each one,
//     wg.Done() when it completes, wg.Wait() blocks until the count reaches
//     zero.
//   - A dedicated goroutine calls wg.Wait() and then close(results) — closing
//     the channel is what lets the `for result := range results` loop below
//     terminate instead of blocking forever waiting for one more value.
func FanOut(checks []Check) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, len(checks))

	for _, check := range checks {
		wg.Add(1)
		go func(c Check) {
			defer wg.Done()
			value, err := c.Run()
			results <- Result{Name: c.Name, Value: value, Err: err}
		}(check)
	}

	// Close the channel only once every goroutine above has sent its
	// result — otherwise the range loop below could exit early and drop
	// results, or block forever if a goroutine hasn't sent yet.
	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]Result, 0, len(checks))
	for result := range results {
		out = append(out, result)
	}
	return out
}
