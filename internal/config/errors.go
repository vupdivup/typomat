package config

import "errors"

var (
	// ErrInit indicates a failure to initialize the app configuration.
	ErrInit = errors.New("failed to initialize configuration directory")
	// ErrCleanup indicates a failure to clean up old files.
	ErrCleanup = errors.New("failed to clean configuration directory")
)
