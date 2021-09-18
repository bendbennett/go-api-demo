package consume

import (
	"context"
	"errors"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
)

type consumerGroupConsumeError struct {
}

func (c *consumerGroupConsumeError) Consume(context.Context, []string, sarama.ConsumerGroupHandler) error {
	return errors.New("consumer group consume error")
}

func (c *consumerGroupConsumeError) Close() error {
	return nil
}

func TestConsumer_Run(t *testing.T) {
	cases := []struct {
		name          string
		consumerGroup consumerGroup
		returnsErr    bool
	}{
		{
			"consumer group consume returns error",
			&consumerGroupConsumeError{},
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ccg := CCG{
				ConsumerGroup: c.consumerGroup,
			}

			err := ccg.Run(context.Background())

			assert.Error(t, err)
		})
	}
}
