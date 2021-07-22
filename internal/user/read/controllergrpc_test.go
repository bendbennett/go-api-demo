package read

import (
	"context"
	"errors"
	"io"
	"testing"

	pb "github.com/bendbennett/go-api-demo/generated"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type interactorMock struct {
}

func (m *interactorMock) read(context.Context) (outputData, error) {
	return outputData{}, nil
}

type interactorMockError struct {
}

func (m *interactorMockError) read(context.Context) (outputData, error) {
	return outputData{}, errors.New("interactor read error")
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

func TestGRPC_Read(t *testing.T) {
	cases := []struct {
		name             string
		interactor       interactor
		presenter        presenter
		request          *pb.ReadRequest
		expectedResponse *pb.UsersResponse
		expectedError    bool
	}{
		{
			"interactor read error",
			&interactorMockError{},
			&presenterMock{},
			&pb.ReadRequest{},
			nil,
			true,
		},
		{
			"success",
			&interactorMock{},
			&presenterMock{},
			&pb.ReadRequest{},
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

	l := log.New()
	l.SetOutput(io.Discard)
	logger := l.WithFields(nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			controller := NewGRPCController(
				c.interactor,
				c.presenter,
				logger,
			)

			resp, err := controller.Read(context.Background(), c.request)

			assert.Equal(t, c.expectedResponse, resp)

			if c.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
