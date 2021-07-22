package read

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestRest_Create(t *testing.T) {
	cases := []struct {
		name                 string
		interactor           interactor
		presenter            presenter
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			"interactor read error",
			&interactorMockError{},
			&presenterMock{},
			http.StatusInternalServerError,
			`{
  									"message": "internal server error"
								}`,
		},
		{
			"success",
			&interactorMock{},
			&presenterMock{},
			http.StatusOK,
			`[
									{
										"id": "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
										"first_name": "john",
										"last_name": "smith",
										"created_at": "2006-01-02T15:04:05-0700"
									},
																	{
										"id": "1a81dec3-3638-4eb4-b04a-83d744f5f3a8",
										"first_name": "joanna",
										"last_name": "smithson",
										"created_at": "2006-01-02T16:04:05-0700"
									}
								]`,
		},
	}

	l := log.New()
	l.SetOutput(io.Discard)
	logger := l.WithFields(nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/user", nil)
			w := httptest.NewRecorder()

			controller := NewHTTPController(
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
