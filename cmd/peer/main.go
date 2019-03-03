package main

import (
	"context"
	"flag"
	"net"
	"time"

	"github.com/ngalayko/p2p/client"
	"github.com/ngalayko/p2p/instance"
	"github.com/ngalayko/p2p/logger"
)

var (
	logLevel          = flag.String("log_level", "info", "logging level [debug|info|error|panic]")
	udp6Multicast     = flag.String("udp6_multicast", "[ff02::114]", "multicast addr for udp6 discrvery")
	udp4Multicast     = flag.String("udp4_multicast", "239.255.255.250", "multicast addr for udp4 discrvery")
	port              = flag.String("port", "30000", "port to listen for messages")
	insecurePort      = flag.String("insecure_port", "30001", "port to listen for greetings")
	discoveryPort     = flag.String("discovery_port", "30002", "port to discover other peers")
	uiPort            = flag.String("ui_port", "30003", "port to serve ui interface")
	discoveryInterval = flag.Duration("discovery_interval", 1*time.Second, "interval to send discovery broadcast")
	statisPath        = flag.String("static_path", "./client/public", "path to static files for ui")
)

func main() {
	flag.Parse()

	log := logger.New(logger.ParseLevel(*logLevel))

	inst := instance.New(
		log,
		*udp4Multicast,
		*udp6Multicast,
		*discoveryPort,
		*port,
		*insecurePort,
		*discoveryInterval,
	)

	client := client.New(
		log,
		net.JoinHostPort("0.0.0.0", *uiPort),
		inst,
		*statisPath,
	)

	ctx := context.Background()

	go func() {
		if err := client.Start(ctx); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := inst.Start(ctx); err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()
}
