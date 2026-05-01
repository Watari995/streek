package config

import "fmt"

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
		"host=%s posrt=%s user= %s password=%s dbname=%s sslmode=disable",
		db.Host,
		db.Port,
		db.User,
		db.Password,
		db.Name,
	)
}
