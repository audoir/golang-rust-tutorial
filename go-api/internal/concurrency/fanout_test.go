package concurrency_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-api/internal/concurrency"
)

// sleepyCheck builds a Check that sleeps for the given duration before
// returning a value — a stand-in for a slow downstream call (a database
// query, another service's API, etc.).
func sleepyCheck(name string, delay time.Duration, value string) concurrency.Check {
	return concurrency.Check{
		Name: name,
		Run: func() (string, error) {
			time.Sleep(delay)
			return value, nil
		},
	}
}

func TestFanOut_RunsChecksConcurrently(t *testing.T) {
	// Three checks that would take ~90ms combined if run one after another.
	// If FanOut truly runs them concurrently, the whole call should take
	// roughly as long as the single slowest check (~30ms), not the sum.
	checks := []concurrency.Check{
		sleepyCheck("inventory", 30*time.Millisecond, "in-stock"),
		sleepyCheck("pricing", 30*time.Millisecond, "$19.99"),
		sleepyCheck("reviews", 30*time.Millisecond, "4.5 stars"),
	}

	start := time.Now()
	results := concurrency.FanOut(checks)
	elapsed := time.Since(start)

	require.Len(t, results, 3)
	// Generous upper bound: sequential execution would take ~90ms; allowing
	// up to 70ms leaves comfortable headroom for slow CI machines while
	// still failing if the checks were run one at a time.
	assert.Less(t, elapsed, 70*time.Millisecond, "expected checks to run concurrently, not sequentially")

	byName := resultsByName(results)
	assert.Equal(t, "in-stock", byName["inventory"].Value)
	assert.Equal(t, "$19.99", byName["pricing"].Value)
	assert.Equal(t, "4.5 stars", byName["reviews"].Value)
}

func TestFanOut_CollectsErrorsPerCheck(t *testing.T) {
	boom := errors.New("boom")
	checks := []concurrency.Check{
		{Name: "ok-check", Run: func() (string, error) { return "fine", nil }},
		{Name: "failing-check", Run: func() (string, error) { return "", boom }},
	}

	results := concurrency.FanOut(checks)
	byName := resultsByName(results)

	require.NoError(t, byName["ok-check"].Err)
	assert.Equal(t, "fine", byName["ok-check"].Value)

	require.Error(t, byName["failing-check"].Err)
	assert.Equal(t, boom, byName["failing-check"].Err)
}

func TestFanOut_EmptyInput(t *testing.T) {
	results := concurrency.FanOut(nil)
	assert.Empty(t, results)
}

// resultsByName re-indexes FanOut's output by name for assertions, since
// FanOut makes no guarantee about result order — goroutines can finish in
// any order.
func resultsByName(results []concurrency.Result) map[string]concurrency.Result {
	out := make(map[string]concurrency.Result, len(results))
	for _, r := range results {
		out[r.Name] = r
	}
	return out
}
