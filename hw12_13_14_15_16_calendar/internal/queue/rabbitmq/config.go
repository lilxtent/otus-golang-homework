package rabbitmq

import "errors"

const (
	DefaultExchange   = "calendar"
	DefaultQueue      = "calendar.notifications"
	DefaultRoutingKey = "calendar.notification"
)

type Config struct {
	URL         string `mapstructure:"url"`
	Exchange    string `mapstructure:"exchange"`
	Queue       string `mapstructure:"queue"`
	RoutingKey  string `mapstructure:"routing_key"`
	ConsumerTag string `mapstructure:"consumer_tag"`
}

func (c Config) withDefaults() Config {
	if c.Exchange == "" {
		c.Exchange = DefaultExchange
	}
	if c.Queue == "" {
		c.Queue = DefaultQueue
	}
	if c.RoutingKey == "" {
		c.RoutingKey = DefaultRoutingKey
	}

	return c
}

func (c Config) validate() error {
	if c.URL == "" {
		return errors.New("rabbitmq url is empty")
	}
	if c.Exchange == "" {
		return errors.New("rabbitmq exchange is empty")
	}
	if c.Queue == "" {
		return errors.New("rabbitmq queue is empty")
	}
	if c.RoutingKey == "" {
		return errors.New("rabbitmq routing key is empty")
	}

	return nil
}
