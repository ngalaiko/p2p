package main

import (
	"flag"
	"time"

	"github.com/ngalayko/p2p/app"
	"github.com/ngalayko/p2p/app/logger"
)

var (
	logLevel          = flag.String("log_level", "info", "logging level [debug|info|error|panic]")
	udp6Multicast     = flag.String("udp6_multicast", "[ff02::114]", "multicast addr for udp6 discrvery")
	udp4Multicast     = flag.String("udp4_multicast", "239.255.255.250", "multicast addr for udp4 discrvery")
	port              = flag.String("port", "30001", "port to use")
	uiPort            = flag.String("ui_port", "30001", "port for ui interface")
	discoveryInterval = flag.Duration("discovery_interval", time.Second, "interval to send discovery broadcast")
)

func main() {
	flag.Parse()

	m := p2p.New(
		logger.ParseLevel(*logLevel),
		*udp4Multicast,
		*udp6Multicast,
		*port,
		*uiPort,
		*discoveryInterval,
	)
	if err := m.Start(); err != nil {
		panic(err)
	}
}
