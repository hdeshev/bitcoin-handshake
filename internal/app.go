package internal

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"deshev.com/bitcoin-handshake/config"
)

type Application struct {
	log    *slog.Logger
	config *config.Config
	ctx    context.Context
}

func NewApplication(ctx context.Context, log *slog.Logger) *Application {
	cfg := config.New()

	return &Application{
		ctx:    ctx,
		log:    log,
		config: cfg,
	}
}

func (a *Application) StartSignalMonitor() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	select {
	case <-a.ctx.Done():
		return context.Canceled
	case <-quit:
		return context.Canceled
	}
}
