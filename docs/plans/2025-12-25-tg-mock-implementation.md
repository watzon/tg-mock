# tg-mock Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a mock Telegram Bot API server for testing bots and bot libraries.

**Architecture:** HTTP server with chi router, codegen from telegram-bot-api-spec, scenario engine for error simulation, control API for test management.

**Tech Stack:** Go 1.21+, chi router, gopkg.in/yaml.v3

---

## Phase 1: Project Foundation

### Task 1: Project Setup

**Files:**
- Create: `cmd/tg-mock/main.go`
- Create: `cmd/codegen/main.go`
- Create: `Makefile`
- Modify: `go.mod`

**Step 1: Create directory structure**

```bash
mkdir -p cmd/tg-mock cmd/codegen internal/server internal/scenario internal/updates internal/storage internal/tokens internal/config gen spec errors
```

**Step 2: Update go.mod with dependencies**

```go
module github.com/watzon/tg-mock

go 1.21

require (
	github.com/go-chi/chi/v5 v5.1.0
	gopkg.in/yaml.v3 v3.0.1
)
```

Run: `go mod tidy`

**Step 3: Create minimal main.go**

```go
// cmd/tg-mock/main.go
package main

import "fmt"

func main() {
	fmt.Println("tg-mock starting...")
}
```

**Step 4: Create Makefile**

```makefile
.PHONY: build run generate test clean

build:
	go build -o bin/tg-mock ./cmd/tg-mock
	go build -o bin/codegen ./cmd/codegen

run: build
	./bin/tg-mock

generate:
	go run ./cmd/codegen -spec spec/api.json -out gen

test:
	go test -v ./...

clean:
	rm -rf bin/

fetch-spec:
	curl -sL https://raw.githubusercontent.com/PaulSonOfLars/telegram-bot-api-spec/main/api.json > spec/api.json

fetch-errors:
	curl -sL https://raw.githubusercontent.com/TelegramBotAPI/errors/master/errors.json > errors/errors.json
```

**Step 5: Verify build**

Run: `go build ./...`
Expected: No errors

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: project setup with directory structure and Makefile"
```

---

### Task 2: Fetch API Spec

**Files:**
- Create: `spec/api.json`

**Step 1: Download spec**

Run: `make fetch-spec`
Expected: `spec/api.json` created with Telegram Bot API spec

**Step 2: Verify spec structure**

```bash
head -50 spec/api.json
```

Expected: JSON with `version`, `methods`, `types` keys

**Step 3: Commit**

```bash
git add spec/api.json
git commit -m "chore: add telegram bot api spec"
```

---

## Phase 2: Code Generation

### Task 3: Codegen - Spec Parser

**Files:**
- Create: `cmd/codegen/main.go`
- Create: `cmd/codegen/spec.go`

**Step 1: Write spec types**

```go
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
```

**Step 2: Write spec loader**

```go
// cmd/codegen/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	specPath := flag.String("spec", "spec/api.json", "path to API spec")
	outDir := flag.String("out", "gen", "output directory")
	flag.Parse()

	spec, err := loadSpec(*specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded spec: %s (%s)\n", spec.Version, spec.ReleaseDate)
	fmt.Printf("Methods: %d, Types: %d\n", len(spec.Methods), len(spec.Types))
	fmt.Printf("Output directory: %s\n", *outDir)
}

func loadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}
```

**Step 3: Test codegen runs**

Run: `go run ./cmd/codegen -spec spec/api.json`
Expected: Output showing version and counts

**Step 4: Commit**

```bash
git add cmd/codegen/
git commit -m "feat(codegen): add spec parser"
```

---

### Task 4: Codegen - Type Generation

**Files:**
- Create: `cmd/codegen/types.go`
- Create: `gen/types.go` (generated)

**Step 1: Write type generator**

```go
// cmd/codegen/types.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func generateTypes(spec *Spec, outDir string) error {
	f, err := os.Create(filepath.Join(outDir, "types.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "// Code generated by codegen. DO NOT EDIT.")
	fmt.Fprintln(f, "package gen")
	fmt.Fprintln(f)

	// Sort types for deterministic output
	names := make([]string, 0, len(spec.Types))
	for name := range spec.Types {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		t := spec.Types[name]
		if err := writeType(f, t); err != nil {
			return err
		}
	}

	return nil
}

func writeType(f *os.File, t Type) error {
	// Write doc comment
	if len(t.Description) > 0 {
		fmt.Fprintf(f, "// %s %s\n", t.Name, t.Description[0])
	}

	fmt.Fprintf(f, "type %s struct {\n", t.Name)

	for _, field := range t.Fields {
		goType := toGoType(field.Types, !field.Required)
		jsonTag := field.Name
		if !field.Required {
			jsonTag += ",omitempty"
		}
		fmt.Fprintf(f, "\t%s %s `json:\"%s\"`\n", toCamelCase(field.Name), goType, jsonTag)
	}

	fmt.Fprintln(f, "}")
	fmt.Fprintln(f)

	return nil
}

func toGoType(types []string, optional bool) string {
	if len(types) == 0 {
		return "interface{}"
	}

	t := types[0]

	// Handle arrays
	if strings.HasPrefix(t, "Array of ") {
		inner := strings.TrimPrefix(t, "Array of ")
		return "[]" + toGoType([]string{inner}, false)
	}

	// Map Telegram types to Go types
	switch t {
	case "Integer":
		if optional {
			return "*int64"
		}
		return "int64"
	case "Float", "Float number":
		if optional {
			return "*float64"
		}
		return "float64"
	case "String":
		return "string"
	case "Boolean", "True":
		if optional {
			return "*bool"
		}
		return "bool"
	case "InputFile":
		return "interface{}" // Can be file upload or string
	default:
		// Union types or complex types
		if len(types) > 1 {
			return "interface{}"
		}
		if optional {
			return "*" + t
		}
		return t
	}
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}
```

**Step 2: Update main.go to call generator**

Add to `cmd/codegen/main.go`:

```go
// Add after the Printf statements in main()
if err := os.MkdirAll(*outDir, 0755); err != nil {
	fmt.Fprintf(os.Stderr, "failed to create output dir: %v\n", err)
	os.Exit(1)
}

if err := generateTypes(spec, *outDir); err != nil {
	fmt.Fprintf(os.Stderr, "failed to generate types: %v\n", err)
	os.Exit(1)
}
fmt.Println("Generated types.go")
```

**Step 3: Run generator**

Run: `make generate`
Expected: `gen/types.go` created

**Step 4: Verify generated code compiles**

Run: `go build ./gen`
Expected: No errors

**Step 5: Commit**

```bash
git add cmd/codegen/ gen/
git commit -m "feat(codegen): generate Go types from spec"
```

---

### Task 5: Codegen - Method Registry

**Files:**
- Create: `cmd/codegen/methods.go`
- Create: `gen/methods.go` (generated)

**Step 1: Write method generator**

```go
// cmd/codegen/methods.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func generateMethods(spec *Spec, outDir string) error {
	f, err := os.Create(filepath.Join(outDir, "methods.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "// Code generated by codegen. DO NOT EDIT.")
	fmt.Fprintln(f, "package gen")
	fmt.Fprintln(f)

	// Generate FieldSpec type
	fmt.Fprintln(f, "// FieldSpec describes a method parameter")
	fmt.Fprintln(f, "type FieldSpec struct {")
	fmt.Fprintln(f, "\tName     string")
	fmt.Fprintln(f, "\tTypes    []string")
	fmt.Fprintln(f, "\tRequired bool")
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f)

	// Generate MethodSpec type
	fmt.Fprintln(f, "// MethodSpec describes a Bot API method")
	fmt.Fprintln(f, "type MethodSpec struct {")
	fmt.Fprintln(f, "\tName    string")
	fmt.Fprintln(f, "\tReturns []string")
	fmt.Fprintln(f, "\tFields  []FieldSpec")
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f)

	// Generate method registry
	fmt.Fprintln(f, "// Methods is the registry of all Bot API methods")
	fmt.Fprintln(f, "var Methods = map[string]MethodSpec{")

	names := make([]string, 0, len(spec.Methods))
	for name := range spec.Methods {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		m := spec.Methods[name]
		fmt.Fprintf(f, "\t%q: {\n", name)
		fmt.Fprintf(f, "\t\tName:    %q,\n", m.Name)
		fmt.Fprintf(f, "\t\tReturns: %#v,\n", m.Returns)

		if len(m.Fields) > 0 {
			fmt.Fprintln(f, "\t\tFields: []FieldSpec{")
			for _, field := range m.Fields {
				fmt.Fprintf(f, "\t\t\t{Name: %q, Types: %#v, Required: %v},\n",
					field.Name, field.Types, field.Required)
			}
			fmt.Fprintln(f, "\t\t},")
		}

		fmt.Fprintln(f, "\t},")
	}

	fmt.Fprintln(f, "}")

	return nil
}
```

**Step 2: Update main.go**

Add to `cmd/codegen/main.go` after generateTypes:

```go
if err := generateMethods(spec, *outDir); err != nil {
	fmt.Fprintf(os.Stderr, "failed to generate methods: %v\n", err)
	os.Exit(1)
}
fmt.Println("Generated methods.go")
```

**Step 3: Run and verify**

Run: `make generate && go build ./gen`
Expected: Both files generated, code compiles

**Step 4: Commit**

```bash
git add cmd/codegen/ gen/
git commit -m "feat(codegen): generate method registry"
```

---

### Task 6: Codegen - Fixtures

**Files:**
- Create: `cmd/codegen/fixtures.go`
- Create: `gen/fixtures.go` (generated)

**Step 1: Write fixture generator**

```go
// cmd/codegen/fixtures.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func generateFixtures(spec *Spec, outDir string) error {
	f, err := os.Create(filepath.Join(outDir, "fixtures.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "// Code generated by codegen. DO NOT EDIT.")
	fmt.Fprintln(f, "package gen")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "import \"time\"")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "var _ = time.Now // Ensure time is used")
	fmt.Fprintln(f)

	names := make([]string, 0, len(spec.Types))
	for name := range spec.Types {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		t := spec.Types[name]
		if err := writeFixture(f, t, spec); err != nil {
			return err
		}
	}

	return nil
}

func writeFixture(f *os.File, t Type, spec *Spec) error {
	fmt.Fprintf(f, "// New%s creates a fixture %s\n", t.Name, t.Name)
	fmt.Fprintf(f, "func New%s() *%s {\n", t.Name, t.Name)
	fmt.Fprintf(f, "\treturn &%s{\n", t.Name)

	for _, field := range t.Fields {
		if !field.Required {
			continue // Skip optional fields in fixtures
		}
		val := fixtureValue(field, spec)
		fmt.Fprintf(f, "\t\t%s: %s,\n", toCamelCase(field.Name), val)
	}

	fmt.Fprintln(f, "\t}")
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f)

	return nil
}

func fixtureValue(field Field, spec *Spec) string {
	if len(field.Types) == 0 {
		return "nil"
	}

	t := field.Types[0]

	if strings.HasPrefix(t, "Array of ") {
		return "nil"
	}

	switch t {
	case "Integer":
		// Special cases
		switch field.Name {
		case "date":
			return "int64(time.Now().Unix())"
		case "message_id":
			return "1"
		case "update_id":
			return "1"
		default:
			return "1"
		}
	case "Float", "Float number":
		return "1.0"
	case "String":
		switch field.Name {
		case "type":
			return `"private"`
		case "text":
			return `"Hello"`
		case "first_name":
			return `"Test"`
		case "username":
			return `"testuser"`
		case "title":
			return `"Test Chat"`
		default:
			return fmt.Sprintf("%q", field.Name)
		}
	case "Boolean", "True":
		return "true"
	default:
		// Check if it's a known type
		if _, ok := spec.Types[t]; ok {
			return fmt.Sprintf("*New%s()", t)
		}
		return "nil"
	}
}
```

**Step 2: Update main.go**

Add after generateMethods:

```go
if err := generateFixtures(spec, *outDir); err != nil {
	fmt.Fprintf(os.Stderr, "failed to generate fixtures: %v\n", err)
	os.Exit(1)
}
fmt.Println("Generated fixtures.go")
```

**Step 3: Run and verify**

Run: `make generate && go build ./gen`
Expected: Compiles (may have warnings about unused time)

**Step 4: Commit**

```bash
git add cmd/codegen/ gen/
git commit -m "feat(codegen): generate fixture constructors"
```

---

## Phase 3: Core Server

### Task 7: HTTP Server Setup

**Files:**
- Create: `internal/server/server.go`
- Modify: `cmd/tg-mock/main.go`

**Step 1: Write server struct**

```go
// internal/server/server.go
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router     chi.Router
	httpServer *http.Server
	port       int
}

type Config struct {
	Port    int
	Verbose bool
}

func New(cfg Config) *Server {
	r := chi.NewRouter()

	if cfg.Verbose {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	s := &Server{
		router: r,
		port:   cfg.Port,
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Control API placeholder
	s.router.Route("/__control", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		})
	})

	// Bot API placeholder
	s.router.Route("/bot{token}", func(r chi.Router) {
		r.Post("/{method}", s.handleBotMethod)
		r.Get("/{method}", s.handleBotMethod)
	})
}

func (s *Server) handleBotMethod(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	method := chi.URLParam(r, "method")

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true,"method":%q,"token":%q}`, method, token[:10]+"...")
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("tg-mock listening on :%d\n", s.port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
```

**Step 2: Update main.go**

```go
// cmd/tg-mock/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/watzon/tg-mock/internal/server"
)

func main() {
	port := flag.Int("port", 8081, "HTTP server port")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	srv := server.New(server.Config{
		Port:    *port,
		Verbose: *verbose,
	})

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down...")
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Test server starts**

Run: `go run ./cmd/tg-mock --verbose`
Expected: "tg-mock listening on :8081"

In another terminal:
```bash
curl http://localhost:8081/health
curl http://localhost:8081/bot123456:ABC/sendMessage
```

**Step 4: Commit**

```bash
git add internal/server/ cmd/tg-mock/
git commit -m "feat: add HTTP server with chi router"
```

---

### Task 8: Token Validation

**Files:**
- Create: `internal/tokens/registry.go`
- Create: `internal/tokens/registry_test.go`

**Step 1: Write failing test**

```go
// internal/tokens/registry_test.go
package tokens

import "testing"

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		token string
		valid bool
	}{
		{"123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11", true},
		{"123456:abcdef", true},
		{"invalid", false},
		{"", false},
		{"123456:", false},
		{":ABC", false},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			if got := ValidateFormat(tt.token); got != tt.valid {
				t.Errorf("ValidateFormat(%q) = %v, want %v", tt.token, got, tt.valid)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Register a token
	r.Register("123:abc", TokenInfo{Status: StatusActive, BotName: "TestBot"})

	// Check status
	info, ok := r.Get("123:abc")
	if !ok {
		t.Fatal("expected token to be registered")
	}
	if info.Status != StatusActive {
		t.Errorf("got status %v, want %v", info.Status, StatusActive)
	}

	// Unknown token
	_, ok = r.Get("unknown:token")
	if ok {
		t.Error("expected unknown token to not be found")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tokens/...`
Expected: FAIL (package doesn't exist)

**Step 3: Write implementation**

```go
// internal/tokens/registry.go
package tokens

import (
	"regexp"
	"sync"
)

type Status string

const (
	StatusActive      Status = "active"
	StatusBanned      Status = "banned"
	StatusDeactivated Status = "deactivated"
)

type TokenInfo struct {
	Status  Status
	BotName string
}

type Registry struct {
	mu     sync.RWMutex
	tokens map[string]TokenInfo
}

var tokenPattern = regexp.MustCompile(`^\d+:[A-Za-z0-9_-]+$`)

func ValidateFormat(token string) bool {
	return tokenPattern.MatchString(token)
}

func NewRegistry() *Registry {
	return &Registry{
		tokens: make(map[string]TokenInfo),
	}
}

func (r *Registry) Register(token string, info TokenInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token] = info
}

func (r *Registry) Get(token string) (TokenInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.tokens[token]
	return info, ok
}

func (r *Registry) Delete(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, token)
}

func (r *Registry) UpdateStatus(token string, status Status) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if info, ok := r.tokens[token]; ok {
		info.Status = status
		r.tokens[token] = info
		return true
	}
	return false
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tokens/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tokens/
git commit -m "feat: add token validation and registry"
```

---

### Task 9: Integrate Token Validation

**Files:**
- Modify: `internal/server/server.go`
- Create: `internal/server/bot_handler.go`

**Step 1: Create bot handler**

```go
// internal/server/bot_handler.go
package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/watzon/tg-mock/gen"
	"github.com/watzon/tg-mock/internal/tokens"
)

type BotHandler struct {
	registry       *tokens.Registry
	registryEnabled bool
}

func NewBotHandler(registry *tokens.Registry, registryEnabled bool) *BotHandler {
	return &BotHandler{
		registry:       registry,
		registryEnabled: registryEnabled,
	}
}

type APIResponse struct {
	OK          bool        `json:"ok"`
	Result      interface{} `json:"result,omitempty"`
	ErrorCode   int         `json:"error_code,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (h *BotHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	method := chi.URLParam(r, "method")

	w.Header().Set("Content-Type", "application/json")

	// Validate token format
	if !tokens.ValidateFormat(token) {
		h.writeError(w, 401, "Unauthorized: invalid token format")
		return
	}

	// Check registry if enabled
	if h.registryEnabled {
		info, ok := h.registry.Get(token)
		if !ok {
			h.writeError(w, 401, "Unauthorized: token not registered")
			return
		}
		switch info.Status {
		case tokens.StatusBanned:
			h.writeError(w, 403, "Forbidden: bot was banned")
			return
		case tokens.StatusDeactivated:
			h.writeError(w, 401, "Unauthorized: bot was deactivated")
			return
		}
	}

	// Check method exists
	spec, ok := gen.Methods[method]
	if !ok {
		h.writeError(w, 404, "Not Found: method not found")
		return
	}

	// For now, return a success stub
	h.writeSuccess(w, map[string]interface{}{
		"method": spec.Name,
	})
}

func (h *BotHandler) writeError(w http.ResponseWriter, code int, desc string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		OK:          false,
		ErrorCode:   code,
		Description: desc,
	})
}

func (h *BotHandler) writeSuccess(w http.ResponseWriter, result interface{}) {
	json.NewEncoder(w).Encode(APIResponse{
		OK:     true,
		Result: result,
	})
}
```

**Step 2: Update server.go to use BotHandler**

```go
// Update Server struct to include:
type Server struct {
	router       chi.Router
	httpServer   *http.Server
	port         int
	tokenRegistry *tokens.Registry
	botHandler   *BotHandler
}

// Update New() function:
func New(cfg Config) *Server {
	r := chi.NewRouter()

	if cfg.Verbose {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	registry := tokens.NewRegistry()

	s := &Server{
		router:        r,
		port:          cfg.Port,
		tokenRegistry: registry,
		botHandler:    NewBotHandler(registry, false), // Registry disabled by default
	}

	s.setupRoutes()

	return s
}

// Update setupRoutes():
func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	s.router.Route("/__control", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		})
	})

	s.router.Route("/bot{token}", func(r chi.Router) {
		r.Post("/{method}", s.botHandler.Handle)
		r.Get("/{method}", s.botHandler.Handle)
	})
}
```

**Step 3: Test token validation**

Run: `go run ./cmd/tg-mock`

```bash
# Valid token format
curl http://localhost:8081/bot123456:ABC/getMe
# Expected: {"ok":true,"result":{"method":"getMe"}}

# Invalid token format
curl http://localhost:8081/botinvalid/getMe
# Expected: {"ok":false,"error_code":401,...}

# Unknown method
curl http://localhost:8081/bot123:abc/unknownMethod
# Expected: {"ok":false,"error_code":404,...}
```

**Step 4: Commit**

```bash
git add internal/server/
git commit -m "feat: integrate token validation into bot handler"
```

---

## Phase 4: Request Validation & Response Generation

### Task 10: Request Validator

**Files:**
- Create: `internal/server/validator.go`
- Create: `internal/server/validator_test.go`

**Step 1: Write failing test**

```go
// internal/server/validator_test.go
package server

import (
	"testing"

	"github.com/watzon/tg-mock/gen"
)

func TestValidateRequest(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		method  string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name:   "sendMessage valid",
			method: "sendMessage",
			params: map[string]interface{}{
				"chat_id": 123,
				"text":    "Hello",
			},
			wantErr: false,
		},
		{
			name:   "sendMessage missing required",
			method: "sendMessage",
			params: map[string]interface{}{
				"chat_id": 123,
			},
			wantErr: true,
		},
		{
			name:    "getMe no params required",
			method:  "getMe",
			params:  map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := gen.Methods[tt.method]
			err := v.Validate(spec, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/server/... -run TestValidateRequest`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/server/validator.go
package server

import (
	"fmt"

	"github.com/watzon/tg-mock/gen"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/server/... -run TestValidateRequest`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/server/validator*.go
git commit -m "feat: add request validator for required fields"
```

---

### Task 11: Response Generator

**Files:**
- Create: `internal/server/responder.go`
- Create: `internal/server/responder_test.go`

**Step 1: Write failing test**

```go
// internal/server/responder_test.go
package server

import (
	"testing"

	"github.com/watzon/tg-mock/gen"
)

func TestGenerateResponse(t *testing.T) {
	r := NewResponder()

	t.Run("getMe returns User", func(t *testing.T) {
		spec := gen.Methods["getMe"]
		params := map[string]interface{}{}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Result should be a User
		user, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		if _, ok := user["id"]; !ok {
			t.Error("User should have id field")
		}
		if _, ok := user["is_bot"]; !ok {
			t.Error("User should have is_bot field")
		}
	})

	t.Run("sendMessage reflects chat_id", func(t *testing.T) {
		spec := gen.Methods["sendMessage"]
		params := map[string]interface{}{
			"chat_id": int64(12345),
			"text":    "Hello world",
		}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		msg, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		chat, ok := msg["chat"].(map[string]interface{})
		if !ok {
			t.Fatal("expected chat field")
		}

		if chat["id"] != int64(12345) {
			t.Errorf("chat.id = %v, want 12345", chat["id"])
		}

		if msg["text"] != "Hello world" {
			t.Errorf("text = %v, want 'Hello world'", msg["text"])
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/server/... -run TestGenerateResponse`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/server/responder.go
package server

import (
	"sync/atomic"
	"time"

	"github.com/watzon/tg-mock/gen"
)

type Responder struct {
	messageIDCounter int64
}

func NewResponder() *Responder {
	return &Responder{}
}

func (r *Responder) nextMessageID() int64 {
	return atomic.AddInt64(&r.messageIDCounter, 1)
}

func (r *Responder) Generate(spec gen.MethodSpec, params map[string]interface{}) (interface{}, error) {
	if len(spec.Returns) == 0 {
		return true, nil
	}

	returnType := spec.Returns[0]

	switch returnType {
	case "Boolean":
		return true, nil
	case "User":
		return r.generateUser(params), nil
	case "Message":
		return r.generateMessage(params), nil
	case "Array of Update":
		return []interface{}{}, nil
	default:
		// Return a basic success response
		return map[string]interface{}{}, nil
	}
}

func (r *Responder) generateUser(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":         int64(123456789),
		"is_bot":     true,
		"first_name": "TestBot",
		"username":   "test_bot",
	}
}

func (r *Responder) generateMessage(params map[string]interface{}) map[string]interface{} {
	chatID := int64(1)
	if id, ok := params["chat_id"].(int64); ok {
		chatID = id
	} else if id, ok := params["chat_id"].(float64); ok {
		chatID = int64(id)
	}

	msg := map[string]interface{}{
		"message_id": r.nextMessageID(),
		"date":       time.Now().Unix(),
		"chat": map[string]interface{}{
			"id":   chatID,
			"type": "private",
		},
	}

	// Reflect input parameters
	if text, ok := params["text"].(string); ok {
		msg["text"] = text
	}

	return msg
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/server/... -run TestGenerateResponse`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/server/responder*.go
git commit -m "feat: add response generator with param reflection"
```

---

### Task 12: Integrate Validator and Responder into BotHandler

**Files:**
- Modify: `internal/server/bot_handler.go`

**Step 1: Update BotHandler**

```go
// Update BotHandler struct:
type BotHandler struct {
	registry        *tokens.Registry
	registryEnabled bool
	validator       *Validator
	responder       *Responder
}

// Update NewBotHandler:
func NewBotHandler(registry *tokens.Registry, registryEnabled bool) *BotHandler {
	return &BotHandler{
		registry:        registry,
		registryEnabled: registryEnabled,
		validator:       NewValidator(),
		responder:       NewResponder(),
	}
}

// Update Handle method to use validator and responder:
func (h *BotHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	method := chi.URLParam(r, "method")

	w.Header().Set("Content-Type", "application/json")

	// Token validation (unchanged)
	if !tokens.ValidateFormat(token) {
		h.writeError(w, 401, "Unauthorized: invalid token format")
		return
	}

	if h.registryEnabled {
		info, ok := h.registry.Get(token)
		if !ok {
			h.writeError(w, 401, "Unauthorized: token not registered")
			return
		}
		switch info.Status {
		case tokens.StatusBanned:
			h.writeError(w, 403, "Forbidden: bot was banned")
			return
		case tokens.StatusDeactivated:
			h.writeError(w, 401, "Unauthorized: bot was deactivated")
			return
		}
	}

	// Check method exists
	spec, ok := gen.Methods[method]
	if !ok {
		h.writeError(w, 404, "Not Found: method not found")
		return
	}

	// Parse parameters
	params, err := h.parseParams(r)
	if err != nil {
		h.writeError(w, 400, "Bad Request: "+err.Error())
		return
	}

	// Validate request
	if err := h.validator.Validate(spec, params); err != nil {
		h.writeError(w, 400, "Bad Request: "+err.Error())
		return
	}

	// Generate response
	result, err := h.responder.Generate(spec, params)
	if err != nil {
		h.writeError(w, 500, "Internal Server Error")
		return
	}

	h.writeSuccess(w, result)
}

func (h *BotHandler) parseParams(r *http.Request) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// Parse query params
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Parse JSON body if present
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			return nil, err
		}
	}

	// Parse form data
	if err := r.ParseForm(); err == nil {
		for key, values := range r.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	return params, nil
}
```

**Step 2: Test integration**

Run: `go run ./cmd/tg-mock`

```bash
# getMe
curl http://localhost:8081/bot123:abc/getMe
# Expected: {"ok":true,"result":{"id":123456789,"is_bot":true,...}}

# sendMessage
curl -X POST http://localhost:8081/bot123:abc/sendMessage \
  -H "Content-Type: application/json" \
  -d '{"chat_id": 999, "text": "Hello!"}'
# Expected: {"ok":true,"result":{"message_id":1,"chat":{"id":999,...},"text":"Hello!"}}

# Missing required param
curl -X POST http://localhost:8081/bot123:abc/sendMessage \
  -H "Content-Type: application/json" \
  -d '{"chat_id": 999}'
# Expected: {"ok":false,"error_code":400,"description":"Bad Request: missing required field: text"}
```

**Step 3: Commit**

```bash
git add internal/server/bot_handler.go
git commit -m "feat: integrate validator and responder into bot handler"
```

---

## Phase 5: Scenario Engine

### Task 13: Scenario Types

**Files:**
- Create: `internal/scenario/scenario.go`
- Create: `internal/scenario/scenario_test.go`

**Step 1: Write failing test**

```go
// internal/scenario/scenario_test.go
package scenario

import "testing"

func TestScenarioMatch(t *testing.T) {
	s := &Scenario{
		ID:     "test-1",
		Method: "sendMessage",
		Match: map[string]interface{}{
			"chat_id": float64(123),
		},
		Times: 1,
		Response: &ErrorResponse{
			ErrorCode:   400,
			Description: "Bad Request: chat not found",
		},
	}

	// Should match
	params := map[string]interface{}{
		"chat_id": float64(123),
		"text":    "hello",
	}
	if !s.Matches("sendMessage", params) {
		t.Error("expected scenario to match")
	}

	// Wrong method
	if s.Matches("getMe", params) {
		t.Error("expected scenario not to match different method")
	}

	// Wrong chat_id
	params2 := map[string]interface{}{
		"chat_id": float64(456),
	}
	if s.Matches("sendMessage", params2) {
		t.Error("expected scenario not to match different chat_id")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/scenario/...`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/scenario/scenario.go
package scenario

import (
	"sync"
	"sync/atomic"
)

type Scenario struct {
	ID       string                 `json:"id"`
	Method   string                 `json:"method"`
	Match    map[string]interface{} `json:"match,omitempty"`
	Times    int                    `json:"times"` // 0 = unlimited
	Response *ErrorResponse         `json:"response,omitempty"`

	used int32 // atomic counter
}

type ErrorResponse struct {
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
	RetryAfter  int    `json:"retry_after,omitempty"`
}

func (s *Scenario) Matches(method string, params map[string]interface{}) bool {
	// Check method
	if s.Method != "*" && s.Method != method {
		return false
	}

	// Check match criteria
	for key, expected := range s.Match {
		actual, ok := params[key]
		if !ok {
			return false
		}
		if actual != expected {
			return false
		}
	}

	return true
}

func (s *Scenario) Use() bool {
	if s.Times == 0 {
		return true // Unlimited
	}

	used := atomic.AddInt32(&s.used, 1)
	return int(used) <= s.Times
}

func (s *Scenario) Exhausted() bool {
	if s.Times == 0 {
		return false
	}
	return int(atomic.LoadInt32(&s.used)) >= s.Times
}

// Engine manages scenarios
type Engine struct {
	mu        sync.RWMutex
	scenarios []*Scenario
	idCounter int64
}

func NewEngine() *Engine {
	return &Engine{
		scenarios: make([]*Scenario, 0),
	}
}

func (e *Engine) Add(s *Scenario) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if s.ID == "" {
		s.ID = e.generateID()
	}

	e.scenarios = append(e.scenarios, s)
	return s.ID
}

func (e *Engine) generateID() string {
	id := atomic.AddInt64(&e.idCounter, 1)
	return fmt.Sprintf("scenario-%d", id)
}

func (e *Engine) Find(method string, params map[string]interface{}) *Scenario {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, s := range e.scenarios {
		if s.Matches(method, params) && !s.Exhausted() {
			return s
		}
	}
	return nil
}

func (e *Engine) List() []*Scenario {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Scenario, len(e.scenarios))
	copy(result, e.scenarios)
	return result
}

func (e *Engine) Remove(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, s := range e.scenarios {
		if s.ID == id {
			e.scenarios = append(e.scenarios[:i], e.scenarios[i+1:]...)
			return true
		}
	}
	return false
}

func (e *Engine) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.scenarios = make([]*Scenario, 0)
}
```

Add missing import:
```go
import (
	"fmt"
	"sync"
	"sync/atomic"
)
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/scenario/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/scenario/
git commit -m "feat: add scenario engine for error simulation"
```

---

### Task 14: Integrate Scenarios into BotHandler

**Files:**
- Modify: `internal/server/server.go`
- Modify: `internal/server/bot_handler.go`

**Step 1: Add scenario engine to Server**

Update `internal/server/server.go`:

```go
import (
	// ... existing imports
	"github.com/watzon/tg-mock/internal/scenario"
)

type Server struct {
	router         chi.Router
	httpServer     *http.Server
	port           int
	tokenRegistry  *tokens.Registry
	scenarioEngine *scenario.Engine
	botHandler     *BotHandler
}

func New(cfg Config) *Server {
	// ...
	scenarioEngine := scenario.NewEngine()

	s := &Server{
		router:         r,
		port:           cfg.Port,
		tokenRegistry:  registry,
		scenarioEngine: scenarioEngine,
		botHandler:     NewBotHandler(registry, scenarioEngine, false),
	}
	// ...
}
```

**Step 2: Update BotHandler to check scenarios**

```go
// Update BotHandler struct:
type BotHandler struct {
	registry        *tokens.Registry
	registryEnabled bool
	scenarios       *scenario.Engine
	validator       *Validator
	responder       *Responder
}

// Update NewBotHandler:
func NewBotHandler(registry *tokens.Registry, scenarios *scenario.Engine, registryEnabled bool) *BotHandler {
	return &BotHandler{
		registry:        registry,
		registryEnabled: registryEnabled,
		scenarios:       scenarios,
		validator:       NewValidator(),
		responder:       NewResponder(),
	}
}

// Update Handle method - add scenario check after token validation:
func (h *BotHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// ... token validation code ...

	// Check method exists
	spec, ok := gen.Methods[method]
	if !ok {
		h.writeError(w, 404, "Not Found: method not found")
		return
	}

	// Parse parameters
	params, err := h.parseParams(r)
	if err != nil {
		h.writeError(w, 400, "Bad Request: "+err.Error())
		return
	}

	// Check for header-based scenario
	if scenarioName := r.Header.Get("X-TG-Mock-Scenario"); scenarioName != "" {
		if resp := h.handleHeaderScenario(w, r, scenarioName); resp {
			return
		}
	}

	// Check for queued scenarios
	if s := h.scenarios.Find(method, params); s != nil {
		s.Use()
		if s.Response != nil {
			h.writeErrorResponse(w, s.Response)
			return
		}
	}

	// Validate request
	if err := h.validator.Validate(spec, params); err != nil {
		h.writeError(w, 400, "Bad Request: "+err.Error())
		return
	}

	// Generate response
	result, err := h.responder.Generate(spec, params)
	if err != nil {
		h.writeError(w, 500, "Internal Server Error")
		return
	}

	h.writeSuccess(w, result)
}

func (h *BotHandler) handleHeaderScenario(w http.ResponseWriter, r *http.Request, name string) bool {
	// Will be implemented with pre-built errors
	return false
}

func (h *BotHandler) writeErrorResponse(w http.ResponseWriter, resp *scenario.ErrorResponse) {
	w.WriteHeader(resp.ErrorCode)
	response := map[string]interface{}{
		"ok":          false,
		"error_code":  resp.ErrorCode,
		"description": resp.Description,
	}
	if resp.RetryAfter > 0 {
		response["parameters"] = map[string]interface{}{
			"retry_after": resp.RetryAfter,
		}
	}
	json.NewEncoder(w).Encode(response)
}
```

**Step 3: Test scenario integration**

Run: `go build ./... && go test ./...`
Expected: All tests pass

**Step 4: Commit**

```bash
git add internal/server/
git commit -m "feat: integrate scenario engine into bot handler"
```

---

## Phase 6: Control API

### Task 15: Control API - Scenarios Endpoints

**Files:**
- Create: `internal/server/control_handler.go`

**Step 1: Create control handler**

```go
// internal/server/control_handler.go
package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/watzon/tg-mock/internal/scenario"
	"github.com/watzon/tg-mock/internal/tokens"
)

type ControlHandler struct {
	scenarios *scenario.Engine
	tokens    *tokens.Registry
}

func NewControlHandler(scenarios *scenario.Engine, tokens *tokens.Registry) *ControlHandler {
	return &ControlHandler{
		scenarios: scenarios,
		tokens:    tokens,
	}
}

func (h *ControlHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Scenarios
	r.Route("/scenarios", func(r chi.Router) {
		r.Get("/", h.listScenarios)
		r.Post("/", h.addScenario)
		r.Delete("/", h.clearScenarios)
		r.Delete("/{id}", h.removeScenario)
	})

	// Tokens
	r.Route("/tokens", func(r chi.Router) {
		r.Post("/", h.registerToken)
		r.Delete("/{token}", h.deleteToken)
		r.Patch("/{token}", h.updateToken)
	})

	// State
	r.Post("/reset", h.reset)
	r.Get("/state", h.getState)

	return r
}

// Scenarios handlers

func (h *ControlHandler) listScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios := h.scenarios.List()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenarios": scenarios,
	})
}

func (h *ControlHandler) addScenario(w http.ResponseWriter, r *http.Request) {
	var s scenario.Scenario
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := h.scenarios.Add(&s)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

func (h *ControlHandler) clearScenarios(w http.ResponseWriter, r *http.Request) {
	h.scenarios.Clear()
	w.WriteHeader(http.StatusNoContent)
}

func (h *ControlHandler) removeScenario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if h.scenarios.Remove(id) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "scenario not found", http.StatusNotFound)
	}
}

// Token handlers

func (h *ControlHandler) registerToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token   string        `json:"token"`
		Status  tokens.Status `json:"status"`
		BotName string        `json:"bot_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = tokens.StatusActive
	}

	h.tokens.Register(req.Token, tokens.TokenInfo{
		Status:  req.Status,
		BotName: req.BotName,
	})

	w.WriteHeader(http.StatusCreated)
}

func (h *ControlHandler) deleteToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	h.tokens.Delete(token)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ControlHandler) updateToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var req struct {
		Status tokens.Status `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.tokens.UpdateStatus(token, req.Status) {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "token not found", http.StatusNotFound)
	}
}

// State handlers

func (h *ControlHandler) reset(w http.ResponseWriter, r *http.Request) {
	h.scenarios.Clear()
	// TODO: Clear updates, files, etc.
	w.WriteHeader(http.StatusNoContent)
}

func (h *ControlHandler) getState(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenarios_count": len(h.scenarios.List()),
	})
}
```

**Step 2: Update server.go to mount control routes**

```go
// Add controlHandler to Server struct
type Server struct {
	// ...
	controlHandler *ControlHandler
}

// Update New():
func New(cfg Config) *Server {
	// ...
	s := &Server{
		router:         r,
		port:           cfg.Port,
		tokenRegistry:  registry,
		scenarioEngine: scenarioEngine,
		botHandler:     NewBotHandler(registry, scenarioEngine, false),
		controlHandler: NewControlHandler(scenarioEngine, registry),
	}
	// ...
}

// Update setupRoutes():
func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	s.router.Mount("/__control", s.controlHandler.Routes())

	s.router.Route("/bot{token}", func(r chi.Router) {
		r.Post("/{method}", s.botHandler.Handle)
		r.Get("/{method}", s.botHandler.Handle)
	})
}
```

**Step 3: Test control API**

Run: `go run ./cmd/tg-mock`

```bash
# Add scenario
curl -X POST http://localhost:8081/__control/scenarios \
  -H "Content-Type: application/json" \
  -d '{"method":"sendMessage","match":{"chat_id":999},"times":1,"response":{"error_code":400,"description":"Bad Request: chat not found"}}'

# List scenarios
curl http://localhost:8081/__control/scenarios

# Test scenario triggers
curl -X POST http://localhost:8081/bot123:abc/sendMessage \
  -H "Content-Type: application/json" \
  -d '{"chat_id":999,"text":"test"}'
# Expected: {"ok":false,"error_code":400,...}
```

**Step 4: Commit**

```bash
git add internal/server/control_handler.go internal/server/server.go
git commit -m "feat: add control API for scenarios and tokens"
```

---

## Phase 7: Updates System

### Task 16: Update Queue

**Files:**
- Create: `internal/updates/queue.go`
- Create: `internal/updates/queue_test.go`

**Step 1: Write failing test**

```go
// internal/updates/queue_test.go
package updates

import "testing"

func TestQueue(t *testing.T) {
	q := NewQueue()

	// Add updates
	q.Add(map[string]interface{}{
		"update_id": int64(1),
		"message":   map[string]interface{}{"text": "hello"},
	})
	q.Add(map[string]interface{}{
		"update_id": int64(2),
		"message":   map[string]interface{}{"text": "world"},
	})

	// Get updates
	updates := q.Get(0, 100)
	if len(updates) != 2 {
		t.Errorf("got %d updates, want 2", len(updates))
	}

	// Get with offset
	updates = q.Get(2, 100)
	if len(updates) != 1 {
		t.Errorf("got %d updates after offset, want 1", len(updates))
	}

	// Acknowledge (offset > update_id)
	q.Acknowledge(2)
	updates = q.Get(0, 100)
	if len(updates) != 1 {
		t.Errorf("got %d updates after ack, want 1", len(updates))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/updates/...`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/updates/queue.go
package updates

import (
	"sync"
	"sync/atomic"
)

type Queue struct {
	mu        sync.RWMutex
	updates   []map[string]interface{}
	idCounter int64
}

func NewQueue() *Queue {
	return &Queue{
		updates: make([]map[string]interface{}, 0),
	}
}

func (q *Queue) Add(update map[string]interface{}) int64 {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Assign update_id if not present
	if _, ok := update["update_id"]; !ok {
		update["update_id"] = atomic.AddInt64(&q.idCounter, 1)
	}

	q.updates = append(q.updates, update)
	return update["update_id"].(int64)
}

func (q *Queue) Get(offset int64, limit int) []map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]map[string]interface{}, 0)

	for _, u := range q.updates {
		updateID := u["update_id"].(int64)
		if offset == 0 || updateID >= offset {
			result = append(result, u)
			if len(result) >= limit {
				break
			}
		}
	}

	return result
}

func (q *Queue) Acknowledge(offset int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Remove updates with update_id < offset
	newUpdates := make([]map[string]interface{}, 0)
	for _, u := range q.updates {
		if u["update_id"].(int64) >= offset {
			newUpdates = append(newUpdates, u)
		}
	}
	q.updates = newUpdates
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.updates = make([]map[string]interface{}, 0)
}

func (q *Queue) Pending() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.updates)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/updates/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/updates/
git commit -m "feat: add update queue for getUpdates"
```

---

### Task 17: Updates Control Endpoints

**Files:**
- Modify: `internal/server/control_handler.go`
- Modify: `internal/server/server.go`

**Step 1: Add updates to control handler**

```go
// Update ControlHandler struct:
type ControlHandler struct {
	scenarios *scenario.Engine
	tokens    *tokens.Registry
	updates   *updates.Queue
}

// Update NewControlHandler:
func NewControlHandler(scenarios *scenario.Engine, tokens *tokens.Registry, updates *updates.Queue) *ControlHandler {
	return &ControlHandler{
		scenarios: scenarios,
		tokens:    tokens,
		updates:   updates,
	}
}

// Add to Routes():
r.Route("/updates", func(r chi.Router) {
	r.Get("/", h.listUpdates)
	r.Post("/", h.addUpdate)
	r.Delete("/", h.clearUpdates)
})

// Add handlers:
func (h *ControlHandler) listUpdates(w http.ResponseWriter, r *http.Request) {
	updates := h.updates.Get(0, 100)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updates": updates,
		"pending": h.updates.Pending(),
	})
}

func (h *ControlHandler) addUpdate(w http.ResponseWriter, r *http.Request) {
	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := h.updates.Add(update)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"update_id": id,
	})
}

func (h *ControlHandler) clearUpdates(w http.ResponseWriter, r *http.Request) {
	h.updates.Clear()
	w.WriteHeader(http.StatusNoContent)
}

// Update reset():
func (h *ControlHandler) reset(w http.ResponseWriter, r *http.Request) {
	h.scenarios.Clear()
	h.updates.Clear()
	w.WriteHeader(http.StatusNoContent)
}

// Update getState():
func (h *ControlHandler) getState(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenarios_count": len(h.scenarios.List()),
		"updates_pending": h.updates.Pending(),
	})
}
```

**Step 2: Update server.go**

```go
import (
	// ...
	"github.com/watzon/tg-mock/internal/updates"
)

type Server struct {
	// ...
	updateQueue *updates.Queue
}

func New(cfg Config) *Server {
	// ...
	updateQueue := updates.NewQueue()

	s := &Server{
		router:         r,
		port:           cfg.Port,
		tokenRegistry:  registry,
		scenarioEngine: scenarioEngine,
		updateQueue:    updateQueue,
		botHandler:     NewBotHandler(registry, scenarioEngine, updateQueue, false),
		controlHandler: NewControlHandler(scenarioEngine, registry, updateQueue),
	}
	// ...
}
```

**Step 3: Update BotHandler for getUpdates**

Add to `bot_handler.go`:

```go
// Update BotHandler struct:
type BotHandler struct {
	// ...
	updates *updates.Queue
}

// Update NewBotHandler:
func NewBotHandler(registry *tokens.Registry, scenarios *scenario.Engine, updates *updates.Queue, registryEnabled bool) *BotHandler {
	return &BotHandler{
		registry:        registry,
		registryEnabled: registryEnabled,
		scenarios:       scenarios,
		updates:         updates,
		validator:       NewValidator(),
		responder:       NewResponder(),
	}
}

// Add getUpdates handling in Handle():
// After response generation, before writeSuccess:
if method == "getUpdates" {
	result := h.handleGetUpdates(params)
	h.writeSuccess(w, result)
	return
}

// Add handler:
func (h *BotHandler) handleGetUpdates(params map[string]interface{}) []map[string]interface{} {
	offset := int64(0)
	if o, ok := params["offset"].(float64); ok {
		offset = int64(o)
	}

	limit := 100
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	// Acknowledge previous updates
	if offset > 0 {
		h.updates.Acknowledge(offset)
	}

	return h.updates.Get(offset, limit)
}
```

**Step 4: Test updates flow**

Run: `go run ./cmd/tg-mock`

```bash
# Add update via control API
curl -X POST http://localhost:8081/__control/updates \
  -H "Content-Type: application/json" \
  -d '{"message":{"message_id":1,"text":"Hello","chat":{"id":123,"type":"private"}}}'

# Get updates via bot API
curl http://localhost:8081/bot123:abc/getUpdates
# Expected: {"ok":true,"result":[{"update_id":1,"message":{...}}]}
```

**Step 5: Commit**

```bash
git add internal/server/ internal/updates/
git commit -m "feat: add updates queue and getUpdates handling"
```

---

## Phase 8: Pre-built Errors

### Task 18: Built-in Error Scenarios

**Files:**
- Create: `internal/scenario/builtin.go`

**Step 1: Create built-in errors**

```go
// internal/scenario/builtin.go
package scenario

var BuiltinErrors = map[string]*ErrorResponse{
	// 400 Bad Request
	"bad_request":              {ErrorCode: 400, Description: "Bad Request"},
	"chat_not_found":           {ErrorCode: 400, Description: "Bad Request: chat not found"},
	"user_not_found":           {ErrorCode: 400, Description: "Bad Request: user not found"},
	"message_not_found":        {ErrorCode: 400, Description: "Bad Request: message to edit not found"},
	"message_not_modified":     {ErrorCode: 400, Description: "Bad Request: message is not modified"},
	"message_text_empty":       {ErrorCode: 400, Description: "Bad Request: message text is empty"},
	"message_too_long":         {ErrorCode: 400, Description: "Bad Request: message is too long"},
	"message_cant_be_edited":   {ErrorCode: 400, Description: "Bad Request: message can't be edited"},
	"message_cant_be_deleted":  {ErrorCode: 400, Description: "Bad Request: message can't be deleted"},
	"reply_message_not_found":  {ErrorCode: 400, Description: "Bad Request: reply message not found"},
	"button_url_invalid":       {ErrorCode: 400, Description: "Bad Request: BUTTON_URL_INVALID"},
	"entities_too_long":        {ErrorCode: 400, Description: "Bad Request: entities too long"},
	"file_too_big":             {ErrorCode: 400, Description: "Bad Request: file is too big"},
	"invalid_file_id":          {ErrorCode: 400, Description: "Bad Request: invalid file id"},
	"member_not_found":         {ErrorCode: 400, Description: "Bad Request: member not found"},
	"group_deactivated":        {ErrorCode: 400, Description: "Bad Request: group is deactivated"},
	"peer_id_invalid":          {ErrorCode: 400, Description: "Bad Request: PEER_ID_INVALID"},
	"wrong_parameter_action":   {ErrorCode: 400, Description: "Bad Request: wrong parameter action in request"},

	// 401 Unauthorized
	"unauthorized": {ErrorCode: 401, Description: "Unauthorized"},

	// 403 Forbidden
	"forbidden":               {ErrorCode: 403, Description: "Forbidden"},
	"bot_blocked":             {ErrorCode: 403, Description: "Forbidden: bot was blocked by the user"},
	"bot_kicked":              {ErrorCode: 403, Description: "Forbidden: bot was kicked from the chat"},
	"cant_initiate":           {ErrorCode: 403, Description: "Forbidden: bot can't initiate conversation with a user"},
	"cant_send_to_bots":       {ErrorCode: 403, Description: "Forbidden: bot can't send messages to bots"},
	"not_member_channel":      {ErrorCode: 403, Description: "Forbidden: bot is not a member of the channel chat"},
	"not_member_supergroup":   {ErrorCode: 403, Description: "Forbidden: bot is not a member of the supergroup chat"},
	"user_deactivated":        {ErrorCode: 403, Description: "Forbidden: user is deactivated"},
	"not_enough_rights_text":  {ErrorCode: 403, Description: "Forbidden: not enough rights to send text messages"},
	"not_enough_rights_photo": {ErrorCode: 403, Description: "Forbidden: not enough rights to send photos"},

	// 409 Conflict
	"webhook_active":           {ErrorCode: 409, Description: "Conflict: can't use getUpdates method while webhook is active"},
	"terminated_by_long_poll":  {ErrorCode: 409, Description: "Conflict: terminated by other long poll"},

	// 429 Rate Limit
	"rate_limit":  {ErrorCode: 429, Description: "Too Many Requests: retry after 30", RetryAfter: 30},
	"flood_wait":  {ErrorCode: 429, Description: "Flood control exceeded. Retry in 60 seconds", RetryAfter: 60},
}

func GetBuiltinError(name string) *ErrorResponse {
	return BuiltinErrors[name]
}
```

**Step 2: Update BotHandler to use builtin errors**

Update `handleHeaderScenario` in `bot_handler.go`:

```go
func (h *BotHandler) handleHeaderScenario(w http.ResponseWriter, r *http.Request, name string) bool {
	resp := scenario.GetBuiltinError(name)
	if resp == nil {
		return false
	}

	// Allow retry_after override via header
	if retryAfter := r.Header.Get("X-TG-Mock-Retry-After"); retryAfter != "" {
		if val, err := strconv.Atoi(retryAfter); err == nil {
			resp = &scenario.ErrorResponse{
				ErrorCode:   resp.ErrorCode,
				Description: resp.Description,
				RetryAfter:  val,
			}
		}
	}

	h.writeErrorResponse(w, resp)
	return true
}
```

Add import: `"strconv"`

**Step 3: Test header-based errors**

Run: `go run ./cmd/tg-mock`

```bash
# Rate limit
curl -H "X-TG-Mock-Scenario: rate_limit" \
  http://localhost:8081/bot123:abc/sendMessage
# Expected: {"ok":false,"error_code":429,...,"parameters":{"retry_after":30}}

# Custom retry_after
curl -H "X-TG-Mock-Scenario: rate_limit" \
  -H "X-TG-Mock-Retry-After: 120" \
  http://localhost:8081/bot123:abc/sendMessage
# Expected: retry_after: 120

# Bot blocked
curl -H "X-TG-Mock-Scenario: bot_blocked" \
  http://localhost:8081/bot123:abc/sendMessage
# Expected: {"ok":false,"error_code":403,"description":"Forbidden: bot was blocked..."}
```

**Step 4: Commit**

```bash
git add internal/scenario/builtin.go internal/server/bot_handler.go
git commit -m "feat: add pre-built error scenarios with header support"
```

---

## Phase 9: Configuration

### Task 19: Config Types and Loading

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write failing test**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yaml := `
server:
  port: 9000
  verbose: true

tokens:
  "123:abc":
    status: active
    bot_name: TestBot
  "456:def":
    status: banned
`

	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString(yaml)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Port != 9000 {
		t.Errorf("port = %d, want 9000", cfg.Server.Port)
	}

	if !cfg.Server.Verbose {
		t.Error("verbose should be true")
	}

	if len(cfg.Tokens) != 2 {
		t.Errorf("got %d tokens, want 2", len(cfg.Tokens))
	}

	if cfg.Tokens["123:abc"].Status != "active" {
		t.Error("token status should be active")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/...`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig            `yaml:"server"`
	Storage   StorageConfig           `yaml:"storage"`
	Tokens    map[string]TokenConfig  `yaml:"tokens"`
	Scenarios []ScenarioConfig        `yaml:"scenarios"`
}

type ServerConfig struct {
	Port    int  `yaml:"port"`
	Verbose bool `yaml:"verbose"`
	Strict  bool `yaml:"strict"`
}

type StorageConfig struct {
	Dir string `yaml:"dir"`
}

type TokenConfig struct {
	Status  string `yaml:"status"`
	BotName string `yaml:"bot_name"`
}

type ScenarioConfig struct {
	Method   string                 `yaml:"method"`
	Match    map[string]interface{} `yaml:"match"`
	Times    int                    `yaml:"times"`
	Response ResponseConfig         `yaml:"response"`
}

type ResponseConfig struct {
	ErrorCode   int    `yaml:"error_code"`
	Description string `yaml:"description"`
	RetryAfter  int    `yaml:"retry_after"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    8081,
			Verbose: false,
			Strict:  false,
		},
		Tokens:    make(map[string]TokenConfig),
		Scenarios: make([]ScenarioConfig, 0),
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add YAML config loading"
```

---

### Task 20: Integrate Config into Main

**Files:**
- Modify: `cmd/tg-mock/main.go`
- Modify: `internal/server/server.go`

**Step 1: Update main.go**

```go
// cmd/tg-mock/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/watzon/tg-mock/internal/config"
	"github.com/watzon/tg-mock/internal/server"
)

func main() {
	port := flag.Int("port", 0, "HTTP server port (overrides config)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging (overrides config)")
	configPath := flag.String("config", "", "Path to config file")
	storageDir := flag.String("storage-dir", "", "Directory for file storage")
	flag.Parse()

	// Load config
	var cfg *config.Config
	if *configPath != "" {
		var err error
		cfg, err = config.Load(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// CLI overrides
	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *verbose {
		cfg.Server.Verbose = true
	}
	if *storageDir != "" {
		cfg.Storage.Dir = *storageDir
	}

	srv := server.New(server.Config{
		Port:       cfg.Server.Port,
		Verbose:    cfg.Server.Verbose,
		Tokens:     cfg.Tokens,
		Scenarios:  cfg.Scenarios,
		StorageDir: cfg.Storage.Dir,
	})

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down...")
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Update server.go Config**

```go
import (
	"github.com/watzon/tg-mock/internal/config"
)

type Config struct {
	Port       int
	Verbose    bool
	Tokens     map[string]config.TokenConfig
	Scenarios  []config.ScenarioConfig
	StorageDir string
}

func New(cfg Config) *Server {
	r := chi.NewRouter()

	if cfg.Verbose {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	registry := tokens.NewRegistry()
	scenarioEngine := scenario.NewEngine()
	updateQueue := updates.NewQueue()

	// Load tokens from config
	for token, info := range cfg.Tokens {
		registry.Register(token, tokens.TokenInfo{
			Status:  tokens.Status(info.Status),
			BotName: info.BotName,
		})
	}

	// Load scenarios from config
	for _, sc := range cfg.Scenarios {
		scenarioEngine.Add(&scenario.Scenario{
			Method: sc.Method,
			Match:  sc.Match,
			Times:  sc.Times,
			Response: &scenario.ErrorResponse{
				ErrorCode:   sc.Response.ErrorCode,
				Description: sc.Response.Description,
				RetryAfter:  sc.Response.RetryAfter,
			},
		})
	}

	registryEnabled := len(cfg.Tokens) > 0

	s := &Server{
		router:         r,
		port:           cfg.Port,
		tokenRegistry:  registry,
		scenarioEngine: scenarioEngine,
		updateQueue:    updateQueue,
		botHandler:     NewBotHandler(registry, scenarioEngine, updateQueue, registryEnabled),
		controlHandler: NewControlHandler(scenarioEngine, registry, updateQueue),
	}

	s.setupRoutes()

	return s
}
```

**Step 3: Test with config file**

Create `test-config.yaml`:
```yaml
server:
  port: 9090
  verbose: true

tokens:
  "123:testtoken":
    status: active
    bot_name: TestBot
```

Run: `go run ./cmd/tg-mock --config test-config.yaml`
Expected: Server starts on port 9090 with verbose logging

**Step 4: Commit**

```bash
git add cmd/tg-mock/ internal/server/server.go
git commit -m "feat: integrate config file with CLI overrides"
```

---

## Phase 10: File Storage

### Task 21: File Storage Interface and Memory Implementation

**Files:**
- Create: `internal/storage/store.go`
- Create: `internal/storage/memory.go`
- Create: `internal/storage/memory_test.go`

**Step 1: Write failing test**

```go
// internal/storage/memory_test.go
package storage

import "testing"

func TestMemoryStore(t *testing.T) {
	s := NewMemoryStore()

	// Store file
	data := []byte("hello world")
	fileID, err := s.Store(data, "test.txt", "text/plain")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if fileID == "" {
		t.Error("fileID should not be empty")
	}

	// Get file
	retrieved, meta, err := s.Get(fileID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(data) {
		t.Errorf("got %q, want %q", retrieved, data)
	}

	if meta.Filename != "test.txt" {
		t.Errorf("filename = %q, want test.txt", meta.Filename)
	}

	// Get path
	path, err := s.GetPath(fileID)
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}
	if path == "" {
		t.Error("path should not be empty")
	}

	// Delete
	if err := s.Delete(fileID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should be gone
	_, _, err = s.Get(fileID)
	if err == nil {
		t.Error("expected error after delete")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/storage/...`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/storage/store.go
package storage

import "errors"

var ErrNotFound = errors.New("file not found")

type FileMetadata struct {
	Filename string
	MimeType string
	Size     int64
}

type Store interface {
	Store(data []byte, filename string, mimeType string) (fileID string, err error)
	Get(fileID string) (data []byte, metadata FileMetadata, err error)
	GetPath(fileID string) (filePath string, err error)
	Delete(fileID string) error
	Clear() error
}
```

```go
// internal/storage/memory.go
package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

type memoryFile struct {
	data     []byte
	metadata FileMetadata
	path     string
}

type MemoryStore struct {
	mu    sync.RWMutex
	files map[string]*memoryFile
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		files: make(map[string]*memoryFile),
	}
}

func (s *MemoryStore) Store(data []byte, filename string, mimeType string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileID := s.generateFileID()
	path := fmt.Sprintf("documents/%s", filename)

	s.files[fileID] = &memoryFile{
		data: data,
		metadata: FileMetadata{
			Filename: filename,
			MimeType: mimeType,
			Size:     int64(len(data)),
		},
		path: path,
	}

	return fileID, nil
}

func (s *MemoryStore) Get(fileID string) ([]byte, FileMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, ok := s.files[fileID]
	if !ok {
		return nil, FileMetadata{}, ErrNotFound
	}

	return file.data, file.metadata, nil
}

func (s *MemoryStore) GetPath(fileID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, ok := s.files[fileID]
	if !ok {
		return "", ErrNotFound
	}

	return file.path, nil
}

func (s *MemoryStore) Delete(fileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.files, fileID)
	return nil
}

func (s *MemoryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.files = make(map[string]*memoryFile)
	return nil
}

func (s *MemoryStore) generateFileID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/storage/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/storage/
git commit -m "feat: add file storage interface and memory implementation"
```

---

### Task 22: File Download Endpoint

**Files:**
- Modify: `internal/server/server.go`

**Step 1: Add file route and storage to server**

```go
import (
	"github.com/watzon/tg-mock/internal/storage"
)

type Server struct {
	// ... existing fields
	fileStore storage.Store
}

func New(cfg Config) *Server {
	// ... existing setup

	var fileStore storage.Store
	if cfg.StorageDir != "" {
		// TODO: disk store
		fileStore = storage.NewMemoryStore()
	} else {
		fileStore = storage.NewMemoryStore()
	}

	s := &Server{
		// ... existing fields
		fileStore: fileStore,
	}
	// ...
}

func (s *Server) setupRoutes() {
	// ... existing routes

	// File download
	s.router.Get("/file/bot{token}/{path:.*}", s.handleFileDownload)
}

func (s *Server) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	path := chi.URLParam(r, "path")

	// Validate token format
	if !tokens.ValidateFormat(token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Find file by path
	// For now, return 404 - actual implementation needs file lookup by path
	http.Error(w, "File not found", http.StatusNotFound)
}
```

**Step 2: Commit**

```bash
git add internal/server/server.go
git commit -m "feat: add file download endpoint placeholder"
```

---

## Phase 11: Finish & Polish

### Task 23: Add getFile Response

**Files:**
- Modify: `internal/server/responder.go`

**Step 1: Update responder to handle getFile**

Add to `Generate` method:

```go
case "File":
	return r.generateFile(params), nil
```

Add method:

```go
func (r *Responder) generateFile(params map[string]interface{}) map[string]interface{} {
	fileID := "placeholder"
	if id, ok := params["file_id"].(string); ok {
		fileID = id
	}

	return map[string]interface{}{
		"file_id":        fileID,
		"file_unique_id": "unique_" + fileID[:8],
		"file_size":      1024,
		"file_path":      "documents/file.txt",
	}
}
```

**Step 2: Commit**

```bash
git add internal/server/responder.go
git commit -m "feat: add getFile response generation"
```

---

### Task 24: README

**Files:**
- Create: `README.md`

**Step 1: Write README**

```markdown
# tg-mock

A mock Telegram Bot API server for testing bots and bot libraries.

## Installation

```bash
go install github.com/watzon/tg-mock@latest
```

## Usage

```bash
# Start with defaults (port 8081)
tg-mock

# Custom port
tg-mock --port 9090

# With config file
tg-mock --config config.yaml

# Verbose logging
tg-mock --verbose
```

## Configuration

```yaml
server:
  port: 8081
  verbose: true

tokens:
  "123456789:ABC-xyz":
    status: active
    bot_name: MyBot

scenarios:
  - method: sendMessage
    match:
      chat_id: 999
    response:
      error_code: 400
      description: "Bad Request: chat not found"
```

## Control API

### Scenarios

```bash
# Add scenario
curl -X POST http://localhost:8081/__control/scenarios \
  -d '{"method":"sendMessage","times":1,"response":{"error_code":429,"description":"Too Many Requests","retry_after":30}}'

# List scenarios
curl http://localhost:8081/__control/scenarios

# Clear scenarios
curl -X DELETE http://localhost:8081/__control/scenarios
```

### Updates

```bash
# Inject update
curl -X POST http://localhost:8081/__control/updates \
  -d '{"message":{"message_id":1,"text":"Hello","chat":{"id":123,"type":"private"}}}'

# View pending updates
curl http://localhost:8081/__control/updates
```

### Header-based Errors

```bash
curl -H "X-TG-Mock-Scenario: rate_limit" \
  http://localhost:8081/bot123:abc/sendMessage

curl -H "X-TG-Mock-Scenario: bot_blocked" \
  http://localhost:8081/bot123:abc/sendMessage
```

## License

MIT
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README"
```

---

### Task 25: Final Integration Test

**Files:**
- Create: `integration_test.go`

**Step 1: Write integration test**

```go
// integration_test.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/watzon/tg-mock/internal/server"
)

func TestIntegration(t *testing.T) {
	srv := server.New(server.Config{
		Port:    0,
		Verbose: false,
	})

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	t.Run("getMe", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/bot123:abc/getMe")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["ok"] != true {
			t.Error("expected ok=true")
		}
	})

	t.Run("sendMessage", func(t *testing.T) {
		body := bytes.NewBufferString(`{"chat_id":123,"text":"Hello"}`)
		resp, err := http.Post(ts.URL+"/bot123:abc/sendMessage", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["ok"] != true {
			t.Errorf("expected ok=true, got %v", result)
		}
	})

	t.Run("scenario error", func(t *testing.T) {
		// Add scenario
		scenario := `{"method":"sendMessage","match":{"chat_id":999},"times":1,"response":{"error_code":400,"description":"Bad Request: chat not found"}}`
		http.Post(ts.URL+"/__control/scenarios", "application/json", bytes.NewBufferString(scenario))

		// Trigger scenario
		body := bytes.NewBufferString(`{"chat_id":999,"text":"test"}`)
		resp, err := http.Post(ts.URL+"/bot123:abc/sendMessage", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("header scenario", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/bot123:abc/sendMessage", bytes.NewBufferString(`{"chat_id":1,"text":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-TG-Mock-Scenario", "rate_limit")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 429 {
			t.Errorf("expected 429, got %d", resp.StatusCode)
		}
	})
}
```

**Step 2: Add Router() method to server**

```go
func (s *Server) Router() chi.Router {
	return s.router
}
```

**Step 3: Run tests**

Run: `go test -v ./...`
Expected: All tests pass

**Step 4: Commit**

```bash
git add integration_test.go internal/server/server.go
git commit -m "test: add integration tests"
```

---

## Summary

This plan covers:

1. **Project Setup** - Directory structure, dependencies, Makefile
2. **Code Generation** - Parse spec, generate types, methods, fixtures
3. **Core Server** - HTTP server with chi, routing, token validation
4. **Request Handling** - Validation, response generation with reflection
5. **Scenario Engine** - Match scenarios, return configured errors
6. **Control API** - Manage scenarios, tokens, updates
7. **Updates System** - Queue for getUpdates, update injection
8. **Pre-built Errors** - Common Telegram errors ready to use
9. **Configuration** - YAML config with CLI overrides
10. **File Storage** - In-memory storage for uploads
11. **Polish** - README, integration tests

Each task is 5-15 minutes and follows TDD where applicable.
