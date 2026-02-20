package rabbitmq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DeliveryHandler is called for each consumed message. Return nil to ack (if manual ack), or an error.
type DeliveryHandler func(d amqp.Delivery) error

// Consume starts consuming from queue and calls handler for each message.
// When AutoReconnect is true (default), the consumer is re-established if the channel/connection is lost.
func (c *Client) Consume(ctx context.Context, queue string, handler DeliveryHandler, opts *ConsumeOpts) error {
	if opts == nil {
		o := DefaultConsumeOpts()
		opts = &o
	}
	for {
		err := c.consumeOnce(ctx, queue, handler, opts)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if !opts.AutoReconnect {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (c *Client) consumeOnce(ctx context.Context, queue string, handler DeliveryHandler, opts *ConsumeOpts) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	deliveries, err := ch.Consume(queue, opts.ConsumerTag, opts.AutoAck, opts.Exclusive, opts.NoLocal, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}
			if err := handler(d); err != nil && !opts.AutoAck {
				_ = d.Nack(false, true)
			} else if !opts.AutoAck {
				_ = d.Ack(false)
			}
		}
	}
}
