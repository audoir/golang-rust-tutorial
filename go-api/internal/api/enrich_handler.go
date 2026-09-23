package api

import (
	"fmt"
	"net/http"
	"time"

	"go-api/internal/concurrency"
)

// enrichResponse is the JSON shape returned by GET /api/items/{id}/enrich.
type enrichResponse struct {
	ItemID    int               `json:"itemId"`
	Checks    map[string]string `json:"checks"`
	Errors    map[string]string `json:"errors,omitempty"`
	ElapsedMs int64             `json:"elapsedMs"`
}

// GET /api/items/{id}/enrich — the concurrency showcase endpoint.
//
// It looks up the item, then fans out three independent, simulated
// downstream checks (inventory, pricing, reviews) concurrently using
// internal/concurrency.FanOut — the same function demoed standalone in
// cmd/concurrency-demo, now doing real work inside an HTTP handler. Because
// the checks run concurrently rather than one after another, the whole
// request takes roughly as long as the single slowest check, not the sum of
// all three. See docs/go/02-concurrency.md and docs/go/04-api.md.
func (a *itemsAPI) enrich(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item id")
		return
	}
	item, ok := a.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}

	checks := []concurrency.Check{
		{
			Name: "inventory",
			Run: func() (string, error) {
				time.Sleep(150 * time.Millisecond) // simulates a slow downstream call
				return fmt.Sprintf("42 units of %q in stock", item.Name), nil
			},
		},
		{
			Name: "pricing",
			Run: func() (string, error) {
				time.Sleep(150 * time.Millisecond)
				return "$19.99", nil
			},
		},
		{
			Name: "reviews",
			Run: func() (string, error) {
				time.Sleep(150 * time.Millisecond)
				return "4.5 stars (128 reviews)", nil
			},
		},
	}

	start := time.Now()
	results := concurrency.FanOut(checks)
	elapsed := time.Since(start)

	resp := enrichResponse{
		ItemID:    item.ID,
		Checks:    map[string]string{},
		ElapsedMs: elapsed.Milliseconds(),
	}
	for _, result := range results {
		if result.Err != nil {
			if resp.Errors == nil {
				resp.Errors = map[string]string{}
			}
			resp.Errors[result.Name] = result.Err.Error()
			continue
		}
		resp.Checks[result.Name] = result.Value
	}

	writeJSON(w, http.StatusOK, resp)
}
