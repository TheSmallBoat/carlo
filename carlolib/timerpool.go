package carlolib

import (
	"sync"
	"time"
)

var zeroTime time.Time
var timerPool = newTimerPool()

type TimerPool struct {
	sp sync.Pool
}

func newTimerPool() *TimerPool {
	return &TimerPool{sp: sync.Pool{}}
}

func (p *TimerPool) acquire(timeout time.Duration) *time.Timer {
	v := p.sp.Get()
	if v == nil {
		return time.NewTimer(timeout)
	}
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
}
