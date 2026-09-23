package api

import (
	"encoding/json"
	"net/http"

	"go-api/internal/items"
)

// GET /api/items — list all items
func (a *itemsAPI) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List())
}

// POST /api/items — create a new item
func (a *itemsAPI) create(w http.ResponseWriter, r *http.Request) {
	var body items.CreateItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := items.Validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, items.FirstValidationError(err))
		return
	}
	item := a.store.Create(body.Name, body.Description)
	writeJSON(w, http.StatusCreated, item)
}

// GET /api/items/{id} — get a single item
func (a *itemsAPI) getOne(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, item)
}

// PUT /api/items/{id} — full update
func (a *itemsAPI) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item id")
		return
	}
	var body items.UpdateItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := items.Validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, items.FirstValidationError(err))
		return
	}
	item, ok := a.store.Replace(id, body.Name, body.Description)
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
	var body items.PatchItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body.Sanitize()
	if err := items.Validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, items.FirstValidationError(err))
		return
	}
	item, ok := a.store.Patch(id, body.Name, body.Description)
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
	item, ok := a.store.Delete(id)
	if !ok {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
