package routing

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/gorilla/mux"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPRouter struct {
	handler http.Handler
	logger  log.Logger
	port    int
}

type HTTPControllers struct {
	UserCreateController func(w http.ResponseWriter, r *http.Request)
	UserReadController   func(w http.ResponseWriter, r *http.Request)
	UserSearchController func(w http.ResponseWriter, r *http.Request)
}

type HTTPPromVec struct {
	HistogramVec *prometheus.HistogramVec
	CounterVec   *prometheus.CounterVec
}

type route struct {
	path    string
	handler http.HandlerFunc
	method  string
}

// NewHTTPRouter returns a pointer to an HTTPRouter struct populated
// with the port for the server, a configured router and a logger.
func NewHTTPRouter(
	controllers HTTPControllers,
	logger log.Logger,
	httpPromVec HTTPPromVec,
	tracer opentracing.Tracer,
	port int,
) *HTTPRouter {
	router := mux.NewRouter()

	router.HandleFunc(
		"/",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	router.HandleFunc(
		"/metrics",
		promhttp.Handler().ServeHTTP,
	)

	routes := []route{
		{
			path:    "/user",
			handler: controllers.UserCreateController,
			method:  http.MethodPost,
		},
		{
			path:    "/user",
			handler: controllers.UserReadController,
			method:  http.MethodGet,
		},
		{
			path:    "/user/search/{searchTerm}",
			handler: controllers.UserSearchController,
			method:  http.MethodGet,
		},
	}

	for _, route := range routes {
		router.HandleFunc(
			route.path,
			wrapMetrics(
				httpPromVec,
				route,
			),
		).Methods(route.method)
	}

	handler := nethttp.Middleware(
		tracer,
		router,
		nethttp.OperationNameFunc(func(r *http.Request) string {
			return "HTTP " + r.Method + ": " + r.URL.Path
		}),
		nethttp.MWSpanFilter(func(r *http.Request) bool {
			return r.URL.Path != "/metrics"
		}),
	)

	return &HTTPRouter{
		handler,
		logger,
		port,
	}
}

func wrapMetrics(
	httpPromVec HTTPPromVec,
	route route,
) http.HandlerFunc {
	if httpPromVec.CounterVec != nil {
		route.handler = promhttp.InstrumentHandlerCounter(
			httpPromVec.CounterVec.MustCurryWith(
				prometheus.Labels{
					"route":  route.path,
					"method": route.method,
				},
			),
			route.handler,
		)
	}

	if httpPromVec.HistogramVec != nil {
		route.handler = promhttp.InstrumentHandlerDuration(
			httpPromVec.HistogramVec.MustCurryWith(
				prometheus.Labels{
					"route":  route.path,
					"method": route.method,
				},
			),
			route.handler,
		)
	}

	return route.handler
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
		Handler: r.handler,
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
