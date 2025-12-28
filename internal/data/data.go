// Package data provides functions to interact with the underlying database.
package data

import (
	"crypto/sha256"
	"encoding/hex"
	"iter"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/glebarez/sqlite"
	"github.com/vupdivup/typomat/internal/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	// batchSize is the number of records to process in a single batch
	// operation.
	batchSize = 100
)

// dbCache is a map of opened database connections keyed by their IDs.
var dbCache = map[string]*gorm.DB{}

// Token represents a token record in the database.
type Token struct {
	// Path is the path to the file from which the token was extracted.
	Path string `gorm:"primaryKey"`
	// Value is the token value.
	Value string `gorm:"primaryKey"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
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
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// VersionEquals checks if two File instances refer to the same version of a file.
// Comparison is based on file path, mtime and size.
func (f *File) VersionEquals(other File) bool {
	return f.Path == other.Path && f.Mtime.Unix() == other.Mtime.Unix() &&
		f.Size == other.Size
}

// UpsertTokens inserts or updates the given tokens in a database.
func UpsertTokens(dbId string, tokens []Token) error {
	db, err := openDb(dbId)
	if err != nil {
		return err
	}

	result := db.Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(tokens, batchSize)
	if result.Error != nil {
		zap.S().Errorw("Failed to upsert tokens into database",
			"db_id", dbId,
			"error", result.Error)
		return ErrQuery
	}

	zap.S().Debugw("Upserted tokens into database",
		"token_count", len(tokens),
		"rows_affected", result.RowsAffected,
		"db_id", dbId)
	return nil
}

// UpsertFiles uploads or updates file records in a database.
func UpsertFiles(dbId string, files []File) error {
	db, err := openDb(dbId)
	if err != nil {
		return err
	}

	result := db.Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(files, batchSize)
	if result.Error != nil {
		zap.S().Errorw("Failed to upsert files into database",
			"db_id", dbId,
			"error", result.Error)
		return ErrQuery
	}

	zap.S().Debugw("Upserted files into database",
		"file_count", len(files),
		"rows_affected", result.RowsAffected,
		"db_id", dbId)
	return nil
}

// DeleteFile removes a file record from the database, optionally cascading
// the deletion to associated tokens.
func DeleteFile(dbId string, file File, cascade bool) error {
	// Open the database
	db, err := openDb(dbId)
	if err != nil {
		return err
	}

	// Delete the file record
	if err := db.Delete(&file).Error; err != nil {
		zap.S().Errorw("Failed to delete file from database",
			"file_path", file.Path,
			"db_id", dbId,
			"error", err)
		return ErrQuery
	}

	if !cascade {
		zap.S().Debugw("Deleted file from database",
			"file_path", file.Path,
			"db_id", dbId)
		zap.S().Debugw("Skipping cascade delete of associated tokens",
			"file_path", file.Path,
			"db_id", dbId)
		return nil
	}

	// Cascade delete associated tokens
	if err := db.
		Where("path = ?", file.Path).
		Delete(&Token{}).Error; err != nil {
		zap.S().Errorw(
			"Failed to cascade delete tokens from database",
			"file_path", file.Path,
			"db_id", dbId,
			"error", err)
		return ErrQuery
	}

	zap.S().Debugw(
		"Deleted file and associated tokens from database",
		"file_path", file.Path,
		"db_id", dbId)
	return nil
}

// DeleteTokensOfFile removes all tokens associated with a specific file
// from the database.
func DeleteTokensOfFile(dbId string, path string) error {
	// Open the database
	db, err := openDb(dbId)
	if err != nil {
		return err
	}

	// Delete associated tokens
	if err := db.
		Where("path = ?", path).
		Delete(&Token{}).Error; err != nil {
		zap.S().Errorw(
			"Failed to delete tokens of file from database",
			"file_path", path,
			"db_id", dbId,
			"error", err)
		return ErrQuery
	}

	zap.S().Debugw("Deleted tokens of file from database",
		"file_path", path,
		"db_id", dbId)

	return nil
}

// IterUniqueTokens returns an iterator over distinct tokens in the database.
func IterUniqueTokens(dbId string) iter.Seq[TokenResult] {
	zap.S().Debugw("Creating iterator for unique tokens",
		"db_id", dbId)

	return func(yield func(TokenResult) bool) {
		// Open the database
		db, err := openDb(dbId)
		if err != nil {
			yield(TokenResult{Err: err})
			return
		}

		// Query distinct tokens
		rows, err := db.Model(&Token{}).Distinct("value").Rows()
		if err != nil {
			zap.S().Errorw("Failed to query distinct tokens",
				"db_id", dbId,
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
					"db_id", dbId,
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
func GetFiles(dbId string) ([]File, error) {
	db, err := openDb(dbId)
	if err != nil {
		return []File{}, err
	}

	var files []File
	if err := db.Find(&files).Error; err != nil {
		zap.S().Errorw("Failed to retrieve files from database",
			"db_id", dbId,
			"error", err)
		return []File{}, ErrQuery
	}

	zap.S().Debugw("Retrieved files from database",
		"file_count", len(files),
		"db_id", dbId)
	return files, nil
}

// openDb opens (or creates) a database with the given identifier.
// Cached database connections are reused.
func openDb(id string) (*gorm.DB, error) {
	// Check if the database is already opened and cached
	if db, ok := dbCache[id]; ok {
		return db, nil
	}

	// Hash the ID to create a filename
	h := sha256.New()
	h.Write([]byte(id))
	hashedId := hex.EncodeToString(h.Sum(nil))

	// Determine the database file path
	dbPath := filepath.Join(config.DbDir(), hashedId+".db")

	// Open (or create) the SQLite database
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		zap.S().Errorw("Failed to open database",
			"db_id", id,
			"db_path", dbPath,
			"error", err)
		return nil, ErrConn
	}
	zap.S().Debugw("Opened database",
		"db_id", id,
		"db_path", dbPath)

	// Perform migrations
	// Retry once if migration fails
	for i := range 2 {
		if err := db.AutoMigrate(&File{}, &Token{}); err == nil {
			break
		}

		zap.S().Errorw("Failed to migrate or create database schema",
			"db_id", id,
			"error", err)

		if i == 1 {
			return nil, ErrQuery
		}

		// Try to purge cache and reopen
		if err := config.PurgeCache(); err != nil {
			zap.S().Errorw("Failed to purge cache after migration error",
				"db_id", id,
				"error", err)
		}
	}

	// Cache the opened database
	dbCache[id] = db

	return db, nil
}
