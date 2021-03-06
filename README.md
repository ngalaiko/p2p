# Peer to Peer messenger [![Go Report Card](https://goreportcard.com/badge/github.com/ngalayko/p2p)](https://goreportcard.com/report/github.com/ngalayko/p2p)

## Features

1. UDP multicast discovery within a local network
2. End-to-end encryption

## Peer local run 

```bash
> go get -u github.com/ngalayko/p2p/...
> cd $GOPATH/src/github.com/ngalayko/p2p
> go run ./cmd/peer/main.go
> open http://127.0.0.1:30003
```

## Help 

```bash
> peer --help
```

## Dispatcher local run

**NOTE: Requires docker swarm and local resolver (or `/etc/hosts` changes)**

```bash
> docker swarm init
> docker stack deploy -c docker-compose.yaml messenger
> open http://localhost
```
