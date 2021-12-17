package bootstrap

import (
	"io"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/routing"
	"github.com/bendbennett/go-api-demo/internal/sanitise"
	"github.com/bendbennett/go-api-demo/internal/trace"
	"github.com/bendbennett/go-api-demo/internal/user"
	usercreate "github.com/bendbennett/go-api-demo/internal/user/create"
	userread "github.com/bendbennett/go-api-demo/internal/user/read"
	usersearch "github.com/bendbennett/go-api-demo/internal/user/search"
	"github.com/bendbennett/go-api-demo/internal/validate"
)

func newRouters(
	conf config.Config,
	logger log.Logger,
	userCache user.CreatorReader,
	userSearch user.Searcher,
) ([]app.Component, []io.Closer) {
	var (
		components []app.Component
		closers    []io.Closer
	)

	validator, err := validate.NewValidator()
	if err != nil {
		logger.Panic(err)
	}

	userStorage, closer, err := newUserStorage(
		conf.MySQL,
		conf.Storage,
		conf.Tracing.Enabled,
	)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

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

	httpProm, err := routing.NewHTTPProm(conf.Metrics.Enabled)
	if err != nil {
		logger.Panic(err)
	}

	tracer, closer, err := trace.NewTracer(logger, conf.Tracing.Enabled)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

	httpRouter := routing.NewHTTPRouter(
		httpControllers,
		logger,
		httpProm,
		tracer,
		conf.HTTPPort,
	)

	components = append(components, httpRouter)

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

	components = append(components, grpcRouter)

	return components, closers
}
