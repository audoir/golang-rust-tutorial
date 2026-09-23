package items

import "sync"

// Store is a plain in-memory collection of Items, guarded by a mutex and
// reset on every restart — good enough for a tutorial, not for production.
//
// Every request the HTTP server handles runs on its own goroutine (see
// docs/go/02-concurrency.md), so multiple requests could read and write the
// underlying map at the same time without this lock.
type Store struct {
	mu     sync.Mutex
	items  map[int]*Item
	nextID int
}

// NewStore creates a Store seeded with one sample item, matching the fixture
// data the frontend expects to see on first load.
func NewStore() *Store {
	return &Store{
		items: map[int]*Item{
			1: {ID: 1, Name: "Sample Item", Description: "This is a sample item"},
		},
		nextID: 2,
	}
}

// List returns every item, ordered by ID.
func (s *Store) List() []*Item {
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

// Get returns the item with the given ID, or ok=false if it doesn't exist.
func (s *Store) Get(id int) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	return item, ok
}

// Create adds a new item and returns it with its assigned ID.
func (s *Store) Create(name, description string) *Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := &Item{ID: s.nextID, Name: name, Description: description}
	s.items[item.ID] = item
	s.nextID++
	return item
}

// Replace fully overwrites an existing item's fields (PUT semantics).
func (s *Store) Replace(id int, name, description string) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return nil, false
	}
	item := &Item{ID: id, Name: name, Description: description}
	s.items[id] = item
	return item, true
}

// Patch updates only the fields that are non-nil (PATCH semantics).
func (s *Store) Patch(id int, name, description *string) (*Item, bool) {
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

// Delete removes an item and returns the item that was removed.
func (s *Store) Delete(id int) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return nil, false
	}
	delete(s.items, id)
	return item, true
}
