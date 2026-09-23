// Package api wires the go-api/internal/items store and
// go-api/internal/concurrency fan-out helper up to real HTTP handlers using
// chi: a minimal, unopinionated router with an explicit middleware pipeline
// that you wire up yourself.
package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"go-api/internal/items"
)

// itemsAPI groups the item store together with its HTTP handlers.
type itemsAPI struct {
	store *items.Store
}

// NewRouter builds the fully configured chi router for this server:
// middleware, CORS, and every route under /api/items.
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// ── Middleware ────────────────────────────────────────────────────────
	// Each middleware is a function that wraps the next handler in the
	// chain and runs before the route handler for every matching request.
	r.Use(middleware.Logger)    // logs method, path, status, and latency
	r.Use(middleware.Recoverer) // recovers from panics with a 500 instead of crashing

	// Allow requests from the Next.js dev server (port 3000).
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	}))

	// ── Routes ────────────────────────────────────────────────────────────
	api := &itemsAPI{store: items.NewStore()}
	r.Route("/api/items", api.routes)

	return r
}

// routes wires this API's handlers onto a chi router at the given base path.
func (a *itemsAPI) routes(r chi.Router) {
	r.Get("/", a.list)
	r.Post("/", a.create)
	r.Get("/{id}", a.getOne)
	r.Put("/{id}", a.update)
	r.Patch("/{id}", a.partialUpdate)
	r.Delete("/{id}", a.remove)
	r.Get("/{id}/enrich", a.enrich)
}
