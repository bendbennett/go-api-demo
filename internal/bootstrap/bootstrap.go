package bootstrap

import (
	"io"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/storage/elastic"
	"github.com/bendbennett/go-api-demo/internal/storage/redis"
)

// New configures a logger for use throughout the application,
// retrieves configuration application, configures HTTP and
// gRPC routers, populates an app.App struct with the configured
// routers and returns a pointer to the populated app.App.
// nolint:gocyclo
func New() *app.App {
	var (
		components []app.Component
		closers    []io.Closer
	)

	conf := config.New()

	logger, err := log.NewLogger(conf.Logging.Production)
	if err != nil {
		panic(err)
	}

	userCache, closer, err := redis.NewUserCache(
		conf.Redis,
		conf.Tracing.Enabled,
	)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

	userSearch, err := elastic.NewUserSearch(
		conf.Elasticsearch,
		conf.Tracing.Enabled,
	)
	if err != nil {
		logger.Panic(err)
	}

	routers, closrs := newRouters(conf, logger, userCache, userSearch)
	components = append(components, routers...)
	closers = addCloser(closers, closrs...)

	consumers, closrs, err := newConsumers(conf, logger, userCache, userSearch)
	if err != nil {
		logger.Panic(err)
	}

	components = append(components, consumers...)
	closers = addCloser(closers, closrs...)

	return app.New(
		components,
		closers,
	)
}

func addCloser(
	closers []io.Closer,
	closer ...io.Closer,
) []io.Closer {
	if closer != nil {
		closers = append(closers, closer...)
	}

	return closers
}
