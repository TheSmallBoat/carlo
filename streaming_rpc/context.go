package streaming_rpc

import (
	"io"

	carlo_ "github.com/TheSmallBoat/carlo/lib"
)

const ChunkSize = 2048

type OpCode = uint8

const (
	OpCodeHandshake OpCode = iota
	OpCodeServiceRequest
	OpCodeServiceResponse
	OpCodeData
	OpCodeFindNodeRequest
	OpCodeFindNodeResponse
)

type Handler func(ctx *Context)

var _ io.Writer = (*Context)(nil)

type Context struct {
	StreamId uint32 // stream id
	Conn     *carlo_.Conn

	Headers map[string]string
	Body    io.ReadCloser

	responseHeaders map[string]string // response headers
	written         bool              // written before?
}

func (c *Context) WriteHeader(key, val string) {
	c.responseHeaders[key] = val
}

// Implement Write function for io.Writer interface
func (c *Context) Write(data []byte) (int, error) {
	if !c.written {
		packet := ServiceResponsePacket{
			StreamId: c.StreamId,
			Handled:  true,
			Headers:  c.responseHeaders,
		}

		c.written = true

		err := c.Conn.Send(packet.AppendTo([]byte{OpCodeServiceResponse}))
		if err != nil {
			return 0, err
		}
	}

	if len(data) == 0 { // disallow writing zero bytes
		return 0, nil
	}

	for i := 0; i < len(data); i += ChunkSize {
		start := i
		end := i + ChunkSize
		if end > len(data) {
			end = len(data)
		}

		packet := DataPacket{
			StreamID: c.StreamId,
			Data:     data[start:end],
		}

		err := c.Conn.Send(packet.AppendTo([]byte{OpCodeData}))
		if err != nil {
			return 0, err
		}
	}

	return len(data), nil
}
