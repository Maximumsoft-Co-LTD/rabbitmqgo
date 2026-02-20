package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// QueueDeclare declares a queue on the channel. Use opts or nil for defaults.
func (c *Client) QueueDeclare(ch *amqp.Channel, name string, opts *QueueDeclareOpts) (amqp.Queue, error) {
	if opts == nil {
		o := DefaultQueueDeclareOpts()
		opts = &o
	}
	args := amqp.Table{}
	for k, v := range opts.Args {
		args[k] = v
	}
	return ch.QueueDeclare(name, opts.Durable, opts.AutoDelete, opts.Exclusive, opts.NoWait, args)
}

// Publish publishes a message to the queue. Body is JSON-marshalled if not []byte.
func (c *Client) Publish(ctx context.Context, queue string, body interface{}) error {
	return c.PublishWithOptions(ctx, queue, body, nil)
}

// PublishWithOptions publishes with optional PublishOpts.
func (c *Client) PublishWithOptions(ctx context.Context, queue string, body interface{}, opts *PublishOpts) error {
	var b []byte
	switch v := body.(type) {
	case []byte:
		b = v
	default:
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	ch, err := c.acquireChannel()
	if err != nil {
		return err
	}
	reuse := true
	defer func() { c.releaseChannel(ch, reuse) }()

	if _, err := c.QueueDeclare(ch, queue, nil); err != nil {
		reuse = false
		return err
	}
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        b,
		Timestamp:   time.Now(),
	}
	if opts != nil {
		if opts.ContentType != "" {
			msg.ContentType = opts.ContentType
		}
		if opts.ContentEncoding != "" {
			msg.ContentEncoding = opts.ContentEncoding
		}
		if opts.CorrelationID != "" {
			msg.CorrelationId = opts.CorrelationID
		}
		if opts.ReplyTo != "" {
			msg.ReplyTo = opts.ReplyTo
		}
		if opts.MessageID != "" {
			msg.MessageId = opts.MessageID
		}
		if opts.DeliveryMode != 0 {
			msg.DeliveryMode = opts.DeliveryMode
		}
		if opts.Priority != 0 {
			msg.Priority = opts.Priority
		}
	}
	err = ch.PublishWithContext(ctx, "", queue, false, false, msg)
	if err != nil {
		reuse = false
		return err
	}
	return nil
}

// PublishWithReply sends a message and waits for a reply (request/reply pattern).
// Reply is JSON-unmarshalled into result (can be nil to discard).
func (c *Client) PublishWithReply(ctx context.Context, queue string, payload interface{}, result interface{}) error {
	ch, err := c.acquireChannel()
	if err != nil {
		return err
	}
	reuse := true
	defer func() { c.releaseChannel(ch, reuse) }()

	replyQueue, err := ch.QueueDeclare("", true, true, false, false, nil)
	if err != nil {
		reuse = false
		return err
	}
	deliveries, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		reuse = false
		return err
	}
	var b []byte
	switch v := payload.(type) {
	case []byte:
		b = v
	default:
		var err error
		b, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	correlationID := time.Now().Format("20060102150405.000000")
	msg := amqp.Publishing{
		ContentType:   "application/json",
		Body:          b,
		CorrelationId: correlationID,
		ReplyTo:       replyQueue.Name,
		Timestamp:     time.Now(),
	}
	if err := ch.PublishWithContext(ctx, "", queue, false, false, msg); err != nil {
		reuse = false
		return err
	}
	for {
		select {
		case <-ctx.Done():
			reuse = false
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				reuse = false
				return errors.New("rabbitmq: delivery channel closed")
			}
			if d.CorrelationId != correlationID {
				continue
			}
			if result != nil {
				if err := json.Unmarshal(d.Body, result); err != nil {
					return err
				}
			}
			return nil
		}
	}
}
