package client

import (
	"context"
	"fmt"
	"time"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/peers"
)

// Client used to send messages to a peer.
type Client struct {
	logger *logger.Logger
	creds  credentials.TransportCredentials
	client *peers.Peer

	conn *grpc.ClientConn
}

// Connect connects to a client.
func Connect(
	ctx context.Context,
	log *logger.Logger,
	creds credentials.TransportCredentials,
	client *peers.Peer,
) (*Client, error) {
	c := &Client{
		logger: log.Prefix("grpc-client-%s", client.ID),
		creds:  creds,
		client: client,
	}

	var err error
	c.conn, err = grpc.DialContext(
		ctx,
		fmt.Sprintf("client:///%s", client.ID),
		grpc.WithTransportCredentials(c.creds),
		grpc.WithBalancerName("pick_first"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %s", err)
	}

	return c, nil
}

// Send sends a message.
func (c *Client) Send(ctx context.Context, m *messages.Message) error {
	streamClient, err := chat.NewChatClient(c.conn).Stream(ctx)
	if err != nil {
		return fmt.Errorf("can't open stream: %s", err)
	}

	if err := streamClient.Send(makeMessage(m)); err != nil {
		return fmt.Errorf("can't send a message: %s", err)
	}

	c.logger.Info("messsage sent")

	return nil
}

func makeMessage(msg *messages.Message) *chat.Message {
	now := time.Now()
	m := &chat.Message{
		Timestamp: &timestamp.Timestamp{
			Seconds: now.Unix(),
		},
	}
	switch msg.Type {
	case messages.TypeText:
		m.Payload = &chat.Message_Text{
			Text: msg.Text,
		}
	}
	return m
}
