package bootstrap

import (
	"fmt"
	"io"

	"github.com/bendbennett/go-api-demo/internal/schema"

	"github.com/bendbennett/go-api-demo/internal/app"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/consume"
	"github.com/bendbennett/go-api-demo/internal/log"
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
		panic(err)
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

	userConsumerProm, err := consume.NewConsumerProm(conf.Metrics.Enabled)
	if err != nil {
		panic(err)
	}

	userConsumerPromUserCache := consume.NewConsumerPromCollector(
		"user",
		"cache",
		conf.Metrics.CollectionInterval,
		userConsumerProm,
	)

	userProcessorCache := userconsume.NewProcessor(userCache)

	consumers, closrs := userconsume.NewConsumers(
		conf.UserConsumerCache,
		conf.Tracing.Enabled,
		userConsumerPromUserCache,
		userProcessorCache,
		userDecoder,
		logger,
	)

	for _, consumer := range consumers {
		components = append(components, consumer)
	}

	closers = addCloser(closers, closrs...)

	userConsumerPromUserSearch := consume.NewConsumerPromCollector(
		"user",
		"search",
		conf.Metrics.CollectionInterval,
		userConsumerProm,
	)

	userProcessorSearch := userconsume.NewProcessor(userSearch)

	consumers, closrs = userconsume.NewConsumers(
		conf.UserConsumerSearch,
		conf.Tracing.Enabled,
		userConsumerPromUserSearch,
		userProcessorSearch,
		userDecoder,
		logger,
	)

	for _, consumer := range consumers {
		components = append(components, consumer)
	}

	closers = addCloser(closers, closrs...)

	return components, closers, nil
}
