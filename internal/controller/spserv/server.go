package spserv

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/MaKcm14/vk-test/internal/controller/spserv/sprpc"
	"github.com/MaKcm14/vk-test/pkg/subpub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SubPubServer defines the logic of handling the grpc's requests.
type SubPubServer struct {
	sprpc.UnimplementedPubSubServer
	serv subpub.SubPub
	log  *slog.Logger

	// flagDone is the flag that stores the current service's condition.
	flagDone atomic.Bool

	subMsgCh syncMap
	curID    atomic.Int64
}

func NewSubPubServer(log *slog.Logger, serv subpub.SubPub) *SubPubServer {
	return &SubPubServer{
		log:      log,
		serv:     serv,
		subMsgCh: newSyncMap(),
	}
}

// Subscribe defines the logic of the handling the subscribe requests.
func (s *SubPubServer) Subscribe(request *sprpc.SubscribeRequest, stream grpc.ServerStreamingServer[sprpc.Event]) error {
	const op = "spserv.Subscribe"

	msgCh := make(chan string)
	curID := int(s.curID.Load())

	s.curID.Add(1)
	s.subMsgCh.Add(int(s.curID.Load()), msgCh)

	sub, err := s.serv.Subscribe(request.Key, func(msg interface{}) {
		msgCh <- msg.(string)
	})

	if err != nil {
		var code codes.Code
		var subErr error

		if errors.Is(err, subpub.ErrInputData) {
			code = codes.InvalidArgument
			subErr = fmt.Errorf("%w: %s", ErrDataRequest, err)
		} else if errors.Is(err, subpub.ErrSystemCondition) {
			code = codes.Unavailable
			subErr = fmt.Errorf("%w: %s", ErrServiceCondition, err)
		}
		s.log.Error(fmt.Sprintf("error of the %s: %s", op, subErr))

		return status.Error(code, subErr.Error())
	}

	for msg := range msgCh {
		if err := stream.Send(&sprpc.Event{Data: msg}); err != nil {
			sub.Unsubscribe()
			close(s.subMsgCh.Delete(curID))

			sendErr := fmt.Errorf("%w: %s", ErrSendingMsg, err)
			s.log.Error(fmt.Sprintf("error of the %s: %s", op, sendErr))

			return status.Error(codes.Aborted, sendErr.Error())
		}
	}

	if s.flagDone.Load() {
		sub.Unsubscribe()
		return status.Error(codes.Aborted, ErrServiceCondition.Error())
	}
	s.curID.Add(-1)

	return nil
}

// Publish defines the logic of the handling the publish requests.
func (s *SubPubServer) Publish(_ context.Context, request *sprpc.PublishRequest) (*emptypb.Empty, error) {
	const op = "spserv.Publish"

	if err := s.serv.Publish(request.Key, request.Data); err != nil {
		var code codes.Code
		var pubErr error

		if errors.Is(err, subpub.ErrInputData) {
			code = codes.InvalidArgument
			pubErr = fmt.Errorf("%w: %s", ErrDataRequest, err)
		} else if errors.Is(err, subpub.ErrSystemCondition) {
			code = codes.Unavailable
			pubErr = fmt.Errorf("%w: %s", ErrServiceCondition, err)
		}
		s.log.Error(fmt.Sprintf("error of the %s: %s", op, pubErr))

		return nil, status.Error(code, pubErr.Error())
	}

	return &emptypb.Empty{}, nil
}

// Close releases the resources of the SubPubServer.
func (s *SubPubServer) Close() {
	s.flagDone.Store(true)

	s.subMsgCh.Range(func(ch chan string) {
		close(ch)
	})

	s.serv.Close(context.Background())
}
