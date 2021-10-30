package search

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bendbennett/go-api-demo/internal/user"
)

type searcherMockError struct {
}

func (m *searcherMockError) Search(context.Context, string) ([]user.User, error) {
	return []user.User{}, errors.New("searcher search error")
}

type searcherMock struct {
}

func (m *searcherMock) Search(context.Context, string) ([]user.User, error) {
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

func TestInteractor_Search(t *testing.T) {
	cases := []struct {
		name               string
		reader             searcher
		expectedOutputData outputData
		returnsErr         bool
	}{
		{
			"searcher returns error",
			&searcherMockError{},
			outputData{},
			true,
		},
		{
			"success",
			&searcherMock{},
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
			od, err := interactor.search(
				context.Background(),
				"",
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
