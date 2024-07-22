package telemetry

import (
	"errors"
	"time"

	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/log"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/net/context"
)

type telemetry struct {
	logger log.Logger
	conf   config.Telemetry
}

func NewTelemetry(
	logger log.Logger,
	conf config.Telemetry,
) (*telemetry, error) {
	return &telemetry{
		logger: logger,
		conf:   conf,
	}, nil
}

func (t *telemetry) Run(ctx context.Context) error {
	var meterProvider *sdkmetric.MeterProvider
	var tracerProvider *sdktrace.TracerProvider

	if t.conf.Enabled {
		resource, err := resource.New(ctx,
			resource.WithFromEnv(),
			resource.WithProcess(),
			resource.WithTelemetrySDK(),
			resource.WithHost(),
			resource.WithAttributes(
				// Service name used in traces and metrics (exported_job).
				semconv.ServiceNameKey.String(t.conf.ServiceName),
			),
		)
		if err != nil {
			return err
		}

		if meterProvider, err = t.meterProvider(ctx, resource); err != nil {
			return err
		}

		if tracerProvider, err = t.tracerProvider(ctx, resource); err != nil {
			return err
		}
	}

	<-ctx.Done()

	var providerShutdownErr error

	if tracerProvider != nil {
		err := tracerProvider.Shutdown(ctx)

		if !errors.Is(err, context.Canceled) {
			t.logger.Error(err)
			providerShutdownErr = errors.Join(err)
		}

		t.logger.Infof(err.Error())
	}

	if meterProvider != nil {
		err := meterProvider.Shutdown(ctx)

		if !errors.Is(err, context.Canceled) {
			t.logger.Error(err)
			providerShutdownErr = errors.Join(providerShutdownErr, err)
		}

		t.logger.Infof(err.Error())
	}

	return providerShutdownErr
}

func (t *telemetry) meterProvider(
	ctx context.Context,
	resource *resource.Resource,
) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(t.conf.ExporterTargetEndPoint),
	)

	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(2*time.Second),
			),
		),
	)

	// Registers meterProvider as the global meter provider.
	otel.SetMeterProvider(meterProvider)

	// Add Go runtime metrics.
	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))

	return meterProvider, err
}

func (t *telemetry) tracerProvider(
	ctx context.Context,
	resource *resource.Resource,
) (*sdktrace.TracerProvider, error) {
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(t.conf.ExporterTargetEndPoint))

	traceExporter, err := otlptrace.New(ctx, traceClient)

	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set global propagator to TraceContext (the default is no-op).
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Registers tracerProvider as the global trace provider.
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}
