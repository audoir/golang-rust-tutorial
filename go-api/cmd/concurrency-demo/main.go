// Command concurrency-demo is a standalone program demonstrating goroutines,
// sync.WaitGroup, and channels via go-api/internal/concurrency.FanOut.
//
// Run it directly:
//
//	go run ./cmd/concurrency-demo
//
// See docs/go/02-concurrency.md for the full walkthrough. The same FanOut
// function is reused for real in internal/api/enrich_handler.go.
package main

import (
	"fmt"
	"time"

	"go-api/internal/concurrency"
)

func main() {
	// Three simulated slow checks, each standing in for a call to a
	// separate downstream service. If these ran one after another
	// (sequentially), the program would take roughly 300+250+200 = 750ms.
	checks := []concurrency.Check{
		{
			Name: "inventory",
			Run: func() (string, error) {
				time.Sleep(300 * time.Millisecond)
				return "42 units in stock", nil
			},
		},
		{
			Name: "pricing",
			Run: func() (string, error) {
				time.Sleep(250 * time.Millisecond)
				return "$19.99", nil
			},
		},
		{
			Name: "reviews",
			Run: func() (string, error) {
				time.Sleep(200 * time.Millisecond)
				return "4.5 stars (128 reviews)", nil
			},
		},
	}

	fmt.Println("Running 3 checks concurrently (would take ~750ms sequentially)...")

	start := time.Now()
	results := concurrency.FanOut(checks)
	elapsed := time.Since(start)

	for _, result := range results {
		if result.Err != nil {
			fmt.Printf("  %-10s error: %v\n", result.Name, result.Err)
			continue
		}
		fmt.Printf("  %-10s %s\n", result.Name, result.Value)
	}

	fmt.Printf("\nDone in %v — roughly the time of the single slowest check (~300ms),\n", elapsed)
	fmt.Println("not the sum of all three, because FanOut ran them concurrently on their own goroutines.")
}
