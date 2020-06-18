package streaming_rpc

import (
	"io"
	"sync"

	carlo "github.com/TheSmallBoat/carlo/lib"
)

type ContextPool struct {
	sp sync.Pool
}

func (p *ContextPool) acquire(headers map[string]string, body io.ReadCloser, id uint32, conn *carlo.Conn) *Context {
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
