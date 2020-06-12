package carlolib

import "fmt"

func startPoolMetrics() {
	timerPool.m.start()
	contextPool.m.start()
	pendingRequestPool.m.start()
	pendingWritePool.m.start()
}

func releasePoolMetrics() {
	timerPool.m.release()
	contextPool.m.release()
	pendingRequestPool.m.release()
	pendingWritePool.m.release()
}

func jsonStringPoolMetrics() string {
	return fmt.Sprintf("{\"TimerPool\" = %s, \"contextPool\" = %s, \"pendingRequestPool\" = %s, \"pendingWritePool\" = %s}",
		timerPool.m.metricsString(),
		contextPool.m.metricsString(),
		pendingRequestPool.m.metricsString(),
		pendingWritePool.m.metricsString(),
	)
}
