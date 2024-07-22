package routing

import (
	"context"
	"fmt"
	"net"

	user "github.com/bendbennett/go-api-demo/generated"
	"github.com/bendbennett/go-api-demo/internal/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCRouter struct {
	userServer       *userServer
	logger           log.Logger
	telemetryEnabled bool
	port             int
}

type GRPCControllers struct {
	UserCreate func(ctx context.Context, in *user.CreateRequest) (*user.UserResponse, error)
	UserRead   func(ctx context.Context, in *user.ReadRequest) (*user.UsersResponse, error)
	UserSearch func(ctx context.Context, in *user.SearchRequest) (*user.UsersResponse, error)
}

// NewGRPCRouter returns a pointer to a GRPCRouter struct
// populated with the port for the server and a logger.
func NewGRPCRouter(
	controllers GRPCControllers,
	logger log.Logger,
	telemetryEnabled bool,
	port int,
) *GRPCRouter {
	return &GRPCRouter{
		&userServer{
			UnimplementedUserServer: user.UnimplementedUserServer{},
			UserCreate:              controllers.UserCreate,
			UserRead:                controllers.UserRead,
			UserSearch:              controllers.UserSearch,
		},
		logger,
		telemetryEnabled,
		port,
	}
}

type UserCreate func(ctx context.Context, in *user.CreateRequest) (*user.UserResponse, error)
type UserRead func(ctx context.Context, in *user.ReadRequest) (*user.UsersResponse, error)
type UserSearch func(ctx context.Context, in *user.SearchRequest) (*user.UsersResponse, error)

type userServer struct {
	user.UnimplementedUserServer
	UserCreate
	UserRead
	UserSearch
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

func (us *userServer) Search(
	ctx context.Context,
	searchReq *user.SearchRequest,
) (*user.UsersResponse, error) {
	sr := searchReq

	return us.UserSearch(ctx, sr)
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
		r.logger.Panicf("failed to listen: %v", err)
	}

	serverOptions := []grpc.ServerOption{}

	if r.telemetryEnabled {
		serverOptions = append(serverOptions, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	s := grpc.NewServer(serverOptions...)
	user.RegisterUserServer(
		s,
		r.userServer,
	)

	// TODO: This should be configurable as it publicly exposes the gRPC endpoints.
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
