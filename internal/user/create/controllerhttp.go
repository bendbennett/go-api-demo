package create

import (
	"encoding/json"
	"net/http"

	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/response"
	"github.com/bendbennett/go-api-demo/internal/validate"
)

type httpController struct {
	validator  validate.Validator
	interactor interactor
	presenter  presenter
	logger     log.Logger
}

type HTTPController interface {
	Create(w http.ResponseWriter, r *http.Request)
}

func NewHTTPController(
	validator validate.Validator,
	interactor interactor,
	presenter presenter,
	logger log.Logger,
) *httpController {
	return &httpController{
		validator,
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
	input := inputData{}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		c.logger.ErrorfContext(ctx, "json body invalid: %v", err)
		response.WriteErrorResponse(
			w,
			http.StatusBadRequest,
			"failed validation",
			map[string]string{"body": "json invalid"},
		)
		return
	}

	errs := c.validator.ValidateStruct(input)
	if errs != nil {
		c.logger.InfofContext(ctx, "input invalid: %v", errs)
		response.WriteErrorResponse(
			w,
			http.StatusBadRequest,
			"failed validation",
			errs,
		)
		return
	}

	od, err := c.interactor.create(
		ctx,
		input,
	)
	if err != nil {
		c.logger.ErrorContext(ctx, err)
		response.Write500Response(w)
		return
	}

	vm := c.presenter.viewModel(od)

	type output struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		CreatedAt string `json:"created_at"`
	}

	response.WriteResponse(
		w,
		http.StatusCreated,
		output(vm),
	)
}
