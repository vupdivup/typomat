package data

import "errors"

var (
	// ErrConn indicates a failure to connect to the database.
	ErrConn = errors.New("failed to connect to database")
	// ErrQuery indicates a failure during a database operation.
	ErrQuery = errors.New("database operation failed")
)
