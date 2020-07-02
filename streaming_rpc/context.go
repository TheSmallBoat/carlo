package streaming_rpc

import (
	"io"

	st "github.com/TheSmallBoat/carlo/streaming_transmit"
	"github.com/lithdew/kademlia"
)

const ChunkSize = 2048

type Handler func(ctx *Context)

type OpCode = uint8

const (
	OpCodeHandshake OpCode = iota
	OpCodeServiceRequest
	OpCodeServiceResponse
	OpCodeData
	OpCodeFindNodeRequest
	OpCodeFindNodeResponse
)

var _ io.Writer = (*Context)(nil)

type Context struct {
	KadId   kademlia.ID
	Headers map[string]string
	Body    io.ReadCloser

	streamId uint32 // stream id
	conn     *st.Conn

	responseHeaders map[string]string // response headers
	written         bool              // written before?
}

func (c *Context) WriteHeader(key, val string) {
	c.responseHeaders[key] = val
}

// Implement Write function for io.Writer interface
func (c *Context) Write(data []byte) (int, error) {
	if len(data) == 0 { // disallow writing zero bytes
		return 0, nil
	}

	if !c.written {
		packet := ServiceResponsePacket{
			StreamId: c.streamId,
			Handled:  true,
			Headers:  c.responseHeaders,
		}

		c.written = true

		err := c.conn.Send(packet.AppendTo([]byte{OpCodeServiceResponse}))
		if err != nil {
			return 0, err
		}
	}

	for i := 0; i < len(data); i += ChunkSize {
		start := i
		end := i + ChunkSize
		if end > len(data) {
			end = len(data)
		}

		packet := DataPacket{
			StreamID: c.streamId,
			Data:     data[start:end],
		}

		err := c.conn.Send(packet.AppendTo([]byte{OpCodeData}))
		if err != nil {
			return 0, err
		}
	}

	return len(data), nil
}
