package lib

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestClientHandshakeTimeout(t *testing.T) {
	defer goleak.VerifyNone(t)

	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	client := &Client{Addr: ln.Addr().String(), HandshakeTimeout: 1 * time.Millisecond}

	defer func() {
		client.Shutdown()
		require.NoError(t, ln.Close())
	}()

	attempts := 16
	go func() {
		for i := 0; i < attempts; i++ {
			_, _ = ln.Accept()
		}
	}()

	for i := 0; i < attempts; i++ {
		require.Error(t, client.Send([]byte("hello\n")))
	}
}

func TestClientSend(t *testing.T) {
	defer goleak.VerifyNone(t)

	n := 4
	m := 1024
	c := uint32(n * m)

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
		require.EqualValues(t, 0, atomic.LoadUint32(&c))
	}()

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < m; j++ {
				require.NoError(t, client.Send([]byte(fmt.Sprintf("[%d] hello %d", i, j))))
				atomic.AddUint32(&c, ^uint32(0))
			}
		}(i)
	}

	wg.Wait()
}

func TestClientRequest(t *testing.T) {
	defer goleak.VerifyNone(t)

	n := 4
	m := 1024
	c := uint32(n * m * 2)

	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	handler := func(ctx *Context) error {
		atomic.AddUint32(&c, ^uint32(0))
		return ctx.Reply([]byte("a reply!"))
	}

	var server Server
	server.Handler = HandlerFunc(handler)

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(t, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(t, ln.Close())
		require.EqualValues(t, 0, atomic.LoadUint32(&c))
	}()

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < m; j++ {
				res, err := client.Request(nil, []byte(fmt.Sprintf("[%d] hello %d", i, j)))
				require.NoError(t, err)
				require.EqualValues(t, []byte("a reply!"), res)
				atomic.AddUint32(&c, ^uint32(0))
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkSend(b *testing.B) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(b, err)

	var server Server

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(b, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(b, ln.Close())
	}()

	buf := make([]byte, 1400)
	_, err = rand.Read(buf)
	require.NoError(b, err)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := client.Send(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSendNoWait(b *testing.B) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(b, err)

	var server Server

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(b, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(b, ln.Close())
	}()

	buf := make([]byte, 1400)
	_, err = rand.Read(buf)
	require.NoError(b, err)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := client.SendNoWait(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRequest(b *testing.B) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(b, err)

	var server Server
	server.Handler = HandlerFunc(func(ctx *Context) error {
		return ctx.Reply(nil)
	})

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(b, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(b, ln.Close())
	}()

	buf := make([]byte, 1400)
	_, err = rand.Read(buf)
	require.NoError(b, err)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, err := client.Request(nil, buf)
		if err != nil {
			b.Fatal(err)
		}
		if len(res) != 0 {
			b.Fatalf("expected empty response, got '%s'", string(res))
		}
	}
}

func BenchmarkParallelSend(b *testing.B) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(b, err)

	var server Server

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(b, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(b, ln.Close())
	}()

	buf := make([]byte, 1400)
	_, err = rand.Read(buf)
	require.NoError(b, err)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := client.Send(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkParallelSendNoWait(b *testing.B) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(b, err)

	var server Server

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(b, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(b, ln.Close())
	}()

	buf := make([]byte, 1400)
	_, err = rand.Read(buf)
	require.NoError(b, err)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := client.SendNoWait(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkParallelRequest(b *testing.B) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(b, err)

	var server Server
	server.Handler = HandlerFunc(func(ctx *Context) error {
		return ctx.Reply(nil)
	})

	client := &Client{Addr: ln.Addr().String()}

	go func() {
		require.NoError(b, server.Serve(ln))
	}()

	defer func() {
		server.Shutdown()
		client.Shutdown()

		require.NoError(b, ln.Close())
	}()

	buf := make([]byte, 1400)
	_, err = rand.Read(buf)
	require.NoError(b, err)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := client.Request(nil, buf)
			if err != nil {
				b.Fatal(err)
			}
			if len(res) != 0 {
				b.Fatalf("expected empty response, got '%s'", string(res))
			}
		}
	})
}
