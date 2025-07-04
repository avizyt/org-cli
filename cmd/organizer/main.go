package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/avizyt/org-cli/internal/oragnizer"
)

func main() {
	fmt.Println("Go File Organizer CLI")

	// 1. Define command-line flags
	sourceDir := flag.String("source", "", "Source directory to organize files from (required)")
	destDir := flag.String("dest", "", "Destination directory to move organized files to (required)")
	dryRun := flag.Bool("dry-run", false, "If true, only simulate actions without moving files")
	recursive := flag.Bool("recursive", false, "If true, scan and organize files in subdirectories")
	workers := flag.Int("workers", 5, "Number of concurrent file operations (default 5)")

	// 2. Parse the flags
	flag.Parse()

	// 3. Basic validation for required arguments
	if *sourceDir == "" {
		fmt.Fprintln(os.Stderr, "Error: --source directory is required.")
		flag.Usage() // Print usage instructions
		os.Exit(1)
	}
	if *destDir == "" {
		fmt.Fprintln(os.Stderr, "Error: --dest directory is required.")
		flag.Usage()
		os.Exit(1)
	}

	// Resolve absolute paths for robustness
	absSourceDir, err := filepath.Abs(*sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving absolute path for source directory '%s': %v\n", *sourceDir, err)
		os.Exit(1)
	}
	absDestDir, err := filepath.Abs(*destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving absolute path for destination directory '%s': %v\n", *destDir, err)
		os.Exit(1)
	}

	// Create the Config struct
	cfg := oragnizer.Config{
		SourceDir: absSourceDir,
		DestDir:   absDestDir,
		DryRun:    *dryRun,
		Recursive: *recursive,
		Workers:   *workers,
	}

	// 4. Call the organizer logic with the parsed config
	err = oragnizer.OrganizeFiles(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during file organization: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Organizer finished.")
}
