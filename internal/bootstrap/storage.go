package bootstrap

import (
	"database/sql"
	"fmt"

	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/storage/memory"
	"github.com/bendbennett/go-api-demo/internal/storage/mysql"
	"github.com/bendbennett/go-api-demo/internal/user"
	sqldriver "github.com/go-sql-driver/mysql"
)

func NewUserStorage(c config.Config) (user.CommandQuery, error) {
	var (
		handle interface{}
		err    error
	)

	if c.Storage.Type == config.StorageTypeSQL {
		handle, err = sqlDB(c.MySQL)
		if err != nil {
			return nil, err
		}
	}

	switch h := handle.(type) {
	case *sql.DB:
		return mysql.NewUserStorage(
			h,
			c.Storage.QueryTimeout,
		), nil
	default:
		return memory.NewUserStorage(), nil
	}
}

func sqlDB(conf *sqldriver.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}
