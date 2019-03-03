package swarm

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"

	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Swarm can create a new peer service in the docker swarm.
type Swarm struct {
	logger *logger.Logger

	cli *client.Client

	networkName string
	imageName   string
}

// New is a swarm creator constructor.
func New(
	ctx context.Context,
	log *logger.Logger,
	imageName string,
	networkName string,
) *Swarm {
	log = log.Prefix("swarm")

	cli, err := client.NewEnvClient()
	if err != nil {
		log.Panic("failed to create docker client: %s", err)
	}

	log.Info("pulling %s", imageName)
	_, err = cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		log.Panic("failed to pull image '%': %s", imageName, err)
	}
	log.Info("%s is up tp date", imageName)

	return &Swarm{
		logger:      log,
		cli:         cli,
		imageName:   imageName,
		networkName: networkName,
	}
}

// Create implements Creator.
// creates a new docker service in a swarm cluster.
func (s *Swarm) Create(ctx context.Context) (*peers.Peer, error) {
	s.logger.Info("creating a new instance of '%s' in '%s' network", s.imageName, s.networkName)

	resp, err := s.cli.ServiceCreate(
		ctx,
		swarm.ServiceSpec{
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: swarm.ContainerSpec{
					Image: s.imageName,
				},
				Networks: []swarm.NetworkAttachmentConfig{
					{Target: s.networkName},
				},
			},
		},
		types.ServiceCreateOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create docker service: %s", err)
	}

	s.logger.Info("%s created", resp.ID)
	for _, wn := range resp.Warnings {
		s.logger.Warning("container %s: %s", wn)
	}

	return &peers.Peer{
		ID: "test",
	}, nil
}
