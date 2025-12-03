// Package data provides functions to interact with the underlying database.
package data

import (
	"crypto/sha256"
	"encoding/hex"
	"iter"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// File represents a file in the user's file system.
type File struct {
	// Path is the path to the file.
	Path string `gorm:"primaryKey"`
	// Fingerprint is the file's hash fingerprint.
	Fingerprint string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// UpsertTokens inserts or updates the given tokens in a database.
func UpsertTokens(dbId string, tokens []Token) error {
	db, err := openDb(dbId)
	if err != nil {
		return err
	}

	return db.Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(tokens, batchSize).Error
}

// UpsertFiles uploads or updates file records in a database.
func UpsertFiles(dbId string, files []File) error {
	db, err := openDb(dbId)
	if err != nil {
		return err
	}

	return db.Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(files, batchSize).Error
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
	if err := db.Delete(&file).Error; !cascade || err != nil {
		return err
	}

	// Cascade delete associated tokens
	if err := db.
		Where("path = ?", file.Path).
		Delete(&Token{}).Error; err != nil {
		return err
	}

	return nil
}

// GetTokens returns an iterator over distinct tokens in the database.
func GetTokens(dbId string) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		db, err := openDb(dbId)
		if err != nil {
			yield(Token{}, err)
			return
		}

		rows, err := db.Model(&Token{}).Distinct("value").Rows()
		if err != nil {
			yield(Token{}, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var token Token
			if err := db.ScanRows(rows, &token); err != nil {
				yield(Token{}, err)
				return
			}
			if !yield(token, nil) {
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
	result := db.Find(&files)
	return files, result.Error
}

// openDb opens (or creates) a database with the given identifier.
// Cached database connections are reused.
func openDb(id string) (*gorm.DB, error) {
	if db, ok := dbCache[id]; ok {
		return db, nil
	}

	// Hash the ID to create a filename
	h := sha256.New()
	h.Write([]byte(id))
	hashedId := hex.EncodeToString(h.Sum(nil))

	// Open (or create) the SQLite database
	db, err := gorm.Open(sqlite.Open(hashedId+".db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Perform migrations
	if err := db.AutoMigrate(&Token{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&File{}); err != nil {
		return nil, err
	}

	// Cache the opened database
	dbCache[id] = db

	return db, nil
}
