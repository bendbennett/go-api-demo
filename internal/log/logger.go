package log

import (
	"context"
	"fmt"

	"github.com/bendbennett/go-api-demo/internal/app"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const callerSkip = 1

type Logger interface {
	Panic(err error)
	Panicf(msg string, args ...interface{})
	Error(err error)
	ErrorContext(ctx context.Context, err error)
	Errorf(msg string, args ...interface{})
	ErrorfContext(ctx context.Context, msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	InfofContext(ctx context.Context, msg string, args ...interface{})
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

func (l logger) ErrorContext(ctx context.Context, err error) {
	msg := fmt.Sprintf("%+v", err)
	span := spanFromContext(ctx)
	spanFields := span.spanFields()
	span.logToSpan("error", msg)
	l.logger.Error(msg, spanFields...)
}

func (l logger) Errorf(msg string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, args...))
}

func (l logger) ErrorfContext(ctx context.Context, msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	span := spanFromContext(ctx)
	spanFields := span.spanFields()
	span.logToSpan("error", m)
	l.logger.Error(m, spanFields...)
}

func (l logger) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

func (l logger) InfofContext(ctx context.Context, msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	span := spanFromContext(ctx)
	spanFields := span.spanFields()
	span.logToSpan("info", m)
	l.logger.Info(m, spanFields...)
}

type span struct {
	trace.Span
}

func spanFromContext(ctx context.Context) span {
	return span{
		trace.SpanFromContext(ctx),
	}
}

func (s span) spanFields() []zapcore.Field {
	var spanFields []zapcore.Field

	if s.Span != nil && s.IsRecording() {
		spanFields = append(spanFields, zap.String("trace_id", s.SpanContext().TraceID().String()))
		spanFields = append(spanFields, zap.String("span_id", s.SpanContext().SpanID().String()))
	}

	return spanFields
}

func (s span) logToSpan(level string, msg string) {
	if s.Span != nil && s.IsRecording() {
		s.SetAttributes(attribute.String("event", msg))
		s.SetAttributes(attribute.String("level", level))
	}
}
