FROM golang:1.11.5-alpine as builder

WORKDIR ${GOPATH}/src/github.com/ngalayko/p2p/app

COPY . .

RUN go build -o /p2p ./cmd/p2p/main.go

FROM alpine:3.9

COPY --from=builder /p2p /p2p

ENTRYPOINT [ "/p2p" ]
