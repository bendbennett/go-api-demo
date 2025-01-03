package consume

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type readerMock struct {
}

func (r *readerMock) FetchMessage(ctx context.Context) (kafka.Message, error) {
	select {
	case <-ctx.Done():
		return kafka.Message{}, ctx.Err()
	default:
	}

	return kafka.Message{}, nil
}

func (r *readerMock) CommitMessages(context.Context, ...kafka.Message) error {
	return nil
}

func (r *readerMock) Stats() kafka.ReaderStats {
	return kafka.ReaderStats{}
}

type processorMock struct {
}

func (p *processorMock) Process(_ context.Context, data any) error {
	return nil
}

type logMock struct {
}

func (l *logMock) Panic(error) {
	panic("implement me")
}

func (l *logMock) Panicf(string, ...interface{}) {
	panic("implement me")
}

func (l *logMock) Error(error) {
	panic("implement me")
}

func (l *logMock) ErrorContext(context.Context, error) {
	panic("implement me")
}

func (l *logMock) Errorf(string, ...interface{}) {
	panic("implement me")
}

func (l *logMock) ErrorfContext(context.Context, string, ...interface{}) {
	panic("implement me")
}

func (l *logMock) Infof(string, ...interface{}) {
}

func (l *logMock) InfofContext(context.Context, string, ...interface{}) {
	panic("implement me")
}

// TODO: Consume only returns error when context is cancelled.
// Once the event handling for error conditions is in place,
// should test whether events are being pushed onto retry topic
// or not. For errors generated by unmarshaling, a separate
// topic from the retry topic should be implemented.
func TestUserConsumer_Consume(t *testing.T) {
	cases := []struct {
		name     string
		consumer *c
	}{
		{
			"user consumer consume logs and returns nil",
			&c{
				reader:      &readerMock{},
				consumeFunc: consume,
				processor:   &processorMock{},
				log:         &logMock{},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			err := c.consumer.Run(ctx)

			assert.NoError(t, err)
		})
	}
}
