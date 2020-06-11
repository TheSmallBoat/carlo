package carlolib

import (
	"sync"
	"time"
)

var zeroTime time.Time
var timePool = newTimePool()

type TimePool struct {
	sp sync.Pool
}

func newTimePool() *TimePool {
	return &TimePool{sp: sync.Pool{}}
}

func (p *TimePool) Acquire(timeout time.Duration) *time.Timer {
	v := p.sp.Get()
	if v == nil {
		return time.NewTimer(timeout)
	}
	t := v.(*time.Timer)
	t.Reset(timeout)
	return t
}

func (p *TimePool) Release(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	p.sp.Put(t)
}
