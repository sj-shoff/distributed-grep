package grep

import (
	"dgrep/internal/config"

	"github.com/rs/zerolog"
)

type Grep struct {
	cfg *config.Config
	log *zerolog.Logger
}

func New(cfg *config.Config, log *zerolog.Logger) *Grep {
	return &Grep{
		cfg: cfg,
		log: log,
	}
}
