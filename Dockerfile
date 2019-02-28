FROM golang:1.11.5-alpine as builder

WORKDIR ${GOPATH}/src/github.com/ngalayko/p2p

COPY . .

RUN go build -o /peer ./cmd/peer/main.go

COPY ./client/public /public

FROM alpine:3.9

COPY --from=builder /peer /peer
COPY --from=builder /public /public

ENTRYPOINT [ "/peer", "--static_path=/public" ]
