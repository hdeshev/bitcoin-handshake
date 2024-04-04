package internal

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"deshev.com/bitcoin-handshake/btc/client"
	"deshev.com/bitcoin-handshake/config"
)

type RemoteClient interface {
	Connect() error
}

type Application struct {
	log    *slog.Logger
	config *config.Config
	ctx    context.Context
	client RemoteClient
}

func NewApplication(ctx context.Context, log *slog.Logger) *Application {
	cfg := config.New()

	return &Application{
		ctx:    ctx,
		log:    log,
		config: cfg,
		client: client.New(ctx, log, cfg),
	}
}

func (a *Application) StartConnection() error {
	return a.client.Connect() //nolint:wrapcheck // boot errors are logged in main
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
