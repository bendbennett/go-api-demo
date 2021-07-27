package routing

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type HTTPRouter struct {
	handler http.Handler
	logger  *log.Entry
	port    int
}

type HTTPControllers struct {
	UserCreateController func(w http.ResponseWriter, r *http.Request)
	UserReadController   func(w http.ResponseWriter, r *http.Request)
}

// NewHTTPRouter returns a pointer to an HTTPRouter struct populated
// with the port for the server, a configured router and a logger.
func NewHTTPRouter(
	controllers HTTPControllers,
	logger *log.Entry,
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

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.001, .002, .005, .01, .02, .05, .1, .2, .5},
		},
		[]string{"route", "method", "code"},
	)

	err := prometheus.Register(duration)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			logger.Panic(err)
		}
	}

	for _, route := range []struct {
		path    string
		handler http.HandlerFunc
		method  string
	}{
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
	} {
		router.HandleFunc(
			route.path,

			promhttp.InstrumentHandlerDuration(
				duration.MustCurryWith(
					prometheus.Labels{
						"route":  route.path,
						"method": route.method,
					},
				),
				route.handler,
			),
		).Methods(route.method)
	}

	return &HTTPRouter{
		router,
		logger,
		port,
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
		Handler: r.handler,
	}

	go func() {
		<-ctx.Done()
		err := s.Shutdown(ctx)
		if err != nil {
			log.Println(err)
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
