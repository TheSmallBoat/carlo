package streaming_rpc

import (
	"io"
	"sync"

	st "github.com/TheSmallBoat/carlo/streaming_transmit"
	"github.com/lithdew/kademlia"
)

type ContextPool struct {
	sp sync.Pool
}

func (p *ContextPool) acquire(kadId kademlia.ID, headers map[string]string, body io.ReadCloser, streamId uint32, conn *st.Conn) *Context {
	v := p.sp.Get()
	if v == nil {
		v = &Context{responseHeaders: make(map[string]string)}
	}
	ctx := v.(*Context)
	ctx.KadId = kadId
	ctx.Headers = headers
	ctx.Body = body

	ctx.streamId = streamId
	ctx.conn = conn

	return ctx
}

func (p *ContextPool) release(ctx *Context) {
	ctx.written = false
	for key := range ctx.responseHeaders {
		delete(ctx.responseHeaders, key)
	}
	p.sp.Put(ctx)
}
