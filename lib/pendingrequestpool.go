package lib

import (
	"sync"
	"sync/atomic"
)

type pendingRequest struct {
	dst []byte         // dst to copy response to
	err error          // error while waiting for response
	wg  sync.WaitGroup // signals the caller that the response has been received
}

type PendingRequestPool struct {
	sp sync.Pool
	m  *PoolMetrics
}

func (p *PendingRequestPool) acquire(dst []byte) *pendingRequest {
	v := p.sp.Get()
	if v == nil {
		v = &pendingRequest{}
		atomic.AddUint32(&p.m.na, uint32(1))
	} else {
		atomic.AddUint32(&p.m.nr, uint32(1))
	}
	pr := v.(*pendingRequest)
	pr.dst = dst
	return pr
}

func (p *PendingRequestPool) release(pr *pendingRequest) {
	pr.dst = nil
	pr.err = nil
	p.sp.Put(pr)
	atomic.AddUint32(&p.m.np, uint32(1))
}
