package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config 应用配置结构
type Config struct {
	// 应用配置
	AppPort int    `json:"app_port"`
	AppEnv  string `json:"app_env"`
	Debug   bool   `json:"debug"`

	// 数据库配置
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	// Redis配置
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// JWT配置
	JWTSecret    string        `json:"jwt_secret"`
	JWTExpiresIn time.Duration `json:"jwt_expires_in"`

	// 加密配置
	EncryptionKey string `json:"encryption_key"`

	// 日志配置
	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`

	// Docker配置
	DockerHost string `json:"docker_host"`

	// pt-online-schema-change 默认参数
	PTDefaultChunkSize    int    `json:"pt_default_chunk_size"`
	PTDefaultMaxLoad      string `json:"pt_default_max_load"`
	PTDefaultCriticalLoad string `json:"pt_default_critical_load"`
}

// Load 加载配置
func Load() *Config {
	// 尝试加载.env文件
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	config := &Config{
		AppPort: getEnvAsInt("APP_PORT", 8090),
		AppEnv:  getEnv("APP_ENV", "development"),
		Debug:   getEnvAsBool("APP_DEBUG", true),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 3307),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "root123"),
		DBName:     getEnv("DB_NAME", "mysqler"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnvAsInt("REDIS_PORT", 6380),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		JWTSecret:    getEnv("JWT_SECRET", "mysqler-default-secret"),
		JWTExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),

		EncryptionKey: getEnv("ENCRYPTION_KEY", "mysqler-encryption-key-default-32char"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFile:  getEnv("LOG_FILE", ""),

		DockerHost: getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),

		PTDefaultChunkSize:    getEnvAsInt("PT_DEFAULT_CHUNK_SIZE", 1000),
		PTDefaultMaxLoad:      getEnv("PT_DEFAULT_MAX_LOAD", "Threads_running=25"),
		PTDefaultCriticalLoad: getEnv("PT_DEFAULT_CRITICAL_LOAD", "Threads_running=50"),
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration 获取环境变量并转换为时间间隔
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
