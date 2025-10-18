package hotreload

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"go-runner/internal/config"

	"github.com/fsnotify/fsnotify"
)

// Watcher handles file watching and hot reload
type Watcher struct {
	config     *config.HotReloadConfig
	watcher    *fsnotify.Watcher
	process    *exec.Cmd
	processCtx context.Context
	processCancel context.CancelFunc
	mu         sync.RWMutex
	restarting bool
}

// NewWatcher creates a new hot reload watcher
func NewWatcher(cfg *config.HotReloadConfig) *Watcher {
	return &Watcher{
		config: cfg,
	}
}

// Start starts the hot reload watcher
func (w *Watcher) Start() error {
	if !w.config.Enabled {
		log.Println("Hot reload is disabled")
		return nil
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	w.watcher = watcher

	// Add watch directories
	for _, dir := range w.config.WatchDirs {
		if err := w.addWatchDir(dir); err != nil {
			log.Printf("Warning: failed to watch directory %s: %v", dir, err)
		}
	}

	// Start the initial process
	if err := w.startProcess(); err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}

	// Start watching for changes
	go w.watch()

	// Handle shutdown signals
	go w.handleSignals()

	log.Println("🔥 Hot reload watcher started")
	return nil
}

// Stop stops the hot reload watcher
func (w *Watcher) Stop() error {
	if w.watcher != nil {
		w.watcher.Close()
	}
	
	if w.processCancel != nil {
		w.processCancel()
	}
	
	if w.process != nil && w.process.Process != nil {
		w.process.Process.Kill()
	}
	
	log.Println("🛑 Hot reload watcher stopped")
	return nil
}

// addWatchDir adds a directory to watch
func (w *Watcher) addWatchDir(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", dir)
	}

	// Walk through directory and add all subdirectories
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if it's a file
		if !info.IsDir() {
			return nil
		}

		// Check if directory should be excluded
		if w.shouldExcludeDir(path) {
			return filepath.SkipDir
		}

		// Add directory to watcher
		if err := w.watcher.Add(path); err != nil {
			log.Printf("Warning: failed to add %s to watcher: %v", path, err)
		} else {
			log.Printf("Watching directory: %s", path)
		}

		return nil
	})
}

// shouldExcludeDir checks if a directory should be excluded
func (w *Watcher) shouldExcludeDir(path string) bool {
	for _, excludeDir := range w.config.ExcludeDirs {
		if strings.Contains(path, excludeDir) {
			return true
		}
	}
	return false
}

// shouldIncludeFile checks if a file should trigger a reload
func (w *Watcher) shouldIncludeFile(filename string) bool {
	ext := filepath.Ext(filename)
	
	// Check if extension should be excluded
	for _, excludeExt := range w.config.ExcludeExts {
		if ext == excludeExt {
			return false
		}
	}
	
	// Check if extension should be included
	for _, includeExt := range w.config.IncludeExts {
		if ext == includeExt {
			return true
		}
	}
	
	return false
}

// watch watches for file changes
func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Check if file should trigger reload
			if w.shouldIncludeFile(event.Name) {
				log.Printf("File changed: %s", event.Name)
				w.restartProcess()
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// startProcess starts the application process
func (w *Watcher) startProcess() error {
	// Create context for process
	w.processCtx, w.processCancel = context.WithCancel(context.Background())

	// Build the application
	log.Println("🔨 Building application...")
	
	// Parse build command and execute directly
	buildArgs := w.parseCommand(w.config.BuildCmd)
	buildCmd := exec.CommandContext(w.processCtx, buildArgs[0], buildArgs[1:]...)
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed: %v", err)
	}

	// Start the application
	log.Println("🚀 Starting application...")
	
	// Parse run command and execute directly
	runArgs := w.parseCommand(w.config.RunCmd)
	w.process = exec.CommandContext(w.processCtx, runArgs[0], runArgs[1:]...)
	w.process.Stdout = os.Stdout
	w.process.Stderr = os.Stderr

	if err := w.process.Start(); err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}

	log.Printf("✅ Application started with PID: %d", w.process.Process.Pid)
	return nil
}

// parseCommand parses a command string into executable and arguments
func (w *Watcher) parseCommand(cmdStr string) []string {
	// Simple command parsing - split by spaces
	// This is a basic implementation, for more complex commands you might need a proper shell parser
	parts := strings.Fields(cmdStr)
	
	if len(parts) == 0 {
		return parts
	}
	
	// Handle cross-platform executable paths and extensions
	if runtime.GOOS == "windows" {
		// On Windows, handle .exe extension and remove ./ prefix
		if parts[0] == "go" {
			// Keep go command as is
		} else {
			// Remove ./ prefix if present
			parts[0] = strings.TrimPrefix(parts[0], "./")
			// Add .exe extension if no extension is present and it's not a system command
			if !strings.Contains(parts[0], ".") && !w.isSystemCommand(parts[0]) {
				parts[0] = parts[0] + ".exe"
			}
		}
	} else {
		// On Unix-like systems, keep ./ prefix for relative paths
		// No special handling needed
	}
	
	return parts
}

// isSystemCommand checks if a command is a system command that shouldn't get .exe extension
func (w *Watcher) isSystemCommand(cmd string) bool {
	systemCommands := []string{"go", "git", "npm", "node", "python", "python3", "java", "javac", "gcc", "g++", "make", "cmake"}
	for _, sysCmd := range systemCommands {
		if cmd == sysCmd {
			return true
		}
	}
	return false
}

// restartProcess restarts the application process
func (w *Watcher) restartProcess() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.restarting {
		return
	}
	w.restarting = true

	// Kill current process
	if w.process != nil && w.process.Process != nil {
		log.Println("🛑 Stopping current process...")
		w.processCancel()
		w.process.Wait()
	}

	// Wait a bit before restarting
	time.Sleep(time.Duration(w.config.Delay) * time.Millisecond)

	// Start new process
	if err := w.startProcess(); err != nil {
		log.Printf("❌ Failed to restart process: %v", err)
	} else {
		log.Println("🔄 Process restarted successfully")
	}

	w.restarting = false
}

// handleSignals handles OS signals for graceful shutdown
func (w *Watcher) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 Received shutdown signal")
	w.Stop()
	os.Exit(0)
}
