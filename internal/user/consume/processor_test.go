package consume

import (
	"context"
	"errors"
	"testing"

	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/stretchr/testify/assert"
)

type creatorMockError struct {
}

func (m *creatorMockError) Create(context.Context, ...user.User) error {
	return errors.New("creator create error")
}

type creatorMock struct {
}

func (m *creatorMock) Create(context.Context, ...user.User) error {
	return nil
}

func TestProcessor_Process(t *testing.T) {
	cases := []struct {
		name       string
		creator    user.Creator
		inputData  inputData
		returnsErr bool
	}{
		{
			"creator returns error",
			&creatorMockError{},
			inputData{},
			true,
		},
		{
			"unimplemented processing",
			&creatorMock{},
			inputData{
				before: user.User{ID: "id"},
			},
			true,
		},
		{
			"success",
			&creatorMock{},
			inputData{},
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			processor := NewProcessor(
				c.creator,
			)
			err := processor.Process(
				context.Background(),
				c.inputData,
			)

			if c.returnsErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
