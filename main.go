package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	var path, str string
	var verbose, dryRun, interactive, regex bool
	flag.StringVar(&path, "p", "", "path to dir")
	flag.StringVar(&str, "s", "", "string to find")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.BoolVar(&dryRun, "d", false, "dry run mode, only print the changes")
	flag.BoolVar(&interactive, "i", false, "interactive mode, ask for confirmation before renaming")
	flag.BoolVar(&regex, "r", false, "regex mode, accept regex")
	flag.Parse()

	if path == "" || str == "" {
		flag.Usage()
		os.Exit(1)
	}

	// If interactive mode is enabled, count the files that would be modified.
	if interactive && !dryRun {
		candidates, err := countRenameCandidates(path, str, regex)
		if err != nil {
			fmt.Println("Error counting files:", err)
			os.Exit(2)
		}
		if candidates == 0 {
			fmt.Println("No files to rename.")
			return
		}
		fmt.Printf("Found %d file(s) to rename.\n", candidates)
		if !YesOrNoPrompt("Proceed with renaming?", false) {
			fmt.Println("Aborted by user.")
			return
		}
	}

	start := time.Now()
	n, err := rename(path, str, dryRun, regex)
	if err != nil {
		fmt.Println("Error renaming files:", err)
		os.Exit(2)
	}
	if verbose {
		mode := "Renamed"
		if dryRun {
			mode = "Processed (dry-run)"
		}
		fmt.Printf("%s %d file(s) in %s.\n", mode, n, time.Since(start))
	}
}

// rename walks the directory and renames files by removing the specified string.
// If dryRun is true, it only prints what would be done.
func rename(base, str string, dryRun bool, regex bool) (int, error) {
	var renamed int
	err := filepath.WalkDir(base, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() {
			return nil
		}

		targetStr, err := FindRegexString(regex, str, file.Name())
		if err != nil {
			return err
		}

		newPath := filepath.Join(filepath.Dir(path), strings.ReplaceAll(file.Name(), targetStr, ""))
		if path == newPath {
			return nil
		}

		if dryRun {
			// Only print the intended renaming in dry-run mode.
			fmt.Printf("[Dry-run] Would rename: %s -> %s\n", path, newPath)
		} else {
			if err = os.Rename(path, newPath); err != nil {
				return err
			}
		}
		renamed++
		return nil
	})
	return renamed, err
}

// countRenameCandidates walks the directory to count files that would be renamed.
func countRenameCandidates(base, str string, regex bool) (int, error) {
	var count int
	err := filepath.WalkDir(base, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() {
			return nil
		}

		targetStr, err := FindRegexString(regex, str, file.Name())
		if err != nil {
			return nil
		}

		newName := strings.ReplaceAll(file.Name(), targetStr, "")
		if file.Name() != newName {
			count++
		}
		return nil
	})
	return count, err
}

// YesOrNoPrompt prompts the user for confirmation and returns true for yes.
func YesOrNoPrompt(label string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

// Handle compiling regex if -r flag is true
func FindRegexString(regex bool, str string, fileName string) (string, error) {
	var targetStr string
	if regex {
		r, err := regexp.Compile(str)
		if err != nil {
			return "", err
		}

		targetStr = r.FindString(fileName)
	} else {
		targetStr = str
	}

	return targetStr, nil
}
