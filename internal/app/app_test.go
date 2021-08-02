package app

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type okComponent struct{}

func (t *okComponent) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

type errorComponent struct{}

func (t *errorComponent) Run(ctx context.Context) error {
	return errors.New("error from component")
}

func TestRun_NoError(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	a := New(&okComponent{}, &okComponent{}, []io.Closer{})

	go func() {
		<-time.After(time.Millisecond)
		cancelFunc()
	}()

	err := a.Run(ctx)
	require.NoError(t, err)
}

func TestRun_Error(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	a := New(&errorComponent{}, &okComponent{}, []io.Closer{})

	err := a.Run(ctx)
	require.Error(t, err)
}

func TestApp_CommitHash(t *testing.T) {
	commitHash = "123abc"
	assert.Equal(t, CommitHash(), "123abc")
}
