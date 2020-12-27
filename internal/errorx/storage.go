package errorx

import (
	"errors"
	"fmt"
)

var (
	// ErrKeyNotFound key not found error
	ErrKeyNotFound = fmt.Errorf("key %w", ErrNotFound)

	// ErrCache error setting cache key
	ErrCache = errors.New("error caching")
)
