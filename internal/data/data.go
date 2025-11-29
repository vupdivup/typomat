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
	batchSize = 100
)

type Token struct {
	Path      string `gorm:"primaryKey"`
	Value     string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type File struct {
	Path        string `gorm:"primaryKey"`
	Fingerprint string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
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

// openDb opens (or creates) a database with the given identifier.
func openDb(id string) (*gorm.DB, error) {
	h := sha256.New()
	h.Write([]byte(id))
	hashedId := hex.EncodeToString(h.Sum(nil))

	db, err := gorm.Open(sqlite.Open(hashedId+".db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Token{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&File{}); err != nil {
		return nil, err
	}
	return db, nil
}
