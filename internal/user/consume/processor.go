package consume

import (
	"context"
	"errors"
	"time"

	"github.com/bendbennett/go-api-demo/internal/user"
)

type p struct {
	creator user.Creator
}

type Processor interface {
	Process(context.Context, inputData) error
}

func NewProcessor(
	creator user.Creator,
) *p {
	return &p{
		creator,
	}
}

// Process examines userBefore and userAfter to determine whether to create, update or delete.
// The first case checks whether the before and after values are the same (i.e., no changes),
// which should only occur for tombstone events (see consumer consume func) which should be
// filtered out in the consumer, but we are just being defensive.
// The second case checks whether before is empty, in which case a user is being created.
// TODO: Implement update and delete.
func (p *p) Process(
	ctx context.Context,
	inputData inputData,
) error {
	diff := p.userBeforeAfter(inputData)

	switch {
	case diff.before == diff.after:
		return nil
	case diff.before == (user.User{}):
		return p.creator.Create(ctx, diff.after)
	default:
		return errors.New("not implemented")
	}
}

// userBeforeAfter converts the inputData to map[string]user.User.
//
// The CreatedAt timestamp (io.debezium.time.Timestamp) is an int64 that represents the unix
// timestamp in msec. As the first arg to time.Unix is expected to be sec, the CreatedAt
// value needs to be divided by 1000. The second arg to time.Unix is nsec, so the msec that
// remain after the sec are subtracted are multiplied by 1000000.
func (p *p) userBeforeAfter(inputData inputData) userChange {
	msecToTime := func(msec int64) time.Time {
		const (
			divisor    = 1e3
			multiplier = 1e6
		)

		t := time.Time{}

		if msec > 0 {
			sec := msec / divisor
			nsec := (msec - (sec * divisor)) * multiplier
			t = time.Unix(sec, nsec)
		}

		return t
	}

	return userChange{
		before: user.User{
			CreatedAt: msecToTime(inputData.Before.CreatedAt),
			ID:        inputData.Before.ID,
			FirstName: inputData.Before.FirstName,
			LastName:  inputData.Before.LastName,
		},
		after: user.User{
			CreatedAt: msecToTime(inputData.After.CreatedAt),
			ID:        inputData.After.ID,
			FirstName: inputData.After.FirstName,
			LastName:  inputData.After.LastName,
		},
	}
}

type userChange struct {
	before user.User
	after  user.User
}
