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
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type UserStorage struct {
	db           DB
	queryTimeout time.Duration
}

func NewUserStorage(
	db DB,
	queryTimeout time.Duration,
) *UserStorage {
	return &UserStorage{
		db:           db,
		queryTimeout: queryTimeout,
	}
}

func (u *UserStorage) Create(
	ctx context.Context,
	users ...user.User,
) error {
	ctx, cancel := context.WithTimeout(
		ctx,
		u.queryTimeout,
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

func (u *UserStorage) Read(ctx context.Context) ([]user.User, error) {
	ctx, cancel := context.WithTimeout(
		ctx,
		u.queryTimeout,
	)
	defer cancel()

	qry := `
SELECT id, first_name, last_name, created_at
FROM users
`

	rows, err := u.db.QueryContext(
		ctx,
		qry,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []user.User

	for rows.Next() {
		var u user.User

		err := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.CreatedAt,
		)
		if err != nil {
			return users, err
		}

		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}
