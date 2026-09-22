package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment          string
	GatewayPort          string
	BiometricPort        string
	KYCPort              string
	BiometricServiceURL  string
	KYCServiceURL        string
	JWTSecret            string
	RedisHost            string
	RedisPort            string
	RedisPassword        string
	RedisDB              int
	PostgresHost         string
	PostgresPort         string
	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
	MaxWorkerPoolSize    int
	MaxQueueCapacity     int
	RateLimitRequests    int
	RateLimitWindowSec   int
	CircuitBreakerThresh int
}

func LoadConfig() *Config {
	return &Config{
		Environment:          getEnv("ENV", "production"),
		GatewayPort:          getEnv("GATEWAY_PORT", "8080"),
		BiometricPort:        getEnv("BIOMETRIC_PORT", "8081"),
		KYCPort:              getEnv("KYC_PORT", "8082"),
		BiometricServiceURL:  getEnv("BIOMETRIC_SERVICE_URL", "http://localhost:8081"),
		KYCServiceURL:        getEnv("KYC_SERVICE_URL", "http://localhost:8082"),
		JWTSecret:            getEnv("JWT_SECRET", "cbe-super-secret-kyc-biometric-key-2026"),
		RedisHost:            getEnv("REDIS_HOST", "localhost"),
		RedisPort:            getEnv("REDIS_PORT", "6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              getEnvAsInt("REDIS_DB", 0),
		PostgresHost:         getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:         getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:         getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:     getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:           getEnv("POSTGRES_DB", "kyc_biometric_db"),
		MaxWorkerPoolSize:    getEnvAsInt("WORKER_POOL_SIZE", 100),
		MaxQueueCapacity:     getEnvAsInt("WORKER_QUEUE_CAPACITY", 10000),
		RateLimitRequests:    getEnvAsInt("RATE_LIMIT_REQUESTS", 5000),
		RateLimitWindowSec:   getEnvAsInt("RATE_LIMIT_WINDOW_SEC", 1),
		CircuitBreakerThresh: getEnvAsInt("CIRCUIT_BREAKER_THRESH", 5),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return fallback
}
