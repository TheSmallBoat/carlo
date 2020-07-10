package streaming_rpc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	st "github.com/TheSmallBoat/carlo/streaming_transmit"
	"github.com/lithdew/kademlia"
)

// Service Provider
type Provider struct {
	kadId *kademlia.ID
	conn  *st.Conn

	services map[string]struct{}

	mu      sync.Mutex         // protects all stream-related structures
	counter uint32             // total number of outgoing streams
	streams map[uint32]*Stream // maps stream ids to stream instances
}

func (p *Provider) NextStream() *Stream {
	reader, writer := createWrappedPipe()

	p.mu.Lock()
	defer p.mu.Unlock()

	id := p.counter
	p.counter += 2

	stream := &Stream{
		ID:     id,
		Reader: reader,
		Writer: writer,
	}
	stream.wg.Add(1)
	p.streams[id] = stream

	return stream
}

func (p *Provider) GetStream(id uint32) (*Stream, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	stream, exists := p.streams[id]
	return stream, exists
}

func (p *Provider) RegisterStreamWithServiceRequestPacket(header ServiceRequestPacket) (*Stream, bool) {
	reader, writer := createWrappedPipe()

	p.mu.Lock()
	defer p.mu.Unlock()

	stream, exists := p.streams[header.StreamId]
	if exists {
		return stream, false
	}

	stream = &Stream{
		ID:     header.StreamId,
		Reader: reader,
		Writer: writer,
	}
	stream.wg.Add(1)
	p.streams[header.StreamId] = stream

	return stream, true
}

func (p *Provider) CloseStreamWithError(stream *Stream, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	_ = stream.Reader.CloseWithError(err)
	_ = stream.Writer.CloseWithError(err)

	stream.once.Do(stream.wg.Done)

	delete(p.streams, stream.ID)
}

func (p *Provider) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for sid, stream := range p.streams {
		err := fmt.Errorf("provider connection closed: %w", io.EOF)

		_ = stream.Reader.CloseWithError(err)
		_ = stream.Writer.CloseWithError(err)

		stream.once.Do(stream.wg.Done)

		delete(p.streams, sid)
	}
}

var ErrProviderNotAvailable = errors.New("provider unable to provide service")

func (p *Provider) Push(services []string, headers map[string]string, body io.ReadCloser) (*Stream, error) {
	buf := make([]byte, ChunkSize)

	stream := p.NextStream()

	header := ServiceRequestPacket{
		StreamId: stream.ID,
		Headers:  headers,
		Services: services,
	}

	err := p.conn.Send(header.AppendTo([]byte{OpCodeServiceRequest}))
	if err != nil {
		err = fmt.Errorf("failed to send stream header: %s: %w", err, ErrProviderNotAvailable)
		p.CloseStreamWithError(stream, err)
		return nil, err
	}

	for {
		nn, err := body.Read(buf[:ChunkSize])
		if err != nil && err != io.EOF {
			err = fmt.Errorf("failed reading body: %w", err)
			p.CloseStreamWithError(stream, err)
			return nil, err
		}

		payload := DataPacket{
			StreamID: stream.ID,
			Data:     buf[:nn],
		}

		if err := p.conn.Send(payload.AppendTo([]byte{OpCodeData})); err != nil {
			err = fmt.Errorf("failed writing body chunk as a data packet to peer: %w", err)
			p.CloseStreamWithError(stream, err)
			return nil, err
		}

		if err == io.EOF && nn == 0 {
			break
		}
	}

	stream.wg.Wait()

	if stream.Header == nil {
		return nil, fmt.Errorf("no response headers were returned: %w", io.EOF)
	}

	if !stream.Header.Handled {
		return nil, fmt.Errorf("provider unable to service: %s", services)
	}

	return stream, nil
}

func (p *Provider) Services() []string {
	services := make([]string, 0, len(p.services))
	for service := range p.services {
		services = append(services, service)
	}
	return services
}

func (p *Provider) KadID() *kademlia.ID {
	return p.kadId
}

func (p *Provider) Addr() string {
	if p.kadId != nil {
		return HostAddr(p.kadId.Host, p.kadId.Port)
	} else {
		return "<anon>"
	}
}

func HostAddr(host net.IP, port uint16) string {
	h := ""
	if len(host) > 0 {
		h = host.String()
	}
	p := strconv.FormatUint(uint64(port), 10)
	return net.JoinHostPort(h, p)
}
