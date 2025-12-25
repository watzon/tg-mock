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
	fakerSeed := flag.Int64("faker-seed", 0, "Seed for faker (0 = random, >0 = deterministic)")
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
	if *fakerSeed != 0 {
		cfg.Server.FakerSeed = *fakerSeed
	}

	srv := server.New(server.Config{
		Port:       cfg.Server.Port,
		Verbose:    cfg.Server.Verbose,
		FakerSeed:  cfg.Server.FakerSeed,
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
