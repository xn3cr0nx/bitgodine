package errorx

import (
	"errors"
	"fmt"
	"io"
)

var (
	// ErrNotFound generic not found error
	ErrNotFound = errors.New("not found")
	// ErrInvalidArgument generic invalid argument error
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrUnknown generic unknown error
	ErrUnknown = errors.New("unknown")
	// ErrConfig configuration error
	ErrConfig = errors.New("configuration error")
	// ErrAlreadyExists already exists error
	ErrAlreadyExists = errors.New("already exists")
	// ErrEOF error
	ErrEOF = fmt.Errorf("EOF: %w", io.EOF)
	// ErrOutOfRange index out of range error
	ErrOutOfRange = errors.New("out of range")
)

var (
	// ErrBlockNotFound block not found error
	ErrBlockNotFound = fmt.Errorf("block %w", ErrNotFound)
	// ErrTxNotFound tx not found error
	ErrTxNotFound = fmt.Errorf("tx %w", ErrNotFound)
	// ErrClusterNotFound cluster not found error
	ErrClusterNotFound = fmt.Errorf("cluster %w", ErrNotFound)
)
