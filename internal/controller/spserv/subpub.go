package spserv

import (
	"fmt"
	"log/slog"
	"net"

	grpc "google.golang.org/grpc"
)

// SubPub is the main service that defines the grpc-server configuring.
type SubPub struct {
	serv *grpc.Server
	log  *slog.Logger
	conn net.Listener
	kern *subPubServer
}

func NewSubPub(log *slog.Logger, socket string) (SubPub, error) {
	const op = "spserv.NewSubPub"

	lis, err := net.Listen("tcp", socket)

	if err != nil {
		errNet := fmt.Errorf("error of the %s: %w: %s", op, ErrNetOpenConn, err)
		log.Error(errNet.Error())
		return SubPub{}, errNet
	}

	serv := grpc.NewServer()

	subPubServ := newSubPubServer(log)
	RegisterPubSubServer(serv, subPubServ)

	return SubPub{
		log:  log,
		serv: serv,
		conn: lis,
		kern: subPubServ,
	}, nil
}

// Run starts the serving new client's requests.
func (s *SubPub) Run() {
	s.serv.Serve(s.conn)
}

// Close releases the resources of the server.
func (s *SubPub) Close() {
	s.conn.Close()
	s.kern.close()
}
