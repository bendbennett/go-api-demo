package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bendbennett/go-api-demo/internal/user"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type UserStorage struct {
	db DB
}

func NewUserStorage(db DB) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

func (u *UserStorage) Create(
	ctx context.Context,
	users ...user.User,
) error {
	ctx, cancel := context.WithTimeout(
		ctx,
		time.Second*3,
	)
	defer cancel()

	values := make([]string, 0, len(users))
	args := make([]interface{}, 0, len(users)*4)

	for _, usr := range users {
		values = append(values, "(?, ?, ?, ?)")
		args = append(args, usr.ID)
		args = append(args, usr.FirstName)
		args = append(args, usr.LastName)
		args = append(args, usr.CreatedAt)
	}

	qry := fmt.Sprintf(
		"INSERT INTO users(id, first_name, last_name, created_at) VALUES %s",
		strings.Join(
			values,
			",",
		),
	)

	_, err := u.db.ExecContext(ctx,
		qry,
		args...,
	)
	if err != nil {
		return err
	}

	return nil
}
