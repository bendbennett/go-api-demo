package search

import (
	"context"
	"errors"
	"testing"

	"github.com/bendbennett/go-api-demo/internal/sanitise"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	pb "github.com/bendbennett/go-api-demo/generated"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/stretchr/testify/assert"
)

type interactorMock struct {
}

func (m *interactorMock) search(context.Context, string) (outputData, error) {
	return outputData{}, nil
}

type interactorMockError struct {
}

func (m *interactorMockError) search(context.Context, string) (outputData, error) {
	return outputData{}, errors.New("interactor search error")
}

type presenterMock struct {
}

func (pm *presenterMock) viewModel(outputData) viewModel {
	return viewModel{
		{
			ID:        "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
			FirstName: "john",
			LastName:  "smith",
			CreatedAt: "2006-01-02T15:04:05-0700",
		},
		{
			ID:        "1a81dec3-3638-4eb4-b04a-83d744f5f3a8",
			FirstName: "joanna",
			LastName:  "smithson",
			CreatedAt: "2006-01-02T16:04:05-0700",
		},
	}
}

func TestGRPC_Search(t *testing.T) {
	cases := []struct {
		name             string
		interactor       interactor
		presenter        presenter
		request          *pb.SearchRequest
		expectedResponse *pb.UsersResponse
		expectedError    bool
	}{
		{
			"search term invalid error",
			&interactorMockError{},
			&presenterMock{},
			&pb.SearchRequest{SearchTerm: "ab"},
			nil,
			true,
		},
		{
			"interactor search error",
			&interactorMockError{},
			&presenterMock{},
			&pb.SearchRequest{SearchTerm: "abc"},
			nil,
			true,
		},
		{
			"success",
			&interactorMock{},
			&presenterMock{},
			&pb.SearchRequest{SearchTerm: "abc"},
			&pb.UsersResponse{
				Users: []*pb.UserResponse{
					{
						Id:        "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
						FirstName: "john",
						LastName:  "smith",
						CreatedAt: "2006-01-02T15:04:05-0700",
					},
					{
						Id:        "1a81dec3-3638-4eb4-b04a-83d744f5f3a8",
						FirstName: "joanna",
						LastName:  "smithson",
						CreatedAt: "2006-01-02T16:04:05-0700",
					},
				},
			},
			false,
		},
	}

	zc, _ := observer.New(zapcore.DebugLevel)
	zl := zap.New(zc)
	logger := log.NewLogger(zl)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			controller := NewGRPCController(
				sanitise.AlphaWithHyphen,
				c.interactor,
				c.presenter,
				logger,
			)

			resp, err := controller.Search(context.Background(), c.request)

			assert.Equal(t, c.expectedResponse, resp)

			if c.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
