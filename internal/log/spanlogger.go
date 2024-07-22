package log

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type spanLogger struct {
	logger     *zap.Logger
	span       trace.Span
	spanFields []zapcore.Field
}

func (sl spanLogger) Panic(err error) {
	msg := fmt.Sprintf("%+v", err)
	sl.logToSpan("panic", msg)
	sl.logger.Panic(msg, sl.spanFields...)
}

func (sl spanLogger) Panicf(msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	sl.logToSpan("panic", m)
	sl.logger.Panic(m, sl.spanFields...)
}

func (sl spanLogger) Error(err error) {
	msg := fmt.Sprintf("%+v", err)
	sl.logToSpan("error", msg)
	sl.logger.Error(msg, sl.spanFields...)
}

func (sl spanLogger) Errorf(msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	sl.logToSpan("error", m)
	sl.logger.Error(m, sl.spanFields...)
}

func (sl spanLogger) Infof(msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	sl.logToSpan("info", m)
	sl.logger.Info(m, sl.spanFields...)
}

func (sl spanLogger) WithSpan(context.Context) Logger {
	return sl
}

func (sl spanLogger) logToSpan(level string, msg string) {
	if sl.span == nil {
		return
	}

	sl.span.SetAttributes(attribute.String("event", msg))
	sl.span.SetAttributes(attribute.String("level", level))
}
