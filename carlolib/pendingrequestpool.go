package carlolib

import "sync"

type pendingRequest struct {
	dst []byte         // dst to copy response to
	wg  sync.WaitGroup // signals the caller that the response has been received
}

type PendingRequestPool struct {
	sp sync.Pool
}

func NewPendingRequestPool() *PendingRequestPool {
	return &PendingRequestPool{sp: sync.Pool{}}
}

func (p *PendingRequestPool) Acquire(dst []byte) *pendingRequest {
	v := p.sp.Get()
	if v == nil {
		v = &pendingRequest{}
	}
	pr := v.(*pendingRequest)
	pr.dst = dst
	return pr
}

func (p *PendingRequestPool) Release(pr *pendingRequest) { p.sp.Put(pr) }
