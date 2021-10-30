package search

import (
	"context"
	"time"

	"github.com/bendbennett/go-api-demo/internal/user"
)

type searcher interface {
	Search(ctx context.Context, searchTerm string) ([]user.User, error)
}

type i struct {
	searcher searcher
}

type interactor interface {
	search(context.Context, string) (outputData, error)
}

var _ interactor = (*i)(nil)

func NewInteractor(searcher searcher) *i {
	return &i{
		searcher,
	}
}

type outputData []item

type item struct {
	CreatedAt time.Time
	ID        string
	FirstName string
	LastName  string
}

func (i *i) search(
	ctx context.Context,
	searchTerm string,
) (outputData, error) {
	users, err := i.searcher.Search(ctx, searchTerm)
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
