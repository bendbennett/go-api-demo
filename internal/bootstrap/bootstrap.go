package bootstrap

import (
	"os"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/routing"
	log "github.com/sirupsen/logrus"
)

// New configures a logger for use throughout the application,
// retrieves configuration values for components used by the
// application, configures the components, populates an app.App
// struct with the configured components and returns a pointer
// to the populated app.App.
func New() *app.App {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	logger := log.WithFields(
		log.Fields{
			"commitHash": app.CommitHash(),
		},
	)

	c := config.New()

	httpRouter := routing.NewHTTPRouter(
		logger,
		c.HTTPPort,
	)
	grpcRouter := routing.NewGRPCRouter(
		logger,
		c.GRPCPort,
	)

	return app.New(
		httpRouter,
		grpcRouter,
	)
}
