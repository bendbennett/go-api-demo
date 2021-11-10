package bootstrap

import (
	"io"

	"github.com/bendbennett/go-api-demo/internal/consume"
	"github.com/bendbennett/go-api-demo/internal/trace"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/routing"
	"github.com/bendbennett/go-api-demo/internal/sanitise"
	"github.com/bendbennett/go-api-demo/internal/storage/elastic"
	"github.com/bendbennett/go-api-demo/internal/storage/redis"
	userconsume "github.com/bendbennett/go-api-demo/internal/user/consume"
	usercreate "github.com/bendbennett/go-api-demo/internal/user/create"
	userread "github.com/bendbennett/go-api-demo/internal/user/read"
	usersearch "github.com/bendbennett/go-api-demo/internal/user/search"
	"github.com/bendbennett/go-api-demo/internal/validate"
)

// New configures a logger for use throughout the application,
// retrieves configuration application, configures HTTP and
// gRPC routers, populates an app.App struct with the configured
// routers and returns a pointer to the populated app.App.
func New() *app.App {
	var closers []io.Closer

	conf := config.New()

	logger, err := log.NewLogger(conf.Logging.Production)
	if err != nil {
		panic(err)
	}

	tracer, closer, err := trace.NewTracer(logger, conf.Tracing.Enabled)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

	httpProm, err := routing.NewHTTPProm(conf.Metrics.Enabled)
	if err != nil {
		logger.Panic(err)
	}

	validator, err := validate.NewValidator()
	if err != nil {
		logger.Panic(err)
	}

	userStorage, closer, err := NewUserStorage(
		conf.MySQL,
		conf.Storage,
		conf.Tracing.Enabled,
	)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

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

	userCreateInteractor := usercreate.NewInteractor(userStorage)
	userCreatePresenter := usercreate.NewPresenter()

	userCreateControllerHTTP := usercreate.NewHTTPController(
		validator,
		userCreateInteractor,
		userCreatePresenter,
		logger,
	)

	userReadInteractor := userread.NewInteractor(userCache)
	userReadPresenter := userread.NewPresenter()

	userReadControllerHTTP := userread.NewHTTPController(
		userReadInteractor,
		userReadPresenter,
		logger,
	)

	userSearchInteractor := usersearch.NewInteractor(userSearch)
	userSearchPresenter := usersearch.NewPresenter()

	userSearchControllerHTTP := usersearch.NewHTTPController(
		sanitise.AlphaWithHyphen,
		userSearchInteractor,
		userSearchPresenter,
		logger,
	)

	httpControllers := routing.HTTPControllers{
		UserCreateController: userCreateControllerHTTP.Create,
		UserReadController:   userReadControllerHTTP.Read,
		UserSearchController: userSearchControllerHTTP.Search,
	}

	httpRouter := routing.NewHTTPRouter(
		httpControllers,
		logger,
		httpProm,
		tracer,
		conf.HTTPPort,
	)

	userCreateControllerGRPC := usercreate.NewGRPCController(
		validator,
		userCreateInteractor,
		userCreatePresenter,
		logger,
	)

	userReadControllerGRPC := userread.NewGRPCController(
		userReadInteractor,
		userReadPresenter,
		logger,
	)

	userSearchControllerGRPC := usersearch.NewGRPCController(
		sanitise.AlphaWithHyphen,
		userSearchInteractor,
		userSearchPresenter,
		logger,
	)

	grpcControllers := routing.GRPCControllers{
		UserCreate: userCreateControllerGRPC.Create,
		UserRead:   userReadControllerGRPC.Read,
		UserSearch: userSearchControllerGRPC.Search,
	}

	grpcRouter := routing.NewGRPCRouter(
		grpcControllers,
		logger,
		conf.Metrics.Enabled,
		conf.Tracing.Enabled,
		conf.GRPCPort,
	)

	userConsumerProm, err := consume.NewConsumerProm(conf.Metrics.Enabled)
	if err != nil {
		panic(err)
	}

	userConsumerPromUserCache := consume.NewConsumerPromCollector(
		"user",
		"cache",
		conf.Metrics.CollectionInterval,
		userConsumerProm,
	)

	userProcessorCache := userconsume.NewProcessor(userCache)

	userConsumerCache, closer := userconsume.NewConsumer(
		conf.UserConsumerCache,
		conf.Tracing.Enabled,
		userConsumerPromUserCache,
		userProcessorCache,
		logger,
	)
	closers = addCloser(closers, closer)

	userConsumerPromUserSearch := consume.NewConsumerPromCollector(
		"user",
		"search",
		conf.Metrics.CollectionInterval,
		userConsumerProm,
	)

	userProcessorSearch := userconsume.NewProcessor(userSearch)

	userConsumerSearch, closer := userconsume.NewConsumer(
		conf.UserConsumerSearch,
		conf.Tracing.Enabled,
		userConsumerPromUserSearch,
		userProcessorSearch,
		logger,
	)
	closers = addCloser(closers, closer)

	return app.New(
		[]app.Component{
			httpRouter,
			grpcRouter,
			userConsumerCache,
			userConsumerSearch,
		},
		closers,
	)
}

func addCloser(
	closers []io.Closer,
	closer io.Closer,
) []io.Closer {
	if closer != nil {
		closers = append(closers, closer)
	}

	return closers
}
