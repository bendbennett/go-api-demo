package log

import (
	"context"
	"fmt"

	"github.com/bendbennett/go-api-demo/internal/app"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const callerSkip = 1

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
		zapLogger, err = zap.NewProduction(
			zap.AddCallerSkip(
				callerSkip,
			),
		)
	default:
		zapLogger, err = zap.NewDevelopment(
			zap.AddCallerSkip(
				callerSkip,
			),
		)
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
	l.Panicf("%+v", err)
}

func (l logger) Panicf(msg string, args ...interface{}) {
	l.logger.Panic(fmt.Sprintf(msg, args...))
}

func (l logger) Error(err error) {
	l.Errorf("%+v", err)
}

func (l logger) Errorf(msg string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, args...))
}

func (l logger) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

func (l logger) WithSpan(ctx context.Context) Logger {
	// span.IsRecording() determines whether span is noopSpan,
	// which is generated when tracing has not been configured.
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		spanLogger := spanLogger{
			logger: l.logger,
			span:   span,
		}

		spanLogger.spanFields = []zapcore.Field{
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		}

		return spanLogger
	}

	return l
}
