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
	peerServiceName = flag.String("peer_service", "", "name of the peer service to scale")
	consulURL       = flag.String("consul", "consul:8500", "url to contact consul catalog")
	staticPath      = flag.String("staticPath", "./dispatcher/public", "path to static files")
	buffer          = flag.Int("buffer", 3, "number peers to create in advance")
)

func main() {
	flag.Parse()

	log := logger.New(logger.ParseLevel(*logLevel))

	if *peerServiceName == "" {
		log.Panic("peer service name can not be empty")
	}

	ctx := context.Background()

	d := dispatcher.New(
		ctx,
		log,
		*jwtSecret,
		*peerServiceName,
		*consulURL,
		*buffer,
	)

	if err := d.Start(ctx, *port, *staticPath); err != nil {
		panic(err)
	}
}
