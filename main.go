package main

import (
	"context"
	"errors"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"deshev.com/bitcoin-handshake/internal"
)

func main() {
	log := slog.Default()
	ops, ctx := errgroup.WithContext(context.Background())

	app := internal.NewApplication(ctx, log)
	log.Info("starting bitcoin-handshake")

	ops.Go(app.StartSignalMonitor)

	err := ops.Wait()
	if !errors.Is(err, context.Canceled) {
		log.Error("server terminated abnormally", "error", err)
	}
}
