package carlolib

import "sync"

type ContextFunc func(seq uint32, buf []byte) error

type Context struct {
	fn ContextFunc
	seq  uint32
	buf  []byte
}

func (c *Context) Body() []byte           { return c.buf }
func (c *Context) Reply(buf []byte) error { return c.fn(c.seq, buf) }

type ContextPool struct {
	sp sync.Pool
}

func NewContextPool() *ContextPool {
	return &ContextPool{sp: sync.Pool{}}
}

func (p *ContextPool) Acquire(fn ContextFunc, seq uint32, buf []byte) *Context {
	v := p.sp.Get()
	if v == nil {
		v = &Context{}
	}
	ctx := v.(*Context)
	ctx.fn = fn
	ctx.seq = seq
	ctx.buf = buf
	return ctx
}

func (p *ContextPool) Release(ctx *Context) { p.sp.Put(ctx) }