package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"dgrep/internal/app"
	"dgrep/internal/config"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	if err := godotenv.Load(); err != nil {
		logger.Warn().Err(err).Msg("Failed to load .env file")
	}

	cfg := config.MustLoad()

	config.ParseFlags(cfg)

	application := app.NewApplication(cfg, &logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info().Msg("Received shutdown signal")
		cancel()
	}()

	if cfg.ServerMode {
		application.StartServer(ctx, cfg.Addr)
		return
	}

	args := cfg.Args
	if len(args) < 1 {
		logger.Fatal().Msg("Usage: dgrep [flags] pattern [file]")
	}
	pattern := args[0]

	var inputFile string
	if len(args) > 1 {
		inputFile = args[1]
	}

	application.RunClient(ctx, pattern, inputFile, os.Stdin)
}
