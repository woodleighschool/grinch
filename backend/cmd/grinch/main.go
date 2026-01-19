package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/woodleighschool/grinch/internal/app"
	"github.com/woodleighschool/grinch/internal/platform/logging"
)

func main() {
	log := logging.NewLogger("info")

	if err := run(); err != nil {
		log.Error("grinch failed", "error", err)
		os.Exit(1)
	}
}

// run builds the app and runs it until the context is canceled.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		return err
	}

	return a.Run(ctx)
}
