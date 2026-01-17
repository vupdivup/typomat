// Package data provides functions to interact with the underlying database.
package data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"iter"
	"math"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/glebarez/sqlite"
	"github.com/vupdivup/typomat/internal/config"
	"github.com/vupdivup/typomat/pkg/files"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	// batchSize is the number of records to process in a single batch
	// operation.
	batchSize = 100
)

var (
	// db is the global database connection.
	db *gorm.DB
	// dbPath is the path to the database file.
	dbPath string
	// dirPath is the directory path associated with the database.
	dirPath string

	// ctx is the data-level context for graceful shutdowns.
	ctx context.Context
	// cancel is the cancel function for the data-level context.
	cancel context.CancelFunc
)

// Token represents a token record in the database.
type Token struct {
	// Path is the path to the file from which the token was extracted.
	Path string `gorm:"primaryKey"`
	// Value is the token value.
	Value string `gorm:"primaryKey"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TokenResult represents the result of a token iteration, containing either
// a token or an error.
type TokenResult struct {
	// Token is the token retrieved from the database.
	Token Token
	// Err is any error encountered during retrieval.
	Err error
}

// File represents a file in the user's file system.
type File struct {
	// Path is the path to the file.
	Path string `gorm:"primaryKey"`
	// Size is the size of the file in bytes.
	Size int
	// Mtime is the modification time of the file.
	Mtime time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}

// VersionEquals checks if two File instances refer to the same version of a file.
// Comparison is based on file path, mtime and size.
func (f *File) VersionEquals(other File) bool {
	return f.Path == other.Path && f.Mtime.Unix() == other.Mtime.Unix() &&
		f.Size == other.Size
}

// UpsertTokens inserts or updates the given tokens in a database.
func UpsertTokens(tokens []Token) error {
	result := db.Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(tokens, batchSize)
	if result.Error != nil {
		zap.S().Errorw("Failed to upsert tokens into database",
			"error", result.Error)
		return ErrQuery
	}

	zap.S().Debugw("Upserted tokens into database",
		"token_count", len(tokens),
		"rows_affected", result.RowsAffected)
	return nil
}

// UpsertFiles uploads or updates file records in a database.
func UpsertFiles(files []File) error {
	result := db.Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(files, batchSize)
	if result.Error != nil {
		zap.S().Errorw("Failed to upsert files into database",
			"error", result.Error)
		return ErrQuery
	}

	zap.S().Debugw("Upserted files into database",
		"file_count", len(files),
		"rows_affected", result.RowsAffected)
	return nil
}

// DeleteFile removes a file record from the database, optionally cascading
// the deletion to associated tokens.
func DeleteFile(file File, cascade bool) error {
	// Delete the file record
	if err := db.Delete(&file).Error; err != nil {
		zap.S().Errorw("Failed to delete file from database",
			"file_path", file.Path,
			"error", err)
		return ErrQuery
	}

	if !cascade {
		zap.S().Debugw("Deleted file from database",
			"file_path", file.Path)
		zap.S().Debugw("Skipping cascade delete of associated tokens",
			"file_path", file.Path)
		return nil
	}

	// Cascade delete associated tokens
	if err := db.
		Where("path = ?", file.Path).
		Delete(&Token{}).Error; err != nil {
		zap.S().Errorw(
			"Failed to cascade delete tokens from database",
			"file_path", file.Path,
			"error", err)
		return ErrQuery
	}

	zap.S().Debugw(
		"Deleted file and associated tokens from database",
		"file_path", file.Path)
	return nil
}

// DeleteTokensOfFile removes all tokens associated with a specific file
// from the database.
func DeleteTokensOfFile(path string) error {
	// Delete associated tokens
	if err := db.
		Where("path = ?", path).
		Delete(&Token{}).Error; err != nil {
		zap.S().Errorw(
			"Failed to delete tokens of file from database",
			"file_path", path,
			"error", err)
		return ErrQuery
	}

	zap.S().Debugw("Deleted tokens of file from database",
		"file_path", path)

	return nil
}

// IterUniqueTokens returns an iterator over distinct tokens in the database.
func IterUniqueTokens() iter.Seq[TokenResult] {
	return func(yield func(TokenResult) bool) {
		// Query distinct tokens
		rows, err := db.Model(&Token{}).Distinct("value").Rows()
		if err != nil {
			zap.S().Errorw("Failed to query distinct tokens",
				"error", err)
			yield(TokenResult{Err: ErrQuery})
			return
		}
		defer rows.Close()

		// Iterate over the result set
		for rows.Next() {
			var token Token
			if err := db.ScanRows(rows, &token); err != nil {
				zap.S().Errorw("Failed to scan token row",
					"error", err)
				yield(TokenResult{Err: ErrQuery})
				return
			}
			if !yield(TokenResult{Token: token}) {
				return
			}
		}
	}
}

// GetFiles retrieves all file records from the database.
func GetFiles() ([]File, error) {
	var files []File
	if err := db.Find(&files).Error; err != nil {
		zap.S().Errorw("Failed to retrieve files from database",
			"error", err)
		return []File{}, ErrQuery
	}

	zap.S().Debugw("Retrieved files from database",
		"file_count", len(files))
	return files, nil
}

// Setup initializes the database connection for the specified directory.
// If a cached database already exists for the directory, it will be used.
// Alternatively, if useCache is true, the cached database will be used or created.
func Setup(dirPath string, useCache bool) error {
	// Hash the ID to create a filename
	h := sha256.New()
	h.Write([]byte(dirPath))
	hashedId := hex.EncodeToString(h.Sum(nil))

	// Check if the database was cached on a previous run
	cachedDbPath := filepath.Join(config.CachedDbDir(), hashedId+".db")
	cacheExists, err := files.FileExists(cachedDbPath)
	if err != nil {
		zap.S().Errorw("Failed to check cached database existence",
			"db_id", dirPath,
			"db_path", cachedDbPath,
			"error", err)
		return ErrConn
	}

	// If cache is enabled or a cached database exists, use it
	if cacheExists || useCache {
		zap.S().Debugw("Using cached database",
			"db_id", dirPath,
			"db_path", cachedDbPath)
		dbPath = cachedDbPath
	} else {
		dbPath = filepath.Join(config.TempDbDir(), hashedId+".db")
	}

	// Open (or create) the SQLite database
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		zap.S().Errorw("Failed to open database",
			"db_id", dirPath,
			"db_path", dbPath,
			"error", err)
		return ErrConn
	}
	db = db.WithContext(ctx)
	zap.S().Infow("Opened database",
		"db_id", dirPath,
		"db_path", dbPath)

	// Perform migrations
	if err := db.AutoMigrate(&File{}, &Token{}); err != nil {
		zap.S().Errorw("Failed to migrate or create database schema",
			"db_id", dirPath,
			"error", err)
		return ErrQuery
	}

	return nil
}

// Teardown closes the database connection and cleans up temporary files.
func Teardown() error {
	// Cancel any ongoing operations
	cancel()
	
	if db == nil {
		return nil
	}


	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Errorw("Failed to get sql.DB from gorm.DB during teardown",
			"error", err)
		return ErrCleanup
	}
	if err := sqlDB.Close(); err != nil {
		zap.S().Errorw("Failed to close sql.DB during teardown",
			"error", err)
		return ErrCleanup
	}

	// Attempt to delete temporary database files with retries
	baseTimeout := time.Millisecond * 50
	deleteRetries := 4
	for i := range deleteRetries {
		err := files.RemoveChildren(config.TempDbDir())
		if err == nil {
			zap.S().Infow("Removed temporary database files during teardown",
				"attempt", i+1)
			break
		}

		if i < deleteRetries-1 {
			zap.S().Warnw("Retrying removal of temporary database files during teardown",
				"error", err)
			time.Sleep(time.Duration(math.Exp2(float64(i))) * baseTimeout)
		} else {
			zap.S().Errorw("Failed to remove temporary database files during teardown",
				"error", err)
			return ErrCleanup
		}
	}

	return nil
}
