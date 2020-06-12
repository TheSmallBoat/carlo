package carlolib

import (
	"sync"
	"sync/atomic"
	"time"
)

var zeroTime time.Time
var timerPool = &TimerPool{sp: sync.Pool{}}

type TimerPool struct {
	sp sync.Pool
	na uint64 // number of new acquires
	nr uint64 // number of reuse from pool
	np uint64 // number of put back to pool
}

func (p *TimerPool) acquire(timeout time.Duration) *time.Timer {
	v := p.sp.Get()
	if v == nil {
		atomic.AddUint64(&p.na, uint64(1))
		return time.NewTimer(timeout)
	}
	atomic.AddUint64(&p.nr, uint64(1))
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
	atomic.AddUint64(&p.np, uint64(1))
}
