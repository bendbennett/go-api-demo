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
	components []Component
	closers    []io.Closer
}

func New(
	components []Component,
	closers []io.Closer,
) *App {
	return &App{
		components,
		closers,
	}
}

// Run uses an errgroup.WithContext for synchronisation,
// error propagation, and Context cancellation,
// facilitating graceful shutdown of HTTP and gRPC servers
// on interrupt or terminate.
func (a *App) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, c := range a.components {
		f := func(c Component) func() error {
			return func() error {
				return c.Run(ctx)
			}
		}

		eg.Go(f(c))
	}

	return eg.Wait()
}

func (a *App) Close() {
	for _, closer := range a.closers {
		_ = closer.Close()
	}
}
