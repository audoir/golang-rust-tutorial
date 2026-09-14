// Package main implements a simple CRUD API using chi, the Go equivalent of
// Express in this tutorial: a minimal, unopinionated router with an explicit
// middleware pipeline that you wire up yourself.
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const port = "3005"

func main() {
	r := chi.NewRouter()

	// ── Middleware ────────────────────────────────────────────────────────────
	// chi.Middlewares are the same concept as Express middleware — functions
	// that run before the route handler for every matching request.
	r.Use(middleware.Logger)    // logs method, path, status, and latency
	r.Use(middleware.Recoverer) // recovers from panics with a 500 instead of crashing

	// Allow requests from the Next.js dev server (port 3000), the same job
	// `cors({ origin: "http://localhost:3000" })` does in Express.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	}))

	// ── Routes ────────────────────────────────────────────────────────────────
	api := newItemsAPI()
	r.Route("/api/items", api.routes)

	log.Printf("Go server running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
