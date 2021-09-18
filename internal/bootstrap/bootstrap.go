package bootstrap

import (
	"io"

	"github.com/bendbennett/go-api-demo/internal/storage/redis"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/routing"
	userconsume "github.com/bendbennett/go-api-demo/internal/user/consume"
	usercreate "github.com/bendbennett/go-api-demo/internal/user/create"
	userread "github.com/bendbennett/go-api-demo/internal/user/read"
	"github.com/bendbennett/go-api-demo/internal/validate"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
)

// New configures a logger for use throughout the application,
// retrieves configuration application, configures HTTP and
// gRPC routers, populates an app.App struct with the configured
// routers and returns a pointer to the populated app.App.
func New() *app.App {
	var closers []io.Closer

	conf := config.New()

	logger, err := NewLogger(conf.Logging.Production)
	if err != nil {
		panic(err)
	}

	tracer, closer, err := NewTracer(logger, conf.Tracing.Enabled)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

	httpPromVec, err := NewHTTPPromVec(conf.Metrics.Enabled)
	if err != nil {
		logger.Panic(err)
	}

	validator, err := validate.NewValidator()
	if err != nil {
		logger.Panic(err)
	}

	userStorage, closer, err := NewUserStorage(conf)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

	userCache, closer, err := redis.NewUserCache(conf.Redis)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

	userCreatePresenter := usercreate.NewPresenter()

	userCreateControllerHTTP := usercreate.NewHTTPController(
		validator,
		usercreate.NewInteractor(
			userStorage,
		),
		userCreatePresenter,
		logger,
	)

	userReadPresenter := userread.NewPresenter()

	userReadControllerHTTP := userread.NewHTTPController(
		userread.NewInteractor(
			userCache,
		),
		userReadPresenter,
		logger,
	)

	httpControllers := routing.HTTPControllers{
		UserCreateController: userCreateControllerHTTP.Create,
		UserReadController:   userReadControllerHTTP.Read,
	}

	httpRouter := routing.NewHTTPRouter(
		httpControllers,
		logger,
		httpPromVec,
		tracer,
		conf.HTTPPort,
	)

	userCreateControllerGRPC := usercreate.NewGRPCController(
		validator,
		usercreate.NewInteractor(
			userStorage,
		),
		userCreatePresenter,
		logger,
	)

	userReadControllerGRPC := userread.NewGRPCController(
		userread.NewInteractor(
			userCache,
		),
		userReadPresenter,
		logger,
	)

	grpcControllers := routing.GRPCControllers{
		UserCreate: userCreateControllerGRPC.Create,
		UserRead:   userReadControllerGRPC.Read,
	}

	grpcRouter := routing.NewGRPCRouter(
		grpcControllers,
		logger,
		conf.Metrics.Enabled,
		conf.Tracing.Enabled,
		conf.GRPCPort,
	)

	userProcessor := userconsume.NewProcessor(userCache)

	userConsumer, closer, err := userconsume.NewConsumer(
		conf.Kafka,
		userProcessor,
		logger,
	)
	if err != nil {
		logger.Panic(err)
	}
	closers = addCloser(closers, closer)

	return app.New(
		httpRouter,
		grpcRouter,
		userConsumer,
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

func NewLogger(prod bool) (log.Logger, error) {
	var (
		zapLogger *zap.Logger
		err       error
	)

	switch prod {
	case true:
		zapLogger, err = zap.NewProduction()
	default:
		zapLogger, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, err
	}

	logger := log.NewLogger(
		zapLogger.With(
			zap.String("commit_hash", app.CommitHash()),
		),
	)

	return logger, nil
}

// NewTracer toggles tracing on the basis of TRACING_ENABLED env var.
// If TRACING_ENABLED is true, the configuration and behaviour of the
// tracer is modified through JAEGER_... env vars.
func NewTracer(
	logger log.Logger,
	tracingEnabled bool,
) (opentracing.Tracer, io.Closer, error) {
	if !tracingEnabled {
		return opentracing.NoopTracer{}, nil, nil
	}

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		logger.Panic(err)
	}

	jaegerLogger := jaegerLoggerAdapter{logger}
	jaegerMetrics := jaegerprom.New()

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerLogger),
		jaegercfg.Metrics(jaegerMetrics),
	)
	if err != nil {
		return nil, nil, err
	}

	opentracing.SetGlobalTracer(tracer)

	return tracer, closer, nil
}

type jaegerLoggerAdapter struct {
	logger log.Logger
}

func (l jaegerLoggerAdapter) Error(msg string) {
	l.logger.Errorf(msg)
}

func (l jaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

func NewHTTPPromVec(metricsEnabled bool) (routing.HTTPPromVec, error) {
	if !metricsEnabled {
		return routing.HTTPPromVec{}, nil
	}

	histogramVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of latencies for HTTP requests.",
			Buckets: []float64{.001, .002, .005, .01, .02, .05, .1, .2, .5, 1, 2, 5},
		},
		[]string{"route", "method", "code"},
	)

	err := prometheus.Register(histogramVec)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return routing.HTTPPromVec{}, err
		}
	}

	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_server_handled_total",
			Help: "Total number of HTTP requests completed on the server, regardless of success or failure",
		},
		[]string{"route", "method", "code"},
	)

	err = prometheus.Register(counterVec)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return routing.HTTPPromVec{}, err
		}
	}

	return routing.HTTPPromVec{
		HistogramVec: histogramVec,
		CounterVec:   counterVec,
	}, nil
}
