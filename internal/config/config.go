package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/epromite/vaultly/internal/winpath"
	"github.com/fatih/color"
)

// Config represents the application configuration.
type Config struct {
	Version     string       `json:"version"`
	Paths       Paths        `json:"paths"`
	Options     Options      `json:"options"`
	CustomRules []CustomRule `json:"customRules"`
}

// Paths holds all folder paths.
type Paths struct {
	Source    string `json:"source"`
	Pictures  string `json:"pictures"`
	Music     string `json:"music"`
	Videos    string `json:"videos"`
	Documents string `json:"documents"`
	Archives  string `json:"archives"`
	Programs  string `json:"programs"`
}

// Options holds configuration options.
type Options struct {
	AutoArchive       bool     `json:"autoArchive"`
	WatchMode         bool     `json:"watchMode"`
	LogFile           string   `json:"logFile"`
	DuplicateHandling string   `json:"duplicateHandling"`
	IgnoreFiles       []string `json:"ignoreFiles"`
}

// CustomRule defines a custom extension-to-destination mapping.
type CustomRule struct {
	Extension   string `json:"extension"`
	Destination string `json:"destination"`
}

// ConfigPath returns the path to the config file in user home.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vaultlyrc.json")
}

// Exists checks if the config file exists.
func Exists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

// Load reads and parses the config file.
func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	return &cfg, nil
}

// Save writes the config as formatted JSON.
func Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0644)
}

// OpenInNotepad opens the config file in Notepad for editing.
func OpenInNotepad() error {
	if !Exists() {
		return fmt.Errorf("config file does not exist yet — run vaultly first to set up")
	}
	cmd := exec.Command("notepad.exe", ConfigPath())
	return cmd.Start()
}

// ValidatePaths checks that all configured paths are accessible.
// Returns warnings for any issues found.
func ValidatePaths(cfg *Config) []string {
	var warnings []string
	check := func(name, path string) {
		if path == "" {
			warnings = append(warnings, fmt.Sprintf("%s path is empty", name))
			return
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("%s path does not exist: %s", name, path))
		}
	}
	check("Source", cfg.Paths.Source)
	check("Pictures", cfg.Paths.Pictures)
	check("Music", cfg.Paths.Music)
	check("Videos", cfg.Paths.Videos)
	check("Documents", cfg.Paths.Documents)
	// Archives folder will be auto-created, skip validation
	return warnings
}

// RunSetupWizard runs the interactive first-run setup experience.
func RunSetupWizard() (*Config, error) {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)

	fmt.Println()
	bold.Println("  \U0001F512 Vaultly - First Run Setup")
	fmt.Println("  ================================")
	fmt.Println("  Welcome! Let's set up Vaultly.")
	fmt.Println()
	fmt.Println("  Vaultly will auto-detect your folders, but you can")
	fmt.Println("  customize each path. Press Enter to use the default.")
	fmt.Println()

	detected := winpath.DetectPaths()
	scanner := bufio.NewScanner(os.Stdin)

	promptPath := func(emoji, label, detectedPath string) string {
		cyan.Printf("  %s %s folder:\n", emoji, label)
		fmt.Printf("     Detected: %s\n", detectedPath)
		fmt.Print("     Custom path (or Enter to use detected): ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return detectedPath
		}
		return input
	}

	source := promptPath("\U0001F4C2", "Source (Downloads)", detected.Downloads)
	fmt.Println()
	pictures := promptPath("\U0001F4F7", "Pictures", detected.Pictures)
	fmt.Println()
	music := promptPath("\U0001F3B5", "Music", detected.Music)
	fmt.Println()
	videos := promptPath("\U0001F3AC", "Videos", detected.Videos)
	fmt.Println()
	documents := promptPath("\U0001F4C4", "Documents", detected.Documents)
	fmt.Println()

	defaultArchives := filepath.Join(source, "Archives")
	archives := promptPath("\U0001F4E6", "Archives (inside Downloads)", defaultArchives)
	fmt.Println()

	defaultPrograms := filepath.Join(source, "Programs")
	programs := promptPath("\U0001F4BB", "Programs (inside Downloads)", defaultPrograms)
	fmt.Println()

	home, _ := os.UserHomeDir()

	cfg := &Config{
		Version: "1",
		Paths: Paths{
			Source:    source,
			Pictures:  pictures,
			Music:     music,
			Videos:    videos,
			Documents: documents,
			Archives:  archives,
			Programs:  programs,
		},
		Options: Options{
			AutoArchive:       false,
			WatchMode:         false,
			LogFile:           filepath.Join(home, "vaultly.log"),
			DuplicateHandling: "rename",
			IgnoreFiles: []string{
				".gitkeep",
				".DS_Store",
				"desktop.ini",
				"thumbs.db",
				"*.tmp",
			},
		},
		CustomRules: []CustomRule{
			{Extension: ".ai", Destination: "pictures"},
			{Extension: ".fig", Destination: "pictures"},
			{Extension: ".psd", Destination: "pictures"},
			{Extension: ".epub", Destination: "documents"},
		},
	}

	if err := Save(cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("  ================================")
	green.Printf("  \u2705 Config saved to: %s\n", ConfigPath())
	fmt.Println()
	fmt.Println("  Tip: Run \"vaultly --config\" anytime to edit your paths.")
	fmt.Println("  ================================")
	fmt.Println()

	return cfg, nil
}
