package read

import (
	"context"

	user "github.com/bendbennett/go-api-demo/generated"
	log "github.com/sirupsen/logrus"
)

type grpcController struct {
	interactor interactor
	presenter  presenter
	logger     *log.Entry
}

type GRPCController interface {
	Read(context.Context, *user.ReadRequest) (*user.UsersResponse, error)
}

func NewGRPCController(
	interactor interactor,
	presenter presenter,
	logger *log.Entry,
) *grpcController {
	return &grpcController{
		interactor,
		presenter,
		logger,
	}
}

func (c *grpcController) Read(
	ctx context.Context,
	readReq *user.ReadRequest,
) (*user.UsersResponse, error) {
	od, err := c.interactor.read(
		ctx,
	)
	if err != nil {
		c.logger.Warn(err)
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
