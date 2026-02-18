package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string     `yaml:"env" env-default:"local"`
	DBConfig DBConfig `yaml:"db"`
	RedisConfig RedisConfig `yaml:"redis"`
	Server      HTTPServer `yaml:"server"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env-default:"localhost:8080"`
	Timeout      time.Duration `yaml:"timeout" env-default:"4s"`
	IddleTimeout time.Duration `yaml:"iddle_timeout" env-default:"60s"`
}

type DBConfig struct {
	Url string `yaml:"url" env:"DATABASE_URL"`
}

type RedisConfig struct {
	Address string `yaml:"address" env:"REDIS_ADDR"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB string `yaml:"database" env:"REDIS_DATABASE"`
}

func MustLoad() Config {
	err := godotenv.Load()
	if err != nil {
		panic(".env file couldn't be loaded")
	}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("config file not found")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("error reading config file:" + err.Error())
	}
	return cfg
}
