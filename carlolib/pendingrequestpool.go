package carlolib

import (
	"sync"
	"sync/atomic"
)

var pendingRequestPool = &PendingRequestPool{sp: sync.Pool{}}

type pendingRequest struct {
	dst []byte         // dst to copy response to
	wg  sync.WaitGroup // signals the caller that the response has been received
}

type PendingRequestPool struct {
	sp sync.Pool
	na uint64 // number of new acquires
	nr uint64 // number of reuse from pool
	np uint64 // number of put back to pool
}

func (p *PendingRequestPool) acquire(dst []byte) *pendingRequest {
	v := p.sp.Get()
	if v == nil {
		v = &pendingRequest{}
		atomic.AddUint64(&p.na, uint64(1))
	} else {
		atomic.AddUint64(&p.nr, uint64(1))
	}
	pr := v.(*pendingRequest)
	pr.dst = dst
	return pr
}

func (p *PendingRequestPool) release(pr *pendingRequest) {
	p.sp.Put(pr)
	atomic.AddUint64(&p.np, uint64(1))
}
