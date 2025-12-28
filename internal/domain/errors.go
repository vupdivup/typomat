package domain

import "errors"

var (
	// ErrInvalidDirPath indicates that the provided directory path is invalid.
	ErrInvalidDirPath = errors.New("invalid directory path")
	// ErrFileOperation indicates a failure during file operations.
	ErrFileOperation = errors.New("file operation failed")
	// ErrTokenization indicates a failure during tokenization.
	ErrTextProcessing = errors.New("text processing failed")
	// ErrEmptyDir indicates that no eligible files were found in the specified
	// directory.
	ErrEmptyDir = errors.New("no eligible files found in directory")
	// ErrTooManyErrors indicates that too many errors occurred during directory
	// processing, leading to an abort.
	ErrTooManyErrors = errors.New("too many errors during directory processing")
	// ErrNoTokensFound indicates that no tokens were found after processing the
	// files in the specified directory.
	ErrNoTokensFound = errors.New("no tokens found in directory")
)
