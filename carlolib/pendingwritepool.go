package carlolib

import (
	"sync"

	"github.com/valyala/bytebufferpool"
)

var pendingWritePool = &PendingWritePool{sp: sync.Pool{}}

type pendingWrite struct {
	buf  *bytebufferpool.ByteBuffer // payload
	wait bool                       // signal to caller if they're waiting
	err  error                      // keeps track of any socket errors on write
	wg   sync.WaitGroup             // signals the caller that this write is complete
}

type PendingWritePool struct {
	sp sync.Pool
}

func (p *PendingWritePool) acquire(buf *bytebufferpool.ByteBuffer, wait bool) *pendingWrite {
	v := p.sp.Get()
	if v == nil {
		v = &pendingWrite{}
	}
	pw := v.(*pendingWrite)
	pw.buf = buf
	pw.wait = wait
	return pw
}

func (p *PendingWritePool) release(pw *pendingWrite) { p.sp.Put(pw) }
