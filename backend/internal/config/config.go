package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
	JWT    JWTConfig
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host string
	Port string
}

type JWTConfig struct {
	Secret string
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host,
		db.Port,
		db.User,
		db.Password,
		db.Name,
	)
}

func Load() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}

	return &Config{
			Server: ServerConfig{
				Port: getEnvOrDefault("SERVER_PORT", "8080"),
			},
			DB: DBConfig{
				Host:     getEnvOrDefault("DB_HOST", "localhost"),
				Port:     getEnvOrDefault("DB_PORT", "5432"),
				User:     getEnvOrDefault("DB_USER", "streek"),
				Password: getEnvOrDefault("DB_PASSWORD", "streek"),
				Name:     getEnvOrDefault("DB_NAME", "streek"),
			},
			Redis: RedisConfig{
				Host: getEnvOrDefault("REDIS_HOST", "localhost"),
				Port: getEnvOrDefault("REDIS_PORT", "6379"),
			},
			JWT: JWTConfig{
				Secret: jwtSecret,
			},
		},
		nil
}

func getEnvOrDefault(envKey, defaultValue string) string {
	// if envKey is set, return its value, otherwise return defaultValue
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultValue
}
