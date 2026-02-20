package example_test

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"rabbitmqgo/rabbitmq"
)

func ExampleClient_Consume() {
	client, err := rabbitmq.New("amqp://guest:guest@localhost:5672/", nil)
	if err != nil {
		log.Printf("New failed: %v", err)
		return
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	opts := &rabbitmq.ConsumeOpts{AutoReconnect: true, AutoAck: true}
	err = client.Consume(ctx, "example_queue", func(d amqp.Delivery) error {
		fmt.Printf("got: %s\n", d.Body)
		return nil
	}, opts)
	if err != nil && err != context.DeadlineExceeded {
		log.Printf("Consume: %v", err)
	}
}
