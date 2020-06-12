package carlolib

import (
	"sync"
	"sync/atomic"

	"github.com/valyala/bytebufferpool"
)

var pendingWritePool = &PendingWritePool{sp: sync.Pool{}, m: newPoolMetrics()}

type pendingWrite struct {
	buf  *bytebufferpool.ByteBuffer // payload
	wait bool                       // signal to caller if they're waiting
	err  error                      // keeps track of any socket errors on write
	wg   sync.WaitGroup             // signals the caller that this write is complete
}

type PendingWritePool struct {
	sp sync.Pool
	m  *PoolMetrics
}

func (p *PendingWritePool) acquire(buf *bytebufferpool.ByteBuffer, wait bool) *pendingWrite {
	v := p.sp.Get()
	if v == nil {
		v = &pendingWrite{}
		atomic.AddUint32(&p.m.na, uint32(1))
	} else {
		atomic.AddUint32(&p.m.nr, uint32(1))
	}

	pw := v.(*pendingWrite)
	pw.buf = buf
	pw.wait = wait
	return pw
}

func (p *PendingWritePool) release(pw *pendingWrite) {
	p.sp.Put(pw)
	atomic.AddUint32(&p.m.np, uint32(1))
}
