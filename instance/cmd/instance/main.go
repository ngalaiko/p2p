package main

import (
	"context"
	"flag"
	"time"

	"github.com/ngalayko/p2p/instance"
	"github.com/ngalayko/p2p/instance/logger"
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
)

func main() {
	flag.Parse()

	m := instance.New(
		logger.ParseLevel(*logLevel),
		*udp4Multicast,
		*udp6Multicast,
		*discoveryPort,
		*port,
		*insecurePort,
		*uiPort,
		*discoveryInterval,
	)
	if err := m.Start(context.Background()); err != nil {
		panic(err)
	}
}
