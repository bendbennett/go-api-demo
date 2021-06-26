package routing

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type HTTPRouter struct {
	handler http.Handler
	logger  *log.Entry
	port    int
}

type HTTPControllers struct {
	UserCreateController func(w http.ResponseWriter, r *http.Request)
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
	} {
		router.HandleFunc(route.path, route.handler).Methods(route.method)
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
