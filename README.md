# Peer to Peer messenger [![Go Report Card](https://goreportcard.com/badge/github.com/ngalayko/p2p)](https://goreportcard.com/report/github.com/ngalayko/p2p)

## Features

1. UDP multicast discovery within a local network
2. End-to-end encryption

## Local run 

```bash
go get -u github.com/ngalayko/p2p/...
go install github.com/ngalayko/p2p/cmd/peer
cd $GOPATH/github.com/ngalayko/p2p
peer
open http://127.0.0.1:30003
```

## Help 

```bash
$ peer --help
```
