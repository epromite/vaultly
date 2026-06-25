package menu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/epromite/vaultly/internal/archive"
	"github.com/epromite/vaultly/internal/config"
	"github.com/epromite/vaultly/internal/logger"
	"github.com/epromite/vaultly/internal/mover"
	"github.com/fatih/color"
)

// Menu provides an interactive post-sort menu.
type Menu struct {
	cfg        *config.Config
	log        *logger.Logger
	mv         *mover.Mover
	archiveHdl *archive.Handler
	scanner    *bufio.Scanner
}

// New creates a new interactive Menu.
func New(cfg *config.Config, log *logger.Logger, mv *mover.Mover, archiveHdl *archive.Handler) *Menu {
	return &Menu{
		cfg:        cfg,
		log:        log,
		mv:         mv,
		archiveHdl: archiveHdl,
		scanner:    bufio.NewScanner(os.Stdin),
	}
}

// Run starts the interactive menu loop.
func (m *Menu) Run() {
	for {
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Println("  \U0001F4CB What would you like to do?")
		fmt.Println("  ================================")
		fmt.Println("  [0] Exit")
		fmt.Println("  [1] Custom Move (assign unknown files)")
		fmt.Println("  [2] Undo Last Sort")
		fmt.Println("  [3] View Unhandled Files")
		fmt.Println("  [4] Re-scan & Sort Again")
		fmt.Println()
		fmt.Print("  Choose: ")

		m.scanner.Scan()
		choice := strings.TrimSpace(m.scanner.Text())

		switch choice {
		case "0":
			color.New(color.FgGreen).Println("\n  \U0001F44B Goodbye!")
			return
		case "1":
			m.customMove()
		case "2":
			m.undoMoves()
		case "3":
			m.viewUnhandled()
		case "4":
			m.rescan()
		default:
			color.New(color.FgYellow).Println("  Invalid choice, try again.")
		}
	}
}

// viewUnhandled shows unhandled files grouped by extension with counts.
func (m *Menu) viewUnhandled() {
	unhandled := m.mv.GetUnhandled()
	if len(unhandled) == 0 {
		color.New(color.FgGreen).Println("\n  \u2705 No unhandled files!")
		return
	}

	type group struct {
		ext   string
		count int
	}
	var groups []group
	for ext, files := range unhandled {
		groups = append(groups, group{ext, len(files)})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].count > groups[j].count })

	fmt.Println()
	color.New(color.FgYellow, color.Bold).Println("  \U0001F4CB Unhandled Files by Extension")
	fmt.Println("  ================================")

	total := 0
	for _, g := range groups {
		fmt.Printf("  %5d Files | %s\n", g.count, g.ext)
		total += g.count
	}
	fmt.Println("  --------------------------------")
	fmt.Printf("  Total: %d unhandled files\n", total)
}

// customMove lets the user assign unknown file types to a destination.
func (m *Menu) customMove() {
	unhandled := m.mv.GetUnhandled()
	if len(unhandled) == 0 {
		color.New(color.FgGreen).Println("\n  \u2705 No unhandled files to move!")
		return
	}

	type group struct {
		ext   string
		files []string
	}
	var groups []group
	for ext, files := range unhandled {
		groups = append(groups, group{ext, files})
	}
	sort.Slice(groups, func(i, j int) bool { return len(groups[i].files) > len(groups[j].files) })

	// Show extension list
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("  \U0001F4C1 Custom Move \u2014 Select extension:")
	fmt.Println("  ================================")
	for i, g := range groups {
		fmt.Printf("  [%d] %s (%d files)\n", i+1, g.ext, len(g.files))
	}
	fmt.Println("  [0] Cancel")
	fmt.Print("  Choose: ")

	m.scanner.Scan()
	idx := 0
	fmt.Sscanf(strings.TrimSpace(m.scanner.Text()), "%d", &idx)
	if idx < 1 || idx > len(groups) {
		return
	}

	selected := groups[idx-1]

	// Show file preview
	fmt.Println()
	color.New(color.FgCyan).Printf("  Files with %s extension:\n", selected.ext)
	showMax := 8
	if len(selected.files) < showMax {
		showMax = len(selected.files)
	}
	for i := 0; i < showMax; i++ {
		color.New(color.FgHiBlack).Printf("     \u2022 %s\n", filepath.Base(selected.files[i]))
	}
	if len(selected.files) > showMax {
		color.New(color.FgHiBlack).Printf("     ... and %d more\n", len(selected.files)-showMax)
	}

	// Show destination options
	fmt.Println()
	color.New(color.FgCyan).Printf("  Move %d %s files to:\n", len(selected.files), selected.ext)
	fmt.Printf("  [1] \U0001F4F7 Pictures  (%s)\n", m.cfg.Paths.Pictures)
	fmt.Printf("  [2] \U0001F3B5 Music     (%s)\n", m.cfg.Paths.Music)
	fmt.Printf("  [3] \U0001F3AC Videos    (%s)\n", m.cfg.Paths.Videos)
	fmt.Printf("  [4] \U0001F4C4 Documents (%s)\n", m.cfg.Paths.Documents)
	fmt.Printf("  [5] \U0001F4E6 Archives  (%s)\n", m.cfg.Paths.Archives)
	fmt.Printf("  [6] \U0001F4BB Programs  (%s)\n", m.cfg.Paths.Programs)
	fmt.Println("  [7] \U0001F4C1 New folder (enter custom path)")
	fmt.Println("  [0] Cancel")
	fmt.Print("  Choose: ")

	m.scanner.Scan()
	destChoice := strings.TrimSpace(m.scanner.Text())

	var destDir string
	switch destChoice {
	case "1":
		destDir = m.cfg.Paths.Pictures
	case "2":
		destDir = m.cfg.Paths.Music
	case "3":
		destDir = m.cfg.Paths.Videos
	case "4":
		destDir = m.cfg.Paths.Documents
	case "5":
		destDir = m.cfg.Paths.Archives
	case "6":
		destDir = m.cfg.Paths.Programs
	case "7":
		fmt.Print("  Enter full folder path: ")
		m.scanner.Scan()
		destDir = strings.TrimSpace(m.scanner.Text())
		if destDir == "" {
			return
		}
	default:
		return
	}

	os.MkdirAll(destDir, 0755)
	moved := 0

	for _, f := range selected.files {
		filename := filepath.Base(f)

		// Check file still exists
		if _, err := os.Stat(f); os.IsNotExist(err) {
			continue
		}

		finalName := mover.GenerateUniqueName(destDir, filename)
		destPath := filepath.Join(destDir, finalName)

		if err := mover.MoveFileAcrossDrives(f, destPath); err != nil {
			m.log.Log(logger.ActionError, filename, err.Error())
			continue
		}

		m.mv.RecordMove(f, destPath)

		if finalName != filename {
			m.log.Log(logger.ActionRenamed, filename, destPath)
		} else {
			m.log.Log(logger.ActionMoved, filename, destPath)
		}
		moved++
	}

	// Remove from unhandled
	delete(unhandled, selected.ext)

	// Ask to save as permanent rule
	fmt.Println()
	catName := m.destToCategory(destDir)
	displayDest := catName
	if displayDest == destDir {
		displayDest = filepath.Base(destDir)
	}
	fmt.Printf("  Save \"%s \u2192 %s\" as permanent rule? (y/n): ", selected.ext, displayDest)
	m.scanner.Scan()
	if strings.EqualFold(strings.TrimSpace(m.scanner.Text()), "y") {
		m.cfg.CustomRules = append(m.cfg.CustomRules, config.CustomRule{
			Extension:   selected.ext,
			Destination: catName,
		})
		config.Save(m.cfg)
		color.New(color.FgGreen).Printf("  \u2705 Rule saved! %s files will auto-move in future runs.\n", selected.ext)
	}

	color.New(color.FgGreen).Printf("  \u2705 Moved %d %s files to %s\n", moved, selected.ext, destDir)
}

// undoMoves reverses all moves from this session.
func (m *Menu) undoMoves() {
	history := m.mv.GetHistory()
	if len(history) == 0 {
		color.New(color.FgYellow).Println("\n  \u26A0\uFE0F No moves to undo!")
		return
	}

	fmt.Println()
	color.New(color.FgYellow, color.Bold).Printf("  \u21A9\uFE0F  Undo %d file moves?\n", len(history))
	fmt.Println("  Files will be moved back to their original location.")
	fmt.Print("  Confirm (y/n): ")

	m.scanner.Scan()
	if !strings.EqualFold(strings.TrimSpace(m.scanner.Text()), "y") {
		return
	}

	fmt.Println("  ================================")
	undone := 0
	for i := len(history) - 1; i >= 0; i-- {
		rec := history[i]

		if _, err := os.Stat(rec.MovedPath); os.IsNotExist(err) {
			continue
		}

		os.MkdirAll(filepath.Dir(rec.OriginalPath), 0755)

		origName := filepath.Base(rec.OriginalPath)
		origDir := filepath.Dir(rec.OriginalPath)
		finalName := mover.GenerateUniqueName(origDir, origName)
		destPath := filepath.Join(origDir, finalName)

		if err := mover.MoveFileAcrossDrives(rec.MovedPath, destPath); err != nil {
			m.log.Log(logger.ActionError, filepath.Base(rec.MovedPath), fmt.Sprintf("undo failed: %v", err))
			continue
		}

		color.New(color.FgGreen).Printf("  \u21A9\uFE0F  %s \u2192 %s\n", filepath.Base(rec.MovedPath), destPath)
		undone++
	}

	m.mv.ClearHistory()
	fmt.Println("  ================================")
	color.New(color.FgGreen).Printf("  \u2705 Undone %d/%d moves.\n", undone, len(history))
}

// rescan performs a fresh scan and sort.
func (m *Menu) rescan() {
	fmt.Println()
	color.New(color.FgCyan).Printf("  \U0001F504 Re-scanning: %s\n", m.cfg.Paths.Source)
	fmt.Println("  ================================")

	m.mv.ResetUnhandled()

	result, err := m.mv.ScanAndMove()
	if err != nil {
		color.New(color.FgRed).Printf("  \u274C %v\n", err)
		return
	}

	m.archiveHdl.BatchHandleArchives(result.ArchiveFiles, m.cfg.Options.AutoArchive)
	m.archiveHdl.BatchHandlePrograms(result.ProgramFiles, m.cfg.Options.AutoArchive)

	m.log.PrintSummary()
}

// destToCategory maps a destination path back to a category name.
func (m *Menu) destToCategory(destDir string) string {
	switch destDir {
	case m.cfg.Paths.Pictures:
		return "pictures"
	case m.cfg.Paths.Music:
		return "music"
	case m.cfg.Paths.Videos:
		return "videos"
	case m.cfg.Paths.Documents:
		return "documents"
	case m.cfg.Paths.Archives:
		return "archives"
	case m.cfg.Paths.Programs:
		return "programs"
	default:
		return destDir
	}
}
