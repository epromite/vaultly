package classifier

import (
	"path/filepath"
	"strings"

	"github.com/epromite/vaultly/internal/config"
)

// Category represents a file category.
type Category string

const (
	Pictures  Category = "pictures"
	Music     Category = "music"
	Videos    Category = "videos"
	Documents Category = "documents"
	Archives  Category = "archives"
	Programs  Category = "programs"
	Unknown   Category = "unknown"
)

// Default extension-to-category mappings.
var defaultMappings = map[string]Category{
	// Pictures
	".jpg": Pictures, ".jpeg": Pictures, ".jfif": Pictures, ".png": Pictures,
	".gif": Pictures, ".bmp": Pictures, ".svg": Pictures, ".webp": Pictures,
	".ico": Pictures, ".tiff": Pictures, ".tif": Pictures, ".avif": Pictures,
	".heic": Pictures, ".heif": Pictures, ".raw": Pictures, ".cr2": Pictures,
	".nef": Pictures,
	// Music
	".mp3": Music, ".wav": Music, ".flac": Music, ".aac": Music,
	".ogg": Music, ".wma": Music, ".m4a": Music,
	// Videos
	".mp4": Videos, ".mkv": Videos, ".avi": Videos, ".mov": Videos,
	".wmv": Videos, ".flv": Videos, ".webm": Videos,
	// Documents
	".pdf": Documents, ".doc": Documents, ".docx": Documents,
	".xls": Documents, ".xlsx": Documents, ".ppt": Documents,
	".pptx": Documents, ".txt": Documents, ".csv": Documents,
	".odt": Documents,
	// Archives
	".zip": Archives, ".rar": Archives, ".7z": Archives,
	".tar": Archives, ".gz": Archives, ".bz2": Archives, ".xz": Archives,
	// Programs / Installers
	".exe": Programs, ".msi": Programs, ".msix": Programs, ".appx": Programs,
	".bat": Programs, ".cmd": Programs, ".ps1": Programs, ".dll": Programs,
}

// Classifier classifies files by their extension.
type Classifier struct {
	mappings map[string]Category
}

// New creates a Classifier with custom rules applied on top of defaults.
func New(customRules []config.CustomRule) *Classifier {
	c := &Classifier{mappings: make(map[string]Category)}

	for ext, cat := range defaultMappings {
		c.mappings[ext] = cat
	}

	// Custom rules override defaults
	for _, rule := range customRules {
		ext := strings.ToLower(rule.Extension)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		c.mappings[ext] = Category(rule.Destination)
	}

	return c
}

// Classify returns the category for a given filename.
func (c *Classifier) Classify(filename string) Category {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return Unknown
	}
	if cat, ok := c.mappings[ext]; ok {
		return cat
	}
	return Unknown
}

// GetDestination returns the destination path for a given category.
func (c *Classifier) GetDestination(cat Category, paths config.Paths) string {
	switch cat {
	case Pictures:
		return paths.Pictures
	case Music:
		return paths.Music
	case Videos:
		return paths.Videos
	case Documents:
		return paths.Documents
	case Archives:
		return paths.Archives
	case Programs:
		return paths.Programs
	default:
		return ""
	}
}
