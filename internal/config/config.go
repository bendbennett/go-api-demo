package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/Shopify/sarama"
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
	UserConsumerCache  KafkaConsumer
	UserConsumerSearch KafkaConsumer
	Kafka              Kafka
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
	Enabled bool
}

type Tracing struct {
	Enabled bool
}

type Logging struct {
	Production bool
}

type Message struct {
}

type Kafka struct {
	Brokers []string
	Version sarama.KafkaVersion
}

type KafkaConsumer struct {
	GroupID   string
	Topics    []string
	IsEnabled bool
}

// New returns a populated Config struct with values
// retrieved from environment variables or default
// values if the env vars do not exist.
func New() Config {
	_ = godotenv.Load()

	kafkaVersion, err := sarama.ParseKafkaVersion(
		GetEnvAsString(
			"KAFKA_VERSION",
			"",
		),
	)
	if err != nil {
		panic(err)
	}

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
			QueryTimeout: time.Millisecond * time.Duration(
				GetEnvAsInt(
					"STORAGE_QUERY_TIMEOUT",
					3000,
				),
			),
		},
		Metrics: Metrics{
			Enabled: GetEnvAsBool(
				"METRICS_ENABLED",
				true,
			),
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
		Kafka: Kafka{
			Version: kafkaVersion,
			Brokers: GetEnvAsSliceOfStrings(
				"KAFKA_BROKERS",
				",",
				[]string{},
			),
		},
		UserConsumerCache: KafkaConsumer{
			GroupID: GetEnvAsString(
				"KAFKA_USER_CONSUMER_CACHE_GROUP_ID",
				"",
			),
			Topics: GetEnvAsSliceOfStrings(
				"KAFKA_USER_CONSUMER_CACHE_TOPICS",
				",",
				[]string{},
			),
			IsEnabled: GetEnvAsBool(
				"KAFKA_USER_CONSUMER_CACHE_IS_ENABLED",
				false,
			),
		},
		UserConsumerSearch: KafkaConsumer{
			GroupID: GetEnvAsString(
				"KAFKA_USER_CONSUMER_SEARCH_GROUP_ID",
				"",
			),
			Topics: GetEnvAsSliceOfStrings(
				"KAFKA_USER_CONSUMER_SEARCH_TOPICS",
				",",
				[]string{},
			),
			IsEnabled: GetEnvAsBool(
				"KAFKA_USER_CONSUMER_SEARCH_IS_ENABLED",
				false,
			),
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
