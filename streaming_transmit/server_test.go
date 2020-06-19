package streaming_transmit

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
}
