package carlolib

import (
	"sync"
	"time"
)

type Conn struct {
	Handler Handler

	ReadBufferSize  int
	WriteBufferSize int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	mu   sync.Mutex
	once sync.Once

	writerQueue []*pendingWrite
	writerCond  sync.Cond
	writerDone  bool

	reqs map[uint32]*pendingRequest
	seq  uint32
}
