package trace

import (
	"io"

	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerprom "github.com/uber/jaeger-lib/metrics/prometheus"
)

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
