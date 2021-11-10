package consume

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/bendbennett/go-api-demo/internal/config"
	prom "github.com/bendbennett/go-api-demo/internal/consume"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/opentracing/opentracing-go"
	kafka "github.com/segmentio/kafka-go"
)

type consumeFunc func(context.Context, *c, kafka.Message) error

type reader interface {
	FetchMessage(context.Context) (kafka.Message, error)
	CommitMessages(context.Context, ...kafka.Message) error
	Stats() kafka.ReaderStats
}

type c struct {
	reader        reader
	consumeFunc   consumeFunc
	processor     Processor
	log           log.Logger
	promCollector prom.PromCollector
}

func NewConsumer(
	conf config.KafkaConsumer,
	isTracingEnabled bool,
	promCollector prom.PromCollector,
	processor Processor,
	log log.Logger,
) (*c, io.Closer) {
	r := kafka.NewReader(conf.ReaderConfig)

	cf := consume

	if isTracingEnabled {
		cf = wrappedConsume
	}

	return &c{
		reader:        r,
		consumeFunc:   cf,
		processor:     processor,
		promCollector: promCollector,
		log:           log,
	}, r
}

// Run is executed in a loop to continuously consume messages.
// A goroutine is used to intermittently collect prometheus metrics
// for the Kafka reader.
// TODO: Implement retry topic for error cases.
func (c *c) Run(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(c.promCollector.Interval)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				kafkaStats := c.reader.Stats()

				stats := prom.Stats{
					Messages:      kafkaStats.Messages,
					QueueLength:   kafkaStats.QueueLength,
					QueueCapacity: kafkaStats.QueueCapacity,
					Lag:           kafkaStats.Lag,
				}

				c.promCollector.Update(stats)
			}
		}
	}()

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if err == ctx.Err() {
				c.log.Infof("%s", err)
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

	var userPayload usrPayload

	err := json.Unmarshal(msg.Value, &userPayload)
	if err != nil {
		return err
	}

	input := inputData{
		Before: userPayload.Payload.Before,
		After:  userPayload.Payload.After,
	}

	err = c.processor.Process(ctx, input)
	if err != nil {
		return err
	}

	err = c.reader.CommitMessages(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

// wrappedConsume decorates consume func with tracing. We avoid wrapping tracing around
// c.reader.FetchMessage as this function blocks, so in cases where we are waiting for
// messages to arrive this produces spans that include the wait time.
func wrappedConsume(
	ctx context.Context,
	c *c,
	msg kafka.Message,
) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"Consume: User",
	)
	defer span.Finish()

	return consume(ctx, c, msg)
}

type usrPayload struct {
	Payload payload `json:"payload"`
}

type payload struct {
	Before usr `json:"before"`
	After  usr `json:"after"`
}

type usr struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt int64  `json:"created_at"`
}

type inputData struct {
	Before usr
	After  usr
}
