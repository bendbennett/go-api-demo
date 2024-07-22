package metrics

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ConsumerMetricsLabels struct {
	entityType  string
	destination string
}

func NewConsumerMetricsLabels(
	entityType string,
	destination string,
) ConsumerMetricsLabels {
	return ConsumerMetricsLabels{
		entityType:  entityType,
		destination: destination,
	}
}

func (c ConsumerMetricsLabels) EntityType() string {
	return c.entityType
}

func (c ConsumerMetricsLabels) Destination() string {
	return c.destination
}

type ConsumerMetrics struct {
	msgCounter         metric.Int64ObservableCounter
	queueLengthGauge   metric.Int64ObservableGauge
	queueCapacityGauge metric.Int64ObservableGauge
	lagGauge           metric.Int64ObservableGauge
	totalMsgCount      int64
}

func NewConsumerMetrics(telemetryEnabled bool) (ConsumerMetrics, error) {
	if !telemetryEnabled {
		return ConsumerMetrics{}, nil
	}

	meter := otel.Meter("")

	msgCounter, err := meter.Int64ObservableCounter(
		"kafka_consumer_messages_total",
		metric.WithDescription("Total number of Kafka messages consumed regardless of success or failure"),
	)

	if err != nil {
		return ConsumerMetrics{}, err
	}

	queueLengthGauge, err := meter.Int64ObservableGauge(
		"kafka_consumer_queue_length",
		metric.WithDescription("Queue length of Kafka consumer"),
	)

	if err != nil {
		return ConsumerMetrics{}, err
	}

	queueCapacityGauge, err := meter.Int64ObservableGauge(
		"kafka_consumer_queue_capacity",
		metric.WithDescription("Queue capacity of Kafka consumer"),
	)

	if err != nil {
		return ConsumerMetrics{}, err
	}

	lagGauge, err := meter.Int64ObservableGauge(
		"kafka_consumer_lag",
		metric.WithDescription("Lag of Kafka consumer"),
	)

	if err != nil {
		return ConsumerMetrics{}, err
	}

	return ConsumerMetrics{
		msgCounter:         msgCounter,
		queueLengthGauge:   queueLengthGauge,
		queueCapacityGauge: queueCapacityGauge,
		lagGauge:           lagGauge,
	}, nil
}

type ConsumerMetricsCollector struct {
	metrics       ConsumerMetrics
	metricsLabels ConsumerMetricsLabels
}

func NewConsumerMetricsCollector(
	metrics ConsumerMetrics,
	metricsLabels ConsumerMetricsLabels,
) ConsumerMetricsCollector {
	return ConsumerMetricsCollector{
		metrics:       metrics,
		metricsLabels: metricsLabels,
	}
}

type statsFunc func() kafka.ReaderStats

func (m ConsumerMetricsCollector) RegisterMetrics(telemetryEnabled bool, statsFunc statsFunc, groupID string) error {
	if !telemetryEnabled {
		return nil
	}

	meter := otel.Meter("")

	metricMeasurementOptions := metric.WithAttributes(
		attribute.String("consumer", groupID),
		attribute.String("entity_type", m.metricsLabels.EntityType()),
		attribute.String("destination", m.metricsLabels.Destination()),
	)

	_, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			stats := statsFunc()

			// Maintain a total message count as statsFunc().Messages
			// returns count of messages since last snapshot.
			// A snapshot is generated each time Stats() is called.
			m.metrics.totalMsgCount += stats.Messages

			o.ObserveInt64(
				m.metrics.msgCounter,
				m.metrics.totalMsgCount,
				metricMeasurementOptions,
			)

			o.ObserveInt64(
				m.metrics.queueLengthGauge,
				stats.QueueLength,
				metricMeasurementOptions,
			)

			o.ObserveInt64(
				m.metrics.queueCapacityGauge,
				stats.QueueCapacity,
				metricMeasurementOptions,
			)

			o.ObserveInt64(
				m.metrics.lagGauge,
				stats.Lag,
				metricMeasurementOptions,
			)

			return nil
		},
		m.metrics.msgCounter,
		m.metrics.queueLengthGauge,
		m.metrics.queueCapacityGauge,
		m.metrics.lagGauge,
	)

	return err
}
