package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string     `yaml:"env" env-default:"local"`
	StoragePath string     `yaml:"storage_path" env-required:"./storage"`
	Server      HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env-default:"localhost:8080"`
	Timeout      time.Duration `yaml:"timeout" env-default:"4s"`
	IddleTimeout time.Duration `yaml:"iddle_timeout" env-default:"60s"`
}

func MustLoad() Config {
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
