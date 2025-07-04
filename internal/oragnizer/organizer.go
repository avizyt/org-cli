package oragnizer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	SourceDir string
	DestDir   string
	DryRun    bool
	Recursive bool
	Workers   int
}

// File move represents a single file operation task
type FileMove struct {
	SourcePath string
	DestPath   string
	DryRun     bool
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

// moveFile performs the actual file moving operation, including collision resolution.
func moveFile(fm FileMove) error {
	// Ensure the destination directory exists
	destDir := filepath.Dir(fm.DestPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if fm.DryRun {
			fmt.Printf(" DRY RUN: Would create directory: %s\n", destDir)
		} else {
			// 0755 permission: owner can read/write/execute. grou/others can read/execute
			err := os.MkdirAll(destDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create destination directory '%s': %w", destDir, err)
			}
			fmt.Printf("    Created directory: %s\n", destDir)
		}
	}

	// Collision Resolution: Check if target file already exists
	finalDestPath := fm.DestPath
	if _, err := os.Stat(finalDestPath); err == nil {
		// file exists, append timestamp to make ot unique
		ext := filepath.Ext(fm.DestPath)
		name := strings.TrimSuffix(filepath.Base(fm.DestPath), ext)
		timestamp := time.Now().Format("20060102_150405")
		finalDestPath = filepath.Join(destDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
		fmt.Printf("    Collision detected! Renaming '%s' to '%s'\n", filepath.Base(fm.DestPath), filepath.Base(finalDestPath))
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking existance of '%s': %w", finalDestPath, err)
	}

	if fm.DryRun {
		fmt.Printf("DRY RUN: Would move '%s' to '%s'\n", fm.SourcePath, finalDestPath)
	} else {
		err := os.Rename(fm.SourcePath, finalDestPath)
		if err != nil {
			return fmt.Errorf("failed to move '%s' to '%s': %w", fm.SourcePath, finalDestPath, err)
		}
		fmt.Printf("    Moved '%s' to '%s'\n", fm.SourcePath, finalDestPath)
	}
	return nil

}

func OrganizeFiles(cfg Config) error {
	fmt.Printf("Starting file organization from '%s' to '%s'...\n", cfg.SourceDir, cfg.DestDir)
	if cfg.DryRun {
		fmt.Println("!!! DRY RUN MODE: No files will be moved or created. !!!")
	}

	workQueue := make(chan FileMove, cfg.Workers*2)

	var wg sync.WaitGroup
	fileCount := 0
	movedCount := 0
	errorCount := 0

	// Start worker goroutines
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1) // Increment WaitGroup counter for each worker
		go func(workerID int) {
			defer wg.Done() // Decrement WaitGroup counter when worker exits
			for fm := range workQueue {
				err := moveFile(fm)
				if err != nil {
					fmt.Printf("    [ERROR] Worker %d failed to process %s: %v\n", workerID, fm.SourcePath, err)
					// We could use a mutex to increment errorCount safely, but for simple counting,
					// direct increment might be acceptable if races are rare and not critical.
					// For production, consider sync.Atomic for counters.
					errorCount++
				} else {
					movedCount++
				}
			}
		}(i) // Pass workerID to the goroutine
	}

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

		// fmt.Printf("  [FILE] %s -> Category: %s\n", fileName, category)
		// fmt.Printf("    Proposed move: %s -> %s\n", path, targetFilePath)

		// In later steps, we'll add the actual file moving logic here,
		// potentially dispatching to a worker pool.

		// Dispatch the file move task to the worker pool
		workQueue <- FileMove{
			SourcePath: path,
			DestPath:   targetFilePath,
			DryRun:     cfg.DryRun,
		}

		return nil
	})

	close(workQueue)

	wg.Wait()

	if err != nil {
		return fmt.Errorf("error walking directory '%s': %w", cfg.SourceDir, err)
	}

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Scanned %d files.\n", fileCount)
	if cfg.DryRun {
		fmt.Printf("Dry run completed. %d files would have been processed.\n", movedCount)
	} else {
		fmt.Printf("Successfully processed %d files.\n", movedCount)
	}
	fmt.Printf("Encountered %d errors during processing.\n", errorCount)
	fmt.Println("File organization process completed.")
	return nil
}
