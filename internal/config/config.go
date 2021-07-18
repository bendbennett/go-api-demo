package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const StorageMemory = "memory"
const StorageSQL = "sql"

type Config struct {
	MySQL    *mysql.Config
	Storage  Storage
	HTTPPort int
	GRPCPort int
}

type Storage struct {
	UserStorage string
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
			UserStorage: GetEnvAsString(
				"USER_STORAGE",
				StorageMemory),
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