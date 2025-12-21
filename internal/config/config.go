// Package config provides configuration logic and constants for the
// application.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vupdivup/typelines/pkg/files"
	"go.uber.org/zap"
)

const (
	// AppName is the name of the application.
	AppName = "Keycap"

	// retentionPeriod defines how long db and log files are kept before
	// being cleaned up.
	retentionPeriod = 7 * 24 * time.Hour
)

var (
	appDir string
	dbDir  string
)

// Init initializes the configuration by setting up necessary directories
// and configuring the logger.
func Init() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ErrInit
	}

	// Create application directory
	appDir = filepath.Join(configDir, AppName)
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
	zap.S().Info("Purging application cache")
	return files.RemoveChildren(dbDir)
}

// RemoveOldFiles removes files in the application directory that are older
// than a week.
func RemoveOldFiles() error {
	err := filepath.WalkDir(
		appDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				zap.S().Errorw("Failed to access path",
					"path", path,
					"error", err)
				return nil // Continue walking despite errors
			}

			// Skip directories
			if d.IsDir() {
				return nil
			}

			stat, err := d.Info()
			if err != nil {
				zap.S().Errorw("Failed to get file info",
					"path", path,
					"error", err)
				return nil
			}

			// Remove file if older than retention period
			if time.Since(stat.ModTime()) > retentionPeriod {
				err := os.Remove(path)
				if err == nil {
					zap.S().Infow("Removed old file",
						"path", path)
				} else {
					zap.S().Errorw("Failed to remove file",
						"path", path,
						"error", err)
				}
			}
			return nil
		})
	if err != nil {
		zap.S().Errorw("Failed to walk application directory",
			"app_dir", appDir,
			"error", err)
		return ErrCleanup
	}

	return nil
}
