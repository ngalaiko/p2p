package merge

import (
	"context"
	"fmt"
	"testing"

	"github.com/ngalayko/p2p/instance/discovery/mock"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/stretchr/testify/assert"
)

func Test_Discover__should_merge_2_services(t *testing.T) {
	pp := []*peers.Peer{}
	for i := 0; i < 10; i++ {
		pp = append(pp, &peers.Peer{ID: fmt.Sprintf("%d", i)})
	}

	d := New(
		mock.New(pp...),
		mock.New(pp...),
	)

	cc := 0
	for range d.Discover(context.Background()) {
		cc++
	}

	assert.Equal(t, 20, cc)
}

func Test_Discover__should_merge_0_services(t *testing.T) {
	d := New()

	cc := 0
	for range d.Discover(context.Background()) {
		cc++
	}

	assert.Equal(t, 0, cc)
}
