package client

import (
	"context"
	"fmt"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

const (
	idLen = 32
)

// Client used to open a stream connection with another peer.
type Client struct {
	chat.Chat_StreamClient

	logger *logger.Logger
	client *peers.Peer
	r      *rand.Rand
}

// Connect connects to a client.
func Connect(
	ctx context.Context,
	r *rand.Rand,
	log *logger.Logger,
	creds credentials.TransportCredentials,
	client *peers.Peer,
) (*Client, error) {
	c := &Client{
		logger: log.Prefix("grpc-client-%s", client.ID),
		client: client,
		r:      r,
	}

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("peer:///%s", client.ID),
		grpc.WithTransportCredentials(creds),
		grpc.WithBalancerName("pick_first"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %s", err)
	}

	streamClient, err := chat.NewChatClient(conn).Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't open stream: %s", err)
	}
	c.Chat_StreamClient = streamClient

	return c, nil
}
