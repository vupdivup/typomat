// Package config provides configuration logic and constants for the
// application.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	// AppName is the name of the application.
	AppName = "recital"
	// ProductName is the product name of the application.
	ProductName = "Recital"
)

var (
	appDir  string
	dbDir   string
	logPath string
)

func init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	// Create application directory
	appDir = filepath.Join(configDir, AppName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		panic(err)
	}

	// Create database directory
	dbDir = filepath.Join(appDir, "db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		panic(err)
	}

	// Create logs directory
	logDir := filepath.Join(appDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic(err)
	}

	// Determine log file path
	logFileName := time.Now().Format("20060102_150405") + ".log"
	logPath = filepath.Join(logDir, logFileName)

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
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

// AppDir returns the application directory path.
func AppDir() string {
	return appDir
}

// DbDir returns the directory path where database files are stored.
func DbDir() string {
	return dbDir
}

// LogPath returns the log file path.
func LogPath() string {
	return logPath
}
