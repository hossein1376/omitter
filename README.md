# 🚀 Omitter

A tool to remove that annoying text from file and folder names. It recursively
walks down the given directory, and remove the provided string from the names. It supports **dry-run** mode for safe previews, **interactive** confirmation for extra control, and a **verbose** option for detailed logging and more.

## Features ✨

- **Dry-Run Mode (`-d`)**: Preview changes without modifying any files.
- **Interactive Mode (`-i`)**: Get a confirmation prompt before applying changes.
- **Regex Mode (`-r`)**: Accept regex(regular expression) on -s flag.
- **File type filter (`-t`)**: Filter files based on provided extention(sample: -t .txt).
- **Verbose Output (`-v`)**: See detailed logs of the operations.
- **Flexible String Matching**: Remove a given substring from file names.
- **Easy Integration**: Can be used in scripts or manually via command-line.

## Installation 🔧

1. **Clone the repository:**

   ```bash
   git clone https://github.com/hossein1376/omitter.git
   cd omitter
   ```

2. **Build the binary:**

   ```bash
   go build -o omitter main.go
   ```

## Usage 📝

Run the omitter with the following options:

```bash
./omitter -p /path/to/directory -s "" [options]
```

Example regex command:

```bash
./omitter -p /path/to/directory -s "\\d+" -r [options]
```

Example filter by extention:

```bash
./omitter -p /path/to/directory -s "\\d+" -r -t ".txt" [options]
```

### Options

- **`-p`**: Path to the directory containing files.
- **`-s`**: The substring to find (and remove).it can be regex(regular expression) too when -r flag is enabled.
- **`-v`**: Enable verbose output.
- **`-d`**: Enable dry-run mode to preview changes.
- **`-i`**: Enable interactive mode to ask for confirmation before renaming.
- **`-r`**: Enable regex mode to accept regular expression.
- **`-t`**: Filter by file type for correction.

## License 📄

Distributed under the MIT License. See LICENSE for more information.
