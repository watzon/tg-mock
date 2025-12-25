// Package faker provides realistic mock data generation for Telegram Bot API types.
package faker

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Faker generates realistic mock data for Telegram Bot API responses.
// It supports configurable randomness via seed for reproducible tests.
type Faker struct {
	messageIDCounter int64
	updateIDCounter  int64
	userIDCounter    int64
	chatIDCounter    int64
	rng              *rand.Rand
	seed             int64
	mu               sync.Mutex

	// Type generators registry
	generators map[string]GeneratorFunc
}

// GeneratorFunc is a function that generates mock data for a specific type.
// params contains the request parameters that can be reflected in the response.
type GeneratorFunc func(f *Faker, params map[string]interface{}) map[string]interface{}

// Config holds configuration for the Faker.
type Config struct {
	// Seed for random number generation.
	// 0 = use current time (non-deterministic)
	// >0 = use fixed seed (deterministic, reproducible)
	Seed int64
}

// New creates a new Faker with the given configuration.
func New(cfg Config) *Faker {
	seed := cfg.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	f := &Faker{
		rng:        rand.New(rand.NewSource(seed)),
		seed:       seed,
		generators: make(map[string]GeneratorFunc),
	}

	// Register all type generators
	f.registerGenerators()

	return f
}

// Reset resets the faker to its initial state with a new seed.
// Useful for test isolation.
func (f *Faker) Reset(seed int64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	f.rng = rand.New(rand.NewSource(seed))
	f.seed = seed
	atomic.StoreInt64(&f.messageIDCounter, 0)
	atomic.StoreInt64(&f.updateIDCounter, 0)
	atomic.StoreInt64(&f.userIDCounter, 0)
	atomic.StoreInt64(&f.chatIDCounter, 0)
}

// Generate creates mock data for the given Telegram API type.
// params are the request parameters that may be reflected in the response.
func (f *Faker) Generate(typeName string, params map[string]interface{}) interface{} {
	return f.GenerateWithOverrides(typeName, params, nil)
}

// GenerateWithOverrides creates mock data with user-specified overrides.
// overrides is a map of field names to values that will replace generated values.
func (f *Faker) GenerateWithOverrides(typeName string, params map[string]interface{}, overrides map[string]interface{}) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Handle primitive types
	switch typeName {
	case "Boolean":
		if overrides != nil {
			if v, ok := overrides["value"]; ok {
				return v
			}
		}
		return true
	case "Integer":
		if overrides != nil {
			if v, ok := overrides["value"]; ok {
				return v
			}
		}
		return int64(f.rng.Intn(1000) + 1)
	case "String":
		if overrides != nil {
			if v, ok := overrides["value"]; ok {
				return v
			}
		}
		return f.generateString("value")
	}

	// Handle array types
	if len(typeName) > 9 && typeName[:9] == "Array of " {
		elementType := typeName[9:]
		return f.generateArray(elementType, params, overrides)
	}

	// Look up type generator
	generator, ok := f.generators[typeName]
	if !ok {
		// Fallback: generate based on type name heuristics
		return f.generateUnknownType(typeName, params, overrides)
	}

	// Generate base data
	result := generator(f, params)

	// Apply overrides
	if overrides != nil {
		result = f.mergeOverrides(result, overrides)
	}

	return result
}

// generateArray generates an array of the specified element type.
func (f *Faker) generateArray(elementType string, params map[string]interface{}, overrides map[string]interface{}) []interface{} {
	// Determine array size based on context
	size := f.arraySize(elementType)

	result := make([]interface{}, size)
	for i := 0; i < size; i++ {
		result[i] = f.GenerateWithOverrides(elementType, params, nil)
	}

	// Apply array overrides if provided
	if overrides != nil {
		if arr, ok := overrides["items"].([]interface{}); ok {
			return arr
		}
	}

	return result
}

// arraySize determines appropriate array size based on element type.
func (f *Faker) arraySize(elementType string) int {
	switch elementType {
	case "PhotoSize":
		return 3 // Standard: small, medium, large
	case "MessageEntity":
		return 0 // Usually empty unless there are entities
	case "Update":
		return 0 // Empty by default, populated via control API
	case "User":
		return 1 // Usually single user in arrays
	default:
		// Random small number for other types
		return f.rng.Intn(3)
	}
}

// generateUnknownType generates data for types without specific generators.
func (f *Faker) generateUnknownType(typeName string, params map[string]interface{}, overrides map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Apply overrides if provided
	if overrides != nil {
		for k, v := range overrides {
			result[k] = v
		}
	}

	return result
}

// mergeOverrides deep merges overrides into the result.
func (f *Faker) mergeOverrides(result, overrides map[string]interface{}) map[string]interface{} {
	for key, override := range overrides {
		if overrideMap, ok := override.(map[string]interface{}); ok {
			if existingMap, ok := result[key].(map[string]interface{}); ok {
				// Recursively merge nested maps
				result[key] = f.mergeOverrides(existingMap, overrideMap)
				continue
			}
		}
		// Direct override
		result[key] = override
	}
	return result
}

// NextMessageID returns the next auto-incrementing message ID.
func (f *Faker) NextMessageID() int64 {
	return atomic.AddInt64(&f.messageIDCounter, 1)
}

// NextUpdateID returns the next auto-incrementing update ID.
func (f *Faker) NextUpdateID() int64 {
	return atomic.AddInt64(&f.updateIDCounter, 1)
}

// NextUserID returns the next auto-incrementing user ID.
func (f *Faker) NextUserID() int64 {
	return atomic.AddInt64(&f.userIDCounter, 1)
}

// NextChatID returns the next auto-incrementing chat ID.
func (f *Faker) NextChatID() int64 {
	return atomic.AddInt64(&f.chatIDCounter, 1)
}

// RandomInt64 generates a random int64 within a range.
func (f *Faker) RandomInt64(min, max int64) int64 {
	if max <= min {
		return min
	}
	return min + f.rng.Int63n(max-min)
}

// RandomFloat64 generates a random float64 within a range.
func (f *Faker) RandomFloat64(min, max float64) float64 {
	return min + f.rng.Float64()*(max-min)
}

// RandomBool generates a random boolean with given probability of true.
func (f *Faker) RandomBool(trueProbability float64) bool {
	return f.rng.Float64() < trueProbability
}

// RandomChoice picks a random element from a slice.
func (f *Faker) RandomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	return choices[f.rng.Intn(len(choices))]
}
