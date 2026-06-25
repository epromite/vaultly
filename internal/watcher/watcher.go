package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/epromite/vaultly/internal/classifier"
	"github.com/epromite/vaultly/internal/config"
	"github.com/epromite/vaultly/internal/logger"
	"github.com/epromite/vaultly/internal/mover"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

// Watcher watches the source directory and auto-moves new files.
type Watcher struct {
	cfg            *config.Config
	cls            *classifier.Classifier
	log            *logger.Logger
	autoArchive    bool
	archiveHandler func(string)
	debounce       map[string]*time.Timer
	mu             sync.Mutex
}

// New creates a new Watcher.
func New(cfg *config.Config, cls *classifier.Classifier, log *logger.Logger, autoArchive bool, archiveHandler func(string)) *Watcher {
	return &Watcher{
		cfg:            cfg,
		cls:            cls,
		log:            log,
		autoArchive:    autoArchive,
		archiveHandler: archiveHandler,
		debounce:       make(map[string]*time.Timer),
	}
}

// Watch starts watching the source directory for new files.
// It blocks until the stop channel is closed.
func (w *Watcher) Watch(stop <-chan os.Signal) error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("cannot create watcher: %w", err)
	}
	defer fsw.Close()

	if err := fsw.Add(w.cfg.Paths.Source); err != nil {
		return fmt.Errorf("cannot watch directory: %w", err)
	}

	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("  \U0001F440 Watching: %s\n", w.cfg.Paths.Source)
	fmt.Println("  Press Ctrl+C to stop.")
	fmt.Println("  ================================")

	for {
		select {
		case event, ok := <-fsw.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				w.scheduleMove(event.Name)
			}

		case err, ok := <-fsw.Errors:
			if !ok {
				return nil
			}
			w.log.Log(logger.ActionError, "watcher", err.Error())

		case <-stop:
			fmt.Println()
			color.New(color.FgYellow).Println("  \u23F9\uFE0F  Watch mode stopped.")
			return nil
		}
	}
}

// scheduleMove debounces file creation events (wait 2s for downloads to complete).
func (w *Watcher) scheduleMove(filePath string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Cancel any pending timer for this file
	if timer, exists := w.debounce[filePath]; exists {
		timer.Stop()
	}

	w.debounce[filePath] = time.AfterFunc(2*time.Second, func() {
		w.processFile(filePath)

		w.mu.Lock()
		delete(w.debounce, filePath)
		w.mu.Unlock()
	})
}

// processFile moves a single new file.
func (w *Watcher) processFile(filePath string) {
	// Verify file still exists (it might have been moved already)
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return
	}

	filename := filepath.Base(filePath)
	m := mover.New(w.cfg, w.cls, w.log, false)

	if m.ShouldIgnore(filename) {
		return
	}

	cat := m.MoveFile(filePath)

	if cat == classifier.Archives && w.archiveHandler != nil {
		w.archiveHandler(filePath)
	}
}
