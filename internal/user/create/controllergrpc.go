package create

import (
	"context"
	"fmt"

	user "github.com/bendbennett/go-api-demo/generated"
	"github.com/bendbennett/go-api-demo/internal/log"
	"github.com/bendbennett/go-api-demo/internal/validate"
)

type grpcController struct {
	validator  validate.Validator
	interactor interactor
	presenter  presenter
	logger     log.Logger
}

type GRPCController interface {
	Create(context.Context, *user.CreateRequest) (*user.UserResponse, error)
}

func NewGRPCController(
	validator validate.Validator,
	interactor interactor,
	presenter presenter,
	logger log.Logger,
) *grpcController {
	return &grpcController{
		validator,
		interactor,
		presenter,
		logger,
	}
}

func (c *grpcController) Create(
	ctx context.Context,
	req *user.CreateRequest,
) (*user.UserResponse, error) {
	input := inputData{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	errs := c.validator.ValidateStruct(input)
	if errs != nil {
		c.logger.WithSpan(ctx).Infof("input invalid: %v", errs)
		return nil, fmt.Errorf("%v", errs)
	}

	od, err := c.interactor.create(
		ctx,
		input,
	)
	if err != nil {
		c.logger.WithSpan(ctx).Error(err)
		return nil, err
	}

	vm := c.presenter.viewModel(od)

	return &user.UserResponse{
		Id:        vm.ID,
		FirstName: vm.FirstName,
		LastName:  vm.LastName,
		CreatedAt: vm.CreatedAt,
	}, nil
}
