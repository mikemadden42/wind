package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const bytesInMB = 1024 * 1024

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing file: %v\n", err)
			return nil // Continue to next file instead of returning error
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func listDirectories(dir string) ([]os.DirEntry, error) {
	return os.ReadDir(dir)
}

func printDirectorySize(dir string, size int64, unit string) {
	var sizeFormatted int64
	switch unit {
	case "B":
		sizeFormatted = size
	case "KB":
		sizeFormatted = size / 1024
	case "MB":
		sizeFormatted = size / bytesInMB
	default:
		fmt.Println("Invalid unit; using MB as default")
		sizeFormatted = size / bytesInMB
	}
	fmt.Printf("%d\t%s\n", sizeFormatted, dir)
}

func main() {
	dir := flag.String("dir", ".", "Specify the directory to summarize disk usage for")
	unit := flag.String("unit", "MB", "Output unit: B (bytes), KB, or MB")
	flag.Parse()

	entries, err := listDirectories(*dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(*dir, entry.Name())
			size, err := dirSize(subDirPath)
			if err != nil {
				fmt.Printf("Error calculating size for %s: %v\n", subDirPath, err)
				continue
			}
			printDirectorySize(entry.Name(), size, *unit)
		}
	}
}
