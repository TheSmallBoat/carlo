package streaming_rpc

import (
	"sync"
)

type Stream struct {
	ID     uint32
	Reader *pipeReader
	Writer *pipeWriter

	Header *ServiceResponsePacket

	received uint64

	wg   sync.WaitGroup
	once sync.Once
}
