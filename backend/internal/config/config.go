package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server       ServerConfig
	DB           DBConfig
	Redis        RedisConfig
	JWT          JWTConfig
	Notification NotificationConfig
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

type NotificationConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	To           string
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

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func (n NotificationConfig) IsSMTPEnabled() bool {
	return n.SMTPHost != "" && n.SMTPPort != "" && n.SMTPUser != "" && n.SMTPPassword != "" && n.SMTPFrom != "" && n.To != ""
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
			Notification: NotificationConfig{
				SMTPHost:     getEnvOrDefault("SMTP_HOST", ""),
				SMTPPort:     getEnvOrDefault("SMTP_PORT", "587"),
				SMTPUser:     getEnvOrDefault("SMTP_USER", ""),
				SMTPPassword: getEnvOrDefault("SMTP_PASSWORD", ""),
				SMTPFrom:     getEnvOrDefault("SMTP_FROM", ""),
				To:           getEnvOrDefault("NOTIFICATION_TO", ""),
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
