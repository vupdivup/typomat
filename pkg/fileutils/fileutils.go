package fileutils

import (
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// IsTextFile checks if the file at the given path is a text file by reading a
// portion of its content.
func IsTextFile(path string) (bool, error) {
	// Number of bytes to read for checking text validity
	lookahead := 512

	file, err := os.Open(path)

	if err != nil {
		return false, err
	}

	defer file.Close()

	// Read a portion of the file to check if it's valid UTF-8
	buf := make([]byte, lookahead)

	if _, err := file.Read(buf); err == io.EOF {
		// An empty file is considered not a text file
		return false, nil
	} else if err != nil {
		return false, err
	}

	return utf8.Valid(buf), nil
}

// GetFilesInTree returns a list of all files in the directory tree rooted at
// the specified path.
// Files are returned as their absolute paths.
func GetFilesInTree(root string) ([]string, error) {
	var files []string

	walk := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	}

	err := filepath.WalkDir(root, walk)
	if err != nil {
		return nil, err
	}

	return files, nil
}
