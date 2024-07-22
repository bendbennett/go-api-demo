package consume

import (
	"context"
	"errors"

	"github.com/bendbennett/go-api-demo/internal/format"
	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/mitchellh/mapstructure"
)

type processor struct {
	creator user.Creator
}

func NewProcessor(
	creator user.Creator,
) *processor {
	return &processor{
		creator,
	}
}

// Process examines userBefore and userAfter to determine whether to create, update or delete.
// The first case checks whether the before and after values are the same (i.e., no changes),
// which should only occur for tombstone events (see consumer consume func) which should be
// filtered out in the consumer, but we are just being defensive.
// The second case checks whether before is empty, in which case a user is being created.
// TODO: Implement update and delete.
func (p *processor) Process(
	ctx context.Context,
	data any,
) error {
	beforeAfter := beforeAfter{}

	err := mapstructure.Decode(data, &beforeAfter)
	if err != nil {
		return err
	}

	userBeforeAfter := beforeAfter.UserBeforeAfter()

	switch {
	case userBeforeAfter.before == userBeforeAfter.after:
		return nil
	case userBeforeAfter.before == (user.User{}):
		return p.creator.Create(ctx, userBeforeAfter.after)
	default:
		return errors.New("not implemented")
	}
}

type beforeAfter struct {
	Before value
	After  value
}

type value struct {
	Value usr `mapstructure:"mysql.go_api_demo.users.Value"`
}

type usr struct {
	ID        string `mapstructure:"id"`
	FirstName string `mapstructure:"first_name"`
	LastName  string `mapstructure:"last_name"`
	CreatedAt int64  `mapstructure:"created_at"`
}

type userBeforeAfter struct {
	before user.User
	after  user.User
}

// UserBeforeAfter converts beforeAfter to struct
// containing user.User for before and after.
//
// The CreatedAt timestamp (io.debezium.time.Timestamp) is an
// int64 that represents the unix timestamp in msec.
func (bf beforeAfter) UserBeforeAfter() userBeforeAfter {
	return userBeforeAfter{
		before: user.User{
			CreatedAt: format.MsecToTime(bf.Before.Value.CreatedAt),
			ID:        bf.Before.Value.ID,
			FirstName: bf.Before.Value.FirstName,
			LastName:  bf.Before.Value.LastName,
		},
		after: user.User{
			CreatedAt: format.MsecToTime(bf.After.Value.CreatedAt),
			ID:        bf.After.Value.ID,
			FirstName: bf.After.Value.FirstName,
			LastName:  bf.After.Value.LastName,
		},
	}
}
