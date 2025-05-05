package spserv

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/MaKcm14/vk-test/internal/controller/spserv/sprpc"
	"google.golang.org/grpc"
)

// SubPub is the main service that defines the grpc-server configuring.
type SubPubService struct {
	grpcServ *grpc.Server
	log      *slog.Logger
	conn     net.Listener
	kern     SPServer
}

func NewSubPubService(log *slog.Logger, socket string, server SPServer) (SubPubService, error) {
	const op = "spserv.NewSubPub"

	log.Info(fmt.Sprintf("opening the connection on the %s", socket))
	lis, err := net.Listen("tcp", socket)

	if err != nil {
		errNet := fmt.Errorf("error of the %s: %w: %s", op, ErrNetOpenConn, err)
		log.Error(errNet.Error())
		return SubPubService{}, errNet
	}

	grpcServ := grpc.NewServer()

	sprpc.RegisterPubSubServer(grpcServ, server)

	return SubPubService{
		log:      log,
		grpcServ: grpcServ,
		conn:     lis,
		kern:     server,
	}, nil
}

// Run starts the serving new client's requests.
func (s *SubPubService) Run() {
	s.log.Info("starting the grpc-server")
	s.grpcServ.Serve(s.conn)
}

// Close releases the resources of the server.
func (s *SubPubService) Close() {
	s.log.Info("releasing resources of the service")
	s.grpcServ.GracefulStop()
	s.conn.Close()
	s.kern.Close()
}
