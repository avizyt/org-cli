# Go-FOC

Go file organizer CLI (foc)
-----

## 🚀 Organize Your Digital Life with Speed and Precision\!

**Go-FOC** is a high-performance, customizable command-line utility built with Go, designed to automatically sort and organize your files into categorized directories. Say goodbye to cluttered download folders and messy document archives – let Go File Organizer do the heavy lifting for you\!

## ✨ Features

This tool is engineered for efficiency and flexibility, incorporating Go's powerful concurrency model to handle large volumes of files with ease.

1.  **Intelligent Core Sorting Logic:** Scans a specified source directory and automatically categorizes files based on their extensions (e.g., `.jpg` to `Images`, `.pdf` to `Documents`). Includes a default set of common file types and an `Others` category for unrecognized extensions.

2.  **Robust Command-Line Interface (CLI):** Provides intuitive flags for configuring source/destination directories, `dry-run` mode, `recursive` scanning, `workers` count, and custom `config` file paths.

3.  **Concurrent File Operations (Worker Pool):** Leverages Go's lightweight goroutines and channels to process and move multiple files simultaneously. The configurable worker pool maximizes I/O throughput, significantly reducing processing time for large datasets.

4.  **Smart Collision Resolution:** Prevents accidental overwrites by automatically detecting existing files with the same name. Duplicates are renamed by appending a timestamp (e.g., `my_photo.jpg` becomes `my_photo_20250704_153000.jpg`).

5.  **Customizable Category Mappings:** Define your own file extension-to-category mappings via a simple JSON configuration file. Custom rules override default categorizations, giving you complete control.

6.  **Beautiful & Informative CLI Experience:**

      * **Colored Output:** Clear, color-coded messages for status updates (success, errors, warnings, dry-run actions).
      * **Interactive Progress Bar:** Real-time visual feedback during file processing, indicating progress.
      * **Quiet Mode (`--quiet`):** Suppress detailed per-file output for faster, cleaner runs on large datasets, showing only the progress bar and final summary.
      * **Execution Time Tracking:** Reports the total time taken for the entire organization process in the final summary.

-----

## ⚙️ Installation

To get started, ensure you have Go installed (version 1.18 or higher recommended).

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/org-cli.git # Replace with your actual repo URL
    cd org-cli
    ```

2.  **Build the executable:**

    ```bash
    go build -o organizer ./cmd/organizer
    ```

    This will create an executable named `organizer` (or `organizer.exe` on Windows) in your project root.

3.  **(Optional) Add to your PATH:** For easy access from any directory, move the `organizer` executable to a location included in your system's PATH (e.g., `/usr/local/bin` on Linux/macOS, or a custom bin directory on Windows).

-----

## 🚀 Usage

### Basic Command Structure

```bash
./organizer --source <source_directory> --dest <destination_directory> [flags]
```

### Flags

  * `--source <path>` (required): The directory containing files to be organized.
  * `--dest <path>` (required): The root directory where organized category folders will be created.
  * `--dry-run` (optional): Simulate the process without moving or creating anything.
  * `--recursive` (optional): Scan and organize files within subdirectories.
  * `--workers <number>` (optional): Number of concurrent file operations (default: `5`). Adjust for optimal performance based on your system.
  * `--config <path>` (optional): Path to a JSON file for custom category mappings.
  * `--quiet` (optional): Suppress detailed per-file output, showing only progress and summary.

### Examples

1.  **Dry Run (Previewing Changes):**

    ```bash
    ./organizer --source ~/Downloads --dest ~/OrganizedFiles --dry-run
    ```

2.  **Organize Recursively:**

    ```bash
    ./organizer --source ~/MyMessyFolder --dest ~/CleanedUp --recursive
    ```

3.  **Organize with More Workers (e.g., for SSDs):**

    ```bash
    ./organizer --source /mnt/data/raw --dest /mnt/data/processed --workers 20
    ```

4.  **Using Custom Mappings (and quiet mode):**

    First, create a `my_mappings.json` file:

    ```json
    {
      ".log": "Application Logs",
      "svg": "Vector Graphics",
      ".md": "Markdown Notes",
      ".bak": "Backups"
    }
    ```

    Then run:

    ```bash
    ./organizer --source ~/ProjectFiles --dest ~/OrganizedProjects --config my_mappings.json --recursive --quiet
    ```

-----

## ⚡ Performance & Concurrency

Go File Organizer leverages Go's concurrency model with goroutines and channels to parallelize file scanning and moving, making efficient use of multi-core processors and I/O bandwidth. The configurable worker pool ensures optimal throughput. For large file sets, increasing the `--workers` count and using the `--quiet` flag are highly recommended to maximize speed.

-----

## 🛡️ Collision Resolution

To prevent data loss, if a file with the same name already exists in the target category folder, the new file will be automatically renamed by appending a timestamp before its extension (e.g., `report.pdf` becomes `report_20250704_220740.pdf`).

-----

## 🤝 Contributing

Contributions are welcome\! If you have ideas for new features, bug fixes, or performance improvements, please feel free to:

1.  Fork the repository.
2.  Create a new branch (`git checkout -b feature/your-feature`).
3.  Make your changes.
4.  Commit your changes (`git commit -m 'Add new feature'`).
5.  Push to the branch (`git push origin feature/your-feature`).
6.  Open a Pull Request.

-----

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.

-----

## 🙏 Acknowledgements

  * [fatih/color](https://github.com/fatih/color) for beautiful colored output.
  * [schollz/progressbar/v3](https://github.com/schollz/progressbar) for the elegant progress bar.