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
	logger *log.Entry
	port   int
}

// NewGRPCRouter returns a pointer to a GRPCRouter struct
// populated with the port for the server and a logger.
func NewGRPCRouter(
	logger *log.Entry,
	port int,
) *GRPCRouter {
	return &GRPCRouter{
		logger,
		port,
	}
}

type server struct {
	pb.UnimplementedHelloServer
}

// Hello is a an implementation of the func required for the
// server struct to satisfy the hello.HelloServer interface.
func (s *server) Hello(
	ctx context.Context,
	in *pb.HelloRequest,
) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Message: "Hello " + in.GetName(),
	}, nil
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
	pb.RegisterHelloServer(
		s,
		&server{},
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
