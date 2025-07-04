package oragnizer

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Config struct {
	SourceDir string
	DestDir   string
	DryRun    bool
	Recursive bool
	Workers   int
}

var defaultCategoryMapping = map[string]string{
	// Images
	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".gif":  "Images",
	".bmp":  "Images",
	".tiff": "Images",
	".webp": "Images",
	".heic": "Images",

	// Documents
	".pdf":  "Documents",
	".doc":  "Documents",
	".docx": "Documents",
	".ppt":  "Documents",
	".pptx": "Documents",
	".xls":  "Documents",
	".xlsx": "Documents",
	".txt":  "Documents",
	".rtf":  "Documents",
	".odt":  "Documents",

	// Videos
	".mp4":  "Videos",
	".mov":  "Videos",
	".avi":  "Videos",
	".mkv":  "Videos",
	".webm": "Videos",

	// Audio
	".mp3":  "Audio",
	".wav":  "Audio",
	".flac": "Audio",
	".aac":  "Audio",

	// Archives
	".zip": "Archives",
	".rar": "Archives",
	".7z":  "Archives",
	".tar": "Archives",
	".gz":  "Archives",

	// Executables
	".exe": "Executables",
	".dmg": "Executables",
	".app": "Executables", // macOS applications
	".deb": "Executables", // Debian packages
	".rpm": "Executables", // Red Hat packages

	// Code
	".go":   "Code",
	".js":   "Code",
	".ts":   "Code",
	".py":   "Code",
	".java": "Code",
	".c":    "Code",
	".cpp":  "Code",
	".h":    "Code",
	".hpp":  "Code",
	".html": "Code",
	".css":  "Code",
	".json": "Code",
	".xml":  "Code",
	".md":   "Code",
}

func OrganizeFiles(cfg Config) error {
	fmt.Printf("Starting file organization from '%s' to '%s'...\n", cfg.SourceDir, cfg.DestDir)
	if cfg.DryRun {
		fmt.Println("!!! DRY RUN MODE: No files will be moved or created. !!!")
	}

	// var wg sync.WaitGroup
	fileCount := 0

	err := filepath.WalkDir(cfg.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return err
		}

		if d.IsDir() {
			if !cfg.Recursive && path != cfg.SourceDir {
				return filepath.SkipDir
			}
			return nil
		}

		fileCount++
		ext := strings.ToLower(filepath.Ext(path))
		fileName := filepath.Base(path)

		category, ok := defaultCategoryMapping[ext]
		if !ok {
			category = "Others"
		}

		targetCategoryDir := filepath.Join(cfg.DestDir, category)
		targetFilePath := filepath.Join(targetCategoryDir, fileName)

		fmt.Printf("  [FILE] %s -> Category: %s\n", fileName, category)
		fmt.Printf("    Proposed move: %s -> %s\n", path, targetFilePath)

		// In later steps, we'll add the actual file moving logic here,
		// potentially dispatching to a worker pool.

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory '%s': %w", cfg.SourceDir, err)
	}

	fmt.Printf("Scanned %d files.\n", fileCount)
	fmt.Println("File organization process completed (dry run).")
	return nil
}
