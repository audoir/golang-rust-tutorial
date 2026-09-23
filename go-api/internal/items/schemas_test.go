package items_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-api/internal/items"
)

func TestCreateItemInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   items.CreateItemInput
		wantErr bool
	}{
		{
			name:    "valid input",
			input:   items.CreateItemInput{Name: "Valid Name", Description: "Optional description"},
			wantErr: false,
		},
		{
			name:    "valid input without description",
			input:   items.CreateItemInput{Name: "Valid Name"},
			wantErr: false,
		},
		{
			name:    "name too short",
			input:   items.CreateItemInput{Name: "x"},
			wantErr: true,
		},
		{
			name:    "name missing",
			input:   items.CreateItemInput{Description: "No name provided"},
			wantErr: true,
		},
		{
			name:    "description too long",
			input:   items.CreateItemInput{Name: "Valid Name", Description: string(make([]byte, 501))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := items.Validate.Struct(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPatchItemInput_OmittedFieldsAreNil(t *testing.T) {
	// Decoding an empty JSON object should leave both pointer fields nil,
	// which is exactly what lets the handler distinguish "field omitted"
	// from "field explicitly set to empty string" (see docs/go/01-basics.md
	// on zero values).
	input := items.PatchItemInput{}

	assert.Nil(t, input.Name)
	assert.Nil(t, input.Description)
	require.NoError(t, items.Validate.Struct(input), "an entirely empty patch should be valid")
}

func TestSanitize_TrimsWhitespace(t *testing.T) {
	t.Run("CreateItemInput", func(t *testing.T) {
		input := items.CreateItemInput{Name: "  Padded Name  "}
		input.Sanitize()
		assert.Equal(t, "Padded Name", input.Name)
	})

	t.Run("PatchItemInput", func(t *testing.T) {
		padded := "  Padded Name  "
		input := items.PatchItemInput{Name: &padded}
		input.Sanitize()
		require.NotNil(t, input.Name)
		assert.Equal(t, "Padded Name", *input.Name)
	})
}

func TestFirstValidationError_Messages(t *testing.T) {
	tests := []struct {
		name    string
		input   items.CreateItemInput
		wantSub string
	}{
		{name: "required", input: items.CreateItemInput{}, wantSub: "required"},
		{name: "min", input: items.CreateItemInput{Name: "x"}, wantSub: "at least"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := items.Validate.Struct(tt.input)
			require.Error(t, err)

			msg := items.FirstValidationError(err)
			assert.Contains(t, msg, tt.wantSub)
		})
	}
}
