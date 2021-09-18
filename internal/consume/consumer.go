package consume

import (
	"context"

	"github.com/Shopify/sarama"
)

type consumer interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error
}

type consumerGroup interface {
	Consume(context.Context, []string, sarama.ConsumerGroupHandler) error
	Close() error
}

type Run interface {
	Run(ctx context.Context) error
}

type CCG struct {
	Consumer      consumer
	ConsumerGroup consumerGroup
	Topics        []string
}

func (c *CCG) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := c.ConsumerGroup.Consume(
			ctx,
			c.Topics,
			c.Consumer,
		); err != nil {
			return err
		}
	}
}

type CCGNoop struct{}

func (cn *CCGNoop) Run(context.Context) error {
	return nil
}
