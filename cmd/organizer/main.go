// cmd/organizer/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync" // For waiting on the progress collector goroutine
	"time"

	"github.com/avizyt/org-cli/internal/organizer" // Replace with your module path
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

func main() {

	startTime := time.Now()
	// Define colors for initial messages
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()

	fmt.Println(blue("✨ Go File Organizer CLI ✨"))

	// 1. Define command-line flags
	sourceDir := flag.String("source", "", "Source directory to organize files from (required)")
	destDir := flag.String("dest", "", "Destination directory to move organized files to (required)")
	dryRun := flag.Bool("dry-run", false, "If true, only simulate actions without moving files")
	recursive := flag.Bool("recursive", false, "If true, scan and organize files in subdirectories")
	workers := flag.Int("workers", 5, "Number of concurrent file operations (default 5)")
	configPath := flag.String("config", "", "Path to a JSON configuration file for custom category mappings")
	quiet := flag.Bool("quiet", false, "Suppress detailed per-file output during processing (show only progress and summary)") // New flag

	// 2. Parse the flags
	flag.Parse()

	// 3. Basic validation for required arguments
	if *sourceDir == "" {
		fmt.Fprintln(os.Stderr, red("Error: --source directory is required."))
		flag.Usage()
		os.Exit(1)
	}
	if *destDir == "" {
		fmt.Fprintln(os.Stderr, red("Error: --dest directory is required."))
		flag.Usage()
		os.Exit(1)
	}

	// Resolve absolute paths for robustness
	absSourceDir, err := filepath.Abs(*sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, red("Error resolving absolute path for source directory '%s': %v\n"), *sourceDir, err)
		os.Exit(1)
	}
	absDestDir, err := filepath.Abs(*destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, red("Error resolving absolute path for destination directory '%s': %v\n"), *destDir, err)
		os.Exit(1)
	}

	// Initialize category mappings with defaults
	categoryMappings := organizer.DefaultCategoryMappings()

	// Load and merge custom mappings if a config path is provided
	if *configPath != "" {
		fmt.Printf("%s Loading custom category mappings from '%s'...\n", blue("⚙️"), *configPath)
		customMappings, err := loadCustomMappings(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, red("Error loading custom mappings from '%s': %v\n"), *configPath, err)
			os.Exit(1)
		}

		// Merge custom mappings (custom overrides defaults)
		for ext, category := range customMappings {
			categoryMappings[ext] = category
		}
		fmt.Println(green("✔ Custom mappings loaded and merged."))
	}

	// Create the Config struct
	cfg := organizer.Config{
		SourceDir:        absSourceDir,
		DestDir:          absDestDir,
		DryRun:           *dryRun,
		Recursive:        *recursive,
		Workers:          *workers,
		CategoryMappings: categoryMappings,
		Quiet:            *quiet,
	}

	// Create a channel for progress updates from the organizer
	progressChan := make(chan organizer.ProgressUpdate, cfg.Workers+10)

	// Initialize the progress bar
	bar := progressbar.NewOptions(0, // Max is 0 initially, will be set after scanning
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription("[cyan]Processing files...[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionClearOnFinish(),
	)

	// Variables to aggregate counts from workers
	var totalProcessed int // Renamed from movedCount to be more general (dry-run counts as processed)
	var totalErrors int
	var wgProgress sync.WaitGroup // New WaitGroup for the progress collector goroutine

	// Goroutine to update the progress bar and collect counts based on messages from progressChan
	wgProgress.Add(1)
	go func() {
		defer wgProgress.Done()
		for update := range progressChan {
			totalProcessed += update.Moved
			totalErrors += update.Errored
			bar.Add(update.Moved)
		}
		bar.Finish() // Ensure bar finishes when channel is closed
	}()

	// 4. Call the organizer logic with the parsed config and progress channel
	totalScanned, totalFilesToProcess, totalSkipped, scanErr := organizer.OrganizeFiles(cfg, progressChan)
	if scanErr != nil {
		fmt.Fprintf(os.Stderr, red("Error during file scanning: %v\n"), scanErr)
		// Don't exit immediately, let summary print
	}

	// Set the max value of the progress bar after scanning
	bar.ChangeMax(totalFilesToProcess)

	// Close the progress channel. This signals the progress collector goroutine to finish.
	close(progressChan)

	// Wait for the progress collector goroutine to finish
	wgProgress.Wait()

	// Final newline after progress bar
	fmt.Println()

	endTime := time.Now() // End timing the operation
	duration := endTime.Sub(startTime)

	fmt.Println(blue("🎉 Organizer finished."))
	fmt.Printf("%s --- Summary ---\n", blue("📄"))
	fmt.Printf("%s Total files scanned: %s\n", blue("🔍"), green(fmt.Sprintf("%d", totalScanned)))
	fmt.Printf("%s Files to process: %s\n", blue("📦"), green(fmt.Sprintf("%d", totalFilesToProcess)))
	fmt.Printf("%s Files skipped (already in dest or access error): %s\n", yellow("⏩"), yellow(fmt.Sprintf("%d", totalSkipped)))
	if *dryRun {
		fmt.Printf("%s Dry run completed. %s files would have been processed.\n", green("✅"), green(fmt.Sprintf("%d", totalProcessed)))
	} else {
		fmt.Printf("%s Successfully processed %s files.\n", green("✅"), green(fmt.Sprintf("%d", totalProcessed)))
	}
	if totalErrors > 0 {
		fmt.Printf("%s Encountered %s errors during processing.\n", red("❌"), red(fmt.Sprintf("%d", totalErrors)))
	} else {
		fmt.Printf("%s No errors encountered during processing.\n", green("✔️"))
	}
	fmt.Printf("%s Total time taken: %s\n", magenta("⏱️"), magenta(duration.Round(time.Millisecond).String())) // Print total time
}

// loadCustomMappings reads a JSON file and unmarshals it into a map.
func loadCustomMappings(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}

	mappings := make(map[string]string)
	err = json.Unmarshal(data, &mappings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON config file '%s': %w", filePath, err)
	}

	// Normalize keys to lowercase (extensions)
	normalizedMappings := make(map[string]string)
	for ext, category := range mappings {
		// Ensure extension starts with a dot
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		normalizedMappings[strings.ToLower(ext)] = category
	}

	return normalizedMappings, nil
}
