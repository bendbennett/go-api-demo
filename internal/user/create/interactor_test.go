package create

import (
	"context"
	"errors"
	"testing"

	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/bendbennett/go-api-demo/internal/validate"
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

func TestInteractor_Create(t *testing.T) {
	cases := []struct {
		name               string
		creator            user.Creator
		expectedOutputData outputData
		returnsErr         bool
	}{
		{
			"creator returns error",
			&creatorMockError{},
			outputData{},
			true,
		},
		{
			"success",
			&creatorMock{},
			outputData{
				FirstName: "john",
				LastName:  "smith",
			},
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			interactor := NewInteractor(
				c.creator,
			)
			od, err := interactor.create(
				context.Background(),
				inputData{
					FirstName: "john",
					LastName:  "smith",
				},
			)

			if c.returnsErr {
				assert.Error(t, err)
				assert.Equal(t, c.expectedOutputData.ID, od.ID)
				assert.Equal(t, c.expectedOutputData.CreatedAt, od.CreatedAt)
			} else {
				assert.NoError(t, err)
				assert.True(t, validate.IsUUID(od.ID))
				assert.True(t, !od.CreatedAt.IsZero())
			}

			assert.Equal(t, c.expectedOutputData.FirstName, od.FirstName)
			assert.Equal(t, c.expectedOutputData.LastName, od.LastName)
		})
	}
}
