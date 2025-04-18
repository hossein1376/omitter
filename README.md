# 🚀 Omitter

A tool to remove that annoying text from file and folder names. It recursively
walks down the given directory, and remove the provided string from the names. It supports **dry-run** mode for safe previews, **interactive** confirmation for extra control, and a **verbose** option for detailed logging and more.

## Features ✨

- **Dry-Run Mode (`-d`)**: Preview changes without modifying any files.
- **Interactive Mode (`-i`)**: Get a confirmation prompt before applying changes.
- **Regex Mode (`-r`)**: Accept regex(regular expression) on -s flag.
- **File type filter (`-t`)**: Filter files based on provided extension(sample: -t .txt).
- **Replace mode (`-replace`)**: Replace instead of removing.
- **Different output (`-output`)**: Copy to desired output dir.
- **Verbose Output (`-v`)**: See detailed logs of the operations.
- **Verbose Output (`-tt`)**: Set transmission type when output is exist. default set to copy.
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

Example filter by extension:

```bash
./omitter -p /path/to/directory -s "\\d+" -r -t ".txt" [options]
```

Example replace mode:

🛎In replace mode, if multiple files resolve to the same name, the utility automatically appends a numeric suffix (e.g., \_1, \_2) to ensure each renamed file remains unique and no data is lost.

```bash
./omitter -p /path/to/directory -s "aaa" --replace bbb [options]
```

Example output flag(copy):

```bash
./omitter -p /path/to/directory -s "aaa" --output /path/to/target/output [options]

or

./omitter -p /path/to/directory -s "aaa" --output /path/to/target/output -tt cp [options]

or

./omitter -p /path/to/directory -s "aaa" --output /path/to/target/output -tt copy [options]
```

Example output flag(move):

```bash
./omitter -p /path/to/directory -s "aaa" --output /path/to/target/output -tt mv [options]

or

./omitter -p /path/to/directory -s "aaa" --output /path/to/target/output -tt move [options]
```

### Options

- **`-p`**: Path to the directory containing files.
- **`-s`**: The substring to find (and remove).it can be regex(regular expression) too when -r flag is enabled.
- **`-v`**: Enable verbose output.
- **`-d`**: Enable dry-run mode to preview changes.
- **`-i`**: Enable interactive mode to ask for confirmation before renaming.
- **`-r`**: Enable regex mode to accept regular expression.
- **`-t`**: Filter by file type for correction.
- **`-tt`**: Set transmission type(copy/move). default is copy.
- **`-replace`**: Replace instead of removing.
- **`-output`**: Copy to new dir instead of rename in path flag dir.
- **`-help`**: Print usage of omitter.

## License 📄

Distributed under the MIT License. See LICENSE for more information.
