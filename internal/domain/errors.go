package domain

import "errors"

var (
	// ErrInvalidDirPath indicates that the provided directory path is invalid.
	ErrInvalidDirPath = errors.New("invalid directory path")
	// ErrFileOperation indicates a failure during file operations.
	ErrFileOperation = errors.New("file operation failed")
	// ErrTokenization indicates a failure during tokenization.
	ErrTextProcessing = errors.New("text processing failed")
)
