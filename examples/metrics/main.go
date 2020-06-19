package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	st "github.com/TheSmallBoat/carlo/streaming_transmit"
)

func main() {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	st.StartPoolMetrics()

	ln, err := net.Listen("tcp", ":4444")
	check(err)

	client := &st.Client{Addr: ln.Addr().String()}

	var server st.Server
	go func() {
		defer ln.Close()
		defer server.Shutdown()
		defer client.Shutdown()

		check(server.Serve(ln))
	}()

	var wg sync.WaitGroup
	n := 4
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 1024*256; j++ {
				check(client.Send([]byte(fmt.Sprintf("[%d %d] Hello from Go!", i, j))))
			}
			fmt.Println(st.JsonStringPoolMetrics())
		}(i)
	}

	wg.Wait()

	st.ReleasePoolMetrics()
	time.Sleep(200 * time.Millisecond)
	fmt.Println(st.JsonStringPoolMetrics())
}
