package organizer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color" // Import for colored output
)

// Config holds the configuration for the file organizer.
type Config struct {
	SourceDir        string            // Directory to scan
	DestDir          string            // Directory where organized files will be moved
	DryRun           bool              // If true, only print actions, don't move files
	Recursive        bool              // If true, scan subdirectories
	Workers          int               // Number of concurrent workers for file operations
	CategoryMappings map[string]string // Custom or merged category mappings
}

// FileMove represents a single file operation task.
type FileMove struct {
	SourcePath string // Original path of the file
	DestPath   string // Target path for the file
	DryRun     bool   // Whether this is a dry run
}

// ProgressUpdate is sent by workers to report their status.
type ProgressUpdate struct {
	Moved   int
	Errored int
}

// DefaultCategoryMappings defines common file extensions and their default categories.
func DefaultCategoryMappings() map[string]string {
	return map[string]string{
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
}

// moveFile performs the actual file moving operation, including collision resolution.
// It sends progress updates to the provided channel.
func moveFile(fm FileMove, progressChan chan<- ProgressUpdate) error {
	defer func() {
		// Ensure a progress update is sent even if an error occurs
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in moveFile: %v\n", r)
			progressChan <- ProgressUpdate{Errored: 1}
		}
	}()

	// Define colors for output
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Ensure the destination directory exists
	destDir := filepath.Dir(fm.DestPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if fm.DryRun {
			fmt.Printf("    %s: Would create directory: %s\n", cyan("DRY RUN"), destDir)
		} else {
			err := os.MkdirAll(destDir, 0755)
			if err != nil {
				progressChan <- ProgressUpdate{Errored: 1}
				return fmt.Errorf("failed to create destination directory '%s': %w", destDir, err)
			}
			fmt.Printf("    %s: Created directory: %s\n", green("CREATED"), destDir)
		}
	}

	// Collision Resolution: Check if target file already exists
	finalDestPath := fm.DestPath
	if _, err := os.Stat(finalDestPath); err == nil {
		// File exists, append timestamp to make it unique
		ext := filepath.Ext(fm.DestPath)
		name := strings.TrimSuffix(filepath.Base(fm.DestPath), ext)
		timestamp := time.Now().Format("20060102_150405") // YYYYMMDD_HHMMSS
		finalDestPath = filepath.Join(destDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
		fmt.Printf("    %s: Renaming '%s' to '%s'\n", yellow("COLLISION"), filepath.Base(fm.DestPath), filepath.Base(finalDestPath))
	} else if !os.IsNotExist(err) {
		// Some other error occurred while checking file existence
		progressChan <- ProgressUpdate{Errored: 1}
		return fmt.Errorf("error checking existence of '%s': %w", finalDestPath, err)
	}

	if fm.DryRun {
		fmt.Printf("    %s: Would move '%s' to '%s'\n", cyan("DRY RUN"), fm.SourcePath, finalDestPath)
		progressChan <- ProgressUpdate{Moved: 1} // Still count as "moved" in dry run for progress
	} else {
		err := os.Rename(fm.SourcePath, finalDestPath)
		if err != nil {
			progressChan <- ProgressUpdate{Errored: 1}
			return fmt.Errorf("failed to move '%s' to '%s': %w", fm.SourcePath, finalDestPath, err)
		}
		fmt.Printf("    %s: Moved '%s' to '%s'\n", green("MOVED"), fm.SourcePath, finalDestPath)
		progressChan <- ProgressUpdate{Moved: 1}
	}
	return nil
}

// OrganizeFiles scans the source directory and dispatches file moves to a worker pool.
// It returns the total files scanned, moved, skipped, and errored.
func OrganizeFiles(cfg Config, progressChan chan<- ProgressUpdate) (int, int, int, int, error) {
	// Define colors for output
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Printf("%s Starting file organization from '%s' to '%s'...\n", blue("🚀"), cfg.SourceDir, cfg.DestDir)
	if cfg.DryRun {
		fmt.Println(yellow("!!! DRY RUN MODE: No files will be moved or created. !!!"))
	}

	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}

	// Phase 1: Scan and Collect Files
	fmt.Printf("%s Scanning files in '%s'...\n", blue("🔍"), cfg.SourceDir)
	var filesToMove []FileMove
	var scanError error
	skippedCount := 0

	err := filepath.WalkDir(cfg.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("%s Error accessing path %s: %v. Skipping.\n", red("❌"), path, err)
			// Don't count as an error for processing, but for scanning errors
			scanError = fmt.Errorf("encountered error during scan: %w", err)
			return nil // Continue walking other paths
		}

		if d.IsDir() {
			if !cfg.Recursive && path != cfg.SourceDir {
				return filepath.SkipDir
			}
			return nil
		}

		// It's a file, process it
		ext := strings.ToLower(filepath.Ext(path))
		fileName := filepath.Base(path)

		category, ok := cfg.CategoryMappings[ext]
		if !ok {
			category = "Others"
		}

		// Skip files that are already in the destination directory (or a subdirectory of it)
		if strings.HasPrefix(path, cfg.DestDir) {
			fmt.Printf("  %s %s is already in the destination directory. Skipping.\n", yellow("⚠️"), fileName)
			skippedCount++
			return nil
		}

		targetCategoryDir := filepath.Join(cfg.DestDir, category)
		targetFilePath := filepath.Join(targetCategoryDir, fileName)

		filesToMove = append(filesToMove, FileMove{
			SourcePath: path,
			DestPath:   targetFilePath,
			DryRun:     cfg.DryRun,
		})

		return nil
	})

	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error walking source directory '%s': %w", cfg.SourceDir, err)
	}
	if scanError != nil {
		fmt.Printf("%s Scan completed with some errors.\n", yellow("⚠️"))
	}

	totalFiles := len(filesToMove)
	if totalFiles == 0 {
		fmt.Printf("%s No files found to organize.\n", blue("ℹ️"))
		return 0, 0, skippedCount, 0, nil
	}

	fmt.Printf("%s Found %d files to process.\n", blue("✅"), totalFiles)

	// Phase 2: Process Files with Worker Pool
	workQueue := make(chan FileMove, cfg.Workers*2)
	var wg sync.WaitGroup
	movedCount := 0
	errorCount := 0

	// Start worker goroutines
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for fm := range workQueue {
				err := moveFile(fm, progressChan)
				if err != nil {
					// Error already reported by moveFile, just log here if needed for deeper debug
					// fmt.Printf("Worker %d: %v\n", workerID, err)
					// errorCount is incremented via progressChan
				}
			}
		}(i)
	}

	// Dispatch tasks to the worker pool
	for _, fm := range filesToMove {
		workQueue <- fm
	}
	close(workQueue) // Close the work queue after all files have been dispatched.

	// Goroutine to collect progress updates
	go func() {
		for update := range progressChan {
			movedCount += update.Moved
			errorCount += update.Errored
		}
	}()

	// Wait for all worker goroutines to finish their tasks.
	wg.Wait()
	close(progressChan) // Close the progress channel after all workers are done

	fmt.Printf("\n%s --- Summary ---\n", blue("📄"))
	fmt.Printf("%s Scanned %d files.\n", blue("🔍"), totalFiles+skippedCount) // Total scanned includes skipped
	fmt.Printf("%s Skipped %d files (already in destination or access error).\n", yellow("⏩"), skippedCount)
	if cfg.DryRun {
		fmt.Printf("%s Dry run completed. %s files would have been processed.\n", green("✅"), green(fmt.Sprintf("%d", movedCount)))
	} else {
		fmt.Printf("%s Successfully processed %s files.\n", green("✅"), green(fmt.Sprintf("%d", movedCount)))
	}
	if errorCount > 0 {
		fmt.Printf("%s Encountered %s errors during processing.\n", red("❌"), red(fmt.Sprintf("%d", errorCount)))
	} else {
		fmt.Printf("%s No errors encountered.\n", green("✔️"))
	}
	fmt.Printf("%s File organization process completed.\n", blue("🎉"))

	return totalFiles + skippedCount, movedCount, skippedCount, errorCount, nil // Return actual counts
}
