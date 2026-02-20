package rabbitmq

import (
	"container/list"
	"errors"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrPoolClosed = errors.New("rabbitmq: pool closed")
	ErrPoolFull   = errors.New("rabbitmq: pool full")
)

type pooledChannel struct {
	ch        *amqp.Channel
	lastUsed  time.Time
	createdAt time.Time
}

// channelPool holds reusable AMQP channels.
type channelPool struct {
	mu            sync.Mutex
	source        func() (*amqp.Channel, error)
	config        PoolConfig
	idle          *list.List
	inUse         int
	closed        bool
}

func newChannelPool(source func() (*amqp.Channel, error), config PoolConfig) *channelPool {
	p := &channelPool{
		source: source,
		config: config,
		idle:   list.New(),
	}
	return p
}

// AcquireChannel returns a channel from the pool or creates a new one.
func (p *channelPool) AcquireChannel() (*amqp.Channel, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, ErrPoolClosed
	}
	now := time.Now()
	for p.idle.Len() > 0 {
		e := p.idle.Front()
		p.idle.Remove(e)
		pc := e.Value.(*pooledChannel)
		if p.config.IdleTimeout > 0 && now.Sub(pc.lastUsed) > p.config.IdleTimeout {
			_ = pc.ch.Close()
			continue
		}
		if p.config.MaxChannelLifetime > 0 && now.Sub(pc.createdAt) > p.config.MaxChannelLifetime {
			_ = pc.ch.Close()
			continue
		}
		p.inUse++
		return pc.ch, nil
	}
	if p.inUse+p.idle.Len() >= p.config.MaxChannels {
		return nil, ErrPoolFull
	}
	ch, err := p.source()
	if err != nil {
		return nil, err
	}
	p.inUse++
	return ch, nil
}

// ReleaseChannel returns a channel to the pool or closes it.
// reuse should be false if the channel had an error or is invalid.
func (p *channelPool) ReleaseChannel(ch *amqp.Channel, reuse bool) {
	if ch == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.inUse--
	if p.inUse < 0 {
		p.inUse = 0
	}
	if p.closed || !reuse || p.idle.Len()+p.inUse >= p.config.MaxChannels {
		_ = ch.Close()
		return
	}
	now := time.Now()
	p.idle.PushBack(&pooledChannel{ch: ch, lastUsed: now, createdAt: now})
}

// closeAll closes all idle and marks pool closed; callers must stop using channels.
func (p *channelPool) closeAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	for e := p.idle.Front(); e != nil; e = p.idle.Front() {
		p.idle.Remove(e)
		pc := e.Value.(*pooledChannel)
		_ = pc.ch.Close()
	}
}

// Close closes the pool and all idle channels.
func (p *channelPool) Close() {
	p.closeAll()
}

