package create

import (
	"context"
	"time"

	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/google/uuid"
)

type i struct {
	userCreator user.Creator
}

type interactor interface {
	create(context.Context, inputData) (outputData, error)
}

var _ interactor = (*i)(nil)

func NewInteractor(
	userCreator user.Creator,
) *i {
	return &i{
		userCreator,
	}
}

type outputData struct {
	CreatedAt time.Time
	ID        string
	FirstName string
	LastName  string
}

func (i *i) create(
	ctx context.Context,
	inputData inputData,
) (outputData, error) {
	u := user.User{
		ID:        uuid.New().String(),
		FirstName: inputData.FirstName,
		LastName:  inputData.LastName,
		CreatedAt: time.Now(),
	}

	err := i.userCreator.Create(
		ctx,
		u,
	)
	if err != nil {
		return outputData{}, err
	}

	return outputData{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	}, nil
}
