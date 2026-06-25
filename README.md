# 🔒 Vaultly

**Auto-organize your Downloads folder in seconds.**

Vaultly is a lightweight CLI tool for Windows that automatically moves files from your Downloads folder to the right place — Pictures, Music, Videos, Documents — based on file type. No install needed, just download and run.

---

## ✨ Features

- 🚀 **One-time sort** — Run once, organize everything instantly
- 📁 **Recursive scanning** — Scans subfolders inside Downloads too
- 👁️ **Watch mode** — Runs in background, auto-moves new files in real-time
- 🔍 **Dry run** — Preview what will happen before moving anything
- 📦 **Archive handling** — Batch prompt for ZIP/RAR/7Z files (move or skip)
- 💻 **Program handling** — Batch prompt for EXE/MSI/DLL files
- 🔄 **Smart duplicate detection** — Compares file size + SHA256 hash, asks you what to do
- 🧙 **Setup wizard** — Auto-detects your Windows folders on first run
- 📂 **Custom paths** — Support for folders on any drive (D:\, E:\, etc.)
- ↩️ **Undo** — Revert all moves from the current session
- 🎯 **Custom Move** — Assign unknown file types to any folder
- 📋 **Unhandled view** — See unknown files grouped by extension
- 📝 **Logging** — Full history of every file moved
- 🎨 **Colored output** — Beautiful terminal output with emoji
- ⚡ **Single binary** — No dependencies, no install, just run

---

## 📥 Download & Install

1. Go to [**Releases**](https://github.com/epromite/vaultly/releases/latest)
2. Download `vaultly-windows-amd64.exe` (64-bit) or the version for your architecture
3. Rename to `vaultly.exe` (optional)
4. Place it anywhere (e.g., `C:\Tools\`) and [add to PATH](https://www.architectryan.com/2018/03/17/add-to-the-path-on-windows-10/)
5. Open terminal and run: `vaultly`

---

## 🧙 First Run

On first run, Vaultly will guide you through setup:

```
🔒 Vaultly - First Run Setup
================================
Welcome! Let's set up Vaultly.

📂 Source (Downloads) folder:
   Detected: C:\Users\John\Downloads
   Custom path (or Enter to use detected): _

📷 Pictures folder:
   Detected: C:\Users\John\Pictures
   Custom path (or Enter to use detected): D:\My Photos

🎵 Music folder:
   Detected: C:\Users\John\Music
   Custom path (or Enter to use detected): _

🎬 Videos folder:
   Detected: C:\Users\John\Videos
   Custom path (or Enter to use detected): _

📄 Documents folder:
   Detected: C:\Users\John\Documents
   Custom path (or Enter to use detected): _

📦 Archives (inside Downloads) folder:
   Default: C:\Users\John\Downloads\Archives
   Custom path (or Enter to use detected): _

💻 Programs (inside Downloads) folder:
   Default: C:\Users\John\Downloads\Programs
   Custom path (or Enter to use detected): _

================================
✅ Config saved to: C:\Users\John\.vaultlyrc.json
```

Vaultly auto-detects your Windows folder locations (even if you've moved them to another drive) using the Windows API (`SHGetKnownFolderPath`).

---

## 🚀 Usage

```bash
# Run one-time sort (default)
vaultly

# Watch mode — auto-move files in real-time
vaultly --watch

# Preview without moving anything
vaultly --dry-run

# Skip archive/program prompts, auto-move
vaultly --auto-archive

# Combine flags
vaultly --watch --auto-archive

# Edit config file in Notepad
vaultly --config

# Show version
vaultly --version

# Show help
vaultly --help
```

---

## 📋 Terminal Output

```
🔒 Vaultly v1.0.0
📂 Scanning: C:\Users\John\Downloads (including subfolders)
================================
✅ Moved    photo.jpg            → D:\My Photos\photo.jpg
✅ Moved    song.mp3             → D:\Music\song.mp3
✅ Moved    movie.mp4            → C:\Users\John\Videos\movie.mp4
✅ Moved    report.pdf           → C:\Users\John\Documents\report.pdf
⚠️  Renamed  photo.jpg            → D:\My Photos\photo_1.jpg
⏭️  Skipped  design.fig           → unknown type
================================
✅ Done! 4 moved, 1 renamed, 1 skipped
📝 Log saved to: C:\Users\John\vaultly.log
```

---

## 📋 Interactive Menu

After sorting, Vaultly stays open with an interactive menu:

```
📋 What would you like to do?
================================
[0] Exit
[1] Custom Move (assign unknown files)
[2] Undo Last Sort
[3] View Unhandled Files
[4] Re-scan & Sort Again

Choose: _
```

### Option 1: Custom Move

Assign unknown file types (like `.jar`, `.apk`, `.json`) to any folder:

```
📁 Custom Move — Select extension:
================================
[1] .json (42 files)
[2] .jar (22 files)
[3] .apk (8 files)
[0] Cancel
Choose: 2

Move 22 .jar files to:
[1] 📷 Pictures
[2] 🎵 Music
[3] 🎬 Videos
[4] 📄 Documents
[5] 📦 Archives
[6] 💻 Programs
[7] 📁 New folder (enter custom path)
Choose: 7
Enter full folder path: D:\Java Projects

✅ Moved 22 .jar files to D:\Java Projects
Save ".jar → Java Projects" as permanent rule? (y/n): y
✅ Rule saved! .jar files will auto-move in future runs.
```

### Option 2: Undo Last Sort

Revert all moves from the current session:

```
↩️  Undo 47 file moves?
Files will be moved back to their original location.
Confirm (y/n): y
================================
↩️  photo.jpg → D:\Downloads\photo.jpg
↩️  song.mp3 → D:\Downloads\song.mp3
...
================================
✅ Undone 47/47 moves.
```

### Option 3: View Unhandled Files

See unknown files grouped by extension with counts:

```
📋 Unhandled Files by Extension
================================
   42 Files | .json
   22 Files | .jar
   15 Files | .mcworld
    8 Files | .apk
    5 Files | .schem
    3 Files | (no extension)
--------------------------------
Total: 95 unhandled files
```

---

## 🔄 Smart Duplicate Handling

When a file with the same name already exists at the destination, Vaultly prompts you:

```
🔄 Duplicate found: photo.jpg
   Source:      D:\Downloads\photo.jpg (2.45 MB)
   Existing:   D:\Pictures\photo.jpg (2.45 MB)
   ✔ Files are IDENTICAL (same size + SHA256)

[1] Move & Rename (keep both)        → photo_1.jpg
[2] Remove Duplicate (delete source, keep existing)
[3] Skip

Choose (1/2/3): _
```

- **Identical check**: Compares file size AND SHA256 hash — not just the filename
- **Move & Rename**: Keeps both files, renames to `photo_1.jpg`, `photo_2.jpg`, etc.
- **Remove Duplicate**: Deletes the source file (only if confirmed identical)
- **Safety warning**: If files have different content, warns before deletion

```
🔄 Duplicate found: report.pdf
   Source:      D:\Downloads\report.pdf (1.20 MB)
   Existing:   D:\Documents\report.pdf (856.00 KB)
   ✖ Files are DIFFERENT (different content)

[1] Move & Rename (keep both)
[2] Remove Duplicate (not recommended — files differ!)
[3] Skip
```

---

## 📁 Recursive Subfolder Scanning

Vaultly automatically scans all subfolders inside your Downloads folder. Files in subfolders are organized the same way as root-level files.

Vaultly's own folders (`Archives`, `Programs`) are automatically excluded from scanning.

---

## 📂 Supported File Types

| Category | Extensions |
|----------|-----------|
| 📷 Pictures | `.jpg` `.jpeg` `.jfif` `.png` `.gif` `.bmp` `.svg` `.webp` `.ico` `.tiff` `.tif` `.avif` `.heic` `.heif` `.raw` `.cr2` `.nef` |
| 🎵 Music | `.mp3` `.wav` `.flac` `.aac` `.ogg` `.wma` `.m4a` |
| 🎬 Videos | `.mp4` `.mkv` `.avi` `.mov` `.wmv` `.flv` `.webm` |
| 📄 Documents | `.pdf` `.doc` `.docx` `.xls` `.xlsx` `.ppt` `.pptx` `.txt` `.csv` `.odt` |
| 📦 Archives | `.zip` `.rar` `.7z` `.tar` `.gz` `.bz2` `.xz` |
| 💻 Programs | `.exe` `.msi` `.msix` `.appx` `.bat` `.cmd` `.ps1` `.dll` |

> Unknown file types are skipped and can be assigned via the Custom Move menu.

---

## 📦 Archive & Program Handling

Archives and programs get a **single batch prompt** (not per-file):

```
📦 Found 56 archive files (.zip, .rar, .7z, .tar, .gz, etc.)
   • project.zip
   • data.rar
   • backup.7z
   ... and 53 more

[1] Move all to D:\Downloads\Archives
[2] Skip all
Choose (1/2): _
```

```
💻 Found 23 program files (.exe, .msi, .dll, etc.)
   • vs_BuildTools.exe
   • runtime.msi
   ... and 21 more

[1] Move all to D:\Downloads\Programs
[2] Skip all
Choose (1/2): _
```

Use `--auto-archive` to skip prompts and auto-move everything.

---

## ⚙️ Config File

Location: `C:\Users\{username}\.vaultlyrc.json`

```json
{
  "version": "1",
  "paths": {
    "source": "C:\\Users\\John\\Downloads",
    "pictures": "D:\\My Photos",
    "music": "D:\\Music",
    "videos": "C:\\Users\\John\\Videos",
    "documents": "C:\\Users\\John\\Documents",
    "archives": "C:\\Users\\John\\Downloads\\Archives",
    "programs": "C:\\Users\\John\\Downloads\\Programs"
  },
  "options": {
    "autoArchive": false,
    "watchMode": false,
    "logFile": "C:\\Users\\John\\vaultly.log",
    "duplicateHandling": "rename",
    "ignoreFiles": [".gitkeep", ".DS_Store", "desktop.ini", "thumbs.db", "*.tmp"]
  },
  "customRules": [
    { "extension": ".psd", "destination": "pictures" },
    { "extension": ".ai", "destination": "pictures" },
    { "extension": ".epub", "destination": "documents" },
    { "extension": ".jar", "destination": "D:\\Java Projects" }
  ]
}
```

To edit: run `vaultly --config` to open in Notepad.

### Custom Rules

Add custom file type mappings in the `customRules` array. The `destination` can be a category name (`pictures`, `music`, `videos`, `documents`, `archives`, `programs`) or a full folder path.

---

## 📝 Log File

Location: `C:\Users\{username}\vaultly.log`

```
[2024-01-15 10:30:00] MOVED     photo.jpg            → D:\My Photos\photo.jpg
[2024-01-15 10:30:01] MOVED     song.mp3             → D:\Music\song.mp3
[2024-01-15 10:30:02] RENAMED   photo.jpg            → D:\My Photos\photo_1.jpg (duplicate)
[2024-01-15 10:30:03] SKIPPED   project.zip          → user chose skip (batch)
[2024-01-15 10:30:04] SKIPPED   movie.mp4            → file is currently in use
[2024-01-15 10:30:05] UNKNOWN   design.fig           → no rule defined
[2024-01-15 10:30:06] ARCHIVED  data.zip             → Downloads\Archives\data.zip
[2024-01-15 10:30:07] ERROR     report.pdf           → permission denied
```

---

## 🤝 Contributing

Contributions are welcome! Here's how:

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

### Development

```bash
# Clone the repo
git clone https://github.com/epromite/vaultly.git
cd vaultly

# Build
go build -o vaultly.exe .

# Run
./vaultly.exe --dry-run
```

### Project Structure

```
vaultly/
├── main.go                      ← CLI entry point
├── go.mod / go.sum
├── .github/workflows/release.yml ← Auto-build binaries
└── internal/
    ├── classifier/classifier.go  ← Extension → category mapping
    ├── mover/mover.go            ← Move logic + duplicates + undo
    ├── archive/archive.go        ← Archive/program batch prompts
    ├── watcher/watcher.go        ← Watch mode (fsnotify)
    ├── config/config.go          ← Config + setup wizard
    ├── logger/logger.go          ← File + terminal logging
    ├── menu/menu.go              ← Interactive post-sort menu
    └── winpath/winpath.go        ← Windows API folder detection
```

---

## 📄 License

[MIT License](LICENSE) — do whatever you want with it.

---

Made with ❤️ for a cleaner Downloads folder.
