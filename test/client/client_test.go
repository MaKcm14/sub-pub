package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/MaKcm14/vk-test/internal/controller/spserv"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientSuite struct {
	suite.Suite
	client spserv.PubSubClient

	conn *grpc.ClientConn
}

func newClientSuite() *ClientSuite {
	const op = "client.newClientSuite"

	c := &ClientSuite{}

	godotenv.Load("../../.env")
	socket := os.Getenv("SOCKET")

	if socket == "" {
		panic(fmt.Sprintf("error of the %s: the SOCKET var is empty", op))
	}

	conn, err := grpc.NewClient(os.Getenv("SOCKET"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		panic(fmt.Sprintf("error of the %s: %s", op, err))
	}
	c.client = spserv.NewPubSubClient(conn)
	c.conn = conn

	return c
}

func (c *ClientSuite) startPublisherCommonWork(msg []*spserv.PublishRequest) {
	time.Sleep(time.Second * 5)
	for _, req := range msg {
		_, err := c.client.Publish(context.Background(), req)

		c.Suite.NoError(err, fmt.Sprintf("expected correct work of Publish: error was got: %s", err))
	}
}

func (c *ClientSuite) TestPositiveCases_PubSubCommonWork() {
	var (
		msg         = make([]*spserv.PublishRequest, 0, 10)
		testChannel = "test-channel"
		testMessage = "test-message"
	)

	for i := 0; i != 10; i++ {
		msg = append(msg, &spserv.PublishRequest{
			Key:  testChannel,
			Data: fmt.Sprintf("%s-%d", testMessage, i),
		})
	}
	stream, err := c.client.Subscribe(context.Background(), &spserv.SubscribeRequest{
		Key: testChannel,
	})

	c.Suite.NoError(err, fmt.Sprintf("expected correct work of Subscribe: error was got: %s", err))

	defer stream.CloseSend()

	go c.startPublisherCommonWork(msg)

	for _, req := range msg {
		ev, err := stream.Recv()

		c.Suite.NoError(err, fmt.Sprintf("expected correct receiving from Recv: error was got: %s", err))

		c.Suite.Equal(req.Data, ev.Data,
			fmt.Sprintf("expected correct order for the corresponding data msg: \nExp: %s <=> Got: %s",
				req.Data, ev.Data))
	}
}

func (c *ClientSuite) TestPublishNegativeCases_PublishUnexistingChannel() {
	var (
		testChannel = "test-channel-unexists"
		testMessage = "test-message"
	)

	_, err := c.client.Publish(context.Background(), &spserv.PublishRequest{
		Key:  testChannel,
		Data: testMessage,
	})
	c.Suite.Error(err, "expected error after the publishing into the unexisting channel")
}

func (c *ClientSuite) TestPublishNegativeCases_PublishEmptyChannel() {
	var testMessage = "test-message"

	_, err := c.client.Publish(context.Background(), &spserv.PublishRequest{
		Key:  "",
		Data: testMessage,
	})
	c.Suite.Error(err, "expected error after the publishing with empty channel")
}

func (c *ClientSuite) TestSubscribeNegativeCases_SubscribeEmptySubject() {
	stream, err := c.client.Subscribe(context.Background(), &spserv.SubscribeRequest{
		Key: "",
	})
	c.Suite.NoError(err, "expected no error after the subscribing with correct grpc config")

	_, err = stream.Recv()
	c.Suite.Error(err, "expected error after using the stream with incorrect channel config")
}

func (c *ClientSuite) close() {
	c.conn.Close()
}

func TestClient(t *testing.T) {
	c := newClientSuite()

	defer c.close()
	suite.Run(t, c)
}
