package config

import (
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/spf13/viper"
)

var (
	instance *Config
	once     sync.Once
)

type Config struct {
	Server   *ServerConfig   `mapstructure:"server"`
	Log      *LogConfig      `mapstructure:"log"`
	Postgres *PostgresConfig `mapstructure:"postgres"`
	Security *SecurityConfig `mapstructure:"security"`
	Redis    *RedisConfig    `mapstructure:"redis"`
	RabbitMQ *RabbitMQConfig `mapstructure:"rabbitmq"`
	Email    *EmailConfig    `mapstructure:"email"`
}

func Instance() *Config {
	once.Do(func() {
		instance = loadEnvConfig()
	})

	return instance
}

func loadEnvConfig() *Config {
	env := utils.GetEnvFromArgs("develop")

	viper := viper.New()
	viper.AddConfigPath("./config/env/")
	viper.SetConfigName(env)
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	if cfg.Server.Env == "" {
		cfg.Server.Env = env
	}

	return &cfg
}
