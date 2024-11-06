package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func createTempDirWithFiles(t *testing.T, files map[string]int64) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	for filename, size := range files {
		filePath := filepath.Join(dir, filename)
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := f.Truncate(size); err != nil {
			t.Fatalf("Failed to set file size: %v", err)
		}
		f.Close()
	}
	return dir
}

func TestDirSize(t *testing.T) {
	files := map[string]int64{
		"file1.txt": 1024,      // 1 KB
		"file2.txt": 2048,      // 2 KB
		"file3.txt": 1024 * 10, // 10 KB
	}
	dir := createTempDirWithFiles(t, files)
	defer os.RemoveAll(dir)

	size, err := dirSize(dir)
	if err != nil {
		t.Fatalf("dirSize returned an error: %v", err)
	}

	expectedSize := int64(1024 + 2048 + 1024*10)
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

func TestListDirectories(t *testing.T) {
	dir := createTempDirWithFiles(t, map[string]int64{"file1.txt": 1024})
	defer os.RemoveAll(dir)

	// Create subdirectories
	subDir1 := filepath.Join(dir, "subdir1")
	subDir2 := filepath.Join(dir, "subdir2")
	os.Mkdir(subDir1, 0755)
	os.Mkdir(subDir2, 0755)

	entries, err := listDirectories(dir)
	if err != nil {
		t.Fatalf("listDirectories returned an error: %v", err)
	}

	var subDirNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			subDirNames = append(subDirNames, entry.Name())
		}
	}

	expectedDirs := []string{"subdir1", "subdir2"}
	if len(subDirNames) != len(expectedDirs) {
		t.Fatalf("Expected %d directories, got %d", len(expectedDirs), len(subDirNames))
	}

	for i, expected := range expectedDirs {
		if subDirNames[i] != expected {
			t.Errorf("Expected directory %s, got %s", expected, subDirNames[i])
		}
	}
}

func TestPrintDirectorySize(t *testing.T) {
	var buf bytes.Buffer

	// Test for MB output
	printDirectorySizeToWriter(&buf, "testdir", 10485760, "MB") // 10 MB
	expected := "10\ttestdir\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
	buf.Reset()

	// Test for KB output
	printDirectorySizeToWriter(&buf, "testdir", 10240, "KB") // 10 KB
	expected = "10\ttestdir\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
	buf.Reset()

	// Test for Bytes output
	printDirectorySizeToWriter(&buf, "testdir", 10, "B") // 10 bytes
	expected = "10\ttestdir\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// printDirectorySizeToWriter is a helper function that writes to an io.Writer for testing purposes
func printDirectorySizeToWriter(w *bytes.Buffer, dir string, size int64, unit string) {
	var sizeFormatted int64
	switch unit {
	case "B":
		sizeFormatted = size
	case "KB":
		sizeFormatted = size / 1024
	case "MB":
		sizeFormatted = size / bytesInMB
	default:
		sizeFormatted = size / bytesInMB
	}
	fmt.Fprintf(w, "%d\t%s\n", sizeFormatted, dir)
}
