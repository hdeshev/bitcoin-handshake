package internal

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/pkg/errors"

	"deshev.com/bitcoin-handshake/btc/client"
	"deshev.com/bitcoin-handshake/btc/encoding"
	"deshev.com/bitcoin-handshake/config"
)

type RemoteClient interface {
	Connect() (<-chan encoding.Message, error)
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
	messageC, err := a.client.Connect()
	if err != nil {
		return errors.Wrap(err, "client connect error")
	}
	for {
		select {
		case <-a.ctx.Done():
			return context.Canceled
		case msg, ok := <-messageC:
			if !ok {
				return errors.New("client connection closed")
			}
			a.log.Info("app received message", "command", msg.GetCommand())
		}
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
