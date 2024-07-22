package search

import (
	"fmt"
	"net/http"

	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/response"
	"github.com/gorilla/mux"
)

type httpController struct {
	sanitise   alphaWithHyphen
	interactor interactor
	presenter  presenter
	logger     log.Logger
}

type HTTPController interface {
	Search(w http.ResponseWriter, r *http.Request)
}

func NewHTTPController(
	alphaWithHyphen alphaWithHyphen,
	interactor interactor,
	presenter presenter,
	logger log.Logger,
) *httpController {
	return &httpController{
		sanitise:   alphaWithHyphen,
		interactor: interactor,
		presenter:  presenter,
		logger:     logger,
	}
}

func (c *httpController) Search(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	params := mux.Vars(r)

	searchTerm, err := c.sanitise(params["searchTerm"])
	if err != nil {
		c.logger.ErrorfContext(ctx, "clean string failed: %v", err)
		response.Write500Response(
			w,
		)
		return
	}

	if len(searchTerm) < searchTermMinLen {
		msg := fmt.Sprintf("search term must be >= %v chars", searchTermMinLen)
		c.logger.InfofContext(ctx, msg)
		response.WriteErrorResponse(
			w,
			http.StatusBadRequest,
			"failed validation",
			map[string]string{"invalid": msg},
		)
		return
	}

	od, err := c.interactor.
		search(
			ctx,
			searchTerm,
		)
	if err != nil {
		c.logger.ErrorContext(ctx, err)
		response.Write500Response(w)
		return
	}

	type user struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		CreatedAt string `json:"created_at"`
	}

	var (
		vm    = c.presenter.viewModel(od)
		users = []user{}
	)

	for _, u := range vm {
		users = append(
			users,
			user(u),
		)
	}

	response.WriteResponse(
		w,
		http.StatusOK,
		users,
	)

}
