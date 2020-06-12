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

	t.Logf("Timer Pool => new:%d,reuse:%d,putback:%d", timerPool.m.na, timerPool.m.nr, timerPool.m.np)
	t.Logf("Context Pool => new:%d,reuse:%d,putback:%d", contextPool.m.na, contextPool.m.nr, contextPool.m.np)
	t.Logf("PendingRequest Pool => new:%d,reuse:%d,putback:%d", pendingRequestPool.m.na, pendingRequestPool.m.nr, pendingRequestPool.m.np)
	t.Logf("PendingWrite Pool => new:%d,reuse:%d,putback:%d", pendingWritePool.m.na, pendingWritePool.m.nr, pendingWritePool.m.np)
}
