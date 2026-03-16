package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env                   string          `yaml:"env" env-default:"local"`
	DBConfig              DBConfig        `yaml:"db"`
	RedisConfig           RedisConfig     `yaml:"redis"`
	MailHogConfig         MailHogConfig   `yaml:"mailhog"`
	SchedulerConfig       SchedulerConfig `yaml:"scheduler"`
	WorkerConfig          WorkerConfig    `yaml:"worker"`
	Server                HTTPServer      `yaml:"server"`
	DefaultTaskMaxRetries int             `yaml:"default-max-retries" env-default:"3"`
}

type WorkerConfig struct {
	MaxConcurrency int           `yaml:"max-concurrency" env-default:"10"`
	LockTimeout    time.Duration `yaml:"lock-timeout" env-default:"5m"`
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
	Address  string `yaml:"address" env:"REDIS_ADDR"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       string `yaml:"database" env:"REDIS_DATABASE"`
}

type MailHogConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type SchedulerConfig struct {
	StaleTaskThreshold time.Duration `yaml:"stale-task-threshold" env:"STALE_TASK_THRESHOLD"`
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
