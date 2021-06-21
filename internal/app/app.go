package app

import (
	"context"

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
}

func New(
	httpRouter Component,
	grpcRouter Component,
) *App {
	return &App{
		httpRouter,
		grpcRouter,
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

	return eg.Wait()
}
