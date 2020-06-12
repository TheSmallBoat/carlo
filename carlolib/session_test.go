package carlolib

import (
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func newTestSession(t testing.TB) Session {
	t.Helper()
	sess, err := NewSession()
	require.NoError(t, err)
	return sess
}

func TestSessionConn(t *testing.T) {
	defer goleak.VerifyNone(t)

	alice, bob := net.Pipe()
	defer func() {
		require.NoError(t, alice.Close())
		require.NoError(t, bob.Close())
	}()

	a := newTestSession(t)
	b := newTestSession(t)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		require.NoError(t, a.DoClient(alice))
	}()

	go func() {
		defer wg.Done()
		require.NoError(t, b.DoServer(bob))
	}()

	wg.Wait()

	aliceConn := NewSessionConn(a.Suite(), alice)
	bobConn := NewSessionConn(b.Suite(), bob)

	trials := 1024

	go func() {
		for i := 0; i < trials; i++ {
			_, err := aliceConn.Write(strconv.AppendUint(nil, uint64(i), 10))
			require.NoError(t, err)
		}
		require.NoError(t, aliceConn.Flush())
	}()

	buf := make([]byte, 1024)

	for i := 0; i < trials; i++ {
		n, err := bobConn.Read(buf)
		require.NoError(t, err)
		require.EqualValues(t, strconv.AppendUint(nil, uint64(i), 10), buf[:n])
	}

	t.Logf("Timer Pool => new:%d,reuse:%d,putback:%d", timerPool.m.na, timerPool.m.nr, timerPool.m.np)
	t.Logf("Context Pool => new:%d,reuse:%d,putback:%d", contextPool.m.na, contextPool.m.nr, contextPool.m.np)
	t.Logf("PendingRequest Pool => new:%d,reuse:%d,putback:%d", pendingRequestPool.m.na, pendingRequestPool.m.nr, pendingRequestPool.m.np)
	t.Logf("PendingWrite Pool => new:%d,reuse:%d,putback:%d", pendingWritePool.m.na, pendingWritePool.m.nr, pendingWritePool.m.np)
}

func TestSession(t *testing.T) {
	defer goleak.VerifyNone(t)

	aliceSession := newTestSession(t)
	bobSession := newTestSession(t)

	bob, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	ch := make(chan []byte, 1)
	go func() {
		alice, err := net.Dial("tcp", bob.Addr().String())
		require.NoError(t, err)

		require.NoError(t, aliceSession.DoClient(alice))
		require.NotNil(t, aliceSession.SharedKey())
		require.NotNil(t, aliceSession.Suite())

		require.NoError(t, alice.Close())

		ch <- aliceSession.SharedKey()
		close(ch)
	}()

	conn, err := bob.Accept()
	require.NoError(t, err)

	require.NoError(t, bobSession.DoServer(conn))
	require.NotNil(t, bobSession.SharedKey())
	require.NotNil(t, bobSession.Suite())

	require.EqualValues(t, bobSession.SharedKey(), <-ch)

	require.NoError(t, conn.Close())
	require.NoError(t, bob.Close())

	t.Logf("Timer Pool => new:%d,reuse:%d,putback:%d", timerPool.m.na, timerPool.m.nr, timerPool.m.np)
	t.Logf("Context Pool => new:%d,reuse:%d,putback:%d", contextPool.m.na, contextPool.m.nr, contextPool.m.np)
	t.Logf("PendingRequest Pool => new:%d,reuse:%d,putback:%d", pendingRequestPool.m.na, pendingRequestPool.m.nr, pendingRequestPool.m.np)
	t.Logf("PendingWrite Pool => new:%d,reuse:%d,putback:%d", pendingWritePool.m.na, pendingWritePool.m.nr, pendingWritePool.m.np)

}
