package routing

import (
	"context"
	"fmt"
	"net"

	pb "github.com/bendbennett/go-api-demo/generated"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GRPCRouter struct {
	userServer *userServer
	logger     *log.Entry
	port       int
}

type GRPCControllers struct {
	UserCreate func(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error)
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
			UnimplementedUserServer: pb.UnimplementedUserServer{},
			CreateUser:              controllers.UserCreate,
		},
		logger,
		port,
	}
}

type CreateUser func(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error)

type userServer struct {
	pb.UnimplementedUserServer
	CreateUser
}

func (us *userServer) Create(
	ctx context.Context,
	createReq *pb.CreateRequest,
) (*pb.CreateResponse, error) {
	return us.CreateUser(ctx, createReq)
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
	pb.RegisterUserServer(
		s,
		r.userServer,
	)

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
