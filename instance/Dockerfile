FROM golang:1.11.5-alpine as builder

WORKDIR ${GOPATH}/src/github.com/ngalayko/p2p/instance

COPY . .

RUN go build -o /instance ./cmd/instance/main.go

FROM alpine:3.9

COPY --from=builder /instance /instance

ENTRYPOINT [ "/instance" ]
