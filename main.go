package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var path, str string
	var verbose bool
	flag.StringVar(&path, "p", "", "path to dir")
	flag.StringVar(&str, "s", "", "string to find")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	if path == "" || str == "" {
		flag.Usage()
		os.Exit(1)
	}

	start := time.Now()
	n, err := rename(path, str)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if verbose {
		fmt.Printf("Renamed %d file(s) in %s.\n", n, time.Since(start))
	}
}

func rename(base, str string) (int, error) {
	var renamed int
	return renamed, filepath.WalkDir(base, func(path string, file fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case file.IsDir():
			return nil
		}

		newPath := filepath.Join(filepath.Dir(path), strings.ReplaceAll(file.Name(), str, ""))
		if path == newPath {
			return nil
		}

		if err = os.Rename(path, newPath); err != nil {
			return err
		}
		renamed++

		return nil
	})
}
