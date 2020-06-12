package carlolib

// na + nr equal the total number of acquires
// na + nr - np equal the number of still running.

type PoolMetrics struct {
	na uint32 // number of new acquires
	nr uint32 // number of reuse from pool
	np uint32 // number of put back to pool
}
