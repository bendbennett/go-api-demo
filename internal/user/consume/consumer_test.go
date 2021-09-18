package consume

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/bendbennett/go-api-demo/internal/log"

	"github.com/Shopify/sarama"
	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/stretchr/testify/assert"
)

type userProcessorMock struct {
}

func (u *userProcessorMock) Process(ctx context.Context, id inputData) error {
	expected := inputData{
		before: user.User{
			CreatedAt: time.Time{},
		},
		after: user.User{
			CreatedAt: time.Unix(1633781642, 0),
			ID:        "135c5b00-5463-4c05-b26b-7f886f2b65dd",
			FirstName: "john",
			LastName:  "smith",
		},
	}

	if !reflect.DeepEqual(expected, id) {
		return errors.New("output data is not equal to expected data")
	}

	return nil
}

type session struct {
}

func (s session) Claims() map[string][]int32 {
	panic("implement me")
}

func (s session) MemberID() string {
	panic("implement me")
}

func (s session) GenerationID() int32 {
	panic("implement me")
}

func (s session) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	panic("implement me")
}

func (s session) Commit() {
	panic("implement me")
}

func (s session) ResetOffset(topic string, partition int32, offset int64, metadata string) {
	panic("implement me")
}

func (s session) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {}

func (s session) Context() context.Context {
	return context.Background()
}

type claim struct {
}

func (c claim) Topic() string {
	panic("implement me")
}

func (c claim) Partition() int32 {
	panic("implement me")
}

func (c claim) InitialOffset() int64 {
	panic("implement me")
}

func (c claim) HighWaterMarkOffset() int64 {
	panic("implement me")
}

func (c claim) Messages() <-chan *sarama.ConsumerMessage {
	chanMsg := make(chan *sarama.ConsumerMessage, 1)

	cm := &sarama.ConsumerMessage{
		Value: []byte(`{
			"payload": {
			"before": null,
			"after": {
				"id": "135c5b00-5463-4c05-b26b-7f886f2b65dd",
				"first_name": "john",
				"last_name": "smith",
				"created_at": 1633781642000
			}}}`),
	}

	chanMsg <- cm

	close(chanMsg)

	return chanMsg
}

type logger struct {
}

func (l logger) Panic(err error) {
	panic("implement me")
}

func (l logger) Panicf(msg string, args ...interface{}) {
	panic("implement me")
}

func (l logger) Error(err error) {
	panic("implement me")
}

func (l logger) Errorf(msg string, args ...interface{}) {
	panic("implement me")
}

func (l logger) Infof(msg string, args ...interface{}) {
	panic("implement me")
}

func (l logger) WithSpan(ctx context.Context) log.Logger {
	panic("implement me")
}

// TODO: ConsumeClaim does not return any errors, so this test
// is really verifying whether log.Error is being called as that
// would result in a panic. Once the event handling for error
// conditions is in place, should test whether events are being
// pushed onto retry topic or not. For unmarshalling errors,
// a separate topic from the retry topic is required.
func TestUserConsumer_ConsumeClaim(t *testing.T) {
	cases := []struct {
		name         string
		userConsumer userConsumer
	}{
		{
			"user consumer consume claim success",
			userConsumer{
				userProcessor: &userProcessorMock{},
				log:           logger{},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.userConsumer.ConsumeClaim(&session{}, &claim{})

			assert.NoError(t, err)
		})
	}
}
