package spserv

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/MaKcm14/vk-test/pkg/subpub"
	grpc "google.golang.org/grpc"
)

// SubPub is the main service that defines the grpc-server configuring.
type SubPubService struct {
	serv *grpc.Server
	log  *slog.Logger
	conn net.Listener
	kern *subPubServer
}

func NewSubPubService(log *slog.Logger, socket string, handler subpub.SubPub) (SubPubService, error) {
	const op = "spserv.NewSubPub"

	lis, err := net.Listen("tcp", socket)

	if err != nil {
		errNet := fmt.Errorf("error of the %s: %w: %s", op, ErrNetOpenConn, err)
		log.Error(errNet.Error())
		return SubPubService{}, errNet
	}

	serv := grpc.NewServer()

	subPubServ := newSubPubServer(log, handler)
	RegisterPubSubServer(serv, subPubServ)

	return SubPubService{
		log:  log,
		serv: serv,
		conn: lis,
		kern: subPubServ,
	}, nil
}

// Run starts the serving new client's requests.
func (s *SubPubService) Run() {
	s.serv.Serve(s.conn)
}

// Close releases the resources of the server.
func (s *SubPubService) Close() {
	s.conn.Close()
	s.kern.close()
}
