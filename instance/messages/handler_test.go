package messages

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/peers"
)

func Test_Handler__should_shotdown_on_done(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	h := testHandler(t)

	done := make(chan bool)
	go func() {
		if err := h.Start(ctx); err != nil {
			t.Fatal(err)
		}

		close(done)
	}()

	cancel()

	<-done
}

func Test_Handler__should_not_sent_a_message_to_unknown_peer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hSender := testHandler(t)
	go run(ctx, t, hSender)

	hReceiver := testHandler(t)
	go run(ctx, t, hReceiver)

	sent := make(chan *Message)
	go func() {
		sent <- <-hReceiver.Received()
	}()

	err := hSender.SendText(ctx, "test", hReceiver.self.ID)
	assert.Error(t, err)
}

func Test_Handler__should_send_a_message_to_another_peer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hSender := testHandler(t)
	go run(ctx, t, hSender)

	hReceiver := testHandler(t)
	go run(ctx, t, hReceiver)

	hSender.self.KnownPeers.Add(hReceiver.self)

	received := make(chan *Message)
	go func() {
		received <- <-hSender.Sent()
	}()

	sent := make(chan *Message)
	go func() {
		sent <- <-hReceiver.Received()
	}()

	err := hSender.SendText(ctx, "test", hReceiver.self.ID)
	assert.NoError(t, err, "can't send a message")

	sentMsg := <-sent

	receivedMsg := <-received

	assert.Equal(t, sentMsg.ID, receivedMsg.ID)
	assert.Equal(t, sentMsg.Text, receivedMsg.Text)
}

func Test_Handler__should_send_a_message_to_self(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hSender := testHandler(t)
	go run(ctx, t, hSender)

	received := make(chan *Message)
	go func() {
		received <- <-hSender.Sent()
	}()

	sent := make(chan *Message)
	go func() {
		sent <- <-hSender.Received()
	}()

	err := hSender.SendText(ctx, "test", hSender.self.ID)
	assert.NoError(t, err, "can't send a message")

	sentMsg := <-sent
	receivedMsg := <-received

	assert.Equal(t, sentMsg.ID, receivedMsg.ID)
	assert.Equal(t, sentMsg.Text, receivedMsg.Text)
}

//
// helpers
//

func run(ctx context.Context, t *testing.T, h *Handler) {
	if err := h.Start(ctx); err != nil {
		t.Fatalf("failed to start a server: %s", err)
	}
}

func testHandler(t *testing.T) *Handler {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	return NewHandler(
		r,
		logger.New(logger.LevelDebug),
		testPeer(t),
	)
}

func testPeer(t *testing.T) *peers.Peer {
	defer func() {
		port += 2
	}()

	r := rand.New(rand.NewSource(int64(port)))
	p, err := peers.New(r, getNextPort(), getNextPort(), 512)
	if err != nil {
		t.Fatalf("can't create a test peer: %s", err)
	}
	p.Addrs.Add(net.ParseIP("127.0.0.1"))
	return p
}

var port = 1000

func getNextPort() string {
	port++
	return strconv.Itoa(port)
}
