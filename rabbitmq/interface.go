package rabbitmq

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mock_client.go -package=rabbitmq -mock_names "ClientInterface=MockClient" -source=interface.go

import (
	"context"
)

// ClientInterface is the set of RabbitMQ operations provided by the library.
// *Client implements it; use MockClient in unit tests.
type ClientInterface interface {
	Close() error
	HealthCheck() error
	Publish(ctx context.Context, queue string, body interface{}) error
	PublishWithOptions(ctx context.Context, queue string, body interface{}, opts *PublishOpts) error
	PublishWithReply(ctx context.Context, queue string, payload interface{}, result interface{}) error
	Consume(ctx context.Context, queue string, handler DeliveryHandler, opts *ConsumeOpts) error
	DeleteQueue(ctx context.Context, name string, ifUnused, ifEmpty, noWait bool) (int, error)
}
