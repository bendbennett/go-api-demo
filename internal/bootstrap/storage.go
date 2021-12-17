package bootstrap

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/storage/memory"
	"github.com/bendbennett/go-api-demo/internal/storage/mysql"
	"github.com/bendbennett/go-api-demo/internal/user"
	sqldriver "github.com/go-sql-driver/mysql"
	sqltracing "github.com/luna-duclos/instrumentedsql"
	sqlopentracing "github.com/luna-duclos/instrumentedsql/opentracing"
)

func newUserStorage(
	mySQLConf *sqldriver.Config,
	storageConf config.Storage,
	isTracingEnabled bool,
) (user.CreatorReader, io.Closer, error) {
	var (
		handle interface{}
		err    error
	)

	if storageConf.Type == config.StorageTypeSQL {
		handle, err = sqlDB(
			mySQLConf,
			isTracingEnabled,
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
	tracingEnabled bool,
) (*sql.DB, error) {
	driverName := "mysql"

	if tracingEnabled {
		driverName = "instrumented-mysql"
		sql.Register(
			driverName,
			sqltracing.WrapDriver(
				sqldriver.MySQLDriver{},
				sqltracing.WithTracer(sqlopentracing.NewTracer(true)),
			),
		)
	}

	db, err := sql.Open(
		driverName,
		conf.FormatDSN(),
	)
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
