package create

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/validate"
	"github.com/stretchr/testify/assert"
)

type validatorMock struct {
}

func (m *validatorMock) ValidateStruct(input interface{}) map[string]string {
	return nil
}

type validatorMockInputInvalid struct {
}

func (m *validatorMockInputInvalid) ValidateStruct(input interface{}) map[string]string {
	return map[string]string{"input": "invalid"}
}

type interactorMock struct {
}

func (m *interactorMock) create(context.Context, inputData) (outputData, error) {
	return outputData{}, nil
}

type interactorMockError struct {
}

func (m *interactorMockError) create(context.Context, inputData) (outputData, error) {
	return outputData{}, errors.New("interactor create error")
}

type presenterMock struct {
}

func (pm *presenterMock) viewModel(outputData) viewModel {
	return viewModel{
		ID:        "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
		FirstName: "john",
		LastName:  "smith",
		CreatedAt: "2006-01-02T15:04:05-0700",
	}
}

func TestRest_Create(t *testing.T) {
	cases := []struct {
		name                 string
		validator            validate.Validator
		interactor           interactor
		presenter            presenter
		body                 io.Reader
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			"json unmarshall error",
			&validatorMock{},
			&interactorMock{},
			&presenterMock{},
			strings.NewReader(`{"first_name:}`),
			http.StatusBadRequest,
			`{
  									"message": "failed validation",
									"errors": {
										"body": "json invalid"
									}
								}`,
		},
		{
			"input invalid",
			&validatorMockInputInvalid{},
			&interactorMock{},
			&presenterMock{},
			strings.NewReader(`{"first_name": "ab"}`),
			http.StatusBadRequest,
			`{
  									"message": "failed validation",
									"errors": {
										"input": "invalid"
									}
								}`,
		},
		{
			"interactor create error",
			&validatorMock{},
			&interactorMockError{},
			&presenterMock{},
			strings.NewReader(`{"first_name": "john", "last_name": "smith"}`),
			http.StatusInternalServerError,
			`{
  									"message": "internal server error"
								}`,
		},
		{
			"success",
			&validatorMock{},
			&interactorMock{},
			&presenterMock{},
			strings.NewReader(`{"first_name": "john", "last_name": "smith"}`),
			http.StatusCreated,
			`{
									"id": "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
									"first_name": "john",
									"last_name": "smith",
									"created_at": "2006-01-02T15:04:05-0700"
								}`,
		},
	}

	zc, _ := observer.New(zapcore.DebugLevel)
	zl := zap.New(zc)
	logger := log.NewLogger(zl)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/user", c.body)
			w := httptest.NewRecorder()

			controller := NewHTTPController(
				c.validator,
				c.interactor,
				c.presenter,
				logger,
			)

			controller.Create(w, r)

			// Flatten JSON formatted response body.
			expectedResponseBody := bytes.NewBuffer(nil)
			_ = json.Compact(expectedResponseBody, []byte(c.expectedResponseBody))

			assert.Equal(t, c.expectedStatus, w.Code)
			assert.JSONEq(t, expectedResponseBody.String(), w.Body.String())
		})
	}
}
