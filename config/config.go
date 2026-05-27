package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Service   ServiceConfig
	HTTP      HTTPConfig
	Auth      AuthConfig
	RateLimit RateLimitConfig
	Infra     InfraConfig
}

type ServiceConfig struct {
	Name    string
	Version string
	Env     string
}

type HTTPConfig struct {
	Addr string
}

type AuthConfig struct {
	JWTSecret      string
	AccessTokenTTL time.Duration
	Issuer         string
}

type RateLimitConfig struct {
	PublicMaxRequests  int
	PrivateMaxRequests int
	OrderMaxRequests   int
	Window             time.Duration
}

type InfraConfig struct {
	Adapter  string
	Database DatabaseConfig
	Redis    RedisConfig
	MQ       MQConfig
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type MQConfig struct {
	Brokers []string
}

func Load() Config {
	return Config{
		Service: ServiceConfig{
			Name:    getEnv("SERVICE_NAME", "commerce-api"),
			Version: getEnv("SERVICE_VERSION", "v1"),
			Env:     getEnv("SERVICE_ENV", "local"),
		},
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8888"),
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
			AccessTokenTTL: 2 * time.Hour,
			Issuer:         getEnv("JWT_ISSUER", "commerce-api"),
		},
		RateLimit: RateLimitConfig{
			PublicMaxRequests:  120,
			PrivateMaxRequests: 60,
			OrderMaxRequests:   10,
			Window:             time.Minute,
		},
		Infra: InfraConfig{
			Adapter: getEnv("APP_ADAPTER", "memory"),
			Database: DatabaseConfig{
				DSN: getEnv("MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/commerce?charset=utf8mb4&parseTime=True&loc=Local"),
			},
			Redis: RedisConfig{
				Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       getEnvInt("REDIS_DB", 0),
			},
			MQ: MQConfig{
				Brokers: getEnvList("KAFKA_BROKERS", []string{"127.0.0.1:9092"}),
			},
		},
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvList(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
