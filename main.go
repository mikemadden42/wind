// Package main implements wind, a command-line tool that reports the total
// size of each subdirectory of a directory.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var unitSizes = map[string]int64{
	"B":  1,
	"KB": 1024,
	"MB": 1024 * 1024,
}

// dirSize returns the total apparent size of the non-directory entries under
// path. Entries below path that cannot be read are reported to errOut and
// skipped; an error is returned only if path itself cannot be walked.
func dirSize(path string, errOut io.Writer) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == path {
				return err
			}
			fmt.Fprintf(errOut, "wind: %v\n", err)
			return nil // Continue with the rest of the tree
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			fmt.Fprintf(errOut, "wind: %v\n", err)
			return nil
		}
		size += info.Size()
		return nil
	})
	return size, err
}

// printDirectorySize writes size in units of unitSize, rounded up so that a
// non-empty directory never reports 0, matching du.
func printDirectorySize(w io.Writer, dir string, size, unitSize int64) {
	fmt.Fprintf(w, "%d\t%s\n", (size+unitSize-1)/unitSize, dir)
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("wind", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dir := flags.String("dir", ".", "Specify the directory to summarize disk usage for")
	unit := flags.String("unit", "MB", "Output unit: B (bytes), KB, or MB")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	unitSize, ok := unitSizes[*unit]
	if !ok {
		fmt.Fprintf(stderr, "wind: invalid unit %q (want B, KB, or MB)\n", *unit)
		return 2
	}

	entries, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "wind: %v\n", err)
		return 1
	}

	status := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDirPath := filepath.Join(*dir, entry.Name())
		size, err := dirSize(subDirPath, stderr)
		if err != nil {
			fmt.Fprintf(stderr, "wind: %v\n", err)
			status = 1
			continue
		}
		printDirectorySize(stdout, entry.Name(), size, unitSize)
	}
	return status
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
