package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-api/internal/api"
)

// newTestServer spins up the real router (same one main() uses) wrapped in
// an httptest.Server, so tests exercise the full middleware + routing +
// handler stack exactly as a real client would hit it over HTTP.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(api.NewRouter())
	t.Cleanup(srv.Close)
	return srv
}

func decodeJSON[T any](t *testing.T, r *http.Response) T {
	t.Helper()
	var out T
	require.NoError(t, json.NewDecoder(r.Body).Decode(&out))
	return out
}

type itemDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func TestList_ReturnsSeededItem(t *testing.T) {
	srv := newTestServer(t)

	res, err := http.Get(srv.URL + "/api/items")
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	got := decodeJSON[[]itemDTO](t, res)
	require.Len(t, got, 1)
	assert.Equal(t, "Sample Item", got[0].Name)
}

func TestCreate_ValidBody(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"name": "New Item", "description": "Created in a test"})
	res, err := http.Post(srv.URL+"/api/items", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	got := decodeJSON[itemDTO](t, res)
	assert.Equal(t, "New Item", got.Name)
	assert.Equal(t, "Created in a test", got.Description)
}

func TestCreate_InvalidBody_MissingName(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"description": "No name provided"})
	res, err := http.Post(srv.URL+"/api/items", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	got := decodeJSON[map[string]string](t, res)
	assert.Contains(t, got["error"], "name")
}

func TestGetOne_NotFound(t *testing.T) {
	srv := newTestServer(t)

	res, err := http.Get(srv.URL + "/api/items/999")
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestUpdate_ThenGet_ReflectsChange(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"name": "Updated Name", "description": "Updated description"})
	req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/items/1", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	getRes, err := http.Get(srv.URL + "/api/items/1")
	require.NoError(t, err)
	defer getRes.Body.Close()

	got := decodeJSON[itemDTO](t, getRes)
	assert.Equal(t, "Updated Name", got.Name)
}

func TestPartialUpdate_OnlyChangesProvidedFields(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"description": "Only description changed"})
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/items/1", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	got := decodeJSON[itemDTO](t, res)
	assert.Equal(t, "Sample Item", got.Name, "name should be untouched by a description-only patch")
	assert.Equal(t, "Only description changed", got.Description)
}

func TestDelete_RemovesItem(t *testing.T) {
	srv := newTestServer(t)

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/items/1", nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	getRes, err := http.Get(srv.URL + "/api/items/1")
	require.NoError(t, err)
	defer getRes.Body.Close()
	assert.Equal(t, http.StatusNotFound, getRes.StatusCode)
}
