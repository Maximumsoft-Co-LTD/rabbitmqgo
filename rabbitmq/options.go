package rabbitmq

import "time"

// PoolConfig configures the channel pool per connection.
type PoolConfig struct {
	// MinChannels is the minimum number of channels to keep in the pool (optional, 0 = no prefill).
	MinChannels int
	// MaxChannels is the maximum number of channels in the pool (required, e.g. 10).
	MaxChannels int
	// IdleTimeout is how long an idle channel stays in the pool before being closed.
	IdleTimeout time.Duration
	// MaxChannelLifetime is the maximum lifetime of a channel; after this it is discarded (0 = no limit).
	MaxChannelLifetime time.Duration
}

// DefaultPoolConfig returns a default channel pool config.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MinChannels:       0,
		MaxChannels:       10,
		IdleTimeout:       5 * time.Minute,
		MaxChannelLifetime: 0,
	}
}

// ReconnectConfig configures auto-reconnect behavior when the connection is lost.
type ReconnectConfig struct {
	// MaxRetries is the maximum number of reconnect attempts (0 = retry indefinitely).
	MaxRetries int
	// InitialInterval is the initial backoff interval (e.g. 1*time.Second).
	InitialInterval time.Duration
	// MaxInterval caps the backoff interval (e.g. 30*time.Second).
	MaxInterval time.Duration
}

// DefaultReconnectConfig returns a default reconnect config.
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxRetries:       0,
		InitialInterval: time.Second,
		MaxInterval:     30 * time.Second,
	}
}

// ConsumeOpts configures Consume behavior.
type ConsumeOpts struct {
	// AutoReconnect re-establishes the consumer when channel/connection is lost (default true).
	AutoReconnect bool
	// ConsumerTag is the consumer tag (optional).
	ConsumerTag string
	// AutoAck when true messages are acked automatically (default false for manual ack).
	AutoAck bool
	// Exclusive when true the consumer is exclusive to this connection.
	Exclusive bool
	// NoLocal when true the server will not deliver messages published by this connection.
	NoLocal bool
}

// DefaultConsumeOpts returns default consume options.
func DefaultConsumeOpts() ConsumeOpts {
	return ConsumeOpts{
		AutoReconnect: true,
		AutoAck:       false,
		Exclusive:     false,
		NoLocal:       false,
	}
}

// QueueDeclareOpts configures QueueDeclare.
type QueueDeclareOpts struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]interface{}
}

// DefaultQueueDeclareOpts returns default queue declare options.
func DefaultQueueDeclareOpts() QueueDeclareOpts {
	return QueueDeclareOpts{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}

// PublishOpts configures Publish (content-type, correlation-id, reply-to, etc.).
type PublishOpts struct {
	ContentType     string
	ContentEncoding string
	CorrelationID   string
	ReplyTo         string
	MessageID       string
	DeliveryMode    uint8
	Priority        uint8
}
