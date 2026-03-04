package app

import (
	"context"
	"io"

	"github.com/rs/zerolog"

	"dgrep/internal/adapters/cli"
	"dgrep/internal/adapters/grpc"
	"dgrep/internal/config"
	"dgrep/internal/domain"
	grepUC "dgrep/internal/usecase/grep"
)

type Application struct {
	cfg    *config.Config
	logger *zerolog.Logger
	grepUC domain.GrepUsecase
}

func NewApplication(cfg *config.Config, logger *zerolog.Logger) *Application {
	grepUsecase := grepUC.NewGrepService(cfg, logger)
	return &Application{
		cfg:    cfg,
		logger: logger,
		grepUC: grepUsecase,
	}
}

func (a *Application) StartServer(ctx context.Context, addr string) {
	grpcAdapter := grpc.NewGRPCAdapter(a.grepUC, a.logger)
	grpcAdapter.StartServer(ctx, addr)
}

func (a *Application) RunClient(ctx context.Context, pattern string, inputFile string, stdin io.Reader) {
	cliAdapter := cli.NewCLIAdapter(a.grepUC, a.logger)
	cliAdapter.RunClient(ctx, a.cfg.Addrs, pattern, inputFile, stdin, a.cfg.GrepOptions)
}
