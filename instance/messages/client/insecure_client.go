package client

import (
	"context"
	"fmt"
	"math/rand"

	"google.golang.org/grpc"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages/proto/greeter"
	"github.com/ngalayko/p2p/instance/peers"
)

// InsecureClient used to open an insecure connection with another peer.
type InsecureClient struct {
	greeter.GreeterClient

	logger *logger.Logger
	client *peers.Peer
	r      *rand.Rand
}

// InsecureConnect connects to a client.
func InsecureConnect(
	ctx context.Context,
	log *logger.Logger,
	client *peers.Peer,
) (*InsecureClient, error) {
	insecureConn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("greet:///%s", client.ID),
		grpc.WithBalancerName("pick_first"),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to insecure: %s", err)
	}

	return &InsecureClient{
		GreeterClient: greeter.NewGreeterClient(insecureConn),
		logger:        log.Prefix("grpc-insecure-client-%s", client.ID),
		client:        client,
	}, nil
}
