package storage

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
)

// // KeyNotFoundError error when redis key not found
// type KeyNotFoundError error

// // NewKeyNotFoundError returns a new KeyNotFoundError error
// func NewKeyNotFoundError(key string) KeyNotFoundError {
// 	return fmt.Errorf("%s not found", key)
// }

// ErrKeyNotFound key not found error
var ErrKeyNotFound = fmt.Errorf("key %w", errorx.ErrNotFound)
