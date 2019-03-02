package main

import (
	"context"
	"flag"

	"github.com/ngalayko/p2p/dispatcher"
	"github.com/ngalayko/p2p/instance/logger"
)

var (
	port      = flag.String("port", "20000", "port to listen")
	jwtSecret = flag.String("jwt_secret", "secret", "secret to sign jwt tokens with")
	logLevel  = flag.String("log_level", "info", "log level [debug|info|error|panic]")
)

func main() {
	flag.Parse()

	d := dispatcher.New(
		logger.New(logger.ParseLevel(*logLevel)),
		*jwtSecret,
	)

	if err := d.Start(context.Background(), *port); err != nil {
		panic(err)
	}
}
