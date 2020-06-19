package streaming_rpc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	st "github.com/TheSmallBoat/carlo/streaming_transmit"
	"github.com/jpillora/backoff"
	"github.com/lithdew/kademlia"
)

var _ st.Handler = (*StreamNode)(nil)
var _ st.ConnStateHandler = (*StreamNode)(nil)

type NetProtocolType = uint8

const (
	NetProtocolTCP NetProtocolType = iota
	NetProtocolUDP
	NetProtocolIP
)

type StreamNode struct {
	start sync.Once
	stop  sync.Once
	wg    sync.WaitGroup

	SecretKey   kademlia.PrivateKey // Used to sign for the HandshakePacket
	NetProtocol NetProtocolType     // Such as TCP,UDP,IP

	tableMu sync.Mutex
	table   *kademlia.Table
	kadId   *kademlia.ID

	srv       *st.Server
	providers *Providers
	Services  map[string]Handler

	clientsMu sync.Mutex
	clients   map[string]*st.Client
}

func GenerateSecretKey() kademlia.PrivateKey {
	_, secret, err := kademlia.GenerateKeys(nil)
	if err != nil {
		panic(err)
	}
	return secret
}

func NewStreamNode(sk kademlia.PrivateKey, kid *kademlia.ID, tab *kademlia.Table) *StreamNode {
	return &StreamNode{
		start:       sync.Once{},
		stop:        sync.Once{},
		wg:          sync.WaitGroup{},
		SecretKey:   sk,
		NetProtocol: NetProtocolTCP,
		tableMu:     sync.Mutex{},
		table:       tab,
		kadId:       kid,
		srv:         nil,
		providers:   NewProviders(),
		Services:    make(map[string]Handler),
		clientsMu:   sync.Mutex{},
		clients:     make(map[string]*st.Client),
	}
}

func (n *StreamNode) SetNetProtocol(np NetProtocolType) {
	if np == NetProtocolTCP || np == NetProtocolUDP || np == NetProtocolIP {
		n.NetProtocol = np
	} else {
		log.Print("Unknown NetProtocol")
	}
}

func (n *StreamNode) Bootstrap() {
	var mu sync.Mutex

	var pub kademlia.PublicKey
	if n.kadId != nil {
		pub = n.kadId.Pub
	}

	busy := make(chan struct{}, kademlia.DefaultBucketSize)
	queue := make(chan kademlia.ID, kademlia.DefaultBucketSize)

	visited := make(map[kademlia.PublicKey]struct{})

	n.tableMu.Lock()
	kids := n.table.ClosestTo(pub, kademlia.DefaultBucketSize)
	n.tableMu.Unlock()

	if len(kids) == 0 {
		return
	}

	visited[pub] = struct{}{}
	for _, kid := range kids {
		queue <- kid
		visited[kid.Pub] = struct{}{}
	}

	var results []kademlia.ID

	for len(queue) > 0 || len(busy) > 0 {
		select {
		case kid := <-queue:
			busy <- struct{}{}
			go func() {
				defer func() { <-busy }()

				client := n.getClient(HostAddr(kid.Host, kid.Port))

				conn, err := client.Get()
				if err != nil {
					return
				}

				err = n.probeWithStreamTransmitConn(conn)
				if err != nil {
					return
				}

				req := n.createFindNodeRequest().AppendTo([]byte{OpCodeFindNodeRequest})

				buf, err := conn.Request(req[:0], req)
				if err != nil {
					return
				}

				res, _, err := kademlia.UnmarshalFindNodeResponse(buf)
				if err != nil {
					return
				}

				mu.Lock()
				for _, kid := range res.Closest {
					if _, seen := visited[kid.Pub]; !seen {
						visited[kid.Pub] = struct{}{}
						results = append(results, kid)
						queue <- kid
					}
				}
				mu.Unlock()
			}()
		default:
			time.Sleep(16 * time.Millisecond)
		}
	}

	log.Printf("Discovered %d peer(s).", len(results))
}

func (n *StreamNode) setStreamTransmitServer() {
	n.srv = &st.Server{
		Handler:   n,
		ConnState: n,
	}
}

func (n *StreamNode) getClient(addr string) *st.Client {
	n.clientsMu.Lock()
	defer n.clientsMu.Unlock()

	client, exists := n.clients[addr]
	if !exists {
		client = &st.Client{
			Addr:      addr,
			Handler:   n,
			ConnState: n,
		}
		n.clients[addr] = client
	}

	return client
}

func (n *StreamNode) probeWithStreamTransmitConn(conn *st.Conn) error {
	provider, exists := n.providers.registerProvider(conn, nil, nil, true)

	req := n.createHandshakePacket(nil).AppendTo([]byte{OpCodeHandshake})
	res, err := conn.Request(req[:0], req)
	if err != nil {
		return err
	}
	packet, err := UnmarshalHandshakePacket(res)
	if err != nil {
		return err
	}
	err = packet.Validate(req[:0])
	if err != nil {
		return err
	}

	if packet.KadId != nil {
		//addr = Addr(packet.ID.Host, packet.ID.Port)
		//if !packet.ID.Host.Equal(resolved.IP) || packet.ID.Port != uint16(resolved.Port) {
		//	return provider, fmt.Errorf("dialed '%s' which advertised '%s'", resolved, addr)
		//}

		// update the provider with id and services info
		provider, _ = n.providers.registerProvider(conn, packet.KadId, packet.Services, true)
		if !exists {
			log.Printf("You are now connected to %s. Services: %s", provider.Addr(), provider.Services())
		} else {
			log.Printf("Re-probed %s. Services: %s", provider.Addr(), provider.Services())
		}

		// update the routing table
		n.tableMu.Lock()
		n.table.Update(*packet.KadId)
		n.tableMu.Unlock()
	}

	return nil
}

func (n *StreamNode) createHandshakePacket(buf []byte) HandshakePacket {
	packet := HandshakePacket{Services: n.getServiceNames()}
	if n.kadId != nil {
		packet.KadId = n.kadId
		packet.Signature = n.SecretKey.Sign(packet.AppendPayloadTo(buf))
	}
	return packet
}

func (n *StreamNode) createFindNodeRequest() kademlia.FindNodeRequest {
	var req kademlia.FindNodeRequest
	if n.kadId != nil {
		req.Target = n.kadId.Pub
	}
	return req
}

func (n *StreamNode) getServiceNames() []string {
	if n.Services == nil {
		return nil
	}
	services := make([]string, 0, len(n.Services))
	for service := range n.Services {
		services = append(services, service)
	}
	return services
}

func (n *StreamNode) Push(services []string, headers map[string]string, body io.ReadCloser) (*Stream, error) {
	providers := n.providers.getProviders(services...)

	for _, provider := range providers {
		stream, err := provider.Push(services, headers, body)
		if err != nil {
			if errors.Is(err, ErrProviderNotAvailable) {
				continue
			}
			return nil, err
		}
		return stream, nil
	}

	return nil, fmt.Errorf("no nodes were able to process your request for service(s): %s", services)
}

// Implement HandleConnState function for the ConnStateHandler interface
func (n *StreamNode) HandleConnState(conn *st.Conn, state st.ConnState) {
	if state != st.StateClosed {
		return
	}

	provider := n.providers.deregisterProvider(conn)
	if provider == nil {
		return
	}

	n.tableMu.Lock()
	if provider.kadId != nil {
		n.table.Delete(provider.kadId.Pub)
	}
	n.tableMu.Unlock()

	addr := provider.Addr()

	log.Printf("%s has disconnected from you. Services: %s", addr, provider.Services())

	n.clientsMu.Lock()
	_, exists := n.clients[addr]
	n.clientsMu.Unlock()

	if !exists {
		return
	}

	go func() {
		b := &backoff.Backoff{
			Factor: 1.25,
			Jitter: true,
			Min:    500 * time.Millisecond,
			Max:    1 * time.Second,
		}

		for i := 0; i < 8; i++ { // 8 attempts max
			err := n.ProbeWithAddr(addr)
			if err == nil {
				break
			}

			duration := b.Duration()

			log.Printf("Trying to reconnect to %s. Sleeping for %s.", addr, duration)
			time.Sleep(duration)
		}

		log.Printf("Tried 8 times reconnecting to %s. Giving up.", addr)
	}()
}

func (n *StreamNode) getResolvedString(addr string) (string, error) {
	switch n.NetProtocol {
	case NetProtocolTCP:
		resolved, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return "", err
		}
		return resolved.String(), nil
	case NetProtocolUDP:
		resolved, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return "", err
		}
		return resolved.String(), nil
	case NetProtocolIP:
		resolved, err := net.ResolveIPAddr("ip", addr)
		if err != nil {
			return "", err
		}
		return resolved.String(), nil
	}
	return "", errors.New("unknown the net protocol type, must be tcp/udp/ip")
}

func (n *StreamNode) ProbeWithAddr(addr string) error {
	k, err := n.getResolvedString(addr)
	if err != nil {
		return err
	}

	conn, err := n.getClient(k).Get()
	if err != nil {
		return err
	}

	err = n.probeWithStreamTransmitConn(conn)
	if err != nil {
		return err
	}

	return nil
}

// Implement HandleMessage function for the Handler interface
func (n *StreamNode) HandleMessage(ctx *st.Context) error {
	var opcode uint8

	body := ctx.Body()
	if len(body) < 1 {
		return errors.New("no opcode recorded in packet")
	}
	opcode, body = body[0], body[1:]

	switch opcode {
	case OpCodeHandshake:
		packet, err := UnmarshalHandshakePacket(body)
		if err != nil {
			return err
		}

		err = packet.Validate(body[:0])
		if err != nil {
			return err
		}

		// register the microservice if it hasn't been registered before
		provider, exists := n.providers.registerProvider(ctx.Conn(), packet.KadId, packet.Services, false)
		if !exists {
			log.Printf("%s has connected. Services: %s", provider.Addr(), provider.Services())
		}

		// register the peer to the routing table
		if packet.KadId != nil {
			n.tableMu.Lock()
			n.table.Update(*packet.KadId)
			n.tableMu.Unlock()
		}

		// always reply back with what services we provide, and our
		// id if we want to publicly advertise our microservice
		return ctx.Reply(n.createHandshakePacket(body[:0]).AppendTo(body[:0]))
	case OpCodeServiceRequest:
		provider := n.providers.findProvider(ctx.Conn())
		if provider == nil {
			return errors.New("conn is not a provider")
		}

		packet, err := UnmarshalServiceRequestPacket(body)
		if err != nil {
			return fmt.Errorf("failed to decode service request packet: %w", err)
		}

		// if stream does not exist, the stream is a request to process some payload! register
		// a new stream, and handle it.

		// if the stream exists, it's a stream we made! proceed to start receiving its payload,
		// b/c the payload will basically be our peer giving us a response.

		// i guess we need to separate streams from the server-side, and streams from the client-side.
		// client-side starts with stream ids that are odd (1, 3, 5, 7, 9, ...) and server-side
		// starts with stream ids that are even (0, 2, 4, 6, 8, 10).

		stream, created := provider.RegisterStreamWithServiceRequestPacket(packet)
		if !created {
			return fmt.Errorf("got service request with stream id %d, but node is making service request"+
				"with the given id already", packet.StreamId)
		}

		for _, service := range packet.Services {
			handler, exists := n.Services[service]
			if !exists {
				continue
			}

			go func() {
				ctx := contextPool.acquire(packet.Headers, stream.Reader, stream.ID, ctx.Conn())
				defer contextPool.release(ctx)

				handler(ctx)

				if !ctx.written {
					packet := ServiceResponsePacket{
						StreamId: ctx.StreamId,
						Handled:  true,
						Headers:  ctx.responseHeaders,
					}

					ctx.written = true

					err := ctx.Conn.Send(packet.AppendTo([]byte{OpCodeServiceResponse}))
					if err != nil {
						provider.CloseStreamWithError(stream, err)
						return
					}

					return
				}

				err := ctx.Conn.Send(DataPacket{StreamID: stream.ID}.AppendTo([]byte{OpCodeData}))
				if err != nil {
					provider.CloseStreamWithError(stream, err)
					return
				}
			}()

			return nil
		}

		response := ServiceResponsePacket{
			StreamId: packet.StreamId,
			Handled:  false,
		}

		return ctx.Conn().SendNoWait(response.AppendTo([]byte{OpCodeServiceResponse}))
	case OpCodeServiceResponse:
		provider := n.providers.findProvider(ctx.Conn())
		if provider == nil {
			return errors.New("conn is not a provider")
		}

		packet, err := UnmarshalServiceResponsePacket(body)
		if err != nil {
			return fmt.Errorf("failed to decode services response packet: %w", err)
		}

		stream, exists := provider.GetStream(packet.StreamId)
		if !exists {
			return fmt.Errorf("stream with id %d got a service response although no service request was sent",
				packet.StreamId)
		}

		stream.Header = &packet
		stream.once.Do(stream.wg.Done)

		return nil
	case OpCodeData:
		provider := n.providers.findProvider(ctx.Conn())
		if provider == nil {
			return errors.New("conn is not a provider")
		}

		packet, err := UnmarshalDataPacket(body)
		if err != nil {
			return fmt.Errorf("failed to decode stream payload packet: %w", err)
		}

		// the stream must always exist

		stream, exists := provider.GetStream(packet.StreamID)
		if !exists {
			return fmt.Errorf("stream with id %d does not exist", packet.StreamID)
		}

		// there should never be any empty payload packets

		if len(packet.Data) > ChunkSize {
			err = fmt.Errorf("stream with id %d got packet over max limit of ChunkSize bytes: got %d bytes",
				packet.StreamID, len(packet.Data))
			provider.CloseStreamWithError(stream, err)
			return err
		}

		if stream.ID%2 == 1 && stream.Header == nil {
			err = fmt.Errorf("outgoing stream with id %d received a payload packet but has not received a header yet",
				packet.StreamID)
			provider.CloseStreamWithError(stream, err)
			return err
		}

		// if the chunk is zero-length, the stream has been closed

		if len(packet.Data) == 0 {
			provider.CloseStreamWithError(stream, io.EOF)
		} else {
			_, err = stream.Writer.Write(packet.Data)
			if err != nil && !errors.Is(err, io.ErrClosedPipe) {
				err = fmt.Errorf("failed to write payload: %w", err)
				provider.CloseStreamWithError(stream, err)
				return err
			}
		}

		return nil
	case OpCodeFindNodeRequest:
		packet, _, err := kademlia.UnmarshalFindNodeRequest(body)
		if err != nil {
			return err
		}

		n.tableMu.Lock()
		res := kademlia.FindNodeResponse{Closest: n.table.ClosestTo(packet.Target, kademlia.DefaultBucketSize)}
		n.tableMu.Unlock()

		return ctx.Reply(res.AppendTo(nil))
	}

	return fmt.Errorf("unknown opcode %d", opcode)
}

func (n *StreamNode) Start() error {
	start := false
	n.start.Do(func() { start = true })
	if !start {
		return errors.New("listener already started")
	}

	n.setStreamTransmitServer()

	return nil
}

func (n *StreamNode) Shutdown() {
	once := false
	n.start.Do(func() { once = true })
	if once {
		return
	}

	stop := false
	n.stop.Do(func() { stop = true })
	if !stop {
		return
	}

	n.srv.Shutdown()
	n.wg.Wait()
}