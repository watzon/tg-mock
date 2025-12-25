// internal/server/server.go
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/watzon/tg-mock/internal/config"
	"github.com/watzon/tg-mock/internal/scenario"
	"github.com/watzon/tg-mock/internal/storage"
	"github.com/watzon/tg-mock/internal/tokens"
	"github.com/watzon/tg-mock/internal/updates"
)

type Server struct {
	router         chi.Router
	httpServer     *http.Server
	port           int
	tokenRegistry  *tokens.Registry
	scenarioEngine *scenario.Engine
	updateQueue    *updates.Queue
	fileStore      storage.Store
	botHandler     *BotHandler
	controlHandler *ControlHandler
}

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

	// Enable token registry if any tokens are configured
	registryEnabled := len(cfg.Tokens) > 0

	// Create file storage (for now always use MemoryStore)
	var fileStore storage.Store
	if cfg.StorageDir != "" {
		// TODO: implement disk store when needed
		fileStore = storage.NewMemoryStore()
	} else {
		fileStore = storage.NewMemoryStore()
	}

	s := &Server{
		router:         r,
		port:           cfg.Port,
		tokenRegistry:  registry,
		scenarioEngine: scenarioEngine,
		updateQueue:    updateQueue,
		fileStore:      fileStore,
		botHandler:     NewBotHandler(registry, scenarioEngine, updateQueue, registryEnabled),
		controlHandler: NewControlHandler(scenarioEngine, registry, updateQueue),
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Control API
	s.router.Mount("/__control", s.controlHandler.Routes())

	// Bot API routes
	s.router.Route("/bot{token}", func(r chi.Router) {
		r.Post("/{method}", s.botHandler.Handle)
		r.Get("/{method}", s.botHandler.Handle)
	})

	// File download endpoint
	s.router.Get("/file/bot{token}/{path:.*}", s.handleFileDownload)
}

func (s *Server) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	// path := chi.URLParam(r, "path") // Will be used in future implementation

	// Validate token format
	if !tokens.ValidateFormat(token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Find file by path
	// For now, return 404 - actual implementation needs file lookup by path
	http.Error(w, "File not found", http.StatusNotFound)
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
