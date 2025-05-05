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

func newSubPubServer(log *slog.Logger, serv subpub.SubPub) *subPubServer {
	return &subPubServer{
		log:  log,
		serv: serv,
	}
}

// Subscribe defines the logic of the handling the subscribing request.
func (s *subPubServer) Subscribe(request *SubscribeRequest, stream grpc.ServerStreamingServer[Event]) error {
	const op = "spserv.Subscribe"

	msgCh := make(chan string)
	sub, err := s.serv.Subscribe(request.Key, func(msg interface{}) {
		msgCh <- msg.(string)
	})

	if err != nil {
		var code codes.Code
		var subErr = fmt.Errorf("error of the %s: %s", op, err)

		if errors.Is(err, subpub.ErrInputData) {
			code = codes.InvalidArgument
		} else if errors.Is(err, subpub.ErrSystemCondition) {
			code = codes.Unavailable
		}
		s.log.Error(subErr.Error())

		return status.Error(code, subErr.Error())
	}

	for msg := range msgCh {
		if err := stream.Send(&Event{Data: msg}); err != nil {
			// TODO: rewrite this part: think there's more efficiency solution can be.
			sub.Unsubscribe()
			close(msgCh)

			sendErr := fmt.Errorf("error of the %s: %w: %s", op, ErrSendingMsg, err)
			s.log.Error(sendErr.Error())
			return status.Error(codes.Aborted, sendErr.Error())
		}
	}

	return nil
}

// Publish defines the logic of the handling the publishing requests.
func (s *subPubServer) Publish(_ context.Context, request *PublishRequest) (*emptypb.Empty, error) {
	const op = "spserv.Publish"

	if err := s.serv.Publish(request.Key, request.Data); err != nil {
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
