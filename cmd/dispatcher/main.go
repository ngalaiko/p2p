package main

import (
	"context"
	"flag"

	"github.com/ngalayko/p2p/dispatcher"
	"github.com/ngalayko/p2p/logger"
)

var (
	logLevel        = flag.String("log_level", "info", "log level [debug|info|warning|error|panic]")
	port            = flag.String("port", "20000", "port to listen")
	jwtSecret       = flag.String("jwt_secret", "secret", "secret to sign jwt tokens with")
	peerImageName   = flag.String("image_name", "docker.io/ngalayko/peer", "name of the peer image to pull")
	peerNetworkName = flag.String("network_name", "p2p", "name of the peer docker network")
	consulURL       = flag.String("consul", "http://consul:8500", "url to contact consul kv api")
	host            = flag.String("hostname", "localhost", "hostname of the dispatcher")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	d := dispatcher.New(
		ctx,
		logger.New(logger.ParseLevel(*logLevel)),
		*jwtSecret,
		*host,
		*peerImageName,
		*peerNetworkName,
		*consulURL,
	)

	if err := d.Start(ctx, *port); err != nil {
		panic(err)
	}
}
