package main

import (
	"log"

	"go-runner/internal/config"
	"go-runner/internal/hotreload"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create and start hot reload watcher
	watcher := hotreload.NewWatcher(&cfg.HotReload)
	
	if err := watcher.Start(); err != nil {
		log.Fatalf("Failed to start hot reload: %v", err)
	}

	// Keep the program running
	select {}
}
