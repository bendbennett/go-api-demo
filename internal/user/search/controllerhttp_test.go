package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bendbennett/go-api-demo/internal/sanitise"
	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
)

func TestRest_Search(t *testing.T) {
	cases := []struct {
		name                 string
		searchTerm           string
		interactor           interactor
		presenter            presenter
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			"search term invalid error",
			"ab",
			&interactorMockError{},
			&presenterMock{},
			http.StatusBadRequest,
			`{
  									"message": "failed validation",
									"errors": {
										"invalid": "search term must be >= 3 chars"
									}
								}`,
		},
		{
			"interactor search error",
			"abc",
			&interactorMockError{},
			&presenterMock{},
			http.StatusInternalServerError,
			`{
  									"message": "internal server error"
								}`,
		},
		{
			"success",
			"abc",
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

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/search/%s", c.searchTerm), nil)
			r = mux.SetURLVars(r, map[string]string{"searchTerm": c.searchTerm})

			w := httptest.NewRecorder()

			controller := NewHTTPController(
				sanitise.AlphaWithHyphen,
				c.interactor,
				c.presenter,
				loggerMock{},
			)

			controller.Search(w, r)

			// Flatten JSON formatted response body.
			expectedResponseBody := bytes.NewBuffer(nil)
			_ = json.Compact(expectedResponseBody, []byte(c.expectedResponseBody))

			assert.Equal(t, c.expectedStatus, w.Code)
			assert.JSONEq(t, expectedResponseBody.String(), w.Body.String())
		})
	}
}
