package tx

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
)

// ErrTxNotFound wraps the key not found error in a transaction not found
var ErrTxNotFound = fmt.Errorf("tx %w", errorx.ErrKeyNotFound)
