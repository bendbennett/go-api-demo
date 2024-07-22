package bootstrap

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/XSAM/otelsql"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/storage/memory"
	"github.com/bendbennett/go-api-demo/internal/storage/mysql"
	"github.com/bendbennett/go-api-demo/internal/user"
	sqldriver "github.com/go-sql-driver/mysql"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func newUserStorage(
	mySQLConf *sqldriver.Config,
	storageConf config.Storage,
	telemetryEnabled bool,
) (user.CreatorReader, io.Closer, error) {
	var (
		handle interface{}
		err    error
	)

	if storageConf.Type == config.StorageTypeSQL {
		handle, err = sqlDB(
			mySQLConf,
			telemetryEnabled,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	switch h := handle.(type) {
	case *sql.DB:
		return mysql.NewUserStorage(
			h,
			storageConf.QueryTimeout,
		), h, nil
	default:
		return memory.NewUserStorage(), nil, nil
	}
}

func sqlDB(
	conf *sqldriver.Config,
	telemetryEnabled bool,
) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch telemetryEnabled {
	case true:
		db, err = otelsql.Open("mysql", conf.FormatDSN(), otelsql.WithAttributes(
			semconv.DBSystemMySQL,
		))
	default:
		db, err = sql.Open("mysql", conf.FormatDSN())
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to db: %w",
			err,
		)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf(
			"failed to ping db: %w",
			err,
		)
	}

	return db, nil
}
