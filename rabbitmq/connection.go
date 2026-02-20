package rabbitmq

import (
	"errors"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrReconnecting = errors.New("rabbitmq: connection is reconnecting")
	ErrClosed = errors.New("rabbitmq: connection closed")
)

// connWrapper holds an AMQP connection and handles NotifyClose + auto-reconnect.
type connWrapper struct {
	mu       sync.RWMutex
	conn     *amqp.Connection
	url      string
	config   ReconnectConfig
	closed   bool
	done     chan struct{}
	doneOnce sync.Once
	onClosed func() // called when connection is lost (e.g. clear pool); optional
}

func newConnWrapper(url string, config ReconnectConfig, onClosed func()) (*connWrapper, error) {
	c := &connWrapper{
		url:      url,
		config:   config,
		done:     make(chan struct{}),
		onClosed: onClosed,
	}
	if err := c.dial(); err != nil {
		return nil, err
	}
	go c.runNotifyClose()
	return c, nil
}

// setOnClosed sets the callback invoked when the connection is lost (e.g. to clear pool).
// Must be called before any reconnect can occur.
func (c *connWrapper) setOnClosed(fn func()) {
	c.mu.Lock()
	c.onClosed = fn
	c.mu.Unlock()
}

func (c *connWrapper) dial() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

func (c *connWrapper) runNotifyClose() {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return
	}
	closeCh := conn.NotifyClose(make(chan *amqp.Error, 1))
	select {
	case err := <-closeCh:
		if err != nil {
			// Connection closed (e.g. broker restart, network)
			c.reconnect()
		}
	case <-c.done:
		return
	}
}

func (c *connWrapper) reconnect() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	oldConn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if oldConn != nil {
		_ = oldConn.Close()
	}
	if c.onClosed != nil {
		c.onClosed()
	}

	interval := c.config.InitialInterval
	if interval <= 0 {
		interval = time.Second
	}
	maxInterval := c.config.MaxInterval
	if maxInterval <= 0 {
		maxInterval = 30 * time.Second
	}
	maxRetries := c.config.MaxRetries
	attempt := 0

	for {
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()

		if maxRetries > 0 && attempt >= maxRetries {
			return
		}
		attempt++

		if err := c.dial(); err != nil {
			time.Sleep(interval)
			if interval < maxInterval {
				interval *= 2
				if interval > maxInterval {
					interval = maxInterval
				}
			}
			continue
		}
		go c.runNotifyClose()
		return
	}
}

// Connection returns the current AMQP connection or nil if reconnecting/closed.
func (c *connWrapper) Connection() *amqp.Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// Channel opens a new channel from the current connection.
func (c *connWrapper) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	conn := c.conn
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return nil, ErrClosed
	}
	if conn == nil {
		return nil, ErrReconnecting
	}
	return conn.Channel()
}

// Close closes the wrapper and the underlying connection.
func (c *connWrapper) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	c.doneOnce.Do(func() { close(c.done) })
	if conn != nil {
		return conn.Close()
	}
	return nil
}
