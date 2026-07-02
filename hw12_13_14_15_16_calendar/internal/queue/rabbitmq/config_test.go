package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const testURL = "amqp://rabbit:password@localhost:5672/"

func TestConfigWithDefaults(t *testing.T) {
	t.Parallel()

	config := Config{URL: testURL}.withDefaults()

	require.Equal(t, testURL, config.URL)
	require.Equal(t, DefaultExchange, config.Exchange)
	require.Equal(t, DefaultQueue, config.Queue)
	require.Equal(t, DefaultRoutingKey, config.RoutingKey)
}

func TestConfigWithDefaultsKeepsExplicitValues(t *testing.T) {
	t.Parallel()

	config := Config{
		URL:        testURL,
		Exchange:   "custom.exchange",
		Queue:      "custom.queue",
		RoutingKey: "custom.key",
	}.withDefaults()

	require.Equal(t, "custom.exchange", config.Exchange)
	require.Equal(t, "custom.queue", config.Queue)
	require.Equal(t, "custom.key", config.RoutingKey)
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	config := Config{
		URL:        testURL,
		Exchange:   DefaultExchange,
		Queue:      DefaultQueue,
		RoutingKey: DefaultRoutingKey,
	}

	require.NoError(t, config.validate())
}

func TestConfigValidateRequiresURL(t *testing.T) {
	t.Parallel()

	config := Config{
		Exchange:   DefaultExchange,
		Queue:      DefaultQueue,
		RoutingKey: DefaultRoutingKey,
	}

	require.EqualError(t, config.validate(), "rabbitmq url is empty")
}

func TestConfigValidateRequiresTopologyNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name: "exchange",
			config: Config{
				URL:        testURL,
				Queue:      DefaultQueue,
				RoutingKey: DefaultRoutingKey,
			},
			wantErr: "rabbitmq exchange is empty",
		},
		{
			name: "queue",
			config: Config{
				URL:        testURL,
				Exchange:   DefaultExchange,
				RoutingKey: DefaultRoutingKey,
			},
			wantErr: "rabbitmq queue is empty",
		},
		{
			name: "routing key",
			config: Config{
				URL:      testURL,
				Exchange: DefaultExchange,
				Queue:    DefaultQueue,
			},
			wantErr: "rabbitmq routing key is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.EqualError(t, tt.config.validate(), tt.wantErr)
		})
	}
}
