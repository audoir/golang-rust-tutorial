// Package items holds the CRUD resource at the heart of this tutorial: the
// Item struct, its in-memory store, and the input structs used to validate
// incoming requests. Nothing in this package knows about HTTP — that keeps
// it easy to unit test in isolation (see store_test.go, schemas_test.go) and
// reusable from any transport the api package chooses to wire it up to.
package items

// Item is the shape of a resource returned by the API.
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
