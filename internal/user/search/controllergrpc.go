package search

import (
	"context"
	"errors"
	"fmt"

	user "github.com/bendbennett/go-api-demo/generated"
	"github.com/bendbennett/go-api-demo/internal/log"
)

type grpcController struct {
	sanitise   alphaWithHyphen
	interactor interactor
	presenter  presenter
	logger     log.Logger
}

type GRPCController interface {
	Search(context.Context, *user.SearchRequest) (*user.UsersResponse, error)
}

func NewGRPCController(
	sanitise alphaWithHyphen,
	interactor interactor,
	presenter presenter,
	logger log.Logger,
) *grpcController {
	return &grpcController{
		sanitise,
		interactor,
		presenter,
		logger,
	}
}

func (c *grpcController) Search(
	ctx context.Context,
	searchReq *user.SearchRequest,
) (*user.UsersResponse, error) {
	searchTerm, err := c.sanitise(searchReq.SearchTerm)
	if err != nil {
		c.logger.ErrorfContext(ctx, "clean string failed: %v", err)
		return nil, err
	}

	if len(searchTerm) < searchTermMinLen {
		msg := fmt.Sprintf("search term must be >= %v chars", searchTermMinLen)
		c.logger.InfofContext(ctx, msg)
		return nil, errors.New(msg)
	}

	od, err := c.interactor.search(
		ctx,
		searchTerm,
	)
	if err != nil {
		c.logger.ErrorContext(ctx, err)
		return nil, err
	}

	vm := c.presenter.viewModel(od)

	var users []*user.UserResponse

	for _, u := range vm {
		users = append(
			users,
			&user.UserResponse{
				Id:        u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				CreatedAt: u.CreatedAt,
			},
		)
	}

	return &user.UsersResponse{
		Users: users,
	}, nil
}
