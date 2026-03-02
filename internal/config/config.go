package config

import (
	"fmt"
	"time"

	"dgrep/internal/domain/errors"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddr    string        `env:"DGREP_SERVER_ADDR" env-default:":8080" validate:"required"`
	DefaultAddrs  string        `env:"DGREP_DEFAULT_ADDRS" env-default:""`
	Timeout       time.Duration `env:"DGREP_TIMEOUT" env-default:"10s" validate:"required"`
	NumGoroutines int           `env:"DGREP_NUM_GOROUTINES" env-default:"4" validate:"gte=1"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Errorf("%w: %v", errors.ErrInvalidConfig, err))
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		panic(fmt.Errorf("%w: %v", errors.ErrInvalidConfig, err))
	}

	return &cfg
}
