package memory

import (
	"context"
	"sync"

	"github.com/bendbennett/go-api-demo/internal/user"
)

type UserStorage struct {
	users map[string]user.User
	mu    sync.Mutex
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		users: make(map[string]user.User),
	}
}

func (u *UserStorage) Create(
	ctx context.Context,
	users ...user.User,
) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	for _, usr := range users {
		u.users[usr.ID] = usr
	}

	return nil
}

func (u *UserStorage) Read(context.Context) ([]user.User, error) {
	var users []user.User

	for _, u := range u.users {
		users = append(users, u)
	}

	return users, nil
}
