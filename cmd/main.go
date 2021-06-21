package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bendbennett/go-api-demo/internal/bootstrap"
)

// main bootstraps and runs the application.
// signalShutdownHandler is run in a go routine and cancels
// the context when an interrupt or termination signal is received.
func main() {
	app := bootstrap.New()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go signalShutdownHandler(cancelFunc)

	err := app.Run(ctx)
	if err != nil {
		log.Fatalf("app run failed: %v", err)
	}

	os.Exit(0)
}

func signalShutdownHandler(cancelFunc context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	cancelFunc()
}
