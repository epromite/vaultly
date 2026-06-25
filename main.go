package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/epromite/vaultly/internal/archive"
	"github.com/epromite/vaultly/internal/classifier"
	"github.com/epromite/vaultly/internal/config"
	"github.com/epromite/vaultly/internal/logger"
	"github.com/epromite/vaultly/internal/menu"
	"github.com/epromite/vaultly/internal/mover"
	"github.com/epromite/vaultly/internal/watcher"
	"github.com/fatih/color"
)

var version = "1.0.0"

func main() {
	// Define CLI flags
	watchMode := flag.Bool("watch", false, "Watch mode: auto-move new files in real-time")
	dryRun := flag.Bool("dry-run", false, "Preview mode: show what would be moved without moving")
	autoArchive := flag.Bool("auto-archive", false, "Skip archive/program prompt and auto-move")
	openConfig := flag.Bool("config", false, "Open config file in Notepad for editing")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		bold := color.New(color.Bold)
		bold.Printf("\n  \U0001F512 Vaultly v%s\n", version)
		fmt.Println("  Auto-organize your Downloads folder")
		fmt.Println()
		bold.Println("  Usage:")
		fmt.Println("    vaultly                  Run one-time sort")
		fmt.Println("    vaultly --watch           Watch mode (real-time)")
		fmt.Println("    vaultly --dry-run         Preview without moving")
		fmt.Println("    vaultly --auto-archive    Skip archive/program prompts")
		fmt.Println("    vaultly --config          Edit config in Notepad")
		fmt.Println("    vaultly --version         Show version")
		fmt.Println("    vaultly --help            Show this help")
		fmt.Println()
		bold.Println("  Config:")
		fmt.Printf("    Path: %s\n", config.ConfigPath())
		fmt.Println("    To change folder paths, run: vaultly --config")
		fmt.Println()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("Vaultly v%s\n", version)
		return
	}

	if *openConfig {
		if !config.Exists() {
			color.New(color.FgRed).Println("  \u274C Config file not found. Run vaultly first to set up.")
			os.Exit(1)
		}
		if err := config.OpenInNotepad(); err != nil {
			color.New(color.FgRed).Printf("  \u274C Cannot open config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  \U0001F4DD Opening config: %s\n", config.ConfigPath())
		return
	}

	// Print header
	color.New(color.FgCyan, color.Bold).Printf("\n  \U0001F512 Vaultly v%s\n", version)

	// Load or create config
	var cfg *config.Config
	if !config.Exists() {
		var err error
		cfg, err = config.RunSetupWizard()
		if err != nil {
			color.New(color.FgRed).Printf("  \u274C Setup failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		var err error
		cfg, err = config.Load()
		if err != nil {
			color.New(color.FgRed).Printf("  \u274C Cannot load config: %v\n", err)
			color.New(color.FgYellow).Println("  Run \"vaultly --config\" to fix your config file.")
			os.Exit(1)
		}
	}

	if *autoArchive {
		cfg.Options.AutoArchive = true
	}

	// Validate paths
	warnings := config.ValidatePaths(cfg)
	for _, w := range warnings {
		color.New(color.FgYellow).Printf("  \u26A0\uFE0F  %s\n", w)
	}

	if _, err := os.Stat(cfg.Paths.Source); os.IsNotExist(err) {
		color.New(color.FgRed).Printf("  \u274C Source folder does not exist: %s\n", cfg.Paths.Source)
		os.Exit(1)
	}

	// Initialize components
	cls := classifier.New(cfg.CustomRules)
	log, err := logger.New(cfg.Options.LogFile)
	if err != nil {
		color.New(color.FgRed).Printf("  \u274C Cannot create log file: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	archiveHdl := archive.New(cfg, log)

	// Handle --dry-run
	if *dryRun {
		mv := mover.New(cfg, cls, log, true)
		if err := mv.DryRun(); err != nil {
			color.New(color.FgRed).Printf("  \u274C %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle --watch mode
	if *watchMode {
		watchArchiveHandler := func(filePath string) {
			archiveHdl.BatchHandleArchives([]string{filePath}, cfg.Options.AutoArchive)
		}
		w := watcher.New(cfg, cls, log, cfg.Options.AutoArchive, watchArchiveHandler)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		if err := w.Watch(stop); err != nil {
			color.New(color.FgRed).Printf("  \u274C %v\n", err)
			os.Exit(1)
		}
		log.PrintSummary()
		return
	}

	// Default: one-time sort
	fmt.Printf("  \U0001F4C2 Scanning: %s (including subfolders)\n", cfg.Paths.Source)
	fmt.Println("  ================================")

	mv := mover.New(cfg, cls, log, false)
	result, err := mv.ScanAndMove()
	if err != nil {
		color.New(color.FgRed).Printf("  \u274C %v\n", err)
		os.Exit(1)
	}

	// Batch prompt for archives and programs
	archiveHdl.BatchHandleArchives(result.ArchiveFiles, cfg.Options.AutoArchive)
	archiveHdl.BatchHandlePrograms(result.ProgramFiles, cfg.Options.AutoArchive)

	log.PrintSummary()

	// Enter interactive menu (don't exit)
	m := menu.New(cfg, log, mv, archiveHdl)
	m.Run()
}
