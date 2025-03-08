package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config содержит все настройки приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Minio    MinioConfig
	Kafka    KafkaConfig
	Auth     AuthConfig
	Storage  StorageConfig
}

// ServerConfig настройки сервера
type ServerConfig struct {
	Port    int
	Timeout time.Duration
	Debug   bool
}

// DatabaseConfig настройки базы данных
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig настройки Redis
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// MinioConfig настройки MinIO
type MinioConfig struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	VideosBucket string
}

// KafkaConfig настройки Kafka
type KafkaConfig struct {
	Brokers              []string
	VideoProcessingTopic string
}

// AuthConfig настройки аутентификации
type AuthConfig struct {
	AccessTokenSecret    string
	RefreshTokenSecret   string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// StorageConfig настройки хранилища
type StorageConfig struct {
	ShardCount int
	BaseURL    string
	UseSSL     bool
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Загружаем .env файл, если он существует
	// Игнорируем ошибку, если файл отсутствует
	_ = godotenv.Load()

	// Создаем конфигурацию
	cfg := &Config{
		Server: ServerConfig{
			Port:    getEnvAsInt("SERVER_PORT", 8080),
			Timeout: getEnvAsDuration("SERVER_TIMEOUT", 30*time.Second),
			Debug:   getEnvAsBool("SERVER_DEBUG", false),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "video_hosting"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
		Minio: MinioConfig{
			Endpoint:     getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:    getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:    getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:       getEnvAsBool("MINIO_USE_SSL", false),
			VideosBucket: getEnv("MINIO_VIDEOS_BUCKET", "video-hosting-videos"),
		},
		Kafka: KafkaConfig{
			Brokers:              getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			VideoProcessingTopic: getEnv("KAFKA_VIDEO_PROCESSING_TOPIC", "video-processing"),
		},
		Auth: AuthConfig{
			AccessTokenSecret:    getEnv("AUTH_ACCESS_TOKEN_SECRET", "your_access_token_secret_key"),
			RefreshTokenSecret:   getEnv("AUTH_REFRESH_TOKEN_SECRET", "your_refresh_token_secret_key"),
			AccessTokenDuration:  getEnvAsDuration("AUTH_ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDuration("AUTH_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		},
		Storage: StorageConfig{
			ShardCount: getEnvAsInt("STORAGE_SHARD_COUNT", 64),
			BaseURL:    getEnv("STORAGE_BASE_URL", "http://localhost:9000"),
			UseSSL:     getEnvAsBool("STORAGE_BASE_USE_SSL", false),
		},
	}

	return cfg, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает числовое значение переменной окружения
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool получает логическое значение переменной окружения
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsDuration получает значение продолжительности из переменной окружения
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsSlice получает срез строк из переменной окружения
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
