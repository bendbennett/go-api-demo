package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

const StorageTypeMemory = "memory"
const StorageTypeSQL = "sql"

type Config struct {
	MySQL              *mysql.Config
	Storage            Storage
	Redis              redis.Options
	Elasticsearch      elasticsearch.Config
	TopicConfigs       TopicConfigs
	SchemaRegistry     SchemaRegistry
	Telemetry          Telemetry
	UserConsumerCache  KafkaConsumer
	UserConsumerSearch KafkaConsumer
	Logging            Logging
	HTTP               HTTP
	GRPCPort           int
}

type HTTP struct {
	Port              int
	ReadHeaderTimeout time.Duration
}

type Storage struct {
	Type         string
	QueryTimeout time.Duration
}

type Telemetry struct {
	ServiceName               string
	ExporterTargetEndPoint    string
	MetricsCollectionInterval time.Duration
	Enabled                   bool
}

type Logging struct {
	Production bool
}

type Message struct {
}

type KafkaConsumer struct {
	ReaderConfig kafka.ReaderConfig
	IsEnabled    bool
	Num          int
}

type TopicConfigs struct {
	Brokers []string
	Conf    []kafka.TopicConfig
}

type SchemaRegistry struct {
	Endpoints     map[string]string
	Domain        string
	ClientTimeout time.Duration
}

// New returns a populated Config struct with values
// retrieved from environment variables or default
// values if the env vars do not exist.
func New() Config {
	_ = godotenv.Load()

	return Config{
		MySQL: &mysql.Config{
			User: GetEnvAsString(
				"MYSQL_USER",
				"",
			),
			Passwd: GetEnvAsString(
				"MYSQL_PASSWORD",
				"",
			),
			Net: "tcp",
			Addr: fmt.Sprintf(
				"%s:%d",
				GetEnvAsString(
					"MYSQL_HOST",
					"localhost"),
				GetEnvAsInt(
					"MYSQL_PORT",
					3306),
			),
			DBName: GetEnvAsString(
				"MYSQL_DBNAME",
				"",
			),
			ParseTime:            true,
			AllowNativePasswords: true,
			InterpolateParams:    true,
		},
		Storage: Storage{
			Type: GetEnvAsString(
				"STORAGE_TYPE",
				StorageTypeMemory,
			),
			QueryTimeout: GetEnvAsDuration(
				"STORAGE_QUERY_TIMEOUT",
				3*time.Second,
			),
		},
		Telemetry: Telemetry{
			ServiceName: GetEnvAsString(
				"TELEMETRY_SERVICE_NAME",
				"go-api-demo",
			),
			// OTEL_EXPORTER_OTLP_ENDPOINT is env var for base endpoint URL for any signal type.
			// https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/
			ExporterTargetEndPoint: GetEnvAsString(
				"TELEMETRY_EXPORTER_TARGET_ENDPOINT",
				"0.0.0.0:4317",
			),
			MetricsCollectionInterval: GetEnvAsDuration(
				"TELEMETRY_METRICS_COLLECTION_INTERVAL",
				5*time.Second,
			),
			Enabled: GetEnvAsBool(
				"TELEMETRY_ENABLED",
				true,
			),
		},
		Logging: Logging{
			Production: GetEnvAsBool(
				"LOGGING_PRODUCTION",
				false,
			),
		},
		UserConsumerCache: KafkaConsumer{
			ReaderConfig: kafka.ReaderConfig{
				Brokers: GetEnvAsSliceOfStrings(
					"KAFKA_BROKERS",
					",",
					[]string{},
				),
				GroupID: GetEnvAsString(
					"KAFKA_USER_CONSUMER_CACHE_GROUP_ID",
					"",
				),
				MaxBytes: GetEnvAsInt(
					"KAFKA_USER_CONSUMER_CACHE_MAX_BYTES",
					200e3,
				),
				MaxWait: GetEnvAsDuration(
					"KAFKA_USER_CONSUMER_CACHE_MAX_WAIT",
					30*time.Second,
				),
				RebalanceTimeout: GetEnvAsDuration(
					"KAFKA_USER_CONSUMER_CACHE_REBALANCE_TIMEOUT",
					30*time.Second,
				),
				Topic: GetEnvAsString(
					"KAFKA_USER_CONSUMER_CACHE_TOPIC",
					"",
				),
			},
			IsEnabled: GetEnvAsBool(
				"KAFKA_USER_CONSUMER_CACHE_IS_ENABLED",
				false,
			),
			Num: GetEnvAsInt(
				"KAFKA_USER_CONSUMER_CACHE_NUM",
				1,
			),
		},
		UserConsumerSearch: KafkaConsumer{
			ReaderConfig: kafka.ReaderConfig{
				Brokers: GetEnvAsSliceOfStrings(
					"KAFKA_BROKERS",
					",",
					[]string{},
				),
				GroupID: GetEnvAsString(
					"KAFKA_USER_CONSUMER_SEARCH_GROUP_ID",
					"",
				),
				MaxBytes: GetEnvAsInt(
					"KAFKA_USER_CONSUMER_SEARCH_MAX_BYTES",
					200e3,
				),
				MaxWait: GetEnvAsDuration(
					"KAFKA_USER_CONSUMER_SEARCH_MAX_WAIT",
					30*time.Second,
				),
				RebalanceTimeout: GetEnvAsDuration(
					"KAFKA_USER_CONSUMER_SEARCH_REBALANCE_TIMEOUT",
					30*time.Second,
				),
				Topic: GetEnvAsString(
					"KAFKA_USER_CONSUMER_SEARCH_TOPIC",
					"",
				),
			},
			IsEnabled: GetEnvAsBool(
				"KAFKA_USER_CONSUMER_SEARCH_IS_ENABLED",
				false,
			),
			Num: GetEnvAsInt(
				"KAFKA_USER_CONSUMER_SEARCH_NUM",
				2,
			),
		},
		TopicConfigs: TopicConfigs{
			Brokers: GetEnvAsSliceOfStrings(
				"KAFKA_BROKERS",
				",",
				[]string{},
			),
			Conf: []kafka.TopicConfig{
				{
					Topic:             "mysql.go_api_demo.users",
					NumPartitions:     2,
					ReplicationFactor: 1,
					ConfigEntries: []kafka.ConfigEntry{
						{
							ConfigName:  "cleanup.policy",
							ConfigValue: "compact",
						},
						{
							ConfigName:  "compression.type",
							ConfigValue: "producer",
						},
					},
				},
			},
		},
		SchemaRegistry: SchemaRegistry{
			ClientTimeout: GetEnvAsDuration(
				"KAFKA_SCHEMA_REGISTRY_CLIENT_TIMEOUT",
				3*time.Second),
			Domain: fmt.Sprintf(
				"%v://%v:%v",
				GetEnvAsString(
					"KAFKA_SCHEMA_REGISTRY_PROTOCOL",
					"http",
				),
				GetEnvAsString(
					"KAFKA_SCHEMA_REGISTRY_DOMAIN",
					"localhost",
				),
				GetEnvAsString(
					"KAFKA_SCHEMA_REGISTRY_PORT",
					"8081",
				),
			),
			Endpoints: map[string]string{
				"usersValue": fmt.Sprintf(
					"/subjects/%v/versions/%v",
					GetEnvAsString(
						"KAFKA_SCHEMA_REGISTRY_USERS_VALUE",
						"mysql.go_api_demo.users-value",
					),
					GetEnvAsString(
						"KAFKA_SCHEMA_REGISTRY_USERS_VALUE_VERSION",
						"1",
					),
				),
			},
		},
		HTTP: HTTP{
			Port: GetEnvAsInt(
				"HTTP_PORT",
				3000,
			),
			ReadHeaderTimeout: GetEnvAsDuration(
				"HTTP_READ_HEADER_TIMEOUT",
				10*time.Second,
			),
		},
		GRPCPort: GetEnvAsInt(
			"GRPC_PORT",
			1234,
		),
		Redis: redis.Options{
			Addr: fmt.Sprintf("%s:%v",
				GetEnvAsString(
					"REDIS_HOST",
					"localhost",
				),
				GetEnvAsInt(
					"REDIS_PORT",
					6379,
				),
			),
			Password: GetEnvAsString(
				"REDIS_PASSWORD",
				"pass",
			),
		},
		Elasticsearch: elasticsearch.Config{
			Addresses: GetEnvAsSliceOfStrings(
				"ELASTICSEARCH_ADDRESSES",
				",",
				[]string{},
			),
		},
	}
}

func GetEnvAsInt(
	key string,
	defaultVal int,
) int {
	if val, ok := os.LookupEnv(key); ok {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		return intVal
	}

	return defaultVal
}

func GetEnvAsString(
	key string,
	defaultVal string,
) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func GetEnvAsBool(
	key string,
	defaultVal bool,
) bool {
	if val, ok := os.LookupEnv(key); ok {
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			panic(err)
		}

		return boolVal
	}

	return defaultVal
}

func GetEnvAsSliceOfStrings(
	key string,
	separator string,
	defaultVal []string,
) []string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.Split(val, separator)
	}

	return defaultVal
}

func GetEnvAsDuration(
	key string,
	defaultVal time.Duration,
) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(fmt.Sprintf("%v", val))
		if err != nil {
			panic(err)
		}

		return d
	}

	return defaultVal
}
