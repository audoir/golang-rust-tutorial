package main

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is a single shared validator instance, analogous to a Zod schema
// object or a Pydantic BaseModel — it is defined once and reused for every
// request.
var validate = validator.New()

// CreateItemInput mirrors the Zod `createItemSchema` / Pydantic
// `CreateItemSchema` used in the other tabs. Struct tags describe the
// validation rules declaratively, the same way `z.string().min(2).max(100)`
// or a Pydantic `@field_validator` does.
type CreateItemInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// UpdateItemInput backs PUT (full update) — name is required, just like the
// other frameworks' "update" schema.
type UpdateItemInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// PatchItemInput backs PATCH (partial update) — every field is optional, so
// pointers are used to distinguish "field omitted" from "field set to zero
// value" (the Go equivalent of Zod's `.optional()` / Pydantic's `Optional[str] = None`).
type PatchItemInput struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// Sanitize trims whitespace from the name before validation runs, mirroring
// the `.strip()` calls in the Pydantic field validators.
func (c *CreateItemInput) Sanitize() {
	c.Name = strings.TrimSpace(c.Name)
}

func (u *UpdateItemInput) Sanitize() {
	u.Name = strings.TrimSpace(u.Name)
}

func (p *PatchItemInput) Sanitize() {
	if p.Name != nil {
		trimmed := strings.TrimSpace(*p.Name)
		p.Name = &trimmed
	}
}

// firstValidationError extracts a single human-readable message from the
// first failing field, the same way the TS tabs read `err.issues[0].message`
// and the Python tabs read `exc.errors()[0]["msg"]`.
func firstValidationError(err error) string {
	verrs, ok := err.(validator.ValidationErrors)
	if !ok || len(verrs) == 0 {
		return "Validation error"
	}
	fe := verrs[0]
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return field + " is required and cannot be empty"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	default:
		return field + " is invalid"
	}
}
