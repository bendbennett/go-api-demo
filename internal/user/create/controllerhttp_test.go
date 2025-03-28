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

type loggerMock struct {
}

func (lm loggerMock) Panic(error)                                           {}
func (lm loggerMock) Panicf(string, ...interface{})                         {}
func (lm loggerMock) Error(error)                                           {}
func (lm loggerMock) ErrorContext(context.Context, error)                   {}
func (lm loggerMock) Errorf(string, ...interface{})                         {}
func (lm loggerMock) ErrorfContext(context.Context, string, ...interface{}) {}
func (lm loggerMock) Infof(string, ...interface{})                          {}
func (lm loggerMock) InfofContext(context.Context, string, ...interface{})  {}

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

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/user", c.body)
			w := httptest.NewRecorder()

			controller := NewHTTPController(
				c.validator,
				c.interactor,
				c.presenter,
				loggerMock{},
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
