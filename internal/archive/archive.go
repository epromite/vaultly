package archive

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/epromite/vaultly/internal/config"
	"github.com/epromite/vaultly/internal/logger"
	"github.com/epromite/vaultly/internal/mover"
	"github.com/fatih/color"
)

// Handler manages archive and program file operations.
type Handler struct {
	cfg *config.Config
	log *logger.Logger
}

// New creates a new archive Handler.
func New(cfg *config.Config, log *logger.Logger) *Handler {
	return &Handler{cfg: cfg, log: log}
}

// BatchHandleArchives prompts once for all archive files, then processes them.
func (h *Handler) BatchHandleArchives(files []string, autoArchive bool) {
	if len(files) == 0 {
		return
	}

	if autoArchive {
		for _, f := range files {
			h.moveToFolder(f, h.cfg.Paths.Archives, logger.ActionArchived)
		}
		return
	}

	// Show batch summary
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("  \U0001F4E6 Found %d archive files (.zip, .rar, .7z, .tar, .gz, etc.)\n", len(files))

	// Show first few files as preview
	showCount := len(files)
	if showCount > 5 {
		showCount = 5
	}
	for i := 0; i < showCount; i++ {
		color.New(color.FgHiBlack).Printf("     \u2022 %s\n", filepath.Base(files[i]))
	}
	if len(files) > 5 {
		color.New(color.FgHiBlack).Printf("     ... and %d more\n", len(files)-5)
	}

	fmt.Println()
	fmt.Printf("  [1] Move all to %s\n", h.cfg.Paths.Archives)
	fmt.Println("  [2] Skip all")
	fmt.Print("  Choose (1/2): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		for _, f := range files {
			h.moveToFolder(f, h.cfg.Paths.Archives, logger.ActionArchived)
		}
	default:
		for _, f := range files {
			h.log.Log(logger.ActionSkipped, filepath.Base(f), "user chose skip (batch)")
		}
	}
}

// BatchHandlePrograms prompts once for all program/installer files, then processes them.
func (h *Handler) BatchHandlePrograms(files []string, autoArchive bool) {
	if len(files) == 0 {
		return
	}

	if autoArchive {
		for _, f := range files {
			h.moveToFolder(f, h.cfg.Paths.Programs, logger.ActionMoved)
		}
		return
	}

	// Show batch summary
	fmt.Println()
	color.New(color.FgMagenta, color.Bold).Printf("  \U0001F4BB Found %d program/installer files (.exe, .msi, .dll, etc.)\n", len(files))

	// Show first few files as preview
	showCount := len(files)
	if showCount > 5 {
		showCount = 5
	}
	for i := 0; i < showCount; i++ {
		color.New(color.FgHiBlack).Printf("     \u2022 %s\n", filepath.Base(files[i]))
	}
	if len(files) > 5 {
		color.New(color.FgHiBlack).Printf("     ... and %d more\n", len(files)-5)
	}

	fmt.Println()
	fmt.Printf("  [1] Move all to %s\n", h.cfg.Paths.Programs)
	fmt.Println("  [2] Skip all")
	fmt.Print("  Choose (1/2): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		for _, f := range files {
			h.moveToFolder(f, h.cfg.Paths.Programs, logger.ActionMoved)
		}
	default:
		for _, f := range files {
			h.log.Log(logger.ActionSkipped, filepath.Base(f), "user chose skip (batch)")
		}
	}
}

// moveToFolder moves a file to the given destination folder.
func (h *Handler) moveToFolder(filePath, destDir string, action logger.Action) {
	filename := filepath.Base(filePath)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		h.log.Log(logger.ActionError, filename, fmt.Sprintf("cannot create dir: %v", err))
		return
	}

	finalName := mover.GenerateUniqueName(destDir, filename)
	destPath := filepath.Join(destDir, finalName)

	// Try rename, fallback to copy+delete for cross-drive
	if err := os.Rename(filePath, destPath); err != nil {
		src, errOpen := os.Open(filePath)
		if errOpen != nil {
			h.log.Log(logger.ActionError, filename, errOpen.Error())
			return
		}

		dst, errCreate := os.Create(destPath)
		if errCreate != nil {
			src.Close()
			h.log.Log(logger.ActionError, filename, errCreate.Error())
			return
		}

		if _, errCopy := io.Copy(dst, src); errCopy != nil {
			src.Close()
			dst.Close()
			h.log.Log(logger.ActionError, filename, errCopy.Error())
			return
		}
		src.Close()
		dst.Close()
		os.Remove(filePath)
	}

	if finalName != filename {
		h.log.Log(logger.ActionRenamed, filename, fmt.Sprintf("%s (duplicate)", destPath))
	} else {
		h.log.Log(action, filename, destPath)
	}
}

// ExtractArchive extracts an archive file in-place (for future use).
func (h *Handler) ExtractArchive(filePath string) {
	filename := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filename))
	baseDir := filepath.Dir(filePath)
	extractDir := filepath.Join(baseDir, strings.TrimSuffix(filename, ext))

	if err := os.MkdirAll(extractDir, 0755); err != nil {
		h.log.Log(logger.ActionError, filename, fmt.Sprintf("cannot create extract dir: %v", err))
		return
	}

	var err error
	switch ext {
	case ".zip":
		err = extractZip(filePath, extractDir)
	case ".gz":
		err = extractTarGz(filePath, extractDir)
	case ".tar":
		err = extractTar(filePath, extractDir)
	default:
		h.log.Log(logger.ActionError, filename, fmt.Sprintf("extraction not supported for %s", ext))
		return
	}

	if err != nil {
		h.log.Log(logger.ActionError, filename, fmt.Sprintf("extraction failed: %v", err))
		return
	}

	color.New(color.FgGreen).Printf("  \u2705 Extracted %s \u2192 %s\n", filename, extractDir)
}

// extractZip extracts a .zip file.
func extractZip(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		targetPath := filepath.Join(destDir, f.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.Create(targetPath)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarGz extracts a .tar.gz file.
func extractTarGz(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	return extractTarReader(tar.NewReader(gzr), destDir)
}

// extractTar extracts a .tar file.
func extractTar(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return extractTarReader(tar.NewReader(f), destDir)
}

// extractTarReader processes a tar reader and writes files.
func extractTarReader(tr *tar.Reader, destDir string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destDir, header.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(targetPath, 0755)
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
