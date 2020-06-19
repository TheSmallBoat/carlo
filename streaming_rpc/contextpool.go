package streaming_rpc

import (
	"io"
	"sync"

	st "github.com/TheSmallBoat/carlo/streaming_transmit"
)

type ContextPool struct {
	sp sync.Pool
}

func (p *ContextPool) acquire(headers map[string]string, body io.ReadCloser, id uint32, conn *st.Conn) *Context {
	v := p.sp.Get()
	if v == nil {
		v = &Context{responseHeaders: make(map[string]string)}
	}
	ctx := v.(*Context)
	ctx.Headers = headers
	ctx.Body = body
	ctx.StreamId = id
	ctx.Conn = conn
	return ctx
}

func (p *ContextPool) release(ctx *Context) {
	ctx.written = false
	for key := range ctx.responseHeaders {
		delete(ctx.responseHeaders, key)
	}
	p.sp.Put(ctx)
}
