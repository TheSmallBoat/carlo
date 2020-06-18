package streaming_rpc

import (
	"io"
	"sync"
)

type Stream struct {
	ID     uint32
	Reader *io.PipeReader
	Writer *io.PipeWriter

	Header *ServiceResponsePacket

	received uint64

	wg   sync.WaitGroup
	once sync.Once
}
