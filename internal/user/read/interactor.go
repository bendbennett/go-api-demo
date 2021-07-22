package read

import (
	"context"
	"time"

	"github.com/bendbennett/go-api-demo/internal/user"
)

type i struct {
	userReader user.Reader
}

type interactor interface {
	read(context.Context) (outputData, error)
}

var _ interactor = (*i)(nil)

func NewInteractor(
	userReader user.Reader,
) *i {
	return &i{
		userReader,
	}
}

type outputData []item

type item struct {
	CreatedAt time.Time
	ID        string
	FirstName string
	LastName  string
}

func (i *i) read(
	ctx context.Context,
) (outputData, error) {
	users, err := i.userReader.Read(
		ctx,
	)
	if err != nil {
		return outputData{}, err
	}

	var od outputData

	for _, u := range users {
		od = append(
			od,
			item{
				CreatedAt: u.CreatedAt,
				ID:        u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			},
		)
	}

	return od, nil
}
