package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	config Config
	conn   *amqp.Connection
	ch     *amqp.Channel
}

var _ queue.Broker = (*Client)(nil)

func New(config Config) (*Client, error) {
	config = config.withDefaults()
	if err := config.validate(); err != nil {
		return nil, err
	}

	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	client := &Client{
		config: config,
		conn:   conn,
		ch:     ch,
	}
	if err := client.declareTopology(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

func (c *Client) Publish(ctx context.Context, message queue.Message) error {
	return c.ch.PublishWithContext(
		ctx,
		c.config.Exchange,
		c.config.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  message.ContentType,
			Body:         message.Body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (c *Client) Consume(ctx context.Context, handler queue.Handler) error {
	if err := c.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("set rabbitmq qos: %w", err)
	}

	deliveries, err := c.ch.ConsumeWithContext(
		ctx,
		c.config.Queue,
		c.config.ConsumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume rabbitmq queue: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}

			message := queue.Message{
				Body:        delivery.Body,
				ContentType: delivery.ContentType,
			}
			if err := handler(ctx, message); err != nil {
				if nackErr := delivery.Nack(false, true); nackErr != nil {
					return errors.Join(err, nackErr)
				}
				continue
			}
			if err := delivery.Ack(false); err != nil {
				return err
			}
		}
	}
}

func (c *Client) Close() error {
	var err error
	if c.ch != nil {
		err = errors.Join(err, c.ch.Close())
	}
	if c.conn != nil {
		err = errors.Join(err, c.conn.Close())
	}

	return err
}

func (c *Client) declareTopology() error {
	if err := c.ch.ExchangeDeclare(
		c.config.Exchange,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare rabbitmq exchange: %w", err)
	}

	if _, err := c.ch.QueueDeclare(
		c.config.Queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare rabbitmq queue: %w", err)
	}

	if err := c.ch.QueueBind(
		c.config.Queue,
		c.config.RoutingKey,
		c.config.Exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind rabbitmq queue: %w", err)
	}

	return nil
}
