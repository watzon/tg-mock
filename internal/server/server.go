// internal/server/server.go
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/watzon/tg-mock/internal/scenario"
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
	botHandler     *BotHandler
	controlHandler *ControlHandler
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

	registry := tokens.NewRegistry()
	scenarioEngine := scenario.NewEngine()
	updateQueue := updates.NewQueue()

	s := &Server{
		router:         r,
		port:           cfg.Port,
		tokenRegistry:  registry,
		scenarioEngine: scenarioEngine,
		updateQueue:    updateQueue,
		botHandler:     NewBotHandler(registry, scenarioEngine, updateQueue, false), // Registry disabled by default
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
