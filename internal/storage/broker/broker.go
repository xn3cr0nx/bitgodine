package broker

import "context"

// Broker interface defines broker implementation interface
type Broker interface {
	Push(ctx context.Context, key, value []byte) (err error)
	Pull(ctx context.Context) (message interface{}, err error)
}
