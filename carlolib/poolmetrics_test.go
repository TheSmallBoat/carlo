package carlolib

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestPoolMetrics(t *testing.T) {
	defer goleak.VerifyNone(t)

	StartPoolMetrics()

	n := 4
	m := 1024

	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	var server Server

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(t, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(t, ln.Close())
	}()

	var wg sync.WaitGroup

	for k := 0; k < 32; k++ {
		wg.Add(n)

		for i := 0; i < n; i++ {
			go func(i int) {
				defer wg.Done()
				for j := 0; j < m; j++ {
					require.NoError(t, client.Send([]byte(fmt.Sprintf("[%d] hello %d", i, j))))
				}
			}(i)
		}

		wg.Wait()
		t.Logf("%s", JsonStringPoolMetrics())
		time.Sleep(200 * time.Millisecond)
	}
	ReleasePoolMetrics()
	time.Sleep(200 * time.Millisecond)
	t.Logf("%s", JsonStringPoolMetrics())
}
