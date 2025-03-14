package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// createTempFile is a helper to create a temporary file with the given content.
func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
	return path
}

// TestWalkerWithoutRegex verifies that walker returns a proper mapping when regex is disabled.
func TestWalkerWithoutRegex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testwalker")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create files.
	file1 := createTempFile(t, tempDir, "example_target.txt", "dummy")
	file2 := createTempFile(t, tempDir, "example.txt", "dummy")

	// Create config
	cfg := config{
		options:         fileOptions{path: tempDir, str: "target", fileType: "", replace: ""},
		withVerbose:     false,
		withDryRun:      false,
		withInteractive: false,
		withRegex:       false,
	}

	// Call walker with regex disabled (pattern is nil) and str "target".
	pairs, err := walker(cfg, nil)
	if err != nil {
		t.Fatalf("walker error: %v", err)
	}
	// file1 should be processed because it contains "target".
	if _, ok := pairs[file1]; !ok {
		t.Errorf("expected file %s to be in pairs", file1)
	}
	// file2 should not be processed.
	if _, ok := pairs[file2]; ok {
		t.Errorf("did not expect file %s in pairs", file2)
	}
}

// TestWalkerWithRegex verifies that walker correctly uses a regex pattern.
func TestWalkerWithRegex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testwalker")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create files.
	file1 := createTempFile(t, tempDir, "example_target.txt", "dummy")
	file2 := createTempFile(t, tempDir, "example.txt", "dummy")

	// Compile regex pattern to match "_target" in the file name.
	pattern, err := regexp.Compile("(_target)")
	if err != nil {
		t.Fatalf("failed to compile regex: %v", err)
	}

	// Create config
	cfg := config{
		options:         fileOptions{path: tempDir, str: "(_target)", fileType: "", replace: ""},
		withVerbose:     false,
		withDryRun:      false,
		withInteractive: false,
		withRegex:       true,
	}

	// Here the second parameter "target" is still passed,
	// but the searchString function uses the regex if provided.
	pairs, err := walker(cfg, pattern)
	if err != nil {
		t.Fatalf("walker error: %v", err)
	}

	// file1 should be processed because it matches the regex.
	if _, ok := pairs[file1]; !ok {
		t.Errorf("expected file %s to be in pairs", file1)
	}
	// file2 should not be processed.
	if _, ok := pairs[file2]; ok {
		t.Errorf("did not expect file %s in pairs", file2)
	}

	// Verify that the new file name is as expected.
	expectedNewName := "example.txt" // "example_target.txt" with "_target" removed.
	newPath, ok := pairs[file1]
	if !ok {
		t.Fatalf("file %s not found in pairs", file1)
	}
	if filepath.Base(newPath) != expectedNewName {
		t.Errorf("expected new file name %q, got %q", expectedNewName, filepath.Base(newPath))
	}
}

func TestWalkerWitFileType(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testwalker")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create files.
	file1 := createTempFile(t, tempDir, "file1.txt", "dummy")
	file2 := createTempFile(t, tempDir, "file1.json", "dummy")
	file3 := createTempFile(t, tempDir, "file2.json", "dummy")
	file4 := createTempFile(t, tempDir, "nothing.json", "dummy")

	// Create config
	cfg := config{
		options:         fileOptions{path: tempDir, str: "ile", fileType: ".json", replace: ""},
		withVerbose:     false,
		withDryRun:      false,
		withInteractive: false,
		withRegex:       false,
	}

	// Call walker with regex disabled (pattern is nil) and str "target".
	pairs, err := walker(cfg, nil)
	if err != nil {
		t.Fatalf("walker error: %v", err)
	}
	// file1 should not be processed because it contains ".txt" instead of ".json".
	if _, ok := pairs[file1]; ok {
		t.Errorf("did not expect file %s in pairs", file1)
	}
	// file2 should be processed.because it contains "ile" in file name and ".json" in file extention
	if _, ok := pairs[file2]; !ok {
		t.Errorf("expected file %s to be in pairs", file2)
	}
	// file3 should be processed.because it contains "ile" in file name and ".json" in file extention
	if _, ok := pairs[file3]; !ok {
		t.Errorf("expected file %s to be in pairs", file3)
	}
	// file4 should not be processed.
	if _, ok := pairs[file4]; ok {
		t.Errorf("did not expect file %s in pairs", file4)
	}
}

func TestCollisionResolution(t *testing.T) {
	// Create a temporary directory.
	tempDir, err := os.MkdirTemp("", "collision_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create two files that will produce the same new name when processed.
	// Using regex "d.*d" on these file names:
	// "aaa.json"   -> pattern "aaa" replaced by "bbb" produces "bbb.json"
	// "aaaaaaa.json" -> pattern "aaaaaaa" replaced by "bbb" also produces "bbb.json"
	_ = createTempFile(t, tempDir, "aaa.json", "dummy")
	_ = createTempFile(t, tempDir, "aaaaaaa.json", "dummy")

	// Set up config with regex mode enabled.
	cfg := config{
		options: fileOptions{
			path:    tempDir,
			str:     "a.*a", // Regex pattern to match the entire part from first d to last d.
			replace: "bbb",  // Replacement string.
		},
		withVerbose:     false,
		withDryRun:      false,
		withInteractive: false,
		withRegex:       true,
	}

	// Compile the regex pattern.
	pattern, err := regexp.Compile(cfg.options.str)
	if err != nil {
		t.Fatalf("failed to compile regex: %v", err)
	}

	// Call walker to generate the mapping of old paths to new paths.
	pairs, err := walker(cfg, pattern)
	if err != nil {
		t.Fatalf("walker error: %v", err)
	}

	// We expect both files to be processed.
	if len(pairs) != 2 {
		t.Fatalf("expected 2 files to be processed, got %d", len(pairs))
	}

	// Collect the new file names.
	newNames := make(map[string]bool)
	for _, newPath := range pairs {
		newNames[filepath.Base(newPath)] = true
	}

	// We expect one file to become "bbb.json" and the other to become "bbb.json".
	if !newNames["bbb.json"] {
		t.Errorf("expected 'bbb.json' in new names, got %v", newNames)
	}
	if !newNames["bbb_1.json"] {
		t.Errorf("expected 'bbb_1.json' in new names, got %v", newNames)
	}
}

// TestRename verifies that the rename function renames files as expected.
func TestRename(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testrename")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file that should be renamed.
	originalFile := createTempFile(t, tempDir, "example_target.txt", "dummy")

	// Expected new name after removing "target".
	newName := "example_.txt"
	newPath := filepath.Join(tempDir, newName)
	pairs := map[string]string{
		originalFile: newPath,
	}

	// Call rename.
	count, err := rename(pairs)
	if err != nil {
		t.Fatalf("rename error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 file renamed, got %d", count)
	}

	// Verify that the original file no longer exists and the new file does.
	if _, err := os.Stat(originalFile); !os.IsNotExist(err) {
		t.Errorf("expected original file %s to be removed", originalFile)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("expected new file %s to exist, error: %v", newPath, err)
	}
}

// TestSearchString verifies the behavior of searchString.
func TestSearchString(t *testing.T) {
	// When pattern is nil, it should simply return the str parameter.
	result := searchString(nil, "target", "example_target.txt")
	if result != "target" {
		t.Errorf("expected 'target', got '%s'", result)
	}

	// With regex pattern provided.
	pattern, err := regexp.Compile("target")
	if err != nil {
		t.Fatalf("failed to compile regex: %v", err)
	}
	result = searchString(pattern, "target", "example_target.txt")
	if result != "target" {
		t.Errorf("expected 'target', got '%s'", result)
	}

	// Test with a non-matching pattern.
	pattern, err = regexp.Compile("nomatch")
	if err != nil {
		t.Fatalf("failed to compile regex: %v", err)
	}
	result = searchString(pattern, "nomatch", "example_target.txt")
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

// TestCanProceedYes simulates a "yes" response for canProceed.
func TestCanProceedYes(t *testing.T) {
	// Save original os.Stdin.
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	// Create a pipe and write "y\n".
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.WriteString("y\n")
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	os.Stdin = r

	if !canProceed() {
		t.Error("expected canProceed() to return true for input 'y'")
	}
}

// TestCanProceedNo simulates a "no" response for canProceed.
func TestCanProceedNo(t *testing.T) {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.WriteString("n\n")
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	os.Stdin = r

	if canProceed() {
		t.Error("expected canProceed() to return false for input 'n'")
	}
}

func TestSearchFileExtention(t *testing.T) {
	result := searchFileExtention("file_name.txt")
	if result != ".txt" {
		t.Errorf("expected '.txt', got '%s'", result)
	}
}
