package streaming_transmit

import (
	"fmt"
	"sync/atomic"
	"time"
)

var DefaultTickerDuration = 1 * time.Second

// na + nr equal the total number of acquires
// na + nr - np equal the number of still running.
type PoolMetrics struct {
	na uint32 // number of new acquires
	nr uint32 // number of reuse from pool
	np uint32 // number of put back to pool

	naa uint64 // accumulative
	nra uint64 // accumulative
	npa uint64 // accumulative

	done chan struct{}
}

func newPoolMetrics() *PoolMetrics {
	pm := &PoolMetrics{}
	pm.done = make(chan struct{})

	return pm
}

func (p *PoolMetrics) release() {
	p.done <- struct{}{}
}

func (p *PoolMetrics) setMetrics() {
	atomic.AddUint64(&p.naa, uint64(atomic.SwapUint32(&p.na, uint32(0))))
	atomic.AddUint64(&p.nra, uint64(atomic.SwapUint32(&p.nr, uint32(0))))
	atomic.AddUint64(&p.npa, uint64(atomic.SwapUint32(&p.np, uint32(0))))
}

func (p *PoolMetrics) start() {
	timer := time.NewTicker(DefaultTickerDuration)

	go func() {
		defer close(p.done)

		for {
			select {
			case <-timer.C:
				p.setMetrics()
			case <-p.done:
				p.setMetrics()
				return
			}
		}
	}()
}

func (p *PoolMetrics) metricsString() string {
	return fmt.Sprintf("[ %v|%v|%v, %v|%v|%v ]", p.na, p.nr, p.np, p.naa, p.nra, p.npa)
}
