package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type enrichDTO struct {
	ItemID    int               `json:"itemId"`
	Checks    map[string]string `json:"checks"`
	Errors    map[string]string `json:"errors"`
	ElapsedMs int64             `json:"elapsedMs"`
}

func TestEnrich_RunsChecksConcurrently(t *testing.T) {
	srv := newTestServer(t)

	start := time.Now()
	res, err := http.Get(srv.URL + "/api/items/1/enrich")
	elapsed := time.Since(start)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	got := decodeJSON[enrichDTO](t, res)

	assert.Equal(t, 1, got.ItemID)
	assert.Len(t, got.Checks, 3)
	assert.Empty(t, got.Errors)

	// Each simulated check sleeps 150ms; sequential execution would take
	// ~450ms. Assert the real wall-clock request time stays well under
	// that, proving the three checks ran concurrently.
	assert.Less(t, elapsed, 400*time.Millisecond, "expected the three checks to run concurrently")
}

func TestEnrich_NotFound(t *testing.T) {
	srv := newTestServer(t)

	res, err := http.Get(srv.URL + "/api/items/999/enrich")
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
