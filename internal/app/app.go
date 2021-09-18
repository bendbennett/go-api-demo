package app

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"
)

// commitHash is populated with the git commit hash
// through linker flags (ldflags) when the binary is
// compiled (see Makefile).
var commitHash string

func CommitHash() string {
	return commitHash
}

type Component interface {
	Run(ctx context.Context) error
}

type App struct {
	httpRouter Component
	grpcRouter Component
	consumer   Component
	closers    []io.Closer
}

func New(
	httpRouter Component,
	grpcRouter Component,
	consumer Component,
	closers []io.Closer,
) *App {
	return &App{
		httpRouter,
		grpcRouter,
		consumer,
		closers,
	}
}

// Run uses an errgroup.WithContext for synchronisation,
// error propagation, and Context cancellation,
// facilitating graceful shutdown of HTTP and gRPC servers
// on interrupt or terminate.
func (a *App) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return a.httpRouter.Run(ctx)
	})

	eg.Go(func() error {
		return a.grpcRouter.Run(ctx)
	})

	eg.Go(func() error {
		return a.consumer.Run(ctx)
	})

	return eg.Wait()
}

func (a *App) Close() {
	for _, closer := range a.closers {
		_ = closer.Close()
	}
}
