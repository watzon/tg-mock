// internal/server/validator.go
package server

import (
	"fmt"

	"github.com/watzon/tg-mock/gen"
)

// Validator validates Bot API requests against method specifications
type Validator struct{}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks that all required fields are present in params
func (v *Validator) Validate(spec gen.MethodSpec, params map[string]interface{}) error {
	// Check required fields
	for _, field := range spec.Fields {
		if field.Required {
			if _, ok := params[field.Name]; !ok {
				return fmt.Errorf("missing required field: %s", field.Name)
			}
		}
	}

	// TODO: Add type validation

	return nil
}
