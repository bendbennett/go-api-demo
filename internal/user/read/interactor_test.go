package read

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bendbennett/go-api-demo/internal/user"
)

type readerMockError struct {
}

func (m *readerMockError) Read(context.Context) ([]user.User, error) {
	return []user.User{}, errors.New("reader read error")
}

type readerMock struct {
}

func (m *readerMock) Read(context.Context) ([]user.User, error) {
	return []user.User{
		{
			ID:        "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
			FirstName: "john",
			LastName:  "smith",
			CreatedAt: createdAt(),
		},
	}, nil
}

func createdAt() time.Time {
	createdAt, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05-0700")

	return createdAt
}

func TestInteractor_Create(t *testing.T) {
	cases := []struct {
		name               string
		reader             user.Reader
		expectedOutputData outputData
		returnsErr         bool
	}{
		{
			"reader returns error",
			&readerMockError{},
			outputData{},
			true,
		},
		{
			"success",
			&readerMock{},
			outputData{
				{
					ID:        "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
					FirstName: "john",
					LastName:  "smith",
					CreatedAt: createdAt(),
				},
			},
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			interactor := NewInteractor(
				c.reader,
			)
			od, err := interactor.read(
				context.Background(),
			)

			if c.returnsErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, c.expectedOutputData, od)
		})
	}
}
