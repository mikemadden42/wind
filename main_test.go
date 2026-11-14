package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func createFiles(t *testing.T, dir string, files map[string]int64) {
	t.Helper()
	for filename, size := range files {
		filePath := filepath.Join(dir, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			t.Fatalf("Failed to create parent directory: %v", err)
		}
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := f.Truncate(size); err != nil {
			t.Fatalf("Failed to set file size: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("Failed to close test file: %v", err)
		}
	}
}

// makeUnreadable removes all permissions from dir for the rest of the test.
func makeUnreadable(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permissions cannot make a directory unreadable here")
	}
	if err := os.Chmod(dir, 0); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

func TestDirSize(t *testing.T) {
	dir := t.TempDir()
	createFiles(t, dir, map[string]int64{
		"file1.txt":           1024,      // 1 KB
		"file2.txt":           2048,      // 2 KB
		"nested/file3.txt":    1024 * 10, // 10 KB
		"nested/deep/file4.x": 1,
	})

	var errOut bytes.Buffer
	size, err := dirSize(dir, &errOut)
	if err != nil {
		t.Fatalf("dirSize returned an error: %v", err)
	}

	expectedSize := int64(1024 + 2048 + 1024*10 + 1)
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
	if errOut.Len() != 0 {
		t.Errorf("Expected no errors, got %q", errOut.String())
	}
}

func TestDirSizeMissing(t *testing.T) {
	var errOut bytes.Buffer
	if _, err := dirSize(filepath.Join(t.TempDir(), "missing"), &errOut); err == nil {
		t.Error("Expected an error for a missing directory")
	}
}

func TestDirSizeUnreadableRoot(t *testing.T) {
	dir := t.TempDir()
	makeUnreadable(t, dir)

	var errOut bytes.Buffer
	if _, err := dirSize(dir, &errOut); err == nil {
		t.Error("Expected an error for an unreadable directory")
	}
}

func TestDirSizeUnreadableChild(t *testing.T) {
	dir := t.TempDir()
	createFiles(t, dir, map[string]int64{"file1.txt": 100, "locked/file2.txt": 200})
	makeUnreadable(t, filepath.Join(dir, "locked"))

	var errOut bytes.Buffer
	size, err := dirSize(dir, &errOut)
	if err != nil {
		t.Fatalf("dirSize returned an error: %v", err)
	}
	if size != 100 {
		t.Errorf("Expected size 100, got %d", size)
	}
	if !strings.Contains(errOut.String(), "locked") {
		t.Errorf("Expected an error mentioning the locked directory, got %q", errOut.String())
	}
}

func TestPrintDirectorySize(t *testing.T) {
	tests := []struct {
		size     int64
		unit     string
		expected string
	}{
		{10 * 1024 * 1024, "MB", "10\ttestdir\n"},
		{10 * 1024, "KB", "10\ttestdir\n"},
		{10, "B", "10\ttestdir\n"},
		{1, "MB", "1\ttestdir\n"},
		{1024 + 1, "KB", "2\ttestdir\n"},
		{0, "MB", "0\ttestdir\n"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		printDirectorySize(&buf, "testdir", tt.size, unitSizes[tt.unit])
		if buf.String() != tt.expected {
			t.Errorf("size %d %s: expected %q, got %q", tt.size, tt.unit, tt.expected, buf.String())
		}
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	createFiles(t, dir, map[string]int64{
		"top.txt":        5000, // Files at the top level are not reported
		"a/file.txt":     2048,
		"b/sub/file.txt": 1024,
	})

	var stdout, stderr bytes.Buffer
	if status := run([]string{"-dir", dir, "-unit", "KB"}, &stdout, &stderr); status != 0 {
		t.Fatalf("Expected status 0, got %d (stderr %q)", status, stderr.String())
	}

	expected := "2\ta\n1\tb\n"
	if stdout.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("Expected no errors, got %q", stderr.String())
	}
}

func TestRunInvalidUnit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if status := run([]string{"-dir", t.TempDir(), "-unit", "GB"}, &stdout, &stderr); status != 2 {
		t.Errorf("Expected status 2, got %d", status)
	}
	if stdout.Len() != 0 {
		t.Errorf("Expected no output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "invalid unit") {
		t.Errorf("Expected an invalid unit error, got %q", stderr.String())
	}
}

func TestRunMissingDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if status := run([]string{"-dir", filepath.Join(t.TempDir(), "missing")}, &stdout, &stderr); status != 1 {
		t.Errorf("Expected status 1, got %d", status)
	}
	if stderr.Len() == 0 {
		t.Error("Expected an error message")
	}
}

func TestRunUnreadableSubdir(t *testing.T) {
	dir := t.TempDir()
	createFiles(t, dir, map[string]int64{"ok/file.txt": 10, "locked/file.txt": 10})
	makeUnreadable(t, filepath.Join(dir, "locked"))

	var stdout, stderr bytes.Buffer
	if status := run([]string{"-dir", dir, "-unit", "B"}, &stdout, &stderr); status != 1 {
		t.Errorf("Expected status 1, got %d", status)
	}
	if stdout.String() != "10\tok\n" {
		t.Errorf("Expected only the readable directory, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "locked") {
		t.Errorf("Expected an error mentioning the locked directory, got %q", stderr.String())
	}
}
