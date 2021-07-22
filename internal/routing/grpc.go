package routing

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc/reflection"

	user "github.com/bendbennett/go-api-demo/generated"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GRPCRouter struct {
	userServer *userServer
	logger     *log.Entry
	port       int
}

type GRPCControllers struct {
	UserCreate func(ctx context.Context, in *user.CreateRequest) (*user.UserResponse, error)
	UserRead   func(ctx context.Context, in *user.ReadRequest) (*user.UsersResponse, error)
}

// NewGRPCRouter returns a pointer to a GRPCRouter struct
// populated with the port for the server and a logger.
func NewGRPCRouter(
	controllers GRPCControllers,
	logger *log.Entry,
	port int,
) *GRPCRouter {
	return &GRPCRouter{
		&userServer{
			UnimplementedUserServer: user.UnimplementedUserServer{},
			UserCreate:              controllers.UserCreate,
			UserRead:                controllers.UserRead,
		},
		logger,
		port,
	}
}

type UserCreate func(ctx context.Context, in *user.CreateRequest) (*user.UserResponse, error)
type UserRead func(ctx context.Context, in *user.ReadRequest) (*user.UsersResponse, error)

type userServer struct {
	user.UnimplementedUserServer
	UserCreate
	UserRead
}

func (us *userServer) Create(
	ctx context.Context,
	createReq *user.CreateRequest,
) (*user.UserResponse, error) {
	return us.UserCreate(ctx, createReq)
}

func (us *userServer) Read(
	ctx context.Context,
	readReq *user.ReadRequest,
) (*user.UsersResponse, error) {
	return us.UserRead(ctx, readReq)
}

// Run configures and starts a gRPC server. A go routine is
// used to listen for context cancellation and triggers
// a call to server stop.
func (r *GRPCRouter) Run(ctx context.Context) error {
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf(
			":%d",
			r.port,
		),
	)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	user.RegisterUserServer(
		s,
		r.userServer,
	)

	reflection.Register(s)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	r.logger.Infof(
		"gRPC server running on port %v",
		r.port,
	)

	return s.Serve(listener)
}
