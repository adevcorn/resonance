package agent

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches agent definition files for changes
type Watcher struct {
	loader   *Loader
	pool     *Pool
	watcher  *fsnotify.Watcher
	enabled  bool
	debounce time.Duration
}

// NewWatcher creates a new file watcher for agent definitions
func NewWatcher(loader *Loader, pool *Pool, enabled bool) (*Watcher, error) {
	if !enabled {
		return &Watcher{
			loader:  loader,
			pool:    pool,
			enabled: false,
		}, nil
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &Watcher{
		loader:   loader,
		pool:     pool,
		watcher:  fsw,
		enabled:  true,
		debounce: 100 * time.Millisecond, // Debounce rapid file changes
	}, nil
}

// Start starts watching for file changes
func (w *Watcher) Start(ctx context.Context) error {
	if !w.enabled {
		log.Println("[agent/watcher] Hot-reload disabled")
		return nil
	}

	// Add agents directory to watch list
	if err := w.watcher.Add(w.loader.agentsPath); err != nil {
		return fmt.Errorf("failed to watch agents directory: %w", err)
	}

	log.Printf("[agent/watcher] Watching %s for changes", w.loader.agentsPath)

	// Track recent events for debouncing
	recentEvents := make(map[string]time.Time)
	eventMutex := make(chan struct{}, 1)
	eventMutex <- struct{}{} // Initialize mutex

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[agent/watcher] Stopping watcher")
				return

			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				// Debounce: ignore events for the same file within debounce period
				<-eventMutex
				lastEvent, exists := recentEvents[event.Name]
				if exists && time.Since(lastEvent) < w.debounce {
					eventMutex <- struct{}{}
					continue
				}
				recentEvents[event.Name] = time.Now()
				eventMutex <- struct{}{}

				// Handle the event
				if err := w.handleEvent(event); err != nil {
					log.Printf("[agent/watcher] Error handling event %s: %v", event.Name, err)
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[agent/watcher] Error: %v", err)
			}
		}
	}()

	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	if !w.enabled || w.watcher == nil {
		return nil
	}
	return w.watcher.Close()
}

// handleEvent processes file system events (create, write, remove)
func (w *Watcher) handleEvent(event fsnotify.Event) error {
	filename := filepath.Base(event.Name)

	// Ignore temporary files
	if w.isTemporaryFile(filename) {
		return nil
	}

	// Only process YAML files
	if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
		return nil
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		return w.handleCreate(filename)

	case event.Op&fsnotify.Write == fsnotify.Write:
		return w.handleWrite(filename)

	case event.Op&fsnotify.Remove == fsnotify.Remove:
		return w.handleRemove(filename)

	case event.Op&fsnotify.Rename == fsnotify.Rename:
		// Treat rename as remove (file moved away)
		return w.handleRemove(filename)
	}

	return nil
}

// handleCreate loads a newly created agent file
func (w *Watcher) handleCreate(filename string) error {
	log.Printf("[agent/watcher] CREATE: %s", filename)

	def, err := w.loader.LoadOne(filename)
	if err != nil {
		return fmt.Errorf("failed to load new agent: %w", err)
	}

	if err := w.pool.Reload(def); err != nil {
		return fmt.Errorf("failed to add agent to pool: %w", err)
	}

	log.Printf("[agent/watcher] Agent %q added successfully", def.Name)
	return nil
}

// handleWrite reloads a modified agent file
func (w *Watcher) handleWrite(filename string) error {
	log.Printf("[agent/watcher] WRITE: %s", filename)

	def, err := w.loader.LoadOne(filename)
	if err != nil {
		return fmt.Errorf("failed to reload agent: %w", err)
	}

	if err := w.pool.Reload(def); err != nil {
		return fmt.Errorf("failed to reload agent in pool: %w", err)
	}

	log.Printf("[agent/watcher] Agent %q reloaded successfully", def.Name)
	return nil
}

// handleRemove removes an agent from the pool
func (w *Watcher) handleRemove(filename string) error {
	log.Printf("[agent/watcher] REMOVE: %s", filename)

	// Extract agent name from filename (remove extension)
	agentName := strings.TrimSuffix(filename, filepath.Ext(filename))

	if err := w.pool.Remove(agentName); err != nil {
		// Don't error if agent doesn't exist (might have been renamed)
		log.Printf("[agent/watcher] Agent %q not in pool (may have been renamed)", agentName)
		return nil
	}

	log.Printf("[agent/watcher] Agent %q removed successfully", agentName)
	return nil
}

// isTemporaryFile checks if a filename is a temporary file
func (w *Watcher) isTemporaryFile(filename string) bool {
	// First check if it's a YAML file (these are never temporary)
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		return false
	}

	// Ignore vim swap files, emacs backup files, and other temp files
	tempPatterns := []string{
		".swp", ".swo", ".swn", // Vim
		"~",             // Emacs backup
		".tmp", ".temp", // Generic temp
		"#",    // Emacs auto-save
		".bak", // Backup files
	}

	for _, pattern := range tempPatterns {
		if strings.HasSuffix(filename, pattern) || strings.HasPrefix(filename, pattern) {
			return true
		}
	}

	// Ignore hidden files
	if strings.HasPrefix(filename, ".") {
		return true
	}

	return false
}
