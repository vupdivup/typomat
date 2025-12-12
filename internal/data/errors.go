package data

import "errors"

var (
	// ErrDbConnection indicates a failure to connect to the database.
	ErrDbConnection = errors.New("failed to connect to database")
	// ErrDbOperation indicates a failure during a database operation.
	ErrDbOperation = errors.New("database operation failed")
)
