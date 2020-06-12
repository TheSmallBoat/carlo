package carlolib

import (
	"sync"
	"sync/atomic"
	"time"
)

var zeroTime time.Time
var timerPool = &TimerPool{sp: sync.Pool{}, m: &PoolMetrics{}}

type TimerPool struct {
	sp sync.Pool
	m  *PoolMetrics
}

func (p *TimerPool) acquire(timeout time.Duration) *time.Timer {
	v := p.sp.Get()
	if v == nil {
		atomic.AddUint32(&p.m.na, uint32(1))
		return time.NewTimer(timeout)
	}
	atomic.AddUint32(&p.m.nr, uint32(1))
	t := v.(*time.Timer)
	t.Reset(timeout)
	return t
}

func (p *TimerPool) release(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	p.sp.Put(t)
	atomic.AddUint32(&p.m.np, uint32(1))
}
