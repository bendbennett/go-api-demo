package consume

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/bendbennett/go-api-demo/internal/consume"

	"github.com/Shopify/sarama"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/user"
)

type userProcessor interface {
	Process(context.Context, inputData) error
}

// Consumer represents a Sarama consumer group consumer
type userConsumer struct {
	userProcessor userProcessor
	log           log.Logger
}

func NewConsumer(
	confKafka config.Kafka,
	userProcessor userProcessor,
	log log.Logger,
) (consume.Run, io.Closer, error) {
	if !confKafka.UserConsumer.IsEnabled {
		return &consume.CCGNoop{}, nil, nil
	}

	userConfig := sarama.NewConfig()
	userConfig.Version = confKafka.Version
	userConfig.Consumer.Group.Rebalance.Strategy = confKafka.UserConsumer.GroupRebalanceStrategy

	cg, err := sarama.NewConsumerGroup(
		confKafka.Brokers,
		confKafka.UserConsumer.GroupID,
		userConfig,
	)
	if err != nil {
		return nil, nil, err
	}

	return &consume.CCG{
		Consumer: &userConsumer{
			userProcessor,
			log,
		},
		ConsumerGroup: cg,
		Topics:        confKafka.UserConsumer.Topics,
	}, cg, nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (uc *userConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (uc *userConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consume loop of ConsumerGroupClaim's Messages().
// TODO: Implement retry topic for error cases.
//
// The Kafka connector emits events with a non-nil key and a nil value as these represent "tombstone"
// events for use by compaction. We therefore need to check whether the inputData is an empty struct,
// returned from userBeforeAfter which indicates the message should be ignored.
func (uc *userConsumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for message := range claim.Messages() {
		userBeforeAfter, err := uc.userBeforeAfter(message)
		if err != nil {
			uc.log.Error(err)
			continue
		}

		if userBeforeAfter == (inputData{}) {
			continue
		}

		err = uc.userProcessor.Process(session.Context(), userBeforeAfter)
		if err != nil {
			uc.log.Error(err)
			continue
		}

		session.MarkMessage(message, "")
	}

	return nil
}

type userPayload struct {
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
	before user.User
	after  user.User
}

// userBeforeAfter converts the JSON payload retrieved from the Kafka message to map[string]user.User.
//
// The Kafka connector emits events with a non-nil key and a nil value as these represent "tombstone"
// events for use by compaction. We therefore need to check whether the message.Value is nil and if
// so, return an empty inputData struct.
//
// The CreatedAt timestamp (io.debezium.time.Timestamp) that is returned in the message is an int64
// that represents the unix timestamp in msec. As the first arg to time.Unix is expected to be sec,
// the CreatedAt value from the message needs to be divided by 1000. The second arg to time.Unix is
// nsec, so the msec part of created at is multiplied by 1000000.
func (uc *userConsumer) userBeforeAfter(message *sarama.ConsumerMessage) (inputData, error) {
	var userPayload userPayload

	if message.Value == nil {
		return inputData{}, nil
	}

	err := json.Unmarshal(message.Value, &userPayload)
	if err != nil {
		return inputData{}, err
	}

	const (
		divisor    = 1e3
		multiplier = 1e6
	)

	beforeCreatedAtSec, beforeCreateAtNanoSec, afterCreatedAtSec, afterCreateAtNanoSec :=
		int64(0), int64(0), int64(0), int64(0)

	beforeCreatedAt, afterCreatedAt := time.Time{}, time.Time{}

	if userPayload.Payload.Before.CreatedAt > 0 {
		beforeCreatedAtSec = userPayload.Payload.Before.CreatedAt / divisor
		beforeCreateAtNanoSec = (userPayload.Payload.Before.CreatedAt - (beforeCreatedAtSec * divisor)) * multiplier
		beforeCreatedAt = time.Unix(beforeCreatedAtSec, beforeCreateAtNanoSec)
	}

	if userPayload.Payload.After.CreatedAt > 0 {
		afterCreatedAtSec = userPayload.Payload.After.CreatedAt / divisor
		afterCreateAtNanoSec = (userPayload.Payload.After.CreatedAt - (afterCreatedAtSec * divisor)) * multiplier
		afterCreatedAt = time.Unix(afterCreatedAtSec, afterCreateAtNanoSec)
	}

	return inputData{
		before: user.User{
			CreatedAt: beforeCreatedAt,
			ID:        userPayload.Payload.Before.ID,
			FirstName: userPayload.Payload.Before.FirstName,
			LastName:  userPayload.Payload.Before.LastName,
		},
		after: user.User{
			CreatedAt: afterCreatedAt,
			ID:        userPayload.Payload.After.ID,
			FirstName: userPayload.Payload.After.FirstName,
			LastName:  userPayload.Payload.After.LastName,
		},
	}, nil
}
