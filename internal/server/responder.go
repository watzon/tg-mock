// internal/server/responder.go
package server

import (
	"github.com/watzon/tg-mock/gen"
	"github.com/watzon/tg-mock/internal/faker"
)

// Responder generates appropriate responses for Bot API methods
// using the faker engine for realistic mock data.
type Responder struct {
	faker *faker.Faker
}

// NewResponder creates a new Responder instance with the given faker.
func NewResponder(f *faker.Faker) *Responder {
	return &Responder{faker: f}
}

// Generate produces an appropriate response based on the method's return type.
// It uses the faker engine to generate realistic mock data.
func (r *Responder) Generate(spec gen.MethodSpec, params map[string]interface{}) (interface{}, error) {
	return r.GenerateWithOverrides(spec, params, nil)
}

// GenerateWithOverrides produces a response with user-specified overrides.
// The overrides map allows scenarios to specify custom values for specific fields.
func (r *Responder) GenerateWithOverrides(spec gen.MethodSpec, params, overrides map[string]interface{}) (interface{}, error) {
	if len(spec.Returns) == 0 {
		return true, nil
	}

	returnType := spec.Returns[0]
	return r.faker.GenerateWithOverrides(returnType, params, overrides), nil
}

// GetFaker returns the faker instance for direct access when needed.
func (r *Responder) GetFaker() *faker.Faker {
	return r.faker
}
