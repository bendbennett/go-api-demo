package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort int
	GRPCPort int
}

// New returns a populated Config struct with values
// retrieved from environment variables or default
// values if the env vars do not exist.
func New() Config {
	return Config{
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
