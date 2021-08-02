package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const StorageTypeMemory = "memory"
const StorageTypeSQL = "sql"

type Config struct {
	MySQL    *mysql.Config
	Storage  Storage
	Metrics  Metrics
	Tracing  Tracing
	Logging  Logging
	HTTPPort int
	GRPCPort int
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
		HTTPPort: GetEnvAsInt(
			"HTTP_PORT",
			3000,
		),
		GRPCPort: GetEnvAsInt(
			"GRPC_PORT",
			1234,
		),
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
