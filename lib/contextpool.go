package lib

import (
	"sync"
	"sync/atomic"
)

type Context struct {
	conn *Conn
	seq  uint32
	buf  []byte
}

func (c *Context) Conn() *Conn            { return c.conn }
func (c *Context) Body() []byte           { return c.buf }
func (c *Context) Reply(buf []byte) error { return c.conn.send(c.seq, buf) }

type ContextPool struct {
	sp sync.Pool
	m  *PoolMetrics
}

func (p *ContextPool) metrics() *PoolMetrics {
	return p.m
}

func (p *ContextPool) acquire(conn *Conn, seq uint32, buf []byte) *Context {
	v := p.sp.Get()
	if v == nil {
		v = &Context{}
		atomic.AddUint32(&p.m.na, uint32(1))
	} else {
		atomic.AddUint32(&p.m.nr, uint32(1))
	}
	ctx := v.(*Context)
	ctx.conn = conn
	ctx.seq = seq
	ctx.buf = buf
	return ctx
}

func (p *ContextPool) release(ctx *Context) {
	p.sp.Put(ctx)
	atomic.AddUint32(&p.m.np, uint32(1))
}
