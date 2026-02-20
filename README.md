# rabbitmqgo

RabbitMQ client for Go with **connection reuse**, **channel pooling**, and **auto-reconnect**. Built on [amqp091-go](https://github.com/rabbitmq/amqp091-go).

- **Go:** 1.21+
- **RabbitMQ:** 2.0+

## Install

```bash
go get rabbitmqgo
```

Or with replace for local development:

```go
require rabbitmqgo v0.0.0
replace rabbitmqgo => /path/to/rabbitmqgo
```

## Quick start

```go
package main

import (
	"context"
	"log"

	"rabbitmqgo/rabbitmq"
)

func main() {
	client, err := rabbitmq.New("amqp://guest:guest@localhost:5672/", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Publish(ctx, "my_queue", map[string]string{"hello": "world"}); err != nil {
		log.Fatal(err)
	}
}
```

## Connection reuse and cache

- **New(url, opts)** — creates a client (uncached). Use for tests or one-off use.
- **Connect(name, url, opts)** — returns a cached client for that name. First call creates and caches.
- **Close(name)** — closes and removes the cached client for that name.
- **CloseAll()** — closes all cached clients.

```go
client, _ := rabbitmq.Connect(rabbitmq.DefaultConn, "amqp://guest:guest@localhost:5672/", nil)
defer rabbitmq.Close(rabbitmq.DefaultConn)
```

## Channel pool

Each client uses a channel pool (configurable via `ClientOptions.PoolConfig`):

- **MinChannels** — minimum channels to keep (0 = no prefill).
- **MaxChannels** — maximum channels in the pool (e.g. 10).
- **IdleTimeout** — how long an idle channel stays in the pool before being closed.
- **MaxChannelLifetime** — optional max lifetime per channel (0 = no limit).

Publish and PublishWithReply acquire a channel from the pool and release it when done.

## Auto-reconnect

- **Connection level:** When the connection is lost (broker restart, network), the SDK listens to `NotifyClose`, clears the channel pool, and retries `Dial` with exponential backoff. Config via `ReconnectConfig`: **MaxRetries** (0 = indefinite), **InitialInterval**, **MaxInterval**.
- **Consumer level:** For `Consume(ctx, queue, handler, opts)`, when the delivery channel closes (channel/connection lost), the consumer is re-established automatically if **ConsumeOpts.AutoReconnect** is true (default).

## API

| Method | Description |
|--------|-------------|
| **Publish(ctx, queue, body)** | Publish message (body JSON-marshalled if not `[]byte`) |
| **PublishWithOptions(ctx, queue, body, opts)** | Publish with content-type, correlation-id, reply-to, etc. |
| **PublishWithReply(ctx, queue, payload, result)** | Request/reply: send and wait for reply, unmarshal into `result` |
| **Consume(ctx, queue, handler, opts)** | Consume with optional auto-reconnect |
| **QueueDeclare(ch, name, opts)** | Declare a queue on a channel (for advanced use) |
| **HealthCheck()** | Verify connection by opening and closing a channel |
| **Close()** | Close client and pool |

## Options

- **ClientOptions** — PoolConfig, ReconnectConfig.
- **PoolConfig** — MinChannels, MaxChannels, IdleTimeout, MaxChannelLifetime.
- **ReconnectConfig** — MaxRetries, InitialInterval, MaxInterval.
- **ConsumeOpts** — AutoReconnect, ConsumerTag, AutoAck, Exclusive, NoLocal.
- **QueueDeclareOpts** — Durable, AutoDelete, Exclusive, NoWait, Args.
- **PublishOpts** — ContentType, CorrelationID, ReplyTo, MessageID, DeliveryMode, Priority.

## Example (consumer with auto-reconnect)

```go
ctx := context.Background()
opts := &rabbitmq.ConsumeOpts{AutoReconnect: true, AutoAck: true}
err := client.Consume(ctx, "my_queue", func(d amqp.Delivery) error {
	log.Printf("Received: %s", d.Body)
	return nil
}, opts)
```

## Development

```bash
go mod tidy
go build ./...
go test ./...
go generate ./rabbitmq/...
```
