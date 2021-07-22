package read

import (
	"net/http"

	"github.com/bendbennett/go-api-demo/internal/response"
	log "github.com/sirupsen/logrus"
)

type httpController struct {
	interactor interactor
	presenter  presenter
	logger     *log.Entry
}

type HTTPController interface {
	Read(w http.ResponseWriter, r *http.Request)
}

func NewHTTPController(
	interactor interactor,
	presenter presenter,
	logger *log.Entry,
) *httpController {
	return &httpController{
		interactor,
		presenter,
		logger,
	}
}

func (c *httpController) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	od, err := c.interactor.read(
		ctx,
	)
	if err != nil {
		c.logger.Warn(err)
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
