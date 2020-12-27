package errorx

import "fmt"

// ErrKeyNotFound key not found error
var ErrKeyNotFound = fmt.Errorf("key %w", ErrNotFound)
