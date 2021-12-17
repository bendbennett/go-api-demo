package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const StorageTypeMemory = "memory"
const StorageTypeSQL = "sql"

type Config struct {
	MySQL              *mysql.Config
	Storage            Storage
	Redis              redis.Options
	Elasticsearch      elasticsearch.Config
	TopicConfigs       TopicConfigs
	UserConsumerCache  KafkaConsumer
	UserConsumerSearch KafkaConsumer
	Metrics            Metrics
	Tracing            Tracing
	Logging            Logging
	HTTPPort           int
	GRPCPort           int
}

type Storage struct {
	Type         string
	QueryTimeout time.Duration
}

type Metrics struct {
	Enabled            bool
	CollectionInterval time.Duration
}

type Tracing struct {
	Enabled bool
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
				GetEnvAsInt(""+
					"MYSQL_PORT",
					3306),
			),
			DBName: GetEnvAsString(
				"MYSQL_DBNAME",
				"",
			),
			ParseTime:            true,
			AllowNativePasswords: true,
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
		Metrics: Metrics{
			Enabled: GetEnvAsBool(
				"METRICS_ENABLED",
				true,
			),
			CollectionInterval: GetEnvAsDuration(
				"METRICS_COLLECTION_INTERVAL",
				5*time.Second),
		},
		Tracing: Tracing{
			Enabled: GetEnvAsBool(
				"TRACING_ENABLED",
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
					Topic:             "go_api_demo_db.go-api-demo.users",
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
		HTTPPort: GetEnvAsInt(
			"HTTP_PORT",
			3000,
		),
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
