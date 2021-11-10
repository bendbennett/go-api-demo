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

type HTTPProm struct {
	RequestDurationHistogram *prometheus.HistogramVec
	RequestCounter           *prometheus.CounterVec
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
	httpProm HTTPProm,
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
				httpProm,
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
	httpProm HTTPProm,
	route route,
) http.HandlerFunc {
	if httpProm.RequestCounter != nil {
		route.handler = promhttp.InstrumentHandlerCounter(
			httpProm.RequestCounter.MustCurryWith(
				prometheus.Labels{
					"route":  route.path,
					"method": route.method,
				},
			),
			route.handler,
		)
	}

	if httpProm.RequestDurationHistogram != nil {
		route.handler = promhttp.InstrumentHandlerDuration(
			httpProm.RequestDurationHistogram.MustCurryWith(
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

func NewHTTPProm(metricsEnabled bool) (HTTPProm, error) {
	if !metricsEnabled {
		return HTTPProm{}, nil
	}

	requestDurationHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of latencies for HTTP requests.",
			Buckets: []float64{.001, .002, .005, .01, .02, .05, .1, .2, .5, 1, 2, 5},
		},
		[]string{"route", "method", "code"},
	)

	err := prometheus.Register(requestDurationHistogram)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return HTTPProm{}, err
		}
	}

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_server_handled_total",
			Help: "Total number of HTTP requests completed on the server, regardless of success or failure",
		},
		[]string{"route", "method", "code"},
	)

	err = prometheus.Register(requestCounter)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return HTTPProm{}, err
		}
	}

	return HTTPProm{
		RequestDurationHistogram: requestDurationHistogram,
		RequestCounter:           requestCounter,
	}, nil
}
