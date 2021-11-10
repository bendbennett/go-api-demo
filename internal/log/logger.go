package log

import (
	"context"
	"fmt"

	"github.com/bendbennett/go-api-demo/internal/app"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

type Logger interface {
	Panic(err error)
	Panicf(msg string, args ...interface{})
	Error(err error)
	Errorf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	WithSpan(ctx context.Context) Logger
}

func NewLogger(prod bool) (logger, error) {
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
		return logger{}, err
	}

	return logger{
		zapLogger.With(
			zap.String("commit_hash", app.CommitHash())),
	}, nil
}

type logger struct {
	logger *zap.Logger
}

func (l logger) Panic(err error) {
	l.Panicf("%v", err)
}

func (l logger) Panicf(msg string, args ...interface{}) {
	l.logger.Panic(fmt.Sprintf(msg, args...))
}

func (l logger) Error(err error) {
	l.Errorf("%v", err)
}

func (l logger) Errorf(msg string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, args...))
}

func (l logger) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

func (l logger) WithSpan(ctx context.Context) Logger {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		spanLogger := spanLogger{
			logger: l.logger,
			span:   span,
		}

		if jaegerCtx, ok := span.Context().(jaeger.SpanContext); ok {
			spanLogger.spanFields = []zapcore.Field{
				zap.String("trace_id", jaegerCtx.TraceID().String()),
				zap.String("span_id", jaegerCtx.SpanID().String()),
			}
		}

		return spanLogger
	}

	return l
}
