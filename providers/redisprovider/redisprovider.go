package redisprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/miqdadyyy/flaggo"
)

// Options holds configuration for the Redis provider.
type Options struct {
	// Addr is the Redis connection URL (e.g., "redis://localhost:6379/0").
	Addr string
	// Database is the Redis database number (0-15).
	Database int
	// Prefix is prepended to all keys (e.g., "flaggo:").
	Prefix string
}

// RedisProvider implements flaggo.Provider using Redis.
type RedisProvider struct {
	pool   *redis.Pool
	db     int
	prefix string
}

// New creates a RedisProvider and verifies the connection.
func New(opts Options) (*RedisProvider, error) {
	pool := &redis.Pool{
		MaxIdle:     3,
		MaxActive:   10,
		IdleTimeout: 10 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(opts.Addr, redis.DialConnectTimeout(10*time.Second))
		},
	}

	conn := pool.Get()
	defer func() { _ = conn.Close() }()

	if _, err := conn.Do("PING"); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("flaggo/redisprovider: ping: %w", err)
	}

	return &RedisProvider{
		pool:   pool,
		db:     opts.Database,
		prefix: opts.Prefix,
	}, nil
}

func (p *RedisProvider) conn() redis.Conn {
	c := p.pool.Get()
	_, _ = c.Do("SELECT", p.db)
	return c
}

// Set stores a flag configuration as JSON in Redis.
func (p *RedisProvider) Set(ctx context.Context, key string, config flaggo.FlagConfig) error {
	c := p.conn()
	defer func() { _ = c.Close() }()

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("flaggo/redisprovider: marshal: %w", err)
	}

	_, err = c.Do("SET", p.prefix+key, data)
	return err
}

// Get returns the configuration for a flag and whether it exists.
func (p *RedisProvider) Get(ctx context.Context, key string) (flaggo.FlagConfig, bool) {
	c := p.conn()
	defer func() { _ = c.Close() }()

	data, err := redis.Bytes(c.Do("GET", p.prefix+key))
	if err != nil {
		return flaggo.FlagConfig{}, false
	}

	var cfg flaggo.FlagConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return flaggo.FlagConfig{}, false
	}
	return cfg, true
}

// All returns all flags matching the provider's prefix.
func (p *RedisProvider) All(ctx context.Context) (map[string]flaggo.FlagConfig, error) {
	c := p.conn()
	defer func() { _ = c.Close() }()

	flags := make(map[string]flaggo.FlagConfig)
	cursor := 0

	for {
		reply, err := redis.Values(c.Do("SCAN", cursor, "MATCH", p.prefix+"*", "COUNT", 100))
		if err != nil {
			return nil, fmt.Errorf("flaggo/redisprovider: scan: %w", err)
		}

		if len(reply) != 2 {
			return nil, fmt.Errorf("flaggo/redisprovider: unexpected SCAN reply format")
		}

		cursor, err = redis.Int(reply[0], nil)
		if err != nil {
			return nil, fmt.Errorf("flaggo/redisprovider: parse cursor: %w", err)
		}

		keys, err := redis.Strings(reply[1], nil)
		if err != nil {
			return nil, fmt.Errorf("flaggo/redisprovider: parse keys: %w", err)
		}

		for _, fullKey := range keys {
			flagName := fullKey[len(p.prefix):]
			data, err := redis.Bytes(c.Do("GET", fullKey))
			if err != nil {
				continue
			}
			var cfg flaggo.FlagConfig
			if err := json.Unmarshal(data, &cfg); err != nil {
				continue
			}
			flags[flagName] = cfg
		}

		if cursor == 0 {
			break
		}
	}

	return flags, nil
}

// Close closes the underlying Redis connection pool.
func (p *RedisProvider) Close() error {
	return p.pool.Close()
}
