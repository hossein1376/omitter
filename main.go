package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pooulad/ravan"
)

const (
	RENAME string = "rename"
	COPY   string = "copy"
	MOVE   string = "move"
)

type fileOptions struct {
	path             string
	str              string
	fileType         string
	replace          string
	output           string
	transmissionType string
}
type config struct {
	options         fileOptions
	withVerbose     bool
	withDryRun      bool
	withInteractive bool
	withRegex       bool
	help            bool
}

func main() {
	cfg := parseFlags()
	if cfg.options.path == "" || cfg.options.str == "" || cfg.help {
		flag.Usage()
		os.Exit(1)
	}

	var pattern *regexp.Regexp
	var err error
	if cfg.withRegex {
		pattern, err = regexp.Compile(cfg.options.str)
		if err != nil {
			fmt.Println("compile pattern:", err)
			os.Exit(1)
		}
	}
	pairs, err := walker(cfg, pattern)
	if err != nil {
		fmt.Println("walk dir:", err)
		os.Exit(2)
	}

	actionName := getActionName(cfg.options.output, cfg.options.transmissionType)

	if cfg.withDryRun {
		fmt.Printf("Found %d file(s) to %s!\n", len(pairs), actionName)
		if cfg.withVerbose {
			for k, v := range pairs {
				fmt.Printf("%s -> %s\n", k, v)
			}
		}
		return
	}
	if cfg.withInteractive {
		fmt.Printf("Found %d file(s) to %s. Proceed?(y/n) ", len(pairs), actionName)
		if !canProceed() {
			fmt.Println("Aborted.")
			return
		}
	}

	start := time.Now()
	var n uint
	if cfg.options.output != "" {
		tt := getTransmissionType(cfg.options.transmissionType)
		var n uint
		var err error
		var message, vMessage string
		if tt == COPY {
			n, err = copyAction(pairs)
			message = fmt.Sprintf("%d file(s) were copied.", n)
			vMessage = fmt.Sprintf("Copied %d file(s)", n)

		} else {
			n, err = moveAction(pairs)
			message = fmt.Sprintf("%d file(s) were moved.", n)
			vMessage = fmt.Sprintf("Moved %d file(s)", n)
		}
		if err != nil {
			fmt.Printf("%s: %t", tt, err)
			fmt.Println(message)
			os.Exit(2)
		}
		if cfg.withVerbose {
			fmt.Printf("%s in %s.\n", vMessage, time.Since(start))
		}
	} else {
		n, err = renameAction(pairs)
		if err != nil {
			fmt.Println("Renaming:", err)
			fmt.Printf("%d file(s) were renamed.\n", n)
			os.Exit(2)
		}
		if cfg.withVerbose {
			fmt.Printf("Renamed %d file(s) in %s.\n", n, time.Since(start))
		}
	}
}

func walker(config config, pattern *regexp.Regexp,
) (map[string]string, error) {
	pairs := make(map[string]string)
	err := filepath.WalkDir(
		config.options.path,
		func(path string, file fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return err
			case file.IsDir():
				return nil
			}
			oldName := file.Name()
			fileExt := filepath.Ext(oldName)
			if config.options.fileType != "" && fileExt != "" {
				if fileExt != config.options.fileType {
					return nil
				}
			}
			targetStr := searchString(pattern, config.options.str, oldName)
			if config.withRegex && targetStr == "" {
				return nil
			}

			newName := strings.ReplaceAll(oldName, targetStr, config.options.replace)
			if newName == oldName || newName == "" {
				return nil
			}

			var targetDir string
			if config.options.output != "" {
				targetDir = config.options.output
			} else {
				targetDir = path
			}
			if config.options.replace != "" {
				newName = resolveConflict(filepath.Dir(targetDir), newName, pairs)
			}
			newPath := filepath.Join(filepath.Dir(targetDir), newName)
			if path == newPath {
				return nil
			}
			pairs[path] = newPath
			return nil
		})
	return pairs, err
}

func copyAction(pairs map[string]string) (uint, error) {
	r, err := ravan.New(ravan.WithWidth(50))
	if err != nil {
		return 0, fmt.Errorf("init raven: %w", err)
	}

	var copied uint
	total := len(pairs)
	for oldName, newName := range pairs {
		if err := copyFile(oldName, newName); err != nil {
			return copied, fmt.Errorf("%q to %q: %w", oldName, newName, err)
		}
		copied++
		r.Draw(float64(copied) / float64(total))
	}
	return copied, nil
}

func moveAction(pairs map[string]string) (uint, error) {
	r, err := ravan.New(ravan.WithWidth(50))
	if err != nil {
		return 0, fmt.Errorf("init raven: %w", err)
	}

	var moved uint
	total := len(pairs)
	for oldName, newName := range pairs {
		if err := moveFile(oldName, newName); err != nil {
			return moved, fmt.Errorf("%q to %q: %w", oldName, newName, err)
		}
		moved++
		r.Draw(float64(moved) / float64(total))
	}
	return moved, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copying data: %w", err)
	}

	if err = out.Sync(); err != nil {
		return fmt.Errorf("sync destination file: %w", err)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get file(%q) info: %w", src, err)
	}
	if err = os.Chmod(dst, info.Mode()); err != nil {
		return fmt.Errorf("failed to set file(%q) permissions: %w", dst, err)
	}

	return nil
}

func moveFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("moving data: %w", err)
	}
	if err = out.Sync(); err != nil {
		return fmt.Errorf("sync destination file: %w", err)
	}
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("get file(%q) info: %w", src, err)
	}
	if err = os.Chmod(dst, info.Mode()); err != nil {
		return fmt.Errorf("set file(%q) permissions: %w", dst, err)
	}
	if err = os.Remove(src); err != nil {
		return fmt.Errorf("remove source file after copy: %w", err)
	}

	return nil
}

func renameAction(pairs map[string]string) (uint, error) {
	r, err := ravan.New(ravan.WithWidth(50))
	if err != nil {
		return 0, fmt.Errorf("init raven: %w", err)
	}

	var renamed uint
	total := len(pairs)
	for oldName, newName := range pairs {
		if err := os.Rename(oldName, newName); err != nil {
			return renamed, fmt.Errorf(
				"%q to %q: %w", oldName, newName, err,
			)
		}
		renamed++
		r.Draw(float64(renamed) / float64(total))
	}

	return renamed, nil
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.options.path, "p", "", "path to dir")
	flag.StringVar(&cfg.options.str, "s", "", "string to find")
	flag.StringVar(&cfg.options.fileType, "t", "", "filter file type to modify")
	flag.StringVar(&cfg.options.replace, "replace", "", "replace str instead of remove it")
	flag.StringVar(&cfg.options.output, "output", "", "copy to new dir instead of rename in path flag dir")
	flag.StringVar(&cfg.options.transmissionType, "tt", "", "determine transmission type. default is copy if output flag is exist.")
	flag.BoolVar(&cfg.withVerbose, "v", false, "verbose")
	flag.BoolVar(&cfg.withDryRun, "d", false, "dry run")
	flag.BoolVar(&cfg.withInteractive, "i", false, "interactive")
	flag.BoolVar(&cfg.withRegex, "r", false, "enable regex")
	flag.BoolVar(&cfg.help, "help", false, "help")
	flag.Parse()
	return cfg
}

func searchString(pattern *regexp.Regexp, str, fileName string) string {
	if pattern == nil {
		return str
	}
	return pattern.FindString(fileName)
}

func canProceed() bool {
	r := bufio.NewReader(os.Stdin)
	s, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func resolveConflict(dir, newName string, pairs map[string]string) string {
	candidate := newName
	count := 1
	for {
		conflict := false
		for _, v := range pairs {
			if filepath.Base(v) == candidate {
				conflict = true
				break
			}
		}
		if _, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
			conflict = true
		}
		if !conflict {
			break
		}
		ext := filepath.Ext(newName)
		nameOnly := strings.TrimSuffix(newName, ext)
		candidate = fmt.Sprintf("%s_%d%s", nameOnly, count, ext)
		count++
	}
	return candidate
}

func getActionName(output, tType string) string {
	tt := getTransmissionType(tType)
	name := RENAME
	if output != "" {
		name = tt
	}
	return name
}

func getTransmissionType(transmissionType string) string {
	switch transmissionType {
	case "mv", "move":
		return MOVE
	default:
		return COPY
	}
}
