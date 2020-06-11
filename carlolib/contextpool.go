package carlolib

import "sync"

var contextPool = &ContextPool{sp: sync.Pool{}}

type Context struct {
	conn *Conn
	seq  uint32
	buf  []byte
}

func (c *Context) Body() []byte           { return c.buf }
func (c *Context) Reply(buf []byte) error { return c.conn.send(c.seq, buf) }

type ContextPool struct {
	sp sync.Pool
}

func (p *ContextPool) acquire(conn *Conn, seq uint32, buf []byte) *Context {
	v := p.sp.Get()
	if v == nil {
		v = &Context{}
	}
	ctx := v.(*Context)
	ctx.conn = conn
	ctx.seq = seq
	ctx.buf = buf
	return ctx
}

func (p *ContextPool) release(ctx *Context) { p.sp.Put(ctx) }
