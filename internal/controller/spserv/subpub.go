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
	conn     net.Listener

	kern SPServer
	log  *slog.Logger
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
	const op = "spserv.Close"
	s.log.Info("releasing the resources of the service")

	s.kern.Close()
	s.log.Info("service condition was changed on 'closed'")

	s.grpcServ.GracefulStop()
	s.log.Info("service was gracefuly stopped")

	if err := s.conn.Close(); err != nil {
		s.log.Error(fmt.Sprintf("error of the %s: %s", op, err))
		return
	}
	s.log.Info("socket was closed")
}
