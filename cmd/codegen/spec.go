// cmd/codegen/spec.go
package main

// Spec represents the telegram-bot-api-spec structure
type Spec struct {
	Version     string            `json:"version"`
	ReleaseDate string            `json:"release_date"`
	Changelog   string            `json:"changelog"`
	Methods     map[string]Method `json:"methods"`
	Types       map[string]Type   `json:"types"`
}

type Method struct {
	Name        string   `json:"name"`
	Href        string   `json:"href"`
	Description []string `json:"description"`
	Returns     []string `json:"returns"`
	Fields      []Field  `json:"fields"`
}

type Type struct {
	Name        string   `json:"name"`
	Href        string   `json:"href"`
	Description []string `json:"description"`
	Fields      []Field  `json:"fields"`
	Subtypes    []string `json:"subtypes"`
	SubtypeOf   []string `json:"subtype_of"`
}

type Field struct {
	Name        string   `json:"name"`
	Types       []string `json:"types"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
}
