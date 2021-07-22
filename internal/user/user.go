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

type CommandQuery interface {
	Creator
	Reader
}

type Creator interface {
	Create(context.Context, ...User) error
}

type Reader interface {
	Read(context.Context) ([]User, error)
}
