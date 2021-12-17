package consume

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func NewConsumerProm(metricsEnabled bool) (Prom, error) {
	if !metricsEnabled {
		return Prom{}, nil
	}

	msgCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consumer_messages_total",
			Help: "Total number of Kafka messages consumed regardless of success or failure",
		},
		[]string{"type", "processor", "consumer"},
	)

	err := prometheus.Register(msgCounter)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return Prom{}, err
		}
	}

	queueLengthGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_queue_length",
			Help: "Queue length of Kafka consumer",
		},
		[]string{"type", "processor", "consumer"},
	)

	err = prometheus.Register(queueLengthGauge)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return Prom{}, err
		}
	}

	queueCapacityGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_queue_capacity",
			Help: "Queue capacity of Kafka consumer",
		},
		[]string{"type", "processor", "consumer"},
	)

	err = prometheus.Register(queueCapacityGauge)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return Prom{}, err
		}
	}

	lagGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_lag",
			Help: "Lag of Kafka consumer",
		},
		[]string{"type", "processor", "consumer"},
	)

	err = prometheus.Register(lagGauge)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return Prom{}, err
		}
	}

	return Prom{
		MsgCounter:         msgCounter,
		QueueLengthGauge:   queueLengthGauge,
		QueueCapacityGauge: queueCapacityGauge,
		LagGauge:           lagGauge,
	}, nil
}

func NewConsumerPromCollector(
	t, p string,
	collectInt time.Duration,
	prom Prom,
) PromCollector {
	return PromCollector{
		Type:      t,
		Processor: p,
		Interval:  collectInt,
		Prom:      prom,
	}
}

type PromCollector struct {
	Prom      Prom
	Type      string
	Processor string
	Interval  time.Duration
}

type Prom struct {
	MsgCounter         *prometheus.CounterVec
	QueueLengthGauge   *prometheus.GaugeVec
	QueueCapacityGauge *prometheus.GaugeVec
	LagGauge           *prometheus.GaugeVec
}

type Stats struct {
	Messages      int64
	QueueLength   int64
	QueueCapacity int64
	Lag           int64
}

func (c PromCollector) Update(stats Stats, consumer string) {
	if c.Prom.MsgCounter != nil {
		c.Prom.MsgCounter.WithLabelValues(c.Type, c.Processor, consumer).Add(float64(stats.Messages))
	}

	if c.Prom.QueueLengthGauge != nil {
		c.Prom.QueueLengthGauge.WithLabelValues(c.Type, c.Processor, consumer).Set(float64(stats.QueueLength))
	}

	if c.Prom.QueueCapacityGauge != nil {
		c.Prom.QueueCapacityGauge.WithLabelValues(c.Type, c.Processor, consumer).Set(float64(stats.QueueCapacity))
	}

	if c.Prom.LagGauge != nil {
		c.Prom.LagGauge.WithLabelValues(c.Type, c.Processor, consumer).Set(float64(stats.Lag))
	}
}
