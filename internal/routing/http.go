package routing

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

type HTTPRouter struct {
	handler           http.Handler
	logger            log.Logger
	port              int
	readHeaderTimeout time.Duration
}

type HTTPControllers struct {
	UserCreateController func(w http.ResponseWriter, r *http.Request)
	UserReadController   func(w http.ResponseWriter, r *http.Request)
	UserSearchController func(w http.ResponseWriter, r *http.Request)
}

type route struct {
	path        string
	handlerFunc http.HandlerFunc
	method      string
}

// NewHTTPRouter returns a pointer to an HTTPRouter struct populated
// with the port for the server, a configured router and a logger.
func NewHTTPRouter(
	controllers HTTPControllers,
	logger log.Logger,
	telemetryEnabled bool,
	port int,
	readHeaderTimeout time.Duration,
) *HTTPRouter {
	router := mux.NewRouter()

	router.HandleFunc(
		"/",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	router.HandleFunc(
		"/debug/pprof/profile",
		pprof.Profile,
	)

	routes := []route{
		{
			path:        "/user",
			handlerFunc: controllers.UserCreateController,
			method:      http.MethodPost,
		},
		{
			path:        "/user",
			handlerFunc: controllers.UserReadController,
			method:      http.MethodGet,
		},
		{
			path:        "/user/search/{searchTerm}",
			handlerFunc: controllers.UserSearchController,
			method:      http.MethodGet,
		},
	}

	telemetryHandlerFunc := func(f http.HandlerFunc, path string) http.HandlerFunc {
		if !telemetryEnabled {
			return f
		}

		return func(w http.ResponseWriter, r *http.Request) {
			labeler, _ := otelhttp.LabelerFromContext(r.Context())
			labeler.Add(attribute.String("http_route", path))

			f(w, r)
		}
	}

	for _, route := range routes {
		router.HandleFunc(
			route.path,
			telemetryHandlerFunc(route.handlerFunc, route.path),
		).Methods(route.method)
	}

	var handler http.Handler = router

	if telemetryEnabled {
		handler = otelhttp.NewHandler(
			router,
			"http",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return operation + ": " + r.Method + ": " + r.URL.Path
			}),
		)
	}

	return &HTTPRouter{
		handler,
		logger,
		port,
		readHeaderTimeout,
	}
}

// Run configures and starts an HTTP server. A go routine is
// used to listen for context cancellation and triggers
// server shutdown.
func (r *HTTPRouter) Run(ctx context.Context) error {
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf(
			":%d",
			r.port,
		),
	)
	if err != nil {
		return err
	}

	s := http.Server{
		Handler:           r.handler,
		ReadHeaderTimeout: r.readHeaderTimeout,
	}

	go func() {
		<-ctx.Done()
		err := s.Shutdown(ctx)
		if err != nil {
			r.logger.Error(err)
		}
	}()

	r.logger.Infof(
		"HTTP server running on port %v",
		r.port,
	)

	err = s.Serve(listener)
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
