package carlolib

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestServerShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	srv := &Server{}

	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		srv.Shutdown()
		ln.Close()
	}()

	require.NoError(t, srv.Serve(ln))

	t.Logf("Timer Pool => new:%d,reuse:%d,putback:%d", timerPool.na, timerPool.nr, timerPool.np)
	t.Logf("Context Pool => new:%d,reuse:%d,putback:%d", contextPool.na, contextPool.nr, contextPool.np)
	t.Logf("PendingRequest Pool => new:%d,reuse:%d,putback:%d", pendingRequestPool.na, pendingRequestPool.nr, pendingRequestPool.np)
	t.Logf("PendingWrite Pool => new:%d,reuse:%d,putback:%d", pendingWritePool.na, pendingWritePool.nr, pendingWritePool.np)
}
