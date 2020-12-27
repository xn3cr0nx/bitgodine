package bitcoin

import "errors"

// ErrEmptySliceParse cannot parse block from empty slice error
var ErrEmptySliceParse = errors.New("cannot parse block from empty slice")

// ErrIncompleteBlockParse cannot parse block due to incomplete file error
var ErrIncompleteBlockParse = errors.New("cannot parse incomplete block")

// ErrBlockParse cannot parse block error
var ErrBlockParse = errors.New("cannot parse block")

// ErrBlockFromBytes cannot generate block from matched bytes error
var ErrBlockFromBytes = errors.New("cannot generate block from matched bytes")

// ErrMagicBytesMatching cannot match magic bytes
var ErrMagicBytesMatching = errors.New("cannot match magic bytes")

// ErrInterrupt interrupt signal error
var ErrInterrupt = errors.New("parser input signal error")
