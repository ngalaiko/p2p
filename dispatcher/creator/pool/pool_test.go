package pool

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ngalayko/p2p/dispatcher/creator/mock"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
	"github.com/stretchr/testify/assert"
)

func Test_Pool_Create__should_fill_pool_on_create(t *testing.T) {
	pool := New(log, mock.New(
		func(context.Context) (*peers.Peer, *url.URL, error) {
			i := rand.Int()
			u, _ := url.Parse(fmt.Sprintf("http://fast.%d/", i))
			return &peers.Peer{
				ID: fmt.Sprint(i),
			}, u, nil
		}), 5)

	<-time.Tick(100 * time.Millisecond)

	assert.Equal(t, 5, pool.Len())
}

func Test_Pool_Create__should_return_from_the_pool(t *testing.T) {
	var c int32

	pool := New(log, mock.New(func(context.Context) (*peers.Peer, *url.URL, error) {
		if atomic.LoadInt32(&c) == 1 {
			time.Sleep(time.Hour)
		}
		atomic.AddInt32(&c, 1)
		return &peers.Peer{ID: "test"}, nil, nil
	}), 1)

	<-time.Tick(100 * time.Millisecond)

	assert.Equal(t, 1, pool.Len())
	_, _, _ = pool.Create(context.Background())
	assert.Equal(t, 0, pool.Len())
}

func Test_Pool_Create__should_fallback_when_pool_is_empty(t *testing.T) {
	var c int32

	pool := New(log, mock.New(func(context.Context) (*peers.Peer, *url.URL, error) {
		t := atomic.LoadInt32(&c)
		if t != 0 {
			return nil, nil, fmt.Errorf("%d call", t+1)
		}
		atomic.AddInt32(&c, 1)
		return &peers.Peer{ID: "test"}, nil, nil
	}), 1)

	<-time.Tick(100 * time.Millisecond)

	assert.Equal(t, 1, pool.Len())
	_, _, _ = pool.Create(context.Background())
	assert.Equal(t, 0, pool.Len())
	_, _, err := pool.Create(context.Background())
	assert.Equal(t, 0, pool.Len())
	assert.Equal(t, "2 call", err.Error())
}

//
// helpers
//

var log = logger.New(logger.LevelDebug)
