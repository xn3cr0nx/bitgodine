package errorx

import "errors"

var (
	// ErrNotFound generic not found error
	ErrNotFound = errors.New("not found")
	// ErrInvalidArgument generic invalid argument error
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrUnknown generic unknown error
	ErrUnknown = errors.New("unknown")
)
