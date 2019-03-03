package main

import (
	"context"
	"flag"

	"github.com/ngalayko/p2p/dispatcher"
	"github.com/ngalayko/p2p/logger"
)

var (
	logLevel        = flag.String("log_level", "info", "log level [debug|info|error|panic]")
	port            = flag.String("port", "20000", "port to listen")
	jwtSecret       = flag.String("jwt_secret", "secret", "secret to sign jwt tokens with")
	peerImageName   = flag.String("image_name", "docker.io/ngalayko/peer", "name of the peer image to pull")
	peerNetworkName = flag.String("network_name", "p2p", "name of the peer docker network")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	d := dispatcher.New(
		ctx,
		logger.New(logger.ParseLevel(*logLevel)),
		*jwtSecret,
		*peerImageName,
		*peerNetworkName,
	)

	if err := d.Start(ctx, *port); err != nil {
		panic(err)
	}
}
