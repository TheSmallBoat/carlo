package carlolib

import "sync"

var contextPool = newContextPool()

type SendFunc func(seq uint32, buf []byte) error

type Context struct {
	seq  uint32
	buf  []byte
	send SendFunc
}

func (c *Context) Body() []byte           { return c.buf }
func (c *Context) Reply(buf []byte) error { return c.send(c.seq, buf) }

type ContextPool struct {
	sp sync.Pool
}

func newContextPool() *ContextPool {
	return &ContextPool{sp: sync.Pool{}}
}

func (p *ContextPool) acquire(send SendFunc, seq uint32, buf []byte) *Context {
	v := p.sp.Get()
	if v == nil {
		v = &Context{}
	}
	ctx := v.(*Context)
	ctx.send = send
	ctx.seq = seq
	ctx.buf = buf
	return ctx
}

func (p *ContextPool) release(ctx *Context) { p.sp.Put(ctx) }
