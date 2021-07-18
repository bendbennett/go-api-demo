package bootstrap

import (
	"os"

	usercreate "github.com/bendbennett/go-api-demo/internal/user/create"
	"github.com/bendbennett/go-api-demo/internal/validate"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/routing"
	log "github.com/sirupsen/logrus"
)

// New configures a logger for use throughout the application,
// retrieves configuration application, configures HTTP and
// gRPC routers, populates an app.App struct with the configured
// routers and returns a pointer to the populated app.App.
func New() *app.App {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	logger := log.WithFields(
		log.Fields{
			"commitHash": app.CommitHash(),
		},
	)

	c := config.New()

	validator, err := validate.NewValidator()
	if err != nil {
		logger.Panic(err)
	}

	userStorage, err := NewUserStorage(c)
	if err != nil {
		logger.Panic(err)
	}

	userCreatePresenter := usercreate.NewPresenter()

	userCreateControllerHTTP := usercreate.NewHTTPController(
		validator,
		usercreate.NewInteractor(
			userStorage,
		),
		userCreatePresenter,
		logger,
	)

	httpControllers := routing.HTTPControllers{
		UserCreateController: userCreateControllerHTTP.Create,
	}

	httpRouter := routing.NewHTTPRouter(
		httpControllers,
		logger,
		c.HTTPPort,
	)

	userCreateControllerGRPC := usercreate.NewGRPCController(
		validator,
		usercreate.NewInteractor(
			userStorage,
		),
		userCreatePresenter,
		logger,
	)

	grpcControllers := routing.GRPCControllers{
		UserCreate: userCreateControllerGRPC.Create,
	}

	grpcRouter := routing.NewGRPCRouter(
		grpcControllers,
		logger,
		c.GRPCPort,
	)

	return app.New(
		httpRouter,
		grpcRouter,
	)
}
