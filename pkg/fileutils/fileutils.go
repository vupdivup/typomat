package fileutils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"
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

// Size returns the size of the file at the given path in bytes.
func Size(path string) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(info.Size()), nil
}

// Mtime returns the modification time of the file at the given path.
func Mtime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// GetFingerprint computes the SHA-256 fingerprint of the file at the given
// path.
func GetFingerprint(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
