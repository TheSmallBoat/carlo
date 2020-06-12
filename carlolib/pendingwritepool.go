package carlolib

import (
	"sync"
	"sync/atomic"

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
	na uint64 // number of new acquires
	nr uint64 // number of reuse from pool
	np uint64 // number of put back to pool
}

func (p *PendingWritePool) acquire(buf *bytebufferpool.ByteBuffer, wait bool) *pendingWrite {
	v := p.sp.Get()
	if v == nil {
		v = &pendingWrite{}
		atomic.AddUint64(&p.na, uint64(1))
	} else {
		atomic.AddUint64(&p.nr, uint64(1))
	}

	pw := v.(*pendingWrite)
	pw.buf = buf
	pw.wait = wait
	return pw
}

func (p *PendingWritePool) release(pw *pendingWrite) {
	p.sp.Put(pw)
	atomic.AddUint64(&p.np, uint64(1))
}
