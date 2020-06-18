package rpc

import (
	"io"

	"github.com/lithdew/bytesutil"
)

type ServiceResponsePacket struct {
	StreamId uint32            // stream id
	Handled  bool              // whether or not the service was handled
	Headers  map[string]string // headers for this packet
}

func (p ServiceResponsePacket) AppendTo(dst []byte) []byte {
	dst = bytesutil.AppendUint32BE(dst, p.StreamId)
	if p.Handled {
		dst = append(dst, 1)
	} else {
		dst = append(dst, 0)
	}
	if p.Headers != nil {
		dst = bytesutil.AppendUint16BE(dst, uint16(len(p.Headers)))
		for name, value := range p.Headers {
			dst = append(dst, byte(len(name)))
			dst = append(dst, name...)
			dst = bytesutil.AppendUint16BE(dst, uint16(len(value)))
			dst = append(dst, value...)
		}
	} else {
		dst = bytesutil.AppendUint16BE(dst, 0)
	}
	return dst
}

func UnmarshalServiceResponsePacket(buf []byte) (ServiceResponsePacket, error) {
	var packet ServiceResponsePacket

	{
		if len(buf) < 5 {
			return packet, io.ErrUnexpectedEOF
		}

		packet.StreamId, buf = bytesutil.Uint32BE(buf[:4]), buf[4:]
		packet.Handled, buf = buf[0] == 1, buf[1:]
	}

	{
		if len(buf) < 2 {
			return packet, io.ErrUnexpectedEOF
		}

		var size uint16
		size, buf = bytesutil.Uint16BE(buf[:2]), buf[2:]

		packet.Headers = make(map[string]string, size)
		for i := uint16(0); i < size; i++ {
			{
				if len(buf) < 1 {
					return packet, io.ErrUnexpectedEOF
				}
				var nameSize uint8
				nameSize, buf = buf[0], buf[1:]
				if len(buf) < int(nameSize) {
					return packet, io.ErrUnexpectedEOF
				}
				var name string
				name, buf = string(buf[:nameSize]), buf[nameSize:]

				if len(buf) < 2 {
					return packet, io.ErrUnexpectedEOF
				}
				var valueSize uint16
				valueSize, buf = bytesutil.Uint16BE(buf[:2]), buf[2:]
				if len(buf) < int(valueSize) {
					return packet, io.ErrUnexpectedEOF
				}
				var value string
				value, buf = string(buf[:valueSize]), buf[valueSize:]
				packet.Headers[name] = value
			}
		}
	}

	return packet, nil
}
