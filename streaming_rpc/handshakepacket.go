package streaming_rpc

import (
	"errors"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
	"unsafe"

	"github.com/lithdew/kademlia"
)

type HandshakePacket struct {
	KadId     *kademlia.ID
	Services  []string
	Signature kademlia.Signature
}

func (h HandshakePacket) AppendPayloadTo(dst []byte) []byte {
	dst = h.KadId.AppendTo(dst)
	for _, service := range h.Services {
		dst = append(dst, service...)
	}
	return dst
}

func (h HandshakePacket) AppendTo(dst []byte) []byte {
	if h.KadId != nil {
		dst = append(dst, 1)
		dst = h.KadId.AppendTo(dst)
	} else {
		dst = append(dst, 0)
	}
	dst = append(dst, uint8(len(h.Services)))
	for _, service := range h.Services {
		dst = append(dst, uint8(len(service)))
		dst = append(dst, service...)
	}
	if h.KadId != nil {
		dst = append(dst, h.Signature[:]...)
	}
	return dst
}

func UnmarshalHandshakePacket(buf []byte) (HandshakePacket, error) {
	var pkt HandshakePacket

	if len(buf) < 1 {
		return pkt, io.ErrUnexpectedEOF
	}

	hasID := buf[0] == 1
	buf = buf[1:]

	if hasID {
		id, leftover, err := kademlia.UnmarshalID(buf)
		if err != nil {
			return pkt, err
		}
		pkt.KadId = &id

		buf = leftover
	}

	if len(buf) < 1 {
		return pkt, io.ErrUnexpectedEOF
	}

	var size uint8
	size, buf = buf[0], buf[1:]

	if len(buf) < int(size) {
		return pkt, io.ErrUnexpectedEOF
	}

	pkt.Services = make([]string, size)
	for i := 0; i < len(pkt.Services); i++ {
		if len(buf) < 1 {
			return pkt, io.ErrUnexpectedEOF
		}
		size, buf = buf[0], buf[1:]
		if len(buf) < int(size) {
			return pkt, io.ErrUnexpectedEOF
		}
		pkt.Services[i] = string(buf[:size])
		buf = buf[size:]
	}

	if hasID {
		if len(buf) < kademlia.SizeSignature {
			return pkt, io.ErrUnexpectedEOF
		}

		pkt.Signature, buf = *(*kademlia.Signature)(unsafe.Pointer(&((buf[:kademlia.SizeSignature])[0]))),
			buf[kademlia.SizeSignature:]
	}

	return pkt, nil
}

func (h HandshakePacket) Validate(dst []byte) error {
	if h.KadId != nil {
		err := h.KadId.Validate()
		if err != nil {
			return err
		}
	}

	for _, service := range h.Services {
		if !utf8.ValidString(service) {
			return fmt.Errorf("service '%s' in hello packet is not valid utf8", service)
		}
		if len(service) > math.MaxUint8 {
			return fmt.Errorf("service '%s' in hello packet is too large - must <= %d bytes",
				service, math.MaxUint8)
		}
	}

	if h.KadId != nil && !h.Signature.Verify(h.KadId.Pub, h.AppendPayloadTo(dst)) {
		return errors.New("signature is malformed")
	}

	return nil
}
