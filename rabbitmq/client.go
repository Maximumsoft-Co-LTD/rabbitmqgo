package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DefaultConn is the default connection name used by Connect/Close.
const DefaultConn = "default"

// ClientOptions configures a Client (pool + reconnect).
type ClientOptions struct {
	PoolConfig      PoolConfig
	ReconnectConfig ReconnectConfig
}

// DefaultClientOptions returns default options.
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		PoolConfig:      DefaultPoolConfig(),
		ReconnectConfig: DefaultReconnectConfig(),
	}
}

// Client wraps an AMQP connection with channel pool and auto-reconnect.
type Client struct {
	conn *connWrapper
	pool *channelPool
}

var connCache sync.Map

// New creates a new Client (uncached). Use for tests or one-off connections.
func New(url string, opts *ClientOptions) (*Client, error) {
	if opts == nil {
		o := DefaultClientOptions()
		opts = &o
	}
	conn, err := newConnWrapper(url, opts.ReconnectConfig, nil)
	if err != nil {
		return nil, err
	}
	pool := newChannelPool(conn.Channel, opts.PoolConfig)
	conn.setOnClosed(pool.closeAll)
	return &Client{conn: conn, pool: pool}, nil
}

// Connect returns a cached Client for the given name. First call for a name creates and caches the client.
func Connect(name string, url string, opts *ClientOptions) (*Client, error) {
	if c, ok := connCache.Load(name); ok {
		return c.(*Client), nil
	}
	client, err := New(url, opts)
	if err != nil {
		return nil, err
	}
	connCache.Store(name, client)
	return client, nil
}

// Close closes this client's connection and pool.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.pool.Close()
	return c.conn.Close()
}

// Close removes the cached client for name and closes it.
func Close(name string) error {
	v, ok := connCache.LoadAndDelete(name)
	if !ok {
		return nil
	}
	return v.(*Client).Close()
}

// CloseAll closes all cached clients and clears the cache.
func CloseAll() {
	connCache.Range(func(key, value interface{}) bool {
		value.(*Client).Close()
		connCache.Delete(key)
		return true
	})
}

// acquireChannel gets a channel from the pool (or creates one).
func (c *Client) acquireChannel() (*amqp.Channel, error) {
	return c.pool.AcquireChannel()
}

// releaseChannel returns a channel to the pool. reuse should be false if the channel had an error.
func (c *Client) releaseChannel(ch *amqp.Channel, reuse bool) {
	c.pool.ReleaseChannel(ch, reuse)
}

// HealthCheck verifies the connection is alive by opening and closing a channel.
func (c *Client) HealthCheck() error {
	ch, err := c.acquireChannel()
	if err != nil {
		return err
	}
	c.releaseChannel(ch, true)
	return nil
}

// DeleteQueue removes a queue from RabbitMQ server.
func (c *Client) DeleteQueue(ctx context.Context, name string, ifUnused, ifEmpty, noWait bool) (int, error) {
	if c == nil {
		return 0, fmt.Errorf("rabbitmq client not initialized")
	}
	ch, err := c.acquireChannel()
	if err != nil {
		return 0, err
	}
	// We don't need context for QueueDelete in amqp091-go, as it's synchronous on the channel,
	// but wrapping it for interface consistency.
	defer c.releaseChannel(ch, true)
	return ch.QueueDelete(name, ifUnused, ifEmpty, noWait)
}
