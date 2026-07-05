package queue

import "context"

type Message struct {
	Body        []byte
	ContentType string
}

type Handler func(ctx context.Context, message Message) error

type Publisher interface {
	Publish(ctx context.Context, message Message) error
}

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
}

type Broker interface {
	Publisher
	Consumer
	Close() error
}
