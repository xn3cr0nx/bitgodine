package bitcoin

import (
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
)

var (
	// ErrEmptySliceParse cannot parse block from empty slice error
	ErrEmptySliceParse = errors.New("cannot parse block from empty slice")

	// ErrIncompleteBlockParse cannot parse block due to incomplete file error
	ErrIncompleteBlockParse = errors.New("cannot parse incomplete block")

	// ErrBlockParse cannot parse block error
	ErrBlockParse = errors.New("cannot parse block")

	// ErrBlockFromBytes cannot generate block from matched bytes error
	ErrBlockFromBytes = errors.New("cannot generate block from matched bytes")

	// ErrMagicBytesMatching cannot match magic bytes
	ErrMagicBytesMatching = errors.New("cannot match magic bytes")
)

var (
	// ErrInterrupt interrupt signal error
	ErrInterrupt = errors.New("parser input signal error")

	// ErrInterruptUnknown interrupt signal error
	ErrInterruptUnknown = errors.New("parser input signal unknown error")
)

var (
	// ErrExceededSize skipped blocks size error
	ErrExceededSize = errors.New("exceed skipped blocks size")

	// ErrCheckpointNotFound checkpoint not found error
	ErrCheckpointNotFound = fmt.Errorf("checkpoint %w", errorx.ErrNotFound)

	// ErrNoBitcoinData returned if no bitcoin data to read from
	ErrNoBitcoinData = errors.New("missing bitcoin data")
)
