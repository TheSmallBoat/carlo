package carlolib

import "sync"

var pendingRequestPool = &PendingRequestPool{sp: sync.Pool{}}

type pendingRequest struct {
	dst []byte         // dst to copy response to
	wg  sync.WaitGroup // signals the caller that the response has been received
}

type PendingRequestPool struct {
	sp sync.Pool
}

func (p *PendingRequestPool) acquire(dst []byte) *pendingRequest {
	v := p.sp.Get()
	if v == nil {
		v = &pendingRequest{}
	}
	pr := v.(*pendingRequest)
	pr.dst = dst
	return pr
}

func (p *PendingRequestPool) release(pr *pendingRequest) { p.sp.Put(pr) }
