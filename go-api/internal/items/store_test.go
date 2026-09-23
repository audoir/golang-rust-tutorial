package items_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-api/internal/items"
)

func TestStore_List_ReturnsSeededItem(t *testing.T) {
	store := items.NewStore()

	got := store.List()

	require.Len(t, got, 1)
	assert.Equal(t, "Sample Item", got[0].Name)
}

func TestStore_Create(t *testing.T) {
	// Table-driven test: each case describes one create() call and the
	// item we expect back. This is the idiomatic Go way to cover several
	// similar scenarios without repeating the same test body.
	tests := []struct {
		name        string
		description string
	}{
		{name: "Widget", description: "A simple widget"},
		{name: "Gadget", description: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := items.NewStore()

			created := store.Create(tt.name, tt.description)

			assert.Equal(t, tt.name, created.Name)
			assert.Equal(t, tt.description, created.Description)

			got, ok := store.Get(created.ID)
			require.True(t, ok, "created item should be retrievable by ID")
			assert.Equal(t, created, got)
		})
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	store := items.NewStore()

	_, ok := store.Get(999)

	assert.False(t, ok)
}

func TestStore_Replace(t *testing.T) {
	store := items.NewStore()
	created := store.Create("Original", "Original description")

	t.Run("existing item", func(t *testing.T) {
		updated, ok := store.Replace(created.ID, "Replaced", "New description")

		require.True(t, ok)
		assert.Equal(t, "Replaced", updated.Name)
		assert.Equal(t, "New description", updated.Description)
	})

	t.Run("missing item", func(t *testing.T) {
		_, ok := store.Replace(999, "Nope", "Nope")

		assert.False(t, ok)
	})
}

func TestStore_Patch(t *testing.T) {
	store := items.NewStore()
	created := store.Create("Original", "Original description")

	newName := "Patched Name"

	t.Run("only name provided", func(t *testing.T) {
		updated, ok := store.Patch(created.ID, &newName, nil)

		require.True(t, ok)
		assert.Equal(t, "Patched Name", updated.Name)
		// Description was not provided (nil pointer), so it should be
		// untouched — this is the zero-value / pointer distinction
		// covered in docs/go/01-basics.md.
		assert.Equal(t, "Original description", updated.Description)
	})

	t.Run("missing item", func(t *testing.T) {
		_, ok := store.Patch(999, &newName, nil)

		assert.False(t, ok)
	})
}

func TestStore_Delete(t *testing.T) {
	store := items.NewStore()
	created := store.Create("Doomed", "Will be deleted")

	t.Run("existing item", func(t *testing.T) {
		deleted, ok := store.Delete(created.ID)

		require.True(t, ok)
		assert.Equal(t, created.ID, deleted.ID)

		_, stillThere := store.Get(created.ID)
		assert.False(t, stillThere, "item should no longer be retrievable after delete")
	})

	t.Run("missing item", func(t *testing.T) {
		_, ok := store.Delete(999)

		assert.False(t, ok)
	})
}
