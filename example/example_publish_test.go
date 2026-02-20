package example_test

import (
	"context"
	"fmt"
	"log"

	"rabbitmqgo/rabbitmq"
)

// ExampleNew creates a client. Requires RabbitMQ at localhost:5672 for success.
func ExampleNew() {
	client, err := rabbitmq.New("amqp://guest:guest@localhost:5672/", nil)
	if err != nil {
		log.Printf("New failed (is RabbitMQ running?): %v", err)
		return
	}
	defer client.Close()
	fmt.Println("client created")
}

// ExampleClient_Publish publishes a message. Requires RabbitMQ at localhost:5672.
func ExampleClient_Publish() {
	client, err := rabbitmq.New("amqp://guest:guest@localhost:5672/", nil)
	if err != nil {
		log.Printf("New failed: %v", err)
		return
	}
	defer client.Close()
	ctx := context.Background()
	err = client.Publish(ctx, "example_queue", map[string]string{"key": "value"})
	if err != nil {
		log.Printf("Publish failed: %v", err)
		return
	}
	fmt.Println("published")
}

// ExampleConnect uses cached connection by name. Requires RabbitMQ at localhost:5672.
func ExampleConnect() {
	client, err := rabbitmq.Connect(rabbitmq.DefaultConn, "amqp://guest:guest@localhost:5672/", nil)
	if err != nil {
		log.Printf("Connect failed: %v", err)
		return
	}
	defer rabbitmq.Close(rabbitmq.DefaultConn)
	ctx := context.Background()
	_ = client.Publish(ctx, "example_queue", []byte("hello"))
	fmt.Println("connected and published")
}
