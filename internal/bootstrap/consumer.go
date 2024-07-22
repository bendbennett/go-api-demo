package bootstrap

import (
	"fmt"
	"io"

	"github.com/bendbennett/go-api-demo/internal/schema"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/consume"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/metrics"
	"github.com/bendbennett/go-api-demo/internal/user"
	userconsume "github.com/bendbennett/go-api-demo/internal/user/consume"
)

func newConsumers(
	conf config.Config,
	logger log.Logger,
	userCache user.CreatorReader,
	userSearch user.CreatorSearcher,
) ([]app.Component, []io.Closer, error) {
	var (
		components []app.Component
		closers    []io.Closer
	)

	err := consume.CreateTopics(conf.TopicConfigs)
	if err != nil {
		return nil, nil, err
	}

	schemaClient := schema.NewClient(conf.SchemaRegistry.ClientTimeout)
	userDecoder, err := schemaClient.GetDecoder(
		fmt.Sprintf(
			"%s%s",
			conf.SchemaRegistry.Domain,
			conf.SchemaRegistry.Endpoints["usersValue"],
		),
	)
	if err != nil {
		return nil, nil, err
	}

	userConsumerMetrics, err := metrics.NewConsumerMetrics(conf.Telemetry.Enabled)
	if err != nil {
		panic(err)
	}

	userConsumerMetricsLabelsCache := metrics.NewConsumerMetricsLabels(
		"user",
		"cache",
	)

	userConsumerMetricsCollectorCache := metrics.NewConsumerMetricsCollector(
		userConsumerMetrics,
		userConsumerMetricsLabelsCache,
	)

	userProcessorCache := userconsume.NewProcessor(userCache)

	consumers, closrs, err := consume.NewConsumers(
		conf.UserConsumerCache,
		conf.Telemetry.Enabled,
		userConsumerMetricsLabelsCache,
		userConsumerMetricsCollectorCache,
		userProcessorCache,
		userDecoder,
		logger,
	)

	if err != nil {
		return nil, nil, err
	}

	for _, consumer := range consumers {
		components = append(components, consumer)
	}

	closers = addCloser(closers, closrs...)

	userConsumerMetricsLabelsSearch := metrics.NewConsumerMetricsLabels(
		"user",
		"search",
	)

	userConsumerMetricsCollectorSearch := metrics.NewConsumerMetricsCollector(
		userConsumerMetrics,
		userConsumerMetricsLabelsSearch,
	)

	userProcessorSearch := userconsume.NewProcessor(userSearch)

	consumers, closrs, err = consume.NewConsumers(
		conf.UserConsumerSearch,
		conf.Telemetry.Enabled,
		userConsumerMetricsLabelsSearch,
		userConsumerMetricsCollectorSearch,
		userProcessorSearch,
		userDecoder,
		logger,
	)

	if err != nil {
		return nil, nil, err
	}

	for _, consumer := range consumers {
		components = append(components, consumer)
	}

	closers = addCloser(closers, closrs...)

	return components, closers, nil
}
