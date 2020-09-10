package redis

import (
	"errors"

	"github.com/go-redis/redis/v8"

	ctx "context"
)

// Redis client wrapper
type Redis struct {
	*redis.Client
}

// Config strcut containing initialization fields
type Config struct {
	URL string
}

// Conf returns default config struct
func Conf(URL string) *Config {
	host := "redis:2379"
	if URL != "" {
		host = URL
	}

	return &Config{
		URL: host,
	}
}

func errorParser(err error) error {
	if err == nil {
		return err
	} else if err.Error() == "redis: nil" {
		err = nil
	}
	return err
}

// NewRedis creates a new instance of the db
func NewRedis(conf *Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: conf.URL,
		// Password: "", // no password set
		DB: 0, // use default DB
	})

	_, err := rdb.Ping(ctx.Background()).Result()

	return &Redis{rdb}, err
}

// Store insert new key-value in redis
func (r *Redis) Store(key string, value []byte) (err error) {
	e := r.Set(ctx.Background(), key, value, 0).Err()
	return errorParser(e)
}

// StoreBatch insert new key-value in redis
func (r *Redis) StoreBatch(batch interface{}) (err error) {
	series := batch.(map[string][]byte)
	for key, value := range series {
		if err = r.Store(key, value); err != nil {
			return
		}
	}
	return
}

// Queue data structure used as broker to load insertion queue
type Queue []interface{}

var queue Queue

// StoreQueueBatch loads a queue until a threshold to perform a bulk insertion
func (r *Redis) StoreQueueBatch(v interface{}) (err error) {
	if queue == nil {
		queue = make(Queue, 0)
	}
	queue = append(queue, v)
	if len(queue) >= 100 {
		r.StoreBatch(queue)
		queue = make(Queue, 0)
	}
	return
}

func (r *Redis) Read(key string) (value []byte, err error) {
	val, err := r.Get(ctx.Background(), key).Result()
	err = errorParser(err)
	if err != nil {
		return
	}
	if val == "" {
		err = errors.New("Key not found")
	}
	value = []byte(val)
	return
}

// ReadKeys concurrent read values based on a prefix
func (r *Redis) ReadKeys() (value []string, err error) {
	c := ctx.Background()
	iter := r.Scan(c, 0, "", 0).Iterator()
	for iter.Next(c) {
		value = append(value, iter.Val())
	}
	err = errorParser(iter.Err())
	return
}

// ReadKeyValues concurrent read values based on a prefix
func (r *Redis) ReadKeyValues() (value map[string][]byte, err error) {
	c := ctx.Background()
	iter := r.Scan(c, 0, "", 0).Iterator()
	for iter.Next(c) {
		val, err := r.Read(iter.Val())
		if err != nil {
			return nil, err
		}
		value[iter.Val()] = val
	}
	err = errorParser(iter.Err())
	return
}

// ReadKeysWithPrefix concurrent read values based on a prefix
func (r *Redis) ReadKeysWithPrefix(prefix string) (keys []string, err error) {
	c := ctx.Background()
	iter := r.Scan(c, 0, prefix, 0).Iterator()
	for iter.Next(c) {
		keys = append(keys, iter.Val())
	}
	err = iter.Err()
	return
}

// ReadPrefix concurrent read values based on a prefix
func (r *Redis) ReadPrefix(prefix string) (value [][]byte, err error) {
	c := ctx.Background()
	iter := r.Scan(c, 0, prefix, 0).Iterator()
	for iter.Next(c) {
		val, err := r.Read(iter.Val())
		if err != nil {
			return nil, err
		}
		value = append(value, val)
	}
	err = errorParser(iter.Err())
	return
}

// ReadFirstValueByPrefix returns the first value matched by prefix
func (r *Redis) ReadFirstValueByPrefix(prefix string) (value []byte, err error) {
	c := ctx.Background()
	iter := r.Scan(c, 0, prefix, 0).Iterator()
	for iter.Next(c) {
		val, err := r.Read(iter.Val())
		if err != nil {
			return nil, err
		}
		return val, err
	}
	err = errorParser(iter.Err())
	return
}

// ReadPrefixWithKey concurrent read values based on a prefix
func (r *Redis) ReadPrefixWithKey(prefix string) (value map[string][]byte, err error) {
	c := ctx.Background()
	iter := r.Scan(c, 0, prefix, 0).Iterator()
	for iter.Next(c) {
		val, err := r.Read(iter.Val())
		if err != nil {
			return nil, err
		}
		value[iter.Val()] = val
	}
	err = errorParser(iter.Err())
	return
}

// IsStored returns true if the block corresponding to passed hash is stored in db
func (r *Redis) IsStored(key string) bool {
	_, err := r.Read(key)
	return err == nil
}

// Delete inserts in the db the block as []byte passed
func (r *Redis) Delete(key string) (err error) {
	err = r.Del(ctx.Background(), key).Err()
	return errorParser(err)
}

// Empty empties the redis store
func (r *Redis) Empty() (err error) {
	keys, err := r.ReadKeys()
	if err != nil {
		return
	}

	for _, key := range keys {
		if err = r.Delete(key); err != nil {
			return
		}
	}

	return
}
