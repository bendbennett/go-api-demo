package create

import (
	"context"
	"io"
	"testing"

	pb "github.com/bendbennett/go-api-demo/generated"
	"github.com/stretchr/testify/assert"

	"github.com/bendbennett/go-api-demo/internal/validate"
	log "github.com/sirupsen/logrus"
)

func TestGRPC_Create(t *testing.T) {
	cases := []struct {
		name             string
		validator        validate.Validator
		interactor       interactor
		presenter        presenter
		request          *pb.CreateRequest
		expectedResponse *pb.UserResponse
		expectedError    bool
	}{
		{
			"input invalid",
			&validatorMockInputInvalid{},
			&interactorMock{},
			&presenterMock{},
			&pb.CreateRequest{
				FirstName: "ab",
			},
			nil,
			true,
		},
		{
			"interactor create error",
			&validatorMock{},
			&interactorMockError{},
			&presenterMock{},
			&pb.CreateRequest{
				FirstName: "john",
				LastName:  "smith",
			},
			nil,
			true,
		},
		{
			"success",
			&validatorMock{},
			&interactorMock{},
			&presenterMock{},
			&pb.CreateRequest{
				FirstName: "john",
				LastName:  "smith",
			},
			&pb.UserResponse{
				Id:        "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
				FirstName: "john",
				LastName:  "smith",
				CreatedAt: "2006-01-02T15:04:05-0700",
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
				c.validator,
				c.interactor,
				c.presenter,
				logger,
			)

			resp, err := controller.Create(context.Background(), c.request)

			assert.Equal(t, c.expectedResponse, resp)

			if c.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
