package config

import (
	"fmt"
	"time"

	"dgrep/internal/domain/errors"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type GrepOptions struct {
	Fixed      bool `validate:"omitempty"`
	IgnoreCase bool `validate:"omitempty"`
	Invert     bool `validate:"omitempty"`
	LineNum    bool `validate:"omitempty"`
	After      int  `validate:"gte=0"`
	Before     int  `validate:"gte=0"`
	Context    int  `validate:"gte=0"`
	Count      bool `validate:"omitempty"`
}

type Config struct {
	ServerAddr    string        `env:"MYGREP_SERVER_ADDR" env-default:":8042" validate:"required"`
	DefaultAddrs  string        `env:"MYGREP_DEFAULT_ADDRS" env-default:""`
	Timeout       time.Duration `env:"MYGREP_TIMEOUT" env-default:"10s" validate:"required"`
	NumGoroutines int           `env:"MYGREP_NUM_GOROUTINES" env-default:"4" validate:"gte=1"`
	ServerMode    bool
	Addr          string
	Addrs         string
	Args          []string
	GrepOptions   GrepOptions `validate:"required"`
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
