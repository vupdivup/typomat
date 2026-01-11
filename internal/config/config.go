// Package config provides configuration logic and constants for the
// application.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vupdivup/typomat/pkg/files"
	"go.uber.org/zap"
)

const (
	// ProductName is the human-readable name of the application.
	ProductName = "typomat"
	// AppCommandName is the command-line name of the application.
	AppName = "typomat"
)

var (
	appDir string
	dbDir  string
)

// Init initializes the configuration by setting up necessary directories
// and configuring the logger.
func Init() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return ErrInit
	}

	// Create application directory
	appDir = filepath.Join(cacheDir, AppName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return ErrInit
	}

	// Create database directory
	dbDir = filepath.Join(appDir, "db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return ErrInit
	}

	// Create logs directory
	logDir := filepath.Join(appDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return ErrInit
	}

	// Determine log file path
	logFileName := time.Now().Format("20060102_150405") + ".log"
	logPath := filepath.Join(logDir, logFileName)

	// Initialize zap logger
	var config zap.Config
	if os.Getenv(strings.ToUpper(AppName)+"_DEBUG") == "" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.OutputPaths = []string{logPath}
	logger, err := config.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

// AppDir returns the application directory path.
func AppDir() string {
	return appDir
}

// DbDir returns the directory path where database files are stored.
func DbDir() string {
	return dbDir
}

// PurgeCache deletes all cached data stored in the database directory.
func PurgeCache() error {
	zap.S().Infow("Purging application cache",
		"db_dir", dbDir)
	return files.RemoveChildren(dbDir)
}
