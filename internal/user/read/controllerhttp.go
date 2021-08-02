package read

import (
	"net/http"

	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/response"
)

type httpController struct {
	interactor interactor
	presenter  presenter
	logger     log.Logger
}

type HTTPController interface {
	Read(w http.ResponseWriter, r *http.Request)
}

func NewHTTPController(
	interactor interactor,
	presenter presenter,
	logger log.Logger,
) *httpController {
	return &httpController{
		interactor,
		presenter,
		logger,
	}
}

func (c *httpController) Read(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	od, err := c.interactor.read(
		ctx,
	)
	if err != nil {
		c.logger.WithSpan(ctx).Error(err)
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
