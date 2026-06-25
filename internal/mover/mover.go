package mover

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/epromite/vaultly/internal/classifier"
	"github.com/epromite/vaultly/internal/config"
	"github.com/epromite/vaultly/internal/logger"
	"github.com/fatih/color"
)

// MoveRecord tracks a single file move for undo support.
type MoveRecord struct {
	OriginalPath string
	MovedPath    string
}

// ScanResult holds collected files that need batch handling.
type ScanResult struct {
	ArchiveFiles []string
	ProgramFiles []string
}

// Mover handles file move operations.
type Mover struct {
	cfg        *config.Config
	classifier *classifier.Classifier
	log        *logger.Logger
	dryRun     bool
	history    []MoveRecord
	unhandled  map[string][]string
	scanner    *bufio.Scanner
}

// New creates a new Mover.
func New(cfg *config.Config, cls *classifier.Classifier, log *logger.Logger, dryRun bool) *Mover {
	return &Mover{
		cfg:        cfg,
		classifier: cls,
		log:        log,
		dryRun:     dryRun,
		unhandled:  make(map[string][]string),
		scanner:    bufio.NewScanner(os.Stdin),
	}
}

// GetHistory returns the move history for undo.
func (m *Mover) GetHistory() []MoveRecord { return m.history }

// GetUnhandled returns unhandled files grouped by extension.
func (m *Mover) GetUnhandled() map[string][]string { return m.unhandled }

// ClearHistory clears the move history after undo.
func (m *Mover) ClearHistory() { m.history = nil }

// RecordMove adds a move record to history.
func (m *Mover) RecordMove(original, moved string) {
	m.history = append(m.history, MoveRecord{OriginalPath: original, MovedPath: moved})
}

// ResetUnhandled clears unhandled files for re-scan.
func (m *Mover) ResetUnhandled() { m.unhandled = make(map[string][]string) }

// ShouldIgnore checks if a file should be ignored based on the ignore list.
func (m *Mover) ShouldIgnore(filename string) bool {
	lower := strings.ToLower(filename)
	for _, pattern := range m.cfg.Options.IgnoreFiles {
		pattern = strings.ToLower(pattern)
		if strings.HasPrefix(pattern, "~$") {
			if strings.HasPrefix(lower, "~$") {
				return true
			}
			continue
		}
		if strings.HasPrefix(pattern, "*") {
			if strings.HasSuffix(lower, pattern[1:]) {
				return true
			}
			continue
		}
		if lower == pattern {
			return true
		}
	}
	return false
}

// isFileLocked tries to detect if a file is currently in use.
func isFileLocked(path string) bool {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return true
	}
	f.Close()
	return false
}

// GenerateUniqueName generates a unique filename to avoid overwriting duplicates.
func GenerateUniqueName(destDir, filename string) string {
	destPath := filepath.Join(destDir, filename)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		return filename
	}
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(filepath.Join(destDir, newName)); os.IsNotExist(err) {
			return newName
		}
	}
}

// MoveFileAcrossDrives moves a file, using copy+delete if os.Rename fails.
func MoveFileAcrossDrives(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyFile(src, dst); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("copied but failed to delete source: %w", err)
	}
	return nil
}

// fileHash computes SHA256 hash of a file for duplicate comparison.
func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// filesAreIdentical checks if two files are truly identical (size + SHA256).
func filesAreIdentical(path1, path2 string) (bool, error) {
	info1, err := os.Stat(path1)
	if err != nil {
		return false, err
	}
	info2, err := os.Stat(path2)
	if err != nil {
		return false, err
	}

	// Quick check: different sizes = not identical
	if info1.Size() != info2.Size() {
		return false, nil
	}

	// Deep check: compare SHA256 hashes
	hash1, err := fileHash(path1)
	if err != nil {
		return false, err
	}
	hash2, err := fileHash(path2)
	if err != nil {
		return false, err
	}

	return hash1 == hash2, nil
}

// promptDuplicate asks the user how to handle a duplicate file.
// Returns: "rename", "remove", or "skip"
func (m *Mover) promptDuplicate(srcPath, existingPath string) string {
	srcName := filepath.Base(srcPath)
	srcInfo, _ := os.Stat(srcPath)
	existInfo, _ := os.Stat(existingPath)

	identical, _ := filesAreIdentical(srcPath, existingPath)

	fmt.Println()
	color.New(color.FgYellow, color.Bold).Printf("  \U0001F504 Duplicate found: %s\n", srcName)
	fmt.Printf("     Source:      %s", srcPath)
	if srcInfo != nil {
		fmt.Printf(" (%s)", formatSize(srcInfo.Size()))
	}
	fmt.Println()
	fmt.Printf("     Existing:    %s", existingPath)
	if existInfo != nil {
		fmt.Printf(" (%s)", formatSize(existInfo.Size()))
	}
	fmt.Println()

	if identical {
		color.New(color.FgGreen).Println("     \u2714 Files are IDENTICAL (same size + SHA256)")
	} else {
		color.New(color.FgYellow).Println("     \u2716 Files are DIFFERENT (different content)")
	}

	fmt.Println()
	fmt.Println("  [1] Move & Rename (keep both)")
	if identical {
		fmt.Println("  [2] Remove Duplicate (delete source, keep existing)")
	} else {
		color.New(color.FgHiBlack).Println("  [2] Remove Duplicate (not recommended \u2014 files differ!)")
	}
	fmt.Println("  [3] Skip")
	fmt.Print("  Choose (1/2/3): ")

	m.scanner.Scan()
	choice := strings.TrimSpace(m.scanner.Text())

	switch choice {
	case "1":
		return "rename"
	case "2":
		if !identical {
			color.New(color.FgRed).Println("  \u26A0\uFE0F  Warning: Files have different content!")
			fmt.Print("  Are you sure you want to delete the source file? (y/n): ")
			m.scanner.Scan()
			if !strings.EqualFold(strings.TrimSpace(m.scanner.Text()), "y") {
				return "skip"
			}
		}
		return "remove"
	default:
		return "skip"
	}
}

// formatSize formats bytes to human-readable size.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// MoveFile moves a single file to its classified destination.
func (m *Mover) MoveFile(filePath string) classifier.Category {
	filename := filepath.Base(filePath)

	if m.ShouldIgnore(filename) {
		return classifier.Unknown
	}

	cat := m.classifier.Classify(filename)

	if cat == classifier.Unknown {
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == "" {
			ext = "(no extension)"
		}
		m.unhandled[ext] = append(m.unhandled[ext], filePath)
		m.log.Log(logger.ActionUnknown, filename, "no rule defined")
		return cat
	}

	// Archives and Programs are handled separately by batch handlers
	if cat == classifier.Archives || cat == classifier.Programs {
		return cat
	}

	destDir := m.classifier.GetDestination(cat, m.cfg.Paths)
	if destDir == "" {
		m.log.Log(logger.ActionError, filename, "no destination configured")
		return cat
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		m.log.Log(logger.ActionError, filename, fmt.Sprintf("cannot create directory: %v", err))
		return cat
	}

	if m.dryRun {
		return cat
	}

	if isFileLocked(filePath) {
		m.log.Log(logger.ActionSkipped, filename, "file is currently in use")
		return cat
	}

	// Check if duplicate exists at destination
	existingPath := filepath.Join(destDir, filename)
	if _, err := os.Stat(existingPath); err == nil {
		// Duplicate found — prompt user
		action := m.promptDuplicate(filePath, existingPath)

		switch action {
		case "rename":
			finalName := GenerateUniqueName(destDir, filename)
			destPath := filepath.Join(destDir, finalName)
			if err := MoveFileAcrossDrives(filePath, destPath); err != nil {
				m.log.Log(logger.ActionError, filename, err.Error())
				return cat
			}
			m.history = append(m.history, MoveRecord{OriginalPath: filePath, MovedPath: destPath})
			m.log.Log(logger.ActionRenamed, filename, fmt.Sprintf("%s (duplicate)", destPath))

		case "remove":
			if err := os.Remove(filePath); err != nil {
				m.log.Log(logger.ActionError, filename, fmt.Sprintf("cannot remove duplicate: %v", err))
				return cat
			}
			m.log.Log(logger.ActionSkipped, filename, "duplicate removed (source deleted)")

		case "skip":
			m.log.Log(logger.ActionSkipped, filename, "duplicate skipped by user")
		}

		return cat
	}

	// No duplicate — move directly
	destPath := filepath.Join(destDir, filename)
	if err := MoveFileAcrossDrives(filePath, destPath); err != nil {
		m.log.Log(logger.ActionError, filename, err.Error())
		return cat
	}

	m.history = append(m.history, MoveRecord{OriginalPath: filePath, MovedPath: destPath})
	m.log.Log(logger.ActionMoved, filename, destPath)
	return cat
}

// skipDirs returns a set of directories to skip during recursive scan.
func (m *Mover) skipDirs() map[string]bool {
	skip := make(map[string]bool)
	for _, p := range []string{m.cfg.Paths.Archives, m.cfg.Paths.Programs} {
		if p != "" {
			skip[filepath.Clean(strings.ToLower(p))] = true
		}
	}
	return skip
}

// ScanAndMove recursively scans the source directory and moves files.
func (m *Mover) ScanAndMove() (*ScanResult, error) {
	result := &ScanResult{}
	skip := m.skipDirs()

	err := filepath.WalkDir(m.cfg.Paths.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path == m.cfg.Paths.Source {
				return nil
			}
			if skip[filepath.Clean(strings.ToLower(path))] {
				return filepath.SkipDir
			}
			return nil
		}

		cat := m.MoveFile(path)
		switch cat {
		case classifier.Archives:
			result.ArchiveFiles = append(result.ArchiveFiles, path)
		case classifier.Programs:
			result.ProgramFiles = append(result.ProgramFiles, path)
		}
		return nil
	})

	return result, err
}

// DryRun performs a preview scan without moving any files.
func (m *Mover) DryRun() error {
	skip := m.skipDirs()

	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("  [DRY RUN] Vaultly - Preview Mode")
	fmt.Println("  ================================")

	willMove, willPrompt, willSkip := 0, 0, 0

	filepath.WalkDir(m.cfg.Paths.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path == m.cfg.Paths.Source {
				return nil
			}
			if skip[filepath.Clean(strings.ToLower(path))] {
				return filepath.SkipDir
			}
			return nil
		}

		filename := d.Name()
		if m.ShouldIgnore(filename) {
			return nil
		}

		cat := m.classifier.Classify(filename)
		destDir := m.classifier.GetDestination(cat, m.cfg.Paths)

		rel, _ := filepath.Rel(m.cfg.Paths.Source, path)
		display := rel

		switch cat {
		case classifier.Archives:
			color.New(color.FgYellow).Printf("  ? %-30s \u2192 Archives\\ (will prompt)\n", display)
			willPrompt++
		case classifier.Programs:
			color.New(color.FgMagenta).Printf("  ? %-30s \u2192 Programs\\ (will prompt)\n", display)
			willPrompt++
		case classifier.Unknown:
			color.New(color.FgHiBlack).Printf("  - %-30s \u2192 SKIP (unknown type)\n", display)
			willSkip++
		default:
			color.New(color.FgGreen).Printf("  \u2713 %-30s \u2192 %s\n", display, destDir)
			willMove++
		}
		return nil
	})

	fmt.Println("  --------------------------------")
	fmt.Printf("  Total: %d will move, %d will prompt, %d will skip\n", willMove, willPrompt, willSkip)
	fmt.Println("  Run without --dry-run to apply changes.")
	fmt.Println()
	return nil
}

// copyFile copies a file preserving modification time.
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("cannot create destination: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	info, err := os.Stat(src)
	if err == nil {
		os.Chtimes(dst, info.ModTime(), info.ModTime())
	}
	return nil
}
