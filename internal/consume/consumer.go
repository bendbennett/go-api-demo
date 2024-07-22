package consume

import (
	"context"
	"fmt"
	"io"

	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/metrics"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type reader interface {
	FetchMessage(context.Context) (kafka.Message, error)
	CommitMessages(context.Context, ...kafka.Message) error
	Stats() kafka.ReaderStats
}

type consumeFunc func(context.Context, *c, kafka.Message) error

type decoder interface {
	Decode([]byte) (interface{}, error)
}

type processor interface {
	Process(context.Context, any) error
}

type c struct {
	reader      reader
	consumeFunc consumeFunc
	processor   processor
	log         log.Logger
	decoder     decoder
	groupID     string
}

func NewConsumers(
	conf config.KafkaConsumer,
	telemetryEnabled bool,
	consumerMetricsLabels metrics.ConsumerMetricsLabels,
	consumerMetricsCollector metrics.ConsumerMetricsCollector,
	processor processor,
	decoder decoder,
	log log.Logger,
) ([]*c, []io.Closer, error) {
	var (
		consumers []*c
		closers   []io.Closer
	)

	for i := 0; i < conf.Num; i++ {
		reader := kafka.NewReader(conf.ReaderConfig)

		consumeFunc := cf(
			telemetryEnabled,
			consumerMetricsLabels.EntityType(),
			consumerMetricsLabels.Destination(),
		)

		groupID := fmt.Sprintf("%v-%v", conf.ReaderConfig.GroupID, i)

		err := consumerMetricsCollector.RegisterMetrics(telemetryEnabled, reader.Stats, groupID)

		if err != nil {
			return nil, nil, err
		}

		consumers = append(consumers, &c{
			reader:      reader,
			consumeFunc: consumeFunc,
			processor:   processor,
			log:         log,
			groupID:     groupID,
			decoder:     decoder,
		})

		closers = append(closers, reader)
	}

	return consumers, closers, nil
}

// Run is executed in a loop to continuously consume messages.
// TODO: Implement retry topic for error cases.
func (c *c) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if err == ctx.Err() {
				c.log.Infof(err.Error())
				return nil
			}

			c.log.Error(err)
			continue
		}

		err = c.consumeFunc(ctx, c, msg)
		if err != nil {
			c.log.Error(err)
		}
	}
}

// consume parses the msg and then calls Process.
// The Kafka connector emits events with a non-nil key and a nil value as these represent "tombstone"
// events for use by compaction. We therefore need to check whether the msg.Value is nil and if so,
// the message should be committed and ignored.
func consume(
	ctx context.Context,
	c *c,
	msg kafka.Message,
) error {
	if msg.Value == nil {
		err := c.reader.CommitMessages(ctx, msg)
		if err != nil {
			return err
		}
		return nil
	}

	// https://stackoverflow.com/questions/40548909/consume-kafka-avro-messages-in-go
	nMsg, err := c.decoder.Decode(msg.Value)
	if err != nil {
		return err
	}

	err = c.processor.Process(ctx, nMsg)
	if err != nil {
		return err
	}

	err = c.reader.CommitMessages(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

// cf decorates consume func with tracing if telemetry is enabled. We avoid wrapping tracing
// around c.reader.FetchMessage as this function blocks, so in cases where we are
// waiting for messages to arrive this produces spans that include the wait time.
func cf(
	telemetryEnabled bool,
	entityType string,
	destination string,
) consumeFunc {
	if telemetryEnabled {
		return func(ctx context.Context, c *c, msg kafka.Message) error {
			ctx, oSpan := otel.GetTracerProvider().Tracer("").Start(
				ctx,
				fmt.Sprintf("consume: %s: %s", destination, entityType),
			)
			defer oSpan.End()

			return consume(ctx, c, msg)
		}
	}

	return consume
}
