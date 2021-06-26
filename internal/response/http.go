package response

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Errors  map[string]string `json:"errors,omitempty"`
	Message string            `json:"message"`
}

func WriteResponse(
	w http.ResponseWriter,
	statusCode int,
	response interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if response == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(response)
}

func WriteErrorResponse(
	w http.ResponseWriter,
	statusCode int,
	msg string,
	errs map[string]string,
) {
	WriteResponse(
		w,
		statusCode,
		errorResponse{
			errs,
			msg,
		},
	)
}

func Write500Response(w http.ResponseWriter) {
	WriteResponse(
		w,
		http.StatusInternalServerError,
		errorResponse{
			Message: "internal server error",
		},
	)
}
