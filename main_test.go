package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestFile is a helper function that creates a file with given content in the specified directory.
func createTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", name, err)
	}
}

// TestCountRenameCandidates verifies that countRenameCandidates correctly counts the files whose names contain the target substring.
func TestCountRenameCandidates(t *testing.T) {
	// Create a temporary directory for testing.
	tempDir, err := os.MkdirTemp("", "omitter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create sample test files.
	createTestFile(t, tempDir, "file_target.txt", "dummy content")
	createTestFile(t, tempDir, "another_target_file.log", "dummy content")
	createTestFile(t, tempDir, "nochange.txt", "dummy content")

	// Count files that would be renamed.
	count, err := countRenameCandidates(tempDir, "target", false)
	if err != nil {
		t.Fatalf("Error counting rename candidates: %v", err)
	}

	expected := 2
	if count != expected {
		t.Errorf("Expected %d rename candidates, got %d", expected, count)
	}
}

// TestRenameDryRun ensures that when dry-run mode is enabled, files are not actually renamed.
func TestRenameDryRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omitter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalFileName := "example_target.txt"
	createTestFile(t, tempDir, originalFileName, "dummy content")

	// Run rename in dry-run mode.
	count, err := rename(tempDir, "target", true, false)
	if err != nil {
		t.Fatalf("Error in rename dry-run: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 candidate processed in dry-run, got %d", count)
	}

	// Verify the file still exists with its original name.
	if _, err := os.Stat(filepath.Join(tempDir, originalFileName)); os.IsNotExist(err) {
		t.Errorf("Expected file %s to still exist in dry-run mode", originalFileName)
	}
}

// TestRenameActual checks that files are actually renamed when not in dry-run mode.
func TestRenameActual(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omitter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalFileName := "example_target.txt"
	createTestFile(t, tempDir, originalFileName, "dummy content")

	// Run the actual renaming.
	count, err := rename(tempDir, "target", false, false)
	if err != nil {
		t.Fatalf("Error in rename: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 file renamed, got %d", count)
	}

	// Calculate the new file name.
	newFileName := strings.ReplaceAll(originalFileName, "target", "")

	// Confirm that the original file no longer exists.
	if _, err := os.Stat(filepath.Join(tempDir, originalFileName)); !os.IsNotExist(err) {
		t.Errorf("Expected original file %s to be renamed", originalFileName)
	}

	// Confirm that the new file exists.
	if _, err := os.Stat(filepath.Join(tempDir, newFileName)); err != nil {
		t.Errorf("Expected new file %s to exist", newFileName)
	}
}

// TestRenameRegexDryRun ensures that when dry-run mode is enabled in regex mode,
// files are not actually renamed.
func TestRenameRegexDryRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ommiter_regex_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file that should be processed by regex.
	originalFileName := "sample123.txt" // regex "\\d+" will match "123"
	createTestFile(t, tempDir, originalFileName, "dummy content")

	// Run rename in dry-run mode with regex enabled.
	count, err := rename(tempDir, "\\d+", true, true)
	if err != nil {
		t.Fatalf("Error in rename dry-run: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 candidate processed in dry-run, got %d", count)
	}
	// Verify the file still exists with its original name.
	if _, err := os.Stat(filepath.Join(tempDir, originalFileName)); os.IsNotExist(err) {
		t.Errorf("Expected file %s to still exist in dry-run mode", originalFileName)
	}
}

// TestRenameRegexActual checks that files are actually renamed when regex mode is enabled.
func TestRenameRegexActual(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ommiter_regex_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalFileName := "sample123.txt" // regex "\\d+" will match "123"
	createTestFile(t, tempDir, originalFileName, "dummy content")

	// Run the renaming in regex mode (actual renaming).
	count, err := rename(tempDir, "\\d+", false, true)
	if err != nil {
		t.Fatalf("Error in rename: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 file renamed, got %d", count)
	}
	// Calculate the new file name: "sample123.txt" with "123" removed.
	newFileName := strings.ReplaceAll(originalFileName, "123", "")

	// Confirm that the original file no longer exists.
	if _, err := os.Stat(filepath.Join(tempDir, originalFileName)); !os.IsNotExist(err) {
		t.Errorf("Expected original file %s to be renamed", originalFileName)
	}

	// Confirm that the new file exists.
	if _, err := os.Stat(filepath.Join(tempDir, newFileName)); err != nil {
		t.Errorf("Expected new file %s to exist", newFileName)
	}
}
