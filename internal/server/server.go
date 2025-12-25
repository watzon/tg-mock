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
	// Truncate token for display if long enough
	displayToken := token
	if len(token) > 10 {
		displayToken = token[:10] + "..."
	}
	fmt.Fprintf(w, `{"ok":true,"method":%q,"token":%q}`, method, displayToken)
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
