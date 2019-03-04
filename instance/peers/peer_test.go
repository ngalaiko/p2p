package peers

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Peer__should_marshal_and_unmarshal_json(t *testing.T) {
	p := newTestPeer(t)

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	r := &Peer{}
	err = r.Unmarshal(data)
	assert.NoError(t, err)

	assert.Equal(t, p.ID, r.ID)
	assert.Equal(t, p.Name, r.Name)
	assert.Equal(t, p.Port, r.Port)
	assert.Equal(t, p.InsecurePort, r.InsecurePort)
}

func newTestPeer(t *testing.T) *Peer {
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
