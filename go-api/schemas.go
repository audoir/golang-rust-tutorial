package main

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is a single shared validator instance — it is defined once and
// reused for every request.
var validate = validator.New()

// CreateItemInput describes the shape of a valid "create" request body.
// Struct tags describe the validation rules declaratively — no separate
// schema object or class is needed.
type CreateItemInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// UpdateItemInput backs PUT (full update) — name is required.
type UpdateItemInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// PatchItemInput backs PATCH (partial update) — every field is optional, so
// pointers are used to distinguish "field omitted" from "field set to zero
// value".
type PatchItemInput struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// Sanitize trims whitespace from the name before validation runs.
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
// first failing field, so the client gets one clean message instead of a
// raw Go error value.
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
