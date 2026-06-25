package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Action represents a log action type.
type Action string

const (
	ActionMoved    Action = "MOVED"
	ActionRenamed  Action = "RENAMED"
	ActionSkipped  Action = "SKIPPED"
	ActionUnknown  Action = "UNKNOWN"
	ActionArchived Action = "ARCHIVED"
	ActionError    Action = "ERROR"
)

// Stats tracks operation statistics.
type Stats struct {
	Moved    int
	Renamed  int
	Skipped  int
	Errors   int
	Unknown  int
	Archived int
}

// Logger handles dual output to log file and colored terminal.
type Logger struct {
	logFile *os.File
	logPath string
	mu      sync.Mutex
	stats   Stats
}

// New creates a new Logger that writes to the given log file path.
func New(logPath string) (*Logger, error) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open log file: %w", err)
	}
	return &Logger{logFile: f, logPath: logPath}, nil
}

// Close closes the log file.
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

// LogPath returns the path to the log file.
func (l *Logger) LogPath() string {
	return l.logPath
}

// GetStats returns current statistics.
func (l *Logger) GetStats() Stats {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stats
}

// Log writes a log entry to both file and terminal.
func (l *Logger) Log(action Action, filename, detail string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch action {
	case ActionMoved:
		l.stats.Moved++
	case ActionRenamed:
		l.stats.Renamed++
	case ActionSkipped:
		l.stats.Skipped++
	case ActionError:
		l.stats.Errors++
	case ActionUnknown:
		l.stats.Unknown++
	case ActionArchived:
		l.stats.Archived++
	}

	// Write to log file
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %-9s %-20s \u2192 %s\n", timestamp, action, filename, detail)
	if l.logFile != nil {
		l.logFile.WriteString(logLine)
	}

	// Write to terminal with colors and emoji
	padded := fmt.Sprintf("%-20s", filename)
	switch action {
	case ActionMoved:
		color.New(color.FgGreen).Printf("  \u2705 Moved    %s \u2192 %s\n", padded, detail)
	case ActionRenamed:
		color.New(color.FgYellow).Printf("  \u26A0\uFE0F  Renamed  %s \u2192 %s\n", padded, detail)
	case ActionSkipped:
		color.New(color.FgYellow).Printf("  \u23ED\uFE0F  Skipped  %s \u2192 %s\n", padded, detail)
	case ActionUnknown:
		color.New(color.FgHiBlack).Printf("  \u23ED\uFE0F  Skipped  %s \u2192 %s\n", padded, detail)
	case ActionArchived:
		color.New(color.FgCyan).Printf("  \U0001F4E6 Archive  %s \u2192 %s\n", padded, detail)
	case ActionError:
		color.New(color.FgRed).Printf("  \u274C Error    %s \u2192 %s\n", padded, detail)
	}
}

// PrintSummary prints the final summary to terminal.
func (l *Logger) PrintSummary() {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Println("  ================================")

	total := l.stats.Moved + l.stats.Archived
	parts := []string{}
	if total > 0 {
		parts = append(parts, fmt.Sprintf("%d moved", total))
	}
	if l.stats.Renamed > 0 {
		parts = append(parts, fmt.Sprintf("%d renamed", l.stats.Renamed))
	}
	skipped := l.stats.Skipped + l.stats.Unknown
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}
	if l.stats.Errors > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", l.stats.Errors))
	}

	if len(parts) == 0 {
		color.New(color.FgGreen).Println("  \u2705 No files to organize!")
	} else {
		summary := ""
		for i, p := range parts {
			if i > 0 {
				summary += ", "
			}
			summary += p
		}
		color.New(color.FgGreen).Printf("  \u2705 Done! %s\n", summary)
	}
	fmt.Printf("  \U0001F4DD Log saved to: %s\n", l.logPath)
}
