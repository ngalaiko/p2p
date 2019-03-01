package peers

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func NewTestPeer(t *testing.T) *Peer {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	defer func() {
		port += 2
	}()
	p, err := New(r, getNextPort(), getNextPort(), 512)
	if err != nil {
		t.Fatalf("can't create a test peer: %s", err)
	}
	return p
}

var port = 1000

func getNextPort() string {
	port++
	return strconv.Itoa(port)
}
