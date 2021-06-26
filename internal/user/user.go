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

type Creator interface {
	Create(context.Context, ...User) error
}

type Reader interface {
	FindAll(context.Context) ([]User, error)
}
