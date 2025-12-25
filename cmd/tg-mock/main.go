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
