package lib

import (
	"fmt"
	"sync"
	"time"
)

var zeroTime time.Time

var timerPool = &TimerPool{sp: sync.Pool{}, m: newPoolMetrics()}
var contextPool = &ContextPool{sp: sync.Pool{}, m: newPoolMetrics()}
var pendingRequestPool = &PendingRequestPool{sp: sync.Pool{}, m: newPoolMetrics()}
var pendingWritePool = &PendingWritePool{sp: sync.Pool{}, m: newPoolMetrics()}

func StartPoolMetrics() {
	timerPool.m.start()
	contextPool.m.start()
	pendingRequestPool.m.start()
	pendingWritePool.m.start()
}

func ReleasePoolMetrics() {
	timerPool.m.release()
	contextPool.m.release()
	pendingRequestPool.m.release()
	pendingWritePool.m.release()
}

func JsonStringPoolMetrics() string {
	return fmt.Sprintf("{\"TimerPool\" = %s, \"contextPool\" = %s, \"pendingRequestPool\" = %s, \"pendingWritePool\" = %s}",
		timerPool.m.metricsString(),
		contextPool.m.metricsString(),
		pendingRequestPool.m.metricsString(),
		pendingWritePool.m.metricsString(),
	)
}
