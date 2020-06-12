package carlolib

import (
	"sync"
	"sync/atomic"
)

var contextPool = &ContextPool{sp: sync.Pool{}}

type Context struct {
	conn *Conn
	seq  uint32
	buf  []byte
}

func (c *Context) Body() []byte           { return c.buf }
func (c *Context) Reply(buf []byte) error { return c.conn.send(c.seq, buf) }

// na + nr equal the total number of acquires
// na + nr - np equal the number of still running.
type ContextPool struct {
	sp sync.Pool
	na uint64 // number of new acquires
	nr uint64 // number of reuse from pool
	np uint64 // number of put back to pool
}

func (p *ContextPool) acquire(conn *Conn, seq uint32, buf []byte) *Context {
	v := p.sp.Get()
	if v == nil {
		v = &Context{}
		atomic.AddUint64(&p.na, uint64(1))
	} else {
		atomic.AddUint64(&p.nr, uint64(1))
	}
	ctx := v.(*Context)
	ctx.conn = conn
	ctx.seq = seq
	ctx.buf = buf
	return ctx
}

func (p *ContextPool) release(ctx *Context) {
	p.sp.Put(ctx)
	atomic.AddUint64(&p.np, uint64(1))
}
