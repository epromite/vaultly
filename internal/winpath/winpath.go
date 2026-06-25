package winpath

import (
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shell32                  = windows.NewLazySystemDLL("shell32.dll")
	ole32                    = windows.NewLazySystemDLL("ole32.dll")
	procSHGetKnownFolderPath = shell32.NewProc("SHGetKnownFolderPath")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
)

// Known Folder GUIDs from Windows API
var (
	folderDownloads = windows.GUID{
		Data1: 0x374DE290, Data2: 0x123F, Data3: 0x4565,
		Data4: [8]byte{0x91, 0x64, 0x39, 0xC4, 0x92, 0x5E, 0x46, 0x7B},
	}
	folderPictures = windows.GUID{
		Data1: 0x33E28130, Data2: 0x4E1E, Data3: 0x4676,
		Data4: [8]byte{0x83, 0x5A, 0x98, 0x39, 0x5C, 0x3B, 0xC3, 0xBB},
	}
	folderMusic = windows.GUID{
		Data1: 0x4BD8D571, Data2: 0x6D19, Data3: 0x48D3,
		Data4: [8]byte{0xBE, 0x97, 0x42, 0x22, 0x20, 0x08, 0x0E, 0x43},
	}
	folderVideos = windows.GUID{
		Data1: 0x18989B1D, Data2: 0x99B5, Data3: 0x455B,
		Data4: [8]byte{0x84, 0x1C, 0xAB, 0x7C, 0x74, 0xE4, 0xDD, 0xFC},
	}
	folderDocuments = windows.GUID{
		Data1: 0xFDD39AD0, Data2: 0x238F, Data3: 0x46AF,
		Data4: [8]byte{0xAD, 0xB4, 0x6C, 0x85, 0x48, 0x03, 0x69, 0xC7},
	}
)

// getKnownFolderPath retrieves the path of a known folder using Windows API.
func getKnownFolderPath(folderID *windows.GUID) (string, error) {
	var pathPtr *uint16
	hr, _, _ := procSHGetKnownFolderPath.Call(
		uintptr(unsafe.Pointer(folderID)),
		0,
		0,
		uintptr(unsafe.Pointer(&pathPtr)),
	)
	if hr != 0 {
		return "", windows.Errno(hr)
	}
	defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(pathPtr)))
	return windows.UTF16PtrToString(pathPtr), nil
}

// DetectedPaths holds all auto-detected Windows folder paths.
type DetectedPaths struct {
	Downloads string
	Pictures  string
	Music     string
	Videos    string
	Documents string
}

// getOrFallback tries Windows API first, then falls back to home + subfolder.
func getOrFallback(guid *windows.GUID, fallbackSub string) string {
	if p, err := getKnownFolderPath(guid); err == nil && p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, fallbackSub)
}

// DetectPaths auto-detects all known folder paths using SHGetKnownFolderPath.
func DetectPaths() DetectedPaths {
	return DetectedPaths{
		Downloads: getOrFallback(&folderDownloads, "Downloads"),
		Pictures:  getOrFallback(&folderPictures, "Pictures"),
		Music:     getOrFallback(&folderMusic, "Music"),
		Videos:    getOrFallback(&folderVideos, "Videos"),
		Documents: getOrFallback(&folderDocuments, "Documents"),
	}
}
