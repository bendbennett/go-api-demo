package consume

import (
	"context"
	"errors"

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

// Process examines userBefore and userAfter to determine
// whether to create, update or delete.
// TODO: Implement update and delete.
func (p *p) Process(
	ctx context.Context,
	userChange inputData,
) error {
	switch {
	case userChange.before == (user.User{}):
		return p.creator.Create(ctx, userChange.after)
	default:
		return errors.New("not implemented")
	}
}
