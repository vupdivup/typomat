// Package config provides configuration logic and constants for the
// application.
package config

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	// AppName is the name of the application.
	AppName = "recital"
)

var (
	appDir  string
	dbDir   string
	logPath string
	logFile *os.File
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

	logFileName := time.Now().Format("20060102_150405") + ".log"
	logPath = filepath.Join(logDir, logFileName)
	logFile, err = os.OpenFile(
		logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
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

// Cleanup performs cleanup operations such as closing log files.
func Cleanup() {
	log.Println("Cleaning up resources")

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			log.Fatalf("Error closing log file: %v", err)
		}
	}
}
