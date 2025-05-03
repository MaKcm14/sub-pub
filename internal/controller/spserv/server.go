package spserv

import (
	context "context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/MaKcm14/vk-test/pkg/subpub"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// subPubServer defines the logic of handling the grpc's requests.
type subPubServer struct {
	UnimplementedPubSubServer
	serv subpub.SubPub
	log  *slog.Logger
}

func newSubPubServer(log *slog.Logger) *subPubServer {
	return &subPubServer{
		log:  log,
		serv: subpub.NewSubPub(),
	}
}

// Subscribe defines the logic of the handling the subscribing request.
func (s *subPubServer) Subscribe(*SubscribeRequest, grpc.ServerStreamingServer[Event]) error {
	const op = "spserv.Subscribe"
	return nil
}

// Publish defines the logic of the handling the publishing requests.
func (s *subPubServer) Publish(_ context.Context, requst *PublishRequest) (*emptypb.Empty, error) {
	const op = "spserv.Publish"

	if err := s.serv.Publish(requst.Key, requst.Data); err != nil {
		var code codes.Code
		var pubErr = fmt.Errorf("error of the %s: %s", op, err)

		if errors.Is(err, subpub.ErrInputData) {
			code = codes.InvalidArgument
		} else if errors.Is(err, subpub.ErrSystemCondition) {
			code = codes.Unavailable
		}
		s.log.Error(pubErr.Error())

		return nil, status.Error(code, pubErr.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *subPubServer) close() {
	s.serv.Close(context.Background())
}
