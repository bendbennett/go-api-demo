package user

import (
	"context"
	"time"
)

type User struct {
	CreatedAt time.Time
	ID        string
	FirstName string
	LastName  string
}

type CreatorReader interface {
	Creator
	Reader
}

type Creator interface {
	Create(context.Context, ...User) error
}

type Reader interface {
	Read(context.Context) ([]User, error)
}

type CreatorSearcher interface {
	Creator
	Searcher
}

type Searcher interface {
	Search(ctx context.Context, searchTerm string) ([]User, error)
}
