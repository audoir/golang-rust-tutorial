package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
)

// Item is the shape of a resource returned by the API.
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ── In-memory store ───────────────────────────────────────────────────────────
// A plain in-memory collection guarded by a mutex, reset on restart. Good
// enough for a tutorial, not for production.
type itemStore struct {
	mu     sync.Mutex
	items  map[int]*Item
	nextID int
}

func newItemStore() *itemStore {
	return &itemStore{
		items: map[int]*Item{
			1: {ID: 1, Name: "Sample Item", Description: "This is a sample item"},
		},
		nextID: 2,
	}
}

func (s *itemStore) list() []*Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Item, 0, len(s.items))
	for id := 1; id < s.nextID; id++ {
		if item, ok := s.items[id]; ok {
			out = append(out, item)
		}
	}
	return out
}

func (s *itemStore) get(id int) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	return item, ok
}

func (s *itemStore) create(name, description string) *Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := &Item{ID: s.nextID, Name: name, Description: description}
	s.items[item.ID] = item
	s.nextID++
	return item
}

func (s *itemStore) replace(id int, name, description string) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return nil, false
	}
	item := &Item{ID: id, Name: name, Description: description}
	s.items[id] = item
	return item, true
}

func (s *itemStore) patch(id int, name, description *string) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return nil, false
	}
	if name != nil {
		item.Name = *name
	}
	if description != nil {
		item.Description = *description
	}
	return item, true
}

func (s *itemStore) delete(id int) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return nil, false
	}
	delete(s.items, id)
	return item, true
}

// ── Handlers ──────────────────────────────────────────────────────────────────
// itemsAPI groups the store together with its HTTP handlers.
type itemsAPI struct {
	store *itemStore
}

func newItemsAPI() *itemsAPI {
	return &itemsAPI{store: newItemStore()}
}

// routes wires this API's handlers onto a chi router at the given base path.
func (a *itemsAPI) routes(r chi.Router) {
	r.Get("/", a.list)
	r.Post("/", a.create)
	r.Get("/{id}", a.getOne)
	r.Put("/{id}", a.update)
	r.Patch("/{id}", a.partialUpdate)
	r.Delete("/{id}", a.remove)
}

// GET /api/items — list all items
func (a *itemsAPI) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.list())
}

// POST /api/items — create a new item
func (a *itemsAPI) create(w http.ResponseWriter, r *http.Request) {
	var body CreateItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, firstValidationError(err))
		return
	}
	item := a.store.create(body.Name, body.Description)
	writeJSON(w, http.StatusCreated, item)
}

// GET /api/items/{id} — get a single item
func (a *itemsAPI) getOne(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item id")
		return
	}
	item, ok := a.store.get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// PUT /api/items/{id} — full update
func (a *itemsAPI) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item id")
		return
	}
	var body UpdateItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, firstValidationError(err))
		return
	}
	item, ok := a.store.replace(id, body.Name, body.Description)
	if !ok {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// PATCH /api/items/{id} — partial update
func (a *itemsAPI) partialUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item id")
		return
	}
	var body PatchItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, firstValidationError(err))
		return
	}
	item, ok := a.store.patch(id, body.Name, body.Description)
	if !ok {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// DELETE /api/items/{id}
func (a *itemsAPI) remove(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item id")
		return
	}
	item, ok := a.store.delete(id)
	if !ok {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
