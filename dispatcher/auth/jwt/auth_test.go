package jwt

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ngalayko/p2p/instance/peers"
)

func Test_Store__should_store_jwt_token(t *testing.T) {
	testPeer := &peers.Peer{
		ID:   "test id",
		Name: "test name",
	}

	recorder := httptest.NewRecorder()

	err := New("test secret").Store(recorder, testPeer)
	assert.NoError(t, err)

	cc := recorder.Result().Cookies()

	if assert.Equal(t, 1, len(cc)) {
		assert.Equal(t, cookieName, cc[0].Name)
	}
}

func Test_Store__should_get_stored_peer(t *testing.T) {
	testPeer := &peers.Peer{
		ID:   "test id",
		Name: "test name",
	}

	recorder := httptest.NewRecorder()

	auth := New("test secret")

	err := auth.Store(recorder, testPeer)
	assert.NoError(t, err)

	cc := recorder.Result().Cookies()

	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(cc[0])

	outPeer, err := auth.Get(r)
	assert.NoError(t, err)

	assert.Equal(t, testPeer, outPeer)
}
