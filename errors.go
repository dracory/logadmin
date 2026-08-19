package logadmin

import "errors"

// Common errors
var (
	// ErrStoreRequired is returned when Store is not provided
	ErrStoreRequired = errors.New("log store is required")

	// ErrLoggerRequired is returned when Logger is not provided
	ErrLoggerRequired = errors.New("logger is required")
)
